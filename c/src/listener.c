// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0
//
// Multi-client listener (T32.2). Mirrors Go's pigeon.Register /
// pigeon.Listener.Accept in api.go — a backend registers once with
// the relay (PIGEON_ROLE_REGISTER_MUX); thereafter every paired
// client that connects on the shared QUIC connection produces a
// fresh pigeon_session.
//
// Threading model: synchronous inline demux on the calling thread.
// pigeon_listener_accept blocks on transport->accept_stream, reads
// the [4-byte clientTag][varint name-len][name] header that the
// relay prepends, and either:
//
//   * tag is new + name empty -> primary stream for a new client.
//     Run T32.1's pigeon_run_backend_activation against the
//     registered pairing callback, derive an AEAD channel from the
//     resolved PairingRecord, allocate a pigeon_session, register
//     it under its tag in the demux table, and return it.
//
//   * tag is known -> sub-stream for an existing client. Park the
//     stream + name on the session's per-session incoming-stream
//     queue (the listener's pump keeps looping until the next new
//     primary).
//
// The hash table is a 16-slot open-addressing linear-probing map.
// The C SDK targets small N (a personal-device-count of paired
// clients per backend), so 16 slots is plenty and keeps the data
// structure trivial.

#include "pigeon/pigeon.h"
#include "pigeon/activation.h"
#include "pigeon/session_gen.h"

#include <stdlib.h>
#include <string.h>

// --- Per-listener state ---

typedef struct {
    bool            in_use;
    uint32_t        client_tag;
    pigeon_session *session;
} listener_slot;

struct pigeon_listener {
    pigeon_transport          transport;
    char                      instance_id[64];

    // Pairing callback (device-id -> PairingRecord lookup) wired in
    // at init time. Invoked synchronously from the accept pump.
    pigeon_resolve_device_fn  resolve;
    void                     *resolve_userdata;

    // Datagram channel definitions copied into each accepted
    // session.
    pigeon_dgchannel_def      datagrams[PIGEON_MAX_DATAGRAM_CHANNELS];
    size_t                    datagram_count;

    // Per-tag demux table. Fixed-size open-addressing linear
    // probing; sized to PIGEON_LISTENER_MAX_CLIENTS slots.
    listener_slot             slots[PIGEON_LISTENER_MAX_CLIENTS];

    // Optional owned-transport closer. When the high-level
    // pigeon_register wires up an ngtcp2 transport on behalf of the
    // caller, it parks a closer + free callback here so
    // pigeon_listener_close tears the transport down. NULL when the
    // caller owns the transport (the low-level pigeon_listener_init
    // path).
    void  (*owned_transport_close)(void *t);
    void  (*owned_transport_free)(void *t);
    void   *owned_transport;

    bool                      closed;
};

// --- Hash table helpers ---
//
// 32-bit tag fold. The relay assigns sequential client tags so even
// a trivial modulo distributes them well; mix in a fixnum-style
// multiplier so an adversarial workload that picks colliding tags
// doesn't fall off a cliff.

static size_t slot_index(uint32_t tag, size_t probe)
{
    uint32_t hash = tag * 2654435761u; // Knuth fibonacci hash
    return ((size_t)hash + probe) % PIGEON_LISTENER_MAX_CLIENTS;
}

// Find an in-use slot with the given tag, or return NULL.
static listener_slot *slot_lookup(pigeon_listener *l, uint32_t tag)
{
    for (size_t probe = 0; probe < PIGEON_LISTENER_MAX_CLIENTS; probe++) {
        listener_slot *s = &l->slots[slot_index(tag, probe)];
        if (!s->in_use) return NULL; // open addressing: first empty -> miss
        if (s->client_tag == tag) return s;
    }
    return NULL;
}

// Find a free slot to host the given tag, or NULL if the table is
// full. Assumes the tag is not already present.
static listener_slot *slot_insert(pigeon_listener *l, uint32_t tag)
{
    for (size_t probe = 0; probe < PIGEON_LISTENER_MAX_CLIENTS; probe++) {
        listener_slot *s = &l->slots[slot_index(tag, probe)];
        if (!s->in_use) return s;
    }
    return NULL;
}

// --- Session lifecycle ---

// Allocate and initialise a pigeon_session for an activated client.
// Returns NULL on allocation failure; on success the caller registers
// it in the demux table.
static pigeon_session *make_session(pigeon_listener *l,
                                    uint32_t tag,
                                    const pigeon_pairing_record *rec)
{
    pigeon_session *s = (pigeon_session *)calloc(1, sizeof(*s));
    if (!s) return NULL;

    // Derive the session AEAD channel from the resolved
    // PairingRecord. Mirrors api.go's rec.DeriveChannel(
    // "backend->client", "client->backend") on the Listener side.
    uint8_t send_key[32], recv_key[32];
    if (pigeon_derive_session_key(rec->local_private_key,
                                  rec->peer_public_key,
                                  (const uint8_t *)"backend->client",
                                  strlen("backend->client"),
                                  send_key) != 0) {
        free(s);
        return NULL;
    }
    if (pigeon_derive_session_key(rec->local_private_key,
                                  rec->peer_public_key,
                                  (const uint8_t *)"client->backend",
                                  strlen("client->backend"),
                                  recv_key) != 0) {
        free(s);
        return NULL;
    }
    pigeon_channel ch;
    pigeon_channel_init(&ch, send_key, recv_key, PIGEON_MODE_STRICT);

    if (pigeon_session_init(s, &l->transport, &ch,
                            /*is_backend=*/true, tag,
                            l->datagrams, l->datagram_count) != 0) {
        free(s);
        return NULL;
    }
    return s;
}

// Free a session: scratch buffers, any unpicked-up incoming sub-
// streams, then the struct itself. Safe on NULL.
static void free_session(pigeon_listener *l, pigeon_session *s)
{
    if (!s) return;
    for (size_t i = 0; i < PIGEON_SESSION_MAX_INCOMING; i++) {
        if (s->incoming[i].in_use && s->incoming[i].handle != NULL
                && l->transport.close_stream != NULL) {
            (void)l->transport.close_stream(l->transport.userdata,
                                            s->incoming[i].handle);
        }
        s->incoming[i].in_use = false;
        s->incoming[i].handle = NULL;
    }
    pigeon_session_close(s);
    free(s);
}

// --- Header decoding ---
//
// Read the [4-byte tag][varint name-len][name] header that backends
// see on every inbound stream. The relay prepended the tag; the
// originating client wrote the (name-len, name) part.

static int read_backend_header(pigeon_listener *l,
                               pigeon_stream_handle *handle,
                               uint32_t *out_tag,
                               char *out_name, size_t out_name_cap,
                               size_t *out_name_len)
{
    uint8_t hdr[PIGEON_MAX_STREAM_HEADER];
    size_t  hdr_len = 0;
    if (l->transport.recv_on_stream(l->transport.userdata, handle,
                                    hdr, sizeof(hdr), &hdr_len) != 0) {
        return -1;
    }
    return pigeon_wire_stream_header_decode_backend(hdr, hdr_len,
                                               out_tag,
                                               out_name, out_name_cap,
                                               out_name_len);
}

// --- Queueing a sub-stream into a session ---

static void queue_substream(pigeon_session *s,
                            pigeon_stream_handle *handle,
                            const char *name)
{
    for (size_t i = 0; i < PIGEON_SESSION_MAX_INCOMING; i++) {
        if (!s->incoming[i].in_use) {
            s->incoming[i].in_use = true;
            s->incoming[i].handle = handle;
            size_t nl = strlen(name);
            if (nl >= PIGEON_MAX_NAME_LEN) nl = PIGEON_MAX_NAME_LEN - 1;
            memcpy(s->incoming[i].name, name, nl);
            s->incoming[i].name[nl] = '\0';
            return;
        }
    }
    // Queue full: drop the stream (close it so the peer doesn't
    // wedge waiting for a reader). Matches "warn and drop" behaviour
    // the Go side falls back to under sustained queue pressure.
}

// --- Public API ---

int pigeon_listener_init(pigeon_listener **out,
                         const pigeon_transport *transport,
                         const char *instance_id,
                         pigeon_resolve_device_fn pairing,
                         void *pairing_userdata,
                         const pigeon_dgchannel_def *datagrams,
                         size_t datagram_count)
{
    if (!out || !transport || !pairing) return -1;
    if (datagram_count > PIGEON_MAX_DATAGRAM_CHANNELS) return -1;
    if (!transport->accept_stream || !transport->recv_on_stream) return -1;

    pigeon_listener *l = (pigeon_listener *)calloc(1, sizeof(*l));
    if (!l) return -1;

    l->transport         = *transport;
    l->resolve           = pairing;
    l->resolve_userdata  = pairing_userdata;
    if (instance_id) {
        size_t n = strlen(instance_id);
        if (n >= sizeof(l->instance_id)) n = sizeof(l->instance_id) - 1;
        memcpy(l->instance_id, instance_id, n);
        l->instance_id[n] = '\0';
    }

    // Validate + copy datagram channels with the same rules as
    // pigeon_session_init: no duplicate ids, no id==0, names bounded.
    for (size_t i = 0; i < datagram_count; i++) {
        if (datagrams[i].channel_id == 0) { free(l); return -1; }
        size_t nl = strlen(datagrams[i].name);
        if (nl == 0 || nl >= PIGEON_MAX_NAME_LEN) { free(l); return -1; }
        for (size_t j = 0; j < i; j++) {
            if (datagrams[j].channel_id == datagrams[i].channel_id) {
                free(l); return -1;
            }
        }
        l->datagrams[i] = datagrams[i];
    }
    l->datagram_count = datagram_count;

    *out = l;
    return 0;
}

const char *pigeon_listener_instance_id(const pigeon_listener *l)
{
    if (!l) return NULL;
    return l->instance_id;
}

int pigeon_listener_step(pigeon_listener *l, pigeon_session **out_session)
{
    if (!l || !out_session) return -1;
    if (l->closed) return -1;

    *out_session = NULL;

    pigeon_stream_handle *handle = NULL;
    if (l->transport.accept_stream(l->transport.userdata, &handle) != 0) {
        return -1;
    }

    uint32_t tag = 0;
    char     name[PIGEON_MAX_NAME_LEN] = {0};
    size_t   name_len = 0;
    if (read_backend_header(l, handle, &tag, name, sizeof(name), &name_len) < 0) {
        if (l->transport.close_stream) {
            l->transport.close_stream(l->transport.userdata, handle);
        }
        return 0; // malformed header from one client shouldn't sink the
                  // whole listener — caller can loop again.
    }

    listener_slot *existing = slot_lookup(l, tag);
    if (existing != NULL) {
        if (name_len == 0) {
            // Duplicate primary for a tag we already activated:
            // protocol error, drop.
            if (l->transport.close_stream) {
                l->transport.close_stream(l->transport.userdata, handle);
            }
            return 0;
        }
        queue_substream(existing->session, handle, name);
        return 0;
    }

    if (name_len != 0) {
        // First-seen tag with a named sub-stream — no Session to
        // route to. Drop the stream.
        if (l->transport.close_stream) {
            l->transport.close_stream(l->transport.userdata, handle);
        }
        return 0;
    }

    // New client primary: run the activation handshake on this
    // stream against the listener's resolver.
    pigeon_backend_machine machine;
    char     device_id[PIGEON_AUTH_MAX_DEVICE_ID + 1] = {0};
    pigeon_pairing_record  record;
    int rc = pigeon_run_backend_activation(&l->transport, handle,
                                           l->resolve,
                                           l->resolve_userdata,
                                           &machine,
                                           device_id, sizeof(device_id),
                                           &record);
    if (rc != 0) {
        // -1 (wire failure) or 1 (decoded but rejected): close
        // the primary and keep accepting. Matches the Go-side
        // behaviour: rejected clients don't materialise a Session.
        if (l->transport.close_stream) {
            l->transport.close_stream(l->transport.userdata, handle);
        }
        return 0;
    }

    // Allocate the session, derive the AEAD channel, register it
    // in the demux table BEFORE returning so any sub-stream the
    // client opens immediately after activation finds the tag in
    // the map when the next pump iteration lands.
    listener_slot *slot = slot_insert(l, tag);
    if (!slot) {
        // Table full: cap reached.
        if (l->transport.close_stream) {
            l->transport.close_stream(l->transport.userdata, handle);
        }
        return -1;
    }
    pigeon_session *sess = make_session(l, tag, &record);
    if (!sess) {
        if (l->transport.close_stream) {
            l->transport.close_stream(l->transport.userdata, handle);
        }
        return -1;
    }
    slot->in_use     = true;
    slot->client_tag = tag;
    slot->session    = sess;

    // Bind (don't close) the activation primary. Matches Go's
    // Listener.acceptPrimary: the primary is kept open so the relay
    // bridge's per-client primary forwarding goroutine stays alive
    // — closing it here tears the whole client connection down on
    // the relay side. The primary stream is owned by the transport;
    // pigeon_listener_close drops the transport which closes all
    // its streams as a unit.
    sess->primary = handle;

    *out_session = sess;
    return 1;
}

int pigeon_listener_accept(pigeon_listener *l, pigeon_session **out_session)
{
    if (!l || !out_session) return -1;
    while (!l->closed) {
        pigeon_session *s = NULL;
        int rc = pigeon_listener_step(l, &s);
        if (rc < 0) return rc;
        if (rc == 1) {
            *out_session = s;
            return 0;
        }
        // rc == 0: sub-stream dispatched or malformed header — keep pumping.
    }
    return -1;
}

void pigeon_listener_close(pigeon_listener *l)
{
    if (!l) return;
    l->closed = true;
    for (size_t i = 0; i < PIGEON_LISTENER_MAX_CLIENTS; i++) {
        if (l->slots[i].in_use) {
            free_session(l, l->slots[i].session);
            l->slots[i].in_use = false;
            l->slots[i].session = NULL;
        }
    }
    if (l->owned_transport != NULL) {
        if (l->owned_transport_close != NULL) {
            l->owned_transport_close(l->owned_transport);
        }
        if (l->owned_transport_free != NULL) {
            l->owned_transport_free(l->owned_transport);
        }
        l->owned_transport = NULL;
    }
    free(l);
}

// Internal hook for pigeon_register (and any other transport-owning
// wrapper) to park the underlying transport so pigeon_listener_close
// tears it down. Not declared in the public header — wrappers in this
// library forward-declare it.
void pigeon_listener_set_owned_transport(
        pigeon_listener *l,
        void *transport,
        void (*close_fn)(void *),
        void (*free_fn)(void *))
{
    if (!l) return;
    l->owned_transport       = transport;
    l->owned_transport_close = close_fn;
    l->owned_transport_free  = free_fn;
}

// --- Session-level: drain a buffered incoming sub-stream ---

int pigeon_session_accept_incoming_stream(pigeon_session *s,
                                          const char *name,
                                          pigeon_stream *out_stream)
{
    if (!s || !name || !out_stream) return -1;
    for (size_t i = 0; i < PIGEON_SESSION_MAX_INCOMING; i++) {
        if (s->incoming[i].in_use
                && strcmp(s->incoming[i].name, name) == 0) {
            out_stream->session = s;
            out_stream->handle  = s->incoming[i].handle;
            size_t nl = strlen(name);
            if (nl >= PIGEON_MAX_NAME_LEN) nl = PIGEON_MAX_NAME_LEN - 1;
            memcpy(out_stream->name, name, nl);
            out_stream->name[nl] = '\0';
            s->incoming[i].in_use = false;
            s->incoming[i].handle = NULL;
            return 0;
        }
    }
    return -1;
}

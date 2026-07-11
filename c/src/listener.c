// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0
//
// Multi-client listener (T45 remote-Listen L1). Mirrors Go's
// pigeon.Register / pigeon.Listener.Accept in api.go.
//
// Under T45 each accepted client rides its OWN end-to-end QUIC pipe:
// the backend holds one register control connection open for the
// instance's lifetime, then dials a fresh `listen` connection per
// client. The relay matches each arriving client to a parked listen
// and bridges the two connections opaquely, so there is no shared
// connection and no per-client clientTag demux.
//
// The listener does not own the per-client transports directly. It
// holds a dial-listen callback that produces a fresh bridged listen
// transport on each accept (ngtcp2 dialer in production,
// loopback dialer in tests). pigeon_listener_accept:
//
//   1. dials a listen connection and waits for the relay to bridge a
//      client onto it;
//   2. runs T32.1's pigeon_run_backend_activation on the bridged
//      primary against the registered pairing callback;
//   3. derives an AEAD channel from the resolved PairingRecord and
//      builds a pigeon_session that ADOPTS the listen transport
//      (closing it on pigeon_session_close).
//
// Memory: each accepted session owns its listen transport; the caller
// owns the session and releases it with pigeon_session_close. The
// register control transport is owned by the listener and torn down by
// pigeon_listener_close.

#include "pigeon/pigeon.h"
#include "pigeon/activation.h"
#include "pigeon/session_gen.h"

#include <stdlib.h>
#include <string.h>

// --- Per-listener state ---

struct pigeon_listener {
    // Register control transport. Kept open for the instance's
    // lifetime; no traffic flows on it.
    pigeon_transport          control;
    char                      instance_id[64];

    // Listen-dialer: produces a fresh bridged listen transport on each
    // accept. owner_close / owner_free tear down the adopted transport
    // when its session closes.
    pigeon_listen_dialer      dial_listen;
    void                     *dial_userdata;
    void                    (*dial_userdata_free)(void *); // optional
    pigeon_listen_owner_fn    owner_close;
    pigeon_listen_owner_fn    owner_free;

    // Pairing callback (device-id -> PairingRecord lookup) wired in
    // at init time. Invoked synchronously from the accept path.
    pigeon_resolve_device_fn  resolve;
    void                     *resolve_userdata;

    // Datagram channel definitions copied into each accepted session.
    pigeon_dgchannel_def      datagrams[PIGEON_MAX_DATAGRAM_CHANNELS];
    size_t                    datagram_count;

    // Optional owned register-control closer. When the high-level
    // pigeon_register wires up an ngtcp2 control transport on behalf of
    // the caller, it parks a closer + free callback here so
    // pigeon_listener_close tears the control connection down. NULL
    // when the caller owns the control transport (the low-level
    // pigeon_listener_init path).
    void  (*owned_control_close)(void *t);
    void  (*owned_control_free)(void *t);
    void   *owned_control;

    bool                      closed;
};

// --- Session lifecycle ---

// Allocate and initialise a pigeon_session for an activated client.
// The session ADOPTS `transport` (a fresh listen connection) and its
// owner cookie / teardown hooks. Returns NULL on failure.
static pigeon_session *make_session(pigeon_listener *l,
                                    const pigeon_transport *transport,
                                    pigeon_stream_handle *primary,
                                    void *owner,
                                    const pigeon_pairing_record *rec,
                                    const uint8_t *nonce)
{
    pigeon_session *s = (pigeon_session *)calloc(1, sizeof(*s));
    if (!s) return NULL;

    // Derive the session AEAD channel from the resolved PairingRecord.
    // Mirrors api.go's DeriveSessionChannel on the Listener side:
    // send=backend->client, recv=client->backend. The per-session nonce
    // (from the client's auth_request) is folded into the HKDF info so
    // concurrent sessions under one record derive distinct keys (🎯T44.1):
    // info = direction-label || nonce.
    uint8_t send_info[sizeof("backend->client") - 1 + PIGEON_AUTH_NONCE_LEN];
    uint8_t recv_info[sizeof("client->backend") - 1 + PIGEON_AUTH_NONCE_LEN];
    memcpy(send_info, "backend->client", sizeof("backend->client") - 1);
    memcpy(send_info + sizeof("backend->client") - 1, nonce, PIGEON_AUTH_NONCE_LEN);
    memcpy(recv_info, "client->backend", sizeof("client->backend") - 1);
    memcpy(recv_info + sizeof("client->backend") - 1, nonce, PIGEON_AUTH_NONCE_LEN);
    uint8_t send_key[32], recv_key[32];
    if (pigeon_derive_session_key(rec->local_private_key,
                                  rec->peer_public_key,
                                  send_info, sizeof(send_info),
                                  send_key) != 0) {
        free(s);
        return NULL;
    }
    if (pigeon_derive_session_key(rec->local_private_key,
                                  rec->peer_public_key,
                                  recv_info, sizeof(recv_info),
                                  recv_key) != 0) {
        free(s);
        return NULL;
    }
    pigeon_channel ch;
    pigeon_channel_init(&ch, send_key, recv_key, PIGEON_MODE_STRICT);

    if (pigeon_session_init(s, transport, &ch,
                            l->datagrams, l->datagram_count) != 0) {
        free(s);
        return NULL;
    }
    s->primary      = primary;
    s->owner        = owner;
    s->owner_close  = l->owner_close;
    s->owner_free   = l->owner_free;
    return s;
}

// --- Public API ---

int pigeon_listener_init(pigeon_listener **out,
                         const pigeon_transport *control,
                         const char *instance_id,
                         pigeon_listen_dialer dial_listen,
                         void *dial_userdata,
                         pigeon_listen_owner_fn owner_close,
                         pigeon_listen_owner_fn owner_free,
                         pigeon_resolve_device_fn pairing,
                         void *pairing_userdata,
                         const pigeon_dgchannel_def *datagrams,
                         size_t datagram_count)
{
    if (!out || !control || !dial_listen || !pairing) return -1;
    if (datagram_count > PIGEON_MAX_DATAGRAM_CHANNELS) return -1;

    pigeon_listener *l = (pigeon_listener *)calloc(1, sizeof(*l));
    if (!l) return -1;

    l->control          = *control;
    l->dial_listen      = dial_listen;
    l->dial_userdata    = dial_userdata;
    l->owner_close      = owner_close;
    l->owner_free       = owner_free;
    l->resolve          = pairing;
    l->resolve_userdata = pairing_userdata;
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

int pigeon_listener_accept(pigeon_listener *l, pigeon_session **out_session)
{
    if (!l || !out_session) return -1;
    if (l->closed) return -1;

    *out_session = NULL;

    // 1. Dial a fresh listen connection and wait for the relay to
    //    bridge a client onto it.
    pigeon_transport      tr;
    pigeon_stream_handle *primary = NULL;
    void                 *owner   = NULL;
    if (l->dial_listen(l->dial_userdata, &tr, &primary, &owner) != 0) {
        return -1;
    }

    // 2. Run the activation handshake on the bridged primary against
    //    the listener's resolver. The client's auth_request is the
    //    first message on the bridged pipe, so it doubles as the
    //    "a client really matched" gate.
    pigeon_backend_machine machine;
    char     device_id[PIGEON_AUTH_MAX_DEVICE_ID + 1] = {0};
    pigeon_pairing_record  record;
    uint8_t  session_nonce[PIGEON_AUTH_NONCE_LEN];
    int rc = pigeon_run_backend_activation(&tr, primary,
                                           l->resolve,
                                           l->resolve_userdata,
                                           &machine,
                                           device_id, sizeof(device_id),
                                           &record,
                                           session_nonce,
                                           /*out_route=*/NULL, 0);
    if (rc != 0) {
        // -1 (wire failure) or 1 (decoded but rejected): tear down the
        // listen connection. Matches the Go-side behaviour: rejected
        // clients don't materialise a Session.
        if (l->owner_close && owner) l->owner_close(owner);
        if (l->owner_free  && owner) l->owner_free(owner);
        return -1;
    }

    // 3. Build the session, deriving its AEAD channel and adopting the
    //    listen transport (and its owner cookie).
    pigeon_session *sess = make_session(l, &tr, primary, owner, &record, session_nonce);
    if (!sess) {
        if (l->owner_close && owner) l->owner_close(owner);
        if (l->owner_free  && owner) l->owner_free(owner);
        return -1;
    }

    *out_session = sess;
    return 0;
}

void pigeon_listener_close(pigeon_listener *l)
{
    if (!l) return;
    l->closed = true;
    if (l->owned_control != NULL) {
        if (l->owned_control_close != NULL) {
            l->owned_control_close(l->owned_control);
        }
        if (l->owned_control_free != NULL) {
            l->owned_control_free(l->owned_control);
        }
        l->owned_control = NULL;
    }
    if (l->dial_userdata_free != NULL && l->dial_userdata != NULL) {
        l->dial_userdata_free(l->dial_userdata);
        l->dial_userdata = NULL;
    }
    free(l);
}

// Internal hook for pigeon_register (and any other control-transport-
// owning wrapper) to park the register control transport so
// pigeon_listener_close tears it down. Not declared in the public
// header — wrappers in this library forward-declare it.
void pigeon_listener_set_owned_control(
        pigeon_listener *l,
        void *transport,
        void (*close_fn)(void *),
        void (*free_fn)(void *))
{
    if (!l) return;
    l->owned_control       = transport;
    l->owned_control_close = close_fn;
    l->owned_control_free  = free_fn;
}

// Internal hook: park a free callback for the listen-dialer's userdata
// so pigeon_listener_close releases it. Not in the public header.
void pigeon_listener_set_dial_userdata_free(
        pigeon_listener *l,
        void (*free_fn)(void *))
{
    if (!l) return;
    l->dial_userdata_free = free_fn;
}

// --- Session-level: accept a peer-opened sub-stream by name ---
//
// Each session owns its own QUIC pipe, so this accepts directly from
// the session's transport. Streams whose name doesn't match are
// buffered into the session's incoming queue for a later matching call.

int pigeon_session_accept_incoming_stream(pigeon_session *s,
                                          const char *name,
                                          pigeon_stream *out_stream)
{
    if (!s || !name || !out_stream) return -1;

    // First, drain any previously-buffered sub-stream that matches.
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
            return pigeon_stream_bind_aead(out_stream);
        }
    }

    if (!s->transport.accept_stream || !s->transport.recv_on_stream) {
        return -1;
    }

    // Pull streams off this session's transport until the requested
    // name arrives. Non-matching streams are buffered.
    for (;;) {
        pigeon_stream_handle *handle = NULL;
        if (s->transport.accept_stream(s->transport.userdata, &handle) != 0) {
            return -1;
        }
        uint8_t hdr[PIGEON_MAX_STREAM_HEADER];
        size_t  hdr_len = 0;
        if (s->transport.recv_on_stream(s->transport.userdata, handle,
                                        hdr, sizeof(hdr), &hdr_len) != 0) {
            if (s->transport.close_stream) {
                s->transport.close_stream(s->transport.userdata, handle);
            }
            continue;
        }
        char   sname[PIGEON_MAX_NAME_LEN] = {0};
        size_t sname_len = 0;
        if (pigeon_wire_stream_header_decode(hdr, hdr_len,
                                             sname, sizeof(sname),
                                             &sname_len) < 0) {
            if (s->transport.close_stream) {
                s->transport.close_stream(s->transport.userdata, handle);
            }
            continue;
        }
        if (strcmp(sname, name) == 0) {
            out_stream->session = s;
            out_stream->handle  = handle;
            memcpy(out_stream->name, sname, sname_len);
            out_stream->name[sname_len] = '\0';
            return pigeon_stream_bind_aead(out_stream);
        }
        // Different name: buffer it for a later matching call.
        bool buffered = false;
        for (size_t i = 0; i < PIGEON_SESSION_MAX_INCOMING; i++) {
            if (!s->incoming[i].in_use) {
                s->incoming[i].in_use = true;
                s->incoming[i].handle = handle;
                memcpy(s->incoming[i].name, sname, sname_len);
                s->incoming[i].name[sname_len] = '\0';
                buffered = true;
                break;
            }
        }
        if (!buffered) {
            // Queue full: drop the stream so the peer doesn't wedge.
            if (s->transport.close_stream) {
                s->transport.close_stream(s->transport.userdata, handle);
            }
        }
    }
}

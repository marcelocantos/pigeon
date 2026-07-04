// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

#include "pigeon/pigeon.h"
#include "pigeon/activation.h"
#include "pigeon/session_gen.h"
#include "pigeon/wire_gen.h"
#include <stdlib.h>
#include <string.h>

void pigeon_init(pigeon_ctx *ctx, const pigeon_transport *transport)
{
    memset(ctx, 0, sizeof(*ctx));
    if (transport) {
        ctx->transport = *transport;
    }
}

int pigeon_send(pigeon_ctx *ctx, const uint8_t *data, size_t len)
{
    if (!ctx->transport.send_stream) return -1;
    if (len > PIGEON_MAX_MSG) return -1;

    const uint8_t *payload = data;
    size_t payload_len = len;

    if (ctx->stream_channel.established) {
        // Encrypt into write_buf+4, leaving room for the 4-byte length prefix.
        // Maximum ciphertext: 8 (seq) + len + 16 (tag).
        if (8 + len + 16 + 4 > sizeof(ctx->write_buf)) return -1;
        int ct_len = pigeon_channel_encrypt(&ctx->stream_channel,
                                            data, len,
                                            ctx->write_buf + 4,
                                            sizeof(ctx->write_buf) - 4);
        if (ct_len < 0) return -1;
        // Write the 4-byte BE length prefix in front of the ciphertext.
        uint32_t ulen = (uint32_t)ct_len;
        ctx->write_buf[0] = (uint8_t)(ulen >> 24);
        ctx->write_buf[1] = (uint8_t)(ulen >> 16);
        ctx->write_buf[2] = (uint8_t)(ulen >> 8);
        ctx->write_buf[3] = (uint8_t)(ulen);
        return ctx->transport.send_stream(ctx->transport.userdata,
                                          ctx->write_buf, (size_t)(4 + ct_len));
    }

    // No encryption: frame plaintext.
    int frame_len = pigeon_frame_message(payload, payload_len,
                                         ctx->write_buf, sizeof(ctx->write_buf));
    if (frame_len < 0) return -1;

    return ctx->transport.send_stream(ctx->transport.userdata,
                                      ctx->write_buf, (size_t)frame_len);
}

int pigeon_recv(pigeon_ctx *ctx, uint8_t *out, size_t out_len)
{
    if (!ctx->transport.recv_stream) return -1;

    // Read 4-byte length prefix.
    uint8_t hdr[4];
    size_t got = 0;
    int err = ctx->transport.recv_stream(ctx->transport.userdata, hdr, 4, &got);
    if (err || got != 4) return -1;

    uint32_t frame_payload_len = pigeon_read_frame_length(hdr);
    if (frame_payload_len > PIGEON_MAX_MSG) return -1;

    if (ctx->stream_channel.established) {
        // Read ciphertext into read_buf, then decrypt into out.
        if (frame_payload_len > sizeof(ctx->read_buf)) return -1;
        got = 0;
        err = ctx->transport.recv_stream(ctx->transport.userdata,
                                         ctx->read_buf, frame_payload_len, &got);
        if (err || got != frame_payload_len) return -1;
        return pigeon_channel_decrypt(&ctx->stream_channel,
                                      ctx->read_buf, frame_payload_len,
                                      out, out_len);
    }

    // No encryption: read plaintext directly into out.
    if (frame_payload_len > out_len) return -1;
    got = 0;
    err = ctx->transport.recv_stream(ctx->transport.userdata, out, frame_payload_len, &got);
    if (err || got != frame_payload_len) return -1;

    return (int)frame_payload_len;
}

int pigeon_send_datagram(pigeon_ctx *ctx, const uint8_t *data, size_t len)
{
    if (!ctx->transport.send_datagram) return -1;

    if (ctx->datagram_channel.established) {
        if (8 + len + 16 > sizeof(ctx->write_buf)) return -1;
        int ct_len = pigeon_channel_encrypt(&ctx->datagram_channel,
                                            data, len,
                                            ctx->write_buf,
                                            sizeof(ctx->write_buf));
        if (ct_len < 0) return -1;
        return ctx->transport.send_datagram(ctx->transport.userdata,
                                            ctx->write_buf, (size_t)ct_len);
    }

    return ctx->transport.send_datagram(ctx->transport.userdata, data, len);
}

int pigeon_recv_datagram(pigeon_ctx *ctx, uint8_t *out, size_t out_len)
{
    if (!ctx->transport.recv_datagram) return -1;

    if (ctx->datagram_channel.established) {
        // Receive into read_buf, then decrypt into out.
        size_t got = 0;
        int err = ctx->transport.recv_datagram(ctx->transport.userdata,
                                               ctx->read_buf, sizeof(ctx->read_buf), &got);
        if (err) return -1;
        return pigeon_channel_decrypt(&ctx->datagram_channel,
                                      ctx->read_buf, got,
                                      out, out_len);
    }

    size_t got = 0;
    int err = ctx->transport.recv_datagram(ctx->transport.userdata, out, out_len, &got);
    if (err) return -1;
    return (int)got;
}

int pigeon_frame_message(const uint8_t *payload, size_t len,
                         uint8_t *buf, size_t buf_len)
{
    if (4 + len > buf_len) return -1;
    buf[0] = (uint8_t)(len >> 24);
    buf[1] = (uint8_t)(len >> 16);
    buf[2] = (uint8_t)(len >> 8);
    buf[3] = (uint8_t)(len);
    memcpy(buf + 4, payload, len);
    return (int)(4 + len);
}

uint32_t pigeon_read_frame_length(const uint8_t *buf)
{
    return ((uint32_t)buf[0] << 24) |
           ((uint32_t)buf[1] << 16) |
           ((uint32_t)buf[2] << 8)  |
           ((uint32_t)buf[3]);
}

// --- Multi-channel wire helpers ---

int pigeon_uvarint_encode(uint64_t v, uint8_t *buf, size_t buf_len)
{
    size_t i = 0;
    while (v >= 0x80) {
        if (i >= buf_len) return -1;
        buf[i++] = (uint8_t)(v) | 0x80u;
        v >>= 7;
    }
    if (i >= buf_len) return -1;
    buf[i++] = (uint8_t)(v);
    return (int)i;
}

int pigeon_uvarint_decode(const uint8_t *buf, size_t buf_len, uint64_t *out)
{
    uint64_t v = 0;
    unsigned shift = 0;
    for (size_t i = 0; i < buf_len; i++) {
        uint8_t b = buf[i];
        if (i == 9 && b > 1) {
            // 10th byte may only contribute 1 bit (uint64 max).
            return -1;
        }
        v |= (uint64_t)(b & 0x7fu) << shift;
        if ((b & 0x80u) == 0) {
            *out = v;
            return (int)(i + 1);
        }
        shift += 7;
        if (shift >= 64) return -1;
    }
    return 0; // truncated
}

// Stream-header encoder/decoder are protogen-generated in
// c/src/wire_gen.c — pigeon_wire_stream_header_{encode,decode}. Under
// T45's remote-Listen L1 model each accepted client rides its own
// end-to-end QUIC pipe, so the header is just [varint name-len][name]
// with no backend/client variant and no 4-byte clientTag prefix. Call
// sites use the generated names directly; we no longer hand-roll them
// here.

// AEAD-wrapped datagram: wire = AEAD(plaintext). The *plaintext* layout
// is protogen-generated (pigeon_wire_datagram_plaintext_encode); this
// wrapper layers AEAD on top. Under T45 there is no clientTag prefix —
// each session owns its own QUIC pipe, so the pipe is the demux.
int pigeon_encode_datagram(pigeon_channel *ch,
                           uint64_t channel_id,
                           const uint8_t *payload, size_t payload_len,
                           uint8_t *out, size_t out_len)
{
    if (!ch || !ch->established) return -1;
    if (payload_len > PIGEON_MAX_MSG) return -1;

    // Compose the AEAD-plaintext (the protogen byte format) on the
    // heap: a per-call 1 MiB stack would overflow every host runtime
    // (see T38 in the audit log).
    size_t plain_cap = PIGEON_MAX_VARINT_LEN + PIGEON_MAX_MSG;
    uint8_t *plain = (uint8_t *)malloc(plain_cap);
    if (!plain) return -1;
    int plain_len = pigeon_wire_datagram_plaintext_encode(channel_id,
                                                          payload, payload_len,
                                                          plain, plain_cap);
    if (plain_len < 0) { free(plain); return -1; }

    int ct = pigeon_channel_encrypt(ch, plain, (size_t)plain_len,
                                    out, out_len);
    free(plain);
    if (ct < 0) return -1;
    return ct;
}

int pigeon_decode_datagram(pigeon_channel *ch,
                           const uint8_t *wire, size_t wire_len,
                           uint64_t *channel_id,
                           uint8_t *payload_buf, size_t payload_buf_len)
{
    if (!ch || !ch->established) return -1;

    // AEAD-decrypt into a heap scratch buffer, then decode the
    // protogen plaintext format. Heap-allocated for the same reason as
    // pigeon_encode_datagram above (T38).
    uint8_t *plain = (uint8_t *)malloc(PIGEON_MAX_MSG);
    if (!plain) return -1;
    int pn = pigeon_channel_decrypt(ch, wire, wire_len,
                                    plain, PIGEON_MAX_MSG);
    if (pn < 0) { free(plain); return -1; }

    uint64_t cid = 0;
    size_t payload_len = 0;
    int dn = pigeon_wire_datagram_plaintext_decode(plain, (size_t)pn,
                                                   &cid,
                                                   payload_buf, payload_buf_len,
                                                   &payload_len);
    free(plain);
    if (dn < 0) return -1;
    if (channel_id) *channel_id = cid;
    return (int)payload_len;
}

// Magic bytes and version for PairingRecord serialisation.
#define PIGEON_PR_MAGIC0 0x50u /* 'P' */
#define PIGEON_PR_MAGIC1 0x47u /* 'G' */
#define PIGEON_PR_MAGIC2 0x52u /* 'R' */
#define PIGEON_PR_VERSION 1u

int pigeon_pairing_record_serialize(const pigeon_pairing_record *rec,
                                    uint8_t *buf, size_t buf_len)
{
    if (buf_len < PIGEON_PAIRING_RECORD_SIZE) return -1;

    buf[0] = PIGEON_PR_MAGIC0;
    buf[1] = PIGEON_PR_MAGIC1;
    buf[2] = PIGEON_PR_MAGIC2;
    buf[3] = PIGEON_PR_VERSION;

    memcpy(buf +   4, rec->peer_instance_id,  64);
    memcpy(buf +  68, rec->relay_url,         256);
    memcpy(buf + 324, rec->local_private_key,  32);
    memcpy(buf + 356, rec->local_public_key,   32);
    memcpy(buf + 388, rec->peer_public_key,    32);

    return PIGEON_PAIRING_RECORD_SIZE;
}

int pigeon_pairing_record_deserialize(pigeon_pairing_record *rec,
                                      const uint8_t *buf, size_t buf_len)
{
    if (buf_len < PIGEON_PAIRING_RECORD_SIZE) return -1;
    if (buf[0] != PIGEON_PR_MAGIC0 ||
        buf[1] != PIGEON_PR_MAGIC1 ||
        buf[2] != PIGEON_PR_MAGIC2) return -1;
    if (buf[3] != PIGEON_PR_VERSION) return -1;

    memcpy(rec->peer_instance_id,  buf +   4,  64);
    memcpy(rec->relay_url,         buf +  68, 256);
    memcpy(rec->local_private_key, buf + 324,  32);
    memcpy(rec->local_public_key,  buf + 356,  32);
    memcpy(rec->peer_public_key,   buf + 388,  32);

    return PIGEON_PAIRING_RECORD_SIZE;
}

// --- Multi-channel session API ---

// Lazily allocate the per-session scratch buffers. Idempotent.
static int pigeon_session_ensure_scratch(pigeon_session *s)
{
    if (s->scratch_a && s->scratch_b) return 0;
    // Sized for the largest single send/recv: AEAD ciphertext expansion
    // is ~32 bytes (8-byte seq + 16-byte tag + slack). 64 bytes of
    // slack is comfortable.
    size_t sz = PIGEON_MAX_MSG + 64;
    if (!s->scratch_a) s->scratch_a = (uint8_t *)malloc(sz);
    if (!s->scratch_b) s->scratch_b = (uint8_t *)malloc(sz);
    if (!s->scratch_a || !s->scratch_b) {
        free(s->scratch_a); free(s->scratch_b);
        s->scratch_a = s->scratch_b = NULL;
        return -1;
    }
    s->scratch_size = sz;
    return 0;
}

void pigeon_session_close(pigeon_session *s)
{
    if (!s) return;
    // Close any peer-opened sub-streams this session buffered but the
    // application never picked up.
    for (size_t i = 0; i < PIGEON_SESSION_MAX_INCOMING; i++) {
        if (s->incoming[i].in_use && s->incoming[i].handle != NULL
                && s->transport.close_stream != NULL) {
            (void)s->transport.close_stream(s->transport.userdata,
                                            s->incoming[i].handle);
        }
        s->incoming[i].in_use = false;
        s->incoming[i].handle = NULL;
    }
    free(s->scratch_a); s->scratch_a = NULL;
    free(s->scratch_b); s->scratch_b = NULL;
    s->scratch_size = 0;
    // Tear down an adopted listen transport (pigeon_listener_accept).
    // owner_close drops the QUIC connection; owner_free releases the
    // heap box. Run once, then clear so a second close is a no-op.
    if (s->owner != NULL) {
        if (s->owner_close != NULL) s->owner_close(s->owner);
        if (s->owner_free  != NULL) s->owner_free(s->owner);
        s->owner       = NULL;
        s->owner_close = NULL;
        s->owner_free  = NULL;
    }
}

int pigeon_session_init(pigeon_session *s,
                        const pigeon_transport *transport,
                        const pigeon_channel *channel,
                        const pigeon_dgchannel_def *datagrams,
                        size_t datagram_count)
{
    if (!s || !transport || !channel) return -1;
    if (datagram_count > PIGEON_MAX_DATAGRAM_CHANNELS) return -1;

    memset(s, 0, sizeof(*s));
    s->transport  = *transport;
    s->channel    = *channel;

    // Validate the (name, id) list: no duplicate ids, no id == 0
    // (reserved), no name overflow.
    for (size_t i = 0; i < datagram_count; i++) {
        if (datagrams[i].channel_id == 0) return -1;
        size_t nl = strlen(datagrams[i].name);
        if (nl == 0 || nl >= PIGEON_MAX_NAME_LEN) return -1;
        for (size_t j = 0; j < i; j++) {
            if (datagrams[j].channel_id == datagrams[i].channel_id) return -1;
        }
        s->datagrams[i] = datagrams[i];
    }
    s->datagram_count = datagram_count;
    return 0;
}

int pigeon_session_open_stream(pigeon_session *s,
                               const char *name,
                               pigeon_stream *out_stream)
{
    if (!s || !name || !out_stream) return -1;
    size_t name_len = strlen(name);
    if (name_len == 0 || name_len >= PIGEON_MAX_NAME_LEN) return -1;
    if (!s->transport.open_stream || !s->transport.send_on_stream) return -1;

    pigeon_stream_handle *h = NULL;
    if (s->transport.open_stream(s->transport.userdata, &h) != 0) return -1;

    // Compose the unencrypted name-binding header and write it as the
    // first message on the stream. Under T45 the header is just
    // [varint name-len][name] — no clientTag prefix.
    uint8_t hdr[PIGEON_MAX_STREAM_HEADER];
    int hn = pigeon_wire_stream_header_encode(name, name_len, hdr, sizeof(hdr));
    if (hn < 0) {
        if (s->transport.close_stream) s->transport.close_stream(s->transport.userdata, h);
        return -1;
    }
    if (s->transport.send_on_stream(s->transport.userdata, h, hdr, (size_t)hn) != 0) {
        if (s->transport.close_stream) s->transport.close_stream(s->transport.userdata, h);
        return -1;
    }

    out_stream->session = s;
    out_stream->handle  = h;
    memcpy(out_stream->name, name, name_len);
    out_stream->name[name_len] = '\0';
    return 0;
}

int pigeon_session_get_datagram(pigeon_session *s,
                                const char *name,
                                pigeon_datagram *out)
{
    if (!s || !name || !out) return -1;
    for (size_t i = 0; i < s->datagram_count; i++) {
        if (strcmp(s->datagrams[i].name, name) == 0) {
            out->session    = s;
            out->channel_id = s->datagrams[i].channel_id;
            size_t nl = strlen(name);
            memcpy(out->name, name, nl);
            out->name[nl] = '\0';
            return 0;
        }
    }
    return -1;
}

int pigeon_stream_send(pigeon_stream *s,
                       const uint8_t *msg, size_t msg_len)
{
    if (!s || !s->session || !s->handle) return -1;
    pigeon_session *sess = s->session;
    if (!sess->transport.send_on_stream) return -1;

    // Pairing-mode session: channel not yet established. Mirror Go's
    // Stream.Send (api.go) — send plaintext as one length-prefixed message
    // so pigeon_session_primary() callers can drive the pairing ceremony
    // over the primary stream before the AEAD channel exists.
    if (!sess->channel.established) {
        return sess->transport.send_on_stream(sess->transport.userdata,
                                              s->handle, msg, msg_len);
    }

    if (pigeon_session_ensure_scratch(sess) != 0) return -1;

    // AEAD-encrypt the application payload and write the ciphertext as
    // one length-prefixed message on the stream.
    int ctn = pigeon_channel_encrypt(&sess->channel, msg, msg_len,
                                     sess->scratch_a, sess->scratch_size);
    if (ctn < 0) return -1;
    return sess->transport.send_on_stream(sess->transport.userdata, s->handle,
                                          sess->scratch_a, (size_t)ctn);
}

int pigeon_stream_recv(pigeon_stream *s,
                       uint8_t *buf, size_t buf_len)
{
    if (!s || !s->session || !s->handle) return -1;
    pigeon_session *sess = s->session;
    if (!sess->transport.recv_on_stream) return -1;

    // Pairing-mode mirror of pigeon_stream_send: read plaintext directly
    // into the caller's buffer when the channel hasn't been established yet.
    if (!sess->channel.established) {
        size_t got = 0;
        if (sess->transport.recv_on_stream(sess->transport.userdata, s->handle,
                                           buf, buf_len, &got) != 0) {
            return -1;
        }
        return (int)got;
    }

    if (pigeon_session_ensure_scratch(sess) != 0) return -1;

    size_t got = 0;
    if (sess->transport.recv_on_stream(sess->transport.userdata, s->handle,
                                       sess->scratch_a, sess->scratch_size, &got) != 0) {
        return -1;
    }
    return pigeon_channel_decrypt(&sess->channel, sess->scratch_a, got, buf, buf_len);
}

int pigeon_stream_close(pigeon_stream *s)
{
    if (!s || !s->session || !s->handle) return -1;
    if (!s->session->transport.close_stream) return 0;
    return s->session->transport.close_stream(s->session->transport.userdata, s->handle);
}

int pigeon_datagram_send(pigeon_datagram *d,
                         const uint8_t *payload, size_t payload_len)
{
    if (!d || !d->session) return -1;
    pigeon_session *sess = d->session;
    if (!sess->transport.send_datagram) return -1;
    if (pigeon_session_ensure_scratch(sess) != 0) return -1;

    int wn = pigeon_encode_datagram(&sess->channel,
                                    d->channel_id,
                                    payload, payload_len,
                                    sess->scratch_a, sess->scratch_size);
    if (wn < 0) return -1;
    return sess->transport.send_datagram(sess->transport.userdata,
                                         sess->scratch_a, (size_t)wn);
}

int pigeon_datagram_recv(pigeon_datagram *d,
                         uint8_t *buf, size_t buf_len)
{
    if (!d || !d->session) return -1;
    pigeon_session *sess = d->session;
    if (!sess->transport.recv_datagram) return -1;
    if (pigeon_session_ensure_scratch(sess) != 0) return -1;

    size_t got = 0;
    if (sess->transport.recv_datagram(sess->transport.userdata,
                                      sess->scratch_a, sess->scratch_size, &got) != 0) {
        return -1;
    }
    if (got == 0) return -1; // recv timed out with no datagram available
    uint64_t cid = 0;
    int pn = pigeon_decode_datagram(&sess->channel,
                                    sess->scratch_a, got, &cid,
                                    buf, buf_len);
    if (pn < 0) return -1;
    if (cid != d->channel_id) {
        // Datagram belongs to a different channel; the caller should
        // route to a sibling pigeon_datagram. Returning 0 (zero-byte
        // application payload) is ambiguous, so we surface a distinct
        // sentinel: -2.
        return -2;
    }
    return pn;
}

// --- pigeon_session_primary / pigeon_connect (T32.3) ---

int pigeon_session_primary(pigeon_session *s, pigeon_stream *out_stream)
{
    if (!s || !out_stream || !s->primary) return -1;
    out_stream->session = s;
    out_stream->handle  = s->primary;
    out_stream->name[0] = '\0'; // Primary stream has empty name.
    return 0;
}

// derive_session_channel mirrors crypto.PairingRecord.DeriveChannel
// (Go): two HKDF-SHA256 expansions from the X25519 shared secret,
// using `send_info` / `recv_info` as separate KDF context labels.
// Writes 32-byte send/recv keys into out_send/out_recv. Returns 0 on
// success.
static int derive_session_channel(const pigeon_pairing_record *rec,
                                  const uint8_t *send_info, size_t send_info_len,
                                  const uint8_t *recv_info, size_t recv_info_len,
                                  const uint8_t *nonce,
                                  uint8_t *out_send, uint8_t *out_recv)
{
    // Bind the per-session nonce into the HKDF info so concurrent /
    // reconnecting sessions under one PairingRecord derive distinct keys
    // (🎯T44.1). info = direction-label || nonce.
    uint8_t send_full[64], recv_full[64];
    if (send_info_len + PIGEON_AUTH_NONCE_LEN > sizeof(send_full)) return -1;
    if (recv_info_len + PIGEON_AUTH_NONCE_LEN > sizeof(recv_full)) return -1;
    memcpy(send_full, send_info, send_info_len);
    memcpy(send_full + send_info_len, nonce, PIGEON_AUTH_NONCE_LEN);
    memcpy(recv_full, recv_info, recv_info_len);
    memcpy(recv_full + recv_info_len, nonce, PIGEON_AUTH_NONCE_LEN);

    if (pigeon_derive_session_key(rec->local_private_key,
                                  rec->peer_public_key,
                                  send_full, send_info_len + PIGEON_AUTH_NONCE_LEN,
                                  out_send) != 0) return -1;
    if (pigeon_derive_session_key(rec->local_private_key,
                                  rec->peer_public_key,
                                  recv_full, recv_info_len + PIGEON_AUTH_NONCE_LEN,
                                  out_recv) != 0) return -1;
    return 0;
}

int pigeon_connect_on_transport(const pigeon_transport *transport,
                                pigeon_stream_handle *primary_handle,
                                const char *peer_instance_id,
                                const char *device_id,
                                const pigeon_pairing_record *record,
                                const pigeon_dgchannel_def *datagrams,
                                size_t datagram_count,
                                pigeon_session *out_session)
{
    if (!transport || !primary_handle || !out_session) return -1;
    if (!transport->send_on_stream || !transport->recv_on_stream) return -1;

    bool pairing_mode = (record == NULL);
    if (!pairing_mode && (device_id == NULL || device_id[0] == '\0')) {
        // Activation mode requires the client's device id.
        return -1;
    }
    if (peer_instance_id == NULL) return -1;

    // Under T45 the relay's "ok" ack means this QUIC connection is
    // already bridged end-to-end onto a backend listen, so the primary
    // is a clean pipe straight to the backend — no relay-level framing.
    //
    // 1. Pairing mode has no activation handshake; the backend gates
    //    Session creation on a real client match by reading one empty
    //    "arrival marker" message on the primary (see Go's Connect /
    //    Listener.activate). Activation mode skips the marker: its first
    //    primary message IS the auth_request, which already serves as
    //    the gate.
    if (pairing_mode) {
        if (transport->send_on_stream(transport->userdata, primary_handle,
                                      NULL, 0) != 0) {
            return -1;
        }
    }

    // 2. Run the client-side activation handshake (or skip in pairing mode).
    //    On success the AEAD channel is derived from the PairingRecord;
    //    in pairing mode the channel stays unestablished and the caller
    //    drives the ceremony over Session.Primary().
    pigeon_channel channel;
    memset(&channel, 0, sizeof(channel));
    if (!pairing_mode) {
        pigeon_client_machine cm;
        uint8_t session_nonce[PIGEON_AUTH_NONCE_LEN];
        if (pigeon_run_client_activation(transport, primary_handle,
                                         device_id, /*route=*/"", &cm,
                                         NULL, 0, session_nonce) != 0) {
            return -1;
        }
        // We don't retain the machine post-activation (matches Go's
        // newClientSessionMachine, which advances the spec state and is
        // then held by Session for future executor-mediated I/O —
        // deferred work in T39).
        uint8_t send_key[32], recv_key[32];
        // Mirror Go's pigeon.Connect: send=client->backend, recv=backend->client.
        const char send_info[] = "client->backend";
        const char recv_info[] = "backend->client";
        if (derive_session_channel(record,
                                   (const uint8_t *)send_info, sizeof(send_info) - 1,
                                   (const uint8_t *)recv_info, sizeof(recv_info) - 1,
                                   session_nonce,
                                   send_key, recv_key) != 0) {
            return -1;
        }
        pigeon_channel_init(&channel, send_key, recv_key, PIGEON_MODE_STRICT);
    }
    // Note: in pairing mode `channel` stays zero-initialised; channel.
    // established is false, which pigeon_stream_send / _datagram_send
    // already reject. The caller switches to a derived channel later.

    // 3. Initialise the session (client side).
    if (pigeon_session_init(out_session, transport, &channel,
                            datagrams, datagram_count) != 0) {
        return -1;
    }
    out_session->primary = primary_handle;
    (void)peer_instance_id; // Reserved for future peer_id storage (Go's
                            // Session.PeerID); not stored in pigeon_session
                            // today.
    return 0;
}

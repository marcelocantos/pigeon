// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

#include "pigeon/pigeon.h"
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

int pigeon_encode_stream_header(bool is_backend, uint32_t client_tag,
                                const char *name, size_t name_len,
                                uint8_t *out, size_t out_len)
{
    size_t off = 0;
    if (is_backend) {
        if (off + 4 > out_len) return -1;
        out[off++] = (uint8_t)(client_tag >> 24);
        out[off++] = (uint8_t)(client_tag >> 16);
        out[off++] = (uint8_t)(client_tag >> 8);
        out[off++] = (uint8_t)(client_tag);
    }
    int n = pigeon_uvarint_encode((uint64_t)name_len, out + off, out_len - off);
    if (n < 0) return -1;
    off += (size_t)n;
    if (off + name_len > out_len) return -1;
    if (name_len > 0) {
        if (!name) return -1;
        memcpy(out + off, name, name_len);
        off += name_len;
    }
    return (int)off;
}

// Internal: shared body of decode_{backend,client}_stream_header.
// `tag_in` is non-NULL on backend side (consumes 4 leading bytes).
static int decode_stream_header_inner(const uint8_t *buf, size_t buf_len,
                                      uint32_t *tag_out,
                                      char *name_buf, size_t name_buf_len,
                                      size_t *name_len_out)
{
    size_t off = 0;
    if (tag_out) {
        if (buf_len < 4) return -1;
        *tag_out = ((uint32_t)buf[0] << 24)
                 | ((uint32_t)buf[1] << 16)
                 | ((uint32_t)buf[2] <<  8)
                 |  (uint32_t)buf[3];
        off = 4;
    }
    uint64_t name_len = 0;
    int n = pigeon_uvarint_decode(buf + off, buf_len - off, &name_len);
    if (n <= 0) return -1;
    off += (size_t)n;
    if (name_len > buf_len - off) return -1;
    // Reserve one byte for the trailing NUL.
    if (name_len + 1 > name_buf_len) return -1;
    if (name_len > 0) memcpy(name_buf, buf + off, (size_t)name_len);
    name_buf[name_len] = '\0';
    if (name_len_out) *name_len_out = (size_t)name_len;
    off += (size_t)name_len;
    return (int)off;
}

int pigeon_decode_backend_stream_header(const uint8_t *buf, size_t buf_len,
                                        uint32_t *client_tag,
                                        char *name_buf, size_t name_buf_len,
                                        size_t *name_len_out)
{
    if (!client_tag) return -1;
    return decode_stream_header_inner(buf, buf_len, client_tag,
                                      name_buf, name_buf_len, name_len_out);
}

int pigeon_decode_client_stream_header(const uint8_t *buf, size_t buf_len,
                                       char *name_buf, size_t name_buf_len,
                                       size_t *name_len_out)
{
    return decode_stream_header_inner(buf, buf_len, NULL,
                                      name_buf, name_buf_len, name_len_out);
}

int pigeon_encode_datagram(pigeon_channel *ch,
                           bool is_backend, uint32_t client_tag,
                           uint64_t channel_id,
                           const uint8_t *payload, size_t payload_len,
                           uint8_t *out, size_t out_len)
{
    if (!ch || !ch->established) return -1;

    // First compose the AEAD-plaintext: [varint channel-id][payload].
    uint8_t plain[PIGEON_MAX_VARINT_LEN + PIGEON_MAX_MSG];
    if (payload_len > PIGEON_MAX_MSG) return -1;
    int idn = pigeon_uvarint_encode(channel_id, plain, sizeof(plain));
    if (idn < 0) return -1;
    if ((size_t)idn + payload_len > sizeof(plain)) return -1;
    if (payload_len > 0) memcpy(plain + idn, payload, payload_len);
    size_t plain_len = (size_t)idn + payload_len;

    // Wire = (optional 4-byte tag) ++ AEAD(plain).
    size_t off = 0;
    if (is_backend) {
        if (off + 4 > out_len) return -1;
        out[off++] = (uint8_t)(client_tag >> 24);
        out[off++] = (uint8_t)(client_tag >> 16);
        out[off++] = (uint8_t)(client_tag >> 8);
        out[off++] = (uint8_t)(client_tag);
    }
    int ct = pigeon_channel_encrypt(ch, plain, plain_len,
                                    out + off, out_len - off);
    if (ct < 0) return -1;
    return (int)off + ct;
}

int pigeon_decode_datagram(pigeon_channel *ch,
                           bool is_backend,
                           const uint8_t *wire, size_t wire_len,
                           uint32_t *client_tag,
                           uint64_t *channel_id,
                           uint8_t *payload_buf, size_t payload_buf_len)
{
    if (!ch || !ch->established) return -1;

    size_t off = 0;
    if (is_backend) {
        if (wire_len < 4) return -1;
        if (client_tag) {
            *client_tag = ((uint32_t)wire[0] << 24)
                        | ((uint32_t)wire[1] << 16)
                        | ((uint32_t)wire[2] <<  8)
                        |  (uint32_t)wire[3];
        }
        off = 4;
    }

    // AEAD-decrypt into a scratch buffer, then peel the channel-id varint.
    uint8_t plain[PIGEON_MAX_MSG];
    int pn = pigeon_channel_decrypt(ch, wire + off, wire_len - off,
                                    plain, sizeof(plain));
    if (pn < 0) return -1;

    uint64_t cid = 0;
    int idn = pigeon_uvarint_decode(plain, (size_t)pn, &cid);
    if (idn <= 0) return -1;
    if (channel_id) *channel_id = cid;

    size_t payload_len = (size_t)pn - (size_t)idn;
    if (payload_len > payload_buf_len) return -1;
    if (payload_len > 0) memcpy(payload_buf, plain + idn, payload_len);
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

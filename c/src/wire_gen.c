// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0
//
// Code generated from protocol/wireformats.yaml. DO NOT EDIT.

#include "pigeon/wire_gen.h"
#include <string.h>

int pigeon_wire_stream_header_encode(bool is_backend, uint32_t client_tag, const char *name, size_t name_len, uint8_t *out, size_t out_len)
{
    size_t off = 0;
    if (is_backend) {
        if (off + 4 > out_len) return -1;
        out[off++] = (uint8_t)(client_tag >> 24);
        out[off++] = (uint8_t)(client_tag >> 16);
        out[off++] = (uint8_t)(client_tag >> 8);
        out[off++] = (uint8_t)(client_tag);
        { int _n = pigeon_uvarint_encode((uint64_t)name_len, out + off, out_len - off); if (_n < 0) return -1; off += (size_t)_n; }
        if (off + name_len > out_len) return -1;
        if (name_len > 0) { if (!name) return -1; memcpy(out + off, name, name_len); off += name_len; }
    } else {
        { int _n = pigeon_uvarint_encode((uint64_t)name_len, out + off, out_len - off); if (_n < 0) return -1; off += (size_t)_n; }
        if (off + name_len > out_len) return -1;
        if (name_len > 0) { if (!name) return -1; memcpy(out + off, name, name_len); off += name_len; }
    }
    return (int)off;
}

int pigeon_wire_stream_header_decode_backend(const uint8_t *buf, size_t buf_len, uint32_t *client_tag, char *name_buf, size_t name_buf_len, size_t *name_len_out)
{
    size_t off = 0;
    if (buf_len - off < 4) return -1;
    if (client_tag) *client_tag = ((uint32_t)buf[off] << 24) | ((uint32_t)buf[off+1] << 16) | ((uint32_t)buf[off+2] << 8) | (uint32_t)buf[off+3];
    off += 4;
    { uint64_t _len = 0; int _n = pigeon_uvarint_decode(buf + off, buf_len - off, &_len); if (_n <= 0) return -1; off += (size_t)_n;
      if (_len > buf_len - off) return -1;
      if (_len + 1 > name_buf_len) return -1;
      if (_len > 0) memcpy(name_buf, buf + off, (size_t)_len);
      name_buf[_len] = '\0';
      if (name_len_out) *name_len_out = (size_t)_len;
      off += (size_t)_len; }
    return (int)off;
}

int pigeon_wire_stream_header_decode_client(const uint8_t *buf, size_t buf_len, char *name_buf, size_t name_buf_len, size_t *name_len_out)
{
    size_t off = 0;
    { uint64_t _len = 0; int _n = pigeon_uvarint_decode(buf + off, buf_len - off, &_len); if (_n <= 0) return -1; off += (size_t)_n;
      if (_len > buf_len - off) return -1;
      if (_len + 1 > name_buf_len) return -1;
      if (_len > 0) memcpy(name_buf, buf + off, (size_t)_len);
      name_buf[_len] = '\0';
      if (name_len_out) *name_len_out = (size_t)_len;
      off += (size_t)_len; }
    return (int)off;
}

int pigeon_wire_datagram_plaintext_encode(uint64_t channel_id, const uint8_t *payload, size_t payload_len, uint8_t *out, size_t out_len)
{
    size_t off = 0;
    { int _n = pigeon_uvarint_encode(channel_id, out + off, out_len - off); if (_n < 0) return -1; off += (size_t)_n; }
    if (off + payload_len > out_len) return -1;
    if (payload_len > 0) { memcpy(out + off, payload, payload_len); off += payload_len; }
    return (int)off;
}

int pigeon_wire_datagram_plaintext_decode(const uint8_t *buf, size_t buf_len, uint64_t *channel_id, uint8_t *payload_buf, size_t payload_buf_len, size_t *payload_len_out)
{
    size_t off = 0;
    { uint64_t _v = 0; int _n = pigeon_uvarint_decode(buf + off, buf_len - off, &_v); if (_n <= 0) return -1; if (channel_id) *channel_id = _v; off += (size_t)_n; }
    { size_t _n = buf_len - off; if (_n > payload_buf_len) return -1; if (_n > 0) memcpy(payload_buf, buf + off, _n); if (payload_len_out) *payload_len_out = _n; off = buf_len; }
    return (int)off;
}

int pigeon_wire_relay_greeting_encode_connect(const char *instance_id, size_t instance_id_len, uint8_t *out, size_t out_len)
{
    size_t off = 0;
    const size_t prefix_len = 8;
    if (off + prefix_len > out_len) return -1;
    memcpy(out + off, "connect:", prefix_len);
    off += prefix_len;
    if (instance_id && instance_id_len > 0) {
        if (off + instance_id_len > out_len) return -1;
        memcpy(out + off, instance_id, instance_id_len);
        off += instance_id_len;
    }
    return (int)off;
}

int pigeon_wire_relay_greeting_encode_register_mux(const char *token, const char *instance_id, uint8_t *out, size_t out_len)
{
    size_t off = 0;
    const size_t prefix_len = 12;
    if (off + prefix_len > out_len) return -1;
    memcpy(out + off, "register-mux", prefix_len);
    off += prefix_len;
    const char *parts[2];
    size_t part_lens[2];
    bool any_non_empty = false;
    parts[0] = token;
    part_lens[0] = (token ? strlen(token) : 0);
    if (parts[0] && part_lens[0] > 0) any_non_empty = true;
    parts[1] = instance_id;
    part_lens[1] = (instance_id ? strlen(instance_id) : 0);
    if (parts[1] && part_lens[1] > 0) any_non_empty = true;
    if (any_non_empty) {
        for (int i = 0; i < 2; i++) {
            if (off + 1 > out_len) return -1;
            out[off++] = ':';
            if (parts[i] && part_lens[i] > 0) {
                if (off + part_lens[i] > out_len) return -1;
                memcpy(out + off, parts[i], part_lens[i]);
                off += part_lens[i];
            }
        }
    }
    return (int)off;
}

int pigeon_wire_relay_greeting_decode(const uint8_t *buf, size_t buf_len, pigeon_wire_relay_greeting_variant *out_variant, char *instance_id_buf, size_t instance_id_buf_len, size_t *instance_id_len_out, char *token_buf, size_t token_buf_len, size_t *token_len_out)
{
    if (instance_id_len_out) *instance_id_len_out = 0;
    if (token_len_out) *token_len_out = 0;
    if (buf_len >= 12 && memcmp(buf, "register-mux", 12) == 0) {
        *out_variant = RELAY_GREETING_REGISTER_MUX;
        size_t off = 12;
        if (off < buf_len) {
            if (buf[off] != ':') return -1;
            off++;
            size_t part_start = off;
            size_t part_idx = 0;
            const size_t expected = 2;
            for (; off <= buf_len; off++) {
                bool atEnd = (off == buf_len);
                if (atEnd || (part_idx + 1 < expected && buf[off] == ':')) {
                    size_t part_len = off - part_start;
                    switch (part_idx) {
                    case 0:
                        if (part_len + 1 > token_buf_len) return -1;
                        if (part_len > 0) memcpy(token_buf, buf + part_start, part_len);
                        token_buf[part_len] = '\0';
                        if (token_len_out) *token_len_out = part_len;
                        break;
                    case 1:
                        if (part_len + 1 > instance_id_buf_len) return -1;
                        if (part_len > 0) memcpy(instance_id_buf, buf + part_start, part_len);
                        instance_id_buf[part_len] = '\0';
                        if (instance_id_len_out) *instance_id_len_out = part_len;
                        break;
                    }
                    part_idx++;
                    part_start = off + 1;
                    if (atEnd) break;
                }
            }
        }
        return (int)buf_len;
    }
    if (buf_len >= 8 && memcmp(buf, "connect:", 8) == 0) {
        *out_variant = RELAY_GREETING_CONNECT;
        size_t off = 8;
        size_t instance_id_n = buf_len - off;
        if (instance_id_n + 1 > instance_id_buf_len) return -1;
        if (instance_id_n > 0) memcpy(instance_id_buf, buf + off, instance_id_n);
        instance_id_buf[instance_id_n] = '\0';
        if (instance_id_len_out) *instance_id_len_out = instance_id_n;
        off = buf_len;
        return (int)buf_len;
    }
    *out_variant = RELAY_GREETING_UNKNOWN;
    return -1;
}


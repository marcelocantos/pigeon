// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0
//
// Code generated from protocol/wireformats.yaml. DO NOT EDIT.

#ifndef PIGEON_WIRE_GEN_H
#define PIGEON_WIRE_GEN_H

#include <stdbool.h>
#include <stddef.h>
#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

// Provided by pigeon.h / c/src/pigeon.c.
int pigeon_uvarint_encode(uint64_t v, uint8_t *buf, size_t buf_len);
int pigeon_uvarint_decode(const uint8_t *buf, size_t buf_len, uint64_t *out);

// stream_header — First message on every stream — binds the stream to a name.
int pigeon_wire_stream_header_encode(const char *name, size_t name_len, uint8_t *out, size_t out_len);
int pigeon_wire_stream_header_decode(const uint8_t *buf, size_t buf_len, char *name_buf, size_t name_buf_len, size_t *name_len_out);

// datagram_plaintext — Plaintext layout of a datagram before AEAD encryption.
int pigeon_wire_datagram_plaintext_encode(uint64_t channel_id, const uint8_t *payload, size_t payload_len, uint8_t *out, size_t out_len);
int pigeon_wire_datagram_plaintext_decode(const uint8_t *buf, size_t buf_len, uint64_t *channel_id, uint8_t *payload_buf, size_t payload_buf_len, size_t *payload_len_out);

// relay_greeting — Relay-side greeting sent on the primary stream after dial.
typedef enum {
    RELAY_GREETING_CONNECT = 0,
    RELAY_GREETING_REGISTER = 1,
    RELAY_GREETING_LISTEN = 2,
    RELAY_GREETING_UNKNOWN = -1
} pigeon_wire_relay_greeting_variant;

int pigeon_wire_relay_greeting_encode_connect(const char *instance_id, size_t instance_id_len, uint8_t *out, size_t out_len);
int pigeon_wire_relay_greeting_encode_register(const char *token, const char *instance_id, uint8_t *out, size_t out_len);
int pigeon_wire_relay_greeting_encode_listen(const char *token, const char *instance_id, uint8_t *out, size_t out_len);
int pigeon_wire_relay_greeting_decode(const uint8_t *buf, size_t buf_len, pigeon_wire_relay_greeting_variant *out_variant, char *instance_id_buf, size_t instance_id_buf_len, size_t *instance_id_len_out, char *token_buf, size_t token_buf_len, size_t *token_len_out);


#ifdef __cplusplus
}
#endif

#endif // PIGEON_WIRE_GEN_H

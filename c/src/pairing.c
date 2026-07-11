// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Wire-level pairing ceremony driver (C side).
//
// pigeon_pair_acceptor and pigeon_pair_initiator implement the same
// hello/welcome/reveal/confirm exchange that Go's pairing.runAcceptor /
// runInitiator drive over a pigeon.Conn. Messages are JSON with
// base64-encoded byte fields — byte-for-byte compatible with the Go
// encoding/json marshaller.
//
// 🎯T52: the initiator commits to its ephemeral pubkey before reveal
// so a relay MitM cannot grind SAS codes after seeing peer keys:
//   hello  = {commit = SHA256("pigeon-sas-commit"||eph||blind), identity, instance}
//   welcome = {eph_pub, identity, instance}
//   reveal = {eph_pub, blind}
//   confirm = {}
//
// The only external I/O is through pigeon_transport.send_on_stream /
// recv_on_stream, plus pigeon_derive_confirmation_code /
// pigeon_sas_commit / pigeon_sas_commit_verify from crypto.c.

#include "pigeon/pigeon.h"

#include <stdlib.h>
#include <string.h>
#include <stdint.h>

// ---------------------------------------------------------------------------
// Minimal base64 encoder (RFC 4648 standard alphabet, with padding).
// Go's encoding/json uses standard base64 with padding for []byte fields,
// so a 32-byte key encodes to 44 characters.
// ---------------------------------------------------------------------------

static const char b64_table[] =
    "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/";

// b64_encode: encode `src_len` bytes from `src` into `dst` (null-terminated).
// `dst` must be at least ceil(src_len/3)*4 + 1 bytes.
// Returns the number of base64 characters written (not counting NUL).
static int b64_encode(const uint8_t *src, size_t src_len, char *dst, size_t dst_len)
{
    size_t out = 0;
    for (size_t i = 0; i < src_len; ) {
        uint32_t b0 = src[i++];
        uint32_t b1 = (i < src_len) ? src[i++] : 0;
        uint32_t b2 = (i < src_len) ? src[i++] : 0;
        if (out + 4 >= dst_len) return -1;
        dst[out++] = b64_table[(b0 >> 2) & 0x3F];
        dst[out++] = b64_table[((b0 & 0x3) << 4) | ((b1 >> 4) & 0xF)];
        dst[out++] = b64_table[((b1 & 0xF) << 2) | ((b2 >> 6) & 0x3)];
        dst[out++] = b64_table[b2 & 0x3F];
    }
    // Add padding.
    size_t rem = src_len % 3;
    if (rem == 1) { dst[out-2] = '='; dst[out-1] = '='; }
    else if (rem == 2) { dst[out-1] = '='; }
    dst[out] = '\0';
    return (int)out;
}

// b64_decode_char: return the 6-bit value for a base64 character, or -1.
static int b64_decode_char(char c)
{
    if (c >= 'A' && c <= 'Z') return c - 'A';
    if (c >= 'a' && c <= 'z') return c - 'a' + 26;
    if (c >= '0' && c <= '9') return c - '0' + 52;
    if (c == '+') return 62;
    if (c == '/') return 63;
    return -1;
}

// b64_decode: decode null-terminated base64 string `src` into `dst`.
// `dst_cap` must be >= src_len*3/4.
// Returns the number of decoded bytes, or -1 on error.
static int b64_decode(const char *src, size_t src_len,
                      uint8_t *dst, size_t dst_cap)
{
    // Strip padding to find actual encoded groups.
    while (src_len > 0 && src[src_len-1] == '=') src_len--;
    size_t out = 0;
    for (size_t i = 0; i < src_len; ) {
        int v0 = b64_decode_char(src[i++]);
        int v1 = (i < src_len) ? b64_decode_char(src[i++]) : -1;
        int v2 = (i < src_len) ? b64_decode_char(src[i++]) : -1;
        int v3 = (i < src_len) ? b64_decode_char(src[i++]) : -1;
        if (v0 < 0 || v1 < 0) return -1;
        if (out >= dst_cap) return -1;
        dst[out++] = (uint8_t)((v0 << 2) | (v1 >> 4));
        if (v2 >= 0) {
            if (out >= dst_cap) return -1;
            dst[out++] = (uint8_t)((v1 << 4) | (v2 >> 2));
        }
        if (v3 >= 0) {
            if (out >= dst_cap) return -1;
            dst[out++] = (uint8_t)((v2 << 6) | v3);
        }
    }
    return (int)out;
}

// ---------------------------------------------------------------------------
// Minimal JSON helpers for the fixed pairing message schema.
// We never build a general-purpose parser — just extract named string fields
// from a well-formed JSON object produced by Go's encoding/json.
// ---------------------------------------------------------------------------

// json_find_string_field: find the value of a JSON string field by key.
// Writes the raw (unescaped) string content into `out` (NUL-terminated).
// Returns the length of the value, or -1 if not found / too long.
// NOTE: does not handle JSON escapes beyond the printable ASCII subset
// produced by encoding/json for these specific fields (instance IDs and
// base64 strings contain no characters that encoding/json would escape).
static int json_find_string_field(const char *json, size_t json_len,
                                  const char *key,
                                  char *out, size_t out_cap)
{
    // Search for "key": pattern.
    char search[256];
    int slen = 0;
    search[slen++] = '"';
    for (const char *k = key; *k; k++) {
        if (slen + 2 >= (int)sizeof(search)) return -1;
        search[slen++] = *k;
    }
    search[slen++] = '"';
    search[slen] = '\0';

    const char *p = json;
    const char *end = json + json_len;

    while (p < end) {
        // Find the key.
        const char *found = NULL;
        for (const char *q = p; q + slen <= end; q++) {
            if (memcmp(q, search, (size_t)slen) == 0) {
                found = q;
                break;
            }
        }
        if (!found) return -1;
        p = found + slen;
        // Skip whitespace and colon.
        while (p < end && (*p == ' ' || *p == '\t' || *p == '\n' || *p == '\r')) p++;
        if (p >= end || *p != ':') { p = found + 1; continue; }
        p++;
        while (p < end && (*p == ' ' || *p == '\t' || *p == '\n' || *p == '\r')) p++;
        if (p >= end || *p != '"') { p = found + 1; continue; }
        p++; // skip opening quote
        // Copy value until closing quote.
        size_t n = 0;
        while (p < end && *p != '"') {
            if (n + 1 >= out_cap) return -1;
            out[n++] = *p++;
        }
        out[n] = '\0';
        return (int)n;
    }
    return -1;
}

// ---------------------------------------------------------------------------
// send/recv helpers using the transport's send_on_stream / recv_on_stream.
// ---------------------------------------------------------------------------

static int pair_send(const pigeon_transport *t, pigeon_stream_handle *h,
                     const uint8_t *msg, size_t len)
{
    if (!t->send_on_stream) return -1;
    return t->send_on_stream(t->userdata, h, msg, len);
}

static int pair_recv(const pigeon_transport *t, pigeon_stream_handle *h,
                     uint8_t *buf, size_t buf_cap, size_t *out_len)
{
    if (!t->recv_on_stream) return -1;
    return t->recv_on_stream(t->userdata, h, buf, buf_cap, out_len);
}

// ---------------------------------------------------------------------------
// JSON message builders.
//
// Each of these builds the same JSON that Go's encoding/json produces for
// the corresponding pairingMessage struct value, so the C driver is wire-
// compatible with the Go driver.
// ---------------------------------------------------------------------------

// build_hello: {"kind":"hello","commit":"<b64>","identity_pub":"<b64>","instance_id":"<str>"}
// commit is a 32-byte SAS commitment (SHA256); eph is revealed later.
static int build_hello(const uint8_t *commit,
                       const uint8_t *identity_pub,
                       const char *instance_id,
                       uint8_t *buf, size_t buf_cap)
{
    char commit_b64[64], id_b64[64];
    if (b64_encode(commit, 32, commit_b64, sizeof(commit_b64)) < 0) return -1;
    if (b64_encode(identity_pub, 32, id_b64, sizeof(id_b64)) < 0) return -1;
    int n = snprintf((char *)buf, buf_cap,
        "{\"kind\":\"hello\",\"commit\":\"%s\",\"identity_pub\":\"%s\","
        "\"instance_id\":\"%s\"}",
        commit_b64, id_b64, instance_id);
    if (n < 0 || (size_t)n >= buf_cap) return -1;
    return n;
}

// build_welcome: {"kind":"welcome","eph_pub":"<b64>","identity_pub":"<b64>","instance_id":"<str>"}
static int build_welcome(const uint8_t *eph_pub,
                         const uint8_t *identity_pub,
                         const char *instance_id,
                         uint8_t *buf, size_t buf_cap)
{
    char eph_b64[64], id_b64[64];
    if (b64_encode(eph_pub, 32, eph_b64, sizeof(eph_b64)) < 0) return -1;
    if (b64_encode(identity_pub, 32, id_b64, sizeof(id_b64)) < 0) return -1;
    int n = snprintf((char *)buf, buf_cap,
        "{\"kind\":\"welcome\",\"eph_pub\":\"%s\",\"identity_pub\":\"%s\","
        "\"instance_id\":\"%s\"}",
        eph_b64, id_b64, instance_id);
    if (n < 0 || (size_t)n >= buf_cap) return -1;
    return n;
}

// build_reveal: {"kind":"reveal","eph_pub":"<b64>","blind":"<b64>"}
static int build_reveal(const uint8_t *eph_pub,
                        const uint8_t *blind,
                        uint8_t *buf, size_t buf_cap)
{
    char eph_b64[64], blind_b64[64];
    if (b64_encode(eph_pub, 32, eph_b64, sizeof(eph_b64)) < 0) return -1;
    if (b64_encode(blind, 32, blind_b64, sizeof(blind_b64)) < 0) return -1;
    int n = snprintf((char *)buf, buf_cap,
        "{\"kind\":\"reveal\",\"eph_pub\":\"%s\",\"blind\":\"%s\"}",
        eph_b64, blind_b64);
    if (n < 0 || (size_t)n >= buf_cap) return -1;
    return n;
}

// build_confirm: {"kind":"confirm"}
static int build_confirm(uint8_t *buf, size_t buf_cap)
{
    int n = snprintf((char *)buf, buf_cap, "{\"kind\":\"confirm\"}");
    if (n < 0 || (size_t)n >= buf_cap) return -1;
    return n;
}

// ---------------------------------------------------------------------------
// pigeon_pair_acceptor
// ---------------------------------------------------------------------------

int pigeon_pair_acceptor(
    const pigeon_transport *transport,
    const uint8_t *local_eph_priv,
    const uint8_t *local_eph_pub,
    const uint8_t *identity_pub,
    const char *instance_id,
    int (*confirm_fn)(void *userdata, const char *code),
    void *userdata,
    pigeon_pairing_record *out_record,
    char *out_code)
{
    if (!transport || !local_eph_priv || !local_eph_pub ||
        !identity_pub || !instance_id || !confirm_fn ||
        !out_record || !out_code) return -1;
    if (!transport->accept_stream || !transport->send_on_stream || !transport->recv_on_stream)
        return -1;

    uint8_t *buf = (uint8_t *)malloc(PIGEON_MAX_MSG);
    if (!buf) return -1;

    // Accept the stream that the initiator opened.
    pigeon_stream_handle *stream = NULL;
    if (transport->accept_stream(transport->userdata, &stream) != 0) { free(buf); return -1; }

    // --- Drive FSM through setup phases (no wire I/O yet) ---
    pigeon_acceptor_machine m;
    pigeon_acceptor_machine_init(&m);
    // All actions are no-ops: key generation and relay registration happened
    // before this function was called. We just walk the state machine.
    if (pigeon_acceptor_step(&m, PIGEON_PAIRINGCEREMONY_EVENT_PAIR_BEGIN) != 1)        { free(buf); return -1; }
    if (pigeon_acceptor_step(&m, PIGEON_PAIRINGCEREMONY_EVENT_EPHEMERAL_READY) != 1)   { free(buf); return -1; }
    if (pigeon_acceptor_step(&m, PIGEON_PAIRINGCEREMONY_EVENT_RELAY_REGISTERED) != 1)  { free(buf); return -1; }

    // --- Read hello (commit only; eph revealed later) ---
    size_t got = 0;
    if (pair_recv(transport, stream, buf, PIGEON_MAX_MSG, &got) != 0) { free(buf); return -1; }
    buf[got < PIGEON_MAX_MSG ? got : PIGEON_MAX_MSG - 1] = '\0';

    char kind[32];
    if (json_find_string_field((char *)buf, got, "kind", kind, sizeof(kind)) < 0
        || strcmp(kind, "hello") != 0) { free(buf); return -1; }

    char commit_b64[64];
    if (json_find_string_field((char *)buf, got, "commit", commit_b64, sizeof(commit_b64)) < 0)
        { free(buf); return -1; }
    uint8_t peer_commit[32];
    if (b64_decode(commit_b64, strlen(commit_b64), peer_commit, 32) != 32)
        { free(buf); return -1; }

    // Extract initiator identity and instance from hello.
    uint8_t peer_id_pub[32];
    char peer_eph_id_b64[64];
    if (json_find_string_field((char *)buf, got, "identity_pub", peer_eph_id_b64, sizeof(peer_eph_id_b64)) >= 0)
        b64_decode(peer_eph_id_b64, strlen(peer_eph_id_b64), peer_id_pub, 32);
    char peer_instance[64] = "";
    json_find_string_field((char *)buf, got, "instance_id", peer_instance, sizeof(peer_instance));

    if (pigeon_acceptor_handle_message(&m, PIGEON_PAIRINGCEREMONY_MSG_HELLO) != 1) { free(buf); return -1; }

    // --- Send welcome (reveal acceptor eph; already bound by OOB token) ---
    int wlen = build_welcome(local_eph_pub, identity_pub, instance_id, buf, PIGEON_MAX_MSG);
    if (wlen < 0) { free(buf); return -1; }
    if (pair_send(transport, stream, buf, (size_t)wlen) != 0) { free(buf); return -1; }

    // --- Read reveal; verify it opens the hello commit ---
    got = 0;
    if (pair_recv(transport, stream, buf, PIGEON_MAX_MSG, &got) != 0) { free(buf); return -1; }
    buf[got < PIGEON_MAX_MSG ? got : PIGEON_MAX_MSG - 1] = '\0';
    if (json_find_string_field((char *)buf, got, "kind", kind, sizeof(kind)) < 0
        || strcmp(kind, "reveal") != 0) { free(buf); return -1; }

    char eph_b64[64], blind_b64[64];
    if (json_find_string_field((char *)buf, got, "eph_pub", eph_b64, sizeof(eph_b64)) < 0)
        { free(buf); return -1; }
    if (json_find_string_field((char *)buf, got, "blind", blind_b64, sizeof(blind_b64)) < 0)
        { free(buf); return -1; }
    uint8_t peer_eph_pub[32], peer_blind[32];
    if (b64_decode(eph_b64, strlen(eph_b64), peer_eph_pub, 32) != 32)
        { free(buf); return -1; }
    if (b64_decode(blind_b64, strlen(blind_b64), peer_blind, 32) != 32)
        { free(buf); return -1; }

    if (pigeon_sas_commit_verify(peer_eph_pub, peer_blind, peer_commit) != 0) {
        (void)pigeon_acceptor_step(&m, PIGEON_PAIRINGCEREMONY_EVENT_COMMIT_FAIL);
        free(buf);
        return -1;
    }

    if (pigeon_acceptor_handle_message(&m, PIGEON_PAIRINGCEREMONY_MSG_REVEAL) != 1) { free(buf); return -1; }
    if (pigeon_acceptor_step(&m, PIGEON_PAIRINGCEREMONY_EVENT_CODE_READY) != 1)    { free(buf); return -1; }

    // --- Derive confirmation code (only after commit verified) ---
    if (pigeon_derive_confirmation_code(local_eph_pub, peer_eph_pub, out_code) != 0)
        { free(buf); return -1; }

    // --- Ask local user to confirm ---
    if (confirm_fn(userdata, out_code) != 1) { free(buf); return -1; }

    if (pigeon_acceptor_step(&m, PIGEON_PAIRINGCEREMONY_EVENT_USER_CONFIRM) != 1) { free(buf); return -1; }

    // --- Send confirm ---
    int clen = build_confirm(buf, PIGEON_MAX_MSG);
    if (clen < 0) { free(buf); return -1; }
    if (pair_send(transport, stream, buf, (size_t)clen) != 0) { free(buf); return -1; }

    // --- Read peer confirm ---
    got = 0;
    if (pair_recv(transport, stream, buf, PIGEON_MAX_MSG, &got) != 0) { free(buf); return -1; }
    buf[got < PIGEON_MAX_MSG ? got : PIGEON_MAX_MSG - 1] = '\0';
    if (json_find_string_field((char *)buf, got, "kind", kind, sizeof(kind)) < 0
        || strcmp(kind, "confirm") != 0) { free(buf); return -1; }

    if (pigeon_acceptor_handle_message(&m, PIGEON_PAIRINGCEREMONY_MSG_CONFIRM_TO_ACCEPTOR) != 1)
        { free(buf); return -1; }

    // --- Build output record ---
    memset(out_record, 0, sizeof(*out_record));
    strncpy(out_record->peer_instance_id, peer_instance, sizeof(out_record->peer_instance_id) - 1);
    // relay_url is not available to the C driver; caller fills it if needed.
    memcpy(out_record->local_private_key, local_eph_priv, 32);
    memcpy(out_record->local_public_key,  local_eph_pub,  32);
    memcpy(out_record->peer_public_key,   peer_eph_pub,   32);

    free(buf);
    return 0;
}

// ---------------------------------------------------------------------------
// pigeon_pair_initiator
// ---------------------------------------------------------------------------

int pigeon_pair_initiator(
    const pigeon_transport *transport,
    const uint8_t *local_eph_priv,
    const uint8_t *local_eph_pub,
    const uint8_t *identity_pub,
    const char *instance_id,
    const uint8_t *acc_eph_pub,
    const char *acc_instance,
    int (*confirm_fn)(void *userdata, const char *code),
    void *userdata,
    pigeon_pairing_record *out_record,
    char *out_code)
{
    if (!transport || !local_eph_priv || !local_eph_pub ||
        !identity_pub || !instance_id || !acc_eph_pub || !acc_instance ||
        !confirm_fn || !out_record || !out_code) return -1;
    if (!transport->open_stream || !transport->send_on_stream || !transport->recv_on_stream)
        return -1;

    uint8_t *buf = (uint8_t *)malloc(PIGEON_MAX_MSG);
    if (!buf) return -1;

    // Open a stream toward the acceptor.
    pigeon_stream_handle *stream = NULL;
    if (transport->open_stream(transport->userdata, &stream) != 0) { free(buf); return -1; }

    // --- Drive FSM through setup phases ---
    pigeon_initiator_machine m;
    pigeon_initiator_machine_init(&m);
    if (pigeon_initiator_step(&m, PIGEON_PAIRINGCEREMONY_EVENT_TOKEN_RECEIVED) != 1)  { free(buf); return -1; }
    if (pigeon_initiator_step(&m, PIGEON_PAIRINGCEREMONY_EVENT_TOKEN_DECODED) != 1)   { free(buf); return -1; }
    if (pigeon_initiator_step(&m, PIGEON_PAIRINGCEREMONY_EVENT_EPHEMERAL_READY) != 1) { free(buf); return -1; }

    // --- Mint SAS blind + commit before revealing eph (🎯T52) ---
    uint8_t blind[32], commit[32];
    pigeon_random_bytes(blind, 32);
    if (pigeon_sas_commit(local_eph_pub, blind, commit) != 0) { free(buf); return -1; }

    if (pigeon_initiator_step(&m, PIGEON_PAIRINGCEREMONY_EVENT_RELAY_CONNECTED) != 1) { free(buf); return -1; }

    // --- Send hello (commitment only) ---
    int hlen = build_hello(commit, identity_pub, instance_id, buf, PIGEON_MAX_MSG);
    if (hlen < 0) { free(buf); return -1; }
    if (pair_send(transport, stream, buf, (size_t)hlen) != 0) { free(buf); return -1; }

    // --- Read welcome ---
    size_t got = 0;
    if (pair_recv(transport, stream, buf, PIGEON_MAX_MSG, &got) != 0) { free(buf); return -1; }
    buf[got < PIGEON_MAX_MSG ? got : PIGEON_MAX_MSG - 1] = '\0';

    char kind[32];
    if (json_find_string_field((char *)buf, got, "kind", kind, sizeof(kind)) < 0
        || strcmp(kind, "welcome") != 0) { free(buf); return -1; }

    char eph_b64[64];
    if (json_find_string_field((char *)buf, got, "eph_pub", eph_b64, sizeof(eph_b64)) < 0)
        { free(buf); return -1; }
    uint8_t peer_eph_pub[32];
    if (b64_decode(eph_b64, strlen(eph_b64), peer_eph_pub, 32) != 32)
        { free(buf); return -1; }

    // Welcome eph must match the OOB token; otherwise a MitM substituted
    // the acceptor key. Token value is authoritative for SAS + record.
    if (memcmp(peer_eph_pub, acc_eph_pub, 32) != 0) { free(buf); return -1; }

    char peer_instance[64] = "";
    json_find_string_field((char *)buf, got, "instance_id", peer_instance, sizeof(peer_instance));

    if (pigeon_initiator_handle_message(&m, PIGEON_PAIRINGCEREMONY_MSG_WELCOME) != 1) { free(buf); return -1; }

    // --- Reveal eph under the prior commit ---
    int rlen = build_reveal(local_eph_pub, blind, buf, PIGEON_MAX_MSG);
    if (rlen < 0) { free(buf); return -1; }
    if (pair_send(transport, stream, buf, (size_t)rlen) != 0) { free(buf); return -1; }

    if (pigeon_initiator_step(&m, PIGEON_PAIRINGCEREMONY_EVENT_REVEAL_SENT) != 1) { free(buf); return -1; }
    if (pigeon_initiator_step(&m, PIGEON_PAIRINGCEREMONY_EVENT_CODE_READY) != 1)  { free(buf); return -1; }

    // --- Derive confirmation code from token-bound acceptor eph + local eph ---
    if (pigeon_derive_confirmation_code(acc_eph_pub, local_eph_pub, out_code) != 0)
        { free(buf); return -1; }

    // --- Ask local user to confirm ---
    if (confirm_fn(userdata, out_code) != 1) { free(buf); return -1; }

    if (pigeon_initiator_step(&m, PIGEON_PAIRINGCEREMONY_EVENT_USER_CONFIRM) != 1) { free(buf); return -1; }

    // --- Send confirm ---
    int clen = build_confirm(buf, PIGEON_MAX_MSG);
    if (clen < 0) { free(buf); return -1; }
    if (pair_send(transport, stream, buf, (size_t)clen) != 0) { free(buf); return -1; }

    // --- Read peer confirm ---
    got = 0;
    if (pair_recv(transport, stream, buf, PIGEON_MAX_MSG, &got) != 0) { free(buf); return -1; }
    buf[got < PIGEON_MAX_MSG ? got : PIGEON_MAX_MSG - 1] = '\0';
    if (json_find_string_field((char *)buf, got, "kind", kind, sizeof(kind)) < 0
        || strcmp(kind, "confirm") != 0) { free(buf); return -1; }

    if (pigeon_initiator_handle_message(&m, PIGEON_PAIRINGCEREMONY_MSG_CONFIRM_TO_INITIATOR) != 1)
        { free(buf); return -1; }

    // --- Build output record ---
    memset(out_record, 0, sizeof(*out_record));
    strncpy(out_record->peer_instance_id, peer_instance[0] ? peer_instance : acc_instance,
            sizeof(out_record->peer_instance_id) - 1);
    memcpy(out_record->local_private_key, local_eph_priv, 32);
    memcpy(out_record->local_public_key,  local_eph_pub,  32);
    memcpy(out_record->peer_public_key,   acc_eph_pub,    32);

    free(buf);
    return 0;
}

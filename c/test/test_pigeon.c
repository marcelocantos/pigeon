// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

#include "pigeon.h"
#include <assert.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <sodium.h>

static int tests_run = 0;
static int tests_passed = 0;

#define TEST(name) \
    do { \
        tests_run++; \
        printf("  %-50s", name); \
    } while (0)

#define PASS() \
    do { \
        tests_passed++; \
        printf("OK\n"); \
    } while (0)

#define FAIL(msg) \
    do { \
        printf("FAIL: %s\n", msg); \
    } while (0)

// --- Keypair generation ---

static void test_keypair(void)
{
    TEST("keypair generation");
    pigeon_keypair kp;
    int ret = pigeon_generate_keypair(&kp);
    if (ret != 0) { FAIL("generate returned error"); return; }

    // Public key should not be all zeroes.
    uint8_t zero[32] = {0};
    if (memcmp(kp.public_key, zero, 32) == 0) { FAIL("public key is zero"); return; }
    if (memcmp(kp.private_key, zero, 32) == 0) { FAIL("private key is zero"); return; }

    // Two keypairs should differ.
    pigeon_keypair kp2;
    pigeon_generate_keypair(&kp2);
    if (memcmp(kp.private_key, kp2.private_key, 32) == 0) { FAIL("duplicate keys"); return; }

    PASS();
}

// --- Session key derivation (ECDH + HKDF) ---

static void test_session_key_derivation(void)
{
    TEST("session key derivation (ECDH + HKDF)");
    pigeon_keypair alice, bob;
    pigeon_generate_keypair(&alice);
    pigeon_generate_keypair(&bob);

    const uint8_t info[] = "test-session";
    uint8_t alice_key[32], bob_key[32];

    int ret1 = pigeon_derive_session_key(alice.private_key, bob.public_key,
                                          info, sizeof(info) - 1, alice_key);
    int ret2 = pigeon_derive_session_key(bob.private_key, alice.public_key,
                                          info, sizeof(info) - 1, bob_key);
    if (ret1 != 0 || ret2 != 0) { FAIL("derivation error"); return; }

    // Both sides should derive the same key.
    if (memcmp(alice_key, bob_key, 32) != 0) { FAIL("keys don't match"); return; }

    // Different info should produce different keys.
    uint8_t other_key[32];
    const uint8_t info2[] = "other-session";
    pigeon_derive_session_key(alice.private_key, bob.public_key,
                              info2, sizeof(info2) - 1, other_key);
    if (memcmp(alice_key, other_key, 32) == 0) { FAIL("different info same key"); return; }

    PASS();
}

// --- Confirmation code ---

static void test_confirmation_code(void)
{
    TEST("confirmation code (order-independent)");
    pigeon_keypair alice, bob;
    pigeon_generate_keypair(&alice);
    pigeon_generate_keypair(&bob);

    char code_ab[7], code_ba[7];
    int ret1 = pigeon_derive_confirmation_code(alice.public_key, bob.public_key, code_ab);
    int ret2 = pigeon_derive_confirmation_code(bob.public_key, alice.public_key, code_ba);
    if (ret1 != 0 || ret2 != 0) { FAIL("derivation error"); return; }

    // Same code regardless of order.
    if (strcmp(code_ab, code_ba) != 0) { FAIL("codes differ"); return; }

    // 6 digits, null-terminated.
    if (strlen(code_ab) != 6) { FAIL("wrong length"); return; }
    for (int i = 0; i < 6; i++) {
        if (code_ab[i] < '0' || code_ab[i] > '9') { FAIL("non-digit"); return; }
    }

    PASS();
}

// --- Channel encrypt/decrypt round-trip ---

static void test_channel_roundtrip(void)
{
    TEST("channel encrypt/decrypt round-trip");
    pigeon_keypair alice, bob;
    pigeon_generate_keypair(&alice);
    pigeon_generate_keypair(&bob);

    // Derive send/recv keys for alice→bob direction.
    uint8_t a2b[32], b2a[32];
    const uint8_t info_a2b[] = "alice-to-bob";
    const uint8_t info_b2a[] = "bob-to-alice";
    pigeon_derive_session_key(alice.private_key, bob.public_key, info_a2b, sizeof(info_a2b) - 1, a2b);
    pigeon_derive_session_key(alice.private_key, bob.public_key, info_b2a, sizeof(info_b2a) - 1, b2a);

    pigeon_channel alice_ch, bob_ch;
    pigeon_channel_init(&alice_ch, a2b, b2a, PIGEON_MODE_STRICT);
    pigeon_channel_init(&bob_ch, b2a, a2b, PIGEON_MODE_STRICT);

    const char *msg = "hello from pigeon";
    uint8_t ciphertext[256], plaintext[256];

    int ct_len = pigeon_channel_encrypt(&alice_ch, (const uint8_t *)msg, strlen(msg),
                                         ciphertext, sizeof(ciphertext));
    if (ct_len < 0) { FAIL("encrypt failed"); return; }

    // Ciphertext should be longer than plaintext (8-byte seq + 16-byte tag).
    if ((size_t)ct_len != 8 + strlen(msg) + 16) { FAIL("wrong ciphertext length"); return; }

    int pt_len = pigeon_channel_decrypt(&bob_ch, ciphertext, (size_t)ct_len,
                                         plaintext, sizeof(plaintext));
    if (pt_len < 0) { FAIL("decrypt failed"); return; }
    if ((size_t)pt_len != strlen(msg)) { FAIL("wrong plaintext length"); return; }
    if (memcmp(plaintext, msg, (size_t)pt_len) != 0) { FAIL("plaintext mismatch"); return; }

    PASS();
}

// --- Symmetric channel ---

static void test_symmetric_channel(void)
{
    TEST("symmetric channel (client/server)");
    uint8_t master[32];
    randombytes_buf(master, 32);

    pigeon_channel client_ch, server_ch;
    int ret1 = pigeon_channel_init_symmetric(&client_ch, master, false);
    int ret2 = pigeon_channel_init_symmetric(&server_ch, master, true);
    if (ret1 != 0 || ret2 != 0) { FAIL("init error"); return; }

    // Client → server.
    const char *msg = "client says hi";
    uint8_t ct[256], pt[256];
    int ct_len = pigeon_channel_encrypt(&client_ch, (const uint8_t *)msg, strlen(msg), ct, sizeof(ct));
    int pt_len = pigeon_channel_decrypt(&server_ch, ct, (size_t)ct_len, pt, sizeof(pt));
    if (pt_len < 0 || (size_t)pt_len != strlen(msg) || memcmp(pt, msg, (size_t)pt_len) != 0) {
        FAIL("client→server failed"); return;
    }

    // Server → client.
    const char *reply = "server replies";
    ct_len = pigeon_channel_encrypt(&server_ch, (const uint8_t *)reply, strlen(reply), ct, sizeof(ct));
    pt_len = pigeon_channel_decrypt(&client_ch, ct, (size_t)ct_len, pt, sizeof(pt));
    if (pt_len < 0 || (size_t)pt_len != strlen(reply) || memcmp(pt, reply, (size_t)pt_len) != 0) {
        FAIL("server→client failed"); return;
    }

    PASS();
}

// --- Sequence number enforcement ---

static void test_sequence_strict(void)
{
    TEST("strict mode rejects out-of-order");
    uint8_t key[32];
    randombytes_buf(key, 32);

    pigeon_channel send_ch, recv_ch;
    pigeon_channel_init(&send_ch, key, key, PIGEON_MODE_STRICT);
    pigeon_channel_init(&recv_ch, key, key, PIGEON_MODE_STRICT);

    const char *msg = "seq test";
    uint8_t ct1[256], ct2[256], pt[256];

    pigeon_channel_encrypt(&send_ch, (const uint8_t *)msg, strlen(msg), ct1, sizeof(ct1));
    int ct2_len = pigeon_channel_encrypt(&send_ch, (const uint8_t *)msg, strlen(msg), ct2, sizeof(ct2));

    // Decrypt ct2 first (seq=1) — should fail because recv expects seq=0.
    int ret = pigeon_channel_decrypt(&recv_ch, ct2, (size_t)ct2_len, pt, sizeof(pt));
    if (ret >= 0) { FAIL("should have rejected out-of-order"); return; }

    PASS();
}

// --- Wire framing ---

static void test_framing(void)
{
    TEST("wire framing (4-byte BE length prefix)");
    const char *payload = "test payload";
    size_t len = strlen(payload);
    uint8_t buf[256];

    int frame_len = pigeon_frame_message((const uint8_t *)payload, len, buf, sizeof(buf));
    if (frame_len != (int)(4 + len)) { FAIL("wrong frame length"); return; }

    uint32_t decoded_len = pigeon_read_frame_length(buf);
    if (decoded_len != len) { FAIL("decoded length mismatch"); return; }
    if (memcmp(buf + 4, payload, len) != 0) { FAIL("payload mismatch"); return; }

    PASS();
}

// --- Mock transport for send/recv round-trip tests ---

typedef struct {
    uint8_t stream_buf[PIGEON_MAX_MSG + 64]; // stream data buffer
    size_t  stream_write_pos;                // next byte to write
    size_t  stream_read_pos;                 // next byte to read
    uint8_t dgram_buf[PIGEON_MAX_MSG + 64];  // datagram buffer
    size_t  dgram_len;                       // current datagram size
} mock_transport_state;

static int mock_send_stream(void *ud, const uint8_t *data, size_t len)
{
    mock_transport_state *s = (mock_transport_state *)ud;
    if (s->stream_write_pos + len > sizeof(s->stream_buf)) return -1;
    memcpy(s->stream_buf + s->stream_write_pos, data, len);
    s->stream_write_pos += len;
    return 0;
}

static int mock_recv_stream(void *ud, uint8_t *buf, size_t buf_len, size_t *out_len)
{
    mock_transport_state *s = (mock_transport_state *)ud;
    size_t avail = s->stream_write_pos - s->stream_read_pos;
    size_t to_read = buf_len < avail ? buf_len : avail;
    if (to_read == 0) { *out_len = 0; return -1; } // nothing to read
    memcpy(buf, s->stream_buf + s->stream_read_pos, to_read);
    s->stream_read_pos += to_read;
    *out_len = to_read;
    return 0;
}

static int mock_send_datagram(void *ud, const uint8_t *data, size_t len)
{
    mock_transport_state *s = (mock_transport_state *)ud;
    if (len > sizeof(s->dgram_buf)) return -1;
    memcpy(s->dgram_buf, data, len);
    s->dgram_len = len;
    return 0;
}

static int mock_recv_datagram(void *ud, uint8_t *buf, size_t buf_len, size_t *out_len)
{
    mock_transport_state *s = (mock_transport_state *)ud;
    if (s->dgram_len == 0) { *out_len = 0; return -1; }
    size_t to_read = s->dgram_len < buf_len ? s->dgram_len : buf_len;
    memcpy(buf, s->dgram_buf, to_read);
    *out_len = to_read;
    s->dgram_len = 0;
    return 0;
}

static void init_mock_ctx(pigeon_ctx *ctx, mock_transport_state *state)
{
    memset(state, 0, sizeof(*state));
    pigeon_transport t = {
        .userdata     = state,
        .send_stream  = mock_send_stream,
        .recv_stream  = mock_recv_stream,
        .send_datagram  = mock_send_datagram,
        .recv_datagram  = mock_recv_datagram,
    };
    pigeon_init(ctx, &t);
}

// --- pigeon_send / pigeon_recv round-trip (unencrypted) ---

static void test_send_recv_unencrypted(void)
{
    TEST("pigeon_send/recv round-trip (unencrypted)");
    static pigeon_ctx ctx;
    static mock_transport_state state;
    init_mock_ctx(&ctx, &state);

    const char *msg = "hello pigeon";
    int ret = pigeon_send(&ctx, (const uint8_t *)msg, strlen(msg));
    if (ret != 0) { FAIL("pigeon_send failed"); return; }

    uint8_t out[256];
    int len = pigeon_recv(&ctx, out, sizeof(out));
    if (len < 0) { FAIL("pigeon_recv failed"); return; }
    if ((size_t)len != strlen(msg)) { FAIL("length mismatch"); return; }
    if (memcmp(out, msg, (size_t)len) != 0) { FAIL("data mismatch"); return; }
    // Channel must not be established (no keys set).
    if (ctx.stream_channel.established) { FAIL("channel should not be established"); return; }

    PASS();
}

// --- pigeon_send / pigeon_recv round-trip (encrypted) ---

static void test_send_recv_encrypted(void)
{
    TEST("pigeon_send/recv round-trip (encrypted)");
    static pigeon_ctx sender_ctx, receiver_ctx;
    static mock_transport_state state;
    init_mock_ctx(&sender_ctx, &state);
    // Receiver shares the same mock transport buffer.
    receiver_ctx = sender_ctx;

    // Establish symmetric channels: sender = client, receiver = server.
    uint8_t master[32];
    randombytes_buf(master, 32);
    if (pigeon_channel_init_symmetric(&sender_ctx.stream_channel, master, false) != 0) {
        FAIL("sender channel init failed"); return;
    }
    if (pigeon_channel_init_symmetric(&receiver_ctx.stream_channel, master, true) != 0) {
        FAIL("receiver channel init failed"); return;
    }
    // Both contexts share the same transport state pointer.
    receiver_ctx.transport = sender_ctx.transport;

    if (!sender_ctx.stream_channel.established) { FAIL("sender channel not established"); return; }
    if (!receiver_ctx.stream_channel.established) { FAIL("receiver channel not established"); return; }

    const char *msg = "encrypted pigeon message";
    int ret = pigeon_send(&sender_ctx, (const uint8_t *)msg, strlen(msg));
    if (ret != 0) { FAIL("pigeon_send failed"); return; }

    // Wire: the transport buffer now contains framed ciphertext.
    // Verify it does NOT contain the plaintext verbatim.
    if (memmem(state.stream_buf, state.stream_write_pos, msg, strlen(msg)) != NULL) {
        FAIL("plaintext found in wire buffer (not encrypted)"); return;
    }

    uint8_t out[256];
    int len = pigeon_recv(&receiver_ctx, out, sizeof(out));
    if (len < 0) { FAIL("pigeon_recv decrypt failed"); return; }
    if ((size_t)len != strlen(msg)) { FAIL("decrypted length mismatch"); return; }
    if (memcmp(out, msg, (size_t)len) != 0) { FAIL("decrypted data mismatch"); return; }

    PASS();
}

// --- pigeon_send_datagram / pigeon_recv_datagram round-trip (unencrypted) ---

static void test_send_recv_datagram_unencrypted(void)
{
    TEST("pigeon_send/recv_datagram round-trip (unencrypted)");
    static pigeon_ctx ctx;
    static mock_transport_state state;
    init_mock_ctx(&ctx, &state);

    const char *msg = "datagram hello";
    int ret = pigeon_send_datagram(&ctx, (const uint8_t *)msg, strlen(msg));
    if (ret != 0) { FAIL("pigeon_send_datagram failed"); return; }

    uint8_t out[256];
    int len = pigeon_recv_datagram(&ctx, out, sizeof(out));
    if (len < 0) { FAIL("pigeon_recv_datagram failed"); return; }
    if ((size_t)len != strlen(msg)) { FAIL("length mismatch"); return; }
    if (memcmp(out, msg, (size_t)len) != 0) { FAIL("data mismatch"); return; }

    PASS();
}

// --- pigeon_send_datagram / pigeon_recv_datagram round-trip (encrypted) ---

static void test_send_recv_datagram_encrypted(void)
{
    TEST("pigeon_send/recv_datagram round-trip (encrypted)");
    static pigeon_ctx sender_ctx, receiver_ctx;
    static mock_transport_state state;
    init_mock_ctx(&sender_ctx, &state);
    receiver_ctx = sender_ctx;

    uint8_t master[32];
    randombytes_buf(master, 32);
    if (pigeon_channel_init_symmetric(&sender_ctx.datagram_channel, master, false) != 0) {
        FAIL("sender datagram channel init failed"); return;
    }
    if (pigeon_channel_init_symmetric(&receiver_ctx.datagram_channel, master, true) != 0) {
        FAIL("receiver datagram channel init failed"); return;
    }
    receiver_ctx.transport = sender_ctx.transport;

    const char *msg = "encrypted datagram";
    int ret = pigeon_send_datagram(&sender_ctx, (const uint8_t *)msg, strlen(msg));
    if (ret != 0) { FAIL("pigeon_send_datagram failed"); return; }

    // Wire buffer must not contain plaintext.
    if (memmem(state.dgram_buf, state.dgram_len, msg, strlen(msg)) != NULL) {
        FAIL("plaintext found in datagram wire buffer"); return;
    }

    uint8_t out[256];
    int len = pigeon_recv_datagram(&receiver_ctx, out, sizeof(out));
    if (len < 0) { FAIL("pigeon_recv_datagram decrypt failed"); return; }
    if ((size_t)len != strlen(msg)) { FAIL("decrypted datagram length mismatch"); return; }
    if (memcmp(out, msg, (size_t)len) != 0) { FAIL("decrypted datagram data mismatch"); return; }

    PASS();
}

// --- State machine init ---

static void test_state_machine_init(void)
{
    TEST("pairing machine init (acceptor + initiator)");
    pigeon_acceptor_machine acc;
    pigeon_acceptor_machine_init(&acc);
    if (acc.state != PIGEON_ACCEPTOR_IDLE) { FAIL("wrong acceptor initial state"); return; }

    pigeon_initiator_machine ini;
    pigeon_initiator_machine_init(&ini);
    if (ini.state != PIGEON_INITIATOR_IDLE) { FAIL("wrong initiator initial state"); return; }

    PASS();
}

// --- pigeon_ctx init ---

static void test_ctx_init(void)
{
    TEST("pigeon_ctx init");
    // Use static storage: pigeon_ctx with PIGEON_MAX_MSG=1MiB is too
    // large for the stack. The pairing FSM is no longer pre-allocated
    // inside pigeon_ctx — callers declare their own
    // pigeon_acceptor_machine / pigeon_initiator_machine — so the only
    // post-init invariants to check are the channel state and the
    // zeroed transport pointers.
    static pigeon_ctx ctx;
    pigeon_init(&ctx, NULL);
    if (ctx.stream_channel.send_seq != 0) { FAIL("send_seq not zero"); return; }
    if (ctx.stream_channel.established) { FAIL("stream channel should not be established"); return; }
    if (ctx.datagram_channel.established) { FAIL("datagram channel should not be established"); return; }

    PASS();
}

// --- Datagram mode (gaps allowed, replays rejected) ---

static void test_datagram_mode(void)
{
    TEST("datagram mode (gaps ok, replays rejected)");
    uint8_t key[32];
    randombytes_buf(key, 32);

    pigeon_channel send_ch, recv_ch;
    pigeon_channel_init(&send_ch, key, key, PIGEON_MODE_DATAGRAMS);
    pigeon_channel_init(&recv_ch, key, key, PIGEON_MODE_DATAGRAMS);

    const char *msg = "dgram";
    uint8_t ct0[256], ct1[256], ct2[256], pt[256];

    // Encrypt seq 0, 1, 2.
    int ct0_len = pigeon_channel_encrypt(&send_ch, (const uint8_t *)msg, strlen(msg), ct0, sizeof(ct0));
    int ct1_len = pigeon_channel_encrypt(&send_ch, (const uint8_t *)msg, strlen(msg), ct1, sizeof(ct1));
    int ct2_len = pigeon_channel_encrypt(&send_ch, (const uint8_t *)msg, strlen(msg), ct2, sizeof(ct2));
    if (ct0_len < 0 || ct1_len < 0 || ct2_len < 0) { FAIL("encrypt failed"); return; }

    // Decrypt seq 2 first — gap allowed in datagram mode; recv_seq advances to 3.
    int ret2 = pigeon_channel_decrypt(&recv_ch, ct2, (size_t)ct2_len, pt, sizeof(pt));
    if (ret2 < 0) { FAIL("decrypt seq 2 should succeed"); return; }

    // Decrypt seq 0 — old (recv_seq is now 3), must be rejected.
    int ret0 = pigeon_channel_decrypt(&recv_ch, ct0, (size_t)ct0_len, pt, sizeof(pt));
    if (ret0 >= 0) { FAIL("replay of seq 0 should be rejected"); return; }

    // Decrypt seq 1 — also old, must be rejected.
    int ret1 = pigeon_channel_decrypt(&recv_ch, ct1, (size_t)ct1_len, pt, sizeof(pt));
    if (ret1 >= 0) { FAIL("replay of seq 1 should be rejected"); return; }

    PASS();
}

// --- Sequence counter increments correctly across many messages ---

static void test_multiple_messages(void)
{
    TEST("multiple messages (sequence counter increments)");
    uint8_t key[32];
    randombytes_buf(key, 32);

    pigeon_channel send_ch, recv_ch;
    pigeon_channel_init(&send_ch, key, key, PIGEON_MODE_STRICT);
    pigeon_channel_init(&recv_ch, key, key, PIGEON_MODE_STRICT);

    const char *msg = "ping";
    uint8_t ct[256], pt[256];

    for (int i = 0; i < 10; i++) {
        int ct_len = pigeon_channel_encrypt(&send_ch, (const uint8_t *)msg, strlen(msg), ct, sizeof(ct));
        if (ct_len < 0) { FAIL("encrypt failed"); return; }

        int pt_len = pigeon_channel_decrypt(&recv_ch, ct, (size_t)ct_len, pt, sizeof(pt));
        if (pt_len < 0) { FAIL("decrypt failed"); return; }
        if ((size_t)pt_len != strlen(msg)) { FAIL("wrong plaintext length"); return; }
        if (memcmp(pt, msg, (size_t)pt_len) != 0) { FAIL("plaintext mismatch"); return; }
    }

    PASS();
}

// --- Empty plaintext ---

static void test_empty_plaintext(void)
{
    TEST("empty plaintext (0-byte payload)");
    uint8_t key[32];
    randombytes_buf(key, 32);

    pigeon_channel send_ch, recv_ch;
    pigeon_channel_init(&send_ch, key, key, PIGEON_MODE_STRICT);
    pigeon_channel_init(&recv_ch, key, key, PIGEON_MODE_STRICT);

    uint8_t ct[64], pt[64];
    int ct_len = pigeon_channel_encrypt(&send_ch, (const uint8_t *)"", 0, ct, sizeof(ct));
    // Expected ciphertext: 8 (seq) + 0 (payload) + 16 (tag) = 24 bytes.
    if (ct_len != 24) { FAIL("wrong ciphertext length for empty payload"); return; }

    int pt_len = pigeon_channel_decrypt(&recv_ch, ct, (size_t)ct_len, pt, sizeof(pt));
    if (pt_len < 0) { FAIL("decrypt failed"); return; }
    if (pt_len != 0) { FAIL("expected zero-length plaintext"); return; }

    PASS();
}

// --- Buffer-too-small edge cases ---

static void test_buffer_too_small(void)
{
    TEST("buffer too small edge cases");
    uint8_t key[32];
    randombytes_buf(key, 32);

    pigeon_channel send_ch, recv_ch;
    pigeon_channel_init(&send_ch, key, key, PIGEON_MODE_STRICT);
    pigeon_channel_init(&recv_ch, key, key, PIGEON_MODE_STRICT);

    const char *msg = "hello";
    uint8_t ct[256], pt[256];

    // Encrypt into a buffer that is too small to hold seq + ciphertext + tag.
    // Minimum output for a 5-byte plaintext is 8 + 5 + 16 = 29 bytes.
    int ret = pigeon_channel_encrypt(&send_ch, (const uint8_t *)msg, strlen(msg), ct, 10);
    if (ret >= 0) { FAIL("encrypt with small out_len should return -1"); return; }

    // Produce a valid ciphertext for the decrypt tests.
    int ct_len = pigeon_channel_encrypt(&send_ch, (const uint8_t *)msg, strlen(msg), ct, sizeof(ct));
    if (ct_len < 0) { FAIL("encrypt failed"); return; }

    // Decrypt into a buffer that is too small for the plaintext.
    ret = pigeon_channel_decrypt(&recv_ch, ct, (size_t)ct_len, pt, 0);
    if (ret >= 0) { FAIL("decrypt with small out_len should return -1"); return; }

    // pigeon_frame_message with buf_len < 4 + payload length.
    uint8_t frame_buf[8];
    ret = pigeon_frame_message((const uint8_t *)msg, strlen(msg), frame_buf, 3);
    if (ret >= 0) { FAIL("frame_message with small buf_len should return -1"); return; }

    PASS();
}

// --- State machine transitions for ios actor ---

// Counters incremented by hooked-in pairing-ceremony actions, used to
// verify the spec dispatches the right action on each transition.
static int s_gen_ephemeral_called;
static int s_register_relay_called;
static int s_emit_token_called;
static int s_derive_code_called;
static int s_store_record_called;
static int s_decode_token_called;
static int s_dial_relay_called;

static int act_gen_ephemeral(void *ctx)   { (void)ctx; s_gen_ephemeral_called++;   return 0; }
static int act_register_relay(void *ctx)  { (void)ctx; s_register_relay_called++;  return 0; }
static int act_emit_token(void *ctx)      { (void)ctx; s_emit_token_called++;      return 0; }
static int act_derive_code(void *ctx)     { (void)ctx; s_derive_code_called++;     return 0; }
static int act_store_record(void *ctx)    { (void)ctx; s_store_record_called++;    return 0; }
static int act_decode_token(void *ctx)    { (void)ctx; s_decode_token_called++;    return 0; }
static int act_dial_relay(void *ctx)      { (void)ctx; s_dial_relay_called++;      return 0; }

static void test_state_machine_transitions(void)
{
    TEST("acceptor + initiator full happy paths");

    s_gen_ephemeral_called = 0;
    s_register_relay_called = 0;
    s_emit_token_called = 0;
    s_derive_code_called = 0;
    s_store_record_called = 0;
    s_decode_token_called = 0;
    s_dial_relay_called = 0;

    // ----- acceptor -----
    pigeon_acceptor_machine acc;
    pigeon_acceptor_machine_init(&acc);
    acc.actions[PIGEON_ACTION_GEN_EPHEMERAL]  = act_gen_ephemeral;
    acc.actions[PIGEON_ACTION_REGISTER_RELAY] = act_register_relay;
    acc.actions[PIGEON_ACTION_EMIT_TOKEN]     = act_emit_token;
    acc.actions[PIGEON_ACTION_DERIVE_CODE]    = act_derive_code;
    acc.actions[PIGEON_ACTION_STORE_RECORD]   = act_store_record;

    if (acc.state != PIGEON_ACCEPTOR_IDLE) { FAIL("acceptor: expected IDLE"); return; }
    if (pigeon_acceptor_step(&acc, PIGEON_EVENT_PAIR_BEGIN) != 1) { FAIL("acceptor: step PAIR_BEGIN"); return; }
    if (acc.state != PIGEON_ACCEPTOR_GENERATING_EPHEMERAL) { FAIL("acceptor: expected GENERATING_EPHEMERAL"); return; }
    if (pigeon_acceptor_step(&acc, PIGEON_EVENT_EPHEMERAL_READY) != 1) { FAIL("acceptor: step EPHEMERAL_READY"); return; }
    if (acc.state != PIGEON_ACCEPTOR_REGISTERING_RELAY) { FAIL("acceptor: expected REGISTERING_RELAY"); return; }
    if (pigeon_acceptor_step(&acc, PIGEON_EVENT_RELAY_REGISTERED) != 1) { FAIL("acceptor: step RELAY_REGISTERED"); return; }
    if (acc.state != PIGEON_ACCEPTOR_WAITING_FOR_HELLO) { FAIL("acceptor: expected WAITING_FOR_HELLO"); return; }
    if (pigeon_acceptor_handle_message(&acc, PIGEON_MSG_HELLO) != 1) { FAIL("acceptor: handle HELLO"); return; }
    if (acc.state != PIGEON_ACCEPTOR_DERIVING_CODE) { FAIL("acceptor: expected DERIVING_CODE"); return; }
    if (pigeon_acceptor_step(&acc, PIGEON_EVENT_CODE_READY) != 1) { FAIL("acceptor: step CODE_READY"); return; }
    if (acc.state != PIGEON_ACCEPTOR_AWAITING_USER_CONFIRM) { FAIL("acceptor: expected AWAITING_USER_CONFIRM"); return; }
    if (pigeon_acceptor_step(&acc, PIGEON_EVENT_USER_CONFIRM) != 1) { FAIL("acceptor: step USER_CONFIRM"); return; }
    if (acc.state != PIGEON_ACCEPTOR_AWAITING_PEER_CONFIRM) { FAIL("acceptor: expected AWAITING_PEER_CONFIRM"); return; }
    if (pigeon_acceptor_handle_message(&acc, PIGEON_MSG_CONFIRM_TO_ACCEPTOR) != 1) { FAIL("acceptor: handle CONFIRM_TO_ACCEPTOR"); return; }
    if (acc.state != PIGEON_ACCEPTOR_PAIRED) { FAIL("acceptor: expected PAIRED"); return; }

    if (s_gen_ephemeral_called  != 1) { FAIL("acceptor: gen_ephemeral did not fire"); return; }
    if (s_register_relay_called != 1) { FAIL("acceptor: register_relay did not fire"); return; }
    if (s_emit_token_called     != 1) { FAIL("acceptor: emit_token did not fire"); return; }
    if (s_derive_code_called    != 1) { FAIL("acceptor: derive_code did not fire"); return; }
    if (s_store_record_called   != 1) { FAIL("acceptor: store_record did not fire"); return; }

    // ----- initiator -----
    pigeon_initiator_machine ini;
    pigeon_initiator_machine_init(&ini);
    ini.actions[PIGEON_ACTION_DECODE_TOKEN]  = act_decode_token;
    ini.actions[PIGEON_ACTION_GEN_EPHEMERAL] = act_gen_ephemeral;
    ini.actions[PIGEON_ACTION_DIAL_RELAY]    = act_dial_relay;
    ini.actions[PIGEON_ACTION_DERIVE_CODE]   = act_derive_code;
    ini.actions[PIGEON_ACTION_STORE_RECORD]  = act_store_record;

    if (pigeon_initiator_step(&ini, PIGEON_EVENT_TOKEN_RECEIVED) != 1) { FAIL("initiator: step TOKEN_RECEIVED"); return; }
    if (ini.state != PIGEON_INITIATOR_DECODING_TOKEN) { FAIL("initiator: expected DECODING_TOKEN"); return; }
    if (pigeon_initiator_step(&ini, PIGEON_EVENT_TOKEN_DECODED) != 1) { FAIL("initiator: step TOKEN_DECODED"); return; }
    if (ini.state != PIGEON_INITIATOR_GENERATING_EPHEMERAL) { FAIL("initiator: expected GENERATING_EPHEMERAL"); return; }
    if (pigeon_initiator_step(&ini, PIGEON_EVENT_EPHEMERAL_READY) != 1) { FAIL("initiator: step EPHEMERAL_READY"); return; }
    if (ini.state != PIGEON_INITIATOR_CONNECTING_RELAY) { FAIL("initiator: expected CONNECTING_RELAY"); return; }
    if (pigeon_initiator_step(&ini, PIGEON_EVENT_RELAY_CONNECTED) != 1) { FAIL("initiator: step RELAY_CONNECTED"); return; }
    if (ini.state != PIGEON_INITIATOR_AWAITING_WELCOME) { FAIL("initiator: expected AWAITING_WELCOME"); return; }
    if (pigeon_initiator_handle_message(&ini, PIGEON_MSG_WELCOME) != 1) { FAIL("initiator: handle WELCOME"); return; }
    if (ini.state != PIGEON_INITIATOR_DERIVING_CODE) { FAIL("initiator: expected DERIVING_CODE"); return; }
    if (pigeon_initiator_step(&ini, PIGEON_EVENT_CODE_READY) != 1) { FAIL("initiator: step CODE_READY"); return; }
    if (ini.state != PIGEON_INITIATOR_AWAITING_USER_CONFIRM) { FAIL("initiator: expected AWAITING_USER_CONFIRM"); return; }
    if (pigeon_initiator_step(&ini, PIGEON_EVENT_USER_CONFIRM) != 1) { FAIL("initiator: step USER_CONFIRM"); return; }
    if (ini.state != PIGEON_INITIATOR_AWAITING_PEER_CONFIRM) { FAIL("initiator: expected AWAITING_PEER_CONFIRM"); return; }
    if (pigeon_initiator_handle_message(&ini, PIGEON_MSG_CONFIRM_TO_INITIATOR) != 1) { FAIL("initiator: handle CONFIRM_TO_INITIATOR"); return; }
    if (ini.state != PIGEON_INITIATOR_PAIRED) { FAIL("initiator: expected PAIRED"); return; }

    if (s_decode_token_called   != 1) { FAIL("initiator: decode_token did not fire"); return; }
    if (s_gen_ephemeral_called  != 2) { FAIL("initiator: gen_ephemeral count"); return; }
    if (s_dial_relay_called     != 1) { FAIL("initiator: dial_relay did not fire"); return; }
    if (s_derive_code_called    != 2) { FAIL("initiator: derive_code count"); return; }
    if (s_store_record_called   != 2) { FAIL("initiator: store_record count"); return; }

    PASS();
}

// --- Cross-language vector validation ---

// Decode a hex string into buf. Returns the number of bytes written on
// success, or -1 if the hex string is malformed or the buffer is too small.
static int hex_decode(const char *hex, uint8_t *buf, size_t buf_len)
{
    size_t hex_len = strlen(hex);
    if (hex_len % 2 != 0 || hex_len / 2 > buf_len) return -1;
    for (size_t i = 0; i < hex_len / 2; i++) {
        unsigned int byte;
        if (sscanf(hex + 2 * i, "%02x", &byte) != 1) return -1;
        buf[i] = (uint8_t)byte;
    }
    return (int)(hex_len / 2);
}

// Extract the value of a JSON string field from a flat JSON object.
// Writes a null-terminated copy into out (up to out_len - 1 chars).
// Returns 0 on success, -1 if not found or buffer too small.
static int json_extract_string(const char *json, const char *key,
                                char *out, size_t out_len)
{
    char pattern[128];
    snprintf(pattern, sizeof(pattern), "\"%s\": \"", key);
    const char *p = strstr(json, pattern);
    if (!p) return -1;
    p += strlen(pattern);
    const char *end = strchr(p, '"');
    if (!end) return -1;
    size_t len = (size_t)(end - p);
    if (len >= out_len) return -1;
    memcpy(out, p, len);
    out[len] = '\0';
    return 0;
}

static void test_cross_language_vectors(void)
{
    TEST("cross-language vector validation (Go->C)");

    // Load vectors.json — expected to be run from the repo root.
    const char *path = "c/test/vectors.json";
    FILE *f = fopen(path, "r");
    if (!f) {
        FAIL("could not open c/test/vectors.json (run from repo root)");
        return;
    }
    fseek(f, 0, SEEK_END);
    long file_size = ftell(f);
    fseek(f, 0, SEEK_SET);
    char *json = malloc((size_t)file_size + 1);
    if (!json) { fclose(f); FAIL("malloc failed"); return; }
    fread(json, 1, (size_t)file_size, f);
    fclose(f);
    json[file_size] = '\0';

    // Extract all fields.
    char alice_private_hex[65], alice_public_hex[65];
    char bob_private_hex[65], bob_public_hex[65];
    char session_key_hex[65];
    char confirmation_code[8];
    char ciphertext_hex[512];
    char symmetric_master_hex[65];
    char symmetric_ciphertext_c2s_hex[512];
    char symmetric_ciphertext_s2c_hex[512];

    if (json_extract_string(json, "alice_private", alice_private_hex, sizeof(alice_private_hex)) != 0 ||
        json_extract_string(json, "alice_public",  alice_public_hex,  sizeof(alice_public_hex))  != 0 ||
        json_extract_string(json, "bob_private",   bob_private_hex,   sizeof(bob_private_hex))   != 0 ||
        json_extract_string(json, "bob_public",    bob_public_hex,    sizeof(bob_public_hex))    != 0 ||
        json_extract_string(json, "session_key",   session_key_hex,   sizeof(session_key_hex))   != 0 ||
        json_extract_string(json, "confirmation_code", confirmation_code, sizeof(confirmation_code)) != 0 ||
        json_extract_string(json, "ciphertext",    ciphertext_hex,    sizeof(ciphertext_hex))    != 0 ||
        json_extract_string(json, "symmetric_master", symmetric_master_hex, sizeof(symmetric_master_hex)) != 0 ||
        json_extract_string(json, "symmetric_ciphertext_c2s", symmetric_ciphertext_c2s_hex,
                            sizeof(symmetric_ciphertext_c2s_hex)) != 0 ||
        json_extract_string(json, "symmetric_ciphertext_s2c", symmetric_ciphertext_s2c_hex,
                            sizeof(symmetric_ciphertext_s2c_hex)) != 0) {
        free(json);
        FAIL("failed to extract one or more fields from vectors.json");
        return;
    }
    free(json);

    // Decode raw bytes.
    uint8_t alice_private[32], alice_public[32];
    uint8_t bob_private[32], bob_public[32];
    uint8_t expected_session_key[32];
    uint8_t ciphertext[256], symmetric_master[32];
    uint8_t sym_ct_c2s[256], sym_ct_s2c[256];

    if (hex_decode(alice_private_hex, alice_private, 32) < 0 ||
        hex_decode(alice_public_hex,  alice_public,  32) < 0 ||
        hex_decode(bob_private_hex,   bob_private,   32) < 0 ||
        hex_decode(bob_public_hex,    bob_public,    32) < 0 ||
        hex_decode(session_key_hex,   expected_session_key, 32) < 0) {
        FAIL("hex_decode failed for key material");
        return;
    }
    int ct_len = hex_decode(ciphertext_hex, ciphertext, sizeof(ciphertext));
    if (ct_len < 0) { FAIL("hex_decode failed for ciphertext"); return; }
    if (hex_decode(symmetric_master_hex, symmetric_master, 32) < 0) {
        FAIL("hex_decode failed for symmetric_master");
        return;
    }
    int sym_ct_c2s_len = hex_decode(symmetric_ciphertext_c2s_hex, sym_ct_c2s, sizeof(sym_ct_c2s));
    int sym_ct_s2c_len = hex_decode(symmetric_ciphertext_s2c_hex, sym_ct_s2c, sizeof(sym_ct_s2c));
    if (sym_ct_c2s_len < 0 || sym_ct_s2c_len < 0) {
        FAIL("hex_decode failed for symmetric ciphertexts");
        return;
    }

    // Verify public keys: X25519(private, base_point) must equal the vector value.
    uint8_t derived_alice_pub[32], derived_bob_pub[32];
    if (crypto_scalarmult_base(derived_alice_pub, alice_private) != 0 ||
        crypto_scalarmult_base(derived_bob_pub, bob_private) != 0) {
        FAIL("crypto_scalarmult_base failed");
        return;
    }
    if (memcmp(derived_alice_pub, alice_public, 32) != 0) {
        FAIL("alice derived public key does not match vector");
        return;
    }
    if (memcmp(derived_bob_pub, bob_public, 32) != 0) {
        FAIL("bob derived public key does not match vector");
        return;
    }

    // Session key: alice's private + bob's public, info="test-session".
    const uint8_t info_session[] = "test-session";
    uint8_t derived_session_key[32];
    if (pigeon_derive_session_key(alice_private, bob_public,
                                   info_session, sizeof(info_session) - 1,
                                   derived_session_key) != 0) {
        FAIL("pigeon_derive_session_key failed");
        return;
    }
    if (memcmp(derived_session_key, expected_session_key, 32) != 0) {
        FAIL("derived session key does not match Go vector");
        return;
    }
    // Bob's side must derive the same key.
    uint8_t bob_session_key[32];
    if (pigeon_derive_session_key(bob_private, alice_public,
                                   info_session, sizeof(info_session) - 1,
                                   bob_session_key) != 0) {
        FAIL("pigeon_derive_session_key (bob) failed");
        return;
    }
    if (memcmp(bob_session_key, expected_session_key, 32) != 0) {
        FAIL("bob session key does not match Go vector");
        return;
    }

    // Confirmation code.
    char derived_code[7];
    if (pigeon_derive_confirmation_code(alice_public, bob_public, derived_code) != 0) {
        FAIL("pigeon_derive_confirmation_code failed");
        return;
    }
    if (strcmp(derived_code, confirmation_code) != 0) {
        FAIL("confirmation code does not match Go vector");
        return;
    }

    // Decrypt the Go-generated ciphertext (alice->bob direction).
    // Go encrypted with alice's channel: sendKey = DeriveSessionKey(alice, bob, "alice-to-bob")
    // Bob's channel to decrypt: recvKey = DeriveSessionKey(bob, alice, "alice-to-bob")
    const uint8_t info_a2b[] = "alice-to-bob";
    const uint8_t info_b2a[] = "bob-to-alice";
    uint8_t bob_send_key[32], bob_recv_key[32];
    if (pigeon_derive_session_key(bob_private, alice_public,
                                   info_b2a, sizeof(info_b2a) - 1, bob_send_key) != 0 ||
        pigeon_derive_session_key(bob_private, alice_public,
                                   info_a2b, sizeof(info_a2b) - 1, bob_recv_key) != 0) {
        FAIL("pigeon_derive_session_key for alice-to-bob channel failed");
        return;
    }

    pigeon_channel bob_ch;
    pigeon_channel_init(&bob_ch, bob_send_key, bob_recv_key, PIGEON_MODE_STRICT);

    uint8_t plaintext[256];
    int pt_len = pigeon_channel_decrypt(&bob_ch, ciphertext, (size_t)ct_len,
                                         plaintext, sizeof(plaintext));
    if (pt_len < 0) {
        FAIL("channel_decrypt failed for Go-generated ciphertext");
        return;
    }
    const char *expected_plaintext = "hello from pigeon";
    if ((size_t)pt_len != strlen(expected_plaintext) ||
        memcmp(plaintext, expected_plaintext, (size_t)pt_len) != 0) {
        FAIL("decrypted plaintext does not match Go vector");
        return;
    }

    // Symmetric channel: server decrypts c2s, client decrypts s2c.
    pigeon_channel server_sym_ch, client_sym_ch;
    if (pigeon_channel_init_symmetric(&server_sym_ch, symmetric_master, true)  != 0 ||
        pigeon_channel_init_symmetric(&client_sym_ch, symmetric_master, false) != 0) {
        FAIL("pigeon_channel_init_symmetric failed");
        return;
    }

    uint8_t sym_pt[256];
    const char *expected_sym_plaintext = "symmetric test";

    int sym_pt_len = pigeon_channel_decrypt(&server_sym_ch, sym_ct_c2s, (size_t)sym_ct_c2s_len,
                                             sym_pt, sizeof(sym_pt));
    if (sym_pt_len < 0) {
        FAIL("symmetric c2s decrypt failed");
        return;
    }
    if ((size_t)sym_pt_len != strlen(expected_sym_plaintext) ||
        memcmp(sym_pt, expected_sym_plaintext, (size_t)sym_pt_len) != 0) {
        FAIL("symmetric c2s plaintext does not match Go vector");
        return;
    }

    sym_pt_len = pigeon_channel_decrypt(&client_sym_ch, sym_ct_s2c, (size_t)sym_ct_s2c_len,
                                         sym_pt, sizeof(sym_pt));
    if (sym_pt_len < 0) {
        FAIL("symmetric s2c decrypt failed");
        return;
    }
    if ((size_t)sym_pt_len != strlen(expected_sym_plaintext) ||
        memcmp(sym_pt, expected_sym_plaintext, (size_t)sym_pt_len) != 0) {
        FAIL("symmetric s2c plaintext does not match Go vector");
        return;
    }

    PASS();
}

// --- PairingRecord serialisation round-trip ---

static void test_pairing_record_roundtrip(void)
{
    TEST("pairing record serialize/deserialize round-trip");

    pigeon_pairing_record orig;
    memset(&orig, 0, sizeof(orig));

    // Fill with recognisable data.
    strncpy(orig.peer_instance_id, "instance-abc-123", sizeof(orig.peer_instance_id) - 1);
    strncpy(orig.relay_url, "https://relay.example.com:443", sizeof(orig.relay_url) - 1);
    for (int i = 0; i < 32; i++) {
        orig.local_private_key[i] = (uint8_t)(0x10 + i);
        orig.local_public_key[i]  = (uint8_t)(0x20 + i);
        orig.peer_public_key[i]   = (uint8_t)(0x30 + i);
    }

    uint8_t buf[PIGEON_PAIRING_RECORD_SIZE];
    int written = pigeon_pairing_record_serialize(&orig, buf, sizeof(buf));
    if (written != PIGEON_PAIRING_RECORD_SIZE) { FAIL("serialize returned wrong size"); return; }

    pigeon_pairing_record decoded;
    memset(&decoded, 0xff, sizeof(decoded)); // Fill with garbage before decode.
    int consumed = pigeon_pairing_record_deserialize(&decoded, buf, sizeof(buf));
    if (consumed != PIGEON_PAIRING_RECORD_SIZE) { FAIL("deserialize returned wrong size"); return; }

    if (memcmp(&orig, &decoded, sizeof(orig)) != 0) { FAIL("round-trip mismatch"); return; }

    // Buffer-too-small rejection.
    if (pigeon_pairing_record_serialize(&orig, buf, PIGEON_PAIRING_RECORD_SIZE - 1) >= 0) {
        FAIL("serialize with small buf should return -1"); return;
    }
    if (pigeon_pairing_record_deserialize(&decoded, buf, PIGEON_PAIRING_RECORD_SIZE - 1) >= 0) {
        FAIL("deserialize with small buf should return -1"); return;
    }

    // Bad magic rejection.
    buf[0] = 0x00;
    if (pigeon_pairing_record_serialize(&orig, buf, sizeof(buf)) != PIGEON_PAIRING_RECORD_SIZE) {
        FAIL("re-serialize failed"); return;
    }
    buf[0] = 0xFF; // Corrupt magic.
    if (pigeon_pairing_record_deserialize(&decoded, buf, sizeof(buf)) >= 0) {
        FAIL("deserialize with bad magic should return -1"); return;
    }

    // Bad version rejection.
    if (pigeon_pairing_record_serialize(&orig, buf, sizeof(buf)) != PIGEON_PAIRING_RECORD_SIZE) {
        FAIL("re-serialize failed"); return;
    }
    buf[3] = 99; // Corrupt version.
    if (pigeon_pairing_record_deserialize(&decoded, buf, sizeof(buf)) >= 0) {
        FAIL("deserialize with bad version should return -1"); return;
    }

    PASS();
}

int main(void)
{
    if (sodium_init() < 0) {
        fprintf(stderr, "sodium_init failed\n");
        return 1;
    }

    printf("pigeon C library tests\n\n");

    test_keypair();
    test_session_key_derivation();
    test_confirmation_code();
    test_channel_roundtrip();
    test_symmetric_channel();
    test_sequence_strict();
    test_framing();
    test_state_machine_init();
    test_ctx_init();
    test_datagram_mode();
    test_multiple_messages();
    test_empty_plaintext();
    test_buffer_too_small();
    test_state_machine_transitions();
    test_cross_language_vectors();
    test_pairing_record_roundtrip();
    test_send_recv_unencrypted();
    test_send_recv_encrypted();
    test_send_recv_datagram_unencrypted();
    test_send_recv_datagram_encrypted();

    printf("\n%d/%d tests passed\n", tests_passed, tests_run);
    return tests_passed == tests_run ? 0 : 1;
}

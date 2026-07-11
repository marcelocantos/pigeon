// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

#include "pigeon.h"
#include <assert.h>
#include <pthread.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
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
    acc.actions[PIGEON_PAIRINGCEREMONY_ACTION_GEN_EPHEMERAL]  = act_gen_ephemeral;
    acc.actions[PIGEON_PAIRINGCEREMONY_ACTION_REGISTER_RELAY] = act_register_relay;
    acc.actions[PIGEON_PAIRINGCEREMONY_ACTION_EMIT_TOKEN]     = act_emit_token;
    acc.actions[PIGEON_PAIRINGCEREMONY_ACTION_DERIVE_CODE]    = act_derive_code;
    acc.actions[PIGEON_PAIRINGCEREMONY_ACTION_STORE_RECORD]   = act_store_record;

    if (acc.state != PIGEON_ACCEPTOR_IDLE) { FAIL("acceptor: expected IDLE"); return; }
    if (pigeon_acceptor_step(&acc, PIGEON_PAIRINGCEREMONY_EVENT_PAIR_BEGIN) != 1) { FAIL("acceptor: step PAIR_BEGIN"); return; }
    if (acc.state != PIGEON_ACCEPTOR_GENERATING_EPHEMERAL) { FAIL("acceptor: expected GENERATING_EPHEMERAL"); return; }
    if (pigeon_acceptor_step(&acc, PIGEON_PAIRINGCEREMONY_EVENT_EPHEMERAL_READY) != 1) { FAIL("acceptor: step EPHEMERAL_READY"); return; }
    if (acc.state != PIGEON_ACCEPTOR_REGISTERING_RELAY) { FAIL("acceptor: expected REGISTERING_RELAY"); return; }
    if (pigeon_acceptor_step(&acc, PIGEON_PAIRINGCEREMONY_EVENT_RELAY_REGISTERED) != 1) { FAIL("acceptor: step RELAY_REGISTERED"); return; }
    if (acc.state != PIGEON_ACCEPTOR_WAITING_FOR_HELLO) { FAIL("acceptor: expected WAITING_FOR_HELLO"); return; }
    if (pigeon_acceptor_handle_message(&acc, PIGEON_PAIRINGCEREMONY_MSG_HELLO) != 1) { FAIL("acceptor: handle HELLO"); return; }
    if (acc.state != PIGEON_ACCEPTOR_DERIVING_CODE) { FAIL("acceptor: expected DERIVING_CODE"); return; }
    if (pigeon_acceptor_step(&acc, PIGEON_PAIRINGCEREMONY_EVENT_CODE_READY) != 1) { FAIL("acceptor: step CODE_READY"); return; }
    if (acc.state != PIGEON_ACCEPTOR_AWAITING_USER_CONFIRM) { FAIL("acceptor: expected AWAITING_USER_CONFIRM"); return; }
    if (pigeon_acceptor_step(&acc, PIGEON_PAIRINGCEREMONY_EVENT_USER_CONFIRM) != 1) { FAIL("acceptor: step USER_CONFIRM"); return; }
    if (acc.state != PIGEON_ACCEPTOR_AWAITING_PEER_CONFIRM) { FAIL("acceptor: expected AWAITING_PEER_CONFIRM"); return; }
    if (pigeon_acceptor_handle_message(&acc, PIGEON_PAIRINGCEREMONY_MSG_CONFIRM_TO_ACCEPTOR) != 1) { FAIL("acceptor: handle CONFIRM_TO_ACCEPTOR"); return; }
    if (acc.state != PIGEON_ACCEPTOR_PAIRED) { FAIL("acceptor: expected PAIRED"); return; }

    if (s_gen_ephemeral_called  != 1) { FAIL("acceptor: gen_ephemeral did not fire"); return; }
    if (s_register_relay_called != 1) { FAIL("acceptor: register_relay did not fire"); return; }
    if (s_emit_token_called     != 1) { FAIL("acceptor: emit_token did not fire"); return; }
    if (s_derive_code_called    != 1) { FAIL("acceptor: derive_code did not fire"); return; }
    if (s_store_record_called   != 1) { FAIL("acceptor: store_record did not fire"); return; }

    // ----- initiator -----
    pigeon_initiator_machine ini;
    pigeon_initiator_machine_init(&ini);
    ini.actions[PIGEON_PAIRINGCEREMONY_ACTION_DECODE_TOKEN]  = act_decode_token;
    ini.actions[PIGEON_PAIRINGCEREMONY_ACTION_GEN_EPHEMERAL] = act_gen_ephemeral;
    ini.actions[PIGEON_PAIRINGCEREMONY_ACTION_DIAL_RELAY]    = act_dial_relay;
    ini.actions[PIGEON_PAIRINGCEREMONY_ACTION_DERIVE_CODE]   = act_derive_code;
    ini.actions[PIGEON_PAIRINGCEREMONY_ACTION_STORE_RECORD]  = act_store_record;

    if (pigeon_initiator_step(&ini, PIGEON_PAIRINGCEREMONY_EVENT_TOKEN_RECEIVED) != 1) { FAIL("initiator: step TOKEN_RECEIVED"); return; }
    if (ini.state != PIGEON_INITIATOR_DECODING_TOKEN) { FAIL("initiator: expected DECODING_TOKEN"); return; }
    if (pigeon_initiator_step(&ini, PIGEON_PAIRINGCEREMONY_EVENT_TOKEN_DECODED) != 1) { FAIL("initiator: step TOKEN_DECODED"); return; }
    if (ini.state != PIGEON_INITIATOR_GENERATING_EPHEMERAL) { FAIL("initiator: expected GENERATING_EPHEMERAL"); return; }
    if (pigeon_initiator_step(&ini, PIGEON_PAIRINGCEREMONY_EVENT_EPHEMERAL_READY) != 1) { FAIL("initiator: step EPHEMERAL_READY"); return; }
    if (ini.state != PIGEON_INITIATOR_CONNECTING_RELAY) { FAIL("initiator: expected CONNECTING_RELAY"); return; }
    if (pigeon_initiator_step(&ini, PIGEON_PAIRINGCEREMONY_EVENT_RELAY_CONNECTED) != 1) { FAIL("initiator: step RELAY_CONNECTED"); return; }
    if (ini.state != PIGEON_INITIATOR_AWAITING_WELCOME) { FAIL("initiator: expected AWAITING_WELCOME"); return; }
    if (pigeon_initiator_handle_message(&ini, PIGEON_PAIRINGCEREMONY_MSG_WELCOME) != 1) { FAIL("initiator: handle WELCOME"); return; }
    if (ini.state != PIGEON_INITIATOR_DERIVING_CODE) { FAIL("initiator: expected DERIVING_CODE"); return; }
    if (pigeon_initiator_step(&ini, PIGEON_PAIRINGCEREMONY_EVENT_CODE_READY) != 1) { FAIL("initiator: step CODE_READY"); return; }
    if (ini.state != PIGEON_INITIATOR_AWAITING_USER_CONFIRM) { FAIL("initiator: expected AWAITING_USER_CONFIRM"); return; }
    if (pigeon_initiator_step(&ini, PIGEON_PAIRINGCEREMONY_EVENT_USER_CONFIRM) != 1) { FAIL("initiator: step USER_CONFIRM"); return; }
    if (ini.state != PIGEON_INITIATOR_AWAITING_PEER_CONFIRM) { FAIL("initiator: expected AWAITING_PEER_CONFIRM"); return; }
    if (pigeon_initiator_handle_message(&ini, PIGEON_PAIRINGCEREMONY_MSG_CONFIRM_TO_INITIATOR) != 1) { FAIL("initiator: handle CONFIRM_TO_INITIATOR"); return; }
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

// --- Multi-channel wire helpers (post-T22) ---

static void test_uvarint(void)
{
    TEST("Go-style uvarint encode/decode + reference vectors");

    // Reference vectors verified against Go's encoding/binary.PutUvarint:
    //   0       -> [0x00]
    //   1       -> [0x01]
    //   127     -> [0x7f]
    //   128     -> [0x80, 0x01]
    //   300     -> [0xac, 0x02]
    //   16384   -> [0x80, 0x80, 0x01]
    //   2^63    -> [0x80,0x80,0x80,0x80,0x80,0x80,0x80,0x80,0x80,0x01]
    struct { uint64_t v; uint8_t bytes[10]; size_t n; } cases[] = {
        {0,        {0x00}, 1},
        {1,        {0x01}, 1},
        {127,      {0x7f}, 1},
        {128,      {0x80, 0x01}, 2},
        {300,      {0xac, 0x02}, 2},
        {16384,    {0x80, 0x80, 0x01}, 3},
        {(uint64_t)1 << 63,
                   {0x80,0x80,0x80,0x80,0x80,0x80,0x80,0x80,0x80,0x01}, 10},
    };

    for (size_t i = 0; i < sizeof(cases)/sizeof(cases[0]); i++) {
        uint8_t buf[16];
        int n = pigeon_uvarint_encode(cases[i].v, buf, sizeof(buf));
        if (n != (int)cases[i].n) { FAIL("encode wrong length"); return; }
        if (memcmp(buf, cases[i].bytes, cases[i].n) != 0) { FAIL("encode bytes mismatch"); return; }

        uint64_t out = 0;
        int consumed = pigeon_uvarint_decode(cases[i].bytes, cases[i].n, &out);
        if (consumed != (int)cases[i].n) { FAIL("decode consumed wrong"); return; }
        if (out != cases[i].v) { FAIL("decode value mismatch"); return; }
    }

    // Truncated input -> 0 (need more bytes).
    uint8_t trunc[] = {0x80};
    uint64_t v;
    if (pigeon_uvarint_decode(trunc, 1, &v) != 0) { FAIL("truncated should return 0"); return; }

    // Buffer-too-small on encode.
    uint8_t small[1];
    if (pigeon_uvarint_encode(128, small, 1) >= 0) { FAIL("encode should fail on small buf"); return; }

    PASS();
}

static void test_stream_header(void)
{
    TEST("stream-header encode/decode (T45: [varint name-len][name])");

    // "chat" (length 4 -> varint 0x04). Expected wire: 04 'c' 'h' 'a' 't'.
    {
        uint8_t out[32];
        int n = pigeon_wire_stream_header_encode("chat", 4, out, sizeof(out));
        if (n != 5) { FAIL("chat encode wrong length"); return; }
        const uint8_t want[] = {0x04, 'c','h','a','t'};
        if (memcmp(out, want, 5) != 0) { FAIL("chat encode bytes mismatch"); return; }

        char name[32]; size_t name_len = 0;
        int consumed = pigeon_wire_stream_header_decode(out, (size_t)n,
                                                        name, sizeof(name),
                                                        &name_len);
        if (consumed != 5) { FAIL("chat decode consumed wrong"); return; }
        if (name_len != 4 || strcmp(name, "chat") != 0) { FAIL("chat decode name mismatch"); return; }
    }

    // Primary: empty name -> [0x00] (just the varint 0).
    {
        uint8_t out[32];
        int n = pigeon_wire_stream_header_encode(NULL, 0, out, sizeof(out));
        if (n != 1) { FAIL("empty encode wrong length"); return; }
        if (out[0] != 0x00) { FAIL("empty encode byte"); return; }

        char name[8]; size_t name_len = 1;
        int consumed = pigeon_wire_stream_header_decode(out, (size_t)n,
                                                        name, sizeof(name),
                                                        &name_len);
        if (consumed != 1) { FAIL("empty decode consumed"); return; }
        if (name_len != 0 || name[0] != '\0') { FAIL("empty decode name"); return; }
    }

    // Named: "control" (length 7 -> varint 0x07).
    {
        uint8_t out[32];
        int n = pigeon_wire_stream_header_encode("control", 7, out, sizeof(out));
        if (n != 8) { FAIL("control encode wrong length"); return; }
        const uint8_t want[] = {0x07, 'c','o','n','t','r','o','l'};
        if (memcmp(out, want, 8) != 0) { FAIL("control encode bytes"); return; }

        char name[16]; size_t name_len = 0;
        int consumed = pigeon_wire_stream_header_decode(out, (size_t)n,
                                                        name, sizeof(name),
                                                        &name_len);
        if (consumed != 8) { FAIL("control decode consumed"); return; }
        if (name_len != 7 || strcmp(name, "control") != 0) { FAIL("control decode name"); return; }
    }

    // Decode rejects a name longer than name_buf (must leave room for NUL).
    {
        uint8_t out[32];
        int n = pigeon_wire_stream_header_encode("abcdef", 6, out, sizeof(out));
        if (n != 7) { FAIL("setup"); return; }
        char small[6]; size_t nl = 0;
        if (pigeon_wire_stream_header_decode(out, (size_t)n, small, sizeof(small), &nl) >= 0) {
            FAIL("decode should fail when name_buf too small for NUL"); return;
        }
    }

    PASS();
}

static void test_datagram_framing(void)
{
    TEST("datagram framing: AEAD([varint id][payload]) (T45: no tag prefix)");

    uint8_t key[32];
    randombytes_buf(key, 32);

    // Two ends sharing the same symmetric key: one sends, the other receives.
    pigeon_channel send_ch, recv_ch;
    pigeon_channel_init(&send_ch, key, key, PIGEON_MODE_DATAGRAMS);
    pigeon_channel_init(&recv_ch, key, key, PIGEON_MODE_DATAGRAMS);

    {
        const uint8_t payload[] = "hello";
        uint8_t wire[256];
        int wn = pigeon_encode_datagram(&send_ch, /*channel_id=*/1,
                                        payload, sizeof(payload) - 1,
                                        wire, sizeof(wire));
        if (wn < 0) { FAIL("encode datagram"); return; }

        uint64_t cid = 0; uint8_t out[256];
        int pn = pigeon_decode_datagram(&recv_ch, wire, (size_t)wn,
                                        &cid, out, sizeof(out));
        if (pn != (int)(sizeof(payload) - 1)) { FAIL("decode wrong len"); return; }
        if (cid != 1) { FAIL("decode channel id"); return; }
        if (memcmp(out, payload, sizeof(payload) - 1) != 0) { FAIL("decode payload"); return; }
    }

    // Multi-byte channel id (300 -> varint ac 02).
    randombytes_buf(key, 32);
    pigeon_channel_init(&send_ch, key, key, PIGEON_MODE_DATAGRAMS);
    pigeon_channel_init(&recv_ch, key, key, PIGEON_MODE_DATAGRAMS);
    {
        const uint8_t payload[] = "ping";
        uint8_t wire[256];
        int wn = pigeon_encode_datagram(&send_ch, /*channel_id=*/300,
                                        payload, sizeof(payload) - 1,
                                        wire, sizeof(wire));
        if (wn < 0) { FAIL("encode datagram (multi-byte cid)"); return; }

        uint64_t cid = 0; uint8_t out[256];
        int pn = pigeon_decode_datagram(&recv_ch, wire, (size_t)wn,
                                        &cid, out, sizeof(out));
        if (pn != (int)(sizeof(payload) - 1)) { FAIL("decode wrong len"); return; }
        if (cid != 300) { FAIL("decode channel id"); return; }
        if (memcmp(out, payload, sizeof(payload) - 1) != 0) { FAIL("decode payload"); return; }
    }

    PASS();
}

// --- In-process loopback transport ---
//
// A minimal pigeon_transport implementation that wires two sessions
// together inside the same process, with no network. Each side has
// its own stream queues and datagram queue. Tests use this to exercise
// the pigeon_session / pigeon_stream / pigeon_datagram API without
// needing ngtcp2 or a live relay.

#define LOOP_MAX_STREAMS 8

// Single in-flight message on a per-stream / per-datagram queue.
// Allocated on send_*, freed on recv_*; payload sized to the actual
// message length so the test transport's memory tracks real traffic
// rather than preallocating the SDK's PIGEON_MAX_MSG ceiling.
typedef struct loopback_msg {
    struct loopback_msg *next;
    size_t  len;
    uint8_t bytes[];
} loopback_msg;

typedef struct loopback_stream {
    int  id;                          // stream identifier (slot index)
    bool in_use;
    bool accepted;                    // matched by accept_stream on the peer
    loopback_msg *msg_head, *msg_tail;
} loopback_stream;

typedef struct loopback_endpoint {
    loopback_stream  streams[LOOP_MAX_STREAMS];
    int              next_stream_id;

    // Datagrams incoming to this endpoint.
    loopback_msg *dgram_head, *dgram_tail;

    // Stream IDs awaiting accept_stream by this endpoint.
    int accept_queue[LOOP_MAX_STREAMS];
    int accept_head, accept_tail, accept_count;

    struct loopback_endpoint *peer;
} loopback_endpoint;

static loopback_msg *loopback_msg_new(const uint8_t *data, size_t len)
{
    loopback_msg *m = (loopback_msg *)malloc(sizeof(*m) + len);
    if (!m) return NULL;
    m->next = NULL;
    m->len  = len;
    memcpy(m->bytes, data, len);
    return m;
}

static void loopback_queue_push(loopback_msg **head, loopback_msg **tail,
                                loopback_msg *m)
{
    if (*tail) (*tail)->next = m;
    else       *head = m;
    *tail = m;
}

static loopback_msg *loopback_queue_pop(loopback_msg **head, loopback_msg **tail)
{
    loopback_msg *m = *head;
    if (!m) return NULL;
    *head = m->next;
    if (!*head) *tail = NULL;
    return m;
}

static void loopback_queue_drain(loopback_msg **head, loopback_msg **tail)
{
    loopback_msg *m;
    while ((m = loopback_queue_pop(head, tail)) != NULL) free(m);
}

static void loopback_endpoint_init(loopback_endpoint *e, loopback_endpoint *peer)
{
    memset(e, 0, sizeof(*e));
    e->peer = peer;
}

static void loopback_endpoint_drain(loopback_endpoint *e)
{
    for (int i = 0; i < LOOP_MAX_STREAMS; i++) {
        loopback_queue_drain(&e->streams[i].msg_head, &e->streams[i].msg_tail);
    }
    loopback_queue_drain(&e->dgram_head, &e->dgram_tail);
}

static loopback_stream *loopback_alloc_stream(loopback_endpoint *e)
{
    for (int i = 0; i < LOOP_MAX_STREAMS; i++) {
        if (!e->streams[i].in_use) {
            loopback_queue_drain(&e->streams[i].msg_head, &e->streams[i].msg_tail);
            memset(&e->streams[i], 0, sizeof(e->streams[i]));
            e->streams[i].in_use = true;
            e->streams[i].id = i;
            return &e->streams[i];
        }
    }
    return NULL;
}

static int loop_open_stream(void *ud, pigeon_stream_handle **out)
{
    loopback_endpoint *e = (loopback_endpoint *)ud;
    loopback_stream *me = loopback_alloc_stream(e);
    if (!me) return -1;
    // Allocate the matching slot on the peer with the same id (paired).
    loopback_endpoint *p = e->peer;
    if (p->streams[me->id].in_use) return -1;
    loopback_queue_drain(&p->streams[me->id].msg_head, &p->streams[me->id].msg_tail);
    memset(&p->streams[me->id], 0, sizeof(p->streams[me->id]));
    p->streams[me->id].in_use = true;
    p->streams[me->id].id = me->id;
    // Queue the new stream for the peer to accept.
    p->accept_queue[p->accept_tail] = me->id;
    p->accept_tail = (p->accept_tail + 1) % LOOP_MAX_STREAMS;
    p->accept_count++;
    *out = (pigeon_stream_handle *)me;
    return 0;
}

static int loop_accept_stream(void *ud, pigeon_stream_handle **out)
{
    loopback_endpoint *e = (loopback_endpoint *)ud;
    // Blocking spin-wait so multi-threaded listener tests (T32.2) can
    // park on accept while client threads open streams. Caps at ~5s.
    // Same shape and rationale as loop_recv_on_stream above; single-
    // threaded tests are unaffected — they always open before accept.
    for (int waited_us = 0; e->accept_count == 0 && waited_us < 5000000; waited_us += 1000) {
        usleep(1000);
    }
    if (e->accept_count == 0) return -1;
    int id = e->accept_queue[e->accept_head];
    e->accept_head = (e->accept_head + 1) % LOOP_MAX_STREAMS;
    e->accept_count--;
    if (id < 0 || id >= LOOP_MAX_STREAMS) return -1;
    if (!e->streams[id].in_use) return -1;
    e->streams[id].accepted = true;
    *out = (pigeon_stream_handle *)&e->streams[id];
    return 0;
}

static int loop_send_on_stream(void *ud, pigeon_stream_handle *h,
                               const uint8_t *data, size_t len)
{
    loopback_endpoint *e = (loopback_endpoint *)ud;
    loopback_stream *me = (loopback_stream *)h;
    if (!me->in_use) return -1;
    // Route to the peer's mirror slot.
    loopback_stream *peer = &e->peer->streams[me->id];
    if (!peer->in_use) return -1;
    loopback_msg *m = loopback_msg_new(data, len);
    if (!m) return -1;
    loopback_queue_push(&peer->msg_head, &peer->msg_tail, m);
    return 0;
}

static int loop_recv_on_stream(void *ud, pigeon_stream_handle *h,
                               uint8_t *buf, size_t buf_len, size_t *out_len)
{
    (void)ud;
    loopback_stream *me = (loopback_stream *)h;
    if (!me->in_use) return -1;
    // Blocking spin-wait so multi-threaded tests (activation, etc.)
    // can drive backend and client concurrently. Caps at ~5s. Single-
    // threaded tests are unaffected — they always send before recv.
    for (int waited_us = 0; me->msg_head == NULL && waited_us < 5000000; waited_us += 1000) {
        if (!me->in_use) return -1;
        usleep(1000);
    }
    if (!me->msg_head) return -1;
    if (me->msg_head->len > buf_len) return -1;  // peek-too-small: leave for retry
    loopback_msg *m = loopback_queue_pop(&me->msg_head, &me->msg_tail);
    memcpy(buf, m->bytes, m->len);
    *out_len = m->len;
    free(m);
    return 0;
}

static int loop_close_stream(void *ud, pigeon_stream_handle *h)
{
    (void)ud;
    loopback_stream *me = (loopback_stream *)h;
    loopback_queue_drain(&me->msg_head, &me->msg_tail);
    me->in_use = false;
    return 0;
}

static int loop_send_datagram(void *ud, const uint8_t *data, size_t len)
{
    loopback_endpoint *e = (loopback_endpoint *)ud;
    loopback_endpoint *p = e->peer;
    loopback_msg *m = loopback_msg_new(data, len);
    if (!m) return -1;
    loopback_queue_push(&p->dgram_head, &p->dgram_tail, m);
    return 0;
}

static int loop_recv_datagram(void *ud, uint8_t *buf, size_t buf_len, size_t *out_len)
{
    loopback_endpoint *e = (loopback_endpoint *)ud;
    if (!e->dgram_head) return -1;
    if (e->dgram_head->len > buf_len) return -1;  // peek-too-small: leave for retry
    loopback_msg *m = loopback_queue_pop(&e->dgram_head, &e->dgram_tail);
    memcpy(buf, m->bytes, m->len);
    *out_len = m->len;
    free(m);
    return 0;
}

static void loopback_make_transport(pigeon_transport *t, loopback_endpoint *e)
{
    memset(t, 0, sizeof(*t));
    t->userdata        = e;
    t->open_stream     = loop_open_stream;
    t->accept_stream   = loop_accept_stream;
    t->send_on_stream  = loop_send_on_stream;
    t->recv_on_stream  = loop_recv_on_stream;
    t->close_stream    = loop_close_stream;
    t->send_datagram   = loop_send_datagram;
    t->recv_datagram   = loop_recv_datagram;
}

static void test_session_stream_roundtrip(void)
{
    TEST("session/stream round-trip via loopback transport (chat)");

    // Two endpoints sharing the same symmetric AEAD key.
    uint8_t key[32]; randombytes_buf(key, 32);
    pigeon_channel ch_a, ch_b;
    pigeon_channel_init(&ch_a, key, key, PIGEON_MODE_STRICT);
    pigeon_channel_init(&ch_b, key, key, PIGEON_MODE_STRICT);

    loopback_endpoint ea, eb;
    loopback_endpoint_init(&ea, &eb);
    loopback_endpoint_init(&eb, &ea);

    pigeon_transport ta, tb;
    loopback_make_transport(&ta, &ea);
    loopback_make_transport(&tb, &eb);

    // Under T45 both peers are symmetric: each owns its own pipe and
    // writes the same [varint name-len][name] header.
    pigeon_session sa, sb;
    if (pigeon_session_init(&sa, &ta, &ch_a, NULL, 0) != 0) { FAIL("init A"); return; }
    if (pigeon_session_init(&sb, &tb, &ch_b, NULL, 0) != 0) { FAIL("init B"); return; }

    // A opens "chat". B accepts the next incoming stream and reads its
    // header to confirm the name.
    pigeon_stream sa_chat;
    if (pigeon_session_open_stream(&sa, "chat", &sa_chat) != 0) { FAIL("A open chat"); return; }

    pigeon_stream_handle *bh = NULL;
    if (tb.accept_stream(tb.userdata, &bh) != 0) { FAIL("B accept_stream"); return; }
    uint8_t hdr[PIGEON_MAX_STREAM_HEADER]; size_t hn = 0;
    if (tb.recv_on_stream(tb.userdata, bh, hdr, sizeof(hdr), &hn) != 0) { FAIL("B recv header"); return; }
    char name[32]; size_t nl = 0;
    if (pigeon_wire_stream_header_decode(hdr, hn, name, sizeof(name), &nl) < 0) {
        FAIL("B decode header"); return;
    }
    if (nl != 4 || strcmp(name, "chat") != 0) { FAIL("name mismatch"); return; }

    // Wrap B's accepted handle in a pigeon_stream and bind per-stream
    // ModeStrict AEAD (🎯T53) before send/recv.
    pigeon_stream sb_chat = (pigeon_stream){ .session = &sb, .handle = bh };
    strcpy(sb_chat.name, "chat");
    if (pigeon_stream_bind_aead(&sb_chat) != 0) { FAIL("B bind aead"); return; }

    // A sends "hello" (encrypted), B reads "hello".
    if (pigeon_stream_send(&sa_chat, (const uint8_t *)"hello", 5) != 0) { FAIL("A send"); return; }
    uint8_t buf[32];
    int got = pigeon_stream_recv(&sb_chat, buf, sizeof(buf));
    if (got != 5 || memcmp(buf, "hello", 5) != 0) { FAIL("B recv"); return; }

    // B sends "world" back, A reads "world".
    if (pigeon_stream_send(&sb_chat, (const uint8_t *)"world", 5) != 0) { FAIL("B send"); return; }
    got = pigeon_stream_recv(&sa_chat, buf, sizeof(buf));
    if (got != 5 || memcmp(buf, "world", 5) != 0) { FAIL("A recv"); return; }

    pigeon_stream_close(&sa_chat);
    pigeon_session_close(&sa);
    pigeon_session_close(&sb);

    loopback_endpoint_drain(&ea);
    loopback_endpoint_drain(&eb);

    PASS();
}

static void test_session_datagram_roundtrip(void)
{
    TEST("session/datagram round-trip via loopback (ping channel)");

    uint8_t key[32]; randombytes_buf(key, 32);
    pigeon_channel ch_a, ch_b;
    pigeon_channel_init(&ch_a, key, key, PIGEON_MODE_DATAGRAMS);
    pigeon_channel_init(&ch_b, key, key, PIGEON_MODE_DATAGRAMS);

    loopback_endpoint ea, eb;
    loopback_endpoint_init(&ea, &eb);
    loopback_endpoint_init(&eb, &ea);

    pigeon_transport ta, tb;
    loopback_make_transport(&ta, &ea);
    loopback_make_transport(&tb, &eb);

    pigeon_dgchannel_def chans[] = { { "ping", 1 }, { "metric", 2 } };

    pigeon_session sa, sb;
    if (pigeon_session_init(&sa, &ta, &ch_a, chans, 2) != 0) { FAIL("init A"); return; }
    if (pigeon_session_init(&sb, &tb, &ch_b, chans, 2) != 0) { FAIL("init B"); return; }

    pigeon_datagram da_ping, db_ping, da_metric;
    if (pigeon_session_get_datagram(&sa, "ping",   &da_ping)   != 0) { FAIL("A ping"); return; }
    if (pigeon_session_get_datagram(&sb, "ping",   &db_ping)   != 0) { FAIL("B ping"); return; }
    if (pigeon_session_get_datagram(&sa, "metric", &da_metric) != 0) { FAIL("A metric"); return; }

    // Under T45 the datagram wire is just AEAD([varint channel-id]
    // [payload]) — both peers are symmetric, no tag prefix to strip.
    if (pigeon_datagram_send(&da_ping, (const uint8_t *)"p1", 2) != 0) { FAIL("A send ping"); return; }
    uint8_t buf[32];
    int got = pigeon_datagram_recv(&db_ping, buf, sizeof(buf));
    if (got != 2 || memcmp(buf, "p1", 2) != 0) { FAIL("B recv ping"); return; }

    pigeon_session_close(&sa);
    pigeon_session_close(&sb);

    loopback_endpoint_drain(&ea);
    loopback_endpoint_drain(&eb);

    PASS();
}

// --- Activation handshake driver (T32.1) ---
//
// Three tests using the in-test loopback transport defined above:
//   * known device path: backend resolves the device, both machines
//     reach SessionActive.
//   * unknown device path: backend's resolve fails, machine takes
//     the device_unknown branch and ends up at Idle.
//   * wire roundtrip: encode/decode auth_request and auth_ok across
//     accepted and rejected variants; byte vectors match the Go
//     side's TestActivation_WireRoundtrip outputs.

typedef struct {
    bool match;          // resolve returns 0 when device_id matches
    const char *want_id; // expected device ID (NULL = match anything)
} activation_resolve_ctx;

static int activation_resolve(void *ud, const char *device_id, void *out_record)
{
    activation_resolve_ctx *r = (activation_resolve_ctx *)ud;
    if (r->want_id == NULL || strcmp(device_id, r->want_id) == 0) {
        if (r->match) {
            // Zero out the record — the activation tests don't
            // exercise it. Real callers fill it from their lookup.
            (void)out_record;
            return 0;
        }
    }
    return -1;
}

// Spawn a transport pair and a single stream on each side that maps
// to the peer. Used by both activation tests below.
static void activation_pair_setup(loopback_endpoint *eb,
                                  loopback_endpoint *ec,
                                  pigeon_transport *tb,
                                  pigeon_transport *tc,
                                  pigeon_stream_handle **out_backend_stream,
                                  pigeon_stream_handle **out_client_stream)
{
    loopback_endpoint_init(eb, ec);
    loopback_endpoint_init(ec, eb);
    loopback_make_transport(tb, eb);
    loopback_make_transport(tc, ec);
    // Client opens the activation stream; backend accepts.
    pigeon_stream_handle *cs = NULL;
    if (tc->open_stream(tc->userdata, &cs) != 0) {
        FAIL("activation: open client stream"); return;
    }
    pigeon_stream_handle *bs = NULL;
    if (tb->accept_stream(tb->userdata, &bs) != 0) {
        FAIL("activation: accept backend stream"); return;
    }
    *out_client_stream  = cs;
    *out_backend_stream = bs;
}

// Backend-side thread argument bundle. The loopback transport's
// recv_on_stream returns -1 immediately on empty (no blocking
// semantics), so the activation tests have to drive backend and
// client concurrently. The backend runs in its own thread; the
// main thread runs the client.
typedef struct {
    pigeon_transport       *transport;
    pigeon_stream_handle   *stream;
    activation_resolve_ctx *rctx;
    pigeon_backend_machine *machine;
    char                   *device_id_seen;
    size_t                  device_id_cap;
    pigeon_pairing_record  *record;
    uint8_t                 nonce[PIGEON_AUTH_NONCE_LEN];
    int                     rc;
} backend_thread_args;

static void *run_backend_thread(void *p)
{
    backend_thread_args *a = (backend_thread_args *)p;
    a->rc = pigeon_run_backend_activation(a->transport, a->stream,
                                          activation_resolve, a->rctx,
                                          a->machine,
                                          a->device_id_seen, a->device_id_cap,
                                          a->record, a->nonce,
                                          /*out_route=*/NULL, 0);
    return NULL;
}

static void test_activation_known_device(void)
{
    TEST("activation: known device → SessionActive on both sides");
    loopback_endpoint eb, ec;
    pigeon_transport tb, tc;
    pigeon_stream_handle *bs = NULL, *cs = NULL;
    activation_pair_setup(&eb, &ec, &tb, &tc, &bs, &cs);
    if (cs == NULL || bs == NULL) { loopback_endpoint_drain(&eb); loopback_endpoint_drain(&ec); return; }

    activation_resolve_ctx rctx = { .match = true, .want_id = "device-known-1" };

    pigeon_backend_machine bm;
    pigeon_client_machine  cm;
    char device_id_seen[PIGEON_AUTH_MAX_DEVICE_ID + 1];
    pigeon_pairing_record record;

    backend_thread_args args = {
        .transport = &tb, .stream = bs, .rctx = &rctx,
        .machine = &bm,
        .device_id_seen = device_id_seen, .device_id_cap = sizeof(device_id_seen),
        .record = &record, .rc = 0,
    };
    pthread_t tid;
    if (pthread_create(&tid, NULL, run_backend_thread, &args) != 0) {
        FAIL("pthread_create"); return;
    }

    uint8_t cnonce[PIGEON_AUTH_NONCE_LEN];
    int crc = pigeon_run_client_activation(&tc, cs, "device-known-1", /*route=*/"",
                                           &cm, NULL, 0, cnonce);
    pthread_join(tid, NULL);

    if (args.rc != 0) { FAIL("backend activation"); return; }
    if (crc != 0)     { FAIL("client activation"); return; }
    if (strcmp(device_id_seen, "device-known-1") != 0) {
        FAIL("backend received wrong device id"); return;
    }
    if (bm.state != PIGEON_BACKEND_SESSION_ACTIVE) {
        FAIL("backend machine not at SessionActive"); return;
    }
    if (cm.state != PIGEON_CLIENT_SESSION_ACTIVE) {
        FAIL("client machine not at SessionActive"); return;
    }
    loopback_endpoint_drain(&eb);
    loopback_endpoint_drain(&ec);
    PASS();
}

static void test_activation_unknown_device(void)
{
    TEST("activation: unknown device → backend at Idle");
    loopback_endpoint eb, ec;
    pigeon_transport tb, tc;
    pigeon_stream_handle *bs = NULL, *cs = NULL;
    activation_pair_setup(&eb, &ec, &tb, &tc, &bs, &cs);
    if (cs == NULL || bs == NULL) { loopback_endpoint_drain(&eb); loopback_endpoint_drain(&ec); return; }

    activation_resolve_ctx rctx = { .match = false, .want_id = NULL };

    pigeon_backend_machine bm;
    pigeon_client_machine  cm;
    char device_id_seen[PIGEON_AUTH_MAX_DEVICE_ID + 1];
    pigeon_pairing_record record;
    char client_reason[PIGEON_AUTH_MAX_REASON];

    backend_thread_args args = {
        .transport = &tb, .stream = bs, .rctx = &rctx,
        .machine = &bm,
        .device_id_seen = device_id_seen, .device_id_cap = sizeof(device_id_seen),
        .record = &record, .rc = 0,
    };
    pthread_t tid;
    if (pthread_create(&tid, NULL, run_backend_thread, &args) != 0) {
        FAIL("pthread_create"); return;
    }

    uint8_t cnonce[PIGEON_AUTH_NONCE_LEN];
    int crc = pigeon_run_client_activation(&tc, cs, "device-stranger", /*route=*/"",
                                           &cm, client_reason, sizeof(client_reason), cnonce);
    pthread_join(tid, NULL);

    if (args.rc != 1) { FAIL("backend should report tri-value 1 (rejected)"); return; }
    if (crc != -1) { FAIL("client should report rejection (-1)"); return; }
    if (bm.state != PIGEON_BACKEND_IDLE) {
        FAIL("backend machine not at Idle after device_unknown branch"); return;
    }
    if (strcmp(client_reason, "unknown client") != 0) {
        FAIL("client reason mismatch"); return;
    }
    loopback_endpoint_drain(&eb);
    loopback_endpoint_drain(&ec);
    PASS();
}

// --- Multi-client listener (T45 remote-Listen L1) ---
//
// Under T45 each accepted client rides its OWN end-to-end pipe. The
// listener holds a register control connection plus a dial-listen
// callback that yields a fresh bridged listen connection per accept.
// These tests wire a loopback dialer: each accept pops a pre-staged
// loopback pair, the dialer returns the backend side, and a client
// thread drives the matching client side through activation.

// Listener resolver: accept any device id matching one of two
// expected values. PairingRecord uses a real generated keypair so
// make_session can run X25519 ECDH (libsodium's crypto_scalarmult
// rejects all-zero inputs). The listener test exercises the dial +
// activation flow, not the AEAD round-trip — the keys just have to
// be a valid curve point pair.
typedef struct {
    const char     *want_id_1;
    const char     *want_id_2;
    pigeon_keypair  kp;
} listener_resolve_ctx;

static int listener_resolve(void *ud, const char *device_id, void *out_record)
{
    listener_resolve_ctx *r = (listener_resolve_ctx *)ud;
    if (strcmp(device_id, r->want_id_1) != 0
            && strcmp(device_id, r->want_id_2) != 0) {
        return -1;
    }
    pigeon_pairing_record *rec = (pigeon_pairing_record *)out_record;
    memset(rec, 0, sizeof(*rec));
    memcpy(rec->local_private_key, r->kp.private_key, 32);
    memcpy(rec->local_public_key,  r->kp.public_key,  32);
    memcpy(rec->peer_public_key,   r->kp.public_key,  32);
    return 0;
}

// One staged per-client pipe: a loopback pair whose backend side the
// dialer hands to the listener and whose client side a client thread
// drives.
#define LISTENER_TEST_MAX_PIPES 4
typedef struct {
    loopback_endpoint backend_ep;
    loopback_endpoint client_ep;
    pigeon_transport  client_tr;
    bool              in_use;
} listener_test_pipe;

// Dialer context: a fixed pool of staged pipes; each accept consumes
// the next one.
typedef struct {
    listener_test_pipe pipes[LISTENER_TEST_MAX_PIPES];
    int                next;
    int                count;
} listener_dial_ctx;

// pigeon_listen_dialer: hand the backend side of the next staged pipe
// to the listener. The owner cookie is the pipe pointer (no teardown
// needed — the pipes live on the test's stack).
static int listener_test_dial(void *userdata,
                              pigeon_transport *out_transport,
                              pigeon_stream_handle **out_primary,
                              void **out_owner)
{
    listener_dial_ctx *d = (listener_dial_ctx *)userdata;
    if (d->next >= d->count) return -1;
    listener_test_pipe *p = &d->pipes[d->next++];
    // The client opened the primary stream before we dial; accept it on
    // the backend side to recover the bridged primary handle.
    pigeon_stream_handle *primary = NULL;
    if (p->backend_ep.peer == NULL) return -1;
    pigeon_transport tb;
    loopback_make_transport(&tb, &p->backend_ep);
    if (tb.accept_stream(tb.userdata, &primary) != 0) return -1;
    *out_transport = tb;
    *out_primary   = primary;
    *out_owner     = p; // pipe lives on the stack — owner hooks are no-ops.
    return 0;
}

static void listener_test_owner_noop(void *owner) { (void)owner; }

// Drive one client through activation over its own pipe: open the
// primary stream and run the client activation driver. No stream
// header — activation mode's auth_request is the first message.
typedef struct {
    pigeon_transport *transport;
    const char       *device_id;
    int               rc;
} listener_client_args;

static void *run_listener_client(void *p)
{
    listener_client_args *a = (listener_client_args *)p;
    pigeon_stream_handle *h = NULL;
    if (a->transport->open_stream(a->transport->userdata, &h) != 0) {
        a->rc = -1;
        return NULL;
    }
    pigeon_client_machine cm;
    uint8_t cnonce[PIGEON_AUTH_NONCE_LEN];
    a->rc = pigeon_run_client_activation(a->transport, h, a->device_id, /*route=*/"",
                                         &cm, NULL, 0, cnonce);
    return NULL;
}

// Stage `n` pipes in the dial context and fill the client transports.
static void stage_pipes(listener_dial_ctx *d, int n)
{
    memset(d, 0, sizeof(*d));
    d->count = n;
    d->next = 0;
    for (int i = 0; i < n; i++) {
        listener_test_pipe *p = &d->pipes[i];
        loopback_endpoint_init(&p->backend_ep, &p->client_ep);
        loopback_endpoint_init(&p->client_ep, &p->backend_ep);
        loopback_make_transport(&p->client_tr, &p->client_ep);
        p->in_use = true;
    }
}

static void drain_pipes(listener_dial_ctx *d)
{
    for (int i = 0; i < d->count; i++) {
        loopback_endpoint_drain(&d->pipes[i].backend_ep);
        loopback_endpoint_drain(&d->pipes[i].client_ep);
    }
}

static void test_listener_two_clients(void)
{
    TEST("listener: two clients, each on its own bridged pipe (T45)");

    // Register control connection (held open; no traffic). A plain
    // loopback endpoint pair stands in for it.
    loopback_endpoint ctrl_a, ctrl_b;
    loopback_endpoint_init(&ctrl_a, &ctrl_b);
    loopback_endpoint_init(&ctrl_b, &ctrl_a);
    pigeon_transport ctrl_tr;
    loopback_make_transport(&ctrl_tr, &ctrl_a);

    listener_resolve_ctx rctx = {
        .want_id_1 = "device-a",
        .want_id_2 = "device-b",
    };
    if (pigeon_generate_keypair(&rctx.kp) != 0) { FAIL("keypair"); return; }

    listener_dial_ctx dctx;
    stage_pipes(&dctx, 2);

    pigeon_listener *l = NULL;
    if (pigeon_listener_init(&l, &ctrl_tr, "backend-instance",
                             listener_test_dial, &dctx,
                             listener_test_owner_noop, listener_test_owner_noop,
                             listener_resolve, &rctx,
                             NULL, 0) != 0) {
        FAIL("listener_init"); return;
    }
    if (strcmp(pigeon_listener_instance_id(l), "backend-instance") != 0) {
        FAIL("instance id mismatch"); return;
    }

    // Drive both clients concurrently; the listener accepts them one at
    // a time (one per staged pipe) on this thread.
    listener_client_args ca = { &dctx.pipes[0].client_tr, "device-a", 0 };
    listener_client_args cb = { &dctx.pipes[1].client_tr, "device-b", 0 };
    pthread_t tid_a, tid_b;
    if (pthread_create(&tid_a, NULL, run_listener_client, &ca) != 0) {
        FAIL("pthread_create A"); return;
    }
    if (pthread_create(&tid_b, NULL, run_listener_client, &cb) != 0) {
        FAIL("pthread_create B"); return;
    }

    pigeon_session *sessions[2] = { NULL, NULL };
    if (pigeon_listener_accept(l, &sessions[0]) != 0) { FAIL("accept 0"); return; }
    if (pigeon_listener_accept(l, &sessions[1]) != 0) { FAIL("accept 1"); return; }

    pthread_join(tid_a, NULL);
    pthread_join(tid_b, NULL);

    if (ca.rc != 0)   { FAIL("client A activation failed"); return; }
    if (cb.rc != 0)   { FAIL("client B activation failed"); return; }
    if (!sessions[0] || !sessions[1]) { FAIL("missing session"); return; }
    if (sessions[0] == sessions[1])    { FAIL("same session twice"); return; }

    // Each session rides its own transport — no shared demux. Confirm
    // the two sessions bind to different (backend-side) endpoints.
    if (sessions[0]->transport.userdata == sessions[1]->transport.userdata) {
        FAIL("sessions share a transport"); return;
    }

    pigeon_session_close(sessions[0]); free(sessions[0]);
    pigeon_session_close(sessions[1]); free(sessions[1]);
    pigeon_listener_close(l);
    drain_pipes(&dctx);
    loopback_endpoint_drain(&ctrl_a);
    loopback_endpoint_drain(&ctrl_b);
    PASS();
}

// Verify a peer-opened sub-stream on a session's own pipe is delivered
// by pigeon_session_accept_incoming_stream (which now reads directly
// from the session's transport — no listener demux).
static void test_listener_substream_demux(void)
{
    TEST("listener: sub-stream on a session's own pipe is delivered (T45)");

    loopback_endpoint ctrl_a, ctrl_b;
    loopback_endpoint_init(&ctrl_a, &ctrl_b);
    loopback_endpoint_init(&ctrl_b, &ctrl_a);
    pigeon_transport ctrl_tr;
    loopback_make_transport(&ctrl_tr, &ctrl_a);

    listener_resolve_ctx rctx = {
        .want_id_1 = "device-first",
        .want_id_2 = "device-second",
    };
    if (pigeon_generate_keypair(&rctx.kp) != 0) { FAIL("keypair"); return; }

    listener_dial_ctx dctx;
    stage_pipes(&dctx, 1);

    pigeon_listener *l = NULL;
    if (pigeon_listener_init(&l, &ctrl_tr, "demux-listener",
                             listener_test_dial, &dctx,
                             listener_test_owner_noop, listener_test_owner_noop,
                             listener_resolve, &rctx,
                             NULL, 0) != 0) {
        FAIL("listener_init"); return;
    }

    // Client A connects and activates.
    listener_client_args ca = { &dctx.pipes[0].client_tr, "device-first", 0 };
    pthread_t tid_a;
    if (pthread_create(&tid_a, NULL, run_listener_client, &ca) != 0) {
        FAIL("pthread_create A"); return;
    }
    pigeon_session *sess = NULL;
    if (pigeon_listener_accept(l, &sess) != 0) { FAIL("accept"); return; }
    pthread_join(tid_a, NULL);
    if (ca.rc != 0) { FAIL("client A activation"); return; }
    if (!sess)      { FAIL("missing session"); return; }

    // The client opens a "logs" sub-stream on its own pipe; the session
    // should surface it via accept_incoming_stream.
    pigeon_stream_handle *sub = NULL;
    if (dctx.pipes[0].client_tr.open_stream(dctx.pipes[0].client_tr.userdata, &sub) != 0) {
        FAIL("open sub"); return;
    }
    uint8_t hdr[PIGEON_MAX_STREAM_HEADER];
    int hn = pigeon_wire_stream_header_encode("logs", strlen("logs"),
                                              hdr, sizeof(hdr));
    if (hn < 0) { FAIL("encode sub header"); return; }
    if (dctx.pipes[0].client_tr.send_on_stream(dctx.pipes[0].client_tr.userdata,
                                               sub, hdr, (size_t)hn) != 0) {
        FAIL("send sub header"); return;
    }

    pigeon_stream out_stream;
    if (pigeon_session_accept_incoming_stream(sess, "logs", &out_stream) != 0) {
        FAIL("sub-stream not delivered"); return;
    }
    if (strcmp(out_stream.name, "logs") != 0) {
        FAIL("sub-stream wrong name"); return;
    }

    pigeon_session_close(sess); free(sess);
    pigeon_listener_close(l);
    drain_pipes(&dctx);
    loopback_endpoint_drain(&ctrl_a);
    loopback_endpoint_drain(&ctrl_b);
    PASS();
}

// --- pigeon_connect / pigeon_session_primary (T32.3) ---
//
// Fake-backend in-process loopback test. Mirrors the T32.1 activation
// test plumbing: pair two loopback endpoints, client opens the primary
// stream, backend accepts. The backend thread reads the empty-name
// primary header that pigeon_connect_on_transport will write, then runs
// pigeon_run_backend_activation; the client thread calls
// pigeon_connect_on_transport, which writes the header + runs the
// client activation handshake + derives the AEAD channel + initialises
// the pigeon_session. After both sides complete, we exercise the
// resulting session end-to-end by opening a named sub-stream and
// sending an AEAD-encrypted application message.
//
// This test does NOT depend on T32.2's pigeon_register / Listener — the
// backend half is open-coded here so T32.3 lands testable on its own
// worktree. Once T32.2 lands, a follow-up will switch this to drive
// pigeon_register on the backend side.

typedef struct {
    pigeon_transport       *transport;
    pigeon_stream_handle   *stream;
    activation_resolve_ctx *rctx;
    pigeon_backend_machine *machine;
    char                   *device_id_seen;
    size_t                  device_id_cap;
    pigeon_pairing_record  *record;
    uint8_t                 nonce[PIGEON_AUTH_NONCE_LEN];
    int                     rc;
} backend_connect_thread_args;

static void *run_backend_connect_thread(void *p)
{
    backend_connect_thread_args *a = (backend_connect_thread_args *)p;

    // Under T45 activation mode writes no primary header — the client's
    // auth_request is the first message on the bridged pipe. Run the
    // activation handshake directly.
    a->rc = pigeon_run_backend_activation(a->transport, a->stream,
                                          activation_resolve, a->rctx,
                                          a->machine,
                                          a->device_id_seen, a->device_id_cap,
                                          a->record, a->nonce,
                                          /*out_route=*/NULL, 0);
    return NULL;
}

// Build a pigeon_pairing_record pair (backend + client) with matching
// X25519 keys so DeriveChannel produces identical send/recv keys on
// both sides. Returns 0 on success.
static int build_paired_records(pigeon_pairing_record *backend_rec,
                                pigeon_pairing_record *client_rec)
{
    memset(backend_rec, 0, sizeof(*backend_rec));
    memset(client_rec,  0, sizeof(*client_rec));
    pigeon_keypair kb, kc;
    if (pigeon_generate_keypair(&kb) != 0) return -1;
    if (pigeon_generate_keypair(&kc) != 0) return -1;
    memcpy(backend_rec->local_private_key, kb.private_key, 32);
    memcpy(backend_rec->local_public_key,  kb.public_key,  32);
    memcpy(backend_rec->peer_public_key,   kc.public_key,  32);
    memcpy(client_rec->local_private_key,  kc.private_key, 32);
    memcpy(client_rec->local_public_key,   kc.public_key,  32);
    memcpy(client_rec->peer_public_key,    kb.public_key,  32);
    snprintf(backend_rec->peer_instance_id, sizeof(backend_rec->peer_instance_id),
             "client-device-1");
    snprintf(client_rec->peer_instance_id,  sizeof(client_rec->peer_instance_id),
             "backend-1");
    snprintf(backend_rec->relay_url, sizeof(backend_rec->relay_url), "test://relay");
    snprintf(client_rec->relay_url,  sizeof(client_rec->relay_url),  "test://relay");
    return 0;
}

static void test_pigeon_connect_loopback(void)
{
    TEST("pigeon_connect_on_transport: client ↔ fake-backend round-trip");

    // 1. Loopback transport pair + primary stream.
    loopback_endpoint eb, ec;
    pigeon_transport tb, tc;
    pigeon_stream_handle *bs = NULL, *cs = NULL;
    activation_pair_setup(&eb, &ec, &tb, &tc, &bs, &cs);
    if (cs == NULL || bs == NULL) { loopback_endpoint_drain(&eb); loopback_endpoint_drain(&ec); return; }

    // 2. Matching pairing records on both sides.
    pigeon_pairing_record backend_rec, client_rec;
    if (build_paired_records(&backend_rec, &client_rec) != 0) {
        FAIL("build paired records"); return;
    }

    activation_resolve_ctx rctx = { .match = true, .want_id = "client-device-1" };
    pigeon_backend_machine bm;
    char device_id_seen[PIGEON_AUTH_MAX_DEVICE_ID + 1] = {0};

    backend_connect_thread_args bargs = {
        .transport      = &tb,
        .stream         = bs,
        .rctx           = &rctx,
        .machine        = &bm,
        .device_id_seen = device_id_seen,
        .device_id_cap  = sizeof(device_id_seen),
        .record         = &backend_rec,
        .rc             = 0,
    };
    pthread_t tid;
    if (pthread_create(&tid, NULL, run_backend_connect_thread, &bargs) != 0) {
        FAIL("pthread_create"); return;
    }

    // 3. Client side: pigeon_connect_on_transport runs client activation
    //    + channel derive + session init (T45: no primary header).
    pigeon_dgchannel_def dgs[] = { { "ping", 1 } };
    pigeon_session csess;
    int crc = pigeon_connect_on_transport(&tc, cs,
                                          /*peer_instance_id=*/"backend-1",
                                          /*device_id=*/"client-device-1",
                                          &client_rec,
                                          dgs, 1,
                                          &csess);
    pthread_join(tid, NULL);

    if (bargs.rc != 0)          { FAIL("backend: activation"); return; }
    if (crc != 0)               { FAIL("client: pigeon_connect_on_transport"); return; }
    if (strcmp(device_id_seen, "client-device-1") != 0) {
        FAIL("backend: wrong device id seen"); return;
    }
    if (csess.primary != cs)    { FAIL("session.primary not bound"); return; }
    if (!csess.channel.established) { FAIL("session.channel not established"); return; }

    // 4. pigeon_session_primary smoke test.
    pigeon_stream prim;
    if (pigeon_session_primary(&csess, &prim) != 0) { FAIL("session_primary"); return; }
    if (prim.handle != cs) { FAIL("session_primary handle mismatch"); return; }
    if (prim.name[0] != '\0') { FAIL("session_primary name should be empty"); return; }

    // 5. End-to-end: open a sub-stream "chat", send AEAD-encrypted msg.
    //    We open the stream on the client side, then accept + decode the
    //    header on the backend side, then read & decrypt one application
    //    message using a transient pigeon_session anchored on the
    //    backend's (matching) AEAD channel.
    pigeon_stream chat;
    if (pigeon_session_open_stream(&csess, "chat", &chat) != 0) {
        FAIL("open chat"); return;
    }

    // Backend-side mirror: derive backend channel (send=backend->client,
    // recv=client->backend; reverse of the client side's labels) and
    // wrap it as a pigeon_session for the recv path.
    // Fold the per-session nonce the backend decoded from the client's
    // auth_request into the HKDF info (🎯T44.1): info = label || nonce.
    uint8_t b_send[32], b_recv[32];
    uint8_t b_send_info[15 + PIGEON_AUTH_NONCE_LEN];
    uint8_t b_recv_info[15 + PIGEON_AUTH_NONCE_LEN];
    memcpy(b_send_info, "backend->client", 15);
    memcpy(b_send_info + 15, bargs.nonce, PIGEON_AUTH_NONCE_LEN);
    memcpy(b_recv_info, "client->backend", 15);
    memcpy(b_recv_info + 15, bargs.nonce, PIGEON_AUTH_NONCE_LEN);
    if (pigeon_derive_session_key(backend_rec.local_private_key,
                                  backend_rec.peer_public_key,
                                  b_send_info, sizeof(b_send_info),
                                  b_send) != 0) { FAIL("derive b_send"); return; }
    if (pigeon_derive_session_key(backend_rec.local_private_key,
                                  backend_rec.peer_public_key,
                                  b_recv_info, sizeof(b_recv_info),
                                  b_recv) != 0) { FAIL("derive b_recv"); return; }
    pigeon_channel b_chan;
    pigeon_channel_init(&b_chan, b_send, b_recv, PIGEON_MODE_STRICT);

    pigeon_session bsess;
    if (pigeon_session_init(&bsess, &tb, &b_chan,
                            dgs, 1) != 0) { FAIL("backend session_init"); return; }

    // Accept the chat stream the client opened, read and decode its header.
    pigeon_stream_handle *b_chat = NULL;
    if (tb.accept_stream(tb.userdata, &b_chat) != 0) { FAIL("backend accept chat"); return; }
    uint8_t hbuf[PIGEON_MAX_STREAM_HEADER]; size_t hlen = 0;
    if (tb.recv_on_stream(tb.userdata, b_chat, hbuf, sizeof(hbuf), &hlen) != 0) {
        FAIL("backend read chat header"); return;
    }
    // T45: header is just [varint name-len][name].
    char chat_name[PIGEON_MAX_NAME_LEN]; size_t cnl = 0;
    if (pigeon_wire_stream_header_decode(hbuf, hlen, chat_name, sizeof(chat_name), &cnl) < 0) {
        FAIL("backend decode chat header"); return;
    }
    if (cnl != 4 || strcmp(chat_name, "chat") != 0) { FAIL("chat name mismatch"); return; }

    // Client sends "hello", backend reads decrypted.
    if (pigeon_stream_send(&chat, (const uint8_t *)"hello", 5) != 0) {
        FAIL("client chat send"); return;
    }
    pigeon_stream b_chat_wrap = { .session = &bsess, .handle = b_chat };
    strcpy(b_chat_wrap.name, "chat");
    if (pigeon_stream_bind_aead(&b_chat_wrap) != 0) { FAIL("backend bind aead"); return; }
    uint8_t buf[32];
    int got = pigeon_stream_recv(&b_chat_wrap, buf, sizeof(buf));
    if (got != 5 || memcmp(buf, "hello", 5) != 0) { FAIL("backend chat recv"); return; }

    pigeon_stream_close(&chat);
    pigeon_session_close(&csess);
    pigeon_session_close(&bsess);

    loopback_endpoint_drain(&eb);
    loopback_endpoint_drain(&ec);
    PASS();
}

static void test_pigeon_connect_pairing_mode(void)
{
    TEST("pigeon_connect_on_transport: pairing mode (no record, no activation)");

    loopback_endpoint eb, ec;
    pigeon_transport tb, tc;
    pigeon_stream_handle *bs = NULL, *cs = NULL;
    activation_pair_setup(&eb, &ec, &tb, &tc, &bs, &cs);
    if (cs == NULL || bs == NULL) { loopback_endpoint_drain(&eb); loopback_endpoint_drain(&ec); return; }

    // In pairing mode the client passes record=NULL, device_id=NULL.
    // No activation handshake runs; channel.established stays false.
    // The backend side wouldn't run pigeon_run_backend_activation either
    // (caller switches to ceremony mode), so for this test we only
    // verify the client-side post-state and that Session.Primary() can
    // round-trip plaintext (the pairing ceremony does its own crypto).
    pigeon_session csess;
    if (pigeon_connect_on_transport(&tc, cs,
                                    /*peer_instance_id=*/"backend-1",
                                    /*device_id=*/NULL,
                                    /*record=*/NULL,
                                    NULL, 0,
                                    &csess) != 0) {
        FAIL("pigeon_connect_on_transport(pairing-mode)"); return;
    }
    if (csess.channel.established) { FAIL("channel should NOT be established"); return; }
    if (csess.primary != cs)       { FAIL("primary handle not bound"); return; }

    // Under T45 pairing mode, the client writes one empty "arrival
    // marker" message on the primary (Connect's writeMessage(nil)) so
    // the backend can gate Session creation on a real client match.
    // Read it on the backend side: it is a zero-length message.
    uint8_t hdr[PIGEON_MAX_STREAM_HEADER]; size_t hn = 0;
    if (tb.recv_on_stream(tb.userdata, bs, hdr, sizeof(hdr), &hn) != 0) {
        FAIL("backend read arrival marker"); return;
    }
    if (hn != 0) { FAIL("arrival marker should be empty"); return; }

    // Drive plaintext over the primary via Session.Primary() (pairing-
    // ceremony idiom). Backend stays raw — read straight off the
    // transport with no AEAD.
    pigeon_stream prim;
    if (pigeon_session_primary(&csess, &prim) != 0) { FAIL("session_primary"); return; }
    if (pigeon_stream_send(&prim, (const uint8_t *)"hi", 2) != 0) {
        FAIL("primary plaintext send"); return;
    }
    uint8_t buf[32]; size_t got = 0;
    if (tb.recv_on_stream(tb.userdata, bs, buf, sizeof(buf), &got) != 0) {
        FAIL("backend raw recv on primary"); return;
    }
    if (got != 2 || memcmp(buf, "hi", 2) != 0) { FAIL("primary plaintext mismatch"); return; }

    pigeon_session_close(&csess);
    loopback_endpoint_drain(&eb);
    loopback_endpoint_drain(&ec);
    PASS();
}

static void test_activation_wire_roundtrip(void)
{
    TEST("activation: wire roundtrip (auth_request + auth_ok variants)");
    uint8_t buf[256];

    // auth_request: tag 0x01, varint(len), device-id, 16-byte nonce.
    uint8_t nonce_in[PIGEON_AUTH_NONCE_LEN];
    for (size_t i = 0; i < PIGEON_AUTH_NONCE_LEN; i++) nonce_in[i] = (uint8_t)(i + 1);
    int n = pigeon_encode_auth_request("device-wire-1", nonce_in, "gs-42", buf, sizeof(buf));
    if (n < 0) { FAIL("encode auth_request"); return; }
    char id_out[64];
    uint8_t nonce_out[PIGEON_AUTH_NONCE_LEN];
    char route_out[PIGEON_AUTH_MAX_ROUTE + 1];
    if (pigeon_decode_auth_request(buf, (size_t)n, id_out, sizeof(id_out), nonce_out,
                                   route_out, sizeof(route_out)) != 0) {
        FAIL("decode auth_request"); return;
    }
    if (strcmp(id_out, "device-wire-1") != 0) {
        FAIL("auth_request roundtrip mismatch"); return;
    }
    if (memcmp(nonce_in, nonce_out, PIGEON_AUTH_NONCE_LEN) != 0) {
        FAIL("auth_request nonce roundtrip mismatch"); return;
    }
    if (strcmp(route_out, "gs-42") != 0) {
        FAIL("auth_request route roundtrip mismatch"); return;
    }

    // auth_ok accepted: tag 0x02, 0x01.
    n = pigeon_encode_auth_ok(true, NULL, buf, sizeof(buf));
    if (n != 2 || buf[0] != 0x02 || buf[1] != 0x01) {
        FAIL("encode auth_ok accepted"); return;
    }
    bool got_ok = false;
    char reason[64];
    if (pigeon_decode_auth_ok(buf, (size_t)n, &got_ok, reason, sizeof(reason)) != 0) {
        FAIL("decode auth_ok accepted"); return;
    }
    if (!got_ok || reason[0] != '\0') { FAIL("decoded accepted: wrong fields"); return; }

    // auth_ok rejected: tag 0x02, 0x00, varint, reason.
    n = pigeon_encode_auth_ok(false, "unknown client", buf, sizeof(buf));
    if (n < 0) { FAIL("encode auth_ok rejected"); return; }
    if (pigeon_decode_auth_ok(buf, (size_t)n, &got_ok, reason, sizeof(reason)) != 0) {
        FAIL("decode auth_ok rejected"); return;
    }
    if (got_ok || strcmp(reason, "unknown client") != 0) {
        FAIL("decoded rejected: wrong fields"); return;
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
    test_uvarint();
    test_stream_header();
    test_datagram_framing();
    test_session_stream_roundtrip();
    test_session_datagram_roundtrip();
    test_send_recv_unencrypted();
    test_send_recv_encrypted();
    test_send_recv_datagram_unencrypted();
    test_send_recv_datagram_encrypted();
    test_activation_wire_roundtrip();
    test_activation_known_device();
    test_activation_unknown_device();
    test_listener_two_clients();
    test_listener_substream_demux();
    test_pigeon_connect_loopback();
    test_pigeon_connect_pairing_mode();

    printf("\n%d/%d tests passed\n", tests_passed, tests_run);
    return tests_passed == tests_run ? 0 : 1;
}

// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0
//
// Pigeon C client library — amalgamated source.
// Compile with -DPIGEON_CRYPTO_LIBSODIUM and link -lsodium.

#include "pigeon.h"
#include <string.h>

// --- Generated state machine ---




void pigeon_server_machine_init(pigeon_server_machine *m)
{
	memset(m, 0, sizeof(*m));
	m->state = PIGEON_SERVER_IDLE;
	m->current_token = "none";
	m->server_ecdh_pub = "none";
	m->received_client_pub = "none";
	m->code_attempts = 0;
	m->device_secret = "none";
	m->received_device_id = "none";
	m->received_auth_nonce = "none";
}

int pigeon_server_handle_message(pigeon_server_machine *m, pairing_ceremony_msg_type msg)
{
	if (m->state == PIGEON_SERVER_IDLE && msg == PIGEON_MSG_PAIR_BEGIN) {
		if (m->actions[PIGEON_ACTION_GENERATE_TOKEN]) {
			int err = m->actions[PIGEON_ACTION_GENERATE_TOKEN](m->userdata);
			if (err) return -err;
		}
		m->current_token = "tok_1";
		if (m->on_change) m->on_change("current_token", m->userdata);
		// active_tokens: active_tokens \union {"tok_1"} (set by action)
		m->state = PIGEON_SERVER_GENERATE_TOKEN;
		return 1;
	}
	if (m->state == PIGEON_SERVER_WAITING_FOR_CLIENT && msg == PIGEON_MSG_PAIR_HELLO && m->guards[PIGEON_GUARD_TOKEN_VALID] && m->guards[PIGEON_GUARD_TOKEN_VALID](m->userdata)) {
		if (m->actions[PIGEON_ACTION_DERIVE_SECRET]) {
			int err = m->actions[PIGEON_ACTION_DERIVE_SECRET](m->userdata);
			if (err) return -err;
		}
		// received_client_pub: recv_msg.pubkey (set by action)
		m->server_ecdh_pub = "server_pub";
		if (m->on_change) m->on_change("server_ecdh_pub", m->userdata);
		// server_shared_key: DeriveKey("server_pub", recv_msg.pubkey) (set by action)
		// server_code: DeriveCode("server_pub", recv_msg.pubkey) (set by action)
		m->state = PIGEON_SERVER_DERIVE_SECRET;
		return 1;
	}
	if (m->state == PIGEON_SERVER_WAITING_FOR_CLIENT && msg == PIGEON_MSG_PAIR_HELLO && m->guards[PIGEON_GUARD_TOKEN_INVALID] && m->guards[PIGEON_GUARD_TOKEN_INVALID](m->userdata)) {
		m->state = PIGEON_SERVER_IDLE;
		return 1;
	}
	if (m->state == PIGEON_SERVER_WAITING_FOR_CODE && msg == PIGEON_MSG_CODE_SUBMIT) {
		// received_code: recv_msg.code (set by action)
		m->state = PIGEON_SERVER_VALIDATE_CODE;
		return 1;
	}
	if (m->state == PIGEON_SERVER_PAIRED && msg == PIGEON_MSG_AUTH_REQUEST) {
		// received_device_id: recv_msg.device_id (set by action)
		// received_auth_nonce: recv_msg.nonce (set by action)
		m->state = PIGEON_SERVER_AUTH_CHECK;
		return 1;
	}
	return 0;
}

int pigeon_server_step(pigeon_server_machine *m, pairing_ceremony_event_id event)
{
	if (m->state == PIGEON_SERVER_GENERATE_TOKEN && event == PIGEON_EVENT_TOKEN_CREATED) {
		if (m->actions[PIGEON_ACTION_REGISTER_RELAY]) {
			int err = m->actions[PIGEON_ACTION_REGISTER_RELAY](m->userdata);
			if (err) return -err;
		}
		m->state = PIGEON_SERVER_REGISTER_RELAY;
		return 1;
	}
	if (m->state == PIGEON_SERVER_REGISTER_RELAY && event == PIGEON_EVENT_RELAY_REGISTERED) {
		m->state = PIGEON_SERVER_WAITING_FOR_CLIENT;
		return 1;
	}
	if (m->state == PIGEON_SERVER_DERIVE_SECRET && event == PIGEON_EVENT_ECDH_COMPLETE) {
		m->state = PIGEON_SERVER_SEND_ACK;
		return 1;
	}
	if (m->state == PIGEON_SERVER_SEND_ACK && event == PIGEON_EVENT_SIGNAL_CODE_DISPLAY) {
		m->state = PIGEON_SERVER_WAITING_FOR_CODE;
		return 1;
	}
	if (m->state == PIGEON_SERVER_VALIDATE_CODE && event == PIGEON_EVENT_CHECK_CODE && m->guards[PIGEON_GUARD_CODE_CORRECT] && m->guards[PIGEON_GUARD_CODE_CORRECT](m->userdata)) {
		m->state = PIGEON_SERVER_STORE_PAIRED;
		return 1;
	}
	if (m->state == PIGEON_SERVER_VALIDATE_CODE && event == PIGEON_EVENT_CHECK_CODE && m->guards[PIGEON_GUARD_CODE_WRONG] && m->guards[PIGEON_GUARD_CODE_WRONG](m->userdata)) {
		m->code_attempts = m->code_attempts + 1;
		if (m->on_change) m->on_change("code_attempts", m->userdata);
		m->state = PIGEON_SERVER_IDLE;
		return 1;
	}
	if (m->state == PIGEON_SERVER_STORE_PAIRED && event == PIGEON_EVENT_FINALISE) {
		if (m->actions[PIGEON_ACTION_STORE_DEVICE]) {
			int err = m->actions[PIGEON_ACTION_STORE_DEVICE](m->userdata);
			if (err) return -err;
		}
		m->device_secret = "dev_secret_1";
		if (m->on_change) m->on_change("device_secret", m->userdata);
		// paired_devices: paired_devices \union {"device_1"} (set by action)
		// active_tokens: active_tokens \ {current_token} (set by action)
		// used_tokens: used_tokens \union {current_token} (set by action)
		m->state = PIGEON_SERVER_PAIRED;
		return 1;
	}
	if (m->state == PIGEON_SERVER_AUTH_CHECK && event == PIGEON_EVENT_VERIFY && m->guards[PIGEON_GUARD_DEVICE_KNOWN] && m->guards[PIGEON_GUARD_DEVICE_KNOWN](m->userdata)) {
		if (m->actions[PIGEON_ACTION_VERIFY_DEVICE]) {
			int err = m->actions[PIGEON_ACTION_VERIFY_DEVICE](m->userdata);
			if (err) return -err;
		}
		// auth_nonces_used: auth_nonces_used \union {received_auth_nonce} (set by action)
		m->state = PIGEON_SERVER_SESSION_ACTIVE;
		return 1;
	}
	if (m->state == PIGEON_SERVER_AUTH_CHECK && event == PIGEON_EVENT_VERIFY && m->guards[PIGEON_GUARD_DEVICE_UNKNOWN] && m->guards[PIGEON_GUARD_DEVICE_UNKNOWN](m->userdata)) {
		m->state = PIGEON_SERVER_IDLE;
		return 1;
	}
	if (m->state == PIGEON_SERVER_SESSION_ACTIVE && event == PIGEON_EVENT_DISCONNECT) {
		m->state = PIGEON_SERVER_PAIRED;
		return 1;
	}
	return 0;
}

void pigeon_ios_machine_init(pigeon_ios_machine *m)
{
	memset(m, 0, sizeof(*m));
	m->state = PIGEON_APP_IDLE;
	m->received_server_pub = "none";
}

int pigeon_ios_handle_message(pigeon_ios_machine *m, pairing_ceremony_msg_type msg)
{
	if (m->state == PIGEON_APP_WAIT_ACK && msg == PIGEON_MSG_PAIR_HELLO_ACK) {
		if (m->actions[PIGEON_ACTION_DERIVE_SECRET]) {
			int err = m->actions[PIGEON_ACTION_DERIVE_SECRET](m->userdata);
			if (err) return -err;
		}
		// received_server_pub: recv_msg.pubkey (set by action)
		// client_shared_key: DeriveKey("client_pub", recv_msg.pubkey) (set by action)
		m->state = PIGEON_APP_E2E_READY;
		return 1;
	}
	if (m->state == PIGEON_APP_E2E_READY && msg == PIGEON_MSG_PAIR_CONFIRM) {
		// ios_code: DeriveCode(received_server_pub, "client_pub") (set by action)
		m->state = PIGEON_APP_SHOW_CODE;
		return 1;
	}
	if (m->state == PIGEON_APP_WAIT_PAIR_COMPLETE && msg == PIGEON_MSG_PAIR_COMPLETE) {
		if (m->actions[PIGEON_ACTION_STORE_SECRET]) {
			int err = m->actions[PIGEON_ACTION_STORE_SECRET](m->userdata);
			if (err) return -err;
		}
		m->state = PIGEON_APP_PAIRED;
		return 1;
	}
	if (m->state == PIGEON_APP_SEND_AUTH && msg == PIGEON_MSG_AUTH_OK) {
		m->state = PIGEON_APP_SESSION_ACTIVE;
		return 1;
	}
	return 0;
}

int pigeon_ios_step(pigeon_ios_machine *m, pairing_ceremony_event_id event)
{
	if (m->state == PIGEON_APP_IDLE && event == PIGEON_EVENT_USER_SCANS_QR) {
		m->state = PIGEON_APP_SCAN_QR;
		return 1;
	}
	if (m->state == PIGEON_APP_SCAN_QR && event == PIGEON_EVENT_QR_PARSED) {
		m->state = PIGEON_APP_CONNECT_RELAY;
		return 1;
	}
	if (m->state == PIGEON_APP_CONNECT_RELAY && event == PIGEON_EVENT_RELAY_CONNECTED) {
		m->state = PIGEON_APP_GEN_KEY_PAIR;
		return 1;
	}
	if (m->state == PIGEON_APP_GEN_KEY_PAIR && event == PIGEON_EVENT_KEY_PAIR_GENERATED) {
		if (m->actions[PIGEON_ACTION_SEND_PAIR_HELLO]) {
			int err = m->actions[PIGEON_ACTION_SEND_PAIR_HELLO](m->userdata);
			if (err) return -err;
		}
		m->state = PIGEON_APP_WAIT_ACK;
		return 1;
	}
	if (m->state == PIGEON_APP_SHOW_CODE && event == PIGEON_EVENT_CODE_DISPLAYED) {
		m->state = PIGEON_APP_WAIT_PAIR_COMPLETE;
		return 1;
	}
	if (m->state == PIGEON_APP_PAIRED && event == PIGEON_EVENT_APP_LAUNCH) {
		m->state = PIGEON_APP_RECONNECT;
		return 1;
	}
	if (m->state == PIGEON_APP_RECONNECT && event == PIGEON_EVENT_RELAY_CONNECTED) {
		m->state = PIGEON_APP_SEND_AUTH;
		return 1;
	}
	if (m->state == PIGEON_APP_SESSION_ACTIVE && event == PIGEON_EVENT_DISCONNECT) {
		m->state = PIGEON_APP_PAIRED;
		return 1;
	}
	return 0;
}

void pigeon_cli_machine_init(pigeon_cli_machine *m)
{
	memset(m, 0, sizeof(*m));
	m->state = PIGEON_CLI_IDLE;
}

int pigeon_cli_handle_message(pigeon_cli_machine *m, pairing_ceremony_msg_type msg)
{
	if (m->state == PIGEON_CLI_BEGIN_PAIR && msg == PIGEON_MSG_TOKEN_RESPONSE) {
		m->state = PIGEON_CLI_SHOW_QR;
		return 1;
	}
	if (m->state == PIGEON_CLI_SHOW_QR && msg == PIGEON_MSG_WAITING_FOR_CODE) {
		m->state = PIGEON_CLI_PROMPT_CODE;
		return 1;
	}
	if (m->state == PIGEON_CLI_SUBMIT_CODE && msg == PIGEON_MSG_PAIR_STATUS) {
		m->state = PIGEON_CLI_DONE;
		return 1;
	}
	return 0;
}

int pigeon_cli_step(pigeon_cli_machine *m, pairing_ceremony_event_id event)
{
	if (m->state == PIGEON_CLI_IDLE && event == PIGEON_EVENT_CLI_INIT) {
		m->state = PIGEON_CLI_GET_KEY;
		return 1;
	}
	if (m->state == PIGEON_CLI_GET_KEY && event == PIGEON_EVENT_KEY_STORED) {
		m->state = PIGEON_CLI_BEGIN_PAIR;
		return 1;
	}
	if (m->state == PIGEON_CLI_PROMPT_CODE && event == PIGEON_EVENT_USER_ENTERS_CODE) {
		m->state = PIGEON_CLI_SUBMIT_CODE;
		return 1;
	}
	return 0;
}


// --- Crypto ---



// Pigeon crypto uses:
// - X25519 ECDH for key exchange
// - HKDF-SHA256 for key derivation
// - AES-256-GCM for symmetric encryption
// - Confirmation codes: HKDF of sorted pubkeys, mod 10^6
//
// Link against libsodium (-lsodium) or OpenSSL (-lcrypto) to provide
// the underlying primitives. Define PIGEON_CRYPTO_LIBSODIUM or
// PIGEON_CRYPTO_OPENSSL to select the backend.

#if defined(PIGEON_CRYPTO_LIBSODIUM)
#include <sodium.h>

int pigeon_generate_keypair(pigeon_keypair *kp)
{
    // libsodium's crypto_box uses X25519 internally.
    // scalarmult base gives us the public key from a random secret.
    randombytes_buf(kp->private_key, 32);
    return crypto_scalarmult_base(kp->public_key, kp->private_key) == 0 ? 0 : -1;
}

// Internal: HKDF-SHA256 extract + expand. libsodium doesn't have HKDF
// natively, so we build it from HMAC-SHA256.
static int hkdf_sha256(const uint8_t *ikm, size_t ikm_len,
                       const uint8_t *info, size_t info_len,
                       uint8_t *out, size_t out_len)
{
    // Extract: PRK = HMAC-SHA256(salt="", IKM)
    uint8_t prk[32];
    uint8_t salt[32] = {0}; // empty salt
    crypto_auth_hmacsha256_state st;

    crypto_auth_hmacsha256_init(&st, salt, 32);
    crypto_auth_hmacsha256_update(&st, ikm, ikm_len);
    crypto_auth_hmacsha256_final(&st, prk);

    // Expand: OKM = HMAC-SHA256(PRK, info || 0x01)
    // For 32 bytes output, one iteration is enough.
    if (out_len > 32) return -1;

    uint8_t expand_input[256 + 1]; // info + counter byte
    if (info_len > 255) return -1;
    memcpy(expand_input, info, info_len);
    expand_input[info_len] = 0x01;

    crypto_auth_hmacsha256_init(&st, prk, 32);
    crypto_auth_hmacsha256_update(&st, expand_input, info_len + 1);
    crypto_auth_hmacsha256_final(&st, out);

    sodium_memzero(prk, sizeof(prk));
    return 0;
}

int pigeon_derive_session_key(const uint8_t *private_key,
                              const uint8_t *peer_public_key,
                              const uint8_t *info, size_t info_len,
                              uint8_t *out_key)
{
    // ECDH: shared_secret = X25519(private_key, peer_public_key)
    uint8_t shared[32];
    if (crypto_scalarmult(shared, private_key, peer_public_key) != 0) {
        return -1;
    }

    // KDF: session_key = HKDF-SHA256(shared_secret, info)
    int ret = hkdf_sha256(shared, 32, info, info_len, out_key, 32);
    sodium_memzero(shared, sizeof(shared));
    return ret;
}

int pigeon_derive_confirmation_code(const uint8_t *pub_a,
                                    const uint8_t *pub_b,
                                    char *out_code)
{
    // Sort keys lexicographically, concatenate.
    uint8_t ikm[64];
    int cmp = memcmp(pub_a, pub_b, 32);
    if (cmp <= 0) {
        memcpy(ikm, pub_a, 32);
        memcpy(ikm + 32, pub_b, 32);
    } else {
        memcpy(ikm, pub_b, 32);
        memcpy(ikm + 32, pub_a, 32);
    }

    // HKDF with info="pairing-confirmation"
    uint8_t derived[32];
    const uint8_t info[] = "pairing-confirmation";
    int ret = hkdf_sha256(ikm, 64, info, sizeof(info) - 1, derived, 32);
    if (ret != 0) return -1;

    // First 4 bytes as big-endian uint32, mod 1000000, zero-padded.
    uint32_t val = ((uint32_t)derived[0] << 24) |
                   ((uint32_t)derived[1] << 16) |
                   ((uint32_t)derived[2] << 8)  |
                   ((uint32_t)derived[3]);
    val %= 1000000;

    // Write 6-digit code + null terminator.
    for (int i = 5; i >= 0; i--) {
        out_code[i] = '0' + (char)(val % 10);
        val /= 10;
    }
    out_code[6] = '\0';
    return 0;
}

void pigeon_channel_init(pigeon_channel *ch,
                         const uint8_t *send_key,
                         const uint8_t *recv_key,
                         pigeon_channel_mode mode)
{
    memcpy(ch->send_key, send_key, 32);
    memcpy(ch->recv_key, recv_key, 32);
    ch->send_seq = 0;
    ch->recv_seq = 0;
    ch->mode = mode;
}

int pigeon_channel_init_symmetric(pigeon_channel *ch,
                                  const uint8_t *master_key,
                                  bool is_server)
{
    uint8_t c2s[32], s2c[32];
    const uint8_t info_c2s[] = "client-to-server";
    const uint8_t info_s2c[] = "server-to-client";

    if (hkdf_sha256(master_key, 32, info_c2s, sizeof(info_c2s) - 1, c2s, 32) != 0)
        return -1;
    if (hkdf_sha256(master_key, 32, info_s2c, sizeof(info_s2c) - 1, s2c, 32) != 0)
        return -1;

    if (is_server) {
        pigeon_channel_init(ch, s2c, c2s, PIGEON_MODE_STRICT);
    } else {
        pigeon_channel_init(ch, c2s, s2c, PIGEON_MODE_STRICT);
    }

    sodium_memzero(c2s, sizeof(c2s));
    sodium_memzero(s2c, sizeof(s2c));
    return 0;
}

int pigeon_channel_encrypt(pigeon_channel *ch,
                           const uint8_t *plaintext, size_t plaintext_len,
                           uint8_t *out, size_t out_len)
{
    // Output: [8-byte LE seq][ciphertext + 16-byte tag]
    size_t needed = 8 + plaintext_len + crypto_aead_aes256gcm_ABYTES;
    if (out_len < needed) return -1;

    // Sequence number as little-endian 8 bytes.
    uint64_t seq = ch->send_seq++;
    for (int i = 0; i < 8; i++) {
        out[i] = (uint8_t)(seq >> (i * 8));
    }

    // Nonce: LE64(seq) zero-padded to 12 bytes.
    uint8_t nonce[12] = {0};
    memcpy(nonce, out, 8);

    // AAD is the 8-byte seq prefix.
    unsigned long long ciphertext_len = 0;
    if (crypto_aead_aes256gcm_encrypt(out + 8, &ciphertext_len,
                                       plaintext, plaintext_len,
                                       out, 8, // AAD = seq bytes
                                       NULL, nonce, ch->send_key) != 0) {
        return -1;
    }

    return (int)(8 + ciphertext_len);
}

int pigeon_channel_decrypt(pigeon_channel *ch,
                           const uint8_t *data, size_t data_len,
                           uint8_t *out, size_t out_len)
{
    if (data_len < 8 + crypto_aead_aes256gcm_ABYTES) return -1;

    // Read sequence number (LE64).
    uint64_t seq = 0;
    for (int i = 0; i < 8; i++) {
        seq |= (uint64_t)data[i] << (i * 8);
    }

    // Sequence validation.
    if (ch->mode == PIGEON_MODE_STRICT) {
        if (seq != ch->recv_seq) return -1;
    } else {
        if (seq < ch->recv_seq) return -1;  // reject replays and old
    }

    // Nonce: LE64(seq) zero-padded to 12 bytes.
    uint8_t nonce[12] = {0};
    memcpy(nonce, data, 8);

    size_t ciphertext_len = data_len - 8;
    if (out_len < ciphertext_len - crypto_aead_aes256gcm_ABYTES) return -1;

    unsigned long long plaintext_len = 0;
    if (crypto_aead_aes256gcm_decrypt(out, &plaintext_len,
                                       NULL,
                                       data + 8, ciphertext_len,
                                       data, 8, // AAD = seq bytes
                                       nonce, ch->recv_key) != 0) {
        return -1;
    }

    if (ch->mode == PIGEON_MODE_STRICT) {
        ch->recv_seq = seq + 1;
    } else {
        if (seq >= ch->recv_seq) {
            ch->recv_seq = seq + 1;
        }
    }

    return (int)plaintext_len;
}

#else
#error "Define PIGEON_CRYPTO_LIBSODIUM or PIGEON_CRYPTO_OPENSSL to select a crypto backend"
#endif

// --- Connection and framing ---



void pigeon_init(pigeon_ctx *ctx, const pigeon_transport *transport)
{
    memset(ctx, 0, sizeof(*ctx));
    if (transport) {
        ctx->transport = *transport;
    }
    pigeon_ios_machine_init(&ctx->pairing);
}

int pigeon_send(pigeon_ctx *ctx, const uint8_t *data, size_t len)
{
    if (!ctx->transport.send_stream) return -1;
    if (len > PIGEON_MAX_MSG) return -1;

    // Frame: 4-byte BE length + payload.
    int frame_len = pigeon_frame_message(data, len,
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

    uint32_t payload_len = pigeon_read_frame_length(hdr);
    if (payload_len > PIGEON_MAX_MSG || payload_len > out_len) return -1;

    got = 0;
    err = ctx->transport.recv_stream(ctx->transport.userdata, out, payload_len, &got);
    if (err || got != payload_len) return -1;

    return (int)payload_len;
}

int pigeon_send_datagram(pigeon_ctx *ctx, const uint8_t *data, size_t len)
{
    if (!ctx->transport.send_datagram) return -1;
    return ctx->transport.send_datagram(ctx->transport.userdata, data, len);
}

int pigeon_recv_datagram(pigeon_ctx *ctx, uint8_t *out, size_t out_len)
{
    if (!ctx->transport.recv_datagram) return -1;
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

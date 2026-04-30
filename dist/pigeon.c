// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0
//
// Pigeon C client library — amalgamated source.
// Compile with -DPIGEON_CRYPTO_LIBSODIUM and link -lsodium.

#include "pigeon.h"
#include <string.h>

// --- Generated state machine ---




void pigeon_acceptor_machine_init(pigeon_acceptor_machine *m)
{
	memset(m, 0, sizeof(*m));
	m->state = PIGEON_ACCEPTOR_IDLE;
	m->acceptor_eph_pub = "none";
	m->acceptor_received_eph_pub = "none";
	m->acceptor_received_identity = "none";
	m->acceptor_received_instance = "none";
	m->acceptor_user_confirmed = "false";
	m->acceptor_received_confirm = "false";
}

int pigeon_acceptor_handle_message(pigeon_acceptor_machine *m, pairing_ceremony_msg_type msg)
{
	if (m->state == PIGEON_ACCEPTOR_WAITING_FOR_HELLO && msg == PIGEON_MSG_HELLO) {
		if (m->actions[PIGEON_ACTION_DERIVE_CODE]) {
			int err = m->actions[PIGEON_ACTION_DERIVE_CODE](m->userdata);
			if (err) return -err;
		}
		// acceptor_received_eph_pub: recv_msg.eph_pub (set by action)
		// acceptor_received_identity: recv_msg.identity_pub (set by action)
		// acceptor_received_instance: recv_msg.instance_id (set by action)
		// acceptor_code: DeriveCode(acceptor_eph_pub, recv_msg.eph_pub) (set by action)
		m->state = PIGEON_ACCEPTOR_DERIVING_CODE;
		return 1;
	}
	if (m->state == PIGEON_ACCEPTOR_AWAITING_PEER_CONFIRM && msg == PIGEON_MSG_CONFIRM_TO_ACCEPTOR) {
		if (m->actions[PIGEON_ACTION_STORE_RECORD]) {
			int err = m->actions[PIGEON_ACTION_STORE_RECORD](m->userdata);
			if (err) return -err;
		}
		m->acceptor_received_confirm = "true";
		if (m->on_change) m->on_change("acceptor_received_confirm", m->userdata);
		m->state = PIGEON_ACCEPTOR_PAIRED;
		return 1;
	}
	return 0;
}

int pigeon_acceptor_step(pigeon_acceptor_machine *m, pairing_ceremony_event_id event)
{
	if (m->state == PIGEON_ACCEPTOR_IDLE && event == PIGEON_EVENT_PAIR_BEGIN) {
		if (m->actions[PIGEON_ACTION_GEN_EPHEMERAL]) {
			int err = m->actions[PIGEON_ACTION_GEN_EPHEMERAL](m->userdata);
			if (err) return -err;
		}
		m->acceptor_eph_pub = "acceptor_eph";
		if (m->on_change) m->on_change("acceptor_eph_pub", m->userdata);
		m->state = PIGEON_ACCEPTOR_GENERATING_EPHEMERAL;
		return 1;
	}
	if (m->state == PIGEON_ACCEPTOR_GENERATING_EPHEMERAL && event == PIGEON_EVENT_EPHEMERAL_READY) {
		if (m->actions[PIGEON_ACTION_REGISTER_RELAY]) {
			int err = m->actions[PIGEON_ACTION_REGISTER_RELAY](m->userdata);
			if (err) return -err;
		}
		m->state = PIGEON_ACCEPTOR_REGISTERING_RELAY;
		return 1;
	}
	if (m->state == PIGEON_ACCEPTOR_REGISTERING_RELAY && event == PIGEON_EVENT_RELAY_REGISTERED) {
		if (m->actions[PIGEON_ACTION_EMIT_TOKEN]) {
			int err = m->actions[PIGEON_ACTION_EMIT_TOKEN](m->userdata);
			if (err) return -err;
		}
		m->state = PIGEON_ACCEPTOR_WAITING_FOR_HELLO;
		return 1;
	}
	if (m->state == PIGEON_ACCEPTOR_DERIVING_CODE && event == PIGEON_EVENT_CODE_READY) {
		m->state = PIGEON_ACCEPTOR_AWAITING_USER_CONFIRM;
		return 1;
	}
	if (m->state == PIGEON_ACCEPTOR_AWAITING_USER_CONFIRM && event == PIGEON_EVENT_USER_CONFIRM) {
		m->acceptor_user_confirmed = "true";
		if (m->on_change) m->on_change("acceptor_user_confirmed", m->userdata);
		m->state = PIGEON_ACCEPTOR_AWAITING_PEER_CONFIRM;
		return 1;
	}
	if (m->state == PIGEON_ACCEPTOR_AWAITING_USER_CONFIRM && event == PIGEON_EVENT_USER_CANCEL) {
		m->state = PIGEON_ACCEPTOR_ABORTED;
		return 1;
	}
	if (m->state == PIGEON_ACCEPTOR_AWAITING_PEER_CONFIRM && event == PIGEON_EVENT_USER_CANCEL) {
		m->state = PIGEON_ACCEPTOR_ABORTED;
		return 1;
	}
	return 0;
}

void pigeon_initiator_machine_init(pigeon_initiator_machine *m)
{
	memset(m, 0, sizeof(*m));
	m->state = PIGEON_INITIATOR_IDLE;
	m->initiator_eph_pub = "none";
	m->received_acceptor_eph_pub = "none";
	m->received_acceptor_identity = "none";
	m->received_acceptor_instance = "none";
	m->initiator_received_eph_pub = "none";
	m->initiator_received_identity = "none";
	m->initiator_received_instance = "none";
	m->initiator_user_confirmed = "false";
	m->initiator_received_confirm = "false";
}

int pigeon_initiator_handle_message(pigeon_initiator_machine *m, pairing_ceremony_msg_type msg)
{
	if (m->state == PIGEON_INITIATOR_AWAITING_WELCOME && msg == PIGEON_MSG_WELCOME) {
		if (m->actions[PIGEON_ACTION_DERIVE_CODE]) {
			int err = m->actions[PIGEON_ACTION_DERIVE_CODE](m->userdata);
			if (err) return -err;
		}
		// initiator_received_eph_pub: recv_msg.eph_pub (set by action)
		// initiator_received_identity: recv_msg.identity_pub (set by action)
		// initiator_received_instance: recv_msg.instance_id (set by action)
		// initiator_code: DeriveCode(initiator_eph_pub, recv_msg.eph_pub) (set by action)
		m->state = PIGEON_INITIATOR_DERIVING_CODE;
		return 1;
	}
	if (m->state == PIGEON_INITIATOR_AWAITING_PEER_CONFIRM && msg == PIGEON_MSG_CONFIRM_TO_INITIATOR) {
		if (m->actions[PIGEON_ACTION_STORE_RECORD]) {
			int err = m->actions[PIGEON_ACTION_STORE_RECORD](m->userdata);
			if (err) return -err;
		}
		m->initiator_received_confirm = "true";
		if (m->on_change) m->on_change("initiator_received_confirm", m->userdata);
		m->state = PIGEON_INITIATOR_PAIRED;
		return 1;
	}
	return 0;
}

int pigeon_initiator_step(pigeon_initiator_machine *m, pairing_ceremony_event_id event)
{
	if (m->state == PIGEON_INITIATOR_IDLE && event == PIGEON_EVENT_TOKEN_RECEIVED) {
		if (m->actions[PIGEON_ACTION_DECODE_TOKEN]) {
			int err = m->actions[PIGEON_ACTION_DECODE_TOKEN](m->userdata);
			if (err) return -err;
		}
		m->received_acceptor_eph_pub = "acceptor_eph";
		if (m->on_change) m->on_change("received_acceptor_eph_pub", m->userdata);
		m->received_acceptor_identity = "acceptor_id";
		if (m->on_change) m->on_change("received_acceptor_identity", m->userdata);
		m->received_acceptor_instance = "acceptor_instance";
		if (m->on_change) m->on_change("received_acceptor_instance", m->userdata);
		m->state = PIGEON_INITIATOR_DECODING_TOKEN;
		return 1;
	}
	if (m->state == PIGEON_INITIATOR_DECODING_TOKEN && event == PIGEON_EVENT_TOKEN_DECODED) {
		if (m->actions[PIGEON_ACTION_GEN_EPHEMERAL]) {
			int err = m->actions[PIGEON_ACTION_GEN_EPHEMERAL](m->userdata);
			if (err) return -err;
		}
		m->initiator_eph_pub = "initiator_eph";
		if (m->on_change) m->on_change("initiator_eph_pub", m->userdata);
		m->state = PIGEON_INITIATOR_GENERATING_EPHEMERAL;
		return 1;
	}
	if (m->state == PIGEON_INITIATOR_GENERATING_EPHEMERAL && event == PIGEON_EVENT_EPHEMERAL_READY) {
		if (m->actions[PIGEON_ACTION_DIAL_RELAY]) {
			int err = m->actions[PIGEON_ACTION_DIAL_RELAY](m->userdata);
			if (err) return -err;
		}
		m->state = PIGEON_INITIATOR_CONNECTING_RELAY;
		return 1;
	}
	if (m->state == PIGEON_INITIATOR_CONNECTING_RELAY && event == PIGEON_EVENT_RELAY_CONNECTED) {
		m->state = PIGEON_INITIATOR_AWAITING_WELCOME;
		return 1;
	}
	if (m->state == PIGEON_INITIATOR_DERIVING_CODE && event == PIGEON_EVENT_CODE_READY) {
		m->state = PIGEON_INITIATOR_AWAITING_USER_CONFIRM;
		return 1;
	}
	if (m->state == PIGEON_INITIATOR_AWAITING_USER_CONFIRM && event == PIGEON_EVENT_USER_CONFIRM) {
		m->initiator_user_confirmed = "true";
		if (m->on_change) m->on_change("initiator_user_confirmed", m->userdata);
		m->state = PIGEON_INITIATOR_AWAITING_PEER_CONFIRM;
		return 1;
	}
	if (m->state == PIGEON_INITIATOR_AWAITING_USER_CONFIRM && event == PIGEON_EVENT_USER_CANCEL) {
		m->state = PIGEON_INITIATOR_ABORTED;
		return 1;
	}
	if (m->state == PIGEON_INITIATOR_AWAITING_PEER_CONFIRM && event == PIGEON_EVENT_USER_CANCEL) {
		m->state = PIGEON_INITIATOR_ABORTED;
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
    ch->established = true;
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

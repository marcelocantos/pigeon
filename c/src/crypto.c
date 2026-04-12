// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

#include "pigeon/pigeon.h"
#include <string.h>

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

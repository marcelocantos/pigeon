// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package com.marcelocantos.pigeon.crypto

import java.security.KeyFactory
import java.security.KeyPairGenerator
import java.security.PrivateKey
import java.security.PublicKey
import java.security.spec.PKCS8EncodedKeySpec
import java.security.spec.X509EncodedKeySpec
import javax.crypto.Cipher
import javax.crypto.KeyAgreement
import javax.crypto.spec.GCMParameterSpec
import javax.crypto.spec.SecretKeySpec

/**
 * End-to-end encryption for traffic relayed through pigeon.
 * Mirrors the Go crypto package and Swift Pigeon.
 *
 * Key exchange: X25519 ECDH
 * Symmetric encryption: AES-256-GCM with counter nonce
 * Key derivation: HKDF-SHA256
 *
 * Requires Java 11+ / Android API 33+.
 */

// X.509 SubjectPublicKeyInfo header for X25519 (12 bytes, fixed).
private val X25519_X509_HEADER = byteArrayOf(
    0x30, 0x2A, 0x30, 0x05, 0x06, 0x03, 0x2B, 0x65, 0x6E, 0x03, 0x21, 0x00
)

// PKCS#8 PrivateKeyInfo header for X25519 (16 bytes, fixed).
// ASN.1: SEQUENCE { INTEGER 0, SEQUENCE { OID 1.3.101.110 }, OCTET STRING { OCTET STRING { key } } }
private val X25519_PKCS8_HEADER = byteArrayOf(
    0x30, 0x2E, 0x02, 0x01, 0x00, 0x30, 0x05, 0x06,
    0x03, 0x2B, 0x65, 0x6E, 0x04, 0x22, 0x04, 0x20
)

// ---- Key exchange ----

/** An X25519 ECDH key pair for key exchange. */
class E2EKeyPair {
    private val keyPair = KeyPairGenerator.getInstance("X25519").generateKeyPair()

    /** Raw public key bytes (32 bytes) for sending to peer. */
    val publicKeyData: ByteArray
        get() {
            val encoded = keyPair.public.encoded
            return encoded.copyOfRange(encoded.size - 32, encoded.size)
        }

    /** Raw private key bytes (32 bytes) for persistent storage. */
    val privateKeyData: ByteArray
        get() {
            val encoded = keyPair.private.encoded
            return encoded.copyOfRange(encoded.size - 32, encoded.size)
        }

    /**
     * Derive a shared secret via ECDH, then derive a 256-bit key via HKDF.
     *
     * @param peerPublicKey raw 32-byte X25519 public key from peer
     * @param info HKDF info parameter (e.g. "client-to-server")
     * @return 32-byte derived session key
     */
    fun deriveSessionKey(peerPublicKey: ByteArray, info: ByteArray): ByteArray {
        require(peerPublicKey.size == 32) { "peer public key must be 32 bytes" }
        val peerKey = rawToX25519PublicKey(peerPublicKey)
        val ka = KeyAgreement.getInstance("X25519")
        ka.init(keyPair.private)
        ka.doPhase(peerKey, true)
        val shared = ka.generateSecret()
        return Hkdf.derive(shared, info)
    }
}

/**
 * Derive a session key from a persistent secret and info via HKDF-SHA256.
 */
fun deriveKeyFromSecret(secret: ByteArray, info: ByteArray): ByteArray =
    Hkdf.derive(secret, info)

/** Generate 32 cryptographically random bytes suitable for use as a nonce. */
fun generateNonce(): ByteArray {
    val b = ByteArray(32)
    java.security.SecureRandom().nextBytes(b)
    return b
}

/** Generate 32 cryptographically random bytes suitable for use as a secret. */
fun generateSecret(): ByteArray = generateNonce()

// ---- Confirmation code ----

/**
 * Derive a 6-digit confirmation code from two X25519 public keys.
 * The code is order-independent and deterministic: swapping the keys
 * produces the same result. Both sides of a key exchange compute this
 * independently; a mismatch indicates a MitM attack.
 */
fun deriveConfirmationCode(pubA: ByteArray, pubB: ByteArray): String {
    // Sort lexicographically for order-independence.
    val cmp = pubA.zip(pubB).map { (x, y) -> (x.toInt() and 0xFF) - (y.toInt() and 0xFF) }
        .firstOrNull { it != 0 } ?: 0
    val (a, b) = if (cmp <= 0) pubA to pubB else pubB to pubA
    val ikm = a + b
    val derived = Hkdf.derive(ikm, "pairing-confirmation".toByteArray(), length = 4)
    val value = ((derived[0].toLong() and 0xFF) shl 24) or
                ((derived[1].toLong() and 0xFF) shl 16) or
                ((derived[2].toLong() and 0xFF) shl 8) or
                (derived[3].toLong() and 0xFF)
    val code = value % 1_000_000
    return String.format("%06d", code)
}

// ---- Channel mode ----

/**
 * Controls how [E2EChannel.decrypt] handles sequence numbers.
 */
enum class ChannelMode {
    /**
     * Strict (default): sequence numbers must be contiguous with no gaps.
     * Any out-of-order or replayed packet is rejected. Suitable for reliable
     * transports (reliable streams).
     */
    STRICT,

    /**
     * Datagrams: gaps in the sequence number space are accepted, as expected
     * on lossy transports (UDP, H.264 video). Packets with a sequence number
     * less than the last accepted one are rejected to prevent replay attacks.
     */
    DATAGRAMS,
}

// ---- Encrypted channel ----

/**
 * Provides symmetric encryption/decryption for a relay connection.
 * Uses AES-256-GCM with a monotonic counter nonce. Thread-safe.
 */
class E2EChannel private constructor(
    private val sendKey: ByteArray,
    private val recvKey: ByteArray,
) {
    // Note: Kotlin Long is signed 64-bit. Overflow at Long.MAX_VALUE wraps
    // to Long.MIN_VALUE, which differs from Go/Swift unsigned uint64.
    // This only matters after 2^63 messages (~9.2 quintillion) — theoretical.
    private var sendSeq: Long = 0
    private var recvSeq: Long = 0
    private val lock = Any()

    /**
     * The sequence-number checking mode used by [decrypt].
     * Defaults to [ChannelMode.STRICT]. Set before the first [decrypt] call.
     */
    @Volatile
    var mode: ChannelMode = ChannelMode.STRICT

    /** Create a channel with separate send/recv keys. */
    constructor(sendKey: ByteArray, recvKey: ByteArray, @Suppress("UNUSED_PARAMETER") marker: Unit = Unit)
        : this(sendKey.copyOf(), recvKey.copyOf())

    /** Create a symmetric channel from a shared key, deriving directional keys via HKDF. */
    constructor(sharedKey: ByteArray, isServer: Boolean) : this(
        sendKey = deriveKeyFromSecret(
            sharedKey,
            if (isServer) "server-to-client".toByteArray() else "client-to-server".toByteArray()
        ),
        recvKey = deriveKeyFromSecret(
            sharedKey,
            if (isServer) "client-to-server".toByteArray() else "server-to-client".toByteArray()
        )
    )

    /**
     * Encrypt a plaintext message. Returns `[8-byte seq LE][ciphertext][16-byte tag]`.
     */
    fun encrypt(plaintext: ByteArray): ByteArray = synchronized(lock) {
        val seq = sendSeq++
        val seqBytes = longToLE(seq)
        val nonce = makeNonce(seq)

        val cipher = Cipher.getInstance("AES/GCM/NoPadding")
        cipher.init(Cipher.ENCRYPT_MODE, SecretKeySpec(sendKey, "AES"), GCMParameterSpec(128, nonce))
        cipher.updateAAD(seqBytes)
        val ciphertext = cipher.doFinal(plaintext)

        seqBytes + ciphertext
    }

    /**
     * Decrypt a ciphertext message. Verifies the sequence number according to [mode].
     *
     * @throws E2EException if the ciphertext is too short or the sequence check fails
     */
    fun decrypt(data: ByteArray): ByteArray {
        if (data.size < 8 + 16) throw E2EException("Ciphertext too short")

        val seqBytes = data.copyOfRange(0, 8)
        val seq = leToLong(seqBytes)
        val payload = data.copyOfRange(8, data.size)

        return synchronized(lock) {
            when (mode) {
                ChannelMode.STRICT -> {
                    if (seq != recvSeq) throw E2EException("Unexpected sequence number")
                    recvSeq++
                }
                ChannelMode.DATAGRAMS -> {
                    // recvSeq holds highest-accepted-seq + 1 (0 if none yet).
                    // Accept seq >= recvSeq (gaps allowed); reject seq < recvSeq (replay).
                    if (java.lang.Long.compareUnsigned(seq, recvSeq) < 0)
                        throw E2EException("Sequence number replayed or too old")
                    recvSeq = seq + 1
                }
            }

            val nonce = makeNonce(seq)
            val cipher = Cipher.getInstance("AES/GCM/NoPadding")
            cipher.init(Cipher.DECRYPT_MODE, SecretKeySpec(recvKey, "AES"), GCMParameterSpec(128, nonce))
            cipher.updateAAD(seqBytes)
            cipher.doFinal(payload)
        }
    }

    private fun makeNonce(seq: Long): ByteArray {
        val nonce = ByteArray(12)
        for (i in 0..7) nonce[i] = (seq ushr (i * 8)).toByte()
        return nonce
    }

    private fun longToLE(v: Long): ByteArray {
        val b = ByteArray(8)
        for (i in 0..7) b[i] = (v ushr (i * 8)).toByte()
        return b
    }

    private fun leToLong(b: ByteArray): Long {
        var v = 0L
        for (i in 0..7) v = v or ((b[i].toLong() and 0xFF) shl (i * 8))
        return v
    }
}

class E2EException(message: String) : Exception(message)

// ---- Session salt (🎯T50) ----

/**
 * Length of the per-session diversifier folded into [PairingRecord.deriveChannel]'s
 * HKDF info. Matches the activation-path nonce length (T44.1) so reconnects
 * under one PairingRecord never share an AEAD key.
 */
const val SESSION_SALT_LEN = 16

/**
 * Fixed peer reply to a session-salt handshake on the artifact/reconnect
 * path (plaintext on the primary stream before AEAD is established).
 */
const val SESSION_SALT_ACK = "session_salt_ack"

/** Generate [SESSION_SALT_LEN] cryptographically random bytes (🎯T50). */
fun generateSessionSalt(): ByteArray {
    val b = ByteArray(SESSION_SALT_LEN)
    java.security.SecureRandom().nextBytes(b)
    return b
}

// ---- Pairing record ----

/**
 * Persistent state from a completed pairing ceremony.
 * Serialize to JSON and store in EncryptedSharedPreferences or similar.
 * On reconnect, load it and call [deriveChannel] to derive session keys
 * without repeating the ECDH ceremony.
 */
data class PairingRecord(
    val peerInstanceID: String,
    val relayURL: String,
    val localPrivateKey: ByteArray,  // raw X25519, 32 bytes
    val localPublicKey: ByteArray,   // raw X25519, 32 bytes
    val peerPublicKey: ByteArray,    // raw X25519, 32 bytes
) {
    constructor(
        peerInstanceID: String,
        relayURL: String,
        localKeyPair: E2EKeyPair,
        peerPublicKey: ByteArray,
    ) : this(
        peerInstanceID = peerInstanceID,
        relayURL = relayURL,
        localPrivateKey = localKeyPair.privateKeyData,
        localPublicKey = localKeyPair.publicKeyData,
        peerPublicKey = peerPublicKey,
    )

    /**
     * Derive an encrypted channel from the stored keys.
     *
     * [sessionSalt] is a per-session diversifier (typically
     * [SESSION_SALT_LEN] random bytes exchanged during reconnect). When
     * non-empty it is appended to each direction label (HKDF info =
     * label || salt), so two sessions under one PairingRecord derive
     * distinct AEAD keys (🎯T50). Pass empty only for tests that
     * intentionally exercise the static pre-T50 derivation. When
     * non-empty, [sessionSalt] must be exactly [SESSION_SALT_LEN] bytes.
     */
    fun deriveChannel(
        sendInfo: ByteArray,
        recvInfo: ByteArray,
        sessionSalt: ByteArray = ByteArray(0),
    ): E2EChannel {
        require(sessionSalt.isEmpty() || sessionSalt.size == SESSION_SALT_LEN) {
            "session salt must be $SESSION_SALT_LEN bytes, got ${sessionSalt.size}"
        }
        val sendKeyInfo = sendInfo + sessionSalt
        val recvKeyInfo = recvInfo + sessionSalt
        val sendKey = deriveSessionKeyFromRaw(localPrivateKey, peerPublicKey, sendKeyInfo)
        val recvKey = deriveSessionKeyFromRaw(localPrivateKey, peerPublicKey, recvKeyInfo)
        return E2EChannel(sendKey, recvKey)
    }

    override fun equals(other: Any?): Boolean {
        if (this === other) return true
        if (other !is PairingRecord) return false
        return peerInstanceID == other.peerInstanceID &&
            relayURL == other.relayURL &&
            localPrivateKey.contentEquals(other.localPrivateKey) &&
            localPublicKey.contentEquals(other.localPublicKey) &&
            peerPublicKey.contentEquals(other.peerPublicKey)
    }

    override fun hashCode(): Int {
        var result = peerInstanceID.hashCode()
        result = 31 * result + relayURL.hashCode()
        result = 31 * result + localPrivateKey.contentHashCode()
        result = 31 * result + localPublicKey.contentHashCode()
        result = 31 * result + peerPublicKey.contentHashCode()
        return result
    }
}

/**
 * Derive a session key from raw private key bytes, raw peer public key bytes,
 * and an info parameter via ECDH + HKDF-SHA256.
 */
internal fun deriveSessionKeyFromRaw(
    privateKeyBytes: ByteArray,
    peerPublicKeyBytes: ByteArray,
    info: ByteArray,
): ByteArray {
    val priv = rawToX25519PrivateKey(privateKeyBytes)
    val pub = rawToX25519PublicKey(peerPublicKeyBytes)
    val ka = KeyAgreement.getInstance("X25519")
    ka.init(priv)
    ka.doPhase(pub, true)
    val shared = ka.generateSecret()
    return Hkdf.derive(shared, info)
}

// ---- Internal helpers ----

private fun rawToX25519PublicKey(raw: ByteArray): PublicKey {
    val encoded = X25519_X509_HEADER + raw
    return KeyFactory.getInstance("X25519").generatePublic(X509EncodedKeySpec(encoded))
}

private fun rawToX25519PrivateKey(raw: ByteArray): PrivateKey {
    val encoded = X25519_PKCS8_HEADER + raw
    return KeyFactory.getInstance("X25519").generatePrivate(PKCS8EncodedKeySpec(encoded))
}

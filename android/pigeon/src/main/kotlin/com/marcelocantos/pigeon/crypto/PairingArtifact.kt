// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package com.marcelocantos.pigeon.crypto

import com.marcelocantos.pigeon.relay.PigeonConn
import com.marcelocantos.pigeon.relay.QuicTransport
import com.marcelocantos.pigeon.relay.connect
import java.time.Instant
import java.util.Base64

/** The default lifetime of a [PairingArtifact] when none is specified. */
val DEFAULT_PAIRING_TTL: java.time.Duration = java.time.Duration.ofDays(30)

/** Thrown when a [PairingArtifact] is used past its `expiresAt` timestamp. */
class PairingExpiredException(val expiresAt: Instant) :
    Exception("pigeon: pairing artifact expired at $expiresAt")

/**
 * The persistable+expirable envelope around a completed pairing.
 *
 * The cryptographic core lives in [record]; the wrapping fields carry
 * the lifecycle metadata that lets a client detect expiry and prompt
 * re-pair.
 *
 * Wire format matches the Go and Swift SDKs: snake_case JSON keys,
 * ISO-8601 timestamps, base64 byte fields. An artifact minted in any
 * SDK can be decoded by the other two.
 */
data class PairingArtifact(
    val record: PairingRecord,
    val token: String = "",
    val issuedAt: Instant = Instant.now(),
    /** `null` means the artifact never expires. */
    val expiresAt: Instant? = null,
) {
    /** Reports whether the artifact is past its expiry at [now]. */
    fun isExpired(now: Instant = Instant.now()): Boolean {
        val exp = expiresAt ?: return false
        return !now.isBefore(exp)
    }

    /** Serialises to canonical JSON. Wire-compatible with the Go and Swift SDKs. */
    fun toJson(): String = pairingArtifactToJson(this)

    /**
     * Returns the canonical single-line text encoding: base64url of
     * the JSON bytes, with no padding. Suitable for transport via QR
     * payload, launch argument, environment variable, pasteboard, or
     * any other channel that wants a single token of text.
     */
    fun toText(): String = base64UrlNoPad(toJson().toByteArray(Charsets.UTF_8))

    companion object {
        /** Mints an artifact wrapping [record] with a TTL (default 30 days). */
        fun mint(
            record: PairingRecord,
            token: String = "",
            issuedAt: Instant = Instant.now(),
            ttl: java.time.Duration = DEFAULT_PAIRING_TTL,
        ): PairingArtifact {
            val expiresAt = if (!ttl.isNegative && !ttl.isZero) issuedAt.plus(ttl) else null
            return PairingArtifact(record, token, issuedAt, expiresAt)
        }

        /** Deserialises from canonical JSON. */
        fun fromJson(json: String): PairingArtifact = pairingArtifactFromJson(json)

        /** Decodes the canonical text encoding produced by [toText]. */
        fun fromText(text: String): PairingArtifact =
            fromJson(String(base64UrlDecode(text), Charsets.UTF_8))
    }
}

/**
 * Connect to the relay using a persisted [PairingArtifact] and wire
 * encrypted I/O on the resulting [PigeonConn].
 *
 * The peer instance ID comes from the artifact; the artifact's expiry
 * is checked up front and throws [PairingExpiredException] (so callers
 * can route to a re-pair flow uniformly).
 *
 * The caller still supplies the [QuicTransport] because the JVM-side
 * library is transport-agnostic — pick a Kwik or Bridge transport
 * connected to the relay's host/port from the artifact.
 *
 * After the relay bridge is live, runs the 🎯T50 session-salt
 * handshake on the primary stream: mints a fresh [SESSION_SALT_LEN]-byte
 * salt, sends it plaintext, waits for [SESSION_SALT_ACK], then derives
 * the AEAD channel with HKDF info = direction-label || salt so each
 * reconnect gets distinct keys.
 *
 * The peer must call [acceptSessionSalt] on its bridged conn with the
 * mirrored PairingRecord.
 *
 * Mirrors `PigeonConn.connect(artifact:)` (Swift).
 */
fun connectWithArtifact(transport: QuicTransport, artifact: PairingArtifact): PigeonConn {
    if (artifact.isExpired()) {
        throw PairingExpiredException(artifact.expiresAt ?: Instant.now())
    }
    val conn = connect(transport, artifact.record.peerInstanceID)
    val channel = initiateSessionSalt(
        conn,
        artifact.record,
        "client-to-server".toByteArray(),
        "server-to-client".toByteArray(),
    )
    conn.setChannel(channel)
    return conn
}

/**
 * Client side of the 🎯T50 session-salt handshake: mint salt, send it,
 * wait for [SESSION_SALT_ACK], derive channel with info = label || salt.
 *
 * Called before [PigeonConn.setChannel] so the exchange is plaintext.
 */
fun initiateSessionSalt(
    conn: PigeonConn,
    record: PairingRecord,
    sendInfo: ByteArray,
    recvInfo: ByteArray,
): E2EChannel {
    val salt = generateSessionSalt()
    conn.send(salt)
    val ack = conn.recv()
    val ackStr = String(ack, Charsets.UTF_8)
    require(ackStr == SESSION_SALT_ACK) {
        "session salt handshake: expected $SESSION_SALT_ACK, got $ackStr"
    }
    return record.deriveChannel(sendInfo, recvInfo, salt)
}

/**
 * Peer/backend side of the 🎯T50 session-salt handshake: read the
 * client's salt, reply [SESSION_SALT_ACK], derive channel with the same
 * salt. Direction labels are the caller's responsibility (typically the
 * reverse of the client's).
 *
 * Called before [PigeonConn.setChannel] so the exchange is plaintext.
 */
fun acceptSessionSalt(
    conn: PigeonConn,
    record: PairingRecord,
    sendInfo: ByteArray,
    recvInfo: ByteArray,
): E2EChannel {
    val salt = conn.recv()
    require(salt.size == SESSION_SALT_LEN) {
        "session salt handshake: expected $SESSION_SALT_LEN-byte salt, got ${salt.size}"
    }
    conn.send(SESSION_SALT_ACK.toByteArray(Charsets.UTF_8))
    return record.deriveChannel(sendInfo, recvInfo, salt)
}

// ---- Hand-rolled JSON ----
//
// kotlinx.serialization would be idiomatic but pulls a plugin and a
// dependency. The artifact has a small, fixed shape with no embedded
// user-controlled strings beyond `peerInstanceID` and `relayURL`,
// which the encoder escapes defensively.

private fun pairingArtifactToJson(a: PairingArtifact): String {
    val b64 = Base64.getEncoder()
    val sb = StringBuilder()
    sb.append('{')
    sb.append("\"record\":{")
    sb.append("\"peer_instance_id\":").append(quote(a.record.peerInstanceID)).append(',')
    sb.append("\"relay_url\":").append(quote(a.record.relayURL)).append(',')
    sb.append("\"local_private_key\":").append(quote(b64.encodeToString(a.record.localPrivateKey))).append(',')
    sb.append("\"local_public_key\":").append(quote(b64.encodeToString(a.record.localPublicKey))).append(',')
    sb.append("\"peer_public_key\":").append(quote(b64.encodeToString(a.record.peerPublicKey)))
    sb.append("},")
    if (a.token.isNotEmpty()) {
        sb.append("\"token\":").append(quote(a.token)).append(',')
    }
    sb.append("\"issued_at\":").append(quote(a.issuedAt.toString())).append(',')
    // Encode zero (epoch) when no expiry, mirroring Go's time.Time zero value.
    val exp = a.expiresAt ?: Instant.EPOCH
    sb.append("\"expires_at\":").append(quote(exp.toString()))
    sb.append('}')
    return sb.toString()
}

private fun pairingArtifactFromJson(json: String): PairingArtifact {
    val obj = JsonParser(json).parseObject()

    @Suppress("UNCHECKED_CAST")
    val rec = obj["record"] as? Map<String, Any?>
        ?: throw IllegalArgumentException("missing record")
    val record = PairingRecord(
        peerInstanceID = rec.getString("peer_instance_id"),
        relayURL = rec.getString("relay_url"),
        localPrivateKey = Base64.getDecoder().decode(rec.getString("local_private_key")),
        localPublicKey = Base64.getDecoder().decode(rec.getString("local_public_key")),
        peerPublicKey = Base64.getDecoder().decode(rec.getString("peer_public_key")),
    )
    val token = (obj["token"] as? String) ?: ""
    val issuedAt = Instant.parse(obj.getString("issued_at"))
    val expiresAtStr = obj.getString("expires_at")
    val expiresAtParsed = Instant.parse(expiresAtStr)
    val expiresAt = if (expiresAtParsed == Instant.EPOCH) null else expiresAtParsed

    return PairingArtifact(record, token, issuedAt, expiresAt)
}

private fun Map<String, Any?>.getString(key: String): String =
    this[key] as? String ?: throw IllegalArgumentException("missing or non-string field: $key")

private fun quote(s: String): String {
    val sb = StringBuilder(s.length + 2)
    sb.append('"')
    for (c in s) {
        when (c) {
            '\\' -> sb.append("\\\\")
            '"' -> sb.append("\\\"")
            '\n' -> sb.append("\\n")
            '\r' -> sb.append("\\r")
            '\t' -> sb.append("\\t")
            '\b' -> sb.append("\\b")
            '\u000C' -> sb.append("\\f")
            else -> if (c.code < 0x20) {
                sb.append("\\u%04x".format(c.code))
            } else {
                sb.append(c)
            }
        }
    }
    sb.append('"')
    return sb.toString()
}

private fun base64UrlNoPad(bytes: ByteArray): String =
    Base64.getUrlEncoder().withoutPadding().encodeToString(bytes)

private fun base64UrlDecode(s: String): ByteArray =
    Base64.getUrlDecoder().decode(s)

// Minimal recursive-descent JSON parser. Handles the subset emitted by
// the Go/Swift SDKs: objects, strings, with no need for arrays or
// numbers.
private class JsonParser(private val src: String) {
    private var i = 0

    fun parseObject(): Map<String, Any?> {
        skipWs()
        expect('{')
        val map = mutableMapOf<String, Any?>()
        skipWs()
        if (peek() == '}') { i++; return map }
        while (true) {
            skipWs()
            val key = parseString()
            skipWs()
            expect(':')
            skipWs()
            val value = parseValue()
            map[key] = value
            skipWs()
            when (peek()) {
                ',' -> { i++; continue }
                '}' -> { i++; return map }
                else -> error("expected , or } at $i")
            }
        }
    }

    private fun parseValue(): Any? {
        skipWs()
        return when (peek()) {
            '"' -> parseString()
            '{' -> parseObject()
            't', 'f' -> parseBool()
            'n' -> parseNull()
            else -> error("unexpected char at $i: ${peek()}")
        }
    }

    private fun parseString(): String {
        expect('"')
        val sb = StringBuilder()
        while (true) {
            if (i >= src.length) error("unterminated string")
            val c = src[i++]
            if (c == '"') return sb.toString()
            if (c == '\\') {
                if (i >= src.length) error("bad escape")
                when (val esc = src[i++]) {
                    '"' -> sb.append('"')
                    '\\' -> sb.append('\\')
                    '/' -> sb.append('/')
                    'n' -> sb.append('\n')
                    'r' -> sb.append('\r')
                    't' -> sb.append('\t')
                    'b' -> sb.append('\b')
                    'f' -> sb.append('\u000C')
                    'u' -> {
                        if (i + 4 > src.length) error("bad unicode escape")
                        sb.append(src.substring(i, i + 4).toInt(16).toChar())
                        i += 4
                    }
                    else -> error("bad escape: $esc")
                }
            } else {
                sb.append(c)
            }
        }
    }

    private fun parseBool(): Boolean {
        return when {
            src.startsWith("true", i) -> { i += 4; true }
            src.startsWith("false", i) -> { i += 5; false }
            else -> error("invalid bool at $i")
        }
    }

    private fun parseNull(): Any? {
        if (src.startsWith("null", i)) { i += 4; return null }
        error("invalid null at $i")
    }

    private fun expect(c: Char) {
        if (i >= src.length || src[i] != c) error("expected '$c' at $i")
        i++
    }

    private fun peek(): Char = if (i < src.length) src[i] else '\u0000'

    private fun skipWs() {
        while (i < src.length && src[i].isWhitespace()) i++
    }

    private fun error(msg: String): Nothing = throw IllegalArgumentException("JSON: $msg")
}

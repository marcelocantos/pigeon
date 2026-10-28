// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package com.marcelocantos.pigeon.peer

import com.marcelocantos.pigeon.crypto.E2EKeyPair
import com.marcelocantos.pigeon.crypto.JsonParser
import com.marcelocantos.pigeon.crypto.PairingRecord
import com.marcelocantos.pigeon.crypto.deriveConfirmationCode
import com.marcelocantos.pigeon.crypto.generateNonce
import java.security.MessageDigest
import java.util.Base64

/** The Go pairing.Register rendezvous token, not an activated credential. */
internal data class Rendezvous(
    val relay: String, val sessionID: String, val acceptorInstance: String,
    val acceptorEphPub: ByteArray, val acceptorIdPub: ByteArray,
) {
    companion object {
        fun decode(text: String): Rendezvous {
            val json = try { String(Base64.getUrlDecoder().decode(text), Charsets.UTF_8) }
                catch (e: IllegalArgumentException) { throw IllegalArgumentException("invalid rendezvous token", e) }
            val obj = JsonParser(json).parseObject()
            fun str(key: String): String = obj[key] as? String ?: throw IllegalArgumentException("missing $key")
            fun key(key: String): ByteArray {
                val b = Base64.getDecoder().decode(str(key))
                require(b.size == 32) { "$key must be a 32-byte public key" }
                return b
            }
            return Rendezvous(str("relay"), str("session_id"), str("acceptor_instance"),
                key("acceptor_eph_pub"), key("acceptor_id_pub")).also {
                require(it.sessionID.isNotEmpty() && it.acceptorInstance.isNotEmpty()) { "empty rendezvous identity" }
            }
        }
    }
}

/** C ngtcp2 transport writes the 4-byte BE frame; Go singleStreamTransport writes
 * the same frame over its RAW RWC. The caller passes JSON only here. */
private suspend fun RelayClientStream.sendPairingMessage(message: String) {
    val bytes = message.toByteArray(Charsets.UTF_8)
    require(bytes.size <= 1_048_576) { "pairing message too large" }
    send(bytes)
}

private suspend fun RelayClientStream.recvPairingMessage(): Map<String, Any?> {
    val message = recv()
    require(message.isNotEmpty() && message.size <= 1_048_576) { "invalid pairing message length" }
    return JsonParser(String(message, Charsets.UTF_8)).parseObject()
}

private fun jsonString(s: String): String = buildString {
    append('"')
    for (c in s) when (c) {
        '"' -> append("\\\"")
        '\\' -> append("\\\\")
        '\n' -> append("\\n")
        '\r' -> append("\\r")
        else -> { require(c.code >= 32) { "invalid identity control character" }; append(c) }
    }
    append('"')
}

/**
 * Drive the modern commit-then-reveal pairing ceremony with a Go backend.
 * The callback must compare the SAS with the other device's displayed code;
 * returning false aborts without emitting a confirmation or record.
 * The library owns and closes the plaintext pairing session and stream.
 * The resulting record is used on a FRESH connection via [connectPaired].
 */
suspend fun pairInitiator(
    rendezvousToken: String,
    identityInstanceID: String,
    identityPublicKey: ByteArray,
    confirm: suspend (String) -> Boolean,
): PairingRecord = pairInitiatorWithConnector(
    rendezvousToken, identityInstanceID, identityPublicKey, confirm,
    { url, id -> connect(url, id) },
)

internal suspend fun pairInitiatorWithConnector(
    rendezvousToken: String,
    identityInstanceID: String,
    identityPublicKey: ByteArray,
    confirm: suspend (String) -> Boolean,
    connector: suspend (String, String) -> RelayClientSession,
): PairingRecord {
    require(identityPublicKey.size == 32) { "identity key must be 32 bytes" }
    require(identityInstanceID.isNotBlank()) { "identity instance ID required" }
    val token = Rendezvous.decode(rendezvousToken)
    val eph = E2EKeyPair()
    val blind = generateNonce()
    val digest = MessageDigest.getInstance("SHA-256")
    val commit = digest.digest("pigeon-sas-commit".toByteArray() + eph.publicKeyData + blind)
    val b64 = Base64.getEncoder()
    val session = connector(token.relay, token.sessionID)
    try {
        val stream = session.openStream("ceremony")
        try {
            stream.sendPairingMessage("""{"kind":"hello","commit":"${b64.encodeToString(commit)}","identity_pub":"${b64.encodeToString(identityPublicKey)}","instance_id":${jsonString(identityInstanceID)}}""")
            val welcome = stream.recvPairingMessage()
            require(welcome["kind"] == "welcome") { "expected pairing welcome" }
            val peerEph = Base64.getDecoder().decode(welcome["eph_pub"] as? String
                ?: throw IllegalArgumentException("welcome missing eph_pub"))
            require(MessageDigest.isEqual(peerEph, token.acceptorEphPub)) { "welcome key differs from rendezvous token" }
            require(welcome["instance_id"] == token.acceptorInstance) { "welcome instance differs from rendezvous token" }
            val peerIdentity = Base64.getDecoder().decode(welcome["identity_pub"] as? String
                ?: throw IllegalArgumentException("welcome missing identity_pub"))
            require(MessageDigest.isEqual(peerIdentity, token.acceptorIdPub)) { "welcome identity differs from rendezvous token" }
            stream.sendPairingMessage("""{"kind":"reveal","eph_pub":"${b64.encodeToString(eph.publicKeyData)}","blind":"${b64.encodeToString(blind)}"}""")
            val code = deriveConfirmationCode(eph.publicKeyData, token.acceptorEphPub)
            if (!confirm(code)) throw IllegalStateException("pairing rejected by user")
            stream.sendPairingMessage("""{"kind":"confirm"}""")
            require(stream.recvPairingMessage()["kind"] == "confirm") { "peer did not confirm pairing" }
            return PairingRecord(token.acceptorInstance, token.relay,
                eph.privateKeyData, eph.publicKeyData, token.acceptorEphPub)
        } finally { stream.close() }
    } finally { session.close() }
}

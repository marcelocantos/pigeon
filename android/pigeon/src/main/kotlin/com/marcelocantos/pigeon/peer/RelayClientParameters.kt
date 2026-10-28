// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package com.marcelocantos.pigeon.peer

import com.marcelocantos.pigeon.crypto.PairingRecord
import java.net.URI

/** Validated ngtcp2 dial arguments. Never silently downgrade TLS verification. */
internal data class RelayEndpoint(val host: String, val port: String) {
    companion object {
        fun parse(relayURL: String): RelayEndpoint {
            val uri = try { URI(relayURL) } catch (e: Exception) {
                throw IllegalArgumentException("invalid relay URL", e)
            }
            require(uri.scheme == "https" || uri.scheme == "quic") {
                "relay URL must use https or quic"
            }
            require(!uri.host.isNullOrBlank() && uri.userInfo == null &&
                (uri.path.isNullOrEmpty() || uri.path == "/") && uri.query == null && uri.fragment == null) {
                "relay URL must contain only a host and optional port"
            }
            require(uri.port == -1 || uri.port in 1..65535) { "invalid relay port" }
            return RelayEndpoint(uri.host, (if (uri.port == -1) 443 else uri.port).toString())
        }
    }
}

/**
 * libpigeon's version-1 420-byte pairing record (pigeon_pairing_record_serialize).
 * Encode explicitly rather than passing a JVM object with an unstable JNI layout.
 * Native code MUST validate the magic, version, length and NUL-terminated fields.
 */
internal fun PairingRecord.toNativeBytes(): ByteArray {
    val out = ByteArray(420)
    out[0] = 'P'.code.toByte()
    out[1] = 'G'.code.toByte()
    out[2] = 'R'.code.toByte()
    out[3] = 1
    fun text(s: String, offset: Int, width: Int) {
        val bytes = s.toByteArray(Charsets.UTF_8)
        require(bytes.size < width && !s.contains('\u0000')) { "pairing record field exceeds ${width - 1} UTF-8 bytes or contains NUL" }
        bytes.copyInto(out, offset)
    }
    fun key(bytes: ByteArray, offset: Int) {
        require(bytes.size == 32) { "pairing record key must be 32 bytes" }
        bytes.copyInto(out, offset)
    }
    text(peerInstanceID, 4, 64)
    text(relayURL, 68, 256)
    key(localPrivateKey, 324)
    key(localPublicKey, 356)
    key(peerPublicKey, 388)
    return out
}

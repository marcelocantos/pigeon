// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package com.marcelocantos.pigeon.peer

import com.marcelocantos.pigeon.crypto.PairingRecord
import kotlin.test.Test
import kotlin.test.assertContentEquals
import kotlin.test.assertEquals
import kotlin.test.assertFailsWith

class RelayClientParametersTest {
    @Test fun relayURLHostPortAndSecurity() {
        assertEquals(RelayEndpoint("relay.example", "443"), RelayEndpoint.parse("https://relay.example"))
        assertEquals(RelayEndpoint("::1", "7443"), RelayEndpoint.parse("quic://[::1]:7443"))
        for (url in listOf("http://relay.example", "https://user@relay.example", "https://relay.example/path",
            "https://relay.example:0", "https://relay.example:65536", "https://relay.example?token=secret", "oops")) {
            assertFailsWith<IllegalArgumentException>(url) { RelayEndpoint.parse(url) }
        }
    }

    @Test fun pairingRecordUsesCWireLayoutAndRejectsTruncation() {
        val priv = ByteArray(32) { it.toByte() }
        val pub = ByteArray(32) { (it + 32).toByte() }
        val peer = ByteArray(32) { (it + 64).toByte() }
        val record = PairingRecord("peer", "https://relay.example", priv, pub, peer)
        val bytes = record.toNativeBytes()
        assertEquals(420, bytes.size)
        assertContentEquals(byteArrayOf('P'.code.toByte(), 'G'.code.toByte(), 'R'.code.toByte(), 1), bytes.copyOfRange(0, 4))
        assertContentEquals("peer".toByteArray(), bytes.copyOfRange(4, 8))
        assertEquals(0, bytes[8].toInt())
        assertContentEquals("https://relay.example".toByteArray(), bytes.copyOfRange(68, 89))
        assertContentEquals(priv, bytes.copyOfRange(324, 356))
        assertContentEquals(pub, bytes.copyOfRange(356, 388))
        assertContentEquals(peer, bytes.copyOfRange(388, 420))
        assertFailsWith<IllegalArgumentException> { record.copy(peerInstanceID = "x".repeat(64)).toNativeBytes() }
        assertFailsWith<IllegalArgumentException> { record.copy(peerPublicKey = byteArrayOf(1)).toNativeBytes() }
        assertFailsWith<IllegalArgumentException> { record.copy(relayURL = "a\u0000b").toNativeBytes() }
    }
}

// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package com.marcelocantos.pigeon.peer

import com.marcelocantos.pigeon.crypto.PairingRecord
import kotlinx.coroutines.runBlocking
import kotlin.test.Test
import kotlin.test.assertContentEquals
import kotlin.test.assertEquals
import kotlin.test.assertFailsWith
import kotlin.test.assertTrue

class RelayClientTest {
    private class Driver : RelayClientDriver {
        val events = mutableListOf<String>()
        var lastRecord: ByteArray? = null
        var lastToken: String? = null
        var lastTLS = RelayTLS()
        var failConnect = false
        var failStreamClose = false
        override fun connect(endpoint: RelayEndpoint, instanceID: String, token: String?, record: ByteArray?, deviceID: String?, tls: RelayTLS): Long {
            events += "connect:${endpoint.host}:${endpoint.port}:$instanceID"
            lastRecord = record
            lastToken = token
            lastTLS = tls
            if (failConnect) throw IllegalStateException("dial rejected")
            return 7
        }
        override fun primaryStream(connection: Long): Long { events += "primary:$connection"; return 9 }
        override fun openStream(connection: Long, name: String): Long {
            events += "open:$connection:$name"
            return 8
        }
        override fun send(stream: Long, data: ByteArray) { events += "send:$stream:${String(data)}" }
        override fun recv(stream: Long): ByteArray { events += "recv:$stream"; return "pong".toByteArray() }
        override fun closeStream(stream: Long) {
            events += "close-stream:$stream"
            if (failStreamClose) throw IllegalStateException("stream close failed")
        }
        override fun closeConnection(connection: Long) { events += "close-connection:$connection" }
    }

    @Test fun pairingHandoffAndOwnedStream() = runBlocking {
        val d = Driver()
        val record = PairingRecord("peer", "https://relay.example", ByteArray(32), ByteArray(32), ByteArray(32))
        val session = connectRelayClient(d, "https://relay.example:4433", "peer", null, record, "device")
        assertEquals(null, d.lastToken)
        assertEquals(RelayTLS(), d.lastTLS)
        assertContentEquals(record.toNativeBytes(), d.lastRecord)
        assertFailsWith<IllegalArgumentException> { connectRelayClient(d, "https://relay.example", "peer", "auth-token") }
        val stream = session.openStream("chat")
        stream.send("ping".toByteArray())
        assertContentEquals("pong".toByteArray(), stream.recv())
        session.close()
        session.close()
        stream.close()
        assertEquals(listOf("connect:relay.example:4433:peer", "open:7:chat", "send:8:ping", "recv:8",
            "close-stream:8", "close-connection:7"), d.events)
        assertFailsWith<IllegalStateException> { session.openStream("again") }
        assertFailsWith<IllegalStateException> { stream.send(byteArrayOf(1)) }
    }

    @Test fun invalidHandoffNeverDialsAndDialErrorsPropagate() = runBlocking {
        val d = Driver()
        val record = PairingRecord("other", "https://relay.example", ByteArray(32), ByteArray(32), ByteArray(32))
        assertFailsWith<IllegalArgumentException> {
            connectRelayClient(d, "https://relay.example", "peer", pairing = record)
        }
        assertFailsWith<IllegalArgumentException> { connectRelayClient(d, "http://relay.example", "peer") }
        assertTrue(d.events.isEmpty())
        d.failConnect = true
        assertFailsWith<IllegalStateException> { connectRelayClient(d, "https://relay.example", "peer") }
        assertEquals(listOf("connect:relay.example:443:peer"), d.events)
    }

    @Test fun localTLSOptOutMustBeExplicit() = runBlocking {
        val d = Driver()
        val session = connectRelayClient(d, "https://localhost", "peer", tls = RelayTLS(verifyPeer = false))
        assertEquals(false, d.lastTLS.verifyPeer)
        session.close()
        assertFailsWith<IllegalArgumentException> { RelayTLS(timeoutMs = -1) }
    }

    @Test fun connectionAlwaysClosesEvenWhenStreamCloseThrows() = runBlocking {
        val d = Driver()
        val session = connectRelayClient(d, "quic://relay.example", "peer")
        session.openStream("a")
        d.failStreamClose = true
        assertFailsWith<IllegalStateException> { session.close() }
        assertEquals("close-connection:7", d.events.last())
    }
}

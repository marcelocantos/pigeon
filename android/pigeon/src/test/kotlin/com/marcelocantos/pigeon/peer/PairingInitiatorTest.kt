// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package com.marcelocantos.pigeon.peer

import com.marcelocantos.pigeon.crypto.JsonParser
import kotlinx.coroutines.runBlocking
import java.util.ArrayDeque
import java.util.Base64
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFailsWith
import kotlin.test.assertTrue
import kotlin.test.assertContentEquals

class PairingInitiatorTest {
    private val acceptorEph = ByteArray(32) { (it + 1).toByte() }
    private val acceptorId = ByteArray(32) { (it + 50).toByte() }
    private val b64 = Base64.getEncoder()
    private fun token(): String = Base64.getUrlEncoder().withoutPadding().encodeToString(
        """{"relay":"https://relay.example","session_id":"ephemeral-relay-id","acceptor_eph_pub":"${b64.encodeToString(acceptorEph)}","acceptor_id_pub":"${b64.encodeToString(acceptorId)}","acceptor_instance":"stable-backend"}""".toByteArray()
    )
    private class Driver(welcome: String) : RelayClientDriver {
        val sent = mutableListOf<String>()
        val closed = mutableListOf<String>()
        val inbound = ArrayDeque(listOf(welcome.toByteArray(), """{"kind":"confirm"}""".toByteArray()))
        var selectedStream = ""
        var dialedInstance = ""
        override fun connect(endpoint: RelayEndpoint, instanceID: String, token: String?, record: ByteArray?, deviceID: String?): Long {
            dialedInstance = instanceID; return 7
        }
        override fun primaryStream(connection: Long): Long = error("ceremony must use named stream")
        override fun openStream(connection: Long, name: String): Long { selectedStream = name; return 8 }
        override fun send(stream: Long, data: ByteArray) { sent += String(data, Charsets.UTF_8) }
        override fun recv(stream: Long): ByteArray = inbound.removeFirst()
        override fun closeStream(stream: Long) { closed += "stream" }
        override fun closeConnection(connection: Long) { closed += "connection" }
    }

    private fun welcome(eph: ByteArray = acceptorEph, id: ByteArray = acceptorId): String =
        """{"kind":"welcome","eph_pub":"${b64.encodeToString(eph)}","identity_pub":"${b64.encodeToString(id)}","instance_id":"stable-backend"}"""

    @Test fun goTokenCeremonyAndPlaintextMessageBoundaries() = runBlocking {
        val d = Driver(welcome())
        var seenCode = ""
        val result = pairInitiatorWithConnector(token(), "device-1", ByteArray(32) { 9 },
            { code -> seenCode = code; assertEquals(2, d.sent.size); true },
            { url, id -> connectRelayClient(d, url, id) })
        assertEquals("ephemeral-relay-id", d.dialedInstance)
        assertEquals("ceremony", d.selectedStream)
        assertTrue(seenCode.matches(Regex("[0-9]{6}")))
        assertEquals(listOf("hello", "reveal", "confirm"), d.sent.map { JsonParser(it).parseObject()["kind"] })
        assertTrue(d.sent[0].startsWith("{\"kind\":\"hello\""), "JSON, not a duplicate 4-byte frame")
        assertTrue("eph_pub" !in d.sent[0], "eph must remain hidden until reveal")
        assertEquals("stable-backend", result.peerInstanceID)
        assertEquals("https://relay.example", result.relayURL)
        assertContentEquals(acceptorEph, result.peerPublicKey)
        assertEquals(listOf("stream", "connection"), d.closed)
    }

    @Test fun tokenSubstitutionAndUserRejectionCannotIssueRecord() = runBlocking {
        val wrong = Driver(welcome(eph = ByteArray(32) { 42 }))
        assertFailsWith<IllegalArgumentException> {
            pairInitiatorWithConnector(token(), "device-1", ByteArray(32), { true },
                { url, id -> connectRelayClient(wrong, url, id) })
        }
        assertEquals(1, wrong.sent.size)
        assertEquals(listOf("stream", "connection"), wrong.closed)
        val denied = Driver(welcome())
        assertFailsWith<IllegalStateException> {
            pairInitiatorWithConnector(token(), "device-1", ByteArray(32), { false },
                { url, id -> connectRelayClient(denied, url, id) })
        }
        assertEquals(2, denied.sent.size)
        assertEquals(listOf("stream", "connection"), denied.closed)
    }
}

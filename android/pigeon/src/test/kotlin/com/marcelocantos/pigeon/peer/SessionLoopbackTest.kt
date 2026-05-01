// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package com.marcelocantos.pigeon.peer

import kotlinx.coroutines.runBlocking
import kotlin.test.Test
import kotlin.test.assertContentEquals
import kotlin.test.assertEquals
import kotlin.test.assertNotNull
import kotlin.test.assertNull

/**
 * Integration tests for the high-level Kotlin peer API ([Session],
 * [Stream], [Datagram]) wired over the C-side loopback transport.
 *
 * These mirror c/test/test_pigeon.c::test_session_stream_roundtrip and
 * test_session_datagram_roundtrip, exercising the same code path
 * through the JNI bridge.
 */
class SessionLoopbackTest {

    @Test
    fun streamRoundTripChat() = runBlocking {
        val key = ByteArray(32) { it.toByte() }
        val chA = Channel.shared(key)
        val chB = Channel.shared(key)
        val (a, b) = Session.loopbackPair(
            chA, chB,
            aIsBackend = true, aTag = 0xcafef00d.toInt(),
            bIsBackend = false, bTag = 0,
        )
        try {
            // A opens "chat"; B accepts the next inbound stream.
            val aChat = a.openStream("chat")
            val bChat = b.acceptStream()
            assertNotNull(bChat, "B should accept the new stream")
            assertEquals("chat", bChat.name)

            // A → B: hello
            aChat.send("hello".toByteArray())
            val got1 = bChat.recv()
            assertContentEquals("hello".toByteArray(), got1)

            // B → A: world
            bChat.send("world".toByteArray())
            val got2 = aChat.recv()
            assertContentEquals("world".toByteArray(), got2)

            aChat.close()
            bChat.close()
        } finally {
            a.close()
            b.close()
        }
    }

    @Test
    fun emptyAcceptReturnsNull() = runBlocking {
        val key = ByteArray(32) { (it + 7).toByte() }
        val chA = Channel.shared(key)
        val chB = Channel.shared(key)
        val (a, b) = Session.loopbackPair(
            chA, chB,
            aIsBackend = false, aTag = 0,
            bIsBackend = false, bTag = 0,
        )
        try {
            // Nothing was opened: accept must yield null without blocking.
            assertNull(b.acceptStream())
        } finally {
            a.close()
            b.close()
        }
    }

    @Test
    fun datagramRoundTripPing() = runBlocking {
        val key = ByteArray(32) { (it * 3 + 1).toByte() }
        val chA = Channel.shared(key, mode = Channel.MODE_DATAGRAMS)
        val chB = Channel.shared(key, mode = Channel.MODE_DATAGRAMS)
        val chans = listOf(
            DatagramChannelDef("ping", 1L),
            DatagramChannelDef("metric", 2L),
        )
        // Both sides client-mode here (matching the C test, which
        // skips the 4-byte tag prefix to make the round-trip easy).
        val (a, b) = Session.loopbackPair(
            chA, chB,
            aIsBackend = false, aTag = 0,
            bIsBackend = false, bTag = 0,
            datagramChannels = chans,
        )
        try {
            val aPing = a.getDatagram("ping")
            val bPing = b.getDatagram("ping")
            try {
                aPing.send("p1".toByteArray())
                val got = bPing.recv()
                assertContentEquals("p1".toByteArray(), got)
            } finally {
                aPing.close()
                bPing.close()
            }
        } finally {
            a.close()
            b.close()
        }
    }

    @Test
    fun blockingApiAlsoWorks() {
        val key = ByteArray(32) { (it - 4).toByte() }
        val chA = Channel.shared(key)
        val chB = Channel.shared(key)
        val (a, b) = Session.loopbackPair(
            chA, chB,
            aIsBackend = true, aTag = 0x11223344,
            bIsBackend = false, bTag = 0,
        )
        try {
            val aS = a.openStreamBlocking("control")
            val bS = b.acceptStreamBlocking()
            assertNotNull(bS)
            assertEquals("control", bS.name)
            aS.sendBlocking("howdy".toByteArray())
            assertContentEquals("howdy".toByteArray(), bS.recvBlocking())
            aS.close()
            bS.close()
        } finally {
            a.close()
            b.close()
        }
    }
}

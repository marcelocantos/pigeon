// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package com.marcelocantos.pigeon.peer

import java.util.concurrent.ConcurrentHashMap
import java.util.concurrent.LinkedBlockingDeque
import java.util.concurrent.atomic.AtomicLong
import kotlin.test.Test
import kotlin.test.assertContentEquals
import kotlin.test.assertEquals
import kotlin.test.assertNotNull
import kotlin.test.assertNull

/**
 * Exercise the JVM-callback transport (T36) end-to-end through the
 * JNI bridge. Two Kotlin transports are paired in-memory; libpigeon
 * drives both sessions through them, applying all AEAD / wire framing
 * on the C side while the byte pipes themselves live entirely in
 * Kotlin code.
 *
 * Symmetric with [SessionLoopbackTest] (C-side loopback) — same
 * round-trips, same wire bytes, exercised through the opposite
 * direction of the C ↔ JVM boundary.
 */
class JniTransportLoopbackTest {

    /**
     * In-memory shared wire used by two [LoopbackJvmTransport]
     * endpoints. Each shared wire is one direction of message flow
     * — the writer's `out` is the reader's `in`.
     *
     * Stream messages key off the opaque handle minted by
     * [LoopbackJvmTransport.openStream]; both sides see the same
     * handle value for the same logical stream because the opener
     * publishes the handle into the peer's accept queue.
     */
    private class SharedWire {
        val streamMsgs = ConcurrentHashMap<Long, LinkedBlockingDeque<ByteArray>>()
        val acceptQueue = LinkedBlockingDeque<Long>()
        val datagrams = LinkedBlockingDeque<ByteArray>()
    }

    /**
     * Kotlin-side loopback transport. Pair two of these by routing
     * each endpoint's `outbound` to the other's `inbound`. A small
     * package-private handle counter mints monotonically-increasing
     * non-zero Long handles, which is all libpigeon needs.
     */
    private class LoopbackJvmTransport(
        private val inbound: SharedWire,
        private val outbound: SharedWire,
    ) : JniQuicTransport {

        override fun openStream(): Long {
            val h = nextHandle.incrementAndGet()
            // Publish the new handle into the peer's accept queue so
            // its acceptStream() returns it; the queue identity is
            // shared with sendOnStream / recvOnStream below.
            outbound.acceptQueue.put(h)
            return h
        }

        override fun acceptStream(): Long {
            // Non-blocking poll — matches the loopback contract where
            // null/zero means "no stream pending right now".
            return inbound.acceptQueue.pollFirst() ?: 0L
        }

        override fun sendOnStream(handle: Long, data: ByteArray) {
            outbound.streamMsgs.computeIfAbsent(handle) { LinkedBlockingDeque() }
                .put(data.copyOf())
        }

        override fun recvOnStream(handle: Long): ByteArray {
            val q = inbound.streamMsgs.computeIfAbsent(handle) { LinkedBlockingDeque() }
            // Block until a message arrives or the test times out the
            // whole JVM. Tests open + send synchronously so a finite
            // poll is fine; use take() for the production semantic.
            return q.takeFirst()
        }

        override fun closeStream(handle: Long) {
            outbound.streamMsgs.remove(handle)
            inbound.streamMsgs.remove(handle)
        }

        override fun sendDatagram(data: ByteArray) {
            outbound.datagrams.put(data.copyOf())
        }

        override fun recvDatagram(): ByteArray {
            return inbound.datagrams.takeFirst()
        }

        override fun close() {
            // Best-effort drain so test resources don't linger across
            // runs. Idempotent; safe to call from defer / finally.
            outbound.streamMsgs.clear()
            inbound.streamMsgs.clear()
            outbound.acceptQueue.clear()
            inbound.acceptQueue.clear()
            outbound.datagrams.clear()
            inbound.datagrams.clear()
        }

        companion object {
            private val nextHandle = AtomicLong(0)

            fun pair(): Pair<LoopbackJvmTransport, LoopbackJvmTransport> {
                val a = SharedWire()
                val b = SharedWire()
                return LoopbackJvmTransport(a, b) to LoopbackJvmTransport(b, a)
            }
        }
    }

    @Test
    fun streamRoundTripChat() {
        val key = ByteArray(32) { it.toByte() }
        val chA = Channel.shared(key)
        val chB = Channel.shared(key)
        val (tA, tB) = LoopbackJvmTransport.pair()
        val a = Session.fromTransport(chA, tA,
            isBackend = true, clientTag = 0xcafef00d.toInt())
        val b = Session.fromTransport(chB, tB,
            isBackend = false, clientTag = 0)
        try {
            // A opens "chat"; B accepts the new inbound stream. The C
            // side wrote the unencrypted name-binding header through
            // tA.sendOnStream, the peer reads it via tB.recvOnStream
            // — both calls thunk through the JNI bridge under the
            // hood of openStream / acceptStreamBlocking.
            val aChat = a.openStreamBlocking("chat")
            val bChat = b.acceptStreamBlocking()
            assertNotNull(bChat, "B should accept the new stream")
            assertEquals("chat", bChat.name)

            // A → B
            aChat.sendBlocking("hello".toByteArray())
            assertContentEquals("hello".toByteArray(), bChat.recvBlocking())

            // B → A
            bChat.sendBlocking("world".toByteArray())
            assertContentEquals("world".toByteArray(), aChat.recvBlocking())

            aChat.close()
            bChat.close()
        } finally {
            a.close()
            b.close()
            tA.close()
            tB.close()
        }
    }

    @Test
    fun emptyAcceptReturnsNull() {
        val key = ByteArray(32) { (it + 7).toByte() }
        val chA = Channel.shared(key)
        val chB = Channel.shared(key)
        val (tA, tB) = LoopbackJvmTransport.pair()
        val a = Session.fromTransport(chA, tA,
            isBackend = false, clientTag = 0)
        val b = Session.fromTransport(chB, tB,
            isBackend = false, clientTag = 0)
        try {
            // Nothing opened — acceptStreamBlocking sees the empty
            // accept-queue and returns 0 from acceptStream(), which
            // the Kotlin wrapper translates to null.
            assertNull(b.acceptStreamBlocking())
        } finally {
            a.close()
            b.close()
            tA.close()
            tB.close()
        }
    }

    @Test
    fun datagramRoundTripPing() {
        val key = ByteArray(32) { (it * 3 + 1).toByte() }
        val chA = Channel.shared(key, mode = Channel.MODE_DATAGRAMS)
        val chB = Channel.shared(key, mode = Channel.MODE_DATAGRAMS)
        val chans = listOf(
            DatagramChannelDef("ping", 1L),
            DatagramChannelDef("metric", 2L),
        )
        val (tA, tB) = LoopbackJvmTransport.pair()
        // Both sides client-mode here (matches the C-side
        // SessionLoopbackTest.datagramRoundTripPing setup, which skips
        // the 4-byte tag prefix for round-trip simplicity).
        val a = Session.fromTransport(chA, tA,
            isBackend = false, clientTag = 0, datagramChannels = chans)
        val b = Session.fromTransport(chB, tB,
            isBackend = false, clientTag = 0, datagramChannels = chans)
        try {
            val aPing = a.getDatagram("ping")
            val bPing = b.getDatagram("ping")
            try {
                aPing.sendBlocking("p1".toByteArray())
                assertContentEquals("p1".toByteArray(), bPing.recvBlocking())

                bPing.sendBlocking("p2".toByteArray())
                assertContentEquals("p2".toByteArray(), aPing.recvBlocking())
            } finally {
                aPing.close()
                bPing.close()
            }
        } finally {
            a.close()
            b.close()
            tA.close()
            tB.close()
        }
    }

    @Test
    fun multipleStreamsRoundTrip() {
        val key = ByteArray(32) { (it - 4).toByte() }
        val chA = Channel.shared(key)
        val chB = Channel.shared(key)
        val (tA, tB) = LoopbackJvmTransport.pair()
        val a = Session.fromTransport(chA, tA,
            isBackend = true, clientTag = 0x11223344)
        val b = Session.fromTransport(chB, tB,
            isBackend = false, clientTag = 0)
        try {
            // Open three independent streams, accept each, and round-
            // trip a tagged payload through each. Exercises the
            // open / accept / send / recv vtable slots multiple times
            // and confirms libpigeon dispatches per-stream correctly.
            val outs = listOf("control", "data", "diagnostics").map { name ->
                val ao = a.openStreamBlocking(name)
                val bi = b.acceptStreamBlocking()
                assertNotNull(bi, "B should accept $name")
                assertEquals(name, bi.name)
                ao to bi
            }

            for ((ao, bi) in outs) {
                val payload = "from-${bi.name}".toByteArray()
                ao.sendBlocking(payload)
                assertContentEquals(payload, bi.recvBlocking())
            }
            for ((ao, bi) in outs) {
                ao.close()
                bi.close()
            }
        } finally {
            a.close()
            b.close()
            tA.close()
            tB.close()
        }
    }
}

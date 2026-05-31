// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package com.marcelocantos.pigeon.peer

import tech.kwik.core.QuicConnection
import tech.kwik.core.QuicStream
import java.io.DataInputStream
import java.io.DataOutputStream
import java.io.EOFException
import java.io.IOException
import java.util.concurrent.ConcurrentHashMap
import java.util.concurrent.LinkedBlockingDeque
import java.util.concurrent.TimeUnit
import java.util.concurrent.atomic.AtomicLong

/**
 * [JniQuicTransport] backed by a Kwik [QuicConnection] (T36 Phase 2).
 *
 * Each `sendOnStream` / `recvOnStream` delivers a discrete pigeon
 * message. QUIC streams are byte streams, so the adapter prepends a
 * 4-byte big-endian length header to every message and reads it back on
 * the receive side — this gives libpigeon the message-boundary semantic
 * its loopback transport already exposes.
 *
 * Stream handles are opaque non-zero `Long`s minted from a monotonically
 * increasing counter; `handle → QuicStream` lives in a concurrent map.
 *
 * Inbound wiring is **not** done in the constructor — the caller hooks
 * either Kwik's connection-level `setPeerInitiatedStreamCallback` (client
 * side) or an `ApplicationProtocolConnection.acceptPeerInitiatedStream`
 * override (server side) to [acceptIncomingStream] / [acceptIncomingDatagram].
 * Keeping intake explicit avoids duplicate callbacks and lets the same
 * adapter serve both roles.
 */
class JniKwikTransport(
    private val connection: QuicConnection,
    private val recvTimeout: Long = 30,
    private val recvTimeoutUnit: TimeUnit = TimeUnit.SECONDS,
) : JniQuicTransport {

    private class StreamCtx(val stream: QuicStream) {
        val output = DataOutputStream(stream.outputStream)
        val input = DataInputStream(stream.inputStream)
    }

    private val nextHandle = AtomicLong(1)
    private val streams = ConcurrentHashMap<Long, StreamCtx>()
    private val inbound = LinkedBlockingDeque<Long>()
    private val datagrams = LinkedBlockingDeque<ByteArray>()

    private fun registerStream(stream: QuicStream): Long {
        val handle = nextHandle.getAndIncrement()
        streams[handle] = StreamCtx(stream)
        return handle
    }

    private fun ctx(handle: Long): StreamCtx =
        streams[handle] ?: throw IllegalStateException("unknown stream handle: $handle")

    /** Hand a peer-initiated stream to the transport's accept queue. */
    fun acceptIncomingStream(stream: QuicStream) {
        inbound.put(registerStream(stream))
    }

    /** Hand a received datagram to the transport's recv queue. */
    fun acceptIncomingDatagram(data: ByteArray) {
        datagrams.put(data)
    }

    override fun openStream(): Long {
        val stream = connection.createStream(true)
        return registerStream(stream)
    }

    override fun acceptStream(): Long {
        // Block (with a generous timeout) on a peer-initiated stream.
        // Returning 0 signals "no stream" to libpigeon, which the Kotlin
        // wrapper translates to null. The timeout exists so a hung test
        // can't pin the JVM forever.
        return inbound.pollFirst(recvTimeout, recvTimeoutUnit) ?: 0L
    }

    override fun sendOnStream(handle: Long, data: ByteArray) {
        val c = ctx(handle)
        c.output.writeInt(data.size)
        if (data.isNotEmpty()) c.output.write(data)
        c.output.flush()
    }

    override fun recvOnStream(handle: Long): ByteArray {
        val c = ctx(handle)
        val len = try {
            c.input.readInt()
        } catch (e: EOFException) {
            throw IllegalStateException("stream $handle: peer closed", e)
        }
        require(len >= 0) { "negative message length on stream $handle: $len" }
        if (len == 0) return ByteArray(0)
        val buf = ByteArray(len)
        c.input.readFully(buf)
        return buf
    }

    override fun closeStream(handle: Long) {
        val c = streams.remove(handle) ?: return
        try { c.output.close() } catch (_: IOException) {}
    }

    override fun sendDatagram(data: ByteArray) {
        connection.sendDatagram(data)
    }

    override fun recvDatagram(): ByteArray {
        return datagrams.pollFirst(recvTimeout, recvTimeoutUnit)
            ?: throw IllegalStateException("datagram receive timed out")
    }

    override fun close() {
        streams.values.forEach { c ->
            try { c.output.close() } catch (_: IOException) {}
        }
        streams.clear()
        inbound.clear()
        datagrams.clear()
        try { connection.close() } catch (_: Exception) {}
    }
}

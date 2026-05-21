// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package com.marcelocantos.pigeon.peer

/**
 * Transport vtable invoked from C via JNI (T36).
 *
 * libpigeon's C core drives all AEAD and wire framing; the JVM
 * implementation supplies the underlying QUIC byte pipes (Kwik on
 * desktop JVM, Cronet / native ngtcp2 on Android). Each method maps
 * one-to-one onto a slot in the C `pigeon_transport` vtable
 * (`open_stream`, `accept_stream`, `send_on_stream`, `recv_on_stream`,
 * `close_stream`, `send_datagram`, `recv_datagram`).
 *
 * The C side resolves these methods by *name and signature*, not by
 * type — so any class that declares them (with the exact signatures
 * below) plugs into a [Session]. Implementations should normally
 * declare `: JniQuicTransport` for clarity, but the JNI shim only
 * cares about the method shapes.
 *
 * Stream handle protocol: [openStream] and [acceptStream] mint
 * opaque `Long` values that the C side stores in
 * `pigeon_stream_handle*` and hands back on per-stream calls
 * ([sendOnStream], [recvOnStream], [closeStream]). The transport
 * picks the mapping — typically the address of a Kotlin-managed
 * descriptor, or an index into an internal table. `0` is the
 * reserved "no stream" sentinel (open-failure / accept-empty).
 *
 * Error model: methods throw on hard failure. The JNI bridge clears
 * the pending exception and returns -1 to libpigeon, which surfaces
 * as a `RuntimeException` at the next API boundary. There is no
 * separate error-return code path.
 *
 * Naming note: lifecycle teardown is `close()` (via [AutoCloseable]),
 * not `cancel` — the transport is a resource handle that defers
 * cleanly.
 *
 * Threading: invoked on whichever JVM thread is currently driving
 * into libpigeon. Implementations must be safe to call from any
 * thread (libpigeon itself is single-threaded per-session, but the
 * application may drive multiple sessions concurrently).
 */
interface JniQuicTransport : AutoCloseable {
    /**
     * Open a new outbound bidirectional stream. Returns an opaque
     * non-zero handle on success; `0` signals failure.
     */
    fun openStream(): Long

    /**
     * Accept the next inbound bidirectional stream. Returns an
     * opaque non-zero handle when a stream is available. Returns
     * `0` if no stream is currently pending — libpigeon translates
     * that to a "no stream" outcome (e.g. [Session.acceptStream]
     * returning null). Throw on hard transport failure.
     *
     * Implementations may block waiting for a stream; the call is
     * driven by the JVM thread that invoked [Session.acceptStream] /
     * [Session.acceptStreamBlocking], so blocking semantics are the
     * caller's responsibility.
     */
    fun acceptStream(): Long

    /**
     * Send one framed message on the supplied stream. The message
     * already contains any AEAD framing libpigeon applies — the
     * transport must deliver it byte-for-byte.
     */
    fun sendOnStream(handle: Long, data: ByteArray)

    /**
     * Receive the next framed message on the supplied stream. Returns
     * the message bytes (may be empty for a zero-byte payload — that's
     * a legitimate value libpigeon handles). Throw on transport
     * failure or peer close.
     */
    fun recvOnStream(handle: Long): ByteArray

    /**
     * Close the supplied stream. Idempotent — repeated closes on the
     * same handle should be no-ops.
     */
    fun closeStream(handle: Long)

    /** Send one connection-level datagram. */
    fun sendDatagram(data: ByteArray)

    /**
     * Receive the next connection-level datagram. Throw on transport
     * failure; return an empty array only if the peer legitimately
     * sent a zero-byte datagram.
     */
    fun recvDatagram(): ByteArray

    /**
     * Release transport resources. Always callable; idempotent. After
     * close, every other method may throw. Defer-friendly: usage is
     * `transport.use { ... }` or an explicit `try / finally`.
     */
    override fun close()
}

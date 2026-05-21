// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package com.marcelocantos.pigeon.peer

import com.marcelocantos.pigeon.jni.PigeonNative
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.flow
import kotlinx.coroutines.flow.flowOn
import java.util.concurrent.atomic.AtomicLong

/**
 * Idiomatic Kotlin API for the pigeon peer library.
 *
 * This is a thin wrapper over [PigeonNative] (the JNI bindings to
 * `libpigeon-jni`). All wire-format and crypto logic lives in the C
 * library; the Kotlin types provide:
 *
 *   - structured ownership ([AutoCloseable]),
 *   - `suspend` send/recv that hop to [Dispatchers.IO] so callers
 *     don't accidentally block the main dispatcher,
 *   - [Flow] of inbound messages on a stream.
 *
 * The shape mirrors `pigeon.Session` (Go) and `Pigeon.Session`
 * (Swift). Application code on Android / desktop JVM should use
 * these classes directly; the older `com.marcelocantos.pigeon.relay.PigeonConn`
 * remains for backwards-compat with the legacy single-channel wire
 * (deprecated — to be removed once the multi-stream ngtcp2 transport
 * lands and we have a Java-callback transport bridge).
 */

/** Pre-declared datagram channel binding. Mirror the C `pigeon_dgchannel_def`. */
data class DatagramChannelDef(val name: String, val id: Long)

/**
 * Symmetric AEAD channel created from a shared secret. The session
 * holds an owned [Channel] and frees it when the session is closed.
 *
 * Construct directly when you have an explicit key (e.g. derived from
 * a [com.marcelocantos.pigeon.crypto.PairingArtifact]); use [Channel.shared]
 * for the symmetric (same-key-both-sides) case.
 */
class Channel internal constructor(internal val handle: Long) : AutoCloseable {

    companion object {
        const val MODE_STRICT = 0
        const val MODE_DATAGRAMS = 1

        /** Channel with separate send / recv keys. */
        fun directional(sendKey: ByteArray, recvKey: ByteArray, mode: Int = MODE_STRICT): Channel {
            require(sendKey.size == 32 && recvKey.size == 32) { "keys must be 32 bytes" }
            return Channel(PigeonNative.channelInit(sendKey, recvKey, mode))
        }

        /** Symmetric channel (both sides use the same key). */
        fun shared(key: ByteArray, mode: Int = MODE_STRICT): Channel {
            require(key.size == 32) { "key must be 32 bytes" }
            return Channel(PigeonNative.channelInit(key, key, mode))
        }
    }

    private var closed = false

    fun encrypt(plaintext: ByteArray): ByteArray {
        check(!closed) { "channel closed" }
        return PigeonNative.channelEncrypt(handle, plaintext)
    }

    fun decrypt(ciphertext: ByteArray): ByteArray {
        check(!closed) { "channel closed" }
        return PigeonNative.channelDecrypt(handle, ciphertext)
    }

    override fun close() {
        if (closed) return
        closed = true
        PigeonNative.channelFree(handle)
    }
}

/**
 * A peer-to-peer association: backend ↔ one paired client (or vice
 * versa from the client side). Wraps a `pigeon_session*` and the
 * underlying transport.
 *
 * Two transport flavours are supported:
 *
 *   - The in-process C loopback (built via [Session.loopbackPair]).
 *     Used by SDK tests; not for production.
 *   - A JVM-callback transport (built via [Session.fromTransport]).
 *     Each vtable slot thunks through JNI back into a Kotlin object
 *     that implements [JniQuicTransport]. This is the production
 *     path — desktop JVM (Kwik) and Android (Cronet / native ngtcp2)
 *     plug a [JniQuicTransport] in here.
 *
 * The session takes ownership of the supplied [Channel] objects and
 * closes them on [close]. The JNI-callback transport object's
 * [JniQuicTransport.close] is *not* called automatically — the
 * application owns the transport's lifetime, since the transport
 * usually outlives any single session (e.g. a relay-side transport
 * accepts many client sessions).
 */
class Session internal constructor(
    private val handleRef: AtomicLong,
    private val channel: Channel,
    val isBackend: Boolean,
    val clientTag: Int,
    private val useGenericAccept: Boolean = false,
) : AutoCloseable {

    val handle: Long get() = handleRef.get()

    /**
     * Open a new outbound stream with the given name. The C library
     * writes the unencrypted name-binding header as the first message
     * on the stream; subsequent messages are AEAD-encrypted.
     */
    suspend fun openStream(name: String): Stream = withIO {
        require(name.length <= 63) { "stream name too long" }
        Stream(PigeonNative.sessionOpenStream(handle, name), name)
    }

    /**
     * Synchronous variant of [openStream] — no dispatcher hop. Useful
     * inside tests and from non-suspend contexts.
     */
    fun openStreamBlocking(name: String): Stream {
        require(name.length <= 63) { "stream name too long" }
        return Stream(PigeonNative.sessionOpenStream(handle, name), name)
    }

    /**
     * Accept the next inbound stream and decode its header. Returns
     * null if no stream is currently pending (loopback transport
     * never blocks; a future ngtcp2-backed implementation will
     * suspend here).
     */
    suspend fun acceptStream(): Stream? = withIO { acceptStreamBlocking() }

    /** Synchronous variant of [acceptStream]. */
    fun acceptStreamBlocking(): Stream? {
        val nameOut = arrayOfNulls<String>(1)
        val h = if (useGenericAccept) {
            PigeonNative.sessionAcceptStreamGeneric(handle, nameOut)
        } else {
            PigeonNative.sessionAcceptStream(handle, nameOut)
        }
        if (h == 0L) return null
        return Stream(h, nameOut[0] ?: "")
    }

    /**
     * Look up a pre-declared datagram channel by name. Throws if the
     * name wasn't in the session's datagram-channel list.
     */
    fun getDatagram(name: String): Datagram {
        return Datagram(PigeonNative.sessionGetDatagram(handle, name), name)
    }

    override fun close() {
        val h = handleRef.getAndSet(0L)
        if (h != 0L) PigeonNative.sessionFree(h)
        channel.close()
    }

    companion object {
        /**
         * Create two sessions wired together over an in-process
         * loopback transport. Returns the (a, b) pair. Convenient
         * for tests; not for production.
         *
         * `aIsBackend` and `aTag` configure the backend-side framing
         * for session A; the matching `b*` parameters do the same for
         * session B. Both sessions get the same datagram-channel
         * declarations.
         */
        /**
         * Build a session over a JVM-callback transport (T36).
         *
         * `transport` must implement [JniQuicTransport]; libpigeon
         * resolves the seven vtable methods by name + signature and
         * thunks through JNI on each call. The session retains a
         * global ref to `transport` until [close], so the caller can
         * drop their local reference safely.
         *
         * `channel` is consumed by the session — closing the session
         * closes the channel. The transport's lifetime is the
         * application's responsibility (closing the session does NOT
         * call `transport.close()`).
         *
         * `isBackend` and `clientTag` mirror [loopbackPair]; the
         * combination of the two governs how stream-header framing
         * looks on outbound streams (backend prepends the 4-byte
         * clientTag prefix).
         */
        fun fromTransport(
            channel: Channel,
            transport: JniQuicTransport,
            isBackend: Boolean,
            clientTag: Int,
            datagramChannels: List<DatagramChannelDef> = emptyList(),
        ): Session {
            val names = if (datagramChannels.isEmpty()) null
                else datagramChannels.map { it.name }.toTypedArray()
            val ids = if (datagramChannels.isEmpty()) null
                else LongArray(datagramChannels.size) { datagramChannels[it].id }
            val handle = PigeonNative.sessionInitWithJniTransport(
                channel.handle, transport,
                isBackend, clientTag,
                names, ids,
            )
            return Session(AtomicLong(handle), channel, isBackend, clientTag,
                useGenericAccept = true)
        }

        fun loopbackPair(
            channelA: Channel,
            channelB: Channel,
            aIsBackend: Boolean,
            aTag: Int,
            bIsBackend: Boolean,
            bTag: Int,
            datagramChannels: List<DatagramChannelDef> = emptyList(),
        ): Pair<Session, Session> {
            val names = if (datagramChannels.isEmpty()) null
                else datagramChannels.map { it.name }.toTypedArray()
            val ids = if (datagramChannels.isEmpty()) null
                else LongArray(datagramChannels.size) { datagramChannels[it].id }
            val handles = PigeonNative.newLoopbackPair(
                channelA.handle, channelB.handle,
                aIsBackend, aTag,
                bIsBackend, bTag,
                names, ids,
            )
            val a = Session(AtomicLong(handles[0]), channelA, aIsBackend, aTag)
            val b = Session(AtomicLong(handles[1]), channelB, bIsBackend, bTag)
            return a to b
        }
    }
}

/**
 * A bidirectional, named, AEAD-encrypted stream. Read with [recv] /
 * [incoming]; write with [send].
 *
 * Streams are not goroutine-safe in C; do not call [send] from two
 * coroutines concurrently. The Kotlin wrapper does not lock on
 * principle (the underlying ngtcp2 transport will impose its own
 * thread-safety model and we want the wrapper to be a thin shell).
 */
class Stream internal constructor(
    private val handleRef: AtomicLong,
    val name: String,
) : AutoCloseable {

    internal constructor(handle: Long, name: String) : this(AtomicLong(handle), name)

    val handle: Long get() = handleRef.get()

    suspend fun send(msg: ByteArray) = withIO {
        PigeonNative.streamSend(handle, msg)
    }

    fun sendBlocking(msg: ByteArray) {
        PigeonNative.streamSend(handle, msg)
    }

    suspend fun recv(): ByteArray = withIO {
        PigeonNative.streamRecv(handle)
    }

    fun recvBlocking(): ByteArray = PigeonNative.streamRecv(handle)

    /**
     * Cold flow of inbound messages. Each collection issues blocking
     * [PigeonNative.streamRecv] calls on [Dispatchers.IO]. The flow
     * terminates when the underlying recv returns an error (e.g. the
     * peer closed the stream) — that propagates as a `RuntimeException`
     * thrown from the C layer.
     */
    val incoming: Flow<ByteArray>
        get() = flow {
            while (true) {
                emit(PigeonNative.streamRecv(handle))
            }
        }.flowOn(Dispatchers.IO)

    override fun close() {
        val h = handleRef.getAndSet(0L)
        if (h != 0L) {
            PigeonNative.streamClose(h)
            PigeonNative.streamFree(h)
        }
    }
}

/**
 * A pre-declared, AEAD-encrypted datagram channel multiplexed onto
 * the connection's single datagram pipe (QUIC has no per-stream
 * datagrams; the 1-byte-or-two channel id is part of the AEAD
 * plaintext).
 */
class Datagram internal constructor(
    private val handleRef: AtomicLong,
    val name: String,
) : AutoCloseable {

    internal constructor(handle: Long, name: String) : this(AtomicLong(handle), name)

    val handle: Long get() = handleRef.get()

    suspend fun send(payload: ByteArray) = withIO {
        PigeonNative.datagramSend(handle, payload)
    }

    fun sendBlocking(payload: ByteArray) {
        PigeonNative.datagramSend(handle, payload)
    }

    /**
     * Receive the next datagram on this channel. The C layer returns
     * an empty array if the channel-id of the next datagram doesn't
     * match this binding; in production code the caller demuxes from
     * a single recv-pump rather than calling this per-channel.
     */
    suspend fun recv(): ByteArray = withIO {
        PigeonNative.datagramRecv(handle)
    }

    fun recvBlocking(): ByteArray = PigeonNative.datagramRecv(handle)

    override fun close() {
        val h = handleRef.getAndSet(0L)
        if (h != 0L) PigeonNative.datagramFree(h)
    }
}

// --- Internals ---

private suspend inline fun <T> withIO(crossinline block: () -> T): T =
    kotlinx.coroutines.withContext(Dispatchers.IO) { block() }

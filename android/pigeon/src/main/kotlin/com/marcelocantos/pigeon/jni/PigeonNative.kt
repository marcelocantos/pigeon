// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package com.marcelocantos.pigeon.jni

/**
 * Low-level JNI bindings to the pigeon C peer library.
 *
 * Each method maps directly to a `pigeon_*` C symbol. Ownership of
 * native memory is expressed via opaque `Long` handles; the higher-
 * level Kotlin types (`com.marcelocantos.pigeon.peer.Session`,
 * `Stream`, `Datagram`) wrap these handles in `AutoCloseable` types
 * and are the API consumers should normally use.
 *
 * The shared library `libpigeon-jni` is loaded eagerly from the class
 * initialiser: the test JVM looks for it on `java.library.path`,
 * which the Gradle build feeds via `pigeon-jni.libDir`. On Android
 * `System.loadLibrary` resolves it from the APK's jniLibs folder.
 *
 * If the library can't be found at class-init time the static
 * initialiser throws `UnsatisfiedLinkError`. Callers that want a
 * graceful skip can bracket their JNI code in
 * `try { PigeonNative.ensureLoaded(); ... } catch (UnsatisfiedLinkError)`.
 */
object PigeonNative {

    init {
        // Allow the build to push native libs onto java.library.path
        // via the Gradle test configuration. Both the Android NDK and
        // a desktop JNI build use the same library name; resolution is
        // the JVM's responsibility.
        try {
            System.loadLibrary("pigeon-jni")
        } catch (e: UnsatisfiedLinkError) {
            // Re-throw with extra context so the test failure message
            // points at the Gradle build instead of a cryptic
            // 'no pigeon-jni in java.library.path' string.
            throw UnsatisfiedLinkError(
                "Failed to load libpigeon-jni: ${e.message}. " +
                "java.library.path=${System.getProperty("java.library.path")}",
            )
        }
    }

    /** No-op accessor that triggers static initialisation. */
    fun ensureLoaded() = Unit

    // --- Crypto ---

    /** Generate an X25519 key pair. Returns 64 bytes: 32-byte priv || 32-byte pub. */
    @JvmStatic external fun keypair(): ByteArray

    /** Derive a 32-byte session key from local priv + peer pub + HKDF info. */
    @JvmStatic external fun deriveSessionKey(
        priv: ByteArray,
        peerPub: ByteArray,
        info: ByteArray?,
    ): ByteArray

    /** 6-digit order-independent confirmation code. Returns the digits as a String. */
    @JvmStatic external fun deriveConfirmationCode(
        pubA: ByteArray,
        pubB: ByteArray,
    ): String

    // --- Wire helpers ---

    /** Encode an unsigned varint in Go's binary.PutUvarint format. */
    @JvmStatic external fun uvarintEncode(v: Long): ByteArray

    /**
     * Encode the post-T22 stream-header. Backend side prepends the
     * 4-byte clientTag; client side starts with the varint name length.
     */
    @JvmStatic external fun encodeStreamHeader(
        isBackend: Boolean,
        clientTag: Int,
        name: String,
    ): ByteArray

    // --- Channel ---

    /** Returns an opaque pigeon_channel pointer. mode = 0 strict, 1 datagrams. */
    @JvmStatic external fun channelInit(
        sendKey: ByteArray,
        recvKey: ByteArray,
        mode: Int,
    ): Long

    @JvmStatic external fun channelFree(handle: Long)

    @JvmStatic external fun channelEncrypt(handle: Long, plaintext: ByteArray): ByteArray

    @JvmStatic external fun channelDecrypt(handle: Long, ciphertext: ByteArray): ByteArray

    // --- Sessions (loopback test transport) ---

    /**
     * Create two paired sessions wired up over an in-process loopback
     * transport. Returns `[handleA, handleB]`. Each handle owns its
     * pigeon_session and a shared loopback_pair (refcounted).
     *
     * The two channels are typically created from the same shared key
     * (symmetric AEAD) so each side's send key matches the other's
     * recv key. The caller still owns the channel handles and must
     * free them after both sessions are closed.
     *
     * `dgChannelNames` and `dgChannelIds` declare the agreed-upon
     * datagram channel table; both peers must declare the same list.
     * Pass null/empty to skip datagram channels.
     */
    @JvmStatic external fun newLoopbackPair(
        channelA: Long,
        channelB: Long,
        aIsBackend: Boolean,
        aTag: Int,
        bIsBackend: Boolean,
        bTag: Int,
        dgChannelNames: Array<String>?,
        dgChannelIds: LongArray?,
    ): LongArray

    @JvmStatic external fun sessionFree(handle: Long)

    /** Open a fresh outbound stream on this session. Returns a pigeon_stream*. */
    @JvmStatic external fun sessionOpenStream(sessionHandle: Long, name: String): Long

    /**
     * Pop the next inbound stream off the loopback transport, decode
     * its header, and return a pigeon_stream*. `nameOut[0]` is set to
     * the decoded stream name. Returns 0 if no stream is pending.
     *
     * This is a loopback-only convenience. A real ngtcp2 transport
     * provides accept_stream natively; that path is the future
     * Java-callback-transport (deferred — see T30 commit notes).
     */
    @JvmStatic external fun sessionAcceptStream(
        sessionHandle: Long,
        nameOut: Array<String?>,
    ): Long

    @JvmStatic external fun streamFree(handle: Long)

    @JvmStatic external fun streamSend(handle: Long, msg: ByteArray)

    @JvmStatic external fun streamRecv(handle: Long): ByteArray

    @JvmStatic external fun streamClose(handle: Long)

    // --- Datagrams ---

    @JvmStatic external fun sessionGetDatagram(sessionHandle: Long, name: String): Long

    @JvmStatic external fun datagramFree(handle: Long)

    @JvmStatic external fun datagramSend(handle: Long, payload: ByteArray)

    /** Returns the application payload, or an empty array if the channel-id didn't match. */
    @JvmStatic external fun datagramRecv(handle: Long): ByteArray
}

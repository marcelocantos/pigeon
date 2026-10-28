// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package com.marcelocantos.pigeon.peer

import com.marcelocantos.pigeon.crypto.PairingRecord
import com.marcelocantos.pigeon.jni.PigeonNative
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.flow
import kotlinx.coroutines.flow.flowOn
import kotlinx.coroutines.withContext

/** TLS is verified by default; disabling verification is only for a local fixture. */
data class RelayTLS(
    val verifyPeer: Boolean = true,
    val caCertFile: String? = null,
    val timeoutMs: Int = 0,
) {
    init { require(timeoutMs >= 0) { "timeout must be nonnegative" } }
}

internal interface RelayClientDriver {
    fun connect(endpoint: RelayEndpoint, instanceID: String, token: String?, record: ByteArray?, deviceID: String?, tls: RelayTLS): Long
    fun primaryStream(connection: Long): Long
    fun openStream(connection: Long, name: String): Long
    fun send(stream: Long, data: ByteArray)
    fun recv(stream: Long): ByteArray
    fun closeStream(stream: Long)
    fun closeConnection(connection: Long)
}

/**
 * Owns one relay connection and its streams. Close streams before the native
 * connection: stream pointers borrow the session owned by that connection.
 * Operations are dispatched off the caller thread; recv does not hold a
 * session-wide lock against send. Full-duplex safety still requires the
 * native transport pump and a device recv-first oracle.
 */
class RelayClientSession internal constructor(
    private val driver: RelayClientDriver,
    private val connection: Long,
) : AutoCloseable {
    private val lock = Any()
    private val streams = mutableSetOf<RelayClientStream>()
    private var closed = false

    /** Plaintext ceremony stream; only available on a pairing-mode connection. */
    suspend fun primaryStream(): RelayClientStream = withContext(Dispatchers.IO) {
        synchronized(lock) {
            check(!closed) { "session closed" }
            val handle = driver.primaryStream(connection)
            check(handle != 0L) { "native primary stream unavailable" }
            RelayClientStream(driver, handle, "", this@RelayClientSession).also { streams.add(it) }
        }
    }

    suspend fun openStream(name: String): RelayClientStream = withContext(Dispatchers.IO) {
        require(name.isNotEmpty() && name.toByteArray(Charsets.UTF_8).size < 64) {
            "stream name must be 1..63 UTF-8 bytes"
        }
        synchronized(lock) {
            check(!closed) { "session closed" }
            val handle = driver.openStream(connection, name)
            check(handle != 0L) { "native stream open failed" }
            RelayClientStream(driver, handle, name, this@RelayClientSession).also { streams.add(it) }
        }
    }

    internal fun checkOpen() = synchronized(lock) { check(!closed) { "session closed" } }

    internal fun forget(stream: RelayClientStream) = synchronized(lock) { streams.remove(stream) }

    override fun close(): Unit = synchronized(lock) {
        if (closed) return
        closed = true
        try {
            // Closing all child streams first prevents use-after-free of
            // their borrowed native session pointer. Keep trying after an
            // individual close failure, and always close the connection.
            var failure: Throwable? = null
            for (stream in streams.toList()) {
                try { stream.close() } catch (e: Throwable) { if (failure == null) failure = e else failure.addSuppressed(e) }
            }
            streams.clear()
            if (failure != null) throw failure
        } finally {
            driver.closeConnection(connection)
        }
    }
}

class RelayClientStream internal constructor(
    private val driver: RelayClientDriver,
    private val handle: Long,
    val name: String,
    private val session: RelayClientSession,
) : AutoCloseable {
    @Volatile private var closed = false
    private val lock = Any()

    suspend fun send(data: ByteArray) = withContext(Dispatchers.IO) {
        synchronized(lock) {
            session.checkOpen()
            check(!closed) { "stream closed" }
            driver.send(handle, data)
        }
    }

    suspend fun recv(): ByteArray = withContext(Dispatchers.IO) {
        // No session-wide mutex across blocking recv. Native pump must
        // provide safe progress for send on another stream/thread.
        session.checkOpen()
        check(!closed) { "stream closed" }
        driver.recv(handle)
    }

    val incoming: Flow<ByteArray> get() = flow {
        while (true) emit(recv())
    }.flowOn(Dispatchers.IO)

    override fun close(): Unit = synchronized(lock) {
        if (closed) return
        closed = true
        try { driver.closeStream(handle) } finally { session.forget(this) }
    }
}

/**
 * Internal staging seam; the public connect(relayURL, instanceID, token?)
 * entry point awaits the native event-loop pump and device recv-first oracle.
 */
internal suspend fun connectRelayClient(
    driver: RelayClientDriver,
    relayURL: String,
    instanceID: String,
    token: String? = null,
    pairing: PairingRecord? = null,
    deviceID: String? = null,
    tls: RelayTLS = RelayTLS(),
): RelayClientSession {
    require(instanceID.isNotBlank() && instanceID.toByteArray(Charsets.UTF_8).size < 64) {
        "instance ID must be 1..63 UTF-8 bytes"
    }
    require(pairing == null || pairing.peerInstanceID == instanceID) {
        "pairing record belongs to another instance"
    }
    require(token.isNullOrEmpty()) { "CONNECT bearer token is not supported by the relay greeting" }
    require((pairing == null) == (deviceID == null)) { "activation requires both a pairing record and device ID" }
    require(deviceID == null || deviceID.isNotBlank()) { "device ID must not be blank" }
    val endpoint = RelayEndpoint.parse(relayURL)
    return withContext(Dispatchers.IO) {
        val connection = driver.connect(endpoint, instanceID, token, pairing?.toNativeBytes(), deviceID, tls)
        check(connection != 0L) { "native relay connect failed" }
        RelayClientSession(driver, connection)
    }
}

/** Native JNI owner: stream → session → connection; no native handle escapes. */
internal object NativeRelayClientDriver : RelayClientDriver {
    private data class Owned(val session: Long, val channel: Long)
    private val owned = mutableMapOf<Long, Owned>()

    override fun connect(endpoint: RelayEndpoint, instanceID: String, token: String?, record: ByteArray?, deviceID: String?, tls: RelayTLS): Long {
        // The CONNECT greeting cannot carry a bearer token; never discard it.
        require(token.isNullOrEmpty()) { "CONNECT bearer token is not supported" }
        val connection = PigeonNative.connectionOpen(endpoint.host, endpoint.port, instanceID,
            0, null, tls.verifyPeer, tls.caCertFile, tls.timeoutMs)
        check(connection != 0L) { "native relay connection failed" }
        var channel = 0L
        try {
            val session = if (record == null) {
                channel = PigeonNative.channelInitPairing()
                PigeonNative.connectionSessionInit(connection, channel, null, null)
            } else {
                PigeonNative.connectionActivate(connection, requireNotNull(deviceID), record, null, null)
            }
            check(session != 0L) { "native relay session failed" }
            synchronized(owned) { owned[connection] = Owned(session, channel) }
            return connection
        } catch (e: Throwable) {
            try { if (channel != 0L) PigeonNative.channelFree(channel) } finally {
                PigeonNative.connectionClose(connection)
            }
            throw e
        }
    }

    private fun session(connection: Long): Long = synchronized(owned) {
        owned[connection]?.session ?: error("connection closed")
    }
    override fun primaryStream(connection: Long): Long = PigeonNative.sessionPrimary(session(connection))
    override fun openStream(connection: Long, name: String): Long = PigeonNative.sessionOpenStream(session(connection), name)
    override fun send(stream: Long, data: ByteArray) = PigeonNative.streamSend(stream, data)
    override fun recv(stream: Long): ByteArray = PigeonNative.streamRecv(stream)
    override fun closeStream(stream: Long) {
        try { PigeonNative.streamClose(stream) } finally { PigeonNative.streamFree(stream) }
    }
    override fun closeConnection(connection: Long) {
        val owner = synchronized(owned) { owned.remove(connection) } ?: return
        try { PigeonNative.sessionFree(owner.session) } finally {
            try { if (owner.channel != 0L) PigeonNative.channelFree(owner.channel) } finally {
                PigeonNative.connectionClose(connection)
            }
        }
    }
}

/**
 * Begin a pairing-mode connection. Native initiation binds the plaintext
 * primary stream and sends the arrival marker; [pairInitiator] performs the
 * ceremony on a separate named "ceremony" stream. Use [connectPaired] on
 * a fresh connection after completing the ceremony and persisting its record.
 * CONNECT does not carry a bearer token on the current relay wire; a
 * nonempty [token] is rejected, never silently discarded.
 *
 * Native transport pump (T66.5) and recv-first device verification are
 * prerequisites before this entry point can be considered runtime-ready.
 */
suspend fun connect(
    relayURL: String, instanceID: String, token: String? = null, tls: RelayTLS = RelayTLS(),
): RelayClientSession = connectRelayClient(NativeRelayClientDriver, relayURL, instanceID, token, tls = tls)

/** Establish an encrypted session from a completed pairing record. */
suspend fun connectPaired(
    relayURL: String,
    instanceID: String,
    deviceID: String,
    pairing: PairingRecord,
    tls: RelayTLS = RelayTLS(),
): RelayClientSession = connectRelayClient(
    NativeRelayClientDriver, relayURL, instanceID, pairing = pairing, deviceID = deviceID, tls = tls,
)

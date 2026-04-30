// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package com.marcelocantos.pigeon.relay

import com.marcelocantos.pigeon.crypto.E2EChannel
import java.io.InputStream
import java.io.OutputStream
import java.net.HttpURLConnection
import java.net.URL
import java.nio.ByteBuffer

/**
 * Transport abstraction for a QUIC-like connection.
 * Applications provide an implementation backed by their QUIC library
 * (Cronet, quiche, etc.).
 *
 * Legacy single-channel wire. New code should target the multi-stream
 * peer API in [com.marcelocantos.pigeon.peer.Session] (T30), which
 * delegates to the C peer library via JNI. This interface and
 * [PigeonConn] remain for backwards-compat with the pre-T22 wire
 * protocol and will be removed once all consumers have ported.
 */
@Deprecated("Use com.marcelocantos.pigeon.peer.Session for the multi-stream wire (T30).")
interface QuicTransport {
    /** The bidirectional stream for reliable messages. */
    val inputStream: InputStream
    val outputStream: OutputStream

    /** Send an unreliable datagram. */
    fun sendDatagram(data: ByteArray)

    /** Receive the next datagram (blocking). */
    fun receiveDatagram(): ByteArray

    /** Close the transport. */
    fun close()
}

// Internal message types -- first byte of encrypted plaintext.
// These match the Go constants in conn.go.
private const val MSG_APP: Byte = 0x00
private const val MAX_MESSAGE_SIZE = 1_048_576 // 1 MiB

/**
 * Connection to a pigeon relay. Wraps a [QuicTransport] with the pigeon
 * protocol: length-prefixed framing, handshake, and optional E2E encryption.
 *
 * Mirrors the Go `pigeon.Conn` type. Create via [register] or [connect].
 *
 * Legacy single-channel wire — superseded by
 * [com.marcelocantos.pigeon.peer.Session] (T30). Retained for
 * backwards-compat with the pre-T22 wire protocol.
 */
@Deprecated("Use com.marcelocantos.pigeon.peer.Session for the multi-stream wire (T30).")
class PigeonConn internal constructor(
    private val transport: QuicTransport,
    /** The relay-assigned instance ID. */
    val instanceID: String,
) {
    private var channel: E2EChannel? = null
    private var dgChannel: E2EChannel? = null
    private val writeLock = Any()

    /** Set E2E encryption for stream messages. */
    fun setChannel(ch: E2EChannel) {
        channel = ch
    }

    /** Set E2E encryption for datagrams. */
    fun setDatagramChannel(ch: E2EChannel) {
        dgChannel = ch
    }

    /**
     * Send a message on the primary stream. Thread-safe.
     * In raw mode, data is sent as-is with length-prefix framing.
     * In encrypted mode, data is framed with a message-type byte and encrypted.
     */
    fun send(data: ByteArray) {
        val ch = channel
        val payload = if (ch != null) {
            val framed = ByteArray(1 + data.size)
            framed[0] = MSG_APP
            System.arraycopy(data, 0, framed, 1, data.size)
            ch.encrypt(framed)
        } else {
            data
        }

        synchronized(writeLock) {
            writeMessage(transport.outputStream, payload)
        }
    }

    /**
     * Receive a message from the primary stream.
     * In raw mode, returns the raw bytes. In encrypted mode, decrypts and
     * returns the application payload, silently discarding control messages.
     */
    fun recv(): ByteArray {
        while (true) {
            val data = readMessage(transport.inputStream)
            val ch = channel ?: return data

            val plaintext = ch.decrypt(data)
            when (plaintext[0]) {
                MSG_APP -> return plaintext.copyOfRange(1, plaintext.size)
                else -> continue // control message, discard
            }
        }
    }

    /** Send an unreliable datagram. */
    fun sendDatagram(data: ByteArray) {
        val payload = dgChannel?.encrypt(data) ?: data
        transport.sendDatagram(payload)
    }

    /** Receive the next datagram. */
    fun receiveDatagram(): ByteArray {
        val data = transport.receiveDatagram()
        return dgChannel?.decrypt(data) ?: data
    }

    /** Close the connection. */
    fun close() {
        transport.close()
    }
}

/**
 * Register as a backend with the relay.
 *
 * Sends the register handshake over [transport] and reads the
 * relay-assigned instance ID.
 *
 * @param transport a QUIC transport connected to the relay with ALPN "pigeon"
 * @param token optional bearer token for authentication
 * @return a [PigeonConn] ready for bidirectional messaging
 */
fun register(transport: QuicTransport, token: String? = null, host: String? = null): PigeonConn {
    if (host != null) wakeRelay(host)
    val handshake = if (token != null) "register:$token" else "register"
    writeMessage(transport.outputStream, handshake.toByteArray())
    val id = readMessage(transport.inputStream)
    return PigeonConn(transport, String(id))
}

/**
 * Connect as a client to a specific backend instance.
 *
 * Sends the connect handshake over [transport].
 *
 * @param transport a QUIC transport connected to the relay with ALPN "pigeon"
 * @param instanceID the relay-assigned instance ID of the target backend
 * @return a [PigeonConn] ready for bidirectional messaging
 */
fun connect(transport: QuicTransport, instanceID: String, host: String? = null): PigeonConn {
    if (host != null) wakeRelay(host)
    writeMessage(transport.outputStream, "connect:$instanceID".toByteArray())
    return PigeonConn(transport, instanceID)
}

/**
 * Wake a Fly.io relay that may be auto-stopped.
 *
 * Sends an HTTPS request to /health, which triggers Fly's proxy to
 * start the machine. No-op if the relay is already running.
 * Best-effort — exceptions are silently ignored.
 *
 * @param host relay hostname (e.g., "carrier-pigeon.fly.dev")
 * @param port HTTPS port (typically 443)
 */
fun wakeRelay(host: String, port: Int = 443) {
    try {
        val url = URL("https://$host:$port/health")
        val conn = url.openConnection() as HttpURLConnection
        conn.connectTimeout = 10_000
        conn.readTimeout = 10_000
        conn.requestMethod = "GET"
        conn.responseCode // triggers the request
        conn.disconnect()
    } catch (_: Exception) {
        // Best-effort.
    }
}

// ---- Length-prefixed framing (matches Go writeMessage/readMessage) ----

internal fun writeMessage(out: OutputStream, data: ByteArray) {
    require(data.size <= MAX_MESSAGE_SIZE) { "message too large: ${data.size} > $MAX_MESSAGE_SIZE" }
    val header = ByteBuffer.allocate(4).putInt(data.size).array()
    out.write(header)
    out.write(data)
    out.flush()
}

internal fun readMessage(inp: InputStream): ByteArray {
    val header = readExact(inp, 4)
    val length = ByteBuffer.wrap(header).int
    if (length < 0 || length > MAX_MESSAGE_SIZE) {
        throw IllegalStateException("message too large: $length")
    }
    return readExact(inp, length)
}

/** Read exactly [n] bytes or throw. */
private fun readExact(inp: InputStream, n: Int): ByteArray {
    val buf = ByteArray(n)
    var offset = 0
    while (offset < n) {
        val read = inp.read(buf, offset, n - offset)
        if (read < 0) throw IllegalStateException("stream closed")
        offset += read
    }
    return buf
}

// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package com.marcelocantos.tern.relay

import tech.kwik.core.QuicClientConnection
import tech.kwik.core.QuicStream
import java.io.InputStream
import java.io.OutputStream
import java.net.URI
import java.util.concurrent.LinkedBlockingQueue
import java.util.concurrent.TimeUnit

/**
 * A [QuicTransport] backed by kwik (pure Java QUIC client).
 * Used for E2E integration tests against a real Go tern relay server.
 *
 * The connection is configured with:
 * - ALPN "tern" (tern's raw QUIC protocol identifier)
 * - Self-signed certificate trust disabled (for test servers)
 * - Datagram extension enabled (for unreliable datagram tests)
 */
class KwikQuicTransport private constructor(
    private val connection: QuicClientConnection,
    private val stream: QuicStream,
) : QuicTransport {

    override val inputStream: InputStream = stream.getInputStream()
    override val outputStream: OutputStream = stream.getOutputStream()

    private val incomingDatagrams = LinkedBlockingQueue<ByteArray>()

    init {
        connection.setDatagramHandler { data ->
            incomingDatagrams.put(data)
        }
    }

    override fun sendDatagram(data: ByteArray) {
        connection.sendDatagram(data)
    }

    override fun receiveDatagram(): ByteArray {
        return incomingDatagrams.poll(30, TimeUnit.SECONDS)
            ?: throw IllegalStateException("datagram receive timed out after 30s")
    }

    override fun close() {
        try { outputStream.close() } catch (_: Exception) {}
        connection.close()
    }

    companion object {
        /**
         * Connect to a tern QUIC relay server.
         *
         * @param host the relay hostname (e.g. "127.0.0.1")
         * @param port the relay QUIC port
         * @param trustAllCerts if true, disable server certificate validation
         *   (for self-signed test certs). Defaults to true for backward compat.
         * @return a connected [KwikQuicTransport] with an open bidirectional stream
         */
        fun connect(host: String, port: Int, trustAllCerts: Boolean = true): KwikQuicTransport {
            val builder = QuicClientConnection.newBuilder()
                .uri(URI.create("https://$host:$port"))
                .applicationProtocol("tern")
                .enableDatagramExtension()

            if (trustAllCerts) {
                builder.noServerCertificateCheck()
            }

            val connection = builder.build()
            connection.connect()

            val stream = connection.createStream(true)

            return KwikQuicTransport(connection, stream)
        }
    }
}

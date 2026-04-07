// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package com.marcelocantos.pigeon.relay

import tech.kwik.core.QuicClientConnection
import tech.kwik.core.QuicStream
import java.io.InputStream
import java.io.OutputStream
import java.net.URI
import java.util.concurrent.LinkedBlockingQueue
import java.util.concurrent.TimeUnit

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
        fun connect(host: String, port: Int): KwikQuicTransport {
            val connection = QuicClientConnection.newBuilder()
                .uri(URI.create("https://$host:$port"))
                .applicationProtocol("pigeon")
                .noServerCertificateCheck()
                .enableDatagramExtension()
                .build()

            connection.connect()
            val stream = connection.createStream(true)
            return KwikQuicTransport(connection, stream)
        }
    }
}

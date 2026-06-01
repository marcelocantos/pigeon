// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package com.marcelocantos.pigeon.peer

import tech.kwik.core.QuicClientConnection
import tech.kwik.core.QuicConnection
import tech.kwik.core.QuicStream
import tech.kwik.core.log.NullLogger
import tech.kwik.core.server.ApplicationProtocolConnection
import tech.kwik.core.server.ApplicationProtocolConnectionFactory
import tech.kwik.core.server.ServerConnectionConfig
import tech.kwik.core.server.ServerConnector
import java.net.DatagramSocket
import java.net.InetAddress
import java.net.URI
import java.time.Duration
import java.util.concurrent.CompletableFuture
import java.util.concurrent.TimeUnit
import kotlin.test.Test
import kotlin.test.assertContentEquals
import kotlin.test.assertEquals
import kotlin.test.assertNotNull

/**
 * End-to-end test for the T36 JVM-callback transport over a real QUIC
 * connection (Phase 2 — symmetric with the in-memory loopback in
 * [JniTransportLoopbackTest]).
 *
 * Two pigeon [Session]s, one client-side and one server-side, are bridged
 * over a real Kwik QUIC connection via [JniKwikTransport]. libpigeon
 * drives all AEAD / wire framing across the JNI boundary; Kwik carries
 * the raw bytes. This validates the same vtable shape Phase 1 exercised
 * with in-memory loopback against a live multi-stream QUIC stack.
 */
class JniKwikE2ETest {

    /**
     * Spin up a Kwik server on an ephemeral UDP port, register the
     * "pigeon" ALPN, and connect a Kwik client to it. Build server-side
     * and client-side [Session]s over [JniKwikTransport] using the supplied
     * channels. Inbound streams and datagrams are routed into each
     * transport via the explicit hook methods.
     */
    private class Harness(
        val server: ServerConnector,
        val clientConn: QuicClientConnection,
        val serverConn: QuicConnection,
        val clientTransport: JniKwikTransport,
        val serverTransport: JniKwikTransport,
        val clientSession: Session,
        val serverSession: Session,
    ) : AutoCloseable {
        override fun close() {
            try { clientSession.close() } catch (_: Exception) {}
            try { serverSession.close() } catch (_: Exception) {}
            try { clientTransport.close() } catch (_: Exception) {}
            try { serverTransport.close() } catch (_: Exception) {}
            try { server.close() } catch (_: Exception) {}
        }
    }

    private fun startHarness(
        clientChannel: Channel,
        serverChannel: Channel,
        clientIsBackend: Boolean,
        clientTag: Int,
        serverIsBackend: Boolean,
        serverTag: Int,
        datagramChannels: List<DatagramChannelDef> = emptyList(),
    ): Harness {
        // Bind a UDP socket explicitly so we can read the bound port —
        // Kwik's ServerConnector builder doesn't expose getPort() once
        // built, but accepts an externally-provided socket.
        val udp = DatagramSocket(0, InetAddress.getByName("127.0.0.1"))
        val port = udp.localPort

        data class ServerSide(
            val conn: QuicConnection,
            val transport: JniKwikTransport,
            val session: Session,
        )
        val serverReady = CompletableFuture<ServerSide>()
        val factory = object : ApplicationProtocolConnectionFactory {
            override fun maxConcurrentPeerInitiatedBidirectionalStreams(): Int = 128
            override fun maxTotalPeerInitiatedBidirectionalStreams(): Long = 1024L
            override fun maxConcurrentPeerInitiatedUnidirectionalStreams(): Int = 0
            override fun maxTotalPeerInitiatedUnidirectionalStreams(): Long = 0L
            override fun minBidirectionalStreamReceiverBufferSize(): Int = 1024
            override fun maxBidirectionalStreamReceiverBufferSize(): Long = 1024L * 1024L
            override fun enableDatagramExtension(): Boolean = true

            override fun createConnection(
                protocol: String,
                conn: QuicConnection,
            ): ApplicationProtocolConnection {
                val transport = JniKwikTransport(conn)
                // Server side: route inbound streams via the
                // ApplicationProtocolConnection callback (the
                // Kwik-recommended path on the server). Datagrams go
                // through the connection-level handler.
                conn.setDatagramHandler { data -> transport.acceptIncomingDatagram(data) }
                val session = Session.fromTransport(
                    serverChannel, transport,
                    isBackend = serverIsBackend, clientTag = serverTag,
                    datagramChannels = datagramChannels,
                )
                serverReady.complete(ServerSide(conn, transport, session))
                return object : ApplicationProtocolConnection {
                    override fun acceptPeerInitiatedStream(stream: QuicStream) {
                        transport.acceptIncomingStream(stream)
                    }
                }
            }
        }

        val certIn = javaClass.classLoader!!.getResourceAsStream("kwik-test.cert.pem")
            ?: error("missing kwik-test.cert.pem on test classpath")
        val keyIn = javaClass.classLoader!!.getResourceAsStream("kwik-test.key.pem")
            ?: error("missing kwik-test.key.pem on test classpath")

        val serverConfig = ServerConnectionConfig.builder()
            .maxIdleTimeoutInSeconds(30)
            .maxConnectionBufferSize(1024L * 1024L)
            .maxBidirectionalStreamBufferSize(1024L * 1024L)
            .maxOpenPeerInitiatedBidirectionalStreams(128)
            .maxTotalPeerInitiatedBidirectionalStreams(1024L)
            .build()

        val server = ServerConnector.builder()
            .withSocket(udp)
            .withCertificate(certIn, keyIn)
            .withSupportedVersion(QuicConnection.QuicVersion.V1)
            .withConfiguration(serverConfig)
            .withLogger(NullLogger())
            .build()
        server.registerApplicationProtocol("pigeon", factory)
        server.start()

        val clientConn = QuicClientConnection.newBuilder()
            .uri(URI.create("https://127.0.0.1:$port"))
            .applicationProtocol("pigeon")
            .noServerCertificateCheck()
            .enableDatagramExtension()
            .maxOpenPeerInitiatedBidirectionalStreams(128)
            .connectTimeout(Duration.ofSeconds(10))
            .build()
        clientConn.connect()

        val clientTransport = JniKwikTransport(clientConn)
        clientConn.setPeerInitiatedStreamCallback { stream ->
            clientTransport.acceptIncomingStream(stream)
        }
        clientConn.setDatagramHandler { data ->
            clientTransport.acceptIncomingDatagram(data)
        }

        val clientSession = Session.fromTransport(
            clientChannel, clientTransport,
            isBackend = clientIsBackend, clientTag = clientTag,
            datagramChannels = datagramChannels,
        )

        val s = serverReady.get(10, TimeUnit.SECONDS)
        return Harness(
            server = server,
            clientConn = clientConn,
            serverConn = s.conn,
            clientTransport = clientTransport,
            serverTransport = s.transport,
            clientSession = clientSession,
            serverSession = s.session,
        )
    }

    @Test
    fun streamRoundTripChat() {
        val key = ByteArray(32) { (it + 11).toByte() }
        val chClient = Channel.shared(key)
        val chServer = Channel.shared(key)
        val h = startHarness(
            clientChannel = chClient,
            serverChannel = chServer,
            clientIsBackend = true, clientTag = 0xabcd1234.toInt(),
            serverIsBackend = false, serverTag = 0,
        )
        h.use {
            val cChat = h.clientSession.openStreamBlocking("chat")
            val sChat = h.serverSession.acceptStreamBlocking()
            assertNotNull(sChat, "server should accept the new stream")
            assertEquals("chat", sChat.name)

            cChat.sendBlocking("hello".toByteArray())
            assertContentEquals("hello".toByteArray(), sChat.recvBlocking())

            sChat.sendBlocking("world".toByteArray())
            assertContentEquals("world".toByteArray(), cChat.recvBlocking())

            cChat.close()
            sChat.close()
        }
    }

    @Test
    fun multipleStreamsRoundTrip() {
        val key = ByteArray(32) { (it - 9).toByte() }
        val chClient = Channel.shared(key)
        val chServer = Channel.shared(key)
        val h = startHarness(
            clientChannel = chClient,
            serverChannel = chServer,
            clientIsBackend = true, clientTag = 0x55667788,
            serverIsBackend = false, serverTag = 0,
        )
        h.use {
            val names = listOf("control", "data", "diagnostics")
            val pairs = names.map { name ->
                val co = h.clientSession.openStreamBlocking(name)
                val si = h.serverSession.acceptStreamBlocking()
                assertNotNull(si, "server should accept $name")
                assertEquals(name, si.name)
                co to si
            }
            for ((co, si) in pairs) {
                val payload = "msg-${si.name}".toByteArray()
                co.sendBlocking(payload)
                assertContentEquals(payload, si.recvBlocking())
            }
            for ((co, si) in pairs) { co.close(); si.close() }
        }
    }

    @Test
    fun datagramRoundTripPing() {
        val key = ByteArray(32) { (it * 5 + 2).toByte() }
        val chClient = Channel.shared(key, mode = Channel.MODE_DATAGRAMS)
        val chServer = Channel.shared(key, mode = Channel.MODE_DATAGRAMS)
        val chans = listOf(
            DatagramChannelDef("ping", 1L),
            DatagramChannelDef("metric", 2L),
        )
        val h = startHarness(
            clientChannel = chClient,
            serverChannel = chServer,
            clientIsBackend = false, clientTag = 0,
            serverIsBackend = false, serverTag = 0,
            datagramChannels = chans,
        )
        h.use {
            val cPing = h.clientSession.getDatagram("ping")
            val sPing = h.serverSession.getDatagram("ping")
            try {
                cPing.sendBlocking("p1".toByteArray())
                assertContentEquals("p1".toByteArray(), sPing.recvBlocking())

                sPing.sendBlocking("p2".toByteArray())
                assertContentEquals("p2".toByteArray(), cPing.recvBlocking())
            } finally {
                cPing.close()
                sPing.close()
            }
        }
    }
}

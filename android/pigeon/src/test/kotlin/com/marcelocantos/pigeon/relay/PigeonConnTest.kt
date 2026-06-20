// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package com.marcelocantos.pigeon.relay

import com.marcelocantos.pigeon.crypto.E2EChannel
import java.io.InputStream
import java.io.OutputStream
import java.io.PipedInputStream
import java.io.PipedOutputStream
import java.nio.ByteBuffer
import java.util.concurrent.LinkedBlockingQueue
import kotlin.test.Test
import kotlin.test.assertContentEquals
import kotlin.test.assertEquals

/**
 * A mock [QuicTransport] backed by piped streams and blocking queues
 * for datagrams. Two linked MockTransports simulate a client/server pair.
 */
class MockTransport(
    override val inputStream: InputStream,
    override val outputStream: OutputStream,
    private val outDatagrams: LinkedBlockingQueue<ByteArray>,
    private val inDatagrams: LinkedBlockingQueue<ByteArray>,
) : QuicTransport {
    var closed = false
        private set

    override fun sendDatagram(data: ByteArray) {
        outDatagrams.put(data)
    }

    override fun receiveDatagram(): ByteArray = inDatagrams.take()

    override fun close() {
        closed = true
    }
}

/** Create a linked pair of MockTransports. Writes on one appear as reads on the other. */
fun createTransportPair(): Pair<MockTransport, MockTransport> {
    val aToB = PipedOutputStream()
    val bFromA = PipedInputStream(aToB, 65536)
    val bToA = PipedOutputStream()
    val aFromB = PipedInputStream(bToA, 65536)

    val dgAtoB = LinkedBlockingQueue<ByteArray>()
    val dgBtoA = LinkedBlockingQueue<ByteArray>()

    val a = MockTransport(aFromB, aToB, dgAtoB, dgBtoA)
    val b = MockTransport(bFromA, bToA, dgBtoA, dgAtoB)
    return a to b
}

class PigeonConnTest {
    @Test
    fun `register handshake without token`() {
        val (client, server) = createTransportPair()

        // Client sends register handshake in a background thread.
        val result = Thread {
            register(client)
        }.also { it.start() }

        // Server reads the handshake and responds with an instance ID.
        val handshake = readMessage(server.inputStream)
        assertEquals("register", String(handshake))
        writeMessage(server.outputStream, "test-id-123".toByteArray())

        result.join(5000)

        // Verify via a second register to check the protocol works.
        // (The first register completed in the thread.)
    }

    @Test
    fun `register handshake with token`() {
        val (client, server) = createTransportPair()

        val thread = Thread {
            register(client, token = "secret-token")
        }.also { it.start() }

        val handshake = readMessage(server.inputStream)
        // Post-T45 the register greeting carries both fields once either
        // is non-empty: `register:<token>:<instanceID>` (id empty here).
        assertEquals("register:secret-token:", String(handshake))
        writeMessage(server.outputStream, "id-456".toByteArray())

        thread.join(5000)
    }

    @Test
    fun `listen handshake`() {
        val (client, server) = createTransportPair()

        val thread = Thread {
            listen(client, "target-instance")
        }.also { it.start() }

        val handshake = readMessage(server.inputStream)
        // listen greeting echoes the instance ID in the id field:
        // `listen::<instanceID>` (token empty).
        assertEquals("listen::target-instance", String(handshake))
        // Relay acks with the instance ID before parking.
        writeMessage(server.outputStream, "target-instance".toByteArray())

        thread.join(5000)
    }

    @Test
    fun `connect handshake`() {
        val (client, server) = createTransportPair()

        val thread = Thread {
            connect(client, "target-instance")
        }.also { it.start() }

        val handshake = readMessage(server.inputStream)
        assertEquals("connect:target-instance", String(handshake))
        // Post-T45 connect waits for the relay's "ok" ack (the bridge is
        // live once it returns).
        writeMessage(server.outputStream, "ok".toByteArray())

        thread.join(5000)
    }

    @Test
    fun `send and recv raw messages`() {
        val (clientTransport, serverTransport) = createTransportPair()

        // Simulate register handshake.
        val connThread = Thread {
            val conn = register(clientTransport)
            conn.send("hello".toByteArray())
            val reply = conn.recv()
            assertEquals("world", String(reply))
            conn
        }.also { it.start() }

        // Server side: complete handshake, then exchange messages.
        val handshake = readMessage(serverTransport.inputStream)
        assertEquals("register", String(handshake))
        writeMessage(serverTransport.outputStream, "id-raw".toByteArray())

        // Read the message the client sent.
        val msg = readMessage(serverTransport.inputStream)
        assertEquals("hello", String(msg))

        // Send a reply.
        writeMessage(serverTransport.outputStream, "world".toByteArray())

        connThread.join(5000)
    }

    @Test
    fun `send and recv with encryption`() {
        val (clientTransport, serverTransport) = createTransportPair()

        val sharedKey = ByteArray(32) { it.toByte() }
        val clientChannel = E2EChannel(sharedKey, isServer = false)
        val serverChannel = E2EChannel(sharedKey, isServer = true)

        var receivedPlaintext: ByteArray? = null

        val connThread = Thread {
            val conn = register(clientTransport)
            conn.setChannel(clientChannel)
            conn.send("encrypted-hello".toByteArray())
            receivedPlaintext = conn.recv()
        }.also { it.start() }

        // Server side: complete handshake.
        readMessage(serverTransport.inputStream)
        writeMessage(serverTransport.outputStream, "id-enc".toByteArray())

        // Read the encrypted message the client sent.
        val ciphertext = readMessage(serverTransport.inputStream)

        // Decrypt server-side (server receives client-to-server).
        val plaintext = serverChannel.decrypt(ciphertext)
        assertEquals(0x00, plaintext[0]) // MSG_APP
        assertEquals("encrypted-hello", String(plaintext, 1, plaintext.size - 1))

        // Send an encrypted reply (server-to-client).
        val replyFramed = ByteArray(1 + "encrypted-world".length)
        replyFramed[0] = 0x00
        System.arraycopy("encrypted-world".toByteArray(), 0, replyFramed, 1, "encrypted-world".length)
        val replyCiphertext = serverChannel.encrypt(replyFramed)
        writeMessage(serverTransport.outputStream, replyCiphertext)

        connThread.join(5000)
        assertContentEquals("encrypted-world".toByteArray(), receivedPlaintext)
    }

    @Test
    fun `datagram send and recv raw`() {
        val (clientTransport, serverTransport) = createTransportPair()

        val connThread = Thread {
            val conn = register(clientTransport)
            conn.sendDatagram("dg-hello".toByteArray())
            val reply = conn.receiveDatagram()
            assertEquals("dg-world", String(reply))
        }.also { it.start() }

        // Complete handshake.
        readMessage(serverTransport.inputStream)
        writeMessage(serverTransport.outputStream, "id-dg".toByteArray())

        // Receive datagram from client.
        val dg = serverTransport.receiveDatagram()
        assertEquals("dg-hello", String(dg))

        // Send datagram reply.
        serverTransport.sendDatagram("dg-world".toByteArray())

        connThread.join(5000)
    }

    @Test
    fun `datagram send and recv with encryption`() {
        val (clientTransport, serverTransport) = createTransportPair()

        val sharedKey = ByteArray(32) { (it + 42).toByte() }
        val clientDgChannel = E2EChannel(sharedKey, isServer = false).apply {
            mode = com.marcelocantos.pigeon.crypto.ChannelMode.DATAGRAMS
        }
        val serverDgChannel = E2EChannel(sharedKey, isServer = true).apply {
            mode = com.marcelocantos.pigeon.crypto.ChannelMode.DATAGRAMS
        }

        var receivedDg: ByteArray? = null

        val connThread = Thread {
            val conn = register(clientTransport)
            conn.setDatagramChannel(clientDgChannel)
            conn.sendDatagram("enc-dg".toByteArray())
            receivedDg = conn.receiveDatagram()
        }.also { it.start() }

        // Complete handshake.
        readMessage(serverTransport.inputStream)
        writeMessage(serverTransport.outputStream, "id-edg".toByteArray())

        // Receive encrypted datagram, decrypt server-side.
        val encDg = serverTransport.receiveDatagram()
        val plainDg = serverDgChannel.decrypt(encDg)
        assertEquals("enc-dg", String(plainDg))

        // Send encrypted datagram reply.
        val replyEnc = serverDgChannel.encrypt("enc-dg-reply".toByteArray())
        serverTransport.sendDatagram(replyEnc)

        connThread.join(5000)
        assertContentEquals("enc-dg-reply".toByteArray(), receivedDg)
    }

    @Test
    fun `length-prefix framing round-trip`() {
        val out = PipedOutputStream()
        val inp = PipedInputStream(out, 65536)

        val messages = listOf(
            "".toByteArray(),
            "short".toByteArray(),
            ByteArray(1000) { (it % 256).toByte() },
        )

        for (msg in messages) {
            writeMessage(out, msg)
        }

        for (msg in messages) {
            val read = readMessage(inp)
            assertContentEquals(msg, read)
        }
    }

    @Test
    fun `close delegates to transport`() {
        val (clientTransport, server) = createTransportPair()
        val thread = Thread {
            // Drain the connect greeting and ack so connect() returns.
            readMessage(server.inputStream)
            writeMessage(server.outputStream, "ok".toByteArray())
        }.also { it.start() }
        val conn = connect(clientTransport, "some-id")
        thread.join(5000)
        conn.close()
        assertEquals(true, clientTransport.closed)
    }
}

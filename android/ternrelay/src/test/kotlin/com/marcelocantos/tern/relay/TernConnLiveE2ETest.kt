// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package com.marcelocantos.tern.relay

import com.marcelocantos.tern.crypto.ChannelMode
import com.marcelocantos.tern.crypto.E2EChannel
import com.marcelocantos.tern.crypto.E2EKeyPair
import com.marcelocantos.tern.crypto.deriveConfirmationCode
import kotlin.test.Test
import kotlin.test.assertContentEquals
import kotlin.test.assertEquals
import kotlin.test.assertTrue
import org.junit.jupiter.api.Assumptions

/**
 * End-to-end integration tests against the live tern relay at tern.fly.dev.
 *
 * These tests connect via raw QUIC (port 4433) using kwik with real TLS
 * certificate validation (the live server has a Let's Encrypt cert).
 *
 * Requires environment variables:
 * - `TERN_TOKEN`: bearer token for /register authentication
 * - `TERN_RELAY_HOST` (optional): relay hostname, defaults to "tern.fly.dev"
 *
 * Tests are skipped when `TERN_TOKEN` is not set.
 */
class TernConnLiveE2ETest {

    private val token: String? = System.getenv("TERN_TOKEN")
    private val host: String = System.getenv("TERN_RELAY_HOST") ?: "tern.fly.dev"
    private val quicPort: Int = 4433

    private fun assumeLive() {
        Assumptions.assumeTrue(
            token != null && token.isNotEmpty(),
            "TERN_TOKEN not set; skipping live relay test"
        )
    }

    private fun connectToRelay(): KwikQuicTransport {
        return KwikQuicTransport.connect(host, quicPort, trustAllCerts = false)
    }

    @Test
    fun `live - register with token assigns instance ID`() {
        assumeLive()
        val transport = connectToRelay()
        try {
            val conn = register(transport, token = token!!)
            assertTrue(conn.instanceID.isNotEmpty(), "instance ID should be non-empty")
        } finally {
            transport.close()
        }
    }

    @Test
    fun `live - bidirectional stream round-trip`() {
        assumeLive()
        val backendTransport = connectToRelay()
        val backend = register(backendTransport, token = token!!)
        assertTrue(backend.instanceID.isNotEmpty())

        val clientTransport = connectToRelay()
        val client = connect(clientTransport, backend.instanceID)

        try {
            // Client -> backend.
            client.send("hello from live kotlin client".toByteArray())
            val received = backend.recv()
            assertEquals("hello from live kotlin client", String(received))

            // Backend -> client.
            backend.send("hello from live kotlin backend".toByteArray())
            val reply = client.recv()
            assertEquals("hello from live kotlin backend", String(reply))
        } finally {
            client.close()
            backend.close()
        }
    }

    @Test
    fun `live - encrypted stream round-trip`() {
        assumeLive()
        val backendTransport = connectToRelay()
        val backend = register(backendTransport, token = token!!)

        val clientTransport = connectToRelay()
        val client = connect(clientTransport, backend.instanceID)

        try {
            val sharedKey = ByteArray(32) { (it * 7 + 13).toByte() }

            val clientChannel = E2EChannel(sharedKey, isServer = false)
            val backendChannel = E2EChannel(sharedKey, isServer = true)

            client.setChannel(clientChannel)
            backend.setChannel(backendChannel)

            // Client -> backend (encrypted).
            client.send("live secret from client".toByteArray())
            val received = backend.recv()
            assertEquals("live secret from client", String(received))

            // Backend -> client (encrypted).
            backend.send("live secret from backend".toByteArray())
            val reply = client.recv()
            assertEquals("live secret from backend", String(reply))
        } finally {
            client.close()
            backend.close()
        }
    }

    @Test
    fun `live - datagram round-trip`() {
        assumeLive()
        val backendTransport = connectToRelay()
        val backend = register(backendTransport, token = token!!)

        val clientTransport = connectToRelay()
        val client = connect(clientTransport, backend.instanceID)

        try {
            // Client -> backend datagram.
            client.sendDatagram("live-dg-from-client".toByteArray())
            val dg = backend.receiveDatagram()
            assertEquals("live-dg-from-client", String(dg))

            // Backend -> client datagram.
            backend.sendDatagram("live-dg-from-backend".toByteArray())
            val reply = client.receiveDatagram()
            assertEquals("live-dg-from-backend", String(reply))
        } finally {
            client.close()
            backend.close()
        }
    }

    @Test
    fun `live - encrypted datagram round-trip`() {
        assumeLive()
        val backendTransport = connectToRelay()
        val backend = register(backendTransport, token = token!!)

        val clientTransport = connectToRelay()
        val client = connect(clientTransport, backend.instanceID)

        try {
            val sharedKey = ByteArray(32) { (it * 3 + 5).toByte() }

            val clientDgChannel = E2EChannel(sharedKey, isServer = false).apply {
                mode = ChannelMode.DATAGRAMS
            }
            val backendDgChannel = E2EChannel(sharedKey, isServer = true).apply {
                mode = ChannelMode.DATAGRAMS
            }

            client.setDatagramChannel(clientDgChannel)
            backend.setDatagramChannel(backendDgChannel)

            // Client -> backend encrypted datagram.
            client.sendDatagram("live-encrypted-dg".toByteArray())
            val dg = backend.receiveDatagram()
            assertEquals("live-encrypted-dg", String(dg))

            // Backend -> client encrypted datagram.
            backend.sendDatagram("live-encrypted-dg-reply".toByteArray())
            val reply = client.receiveDatagram()
            assertEquals("live-encrypted-dg-reply", String(reply))
        } finally {
            client.close()
            backend.close()
        }
    }

    @Test
    fun `live - full pairing ceremony with ECDH key exchange`() {
        assumeLive()
        val backendTransport = connectToRelay()
        val backend = register(backendTransport, token = token!!)

        val clientTransport = connectToRelay()
        val client = connect(clientTransport, backend.instanceID)

        try {
            // Both sides generate ECDH key pairs.
            val clientKP = E2EKeyPair()
            val backendKP = E2EKeyPair()

            // Client sends its public key through the relay.
            client.send(clientKP.publicKeyData)
            val clientPubAtBackend = backend.recv()

            // Backend sends its public key through the relay.
            backend.send(backendKP.publicKeyData)
            val backendPubAtClient = client.recv()

            // Verify confirmation codes match (no MitM).
            val clientCode = deriveConfirmationCode(backendPubAtClient, clientKP.publicKeyData)
            val backendCode = deriveConfirmationCode(backendKP.publicKeyData, clientPubAtBackend)
            assertEquals(clientCode, backendCode, "Confirmation codes should match")

            // Derive session keys.
            val clientSendKey = clientKP.deriveSessionKey(backendPubAtClient, "client-to-server".toByteArray())
            val clientRecvKey = clientKP.deriveSessionKey(backendPubAtClient, "server-to-client".toByteArray())
            val backendSendKey = backendKP.deriveSessionKey(clientPubAtBackend, "server-to-client".toByteArray())
            val backendRecvKey = backendKP.deriveSessionKey(clientPubAtBackend, "client-to-server".toByteArray())

            // Create encrypted channels.
            val clientChannel = E2EChannel(clientSendKey, clientRecvKey)
            val backendChannel = E2EChannel(backendSendKey, backendRecvKey)

            client.setChannel(clientChannel)
            backend.setChannel(backendChannel)

            // Exchange encrypted messages.
            val secret = "live top secret pairing data"
            client.send(secret.toByteArray())
            val decrypted = backend.recv()
            assertEquals(secret, String(decrypted))

            backend.send("live acknowledged".toByteArray())
            val ack = client.recv()
            assertEquals("live acknowledged", String(ack))
        } finally {
            client.close()
            backend.close()
        }
    }
}

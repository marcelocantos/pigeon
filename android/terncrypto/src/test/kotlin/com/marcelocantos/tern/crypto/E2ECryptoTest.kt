// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package com.marcelocantos.tern.crypto

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFailsWith
import kotlin.test.assertNotEquals

class E2ECryptoTest {

    @Test
    fun `key exchange produces 32 byte public key`() {
        val kp = E2EKeyPair()
        assertEquals(32, kp.publicKeyData.size)
    }

    @Test
    fun `two key pairs derive same session key`() {
        val alice = E2EKeyPair()
        val bob = E2EKeyPair()

        val info = "test-info".toByteArray()
        val aliceKey = alice.deriveSessionKey(bob.publicKeyData, info)
        val bobKey = bob.deriveSessionKey(alice.publicKeyData, info)

        assertEquals(32, aliceKey.size)
        assertContentEquals(aliceKey, bobKey)
    }

    @Test
    fun `symmetric channel encrypt decrypt round trip`() {
        val sharedKey = ByteArray(32) { it.toByte() }
        val server = E2EChannel(sharedKey, isServer = true)
        val client = E2EChannel(sharedKey, isServer = false)

        val plaintext = "hello from client".toByteArray()
        val encrypted = client.encrypt(plaintext)
        val decrypted = server.decrypt(encrypted)

        assertContentEquals(plaintext, decrypted)
    }

    @Test
    fun `bidirectional channel communication`() {
        val sharedKey = ByteArray(32) { it.toByte() }
        val server = E2EChannel(sharedKey, isServer = true)
        val client = E2EChannel(sharedKey, isServer = false)

        // Client → server
        val msg1 = "client message".toByteArray()
        assertContentEquals(msg1, server.decrypt(client.encrypt(msg1)))

        // Server → client
        val msg2 = "server message".toByteArray()
        assertContentEquals(msg2, client.decrypt(server.encrypt(msg2)))

        // Multiple messages in each direction
        val msg3 = "second client msg".toByteArray()
        assertContentEquals(msg3, server.decrypt(client.encrypt(msg3)))
    }

    @Test
    fun `replay attack rejected`() {
        val sharedKey = ByteArray(32) { it.toByte() }
        val server = E2EChannel(sharedKey, isServer = true)
        val client = E2EChannel(sharedKey, isServer = false)

        val encrypted = client.encrypt("hello".toByteArray())
        server.decrypt(encrypted) // first time OK

        // Replay same message → wrong sequence number
        assertFailsWith<E2EException> {
            server.decrypt(encrypted)
        }
    }

    @Test
    fun `ciphertext too short rejected`() {
        val sharedKey = ByteArray(32) { it.toByte() }
        val channel = E2EChannel(sharedKey, isServer = true)

        assertFailsWith<E2EException> {
            channel.decrypt(ByteArray(10)) // less than 8 seq + 16 tag
        }
    }

    @Test
    fun `different shared keys produce different ciphertexts`() {
        val key1 = ByteArray(32) { 1 }
        val key2 = ByteArray(32) { 2 }
        val ch1 = E2EChannel(key1, isServer = false)
        val ch2 = E2EChannel(key2, isServer = false)

        val plaintext = "same message".toByteArray()
        val enc1 = ch1.encrypt(plaintext)
        val enc2 = ch2.encrypt(plaintext)

        // Ciphertexts should differ (different keys)
        assertNotEquals(enc1.toList(), enc2.toList())
    }

    @Test
    fun `full ECDH pairing flow`() {
        // Simulate the full pairing ceremony crypto flow:
        // 1. Both sides generate key pairs
        val serverKp = E2EKeyPair()
        val clientKp = E2EKeyPair()

        // 2. Exchange public keys and derive session keys
        val serverSendKey = serverKp.deriveSessionKey(clientKp.publicKeyData, "server-to-client".toByteArray())
        val serverRecvKey = serverKp.deriveSessionKey(clientKp.publicKeyData, "client-to-server".toByteArray())
        val clientSendKey = clientKp.deriveSessionKey(serverKp.publicKeyData, "client-to-server".toByteArray())
        val clientRecvKey = clientKp.deriveSessionKey(serverKp.publicKeyData, "server-to-client".toByteArray())

        // Keys should match cross-directionally
        assertContentEquals(serverSendKey, clientRecvKey)
        assertContentEquals(serverRecvKey, clientSendKey)

        // 3. Create channels and exchange encrypted messages
        val serverCh = E2EChannel(serverSendKey, serverRecvKey)
        val clientCh = E2EChannel(clientSendKey, clientRecvKey)

        val msg = "encrypted pairing message".toByteArray()
        assertContentEquals(msg, clientCh.decrypt(serverCh.encrypt(msg)))
        assertContentEquals(msg, serverCh.decrypt(clientCh.encrypt(msg)))
    }

    @Test
    fun `confirmation code crossplatform vector`() {
        // Fixed X25519 public keys (any 32-byte value works for derivation).
        val keyA = ByteArray(32) { 0x01 }
        val keyB = ByteArray(32) { 0x02 }

        // This expected value must match Go, TypeScript, and Swift tests.
        val expected = "629624"
        assertEquals(expected, deriveConfirmationCode(keyA, keyB))

        // Verify order-independence.
        assertEquals(expected, deriveConfirmationCode(keyB, keyA))
    }

    private fun assertContentEquals(expected: ByteArray, actual: ByteArray) {
        assertEquals(expected.toList(), actual.toList())
    }
}

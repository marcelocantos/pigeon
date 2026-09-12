// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package com.marcelocantos.pigeon.crypto

import java.io.File
import kotlin.test.Test
import kotlin.test.assertEquals

/**
 * Loads c/test/vectors.json and checks the Kotlin AEAD/HKDF stack
 * against the Go-generated vectors (ENT-003). Uses PairingRecord /
 * E2EChannel — the shipped JCA path — as the oracle.
 */
class CrossLanguageVectorTest {

    @Test
    fun `confirmation code matches Go vector`() {
        val v = loadVectors()
        val code = deriveConfirmationCode(v.hex("alice_public"), v.hex("bob_public"))
        assertEquals(v.str("confirmation_code"), code)
    }

    @Test
    fun `ECDH ciphertext decrypts and re-encrypts`() {
        val v = loadVectors()
        val alice = PairingRecord(
            peerInstanceID = "bob",
            relayURL = "test",
            localPrivateKey = v.hex("alice_private"),
            localPublicKey = v.hex("alice_public"),
            peerPublicKey = v.hex("bob_public"),
        )
        val bob = PairingRecord(
            peerInstanceID = "alice",
            relayURL = "test",
            localPrivateKey = v.hex("bob_private"),
            localPublicKey = v.hex("bob_public"),
            peerPublicKey = v.hex("alice_public"),
        )
        val aliceCh = alice.deriveChannel(
            "alice-to-bob".toByteArray(),
            "bob-to-alice".toByteArray(),
        )
        val bobCh = bob.deriveChannel(
            "bob-to-alice".toByteArray(),
            "alice-to-bob".toByteArray(),
        )
        val pt = v.str("plaintext").toByteArray()
        assertContentEquals(pt, bobCh.decrypt(v.hex("ciphertext")))
        assertContentEquals(v.hex("ciphertext"), aliceCh.encrypt(pt))
    }

    @Test
    fun `symmetric ciphertexts match Go vector`() {
        val v = loadVectors()
        val master = v.hex("symmetric_master")
        val client = E2EChannel(master, isServer = false)
        val server = E2EChannel(master, isServer = true)
        val pt = v.str("symmetric_plaintext").toByteArray()
        assertContentEquals(pt, server.decrypt(v.hex("symmetric_ciphertext_c2s")))
        assertContentEquals(pt, client.decrypt(v.hex("symmetric_ciphertext_s2c")))
        assertContentEquals(v.hex("symmetric_ciphertext_c2s"), client.encrypt(pt))
        assertContentEquals(v.hex("symmetric_ciphertext_s2c"), server.encrypt(pt))
    }

    private class Vectors(private val obj: Map<String, String>) {
        fun str(k: String): String = obj.getValue(k)
        fun hex(k: String): ByteArray = hexDecode(obj.getValue(k))
    }

    private fun loadVectors(): Vectors {
        val candidates = listOf(
            File("c/test/vectors.json"),
            File("../../c/test/vectors.json"),
            File("../../../c/test/vectors.json"),
            File("../../../../c/test/vectors.json"),
        )
        val file = candidates.firstOrNull { it.isFile }
            ?: error("c/test/vectors.json not found from ${File(".").absolutePath}")
        val map = mutableMapOf<String, String>()
        Regex(""""([^"]+)"\s*:\s*"([^"]*)"""").findAll(file.readText()).forEach {
            map[it.groupValues[1]] = it.groupValues[2]
        }
        require(map.isNotEmpty()) { "${file.path} parsed to no keys" }
        return Vectors(map)
    }

    private fun hexDecode(s: String): ByteArray {
        require(s.length % 2 == 0) { "odd hex length" }
        return ByteArray(s.length / 2) { i ->
            s.substring(i * 2, i * 2 + 2).toInt(16).toByte()
        }
    }

    private fun assertContentEquals(expected: ByteArray, actual: ByteArray) {
        assertEquals(expected.toList(), actual.toList())
    }
}

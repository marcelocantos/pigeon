// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package com.marcelocantos.pigeon.jni

import kotlin.test.Test
import kotlin.test.assertContentEquals
import kotlin.test.assertEquals
import kotlin.test.assertNotNull
import kotlin.test.assertTrue

/**
 * Low-level tests for [PigeonNative] — exercise the JNI surface
 * directly, no Kotlin idiomatic wrappers.
 *
 * Pinned wire-byte vectors mirror c/test/test_pigeon.c::test_stream_header
 * so any drift between implementations breaks here too. Three independent
 * implementations (Go native, C, Kotlin-via-C) all agree on the same
 * bytes by construction.
 */
class PigeonNativeTest {

    @Test
    fun keypairProducesDistinctKeys() {
        val a = PigeonNative.keypair()
        val b = PigeonNative.keypair()
        assertEquals(64, a.size)
        assertEquals(64, b.size)
        // Vanishingly unlikely to collide.
        assertTrue(!a.contentEquals(b), "two fresh keypairs should differ")
    }

    @Test
    fun deriveSessionKeySymmetric() {
        // X25519 ECDH is symmetric: derive(privA, pubB) == derive(privB, pubA).
        val kpA = PigeonNative.keypair()
        val kpB = PigeonNative.keypair()
        val privA = kpA.copyOfRange(0, 32); val pubA = kpA.copyOfRange(32, 64)
        val privB = kpB.copyOfRange(0, 32); val pubB = kpB.copyOfRange(32, 64)

        val info = "pigeon-test".toByteArray()
        val ka = PigeonNative.deriveSessionKey(privA, pubB, info)
        val kb = PigeonNative.deriveSessionKey(privB, pubA, info)
        assertEquals(32, ka.size)
        assertContentEquals(ka, kb, "ECDH should be symmetric")
    }

    @Test
    fun confirmationCodeOrderIndependent() {
        val kpA = PigeonNative.keypair()
        val kpB = PigeonNative.keypair()
        val pubA = kpA.copyOfRange(32, 64)
        val pubB = kpB.copyOfRange(32, 64)

        val ab = PigeonNative.deriveConfirmationCode(pubA, pubB)
        val ba = PigeonNative.deriveConfirmationCode(pubB, pubA)
        assertEquals(6, ab.length)
        assertEquals(ab, ba, "confirmation code should be order-independent")
        assertTrue(ab.all { it.isDigit() }, "confirmation code should be 6 digits")
    }

    @Test
    fun streamHeaderEmptyNamePrimary() {
        // Pinned vector: client primary is `[0x00]` (single varint zero).
        val hdr = PigeonNative.encodeStreamHeader(false, 0, "")
        assertContentEquals(byteArrayOf(0x00), hdr,
            "empty-name primary header must be a single zero byte")
    }

    @Test
    fun streamHeaderClientControl() {
        // Pinned vector: "control" client header.
        val hdr = PigeonNative.encodeStreamHeader(false, 0, "control")
        assertContentEquals(
            byteArrayOf(0x07, 'c'.code.toByte(), 'o'.code.toByte(), 'n'.code.toByte(),
                't'.code.toByte(), 'r'.code.toByte(), 'o'.code.toByte(), 'l'.code.toByte()),
            hdr,
        )
    }

    @Test
    fun streamHeaderBackendChat() {
        // Pinned vector: backend "chat" stream with tag 0x01020304.
        val hdr = PigeonNative.encodeStreamHeader(true, 0x01020304, "chat")
        assertContentEquals(
            byteArrayOf(
                0x01, 0x02, 0x03, 0x04,         // tag (BE)
                0x04,                            // varint(4)
                'c'.code.toByte(), 'h'.code.toByte(),
                'a'.code.toByte(), 't'.code.toByte(),
            ),
            hdr,
        )
    }

    @Test
    fun uvarintEncodesGoFormat() {
        assertContentEquals(byteArrayOf(0x00), PigeonNative.uvarintEncode(0))
        assertContentEquals(byteArrayOf(0x7f), PigeonNative.uvarintEncode(127))
        // Two bytes for 128.
        val v128 = PigeonNative.uvarintEncode(128)
        assertEquals(2, v128.size)
        assertEquals(0x80.toByte(), v128[0])
        assertEquals(0x01.toByte(), v128[1])
    }

    @Test
    fun channelEncryptDecryptRoundTrip() {
        val key = ByteArray(32) { it.toByte() }
        val ch = PigeonNative.channelInit(key, key, 0)  // STRICT
        try {
            val plain = "the carrier pigeon flies at midnight".toByteArray()
            val ct = PigeonNative.channelEncrypt(ch, plain)
            assertTrue(ct.size > plain.size, "ciphertext should grow vs plaintext")

            // Symmetric decrypt with a fresh channel that has the same key
            // (decrypt advances the recv counter, so we use a separate one).
            val ch2 = PigeonNative.channelInit(key, key, 0)
            try {
                val pt = PigeonNative.channelDecrypt(ch2, ct)
                assertContentEquals(plain, pt)
            } finally {
                PigeonNative.channelFree(ch2)
            }
        } finally {
            PigeonNative.channelFree(ch)
        }
    }
}

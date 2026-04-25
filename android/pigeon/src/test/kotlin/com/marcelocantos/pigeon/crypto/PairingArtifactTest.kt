// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package com.marcelocantos.pigeon.crypto

import org.junit.jupiter.api.Test
import org.junit.jupiter.api.assertThrows
import org.junit.jupiter.api.io.TempDir
import java.io.File
import java.nio.file.Path
import java.time.Duration
import java.time.Instant
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertNotNull
import kotlin.test.assertNull
import kotlin.test.assertTrue

class PairingArtifactTest {

    private fun makeArtifact(ttl: Duration = DEFAULT_PAIRING_TTL, issued: Instant = Instant.parse("2026-04-25T12:00:00Z")): Pair<PairingArtifact, Instant> {
        val kp = E2EKeyPair()
        val peerKP = E2EKeyPair()
        val record = PairingRecord(
            peerInstanceID = "inst-abc",
            relayURL = "https://relay.example.com",
            localKeyPair = kp,
            peerPublicKey = peerKP.publicKeyData,
        )
        return PairingArtifact.mint(record, "tok-xyz", issued, ttl) to issued
    }

    @Test
    fun jsonRoundTrip() {
        val (a, issued) = makeArtifact()
        val json = a.toJson()
        val restored = PairingArtifact.fromJson(json)
        assertEquals("tok-xyz", restored.token)
        assertEquals(issued, restored.issuedAt)
        assertEquals(issued.plus(DEFAULT_PAIRING_TTL), restored.expiresAt)
        assertEquals("inst-abc", restored.record.peerInstanceID)
        assertTrue(restored.record.peerPublicKey.contentEquals(a.record.peerPublicKey))
    }

    @Test
    fun textRoundTrip() {
        val (a, _) = makeArtifact()
        val text = a.toText()
        assertFalse(text.contains('+'))
        assertFalse(text.contains('/'))
        assertFalse(text.contains('='))
        val restored = PairingArtifact.fromText(text)
        assertEquals("inst-abc", restored.record.peerInstanceID)
        assertEquals("tok-xyz", restored.token)
    }

    @Test
    fun isExpired() {
        val (a, issued) = makeArtifact(ttl = Duration.ofHours(24))
        assertFalse(a.isExpired(issued))
        assertFalse(a.isExpired(issued.plus(Duration.ofHours(1))))
        assertTrue(a.isExpired(issued.plus(Duration.ofHours(24))))
        assertTrue(a.isExpired(issued.plus(Duration.ofHours(48))))
    }

    @Test
    fun zeroTtlNeverExpires() {
        val (a, _) = makeArtifact(ttl = Duration.ZERO)
        assertNull(a.expiresAt)
        assertFalse(a.isExpired(Instant.now().plus(Duration.ofDays(36500))))
    }

    @Test
    fun wireFormatIsSnakeCase() {
        val (a, _) = makeArtifact()
        val json = a.toJson()
        assertTrue(json.contains("\"peer_instance_id\""))
        assertTrue(json.contains("\"relay_url\""))
        assertTrue(json.contains("\"local_private_key\""))
        assertTrue(json.contains("\"local_public_key\""))
        assertTrue(json.contains("\"peer_public_key\""))
        assertTrue(json.contains("\"issued_at\""))
        assertTrue(json.contains("\"expires_at\""))
    }

    @Test
    fun fileCredentialStoreRoundTrip(@TempDir dir: Path) {
        val store = FileCredentialStore(File(dir.toFile(), "nested/artifact.json"))
        assertThrows<NoCredentialException> { store.load() }
        assertThrows<NoCredentialException> { store.isExpired() }

        val (a, _) = makeArtifact()
        store.save(a)
        val restored = store.load()
        assertEquals("tok-xyz", restored.token)
        assertEquals("inst-abc", restored.record.peerInstanceID)

        store.delete()
        assertFalse(store.path.exists())
        store.delete() // no-op when absent
    }

    @Test
    fun fileCredentialStoreIsExpired(@TempDir dir: Path) {
        val store = FileCredentialStore(File(dir.toFile(), "artifact.json"))
        val (a, issued) = makeArtifact(ttl = Duration.ofHours(24))
        store.save(a)

        store.now = { issued.plus(Duration.ofHours(1)) }
        assertFalse(store.isExpired())

        store.now = { issued.plus(Duration.ofHours(48)) }
        assertTrue(store.isExpired())
    }

    @Test
    fun connectWithArtifactRejectsExpired() {
        val kp = E2EKeyPair()
        val peerKP = E2EKeyPair()
        val record = PairingRecord(
            peerInstanceID = "inst",
            relayURL = "https://relay.example.com",
            localKeyPair = kp,
            peerPublicKey = peerKP.publicKeyData,
        )
        val stale = PairingArtifact.mint(
            record,
            "",
            issuedAt = Instant.now().minus(Duration.ofDays(31)),
            ttl = DEFAULT_PAIRING_TTL,
        )
        assertTrue(stale.isExpired())

        // Sentinel transport — should never be touched because expiry
        // check fires first.
        val transport = object : com.marcelocantos.pigeon.relay.QuicTransport {
            override val inputStream get() = error("unreachable")
            override val outputStream get() = error("unreachable")
            override fun sendDatagram(data: ByteArray) = error("unreachable")
            override fun receiveDatagram() = error("unreachable")
            override fun close() = Unit
        }
        assertThrows<PairingExpiredException> {
            connectWithArtifact(transport, stale)
        }
    }

    @Test
    fun decodesGoMintedJson() {
        // Realistic JSON shape produced by the Go SDK / pigeon-pair CLI.
        // Verifies cross-language wire compatibility.
        val sample = """
        {
          "record": {
            "peer_instance_id": "device-42",
            "relay_url": "https://relay.example.com",
            "local_private_key": "AAECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh8=",
            "local_public_key":  "ICEiIyQlJicoKSorLC0uLzAxMjM0NTY3ODk6Ozw9Pj8=",
            "peer_public_key":   "QEFCQ0RFRkdISUpLTE1OT1BRUlNUVVZXWFlaW1xdXl8="
          },
          "token": "tok-deploy",
          "issued_at": "2026-04-25T12:00:00Z",
          "expires_at": "2026-05-25T12:00:00Z"
        }
        """.trimIndent()
        val a = PairingArtifact.fromJson(sample)
        assertEquals("device-42", a.record.peerInstanceID)
        assertEquals("tok-deploy", a.token)
        assertNotNull(a.expiresAt)
        assertEquals(Instant.parse("2026-05-25T12:00:00Z"), a.expiresAt)
        assertEquals(32, a.record.localPrivateKey.size)
        assertEquals(32, a.record.localPublicKey.size)
        assertEquals(32, a.record.peerPublicKey.size)
    }
}

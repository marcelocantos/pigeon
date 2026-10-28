// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0
package com.marcelocantos.pigeon

import androidx.test.ext.junit.runners.AndroidJUnit4
import androidx.test.platform.app.InstrumentationRegistry
import com.marcelocantos.pigeon.crypto.E2EKeyPair
import com.marcelocantos.pigeon.peer.RelayTLS
import com.marcelocantos.pigeon.peer.connectPaired
import com.marcelocantos.pigeon.peer.pairInitiator
import java.net.HttpURLConnection
import java.net.URL
import java.security.MessageDigest
import java.util.Base64
import kotlinx.coroutines.runBlocking
import org.junit.Assert.assertEquals
import org.junit.Test
import org.junit.runner.RunWith

/** A physical-device oracle: fresh Go/Kotlin SAS, then a new encrypted session.
 * No JVM loopback, mock transport, or success-shaped skip is permitted. */
@RunWith(AndroidJUnit4::class)
class DeviceEchoTest {
    private fun control(url: String, method: String, body: String? = null): Int {
        val connection = URL(url).openConnection() as HttpURLConnection
        try {
            connection.requestMethod = method
            connection.connectTimeout = 5_000
            connection.readTimeout = 45_000
            if (body != null) {
                connection.doOutput = true
                connection.outputStream.use { it.write(body.toByteArray(Charsets.UTF_8)) }
            }
            return connection.responseCode
        } finally { connection.disconnect() }
    }

    @Test fun freshPairingThenEncryptedChat() = runBlocking {
        val args = InstrumentationRegistry.getArguments()
        val token = requireNotNull(args.getString("token")) { "missing fresh fixture token" }
        val relay = requireNotNull(args.getString("relay")) { "missing UDP relay URL" }
        val controlURL = requireNotNull(args.getString("control")) { "missing LAN fixture control URL" }
        require(token.isNotBlank() && relay.startsWith("https://") && controlURL.startsWith("http://")) {
            "device oracle arguments invalid"
        }
        val identity = E2EKeyPair()
        val digest = MessageDigest.getInstance("SHA-256").digest(identity.publicKeyData)
        val deviceID = Base64.getUrlEncoder().withoutPadding().encodeToString(digest.copyOfRange(0, 16))
        // A self-signed LAN relay is allowed ONLY by this explicit test parameter.
        // The public API and all production calls default to certificate verification.
        val localTLS = RelayTLS(verifyPeer = false)
        val pairing = pairInitiator(token, deviceID, identity.publicKeyData,
            confirm = { sas -> control("$controlURL/sas", "POST", sas) == 204 }, tls = localTLS)
        assertEquals("fixture backend did not register after SAS", 204, control("$controlURL/ready", "GET"))
        connectPaired(relay, pairing.peerInstanceID, deviceID, pairing, tls = localTLS).use { session ->
            session.openStream("chat").use { chat ->
                chat.send("ping".toByteArray(Charsets.UTF_8))
                assertEquals("echo: ping", String(chat.recv(), Charsets.UTF_8))
                // ACK over the same encrypted stream, only after byte-exact echo.
                chat.send("ack: echo: ping".toByteArray(Charsets.UTF_8))
            }
        }
    }
}

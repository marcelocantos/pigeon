// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0
package com.marcelocantos.pigeon

import androidx.test.ext.junit.runners.AndroidJUnit4
import androidx.test.platform.app.InstrumentationRegistry
import org.junit.Assert.fail
import org.junit.Test
import org.junit.runner.RunWith

/** Real-device integration test; deliberately RED until the public Kotlin
 * pairing ceremony contract lands. Never substitute a JVM loopback here. */
@RunWith(AndroidJUnit4::class)
class DeviceEchoTest {
    @Test fun freshPairingThenEncryptedChat() {
        val args = InstrumentationRegistry.getArguments()
        val token = requireNotNull(args.getString("token")) { "missing fresh fixture token" }
        val relay = requireNotNull(args.getString("relay")) { "missing UDP relay URL" }
        val control = requireNotNull(args.getString("control")) { "missing LAN fixture control URL" }
        require(token.isNotBlank() && relay.startsWith("https://") && control.startsWith("http://")) {
            "device oracle arguments invalid"
        }
        // Await precise Kotlin pairing API from T66.3. This must derive a
        // fresh SAS on-device, POST it to $control/sas for comparison with
        // the Go backend's independently derived code, confirm both sides,
        // reconnect using the new PairingRecord, and use a named "chat"
        // encrypted stream for ping -> echo: ping. No success-shaped skip.
        fail("T66.3 Kotlin public pairing ceremony API not integrated")
    }
}

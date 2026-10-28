// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0
package com.marcelocantos.pigeon.jni

import org.junit.jupiter.api.Assertions.assertEquals
import org.junit.jupiter.api.Assertions.assertThrows
import org.junit.jupiter.api.Assertions.assertTrue
import org.junit.jupiter.api.Test

/** JNI boundary checks that fail before any network dial. */
class NativeConnectionContractTest {
    @Test fun pairingChannelIsDistinctFromEstablishedChannel() {
        val pairing = PigeonNative.channelInitPairing()
        assertTrue(pairing != 0L)
        PigeonNative.channelFree(pairing)
    }

    @Test fun connectTokenIsRejectedRatherThanIgnored() {
        val error = assertThrows(IllegalArgumentException::class.java) {
            PigeonNative.connectionOpen("localhost", "4433", "peer", 0,
                "secret", true, null, 100)
        }
        assertTrue(error.message!!.contains("token"))
    }

    @Test fun badConfigurationAndNullHandlesFailBeforeDial() {
        assertThrows(IllegalArgumentException::class.java) {
            PigeonNative.connectionOpen("localhost", "4433", "peer", 3,
                null, true, null, 100)
        }
        assertThrows(IllegalArgumentException::class.java) {
            PigeonNative.connectionSessionInit(0, 0, null, null)
        }
        assertThrows(IllegalArgumentException::class.java) {
            PigeonNative.sessionPrimary(0)
        }
        PigeonNative.connectionClose(0)
    }
}

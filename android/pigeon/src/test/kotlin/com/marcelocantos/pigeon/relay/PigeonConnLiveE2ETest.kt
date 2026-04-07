// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package com.marcelocantos.pigeon.relay

import java.io.InputStream
import java.io.OutputStream
import java.nio.ByteBuffer
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertTrue
import org.junit.jupiter.api.Assumptions

/**
 * End-to-end integration tests against the live pigeon relay.
 *
 * Uses pigeon-bridge (Go binary) for QUIC connectivity. The bridge
 * handles the QUIC connection and pigeon handshake; messages are bridged
 * via stdin/stdout using the same length-prefixed framing.
 *
 * Requires PIGEON_TOKEN env var. Tests are skipped when not set.
 */
class PigeonConnLiveE2ETest {

    private val token: String? = System.getenv("PIGEON_TOKEN")
    private val relayUrl: String = System.getenv("PIGEON_RELAY_URL") ?: "https://carrier-pigeon.fly.dev:4433"
    private val repoRoot: String = findRepoRoot()

    private fun assumeLive() {
        Assumptions.assumeTrue(
            token != null && token.isNotEmpty(),
            "PIGEON_TOKEN not set; skipping live relay test"
        )
    }

    private fun buildBridge() {
        val build = ProcessBuilder("go", "build", "-o", "/tmp/pigeon-bridge-kt", "./cmd/pigeon-bridge")
            .directory(java.io.File(repoRoot))
            .redirectErrorStream(true)
            .start()
        val exit = build.waitFor()
        require(exit == 0) { "go build failed: ${build.inputStream.readAllBytes().decodeToString()}" }
    }

    private fun startBridge(vararg args: String): Process {
        return ProcessBuilder("/tmp/pigeon-bridge-kt", *args)
            .redirectError(ProcessBuilder.Redirect.INHERIT)
            .start()
    }

    private fun writeMsg(out: OutputStream, data: ByteArray) {
        val header = ByteBuffer.allocate(4).putInt(data.size).array()
        out.write(header)
        out.write(data)
        out.flush()
    }

    private fun readMsg(inp: InputStream): ByteArray {
        val header = inp.readNBytes(4)
        if (header.size < 4) throw IllegalStateException("stream closed reading header")
        val length = ByteBuffer.wrap(header).int
        val data = inp.readNBytes(length)
        if (data.size < length) throw IllegalStateException("stream closed reading payload")
        return data
    }

    @Test
    fun `live - register assigns instance ID`() {
        assumeLive()
        buildBridge()

        val proc = startBridge("register", relayUrl, token!!)
        try {
            // Bridge sends instance ID as first message on stdout.
            val id = String(readMsg(proc.inputStream))
            assertTrue(id.isNotEmpty(), "instance ID should be non-empty")
        } finally {
            proc.destroyForcibly()
        }
    }

    @Test
    fun `live - bidirectional stream round-trip`() {
        assumeLive()
        buildBridge()

        val backend = startBridge("register", relayUrl, token!!)
        val id = String(readMsg(backend.inputStream))

        val client = startBridge("connect", relayUrl, id)
        Thread.sleep(1000)  // Let bridge establish QUIC connection

        try {
            // Client → backend
            writeMsg(client.outputStream, "hello from kotlin".toByteArray())
            val msg = String(readMsg(backend.inputStream))
            assertEquals("hello from kotlin", msg)

            // Backend → client
            writeMsg(backend.outputStream, "reply from kotlin".toByteArray())
            val reply = String(readMsg(client.inputStream))
            assertEquals("reply from kotlin", reply)
        } finally {
            client.destroyForcibly()
            backend.destroyForcibly()
        }
    }

    @Test
    fun `live - 10 messages in order`() {
        assumeLive()
        buildBridge()

        val backend = startBridge("register", relayUrl, token!!)
        val id = String(readMsg(backend.inputStream))

        val client = startBridge("connect", relayUrl, id)
        Thread.sleep(1000)

        try {
            for (i in 0 until 10) {
                writeMsg(client.outputStream, "msg-$i".toByteArray())
                val msg = String(readMsg(backend.inputStream))
                assertEquals("msg-$i", msg)
            }
        } finally {
            client.destroyForcibly()
            backend.destroyForcibly()
        }
    }

    companion object {
        private fun findRepoRoot(): String {
            // Walk up from CWD to find go.mod
            var dir = java.io.File(System.getProperty("user.dir"))
            while (dir.parentFile != null) {
                if (java.io.File(dir, "go.mod").exists()) return dir.absolutePath
                dir = dir.parentFile
            }
            return System.getProperty("user.dir")
        }
    }
}

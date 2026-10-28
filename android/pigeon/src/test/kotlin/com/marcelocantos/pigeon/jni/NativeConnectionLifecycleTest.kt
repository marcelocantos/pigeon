// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0
package com.marcelocantos.pigeon.jni

import com.marcelocantos.pigeon.crypto.E2EKeyPair
import com.marcelocantos.pigeon.peer.RelayTLS
import com.marcelocantos.pigeon.peer.pairInitiator
import com.marcelocantos.pigeon.peer.toNativeBytes
import java.io.File
import java.net.DatagramSocket
import java.net.HttpURLConnection
import java.net.ServerSocket
import java.net.URL
import java.nio.file.Files
import java.nio.file.Path
import java.security.MessageDigest
import java.util.Base64
import java.util.concurrent.CountDownLatch
import java.util.concurrent.Executors
import java.util.concurrent.TimeUnit
import kotlinx.coroutines.runBlocking
import org.junit.jupiter.api.Assertions.assertArrayEquals
import org.junit.jupiter.api.Assertions.assertThrows
import org.junit.jupiter.api.Assertions.assertTrue
import org.junit.jupiter.api.Test

/** A real desktop JNI ngtcp2 connection against a private Go relay/one-shot backend.
 * No device, mock transport, public relay, or success-shaped skip. The native
 * Android-arm64 binary still needs the separate Fold instrumentation oracle. */
class NativeConnectionLifecycleTest {
    private fun root(): Path {
        var dir = Path.of(System.getProperty("user.dir")).toAbsolutePath()
        while (!Files.exists(dir.resolve("go.mod"))) {
            dir = dir.parent ?: error("go.mod not found above working directory")
        }
        return dir
    }

    private fun port(tcp: Boolean): Int = if (tcp) ServerSocket(0).use { it.localPort }
        else DatagramSocket(0).use { it.localPort }

    private fun http(url: String, body: String? = null): String {
        val c = URL(url).openConnection() as HttpURLConnection
        try {
            c.connectTimeout = 3_000
            c.readTimeout = 40_000
            if (body != null) {
                c.requestMethod = "POST"
                c.doOutput = true
                c.outputStream.use { it.write(body.toByteArray(Charsets.UTF_8)) }
            }
            check(c.responseCode in 200..299) { "$url: HTTP ${c.responseCode}" }
            return c.inputStream.bufferedReader().use { it.readText() }
        } finally { c.disconnect() }
    }

    private fun build(repo: Path, path: Path, pkg: String) {
        val p = ProcessBuilder("go", "build", "-o", path.toString(), pkg)
            .directory(repo.toFile()).redirectErrorStream(true).start()
        val output = p.inputStream.bufferedReader().use { it.readText() }
        check(p.waitFor(120, TimeUnit.SECONDS) && p.exitValue() == 0) { "go build $pkg: $output" }
    }

    private fun start(repo: Path, log: Path, vararg argv: String): Process =
        ProcessBuilder(*argv).directory(repo.toFile()).redirectErrorStream(true)
            .redirectOutput(log.toFile()).start()

    private fun text(path: Path): String = String(Files.readAllBytes(path), Charsets.UTF_8)

    private fun waitForControl(process: Process, log: Path): Int {
        repeat(160) {
            check(process.isAlive) { "fixture exited: ${text(log)}" }
            val match = Regex("(?m)^CONTROL=127\\.0\\.0\\.1:(\\d+)$")
                .find(text(log))
            if (match != null) return match.groupValues[1].toInt()
            Thread.sleep(100)
        }
        error("fixture control port unavailable: ${text(log)}")
    }

    @Test fun failedConnectLiveOwnershipAndRecvFirstReply() = runBlocking {
        val repo = root()
        val tmp = Files.createTempDirectory("pigeon-jni-lifecycle-")
        val relayBin = tmp.resolve("relay")
        val fixtureBin = tmp.resolve("fixture")
        build(repo, relayBin, "./cmd/pigeon")
        build(repo, fixtureBin, "./cmd/pigeon-android-fixture")
        val udp = port(false)
        val tcp = port(true)
        val relayUrl = "https://127.0.0.1:$udp"
        val relay = ProcessBuilder(relayBin.toString(), "-port", tcp.toString(),
            "-quic-port", udp.toString()).directory(repo.toFile())
            .redirectErrorStream(true).redirectOutput(tmp.resolve("relay.log").toFile())
            .apply { environment()["PIGEON_TOKEN"] = "" }.start()
        var fixture: Process? = null
        try {
            fixture = start(repo, tmp.resolve("fixture.log"), fixtureBin.toString(),
                "-relay", relayUrl, "-control", "127.0.0.1:0", "-timeout", "150s")
            val controlPort = waitForControl(fixture, tmp.resolve("fixture.log"))
            val control = "http://127.0.0.1:$controlPort"
            val token = http("$control/token").trim()
            check(token.isNotEmpty()) { "fixture returned empty pairing token" }
            val identity = E2EKeyPair()
            val digest = MessageDigest.getInstance("SHA-256").digest(identity.publicKeyData)
            val deviceID = Base64.getUrlEncoder().withoutPadding()
                .encodeToString(digest.copyOfRange(0, 16))
            val record = pairInitiator(token, deviceID, identity.publicKeyData,
                confirm = { code -> http("$control/sas", code); true },
                tls = RelayTLS(verifyPeer = false))
            http("$control/ready")

            // Native QUIC failure must return no handle and leave a subsequent
            // legitimate connection possible. The unused port is local/private.
            val unused = port(false)
            val beforeFailedDials = File("/dev/fd").list()?.size
                ?: error("/dev/fd unavailable for native failed-connect leak oracle")
            repeat(8) {
                assertThrows(RuntimeException::class.java) {
                    PigeonNative.connectionOpen("127.0.0.1", unused.toString(),
                        record.peerInstanceID, 0, null, false, null, 250)
                }
            }
            val afterFailedDials = File("/dev/fd").list()?.size
                ?: error("/dev/fd disappeared during failed-connect leak oracle")
            assertTrue(afterFailedDials <= beforeFailedDials + 3,
                "failed native connects leaked descriptors: $beforeFailedDials -> $afterFailedDials")
            val conn = PigeonNative.connectionOpen("127.0.0.1", udp.toString(),
                record.peerInstanceID, 0, null, false, null, 5_000)
            assertTrue(conn != 0L)
            val session = try {
                PigeonNative.connectionActivate(conn, deviceID, record.toNativeBytes(), null, null)
            } catch (t: Throwable) {
                PigeonNative.connectionClose(conn)
                throw t
            }
            assertTrue(session != 0L)
            try {
                assertThrows(IllegalStateException::class.java) { PigeonNative.connectionClose(conn) }
                val stream = PigeonNative.sessionOpenStream(session, "chat")
                assertTrue(stream != 0L)
                val pool = Executors.newSingleThreadExecutor()
                try {
                    // Receive is already waiting before the second JVM lane
                    // enters native send. A connection-wide blocking mutex
                    // would time out here instead of delivering the echo.
                    val entered = CountDownLatch(1)
                    val receiving = pool.submit<ByteArray> {
                        entered.countDown()
                        PigeonNative.streamRecv(stream)
                    }
                    check(entered.await(2, TimeUnit.SECONDS)) { "receive lane never started" }
                    Thread.sleep(100)
                    PigeonNative.streamSend(stream, "ping".toByteArray())
                    val reply = try { receiving.get(12, TimeUnit.SECONDS) } catch (t: Throwable) {
                        throw AssertionError("native recv failed; fixture: ${text(tmp.resolve("fixture.log"))}; " +
                            "relay: ${text(tmp.resolve("relay.log"))}", t)
                    }
                    assertArrayEquals("echo: ping".toByteArray(), reply)
                } finally {
                    pool.shutdownNow()
                    PigeonNative.streamClose(stream)
                    PigeonNative.streamFree(stream)
                }
            } finally {
                PigeonNative.sessionFree(session)
                PigeonNative.connectionClose(conn)
            }
            check(fixture.waitFor(10, TimeUnit.SECONDS) && fixture.exitValue() == 0) {
                "fixture did not finish: ${text(tmp.resolve("fixture.log"))}"
            }
            assertTrue(text(tmp.resolve("fixture.log")).contains("DEVICE_ORACLE_PASS"))
        } finally {
            fixture?.destroyForcibly()
            relay.destroyForcibly()
            fixture?.waitFor(5, TimeUnit.SECONDS)
            relay.waitFor(5, TimeUnit.SECONDS)
        }
    }
}

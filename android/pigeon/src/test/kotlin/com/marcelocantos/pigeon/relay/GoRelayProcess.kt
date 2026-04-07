// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package com.marcelocantos.pigeon.relay

import java.io.BufferedReader
import java.io.File
import java.io.InputStreamReader
import java.net.DatagramSocket
import java.nio.file.Path
import java.util.concurrent.CountDownLatch
import java.util.concurrent.TimeUnit

/**
 * Manages a Go pigeon relay server as a subprocess for E2E tests.
 *
 * Builds the relay binary (if not already built), starts it on ephemeral
 * ports, waits for the "pigeon starting" log line to confirm readiness,
 * and extracts the listening ports from server output.
 *
 * The relay uses a self-signed certificate (development mode) and
 * listens on 127.0.0.1 for both WebTransport and raw QUIC.
 */
class GoRelayProcess private constructor(
    private val process: Process,
    /** The raw QUIC port the relay is listening on. */
    val quicPort: Int,
    /** The WebTransport port the relay is listening on. */
    val wtPort: Int,
    /** Optional bearer token required for registration. */
    val token: String?,
) : AutoCloseable {

    override fun close() {
        process.destroyForcibly()
        process.waitFor(5, TimeUnit.SECONDS)
    }

    companion object {
        // Find the project root by walking up from the current working dir.
        private fun findProjectRoot(): Path {
            var dir = Path.of(System.getProperty("user.dir"))
            // Walk up until we find go.mod in the pigeon root.
            while (dir.parent != null) {
                if (dir.resolve("go.mod").toFile().exists() &&
                    dir.resolve("cmd/pigeon/main.go").toFile().exists()) {
                    return dir
                }
                dir = dir.parent
            }
            throw IllegalStateException(
                "Could not find pigeon project root (go.mod + cmd/pigeon/main.go) " +
                "walking up from ${System.getProperty("user.dir")}"
            )
        }

        /**
         * Find an available UDP port by briefly binding a DatagramSocket.
         */
        private fun findAvailablePort(): Int {
            DatagramSocket(0).use { return it.localPort }
        }

        /**
         * Generate a self-signed RSA certificate and key in PEM format.
         * RSA is used because kwik's TLS stack (Agent15) has better
         * compatibility with RSA certificates than ECDSA.
         *
         * @return pair of (cert file, key file) in a temp directory
         */
        private fun generateTestCert(): Pair<File, File> {
            val tmpDir = File(System.getProperty("java.io.tmpdir"), "pigeon-test-certs-${System.nanoTime()}")
            tmpDir.mkdirs()
            val keyFile = File(tmpDir, "key.pem")
            val certFile = File(tmpDir, "cert.pem")

            val genResult = ProcessBuilder(
                "openssl", "req", "-x509", "-newkey", "rsa:2048",
                "-keyout", keyFile.absolutePath,
                "-out", certFile.absolutePath,
                "-days", "1",
                "-nodes",
                "-subj", "/CN=localhost",
                "-addext", "subjectAltName=IP:127.0.0.1,DNS:localhost",
            ).redirectErrorStream(true).start()
            val genOutput = genResult.inputStream.bufferedReader().readText()
            val genExit = genResult.waitFor()
            if (genExit != 0) {
                throw IllegalStateException("Failed to generate test certificate (exit $genExit):\n$genOutput")
            }

            return certFile to keyFile
        }

        /**
         * Build and start a pigeon relay server on ephemeral ports.
         *
         * @param token optional bearer token for /register authentication
         * @return a running [GoRelayProcess] ready to accept connections
         */
        fun start(token: String? = null): GoRelayProcess {
            val projectRoot = findProjectRoot()

            // Build the pigeon binary.
            val buildResult = ProcessBuilder("go", "build", "-o", "pigeon-test-binary", "./cmd/pigeon/")
                .directory(projectRoot.toFile())
                .redirectErrorStream(true)
                .start()
            val buildOutput = buildResult.inputStream.bufferedReader().readText()
            val buildExit = buildResult.waitFor()
            if (buildExit != 0) {
                throw IllegalStateException("Failed to build pigeon binary (exit $buildExit):\n$buildOutput")
            }

            // Generate an RSA TLS certificate for the test server.
            val (certFile, keyFile) = generateTestCert()

            // Pick available ports before starting the server.
            val wtPort = findAvailablePort()
            val quicPort = findAvailablePort()

            val env = mutableMapOf<String, String>()
            if (token != null) {
                env["PIGEON_TOKEN"] = token
            }

            val cmd = listOf(
                projectRoot.resolve("pigeon-test-binary").toString(),
                "--port", wtPort.toString(),
                "--quic-port", quicPort.toString(),
                "--cert", certFile.absolutePath,
                "--key", keyFile.absolutePath,
            )

            val pb = ProcessBuilder(cmd)
                .directory(projectRoot.toFile())
                .redirectErrorStream(true)
            pb.environment().putAll(env)

            val process = pb.start()

            // Wait for the "pigeon starting" log line to confirm the server is ready.
            val reader = BufferedReader(InputStreamReader(process.inputStream))
            val readyLatch = CountDownLatch(1)

            val logThread = Thread {
                try {
                    var line: String?
                    while (reader.readLine().also { line = it } != null) {
                        val l = line ?: continue
                        System.err.println("[pigeon-relay] $l")

                        if (l.contains("msg=\"pigeon starting\"") || l.contains("msg=pigeon starting")) {
                            readyLatch.countDown()
                        }
                    }
                } catch (_: Exception) {
                    // Process closed.
                }
            }
            logThread.isDaemon = true
            logThread.start()

            if (!readyLatch.await(30, TimeUnit.SECONDS)) {
                process.destroyForcibly()
                throw IllegalStateException("Timed out waiting for pigeon relay to start")
            }

            return GoRelayProcess(process, quicPort, wtPort, token)
        }
    }
}

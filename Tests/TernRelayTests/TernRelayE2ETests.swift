// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

#if canImport(Network)

import CryptoKit
import Foundation
import Network
import XCTest

@testable import TernCrypto
@testable import TernRelay

// MARK: - Test relay server helper

/// Manages a local tern relay server subprocess for integration tests.
private final class TestRelayServer {
    let process: Process
    let quicPort: UInt16

    /// Build the tern binary and start a local relay on a random port.
    /// The server generates a self-signed TLS certificate automatically.
    /// If `token` is provided, the server requires it for registration.
    static func start(token: String? = nil) throws -> TestRelayServer {
        let repoRoot = URL(fileURLWithPath: #filePath)
            .deletingLastPathComponent()  // TernRelayTests/
            .deletingLastPathComponent()  // Tests/
            .deletingLastPathComponent()  // repo root

        // Build the tern binary.
        let buildProcess = Process()
        buildProcess.executableURL = URL(fileURLWithPath: "/usr/bin/env")
        buildProcess.arguments = ["go", "build", "-o", "/tmp/tern-test-server", "./cmd/tern"]
        buildProcess.currentDirectoryURL = repoRoot
        buildProcess.standardOutput = FileHandle.nullDevice
        buildProcess.standardError = FileHandle.nullDevice
        try buildProcess.run()
        buildProcess.waitUntilExit()
        guard buildProcess.terminationStatus == 0 else {
            throw NSError(
                domain: "TestRelayServer", code: 1,
                userInfo: [NSLocalizedDescriptionKey: "go build failed with status \(buildProcess.terminationStatus)"])
        }

        // Pick a random high port for QUIC.
        let quicPort = try randomAvailablePort()

        // Also pick a random port for WebTransport (required but unused).
        let wtPort = try randomAvailablePort()

        let serverProcess = Process()
        serverProcess.executableURL = URL(fileURLWithPath: "/tmp/tern-test-server")
        serverProcess.arguments = [
            "--port", String(wtPort),
            "--quic-port", String(quicPort),
        ]
        serverProcess.standardOutput = FileHandle.nullDevice
        // Let server stderr go to test output for debugging.
        serverProcess.standardError = FileHandle.standardError

        if let token = token {
            var env = ProcessInfo.processInfo.environment
            env["TERN_TOKEN"] = token
            serverProcess.environment = env
        }

        try serverProcess.run()

        return TestRelayServer(process: serverProcess, quicPort: quicPort)
    }

    private init(process: Process, quicPort: UInt16) {
        self.process = process
        self.quicPort = quicPort
    }

    func stop() {
        if process.isRunning {
            process.terminate()
            process.waitUntilExit()
        }
    }

    /// Wait for the server to be ready by attempting QUIC connections.
    func waitUntilReady(timeout: TimeInterval = 10) async throws {
        let deadline = Date().addingTimeInterval(timeout)
        while Date() < deadline {
            do {
                let conn = try await TernConn.register(
                    host: "127.0.0.1", port: quicPort,
                    quicOptions: Self.insecureQUICOptions()
                )
                conn.close()
                return
            } catch {
                try await Task.sleep(nanoseconds: 200_000_000)  // 200ms
            }
        }
        throw NSError(
            domain: "TestRelayServer", code: 2,
            userInfo: [NSLocalizedDescriptionKey: "Server did not become ready within \(timeout)s"])
    }

    /// Create QUIC options that accept any TLS certificate (for self-signed).
    static func insecureQUICOptions() -> NWProtocolQUIC.Options {
        let options = NWProtocolQUIC.Options(alpn: ["tern"])
        sec_protocol_options_set_verify_block(
            options.securityProtocolOptions,
            { _, _, completion in
                completion(true)
            }, DispatchQueue.global())
        return options
    }

    /// Find a random available UDP port.
    private static func randomAvailablePort() throws -> UInt16 {
        // Bind to port 0 to get an OS-assigned port, then close it.
        let fd = socket(AF_INET, SOCK_DGRAM, 0)
        guard fd >= 0 else {
            throw NSError(
                domain: "TestRelayServer", code: 3,
                userInfo: [NSLocalizedDescriptionKey: "socket() failed"])
        }
        defer { close(fd) }

        var addr = sockaddr_in()
        addr.sin_family = sa_family_t(AF_INET)
        addr.sin_port = 0
        addr.sin_addr.s_addr = INADDR_LOOPBACK.bigEndian

        let bindResult = withUnsafePointer(to: &addr) { ptr in
            ptr.withMemoryRebound(to: sockaddr.self, capacity: 1) { sockPtr in
                bind(fd, sockPtr, socklen_t(MemoryLayout<sockaddr_in>.size))
            }
        }
        guard bindResult == 0 else {
            throw NSError(
                domain: "TestRelayServer", code: 4,
                userInfo: [NSLocalizedDescriptionKey: "bind() failed: \(errno)"])
        }

        var boundAddr = sockaddr_in()
        var addrLen = socklen_t(MemoryLayout<sockaddr_in>.size)
        let nameResult = withUnsafeMutablePointer(to: &boundAddr) { ptr in
            ptr.withMemoryRebound(to: sockaddr.self, capacity: 1) { sockPtr in
                getsockname(fd, sockPtr, &addrLen)
            }
        }
        guard nameResult == 0 else {
            throw NSError(
                domain: "TestRelayServer", code: 5,
                userInfo: [NSLocalizedDescriptionKey: "getsockname() failed"])
        }

        return UInt16(bigEndian: boundAddr.sin_port)
    }
}

// MARK: - E2E Tests

final class TernRelayE2ETests: XCTestCase {

    private var server: TestRelayServer?

    override func setUp() async throws {
        try await super.setUp()
        server = try TestRelayServer.start()
        try await server!.waitUntilReady()
    }

    override func tearDown() async throws {
        server?.stop()
        server = nil
        try await super.tearDown()
    }

    private var quicPort: UInt16 { server!.quicPort }
    private var quicOptions: NWProtocolQUIC.Options { TestRelayServer.insecureQUICOptions() }

    // MARK: - Test 1: Register and get instance ID

    func testRegisterReturnsInstanceID() async throws {
        let backend = try await TernConn.register(
            host: "127.0.0.1", port: quicPort,
            quicOptions: quicOptions
        )
        defer { backend.close() }

        XCTAssertFalse(backend.instanceID.isEmpty, "Instance ID should be non-empty")
        // The server generates hex IDs (8 hex chars = 32 bits).
        XCTAssertTrue(
            backend.instanceID.allSatisfy { $0.isHexDigit },
            "Instance ID should be hex: \(backend.instanceID)")
    }

    // MARK: - Test 2: Bidirectional messaging

    func testBidirectionalMessaging() async throws {
        let backend = try await TernConn.register(
            host: "127.0.0.1", port: quicPort,
            quicOptions: quicOptions
        )
        defer { backend.close() }

        let client = try await TernConn.connect(
            host: "127.0.0.1", port: quicPort,
            instanceID: backend.instanceID,
            quicOptions: quicOptions
        )
        defer { client.close() }

        // Client sends, backend receives.
        try await client.send(Data("hello".utf8))
        let received = try await backend.recv()
        XCTAssertEqual(String(decoding: received, as: UTF8.self), "hello")

        // Backend sends, client receives.
        try await backend.send(Data("world".utf8))
        let reply = try await client.recv()
        XCTAssertEqual(String(decoding: reply, as: UTF8.self), "world")
    }

    // MARK: - Test 3: Multiple messages in sequence

    func testMultipleMessages() async throws {
        let backend = try await TernConn.register(
            host: "127.0.0.1", port: quicPort,
            quicOptions: quicOptions
        )
        defer { backend.close() }

        let client = try await TernConn.connect(
            host: "127.0.0.1", port: quicPort,
            instanceID: backend.instanceID,
            quicOptions: quicOptions
        )
        defer { client.close() }

        let messageCount = 10
        for i in 0..<messageCount {
            try await client.send(Data("msg-\(i)".utf8))
        }

        for i in 0..<messageCount {
            let data = try await backend.recv()
            XCTAssertEqual(
                String(decoding: data, as: UTF8.self), "msg-\(i)",
                "Message \(i) mismatch")
        }
    }

    // MARK: - Test 4: Encrypted messaging via E2EChannel

    func testEncryptedMessaging() async throws {
        let backend = try await TernConn.register(
            host: "127.0.0.1", port: quicPort,
            quicOptions: quicOptions
        )
        defer { backend.close() }

        let client = try await TernConn.connect(
            host: "127.0.0.1", port: quicPort,
            instanceID: backend.instanceID,
            quicOptions: quicOptions
        )
        defer { client.close() }

        // Generate key pairs.
        let backendKP = E2EKeyPair()
        let clientKP = E2EKeyPair()

        // Derive session keys.
        let bSendKey = try backendKP.deriveSessionKey(
            peerPublicKey: clientKP.publicKeyData, info: Data("b-to-c".utf8))
        let bRecvKey = try backendKP.deriveSessionKey(
            peerPublicKey: clientKP.publicKeyData, info: Data("c-to-b".utf8))
        let cSendKey = try clientKP.deriveSessionKey(
            peerPublicKey: backendKP.publicKeyData, info: Data("c-to-b".utf8))
        let cRecvKey = try clientKP.deriveSessionKey(
            peerPublicKey: backendKP.publicKeyData, info: Data("b-to-c".utf8))

        let backendChannel = E2EChannel(sendKey: bSendKey, recvKey: bRecvKey)
        let clientChannel = E2EChannel(sendKey: cSendKey, recvKey: cRecvKey)

        // Client encrypts and sends.
        let plaintext = Data("secret message".utf8)
        let ciphertext = try clientChannel.encrypt(plaintext)
        try await client.send(ciphertext)

        // Backend receives and decrypts.
        let receivedCiphertext = try await backend.recv()
        let decrypted = try backendChannel.decrypt(receivedCiphertext)
        XCTAssertEqual(
            String(decoding: decrypted, as: UTF8.self), "secret message")

        // Backend encrypts and sends back.
        let replyPlain = Data("secret reply".utf8)
        let replyCipher = try backendChannel.encrypt(replyPlain)
        try await backend.send(replyCipher)

        // Client receives and decrypts.
        let receivedReply = try await client.recv()
        let decryptedReply = try clientChannel.decrypt(receivedReply)
        XCTAssertEqual(
            String(decoding: decryptedReply, as: UTF8.self), "secret reply")
    }

    // MARK: - Test 5: Connect to non-existent instance fails

    func testConnectToNonExistentInstanceFails() async throws {
        do {
            _ = try await TernConn.connect(
                host: "127.0.0.1", port: quicPort,
                instanceID: "nonexistent",
                quicOptions: quicOptions
            )
            XCTFail("Should have thrown for non-existent instance")
        } catch {
            // Expected: the server closes the connection with an error.
        }
    }

    // MARK: - Test 6: Multiple backends get unique IDs

    func testMultipleBackendsGetUniqueIDs() async throws {
        let backend1 = try await TernConn.register(
            host: "127.0.0.1", port: quicPort,
            quicOptions: quicOptions
        )
        defer { backend1.close() }

        let backend2 = try await TernConn.register(
            host: "127.0.0.1", port: quicPort,
            quicOptions: quicOptions
        )
        defer { backend2.close() }

        let backend3 = try await TernConn.register(
            host: "127.0.0.1", port: quicPort,
            quicOptions: quicOptions
        )
        defer { backend3.close() }

        let ids = Set([backend1.instanceID, backend2.instanceID, backend3.instanceID])
        XCTAssertEqual(ids.count, 3, "Each backend should receive a unique instance ID")
    }

    // MARK: - Test 7: Binary data with full byte range

    func testBinaryDataRoundTrip() async throws {
        let backend = try await TernConn.register(
            host: "127.0.0.1", port: quicPort,
            quicOptions: quicOptions
        )
        defer { backend.close() }

        let client = try await TernConn.connect(
            host: "127.0.0.1", port: quicPort,
            instanceID: backend.instanceID,
            quicOptions: quicOptions
        )
        defer { client.close() }

        // All 256 byte values including null bytes.
        var binaryData = Data(count: 256)
        for i in 0..<256 {
            binaryData[i] = UInt8(i)
        }

        try await client.send(binaryData)
        let received = try await backend.recv()
        XCTAssertEqual(received, binaryData, "Binary data should survive relay round-trip intact")
    }

    // MARK: - Test 8: Empty message

    func testEmptyMessage() async throws {
        let backend = try await TernConn.register(
            host: "127.0.0.1", port: quicPort,
            quicOptions: quicOptions
        )
        defer { backend.close() }

        let client = try await TernConn.connect(
            host: "127.0.0.1", port: quicPort,
            instanceID: backend.instanceID,
            quicOptions: quicOptions
        )
        defer { client.close() }

        try await client.send(Data())
        let received = try await backend.recv()
        XCTAssertEqual(received, Data(), "Empty message should round-trip as empty Data")
    }

    // MARK: - Test 9: Large message (64 KB)

    func testLargeMessage() async throws {
        let backend = try await TernConn.register(
            host: "127.0.0.1", port: quicPort,
            quicOptions: quicOptions
        )
        defer { backend.close() }

        let client = try await TernConn.connect(
            host: "127.0.0.1", port: quicPort,
            instanceID: backend.instanceID,
            quicOptions: quicOptions
        )
        defer { client.close() }

        let largeData = Data(repeating: 0xAB, count: 64 * 1024)
        try await client.send(largeData)
        let received = try await backend.recv()
        XCTAssertEqual(received.count, largeData.count)
        XCTAssertEqual(received, largeData)
    }

    // MARK: - Test 10: Datagrams

    func testDatagramForwarding() async throws {
        let backend = try await TernConn.register(
            host: "127.0.0.1", port: quicPort,
            quicOptions: quicOptions
        )
        defer { backend.close() }

        let client = try await TernConn.connect(
            host: "127.0.0.1", port: quicPort,
            instanceID: backend.instanceID,
            quicOptions: quicOptions
        )
        defer { client.close() }

        // Client -> backend datagram.
        try await client.sendDatagram(Data("dgram-c2b".utf8))
        let received = try await backend.recvDatagram()
        XCTAssertEqual(String(decoding: received, as: UTF8.self), "dgram-c2b")

        // Backend -> client datagram.
        try await backend.sendDatagram(Data("dgram-b2c".utf8))
        let reply = try await client.recvDatagram()
        XCTAssertEqual(String(decoding: reply, as: UTF8.self), "dgram-b2c")
    }
}

// MARK: - Token auth E2E tests

/// Separate test class for token-authenticated relay — needs its own server.
final class TernRelayTokenAuthE2ETests: XCTestCase {

    private var server: TestRelayServer?
    private let testToken = "e2e-test-secret"

    override func setUp() async throws {
        try await super.setUp()
        server = try TestRelayServer.start(token: testToken)
        // Wait for ready using the correct token.
        let deadline = Date().addingTimeInterval(10)
        while Date() < deadline {
            do {
                let conn = try await TernConn.register(
                    host: "127.0.0.1", port: server!.quicPort,
                    token: testToken,
                    quicOptions: TestRelayServer.insecureQUICOptions()
                )
                conn.close()
                return
            } catch {
                try await Task.sleep(nanoseconds: 200_000_000)
            }
        }
        XCTFail("Token-auth server did not become ready")
    }

    override func tearDown() async throws {
        server?.stop()
        server = nil
        try await super.tearDown()
    }

    private var quicPort: UInt16 { server!.quicPort }
    private var quicOptions: NWProtocolQUIC.Options { TestRelayServer.insecureQUICOptions() }

    func testRegisterWithCorrectToken() async throws {
        let backend = try await TernConn.register(
            host: "127.0.0.1", port: quicPort,
            token: testToken,
            quicOptions: quicOptions
        )
        defer { backend.close() }

        XCTAssertFalse(backend.instanceID.isEmpty)
    }

    func testRegisterWithoutTokenFails() async throws {
        // Without a token, registration should fail. The server closes the
        // QUIC connection, which may surface as a handshake read failure.
        do {
            let conn = try await TernConn.register(
                host: "127.0.0.1", port: quicPort,
                quicOptions: quicOptions
            )
            // If the handshake appeared to succeed, verify the connection is
            // actually dead by trying to use it.
            defer { conn.close() }
            try await Task.sleep(nanoseconds: 500_000_000)
            try await conn.send(Data("probe".utf8))
            let _ = try await conn.recv()
            XCTFail("Expected rejection without token")
        } catch {
            // Expected — server rejected the connection.
        }
    }

    func testRegisterWithWrongTokenFails() async throws {
        do {
            let conn = try await TernConn.register(
                host: "127.0.0.1", port: quicPort,
                token: "wrong-token",
                quicOptions: quicOptions
            )
            defer { conn.close() }
            try await Task.sleep(nanoseconds: 500_000_000)
            try await conn.send(Data("probe".utf8))
            let _ = try await conn.recv()
            XCTFail("Expected rejection with wrong token")
        } catch {
            // Expected.
        }
    }

    func testTokenAuthBidirectional() async throws {
        let backend = try await TernConn.register(
            host: "127.0.0.1", port: quicPort,
            token: testToken,
            quicOptions: quicOptions
        )
        defer { backend.close() }

        let client = try await TernConn.connect(
            host: "127.0.0.1", port: quicPort,
            instanceID: backend.instanceID,
            quicOptions: quicOptions
        )
        defer { client.close() }

        try await client.send(Data("authed-hello".utf8))
        let received = try await backend.recv()
        XCTAssertEqual(String(decoding: received, as: UTF8.self), "authed-hello")
    }
}

// MARK: - Live relay tests

final class TernRelayLiveE2ETests: XCTestCase {

    private var token: String?
    private let liveHost = "tern.fly.dev"
    private let livePort: UInt16 = 4433

    override func setUp() async throws {
        try await super.setUp()
        token = ProcessInfo.processInfo.environment["TERN_TOKEN"]
        try XCTSkipIf(token == nil || token!.isEmpty, "TERN_TOKEN not set; skipping live relay tests")
    }

    func testLiveRegisterAndConnect() async throws {
        let backend = try await TernConn.register(
            host: liveHost, port: livePort,
            token: token
        )
        defer { backend.close() }

        XCTAssertFalse(backend.instanceID.isEmpty, "Live instance ID should be non-empty")

        let client = try await TernConn.connect(
            host: liveHost, port: livePort,
            instanceID: backend.instanceID
        )
        defer { client.close() }

        // Bidirectional message exchange.
        try await client.send(Data("live-hello".utf8))
        let received = try await backend.recv()
        XCTAssertEqual(String(decoding: received, as: UTF8.self), "live-hello")

        try await backend.send(Data("live-reply".utf8))
        let reply = try await client.recv()
        XCTAssertEqual(String(decoding: reply, as: UTF8.self), "live-reply")
    }
}

#endif

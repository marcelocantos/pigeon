// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// E2E relay tests that start a local Go relay subprocess, run
// register/connect/stream/crypto round-trips, and tear down.
// Adapted from e2e/swift/main.swift for XCTest integration.

#if canImport(Network)

import XCTest
@testable import Pigeon
import Foundation
import Network

final class RelayE2ETests: XCTestCase {

    private var relayProcess: Process!
    private var relayPort: UInt16!

    override func setUpWithError() throws {
        try super.setUpWithError()

        // Build the relay binary.
        let repoRoot = URL(fileURLWithPath: #filePath)
            .deletingLastPathComponent()  // PigeonRelayE2ETests/
            .deletingLastPathComponent()  // Tests/
            .deletingLastPathComponent()  // repo root

        let build = Process()
        build.executableURL = URL(fileURLWithPath: "/usr/bin/env")
        build.arguments = ["go", "build", "-o", "/tmp/pigeon-e2e-server", "./cmd/pigeon"]
        build.currentDirectoryURL = repoRoot
        build.standardOutput = FileHandle.nullDevice
        build.standardError = FileHandle.nullDevice
        try build.run()
        build.waitUntilExit()
        guard build.terminationStatus == 0 else {
            throw NSError(domain: "Build", code: 1,
                          userInfo: [NSLocalizedDescriptionKey: "go build failed (\(build.terminationStatus))"])
        }

        // Find a free UDP port.
        relayPort = Self.findFreePort()

        // Start the relay.
        let proc = Process()
        proc.executableURL = URL(fileURLWithPath: "/tmp/pigeon-e2e-server")
        proc.arguments = ["--quic-port", String(relayPort)]
        proc.standardOutput = FileHandle.nullDevice
        let pipe = Pipe()
        proc.standardError = pipe
        try proc.run()

        // Wait for "pigeon starting" on stderr.
        let deadline = Date().addingTimeInterval(15)
        var ready = false
        pipe.fileHandleForReading.readabilityHandler = { h in
            if let s = String(data: h.availableData, encoding: .utf8),
               s.contains("pigeon starting") {
                ready = true
            }
        }
        while !ready && Date() < deadline {
            Thread.sleep(forTimeInterval: 0.1)
        }
        pipe.fileHandleForReading.readabilityHandler = nil

        guard ready else {
            proc.terminate()
            throw NSError(domain: "Server", code: 2,
                          userInfo: [NSLocalizedDescriptionKey: "relay did not start within 15s"])
        }

        relayProcess = proc
    }

    override func tearDown() {
        relayProcess?.terminate()
        relayProcess?.waitUntilExit()
        relayProcess = nil
        super.tearDown()
    }

    // MARK: - Tests

    func testRegister() async throws {
        let (conn, id) = try await register()
        XCTAssertFalse(id.isEmpty, "instance ID should not be empty")
        conn.cancel()
    }

    func testStreamRoundTrip() async throws {
        let (backend, id) = try await register()
        let client = try await connect(id)

        try await writeMsg(client, Data("hello from swift".utf8))
        let msg = try await readMsg(backend)
        XCTAssertEqual(String(decoding: msg, as: UTF8.self), "hello from swift")

        try await writeMsg(backend, Data("reply from swift".utf8))
        let reply = try await readMsg(client)
        XCTAssertEqual(String(decoding: reply, as: UTF8.self), "reply from swift")

        backend.cancel(); client.cancel()
    }

    func testTenMessagesInOrder() async throws {
        let (backend, id) = try await register()
        let client = try await connect(id)

        for i in 0..<10 {
            try await writeMsg(client, Data("msg-\(i)".utf8))
        }
        for i in 0..<10 {
            let d = try await readMsg(backend)
            XCTAssertEqual(String(decoding: d, as: UTF8.self), "msg-\(i)")
        }

        backend.cancel(); client.cancel()
    }

    func testEncryptedRoundTrip() async throws {
        let (backend, id) = try await register()
        let client = try await connect(id)

        let bKP = E2EKeyPair(), cKP = E2EKeyPair()

        // Exchange public keys through relay.
        try await writeMsg(client, cKP.publicKeyData)
        try await writeMsg(backend, bKP.publicKeyData)
        let cPub = try await readMsg(backend)
        let bPub = try await readMsg(client)

        // Derive keys, create channels.
        let bSend = try bKP.deriveSessionKey(peerPublicKey: cPub, info: Data("b2c".utf8))
        let bRecv = try bKP.deriveSessionKey(peerPublicKey: cPub, info: Data("c2b".utf8))
        let cSend = try cKP.deriveSessionKey(peerPublicKey: bPub, info: Data("c2b".utf8))
        let cRecv = try cKP.deriveSessionKey(peerPublicKey: bPub, info: Data("b2c".utf8))
        let bCh = E2EChannel(sendKey: bSend, recvKey: bRecv)
        let cCh = E2EChannel(sendKey: cSend, recvKey: cRecv)

        // Client → backend encrypted.
        let pt = Data("secret from swift".utf8)
        try await writeMsg(client, try cCh.encrypt(pt))
        let ct = try await readMsg(backend)
        let decrypted = try bCh.decrypt(ct)
        XCTAssertEqual(decrypted, pt)

        // Backend → client encrypted.
        let reply = Data("secret reply".utf8)
        try await writeMsg(backend, try bCh.encrypt(reply))
        let replyCt = try await readMsg(client)
        XCTAssertEqual(try cCh.decrypt(replyCt), reply)

        backend.cancel(); client.cancel()
    }

    func testConfirmationCodeCrossplatformVector() {
        let code = deriveConfirmationCode(
            Data(repeating: 0x01, count: 32),
            Data(repeating: 0x02, count: 32)
        )
        XCTAssertEqual(code, "629624")
    }

    func testConfirmationCodesMatch() {
        let a = E2EKeyPair(), b = E2EKeyPair()
        let codeA = deriveConfirmationCode(a.publicKeyData, b.publicKeyData)
        let codeB = deriveConfirmationCode(b.publicKeyData, a.publicKeyData)
        XCTAssertEqual(codeA, codeB)
        XCTAssertEqual(codeA.count, 6)
    }

    // MARK: - QUIC helpers

    private func quicConnect() async throws -> NWConnection {
        let opts = NWProtocolQUIC.Options(alpn: ["pigeon"])
        sec_protocol_options_set_verify_block(
            opts.securityProtocolOptions, { _, _, c in c(true) }, .main
        )
        let params = NWParameters(quic: opts)
        let ep = NWEndpoint.hostPort(
            host: .init("127.0.0.1"),
            port: NWEndpoint.Port(rawValue: relayPort)!
        )
        let q = DispatchQueue(label: "e2e.\(arc4random())")
        let conn = NWConnection(to: ep, using: params)

        try await withThrowingTaskGroup(of: Void.self) { group in
            group.addTask {
                try await withCheckedThrowingContinuation { (c: CheckedContinuation<Void, Error>) in
                    final class Guard: @unchecked Sendable { var done = false }
                    let g = Guard()
                    conn.stateUpdateHandler = { s in
                        guard !g.done else { return }
                        if case .ready = s { g.done = true; c.resume() }
                        else if case .failed(let e) = s { g.done = true; c.resume(throwing: e) }
                    }
                    conn.start(queue: q)
                }
            }
            group.addTask {
                try await Task.sleep(nanoseconds: 10_000_000_000)
                throw NSError(domain: "Timeout", code: 0,
                              userInfo: [NSLocalizedDescriptionKey: "QUIC connect timeout"])
            }
            try await group.next()!
            group.cancelAll()
        }
        return conn
    }

    private func register() async throws -> (NWConnection, String) {
        let c = try await quicConnect()
        try await writeMsg(c, Data("register".utf8))
        let id = String(decoding: try await readMsg(c), as: UTF8.self)
        return (c, id)
    }

    private func connect(_ id: String) async throws -> NWConnection {
        let c = try await quicConnect()
        try await writeMsg(c, Data("connect:\(id)".utf8))
        return c
    }

    private func writeMsg(_ c: NWConnection, _ payload: Data) async throws {
        var h = Data(count: 4)
        let len = UInt32(payload.count)
        h[0] = UInt8((len >> 24) & 0xFF); h[1] = UInt8((len >> 16) & 0xFF)
        h[2] = UInt8((len >> 8) & 0xFF); h[3] = UInt8(len & 0xFF)
        try await withCheckedThrowingContinuation { (cont: CheckedContinuation<Void, Error>) in
            c.send(content: h + payload, completion: .contentProcessed { err in
                if let err = err { cont.resume(throwing: err) } else { cont.resume() }
            })
        }
    }

    private func readMsg(_ c: NWConnection) async throws -> Data {
        let hdr = try await readExact(c, 4)
        let b0 = UInt32(hdr[0]) << 24
        let b1 = UInt32(hdr[1]) << 16
        let b2 = UInt32(hdr[2]) << 8
        let b3 = UInt32(hdr[3])
        let len = Int(b0 | b1 | b2 | b3)
        if len == 0 { return Data() }
        return try await readExact(c, len)
    }

    private func readExact(_ c: NWConnection, _ count: Int) async throws -> Data {
        try await withCheckedThrowingContinuation { cont in
            c.receive(minimumIncompleteLength: count, maximumLength: count) { d, _, _, e in
                if let e = e { cont.resume(throwing: e) }
                else if let d = d, d.count >= count { cont.resume(returning: d) }
                else {
                    cont.resume(throwing: NSError(
                        domain: "EOF", code: 0,
                        userInfo: [NSLocalizedDescriptionKey: "expected \(count) bytes, got \(d?.count ?? 0)"]
                    ))
                }
            }
        }
    }

    // MARK: - Port helper

    private static func findFreePort() -> UInt16 {
        let sock = socket(AF_INET, SOCK_DGRAM, 0)
        defer { close(sock) }
        var addr = sockaddr_in()
        addr.sin_family = sa_family_t(AF_INET)
        addr.sin_port = 0
        addr.sin_addr.s_addr = INADDR_ANY.bigEndian
        var len = socklen_t(MemoryLayout<sockaddr_in>.size)
        withUnsafePointer(to: &addr) {
            $0.withMemoryRebound(to: sockaddr.self, capacity: 1) { _ = Darwin.bind(sock, $0, len) }
        }
        withUnsafeMutablePointer(to: &addr) {
            $0.withMemoryRebound(to: sockaddr.self, capacity: 1) { _ = getsockname(sock, $0, &len) }
        }
        return UInt16(bigEndian: addr.sin_port)
    }
}

#endif

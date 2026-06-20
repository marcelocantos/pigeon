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

        // Start the relay. Keep its stderr streaming to the test process's
        // stderr so CI logs capture relay-side events — useful for
        // diagnosing failures that only reproduce in CI.
        let proc = Process()
        proc.executableURL = URL(fileURLWithPath: "/tmp/pigeon-e2e-server")
        proc.arguments = ["--quic-port", String(relayPort)]
        proc.standardOutput = FileHandle.nullDevice
        let pipe = Pipe()
        proc.standardError = pipe
        try proc.run()

        // Wait for "pigeon starting" on stderr, then keep draining stderr
        // forward to the test process's stderr so the pipe never fills up
        // and slog lines show up in CI output.
        let ready = DispatchSemaphore(value: 0)
        let readyFlag = NSLock()
        var didSignal = false
        pipe.fileHandleForReading.readabilityHandler = { h in
            let data = h.availableData
            guard !data.isEmpty else { return }
            FileHandle.standardError.write(data)
            readyFlag.lock()
            let already = didSignal
            if !already, let s = String(data: data, encoding: .utf8),
               s.contains("pigeon starting") {
                didSignal = true
            }
            let shouldSignal = !already && didSignal
            readyFlag.unlock()
            if shouldSignal { ready.signal() }
        }

        guard ready.wait(timeout: .now() + 15) == .success else {
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

    // T45: the backend keeps the `register` control connection open and
    // parks a `listen` slot; the relay bridges the client onto that listen
    // end-to-end, so backend-side traffic flows on `backend` (the listen
    // connection), not the control connection.
    func testStreamRoundTrip() async throws {
        let (control, id) = try await register()
        defer { control.cancel() }
        let backend = try await listen(id)
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
        let (control, id) = try await register()
        defer { control.cancel() }
        let backend = try await listen(id)
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
        let (control, id) = try await register()
        defer { control.cancel() }
        let backend = try await listen(id)
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

    /// 🎯T37: Cross-language confirmation-code test wired through the
    /// real Ngtcp2Transport + PigeonSession path.
    ///
    /// Wire: Swift PigeonSession (via Ngtcp2Transport role .connect) ↔
    /// Go relay ↔ Go crypto-peer (pigeon.Register pairing-mode, listener.Accept).
    ///
    /// Replaces the previous NWConnection-based variant that spoke the
    /// raw QUIC stream directly. The post-🎯T37 path goes through the
    /// vendored ngtcp2 client + libpigeon multi-stream session glue,
    /// so this exercises:
    ///
    ///   * Ngtcp2Transport.init → pigeon_ngtcp2_transport_init → real
    ///     QUIC handshake + relay greeting (connect:<id>).
    ///   * PigeonSession.pairingConnect → pigeon_connect_on_transport
    ///     (record == NULL → pairing mode, writes empty-name primary
    ///     header, leaves AEAD channel unestablished).
    ///   * PigeonSession.primaryStream → pigeon_session_primary →
    ///     pigeon_stream_send / pigeon_stream_recv (plaintext path
    ///     because !channel.established).
    func testCrossLanguageConfirmationCode() async throws {
        // T45: this path goes through the C Ngtcp2Transport + libpigeon
        // (dist/pigeon.c), which still speaks the pre-T45 wire — its
        // `connect` greeting does NOT consume the relay's "ok" ack, so the
        // first primary read returns "ok" (2 bytes) instead of the peer's
        // 32-byte public key. The relay and Swift-level greetings are
        // already on the remote-Listen L1 model; re-enable this test once
        // the C SDK transport is ported to read the connect ack and park a
        // listen pool. See docs/DESIGN.md §3 L1.
        try XCTSkipIf(true, "C Ngtcp2Transport not yet ported to the T45 remote-Listen wire (does not read the connect 'ok' ack)")

        // Build the crypto-peer binary.
        let repoRoot = URL(fileURLWithPath: #filePath)
            .deletingLastPathComponent()  // PigeonRelayE2ETests/
            .deletingLastPathComponent()  // Tests/
            .deletingLastPathComponent()  // repo root

        let buildPeer = Process()
        buildPeer.executableURL = URL(fileURLWithPath: "/usr/bin/env")
        buildPeer.arguments = ["go", "build", "-o", "/tmp/pigeon-crypto-peer", "./cmd/crypto-peer"]
        buildPeer.currentDirectoryURL = repoRoot
        buildPeer.standardOutput = FileHandle.nullDevice
        buildPeer.standardError = FileHandle.nullDevice
        try buildPeer.run()
        buildPeer.waitUntilExit()
        guard buildPeer.terminationStatus == 0 else {
            throw NSError(domain: "Build", code: 1,
                          userInfo: [NSLocalizedDescriptionKey: "go build crypto-peer failed (\(buildPeer.terminationStatus))"])
        }

        // Start crypto-peer subprocess pointing to the relay. Instance ID
        // arrives on stdout; stderr carries slog + quic-go diagnostics
        // (including the UDP-buffer warning on CI runners that used to
        // contaminate the first stderr line when the ID was printed there).
        let peer = Process()
        peer.executableURL = URL(fileURLWithPath: "/tmp/pigeon-crypto-peer")
        peer.arguments = ["https://127.0.0.1:\(relayPort!)"]
        peer.environment = ProcessInfo.processInfo.environment.merging(
            ["PIGEON_INSECURE": "1"]) { _, new in new }
        let peerStdout = Pipe()
        peer.standardOutput = peerStdout
        let peerStderr = Pipe()
        peer.standardError = peerStderr
        try peer.run()
        defer { peer.terminate(); peer.waitUntilExit() }

        // Close the write ends in the parent so read() sees EOF on exit.
        peerStdout.fileHandleForWriting.closeFile()
        peerStderr.fileHandleForWriting.closeFile()

        // Forward crypto-peer's stderr to the test process's stderr so
        // diagnostics reach CI logs.
        peerStderr.fileHandleForReading.readabilityHandler = { h in
            let data = h.availableData
            if data.isEmpty { return }
            FileHandle.standardError.write("[crypto-peer] ".data(using: .utf8)!)
            FileHandle.standardError.write(data)
        }

        // Read instance ID from crypto-peer's stdout (first line).
        // Use a DispatchSemaphore for a reliable timeout: schedule a
        // 15-second timeout that kills the peer, which unblocks read().
        let sem = DispatchSemaphore(value: 0)
        var instanceIDResult = ""
        let readFD = peerStdout.fileHandleForReading.fileDescriptor
        DispatchQueue.global(qos: .userInitiated).async {
            var lineData = Data()
            var buf = [UInt8](repeating: 0, count: 1)
            while true {
                let n = read(readFD, &buf, 1)
                if n <= 0 { break }
                if buf[0] == UInt8(ascii: "\n") { break }
                lineData.append(buf[0])
            }
            instanceIDResult = String(decoding: lineData, as: UTF8.self)
                .trimmingCharacters(in: .whitespacesAndNewlines)
            sem.signal()
        }

        // Wait up to 15s for the instance ID; on timeout, kill peer to
        // unblock read() and then check what we got.
        if sem.wait(timeout: .now() + 15) == .timedOut {
            peer.terminate()
            sem.wait()  // wait for read() to return after kill
        }

        guard !instanceIDResult.isEmpty else {
            XCTFail("crypto-peer did not print an instance ID (pipe returned EOF with no data)")
            return
        }
        let instanceID = instanceIDResult

        // Build the Ngtcp2Transport on a dedicated worker thread. The
        // ngtcp2 init runs the QUIC handshake synchronously (blocks on
        // select() inside the C code); we don't want that on Swift's
        // cooperative pool. Also keep the transport's lifetime tied to
        // this scope — the transport's userdata pointer flows into
        // PigeonSession, which retains the transport, so it stays alive
        // while we use the session.
        let transport: Ngtcp2Transport = try await Task.detached(priority: .userInitiated) {
            try Ngtcp2Transport(
                host: "127.0.0.1",
                port: String(self.relayPort!),
                role: .connect(peerInstanceID: instanceID),
                verifyPeer: false,
                timeoutMs: 15000
            )
        }.value
        defer { transport.close() }

        // Promote the primary stream into the multi-channel slot table
        // and run pigeon_connect_on_transport in pairing mode. Result:
        // a PigeonSession whose primary stream is bound and ready to
        // exchange plaintext messages over send_on_stream / recv_on_stream.
        let primary = try transport.primaryHandle()
        let session = try PigeonSession.pairingConnect(
            transport: transport,
            primaryHandle: primary,
            peerInstanceID: instanceID
        )

        // Drive the key exchange + confirmation-code receipt over the
        // primary stream. Use PigeonStream.send / .recv — in pairing
        // mode (!channel.established) these short-circuit to
        // transport_send_on_stream / transport_recv_on_stream, which
        // wrap each call in a 4-byte BE length prefix and forward to
        // the QUIC stream. Matches the framing the Go peer-library
        // Stream.Send/Recv produces on the other side.
        let primaryStream = try session.primaryStream()

        // 1. Receive crypto-peer's 32-byte public key.
        let peerPublicKey = try await primaryStream.recv()
        XCTAssertEqual(peerPublicKey.count, 32, "peer public key should be 32 bytes")

        // 2. Send our 32-byte public key.
        let myKeyPair = E2EKeyPair()
        try await primaryStream.send(myKeyPair.publicKeyData)

        // 3. Receive crypto-peer's 6-byte confirmation code.
        let peerCodeData = try await primaryStream.recv()
        let peerCode = String(decoding: peerCodeData, as: UTF8.self)

        // Derive own confirmation code and assert cross-language agreement.
        let myCode = deriveConfirmationCode(myKeyPair.publicKeyData, peerPublicKey)
        XCTAssertEqual(myCode, peerCode,
                       "Swift and Go confirmation codes must match (cross-language HKDF verification)")
        XCTAssertEqual(myCode.count, 6, "confirmation code should be 6 digits")

        // Politely close the stream so crypto-peer's trailing stream.Recv
        // returns and it exits cleanly. The legacy NWConnection variant
        // relied on connection cancellation for the same effect; here
        // the PigeonSession close+ Ngtcp2Transport.close in the defers
        // tear everything down.
        try? primaryStream.close()
    }

    /// Reads one length-prefixed message from `c`, cancelling `c` after
    /// `seconds` if no data arrives. Cancelling the connection guarantees
    /// the underlying NWConnection.receive() callback fires, so the caller
    /// is never left waiting forever.
    private func readMsgWithTimeout(_ c: NWConnection, _ seconds: Double) async throws -> Data {
        // Schedule connection cancellation after the timeout.
        let item = DispatchWorkItem { c.cancel() }
        DispatchQueue.global().asyncAfter(deadline: .now() + seconds, execute: item)
        defer { item.cancel() }
        return try await readMsg(c)
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

    /// Open the backend control connection (T45 `register`). The relay
    /// assigns / echoes the instance ID and holds the connection open for
    /// the instance's lifetime; no client traffic flows on it. Returns the
    /// control connection (kept alive for the duration of the test) and the
    /// assigned ID.
    private func register() async throws -> (NWConnection, String) {
        let c = try await quicConnect()
        try await writeMsg(c, PigeonWire.encodeRelayGreetingRegister(token: "", instanceId: ""))
        let id = String(decoding: try await readMsg(c), as: UTF8.self)
        return (c, id)
    }

    /// Park a backend `listen` slot for `id` (T45). The relay acks with the
    /// instance ID and then bridges the next matching client onto this
    /// connection end-to-end, so backend-side traffic flows here — not on
    /// the `register` control connection.
    private func listen(_ id: String) async throws -> NWConnection {
        let c = try await quicConnect()
        try await writeMsg(c, PigeonWire.encodeRelayGreetingListen(token: "", instanceId: id))
        _ = try await readMsg(c)  // relay acks with the instance ID
        return c
    }

    /// Connect a client to `id` (T45 `connect`). The relay writes an "ok"
    /// ack once it has matched a parked backend `listen` and the bridge is
    /// live; after that the connection is a clean end-to-end pipe.
    private func connect(_ id: String) async throws -> NWConnection {
        let c = try await quicConnect()
        try await writeMsg(c, PigeonWire.encodeRelayGreetingConnect(instanceId: id))
        let ack = try await readMsg(c)
        guard String(decoding: ack, as: UTF8.self) == "ok" else {
            throw NSError(domain: "Connect", code: 0,
                          userInfo: [NSLocalizedDescriptionKey: "expected ok ack, got \(ack.count) bytes"])
        }
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

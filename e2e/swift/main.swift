// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Standalone E2E test for Swift QUIC relay connectivity.
// Uses raw NWConnection (not the PigeonRelay wrapper) to prove the
// protocol works end-to-end. PigeonRelay wrapper tests are separate.
//
// Usage:
//   swift build && .build/debug/pigeon-e2e-swift                                # local (auto-starts relay)
//   PIGEON_RELAY_HOST=carrier-pigeon.fly.dev PIGEON_RELAY_PORT=4433 PIGEON_TOKEN=<t> \
//     .build/debug/pigeon-e2e-swift                                              # live relay

#if canImport(Network)

import Foundation
import Network
import Pigeon

// MARK: - Config

let relayHost = ProcessInfo.processInfo.environment["PIGEON_RELAY_HOST"] ?? "127.0.0.1"
let relayPort = UInt16(ProcessInfo.processInfo.environment["PIGEON_RELAY_PORT"] ?? "4433")!
let relayToken = ProcessInfo.processInfo.environment["PIGEON_TOKEN"]
let isLocal = relayHost == "127.0.0.1" || relayHost == "localhost"

// MARK: - QUIC helpers

func quicConnect(_ host: String, _ port: UInt16) async throws -> NWConnection {
    let opts = NWProtocolQUIC.Options(alpn: ["pigeon"])
    sec_protocol_options_set_verify_block(opts.securityProtocolOptions, { _, _, c in c(true) }, .main)
    let params = NWParameters(quic: opts)
    let ep = NWEndpoint.hostPort(host: .init(host), port: NWEndpoint.Port(rawValue: port)!)
    let q = DispatchQueue(label: "pigeon.\(arc4random())")
    let conn = NWConnection(to: ep, using: params)

    try await withThrowingTaskGroup(of: Void.self) { group in
        group.addTask {
            try await withCheckedThrowingContinuation { (c: CheckedContinuation<Void, Error>) in
                final class G: @unchecked Sendable { var done = false }
                let g = G()
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
            throw NSError(domain: "Timeout", code: 0, userInfo: [NSLocalizedDescriptionKey: "QUIC connect timeout"])
        }
        try await group.next()!
        group.cancelAll()
    }
    return conn
}

func writeMsg(_ c: NWConnection, _ payload: Data) async throws {
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

func readMsg(_ c: NWConnection) async throws -> Data {
    let hdr = try await readExact(c, 4)
    let b0 = UInt32(hdr[0]) << 24
    let b1 = UInt32(hdr[1]) << 16
    let b2 = UInt32(hdr[2]) << 8
    let b3 = UInt32(hdr[3])
    let len = Int(b0 | b1 | b2 | b3)
    if len == 0 { return Data() }
    return try await readExact(c, len)
}

func readExact(_ c: NWConnection, _ count: Int) async throws -> Data {
    try await withCheckedThrowingContinuation { cont in
        c.receive(minimumIncompleteLength: count, maximumLength: count) { d, _, _, e in
            if let e = e { cont.resume(throwing: e) }
            else if let d = d, d.count >= count { cont.resume(returning: d) }
            else { cont.resume(throwing: NSError(domain: "EOF", code: 0,
                userInfo: [NSLocalizedDescriptionKey: "expected \(count) bytes, got \(d?.count ?? 0)"])) }
        }
    }
}

// MARK: - Local relay

class LocalRelay {
    let process: Process
    let quicPort: UInt16

    static func start() throws -> LocalRelay {
        let repoRoot = URL(fileURLWithPath: #filePath)
            .deletingLastPathComponent().deletingLastPathComponent().deletingLastPathComponent()

        let build = Process()
        build.executableURL = URL(fileURLWithPath: "/usr/bin/env")
        build.arguments = ["go", "build", "-o", "/tmp/pigeon-e2e-server", "./cmd/pigeon"]
        build.currentDirectoryURL = repoRoot
        build.standardOutput = FileHandle.nullDevice
        build.standardError = FileHandle.nullDevice
        try build.run()
        build.waitUntilExit()
        guard build.terminationStatus == 0 else {
            throw NSError(domain: "Build", code: 1, userInfo: [NSLocalizedDescriptionKey: "go build failed"])
        }

        let port = findFreePort()
        let proc = Process()
        proc.executableURL = URL(fileURLWithPath: "/tmp/pigeon-e2e-server")
        proc.arguments = ["--quic-port", String(port)]
        proc.standardOutput = FileHandle.nullDevice
        let pipe = Pipe()
        proc.standardError = pipe
        try proc.run()

        // Wait for server ready.
        let deadline = Date().addingTimeInterval(10)
        var ready = false
        pipe.fileHandleForReading.readabilityHandler = { h in
            if let s = String(data: h.availableData, encoding: .utf8), s.contains("pigeon starting") { ready = true }
        }
        while !ready && Date() < deadline { Thread.sleep(forTimeInterval: 0.1) }
        pipe.fileHandleForReading.readabilityHandler = nil
        guard ready else { proc.terminate(); throw NSError(domain: "Server", code: 2, userInfo: [NSLocalizedDescriptionKey: "timeout"]) }

        return LocalRelay(process: proc, quicPort: port)
    }

    private init(process: Process, quicPort: UInt16) { self.process = process; self.quicPort = quicPort }
    func stop() { process.terminate(); process.waitUntilExit() }

    private static func findFreePort() -> UInt16 {
        let sock = socket(AF_INET, SOCK_DGRAM, 0)
        defer { close(sock) }
        var addr = sockaddr_in()
        addr.sin_family = sa_family_t(AF_INET)
        addr.sin_port = 0
        addr.sin_addr.s_addr = INADDR_ANY.bigEndian
        var len = socklen_t(MemoryLayout<sockaddr_in>.size)
        withUnsafePointer(to: &addr) { $0.withMemoryRebound(to: sockaddr.self, capacity: 1) { _ = bind(sock, $0, len) } }
        withUnsafeMutablePointer(to: &addr) { $0.withMemoryRebound(to: sockaddr.self, capacity: 1) { _ = getsockname(sock, $0, &len) } }
        return UInt16(bigEndian: addr.sin_port)
    }
}

// MARK: - Test runner

var passed = 0
var failed = 0

func test(_ name: String, _ body: @escaping () async throws -> Void) async {
    do {
        try await body()
        passed += 1
        print("[PASS] \(name)")
    } catch {
        failed += 1
        print("[FAIL] \(name): \(error.localizedDescription)")
    }
}

func register(_ host: String, _ port: UInt16) async throws -> (NWConnection, String) {
    let c = try await quicConnect(host, port)
    let hs = relayToken.map { "register:\($0)" } ?? "register"
    try await writeMsg(c, Data(hs.utf8))
    let id = String(decoding: try await readMsg(c), as: UTF8.self)
    return (c, id)
}

func connect(_ host: String, _ port: UInt16, _ id: String) async throws -> NWConnection {
    let c = try await quicConnect(host, port)
    try await writeMsg(c, Data("connect:\(id)".utf8))
    return c
}

// MARK: - Tests

func runTests(_ host: String, _ port: UInt16) async {
    await test("register") {
        let (c, id) = try await register(host, port)
        guard !id.isEmpty else { throw NSError(domain: "", code: 0, userInfo: [NSLocalizedDescriptionKey: "empty ID"]) }
        c.cancel()
    }

    await test("stream round-trip") {
        let (backend, id) = try await register(host, port)
        let client = try await connect(host, port, id)

        try await client.send("hello from swift", writeMsg)
        let msg = try await readMsg(backend)
        guard String(decoding: msg, as: UTF8.self) == "hello from swift" else {
            throw NSError(domain: "", code: 0, userInfo: [NSLocalizedDescriptionKey: "got '\(String(decoding: msg, as: UTF8.self))'"])
        }

        try await backend.send("reply from swift", writeMsg)
        let reply = try await readMsg(client)
        guard String(decoding: reply, as: UTF8.self) == "reply from swift" else {
            throw NSError(domain: "", code: 0, userInfo: [NSLocalizedDescriptionKey: "got '\(String(decoding: reply, as: UTF8.self))'"])
        }

        backend.cancel(); client.cancel()
    }

    await test("10 messages in order") {
        let (backend, id) = try await register(host, port)
        let client = try await connect(host, port, id)

        for i in 0..<10 { try await writeMsg(client, Data("msg-\(i)".utf8)) }
        for i in 0..<10 {
            let d = try await readMsg(backend)
            let t = String(decoding: d, as: UTF8.self)
            guard t == "msg-\(i)" else { throw NSError(domain: "", code: 0, userInfo: [NSLocalizedDescriptionKey: "msg \(i): got '\(t)'"]) }
        }
        backend.cancel(); client.cancel()
    }

    await test("encrypted round-trip") {
        let (backend, id) = try await register(host, port)
        let client = try await connect(host, port, id)

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
        guard decrypted == pt else { throw NSError(domain: "", code: 0, userInfo: [NSLocalizedDescriptionKey: "decrypt mismatch"]) }

        // Backend → client encrypted.
        let reply = Data("secret reply".utf8)
        try await writeMsg(backend, try bCh.encrypt(reply))
        let replyCt = try await readMsg(client)
        guard try cCh.decrypt(replyCt) == reply else { throw NSError(domain: "", code: 0, userInfo: [NSLocalizedDescriptionKey: "reply decrypt mismatch"]) }

        backend.cancel(); client.cancel()
    }

    await test("confirmation code cross-platform vector") {
        let code = deriveConfirmationCode(Data(repeating: 0x01, count: 32), Data(repeating: 0x02, count: 32))
        guard code == "629624" else { throw NSError(domain: "", code: 0, userInfo: [NSLocalizedDescriptionKey: "got '\(code)'"]) }
    }

    await test("confirmation codes match") {
        let a = E2EKeyPair(), b = E2EKeyPair()
        let codeA = deriveConfirmationCode(a.publicKeyData, b.publicKeyData)
        let codeB = deriveConfirmationCode(b.publicKeyData, a.publicKeyData)
        guard codeA == codeB else { throw NSError(domain: "", code: 0, userInfo: [NSLocalizedDescriptionKey: "\(codeA) != \(codeB)"]) }
        guard codeA.count == 6 else { throw NSError(domain: "", code: 0, userInfo: [NSLocalizedDescriptionKey: "not 6 digits"]) }
    }
}

// MARK: - Convenience

extension NWConnection {
    func send(_ text: String, _ writer: (NWConnection, Data) async throws -> Void) async throws {
        try await writer(self, Data(text.utf8))
    }
}

// MARK: - Entry point (RunLoop-based for Network.framework)

func main() {
    Task {
        let host: String
        let port: UInt16
        var relay: LocalRelay? = nil

        if isLocal {
            print("Starting local relay...")
            do {
                relay = try LocalRelay.start()
                host = "127.0.0.1"
                port = relay!.quicPort
                print("Relay on port \(port)")
            } catch {
                print("[FATAL] \(error)")
                exit(1)
            }
        } else {
            host = relayHost
            port = relayPort
            print("Using \(host):\(port)")
        }

        await runTests(host, port)

        relay?.stop()
        print("\n=== \(passed) passed, \(failed) failed ===")
        exit(failed > 0 ? 1 : 0)
    }
    dispatchMain()
}

main()

#else

func main() {
    print("Network.framework not available — skipping")
    exit(0)
}
main()

#endif

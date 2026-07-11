// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

import XCTest
@testable import Pigeon

final class E2ECryptoTests: XCTestCase {

    func testKeyExchangeAndEncrypt() throws {
        let client = E2EKeyPair()
        let server = E2EKeyPair()

        // Derive session keys — client sends, server receives.
        let clientSendKey = try client.deriveSessionKey(
            peerPublicKey: server.publicKeyData,
            info: Data("client-to-server".utf8)
        )
        let serverRecvKey = try server.deriveSessionKey(
            peerPublicKey: client.publicKeyData,
            info: Data("client-to-server".utf8)
        )

        // Reverse direction.
        let clientRecvKey = try client.deriveSessionKey(
            peerPublicKey: server.publicKeyData,
            info: Data("server-to-client".utf8)
        )
        let serverSendKey = try server.deriveSessionKey(
            peerPublicKey: client.publicKeyData,
            info: Data("server-to-client".utf8)
        )

        // Create channels.
        let clientCh = E2EChannel(sendKey: clientSendKey, recvKey: clientRecvKey)
        let serverCh = E2EChannel(sendKey: serverSendKey, recvKey: serverRecvKey)

        // Client → server.
        let msg = Data("hello from client".utf8)
        let ct = try clientCh.encrypt(msg)
        let pt = try serverCh.decrypt(ct)
        XCTAssertEqual(pt, msg)

        // Server → client.
        let msg2 = Data("hello from server".utf8)
        let ct2 = try serverCh.encrypt(msg2)
        let pt2 = try clientCh.decrypt(ct2)
        XCTAssertEqual(pt2, msg2)
    }

    func testSymmetricChannel() throws {
        let secret = Data((0..<32).map { _ in UInt8.random(in: 0...255) })

        let clientCh = E2EChannel(sharedKey: secret, isServer: false)
        let serverCh = E2EChannel(sharedKey: secret, isServer: true)

        // Client → server.
        let msg = Data(#"{"type":"auth","device":"abc123"}"#.utf8)
        let ct = try clientCh.encrypt(msg)
        let pt = try serverCh.decrypt(ct)
        XCTAssertEqual(pt, msg)

        // Server → client.
        let msg2 = Data(#"{"type":"auth_ok"}"#.utf8)
        let ct2 = try serverCh.encrypt(msg2)
        let pt2 = try clientCh.decrypt(ct2)
        XCTAssertEqual(pt2, msg2)

        // Multiple messages in sequence.
        for i in 0..<100 {
            let m = Data("message \(i)".utf8)
            let c = try clientCh.encrypt(m)
            let p = try serverCh.decrypt(c)
            XCTAssertEqual(p, m)
        }
    }

    func testReplayRejected() throws {
        let secret = Data((0..<32).map { _ in UInt8.random(in: 0...255) })
        let clientCh = E2EChannel(sharedKey: secret, isServer: false)
        let serverCh = E2EChannel(sharedKey: secret, isServer: true)

        let ct = try clientCh.encrypt(Data("first".utf8))
        _ = try serverCh.decrypt(ct)

        // Replay the same ciphertext — should fail.
        XCTAssertThrowsError(try serverCh.decrypt(ct))
    }

    func testCiphertextTooShort() throws {
        let secret = Data((0..<32).map { _ in UInt8.random(in: 0...255) })
        let ch = E2EChannel(sharedKey: secret, isServer: true)

        XCTAssertThrowsError(try ch.decrypt(Data([1, 2, 3])))
    }

    func testDatagramModeAcceptsGaps() throws {
        let secret = Data((0..<32).map { UInt8($0) })
        let sender = E2EChannel(sharedKey: secret, isServer: false)
        let receiver = E2EChannel(sharedKey: secret, isServer: true)
        receiver.mode = .datagrams

        // Encrypt seq 0..5
        var encrypted: [Data] = []
        for i in 0..<6 {
            encrypted.append(try sender.encrypt(Data("msg\(i)".utf8)))
        }

        // Decrypt 0, 1, 5 (skip 2, 3, 4)
        XCTAssertEqual(try receiver.decrypt(encrypted[0]), Data("msg0".utf8))
        XCTAssertEqual(try receiver.decrypt(encrypted[1]), Data("msg1".utf8))
        XCTAssertEqual(try receiver.decrypt(encrypted[5]), Data("msg5".utf8))
    }

    func testDatagramModeRejectsReplay() throws {
        let secret = Data((0..<32).map { UInt8($0) })
        let sender = E2EChannel(sharedKey: secret, isServer: false)
        let receiver = E2EChannel(sharedKey: secret, isServer: true)
        receiver.mode = .datagrams

        let ct = try sender.encrypt(Data("hello".utf8))
        _ = try receiver.decrypt(ct) // first time OK
        XCTAssertThrowsError(try receiver.decrypt(ct)) // replay rejected
    }

    func testDatagramModeRejectsOldSeq() throws {
        let secret = Data((0..<32).map { UInt8($0) })
        let sender = E2EChannel(sharedKey: secret, isServer: false)
        let receiver = E2EChannel(sharedKey: secret, isServer: true)
        receiver.mode = .datagrams

        // Encrypt seq 0..5
        var encrypted: [Data] = []
        for i in 0..<6 {
            encrypted.append(try sender.encrypt(Data("msg\(i)".utf8)))
        }

        // Accept seq 5 first
        _ = try receiver.decrypt(encrypted[5])

        // Seq 3 is now too old — should be rejected
        XCTAssertThrowsError(try receiver.decrypt(encrypted[3]))
    }

    func testPairingRecordRoundTrip() throws {
        // Simulate a pairing: two key pairs, record one side.
        let serverKP = E2EKeyPair()
        let clientKP = E2EKeyPair()

        let record = PairingRecord(
            peerInstanceID: "server-uuid",
            relayURL: "https://relay.example.com",
            localKeyPair: clientKP,
            peerPublicKey: serverKP.publicKeyData
        )

        // Marshal/unmarshal via JSON.
        let data = try JSONEncoder().encode(record)
        let restored = try JSONDecoder().decode(PairingRecord.self, from: data)

        // Derive channel from restored record.
        let ch = try restored.deriveChannel(
            sendInfo: Data("c2s".utf8),
            recvInfo: Data("s2c".utf8)
        )

        // Derive the same channel from the original keys (server side).
        let serverSendKey = try serverKP.deriveSessionKey(
            peerPublicKey: clientKP.publicKeyData,
            info: Data("s2c".utf8)
        )
        let serverRecvKey = try serverKP.deriveSessionKey(
            peerPublicKey: clientKP.publicKeyData,
            info: Data("c2s".utf8)
        )
        let serverCh = E2EChannel(sendKey: serverSendKey, recvKey: serverRecvKey)

        // Verify they can communicate.
        let ct = try ch.encrypt(Data("hello from restored record".utf8))
        let pt = try serverCh.decrypt(ct)
        XCTAssertEqual(String(data: pt, encoding: .utf8), "hello from restored record")
    }

    /// 🎯T50 oracle: two sessions under one PairingRecord with distinct
    /// session salts must produce distinct AEAD keys (reconnect must not
    /// reuse the prior key with the GCM counter reset to 0).
    func testDeriveChannelPerSessionSaltDistinct() throws {
        let localKP = E2EKeyPair()
        let peerKP = E2EKeyPair()
        let rec = PairingRecord(
            peerInstanceID: "peer-001",
            relayURL: "https://relay.example.com",
            localKeyPair: localKP,
            peerPublicKey: peerKP.publicKeyData)

        let salt1 = Data(repeating: 0x11, count: sessionSaltLen)
        let salt2 = Data(repeating: 0x22, count: sessionSaltLen)
        let plaintext = Data("the quick brown fox".utf8)

        let ch1 = try rec.deriveChannel(
            sendInfo: Data("client-to-server".utf8),
            recvInfo: Data("server-to-client".utf8),
            sessionSalt: salt1)
        let ch2 = try rec.deriveChannel(
            sendInfo: Data("client-to-server".utf8),
            recvInfo: Data("server-to-client".utf8),
            sessionSalt: salt2)

        let ct1 = try ch1.encrypt(plaintext)
        let ct2 = try ch2.encrypt(plaintext)
        // Leading 8 bytes are the (zero) sequence prefix; body must differ.
        XCTAssertNotEqual(ct1.dropFirst(8), ct2.dropFirst(8),
            "distinct salts produced identical keystream: per-session key derivation is broken")

        // Same salt is deterministic.
        let ch1b = try rec.deriveChannel(
            sendInfo: Data("client-to-server".utf8),
            recvInfo: Data("server-to-client".utf8),
            sessionSalt: salt1)
        let ct1b = try ch1b.encrypt(plaintext)
        XCTAssertEqual(ct1.dropFirst(8), ct1b.dropFirst(8),
            "same salt produced different keys")

        // Wrong-length salt is rejected.
        XCTAssertThrowsError(
            try rec.deriveChannel(
                sendInfo: Data("c2s".utf8),
                recvInfo: Data("s2c".utf8),
                sessionSalt: Data([0x01]))
        )
    }

    func testConfirmationCodeCrossplatformVector() {
        // Fixed X25519 public keys (any 32-byte value works for derivation).
        let keyA = Data(repeating: 0x01, count: 32)
        let keyB = Data(repeating: 0x02, count: 32)

        // This expected value must match Go, TypeScript, and Kotlin tests.
        let expected = "629624"
        XCTAssertEqual(deriveConfirmationCode(keyA, keyB), expected)

        // Verify order-independence.
        XCTAssertEqual(deriveConfirmationCode(keyB, keyA), expected)
    }
}

// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

import XCTest
@testable import TernCrypto

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

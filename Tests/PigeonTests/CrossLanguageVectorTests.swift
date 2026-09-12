// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

import XCTest
@testable import PigeonCore

/// Loads c/test/vectors.json and checks the Swift AEAD/HKDF stack
/// against the Go-generated vectors (ENT-003).
final class CrossLanguageVectorTests: XCTestCase {

    func testConfirmationCodeMatchesGoVector() throws {
        let v = try loadVectors()
        let code = deriveConfirmationCode(v.data("alice_public"), v.data("bob_public"))
        XCTAssertEqual(code, v.string("confirmation_code"))
    }

    func testECDHCiphertextDecrypts() throws {
        let v = try loadVectors()
        let record = PairingRecord(
            peerInstanceID: "bob",
            relayURL: "test",
            localPrivateKey: v.data("alice_private"),
            localPublicKey: v.data("alice_public"),
            peerPublicKey: v.data("bob_public")
        )
        let ch = try record.deriveChannel(
            sendInfo: Data("alice-to-bob".utf8),
            recvInfo: Data("bob-to-alice".utf8)
        )
        // The recorded ciphertext is alice→bob, so bob's recv side
        // is the inverse record. Decrypt from bob's view.
        let bob = PairingRecord(
            peerInstanceID: "alice",
            relayURL: "test",
            localPrivateKey: v.data("bob_private"),
            localPublicKey: v.data("bob_public"),
            peerPublicKey: v.data("alice_public")
        )
        let bobCh = try bob.deriveChannel(
            sendInfo: Data("bob-to-alice".utf8),
            recvInfo: Data("alice-to-bob".utf8)
        )
        let pt = try bobCh.decrypt(v.data("ciphertext"))
        XCTAssertEqual(pt, Data(v.string("plaintext").utf8))
        // And a fresh encrypt from alice matches the recorded first frame.
        let ct = try ch.encrypt(Data(v.string("plaintext").utf8))
        XCTAssertEqual(ct, v.data("ciphertext"))
    }

    func testSymmetricCiphertextsMatch() throws {
        let v = try loadVectors()
        let master = v.data("symmetric_master")
        let client = E2EChannel(sharedKey: master, isServer: false)
        let server = E2EChannel(sharedKey: master, isServer: true)
        let pt = Data(v.string("symmetric_plaintext").utf8)
        XCTAssertEqual(try server.decrypt(v.data("symmetric_ciphertext_c2s")), pt)
        XCTAssertEqual(try client.decrypt(v.data("symmetric_ciphertext_s2c")), pt)
        XCTAssertEqual(try client.encrypt(pt), v.data("symmetric_ciphertext_c2s"))
        XCTAssertEqual(try server.encrypt(pt), v.data("symmetric_ciphertext_s2c"))
    }

    private struct Vectors {
        let raw: [String: String]
        func string(_ k: String) throws -> String {
            guard let s = raw[k] else { throw XCTSkip("missing vector \(k)") }
            return s
        }
        func data(_ k: String) throws -> Data {
            let s = try string(k)
            guard let d = Data(hexString: s) else {
                return Data(s.utf8)
            }
            return d
        }
    }

    private func loadVectors() throws -> Vectors {
        var dir = URL(fileURLWithPath: #filePath)
        for _ in 0..<8 {
            dir.deleteLastPathComponent()
            let candidate = dir.appendingPathComponent("c/test/vectors.json")
            if FileManager.default.fileExists(atPath: candidate.path) {
                let data = try Data(contentsOf: candidate)
                let raw = try JSONDecoder().decode([String: String].self, from: data)
                return Vectors(raw: raw)
            }
        }
        throw XCTSkip("c/test/vectors.json not found above \(#filePath)")
    }
}

private extension Data {
    init?(hexString: String) {
        let hex = hexString
        guard hex.count % 2 == 0, !hex.isEmpty, hex.allSatisfy(\.isHexDigit) else {
            return nil
        }
        var bytes = [UInt8]()
        bytes.reserveCapacity(hex.count / 2)
        var idx = hex.startIndex
        while idx < hex.endIndex {
            let next = hex.index(idx, offsetBy: 2)
            guard let b = UInt8(hex[idx..<next], radix: 16) else { return nil }
            bytes.append(b)
            idx = next
        }
        self.init(bytes)
    }
}

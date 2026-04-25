// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

import XCTest
@testable import Pigeon

final class PairingArtifactTests: XCTestCase {

    private func makeArtifact(ttl: TimeInterval = defaultPairingTTL) throws -> (PairingArtifact, Date) {
        let kp = E2EKeyPair()
        let peerKP = E2EKeyPair()
        let record = PairingRecord(
            peerInstanceID: "inst-abc",
            relayURL: "https://relay.example.com",
            localKeyPair: kp,
            peerPublicKey: peerKP.publicKeyData)
        let issued = Date(timeIntervalSince1970: 1_777_000_000)  // deterministic
        var a = PairingArtifact(record: record, token: "tok-xyz", issuedAt: issued, ttl: ttl)
        // re-pin token in case ttl<=0 zeroed expiry
        a.token = "tok-xyz"
        return (a, issued)
    }

    func testJSONRoundTrip() throws {
        let (a, issued) = try makeArtifact()
        let data = try a.toJSON()
        let restored = try PairingArtifact.fromJSON(data)

        XCTAssertEqual(restored.token, "tok-xyz")
        XCTAssertEqual(restored.issuedAt, issued)
        XCTAssertEqual(restored.expiresAt, issued.addingTimeInterval(defaultPairingTTL))
        XCTAssertEqual(restored.record.peerInstanceID, "inst-abc")
        XCTAssertEqual(restored.record.peerPublicKey, a.record.peerPublicKey)
    }

    func testTextRoundTrip() throws {
        let (a, _) = try makeArtifact()
        let text = try a.toText()
        XCTAssertFalse(text.contains("="))
        XCTAssertFalse(text.contains("+"))
        XCTAssertFalse(text.contains("/"))

        let restored = try PairingArtifact.fromText(text)
        XCTAssertEqual(restored.record.peerInstanceID, "inst-abc")
        XCTAssertEqual(restored.token, "tok-xyz")
    }

    func testIsExpired() throws {
        let (a, issued) = try makeArtifact(ttl: 24 * 60 * 60)
        XCTAssertFalse(a.isExpired(now: issued))
        XCTAssertFalse(a.isExpired(now: issued.addingTimeInterval(60 * 60)))
        XCTAssertTrue(a.isExpired(now: issued.addingTimeInterval(24 * 60 * 60)))
        XCTAssertTrue(a.isExpired(now: issued.addingTimeInterval(48 * 60 * 60)))
    }

    func testZeroTTLNeverExpires() throws {
        let kp = E2EKeyPair()
        let peerKP = E2EKeyPair()
        let record = PairingRecord(
            peerInstanceID: "i",
            relayURL: "https://r",
            localKeyPair: kp,
            peerPublicKey: peerKP.publicKeyData)
        let a = PairingArtifact(record: record, token: "", issuedAt: Date(), ttl: 0)
        XCTAssertNil(a.expiresAt)
        XCTAssertFalse(a.isExpired(now: Date(timeIntervalSinceNow: 1_000_000_000)))
    }

    func testWireFormatMatchesGoSnakeCase() throws {
        // The Go SDK emits snake_case keys. Verify Swift produces the
        // same shape.
        let (a, _) = try makeArtifact()
        let data = try a.toJSON()
        let object = try JSONSerialization.jsonObject(with: data) as? [String: Any]
        XCTAssertNotNil(object?["record"])
        XCTAssertNotNil(object?["issued_at"])
        XCTAssertNotNil(object?["expires_at"])
        let record = object?["record"] as? [String: Any]
        XCTAssertNotNil(record?["peer_instance_id"])
        XCTAssertNotNil(record?["relay_url"])
        XCTAssertNotNil(record?["local_private_key"])
        XCTAssertNotNil(record?["local_public_key"])
        XCTAssertNotNil(record?["peer_public_key"])
    }

    func testFileCredentialStoreRoundTrip() throws {
        let dir = FileManager.default.temporaryDirectory
            .appendingPathComponent(UUID().uuidString)
        defer { try? FileManager.default.removeItem(at: dir) }
        let store = FileCredentialStore(url: dir.appendingPathComponent("artifact.json"))

        XCTAssertThrowsError(try store.load())
        XCTAssertThrowsError(try store.isExpired())

        let (a, _) = try makeArtifact()
        try store.save(a)
        let restored = try store.load()
        XCTAssertEqual(restored.token, "tok-xyz")
        XCTAssertEqual(restored.record.peerInstanceID, "inst-abc")

        try store.delete()
        XCTAssertThrowsError(try store.load())
    }

    func testConnectWithArtifactRejectsExpired() async throws {
        let kp = E2EKeyPair()
        let peerKP = E2EKeyPair()
        let record = PairingRecord(
            peerInstanceID: "inst",
            relayURL: "https://relay.example.com",
            localKeyPair: kp,
            peerPublicKey: peerKP.publicKeyData)
        let stale = PairingArtifact(
            record: record,
            token: "",
            issuedAt: Date(timeIntervalSinceNow: -31 * 24 * 60 * 60),
            ttl: defaultPairingTTL)
        XCTAssertTrue(stale.isExpired())
        do {
            _ = try await PigeonConn.connect(artifact: stale)
            XCTFail("expected PairingError.expired")
        } catch PairingError.expired {
            // expected
        } catch {
            XCTFail("unexpected error: \(error)")
        }
    }

    func testFileCredentialStoreIsExpired() throws {
        let dir = FileManager.default.temporaryDirectory
            .appendingPathComponent(UUID().uuidString)
        defer { try? FileManager.default.removeItem(at: dir) }
        let store = FileCredentialStore(url: dir.appendingPathComponent("artifact.json"))

        let (a, issued) = try makeArtifact(ttl: 24 * 60 * 60)
        try store.save(a)

        store.now = { issued.addingTimeInterval(60 * 60) }
        XCTAssertFalse(try store.isExpired())

        store.now = { issued.addingTimeInterval(48 * 60 * 60) }
        XCTAssertTrue(try store.isExpired())
    }
}

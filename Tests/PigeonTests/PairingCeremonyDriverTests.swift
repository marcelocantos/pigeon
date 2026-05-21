// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// In-process pairing-ceremony driver tests. Swift acceptor ↔ Swift
// initiator over an in-memory paired transport; verifies that the
// generated PairingCeremonyAcceptor/InitiatorMachine sequence the
// hello/welcome/confirm exchange end-to-end, that both sides derive
// identical 6-digit codes, and that the resulting PairingRecords yield
// AEAD channels that can decrypt each other's ciphertext.
//
// This is the Swift counterpart to pairing/pairing_test.go::
// TestCeremonyEndToEnd (which exercises the Go runAcceptor /
// runInitiator over a real relay).

import XCTest
@testable import Pigeon
import CryptoKit
import Foundation

final class PairingCeremonyDriverTests: XCTestCase {

    func testSwiftAcceptorMeetsSwiftInitiator() async throws {
        let (aTr, bTr) = InMemoryPairingTransport.pair()

        // Acceptor side: ephemeral X25519 + identity.
        let aEph = E2EKeyPair()
        let aId = E2EKeyPair()
        let aIdentity = PairingIdentity(
            publicKey: aId.publicKeyData,
            instanceID: "acceptor-instance"
        )

        // Initiator side.
        let bEph = E2EKeyPair()
        let bId = E2EKeyPair()
        let bIdentity = PairingIdentity(
            publicKey: bId.publicKeyData,
            instanceID: "initiator-instance"
        )

        // Run both halves concurrently — pairAcceptor blocks on the
        // initiator's hello, pairInitiator blocks on the acceptor's welcome.
        async let acceptorResult = pairAcceptor(
            transport: aTr,
            identity: aIdentity,
            acceptorEphPriv: Data(aEph.privateKey.rawRepresentation),
            acceptorEphPub: aEph.publicKeyData,
            relayURL: "https://relay.example.com",
            confirm: { _ in true }
        )
        async let initiatorResult = pairInitiator(
            transport: bTr,
            identity: bIdentity,
            initiatorEphPriv: Data(bEph.privateKey.rawRepresentation),
            initiatorEphPub: bEph.publicKeyData,
            acceptorEphPub: aEph.publicKeyData,
            acceptorInstance: aIdentity.instanceID,
            relayURL: "https://relay.example.com",
            confirm: { _ in true }
        )

        let aRes = try await acceptorResult
        let bRes = try await initiatorResult

        XCTAssertEqual(aRes.code, bRes.code, "confirmation codes must agree")
        XCTAssertEqual(aRes.code.count, 6, "code must be 6 digits")

        // Each side records the other's instance ID and ephemeral pubkey.
        XCTAssertEqual(aRes.record.peerInstanceID, bIdentity.instanceID)
        XCTAssertEqual(bRes.record.peerInstanceID, aIdentity.instanceID)
        XCTAssertEqual(aRes.record.peerPublicKey, bEph.publicKeyData)
        XCTAssertEqual(bRes.record.peerPublicKey, aEph.publicKeyData)

        // DeriveChannel interop: encrypt on acceptor, decrypt on initiator
        // (and vice-versa). Using the same info-label convention as Go's
        // TestCeremonyEndToEnd in pairing/pairing_test.go.
        let aChan = try aRes.record.deriveChannel(
            sendInfo: Data("a->b".utf8),
            recvInfo: Data("b->a".utf8)
        )
        let bChan = try bRes.record.deriveChannel(
            sendInfo: Data("b->a".utf8),
            recvInfo: Data("a->b".utf8)
        )

        let pt = Data("hello over a derived channel".utf8)
        let ct = try aChan.encrypt(pt)
        let dec = try bChan.decrypt(ct)
        XCTAssertEqual(dec, pt, "a→b round-trip via PairingRecord-derived channels")

        let pt2 = Data("reply over a derived channel".utf8)
        let ct2 = try bChan.encrypt(pt2)
        let dec2 = try aChan.decrypt(ct2)
        XCTAssertEqual(dec2, pt2, "b→a round-trip via PairingRecord-derived channels")
    }

    func testInitiatorRejectsWelcomeWithMismatchedAcceptorEph() async throws {
        // If the welcome's eph_pub disagrees with the rendezvous token's
        // acceptor_eph_pub, the initiator must abort — this is the
        // pre-derivation MitM check.
        let (aTr, bTr) = InMemoryPairingTransport.pair()
        let aEph = E2EKeyPair()
        let bEph = E2EKeyPair()
        let attackerEph = E2EKeyPair()  // claims aEph in token, sends attackerEph in welcome

        // Spawn a fake acceptor that responds with attackerEph's pubkey
        // instead of aEph (the value the token committed to).
        Task {
            // Wait for hello.
            let helloBytes = (try? await aTr.recv()) ?? Data()
            _ = helloBytes
            // Send fraudulent welcome.
            let fake = #"{"kind":"welcome","eph_pub":"\#(attackerEph.publicKeyData.base64EncodedString())","identity_pub":"\#(E2EKeyPair().publicKeyData.base64EncodedString())","instance_id":"acceptor"}"#
            try? await aTr.send(Data(fake.utf8))
        }

        let bId = E2EKeyPair()
        do {
            _ = try await pairInitiator(
                transport: bTr,
                identity: PairingIdentity(publicKey: bId.publicKeyData,
                                          instanceID: "initiator"),
                initiatorEphPriv: Data(bEph.privateKey.rawRepresentation),
                initiatorEphPub: bEph.publicKeyData,
                acceptorEphPub: aEph.publicKeyData,
                acceptorInstance: "acceptor",
                relayURL: "",
                confirm: { _ in true }
            )
            XCTFail("expected pairInitiator to reject welcome with mismatched eph_pub")
        } catch let err as PairingCeremonyError {
            if case .malformedMessage = err { return }
            XCTFail("unexpected error: \(err)")
        } catch {
            XCTFail("unexpected error: \(error)")
        }
    }

    func testTokenRoundTrip() throws {
        // Encode a token in the Go pairing.tokenPayload shape, then
        // decode it via PairingToken.decode and check field-by-field.
        let acceptorEph = E2EKeyPair()
        let acceptorId = E2EKeyPair()
        let payload: [String: Any] = [
            "relay": "https://relay.example.com",
            "session_id": "sess-xyz",
            "acceptor_eph_pub": acceptorEph.publicKeyData.base64EncodedString(),
            "acceptor_id_pub": acceptorId.publicKeyData.base64EncodedString(),
            "acceptor_instance": "acceptor-instance-id",
        ]
        let json = try JSONSerialization.data(withJSONObject: payload, options: [])
        // base64url-no-pad — what pairing.Pairer.Accept does.
        let token = json.base64EncodedString()
            .replacingOccurrences(of: "+", with: "-")
            .replacingOccurrences(of: "/", with: "_")
            .replacingOccurrences(of: "=", with: "")

        let decoded = try PairingToken.decode(token)
        XCTAssertEqual(decoded.relay, "https://relay.example.com")
        XCTAssertEqual(decoded.sessionID, "sess-xyz")
        XCTAssertEqual(decoded.acceptorInstance, "acceptor-instance-id")
        XCTAssertEqual(decoded.acceptorEphPub, acceptorEph.publicKeyData)
        XCTAssertEqual(decoded.acceptorIdPub, acceptorId.publicKeyData)
    }
}

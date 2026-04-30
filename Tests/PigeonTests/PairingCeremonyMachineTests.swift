// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Unit tests for the regenerated pairing-ceremony state machines.
// These exercise the spec produced by `cmd/protogen protocol/pairing.yaml`
// and so verify that the Swift, Kotlin, Go, C, TS, and TLA+ outputs all
// agree on state names, transition order, and action wiring.

import XCTest
@testable import Pigeon

typealias Acceptor = PairingCeremonyAcceptorMachine
typealias Initiator = PairingCeremonyInitiatorMachine

final class PairingCeremonyMachineTests: XCTestCase {

    // MARK: - acceptor

    func testAcceptorInitialStateIsIdle() {
        let m = Acceptor()
        XCTAssertEqual(m.state, .idle)
    }

    func testAcceptorHappyPath() throws {
        let m = Acceptor()
        var fired: [Acceptor.ActionID] = []
        let track: (Acceptor.ActionID) -> () throws -> Void = { id in
            { fired.append(id) }
        }
        m.actions[.genEphemeral]    = track(.genEphemeral)
        m.actions[.registerRelay]   = track(.registerRelay)
        m.actions[.emitToken]       = track(.emitToken)
        m.actions[.deriveCode]      = track(.deriveCode)
        m.actions[.storeRecord]     = track(.storeRecord)

        try m.handleEvent(.pairBegin)
        XCTAssertEqual(m.state, .generatingEphemeral)
        try m.handleEvent(.ephemeralReady)
        XCTAssertEqual(m.state, .registeringRelay)
        try m.handleEvent(.relayRegistered)
        XCTAssertEqual(m.state, .waitingForHello)
        try m.handleEvent(.recvHello)
        XCTAssertEqual(m.state, .derivingCode)
        try m.handleEvent(.codeReady)
        XCTAssertEqual(m.state, .awaitingUserConfirm)
        try m.handleEvent(.userConfirm)
        XCTAssertEqual(m.state, .awaitingPeerConfirm)
        try m.handleEvent(.recvConfirmToAcceptor)
        XCTAssertEqual(m.state, .paired)

        // Actions fire in the order the spec dictates: gen_ephemeral
        // (Idle→GeneratingEphemeral), register_relay (→RegisteringRelay),
        // emit_token (→WaitingForHello), derive_code (on recv hello),
        // store_record (on recv confirm_to_acceptor).
        XCTAssertEqual(fired, [
            .genEphemeral, .registerRelay, .emitToken,
            .deriveCode, .storeRecord,
        ])
    }

    func testAcceptorUserCancelFromAwaitingUserConfirm() throws {
        let m = Acceptor()
        m.actions[.genEphemeral]  = {}
        m.actions[.registerRelay] = {}
        m.actions[.emitToken]     = {}
        m.actions[.deriveCode]    = {}

        try m.handleEvent(.pairBegin)
        try m.handleEvent(.ephemeralReady)
        try m.handleEvent(.relayRegistered)
        try m.handleEvent(.recvHello)
        try m.handleEvent(.codeReady)
        XCTAssertEqual(m.state, .awaitingUserConfirm)

        try m.handleEvent(.userCancel)
        XCTAssertEqual(m.state, .aborted)
    }

    func testAcceptorUserCancelFromAwaitingPeerConfirm() throws {
        let m = Acceptor()
        m.actions[.genEphemeral]  = {}
        m.actions[.registerRelay] = {}
        m.actions[.emitToken]     = {}
        m.actions[.deriveCode]    = {}

        try m.handleEvent(.pairBegin)
        try m.handleEvent(.ephemeralReady)
        try m.handleEvent(.relayRegistered)
        try m.handleEvent(.recvHello)
        try m.handleEvent(.codeReady)
        try m.handleEvent(.userConfirm)
        XCTAssertEqual(m.state, .awaitingPeerConfirm)

        try m.handleEvent(.userCancel)
        XCTAssertEqual(m.state, .aborted)
    }

    func testAcceptorActionThrowsPropagates() {
        let m = Acceptor()
        struct Boom: Error {}
        m.actions[.genEphemeral] = { throw Boom() }
        XCTAssertThrowsError(try m.handleEvent(.pairBegin)) { err in
            XCTAssertTrue(err is Boom)
        }
    }

    // MARK: - initiator

    func testInitiatorInitialStateIsIdle() {
        let m = Initiator()
        XCTAssertEqual(m.state, .idle)
    }

    func testInitiatorHappyPath() throws {
        let m = Initiator()
        var fired: [Initiator.ActionID] = []
        let track: (Initiator.ActionID) -> () throws -> Void = { id in
            { fired.append(id) }
        }
        m.actions[.decodeToken]   = track(.decodeToken)
        m.actions[.genEphemeral]  = track(.genEphemeral)
        m.actions[.dialRelay]     = track(.dialRelay)
        m.actions[.deriveCode]    = track(.deriveCode)
        m.actions[.storeRecord]   = track(.storeRecord)

        try m.handleEvent(.tokenReceived)
        XCTAssertEqual(m.state, .decodingToken)
        try m.handleEvent(.tokenDecoded)
        XCTAssertEqual(m.state, .generatingEphemeral)
        try m.handleEvent(.ephemeralReady)
        XCTAssertEqual(m.state, .connectingRelay)
        try m.handleEvent(.relayConnected)
        XCTAssertEqual(m.state, .awaitingWelcome)
        try m.handleEvent(.recvWelcome)
        XCTAssertEqual(m.state, .derivingCode)
        try m.handleEvent(.codeReady)
        XCTAssertEqual(m.state, .awaitingUserConfirm)
        try m.handleEvent(.userConfirm)
        XCTAssertEqual(m.state, .awaitingPeerConfirm)
        try m.handleEvent(.recvConfirmToInitiator)
        XCTAssertEqual(m.state, .paired)

        // decode_token (Idle→DecodingToken), gen_ephemeral
        // (DecodingToken→GeneratingEphemeral), dial_relay
        // (GeneratingEphemeral→ConnectingRelay), derive_code
        // (on recv welcome), store_record (on recv confirm).
        XCTAssertEqual(fired, [
            .decodeToken, .genEphemeral, .dialRelay, .deriveCode, .storeRecord,
        ])
    }

    func testInitiatorUserCancel() throws {
        let m = Initiator()
        m.actions[.decodeToken]  = {}
        m.actions[.genEphemeral] = {}
        m.actions[.deriveCode]   = {}

        try m.handleEvent(.tokenReceived)
        try m.handleEvent(.tokenDecoded)
        try m.handleEvent(.ephemeralReady)
        try m.handleEvent(.relayConnected)
        try m.handleEvent(.recvWelcome)
        try m.handleEvent(.codeReady)
        XCTAssertEqual(m.state, .awaitingUserConfirm)

        try m.handleEvent(.userCancel)
        XCTAssertEqual(m.state, .aborted)
    }

    // MARK: - protocol surface

    func testMessageTypesMatchYAML() {
        XCTAssertEqual(PairingCeremonyProtocol.MessageType.hello.rawValue,              "hello")
        XCTAssertEqual(PairingCeremonyProtocol.MessageType.welcome.rawValue,            "welcome")
        XCTAssertEqual(PairingCeremonyProtocol.MessageType.confirmToInitiator.rawValue, "confirm_to_initiator")
        XCTAssertEqual(PairingCeremonyProtocol.MessageType.confirmToAcceptor.rawValue,  "confirm_to_acceptor")
    }

    func testActionIDsMatchYAML() {
        XCTAssertEqual(PairingCeremonyProtocol.ActionID.genEphemeral.rawValue,  "gen_ephemeral")
        XCTAssertEqual(PairingCeremonyProtocol.ActionID.registerRelay.rawValue, "register_relay")
        XCTAssertEqual(PairingCeremonyProtocol.ActionID.emitToken.rawValue,     "emit_token")
        XCTAssertEqual(PairingCeremonyProtocol.ActionID.deriveCode.rawValue,    "derive_code")
        XCTAssertEqual(PairingCeremonyProtocol.ActionID.storeRecord.rawValue,   "store_record")
        XCTAssertEqual(PairingCeremonyProtocol.ActionID.decodeToken.rawValue,   "decode_token")
        XCTAssertEqual(PairingCeremonyProtocol.ActionID.dialRelay.rawValue,     "dial_relay")
    }
}

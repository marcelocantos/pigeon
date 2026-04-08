// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

import XCTest
@testable import Pigeon

final class PairingCeremonyMachineTests: XCTestCase {

    // MARK: - Server Machine

    func testServerInitialState() {
        let m = PairingCeremonyServerMachine()
        XCTAssertEqual(m.state, .idle)
    }

    func testServerIdleToGenerateToken() throws {
        let m = PairingCeremonyServerMachine()
        var actionCalled = false
        m.actions[.generateToken] = { actionCalled = true }

        let cmds = try m.handleEvent(.recvPairBegin)
        XCTAssertEqual(m.state, .generateToken)
        XCTAssertTrue(actionCalled)
        XCTAssertEqual(cmds, [])
        XCTAssertEqual(m.currentToken, "tok_1")
    }

    func testServerGenerateTokenToRegisterRelay() throws {
        let m = PairingCeremonyServerMachine()
        m.actions[.generateToken] = {}
        try m.handleEvent(.recvPairBegin)

        var actionCalled = false
        m.actions[.registerRelay] = { actionCalled = true }

        let cmds = try m.handleEvent(.tokenCreated)
        XCTAssertEqual(m.state, .registerRelay)
        XCTAssertTrue(actionCalled)
        XCTAssertEqual(cmds, [])
    }

    func testServerRegisterRelayToWaitingForClient() throws {
        let m = serverAtState(.registerRelay)
        let cmds = try m.handleEvent(.relayRegistered)
        XCTAssertEqual(m.state, .waitingForClient)
        XCTAssertEqual(cmds, [])
    }

    func testServerTokenValidGuardAllowsTransition() throws {
        let m = serverAtState(.waitingForClient)
        m.guards[.tokenValid] = { true }
        m.guards[.tokenInvalid] = { false }

        var actionCalled = false
        m.actions[.deriveSecret] = { actionCalled = true }

        let cmds = try m.handleEvent(.recvPairHello)
        XCTAssertEqual(m.state, .deriveSecret)
        XCTAssertTrue(actionCalled)
        XCTAssertEqual(cmds, [])
        XCTAssertEqual(m.serverEcdhPub, "server_pub")
    }

    func testServerTokenInvalidGuardResetsToIdle() throws {
        let m = serverAtState(.waitingForClient)
        m.guards[.tokenValid] = { false }
        m.guards[.tokenInvalid] = { true }

        let cmds = try m.handleEvent(.recvPairHello)
        XCTAssertEqual(m.state, .idle)
        XCTAssertEqual(cmds, [])
    }

    func testServerCodeCorrectGuard() throws {
        let m = serverAtState(.validateCode)
        m.guards[.codeCorrect] = { true }
        m.guards[.codeWrong] = { false }

        let cmds = try m.handleEvent(.checkCode)
        XCTAssertEqual(m.state, .storePaired)
        XCTAssertEqual(cmds, [])
    }

    func testServerCodeWrongGuardResetsToIdle() throws {
        let m = serverAtState(.validateCode)
        m.guards[.codeCorrect] = { false }
        m.guards[.codeWrong] = { true }

        let cmds = try m.handleEvent(.checkCode)
        XCTAssertEqual(m.state, .idle)
        XCTAssertEqual(cmds, [])
    }

    func testServerFinaliseStoresDevice() throws {
        let m = serverAtState(.storePaired)
        var actionCalled = false
        m.actions[.storeDevice] = { actionCalled = true }

        let cmds = try m.handleEvent(.finalise)
        XCTAssertEqual(m.state, .paired)
        XCTAssertTrue(actionCalled)
        XCTAssertEqual(cmds, [])
        XCTAssertEqual(m.deviceSecret, "dev_secret_1")
    }

    func testServerDeviceKnownGuard() throws {
        let m = serverAtState(.authCheck)
        m.guards[.deviceKnown] = { true }
        m.guards[.deviceUnknown] = { false }
        var actionCalled = false
        m.actions[.verifyDevice] = { actionCalled = true }

        let cmds = try m.handleEvent(.verify)
        XCTAssertEqual(m.state, .sessionActive)
        XCTAssertTrue(actionCalled)
        XCTAssertEqual(cmds, [])
    }

    func testServerDeviceUnknownGuardResetsToIdle() throws {
        let m = serverAtState(.authCheck)
        m.guards[.deviceKnown] = { false }
        m.guards[.deviceUnknown] = { true }

        let cmds = try m.handleEvent(.verify)
        XCTAssertEqual(m.state, .idle)
        XCTAssertEqual(cmds, [])
    }

    func testServerDisconnectReturnsToPaired() throws {
        let m = serverAtState(.sessionActive)
        let cmds = try m.handleEvent(.disconnect)
        XCTAssertEqual(m.state, .paired)
        XCTAssertEqual(cmds, [])
    }

    func testServerInvalidEventNoStateChange() throws {
        let m = PairingCeremonyServerMachine()
        let cmds = try m.handleEvent(.disconnect) // invalid from Idle
        XCTAssertEqual(m.state, .idle)
        XCTAssertEqual(cmds, [])
    }

    func testServerFullPairingFlow() throws {
        let m = PairingCeremonyServerMachine()
        m.actions[.generateToken] = {}
        m.actions[.registerRelay] = {}
        m.actions[.deriveSecret] = {}
        m.actions[.storeDevice] = {}
        m.guards[.tokenValid] = { true }
        m.guards[.tokenInvalid] = { false }
        m.guards[.codeCorrect] = { true }
        m.guards[.codeWrong] = { false }

        try m.handleEvent(.recvPairBegin)
        XCTAssertEqual(m.state, .generateToken)
        try m.handleEvent(.tokenCreated)
        XCTAssertEqual(m.state, .registerRelay)
        try m.handleEvent(.relayRegistered)
        XCTAssertEqual(m.state, .waitingForClient)
        try m.handleEvent(.recvPairHello)
        XCTAssertEqual(m.state, .deriveSecret)
        try m.handleEvent(.eCDHComplete)
        XCTAssertEqual(m.state, .sendAck)
        try m.handleEvent(.signalCodeDisplay)
        XCTAssertEqual(m.state, .waitingForCode)
        try m.handleEvent(.recvCodeSubmit)
        XCTAssertEqual(m.state, .validateCode)
        try m.handleEvent(.checkCode)
        XCTAssertEqual(m.state, .storePaired)
        try m.handleEvent(.finalise)
        XCTAssertEqual(m.state, .paired)
    }

    // MARK: - iOS Machine

    func testIosInitialState() {
        let m = PairingCeremonyIosMachine()
        XCTAssertEqual(m.state, .idle)
    }

    func testIosFullPairingFlow() throws {
        let m = PairingCeremonyIosMachine()
        m.actions[.sendPairHello] = {}
        m.actions[.deriveSecret] = {}
        m.actions[.storeSecret] = {}

        try m.handleEvent(.userScansQR)
        XCTAssertEqual(m.state, .scanQR)
        try m.handleEvent(.qRParsed)
        XCTAssertEqual(m.state, .connectRelay)
        try m.handleEvent(.relayConnected)
        XCTAssertEqual(m.state, .genKeyPair)
        try m.handleEvent(.keyPairGenerated)
        XCTAssertEqual(m.state, .waitAck)
        try m.handleEvent(.recvPairHelloAck)
        XCTAssertEqual(m.state, .e2EReady)
        try m.handleEvent(.recvPairConfirm)
        XCTAssertEqual(m.state, .showCode)
        try m.handleEvent(.codeDisplayed)
        XCTAssertEqual(m.state, .waitPairComplete)
        try m.handleEvent(.recvPairComplete)
        XCTAssertEqual(m.state, .paired)
    }

    func testIosKeyPairGeneratedCallsAction() throws {
        let m = PairingCeremonyIosMachine()
        try m.handleEvent(.userScansQR)
        try m.handleEvent(.qRParsed)
        try m.handleEvent(.relayConnected)

        var actionCalled = false
        m.actions[.sendPairHello] = { actionCalled = true }

        try m.handleEvent(.keyPairGenerated)
        XCTAssertTrue(actionCalled)
        XCTAssertEqual(m.state, .waitAck)
    }

    func testIosReconnectAndAuth() throws {
        let m = PairingCeremonyIosMachine()
        // Fast-forward to Paired via step/handleEvent
        m.actions[.sendPairHello] = {}
        m.actions[.deriveSecret] = {}
        m.actions[.storeSecret] = {}

        try m.handleEvent(.userScansQR)
        try m.handleEvent(.qRParsed)
        try m.handleEvent(.relayConnected)
        try m.handleEvent(.keyPairGenerated)
        try m.handleEvent(.recvPairHelloAck)
        try m.handleEvent(.recvPairConfirm)
        try m.handleEvent(.codeDisplayed)
        try m.handleEvent(.recvPairComplete)
        XCTAssertEqual(m.state, .paired)

        try m.handleEvent(.appLaunch)
        XCTAssertEqual(m.state, .reconnect)
        try m.handleEvent(.relayConnected)
        XCTAssertEqual(m.state, .sendAuth)
        try m.handleEvent(.recvAuthOk)
        XCTAssertEqual(m.state, .sessionActive)
        try m.handleEvent(.disconnect)
        XCTAssertEqual(m.state, .paired)
    }

    func testIosInvalidEventNoStateChange() throws {
        let m = PairingCeremonyIosMachine()
        let cmds = try m.handleEvent(.disconnect) // invalid from Idle
        XCTAssertEqual(m.state, .idle)
        XCTAssertEqual(cmds, [])
    }

    // MARK: - CLI Machine

    func testCliInitialState() {
        let m = PairingCeremonyCliMachine()
        XCTAssertEqual(m.state, .idle)
    }

    func testCliFullFlow() throws {
        let m = PairingCeremonyCliMachine()

        try m.handleEvent(.cliInit)
        XCTAssertEqual(m.state, .getKey)
        try m.handleEvent(.keyStored)
        XCTAssertEqual(m.state, .beginPair)
        try m.handleEvent(.recvTokenResponse)
        XCTAssertEqual(m.state, .showQR)
        try m.handleEvent(.recvWaitingForCode)
        XCTAssertEqual(m.state, .promptCode)
        try m.handleEvent(.userEntersCode)
        XCTAssertEqual(m.state, .submitCode)
        try m.handleEvent(.recvPairStatus)
        XCTAssertEqual(m.state, .done)
    }

    func testCliInvalidEventNoStateChange() throws {
        let m = PairingCeremonyCliMachine()
        let cmds = try m.handleEvent(.keyStored) // invalid from Idle
        XCTAssertEqual(m.state, .idle)
        XCTAssertEqual(cmds, [])
    }

    // MARK: - Helpers

    private func serverAtState(_ target: PairingCeremonyServerState) -> PairingCeremonyServerMachine {
        let m = PairingCeremonyServerMachine()
        m.actions[.generateToken] = {}
        m.actions[.registerRelay] = {}
        m.actions[.deriveSecret] = {}
        m.actions[.storeDevice] = {}
        m.actions[.verifyDevice] = {}
        m.guards[.tokenValid] = { true }
        m.guards[.tokenInvalid] = { false }
        m.guards[.codeCorrect] = { true }
        m.guards[.codeWrong] = { false }
        m.guards[.deviceKnown] = { true }
        m.guards[.deviceUnknown] = { false }

        // Walk the machine to the target state.
        let path: [PairingCeremonyProtocol.EventID]
        switch target {
        case .idle: path = []
        case .generateToken: path = [.recvPairBegin]
        case .registerRelay: path = [.recvPairBegin, .tokenCreated]
        case .waitingForClient: path = [.recvPairBegin, .tokenCreated, .relayRegistered]
        case .deriveSecret: path = [.recvPairBegin, .tokenCreated, .relayRegistered, .recvPairHello]
        case .sendAck: path = [.recvPairBegin, .tokenCreated, .relayRegistered, .recvPairHello, .eCDHComplete]
        case .waitingForCode: path = [.recvPairBegin, .tokenCreated, .relayRegistered, .recvPairHello, .eCDHComplete, .signalCodeDisplay]
        case .validateCode: path = [.recvPairBegin, .tokenCreated, .relayRegistered, .recvPairHello, .eCDHComplete, .signalCodeDisplay, .recvCodeSubmit]
        case .storePaired: path = [.recvPairBegin, .tokenCreated, .relayRegistered, .recvPairHello, .eCDHComplete, .signalCodeDisplay, .recvCodeSubmit, .checkCode]
        case .paired: path = [.recvPairBegin, .tokenCreated, .relayRegistered, .recvPairHello, .eCDHComplete, .signalCodeDisplay, .recvCodeSubmit, .checkCode, .finalise]
        case .authCheck: path = [.recvPairBegin, .tokenCreated, .relayRegistered, .recvPairHello, .eCDHComplete, .signalCodeDisplay, .recvCodeSubmit, .checkCode, .finalise, .recvAuthRequest]
        case .sessionActive: path = [.recvPairBegin, .tokenCreated, .relayRegistered, .recvPairHello, .eCDHComplete, .signalCodeDisplay, .recvCodeSubmit, .checkCode, .finalise, .recvAuthRequest, .verify]
        }
        for ev in path {
            try! m.handleEvent(ev)
        }
        return m
    }
}

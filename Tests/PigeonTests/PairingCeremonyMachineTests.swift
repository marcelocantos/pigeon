// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

import XCTest
@testable import Pigeon

final class PairingCeremonyMachineTests: XCTestCase {

    // MARK: - Server Pairing Sub-Machine

    func testServerPairingInitialState() {
        let m = PairingCeremonyServerPairingMachine()
        XCTAssertEqual(m.state, .idle)
    }

    func testServerPairingIdleToGenerateToken() throws {
        let m = PairingCeremonyServerPairingMachine()
        var actionCalled = false
        m.actions[.generateToken] = { actionCalled = true }

        let cmds = try m.handleEvent(.recvPairBegin)
        XCTAssertEqual(m.state, .generateToken)
        XCTAssertTrue(actionCalled)
        XCTAssertEqual(cmds, [])
    }

    func testServerPairingGenerateTokenToRegisterRelay() throws {
        let m = PairingCeremonyServerPairingMachine()
        m.actions[.generateToken] = {}
        try m.handleEvent(.recvPairBegin)

        var actionCalled = false
        m.actions[.registerRelay] = { actionCalled = true }

        let cmds = try m.handleEvent(.tokenCreated)
        XCTAssertEqual(m.state, .registerRelay)
        XCTAssertTrue(actionCalled)
        XCTAssertEqual(cmds, [])
    }

    func testServerPairingRegisterRelayToWaitingForClient() throws {
        let m = serverPairingAtState(.registerRelay)
        let cmds = try m.handleEvent(.relayRegistered)
        XCTAssertEqual(m.state, .waitingForClient)
        XCTAssertEqual(cmds, [])
    }

    func testServerPairingTokenValidGuardAllowsTransition() throws {
        let m = serverPairingAtState(.waitingForClient)
        m.guards[.tokenValid] = { true }
        m.guards[.tokenInvalid] = { false }

        var actionCalled = false
        m.actions[.deriveSecret] = { actionCalled = true }

        let cmds = try m.handleEvent(.recvPairHello)
        XCTAssertEqual(m.state, .deriveSecret)
        XCTAssertTrue(actionCalled)
        XCTAssertEqual(cmds, [])
    }

    func testServerPairingTokenInvalidGuardResetsToIdle() throws {
        let m = serverPairingAtState(.waitingForClient)
        m.guards[.tokenValid] = { false }
        m.guards[.tokenInvalid] = { true }

        let cmds = try m.handleEvent(.recvPairHello)
        XCTAssertEqual(m.state, .idle)
        XCTAssertEqual(cmds, [])
    }

    func testServerPairingCodeCorrectGuard() throws {
        let m = serverPairingAtState(.validateCode)
        m.guards[.codeCorrect] = { true }
        m.guards[.codeWrong] = { false }

        let cmds = try m.handleEvent(.checkCode)
        XCTAssertEqual(m.state, .storePaired)
        XCTAssertEqual(cmds, [])
    }

    func testServerPairingCodeWrongGuardResetsToIdle() throws {
        let m = serverPairingAtState(.validateCode)
        m.guards[.codeCorrect] = { false }
        m.guards[.codeWrong] = { true }

        let cmds = try m.handleEvent(.checkCode)
        XCTAssertEqual(m.state, .idle)
        XCTAssertEqual(cmds, [])
    }

    func testServerPairingFinaliseStoresDevice() throws {
        let m = serverPairingAtState(.storePaired)
        var actionCalled = false
        m.actions[.storeDevice] = { actionCalled = true }

        let cmds = try m.handleEvent(.finalise)
        XCTAssertEqual(m.state, .pairingComplete)
        XCTAssertTrue(actionCalled)
        XCTAssertEqual(cmds, [])
    }

    func testServerPairingInvalidEventNoStateChange() throws {
        let m = PairingCeremonyServerPairingMachine()
        let cmds = try m.handleEvent(.disconnect) // invalid from Idle
        XCTAssertEqual(m.state, .idle)
        XCTAssertEqual(cmds, [])
    }

    func testServerPairingFullFlow() throws {
        let m = PairingCeremonyServerPairingMachine()
        m.actions[.generateToken] = {}
        m.actions[.registerRelay] = {}
        m.actions[.deriveSecret] = {}
        m.actions[.storeDevice] = {}
        m.guards[.tokenValid] = { true }
        m.guards[.tokenInvalid] = { false }
        m.guards[.codeCorrect] = { true }
        m.guards[.codeWrong] = { false }

        try m.handleEvent(.recvPairBegin);     XCTAssertEqual(m.state, .generateToken)
        try m.handleEvent(.tokenCreated);      XCTAssertEqual(m.state, .registerRelay)
        try m.handleEvent(.relayRegistered);   XCTAssertEqual(m.state, .waitingForClient)
        try m.handleEvent(.recvPairHello);     XCTAssertEqual(m.state, .deriveSecret)
        try m.handleEvent(.eCDHComplete);      XCTAssertEqual(m.state, .sendAck)
        try m.handleEvent(.signalCodeDisplay); XCTAssertEqual(m.state, .waitingForCode)
        try m.handleEvent(.recvCodeSubmit);    XCTAssertEqual(m.state, .validateCode)
        try m.handleEvent(.checkCode);         XCTAssertEqual(m.state, .storePaired)
        try m.handleEvent(.finalise);          XCTAssertEqual(m.state, .pairingComplete)
    }

    // MARK: - Server Auth Sub-Machine

    func testServerAuthInitialState() {
        let m = PairingCeremonyServerAuthMachine()
        XCTAssertEqual(m.state, .idle)
    }

    func testServerAuthCredentialReadyMovesToPaired() throws {
        let m = PairingCeremonyServerAuthMachine()
        let cmds = try m.handleEvent(.credentialReady)
        XCTAssertEqual(m.state, .paired)
        XCTAssertEqual(cmds, [])
    }

    func testServerAuthRequestMovesToAuthCheck() throws {
        let m = serverAuthAtState(.paired)
        let cmds = try m.handleEvent(.recvAuthRequest)
        XCTAssertEqual(m.state, .authCheck)
        XCTAssertEqual(cmds, [])
    }

    func testServerAuthDeviceKnownGuard() throws {
        let m = serverAuthAtState(.authCheck)
        m.guards[.deviceKnown] = { true }
        m.guards[.deviceUnknown] = { false }
        var actionCalled = false
        m.actions[.verifyDevice] = { actionCalled = true }

        let cmds = try m.handleEvent(.verify)
        XCTAssertEqual(m.state, .sessionActive)
        XCTAssertTrue(actionCalled)
        XCTAssertEqual(cmds, [])
    }

    func testServerAuthDeviceUnknownGuardResetsToIdle() throws {
        let m = serverAuthAtState(.authCheck)
        m.guards[.deviceKnown] = { false }
        m.guards[.deviceUnknown] = { true }

        let cmds = try m.handleEvent(.verify)
        XCTAssertEqual(m.state, .idle)
        XCTAssertEqual(cmds, [])
    }

    func testServerAuthDisconnectReturnsToPaired() throws {
        let m = serverAuthAtState(.sessionActive)
        let cmds = try m.handleEvent(.disconnect)
        XCTAssertEqual(m.state, .paired)
        XCTAssertEqual(cmds, [])
    }

    // MARK: - Server Composite (cross-sub-machine routing)

    func testServerCompositeInitialState() {
        let c = PairingCeremonyServerComposite()
        XCTAssertEqual(c.pairing.state, .idle)
        XCTAssertEqual(c.auth.state, .idle)
    }

    func testServerCompositeRoutePairedTriggersCredentialReady() throws {
        let c = PairingCeremonyServerComposite()
        XCTAssertEqual(c.auth.state, .idle)
        let routed = try c.route(from: "pairing", event: .paired)
        XCTAssertTrue(routed)
        XCTAssertEqual(c.auth.state, .paired)
    }

    func testServerCompositeUnknownRouteIsNoOp() throws {
        let c = PairingCeremonyServerComposite()
        let routed = try c.route(from: "auth", event: .paired)
        XCTAssertFalse(routed)
        XCTAssertEqual(c.auth.state, .idle)
    }

    func testServerCompositeFullPairThenAuthFlow() throws {
        let c = PairingCeremonyServerComposite()
        c.pairing.actions[.generateToken] = {}
        c.pairing.actions[.registerRelay] = {}
        c.pairing.actions[.deriveSecret] = {}
        c.pairing.actions[.storeDevice] = {}
        c.pairing.guards[.tokenValid] = { true }
        c.pairing.guards[.codeCorrect] = { true }
        c.auth.actions[.verifyDevice] = {}
        c.auth.guards[.deviceKnown] = { true }

        // Walk pairing sub-machine to PairingComplete.
        try c.pairing.handleEvent(.recvPairBegin)
        try c.pairing.handleEvent(.tokenCreated)
        try c.pairing.handleEvent(.relayRegistered)
        try c.pairing.handleEvent(.recvPairHello)
        try c.pairing.handleEvent(.eCDHComplete)
        try c.pairing.handleEvent(.signalCodeDisplay)
        try c.pairing.handleEvent(.recvCodeSubmit)
        try c.pairing.handleEvent(.checkCode)
        try c.pairing.handleEvent(.finalise)
        XCTAssertEqual(c.pairing.state, .pairingComplete)

        // Route fires the credential_ready into auth.
        try c.route(from: "pairing", event: .paired)
        XCTAssertEqual(c.auth.state, .paired)

        // Auth flow.
        try c.auth.handleEvent(.recvAuthRequest)
        XCTAssertEqual(c.auth.state, .authCheck)
        try c.auth.handleEvent(.verify)
        XCTAssertEqual(c.auth.state, .sessionActive)
        try c.auth.handleEvent(.disconnect)
        XCTAssertEqual(c.auth.state, .paired)
    }

    // MARK: - iOS Pairing Sub-Machine

    func testIosPairingInitialState() {
        let m = PairingCeremonyIosPairingMachine()
        XCTAssertEqual(m.state, .idle)
    }

    func testIosPairingFullFlow() throws {
        let m = PairingCeremonyIosPairingMachine()
        m.actions[.sendPairHello] = {}
        m.actions[.deriveSecret] = {}
        m.actions[.storeSecret] = {}

        try m.handleEvent(.userScansQR);       XCTAssertEqual(m.state, .scanQR)
        try m.handleEvent(.qRParsed);          XCTAssertEqual(m.state, .connectRelay)
        try m.handleEvent(.relayConnected);    XCTAssertEqual(m.state, .genKeyPair)
        try m.handleEvent(.keyPairGenerated);  XCTAssertEqual(m.state, .waitAck)
        try m.handleEvent(.recvPairHelloAck);  XCTAssertEqual(m.state, .e2EReady)
        try m.handleEvent(.recvPairConfirm);   XCTAssertEqual(m.state, .showCode)
        try m.handleEvent(.codeDisplayed);     XCTAssertEqual(m.state, .waitPairComplete)
        try m.handleEvent(.recvPairComplete);  XCTAssertEqual(m.state, .pairingComplete)
    }

    func testIosPairingKeyPairGeneratedCallsAction() throws {
        let m = PairingCeremonyIosPairingMachine()
        try m.handleEvent(.userScansQR)
        try m.handleEvent(.qRParsed)
        try m.handleEvent(.relayConnected)

        var actionCalled = false
        m.actions[.sendPairHello] = { actionCalled = true }

        try m.handleEvent(.keyPairGenerated)
        XCTAssertTrue(actionCalled)
        XCTAssertEqual(m.state, .waitAck)
    }

    func testIosPairingInvalidEventNoStateChange() throws {
        let m = PairingCeremonyIosPairingMachine()
        let cmds = try m.handleEvent(.disconnect) // invalid from Idle
        XCTAssertEqual(m.state, .idle)
        XCTAssertEqual(cmds, [])
    }

    // MARK: - iOS Auth Sub-Machine

    func testIosAuthInitialState() {
        let m = PairingCeremonyIosAuthMachine()
        XCTAssertEqual(m.state, .idle)
    }

    func testIosAuthReconnectAndAuthFlow() throws {
        let m = PairingCeremonyIosAuthMachine()
        try m.handleEvent(.credentialReady); XCTAssertEqual(m.state, .paired)
        try m.handleEvent(.appLaunch);       XCTAssertEqual(m.state, .reconnect)
        try m.handleEvent(.relayConnected);  XCTAssertEqual(m.state, .sendAuth)
        try m.handleEvent(.recvAuthOk);      XCTAssertEqual(m.state, .sessionActive)
        try m.handleEvent(.disconnect);      XCTAssertEqual(m.state, .paired)
    }

    // MARK: - iOS Composite

    func testIosCompositeRoutePairedTriggersCredentialReady() throws {
        let c = PairingCeremonyIosComposite()
        XCTAssertEqual(c.auth.state, .idle)
        let routed = try c.route(from: "pairing", event: .paired)
        XCTAssertTrue(routed)
        XCTAssertEqual(c.auth.state, .paired)
    }

    func testIosCompositeFullPairThenAuthFlow() throws {
        let c = PairingCeremonyIosComposite()
        c.pairing.actions[.sendPairHello] = {}
        c.pairing.actions[.deriveSecret] = {}
        c.pairing.actions[.storeSecret] = {}

        try c.pairing.handleEvent(.userScansQR)
        try c.pairing.handleEvent(.qRParsed)
        try c.pairing.handleEvent(.relayConnected)
        try c.pairing.handleEvent(.keyPairGenerated)
        try c.pairing.handleEvent(.recvPairHelloAck)
        try c.pairing.handleEvent(.recvPairConfirm)
        try c.pairing.handleEvent(.codeDisplayed)
        try c.pairing.handleEvent(.recvPairComplete)
        XCTAssertEqual(c.pairing.state, .pairingComplete)

        try c.route(from: "pairing", event: .paired)
        XCTAssertEqual(c.auth.state, .paired)

        try c.auth.handleEvent(.appLaunch)
        XCTAssertEqual(c.auth.state, .reconnect)
        try c.auth.handleEvent(.relayConnected)
        XCTAssertEqual(c.auth.state, .sendAuth)
        try c.auth.handleEvent(.recvAuthOk)
        XCTAssertEqual(c.auth.state, .sessionActive)
        try c.auth.handleEvent(.disconnect)
        XCTAssertEqual(c.auth.state, .paired)
    }

    // MARK: - CLI Machine (no decomposition)

    func testCliInitialState() {
        let m = PairingCeremonyCliMachine()
        XCTAssertEqual(m.state, .idle)
    }

    func testCliFullFlow() throws {
        let m = PairingCeremonyCliMachine()

        try m.handleEvent(.cliInit);            XCTAssertEqual(m.state, .getKey)
        try m.handleEvent(.keyStored);          XCTAssertEqual(m.state, .beginPair)
        try m.handleEvent(.recvTokenResponse);  XCTAssertEqual(m.state, .showQR)
        try m.handleEvent(.recvWaitingForCode); XCTAssertEqual(m.state, .promptCode)
        try m.handleEvent(.userEntersCode);     XCTAssertEqual(m.state, .submitCode)
        try m.handleEvent(.recvPairStatus);     XCTAssertEqual(m.state, .done)
    }

    func testCliInvalidEventNoStateChange() throws {
        let m = PairingCeremonyCliMachine()
        let cmds = try m.handleEvent(.keyStored) // invalid from Idle
        XCTAssertEqual(m.state, .idle)
        XCTAssertEqual(cmds, [])
    }

    // MARK: - Helpers

    private func serverPairingAtState(_ target: PairingCeremonyServerPairingState) -> PairingCeremonyServerPairingMachine {
        let m = PairingCeremonyServerPairingMachine()
        m.actions[.generateToken] = {}
        m.actions[.registerRelay] = {}
        m.actions[.deriveSecret] = {}
        m.actions[.storeDevice] = {}
        m.guards[.tokenValid] = { true }
        m.guards[.tokenInvalid] = { false }
        m.guards[.codeCorrect] = { true }
        m.guards[.codeWrong] = { false }

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
        case .pairingComplete: path = [.recvPairBegin, .tokenCreated, .relayRegistered, .recvPairHello, .eCDHComplete, .signalCodeDisplay, .recvCodeSubmit, .checkCode, .finalise]
        }
        for ev in path {
            try! m.handleEvent(ev)
        }
        return m
    }

    private func serverAuthAtState(_ target: PairingCeremonyServerAuthState) -> PairingCeremonyServerAuthMachine {
        let m = PairingCeremonyServerAuthMachine()
        m.actions[.verifyDevice] = {}
        m.guards[.deviceKnown] = { true }
        m.guards[.deviceUnknown] = { false }

        let path: [PairingCeremonyProtocol.EventID]
        switch target {
        case .idle: path = []
        case .paired: path = [.credentialReady]
        case .authCheck: path = [.credentialReady, .recvAuthRequest]
        case .sessionActive: path = [.credentialReady, .recvAuthRequest, .verify]
        }
        for ev in path {
            try! m.handleEvent(ev)
        }
        return m
    }
}

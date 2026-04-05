// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Auto-generated from protocol definition. Do not edit.
// Source of truth: protocol/*.yaml

import Foundation

public enum MessageType: String, Sendable {
    case pairBegin = "pair_begin"
    case tokenResponse = "token_response"
    case pairHello = "pair_hello"
    case pairHelloAck = "pair_hello_ack"
    case pairConfirm = "pair_confirm"
    case waitingForCode = "waiting_for_code"
    case codeSubmit = "code_submit"
    case pairComplete = "pair_complete"
    case pairStatus = "pair_status"
    case authRequest = "auth_request"
    case authOk = "auth_ok"
}

public enum ServerState: String, Sendable {
    case idle = "Idle"
    case generateToken = "GenerateToken"
    case registerRelay = "RegisterRelay"
    case waitingForClient = "WaitingForClient"
    case deriveSecret = "DeriveSecret"
    case sendAck = "SendAck"
    case waitingForCode = "WaitingForCode"
    case validateCode = "ValidateCode"
    case storePaired = "StorePaired"
    case paired = "Paired"
    case authCheck = "AuthCheck"
    case sessionActive = "SessionActive"
}

public enum IosState: String, Sendable {
    case idle = "Idle"
    case scanQR = "ScanQR"
    case connectRelay = "ConnectRelay"
    case genKeyPair = "GenKeyPair"
    case waitAck = "WaitAck"
    case e2EReady = "E2EReady"
    case showCode = "ShowCode"
    case waitPairComplete = "WaitPairComplete"
    case paired = "Paired"
    case reconnect = "Reconnect"
    case sendAuth = "SendAuth"
    case sessionActive = "SessionActive"
}

public enum CliState: String, Sendable {
    case idle = "Idle"
    case getKey = "GetKey"
    case beginPair = "BeginPair"
    case showQR = "ShowQR"
    case promptCode = "PromptCode"
    case submitCode = "SubmitCode"
    case done = "Done"
}

public enum GuardID: String, Sendable {
    case tokenValid = "token_valid"
    case tokenInvalid = "token_invalid"
    case codeCorrect = "code_correct"
    case codeWrong = "code_wrong"
    case deviceKnown = "device_known"
    case deviceUnknown = "device_unknown"
    case nonceFresh = "nonce_fresh"
}

public enum ActionID: String, Sendable {
    case generateToken = "generate_token"
    case registerRelay = "register_relay"
    case deriveSecret = "derive_secret"
    case storeDevice = "store_device"
    case verifyDevice = "verify_device"
    case sendPairHello = "send_pair_hello"
    case storeSecret = "store_secret"
}

public enum EventID: String, Sendable {
    case tokenCreated = "token created"
    case relayRegistered = "relay registered"
    case eCDHComplete = "ECDH complete"
    case signalCodeDisplay = "signal code display"
    case checkCode = "check code"
    case finalise = "finalise"
    case verify = "verify"
    case disconnect = "disconnect"
    case userScansQR = "user scans QR"
    case qRParsed = "QR parsed"
    case relayConnected = "relay connected"
    case keyPairGenerated = "key pair generated"
    case codeDisplayed = "code displayed"
    case appLaunch = "app launch"
    case cliInit = "cli --init"
    case keyStored = "key stored"
    case userEntersCode = "user enters code"
    case recvPairBegin = "recv_pair_begin"
    case recvPairHello = "recv_pair_hello"
    case recvCodeSubmit = "recv_code_submit"
    case recvAuthRequest = "recv_auth_request"
    case recvPairHelloAck = "recv_pair_hello_ack"
    case recvPairConfirm = "recv_pair_confirm"
    case recvPairComplete = "recv_pair_complete"
    case recvAuthOk = "recv_auth_ok"
    case recvTokenResponse = "recv_token_response"
    case recvWaitingForCode = "recv_waiting_for_code"
    case recvPairStatus = "recv_pair_status"
}

/// The protocol transition table. Fed to Machine for execution.
public enum PairingCeremonyProtocol {

    /// server transitions.
    public static let serverInitial: ServerState = .idle

    public static let serverTransitions: [(from: String, to: String, on: String, onKind: String, guard: String?, action: String?, sends: [(to: String, msg: String)])] = [
        (from: "Idle", to: "GenerateToken", on: "pair_begin", onKind: "recv", guard: nil, action: "generate_token", sends: []),
        (from: "GenerateToken", to: "RegisterRelay", on: "token created", onKind: "internal", guard: nil, action: "register_relay", sends: []),
        (from: "RegisterRelay", to: "WaitingForClient", on: "relay registered", onKind: "internal", guard: nil, action: nil, sends: [(to: "cli", msg: "token_response")]),
        (from: "WaitingForClient", to: "DeriveSecret", on: "pair_hello", onKind: "recv", guard: "token_valid", action: "derive_secret", sends: []),
        (from: "WaitingForClient", to: "Idle", on: "pair_hello", onKind: "recv", guard: "token_invalid", action: nil, sends: []),
        (from: "DeriveSecret", to: "SendAck", on: "ECDH complete", onKind: "internal", guard: nil, action: nil, sends: [(to: "ios", msg: "pair_hello_ack")]),
        (from: "SendAck", to: "WaitingForCode", on: "signal code display", onKind: "internal", guard: nil, action: nil, sends: [(to: "ios", msg: "pair_confirm"), (to: "cli", msg: "waiting_for_code")]),
        (from: "WaitingForCode", to: "ValidateCode", on: "code_submit", onKind: "recv", guard: nil, action: nil, sends: []),
        (from: "ValidateCode", to: "StorePaired", on: "check code", onKind: "internal", guard: "code_correct", action: nil, sends: []),
        (from: "ValidateCode", to: "Idle", on: "check code", onKind: "internal", guard: "code_wrong", action: nil, sends: []),
        (from: "StorePaired", to: "Paired", on: "finalise", onKind: "internal", guard: nil, action: "store_device", sends: [(to: "ios", msg: "pair_complete"), (to: "cli", msg: "pair_status")]),
        (from: "Paired", to: "AuthCheck", on: "auth_request", onKind: "recv", guard: nil, action: nil, sends: []),
        (from: "AuthCheck", to: "SessionActive", on: "verify", onKind: "internal", guard: "device_known", action: "verify_device", sends: [(to: "ios", msg: "auth_ok")]),
        (from: "AuthCheck", to: "Idle", on: "verify", onKind: "internal", guard: "device_unknown", action: nil, sends: []),
        (from: "SessionActive", to: "Paired", on: "disconnect", onKind: "internal", guard: nil, action: nil, sends: []),
    ]

    /// ios transitions.
    public static let iosInitial: IosState = .idle

    public static let iosTransitions: [(from: String, to: String, on: String, onKind: String, guard: String?, action: String?, sends: [(to: String, msg: String)])] = [
        (from: "Idle", to: "ScanQR", on: "user scans QR", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "ScanQR", to: "ConnectRelay", on: "QR parsed", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "ConnectRelay", to: "GenKeyPair", on: "relay connected", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "GenKeyPair", to: "WaitAck", on: "key pair generated", onKind: "internal", guard: nil, action: "send_pair_hello", sends: [(to: "server", msg: "pair_hello")]),
        (from: "WaitAck", to: "E2EReady", on: "pair_hello_ack", onKind: "recv", guard: nil, action: "derive_secret", sends: []),
        (from: "E2EReady", to: "ShowCode", on: "pair_confirm", onKind: "recv", guard: nil, action: nil, sends: []),
        (from: "ShowCode", to: "WaitPairComplete", on: "code displayed", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "WaitPairComplete", to: "Paired", on: "pair_complete", onKind: "recv", guard: nil, action: "store_secret", sends: []),
        (from: "Paired", to: "Reconnect", on: "app launch", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "Reconnect", to: "SendAuth", on: "relay connected", onKind: "internal", guard: nil, action: nil, sends: [(to: "server", msg: "auth_request")]),
        (from: "SendAuth", to: "SessionActive", on: "auth_ok", onKind: "recv", guard: nil, action: nil, sends: []),
        (from: "SessionActive", to: "Paired", on: "disconnect", onKind: "internal", guard: nil, action: nil, sends: []),
    ]

    /// cli transitions.
    public static let cliInitial: CliState = .idle

    public static let cliTransitions: [(from: String, to: String, on: String, onKind: String, guard: String?, action: String?, sends: [(to: String, msg: String)])] = [
        (from: "Idle", to: "GetKey", on: "cli --init", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "GetKey", to: "BeginPair", on: "key stored", onKind: "internal", guard: nil, action: nil, sends: [(to: "server", msg: "pair_begin")]),
        (from: "BeginPair", to: "ShowQR", on: "token_response", onKind: "recv", guard: nil, action: nil, sends: []),
        (from: "ShowQR", to: "PromptCode", on: "waiting_for_code", onKind: "recv", guard: nil, action: nil, sends: []),
        (from: "PromptCode", to: "SubmitCode", on: "user enters code", onKind: "internal", guard: nil, action: nil, sends: [(to: "server", msg: "code_submit")]),
        (from: "SubmitCode", to: "Done", on: "pair_status", onKind: "recv", guard: nil, action: nil, sends: []),
    ]
}

/// ServerMachine is the generated state machine for the server actor.
public final class ServerMachine: @unchecked Sendable {
    public private(set) var state: ServerState
    public var currentToken: String // pairing token currently in play
    public var activeTokens: String // set of valid (non-revoked) tokens
    public var usedTokens: String // set of revoked tokens
    public var serverEcdhPub: String // server ECDH public key
    public var receivedClientPub: String // pubkey server received in pair_hello (may be adversary's)
    public var serverSharedKey: String // ECDH key derived by server (tuple to match DeriveKey output type)
    public var serverCode: String // code computed by server from its view of the pubkeys (tuple to match DeriveCode output type)
    public var receivedCode: String // code received in code_submit (tuple to match DeriveCode output type)
    public var codeAttempts: String // failed code submission attempts
    public var deviceSecret: String // persistent device secret
    public var pairedDevices: String // device IDs that completed pairing
    public var receivedDeviceId: String // device_id from auth_request
    public var authNoncesUsed: String // set of consumed auth nonces
    public var receivedAuthNonce: String // nonce from auth_request

    public var guards: [GuardID: () -> Bool] = [:]
    public var actions: [ActionID: () throws -> Void] = [:]

    public init() {
        self.state = .idle
        self.currentToken = "none"
        self.activeTokens = ""
        self.usedTokens = ""
        self.serverEcdhPub = "none"
        self.receivedClientPub = "none"
        self.serverSharedKey = ""
        self.serverCode = ""
        self.receivedCode = ""
        self.codeAttempts = 0
        self.deviceSecret = "none"
        self.pairedDevices = ""
        self.receivedDeviceId = "none"
        self.authNoncesUsed = ""
        self.receivedAuthNonce = "none"
    }

    /// Handle any event (message receipt or internal). Returns emitted commands.
    @discardableResult
    public func handleEvent(_ ev: EventID) throws -> [CmdID] {
        switch (state, ev) {
        case (.idle, .recvPairBegin):
            try actions[.generateToken]?()
            currentToken = "tok_1"
            // active_tokens: active_tokens \union {"tok_1"} (set by action)
            state = .generateToken
            return []
        case (.generateToken, .tokenCreated):
            try actions[.registerRelay]?()
            state = .registerRelay
            return []
        case (.registerRelay, .relayRegistered):
            state = .waitingForClient
            return []
        case (.waitingForClient, .recvPairHello) where guards[.tokenValid]?() == true:
            try actions[.deriveSecret]?()
            // received_client_pub: recv_msg.pubkey (set by action)
            serverEcdhPub = "server_pub"
            // server_shared_key: DeriveKey("server_pub", recv_msg.pubkey) (set by action)
            // server_code: DeriveCode("server_pub", recv_msg.pubkey) (set by action)
            state = .deriveSecret
            return []
        case (.waitingForClient, .recvPairHello) where guards[.tokenInvalid]?() == true:
            state = .idle
            return []
        case (.deriveSecret, .eCDHComplete):
            state = .sendAck
            return []
        case (.sendAck, .signalCodeDisplay):
            state = .waitingForCode
            return []
        case (.waitingForCode, .recvCodeSubmit):
            // received_code: recv_msg.code (set by action)
            state = .validateCode
            return []
        case (.validateCode, .checkCode) where guards[.codeCorrect]?() == true:
            state = .storePaired
            return []
        case (.validateCode, .checkCode) where guards[.codeWrong]?() == true:
            // code_attempts: code_attempts + 1 (set by action)
            state = .idle
            return []
        case (.storePaired, .finalise):
            try actions[.storeDevice]?()
            deviceSecret = "dev_secret_1"
            // paired_devices: paired_devices \union {"device_1"} (set by action)
            // active_tokens: active_tokens \ {current_token} (set by action)
            // used_tokens: used_tokens \union {current_token} (set by action)
            state = .paired
            return []
        case (.paired, .recvAuthRequest):
            // received_device_id: recv_msg.device_id (set by action)
            // received_auth_nonce: recv_msg.nonce (set by action)
            state = .authCheck
            return []
        case (.authCheck, .verify) where guards[.deviceKnown]?() == true:
            try actions[.verifyDevice]?()
            // auth_nonces_used: auth_nonces_used \union {received_auth_nonce} (set by action)
            state = .sessionActive
            return []
        case (.authCheck, .verify) where guards[.deviceUnknown]?() == true:
            state = .idle
            return []
        case (.sessionActive, .disconnect):
            state = .paired
            return []
        default:
            return []
        }
    }

    /// Process a received message. Returns the new state, or nil if rejected.
    @discardableResult
    public func handleMessage(_ msg: MessageType) throws -> ServerState? {
        switch (state, msg) {
        case (.idle, .pairBegin):
            try actions[.generateToken]?()
            currentToken = "tok_1"
            // active_tokens: active_tokens \union {"tok_1"} (set by action)
            state = .generateToken
            return state
        case (.waitingForClient, .pairHello) where guards[.tokenValid]?() == true:
            try actions[.deriveSecret]?()
            // received_client_pub: recv_msg.pubkey (set by action)
            serverEcdhPub = "server_pub"
            // server_shared_key: DeriveKey("server_pub", recv_msg.pubkey) (set by action)
            // server_code: DeriveCode("server_pub", recv_msg.pubkey) (set by action)
            state = .deriveSecret
            return state
        case (.waitingForClient, .pairHello) where guards[.tokenInvalid]?() == true:
            state = .idle
            return state
        case (.waitingForCode, .codeSubmit):
            // received_code: recv_msg.code (set by action)
            state = .validateCode
            return state
        case (.paired, .authRequest):
            // received_device_id: recv_msg.device_id (set by action)
            // received_auth_nonce: recv_msg.nonce (set by action)
            state = .authCheck
            return state
        default:
            return nil
        }
    }

    /// Attempt an internal transition. Returns the new state, or nil if none available.
    @discardableResult
    public func step() throws -> ServerState? {
        switch state {
        case .generateToken:
            try actions[.registerRelay]?()
            state = .registerRelay
            return state
        case .registerRelay:
            state = .waitingForClient
            return state
        case .deriveSecret:
            state = .sendAck
            return state
        case .sendAck:
            state = .waitingForCode
            return state
        case .validateCode:
            if guards[.codeCorrect]?() == true {
                state = .storePaired
                return state
            }
            if guards[.codeWrong]?() == true {
                // code_attempts: code_attempts + 1 (set by action)
                state = .idle
                return state
            }
            return nil
        case .storePaired:
            try actions[.storeDevice]?()
            deviceSecret = "dev_secret_1"
            // paired_devices: paired_devices \union {"device_1"} (set by action)
            // active_tokens: active_tokens \ {current_token} (set by action)
            // used_tokens: used_tokens \union {current_token} (set by action)
            state = .paired
            return state
        case .authCheck:
            if guards[.deviceKnown]?() == true {
                try actions[.verifyDevice]?()
                // auth_nonces_used: auth_nonces_used \union {received_auth_nonce} (set by action)
                state = .sessionActive
                return state
            }
            if guards[.deviceUnknown]?() == true {
                state = .idle
                return state
            }
            return nil
        case .sessionActive:
            state = .paired
            return state
        default:
            return nil
        }
    }
}

/// IosMachine is the generated state machine for the ios actor.
public final class IosMachine: @unchecked Sendable {
    public private(set) var state: IosState
    public var receivedServerPub: String // pubkey ios received in pair_hello_ack (may be adversary's)
    public var clientSharedKey: String // ECDH key derived by ios (tuple to match DeriveKey output type)
    public var iosCode: String // code computed by ios from its view of the pubkeys (tuple to match DeriveCode output type)

    public var guards: [GuardID: () -> Bool] = [:]
    public var actions: [ActionID: () throws -> Void] = [:]

    public init() {
        self.state = .idle
        self.receivedServerPub = "none"
        self.clientSharedKey = ""
        self.iosCode = ""
    }

    /// Handle any event (message receipt or internal). Returns emitted commands.
    @discardableResult
    public func handleEvent(_ ev: EventID) throws -> [CmdID] {
        switch (state, ev) {
        case (.idle, .userScansQR):
            state = .scanQR
            return []
        case (.scanQR, .qRParsed):
            state = .connectRelay
            return []
        case (.connectRelay, .relayConnected):
            state = .genKeyPair
            return []
        case (.genKeyPair, .keyPairGenerated):
            try actions[.sendPairHello]?()
            state = .waitAck
            return []
        case (.waitAck, .recvPairHelloAck):
            try actions[.deriveSecret]?()
            // received_server_pub: recv_msg.pubkey (set by action)
            // client_shared_key: DeriveKey("client_pub", recv_msg.pubkey) (set by action)
            state = .e2EReady
            return []
        case (.e2EReady, .recvPairConfirm):
            // ios_code: DeriveCode(received_server_pub, "client_pub") (set by action)
            state = .showCode
            return []
        case (.showCode, .codeDisplayed):
            state = .waitPairComplete
            return []
        case (.waitPairComplete, .recvPairComplete):
            try actions[.storeSecret]?()
            state = .paired
            return []
        case (.paired, .appLaunch):
            state = .reconnect
            return []
        case (.reconnect, .relayConnected):
            state = .sendAuth
            return []
        case (.sendAuth, .recvAuthOk):
            state = .sessionActive
            return []
        case (.sessionActive, .disconnect):
            state = .paired
            return []
        default:
            return []
        }
    }

    /// Process a received message. Returns the new state, or nil if rejected.
    @discardableResult
    public func handleMessage(_ msg: MessageType) throws -> IosState? {
        switch (state, msg) {
        case (.waitAck, .pairHelloAck):
            try actions[.deriveSecret]?()
            // received_server_pub: recv_msg.pubkey (set by action)
            // client_shared_key: DeriveKey("client_pub", recv_msg.pubkey) (set by action)
            state = .e2EReady
            return state
        case (.e2EReady, .pairConfirm):
            // ios_code: DeriveCode(received_server_pub, "client_pub") (set by action)
            state = .showCode
            return state
        case (.waitPairComplete, .pairComplete):
            try actions[.storeSecret]?()
            state = .paired
            return state
        case (.sendAuth, .authOk):
            state = .sessionActive
            return state
        default:
            return nil
        }
    }

    /// Attempt an internal transition. Returns the new state, or nil if none available.
    @discardableResult
    public func step() throws -> IosState? {
        switch state {
        case .idle:
            state = .scanQR
            return state
        case .scanQR:
            state = .connectRelay
            return state
        case .connectRelay:
            state = .genKeyPair
            return state
        case .genKeyPair:
            try actions[.sendPairHello]?()
            state = .waitAck
            return state
        case .showCode:
            state = .waitPairComplete
            return state
        case .paired:
            state = .reconnect
            return state
        case .reconnect:
            state = .sendAuth
            return state
        case .sessionActive:
            state = .paired
            return state
        default:
            return nil
        }
    }
}

/// CliMachine is the generated state machine for the cli actor.
public final class CliMachine: @unchecked Sendable {
    public private(set) var state: CliState

    public var guards: [GuardID: () -> Bool] = [:]
    public var actions: [ActionID: () throws -> Void] = [:]

    public init() {
        self.state = .idle
    }

    /// Handle any event (message receipt or internal). Returns emitted commands.
    @discardableResult
    public func handleEvent(_ ev: EventID) throws -> [CmdID] {
        switch (state, ev) {
        case (.idle, .cliInit):
            state = .getKey
            return []
        case (.getKey, .keyStored):
            state = .beginPair
            return []
        case (.beginPair, .recvTokenResponse):
            state = .showQR
            return []
        case (.showQR, .recvWaitingForCode):
            state = .promptCode
            return []
        case (.promptCode, .userEntersCode):
            state = .submitCode
            return []
        case (.submitCode, .recvPairStatus):
            state = .done
            return []
        default:
            return []
        }
    }

    /// Process a received message. Returns the new state, or nil if rejected.
    @discardableResult
    public func handleMessage(_ msg: MessageType) throws -> CliState? {
        switch (state, msg) {
        case (.beginPair, .tokenResponse):
            state = .showQR
            return state
        case (.showQR, .waitingForCode):
            state = .promptCode
            return state
        case (.submitCode, .pairStatus):
            state = .done
            return state
        default:
            return nil
        }
    }

    /// Attempt an internal transition. Returns the new state, or nil if none available.
    @discardableResult
    public func step() throws -> CliState? {
        switch state {
        case .idle:
            state = .getKey
            return state
        case .getKey:
            state = .beginPair
            return state
        case .promptCode:
            state = .submitCode
            return state
        default:
            return nil
        }
    }
}


// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Auto-generated from protocol definition. Do not edit.
// Source of truth: protocol/*.yaml

import Foundation

public enum SessionBackendState: String, Sendable {
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
    case relayConnected = "RelayConnected"
    case candidatesAdvertised = "CandidatesAdvertised"
    case altActive = "AltActive"
    case relayBackoff = "RelayBackoff"
    case altDegraded = "AltDegraded"
}

public enum SessionClientState: String, Sendable {
    case idle = "Idle"
    case obtainBackchannelSecret = "ObtainBackchannelSecret"
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
    case relayConnected = "RelayConnected"
    case pairDialing = "PairDialing"
    case pairChecking = "PairChecking"
    case altActive = "AltActive"
    case relayFallback = "RelayFallback"
}

public enum SessionRelayState: String, Sendable {
    case idle = "Idle"
    case backendRegistered = "BackendRegistered"
    case bridged = "Bridged"
}

/// Protocol wire constants shared across all platforms.
public enum SessionWire {
    public static let dgConnWhole: UInt8 = 0x00
    public static let dgPing: UInt8 = 0x10
    public static let dgPong: UInt8 = 0x11
    public static let dgConnFragment: UInt8 = 0x40
    public static let dgChanWhole: UInt8 = 0x80
    public static let dgChanFragment: UInt8 = 0xC0
    public static let fragHeaderSize = 8
    public static let chanIdSize = 2
    public static let maxDatagramPayload = 1200
    public static let fragmentTimeoutMs = 5000 // ms
    public static let frameApp: UInt8 = 0x00
    public static let frameCandidates: UInt8 = 0x01
    public static let frameCutover: UInt8 = 0x02
    public static let framePairCheck: UInt8 = 0x03
    public static let framePairCheckAck: UInt8 = 0x04
    public static let candHost = "host"
    public static let candSrflx = "srflx"
    public static let maxMessageSize = 1048576
    public static let lengthPrefixSize = 4
    public static let pingIntervalMs = 5000 // ms
    public static let pongTimeoutMs = 4000 // ms
    public static let maxPingFailures = 3
    public static let maxBackoffLevel = 5
    public static let streamChannelOpenerSuffix = ":o2a"
    public static let streamChannelAcceptSuffix = ":a2o"
    public static let dgChannelSendSuffix = ":dg:send"
    public static let dgChannelRecvSuffix = ":dg:recv"
    public static let channelIdHashMultiplier = 31
}

/// The protocol transition table and shared type enums.
public enum SessionProtocol {

    public enum MessageType: String, Sendable {
        case pairHello = "pair_hello"
        case pairHelloAck = "pair_hello_ack"
        case pairConfirm = "pair_confirm"
        case pairComplete = "pair_complete"
        case authRequest = "auth_request"
        case authOk = "auth_ok"
        case candidates = "candidates"
        case pairCheck = "pair_check"
        case pairCheckAck = "pair_check_ack"
        case pathPing = "path_ping"
        case pathPong = "path_pong"
    }

    public enum GuardID: String, Sendable {
        case tokenValid = "token_valid"
        case tokenInvalid = "token_invalid"
        case codeCorrect = "code_correct"
        case codeWrong = "code_wrong"
        case deviceKnown = "device_known"
        case deviceUnknown = "device_unknown"
        case nonceFresh = "nonce_fresh"
        case challengeValid = "challenge_valid"
        case challengeInvalid = "challenge_invalid"
        case altEnabled = "alt_enabled"
        case altDisabled = "alt_disabled"
        case localCandidatesAvailable = "local_candidates_available"
        case underMaxFailures = "under_max_failures"
        case atMaxFailures = "at_max_failures"
    }

    public enum ActionID: String, Sendable {
        case generateToken = "generate_token"
        case registerRelay = "register_relay"
        case deriveSecret = "derive_secret"
        case storeDevice = "store_device"
        case verifyDevice = "verify_device"
        case activateLan = "activate_lan"
        case fallbackToRelay = "fallback_to_relay"
        case resetFailures = "reset_failures"
        case sendPairHello = "send_pair_hello"
        case storeSecret = "store_secret"
        case dialCandidate = "dial_candidate"
        case bridgeStreams = "bridge_streams"
        case unbridge = "unbridge"
    }

    public enum EventID: String, Sendable {
        case appSend = "app_send"
        case appRecv = "app_recv"
        case appSendDatagram = "app_send_datagram"
        case appRecvDatagram = "app_recv_datagram"
        case appClose = "app_close"
        case appForceFallback = "app_force_fallback"
        case relayStreamData = "relay_stream_data"
        case relayStreamError = "relay_stream_error"
        case relayDatagram = "relay_datagram"
        case altStreamData = "alt_stream_data"
        case altStreamError = "alt_stream_error"
        case altDatagram = "alt_datagram"
        case dialOk = "dial_ok"
        case dialFailed = "dial_failed"
        case pairCheckOk = "pair_check_ok"
        case pingTimeout = "ping_timeout"
        case pingTick = "ping_tick"
        case backoffExpired = "backoff_expired"
        case candidatesTimeout = "candidates_timeout"
        case cliInitPair = "cli_init_pair"
        case tokenCreated = "token_created"
        case relayRegistered = "relay_registered"
        case ecdhComplete = "ecdh_complete"
        case signalCodeDisplay = "signal_code_display"
        case cliCodeEntered = "cli_code_entered"
        case checkCode = "check_code"
        case finalise = "finalise"
        case verify = "verify"
        case sessionEstablished = "session_established"
        case candidatesGathered = "candidates_gathered"
        case candidatesChanged = "candidates_changed"
        case candidatesRefreshTick = "candidates_refresh_tick"
        case disconnect = "disconnect"
        case backchannelReceived = "backchannel_received"
        case secretParsed = "secret_parsed"
        case relayConnected = "relay_connected"
        case keyPairGenerated = "key_pair_generated"
        case codeDisplayed = "code_displayed"
        case appLaunch = "app_launch"
        case verifyTimeout = "verify_timeout"
        case altError = "alt_error"
        case relayOk = "relay_ok"
        case backendRegister = "backend_register"
        case clientConnect = "client_connect"
        case clientDisconnect = "client_disconnect"
        case backendDisconnect = "backend_disconnect"
        case recvPairHello = "recv_pair_hello"
        case recvAuthRequest = "recv_auth_request"
        case recvPairCheck = "recv_pair_check"
        case recvPathPong = "recv_path_pong"
        case recvPairHelloAck = "recv_pair_hello_ack"
        case recvPairConfirm = "recv_pair_confirm"
        case recvPairComplete = "recv_pair_complete"
        case recvAuthOk = "recv_auth_ok"
        case recvCandidates = "recv_candidates"
        case recvPairCheckAck = "recv_pair_check_ack"
        case recvPathPing = "recv_path_ping"
    }

    public enum CmdID: String, Sendable {
        case writeActiveStream = "write_active_stream"
        case sendActiveDatagram = "send_active_datagram"
        case sendPathPing = "send_path_ping"
        case sendPathPong = "send_path_pong"
        case sendCandidates = "send_candidates"
        case sendPairCheck = "send_pair_check"
        case sendPairCheckAck = "send_pair_check_ack"
        case dialCandidate = "dial_candidate"
        case deliverRecv = "deliver_recv"
        case deliverRecvError = "deliver_recv_error"
        case deliverRecvDatagram = "deliver_recv_datagram"
        case startAltStreamReader = "start_alt_stream_reader"
        case stopAltStreamReader = "stop_alt_stream_reader"
        case startAltDgReader = "start_alt_dg_reader"
        case stopAltDgReader = "stop_alt_dg_reader"
        case startMonitor = "start_monitor"
        case stopMonitor = "stop_monitor"
        case startPongTimeout = "start_pong_timeout"
        case cancelPongTimeout = "cancel_pong_timeout"
        case startBackoffTimer = "start_backoff_timer"
        case closeAltPath = "close_alt_path"
        case signalAltReady = "signal_alt_ready"
        case resetAltReady = "reset_alt_ready"
        case setCryptoDatagram = "set_crypto_datagram"
    }


    /// backend transitions.
    public static let backendInitial: SessionBackendState = .idle

    public static let backendTransitions: [(from: String, to: String, on: String, onKind: String, guard: String?, action: String?, sends: [(to: String, msg: String)])] = [
        (from: "Idle", to: "GenerateToken", on: "cli_init_pair", onKind: "internal", guard: nil, action: "generate_token", sends: []),
        (from: "GenerateToken", to: "RegisterRelay", on: "token_created", onKind: "internal", guard: nil, action: "register_relay", sends: []),
        (from: "RegisterRelay", to: "WaitingForClient", on: "relay_registered", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "WaitingForClient", to: "DeriveSecret", on: "pair_hello", onKind: "recv", guard: "token_valid", action: "derive_secret", sends: []),
        (from: "WaitingForClient", to: "Idle", on: "pair_hello", onKind: "recv", guard: "token_invalid", action: nil, sends: []),
        (from: "DeriveSecret", to: "SendAck", on: "ecdh_complete", onKind: "internal", guard: nil, action: nil, sends: [(to: "client", msg: "pair_hello_ack")]),
        (from: "SendAck", to: "WaitingForCode", on: "signal_code_display", onKind: "internal", guard: nil, action: nil, sends: [(to: "client", msg: "pair_confirm")]),
        (from: "WaitingForCode", to: "ValidateCode", on: "cli_code_entered", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "ValidateCode", to: "StorePaired", on: "check_code", onKind: "internal", guard: "code_correct", action: nil, sends: []),
        (from: "ValidateCode", to: "Idle", on: "check_code", onKind: "internal", guard: "code_wrong", action: nil, sends: []),
        (from: "StorePaired", to: "Paired", on: "finalise", onKind: "internal", guard: nil, action: "store_device", sends: [(to: "client", msg: "pair_complete")]),
        (from: "Paired", to: "AuthCheck", on: "auth_request", onKind: "recv", guard: nil, action: nil, sends: []),
        (from: "AuthCheck", to: "SessionActive", on: "verify", onKind: "internal", guard: "device_known", action: "verify_device", sends: [(to: "client", msg: "auth_ok")]),
        (from: "AuthCheck", to: "Idle", on: "verify", onKind: "internal", guard: "device_unknown", action: nil, sends: []),
        (from: "SessionActive", to: "RelayConnected", on: "session_established", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "RelayConnected", to: "CandidatesAdvertised", on: "candidates_gathered", onKind: "internal", guard: nil, action: nil, sends: [(to: "client", msg: "candidates")]),
        (from: "CandidatesAdvertised", to: "AltActive", on: "pair_check", onKind: "recv", guard: "challenge_valid", action: "activate_lan", sends: [(to: "client", msg: "pair_check_ack")]),
        (from: "CandidatesAdvertised", to: "RelayConnected", on: "pair_check", onKind: "recv", guard: "challenge_invalid", action: nil, sends: []),
        (from: "CandidatesAdvertised", to: "RelayBackoff", on: "candidates_timeout", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "AltActive", to: "AltActive", on: "ping_tick", onKind: "internal", guard: nil, action: nil, sends: [(to: "client", msg: "path_ping")]),
        (from: "AltActive", to: "AltDegraded", on: "ping_timeout", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "AltDegraded", to: "AltDegraded", on: "ping_tick", onKind: "internal", guard: nil, action: nil, sends: [(to: "client", msg: "path_ping")]),
        (from: "AltActive", to: "RelayBackoff", on: "alt_stream_error", onKind: "internal", guard: nil, action: "fallback_to_relay", sends: []),
        (from: "AltDegraded", to: "RelayBackoff", on: "alt_stream_error", onKind: "internal", guard: nil, action: "fallback_to_relay", sends: []),
        (from: "AltDegraded", to: "AltActive", on: "path_pong", onKind: "recv", guard: nil, action: "reset_failures", sends: []),
        (from: "AltDegraded", to: "AltDegraded", on: "ping_timeout", onKind: "internal", guard: "under_max_failures", action: nil, sends: []),
        (from: "AltDegraded", to: "RelayBackoff", on: "ping_timeout", onKind: "internal", guard: "at_max_failures", action: "fallback_to_relay", sends: []),
        (from: "RelayBackoff", to: "CandidatesAdvertised", on: "backoff_expired", onKind: "internal", guard: nil, action: nil, sends: [(to: "client", msg: "candidates")]),
        (from: "RelayBackoff", to: "CandidatesAdvertised", on: "candidates_changed", onKind: "internal", guard: nil, action: nil, sends: [(to: "client", msg: "candidates")]),
        (from: "RelayConnected", to: "CandidatesAdvertised", on: "candidates_refresh_tick", onKind: "internal", guard: "local_candidates_available", action: nil, sends: [(to: "client", msg: "candidates")]),
        (from: "CandidatesAdvertised", to: "RelayConnected", on: "app_force_fallback", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "AltActive", to: "RelayBackoff", on: "app_force_fallback", onKind: "internal", guard: nil, action: "fallback_to_relay", sends: []),
        (from: "AltDegraded", to: "RelayBackoff", on: "app_force_fallback", onKind: "internal", guard: nil, action: "fallback_to_relay", sends: []),
        (from: "RelayConnected", to: "Paired", on: "disconnect", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "RelayConnected", to: "RelayConnected", on: "app_send", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "CandidatesAdvertised", to: "CandidatesAdvertised", on: "app_send", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "AltActive", to: "AltActive", on: "app_send", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "AltDegraded", to: "AltDegraded", on: "app_send", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "RelayBackoff", to: "RelayBackoff", on: "app_send", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "RelayConnected", to: "RelayConnected", on: "relay_stream_data", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "CandidatesAdvertised", to: "CandidatesAdvertised", on: "relay_stream_data", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "AltActive", to: "AltActive", on: "relay_stream_data", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "AltDegraded", to: "AltDegraded", on: "relay_stream_data", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "RelayBackoff", to: "RelayBackoff", on: "relay_stream_data", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "RelayConnected", to: "RelayConnected", on: "relay_stream_error", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "CandidatesAdvertised", to: "CandidatesAdvertised", on: "relay_stream_error", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "AltActive", to: "AltActive", on: "relay_stream_error", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "AltDegraded", to: "AltDegraded", on: "relay_stream_error", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "RelayBackoff", to: "RelayBackoff", on: "relay_stream_error", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "RelayConnected", to: "RelayConnected", on: "app_send_datagram", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "CandidatesAdvertised", to: "CandidatesAdvertised", on: "app_send_datagram", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "AltActive", to: "AltActive", on: "app_send_datagram", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "AltDegraded", to: "AltDegraded", on: "app_send_datagram", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "RelayBackoff", to: "RelayBackoff", on: "app_send_datagram", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "RelayConnected", to: "RelayConnected", on: "relay_datagram", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "CandidatesAdvertised", to: "CandidatesAdvertised", on: "relay_datagram", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "AltActive", to: "AltActive", on: "relay_datagram", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "AltDegraded", to: "AltDegraded", on: "relay_datagram", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "RelayBackoff", to: "RelayBackoff", on: "relay_datagram", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "AltActive", to: "AltActive", on: "alt_stream_data", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "AltDegraded", to: "AltDegraded", on: "alt_stream_data", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "AltActive", to: "AltActive", on: "alt_datagram", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "AltDegraded", to: "AltDegraded", on: "alt_datagram", onKind: "internal", guard: nil, action: nil, sends: []),
    ]

    /// client transitions.
    public static let clientInitial: SessionClientState = .idle

    public static let clientTransitions: [(from: String, to: String, on: String, onKind: String, guard: String?, action: String?, sends: [(to: String, msg: String)])] = [
        (from: "Idle", to: "ObtainBackchannelSecret", on: "backchannel_received", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "ObtainBackchannelSecret", to: "ConnectRelay", on: "secret_parsed", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "ConnectRelay", to: "GenKeyPair", on: "relay_connected", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "GenKeyPair", to: "WaitAck", on: "key_pair_generated", onKind: "internal", guard: nil, action: "send_pair_hello", sends: [(to: "backend", msg: "pair_hello")]),
        (from: "WaitAck", to: "E2EReady", on: "pair_hello_ack", onKind: "recv", guard: nil, action: "derive_secret", sends: []),
        (from: "E2EReady", to: "ShowCode", on: "pair_confirm", onKind: "recv", guard: nil, action: nil, sends: []),
        (from: "ShowCode", to: "WaitPairComplete", on: "code_displayed", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "WaitPairComplete", to: "Paired", on: "pair_complete", onKind: "recv", guard: nil, action: "store_secret", sends: []),
        (from: "Paired", to: "Reconnect", on: "app_launch", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "Reconnect", to: "SendAuth", on: "relay_connected", onKind: "internal", guard: nil, action: nil, sends: [(to: "backend", msg: "auth_request")]),
        (from: "SendAuth", to: "SessionActive", on: "auth_ok", onKind: "recv", guard: nil, action: nil, sends: []),
        (from: "SessionActive", to: "RelayConnected", on: "session_established", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "RelayConnected", to: "PairDialing", on: "candidates", onKind: "recv", guard: "alt_enabled", action: "dial_candidate", sends: []),
        (from: "RelayConnected", to: "RelayConnected", on: "candidates", onKind: "recv", guard: "alt_disabled", action: nil, sends: []),
        (from: "PairDialing", to: "PairChecking", on: "dial_ok", onKind: "internal", guard: nil, action: nil, sends: [(to: "backend", msg: "pair_check")]),
        (from: "PairDialing", to: "RelayConnected", on: "dial_failed", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "PairChecking", to: "AltActive", on: "pair_check_ack", onKind: "recv", guard: nil, action: "activate_lan", sends: []),
        (from: "PairChecking", to: "RelayConnected", on: "verify_timeout", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "AltActive", to: "AltActive", on: "path_ping", onKind: "recv", guard: nil, action: nil, sends: [(to: "backend", msg: "path_pong")]),
        (from: "AltActive", to: "RelayFallback", on: "alt_error", onKind: "internal", guard: nil, action: "fallback_to_relay", sends: []),
        (from: "AltActive", to: "RelayFallback", on: "alt_stream_error", onKind: "internal", guard: nil, action: "fallback_to_relay", sends: []),
        (from: "RelayFallback", to: "RelayConnected", on: "relay_ok", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "AltActive", to: "PairDialing", on: "candidates", onKind: "recv", guard: "alt_enabled", action: "dial_candidate", sends: []),
        (from: "PairDialing", to: "RelayConnected", on: "app_force_fallback", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "PairChecking", to: "RelayConnected", on: "app_force_fallback", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "AltActive", to: "RelayConnected", on: "app_force_fallback", onKind: "internal", guard: nil, action: "fallback_to_relay", sends: []),
        (from: "RelayConnected", to: "Paired", on: "disconnect", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "RelayConnected", to: "RelayConnected", on: "app_send", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "PairDialing", to: "PairDialing", on: "app_send", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "PairChecking", to: "PairChecking", on: "app_send", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "AltActive", to: "AltActive", on: "app_send", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "RelayFallback", to: "RelayFallback", on: "app_send", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "RelayConnected", to: "RelayConnected", on: "relay_stream_data", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "PairDialing", to: "PairDialing", on: "relay_stream_data", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "PairChecking", to: "PairChecking", on: "relay_stream_data", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "AltActive", to: "AltActive", on: "relay_stream_data", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "RelayFallback", to: "RelayFallback", on: "relay_stream_data", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "RelayConnected", to: "RelayConnected", on: "relay_stream_error", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "PairDialing", to: "PairDialing", on: "relay_stream_error", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "PairChecking", to: "PairChecking", on: "relay_stream_error", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "AltActive", to: "AltActive", on: "relay_stream_error", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "RelayFallback", to: "RelayFallback", on: "relay_stream_error", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "RelayConnected", to: "RelayConnected", on: "app_send_datagram", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "PairDialing", to: "PairDialing", on: "app_send_datagram", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "PairChecking", to: "PairChecking", on: "app_send_datagram", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "AltActive", to: "AltActive", on: "app_send_datagram", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "RelayFallback", to: "RelayFallback", on: "app_send_datagram", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "RelayConnected", to: "RelayConnected", on: "relay_datagram", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "PairDialing", to: "PairDialing", on: "relay_datagram", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "PairChecking", to: "PairChecking", on: "relay_datagram", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "AltActive", to: "AltActive", on: "relay_datagram", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "RelayFallback", to: "RelayFallback", on: "relay_datagram", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "AltActive", to: "AltActive", on: "alt_stream_data", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "AltActive", to: "AltActive", on: "alt_datagram", onKind: "internal", guard: nil, action: nil, sends: []),
    ]

    /// relay transitions.
    public static let relayInitial: SessionRelayState = .idle

    public static let relayTransitions: [(from: String, to: String, on: String, onKind: String, guard: String?, action: String?, sends: [(to: String, msg: String)])] = [
        (from: "Idle", to: "BackendRegistered", on: "backend_register", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "BackendRegistered", to: "Bridged", on: "client_connect", onKind: "internal", guard: nil, action: "bridge_streams", sends: []),
        (from: "Bridged", to: "BackendRegistered", on: "client_disconnect", onKind: "internal", guard: nil, action: "unbridge", sends: []),
        (from: "BackendRegistered", to: "Idle", on: "backend_disconnect", onKind: "internal", guard: nil, action: nil, sends: []),
    ]
}

/// SessionBackendMachine is the generated state machine for the backend actor.
public final class SessionBackendMachine: @unchecked Sendable {
    public typealias MessageType = SessionProtocol.MessageType
    public typealias GuardID = SessionProtocol.GuardID
    public typealias ActionID = SessionProtocol.ActionID
    public typealias EventID = SessionProtocol.EventID
    public typealias CmdID = SessionProtocol.CmdID

    public private(set) var state: SessionBackendState
    public var currentToken: String // pairing token currently in play
    public var activeTokens: String // set of valid (non-revoked) tokens
    public var usedTokens: String // set of revoked tokens
    public var backendEcdhPub: String // backend ECDH public key
    public var receivedClientPub: String // pubkey backend received in pair_hello
    public var backendSharedKey: String // ECDH key derived by backend
    public var backendCode: String // code computed by backend
    public var receivedCode: String // code entered via CLI
    public var codeAttempts: Int // failed code submission attempts
    public var deviceSecret: String // persistent device secret
    public var pairedDevices: String // device IDs that completed pairing
    public var receivedDeviceId: String // device_id from auth_request
    public var authNoncesUsed: String // set of consumed auth nonces
    public var receivedAuthNonce: String // nonce from auth_request
    public var secretPublished: Bool // whether token has been published via backchannel
    public var pingFailures: Int // consecutive failed pings
    public var backoffLevel: Int // exponential backoff level
    public var bActivePath: String // backend active path
    public var bDispatcherPath: String // backend datagram dispatcher binding
    public var monitorTarget: String // health monitor target
    public var altSignal: String // AltReady notification state

    public var guards: [GuardID: () -> Bool] = [:]
    public var actions: [ActionID: () throws -> Void] = [:]

    public init() {
        self.state = .idle
        self.currentToken = "none"
        self.activeTokens = ""
        self.usedTokens = ""
        self.backendEcdhPub = "none"
        self.receivedClientPub = "none"
        self.backendSharedKey = ""
        self.backendCode = ""
        self.receivedCode = ""
        self.codeAttempts = 0
        self.deviceSecret = "none"
        self.pairedDevices = ""
        self.receivedDeviceId = "none"
        self.authNoncesUsed = ""
        self.receivedAuthNonce = "none"
        self.secretPublished = false
        self.pingFailures = 0
        self.backoffLevel = 0
        self.bActivePath = "relay"
        self.bDispatcherPath = "relay"
        self.monitorTarget = "none"
        self.altSignal = "pending"
    }

    /// Handle any event (message receipt or internal). Returns emitted commands.
    @discardableResult
    public func handleEvent(_ ev: EventID) throws -> [CmdID] {
        switch (state, ev) {
        case (.idle, .cliInitPair):
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
            secretPublished = true
            state = .waitingForClient
            return []
        case (.waitingForClient, .recvPairHello) where guards[.tokenValid]?() == true:
            try actions[.deriveSecret]?()
            // received_client_pub: recv_msg.pubkey (set by action)
            backendEcdhPub = "backend_pub"
            // backend_shared_key: DeriveKey("backend_pub", recv_msg.pubkey) (set by action)
            // backend_code: DeriveCode("backend_pub", recv_msg.pubkey) (set by action)
            state = .deriveSecret
            return []
        case (.waitingForClient, .recvPairHello) where guards[.tokenInvalid]?() == true:
            state = .idle
            return []
        case (.deriveSecret, .ecdhComplete):
            state = .sendAck
            return []
        case (.sendAck, .signalCodeDisplay):
            state = .waitingForCode
            return []
        case (.waitingForCode, .cliCodeEntered):
            // received_code: cli_entered_code (set by action)
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
        case (.sessionActive, .sessionEstablished):
            state = .relayConnected
            return []
        case (.relayConnected, .candidatesGathered):
            state = .candidatesAdvertised
            return [.sendCandidates]
        case (.candidatesAdvertised, .recvPairCheck) where guards[.challengeValid]?() == true:
            try actions[.activateLan]?()
            pingFailures = 0
            backoffLevel = 0
            bActivePath = "alt"
            bDispatcherPath = "alt"
            monitorTarget = "alt"
            altSignal = "ready"
            state = .altActive
            return [.sendPairCheckAck, .startAltStreamReader, .startAltDgReader, .startMonitor, .signalAltReady, .setCryptoDatagram]
        case (.candidatesAdvertised, .recvPairCheck) where guards[.challengeInvalid]?() == true:
            state = .relayConnected
            return []
        case (.candidatesAdvertised, .candidatesTimeout):
            // backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
            altSignal = "pending"
            state = .relayBackoff
            return [.resetAltReady, .startBackoffTimer]
        case (.altActive, .pingTick):
            state = .altActive
            return [.sendPathPing, .startPongTimeout]
        case (.altActive, .pingTimeout):
            pingFailures = 1
            state = .altDegraded
            return []
        case (.altDegraded, .pingTick):
            state = .altDegraded
            return [.sendPathPing, .startPongTimeout]
        case (.altActive, .altStreamError):
            try actions[.fallbackToRelay]?()
            // backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
            bActivePath = "relay"
            bDispatcherPath = "relay"
            monitorTarget = "none"
            altSignal = "pending"
            pingFailures = 0
            state = .relayBackoff
            return [.stopMonitor, .stopAltStreamReader, .stopAltDgReader, .closeAltPath, .resetAltReady, .startBackoffTimer]
        case (.altDegraded, .altStreamError):
            try actions[.fallbackToRelay]?()
            // backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
            bActivePath = "relay"
            bDispatcherPath = "relay"
            monitorTarget = "none"
            altSignal = "pending"
            pingFailures = 0
            state = .relayBackoff
            return [.stopMonitor, .stopAltStreamReader, .stopAltDgReader, .closeAltPath, .resetAltReady, .startBackoffTimer]
        case (.altDegraded, .recvPathPong):
            try actions[.resetFailures]?()
            pingFailures = 0
            state = .altActive
            return [.cancelPongTimeout]
        case (.altDegraded, .pingTimeout) where guards[.underMaxFailures]?() == true:
            // ping_failures: ping_failures + 1 (set by action)
            state = .altDegraded
            return []
        case (.altDegraded, .pingTimeout) where guards[.atMaxFailures]?() == true:
            try actions[.fallbackToRelay]?()
            // backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
            bActivePath = "relay"
            bDispatcherPath = "relay"
            monitorTarget = "none"
            altSignal = "pending"
            pingFailures = 0
            state = .relayBackoff
            return [.stopMonitor, .stopAltStreamReader, .stopAltDgReader, .closeAltPath, .resetAltReady, .startBackoffTimer]
        case (.relayBackoff, .backoffExpired):
            state = .candidatesAdvertised
            return [.sendCandidates]
        case (.relayBackoff, .candidatesChanged):
            backoffLevel = 0
            state = .candidatesAdvertised
            return [.sendCandidates]
        case (.relayConnected, .candidatesRefreshTick) where guards[.localCandidatesAvailable]?() == true:
            state = .candidatesAdvertised
            return [.sendCandidates]
        case (.candidatesAdvertised, .appForceFallback):
            altSignal = "pending"
            state = .relayConnected
            return [.resetAltReady]
        case (.altActive, .appForceFallback):
            try actions[.fallbackToRelay]?()
            // backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
            bActivePath = "relay"
            bDispatcherPath = "relay"
            monitorTarget = "none"
            altSignal = "pending"
            pingFailures = 0
            state = .relayBackoff
            return [.stopMonitor, .cancelPongTimeout, .stopAltStreamReader, .stopAltDgReader, .closeAltPath, .resetAltReady, .startBackoffTimer]
        case (.altDegraded, .appForceFallback):
            try actions[.fallbackToRelay]?()
            // backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
            bActivePath = "relay"
            bDispatcherPath = "relay"
            monitorTarget = "none"
            altSignal = "pending"
            pingFailures = 0
            state = .relayBackoff
            return [.stopMonitor, .cancelPongTimeout, .stopAltStreamReader, .stopAltDgReader, .closeAltPath, .resetAltReady, .startBackoffTimer]
        case (.relayConnected, .disconnect):
            state = .paired
            return []
        case (.relayConnected, .appSend):
            state = .relayConnected
            return [.writeActiveStream]
        case (.candidatesAdvertised, .appSend):
            state = .candidatesAdvertised
            return [.writeActiveStream]
        case (.altActive, .appSend):
            state = .altActive
            return [.writeActiveStream]
        case (.altDegraded, .appSend):
            state = .altDegraded
            return [.writeActiveStream]
        case (.relayBackoff, .appSend):
            state = .relayBackoff
            return [.writeActiveStream]
        case (.relayConnected, .relayStreamData):
            state = .relayConnected
            return [.deliverRecv]
        case (.candidatesAdvertised, .relayStreamData):
            state = .candidatesAdvertised
            return [.deliverRecv]
        case (.altActive, .relayStreamData):
            state = .altActive
            return [.deliverRecv]
        case (.altDegraded, .relayStreamData):
            state = .altDegraded
            return [.deliverRecv]
        case (.relayBackoff, .relayStreamData):
            state = .relayBackoff
            return [.deliverRecv]
        case (.relayConnected, .relayStreamError):
            state = .relayConnected
            return [.deliverRecvError]
        case (.candidatesAdvertised, .relayStreamError):
            state = .candidatesAdvertised
            return [.deliverRecvError]
        case (.altActive, .relayStreamError):
            state = .altActive
            return [.deliverRecvError]
        case (.altDegraded, .relayStreamError):
            state = .altDegraded
            return [.deliverRecvError]
        case (.relayBackoff, .relayStreamError):
            state = .relayBackoff
            return [.deliverRecvError]
        case (.relayConnected, .appSendDatagram):
            state = .relayConnected
            return [.sendActiveDatagram]
        case (.candidatesAdvertised, .appSendDatagram):
            state = .candidatesAdvertised
            return [.sendActiveDatagram]
        case (.altActive, .appSendDatagram):
            state = .altActive
            return [.sendActiveDatagram]
        case (.altDegraded, .appSendDatagram):
            state = .altDegraded
            return [.sendActiveDatagram]
        case (.relayBackoff, .appSendDatagram):
            state = .relayBackoff
            return [.sendActiveDatagram]
        case (.relayConnected, .relayDatagram):
            state = .relayConnected
            return [.deliverRecvDatagram]
        case (.candidatesAdvertised, .relayDatagram):
            state = .candidatesAdvertised
            return [.deliverRecvDatagram]
        case (.altActive, .relayDatagram):
            state = .altActive
            return [.deliverRecvDatagram]
        case (.altDegraded, .relayDatagram):
            state = .altDegraded
            return [.deliverRecvDatagram]
        case (.relayBackoff, .relayDatagram):
            state = .relayBackoff
            return [.deliverRecvDatagram]
        case (.altActive, .altStreamData):
            state = .altActive
            return [.deliverRecv]
        case (.altDegraded, .altStreamData):
            state = .altDegraded
            return [.deliverRecv]
        case (.altActive, .altDatagram):
            state = .altActive
            return [.deliverRecvDatagram]
        case (.altDegraded, .altDatagram):
            state = .altDegraded
            return [.deliverRecvDatagram]
        default:
            return []
        }
    }

    /// Process a received message. Returns the new state, or nil if rejected.
    @discardableResult
    public func handleMessage(_ msg: MessageType) throws -> SessionBackendState? {
        switch (state, msg) {
        case (.waitingForClient, .pairHello) where guards[.tokenValid]?() == true:
            try actions[.deriveSecret]?()
            // received_client_pub: recv_msg.pubkey (set by action)
            backendEcdhPub = "backend_pub"
            // backend_shared_key: DeriveKey("backend_pub", recv_msg.pubkey) (set by action)
            // backend_code: DeriveCode("backend_pub", recv_msg.pubkey) (set by action)
            state = .deriveSecret
            return state
        case (.waitingForClient, .pairHello) where guards[.tokenInvalid]?() == true:
            state = .idle
            return state
        case (.paired, .authRequest):
            // received_device_id: recv_msg.device_id (set by action)
            // received_auth_nonce: recv_msg.nonce (set by action)
            state = .authCheck
            return state
        case (.candidatesAdvertised, .pairCheck) where guards[.challengeValid]?() == true:
            try actions[.activateLan]?()
            pingFailures = 0
            backoffLevel = 0
            bActivePath = "alt"
            bDispatcherPath = "alt"
            monitorTarget = "alt"
            altSignal = "ready"
            state = .altActive
            return state
        case (.candidatesAdvertised, .pairCheck) where guards[.challengeInvalid]?() == true:
            state = .relayConnected
            return state
        case (.altDegraded, .pathPong):
            try actions[.resetFailures]?()
            pingFailures = 0
            state = .altActive
            return state
        default:
            return nil
        }
    }

    /// Attempt an internal transition. Returns the new state, or nil if none available.
    @discardableResult
    public func step() throws -> SessionBackendState? {
        switch state {
        case .idle:
            try actions[.generateToken]?()
            currentToken = "tok_1"
            // active_tokens: active_tokens \union {"tok_1"} (set by action)
            state = .generateToken
            return state
        case .generateToken:
            try actions[.registerRelay]?()
            state = .registerRelay
            return state
        case .registerRelay:
            secretPublished = true
            state = .waitingForClient
            return state
        case .deriveSecret:
            state = .sendAck
            return state
        case .sendAck:
            state = .waitingForCode
            return state
        case .waitingForCode:
            // received_code: cli_entered_code (set by action)
            state = .validateCode
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
            state = .relayConnected
            return state
        default:
            return nil
        }
    }
}

/// SessionClientMachine is the generated state machine for the client actor.
public final class SessionClientMachine: @unchecked Sendable {
    public typealias MessageType = SessionProtocol.MessageType
    public typealias GuardID = SessionProtocol.GuardID
    public typealias ActionID = SessionProtocol.ActionID
    public typealias EventID = SessionProtocol.EventID
    public typealias CmdID = SessionProtocol.CmdID

    public private(set) var state: SessionClientState
    public var receivedBackendPub: String // pubkey client received in pair_hello_ack
    public var clientSharedKey: String // ECDH key derived by client
    public var clientCode: String // code computed by client
    public var cActivePath: String // client active path
    public var cDispatcherPath: String // client datagram dispatcher binding
    public var altSignal: String // AltReady notification state

    public var guards: [GuardID: () -> Bool] = [:]
    public var actions: [ActionID: () throws -> Void] = [:]

    public init() {
        self.state = .idle
        self.receivedBackendPub = "none"
        self.clientSharedKey = ""
        self.clientCode = ""
        self.cActivePath = "relay"
        self.cDispatcherPath = "relay"
        self.altSignal = "pending"
    }

    /// Handle any event (message receipt or internal). Returns emitted commands.
    @discardableResult
    public func handleEvent(_ ev: EventID) throws -> [CmdID] {
        switch (state, ev) {
        case (.idle, .backchannelReceived):
            state = .obtainBackchannelSecret
            return []
        case (.obtainBackchannelSecret, .secretParsed):
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
            // received_backend_pub: recv_msg.pubkey (set by action)
            // client_shared_key: DeriveKey("client_pub", recv_msg.pubkey) (set by action)
            state = .e2EReady
            return []
        case (.e2EReady, .recvPairConfirm):
            // client_code: DeriveCode(received_backend_pub, "client_pub") (set by action)
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
        case (.sessionActive, .sessionEstablished):
            state = .relayConnected
            return []
        case (.relayConnected, .recvCandidates) where guards[.altEnabled]?() == true:
            try actions[.dialCandidate]?()
            state = .pairDialing
            return [.dialCandidate]
        case (.relayConnected, .recvCandidates) where guards[.altDisabled]?() == true:
            state = .relayConnected
            return []
        case (.pairDialing, .dialOk):
            state = .pairChecking
            return [.sendPairCheck]
        case (.pairDialing, .dialFailed):
            state = .relayConnected
            return []
        case (.pairChecking, .recvPairCheckAck):
            try actions[.activateLan]?()
            cActivePath = "alt"
            cDispatcherPath = "alt"
            altSignal = "ready"
            state = .altActive
            return [.startAltStreamReader, .startAltDgReader, .signalAltReady, .setCryptoDatagram]
        case (.pairChecking, .verifyTimeout):
            cDispatcherPath = "relay"
            state = .relayConnected
            return []
        case (.altActive, .recvPathPing):
            state = .altActive
            return [.sendPathPong]
        case (.altActive, .altError):
            try actions[.fallbackToRelay]?()
            cActivePath = "relay"
            cDispatcherPath = "relay"
            altSignal = "pending"
            state = .relayFallback
            return [.stopAltStreamReader, .stopAltDgReader, .closeAltPath, .resetAltReady]
        case (.altActive, .altStreamError):
            try actions[.fallbackToRelay]?()
            cActivePath = "relay"
            cDispatcherPath = "relay"
            altSignal = "pending"
            state = .relayFallback
            return [.stopAltStreamReader, .stopAltDgReader, .closeAltPath, .resetAltReady]
        case (.relayFallback, .relayOk):
            state = .relayConnected
            return []
        case (.altActive, .recvCandidates) where guards[.altEnabled]?() == true:
            try actions[.dialCandidate]?()
            state = .pairDialing
            return [.stopAltStreamReader, .stopAltDgReader, .closeAltPath, .dialCandidate]
        case (.pairDialing, .appForceFallback):
            state = .relayConnected
            return []
        case (.pairChecking, .appForceFallback):
            cDispatcherPath = "relay"
            state = .relayConnected
            return [.stopAltStreamReader, .stopAltDgReader, .closeAltPath]
        case (.altActive, .appForceFallback):
            try actions[.fallbackToRelay]?()
            cActivePath = "relay"
            cDispatcherPath = "relay"
            altSignal = "pending"
            state = .relayConnected
            return [.stopAltStreamReader, .stopAltDgReader, .closeAltPath, .resetAltReady]
        case (.relayConnected, .disconnect):
            state = .paired
            return []
        case (.relayConnected, .appSend):
            state = .relayConnected
            return [.writeActiveStream]
        case (.pairDialing, .appSend):
            state = .pairDialing
            return [.writeActiveStream]
        case (.pairChecking, .appSend):
            state = .pairChecking
            return [.writeActiveStream]
        case (.altActive, .appSend):
            state = .altActive
            return [.writeActiveStream]
        case (.relayFallback, .appSend):
            state = .relayFallback
            return [.writeActiveStream]
        case (.relayConnected, .relayStreamData):
            state = .relayConnected
            return [.deliverRecv]
        case (.pairDialing, .relayStreamData):
            state = .pairDialing
            return [.deliverRecv]
        case (.pairChecking, .relayStreamData):
            state = .pairChecking
            return [.deliverRecv]
        case (.altActive, .relayStreamData):
            state = .altActive
            return [.deliverRecv]
        case (.relayFallback, .relayStreamData):
            state = .relayFallback
            return [.deliverRecv]
        case (.relayConnected, .relayStreamError):
            state = .relayConnected
            return [.deliverRecvError]
        case (.pairDialing, .relayStreamError):
            state = .pairDialing
            return [.deliverRecvError]
        case (.pairChecking, .relayStreamError):
            state = .pairChecking
            return [.deliverRecvError]
        case (.altActive, .relayStreamError):
            state = .altActive
            return [.deliverRecvError]
        case (.relayFallback, .relayStreamError):
            state = .relayFallback
            return [.deliverRecvError]
        case (.relayConnected, .appSendDatagram):
            state = .relayConnected
            return [.sendActiveDatagram]
        case (.pairDialing, .appSendDatagram):
            state = .pairDialing
            return [.sendActiveDatagram]
        case (.pairChecking, .appSendDatagram):
            state = .pairChecking
            return [.sendActiveDatagram]
        case (.altActive, .appSendDatagram):
            state = .altActive
            return [.sendActiveDatagram]
        case (.relayFallback, .appSendDatagram):
            state = .relayFallback
            return [.sendActiveDatagram]
        case (.relayConnected, .relayDatagram):
            state = .relayConnected
            return [.deliverRecvDatagram]
        case (.pairDialing, .relayDatagram):
            state = .pairDialing
            return [.deliverRecvDatagram]
        case (.pairChecking, .relayDatagram):
            state = .pairChecking
            return [.deliverRecvDatagram]
        case (.altActive, .relayDatagram):
            state = .altActive
            return [.deliverRecvDatagram]
        case (.relayFallback, .relayDatagram):
            state = .relayFallback
            return [.deliverRecvDatagram]
        case (.altActive, .altStreamData):
            state = .altActive
            return [.deliverRecv]
        case (.altActive, .altDatagram):
            state = .altActive
            return [.deliverRecvDatagram]
        default:
            return []
        }
    }

    /// Process a received message. Returns the new state, or nil if rejected.
    @discardableResult
    public func handleMessage(_ msg: MessageType) throws -> SessionClientState? {
        switch (state, msg) {
        case (.waitAck, .pairHelloAck):
            try actions[.deriveSecret]?()
            // received_backend_pub: recv_msg.pubkey (set by action)
            // client_shared_key: DeriveKey("client_pub", recv_msg.pubkey) (set by action)
            state = .e2EReady
            return state
        case (.e2EReady, .pairConfirm):
            // client_code: DeriveCode(received_backend_pub, "client_pub") (set by action)
            state = .showCode
            return state
        case (.waitPairComplete, .pairComplete):
            try actions[.storeSecret]?()
            state = .paired
            return state
        case (.sendAuth, .authOk):
            state = .sessionActive
            return state
        case (.relayConnected, .candidates) where guards[.altEnabled]?() == true:
            try actions[.dialCandidate]?()
            state = .pairDialing
            return state
        case (.relayConnected, .candidates) where guards[.altDisabled]?() == true:
            state = .relayConnected
            return state
        case (.pairChecking, .pairCheckAck):
            try actions[.activateLan]?()
            cActivePath = "alt"
            cDispatcherPath = "alt"
            altSignal = "ready"
            state = .altActive
            return state
        case (.altActive, .pathPing):
            state = .altActive
            return state
        case (.altActive, .candidates) where guards[.altEnabled]?() == true:
            try actions[.dialCandidate]?()
            state = .pairDialing
            return state
        default:
            return nil
        }
    }

    /// Attempt an internal transition. Returns the new state, or nil if none available.
    @discardableResult
    public func step() throws -> SessionClientState? {
        switch state {
        case .idle:
            state = .obtainBackchannelSecret
            return state
        case .obtainBackchannelSecret:
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
            state = .relayConnected
            return state
        default:
            return nil
        }
    }
}

/// SessionRelayMachine is the generated state machine for the relay actor.
public final class SessionRelayMachine: @unchecked Sendable {
    public typealias MessageType = SessionProtocol.MessageType
    public typealias GuardID = SessionProtocol.GuardID
    public typealias ActionID = SessionProtocol.ActionID
    public typealias EventID = SessionProtocol.EventID
    public typealias CmdID = SessionProtocol.CmdID

    public private(set) var state: SessionRelayState
    public var relayBridge: String // relay bridge state

    public var guards: [GuardID: () -> Bool] = [:]
    public var actions: [ActionID: () throws -> Void] = [:]

    public init() {
        self.state = .idle
        self.relayBridge = "idle"
    }

    /// Handle any event (message receipt or internal). Returns emitted commands.
    @discardableResult
    public func handleEvent(_ ev: EventID) throws -> [CmdID] {
        switch (state, ev) {
        case (.idle, .backendRegister):
            state = .backendRegistered
            return []
        case (.backendRegistered, .clientConnect):
            try actions[.bridgeStreams]?()
            relayBridge = "active"
            state = .bridged
            return []
        case (.bridged, .clientDisconnect):
            try actions[.unbridge]?()
            relayBridge = "idle"
            state = .backendRegistered
            return []
        case (.backendRegistered, .backendDisconnect):
            state = .idle
            return []
        default:
            return []
        }
    }

    /// Process a received message. Returns the new state, or nil if rejected.
    @discardableResult
    public func handleMessage(_ msg: MessageType) throws -> SessionRelayState? {
        switch (state, msg) {
        default:
            return nil
        }
    }

    /// Attempt an internal transition. Returns the new state, or nil if none available.
    @discardableResult
    public func step() throws -> SessionRelayState? {
        switch state {
        case .idle:
            state = .backendRegistered
            return state
        case .bridged:
            try actions[.unbridge]?()
            relayBridge = "idle"
            state = .backendRegistered
            return state
        default:
            return nil
        }
    }
}


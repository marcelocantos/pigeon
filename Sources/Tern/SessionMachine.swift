// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Auto-generated from protocol definition. Do not edit.
// Source of truth: protocol/*.yaml

import Foundation

public enum MessageType: String, Sendable {
    case pairhello = "pair_hello"
    case pairhelloack = "pair_hello_ack"
    case pairconfirm = "pair_confirm"
    case paircomplete = "pair_complete"
    case authrequest = "auth_request"
    case authok = "auth_ok"
    case lanoffer = "lan_offer"
    case lanverify = "lan_verify"
    case lanconfirm = "lan_confirm"
    case pathping = "path_ping"
    case pathpong = "path_pong"
}

public enum BackendState: String, Sendable {
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
    case lANOffered = "LANOffered"
    case lANActive = "LANActive"
    case lANDegraded = "LANDegraded"
    case relayBackoff = "RelayBackoff"
}

public enum ClientState: String, Sendable {
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
    case lANConnecting = "LANConnecting"
    case lANVerifying = "LANVerifying"
    case lANActive = "LANActive"
    case relayFallback = "RelayFallback"
}

public enum RelayState: String, Sendable {
    case idle = "Idle"
    case backendRegistered = "BackendRegistered"
    case bridged = "Bridged"
}

public enum GuardID: String, Sendable {
    case tokenvalid = "token_valid"
    case tokeninvalid = "token_invalid"
    case codecorrect = "code_correct"
    case codewrong = "code_wrong"
    case deviceknown = "device_known"
    case deviceunknown = "device_unknown"
    case noncefresh = "nonce_fresh"
    case challengevalid = "challenge_valid"
    case challengeinvalid = "challenge_invalid"
    case lanenabled = "lan_enabled"
    case landisabled = "lan_disabled"
    case lanserveravailable = "lan_server_available"
    case undermaxfailures = "under_max_failures"
    case atmaxfailures = "at_max_failures"
}

public enum ActionID: String, Sendable {
    case generatetoken = "generate_token"
    case registerrelay = "register_relay"
    case derivesecret = "derive_secret"
    case storedevice = "store_device"
    case verifydevice = "verify_device"
    case activatelan = "activate_lan"
    case resetfailures = "reset_failures"
    case fallbacktorelay = "fallback_to_relay"
    case sendpairhello = "send_pair_hello"
    case storesecret = "store_secret"
    case diallan = "dial_lan"
    case bridgestreams = "bridge_streams"
    case unbridge = "unbridge"
}

/// The protocol transition table. Fed to Machine for execution.
public enum SessionProtocol {

    /// backend transitions.
    public static let backendInitial: BackendState = .idle

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
        (from: "RelayConnected", to: "RelayConnected", on: "app_send", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "RelayConnected", to: "RelayConnected", on: "relay_stream_data", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "LANOffered", to: "LANOffered", on: "app_send", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "LANOffered", to: "LANOffered", on: "relay_stream_data", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "LANActive", to: "LANActive", on: "app_send", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "LANActive", to: "LANActive", on: "lan_stream_data", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "LANActive", to: "LANActive", on: "relay_stream_data", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "LANDegraded", to: "LANDegraded", on: "app_send", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "LANDegraded", to: "LANDegraded", on: "lan_stream_data", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "LANDegraded", to: "LANDegraded", on: "relay_stream_data", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "RelayBackoff", to: "RelayBackoff", on: "app_send", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "RelayBackoff", to: "RelayBackoff", on: "relay_stream_data", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "RelayConnected", to: "LANOffered", on: "lan_server_ready", onKind: "internal", guard: nil, action: nil, sends: [(to: "client", msg: "lan_offer")]),
        (from: "LANOffered", to: "LANActive", on: "lan_verify", onKind: "recv", guard: "challenge_valid", action: "activate_lan", sends: [(to: "client", msg: "lan_confirm")]),
        (from: "LANOffered", to: "RelayConnected", on: "lan_verify", onKind: "recv", guard: "challenge_invalid", action: nil, sends: []),
        (from: "LANOffered", to: "RelayBackoff", on: "offer_timeout", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "LANActive", to: "LANActive", on: "ping_tick", onKind: "internal", guard: nil, action: nil, sends: [(to: "client", msg: "path_ping")]),
        (from: "LANActive", to: "LANDegraded", on: "ping_timeout", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "LANDegraded", to: "LANDegraded", on: "ping_tick", onKind: "internal", guard: nil, action: nil, sends: [(to: "client", msg: "path_ping")]),
        (from: "LANDegraded", to: "LANActive", on: "path_pong", onKind: "recv", guard: nil, action: "reset_failures", sends: []),
        (from: "LANDegraded", to: "LANDegraded", on: "ping_timeout", onKind: "internal", guard: "under_max_failures", action: nil, sends: []),
        (from: "LANDegraded", to: "RelayBackoff", on: "ping_timeout", onKind: "internal", guard: "at_max_failures", action: "fallback_to_relay", sends: []),
        (from: "RelayBackoff", to: "LANOffered", on: "backoff_expired", onKind: "internal", guard: nil, action: nil, sends: [(to: "client", msg: "lan_offer")]),
        (from: "RelayBackoff", to: "LANOffered", on: "lan_server_changed", onKind: "internal", guard: nil, action: nil, sends: [(to: "client", msg: "lan_offer")]),
        (from: "RelayConnected", to: "LANOffered", on: "readvertise_tick", onKind: "internal", guard: "lan_server_available", action: nil, sends: [(to: "client", msg: "lan_offer")]),
        (from: "RelayConnected", to: "Paired", on: "disconnect", onKind: "internal", guard: nil, action: nil, sends: []),
    ]

    /// client transitions.
    public static let clientInitial: ClientState = .idle

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
        (from: "RelayConnected", to: "RelayConnected", on: "app_send", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "RelayConnected", to: "RelayConnected", on: "relay_stream_data", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "LANConnecting", to: "LANConnecting", on: "app_send", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "LANConnecting", to: "LANConnecting", on: "relay_stream_data", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "LANVerifying", to: "LANVerifying", on: "app_send", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "LANVerifying", to: "LANVerifying", on: "relay_stream_data", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "LANActive", to: "LANActive", on: "app_send", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "LANActive", to: "LANActive", on: "lan_stream_data", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "LANActive", to: "LANActive", on: "relay_stream_data", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "RelayFallback", to: "RelayFallback", on: "app_send", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "RelayFallback", to: "RelayFallback", on: "relay_stream_data", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "RelayConnected", to: "LANConnecting", on: "lan_offer", onKind: "recv", guard: "lan_enabled", action: "dial_lan", sends: []),
        (from: "RelayConnected", to: "RelayConnected", on: "lan_offer", onKind: "recv", guard: "lan_disabled", action: nil, sends: []),
        (from: "LANConnecting", to: "LANVerifying", on: "lan_dial_ok", onKind: "internal", guard: nil, action: nil, sends: [(to: "backend", msg: "lan_verify")]),
        (from: "LANConnecting", to: "RelayConnected", on: "lan_dial_failed", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "LANVerifying", to: "LANActive", on: "lan_confirm", onKind: "recv", guard: nil, action: "activate_lan", sends: []),
        (from: "LANVerifying", to: "RelayConnected", on: "verify_timeout", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "LANActive", to: "LANActive", on: "path_ping", onKind: "recv", guard: nil, action: nil, sends: [(to: "backend", msg: "path_pong")]),
        (from: "LANActive", to: "RelayFallback", on: "lan_error", onKind: "internal", guard: nil, action: "fallback_to_relay", sends: []),
        (from: "RelayFallback", to: "RelayConnected", on: "relay_ok", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "LANActive", to: "LANConnecting", on: "lan_offer", onKind: "recv", guard: "lan_enabled", action: "dial_lan", sends: []),
        (from: "RelayConnected", to: "Paired", on: "disconnect", onKind: "internal", guard: nil, action: nil, sends: []),
    ]

    /// relay transitions.
    public static let relayInitial: RelayState = .idle

    public static let relayTransitions: [(from: String, to: String, on: String, onKind: String, guard: String?, action: String?, sends: [(to: String, msg: String)])] = [
        (from: "Idle", to: "BackendRegistered", on: "backend_register", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "BackendRegistered", to: "Bridged", on: "client_connect", onKind: "internal", guard: nil, action: "bridge_streams", sends: []),
        (from: "Bridged", to: "BackendRegistered", on: "client_disconnect", onKind: "internal", guard: nil, action: "unbridge", sends: []),
        (from: "BackendRegistered", to: "Idle", on: "backend_disconnect", onKind: "internal", guard: nil, action: nil, sends: []),
    ]
}

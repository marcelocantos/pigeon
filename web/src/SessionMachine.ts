// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Auto-generated from protocol definition. Do not edit.
// Source of truth: protocol/*.yaml

export enum MessageType {
    PairHello = "pair_hello",
    PairHelloAck = "pair_hello_ack",
    PairConfirm = "pair_confirm",
    PairComplete = "pair_complete",
    AuthRequest = "auth_request",
    AuthOk = "auth_ok",
    LanOffer = "lan_offer",
    LanVerify = "lan_verify",
    LanConfirm = "lan_confirm",
    PathPing = "path_ping",
    PathPong = "path_pong",
}

export enum BackendState {
    Idle = "Idle",
    GenerateToken = "GenerateToken",
    RegisterRelay = "RegisterRelay",
    WaitingForClient = "WaitingForClient",
    DeriveSecret = "DeriveSecret",
    SendAck = "SendAck",
    WaitingForCode = "WaitingForCode",
    ValidateCode = "ValidateCode",
    StorePaired = "StorePaired",
    Paired = "Paired",
    AuthCheck = "AuthCheck",
    SessionActive = "SessionActive",
    RelayConnected = "RelayConnected",
    LANOffered = "LANOffered",
    LANActive = "LANActive",
    RelayBackoff = "RelayBackoff",
    LANDegraded = "LANDegraded",
}

export enum ClientState {
    Idle = "Idle",
    ObtainBackchannelSecret = "ObtainBackchannelSecret",
    ConnectRelay = "ConnectRelay",
    GenKeyPair = "GenKeyPair",
    WaitAck = "WaitAck",
    E2EReady = "E2EReady",
    ShowCode = "ShowCode",
    WaitPairComplete = "WaitPairComplete",
    Paired = "Paired",
    Reconnect = "Reconnect",
    SendAuth = "SendAuth",
    SessionActive = "SessionActive",
    RelayConnected = "RelayConnected",
    LANConnecting = "LANConnecting",
    LANVerifying = "LANVerifying",
    LANActive = "LANActive",
    RelayFallback = "RelayFallback",
}

export enum RelayState {
    Idle = "Idle",
    BackendRegistered = "BackendRegistered",
    Bridged = "Bridged",
}

export enum GuardID {
    TokenValid = "token_valid",
    TokenInvalid = "token_invalid",
    CodeCorrect = "code_correct",
    CodeWrong = "code_wrong",
    DeviceKnown = "device_known",
    DeviceUnknown = "device_unknown",
    NonceFresh = "nonce_fresh",
    ChallengeValid = "challenge_valid",
    ChallengeInvalid = "challenge_invalid",
    LanEnabled = "lan_enabled",
    LanDisabled = "lan_disabled",
    LanServerAvailable = "lan_server_available",
    UnderMaxFailures = "under_max_failures",
    AtMaxFailures = "at_max_failures",
}

export enum ActionID {
    GenerateToken = "generate_token",
    RegisterRelay = "register_relay",
    DeriveSecret = "derive_secret",
    StoreDevice = "store_device",
    VerifyDevice = "verify_device",
    ActivateLan = "activate_lan",
    ResetFailures = "reset_failures",
    FallbackToRelay = "fallback_to_relay",
    SendPairHello = "send_pair_hello",
    StoreSecret = "store_secret",
    DialLan = "dial_lan",
    BridgeStreams = "bridge_streams",
    Unbridge = "unbridge",
}

export interface Transition {
    readonly from: string;
    readonly to: string;
    readonly on: string;
    readonly onKind: "recv" | "internal";
    readonly guard?: string;
    readonly action?: string;
    readonly sends?: ReadonlyArray<{ readonly to: string; readonly msg: string }>;
}

export interface ActorTable {
    readonly initial: string;
    readonly transitions: ReadonlyArray<Transition>;
}

/** backend transition table. */
export const backendTable: ActorTable = {
    initial: BackendState.Idle,
    transitions: [
        { from: "Idle", to: "GenerateToken", on: "cli_init_pair", onKind: "internal", action: "generate_token" },
        { from: "GenerateToken", to: "RegisterRelay", on: "token_created", onKind: "internal", action: "register_relay" },
        { from: "RegisterRelay", to: "WaitingForClient", on: "relay_registered", onKind: "internal" },
        { from: "WaitingForClient", to: "DeriveSecret", on: "pair_hello", onKind: "recv", guard: "token_valid", action: "derive_secret" },
        { from: "WaitingForClient", to: "Idle", on: "pair_hello", onKind: "recv", guard: "token_invalid" },
        { from: "DeriveSecret", to: "SendAck", on: "ecdh_complete", onKind: "internal", sends: [{ to: "client", msg: "pair_hello_ack" }] },
        { from: "SendAck", to: "WaitingForCode", on: "signal_code_display", onKind: "internal", sends: [{ to: "client", msg: "pair_confirm" }] },
        { from: "WaitingForCode", to: "ValidateCode", on: "cli_code_entered", onKind: "internal" },
        { from: "ValidateCode", to: "StorePaired", on: "check_code", onKind: "internal", guard: "code_correct" },
        { from: "ValidateCode", to: "Idle", on: "check_code", onKind: "internal", guard: "code_wrong" },
        { from: "StorePaired", to: "Paired", on: "finalise", onKind: "internal", action: "store_device", sends: [{ to: "client", msg: "pair_complete" }] },
        { from: "Paired", to: "AuthCheck", on: "auth_request", onKind: "recv" },
        { from: "AuthCheck", to: "SessionActive", on: "verify", onKind: "internal", guard: "device_known", action: "verify_device", sends: [{ to: "client", msg: "auth_ok" }] },
        { from: "AuthCheck", to: "Idle", on: "verify", onKind: "internal", guard: "device_unknown" },
        { from: "SessionActive", to: "RelayConnected", on: "session_established", onKind: "internal" },
        { from: "RelayConnected", to: "LANOffered", on: "lan_server_ready", onKind: "internal", sends: [{ to: "client", msg: "lan_offer" }] },
        { from: "LANOffered", to: "LANActive", on: "lan_verify", onKind: "recv", guard: "challenge_valid", action: "activate_lan", sends: [{ to: "client", msg: "lan_confirm" }] },
        { from: "LANOffered", to: "RelayConnected", on: "lan_verify", onKind: "recv", guard: "challenge_invalid" },
        { from: "LANOffered", to: "RelayBackoff", on: "offer_timeout", onKind: "internal" },
        { from: "LANActive", to: "LANActive", on: "ping_tick", onKind: "internal", sends: [{ to: "client", msg: "path_ping" }] },
        { from: "LANActive", to: "LANDegraded", on: "ping_timeout", onKind: "internal" },
        { from: "LANDegraded", to: "LANDegraded", on: "ping_tick", onKind: "internal", sends: [{ to: "client", msg: "path_ping" }] },
        { from: "LANDegraded", to: "LANActive", on: "path_pong", onKind: "recv", action: "reset_failures" },
        { from: "LANDegraded", to: "LANDegraded", on: "ping_timeout", onKind: "internal", guard: "under_max_failures" },
        { from: "LANDegraded", to: "RelayBackoff", on: "ping_timeout", onKind: "internal", guard: "at_max_failures", action: "fallback_to_relay" },
        { from: "RelayBackoff", to: "LANOffered", on: "backoff_expired", onKind: "internal", sends: [{ to: "client", msg: "lan_offer" }] },
        { from: "RelayBackoff", to: "LANOffered", on: "lan_server_changed", onKind: "internal", sends: [{ to: "client", msg: "lan_offer" }] },
        { from: "RelayConnected", to: "LANOffered", on: "readvertise_tick", onKind: "internal", guard: "lan_server_available", sends: [{ to: "client", msg: "lan_offer" }] },
        { from: "RelayConnected", to: "Paired", on: "disconnect", onKind: "internal" },
    ],
};

/** client transition table. */
export const clientTable: ActorTable = {
    initial: ClientState.Idle,
    transitions: [
        { from: "Idle", to: "ObtainBackchannelSecret", on: "backchannel_received", onKind: "internal" },
        { from: "ObtainBackchannelSecret", to: "ConnectRelay", on: "secret_parsed", onKind: "internal" },
        { from: "ConnectRelay", to: "GenKeyPair", on: "relay_connected", onKind: "internal" },
        { from: "GenKeyPair", to: "WaitAck", on: "key_pair_generated", onKind: "internal", action: "send_pair_hello", sends: [{ to: "backend", msg: "pair_hello" }] },
        { from: "WaitAck", to: "E2EReady", on: "pair_hello_ack", onKind: "recv", action: "derive_secret" },
        { from: "E2EReady", to: "ShowCode", on: "pair_confirm", onKind: "recv" },
        { from: "ShowCode", to: "WaitPairComplete", on: "code_displayed", onKind: "internal" },
        { from: "WaitPairComplete", to: "Paired", on: "pair_complete", onKind: "recv", action: "store_secret" },
        { from: "Paired", to: "Reconnect", on: "app_launch", onKind: "internal" },
        { from: "Reconnect", to: "SendAuth", on: "relay_connected", onKind: "internal", sends: [{ to: "backend", msg: "auth_request" }] },
        { from: "SendAuth", to: "SessionActive", on: "auth_ok", onKind: "recv" },
        { from: "SessionActive", to: "RelayConnected", on: "session_established", onKind: "internal" },
        { from: "RelayConnected", to: "LANConnecting", on: "lan_offer", onKind: "recv", guard: "lan_enabled", action: "dial_lan" },
        { from: "RelayConnected", to: "RelayConnected", on: "lan_offer", onKind: "recv", guard: "lan_disabled" },
        { from: "LANConnecting", to: "LANVerifying", on: "lan_dial_ok", onKind: "internal", sends: [{ to: "backend", msg: "lan_verify" }] },
        { from: "LANConnecting", to: "RelayConnected", on: "lan_dial_failed", onKind: "internal" },
        { from: "LANVerifying", to: "LANActive", on: "lan_confirm", onKind: "recv", action: "activate_lan" },
        { from: "LANVerifying", to: "RelayConnected", on: "verify_timeout", onKind: "internal" },
        { from: "LANActive", to: "LANActive", on: "path_ping", onKind: "recv", sends: [{ to: "backend", msg: "path_pong" }] },
        { from: "LANActive", to: "RelayFallback", on: "lan_error", onKind: "internal", action: "fallback_to_relay" },
        { from: "RelayFallback", to: "RelayConnected", on: "relay_ok", onKind: "internal" },
        { from: "LANActive", to: "LANConnecting", on: "lan_offer", onKind: "recv", guard: "lan_enabled", action: "dial_lan" },
        { from: "RelayConnected", to: "Paired", on: "disconnect", onKind: "internal" },
    ],
};

/** relay transition table. */
export const relayTable: ActorTable = {
    initial: RelayState.Idle,
    transitions: [
        { from: "Idle", to: "BackendRegistered", on: "backend_register", onKind: "internal" },
        { from: "BackendRegistered", to: "Bridged", on: "client_connect", onKind: "internal", action: "bridge_streams" },
        { from: "Bridged", to: "BackendRegistered", on: "client_disconnect", onKind: "internal", action: "unbridge" },
        { from: "BackendRegistered", to: "Idle", on: "backend_disconnect", onKind: "internal" },
    ],
};

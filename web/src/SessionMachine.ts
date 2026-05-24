// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Auto-generated from protocol definition. Do not edit.
// Source of truth: protocol/*.yaml

export enum SessionBackendState {
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
    CandidatesAdvertised = "CandidatesAdvertised",
    AltActive = "AltActive",
    RelayBackoff = "RelayBackoff",
    AltDegraded = "AltDegraded",
}

export enum SessionClientState {
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
    PairDialing = "PairDialing",
    PairChecking = "PairChecking",
    AltActive = "AltActive",
    RelayFallback = "RelayFallback",
}

export enum SessionRelayState {
    Idle = "Idle",
    BackendRegistered = "BackendRegistered",
    Bridged = "Bridged",
}

/** The protocol transition table and shared type enums. */
export namespace SessionProtocol {

    export enum MessageType {
        PairHello = "pair_hello",
        PairHelloAck = "pair_hello_ack",
        PairConfirm = "pair_confirm",
        PairComplete = "pair_complete",
        AuthRequest = "auth_request",
        AuthOk = "auth_ok",
        Candidates = "candidates",
        PairCheck = "pair_check",
        PairCheckAck = "pair_check_ack",
        PathPing = "path_ping",
        PathPong = "path_pong",
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
        AltEnabled = "alt_enabled",
        AltDisabled = "alt_disabled",
        LocalCandidatesAvailable = "local_candidates_available",
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
        FallbackToRelay = "fallback_to_relay",
        ResetFailures = "reset_failures",
        SendPairHello = "send_pair_hello",
        StoreSecret = "store_secret",
        DialCandidate = "dial_candidate",
        BridgeStreams = "bridge_streams",
        Unbridge = "unbridge",
    }

    export enum EventID {
        AppSend = "app_send",
        AppRecv = "app_recv",
        AppSendDatagram = "app_send_datagram",
        AppRecvDatagram = "app_recv_datagram",
        AppClose = "app_close",
        AppForceFallback = "app_force_fallback",
        RelayStreamData = "relay_stream_data",
        RelayStreamError = "relay_stream_error",
        RelayDatagram = "relay_datagram",
        AltStreamData = "alt_stream_data",
        AltStreamError = "alt_stream_error",
        AltDatagram = "alt_datagram",
        DialOk = "dial_ok",
        DialFailed = "dial_failed",
        PairCheckOk = "pair_check_ok",
        PingTimeout = "ping_timeout",
        PingTick = "ping_tick",
        BackoffExpired = "backoff_expired",
        CandidatesTimeout = "candidates_timeout",
        CliInitPair = "cli_init_pair",
        TokenCreated = "token_created",
        RelayRegistered = "relay_registered",
        EcdhComplete = "ecdh_complete",
        SignalCodeDisplay = "signal_code_display",
        CliCodeEntered = "cli_code_entered",
        CheckCode = "check_code",
        Finalise = "finalise",
        Verify = "verify",
        SessionEstablished = "session_established",
        CandidatesGathered = "candidates_gathered",
        CandidatesChanged = "candidates_changed",
        CandidatesRefreshTick = "candidates_refresh_tick",
        Disconnect = "disconnect",
        BackchannelReceived = "backchannel_received",
        SecretParsed = "secret_parsed",
        RelayConnected = "relay_connected",
        KeyPairGenerated = "key_pair_generated",
        CodeDisplayed = "code_displayed",
        AppLaunch = "app_launch",
        VerifyTimeout = "verify_timeout",
        AltError = "alt_error",
        RelayOk = "relay_ok",
        BackendRegister = "backend_register",
        ClientConnect = "client_connect",
        ClientDisconnect = "client_disconnect",
        BackendDisconnect = "backend_disconnect",
        RecvPairHello = "recv_pair_hello",
        RecvAuthRequest = "recv_auth_request",
        RecvPairCheck = "recv_pair_check",
        RecvPathPong = "recv_path_pong",
        RecvPairHelloAck = "recv_pair_hello_ack",
        RecvPairConfirm = "recv_pair_confirm",
        RecvPairComplete = "recv_pair_complete",
        RecvAuthOk = "recv_auth_ok",
        RecvCandidates = "recv_candidates",
        RecvPairCheckAck = "recv_pair_check_ack",
        RecvPathPing = "recv_path_ping",
    }

    export enum CmdID {
        WriteActiveStream = "write_active_stream",
        SendActiveDatagram = "send_active_datagram",
        SendPathPing = "send_path_ping",
        SendPathPong = "send_path_pong",
        SendCandidates = "send_candidates",
        SendPairCheck = "send_pair_check",
        SendPairCheckAck = "send_pair_check_ack",
        DialCandidate = "dial_candidate",
        DeliverRecv = "deliver_recv",
        DeliverRecvError = "deliver_recv_error",
        DeliverRecvDatagram = "deliver_recv_datagram",
        StartAltStreamReader = "start_alt_stream_reader",
        StopAltStreamReader = "stop_alt_stream_reader",
        StartAltDgReader = "start_alt_dg_reader",
        StopAltDgReader = "stop_alt_dg_reader",
        StartMonitor = "start_monitor",
        StopMonitor = "stop_monitor",
        StartPongTimeout = "start_pong_timeout",
        CancelPongTimeout = "cancel_pong_timeout",
        StartBackoffTimer = "start_backoff_timer",
        CloseAltPath = "close_alt_path",
        SignalAltReady = "signal_alt_ready",
        ResetAltReady = "reset_alt_ready",
        SetCryptoDatagram = "set_crypto_datagram",
    }

    /** Protocol wire constants shared across all platforms. */
    export const Wire = {
        DG_CONN_WHOLE: 0x00,
        DG_PING: 0x10,
        DG_PONG: 0x11,
        DG_CONN_FRAGMENT: 0x40,
        DG_CHAN_WHOLE: 0x80,
        DG_CHAN_FRAGMENT: 0xC0,
        FRAG_HEADER_SIZE: 8,
        CHAN_ID_SIZE: 2,
        MAX_DATAGRAM_PAYLOAD: 1200,
        FRAGMENT_TIMEOUT_MS: 5000, // ms
        FRAME_APP: 0x00,
        FRAME_CANDIDATES: 0x01,
        FRAME_CUTOVER: 0x02,
        FRAME_PAIR_CHECK: 0x03,
        FRAME_PAIR_CHECK_ACK: 0x04,
        CAND_HOST: "host",
        CAND_SRFLX: "srflx",
        MAX_MESSAGE_SIZE: 1048576,
        LENGTH_PREFIX_SIZE: 4,
        PING_INTERVAL_MS: 5000, // ms
        PONG_TIMEOUT_MS: 4000, // ms
        MAX_PING_FAILURES: 3,
        MAX_BACKOFF_LEVEL: 5,
        STREAM_CHANNEL_OPENER_SUFFIX: ":o2a",
        STREAM_CHANNEL_ACCEPT_SUFFIX: ":a2o",
        DG_CHANNEL_SEND_SUFFIX: ":dg:send",
        DG_CHANNEL_RECV_SUFFIX: ":dg:recv",
        CHANNEL_ID_HASH_MULTIPLIER: 31,
    } as const;

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
        initial: SessionBackendState.Idle,
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
            { from: "RelayConnected", to: "CandidatesAdvertised", on: "candidates_gathered", onKind: "internal", sends: [{ to: "client", msg: "candidates" }] },
            { from: "CandidatesAdvertised", to: "AltActive", on: "pair_check", onKind: "recv", guard: "challenge_valid", action: "activate_lan", sends: [{ to: "client", msg: "pair_check_ack" }] },
            { from: "CandidatesAdvertised", to: "RelayConnected", on: "pair_check", onKind: "recv", guard: "challenge_invalid" },
            { from: "CandidatesAdvertised", to: "RelayBackoff", on: "candidates_timeout", onKind: "internal" },
            { from: "AltActive", to: "AltActive", on: "ping_tick", onKind: "internal", sends: [{ to: "client", msg: "path_ping" }] },
            { from: "AltActive", to: "AltDegraded", on: "ping_timeout", onKind: "internal" },
            { from: "AltDegraded", to: "AltDegraded", on: "ping_tick", onKind: "internal", sends: [{ to: "client", msg: "path_ping" }] },
            { from: "AltActive", to: "RelayBackoff", on: "alt_stream_error", onKind: "internal", action: "fallback_to_relay" },
            { from: "AltDegraded", to: "RelayBackoff", on: "alt_stream_error", onKind: "internal", action: "fallback_to_relay" },
            { from: "AltDegraded", to: "AltActive", on: "path_pong", onKind: "recv", action: "reset_failures" },
            { from: "AltDegraded", to: "AltDegraded", on: "ping_timeout", onKind: "internal", guard: "under_max_failures" },
            { from: "AltDegraded", to: "RelayBackoff", on: "ping_timeout", onKind: "internal", guard: "at_max_failures", action: "fallback_to_relay" },
            { from: "RelayBackoff", to: "CandidatesAdvertised", on: "backoff_expired", onKind: "internal", sends: [{ to: "client", msg: "candidates" }] },
            { from: "RelayBackoff", to: "CandidatesAdvertised", on: "candidates_changed", onKind: "internal", sends: [{ to: "client", msg: "candidates" }] },
            { from: "RelayConnected", to: "CandidatesAdvertised", on: "candidates_refresh_tick", onKind: "internal", guard: "local_candidates_available", sends: [{ to: "client", msg: "candidates" }] },
            { from: "CandidatesAdvertised", to: "RelayConnected", on: "app_force_fallback", onKind: "internal" },
            { from: "AltActive", to: "RelayBackoff", on: "app_force_fallback", onKind: "internal", action: "fallback_to_relay" },
            { from: "AltDegraded", to: "RelayBackoff", on: "app_force_fallback", onKind: "internal", action: "fallback_to_relay" },
            { from: "RelayConnected", to: "Paired", on: "disconnect", onKind: "internal" },
            { from: "RelayConnected", to: "RelayConnected", on: "app_send", onKind: "internal" },
            { from: "CandidatesAdvertised", to: "CandidatesAdvertised", on: "app_send", onKind: "internal" },
            { from: "AltActive", to: "AltActive", on: "app_send", onKind: "internal" },
            { from: "AltDegraded", to: "AltDegraded", on: "app_send", onKind: "internal" },
            { from: "RelayBackoff", to: "RelayBackoff", on: "app_send", onKind: "internal" },
            { from: "RelayConnected", to: "RelayConnected", on: "relay_stream_data", onKind: "internal" },
            { from: "CandidatesAdvertised", to: "CandidatesAdvertised", on: "relay_stream_data", onKind: "internal" },
            { from: "AltActive", to: "AltActive", on: "relay_stream_data", onKind: "internal" },
            { from: "AltDegraded", to: "AltDegraded", on: "relay_stream_data", onKind: "internal" },
            { from: "RelayBackoff", to: "RelayBackoff", on: "relay_stream_data", onKind: "internal" },
            { from: "RelayConnected", to: "RelayConnected", on: "relay_stream_error", onKind: "internal" },
            { from: "CandidatesAdvertised", to: "CandidatesAdvertised", on: "relay_stream_error", onKind: "internal" },
            { from: "AltActive", to: "AltActive", on: "relay_stream_error", onKind: "internal" },
            { from: "AltDegraded", to: "AltDegraded", on: "relay_stream_error", onKind: "internal" },
            { from: "RelayBackoff", to: "RelayBackoff", on: "relay_stream_error", onKind: "internal" },
            { from: "RelayConnected", to: "RelayConnected", on: "app_send_datagram", onKind: "internal" },
            { from: "CandidatesAdvertised", to: "CandidatesAdvertised", on: "app_send_datagram", onKind: "internal" },
            { from: "AltActive", to: "AltActive", on: "app_send_datagram", onKind: "internal" },
            { from: "AltDegraded", to: "AltDegraded", on: "app_send_datagram", onKind: "internal" },
            { from: "RelayBackoff", to: "RelayBackoff", on: "app_send_datagram", onKind: "internal" },
            { from: "RelayConnected", to: "RelayConnected", on: "relay_datagram", onKind: "internal" },
            { from: "CandidatesAdvertised", to: "CandidatesAdvertised", on: "relay_datagram", onKind: "internal" },
            { from: "AltActive", to: "AltActive", on: "relay_datagram", onKind: "internal" },
            { from: "AltDegraded", to: "AltDegraded", on: "relay_datagram", onKind: "internal" },
            { from: "RelayBackoff", to: "RelayBackoff", on: "relay_datagram", onKind: "internal" },
            { from: "AltActive", to: "AltActive", on: "alt_stream_data", onKind: "internal" },
            { from: "AltDegraded", to: "AltDegraded", on: "alt_stream_data", onKind: "internal" },
            { from: "AltActive", to: "AltActive", on: "alt_datagram", onKind: "internal" },
            { from: "AltDegraded", to: "AltDegraded", on: "alt_datagram", onKind: "internal" },
        ],
    };

    /** client transition table. */
    export const clientTable: ActorTable = {
        initial: SessionClientState.Idle,
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
            { from: "RelayConnected", to: "PairDialing", on: "candidates", onKind: "recv", guard: "alt_enabled", action: "dial_candidate" },
            { from: "RelayConnected", to: "RelayConnected", on: "candidates", onKind: "recv", guard: "alt_disabled" },
            { from: "PairDialing", to: "PairChecking", on: "dial_ok", onKind: "internal", sends: [{ to: "backend", msg: "pair_check" }] },
            { from: "PairDialing", to: "RelayConnected", on: "dial_failed", onKind: "internal" },
            { from: "PairChecking", to: "AltActive", on: "pair_check_ack", onKind: "recv", action: "activate_lan" },
            { from: "PairChecking", to: "RelayConnected", on: "verify_timeout", onKind: "internal" },
            { from: "AltActive", to: "AltActive", on: "path_ping", onKind: "recv", sends: [{ to: "backend", msg: "path_pong" }] },
            { from: "AltActive", to: "RelayFallback", on: "alt_error", onKind: "internal", action: "fallback_to_relay" },
            { from: "AltActive", to: "RelayFallback", on: "alt_stream_error", onKind: "internal", action: "fallback_to_relay" },
            { from: "RelayFallback", to: "RelayConnected", on: "relay_ok", onKind: "internal" },
            { from: "AltActive", to: "PairDialing", on: "candidates", onKind: "recv", guard: "alt_enabled", action: "dial_candidate" },
            { from: "PairDialing", to: "RelayConnected", on: "app_force_fallback", onKind: "internal" },
            { from: "PairChecking", to: "RelayConnected", on: "app_force_fallback", onKind: "internal" },
            { from: "AltActive", to: "RelayConnected", on: "app_force_fallback", onKind: "internal", action: "fallback_to_relay" },
            { from: "RelayConnected", to: "Paired", on: "disconnect", onKind: "internal" },
            { from: "RelayConnected", to: "RelayConnected", on: "app_send", onKind: "internal" },
            { from: "PairDialing", to: "PairDialing", on: "app_send", onKind: "internal" },
            { from: "PairChecking", to: "PairChecking", on: "app_send", onKind: "internal" },
            { from: "AltActive", to: "AltActive", on: "app_send", onKind: "internal" },
            { from: "RelayFallback", to: "RelayFallback", on: "app_send", onKind: "internal" },
            { from: "RelayConnected", to: "RelayConnected", on: "relay_stream_data", onKind: "internal" },
            { from: "PairDialing", to: "PairDialing", on: "relay_stream_data", onKind: "internal" },
            { from: "PairChecking", to: "PairChecking", on: "relay_stream_data", onKind: "internal" },
            { from: "AltActive", to: "AltActive", on: "relay_stream_data", onKind: "internal" },
            { from: "RelayFallback", to: "RelayFallback", on: "relay_stream_data", onKind: "internal" },
            { from: "RelayConnected", to: "RelayConnected", on: "relay_stream_error", onKind: "internal" },
            { from: "PairDialing", to: "PairDialing", on: "relay_stream_error", onKind: "internal" },
            { from: "PairChecking", to: "PairChecking", on: "relay_stream_error", onKind: "internal" },
            { from: "AltActive", to: "AltActive", on: "relay_stream_error", onKind: "internal" },
            { from: "RelayFallback", to: "RelayFallback", on: "relay_stream_error", onKind: "internal" },
            { from: "RelayConnected", to: "RelayConnected", on: "app_send_datagram", onKind: "internal" },
            { from: "PairDialing", to: "PairDialing", on: "app_send_datagram", onKind: "internal" },
            { from: "PairChecking", to: "PairChecking", on: "app_send_datagram", onKind: "internal" },
            { from: "AltActive", to: "AltActive", on: "app_send_datagram", onKind: "internal" },
            { from: "RelayFallback", to: "RelayFallback", on: "app_send_datagram", onKind: "internal" },
            { from: "RelayConnected", to: "RelayConnected", on: "relay_datagram", onKind: "internal" },
            { from: "PairDialing", to: "PairDialing", on: "relay_datagram", onKind: "internal" },
            { from: "PairChecking", to: "PairChecking", on: "relay_datagram", onKind: "internal" },
            { from: "AltActive", to: "AltActive", on: "relay_datagram", onKind: "internal" },
            { from: "RelayFallback", to: "RelayFallback", on: "relay_datagram", onKind: "internal" },
            { from: "AltActive", to: "AltActive", on: "alt_stream_data", onKind: "internal" },
            { from: "AltActive", to: "AltActive", on: "alt_datagram", onKind: "internal" },
        ],
    };

    /** relay transition table. */
    export const relayTable: ActorTable = {
        initial: SessionRelayState.Idle,
        transitions: [
            { from: "Idle", to: "BackendRegistered", on: "backend_register", onKind: "internal" },
            { from: "BackendRegistered", to: "Bridged", on: "client_connect", onKind: "internal", action: "bridge_streams" },
            { from: "Bridged", to: "BackendRegistered", on: "client_disconnect", onKind: "internal", action: "unbridge" },
            { from: "BackendRegistered", to: "Idle", on: "backend_disconnect", onKind: "internal" },
        ],
    };

}

/** SessionBackendMachine is the generated state machine for the backend actor. */
export class SessionBackendMachine {
    readonly protocol = SessionProtocol;
    state: SessionBackendState;
    currentToken: string = "none"; // pairing token currently in play
    activeTokens: Set<string> = new Set(); // set of valid (non-revoked) tokens
    usedTokens: Set<string> = new Set(); // set of revoked tokens
    backendEcdhPub: string = "none"; // backend ECDH public key
    receivedClientPub: string = "none"; // pubkey backend received in pair_hello
    backendSharedKey: string = ""; // ECDH key derived by backend
    backendCode: string = ""; // code computed by backend
    receivedCode: string = ""; // code entered via CLI
    codeAttempts: number = 0; // failed code submission attempts
    deviceSecret: string = "none"; // persistent device secret
    pairedDevices: Set<string> = new Set(); // device IDs that completed pairing
    receivedDeviceId: string = "none"; // device_id from auth_request
    authNoncesUsed: Set<string> = new Set(); // set of consumed auth nonces
    receivedAuthNonce: string = "none"; // nonce from auth_request
    secretPublished: boolean = false; // whether token has been published via backchannel
    pingFailures: number = 0; // consecutive failed pings
    backoffLevel: number = 0; // exponential backoff level
    bActivePath: string = "relay"; // backend active path
    bDispatcherPath: string = "relay"; // backend datagram dispatcher binding
    monitorTarget: string = "none"; // health monitor target
    altSignal: string = "pending"; // AltReady notification state
    guards: Map<SessionProtocol.GuardID, () => boolean> = new Map();
    actions: Map<SessionProtocol.ActionID, () => void> = new Map();

    constructor() {
        this.state = SessionBackendState.Idle;
    }

    handleEvent(ev: SessionProtocol.EventID): SessionProtocol.CmdID[] {
        switch (true) {
            case this.state === SessionBackendState.Idle && ev === SessionProtocol.EventID.CliInitPair: {
                this.actions.get(SessionProtocol.ActionID.GenerateToken)?.();
                this.currentToken = "tok_1";
                // active_tokens: active_tokens \union {"tok_1"} (set by action)
                this.state = SessionBackendState.GenerateToken;
                return [];
            }
            case this.state === SessionBackendState.GenerateToken && ev === SessionProtocol.EventID.TokenCreated: {
                this.actions.get(SessionProtocol.ActionID.RegisterRelay)?.();
                this.state = SessionBackendState.RegisterRelay;
                return [];
            }
            case this.state === SessionBackendState.RegisterRelay && ev === SessionProtocol.EventID.RelayRegistered: {
                this.secretPublished = true;
                this.state = SessionBackendState.WaitingForClient;
                return [];
            }
            case this.state === SessionBackendState.WaitingForClient && ev === SessionProtocol.EventID.RecvPairHello && this.guards.get(SessionProtocol.GuardID.TokenValid)?.() === true: {
                this.actions.get(SessionProtocol.ActionID.DeriveSecret)?.();
                // received_client_pub: recv_msg.pubkey (set by action)
                this.backendEcdhPub = "backend_pub";
                // backend_shared_key: DeriveKey("backend_pub", recv_msg.pubkey) (set by action)
                // backend_code: DeriveCode("backend_pub", recv_msg.pubkey) (set by action)
                this.state = SessionBackendState.DeriveSecret;
                return [];
            }
            case this.state === SessionBackendState.WaitingForClient && ev === SessionProtocol.EventID.RecvPairHello && this.guards.get(SessionProtocol.GuardID.TokenInvalid)?.() === true: {
                this.state = SessionBackendState.Idle;
                return [];
            }
            case this.state === SessionBackendState.DeriveSecret && ev === SessionProtocol.EventID.EcdhComplete: {
                this.state = SessionBackendState.SendAck;
                return [];
            }
            case this.state === SessionBackendState.SendAck && ev === SessionProtocol.EventID.SignalCodeDisplay: {
                this.state = SessionBackendState.WaitingForCode;
                return [];
            }
            case this.state === SessionBackendState.WaitingForCode && ev === SessionProtocol.EventID.CliCodeEntered: {
                // received_code: cli_entered_code (set by action)
                this.state = SessionBackendState.ValidateCode;
                return [];
            }
            case this.state === SessionBackendState.ValidateCode && ev === SessionProtocol.EventID.CheckCode && this.guards.get(SessionProtocol.GuardID.CodeCorrect)?.() === true: {
                this.state = SessionBackendState.StorePaired;
                return [];
            }
            case this.state === SessionBackendState.ValidateCode && ev === SessionProtocol.EventID.CheckCode && this.guards.get(SessionProtocol.GuardID.CodeWrong)?.() === true: {
                // code_attempts: code_attempts + 1 (set by action)
                this.state = SessionBackendState.Idle;
                return [];
            }
            case this.state === SessionBackendState.StorePaired && ev === SessionProtocol.EventID.Finalise: {
                this.actions.get(SessionProtocol.ActionID.StoreDevice)?.();
                this.deviceSecret = "dev_secret_1";
                // paired_devices: paired_devices \union {"device_1"} (set by action)
                // active_tokens: active_tokens \ {current_token} (set by action)
                // used_tokens: used_tokens \union {current_token} (set by action)
                this.state = SessionBackendState.Paired;
                return [];
            }
            case this.state === SessionBackendState.Paired && ev === SessionProtocol.EventID.RecvAuthRequest: {
                // received_device_id: recv_msg.device_id (set by action)
                // received_auth_nonce: recv_msg.nonce (set by action)
                this.state = SessionBackendState.AuthCheck;
                return [];
            }
            case this.state === SessionBackendState.AuthCheck && ev === SessionProtocol.EventID.Verify && this.guards.get(SessionProtocol.GuardID.DeviceKnown)?.() === true: {
                this.actions.get(SessionProtocol.ActionID.VerifyDevice)?.();
                // auth_nonces_used: auth_nonces_used \union {received_auth_nonce} (set by action)
                this.state = SessionBackendState.SessionActive;
                return [];
            }
            case this.state === SessionBackendState.AuthCheck && ev === SessionProtocol.EventID.Verify && this.guards.get(SessionProtocol.GuardID.DeviceUnknown)?.() === true: {
                this.state = SessionBackendState.Idle;
                return [];
            }
            case this.state === SessionBackendState.SessionActive && ev === SessionProtocol.EventID.SessionEstablished: {
                this.state = SessionBackendState.RelayConnected;
                return [];
            }
            case this.state === SessionBackendState.RelayConnected && ev === SessionProtocol.EventID.CandidatesGathered: {
                this.state = SessionBackendState.CandidatesAdvertised;
                return [SessionProtocol.CmdID.SendCandidates];
            }
            case this.state === SessionBackendState.CandidatesAdvertised && ev === SessionProtocol.EventID.RecvPairCheck && this.guards.get(SessionProtocol.GuardID.ChallengeValid)?.() === true: {
                this.actions.get(SessionProtocol.ActionID.ActivateLan)?.();
                this.pingFailures = 0;
                this.backoffLevel = 0;
                this.bActivePath = "alt";
                this.bDispatcherPath = "alt";
                this.monitorTarget = "alt";
                this.altSignal = "ready";
                this.state = SessionBackendState.AltActive;
                return [SessionProtocol.CmdID.SendPairCheckAck, SessionProtocol.CmdID.StartAltStreamReader, SessionProtocol.CmdID.StartAltDgReader, SessionProtocol.CmdID.StartMonitor, SessionProtocol.CmdID.SignalAltReady, SessionProtocol.CmdID.SetCryptoDatagram];
            }
            case this.state === SessionBackendState.CandidatesAdvertised && ev === SessionProtocol.EventID.RecvPairCheck && this.guards.get(SessionProtocol.GuardID.ChallengeInvalid)?.() === true: {
                this.state = SessionBackendState.RelayConnected;
                return [];
            }
            case this.state === SessionBackendState.CandidatesAdvertised && ev === SessionProtocol.EventID.CandidatesTimeout: {
                // backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
                this.altSignal = "pending";
                this.state = SessionBackendState.RelayBackoff;
                return [SessionProtocol.CmdID.ResetAltReady, SessionProtocol.CmdID.StartBackoffTimer];
            }
            case this.state === SessionBackendState.AltActive && ev === SessionProtocol.EventID.PingTick: {
                this.state = SessionBackendState.AltActive;
                return [SessionProtocol.CmdID.SendPathPing, SessionProtocol.CmdID.StartPongTimeout];
            }
            case this.state === SessionBackendState.AltActive && ev === SessionProtocol.EventID.PingTimeout: {
                this.pingFailures = 1;
                this.state = SessionBackendState.AltDegraded;
                return [];
            }
            case this.state === SessionBackendState.AltDegraded && ev === SessionProtocol.EventID.PingTick: {
                this.state = SessionBackendState.AltDegraded;
                return [SessionProtocol.CmdID.SendPathPing, SessionProtocol.CmdID.StartPongTimeout];
            }
            case this.state === SessionBackendState.AltActive && ev === SessionProtocol.EventID.AltStreamError: {
                this.actions.get(SessionProtocol.ActionID.FallbackToRelay)?.();
                // backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
                this.bActivePath = "relay";
                this.bDispatcherPath = "relay";
                this.monitorTarget = "none";
                this.altSignal = "pending";
                this.pingFailures = 0;
                this.state = SessionBackendState.RelayBackoff;
                return [SessionProtocol.CmdID.StopMonitor, SessionProtocol.CmdID.StopAltStreamReader, SessionProtocol.CmdID.StopAltDgReader, SessionProtocol.CmdID.CloseAltPath, SessionProtocol.CmdID.ResetAltReady, SessionProtocol.CmdID.StartBackoffTimer];
            }
            case this.state === SessionBackendState.AltDegraded && ev === SessionProtocol.EventID.AltStreamError: {
                this.actions.get(SessionProtocol.ActionID.FallbackToRelay)?.();
                // backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
                this.bActivePath = "relay";
                this.bDispatcherPath = "relay";
                this.monitorTarget = "none";
                this.altSignal = "pending";
                this.pingFailures = 0;
                this.state = SessionBackendState.RelayBackoff;
                return [SessionProtocol.CmdID.StopMonitor, SessionProtocol.CmdID.StopAltStreamReader, SessionProtocol.CmdID.StopAltDgReader, SessionProtocol.CmdID.CloseAltPath, SessionProtocol.CmdID.ResetAltReady, SessionProtocol.CmdID.StartBackoffTimer];
            }
            case this.state === SessionBackendState.AltDegraded && ev === SessionProtocol.EventID.RecvPathPong: {
                this.actions.get(SessionProtocol.ActionID.ResetFailures)?.();
                this.pingFailures = 0;
                this.state = SessionBackendState.AltActive;
                return [SessionProtocol.CmdID.CancelPongTimeout];
            }
            case this.state === SessionBackendState.AltDegraded && ev === SessionProtocol.EventID.PingTimeout && this.guards.get(SessionProtocol.GuardID.UnderMaxFailures)?.() === true: {
                // ping_failures: ping_failures + 1 (set by action)
                this.state = SessionBackendState.AltDegraded;
                return [];
            }
            case this.state === SessionBackendState.AltDegraded && ev === SessionProtocol.EventID.PingTimeout && this.guards.get(SessionProtocol.GuardID.AtMaxFailures)?.() === true: {
                this.actions.get(SessionProtocol.ActionID.FallbackToRelay)?.();
                // backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
                this.bActivePath = "relay";
                this.bDispatcherPath = "relay";
                this.monitorTarget = "none";
                this.altSignal = "pending";
                this.pingFailures = 0;
                this.state = SessionBackendState.RelayBackoff;
                return [SessionProtocol.CmdID.StopMonitor, SessionProtocol.CmdID.StopAltStreamReader, SessionProtocol.CmdID.StopAltDgReader, SessionProtocol.CmdID.CloseAltPath, SessionProtocol.CmdID.ResetAltReady, SessionProtocol.CmdID.StartBackoffTimer];
            }
            case this.state === SessionBackendState.RelayBackoff && ev === SessionProtocol.EventID.BackoffExpired: {
                this.state = SessionBackendState.CandidatesAdvertised;
                return [SessionProtocol.CmdID.SendCandidates];
            }
            case this.state === SessionBackendState.RelayBackoff && ev === SessionProtocol.EventID.CandidatesChanged: {
                this.backoffLevel = 0;
                this.state = SessionBackendState.CandidatesAdvertised;
                return [SessionProtocol.CmdID.SendCandidates];
            }
            case this.state === SessionBackendState.RelayConnected && ev === SessionProtocol.EventID.CandidatesRefreshTick && this.guards.get(SessionProtocol.GuardID.LocalCandidatesAvailable)?.() === true: {
                this.state = SessionBackendState.CandidatesAdvertised;
                return [SessionProtocol.CmdID.SendCandidates];
            }
            case this.state === SessionBackendState.CandidatesAdvertised && ev === SessionProtocol.EventID.AppForceFallback: {
                this.altSignal = "pending";
                this.state = SessionBackendState.RelayConnected;
                return [SessionProtocol.CmdID.ResetAltReady];
            }
            case this.state === SessionBackendState.AltActive && ev === SessionProtocol.EventID.AppForceFallback: {
                this.actions.get(SessionProtocol.ActionID.FallbackToRelay)?.();
                // backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
                this.bActivePath = "relay";
                this.bDispatcherPath = "relay";
                this.monitorTarget = "none";
                this.altSignal = "pending";
                this.pingFailures = 0;
                this.state = SessionBackendState.RelayBackoff;
                return [SessionProtocol.CmdID.StopMonitor, SessionProtocol.CmdID.CancelPongTimeout, SessionProtocol.CmdID.StopAltStreamReader, SessionProtocol.CmdID.StopAltDgReader, SessionProtocol.CmdID.CloseAltPath, SessionProtocol.CmdID.ResetAltReady, SessionProtocol.CmdID.StartBackoffTimer];
            }
            case this.state === SessionBackendState.AltDegraded && ev === SessionProtocol.EventID.AppForceFallback: {
                this.actions.get(SessionProtocol.ActionID.FallbackToRelay)?.();
                // backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
                this.bActivePath = "relay";
                this.bDispatcherPath = "relay";
                this.monitorTarget = "none";
                this.altSignal = "pending";
                this.pingFailures = 0;
                this.state = SessionBackendState.RelayBackoff;
                return [SessionProtocol.CmdID.StopMonitor, SessionProtocol.CmdID.CancelPongTimeout, SessionProtocol.CmdID.StopAltStreamReader, SessionProtocol.CmdID.StopAltDgReader, SessionProtocol.CmdID.CloseAltPath, SessionProtocol.CmdID.ResetAltReady, SessionProtocol.CmdID.StartBackoffTimer];
            }
            case this.state === SessionBackendState.RelayConnected && ev === SessionProtocol.EventID.Disconnect: {
                this.state = SessionBackendState.Paired;
                return [];
            }
            case this.state === SessionBackendState.RelayConnected && ev === SessionProtocol.EventID.AppSend: {
                this.state = SessionBackendState.RelayConnected;
                return [SessionProtocol.CmdID.WriteActiveStream];
            }
            case this.state === SessionBackendState.CandidatesAdvertised && ev === SessionProtocol.EventID.AppSend: {
                this.state = SessionBackendState.CandidatesAdvertised;
                return [SessionProtocol.CmdID.WriteActiveStream];
            }
            case this.state === SessionBackendState.AltActive && ev === SessionProtocol.EventID.AppSend: {
                this.state = SessionBackendState.AltActive;
                return [SessionProtocol.CmdID.WriteActiveStream];
            }
            case this.state === SessionBackendState.AltDegraded && ev === SessionProtocol.EventID.AppSend: {
                this.state = SessionBackendState.AltDegraded;
                return [SessionProtocol.CmdID.WriteActiveStream];
            }
            case this.state === SessionBackendState.RelayBackoff && ev === SessionProtocol.EventID.AppSend: {
                this.state = SessionBackendState.RelayBackoff;
                return [SessionProtocol.CmdID.WriteActiveStream];
            }
            case this.state === SessionBackendState.RelayConnected && ev === SessionProtocol.EventID.RelayStreamData: {
                this.state = SessionBackendState.RelayConnected;
                return [SessionProtocol.CmdID.DeliverRecv];
            }
            case this.state === SessionBackendState.CandidatesAdvertised && ev === SessionProtocol.EventID.RelayStreamData: {
                this.state = SessionBackendState.CandidatesAdvertised;
                return [SessionProtocol.CmdID.DeliverRecv];
            }
            case this.state === SessionBackendState.AltActive && ev === SessionProtocol.EventID.RelayStreamData: {
                this.state = SessionBackendState.AltActive;
                return [SessionProtocol.CmdID.DeliverRecv];
            }
            case this.state === SessionBackendState.AltDegraded && ev === SessionProtocol.EventID.RelayStreamData: {
                this.state = SessionBackendState.AltDegraded;
                return [SessionProtocol.CmdID.DeliverRecv];
            }
            case this.state === SessionBackendState.RelayBackoff && ev === SessionProtocol.EventID.RelayStreamData: {
                this.state = SessionBackendState.RelayBackoff;
                return [SessionProtocol.CmdID.DeliverRecv];
            }
            case this.state === SessionBackendState.RelayConnected && ev === SessionProtocol.EventID.RelayStreamError: {
                this.state = SessionBackendState.RelayConnected;
                return [SessionProtocol.CmdID.DeliverRecvError];
            }
            case this.state === SessionBackendState.CandidatesAdvertised && ev === SessionProtocol.EventID.RelayStreamError: {
                this.state = SessionBackendState.CandidatesAdvertised;
                return [SessionProtocol.CmdID.DeliverRecvError];
            }
            case this.state === SessionBackendState.AltActive && ev === SessionProtocol.EventID.RelayStreamError: {
                this.state = SessionBackendState.AltActive;
                return [SessionProtocol.CmdID.DeliverRecvError];
            }
            case this.state === SessionBackendState.AltDegraded && ev === SessionProtocol.EventID.RelayStreamError: {
                this.state = SessionBackendState.AltDegraded;
                return [SessionProtocol.CmdID.DeliverRecvError];
            }
            case this.state === SessionBackendState.RelayBackoff && ev === SessionProtocol.EventID.RelayStreamError: {
                this.state = SessionBackendState.RelayBackoff;
                return [SessionProtocol.CmdID.DeliverRecvError];
            }
            case this.state === SessionBackendState.RelayConnected && ev === SessionProtocol.EventID.AppSendDatagram: {
                this.state = SessionBackendState.RelayConnected;
                return [SessionProtocol.CmdID.SendActiveDatagram];
            }
            case this.state === SessionBackendState.CandidatesAdvertised && ev === SessionProtocol.EventID.AppSendDatagram: {
                this.state = SessionBackendState.CandidatesAdvertised;
                return [SessionProtocol.CmdID.SendActiveDatagram];
            }
            case this.state === SessionBackendState.AltActive && ev === SessionProtocol.EventID.AppSendDatagram: {
                this.state = SessionBackendState.AltActive;
                return [SessionProtocol.CmdID.SendActiveDatagram];
            }
            case this.state === SessionBackendState.AltDegraded && ev === SessionProtocol.EventID.AppSendDatagram: {
                this.state = SessionBackendState.AltDegraded;
                return [SessionProtocol.CmdID.SendActiveDatagram];
            }
            case this.state === SessionBackendState.RelayBackoff && ev === SessionProtocol.EventID.AppSendDatagram: {
                this.state = SessionBackendState.RelayBackoff;
                return [SessionProtocol.CmdID.SendActiveDatagram];
            }
            case this.state === SessionBackendState.RelayConnected && ev === SessionProtocol.EventID.RelayDatagram: {
                this.state = SessionBackendState.RelayConnected;
                return [SessionProtocol.CmdID.DeliverRecvDatagram];
            }
            case this.state === SessionBackendState.CandidatesAdvertised && ev === SessionProtocol.EventID.RelayDatagram: {
                this.state = SessionBackendState.CandidatesAdvertised;
                return [SessionProtocol.CmdID.DeliverRecvDatagram];
            }
            case this.state === SessionBackendState.AltActive && ev === SessionProtocol.EventID.RelayDatagram: {
                this.state = SessionBackendState.AltActive;
                return [SessionProtocol.CmdID.DeliverRecvDatagram];
            }
            case this.state === SessionBackendState.AltDegraded && ev === SessionProtocol.EventID.RelayDatagram: {
                this.state = SessionBackendState.AltDegraded;
                return [SessionProtocol.CmdID.DeliverRecvDatagram];
            }
            case this.state === SessionBackendState.RelayBackoff && ev === SessionProtocol.EventID.RelayDatagram: {
                this.state = SessionBackendState.RelayBackoff;
                return [SessionProtocol.CmdID.DeliverRecvDatagram];
            }
            case this.state === SessionBackendState.AltActive && ev === SessionProtocol.EventID.AltStreamData: {
                this.state = SessionBackendState.AltActive;
                return [SessionProtocol.CmdID.DeliverRecv];
            }
            case this.state === SessionBackendState.AltDegraded && ev === SessionProtocol.EventID.AltStreamData: {
                this.state = SessionBackendState.AltDegraded;
                return [SessionProtocol.CmdID.DeliverRecv];
            }
            case this.state === SessionBackendState.AltActive && ev === SessionProtocol.EventID.AltDatagram: {
                this.state = SessionBackendState.AltActive;
                return [SessionProtocol.CmdID.DeliverRecvDatagram];
            }
            case this.state === SessionBackendState.AltDegraded && ev === SessionProtocol.EventID.AltDatagram: {
                this.state = SessionBackendState.AltDegraded;
                return [SessionProtocol.CmdID.DeliverRecvDatagram];
            }
        }
        return [];
    }
}

/** SessionClientMachine is the generated state machine for the client actor. */
export class SessionClientMachine {
    readonly protocol = SessionProtocol;
    state: SessionClientState;
    receivedBackendPub: string = "none"; // pubkey client received in pair_hello_ack
    clientSharedKey: string = ""; // ECDH key derived by client
    clientCode: string = ""; // code computed by client
    cActivePath: string = "relay"; // client active path
    cDispatcherPath: string = "relay"; // client datagram dispatcher binding
    altSignal: string = "pending"; // AltReady notification state
    guards: Map<SessionProtocol.GuardID, () => boolean> = new Map();
    actions: Map<SessionProtocol.ActionID, () => void> = new Map();

    constructor() {
        this.state = SessionClientState.Idle;
    }

    handleEvent(ev: SessionProtocol.EventID): SessionProtocol.CmdID[] {
        switch (true) {
            case this.state === SessionClientState.Idle && ev === SessionProtocol.EventID.BackchannelReceived: {
                this.state = SessionClientState.ObtainBackchannelSecret;
                return [];
            }
            case this.state === SessionClientState.ObtainBackchannelSecret && ev === SessionProtocol.EventID.SecretParsed: {
                this.state = SessionClientState.ConnectRelay;
                return [];
            }
            case this.state === SessionClientState.ConnectRelay && ev === SessionProtocol.EventID.RelayConnected: {
                this.state = SessionClientState.GenKeyPair;
                return [];
            }
            case this.state === SessionClientState.GenKeyPair && ev === SessionProtocol.EventID.KeyPairGenerated: {
                this.actions.get(SessionProtocol.ActionID.SendPairHello)?.();
                this.state = SessionClientState.WaitAck;
                return [];
            }
            case this.state === SessionClientState.WaitAck && ev === SessionProtocol.EventID.RecvPairHelloAck: {
                this.actions.get(SessionProtocol.ActionID.DeriveSecret)?.();
                // received_backend_pub: recv_msg.pubkey (set by action)
                // client_shared_key: DeriveKey("client_pub", recv_msg.pubkey) (set by action)
                this.state = SessionClientState.E2EReady;
                return [];
            }
            case this.state === SessionClientState.E2EReady && ev === SessionProtocol.EventID.RecvPairConfirm: {
                // client_code: DeriveCode(received_backend_pub, "client_pub") (set by action)
                this.state = SessionClientState.ShowCode;
                return [];
            }
            case this.state === SessionClientState.ShowCode && ev === SessionProtocol.EventID.CodeDisplayed: {
                this.state = SessionClientState.WaitPairComplete;
                return [];
            }
            case this.state === SessionClientState.WaitPairComplete && ev === SessionProtocol.EventID.RecvPairComplete: {
                this.actions.get(SessionProtocol.ActionID.StoreSecret)?.();
                this.state = SessionClientState.Paired;
                return [];
            }
            case this.state === SessionClientState.Paired && ev === SessionProtocol.EventID.AppLaunch: {
                this.state = SessionClientState.Reconnect;
                return [];
            }
            case this.state === SessionClientState.Reconnect && ev === SessionProtocol.EventID.RelayConnected: {
                this.state = SessionClientState.SendAuth;
                return [];
            }
            case this.state === SessionClientState.SendAuth && ev === SessionProtocol.EventID.RecvAuthOk: {
                this.state = SessionClientState.SessionActive;
                return [];
            }
            case this.state === SessionClientState.SessionActive && ev === SessionProtocol.EventID.SessionEstablished: {
                this.state = SessionClientState.RelayConnected;
                return [];
            }
            case this.state === SessionClientState.RelayConnected && ev === SessionProtocol.EventID.RecvCandidates && this.guards.get(SessionProtocol.GuardID.AltEnabled)?.() === true: {
                this.actions.get(SessionProtocol.ActionID.DialCandidate)?.();
                this.state = SessionClientState.PairDialing;
                return [SessionProtocol.CmdID.DialCandidate];
            }
            case this.state === SessionClientState.RelayConnected && ev === SessionProtocol.EventID.RecvCandidates && this.guards.get(SessionProtocol.GuardID.AltDisabled)?.() === true: {
                this.state = SessionClientState.RelayConnected;
                return [];
            }
            case this.state === SessionClientState.PairDialing && ev === SessionProtocol.EventID.DialOk: {
                this.state = SessionClientState.PairChecking;
                return [SessionProtocol.CmdID.SendPairCheck];
            }
            case this.state === SessionClientState.PairDialing && ev === SessionProtocol.EventID.DialFailed: {
                this.state = SessionClientState.RelayConnected;
                return [];
            }
            case this.state === SessionClientState.PairChecking && ev === SessionProtocol.EventID.RecvPairCheckAck: {
                this.actions.get(SessionProtocol.ActionID.ActivateLan)?.();
                this.cActivePath = "alt";
                this.cDispatcherPath = "alt";
                this.altSignal = "ready";
                this.state = SessionClientState.AltActive;
                return [SessionProtocol.CmdID.StartAltStreamReader, SessionProtocol.CmdID.StartAltDgReader, SessionProtocol.CmdID.SignalAltReady, SessionProtocol.CmdID.SetCryptoDatagram];
            }
            case this.state === SessionClientState.PairChecking && ev === SessionProtocol.EventID.VerifyTimeout: {
                this.cDispatcherPath = "relay";
                this.state = SessionClientState.RelayConnected;
                return [];
            }
            case this.state === SessionClientState.AltActive && ev === SessionProtocol.EventID.RecvPathPing: {
                this.state = SessionClientState.AltActive;
                return [SessionProtocol.CmdID.SendPathPong];
            }
            case this.state === SessionClientState.AltActive && ev === SessionProtocol.EventID.AltError: {
                this.actions.get(SessionProtocol.ActionID.FallbackToRelay)?.();
                this.cActivePath = "relay";
                this.cDispatcherPath = "relay";
                this.altSignal = "pending";
                this.state = SessionClientState.RelayFallback;
                return [SessionProtocol.CmdID.StopAltStreamReader, SessionProtocol.CmdID.StopAltDgReader, SessionProtocol.CmdID.CloseAltPath, SessionProtocol.CmdID.ResetAltReady];
            }
            case this.state === SessionClientState.AltActive && ev === SessionProtocol.EventID.AltStreamError: {
                this.actions.get(SessionProtocol.ActionID.FallbackToRelay)?.();
                this.cActivePath = "relay";
                this.cDispatcherPath = "relay";
                this.altSignal = "pending";
                this.state = SessionClientState.RelayFallback;
                return [SessionProtocol.CmdID.StopAltStreamReader, SessionProtocol.CmdID.StopAltDgReader, SessionProtocol.CmdID.CloseAltPath, SessionProtocol.CmdID.ResetAltReady];
            }
            case this.state === SessionClientState.RelayFallback && ev === SessionProtocol.EventID.RelayOk: {
                this.state = SessionClientState.RelayConnected;
                return [];
            }
            case this.state === SessionClientState.AltActive && ev === SessionProtocol.EventID.RecvCandidates && this.guards.get(SessionProtocol.GuardID.AltEnabled)?.() === true: {
                this.actions.get(SessionProtocol.ActionID.DialCandidate)?.();
                this.state = SessionClientState.PairDialing;
                return [SessionProtocol.CmdID.StopAltStreamReader, SessionProtocol.CmdID.StopAltDgReader, SessionProtocol.CmdID.CloseAltPath, SessionProtocol.CmdID.DialCandidate];
            }
            case this.state === SessionClientState.PairDialing && ev === SessionProtocol.EventID.AppForceFallback: {
                this.state = SessionClientState.RelayConnected;
                return [];
            }
            case this.state === SessionClientState.PairChecking && ev === SessionProtocol.EventID.AppForceFallback: {
                this.cDispatcherPath = "relay";
                this.state = SessionClientState.RelayConnected;
                return [SessionProtocol.CmdID.StopAltStreamReader, SessionProtocol.CmdID.StopAltDgReader, SessionProtocol.CmdID.CloseAltPath];
            }
            case this.state === SessionClientState.AltActive && ev === SessionProtocol.EventID.AppForceFallback: {
                this.actions.get(SessionProtocol.ActionID.FallbackToRelay)?.();
                this.cActivePath = "relay";
                this.cDispatcherPath = "relay";
                this.altSignal = "pending";
                this.state = SessionClientState.RelayConnected;
                return [SessionProtocol.CmdID.StopAltStreamReader, SessionProtocol.CmdID.StopAltDgReader, SessionProtocol.CmdID.CloseAltPath, SessionProtocol.CmdID.ResetAltReady];
            }
            case this.state === SessionClientState.RelayConnected && ev === SessionProtocol.EventID.Disconnect: {
                this.state = SessionClientState.Paired;
                return [];
            }
            case this.state === SessionClientState.RelayConnected && ev === SessionProtocol.EventID.AppSend: {
                this.state = SessionClientState.RelayConnected;
                return [SessionProtocol.CmdID.WriteActiveStream];
            }
            case this.state === SessionClientState.PairDialing && ev === SessionProtocol.EventID.AppSend: {
                this.state = SessionClientState.PairDialing;
                return [SessionProtocol.CmdID.WriteActiveStream];
            }
            case this.state === SessionClientState.PairChecking && ev === SessionProtocol.EventID.AppSend: {
                this.state = SessionClientState.PairChecking;
                return [SessionProtocol.CmdID.WriteActiveStream];
            }
            case this.state === SessionClientState.AltActive && ev === SessionProtocol.EventID.AppSend: {
                this.state = SessionClientState.AltActive;
                return [SessionProtocol.CmdID.WriteActiveStream];
            }
            case this.state === SessionClientState.RelayFallback && ev === SessionProtocol.EventID.AppSend: {
                this.state = SessionClientState.RelayFallback;
                return [SessionProtocol.CmdID.WriteActiveStream];
            }
            case this.state === SessionClientState.RelayConnected && ev === SessionProtocol.EventID.RelayStreamData: {
                this.state = SessionClientState.RelayConnected;
                return [SessionProtocol.CmdID.DeliverRecv];
            }
            case this.state === SessionClientState.PairDialing && ev === SessionProtocol.EventID.RelayStreamData: {
                this.state = SessionClientState.PairDialing;
                return [SessionProtocol.CmdID.DeliverRecv];
            }
            case this.state === SessionClientState.PairChecking && ev === SessionProtocol.EventID.RelayStreamData: {
                this.state = SessionClientState.PairChecking;
                return [SessionProtocol.CmdID.DeliverRecv];
            }
            case this.state === SessionClientState.AltActive && ev === SessionProtocol.EventID.RelayStreamData: {
                this.state = SessionClientState.AltActive;
                return [SessionProtocol.CmdID.DeliverRecv];
            }
            case this.state === SessionClientState.RelayFallback && ev === SessionProtocol.EventID.RelayStreamData: {
                this.state = SessionClientState.RelayFallback;
                return [SessionProtocol.CmdID.DeliverRecv];
            }
            case this.state === SessionClientState.RelayConnected && ev === SessionProtocol.EventID.RelayStreamError: {
                this.state = SessionClientState.RelayConnected;
                return [SessionProtocol.CmdID.DeliverRecvError];
            }
            case this.state === SessionClientState.PairDialing && ev === SessionProtocol.EventID.RelayStreamError: {
                this.state = SessionClientState.PairDialing;
                return [SessionProtocol.CmdID.DeliverRecvError];
            }
            case this.state === SessionClientState.PairChecking && ev === SessionProtocol.EventID.RelayStreamError: {
                this.state = SessionClientState.PairChecking;
                return [SessionProtocol.CmdID.DeliverRecvError];
            }
            case this.state === SessionClientState.AltActive && ev === SessionProtocol.EventID.RelayStreamError: {
                this.state = SessionClientState.AltActive;
                return [SessionProtocol.CmdID.DeliverRecvError];
            }
            case this.state === SessionClientState.RelayFallback && ev === SessionProtocol.EventID.RelayStreamError: {
                this.state = SessionClientState.RelayFallback;
                return [SessionProtocol.CmdID.DeliverRecvError];
            }
            case this.state === SessionClientState.RelayConnected && ev === SessionProtocol.EventID.AppSendDatagram: {
                this.state = SessionClientState.RelayConnected;
                return [SessionProtocol.CmdID.SendActiveDatagram];
            }
            case this.state === SessionClientState.PairDialing && ev === SessionProtocol.EventID.AppSendDatagram: {
                this.state = SessionClientState.PairDialing;
                return [SessionProtocol.CmdID.SendActiveDatagram];
            }
            case this.state === SessionClientState.PairChecking && ev === SessionProtocol.EventID.AppSendDatagram: {
                this.state = SessionClientState.PairChecking;
                return [SessionProtocol.CmdID.SendActiveDatagram];
            }
            case this.state === SessionClientState.AltActive && ev === SessionProtocol.EventID.AppSendDatagram: {
                this.state = SessionClientState.AltActive;
                return [SessionProtocol.CmdID.SendActiveDatagram];
            }
            case this.state === SessionClientState.RelayFallback && ev === SessionProtocol.EventID.AppSendDatagram: {
                this.state = SessionClientState.RelayFallback;
                return [SessionProtocol.CmdID.SendActiveDatagram];
            }
            case this.state === SessionClientState.RelayConnected && ev === SessionProtocol.EventID.RelayDatagram: {
                this.state = SessionClientState.RelayConnected;
                return [SessionProtocol.CmdID.DeliverRecvDatagram];
            }
            case this.state === SessionClientState.PairDialing && ev === SessionProtocol.EventID.RelayDatagram: {
                this.state = SessionClientState.PairDialing;
                return [SessionProtocol.CmdID.DeliverRecvDatagram];
            }
            case this.state === SessionClientState.PairChecking && ev === SessionProtocol.EventID.RelayDatagram: {
                this.state = SessionClientState.PairChecking;
                return [SessionProtocol.CmdID.DeliverRecvDatagram];
            }
            case this.state === SessionClientState.AltActive && ev === SessionProtocol.EventID.RelayDatagram: {
                this.state = SessionClientState.AltActive;
                return [SessionProtocol.CmdID.DeliverRecvDatagram];
            }
            case this.state === SessionClientState.RelayFallback && ev === SessionProtocol.EventID.RelayDatagram: {
                this.state = SessionClientState.RelayFallback;
                return [SessionProtocol.CmdID.DeliverRecvDatagram];
            }
            case this.state === SessionClientState.AltActive && ev === SessionProtocol.EventID.AltStreamData: {
                this.state = SessionClientState.AltActive;
                return [SessionProtocol.CmdID.DeliverRecv];
            }
            case this.state === SessionClientState.AltActive && ev === SessionProtocol.EventID.AltDatagram: {
                this.state = SessionClientState.AltActive;
                return [SessionProtocol.CmdID.DeliverRecvDatagram];
            }
        }
        return [];
    }
}

/** SessionRelayMachine is the generated state machine for the relay actor. */
export class SessionRelayMachine {
    readonly protocol = SessionProtocol;
    state: SessionRelayState;
    relayBridge: string = "idle"; // relay bridge state
    guards: Map<SessionProtocol.GuardID, () => boolean> = new Map();
    actions: Map<SessionProtocol.ActionID, () => void> = new Map();

    constructor() {
        this.state = SessionRelayState.Idle;
    }

    handleEvent(ev: SessionProtocol.EventID): SessionProtocol.CmdID[] {
        switch (true) {
            case this.state === SessionRelayState.Idle && ev === SessionProtocol.EventID.BackendRegister: {
                this.state = SessionRelayState.BackendRegistered;
                return [];
            }
            case this.state === SessionRelayState.BackendRegistered && ev === SessionProtocol.EventID.ClientConnect: {
                this.actions.get(SessionProtocol.ActionID.BridgeStreams)?.();
                this.relayBridge = "active";
                this.state = SessionRelayState.Bridged;
                return [];
            }
            case this.state === SessionRelayState.Bridged && ev === SessionProtocol.EventID.ClientDisconnect: {
                this.actions.get(SessionProtocol.ActionID.Unbridge)?.();
                this.relayBridge = "idle";
                this.state = SessionRelayState.BackendRegistered;
                return [];
            }
            case this.state === SessionRelayState.BackendRegistered && ev === SessionProtocol.EventID.BackendDisconnect: {
                this.state = SessionRelayState.Idle;
                return [];
            }
        }
        return [];
    }
}

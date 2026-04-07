// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Auto-generated from protocol definition. Do not edit.
// Source of truth: protocol/*.yaml

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

/** The protocol transition table and shared type enums. */
export namespace SessionProtocol {

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
        FallbackToRelay = "fallback_to_relay",
        ResetFailures = "reset_failures",
        SendPairHello = "send_pair_hello",
        StoreSecret = "store_secret",
        DialLan = "dial_lan",
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
        LanStreamData = "lan_stream_data",
        LanStreamError = "lan_stream_error",
        LanDatagram = "lan_datagram",
        LanDialOk = "lan_dial_ok",
        LanDialFailed = "lan_dial_failed",
        LanVerifyOk = "lan_verify_ok",
        PingTimeout = "ping_timeout",
        PingTick = "ping_tick",
        BackoffExpired = "backoff_expired",
        OfferTimeout = "offer_timeout",
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
        LanServerReady = "lan_server_ready",
        LanServerChanged = "lan_server_changed",
        ReadvertiseTick = "readvertise_tick",
        Disconnect = "disconnect",
        BackchannelReceived = "backchannel_received",
        SecretParsed = "secret_parsed",
        RelayConnected = "relay_connected",
        KeyPairGenerated = "key_pair_generated",
        CodeDisplayed = "code_displayed",
        AppLaunch = "app_launch",
        VerifyTimeout = "verify_timeout",
        LanError = "lan_error",
        RelayOk = "relay_ok",
        BackendRegister = "backend_register",
        ClientConnect = "client_connect",
        ClientDisconnect = "client_disconnect",
        BackendDisconnect = "backend_disconnect",
        RecvPairHello = "recv_pair_hello",
        RecvAuthRequest = "recv_auth_request",
        RecvLanVerify = "recv_lan_verify",
        RecvPathPong = "recv_path_pong",
        RecvPairHelloAck = "recv_pair_hello_ack",
        RecvPairConfirm = "recv_pair_confirm",
        RecvPairComplete = "recv_pair_complete",
        RecvAuthOk = "recv_auth_ok",
        RecvLanOffer = "recv_lan_offer",
        RecvLanConfirm = "recv_lan_confirm",
        RecvPathPing = "recv_path_ping",
    }

    export enum CmdID {
        WriteActiveStream = "write_active_stream",
        SendActiveDatagram = "send_active_datagram",
        SendPathPing = "send_path_ping",
        SendPathPong = "send_path_pong",
        SendLanOffer = "send_lan_offer",
        SendLanVerify = "send_lan_verify",
        SendLanConfirm = "send_lan_confirm",
        DialLan = "dial_lan",
        DeliverRecv = "deliver_recv",
        DeliverRecvError = "deliver_recv_error",
        DeliverRecvDatagram = "deliver_recv_datagram",
        StartLanStreamReader = "start_lan_stream_reader",
        StopLanStreamReader = "stop_lan_stream_reader",
        StartLanDgReader = "start_lan_dg_reader",
        StopLanDgReader = "stop_lan_dg_reader",
        StartMonitor = "start_monitor",
        StopMonitor = "stop_monitor",
        StartPongTimeout = "start_pong_timeout",
        CancelPongTimeout = "cancel_pong_timeout",
        StartBackoffTimer = "start_backoff_timer",
        CloseLanPath = "close_lan_path",
        SignalLanReady = "signal_lan_ready",
        ResetLanReady = "reset_lan_ready",
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
        FRAME_LAN_OFFER: 0x01,
        FRAME_CUTOVER: 0x02,
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
            { from: "LANActive", to: "RelayBackoff", on: "lan_stream_error", onKind: "internal", action: "fallback_to_relay" },
            { from: "LANDegraded", to: "RelayBackoff", on: "lan_stream_error", onKind: "internal", action: "fallback_to_relay" },
            { from: "LANDegraded", to: "LANActive", on: "path_pong", onKind: "recv", action: "reset_failures" },
            { from: "LANDegraded", to: "LANDegraded", on: "ping_timeout", onKind: "internal", guard: "under_max_failures" },
            { from: "LANDegraded", to: "RelayBackoff", on: "ping_timeout", onKind: "internal", guard: "at_max_failures", action: "fallback_to_relay" },
            { from: "RelayBackoff", to: "LANOffered", on: "backoff_expired", onKind: "internal", sends: [{ to: "client", msg: "lan_offer" }] },
            { from: "RelayBackoff", to: "LANOffered", on: "lan_server_changed", onKind: "internal", sends: [{ to: "client", msg: "lan_offer" }] },
            { from: "RelayConnected", to: "LANOffered", on: "readvertise_tick", onKind: "internal", guard: "lan_server_available", sends: [{ to: "client", msg: "lan_offer" }] },
            { from: "LANOffered", to: "RelayConnected", on: "app_force_fallback", onKind: "internal" },
            { from: "LANActive", to: "RelayBackoff", on: "app_force_fallback", onKind: "internal", action: "fallback_to_relay" },
            { from: "LANDegraded", to: "RelayBackoff", on: "app_force_fallback", onKind: "internal", action: "fallback_to_relay" },
            { from: "RelayConnected", to: "Paired", on: "disconnect", onKind: "internal" },
            { from: "RelayConnected", to: "RelayConnected", on: "app_send", onKind: "internal" },
            { from: "LANOffered", to: "LANOffered", on: "app_send", onKind: "internal" },
            { from: "LANActive", to: "LANActive", on: "app_send", onKind: "internal" },
            { from: "LANDegraded", to: "LANDegraded", on: "app_send", onKind: "internal" },
            { from: "RelayBackoff", to: "RelayBackoff", on: "app_send", onKind: "internal" },
            { from: "RelayConnected", to: "RelayConnected", on: "relay_stream_data", onKind: "internal" },
            { from: "LANOffered", to: "LANOffered", on: "relay_stream_data", onKind: "internal" },
            { from: "LANActive", to: "LANActive", on: "relay_stream_data", onKind: "internal" },
            { from: "LANDegraded", to: "LANDegraded", on: "relay_stream_data", onKind: "internal" },
            { from: "RelayBackoff", to: "RelayBackoff", on: "relay_stream_data", onKind: "internal" },
            { from: "RelayConnected", to: "RelayConnected", on: "relay_stream_error", onKind: "internal" },
            { from: "LANOffered", to: "LANOffered", on: "relay_stream_error", onKind: "internal" },
            { from: "LANActive", to: "LANActive", on: "relay_stream_error", onKind: "internal" },
            { from: "LANDegraded", to: "LANDegraded", on: "relay_stream_error", onKind: "internal" },
            { from: "RelayBackoff", to: "RelayBackoff", on: "relay_stream_error", onKind: "internal" },
            { from: "RelayConnected", to: "RelayConnected", on: "app_send_datagram", onKind: "internal" },
            { from: "LANOffered", to: "LANOffered", on: "app_send_datagram", onKind: "internal" },
            { from: "LANActive", to: "LANActive", on: "app_send_datagram", onKind: "internal" },
            { from: "LANDegraded", to: "LANDegraded", on: "app_send_datagram", onKind: "internal" },
            { from: "RelayBackoff", to: "RelayBackoff", on: "app_send_datagram", onKind: "internal" },
            { from: "RelayConnected", to: "RelayConnected", on: "relay_datagram", onKind: "internal" },
            { from: "LANOffered", to: "LANOffered", on: "relay_datagram", onKind: "internal" },
            { from: "LANActive", to: "LANActive", on: "relay_datagram", onKind: "internal" },
            { from: "LANDegraded", to: "LANDegraded", on: "relay_datagram", onKind: "internal" },
            { from: "RelayBackoff", to: "RelayBackoff", on: "relay_datagram", onKind: "internal" },
            { from: "LANActive", to: "LANActive", on: "lan_stream_data", onKind: "internal" },
            { from: "LANDegraded", to: "LANDegraded", on: "lan_stream_data", onKind: "internal" },
            { from: "LANActive", to: "LANActive", on: "lan_datagram", onKind: "internal" },
            { from: "LANDegraded", to: "LANDegraded", on: "lan_datagram", onKind: "internal" },
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
            { from: "LANActive", to: "RelayFallback", on: "lan_stream_error", onKind: "internal", action: "fallback_to_relay" },
            { from: "RelayFallback", to: "RelayConnected", on: "relay_ok", onKind: "internal" },
            { from: "LANActive", to: "LANConnecting", on: "lan_offer", onKind: "recv", guard: "lan_enabled", action: "dial_lan" },
            { from: "LANConnecting", to: "RelayConnected", on: "app_force_fallback", onKind: "internal" },
            { from: "LANVerifying", to: "RelayConnected", on: "app_force_fallback", onKind: "internal" },
            { from: "LANActive", to: "RelayConnected", on: "app_force_fallback", onKind: "internal", action: "fallback_to_relay" },
            { from: "RelayConnected", to: "Paired", on: "disconnect", onKind: "internal" },
            { from: "RelayConnected", to: "RelayConnected", on: "app_send", onKind: "internal" },
            { from: "LANConnecting", to: "LANConnecting", on: "app_send", onKind: "internal" },
            { from: "LANVerifying", to: "LANVerifying", on: "app_send", onKind: "internal" },
            { from: "LANActive", to: "LANActive", on: "app_send", onKind: "internal" },
            { from: "RelayFallback", to: "RelayFallback", on: "app_send", onKind: "internal" },
            { from: "RelayConnected", to: "RelayConnected", on: "relay_stream_data", onKind: "internal" },
            { from: "LANConnecting", to: "LANConnecting", on: "relay_stream_data", onKind: "internal" },
            { from: "LANVerifying", to: "LANVerifying", on: "relay_stream_data", onKind: "internal" },
            { from: "LANActive", to: "LANActive", on: "relay_stream_data", onKind: "internal" },
            { from: "RelayFallback", to: "RelayFallback", on: "relay_stream_data", onKind: "internal" },
            { from: "RelayConnected", to: "RelayConnected", on: "relay_stream_error", onKind: "internal" },
            { from: "LANConnecting", to: "LANConnecting", on: "relay_stream_error", onKind: "internal" },
            { from: "LANVerifying", to: "LANVerifying", on: "relay_stream_error", onKind: "internal" },
            { from: "LANActive", to: "LANActive", on: "relay_stream_error", onKind: "internal" },
            { from: "RelayFallback", to: "RelayFallback", on: "relay_stream_error", onKind: "internal" },
            { from: "RelayConnected", to: "RelayConnected", on: "app_send_datagram", onKind: "internal" },
            { from: "LANConnecting", to: "LANConnecting", on: "app_send_datagram", onKind: "internal" },
            { from: "LANVerifying", to: "LANVerifying", on: "app_send_datagram", onKind: "internal" },
            { from: "LANActive", to: "LANActive", on: "app_send_datagram", onKind: "internal" },
            { from: "RelayFallback", to: "RelayFallback", on: "app_send_datagram", onKind: "internal" },
            { from: "RelayConnected", to: "RelayConnected", on: "relay_datagram", onKind: "internal" },
            { from: "LANConnecting", to: "LANConnecting", on: "relay_datagram", onKind: "internal" },
            { from: "LANVerifying", to: "LANVerifying", on: "relay_datagram", onKind: "internal" },
            { from: "LANActive", to: "LANActive", on: "relay_datagram", onKind: "internal" },
            { from: "RelayFallback", to: "RelayFallback", on: "relay_datagram", onKind: "internal" },
            { from: "LANActive", to: "LANActive", on: "lan_stream_data", onKind: "internal" },
            { from: "LANActive", to: "LANActive", on: "lan_datagram", onKind: "internal" },
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

}

/** BackendMachine is the generated state machine for the backend actor. */
export class BackendMachine {
    readonly protocol = SessionProtocol;
    state: BackendState;
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
    lanSignal: string = "pending"; // LANReady notification state
    guards: Map<SessionProtocol.GuardID, () => boolean> = new Map();
    actions: Map<SessionProtocol.ActionID, () => void> = new Map();

    constructor() {
        this.state = BackendState.Idle;
    }

    handleEvent(ev: SessionProtocol.EventID): SessionProtocol.CmdID[] {
        switch (true) {
            case this.state === BackendState.Idle && ev === SessionProtocol.EventID.CliInitPair: {
                this.actions.get(SessionProtocol.ActionID.GenerateToken)?.();
                this.currentToken = "tok_1";
                // active_tokens: active_tokens \union {"tok_1"} (set by action)
                this.state = BackendState.GenerateToken;
                return [];
            }
            case this.state === BackendState.GenerateToken && ev === SessionProtocol.EventID.TokenCreated: {
                this.actions.get(SessionProtocol.ActionID.RegisterRelay)?.();
                this.state = BackendState.RegisterRelay;
                return [];
            }
            case this.state === BackendState.RegisterRelay && ev === SessionProtocol.EventID.RelayRegistered: {
                this.secretPublished = true;
                this.state = BackendState.WaitingForClient;
                return [];
            }
            case this.state === BackendState.WaitingForClient && ev === SessionProtocol.EventID.RecvPairHello && this.guards.get(SessionProtocol.GuardID.TokenValid)?.() === true: {
                this.actions.get(SessionProtocol.ActionID.DeriveSecret)?.();
                // received_client_pub: recv_msg.pubkey (set by action)
                this.backendEcdhPub = "backend_pub";
                // backend_shared_key: DeriveKey("backend_pub", recv_msg.pubkey) (set by action)
                // backend_code: DeriveCode("backend_pub", recv_msg.pubkey) (set by action)
                this.state = BackendState.DeriveSecret;
                return [];
            }
            case this.state === BackendState.WaitingForClient && ev === SessionProtocol.EventID.RecvPairHello && this.guards.get(SessionProtocol.GuardID.TokenInvalid)?.() === true: {
                this.state = BackendState.Idle;
                return [];
            }
            case this.state === BackendState.DeriveSecret && ev === SessionProtocol.EventID.EcdhComplete: {
                this.state = BackendState.SendAck;
                return [];
            }
            case this.state === BackendState.SendAck && ev === SessionProtocol.EventID.SignalCodeDisplay: {
                this.state = BackendState.WaitingForCode;
                return [];
            }
            case this.state === BackendState.WaitingForCode && ev === SessionProtocol.EventID.CliCodeEntered: {
                // received_code: cli_entered_code (set by action)
                this.state = BackendState.ValidateCode;
                return [];
            }
            case this.state === BackendState.ValidateCode && ev === SessionProtocol.EventID.CheckCode && this.guards.get(SessionProtocol.GuardID.CodeCorrect)?.() === true: {
                this.state = BackendState.StorePaired;
                return [];
            }
            case this.state === BackendState.ValidateCode && ev === SessionProtocol.EventID.CheckCode && this.guards.get(SessionProtocol.GuardID.CodeWrong)?.() === true: {
                // code_attempts: code_attempts + 1 (set by action)
                this.state = BackendState.Idle;
                return [];
            }
            case this.state === BackendState.StorePaired && ev === SessionProtocol.EventID.Finalise: {
                this.actions.get(SessionProtocol.ActionID.StoreDevice)?.();
                this.deviceSecret = "dev_secret_1";
                // paired_devices: paired_devices \union {"device_1"} (set by action)
                // active_tokens: active_tokens \ {current_token} (set by action)
                // used_tokens: used_tokens \union {current_token} (set by action)
                this.state = BackendState.Paired;
                return [];
            }
            case this.state === BackendState.Paired && ev === SessionProtocol.EventID.RecvAuthRequest: {
                // received_device_id: recv_msg.device_id (set by action)
                // received_auth_nonce: recv_msg.nonce (set by action)
                this.state = BackendState.AuthCheck;
                return [];
            }
            case this.state === BackendState.AuthCheck && ev === SessionProtocol.EventID.Verify && this.guards.get(SessionProtocol.GuardID.DeviceKnown)?.() === true: {
                this.actions.get(SessionProtocol.ActionID.VerifyDevice)?.();
                // auth_nonces_used: auth_nonces_used \union {received_auth_nonce} (set by action)
                this.state = BackendState.SessionActive;
                return [];
            }
            case this.state === BackendState.AuthCheck && ev === SessionProtocol.EventID.Verify && this.guards.get(SessionProtocol.GuardID.DeviceUnknown)?.() === true: {
                this.state = BackendState.Idle;
                return [];
            }
            case this.state === BackendState.SessionActive && ev === SessionProtocol.EventID.SessionEstablished: {
                this.state = BackendState.RelayConnected;
                return [];
            }
            case this.state === BackendState.RelayConnected && ev === SessionProtocol.EventID.LanServerReady: {
                this.state = BackendState.LANOffered;
                return [SessionProtocol.CmdID.SendLanOffer];
            }
            case this.state === BackendState.LANOffered && ev === SessionProtocol.EventID.RecvLanVerify && this.guards.get(SessionProtocol.GuardID.ChallengeValid)?.() === true: {
                this.actions.get(SessionProtocol.ActionID.ActivateLan)?.();
                this.pingFailures = 0;
                this.backoffLevel = 0;
                this.bActivePath = "lan";
                this.bDispatcherPath = "lan";
                this.monitorTarget = "lan";
                this.lanSignal = "ready";
                this.state = BackendState.LANActive;
                return [SessionProtocol.CmdID.SendLanConfirm, SessionProtocol.CmdID.StartLanStreamReader, SessionProtocol.CmdID.StartLanDgReader, SessionProtocol.CmdID.StartMonitor, SessionProtocol.CmdID.SignalLanReady, SessionProtocol.CmdID.SetCryptoDatagram];
            }
            case this.state === BackendState.LANOffered && ev === SessionProtocol.EventID.RecvLanVerify && this.guards.get(SessionProtocol.GuardID.ChallengeInvalid)?.() === true: {
                this.state = BackendState.RelayConnected;
                return [];
            }
            case this.state === BackendState.LANOffered && ev === SessionProtocol.EventID.OfferTimeout: {
                // backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
                this.lanSignal = "pending";
                this.state = BackendState.RelayBackoff;
                return [SessionProtocol.CmdID.ResetLanReady, SessionProtocol.CmdID.StartBackoffTimer];
            }
            case this.state === BackendState.LANActive && ev === SessionProtocol.EventID.PingTick: {
                this.state = BackendState.LANActive;
                return [SessionProtocol.CmdID.SendPathPing, SessionProtocol.CmdID.StartPongTimeout];
            }
            case this.state === BackendState.LANActive && ev === SessionProtocol.EventID.PingTimeout: {
                this.pingFailures = 1;
                this.state = BackendState.LANDegraded;
                return [];
            }
            case this.state === BackendState.LANDegraded && ev === SessionProtocol.EventID.PingTick: {
                this.state = BackendState.LANDegraded;
                return [SessionProtocol.CmdID.SendPathPing, SessionProtocol.CmdID.StartPongTimeout];
            }
            case this.state === BackendState.LANActive && ev === SessionProtocol.EventID.LanStreamError: {
                this.actions.get(SessionProtocol.ActionID.FallbackToRelay)?.();
                // backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
                this.bActivePath = "relay";
                this.bDispatcherPath = "relay";
                this.monitorTarget = "none";
                this.lanSignal = "pending";
                this.pingFailures = 0;
                this.state = BackendState.RelayBackoff;
                return [SessionProtocol.CmdID.StopMonitor, SessionProtocol.CmdID.StopLanStreamReader, SessionProtocol.CmdID.StopLanDgReader, SessionProtocol.CmdID.CloseLanPath, SessionProtocol.CmdID.ResetLanReady, SessionProtocol.CmdID.StartBackoffTimer];
            }
            case this.state === BackendState.LANDegraded && ev === SessionProtocol.EventID.LanStreamError: {
                this.actions.get(SessionProtocol.ActionID.FallbackToRelay)?.();
                // backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
                this.bActivePath = "relay";
                this.bDispatcherPath = "relay";
                this.monitorTarget = "none";
                this.lanSignal = "pending";
                this.pingFailures = 0;
                this.state = BackendState.RelayBackoff;
                return [SessionProtocol.CmdID.StopMonitor, SessionProtocol.CmdID.StopLanStreamReader, SessionProtocol.CmdID.StopLanDgReader, SessionProtocol.CmdID.CloseLanPath, SessionProtocol.CmdID.ResetLanReady, SessionProtocol.CmdID.StartBackoffTimer];
            }
            case this.state === BackendState.LANDegraded && ev === SessionProtocol.EventID.RecvPathPong: {
                this.actions.get(SessionProtocol.ActionID.ResetFailures)?.();
                this.pingFailures = 0;
                this.state = BackendState.LANActive;
                return [SessionProtocol.CmdID.CancelPongTimeout];
            }
            case this.state === BackendState.LANDegraded && ev === SessionProtocol.EventID.PingTimeout && this.guards.get(SessionProtocol.GuardID.UnderMaxFailures)?.() === true: {
                // ping_failures: ping_failures + 1 (set by action)
                this.state = BackendState.LANDegraded;
                return [];
            }
            case this.state === BackendState.LANDegraded && ev === SessionProtocol.EventID.PingTimeout && this.guards.get(SessionProtocol.GuardID.AtMaxFailures)?.() === true: {
                this.actions.get(SessionProtocol.ActionID.FallbackToRelay)?.();
                // backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
                this.bActivePath = "relay";
                this.bDispatcherPath = "relay";
                this.monitorTarget = "none";
                this.lanSignal = "pending";
                this.pingFailures = 0;
                this.state = BackendState.RelayBackoff;
                return [SessionProtocol.CmdID.StopMonitor, SessionProtocol.CmdID.StopLanStreamReader, SessionProtocol.CmdID.StopLanDgReader, SessionProtocol.CmdID.CloseLanPath, SessionProtocol.CmdID.ResetLanReady, SessionProtocol.CmdID.StartBackoffTimer];
            }
            case this.state === BackendState.RelayBackoff && ev === SessionProtocol.EventID.BackoffExpired: {
                this.state = BackendState.LANOffered;
                return [SessionProtocol.CmdID.SendLanOffer];
            }
            case this.state === BackendState.RelayBackoff && ev === SessionProtocol.EventID.LanServerChanged: {
                this.backoffLevel = 0;
                this.state = BackendState.LANOffered;
                return [SessionProtocol.CmdID.SendLanOffer];
            }
            case this.state === BackendState.RelayConnected && ev === SessionProtocol.EventID.ReadvertiseTick && this.guards.get(SessionProtocol.GuardID.LanServerAvailable)?.() === true: {
                this.state = BackendState.LANOffered;
                return [SessionProtocol.CmdID.SendLanOffer];
            }
            case this.state === BackendState.LANOffered && ev === SessionProtocol.EventID.AppForceFallback: {
                this.lanSignal = "pending";
                this.state = BackendState.RelayConnected;
                return [SessionProtocol.CmdID.ResetLanReady];
            }
            case this.state === BackendState.LANActive && ev === SessionProtocol.EventID.AppForceFallback: {
                this.actions.get(SessionProtocol.ActionID.FallbackToRelay)?.();
                // backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
                this.bActivePath = "relay";
                this.bDispatcherPath = "relay";
                this.monitorTarget = "none";
                this.lanSignal = "pending";
                this.pingFailures = 0;
                this.state = BackendState.RelayBackoff;
                return [SessionProtocol.CmdID.StopMonitor, SessionProtocol.CmdID.CancelPongTimeout, SessionProtocol.CmdID.StopLanStreamReader, SessionProtocol.CmdID.StopLanDgReader, SessionProtocol.CmdID.CloseLanPath, SessionProtocol.CmdID.ResetLanReady, SessionProtocol.CmdID.StartBackoffTimer];
            }
            case this.state === BackendState.LANDegraded && ev === SessionProtocol.EventID.AppForceFallback: {
                this.actions.get(SessionProtocol.ActionID.FallbackToRelay)?.();
                // backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
                this.bActivePath = "relay";
                this.bDispatcherPath = "relay";
                this.monitorTarget = "none";
                this.lanSignal = "pending";
                this.pingFailures = 0;
                this.state = BackendState.RelayBackoff;
                return [SessionProtocol.CmdID.StopMonitor, SessionProtocol.CmdID.CancelPongTimeout, SessionProtocol.CmdID.StopLanStreamReader, SessionProtocol.CmdID.StopLanDgReader, SessionProtocol.CmdID.CloseLanPath, SessionProtocol.CmdID.ResetLanReady, SessionProtocol.CmdID.StartBackoffTimer];
            }
            case this.state === BackendState.RelayConnected && ev === SessionProtocol.EventID.Disconnect: {
                this.state = BackendState.Paired;
                return [];
            }
            case this.state === BackendState.RelayConnected && ev === SessionProtocol.EventID.AppSend: {
                this.state = BackendState.RelayConnected;
                return [SessionProtocol.CmdID.WriteActiveStream];
            }
            case this.state === BackendState.LANOffered && ev === SessionProtocol.EventID.AppSend: {
                this.state = BackendState.LANOffered;
                return [SessionProtocol.CmdID.WriteActiveStream];
            }
            case this.state === BackendState.LANActive && ev === SessionProtocol.EventID.AppSend: {
                this.state = BackendState.LANActive;
                return [SessionProtocol.CmdID.WriteActiveStream];
            }
            case this.state === BackendState.LANDegraded && ev === SessionProtocol.EventID.AppSend: {
                this.state = BackendState.LANDegraded;
                return [SessionProtocol.CmdID.WriteActiveStream];
            }
            case this.state === BackendState.RelayBackoff && ev === SessionProtocol.EventID.AppSend: {
                this.state = BackendState.RelayBackoff;
                return [SessionProtocol.CmdID.WriteActiveStream];
            }
            case this.state === BackendState.RelayConnected && ev === SessionProtocol.EventID.RelayStreamData: {
                this.state = BackendState.RelayConnected;
                return [SessionProtocol.CmdID.DeliverRecv];
            }
            case this.state === BackendState.LANOffered && ev === SessionProtocol.EventID.RelayStreamData: {
                this.state = BackendState.LANOffered;
                return [SessionProtocol.CmdID.DeliverRecv];
            }
            case this.state === BackendState.LANActive && ev === SessionProtocol.EventID.RelayStreamData: {
                this.state = BackendState.LANActive;
                return [SessionProtocol.CmdID.DeliverRecv];
            }
            case this.state === BackendState.LANDegraded && ev === SessionProtocol.EventID.RelayStreamData: {
                this.state = BackendState.LANDegraded;
                return [SessionProtocol.CmdID.DeliverRecv];
            }
            case this.state === BackendState.RelayBackoff && ev === SessionProtocol.EventID.RelayStreamData: {
                this.state = BackendState.RelayBackoff;
                return [SessionProtocol.CmdID.DeliverRecv];
            }
            case this.state === BackendState.RelayConnected && ev === SessionProtocol.EventID.RelayStreamError: {
                this.state = BackendState.RelayConnected;
                return [SessionProtocol.CmdID.DeliverRecvError];
            }
            case this.state === BackendState.LANOffered && ev === SessionProtocol.EventID.RelayStreamError: {
                this.state = BackendState.LANOffered;
                return [SessionProtocol.CmdID.DeliverRecvError];
            }
            case this.state === BackendState.LANActive && ev === SessionProtocol.EventID.RelayStreamError: {
                this.state = BackendState.LANActive;
                return [SessionProtocol.CmdID.DeliverRecvError];
            }
            case this.state === BackendState.LANDegraded && ev === SessionProtocol.EventID.RelayStreamError: {
                this.state = BackendState.LANDegraded;
                return [SessionProtocol.CmdID.DeliverRecvError];
            }
            case this.state === BackendState.RelayBackoff && ev === SessionProtocol.EventID.RelayStreamError: {
                this.state = BackendState.RelayBackoff;
                return [SessionProtocol.CmdID.DeliverRecvError];
            }
            case this.state === BackendState.RelayConnected && ev === SessionProtocol.EventID.AppSendDatagram: {
                this.state = BackendState.RelayConnected;
                return [SessionProtocol.CmdID.SendActiveDatagram];
            }
            case this.state === BackendState.LANOffered && ev === SessionProtocol.EventID.AppSendDatagram: {
                this.state = BackendState.LANOffered;
                return [SessionProtocol.CmdID.SendActiveDatagram];
            }
            case this.state === BackendState.LANActive && ev === SessionProtocol.EventID.AppSendDatagram: {
                this.state = BackendState.LANActive;
                return [SessionProtocol.CmdID.SendActiveDatagram];
            }
            case this.state === BackendState.LANDegraded && ev === SessionProtocol.EventID.AppSendDatagram: {
                this.state = BackendState.LANDegraded;
                return [SessionProtocol.CmdID.SendActiveDatagram];
            }
            case this.state === BackendState.RelayBackoff && ev === SessionProtocol.EventID.AppSendDatagram: {
                this.state = BackendState.RelayBackoff;
                return [SessionProtocol.CmdID.SendActiveDatagram];
            }
            case this.state === BackendState.RelayConnected && ev === SessionProtocol.EventID.RelayDatagram: {
                this.state = BackendState.RelayConnected;
                return [SessionProtocol.CmdID.DeliverRecvDatagram];
            }
            case this.state === BackendState.LANOffered && ev === SessionProtocol.EventID.RelayDatagram: {
                this.state = BackendState.LANOffered;
                return [SessionProtocol.CmdID.DeliverRecvDatagram];
            }
            case this.state === BackendState.LANActive && ev === SessionProtocol.EventID.RelayDatagram: {
                this.state = BackendState.LANActive;
                return [SessionProtocol.CmdID.DeliverRecvDatagram];
            }
            case this.state === BackendState.LANDegraded && ev === SessionProtocol.EventID.RelayDatagram: {
                this.state = BackendState.LANDegraded;
                return [SessionProtocol.CmdID.DeliverRecvDatagram];
            }
            case this.state === BackendState.RelayBackoff && ev === SessionProtocol.EventID.RelayDatagram: {
                this.state = BackendState.RelayBackoff;
                return [SessionProtocol.CmdID.DeliverRecvDatagram];
            }
            case this.state === BackendState.LANActive && ev === SessionProtocol.EventID.LanStreamData: {
                this.state = BackendState.LANActive;
                return [SessionProtocol.CmdID.DeliverRecv];
            }
            case this.state === BackendState.LANDegraded && ev === SessionProtocol.EventID.LanStreamData: {
                this.state = BackendState.LANDegraded;
                return [SessionProtocol.CmdID.DeliverRecv];
            }
            case this.state === BackendState.LANActive && ev === SessionProtocol.EventID.LanDatagram: {
                this.state = BackendState.LANActive;
                return [SessionProtocol.CmdID.DeliverRecvDatagram];
            }
            case this.state === BackendState.LANDegraded && ev === SessionProtocol.EventID.LanDatagram: {
                this.state = BackendState.LANDegraded;
                return [SessionProtocol.CmdID.DeliverRecvDatagram];
            }
        }
        return [];
    }
}

/** ClientMachine is the generated state machine for the client actor. */
export class ClientMachine {
    readonly protocol = SessionProtocol;
    state: ClientState;
    receivedBackendPub: string = "none"; // pubkey client received in pair_hello_ack
    clientSharedKey: string = ""; // ECDH key derived by client
    clientCode: string = ""; // code computed by client
    cActivePath: string = "relay"; // client active path
    cDispatcherPath: string = "relay"; // client datagram dispatcher binding
    lanSignal: string = "pending"; // LANReady notification state
    guards: Map<SessionProtocol.GuardID, () => boolean> = new Map();
    actions: Map<SessionProtocol.ActionID, () => void> = new Map();

    constructor() {
        this.state = ClientState.Idle;
    }

    handleEvent(ev: SessionProtocol.EventID): SessionProtocol.CmdID[] {
        switch (true) {
            case this.state === ClientState.Idle && ev === SessionProtocol.EventID.BackchannelReceived: {
                this.state = ClientState.ObtainBackchannelSecret;
                return [];
            }
            case this.state === ClientState.ObtainBackchannelSecret && ev === SessionProtocol.EventID.SecretParsed: {
                this.state = ClientState.ConnectRelay;
                return [];
            }
            case this.state === ClientState.ConnectRelay && ev === SessionProtocol.EventID.RelayConnected: {
                this.state = ClientState.GenKeyPair;
                return [];
            }
            case this.state === ClientState.GenKeyPair && ev === SessionProtocol.EventID.KeyPairGenerated: {
                this.actions.get(SessionProtocol.ActionID.SendPairHello)?.();
                this.state = ClientState.WaitAck;
                return [];
            }
            case this.state === ClientState.WaitAck && ev === SessionProtocol.EventID.RecvPairHelloAck: {
                this.actions.get(SessionProtocol.ActionID.DeriveSecret)?.();
                // received_backend_pub: recv_msg.pubkey (set by action)
                // client_shared_key: DeriveKey("client_pub", recv_msg.pubkey) (set by action)
                this.state = ClientState.E2EReady;
                return [];
            }
            case this.state === ClientState.E2EReady && ev === SessionProtocol.EventID.RecvPairConfirm: {
                // client_code: DeriveCode(received_backend_pub, "client_pub") (set by action)
                this.state = ClientState.ShowCode;
                return [];
            }
            case this.state === ClientState.ShowCode && ev === SessionProtocol.EventID.CodeDisplayed: {
                this.state = ClientState.WaitPairComplete;
                return [];
            }
            case this.state === ClientState.WaitPairComplete && ev === SessionProtocol.EventID.RecvPairComplete: {
                this.actions.get(SessionProtocol.ActionID.StoreSecret)?.();
                this.state = ClientState.Paired;
                return [];
            }
            case this.state === ClientState.Paired && ev === SessionProtocol.EventID.AppLaunch: {
                this.state = ClientState.Reconnect;
                return [];
            }
            case this.state === ClientState.Reconnect && ev === SessionProtocol.EventID.RelayConnected: {
                this.state = ClientState.SendAuth;
                return [];
            }
            case this.state === ClientState.SendAuth && ev === SessionProtocol.EventID.RecvAuthOk: {
                this.state = ClientState.SessionActive;
                return [];
            }
            case this.state === ClientState.SessionActive && ev === SessionProtocol.EventID.SessionEstablished: {
                this.state = ClientState.RelayConnected;
                return [];
            }
            case this.state === ClientState.RelayConnected && ev === SessionProtocol.EventID.RecvLanOffer && this.guards.get(SessionProtocol.GuardID.LanEnabled)?.() === true: {
                this.actions.get(SessionProtocol.ActionID.DialLan)?.();
                this.state = ClientState.LANConnecting;
                return [SessionProtocol.CmdID.DialLan];
            }
            case this.state === ClientState.RelayConnected && ev === SessionProtocol.EventID.RecvLanOffer && this.guards.get(SessionProtocol.GuardID.LanDisabled)?.() === true: {
                this.state = ClientState.RelayConnected;
                return [];
            }
            case this.state === ClientState.LANConnecting && ev === SessionProtocol.EventID.LanDialOk: {
                this.state = ClientState.LANVerifying;
                return [SessionProtocol.CmdID.SendLanVerify];
            }
            case this.state === ClientState.LANConnecting && ev === SessionProtocol.EventID.LanDialFailed: {
                this.state = ClientState.RelayConnected;
                return [];
            }
            case this.state === ClientState.LANVerifying && ev === SessionProtocol.EventID.RecvLanConfirm: {
                this.actions.get(SessionProtocol.ActionID.ActivateLan)?.();
                this.cActivePath = "lan";
                this.cDispatcherPath = "lan";
                this.lanSignal = "ready";
                this.state = ClientState.LANActive;
                return [SessionProtocol.CmdID.StartLanStreamReader, SessionProtocol.CmdID.StartLanDgReader, SessionProtocol.CmdID.SignalLanReady, SessionProtocol.CmdID.SetCryptoDatagram];
            }
            case this.state === ClientState.LANVerifying && ev === SessionProtocol.EventID.VerifyTimeout: {
                this.cDispatcherPath = "relay";
                this.state = ClientState.RelayConnected;
                return [];
            }
            case this.state === ClientState.LANActive && ev === SessionProtocol.EventID.RecvPathPing: {
                this.state = ClientState.LANActive;
                return [SessionProtocol.CmdID.SendPathPong];
            }
            case this.state === ClientState.LANActive && ev === SessionProtocol.EventID.LanError: {
                this.actions.get(SessionProtocol.ActionID.FallbackToRelay)?.();
                this.cActivePath = "relay";
                this.cDispatcherPath = "relay";
                this.lanSignal = "pending";
                this.state = ClientState.RelayFallback;
                return [SessionProtocol.CmdID.StopLanStreamReader, SessionProtocol.CmdID.StopLanDgReader, SessionProtocol.CmdID.CloseLanPath, SessionProtocol.CmdID.ResetLanReady];
            }
            case this.state === ClientState.LANActive && ev === SessionProtocol.EventID.LanStreamError: {
                this.actions.get(SessionProtocol.ActionID.FallbackToRelay)?.();
                this.cActivePath = "relay";
                this.cDispatcherPath = "relay";
                this.lanSignal = "pending";
                this.state = ClientState.RelayFallback;
                return [SessionProtocol.CmdID.StopLanStreamReader, SessionProtocol.CmdID.StopLanDgReader, SessionProtocol.CmdID.CloseLanPath, SessionProtocol.CmdID.ResetLanReady];
            }
            case this.state === ClientState.RelayFallback && ev === SessionProtocol.EventID.RelayOk: {
                this.state = ClientState.RelayConnected;
                return [];
            }
            case this.state === ClientState.LANActive && ev === SessionProtocol.EventID.RecvLanOffer && this.guards.get(SessionProtocol.GuardID.LanEnabled)?.() === true: {
                this.actions.get(SessionProtocol.ActionID.DialLan)?.();
                this.state = ClientState.LANConnecting;
                return [SessionProtocol.CmdID.StopLanStreamReader, SessionProtocol.CmdID.StopLanDgReader, SessionProtocol.CmdID.CloseLanPath, SessionProtocol.CmdID.DialLan];
            }
            case this.state === ClientState.LANConnecting && ev === SessionProtocol.EventID.AppForceFallback: {
                this.state = ClientState.RelayConnected;
                return [];
            }
            case this.state === ClientState.LANVerifying && ev === SessionProtocol.EventID.AppForceFallback: {
                this.cDispatcherPath = "relay";
                this.state = ClientState.RelayConnected;
                return [SessionProtocol.CmdID.StopLanStreamReader, SessionProtocol.CmdID.StopLanDgReader, SessionProtocol.CmdID.CloseLanPath];
            }
            case this.state === ClientState.LANActive && ev === SessionProtocol.EventID.AppForceFallback: {
                this.actions.get(SessionProtocol.ActionID.FallbackToRelay)?.();
                this.cActivePath = "relay";
                this.cDispatcherPath = "relay";
                this.lanSignal = "pending";
                this.state = ClientState.RelayConnected;
                return [SessionProtocol.CmdID.StopLanStreamReader, SessionProtocol.CmdID.StopLanDgReader, SessionProtocol.CmdID.CloseLanPath, SessionProtocol.CmdID.ResetLanReady];
            }
            case this.state === ClientState.RelayConnected && ev === SessionProtocol.EventID.Disconnect: {
                this.state = ClientState.Paired;
                return [];
            }
            case this.state === ClientState.RelayConnected && ev === SessionProtocol.EventID.AppSend: {
                this.state = ClientState.RelayConnected;
                return [SessionProtocol.CmdID.WriteActiveStream];
            }
            case this.state === ClientState.LANConnecting && ev === SessionProtocol.EventID.AppSend: {
                this.state = ClientState.LANConnecting;
                return [SessionProtocol.CmdID.WriteActiveStream];
            }
            case this.state === ClientState.LANVerifying && ev === SessionProtocol.EventID.AppSend: {
                this.state = ClientState.LANVerifying;
                return [SessionProtocol.CmdID.WriteActiveStream];
            }
            case this.state === ClientState.LANActive && ev === SessionProtocol.EventID.AppSend: {
                this.state = ClientState.LANActive;
                return [SessionProtocol.CmdID.WriteActiveStream];
            }
            case this.state === ClientState.RelayFallback && ev === SessionProtocol.EventID.AppSend: {
                this.state = ClientState.RelayFallback;
                return [SessionProtocol.CmdID.WriteActiveStream];
            }
            case this.state === ClientState.RelayConnected && ev === SessionProtocol.EventID.RelayStreamData: {
                this.state = ClientState.RelayConnected;
                return [SessionProtocol.CmdID.DeliverRecv];
            }
            case this.state === ClientState.LANConnecting && ev === SessionProtocol.EventID.RelayStreamData: {
                this.state = ClientState.LANConnecting;
                return [SessionProtocol.CmdID.DeliverRecv];
            }
            case this.state === ClientState.LANVerifying && ev === SessionProtocol.EventID.RelayStreamData: {
                this.state = ClientState.LANVerifying;
                return [SessionProtocol.CmdID.DeliverRecv];
            }
            case this.state === ClientState.LANActive && ev === SessionProtocol.EventID.RelayStreamData: {
                this.state = ClientState.LANActive;
                return [SessionProtocol.CmdID.DeliverRecv];
            }
            case this.state === ClientState.RelayFallback && ev === SessionProtocol.EventID.RelayStreamData: {
                this.state = ClientState.RelayFallback;
                return [SessionProtocol.CmdID.DeliverRecv];
            }
            case this.state === ClientState.RelayConnected && ev === SessionProtocol.EventID.RelayStreamError: {
                this.state = ClientState.RelayConnected;
                return [SessionProtocol.CmdID.DeliverRecvError];
            }
            case this.state === ClientState.LANConnecting && ev === SessionProtocol.EventID.RelayStreamError: {
                this.state = ClientState.LANConnecting;
                return [SessionProtocol.CmdID.DeliverRecvError];
            }
            case this.state === ClientState.LANVerifying && ev === SessionProtocol.EventID.RelayStreamError: {
                this.state = ClientState.LANVerifying;
                return [SessionProtocol.CmdID.DeliverRecvError];
            }
            case this.state === ClientState.LANActive && ev === SessionProtocol.EventID.RelayStreamError: {
                this.state = ClientState.LANActive;
                return [SessionProtocol.CmdID.DeliverRecvError];
            }
            case this.state === ClientState.RelayFallback && ev === SessionProtocol.EventID.RelayStreamError: {
                this.state = ClientState.RelayFallback;
                return [SessionProtocol.CmdID.DeliverRecvError];
            }
            case this.state === ClientState.RelayConnected && ev === SessionProtocol.EventID.AppSendDatagram: {
                this.state = ClientState.RelayConnected;
                return [SessionProtocol.CmdID.SendActiveDatagram];
            }
            case this.state === ClientState.LANConnecting && ev === SessionProtocol.EventID.AppSendDatagram: {
                this.state = ClientState.LANConnecting;
                return [SessionProtocol.CmdID.SendActiveDatagram];
            }
            case this.state === ClientState.LANVerifying && ev === SessionProtocol.EventID.AppSendDatagram: {
                this.state = ClientState.LANVerifying;
                return [SessionProtocol.CmdID.SendActiveDatagram];
            }
            case this.state === ClientState.LANActive && ev === SessionProtocol.EventID.AppSendDatagram: {
                this.state = ClientState.LANActive;
                return [SessionProtocol.CmdID.SendActiveDatagram];
            }
            case this.state === ClientState.RelayFallback && ev === SessionProtocol.EventID.AppSendDatagram: {
                this.state = ClientState.RelayFallback;
                return [SessionProtocol.CmdID.SendActiveDatagram];
            }
            case this.state === ClientState.RelayConnected && ev === SessionProtocol.EventID.RelayDatagram: {
                this.state = ClientState.RelayConnected;
                return [SessionProtocol.CmdID.DeliverRecvDatagram];
            }
            case this.state === ClientState.LANConnecting && ev === SessionProtocol.EventID.RelayDatagram: {
                this.state = ClientState.LANConnecting;
                return [SessionProtocol.CmdID.DeliverRecvDatagram];
            }
            case this.state === ClientState.LANVerifying && ev === SessionProtocol.EventID.RelayDatagram: {
                this.state = ClientState.LANVerifying;
                return [SessionProtocol.CmdID.DeliverRecvDatagram];
            }
            case this.state === ClientState.LANActive && ev === SessionProtocol.EventID.RelayDatagram: {
                this.state = ClientState.LANActive;
                return [SessionProtocol.CmdID.DeliverRecvDatagram];
            }
            case this.state === ClientState.RelayFallback && ev === SessionProtocol.EventID.RelayDatagram: {
                this.state = ClientState.RelayFallback;
                return [SessionProtocol.CmdID.DeliverRecvDatagram];
            }
            case this.state === ClientState.LANActive && ev === SessionProtocol.EventID.LanStreamData: {
                this.state = ClientState.LANActive;
                return [SessionProtocol.CmdID.DeliverRecv];
            }
            case this.state === ClientState.LANActive && ev === SessionProtocol.EventID.LanDatagram: {
                this.state = ClientState.LANActive;
                return [SessionProtocol.CmdID.DeliverRecvDatagram];
            }
        }
        return [];
    }
}

/** RelayMachine is the generated state machine for the relay actor. */
export class RelayMachine {
    readonly protocol = SessionProtocol;
    state: RelayState;
    relayBridge: string = "idle"; // relay bridge state
    guards: Map<SessionProtocol.GuardID, () => boolean> = new Map();
    actions: Map<SessionProtocol.ActionID, () => void> = new Map();

    constructor() {
        this.state = RelayState.Idle;
    }

    handleEvent(ev: SessionProtocol.EventID): SessionProtocol.CmdID[] {
        switch (true) {
            case this.state === RelayState.Idle && ev === SessionProtocol.EventID.BackendRegister: {
                this.state = RelayState.BackendRegistered;
                return [];
            }
            case this.state === RelayState.BackendRegistered && ev === SessionProtocol.EventID.ClientConnect: {
                this.actions.get(SessionProtocol.ActionID.BridgeStreams)?.();
                this.relayBridge = "active";
                this.state = RelayState.Bridged;
                return [];
            }
            case this.state === RelayState.Bridged && ev === SessionProtocol.EventID.ClientDisconnect: {
                this.actions.get(SessionProtocol.ActionID.Unbridge)?.();
                this.relayBridge = "idle";
                this.state = RelayState.BackendRegistered;
                return [];
            }
            case this.state === RelayState.BackendRegistered && ev === SessionProtocol.EventID.BackendDisconnect: {
                this.state = RelayState.Idle;
                return [];
            }
        }
        return [];
    }
}

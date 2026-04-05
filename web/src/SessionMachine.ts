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
    LANDegraded = "LANDegraded",
    RelayBackoff = "RelayBackoff",
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
        { from: "RelayConnected", to: "RelayConnected", on: "app_send", onKind: "internal" },
        { from: "RelayConnected", to: "RelayConnected", on: "relay_stream_data", onKind: "internal" },
        { from: "LANOffered", to: "LANOffered", on: "app_send", onKind: "internal" },
        { from: "LANOffered", to: "LANOffered", on: "relay_stream_data", onKind: "internal" },
        { from: "LANActive", to: "LANActive", on: "app_send", onKind: "internal" },
        { from: "LANActive", to: "LANActive", on: "lan_stream_data", onKind: "internal" },
        { from: "LANActive", to: "LANActive", on: "relay_stream_data", onKind: "internal" },
        { from: "LANDegraded", to: "LANDegraded", on: "app_send", onKind: "internal" },
        { from: "LANDegraded", to: "LANDegraded", on: "lan_stream_data", onKind: "internal" },
        { from: "LANDegraded", to: "LANDegraded", on: "relay_stream_data", onKind: "internal" },
        { from: "RelayBackoff", to: "RelayBackoff", on: "app_send", onKind: "internal" },
        { from: "RelayBackoff", to: "RelayBackoff", on: "relay_stream_data", onKind: "internal" },
        { from: "RelayConnected", to: "RelayConnected", on: "app_send_datagram", onKind: "internal" },
        { from: "RelayConnected", to: "RelayConnected", on: "relay_datagram", onKind: "internal" },
        { from: "LANOffered", to: "LANOffered", on: "app_send_datagram", onKind: "internal" },
        { from: "LANOffered", to: "LANOffered", on: "relay_datagram", onKind: "internal" },
        { from: "LANActive", to: "LANActive", on: "app_send_datagram", onKind: "internal" },
        { from: "LANActive", to: "LANActive", on: "lan_datagram", onKind: "internal" },
        { from: "LANActive", to: "LANActive", on: "relay_datagram", onKind: "internal" },
        { from: "LANDegraded", to: "LANDegraded", on: "app_send_datagram", onKind: "internal" },
        { from: "LANDegraded", to: "LANDegraded", on: "lan_datagram", onKind: "internal" },
        { from: "LANDegraded", to: "LANDegraded", on: "relay_datagram", onKind: "internal" },
        { from: "RelayBackoff", to: "RelayBackoff", on: "app_send_datagram", onKind: "internal" },
        { from: "RelayBackoff", to: "RelayBackoff", on: "relay_datagram", onKind: "internal" },
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
        { from: "RelayConnected", to: "RelayConnected", on: "app_send", onKind: "internal" },
        { from: "RelayConnected", to: "RelayConnected", on: "relay_stream_data", onKind: "internal" },
        { from: "LANConnecting", to: "LANConnecting", on: "app_send", onKind: "internal" },
        { from: "LANConnecting", to: "LANConnecting", on: "relay_stream_data", onKind: "internal" },
        { from: "LANVerifying", to: "LANVerifying", on: "app_send", onKind: "internal" },
        { from: "LANVerifying", to: "LANVerifying", on: "relay_stream_data", onKind: "internal" },
        { from: "LANActive", to: "LANActive", on: "app_send", onKind: "internal" },
        { from: "LANActive", to: "LANActive", on: "lan_stream_data", onKind: "internal" },
        { from: "LANActive", to: "LANActive", on: "relay_stream_data", onKind: "internal" },
        { from: "RelayFallback", to: "RelayFallback", on: "app_send", onKind: "internal" },
        { from: "RelayFallback", to: "RelayFallback", on: "relay_stream_data", onKind: "internal" },
        { from: "RelayConnected", to: "RelayConnected", on: "app_send_datagram", onKind: "internal" },
        { from: "RelayConnected", to: "RelayConnected", on: "relay_datagram", onKind: "internal" },
        { from: "LANConnecting", to: "LANConnecting", on: "app_send_datagram", onKind: "internal" },
        { from: "LANConnecting", to: "LANConnecting", on: "relay_datagram", onKind: "internal" },
        { from: "LANVerifying", to: "LANVerifying", on: "app_send_datagram", onKind: "internal" },
        { from: "LANVerifying", to: "LANVerifying", on: "relay_datagram", onKind: "internal" },
        { from: "LANActive", to: "LANActive", on: "app_send_datagram", onKind: "internal" },
        { from: "LANActive", to: "LANActive", on: "lan_datagram", onKind: "internal" },
        { from: "LANActive", to: "LANActive", on: "relay_datagram", onKind: "internal" },
        { from: "RelayFallback", to: "RelayFallback", on: "app_send_datagram", onKind: "internal" },
        { from: "RelayFallback", to: "RelayFallback", on: "relay_datagram", onKind: "internal" },
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

/** BackendMachine is the generated state machine for the backend actor. */
export class BackendMachine {
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
    guards: Map<GuardID, () => boolean> = new Map();
    actions: Map<ActionID, () => void> = new Map();

    constructor() {
        this.state = BackendState.Idle;
    }

    handleEvent(ev: EventID): CmdID[] {
        switch (true) {
            case this.state === BackendState.Idle && ev === EventID.CliInitPair: {
                this.actions.get(ActionID.GenerateToken)?.();
                this.currentToken = "tok_1";
                // active_tokens: active_tokens \union {"tok_1"} (set by action)
                this.state = BackendState.GenerateToken;
                return [];
            }
            case this.state === BackendState.GenerateToken && ev === EventID.TokenCreated: {
                this.actions.get(ActionID.RegisterRelay)?.();
                this.state = BackendState.RegisterRelay;
                return [];
            }
            case this.state === BackendState.RegisterRelay && ev === EventID.RelayRegistered: {
                this.secretPublished = true;
                this.state = BackendState.WaitingForClient;
                return [];
            }
            case this.state === BackendState.WaitingForClient && ev === EventID.RecvPairHello && this.guards.get(GuardID.TokenValid)?.() === true: {
                this.actions.get(ActionID.DeriveSecret)?.();
                // received_client_pub: recv_msg.pubkey (set by action)
                this.backendEcdhPub = "backend_pub";
                // backend_shared_key: DeriveKey("backend_pub", recv_msg.pubkey) (set by action)
                // backend_code: DeriveCode("backend_pub", recv_msg.pubkey) (set by action)
                this.state = BackendState.DeriveSecret;
                return [];
            }
            case this.state === BackendState.WaitingForClient && ev === EventID.RecvPairHello && this.guards.get(GuardID.TokenInvalid)?.() === true: {
                this.state = BackendState.Idle;
                return [];
            }
            case this.state === BackendState.DeriveSecret && ev === EventID.EcdhComplete: {
                this.state = BackendState.SendAck;
                return [];
            }
            case this.state === BackendState.SendAck && ev === EventID.SignalCodeDisplay: {
                this.state = BackendState.WaitingForCode;
                return [];
            }
            case this.state === BackendState.WaitingForCode && ev === EventID.CliCodeEntered: {
                // received_code: cli_entered_code (set by action)
                this.state = BackendState.ValidateCode;
                return [];
            }
            case this.state === BackendState.ValidateCode && ev === EventID.CheckCode && this.guards.get(GuardID.CodeCorrect)?.() === true: {
                this.state = BackendState.StorePaired;
                return [];
            }
            case this.state === BackendState.ValidateCode && ev === EventID.CheckCode && this.guards.get(GuardID.CodeWrong)?.() === true: {
                // code_attempts: code_attempts + 1 (set by action)
                this.state = BackendState.Idle;
                return [];
            }
            case this.state === BackendState.StorePaired && ev === EventID.Finalise: {
                this.actions.get(ActionID.StoreDevice)?.();
                this.deviceSecret = "dev_secret_1";
                // paired_devices: paired_devices \union {"device_1"} (set by action)
                // active_tokens: active_tokens \ {current_token} (set by action)
                // used_tokens: used_tokens \union {current_token} (set by action)
                this.state = BackendState.Paired;
                return [];
            }
            case this.state === BackendState.Paired && ev === EventID.RecvAuthRequest: {
                // received_device_id: recv_msg.device_id (set by action)
                // received_auth_nonce: recv_msg.nonce (set by action)
                this.state = BackendState.AuthCheck;
                return [];
            }
            case this.state === BackendState.AuthCheck && ev === EventID.Verify && this.guards.get(GuardID.DeviceKnown)?.() === true: {
                this.actions.get(ActionID.VerifyDevice)?.();
                // auth_nonces_used: auth_nonces_used \union {received_auth_nonce} (set by action)
                this.state = BackendState.SessionActive;
                return [];
            }
            case this.state === BackendState.AuthCheck && ev === EventID.Verify && this.guards.get(GuardID.DeviceUnknown)?.() === true: {
                this.state = BackendState.Idle;
                return [];
            }
            case this.state === BackendState.SessionActive && ev === EventID.SessionEstablished: {
                this.state = BackendState.RelayConnected;
                return [];
            }
            case this.state === BackendState.RelayConnected && ev === EventID.AppSend: {
                this.state = BackendState.RelayConnected;
                return [CmdID.WriteActiveStream];
            }
            case this.state === BackendState.RelayConnected && ev === EventID.RelayStreamData: {
                this.state = BackendState.RelayConnected;
                return [CmdID.DeliverRecv];
            }
            case this.state === BackendState.LANOffered && ev === EventID.AppSend: {
                this.state = BackendState.LANOffered;
                return [CmdID.WriteActiveStream];
            }
            case this.state === BackendState.LANOffered && ev === EventID.RelayStreamData: {
                this.state = BackendState.LANOffered;
                return [CmdID.DeliverRecv];
            }
            case this.state === BackendState.LANActive && ev === EventID.AppSend: {
                this.state = BackendState.LANActive;
                return [CmdID.WriteActiveStream];
            }
            case this.state === BackendState.LANActive && ev === EventID.LanStreamData: {
                this.state = BackendState.LANActive;
                return [CmdID.DeliverRecv];
            }
            case this.state === BackendState.LANActive && ev === EventID.RelayStreamData: {
                this.state = BackendState.LANActive;
                return [CmdID.DeliverRecv];
            }
            case this.state === BackendState.LANDegraded && ev === EventID.AppSend: {
                this.state = BackendState.LANDegraded;
                return [CmdID.WriteActiveStream];
            }
            case this.state === BackendState.LANDegraded && ev === EventID.LanStreamData: {
                this.state = BackendState.LANDegraded;
                return [CmdID.DeliverRecv];
            }
            case this.state === BackendState.LANDegraded && ev === EventID.RelayStreamData: {
                this.state = BackendState.LANDegraded;
                return [CmdID.DeliverRecv];
            }
            case this.state === BackendState.RelayBackoff && ev === EventID.AppSend: {
                this.state = BackendState.RelayBackoff;
                return [CmdID.WriteActiveStream];
            }
            case this.state === BackendState.RelayBackoff && ev === EventID.RelayStreamData: {
                this.state = BackendState.RelayBackoff;
                return [CmdID.DeliverRecv];
            }
            case this.state === BackendState.RelayConnected && ev === EventID.AppSendDatagram: {
                this.state = BackendState.RelayConnected;
                return [CmdID.SendActiveDatagram];
            }
            case this.state === BackendState.RelayConnected && ev === EventID.RelayDatagram: {
                this.state = BackendState.RelayConnected;
                return [CmdID.DeliverRecvDatagram];
            }
            case this.state === BackendState.LANOffered && ev === EventID.AppSendDatagram: {
                this.state = BackendState.LANOffered;
                return [CmdID.SendActiveDatagram];
            }
            case this.state === BackendState.LANOffered && ev === EventID.RelayDatagram: {
                this.state = BackendState.LANOffered;
                return [CmdID.DeliverRecvDatagram];
            }
            case this.state === BackendState.LANActive && ev === EventID.AppSendDatagram: {
                this.state = BackendState.LANActive;
                return [CmdID.SendActiveDatagram];
            }
            case this.state === BackendState.LANActive && ev === EventID.LanDatagram: {
                this.state = BackendState.LANActive;
                return [CmdID.DeliverRecvDatagram];
            }
            case this.state === BackendState.LANActive && ev === EventID.RelayDatagram: {
                this.state = BackendState.LANActive;
                return [CmdID.DeliverRecvDatagram];
            }
            case this.state === BackendState.LANDegraded && ev === EventID.AppSendDatagram: {
                this.state = BackendState.LANDegraded;
                return [CmdID.SendActiveDatagram];
            }
            case this.state === BackendState.LANDegraded && ev === EventID.LanDatagram: {
                this.state = BackendState.LANDegraded;
                return [CmdID.DeliverRecvDatagram];
            }
            case this.state === BackendState.LANDegraded && ev === EventID.RelayDatagram: {
                this.state = BackendState.LANDegraded;
                return [CmdID.DeliverRecvDatagram];
            }
            case this.state === BackendState.RelayBackoff && ev === EventID.AppSendDatagram: {
                this.state = BackendState.RelayBackoff;
                return [CmdID.SendActiveDatagram];
            }
            case this.state === BackendState.RelayBackoff && ev === EventID.RelayDatagram: {
                this.state = BackendState.RelayBackoff;
                return [CmdID.DeliverRecvDatagram];
            }
            case this.state === BackendState.RelayConnected && ev === EventID.LanServerReady: {
                this.state = BackendState.LANOffered;
                return [CmdID.SendLanOffer];
            }
            case this.state === BackendState.LANOffered && ev === EventID.RecvLanVerify && this.guards.get(GuardID.ChallengeValid)?.() === true: {
                this.actions.get(ActionID.ActivateLan)?.();
                this.pingFailures = 0;
                this.backoffLevel = 0;
                this.bActivePath = "lan";
                this.bDispatcherPath = "lan";
                this.monitorTarget = "lan";
                this.lanSignal = "ready";
                this.state = BackendState.LANActive;
                return [CmdID.SendLanConfirm, CmdID.StartLanStreamReader, CmdID.StartLanDgReader, CmdID.StartMonitor, CmdID.SignalLanReady, CmdID.SetCryptoDatagram];
            }
            case this.state === BackendState.LANOffered && ev === EventID.RecvLanVerify && this.guards.get(GuardID.ChallengeInvalid)?.() === true: {
                this.state = BackendState.RelayConnected;
                return [];
            }
            case this.state === BackendState.LANOffered && ev === EventID.OfferTimeout: {
                // backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
                this.lanSignal = "pending";
                this.state = BackendState.RelayBackoff;
                return [CmdID.ResetLanReady, CmdID.StartBackoffTimer];
            }
            case this.state === BackendState.LANActive && ev === EventID.PingTick: {
                this.state = BackendState.LANActive;
                return [CmdID.SendPathPing, CmdID.StartPongTimeout];
            }
            case this.state === BackendState.LANActive && ev === EventID.PingTimeout: {
                this.pingFailures = 1;
                this.state = BackendState.LANDegraded;
                return [];
            }
            case this.state === BackendState.LANDegraded && ev === EventID.PingTick: {
                this.state = BackendState.LANDegraded;
                return [CmdID.SendPathPing, CmdID.StartPongTimeout];
            }
            case this.state === BackendState.LANActive && ev === EventID.LanStreamError: {
                this.actions.get(ActionID.FallbackToRelay)?.();
                // backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
                this.bActivePath = "relay";
                this.bDispatcherPath = "relay";
                this.monitorTarget = "none";
                this.lanSignal = "pending";
                this.pingFailures = 0;
                this.state = BackendState.RelayBackoff;
                return [CmdID.StopMonitor, CmdID.StopLanStreamReader, CmdID.StopLanDgReader, CmdID.CloseLanPath, CmdID.ResetLanReady, CmdID.StartBackoffTimer];
            }
            case this.state === BackendState.LANDegraded && ev === EventID.LanStreamError: {
                this.actions.get(ActionID.FallbackToRelay)?.();
                // backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
                this.bActivePath = "relay";
                this.bDispatcherPath = "relay";
                this.monitorTarget = "none";
                this.lanSignal = "pending";
                this.pingFailures = 0;
                this.state = BackendState.RelayBackoff;
                return [CmdID.StopMonitor, CmdID.StopLanStreamReader, CmdID.StopLanDgReader, CmdID.CloseLanPath, CmdID.ResetLanReady, CmdID.StartBackoffTimer];
            }
            case this.state === BackendState.LANDegraded && ev === EventID.RecvPathPong: {
                this.actions.get(ActionID.ResetFailures)?.();
                this.pingFailures = 0;
                this.state = BackendState.LANActive;
                return [CmdID.CancelPongTimeout];
            }
            case this.state === BackendState.LANDegraded && ev === EventID.PingTimeout && this.guards.get(GuardID.UnderMaxFailures)?.() === true: {
                // ping_failures: ping_failures + 1 (set by action)
                this.state = BackendState.LANDegraded;
                return [];
            }
            case this.state === BackendState.LANDegraded && ev === EventID.PingTimeout && this.guards.get(GuardID.AtMaxFailures)?.() === true: {
                this.actions.get(ActionID.FallbackToRelay)?.();
                // backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
                this.bActivePath = "relay";
                this.bDispatcherPath = "relay";
                this.monitorTarget = "none";
                this.lanSignal = "pending";
                this.pingFailures = 0;
                this.state = BackendState.RelayBackoff;
                return [CmdID.StopMonitor, CmdID.StopLanStreamReader, CmdID.StopLanDgReader, CmdID.CloseLanPath, CmdID.ResetLanReady, CmdID.StartBackoffTimer];
            }
            case this.state === BackendState.RelayBackoff && ev === EventID.BackoffExpired: {
                this.state = BackendState.LANOffered;
                return [CmdID.SendLanOffer];
            }
            case this.state === BackendState.RelayBackoff && ev === EventID.LanServerChanged: {
                this.backoffLevel = 0;
                this.state = BackendState.LANOffered;
                return [CmdID.SendLanOffer];
            }
            case this.state === BackendState.RelayConnected && ev === EventID.ReadvertiseTick && this.guards.get(GuardID.LanServerAvailable)?.() === true: {
                this.state = BackendState.LANOffered;
                return [CmdID.SendLanOffer];
            }
            case this.state === BackendState.RelayConnected && ev === EventID.Disconnect: {
                this.state = BackendState.Paired;
                return [];
            }
        }
        return [];
    }
}

/** ClientMachine is the generated state machine for the client actor. */
export class ClientMachine {
    state: ClientState;
    receivedBackendPub: string = "none"; // pubkey client received in pair_hello_ack
    clientSharedKey: string = ""; // ECDH key derived by client
    clientCode: string = ""; // code computed by client
    cActivePath: string = "relay"; // client active path
    cDispatcherPath: string = "relay"; // client datagram dispatcher binding
    lanSignal: string = "pending"; // LANReady notification state
    guards: Map<GuardID, () => boolean> = new Map();
    actions: Map<ActionID, () => void> = new Map();

    constructor() {
        this.state = ClientState.Idle;
    }

    handleEvent(ev: EventID): CmdID[] {
        switch (true) {
            case this.state === ClientState.Idle && ev === EventID.BackchannelReceived: {
                this.state = ClientState.ObtainBackchannelSecret;
                return [];
            }
            case this.state === ClientState.ObtainBackchannelSecret && ev === EventID.SecretParsed: {
                this.state = ClientState.ConnectRelay;
                return [];
            }
            case this.state === ClientState.ConnectRelay && ev === EventID.RelayConnected: {
                this.state = ClientState.GenKeyPair;
                return [];
            }
            case this.state === ClientState.GenKeyPair && ev === EventID.KeyPairGenerated: {
                this.actions.get(ActionID.SendPairHello)?.();
                this.state = ClientState.WaitAck;
                return [];
            }
            case this.state === ClientState.WaitAck && ev === EventID.RecvPairHelloAck: {
                this.actions.get(ActionID.DeriveSecret)?.();
                // received_backend_pub: recv_msg.pubkey (set by action)
                // client_shared_key: DeriveKey("client_pub", recv_msg.pubkey) (set by action)
                this.state = ClientState.E2EReady;
                return [];
            }
            case this.state === ClientState.E2EReady && ev === EventID.RecvPairConfirm: {
                // client_code: DeriveCode(received_backend_pub, "client_pub") (set by action)
                this.state = ClientState.ShowCode;
                return [];
            }
            case this.state === ClientState.ShowCode && ev === EventID.CodeDisplayed: {
                this.state = ClientState.WaitPairComplete;
                return [];
            }
            case this.state === ClientState.WaitPairComplete && ev === EventID.RecvPairComplete: {
                this.actions.get(ActionID.StoreSecret)?.();
                this.state = ClientState.Paired;
                return [];
            }
            case this.state === ClientState.Paired && ev === EventID.AppLaunch: {
                this.state = ClientState.Reconnect;
                return [];
            }
            case this.state === ClientState.Reconnect && ev === EventID.RelayConnected: {
                this.state = ClientState.SendAuth;
                return [];
            }
            case this.state === ClientState.SendAuth && ev === EventID.RecvAuthOk: {
                this.state = ClientState.SessionActive;
                return [];
            }
            case this.state === ClientState.SessionActive && ev === EventID.SessionEstablished: {
                this.state = ClientState.RelayConnected;
                return [];
            }
            case this.state === ClientState.RelayConnected && ev === EventID.AppSend: {
                this.state = ClientState.RelayConnected;
                return [CmdID.WriteActiveStream];
            }
            case this.state === ClientState.RelayConnected && ev === EventID.RelayStreamData: {
                this.state = ClientState.RelayConnected;
                return [CmdID.DeliverRecv];
            }
            case this.state === ClientState.LANConnecting && ev === EventID.AppSend: {
                this.state = ClientState.LANConnecting;
                return [CmdID.WriteActiveStream];
            }
            case this.state === ClientState.LANConnecting && ev === EventID.RelayStreamData: {
                this.state = ClientState.LANConnecting;
                return [CmdID.DeliverRecv];
            }
            case this.state === ClientState.LANVerifying && ev === EventID.AppSend: {
                this.state = ClientState.LANVerifying;
                return [CmdID.WriteActiveStream];
            }
            case this.state === ClientState.LANVerifying && ev === EventID.RelayStreamData: {
                this.state = ClientState.LANVerifying;
                return [CmdID.DeliverRecv];
            }
            case this.state === ClientState.LANActive && ev === EventID.AppSend: {
                this.state = ClientState.LANActive;
                return [CmdID.WriteActiveStream];
            }
            case this.state === ClientState.LANActive && ev === EventID.LanStreamData: {
                this.state = ClientState.LANActive;
                return [CmdID.DeliverRecv];
            }
            case this.state === ClientState.LANActive && ev === EventID.RelayStreamData: {
                this.state = ClientState.LANActive;
                return [CmdID.DeliverRecv];
            }
            case this.state === ClientState.RelayFallback && ev === EventID.AppSend: {
                this.state = ClientState.RelayFallback;
                return [CmdID.WriteActiveStream];
            }
            case this.state === ClientState.RelayFallback && ev === EventID.RelayStreamData: {
                this.state = ClientState.RelayFallback;
                return [CmdID.DeliverRecv];
            }
            case this.state === ClientState.RelayConnected && ev === EventID.AppSendDatagram: {
                this.state = ClientState.RelayConnected;
                return [CmdID.SendActiveDatagram];
            }
            case this.state === ClientState.RelayConnected && ev === EventID.RelayDatagram: {
                this.state = ClientState.RelayConnected;
                return [CmdID.DeliverRecvDatagram];
            }
            case this.state === ClientState.LANConnecting && ev === EventID.AppSendDatagram: {
                this.state = ClientState.LANConnecting;
                return [CmdID.SendActiveDatagram];
            }
            case this.state === ClientState.LANConnecting && ev === EventID.RelayDatagram: {
                this.state = ClientState.LANConnecting;
                return [CmdID.DeliverRecvDatagram];
            }
            case this.state === ClientState.LANVerifying && ev === EventID.AppSendDatagram: {
                this.state = ClientState.LANVerifying;
                return [CmdID.SendActiveDatagram];
            }
            case this.state === ClientState.LANVerifying && ev === EventID.RelayDatagram: {
                this.state = ClientState.LANVerifying;
                return [CmdID.DeliverRecvDatagram];
            }
            case this.state === ClientState.LANActive && ev === EventID.AppSendDatagram: {
                this.state = ClientState.LANActive;
                return [CmdID.SendActiveDatagram];
            }
            case this.state === ClientState.LANActive && ev === EventID.LanDatagram: {
                this.state = ClientState.LANActive;
                return [CmdID.DeliverRecvDatagram];
            }
            case this.state === ClientState.LANActive && ev === EventID.RelayDatagram: {
                this.state = ClientState.LANActive;
                return [CmdID.DeliverRecvDatagram];
            }
            case this.state === ClientState.RelayFallback && ev === EventID.AppSendDatagram: {
                this.state = ClientState.RelayFallback;
                return [CmdID.SendActiveDatagram];
            }
            case this.state === ClientState.RelayFallback && ev === EventID.RelayDatagram: {
                this.state = ClientState.RelayFallback;
                return [CmdID.DeliverRecvDatagram];
            }
            case this.state === ClientState.RelayConnected && ev === EventID.RecvLanOffer && this.guards.get(GuardID.LanEnabled)?.() === true: {
                this.actions.get(ActionID.DialLan)?.();
                this.state = ClientState.LANConnecting;
                return [CmdID.DialLan];
            }
            case this.state === ClientState.RelayConnected && ev === EventID.RecvLanOffer && this.guards.get(GuardID.LanDisabled)?.() === true: {
                this.state = ClientState.RelayConnected;
                return [];
            }
            case this.state === ClientState.LANConnecting && ev === EventID.LanDialOk: {
                this.state = ClientState.LANVerifying;
                return [CmdID.SendLanVerify];
            }
            case this.state === ClientState.LANConnecting && ev === EventID.LanDialFailed: {
                this.state = ClientState.RelayConnected;
                return [];
            }
            case this.state === ClientState.LANVerifying && ev === EventID.RecvLanConfirm: {
                this.actions.get(ActionID.ActivateLan)?.();
                this.cActivePath = "lan";
                this.cDispatcherPath = "lan";
                this.lanSignal = "ready";
                this.state = ClientState.LANActive;
                return [CmdID.StartLanStreamReader, CmdID.StartLanDgReader, CmdID.SignalLanReady, CmdID.SetCryptoDatagram];
            }
            case this.state === ClientState.LANVerifying && ev === EventID.VerifyTimeout: {
                this.cDispatcherPath = "relay";
                this.state = ClientState.RelayConnected;
                return [];
            }
            case this.state === ClientState.LANActive && ev === EventID.RecvPathPing: {
                this.state = ClientState.LANActive;
                return [CmdID.SendPathPong];
            }
            case this.state === ClientState.LANActive && ev === EventID.LanError: {
                this.actions.get(ActionID.FallbackToRelay)?.();
                this.cActivePath = "relay";
                this.cDispatcherPath = "relay";
                this.lanSignal = "pending";
                this.state = ClientState.RelayFallback;
                return [CmdID.StopLanStreamReader, CmdID.StopLanDgReader, CmdID.CloseLanPath, CmdID.ResetLanReady];
            }
            case this.state === ClientState.RelayFallback && ev === EventID.RelayOk: {
                this.state = ClientState.RelayConnected;
                return [];
            }
            case this.state === ClientState.LANActive && ev === EventID.RecvLanOffer && this.guards.get(GuardID.LanEnabled)?.() === true: {
                this.actions.get(ActionID.DialLan)?.();
                this.state = ClientState.LANConnecting;
                return [CmdID.StopLanStreamReader, CmdID.StopLanDgReader, CmdID.CloseLanPath, CmdID.DialLan];
            }
            case this.state === ClientState.RelayConnected && ev === EventID.Disconnect: {
                this.state = ClientState.Paired;
                return [];
            }
        }
        return [];
    }
}

/** RelayMachine is the generated state machine for the relay actor. */
export class RelayMachine {
    state: RelayState;
    relayBridge: string = "idle"; // relay bridge state
    guards: Map<GuardID, () => boolean> = new Map();
    actions: Map<ActionID, () => void> = new Map();

    constructor() {
        this.state = RelayState.Idle;
    }

    handleEvent(ev: EventID): CmdID[] {
        switch (true) {
            case this.state === RelayState.Idle && ev === EventID.BackendRegister: {
                this.state = RelayState.BackendRegistered;
                return [];
            }
            case this.state === RelayState.BackendRegistered && ev === EventID.ClientConnect: {
                this.actions.get(ActionID.BridgeStreams)?.();
                this.relayBridge = "active";
                this.state = RelayState.Bridged;
                return [];
            }
            case this.state === RelayState.Bridged && ev === EventID.ClientDisconnect: {
                this.actions.get(ActionID.Unbridge)?.();
                this.relayBridge = "idle";
                this.state = RelayState.BackendRegistered;
                return [];
            }
            case this.state === RelayState.BackendRegistered && ev === EventID.BackendDisconnect: {
                this.state = RelayState.Idle;
                return [];
            }
        }
        return [];
    }
}

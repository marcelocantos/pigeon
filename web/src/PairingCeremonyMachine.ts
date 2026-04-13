// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Auto-generated from protocol definition. Do not edit.
// Source of truth: protocol/*.yaml

export enum PairingCeremonyServerPairingState {
    Idle = "Idle",
    GenerateToken = "GenerateToken",
    RegisterRelay = "RegisterRelay",
    WaitingForClient = "WaitingForClient",
    DeriveSecret = "DeriveSecret",
    SendAck = "SendAck",
    WaitingForCode = "WaitingForCode",
    ValidateCode = "ValidateCode",
    StorePaired = "StorePaired",
    PairingComplete = "PairingComplete",
}

export enum PairingCeremonyServerAuthState {
    Idle = "Idle",
    Paired = "Paired",
    AuthCheck = "AuthCheck",
    SessionActive = "SessionActive",
}

export enum PairingCeremonyIosPairingState {
    Idle = "Idle",
    ScanQR = "ScanQR",
    ConnectRelay = "ConnectRelay",
    GenKeyPair = "GenKeyPair",
    WaitAck = "WaitAck",
    E2EReady = "E2EReady",
    ShowCode = "ShowCode",
    WaitPairComplete = "WaitPairComplete",
    PairingComplete = "PairingComplete",
}

export enum PairingCeremonyIosAuthState {
    Idle = "Idle",
    Paired = "Paired",
    Reconnect = "Reconnect",
    SendAuth = "SendAuth",
    SessionActive = "SessionActive",
}

export enum PairingCeremonyCliState {
    Idle = "Idle",
    GetKey = "GetKey",
    BeginPair = "BeginPair",
    ShowQR = "ShowQR",
    PromptCode = "PromptCode",
    SubmitCode = "SubmitCode",
    Done = "Done",
}

/** The protocol transition table and shared type enums. */
export namespace PairingCeremonyProtocol {

    export enum MessageType {
        PairBegin = "pair_begin",
        TokenResponse = "token_response",
        PairHello = "pair_hello",
        PairHelloAck = "pair_hello_ack",
        PairConfirm = "pair_confirm",
        WaitingForCode = "waiting_for_code",
        CodeSubmit = "code_submit",
        PairComplete = "pair_complete",
        PairStatus = "pair_status",
        AuthRequest = "auth_request",
        AuthOk = "auth_ok",
    }

    export enum GuardID {
        TokenValid = "token_valid",
        TokenInvalid = "token_invalid",
        CodeCorrect = "code_correct",
        CodeWrong = "code_wrong",
        DeviceKnown = "device_known",
        DeviceUnknown = "device_unknown",
        NonceFresh = "nonce_fresh",
    }

    export enum ActionID {
        GenerateToken = "generate_token",
        RegisterRelay = "register_relay",
        DeriveSecret = "derive_secret",
        StoreDevice = "store_device",
        VerifyDevice = "verify_device",
        SendPairHello = "send_pair_hello",
        StoreSecret = "store_secret",
    }

    export enum EventID {
        TokenCreated = "token created",
        RelayRegistered = "relay registered",
        ECDHComplete = "ECDH complete",
        SignalCodeDisplay = "signal code display",
        CheckCode = "check code",
        Finalise = "finalise",
        CredentialReady = "credential_ready",
        Verify = "verify",
        Disconnect = "disconnect",
        UserScansQR = "user scans QR",
        QRParsed = "QR parsed",
        RelayConnected = "relay connected",
        KeyPairGenerated = "key pair generated",
        CodeDisplayed = "code displayed",
        AppLaunch = "app launch",
        CliInit = "cli --init",
        KeyStored = "key stored",
        UserEntersCode = "user enters code",
        RecvPairBegin = "recv_pair_begin",
        RecvPairHello = "recv_pair_hello",
        RecvCodeSubmit = "recv_code_submit",
        RecvAuthRequest = "recv_auth_request",
        RecvPairHelloAck = "recv_pair_hello_ack",
        RecvPairConfirm = "recv_pair_confirm",
        RecvPairComplete = "recv_pair_complete",
        RecvAuthOk = "recv_auth_ok",
        RecvTokenResponse = "recv_token_response",
        RecvWaitingForCode = "recv_waiting_for_code",
        RecvPairStatus = "recv_pair_status",
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

    /** server/pairing transition table. */
    export const serverPairingTable: ActorTable = {
        initial: PairingCeremonyServerPairingState.Idle,
        transitions: [
            { from: "Idle", to: "GenerateToken", on: "pair_begin", onKind: "recv", action: "generate_token" },
            { from: "GenerateToken", to: "RegisterRelay", on: "token created", onKind: "internal", action: "register_relay" },
            { from: "RegisterRelay", to: "WaitingForClient", on: "relay registered", onKind: "internal", sends: [{ to: "cli", msg: "token_response" }] },
            { from: "WaitingForClient", to: "DeriveSecret", on: "pair_hello", onKind: "recv", guard: "token_valid", action: "derive_secret" },
            { from: "WaitingForClient", to: "Idle", on: "pair_hello", onKind: "recv", guard: "token_invalid" },
            { from: "DeriveSecret", to: "SendAck", on: "ECDH complete", onKind: "internal", sends: [{ to: "ios", msg: "pair_hello_ack" }] },
            { from: "SendAck", to: "WaitingForCode", on: "signal code display", onKind: "internal", sends: [{ to: "ios", msg: "pair_confirm" }, { to: "cli", msg: "waiting_for_code" }] },
            { from: "WaitingForCode", to: "ValidateCode", on: "code_submit", onKind: "recv" },
            { from: "ValidateCode", to: "StorePaired", on: "check code", onKind: "internal", guard: "code_correct" },
            { from: "ValidateCode", to: "Idle", on: "check code", onKind: "internal", guard: "code_wrong" },
            { from: "StorePaired", to: "PairingComplete", on: "finalise", onKind: "internal", action: "store_device", sends: [{ to: "ios", msg: "pair_complete" }, { to: "cli", msg: "pair_status" }] },
        ],
    };

    /** server/auth transition table. */
    export const serverAuthTable: ActorTable = {
        initial: PairingCeremonyServerAuthState.Idle,
        transitions: [
            { from: "Idle", to: "Paired", on: "credential_ready", onKind: "internal" },
            { from: "Paired", to: "AuthCheck", on: "auth_request", onKind: "recv" },
            { from: "AuthCheck", to: "SessionActive", on: "verify", onKind: "internal", guard: "device_known", action: "verify_device", sends: [{ to: "ios", msg: "auth_ok" }] },
            { from: "AuthCheck", to: "Idle", on: "verify", onKind: "internal", guard: "device_unknown" },
            { from: "SessionActive", to: "Paired", on: "disconnect", onKind: "internal" },
        ],
    };

    /** ios/pairing transition table. */
    export const iosPairingTable: ActorTable = {
        initial: PairingCeremonyIosPairingState.Idle,
        transitions: [
            { from: "Idle", to: "ScanQR", on: "user scans QR", onKind: "internal" },
            { from: "ScanQR", to: "ConnectRelay", on: "QR parsed", onKind: "internal" },
            { from: "ConnectRelay", to: "GenKeyPair", on: "relay connected", onKind: "internal" },
            { from: "GenKeyPair", to: "WaitAck", on: "key pair generated", onKind: "internal", action: "send_pair_hello", sends: [{ to: "server", msg: "pair_hello" }] },
            { from: "WaitAck", to: "E2EReady", on: "pair_hello_ack", onKind: "recv", action: "derive_secret" },
            { from: "E2EReady", to: "ShowCode", on: "pair_confirm", onKind: "recv" },
            { from: "ShowCode", to: "WaitPairComplete", on: "code displayed", onKind: "internal" },
            { from: "WaitPairComplete", to: "PairingComplete", on: "pair_complete", onKind: "recv", action: "store_secret" },
        ],
    };

    /** ios/auth transition table. */
    export const iosAuthTable: ActorTable = {
        initial: PairingCeremonyIosAuthState.Idle,
        transitions: [
            { from: "Idle", to: "Paired", on: "credential_ready", onKind: "internal" },
            { from: "Paired", to: "Reconnect", on: "app launch", onKind: "internal" },
            { from: "Reconnect", to: "SendAuth", on: "relay connected", onKind: "internal", sends: [{ to: "server", msg: "auth_request" }] },
            { from: "SendAuth", to: "SessionActive", on: "auth_ok", onKind: "recv" },
            { from: "SessionActive", to: "Paired", on: "disconnect", onKind: "internal" },
        ],
    };

    /** cli transition table. */
    export const cliTable: ActorTable = {
        initial: PairingCeremonyCliState.Idle,
        transitions: [
            { from: "Idle", to: "GetKey", on: "cli --init", onKind: "internal" },
            { from: "GetKey", to: "BeginPair", on: "key stored", onKind: "internal", sends: [{ to: "server", msg: "pair_begin" }] },
            { from: "BeginPair", to: "ShowQR", on: "token_response", onKind: "recv" },
            { from: "ShowQR", to: "PromptCode", on: "waiting_for_code", onKind: "recv" },
            { from: "PromptCode", to: "SubmitCode", on: "user enters code", onKind: "internal", sends: [{ to: "server", msg: "code_submit" }] },
            { from: "SubmitCode", to: "Done", on: "pair_status", onKind: "recv" },
        ],
    };

}

/** PairingCeremonyServerPairingMachine is the generated state machine for server/pairing. */
export class PairingCeremonyServerPairingMachine {
    readonly protocol = PairingCeremonyProtocol;
    state: PairingCeremonyServerPairingState;
    currentToken: string = "none"; // pairing token currently in play
    activeTokens: string = ""; // set of valid (non-revoked) tokens
    usedTokens: string = ""; // set of revoked tokens
    serverEcdhPub: string = "none"; // server ECDH public key
    receivedClientPub: string = "none"; // pubkey server received in pair_hello (may be adversary's)
    serverSharedKey: string = ""; // ECDH key derived by server (tuple to match DeriveKey output type)
    serverCode: string = ""; // code computed by server from its view of the pubkeys (tuple to match DeriveCode output type)
    receivedCode: string = ""; // code received in code_submit (tuple to match DeriveCode output type)
    codeAttempts: number = 0; // failed code submission attempts
    deviceSecret: string = "none"; // persistent device secret
    pairedDevices: string = ""; // device IDs that completed pairing
    guards: Map<PairingCeremonyProtocol.GuardID, () => boolean> = new Map();
    actions: Map<PairingCeremonyProtocol.ActionID, () => void> = new Map();

    constructor() {
        this.state = PairingCeremonyServerPairingState.Idle;
    }

    step(ev: PairingCeremonyProtocol.EventID): boolean {
        switch (true) {
            case this.state === PairingCeremonyServerPairingState.GenerateToken && ev === PairingCeremonyProtocol.EventID.TokenCreated: {
                this.actions.get(PairingCeremonyProtocol.ActionID.RegisterRelay)?.();
                this.state = PairingCeremonyServerPairingState.RegisterRelay;
                return true;
            }
            case this.state === PairingCeremonyServerPairingState.RegisterRelay && ev === PairingCeremonyProtocol.EventID.RelayRegistered: {
                this.state = PairingCeremonyServerPairingState.WaitingForClient;
                return true;
            }
            case this.state === PairingCeremonyServerPairingState.DeriveSecret && ev === PairingCeremonyProtocol.EventID.ECDHComplete: {
                this.state = PairingCeremonyServerPairingState.SendAck;
                return true;
            }
            case this.state === PairingCeremonyServerPairingState.SendAck && ev === PairingCeremonyProtocol.EventID.SignalCodeDisplay: {
                this.state = PairingCeremonyServerPairingState.WaitingForCode;
                return true;
            }
            case this.state === PairingCeremonyServerPairingState.ValidateCode && ev === PairingCeremonyProtocol.EventID.CheckCode && this.guards.get(PairingCeremonyProtocol.GuardID.CodeCorrect)?.() === true: {
                this.state = PairingCeremonyServerPairingState.StorePaired;
                return true;
            }
            case this.state === PairingCeremonyServerPairingState.ValidateCode && ev === PairingCeremonyProtocol.EventID.CheckCode && this.guards.get(PairingCeremonyProtocol.GuardID.CodeWrong)?.() === true: {
                // code_attempts: code_attempts + 1 (set by action)
                this.state = PairingCeremonyServerPairingState.Idle;
                return true;
            }
            case this.state === PairingCeremonyServerPairingState.StorePaired && ev === PairingCeremonyProtocol.EventID.Finalise: {
                this.actions.get(PairingCeremonyProtocol.ActionID.StoreDevice)?.();
                this.deviceSecret = "dev_secret_1";
                // paired_devices: paired_devices \union {"device_1"} (set by action)
                // active_tokens: active_tokens \ {current_token} (set by action)
                // used_tokens: used_tokens \union {current_token} (set by action)
                this.state = PairingCeremonyServerPairingState.PairingComplete;
                return true;
            }
        }
        return false;
    }

    handleMessage(msg: string): boolean {
        switch (true) {
            case this.state === PairingCeremonyServerPairingState.Idle && msg === PairingCeremonyProtocol.MessageType.PairBegin: {
                this.actions.get(PairingCeremonyProtocol.ActionID.GenerateToken)?.();
                this.currentToken = "tok_1";
                // active_tokens: active_tokens \union {"tok_1"} (set by action)
                this.state = PairingCeremonyServerPairingState.GenerateToken;
                return true;
            }
            case this.state === PairingCeremonyServerPairingState.WaitingForClient && msg === PairingCeremonyProtocol.MessageType.PairHello && this.guards.get(PairingCeremonyProtocol.GuardID.TokenValid)?.() === true: {
                this.actions.get(PairingCeremonyProtocol.ActionID.DeriveSecret)?.();
                // received_client_pub: recv_msg.pubkey (set by action)
                this.serverEcdhPub = "server_pub";
                // server_shared_key: DeriveKey("server_pub", recv_msg.pubkey) (set by action)
                // server_code: DeriveCode("server_pub", recv_msg.pubkey) (set by action)
                this.state = PairingCeremonyServerPairingState.DeriveSecret;
                return true;
            }
            case this.state === PairingCeremonyServerPairingState.WaitingForClient && msg === PairingCeremonyProtocol.MessageType.PairHello && this.guards.get(PairingCeremonyProtocol.GuardID.TokenInvalid)?.() === true: {
                this.state = PairingCeremonyServerPairingState.Idle;
                return true;
            }
            case this.state === PairingCeremonyServerPairingState.WaitingForCode && msg === PairingCeremonyProtocol.MessageType.CodeSubmit: {
                // received_code: recv_msg.code (set by action)
                this.state = PairingCeremonyServerPairingState.ValidateCode;
                return true;
            }
        }
        return false;
    }
}

/** PairingCeremonyServerAuthMachine is the generated state machine for server/auth. */
export class PairingCeremonyServerAuthMachine {
    readonly protocol = PairingCeremonyProtocol;
    state: PairingCeremonyServerAuthState;
    receivedDeviceId: string = "none"; // device_id from auth_request
    authNoncesUsed: string = ""; // set of consumed auth nonces
    receivedAuthNonce: string = "none"; // nonce from auth_request
    guards: Map<PairingCeremonyProtocol.GuardID, () => boolean> = new Map();
    actions: Map<PairingCeremonyProtocol.ActionID, () => void> = new Map();

    constructor() {
        this.state = PairingCeremonyServerAuthState.Idle;
    }

    step(ev: PairingCeremonyProtocol.EventID): boolean {
        switch (true) {
            case this.state === PairingCeremonyServerAuthState.Idle && ev === PairingCeremonyProtocol.EventID.CredentialReady: {
                this.state = PairingCeremonyServerAuthState.Paired;
                return true;
            }
            case this.state === PairingCeremonyServerAuthState.AuthCheck && ev === PairingCeremonyProtocol.EventID.Verify && this.guards.get(PairingCeremonyProtocol.GuardID.DeviceKnown)?.() === true: {
                this.actions.get(PairingCeremonyProtocol.ActionID.VerifyDevice)?.();
                // auth_nonces_used: auth_nonces_used \union {received_auth_nonce} (set by action)
                this.state = PairingCeremonyServerAuthState.SessionActive;
                return true;
            }
            case this.state === PairingCeremonyServerAuthState.AuthCheck && ev === PairingCeremonyProtocol.EventID.Verify && this.guards.get(PairingCeremonyProtocol.GuardID.DeviceUnknown)?.() === true: {
                this.state = PairingCeremonyServerAuthState.Idle;
                return true;
            }
            case this.state === PairingCeremonyServerAuthState.SessionActive && ev === PairingCeremonyProtocol.EventID.Disconnect: {
                this.state = PairingCeremonyServerAuthState.Paired;
                return true;
            }
        }
        return false;
    }

    handleMessage(msg: string): boolean {
        switch (true) {
            case this.state === PairingCeremonyServerAuthState.Paired && msg === PairingCeremonyProtocol.MessageType.AuthRequest: {
                // received_device_id: recv_msg.device_id (set by action)
                // received_auth_nonce: recv_msg.nonce (set by action)
                this.state = PairingCeremonyServerAuthState.AuthCheck;
                return true;
            }
        }
        return false;
    }
}

/** PairingCeremonyServerComposite holds all sub-machines for the server actor. */
export class PairingCeremonyServerComposite {
    pairing = new PairingCeremonyServerPairingMachine();
    auth = new PairingCeremonyServerAuthMachine();
    routeGuards: Map<PairingCeremonyProtocol.GuardID, () => boolean> = new Map();

    /** route dispatches inter-machine events according to the routing table. */
    route(from: string, event: PairingCeremonyProtocol.EventID): void {
        switch (true) {
            case from === "pairing" && event === PairingCeremonyProtocol.EventID.Paired: {
                this.auth.step(PairingCeremonyProtocol.EventID.CredentialReady);
                break;
            }
        }
    }
}

/** PairingCeremonyIosPairingMachine is the generated state machine for ios/pairing. */
export class PairingCeremonyIosPairingMachine {
    readonly protocol = PairingCeremonyProtocol;
    state: PairingCeremonyIosPairingState;
    receivedServerPub: string = "none"; // pubkey ios received in pair_hello_ack (may be adversary's)
    clientSharedKey: string = ""; // ECDH key derived by ios (tuple to match DeriveKey output type)
    iosCode: string = ""; // code computed by ios from its view of the pubkeys (tuple to match DeriveCode output type)
    guards: Map<PairingCeremonyProtocol.GuardID, () => boolean> = new Map();
    actions: Map<PairingCeremonyProtocol.ActionID, () => void> = new Map();

    constructor() {
        this.state = PairingCeremonyIosPairingState.Idle;
    }

    step(ev: PairingCeremonyProtocol.EventID): boolean {
        switch (true) {
            case this.state === PairingCeremonyIosPairingState.Idle && ev === PairingCeremonyProtocol.EventID.UserScansQR: {
                this.state = PairingCeremonyIosPairingState.ScanQR;
                return true;
            }
            case this.state === PairingCeremonyIosPairingState.ScanQR && ev === PairingCeremonyProtocol.EventID.QRParsed: {
                this.state = PairingCeremonyIosPairingState.ConnectRelay;
                return true;
            }
            case this.state === PairingCeremonyIosPairingState.ConnectRelay && ev === PairingCeremonyProtocol.EventID.RelayConnected: {
                this.state = PairingCeremonyIosPairingState.GenKeyPair;
                return true;
            }
            case this.state === PairingCeremonyIosPairingState.GenKeyPair && ev === PairingCeremonyProtocol.EventID.KeyPairGenerated: {
                this.actions.get(PairingCeremonyProtocol.ActionID.SendPairHello)?.();
                this.state = PairingCeremonyIosPairingState.WaitAck;
                return true;
            }
            case this.state === PairingCeremonyIosPairingState.ShowCode && ev === PairingCeremonyProtocol.EventID.CodeDisplayed: {
                this.state = PairingCeremonyIosPairingState.WaitPairComplete;
                return true;
            }
        }
        return false;
    }

    handleMessage(msg: string): boolean {
        switch (true) {
            case this.state === PairingCeremonyIosPairingState.WaitAck && msg === PairingCeremonyProtocol.MessageType.PairHelloAck: {
                this.actions.get(PairingCeremonyProtocol.ActionID.DeriveSecret)?.();
                // received_server_pub: recv_msg.pubkey (set by action)
                // client_shared_key: DeriveKey("client_pub", recv_msg.pubkey) (set by action)
                this.state = PairingCeremonyIosPairingState.E2EReady;
                return true;
            }
            case this.state === PairingCeremonyIosPairingState.E2EReady && msg === PairingCeremonyProtocol.MessageType.PairConfirm: {
                // ios_code: DeriveCode(received_server_pub, "client_pub") (set by action)
                this.state = PairingCeremonyIosPairingState.ShowCode;
                return true;
            }
            case this.state === PairingCeremonyIosPairingState.WaitPairComplete && msg === PairingCeremonyProtocol.MessageType.PairComplete: {
                this.actions.get(PairingCeremonyProtocol.ActionID.StoreSecret)?.();
                this.state = PairingCeremonyIosPairingState.PairingComplete;
                return true;
            }
        }
        return false;
    }
}

/** PairingCeremonyIosAuthMachine is the generated state machine for ios/auth. */
export class PairingCeremonyIosAuthMachine {
    readonly protocol = PairingCeremonyProtocol;
    state: PairingCeremonyIosAuthState;
    guards: Map<PairingCeremonyProtocol.GuardID, () => boolean> = new Map();
    actions: Map<PairingCeremonyProtocol.ActionID, () => void> = new Map();

    constructor() {
        this.state = PairingCeremonyIosAuthState.Idle;
    }

    step(ev: PairingCeremonyProtocol.EventID): boolean {
        switch (true) {
            case this.state === PairingCeremonyIosAuthState.Idle && ev === PairingCeremonyProtocol.EventID.CredentialReady: {
                this.state = PairingCeremonyIosAuthState.Paired;
                return true;
            }
            case this.state === PairingCeremonyIosAuthState.Paired && ev === PairingCeremonyProtocol.EventID.AppLaunch: {
                this.state = PairingCeremonyIosAuthState.Reconnect;
                return true;
            }
            case this.state === PairingCeremonyIosAuthState.Reconnect && ev === PairingCeremonyProtocol.EventID.RelayConnected: {
                this.state = PairingCeremonyIosAuthState.SendAuth;
                return true;
            }
            case this.state === PairingCeremonyIosAuthState.SessionActive && ev === PairingCeremonyProtocol.EventID.Disconnect: {
                this.state = PairingCeremonyIosAuthState.Paired;
                return true;
            }
        }
        return false;
    }

    handleMessage(msg: string): boolean {
        switch (true) {
            case this.state === PairingCeremonyIosAuthState.SendAuth && msg === PairingCeremonyProtocol.MessageType.AuthOk: {
                this.state = PairingCeremonyIosAuthState.SessionActive;
                return true;
            }
        }
        return false;
    }
}

/** PairingCeremonyIosComposite holds all sub-machines for the ios actor. */
export class PairingCeremonyIosComposite {
    pairing = new PairingCeremonyIosPairingMachine();
    auth = new PairingCeremonyIosAuthMachine();
    routeGuards: Map<PairingCeremonyProtocol.GuardID, () => boolean> = new Map();

    /** route dispatches inter-machine events according to the routing table. */
    route(from: string, event: PairingCeremonyProtocol.EventID): void {
        switch (true) {
            case from === "pairing" && event === PairingCeremonyProtocol.EventID.Paired: {
                this.auth.step(PairingCeremonyProtocol.EventID.CredentialReady);
                break;
            }
        }
    }
}

/** PairingCeremonyCliMachine is the generated state machine for the cli actor. */
export class PairingCeremonyCliMachine {
    readonly protocol = PairingCeremonyProtocol;
    state: PairingCeremonyCliState;
    guards: Map<PairingCeremonyProtocol.GuardID, () => boolean> = new Map();
    actions: Map<PairingCeremonyProtocol.ActionID, () => void> = new Map();

    constructor() {
        this.state = PairingCeremonyCliState.Idle;
    }

    handleEvent(ev: PairingCeremonyProtocol.EventID): string[] {
        switch (true) {
            case this.state === PairingCeremonyCliState.Idle && ev === PairingCeremonyProtocol.EventID.CliInit: {
                this.state = PairingCeremonyCliState.GetKey;
                return [];
            }
            case this.state === PairingCeremonyCliState.GetKey && ev === PairingCeremonyProtocol.EventID.KeyStored: {
                this.state = PairingCeremonyCliState.BeginPair;
                return [];
            }
            case this.state === PairingCeremonyCliState.BeginPair && ev === PairingCeremonyProtocol.EventID.RecvTokenResponse: {
                this.state = PairingCeremonyCliState.ShowQR;
                return [];
            }
            case this.state === PairingCeremonyCliState.ShowQR && ev === PairingCeremonyProtocol.EventID.RecvWaitingForCode: {
                this.state = PairingCeremonyCliState.PromptCode;
                return [];
            }
            case this.state === PairingCeremonyCliState.PromptCode && ev === PairingCeremonyProtocol.EventID.UserEntersCode: {
                this.state = PairingCeremonyCliState.SubmitCode;
                return [];
            }
            case this.state === PairingCeremonyCliState.SubmitCode && ev === PairingCeremonyProtocol.EventID.RecvPairStatus: {
                this.state = PairingCeremonyCliState.Done;
                return [];
            }
        }
        return [];
    }
}

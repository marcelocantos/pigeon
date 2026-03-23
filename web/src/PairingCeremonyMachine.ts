// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Auto-generated from protocol definition. Do not edit.
// Source of truth: protocol/*.yaml

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

export enum ServerState {
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
}

export class ServerMachine {
    private _state: ServerState = ServerState.Idle;

    get state(): ServerState {
        return this._state;
    }

    /** Process a received message. Returns the new state, or null if rejected. */
    handleMessage(msg: MessageType, guard?: (name: string) => boolean): ServerState | null {
        const check = guard ?? (() => true);
        let newState: ServerState | null = null;
        if (this._state === ServerState.Idle && msg === MessageType.PairBegin) {
            newState = ServerState.GenerateToken;
        } else if (this._state === ServerState.WaitingForClient && msg === MessageType.PairHello && check("token_valid")) {
            newState = ServerState.DeriveSecret;
        } else if (this._state === ServerState.WaitingForClient && msg === MessageType.PairHello && check("token_invalid")) {
            newState = ServerState.Idle;
        } else if (this._state === ServerState.WaitingForCode && msg === MessageType.CodeSubmit) {
            newState = ServerState.ValidateCode;
        } else if (this._state === ServerState.Paired && msg === MessageType.AuthRequest) {
            newState = ServerState.AuthCheck;
        }
        if (newState !== null) this._state = newState;
        return newState;
    }

    /** Attempt an internal transition. Returns the new state, or null if none available. */
    step(guard?: (name: string) => boolean): ServerState | null {
        const check = guard ?? (() => true);
        if (this._state === ServerState.GenerateToken) {
            this._state = ServerState.RegisterRelay;
            return this._state;
        } else if (this._state === ServerState.RegisterRelay) {
            this._state = ServerState.WaitingForClient;
            return this._state;
        } else if (this._state === ServerState.DeriveSecret) {
            this._state = ServerState.SendAck;
            return this._state;
        } else if (this._state === ServerState.SendAck) {
            this._state = ServerState.WaitingForCode;
            return this._state;
        } else if (this._state === ServerState.ValidateCode && check("code_correct")) {
            this._state = ServerState.StorePaired;
            return this._state;
        } else if (this._state === ServerState.ValidateCode && check("code_wrong")) {
            this._state = ServerState.Idle;
            return this._state;
        } else if (this._state === ServerState.StorePaired) {
            this._state = ServerState.Paired;
            return this._state;
        } else if (this._state === ServerState.AuthCheck && check("device_known")) {
            this._state = ServerState.SessionActive;
            return this._state;
        } else if (this._state === ServerState.AuthCheck && check("device_unknown")) {
            this._state = ServerState.Idle;
            return this._state;
        } else if (this._state === ServerState.SessionActive) {
            this._state = ServerState.Paired;
            return this._state;
        }
        return null;
    }
}

export enum IosState {
    Idle = "Idle",
    ScanQR = "ScanQR",
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
}

export class IosMachine {
    private _state: IosState = IosState.Idle;

    get state(): IosState {
        return this._state;
    }

    /** Process a received message. Returns the new state, or null if rejected. */
    handleMessage(msg: MessageType, guard?: (name: string) => boolean): IosState | null {
        const check = guard ?? (() => true);
        let newState: IosState | null = null;
        if (this._state === IosState.WaitAck && msg === MessageType.PairHelloAck) {
            newState = IosState.E2EReady;
        } else if (this._state === IosState.E2EReady && msg === MessageType.PairConfirm) {
            newState = IosState.ShowCode;
        } else if (this._state === IosState.WaitPairComplete && msg === MessageType.PairComplete) {
            newState = IosState.Paired;
        } else if (this._state === IosState.SendAuth && msg === MessageType.AuthOk) {
            newState = IosState.SessionActive;
        }
        if (newState !== null) this._state = newState;
        return newState;
    }

    /** Attempt an internal transition. Returns the new state, or null if none available. */
    step(guard?: (name: string) => boolean): IosState | null {
        const check = guard ?? (() => true);
        if (this._state === IosState.Idle) {
            this._state = IosState.ScanQR;
            return this._state;
        } else if (this._state === IosState.ScanQR) {
            this._state = IosState.ConnectRelay;
            return this._state;
        } else if (this._state === IosState.ConnectRelay) {
            this._state = IosState.GenKeyPair;
            return this._state;
        } else if (this._state === IosState.GenKeyPair) {
            this._state = IosState.WaitAck;
            return this._state;
        } else if (this._state === IosState.ShowCode) {
            this._state = IosState.WaitPairComplete;
            return this._state;
        } else if (this._state === IosState.Paired) {
            this._state = IosState.Reconnect;
            return this._state;
        } else if (this._state === IosState.Reconnect) {
            this._state = IosState.SendAuth;
            return this._state;
        } else if (this._state === IosState.SessionActive) {
            this._state = IosState.Paired;
            return this._state;
        }
        return null;
    }
}

export enum CliState {
    Idle = "Idle",
    GetKey = "GetKey",
    BeginPair = "BeginPair",
    ShowQR = "ShowQR",
    PromptCode = "PromptCode",
    SubmitCode = "SubmitCode",
    Done = "Done",
}

export class CliMachine {
    private _state: CliState = CliState.Idle;

    get state(): CliState {
        return this._state;
    }

    /** Process a received message. Returns the new state, or null if rejected. */
    handleMessage(msg: MessageType, guard?: (name: string) => boolean): CliState | null {
        const check = guard ?? (() => true);
        let newState: CliState | null = null;
        if (this._state === CliState.BeginPair && msg === MessageType.TokenResponse) {
            newState = CliState.ShowQR;
        } else if (this._state === CliState.ShowQR && msg === MessageType.WaitingForCode) {
            newState = CliState.PromptCode;
        } else if (this._state === CliState.SubmitCode && msg === MessageType.PairStatus) {
            newState = CliState.Done;
        }
        if (newState !== null) this._state = newState;
        return newState;
    }

    /** Attempt an internal transition. Returns the new state, or null if none available. */
    step(guard?: (name: string) => boolean): CliState | null {
        const check = guard ?? (() => true);
        if (this._state === CliState.Idle) {
            this._state = CliState.GetKey;
            return this._state;
        } else if (this._state === CliState.GetKey) {
            this._state = CliState.BeginPair;
            return this._state;
        } else if (this._state === CliState.PromptCode) {
            this._state = CliState.SubmitCode;
            return this._state;
        }
        return null;
    }
}

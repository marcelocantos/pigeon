// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Auto-generated from protocol definition. Do not edit.
// Source of truth: protocol/*.yaml

export enum BackendState {
    RelayConnected = "RelayConnected",
    LANOffered = "LANOffered",
    LANActive = "LANActive",
    RelayBackoff = "RelayBackoff",
    LANDegraded = "LANDegraded",
}

export enum ClientState {
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
export namespace PathSwitchProtocol {

    export enum MessageType {
        LanOffer = "lan_offer",
        LanVerify = "lan_verify",
        LanConfirm = "lan_confirm",
        PathPing = "path_ping",
        PathPong = "path_pong",
        RelayResume = "relay_resume",
        RelayResumed = "relay_resumed",
    }

    export enum GuardID {
        ChallengeValid = "challenge_valid",
        ChallengeInvalid = "challenge_invalid",
        LanEnabled = "lan_enabled",
        LanDisabled = "lan_disabled",
        LanServerAvailable = "lan_server_available",
        UnderMaxFailures = "under_max_failures",
        AtMaxFailures = "at_max_failures",
    }

    export enum ActionID {
        ActivateLan = "activate_lan",
        ResetFailures = "reset_failures",
        FallbackToRelay = "fallback_to_relay",
        DialLan = "dial_lan",
        BridgeStreams = "bridge_streams",
        Unbridge = "unbridge",
        RebridgeStreams = "rebridge_streams",
    }

    export enum EventID {
        LanServerReady = "lan_server_ready",
        OfferTimeout = "offer_timeout",
        PingTick = "ping_tick",
        PingTimeout = "ping_timeout",
        BackoffExpired = "backoff_expired",
        LanServerChanged = "lan_server_changed",
        ReadvertiseTick = "readvertise_tick",
        LanDialOk = "lan_dial_ok",
        LanDialFailed = "lan_dial_failed",
        VerifyTimeout = "verify_timeout",
        LanError = "lan_error",
        RelayOk = "relay_ok",
        BackendRegister = "backend_register",
        ClientConnect = "client_connect",
        ClientDisconnect = "client_disconnect",
        BackendDisconnect = "backend_disconnect",
        RecvLanVerify = "recv_lan_verify",
        RecvPathPong = "recv_path_pong",
        RecvLanOffer = "recv_lan_offer",
        RecvLanConfirm = "recv_lan_confirm",
        RecvPathPing = "recv_path_ping",
        RecvRelayResume = "recv_relay_resume",
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
        initial: BackendState.RelayConnected,
        transitions: [
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
        ],
    };

    /** client transition table. */
    export const clientTable: ActorTable = {
        initial: ClientState.RelayConnected,
        transitions: [
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
        ],
    };

    /** relay transition table. */
    export const relayTable: ActorTable = {
        initial: RelayState.Idle,
        transitions: [
            { from: "Idle", to: "BackendRegistered", on: "backend_register", onKind: "internal" },
            { from: "BackendRegistered", to: "Bridged", on: "client_connect", onKind: "internal", action: "bridge_streams" },
            { from: "Bridged", to: "BackendRegistered", on: "client_disconnect", onKind: "internal", action: "unbridge" },
            { from: "Bridged", to: "Bridged", on: "relay_resume", onKind: "recv", action: "rebridge_streams", sends: [{ to: "client", msg: "relay_resumed" }] },
            { from: "BackendRegistered", to: "Idle", on: "backend_disconnect", onKind: "internal" },
        ],
    };

}

/** BackendMachine is the generated state machine for the backend actor. */
export class BackendMachine {
    readonly protocol = PathSwitchProtocol;
    state: BackendState;
    pingFailures: number = 0; // consecutive failed pings on the direct path
    backoffLevel: number = 0; // current exponential backoff level (0 = no backoff)
    activePath: string = "relay"; // "relay" or "lan" — which path carries application traffic
    dispatcherPath: string = "relay"; // which path the datagram dispatcher reads from ("relay", "lan", "none")
    monitorTarget: string = "none"; // which path the health monitor pings ("lan", "none")
    lanSignal: string = "pending"; // LANReady notification state ("pending" = not yet, "ready" = closed/signalled)
    guards: Map<PathSwitchProtocol.GuardID, () => boolean> = new Map();
    actions: Map<PathSwitchProtocol.ActionID, () => void> = new Map();

    constructor() {
        this.state = BackendState.RelayConnected;
    }

    handleEvent(ev: PathSwitchProtocol.EventID): PathSwitchProtocol.CmdID[] {
        switch (true) {
            case this.state === BackendState.RelayConnected && ev === PathSwitchProtocol.EventID.LanServerReady: {
                this.state = BackendState.LANOffered;
                return [];
            }
            case this.state === BackendState.LANOffered && ev === PathSwitchProtocol.EventID.RecvLanVerify && this.guards.get(PathSwitchProtocol.GuardID.ChallengeValid)?.() === true: {
                this.actions.get(PathSwitchProtocol.ActionID.ActivateLan)?.();
                this.pingFailures = 0;
                this.backoffLevel = 0;
                this.activePath = "lan";
                this.monitorTarget = "lan";
                this.dispatcherPath = "lan";
                this.lanSignal = "ready";
                this.state = BackendState.LANActive;
                return [];
            }
            case this.state === BackendState.LANOffered && ev === PathSwitchProtocol.EventID.RecvLanVerify && this.guards.get(PathSwitchProtocol.GuardID.ChallengeInvalid)?.() === true: {
                this.state = BackendState.RelayConnected;
                return [];
            }
            case this.state === BackendState.LANOffered && ev === PathSwitchProtocol.EventID.OfferTimeout: {
                // backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
                this.state = BackendState.RelayBackoff;
                return [];
            }
            case this.state === BackendState.LANActive && ev === PathSwitchProtocol.EventID.PingTick: {
                this.state = BackendState.LANActive;
                return [];
            }
            case this.state === BackendState.LANActive && ev === PathSwitchProtocol.EventID.PingTimeout: {
                this.pingFailures = 1;
                this.state = BackendState.LANDegraded;
                return [];
            }
            case this.state === BackendState.LANDegraded && ev === PathSwitchProtocol.EventID.PingTick: {
                this.state = BackendState.LANDegraded;
                return [];
            }
            case this.state === BackendState.LANDegraded && ev === PathSwitchProtocol.EventID.RecvPathPong: {
                this.actions.get(PathSwitchProtocol.ActionID.ResetFailures)?.();
                this.pingFailures = 0;
                this.state = BackendState.LANActive;
                return [];
            }
            case this.state === BackendState.LANDegraded && ev === PathSwitchProtocol.EventID.PingTimeout && this.guards.get(PathSwitchProtocol.GuardID.UnderMaxFailures)?.() === true: {
                // ping_failures: ping_failures + 1 (set by action)
                this.state = BackendState.LANDegraded;
                return [];
            }
            case this.state === BackendState.LANDegraded && ev === PathSwitchProtocol.EventID.PingTimeout && this.guards.get(PathSwitchProtocol.GuardID.AtMaxFailures)?.() === true: {
                this.actions.get(PathSwitchProtocol.ActionID.FallbackToRelay)?.();
                // backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
                this.activePath = "relay";
                this.monitorTarget = "none";
                this.dispatcherPath = "relay";
                this.lanSignal = "pending";
                this.pingFailures = 0;
                this.state = BackendState.RelayBackoff;
                return [];
            }
            case this.state === BackendState.RelayBackoff && ev === PathSwitchProtocol.EventID.BackoffExpired: {
                this.state = BackendState.LANOffered;
                return [];
            }
            case this.state === BackendState.RelayBackoff && ev === PathSwitchProtocol.EventID.LanServerChanged: {
                this.backoffLevel = 0;
                this.state = BackendState.LANOffered;
                return [];
            }
            case this.state === BackendState.RelayConnected && ev === PathSwitchProtocol.EventID.ReadvertiseTick && this.guards.get(PathSwitchProtocol.GuardID.LanServerAvailable)?.() === true: {
                this.state = BackendState.LANOffered;
                return [];
            }
        }
        return [];
    }
}

/** ClientMachine is the generated state machine for the client actor. */
export class ClientMachine {
    readonly protocol = PathSwitchProtocol;
    state: ClientState;
    activePath: string = "relay"; // "relay" or "lan" — which path carries application traffic
    dispatcherPath: string = "relay"; // which path the datagram dispatcher reads from ("relay", "lan", "none")
    lanSignal: string = "pending"; // LANReady notification state ("pending" = not yet, "ready" = closed/signalled)
    guards: Map<PathSwitchProtocol.GuardID, () => boolean> = new Map();
    actions: Map<PathSwitchProtocol.ActionID, () => void> = new Map();

    constructor() {
        this.state = ClientState.RelayConnected;
    }

    handleEvent(ev: PathSwitchProtocol.EventID): PathSwitchProtocol.CmdID[] {
        switch (true) {
            case this.state === ClientState.RelayConnected && ev === PathSwitchProtocol.EventID.RecvLanOffer && this.guards.get(PathSwitchProtocol.GuardID.LanEnabled)?.() === true: {
                this.actions.get(PathSwitchProtocol.ActionID.DialLan)?.();
                this.state = ClientState.LANConnecting;
                return [];
            }
            case this.state === ClientState.RelayConnected && ev === PathSwitchProtocol.EventID.RecvLanOffer && this.guards.get(PathSwitchProtocol.GuardID.LanDisabled)?.() === true: {
                this.state = ClientState.RelayConnected;
                return [];
            }
            case this.state === ClientState.LANConnecting && ev === PathSwitchProtocol.EventID.LanDialOk: {
                this.state = ClientState.LANVerifying;
                return [];
            }
            case this.state === ClientState.LANConnecting && ev === PathSwitchProtocol.EventID.LanDialFailed: {
                this.state = ClientState.RelayConnected;
                return [];
            }
            case this.state === ClientState.LANVerifying && ev === PathSwitchProtocol.EventID.RecvLanConfirm: {
                this.actions.get(PathSwitchProtocol.ActionID.ActivateLan)?.();
                this.activePath = "lan";
                this.dispatcherPath = "lan";
                this.lanSignal = "ready";
                this.state = ClientState.LANActive;
                return [];
            }
            case this.state === ClientState.LANVerifying && ev === PathSwitchProtocol.EventID.VerifyTimeout: {
                this.dispatcherPath = "relay";
                this.state = ClientState.RelayConnected;
                return [];
            }
            case this.state === ClientState.LANActive && ev === PathSwitchProtocol.EventID.RecvPathPing: {
                this.state = ClientState.LANActive;
                return [];
            }
            case this.state === ClientState.LANActive && ev === PathSwitchProtocol.EventID.LanError: {
                this.actions.get(PathSwitchProtocol.ActionID.FallbackToRelay)?.();
                this.activePath = "relay";
                this.dispatcherPath = "relay";
                this.lanSignal = "pending";
                this.state = ClientState.RelayFallback;
                return [];
            }
            case this.state === ClientState.RelayFallback && ev === PathSwitchProtocol.EventID.RelayOk: {
                this.state = ClientState.RelayConnected;
                return [];
            }
            case this.state === ClientState.LANActive && ev === PathSwitchProtocol.EventID.RecvLanOffer && this.guards.get(PathSwitchProtocol.GuardID.LanEnabled)?.() === true: {
                this.actions.get(PathSwitchProtocol.ActionID.DialLan)?.();
                this.state = ClientState.LANConnecting;
                return [];
            }
        }
        return [];
    }
}

/** RelayMachine is the generated state machine for the relay actor. */
export class RelayMachine {
    readonly protocol = PathSwitchProtocol;
    state: RelayState;
    relayBridge: string = "idle"; // relay bridge state ("active" = bridging, "idle" = backend registered but no client)
    guards: Map<PathSwitchProtocol.GuardID, () => boolean> = new Map();
    actions: Map<PathSwitchProtocol.ActionID, () => void> = new Map();

    constructor() {
        this.state = RelayState.Idle;
    }

    handleEvent(ev: PathSwitchProtocol.EventID): PathSwitchProtocol.CmdID[] {
        switch (true) {
            case this.state === RelayState.Idle && ev === PathSwitchProtocol.EventID.BackendRegister: {
                this.state = RelayState.BackendRegistered;
                return [];
            }
            case this.state === RelayState.BackendRegistered && ev === PathSwitchProtocol.EventID.ClientConnect: {
                this.actions.get(PathSwitchProtocol.ActionID.BridgeStreams)?.();
                this.relayBridge = "active";
                this.state = RelayState.Bridged;
                return [];
            }
            case this.state === RelayState.Bridged && ev === PathSwitchProtocol.EventID.ClientDisconnect: {
                this.actions.get(PathSwitchProtocol.ActionID.Unbridge)?.();
                this.relayBridge = "idle";
                this.state = RelayState.BackendRegistered;
                return [];
            }
            case this.state === RelayState.Bridged && ev === PathSwitchProtocol.EventID.RecvRelayResume: {
                this.actions.get(PathSwitchProtocol.ActionID.RebridgeStreams)?.();
                this.state = RelayState.Bridged;
                return [];
            }
            case this.state === RelayState.BackendRegistered && ev === PathSwitchProtocol.EventID.BackendDisconnect: {
                this.state = RelayState.Idle;
                return [];
            }
        }
        return [];
    }
}

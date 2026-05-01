// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Auto-generated from protocol definition. Do not edit.
// Source of truth: protocol/*.yaml

export enum PairingCeremonyAcceptorState {
    Idle = "Idle",
    GeneratingEphemeral = "GeneratingEphemeral",
    RegisteringRelay = "RegisteringRelay",
    WaitingForHello = "WaitingForHello",
    DerivingCode = "DerivingCode",
    AwaitingUserConfirm = "AwaitingUserConfirm",
    AwaitingPeerConfirm = "AwaitingPeerConfirm",
    Paired = "Paired",
    Aborted = "Aborted",
}

export enum PairingCeremonyInitiatorState {
    Idle = "Idle",
    DecodingToken = "DecodingToken",
    GeneratingEphemeral = "GeneratingEphemeral",
    ConnectingRelay = "ConnectingRelay",
    AwaitingWelcome = "AwaitingWelcome",
    DerivingCode = "DerivingCode",
    AwaitingUserConfirm = "AwaitingUserConfirm",
    AwaitingPeerConfirm = "AwaitingPeerConfirm",
    Paired = "Paired",
    Aborted = "Aborted",
}

/** The protocol transition table and shared type enums. */
export namespace PairingCeremonyProtocol {

    export enum MessageType {
        Hello = "hello",
        Welcome = "welcome",
        ConfirmToInitiator = "confirm_to_initiator",
        ConfirmToAcceptor = "confirm_to_acceptor",
    }

    export enum ActionID {
        GenEphemeral = "gen_ephemeral",
        RegisterRelay = "register_relay",
        EmitToken = "emit_token",
        DeriveCode = "derive_code",
        StoreRecord = "store_record",
        DecodeToken = "decode_token",
        DialRelay = "dial_relay",
    }

    export enum EventID {
        PairBegin = "pair_begin",
        EphemeralReady = "ephemeral_ready",
        RelayRegistered = "relay_registered",
        CodeReady = "code_ready",
        UserConfirm = "user_confirm",
        UserCancel = "user_cancel",
        TokenReceived = "token_received",
        TokenDecoded = "token_decoded",
        RelayConnected = "relay_connected",
        RecvHello = "recv_hello",
        RecvConfirmToAcceptor = "recv_confirm_to_acceptor",
        RecvWelcome = "recv_welcome",
        RecvConfirmToInitiator = "recv_confirm_to_initiator",
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

    /** acceptor transition table. */
    export const acceptorTable: ActorTable = {
        initial: PairingCeremonyAcceptorState.Idle,
        transitions: [
            { from: "Idle", to: "GeneratingEphemeral", on: "pair_begin", onKind: "internal", action: "gen_ephemeral" },
            { from: "GeneratingEphemeral", to: "RegisteringRelay", on: "ephemeral_ready", onKind: "internal", action: "register_relay" },
            { from: "RegisteringRelay", to: "WaitingForHello", on: "relay_registered", onKind: "internal", action: "emit_token" },
            { from: "WaitingForHello", to: "DerivingCode", on: "hello", onKind: "recv", action: "derive_code" },
            { from: "DerivingCode", to: "AwaitingUserConfirm", on: "code_ready", onKind: "internal", sends: [{ to: "initiator", msg: "welcome" }] },
            { from: "AwaitingUserConfirm", to: "AwaitingPeerConfirm", on: "user_confirm", onKind: "internal", sends: [{ to: "initiator", msg: "confirm_to_initiator" }] },
            { from: "AwaitingPeerConfirm", to: "Paired", on: "confirm_to_acceptor", onKind: "recv", action: "store_record" },
            { from: "AwaitingUserConfirm", to: "Aborted", on: "user_cancel", onKind: "internal" },
            { from: "AwaitingPeerConfirm", to: "Aborted", on: "user_cancel", onKind: "internal" },
        ],
    };

    /** initiator transition table. */
    export const initiatorTable: ActorTable = {
        initial: PairingCeremonyInitiatorState.Idle,
        transitions: [
            { from: "Idle", to: "DecodingToken", on: "token_received", onKind: "internal", action: "decode_token" },
            { from: "DecodingToken", to: "GeneratingEphemeral", on: "token_decoded", onKind: "internal", action: "gen_ephemeral" },
            { from: "GeneratingEphemeral", to: "ConnectingRelay", on: "ephemeral_ready", onKind: "internal", action: "dial_relay" },
            { from: "ConnectingRelay", to: "AwaitingWelcome", on: "relay_connected", onKind: "internal", sends: [{ to: "acceptor", msg: "hello" }] },
            { from: "AwaitingWelcome", to: "DerivingCode", on: "welcome", onKind: "recv", action: "derive_code" },
            { from: "DerivingCode", to: "AwaitingUserConfirm", on: "code_ready", onKind: "internal" },
            { from: "AwaitingUserConfirm", to: "AwaitingPeerConfirm", on: "user_confirm", onKind: "internal", sends: [{ to: "acceptor", msg: "confirm_to_acceptor" }] },
            { from: "AwaitingPeerConfirm", to: "Paired", on: "confirm_to_initiator", onKind: "recv", action: "store_record" },
            { from: "AwaitingUserConfirm", to: "Aborted", on: "user_cancel", onKind: "internal" },
            { from: "AwaitingPeerConfirm", to: "Aborted", on: "user_cancel", onKind: "internal" },
        ],
    };

}

/** PairingCeremonyAcceptorMachine is the generated state machine for the acceptor actor. */
export class PairingCeremonyAcceptorMachine {
    readonly protocol = PairingCeremonyProtocol;
    state: PairingCeremonyAcceptorState;
    acceptorEphPub: string = "none"; // acceptor's ephemeral X25519 public key
    acceptorReceivedEphPub: string = "none"; // ephemeral pubkey acceptor saw in hello (may be adversary's)
    acceptorReceivedIdentity: string = "none"; // identity pubkey acceptor saw in hello
    acceptorReceivedInstance: string = "none"; // instance ID acceptor saw in hello
    acceptorCode: string = ""; // confirmation code acceptor derived from its (ephA, ephB) view
    acceptorUserConfirmed: string = "false"; // has the acceptor's local human pressed y?
    acceptorReceivedConfirm: string = "false"; // has the acceptor received initiator's confirm message?
    actions: Map<PairingCeremonyProtocol.ActionID, () => void> = new Map();

    constructor() {
        this.state = PairingCeremonyAcceptorState.Idle;
    }

    handleEvent(ev: PairingCeremonyProtocol.EventID): string[] {
        switch (true) {
            case this.state === PairingCeremonyAcceptorState.Idle && ev === PairingCeremonyProtocol.EventID.PairBegin: {
                this.actions.get(PairingCeremonyProtocol.ActionID.GenEphemeral)?.();
                this.acceptorEphPub = "acceptor_eph";
                this.state = PairingCeremonyAcceptorState.GeneratingEphemeral;
                return [];
            }
            case this.state === PairingCeremonyAcceptorState.GeneratingEphemeral && ev === PairingCeremonyProtocol.EventID.EphemeralReady: {
                this.actions.get(PairingCeremonyProtocol.ActionID.RegisterRelay)?.();
                this.state = PairingCeremonyAcceptorState.RegisteringRelay;
                return [];
            }
            case this.state === PairingCeremonyAcceptorState.RegisteringRelay && ev === PairingCeremonyProtocol.EventID.RelayRegistered: {
                this.actions.get(PairingCeremonyProtocol.ActionID.EmitToken)?.();
                this.state = PairingCeremonyAcceptorState.WaitingForHello;
                return [];
            }
            case this.state === PairingCeremonyAcceptorState.WaitingForHello && ev === PairingCeremonyProtocol.EventID.RecvHello: {
                this.actions.get(PairingCeremonyProtocol.ActionID.DeriveCode)?.();
                // acceptor_received_eph_pub: recv_msg.eph_pub (set by action)
                // acceptor_received_identity: recv_msg.identity_pub (set by action)
                // acceptor_received_instance: recv_msg.instance_id (set by action)
                // acceptor_code: DeriveCode(acceptor_eph_pub, recv_msg.eph_pub) (set by action)
                this.state = PairingCeremonyAcceptorState.DerivingCode;
                return [];
            }
            case this.state === PairingCeremonyAcceptorState.DerivingCode && ev === PairingCeremonyProtocol.EventID.CodeReady: {
                this.state = PairingCeremonyAcceptorState.AwaitingUserConfirm;
                return [];
            }
            case this.state === PairingCeremonyAcceptorState.AwaitingUserConfirm && ev === PairingCeremonyProtocol.EventID.UserConfirm: {
                this.acceptorUserConfirmed = "true";
                this.state = PairingCeremonyAcceptorState.AwaitingPeerConfirm;
                return [];
            }
            case this.state === PairingCeremonyAcceptorState.AwaitingPeerConfirm && ev === PairingCeremonyProtocol.EventID.RecvConfirmToAcceptor: {
                this.actions.get(PairingCeremonyProtocol.ActionID.StoreRecord)?.();
                this.acceptorReceivedConfirm = "true";
                this.state = PairingCeremonyAcceptorState.Paired;
                return [];
            }
            case this.state === PairingCeremonyAcceptorState.AwaitingUserConfirm && ev === PairingCeremonyProtocol.EventID.UserCancel: {
                this.state = PairingCeremonyAcceptorState.Aborted;
                return [];
            }
            case this.state === PairingCeremonyAcceptorState.AwaitingPeerConfirm && ev === PairingCeremonyProtocol.EventID.UserCancel: {
                this.state = PairingCeremonyAcceptorState.Aborted;
                return [];
            }
        }
        return [];
    }
}

/** PairingCeremonyInitiatorMachine is the generated state machine for the initiator actor. */
export class PairingCeremonyInitiatorMachine {
    readonly protocol = PairingCeremonyProtocol;
    state: PairingCeremonyInitiatorState;
    initiatorEphPub: string = "none"; // initiator's ephemeral X25519 public key
    receivedAcceptorEphPub: string = "none"; // acceptor ephemeral pubkey from token (trusted, out-of-band)
    receivedAcceptorIdentity: string = "none"; // acceptor identity pubkey from token
    receivedAcceptorInstance: string = "none"; // acceptor instance ID from token
    initiatorReceivedEphPub: string = "none"; // ephemeral pubkey initiator saw in welcome (may be adversary's)
    initiatorReceivedIdentity: string = "none"; // identity pubkey initiator saw in welcome
    initiatorReceivedInstance: string = "none"; // instance ID initiator saw in welcome
    initiatorCode: string = ""; // confirmation code initiator derived from its (ephA, ephB) view
    initiatorUserConfirmed: string = "false"; // has the initiator's local human pressed y?
    initiatorReceivedConfirm: string = "false"; // has the initiator received acceptor's confirm message?
    actions: Map<PairingCeremonyProtocol.ActionID, () => void> = new Map();

    constructor() {
        this.state = PairingCeremonyInitiatorState.Idle;
    }

    handleEvent(ev: PairingCeremonyProtocol.EventID): string[] {
        switch (true) {
            case this.state === PairingCeremonyInitiatorState.Idle && ev === PairingCeremonyProtocol.EventID.TokenReceived: {
                this.actions.get(PairingCeremonyProtocol.ActionID.DecodeToken)?.();
                this.receivedAcceptorEphPub = "acceptor_eph";
                this.receivedAcceptorIdentity = "acceptor_id";
                this.receivedAcceptorInstance = "acceptor_instance";
                this.state = PairingCeremonyInitiatorState.DecodingToken;
                return [];
            }
            case this.state === PairingCeremonyInitiatorState.DecodingToken && ev === PairingCeremonyProtocol.EventID.TokenDecoded: {
                this.actions.get(PairingCeremonyProtocol.ActionID.GenEphemeral)?.();
                this.initiatorEphPub = "initiator_eph";
                this.state = PairingCeremonyInitiatorState.GeneratingEphemeral;
                return [];
            }
            case this.state === PairingCeremonyInitiatorState.GeneratingEphemeral && ev === PairingCeremonyProtocol.EventID.EphemeralReady: {
                this.actions.get(PairingCeremonyProtocol.ActionID.DialRelay)?.();
                this.state = PairingCeremonyInitiatorState.ConnectingRelay;
                return [];
            }
            case this.state === PairingCeremonyInitiatorState.ConnectingRelay && ev === PairingCeremonyProtocol.EventID.RelayConnected: {
                this.state = PairingCeremonyInitiatorState.AwaitingWelcome;
                return [];
            }
            case this.state === PairingCeremonyInitiatorState.AwaitingWelcome && ev === PairingCeremonyProtocol.EventID.RecvWelcome: {
                this.actions.get(PairingCeremonyProtocol.ActionID.DeriveCode)?.();
                // initiator_received_eph_pub: recv_msg.eph_pub (set by action)
                // initiator_received_identity: recv_msg.identity_pub (set by action)
                // initiator_received_instance: recv_msg.instance_id (set by action)
                // initiator_code: DeriveCode(initiator_eph_pub, recv_msg.eph_pub) (set by action)
                this.state = PairingCeremonyInitiatorState.DerivingCode;
                return [];
            }
            case this.state === PairingCeremonyInitiatorState.DerivingCode && ev === PairingCeremonyProtocol.EventID.CodeReady: {
                this.state = PairingCeremonyInitiatorState.AwaitingUserConfirm;
                return [];
            }
            case this.state === PairingCeremonyInitiatorState.AwaitingUserConfirm && ev === PairingCeremonyProtocol.EventID.UserConfirm: {
                this.initiatorUserConfirmed = "true";
                this.state = PairingCeremonyInitiatorState.AwaitingPeerConfirm;
                return [];
            }
            case this.state === PairingCeremonyInitiatorState.AwaitingPeerConfirm && ev === PairingCeremonyProtocol.EventID.RecvConfirmToInitiator: {
                this.actions.get(PairingCeremonyProtocol.ActionID.StoreRecord)?.();
                this.initiatorReceivedConfirm = "true";
                this.state = PairingCeremonyInitiatorState.Paired;
                return [];
            }
            case this.state === PairingCeremonyInitiatorState.AwaitingUserConfirm && ev === PairingCeremonyProtocol.EventID.UserCancel: {
                this.state = PairingCeremonyInitiatorState.Aborted;
                return [];
            }
            case this.state === PairingCeremonyInitiatorState.AwaitingPeerConfirm && ev === PairingCeremonyProtocol.EventID.UserCancel: {
                this.state = PairingCeremonyInitiatorState.Aborted;
                return [];
            }
        }
        return [];
    }
}

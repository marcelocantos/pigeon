// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Auto-generated from protocol definition. Do not edit.
// Source of truth: protocol/*.yaml

package com.marcelocantos.pigeon.crypto

enum class PairingCeremonyAcceptorState(val value: String) {
    Idle("Idle"),
    GeneratingEphemeral("GeneratingEphemeral"),
    RegisteringRelay("RegisteringRelay"),
    WaitingForHello("WaitingForHello"),
    WaitingForReveal("WaitingForReveal"),
    DerivingCode("DerivingCode"),
    AwaitingUserConfirm("AwaitingUserConfirm"),
    AwaitingPeerConfirm("AwaitingPeerConfirm"),
    Paired("Paired"),
    Aborted("Aborted");
}

enum class PairingCeremonyInitiatorState(val value: String) {
    Idle("Idle"),
    DecodingToken("DecodingToken"),
    GeneratingEphemeral("GeneratingEphemeral"),
    ConnectingRelay("ConnectingRelay"),
    AwaitingWelcome("AwaitingWelcome"),
    Revealing("Revealing"),
    DerivingCode("DerivingCode"),
    AwaitingUserConfirm("AwaitingUserConfirm"),
    AwaitingPeerConfirm("AwaitingPeerConfirm"),
    Paired("Paired"),
    Aborted("Aborted");
}

/** The protocol transition table and shared type enums. */
object PairingCeremonyProtocol {

    enum class MessageType(val value: String) {
        Hello("hello"),
        Welcome("welcome"),
        Reveal("reveal"),
        ConfirmToInitiator("confirm_to_initiator"),
        ConfirmToAcceptor("confirm_to_acceptor");
    }

    enum class ActionID(val value: String) {
        GenEphemeral("gen_ephemeral"),
        RegisterRelay("register_relay"),
        EmitToken("emit_token"),
        StoreCommit("store_commit"),
        VerifyCommitAndDerive("verify_commit_and_derive"),
        StoreRecord("store_record"),
        DecodeToken("decode_token"),
        DialRelay("dial_relay"),
        SendReveal("send_reveal"),
        DeriveCode("derive_code");
    }

    enum class EventID(val value: String) {
        CodeReady("code_ready"),
        CommitFail("commit_fail"),
        EphemeralReady("ephemeral_ready"),
        PairBegin("pair_begin"),
        RecvConfirmToAcceptor("recv_confirm_to_acceptor"),
        RecvConfirmToInitiator("recv_confirm_to_initiator"),
        RecvHello("recv_hello"),
        RecvReveal("recv_reveal"),
        RecvWelcome("recv_welcome"),
        RelayConnected("relay_connected"),
        RelayRegistered("relay_registered"),
        RevealSent("reveal_sent"),
        TokenDecoded("token_decoded"),
        TokenReceived("token_received"),
        UserCancel("user_cancel"),
        UserConfirm("user_confirm");
    }

    /** acceptor transition table. */
    object AcceptorTable {
        val initial = PairingCeremonyAcceptorState.Idle

        data class Transition(
            val from: String,
            val to: String,
            val on: String,
            val onKind: String,
            val guard: String? = null,
            val action: String? = null,
            val sends: List<Pair<String, String>> = emptyList(),
        )

        val transitions = listOf(
            Transition("Idle", "GeneratingEphemeral", "pair_begin", "internal", null, "gen_ephemeral", emptyList()),
            Transition("GeneratingEphemeral", "RegisteringRelay", "ephemeral_ready", "internal", null, "register_relay", emptyList()),
            Transition("RegisteringRelay", "WaitingForHello", "relay_registered", "internal", null, "emit_token", emptyList()),
            Transition("WaitingForHello", "WaitingForReveal", "hello", "recv", null, "store_commit", listOf("initiator" to "welcome")),
            Transition("WaitingForReveal", "DerivingCode", "reveal", "recv", null, "verify_commit_and_derive", emptyList()),
            Transition("DerivingCode", "AwaitingUserConfirm", "code_ready", "internal", null, null, emptyList()),
            Transition("AwaitingUserConfirm", "AwaitingPeerConfirm", "user_confirm", "internal", null, null, listOf("initiator" to "confirm_to_initiator")),
            Transition("AwaitingPeerConfirm", "Paired", "confirm_to_acceptor", "recv", null, "store_record", emptyList()),
            Transition("AwaitingUserConfirm", "Aborted", "user_cancel", "internal", null, null, emptyList()),
            Transition("AwaitingPeerConfirm", "Aborted", "user_cancel", "internal", null, null, emptyList()),
            Transition("WaitingForReveal", "Aborted", "commit_fail", "internal", null, null, emptyList()),
        )
    }

    /** initiator transition table. */
    object InitiatorTable {
        val initial = PairingCeremonyInitiatorState.Idle

        data class Transition(
            val from: String,
            val to: String,
            val on: String,
            val onKind: String,
            val guard: String? = null,
            val action: String? = null,
            val sends: List<Pair<String, String>> = emptyList(),
        )

        val transitions = listOf(
            Transition("Idle", "DecodingToken", "token_received", "internal", null, "decode_token", emptyList()),
            Transition("DecodingToken", "GeneratingEphemeral", "token_decoded", "internal", null, "gen_ephemeral", emptyList()),
            Transition("GeneratingEphemeral", "ConnectingRelay", "ephemeral_ready", "internal", null, "dial_relay", emptyList()),
            Transition("ConnectingRelay", "AwaitingWelcome", "relay_connected", "internal", null, null, listOf("acceptor" to "hello")),
            Transition("AwaitingWelcome", "Revealing", "welcome", "recv", null, "send_reveal", listOf("acceptor" to "reveal")),
            Transition("Revealing", "DerivingCode", "reveal_sent", "internal", null, "derive_code", emptyList()),
            Transition("DerivingCode", "AwaitingUserConfirm", "code_ready", "internal", null, null, emptyList()),
            Transition("AwaitingUserConfirm", "AwaitingPeerConfirm", "user_confirm", "internal", null, null, listOf("acceptor" to "confirm_to_acceptor")),
            Transition("AwaitingPeerConfirm", "Paired", "confirm_to_initiator", "recv", null, "store_record", emptyList()),
            Transition("AwaitingUserConfirm", "Aborted", "user_cancel", "internal", null, null, emptyList()),
            Transition("AwaitingPeerConfirm", "Aborted", "user_cancel", "internal", null, null, emptyList()),
        )
    }

}

/** PairingCeremonyAcceptorMachine is the generated state machine for the acceptor actor. */
class PairingCeremonyAcceptorMachine {
    var state: PairingCeremonyAcceptorState = PairingCeremonyAcceptorState.Idle
        private set
    var acceptorEphPub: String = "none" // acceptor's ephemeral X25519 public key
    var acceptorReceivedCommit: String = "none" // SAS commit from hello (binds peer eph before reveal)
    var acceptorReceivedEphPub: String = "none" // ephemeral pubkey acceptor saw in reveal (may be adversary's)
    var acceptorReceivedIdentity: String = "none" // identity pubkey acceptor saw in hello
    var acceptorReceivedInstance: String = "none" // instance ID acceptor saw in hello
    var acceptorCommitOk: String = "false" // did the reveal open the hello commit? "true" only after CommitMatches
    var acceptorCode: String = "" // confirmation code acceptor derived from its (ephA, ephB) view
    var acceptorUserConfirmed: String = "false" // has the acceptor's local human pressed y?
    var acceptorReceivedConfirm: String = "false" // has the acceptor received initiator's confirm message?
    val actions = mutableMapOf<PairingCeremonyProtocol.ActionID, () -> Unit>()

    /** Handle an event and return the list of commands to execute. */
    fun handleEvent(ev: PairingCeremonyProtocol.EventID): List<String> {
        val cmds: List<String> = when {
            state == PairingCeremonyAcceptorState.Idle && ev == PairingCeremonyProtocol.EventID.PairBegin ->
                run {
                    actions[PairingCeremonyProtocol.ActionID.GenEphemeral]?.invoke()
                    acceptorEphPub = "acceptor_eph"
                    state = PairingCeremonyAcceptorState.GeneratingEphemeral
                    emptyList()
                }
            state == PairingCeremonyAcceptorState.GeneratingEphemeral && ev == PairingCeremonyProtocol.EventID.EphemeralReady ->
                run {
                    actions[PairingCeremonyProtocol.ActionID.RegisterRelay]?.invoke()
                    state = PairingCeremonyAcceptorState.RegisteringRelay
                    emptyList()
                }
            state == PairingCeremonyAcceptorState.RegisteringRelay && ev == PairingCeremonyProtocol.EventID.RelayRegistered ->
                run {
                    actions[PairingCeremonyProtocol.ActionID.EmitToken]?.invoke()
                    state = PairingCeremonyAcceptorState.WaitingForHello
                    emptyList()
                }
            state == PairingCeremonyAcceptorState.WaitingForHello && ev == PairingCeremonyProtocol.EventID.RecvHello ->
                run {
                    actions[PairingCeremonyProtocol.ActionID.StoreCommit]?.invoke()
                    // acceptor_received_commit: recv_msg.commit (set by action)
                    // acceptor_received_identity: recv_msg.identity_pub (set by action)
                    // acceptor_received_instance: recv_msg.instance_id (set by action)
                    state = PairingCeremonyAcceptorState.WaitingForReveal
                    emptyList()
                }
            state == PairingCeremonyAcceptorState.WaitingForReveal && ev == PairingCeremonyProtocol.EventID.RecvReveal ->
                run {
                    actions[PairingCeremonyProtocol.ActionID.VerifyCommitAndDerive]?.invoke()
                    // acceptor_received_eph_pub: recv_msg.eph_pub (set by action)
                    // acceptor_commit_ok: IF CommitMatches(acceptor_received_commit, recv_msg.eph_pub) THEN "true" ELSE "false" (set by action)
                    // acceptor_code: IF CommitMatches(acceptor_received_commit, recv_msg.eph_pub) THEN DeriveCode(acceptor_eph_pub, recv_msg.eph_pub) ELSE <<"none">> (set by action)
                    state = PairingCeremonyAcceptorState.DerivingCode
                    emptyList()
                }
            state == PairingCeremonyAcceptorState.DerivingCode && ev == PairingCeremonyProtocol.EventID.CodeReady ->
                run {
                    state = PairingCeremonyAcceptorState.AwaitingUserConfirm
                    emptyList()
                }
            state == PairingCeremonyAcceptorState.AwaitingUserConfirm && ev == PairingCeremonyProtocol.EventID.UserConfirm ->
                run {
                    acceptorUserConfirmed = "true"
                    state = PairingCeremonyAcceptorState.AwaitingPeerConfirm
                    emptyList()
                }
            state == PairingCeremonyAcceptorState.AwaitingPeerConfirm && ev == PairingCeremonyProtocol.EventID.RecvConfirmToAcceptor ->
                run {
                    actions[PairingCeremonyProtocol.ActionID.StoreRecord]?.invoke()
                    acceptorReceivedConfirm = "true"
                    state = PairingCeremonyAcceptorState.Paired
                    emptyList()
                }
            state == PairingCeremonyAcceptorState.AwaitingUserConfirm && ev == PairingCeremonyProtocol.EventID.UserCancel ->
                run {
                    state = PairingCeremonyAcceptorState.Aborted
                    emptyList()
                }
            state == PairingCeremonyAcceptorState.AwaitingPeerConfirm && ev == PairingCeremonyProtocol.EventID.UserCancel ->
                run {
                    state = PairingCeremonyAcceptorState.Aborted
                    emptyList()
                }
            state == PairingCeremonyAcceptorState.WaitingForReveal && ev == PairingCeremonyProtocol.EventID.CommitFail ->
                run {
                    state = PairingCeremonyAcceptorState.Aborted
                    emptyList()
                }
            else -> emptyList()
        }
        return cmds
    }
}

/** PairingCeremonyInitiatorMachine is the generated state machine for the initiator actor. */
class PairingCeremonyInitiatorMachine {
    var state: PairingCeremonyInitiatorState = PairingCeremonyInitiatorState.Idle
        private set
    var initiatorEphPub: String = "none" // initiator's ephemeral X25519 public key
    var initiatorCommit: String = "none" // SHA256("pigeon-sas-commit"||eph||blind) for initiator_eph_pub
    var receivedAcceptorEphPub: String = "none" // acceptor ephemeral pubkey from token (trusted, out-of-band)
    var receivedAcceptorIdentity: String = "none" // acceptor identity pubkey from token
    var receivedAcceptorInstance: String = "none" // acceptor instance ID from token
    var initiatorReceivedEphPub: String = "none" // ephemeral pubkey initiator saw in welcome (may be adversary's)
    var initiatorReceivedIdentity: String = "none" // identity pubkey initiator saw in welcome
    var initiatorReceivedInstance: String = "none" // instance ID initiator saw in welcome
    var initiatorCode: String = "" // confirmation code initiator derived from its (ephA, ephB) view
    var initiatorUserConfirmed: String = "false" // has the initiator's local human pressed y?
    var initiatorReceivedConfirm: String = "false" // has the initiator received acceptor's confirm message?
    val actions = mutableMapOf<PairingCeremonyProtocol.ActionID, () -> Unit>()

    /** Handle an event and return the list of commands to execute. */
    fun handleEvent(ev: PairingCeremonyProtocol.EventID): List<String> {
        val cmds: List<String> = when {
            state == PairingCeremonyInitiatorState.Idle && ev == PairingCeremonyProtocol.EventID.TokenReceived ->
                run {
                    actions[PairingCeremonyProtocol.ActionID.DecodeToken]?.invoke()
                    receivedAcceptorEphPub = "acceptor_eph"
                    receivedAcceptorIdentity = "acceptor_id"
                    receivedAcceptorInstance = "acceptor_instance"
                    state = PairingCeremonyInitiatorState.DecodingToken
                    emptyList()
                }
            state == PairingCeremonyInitiatorState.DecodingToken && ev == PairingCeremonyProtocol.EventID.TokenDecoded ->
                run {
                    actions[PairingCeremonyProtocol.ActionID.GenEphemeral]?.invoke()
                    initiatorEphPub = "initiator_eph"
                    initiatorCommit = "initiator_commit"
                    state = PairingCeremonyInitiatorState.GeneratingEphemeral
                    emptyList()
                }
            state == PairingCeremonyInitiatorState.GeneratingEphemeral && ev == PairingCeremonyProtocol.EventID.EphemeralReady ->
                run {
                    actions[PairingCeremonyProtocol.ActionID.DialRelay]?.invoke()
                    state = PairingCeremonyInitiatorState.ConnectingRelay
                    emptyList()
                }
            state == PairingCeremonyInitiatorState.ConnectingRelay && ev == PairingCeremonyProtocol.EventID.RelayConnected ->
                run {
                    state = PairingCeremonyInitiatorState.AwaitingWelcome
                    emptyList()
                }
            state == PairingCeremonyInitiatorState.AwaitingWelcome && ev == PairingCeremonyProtocol.EventID.RecvWelcome ->
                run {
                    actions[PairingCeremonyProtocol.ActionID.SendReveal]?.invoke()
                    // initiator_received_eph_pub: recv_msg.eph_pub (set by action)
                    // initiator_received_identity: recv_msg.identity_pub (set by action)
                    // initiator_received_instance: recv_msg.instance_id (set by action)
                    state = PairingCeremonyInitiatorState.Revealing
                    emptyList()
                }
            state == PairingCeremonyInitiatorState.Revealing && ev == PairingCeremonyProtocol.EventID.RevealSent ->
                run {
                    actions[PairingCeremonyProtocol.ActionID.DeriveCode]?.invoke()
                    // initiator_code: DeriveCode(initiator_eph_pub, initiator_received_eph_pub) (set by action)
                    state = PairingCeremonyInitiatorState.DerivingCode
                    emptyList()
                }
            state == PairingCeremonyInitiatorState.DerivingCode && ev == PairingCeremonyProtocol.EventID.CodeReady ->
                run {
                    state = PairingCeremonyInitiatorState.AwaitingUserConfirm
                    emptyList()
                }
            state == PairingCeremonyInitiatorState.AwaitingUserConfirm && ev == PairingCeremonyProtocol.EventID.UserConfirm ->
                run {
                    initiatorUserConfirmed = "true"
                    state = PairingCeremonyInitiatorState.AwaitingPeerConfirm
                    emptyList()
                }
            state == PairingCeremonyInitiatorState.AwaitingPeerConfirm && ev == PairingCeremonyProtocol.EventID.RecvConfirmToInitiator ->
                run {
                    actions[PairingCeremonyProtocol.ActionID.StoreRecord]?.invoke()
                    initiatorReceivedConfirm = "true"
                    state = PairingCeremonyInitiatorState.Paired
                    emptyList()
                }
            state == PairingCeremonyInitiatorState.AwaitingUserConfirm && ev == PairingCeremonyProtocol.EventID.UserCancel ->
                run {
                    state = PairingCeremonyInitiatorState.Aborted
                    emptyList()
                }
            state == PairingCeremonyInitiatorState.AwaitingPeerConfirm && ev == PairingCeremonyProtocol.EventID.UserCancel ->
                run {
                    state = PairingCeremonyInitiatorState.Aborted
                    emptyList()
                }
            else -> emptyList()
        }
        return cmds
    }
}


// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package com.marcelocantos.pigeon.crypto

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFailsWith

/**
 * Unit tests for the regenerated pairing-ceremony state machines.
 * These exercise the spec produced by `cmd/protogen protocol/pairing.yaml`
 * — Go, Swift, Kotlin, C, TS, and TLA+ outputs all share the same
 * structure, and these tests pin the Kotlin side to the spec.
 */
class PairingCeremonyMachineTest {

    // ---------- acceptor ----------

    @Test
    fun acceptorInitialStateIsIdle() {
        val m = PairingCeremonyAcceptorMachine()
        assertEquals(PairingCeremonyAcceptorState.Idle, m.state)
    }

    @Test
    fun acceptorHappyPath() {
        val m = PairingCeremonyAcceptorMachine()
        val fired = mutableListOf<PairingCeremonyProtocol.ActionID>()
        for (id in listOf(
            PairingCeremonyProtocol.ActionID.GenEphemeral,
            PairingCeremonyProtocol.ActionID.RegisterRelay,
            PairingCeremonyProtocol.ActionID.EmitToken,
            PairingCeremonyProtocol.ActionID.StoreCommit,
            PairingCeremonyProtocol.ActionID.VerifyCommitAndDerive,
            PairingCeremonyProtocol.ActionID.StoreRecord,
        )) {
            m.actions[id] = { fired.add(id) }
        }

        m.handleEvent(PairingCeremonyProtocol.EventID.PairBegin)
        assertEquals(PairingCeremonyAcceptorState.GeneratingEphemeral, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.EphemeralReady)
        assertEquals(PairingCeremonyAcceptorState.RegisteringRelay, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.RelayRegistered)
        assertEquals(PairingCeremonyAcceptorState.WaitingForHello, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.RecvHello)
        assertEquals(PairingCeremonyAcceptorState.WaitingForReveal, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.RecvReveal)
        assertEquals(PairingCeremonyAcceptorState.DerivingCode, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.CodeReady)
        assertEquals(PairingCeremonyAcceptorState.AwaitingUserConfirm, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.UserConfirm)
        assertEquals(PairingCeremonyAcceptorState.AwaitingPeerConfirm, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.RecvConfirmToAcceptor)
        assertEquals(PairingCeremonyAcceptorState.Paired, m.state)

        assertEquals(
            listOf(
                PairingCeremonyProtocol.ActionID.GenEphemeral,
                PairingCeremonyProtocol.ActionID.RegisterRelay,
                PairingCeremonyProtocol.ActionID.EmitToken,
                PairingCeremonyProtocol.ActionID.StoreCommit,
                PairingCeremonyProtocol.ActionID.VerifyCommitAndDerive,
                PairingCeremonyProtocol.ActionID.StoreRecord,
            ),
            fired,
        )
    }

    @Test
    fun acceptorUserCancelFromAwaitingUserConfirm() {
        val m = PairingCeremonyAcceptorMachine()
        m.actions[PairingCeremonyProtocol.ActionID.GenEphemeral] = {}
        m.actions[PairingCeremonyProtocol.ActionID.RegisterRelay] = {}
        m.actions[PairingCeremonyProtocol.ActionID.EmitToken] = {}
        m.actions[PairingCeremonyProtocol.ActionID.StoreCommit] = {}
        m.actions[PairingCeremonyProtocol.ActionID.VerifyCommitAndDerive] = {}

        m.handleEvent(PairingCeremonyProtocol.EventID.PairBegin)
        m.handleEvent(PairingCeremonyProtocol.EventID.EphemeralReady)
        m.handleEvent(PairingCeremonyProtocol.EventID.RelayRegistered)
        m.handleEvent(PairingCeremonyProtocol.EventID.RecvHello)
        m.handleEvent(PairingCeremonyProtocol.EventID.RecvReveal)
        m.handleEvent(PairingCeremonyProtocol.EventID.CodeReady)
        assertEquals(PairingCeremonyAcceptorState.AwaitingUserConfirm, m.state)

        m.handleEvent(PairingCeremonyProtocol.EventID.UserCancel)
        assertEquals(PairingCeremonyAcceptorState.Aborted, m.state)
    }

    @Test
    fun acceptorUserCancelFromAwaitingPeerConfirm() {
        val m = PairingCeremonyAcceptorMachine()
        m.actions[PairingCeremonyProtocol.ActionID.GenEphemeral] = {}
        m.actions[PairingCeremonyProtocol.ActionID.RegisterRelay] = {}
        m.actions[PairingCeremonyProtocol.ActionID.EmitToken] = {}
        m.actions[PairingCeremonyProtocol.ActionID.StoreCommit] = {}
        m.actions[PairingCeremonyProtocol.ActionID.VerifyCommitAndDerive] = {}

        m.handleEvent(PairingCeremonyProtocol.EventID.PairBegin)
        m.handleEvent(PairingCeremonyProtocol.EventID.EphemeralReady)
        m.handleEvent(PairingCeremonyProtocol.EventID.RelayRegistered)
        m.handleEvent(PairingCeremonyProtocol.EventID.RecvHello)
        m.handleEvent(PairingCeremonyProtocol.EventID.RecvReveal)
        m.handleEvent(PairingCeremonyProtocol.EventID.CodeReady)
        m.handleEvent(PairingCeremonyProtocol.EventID.UserConfirm)
        assertEquals(PairingCeremonyAcceptorState.AwaitingPeerConfirm, m.state)

        m.handleEvent(PairingCeremonyProtocol.EventID.UserCancel)
        assertEquals(PairingCeremonyAcceptorState.Aborted, m.state)
    }

    @Test
    fun acceptorCommitFailAborts() {
        val m = PairingCeremonyAcceptorMachine()
        m.actions[PairingCeremonyProtocol.ActionID.GenEphemeral] = {}
        m.actions[PairingCeremonyProtocol.ActionID.RegisterRelay] = {}
        m.actions[PairingCeremonyProtocol.ActionID.EmitToken] = {}
        m.actions[PairingCeremonyProtocol.ActionID.StoreCommit] = {}

        m.handleEvent(PairingCeremonyProtocol.EventID.PairBegin)
        m.handleEvent(PairingCeremonyProtocol.EventID.EphemeralReady)
        m.handleEvent(PairingCeremonyProtocol.EventID.RelayRegistered)
        m.handleEvent(PairingCeremonyProtocol.EventID.RecvHello)
        assertEquals(PairingCeremonyAcceptorState.WaitingForReveal, m.state)

        m.handleEvent(PairingCeremonyProtocol.EventID.CommitFail)
        assertEquals(PairingCeremonyAcceptorState.Aborted, m.state)
    }

    @Test
    fun acceptorActionThrowingPropagates() {
        val m = PairingCeremonyAcceptorMachine()
        class Boom : RuntimeException()
        m.actions[PairingCeremonyProtocol.ActionID.GenEphemeral] = { throw Boom() }
        assertFailsWith<Boom> {
            m.handleEvent(PairingCeremonyProtocol.EventID.PairBegin)
        }
    }

    // ---------- initiator ----------

    @Test
    fun initiatorInitialStateIsIdle() {
        val m = PairingCeremonyInitiatorMachine()
        assertEquals(PairingCeremonyInitiatorState.Idle, m.state)
    }

    @Test
    fun initiatorHappyPath() {
        val m = PairingCeremonyInitiatorMachine()
        val fired = mutableListOf<PairingCeremonyProtocol.ActionID>()
        for (id in listOf(
            PairingCeremonyProtocol.ActionID.DecodeToken,
            PairingCeremonyProtocol.ActionID.GenEphemeral,
            PairingCeremonyProtocol.ActionID.DialRelay,
            PairingCeremonyProtocol.ActionID.SendReveal,
            PairingCeremonyProtocol.ActionID.DeriveCode,
            PairingCeremonyProtocol.ActionID.StoreRecord,
        )) {
            m.actions[id] = { fired.add(id) }
        }

        m.handleEvent(PairingCeremonyProtocol.EventID.TokenReceived)
        assertEquals(PairingCeremonyInitiatorState.DecodingToken, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.TokenDecoded)
        assertEquals(PairingCeremonyInitiatorState.GeneratingEphemeral, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.EphemeralReady)
        assertEquals(PairingCeremonyInitiatorState.ConnectingRelay, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.RelayConnected)
        assertEquals(PairingCeremonyInitiatorState.AwaitingWelcome, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.RecvWelcome)
        assertEquals(PairingCeremonyInitiatorState.Revealing, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.RevealSent)
        assertEquals(PairingCeremonyInitiatorState.DerivingCode, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.CodeReady)
        assertEquals(PairingCeremonyInitiatorState.AwaitingUserConfirm, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.UserConfirm)
        assertEquals(PairingCeremonyInitiatorState.AwaitingPeerConfirm, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.RecvConfirmToInitiator)
        assertEquals(PairingCeremonyInitiatorState.Paired, m.state)

        assertEquals(
            listOf(
                PairingCeremonyProtocol.ActionID.DecodeToken,
                PairingCeremonyProtocol.ActionID.GenEphemeral,
                PairingCeremonyProtocol.ActionID.DialRelay,
                PairingCeremonyProtocol.ActionID.SendReveal,
                PairingCeremonyProtocol.ActionID.DeriveCode,
                PairingCeremonyProtocol.ActionID.StoreRecord,
            ),
            fired,
        )
    }

    @Test
    fun initiatorUserCancel() {
        val m = PairingCeremonyInitiatorMachine()
        m.actions[PairingCeremonyProtocol.ActionID.DecodeToken] = {}
        m.actions[PairingCeremonyProtocol.ActionID.GenEphemeral] = {}
        m.actions[PairingCeremonyProtocol.ActionID.DialRelay] = {}
        m.actions[PairingCeremonyProtocol.ActionID.SendReveal] = {}
        m.actions[PairingCeremonyProtocol.ActionID.DeriveCode] = {}

        m.handleEvent(PairingCeremonyProtocol.EventID.TokenReceived)
        m.handleEvent(PairingCeremonyProtocol.EventID.TokenDecoded)
        m.handleEvent(PairingCeremonyProtocol.EventID.EphemeralReady)
        m.handleEvent(PairingCeremonyProtocol.EventID.RelayConnected)
        m.handleEvent(PairingCeremonyProtocol.EventID.RecvWelcome)
        m.handleEvent(PairingCeremonyProtocol.EventID.RevealSent)
        m.handleEvent(PairingCeremonyProtocol.EventID.CodeReady)
        assertEquals(PairingCeremonyInitiatorState.AwaitingUserConfirm, m.state)

        m.handleEvent(PairingCeremonyProtocol.EventID.UserCancel)
        assertEquals(PairingCeremonyInitiatorState.Aborted, m.state)
    }

    // ---------- protocol surface ----------

    @Test
    fun messageTypeWireValuesMatchYAML() {
        assertEquals("hello", PairingCeremonyProtocol.MessageType.Hello.value)
        assertEquals("welcome", PairingCeremonyProtocol.MessageType.Welcome.value)
        assertEquals("reveal", PairingCeremonyProtocol.MessageType.Reveal.value)
        assertEquals("confirm_to_initiator", PairingCeremonyProtocol.MessageType.ConfirmToInitiator.value)
        assertEquals("confirm_to_acceptor", PairingCeremonyProtocol.MessageType.ConfirmToAcceptor.value)
    }

    @Test
    fun actionIDWireValuesMatchYAML() {
        assertEquals("gen_ephemeral", PairingCeremonyProtocol.ActionID.GenEphemeral.value)
        assertEquals("register_relay", PairingCeremonyProtocol.ActionID.RegisterRelay.value)
        assertEquals("emit_token", PairingCeremonyProtocol.ActionID.EmitToken.value)
        assertEquals("store_commit", PairingCeremonyProtocol.ActionID.StoreCommit.value)
        assertEquals("verify_commit_and_derive", PairingCeremonyProtocol.ActionID.VerifyCommitAndDerive.value)
        assertEquals("store_record", PairingCeremonyProtocol.ActionID.StoreRecord.value)
        assertEquals("decode_token", PairingCeremonyProtocol.ActionID.DecodeToken.value)
        assertEquals("dial_relay", PairingCeremonyProtocol.ActionID.DialRelay.value)
        assertEquals("send_reveal", PairingCeremonyProtocol.ActionID.SendReveal.value)
        assertEquals("derive_code", PairingCeremonyProtocol.ActionID.DeriveCode.value)
    }
}

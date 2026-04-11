// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package com.marcelocantos.pigeon.crypto

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertTrue

class PairingCeremonyMachineTest {

    // MARK: - Server Machine

    @Test
    fun `server starts in Idle`() {
        val m = PairingCeremonyServerMachine()
        assertEquals(PairingCeremonyServerState.Idle, m.state)
    }

    @Test
    fun `server Idle to GenerateToken on recv pair_begin`() {
        val m = PairingCeremonyServerMachine()
        var actionCalled = false
        m.actions[PairingCeremonyProtocol.ActionID.GenerateToken] = { actionCalled = true }

        m.handleEvent(PairingCeremonyProtocol.EventID.RecvPairBegin)
        assertEquals(PairingCeremonyServerState.GenerateToken, m.state)
        assertTrue(actionCalled)
        assertEquals("tok_1", m.currentToken)
    }

    @Test
    fun `server GenerateToken to RegisterRelay on token created`() {
        val m = serverAtState(PairingCeremonyServerState.GenerateToken)
        var actionCalled = false
        m.actions[PairingCeremonyProtocol.ActionID.RegisterRelay] = { actionCalled = true }

        m.handleEvent(PairingCeremonyProtocol.EventID.TokenCreated)
        assertEquals(PairingCeremonyServerState.RegisterRelay, m.state)
        assertTrue(actionCalled)
    }

    @Test
    fun `server RegisterRelay to WaitingForClient on relay registered`() {
        val m = serverAtState(PairingCeremonyServerState.RegisterRelay)
        m.handleEvent(PairingCeremonyProtocol.EventID.RelayRegistered)
        assertEquals(PairingCeremonyServerState.WaitingForClient, m.state)
    }

    @Test
    fun `server token_valid guard allows transition to DeriveSecret`() {
        val m = serverAtState(PairingCeremonyServerState.WaitingForClient)
        m.guards[PairingCeremonyProtocol.GuardID.TokenValid] = { true }
        m.guards[PairingCeremonyProtocol.GuardID.TokenInvalid] = { false }
        var actionCalled = false
        m.actions[PairingCeremonyProtocol.ActionID.DeriveSecret] = { actionCalled = true }

        m.handleEvent(PairingCeremonyProtocol.EventID.RecvPairHello)
        assertEquals(PairingCeremonyServerState.DeriveSecret, m.state)
        assertTrue(actionCalled)
        assertEquals("server_pub", m.serverEcdhPub)
    }

    @Test
    fun `server token_invalid guard resets to Idle`() {
        val m = serverAtState(PairingCeremonyServerState.WaitingForClient)
        m.guards[PairingCeremonyProtocol.GuardID.TokenValid] = { false }
        m.guards[PairingCeremonyProtocol.GuardID.TokenInvalid] = { true }

        m.handleEvent(PairingCeremonyProtocol.EventID.RecvPairHello)
        assertEquals(PairingCeremonyServerState.Idle, m.state)
    }

    @Test
    fun `server code_correct guard transitions to StorePaired`() {
        val m = serverAtState(PairingCeremonyServerState.ValidateCode)
        m.guards[PairingCeremonyProtocol.GuardID.CodeCorrect] = { true }
        m.guards[PairingCeremonyProtocol.GuardID.CodeWrong] = { false }

        m.handleEvent(PairingCeremonyProtocol.EventID.CheckCode)
        assertEquals(PairingCeremonyServerState.StorePaired, m.state)
    }

    @Test
    fun `server code_wrong guard resets to Idle`() {
        val m = serverAtState(PairingCeremonyServerState.ValidateCode)
        m.guards[PairingCeremonyProtocol.GuardID.CodeCorrect] = { false }
        m.guards[PairingCeremonyProtocol.GuardID.CodeWrong] = { true }

        m.handleEvent(PairingCeremonyProtocol.EventID.CheckCode)
        assertEquals(PairingCeremonyServerState.Idle, m.state)
    }

    @Test
    fun `server finalise stores device and transitions to Paired`() {
        val m = serverAtState(PairingCeremonyServerState.StorePaired)
        var actionCalled = false
        m.actions[PairingCeremonyProtocol.ActionID.StoreDevice] = { actionCalled = true }

        m.handleEvent(PairingCeremonyProtocol.EventID.Finalise)
        assertEquals(PairingCeremonyServerState.Paired, m.state)
        assertTrue(actionCalled)
        assertEquals("dev_secret_1", m.deviceSecret)
    }

    @Test
    fun `server device_known guard transitions to SessionActive`() {
        val m = serverAtState(PairingCeremonyServerState.AuthCheck)
        m.guards[PairingCeremonyProtocol.GuardID.DeviceKnown] = { true }
        m.guards[PairingCeremonyProtocol.GuardID.DeviceUnknown] = { false }
        var actionCalled = false
        m.actions[PairingCeremonyProtocol.ActionID.VerifyDevice] = { actionCalled = true }

        m.handleEvent(PairingCeremonyProtocol.EventID.Verify)
        assertEquals(PairingCeremonyServerState.SessionActive, m.state)
        assertTrue(actionCalled)
    }

    @Test
    fun `server device_unknown guard resets to Idle`() {
        val m = serverAtState(PairingCeremonyServerState.AuthCheck)
        m.guards[PairingCeremonyProtocol.GuardID.DeviceKnown] = { false }
        m.guards[PairingCeremonyProtocol.GuardID.DeviceUnknown] = { true }

        m.handleEvent(PairingCeremonyProtocol.EventID.Verify)
        assertEquals(PairingCeremonyServerState.Idle, m.state)
    }

    @Test
    fun `server disconnect returns to Paired`() {
        val m = serverAtState(PairingCeremonyServerState.SessionActive)
        m.handleEvent(PairingCeremonyProtocol.EventID.Disconnect)
        assertEquals(PairingCeremonyServerState.Paired, m.state)
    }

    @Test
    fun `server invalid event does not change state`() {
        val m = PairingCeremonyServerMachine()
        m.handleEvent(PairingCeremonyProtocol.EventID.Disconnect) // invalid from Idle
        assertEquals(PairingCeremonyServerState.Idle, m.state)
    }

    @Test
    fun `server full pairing flow`() {
        val m = PairingCeremonyServerMachine()
        m.actions[PairingCeremonyProtocol.ActionID.GenerateToken] = {}
        m.actions[PairingCeremonyProtocol.ActionID.RegisterRelay] = {}
        m.actions[PairingCeremonyProtocol.ActionID.DeriveSecret] = {}
        m.actions[PairingCeremonyProtocol.ActionID.StoreDevice] = {}
        m.guards[PairingCeremonyProtocol.GuardID.TokenValid] = { true }
        m.guards[PairingCeremonyProtocol.GuardID.TokenInvalid] = { false }
        m.guards[PairingCeremonyProtocol.GuardID.CodeCorrect] = { true }
        m.guards[PairingCeremonyProtocol.GuardID.CodeWrong] = { false }

        m.handleEvent(PairingCeremonyProtocol.EventID.RecvPairBegin)
        assertEquals(PairingCeremonyServerState.GenerateToken, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.TokenCreated)
        assertEquals(PairingCeremonyServerState.RegisterRelay, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.RelayRegistered)
        assertEquals(PairingCeremonyServerState.WaitingForClient, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.RecvPairHello)
        assertEquals(PairingCeremonyServerState.DeriveSecret, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.ECDHComplete)
        assertEquals(PairingCeremonyServerState.SendAck, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.SignalCodeDisplay)
        assertEquals(PairingCeremonyServerState.WaitingForCode, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.RecvCodeSubmit)
        assertEquals(PairingCeremonyServerState.ValidateCode, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.CheckCode)
        assertEquals(PairingCeremonyServerState.StorePaired, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.Finalise)
        assertEquals(PairingCeremonyServerState.Paired, m.state)
    }

    // MARK: - iOS Machine

    @Test
    fun `ios starts in Idle`() {
        val m = PairingCeremonyIosMachine()
        assertEquals(PairingCeremonyIosState.Idle, m.state)
    }

    @Test
    fun `ios full pairing flow`() {
        val m = PairingCeremonyIosMachine()
        m.actions[PairingCeremonyProtocol.ActionID.SendPairHello] = {}
        m.actions[PairingCeremonyProtocol.ActionID.DeriveSecret] = {}
        m.actions[PairingCeremonyProtocol.ActionID.StoreSecret] = {}

        m.handleEvent(PairingCeremonyProtocol.EventID.UserScansQR)
        assertEquals(PairingCeremonyIosState.ScanQR, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.QRParsed)
        assertEquals(PairingCeremonyIosState.ConnectRelay, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.RelayConnected)
        assertEquals(PairingCeremonyIosState.GenKeyPair, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.KeyPairGenerated)
        assertEquals(PairingCeremonyIosState.WaitAck, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.RecvPairHelloAck)
        assertEquals(PairingCeremonyIosState.E2EReady, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.RecvPairConfirm)
        assertEquals(PairingCeremonyIosState.ShowCode, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.CodeDisplayed)
        assertEquals(PairingCeremonyIosState.WaitPairComplete, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.RecvPairComplete)
        assertEquals(PairingCeremonyIosState.Paired, m.state)
    }

    @Test
    fun `ios key pair generated calls sendPairHello action`() {
        val m = PairingCeremonyIosMachine()
        m.handleEvent(PairingCeremonyProtocol.EventID.UserScansQR)
        m.handleEvent(PairingCeremonyProtocol.EventID.QRParsed)
        m.handleEvent(PairingCeremonyProtocol.EventID.RelayConnected)

        var actionCalled = false
        m.actions[PairingCeremonyProtocol.ActionID.SendPairHello] = { actionCalled = true }

        m.handleEvent(PairingCeremonyProtocol.EventID.KeyPairGenerated)
        assertTrue(actionCalled)
        assertEquals(PairingCeremonyIosState.WaitAck, m.state)
    }

    @Test
    fun `ios reconnect and auth flow`() {
        val m = PairingCeremonyIosMachine()
        m.actions[PairingCeremonyProtocol.ActionID.SendPairHello] = {}
        m.actions[PairingCeremonyProtocol.ActionID.DeriveSecret] = {}
        m.actions[PairingCeremonyProtocol.ActionID.StoreSecret] = {}

        // Walk to Paired
        for (ev in listOf(
            PairingCeremonyProtocol.EventID.UserScansQR,
            PairingCeremonyProtocol.EventID.QRParsed,
            PairingCeremonyProtocol.EventID.RelayConnected,
            PairingCeremonyProtocol.EventID.KeyPairGenerated,
            PairingCeremonyProtocol.EventID.RecvPairHelloAck,
            PairingCeremonyProtocol.EventID.RecvPairConfirm,
            PairingCeremonyProtocol.EventID.CodeDisplayed,
            PairingCeremonyProtocol.EventID.RecvPairComplete,
        )) {
            m.handleEvent(ev)
        }
        assertEquals(PairingCeremonyIosState.Paired, m.state)

        m.handleEvent(PairingCeremonyProtocol.EventID.AppLaunch)
        assertEquals(PairingCeremonyIosState.Reconnect, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.RelayConnected)
        assertEquals(PairingCeremonyIosState.SendAuth, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.RecvAuthOk)
        assertEquals(PairingCeremonyIosState.SessionActive, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.Disconnect)
        assertEquals(PairingCeremonyIosState.Paired, m.state)
    }

    @Test
    fun `ios invalid event does not change state`() {
        val m = PairingCeremonyIosMachine()
        m.handleEvent(PairingCeremonyProtocol.EventID.Disconnect) // invalid from Idle
        assertEquals(PairingCeremonyIosState.Idle, m.state)
    }

    // MARK: - CLI Machine

    @Test
    fun `cli starts in Idle`() {
        val m = PairingCeremonyCliMachine()
        assertEquals(PairingCeremonyCliState.Idle, m.state)
    }

    @Test
    fun `cli full flow`() {
        val m = PairingCeremonyCliMachine()

        m.handleEvent(PairingCeremonyProtocol.EventID.CliInit)
        assertEquals(PairingCeremonyCliState.GetKey, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.KeyStored)
        assertEquals(PairingCeremonyCliState.BeginPair, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.RecvTokenResponse)
        assertEquals(PairingCeremonyCliState.ShowQR, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.RecvWaitingForCode)
        assertEquals(PairingCeremonyCliState.PromptCode, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.UserEntersCode)
        assertEquals(PairingCeremonyCliState.SubmitCode, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.RecvPairStatus)
        assertEquals(PairingCeremonyCliState.Done, m.state)
    }

    @Test
    fun `cli invalid event does not change state`() {
        val m = PairingCeremonyCliMachine()
        m.handleEvent(PairingCeremonyProtocol.EventID.KeyStored) // invalid from Idle
        assertEquals(PairingCeremonyCliState.Idle, m.state)
    }

    // MARK: - Helpers

    private fun serverAtState(target: PairingCeremonyServerState): PairingCeremonyServerMachine {
        val m = PairingCeremonyServerMachine()
        m.actions[PairingCeremonyProtocol.ActionID.GenerateToken] = {}
        m.actions[PairingCeremonyProtocol.ActionID.RegisterRelay] = {}
        m.actions[PairingCeremonyProtocol.ActionID.DeriveSecret] = {}
        m.actions[PairingCeremonyProtocol.ActionID.StoreDevice] = {}
        m.actions[PairingCeremonyProtocol.ActionID.VerifyDevice] = {}
        m.guards[PairingCeremonyProtocol.GuardID.TokenValid] = { true }
        m.guards[PairingCeremonyProtocol.GuardID.TokenInvalid] = { false }
        m.guards[PairingCeremonyProtocol.GuardID.CodeCorrect] = { true }
        m.guards[PairingCeremonyProtocol.GuardID.CodeWrong] = { false }
        m.guards[PairingCeremonyProtocol.GuardID.DeviceKnown] = { true }
        m.guards[PairingCeremonyProtocol.GuardID.DeviceUnknown] = { false }

        val path = when (target) {
            PairingCeremonyServerState.Idle -> emptyList()
            PairingCeremonyServerState.GenerateToken -> listOf(PairingCeremonyProtocol.EventID.RecvPairBegin)
            PairingCeremonyServerState.RegisterRelay -> listOf(PairingCeremonyProtocol.EventID.RecvPairBegin, PairingCeremonyProtocol.EventID.TokenCreated)
            PairingCeremonyServerState.WaitingForClient -> listOf(PairingCeremonyProtocol.EventID.RecvPairBegin, PairingCeremonyProtocol.EventID.TokenCreated, PairingCeremonyProtocol.EventID.RelayRegistered)
            PairingCeremonyServerState.DeriveSecret -> listOf(PairingCeremonyProtocol.EventID.RecvPairBegin, PairingCeremonyProtocol.EventID.TokenCreated, PairingCeremonyProtocol.EventID.RelayRegistered, PairingCeremonyProtocol.EventID.RecvPairHello)
            PairingCeremonyServerState.SendAck -> listOf(PairingCeremonyProtocol.EventID.RecvPairBegin, PairingCeremonyProtocol.EventID.TokenCreated, PairingCeremonyProtocol.EventID.RelayRegistered, PairingCeremonyProtocol.EventID.RecvPairHello, PairingCeremonyProtocol.EventID.ECDHComplete)
            PairingCeremonyServerState.WaitingForCode -> listOf(PairingCeremonyProtocol.EventID.RecvPairBegin, PairingCeremonyProtocol.EventID.TokenCreated, PairingCeremonyProtocol.EventID.RelayRegistered, PairingCeremonyProtocol.EventID.RecvPairHello, PairingCeremonyProtocol.EventID.ECDHComplete, PairingCeremonyProtocol.EventID.SignalCodeDisplay)
            PairingCeremonyServerState.ValidateCode -> listOf(PairingCeremonyProtocol.EventID.RecvPairBegin, PairingCeremonyProtocol.EventID.TokenCreated, PairingCeremonyProtocol.EventID.RelayRegistered, PairingCeremonyProtocol.EventID.RecvPairHello, PairingCeremonyProtocol.EventID.ECDHComplete, PairingCeremonyProtocol.EventID.SignalCodeDisplay, PairingCeremonyProtocol.EventID.RecvCodeSubmit)
            PairingCeremonyServerState.StorePaired -> listOf(PairingCeremonyProtocol.EventID.RecvPairBegin, PairingCeremonyProtocol.EventID.TokenCreated, PairingCeremonyProtocol.EventID.RelayRegistered, PairingCeremonyProtocol.EventID.RecvPairHello, PairingCeremonyProtocol.EventID.ECDHComplete, PairingCeremonyProtocol.EventID.SignalCodeDisplay, PairingCeremonyProtocol.EventID.RecvCodeSubmit, PairingCeremonyProtocol.EventID.CheckCode)
            PairingCeremonyServerState.Paired -> listOf(PairingCeremonyProtocol.EventID.RecvPairBegin, PairingCeremonyProtocol.EventID.TokenCreated, PairingCeremonyProtocol.EventID.RelayRegistered, PairingCeremonyProtocol.EventID.RecvPairHello, PairingCeremonyProtocol.EventID.ECDHComplete, PairingCeremonyProtocol.EventID.SignalCodeDisplay, PairingCeremonyProtocol.EventID.RecvCodeSubmit, PairingCeremonyProtocol.EventID.CheckCode, PairingCeremonyProtocol.EventID.Finalise)
            PairingCeremonyServerState.AuthCheck -> listOf(PairingCeremonyProtocol.EventID.RecvPairBegin, PairingCeremonyProtocol.EventID.TokenCreated, PairingCeremonyProtocol.EventID.RelayRegistered, PairingCeremonyProtocol.EventID.RecvPairHello, PairingCeremonyProtocol.EventID.ECDHComplete, PairingCeremonyProtocol.EventID.SignalCodeDisplay, PairingCeremonyProtocol.EventID.RecvCodeSubmit, PairingCeremonyProtocol.EventID.CheckCode, PairingCeremonyProtocol.EventID.Finalise, PairingCeremonyProtocol.EventID.RecvAuthRequest)
            PairingCeremonyServerState.SessionActive -> listOf(PairingCeremonyProtocol.EventID.RecvPairBegin, PairingCeremonyProtocol.EventID.TokenCreated, PairingCeremonyProtocol.EventID.RelayRegistered, PairingCeremonyProtocol.EventID.RecvPairHello, PairingCeremonyProtocol.EventID.ECDHComplete, PairingCeremonyProtocol.EventID.SignalCodeDisplay, PairingCeremonyProtocol.EventID.RecvCodeSubmit, PairingCeremonyProtocol.EventID.CheckCode, PairingCeremonyProtocol.EventID.Finalise, PairingCeremonyProtocol.EventID.RecvAuthRequest, PairingCeremonyProtocol.EventID.Verify)
        }
        for (ev in path) {
            m.handleEvent(ev)
        }
        return m
    }
}

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
        val m = ServerMachine()
        assertEquals(ServerState.Idle, m.state)
    }

    @Test
    fun `server Idle to GenerateToken on recv pair_begin`() {
        val m = ServerMachine()
        var actionCalled = false
        m.actions[PairingCeremonyProtocol.ActionID.GenerateToken] = { actionCalled = true }

        m.handleEvent(PairingCeremonyProtocol.EventID.RecvPairBegin)
        assertEquals(ServerState.GenerateToken, m.state)
        assertTrue(actionCalled)
        assertEquals("tok_1", m.currentToken)
    }

    @Test
    fun `server GenerateToken to RegisterRelay on token created`() {
        val m = serverAtState(ServerState.GenerateToken)
        var actionCalled = false
        m.actions[PairingCeremonyProtocol.ActionID.RegisterRelay] = { actionCalled = true }

        m.handleEvent(PairingCeremonyProtocol.EventID.TokenCreated)
        assertEquals(ServerState.RegisterRelay, m.state)
        assertTrue(actionCalled)
    }

    @Test
    fun `server RegisterRelay to WaitingForClient on relay registered`() {
        val m = serverAtState(ServerState.RegisterRelay)
        m.handleEvent(PairingCeremonyProtocol.EventID.RelayRegistered)
        assertEquals(ServerState.WaitingForClient, m.state)
    }

    @Test
    fun `server token_valid guard allows transition to DeriveSecret`() {
        val m = serverAtState(ServerState.WaitingForClient)
        m.guards[PairingCeremonyProtocol.GuardID.TokenValid] = { true }
        m.guards[PairingCeremonyProtocol.GuardID.TokenInvalid] = { false }
        var actionCalled = false
        m.actions[PairingCeremonyProtocol.ActionID.DeriveSecret] = { actionCalled = true }

        m.handleEvent(PairingCeremonyProtocol.EventID.RecvPairHello)
        assertEquals(ServerState.DeriveSecret, m.state)
        assertTrue(actionCalled)
        assertEquals("server_pub", m.serverEcdhPub)
    }

    @Test
    fun `server token_invalid guard resets to Idle`() {
        val m = serverAtState(ServerState.WaitingForClient)
        m.guards[PairingCeremonyProtocol.GuardID.TokenValid] = { false }
        m.guards[PairingCeremonyProtocol.GuardID.TokenInvalid] = { true }

        m.handleEvent(PairingCeremonyProtocol.EventID.RecvPairHello)
        assertEquals(ServerState.Idle, m.state)
    }

    @Test
    fun `server code_correct guard transitions to StorePaired`() {
        val m = serverAtState(ServerState.ValidateCode)
        m.guards[PairingCeremonyProtocol.GuardID.CodeCorrect] = { true }
        m.guards[PairingCeremonyProtocol.GuardID.CodeWrong] = { false }

        m.handleEvent(PairingCeremonyProtocol.EventID.CheckCode)
        assertEquals(ServerState.StorePaired, m.state)
    }

    @Test
    fun `server code_wrong guard resets to Idle`() {
        val m = serverAtState(ServerState.ValidateCode)
        m.guards[PairingCeremonyProtocol.GuardID.CodeCorrect] = { false }
        m.guards[PairingCeremonyProtocol.GuardID.CodeWrong] = { true }

        m.handleEvent(PairingCeremonyProtocol.EventID.CheckCode)
        assertEquals(ServerState.Idle, m.state)
    }

    @Test
    fun `server finalise stores device and transitions to Paired`() {
        val m = serverAtState(ServerState.StorePaired)
        var actionCalled = false
        m.actions[PairingCeremonyProtocol.ActionID.StoreDevice] = { actionCalled = true }

        m.handleEvent(PairingCeremonyProtocol.EventID.Finalise)
        assertEquals(ServerState.Paired, m.state)
        assertTrue(actionCalled)
        assertEquals("dev_secret_1", m.deviceSecret)
    }

    @Test
    fun `server device_known guard transitions to SessionActive`() {
        val m = serverAtState(ServerState.AuthCheck)
        m.guards[PairingCeremonyProtocol.GuardID.DeviceKnown] = { true }
        m.guards[PairingCeremonyProtocol.GuardID.DeviceUnknown] = { false }
        var actionCalled = false
        m.actions[PairingCeremonyProtocol.ActionID.VerifyDevice] = { actionCalled = true }

        m.handleEvent(PairingCeremonyProtocol.EventID.Verify)
        assertEquals(ServerState.SessionActive, m.state)
        assertTrue(actionCalled)
    }

    @Test
    fun `server device_unknown guard resets to Idle`() {
        val m = serverAtState(ServerState.AuthCheck)
        m.guards[PairingCeremonyProtocol.GuardID.DeviceKnown] = { false }
        m.guards[PairingCeremonyProtocol.GuardID.DeviceUnknown] = { true }

        m.handleEvent(PairingCeremonyProtocol.EventID.Verify)
        assertEquals(ServerState.Idle, m.state)
    }

    @Test
    fun `server disconnect returns to Paired`() {
        val m = serverAtState(ServerState.SessionActive)
        m.handleEvent(PairingCeremonyProtocol.EventID.Disconnect)
        assertEquals(ServerState.Paired, m.state)
    }

    @Test
    fun `server invalid event does not change state`() {
        val m = ServerMachine()
        m.handleEvent(PairingCeremonyProtocol.EventID.Disconnect) // invalid from Idle
        assertEquals(ServerState.Idle, m.state)
    }

    @Test
    fun `server full pairing flow`() {
        val m = ServerMachine()
        m.actions[PairingCeremonyProtocol.ActionID.GenerateToken] = {}
        m.actions[PairingCeremonyProtocol.ActionID.RegisterRelay] = {}
        m.actions[PairingCeremonyProtocol.ActionID.DeriveSecret] = {}
        m.actions[PairingCeremonyProtocol.ActionID.StoreDevice] = {}
        m.guards[PairingCeremonyProtocol.GuardID.TokenValid] = { true }
        m.guards[PairingCeremonyProtocol.GuardID.TokenInvalid] = { false }
        m.guards[PairingCeremonyProtocol.GuardID.CodeCorrect] = { true }
        m.guards[PairingCeremonyProtocol.GuardID.CodeWrong] = { false }

        m.handleEvent(PairingCeremonyProtocol.EventID.RecvPairBegin)
        assertEquals(ServerState.GenerateToken, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.TokenCreated)
        assertEquals(ServerState.RegisterRelay, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.RelayRegistered)
        assertEquals(ServerState.WaitingForClient, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.RecvPairHello)
        assertEquals(ServerState.DeriveSecret, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.ECDHComplete)
        assertEquals(ServerState.SendAck, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.SignalCodeDisplay)
        assertEquals(ServerState.WaitingForCode, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.RecvCodeSubmit)
        assertEquals(ServerState.ValidateCode, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.CheckCode)
        assertEquals(ServerState.StorePaired, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.Finalise)
        assertEquals(ServerState.Paired, m.state)
    }

    // MARK: - iOS Machine

    @Test
    fun `ios starts in Idle`() {
        val m = IosMachine()
        assertEquals(IosState.Idle, m.state)
    }

    @Test
    fun `ios full pairing flow`() {
        val m = IosMachine()
        m.actions[PairingCeremonyProtocol.ActionID.SendPairHello] = {}
        m.actions[PairingCeremonyProtocol.ActionID.DeriveSecret] = {}
        m.actions[PairingCeremonyProtocol.ActionID.StoreSecret] = {}

        m.handleEvent(PairingCeremonyProtocol.EventID.UserScansQR)
        assertEquals(IosState.ScanQR, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.QRParsed)
        assertEquals(IosState.ConnectRelay, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.RelayConnected)
        assertEquals(IosState.GenKeyPair, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.KeyPairGenerated)
        assertEquals(IosState.WaitAck, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.RecvPairHelloAck)
        assertEquals(IosState.E2EReady, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.RecvPairConfirm)
        assertEquals(IosState.ShowCode, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.CodeDisplayed)
        assertEquals(IosState.WaitPairComplete, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.RecvPairComplete)
        assertEquals(IosState.Paired, m.state)
    }

    @Test
    fun `ios key pair generated calls sendPairHello action`() {
        val m = IosMachine()
        m.handleEvent(PairingCeremonyProtocol.EventID.UserScansQR)
        m.handleEvent(PairingCeremonyProtocol.EventID.QRParsed)
        m.handleEvent(PairingCeremonyProtocol.EventID.RelayConnected)

        var actionCalled = false
        m.actions[PairingCeremonyProtocol.ActionID.SendPairHello] = { actionCalled = true }

        m.handleEvent(PairingCeremonyProtocol.EventID.KeyPairGenerated)
        assertTrue(actionCalled)
        assertEquals(IosState.WaitAck, m.state)
    }

    @Test
    fun `ios reconnect and auth flow`() {
        val m = IosMachine()
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
        assertEquals(IosState.Paired, m.state)

        m.handleEvent(PairingCeremonyProtocol.EventID.AppLaunch)
        assertEquals(IosState.Reconnect, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.RelayConnected)
        assertEquals(IosState.SendAuth, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.RecvAuthOk)
        assertEquals(IosState.SessionActive, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.Disconnect)
        assertEquals(IosState.Paired, m.state)
    }

    @Test
    fun `ios invalid event does not change state`() {
        val m = IosMachine()
        m.handleEvent(PairingCeremonyProtocol.EventID.Disconnect) // invalid from Idle
        assertEquals(IosState.Idle, m.state)
    }

    // MARK: - CLI Machine

    @Test
    fun `cli starts in Idle`() {
        val m = CliMachine()
        assertEquals(CliState.Idle, m.state)
    }

    @Test
    fun `cli full flow`() {
        val m = CliMachine()

        m.handleEvent(PairingCeremonyProtocol.EventID.CliInit)
        assertEquals(CliState.GetKey, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.KeyStored)
        assertEquals(CliState.BeginPair, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.RecvTokenResponse)
        assertEquals(CliState.ShowQR, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.RecvWaitingForCode)
        assertEquals(CliState.PromptCode, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.UserEntersCode)
        assertEquals(CliState.SubmitCode, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.RecvPairStatus)
        assertEquals(CliState.Done, m.state)
    }

    @Test
    fun `cli invalid event does not change state`() {
        val m = CliMachine()
        m.handleEvent(PairingCeremonyProtocol.EventID.KeyStored) // invalid from Idle
        assertEquals(CliState.Idle, m.state)
    }

    // MARK: - Helpers

    private fun serverAtState(target: ServerState): ServerMachine {
        val m = ServerMachine()
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
            ServerState.Idle -> emptyList()
            ServerState.GenerateToken -> listOf(PairingCeremonyProtocol.EventID.RecvPairBegin)
            ServerState.RegisterRelay -> listOf(PairingCeremonyProtocol.EventID.RecvPairBegin, PairingCeremonyProtocol.EventID.TokenCreated)
            ServerState.WaitingForClient -> listOf(PairingCeremonyProtocol.EventID.RecvPairBegin, PairingCeremonyProtocol.EventID.TokenCreated, PairingCeremonyProtocol.EventID.RelayRegistered)
            ServerState.DeriveSecret -> listOf(PairingCeremonyProtocol.EventID.RecvPairBegin, PairingCeremonyProtocol.EventID.TokenCreated, PairingCeremonyProtocol.EventID.RelayRegistered, PairingCeremonyProtocol.EventID.RecvPairHello)
            ServerState.SendAck -> listOf(PairingCeremonyProtocol.EventID.RecvPairBegin, PairingCeremonyProtocol.EventID.TokenCreated, PairingCeremonyProtocol.EventID.RelayRegistered, PairingCeremonyProtocol.EventID.RecvPairHello, PairingCeremonyProtocol.EventID.ECDHComplete)
            ServerState.WaitingForCode -> listOf(PairingCeremonyProtocol.EventID.RecvPairBegin, PairingCeremonyProtocol.EventID.TokenCreated, PairingCeremonyProtocol.EventID.RelayRegistered, PairingCeremonyProtocol.EventID.RecvPairHello, PairingCeremonyProtocol.EventID.ECDHComplete, PairingCeremonyProtocol.EventID.SignalCodeDisplay)
            ServerState.ValidateCode -> listOf(PairingCeremonyProtocol.EventID.RecvPairBegin, PairingCeremonyProtocol.EventID.TokenCreated, PairingCeremonyProtocol.EventID.RelayRegistered, PairingCeremonyProtocol.EventID.RecvPairHello, PairingCeremonyProtocol.EventID.ECDHComplete, PairingCeremonyProtocol.EventID.SignalCodeDisplay, PairingCeremonyProtocol.EventID.RecvCodeSubmit)
            ServerState.StorePaired -> listOf(PairingCeremonyProtocol.EventID.RecvPairBegin, PairingCeremonyProtocol.EventID.TokenCreated, PairingCeremonyProtocol.EventID.RelayRegistered, PairingCeremonyProtocol.EventID.RecvPairHello, PairingCeremonyProtocol.EventID.ECDHComplete, PairingCeremonyProtocol.EventID.SignalCodeDisplay, PairingCeremonyProtocol.EventID.RecvCodeSubmit, PairingCeremonyProtocol.EventID.CheckCode)
            ServerState.Paired -> listOf(PairingCeremonyProtocol.EventID.RecvPairBegin, PairingCeremonyProtocol.EventID.TokenCreated, PairingCeremonyProtocol.EventID.RelayRegistered, PairingCeremonyProtocol.EventID.RecvPairHello, PairingCeremonyProtocol.EventID.ECDHComplete, PairingCeremonyProtocol.EventID.SignalCodeDisplay, PairingCeremonyProtocol.EventID.RecvCodeSubmit, PairingCeremonyProtocol.EventID.CheckCode, PairingCeremonyProtocol.EventID.Finalise)
            ServerState.AuthCheck -> listOf(PairingCeremonyProtocol.EventID.RecvPairBegin, PairingCeremonyProtocol.EventID.TokenCreated, PairingCeremonyProtocol.EventID.RelayRegistered, PairingCeremonyProtocol.EventID.RecvPairHello, PairingCeremonyProtocol.EventID.ECDHComplete, PairingCeremonyProtocol.EventID.SignalCodeDisplay, PairingCeremonyProtocol.EventID.RecvCodeSubmit, PairingCeremonyProtocol.EventID.CheckCode, PairingCeremonyProtocol.EventID.Finalise, PairingCeremonyProtocol.EventID.RecvAuthRequest)
            ServerState.SessionActive -> listOf(PairingCeremonyProtocol.EventID.RecvPairBegin, PairingCeremonyProtocol.EventID.TokenCreated, PairingCeremonyProtocol.EventID.RelayRegistered, PairingCeremonyProtocol.EventID.RecvPairHello, PairingCeremonyProtocol.EventID.ECDHComplete, PairingCeremonyProtocol.EventID.SignalCodeDisplay, PairingCeremonyProtocol.EventID.RecvCodeSubmit, PairingCeremonyProtocol.EventID.CheckCode, PairingCeremonyProtocol.EventID.Finalise, PairingCeremonyProtocol.EventID.RecvAuthRequest, PairingCeremonyProtocol.EventID.Verify)
        }
        for (ev in path) {
            m.handleEvent(ev)
        }
        return m
    }
}

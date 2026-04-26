// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package com.marcelocantos.pigeon.crypto

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertTrue

class PairingCeremonyServerPairingMachineTest {

    @Test
    fun `server pairing starts in Idle`() {
        val m = PairingCeremonyServerPairingMachine()
        assertEquals(PairingCeremonyServerPairingState.Idle, m.state)
    }

    @Test
    fun `server Idle to GenerateToken on recv pair_begin`() {
        val m = PairingCeremonyServerPairingMachine()
        var actionCalled = false
        m.actions[PairingCeremonyProtocol.ActionID.GenerateToken] = { actionCalled = true }

        m.handleEvent(PairingCeremonyProtocol.EventID.RecvPairBegin)
        assertEquals(PairingCeremonyServerPairingState.GenerateToken, m.state)
        assertTrue(actionCalled)
        assertEquals("tok_1", m.currentToken)
    }

    @Test
    fun `server GenerateToken to RegisterRelay on token created`() {
        val m = serverPairingAtState(PairingCeremonyServerPairingState.GenerateToken)
        var actionCalled = false
        m.actions[PairingCeremonyProtocol.ActionID.RegisterRelay] = { actionCalled = true }

        m.handleEvent(PairingCeremonyProtocol.EventID.TokenCreated)
        assertEquals(PairingCeremonyServerPairingState.RegisterRelay, m.state)
        assertTrue(actionCalled)
    }

    @Test
    fun `server RegisterRelay to WaitingForClient on relay registered`() {
        val m = serverPairingAtState(PairingCeremonyServerPairingState.RegisterRelay)
        m.handleEvent(PairingCeremonyProtocol.EventID.RelayRegistered)
        assertEquals(PairingCeremonyServerPairingState.WaitingForClient, m.state)
    }

    @Test
    fun `server token_valid guard allows transition to DeriveSecret`() {
        val m = serverPairingAtState(PairingCeremonyServerPairingState.WaitingForClient)
        m.guards[PairingCeremonyProtocol.GuardID.TokenValid] = { true }
        m.guards[PairingCeremonyProtocol.GuardID.TokenInvalid] = { false }
        var actionCalled = false
        m.actions[PairingCeremonyProtocol.ActionID.DeriveSecret] = { actionCalled = true }

        m.handleEvent(PairingCeremonyProtocol.EventID.RecvPairHello)
        assertEquals(PairingCeremonyServerPairingState.DeriveSecret, m.state)
        assertTrue(actionCalled)
        assertEquals("server_pub", m.serverEcdhPub)
    }

    @Test
    fun `server token_invalid guard resets to Idle`() {
        val m = serverPairingAtState(PairingCeremonyServerPairingState.WaitingForClient)
        m.guards[PairingCeremonyProtocol.GuardID.TokenValid] = { false }
        m.guards[PairingCeremonyProtocol.GuardID.TokenInvalid] = { true }

        m.handleEvent(PairingCeremonyProtocol.EventID.RecvPairHello)
        assertEquals(PairingCeremonyServerPairingState.Idle, m.state)
    }

    @Test
    fun `server code_correct guard transitions to StorePaired`() {
        val m = serverPairingAtState(PairingCeremonyServerPairingState.ValidateCode)
        m.guards[PairingCeremonyProtocol.GuardID.CodeCorrect] = { true }
        m.guards[PairingCeremonyProtocol.GuardID.CodeWrong] = { false }

        m.handleEvent(PairingCeremonyProtocol.EventID.CheckCode)
        assertEquals(PairingCeremonyServerPairingState.StorePaired, m.state)
    }

    @Test
    fun `server code_wrong guard resets to Idle`() {
        val m = serverPairingAtState(PairingCeremonyServerPairingState.ValidateCode)
        m.guards[PairingCeremonyProtocol.GuardID.CodeCorrect] = { false }
        m.guards[PairingCeremonyProtocol.GuardID.CodeWrong] = { true }

        m.handleEvent(PairingCeremonyProtocol.EventID.CheckCode)
        assertEquals(PairingCeremonyServerPairingState.Idle, m.state)
    }

    @Test
    fun `server finalise stores device and transitions to PairingComplete`() {
        val m = serverPairingAtState(PairingCeremonyServerPairingState.StorePaired)
        var actionCalled = false
        m.actions[PairingCeremonyProtocol.ActionID.StoreDevice] = { actionCalled = true }

        m.handleEvent(PairingCeremonyProtocol.EventID.Finalise)
        assertEquals(PairingCeremonyServerPairingState.PairingComplete, m.state)
        assertTrue(actionCalled)
        assertEquals("dev_secret_1", m.deviceSecret)
    }

    @Test
    fun `server pairing invalid event does not change state`() {
        val m = PairingCeremonyServerPairingMachine()
        m.handleEvent(PairingCeremonyProtocol.EventID.Disconnect) // invalid from Idle
        assertEquals(PairingCeremonyServerPairingState.Idle, m.state)
    }

    @Test
    fun `server full pairing flow`() {
        val m = PairingCeremonyServerPairingMachine()
        m.actions[PairingCeremonyProtocol.ActionID.GenerateToken] = {}
        m.actions[PairingCeremonyProtocol.ActionID.RegisterRelay] = {}
        m.actions[PairingCeremonyProtocol.ActionID.DeriveSecret] = {}
        m.actions[PairingCeremonyProtocol.ActionID.StoreDevice] = {}
        m.guards[PairingCeremonyProtocol.GuardID.TokenValid] = { true }
        m.guards[PairingCeremonyProtocol.GuardID.TokenInvalid] = { false }
        m.guards[PairingCeremonyProtocol.GuardID.CodeCorrect] = { true }
        m.guards[PairingCeremonyProtocol.GuardID.CodeWrong] = { false }

        m.handleEvent(PairingCeremonyProtocol.EventID.RecvPairBegin)
        assertEquals(PairingCeremonyServerPairingState.GenerateToken, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.TokenCreated)
        assertEquals(PairingCeremonyServerPairingState.RegisterRelay, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.RelayRegistered)
        assertEquals(PairingCeremonyServerPairingState.WaitingForClient, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.RecvPairHello)
        assertEquals(PairingCeremonyServerPairingState.DeriveSecret, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.ECDHComplete)
        assertEquals(PairingCeremonyServerPairingState.SendAck, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.SignalCodeDisplay)
        assertEquals(PairingCeremonyServerPairingState.WaitingForCode, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.RecvCodeSubmit)
        assertEquals(PairingCeremonyServerPairingState.ValidateCode, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.CheckCode)
        assertEquals(PairingCeremonyServerPairingState.StorePaired, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.Finalise)
        assertEquals(PairingCeremonyServerPairingState.PairingComplete, m.state)
    }

    private fun serverPairingAtState(target: PairingCeremonyServerPairingState): PairingCeremonyServerPairingMachine {
        val m = PairingCeremonyServerPairingMachine()
        m.actions[PairingCeremonyProtocol.ActionID.GenerateToken] = {}
        m.actions[PairingCeremonyProtocol.ActionID.RegisterRelay] = {}
        m.actions[PairingCeremonyProtocol.ActionID.DeriveSecret] = {}
        m.actions[PairingCeremonyProtocol.ActionID.StoreDevice] = {}
        m.guards[PairingCeremonyProtocol.GuardID.TokenValid] = { true }
        m.guards[PairingCeremonyProtocol.GuardID.TokenInvalid] = { false }
        m.guards[PairingCeremonyProtocol.GuardID.CodeCorrect] = { true }
        m.guards[PairingCeremonyProtocol.GuardID.CodeWrong] = { false }

        val path = when (target) {
            PairingCeremonyServerPairingState.Idle -> emptyList()
            PairingCeremonyServerPairingState.GenerateToken -> listOf(PairingCeremonyProtocol.EventID.RecvPairBegin)
            PairingCeremonyServerPairingState.RegisterRelay -> listOf(PairingCeremonyProtocol.EventID.RecvPairBegin, PairingCeremonyProtocol.EventID.TokenCreated)
            PairingCeremonyServerPairingState.WaitingForClient -> listOf(PairingCeremonyProtocol.EventID.RecvPairBegin, PairingCeremonyProtocol.EventID.TokenCreated, PairingCeremonyProtocol.EventID.RelayRegistered)
            PairingCeremonyServerPairingState.DeriveSecret -> listOf(PairingCeremonyProtocol.EventID.RecvPairBegin, PairingCeremonyProtocol.EventID.TokenCreated, PairingCeremonyProtocol.EventID.RelayRegistered, PairingCeremonyProtocol.EventID.RecvPairHello)
            PairingCeremonyServerPairingState.SendAck -> listOf(PairingCeremonyProtocol.EventID.RecvPairBegin, PairingCeremonyProtocol.EventID.TokenCreated, PairingCeremonyProtocol.EventID.RelayRegistered, PairingCeremonyProtocol.EventID.RecvPairHello, PairingCeremonyProtocol.EventID.ECDHComplete)
            PairingCeremonyServerPairingState.WaitingForCode -> listOf(PairingCeremonyProtocol.EventID.RecvPairBegin, PairingCeremonyProtocol.EventID.TokenCreated, PairingCeremonyProtocol.EventID.RelayRegistered, PairingCeremonyProtocol.EventID.RecvPairHello, PairingCeremonyProtocol.EventID.ECDHComplete, PairingCeremonyProtocol.EventID.SignalCodeDisplay)
            PairingCeremonyServerPairingState.ValidateCode -> listOf(PairingCeremonyProtocol.EventID.RecvPairBegin, PairingCeremonyProtocol.EventID.TokenCreated, PairingCeremonyProtocol.EventID.RelayRegistered, PairingCeremonyProtocol.EventID.RecvPairHello, PairingCeremonyProtocol.EventID.ECDHComplete, PairingCeremonyProtocol.EventID.SignalCodeDisplay, PairingCeremonyProtocol.EventID.RecvCodeSubmit)
            PairingCeremonyServerPairingState.StorePaired -> listOf(PairingCeremonyProtocol.EventID.RecvPairBegin, PairingCeremonyProtocol.EventID.TokenCreated, PairingCeremonyProtocol.EventID.RelayRegistered, PairingCeremonyProtocol.EventID.RecvPairHello, PairingCeremonyProtocol.EventID.ECDHComplete, PairingCeremonyProtocol.EventID.SignalCodeDisplay, PairingCeremonyProtocol.EventID.RecvCodeSubmit, PairingCeremonyProtocol.EventID.CheckCode)
            PairingCeremonyServerPairingState.PairingComplete -> listOf(PairingCeremonyProtocol.EventID.RecvPairBegin, PairingCeremonyProtocol.EventID.TokenCreated, PairingCeremonyProtocol.EventID.RelayRegistered, PairingCeremonyProtocol.EventID.RecvPairHello, PairingCeremonyProtocol.EventID.ECDHComplete, PairingCeremonyProtocol.EventID.SignalCodeDisplay, PairingCeremonyProtocol.EventID.RecvCodeSubmit, PairingCeremonyProtocol.EventID.CheckCode, PairingCeremonyProtocol.EventID.Finalise)
        }
        for (ev in path) {
            m.handleEvent(ev)
        }
        return m
    }
}

class PairingCeremonyServerAuthMachineTest {

    @Test
    fun `server auth starts in Idle`() {
        val m = PairingCeremonyServerAuthMachine()
        assertEquals(PairingCeremonyServerAuthState.Idle, m.state)
    }

    @Test
    fun `server auth Idle to Paired on credential_ready`() {
        val m = PairingCeremonyServerAuthMachine()
        m.handleEvent(PairingCeremonyProtocol.EventID.CredentialReady)
        assertEquals(PairingCeremonyServerAuthState.Paired, m.state)
    }

    @Test
    fun `server auth Paired to AuthCheck on recv_auth_request`() {
        val m = serverAuthAtState(PairingCeremonyServerAuthState.Paired)
        m.handleEvent(PairingCeremonyProtocol.EventID.RecvAuthRequest)
        assertEquals(PairingCeremonyServerAuthState.AuthCheck, m.state)
    }

    @Test
    fun `server device_known guard transitions to SessionActive`() {
        val m = serverAuthAtState(PairingCeremonyServerAuthState.AuthCheck)
        m.guards[PairingCeremonyProtocol.GuardID.DeviceKnown] = { true }
        m.guards[PairingCeremonyProtocol.GuardID.DeviceUnknown] = { false }
        var actionCalled = false
        m.actions[PairingCeremonyProtocol.ActionID.VerifyDevice] = { actionCalled = true }

        m.handleEvent(PairingCeremonyProtocol.EventID.Verify)
        assertEquals(PairingCeremonyServerAuthState.SessionActive, m.state)
        assertTrue(actionCalled)
    }

    @Test
    fun `server device_unknown guard resets to Idle`() {
        val m = serverAuthAtState(PairingCeremonyServerAuthState.AuthCheck)
        m.guards[PairingCeremonyProtocol.GuardID.DeviceKnown] = { false }
        m.guards[PairingCeremonyProtocol.GuardID.DeviceUnknown] = { true }

        m.handleEvent(PairingCeremonyProtocol.EventID.Verify)
        assertEquals(PairingCeremonyServerAuthState.Idle, m.state)
    }

    @Test
    fun `server disconnect returns to Paired`() {
        val m = serverAuthAtState(PairingCeremonyServerAuthState.SessionActive)
        m.handleEvent(PairingCeremonyProtocol.EventID.Disconnect)
        assertEquals(PairingCeremonyServerAuthState.Paired, m.state)
    }

    @Test
    fun `server auth invalid event does not change state`() {
        val m = PairingCeremonyServerAuthMachine()
        m.handleEvent(PairingCeremonyProtocol.EventID.Disconnect) // invalid from Idle
        assertEquals(PairingCeremonyServerAuthState.Idle, m.state)
    }

    @Test
    fun `server full auth flow`() {
        val m = PairingCeremonyServerAuthMachine()
        m.actions[PairingCeremonyProtocol.ActionID.VerifyDevice] = {}
        m.guards[PairingCeremonyProtocol.GuardID.DeviceKnown] = { true }
        m.guards[PairingCeremonyProtocol.GuardID.DeviceUnknown] = { false }

        m.handleEvent(PairingCeremonyProtocol.EventID.CredentialReady)
        assertEquals(PairingCeremonyServerAuthState.Paired, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.RecvAuthRequest)
        assertEquals(PairingCeremonyServerAuthState.AuthCheck, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.Verify)
        assertEquals(PairingCeremonyServerAuthState.SessionActive, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.Disconnect)
        assertEquals(PairingCeremonyServerAuthState.Paired, m.state)
    }

    private fun serverAuthAtState(target: PairingCeremonyServerAuthState): PairingCeremonyServerAuthMachine {
        val m = PairingCeremonyServerAuthMachine()
        m.actions[PairingCeremonyProtocol.ActionID.VerifyDevice] = {}
        m.guards[PairingCeremonyProtocol.GuardID.DeviceKnown] = { true }
        m.guards[PairingCeremonyProtocol.GuardID.DeviceUnknown] = { false }

        val path = when (target) {
            PairingCeremonyServerAuthState.Idle -> emptyList()
            PairingCeremonyServerAuthState.Paired -> listOf(PairingCeremonyProtocol.EventID.CredentialReady)
            PairingCeremonyServerAuthState.AuthCheck -> listOf(PairingCeremonyProtocol.EventID.CredentialReady, PairingCeremonyProtocol.EventID.RecvAuthRequest)
            PairingCeremonyServerAuthState.SessionActive -> listOf(PairingCeremonyProtocol.EventID.CredentialReady, PairingCeremonyProtocol.EventID.RecvAuthRequest, PairingCeremonyProtocol.EventID.Verify)
        }
        for (ev in path) {
            m.handleEvent(ev)
        }
        return m
    }
}

class PairingCeremonyIosPairingMachineTest {

    @Test
    fun `ios pairing starts in Idle`() {
        val m = PairingCeremonyIosPairingMachine()
        assertEquals(PairingCeremonyIosPairingState.Idle, m.state)
    }

    @Test
    fun `ios full pairing flow`() {
        val m = PairingCeremonyIosPairingMachine()
        m.actions[PairingCeremonyProtocol.ActionID.SendPairHello] = {}
        m.actions[PairingCeremonyProtocol.ActionID.DeriveSecret] = {}
        m.actions[PairingCeremonyProtocol.ActionID.StoreSecret] = {}

        m.handleEvent(PairingCeremonyProtocol.EventID.UserScansQR)
        assertEquals(PairingCeremonyIosPairingState.ScanQR, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.QRParsed)
        assertEquals(PairingCeremonyIosPairingState.ConnectRelay, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.RelayConnected)
        assertEquals(PairingCeremonyIosPairingState.GenKeyPair, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.KeyPairGenerated)
        assertEquals(PairingCeremonyIosPairingState.WaitAck, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.RecvPairHelloAck)
        assertEquals(PairingCeremonyIosPairingState.E2EReady, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.RecvPairConfirm)
        assertEquals(PairingCeremonyIosPairingState.ShowCode, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.CodeDisplayed)
        assertEquals(PairingCeremonyIosPairingState.WaitPairComplete, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.RecvPairComplete)
        assertEquals(PairingCeremonyIosPairingState.PairingComplete, m.state)
    }

    @Test
    fun `ios key pair generated calls sendPairHello action`() {
        val m = PairingCeremonyIosPairingMachine()
        m.handleEvent(PairingCeremonyProtocol.EventID.UserScansQR)
        m.handleEvent(PairingCeremonyProtocol.EventID.QRParsed)
        m.handleEvent(PairingCeremonyProtocol.EventID.RelayConnected)

        var actionCalled = false
        m.actions[PairingCeremonyProtocol.ActionID.SendPairHello] = { actionCalled = true }

        m.handleEvent(PairingCeremonyProtocol.EventID.KeyPairGenerated)
        assertTrue(actionCalled)
        assertEquals(PairingCeremonyIosPairingState.WaitAck, m.state)
    }

    @Test
    fun `ios pairing invalid event does not change state`() {
        val m = PairingCeremonyIosPairingMachine()
        m.handleEvent(PairingCeremonyProtocol.EventID.Disconnect) // invalid from Idle
        assertEquals(PairingCeremonyIosPairingState.Idle, m.state)
    }
}

class PairingCeremonyIosAuthMachineTest {

    @Test
    fun `ios auth starts in Idle`() {
        val m = PairingCeremonyIosAuthMachine()
        assertEquals(PairingCeremonyIosAuthState.Idle, m.state)
    }

    @Test
    fun `ios auth Idle to Paired on credential_ready`() {
        val m = PairingCeremonyIosAuthMachine()
        m.handleEvent(PairingCeremonyProtocol.EventID.CredentialReady)
        assertEquals(PairingCeremonyIosAuthState.Paired, m.state)
    }

    @Test
    fun `ios reconnect and auth flow`() {
        val m = PairingCeremonyIosAuthMachine()

        m.handleEvent(PairingCeremonyProtocol.EventID.CredentialReady)
        assertEquals(PairingCeremonyIosAuthState.Paired, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.AppLaunch)
        assertEquals(PairingCeremonyIosAuthState.Reconnect, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.RelayConnected)
        assertEquals(PairingCeremonyIosAuthState.SendAuth, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.RecvAuthOk)
        assertEquals(PairingCeremonyIosAuthState.SessionActive, m.state)
        m.handleEvent(PairingCeremonyProtocol.EventID.Disconnect)
        assertEquals(PairingCeremonyIosAuthState.Paired, m.state)
    }

    @Test
    fun `ios auth invalid event does not change state`() {
        val m = PairingCeremonyIosAuthMachine()
        m.handleEvent(PairingCeremonyProtocol.EventID.Disconnect) // invalid from Idle
        assertEquals(PairingCeremonyIosAuthState.Idle, m.state)
    }
}

class PairingCeremonyCliMachineTest {

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
}

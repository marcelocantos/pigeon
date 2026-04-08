// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Auto-generated from protocol definition. Do not edit.
// Source of truth: protocol/*.yaml

package com.marcelocantos.pigeon.crypto

enum class PairingCeremonyServerState(val value: String) {
    Idle("Idle"),
    GenerateToken("GenerateToken"),
    RegisterRelay("RegisterRelay"),
    WaitingForClient("WaitingForClient"),
    DeriveSecret("DeriveSecret"),
    SendAck("SendAck"),
    WaitingForCode("WaitingForCode"),
    ValidateCode("ValidateCode"),
    StorePaired("StorePaired"),
    Paired("Paired"),
    AuthCheck("AuthCheck"),
    SessionActive("SessionActive");
}

enum class PairingCeremonyIosState(val value: String) {
    Idle("Idle"),
    ScanQR("ScanQR"),
    ConnectRelay("ConnectRelay"),
    GenKeyPair("GenKeyPair"),
    WaitAck("WaitAck"),
    E2EReady("E2EReady"),
    ShowCode("ShowCode"),
    WaitPairComplete("WaitPairComplete"),
    Paired("Paired"),
    Reconnect("Reconnect"),
    SendAuth("SendAuth"),
    SessionActive("SessionActive");
}

enum class PairingCeremonyCliState(val value: String) {
    Idle("Idle"),
    GetKey("GetKey"),
    BeginPair("BeginPair"),
    ShowQR("ShowQR"),
    PromptCode("PromptCode"),
    SubmitCode("SubmitCode"),
    Done("Done");
}

/** The protocol transition table and shared type enums. */
object PairingCeremonyProtocol {

    enum class MessageType(val value: String) {
        PairBegin("pair_begin"),
        TokenResponse("token_response"),
        PairHello("pair_hello"),
        PairHelloAck("pair_hello_ack"),
        PairConfirm("pair_confirm"),
        WaitingForCode("waiting_for_code"),
        CodeSubmit("code_submit"),
        PairComplete("pair_complete"),
        PairStatus("pair_status"),
        AuthRequest("auth_request"),
        AuthOk("auth_ok");
    }

    enum class GuardID(val value: String) {
        TokenValid("token_valid"),
        TokenInvalid("token_invalid"),
        CodeCorrect("code_correct"),
        CodeWrong("code_wrong"),
        DeviceKnown("device_known"),
        DeviceUnknown("device_unknown"),
        NonceFresh("nonce_fresh");
    }

    enum class ActionID(val value: String) {
        GenerateToken("generate_token"),
        RegisterRelay("register_relay"),
        DeriveSecret("derive_secret"),
        StoreDevice("store_device"),
        VerifyDevice("verify_device"),
        SendPairHello("send_pair_hello"),
        StoreSecret("store_secret");
    }

    enum class EventID(val value: String) {
        ECDHComplete("ECDH complete"),
        QRParsed("QR parsed"),
        AppLaunch("app launch"),
        CheckCode("check code"),
        CliInit("cli --init"),
        CodeDisplayed("code displayed"),
        Disconnect("disconnect"),
        Finalise("finalise"),
        KeyPairGenerated("key pair generated"),
        KeyStored("key stored"),
        RecvAuthOk("recv_auth_ok"),
        RecvAuthRequest("recv_auth_request"),
        RecvCodeSubmit("recv_code_submit"),
        RecvPairBegin("recv_pair_begin"),
        RecvPairComplete("recv_pair_complete"),
        RecvPairConfirm("recv_pair_confirm"),
        RecvPairHello("recv_pair_hello"),
        RecvPairHelloAck("recv_pair_hello_ack"),
        RecvPairStatus("recv_pair_status"),
        RecvTokenResponse("recv_token_response"),
        RecvWaitingForCode("recv_waiting_for_code"),
        RelayConnected("relay connected"),
        RelayRegistered("relay registered"),
        SignalCodeDisplay("signal code display"),
        TokenCreated("token created"),
        UserEntersCode("user enters code"),
        UserScansQR("user scans QR"),
        Verify("verify");
    }

    /** server transition table. */
    object ServerTable {
        val initial = PairingCeremonyServerState.Idle

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
            Transition("Idle", "GenerateToken", "pair_begin", "recv", null, "generate_token", emptyList()),
            Transition("GenerateToken", "RegisterRelay", "token created", "internal", null, "register_relay", emptyList()),
            Transition("RegisterRelay", "WaitingForClient", "relay registered", "internal", null, null, listOf("cli" to "token_response")),
            Transition("WaitingForClient", "DeriveSecret", "pair_hello", "recv", "token_valid", "derive_secret", emptyList()),
            Transition("WaitingForClient", "Idle", "pair_hello", "recv", "token_invalid", null, emptyList()),
            Transition("DeriveSecret", "SendAck", "ECDH complete", "internal", null, null, listOf("ios" to "pair_hello_ack")),
            Transition("SendAck", "WaitingForCode", "signal code display", "internal", null, null, listOf("ios" to "pair_confirm", "cli" to "waiting_for_code")),
            Transition("WaitingForCode", "ValidateCode", "code_submit", "recv", null, null, emptyList()),
            Transition("ValidateCode", "StorePaired", "check code", "internal", "code_correct", null, emptyList()),
            Transition("ValidateCode", "Idle", "check code", "internal", "code_wrong", null, emptyList()),
            Transition("StorePaired", "Paired", "finalise", "internal", null, "store_device", listOf("ios" to "pair_complete", "cli" to "pair_status")),
            Transition("Paired", "AuthCheck", "auth_request", "recv", null, null, emptyList()),
            Transition("AuthCheck", "SessionActive", "verify", "internal", "device_known", "verify_device", listOf("ios" to "auth_ok")),
            Transition("AuthCheck", "Idle", "verify", "internal", "device_unknown", null, emptyList()),
            Transition("SessionActive", "Paired", "disconnect", "internal", null, null, emptyList()),
        )
    }

    /** ios transition table. */
    object IosTable {
        val initial = PairingCeremonyIosState.Idle

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
            Transition("Idle", "ScanQR", "user scans QR", "internal", null, null, emptyList()),
            Transition("ScanQR", "ConnectRelay", "QR parsed", "internal", null, null, emptyList()),
            Transition("ConnectRelay", "GenKeyPair", "relay connected", "internal", null, null, emptyList()),
            Transition("GenKeyPair", "WaitAck", "key pair generated", "internal", null, "send_pair_hello", listOf("server" to "pair_hello")),
            Transition("WaitAck", "E2EReady", "pair_hello_ack", "recv", null, "derive_secret", emptyList()),
            Transition("E2EReady", "ShowCode", "pair_confirm", "recv", null, null, emptyList()),
            Transition("ShowCode", "WaitPairComplete", "code displayed", "internal", null, null, emptyList()),
            Transition("WaitPairComplete", "Paired", "pair_complete", "recv", null, "store_secret", emptyList()),
            Transition("Paired", "Reconnect", "app launch", "internal", null, null, emptyList()),
            Transition("Reconnect", "SendAuth", "relay connected", "internal", null, null, listOf("server" to "auth_request")),
            Transition("SendAuth", "SessionActive", "auth_ok", "recv", null, null, emptyList()),
            Transition("SessionActive", "Paired", "disconnect", "internal", null, null, emptyList()),
        )
    }

    /** cli transition table. */
    object CliTable {
        val initial = PairingCeremonyCliState.Idle

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
            Transition("Idle", "GetKey", "cli --init", "internal", null, null, emptyList()),
            Transition("GetKey", "BeginPair", "key stored", "internal", null, null, listOf("server" to "pair_begin")),
            Transition("BeginPair", "ShowQR", "token_response", "recv", null, null, emptyList()),
            Transition("ShowQR", "PromptCode", "waiting_for_code", "recv", null, null, emptyList()),
            Transition("PromptCode", "SubmitCode", "user enters code", "internal", null, null, listOf("server" to "code_submit")),
            Transition("SubmitCode", "Done", "pair_status", "recv", null, null, emptyList()),
        )
    }

}

/** PairingCeremonyServerMachine is the generated state machine for the server actor. */
class PairingCeremonyServerMachine {
    var state: PairingCeremonyServerState = PairingCeremonyServerState.Idle
        private set
    var currentToken: String = "none" // pairing token currently in play
    var activeTokens: String = "" // set of valid (non-revoked) tokens
    var usedTokens: String = "" // set of revoked tokens
    var serverEcdhPub: String = "none" // server ECDH public key
    var receivedClientPub: String = "none" // pubkey server received in pair_hello (may be adversary's)
    var serverSharedKey: String = "" // ECDH key derived by server (tuple to match DeriveKey output type)
    var serverCode: String = "" // code computed by server from its view of the pubkeys (tuple to match DeriveCode output type)
    var receivedCode: String = "" // code received in code_submit (tuple to match DeriveCode output type)
    var codeAttempts: Int = 0 // failed code submission attempts
    var deviceSecret: String = "none" // persistent device secret
    var pairedDevices: String = "" // device IDs that completed pairing
    var receivedDeviceId: String = "none" // device_id from auth_request
    var authNoncesUsed: String = "" // set of consumed auth nonces
    var receivedAuthNonce: String = "none" // nonce from auth_request
    val guards = mutableMapOf<PairingCeremonyProtocol.GuardID, () -> Boolean>()
    val actions = mutableMapOf<PairingCeremonyProtocol.ActionID, () -> Unit>()

    /** Handle an event and return the list of commands to execute. */
    fun handleEvent(ev: PairingCeremonyProtocol.EventID): List<String> {
        val cmds: List<String> = when {
            state == PairingCeremonyServerState.Idle && ev == PairingCeremonyProtocol.EventID.RecvPairBegin ->
                run {
                    actions[PairingCeremonyProtocol.ActionID.GenerateToken]?.invoke()
                    currentToken = "tok_1"
                    // active_tokens: active_tokens \union {"tok_1"} (set by action)
                    state = PairingCeremonyServerState.GenerateToken
                    emptyList()
                }
            state == PairingCeremonyServerState.GenerateToken && ev == PairingCeremonyProtocol.EventID.TokenCreated ->
                run {
                    actions[PairingCeremonyProtocol.ActionID.RegisterRelay]?.invoke()
                    state = PairingCeremonyServerState.RegisterRelay
                    emptyList()
                }
            state == PairingCeremonyServerState.RegisterRelay && ev == PairingCeremonyProtocol.EventID.RelayRegistered ->
                run {
                    state = PairingCeremonyServerState.WaitingForClient
                    emptyList()
                }
            state == PairingCeremonyServerState.WaitingForClient && ev == PairingCeremonyProtocol.EventID.RecvPairHello && guards[PairingCeremonyProtocol.GuardID.TokenValid]?.invoke() == true ->
                run {
                    actions[PairingCeremonyProtocol.ActionID.DeriveSecret]?.invoke()
                    // received_client_pub: recv_msg.pubkey (set by action)
                    serverEcdhPub = "server_pub"
                    // server_shared_key: DeriveKey("server_pub", recv_msg.pubkey) (set by action)
                    // server_code: DeriveCode("server_pub", recv_msg.pubkey) (set by action)
                    state = PairingCeremonyServerState.DeriveSecret
                    emptyList()
                }
            state == PairingCeremonyServerState.WaitingForClient && ev == PairingCeremonyProtocol.EventID.RecvPairHello && guards[PairingCeremonyProtocol.GuardID.TokenInvalid]?.invoke() == true ->
                run {
                    state = PairingCeremonyServerState.Idle
                    emptyList()
                }
            state == PairingCeremonyServerState.DeriveSecret && ev == PairingCeremonyProtocol.EventID.ECDHComplete ->
                run {
                    state = PairingCeremonyServerState.SendAck
                    emptyList()
                }
            state == PairingCeremonyServerState.SendAck && ev == PairingCeremonyProtocol.EventID.SignalCodeDisplay ->
                run {
                    state = PairingCeremonyServerState.WaitingForCode
                    emptyList()
                }
            state == PairingCeremonyServerState.WaitingForCode && ev == PairingCeremonyProtocol.EventID.RecvCodeSubmit ->
                run {
                    // received_code: recv_msg.code (set by action)
                    state = PairingCeremonyServerState.ValidateCode
                    emptyList()
                }
            state == PairingCeremonyServerState.ValidateCode && ev == PairingCeremonyProtocol.EventID.CheckCode && guards[PairingCeremonyProtocol.GuardID.CodeCorrect]?.invoke() == true ->
                run {
                    state = PairingCeremonyServerState.StorePaired
                    emptyList()
                }
            state == PairingCeremonyServerState.ValidateCode && ev == PairingCeremonyProtocol.EventID.CheckCode && guards[PairingCeremonyProtocol.GuardID.CodeWrong]?.invoke() == true ->
                run {
                    // code_attempts: code_attempts + 1 (set by action)
                    state = PairingCeremonyServerState.Idle
                    emptyList()
                }
            state == PairingCeremonyServerState.StorePaired && ev == PairingCeremonyProtocol.EventID.Finalise ->
                run {
                    actions[PairingCeremonyProtocol.ActionID.StoreDevice]?.invoke()
                    deviceSecret = "dev_secret_1"
                    // paired_devices: paired_devices \union {"device_1"} (set by action)
                    // active_tokens: active_tokens \ {current_token} (set by action)
                    // used_tokens: used_tokens \union {current_token} (set by action)
                    state = PairingCeremonyServerState.Paired
                    emptyList()
                }
            state == PairingCeremonyServerState.Paired && ev == PairingCeremonyProtocol.EventID.RecvAuthRequest ->
                run {
                    // received_device_id: recv_msg.device_id (set by action)
                    // received_auth_nonce: recv_msg.nonce (set by action)
                    state = PairingCeremonyServerState.AuthCheck
                    emptyList()
                }
            state == PairingCeremonyServerState.AuthCheck && ev == PairingCeremonyProtocol.EventID.Verify && guards[PairingCeremonyProtocol.GuardID.DeviceKnown]?.invoke() == true ->
                run {
                    actions[PairingCeremonyProtocol.ActionID.VerifyDevice]?.invoke()
                    // auth_nonces_used: auth_nonces_used \union {received_auth_nonce} (set by action)
                    state = PairingCeremonyServerState.SessionActive
                    emptyList()
                }
            state == PairingCeremonyServerState.AuthCheck && ev == PairingCeremonyProtocol.EventID.Verify && guards[PairingCeremonyProtocol.GuardID.DeviceUnknown]?.invoke() == true ->
                run {
                    state = PairingCeremonyServerState.Idle
                    emptyList()
                }
            state == PairingCeremonyServerState.SessionActive && ev == PairingCeremonyProtocol.EventID.Disconnect ->
                run {
                    state = PairingCeremonyServerState.Paired
                    emptyList()
                }
            else -> emptyList()
        }
        return cmds
    }
}

/** PairingCeremonyIosMachine is the generated state machine for the ios actor. */
class PairingCeremonyIosMachine {
    var state: PairingCeremonyIosState = PairingCeremonyIosState.Idle
        private set
    var receivedServerPub: String = "none" // pubkey ios received in pair_hello_ack (may be adversary's)
    var clientSharedKey: String = "" // ECDH key derived by ios (tuple to match DeriveKey output type)
    var iosCode: String = "" // code computed by ios from its view of the pubkeys (tuple to match DeriveCode output type)
    val guards = mutableMapOf<PairingCeremonyProtocol.GuardID, () -> Boolean>()
    val actions = mutableMapOf<PairingCeremonyProtocol.ActionID, () -> Unit>()

    /** Handle an event and return the list of commands to execute. */
    fun handleEvent(ev: PairingCeremonyProtocol.EventID): List<String> {
        val cmds: List<String> = when {
            state == PairingCeremonyIosState.Idle && ev == PairingCeremonyProtocol.EventID.UserScansQR ->
                run {
                    state = PairingCeremonyIosState.ScanQR
                    emptyList()
                }
            state == PairingCeremonyIosState.ScanQR && ev == PairingCeremonyProtocol.EventID.QRParsed ->
                run {
                    state = PairingCeremonyIosState.ConnectRelay
                    emptyList()
                }
            state == PairingCeremonyIosState.ConnectRelay && ev == PairingCeremonyProtocol.EventID.RelayConnected ->
                run {
                    state = PairingCeremonyIosState.GenKeyPair
                    emptyList()
                }
            state == PairingCeremonyIosState.GenKeyPair && ev == PairingCeremonyProtocol.EventID.KeyPairGenerated ->
                run {
                    actions[PairingCeremonyProtocol.ActionID.SendPairHello]?.invoke()
                    state = PairingCeremonyIosState.WaitAck
                    emptyList()
                }
            state == PairingCeremonyIosState.WaitAck && ev == PairingCeremonyProtocol.EventID.RecvPairHelloAck ->
                run {
                    actions[PairingCeremonyProtocol.ActionID.DeriveSecret]?.invoke()
                    // received_server_pub: recv_msg.pubkey (set by action)
                    // client_shared_key: DeriveKey("client_pub", recv_msg.pubkey) (set by action)
                    state = PairingCeremonyIosState.E2EReady
                    emptyList()
                }
            state == PairingCeremonyIosState.E2EReady && ev == PairingCeremonyProtocol.EventID.RecvPairConfirm ->
                run {
                    // ios_code: DeriveCode(received_server_pub, "client_pub") (set by action)
                    state = PairingCeremonyIosState.ShowCode
                    emptyList()
                }
            state == PairingCeremonyIosState.ShowCode && ev == PairingCeremonyProtocol.EventID.CodeDisplayed ->
                run {
                    state = PairingCeremonyIosState.WaitPairComplete
                    emptyList()
                }
            state == PairingCeremonyIosState.WaitPairComplete && ev == PairingCeremonyProtocol.EventID.RecvPairComplete ->
                run {
                    actions[PairingCeremonyProtocol.ActionID.StoreSecret]?.invoke()
                    state = PairingCeremonyIosState.Paired
                    emptyList()
                }
            state == PairingCeremonyIosState.Paired && ev == PairingCeremonyProtocol.EventID.AppLaunch ->
                run {
                    state = PairingCeremonyIosState.Reconnect
                    emptyList()
                }
            state == PairingCeremonyIosState.Reconnect && ev == PairingCeremonyProtocol.EventID.RelayConnected ->
                run {
                    state = PairingCeremonyIosState.SendAuth
                    emptyList()
                }
            state == PairingCeremonyIosState.SendAuth && ev == PairingCeremonyProtocol.EventID.RecvAuthOk ->
                run {
                    state = PairingCeremonyIosState.SessionActive
                    emptyList()
                }
            state == PairingCeremonyIosState.SessionActive && ev == PairingCeremonyProtocol.EventID.Disconnect ->
                run {
                    state = PairingCeremonyIosState.Paired
                    emptyList()
                }
            else -> emptyList()
        }
        return cmds
    }
}

/** PairingCeremonyCliMachine is the generated state machine for the cli actor. */
class PairingCeremonyCliMachine {
    var state: PairingCeremonyCliState = PairingCeremonyCliState.Idle
        private set
    val guards = mutableMapOf<PairingCeremonyProtocol.GuardID, () -> Boolean>()
    val actions = mutableMapOf<PairingCeremonyProtocol.ActionID, () -> Unit>()

    /** Handle an event and return the list of commands to execute. */
    fun handleEvent(ev: PairingCeremonyProtocol.EventID): List<String> {
        val cmds: List<String> = when {
            state == PairingCeremonyCliState.Idle && ev == PairingCeremonyProtocol.EventID.CliInit ->
                run {
                    state = PairingCeremonyCliState.GetKey
                    emptyList()
                }
            state == PairingCeremonyCliState.GetKey && ev == PairingCeremonyProtocol.EventID.KeyStored ->
                run {
                    state = PairingCeremonyCliState.BeginPair
                    emptyList()
                }
            state == PairingCeremonyCliState.BeginPair && ev == PairingCeremonyProtocol.EventID.RecvTokenResponse ->
                run {
                    state = PairingCeremonyCliState.ShowQR
                    emptyList()
                }
            state == PairingCeremonyCliState.ShowQR && ev == PairingCeremonyProtocol.EventID.RecvWaitingForCode ->
                run {
                    state = PairingCeremonyCliState.PromptCode
                    emptyList()
                }
            state == PairingCeremonyCliState.PromptCode && ev == PairingCeremonyProtocol.EventID.UserEntersCode ->
                run {
                    state = PairingCeremonyCliState.SubmitCode
                    emptyList()
                }
            state == PairingCeremonyCliState.SubmitCode && ev == PairingCeremonyProtocol.EventID.RecvPairStatus ->
                run {
                    state = PairingCeremonyCliState.Done
                    emptyList()
                }
            else -> emptyList()
        }
        return cmds
    }
}


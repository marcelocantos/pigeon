// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Auto-generated from protocol definition. Do not edit.
// Source of truth: protocol/*.yaml

package com.marcelocantos.tern.crypto

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

enum class ServerState(val value: String) {
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

enum class IosState(val value: String) {
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

enum class CliState(val value: String) {
    Idle("Idle"),
    GetKey("GetKey"),
    BeginPair("BeginPair"),
    ShowQR("ShowQR"),
    PromptCode("PromptCode"),
    SubmitCode("SubmitCode"),
    Done("Done");
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
    val initial = ServerState.Idle

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
    val initial = IosState.Idle

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
    val initial = CliState.Idle

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

/** ServerMachine is the generated state machine for the server actor. */
class ServerMachine {
    var state: ServerState = ServerState.Idle
        private set
    var currentToken: String = "none" // pairing token currently in play
    var activeTokens: String = "" // set of valid (non-revoked) tokens
    var usedTokens: String = "" // set of revoked tokens
    var serverEcdhPub: String = "none" // server ECDH public key
    var receivedClientPub: String = "none" // pubkey server received in pair_hello (may be adversary's)
    var serverSharedKey: String = "" // ECDH key derived by server (tuple to match DeriveKey output type)
    var serverCode: String = "" // code computed by server from its view of the pubkeys (tuple to match DeriveCode output type)
    var receivedCode: String = "" // code received in code_submit (tuple to match DeriveCode output type)
    var codeAttempts: String = 0 // failed code submission attempts
    var deviceSecret: String = "none" // persistent device secret
    var pairedDevices: String = "" // device IDs that completed pairing
    var receivedDeviceId: String = "none" // device_id from auth_request
    var authNoncesUsed: String = "" // set of consumed auth nonces
    var receivedAuthNonce: String = "none" // nonce from auth_request
    val guards = mutableMapOf<GuardID, () -> Boolean>()
    val actions = mutableMapOf<ActionID, () -> Unit>()

    /** Handle an event and return the list of commands to execute. */
    fun handleEvent(ev: EventID): List<CmdID> {
        val cmds = when {
            state == ServerState.Idle && ev == EventID.RecvPairBegin ->
                run {
                    actions[ActionID.GenerateToken]?.invoke()
                    currentToken = "tok_1"
                    // active_tokens: active_tokens \union {"tok_1"} (set by action)
                    state = ServerState.GenerateToken
                    emptyList()
                }
            state == ServerState.GenerateToken && ev == EventID.TokenCreated ->
                run {
                    actions[ActionID.RegisterRelay]?.invoke()
                    state = ServerState.RegisterRelay
                    emptyList()
                }
            state == ServerState.RegisterRelay && ev == EventID.RelayRegistered ->
                run {
                    state = ServerState.WaitingForClient
                    emptyList()
                }
            state == ServerState.WaitingForClient && ev == EventID.RecvPairHello && guards[GuardID.TokenValid]?.invoke() == true ->
                run {
                    actions[ActionID.DeriveSecret]?.invoke()
                    // received_client_pub: recv_msg.pubkey (set by action)
                    serverEcdhPub = "server_pub"
                    // server_shared_key: DeriveKey("server_pub", recv_msg.pubkey) (set by action)
                    // server_code: DeriveCode("server_pub", recv_msg.pubkey) (set by action)
                    state = ServerState.DeriveSecret
                    emptyList()
                }
            state == ServerState.WaitingForClient && ev == EventID.RecvPairHello && guards[GuardID.TokenInvalid]?.invoke() == true ->
                run {
                    state = ServerState.Idle
                    emptyList()
                }
            state == ServerState.DeriveSecret && ev == EventID.ECDHComplete ->
                run {
                    state = ServerState.SendAck
                    emptyList()
                }
            state == ServerState.SendAck && ev == EventID.SignalCodeDisplay ->
                run {
                    state = ServerState.WaitingForCode
                    emptyList()
                }
            state == ServerState.WaitingForCode && ev == EventID.RecvCodeSubmit ->
                run {
                    // received_code: recv_msg.code (set by action)
                    state = ServerState.ValidateCode
                    emptyList()
                }
            state == ServerState.ValidateCode && ev == EventID.CheckCode && guards[GuardID.CodeCorrect]?.invoke() == true ->
                run {
                    state = ServerState.StorePaired
                    emptyList()
                }
            state == ServerState.ValidateCode && ev == EventID.CheckCode && guards[GuardID.CodeWrong]?.invoke() == true ->
                run {
                    // code_attempts: code_attempts + 1 (set by action)
                    state = ServerState.Idle
                    emptyList()
                }
            state == ServerState.StorePaired && ev == EventID.Finalise ->
                run {
                    actions[ActionID.StoreDevice]?.invoke()
                    deviceSecret = "dev_secret_1"
                    // paired_devices: paired_devices \union {"device_1"} (set by action)
                    // active_tokens: active_tokens \ {current_token} (set by action)
                    // used_tokens: used_tokens \union {current_token} (set by action)
                    state = ServerState.Paired
                    emptyList()
                }
            state == ServerState.Paired && ev == EventID.RecvAuthRequest ->
                run {
                    // received_device_id: recv_msg.device_id (set by action)
                    // received_auth_nonce: recv_msg.nonce (set by action)
                    state = ServerState.AuthCheck
                    emptyList()
                }
            state == ServerState.AuthCheck && ev == EventID.Verify && guards[GuardID.DeviceKnown]?.invoke() == true ->
                run {
                    actions[ActionID.VerifyDevice]?.invoke()
                    // auth_nonces_used: auth_nonces_used \union {received_auth_nonce} (set by action)
                    state = ServerState.SessionActive
                    emptyList()
                }
            state == ServerState.AuthCheck && ev == EventID.Verify && guards[GuardID.DeviceUnknown]?.invoke() == true ->
                run {
                    state = ServerState.Idle
                    emptyList()
                }
            state == ServerState.SessionActive && ev == EventID.Disconnect ->
                run {
                    state = ServerState.Paired
                    emptyList()
                }
            else -> emptyList()
        }
        return cmds
    }
}

/** IosMachine is the generated state machine for the ios actor. */
class IosMachine {
    var state: IosState = IosState.Idle
        private set
    var receivedServerPub: String = "none" // pubkey ios received in pair_hello_ack (may be adversary's)
    var clientSharedKey: String = "" // ECDH key derived by ios (tuple to match DeriveKey output type)
    var iosCode: String = "" // code computed by ios from its view of the pubkeys (tuple to match DeriveCode output type)
    val guards = mutableMapOf<GuardID, () -> Boolean>()
    val actions = mutableMapOf<ActionID, () -> Unit>()

    /** Handle an event and return the list of commands to execute. */
    fun handleEvent(ev: EventID): List<CmdID> {
        val cmds = when {
            state == IosState.Idle && ev == EventID.UserScansQR ->
                run {
                    state = IosState.ScanQR
                    emptyList()
                }
            state == IosState.ScanQR && ev == EventID.QRParsed ->
                run {
                    state = IosState.ConnectRelay
                    emptyList()
                }
            state == IosState.ConnectRelay && ev == EventID.RelayConnected ->
                run {
                    state = IosState.GenKeyPair
                    emptyList()
                }
            state == IosState.GenKeyPair && ev == EventID.KeyPairGenerated ->
                run {
                    actions[ActionID.SendPairHello]?.invoke()
                    state = IosState.WaitAck
                    emptyList()
                }
            state == IosState.WaitAck && ev == EventID.RecvPairHelloAck ->
                run {
                    actions[ActionID.DeriveSecret]?.invoke()
                    // received_server_pub: recv_msg.pubkey (set by action)
                    // client_shared_key: DeriveKey("client_pub", recv_msg.pubkey) (set by action)
                    state = IosState.E2EReady
                    emptyList()
                }
            state == IosState.E2EReady && ev == EventID.RecvPairConfirm ->
                run {
                    // ios_code: DeriveCode(received_server_pub, "client_pub") (set by action)
                    state = IosState.ShowCode
                    emptyList()
                }
            state == IosState.ShowCode && ev == EventID.CodeDisplayed ->
                run {
                    state = IosState.WaitPairComplete
                    emptyList()
                }
            state == IosState.WaitPairComplete && ev == EventID.RecvPairComplete ->
                run {
                    actions[ActionID.StoreSecret]?.invoke()
                    state = IosState.Paired
                    emptyList()
                }
            state == IosState.Paired && ev == EventID.AppLaunch ->
                run {
                    state = IosState.Reconnect
                    emptyList()
                }
            state == IosState.Reconnect && ev == EventID.RelayConnected ->
                run {
                    state = IosState.SendAuth
                    emptyList()
                }
            state == IosState.SendAuth && ev == EventID.RecvAuthOk ->
                run {
                    state = IosState.SessionActive
                    emptyList()
                }
            state == IosState.SessionActive && ev == EventID.Disconnect ->
                run {
                    state = IosState.Paired
                    emptyList()
                }
            else -> emptyList()
        }
        return cmds
    }
}

/** CliMachine is the generated state machine for the cli actor. */
class CliMachine {
    var state: CliState = CliState.Idle
        private set
    val guards = mutableMapOf<GuardID, () -> Boolean>()
    val actions = mutableMapOf<ActionID, () -> Unit>()

    /** Handle an event and return the list of commands to execute. */
    fun handleEvent(ev: EventID): List<CmdID> {
        val cmds = when {
            state == CliState.Idle && ev == EventID.CliInit ->
                run {
                    state = CliState.GetKey
                    emptyList()
                }
            state == CliState.GetKey && ev == EventID.KeyStored ->
                run {
                    state = CliState.BeginPair
                    emptyList()
                }
            state == CliState.BeginPair && ev == EventID.RecvTokenResponse ->
                run {
                    state = CliState.ShowQR
                    emptyList()
                }
            state == CliState.ShowQR && ev == EventID.RecvWaitingForCode ->
                run {
                    state = CliState.PromptCode
                    emptyList()
                }
            state == CliState.PromptCode && ev == EventID.UserEntersCode ->
                run {
                    state = CliState.SubmitCode
                    emptyList()
                }
            state == CliState.SubmitCode && ev == EventID.RecvPairStatus ->
                run {
                    state = CliState.Done
                    emptyList()
                }
            else -> emptyList()
        }
        return cmds
    }
}


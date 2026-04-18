// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Auto-generated from protocol definition. Do not edit.
// Source of truth: protocol/*.yaml

package com.marcelocantos.pigeon.crypto

enum class PairingCeremonyServerPairingState(val value: String) {
    Idle("Idle"),
    GenerateToken("GenerateToken"),
    RegisterRelay("RegisterRelay"),
    WaitingForClient("WaitingForClient"),
    DeriveSecret("DeriveSecret"),
    SendAck("SendAck"),
    WaitingForCode("WaitingForCode"),
    ValidateCode("ValidateCode"),
    StorePaired("StorePaired"),
    PairingComplete("PairingComplete");
}

enum class PairingCeremonyServerAuthState(val value: String) {
    Idle("Idle"),
    Paired("Paired"),
    AuthCheck("AuthCheck"),
    SessionActive("SessionActive");
}

enum class PairingCeremonyIosPairingState(val value: String) {
    Idle("Idle"),
    ScanQR("ScanQR"),
    ConnectRelay("ConnectRelay"),
    GenKeyPair("GenKeyPair"),
    WaitAck("WaitAck"),
    E2EReady("E2EReady"),
    ShowCode("ShowCode"),
    WaitPairComplete("WaitPairComplete"),
    PairingComplete("PairingComplete");
}

enum class PairingCeremonyIosAuthState(val value: String) {
    Idle("Idle"),
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
        CredentialReady("credential_ready"),
        Disconnect("disconnect"),
        Finalise("finalise"),
        KeyPairGenerated("key pair generated"),
        KeyStored("key stored"),
        Paired("paired"),
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

    /** server/pairing transition table. */
    object ServerPairingTable {
        val initial = PairingCeremonyServerPairingState.Idle

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
            Transition("StorePaired", "PairingComplete", "finalise", "internal", null, "store_device", listOf("ios" to "pair_complete", "cli" to "pair_status")),
        )
    }

    /** server/auth transition table. */
    object ServerAuthTable {
        val initial = PairingCeremonyServerAuthState.Idle

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
            Transition("Idle", "Paired", "credential_ready", "internal", null, null, emptyList()),
            Transition("Paired", "AuthCheck", "auth_request", "recv", null, null, emptyList()),
            Transition("AuthCheck", "SessionActive", "verify", "internal", "device_known", "verify_device", listOf("ios" to "auth_ok")),
            Transition("AuthCheck", "Idle", "verify", "internal", "device_unknown", null, emptyList()),
            Transition("SessionActive", "Paired", "disconnect", "internal", null, null, emptyList()),
        )
    }

    /** ios/pairing transition table. */
    object IosPairingTable {
        val initial = PairingCeremonyIosPairingState.Idle

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
            Transition("WaitPairComplete", "PairingComplete", "pair_complete", "recv", null, "store_secret", emptyList()),
        )
    }

    /** ios/auth transition table. */
    object IosAuthTable {
        val initial = PairingCeremonyIosAuthState.Idle

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
            Transition("Idle", "Paired", "credential_ready", "internal", null, null, emptyList()),
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

/** PairingCeremonyServerPairingMachine is the generated state machine for server/pairing. */
class PairingCeremonyServerPairingMachine {
    var state: PairingCeremonyServerPairingState = PairingCeremonyServerPairingState.Idle
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
    val guards = mutableMapOf<PairingCeremonyProtocol.GuardID, () -> Boolean>()
    val actions = mutableMapOf<PairingCeremonyProtocol.ActionID, () -> Unit>()

    /** Handle an event and return the list of commands to execute. */
    fun handleEvent(ev: PairingCeremonyProtocol.EventID): List<String> {
        val cmds: List<String> = when {
            state == PairingCeremonyServerPairingState.Idle && ev == PairingCeremonyProtocol.EventID.RecvPairBegin ->
                run {
                    actions[PairingCeremonyProtocol.ActionID.GenerateToken]?.invoke()
                    currentToken = "tok_1"
                    // active_tokens: active_tokens \union {"tok_1"} (set by action)
                    state = PairingCeremonyServerPairingState.GenerateToken
                    emptyList()
                }
            state == PairingCeremonyServerPairingState.GenerateToken && ev == PairingCeremonyProtocol.EventID.TokenCreated ->
                run {
                    actions[PairingCeremonyProtocol.ActionID.RegisterRelay]?.invoke()
                    state = PairingCeremonyServerPairingState.RegisterRelay
                    emptyList()
                }
            state == PairingCeremonyServerPairingState.RegisterRelay && ev == PairingCeremonyProtocol.EventID.RelayRegistered ->
                run {
                    state = PairingCeremonyServerPairingState.WaitingForClient
                    emptyList()
                }
            state == PairingCeremonyServerPairingState.WaitingForClient && ev == PairingCeremonyProtocol.EventID.RecvPairHello && guards[PairingCeremonyProtocol.GuardID.TokenValid]?.invoke() == true ->
                run {
                    actions[PairingCeremonyProtocol.ActionID.DeriveSecret]?.invoke()
                    // received_client_pub: recv_msg.pubkey (set by action)
                    serverEcdhPub = "server_pub"
                    // server_shared_key: DeriveKey("server_pub", recv_msg.pubkey) (set by action)
                    // server_code: DeriveCode("server_pub", recv_msg.pubkey) (set by action)
                    // active_tokens: active_tokens \ {current_token} (set by action)
                    // used_tokens: used_tokens \union {current_token} (set by action)
                    state = PairingCeremonyServerPairingState.DeriveSecret
                    emptyList()
                }
            state == PairingCeremonyServerPairingState.WaitingForClient && ev == PairingCeremonyProtocol.EventID.RecvPairHello && guards[PairingCeremonyProtocol.GuardID.TokenInvalid]?.invoke() == true ->
                run {
                    state = PairingCeremonyServerPairingState.Idle
                    emptyList()
                }
            state == PairingCeremonyServerPairingState.DeriveSecret && ev == PairingCeremonyProtocol.EventID.ECDHComplete ->
                run {
                    state = PairingCeremonyServerPairingState.SendAck
                    emptyList()
                }
            state == PairingCeremonyServerPairingState.SendAck && ev == PairingCeremonyProtocol.EventID.SignalCodeDisplay ->
                run {
                    state = PairingCeremonyServerPairingState.WaitingForCode
                    emptyList()
                }
            state == PairingCeremonyServerPairingState.WaitingForCode && ev == PairingCeremonyProtocol.EventID.RecvCodeSubmit ->
                run {
                    // received_code: recv_msg.code (set by action)
                    state = PairingCeremonyServerPairingState.ValidateCode
                    emptyList()
                }
            state == PairingCeremonyServerPairingState.ValidateCode && ev == PairingCeremonyProtocol.EventID.CheckCode && guards[PairingCeremonyProtocol.GuardID.CodeCorrect]?.invoke() == true ->
                run {
                    state = PairingCeremonyServerPairingState.StorePaired
                    emptyList()
                }
            state == PairingCeremonyServerPairingState.ValidateCode && ev == PairingCeremonyProtocol.EventID.CheckCode && guards[PairingCeremonyProtocol.GuardID.CodeWrong]?.invoke() == true ->
                run {
                    // code_attempts: code_attempts + 1 (set by action)
                    state = PairingCeremonyServerPairingState.Idle
                    emptyList()
                }
            state == PairingCeremonyServerPairingState.StorePaired && ev == PairingCeremonyProtocol.EventID.Finalise ->
                run {
                    actions[PairingCeremonyProtocol.ActionID.StoreDevice]?.invoke()
                    deviceSecret = "dev_secret_1"
                    // paired_devices: paired_devices \union {"device_1"} (set by action)
                    state = PairingCeremonyServerPairingState.PairingComplete
                    emptyList()
                }
            else -> emptyList()
        }
        return cmds
    }
}

/** PairingCeremonyServerAuthMachine is the generated state machine for server/auth. */
class PairingCeremonyServerAuthMachine {
    var state: PairingCeremonyServerAuthState = PairingCeremonyServerAuthState.Idle
        private set
    var receivedDeviceId: String = "none" // device_id from auth_request
    var authNoncesUsed: String = "" // set of consumed auth nonces
    var receivedAuthNonce: String = "none" // nonce from auth_request
    val guards = mutableMapOf<PairingCeremonyProtocol.GuardID, () -> Boolean>()
    val actions = mutableMapOf<PairingCeremonyProtocol.ActionID, () -> Unit>()

    /** Handle an event and return the list of commands to execute. */
    fun handleEvent(ev: PairingCeremonyProtocol.EventID): List<String> {
        val cmds: List<String> = when {
            state == PairingCeremonyServerAuthState.Idle && ev == PairingCeremonyProtocol.EventID.CredentialReady ->
                run {
                    state = PairingCeremonyServerAuthState.Paired
                    emptyList()
                }
            state == PairingCeremonyServerAuthState.Paired && ev == PairingCeremonyProtocol.EventID.RecvAuthRequest ->
                run {
                    // received_device_id: recv_msg.device_id (set by action)
                    // received_auth_nonce: recv_msg.nonce (set by action)
                    state = PairingCeremonyServerAuthState.AuthCheck
                    emptyList()
                }
            state == PairingCeremonyServerAuthState.AuthCheck && ev == PairingCeremonyProtocol.EventID.Verify && guards[PairingCeremonyProtocol.GuardID.DeviceKnown]?.invoke() == true ->
                run {
                    actions[PairingCeremonyProtocol.ActionID.VerifyDevice]?.invoke()
                    // auth_nonces_used: auth_nonces_used \union {received_auth_nonce} (set by action)
                    state = PairingCeremonyServerAuthState.SessionActive
                    emptyList()
                }
            state == PairingCeremonyServerAuthState.AuthCheck && ev == PairingCeremonyProtocol.EventID.Verify && guards[PairingCeremonyProtocol.GuardID.DeviceUnknown]?.invoke() == true ->
                run {
                    state = PairingCeremonyServerAuthState.Idle
                    emptyList()
                }
            state == PairingCeremonyServerAuthState.SessionActive && ev == PairingCeremonyProtocol.EventID.Disconnect ->
                run {
                    state = PairingCeremonyServerAuthState.Paired
                    emptyList()
                }
            else -> emptyList()
        }
        return cmds
    }
}

/** PairingCeremonyServerComposite holds all sub-machines for the server actor. */
class PairingCeremonyServerComposite {
    val pairing = PairingCeremonyServerPairingMachine()
    val auth = PairingCeremonyServerAuthMachine()

    /** Route dispatches inter-machine events according to the routing table. */
    fun route(from: String, event: PairingCeremonyProtocol.EventID) {
        when {
            from == "pairing" && event == PairingCeremonyProtocol.EventID.Paired -> {
                auth.handleEvent(PairingCeremonyProtocol.EventID.CredentialReady)
            }
        }
    }
}

/** PairingCeremonyIosPairingMachine is the generated state machine for ios/pairing. */
class PairingCeremonyIosPairingMachine {
    var state: PairingCeremonyIosPairingState = PairingCeremonyIosPairingState.Idle
        private set
    var receivedServerPub: String = "none" // pubkey ios received in pair_hello_ack (may be adversary's)
    var clientSharedKey: String = "" // ECDH key derived by ios (tuple to match DeriveKey output type)
    var iosCode: String = "" // code computed by ios from its view of the pubkeys (tuple to match DeriveCode output type)
    val guards = mutableMapOf<PairingCeremonyProtocol.GuardID, () -> Boolean>()
    val actions = mutableMapOf<PairingCeremonyProtocol.ActionID, () -> Unit>()

    /** Handle an event and return the list of commands to execute. */
    fun handleEvent(ev: PairingCeremonyProtocol.EventID): List<String> {
        val cmds: List<String> = when {
            state == PairingCeremonyIosPairingState.Idle && ev == PairingCeremonyProtocol.EventID.UserScansQR ->
                run {
                    state = PairingCeremonyIosPairingState.ScanQR
                    emptyList()
                }
            state == PairingCeremonyIosPairingState.ScanQR && ev == PairingCeremonyProtocol.EventID.QRParsed ->
                run {
                    state = PairingCeremonyIosPairingState.ConnectRelay
                    emptyList()
                }
            state == PairingCeremonyIosPairingState.ConnectRelay && ev == PairingCeremonyProtocol.EventID.RelayConnected ->
                run {
                    state = PairingCeremonyIosPairingState.GenKeyPair
                    emptyList()
                }
            state == PairingCeremonyIosPairingState.GenKeyPair && ev == PairingCeremonyProtocol.EventID.KeyPairGenerated ->
                run {
                    actions[PairingCeremonyProtocol.ActionID.SendPairHello]?.invoke()
                    state = PairingCeremonyIosPairingState.WaitAck
                    emptyList()
                }
            state == PairingCeremonyIosPairingState.WaitAck && ev == PairingCeremonyProtocol.EventID.RecvPairHelloAck ->
                run {
                    actions[PairingCeremonyProtocol.ActionID.DeriveSecret]?.invoke()
                    // received_server_pub: recv_msg.pubkey (set by action)
                    // client_shared_key: DeriveKey("client_pub", recv_msg.pubkey) (set by action)
                    state = PairingCeremonyIosPairingState.E2EReady
                    emptyList()
                }
            state == PairingCeremonyIosPairingState.E2EReady && ev == PairingCeremonyProtocol.EventID.RecvPairConfirm ->
                run {
                    // ios_code: DeriveCode(received_server_pub, "client_pub") (set by action)
                    state = PairingCeremonyIosPairingState.ShowCode
                    emptyList()
                }
            state == PairingCeremonyIosPairingState.ShowCode && ev == PairingCeremonyProtocol.EventID.CodeDisplayed ->
                run {
                    state = PairingCeremonyIosPairingState.WaitPairComplete
                    emptyList()
                }
            state == PairingCeremonyIosPairingState.WaitPairComplete && ev == PairingCeremonyProtocol.EventID.RecvPairComplete ->
                run {
                    actions[PairingCeremonyProtocol.ActionID.StoreSecret]?.invoke()
                    state = PairingCeremonyIosPairingState.PairingComplete
                    emptyList()
                }
            else -> emptyList()
        }
        return cmds
    }
}

/** PairingCeremonyIosAuthMachine is the generated state machine for ios/auth. */
class PairingCeremonyIosAuthMachine {
    var state: PairingCeremonyIosAuthState = PairingCeremonyIosAuthState.Idle
        private set
    val guards = mutableMapOf<PairingCeremonyProtocol.GuardID, () -> Boolean>()
    val actions = mutableMapOf<PairingCeremonyProtocol.ActionID, () -> Unit>()

    /** Handle an event and return the list of commands to execute. */
    fun handleEvent(ev: PairingCeremonyProtocol.EventID): List<String> {
        val cmds: List<String> = when {
            state == PairingCeremonyIosAuthState.Idle && ev == PairingCeremonyProtocol.EventID.CredentialReady ->
                run {
                    state = PairingCeremonyIosAuthState.Paired
                    emptyList()
                }
            state == PairingCeremonyIosAuthState.Paired && ev == PairingCeremonyProtocol.EventID.AppLaunch ->
                run {
                    state = PairingCeremonyIosAuthState.Reconnect
                    emptyList()
                }
            state == PairingCeremonyIosAuthState.Reconnect && ev == PairingCeremonyProtocol.EventID.RelayConnected ->
                run {
                    state = PairingCeremonyIosAuthState.SendAuth
                    emptyList()
                }
            state == PairingCeremonyIosAuthState.SendAuth && ev == PairingCeremonyProtocol.EventID.RecvAuthOk ->
                run {
                    state = PairingCeremonyIosAuthState.SessionActive
                    emptyList()
                }
            state == PairingCeremonyIosAuthState.SessionActive && ev == PairingCeremonyProtocol.EventID.Disconnect ->
                run {
                    state = PairingCeremonyIosAuthState.Paired
                    emptyList()
                }
            else -> emptyList()
        }
        return cmds
    }
}

/** PairingCeremonyIosComposite holds all sub-machines for the ios actor. */
class PairingCeremonyIosComposite {
    val pairing = PairingCeremonyIosPairingMachine()
    val auth = PairingCeremonyIosAuthMachine()

    /** Route dispatches inter-machine events according to the routing table. */
    fun route(from: String, event: PairingCeremonyProtocol.EventID) {
        when {
            from == "pairing" && event == PairingCeremonyProtocol.EventID.Paired -> {
                auth.handleEvent(PairingCeremonyProtocol.EventID.CredentialReady)
            }
        }
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


// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Auto-generated from protocol definition. Do not edit.
// Source of truth: protocol/*.yaml

package com.marcelocantos.pigeon.crypto

enum class BackendState(val value: String) {
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
    SessionActive("SessionActive"),
    RelayConnected("RelayConnected"),
    LANOffered("LANOffered"),
    LANActive("LANActive"),
    RelayBackoff("RelayBackoff"),
    LANDegraded("LANDegraded");
}

enum class ClientState(val value: String) {
    Idle("Idle"),
    ObtainBackchannelSecret("ObtainBackchannelSecret"),
    ConnectRelay("ConnectRelay"),
    GenKeyPair("GenKeyPair"),
    WaitAck("WaitAck"),
    E2EReady("E2EReady"),
    ShowCode("ShowCode"),
    WaitPairComplete("WaitPairComplete"),
    Paired("Paired"),
    Reconnect("Reconnect"),
    SendAuth("SendAuth"),
    SessionActive("SessionActive"),
    RelayConnected("RelayConnected"),
    LANConnecting("LANConnecting"),
    LANVerifying("LANVerifying"),
    LANActive("LANActive"),
    RelayFallback("RelayFallback");
}

enum class RelayState(val value: String) {
    Idle("Idle"),
    BackendRegistered("BackendRegistered"),
    Bridged("Bridged");
}

/** The protocol transition table and shared type enums. */
object SessionProtocol {

    enum class MessageType(val value: String) {
        PairHello("pair_hello"),
        PairHelloAck("pair_hello_ack"),
        PairConfirm("pair_confirm"),
        PairComplete("pair_complete"),
        AuthRequest("auth_request"),
        AuthOk("auth_ok"),
        LanOffer("lan_offer"),
        LanVerify("lan_verify"),
        LanConfirm("lan_confirm"),
        PathPing("path_ping"),
        PathPong("path_pong");
    }

    enum class GuardID(val value: String) {
        TokenValid("token_valid"),
        TokenInvalid("token_invalid"),
        CodeCorrect("code_correct"),
        CodeWrong("code_wrong"),
        DeviceKnown("device_known"),
        DeviceUnknown("device_unknown"),
        NonceFresh("nonce_fresh"),
        ChallengeValid("challenge_valid"),
        ChallengeInvalid("challenge_invalid"),
        LanEnabled("lan_enabled"),
        LanDisabled("lan_disabled"),
        LanServerAvailable("lan_server_available"),
        UnderMaxFailures("under_max_failures"),
        AtMaxFailures("at_max_failures");
    }

    enum class ActionID(val value: String) {
        GenerateToken("generate_token"),
        RegisterRelay("register_relay"),
        DeriveSecret("derive_secret"),
        StoreDevice("store_device"),
        VerifyDevice("verify_device"),
        ActivateLan("activate_lan"),
        FallbackToRelay("fallback_to_relay"),
        ResetFailures("reset_failures"),
        SendPairHello("send_pair_hello"),
        StoreSecret("store_secret"),
        DialLan("dial_lan"),
        BridgeStreams("bridge_streams"),
        Unbridge("unbridge");
    }

    enum class EventID(val value: String) {
        AppClose("app_close"),
        AppForceFallback("app_force_fallback"),
        AppLaunch("app_launch"),
        AppRecv("app_recv"),
        AppRecvDatagram("app_recv_datagram"),
        AppSend("app_send"),
        AppSendDatagram("app_send_datagram"),
        BackchannelReceived("backchannel_received"),
        BackendDisconnect("backend_disconnect"),
        BackendRegister("backend_register"),
        BackoffExpired("backoff_expired"),
        CheckCode("check_code"),
        CliCodeEntered("cli_code_entered"),
        CliInitPair("cli_init_pair"),
        ClientConnect("client_connect"),
        ClientDisconnect("client_disconnect"),
        CodeDisplayed("code_displayed"),
        Disconnect("disconnect"),
        EcdhComplete("ecdh_complete"),
        Finalise("finalise"),
        KeyPairGenerated("key_pair_generated"),
        LanDatagram("lan_datagram"),
        LanDialFailed("lan_dial_failed"),
        LanDialOk("lan_dial_ok"),
        LanError("lan_error"),
        LanServerChanged("lan_server_changed"),
        LanServerReady("lan_server_ready"),
        LanStreamData("lan_stream_data"),
        LanStreamError("lan_stream_error"),
        LanVerifyOk("lan_verify_ok"),
        OfferTimeout("offer_timeout"),
        PingTick("ping_tick"),
        PingTimeout("ping_timeout"),
        ReadvertiseTick("readvertise_tick"),
        RecvAuthOk("recv_auth_ok"),
        RecvAuthRequest("recv_auth_request"),
        RecvLanConfirm("recv_lan_confirm"),
        RecvLanOffer("recv_lan_offer"),
        RecvLanVerify("recv_lan_verify"),
        RecvPairComplete("recv_pair_complete"),
        RecvPairConfirm("recv_pair_confirm"),
        RecvPairHello("recv_pair_hello"),
        RecvPairHelloAck("recv_pair_hello_ack"),
        RecvPathPing("recv_path_ping"),
        RecvPathPong("recv_path_pong"),
        RelayConnected("relay_connected"),
        RelayDatagram("relay_datagram"),
        RelayOk("relay_ok"),
        RelayRegistered("relay_registered"),
        RelayStreamData("relay_stream_data"),
        RelayStreamError("relay_stream_error"),
        SecretParsed("secret_parsed"),
        SessionEstablished("session_established"),
        SignalCodeDisplay("signal_code_display"),
        TokenCreated("token_created"),
        Verify("verify"),
        VerifyTimeout("verify_timeout");
    }

    enum class CmdID(val value: String) {
        WriteActiveStream("write_active_stream"),
        SendActiveDatagram("send_active_datagram"),
        SendPathPing("send_path_ping"),
        SendPathPong("send_path_pong"),
        SendLanOffer("send_lan_offer"),
        SendLanVerify("send_lan_verify"),
        SendLanConfirm("send_lan_confirm"),
        DialLan("dial_lan"),
        DeliverRecv("deliver_recv"),
        DeliverRecvError("deliver_recv_error"),
        DeliverRecvDatagram("deliver_recv_datagram"),
        StartLanStreamReader("start_lan_stream_reader"),
        StopLanStreamReader("stop_lan_stream_reader"),
        StartLanDgReader("start_lan_dg_reader"),
        StopLanDgReader("stop_lan_dg_reader"),
        StartMonitor("start_monitor"),
        StopMonitor("stop_monitor"),
        StartPongTimeout("start_pong_timeout"),
        CancelPongTimeout("cancel_pong_timeout"),
        StartBackoffTimer("start_backoff_timer"),
        CloseLanPath("close_lan_path"),
        SignalLanReady("signal_lan_ready"),
        ResetLanReady("reset_lan_ready"),
        SetCryptoDatagram("set_crypto_datagram");
    }

    /** Protocol wire constants shared across all platforms. */
    object Wire {
        const val DG_CONN_WHOLE: Byte = 0x00.toByte()
        const val DG_PING: Byte = 0x10.toByte()
        const val DG_PONG: Byte = 0x11.toByte()
        const val DG_CONN_FRAGMENT: Byte = 0x40.toByte()
        const val DG_CHAN_WHOLE: Byte = 0x80.toByte()
        const val DG_CHAN_FRAGMENT: Byte = 0xC0.toByte()
        const val FRAG_HEADER_SIZE = 8
        const val CHAN_ID_SIZE = 2
        const val MAX_DATAGRAM_PAYLOAD = 1200
        const val FRAGMENT_TIMEOUT_MS = 5000L // ms
        const val FRAME_APP: Byte = 0x00.toByte()
        const val FRAME_LAN_OFFER: Byte = 0x01.toByte()
        const val FRAME_CUTOVER: Byte = 0x02.toByte()
        const val MAX_MESSAGE_SIZE = 1048576
        const val LENGTH_PREFIX_SIZE = 4
        const val PING_INTERVAL_MS = 5000L // ms
        const val PONG_TIMEOUT_MS = 4000L // ms
        const val MAX_PING_FAILURES = 3
        const val MAX_BACKOFF_LEVEL = 5
        const val STREAM_CHANNEL_OPENER_SUFFIX = ":o2a"
        const val STREAM_CHANNEL_ACCEPT_SUFFIX = ":a2o"
        const val DG_CHANNEL_SEND_SUFFIX = ":dg:send"
        const val DG_CHANNEL_RECV_SUFFIX = ":dg:recv"
        const val CHANNEL_ID_HASH_MULTIPLIER = 31
    }

    /** backend transition table. */
    object BackendTable {
        val initial = BackendState.Idle

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
            Transition("Idle", "GenerateToken", "cli_init_pair", "internal", null, "generate_token", emptyList()),
            Transition("GenerateToken", "RegisterRelay", "token_created", "internal", null, "register_relay", emptyList()),
            Transition("RegisterRelay", "WaitingForClient", "relay_registered", "internal", null, null, emptyList()),
            Transition("WaitingForClient", "DeriveSecret", "pair_hello", "recv", "token_valid", "derive_secret", emptyList()),
            Transition("WaitingForClient", "Idle", "pair_hello", "recv", "token_invalid", null, emptyList()),
            Transition("DeriveSecret", "SendAck", "ecdh_complete", "internal", null, null, listOf("client" to "pair_hello_ack")),
            Transition("SendAck", "WaitingForCode", "signal_code_display", "internal", null, null, listOf("client" to "pair_confirm")),
            Transition("WaitingForCode", "ValidateCode", "cli_code_entered", "internal", null, null, emptyList()),
            Transition("ValidateCode", "StorePaired", "check_code", "internal", "code_correct", null, emptyList()),
            Transition("ValidateCode", "Idle", "check_code", "internal", "code_wrong", null, emptyList()),
            Transition("StorePaired", "Paired", "finalise", "internal", null, "store_device", listOf("client" to "pair_complete")),
            Transition("Paired", "AuthCheck", "auth_request", "recv", null, null, emptyList()),
            Transition("AuthCheck", "SessionActive", "verify", "internal", "device_known", "verify_device", listOf("client" to "auth_ok")),
            Transition("AuthCheck", "Idle", "verify", "internal", "device_unknown", null, emptyList()),
            Transition("SessionActive", "RelayConnected", "session_established", "internal", null, null, emptyList()),
            Transition("RelayConnected", "LANOffered", "lan_server_ready", "internal", null, null, listOf("client" to "lan_offer")),
            Transition("LANOffered", "LANActive", "lan_verify", "recv", "challenge_valid", "activate_lan", listOf("client" to "lan_confirm")),
            Transition("LANOffered", "RelayConnected", "lan_verify", "recv", "challenge_invalid", null, emptyList()),
            Transition("LANOffered", "RelayBackoff", "offer_timeout", "internal", null, null, emptyList()),
            Transition("LANActive", "LANActive", "ping_tick", "internal", null, null, listOf("client" to "path_ping")),
            Transition("LANActive", "LANDegraded", "ping_timeout", "internal", null, null, emptyList()),
            Transition("LANDegraded", "LANDegraded", "ping_tick", "internal", null, null, listOf("client" to "path_ping")),
            Transition("LANActive", "RelayBackoff", "lan_stream_error", "internal", null, "fallback_to_relay", emptyList()),
            Transition("LANDegraded", "RelayBackoff", "lan_stream_error", "internal", null, "fallback_to_relay", emptyList()),
            Transition("LANDegraded", "LANActive", "path_pong", "recv", null, "reset_failures", emptyList()),
            Transition("LANDegraded", "LANDegraded", "ping_timeout", "internal", "under_max_failures", null, emptyList()),
            Transition("LANDegraded", "RelayBackoff", "ping_timeout", "internal", "at_max_failures", "fallback_to_relay", emptyList()),
            Transition("RelayBackoff", "LANOffered", "backoff_expired", "internal", null, null, listOf("client" to "lan_offer")),
            Transition("RelayBackoff", "LANOffered", "lan_server_changed", "internal", null, null, listOf("client" to "lan_offer")),
            Transition("RelayConnected", "LANOffered", "readvertise_tick", "internal", "lan_server_available", null, listOf("client" to "lan_offer")),
            Transition("LANOffered", "RelayConnected", "app_force_fallback", "internal", null, null, emptyList()),
            Transition("LANActive", "RelayBackoff", "app_force_fallback", "internal", null, "fallback_to_relay", emptyList()),
            Transition("LANDegraded", "RelayBackoff", "app_force_fallback", "internal", null, "fallback_to_relay", emptyList()),
            Transition("RelayConnected", "Paired", "disconnect", "internal", null, null, emptyList()),
            Transition("RelayConnected", "RelayConnected", "app_send", "internal", null, null, emptyList()),
            Transition("LANOffered", "LANOffered", "app_send", "internal", null, null, emptyList()),
            Transition("LANActive", "LANActive", "app_send", "internal", null, null, emptyList()),
            Transition("LANDegraded", "LANDegraded", "app_send", "internal", null, null, emptyList()),
            Transition("RelayBackoff", "RelayBackoff", "app_send", "internal", null, null, emptyList()),
            Transition("RelayConnected", "RelayConnected", "relay_stream_data", "internal", null, null, emptyList()),
            Transition("LANOffered", "LANOffered", "relay_stream_data", "internal", null, null, emptyList()),
            Transition("LANActive", "LANActive", "relay_stream_data", "internal", null, null, emptyList()),
            Transition("LANDegraded", "LANDegraded", "relay_stream_data", "internal", null, null, emptyList()),
            Transition("RelayBackoff", "RelayBackoff", "relay_stream_data", "internal", null, null, emptyList()),
            Transition("RelayConnected", "RelayConnected", "relay_stream_error", "internal", null, null, emptyList()),
            Transition("LANOffered", "LANOffered", "relay_stream_error", "internal", null, null, emptyList()),
            Transition("LANActive", "LANActive", "relay_stream_error", "internal", null, null, emptyList()),
            Transition("LANDegraded", "LANDegraded", "relay_stream_error", "internal", null, null, emptyList()),
            Transition("RelayBackoff", "RelayBackoff", "relay_stream_error", "internal", null, null, emptyList()),
            Transition("RelayConnected", "RelayConnected", "app_send_datagram", "internal", null, null, emptyList()),
            Transition("LANOffered", "LANOffered", "app_send_datagram", "internal", null, null, emptyList()),
            Transition("LANActive", "LANActive", "app_send_datagram", "internal", null, null, emptyList()),
            Transition("LANDegraded", "LANDegraded", "app_send_datagram", "internal", null, null, emptyList()),
            Transition("RelayBackoff", "RelayBackoff", "app_send_datagram", "internal", null, null, emptyList()),
            Transition("RelayConnected", "RelayConnected", "relay_datagram", "internal", null, null, emptyList()),
            Transition("LANOffered", "LANOffered", "relay_datagram", "internal", null, null, emptyList()),
            Transition("LANActive", "LANActive", "relay_datagram", "internal", null, null, emptyList()),
            Transition("LANDegraded", "LANDegraded", "relay_datagram", "internal", null, null, emptyList()),
            Transition("RelayBackoff", "RelayBackoff", "relay_datagram", "internal", null, null, emptyList()),
            Transition("LANActive", "LANActive", "lan_stream_data", "internal", null, null, emptyList()),
            Transition("LANDegraded", "LANDegraded", "lan_stream_data", "internal", null, null, emptyList()),
            Transition("LANActive", "LANActive", "lan_datagram", "internal", null, null, emptyList()),
            Transition("LANDegraded", "LANDegraded", "lan_datagram", "internal", null, null, emptyList()),
        )
    }

    /** client transition table. */
    object ClientTable {
        val initial = ClientState.Idle

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
            Transition("Idle", "ObtainBackchannelSecret", "backchannel_received", "internal", null, null, emptyList()),
            Transition("ObtainBackchannelSecret", "ConnectRelay", "secret_parsed", "internal", null, null, emptyList()),
            Transition("ConnectRelay", "GenKeyPair", "relay_connected", "internal", null, null, emptyList()),
            Transition("GenKeyPair", "WaitAck", "key_pair_generated", "internal", null, "send_pair_hello", listOf("backend" to "pair_hello")),
            Transition("WaitAck", "E2EReady", "pair_hello_ack", "recv", null, "derive_secret", emptyList()),
            Transition("E2EReady", "ShowCode", "pair_confirm", "recv", null, null, emptyList()),
            Transition("ShowCode", "WaitPairComplete", "code_displayed", "internal", null, null, emptyList()),
            Transition("WaitPairComplete", "Paired", "pair_complete", "recv", null, "store_secret", emptyList()),
            Transition("Paired", "Reconnect", "app_launch", "internal", null, null, emptyList()),
            Transition("Reconnect", "SendAuth", "relay_connected", "internal", null, null, listOf("backend" to "auth_request")),
            Transition("SendAuth", "SessionActive", "auth_ok", "recv", null, null, emptyList()),
            Transition("SessionActive", "RelayConnected", "session_established", "internal", null, null, emptyList()),
            Transition("RelayConnected", "LANConnecting", "lan_offer", "recv", "lan_enabled", "dial_lan", emptyList()),
            Transition("RelayConnected", "RelayConnected", "lan_offer", "recv", "lan_disabled", null, emptyList()),
            Transition("LANConnecting", "LANVerifying", "lan_dial_ok", "internal", null, null, listOf("backend" to "lan_verify")),
            Transition("LANConnecting", "RelayConnected", "lan_dial_failed", "internal", null, null, emptyList()),
            Transition("LANVerifying", "LANActive", "lan_confirm", "recv", null, "activate_lan", emptyList()),
            Transition("LANVerifying", "RelayConnected", "verify_timeout", "internal", null, null, emptyList()),
            Transition("LANActive", "LANActive", "path_ping", "recv", null, null, listOf("backend" to "path_pong")),
            Transition("LANActive", "RelayFallback", "lan_error", "internal", null, "fallback_to_relay", emptyList()),
            Transition("LANActive", "RelayFallback", "lan_stream_error", "internal", null, "fallback_to_relay", emptyList()),
            Transition("RelayFallback", "RelayConnected", "relay_ok", "internal", null, null, emptyList()),
            Transition("LANActive", "LANConnecting", "lan_offer", "recv", "lan_enabled", "dial_lan", emptyList()),
            Transition("LANConnecting", "RelayConnected", "app_force_fallback", "internal", null, null, emptyList()),
            Transition("LANVerifying", "RelayConnected", "app_force_fallback", "internal", null, null, emptyList()),
            Transition("LANActive", "RelayConnected", "app_force_fallback", "internal", null, "fallback_to_relay", emptyList()),
            Transition("RelayConnected", "Paired", "disconnect", "internal", null, null, emptyList()),
            Transition("RelayConnected", "RelayConnected", "app_send", "internal", null, null, emptyList()),
            Transition("LANConnecting", "LANConnecting", "app_send", "internal", null, null, emptyList()),
            Transition("LANVerifying", "LANVerifying", "app_send", "internal", null, null, emptyList()),
            Transition("LANActive", "LANActive", "app_send", "internal", null, null, emptyList()),
            Transition("RelayFallback", "RelayFallback", "app_send", "internal", null, null, emptyList()),
            Transition("RelayConnected", "RelayConnected", "relay_stream_data", "internal", null, null, emptyList()),
            Transition("LANConnecting", "LANConnecting", "relay_stream_data", "internal", null, null, emptyList()),
            Transition("LANVerifying", "LANVerifying", "relay_stream_data", "internal", null, null, emptyList()),
            Transition("LANActive", "LANActive", "relay_stream_data", "internal", null, null, emptyList()),
            Transition("RelayFallback", "RelayFallback", "relay_stream_data", "internal", null, null, emptyList()),
            Transition("RelayConnected", "RelayConnected", "relay_stream_error", "internal", null, null, emptyList()),
            Transition("LANConnecting", "LANConnecting", "relay_stream_error", "internal", null, null, emptyList()),
            Transition("LANVerifying", "LANVerifying", "relay_stream_error", "internal", null, null, emptyList()),
            Transition("LANActive", "LANActive", "relay_stream_error", "internal", null, null, emptyList()),
            Transition("RelayFallback", "RelayFallback", "relay_stream_error", "internal", null, null, emptyList()),
            Transition("RelayConnected", "RelayConnected", "app_send_datagram", "internal", null, null, emptyList()),
            Transition("LANConnecting", "LANConnecting", "app_send_datagram", "internal", null, null, emptyList()),
            Transition("LANVerifying", "LANVerifying", "app_send_datagram", "internal", null, null, emptyList()),
            Transition("LANActive", "LANActive", "app_send_datagram", "internal", null, null, emptyList()),
            Transition("RelayFallback", "RelayFallback", "app_send_datagram", "internal", null, null, emptyList()),
            Transition("RelayConnected", "RelayConnected", "relay_datagram", "internal", null, null, emptyList()),
            Transition("LANConnecting", "LANConnecting", "relay_datagram", "internal", null, null, emptyList()),
            Transition("LANVerifying", "LANVerifying", "relay_datagram", "internal", null, null, emptyList()),
            Transition("LANActive", "LANActive", "relay_datagram", "internal", null, null, emptyList()),
            Transition("RelayFallback", "RelayFallback", "relay_datagram", "internal", null, null, emptyList()),
            Transition("LANActive", "LANActive", "lan_stream_data", "internal", null, null, emptyList()),
            Transition("LANActive", "LANActive", "lan_datagram", "internal", null, null, emptyList()),
        )
    }

    /** relay transition table. */
    object RelayTable {
        val initial = RelayState.Idle

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
            Transition("Idle", "BackendRegistered", "backend_register", "internal", null, null, emptyList()),
            Transition("BackendRegistered", "Bridged", "client_connect", "internal", null, "bridge_streams", emptyList()),
            Transition("Bridged", "BackendRegistered", "client_disconnect", "internal", null, "unbridge", emptyList()),
            Transition("BackendRegistered", "Idle", "backend_disconnect", "internal", null, null, emptyList()),
        )
    }

}

/** BackendMachine is the generated state machine for the backend actor. */
class BackendMachine {
    var state: BackendState = BackendState.Idle
        private set
    var currentToken: String = "none" // pairing token currently in play
    var activeTokens: String = "" // set of valid (non-revoked) tokens
    var usedTokens: String = "" // set of revoked tokens
    var backendEcdhPub: String = "none" // backend ECDH public key
    var receivedClientPub: String = "none" // pubkey backend received in pair_hello
    var backendSharedKey: String = "" // ECDH key derived by backend
    var backendCode: String = "" // code computed by backend
    var receivedCode: String = "" // code entered via CLI
    var codeAttempts: Int = 0 // failed code submission attempts
    var deviceSecret: String = "none" // persistent device secret
    var pairedDevices: String = "" // device IDs that completed pairing
    var receivedDeviceId: String = "none" // device_id from auth_request
    var authNoncesUsed: String = "" // set of consumed auth nonces
    var receivedAuthNonce: String = "none" // nonce from auth_request
    var secretPublished: Boolean = false // whether token has been published via backchannel
    var pingFailures: Int = 0 // consecutive failed pings
    var backoffLevel: Int = 0 // exponential backoff level
    var bActivePath: String = "relay" // backend active path
    var bDispatcherPath: String = "relay" // backend datagram dispatcher binding
    var monitorTarget: String = "none" // health monitor target
    var lanSignal: String = "pending" // LANReady notification state
    val guards = mutableMapOf<SessionProtocol.GuardID, () -> Boolean>()
    val actions = mutableMapOf<SessionProtocol.ActionID, () -> Unit>()

    /** Handle an event and return the list of commands to execute. */
    fun handleEvent(ev: SessionProtocol.EventID): List<SessionProtocol.CmdID> {
        val cmds = when {
            state == BackendState.Idle && ev == SessionProtocol.EventID.CliInitPair ->
                run {
                    actions[SessionProtocol.ActionID.GenerateToken]?.invoke()
                    currentToken = "tok_1"
                    // active_tokens: active_tokens \union {"tok_1"} (set by action)
                    state = BackendState.GenerateToken
                    emptyList()
                }
            state == BackendState.GenerateToken && ev == SessionProtocol.EventID.TokenCreated ->
                run {
                    actions[SessionProtocol.ActionID.RegisterRelay]?.invoke()
                    state = BackendState.RegisterRelay
                    emptyList()
                }
            state == BackendState.RegisterRelay && ev == SessionProtocol.EventID.RelayRegistered ->
                run {
                    secretPublished = true
                    state = BackendState.WaitingForClient
                    emptyList()
                }
            state == BackendState.WaitingForClient && ev == SessionProtocol.EventID.RecvPairHello && guards[SessionProtocol.GuardID.TokenValid]?.invoke() == true ->
                run {
                    actions[SessionProtocol.ActionID.DeriveSecret]?.invoke()
                    // received_client_pub: recv_msg.pubkey (set by action)
                    backendEcdhPub = "backend_pub"
                    // backend_shared_key: DeriveKey("backend_pub", recv_msg.pubkey) (set by action)
                    // backend_code: DeriveCode("backend_pub", recv_msg.pubkey) (set by action)
                    state = BackendState.DeriveSecret
                    emptyList()
                }
            state == BackendState.WaitingForClient && ev == SessionProtocol.EventID.RecvPairHello && guards[SessionProtocol.GuardID.TokenInvalid]?.invoke() == true ->
                run {
                    state = BackendState.Idle
                    emptyList()
                }
            state == BackendState.DeriveSecret && ev == SessionProtocol.EventID.EcdhComplete ->
                run {
                    state = BackendState.SendAck
                    emptyList()
                }
            state == BackendState.SendAck && ev == SessionProtocol.EventID.SignalCodeDisplay ->
                run {
                    state = BackendState.WaitingForCode
                    emptyList()
                }
            state == BackendState.WaitingForCode && ev == SessionProtocol.EventID.CliCodeEntered ->
                run {
                    // received_code: cli_entered_code (set by action)
                    state = BackendState.ValidateCode
                    emptyList()
                }
            state == BackendState.ValidateCode && ev == SessionProtocol.EventID.CheckCode && guards[SessionProtocol.GuardID.CodeCorrect]?.invoke() == true ->
                run {
                    state = BackendState.StorePaired
                    emptyList()
                }
            state == BackendState.ValidateCode && ev == SessionProtocol.EventID.CheckCode && guards[SessionProtocol.GuardID.CodeWrong]?.invoke() == true ->
                run {
                    // code_attempts: code_attempts + 1 (set by action)
                    state = BackendState.Idle
                    emptyList()
                }
            state == BackendState.StorePaired && ev == SessionProtocol.EventID.Finalise ->
                run {
                    actions[SessionProtocol.ActionID.StoreDevice]?.invoke()
                    deviceSecret = "dev_secret_1"
                    // paired_devices: paired_devices \union {"device_1"} (set by action)
                    // active_tokens: active_tokens \ {current_token} (set by action)
                    // used_tokens: used_tokens \union {current_token} (set by action)
                    state = BackendState.Paired
                    emptyList()
                }
            state == BackendState.Paired && ev == SessionProtocol.EventID.RecvAuthRequest ->
                run {
                    // received_device_id: recv_msg.device_id (set by action)
                    // received_auth_nonce: recv_msg.nonce (set by action)
                    state = BackendState.AuthCheck
                    emptyList()
                }
            state == BackendState.AuthCheck && ev == SessionProtocol.EventID.Verify && guards[SessionProtocol.GuardID.DeviceKnown]?.invoke() == true ->
                run {
                    actions[SessionProtocol.ActionID.VerifyDevice]?.invoke()
                    // auth_nonces_used: auth_nonces_used \union {received_auth_nonce} (set by action)
                    state = BackendState.SessionActive
                    emptyList()
                }
            state == BackendState.AuthCheck && ev == SessionProtocol.EventID.Verify && guards[SessionProtocol.GuardID.DeviceUnknown]?.invoke() == true ->
                run {
                    state = BackendState.Idle
                    emptyList()
                }
            state == BackendState.SessionActive && ev == SessionProtocol.EventID.SessionEstablished ->
                run {
                    state = BackendState.RelayConnected
                    emptyList()
                }
            state == BackendState.RelayConnected && ev == SessionProtocol.EventID.LanServerReady ->
                run {
                    state = BackendState.LANOffered
                    listOf(SessionProtocol.CmdID.SendLanOffer)
                }
            state == BackendState.LANOffered && ev == SessionProtocol.EventID.RecvLanVerify && guards[SessionProtocol.GuardID.ChallengeValid]?.invoke() == true ->
                run {
                    actions[SessionProtocol.ActionID.ActivateLan]?.invoke()
                    pingFailures = 0
                    backoffLevel = 0
                    bActivePath = "lan"
                    bDispatcherPath = "lan"
                    monitorTarget = "lan"
                    lanSignal = "ready"
                    state = BackendState.LANActive
                    listOf(SessionProtocol.CmdID.SendLanConfirm, SessionProtocol.CmdID.StartLanStreamReader, SessionProtocol.CmdID.StartLanDgReader, SessionProtocol.CmdID.StartMonitor, SessionProtocol.CmdID.SignalLanReady, SessionProtocol.CmdID.SetCryptoDatagram)
                }
            state == BackendState.LANOffered && ev == SessionProtocol.EventID.RecvLanVerify && guards[SessionProtocol.GuardID.ChallengeInvalid]?.invoke() == true ->
                run {
                    state = BackendState.RelayConnected
                    emptyList()
                }
            state == BackendState.LANOffered && ev == SessionProtocol.EventID.OfferTimeout ->
                run {
                    // backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
                    lanSignal = "pending"
                    state = BackendState.RelayBackoff
                    listOf(SessionProtocol.CmdID.ResetLanReady, SessionProtocol.CmdID.StartBackoffTimer)
                }
            state == BackendState.LANActive && ev == SessionProtocol.EventID.PingTick ->
                run {
                    state = BackendState.LANActive
                    listOf(SessionProtocol.CmdID.SendPathPing, SessionProtocol.CmdID.StartPongTimeout)
                }
            state == BackendState.LANActive && ev == SessionProtocol.EventID.PingTimeout ->
                run {
                    pingFailures = 1
                    state = BackendState.LANDegraded
                    emptyList()
                }
            state == BackendState.LANDegraded && ev == SessionProtocol.EventID.PingTick ->
                run {
                    state = BackendState.LANDegraded
                    listOf(SessionProtocol.CmdID.SendPathPing, SessionProtocol.CmdID.StartPongTimeout)
                }
            state == BackendState.LANActive && ev == SessionProtocol.EventID.LanStreamError ->
                run {
                    actions[SessionProtocol.ActionID.FallbackToRelay]?.invoke()
                    // backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
                    bActivePath = "relay"
                    bDispatcherPath = "relay"
                    monitorTarget = "none"
                    lanSignal = "pending"
                    pingFailures = 0
                    state = BackendState.RelayBackoff
                    listOf(SessionProtocol.CmdID.StopMonitor, SessionProtocol.CmdID.StopLanStreamReader, SessionProtocol.CmdID.StopLanDgReader, SessionProtocol.CmdID.CloseLanPath, SessionProtocol.CmdID.ResetLanReady, SessionProtocol.CmdID.StartBackoffTimer)
                }
            state == BackendState.LANDegraded && ev == SessionProtocol.EventID.LanStreamError ->
                run {
                    actions[SessionProtocol.ActionID.FallbackToRelay]?.invoke()
                    // backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
                    bActivePath = "relay"
                    bDispatcherPath = "relay"
                    monitorTarget = "none"
                    lanSignal = "pending"
                    pingFailures = 0
                    state = BackendState.RelayBackoff
                    listOf(SessionProtocol.CmdID.StopMonitor, SessionProtocol.CmdID.StopLanStreamReader, SessionProtocol.CmdID.StopLanDgReader, SessionProtocol.CmdID.CloseLanPath, SessionProtocol.CmdID.ResetLanReady, SessionProtocol.CmdID.StartBackoffTimer)
                }
            state == BackendState.LANDegraded && ev == SessionProtocol.EventID.RecvPathPong ->
                run {
                    actions[SessionProtocol.ActionID.ResetFailures]?.invoke()
                    pingFailures = 0
                    state = BackendState.LANActive
                    listOf(SessionProtocol.CmdID.CancelPongTimeout)
                }
            state == BackendState.LANDegraded && ev == SessionProtocol.EventID.PingTimeout && guards[SessionProtocol.GuardID.UnderMaxFailures]?.invoke() == true ->
                run {
                    // ping_failures: ping_failures + 1 (set by action)
                    state = BackendState.LANDegraded
                    emptyList()
                }
            state == BackendState.LANDegraded && ev == SessionProtocol.EventID.PingTimeout && guards[SessionProtocol.GuardID.AtMaxFailures]?.invoke() == true ->
                run {
                    actions[SessionProtocol.ActionID.FallbackToRelay]?.invoke()
                    // backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
                    bActivePath = "relay"
                    bDispatcherPath = "relay"
                    monitorTarget = "none"
                    lanSignal = "pending"
                    pingFailures = 0
                    state = BackendState.RelayBackoff
                    listOf(SessionProtocol.CmdID.StopMonitor, SessionProtocol.CmdID.StopLanStreamReader, SessionProtocol.CmdID.StopLanDgReader, SessionProtocol.CmdID.CloseLanPath, SessionProtocol.CmdID.ResetLanReady, SessionProtocol.CmdID.StartBackoffTimer)
                }
            state == BackendState.RelayBackoff && ev == SessionProtocol.EventID.BackoffExpired ->
                run {
                    state = BackendState.LANOffered
                    listOf(SessionProtocol.CmdID.SendLanOffer)
                }
            state == BackendState.RelayBackoff && ev == SessionProtocol.EventID.LanServerChanged ->
                run {
                    backoffLevel = 0
                    state = BackendState.LANOffered
                    listOf(SessionProtocol.CmdID.SendLanOffer)
                }
            state == BackendState.RelayConnected && ev == SessionProtocol.EventID.ReadvertiseTick && guards[SessionProtocol.GuardID.LanServerAvailable]?.invoke() == true ->
                run {
                    state = BackendState.LANOffered
                    listOf(SessionProtocol.CmdID.SendLanOffer)
                }
            state == BackendState.LANOffered && ev == SessionProtocol.EventID.AppForceFallback ->
                run {
                    lanSignal = "pending"
                    state = BackendState.RelayConnected
                    listOf(SessionProtocol.CmdID.ResetLanReady)
                }
            state == BackendState.LANActive && ev == SessionProtocol.EventID.AppForceFallback ->
                run {
                    actions[SessionProtocol.ActionID.FallbackToRelay]?.invoke()
                    // backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
                    bActivePath = "relay"
                    bDispatcherPath = "relay"
                    monitorTarget = "none"
                    lanSignal = "pending"
                    pingFailures = 0
                    state = BackendState.RelayBackoff
                    listOf(SessionProtocol.CmdID.StopMonitor, SessionProtocol.CmdID.CancelPongTimeout, SessionProtocol.CmdID.StopLanStreamReader, SessionProtocol.CmdID.StopLanDgReader, SessionProtocol.CmdID.CloseLanPath, SessionProtocol.CmdID.ResetLanReady, SessionProtocol.CmdID.StartBackoffTimer)
                }
            state == BackendState.LANDegraded && ev == SessionProtocol.EventID.AppForceFallback ->
                run {
                    actions[SessionProtocol.ActionID.FallbackToRelay]?.invoke()
                    // backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
                    bActivePath = "relay"
                    bDispatcherPath = "relay"
                    monitorTarget = "none"
                    lanSignal = "pending"
                    pingFailures = 0
                    state = BackendState.RelayBackoff
                    listOf(SessionProtocol.CmdID.StopMonitor, SessionProtocol.CmdID.CancelPongTimeout, SessionProtocol.CmdID.StopLanStreamReader, SessionProtocol.CmdID.StopLanDgReader, SessionProtocol.CmdID.CloseLanPath, SessionProtocol.CmdID.ResetLanReady, SessionProtocol.CmdID.StartBackoffTimer)
                }
            state == BackendState.RelayConnected && ev == SessionProtocol.EventID.Disconnect ->
                run {
                    state = BackendState.Paired
                    emptyList()
                }
            state == BackendState.RelayConnected && ev == SessionProtocol.EventID.AppSend ->
                run {
                    state = BackendState.RelayConnected
                    listOf(SessionProtocol.CmdID.WriteActiveStream)
                }
            state == BackendState.LANOffered && ev == SessionProtocol.EventID.AppSend ->
                run {
                    state = BackendState.LANOffered
                    listOf(SessionProtocol.CmdID.WriteActiveStream)
                }
            state == BackendState.LANActive && ev == SessionProtocol.EventID.AppSend ->
                run {
                    state = BackendState.LANActive
                    listOf(SessionProtocol.CmdID.WriteActiveStream)
                }
            state == BackendState.LANDegraded && ev == SessionProtocol.EventID.AppSend ->
                run {
                    state = BackendState.LANDegraded
                    listOf(SessionProtocol.CmdID.WriteActiveStream)
                }
            state == BackendState.RelayBackoff && ev == SessionProtocol.EventID.AppSend ->
                run {
                    state = BackendState.RelayBackoff
                    listOf(SessionProtocol.CmdID.WriteActiveStream)
                }
            state == BackendState.RelayConnected && ev == SessionProtocol.EventID.RelayStreamData ->
                run {
                    state = BackendState.RelayConnected
                    listOf(SessionProtocol.CmdID.DeliverRecv)
                }
            state == BackendState.LANOffered && ev == SessionProtocol.EventID.RelayStreamData ->
                run {
                    state = BackendState.LANOffered
                    listOf(SessionProtocol.CmdID.DeliverRecv)
                }
            state == BackendState.LANActive && ev == SessionProtocol.EventID.RelayStreamData ->
                run {
                    state = BackendState.LANActive
                    listOf(SessionProtocol.CmdID.DeliverRecv)
                }
            state == BackendState.LANDegraded && ev == SessionProtocol.EventID.RelayStreamData ->
                run {
                    state = BackendState.LANDegraded
                    listOf(SessionProtocol.CmdID.DeliverRecv)
                }
            state == BackendState.RelayBackoff && ev == SessionProtocol.EventID.RelayStreamData ->
                run {
                    state = BackendState.RelayBackoff
                    listOf(SessionProtocol.CmdID.DeliverRecv)
                }
            state == BackendState.RelayConnected && ev == SessionProtocol.EventID.RelayStreamError ->
                run {
                    state = BackendState.RelayConnected
                    listOf(SessionProtocol.CmdID.DeliverRecvError)
                }
            state == BackendState.LANOffered && ev == SessionProtocol.EventID.RelayStreamError ->
                run {
                    state = BackendState.LANOffered
                    listOf(SessionProtocol.CmdID.DeliverRecvError)
                }
            state == BackendState.LANActive && ev == SessionProtocol.EventID.RelayStreamError ->
                run {
                    state = BackendState.LANActive
                    listOf(SessionProtocol.CmdID.DeliverRecvError)
                }
            state == BackendState.LANDegraded && ev == SessionProtocol.EventID.RelayStreamError ->
                run {
                    state = BackendState.LANDegraded
                    listOf(SessionProtocol.CmdID.DeliverRecvError)
                }
            state == BackendState.RelayBackoff && ev == SessionProtocol.EventID.RelayStreamError ->
                run {
                    state = BackendState.RelayBackoff
                    listOf(SessionProtocol.CmdID.DeliverRecvError)
                }
            state == BackendState.RelayConnected && ev == SessionProtocol.EventID.AppSendDatagram ->
                run {
                    state = BackendState.RelayConnected
                    listOf(SessionProtocol.CmdID.SendActiveDatagram)
                }
            state == BackendState.LANOffered && ev == SessionProtocol.EventID.AppSendDatagram ->
                run {
                    state = BackendState.LANOffered
                    listOf(SessionProtocol.CmdID.SendActiveDatagram)
                }
            state == BackendState.LANActive && ev == SessionProtocol.EventID.AppSendDatagram ->
                run {
                    state = BackendState.LANActive
                    listOf(SessionProtocol.CmdID.SendActiveDatagram)
                }
            state == BackendState.LANDegraded && ev == SessionProtocol.EventID.AppSendDatagram ->
                run {
                    state = BackendState.LANDegraded
                    listOf(SessionProtocol.CmdID.SendActiveDatagram)
                }
            state == BackendState.RelayBackoff && ev == SessionProtocol.EventID.AppSendDatagram ->
                run {
                    state = BackendState.RelayBackoff
                    listOf(SessionProtocol.CmdID.SendActiveDatagram)
                }
            state == BackendState.RelayConnected && ev == SessionProtocol.EventID.RelayDatagram ->
                run {
                    state = BackendState.RelayConnected
                    listOf(SessionProtocol.CmdID.DeliverRecvDatagram)
                }
            state == BackendState.LANOffered && ev == SessionProtocol.EventID.RelayDatagram ->
                run {
                    state = BackendState.LANOffered
                    listOf(SessionProtocol.CmdID.DeliverRecvDatagram)
                }
            state == BackendState.LANActive && ev == SessionProtocol.EventID.RelayDatagram ->
                run {
                    state = BackendState.LANActive
                    listOf(SessionProtocol.CmdID.DeliverRecvDatagram)
                }
            state == BackendState.LANDegraded && ev == SessionProtocol.EventID.RelayDatagram ->
                run {
                    state = BackendState.LANDegraded
                    listOf(SessionProtocol.CmdID.DeliverRecvDatagram)
                }
            state == BackendState.RelayBackoff && ev == SessionProtocol.EventID.RelayDatagram ->
                run {
                    state = BackendState.RelayBackoff
                    listOf(SessionProtocol.CmdID.DeliverRecvDatagram)
                }
            state == BackendState.LANActive && ev == SessionProtocol.EventID.LanStreamData ->
                run {
                    state = BackendState.LANActive
                    listOf(SessionProtocol.CmdID.DeliverRecv)
                }
            state == BackendState.LANDegraded && ev == SessionProtocol.EventID.LanStreamData ->
                run {
                    state = BackendState.LANDegraded
                    listOf(SessionProtocol.CmdID.DeliverRecv)
                }
            state == BackendState.LANActive && ev == SessionProtocol.EventID.LanDatagram ->
                run {
                    state = BackendState.LANActive
                    listOf(SessionProtocol.CmdID.DeliverRecvDatagram)
                }
            state == BackendState.LANDegraded && ev == SessionProtocol.EventID.LanDatagram ->
                run {
                    state = BackendState.LANDegraded
                    listOf(SessionProtocol.CmdID.DeliverRecvDatagram)
                }
            else -> emptyList()
        }
        return cmds
    }
}

/** ClientMachine is the generated state machine for the client actor. */
class ClientMachine {
    var state: ClientState = ClientState.Idle
        private set
    var receivedBackendPub: String = "none" // pubkey client received in pair_hello_ack
    var clientSharedKey: String = "" // ECDH key derived by client
    var clientCode: String = "" // code computed by client
    var cActivePath: String = "relay" // client active path
    var cDispatcherPath: String = "relay" // client datagram dispatcher binding
    var lanSignal: String = "pending" // LANReady notification state
    val guards = mutableMapOf<SessionProtocol.GuardID, () -> Boolean>()
    val actions = mutableMapOf<SessionProtocol.ActionID, () -> Unit>()

    /** Handle an event and return the list of commands to execute. */
    fun handleEvent(ev: SessionProtocol.EventID): List<SessionProtocol.CmdID> {
        val cmds = when {
            state == ClientState.Idle && ev == SessionProtocol.EventID.BackchannelReceived ->
                run {
                    state = ClientState.ObtainBackchannelSecret
                    emptyList()
                }
            state == ClientState.ObtainBackchannelSecret && ev == SessionProtocol.EventID.SecretParsed ->
                run {
                    state = ClientState.ConnectRelay
                    emptyList()
                }
            state == ClientState.ConnectRelay && ev == SessionProtocol.EventID.RelayConnected ->
                run {
                    state = ClientState.GenKeyPair
                    emptyList()
                }
            state == ClientState.GenKeyPair && ev == SessionProtocol.EventID.KeyPairGenerated ->
                run {
                    actions[SessionProtocol.ActionID.SendPairHello]?.invoke()
                    state = ClientState.WaitAck
                    emptyList()
                }
            state == ClientState.WaitAck && ev == SessionProtocol.EventID.RecvPairHelloAck ->
                run {
                    actions[SessionProtocol.ActionID.DeriveSecret]?.invoke()
                    // received_backend_pub: recv_msg.pubkey (set by action)
                    // client_shared_key: DeriveKey("client_pub", recv_msg.pubkey) (set by action)
                    state = ClientState.E2EReady
                    emptyList()
                }
            state == ClientState.E2EReady && ev == SessionProtocol.EventID.RecvPairConfirm ->
                run {
                    // client_code: DeriveCode(received_backend_pub, "client_pub") (set by action)
                    state = ClientState.ShowCode
                    emptyList()
                }
            state == ClientState.ShowCode && ev == SessionProtocol.EventID.CodeDisplayed ->
                run {
                    state = ClientState.WaitPairComplete
                    emptyList()
                }
            state == ClientState.WaitPairComplete && ev == SessionProtocol.EventID.RecvPairComplete ->
                run {
                    actions[SessionProtocol.ActionID.StoreSecret]?.invoke()
                    state = ClientState.Paired
                    emptyList()
                }
            state == ClientState.Paired && ev == SessionProtocol.EventID.AppLaunch ->
                run {
                    state = ClientState.Reconnect
                    emptyList()
                }
            state == ClientState.Reconnect && ev == SessionProtocol.EventID.RelayConnected ->
                run {
                    state = ClientState.SendAuth
                    emptyList()
                }
            state == ClientState.SendAuth && ev == SessionProtocol.EventID.RecvAuthOk ->
                run {
                    state = ClientState.SessionActive
                    emptyList()
                }
            state == ClientState.SessionActive && ev == SessionProtocol.EventID.SessionEstablished ->
                run {
                    state = ClientState.RelayConnected
                    emptyList()
                }
            state == ClientState.RelayConnected && ev == SessionProtocol.EventID.RecvLanOffer && guards[SessionProtocol.GuardID.LanEnabled]?.invoke() == true ->
                run {
                    actions[SessionProtocol.ActionID.DialLan]?.invoke()
                    state = ClientState.LANConnecting
                    listOf(SessionProtocol.CmdID.DialLan)
                }
            state == ClientState.RelayConnected && ev == SessionProtocol.EventID.RecvLanOffer && guards[SessionProtocol.GuardID.LanDisabled]?.invoke() == true ->
                run {
                    state = ClientState.RelayConnected
                    emptyList()
                }
            state == ClientState.LANConnecting && ev == SessionProtocol.EventID.LanDialOk ->
                run {
                    state = ClientState.LANVerifying
                    listOf(SessionProtocol.CmdID.SendLanVerify)
                }
            state == ClientState.LANConnecting && ev == SessionProtocol.EventID.LanDialFailed ->
                run {
                    state = ClientState.RelayConnected
                    emptyList()
                }
            state == ClientState.LANVerifying && ev == SessionProtocol.EventID.RecvLanConfirm ->
                run {
                    actions[SessionProtocol.ActionID.ActivateLan]?.invoke()
                    cActivePath = "lan"
                    cDispatcherPath = "lan"
                    lanSignal = "ready"
                    state = ClientState.LANActive
                    listOf(SessionProtocol.CmdID.StartLanStreamReader, SessionProtocol.CmdID.StartLanDgReader, SessionProtocol.CmdID.SignalLanReady, SessionProtocol.CmdID.SetCryptoDatagram)
                }
            state == ClientState.LANVerifying && ev == SessionProtocol.EventID.VerifyTimeout ->
                run {
                    cDispatcherPath = "relay"
                    state = ClientState.RelayConnected
                    emptyList()
                }
            state == ClientState.LANActive && ev == SessionProtocol.EventID.RecvPathPing ->
                run {
                    state = ClientState.LANActive
                    listOf(SessionProtocol.CmdID.SendPathPong)
                }
            state == ClientState.LANActive && ev == SessionProtocol.EventID.LanError ->
                run {
                    actions[SessionProtocol.ActionID.FallbackToRelay]?.invoke()
                    cActivePath = "relay"
                    cDispatcherPath = "relay"
                    lanSignal = "pending"
                    state = ClientState.RelayFallback
                    listOf(SessionProtocol.CmdID.StopLanStreamReader, SessionProtocol.CmdID.StopLanDgReader, SessionProtocol.CmdID.CloseLanPath, SessionProtocol.CmdID.ResetLanReady)
                }
            state == ClientState.LANActive && ev == SessionProtocol.EventID.LanStreamError ->
                run {
                    actions[SessionProtocol.ActionID.FallbackToRelay]?.invoke()
                    cActivePath = "relay"
                    cDispatcherPath = "relay"
                    lanSignal = "pending"
                    state = ClientState.RelayFallback
                    listOf(SessionProtocol.CmdID.StopLanStreamReader, SessionProtocol.CmdID.StopLanDgReader, SessionProtocol.CmdID.CloseLanPath, SessionProtocol.CmdID.ResetLanReady)
                }
            state == ClientState.RelayFallback && ev == SessionProtocol.EventID.RelayOk ->
                run {
                    state = ClientState.RelayConnected
                    emptyList()
                }
            state == ClientState.LANActive && ev == SessionProtocol.EventID.RecvLanOffer && guards[SessionProtocol.GuardID.LanEnabled]?.invoke() == true ->
                run {
                    actions[SessionProtocol.ActionID.DialLan]?.invoke()
                    state = ClientState.LANConnecting
                    listOf(SessionProtocol.CmdID.StopLanStreamReader, SessionProtocol.CmdID.StopLanDgReader, SessionProtocol.CmdID.CloseLanPath, SessionProtocol.CmdID.DialLan)
                }
            state == ClientState.LANConnecting && ev == SessionProtocol.EventID.AppForceFallback ->
                run {
                    state = ClientState.RelayConnected
                    emptyList()
                }
            state == ClientState.LANVerifying && ev == SessionProtocol.EventID.AppForceFallback ->
                run {
                    cDispatcherPath = "relay"
                    state = ClientState.RelayConnected
                    listOf(SessionProtocol.CmdID.StopLanStreamReader, SessionProtocol.CmdID.StopLanDgReader, SessionProtocol.CmdID.CloseLanPath)
                }
            state == ClientState.LANActive && ev == SessionProtocol.EventID.AppForceFallback ->
                run {
                    actions[SessionProtocol.ActionID.FallbackToRelay]?.invoke()
                    cActivePath = "relay"
                    cDispatcherPath = "relay"
                    lanSignal = "pending"
                    state = ClientState.RelayConnected
                    listOf(SessionProtocol.CmdID.StopLanStreamReader, SessionProtocol.CmdID.StopLanDgReader, SessionProtocol.CmdID.CloseLanPath, SessionProtocol.CmdID.ResetLanReady)
                }
            state == ClientState.RelayConnected && ev == SessionProtocol.EventID.Disconnect ->
                run {
                    state = ClientState.Paired
                    emptyList()
                }
            state == ClientState.RelayConnected && ev == SessionProtocol.EventID.AppSend ->
                run {
                    state = ClientState.RelayConnected
                    listOf(SessionProtocol.CmdID.WriteActiveStream)
                }
            state == ClientState.LANConnecting && ev == SessionProtocol.EventID.AppSend ->
                run {
                    state = ClientState.LANConnecting
                    listOf(SessionProtocol.CmdID.WriteActiveStream)
                }
            state == ClientState.LANVerifying && ev == SessionProtocol.EventID.AppSend ->
                run {
                    state = ClientState.LANVerifying
                    listOf(SessionProtocol.CmdID.WriteActiveStream)
                }
            state == ClientState.LANActive && ev == SessionProtocol.EventID.AppSend ->
                run {
                    state = ClientState.LANActive
                    listOf(SessionProtocol.CmdID.WriteActiveStream)
                }
            state == ClientState.RelayFallback && ev == SessionProtocol.EventID.AppSend ->
                run {
                    state = ClientState.RelayFallback
                    listOf(SessionProtocol.CmdID.WriteActiveStream)
                }
            state == ClientState.RelayConnected && ev == SessionProtocol.EventID.RelayStreamData ->
                run {
                    state = ClientState.RelayConnected
                    listOf(SessionProtocol.CmdID.DeliverRecv)
                }
            state == ClientState.LANConnecting && ev == SessionProtocol.EventID.RelayStreamData ->
                run {
                    state = ClientState.LANConnecting
                    listOf(SessionProtocol.CmdID.DeliverRecv)
                }
            state == ClientState.LANVerifying && ev == SessionProtocol.EventID.RelayStreamData ->
                run {
                    state = ClientState.LANVerifying
                    listOf(SessionProtocol.CmdID.DeliverRecv)
                }
            state == ClientState.LANActive && ev == SessionProtocol.EventID.RelayStreamData ->
                run {
                    state = ClientState.LANActive
                    listOf(SessionProtocol.CmdID.DeliverRecv)
                }
            state == ClientState.RelayFallback && ev == SessionProtocol.EventID.RelayStreamData ->
                run {
                    state = ClientState.RelayFallback
                    listOf(SessionProtocol.CmdID.DeliverRecv)
                }
            state == ClientState.RelayConnected && ev == SessionProtocol.EventID.RelayStreamError ->
                run {
                    state = ClientState.RelayConnected
                    listOf(SessionProtocol.CmdID.DeliverRecvError)
                }
            state == ClientState.LANConnecting && ev == SessionProtocol.EventID.RelayStreamError ->
                run {
                    state = ClientState.LANConnecting
                    listOf(SessionProtocol.CmdID.DeliverRecvError)
                }
            state == ClientState.LANVerifying && ev == SessionProtocol.EventID.RelayStreamError ->
                run {
                    state = ClientState.LANVerifying
                    listOf(SessionProtocol.CmdID.DeliverRecvError)
                }
            state == ClientState.LANActive && ev == SessionProtocol.EventID.RelayStreamError ->
                run {
                    state = ClientState.LANActive
                    listOf(SessionProtocol.CmdID.DeliverRecvError)
                }
            state == ClientState.RelayFallback && ev == SessionProtocol.EventID.RelayStreamError ->
                run {
                    state = ClientState.RelayFallback
                    listOf(SessionProtocol.CmdID.DeliverRecvError)
                }
            state == ClientState.RelayConnected && ev == SessionProtocol.EventID.AppSendDatagram ->
                run {
                    state = ClientState.RelayConnected
                    listOf(SessionProtocol.CmdID.SendActiveDatagram)
                }
            state == ClientState.LANConnecting && ev == SessionProtocol.EventID.AppSendDatagram ->
                run {
                    state = ClientState.LANConnecting
                    listOf(SessionProtocol.CmdID.SendActiveDatagram)
                }
            state == ClientState.LANVerifying && ev == SessionProtocol.EventID.AppSendDatagram ->
                run {
                    state = ClientState.LANVerifying
                    listOf(SessionProtocol.CmdID.SendActiveDatagram)
                }
            state == ClientState.LANActive && ev == SessionProtocol.EventID.AppSendDatagram ->
                run {
                    state = ClientState.LANActive
                    listOf(SessionProtocol.CmdID.SendActiveDatagram)
                }
            state == ClientState.RelayFallback && ev == SessionProtocol.EventID.AppSendDatagram ->
                run {
                    state = ClientState.RelayFallback
                    listOf(SessionProtocol.CmdID.SendActiveDatagram)
                }
            state == ClientState.RelayConnected && ev == SessionProtocol.EventID.RelayDatagram ->
                run {
                    state = ClientState.RelayConnected
                    listOf(SessionProtocol.CmdID.DeliverRecvDatagram)
                }
            state == ClientState.LANConnecting && ev == SessionProtocol.EventID.RelayDatagram ->
                run {
                    state = ClientState.LANConnecting
                    listOf(SessionProtocol.CmdID.DeliverRecvDatagram)
                }
            state == ClientState.LANVerifying && ev == SessionProtocol.EventID.RelayDatagram ->
                run {
                    state = ClientState.LANVerifying
                    listOf(SessionProtocol.CmdID.DeliverRecvDatagram)
                }
            state == ClientState.LANActive && ev == SessionProtocol.EventID.RelayDatagram ->
                run {
                    state = ClientState.LANActive
                    listOf(SessionProtocol.CmdID.DeliverRecvDatagram)
                }
            state == ClientState.RelayFallback && ev == SessionProtocol.EventID.RelayDatagram ->
                run {
                    state = ClientState.RelayFallback
                    listOf(SessionProtocol.CmdID.DeliverRecvDatagram)
                }
            state == ClientState.LANActive && ev == SessionProtocol.EventID.LanStreamData ->
                run {
                    state = ClientState.LANActive
                    listOf(SessionProtocol.CmdID.DeliverRecv)
                }
            state == ClientState.LANActive && ev == SessionProtocol.EventID.LanDatagram ->
                run {
                    state = ClientState.LANActive
                    listOf(SessionProtocol.CmdID.DeliverRecvDatagram)
                }
            else -> emptyList()
        }
        return cmds
    }
}

/** RelayMachine is the generated state machine for the relay actor. */
class RelayMachine {
    var state: RelayState = RelayState.Idle
        private set
    var relayBridge: String = "idle" // relay bridge state
    val guards = mutableMapOf<SessionProtocol.GuardID, () -> Boolean>()
    val actions = mutableMapOf<SessionProtocol.ActionID, () -> Unit>()

    /** Handle an event and return the list of commands to execute. */
    fun handleEvent(ev: SessionProtocol.EventID): List<SessionProtocol.CmdID> {
        val cmds = when {
            state == RelayState.Idle && ev == SessionProtocol.EventID.BackendRegister ->
                run {
                    state = RelayState.BackendRegistered
                    emptyList()
                }
            state == RelayState.BackendRegistered && ev == SessionProtocol.EventID.ClientConnect ->
                run {
                    actions[SessionProtocol.ActionID.BridgeStreams]?.invoke()
                    relayBridge = "active"
                    state = RelayState.Bridged
                    emptyList()
                }
            state == RelayState.Bridged && ev == SessionProtocol.EventID.ClientDisconnect ->
                run {
                    actions[SessionProtocol.ActionID.Unbridge]?.invoke()
                    relayBridge = "idle"
                    state = RelayState.BackendRegistered
                    emptyList()
                }
            state == RelayState.BackendRegistered && ev == SessionProtocol.EventID.BackendDisconnect ->
                run {
                    state = RelayState.Idle
                    emptyList()
                }
            else -> emptyList()
        }
        return cmds
    }
}


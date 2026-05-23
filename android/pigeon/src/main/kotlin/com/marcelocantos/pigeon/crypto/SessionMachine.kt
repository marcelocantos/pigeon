// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Auto-generated from protocol definition. Do not edit.
// Source of truth: protocol/*.yaml

package com.marcelocantos.pigeon.crypto

enum class SessionBackendState(val value: String) {
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
    CandidatesAdvertised("CandidatesAdvertised"),
    AltActive("AltActive"),
    RelayBackoff("RelayBackoff"),
    AltDegraded("AltDegraded");
}

enum class SessionClientState(val value: String) {
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
    PairDialing("PairDialing"),
    PairChecking("PairChecking"),
    AltActive("AltActive"),
    RelayFallback("RelayFallback");
}

enum class SessionRelayState(val value: String) {
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
        Candidates("candidates"),
        PairCheck("pair_check"),
        PairCheckAck("pair_check_ack"),
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
        AltEnabled("alt_enabled"),
        AltDisabled("alt_disabled"),
        LocalCandidatesAvailable("local_candidates_available"),
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
        DialCandidate("dial_candidate"),
        BridgeStreams("bridge_streams"),
        Unbridge("unbridge");
    }

    enum class EventID(val value: String) {
        AltDatagram("alt_datagram"),
        AltError("alt_error"),
        AltStreamData("alt_stream_data"),
        AltStreamError("alt_stream_error"),
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
        CandidatesChanged("candidates_changed"),
        CandidatesGathered("candidates_gathered"),
        CandidatesRefreshTick("candidates_refresh_tick"),
        CandidatesTimeout("candidates_timeout"),
        CheckCode("check_code"),
        CliCodeEntered("cli_code_entered"),
        CliInitPair("cli_init_pair"),
        ClientConnect("client_connect"),
        ClientDisconnect("client_disconnect"),
        CodeDisplayed("code_displayed"),
        DialFailed("dial_failed"),
        DialOk("dial_ok"),
        Disconnect("disconnect"),
        EcdhComplete("ecdh_complete"),
        Finalise("finalise"),
        KeyPairGenerated("key_pair_generated"),
        PairCheckOk("pair_check_ok"),
        PingTick("ping_tick"),
        PingTimeout("ping_timeout"),
        RecvAuthOk("recv_auth_ok"),
        RecvAuthRequest("recv_auth_request"),
        RecvCandidates("recv_candidates"),
        RecvPairCheck("recv_pair_check"),
        RecvPairCheckAck("recv_pair_check_ack"),
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
        SendCandidates("send_candidates"),
        SendPairCheck("send_pair_check"),
        SendPairCheckAck("send_pair_check_ack"),
        DialCandidate("dial_candidate"),
        DeliverRecv("deliver_recv"),
        DeliverRecvError("deliver_recv_error"),
        DeliverRecvDatagram("deliver_recv_datagram"),
        StartAltStreamReader("start_alt_stream_reader"),
        StopAltStreamReader("stop_alt_stream_reader"),
        StartAltDgReader("start_alt_dg_reader"),
        StopAltDgReader("stop_alt_dg_reader"),
        StartMonitor("start_monitor"),
        StopMonitor("stop_monitor"),
        StartPongTimeout("start_pong_timeout"),
        CancelPongTimeout("cancel_pong_timeout"),
        StartBackoffTimer("start_backoff_timer"),
        CloseAltPath("close_alt_path"),
        SignalAltReady("signal_alt_ready"),
        ResetAltReady("reset_alt_ready"),
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
        const val FRAME_CANDIDATES: Byte = 0x01.toByte()
        const val FRAME_CUTOVER: Byte = 0x02.toByte()
        const val FRAME_PAIR_CHECK: Byte = 0x03.toByte()
        const val FRAME_PAIR_CHECK_ACK: Byte = 0x04.toByte()
        const val CAND_HOST = "host"
        const val CAND_SRFLX = "srflx"
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
        val initial = SessionBackendState.Idle

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
            Transition("RelayConnected", "CandidatesAdvertised", "candidates_gathered", "internal", null, null, listOf("client" to "candidates")),
            Transition("CandidatesAdvertised", "AltActive", "pair_check", "recv", "challenge_valid", "activate_lan", listOf("client" to "pair_check_ack")),
            Transition("CandidatesAdvertised", "RelayConnected", "pair_check", "recv", "challenge_invalid", null, emptyList()),
            Transition("CandidatesAdvertised", "RelayBackoff", "candidates_timeout", "internal", null, null, emptyList()),
            Transition("AltActive", "AltActive", "ping_tick", "internal", null, null, listOf("client" to "path_ping")),
            Transition("AltActive", "AltDegraded", "ping_timeout", "internal", null, null, emptyList()),
            Transition("AltDegraded", "AltDegraded", "ping_tick", "internal", null, null, listOf("client" to "path_ping")),
            Transition("AltActive", "RelayBackoff", "alt_stream_error", "internal", null, "fallback_to_relay", emptyList()),
            Transition("AltDegraded", "RelayBackoff", "alt_stream_error", "internal", null, "fallback_to_relay", emptyList()),
            Transition("AltDegraded", "AltActive", "path_pong", "recv", null, "reset_failures", emptyList()),
            Transition("AltDegraded", "AltDegraded", "ping_timeout", "internal", "under_max_failures", null, emptyList()),
            Transition("AltDegraded", "RelayBackoff", "ping_timeout", "internal", "at_max_failures", "fallback_to_relay", emptyList()),
            Transition("RelayBackoff", "CandidatesAdvertised", "backoff_expired", "internal", null, null, listOf("client" to "candidates")),
            Transition("RelayBackoff", "CandidatesAdvertised", "candidates_changed", "internal", null, null, listOf("client" to "candidates")),
            Transition("RelayConnected", "CandidatesAdvertised", "candidates_refresh_tick", "internal", "local_candidates_available", null, listOf("client" to "candidates")),
            Transition("CandidatesAdvertised", "RelayConnected", "app_force_fallback", "internal", null, null, emptyList()),
            Transition("AltActive", "RelayBackoff", "app_force_fallback", "internal", null, "fallback_to_relay", emptyList()),
            Transition("AltDegraded", "RelayBackoff", "app_force_fallback", "internal", null, "fallback_to_relay", emptyList()),
            Transition("RelayConnected", "Paired", "disconnect", "internal", null, null, emptyList()),
            Transition("RelayConnected", "RelayConnected", "app_send", "internal", null, null, emptyList()),
            Transition("CandidatesAdvertised", "CandidatesAdvertised", "app_send", "internal", null, null, emptyList()),
            Transition("AltActive", "AltActive", "app_send", "internal", null, null, emptyList()),
            Transition("AltDegraded", "AltDegraded", "app_send", "internal", null, null, emptyList()),
            Transition("RelayBackoff", "RelayBackoff", "app_send", "internal", null, null, emptyList()),
            Transition("RelayConnected", "RelayConnected", "relay_stream_data", "internal", null, null, emptyList()),
            Transition("CandidatesAdvertised", "CandidatesAdvertised", "relay_stream_data", "internal", null, null, emptyList()),
            Transition("AltActive", "AltActive", "relay_stream_data", "internal", null, null, emptyList()),
            Transition("AltDegraded", "AltDegraded", "relay_stream_data", "internal", null, null, emptyList()),
            Transition("RelayBackoff", "RelayBackoff", "relay_stream_data", "internal", null, null, emptyList()),
            Transition("RelayConnected", "RelayConnected", "relay_stream_error", "internal", null, null, emptyList()),
            Transition("CandidatesAdvertised", "CandidatesAdvertised", "relay_stream_error", "internal", null, null, emptyList()),
            Transition("AltActive", "AltActive", "relay_stream_error", "internal", null, null, emptyList()),
            Transition("AltDegraded", "AltDegraded", "relay_stream_error", "internal", null, null, emptyList()),
            Transition("RelayBackoff", "RelayBackoff", "relay_stream_error", "internal", null, null, emptyList()),
            Transition("RelayConnected", "RelayConnected", "app_send_datagram", "internal", null, null, emptyList()),
            Transition("CandidatesAdvertised", "CandidatesAdvertised", "app_send_datagram", "internal", null, null, emptyList()),
            Transition("AltActive", "AltActive", "app_send_datagram", "internal", null, null, emptyList()),
            Transition("AltDegraded", "AltDegraded", "app_send_datagram", "internal", null, null, emptyList()),
            Transition("RelayBackoff", "RelayBackoff", "app_send_datagram", "internal", null, null, emptyList()),
            Transition("RelayConnected", "RelayConnected", "relay_datagram", "internal", null, null, emptyList()),
            Transition("CandidatesAdvertised", "CandidatesAdvertised", "relay_datagram", "internal", null, null, emptyList()),
            Transition("AltActive", "AltActive", "relay_datagram", "internal", null, null, emptyList()),
            Transition("AltDegraded", "AltDegraded", "relay_datagram", "internal", null, null, emptyList()),
            Transition("RelayBackoff", "RelayBackoff", "relay_datagram", "internal", null, null, emptyList()),
            Transition("AltActive", "AltActive", "alt_stream_data", "internal", null, null, emptyList()),
            Transition("AltDegraded", "AltDegraded", "alt_stream_data", "internal", null, null, emptyList()),
            Transition("AltActive", "AltActive", "alt_datagram", "internal", null, null, emptyList()),
            Transition("AltDegraded", "AltDegraded", "alt_datagram", "internal", null, null, emptyList()),
        )
    }

    /** client transition table. */
    object ClientTable {
        val initial = SessionClientState.Idle

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
            Transition("RelayConnected", "PairDialing", "candidates", "recv", "alt_enabled", "dial_candidate", emptyList()),
            Transition("RelayConnected", "RelayConnected", "candidates", "recv", "alt_disabled", null, emptyList()),
            Transition("PairDialing", "PairChecking", "dial_ok", "internal", null, null, listOf("backend" to "pair_check")),
            Transition("PairDialing", "RelayConnected", "dial_failed", "internal", null, null, emptyList()),
            Transition("PairChecking", "AltActive", "pair_check_ack", "recv", null, "activate_lan", emptyList()),
            Transition("PairChecking", "RelayConnected", "verify_timeout", "internal", null, null, emptyList()),
            Transition("AltActive", "AltActive", "path_ping", "recv", null, null, listOf("backend" to "path_pong")),
            Transition("AltActive", "RelayFallback", "alt_error", "internal", null, "fallback_to_relay", emptyList()),
            Transition("AltActive", "RelayFallback", "alt_stream_error", "internal", null, "fallback_to_relay", emptyList()),
            Transition("RelayFallback", "RelayConnected", "relay_ok", "internal", null, null, emptyList()),
            Transition("AltActive", "PairDialing", "candidates", "recv", "alt_enabled", "dial_candidate", emptyList()),
            Transition("PairDialing", "RelayConnected", "app_force_fallback", "internal", null, null, emptyList()),
            Transition("PairChecking", "RelayConnected", "app_force_fallback", "internal", null, null, emptyList()),
            Transition("AltActive", "RelayConnected", "app_force_fallback", "internal", null, "fallback_to_relay", emptyList()),
            Transition("RelayConnected", "Paired", "disconnect", "internal", null, null, emptyList()),
            Transition("RelayConnected", "RelayConnected", "app_send", "internal", null, null, emptyList()),
            Transition("PairDialing", "PairDialing", "app_send", "internal", null, null, emptyList()),
            Transition("PairChecking", "PairChecking", "app_send", "internal", null, null, emptyList()),
            Transition("AltActive", "AltActive", "app_send", "internal", null, null, emptyList()),
            Transition("RelayFallback", "RelayFallback", "app_send", "internal", null, null, emptyList()),
            Transition("RelayConnected", "RelayConnected", "relay_stream_data", "internal", null, null, emptyList()),
            Transition("PairDialing", "PairDialing", "relay_stream_data", "internal", null, null, emptyList()),
            Transition("PairChecking", "PairChecking", "relay_stream_data", "internal", null, null, emptyList()),
            Transition("AltActive", "AltActive", "relay_stream_data", "internal", null, null, emptyList()),
            Transition("RelayFallback", "RelayFallback", "relay_stream_data", "internal", null, null, emptyList()),
            Transition("RelayConnected", "RelayConnected", "relay_stream_error", "internal", null, null, emptyList()),
            Transition("PairDialing", "PairDialing", "relay_stream_error", "internal", null, null, emptyList()),
            Transition("PairChecking", "PairChecking", "relay_stream_error", "internal", null, null, emptyList()),
            Transition("AltActive", "AltActive", "relay_stream_error", "internal", null, null, emptyList()),
            Transition("RelayFallback", "RelayFallback", "relay_stream_error", "internal", null, null, emptyList()),
            Transition("RelayConnected", "RelayConnected", "app_send_datagram", "internal", null, null, emptyList()),
            Transition("PairDialing", "PairDialing", "app_send_datagram", "internal", null, null, emptyList()),
            Transition("PairChecking", "PairChecking", "app_send_datagram", "internal", null, null, emptyList()),
            Transition("AltActive", "AltActive", "app_send_datagram", "internal", null, null, emptyList()),
            Transition("RelayFallback", "RelayFallback", "app_send_datagram", "internal", null, null, emptyList()),
            Transition("RelayConnected", "RelayConnected", "relay_datagram", "internal", null, null, emptyList()),
            Transition("PairDialing", "PairDialing", "relay_datagram", "internal", null, null, emptyList()),
            Transition("PairChecking", "PairChecking", "relay_datagram", "internal", null, null, emptyList()),
            Transition("AltActive", "AltActive", "relay_datagram", "internal", null, null, emptyList()),
            Transition("RelayFallback", "RelayFallback", "relay_datagram", "internal", null, null, emptyList()),
            Transition("AltActive", "AltActive", "alt_stream_data", "internal", null, null, emptyList()),
            Transition("AltActive", "AltActive", "alt_datagram", "internal", null, null, emptyList()),
        )
    }

    /** relay transition table. */
    object RelayTable {
        val initial = SessionRelayState.Idle

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

/** SessionBackendMachine is the generated state machine for the backend actor. */
class SessionBackendMachine {
    var state: SessionBackendState = SessionBackendState.Idle
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
    var altSignal: String = "pending" // AltReady notification state
    val guards = mutableMapOf<SessionProtocol.GuardID, () -> Boolean>()
    val actions = mutableMapOf<SessionProtocol.ActionID, () -> Unit>()

    /** Handle an event and return the list of commands to execute. */
    fun handleEvent(ev: SessionProtocol.EventID): List<SessionProtocol.CmdID> {
        val cmds: List<SessionProtocol.CmdID> = when {
            state == SessionBackendState.Idle && ev == SessionProtocol.EventID.CliInitPair ->
                run {
                    actions[SessionProtocol.ActionID.GenerateToken]?.invoke()
                    currentToken = "tok_1"
                    // active_tokens: active_tokens \union {"tok_1"} (set by action)
                    state = SessionBackendState.GenerateToken
                    emptyList()
                }
            state == SessionBackendState.GenerateToken && ev == SessionProtocol.EventID.TokenCreated ->
                run {
                    actions[SessionProtocol.ActionID.RegisterRelay]?.invoke()
                    state = SessionBackendState.RegisterRelay
                    emptyList()
                }
            state == SessionBackendState.RegisterRelay && ev == SessionProtocol.EventID.RelayRegistered ->
                run {
                    secretPublished = true
                    state = SessionBackendState.WaitingForClient
                    emptyList()
                }
            state == SessionBackendState.WaitingForClient && ev == SessionProtocol.EventID.RecvPairHello && guards[SessionProtocol.GuardID.TokenValid]?.invoke() == true ->
                run {
                    actions[SessionProtocol.ActionID.DeriveSecret]?.invoke()
                    // received_client_pub: recv_msg.pubkey (set by action)
                    backendEcdhPub = "backend_pub"
                    // backend_shared_key: DeriveKey("backend_pub", recv_msg.pubkey) (set by action)
                    // backend_code: DeriveCode("backend_pub", recv_msg.pubkey) (set by action)
                    state = SessionBackendState.DeriveSecret
                    emptyList()
                }
            state == SessionBackendState.WaitingForClient && ev == SessionProtocol.EventID.RecvPairHello && guards[SessionProtocol.GuardID.TokenInvalid]?.invoke() == true ->
                run {
                    state = SessionBackendState.Idle
                    emptyList()
                }
            state == SessionBackendState.DeriveSecret && ev == SessionProtocol.EventID.EcdhComplete ->
                run {
                    state = SessionBackendState.SendAck
                    emptyList()
                }
            state == SessionBackendState.SendAck && ev == SessionProtocol.EventID.SignalCodeDisplay ->
                run {
                    state = SessionBackendState.WaitingForCode
                    emptyList()
                }
            state == SessionBackendState.WaitingForCode && ev == SessionProtocol.EventID.CliCodeEntered ->
                run {
                    // received_code: cli_entered_code (set by action)
                    state = SessionBackendState.ValidateCode
                    emptyList()
                }
            state == SessionBackendState.ValidateCode && ev == SessionProtocol.EventID.CheckCode && guards[SessionProtocol.GuardID.CodeCorrect]?.invoke() == true ->
                run {
                    state = SessionBackendState.StorePaired
                    emptyList()
                }
            state == SessionBackendState.ValidateCode && ev == SessionProtocol.EventID.CheckCode && guards[SessionProtocol.GuardID.CodeWrong]?.invoke() == true ->
                run {
                    // code_attempts: code_attempts + 1 (set by action)
                    state = SessionBackendState.Idle
                    emptyList()
                }
            state == SessionBackendState.StorePaired && ev == SessionProtocol.EventID.Finalise ->
                run {
                    actions[SessionProtocol.ActionID.StoreDevice]?.invoke()
                    deviceSecret = "dev_secret_1"
                    // paired_devices: paired_devices \union {"device_1"} (set by action)
                    // active_tokens: active_tokens \ {current_token} (set by action)
                    // used_tokens: used_tokens \union {current_token} (set by action)
                    state = SessionBackendState.Paired
                    emptyList()
                }
            state == SessionBackendState.Paired && ev == SessionProtocol.EventID.RecvAuthRequest ->
                run {
                    // received_device_id: recv_msg.device_id (set by action)
                    // received_auth_nonce: recv_msg.nonce (set by action)
                    state = SessionBackendState.AuthCheck
                    emptyList()
                }
            state == SessionBackendState.AuthCheck && ev == SessionProtocol.EventID.Verify && guards[SessionProtocol.GuardID.DeviceKnown]?.invoke() == true ->
                run {
                    actions[SessionProtocol.ActionID.VerifyDevice]?.invoke()
                    // auth_nonces_used: auth_nonces_used \union {received_auth_nonce} (set by action)
                    state = SessionBackendState.SessionActive
                    emptyList()
                }
            state == SessionBackendState.AuthCheck && ev == SessionProtocol.EventID.Verify && guards[SessionProtocol.GuardID.DeviceUnknown]?.invoke() == true ->
                run {
                    state = SessionBackendState.Idle
                    emptyList()
                }
            state == SessionBackendState.SessionActive && ev == SessionProtocol.EventID.SessionEstablished ->
                run {
                    state = SessionBackendState.RelayConnected
                    emptyList()
                }
            state == SessionBackendState.RelayConnected && ev == SessionProtocol.EventID.CandidatesGathered ->
                run {
                    state = SessionBackendState.CandidatesAdvertised
                    listOf(SessionProtocol.CmdID.SendCandidates)
                }
            state == SessionBackendState.CandidatesAdvertised && ev == SessionProtocol.EventID.RecvPairCheck && guards[SessionProtocol.GuardID.ChallengeValid]?.invoke() == true ->
                run {
                    actions[SessionProtocol.ActionID.ActivateLan]?.invoke()
                    pingFailures = 0
                    backoffLevel = 0
                    bActivePath = "alt"
                    bDispatcherPath = "alt"
                    monitorTarget = "alt"
                    altSignal = "ready"
                    state = SessionBackendState.AltActive
                    listOf(SessionProtocol.CmdID.SendPairCheckAck, SessionProtocol.CmdID.StartAltStreamReader, SessionProtocol.CmdID.StartAltDgReader, SessionProtocol.CmdID.StartMonitor, SessionProtocol.CmdID.SignalAltReady, SessionProtocol.CmdID.SetCryptoDatagram)
                }
            state == SessionBackendState.CandidatesAdvertised && ev == SessionProtocol.EventID.RecvPairCheck && guards[SessionProtocol.GuardID.ChallengeInvalid]?.invoke() == true ->
                run {
                    state = SessionBackendState.RelayConnected
                    emptyList()
                }
            state == SessionBackendState.CandidatesAdvertised && ev == SessionProtocol.EventID.CandidatesTimeout ->
                run {
                    // backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
                    altSignal = "pending"
                    state = SessionBackendState.RelayBackoff
                    listOf(SessionProtocol.CmdID.ResetAltReady, SessionProtocol.CmdID.StartBackoffTimer)
                }
            state == SessionBackendState.AltActive && ev == SessionProtocol.EventID.PingTick ->
                run {
                    state = SessionBackendState.AltActive
                    listOf(SessionProtocol.CmdID.SendPathPing, SessionProtocol.CmdID.StartPongTimeout)
                }
            state == SessionBackendState.AltActive && ev == SessionProtocol.EventID.PingTimeout ->
                run {
                    pingFailures = 1
                    state = SessionBackendState.AltDegraded
                    emptyList()
                }
            state == SessionBackendState.AltDegraded && ev == SessionProtocol.EventID.PingTick ->
                run {
                    state = SessionBackendState.AltDegraded
                    listOf(SessionProtocol.CmdID.SendPathPing, SessionProtocol.CmdID.StartPongTimeout)
                }
            state == SessionBackendState.AltActive && ev == SessionProtocol.EventID.AltStreamError ->
                run {
                    actions[SessionProtocol.ActionID.FallbackToRelay]?.invoke()
                    // backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
                    bActivePath = "relay"
                    bDispatcherPath = "relay"
                    monitorTarget = "none"
                    altSignal = "pending"
                    pingFailures = 0
                    state = SessionBackendState.RelayBackoff
                    listOf(SessionProtocol.CmdID.StopMonitor, SessionProtocol.CmdID.StopAltStreamReader, SessionProtocol.CmdID.StopAltDgReader, SessionProtocol.CmdID.CloseAltPath, SessionProtocol.CmdID.ResetAltReady, SessionProtocol.CmdID.StartBackoffTimer)
                }
            state == SessionBackendState.AltDegraded && ev == SessionProtocol.EventID.AltStreamError ->
                run {
                    actions[SessionProtocol.ActionID.FallbackToRelay]?.invoke()
                    // backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
                    bActivePath = "relay"
                    bDispatcherPath = "relay"
                    monitorTarget = "none"
                    altSignal = "pending"
                    pingFailures = 0
                    state = SessionBackendState.RelayBackoff
                    listOf(SessionProtocol.CmdID.StopMonitor, SessionProtocol.CmdID.StopAltStreamReader, SessionProtocol.CmdID.StopAltDgReader, SessionProtocol.CmdID.CloseAltPath, SessionProtocol.CmdID.ResetAltReady, SessionProtocol.CmdID.StartBackoffTimer)
                }
            state == SessionBackendState.AltDegraded && ev == SessionProtocol.EventID.RecvPathPong ->
                run {
                    actions[SessionProtocol.ActionID.ResetFailures]?.invoke()
                    pingFailures = 0
                    state = SessionBackendState.AltActive
                    listOf(SessionProtocol.CmdID.CancelPongTimeout)
                }
            state == SessionBackendState.AltDegraded && ev == SessionProtocol.EventID.PingTimeout && guards[SessionProtocol.GuardID.UnderMaxFailures]?.invoke() == true ->
                run {
                    // ping_failures: ping_failures + 1 (set by action)
                    state = SessionBackendState.AltDegraded
                    emptyList()
                }
            state == SessionBackendState.AltDegraded && ev == SessionProtocol.EventID.PingTimeout && guards[SessionProtocol.GuardID.AtMaxFailures]?.invoke() == true ->
                run {
                    actions[SessionProtocol.ActionID.FallbackToRelay]?.invoke()
                    // backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
                    bActivePath = "relay"
                    bDispatcherPath = "relay"
                    monitorTarget = "none"
                    altSignal = "pending"
                    pingFailures = 0
                    state = SessionBackendState.RelayBackoff
                    listOf(SessionProtocol.CmdID.StopMonitor, SessionProtocol.CmdID.StopAltStreamReader, SessionProtocol.CmdID.StopAltDgReader, SessionProtocol.CmdID.CloseAltPath, SessionProtocol.CmdID.ResetAltReady, SessionProtocol.CmdID.StartBackoffTimer)
                }
            state == SessionBackendState.RelayBackoff && ev == SessionProtocol.EventID.BackoffExpired ->
                run {
                    state = SessionBackendState.CandidatesAdvertised
                    listOf(SessionProtocol.CmdID.SendCandidates)
                }
            state == SessionBackendState.RelayBackoff && ev == SessionProtocol.EventID.CandidatesChanged ->
                run {
                    backoffLevel = 0
                    state = SessionBackendState.CandidatesAdvertised
                    listOf(SessionProtocol.CmdID.SendCandidates)
                }
            state == SessionBackendState.RelayConnected && ev == SessionProtocol.EventID.CandidatesRefreshTick && guards[SessionProtocol.GuardID.LocalCandidatesAvailable]?.invoke() == true ->
                run {
                    state = SessionBackendState.CandidatesAdvertised
                    listOf(SessionProtocol.CmdID.SendCandidates)
                }
            state == SessionBackendState.CandidatesAdvertised && ev == SessionProtocol.EventID.AppForceFallback ->
                run {
                    altSignal = "pending"
                    state = SessionBackendState.RelayConnected
                    listOf(SessionProtocol.CmdID.ResetAltReady)
                }
            state == SessionBackendState.AltActive && ev == SessionProtocol.EventID.AppForceFallback ->
                run {
                    actions[SessionProtocol.ActionID.FallbackToRelay]?.invoke()
                    // backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
                    bActivePath = "relay"
                    bDispatcherPath = "relay"
                    monitorTarget = "none"
                    altSignal = "pending"
                    pingFailures = 0
                    state = SessionBackendState.RelayBackoff
                    listOf(SessionProtocol.CmdID.StopMonitor, SessionProtocol.CmdID.CancelPongTimeout, SessionProtocol.CmdID.StopAltStreamReader, SessionProtocol.CmdID.StopAltDgReader, SessionProtocol.CmdID.CloseAltPath, SessionProtocol.CmdID.ResetAltReady, SessionProtocol.CmdID.StartBackoffTimer)
                }
            state == SessionBackendState.AltDegraded && ev == SessionProtocol.EventID.AppForceFallback ->
                run {
                    actions[SessionProtocol.ActionID.FallbackToRelay]?.invoke()
                    // backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
                    bActivePath = "relay"
                    bDispatcherPath = "relay"
                    monitorTarget = "none"
                    altSignal = "pending"
                    pingFailures = 0
                    state = SessionBackendState.RelayBackoff
                    listOf(SessionProtocol.CmdID.StopMonitor, SessionProtocol.CmdID.CancelPongTimeout, SessionProtocol.CmdID.StopAltStreamReader, SessionProtocol.CmdID.StopAltDgReader, SessionProtocol.CmdID.CloseAltPath, SessionProtocol.CmdID.ResetAltReady, SessionProtocol.CmdID.StartBackoffTimer)
                }
            state == SessionBackendState.RelayConnected && ev == SessionProtocol.EventID.Disconnect ->
                run {
                    state = SessionBackendState.Paired
                    emptyList()
                }
            state == SessionBackendState.RelayConnected && ev == SessionProtocol.EventID.AppSend ->
                run {
                    state = SessionBackendState.RelayConnected
                    listOf(SessionProtocol.CmdID.WriteActiveStream)
                }
            state == SessionBackendState.CandidatesAdvertised && ev == SessionProtocol.EventID.AppSend ->
                run {
                    state = SessionBackendState.CandidatesAdvertised
                    listOf(SessionProtocol.CmdID.WriteActiveStream)
                }
            state == SessionBackendState.AltActive && ev == SessionProtocol.EventID.AppSend ->
                run {
                    state = SessionBackendState.AltActive
                    listOf(SessionProtocol.CmdID.WriteActiveStream)
                }
            state == SessionBackendState.AltDegraded && ev == SessionProtocol.EventID.AppSend ->
                run {
                    state = SessionBackendState.AltDegraded
                    listOf(SessionProtocol.CmdID.WriteActiveStream)
                }
            state == SessionBackendState.RelayBackoff && ev == SessionProtocol.EventID.AppSend ->
                run {
                    state = SessionBackendState.RelayBackoff
                    listOf(SessionProtocol.CmdID.WriteActiveStream)
                }
            state == SessionBackendState.RelayConnected && ev == SessionProtocol.EventID.RelayStreamData ->
                run {
                    state = SessionBackendState.RelayConnected
                    listOf(SessionProtocol.CmdID.DeliverRecv)
                }
            state == SessionBackendState.CandidatesAdvertised && ev == SessionProtocol.EventID.RelayStreamData ->
                run {
                    state = SessionBackendState.CandidatesAdvertised
                    listOf(SessionProtocol.CmdID.DeliverRecv)
                }
            state == SessionBackendState.AltActive && ev == SessionProtocol.EventID.RelayStreamData ->
                run {
                    state = SessionBackendState.AltActive
                    listOf(SessionProtocol.CmdID.DeliverRecv)
                }
            state == SessionBackendState.AltDegraded && ev == SessionProtocol.EventID.RelayStreamData ->
                run {
                    state = SessionBackendState.AltDegraded
                    listOf(SessionProtocol.CmdID.DeliverRecv)
                }
            state == SessionBackendState.RelayBackoff && ev == SessionProtocol.EventID.RelayStreamData ->
                run {
                    state = SessionBackendState.RelayBackoff
                    listOf(SessionProtocol.CmdID.DeliverRecv)
                }
            state == SessionBackendState.RelayConnected && ev == SessionProtocol.EventID.RelayStreamError ->
                run {
                    state = SessionBackendState.RelayConnected
                    listOf(SessionProtocol.CmdID.DeliverRecvError)
                }
            state == SessionBackendState.CandidatesAdvertised && ev == SessionProtocol.EventID.RelayStreamError ->
                run {
                    state = SessionBackendState.CandidatesAdvertised
                    listOf(SessionProtocol.CmdID.DeliverRecvError)
                }
            state == SessionBackendState.AltActive && ev == SessionProtocol.EventID.RelayStreamError ->
                run {
                    state = SessionBackendState.AltActive
                    listOf(SessionProtocol.CmdID.DeliverRecvError)
                }
            state == SessionBackendState.AltDegraded && ev == SessionProtocol.EventID.RelayStreamError ->
                run {
                    state = SessionBackendState.AltDegraded
                    listOf(SessionProtocol.CmdID.DeliverRecvError)
                }
            state == SessionBackendState.RelayBackoff && ev == SessionProtocol.EventID.RelayStreamError ->
                run {
                    state = SessionBackendState.RelayBackoff
                    listOf(SessionProtocol.CmdID.DeliverRecvError)
                }
            state == SessionBackendState.RelayConnected && ev == SessionProtocol.EventID.AppSendDatagram ->
                run {
                    state = SessionBackendState.RelayConnected
                    listOf(SessionProtocol.CmdID.SendActiveDatagram)
                }
            state == SessionBackendState.CandidatesAdvertised && ev == SessionProtocol.EventID.AppSendDatagram ->
                run {
                    state = SessionBackendState.CandidatesAdvertised
                    listOf(SessionProtocol.CmdID.SendActiveDatagram)
                }
            state == SessionBackendState.AltActive && ev == SessionProtocol.EventID.AppSendDatagram ->
                run {
                    state = SessionBackendState.AltActive
                    listOf(SessionProtocol.CmdID.SendActiveDatagram)
                }
            state == SessionBackendState.AltDegraded && ev == SessionProtocol.EventID.AppSendDatagram ->
                run {
                    state = SessionBackendState.AltDegraded
                    listOf(SessionProtocol.CmdID.SendActiveDatagram)
                }
            state == SessionBackendState.RelayBackoff && ev == SessionProtocol.EventID.AppSendDatagram ->
                run {
                    state = SessionBackendState.RelayBackoff
                    listOf(SessionProtocol.CmdID.SendActiveDatagram)
                }
            state == SessionBackendState.RelayConnected && ev == SessionProtocol.EventID.RelayDatagram ->
                run {
                    state = SessionBackendState.RelayConnected
                    listOf(SessionProtocol.CmdID.DeliverRecvDatagram)
                }
            state == SessionBackendState.CandidatesAdvertised && ev == SessionProtocol.EventID.RelayDatagram ->
                run {
                    state = SessionBackendState.CandidatesAdvertised
                    listOf(SessionProtocol.CmdID.DeliverRecvDatagram)
                }
            state == SessionBackendState.AltActive && ev == SessionProtocol.EventID.RelayDatagram ->
                run {
                    state = SessionBackendState.AltActive
                    listOf(SessionProtocol.CmdID.DeliverRecvDatagram)
                }
            state == SessionBackendState.AltDegraded && ev == SessionProtocol.EventID.RelayDatagram ->
                run {
                    state = SessionBackendState.AltDegraded
                    listOf(SessionProtocol.CmdID.DeliverRecvDatagram)
                }
            state == SessionBackendState.RelayBackoff && ev == SessionProtocol.EventID.RelayDatagram ->
                run {
                    state = SessionBackendState.RelayBackoff
                    listOf(SessionProtocol.CmdID.DeliverRecvDatagram)
                }
            state == SessionBackendState.AltActive && ev == SessionProtocol.EventID.AltStreamData ->
                run {
                    state = SessionBackendState.AltActive
                    listOf(SessionProtocol.CmdID.DeliverRecv)
                }
            state == SessionBackendState.AltDegraded && ev == SessionProtocol.EventID.AltStreamData ->
                run {
                    state = SessionBackendState.AltDegraded
                    listOf(SessionProtocol.CmdID.DeliverRecv)
                }
            state == SessionBackendState.AltActive && ev == SessionProtocol.EventID.AltDatagram ->
                run {
                    state = SessionBackendState.AltActive
                    listOf(SessionProtocol.CmdID.DeliverRecvDatagram)
                }
            state == SessionBackendState.AltDegraded && ev == SessionProtocol.EventID.AltDatagram ->
                run {
                    state = SessionBackendState.AltDegraded
                    listOf(SessionProtocol.CmdID.DeliverRecvDatagram)
                }
            else -> emptyList()
        }
        return cmds
    }
}

/** SessionClientMachine is the generated state machine for the client actor. */
class SessionClientMachine {
    var state: SessionClientState = SessionClientState.Idle
        private set
    var receivedBackendPub: String = "none" // pubkey client received in pair_hello_ack
    var clientSharedKey: String = "" // ECDH key derived by client
    var clientCode: String = "" // code computed by client
    var cActivePath: String = "relay" // client active path
    var cDispatcherPath: String = "relay" // client datagram dispatcher binding
    var altSignal: String = "pending" // AltReady notification state
    val guards = mutableMapOf<SessionProtocol.GuardID, () -> Boolean>()
    val actions = mutableMapOf<SessionProtocol.ActionID, () -> Unit>()

    /** Handle an event and return the list of commands to execute. */
    fun handleEvent(ev: SessionProtocol.EventID): List<SessionProtocol.CmdID> {
        val cmds: List<SessionProtocol.CmdID> = when {
            state == SessionClientState.Idle && ev == SessionProtocol.EventID.BackchannelReceived ->
                run {
                    state = SessionClientState.ObtainBackchannelSecret
                    emptyList()
                }
            state == SessionClientState.ObtainBackchannelSecret && ev == SessionProtocol.EventID.SecretParsed ->
                run {
                    state = SessionClientState.ConnectRelay
                    emptyList()
                }
            state == SessionClientState.ConnectRelay && ev == SessionProtocol.EventID.RelayConnected ->
                run {
                    state = SessionClientState.GenKeyPair
                    emptyList()
                }
            state == SessionClientState.GenKeyPair && ev == SessionProtocol.EventID.KeyPairGenerated ->
                run {
                    actions[SessionProtocol.ActionID.SendPairHello]?.invoke()
                    state = SessionClientState.WaitAck
                    emptyList()
                }
            state == SessionClientState.WaitAck && ev == SessionProtocol.EventID.RecvPairHelloAck ->
                run {
                    actions[SessionProtocol.ActionID.DeriveSecret]?.invoke()
                    // received_backend_pub: recv_msg.pubkey (set by action)
                    // client_shared_key: DeriveKey("client_pub", recv_msg.pubkey) (set by action)
                    state = SessionClientState.E2EReady
                    emptyList()
                }
            state == SessionClientState.E2EReady && ev == SessionProtocol.EventID.RecvPairConfirm ->
                run {
                    // client_code: DeriveCode(received_backend_pub, "client_pub") (set by action)
                    state = SessionClientState.ShowCode
                    emptyList()
                }
            state == SessionClientState.ShowCode && ev == SessionProtocol.EventID.CodeDisplayed ->
                run {
                    state = SessionClientState.WaitPairComplete
                    emptyList()
                }
            state == SessionClientState.WaitPairComplete && ev == SessionProtocol.EventID.RecvPairComplete ->
                run {
                    actions[SessionProtocol.ActionID.StoreSecret]?.invoke()
                    state = SessionClientState.Paired
                    emptyList()
                }
            state == SessionClientState.Paired && ev == SessionProtocol.EventID.AppLaunch ->
                run {
                    state = SessionClientState.Reconnect
                    emptyList()
                }
            state == SessionClientState.Reconnect && ev == SessionProtocol.EventID.RelayConnected ->
                run {
                    state = SessionClientState.SendAuth
                    emptyList()
                }
            state == SessionClientState.SendAuth && ev == SessionProtocol.EventID.RecvAuthOk ->
                run {
                    state = SessionClientState.SessionActive
                    emptyList()
                }
            state == SessionClientState.SessionActive && ev == SessionProtocol.EventID.SessionEstablished ->
                run {
                    state = SessionClientState.RelayConnected
                    emptyList()
                }
            state == SessionClientState.RelayConnected && ev == SessionProtocol.EventID.RecvCandidates && guards[SessionProtocol.GuardID.AltEnabled]?.invoke() == true ->
                run {
                    actions[SessionProtocol.ActionID.DialCandidate]?.invoke()
                    state = SessionClientState.PairDialing
                    listOf(SessionProtocol.CmdID.DialCandidate)
                }
            state == SessionClientState.RelayConnected && ev == SessionProtocol.EventID.RecvCandidates && guards[SessionProtocol.GuardID.AltDisabled]?.invoke() == true ->
                run {
                    state = SessionClientState.RelayConnected
                    emptyList()
                }
            state == SessionClientState.PairDialing && ev == SessionProtocol.EventID.DialOk ->
                run {
                    state = SessionClientState.PairChecking
                    listOf(SessionProtocol.CmdID.SendPairCheck)
                }
            state == SessionClientState.PairDialing && ev == SessionProtocol.EventID.DialFailed ->
                run {
                    state = SessionClientState.RelayConnected
                    emptyList()
                }
            state == SessionClientState.PairChecking && ev == SessionProtocol.EventID.RecvPairCheckAck ->
                run {
                    actions[SessionProtocol.ActionID.ActivateLan]?.invoke()
                    cActivePath = "alt"
                    cDispatcherPath = "alt"
                    altSignal = "ready"
                    state = SessionClientState.AltActive
                    listOf(SessionProtocol.CmdID.StartAltStreamReader, SessionProtocol.CmdID.StartAltDgReader, SessionProtocol.CmdID.SignalAltReady, SessionProtocol.CmdID.SetCryptoDatagram)
                }
            state == SessionClientState.PairChecking && ev == SessionProtocol.EventID.VerifyTimeout ->
                run {
                    cDispatcherPath = "relay"
                    state = SessionClientState.RelayConnected
                    emptyList()
                }
            state == SessionClientState.AltActive && ev == SessionProtocol.EventID.RecvPathPing ->
                run {
                    state = SessionClientState.AltActive
                    listOf(SessionProtocol.CmdID.SendPathPong)
                }
            state == SessionClientState.AltActive && ev == SessionProtocol.EventID.AltError ->
                run {
                    actions[SessionProtocol.ActionID.FallbackToRelay]?.invoke()
                    cActivePath = "relay"
                    cDispatcherPath = "relay"
                    altSignal = "pending"
                    state = SessionClientState.RelayFallback
                    listOf(SessionProtocol.CmdID.StopAltStreamReader, SessionProtocol.CmdID.StopAltDgReader, SessionProtocol.CmdID.CloseAltPath, SessionProtocol.CmdID.ResetAltReady)
                }
            state == SessionClientState.AltActive && ev == SessionProtocol.EventID.AltStreamError ->
                run {
                    actions[SessionProtocol.ActionID.FallbackToRelay]?.invoke()
                    cActivePath = "relay"
                    cDispatcherPath = "relay"
                    altSignal = "pending"
                    state = SessionClientState.RelayFallback
                    listOf(SessionProtocol.CmdID.StopAltStreamReader, SessionProtocol.CmdID.StopAltDgReader, SessionProtocol.CmdID.CloseAltPath, SessionProtocol.CmdID.ResetAltReady)
                }
            state == SessionClientState.RelayFallback && ev == SessionProtocol.EventID.RelayOk ->
                run {
                    state = SessionClientState.RelayConnected
                    emptyList()
                }
            state == SessionClientState.AltActive && ev == SessionProtocol.EventID.RecvCandidates && guards[SessionProtocol.GuardID.AltEnabled]?.invoke() == true ->
                run {
                    actions[SessionProtocol.ActionID.DialCandidate]?.invoke()
                    state = SessionClientState.PairDialing
                    listOf(SessionProtocol.CmdID.StopAltStreamReader, SessionProtocol.CmdID.StopAltDgReader, SessionProtocol.CmdID.CloseAltPath, SessionProtocol.CmdID.DialCandidate)
                }
            state == SessionClientState.PairDialing && ev == SessionProtocol.EventID.AppForceFallback ->
                run {
                    state = SessionClientState.RelayConnected
                    emptyList()
                }
            state == SessionClientState.PairChecking && ev == SessionProtocol.EventID.AppForceFallback ->
                run {
                    cDispatcherPath = "relay"
                    state = SessionClientState.RelayConnected
                    listOf(SessionProtocol.CmdID.StopAltStreamReader, SessionProtocol.CmdID.StopAltDgReader, SessionProtocol.CmdID.CloseAltPath)
                }
            state == SessionClientState.AltActive && ev == SessionProtocol.EventID.AppForceFallback ->
                run {
                    actions[SessionProtocol.ActionID.FallbackToRelay]?.invoke()
                    cActivePath = "relay"
                    cDispatcherPath = "relay"
                    altSignal = "pending"
                    state = SessionClientState.RelayConnected
                    listOf(SessionProtocol.CmdID.StopAltStreamReader, SessionProtocol.CmdID.StopAltDgReader, SessionProtocol.CmdID.CloseAltPath, SessionProtocol.CmdID.ResetAltReady)
                }
            state == SessionClientState.RelayConnected && ev == SessionProtocol.EventID.Disconnect ->
                run {
                    state = SessionClientState.Paired
                    emptyList()
                }
            state == SessionClientState.RelayConnected && ev == SessionProtocol.EventID.AppSend ->
                run {
                    state = SessionClientState.RelayConnected
                    listOf(SessionProtocol.CmdID.WriteActiveStream)
                }
            state == SessionClientState.PairDialing && ev == SessionProtocol.EventID.AppSend ->
                run {
                    state = SessionClientState.PairDialing
                    listOf(SessionProtocol.CmdID.WriteActiveStream)
                }
            state == SessionClientState.PairChecking && ev == SessionProtocol.EventID.AppSend ->
                run {
                    state = SessionClientState.PairChecking
                    listOf(SessionProtocol.CmdID.WriteActiveStream)
                }
            state == SessionClientState.AltActive && ev == SessionProtocol.EventID.AppSend ->
                run {
                    state = SessionClientState.AltActive
                    listOf(SessionProtocol.CmdID.WriteActiveStream)
                }
            state == SessionClientState.RelayFallback && ev == SessionProtocol.EventID.AppSend ->
                run {
                    state = SessionClientState.RelayFallback
                    listOf(SessionProtocol.CmdID.WriteActiveStream)
                }
            state == SessionClientState.RelayConnected && ev == SessionProtocol.EventID.RelayStreamData ->
                run {
                    state = SessionClientState.RelayConnected
                    listOf(SessionProtocol.CmdID.DeliverRecv)
                }
            state == SessionClientState.PairDialing && ev == SessionProtocol.EventID.RelayStreamData ->
                run {
                    state = SessionClientState.PairDialing
                    listOf(SessionProtocol.CmdID.DeliverRecv)
                }
            state == SessionClientState.PairChecking && ev == SessionProtocol.EventID.RelayStreamData ->
                run {
                    state = SessionClientState.PairChecking
                    listOf(SessionProtocol.CmdID.DeliverRecv)
                }
            state == SessionClientState.AltActive && ev == SessionProtocol.EventID.RelayStreamData ->
                run {
                    state = SessionClientState.AltActive
                    listOf(SessionProtocol.CmdID.DeliverRecv)
                }
            state == SessionClientState.RelayFallback && ev == SessionProtocol.EventID.RelayStreamData ->
                run {
                    state = SessionClientState.RelayFallback
                    listOf(SessionProtocol.CmdID.DeliverRecv)
                }
            state == SessionClientState.RelayConnected && ev == SessionProtocol.EventID.RelayStreamError ->
                run {
                    state = SessionClientState.RelayConnected
                    listOf(SessionProtocol.CmdID.DeliverRecvError)
                }
            state == SessionClientState.PairDialing && ev == SessionProtocol.EventID.RelayStreamError ->
                run {
                    state = SessionClientState.PairDialing
                    listOf(SessionProtocol.CmdID.DeliverRecvError)
                }
            state == SessionClientState.PairChecking && ev == SessionProtocol.EventID.RelayStreamError ->
                run {
                    state = SessionClientState.PairChecking
                    listOf(SessionProtocol.CmdID.DeliverRecvError)
                }
            state == SessionClientState.AltActive && ev == SessionProtocol.EventID.RelayStreamError ->
                run {
                    state = SessionClientState.AltActive
                    listOf(SessionProtocol.CmdID.DeliverRecvError)
                }
            state == SessionClientState.RelayFallback && ev == SessionProtocol.EventID.RelayStreamError ->
                run {
                    state = SessionClientState.RelayFallback
                    listOf(SessionProtocol.CmdID.DeliverRecvError)
                }
            state == SessionClientState.RelayConnected && ev == SessionProtocol.EventID.AppSendDatagram ->
                run {
                    state = SessionClientState.RelayConnected
                    listOf(SessionProtocol.CmdID.SendActiveDatagram)
                }
            state == SessionClientState.PairDialing && ev == SessionProtocol.EventID.AppSendDatagram ->
                run {
                    state = SessionClientState.PairDialing
                    listOf(SessionProtocol.CmdID.SendActiveDatagram)
                }
            state == SessionClientState.PairChecking && ev == SessionProtocol.EventID.AppSendDatagram ->
                run {
                    state = SessionClientState.PairChecking
                    listOf(SessionProtocol.CmdID.SendActiveDatagram)
                }
            state == SessionClientState.AltActive && ev == SessionProtocol.EventID.AppSendDatagram ->
                run {
                    state = SessionClientState.AltActive
                    listOf(SessionProtocol.CmdID.SendActiveDatagram)
                }
            state == SessionClientState.RelayFallback && ev == SessionProtocol.EventID.AppSendDatagram ->
                run {
                    state = SessionClientState.RelayFallback
                    listOf(SessionProtocol.CmdID.SendActiveDatagram)
                }
            state == SessionClientState.RelayConnected && ev == SessionProtocol.EventID.RelayDatagram ->
                run {
                    state = SessionClientState.RelayConnected
                    listOf(SessionProtocol.CmdID.DeliverRecvDatagram)
                }
            state == SessionClientState.PairDialing && ev == SessionProtocol.EventID.RelayDatagram ->
                run {
                    state = SessionClientState.PairDialing
                    listOf(SessionProtocol.CmdID.DeliverRecvDatagram)
                }
            state == SessionClientState.PairChecking && ev == SessionProtocol.EventID.RelayDatagram ->
                run {
                    state = SessionClientState.PairChecking
                    listOf(SessionProtocol.CmdID.DeliverRecvDatagram)
                }
            state == SessionClientState.AltActive && ev == SessionProtocol.EventID.RelayDatagram ->
                run {
                    state = SessionClientState.AltActive
                    listOf(SessionProtocol.CmdID.DeliverRecvDatagram)
                }
            state == SessionClientState.RelayFallback && ev == SessionProtocol.EventID.RelayDatagram ->
                run {
                    state = SessionClientState.RelayFallback
                    listOf(SessionProtocol.CmdID.DeliverRecvDatagram)
                }
            state == SessionClientState.AltActive && ev == SessionProtocol.EventID.AltStreamData ->
                run {
                    state = SessionClientState.AltActive
                    listOf(SessionProtocol.CmdID.DeliverRecv)
                }
            state == SessionClientState.AltActive && ev == SessionProtocol.EventID.AltDatagram ->
                run {
                    state = SessionClientState.AltActive
                    listOf(SessionProtocol.CmdID.DeliverRecvDatagram)
                }
            else -> emptyList()
        }
        return cmds
    }
}

/** SessionRelayMachine is the generated state machine for the relay actor. */
class SessionRelayMachine {
    var state: SessionRelayState = SessionRelayState.Idle
        private set
    var relayBridge: String = "idle" // relay bridge state
    val guards = mutableMapOf<SessionProtocol.GuardID, () -> Boolean>()
    val actions = mutableMapOf<SessionProtocol.ActionID, () -> Unit>()

    /** Handle an event and return the list of commands to execute. */
    fun handleEvent(ev: SessionProtocol.EventID): List<SessionProtocol.CmdID> {
        val cmds: List<SessionProtocol.CmdID> = when {
            state == SessionRelayState.Idle && ev == SessionProtocol.EventID.BackendRegister ->
                run {
                    state = SessionRelayState.BackendRegistered
                    emptyList()
                }
            state == SessionRelayState.BackendRegistered && ev == SessionProtocol.EventID.ClientConnect ->
                run {
                    actions[SessionProtocol.ActionID.BridgeStreams]?.invoke()
                    relayBridge = "active"
                    state = SessionRelayState.Bridged
                    emptyList()
                }
            state == SessionRelayState.Bridged && ev == SessionProtocol.EventID.ClientDisconnect ->
                run {
                    actions[SessionProtocol.ActionID.Unbridge]?.invoke()
                    relayBridge = "idle"
                    state = SessionRelayState.BackendRegistered
                    emptyList()
                }
            state == SessionRelayState.BackendRegistered && ev == SessionProtocol.EventID.BackendDisconnect ->
                run {
                    state = SessionRelayState.Idle
                    emptyList()
                }
            else -> emptyList()
        }
        return cmds
    }
}


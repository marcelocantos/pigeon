// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Auto-generated from protocol definition. Do not edit.
// Source of truth: protocol/*.yaml

package com.marcelocantos.tern.crypto

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
    ResetFailures("reset_failures"),
    FallbackToRelay("fallback_to_relay"),
    SendPairHello("send_pair_hello"),
    StoreSecret("store_secret"),
    DialLan("dial_lan"),
    BridgeStreams("bridge_streams"),
    Unbridge("unbridge");
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
        Transition("LANDegraded", "LANActive", "path_pong", "recv", null, "reset_failures", emptyList()),
        Transition("LANDegraded", "LANDegraded", "ping_timeout", "internal", "under_max_failures", null, emptyList()),
        Transition("LANDegraded", "RelayBackoff", "ping_timeout", "internal", "at_max_failures", "fallback_to_relay", emptyList()),
        Transition("RelayBackoff", "LANOffered", "backoff_expired", "internal", null, null, listOf("client" to "lan_offer")),
        Transition("RelayBackoff", "LANOffered", "lan_server_changed", "internal", null, null, listOf("client" to "lan_offer")),
        Transition("RelayConnected", "LANOffered", "readvertise_tick", "internal", "lan_server_available", null, listOf("client" to "lan_offer")),
        Transition("RelayConnected", "Paired", "disconnect", "internal", null, null, emptyList()),
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
        Transition("RelayFallback", "RelayConnected", "relay_ok", "internal", null, null, emptyList()),
        Transition("LANActive", "LANConnecting", "lan_offer", "recv", "lan_enabled", "dial_lan", emptyList()),
        Transition("RelayConnected", "Paired", "disconnect", "internal", null, null, emptyList()),
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


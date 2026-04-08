// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Auto-generated from protocol definition. Do not edit.
// Source of truth: protocol/*.yaml

package com.marcelocantos.pigeon.crypto

enum class PathSwitchBackendState(val value: String) {
    RelayConnected("RelayConnected"),
    LANOffered("LANOffered"),
    LANActive("LANActive"),
    RelayBackoff("RelayBackoff"),
    LANDegraded("LANDegraded");
}

enum class PathSwitchClientState(val value: String) {
    RelayConnected("RelayConnected"),
    LANConnecting("LANConnecting"),
    LANVerifying("LANVerifying"),
    LANActive("LANActive"),
    RelayFallback("RelayFallback");
}

enum class PathSwitchRelayState(val value: String) {
    Idle("Idle"),
    BackendRegistered("BackendRegistered"),
    Bridged("Bridged");
}

/** The protocol transition table and shared type enums. */
object PathSwitchProtocol {

    enum class MessageType(val value: String) {
        LanOffer("lan_offer"),
        LanVerify("lan_verify"),
        LanConfirm("lan_confirm"),
        PathPing("path_ping"),
        PathPong("path_pong"),
        RelayResume("relay_resume"),
        RelayResumed("relay_resumed");
    }

    enum class GuardID(val value: String) {
        ChallengeValid("challenge_valid"),
        ChallengeInvalid("challenge_invalid"),
        LanEnabled("lan_enabled"),
        LanDisabled("lan_disabled"),
        LanServerAvailable("lan_server_available"),
        UnderMaxFailures("under_max_failures"),
        AtMaxFailures("at_max_failures");
    }

    enum class ActionID(val value: String) {
        ActivateLan("activate_lan"),
        ResetFailures("reset_failures"),
        FallbackToRelay("fallback_to_relay"),
        DialLan("dial_lan"),
        BridgeStreams("bridge_streams"),
        Unbridge("unbridge"),
        RebridgeStreams("rebridge_streams");
    }

    enum class EventID(val value: String) {
        BackendDisconnect("backend_disconnect"),
        BackendRegister("backend_register"),
        BackoffExpired("backoff_expired"),
        ClientConnect("client_connect"),
        ClientDisconnect("client_disconnect"),
        LanDialFailed("lan_dial_failed"),
        LanDialOk("lan_dial_ok"),
        LanError("lan_error"),
        LanServerChanged("lan_server_changed"),
        LanServerReady("lan_server_ready"),
        OfferTimeout("offer_timeout"),
        PingTick("ping_tick"),
        PingTimeout("ping_timeout"),
        ReadvertiseTick("readvertise_tick"),
        RecvLanConfirm("recv_lan_confirm"),
        RecvLanOffer("recv_lan_offer"),
        RecvLanVerify("recv_lan_verify"),
        RecvPathPing("recv_path_ping"),
        RecvPathPong("recv_path_pong"),
        RecvRelayResume("recv_relay_resume"),
        RelayOk("relay_ok"),
        VerifyTimeout("verify_timeout");
    }

    /** backend transition table. */
    object BackendTable {
        val initial = PathSwitchBackendState.RelayConnected

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
        )
    }

    /** client transition table. */
    object ClientTable {
        val initial = PathSwitchClientState.RelayConnected

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
        )
    }

    /** relay transition table. */
    object RelayTable {
        val initial = PathSwitchRelayState.Idle

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
            Transition("Bridged", "Bridged", "relay_resume", "recv", null, "rebridge_streams", listOf("client" to "relay_resumed")),
            Transition("BackendRegistered", "Idle", "backend_disconnect", "internal", null, null, emptyList()),
        )
    }

}

/** PathSwitchBackendMachine is the generated state machine for the backend actor. */
class PathSwitchBackendMachine {
    var state: PathSwitchBackendState = PathSwitchBackendState.RelayConnected
        private set
    var pingFailures: Int = 0 // consecutive failed pings on the direct path
    var backoffLevel: Int = 0 // current exponential backoff level (0 = no backoff)
    var activePath: String = "relay" // "relay" or "lan" — which path carries application traffic
    var dispatcherPath: String = "relay" // which path the datagram dispatcher reads from ("relay", "lan", "none")
    var monitorTarget: String = "none" // which path the health monitor pings ("lan", "none")
    var lanSignal: String = "pending" // LANReady notification state ("pending" = not yet, "ready" = closed/signalled)
    val guards = mutableMapOf<PathSwitchProtocol.GuardID, () -> Boolean>()
    val actions = mutableMapOf<PathSwitchProtocol.ActionID, () -> Unit>()

    /** Handle an event and return the list of commands to execute. */
    fun handleEvent(ev: PathSwitchProtocol.EventID): List<String> {
        val cmds: List<String> = when {
            state == PathSwitchBackendState.RelayConnected && ev == PathSwitchProtocol.EventID.LanServerReady ->
                run {
                    state = PathSwitchBackendState.LANOffered
                    emptyList()
                }
            state == PathSwitchBackendState.LANOffered && ev == PathSwitchProtocol.EventID.RecvLanVerify && guards[PathSwitchProtocol.GuardID.ChallengeValid]?.invoke() == true ->
                run {
                    actions[PathSwitchProtocol.ActionID.ActivateLan]?.invoke()
                    pingFailures = 0
                    backoffLevel = 0
                    activePath = "lan"
                    monitorTarget = "lan"
                    dispatcherPath = "lan"
                    lanSignal = "ready"
                    state = PathSwitchBackendState.LANActive
                    emptyList()
                }
            state == PathSwitchBackendState.LANOffered && ev == PathSwitchProtocol.EventID.RecvLanVerify && guards[PathSwitchProtocol.GuardID.ChallengeInvalid]?.invoke() == true ->
                run {
                    state = PathSwitchBackendState.RelayConnected
                    emptyList()
                }
            state == PathSwitchBackendState.LANOffered && ev == PathSwitchProtocol.EventID.OfferTimeout ->
                run {
                    // backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
                    state = PathSwitchBackendState.RelayBackoff
                    emptyList()
                }
            state == PathSwitchBackendState.LANActive && ev == PathSwitchProtocol.EventID.PingTick ->
                run {
                    state = PathSwitchBackendState.LANActive
                    emptyList()
                }
            state == PathSwitchBackendState.LANActive && ev == PathSwitchProtocol.EventID.PingTimeout ->
                run {
                    pingFailures = 1
                    state = PathSwitchBackendState.LANDegraded
                    emptyList()
                }
            state == PathSwitchBackendState.LANDegraded && ev == PathSwitchProtocol.EventID.PingTick ->
                run {
                    state = PathSwitchBackendState.LANDegraded
                    emptyList()
                }
            state == PathSwitchBackendState.LANDegraded && ev == PathSwitchProtocol.EventID.RecvPathPong ->
                run {
                    actions[PathSwitchProtocol.ActionID.ResetFailures]?.invoke()
                    pingFailures = 0
                    state = PathSwitchBackendState.LANActive
                    emptyList()
                }
            state == PathSwitchBackendState.LANDegraded && ev == PathSwitchProtocol.EventID.PingTimeout && guards[PathSwitchProtocol.GuardID.UnderMaxFailures]?.invoke() == true ->
                run {
                    // ping_failures: ping_failures + 1 (set by action)
                    state = PathSwitchBackendState.LANDegraded
                    emptyList()
                }
            state == PathSwitchBackendState.LANDegraded && ev == PathSwitchProtocol.EventID.PingTimeout && guards[PathSwitchProtocol.GuardID.AtMaxFailures]?.invoke() == true ->
                run {
                    actions[PathSwitchProtocol.ActionID.FallbackToRelay]?.invoke()
                    // backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
                    activePath = "relay"
                    monitorTarget = "none"
                    dispatcherPath = "relay"
                    lanSignal = "pending"
                    pingFailures = 0
                    state = PathSwitchBackendState.RelayBackoff
                    emptyList()
                }
            state == PathSwitchBackendState.RelayBackoff && ev == PathSwitchProtocol.EventID.BackoffExpired ->
                run {
                    state = PathSwitchBackendState.LANOffered
                    emptyList()
                }
            state == PathSwitchBackendState.RelayBackoff && ev == PathSwitchProtocol.EventID.LanServerChanged ->
                run {
                    backoffLevel = 0
                    state = PathSwitchBackendState.LANOffered
                    emptyList()
                }
            state == PathSwitchBackendState.RelayConnected && ev == PathSwitchProtocol.EventID.ReadvertiseTick && guards[PathSwitchProtocol.GuardID.LanServerAvailable]?.invoke() == true ->
                run {
                    state = PathSwitchBackendState.LANOffered
                    emptyList()
                }
            else -> emptyList()
        }
        return cmds
    }
}

/** PathSwitchClientMachine is the generated state machine for the client actor. */
class PathSwitchClientMachine {
    var state: PathSwitchClientState = PathSwitchClientState.RelayConnected
        private set
    var activePath: String = "relay" // "relay" or "lan" — which path carries application traffic
    var dispatcherPath: String = "relay" // which path the datagram dispatcher reads from ("relay", "lan", "none")
    var lanSignal: String = "pending" // LANReady notification state ("pending" = not yet, "ready" = closed/signalled)
    val guards = mutableMapOf<PathSwitchProtocol.GuardID, () -> Boolean>()
    val actions = mutableMapOf<PathSwitchProtocol.ActionID, () -> Unit>()

    /** Handle an event and return the list of commands to execute. */
    fun handleEvent(ev: PathSwitchProtocol.EventID): List<String> {
        val cmds: List<String> = when {
            state == PathSwitchClientState.RelayConnected && ev == PathSwitchProtocol.EventID.RecvLanOffer && guards[PathSwitchProtocol.GuardID.LanEnabled]?.invoke() == true ->
                run {
                    actions[PathSwitchProtocol.ActionID.DialLan]?.invoke()
                    state = PathSwitchClientState.LANConnecting
                    emptyList()
                }
            state == PathSwitchClientState.RelayConnected && ev == PathSwitchProtocol.EventID.RecvLanOffer && guards[PathSwitchProtocol.GuardID.LanDisabled]?.invoke() == true ->
                run {
                    state = PathSwitchClientState.RelayConnected
                    emptyList()
                }
            state == PathSwitchClientState.LANConnecting && ev == PathSwitchProtocol.EventID.LanDialOk ->
                run {
                    state = PathSwitchClientState.LANVerifying
                    emptyList()
                }
            state == PathSwitchClientState.LANConnecting && ev == PathSwitchProtocol.EventID.LanDialFailed ->
                run {
                    state = PathSwitchClientState.RelayConnected
                    emptyList()
                }
            state == PathSwitchClientState.LANVerifying && ev == PathSwitchProtocol.EventID.RecvLanConfirm ->
                run {
                    actions[PathSwitchProtocol.ActionID.ActivateLan]?.invoke()
                    activePath = "lan"
                    dispatcherPath = "lan"
                    lanSignal = "ready"
                    state = PathSwitchClientState.LANActive
                    emptyList()
                }
            state == PathSwitchClientState.LANVerifying && ev == PathSwitchProtocol.EventID.VerifyTimeout ->
                run {
                    dispatcherPath = "relay"
                    state = PathSwitchClientState.RelayConnected
                    emptyList()
                }
            state == PathSwitchClientState.LANActive && ev == PathSwitchProtocol.EventID.RecvPathPing ->
                run {
                    state = PathSwitchClientState.LANActive
                    emptyList()
                }
            state == PathSwitchClientState.LANActive && ev == PathSwitchProtocol.EventID.LanError ->
                run {
                    actions[PathSwitchProtocol.ActionID.FallbackToRelay]?.invoke()
                    activePath = "relay"
                    dispatcherPath = "relay"
                    lanSignal = "pending"
                    state = PathSwitchClientState.RelayFallback
                    emptyList()
                }
            state == PathSwitchClientState.RelayFallback && ev == PathSwitchProtocol.EventID.RelayOk ->
                run {
                    state = PathSwitchClientState.RelayConnected
                    emptyList()
                }
            state == PathSwitchClientState.LANActive && ev == PathSwitchProtocol.EventID.RecvLanOffer && guards[PathSwitchProtocol.GuardID.LanEnabled]?.invoke() == true ->
                run {
                    actions[PathSwitchProtocol.ActionID.DialLan]?.invoke()
                    state = PathSwitchClientState.LANConnecting
                    emptyList()
                }
            else -> emptyList()
        }
        return cmds
    }
}

/** PathSwitchRelayMachine is the generated state machine for the relay actor. */
class PathSwitchRelayMachine {
    var state: PathSwitchRelayState = PathSwitchRelayState.Idle
        private set
    var relayBridge: String = "idle" // relay bridge state ("active" = bridging, "idle" = backend registered but no client)
    val guards = mutableMapOf<PathSwitchProtocol.GuardID, () -> Boolean>()
    val actions = mutableMapOf<PathSwitchProtocol.ActionID, () -> Unit>()

    /** Handle an event and return the list of commands to execute. */
    fun handleEvent(ev: PathSwitchProtocol.EventID): List<String> {
        val cmds: List<String> = when {
            state == PathSwitchRelayState.Idle && ev == PathSwitchProtocol.EventID.BackendRegister ->
                run {
                    state = PathSwitchRelayState.BackendRegistered
                    emptyList()
                }
            state == PathSwitchRelayState.BackendRegistered && ev == PathSwitchProtocol.EventID.ClientConnect ->
                run {
                    actions[PathSwitchProtocol.ActionID.BridgeStreams]?.invoke()
                    relayBridge = "active"
                    state = PathSwitchRelayState.Bridged
                    emptyList()
                }
            state == PathSwitchRelayState.Bridged && ev == PathSwitchProtocol.EventID.ClientDisconnect ->
                run {
                    actions[PathSwitchProtocol.ActionID.Unbridge]?.invoke()
                    relayBridge = "idle"
                    state = PathSwitchRelayState.BackendRegistered
                    emptyList()
                }
            state == PathSwitchRelayState.Bridged && ev == PathSwitchProtocol.EventID.RecvRelayResume ->
                run {
                    actions[PathSwitchProtocol.ActionID.RebridgeStreams]?.invoke()
                    state = PathSwitchRelayState.Bridged
                    emptyList()
                }
            state == PathSwitchRelayState.BackendRegistered && ev == PathSwitchProtocol.EventID.BackendDisconnect ->
                run {
                    state = PathSwitchRelayState.Idle
                    emptyList()
                }
            else -> emptyList()
        }
        return cmds
    }
}


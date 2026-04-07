// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Auto-generated from protocol definition. Do not edit.
// Source of truth: protocol/*.yaml

import Foundation

public enum PathSwitchBackendState: String, Sendable {
    case relayConnected = "RelayConnected"
    case lANOffered = "LANOffered"
    case lANActive = "LANActive"
    case relayBackoff = "RelayBackoff"
    case lANDegraded = "LANDegraded"
}

public enum PathSwitchClientState: String, Sendable {
    case relayConnected = "RelayConnected"
    case lANConnecting = "LANConnecting"
    case lANVerifying = "LANVerifying"
    case lANActive = "LANActive"
    case relayFallback = "RelayFallback"
}

public enum PathSwitchRelayState: String, Sendable {
    case idle = "Idle"
    case backendRegistered = "BackendRegistered"
    case bridged = "Bridged"
}

/// The protocol transition table and shared type enums.
public enum PathSwitchProtocol {

    public enum MessageType: String, Sendable {
        case lanOffer = "lan_offer"
        case lanVerify = "lan_verify"
        case lanConfirm = "lan_confirm"
        case pathPing = "path_ping"
        case pathPong = "path_pong"
        case relayResume = "relay_resume"
        case relayResumed = "relay_resumed"
    }

    public enum GuardID: String, Sendable {
        case challengeValid = "challenge_valid"
        case challengeInvalid = "challenge_invalid"
        case lanEnabled = "lan_enabled"
        case lanDisabled = "lan_disabled"
        case lanServerAvailable = "lan_server_available"
        case underMaxFailures = "under_max_failures"
        case atMaxFailures = "at_max_failures"
    }

    public enum ActionID: String, Sendable {
        case activateLan = "activate_lan"
        case resetFailures = "reset_failures"
        case fallbackToRelay = "fallback_to_relay"
        case dialLan = "dial_lan"
        case bridgeStreams = "bridge_streams"
        case unbridge = "unbridge"
        case rebridgeStreams = "rebridge_streams"
    }

    public enum EventID: String, Sendable {
        case lanServerReady = "lan_server_ready"
        case offerTimeout = "offer_timeout"
        case pingTick = "ping_tick"
        case pingTimeout = "ping_timeout"
        case backoffExpired = "backoff_expired"
        case lanServerChanged = "lan_server_changed"
        case readvertiseTick = "readvertise_tick"
        case lanDialOk = "lan_dial_ok"
        case lanDialFailed = "lan_dial_failed"
        case verifyTimeout = "verify_timeout"
        case lanError = "lan_error"
        case relayOk = "relay_ok"
        case backendRegister = "backend_register"
        case clientConnect = "client_connect"
        case clientDisconnect = "client_disconnect"
        case backendDisconnect = "backend_disconnect"
        case recvLanVerify = "recv_lan_verify"
        case recvPathPong = "recv_path_pong"
        case recvLanOffer = "recv_lan_offer"
        case recvLanConfirm = "recv_lan_confirm"
        case recvPathPing = "recv_path_ping"
        case recvRelayResume = "recv_relay_resume"
    }


    /// backend transitions.
    public static let backendInitial: PathSwitchBackendState = .relayConnected

    public static let backendTransitions: [(from: String, to: String, on: String, onKind: String, guard: String?, action: String?, sends: [(to: String, msg: String)])] = [
        (from: "RelayConnected", to: "LANOffered", on: "lan_server_ready", onKind: "internal", guard: nil, action: nil, sends: [(to: "client", msg: "lan_offer")]),
        (from: "LANOffered", to: "LANActive", on: "lan_verify", onKind: "recv", guard: "challenge_valid", action: "activate_lan", sends: [(to: "client", msg: "lan_confirm")]),
        (from: "LANOffered", to: "RelayConnected", on: "lan_verify", onKind: "recv", guard: "challenge_invalid", action: nil, sends: []),
        (from: "LANOffered", to: "RelayBackoff", on: "offer_timeout", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "LANActive", to: "LANActive", on: "ping_tick", onKind: "internal", guard: nil, action: nil, sends: [(to: "client", msg: "path_ping")]),
        (from: "LANActive", to: "LANDegraded", on: "ping_timeout", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "LANDegraded", to: "LANDegraded", on: "ping_tick", onKind: "internal", guard: nil, action: nil, sends: [(to: "client", msg: "path_ping")]),
        (from: "LANDegraded", to: "LANActive", on: "path_pong", onKind: "recv", guard: nil, action: "reset_failures", sends: []),
        (from: "LANDegraded", to: "LANDegraded", on: "ping_timeout", onKind: "internal", guard: "under_max_failures", action: nil, sends: []),
        (from: "LANDegraded", to: "RelayBackoff", on: "ping_timeout", onKind: "internal", guard: "at_max_failures", action: "fallback_to_relay", sends: []),
        (from: "RelayBackoff", to: "LANOffered", on: "backoff_expired", onKind: "internal", guard: nil, action: nil, sends: [(to: "client", msg: "lan_offer")]),
        (from: "RelayBackoff", to: "LANOffered", on: "lan_server_changed", onKind: "internal", guard: nil, action: nil, sends: [(to: "client", msg: "lan_offer")]),
        (from: "RelayConnected", to: "LANOffered", on: "readvertise_tick", onKind: "internal", guard: "lan_server_available", action: nil, sends: [(to: "client", msg: "lan_offer")]),
    ]

    /// client transitions.
    public static let clientInitial: PathSwitchClientState = .relayConnected

    public static let clientTransitions: [(from: String, to: String, on: String, onKind: String, guard: String?, action: String?, sends: [(to: String, msg: String)])] = [
        (from: "RelayConnected", to: "LANConnecting", on: "lan_offer", onKind: "recv", guard: "lan_enabled", action: "dial_lan", sends: []),
        (from: "RelayConnected", to: "RelayConnected", on: "lan_offer", onKind: "recv", guard: "lan_disabled", action: nil, sends: []),
        (from: "LANConnecting", to: "LANVerifying", on: "lan_dial_ok", onKind: "internal", guard: nil, action: nil, sends: [(to: "backend", msg: "lan_verify")]),
        (from: "LANConnecting", to: "RelayConnected", on: "lan_dial_failed", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "LANVerifying", to: "LANActive", on: "lan_confirm", onKind: "recv", guard: nil, action: "activate_lan", sends: []),
        (from: "LANVerifying", to: "RelayConnected", on: "verify_timeout", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "LANActive", to: "LANActive", on: "path_ping", onKind: "recv", guard: nil, action: nil, sends: [(to: "backend", msg: "path_pong")]),
        (from: "LANActive", to: "RelayFallback", on: "lan_error", onKind: "internal", guard: nil, action: "fallback_to_relay", sends: []),
        (from: "RelayFallback", to: "RelayConnected", on: "relay_ok", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "LANActive", to: "LANConnecting", on: "lan_offer", onKind: "recv", guard: "lan_enabled", action: "dial_lan", sends: []),
    ]

    /// relay transitions.
    public static let relayInitial: PathSwitchRelayState = .idle

    public static let relayTransitions: [(from: String, to: String, on: String, onKind: String, guard: String?, action: String?, sends: [(to: String, msg: String)])] = [
        (from: "Idle", to: "BackendRegistered", on: "backend_register", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "BackendRegistered", to: "Bridged", on: "client_connect", onKind: "internal", guard: nil, action: "bridge_streams", sends: []),
        (from: "Bridged", to: "BackendRegistered", on: "client_disconnect", onKind: "internal", guard: nil, action: "unbridge", sends: []),
        (from: "Bridged", to: "Bridged", on: "relay_resume", onKind: "recv", guard: nil, action: "rebridge_streams", sends: [(to: "client", msg: "relay_resumed")]),
        (from: "BackendRegistered", to: "Idle", on: "backend_disconnect", onKind: "internal", guard: nil, action: nil, sends: []),
    ]
}

/// PathSwitchBackendMachine is the generated state machine for the backend actor.
public final class PathSwitchBackendMachine: @unchecked Sendable {
    public typealias MessageType = PathSwitchProtocol.MessageType
    public typealias GuardID = PathSwitchProtocol.GuardID
    public typealias ActionID = PathSwitchProtocol.ActionID
    public typealias EventID = PathSwitchProtocol.EventID

    public private(set) var state: PathSwitchBackendState
    public var pingFailures: Int // consecutive failed pings on the direct path
    public var backoffLevel: Int // current exponential backoff level (0 = no backoff)
    public var activePath: String // "relay" or "lan" — which path carries application traffic
    public var dispatcherPath: String // which path the datagram dispatcher reads from ("relay", "lan", "none")
    public var monitorTarget: String // which path the health monitor pings ("lan", "none")
    public var lanSignal: String // LANReady notification state ("pending" = not yet, "ready" = closed/signalled)

    public var guards: [GuardID: () -> Bool] = [:]
    public var actions: [ActionID: () throws -> Void] = [:]

    public init() {
        self.state = .relayConnected
        self.pingFailures = 0
        self.backoffLevel = 0
        self.activePath = "relay"
        self.dispatcherPath = "relay"
        self.monitorTarget = "none"
        self.lanSignal = "pending"
    }

    /// Handle any event (message receipt or internal). Returns emitted commands.
    @discardableResult
    public func handleEvent(_ ev: EventID) throws -> [String] {
        switch (state, ev) {
        case (.relayConnected, .lanServerReady):
            state = .lANOffered
            return []
        case (.lANOffered, .recvLanVerify) where guards[.challengeValid]?() == true:
            try actions[.activateLan]?()
            pingFailures = 0
            backoffLevel = 0
            activePath = "lan"
            monitorTarget = "lan"
            dispatcherPath = "lan"
            lanSignal = "ready"
            state = .lANActive
            return []
        case (.lANOffered, .recvLanVerify) where guards[.challengeInvalid]?() == true:
            state = .relayConnected
            return []
        case (.lANOffered, .offerTimeout):
            // backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
            state = .relayBackoff
            return []
        case (.lANActive, .pingTick):
            state = .lANActive
            return []
        case (.lANActive, .pingTimeout):
            pingFailures = 1
            state = .lANDegraded
            return []
        case (.lANDegraded, .pingTick):
            state = .lANDegraded
            return []
        case (.lANDegraded, .recvPathPong):
            try actions[.resetFailures]?()
            pingFailures = 0
            state = .lANActive
            return []
        case (.lANDegraded, .pingTimeout) where guards[.underMaxFailures]?() == true:
            // ping_failures: ping_failures + 1 (set by action)
            state = .lANDegraded
            return []
        case (.lANDegraded, .pingTimeout) where guards[.atMaxFailures]?() == true:
            try actions[.fallbackToRelay]?()
            // backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
            activePath = "relay"
            monitorTarget = "none"
            dispatcherPath = "relay"
            lanSignal = "pending"
            pingFailures = 0
            state = .relayBackoff
            return []
        case (.relayBackoff, .backoffExpired):
            state = .lANOffered
            return []
        case (.relayBackoff, .lanServerChanged):
            backoffLevel = 0
            state = .lANOffered
            return []
        case (.relayConnected, .readvertiseTick) where guards[.lanServerAvailable]?() == true:
            state = .lANOffered
            return []
        default:
            return []
        }
    }

    /// Process a received message. Returns the new state, or nil if rejected.
    @discardableResult
    public func handleMessage(_ msg: MessageType) throws -> PathSwitchBackendState? {
        switch (state, msg) {
        case (.lANOffered, .lanVerify) where guards[.challengeValid]?() == true:
            try actions[.activateLan]?()
            pingFailures = 0
            backoffLevel = 0
            activePath = "lan"
            monitorTarget = "lan"
            dispatcherPath = "lan"
            lanSignal = "ready"
            state = .lANActive
            return state
        case (.lANOffered, .lanVerify) where guards[.challengeInvalid]?() == true:
            state = .relayConnected
            return state
        case (.lANDegraded, .pathPong):
            try actions[.resetFailures]?()
            pingFailures = 0
            state = .lANActive
            return state
        default:
            return nil
        }
    }

    /// Attempt an internal transition. Returns the new state, or nil if none available.
    @discardableResult
    public func step() throws -> PathSwitchBackendState? {
        switch state {
        case .lANOffered:
            // backoff_level: Min(backoff_level + 1, max_backoff_level) (set by action)
            state = .relayBackoff
            return state
        default:
            return nil
        }
    }
}

/// PathSwitchClientMachine is the generated state machine for the client actor.
public final class PathSwitchClientMachine: @unchecked Sendable {
    public typealias MessageType = PathSwitchProtocol.MessageType
    public typealias GuardID = PathSwitchProtocol.GuardID
    public typealias ActionID = PathSwitchProtocol.ActionID
    public typealias EventID = PathSwitchProtocol.EventID

    public private(set) var state: PathSwitchClientState
    public var activePath: String // "relay" or "lan" — which path carries application traffic
    public var dispatcherPath: String // which path the datagram dispatcher reads from ("relay", "lan", "none")
    public var lanSignal: String // LANReady notification state ("pending" = not yet, "ready" = closed/signalled)

    public var guards: [GuardID: () -> Bool] = [:]
    public var actions: [ActionID: () throws -> Void] = [:]

    public init() {
        self.state = .relayConnected
        self.activePath = "relay"
        self.dispatcherPath = "relay"
        self.lanSignal = "pending"
    }

    /// Handle any event (message receipt or internal). Returns emitted commands.
    @discardableResult
    public func handleEvent(_ ev: EventID) throws -> [String] {
        switch (state, ev) {
        case (.relayConnected, .recvLanOffer) where guards[.lanEnabled]?() == true:
            try actions[.dialLan]?()
            state = .lANConnecting
            return []
        case (.relayConnected, .recvLanOffer) where guards[.lanDisabled]?() == true:
            state = .relayConnected
            return []
        case (.lANConnecting, .lanDialOk):
            state = .lANVerifying
            return []
        case (.lANConnecting, .lanDialFailed):
            state = .relayConnected
            return []
        case (.lANVerifying, .recvLanConfirm):
            try actions[.activateLan]?()
            activePath = "lan"
            dispatcherPath = "lan"
            lanSignal = "ready"
            state = .lANActive
            return []
        case (.lANVerifying, .verifyTimeout):
            dispatcherPath = "relay"
            state = .relayConnected
            return []
        case (.lANActive, .recvPathPing):
            state = .lANActive
            return []
        case (.lANActive, .lanError):
            try actions[.fallbackToRelay]?()
            activePath = "relay"
            dispatcherPath = "relay"
            lanSignal = "pending"
            state = .relayFallback
            return []
        case (.relayFallback, .relayOk):
            state = .relayConnected
            return []
        case (.lANActive, .recvLanOffer) where guards[.lanEnabled]?() == true:
            try actions[.dialLan]?()
            state = .lANConnecting
            return []
        default:
            return []
        }
    }

    /// Process a received message. Returns the new state, or nil if rejected.
    @discardableResult
    public func handleMessage(_ msg: MessageType) throws -> PathSwitchClientState? {
        switch (state, msg) {
        case (.relayConnected, .lanOffer) where guards[.lanEnabled]?() == true:
            try actions[.dialLan]?()
            state = .lANConnecting
            return state
        case (.relayConnected, .lanOffer) where guards[.lanDisabled]?() == true:
            state = .relayConnected
            return state
        case (.lANVerifying, .lanConfirm):
            try actions[.activateLan]?()
            activePath = "lan"
            dispatcherPath = "lan"
            lanSignal = "ready"
            state = .lANActive
            return state
        case (.lANActive, .pathPing):
            state = .lANActive
            return state
        case (.lANActive, .lanOffer) where guards[.lanEnabled]?() == true:
            try actions[.dialLan]?()
            state = .lANConnecting
            return state
        default:
            return nil
        }
    }

    /// Attempt an internal transition. Returns the new state, or nil if none available.
    @discardableResult
    public func step() throws -> PathSwitchClientState? {
        switch state {
        case .lANVerifying:
            dispatcherPath = "relay"
            state = .relayConnected
            return state
        case .lANActive:
            try actions[.fallbackToRelay]?()
            activePath = "relay"
            dispatcherPath = "relay"
            lanSignal = "pending"
            state = .relayFallback
            return state
        case .relayFallback:
            state = .relayConnected
            return state
        default:
            return nil
        }
    }
}

/// PathSwitchRelayMachine is the generated state machine for the relay actor.
public final class PathSwitchRelayMachine: @unchecked Sendable {
    public typealias MessageType = PathSwitchProtocol.MessageType
    public typealias GuardID = PathSwitchProtocol.GuardID
    public typealias ActionID = PathSwitchProtocol.ActionID
    public typealias EventID = PathSwitchProtocol.EventID

    public private(set) var state: PathSwitchRelayState
    public var relayBridge: String // relay bridge state ("active" = bridging, "idle" = backend registered but no client)

    public var guards: [GuardID: () -> Bool] = [:]
    public var actions: [ActionID: () throws -> Void] = [:]

    public init() {
        self.state = .idle
        self.relayBridge = "idle"
    }

    /// Handle any event (message receipt or internal). Returns emitted commands.
    @discardableResult
    public func handleEvent(_ ev: EventID) throws -> [String] {
        switch (state, ev) {
        case (.idle, .backendRegister):
            state = .backendRegistered
            return []
        case (.backendRegistered, .clientConnect):
            try actions[.bridgeStreams]?()
            relayBridge = "active"
            state = .bridged
            return []
        case (.bridged, .clientDisconnect):
            try actions[.unbridge]?()
            relayBridge = "idle"
            state = .backendRegistered
            return []
        case (.bridged, .recvRelayResume):
            try actions[.rebridgeStreams]?()
            state = .bridged
            return []
        case (.backendRegistered, .backendDisconnect):
            state = .idle
            return []
        default:
            return []
        }
    }

    /// Process a received message. Returns the new state, or nil if rejected.
    @discardableResult
    public func handleMessage(_ msg: MessageType) throws -> PathSwitchRelayState? {
        switch (state, msg) {
        case (.bridged, .relayResume):
            try actions[.rebridgeStreams]?()
            state = .bridged
            return state
        default:
            return nil
        }
    }

    /// Attempt an internal transition. Returns the new state, or nil if none available.
    @discardableResult
    public func step() throws -> PathSwitchRelayState? {
        switch state {
        case .idle:
            state = .backendRegistered
            return state
        case .bridged:
            try actions[.unbridge]?()
            relayBridge = "idle"
            state = .backendRegistered
            return state
        default:
            return nil
        }
    }
}


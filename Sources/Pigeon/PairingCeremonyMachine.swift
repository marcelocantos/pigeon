// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Auto-generated from protocol definition. Do not edit.
// Source of truth: protocol/*.yaml

import Foundation

public enum PairingCeremonyAcceptorState: String, Sendable {
    case idle = "Idle"
    case generatingEphemeral = "GeneratingEphemeral"
    case registeringRelay = "RegisteringRelay"
    case waitingForHello = "WaitingForHello"
    case waitingForReveal = "WaitingForReveal"
    case derivingCode = "DerivingCode"
    case awaitingUserConfirm = "AwaitingUserConfirm"
    case awaitingPeerConfirm = "AwaitingPeerConfirm"
    case paired = "Paired"
    case aborted = "Aborted"
}

public enum PairingCeremonyInitiatorState: String, Sendable {
    case idle = "Idle"
    case decodingToken = "DecodingToken"
    case generatingEphemeral = "GeneratingEphemeral"
    case connectingRelay = "ConnectingRelay"
    case awaitingWelcome = "AwaitingWelcome"
    case revealing = "Revealing"
    case derivingCode = "DerivingCode"
    case awaitingUserConfirm = "AwaitingUserConfirm"
    case awaitingPeerConfirm = "AwaitingPeerConfirm"
    case paired = "Paired"
    case aborted = "Aborted"
}

/// The protocol transition table and shared type enums.
public enum PairingCeremonyProtocol {

    public enum MessageType: String, Sendable {
        case hello = "hello"
        case welcome = "welcome"
        case reveal = "reveal"
        case confirmToInitiator = "confirm_to_initiator"
        case confirmToAcceptor = "confirm_to_acceptor"
    }

    public enum ActionID: String, Sendable {
        case genEphemeral = "gen_ephemeral"
        case registerRelay = "register_relay"
        case emitToken = "emit_token"
        case storeCommit = "store_commit"
        case verifyCommitAndDerive = "verify_commit_and_derive"
        case storeRecord = "store_record"
        case decodeToken = "decode_token"
        case dialRelay = "dial_relay"
        case sendReveal = "send_reveal"
        case deriveCode = "derive_code"
    }

    public enum EventID: String, Sendable {
        case pairBegin = "pair_begin"
        case ephemeralReady = "ephemeral_ready"
        case relayRegistered = "relay_registered"
        case codeReady = "code_ready"
        case userConfirm = "user_confirm"
        case userCancel = "user_cancel"
        case commitFail = "commit_fail"
        case tokenReceived = "token_received"
        case tokenDecoded = "token_decoded"
        case relayConnected = "relay_connected"
        case revealSent = "reveal_sent"
        case recvHello = "recv_hello"
        case recvReveal = "recv_reveal"
        case recvConfirmToAcceptor = "recv_confirm_to_acceptor"
        case recvWelcome = "recv_welcome"
        case recvConfirmToInitiator = "recv_confirm_to_initiator"
    }


    /// acceptor transitions.
    public static let acceptorInitial: PairingCeremonyAcceptorState = .idle

    public static let acceptorTransitions: [(from: String, to: String, on: String, onKind: String, guard: String?, action: String?, sends: [(to: String, msg: String)])] = [
        (from: "Idle", to: "GeneratingEphemeral", on: "pair_begin", onKind: "internal", guard: nil, action: "gen_ephemeral", sends: []),
        (from: "GeneratingEphemeral", to: "RegisteringRelay", on: "ephemeral_ready", onKind: "internal", guard: nil, action: "register_relay", sends: []),
        (from: "RegisteringRelay", to: "WaitingForHello", on: "relay_registered", onKind: "internal", guard: nil, action: "emit_token", sends: []),
        (from: "WaitingForHello", to: "WaitingForReveal", on: "hello", onKind: "recv", guard: nil, action: "store_commit", sends: [(to: "initiator", msg: "welcome")]),
        (from: "WaitingForReveal", to: "DerivingCode", on: "reveal", onKind: "recv", guard: nil, action: "verify_commit_and_derive", sends: []),
        (from: "DerivingCode", to: "AwaitingUserConfirm", on: "code_ready", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "AwaitingUserConfirm", to: "AwaitingPeerConfirm", on: "user_confirm", onKind: "internal", guard: nil, action: nil, sends: [(to: "initiator", msg: "confirm_to_initiator")]),
        (from: "AwaitingPeerConfirm", to: "Paired", on: "confirm_to_acceptor", onKind: "recv", guard: nil, action: "store_record", sends: []),
        (from: "AwaitingUserConfirm", to: "Aborted", on: "user_cancel", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "AwaitingPeerConfirm", to: "Aborted", on: "user_cancel", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "WaitingForReveal", to: "Aborted", on: "commit_fail", onKind: "internal", guard: nil, action: nil, sends: []),
    ]

    /// initiator transitions.
    public static let initiatorInitial: PairingCeremonyInitiatorState = .idle

    public static let initiatorTransitions: [(from: String, to: String, on: String, onKind: String, guard: String?, action: String?, sends: [(to: String, msg: String)])] = [
        (from: "Idle", to: "DecodingToken", on: "token_received", onKind: "internal", guard: nil, action: "decode_token", sends: []),
        (from: "DecodingToken", to: "GeneratingEphemeral", on: "token_decoded", onKind: "internal", guard: nil, action: "gen_ephemeral", sends: []),
        (from: "GeneratingEphemeral", to: "ConnectingRelay", on: "ephemeral_ready", onKind: "internal", guard: nil, action: "dial_relay", sends: []),
        (from: "ConnectingRelay", to: "AwaitingWelcome", on: "relay_connected", onKind: "internal", guard: nil, action: nil, sends: [(to: "acceptor", msg: "hello")]),
        (from: "AwaitingWelcome", to: "Revealing", on: "welcome", onKind: "recv", guard: nil, action: "send_reveal", sends: [(to: "acceptor", msg: "reveal")]),
        (from: "Revealing", to: "DerivingCode", on: "reveal_sent", onKind: "internal", guard: nil, action: "derive_code", sends: []),
        (from: "DerivingCode", to: "AwaitingUserConfirm", on: "code_ready", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "AwaitingUserConfirm", to: "AwaitingPeerConfirm", on: "user_confirm", onKind: "internal", guard: nil, action: nil, sends: [(to: "acceptor", msg: "confirm_to_acceptor")]),
        (from: "AwaitingPeerConfirm", to: "Paired", on: "confirm_to_initiator", onKind: "recv", guard: nil, action: "store_record", sends: []),
        (from: "AwaitingUserConfirm", to: "Aborted", on: "user_cancel", onKind: "internal", guard: nil, action: nil, sends: []),
        (from: "AwaitingPeerConfirm", to: "Aborted", on: "user_cancel", onKind: "internal", guard: nil, action: nil, sends: []),
    ]
}

/// PairingCeremonyAcceptorMachine is the generated state machine for the acceptor actor.
public final class PairingCeremonyAcceptorMachine: @unchecked Sendable {
    public typealias MessageType = PairingCeremonyProtocol.MessageType
    public typealias ActionID = PairingCeremonyProtocol.ActionID
    public typealias EventID = PairingCeremonyProtocol.EventID

    public private(set) var state: PairingCeremonyAcceptorState
    public var acceptorEphPub: String // acceptor's ephemeral X25519 public key
    public var acceptorReceivedCommit: String // SAS commit from hello (binds peer eph before reveal)
    public var acceptorReceivedEphPub: String // ephemeral pubkey acceptor saw in reveal (may be adversary's)
    public var acceptorReceivedIdentity: String // identity pubkey acceptor saw in hello
    public var acceptorReceivedInstance: String // instance ID acceptor saw in hello
    public var acceptorCommitOk: String // did the reveal open the hello commit? "true" only after CommitMatches
    public var acceptorCode: String // confirmation code acceptor derived from its (ephA, ephB) view
    public var acceptorUserConfirmed: String // has the acceptor's local human pressed y?
    public var acceptorReceivedConfirm: String // has the acceptor received initiator's confirm message?

    public var actions: [ActionID: () throws -> Void] = [:]

    public init() {
        self.state = .idle
        self.acceptorEphPub = "none"
        self.acceptorReceivedCommit = "none"
        self.acceptorReceivedEphPub = "none"
        self.acceptorReceivedIdentity = "none"
        self.acceptorReceivedInstance = "none"
        self.acceptorCommitOk = "false"
        self.acceptorCode = ""
        self.acceptorUserConfirmed = "false"
        self.acceptorReceivedConfirm = "false"
    }

    /// Handle any event (message receipt or internal). Returns emitted commands.
    @discardableResult
    public func handleEvent(_ ev: EventID) throws -> [String] {
        switch (state, ev) {
        case (.idle, .pairBegin):
            try actions[.genEphemeral]?()
            acceptorEphPub = "acceptor_eph"
            state = .generatingEphemeral
            return []
        case (.generatingEphemeral, .ephemeralReady):
            try actions[.registerRelay]?()
            state = .registeringRelay
            return []
        case (.registeringRelay, .relayRegistered):
            try actions[.emitToken]?()
            state = .waitingForHello
            return []
        case (.waitingForHello, .recvHello):
            try actions[.storeCommit]?()
            // acceptor_received_commit: recv_msg.commit (set by action)
            // acceptor_received_identity: recv_msg.identity_pub (set by action)
            // acceptor_received_instance: recv_msg.instance_id (set by action)
            state = .waitingForReveal
            return []
        case (.waitingForReveal, .recvReveal):
            try actions[.verifyCommitAndDerive]?()
            // acceptor_received_eph_pub: recv_msg.eph_pub (set by action)
            // acceptor_commit_ok: IF CommitMatches(acceptor_received_commit, recv_msg.eph_pub) THEN "true" ELSE "false" (set by action)
            // acceptor_code: IF CommitMatches(acceptor_received_commit, recv_msg.eph_pub) THEN DeriveCode(acceptor_eph_pub, recv_msg.eph_pub) ELSE <<"none">> (set by action)
            state = .derivingCode
            return []
        case (.derivingCode, .codeReady):
            state = .awaitingUserConfirm
            return []
        case (.awaitingUserConfirm, .userConfirm):
            acceptorUserConfirmed = "true"
            state = .awaitingPeerConfirm
            return []
        case (.awaitingPeerConfirm, .recvConfirmToAcceptor):
            try actions[.storeRecord]?()
            acceptorReceivedConfirm = "true"
            state = .paired
            return []
        case (.awaitingUserConfirm, .userCancel):
            state = .aborted
            return []
        case (.awaitingPeerConfirm, .userCancel):
            state = .aborted
            return []
        case (.waitingForReveal, .commitFail):
            state = .aborted
            return []
        default:
            return []
        }
    }

    /// Process a received message. Returns the new state, or nil if rejected.
    @discardableResult
    public func handleMessage(_ msg: MessageType) throws -> PairingCeremonyAcceptorState? {
        switch (state, msg) {
        case (.waitingForHello, .hello):
            try actions[.storeCommit]?()
            // acceptor_received_commit: recv_msg.commit (set by action)
            // acceptor_received_identity: recv_msg.identity_pub (set by action)
            // acceptor_received_instance: recv_msg.instance_id (set by action)
            state = .waitingForReveal
            return state
        case (.waitingForReveal, .reveal):
            try actions[.verifyCommitAndDerive]?()
            // acceptor_received_eph_pub: recv_msg.eph_pub (set by action)
            // acceptor_commit_ok: IF CommitMatches(acceptor_received_commit, recv_msg.eph_pub) THEN "true" ELSE "false" (set by action)
            // acceptor_code: IF CommitMatches(acceptor_received_commit, recv_msg.eph_pub) THEN DeriveCode(acceptor_eph_pub, recv_msg.eph_pub) ELSE <<"none">> (set by action)
            state = .derivingCode
            return state
        case (.awaitingPeerConfirm, .confirmToAcceptor):
            try actions[.storeRecord]?()
            acceptorReceivedConfirm = "true"
            state = .paired
            return state
        default:
            return nil
        }
    }

    /// Attempt an internal transition. Returns the new state, or nil if none available.
    @discardableResult
    public func step() throws -> PairingCeremonyAcceptorState? {
        switch state {
        case .idle:
            try actions[.genEphemeral]?()
            acceptorEphPub = "acceptor_eph"
            state = .generatingEphemeral
            return state
        case .generatingEphemeral:
            try actions[.registerRelay]?()
            state = .registeringRelay
            return state
        case .registeringRelay:
            try actions[.emitToken]?()
            state = .waitingForHello
            return state
        case .derivingCode:
            state = .awaitingUserConfirm
            return state
        case .awaitingPeerConfirm:
            state = .aborted
            return state
        case .waitingForReveal:
            state = .aborted
            return state
        default:
            return nil
        }
    }
}

/// PairingCeremonyInitiatorMachine is the generated state machine for the initiator actor.
public final class PairingCeremonyInitiatorMachine: @unchecked Sendable {
    public typealias MessageType = PairingCeremonyProtocol.MessageType
    public typealias ActionID = PairingCeremonyProtocol.ActionID
    public typealias EventID = PairingCeremonyProtocol.EventID

    public private(set) var state: PairingCeremonyInitiatorState
    public var initiatorEphPub: String // initiator's ephemeral X25519 public key
    public var initiatorCommit: String // SHA256("pigeon-sas-commit"||eph||blind) for initiator_eph_pub
    public var receivedAcceptorEphPub: String // acceptor ephemeral pubkey from token (trusted, out-of-band)
    public var receivedAcceptorIdentity: String // acceptor identity pubkey from token
    public var receivedAcceptorInstance: String // acceptor instance ID from token
    public var initiatorReceivedEphPub: String // ephemeral pubkey initiator saw in welcome (may be adversary's)
    public var initiatorReceivedIdentity: String // identity pubkey initiator saw in welcome
    public var initiatorReceivedInstance: String // instance ID initiator saw in welcome
    public var initiatorCode: String // confirmation code initiator derived from its (ephA, ephB) view
    public var initiatorUserConfirmed: String // has the initiator's local human pressed y?
    public var initiatorReceivedConfirm: String // has the initiator received acceptor's confirm message?

    public var actions: [ActionID: () throws -> Void] = [:]

    public init() {
        self.state = .idle
        self.initiatorEphPub = "none"
        self.initiatorCommit = "none"
        self.receivedAcceptorEphPub = "none"
        self.receivedAcceptorIdentity = "none"
        self.receivedAcceptorInstance = "none"
        self.initiatorReceivedEphPub = "none"
        self.initiatorReceivedIdentity = "none"
        self.initiatorReceivedInstance = "none"
        self.initiatorCode = ""
        self.initiatorUserConfirmed = "false"
        self.initiatorReceivedConfirm = "false"
    }

    /// Handle any event (message receipt or internal). Returns emitted commands.
    @discardableResult
    public func handleEvent(_ ev: EventID) throws -> [String] {
        switch (state, ev) {
        case (.idle, .tokenReceived):
            try actions[.decodeToken]?()
            receivedAcceptorEphPub = "acceptor_eph"
            receivedAcceptorIdentity = "acceptor_id"
            receivedAcceptorInstance = "acceptor_instance"
            state = .decodingToken
            return []
        case (.decodingToken, .tokenDecoded):
            try actions[.genEphemeral]?()
            initiatorEphPub = "initiator_eph"
            initiatorCommit = "initiator_commit"
            state = .generatingEphemeral
            return []
        case (.generatingEphemeral, .ephemeralReady):
            try actions[.dialRelay]?()
            state = .connectingRelay
            return []
        case (.connectingRelay, .relayConnected):
            state = .awaitingWelcome
            return []
        case (.awaitingWelcome, .recvWelcome):
            try actions[.sendReveal]?()
            // initiator_received_eph_pub: recv_msg.eph_pub (set by action)
            // initiator_received_identity: recv_msg.identity_pub (set by action)
            // initiator_received_instance: recv_msg.instance_id (set by action)
            state = .revealing
            return []
        case (.revealing, .revealSent):
            try actions[.deriveCode]?()
            // initiator_code: DeriveCode(initiator_eph_pub, initiator_received_eph_pub) (set by action)
            state = .derivingCode
            return []
        case (.derivingCode, .codeReady):
            state = .awaitingUserConfirm
            return []
        case (.awaitingUserConfirm, .userConfirm):
            initiatorUserConfirmed = "true"
            state = .awaitingPeerConfirm
            return []
        case (.awaitingPeerConfirm, .recvConfirmToInitiator):
            try actions[.storeRecord]?()
            initiatorReceivedConfirm = "true"
            state = .paired
            return []
        case (.awaitingUserConfirm, .userCancel):
            state = .aborted
            return []
        case (.awaitingPeerConfirm, .userCancel):
            state = .aborted
            return []
        default:
            return []
        }
    }

    /// Process a received message. Returns the new state, or nil if rejected.
    @discardableResult
    public func handleMessage(_ msg: MessageType) throws -> PairingCeremonyInitiatorState? {
        switch (state, msg) {
        case (.awaitingWelcome, .welcome):
            try actions[.sendReveal]?()
            // initiator_received_eph_pub: recv_msg.eph_pub (set by action)
            // initiator_received_identity: recv_msg.identity_pub (set by action)
            // initiator_received_instance: recv_msg.instance_id (set by action)
            state = .revealing
            return state
        case (.awaitingPeerConfirm, .confirmToInitiator):
            try actions[.storeRecord]?()
            initiatorReceivedConfirm = "true"
            state = .paired
            return state
        default:
            return nil
        }
    }

    /// Attempt an internal transition. Returns the new state, or nil if none available.
    @discardableResult
    public func step() throws -> PairingCeremonyInitiatorState? {
        switch state {
        case .idle:
            try actions[.decodeToken]?()
            receivedAcceptorEphPub = "acceptor_eph"
            receivedAcceptorIdentity = "acceptor_id"
            receivedAcceptorInstance = "acceptor_instance"
            state = .decodingToken
            return state
        case .decodingToken:
            try actions[.genEphemeral]?()
            initiatorEphPub = "initiator_eph"
            initiatorCommit = "initiator_commit"
            state = .generatingEphemeral
            return state
        case .generatingEphemeral:
            try actions[.dialRelay]?()
            state = .connectingRelay
            return state
        case .connectingRelay:
            state = .awaitingWelcome
            return state
        case .revealing:
            try actions[.deriveCode]?()
            // initiator_code: DeriveCode(initiator_eph_pub, initiator_received_eph_pub) (set by action)
            state = .derivingCode
            return state
        case .derivingCode:
            state = .awaitingUserConfirm
            return state
        case .awaitingPeerConfirm:
            state = .aborted
            return state
        default:
            return nil
        }
    }
}


// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Swift-side pairing-ceremony driver.
//
// Mirrors pairing/pairing.go's runAcceptor / runInitiator: drives the
// generated PairingCeremonyAcceptorMachine / PairingCeremonyInitiatorMachine
// through the hello / welcome / confirm wire exchange, then returns a
// PairingRecord whose `deriveChannel` is wire-compatible with the Go and
// Kotlin SDKs.
//
// The wire format matches c/src/pairing.c exactly (🎯T52 commit round):
//   hello   = {"kind":"hello",  "commit":"<b64>","identity_pub":"<b64>","instance_id":"<str>"}
//   welcome = {"kind":"welcome","eph_pub":"<b64>","identity_pub":"<b64>","instance_id":"<str>"}
//   reveal  = {"kind":"reveal", "eph_pub":"<b64>","blind":"<b64>"}
//   confirm = {"kind":"confirm"}
// Byte fields are standard base64 with padding (Go's encoding/json default
// for []byte). Each message is framed by the transport's send/recv —
// typically the singleStreamTransport 4-byte big-endian length prefix used
// by pairing/cgo_pair.go and the C reference transport.
//
// Transports are abstracted via `PairingWireTransport`: callers wire any
// byte-pipe (in-memory pair, stdio of a subprocess, an opened
// pigeon.Stream) by conforming to that protocol.

import CryptoKit
import Foundation

// MARK: - Errors

/// Errors specific to the pairing ceremony.
public enum PairingCeremonyError: LocalizedError, Equatable {
    case malformedMessage(String)
    case unexpectedKind(String)
    case userCancelled
    case peerAborted
    case decode(String)

    public var errorDescription: String? {
        switch self {
        case .malformedMessage(let why): "malformed pairing message: \(why)"
        case .unexpectedKind(let kind): "unexpected pairing message kind: \(kind)"
        case .userCancelled: "user cancelled pairing"
        case .peerAborted: "peer aborted pairing"
        case .decode(let why): "decode error: \(why)"
        }
    }
}

// MARK: - Transport

/// A byte-pipe interface for the pairing ceremony. Each `send` writes one
/// framed message to the peer; each `recv` returns the next one. The
/// underlying transport supplies the framing (e.g. 4-byte length prefix
/// over a TCP-like byte stream, or one logical pigeon-stream message per
/// call over libpigeon).
///
/// Mirrors the Go `singleStreamTransport` in pairing/cgo_pair.go: one
/// `Send` writes hdr+payload; one `Recv` reads hdr then exactly hdr-bytes.
public protocol PairingWireTransport: AnyObject {
    func send(_ data: Data) async throws
    func recv() async throws -> Data
    func close() throws
}

// MARK: - Identity

/// Local-side identity for the ceremony. Mirrors a subset of Go's
/// `crypto.Identity` — only what the wire exchange needs.
public struct PairingIdentity: Sendable {
    public let publicKey: Data    // 32-byte X25519
    public let instanceID: String

    public init(publicKey: Data, instanceID: String) {
        precondition(publicKey.count == 32, "identity public key must be 32 bytes")
        self.publicKey = publicKey
        self.instanceID = instanceID
    }
}

// MARK: - Token payload

/// Decoded acceptor-side rendezvous token. Same fields as Go's
/// `pairing.tokenPayload`, base64-decoded for direct use.
///
/// Encoding: JSON with snake_case keys, then base64-RawURL of the JSON
/// bytes (no padding). Matches `pairing.Pairer.Accept` in Go.
public struct PairingToken: Sendable {
    public let relay: String
    public let sessionID: String
    public let acceptorEphPub: Data     // 32 bytes
    public let acceptorIdPub: Data      // 32 bytes
    public let acceptorInstance: String

    public init(relay: String, sessionID: String,
                acceptorEphPub: Data, acceptorIdPub: Data,
                acceptorInstance: String) {
        self.relay = relay
        self.sessionID = sessionID
        self.acceptorEphPub = acceptorEphPub
        self.acceptorIdPub = acceptorIdPub
        self.acceptorInstance = acceptorInstance
    }

    /// Decode a base64-RawURL token produced by Go's `pairing.Pairer.Accept`.
    public static func decode(_ token: String) throws -> PairingToken {
        // base64url-RawNoPad → standard base64 + padding for Data(base64Encoded:).
        var t = token.replacingOccurrences(of: "-", with: "+")
                     .replacingOccurrences(of: "_", with: "/")
        let pad = (4 - t.count % 4) % 4
        t += String(repeating: "=", count: pad)
        guard let raw = Data(base64Encoded: t) else {
            throw PairingCeremonyError.decode("invalid base64 in token")
        }
        guard let obj = try JSONSerialization.jsonObject(with: raw) as? [String: Any] else {
            throw PairingCeremonyError.decode("token is not a JSON object")
        }
        guard let relay = obj["relay"] as? String,
              let sessionID = obj["session_id"] as? String,
              let instance = obj["acceptor_instance"] as? String else {
            throw PairingCeremonyError.decode("token missing required fields")
        }
        // Go encodes []byte as standard base64-with-padding.
        let eph = try decodeStdB64(obj["acceptor_eph_pub"], field: "acceptor_eph_pub")
        let idp = try decodeStdB64(obj["acceptor_id_pub"], field: "acceptor_id_pub")
        return PairingToken(relay: relay, sessionID: sessionID,
                            acceptorEphPub: eph, acceptorIdPub: idp,
                            acceptorInstance: instance)
    }
}

private func decodeStdB64(_ v: Any?, field: String) throws -> Data {
    guard let s = v as? String, let d = Data(base64Encoded: s) else {
        throw PairingCeremonyError.decode("field \(field) is not base64")
    }
    return d
}

// MARK: - Wire message

/// Pairing-ceremony message types (🎯T52 commit-then-reveal):
/// hello carries a SAS commit (not eph); welcome carries acceptor eph;
/// reveal opens the commit; confirm is bare `{"kind":"confirm"}`.
fileprivate enum WireMsg {
    case hello(commit: Data, identityPub: Data, instanceID: String)
    case welcome(ephPub: Data, identityPub: Data, instanceID: String)
    case reveal(ephPub: Data, blind: Data)
    case confirm

    /// Encode this message to the JSON form c/src/pairing.c emits.
    /// The C parser is order-independent; field order matches the C
    /// builders for ergonomic byte-level equivalence in tests.
    func encode() throws -> Data {
        switch self {
        case .hello(let commit, let id, let iid):
            precondition(commit.count == 32 && id.count == 32)
            let cB64 = commit.base64EncodedString()
            let idB64 = id.base64EncodedString()
            let json = #"{"kind":"hello","commit":"\#(cB64)","identity_pub":"\#(idB64)","instance_id":"\#(iid)"}"#
            return Data(json.utf8)
        case .welcome(let eph, let id, let iid):
            precondition(eph.count == 32 && id.count == 32)
            let ephB64 = eph.base64EncodedString()
            let idB64 = id.base64EncodedString()
            let json = #"{"kind":"welcome","eph_pub":"\#(ephB64)","identity_pub":"\#(idB64)","instance_id":"\#(iid)"}"#
            return Data(json.utf8)
        case .reveal(let eph, let blind):
            precondition(eph.count == 32 && blind.count == 32)
            let ephB64 = eph.base64EncodedString()
            let blindB64 = blind.base64EncodedString()
            let json = #"{"kind":"reveal","eph_pub":"\#(ephB64)","blind":"\#(blindB64)"}"#
            return Data(json.utf8)
        case .confirm:
            return Data(#"{"kind":"confirm"}"#.utf8)
        }
    }

    /// Parse the next inbound message. Throws on anything unparseable;
    /// returns the (kind, fields) — caller dispatches by kind via the
    /// generated state machine.
    static func decode(_ data: Data) throws -> ParsedMsg {
        guard let obj = try? JSONSerialization.jsonObject(with: data) as? [String: Any] else {
            throw PairingCeremonyError.malformedMessage("not a JSON object")
        }
        guard let kind = obj["kind"] as? String else {
            throw PairingCeremonyError.malformedMessage("missing 'kind'")
        }
        switch kind {
        case "hello":
            let commit = try decodeStdB64(obj["commit"], field: "commit")
            let id = try decodeStdB64(obj["identity_pub"], field: "identity_pub")
            guard let iid = obj["instance_id"] as? String else {
                throw PairingCeremonyError.malformedMessage("missing instance_id")
            }
            if commit.count != 32 || id.count != 32 {
                throw PairingCeremonyError.malformedMessage("commit/identity must be 32 bytes")
            }
            return ParsedMsg(kind: kind, commit: commit, ephPub: nil, blind: nil,
                             identityPub: id, instanceID: iid)
        case "welcome":
            let eph = try decodeStdB64(obj["eph_pub"], field: "eph_pub")
            let id = try decodeStdB64(obj["identity_pub"], field: "identity_pub")
            guard let iid = obj["instance_id"] as? String else {
                throw PairingCeremonyError.malformedMessage("missing instance_id")
            }
            if eph.count != 32 || id.count != 32 {
                throw PairingCeremonyError.malformedMessage("eph/identity must be 32 bytes")
            }
            return ParsedMsg(kind: kind, commit: nil, ephPub: eph, blind: nil,
                             identityPub: id, instanceID: iid)
        case "reveal":
            let eph = try decodeStdB64(obj["eph_pub"], field: "eph_pub")
            let blind = try decodeStdB64(obj["blind"], field: "blind")
            if eph.count != 32 || blind.count != 32 {
                throw PairingCeremonyError.malformedMessage("eph/blind must be 32 bytes")
            }
            return ParsedMsg(kind: kind, commit: nil, ephPub: eph, blind: blind,
                             identityPub: nil, instanceID: nil)
        case "confirm":
            return ParsedMsg(kind: "confirm", commit: nil, ephPub: nil, blind: nil,
                             identityPub: nil, instanceID: nil)
        default:
            throw PairingCeremonyError.unexpectedKind(kind)
        }
    }
}

fileprivate struct ParsedMsg {
    let kind: String
    let commit: Data?
    let ephPub: Data?
    let blind: Data?
    let identityPub: Data?
    let instanceID: String?
}

// MARK: - Result

/// What a successful ceremony produces. `code` is the 6-digit
/// confirmation code both sides derived; `record` is the persistable
/// `PairingRecord` whose `deriveChannel` is cross-language compatible.
public struct PairingCeremonyResult: Sendable {
    public let code: String
    public let record: PairingRecord
}

// MARK: - Acceptor

/// Drive the acceptor side of the pairing ceremony over `transport`.
///
/// `confirm` is called synchronously with the 6-digit code once derivable;
/// return `true` to confirm (sends `confirm` to peer and awaits theirs),
/// `false` to abort. This mirrors the Go `cer.Code(ctx)` → `cer.Confirm(ctx)`
/// rendezvous (per MEMORY.md feedback_no_progress_callbacks: the multi-
/// step ceremony returns a result, not a stream of callbacks).
///
/// `acceptorEphPriv` and `acceptorEphPub` are the 32-byte X25519 key pair
/// the caller generated for this ceremony (the public half should already
/// have been embedded in the rendezvous token).
public func pairAcceptor(
    transport: PairingWireTransport,
    identity: PairingIdentity,
    acceptorEphPriv: Data,
    acceptorEphPub: Data,
    relayURL: String = "",
    confirm: @Sendable (String) async -> Bool
) async throws -> PairingCeremonyResult {
    precondition(acceptorEphPriv.count == 32 && acceptorEphPub.count == 32)

    let m = PairingCeremonyAcceptorMachine()
    // Setup-phase events have no wire I/O — they correspond to work the
    // caller has already done (keygen, relay registration, token emit).
    // Wire them as no-ops so the FSM advances through Idle → WaitingForHello.
    m.actions[.genEphemeral]           = {}
    m.actions[.registerRelay]          = {}
    m.actions[.emitToken]              = {}
    m.actions[.storeCommit]            = {}
    m.actions[.verifyCommitAndDerive]  = {}
    m.actions[.storeRecord]            = {}

    try m.handleEvent(.pairBegin)
    try m.handleEvent(.ephemeralReady)
    try m.handleEvent(.relayRegistered)

    // --- Read hello (commit only; eph revealed later) ---
    let helloBytes = try await transport.recv()
    let hello = try WireMsg.decode(helloBytes)
    guard hello.kind == "hello",
          let peerCommit = hello.commit,
          let peerInstance = hello.instanceID else {
        throw PairingCeremonyError.unexpectedKind(hello.kind)
    }
    try m.handleMessage(.hello)

    // --- Send welcome (acceptor eph already bound by OOB token) ---
    let welcome = WireMsg.welcome(
        ephPub: acceptorEphPub,
        identityPub: identity.publicKey,
        instanceID: identity.instanceID
    )
    try await transport.send(try welcome.encode())

    // --- Read reveal; verify it opens the hello commit ---
    let revealBytes = try await transport.recv()
    let reveal = try WireMsg.decode(revealBytes)
    guard reveal.kind == "reveal",
          let peerEph = reveal.ephPub,
          let peerBlind = reveal.blind else {
        throw PairingCeremonyError.unexpectedKind(reveal.kind)
    }
    guard sasCommitVerify(ephPub: peerEph, blind: peerBlind, commit: peerCommit) else {
        try m.handleEvent(.commitFail)
        throw PairingCeremonyError.malformedMessage("reveal does not open hello commit")
    }
    try m.handleMessage(.reveal)
    try m.handleEvent(.codeReady)

    // --- Derive code only after commit verified ---
    let code = deriveConfirmationCode(acceptorEphPub, peerEph)

    // --- Ask local user (callback rendezvous) ---
    if !(await confirm(code)) {
        try m.handleEvent(.userCancel)
        throw PairingCeremonyError.userCancelled
    }
    try m.handleEvent(.userConfirm)

    // --- Send confirm, then await peer's confirm ---
    try await transport.send(try WireMsg.confirm.encode())

    let peerConfirmBytes = try await transport.recv()
    let peerConfirm = try WireMsg.decode(peerConfirmBytes)
    guard peerConfirm.kind == "confirm" else {
        throw PairingCeremonyError.peerAborted
    }
    try m.handleMessage(.confirmToAcceptor)

    let record = PairingRecord(
        peerInstanceID: peerInstance,
        relayURL: relayURL,
        localPrivateKey: acceptorEphPriv,
        localPublicKey: acceptorEphPub,
        peerPublicKey: peerEph
    )
    return PairingCeremonyResult(code: code, record: record)
}

// MARK: - Initiator

/// Drive the initiator side of the pairing ceremony over `transport`.
///
/// The acceptor's ephemeral pubkey and instance ID come from the
/// rendezvous token; they're passed in directly so this function stays
/// transport-agnostic. Callers that have a base64 token can decode it
/// first with `PairingToken.decode(_:)`.
public func pairInitiator(
    transport: PairingWireTransport,
    identity: PairingIdentity,
    initiatorEphPriv: Data,
    initiatorEphPub: Data,
    acceptorEphPub: Data,
    acceptorInstance: String,
    relayURL: String = "",
    confirm: @Sendable (String) async -> Bool
) async throws -> PairingCeremonyResult {
    precondition(initiatorEphPriv.count == 32 && initiatorEphPub.count == 32 && acceptorEphPub.count == 32)

    let m = PairingCeremonyInitiatorMachine()
    m.actions[.decodeToken]  = {}
    m.actions[.genEphemeral] = {}
    m.actions[.dialRelay]    = {}
    m.actions[.sendReveal]   = {}
    m.actions[.deriveCode]   = {}
    m.actions[.storeRecord]  = {}

    try m.handleEvent(.tokenReceived)
    try m.handleEvent(.tokenDecoded)
    try m.handleEvent(.ephemeralReady)

    // --- Mint SAS blind + commit before revealing eph (🎯T52) ---
    let blind = generateNonce()
    let commit = sasCommit(ephPub: initiatorEphPub, blind: blind)

    try m.handleEvent(.relayConnected)

    // --- Send hello (commitment only) ---
    let hello = WireMsg.hello(
        commit: commit,
        identityPub: identity.publicKey,
        instanceID: identity.instanceID
    )
    try await transport.send(try hello.encode())

    // --- Read welcome ---
    let welcomeBytes = try await transport.recv()
    let welcome = try WireMsg.decode(welcomeBytes)
    guard welcome.kind == "welcome",
          let peerEph = welcome.ephPub,
          let peerInstance = welcome.instanceID else {
        throw PairingCeremonyError.unexpectedKind(welcome.kind)
    }
    // The acceptor's ephemeral pubkey in the welcome must match the one
    // the rendezvous token committed to; otherwise we're MitM'd before
    // even deriving the code. The token's value is the authoritative one.
    if peerEph != acceptorEphPub {
        throw PairingCeremonyError.malformedMessage("welcome eph_pub mismatch vs token")
    }
    try m.handleMessage(.welcome)

    // --- Reveal eph under the prior commit ---
    try await transport.send(try WireMsg.reveal(ephPub: initiatorEphPub, blind: blind).encode())
    try m.handleEvent(.revealSent)
    try m.handleEvent(.codeReady)

    // --- Derive code (order-independent, token-bound acceptor eph) ---
    let code = deriveConfirmationCode(initiatorEphPub, acceptorEphPub)

    if !(await confirm(code)) {
        try m.handleEvent(.userCancel)
        throw PairingCeremonyError.userCancelled
    }
    try m.handleEvent(.userConfirm)

    try await transport.send(try WireMsg.confirm.encode())

    let peerConfirmBytes = try await transport.recv()
    let peerConfirm = try WireMsg.decode(peerConfirmBytes)
    guard peerConfirm.kind == "confirm" else {
        throw PairingCeremonyError.peerAborted
    }
    try m.handleMessage(.confirmToInitiator)

    let record = PairingRecord(
        peerInstanceID: peerInstance.isEmpty ? acceptorInstance : peerInstance,
        relayURL: relayURL,
        localPrivateKey: initiatorEphPriv,
        localPublicKey: initiatorEphPub,
        peerPublicKey: acceptorEphPub
    )
    return PairingCeremonyResult(code: code, record: record)
}


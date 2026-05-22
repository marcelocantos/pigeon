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
// The wire format matches c/src/pairing.c exactly:
//   hello   = {"kind":"hello",  "eph_pub":"<b64>","identity_pub":"<b64>","instance_id":"<str>"}
//   welcome = {"kind":"welcome","eph_pub":"<b64>","identity_pub":"<b64>","instance_id":"<str>"}
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

/// The three pairing-ceremony message types. Hello/welcome carry the
/// sender's ephemeral pubkey + identity pubkey + instance ID; confirm
/// is a bare {"kind":"confirm"}.
fileprivate enum WireMsg {
    case hello(ephPub: Data, identityPub: Data, instanceID: String)
    case welcome(ephPub: Data, identityPub: Data, instanceID: String)
    case confirm

    /// Encode this message to the JSON form c/src/pairing.c emits.
    /// Fields go in (kind, eph_pub, identity_pub, instance_id) order —
    /// the C parser is order-independent so this is purely for ergonomic
    /// equivalence with c/src/pairing.c::build_hello / build_welcome.
    func encode() throws -> Data {
        switch self {
        case .hello(let eph, let id, let iid):
            return try encodeKeyed(kind: "hello", eph: eph, id: id, iid: iid)
        case .welcome(let eph, let id, let iid):
            return try encodeKeyed(kind: "welcome", eph: eph, id: id, iid: iid)
        case .confirm:
            return Data(#"{"kind":"confirm"}"#.utf8)
        }
    }

    private func encodeKeyed(kind: String, eph: Data, id: Data, iid: String) throws -> Data {
        precondition(eph.count == 32, "eph_pub must be 32 bytes")
        precondition(id.count == 32, "identity_pub must be 32 bytes")
        // Go's encoding/json escapes only the JSON metacharacters, and
        // instance IDs are base64url alphabet so no escaping fires. We
        // build the string manually so the bytes match c/src/pairing.c
        // exactly (incl. no extra whitespace).
        let ephB64 = eph.base64EncodedString()
        let idB64 = id.base64EncodedString()
        let json = #"{"kind":"\#(kind)","eph_pub":"\#(ephB64)","identity_pub":"\#(idB64)","instance_id":"\#(iid)"}"#
        return Data(json.utf8)
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
        case "hello", "welcome":
            let eph = try decodeStdB64(obj["eph_pub"], field: "eph_pub")
            let id = try decodeStdB64(obj["identity_pub"], field: "identity_pub")
            guard let iid = obj["instance_id"] as? String else {
                throw PairingCeremonyError.malformedMessage("missing instance_id")
            }
            if eph.count != 32 || id.count != 32 {
                throw PairingCeremonyError.malformedMessage("eph/identity must be 32 bytes")
            }
            return ParsedMsg(kind: kind, ephPub: eph, identityPub: id, instanceID: iid)
        case "confirm":
            return ParsedMsg(kind: "confirm", ephPub: nil, identityPub: nil, instanceID: nil)
        default:
            throw PairingCeremonyError.unexpectedKind(kind)
        }
    }
}

fileprivate struct ParsedMsg {
    let kind: String
    let ephPub: Data?
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
    m.actions[.genEphemeral]  = {}
    m.actions[.registerRelay] = {}
    m.actions[.emitToken]     = {}
    m.actions[.deriveCode]    = {}
    m.actions[.storeRecord]   = {}

    try m.handleEvent(.pairBegin)
    try m.handleEvent(.ephemeralReady)
    try m.handleEvent(.relayRegistered)

    // --- Read hello ---
    let helloBytes = try await transport.recv()
    let hello = try WireMsg.decode(helloBytes)
    guard hello.kind == "hello",
          let peerEph = hello.ephPub,
          let peerInstance = hello.instanceID else {
        throw PairingCeremonyError.unexpectedKind(hello.kind)
    }
    try m.handleMessage(.hello)
    try m.handleEvent(.codeReady)

    // --- Send welcome ---
    let welcome = WireMsg.welcome(
        ephPub: acceptorEphPub,
        identityPub: identity.publicKey,
        instanceID: identity.instanceID
    )
    try await transport.send(try welcome.encode())

    // --- Derive code ---
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
    m.actions[.deriveCode]   = {}
    m.actions[.storeRecord]  = {}

    try m.handleEvent(.tokenReceived)
    try m.handleEvent(.tokenDecoded)
    try m.handleEvent(.ephemeralReady)
    try m.handleEvent(.relayConnected)

    // --- Send hello ---
    let hello = WireMsg.hello(
        ephPub: initiatorEphPub,
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
    try m.handleEvent(.codeReady)

    // --- Derive code (order-independent, matches Go acceptor) ---
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


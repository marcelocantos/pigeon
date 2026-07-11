// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

import Foundation

/// The default lifetime of a `PairingArtifact` when none is specified.
/// After this period elapses the client must re-pair.
public let defaultPairingTTL: TimeInterval = 30 * 24 * 60 * 60

/// Errors specific to pairing artifacts.
public enum PairingError: LocalizedError, Equatable {
    /// The artifact is past its `expiresAt` timestamp.
    case expired(at: Date)
    /// The artifact is missing a required field.
    case missingField(String)
    /// The text encoding could not be decoded.
    case malformedText
    /// The session-salt handshake (🎯T50) failed.
    case sessionSaltHandshake(String)

    public var errorDescription: String? {
        switch self {
        case .expired(let at): "Pairing artifact expired at \(at)"
        case .missingField(let f): "Pairing artifact missing field: \(f)"
        case .malformedText: "Malformed pairing artifact text encoding"
        case .sessionSaltHandshake(let msg): "Session salt handshake failed: \(msg)"
        }
    }
}

/// The persistable+expirable envelope around a completed pairing.
/// The cryptographic core lives in `record`; the wrapping fields carry
/// the lifecycle metadata that lets a client detect expiry and prompt
/// re-pair.
///
/// The artifact is what gets QR-encoded for transport from the
/// pairing host to the client device, and what a `CredentialStore`
/// persists across restarts.
///
/// Wire format matches the Go and Kotlin SDKs: snake_case JSON keys,
/// ISO-8601 timestamps, base64 byte fields. An artifact minted in Go
/// and emitted by `pigeon-pair` decodes correctly in Swift via
/// `PairingArtifact.fromJSON(_:)` / `.fromText(_:)`.
public struct PairingArtifact: Codable, Sendable {
    public var record: PairingRecord
    public var token: String
    public var issuedAt: Date
    public var expiresAt: Date?

    public init(record: PairingRecord, token: String = "", issuedAt: Date = Date(), ttl: TimeInterval = defaultPairingTTL) {
        self.record = record
        self.token = token
        self.issuedAt = issuedAt
        if ttl > 0 {
            self.expiresAt = issuedAt.addingTimeInterval(ttl)
        } else {
            self.expiresAt = nil
        }
    }

    enum CodingKeys: String, CodingKey {
        case record
        case token
        case issuedAt = "issued_at"
        case expiresAt = "expires_at"
    }

    public init(from decoder: Decoder) throws {
        let c = try decoder.container(keyedBy: CodingKeys.self)
        self.record = try c.decode(PairingRecord.self, forKey: .record)
        self.token = (try? c.decode(String.self, forKey: .token)) ?? ""
        self.issuedAt = try c.decode(Date.self, forKey: .issuedAt)
        let zero = Date(timeIntervalSince1970: 0)
        if let exp = try? c.decode(Date.self, forKey: .expiresAt), exp != zero {
            self.expiresAt = exp
        } else {
            self.expiresAt = nil
        }
    }

    public func encode(to encoder: Encoder) throws {
        var c = encoder.container(keyedBy: CodingKeys.self)
        try c.encode(record, forKey: .record)
        try c.encode(token, forKey: .token)
        try c.encode(issuedAt, forKey: .issuedAt)
        // Encode zero (epoch) when no expiry, mirroring Go's time.Time zero value.
        try c.encode(expiresAt ?? Date(timeIntervalSince1970: 0), forKey: .expiresAt)
    }

    /// Reports whether the artifact's `expiresAt` is in the past at the
    /// given instant. An artifact with no expiry never expires.
    public func isExpired(now: Date = Date()) -> Bool {
        guard let exp = expiresAt else { return false }
        return now >= exp
    }

    /// Serialises the artifact to canonical JSON. Wire-compatible with
    /// the Go and Kotlin SDKs.
    public func toJSON() throws -> Data {
        let enc = JSONEncoder.pigeonArtifact
        return try enc.encode(self)
    }

    /// Deserialises an artifact from canonical JSON.
    public static func fromJSON(_ data: Data) throws -> PairingArtifact {
        let dec = JSONDecoder.pigeonArtifact
        return try dec.decode(PairingArtifact.self, from: data)
    }

    /// Returns the canonical single-line text encoding: base64url of
    /// the JSON bytes, with no padding. Suitable for transport via QR
    /// payload, launch argument, environment variable, pasteboard, or
    /// any other channel that wants a single token of text.
    public func toText() throws -> String {
        let json = try toJSON()
        return base64URLNoPad(json)
    }

    /// Decodes the canonical text encoding produced by `toText()`.
    public static func fromText(_ text: String) throws -> PairingArtifact {
        guard let data = base64URLDecode(text) else {
            throw PairingError.malformedText
        }
        return try fromJSON(data)
    }
}

#if canImport(Network)
import Network

extension PigeonConn {
    /// Connect to the relay using a persisted [PairingArtifact] and
    /// return the conn alongside the derived [E2EChannel] ready for
    /// encrypted send/recv.
    ///
    /// The relay URL and peer instance ID come from the artifact; the
    /// artifact's expiry is checked up front and throws
    /// `PairingError.expired(at:)` (matchable for re-pair routing).
    ///
    /// After the relay bridge is live, runs the 🎯T50 session-salt
    /// handshake on the primary stream: mints a fresh
    /// `sessionSaltLen`-byte salt, sends it plaintext, waits for
    /// `sessionSaltAck`, then derives the AEAD channel with
    /// HKDF info = direction-label || salt so each reconnect gets
    /// distinct keys (counter-reset-to-0 under a reused key is no
    /// longer possible).
    ///
    /// The peer must call `acceptSessionSalt(record:)` on its bridged
    /// conn with the mirrored PairingRecord.
    public static func connect(
        artifact: PairingArtifact,
        quicOptions: NWProtocolQUIC.Options? = nil
    ) async throws -> (PigeonConn, E2EChannel) {
        if artifact.isExpired() {
            throw PairingError.expired(at: artifact.expiresAt ?? Date())
        }
        let (host, port) = try parseRelayURL(artifact.record.relayURL)
        let conn = try await PigeonConn.connect(
            host: host, port: port,
            instanceID: artifact.record.peerInstanceID,
            quicOptions: quicOptions)
        let channel = try await conn.initiateSessionSalt(
            record: artifact.record,
            sendInfo: Data("client-to-server".utf8),
            recvInfo: Data("server-to-client".utf8))
        return (conn, channel)
    }

    /// Client side of the 🎯T50 session-salt handshake: mint salt,
    /// send it, wait for `sessionSaltAck`, derive channel with
    /// info = label || salt.
    public func initiateSessionSalt(
        record: PairingRecord,
        sendInfo: Data,
        recvInfo: Data
    ) async throws -> E2EChannel {
        let salt = generateSessionSalt()
        try await send(salt)
        let ack = try await recv()
        guard String(data: ack, encoding: .utf8) == sessionSaltAck else {
            throw PairingError.sessionSaltHandshake(
                "expected \(sessionSaltAck), got \(String(data: ack, encoding: .utf8) ?? "<binary>")")
        }
        return try record.deriveChannel(
            sendInfo: sendInfo, recvInfo: recvInfo, sessionSalt: salt)
    }

    /// Peer/backend side of the 🎯T50 session-salt handshake: read the
    /// client's salt, reply `sessionSaltAck`, derive channel with the
    /// same salt (direction labels are the caller's responsibility —
    /// typically the reverse of the client's).
    public func acceptSessionSalt(
        record: PairingRecord,
        sendInfo: Data,
        recvInfo: Data
    ) async throws -> E2EChannel {
        let salt = try await recv()
        guard salt.count == sessionSaltLen else {
            throw PairingError.sessionSaltHandshake(
                "expected \(sessionSaltLen)-byte salt, got \(salt.count)")
        }
        try await send(Data(sessionSaltAck.utf8))
        return try record.deriveChannel(
            sendInfo: sendInfo, recvInfo: recvInfo, sessionSalt: salt)
    }
}

private func parseRelayURL(_ urlString: String) throws -> (String, UInt16) {
    guard let url = URL(string: urlString),
          let host = url.host, !host.isEmpty else {
        throw PairingError.missingField("relay_url")
    }
    let port: UInt16
    if let p = url.port {
        port = UInt16(p)
    } else {
        // Default to the raw-QUIC ALPN port; matches Go pigeon.quicAddr.
        port = 4433
    }
    return (host, port)
}

#endif

private extension JSONEncoder {
    static var pigeonArtifact: JSONEncoder {
        let enc = JSONEncoder()
        enc.dateEncodingStrategy = .iso8601
        return enc
    }
}

private extension JSONDecoder {
    static var pigeonArtifact: JSONDecoder {
        let dec = JSONDecoder()
        dec.dateDecodingStrategy = .iso8601
        return dec
    }
}

private func base64URLNoPad(_ data: Data) -> String {
    let s = data.base64EncodedString()
    return s.replacingOccurrences(of: "+", with: "-")
        .replacingOccurrences(of: "/", with: "_")
        .replacingOccurrences(of: "=", with: "")
}

private func base64URLDecode(_ s: String) -> Data? {
    var t = s.replacingOccurrences(of: "-", with: "+")
        .replacingOccurrences(of: "_", with: "/")
    let pad = (4 - t.count % 4) % 4
    t += String(repeating: "=", count: pad)
    return Data(base64Encoded: t)
}

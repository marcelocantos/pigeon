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

    public var errorDescription: String? {
        switch self {
        case .expired(let at): "Pairing artifact expired at \(at)"
        case .missingField(let f): "Pairing artifact missing field: \(f)"
        case .malformedText: "Malformed pairing artifact text encoding"
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

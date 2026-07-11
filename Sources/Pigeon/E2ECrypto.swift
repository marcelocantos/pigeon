// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

import CryptoKit
import Foundation
import Security

/// End-to-end encryption for relay traffic relayed through pigeon.
/// Mirrors the Go crypto package in crypto/.
///
/// Key exchange: X25519 ECDH
/// Symmetric encryption: AES-256-GCM with counter nonce
/// Key derivation: HKDF-SHA256

// MARK: - Key exchange

public struct E2EKeyPair {
    public let privateKey: Curve25519.KeyAgreement.PrivateKey
    public var publicKey: Curve25519.KeyAgreement.PublicKey { privateKey.publicKey }

    public init() {
        privateKey = .init()
    }

    /// Raw public key bytes (32 bytes) for sending to peer.
    public var publicKeyData: Data {
        Data(publicKey.rawRepresentation)
    }

    /// Derive a shared secret via ECDH, then derive a 256-bit key via HKDF.
    public func deriveSessionKey(peerPublicKey: Data, info: Data) throws -> SymmetricKey {
        let peerKey = try Curve25519.KeyAgreement.PublicKey(rawRepresentation: peerPublicKey)
        let shared = try privateKey.sharedSecretFromKeyAgreement(with: peerKey)
        return shared.hkdfDerivedSymmetricKey(
            using: SHA256.self,
            salt: Data(),
            sharedInfo: info,
            outputByteCount: 32
        )
    }
}

/// Generate 32 cryptographically random bytes suitable for use as a nonce.
public func generateNonce() -> Data {
    var d = Data(count: 32)
    _ = d.withUnsafeMutableBytes { SecRandomCopyBytes(kSecRandomDefault, 32, $0.baseAddress!) }
    return d
}

/// Generate 32 cryptographically random bytes suitable for use as a secret.
public func generateSecret() -> Data { generateNonce() }

/// Derive a session key from a persistent secret and nonce via HKDF.
public func deriveKeyFromSecret(_ secret: Data, info: Data) -> SymmetricKey {
    HKDF<SHA256>.deriveKey(
        inputKeyMaterial: SymmetricKey(data: secret),
        salt: Data(),
        info: info,
        outputByteCount: 32
    )
}

// MARK: - Confirmation code

/// Derive a 6-digit confirmation code from two X25519 public keys.
/// The code is order-independent and deterministic: swapping the keys
/// produces the same result. Both sides of a key exchange compute this
/// independently; a mismatch indicates a MitM attack.
public func deriveConfirmationCode(_ pubA: Data, _ pubB: Data) -> String {
    // Sort lexicographically for order-independence.
    let (a, b) = pubA.lexicographicallyPrecedes(pubB) ? (pubA, pubB) : (pubB, pubA)
    let ikm = a + b
    let key = HKDF<SHA256>.deriveKey(
        inputKeyMaterial: SymmetricKey(data: ikm),
        salt: Data(),
        info: Data("pairing-confirmation".utf8),
        outputByteCount: 4
    )
    let bytes = key.withUnsafeBytes { Array($0) }
    let value = UInt32(bytes[0]) << 24 | UInt32(bytes[1]) << 16 | UInt32(bytes[2]) << 8 | UInt32(bytes[3])
    let code = value % 1_000_000
    return String(format: "%06d", code)
}

// MARK: - Session salt (🎯T50)

/// Length of the per-session diversifier folded into `deriveChannel`'s
/// HKDF info. Matches the activation-path nonce length (T44.1) so
/// reconnects under one PairingRecord never share an AEAD key.
public let sessionSaltLen = 16

/// Fixed peer reply to a session-salt handshake on the artifact/reconnect
/// path (plaintext on the primary stream before AEAD is established).
public let sessionSaltAck = "session_salt_ack"

/// Generate `sessionSaltLen` cryptographically random bytes for use as a
/// per-session `deriveChannel` diversifier (🎯T50).
public func generateSessionSalt() -> Data {
    var d = Data(count: sessionSaltLen)
    _ = d.withUnsafeMutableBytes { SecRandomCopyBytes(kSecRandomDefault, sessionSaltLen, $0.baseAddress!) }
    return d
}

// MARK: - Pairing record

/// Persistent state from a completed pairing ceremony.
/// Serialize to JSON (Codable) and store in Keychain or secure storage.
/// On reconnect, load it and call `deriveChannel` to derive session keys
/// without repeating the ECDH ceremony.
public struct PairingRecord: Codable, Sendable {
    public let peerInstanceID: String
    public let relayURL: String
    public let localPrivateKey: Data  // raw X25519, 32 bytes
    public let localPublicKey: Data   // raw X25519, 32 bytes
    public let peerPublicKey: Data    // raw X25519, 32 bytes

    // JSON keys use snake_case so the wire format matches the Go and
    // Kotlin SDKs. A PairingRecord (or PairingArtifact) minted in any
    // SDK can be decoded by the other two.
    enum CodingKeys: String, CodingKey {
        case peerInstanceID = "peer_instance_id"
        case relayURL = "relay_url"
        case localPrivateKey = "local_private_key"
        case localPublicKey = "local_public_key"
        case peerPublicKey = "peer_public_key"
    }

    public init(peerInstanceID: String, relayURL: String, localKeyPair: E2EKeyPair, peerPublicKey: Data) {
        self.peerInstanceID = peerInstanceID
        self.relayURL = relayURL
        self.localPrivateKey = Data(localKeyPair.privateKey.rawRepresentation)
        self.localPublicKey = localKeyPair.publicKeyData
        self.peerPublicKey = peerPublicKey
    }

    /// Construct a `PairingRecord` directly from raw key bytes. Used by
    /// the pairing-ceremony driver which already has the raw 32-byte
    /// X25519 material on hand.
    public init(peerInstanceID: String, relayURL: String,
                localPrivateKey: Data, localPublicKey: Data, peerPublicKey: Data) {
        self.peerInstanceID = peerInstanceID
        self.relayURL = relayURL
        self.localPrivateKey = localPrivateKey
        self.localPublicKey = localPublicKey
        self.peerPublicKey = peerPublicKey
    }

    /// Derive an encrypted channel from the stored keys.
    /// The info parameters should match what was used during the original pairing.
    ///
    /// `sessionSalt` is a per-session diversifier (typically
    /// `sessionSaltLen` random bytes exchanged during reconnect). When
    /// non-empty it is appended to each direction label (HKDF info =
    /// label || salt), so two sessions under one PairingRecord derive
    /// distinct AEAD keys (🎯T50). Pass empty only for tests that
    /// intentionally exercise the static pre-T50 derivation. When
    /// non-empty, `sessionSalt.count` must equal `sessionSaltLen`.
    public func deriveChannel(sendInfo: Data, recvInfo: Data, sessionSalt: Data = Data()) throws -> E2EChannel {
        if !sessionSalt.isEmpty && sessionSalt.count != sessionSaltLen {
            throw E2ECryptoError.invalidSessionSaltLength(sessionSalt.count)
        }
        let privateKey = try Curve25519.KeyAgreement.PrivateKey(rawRepresentation: localPrivateKey)
        let peerKey = try Curve25519.KeyAgreement.PublicKey(rawRepresentation: peerPublicKey)
        let shared = try privateKey.sharedSecretFromKeyAgreement(with: peerKey)

        let sendKeyInfo = sendInfo + sessionSalt
        let recvKeyInfo = recvInfo + sessionSalt
        let sendKey = shared.hkdfDerivedSymmetricKey(
            using: SHA256.self, salt: Data(), sharedInfo: sendKeyInfo, outputByteCount: 32)
        let recvKey = shared.hkdfDerivedSymmetricKey(
            using: SHA256.self, salt: Data(), sharedInfo: recvKeyInfo, outputByteCount: 32)

        return E2EChannel(sendKey: sendKey, recvKey: recvKey)
    }
}

/// Errors from the pure-crypto layer (key derivation, salt validation).
public enum E2ECryptoError: LocalizedError, Equatable {
    case invalidSessionSaltLength(Int)

    public var errorDescription: String? {
        switch self {
        case .invalidSessionSaltLength(let n):
            return "session salt must be \(sessionSaltLen) bytes, got \(n)"
        }
    }
}

// MARK: - Channel mode

/// Controls how `E2EChannel.decrypt` handles sequence numbers.
public enum ChannelMode {
    /// Strict (default): sequence numbers must be contiguous with no gaps.
    /// Any out-of-order or replayed packet is rejected. Suitable for reliable
    /// transports (TCP / WebSocket).
    case strict

    /// Datagrams: gaps in the sequence number space are accepted, as expected
    /// on lossy transports (UDP, H.264 video). Packets with a sequence number
    /// less than the last accepted one are rejected to prevent replay attacks.
    case datagrams
}

// MARK: - Encrypted channel

/// Provides symmetric encryption/decryption for a relay connection.
/// Uses AES-256-GCM with a monotonic counter nonce.
public final class E2EChannel: @unchecked Sendable {
    private let sendKey: SymmetricKey
    private let recvKey: SymmetricKey
    private var sendSeq: UInt64 = 0
    private var recvSeq: UInt64 = 0
    private let lock = NSLock()

    /// The sequence-number checking mode used by `decrypt`.
    /// Defaults to `.strict`. Set before the first `decrypt` call.
    public var mode: ChannelMode = .strict

    /// Create a channel with separate send/recv keys.
    public init(sendKey: SymmetricKey, recvKey: SymmetricKey) {
        self.sendKey = sendKey
        self.recvKey = recvKey
    }

    /// Create a symmetric channel from a shared key, deriving
    /// directional keys via HKDF.
    public convenience init(sharedKey: Data, isServer: Bool) {
        let sendInfo = isServer ? Data("server-to-client".utf8) : Data("client-to-server".utf8)
        let recvInfo = isServer ? Data("client-to-server".utf8) : Data("server-to-client".utf8)

        let sk = deriveKeyFromSecret(sharedKey, info: sendInfo)
        let rk = deriveKeyFromSecret(sharedKey, info: recvInfo)
        self.init(sendKey: sk, recvKey: rk)
    }

    /// Encrypt a plaintext message. Returns [8-byte seq][ciphertext+tag].
    public func encrypt(_ plaintext: Data) throws -> Data {
        lock.lock()
        defer { lock.unlock() }
        let seq = sendSeq
        sendSeq += 1

        var seqBytes = Data(count: 8)
        seqBytes.withUnsafeMutableBytes { ptr in
            ptr.storeBytes(of: seq.littleEndian, as: UInt64.self)
        }

        let nonce = try makeNonce(seq)
        let sealed = try AES.GCM.seal(
            plaintext,
            using: sendKey,
            nonce: nonce,
            authenticating: seqBytes
        )

        return seqBytes + sealed.ciphertext + sealed.tag
    }

    /// Decrypt a ciphertext message. Verifies the sequence number according to `mode`.
    public func decrypt(_ data: Data) throws -> Data {
        guard data.count >= 8 + 16 else { // 8 seq + 16 tag minimum
            throw E2EError.ciphertextTooShort
        }

        let seqBytes = data.prefix(8)
        let seq = seqBytes.withUnsafeBytes { $0.load(as: UInt64.self).littleEndian }
        let payload = data.dropFirst(8)

        lock.lock()
        defer { lock.unlock() }

        switch mode {
        case .strict:
            guard seq == recvSeq else {
                throw E2EError.unexpectedSequence
            }
            recvSeq += 1
        case .datagrams:
            // recvSeq holds highest-accepted-seq + 1 (0 if none yet).
            // Accept seq >= recvSeq (gaps allowed); reject seq < recvSeq (replay).
            guard seq >= recvSeq else {
                throw E2EError.sequenceReplayed
            }
            recvSeq = seq + 1
        }

        let tagStart = payload.count - 16
        let ciphertext = payload.prefix(tagStart)
        let tag = payload.suffix(16)

        let nonce = try makeNonce(seq)
        let sealedBox = try AES.GCM.SealedBox(
            nonce: nonce,
            ciphertext: ciphertext,
            tag: tag
        )

        return try AES.GCM.open(sealedBox, using: recvKey, authenticating: seqBytes)
    }

    private func makeNonce(_ seq: UInt64) throws -> AES.GCM.Nonce {
        var nonceBytes = Data(count: 12)
        nonceBytes.withUnsafeMutableBytes { ptr in
            ptr.storeBytes(of: seq.littleEndian, as: UInt64.self)
        }
        return try AES.GCM.Nonce(data: nonceBytes)
    }

    public enum E2EError: LocalizedError {
        case ciphertextTooShort
        case unexpectedSequence
        case sequenceReplayed

        public var errorDescription: String? {
            switch self {
            case .ciphertextTooShort: "Ciphertext too short"
            case .unexpectedSequence: "Unexpected sequence number"
            case .sequenceReplayed: "Sequence number replayed or too old"
            }
        }
    }
}

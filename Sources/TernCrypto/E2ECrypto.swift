// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

import CryptoKit
import Foundation

/// End-to-end encryption for relay traffic relayed through tern.
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

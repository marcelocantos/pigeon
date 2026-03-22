// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

import CryptoKit
import Foundation

/// End-to-end encryption for WebSocket traffic relayed through tern.
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

// MARK: - Encrypted channel

/// Provides symmetric encryption/decryption for a WebSocket connection.
/// Uses AES-256-GCM with a monotonic counter nonce.
public final class E2EChannel: @unchecked Sendable {
    private let sendKey: SymmetricKey
    private let recvKey: SymmetricKey
    private var sendSeq: UInt64 = 0
    private var recvSeq: UInt64 = 0
    private let lock = NSLock()

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

    /// Decrypt a ciphertext message. Verifies sequence number.
    public func decrypt(_ data: Data) throws -> Data {
        guard data.count >= 8 + 16 else { // 8 seq + 16 tag minimum
            throw E2EError.ciphertextTooShort
        }

        let seqBytes = data.prefix(8)
        let seq = seqBytes.withUnsafeBytes { $0.load(as: UInt64.self).littleEndian }
        let payload = data.dropFirst(8)

        lock.lock()
        guard seq == recvSeq else {
            lock.unlock()
            throw E2EError.unexpectedSequence
        }
        recvSeq += 1
        lock.unlock()

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

        public var errorDescription: String? {
            switch self {
            case .ciphertextTooShort: "Ciphertext too short"
            case .unexpectedSequence: "Unexpected sequence number"
            }
        }
    }
}

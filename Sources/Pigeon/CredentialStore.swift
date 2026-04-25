// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

import Foundation

#if canImport(Security)
import Security
#endif

/// Errors specific to `CredentialStore`.
public enum CredentialStoreError: LocalizedError {
    /// No credential is currently stored.
    case noCredential
    /// The platform-specific backing store reported an error.
    case backingStore(String)

    public var errorDescription: String? {
        switch self {
        case .noCredential: "No credential stored"
        case .backingStore(let m): "Credential store: \(m)"
        }
    }
}

/// A uniform interface for persisting a `PairingArtifact` across
/// processes/restarts. Backing implementations are platform-specific
/// (Keychain on iOS/macOS, file on linux/desktop fallback); the
/// interface is the same.
public protocol CredentialStore {
    /// Replaces any existing artifact.
    func save(_ artifact: PairingArtifact) throws
    /// Returns the stored artifact or throws `.noCredential`.
    func load() throws -> PairingArtifact
    /// Removes the stored artifact. No-op if absent.
    func delete() throws
    /// Convenience: loads and reports whether the stored artifact's
    /// `expiresAt` is in the past.
    func isExpired() throws -> Bool
}

#if canImport(Security)

/// Reference Keychain-backed credential store for iOS / macOS.
/// Stores the artifact's canonical JSON as a `kSecClassGenericPassword`
/// item under a configurable service+account.
public final class KeychainCredentialStore: CredentialStore {
    public let service: String
    public let account: String

    public init(service: String, account: String = "pigeon-pairing-artifact") {
        self.service = service
        self.account = account
    }

    public func save(_ artifact: PairingArtifact) throws {
        let data = try artifact.toJSON()
        try delete()
        let attributes: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: account,
            kSecValueData as String: data,
            kSecAttrAccessible as String: kSecAttrAccessibleAfterFirstUnlockThisDeviceOnly,
        ]
        let status = SecItemAdd(attributes as CFDictionary, nil)
        guard status == errSecSuccess else {
            throw CredentialStoreError.backingStore("SecItemAdd: OSStatus=\(status)")
        }
    }

    public func load() throws -> PairingArtifact {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: account,
            kSecReturnData as String: true,
            kSecMatchLimit as String: kSecMatchLimitOne,
        ]
        var result: CFTypeRef?
        let status = SecItemCopyMatching(query as CFDictionary, &result)
        if status == errSecItemNotFound {
            throw CredentialStoreError.noCredential
        }
        guard status == errSecSuccess, let data = result as? Data else {
            throw CredentialStoreError.backingStore("SecItemCopyMatching: OSStatus=\(status)")
        }
        return try PairingArtifact.fromJSON(data)
    }

    public func delete() throws {
        let query: [String: Any] = [
            kSecClass as String: kSecClassGenericPassword,
            kSecAttrService as String: service,
            kSecAttrAccount as String: account,
        ]
        let status = SecItemDelete(query as CFDictionary)
        if status != errSecSuccess && status != errSecItemNotFound {
            throw CredentialStoreError.backingStore("SecItemDelete: OSStatus=\(status)")
        }
    }

    public func isExpired() throws -> Bool {
        let a = try load()
        return a.isExpired()
    }
}

#endif

/// File-backed credential store. Useful for tests, Linux, or any
/// scenario where Keychain is unavailable. Stores the artifact JSON
/// at the given URL with permissions `0600`.
public final class FileCredentialStore: CredentialStore {
    public let url: URL
    public var now: () -> Date = { Date() }

    public init(url: URL) {
        self.url = url
    }

    public func save(_ artifact: PairingArtifact) throws {
        let data = try artifact.toJSON()
        try FileManager.default.createDirectory(
            at: url.deletingLastPathComponent(),
            withIntermediateDirectories: true,
            attributes: [.posixPermissions: 0o700])
        try data.write(to: url, options: [.atomic])
        try? FileManager.default.setAttributes(
            [.posixPermissions: 0o600], ofItemAtPath: url.path)
    }

    public func load() throws -> PairingArtifact {
        guard FileManager.default.fileExists(atPath: url.path) else {
            throw CredentialStoreError.noCredential
        }
        let data = try Data(contentsOf: url)
        return try PairingArtifact.fromJSON(data)
    }

    public func delete() throws {
        if FileManager.default.fileExists(atPath: url.path) {
            try FileManager.default.removeItem(at: url)
        }
    }

    public func isExpired() throws -> Bool {
        let a = try load()
        return a.isExpired(now: now())
    }
}

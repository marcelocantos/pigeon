// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package com.marcelocantos.pigeon.crypto

import java.io.File
import java.nio.file.Files
import java.nio.file.attribute.PosixFilePermission
import java.time.Instant

/** Thrown when no credential is currently stored. */
class NoCredentialException : Exception("pigeon: no credential stored")

/**
 * A uniform interface for persisting a [PairingArtifact] across
 * processes/restarts. Backing implementations are platform-specific
 * (EncryptedSharedPreferences on Android, file on JVM/desktop); the
 * interface is the same.
 */
interface CredentialStore {
    /** Replaces any existing artifact. */
    fun save(artifact: PairingArtifact)
    /** Returns the stored artifact or throws [NoCredentialException]. */
    fun load(): PairingArtifact
    /** Removes the stored artifact. No-op if absent. */
    fun delete()
    /**
     * Convenience: loads and reports whether the stored artifact's
     * `expiresAt` is in the past. Throws [NoCredentialException] when
     * nothing is stored.
     */
    fun isExpired(): Boolean
}

/**
 * File-backed credential store. Stores the artifact's canonical JSON
 * at the given path. On POSIX systems, restricts permissions to 0600;
 * on Windows the default ACLs apply.
 *
 * Suitable for JVM/desktop scenarios. Production deployments on
 * Android should use [EncryptedSharedPreferencesCredentialStore]
 * (which lives in the Android-specific module — not in this JVM
 * library because it requires `androidx.security`).
 */
class FileCredentialStore(
    val path: File,
    var now: () -> Instant = { Instant.now() },
) : CredentialStore {

    override fun save(artifact: PairingArtifact) {
        val parent = path.parentFile
        if (parent != null && !parent.exists()) {
            parent.mkdirs()
        }
        val tmp = File(path.absolutePath + ".tmp")
        tmp.writeText(artifact.toJson(), Charsets.UTF_8)
        if (!tmp.renameTo(path)) {
            tmp.delete()
            throw java.io.IOException("rename ${tmp.absolutePath} -> ${path.absolutePath}")
        }
        applyOwnerOnlyPermissions(path)
    }

    override fun load(): PairingArtifact {
        if (!path.exists()) throw NoCredentialException()
        val text = path.readText(Charsets.UTF_8)
        return PairingArtifact.fromJson(text)
    }

    override fun delete() {
        if (path.exists() && !path.delete()) {
            throw java.io.IOException("delete ${path.absolutePath}")
        }
    }

    override fun isExpired(): Boolean = load().isExpired(now())
}

private fun applyOwnerOnlyPermissions(file: File) {
    val supported = file.toPath().fileSystem.supportedFileAttributeViews().contains("posix")
    if (!supported) return
    val perms = setOf(PosixFilePermission.OWNER_READ, PosixFilePermission.OWNER_WRITE)
    Files.setPosixFilePermissions(file.toPath(), perms)
}

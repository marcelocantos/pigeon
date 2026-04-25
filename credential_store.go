// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package pigeon

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ErrNoCredential is returned by CredentialStore.Load when no artifact
// has been saved.
var ErrNoCredential = errors.New("pigeon: no credential stored")

// CredentialStore is a uniform interface for persisting a
// PairingArtifact across processes/restarts. Backing implementations
// are platform-specific (file+OS-keyring on desktop, Keychain on iOS,
// EncryptedSharedPreferences on Android); the interface is the same.
//
// Save replaces any existing artifact. Load returns ErrNoCredential
// when nothing is stored. Delete removes the stored artifact (no-op
// if absent). IsExpired is a convenience that loads and checks
// expiry; it returns (false, ErrNoCredential) if nothing is stored.
type CredentialStore interface {
	Save(artifact *PairingArtifact) error
	Load() (*PairingArtifact, error)
	Delete() error
	IsExpired() (bool, error)
}

// FileCredentialStore is a reference implementation that persists the
// artifact as canonical JSON in a single file with 0600 permissions.
// Suitable for desktop scenarios. Production deployments on iOS or
// Android should use the Keychain or EncryptedSharedPreferences
// implementations in the Swift / Kotlin SDKs respectively.
type FileCredentialStore struct {
	// Path is the absolute file path where the artifact is stored.
	Path string
	// Now is injectable for testing; defaults to time.Now.
	Now func() time.Time
}

// NewFileCredentialStore returns a store backed by the given path. The
// parent directory is created on first Save.
func NewFileCredentialStore(path string) *FileCredentialStore {
	return &FileCredentialStore{Path: path}
}

func (s *FileCredentialStore) now() time.Time {
	if s.Now != nil {
		return s.Now()
	}
	return time.Now()
}

// Save writes the artifact to disk, replacing any existing file.
func (s *FileCredentialStore) Save(a *PairingArtifact) error {
	if a == nil {
		return errors.New("pigeon: cannot save nil artifact")
	}
	data, err := a.Marshal()
	if err != nil {
		return fmt.Errorf("marshal artifact: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(s.Path), 0o700); err != nil {
		return fmt.Errorf("create credential dir: %w", err)
	}
	tmp := s.Path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return fmt.Errorf("write credential: %w", err)
	}
	if err := os.Rename(tmp, s.Path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("rename credential: %w", err)
	}
	return nil
}

// Load reads the artifact from disk. Returns ErrNoCredential if the
// file does not exist.
func (s *FileCredentialStore) Load() (*PairingArtifact, error) {
	data, err := os.ReadFile(s.Path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrNoCredential
		}
		return nil, fmt.Errorf("read credential: %w", err)
	}
	return UnmarshalPairingArtifact(data)
}

// Delete removes the stored artifact. No-op if no artifact is stored.
func (s *FileCredentialStore) Delete() error {
	if err := os.Remove(s.Path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("delete credential: %w", err)
	}
	return nil
}

// IsExpired loads the artifact and reports whether its ExpiresAt is in
// the past at the store's current time.
func (s *FileCredentialStore) IsExpired() (bool, error) {
	a, err := s.Load()
	if err != nil {
		return false, err
	}
	return a.IsExpired(s.now()), nil
}

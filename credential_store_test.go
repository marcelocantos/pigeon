// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package pigeon

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func newTestArtifact(t *testing.T, ttl time.Duration) (*PairingArtifact, time.Time) {
	t.Helper()
	cred, _, err := IssueCredential("inst-test", "https://relay.example.com")
	if err != nil {
		t.Fatal("IssueCredential:", err)
	}
	issued := time.Date(2026, 4, 25, 12, 0, 0, 0, time.UTC)
	return NewPairingArtifact(cred, "tok", issued, ttl), issued
}

func TestFileCredentialStoreRoundTrip(t *testing.T) {
	dir := t.TempDir()
	store := NewFileCredentialStore(filepath.Join(dir, "nested", "artifact.json"))

	if _, err := store.Load(); !errors.Is(err, ErrNoCredential) {
		t.Fatalf("Load on empty store: got %v, want ErrNoCredential", err)
	}
	if _, err := store.IsExpired(); !errors.Is(err, ErrNoCredential) {
		t.Fatalf("IsExpired on empty store: got %v, want ErrNoCredential", err)
	}

	a, _ := newTestArtifact(t, DefaultPairingTTL)
	if err := store.Save(a); err != nil {
		t.Fatal("Save:", err)
	}

	info, err := os.Stat(store.Path)
	if err != nil {
		t.Fatal("Stat:", err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Errorf("file mode = %v, want 0600", info.Mode().Perm())
	}

	restored, err := store.Load()
	if err != nil {
		t.Fatal("Load:", err)
	}
	if restored.Token != a.Token {
		t.Errorf("token = %q, want %q", restored.Token, a.Token)
	}
	if restored.Record.PeerInstanceID != a.Record.PeerInstanceID {
		t.Errorf("peer instance ID = %q, want %q", restored.Record.PeerInstanceID, a.Record.PeerInstanceID)
	}

	if err := store.Delete(); err != nil {
		t.Fatal("Delete:", err)
	}
	if _, err := os.Stat(store.Path); !errors.Is(err, os.ErrNotExist) {
		t.Error("file still present after Delete")
	}
	if err := store.Delete(); err != nil {
		t.Errorf("Delete on absent store: %v", err)
	}
}

func TestFileCredentialStoreIsExpired(t *testing.T) {
	dir := t.TempDir()
	store := NewFileCredentialStore(filepath.Join(dir, "artifact.json"))

	a, issued := newTestArtifact(t, 24*time.Hour)
	if err := store.Save(a); err != nil {
		t.Fatal("Save:", err)
	}

	store.Now = func() time.Time { return issued.Add(time.Hour) }
	expired, err := store.IsExpired()
	if err != nil {
		t.Fatal("IsExpired:", err)
	}
	if expired {
		t.Error("artifact should not be expired one hour after issuance")
	}

	store.Now = func() time.Time { return issued.Add(48 * time.Hour) }
	expired, err = store.IsExpired()
	if err != nil {
		t.Fatal("IsExpired:", err)
	}
	if !expired {
		t.Error("artifact should be expired 48 hours after issuance with 24h TTL")
	}
}

func TestFileCredentialStoreSaveNil(t *testing.T) {
	store := NewFileCredentialStore(filepath.Join(t.TempDir(), "artifact.json"))
	if err := store.Save(nil); err == nil {
		t.Fatal("Save(nil) should error")
	}
}

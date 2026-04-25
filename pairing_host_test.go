// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package pigeon

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestPairingHostMintHappyPath(t *testing.T) {
	fixed := time.Date(2026, 4, 25, 12, 0, 0, 0, time.UTC)
	host := &PairingHost{
		RelayURL: "https://relay.example.com",
		TTL:      DefaultPairingTTL,
		Now:      func() time.Time { return fixed },
		IssueToken: func(peerInstanceID string) (string, error) {
			return "tok-" + peerInstanceID, nil
		},
	}

	artifact, serverRec, err := host.Mint("inst-42")
	if err != nil {
		t.Fatal("Mint:", err)
	}
	if artifact.Token != "tok-inst-42" {
		t.Errorf("token = %q, want tok-inst-42", artifact.Token)
	}
	if !artifact.IssuedAt.Equal(fixed) {
		t.Errorf("issuedAt = %v, want %v", artifact.IssuedAt, fixed)
	}
	if !artifact.ExpiresAt.Equal(fixed.Add(DefaultPairingTTL)) {
		t.Errorf("expiresAt = %v, want %v", artifact.ExpiresAt, fixed.Add(DefaultPairingTTL))
	}
	if artifact.Record.RelayURL != host.RelayURL {
		t.Errorf("relay URL = %q, want %q", artifact.Record.RelayURL, host.RelayURL)
	}
	if artifact.Record.PeerInstanceID != "inst-42" {
		t.Errorf("peer instance ID = %q, want inst-42", artifact.Record.PeerInstanceID)
	}

	// Server record's peer key matches device's local key (and vice versa).
	if string(serverRec.PeerPublicKey) != string(artifact.Record.LocalPublicKey) {
		t.Error("server peer key does not match device local key")
	}

	// Both sides derive a matching channel.
	deviceCh, err := artifact.Record.DeriveChannel([]byte("client-to-server"), []byte("server-to-client"))
	if err != nil {
		t.Fatal("device derive:", err)
	}
	serverCh, err := serverRec.DeriveChannel([]byte("server-to-client"), []byte("client-to-server"))
	if err != nil {
		t.Fatal("server derive:", err)
	}
	ct := deviceCh.Encrypt([]byte("hello"))
	pt, err := serverCh.Decrypt(ct)
	if err != nil {
		t.Fatal("server decrypt:", err)
	}
	if string(pt) != "hello" {
		t.Errorf("decrypted = %q, want hello", pt)
	}
}

func TestPairingHostDefaultTTL(t *testing.T) {
	fixed := time.Date(2026, 4, 25, 12, 0, 0, 0, time.UTC)
	host := &PairingHost{
		RelayURL: "https://relay.example.com",
		Now:      func() time.Time { return fixed },
	}
	a, _, err := host.Mint("inst-1")
	if err != nil {
		t.Fatal("Mint:", err)
	}
	if !a.ExpiresAt.Equal(fixed.Add(DefaultPairingTTL)) {
		t.Errorf("expected default TTL of %v, got expiry %v", DefaultPairingTTL, a.ExpiresAt)
	}
}

func TestPairingHostNegativeTTLNeverExpires(t *testing.T) {
	host := &PairingHost{RelayURL: "https://relay.example.com", TTL: -1}
	a, _, err := host.Mint("inst-1")
	if err != nil {
		t.Fatal("Mint:", err)
	}
	if !a.ExpiresAt.IsZero() {
		t.Errorf("negative TTL should produce zero ExpiresAt, got %v", a.ExpiresAt)
	}
	if a.IsExpired(time.Now().Add(100 * 365 * 24 * time.Hour)) {
		t.Error("negative-TTL artifact should never expire")
	}
}

func TestPairingHostMintRejectsBadInputs(t *testing.T) {
	host := &PairingHost{RelayURL: ""}
	if _, _, err := host.Mint("inst"); err == nil {
		t.Error("Mint with empty RelayURL should error")
	}
	host2 := NewPairingHost("https://relay.example.com")
	if _, _, err := host2.Mint(""); err == nil {
		t.Error("Mint with empty peer instance ID should error")
	}
}

func TestPairingHostExpiredArtifactRejectedByConnect(t *testing.T) {
	// The TTL the host issues governs whether ConnectWithArtifact will
	// accept the artifact. An artifact past its TTL should be refused
	// before any network IO happens.
	host := &PairingHost{
		RelayURL: "https://relay.example.com",
		TTL:      time.Hour,
		Now:      func() time.Time { return time.Now().Add(-2 * time.Hour) }, // issued 2h ago
	}
	stale, _, err := host.Mint("inst-stale")
	if err != nil {
		t.Fatal("Mint:", err)
	}
	if !stale.IsExpired(time.Now()) {
		t.Fatal("setup: artifact should be expired by now")
	}

	_, err = ConnectWithArtifact(context.Background(), stale, Config{})
	if !errors.Is(err, ErrPairingExpired) {
		t.Fatalf("want ErrPairingExpired, got %v", err)
	}
}

func TestPairingHostIssueTokenError(t *testing.T) {
	host := &PairingHost{
		RelayURL: "https://relay.example.com",
		IssueToken: func(string) (string, error) {
			return "", errors.New("token service down")
		},
	}
	if _, _, err := host.Mint("inst-1"); err == nil {
		t.Error("expected error from failing IssueToken")
	}
}

// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package pigeon

import (
	"errors"
	"testing"
	"time"
)

func TestPairingArtifactRoundTripJSON(t *testing.T) {
	deviceCred, _, err := IssueCredential("inst-abc", "https://relay.example.com")
	if err != nil {
		t.Fatal("IssueCredential:", err)
	}

	issued := time.Date(2026, 4, 25, 12, 0, 0, 0, time.UTC)
	a := NewPairingArtifact(deviceCred, "tok-xyz", issued, DefaultPairingTTL)

	data, err := a.Marshal()
	if err != nil {
		t.Fatal("Marshal:", err)
	}
	restored, err := UnmarshalPairingArtifact(data)
	if err != nil {
		t.Fatal("Unmarshal:", err)
	}

	if restored.Token != "tok-xyz" {
		t.Errorf("token = %q, want tok-xyz", restored.Token)
	}
	if !restored.IssuedAt.Equal(issued) {
		t.Errorf("issuedAt = %v, want %v", restored.IssuedAt, issued)
	}
	if !restored.ExpiresAt.Equal(issued.Add(DefaultPairingTTL)) {
		t.Errorf("expiresAt = %v, want %v", restored.ExpiresAt, issued.Add(DefaultPairingTTL))
	}
	if restored.Record == nil {
		t.Fatal("record nil after round-trip")
	}
	if restored.Record.PeerInstanceID != "inst-abc" {
		t.Errorf("peer instance ID = %q, want inst-abc", restored.Record.PeerInstanceID)
	}
	if string(restored.Record.PeerPublicKey) != string(deviceCred.PeerPublicKey) {
		t.Error("peer public key not preserved across round-trip")
	}
}

func TestPairingArtifactRoundTripText(t *testing.T) {
	deviceCred, _, err := IssueCredential("inst-abc", "https://relay.example.com")
	if err != nil {
		t.Fatal("IssueCredential:", err)
	}

	issued := time.Date(2026, 4, 25, 12, 0, 0, 0, time.UTC)
	a := NewPairingArtifact(deviceCred, "tok-xyz", issued, DefaultPairingTTL)

	text, err := a.MarshalText()
	if err != nil {
		t.Fatal("MarshalText:", err)
	}
	if len(text) == 0 {
		t.Fatal("empty text encoding")
	}
	for _, b := range text {
		if b == '+' || b == '/' || b == '=' {
			t.Fatalf("text encoding should be base64url with no padding, got %q", string(text))
		}
	}

	restored, err := ParsePairingArtifactText(text)
	if err != nil {
		t.Fatal("ParsePairingArtifactText:", err)
	}
	if restored.Token != "tok-xyz" {
		t.Errorf("token = %q, want tok-xyz", restored.Token)
	}
	if !restored.IssuedAt.Equal(issued) {
		t.Errorf("issuedAt = %v, want %v", restored.IssuedAt, issued)
	}
	if restored.Record.PeerInstanceID != "inst-abc" {
		t.Errorf("peer instance ID = %q, want inst-abc", restored.Record.PeerInstanceID)
	}
}

func TestPairingArtifactIsExpired(t *testing.T) {
	deviceCred, _, err := IssueCredential("inst-abc", "https://relay.example.com")
	if err != nil {
		t.Fatal("IssueCredential:", err)
	}

	issued := time.Date(2026, 4, 25, 12, 0, 0, 0, time.UTC)
	a := NewPairingArtifact(deviceCred, "", issued, 24*time.Hour)

	if a.IsExpired(issued) {
		t.Error("artifact should not be expired at issuance instant")
	}
	if a.IsExpired(issued.Add(23 * time.Hour)) {
		t.Error("artifact should not be expired before TTL")
	}
	if !a.IsExpired(issued.Add(24 * time.Hour)) {
		t.Error("artifact should be expired exactly at TTL boundary")
	}
	if !a.IsExpired(issued.Add(48 * time.Hour)) {
		t.Error("artifact should be expired well past TTL")
	}

	// Zero ExpiresAt means never expires.
	noExpiry := NewPairingArtifact(deviceCred, "", issued, 0)
	if !noExpiry.ExpiresAt.IsZero() {
		t.Error("ttl=0 should produce zero ExpiresAt")
	}
	if noExpiry.IsExpired(issued.Add(100 * 365 * 24 * time.Hour)) {
		t.Error("ttl=0 artifact should never expire")
	}
}

func TestPairingArtifactErrSentinel(t *testing.T) {
	wrapped := errors.Join(errors.New("connect failed"), ErrPairingExpired)
	if !errors.Is(wrapped, ErrPairingExpired) {
		t.Error("errors.Is should match ErrPairingExpired through wrapping")
	}
}

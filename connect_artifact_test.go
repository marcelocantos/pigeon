// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package pigeon

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestConnectWithArtifactExpired(t *testing.T) {
	deviceCred, _, err := IssueCredential("inst-abc", "https://relay.example.com")
	if err != nil {
		t.Fatal("IssueCredential:", err)
	}

	// Artifact issued 31 days ago with the default 30-day TTL — expired.
	stale := NewPairingArtifact(deviceCred, "tok", time.Now().Add(-31*24*time.Hour), DefaultPairingTTL)

	_, err = ConnectWithArtifact(context.Background(), stale, Config{})
	if err == nil {
		t.Fatal("expected error from expired artifact")
	}
	if !errors.Is(err, ErrPairingExpired) {
		t.Fatalf("expected errors.Is(err, ErrPairingExpired), got %v", err)
	}
}

func TestConnectWithArtifactRejectsMissingFields(t *testing.T) {
	cases := []struct {
		name string
		a    *PairingArtifact
	}{
		{"nil", nil},
		{"no record", &PairingArtifact{ExpiresAt: time.Now().Add(time.Hour)}},
		{"no relay URL", func() *PairingArtifact {
			cred, _, _ := IssueCredential("inst", "https://relay")
			cred.RelayURL = ""
			return NewPairingArtifact(cred, "", time.Now(), time.Hour)
		}()},
		{"no instance ID", func() *PairingArtifact {
			cred, _, _ := IssueCredential("inst", "https://relay")
			cred.PeerInstanceID = ""
			return NewPairingArtifact(cred, "", time.Now(), time.Hour)
		}()},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := ConnectWithArtifact(context.Background(), tc.a, Config{}); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}

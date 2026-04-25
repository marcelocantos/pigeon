// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package pigeon

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/marcelocantos/pigeon/crypto"
)

// DefaultPairingTTL is the default lifetime of a PairingArtifact when
// none is specified. After this period elapses the client must re-pair.
const DefaultPairingTTL = 30 * 24 * time.Hour

// ErrPairingExpired is returned when an attempt is made to use a
// PairingArtifact past its ExpiresAt timestamp. Callers should match
// it with errors.Is and route the user to a re-pair flow.
var ErrPairingExpired = errors.New("pigeon: pairing artifact expired")

// PairingArtifact is the persistable+expirable envelope around a
// completed pairing. The cryptographic core lives in Record (a
// crypto.PairingRecord); the wrapping fields carry the lifecycle
// metadata that lets a client detect expiry and prompt re-pair.
//
// The artifact is what gets QR-encoded for transport from the
// pairing host to the client device, and what a CredentialStore
// persists across restarts.
type PairingArtifact struct {
	// Record is the cryptographic state from a completed pairing
	// ceremony. It is non-nil for any artifact produced by
	// PairingHost or NewPairingArtifact.
	Record *crypto.PairingRecord `json:"record"`

	// Token is the bearer token the client presents when reconnecting
	// to the relay. Empty if the relay does not require token auth.
	Token string `json:"token,omitempty"`

	// IssuedAt is when this artifact was minted.
	IssuedAt time.Time `json:"issued_at"`

	// ExpiresAt is when this artifact stops being valid. The client
	// must re-pair after this point. Zero means no expiry — use
	// sparingly and only for short-lived development scenarios.
	ExpiresAt time.Time `json:"expires_at"`
}

// NewPairingArtifact constructs an artifact from a freshly produced
// PairingRecord and a TTL. issuedAt defaults to time.Now() if zero.
// A zero TTL produces an artifact with no expiry.
func NewPairingArtifact(record *crypto.PairingRecord, token string, issuedAt time.Time, ttl time.Duration) *PairingArtifact {
	if issuedAt.IsZero() {
		issuedAt = time.Now()
	}
	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = issuedAt.Add(ttl)
	}
	return &PairingArtifact{
		Record:    record,
		Token:     token,
		IssuedAt:  issuedAt,
		ExpiresAt: expiresAt,
	}
}

// IsExpired reports whether the artifact's ExpiresAt is in the past
// at the given instant. An artifact with zero ExpiresAt never expires.
func (a *PairingArtifact) IsExpired(now time.Time) bool {
	if a.ExpiresAt.IsZero() {
		return false
	}
	return !now.Before(a.ExpiresAt)
}

// artifactJSON is an alias of PairingArtifact stripped of method
// promotion. Encoding through this type bypasses the TextMarshaler
// interface that PairingArtifact implements, preventing infinite
// recursion when MarshalText calls json.Marshal.
type artifactJSON PairingArtifact

// Marshal serialises the artifact to canonical JSON. The same encoding
// is used across all SDKs so an artifact minted in Go can be decoded
// by the Swift or Kotlin clients (and vice versa).
//
// Keep the artifact small: QR codes practically cap a scannable
// payload at ~2-3 KB. The base fields fit comfortably; resist adding
// new fields without weighing the QR-payload impact.
func (a *PairingArtifact) Marshal() ([]byte, error) {
	return json.Marshal((*artifactJSON)(a))
}

// UnmarshalPairingArtifact deserialises an artifact from canonical JSON.
func UnmarshalPairingArtifact(data []byte) (*PairingArtifact, error) {
	var a artifactJSON
	if err := json.Unmarshal(data, &a); err != nil {
		return nil, err
	}
	out := PairingArtifact(a)
	return &out, nil
}

// MarshalText returns the canonical single-line text encoding:
// base64url(jsonBytes) with no padding. This form is suitable for
// transport via QR payload, launch argument, environment variable,
// pasteboard, or any other channel that wants a single token of text.
func (a *PairingArtifact) MarshalText() ([]byte, error) {
	jsonBytes, err := a.Marshal()
	if err != nil {
		return nil, err
	}
	enc := base64.RawURLEncoding
	out := make([]byte, enc.EncodedLen(len(jsonBytes)))
	enc.Encode(out, jsonBytes)
	return out, nil
}

// UnmarshalText decodes the canonical text encoding produced by
// MarshalText into the receiver.
func (a *PairingArtifact) UnmarshalText(text []byte) error {
	enc := base64.RawURLEncoding
	jsonBytes := make([]byte, enc.DecodedLen(len(text)))
	n, err := enc.Decode(jsonBytes, text)
	if err != nil {
		return fmt.Errorf("decode pairing artifact: %w", err)
	}
	var decoded artifactJSON
	if err := json.Unmarshal(jsonBytes[:n], &decoded); err != nil {
		return fmt.Errorf("parse pairing artifact: %w", err)
	}
	*a = PairingArtifact(decoded)
	return nil
}

// ParsePairingArtifactText is the package-level helper inverse of
// (*PairingArtifact).MarshalText.
func ParsePairingArtifactText(text []byte) (*PairingArtifact, error) {
	var a PairingArtifact
	if err := a.UnmarshalText(text); err != nil {
		return nil, err
	}
	return &a, nil
}

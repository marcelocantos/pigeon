// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package pigeon

import (
	"errors"
	"time"

	"github.com/marcelocantos/pigeon/crypto"
)

// PairingHost is the server-side counterpart to PairingArtifact. It
// produces artifacts that a client device can persist and present on
// reconnect, without running the visual pairing ceremony. Use it when
// the server can authenticate the client out-of-band — e.g. an admin
// API issuing credentials for managed devices, or a developer-deploy
// pipeline injecting credentials via xcrun during install.
//
// The artifact carries a TTL (default 30 days). When the TTL expires,
// the client's next ConnectWithArtifact call returns ErrPairingExpired
// and the client must re-pair.
type PairingHost struct {
	// RelayURL is the relay URL embedded in every minted artifact.
	// Required.
	RelayURL string

	// TTL is the lifetime of every minted artifact. Defaults to
	// DefaultPairingTTL (30 days). Set to a negative value to
	// produce non-expiring artifacts (use sparingly).
	TTL time.Duration

	// Now is injectable for testing; defaults to time.Now.
	Now func() time.Time

	// IssueToken, if set, mints a bearer token to embed in each
	// artifact. The token is not interpreted by the host — wire it
	// up with whatever auth scheme the relay uses. Returns the empty
	// string when unset.
	IssueToken func(peerInstanceID string) (string, error)
}

// NewPairingHost is a convenience constructor.
func NewPairingHost(relayURL string) *PairingHost {
	return &PairingHost{RelayURL: relayURL, TTL: DefaultPairingTTL}
}

func (h *PairingHost) now() time.Time {
	if h.Now != nil {
		return h.Now()
	}
	return time.Now()
}

func (h *PairingHost) ttl() time.Duration {
	if h.TTL == 0 {
		return DefaultPairingTTL
	}
	if h.TTL < 0 {
		return 0
	}
	return h.TTL
}

// Mint issues a fresh PairingArtifact for the given peer instance ID.
// It generates the device + server keypairs, derives matching
// PairingRecords for both sides, and wraps the device-side record in
// an artifact with the host's TTL. The returned serverRecord should
// be registered with the auth sub-machine so subsequent auth_request
// messages from the device are accepted.
//
// Equivalent to IssueCredential plus PairingArtifact wrapping, plus
// optional token issuance.
func (h *PairingHost) Mint(peerInstanceID string) (artifact *PairingArtifact, serverRecord *crypto.PairingRecord, err error) {
	if h.RelayURL == "" {
		return nil, nil, errors.New("pigeon: PairingHost requires RelayURL")
	}
	if peerInstanceID == "" {
		return nil, nil, errors.New("pigeon: Mint requires peerInstanceID")
	}

	deviceCred, serverRec, err := IssueCredential(peerInstanceID, h.RelayURL)
	if err != nil {
		return nil, nil, err
	}

	var token string
	if h.IssueToken != nil {
		token, err = h.IssueToken(peerInstanceID)
		if err != nil {
			return nil, nil, err
		}
	}

	a := NewPairingArtifact(deviceCred, token, h.now(), h.ttl())
	return a, serverRec, nil
}

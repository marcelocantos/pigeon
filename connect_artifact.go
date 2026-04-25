// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package pigeon

import (
	"context"
	"fmt"
	"time"
)

// Direction info strings used by the artifact-driven connect helpers
// when deriving the encrypted channel. They mirror the conventions
// already established by IssueCredential and the channel test suite.
const (
	clientToServerInfo = "client-to-server"
	serverToClientInfo = "server-to-client"
)

// ConnectWithArtifact dials the relay using the persisted PairingArtifact
// and wires the resulting Conn for encrypted traffic. The relay URL and
// peer instance ID are taken from the artifact; the bearer token is
// applied automatically. Caller may pass extra Config to override
// transport-level settings (TLS, WebTransport, LAN, etc.); fields
// already provided by the artifact (Token) take precedence.
//
// Returns ErrPairingExpired (matchable via errors.Is) when the artifact
// is past its ExpiresAt. Callers should trap this and route the user
// to a re-pair flow.
func ConnectWithArtifact(ctx context.Context, a *PairingArtifact, c Config) (*Conn, error) {
	if a == nil || a.Record == nil {
		return nil, fmt.Errorf("pigeon: pairing artifact missing record")
	}
	if a.IsExpired(time.Now()) {
		return nil, fmt.Errorf("%w: expired at %s", ErrPairingExpired, a.ExpiresAt.Format(time.RFC3339))
	}
	if a.Record.RelayURL == "" {
		return nil, fmt.Errorf("pigeon: pairing artifact missing relay URL")
	}
	if a.Record.PeerInstanceID == "" {
		return nil, fmt.Errorf("pigeon: pairing artifact missing peer instance ID")
	}

	if c.Token == "" {
		c.Token = a.Token
	}

	conn, err := Connect(ctx, a.Record.RelayURL, a.Record.PeerInstanceID, c)
	if err != nil {
		return nil, err
	}

	ch, err := a.Record.DeriveChannel([]byte(clientToServerInfo), []byte(serverToClientInfo))
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("derive channel from artifact: %w", err)
	}
	conn.SetChannel(ch)
	conn.SetPairingRecord(a.Record)

	return conn, nil
}

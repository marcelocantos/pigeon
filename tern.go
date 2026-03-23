// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Package tern provides client-side connectivity to a tern relay server.
// Backends call Register to obtain an instance ID; clients call Connect
// with a known instance ID. Both return a Conn for bidirectional
// message exchange over WebTransport (QUIC).
//
// After establishing an encrypted channel (via crypto.Channel), call
// Conn.SetChannel to enable automatic encryption on the primary stream,
// and Conn.SetDatagramChannel for encrypted datagrams.
//
// Sub-packages provide E2E encryption (crypto/), protocol state machines
// (protocol/), and QR code rendering (qr/).
package tern

import (
	"context"
	"crypto/tls"
	_ "embed"
	"fmt"
	"net/http"

	"github.com/quic-go/webtransport-go"
)

//go:embed agents-guide.md
var AgentGuide string

// Option configures a relay connection.
type Option func(*options)

type options struct {
	token     string
	tlsConfig *tls.Config
}

// WithToken sets the bearer token for authentication on /register.
func WithToken(token string) Option {
	return func(o *options) { o.token = token }
}

// WithTLS sets the TLS config for the QUIC connection. Use this to
// trust self-signed certificates (set RootCAs or InsecureSkipVerify).
func WithTLS(tlsConfig *tls.Config) Option {
	return func(o *options) { o.tlsConfig = tlsConfig }
}

func buildOptions(opts []Option) options {
	var o options
	for _, fn := range opts {
		fn(&o)
	}
	return o
}

// Register connects to the relay's /register endpoint as a backend
// over WebTransport. The relay assigns an instance ID, returned via
// InstanceID(). The caller is responsible for closing the connection.
func Register(ctx context.Context, relayURL string, opts ...Option) (*Conn, error) {
	o := buildOptions(opts)

	tlsConfig := o.tlsConfig
	if tlsConfig == nil {
		tlsConfig = &tls.Config{}
	}

	d := webtransport.Dialer{
		TLSClientConfig: tlsConfig,
	}

	hdr := http.Header{}
	if o.token != "" {
		hdr.Set("Authorization", "Bearer "+o.token)
	}

	_, session, err := d.Dial(ctx, relayURL+"/register", hdr)
	if err != nil {
		return nil, fmt.Errorf("register: %w", err)
	}

	// Open the bidirectional stream for message relay.
	stream, err := session.OpenStream()
	if err != nil {
		session.CloseWithError(0, "failed to open stream")
		return nil, fmt.Errorf("register: open stream: %w", err)
	}

	// Send a handshake to trigger the stream header.
	if err := writeWTMessage(stream, []byte("register")); err != nil {
		session.CloseWithError(0, "failed to send handshake")
		return nil, fmt.Errorf("register: handshake: %w", err)
	}

	// Read the instance ID.
	idBytes, err := readWTMessage(stream)
	if err != nil {
		session.CloseWithError(0, "failed to read ID")
		return nil, fmt.Errorf("register: read ID: %w", err)
	}

	return newConn(session, stream, string(idBytes)), nil
}

// Connect connects to a relay as a client over WebTransport, targeting
// a specific backend instance ID.
func Connect(ctx context.Context, relayURL, instanceID string, opts ...Option) (*Conn, error) {
	o := buildOptions(opts)

	tlsConfig := o.tlsConfig
	if tlsConfig == nil {
		tlsConfig = &tls.Config{}
	}

	d := webtransport.Dialer{
		TLSClientConfig: tlsConfig,
	}

	_, session, err := d.Dial(ctx, relayURL+"/ws/"+instanceID, nil)
	if err != nil {
		return nil, fmt.Errorf("connect to %s: %w", instanceID, err)
	}

	// Open the bidirectional stream for message relay.
	stream, err := session.OpenStream()
	if err != nil {
		session.CloseWithError(0, "failed to open stream")
		return nil, fmt.Errorf("connect: open stream: %w", err)
	}

	// Send a handshake to trigger the stream header.
	if err := writeWTMessage(stream, []byte("connect")); err != nil {
		session.CloseWithError(0, "failed to send handshake")
		return nil, fmt.Errorf("connect: handshake: %w", err)
	}

	return newConn(session, stream, instanceID), nil
}

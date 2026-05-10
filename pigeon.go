// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Package pigeon provides client-side connectivity to a pigeon relay server.
// Backends call Register to obtain an instance ID; clients call Connect
// with a known instance ID. Both return a Conn for bidirectional
// message exchange over QUIC.
//
// By default, Register and Connect use raw QUIC (ALPN "pigeon") for
// native clients. Use WithWebTransport() for browser-oriented paths
// that require WebTransport (HTTP/3).
//
// After establishing an encrypted channel (via crypto.Channel), call
// Conn.SetChannel to enable automatic encryption on the primary stream,
// and Conn.SetDatagramChannel for encrypted datagrams.
//
// Sub-packages provide E2E encryption (crypto/), protocol state machines
// (protocol/), and QR code rendering (qr/).
package pigeon

import (
	"context"
	"crypto/tls"
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/webtransport-go"
)

//go:embed agents-guide.md
var AgentGuide string

// Config configures a relay connection.
type Config struct {
	// Token is the bearer token for authentication on /register.
	Token string

	// InstanceID sets a persistent instance ID for registration. If set,
	// the relay uses this ID instead of generating a random one.
	InstanceID string

	// TLS is the TLS config for the QUIC connection. Use this to trust
	// self-signed certificates (set RootCAs or InsecureSkipVerify).
	TLS *tls.Config

	// WebTransport forces WebTransport (HTTP/3) instead of raw QUIC.
	WebTransport bool

	// QUICPort overrides the default QUIC port (4433) for raw QUIC
	// connections.
	QUICPort string

	// LANServer, if set, advertises a local LAN listener to connecting
	// clients. Use with Register — when SetChannel is called, the LAN
	// address is sent to the peer via the encrypted relay channel.
	LANServer *LANServer

	// LAN enables LAN upgrade on the client side. When the backend
	// advertises a LAN address, the client attempts a direct connection.
	LAN bool

	// LANTLS is the TLS config for LAN connections (client side).
	// If nil and LAN is true, InsecureSkipVerify is used.
	LANTLS *tls.Config
}

// WakeRelay sends an HTTPS request to the relay's /health endpoint,
// which triggers Fly.io's proxy to start a stopped machine. Call this
// before Register or Connect if the relay may be auto-stopped.
// The relay URL should be the same one passed to Register/Connect.
func WakeRelay(ctx context.Context, relayURL string, c Config) error {
	tlsConfig := c.TLS
	if tlsConfig == nil {
		tlsConfig = &tls.Config{}
	}

	client := &http.Client{
		Transport: &http.Transport{TLSClientConfig: tlsConfig},
		Timeout:   10 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", relayURL+"/health", nil)
	if err != nil {
		return fmt.Errorf("wake relay: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("wake relay: %w", err)
	}
	resp.Body.Close()
	return nil
}

// --- Raw QUIC client ---

func quicTLSConfig(c Config) *tls.Config {
	cfg := c.TLS
	if cfg == nil {
		cfg = &tls.Config{}
	} else {
		cfg = cfg.Clone()
	}
	cfg.NextProtos = []string{pigeonALPN}
	return cfg
}

// quicAddr derives the raw QUIC address from a relay URL. Precedence:
// (1) explicit Config.QUICPort, (2) port encoded in the URL, (3) 4433.
func quicAddr(relayURL string, c Config) (string, error) {
	u, err := url.Parse(relayURL)
	if err != nil {
		return "", fmt.Errorf("parse relay URL: %w", err)
	}
	host := u.Hostname()
	port := c.QUICPort
	if port == "" {
		port = u.Port()
	}
	if port == "" {
		port = "4433"
	}
	return host + ":" + port, nil
}

// quicCloser wraps *quic.Conn to satisfy io.Closer.
type quicCloser struct {
	conn *quic.Conn
}

func (c quicCloser) Close() error {
	return c.conn.CloseWithError(0, "")
}

// quicOpener adapts *quic.Conn to the streamOpener interface.
type quicOpener struct{ conn *quic.Conn }

func (o quicOpener) OpenStream() (io.ReadWriteCloser, error) {
	return o.conn.OpenStream()
}

type quicAcceptor struct{ conn *quic.Conn }

func (a quicAcceptor) AcceptStream(ctx context.Context) (io.ReadWriteCloser, error) {
	return a.conn.AcceptStream(ctx)
}

// wtCloser wraps webtransport.Session to satisfy io.Closer.
type wtCloser struct {
	session *webtransport.Session
}

func (c wtCloser) Close() error {
	return c.session.CloseWithError(0, "")
}

// wtOpener adapts *webtransport.Session to the streamOpener interface.
type wtOpener struct{ session *webtransport.Session }

func (o wtOpener) OpenStream() (io.ReadWriteCloser, error) {
	return o.session.OpenStream()
}

type wtAcceptor struct{ session *webtransport.Session }

func (a wtAcceptor) AcceptStream(ctx context.Context) (io.ReadWriteCloser, error) {
	return a.session.AcceptStream(ctx)
}

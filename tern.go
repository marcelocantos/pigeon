// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Package tern provides client-side connectivity to a tern relay server.
// Backends call Register to obtain an instance ID; clients call Connect
// with a known instance ID. Both return a Conn for bidirectional
// message exchange over QUIC.
//
// By default, Register and Connect use raw QUIC (ALPN "tern") for
// native clients. Use WithWebTransport() for browser-oriented paths
// that require WebTransport (HTTP/3).
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
	"io"
	"net/http"
	"net/url"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/webtransport-go"
)

//go:embed agents-guide.md
var AgentGuide string

// Option configures a relay connection.
type Option func(*options)

type options struct {
	token         string
	instanceID    string // persistent instance ID (empty = server assigns random)
	tlsConfig     *tls.Config
	webTransport  bool
	quicPort      string // override the QUIC port (default: "4433")
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

// WithWebTransport forces WebTransport (HTTP/3) instead of raw QUIC.
// Use this for browser clients or when connecting to a relay that only
// supports WebTransport.
func WithWebTransport() Option {
	return func(o *options) { o.webTransport = true }
}

// WithInstanceID sets a persistent instance ID for registration. If set,
// the relay uses this ID instead of generating a random one. This enables
// persistent pairing — clients that know the ID can reconnect across
// reboots and network changes without re-scanning a QR code.
func WithInstanceID(id string) Option {
	return func(o *options) { o.instanceID = id }
}

// WithQUICPort overrides the default QUIC port (4433) for raw QUIC
// connections. This is the port the relay's QUICServer listens on.
func WithQUICPort(port string) Option {
	return func(o *options) { o.quicPort = port }
}

func buildOptions(opts []Option) options {
	var o options
	for _, fn := range opts {
		fn(&o)
	}
	return o
}

// Register connects to the relay as a backend. By default uses raw QUIC
// (ALPN "tern"). The relay assigns an instance ID, returned via
// InstanceID(). The caller is responsible for closing the connection.
func Register(ctx context.Context, relayURL string, opts ...Option) (*Conn, error) {
	o := buildOptions(opts)

	if o.webTransport {
		return registerWebTransport(ctx, relayURL, o)
	}
	return registerQUIC(ctx, relayURL, o)
}

// Connect connects to a relay as a client targeting a specific backend
// instance ID. By default uses raw QUIC (ALPN "tern").
func Connect(ctx context.Context, relayURL, instanceID string, opts ...Option) (*Conn, error) {
	o := buildOptions(opts)

	if o.webTransport {
		return connectWebTransport(ctx, relayURL, instanceID, o)
	}
	return connectQUIC(ctx, relayURL, instanceID, o)
}

// --- Raw QUIC client ---

func quicTLSConfig(o options) *tls.Config {
	cfg := o.tlsConfig
	if cfg == nil {
		cfg = &tls.Config{}
	} else {
		cfg = cfg.Clone()
	}
	cfg.NextProtos = []string{ternALPN}
	return cfg
}

// quicAddr derives the raw QUIC address from a relay URL. The default
// QUIC port is 4433 unless overridden by WithQUICPort.
func quicAddr(relayURL string, o options) (string, error) {
	u, err := url.Parse(relayURL)
	if err != nil {
		return "", fmt.Errorf("parse relay URL: %w", err)
	}
	host := u.Hostname()
	port := o.quicPort
	if port == "" {
		port = "4433"
	}
	return host + ":" + port, nil
}

func registerQUIC(ctx context.Context, relayURL string, o options) (*Conn, error) {
	addr, err := quicAddr(relayURL, o)
	if err != nil {
		return nil, err
	}

	conn, err := quic.DialAddr(ctx, addr, quicTLSConfig(o), &quic.Config{EnableDatagrams: true})
	if err != nil {
		return nil, fmt.Errorf("register: quic dial: %w", err)
	}

	stream, err := conn.OpenStream()
	if err != nil {
		conn.CloseWithError(0, "failed to open stream")
		return nil, fmt.Errorf("register: open stream: %w", err)
	}

	// Send handshake: "register[:TOKEN[:INSTANCE_ID]]"
	handshake := "register"
	if o.token != "" || o.instanceID != "" {
		handshake = "register:" + o.token + ":" + o.instanceID
	}
	if err := writeMessage(stream, []byte(handshake)); err != nil {
		conn.CloseWithError(0, "failed to send handshake")
		return nil, fmt.Errorf("register: handshake: %w", err)
	}

	// Read the instance ID.
	idBytes, err := readMessage(stream)
	if err != nil {
		conn.CloseWithError(0, "failed to read ID")
		return nil, fmt.Errorf("register: read ID: %w", err)
	}

	closer := quicCloser{conn}
	return newConn(stream, conn, closer, quicOpener{conn}, quicAcceptor{conn}, string(idBytes)), nil
}

func connectQUIC(ctx context.Context, relayURL, instanceID string, o options) (*Conn, error) {
	addr, err := quicAddr(relayURL, o)
	if err != nil {
		return nil, err
	}

	conn, err := quic.DialAddr(ctx, addr, quicTLSConfig(o), &quic.Config{EnableDatagrams: true})
	if err != nil {
		return nil, fmt.Errorf("connect to %s: quic dial: %w", instanceID, err)
	}

	stream, err := conn.OpenStream()
	if err != nil {
		conn.CloseWithError(0, "failed to open stream")
		return nil, fmt.Errorf("connect: open stream: %w", err)
	}

	// Send handshake: "connect:<instanceID>".
	if err := writeMessage(stream, []byte("connect:"+instanceID)); err != nil {
		conn.CloseWithError(0, "failed to send handshake")
		return nil, fmt.Errorf("connect: handshake: %w", err)
	}

	closer := quicCloser{conn}
	return newConn(stream, conn, closer, quicOpener{conn}, quicAcceptor{conn}, instanceID), nil
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

// --- WebTransport client (for browsers / backward compat) ---

func registerWebTransport(ctx context.Context, relayURL string, o options) (*Conn, error) {
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
	handshake := "register"
	if o.instanceID != "" {
		handshake = "register::" + o.instanceID
	}
	if err := writeMessage(stream, []byte(handshake)); err != nil {
		session.CloseWithError(0, "failed to send handshake")
		return nil, fmt.Errorf("register: handshake: %w", err)
	}

	// Read the instance ID.
	idBytes, err := readMessage(stream)
	if err != nil {
		session.CloseWithError(0, "failed to read ID")
		return nil, fmt.Errorf("register: read ID: %w", err)
	}

	closer := wtCloser{session}
	return newConn(stream, session, closer, wtOpener{session}, wtAcceptor{session}, string(idBytes)), nil
}

func connectWebTransport(ctx context.Context, relayURL, instanceID string, o options) (*Conn, error) {
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
	if err := writeMessage(stream, []byte("connect")); err != nil {
		session.CloseWithError(0, "failed to send handshake")
		return nil, fmt.Errorf("connect: handshake: %w", err)
	}

	closer := wtCloser{session}
	return newConn(stream, session, closer, wtOpener{session}, wtAcceptor{session}, instanceID), nil
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

// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Package tern provides client-side connectivity to a tern relay server.
// Backends call Register to obtain an instance ID; clients call Connect
// with a known instance ID. Both return a Conn for bidirectional
// message exchange.
//
// Sub-packages provide E2E encryption (crypto/), protocol state machines
// (protocol/), and QR code rendering (qr/).
package tern

import (
	"context"
	_ "embed"
	"fmt"
	"net/http"

	"github.com/coder/websocket"
)

//go:embed agents-guide.md
var AgentGuide string

// Conn wraps a WebSocket connection to the relay. It provides
// send/receive for both backend and client sides.
type Conn struct {
	ws         *websocket.Conn
	instanceID string // set for backend connections
}

// Option configures a relay connection.
type Option func(*options)

type options struct {
	token    string
	dialOpts *websocket.DialOptions
}

// WithToken sets the bearer token for authentication on /register.
func WithToken(token string) Option {
	return func(o *options) { o.token = token }
}

// WithDialOptions sets custom WebSocket dial options.
func WithDialOptions(opts *websocket.DialOptions) Option {
	return func(o *options) { o.dialOpts = opts }
}

func buildOptions(opts []Option) options {
	var o options
	for _, fn := range opts {
		fn(&o)
	}
	return o
}

func buildDialOpts(o options) *websocket.DialOptions {
	d := o.dialOpts
	if d == nil {
		d = &websocket.DialOptions{}
	}
	if o.token != "" {
		if d.HTTPHeader == nil {
			d.HTTPHeader = http.Header{}
		}
		d.HTTPHeader.Set("Authorization", "Bearer "+o.token)
	}
	return d
}

// Register connects to the relay's /register endpoint as a backend.
// The relay assigns an instance ID, returned via InstanceID().
// The caller is responsible for closing the connection.
func Register(ctx context.Context, relayURL string, opts ...Option) (*Conn, error) {
	o := buildOptions(opts)
	d := buildDialOpts(o)

	ws, _, err := websocket.Dial(ctx, relayURL+"/register", d)
	if err != nil {
		return nil, fmt.Errorf("register: %w", err)
	}

	// The relay sends the instance ID as the first text message.
	_, idBytes, err := ws.Read(ctx)
	if err != nil {
		ws.CloseNow()
		return nil, fmt.Errorf("read instance ID: %w", err)
	}

	return &Conn{ws: ws, instanceID: string(idBytes)}, nil
}

// Connect connects to a relay as a client, targeting a specific
// backend instance ID.
func Connect(ctx context.Context, relayURL, instanceID string, opts ...Option) (*Conn, error) {
	o := buildOptions(opts)
	d := o.dialOpts // no auth needed for client connections

	ws, _, err := websocket.Dial(ctx, relayURL+"/ws/"+instanceID, d)
	if err != nil {
		return nil, fmt.Errorf("connect to %s: %w", instanceID, err)
	}

	return &Conn{ws: ws, instanceID: instanceID}, nil
}

// InstanceID returns the relay-assigned instance ID.
func (c *Conn) InstanceID() string {
	return c.instanceID
}

// Send writes a message to the relay.
func (c *Conn) Send(ctx context.Context, mt websocket.MessageType, data []byte) error {
	return c.ws.Write(ctx, mt, data)
}

// Recv reads a message from the relay.
func (c *Conn) Recv(ctx context.Context) (websocket.MessageType, []byte, error) {
	return c.ws.Read(ctx)
}

// Close gracefully closes the relay connection.
func (c *Conn) Close() error {
	return c.ws.Close(websocket.StatusNormalClosure, "")
}

// CloseNow immediately closes the connection without a close handshake.
func (c *Conn) CloseNow() error {
	return c.ws.CloseNow()
}

// SetReadLimit sets the maximum message size the connection will read.
func (c *Conn) SetReadLimit(n int64) {
	c.ws.SetReadLimit(n)
}

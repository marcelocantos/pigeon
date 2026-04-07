// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package pigeon

import (
	"context"
	"crypto/tls"
	"io"
	"sync"
	"time"

	"github.com/marcelocantos/pigeon/crypto"
)

// datagrammer provides unreliable datagram send/receive.
type datagrammer interface {
	SendDatagram([]byte) error
	ReceiveDatagram(context.Context) ([]byte, error)
}

// streamOpener can open additional bidirectional streams on the underlying
// QUIC connection or WebTransport session.
type streamOpener interface {
	OpenStream() (io.ReadWriteCloser, error)
}

// streamAcceptor can accept incoming bidirectional streams.
type streamAcceptor interface {
	AcceptStream(context.Context) (io.ReadWriteCloser, error)
}

// deadliner can set read/write deadlines on a stream.
type deadliner interface {
	SetReadDeadline(time.Time) error
	SetWriteDeadline(time.Time) error
}

// Conn manages communication with a peer through a QUIC-based session
// (either raw QUIC or WebTransport). The primary bidirectional stream
// is used for reliable ordered messages. Datagrams provide an unreliable
// channel for latency-sensitive data.
//
// In raw mode (before SetChannel), messages pass through unmodified.
// In encrypted mode (after SetChannel), Send/Recv automatically encrypt
// and decrypt using the configured crypto.Channel.
type Conn struct {
	mu         sync.Mutex
	instanceID string

	// The executor mediates all I/O through the session state machine.
	exec *executor

	// Encryption for the primary stream. Nil means raw mode.
	channel *crypto.Channel

	// Encryption for datagrams. Nil means raw datagram mode.
	dgChannel *crypto.Channel

	// PairingRecord for deriving per-channel encryption keys.
	pairingRecord *crypto.PairingRecord

	// Datagram channel cache (keyed by channel ID).
	dgChannels map[uint16]*DatagramChannel

	// LAN upgrade config.
	lanServer  *LANServer
	lanEnabled bool
	lanTLS     *tls.Config

	ctx    context.Context
	cancel context.CancelFunc
}

// connRole indicates whether a Conn is a backend or client.
type connRole int

const (
	roleBackend connRole = iota
	roleClient
)

func newConn(stream io.ReadWriteCloser, dg datagrammer, closer io.Closer, opener streamOpener, acceptor streamAcceptor, instanceID string, role connRole) *Conn {
	ctx, cancel := context.WithCancel(context.Background())
	relay := newPath("relay", stream, dg, closer, opener, acceptor)

	// Create the transport state machine for this role.
	var machine interface {
		HandleEvent(ev EventID) ([]CmdID, error)
	}
	switch role {
	case roleBackend:
		m := NewSessionProtocolBackendMachine()
		m.State = SessionProtocolBackendRelayConnected
		m.Guards[SessionProtocolGuardChallengeValid] = func() bool { return true }
		m.Guards[SessionProtocolGuardChallengeInvalid] = func() bool { return false }
		m.Guards[SessionProtocolGuardLanServerAvailable] = func() bool { return true }
		m.Guards[SessionProtocolGuardUnderMaxFailures] = func() bool { return m.PingFailures+1 < 3 }
		m.Guards[SessionProtocolGuardAtMaxFailures] = func() bool { return m.PingFailures+1 >= 3 }
		m.Actions[SessionProtocolActionActivateLan] = func() error { return nil }
		m.Actions[SessionProtocolActionResetFailures] = func() error { return nil }
		m.Actions[SessionProtocolActionFallbackToRelay] = func() error {
			// Increment backoff (complex expr not handled by codegen).
			m.BackoffLevel++
			if m.BackoffLevel > 5 {
				m.BackoffLevel = 5
			}
			return nil
		}
		machine = m
	case roleClient:
		m := NewSessionProtocolClientMachine()
		m.State = SessionProtocolClientRelayConnected
		m.Guards[SessionProtocolGuardLanEnabled] = func() bool { return true }
		m.Guards[SessionProtocolGuardLanDisabled] = func() bool { return false }
		m.Actions[SessionProtocolActionDialLan] = func() error { return nil }
		m.Actions[SessionProtocolActionActivateLan] = func() error { return nil }
		m.Actions[SessionProtocolActionFallbackToRelay] = func() error { return nil }
		machine = m
	}

	exec := newExecutor(ctx, cancel, machine, relay)
	exec.instanceID = instanceID

	return &Conn{
		instanceID: instanceID,
		exec:       exec,
		ctx:        ctx,
		cancel:     cancel,
	}
}

// SetPairingRecord stores a PairingRecord for deriving per-channel
// encryption keys. Call this after loading a saved pairing record
// and before opening/accepting channels.
func (c *Conn) SetPairingRecord(rec *crypto.PairingRecord) {
	c.mu.Lock()
	c.pairingRecord = rec
	c.mu.Unlock()
}

// active returns the active path from the executor.
func (c *Conn) active() *path { return c.exec.activePath() }

// acceptStream accepts the next incoming bidirectional stream from the peer.
func (c *Conn) acceptStream(ctx context.Context) (io.ReadWriteCloser, error) {
	p := c.active()
	if p.acceptor == nil {
		return nil, io.ErrClosedPipe
	}
	return p.acceptor.AcceptStream(ctx)
}

// OpenStream opens a new bidirectional stream on the underlying QUIC
// connection or WebTransport session. The returned stream implements
// io.ReadWriteCloser; use writeMessage/readMessage for length-prefixed framing.
//
// NOTE: The relay server currently bridges only the primary stream.
// Additional streams opened via OpenStream are not forwarded to the peer.
// TODO: Add multi-stream relay support in session.go (bridgeClient).
func (c *Conn) OpenStream() (io.ReadWriteCloser, error) {
	return c.active().opener.OpenStream()
}

// InstanceID returns the relay-assigned instance ID.
func (c *Conn) InstanceID() string {
	return c.instanceID
}

// LANReady returns a channel that is closed when the LAN transport
// is active. Use this to wait for the LAN upgrade to complete:
//
//	select {
//	case <-conn.LANReady():
//	    // LAN is active
//	case <-ctx.Done():
//	    // timed out, still on relay
//	}
func (c *Conn) LANReady() <-chan struct{} {
	return c.exec.lanReady
}

// SetChannel enables encrypted mode on the primary stream. After this
// call, Send encrypts plaintext and Recv decrypts ciphertext automatically.
//
// If a LANServer was configured (via WithLANServer), SetChannel also
// advertises the LAN address to the peer.
func (c *Conn) SetChannel(ch *crypto.Channel) {
	c.mu.Lock()
	c.channel = ch
	c.mu.Unlock()

	// Configure the executor's encryption.
	c.exec.channel = ch

	// Trigger LAN advertisement via the machine.
	if c.lanServer != nil {
		c.exec.submit(event{id: SessionProtocolEventLanServerReady})
	}
}

// SetDatagramChannel enables encrypted mode for datagrams. After this
// call, SendDatagram encrypts and RecvDatagram decrypts automatically.
func (c *Conn) SetDatagramChannel(ch *crypto.Channel) {
	c.mu.Lock()
	c.dgChannel = ch
	c.mu.Unlock()

	c.exec.dgChannel = ch
}

// Send writes a message on the primary bidirectional stream. In raw mode,
// data is sent as-is with length-prefix framing. In encrypted mode, data
// is treated as plaintext and encrypted before sending.
//
// Send is safe for concurrent use from multiple goroutines. The executor
// serializes all writes through the event loop.
func (c *Conn) Send(ctx context.Context, data []byte) error {
	return c.exec.send(ctx, data)
}

// Recv reads the next message from the primary bidirectional stream.
// In raw mode, returns the raw bytes. In encrypted mode, decrypts and
// returns the application payload. Control messages (LAN offers, etc.)
// are handled by the executor's stream reader and never delivered here.
func (c *Conn) Recv(ctx context.Context) ([]byte, error) {
	return c.exec.recv(ctx)
}

// SendDatagram sends an unreliable datagram to the peer. The executor
// handles encryption, framing, and fragmentation.
func (c *Conn) SendDatagram(data []byte) error {
	return c.exec.sendDatagram(data)
}

// RecvDatagram receives the next datagram from the peer. The executor
// handles reassembly, decryption, and channel demux.
func (c *Conn) RecvDatagram(ctx context.Context) ([]byte, error) {
	return c.exec.recvDatagram(ctx)
}

// Close gracefully closes the session.
func (c *Conn) Close() error {
	c.cancel()
	if c.exec.lan != nil {
		c.exec.lan.close()
	}
	relay := c.exec.relay
	if relay.stream != nil {
		relay.stream.Close()
	}
	return relay.closer.Close()
}

// fallbackToRelay forces an immediate fallback from the direct path
// to relay. The machine handles this via app_force_fallback transitions
// from any LAN-related state.
func (c *Conn) fallbackToRelay() {
	c.exec.submitSync(event{id: SessionProtocolEventAppForceFallback})
}

// isDirectActive returns true if the direct path is active.
func (c *Conn) isDirectActive() bool {
	return c.exec.lan != nil
}

// CloseNow immediately closes the session.
func (c *Conn) CloseNow() error {
	c.cancel()
	if c.exec.lan != nil {
		c.exec.lan.close()
	}
	return c.exec.relay.closer.Close()
}

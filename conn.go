// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package tern

import (
	"context"
	"crypto/tls"
	"encoding/binary"
	"io"
	"sync"
	"time"

	"github.com/marcelocantos/tern/crypto"
)

// Internal message types — first byte of encrypted plaintext.
// These are invisible to callers; only application messages are delivered.
// Retained for future use (LAN upgrade over WebTransport).
const (
	msgApp      byte = 0x00 // application message
	msgLANOffer byte = 0x01 // LAN address exchange (reserved)
	msgCutover  byte = 0x02 // transport cutover marker (reserved)
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

	// Datagram channel dispatch (legacy — migrating to executor).
	dgChannels  map[uint16]*DatagramChannel
	dgIncoming  map[uint16]chan []byte // per-channel datagram queues
	connDg      chan []byte            // conn-level datagram queue
	dgDispatch  sync.Once             // starts dispatcher once

	// Fragment reassembly (legacy — executor has its own).
	reasm        *reassembler
	maxDgPayload int

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
		m := NewBackendMachine()
		m.State = BackendRelayConnected
		m.Guards[GuardChallengeValid] = func() bool { return true }
		m.Guards[GuardChallengeInvalid] = func() bool { return false }
		m.Guards[GuardLanServerAvailable] = func() bool { return true }
		m.Guards[GuardUnderMaxFailures] = func() bool { return m.PingFailures+1 < 3 }
		m.Guards[GuardAtMaxFailures] = func() bool { return m.PingFailures+1 >= 3 }
		m.Actions[ActionActivateLan] = func() error { return nil }
		m.Actions[ActionResetFailures] = func() error { return nil }
		m.Actions[ActionFallbackToRelay] = func() error { return nil }
		machine = m
	case roleClient:
		m := NewClientMachine()
		m.State = ClientRelayConnected
		m.Guards[GuardLanEnabled] = func() bool { return true }
		m.Guards[GuardLanDisabled] = func() bool { return false }
		m.Actions[ActionDialLan] = func() error { return nil }
		m.Actions[ActionActivateLan] = func() error { return nil }
		m.Actions[ActionFallbackToRelay] = func() error { return nil }
		machine = m
	}

	exec := newExecutor(ctx, cancel, machine, relay)
	exec.instanceID = instanceID

	// Legacy reassembler for old code paths still using Conn directly.
	done := make(chan struct{})
	go func() {
		<-ctx.Done()
		close(done)
	}()

	return &Conn{
		instanceID:   instanceID,
		exec:         exec,
		reasm:        newReassembler(DefaultFragmentTimeout, done),
		maxDgPayload: DefaultMaxDatagramPayload,
		ctx:          ctx,
		cancel:       cancel,
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

// datagramDispatcher reads datagrams from the QUIC connection and
// routes them: conn-level datagrams go to the connDg queue, channel-
// tagged datagrams go to the appropriate channel queue. Handles both
// whole and fragmented datagrams.
func (c *Conn) datagramDispatcher() {
	for {
		payload, chanID, err := c.recvRawDatagram(c.ctx)
		if err != nil {
			return
		}
		if chanID >= 0 {
			c.routeToChannel(uint16(chanID), payload)
		} else {
			select {
			case c.connDg <- payload:
			default:
			}
		}
	}
}

// recvTaggedDatagram receives the next datagram for a specific channel ID.
func (c *Conn) recvTaggedDatagram(ctx context.Context, id uint16) ([]byte, error) {
	c.mu.Lock()
	ch, ok := c.dgIncoming[id]
	if !ok {
		if c.dgIncoming == nil {
			c.dgIncoming = make(map[uint16]chan []byte)
		}
		ch = make(chan []byte, 64)
		c.dgIncoming[id] = ch
	}
	c.mu.Unlock()

	select {
	case data := <-ch:
		return data, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-c.ctx.Done():
		return nil, c.ctx.Err()
	}
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
		c.exec.submit(event{id: EventLanServerReady})
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

func (c *Conn) decryptDatagram(payload []byte) ([]byte, error) {
	c.mu.Lock()
	ch := c.dgChannel
	c.mu.Unlock()
	if ch != nil {
		return ch.Decrypt(payload)
	}
	return payload, nil
}

func (c *Conn) ensureDispatcher() {
	c.dgDispatch.Do(func() {
		c.connDg = make(chan []byte, 64)
		go c.datagramDispatcher()
	})
}

// recvRawDatagram reads one logical datagram from the QUIC layer,
// handling framing and fragment reassembly. Returns the payload and
// the channel ID (-1 for conn-level datagrams).
func (c *Conn) recvRawDatagram(ctx context.Context) (payload []byte, chanID int, err error) {
	for {
		data, err := c.active().dg.ReceiveDatagram(ctx)
		if err != nil {
			return nil, -1, err
		}
		if len(data) == 0 {
			continue
		}

		switch data[0] {
		case dgConnWhole:
			return data[1:], -1, nil

		case dgConnFragment:
			if len(data) < 1+fragHeaderSize {
				continue
			}
			msgID := binary.BigEndian.Uint32(data[1:5])
			fragIdx := int(binary.BigEndian.Uint16(data[5:7]))
			totalFrags := int(binary.BigEndian.Uint16(data[7:9]))
			chunk := data[1+fragHeaderSize:]
			if totalFrags < 2 || fragIdx >= totalFrags {
				continue
			}
			assembled := c.reasm.feed(msgID, fragIdx, totalFrags, chunk)
			if assembled == nil {
				continue
			}
			return assembled, -1, nil

		case dgChanWhole:
			if len(data) < 1+chanIDSize {
				continue
			}
			id := int(binary.BigEndian.Uint16(data[1:3]))
			return data[1+chanIDSize:], id, nil

		case dgChanFragment:
			if len(data) < 1+chanIDSize+fragHeaderSize {
				continue
			}
			id := int(binary.BigEndian.Uint16(data[1:3]))
			off := 1 + chanIDSize
			msgID := binary.BigEndian.Uint32(data[off : off+4])
			fragIdx := int(binary.BigEndian.Uint16(data[off+4 : off+6]))
			totalFrags := int(binary.BigEndian.Uint16(data[off+6 : off+8]))
			chunk := data[off+fragHeaderSize:]
			if totalFrags < 2 || fragIdx >= totalFrags {
				continue
			}
			assembled := c.reasm.feed(msgID, fragIdx, totalFrags, chunk)
			if assembled == nil {
				continue
			}
			return assembled, id, nil

		default:
			continue
		}
	}
}

// routeToChannel delivers a datagram payload to the named channel's queue.
func (c *Conn) routeToChannel(id uint16, payload []byte) {
	c.mu.Lock()
	ch, ok := c.dgIncoming[id]
	if !ok {
		if c.dgIncoming == nil {
			c.dgIncoming = make(map[uint16]chan []byte)
		}
		ch = make(chan []byte, 64)
		c.dgIncoming[id] = ch
	}
	c.mu.Unlock()

	select {
	case ch <- payload:
	default:
		// Drop if full (datagram semantics).
	}
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

// fallbackToRelay forces a fallback from the direct path to relay.
// Drives the session machine through the appropriate transitions.
func (c *Conn) fallbackToRelay() {
	switch m := c.exec.machine.(type) {
	case *BackendMachine:
		switch m.State {
		case BackendLANActive:
			m.HandleEvent(EventPingTimeout) // → LANDegraded
			m.PingFailures = 2
			m.HandleEvent(EventPingTimeout) // → RelayBackoff
		case BackendLANDegraded:
			m.PingFailures = 2
			m.HandleEvent(EventPingTimeout)
		case BackendLANOffered:
			m.HandleEvent(EventOfferTimeout)
		}
	case *ClientMachine:
		switch m.State {
		case ClientLANActive:
			m.HandleEvent(EventLanError)
			m.HandleEvent(EventRelayOk)
		}
	}
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

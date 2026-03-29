// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package tern

import (
	"context"
	"encoding/binary"
	"io"
	"log/slog"
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
	mu      sync.Mutex
	writeMu sync.Mutex // serialises writes to the primary stream
	instanceID string

	stream   io.ReadWriteCloser // primary bidirectional stream (quic.Stream or webtransport.Stream)
	dg       datagrammer        // datagram interface (quic.Connection or webtransport.Session)
	closer   io.Closer          // the session/connection itself
	opener   streamOpener       // opens additional bidirectional streams
	acceptor streamAcceptor     // accepts incoming streams from peer

	// Encryption for the primary stream. Nil means raw mode.
	channel *crypto.Channel

	// Encryption for datagrams. Nil means raw datagram mode.
	dgChannel *crypto.Channel

	// PairingRecord for deriving per-channel encryption keys.
	pairingRecord *crypto.PairingRecord

	// Datagram channel dispatch.
	dgChannels map[uint16]*DatagramChannel
	dgIncoming map[uint16]chan []byte // per-channel datagram queues

	ctx    context.Context    // Conn lifecycle context
	cancel context.CancelFunc // cancels background goroutines
}

func newConn(stream io.ReadWriteCloser, dg datagrammer, closer io.Closer, opener streamOpener, acceptor streamAcceptor, instanceID string) *Conn {
	ctx, cancel := context.WithCancel(context.Background())
	return &Conn{
		instanceID: instanceID,
		stream:     stream,
		dg:         dg,
		closer:     closer,
		opener:     opener,
		acceptor:   acceptor,
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

// acceptStream accepts the next incoming bidirectional stream from the peer.
func (c *Conn) acceptStream(ctx context.Context) (io.ReadWriteCloser, error) {
	if c.acceptor == nil {
		return nil, io.ErrClosedPipe
	}
	return c.acceptor.AcceptStream(ctx)
}

// datagramDispatcher reads datagrams from the QUIC connection and
// routes them to the appropriate DatagramChannel based on the 2-byte
// channel ID prefix.
func (c *Conn) datagramDispatcher() {
	for {
		data, err := c.dg.ReceiveDatagram(c.ctx)
		if err != nil {
			return
		}
		if len(data) < 2 {
			continue
		}
		id := binary.BigEndian.Uint16(data[:2])
		payload := data[2:]

		c.mu.Lock()
		ch, ok := c.dgIncoming[id]
		if !ok {
			// Lazily create the incoming queue.
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
			// Drop if queue is full (datagram semantics — loss is expected).
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
	return c.opener.OpenStream()
}

// InstanceID returns the relay-assigned instance ID.
func (c *Conn) InstanceID() string {
	return c.instanceID
}

// SetChannel enables encrypted mode on the primary stream. After this
// call, Send encrypts plaintext and Recv decrypts ciphertext automatically.
func (c *Conn) SetChannel(ch *crypto.Channel) {
	c.mu.Lock()
	c.channel = ch
	c.mu.Unlock()
}

// SetDatagramChannel enables encrypted mode for datagrams. After this
// call, SendDatagram encrypts and RecvDatagram decrypts automatically.
func (c *Conn) SetDatagramChannel(ch *crypto.Channel) {
	c.mu.Lock()
	c.dgChannel = ch
	c.mu.Unlock()
}

// Send writes a message on the primary bidirectional stream. In raw mode,
// data is sent as-is with length-prefix framing. In encrypted mode, data
// is treated as plaintext and encrypted before sending.
//
// Send is safe for concurrent use from multiple goroutines.
//
// If ctx carries a deadline, it is applied to the underlying stream write
// via SetWriteDeadline. Cancellation without a deadline is not supported
// (the underlying stream write is not interruptible).
func (c *Conn) Send(ctx context.Context, data []byte) error {
	c.mu.Lock()
	ch := c.channel
	c.mu.Unlock()

	payload := data
	if ch != nil {
		framed := make([]byte, 1+len(data))
		framed[0] = msgApp
		copy(framed[1:], data)
		payload = ch.Encrypt(framed)
	}

	// Serialise writes: writeMessage performs two Write calls
	// (length header + payload) which must not interleave.
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	if deadline, ok := ctx.Deadline(); ok {
		if dl, ok := c.stream.(deadliner); ok {
			dl.SetWriteDeadline(deadline)
			defer dl.SetWriteDeadline(time.Time{})
		}
	}

	return writeMessage(c.stream, payload)
}

// Recv reads the next message from the primary bidirectional stream.
// In raw mode, returns the raw bytes. In encrypted mode, decrypts and
// returns the application payload, silently discarding control messages.
//
// If ctx carries a deadline, it is applied to the underlying stream read
// via SetReadDeadline. Cancellation without a deadline is not supported
// (the underlying stream read is not interruptible).
func (c *Conn) Recv(ctx context.Context) ([]byte, error) {
	if deadline, ok := ctx.Deadline(); ok {
		if dl, ok := c.stream.(deadliner); ok {
			dl.SetReadDeadline(deadline)
			defer dl.SetReadDeadline(time.Time{})
		}
	}

	for {
		c.mu.Lock()
		ch := c.channel
		c.mu.Unlock()

		data, err := readMessage(c.stream)
		if err != nil {
			return nil, err
		}

		if ch == nil {
			return data, nil
		}

		plaintext, err := ch.Decrypt(data)
		if err != nil {
			return nil, err
		}

		if len(plaintext) == 0 {
			return nil, nil
		}

		switch plaintext[0] {
		case msgApp:
			return plaintext[1:], nil
		case msgLANOffer, msgCutover:
			slog.Debug("discarding control message", "type", plaintext[0])
			continue
		default:
			slog.Warn("discarding unknown message type", "type", plaintext[0])
			continue
		}
	}
}

// SendDatagram sends an unreliable datagram to the peer via the
// QUIC session. In raw mode, data is sent as-is. In encrypted
// mode (after SetDatagramChannel), data is encrypted before sending.
func (c *Conn) SendDatagram(data []byte) error {
	c.mu.Lock()
	ch := c.dgChannel
	c.mu.Unlock()

	payload := data
	if ch != nil {
		payload = ch.Encrypt(data)
	}
	return c.dg.SendDatagram(payload)
}

// RecvDatagram receives the next datagram from the peer. In raw mode,
// returns the raw bytes. In encrypted mode (after SetDatagramChannel),
// decrypts the datagram before returning.
func (c *Conn) RecvDatagram(ctx context.Context) ([]byte, error) {
	data, err := c.dg.ReceiveDatagram(ctx)
	if err != nil {
		return nil, err
	}

	c.mu.Lock()
	ch := c.dgChannel
	c.mu.Unlock()

	if ch == nil {
		return data, nil
	}

	return ch.Decrypt(data)
}

// Close gracefully closes the session. It closes the primary
// bidirectional stream first (allowing buffered data to flush) before
// closing the session.
func (c *Conn) Close() error {
	c.cancel()
	if c.stream != nil {
		c.stream.Close()
	}
	return c.closer.Close()
}

// CloseNow immediately closes the session.
func (c *Conn) CloseNow() error {
	c.cancel()
	return c.closer.Close()
}

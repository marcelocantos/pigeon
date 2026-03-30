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
	dgChannels  map[uint16]*DatagramChannel
	dgIncoming  map[uint16]chan []byte // per-channel datagram queues
	connDg      chan []byte            // conn-level datagram queue (when dispatcher is running)
	dgDispatch  sync.Once             // starts dispatcher once

	// Fragment reassembly for large datagrams.
	reasm        *reassembler
	maxDgPayload int // max bytes per raw datagram (default 1200)

	ctx    context.Context    // Conn lifecycle context
	cancel context.CancelFunc // cancels background goroutines
}

func newConn(stream io.ReadWriteCloser, dg datagrammer, closer io.Closer, opener streamOpener, acceptor streamAcceptor, instanceID string) *Conn {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	c := &Conn{
		instanceID:   instanceID,
		stream:       stream,
		dg:           dg,
		closer:       closer,
		opener:       opener,
		acceptor:     acceptor,
		reasm:        newReassembler(DefaultFragmentTimeout, done),
		maxDgPayload: DefaultMaxDatagramPayload,
		ctx:          ctx,
		cancel:       cancel,
	}
	// Close the reassembler when the conn context is cancelled.
	go func() {
		<-ctx.Done()
		close(done)
	}()
	return c
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

// SendDatagram sends an unreliable datagram to the peer. Payloads that
// fit in a single QUIC datagram are sent as-is (with a 1-byte framing
// prefix). Larger payloads are automatically split into fragments and
// reassembled by the receiver. If any fragment is lost, the entire
// message is discarded after a timeout (datagram semantics).
//
// In encrypted mode (after SetDatagramChannel), data is encrypted
// before framing/fragmentation.
func (c *Conn) SendDatagram(data []byte) error {
	c.mu.Lock()
	ch := c.dgChannel
	c.mu.Unlock()

	payload := data
	if ch != nil {
		payload = ch.Encrypt(data)
	}

	// Does it fit in a single datagram? (1 byte prefix + payload)
	if 1+len(payload) <= c.maxDgPayload {
		frame := make([]byte, 1+len(payload))
		frame[0] = dgConnWhole
		copy(frame[1:], payload)
		return c.dg.SendDatagram(frame)
	}

	// Fragment it.
	msgID := nextMsgID.Add(1)
	return sendFragmented(c.dg, payload, c.maxDgPayload, msgID, dgConnFragment, nil)
}

// RecvDatagram receives the next datagram from the peer. Fragmented
// datagrams are reassembled transparently. If reassembly of a message
// times out (missing fragments), it is silently discarded and the next
// complete message is returned.
//
// Channel-tagged datagrams (from DatagramChannel) are routed to the
// appropriate channel queue and not returned here.
//
// In encrypted mode (after SetDatagramChannel), the datagram is
// decrypted after reassembly.
func (c *Conn) RecvDatagram(ctx context.Context) ([]byte, error) {
	c.mu.Lock()
	hasChannels := len(c.dgChannels) > 0
	c.mu.Unlock()

	if hasChannels {
		// Dispatcher is running — read from the conn-level queue.
		c.ensureDispatcher()
		select {
		case payload := <-c.connDg:
			return c.decryptDatagram(payload)
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-c.ctx.Done():
			return nil, c.ctx.Err()
		}
	}

	// No channels — read directly.
	for {
		payload, chanID, err := c.recvRawDatagram(ctx)
		if err != nil {
			return nil, err
		}
		if chanID >= 0 {
			c.routeToChannel(uint16(chanID), payload)
			continue
		}
		return c.decryptDatagram(payload)
	}
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
		data, err := c.dg.ReceiveDatagram(ctx)
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

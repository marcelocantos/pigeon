// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package tern

import (
	"context"
	"encoding/binary"
	"io"
	"sync"

	"github.com/marcelocantos/tern/crypto"
)

// StreamChannel is a named, encrypted bidirectional stream.
// Each StreamChannel has its own QUIC stream and crypto.Channel,
// providing independent ordering and encryption from other channels.
type StreamChannel struct {
	name    string
	stream  io.ReadWriteCloser
	ch      *crypto.Channel
	writeMu sync.Mutex
}

// Name returns the channel name.
func (sc *StreamChannel) Name() string { return sc.name }

// Send writes an encrypted message on this channel.
func (sc *StreamChannel) Send(ctx context.Context, data []byte) error {
	payload := data
	if sc.ch != nil {
		payload = sc.ch.Encrypt(data)
	}
	sc.writeMu.Lock()
	defer sc.writeMu.Unlock()
	return writeMessage(sc.stream, payload)
}

// Recv reads and decrypts the next message on this channel.
func (sc *StreamChannel) Recv(ctx context.Context) ([]byte, error) {
	data, err := readMessage(sc.stream)
	if err != nil {
		return nil, err
	}
	if sc.ch != nil {
		return sc.ch.Decrypt(data)
	}
	return data, nil
}

// Close closes the underlying stream.
func (sc *StreamChannel) Close() error {
	return sc.stream.Close()
}

// OpenChannel opens a named streaming channel to the peer. A new QUIC
// stream is created, the channel name is sent as the first message, and
// encryption keys are derived from the connection's master key pair
// (if set via SetChannel) using the channel name as the HKDF info.
//
// The peer receives this channel via AcceptChannel.
func (c *Conn) OpenChannel(name string) (*StreamChannel, error) {
	stream, err := c.opener.OpenStream()
	if err != nil {
		return nil, err
	}

	// Send the channel name as the first message on the stream.
	if err := writeMessage(stream, []byte(name)); err != nil {
		stream.Close()
		return nil, err
	}

	// Derive per-channel encryption if a master channel is set.
	var ch *crypto.Channel
	c.mu.Lock()
	masterCh := c.channel
	c.mu.Unlock()
	if masterCh != nil {
		ch, err = c.deriveChannelKeys(name, true)
		if err != nil {
			stream.Close()
			return nil, err
		}
	}

	return &StreamChannel{name: name, stream: stream, ch: ch}, nil
}

// AcceptChannel waits for the peer to open a named streaming channel.
// Reads the channel name from the new QUIC stream and derives matching
// encryption keys.
func (c *Conn) AcceptChannel(ctx context.Context) (*StreamChannel, error) {
	stream, err := c.acceptStream(ctx)
	if err != nil {
		return nil, err
	}

	// Read the channel name.
	nameBytes, err := readMessage(stream)
	if err != nil {
		stream.Close()
		return nil, err
	}
	name := string(nameBytes)

	// Derive per-channel encryption if a master channel is set.
	var ch *crypto.Channel
	c.mu.Lock()
	masterCh := c.channel
	c.mu.Unlock()
	if masterCh != nil {
		ch, err = c.deriveChannelKeys(name, false)
		if err != nil {
			stream.Close()
			return nil, err
		}
	}

	return &StreamChannel{name: name, stream: stream, ch: ch}, nil
}

// deriveChannelKeys derives a crypto.Channel for a named channel from
// the PairingRecord or stored keys. The isOpener flag determines the
// send/recv key direction (opener sends on "name:o2a", accepts on "name:a2o").
func (c *Conn) deriveChannelKeys(name string, isOpener bool) (*crypto.Channel, error) {
	c.mu.Lock()
	rec := c.pairingRecord
	c.mu.Unlock()

	if rec == nil {
		return nil, nil
	}

	sendInfo := []byte(name + ":o2a")
	recvInfo := []byte(name + ":a2o")
	if !isOpener {
		sendInfo, recvInfo = recvInfo, sendInfo
	}

	return rec.DeriveChannel(sendInfo, recvInfo)
}

// DatagramChannel provides a named, encrypted datagram sub-channel.
// All datagram channels share the single QUIC datagram pipe but are
// demuxed by a 2-byte channel ID prefix. Each has its own crypto.Channel.
type DatagramChannel struct {
	id   uint16
	conn *Conn
	ch   *crypto.Channel
}

// Send sends an encrypted datagram on this channel.
func (dc *DatagramChannel) Send(data []byte) error {
	payload := data
	if dc.ch != nil {
		payload = dc.ch.Encrypt(data)
	}
	// Prefix with 2-byte channel ID.
	frame := make([]byte, 2+len(payload))
	binary.BigEndian.PutUint16(frame[:2], dc.id)
	copy(frame[2:], payload)
	return dc.conn.dg.SendDatagram(frame)
}

// Recv receives the next datagram on this channel. Blocks until a
// datagram with this channel's ID arrives.
func (dc *DatagramChannel) Recv(ctx context.Context) ([]byte, error) {
	for {
		data, err := dc.conn.recvTaggedDatagram(ctx, dc.id)
		if err != nil {
			return nil, err
		}
		if dc.ch != nil {
			return dc.ch.Decrypt(data)
		}
		return data, nil
	}
}

// DatagramChannel creates or returns a named datagram channel. Both
// sides must create the channel with the same name. The channel ID is
// derived deterministically from the name (CRC16), and encryption keys
// are derived from the master key pair using the channel name.
//
// Unlike streaming channels, there is no open/accept — both sides
// create the channel by name and it works immediately.
func (c *Conn) DatagramChannel(name string) *DatagramChannel {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.dgChannels == nil {
		c.dgChannels = make(map[uint16]*DatagramChannel)
		// Start the datagram dispatcher if not already running.
		go c.datagramDispatcher()
	}

	id := channelID(name)
	if dc, ok := c.dgChannels[id]; ok {
		return dc
	}

	var ch *crypto.Channel
	if c.pairingRecord != nil {
		sendInfo := []byte(name + ":dg:o2a")
		recvInfo := []byte(name + ":dg:a2o")
		// For datagram channels, both sides use the same direction keys
		// because there's no opener/accepter distinction. Use sorted
		// name-based derivation instead.
		sendInfo = []byte(name + ":dg:send")
		recvInfo = []byte(name + ":dg:recv")
		ch, _ = c.pairingRecord.DeriveChannel(sendInfo, recvInfo)
		if ch != nil {
			ch.SetMode(crypto.ModeDatagrams)
		}
	}

	dc := &DatagramChannel{id: id, conn: c, ch: ch}
	c.dgChannels[id] = dc
	return dc
}

// channelID derives a deterministic 16-bit channel ID from a name.
func channelID(name string) uint16 {
	// Simple hash: sum of bytes mod 65536. Good enough for channel demux.
	var h uint32
	for _, b := range []byte(name) {
		h = h*31 + uint32(b)
	}
	return uint16(h)
}

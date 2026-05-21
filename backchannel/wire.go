// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package backchannel

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

// MsgType names the five typed pairing-ceremony events the CLI and
// daemon exchange. Values are strings so they survive JSON round-trip
// unchanged.
type MsgType string

const (
	// MsgPairBegin — CLI → daemon. "Start a new pairing ceremony."
	MsgPairBegin MsgType = "pair_begin"
	// MsgTokenResponse — daemon → CLI. Carries the relay-issued
	// instance ID + opaque rendezvous token the CLI should render
	// (typically as a QR code) for the mobile peer to scan.
	MsgTokenResponse MsgType = "token_response"
	// MsgWaitingForCode — daemon → CLI. Mobile peer has connected and
	// the ECDH-derived confirmation code is now displayable on both
	// sides. The CLI prompts the user.
	MsgWaitingForCode MsgType = "waiting_for_code"
	// MsgCodeSubmit — CLI → daemon. The local user accepted the
	// confirmation code.
	MsgCodeSubmit MsgType = "code_submit"
	// MsgPairStatus — daemon → CLI. Terminal status: "paired",
	// "aborted", or an error string. Triggers CLI exit.
	MsgPairStatus MsgType = "pair_status"
)

// Message is the single typed union used on the wire. Empty fields are
// omitted; absence is part of the contract for each Type.
type Message struct {
	Type       MsgType `json:"type"`
	InstanceID string  `json:"instance_id,omitempty"`
	Token      string  `json:"token,omitempty"`
	Code       string  `json:"code,omitempty"`
	Status     string  `json:"status,omitempty"`
}

// MaxFrameSize caps a single backchannel JSON payload. 64 KiB is far
// more than any of the five messages need (the largest is a base64
// rendezvous token, ~hundreds of bytes), and bounding the prefix
// protects against a malicious or buggy peer streaming an unbounded
// length. Mirrors the defensive-coding guidance for length-prefixed
// protocols.
const MaxFrameSize = 64 * 1024

// LengthPrefixSize is the framing header width in bytes. Matches the
// relay-stream framing (uint32 big-endian).
const LengthPrefixSize = 4

// ErrOversize is returned when a frame's length prefix exceeds
// MaxFrameSize. Wraps any underlying decode/write context.
var ErrOversize = errors.New("backchannel: frame exceeds MaxFrameSize")

// writeFrame serialises one Message and writes the length-prefixed JSON
// payload. The caller is responsible for serialising concurrent writes
// to the same io.Writer; a Unix domain socket allows parallel read +
// write but does not safely interleave parallel writes.
func writeFrame(w io.Writer, msg Message) error {
	if msg.Type == "" {
		return errors.New("backchannel: message has empty Type")
	}
	payload, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("backchannel marshal: %w", err)
	}
	if len(payload) > MaxFrameSize {
		return fmt.Errorf("%w: %d > %d", ErrOversize, len(payload), MaxFrameSize)
	}
	var hdr [LengthPrefixSize]byte
	binary.BigEndian.PutUint32(hdr[:], uint32(len(payload)))
	// One Write per record so a single short write doesn't split a
	// frame across syscalls on the receiver side. net.UnixConn honours
	// this contract for stream sockets.
	buf := make([]byte, 0, len(hdr)+len(payload))
	buf = append(buf, hdr[:]...)
	buf = append(buf, payload...)
	if _, err := w.Write(buf); err != nil {
		return fmt.Errorf("backchannel write: %w", err)
	}
	return nil
}

// readFrame reads one length-prefixed JSON payload and decodes it.
// Unknown Type values are returned to the caller (which can ignore or
// reject); JSON parse failures and oversize prefixes are errors.
func readFrame(r io.Reader) (Message, error) {
	var hdr [LengthPrefixSize]byte
	if _, err := io.ReadFull(r, hdr[:]); err != nil {
		return Message{}, err
	}
	length := binary.BigEndian.Uint32(hdr[:])
	if length > MaxFrameSize {
		return Message{}, fmt.Errorf("%w: %d > %d", ErrOversize, length, MaxFrameSize)
	}
	buf := make([]byte, length)
	if _, err := io.ReadFull(r, buf); err != nil {
		return Message{}, fmt.Errorf("backchannel read payload: %w", err)
	}
	var msg Message
	if err := json.Unmarshal(buf, &msg); err != nil {
		return Message{}, fmt.Errorf("backchannel decode: %w", err)
	}
	if msg.Type == "" {
		return Message{}, errors.New("backchannel: message has empty Type")
	}
	return msg, nil
}

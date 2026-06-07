// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Code generated from protocol/wireformats.yaml. DO NOT EDIT.

package pigeon

import (
	"encoding/binary"
	"errors"
	"strings"
)

var (
	_ = binary.BigEndian
	_ = errors.New
	_ = strings.HasPrefix
)

// EncodeStreamHeader — First message on every stream — binds the stream to a name.
func EncodeStreamHeader(name string) []byte {
	var buf []byte
	{
		var tmp [binary.MaxVarintLen64]byte
		n := binary.PutUvarint(tmp[:], uint64(len(name)))
		buf = append(buf, tmp[:n]...)
		buf = append(buf, []byte(name)...)
	}
	return buf
}

// DecodeStreamHeader — First message on every stream — binds the stream to a name.
func DecodeStreamHeader(buf []byte) (name string, err error) {
	off := 0
	{
		ln, n := binary.Uvarint(buf[off:])
		if n <= 0 {
			err = errors.New("bad varint length name")
			return
		}
		off += n
		if uint64(len(buf)-off) < ln {
			err = errors.New("truncated name")
			return
		}
		name = string(buf[off : off+int(ln)])
		off += int(ln)
	}
	_ = off
	return
}

// EncodeDatagramPlaintext — Plaintext layout of a datagram before AEAD encryption.
func EncodeDatagramPlaintext(channelId uint64, payload []byte) []byte {
	var buf []byte
	{
		var tmp [binary.MaxVarintLen64]byte
		n := binary.PutUvarint(tmp[:], channelId)
		buf = append(buf, tmp[:n]...)
	}
	buf = append(buf, payload...)
	return buf
}

// DecodeDatagramPlaintext — Plaintext layout of a datagram before AEAD encryption.
func DecodeDatagramPlaintext(buf []byte) (channelId uint64, payload []byte, err error) {
	off := 0
	{
		v, n := binary.Uvarint(buf[off:])
		if n <= 0 {
			err = errors.New("bad varint channel_id")
			return
		}
		channelId = v
		off += n
	}
	payload = append([]byte(nil), buf[off:]...)
	off = len(buf)
	_ = off
	return
}

// RelayGreetingVariant identifies which union variant a decoded
// relay_greeting message belongs to.
type RelayGreetingVariant int

const (
	RelayGreetingConnect  RelayGreetingVariant = 0
	RelayGreetingRegister RelayGreetingVariant = 1
	RelayGreetingListen   RelayGreetingVariant = 2
)

// RelayGreetingDecoded is the result of decoding a relay_greeting.
type RelayGreetingDecoded struct {
	Variant    RelayGreetingVariant
	InstanceId string
	Token      string
}

// EncodeRelayGreetingConnect — Relay-side greeting sent on the primary stream after dial.
func EncodeRelayGreetingConnect(instanceId string) []byte {
	var sb strings.Builder
	sb.WriteString("connect:")
	sb.WriteString(instanceId)
	return []byte(sb.String())
}

// EncodeRelayGreetingRegister — Relay-side greeting sent on the primary stream after dial.
func EncodeRelayGreetingRegister(token string, instanceId string) []byte {
	var sb strings.Builder
	sb.WriteString("register")
	suffix := []string{token, instanceId}
	anyNonEmpty := false
	for _, p := range suffix {
		if p != "" {
			anyNonEmpty = true
		}
	}
	if anyNonEmpty {
		for i := range suffix {
			sb.WriteByte(':')
			sb.WriteString(suffix[i])
		}
	}
	return []byte(sb.String())
}

// EncodeRelayGreetingListen — Relay-side greeting sent on the primary stream after dial.
func EncodeRelayGreetingListen(token string, instanceId string) []byte {
	var sb strings.Builder
	sb.WriteString("listen")
	suffix := []string{token, instanceId}
	anyNonEmpty := false
	for _, p := range suffix {
		if p != "" {
			anyNonEmpty = true
		}
	}
	if anyNonEmpty {
		for i := range suffix {
			sb.WriteByte(':')
			sb.WriteString(suffix[i])
		}
	}
	return []byte(sb.String())
}

// DecodeRelayGreeting parses a relay_greeting into a tagged union.
func DecodeRelayGreeting(buf []byte) (RelayGreetingDecoded, error) {
	var out RelayGreetingDecoded
	s := string(buf)
	if strings.HasPrefix(s, "connect:") {
		out.Variant = RelayGreetingConnect
		rest := s[len("connect:"):]
		out.InstanceId = rest
		return out, nil
	}
	if strings.HasPrefix(s, "register") {
		out.Variant = RelayGreetingRegister
		rest := s[len("register"):]
		if rest != "" {
			if rest[0] != ':' {
				return out, errors.New("relay_greeting: malformed register: missing ':' after prefix")
			}
			rest = rest[1:]
			parts := strings.SplitN(rest, ":", 2)
			if len(parts) > 0 {
				out.Token = parts[0]
			}
			if len(parts) > 1 {
				out.InstanceId = parts[1]
			}
		}
		return out, nil
	}
	if strings.HasPrefix(s, "listen") {
		out.Variant = RelayGreetingListen
		rest := s[len("listen"):]
		if rest != "" {
			if rest[0] != ':' {
				return out, errors.New("relay_greeting: malformed listen: missing ':' after prefix")
			}
			rest = rest[1:]
			parts := strings.SplitN(rest, ":", 2)
			if len(parts) > 0 {
				out.Token = parts[0]
			}
			if len(parts) > 1 {
				out.InstanceId = parts[1]
			}
		}
		return out, nil
	}
	return out, errors.New("relay_greeting: unrecognised prefix")
}

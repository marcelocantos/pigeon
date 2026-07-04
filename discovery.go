// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package pigeon

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
)

// discoveryStreamName is the reserved sub-stream a client opens to enumerate a
// node's routes (🎯T44.3). The NUL prefix keeps it out of the application name
// space (apps use plain names like "chat"). When a backend Session has a
// Discover hook, its accept loop answers this stream internally — calling
// Discover and replying — instead of delivering it to the application.
const discoveryStreamName = "\x00pigeon.enumerate"

// RouteEntry is one advertised service under a multi-service node (🎯T44.3).
// Route is the sub-address a client passes as ConnectArgs.Route to reach the
// service; Metadata is opaque application-defined bytes (service name, player
// count, ...) that pigeon never interprets — the app defines its shape.
type RouteEntry struct {
	Route    string
	Metadata []byte
}

// Enumerate asks the peer node for the routes visible to this client and
// returns a snapshot (🎯T44.3). It opens the reserved, AEAD-encrypted
// discovery stream, sends an enumerate request, and decodes the reply. The
// peer must have been configured with RegisterArgs.Discover; otherwise the
// request is never answered and this blocks until ctx expires.
//
// Discovery is a snapshot, not a subscription: call Enumerate again to
// refresh. The relay never sees the exchange (it rides an encrypted stream).
func (s *Session) Enumerate(ctx context.Context) ([]RouteEntry, error) {
	st, err := s.OpenStream(ctx, discoveryStreamName)
	if err != nil {
		return nil, fmt.Errorf("enumerate: open: %w", err)
	}
	defer st.Close()
	// The request carries no fields today (an optional filter blob could go
	// here later); a single byte marks a well-formed request.
	if err := st.Send([]byte{0x01}); err != nil {
		return nil, fmt.Errorf("enumerate: send: %w", err)
	}
	resp, err := st.Recv(ctx)
	if err != nil {
		return nil, fmt.Errorf("enumerate: recv: %w", err)
	}
	return decodeRoutes(resp)
}

// handleEnumerate answers one discovery request on the backend side. It runs
// (in its own goroutine) when acceptLoop sees the reserved discovery stream
// and s.discover is set.
func (s *Session) handleEnumerate(st *Stream) {
	defer st.Close()
	if _, err := st.Recv(s.ctx); err != nil {
		return
	}
	entries, err := s.discover(s.peerID)
	if err != nil {
		// On a hook error, return an empty set rather than leak the reason.
		entries = nil
	}
	_ = st.Send(encodeRoutes(entries))
}

// encodeRoutes wire format:
//
//	[uvarint count]{ [uvarint route-len][route][uvarint meta-len][meta] }*
func encodeRoutes(entries []RouteEntry) []byte {
	buf := binary.AppendUvarint(nil, uint64(len(entries)))
	for _, e := range entries {
		buf = binary.AppendUvarint(buf, uint64(len(e.Route)))
		buf = append(buf, e.Route...)
		buf = binary.AppendUvarint(buf, uint64(len(e.Metadata)))
		buf = append(buf, e.Metadata...)
	}
	return buf
}

func decodeRoutes(b []byte) ([]RouteEntry, error) {
	count, n := binary.Uvarint(b)
	if n <= 0 {
		return nil, errors.New("enumerate: bad count")
	}
	b = b[n:]
	entries := make([]RouteEntry, 0, count)
	for i := uint64(0); i < count; i++ {
		route, rest, err := readLenPrefixed(b)
		if err != nil {
			return nil, fmt.Errorf("enumerate: route %d: %w", i, err)
		}
		meta, rest, err := readLenPrefixed(rest)
		if err != nil {
			return nil, fmt.Errorf("enumerate: metadata %d: %w", i, err)
		}
		metaCopy := make([]byte, len(meta))
		copy(metaCopy, meta)
		entries = append(entries, RouteEntry{Route: string(route), Metadata: metaCopy})
		b = rest
	}
	return entries, nil
}

// readLenPrefixed reads a [uvarint len][bytes] field, returning the bytes
// (aliasing b) and the remaining buffer.
func readLenPrefixed(b []byte) (val, rest []byte, err error) {
	l, n := binary.Uvarint(b)
	if n <= 0 {
		return nil, nil, errors.New("bad length")
	}
	b = b[n:]
	if uint64(len(b)) < l {
		return nil, nil, errors.New("truncated")
	}
	return b[:l], b[l:], nil
}

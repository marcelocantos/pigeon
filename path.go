// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package tern

import (
	"io"
	"time"
)

// path represents a single transport path (relay, LAN, STUN, etc.).
// It bundles the stream, datagram, and connection management interfaces
// needed to send and receive over that path.
type path struct {
	name     string             // "relay", "lan", etc.
	stream   io.ReadWriteCloser // primary bidirectional stream
	dg       datagrammer        // datagram interface
	closer   io.Closer          // the session/connection itself
	opener   streamOpener       // opens additional bidirectional streams
	acceptor streamAcceptor     // accepts incoming streams from peer

	// Health monitoring.
	healthy   bool
	lastSend  time.Time
	lastRecv  time.Time
	failures  int
}

func newPath(name string, stream io.ReadWriteCloser, dg datagrammer, closer io.Closer, opener streamOpener, acceptor streamAcceptor) *path {
	now := time.Now()
	return &path{
		name:     name,
		stream:   stream,
		dg:       dg,
		closer:   closer,
		opener:   opener,
		acceptor: acceptor,
		healthy:  true,
		lastSend: now,
		lastRecv: now,
	}
}

func (p *path) close() {
	if p.stream != nil {
		p.stream.Close()
	}
	if p.closer != nil {
		p.closer.Close()
	}
}

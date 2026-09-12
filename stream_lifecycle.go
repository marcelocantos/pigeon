// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package pigeon

import (
	"errors"
	"io"
	"log/slog"
	"sync/atomic"
	"time"
)

// Default L4 stream-open policy (🎯T63.1 / 🎯T63.2).
const (
	// DefaultStreamRendezvous is how long an unmatched incoming stream
	// is parked for a local AcceptStream before it is reset.
	DefaultStreamRendezvous = 2 * time.Second

	// DefaultMaxBufferedStreams is the per-name cap on parked incoming
	// streams. Excess OpenStream from the peer is reset immediately.
	DefaultMaxBufferedStreams = 1
)

// goAwayStreamName is the reserved control stream that carries a
// session go-away. It is not surfaced to AcceptStream.
const goAwayStreamName = "\x00pigeon.goaway"

// ErrSessionDraining is returned by OpenStream after GoAway.
var ErrSessionDraining = errors.New("pigeon: session is draining")

// streamWaiter is one AcceptStream rendezvous slot. done is set by
// either the waiter (cancel) or deliverIncomingStream so a cancelled
// AcceptStream cannot leak the delivered rwc (Fable F10).
type streamWaiter struct {
	ch   chan io.ReadWriteCloser
	done atomic.Bool
}

func reservedStreamName(name string) bool {
	return name == discoveryStreamName || name == lanControlStreamName || name == goAwayStreamName
}

func (s *Session) applyStreamPolicy(names []string) {
	if len(names) == 0 {
		return
	}
	allowed := make(map[string]struct{}, len(names))
	for _, n := range names {
		if n != "" {
			allowed[n] = struct{}{}
		}
	}
	s.streamMu.Lock()
	s.allowedStreams = allowed
	s.streamMu.Unlock()
}

func (s *Session) rendezvousTimeout() time.Duration {
	if s.rendezvousWindow <= 0 {
		return DefaultStreamRendezvous
	}
	return s.rendezvousWindow
}

func (s *Session) bufferedCap() int {
	if s.maxBufferedPerName <= 0 {
		return DefaultMaxBufferedStreams
	}
	return s.maxBufferedPerName
}

func (s *Session) isDraining() bool {
	s.streamMu.Lock()
	defer s.streamMu.Unlock()
	return s.draining
}

func nameAllowed(allowed map[string]struct{}, name string) bool {
	if reservedStreamName(name) {
		return true
	}
	if allowed == nil {
		return true
	}
	_, ok := allowed[name]
	return ok
}

func (s *Session) expireBuffered(name string, rwc io.ReadWriteCloser) {
	s.streamMu.Lock()
	q := s.bufferedStreams[name]
	for i, r := range q {
		if r == rwc {
			s.bufferedStreams[name] = append(q[:i], q[i+1:]...)
			if len(s.bufferedStreams[name]) == 0 {
				delete(s.bufferedStreams, name)
			}
			s.streamMu.Unlock()
			_ = resetStream(rwc)
			return
		}
	}
	s.streamMu.Unlock()
}

func (s *Session) rejectParkedStreams() {
	s.streamMu.Lock()
	buf := s.bufferedStreams
	s.bufferedStreams = make(map[string][]io.ReadWriteCloser)
	s.streamMu.Unlock()
	for _, q := range buf {
		for _, rwc := range q {
			_ = resetStream(rwc)
		}
	}
}

func (s *Session) enterDrain() {
	s.streamMu.Lock()
	if s.draining {
		s.streamMu.Unlock()
		return
	}
	s.draining = true
	s.streamMu.Unlock()
	s.rejectParkedStreams()
}

// GoAway enters a draining state: local OpenStream fails (except
// reserved control names), new peer-opened application streams are
// reset, and in-flight streams are left alone so the application can
// finish them. Session.Close remains the abort path. A go-away is
// signalled to the peer on the reserved control stream.
func (s *Session) GoAway() error {
	if s.isDraining() {
		return nil
	}
	s.enterDrain()
	go s.sendGoAway()
	return nil
}

func (s *Session) sendGoAway() {
	tr := s.currentTransport()
	if tr == nil {
		return
	}
	rwc, err := tr.OpenStream()
	if err != nil {
		slog.Debug("session: go-away open", "err", err)
		return
	}
	if err := writeMessage(rwc, EncodeStreamHeader(goAwayStreamName)); err != nil {
		_ = resetStream(rwc)
		slog.Debug("session: go-away header", "err", err)
		return
	}
	_ = closeWrite(rwc)
}

func (s *Session) bufferedStreamCount() int {
	s.streamMu.Lock()
	defer s.streamMu.Unlock()
	n := 0
	for _, q := range s.bufferedStreams {
		n += len(q)
	}
	return n
}

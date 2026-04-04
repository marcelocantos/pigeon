// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package tern

import (
	"context"
	"io"
	"log/slog"
	"math"
	"math/rand/v2"
	"sync"
	"time"
)

// Backoff configuration for re-establishment attempts.
const (
	backoffBaseInterval = 1 * time.Second  // first retry after 1s
	maxBackoffLevel     = 5                // cap: 2^5 = 32x base = 32s
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

// pathRouter manages multiple paths and routes traffic through the
// best available one. The relay path is permanent; direct paths
// (LAN, STUN) are optional and come and go.
//
// The routing rule is simple: use the direct path if it's healthy,
// otherwise fall back to relay. When the direct path fails, close it
// and re-advertise (triggering a new LAN/STUN attempt).
type pathRouter struct {
	mu     sync.Mutex
	relay  *path   // permanent — never nil after init
	direct *path   // optional — LAN, STUN, etc.
	active *path   // points to either relay or direct

	// Backoff state for re-establishment.
	backoffLevel int

	// Callback when the active path changes.
	onSwitch func(from, to string)
}

func newPathRouter(relay *path) *pathRouter {
	return &pathRouter{
		relay:  relay,
		active: relay,
	}
}

// activePath returns the current active path under the lock.
func (r *pathRouter) activePath() *path {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.active
}

// setDirect installs a direct path and switches to it. Resets backoff.
func (r *pathRouter) setDirect(p *path) {
	r.mu.Lock()
	old := r.active
	r.direct = p
	r.active = p
	r.backoffLevel = 0 // success resets backoff
	r.mu.Unlock()

	slog.Info("path switched", "from", old.name, "to", p.name)
	if r.onSwitch != nil {
		r.onSwitch(old.name, p.name)
	}
}

// fallbackToRelay closes the direct path, reverts to relay, and
// increments the backoff level (capped at maxBackoffLevel).
func (r *pathRouter) fallbackToRelay() {
	r.mu.Lock()
	direct := r.direct
	if direct == nil {
		r.mu.Unlock()
		return
	}
	old := r.active
	r.direct = nil
	r.active = r.relay
	r.backoffLevel++
	if r.backoffLevel > maxBackoffLevel {
		r.backoffLevel = maxBackoffLevel
	}
	level := r.backoffLevel
	r.mu.Unlock()

	slog.Info("path fallback", "from", old.name, "to", "relay",
		"backoff_level", level)
	direct.close()
	if r.onSwitch != nil {
		r.onSwitch(old.name, "relay")
	}
}

// backoffDelay returns the current backoff delay with ±25% jitter.
func (r *pathRouter) backoffDelay() time.Duration {
	r.mu.Lock()
	level := r.backoffLevel
	r.mu.Unlock()

	if level <= 0 {
		return 0
	}

	base := backoffBaseInterval * time.Duration(math.Pow(2, float64(level-1)))
	jitter := time.Duration(rand.Int64N(int64(base) / 2)) - base/4
	return base + jitter
}

// hasDirect returns true if a direct path is currently installed.
func (r *pathRouter) hasDirect() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.direct != nil
}

// isDirectActive returns true if the direct path is the active one.
func (r *pathRouter) isDirectActive() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.direct != nil && r.active == r.direct
}

// monitor watches the active path's health and triggers failover.
// It sends periodic pings on the active path and falls back to relay
// if pings fail.
func (r *pathRouter) monitor(ctx context.Context, pingFn func(*path) error) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	consecutiveFailures := 0
	const maxFailures = 3

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}

		active := r.activePath()
		if active == r.relay {
			// On relay — nothing to monitor (relay is the fallback).
			consecutiveFailures = 0
			continue
		}

		// Ping the direct path.
		if err := pingFn(active); err != nil {
			consecutiveFailures++
			slog.Debug("direct path ping failed",
				"path", active.name,
				"failures", consecutiveFailures,
				"err", err)

			if consecutiveFailures >= maxFailures {
				slog.Warn("direct path unhealthy, falling back to relay",
					"path", active.name,
					"failures", consecutiveFailures)
				r.fallbackToRelay()
				consecutiveFailures = 0
			}
		} else {
			consecutiveFailures = 0
		}
	}
}

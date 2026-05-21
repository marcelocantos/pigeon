// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package backchannel

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"
)

// Session is one accepted CLI ↔ daemon connection (daemon side).
// Lifetime is the duration of a single pairing ceremony — typically
// seconds. Both Send and Recv are context-aware; Close releases the
// underlying socket and unblocks anything in flight.
type Session struct {
	conn net.Conn

	writeMu sync.Mutex // serialises writeFrame on conn
	readMu  sync.Mutex // serialises readFrame on conn (one logical reader)

	closeOnce sync.Once
	closeErr  error
}

// Client is the CLI-side endpoint. Identical wire contract to Session
// — it's a separate type only to express the asymmetric naming the
// design doc reaches for. They share the same connection mechanics.
type Client = Session

// newSession wraps an accepted (server) or dialled (client) net.Conn.
// Internal — callers obtain a *Session from Server.Accept or a *Client
// (alias) from Dial.
func newSession(conn net.Conn) *Session {
	return &Session{conn: conn}
}

// DialArgs configures the CLI-side connect.
type DialArgs struct {
	// AppName addresses the well-known SocketPath. Required when Path
	// is empty.
	AppName string

	// Path overrides SocketPath lookup. Useful for tests.
	Path string

	// Timeout bounds the connect itself (the kernel returns ENOENT
	// immediately when the daemon isn't running, so the only thing
	// this guards is a slow filesystem). Zero ⇒ no timeout.
	Timeout time.Duration
}

// Dial connects to a running daemon's backchannel socket.
func Dial(args *DialArgs) (*Client, error) {
	if args == nil {
		return nil, errors.New("backchannel.Dial: args is nil")
	}
	path := args.Path
	if path == "" {
		p, err := SocketPath(args.AppName)
		if err != nil {
			return nil, err
		}
		path = p
	} else if err := validateAppName(args.AppName); err != nil {
		return nil, err
	}

	var conn net.Conn
	var err error
	if args.Timeout > 0 {
		conn, err = net.DialTimeout("unix", path, args.Timeout)
	} else {
		conn, err = net.Dial("unix", path)
	}
	if err != nil {
		return nil, fmt.Errorf("backchannel dial %q: %w", path, err)
	}
	return newSession(conn), nil
}

// Send writes one Message frame. Honours ctx by setting a write
// deadline before the call; on ctx.Done() the write returns a deadline
// error which we map back to ctx.Err().
func (s *Session) Send(ctx context.Context, msg Message) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	if err := applyDeadline(ctx, s.conn.SetWriteDeadline); err != nil {
		return err
	}
	defer func() { _ = s.conn.SetWriteDeadline(time.Time{}) }()
	if err := writeFrame(s.conn, msg); err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}
		return err
	}
	return nil
}

// Recv reads one Message frame. Honours ctx the same way Send does.
func (s *Session) Recv(ctx context.Context) (Message, error) {
	s.readMu.Lock()
	defer s.readMu.Unlock()
	if err := applyDeadline(ctx, s.conn.SetReadDeadline); err != nil {
		return Message{}, err
	}
	defer func() { _ = s.conn.SetReadDeadline(time.Time{}) }()
	msg, err := readFrame(s.conn)
	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return Message{}, ctxErr
		}
		return Message{}, err
	}
	return msg, nil
}

// Close closes the underlying socket. Idempotent. Safe to call from
// any goroutine — unblocks any in-flight Send/Recv with an error.
func (s *Session) Close() error {
	s.closeOnce.Do(func() {
		s.closeErr = s.conn.Close()
	})
	return s.closeErr
}

// applyDeadline plumbs ctx into a net.Conn deadline setter. A nil
// ctx.Deadline() (the common case — Background or ceremony-scoped
// cancel) installs no deadline; cancellation is then handled by Close
// from a goroutine the caller may spawn, but in practice the ceremony
// already has a context-aware close path. When ctx has a deadline we
// pin the socket to it so the syscall returns when the user's overall
// timeout fires.
func applyDeadline(ctx context.Context, set func(time.Time) error) error {
	if dl, ok := ctx.Deadline(); ok {
		if err := set(dl); err != nil {
			return fmt.Errorf("backchannel deadline: %w", err)
		}
	} else if err := set(time.Time{}); err != nil {
		return fmt.Errorf("backchannel deadline reset: %w", err)
	}
	return nil
}

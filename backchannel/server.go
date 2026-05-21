// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package backchannel

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

// ListenArgs configures the daemon-side socket listener.
type ListenArgs struct {
	// AppName disambiguates concurrent daemons on one host. The socket
	// path is derived from this — see SocketPath. Required.
	AppName string

	// Path overrides the well-known SocketPath(AppName) lookup. Useful
	// for tests (TempDir) and for unconventional installations. When
	// non-empty AppName is still required for sanity logging but the
	// host filesystem layout is irrelevant.
	Path string
}

// Server is the daemon-side Unix-socket listener. Each Accept yields a
// fresh *Session bound to a connected CLI.
type Server struct {
	path     string
	listener *net.UnixListener

	mu     sync.Mutex
	closed bool
}

// Listen creates the Unix-socket directory (mode 0700), unlinks any
// stale socket from a previously-crashed daemon, and starts listening.
// Close releases the listener and removes the socket file.
func Listen(args *ListenArgs) (*Server, error) {
	if args == nil {
		return nil, errors.New("backchannel.Listen: args is nil")
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

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("backchannel: create socket dir %q: %w", filepath.Dir(path), err)
	}
	// Best-effort tighten in case the dir pre-existed with looser perms
	// (umask, prior daemon version).
	if err := os.Chmod(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("backchannel: chmod socket dir %q: %w", filepath.Dir(path), err)
	}

	if err := clearStaleSocket(path); err != nil {
		return nil, err
	}

	// Set umask while binding so the resulting socket inode is 0600;
	// some kernels and Go versions ignore the explicit os.Chmod-after
	// path because the inode is created during ListenUnix.
	addr := &net.UnixAddr{Name: path, Net: "unix"}
	var listener *net.UnixListener
	if err := withUmask(0o077, func() error {
		l, err := net.ListenUnix("unix", addr)
		if err != nil {
			return err
		}
		listener = l
		return nil
	}); err != nil {
		return nil, fmt.Errorf("backchannel: listen %q: %w", path, err)
	}
	// Belt-and-braces: explicit chmod in case the umask path didn't
	// take (e.g. macOS sandbox edge cases). Errors here are non-fatal
	// since the umask already made the inode 0600; warn-and-continue
	// would require a logger we don't have, so we propagate.
	if err := os.Chmod(path, 0o600); err != nil {
		_ = listener.Close()
		_ = os.Remove(path)
		return nil, fmt.Errorf("backchannel: chmod socket %q: %w", path, err)
	}

	return &Server{path: path, listener: listener}, nil
}

// clearStaleSocket removes a socket file left over from a crashed prior
// daemon, distinguishing "stale" (no one is listening) from "live"
// (another daemon owns it). The probe is a non-blocking Dial; if the
// kernel says ECONNREFUSED the socket inode exists but has no listener
// — safe to unlink. Any successful dial means another instance is live
// and we refuse to bind.
func clearStaleSocket(path string) error {
	info, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("backchannel: stat %q: %w", path, err)
	}
	if info.Mode()&os.ModeSocket == 0 {
		return fmt.Errorf("backchannel: path %q exists and is not a socket", path)
	}
	conn, err := net.Dial("unix", path)
	if err == nil {
		_ = conn.Close()
		return fmt.Errorf("backchannel: another daemon is already listening on %q", path)
	}
	// ECONNREFUSED ⇒ inode without a listener; safe to unlink.
	// Anything else (permission, file gone in race) — surface as-is.
	var sysErr *net.OpError
	if errors.As(err, &sysErr) {
		var errno syscall.Errno
		if errors.As(sysErr.Err, &errno) && (errno == syscall.ECONNREFUSED || errno == syscall.ENOENT) {
			if rmErr := os.Remove(path); rmErr != nil && !errors.Is(rmErr, os.ErrNotExist) {
				return fmt.Errorf("backchannel: unlink stale socket %q: %w", path, rmErr)
			}
			return nil
		}
	}
	return fmt.Errorf("backchannel: probe stale socket %q: %w", path, err)
}

// Path returns the absolute filesystem path of the listening socket.
func (s *Server) Path() string { return s.path }

// Accept blocks until the next CLI connects. The returned *Session
// owns the connection — defer-close it from the handler goroutine.
//
// Accept honours ctx cancellation by closing the listener under the
// hood; callers running an accept loop should also call Close on
// shutdown to ensure the socket file is removed.
func (s *Server) Accept(ctx context.Context) (*Session, error) {
	// Hand cancellation to the listener: a goroutine closes the
	// listener when ctx fires, which unblocks AcceptUnix with an
	// "use of closed network connection" error. We translate that
	// back to ctx.Err() for the caller.
	stopCancel := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			_ = s.listener.SetDeadline(deadlineEpoch())
		case <-stopCancel:
		}
	}()
	conn, err := s.listener.AcceptUnix()
	close(stopCancel)
	if err != nil {
		if ctxErr := ctx.Err(); ctxErr != nil {
			return nil, ctxErr
		}
		s.mu.Lock()
		closed := s.closed
		s.mu.Unlock()
		if closed {
			return nil, net.ErrClosed
		}
		return nil, fmt.Errorf("backchannel accept: %w", err)
	}
	// Reset the deadline so subsequent Accepts aren't poisoned.
	_ = s.listener.SetDeadline(time.Time{})
	return newSession(conn), nil
}

// Close stops the listener and removes the socket file. Idempotent.
// Safe to call from a goroutine other than the one running Accept.
func (s *Server) Close() error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	s.closed = true
	listener := s.listener
	path := s.path
	s.mu.Unlock()

	err := listener.Close()
	if rmErr := os.Remove(path); rmErr != nil && !errors.Is(rmErr, os.ErrNotExist) && err == nil {
		err = fmt.Errorf("backchannel: remove socket %q: %w", path, rmErr)
	}
	return err
}

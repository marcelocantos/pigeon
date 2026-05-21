// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package backchannel_test

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/marcelocantos/pigeon/backchannel"
)

// withTempServer spins up a Server listening at a short-path socket.
// macOS limits Unix-socket paths to 104 bytes (sun_path), so t.TempDir
// — which embeds the full test name — usually overflows. We mkdir a
// short directory under os.TempDir instead, with t.Cleanup teardown.
func withTempServer(t *testing.T) (*backchannel.Server, string) {
	t.Helper()
	dir, err := os.MkdirTemp("", "bc-")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	path := filepath.Join(dir, "t.sock")
	srv, err := backchannel.Listen(&backchannel.ListenArgs{AppName: "test", Path: path})
	if err != nil {
		t.Fatalf("Listen: %v", err)
	}
	t.Cleanup(func() { _ = srv.Close() })
	return srv, path
}

// shortTempDir returns a short-path tempdir for socket creation.
// Equivalent to os.MkdirTemp(""; cleanup via t.Cleanup).
func shortTempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "bc-")
	if err != nil {
		t.Fatalf("MkdirTemp: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	return dir
}

func TestPathSanity(t *testing.T) {
	t.Parallel()
	if _, err := backchannel.SocketPath(""); err == nil {
		t.Fatal("expected error for empty AppName")
	}
	if _, err := backchannel.SocketPath("Has Spaces"); err == nil {
		t.Fatal("expected error for AppName with spaces")
	}
	if _, err := backchannel.SocketPath("../escape"); err == nil {
		t.Fatal("expected error for path-traversing AppName")
	}
	if _, err := backchannel.SocketPath(".hidden"); err == nil {
		t.Fatal("expected error for leading-dot AppName")
	}
	p, err := backchannel.SocketPath("pairdroid")
	if err != nil {
		t.Fatalf("happy-path SocketPath: %v", err)
	}
	if !filepath.IsAbs(p) {
		t.Fatalf("SocketPath should return an absolute path, got %q", p)
	}
}

func TestSocketPermissions(t *testing.T) {
	t.Parallel()
	_, path := withTempServer(t)

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("stat socket: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("socket perm = %#o, want 0600", perm)
	}
	dirInfo, err := os.Stat(filepath.Dir(path))
	if err != nil {
		t.Fatalf("stat socket dir: %v", err)
	}
	if perm := dirInfo.Mode().Perm(); perm != 0o700 {
		t.Errorf("dir perm = %#o, want 0700", perm)
	}
}

func TestStaleSocketUnlink(t *testing.T) {
	t.Parallel()
	dir := shortTempDir(t)
	path := filepath.Join(dir, "s.sock")

	// Synthesize a stale socket inode by binding a raw Unix socket
	// through the syscall path (which does NOT register the
	// unlink-on-close handler that net.ListenUnix installs) and
	// closing only the fd. The inode survives, so the next
	// backchannel.Listen must clean it up before binding.
	fd, err := syscall.Socket(syscall.AF_UNIX, syscall.SOCK_STREAM, 0)
	if err != nil {
		t.Fatalf("socket: %v", err)
	}
	sa := &syscall.SockaddrUnix{Name: path}
	if err := syscall.Bind(fd, sa); err != nil {
		_ = syscall.Close(fd)
		t.Fatalf("bind stale: %v", err)
	}
	_ = syscall.Close(fd)
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("stale socket vanished before test: %v", err)
	}

	// Listen should detect the stale inode (no one is accept()ing),
	// unlink it, and rebind.
	srv, err := backchannel.Listen(&backchannel.ListenArgs{AppName: "stale", Path: path})
	if err != nil {
		t.Fatalf("Listen against stale socket: %v", err)
	}
	defer srv.Close()
}

func TestRefuseWhenLiveDaemon(t *testing.T) {
	t.Parallel()
	first, path := withTempServer(t)
	_ = first
	_, err := backchannel.Listen(&backchannel.ListenArgs{AppName: "test", Path: path})
	if err == nil {
		t.Fatal("expected Listen to refuse when another daemon is live")
	}
}

func TestRoundTrip(t *testing.T) {
	t.Parallel()
	srv, path := withTempServer(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Daemon side: accept one session, echo each inbound back with the
	// Token field reversed so we can prove direction-asymmetry too.
	srvDone := make(chan error, 1)
	go func() {
		sess, err := srv.Accept(ctx)
		if err != nil {
			srvDone <- fmt.Errorf("accept: %w", err)
			return
		}
		defer sess.Close()
		for {
			msg, err := sess.Recv(ctx)
			if err != nil {
				srvDone <- nil
				return
			}
			if err := sess.Send(ctx, backchannel.Message{
				Type:       backchannel.MsgTokenResponse,
				InstanceID: "id-" + msg.InstanceID,
				Token:      reverse(msg.Token),
			}); err != nil {
				srvDone <- fmt.Errorf("send: %w", err)
				return
			}
		}
	}()

	// CLI side: dial, send pair_begin twice with different tokens, read
	// each reply, then close.
	cli, err := backchannel.Dial(&backchannel.DialArgs{AppName: "test", Path: path})
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}

	probes := []backchannel.Message{
		{Type: backchannel.MsgPairBegin, InstanceID: "1", Token: "abc"},
		{Type: backchannel.MsgCodeSubmit, InstanceID: "2", Token: "xyzzy", Code: "482901"},
	}
	for _, probe := range probes {
		if err := cli.Send(ctx, probe); err != nil {
			t.Fatalf("client send: %v", err)
		}
		got, err := cli.Recv(ctx)
		if err != nil {
			t.Fatalf("client recv: %v", err)
		}
		want := backchannel.Message{
			Type:       backchannel.MsgTokenResponse,
			InstanceID: "id-" + probe.InstanceID,
			Token:      reverse(probe.Token),
		}
		if got != want {
			t.Errorf("round-trip got %+v, want %+v", got, want)
		}
	}
	_ = cli.Close()

	select {
	case err := <-srvDone:
		if err != nil {
			t.Fatalf("server: %v", err)
		}
	case <-ctx.Done():
		t.Fatal("server goroutine never exited")
	}
}

func TestServeSessionDispatch(t *testing.T) {
	t.Parallel()
	srv, path := withTempServer(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Daemon side: run ServeSession with an EventSink that records
	// every event the executor would see.
	type seen struct {
		name string
		msg  backchannel.Message
	}
	var (
		mu     sync.Mutex
		events []seen
	)
	srvDone := make(chan error, 1)
	go func() {
		sess, err := srv.Accept(ctx)
		if err != nil {
			srvDone <- err
			return
		}
		defer sess.Close()
		err = backchannel.ServeSession(ctx, sess, func(_ context.Context, name string, msg backchannel.Message) error {
			mu.Lock()
			events = append(events, seen{name: name, msg: msg})
			mu.Unlock()
			return nil
		})
		// io.EOF / closed-connection is expected on client-side Close.
		srvDone <- ignoreNetClose(err)
	}()

	cli, err := backchannel.Dial(&backchannel.DialArgs{AppName: "test", Path: path})
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	if err := cli.Send(ctx, backchannel.Message{Type: backchannel.MsgPairBegin}); err != nil {
		t.Fatalf("send pair_begin: %v", err)
	}
	if err := cli.Send(ctx, backchannel.Message{Type: backchannel.MsgCodeSubmit, Code: "123456"}); err != nil {
		t.Fatalf("send code_submit: %v", err)
	}
	_ = cli.Close()

	select {
	case err := <-srvDone:
		if err != nil {
			t.Fatalf("ServeSession: %v", err)
		}
	case <-ctx.Done():
		t.Fatal("ServeSession never returned")
	}

	mu.Lock()
	defer mu.Unlock()
	if len(events) != 2 {
		t.Fatalf("got %d events, want 2: %+v", len(events), events)
	}
	if events[0].name != backchannel.EventCliInitPair {
		t.Errorf("event[0] name = %q, want %q", events[0].name, backchannel.EventCliInitPair)
	}
	if events[1].name != backchannel.EventCliCodeEntered {
		t.Errorf("event[1] name = %q, want %q", events[1].name, backchannel.EventCliCodeEntered)
	}
	if events[1].msg.Code != "123456" {
		t.Errorf("event[1] payload code = %q, want %q", events[1].msg.Code, "123456")
	}
}

func TestServeSessionRejectsUnknownType(t *testing.T) {
	t.Parallel()
	srv, path := withTempServer(t)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	srvDone := make(chan error, 1)
	go func() {
		sess, err := srv.Accept(ctx)
		if err != nil {
			srvDone <- err
			return
		}
		defer sess.Close()
		srvDone <- backchannel.ServeSession(ctx, sess, func(context.Context, string, backchannel.Message) error {
			return nil
		})
	}()

	cli, err := backchannel.Dial(&backchannel.DialArgs{AppName: "test", Path: path})
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}
	defer cli.Close()

	// CLI must not originate token_response; ServeSession rejects.
	if err := cli.Send(ctx, backchannel.Message{Type: backchannel.MsgTokenResponse, Token: "naughty"}); err != nil {
		t.Fatalf("send: %v", err)
	}

	select {
	case err := <-srvDone:
		if err == nil {
			t.Fatal("expected ServeSession to error on disallowed inbound MsgType")
		}
	case <-ctx.Done():
		t.Fatal("ServeSession did not return")
	}
}

func TestRejectsOversizeFrame(t *testing.T) {
	t.Parallel()
	srv, path := withTempServer(t)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	srvDone := make(chan error, 1)
	go func() {
		sess, err := srv.Accept(ctx)
		if err != nil {
			srvDone <- err
			return
		}
		defer sess.Close()
		_, err = sess.Recv(ctx)
		srvDone <- err
	}()

	// Speak the wire directly so we can fabricate a too-large length
	// prefix without going through writeFrame's local guard.
	raw, err := net.Dial("unix", path)
	if err != nil {
		t.Fatalf("raw dial: %v", err)
	}
	defer raw.Close()
	var hdr [4]byte
	binary.BigEndian.PutUint32(hdr[:], backchannel.MaxFrameSize+1)
	if _, err := raw.Write(hdr[:]); err != nil {
		t.Fatalf("write hdr: %v", err)
	}

	select {
	case err := <-srvDone:
		if !errors.Is(err, backchannel.ErrOversize) {
			t.Fatalf("got %v, want ErrOversize", err)
		}
	case <-ctx.Done():
		t.Fatal("server did not reject oversize frame")
	}
}

func TestAcceptHonoursCtx(t *testing.T) {
	t.Parallel()
	srv, _ := withTempServer(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := srv.Accept(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("got %v, want context.Canceled", err)
	}
}

func TestConcurrentSends(t *testing.T) {
	t.Parallel()
	srv, path := withTempServer(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var got atomic.Int64
	srvDone := make(chan error, 1)
	go func() {
		sess, err := srv.Accept(ctx)
		if err != nil {
			srvDone <- err
			return
		}
		defer sess.Close()
		for {
			_, err := sess.Recv(ctx)
			if err != nil {
				srvDone <- ignoreNetClose(err)
				return
			}
			got.Add(1)
		}
	}()

	cli, err := backchannel.Dial(&backchannel.DialArgs{AppName: "test", Path: path})
	if err != nil {
		t.Fatalf("Dial: %v", err)
	}

	const N = 50
	var wg sync.WaitGroup
	for i := range N {
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := cli.Send(ctx, backchannel.Message{
				Type:       backchannel.MsgPairBegin,
				InstanceID: fmt.Sprintf("%d", i),
			})
			if err != nil {
				t.Errorf("send %d: %v", i, err)
			}
		}()
	}
	wg.Wait()
	_ = cli.Close()

	select {
	case err := <-srvDone:
		if err != nil {
			t.Fatalf("server: %v", err)
		}
	case <-ctx.Done():
		t.Fatal("server stuck")
	}
	if got.Load() != int64(N) {
		t.Errorf("server got %d frames, want %d", got.Load(), N)
	}
}

func TestRejectsInvalidJSON(t *testing.T) {
	t.Parallel()
	srv, path := withTempServer(t)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	srvDone := make(chan error, 1)
	go func() {
		sess, err := srv.Accept(ctx)
		if err != nil {
			srvDone <- err
			return
		}
		defer sess.Close()
		_, err = sess.Recv(ctx)
		srvDone <- err
	}()

	raw, err := net.Dial("unix", path)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer raw.Close()
	junk := []byte("not-json{")
	var hdr [4]byte
	binary.BigEndian.PutUint32(hdr[:], uint32(len(junk)))
	frame := append(hdr[:], junk...)
	if _, err := raw.Write(frame); err != nil {
		t.Fatalf("write: %v", err)
	}

	select {
	case err := <-srvDone:
		if err == nil {
			t.Fatal("expected JSON decode error")
		}
	case <-ctx.Done():
		t.Fatal("server did not reject malformed JSON")
	}
}

func TestMessageJSONShape(t *testing.T) {
	t.Parallel()
	// Empty fields must be omitted so the wire matches the proposal in
	// docs/local-backchannel.md verbatim.
	msg := backchannel.Message{Type: backchannel.MsgPairBegin}
	b, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	want := []byte(`{"type":"pair_begin"}`)
	if !bytes.Equal(b, want) {
		t.Errorf("got %s, want %s", b, want)
	}
}

func TestEventMappingsTotal(t *testing.T) {
	t.Parallel()
	// Every MsgType the CLI may originate must map to an executor
	// event, and every daemon→CLI MsgType must have a reverse mapping.
	for _, mt := range []backchannel.MsgType{backchannel.MsgPairBegin, backchannel.MsgCodeSubmit} {
		if _, ok := backchannel.EventForMessage(mt); !ok {
			t.Errorf("EventForMessage(%q) = _, false; want ok", mt)
		}
	}
	for _, ev := range []string{backchannel.EventTokenCreated, backchannel.EventSignalCodeDisplay} {
		if _, ok := backchannel.MessageForEvent(ev); !ok {
			t.Errorf("MessageForEvent(%q) = _, false; want ok", ev)
		}
	}
	// Reverse direction: token_response / waiting_for_code MUST NOT
	// originate at the CLI.
	for _, mt := range []backchannel.MsgType{backchannel.MsgTokenResponse, backchannel.MsgWaitingForCode, backchannel.MsgPairStatus} {
		if _, ok := backchannel.EventForMessage(mt); ok {
			t.Errorf("EventForMessage(%q) = _, true; want false (daemon→CLI only)", mt)
		}
	}
}

// ---------- helpers ----------

func reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}

// ignoreNetClose collapses the various "peer closed" errors to nil.
// Connection close during a blocking Recv is the normal termination
// signal in the tests; treating it as failure would force every test
// to do its own teardown handshake.
func ignoreNetClose(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, net.ErrClosed) || errors.Is(err, context.Canceled) {
		return nil
	}
	// io.EOF surfaces as the wrapped read error on the server side
	// when the client closes its half cleanly.
	if err.Error() == "EOF" {
		return nil
	}
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return nil
	}
	return err
}

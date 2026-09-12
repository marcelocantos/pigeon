// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package pigeon

import (
	"bytes"
	"context"
	"encoding/binary"
	"io"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

var (
	_ streamCtl = (*quicStreamWrapper)(nil)
	_ streamCtl = (*wtStreamWrapper)(nil)
)

type recordCloser struct {
	closes atomic.Int32
}

func (r *recordCloser) Read([]byte) (int, error)    { return 0, io.EOF }
func (r *recordCloser) Write(p []byte) (int, error) { return len(p), nil }
func (r *recordCloser) Close() error {
	r.closes.Add(1)
	return nil
}

func testSession(t *testing.T) *Session {
	t.Helper()
	s := newSession(context.Background(), nil, nil, "peer", nil, true)
	s.rendezvousWindow = 40 * time.Millisecond
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestDeliverIncomingStreamExpiresUnmatched(t *testing.T) {
	s := testSession(t)
	rc := &recordCloser{}
	s.deliverIncomingStream("unexpected", rc)

	if n := s.bufferedStreamCount(); n != 1 {
		t.Fatalf("parked count = %d, want 1 during rendezvous", n)
	}

	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		if s.bufferedStreamCount() == 0 && rc.closes.Load() >= 1 {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}
	if s.bufferedStreamCount() != 0 {
		t.Fatalf("bufferedStreams still holding %d after rendezvous", s.bufferedStreamCount())
	}
	if got := rc.closes.Load(); got != 1 {
		t.Fatalf("Close called %d times, want 1", got)
	}
}

func TestDeliverIncomingStreamCapsPerName(t *testing.T) {
	s := testSession(t)
	a, b := &recordCloser{}, &recordCloser{}
	s.deliverIncomingStream("chat", a)
	s.deliverIncomingStream("chat", b) // over cap → immediate reset
	if got := b.closes.Load(); got != 1 {
		t.Fatalf("excess stream Close = %d, want 1", got)
	}
	if s.bufferedStreamCount() != 1 {
		t.Fatalf("buffered = %d, want 1", s.bufferedStreamCount())
	}
}

func TestDeliverIncomingStreamRespectsAllowlist(t *testing.T) {
	s := testSession(t)
	s.applyStreamPolicy([]string{"chat"})
	rc := &recordCloser{}
	s.deliverIncomingStream("nope", rc)
	if got := rc.closes.Load(); got != 1 {
		t.Fatalf("undeclared stream Close = %d, want 1", got)
	}
	if s.bufferedStreamCount() != 0 {
		t.Fatalf("undeclared name was parked")
	}
	ok := &recordCloser{}
	s.deliverIncomingStream("chat", ok)
	if ok.closes.Load() != 0 {
		t.Fatalf("declared name was reset during rendezvous")
	}
	if s.bufferedStreamCount() != 1 {
		t.Fatalf("declared name not parked")
	}
}

func TestReservedNamesBypassAllowlist(t *testing.T) {
	s := testSession(t)
	s.applyStreamPolicy([]string{"chat"})
	rc := &recordCloser{}
	s.deliverIncomingStream(discoveryStreamName, rc)
	if rc.closes.Load() != 0 {
		t.Fatalf("reserved discovery stream was reset")
	}
	if s.bufferedStreamCount() != 1 {
		t.Fatalf("reserved name should park when no handler is waiting")
	}
}

func TestAcceptStreamCancelClosesRacedDelivery(t *testing.T) {
	s := testSession(t)
	// Fable F10: cancelled AcceptStream racing deliverIncomingStream
	// must Close the delivered rwc exactly once.
	const rounds = 200
	for i := 0; i < rounds; i++ {
		rc := &recordCloser{}
		ctx, cancel := context.WithCancel(context.Background())
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			_, _ = s.AcceptStream(ctx, "s")
		}()
		go func() {
			defer wg.Done()
			time.Sleep(time.Microsecond)
			s.deliverIncomingStream("s", rc)
		}()
		time.Sleep(time.Microsecond)
		cancel()
		wg.Wait()

		// Either AcceptStream won (owns rwc, test doesn't close it) or
		// cancel/reset won. Drain any leftover park so the next round
		// is clean; Close the owned stream if Accept returned it.
		if buf := s.bufferedStreamCount(); buf > 0 {
			s.rejectParkedStreams()
		}
		if got := rc.closes.Load(); got > 1 {
			t.Fatalf("round %d: Close called %d times", i, got)
		}
	}
}

func TestAcceptStreamCancelClosesAfterDeliver(t *testing.T) {
	s := testSession(t)
	rc := &recordCloser{}
	ctx, cancel := context.WithCancel(context.Background())

	started := make(chan struct{})
	errCh := make(chan error, 1)
	go func() {
		close(started)
		_, err := s.AcceptStream(ctx, "s")
		errCh <- err
	}()
	<-started
	time.Sleep(5 * time.Millisecond) // waiter is registered
	s.deliverIncomingStream("s", rc)
	cancel()

	if err := <-errCh; err == nil {
		// Accept won the race and returned the stream — Close it.
		if rc.closes.Load() != 0 {
			t.Fatalf("AcceptStream returned but rwc was already closed")
		}
		_ = rc.Close()
		if rc.closes.Load() != 1 {
			t.Fatalf("Close after Accept = %d, want 1", rc.closes.Load())
		}
		return
	}
	deadline := time.Now().Add(200 * time.Millisecond)
	for time.Now().Before(deadline) && rc.closes.Load() == 0 {
		time.Sleep(time.Millisecond)
	}
	if got := rc.closes.Load(); got != 1 {
		t.Fatalf("cancelled AcceptStream left rwc unclosed (Close=%d)", got)
	}
}

func TestGoAwayResetsIncomingAndBlocksOpen(t *testing.T) {
	s := testSession(t)
	if err := s.GoAway(); err != nil {
		t.Fatal(err)
	}
	rc := &recordCloser{}
	s.deliverIncomingStream("chat", rc)
	if rc.closes.Load() != 1 {
		t.Fatalf("incoming after GoAway Close = %d, want 1", rc.closes.Load())
	}
	if _, err := s.OpenStream(context.Background(), "chat"); err != ErrSessionDraining {
		t.Fatalf("OpenStream after GoAway: %v, want ErrSessionDraining", err)
	}
}

func TestAcceptStreamCancelThenDeliverCloses(t *testing.T) {
	s := testSession(t)
	ctx, cancel := context.WithCancel(context.Background())
	started := make(chan struct{})
	errc := make(chan error, 1)
	s.deliverAfterPop = func() {
		cancel()
		time.Sleep(10 * time.Millisecond)
	}
	go func() {
		close(started)
		_, err := s.AcceptStream(ctx, "s")
		errc <- err
	}()
	<-started
	time.Sleep(15 * time.Millisecond)
	rc := &recordCloser{}
	s.deliverIncomingStream("s", rc)
	err := <-errc
	if err == nil {
		// Accept won; it owns rwc. Close it so the count is 1.
		_ = rc.Close()
	}
	if got := rc.closes.Load(); got != 1 {
		t.Fatalf("F10 race Close = %d, want 1", got)
	}
	if s.bufferedStreamCount() != 0 {
		t.Fatal("raced stream still parked")
	}
}

func TestUndeclaredStreamsDoNotGrowGoroutines(t *testing.T) {
	s := testSession(t)
	s.applyStreamPolicy([]string{"chat"})
	runtime.GC()
	before := runtime.NumGoroutine()
	for i := 0; i < 80; i++ {
		s.deliverIncomingStream("nope", &recordCloser{})
	}
	time.Sleep(20 * time.Millisecond)
	runtime.GC()
	after := runtime.NumGoroutine()
	if after > before+8 {
		t.Fatalf("goroutines grew from %d to %d after 80 undeclared streams", before, after)
	}
	if s.bufferedStreamCount() != 0 {
		t.Fatalf("undeclared streams parked: %d", s.bufferedStreamCount())
	}
}

type scriptedAcceptor struct {
	next chan io.ReadWriteCloser
}

func (a *scriptedAcceptor) AcceptStream(ctx context.Context) (io.ReadWriteCloser, error) {
	select {
	case rwc := <-a.next:
		return rwc, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

type framedRWC struct {
	r *bytes.Reader
	recordCloser
}

func (f *framedRWC) Read(p []byte) (int, error) { return f.r.Read(p) }

func frameMessage(payload []byte) []byte {
	var hdr [4]byte
	binary.BigEndian.PutUint32(hdr[:], uint32(len(payload)))
	return append(hdr[:], payload...)
}

func TestAcceptLoopClosesOnBadHeader(t *testing.T) {
	acc := &scriptedAcceptor{next: make(chan io.ReadWriteCloser, 1)}
	s := testSession(t)
	s.transport = &transport{acceptor: acc}
	go s.acceptLoop()

	rc := &framedRWC{r: bytes.NewReader(frameMessage(nil))}
	acc.next <- rc
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) && rc.closes.Load() == 0 {
		time.Sleep(time.Millisecond)
	}
	if got := rc.closes.Load(); got != 1 {
		t.Fatalf("acceptLoop parse-failure Close = %d, want 1", got)
	}
	if s.bufferedStreamCount() != 0 {
		t.Fatal("bad header was parked")
	}
}

func TestAcceptLoopReservedDiscoveryNotParked(t *testing.T) {
	acc := &scriptedAcceptor{next: make(chan io.ReadWriteCloser, 1)}
	s := testSession(t)
	s.transport = &transport{acceptor: acc}
	hit := make(chan struct{}, 1)
	s.discover = func(string) ([]RouteEntry, error) {
		select {
		case hit <- struct{}{}:
		default:
		}
		return nil, nil
	}
	go s.acceptLoop()

	// Header frame for acceptLoop, then an enumerate request so
	// handleEnumerate Recv succeeds and calls discover. A vacuous
	// bufferedStreamCount==0 check would pass before acceptLoop runs.
	payload := append(frameMessage(EncodeStreamHeader(discoveryStreamName)), frameMessage([]byte{0x01})...)
	acc.next <- &framedRWC{r: bytes.NewReader(payload)}
	select {
	case <-hit:
	case <-time.After(time.Second):
		t.Fatal("discover hook was not called; reserved path did not take the stream")
	}
	if s.bufferedStreamCount() != 0 {
		t.Fatal("discovery stream was parked as an application name")
	}
}

func TestAcceptLoopLANControlClosedWhenNotPreferLAN(t *testing.T) {
	acc := &scriptedAcceptor{next: make(chan io.ReadWriteCloser, 1)}
	s := testSession(t)
	s.transport = &transport{acceptor: acc}
	go s.acceptLoop()

	payload := EncodeStreamHeader(lanControlStreamName)
	rc := &framedRWC{r: bytes.NewReader(frameMessage(payload))}
	acc.next <- rc
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) && rc.closes.Load() == 0 {
		time.Sleep(time.Millisecond)
	}
	if rc.closes.Load() != 1 {
		t.Fatal("lan-control stream was not closed on non-PreferLAN session")
	}
	if s.bufferedStreamCount() != 0 {
		t.Fatal("lan-control stream was parked as an application name")
	}
}

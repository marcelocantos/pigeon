// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package tern

import (
	"bytes"
	"context"
	"crypto/rand"
	"strconv"
	"testing"
	"time"
)

// TestDatagramSmallPayload sends a small datagram (fits in one packet).
func TestDatagramSmallPayload(t *testing.T) {
	env := localRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	b, c := connectPair(t, env)

	msg := []byte("hello datagrams")
	if err := c.SendDatagram(msg); err != nil {
		t.Fatal("send:", err)
	}

	got, err := b.RecvDatagram(ctx)
	if err != nil {
		t.Fatal("recv:", err)
	}
	if string(got) != string(msg) {
		t.Fatalf("got %q, want %q", got, msg)
	}
}

// TestDatagramLargePayload sends a payload larger than one QUIC datagram.
// The integrated fragmentation should split and reassemble transparently.
func TestDatagramLargePayload(t *testing.T) {
	env := localRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	b, c := connectPair(t, env)

	// Force small max payload to trigger fragmentation.
	c.maxDgPayload = 50
	b.maxDgPayload = 50

	// 500 bytes → multiple fragments.
	payload := make([]byte, 500)
	for i := range payload {
		payload[i] = byte(i % 256)
	}

	if err := c.SendDatagram(payload); err != nil {
		t.Fatal("send:", err)
	}

	got, err := b.RecvDatagram(ctx)
	if err != nil {
		t.Fatal("recv:", err)
	}
	if !bytes.Equal(got, payload) {
		t.Fatalf("payload mismatch: got %d bytes, want %d", len(got), len(payload))
	}
}

// TestDatagramEmptyPayload sends an empty datagram.
func TestDatagramEmptyPayload(t *testing.T) {
	env := localRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	b, c := connectPair(t, env)

	if err := c.SendDatagram([]byte{}); err != nil {
		t.Fatal("send:", err)
	}

	got, err := b.RecvDatagram(ctx)
	if err != nil {
		t.Fatal("recv:", err)
	}
	if len(got) != 0 {
		t.Fatalf("got %d bytes, want 0", len(got))
	}
}

// TestDatagramMultipleLargeMessages sends several large messages and
// verifies they all arrive correctly.
func TestDatagramMultipleLargeMessages(t *testing.T) {
	env := localRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	b, c := connectPair(t, env)
	c.maxDgPayload = 100
	b.maxDgPayload = 100

	msgs := make([][]byte, 5)
	for i := range msgs {
		msgs[i] = []byte("message-" + strconv.Itoa(i) + "-" + string(make([]byte, 200)))
	}

	for _, msg := range msgs {
		if err := c.SendDatagram(msg); err != nil {
			t.Fatal("send:", err)
		}
	}

	for i, want := range msgs {
		got, err := b.RecvDatagram(ctx)
		if err != nil {
			t.Fatalf("recv %d: %v", i, err)
		}
		if !bytes.Equal(got, want) {
			t.Fatalf("message %d: got %d bytes, want %d", i, len(got), len(want))
		}
	}
}

// TestDatagramFragmentTimeout verifies that incomplete assemblies are
// dropped after the timeout expires.
func TestDatagramFragmentTimeout(t *testing.T) {
	// Use a mock datagrammer that only delivers some fragments.
	ch := make(chan []byte, 100)
	mock := &mockDatagram{ch: ch}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	reasm := newReassembler(500*time.Millisecond, done)
	defer close(done)

	// Manually send fragments for a 3-fragment message, but omit fragment 1.
	for _, idx := range []int{0, 2} {
		frame := make([]byte, 1+fragHeaderSize+10)
		frame[0] = dgConnFragment
		putUint32BE(frame[1:5], 42)
		putUint16BE(frame[5:7], uint16(idx))
		putUint16BE(frame[7:9], 3)
		copy(frame[1+fragHeaderSize:], []byte("payload..."))
		ch <- frame
	}

	// Create a conn that reads from mock, with the test reassembler.
	c := &Conn{
		dg:           mock,
		reasm:        reasm,
		maxDgPayload: DefaultMaxDatagramPayload,
	}
	c.ctx, c.cancel = context.WithCancel(ctx)

	// RecvDatagram should timeout — fragment 1 never arrives.
	recvCtx, recvCancel := context.WithTimeout(ctx, 2*time.Second)
	defer recvCancel()
	_, err := c.RecvDatagram(recvCtx)
	if err == nil {
		t.Fatal("expected timeout for incomplete assembly")
	}

	// Verify the assembly was cleaned up.
	time.Sleep(time.Second)
	reasm.mu.Lock()
	remaining := len(reasm.assemblies)
	reasm.mu.Unlock()
	if remaining != 0 {
		t.Fatalf("expected 0 assemblies after timeout, got %d", remaining)
	}
}

// TestDatagramBidirectionalLarge sends large messages in both directions.
func TestDatagramBidirectionalLarge(t *testing.T) {
	env := localRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	b, c := connectPair(t, env)
	c.maxDgPayload = 80
	b.maxDgPayload = 80

	payload1 := bytes.Repeat([]byte("A"), 300)
	c.SendDatagram(payload1)

	payload2 := bytes.Repeat([]byte("B"), 400)
	b.SendDatagram(payload2)

	got1, err := b.RecvDatagram(ctx)
	if err != nil {
		t.Fatal("recv on backend:", err)
	}
	if !bytes.Equal(got1, payload1) {
		t.Fatalf("backend got %d bytes, want %d", len(got1), len(payload1))
	}

	got2, err := c.RecvDatagram(ctx)
	if err != nil {
		t.Fatal("recv on client:", err)
	}
	if !bytes.Equal(got2, payload2) {
		t.Fatalf("client got %d bytes, want %d", len(got2), len(payload2))
	}
}

// TestDatagramLargeUnderLoss sends fragmented datagrams through a
// lossy path. Some messages may not arrive. Verify no crashes.
func TestDatagramLargeUnderLoss(t *testing.T) {
	env := localRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	b, c := connectPair(t, env)
	c.maxDgPayload = 80
	b.maxDgPayload = 80

	for i := range 10 {
		payload := make([]byte, 200)
		rand.Read(payload)
		payload[0] = byte(i)
		c.SendDatagram(payload)
	}

	received := 0
	recvCtx, recvCancel := context.WithTimeout(ctx, 5*time.Second)
	defer recvCancel()
	for {
		_, err := b.RecvDatagram(recvCtx)
		if err != nil {
			break
		}
		received++
	}
	t.Logf("large datagrams: sent=10, received=%d", received)
}

// TestDatagramTooLarge verifies rejection of payloads that would
// exceed the fragment count limit.
func TestDatagramTooLarge(t *testing.T) {
	env := localRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, c := connectPair(t, env)
	// maxPayload = 10 means 1 byte per chunk (10 - 1 prefix - 8 header = 1).
	// 65536 bytes → 65536 fragments > 65535 max.
	c.maxDgPayload = 10

	payload := make([]byte, 65536)
	err := c.SendDatagram(payload)
	if err != ErrDatagramTooLarge {
		t.Fatalf("expected ErrDatagramTooLarge, got %v", err)
	}
	_ = ctx
}

// TestDatagramMixedSmallAndLarge interleaves small (single-packet) and
// large (fragmented) datagrams and verifies correct delivery of both.
func TestDatagramMixedSmallAndLarge(t *testing.T) {
	env := localRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	b, c := connectPair(t, env)
	c.maxDgPayload = 80
	b.maxDgPayload = 80

	// Send: small, large, small, large.
	c.SendDatagram([]byte("small-1"))
	c.SendDatagram(bytes.Repeat([]byte("L"), 200))
	c.SendDatagram([]byte("small-2"))
	c.SendDatagram(bytes.Repeat([]byte("M"), 300))

	got1, _ := b.RecvDatagram(ctx)
	if string(got1) != "small-1" {
		t.Fatalf("msg 1: got %d bytes %q", len(got1), got1[:min(len(got1), 20)])
	}

	got2, _ := b.RecvDatagram(ctx)
	if len(got2) != 200 {
		t.Fatalf("msg 2: got %d bytes, want 200", len(got2))
	}

	got3, _ := b.RecvDatagram(ctx)
	if string(got3) != "small-2" {
		t.Fatalf("msg 3: got %q", got3)
	}

	got4, _ := b.RecvDatagram(ctx)
	if len(got4) != 300 {
		t.Fatalf("msg 4: got %d bytes, want 300", len(got4))
	}
}

// --- helpers ---

type mockDatagram struct {
	ch chan []byte
}

func (m *mockDatagram) SendDatagram(data []byte) error {
	m.ch <- data
	return nil
}

func (m *mockDatagram) ReceiveDatagram(ctx context.Context) ([]byte, error) {
	select {
	case d := <-m.ch:
		return d, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func putUint32BE(b []byte, v uint32) { b[0] = byte(v >> 24); b[1] = byte(v >> 16); b[2] = byte(v >> 8); b[3] = byte(v) }
func putUint16BE(b []byte, v uint16) { b[0] = byte(v >> 8); b[1] = byte(v) }

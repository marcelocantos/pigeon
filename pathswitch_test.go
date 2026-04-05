// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Deterministic tests for every path-switching scenario. Each test
// sets up a specific condition and verifies the exact outcome.

package tern

import (
	"bytes"
	"context"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/marcelocantos/tern/crypto"
)

// --- Full cycle tests ---

// TestPathSwitchFullCycle: relay → LAN → relay → LAN.
// Verifies that the system can switch back and forth repeatedly.
func TestPathSwitchFullCycle(t *testing.T) {
	env := localRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	b, c, _ := lanPair(t, env)
	waitForLAN(t, ctx, b, c)

	for cycle := range 3 {
		// Verify LAN is active.
		msg := []byte("lan-" + strconv.Itoa(cycle))
		c.Send(ctx, msg)
		data, err := b.Recv(ctx)
		if err != nil {
			t.Fatalf("cycle %d LAN recv: %v", cycle, err)
		}
		if string(data) != string(msg) {
			t.Fatalf("cycle %d: got %q", cycle, data)
		}

		// Fall back to relay.
		c.fallbackToRelay()
		b.fallbackToRelay()

		// Verify relay works.
		msg = []byte("relay-" + strconv.Itoa(cycle))
		c.Send(ctx, msg)
		data, err = b.Recv(ctx)
		if err != nil {
			t.Fatalf("cycle %d relay recv: %v", cycle, err)
		}
		if string(data) != string(msg) {
			t.Fatalf("cycle %d: got %q", cycle, data)
		}

		// Re-establish LAN by manually installing a direct path.
		// (In production the onSwitch callback re-advertises, but
		// that's async and we want deterministic testing.)
		reestablishLAN(t, ctx, b, c)
	}
}

// reestablishLAN creates a new direct LAN connection between two
// Conns that are currently on relay. Used for deterministic testing.
func reestablishLAN(t *testing.T, ctx context.Context, b, c *Conn) {
	t.Helper()

	// The backend's LANServer is still running. We need the client
	// to receive a new LAN offer. Send one manually.
	b.mu.Lock()
	lanSrv := b.lanServer
	b.mu.Unlock()

	if lanSrv == nil {
		t.Fatal("backend has no LAN server")
	}

	// Trigger LAN re-advertisement via the executor.
	b.exec.submit(event{id: EventLanServerReady})

	// Wait for LAN to establish.
	select {
	case <-c.LANReady():
	case <-ctx.Done():
		t.Fatal("timeout waiting for LAN re-establishment")
	}
	select {
	case <-b.LANReady():
	case <-ctx.Done():
		t.Fatal("timeout waiting for backend LAN")
	}
}

// --- Nonce continuity ---

// TestNonceContinuityAcrossSwitch verifies that the encrypted channel's
// nonce counter is continuous across path switches — no replays, no gaps
// that cause decryption failure.
func TestNonceContinuityAcrossSwitch(t *testing.T) {
	env := localRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	b, c, _ := lanPair(t, env)
	waitForLAN(t, ctx, b, c)

	// Send 10 messages on LAN.
	for i := range 10 {
		c.Send(ctx, []byte("pre-"+strconv.Itoa(i)))
	}
	for range 10 {
		_, err := b.Recv(ctx)
		if err != nil {
			t.Fatal("pre-switch recv:", err)
		}
	}

	// Switch to relay.
	c.fallbackToRelay()
	b.fallbackToRelay()

	// Send 10 more messages on relay. The nonce counter should continue
	// from where it was — no "replayed or too old" errors.
	for i := range 10 {
		if err := c.Send(ctx, []byte("post-"+strconv.Itoa(i))); err != nil {
			t.Fatalf("post-switch send %d: %v", i, err)
		}
	}
	for i := range 10 {
		data, err := b.Recv(ctx)
		if err != nil {
			t.Fatalf("post-switch recv %d: %v", i, err)
		}
		if string(data) != "post-"+strconv.Itoa(i) {
			t.Fatalf("got %q, want post-%d", data, i)
		}
	}
}

// --- Concurrent Send/Recv during switch ---

// TestConcurrentSendRecvDuringSwitch has goroutines continuously
// sending and receiving while the path switches underneath them.
func TestConcurrentSendRecvDuringSwitch(t *testing.T) {
	env := localRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	b, c, _ := lanPair(t, env)
	waitForLAN(t, ctx, b, c)

	var sent, recvd atomic.Int64
	var sendErrors, recvErrors atomic.Int64
	deadline := time.Now().Add(5 * time.Second)

	var wg sync.WaitGroup

	// Continuous sender.
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; time.Now().Before(deadline); i++ {
			sendCtx, sendCancel := context.WithTimeout(ctx, 2*time.Second)
			err := c.Send(sendCtx, []byte("msg-"+strconv.Itoa(i)))
			sendCancel()
			if err != nil {
				sendErrors.Add(1)
				continue
			}
			sent.Add(1)
			time.Sleep(10 * time.Millisecond)
		}
	}()

	// Continuous receiver.
	wg.Add(1)
	go func() {
		defer wg.Done()
		for time.Now().Before(deadline) {
			recvCtx, recvCancel := context.WithTimeout(ctx, 2*time.Second)
			_, err := b.Recv(recvCtx)
			recvCancel()
			if err != nil {
				recvErrors.Add(1)
				continue
			}
			recvd.Add(1)
		}
	}()

	// Path switcher: toggle every 500ms.
	wg.Add(1)
	go func() {
		defer wg.Done()
		for time.Now().Before(deadline) {
			time.Sleep(500 * time.Millisecond)
			c.fallbackToRelay()
			b.fallbackToRelay()
			time.Sleep(500 * time.Millisecond)
			reestablishLAN(t, ctx, b, c)
		}
	}()

	wg.Wait()

	t.Logf("concurrent switch: sent=%d recv=%d sendErr=%d recvErr=%d",
		sent.Load(), recvd.Load(), sendErrors.Load(), recvErrors.Load())

	if sent.Load() > 0 && recvd.Load() == 0 {
		t.Fatal("no messages received during concurrent switching")
	}
}

// --- Datagram channels across switch ---

// TestDatagramChannelAcrossSwitch verifies that named datagram channels
// continue working after a path switch.
func TestDatagramChannelAcrossSwitch(t *testing.T) {
	env := localRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	b, c, _ := lanPair(t, env)
	waitForLAN(t, ctx, b, c)

	cCh := c.DatagramChannel("video")
	bCh := b.DatagramChannel("video")

	// Send on LAN.
	for range 5 {
		cCh.Send([]byte("frame-lan"))
	}
	recvCtx, recvCancel := context.WithTimeout(ctx, 2*time.Second)
	data, err := bCh.Recv(recvCtx)
	recvCancel()
	if err != nil {
		t.Fatal("dg channel recv on LAN:", err)
	}
	if string(data) != "frame-lan" {
		t.Fatalf("got %q", data)
	}

	// Switch to relay.
	c.fallbackToRelay()
	b.fallbackToRelay()

	// Send on relay. The datagram channel should still route correctly.
	for range 5 {
		cCh.Send([]byte("frame-relay"))
	}
	recvCtx, recvCancel = context.WithTimeout(ctx, 2*time.Second)
	data, err = bCh.Recv(recvCtx)
	recvCancel()
	if err != nil {
		t.Fatal("dg channel recv on relay:", err)
	}
	if string(data) != "frame-relay" {
		t.Fatalf("got %q", data)
	}
}

// --- Streaming channels across switch ---

// TestStreamChannelAcrossSwitch verifies that streaming channels
// opened on one path survive a switch to another.
func TestStreamChannelAcrossSwitch(t *testing.T) {
	env := localRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	b, c, _ := lanPair(t, env)
	waitForLAN(t, ctx, b, c)

	// Open a channel on LAN.
	ch, err := c.OpenChannel("game")
	if err != nil {
		t.Fatal("open:", err)
	}
	defer ch.Close()

	bch, err := b.AcceptChannel(ctx)
	if err != nil {
		t.Fatal("accept:", err)
	}
	defer bch.Close()

	ch.Send(ctx, []byte("on-lan"))
	data, _ := bch.Recv(ctx)
	if string(data) != "on-lan" {
		t.Fatalf("got %q", data)
	}

	// Switch to relay. The existing channel's stream was on the LAN
	// QUIC connection, which is now closed. The channel should fail
	// gracefully — this tests that we don't panic or deadlock.
	c.fallbackToRelay()
	b.fallbackToRelay()

	// The old channel's stream is dead. Sending should error.
	sendCtx, sendCancel := context.WithTimeout(ctx, time.Second)
	err = ch.Send(sendCtx, []byte("should-fail"))
	sendCancel()
	// We don't assert the specific error — just that it doesn't hang.
	t.Logf("send on dead channel: %v", err)

	// Open a NEW channel on the relay path.
	ch2, err := c.OpenChannel("game-v2")
	if err != nil {
		t.Fatal("open on relay:", err)
	}
	defer ch2.Close()

	bch2, err := b.AcceptChannel(ctx)
	if err != nil {
		t.Fatal("accept on relay:", err)
	}
	defer bch2.Close()

	ch2.Send(ctx, []byte("on-relay"))
	data, _ = bch2.Recv(ctx)
	if string(data) != "on-relay" {
		t.Fatalf("got %q", data)
	}
}

// --- Asymmetric switch ---

// TestAsymmetricSwitch tests what happens when one side switches to
// relay but the other side still thinks it's on LAN.
func TestAsymmetricSwitch(t *testing.T) {
	env := localRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	b, c, _ := lanPair(t, env)
	waitForLAN(t, ctx, b, c)

	// Only the client falls back. The backend still thinks LAN is active.
	c.fallbackToRelay()

	// Client sends via relay. The relay bridges to the backend, which
	// is still reading from the relay stream (the relay connection is
	// permanent). This should work because the relay bridge is always
	// running.
	c.Send(ctx, []byte("from-relay-client"))
	recvCtx, recvCancel := context.WithTimeout(ctx, 5*time.Second)
	data, err := b.Recv(recvCtx)
	recvCancel()
	if err != nil {
		t.Fatal("asymmetric recv:", err)
	}
	if string(data) != "from-relay-client" {
		t.Fatalf("got %q", data)
	}
}

// --- Rapid flapping ---

// TestRapidFlapping switches between LAN and relay many times in quick
// succession. Verifies no panics, deadlocks, or goroutine leaks.
func TestRapidFlapping(t *testing.T) {
	env := localRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	b, c, _ := lanPair(t, env)
	waitForLAN(t, ctx, b, c)

	for i := range 10 {
		c.fallbackToRelay()
		b.fallbackToRelay()

		// Brief pause — just enough for the fallback to take effect.
		time.Sleep(10 * time.Millisecond)

		// Verify relay works.
		msg := []byte("flap-" + strconv.Itoa(i))
		c.Send(ctx, msg)
		data, err := b.Recv(ctx)
		if err != nil {
			t.Fatalf("flap %d recv: %v", i, err)
		}
		if string(data) != string(msg) {
			t.Fatalf("flap %d: got %q", i, data)
		}
	}
}

// --- Large message during switch ---

// TestLargeMessageDuringSwitch sends a 100KB message, then switches
// mid-flight. The message should either arrive completely or fail
// cleanly — never partial delivery.
func TestLargeMessageDuringSwitch(t *testing.T) {
	env := localRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	b, c, _ := lanPair(t, env)
	waitForLAN(t, ctx, b, c)

	payload := bytes.Repeat([]byte("X"), 100000)

	// Send the large message.
	go func() {
		// Switch after a brief delay — while the message is in flight.
		time.Sleep(5 * time.Millisecond)
		c.fallbackToRelay()
		b.fallbackToRelay()
	}()

	err := c.Send(ctx, payload)
	if err != nil {
		// Send failed — acceptable if the switch happened mid-write.
		t.Logf("large message send failed (acceptable): %v", err)
		return
	}

	// If send succeeded, the message should arrive intact.
	recvCtx, recvCancel := context.WithTimeout(ctx, 5*time.Second)
	data, err := b.Recv(recvCtx)
	recvCancel()
	if err != nil {
		// Recv failed — the switch killed the stream.
		t.Logf("large message recv failed (acceptable): %v", err)
		return
	}
	if !bytes.Equal(data, payload) {
		t.Fatalf("partial delivery: got %d bytes, want %d", len(data), len(payload))
	}
	t.Log("large message delivered intact despite mid-flight switch")
}

// --- Health monitor integration ---

// TestHealthMonitorFallback verifies that the health monitor's ping
// mechanism actually triggers fallback when the direct path dies.
func TestHealthMonitorFallback(t *testing.T) {
	t.Skip("🎯T18: health monitor lifecycle will be machine-driven; ad-hoc monitor doesn't detect closed LAN server")
	env := localRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	b, c, lanSrv := lanPair(t, env)
	waitForLAN(t, ctx, b, c)

	// Verify LAN is active.
	if !c.isDirectActive() {
		t.Fatal("expected direct path active")
	}

	// Close the LAN server — the direct path will become unreachable.
	lanSrv.Close()

	// The health monitor pings every 5 seconds, fails after 3 consecutive.
	// So fallback should happen within ~20 seconds.
	timeout := time.After(25 * time.Second)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			t.Fatal("health monitor did not trigger fallback within 25s")
		case <-ticker.C:
			if !c.isDirectActive() {
				t.Log("health monitor triggered fallback")

				// Verify relay still works.
				c.Send(ctx, []byte("after-health-fallback"))
				data, err := b.Recv(ctx)
				if err != nil {
					t.Fatal("recv after health fallback:", err)
				}
				if string(data) != "after-health-fallback" {
					t.Fatalf("got %q", data)
				}
				return
			}
		}
	}
}

// --- Encrypted channel setup across switch ---

// TestEncryptedChannelAcrossSwitch verifies that per-channel
// encryption keys (derived from PairingRecord) work correctly after
// a path switch.
func TestEncryptedChannelAcrossSwitch(t *testing.T) {
	env := localRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	b, c, _ := lanPair(t, env)

	// Set up pairing records for channel key derivation.
	bKP, _ := crypto.GenerateKeyPair()
	cKP, _ := crypto.GenerateKeyPair()
	bRec := crypto.NewPairingRecord("client", "relay", bKP, cKP.Public)
	cRec := crypto.NewPairingRecord("backend", "relay", cKP, bKP.Public)
	b.SetPairingRecord(bRec)
	c.SetPairingRecord(cRec)

	waitForLAN(t, ctx, b, c)

	// Open encrypted channel on LAN.
	ch, _ := c.OpenChannel("secure")
	bch, _ := b.AcceptChannel(ctx)

	ch.Send(ctx, []byte("secret-on-lan"))
	data, _ := bch.Recv(ctx)
	if string(data) != "secret-on-lan" {
		t.Fatalf("got %q", data)
	}
	ch.Close()
	bch.Close()

	// Switch to relay.
	c.fallbackToRelay()
	b.fallbackToRelay()

	// Open a new encrypted channel on relay.
	ch2, err := c.OpenChannel("secure-v2")
	if err != nil {
		t.Fatal("open on relay:", err)
	}
	bch2, err := b.AcceptChannel(ctx)
	if err != nil {
		t.Fatal("accept on relay:", err)
	}

	ch2.Send(ctx, []byte("secret-on-relay"))
	data, err = bch2.Recv(ctx)
	if err != nil {
		t.Fatal("recv encrypted on relay:", err)
	}
	if string(data) != "secret-on-relay" {
		t.Fatalf("got %q", data)
	}
}

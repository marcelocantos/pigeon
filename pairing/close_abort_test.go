// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package pairing_test

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/marcelocantos/pigeon/crypto"
	"github.com/marcelocantos/pigeon/pairing"
)

// waitGoroutineGone polls the full goroutine stack dump until no goroutine
// has substr in its stack, or fails after timeout. It is the direct oracle
// for "the ceremony driver goroutine unwound".
func waitGoroutineGone(t *testing.T, substr string, timeout time.Duration) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	buf := make([]byte, 1<<20)
	for {
		n := runtime.Stack(buf, true)
		if !strings.Contains(string(buf[:n]), substr) {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("goroutine %q still parked %v after Close — pre-confirm Close leaked it\n%s",
				substr, timeout, buf[:n])
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// TestCeremonyCloseAbortsPendingConfirm reproduces Fable-5 F6 (🎯T54): a
// Ceremony blocked awaiting user confirmation (the C driver parked inside
// the confirm callback) must be released by Close. Before the fix the
// initiator's confirm select watched context.Background() (never fires) and
// Close touched neither that ctx nor confirmCh, so runInitiator wedged
// forever — leaking the goroutine, its two cgo.Handles and the QUIC session.
//
// The test drives a real ceremony over an in-process relay until BOTH sides
// are parked at the confirm wait (Code() has returned on each), then calls
// Close WITHOUT Confirm and asserts both driver goroutines exit.
func TestCeremonyCloseAbortsPendingConfirm(t *testing.T) {
	relayURL := startTestRelay(t)
	tmp := t.TempDir()

	bid, err := crypto.NewFileIdentity(tmp + "/acceptor-id.json")
	if err != nil {
		t.Fatalf("acceptor identity: %v", err)
	}
	cid, err := crypto.NewFileIdentity(tmp + "/initiator-id.json")
	if err != nil {
		t.Fatalf("initiator identity: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	pairer, err := pairing.Register(ctx, &pairing.Args{Relay: relayURL, Identity: bid})
	if err != nil {
		t.Fatalf("pairing.Register: %v", err)
	}
	// Keep the Pairer alive across the assertion: the acceptor's leak must
	// be released by cer.Close alone, not by Pairer.Close.
	defer pairer.Close()

	tokenCh := make(chan string, 1)
	accCerCh := make(chan *pairing.Ceremony, 1)
	accCodeErr := make(chan error, 1)
	go func() {
		cer, err := pairer.Accept(ctx)
		if err != nil {
			accCodeErr <- fmt.Errorf("accept: %w", err)
			return
		}
		accCerCh <- cer
		tokenCh <- cer.Token
		_, err = cer.Code(ctx) // returns once parked at the confirm wait
		accCodeErr <- err
	}()

	var token string
	select {
	case token = <-tokenCh:
	case <-time.After(8 * time.Second):
		t.Fatalf("timed out waiting for acceptor token")
	}

	icer, err := pairing.Initiate(ctx, &pairing.InitiateArgs{Rendezvous: token, Identity: cid})
	if err != nil {
		t.Fatalf("pairing.Initiate: %v", err)
	}

	if _, err := icer.Code(ctx); err != nil {
		t.Fatalf("initiator Code: %v", err)
	}
	if err := <-accCodeErr; err != nil {
		t.Fatalf("acceptor Code: %v", err)
	}
	accCer := <-accCerCh

	// Both drivers are now parked inside the confirm callback. Abort by
	// Close, never calling Confirm.
	if err := icer.Close(); err != nil {
		t.Fatalf("initiator Close: %v", err)
	}
	if err := accCer.Close(); err != nil {
		t.Fatalf("acceptor Close: %v", err)
	}

	waitGoroutineGone(t, "pigeon/pairing.runInitiator", 8*time.Second)
	waitGoroutineGone(t, "pigeon/pairing.runAcceptor", 8*time.Second)
}

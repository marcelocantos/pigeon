// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package tern

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/marcelocantos/tern/crypto"
)

// TestStressStreamBurst sends a burst of messages and verifies all arrive in order.
func TestStressStreamBurst(t *testing.T) {
	forEachRelay(t, func(t *testing.T, env relayEnv) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		b, c := connectPair(t, env)

		const n = 1000

		// Backend collects all messages.
		done := make(chan []string, 1)
		go func() {
			var received []string
			for range n {
				data, err := b.Recv(ctx)
				if err != nil {
					break
				}
				received = append(received, string(data))
			}
			done <- received
		}()

		// Client sends burst.
		for i := range n {
			if err := c.Send(ctx, []byte("msg-"+strconv.Itoa(i))); err != nil {
				t.Fatalf("send %d: %v", i, err)
			}
		}

		received := <-done
		if len(received) != n {
			t.Fatalf("received %d/%d messages", len(received), n)
		}
		for i, msg := range received {
			if msg != "msg-"+strconv.Itoa(i) {
				t.Fatalf("msg %d: got %q, want %q", i, msg, "msg-"+strconv.Itoa(i))
			}
		}
		t.Logf("burst: %d messages delivered in order", n)
	})
}

// TestStressBidirectionalConcurrent sends messages in both directions concurrently.
func TestStressBidirectionalConcurrent(t *testing.T) {
	forEachRelay(t, func(t *testing.T, env relayEnv) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		b, c := connectPair(t, env)

		const n = 500
		var wg sync.WaitGroup

		// Client → backend.
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range n {
				if err := c.Send(ctx, []byte(fmt.Sprintf("c2b-%d", i))); err != nil {
					return
				}
			}
		}()

		// Backend → client.
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range n {
				if err := b.Send(ctx, []byte(fmt.Sprintf("b2c-%d", i))); err != nil {
					return
				}
			}
		}()

		// Receive client→backend.
		wg.Add(1)
		c2bCount := 0
		go func() {
			defer wg.Done()
			for range n {
				if _, err := b.Recv(ctx); err != nil {
					return
				}
				c2bCount++
			}
		}()

		// Receive backend→client.
		wg.Add(1)
		b2cCount := 0
		go func() {
			defer wg.Done()
			for range n {
				if _, err := c.Recv(ctx); err != nil {
					return
				}
				b2cCount++
			}
		}()

		wg.Wait()
		t.Logf("bidirectional: c2b=%d/%d, b2c=%d/%d", c2bCount, n, b2cCount, n)
		if c2bCount != n {
			t.Fatalf("c2b: got %d, want %d", c2bCount, n)
		}
		if b2cCount != n {
			t.Fatalf("b2c: got %d, want %d", b2cCount, n)
		}
	})
}

// TestStressEncryptedBurst sends encrypted messages at high volume.
func TestStressEncryptedBurst(t *testing.T) {
	forEachRelay(t, func(t *testing.T, env relayEnv) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		b, c := connectPair(t, env)

		bKP, _ := crypto.GenerateKeyPair()
		cKP, _ := crypto.GenerateKeyPair()
		bSendKey, _ := crypto.DeriveSessionKey(bKP.Private, cKP.Public, []byte("b2c"))
		bRecvKey, _ := crypto.DeriveSessionKey(bKP.Private, cKP.Public, []byte("c2b"))
		cSendKey, _ := crypto.DeriveSessionKey(cKP.Private, bKP.Public, []byte("c2b"))
		cRecvKey, _ := crypto.DeriveSessionKey(cKP.Private, bKP.Public, []byte("b2c"))
		bCh, _ := crypto.NewChannel(bSendKey, bRecvKey)
		cCh, _ := crypto.NewChannel(cSendKey, cRecvKey)
		b.SetChannel(bCh)
		c.SetChannel(cCh)

		const n = 1000

		done := make(chan int, 1)
		go func() {
			count := 0
			for range n {
				if _, err := b.Recv(ctx); err != nil {
					break
				}
				count++
			}
			done <- count
		}()

		for i := range n {
			if err := c.Send(ctx, []byte("enc-"+strconv.Itoa(i))); err != nil {
				t.Fatalf("send %d: %v", i, err)
			}
		}

		count := <-done
		if count != n {
			t.Fatalf("encrypted burst: received %d/%d", count, n)
		}
		t.Logf("encrypted burst: %d messages", n)
	})
}

// TestStressDatagramFlood sends a flood of datagrams and checks how many arrive.
// Datagram delivery is unreliable, so we just verify no crashes and report stats.
func TestStressDatagramFlood(t *testing.T) {
	forEachRelay(t, func(t *testing.T, env relayEnv) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		b, c := connectPair(t, env)

		const n = 1000

		done := make(chan int, 1)
		go func() {
			count := 0
			for {
				recvCtx, recvCancel := context.WithTimeout(ctx, 2*time.Second)
				_, err := b.RecvDatagram(recvCtx)
				recvCancel()
				if err != nil {
					break
				}
				count++
			}
			done <- count
		}()

		sent := 0
		for i := range n {
			if err := c.SendDatagram([]byte("dg-" + strconv.Itoa(i))); err != nil {
				break
			}
			sent++
		}

		received := <-done
		lossRate := float64(sent-received) / float64(sent) * 100
		t.Logf("datagram flood: sent=%d received=%d loss=%.1f%%", sent, received, lossRate)
		if received == 0 {
			t.Fatal("no datagrams received")
		}
	})
}

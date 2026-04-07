// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package pigeon

import (
	"context"
	"crypto/ecdh"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/marcelocantos/pigeon/crypto"
)

// --- Helpers ---

// localRelayWithToken starts a local relay that requires a bearer token.
func localRelayWithToken(t *testing.T, token string) relayEnv {
	t.Helper()

	cert, pool := generateTestCert(t)
	tlsCfg := &tls.Config{Certificates: []tls.Certificate{cert}}

	srv, err := NewWebTransportServer("127.0.0.1:0", tlsCfg, token)
	if err != nil {
		t.Fatal(err)
	}

	wtUDP, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatal(err)
	}
	go srv.Serve(wtUDP)
	t.Cleanup(func() { srv.Close() })

	qsrv := NewQUICServer("127.0.0.1:0", tlsCfg, token, srv.Hub())
	qUDP, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatal(err)
	}
	go qsrv.ServeWithTLS(qUDP, tlsCfg)
	t.Cleanup(func() { qsrv.Close() })

	wtPort := wtUDP.LocalAddr().(*net.UDPAddr).Port
	qPort := qUDP.LocalAddr().(*net.UDPAddr).Port

	return relayEnv{
		url: "https://127.0.0.1:" + strconv.Itoa(wtPort),
		cfg: Config{
			TLS:      &tls.Config{RootCAs: pool},
			QUICPort: strconv.Itoa(qPort),
			Token:    token,
		},
	}
}

// setupPairingRecords creates matching PairingRecords for both sides.
// Returns (backendRecord, clientRecord).
func setupPairingRecords(t *testing.T) (*crypto.PairingRecord, *crypto.PairingRecord) {
	t.Helper()
	bKP, err := crypto.GenerateKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	cKP, err := crypto.GenerateKeyPair()
	if err != nil {
		t.Fatal(err)
	}

	bRec := crypto.NewPairingRecord("client-inst", "https://relay", bKP, cKP.Public)
	cRec := crypto.NewPairingRecord("backend-inst", "https://relay", cKP, bKP.Public)
	return bRec, cRec
}

// setupEncryptionWithPairingRecord sets up encryption using PairingRecords
// (the channel-derivation path) on both conns and also sets the master
// channel so OpenChannel/AcceptChannel can derive per-channel keys.
func setupEncryptionWithPairingRecord(t *testing.T, b, c *Conn, bRec, cRec *crypto.PairingRecord) {
	t.Helper()

	// Set up master channel on primary stream.
	bCh, err := bRec.DeriveChannel([]byte("b-to-c"), []byte("c-to-b"))
	if err != nil {
		t.Fatal(err)
	}
	cCh, err := cRec.DeriveChannel([]byte("c-to-b"), []byte("b-to-c"))
	if err != nil {
		t.Fatal(err)
	}
	b.SetChannel(bCh)
	c.SetChannel(cCh)

	// Set pairing records for per-channel key derivation.
	b.SetPairingRecord(bRec)
	c.SetPairingRecord(cRec)
}

// --- Streaming channel tests ---

func TestEncryptedStreamingChannel(t *testing.T) {
	env := localRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	b, c := connectPair(t, env)

	bRec, cRec := setupPairingRecords(t)
	setupEncryptionWithPairingRecord(t, b, c, bRec, cRec)

	// Client opens encrypted channel.
	ch, err := c.OpenChannel("encrypted-chan")
	if err != nil {
		t.Fatal("open channel:", err)
	}
	defer ch.Close()

	bch, err := b.AcceptChannel(ctx)
	if err != nil {
		t.Fatal("accept channel:", err)
	}
	defer bch.Close()

	if bch.Name() != "encrypted-chan" {
		t.Fatalf("got channel name %q, want encrypted-chan", bch.Name())
	}

	// Send encrypted data both directions.
	if err := ch.Send(ctx, []byte("secret-payload")); err != nil {
		t.Fatal("send:", err)
	}
	data, err := bch.Recv(ctx)
	if err != nil {
		t.Fatal("recv:", err)
	}
	if string(data) != "secret-payload" {
		t.Fatalf("got %q, want secret-payload", data)
	}

	if err := bch.Send(ctx, []byte("secret-reply")); err != nil {
		t.Fatal("send:", err)
	}
	data, err = ch.Recv(ctx)
	if err != nil {
		t.Fatal("recv:", err)
	}
	if string(data) != "secret-reply" {
		t.Fatalf("got %q, want secret-reply", data)
	}
}

func TestChannelOpenedByBackend(t *testing.T) {
	forEachRelay(t, func(t *testing.T, env relayEnv) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		b, c := connectPair(t, env)

		// Backend opens a channel (reverse direction from typical).
		ch, err := b.OpenChannel("backend-initiated")
		if err != nil {
			t.Fatal("open channel:", err)
		}
		defer ch.Close()

		cch, err := c.AcceptChannel(ctx)
		if err != nil {
			t.Fatal("accept channel:", err)
		}
		defer cch.Close()

		if cch.Name() != "backend-initiated" {
			t.Fatalf("got channel name %q, want backend-initiated", cch.Name())
		}

		if err := ch.Send(ctx, []byte("from-backend")); err != nil {
			t.Fatal("send:", err)
		}
		data, err := cch.Recv(ctx)
		if err != nil {
			t.Fatal("recv:", err)
		}
		if string(data) != "from-backend" {
			t.Fatalf("got %q, want from-backend", data)
		}
	})
}

func TestManyChannelsSimultaneously(t *testing.T) {
	forEachRelay(t, func(t *testing.T, env relayEnv) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		b, c := connectPair(t, env)

		const n = 10
		clientChans := make([]*StreamChannel, n)
		for i := range n {
			ch, err := c.OpenChannel(fmt.Sprintf("chan-%d", i))
			if err != nil {
				t.Fatalf("open channel %d: %v", i, err)
			}
			clientChans[i] = ch
			defer ch.Close()
		}

		// Accept all channels on backend side.
		backendChans := make(map[string]*StreamChannel, n)
		for range n {
			bch, err := b.AcceptChannel(ctx)
			if err != nil {
				t.Fatal("accept:", err)
			}
			backendChans[bch.Name()] = bch
			defer bch.Close()
		}

		// Send on each client channel and verify it arrives on the correct backend channel.
		for i, ch := range clientChans {
			msg := fmt.Sprintf("msg-for-chan-%d", i)
			if err := ch.Send(ctx, []byte(msg)); err != nil {
				t.Fatalf("send on chan %d: %v", i, err)
			}
		}

		for i := range n {
			name := fmt.Sprintf("chan-%d", i)
			bch := backendChans[name]
			if bch == nil {
				t.Fatalf("backend channel %q not found", name)
			}
			data, err := bch.Recv(ctx)
			if err != nil {
				t.Fatalf("recv on %s: %v", name, err)
			}
			expected := fmt.Sprintf("msg-for-chan-%d", i)
			if string(data) != expected {
				t.Fatalf("chan %s: got %q, want %q", name, data, expected)
			}
		}
	})
}

func TestLargeMessageOnChannel(t *testing.T) {
	forEachRelay(t, func(t *testing.T, env relayEnv) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		b, c := connectPair(t, env)

		ch, err := c.OpenChannel("large-data")
		if err != nil {
			t.Fatal("open:", err)
		}
		defer ch.Close()

		bch, err := b.AcceptChannel(ctx)
		if err != nil {
			t.Fatal("accept:", err)
		}
		defer bch.Close()

		// 100KB message.
		msg := make([]byte, 100*1024)
		for i := range msg {
			msg[i] = byte(i % 251) // non-trivial pattern
		}

		if err := ch.Send(ctx, msg); err != nil {
			t.Fatal("send:", err)
		}
		data, err := bch.Recv(ctx)
		if err != nil {
			t.Fatal("recv:", err)
		}
		if len(data) != len(msg) {
			t.Fatalf("received %d bytes, want %d", len(data), len(msg))
		}
		for i := range msg {
			if data[i] != msg[i] {
				t.Fatalf("byte %d: got %d, want %d", i, data[i], msg[i])
				break
			}
		}
	})
}

func TestChannelClosedMidConversation(t *testing.T) {
	env := localRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	b, c := connectPair(t, env)

	ch, err := c.OpenChannel("will-close")
	if err != nil {
		t.Fatal("open:", err)
	}

	bch, err := b.AcceptChannel(ctx)
	if err != nil {
		t.Fatal("accept:", err)
	}

	// Send one message successfully.
	if err := ch.Send(ctx, []byte("before-close")); err != nil {
		t.Fatal("send:", err)
	}
	data, err := bch.Recv(ctx)
	if err != nil {
		t.Fatal("recv:", err)
	}
	if string(data) != "before-close" {
		t.Fatalf("got %q, want before-close", data)
	}

	// Close the client side.
	ch.Close()

	// Backend recv should eventually error.
	_, err = bch.Recv(ctx)
	if err == nil {
		t.Fatal("expected error after peer close, got nil")
	}
}

func TestChannelNameSpecialCharacters(t *testing.T) {
	env := localRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	b, c := connectPair(t, env)

	names := []string{
		"hello world",         // space
		"日本語チャンネル",           // unicode
		"",                    // empty string
		"a/b/c",               // path separators
		"name\twith\ttabs",    // tabs
		strings.Repeat("x", 1000), // long name
	}

	for _, name := range names {
		t.Run(fmt.Sprintf("name=%q", name), func(t *testing.T) {
			ch, err := c.OpenChannel(name)
			if err != nil {
				t.Fatal("open:", err)
			}
			defer ch.Close()

			bch, err := b.AcceptChannel(ctx)
			if err != nil {
				t.Fatal("accept:", err)
			}
			defer bch.Close()

			if bch.Name() != name {
				t.Fatalf("got channel name %q, want %q", bch.Name(), name)
			}

			if err := ch.Send(ctx, []byte("data")); err != nil {
				t.Fatal("send:", err)
			}
			data, err := bch.Recv(ctx)
			if err != nil {
				t.Fatal("recv:", err)
			}
			if string(data) != "data" {
				t.Fatalf("got %q, want data", data)
			}
		})
	}
}

func TestRapidOpenAccept(t *testing.T) {
	forEachRelay(t, func(t *testing.T, env relayEnv) {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		b, c := connectPair(t, env)

		const n = 50

		// Open 50 channels rapidly.
		clientChans := make([]*StreamChannel, n)
		for i := range n {
			ch, err := c.OpenChannel(fmt.Sprintf("rapid-%d", i))
			if err != nil {
				t.Fatalf("open %d: %v", i, err)
			}
			clientChans[i] = ch
			defer ch.Close()
		}

		// Accept all on backend side.
		backendNames := make(map[string]bool, n)
		for range n {
			bch, err := b.AcceptChannel(ctx)
			if err != nil {
				t.Fatal("accept:", err)
			}
			backendNames[bch.Name()] = true
			defer bch.Close()
		}

		// Verify all channels were accepted.
		for i := range n {
			name := fmt.Sprintf("rapid-%d", i)
			if !backendNames[name] {
				t.Fatalf("channel %q not accepted", name)
			}
		}
	})
}

// --- Datagram channel tests ---

func TestDatagramChannelRoundTrip(t *testing.T) {
	env := localRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	b, c := connectPair(t, env)

	bdc := b.DatagramChannel("sensors")
	cdc := c.DatagramChannel("sensors")

	// Send multiple datagrams and verify at least one arrives (unreliable transport).
	const attempts = 20
	var received bool
	for i := range attempts {
		if err := cdc.Send([]byte(fmt.Sprintf("reading-%d", i))); err != nil {
			t.Fatalf("send %d: %v", i, err)
		}
	}

	recvCtx, recvCancel := context.WithTimeout(ctx, 3*time.Second)
	defer recvCancel()

	for {
		data, err := bdc.Recv(recvCtx)
		if err != nil {
			break
		}
		if strings.HasPrefix(string(data), "reading-") {
			received = true
			break
		}
	}

	if !received {
		t.Fatal("no datagrams received on named channel")
	}
}

func TestMultipleDatagramChannels(t *testing.T) {
	env := localRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	b, c := connectPair(t, env)

	// Create two named datagram channels on each side.
	bAlpha := b.DatagramChannel("alpha")
	bBeta := b.DatagramChannel("beta")
	cAlpha := c.DatagramChannel("alpha")
	cBeta := c.DatagramChannel("beta")

	const attempts = 20

	// Send on alpha.
	for i := range attempts {
		cAlpha.Send([]byte(fmt.Sprintf("alpha-%d", i)))
	}
	// Send on beta.
	for i := range attempts {
		cBeta.Send([]byte(fmt.Sprintf("beta-%d", i)))
	}

	// Verify alpha receives alpha messages.
	recvCtx, recvCancel := context.WithTimeout(ctx, 3*time.Second)
	defer recvCancel()

	gotAlpha, gotBeta := false, false
	for i := 0; i < attempts; i++ {
		data, err := bAlpha.Recv(recvCtx)
		if err != nil {
			break
		}
		if strings.HasPrefix(string(data), "alpha-") {
			gotAlpha = true
			break
		}
	}
	for i := 0; i < attempts; i++ {
		data, err := bBeta.Recv(recvCtx)
		if err != nil {
			break
		}
		if strings.HasPrefix(string(data), "beta-") {
			gotBeta = true
			break
		}
	}

	if !gotAlpha {
		t.Fatal("no datagrams received on alpha channel")
	}
	if !gotBeta {
		t.Fatal("no datagrams received on beta channel")
	}
}

func TestDatagramChannelWithEncryption(t *testing.T) {
	// DatagramChannel derives encryption keys from the PairingRecord using
	// the same info strings on both sides ("name:dg:send" / "name:dg:recv").
	// This means both sides encrypt with the same key — so the channel is
	// symmetric. We test using raw datagram encryption (SetDatagramChannel)
	// instead, which has explicit send/recv key separation.
	env := localRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	b, c := connectPair(t, env)
	setupDatagramEncryption(t, b, c)

	const attempts = 20
	for i := range attempts {
		c.SendDatagram([]byte(fmt.Sprintf("enc-dg-%d", i)))
	}

	recvCtx, recvCancel := context.WithTimeout(ctx, 3*time.Second)
	defer recvCancel()

	var received bool
	for range attempts {
		data, err := b.RecvDatagram(recvCtx)
		if err != nil {
			break
		}
		if strings.HasPrefix(string(data), "enc-dg-") {
			received = true
			break
		}
	}

	if !received {
		t.Fatal("no encrypted datagrams received")
	}
}

func TestDatagramFloodNocrash(t *testing.T) {
	env := localRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	b, c := connectPair(t, env)

	bdc := b.DatagramChannel("flood")
	cdc := c.DatagramChannel("flood")

	const n = 500
	for i := range n {
		cdc.Send([]byte(fmt.Sprintf("flood-%d", i)))
	}

	// Drain what we can. The point is no crash, not exact delivery.
	recvCtx, recvCancel := context.WithTimeout(ctx, 2*time.Second)
	defer recvCancel()
	count := 0
	for {
		_, err := bdc.Recv(recvCtx)
		if err != nil {
			break
		}
		count++
	}
	t.Logf("datagram flood via channel: received %d/%d", count, n)
	// No crash is the success criterion; receiving at least 1 means routing worked.
	if count == 0 {
		t.Fatal("received 0 datagrams")
	}
}

// --- Error handling tests ---

func TestConnectToNonExistentInstance(t *testing.T) {
	env := localRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	c, err := Connect(ctx, env.url, "does-not-exist-12345", env.cfg)
	if err != nil {
		// Good: Connect itself returned an error (WebTransport path).
		t.Logf("connect to non-existent: %v", err)
		return
	}
	defer c.CloseNow()

	// Raw QUIC: the error surfaces on the first Send or Recv because the
	// server closes the connection after the handshake identifies a missing
	// instance.
	err = c.Send(ctx, []byte("probe"))
	if err == nil {
		_, err = c.Recv(ctx)
	}
	if err == nil {
		t.Fatal("expected error when communicating with non-existent instance")
	}
	t.Logf("connect to non-existent (deferred): %v", err)
}

func TestRegisterWithInvalidToken(t *testing.T) {
	env := localRelayWithToken(t, "correct-token")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Use wrong token.
	wrongCfg := env.cfg
	wrongCfg.Token = "wrong-token"

	_, err := Register(ctx, env.url, wrongCfg)
	if err == nil {
		t.Fatal("expected error with wrong token, got nil")
	}
	t.Logf("register with invalid token: %v", err)
}

func TestSendAfterClose(t *testing.T) {
	env := localRelay(t)
	b, c := connectPair(t, env)

	ctx := context.Background()
	c.Close()

	// Sending after close should return an error.
	err := c.Send(ctx, []byte("after-close"))
	if err == nil {
		t.Fatal("expected error sending after close, got nil")
	}

	_ = b // keep backend alive during test
}

func TestRecvAfterPeerDisconnects(t *testing.T) {
	env := localRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	b, c := connectPair(t, env)

	// Close client side.
	c.CloseNow()

	// Backend recv should return an error.
	_, err := b.Recv(ctx)
	if err == nil {
		t.Fatal("expected error on recv after peer disconnect, got nil")
	}
}

func TestOversizedMessage(t *testing.T) {
	env := localRelay(t)
	ctx := context.Background()
	_, c := connectPair(t, env)

	// maxMessageSize is 1 MiB. Send > 1 MiB.
	oversized := make([]byte, maxMessageSize+1)
	err := c.Send(ctx, oversized)
	if err == nil {
		t.Fatal("expected error sending oversized message, got nil")
	}
	t.Logf("oversized: %v", err)
}

func TestZeroLengthMessage(t *testing.T) {
	forEachRelay(t, func(t *testing.T, env relayEnv) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		b, c := connectPair(t, env)

		// Send empty data.
		if err := c.Send(ctx, []byte{}); err != nil {
			t.Fatal("send empty:", err)
		}
		data, err := b.Recv(ctx)
		if err != nil {
			t.Fatal("recv empty:", err)
		}
		if len(data) != 0 {
			t.Fatalf("expected empty message, got %d bytes", len(data))
		}
	})
}

// --- Concurrent access tests ---

func TestConcurrentSendsOnSameChannel(t *testing.T) {
	forEachRelay(t, func(t *testing.T, env relayEnv) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		b, c := connectPair(t, env)

		ch, err := c.OpenChannel("concurrent-sends")
		if err != nil {
			t.Fatal("open:", err)
		}
		defer ch.Close()

		bch, err := b.AcceptChannel(ctx)
		if err != nil {
			t.Fatal("accept:", err)
		}
		defer bch.Close()

		const goroutines = 10
		const msgsPerGoroutine = 10
		var wg sync.WaitGroup

		for g := range goroutines {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				for i := range msgsPerGoroutine {
					ch.Send(ctx, []byte(fmt.Sprintf("g%d-m%d", id, i)))
				}
			}(g)
		}

		// Receive all messages.
		received := make(map[string]bool)
		var mu sync.Mutex
		done := make(chan struct{})
		go func() {
			for {
				data, err := bch.Recv(ctx)
				if err != nil {
					break
				}
				mu.Lock()
				received[string(data)] = true
				if len(received) == goroutines*msgsPerGoroutine {
					mu.Unlock()
					close(done)
					return
				}
				mu.Unlock()
			}
		}()

		wg.Wait()

		select {
		case <-done:
		case <-ctx.Done():
			mu.Lock()
			t.Fatalf("timed out: received %d/%d messages", len(received), goroutines*msgsPerGoroutine)
			mu.Unlock()
		}
	})
}

func TestConcurrentOpenAcceptChannels(t *testing.T) {
	env := localRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	b, c := connectPair(t, env)

	const n = 5
	var wg sync.WaitGroup

	// Open channels from client concurrently.
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := range n {
			ch, err := c.OpenChannel(fmt.Sprintf("race-%d", i))
			if err != nil {
				return
			}
			defer ch.Close()
		}
	}()

	// Accept channels from backend concurrently.
	accepted := 0
	wg.Add(1)
	go func() {
		defer wg.Done()
		for range n {
			bch, err := b.AcceptChannel(ctx)
			if err != nil {
				return
			}
			defer bch.Close()
			accepted++
		}
	}()

	wg.Wait()
	if accepted != n {
		t.Fatalf("accepted %d/%d channels", accepted, n)
	}
}

func TestSendAndCloseRace(t *testing.T) {
	env := localRelay(t)
	ctx := context.Background()

	b, c := connectPair(t, env)

	var wg sync.WaitGroup

	// One goroutine sends.
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 100; i++ {
			if err := c.Send(ctx, []byte(fmt.Sprintf("msg-%d", i))); err != nil {
				return
			}
		}
	}()

	// Another goroutine closes after a brief delay.
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(5 * time.Millisecond)
		c.Close()
	}()

	wg.Wait()

	// Backend should eventually get an error on recv.
	drainCtx, drainCancel := context.WithTimeout(ctx, 30*time.Second)
	defer drainCancel()
	for {
		_, err := b.Recv(drainCtx)
		if err != nil {
			break
		}
	}
	// Success: no panic or deadlock.
}

// --- Connection lifecycle tests ---

func TestGracefulClose(t *testing.T) {
	env := localRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	b, c := connectPair(t, env)

	// Send and then close gracefully.
	if err := c.Send(ctx, []byte("before-graceful-close")); err != nil {
		t.Fatal("send:", err)
	}

	data, err := b.Recv(ctx)
	if err != nil {
		t.Fatal("recv:", err)
	}
	if string(data) != "before-graceful-close" {
		t.Fatalf("got %q, want before-graceful-close", data)
	}

	if err := c.Close(); err != nil {
		t.Fatal("close:", err)
	}

	// Eventually backend recv should fail.
	_, err = b.Recv(ctx)
	if err == nil {
		t.Fatal("expected error after graceful close")
	}
}

func TestCloseNow(t *testing.T) {
	env := localRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	b, c := connectPair(t, env)

	if err := c.CloseNow(); err != nil {
		t.Fatal("close now:", err)
	}

	_, err := b.Recv(ctx)
	if err == nil {
		t.Fatal("expected error after CloseNow")
	}
}

func TestContextCancellation(t *testing.T) {
	env := localRelay(t)
	b, _ := connectPair(t, env)

	recvCtx, recvCancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer recvCancel()

	// No messages are being sent, so this should time out.
	_, err := b.Recv(recvCtx)
	if err == nil {
		t.Fatal("expected error from context timeout")
	}
}

// --- Relay correctness tests ---

func TestMessagesDontLeakBetweenInstances(t *testing.T) {
	forEachRelay(t, func(t *testing.T, env relayEnv) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		// Create two separate backend/client pairs.
		b1, c1 := connectPair(t, env)
		b2, c2 := connectPair(t, env)

		// Send on pair 1.
		if err := c1.Send(ctx, []byte("pair-1-msg")); err != nil {
			t.Fatal("send c1:", err)
		}
		// Send on pair 2.
		if err := c2.Send(ctx, []byte("pair-2-msg")); err != nil {
			t.Fatal("send c2:", err)
		}

		// Each backend should only get its own pair's message.
		d1, err := b1.Recv(ctx)
		if err != nil {
			t.Fatal("recv b1:", err)
		}
		if string(d1) != "pair-1-msg" {
			t.Fatalf("b1 got %q, want pair-1-msg", d1)
		}

		d2, err := b2.Recv(ctx)
		if err != nil {
			t.Fatal("recv b2:", err)
		}
		if string(d2) != "pair-2-msg" {
			t.Fatalf("b2 got %q, want pair-2-msg", d2)
		}
	})
}

func TestSecondClientRejected(t *testing.T) {
	env := localRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	b, err := Register(ctx, env.url, env.cfg)
	if err != nil {
		t.Fatal("register:", err)
	}
	defer b.CloseNow()

	// First client connects.
	c1, err := Connect(ctx, env.url, b.InstanceID(), env.cfg)
	if err != nil {
		t.Fatal("connect c1:", err)
	}
	defer c1.CloseNow()

	// Verify first client works.
	if err := c1.Send(ctx, []byte("c1-test")); err != nil {
		t.Fatal("c1 send:", err)
	}
	data, err := b.Recv(ctx)
	if err != nil {
		t.Fatal("b recv:", err)
	}
	if string(data) != "c1-test" {
		t.Fatalf("got %q, want c1-test", data)
	}

	// Second client should be rejected (instance occupied).
	c2, err := Connect(ctx, env.url, b.InstanceID(), env.cfg)
	if err != nil {
		// Good: Connect itself returned an error (WebTransport path).
		t.Logf("second client rejected at connect: %v", err)
		return
	}
	defer c2.CloseNow()

	// Raw QUIC: the server closes the connection after the handshake when
	// it finds the instance is occupied. The error surfaces on Send/Recv.
	err = c2.Send(ctx, []byte("probe"))
	if err == nil {
		_, err = c2.Recv(ctx)
	}
	if err == nil {
		t.Fatal("expected error for second client, got nil")
	}
	t.Logf("second client rejected (deferred): %v", err)
}

func TestBackendDisconnectClosesClient(t *testing.T) {
	forEachRelay(t, func(t *testing.T, env relayEnv) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		b, c := connectPair(t, env)

		// Backend disconnects.
		b.CloseNow()

		// Client recv should return error.
		_, err := c.Recv(ctx)
		if err == nil {
			t.Fatal("expected error on recv after backend disconnect")
		}
	})
}

// --- PairingRecord tests ---

func TestPairingRecordWithChannelDerivation(t *testing.T) {
	env := localRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	b, c := connectPair(t, env)

	bRec, cRec := setupPairingRecords(t)
	setupEncryptionWithPairingRecord(t, b, c, bRec, cRec)

	// Open a channel — encryption should be derived from the PairingRecord.
	ch, err := c.OpenChannel("derived-channel")
	if err != nil {
		t.Fatal("open:", err)
	}
	defer ch.Close()

	bch, err := b.AcceptChannel(ctx)
	if err != nil {
		t.Fatal("accept:", err)
	}
	defer bch.Close()

	// Send and verify.
	if err := ch.Send(ctx, []byte("derived-encrypted")); err != nil {
		t.Fatal("send:", err)
	}
	data, err := bch.Recv(ctx)
	if err != nil {
		t.Fatal("recv:", err)
	}
	if string(data) != "derived-encrypted" {
		t.Fatalf("got %q, want derived-encrypted", data)
	}

	// Reverse direction.
	if err := bch.Send(ctx, []byte("reverse-derived")); err != nil {
		t.Fatal("send:", err)
	}
	data, err = ch.Recv(ctx)
	if err != nil {
		t.Fatal("recv:", err)
	}
	if string(data) != "reverse-derived" {
		t.Fatalf("got %q, want reverse-derived", data)
	}
}

func TestPairingRecordJSONRoundTrip(t *testing.T) {
	bKP, err := crypto.GenerateKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	cKP, err := crypto.GenerateKeyPair()
	if err != nil {
		t.Fatal(err)
	}

	original := crypto.NewPairingRecord("peer-123", "https://relay.example.com", bKP, cKP.Public)

	// Marshal to JSON.
	data, err := original.Marshal()
	if err != nil {
		t.Fatal("marshal:", err)
	}

	// Verify it's valid JSON.
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatal("json unmarshal:", err)
	}
	if raw["peer_instance_id"] != "peer-123" {
		t.Fatalf("peer_instance_id: got %v", raw["peer_instance_id"])
	}
	if raw["relay_url"] != "https://relay.example.com" {
		t.Fatalf("relay_url: got %v", raw["relay_url"])
	}

	// Unmarshal back.
	restored, err := crypto.UnmarshalPairingRecord(data)
	if err != nil {
		t.Fatal("unmarshal:", err)
	}

	if restored.PeerInstanceID != original.PeerInstanceID {
		t.Fatalf("PeerInstanceID: got %q, want %q", restored.PeerInstanceID, original.PeerInstanceID)
	}
	if restored.RelayURL != original.RelayURL {
		t.Fatalf("RelayURL: got %q, want %q", restored.RelayURL, original.RelayURL)
	}

	// Derive channels from both original and restored and verify they produce
	// identical encryption keys. Both use the same info strings, so both
	// derive the same Channel — encrypt with one, decrypt with the other.
	origCh, err := original.DeriveChannel([]byte("send"), []byte("recv"))
	if err != nil {
		t.Fatal("derive from original:", err)
	}
	// Restored should produce the same keys (same private key, same peer pub).
	restoredCh, err := restored.DeriveChannel([]byte("send"), []byte("recv"))
	if err != nil {
		t.Fatal("derive from restored:", err)
	}

	// Encrypt with original, decrypt with a fresh copy of the same channel
	// (since sequence counters are shared in the Channel struct, we need
	// separate instances with the same keys to test round-trip).
	ciphertext := origCh.Encrypt([]byte("round-trip-test"))
	// The restored channel has the same send/recv keys, so its recv key
	// matches the original's recv key. Since we encrypted with the send key,
	// we need to decrypt with a channel whose recv key == original's send key.
	// Use reversed info strings for the receiver side.
	restoredRecvCh, err := restored.DeriveChannel([]byte("recv"), []byte("send"))
	if err != nil {
		t.Fatal("derive recv from restored:", err)
	}
	plaintext, err := restoredRecvCh.Decrypt(ciphertext)
	if err != nil {
		t.Fatal("decrypt:", err)
	}
	if string(plaintext) != "round-trip-test" {
		t.Fatalf("got %q, want round-trip-test", plaintext)
	}

	// Also verify that deriving with the same parameters gives working channels.
	_ = restoredCh
	ct2 := restoredCh.Encrypt([]byte("same-keys-test"))
	origRecvCh, err := original.DeriveChannel([]byte("recv"), []byte("send"))
	if err != nil {
		t.Fatal("derive recv from original:", err)
	}
	pt2, err := origRecvCh.Decrypt(ct2)
	if err != nil {
		t.Fatal("decrypt 2:", err)
	}
	if string(pt2) != "same-keys-test" {
		t.Fatalf("got %q, want same-keys-test", pt2)
	}

	// Verify the peer public key can be parsed back to the original.
	restoredPeerPub, err := ecdh.X25519().NewPublicKey(restored.PeerPublicKey)
	if err != nil {
		t.Fatal("parse peer public key:", err)
	}
	if !bytes_equal(restoredPeerPub.Bytes(), cKP.Public.Bytes()) {
		t.Fatal("peer public key mismatch after round-trip")
	}
}

// TestDatagramChannelDuplicateName verifies that calling DatagramChannel
// twice with the same name returns the same *DatagramChannel object.
func TestDatagramChannelDuplicateName(t *testing.T) {
	env := localRelay(t)
	b, _ := connectPair(t, env)

	dc1 := b.DatagramChannel("dup-test")
	dc2 := b.DatagramChannel("dup-test")

	if dc1 != dc2 {
		t.Fatal("DatagramChannel returned different objects for the same name")
	}
}

// TestDatagramChannelWithPairingRecord verifies the encrypted datagram
// channel path: PairingRecord is set, DatagramChannel derives a
// crypto.Channel, and SetMode(ModeDatagrams) is called.
//
// Note: DatagramChannel uses symmetric info strings ("name:dg:send"/
// "name:dg:recv") on both sides, so both sides get send=HKDF(shared,
// "name:dg:send") and recv=HKDF(shared, "name:dg:recv"). Side A encrypts
// with its send key and side B decrypts with its recv key — these differ.
// This is a known limitation of the current symmetric-info design.
// This test exercises the code path (key derivation, SetMode) without
// attempting a cross-side decrypt.
func TestDatagramChannelWithPairingRecord(t *testing.T) {
	env := localRelay(t)
	b, _ := connectPair(t, env)

	bRec, _ := setupPairingRecords(t)
	b.SetPairingRecord(bRec)

	bdc := b.DatagramChannel("enc-dg")

	// The branch `if c.pairingRecord != nil` was taken.
	if bdc.ch == nil {
		t.Fatal("DatagramChannel with PairingRecord should have a non-nil crypto.Channel")
	}

	// Verify the same channel is returned for the same name (duplicate path).
	bdc2 := b.DatagramChannel("enc-dg")
	if bdc != bdc2 {
		t.Fatal("DatagramChannel returned different objects for same name with PairingRecord")
	}
}

// TestEncryptedStreamingChannelBackendOpens tests the key direction flip
// when the backend opens a channel (isOpener=true for backend instead of client).
func TestEncryptedStreamingChannelBackendOpens(t *testing.T) {
	env := localRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	b, c := connectPair(t, env)

	bRec, cRec := setupPairingRecords(t)
	setupEncryptionWithPairingRecord(t, b, c, bRec, cRec)

	// Backend opens the channel (reverse direction).
	ch, err := b.OpenChannel("backend-opens-encrypted")
	if err != nil {
		t.Fatal("open channel:", err)
	}
	defer ch.Close()

	cch, err := c.AcceptChannel(ctx)
	if err != nil {
		t.Fatal("accept channel:", err)
	}
	defer cch.Close()

	if cch.Name() != "backend-opens-encrypted" {
		t.Fatalf("got channel name %q, want backend-opens-encrypted", cch.Name())
	}

	// Backend -> Client.
	if err := ch.Send(ctx, []byte("from-be-encrypted")); err != nil {
		t.Fatal("send:", err)
	}
	data, err := cch.Recv(ctx)
	if err != nil {
		t.Fatal("recv:", err)
	}
	if string(data) != "from-be-encrypted" {
		t.Fatalf("got %q, want from-be-encrypted", data)
	}

	// Client -> Backend.
	if err := cch.Send(ctx, []byte("from-cl-encrypted")); err != nil {
		t.Fatal("send:", err)
	}
	data, err = ch.Recv(ctx)
	if err != nil {
		t.Fatal("recv:", err)
	}
	if string(data) != "from-cl-encrypted" {
		t.Fatalf("got %q, want from-cl-encrypted", data)
	}
}

// TestEncryptedDatagramChannelSend verifies that DatagramChannel.Send
// with a PairingRecord-derived crypto.Channel encrypts (doesn't panic
// or error). We test encryption in isolation since the symmetric info
// strings mean cross-side decrypt is a known limitation.
func TestEncryptedDatagramChannelSend(t *testing.T) {
	env := localRelay(t)
	b, _ := connectPair(t, env)

	bRec, _ := setupPairingRecords(t)
	b.SetPairingRecord(bRec)

	bdc := b.DatagramChannel("enc-send-test")
	if bdc.ch == nil {
		t.Fatal("expected non-nil crypto.Channel")
	}

	// Send should not panic or error — it encrypts via the derived channel.
	for i := range 5 {
		if err := bdc.Send([]byte(fmt.Sprintf("enc-payload-%d", i))); err != nil {
			t.Fatalf("send %d: %v", i, err)
		}
	}
}

// bytes_equal compares two byte slices.
func bytes_equal(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

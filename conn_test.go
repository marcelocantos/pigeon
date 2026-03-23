// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package tern

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"

	"github.com/marcelocantos/tern/crypto"
)

// startTestRelay starts a minimal relay server for testing.
func startTestRelay(t *testing.T) string {
	t.Helper()

	type inst struct {
		conn *websocket.Conn
		ctx  context.Context
	}
	var backend *inst

	mux := http.NewServeMux()
	mux.HandleFunc("GET /register", func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		defer conn.CloseNow()
		conn.Write(r.Context(), websocket.MessageText, []byte("test-id"))
		backend = &inst{conn: conn, ctx: r.Context()}
		<-r.Context().Done()
	})
	mux.HandleFunc("GET /ws/{id}", func(w http.ResponseWriter, r *http.Request) {
		clientConn, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		defer clientConn.CloseNow()
		if backend == nil {
			return
		}
		go func() {
			for {
				mt, data, err := backend.conn.Read(backend.ctx)
				if err != nil {
					return
				}
				clientConn.Write(r.Context(), mt, data)
			}
		}()
		for {
			mt, data, err := clientConn.Read(r.Context())
			if err != nil {
				return
			}
			backend.conn.Write(backend.ctx, mt, data)
		}
	})

	ts := httptest.NewServer(mux)
	t.Cleanup(ts.Close)
	return "ws" + strings.TrimPrefix(ts.URL, "http")
}

func TestRawModeRoundTrip(t *testing.T) {
	wsBase := startTestRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	b, err := Register(ctx, wsBase)
	if err != nil {
		t.Fatal(err)
	}
	defer b.CloseNow()

	c, err := Connect(ctx, wsBase, b.InstanceID())
	if err != nil {
		t.Fatal(err)
	}
	defer c.CloseNow()

	// Client → backend (raw mode).
	if err := c.Send(ctx, websocket.MessageText, []byte("hello")); err != nil {
		t.Fatal(err)
	}
	mt, data, err := b.Recv(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if mt != websocket.MessageText || string(data) != "hello" {
		t.Fatalf("got (%v, %q), want (text, hello)", mt, data)
	}
}

func TestEncryptedModeRoundTrip(t *testing.T) {
	wsBase := startTestRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	b, err := Register(ctx, wsBase)
	if err != nil {
		t.Fatal(err)
	}
	defer b.CloseNow()

	c, err := Connect(ctx, wsBase, b.InstanceID())
	if err != nil {
		t.Fatal(err)
	}
	defer c.CloseNow()

	// Set up matching encrypted channels.
	bKP, _ := crypto.GenerateKeyPair()
	cKP, _ := crypto.GenerateKeyPair()

	bSendKey, _ := crypto.DeriveSessionKey(bKP.Private, cKP.Public, []byte("b-to-c"))
	bRecvKey, _ := crypto.DeriveSessionKey(bKP.Private, cKP.Public, []byte("c-to-b"))
	cSendKey, _ := crypto.DeriveSessionKey(cKP.Private, bKP.Public, []byte("c-to-b"))
	cRecvKey, _ := crypto.DeriveSessionKey(cKP.Private, bKP.Public, []byte("b-to-c"))

	bCh, _ := crypto.NewChannel(bSendKey, bRecvKey)
	cCh, _ := crypto.NewChannel(cSendKey, cRecvKey)

	b.SetChannel(bCh)
	c.SetChannel(cCh)

	// Give LAN discovery a moment to run. On localhost, the LAN upgrade
	// may succeed — that's fine, messages flow transparently either way.
	time.Sleep(100 * time.Millisecond)

	// Client → backend (encrypted mode — caller sends plaintext).
	if err := c.Send(ctx, websocket.MessageBinary, []byte("secret")); err != nil {
		t.Fatal(err)
	}

	mt, data, err := b.Recv(ctx)
	if err != nil {
		t.Fatalf("recv: %v", err)
	}
	if string(data) != "secret" {
		t.Fatalf("got %q, want secret", data)
	}
	if mt != websocket.MessageBinary {
		t.Fatalf("got mt=%v, want binary", mt)
	}

	// Backend → client.
	if err := b.Send(ctx, websocket.MessageBinary, []byte("reply")); err != nil {
		t.Fatal(err)
	}

	mt, data, err = c.Recv(ctx)
	if err != nil {
		t.Fatalf("recv: %v", err)
	}
	if string(data) != "reply" {
		t.Fatalf("got %q, want reply", data)
	}
}

func TestLANUpgrade(t *testing.T) {
	wsBase := startTestRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	b, err := Register(ctx, wsBase)
	if err != nil {
		t.Fatal(err)
	}
	defer b.CloseNow()

	c, err := Connect(ctx, wsBase, b.InstanceID())
	if err != nil {
		t.Fatal(err)
	}
	defer c.CloseNow()

	// Set up matching encrypted channels.
	bKP, _ := crypto.GenerateKeyPair()
	cKP, _ := crypto.GenerateKeyPair()

	bSendKey, _ := crypto.DeriveSessionKey(bKP.Private, cKP.Public, []byte("b-to-c"))
	bRecvKey, _ := crypto.DeriveSessionKey(bKP.Private, cKP.Public, []byte("c-to-b"))
	cSendKey, _ := crypto.DeriveSessionKey(cKP.Private, bKP.Public, []byte("c-to-b"))
	cRecvKey, _ := crypto.DeriveSessionKey(cKP.Private, bKP.Public, []byte("b-to-c"))

	bCh, _ := crypto.NewChannel(bSendKey, bRecvKey)
	cCh, _ := crypto.NewChannel(cSendKey, cRecvKey)

	// Both sides start on the relay.
	if got := b.PreferredTransport(); got != "relay" {
		t.Fatalf("backend initial transport: got %q, want relay", got)
	}
	if got := c.PreferredTransport(); got != "relay" {
		t.Fatalf("client initial transport: got %q, want relay", got)
	}

	// Enable encryption. This triggers LAN discovery automatically:
	// both sides start listeners, exchange offers, and connect directly.
	b.SetChannel(bCh)
	c.SetChannel(cCh)

	// Wait for LAN upgrade to complete. Both sides need to discover
	// each other's addresses, connect, and cut over.
	deadline := time.After(5 * time.Second)
	tick := time.NewTicker(50 * time.Millisecond)
	defer tick.Stop()
	for {
		select {
		case <-deadline:
			t.Fatalf("LAN upgrade did not complete: backend=%q, client=%q",
				b.PreferredTransport(), c.PreferredTransport())
		case <-tick.C:
			if b.PreferredTransport() == "lan" && c.PreferredTransport() == "lan" {
				goto upgraded
			}
		}
	}
upgraded:

	// Now send messages — they should flow through the LAN transport.
	if err := c.Send(ctx, websocket.MessageBinary, []byte("via LAN")); err != nil {
		t.Fatal(err)
	}
	mt, data, err := b.Recv(ctx)
	if err != nil {
		t.Fatalf("recv: %v", err)
	}
	if string(data) != "via LAN" {
		t.Fatalf("got %q, want 'via LAN'", data)
	}
	if mt != websocket.MessageBinary {
		t.Fatalf("got mt=%v, want binary", mt)
	}

	// Reverse direction.
	if err := b.Send(ctx, websocket.MessageBinary, []byte("reply via LAN")); err != nil {
		t.Fatal(err)
	}
	_, data, err = c.Recv(ctx)
	if err != nil {
		t.Fatalf("recv: %v", err)
	}
	if string(data) != "reply via LAN" {
		t.Fatalf("got %q, want 'reply via LAN'", data)
	}

	// Verify both sides are using LAN.
	if got := b.PreferredTransport(); got != "lan" {
		t.Fatalf("backend transport after messaging: got %q, want lan", got)
	}
	if got := c.PreferredTransport(); got != "lan" {
		t.Fatalf("client transport after messaging: got %q, want lan", got)
	}
}

func TestEncryptedModeFiltersControlMessages(t *testing.T) {
	wsBase := startTestRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	b, err := Register(ctx, wsBase)
	if err != nil {
		t.Fatal(err)
	}
	defer b.CloseNow()

	c, err := Connect(ctx, wsBase, b.InstanceID())
	if err != nil {
		t.Fatal(err)
	}
	defer c.CloseNow()

	// Set up encrypted channels.
	bKP, _ := crypto.GenerateKeyPair()
	cKP, _ := crypto.GenerateKeyPair()

	bSendKey, _ := crypto.DeriveSessionKey(bKP.Private, cKP.Public, []byte("b-to-c"))
	bRecvKey, _ := crypto.DeriveSessionKey(bKP.Private, cKP.Public, []byte("c-to-b"))
	cSendKey, _ := crypto.DeriveSessionKey(cKP.Private, bKP.Public, []byte("c-to-b"))
	cRecvKey, _ := crypto.DeriveSessionKey(cKP.Private, bKP.Public, []byte("b-to-c"))

	bCh, _ := crypto.NewChannel(bSendKey, bRecvKey)
	cCh, _ := crypto.NewChannel(cSendKey, cRecvKey)

	// Don't call SetChannel yet — we want to control the sequence manually.
	// Instead, we'll encrypt and send a LAN offer control message directly,
	// then an application message, and verify that Recv only returns the
	// application message.

	// Manually encrypt a LAN offer control message.
	lanOffer := append([]byte{msgLANOffer}, []byte(`{"addrs":["10.0.0.1"],"port":"9999"}`)...)
	encOffer := cCh.Encrypt(lanOffer)

	// Then an application message.
	appMsg := append([]byte{msgApp}, []byte("the real message")...)
	encApp := cCh.Encrypt(appMsg)

	// Send both through the relay (raw, since we're encrypting manually).
	if err := c.Send(ctx, websocket.MessageBinary, encOffer); err != nil {
		t.Fatal(err)
	}
	if err := c.Send(ctx, websocket.MessageBinary, encApp); err != nil {
		t.Fatal(err)
	}

	// Set channel on receiver. The LAN offer should be consumed internally;
	// only the application message should be delivered.
	b.SetChannel(bCh)

	_, data, err := b.Recv(ctx)
	if err != nil {
		t.Fatalf("recv: %v", err)
	}
	if string(data) != "the real message" {
		t.Fatalf("got %q, want 'the real message'", data)
	}
}

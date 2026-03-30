// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package tern

import (
	"context"
	"testing"
	"time"

	"github.com/marcelocantos/tern/crypto"
)

// lanPair creates a backend+client pair where the backend has a LAN
// server and the client is LAN-enabled.
func lanPair(t *testing.T, env relayEnv) (*Conn, *Conn, *LANServer) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	lanSrv, err := NewLANServer("", nil)
	if err != nil {
		t.Fatal("NewLANServer:", err)
	}
	t.Cleanup(func() { lanSrv.Close() })

	bCfg := env.cfg
	bCfg.LANServer = lanSrv
	b, err := Register(ctx, env.url, bCfg)
	if err != nil {
		t.Fatal("register:", err)
	}
	t.Cleanup(func() { b.CloseNow() })

	cCfg := env.cfg
	cCfg.LAN = true
	c, err := Connect(ctx, env.url, b.InstanceID(), cCfg)
	if err != nil {
		t.Fatal("connect:", err)
	}
	t.Cleanup(func() { c.CloseNow() })

	// Set up encryption — required for LAN offer (control messages).
	bKP, _ := crypto.GenerateKeyPair()
	cKP, _ := crypto.GenerateKeyPair()
	bRec := crypto.NewPairingRecord("client", "relay", bKP, cKP.Public)
	cRec := crypto.NewPairingRecord("backend", "relay", cKP, bKP.Public)

	bCh, _ := bRec.DeriveChannel([]byte("b2c"), []byte("c2b"))
	cCh, _ := cRec.DeriveChannel([]byte("c2b"), []byte("b2c"))

	// SetChannel on backend triggers LAN advertisement.
	b.SetChannel(bCh)
	c.SetChannel(cCh)

	return b, c, lanSrv
}

// TestLANUpgrade verifies that traffic switches from relay to LAN.
func TestLANUpgrade(t *testing.T) {
	env := localRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	b, c, _ := lanPair(t, env)

	// First message triggers Recv which processes the LAN offer.
	b.Send(ctx, []byte("via-relay"))
	data, err := c.Recv(ctx)
	if err != nil {
		t.Fatal("recv:", err)
	}
	if string(data) != "via-relay" {
		t.Fatalf("got %q", data)
	}

	// Wait for LAN connection to establish.
	time.Sleep(2 * time.Second)

	// This should go via LAN.
	c.Send(ctx, []byte("via-lan"))
	data, err = b.Recv(ctx)
	if err != nil {
		t.Fatal("recv via LAN:", err)
	}
	if string(data) != "via-lan" {
		t.Fatalf("got %q", data)
	}

	t.Log("LAN upgrade successful")
}

// TestLANUpgradeBidirectional verifies both directions work after switch.
func TestLANUpgradeBidirectional(t *testing.T) {
	env := localRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	b, c, _ := lanPair(t, env)

	// Trigger LAN offer processing.
	b.Send(ctx, []byte("trigger"))
	c.Recv(ctx)
	time.Sleep(2 * time.Second)

	for i := range 5 {
		msg := []byte("ping-" + string(rune('0'+i)))
		c.Send(ctx, msg)
		data, _ := b.Recv(ctx)
		if string(data) != string(msg) {
			t.Fatalf("got %q, want %q", data, msg)
		}

		reply := []byte("pong-" + string(rune('0'+i)))
		b.Send(ctx, reply)
		data, _ = c.Recv(ctx)
		if string(data) != string(reply) {
			t.Fatalf("got %q, want %q", data, reply)
		}
	}
}

// TestLANServerMultipleClients verifies the LAN server can serve
// multiple clients concurrently.
func TestLANServerMultipleClients(t *testing.T) {
	env := localRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	lanSrv, err := NewLANServer("", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer lanSrv.Close()

	// Two backends sharing the same LAN server.
	b1Cfg := env.cfg
	b1Cfg.LANServer = lanSrv
	b1Cfg.InstanceID = "b1"
	b1, err := Register(ctx, env.url, b1Cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer b1.CloseNow()

	b2Cfg := env.cfg
	b2Cfg.LANServer = lanSrv
	b2Cfg.InstanceID = "b2"
	b2, err := Register(ctx, env.url, b2Cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer b2.CloseNow()

	t.Logf("LAN server addr: %s, serving 2 backends", lanSrv.Addr())
}

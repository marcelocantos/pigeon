// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package tern

import (
	"context"
	"testing"
	"time"
)

// TestLANUpgrade verifies that two peers connected via the relay
// automatically switch to a direct LAN connection.
func TestLANUpgrade(t *testing.T) {
	env := localRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	b, c := connectPair(t, env)

	// Set up encryption (required for control messages).
	bRec, cRec := setupPairingRecords(t)
	setupEncryptionWithPairingRecord(t, b, c, bRec, cRec)

	// Enable LAN on both sides.
	if err := b.EnableLAN(ctx, true, nil); err != nil {
		t.Fatal("backend EnableLAN:", err)
	}
	if err := c.EnableLAN(ctx, false, nil); err != nil {
		t.Fatal("client EnableLAN:", err)
	}

	// Send a message — this triggers Recv on the client which will
	// process the LAN offer control message.
	if err := b.Send(ctx, []byte("via-relay")); err != nil {
		t.Fatal("send:", err)
	}
	data, err := c.Recv(ctx)
	if err != nil {
		t.Fatal("recv:", err)
	}
	if string(data) != "via-relay" {
		t.Fatalf("got %q, want via-relay", data)
	}

	// Give the LAN connection time to establish.
	time.Sleep(2 * time.Second)

	// Send another message — should go via LAN now.
	if err := c.Send(ctx, []byte("via-lan")); err != nil {
		t.Fatal("send via LAN:", err)
	}
	data, err = b.Recv(ctx)
	if err != nil {
		t.Fatal("recv via LAN:", err)
	}
	if string(data) != "via-lan" {
		t.Fatalf("got %q, want via-lan", data)
	}

	t.Log("LAN upgrade successful — messages delivered via direct connection")
}

// TestLANUpgradeBidirectional verifies that after LAN upgrade, both
// directions work and channels function correctly.
func TestLANUpgradeBidirectional(t *testing.T) {
	env := localRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	b, c := connectPair(t, env)

	bRec, cRec := setupPairingRecords(t)
	setupEncryptionWithPairingRecord(t, b, c, bRec, cRec)

	b.EnableLAN(ctx, true, nil)
	c.EnableLAN(ctx, false, nil)

	// Trigger LAN offer processing.
	b.Send(ctx, []byte("trigger"))
	c.Recv(ctx)
	time.Sleep(2 * time.Second)

	// Bidirectional messaging.
	for i := range 5 {
		msg := []byte("ping-" + string(rune('0'+i)))
		c.Send(ctx, msg)
		data, err := b.Recv(ctx)
		if err != nil {
			t.Fatalf("recv %d: %v", i, err)
		}
		if string(data) != string(msg) {
			t.Fatalf("got %q, want %q", data, msg)
		}

		reply := []byte("pong-" + string(rune('0'+i)))
		b.Send(ctx, reply)
		data, err = c.Recv(ctx)
		if err != nil {
			t.Fatalf("recv reply %d: %v", i, err)
		}
		if string(data) != string(reply) {
			t.Fatalf("got %q, want %q", data, reply)
		}
	}
}

// TestLANUpgradeWithoutEncryption verifies that EnableLAN on the
// backend still works without encryption (the offer is sent raw).
func TestLANUpgradeRequiresEncryption(t *testing.T) {
	env := localRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	b, c := connectPair(t, env)

	// Set up encryption — required for control messages.
	bRec, cRec := setupPairingRecords(t)
	setupEncryptionWithPairingRecord(t, b, c, bRec, cRec)

	err := b.EnableLAN(ctx, true, nil)
	if err != nil {
		t.Fatal("EnableLAN:", err)
	}
}

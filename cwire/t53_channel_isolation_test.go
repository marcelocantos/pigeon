// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package cwire_test

import (
	"bytes"
	"testing"

	"github.com/marcelocantos/pigeon/cwire"
)

// TestT53StreamAndDatagramIndependentCounters is the 🎯T53 oracle:
// streams use ModeStrict and datagrams use ModeDatagrams with distinct
// keys forked from the same SessionMaterial. A gap on the datagram
// channel must not prevent subsequent stream decrypts, and two named
// streams may each use seq 0 under different keys.
func TestT53StreamAndDatagramIndependentCounters(t *testing.T) {
	kp, err := cwire.GenerateKeypair()
	if err != nil {
		t.Fatalf("GenerateKeypair: %v", err)
	}
	peer, err := cwire.GenerateKeypair()
	if err != nil {
		t.Fatalf("GenerateKeypair peer: %v", err)
	}
	nonce := bytes.Repeat([]byte{0xab}, 16)
	client, err := cwire.DeriveSessionMaterial(&cwire.PairingRecord{
		LocalPrivKey: kp.PrivKey,
		PeerPubKey:   peer.PubKey,
	}, false, nonce)
	if err != nil {
		t.Fatalf("client material: %v", err)
	}
	backend, err := cwire.DeriveSessionMaterial(&cwire.PairingRecord{
		LocalPrivKey: peer.PrivKey,
		PeerPubKey:   kp.PubKey,
	}, true, nonce)
	if err != nil {
		t.Fatalf("backend material: %v", err)
	}

	chatC, err := client.StreamChannel("chat")
	if err != nil {
		t.Fatal(err)
	}
	ctrlC, err := client.StreamChannel("control")
	if err != nil {
		t.Fatal(err)
	}
	dgC, err := client.DatagramChannel()
	if err != nil {
		t.Fatal(err)
	}
	chatB, err := backend.StreamChannel("chat")
	if err != nil {
		t.Fatal(err)
	}
	ctrlB, err := backend.StreamChannel("control")
	if err != nil {
		t.Fatal(err)
	}
	dgB, err := backend.DatagramChannel()
	if err != nil {
		t.Fatal(err)
	}

	// 1. Concurrent streams: each may start at seq 0 under its own key.
	ctChat, err := chatC.Encrypt([]byte("hello-chat"))
	if err != nil {
		t.Fatal(err)
	}
	ctCtrl, err := ctrlC.Encrypt([]byte("hello-ctrl"))
	if err != nil {
		t.Fatal(err)
	}
	if p, err := chatB.Decrypt(ctChat); err != nil || string(p) != "hello-chat" {
		t.Fatalf("chat decrypt: %q %v", p, err)
	}
	if p, err := ctrlB.Decrypt(ctCtrl); err != nil || string(p) != "hello-ctrl" {
		t.Fatalf("control decrypt: %q %v", p, err)
	}

	// 2. Datagram gap: encrypt seq 0,1,2 on dg; deliver 0 and 2 only.
	d0, err := dgC.Encrypt([]byte("d0"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := dgC.Encrypt([]byte("d1")); err != nil {
		t.Fatal(err)
	}
	d2, err := dgC.Encrypt([]byte("d2"))
	if err != nil {
		t.Fatal(err)
	}
	if p, err := dgB.Decrypt(d0); err != nil || string(p) != "d0" {
		t.Fatalf("dg0: %q %v", p, err)
	}
	if p, err := dgB.Decrypt(d2); err != nil || string(p) != "d2" {
		t.Fatalf("dg2 after gap: %q %v (ModeDatagrams must allow gaps)", p, err)
	}

	// 3. Stream still works after datagram traffic / gap.
	ct2, err := chatC.Encrypt([]byte("still-ok"))
	if err != nil {
		t.Fatal(err)
	}
	if p, err := chatB.Decrypt(ct2); err != nil || string(p) != "still-ok" {
		t.Fatalf("chat after dg gap: %q %v", p, err)
	}
}

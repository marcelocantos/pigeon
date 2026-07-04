// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package cwire_test

import (
	"bytes"
	"testing"

	"github.com/marcelocantos/pigeon/cwire"
)

// TestDeriveSessionChannelPerNonceDistinct is the 🎯T44.1 regression oracle
// (docs/audit/fable-2026-07.md F2): two sessions activated from the SAME
// PairingRecord must derive DISTINCT AEAD keys, because the per-session nonce
// exchanged during activation is folded into the HKDF info. Before the fix,
// DeriveSessionChannel was a deterministic function of the static record and
// every channel restarted the AES-GCM counter at 0 — two ciphertexts under one
// (key, nonce=0) leak plaintext via keystream reuse (P1 XOR P2 = C1 XOR C2) and
// expose the GHASH auth subkey.
func TestDeriveSessionChannelPerNonceDistinct(t *testing.T) {
	localKP, err := cwire.GenerateKeypair()
	if err != nil {
		t.Fatalf("GenerateKeypair (local): %v", err)
	}
	peerKP, err := cwire.GenerateKeypair()
	if err != nil {
		t.Fatalf("GenerateKeypair (peer): %v", err)
	}
	rec := &cwire.PairingRecord{
		PeerInstanceID: "peer-001",
		LocalPrivKey:   localKP.PrivKey,
		LocalPubKey:    localKP.PubKey,
		PeerPubKey:     peerKP.PubKey,
	}

	nonce1 := bytes.Repeat([]byte{0x11}, 16)
	nonce2 := bytes.Repeat([]byte{0x22}, 16)
	plaintext := []byte("the quick brown fox")

	// Same record + same direction, two different per-session nonces.
	ch1, err := cwire.DeriveSessionChannel(rec, false, nonce1)
	if err != nil {
		t.Fatalf("DeriveSessionChannel (nonce1): %v", err)
	}
	ch2, err := cwire.DeriveSessionChannel(rec, false, nonce2)
	if err != nil {
		t.Fatalf("DeriveSessionChannel (nonce2): %v", err)
	}

	// Fresh channels both start at GCM counter 0, so a shared key would
	// produce identical ciphertext bodies for identical plaintext. The
	// leading 8 bytes are the (zero) sequence prefix; the body follows.
	ct1, err := ch1.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt (ch1): %v", err)
	}
	ct2, err := ch2.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt (ch2): %v", err)
	}
	if bytes.Equal(ct1[8:], ct2[8:]) {
		t.Fatal("distinct nonces produced identical keystream: per-session key derivation is broken (AES-GCM nonce reuse across sessions)")
	}

	// Sanity: derivation is deterministic in (record, nonce) so both peers,
	// folding in the same negotiated nonce, agree on the keys.
	ch1b, err := cwire.DeriveSessionChannel(rec, false, nonce1)
	if err != nil {
		t.Fatalf("DeriveSessionChannel (nonce1 again): %v", err)
	}
	ct1b, err := ch1b.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt (ch1b): %v", err)
	}
	if !bytes.Equal(ct1[8:], ct1b[8:]) {
		t.Fatal("same nonce produced different keys: derivation is not deterministic in (record, nonce)")
	}

	// A wrong-length nonce is rejected rather than silently truncated.
	if _, err := cwire.DeriveSessionChannel(rec, false, []byte{0x01}); err == nil {
		t.Fatal("expected error for short nonce, got nil")
	}
}

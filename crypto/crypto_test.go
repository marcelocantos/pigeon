// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package crypto

import (
	"bytes"
	"crypto/ecdh"
	"crypto/rand"
	"fmt"
	"testing"
)

func TestKeyExchangeAndEncrypt(t *testing.T) {
	// Simulate client and server generating key pairs.
	client, err := GenerateKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	server, err := GenerateKeyPair()
	if err != nil {
		t.Fatal(err)
	}

	// Derive session keys — client sends, server receives.
	clientSendKey, err := DeriveSessionKey(client.Private, server.Public, []byte("client-to-server"))
	if err != nil {
		t.Fatal(err)
	}
	serverRecvKey, err := DeriveSessionKey(server.Private, client.Public, []byte("client-to-server"))
	if err != nil {
		t.Fatal(err)
	}

	// Keys should match.
	if !bytes.Equal(clientSendKey, serverRecvKey) {
		t.Fatal("derived keys don't match")
	}

	// Reverse direction.
	clientRecvKey, err := DeriveSessionKey(client.Private, server.Public, []byte("server-to-client"))
	if err != nil {
		t.Fatal(err)
	}
	serverSendKey, err := DeriveSessionKey(server.Private, client.Public, []byte("server-to-client"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(clientRecvKey, serverSendKey) {
		t.Fatal("reverse derived keys don't match")
	}

	// Create channels.
	clientCh, err := NewChannel(clientSendKey, clientRecvKey)
	if err != nil {
		t.Fatal(err)
	}
	serverCh, err := NewChannel(serverSendKey, serverRecvKey)
	if err != nil {
		t.Fatal(err)
	}

	// Client → server.
	msg := []byte("hello from client")
	ct := clientCh.Encrypt(msg)
	pt, err := serverCh.Decrypt(ct)
	if err != nil {
		t.Fatal("server decrypt:", err)
	}
	if !bytes.Equal(pt, msg) {
		t.Fatalf("got %q, want %q", pt, msg)
	}

	// Server → client.
	msg2 := []byte("hello from server")
	ct2 := serverCh.Encrypt(msg2)
	pt2, err := clientCh.Decrypt(ct2)
	if err != nil {
		t.Fatal("client decrypt:", err)
	}
	if !bytes.Equal(pt2, msg2) {
		t.Fatalf("got %q, want %q", pt2, msg2)
	}
}

func TestSymmetricChannel(t *testing.T) {
	secret, err := GenerateSecret()
	if err != nil {
		t.Fatal(err)
	}

	clientCh, err := NewSymmetricChannel(secret, false)
	if err != nil {
		t.Fatal(err)
	}
	serverCh, err := NewSymmetricChannel(secret, true)
	if err != nil {
		t.Fatal(err)
	}

	// Client → server.
	msg := []byte(`{"type":"auth","device":"abc123"}`)
	ct := clientCh.Encrypt(msg)
	pt, err := serverCh.Decrypt(ct)
	if err != nil {
		t.Fatal("server decrypt:", err)
	}
	if !bytes.Equal(pt, msg) {
		t.Fatalf("got %q, want %q", pt, msg)
	}

	// Server → client.
	msg2 := []byte(`{"type":"auth_ok"}`)
	ct2 := serverCh.Encrypt(msg2)
	pt2, err := clientCh.Decrypt(ct2)
	if err != nil {
		t.Fatal("client decrypt:", err)
	}
	if !bytes.Equal(pt2, msg2) {
		t.Fatalf("got %q, want %q", pt2, msg2)
	}

	// Multiple messages in sequence.
	for i := 0; i < 100; i++ {
		msg := []byte("message from client")
		ct := clientCh.Encrypt(msg)
		pt, err := serverCh.Decrypt(ct)
		if err != nil {
			t.Fatalf("round %d: %v", i, err)
		}
		if !bytes.Equal(pt, msg) {
			t.Fatalf("round %d: mismatch", i)
		}
	}
}

func TestReplayRejected(t *testing.T) {
	secret, err := GenerateSecret()
	if err != nil {
		t.Fatal(err)
	}
	clientCh, err := NewSymmetricChannel(secret, false)
	if err != nil {
		t.Fatal(err)
	}
	serverCh, err := NewSymmetricChannel(secret, true)
	if err != nil {
		t.Fatal(err)
	}

	ct := clientCh.Encrypt([]byte("first"))
	if _, err := serverCh.Decrypt(ct); err != nil {
		t.Fatal(err)
	}

	// Replay the same ciphertext — should fail.
	if _, err := serverCh.Decrypt(ct); err == nil {
		t.Fatal("expected replay to be rejected")
	}
}

func TestDeriveKeyFromSecret(t *testing.T) {
	secret, err := GenerateSecret()
	if err != nil {
		t.Fatal(err)
	}
	nonce, err := GenerateNonce()
	if err != nil {
		t.Fatal(err)
	}

	key1, err := DeriveKeyFromSecret(secret, nonce)
	if err != nil {
		t.Fatal(err)
	}
	key2, err := DeriveKeyFromSecret(secret, nonce)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(key1, key2) {
		t.Fatal("same inputs should produce same key")
	}

	// Different nonce → different key.
	nonce2, _ := GenerateNonce()
	key3, err := DeriveKeyFromSecret(secret, nonce2)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(key1, key3) {
		t.Fatal("different nonces should produce different keys")
	}
}

func newTestChannel(t *testing.T) (*Channel, *Channel) {
	t.Helper()
	secret, err := GenerateSecret()
	if err != nil {
		t.Fatal(err)
	}
	clientCh, err := NewSymmetricChannel(secret, false)
	if err != nil {
		t.Fatal(err)
	}
	serverCh, err := NewSymmetricChannel(secret, true)
	if err != nil {
		t.Fatal(err)
	}
	return clientCh, serverCh
}

func newTestDatagramChannel(t *testing.T) (*Channel, *Channel) {
	t.Helper()
	secret, err := GenerateSecret()
	if err != nil {
		t.Fatal(err)
	}

	sendKey, err := hkdfDerive(secret, []byte("client-to-server"))
	if err != nil {
		t.Fatal(err)
	}
	recvKey, err := hkdfDerive(secret, []byte("server-to-client"))
	if err != nil {
		t.Fatal(err)
	}

	clientCh, err := NewDatagramChannel(sendKey, recvKey)
	if err != nil {
		t.Fatal(err)
	}
	serverCh, err := NewDatagramChannel(recvKey, sendKey)
	if err != nil {
		t.Fatal(err)
	}
	return clientCh, serverCh
}

func TestDatagramModeAcceptsGaps(t *testing.T) {
	clientCh, serverCh := newTestDatagramChannel(t)

	// Encrypt seq 0, 1, 2, 3, 4, 5 on the sender side.
	cts := make([][]byte, 6)
	for i := range cts {
		cts[i] = clientCh.Encrypt([]byte("msg"))
	}

	// Receiver gets seq 0, 1, and 5 — skipping 2, 3, 4.
	for _, idx := range []int{0, 1, 5} {
		if _, err := serverCh.Decrypt(cts[idx]); err != nil {
			t.Fatalf("seq %d: unexpected error: %v", idx, err)
		}
	}
}

func TestDatagramModeRejectsReplay(t *testing.T) {
	clientCh, serverCh := newTestDatagramChannel(t)

	ct := clientCh.Encrypt([]byte("first"))
	if _, err := serverCh.Decrypt(ct); err != nil {
		t.Fatal(err)
	}
	// Replay the same packet.
	if _, err := serverCh.Decrypt(ct); err == nil {
		t.Fatal("expected replay to be rejected in datagram mode")
	}
}

func TestDatagramModeRejectsOldSeq(t *testing.T) {
	clientCh, serverCh := newTestDatagramChannel(t)

	// Encrypt 6 packets (seq 0..5).
	cts := make([][]byte, 6)
	for i := range cts {
		cts[i] = clientCh.Encrypt([]byte("msg"))
	}

	// Receive seq=5 first (jump forward).
	if _, err := serverCh.Decrypt(cts[5]); err != nil {
		t.Fatal(err)
	}
	// Now seq=3 is in the past — must be rejected.
	if _, err := serverCh.Decrypt(cts[3]); err == nil {
		t.Fatal("expected old sequence number to be rejected in datagram mode")
	}
}

func TestStrictModeRejectsGaps(t *testing.T) {
	clientCh, serverCh := newTestChannel(t)

	// Encrypt two packets.
	ct0 := clientCh.Encrypt([]byte("first"))
	ct1 := clientCh.Encrypt([]byte("second"))
	_ = ct0

	// Receiver tries to decrypt seq=1 without first receiving seq=0.
	if _, err := serverCh.Decrypt(ct1); err == nil {
		t.Fatal("strict mode should reject a gap (missing seq 0)")
	}
}

func TestNewDatagramChannel(t *testing.T) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		t.Fatal(err)
	}
	sendKey, err := hkdfDerive(key, []byte("s2c"))
	if err != nil {
		t.Fatal(err)
	}
	recvKey, err := hkdfDerive(key, []byte("c2s"))
	if err != nil {
		t.Fatal(err)
	}

	ch, err := NewDatagramChannel(sendKey, recvKey)
	if err != nil {
		t.Fatal(err)
	}
	if ch.mode != ModeDatagrams {
		t.Fatal("NewDatagramChannel should create a channel in ModeDatagrams")
	}
}

func TestDatagramModeSimulatedPacketLoss(t *testing.T) {
	clientCh, serverCh := newTestDatagramChannel(t)

	const total = 1000
	lossRates := []float64{0.05, 0.10, 0.20, 0.50}

	for _, lossRate := range lossRates {
		t.Run(fmt.Sprintf("loss_%.0f%%", lossRate*100), func(t *testing.T) {
			// Reset channels for each sub-test.
			clientCh, serverCh = newTestDatagramChannel(t)

			// Encrypt all messages.
			ciphertexts := make([][]byte, total)
			for i := range total {
				ciphertexts[i] = clientCh.Encrypt([]byte(fmt.Sprintf("msg-%d", i)))
			}

			// Simulate loss: drop packets deterministically based on index.
			// Use a simple hash to distribute drops evenly.
			delivered := 0
			dropped := 0
			var lastDelivered int
			for i, ct := range ciphertexts {
				// Deterministic "random" drop based on index.
				drop := float64((i*7+13)%100) / 100.0 < lossRate
				if drop {
					dropped++
					continue
				}

				plaintext, err := serverCh.Decrypt(ct)
				if err != nil {
					t.Fatalf("seq %d: decrypt failed after %d delivered, %d dropped: %v",
						i, delivered, dropped, err)
				}

				expected := fmt.Sprintf("msg-%d", i)
				if string(plaintext) != expected {
					t.Fatalf("seq %d: got %q, want %q", i, plaintext, expected)
				}
				delivered++
				lastDelivered = i
			}

			t.Logf("%.0f%% loss: %d/%d delivered, %d dropped, last seq=%d",
				lossRate*100, delivered, total, dropped, lastDelivered)

			if delivered == 0 {
				t.Fatal("no messages delivered")
			}
			if delivered+dropped != total {
				t.Fatalf("delivered(%d) + dropped(%d) != total(%d)", delivered, dropped, total)
			}
		})
	}
}

func TestDatagramModeStress(t *testing.T) {
	clientCh, serverCh := newTestDatagramChannel(t)

	// Rapid-fire 10,000 messages with random gaps.
	const total = 10000
	ciphertexts := make([][]byte, total)
	for i := range total {
		payload := make([]byte, 64)
		rand.Read(payload)
		ciphertexts[i] = clientCh.Encrypt(payload)
	}

	// Deliver only even-numbered packets (50% loss, every other packet).
	delivered := 0
	for i := 0; i < total; i += 2 {
		if _, err := serverCh.Decrypt(ciphertexts[i]); err != nil {
			t.Fatalf("seq %d: %v (after %d delivered)", i, err, delivered)
		}
		delivered++
	}

	t.Logf("stress: %d/%d delivered (50%% systematic loss)", delivered, total)
	if delivered != total/2 {
		t.Fatalf("expected %d delivered, got %d", total/2, delivered)
	}
}

func TestConfirmationCodeCrossplatformVector(t *testing.T) {
	// Fixed X25519 public keys (any 32-byte value is a valid X25519 point).
	keyA := bytes.Repeat([]byte{0x01}, 32)
	keyB := bytes.Repeat([]byte{0x02}, 32)

	pubA, err := ecdh.X25519().NewPublicKey(keyA)
	if err != nil {
		t.Fatal(err)
	}
	pubB, err := ecdh.X25519().NewPublicKey(keyB)
	if err != nil {
		t.Fatal(err)
	}

	code, err := DeriveConfirmationCode(pubA, pubB)
	if err != nil {
		t.Fatal(err)
	}

	// This expected value must match the TypeScript test in web/src/crypto.test.ts.
	const expected = "629624"
	if code != expected {
		t.Fatalf("got %q, want %q", code, expected)
	}

	// Verify order-independence with the same expected value.
	codeBA, err := DeriveConfirmationCode(pubB, pubA)
	if err != nil {
		t.Fatal(err)
	}
	if codeBA != expected {
		t.Fatalf("reversed order: got %q, want %q", codeBA, expected)
	}
}

func TestDeriveConfirmationCode(t *testing.T) {
	client, err := GenerateKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	server, err := GenerateKeyPair()
	if err != nil {
		t.Fatal(err)
	}

	// Both sides compute the same code.
	code1, err := DeriveConfirmationCode(server.Public, client.Public)
	if err != nil {
		t.Fatal(err)
	}
	code2, err := DeriveConfirmationCode(client.Public, server.Public)
	if err != nil {
		t.Fatal(err)
	}
	if code1 != code2 {
		t.Fatalf("codes should be order-independent: %s vs %s", code1, code2)
	}
	if len(code1) != 6 {
		t.Fatalf("expected 6-digit code, got %q", code1)
	}

	// MitM: adversary substitutes its own key — codes diverge.
	adversary, err := GenerateKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	codeServer, err := DeriveConfirmationCode(server.Public, adversary.Public)
	if err != nil {
		t.Fatal(err)
	}
	codeClient, err := DeriveConfirmationCode(adversary.Public, client.Public)
	if err != nil {
		t.Fatal(err)
	}
	if codeServer == codeClient {
		t.Fatal("MitM codes should differ (collision — extremely unlikely)")
	}
	if codeServer == code1 {
		t.Fatal("MitM server code should differ from honest code")
	}
	if codeClient == code1 {
		t.Fatal("MitM client code should differ from honest code")
	}
}

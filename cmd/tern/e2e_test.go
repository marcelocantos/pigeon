// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"crypto/ecdh"
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"

	"github.com/marcelocantos/tern"
	"github.com/marcelocantos/tern/crypto"
)

// TestE2EPairingAndEncryptedRelay_WT exercises the full stack via WebTransport.
func TestE2EPairingAndEncryptedRelay_WT(t *testing.T) {
	url, tlsConfig := startTestRelay(t, "")
	runE2EPairingTest(t, url, tern.WithTLS(tlsConfig), tern.WithWebTransport())
}

// TestE2EPairingAndEncryptedRelay_QUIC exercises the full stack via raw QUIC.
func TestE2EPairingAndEncryptedRelay_QUIC(t *testing.T) {
	tr := startTestRelayWithQUIC(t, "")
	runE2EPairingTest(t, tr.wtURL, tern.WithTLS(tr.tlsConfig), tern.WithQUICPort(tr.quicPort))
}

// TestE2ECrossProtocol exercises QUIC backend + WebTransport client.
func TestE2ECrossProtocol(t *testing.T) {
	tr := startTestRelayWithQUIC(t, "")
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Backend registers via raw QUIC.
	backend, err := tern.Register(ctx, tr.wtURL,
		tern.WithTLS(tr.tlsConfig),
		tern.WithQUICPort(tr.quicPort),
	)
	if err != nil {
		t.Fatalf("backend register: %v", err)
	}
	defer backend.CloseNow()

	// Client connects via WebTransport.
	client, err := tern.Connect(ctx, tr.wtURL, backend.InstanceID(),
		tern.WithTLS(tr.tlsConfig),
		tern.WithWebTransport(),
	)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	defer client.CloseNow()

	// Bidirectional message exchange.
	if err := client.Send(ctx, []byte("cross-protocol hello")); err != nil {
		t.Fatalf("client send: %v", err)
	}
	data, err := backend.Recv(ctx)
	if err != nil {
		t.Fatalf("backend recv: %v", err)
	}
	if string(data) != "cross-protocol hello" {
		t.Fatalf("got %q, want %q", data, "cross-protocol hello")
	}

	if err := backend.Send(ctx, []byte("cross-protocol reply")); err != nil {
		t.Fatalf("backend send: %v", err)
	}
	data, err = client.Recv(ctx)
	if err != nil {
		t.Fatalf("client recv: %v", err)
	}
	if string(data) != "cross-protocol reply" {
		t.Fatalf("got %q, want %q", data, "cross-protocol reply")
	}

	t.Log("Cross-protocol (QUIC backend + WT client): PASS")
}

// runE2EPairingTest exercises:
//  1. Backend registers, gets instance ID
//  2. Client connects via instance ID
//  3. Pairing ceremony: ECDH key exchange through relay, confirmation code
//  4. Encrypted channel established
//  5. Encrypted messages flow bidirectionally
//  6. Relay sees only ciphertext
func runE2EPairingTest(t *testing.T, relayURL string, opts ...tern.Option) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Backend registers via relay package.
	backend, err := tern.Register(ctx, relayURL, opts...)
	if err != nil {
		t.Fatalf("backend register: %v", err)
	}
	defer backend.CloseNow()
	t.Logf("Backend registered as %s", backend.InstanceID())

	// Client connects via relay package.
	client, err := tern.Connect(ctx, relayURL, backend.InstanceID(), opts...)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	defer client.CloseNow()

	// --- Pairing ceremony ---

	// Both sides generate ECDH key pairs.
	backendKP, err := crypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("backend keygen: %v", err)
	}
	clientKP, err := crypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("client keygen: %v", err)
	}

	// Client sends pair_hello with its public key (through relay).
	pairHello, _ := json.Marshal(map[string]string{
		"type":   "pair_hello",
		"pubkey": base64.StdEncoding.EncodeToString(clientKP.Public.Bytes()),
	})
	if err := client.Send(ctx, pairHello); err != nil {
		t.Fatalf("client write pair_hello: %v", err)
	}

	// Backend receives pair_hello through relay.
	helloData, err := backend.Recv(ctx)
	if err != nil {
		t.Fatalf("backend read pair_hello: %v", err)
	}

	var hello struct {
		Type   string `json:"type"`
		Pubkey string `json:"pubkey"`
	}
	if err := json.Unmarshal(helloData, &hello); err != nil {
		t.Fatalf("parse pair_hello: %v", err)
	}
	if hello.Type != "pair_hello" {
		t.Fatalf("expected pair_hello, got %s", hello.Type)
	}

	// Backend extracts client public key.
	clientPubBytes, _ := base64.StdEncoding.DecodeString(hello.Pubkey)
	clientPub, err := ecdh.X25519().NewPublicKey(clientPubBytes)
	if err != nil {
		t.Fatalf("parse client pubkey: %v", err)
	}

	// Backend sends pair_hello_ack with its public key.
	ack, _ := json.Marshal(map[string]string{
		"type":   "pair_hello_ack",
		"pubkey": base64.StdEncoding.EncodeToString(backendKP.Public.Bytes()),
	})
	if err := backend.Send(ctx, ack); err != nil {
		t.Fatalf("backend write pair_hello_ack: %v", err)
	}

	// Client receives pair_hello_ack.
	ackData, err := client.Recv(ctx)
	if err != nil {
		t.Fatalf("client read pair_hello_ack: %v", err)
	}

	var ackMsg struct {
		Type   string `json:"type"`
		Pubkey string `json:"pubkey"`
	}
	if err := json.Unmarshal(ackData, &ackMsg); err != nil {
		t.Fatalf("parse pair_hello_ack: %v", err)
	}

	backendPubBytes, _ := base64.StdEncoding.DecodeString(ackMsg.Pubkey)
	backendPub, err := ecdh.X25519().NewPublicKey(backendPubBytes)
	if err != nil {
		t.Fatalf("parse backend pubkey: %v", err)
	}

	// Both sides derive confirmation codes (should match — no MitM).
	backendCode, err := crypto.DeriveConfirmationCode(backendKP.Public, clientPub)
	if err != nil {
		t.Fatalf("backend derive code: %v", err)
	}
	clientCode, err := crypto.DeriveConfirmationCode(backendPub, clientKP.Public)
	if err != nil {
		t.Fatalf("client derive code: %v", err)
	}

	if backendCode != clientCode {
		t.Fatalf("confirmation codes don't match: backend=%s client=%s", backendCode, clientCode)
	}
	t.Logf("Confirmation codes match: %s", backendCode)

	// --- Derive session keys and create encrypted channels ---

	backendSendKey, err := crypto.DeriveSessionKey(backendKP.Private, clientPub, []byte("server-to-client"))
	if err != nil {
		t.Fatalf("backend derive send key: %v", err)
	}
	backendRecvKey, err := crypto.DeriveSessionKey(backendKP.Private, clientPub, []byte("client-to-server"))
	if err != nil {
		t.Fatalf("backend derive recv key: %v", err)
	}

	clientSendKey, err := crypto.DeriveSessionKey(clientKP.Private, backendPub, []byte("client-to-server"))
	if err != nil {
		t.Fatalf("client derive send key: %v", err)
	}
	clientRecvKey, err := crypto.DeriveSessionKey(clientKP.Private, backendPub, []byte("server-to-client"))
	if err != nil {
		t.Fatalf("client derive recv key: %v", err)
	}

	backendCh, err := crypto.NewChannel(backendSendKey, backendRecvKey)
	if err != nil {
		t.Fatalf("backend channel: %v", err)
	}
	clientCh, err := crypto.NewChannel(clientSendKey, clientRecvKey)
	if err != nil {
		t.Fatalf("client channel: %v", err)
	}

	// --- Encrypted message exchange through relay ---

	// Client sends encrypted message.
	plaintext := []byte("secret message from client")
	ciphertext := clientCh.Encrypt(plaintext)

	if err := client.Send(ctx, ciphertext); err != nil {
		t.Fatalf("client write encrypted: %v", err)
	}

	// Backend receives and decrypts.
	relayedData, err := backend.Recv(ctx)
	if err != nil {
		t.Fatalf("backend read encrypted: %v", err)
	}

	// Verify relay passed through ciphertext (not plaintext).
	if string(relayedData) == string(plaintext) {
		t.Fatal("relay leaked plaintext — encryption not working")
	}

	decrypted, err := backendCh.Decrypt(relayedData)
	if err != nil {
		t.Fatalf("backend decrypt: %v", err)
	}
	if string(decrypted) != string(plaintext) {
		t.Fatalf("decrypted %q, want %q", decrypted, plaintext)
	}

	// Backend sends encrypted reply.
	reply := []byte("secret reply from backend")
	replyCiphertext := backendCh.Encrypt(reply)

	if err := backend.Send(ctx, replyCiphertext); err != nil {
		t.Fatalf("backend write encrypted: %v", err)
	}

	// Client receives and decrypts.
	relayedReply, err := client.Recv(ctx)
	if err != nil {
		t.Fatalf("client read encrypted: %v", err)
	}

	if string(relayedReply) == string(reply) {
		t.Fatal("relay leaked plaintext on reply")
	}

	decryptedReply, err := clientCh.Decrypt(relayedReply)
	if err != nil {
		t.Fatalf("client decrypt: %v", err)
	}
	if string(decryptedReply) != string(reply) {
		t.Fatalf("decrypted %q, want %q", decryptedReply, reply)
	}

	t.Log("E2E pairing + encrypted relay: PASS")
}

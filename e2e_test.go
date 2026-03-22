// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"crypto/ecdh"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"

	"github.com/marcelocantos/tern/crypto"
)

// TestE2EPairingAndEncryptedRelay exercises the full stack:
//
//  1. Start relay
//  2. Backend registers, gets instance ID
//  3. Client connects via /ws/{id}
//  4. Pairing ceremony: ECDH key exchange through relay, confirmation code
//  5. Encrypted channel established
//  6. Encrypted messages flow bidirectionally
//  7. Relay sees only ciphertext
func TestE2EPairingAndEncryptedRelay(t *testing.T) {
	// Start relay.
	r := newRelay()
	mux := http.NewServeMux()
	registerRoutes(mux, r)
	ts := httptest.NewServer(mux)
	defer ts.Close()

	wsBase := "ws" + strings.TrimPrefix(ts.URL, "http")
	ctx := context.Background()

	// Backend registers.
	backendConn, _, err := websocket.Dial(ctx, wsBase+"/register", nil)
	if err != nil {
		t.Fatalf("backend register: %v", err)
	}
	defer backendConn.CloseNow()

	// Read instance ID.
	_, idBytes, err := backendConn.Read(ctx)
	if err != nil {
		t.Fatalf("read instance ID: %v", err)
	}
	instanceID := string(idBytes)
	t.Logf("Backend registered as %s", instanceID)

	// Client connects.
	clientConn, _, err := websocket.Dial(ctx, wsBase+"/ws/"+instanceID, nil)
	if err != nil {
		t.Fatalf("client connect: %v", err)
	}
	defer clientConn.CloseNow()

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
	if err := clientConn.Write(ctx, websocket.MessageText, pairHello); err != nil {
		t.Fatalf("client write pair_hello: %v", err)
	}

	// Backend receives pair_hello through relay.
	rctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	_, helloData, err := backendConn.Read(rctx)
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
	if err := backendConn.Write(ctx, websocket.MessageText, ack); err != nil {
		t.Fatalf("backend write pair_hello_ack: %v", err)
	}

	// Client receives pair_hello_ack.
	rctx2, cancel2 := context.WithTimeout(ctx, 2*time.Second)
	defer cancel2()
	_, ackData, err := clientConn.Read(rctx2)
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

	if err := clientConn.Write(ctx, websocket.MessageBinary, ciphertext); err != nil {
		t.Fatalf("client write encrypted: %v", err)
	}

	// Backend receives and decrypts.
	rctx3, cancel3 := context.WithTimeout(ctx, 2*time.Second)
	defer cancel3()
	mt, relayedData, err := backendConn.Read(rctx3)
	if err != nil {
		t.Fatalf("backend read encrypted: %v", err)
	}
	if mt != websocket.MessageBinary {
		t.Fatalf("expected binary, got %v", mt)
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

	if err := backendConn.Write(ctx, websocket.MessageBinary, replyCiphertext); err != nil {
		t.Fatalf("backend write encrypted: %v", err)
	}

	// Client receives and decrypts.
	rctx4, cancel4 := context.WithTimeout(ctx, 2*time.Second)
	defer cancel4()
	mt, relayedReply, err := clientConn.Read(rctx4)
	if err != nil {
		t.Fatalf("client read encrypted: %v", err)
	}
	if mt != websocket.MessageBinary {
		t.Fatalf("expected binary, got %v", mt)
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

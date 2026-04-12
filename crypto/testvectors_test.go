// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package crypto

import (
	"crypto/ecdh"
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"
)

// TestGenerateVectors writes deterministic test vectors to c/test/vectors.json
// for cross-language validation by the C test suite.
func TestGenerateVectors(t *testing.T) {
	// Fixed private key seeds — deterministic across runs.
	aliceSeed := make([]byte, 32)
	for i := range aliceSeed {
		aliceSeed[i] = 0x01
	}
	bobSeed := make([]byte, 32)
	for i := range bobSeed {
		bobSeed[i] = 0x02
	}

	// Construct keypairs from fixed seeds.
	alicePriv, err := ecdh.X25519().NewPrivateKey(aliceSeed)
	if err != nil {
		t.Fatalf("alice NewPrivateKey: %v", err)
	}
	bobPriv, err := ecdh.X25519().NewPrivateKey(bobSeed)
	if err != nil {
		t.Fatalf("bob NewPrivateKey: %v", err)
	}
	alicePub := alicePriv.PublicKey()
	bobPub := bobPriv.PublicKey()

	// Derive session key (alice→bob direction, info="test-session").
	info := []byte("test-session")
	sessionKey, err := DeriveSessionKey(alicePriv, bobPub, info)
	if err != nil {
		t.Fatalf("DeriveSessionKey: %v", err)
	}
	// Verify both sides derive the same key.
	sessionKeyCheck, err := DeriveSessionKey(bobPriv, alicePub, info)
	if err != nil {
		t.Fatalf("DeriveSessionKey (bob side): %v", err)
	}
	if hex.EncodeToString(sessionKey) != hex.EncodeToString(sessionKeyCheck) {
		t.Fatal("session keys don't match between alice and bob")
	}

	// Derive confirmation code.
	confirmCode, err := DeriveConfirmationCode(alicePub, bobPub)
	if err != nil {
		t.Fatalf("DeriveConfirmationCode: %v", err)
	}

	// Derive send/recv keys for alice channel (alice sends, bob receives).
	sendKey, err := DeriveSessionKey(alicePriv, bobPub, []byte("alice-to-bob"))
	if err != nil {
		t.Fatalf("DeriveSessionKey (send): %v", err)
	}
	recvKey, err := DeriveSessionKey(alicePriv, bobPub, []byte("bob-to-alice"))
	if err != nil {
		t.Fatalf("DeriveSessionKey (recv): %v", err)
	}

	// Create alice's channel and encrypt.
	aliceCh, err := NewChannel(sendKey, recvKey)
	if err != nil {
		t.Fatalf("NewChannel: %v", err)
	}
	plaintext := []byte("hello from pigeon")
	ciphertext := aliceCh.Encrypt(plaintext)

	// Verify: bob can decrypt.
	bobSendKey, err := DeriveSessionKey(bobPriv, alicePub, []byte("bob-to-alice"))
	if err != nil {
		t.Fatalf("DeriveSessionKey (bob send): %v", err)
	}
	bobRecvKey, err := DeriveSessionKey(bobPriv, alicePub, []byte("alice-to-bob"))
	if err != nil {
		t.Fatalf("DeriveSessionKey (bob recv): %v", err)
	}
	bobCh, err := NewChannel(bobSendKey, bobRecvKey)
	if err != nil {
		t.Fatalf("NewChannel (bob): %v", err)
	}
	decrypted, err := bobCh.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Decrypt verification: %v", err)
	}
	if string(decrypted) != string(plaintext) {
		t.Fatalf("decrypt mismatch: got %q, want %q", decrypted, plaintext)
	}

	// Symmetric channel: known master key.
	masterKey := make([]byte, 32)
	for i := range masterKey {
		masterKey[i] = 0x42
	}
	clientSymCh, err := NewSymmetricChannel(masterKey, false)
	if err != nil {
		t.Fatalf("NewSymmetricChannel (client): %v", err)
	}
	serverSymCh, err := NewSymmetricChannel(masterKey, true)
	if err != nil {
		t.Fatalf("NewSymmetricChannel (server): %v", err)
	}

	symPlaintext := []byte("symmetric test")

	// Client → server (c2s).
	c2sCiphertext := clientSymCh.Encrypt(symPlaintext)
	// Verify server can decrypt.
	c2sDecrypted, err := serverSymCh.Decrypt(c2sCiphertext)
	if err != nil {
		t.Fatalf("symmetric c2s decrypt verification: %v", err)
	}
	if string(c2sDecrypted) != string(symPlaintext) {
		t.Fatalf("symmetric c2s mismatch: got %q, want %q", c2sDecrypted, symPlaintext)
	}

	// Server → client (s2c).
	s2cCiphertext := serverSymCh.Encrypt(symPlaintext)
	// Verify client can decrypt.
	s2cDecrypted, err := clientSymCh.Decrypt(s2cCiphertext)
	if err != nil {
		t.Fatalf("symmetric s2c decrypt verification: %v", err)
	}
	if string(s2cDecrypted) != string(symPlaintext) {
		t.Fatalf("symmetric s2c mismatch: got %q, want %q", s2cDecrypted, symPlaintext)
	}

	// Assemble vectors.
	vectors := map[string]string{
		"alice_private":            hex.EncodeToString(alicePriv.Bytes()),
		"alice_public":             hex.EncodeToString(alicePub.Bytes()),
		"bob_private":              hex.EncodeToString(bobPriv.Bytes()),
		"bob_public":               hex.EncodeToString(bobPub.Bytes()),
		"session_key":              hex.EncodeToString(sessionKey),
		"confirmation_code":        confirmCode,
		"plaintext":                string(plaintext),
		"ciphertext":               hex.EncodeToString(ciphertext),
		"symmetric_master":         hex.EncodeToString(masterKey),
		"symmetric_plaintext":      string(symPlaintext),
		"symmetric_ciphertext_c2s": hex.EncodeToString(c2sCiphertext),
		"symmetric_ciphertext_s2c": hex.EncodeToString(s2cCiphertext),
	}

	data, err := json.MarshalIndent(vectors, "", "  ")
	if err != nil {
		t.Fatalf("json.MarshalIndent: %v", err)
	}

	outPath := "../c/test/vectors.json"
	if err := os.WriteFile(outPath, data, 0o644); err != nil {
		t.Fatalf("WriteFile %s: %v", outPath, err)
	}
	t.Logf("Wrote test vectors to %s", outPath)
}

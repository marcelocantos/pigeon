// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

//go:build darwin && cgo

package main

import (
	"bytes"
	"crypto/ecdh"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"os"
	"testing"

	"github.com/keybase/go-keychain"
)

// randomService returns a fresh keychain service name per test so parallel
// or stale runs don't collide on the login keychain.
func randomService(t *testing.T) string {
	t.Helper()
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		t.Fatalf("randomService: %v", err)
	}
	return "com.marcelocantos.pigeon.test." + hex.EncodeToString(b[:])
}

// deleteItem removes the Keychain item written by newKeychainIdentity so
// the test doesn't leak entries into the user's login keychain.
func deleteItem(service, account string) error {
	q := keychain.NewItem()
	q.SetSecClass(keychain.SecClassGenericPassword)
	q.SetService(service)
	q.SetAccount(account)
	return keychain.DeleteItem(q)
}

// skipIfKeychainUnavailable bails out when the test host can't reach the
// login keychain (headless CI, locked session). Anything that isn't a
// "no such item" or "user not interactive"-style error means the test
// actually exercised the Keychain and any failure is a real failure.
func skipIfKeychainUnavailable(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		return
	}
	var kErr keychain.Error
	if errors.As(err, &kErr) {
		switch kErr {
		case keychain.ErrorInteractionNotAllowed,
			keychain.ErrorNotAvailable,
			keychain.ErrorNoSuchKeychain:
			t.Skipf("keychain unavailable in this environment: %v", err)
		}
	}
}

// TestKeychainIdentity_RoundTrip checks the full Tier 1 acceptance:
// first call generates and stores; second call loads the same material;
// PublicKey/InstanceID are stable; DeriveSharedSecret produces the same
// 32-byte HKDF output across both Identity instances.
func TestKeychainIdentity_RoundTrip(t *testing.T) {
	service := randomService(t)
	const account = "default"

	first, err := newKeychainIdentity(service, account)
	skipIfKeychainUnavailable(t, err)
	if err != nil {
		t.Fatalf("first load: %v", err)
	}
	t.Cleanup(func() {
		if err := deleteItem(service, account); err != nil {
			// best-effort: a stale item only matters to the user's keychain,
			// not to subsequent runs (each picks a fresh randomService).
			t.Logf("cleanup deleteItem: %v", err)
		}
	})

	if got := len(first.PublicKey()); got != 32 {
		t.Fatalf("PublicKey length: want 32, got %d", got)
	}
	if first.InstanceID() == "" {
		t.Fatal("InstanceID empty")
	}

	// Second load must hit the existing Keychain item, not generate a new one.
	second, err := newKeychainIdentity(service, account)
	if err != nil {
		t.Fatalf("second load: %v", err)
	}

	if !bytes.Equal(first.PublicKey(), second.PublicKey()) {
		t.Fatal("PublicKey changed across loads; Keychain item not reused")
	}
	if first.InstanceID() != second.InstanceID() {
		t.Fatalf("InstanceID changed across loads: %q vs %q", first.InstanceID(), second.InstanceID())
	}

	// DeriveSharedSecret: a fresh peer keypair against both instances should
	// yield the same 32-byte HKDF output. This confirms the private X25519
	// key was decoded identically on both reads.
	peer, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("peer keygen: %v", err)
	}
	info := []byte("pigeon/test/keychain-roundtrip")

	s1, err := first.DeriveSharedSecret(peer.PublicKey().Bytes(), info)
	if err != nil {
		t.Fatalf("first DeriveSharedSecret: %v", err)
	}
	s2, err := second.DeriveSharedSecret(peer.PublicKey().Bytes(), info)
	if err != nil {
		t.Fatalf("second DeriveSharedSecret: %v", err)
	}
	if !bytes.Equal(s1, s2) {
		t.Fatal("DeriveSharedSecret mismatch across reloaded identity")
	}
	if len(s1) != 32 {
		t.Fatalf("derived secret length: want 32, got %d", len(s1))
	}

	// Sign must also work post-reload.
	sig, err := second.Sign([]byte("hello"))
	if err != nil {
		t.Fatalf("Sign: %v", err)
	}
	if len(sig) == 0 {
		t.Fatal("Sign returned empty signature")
	}
}

// TestKeychainIdentity_GeneratesOnFirstUse verifies that newKeychainIdentity
// creates a fresh item when none exists — and that calling it on an empty
// service+account does not return the "load" error path.
func TestKeychainIdentity_GeneratesOnFirstUse(t *testing.T) {
	service := randomService(t)
	const account = "default"

	// Sanity: nothing there yet.
	q := keychain.NewItem()
	q.SetSecClass(keychain.SecClassGenericPassword)
	q.SetService(service)
	q.SetAccount(account)
	q.SetMatchLimit(keychain.MatchLimitOne)
	q.SetReturnData(true)
	results, err := keychain.QueryItem(q)
	skipIfKeychainUnavailable(t, err)
	if err != nil {
		t.Fatalf("pre-check query: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("pre-check: service %q already has %d items", service, len(results))
	}

	id, err := newKeychainIdentity(service, account)
	if err != nil {
		t.Fatalf("newKeychainIdentity: %v", err)
	}
	t.Cleanup(func() { _ = deleteItem(service, account) })

	_ = id
	// Confirm the item is now present.
	results, err = keychain.QueryItem(q)
	if err != nil {
		t.Fatalf("post-check query: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("post-check: want 1 item under %q, got %d", service, len(results))
	}
}

// TestMain ensures TMPDIR is set so any incidental os.TempDir() calls in
// the keychain package don't interact with $HOME-less CI sandboxes.
func TestMain(m *testing.M) {
	if os.Getenv("TMPDIR") == "" {
		_ = os.Setenv("TMPDIR", "/tmp")
	}
	os.Exit(m.Run())
}

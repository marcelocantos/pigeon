// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

//go:build darwin && cgo

// macOS Keychain-backed crypto.Identity for echo-backend.
//
// First run generates an X25519/Ed25519 keypair and stores both private
// halves as a single Keychain generic-password item
// (kSecClassGenericPassword) with kSecAttrAccessible=WhenUnlocked and
// kSecAttrSynchronizable=false. Subsequent runs read the item and reuse
// the keypair, so the derived InstanceID — and any pairings already
// recorded against it — survive daemon restarts and re-launches.
//
// Re-signing the binary (e.g. dev build → notarised distribution build)
// changes the Keychain partition associated with the item: macOS will
// refuse to return the data to the new binary, the daemon will treat
// this as "no item yet", a fresh keypair is generated, and previously
// paired clients have to re-pair once. This is a property of the OS,
// not pigeon.

package main

import (
	"crypto/ecdh"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/keybase/go-keychain"
	"github.com/marcelocantos/pigeon/crypto"
	"golang.org/x/crypto/hkdf"
)

// keychainBlob is the on-Keychain representation: both private halves
// in one generic-password item. JSON over the wire is overkill for two
// fixed-length byte arrays but keeps debugging trivial and matches the
// file-backed identity's storage shape.
type keychainBlob struct {
	X25519Private  []byte `json:"x25519_private"`  // 32 bytes
	Ed25519Private []byte `json:"ed25519_private"` // 64 bytes (seed + public)
}

type keychainIdentity struct {
	xPriv  *ecdh.PrivateKey
	edPriv ed25519.PrivateKey
	xPub   []byte
}

// newKeychainIdentity loads the long-term Identity from the macOS
// Keychain under (service, account), generating and storing a new
// keypair if no item exists yet. Decoded keys are cached in process
// memory; Sign and DeriveSharedSecret do not re-hit the Keychain.
//
// service is the user-visible "Where" column in Keychain Access (use
// reverse-DNS, e.g. "com.example.echo-backend"). account distinguishes
// multiple identities under the same service (e.g. "default", "staging").
func newKeychainIdentity(service, account string) (crypto.Identity, error) {
	blob, err := keychainLoad(service, account)
	if errors.Is(err, errKeychainNotFound) {
		blob, err = keychainGenerateAndStore(service, account)
	}
	if err != nil {
		return nil, err
	}

	if len(blob.X25519Private) != 32 {
		return nil, fmt.Errorf("keychain identity x25519 private: want 32 bytes, got %d", len(blob.X25519Private))
	}
	if len(blob.Ed25519Private) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("keychain identity ed25519 private: want %d bytes, got %d", ed25519.PrivateKeySize, len(blob.Ed25519Private))
	}
	xPriv, err := ecdh.X25519().NewPrivateKey(blob.X25519Private)
	if err != nil {
		return nil, fmt.Errorf("parse x25519 from keychain: %w", err)
	}
	return &keychainIdentity{
		xPriv:  xPriv,
		edPriv: ed25519.PrivateKey(blob.Ed25519Private),
		xPub:   xPriv.PublicKey().Bytes(),
	}, nil
}

var errKeychainNotFound = errors.New("keychain item not found")

func keychainLoad(service, account string) (*keychainBlob, error) {
	q := keychain.NewItem()
	q.SetSecClass(keychain.SecClassGenericPassword)
	q.SetService(service)
	q.SetAccount(account)
	q.SetMatchLimit(keychain.MatchLimitOne)
	q.SetReturnData(true)
	results, err := keychain.QueryItem(q)
	if err != nil {
		return nil, fmt.Errorf("keychain query: %w", err)
	}
	if len(results) == 0 {
		return nil, errKeychainNotFound
	}
	var blob keychainBlob
	if err := json.Unmarshal(results[0].Data, &blob); err != nil {
		return nil, fmt.Errorf("parse keychain identity: %w", err)
	}
	return &blob, nil
}

func keychainGenerateAndStore(service, account string) (*keychainBlob, error) {
	xPriv, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generate x25519: %w", err)
	}
	_, edPriv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generate ed25519: %w", err)
	}
	blob := &keychainBlob{
		X25519Private:  xPriv.Bytes(),
		Ed25519Private: []byte(edPriv),
	}
	data, err := json.Marshal(blob)
	if err != nil {
		return nil, fmt.Errorf("marshal keychain identity: %w", err)
	}

	item := keychain.NewItem()
	item.SetSecClass(keychain.SecClassGenericPassword)
	item.SetService(service)
	item.SetAccount(account)
	item.SetLabel("pigeon backend identity")
	item.SetData(data)
	item.SetAccessible(keychain.AccessibleWhenUnlocked)
	item.SetSynchronizable(keychain.SynchronizableNo)

	if err := keychain.AddItem(item); err != nil {
		return nil, fmt.Errorf("keychain add: %w", err)
	}
	return blob, nil
}

func (k *keychainIdentity) PublicKey() []byte { return k.xPub }

func (k *keychainIdentity) InstanceID() string {
	sum := sha256.Sum256(k.xPub)
	return base64.RawURLEncoding.EncodeToString(sum[:16])
}

func (k *keychainIdentity) DeriveSharedSecret(peerPub, info []byte) ([]byte, error) {
	pub, err := ecdh.X25519().NewPublicKey(peerPub)
	if err != nil {
		return nil, fmt.Errorf("parse peer public key: %w", err)
	}
	shared, err := k.xPriv.ECDH(pub)
	if err != nil {
		return nil, err
	}
	r := hkdf.New(sha256.New, shared, nil, info)
	key := make([]byte, 32)
	if _, err := r.Read(key); err != nil {
		return nil, err
	}
	return key, nil
}

func (k *keychainIdentity) Sign(msg []byte) ([]byte, error) {
	return ed25519.Sign(k.edPriv, msg), nil
}

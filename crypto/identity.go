// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package crypto

import (
	"crypto/ecdh"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
)

// Identity is a long-term peer identity. Implementations may hold private
// material in process memory, on disk, in the OS keychain, in a hardware
// enclave, or anywhere else; pigeon never sees the private bytes directly.
type Identity interface {
	// PublicKey returns the X25519 public key (32 bytes). Always extractable.
	PublicKey() []byte

	// InstanceID returns the stable identifier derived from PublicKey.
	InstanceID() string

	// DeriveSharedSecret performs X25519 ECDH against peerPub, then HKDF
	// with the given info label. Used during pairing and on every connect.
	DeriveSharedSecret(peerPub, info []byte) ([]byte, error)

	// Sign produces an Ed25519 signature over msg using the identity's
	// signing key. Used to authenticate during pairing and connect handshakes.
	Sign(msg []byte) ([]byte, error)
}

// fileIdentity is the file-backed Identity implementation returned by
// NewFileIdentity. Holds an X25519 ECDH keypair (for DeriveSharedSecret)
// and an Ed25519 signing keypair (for Sign), both decoded once at load
// time and held in process memory thereafter.
type fileIdentity struct {
	xPriv  *ecdh.PrivateKey
	edPriv ed25519.PrivateKey
	xPub   []byte
}

// fileIdentityBlob is the on-disk representation written by NewFileIdentity.
type fileIdentityBlob struct {
	X25519Private  []byte `json:"x25519_private"`  // 32 bytes
	Ed25519Private []byte `json:"ed25519_private"` // 64 bytes (seed + public)
}

// NewFileIdentity loads an Identity from path, or generates one and writes
// it there if the file does not exist. Suitable for development and CLI
// demos; production deployments supply their own Identity implementation
// backed by a keychain, enclave, or HSM.
func NewFileIdentity(path string) (Identity, error) {
	b, err := os.ReadFile(path)
	if err == nil {
		var blob fileIdentityBlob
		if err := json.Unmarshal(b, &blob); err != nil {
			return nil, fmt.Errorf("parse identity %s: %w", path, err)
		}
		return decodeFileIdentity(&blob)
	}
	if !os.IsNotExist(err) {
		return nil, fmt.Errorf("read identity %s: %w", path, err)
	}

	xPriv, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generate x25519: %w", err)
	}
	_, edPriv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generate ed25519: %w", err)
	}

	blob := fileIdentityBlob{
		X25519Private:  xPriv.Bytes(),
		Ed25519Private: []byte(edPriv),
	}
	out, err := json.MarshalIndent(blob, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal identity: %w", err)
	}
	if err := os.WriteFile(path, out, 0o600); err != nil {
		return nil, fmt.Errorf("write identity %s: %w", path, err)
	}

	return decodeFileIdentity(&blob)
}

func decodeFileIdentity(blob *fileIdentityBlob) (*fileIdentity, error) {
	if len(blob.X25519Private) != 32 {
		return nil, fmt.Errorf("identity x25519 private: want 32 bytes, got %d", len(blob.X25519Private))
	}
	if len(blob.Ed25519Private) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("identity ed25519 private: want %d bytes, got %d", ed25519.PrivateKeySize, len(blob.Ed25519Private))
	}
	xPriv, err := ecdh.X25519().NewPrivateKey(blob.X25519Private)
	if err != nil {
		return nil, fmt.Errorf("parse x25519: %w", err)
	}
	return &fileIdentity{
		xPriv:  xPriv,
		edPriv: ed25519.PrivateKey(blob.Ed25519Private),
		xPub:   xPriv.PublicKey().Bytes(),
	}, nil
}

func (f *fileIdentity) PublicKey() []byte { return f.xPub }

func (f *fileIdentity) InstanceID() string {
	sum := sha256.Sum256(f.xPub)
	return base64.RawURLEncoding.EncodeToString(sum[:16])
}

func (f *fileIdentity) DeriveSharedSecret(peerPub, info []byte) ([]byte, error) {
	pub, err := ecdh.X25519().NewPublicKey(peerPub)
	if err != nil {
		return nil, fmt.Errorf("parse peer public key: %w", err)
	}
	shared, err := f.xPriv.ECDH(pub)
	if err != nil {
		return nil, err
	}
	return hkdfDerive(shared, info)
}

func (f *fileIdentity) Sign(msg []byte) ([]byte, error) {
	return ed25519.Sign(f.edPriv, msg), nil
}

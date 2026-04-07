// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Package crypto provides end-to-end encryption for traffic relayed
// through pigeon. The relay sees only ciphertext.
//
// Key exchange uses ECDH (X25519). Symmetric encryption uses AES-256-GCM
// with a monotonic counter nonce. Session keys are derived via HKDF-SHA256.
package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdh"
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"golang.org/x/crypto/hkdf"
)

// KeyPair holds an ECDH X25519 key pair for key exchange.
type KeyPair struct {
	Private *ecdh.PrivateKey
	Public  *ecdh.PublicKey
}

// GenerateKeyPair creates a new X25519 key pair.
func GenerateKeyPair() (*KeyPair, error) {
	priv, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	return &KeyPair{Private: priv, Public: priv.PublicKey()}, nil
}

// DeriveSessionKey performs ECDH with the peer's public key and derives
// a 256-bit AES key via HKDF-SHA256.
func DeriveSessionKey(priv *ecdh.PrivateKey, peerPub *ecdh.PublicKey, info []byte) ([]byte, error) {
	shared, err := priv.ECDH(peerPub)
	if err != nil {
		return nil, err
	}
	return hkdfDerive(shared, info)
}

// DeriveKeyFromSecret derives a session key from a persistent secret
// and a random nonce via HKDF-SHA256.
func DeriveKeyFromSecret(secret, nonce []byte) ([]byte, error) {
	return hkdfDerive(secret, nonce)
}

func hkdfDerive(ikm, info []byte) ([]byte, error) {
	r := hkdf.New(sha256.New, ikm, nil, info)
	key := make([]byte, 32)
	if _, err := r.Read(key); err != nil {
		return nil, err
	}
	return key, nil
}

// GenerateNonce creates a random 32-byte nonce for session key derivation.
func GenerateNonce() ([]byte, error) {
	nonce := make([]byte, 32)
	_, err := rand.Read(nonce)
	return nonce, err
}

// GenerateSecret creates a random 32-byte persistent device secret.
func GenerateSecret() ([]byte, error) {
	secret := make([]byte, 32)
	_, err := rand.Read(secret)
	return secret, err
}

// DeriveConfirmationCode computes a 6-digit confirmation code from
// both ECDH public keys. The code is deterministic and
// order-independent: DeriveConfirmationCode(a, b) == DeriveConfirmationCode(b, a).
//
// Both sides of the key exchange compute this independently. Under
// honest conditions, both derive the same code. Under a MitM attack
// (where the adversary substituted its own public key), each side
// computes a different code — the mismatch aborts pairing.
func DeriveConfirmationCode(pubA, pubB *ecdh.PublicKey) (string, error) {
	a, b := pubA.Bytes(), pubB.Bytes()
	if bytes.Compare(a, b) > 0 {
		a, b = b, a
	}
	ikm := append(a, b...)
	r := hkdf.New(sha256.New, ikm, nil, []byte("pairing-confirmation"))
	buf := make([]byte, 4)
	if _, err := r.Read(buf); err != nil {
		return "", err
	}
	code := binary.BigEndian.Uint32(buf) % 1000000
	return fmt.Sprintf("%06d", code), nil
}

// ChannelMode controls how Decrypt handles sequence numbers.
type ChannelMode int

const (
	// ModeStrict (default) requires sequence numbers to be strictly
	// monotonic with no gaps. Any out-of-order or replayed packet is
	// rejected. Suitable for reliable transports (streams).
	ModeStrict ChannelMode = iota

	// ModeDatagrams allows gaps in the sequence number space, which
	// is expected on lossy transports (UDP, H.264 video streams).
	// Any packet with a sequence number greater than the last received
	// is accepted and the counter jumps forward. Packets with a
	// sequence number less than or equal to the last received are
	// rejected to prevent replay attacks.
	ModeDatagrams
)

// Channel provides symmetric encryption/decryption for a transport
// connection. Uses AES-256-GCM with a monotonic counter nonce to
// prevent nonce reuse.
type Channel struct {
	sendGCM cipher.AEAD
	recvGCM cipher.AEAD
	sendSeq atomic.Uint64
	recvSeq uint64
	recvMu  sync.Mutex
	mode    ChannelMode
}

// NewChannel creates an encrypted channel from a shared key.
// sendKey and recvKey should be different (derived with different info
// strings) to prevent nonce collision on the two directions.
func NewChannel(sendKey, recvKey []byte) (*Channel, error) {
	sendBlock, err := aes.NewCipher(sendKey)
	if err != nil {
		return nil, err
	}
	sendGCM, err := cipher.NewGCM(sendBlock)
	if err != nil {
		return nil, err
	}

	recvBlock, err := aes.NewCipher(recvKey)
	if err != nil {
		return nil, err
	}
	recvGCM, err := cipher.NewGCM(recvBlock)
	if err != nil {
		return nil, err
	}

	return &Channel{sendGCM: sendGCM, recvGCM: recvGCM}, nil
}

// NewDatagramChannel creates a channel in ModeDatagrams. It is a
// convenience wrapper around NewChannel for lossy transports such as
// UDP or H.264 video streams.
func NewDatagramChannel(sendKey, recvKey []byte) (*Channel, error) {
	ch, err := NewChannel(sendKey, recvKey)
	if err != nil {
		return nil, err
	}
	ch.mode = ModeDatagrams
	return ch, nil
}

// SetMode changes the sequence-number checking behaviour of Decrypt.
// It is safe to call before any Decrypt calls; calling it concurrently
// with Decrypt is not supported.
func (c *Channel) SetMode(mode ChannelMode) {
	c.recvMu.Lock()
	defer c.recvMu.Unlock()
	c.mode = mode
}

// NewSymmetricChannel creates a channel where both directions use the
// same key but different nonce prefixes (0x00 for client→server,
// 0x01 for server→client). Use this for the persistent secret path.
func NewSymmetricChannel(key []byte, isServer bool) (*Channel, error) {
	sendInfo := []byte("client-to-server")
	recvInfo := []byte("server-to-client")
	if isServer {
		sendInfo, recvInfo = recvInfo, sendInfo
	}

	sendKey, err := hkdfDerive(key, sendInfo)
	if err != nil {
		return nil, err
	}
	recvKey, err := hkdfDerive(key, recvInfo)
	if err != nil {
		return nil, err
	}

	return NewChannel(sendKey, recvKey)
}

// Encrypt encrypts a plaintext message. The returned ciphertext includes
// the 8-byte sequence number prefix.
func (c *Channel) Encrypt(plaintext []byte) []byte {
	seq := c.sendSeq.Add(1) - 1
	nonce := makeNonce(c.sendGCM.NonceSize(), seq)
	seqBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(seqBytes, seq)
	ciphertext := c.sendGCM.Seal(nil, nonce, plaintext, seqBytes)
	return append(seqBytes, ciphertext...)
}

// Decrypt decrypts a ciphertext message. Verifies the sequence number
// to prevent replay attacks.
func (c *Channel) Decrypt(data []byte) ([]byte, error) {
	if len(data) < 8 {
		return nil, errors.New("ciphertext too short")
	}

	seqBytes := data[:8]
	seq := binary.LittleEndian.Uint64(seqBytes)
	ciphertext := data[8:]

	c.recvMu.Lock()
	defer c.recvMu.Unlock()

	switch c.mode {
	case ModeStrict:
		if seq != c.recvSeq {
			return nil, errors.New("unexpected sequence number")
		}
		c.recvSeq++
	case ModeDatagrams:
		// c.recvSeq holds the highest accepted seq + 1 (or 0 if none yet).
		// Accept any seq >= c.recvSeq (gaps allowed); reject seq < c.recvSeq
		// (replay protection — we have already received something at or after
		// this position).
		if seq < c.recvSeq {
			return nil, errors.New("sequence number replayed or too old")
		}
		c.recvSeq = seq + 1
	}

	nonce := makeNonce(c.recvGCM.NonceSize(), seq)
	return c.recvGCM.Open(nil, nonce, ciphertext, seqBytes)
}

func makeNonce(size int, seq uint64) []byte {
	nonce := make([]byte, size)
	binary.LittleEndian.PutUint64(nonce, seq)
	return nonce
}

// PairingRecord holds the persistent state from a completed pairing
// ceremony. Serialize this (e.g., JSON) and store it securely. On
// reconnect, load it and call DeriveChannel() to derive session keys
// without repeating the ECDH ceremony.
type PairingRecord struct {
	PeerInstanceID  string `json:"peer_instance_id"`
	RelayURL        string `json:"relay_url"`
	LocalPrivateKey []byte `json:"local_private_key"` // raw X25519, 32 bytes
	LocalPublicKey  []byte `json:"local_public_key"`  // raw X25519, 32 bytes
	PeerPublicKey   []byte `json:"peer_public_key"`   // raw X25519, 32 bytes
}

// NewPairingRecord creates a record from a completed pairing ceremony.
func NewPairingRecord(peerInstanceID, relayURL string, localKP *KeyPair, peerPubKey *ecdh.PublicKey) *PairingRecord {
	return &PairingRecord{
		PeerInstanceID:  peerInstanceID,
		RelayURL:        relayURL,
		LocalPrivateKey: localKP.Private.Bytes(),
		LocalPublicKey:  localKP.Public.Bytes(),
		PeerPublicKey:   peerPubKey.Bytes(),
	}
}

// DeriveChannel derives an encrypted channel from the stored keys.
// The info parameters should match what was used during the original pairing
// (e.g., "server-to-client" and "client-to-server").
func (r *PairingRecord) DeriveChannel(sendInfo, recvInfo []byte) (*Channel, error) {
	priv, err := ecdh.X25519().NewPrivateKey(r.LocalPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("parse local key: %w", err)
	}
	peerPub, err := ecdh.X25519().NewPublicKey(r.PeerPublicKey)
	if err != nil {
		return nil, fmt.Errorf("parse peer key: %w", err)
	}

	sendKey, err := DeriveSessionKey(priv, peerPub, sendInfo)
	if err != nil {
		return nil, fmt.Errorf("derive send key: %w", err)
	}
	recvKey, err := DeriveSessionKey(priv, peerPub, recvInfo)
	if err != nil {
		return nil, fmt.Errorf("derive recv key: %w", err)
	}

	return NewChannel(sendKey, recvKey)
}

// Marshal serializes the record to JSON.
func (r *PairingRecord) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

// UnmarshalPairingRecord deserializes a record from JSON.
func UnmarshalPairingRecord(data []byte) (*PairingRecord, error) {
	var r PairingRecord
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, err
	}
	return &r, nil
}

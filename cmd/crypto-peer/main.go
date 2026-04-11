// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// crypto-peer registers with a pigeon relay, generates an X25519
// keypair, and performs a key exchange with the connecting client.
// After exchanging public keys, both sides independently derive a
// 6-digit confirmation code. The peer sends its code so the client
// can verify cross-language agreement.
//
// Usage:
//
//	crypto-peer <relay-url>
//
// Protocol (length-prefixed messages over QUIC stream):
//  1. peer → client: 32-byte X25519 public key
//  2. client → peer: 32-byte X25519 public key
//  3. peer → client: 6-byte ASCII confirmation code
//
// The instance ID is printed to stderr so the client can connect.
// Set PIGEON_INSECURE=1 for self-signed relay certificates.
package main

import (
	"context"
	"crypto/ecdh"
	"crypto/tls"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/marcelocantos/pigeon"
	"github.com/marcelocantos/pigeon/crypto"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: crypto-peer <relay-url>")
		os.Exit(1)
	}

	relayURL := os.Args[1]

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tlsConfig := &tls.Config{
		InsecureSkipVerify: os.Getenv("PIGEON_INSECURE") == "1",
	}

	var quicPort string
	if u, err := url.Parse(relayURL); err == nil {
		if p := u.Port(); p != "" {
			quicPort = p
		}
	}

	conn, err := pigeon.Register(ctx, relayURL, pigeon.Config{
		TLS:      tlsConfig,
		QUICPort: quicPort,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "register: %v\n", err)
		os.Exit(1)
	}
	defer conn.CloseNow()

	// Print instance ID so the client can connect.
	fmt.Fprintln(os.Stderr, conn.InstanceID())

	// Generate X25519 keypair.
	kp, err := crypto.GenerateKeyPair()
	if err != nil {
		fmt.Fprintf(os.Stderr, "keygen: %v\n", err)
		os.Exit(1)
	}

	// Send our public key (32 bytes).
	if err := conn.Send(ctx, kp.Public.Bytes()); err != nil {
		fmt.Fprintf(os.Stderr, "send pubkey: %v\n", err)
		os.Exit(1)
	}

	// Receive client's public key (32 bytes).
	peerPubBytes, err := conn.Recv(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "recv pubkey: %v\n", err)
		os.Exit(1)
	}

	peerPub, err := ecdh.X25519().NewPublicKey(peerPubBytes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "parse pubkey: %v\n", err)
		os.Exit(1)
	}

	// Derive confirmation code.
	code, err := crypto.DeriveConfirmationCode(kp.Public, peerPub)
	if err != nil {
		fmt.Fprintf(os.Stderr, "derive code: %v\n", err)
		os.Exit(1)
	}

	// Send confirmation code.
	if err := conn.Send(ctx, []byte(code)); err != nil {
		fmt.Fprintf(os.Stderr, "send code: %v\n", err)
		os.Exit(1)
	}

	// Wait for the client to close the connection before exiting.
	// Without this, CloseNow races with the relay forwarding the
	// confirmation code — the QUIC CONNECTION_CLOSE frame can arrive
	// at the relay before the stream data is bridged to the client.
	conn.Recv(ctx) //nolint:errcheck
}

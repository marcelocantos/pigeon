// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// crypto-peer registers with a pigeon relay (pairing-mode), generates
// an X25519 keypair, and performs a key exchange with the connecting
// client over a single named "crypto" stream. After exchanging public
// keys, both sides independently derive a 6-digit confirmation code;
// the peer sends its code so the client can verify cross-language
// agreement.
//
// Wire shape (modern, post-T39.6.1): the client opens a sub-stream
// named "crypto" on its pigeon Session; crypto-peer accepts that
// stream and runs three length-prefixed exchanges over it:
//
//  1. peer → client: 32-byte X25519 public key
//  2. client → peer: 32-byte X25519 public key
//  3. peer → client: 6-byte ASCII confirmation code
//
// Usage:
//
//	crypto-peer <relay-url>
//
// The instance ID is printed to stdout so the client can connect.
// Set PIGEON_INSECURE=1 for self-signed relay certificates.
package main

import (
	"context"
	"crypto/ecdh"
	"crypto/tls"
	"fmt"
	"os"
	"time"

	"github.com/marcelocantos/pigeon"
	"github.com/marcelocantos/pigeon/crypto"
)

const cryptoStreamName = "crypto"

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

	// pairing-mode Register (no Pairing callback) gives us a Session
	// per accepted client without requiring a PairingRecord — the
	// cross-language confirmation-code test does its own crypto.
	listener, instanceID, err := pigeon.Register(ctx, &pigeon.RegisterArgs{
		Relay: relayURL,
		TLS:   tlsConfig,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "register: %v\n", err)
		os.Exit(1)
	}
	defer listener.Close()

	// Print instance ID to stdout so the client can connect. stderr is
	// reserved for slog / quic-go diagnostics, which may otherwise
	// interleave with the ID and confuse consumers.
	fmt.Println(instanceID)

	sess, err := listener.Accept(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "accept: %v\n", err)
		os.Exit(1)
	}
	defer sess.Close()

	// crypto-peer's cross-language test fixtures (Swift NWConnection,
	// raw kotlin/etc.) talk a single QUIC stream — the primary —
	// rather than opening a sub-stream. Pairing-mode skips activation,
	// so the primary is free for the crypto exchange.
	stream := sess.Primary()
	_ = cryptoStreamName // reserved for the named-sub-stream variant if/when consumers gain multi-stream QUIC support

	// Generate X25519 keypair.
	kp, err := crypto.GenerateKeyPair()
	if err != nil {
		fmt.Fprintf(os.Stderr, "keygen: %v\n", err)
		os.Exit(1)
	}

	// Send our public key (32 bytes).
	if err := stream.Send(kp.Public.Bytes()); err != nil {
		fmt.Fprintf(os.Stderr, "send pubkey: %v\n", err)
		os.Exit(1)
	}

	// Receive client's public key (32 bytes).
	peerPubBytes, err := stream.Recv(ctx)
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
	if err := stream.Send([]byte(code)); err != nil {
		fmt.Fprintf(os.Stderr, "send code: %v\n", err)
		os.Exit(1)
	}

	// Wait for the client to close the stream before exiting; without
	// this, Session.Close races with the relay forwarding the
	// confirmation code — the QUIC CONNECTION_CLOSE frame can arrive
	// at the relay before the stream data is bridged to the client.
	stream.Recv(ctx) //nolint:errcheck
}

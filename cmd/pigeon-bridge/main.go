// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// pigeon-bridge connects to a pigeon relay via QUIC and bridges messages
// to/from stdin/stdout using length-prefixed framing. Intended for
// driving E2E tests from languages without native QUIC support.
//
// Usage:
//   pigeon-bridge register <relay-url> [token]
//   pigeon-bridge connect <relay-url> <instance-id>
//
// Once connected, reads length-prefixed messages from stdin and sends
// them to the relay. Messages from the relay are written to stdout
// with length-prefixed framing.
package main

import (
	"context"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/marcelocantos/pigeon"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "usage: pigeon-bridge register <url> [token]")
		fmt.Fprintln(os.Stderr, "       pigeon-bridge connect <url> <instance-id>")
		os.Exit(1)
	}

	cmd := os.Args[1]
	url := os.Args[2]

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tlsConfig := &tls.Config{
		InsecureSkipVerify: os.Getenv("PIGEON_INSECURE") == "1",
	}

	var conn *pigeon.Conn
	var err error

	switch cmd {
	case "register":
		cfg := pigeon.Config{TLS: tlsConfig}
		if len(os.Args) > 3 && os.Args[3] != "" {
			cfg.Token = os.Args[3]
		}
		conn, err = pigeon.Register(ctx, url, cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "register: %v\n", err)
			os.Exit(1)
		}
		// Write instance ID to stdout as a length-prefixed message.
		writeStdout([]byte(conn.InstanceID()))

	case "connect":
		if len(os.Args) < 4 {
			fmt.Fprintln(os.Stderr, "connect requires instance-id")
			os.Exit(1)
		}
		instanceID := os.Args[3]
		conn, err = pigeon.Connect(ctx, url, instanceID, pigeon.Config{TLS: tlsConfig})
		if err != nil {
			fmt.Fprintf(os.Stderr, "connect: %v\n", err)
			os.Exit(1)
		}

	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", cmd)
		os.Exit(1)
	}

	defer conn.CloseNow()

	// Bridge: relay → stdout. Exit the process when the relay closes so
	// the stdin loop doesn't hang indefinitely waiting for more input.
	go func() {
		for {
			data, err := conn.Recv(ctx)
			if err != nil {
				os.Exit(0)
				return
			}
			writeStdout(data)
		}
	}()

	// Bridge: stdin → relay
	for {
		data, err := readStdin()
		if err != nil {
			return
		}
		if err := conn.Send(ctx, data); err != nil {
			return
		}
	}
}

func writeStdout(data []byte) {
	var hdr [4]byte
	binary.BigEndian.PutUint32(hdr[:], uint32(len(data)))
	os.Stdout.Write(hdr[:])
	os.Stdout.Write(data)
}

func readStdin() ([]byte, error) {
	var hdr [4]byte
	if _, err := io.ReadFull(os.Stdin, hdr[:]); err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint32(hdr[:])
	if length > 1<<20 {
		return nil, fmt.Errorf("message too large: %d bytes", length)
	}
	buf := make([]byte, length)
	if _, err := io.ReadFull(os.Stdin, buf); err != nil {
		return nil, err
	}
	return buf, nil
}

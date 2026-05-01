// echo-client — connects to a paired backend and demonstrates the
// new multi-channel pigeon API:
//   - "chat" stream: send each stdin line, print the backend's echo.
//   - "control" stream: type "/stats" to ask the backend its echo count.
//   - "ping" datagram: type "/ping <text>" to send a one-shot ping.
//   - "metric" datagram: type "/metric" to fetch the backend's count.
//
// Pair new backends with the sibling `pair` subcommand
// (./client pair --token ...).
//
// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/marcelocantos/pigeon"
	"github.com/marcelocantos/pigeon/crypto"
)

// dgChannels matches the backend's declarations.
var dgChannels = map[string]uint64{
	"ping":   1,
	"metric": 2,
}

// peerEntry bundles what the client needs to reach a paired backend.
type peerEntry struct {
	InstanceID string                `json:"instance_id"`
	Record     *crypto.PairingRecord `json:"record"`
}

func loadPeers(path string) (map[string]*peerEntry, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]*peerEntry{}, nil
		}
		return nil, err
	}
	out := map[string]*peerEntry{}
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "pair" {
		pairCmd(os.Args[2:])
		return
	}

	store := flag.String("store", "peers.json", "path to persisted peer entries")
	idPath := flag.String("identity", "identity.json", "client identity keypair")
	peer := flag.String("peer", "", "name of the paired backend to connect to")
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if *peer == "" {
		slog.Error("--peer is required (or use `./client pair` to add a new one)")
		os.Exit(1)
	}

	identity, err := crypto.NewFileIdentity(*idPath)
	if err != nil {
		slog.Error("identity", "err", err)
		os.Exit(1)
	}

	peers, err := loadPeers(*store)
	if err != nil {
		slog.Error("load store", "err", err)
		os.Exit(1)
	}
	entry, ok := peers[*peer]
	if !ok {
		slog.Error("unknown peer", "name", *peer)
		os.Exit(1)
	}

	sess, err := pigeon.Connect(ctx, &pigeon.ConnectArgs{
		InstanceID: entry.InstanceID,
		Record:     entry.Record,
		Identity:   identity,
		Relay:      entry.Record.RelayURL,
		Datagrams:  dgChannels,
	})
	if err != nil {
		slog.Error("connect", "err", err)
		os.Exit(1)
	}
	defer sess.Close()

	chat, err := sess.OpenStream(ctx, "chat")
	if err != nil {
		slog.Error("open chat", "err", err)
		os.Exit(1)
	}
	ctrl, err := sess.OpenStream(ctx, "control")
	if err != nil {
		slog.Error("open control", "err", err)
		os.Exit(1)
	}
	ping := sess.Datagram("ping")
	metric := sess.Datagram("metric")

	// Reader goroutines for each incoming channel.
	go func() {
		for {
			msg, err := chat.Recv(ctx)
			if err != nil {
				return
			}
			fmt.Printf("[chat] %s\n", msg)
		}
	}()
	go func() {
		for {
			msg, err := ctrl.Recv(ctx)
			if err != nil {
				return
			}
			fmt.Printf("[ctrl] %s\n", msg)
		}
	}()
	go func() {
		for {
			p, err := ping.Recv(ctx)
			if err != nil {
				return
			}
			fmt.Printf("[ping] %s\n", p)
		}
	}()
	go func() {
		for {
			m, err := metric.Recv(ctx)
			if err != nil {
				return
			}
			fmt.Printf("[metric] %s\n", m)
		}
	}()

	fmt.Println("Connected. Type a chat message, or use one of:")
	fmt.Println("  /stats        — ask backend its echo count (control stream)")
	fmt.Println("  /ping <text>  — send a one-shot ping datagram")
	fmt.Println("  /metric       — request a metric datagram")

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case line == "/stats":
			if err := ctrl.Send([]byte("stats")); err != nil {
				slog.Warn("ctrl send", "err", err)
				return
			}
		case strings.HasPrefix(line, "/ping "):
			if err := ping.Send([]byte(strings.TrimPrefix(line, "/ping "))); err != nil {
				slog.Warn("ping send", "err", err)
				return
			}
		case line == "/metric":
			if err := metric.Send([]byte("get")); err != nil {
				slog.Warn("metric send", "err", err)
				return
			}
		default:
			if err := chat.Send([]byte(line)); err != nil {
				slog.Warn("chat send", "err", err)
				return
			}
		}
	}
	if err := scanner.Err(); err != nil {
		slog.Warn("stdin", "err", err)
	}
}

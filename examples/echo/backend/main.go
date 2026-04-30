// echo-backend daemon — accepts pigeon connections from paired clients
// and serves them on:
//   - "chat" stream: echoes each message back with an "echo: " prefix
//   - "control" stream: replies to "stats" with the running echo count
//   - "ping" datagram channel (id 1): echoes pings as "pong:..."
//   - "metric" datagram channel (id 2): replies with "echoes=N"
//
// Pair new clients with the sibling `pair` subcommand (./backend pair);
// the running daemon picks up new entries via fsnotify on pairings.json.
//
// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"sync/atomic"
	"syscall"

	"github.com/fsnotify/fsnotify"
	"github.com/marcelocantos/pigeon"
	"github.com/marcelocantos/pigeon/crypto"
)

// dgChannels declares the datagram channels both peers must agree on.
// Channel IDs are part of the wire protocol — changing them is a
// breaking change for clients in the field.
var dgChannels = map[string]uint64{
	"ping":   1,
	"metric": 2,
}

func loadPairings(path string) (map[string]*crypto.PairingRecord, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]*crypto.PairingRecord{}, nil
		}
		return nil, err
	}
	out := map[string]*crypto.PairingRecord{}
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

	relay := flag.String("relay", "relay.example.com", "pigeon relay address")
	store := flag.String("store", "pairings.json", "path to persisted PairingRecords")
	idPath := flag.String("identity", "identity.json", "backend identity keypair")
	flag.Parse()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	identity, err := crypto.NewFileIdentity(*idPath)
	if err != nil {
		slog.Error("identity", "err", err)
		os.Exit(1)
	}

	pairings, err := loadPairings(*store)
	if err != nil {
		slog.Error("load store", "err", err)
		os.Exit(1)
	}
	slog.Info("loaded pairings", "total", len(pairings))

	var mu sync.RWMutex

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		slog.Error("watcher", "err", err)
		os.Exit(1)
	}
	defer watcher.Close()
	dir, _ := filepath.Abs(filepath.Dir(*store))
	if err := watcher.Add(dir); err != nil {
		slog.Error("watch", "err", err)
		os.Exit(1)
	}
	target, _ := filepath.Abs(*store)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case ev := <-watcher.Events:
				if ev.Name != target {
					continue
				}
				if ev.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename) == 0 {
					continue
				}
				next, err := loadPairings(*store)
				if err != nil {
					slog.Warn("reload pairings", "err", err)
					continue
				}
				mu.Lock()
				pairings = next
				mu.Unlock()
				slog.Info("pairings reloaded", "total", len(next))
			case err := <-watcher.Errors:
				slog.Warn("watcher error", "err", err)
			}
		}
	}()

	lis, id, err := pigeon.Register(ctx, &pigeon.RegisterArgs{
		Identity: identity,
		Pairing: func(clientID string) (*crypto.PairingRecord, error) {
			mu.RLock()
			defer mu.RUnlock()
			rec, ok := pairings[clientID]
			if !ok {
				return nil, fmt.Errorf("unknown client %q", clientID)
			}
			return rec, nil
		},
		Relay:     *relay,
		Datagrams: dgChannels,
	})
	if err != nil {
		slog.Error("register", "err", err)
		os.Exit(1)
	}
	defer lis.Close()
	slog.Info("backend ready", "instance_id", id)

	for {
		sess, err := lis.Accept(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			slog.Warn("accept", "err", err)
			continue
		}
		go serveClient(ctx, sess)
	}
}

// serveClient handles all four channels for one connected client.
func serveClient(ctx context.Context, sess *pigeon.Session) {
	defer sess.Close()
	peer := sess.PeerID()
	slog.Info("client connected", "peer", peer)
	defer slog.Info("client disconnected", "peer", peer)

	var echoes atomic.Int64

	// chat: echo with prefix.
	go func() {
		chat, err := sess.AcceptStream(ctx, "chat")
		if err != nil {
			return
		}
		for {
			msg, err := chat.Recv(ctx)
			if err != nil {
				return
			}
			if err := chat.Send(append([]byte("echo: "), msg...)); err != nil {
				return
			}
			echoes.Add(1)
		}
	}()

	// control: reply to "stats" with the current echo count.
	go func() {
		ctrl, err := sess.AcceptStream(ctx, "control")
		if err != nil {
			return
		}
		for {
			msg, err := ctrl.Recv(ctx)
			if err != nil {
				return
			}
			switch string(msg) {
			case "stats":
				_ = ctrl.Send(fmt.Appendf(nil, "echoes=%d", echoes.Load()))
			default:
				_ = ctrl.Send(fmt.Appendf(nil, "unknown command: %s", msg))
			}
		}
	}()

	// ping: echo as "pong:<payload>", count toward echoes.
	go func() {
		ping := sess.Datagram("ping")
		for {
			p, err := ping.Recv(ctx)
			if err != nil {
				return
			}
			_ = ping.Send(append([]byte("pong:"), p...))
			echoes.Add(1)
		}
	}()

	// metric: ignore inbound payload, reply with current echo count.
	go func() {
		metric := sess.Datagram("metric")
		for {
			_, err := metric.Recv(ctx)
			if err != nil {
				return
			}
			_ = metric.Send(fmt.Appendf(nil, "echoes=%d", echoes.Load()))
		}
	}()

	<-ctx.Done()
}

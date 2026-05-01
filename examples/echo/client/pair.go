// echo-client `pair` subcommand — runs one initiator-side pairing
// ceremony against a token produced by `./backend pair`, prompts for
// confirmation, and writes the resulting peer entry to peers.json.
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

	"github.com/marcelocantos/pigeon/crypto"
	"github.com/marcelocantos/pigeon/pairing"
)

func pairCmd(args []string) {
	fs := flag.NewFlagSet("pair", flag.ExitOnError)
	store := fs.String("store", "peers.json", "path to persisted peer entries")
	idPath := fs.String("identity", "identity.json", "client identity keypair")
	name := fs.String("name", "", "name to store the new pairing under")
	token := fs.String("token", "", "rendezvous token from `backend pair`")
	fs.Parse(args)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	if *name == "" || *token == "" {
		slog.Error("--name and --token are required")
		os.Exit(1)
	}

	identity, err := crypto.NewFileIdentity(*idPath)
	if err != nil {
		slog.Error("identity", "err", err)
		os.Exit(1)
	}

	cer, err := pairing.Initiate(ctx, &pairing.InitiateArgs{
		Rendezvous: *token,
		Identity:   identity,
	})
	if err != nil {
		slog.Error("initiate", "err", err)
		os.Exit(1)
	}
	defer cer.Close()

	code, err := cer.Code(ctx)
	if err != nil {
		slog.Error("derive code", "err", err)
		return
	}
	fmt.Printf("Confirmation code: %s\n", code)
	fmt.Print("Does the code match the one shown on the backend? [y/N] ")

	ans, _ := bufio.NewReader(os.Stdin).ReadString('\n')
	if strings.TrimSpace(strings.ToLower(ans)) != "y" {
		slog.Info("ceremony cancelled")
		return
	}

	rec, err := cer.Confirm(ctx)
	if err != nil {
		slog.Error("confirm ceremony", "err", err)
		return
	}

	peers, err := loadPeers(*store)
	if err != nil {
		slog.Error("load store", "err", err)
		return
	}
	peers[*name] = &peerEntry{
		InstanceID: rec.PeerInstanceID,
		Record:     rec,
	}
	b, _ := json.MarshalIndent(peers, "", "  ")
	if err := os.WriteFile(*store, b, 0o600); err != nil {
		slog.Error("save store", "err", err)
		return
	}
	slog.Info("paired", "name", *name, "peer", rec.PeerInstanceID, "total", len(peers))
}

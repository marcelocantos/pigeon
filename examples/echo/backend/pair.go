// echo-backend `pair` subcommand — runs one acceptor-side pairing
// ceremony, prints the rendezvous token (as a QR plus the raw string)
// for the user to transfer to the new device, prompts for confirmation,
// and appends the resulting PairingRecord to pairings.json.
//
// A running ./backend daemon picks up the new entry via fsnotify;
// no daemon restart is required.
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
	"github.com/marcelocantos/pigeon/qr"
)

func pairCmd(args []string) {
	fs := flag.NewFlagSet("pair", flag.ExitOnError)
	relay := fs.String("relay", "relay.example.com", "pigeon relay address")
	store := fs.String("store", "pairings.json", "path to persisted PairingRecords")
	idPath := fs.String("identity", "identity.json", "backend identity keypair")
	fs.Parse(args)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	identity, err := crypto.NewFileIdentity(*idPath)
	if err != nil {
		slog.Error("identity", "err", err)
		os.Exit(1)
	}

	pairer, err := pairing.Register(ctx, &pairing.Args{
		Relay:    *relay,
		Identity: identity,
	})
	if err != nil {
		slog.Error("register pairer", "err", err)
		os.Exit(1)
	}
	defer pairer.Close()

	cer, err := pairer.Accept(ctx)
	if err != nil {
		slog.Error("accept ceremony", "err", err)
		return
	}
	defer cer.Close()

	fmt.Println("Show this on the new device, or paste the token below:")
	qr.Print(os.Stdout, cer.Token)
	fmt.Println()
	fmt.Println(cer.Token)
	fmt.Println()

	code, err := cer.Code(ctx)
	if err != nil {
		slog.Error("derive code", "err", err)
		return
	}
	fmt.Printf("Confirmation code: %s\n", code)
	fmt.Print("Does the code match the one shown on the new device? [y/N] ")

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

	pairings, err := loadPairings(*store)
	if err != nil {
		slog.Error("load store", "err", err)
		return
	}
	pairings[rec.PeerInstanceID] = rec
	b, _ := json.MarshalIndent(pairings, "", "  ")
	if err := os.WriteFile(*store, b, 0o600); err != nil {
		slog.Error("save store", "err", err)
		return
	}
	slog.Info("paired", "peer", rec.PeerInstanceID, "total", len(pairings))
}

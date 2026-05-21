// echo-backend pairing wiring.
//
// Two halves live here:
//
//   - pairCmd: the user-facing `./backend pair` CLI. It dials the
//     running daemon's local backchannel socket and drives the
//     ceremony end-to-end over typed messages — no in-process pairer,
//     no file-handoff. Replaces the old fsnotify(pairings.json) path.
//
//   - handlePairingCLI: the daemon-side handler for one accepted
//     backchannel session. Runs a single pairer.Accept ceremony,
//     emits token_response and waiting_for_code back to the CLI, waits
//     for code_submit, persists the new PairingRecord into the
//     in-memory map and pairings.json, and sends pair_status before
//     closing.
//
// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/marcelocantos/pigeon/backchannel"
	"github.com/marcelocantos/pigeon/crypto"
	"github.com/marcelocantos/pigeon/pairing"
	"github.com/marcelocantos/pigeon/qr"
)

// pairCmd is `./backend pair`. It contacts the running daemon over
// the local backchannel and prints the token / code the user needs.
func pairCmd(args []string) {
	fs := flag.NewFlagSet("pair", flag.ExitOnError)
	fs.Parse(args)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	cli, err := backchannel.Dial(&backchannel.DialArgs{AppName: appName})
	if err != nil {
		slog.Error("dial daemon (is ./backend running?)", "err", err)
		os.Exit(1)
	}
	defer cli.Close()

	if err := cli.Send(ctx, backchannel.Message{Type: backchannel.MsgPairBegin}); err != nil {
		slog.Error("send pair_begin", "err", err)
		return
	}

	// 1) token_response: daemon has registered with the relay and
	//    issued a rendezvous token. Display it to the user.
	tokenMsg, err := cli.Recv(ctx)
	if err != nil {
		slog.Error("recv token_response", "err", err)
		return
	}
	if tokenMsg.Type != backchannel.MsgTokenResponse {
		slog.Error("unexpected daemon reply", "type", tokenMsg.Type)
		return
	}
	fmt.Println("Show this on the new device, or paste the token below:")
	qr.Print(os.Stdout, tokenMsg.Token)
	fmt.Println()
	fmt.Println(tokenMsg.Token)
	fmt.Println()

	// 2) waiting_for_code: peer has connected, both sides have derived
	//    the 6-digit confirmation code. The daemon carries the code in
	//    the message's Code field.
	codeMsg, err := cli.Recv(ctx)
	if err != nil {
		slog.Error("recv waiting_for_code", "err", err)
		return
	}
	if codeMsg.Type != backchannel.MsgWaitingForCode {
		slog.Error("unexpected daemon reply", "type", codeMsg.Type)
		return
	}
	fmt.Printf("Confirmation code: %s\n", codeMsg.Code)
	fmt.Print("Does the code match the one shown on the new device? [y/N] ")

	var ans string
	fmt.Fscanln(os.Stdin, &ans)
	confirmed := ans == "y" || ans == "Y"

	// 3) code_submit: send the user's decision back. status="ok" means
	//    "yes, proceed"; status="abort" means "no, cancel".
	status := "ok"
	if !confirmed {
		status = "abort"
	}
	if err := cli.Send(ctx, backchannel.Message{Type: backchannel.MsgCodeSubmit, Status: status}); err != nil {
		slog.Error("send code_submit", "err", err)
		return
	}

	// 4) pair_status: terminal acknowledgement.
	final, err := cli.Recv(ctx)
	if err != nil {
		slog.Error("recv pair_status", "err", err)
		return
	}
	if final.Type != backchannel.MsgPairStatus {
		slog.Error("unexpected daemon reply", "type", final.Type)
		return
	}
	if final.Status == "paired" {
		slog.Info("paired", "peer", final.InstanceID)
	} else {
		slog.Info("pairing ended", "status", final.Status)
	}
}

// handlePairingCLI runs one acceptor-side ceremony on behalf of a
// connected CLI. Exits when the ceremony resolves (success or abort)
// or when sess closes.
func handlePairingCLI(
	ctx context.Context,
	sess *backchannel.Session,
	pairer *pairing.Pairer,
	mu *sync.RWMutex,
	pairings *map[string]*crypto.PairingRecord,
	storePath string,
) {
	defer sess.Close()

	begin, err := sess.Recv(ctx)
	if err != nil {
		slog.Warn("backchannel recv pair_begin", "err", err)
		return
	}
	if begin.Type != backchannel.MsgPairBegin {
		slog.Warn("backchannel: first message was not pair_begin", "type", begin.Type)
		_ = sess.Send(ctx, backchannel.Message{Type: backchannel.MsgPairStatus, Status: "protocol_error"})
		return
	}

	cer, err := pairer.Accept(ctx)
	if err != nil {
		slog.Warn("pairer accept", "err", err)
		_ = sess.Send(ctx, backchannel.Message{Type: backchannel.MsgPairStatus, Status: fmt.Sprintf("error: %v", err)})
		return
	}
	defer cer.Close()

	if err := sess.Send(ctx, backchannel.Message{
		Type:       backchannel.MsgTokenResponse,
		InstanceID: pairer.InstanceID(),
		Token:      cer.Token,
	}); err != nil {
		slog.Warn("send token_response", "err", err)
		return
	}

	code, err := cer.Code(ctx)
	if err != nil {
		_ = sess.Send(ctx, backchannel.Message{Type: backchannel.MsgPairStatus, Status: fmt.Sprintf("error: %v", err)})
		return
	}

	if err := sess.Send(ctx, backchannel.Message{
		Type: backchannel.MsgWaitingForCode,
		Code: code,
	}); err != nil {
		slog.Warn("send waiting_for_code", "err", err)
		return
	}

	submit, err := sess.Recv(ctx)
	if err != nil {
		slog.Warn("recv code_submit", "err", err)
		return
	}
	if submit.Type != backchannel.MsgCodeSubmit {
		slog.Warn("expected code_submit", "type", submit.Type)
		_ = sess.Send(ctx, backchannel.Message{Type: backchannel.MsgPairStatus, Status: "protocol_error"})
		return
	}
	if submit.Status != "ok" {
		// User rejected — Close on cer aborts the ceremony.
		_ = sess.Send(ctx, backchannel.Message{Type: backchannel.MsgPairStatus, Status: "aborted"})
		return
	}

	rec, err := cer.Confirm(ctx)
	if err != nil {
		_ = sess.Send(ctx, backchannel.Message{Type: backchannel.MsgPairStatus, Status: fmt.Sprintf("error: %v", err)})
		return
	}

	mu.Lock()
	(*pairings)[rec.PeerInstanceID] = rec
	snap := make(map[string]*crypto.PairingRecord, len(*pairings))
	for k, v := range *pairings {
		snap[k] = v
	}
	mu.Unlock()

	if err := savePairings(storePath, snap); err != nil {
		slog.Warn("save store", "err", err)
		_ = sess.Send(ctx, backchannel.Message{Type: backchannel.MsgPairStatus, Status: fmt.Sprintf("save error: %v", err)})
		return
	}

	_ = sess.Send(ctx, backchannel.Message{
		Type:       backchannel.MsgPairStatus,
		Status:     "paired",
		InstanceID: rec.PeerInstanceID,
	})
	slog.Info("paired via backchannel", "peer", rec.PeerInstanceID, "total", len(snap))
}

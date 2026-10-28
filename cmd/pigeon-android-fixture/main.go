// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// pigeon-android-fixture is a one-shot LAN device oracle. The HTTP control
// endpoint is intentionally unauthenticated: bind it only on a trusted test LAN.
// It accepts exactly one fresh pairing and one encrypted chat exchange.
package main

import (
	"context"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/marcelocantos/pigeon"
	"github.com/marcelocantos/pigeon/crypto"
	"github.com/marcelocantos/pigeon/pairing"
)

type state struct {
	mu          sync.Mutex
	code        string
	codeReady   chan struct{}
	matched     chan struct{}
	ready       chan struct{}
	acked       chan struct{}
	ackReported chan struct{}
	result      chan error
}

func (s *state) checkCode(code string) error {
	select {
	case <-s.codeReady:
	case <-time.After(40 * time.Second):
		return errors.New("pairing SAS never arrived")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if code == "" || code != s.code {
		return errors.New("SAS mismatch")
	}
	log.Printf("DEVICE_ORACLE SAS_MATCH code=%s", code)
	select {
	case <-s.matched:
	default:
		close(s.matched)
	}
	return nil
}

// encryptedChat waits for the device's acknowledgement before the deferred stream
// and session closes. Recv is bounded by the fixture's hard deadline.
type chatExchange interface {
	Send([]byte) error
	Recv(context.Context) ([]byte, error)
}

func encryptedChat(ctx context.Context, chat chatExchange) error {
	msg, err := chat.Recv(ctx)
	if err != nil {
		return fmt.Errorf("chat recv: %w", err)
	}
	if string(msg) != "ping" {
		return fmt.Errorf("chat payload = %q, want ping", msg)
	}
	if err := chat.Send([]byte("echo: ping")); err != nil {
		return fmt.Errorf("chat send: %w", err)
	}
	ack, err := chat.Recv(ctx)
	if err != nil {
		return fmt.Errorf("chat ack recv: %w", err)
	}
	if string(ack) != "ack: echo: ping" {
		return fmt.Errorf("chat ack = %q, want ack: echo: ping", ack)
	}
	return nil
}

func run(ctx context.Context, relay string, s *state, token chan<- string) error {
	identityPath, err := os.CreateTemp("", "pigeon-device-identity-*.json")
	if err != nil {
		return err
	}
	path := identityPath.Name()
	identityPath.Close()
	os.Remove(path)
	defer os.Remove(path)
	identity, err := crypto.NewFileIdentity(path)
	if err != nil {
		return err
	}
	pairer, err := pairing.Register(ctx, &pairing.Args{Relay: relay, Identity: identity})
	if err != nil {
		return err
	}
	defer pairer.Close()
	cer, err := pairer.Accept(ctx)
	if err != nil {
		return fmt.Errorf("pair accept: %w", err)
	}
	defer cer.Close()
	token <- cer.Token
	code, err := cer.Code(ctx)
	if err != nil {
		return fmt.Errorf("pair code: %w", err)
	}
	s.mu.Lock()
	s.code = code
	close(s.codeReady)
	s.mu.Unlock()
	select {
	case <-s.matched:
	case <-ctx.Done():
		return ctx.Err()
	}
	record, err := cer.Confirm(ctx)
	if err != nil {
		return fmt.Errorf("pair confirm: %w", err)
	}
	// A fresh device must reconnect with its just-derived PairingRecord.
	listener, id, err := pigeon.Register(ctx, &pigeon.RegisterArgs{
		Identity: identity, Relay: relay, TLS: &tls.Config{InsecureSkipVerify: true},
		Streams: []string{"chat"}, Pairing: func(peer string) (*crypto.PairingRecord, error) {
			if peer != record.PeerInstanceID {
				return nil, fmt.Errorf("unpaired peer %q", peer)
			}
			return record, nil
		},
	})
	if err != nil {
		return fmt.Errorf("backend register: %w", err)
	}
	defer listener.Close()
	if id != identity.InstanceID() {
		return fmt.Errorf("backend instance ID changed: %s", id)
	}
	close(s.ready) // Backend listener is registered before device reconnect.
	sess, err := listener.Accept(ctx)
	if err != nil {
		return fmt.Errorf("backend accept: %w", err)
	}
	defer sess.Close()
	chat, err := sess.AcceptStream(ctx, "chat")
	if err != nil {
		return fmt.Errorf("chat accept: %w", err)
	}
	defer chat.Close()
	if err := encryptedChat(ctx, chat); err != nil {
		return err
	}
	close(s.acked)
	log.Printf("DEVICE_ORACLE encrypted chat ping -> echo: ping -> ack: echo: ping peer=%s", record.PeerInstanceID)
	return nil
}

func main() {
	relay := flag.String("relay", "", "LAN relay URL (https://host:udp-port)")
	control := flag.String("control", "127.0.0.1:0", "trusted-LAN HTTP control bind")
	timeout := flag.Duration("timeout", 120*time.Second, "hard deadline for the oracle")
	flag.Parse()
	if *relay == "" {
		log.Fatal("-relay required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()
	s := &state{codeReady: make(chan struct{}), matched: make(chan struct{}), ready: make(chan struct{}), acked: make(chan struct{}), ackReported: make(chan struct{}), result: make(chan error, 1)}
	tokens := make(chan string, 1)
	go func() { s.result <- run(ctx, *relay, s, tokens) }()
	mux := http.NewServeMux()
	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "GET required", 405)
			return
		}
		select {
		case t := <-tokens:
			fmt.Fprint(w, t)
		case <-ctx.Done():
			http.Error(w, "fixture timed out", 503)
		}
	})
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "GET required", http.StatusMethodNotAllowed)
			return
		}
		select {
		case <-s.ready:
			w.WriteHeader(http.StatusNoContent)
		case <-ctx.Done():
			http.Error(w, "backend never registered", http.StatusServiceUnavailable)
		}
	})
	mux.HandleFunc("/ack", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "GET required", http.StatusMethodNotAllowed)
			return
		}
		select {
		case <-s.acked:
			w.WriteHeader(http.StatusNoContent)
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
			close(s.ackReported)
		case <-ctx.Done():
			http.Error(w, "encrypted chat ACK not received", http.StatusServiceUnavailable)
		}
	})
	mux.HandleFunc("/sas", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST required", 405)
			return
		}
		body, err := io.ReadAll(io.LimitReader(r.Body, 64))
		if err != nil {
			http.Error(w, err.Error(), 400)
			return
		}
		if err := s.checkCode(string(body)); err != nil {
			http.Error(w, err.Error(), 409)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
	ln, err := net.Listen("tcp4", *control)
	if err != nil {
		log.Fatal(err)
	}
	srv := &http.Server{Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	go srv.Serve(ln)
	defer srv.Close()
	fmt.Printf("CONTROL=%s\n", ln.Addr().String())
	select {
	case err := <-s.result:
		if err != nil {
			log.Fatal(err)
		}
		select {
		case <-s.ackReported:
		case <-ctx.Done():
			log.Fatal("device did not confirm fixture ACK")
		}
		fmt.Println("DEVICE_ORACLE_PASS")
	case <-ctx.Done():
		log.Fatal(ctx.Err())
	}
}

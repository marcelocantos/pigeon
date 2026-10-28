// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestSASMustMatchBeforeConfirmation(t *testing.T) {
	s := &state{codeReady: make(chan struct{}), matched: make(chan struct{})}
	s.code = "123456"
	close(s.codeReady)
	if err := s.checkCode("654321"); err == nil || !strings.Contains(err.Error(), "mismatch") {
		t.Fatalf("mismatch: %v", err)
	}
	select {
	case <-s.matched:
		t.Fatal("wrong SAS confirmed")
	default:
	}
	if err := s.checkCode("123456"); err != nil {
		t.Fatal(err)
	}
	select {
	case <-s.matched:
	default:
		t.Fatal("matching SAS not confirmed")
	}
}

type fakeChat struct {
	incoming chan []byte
	sent     chan []byte
}

func (f *fakeChat) Send(b []byte) error { f.sent <- append([]byte(nil), b...); return nil }
func (f *fakeChat) Recv(ctx context.Context) ([]byte, error) {
	select {
	case b := <-f.incoming:
		return b, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func TestEncryptedChatWaitsForDeviceAckBeforePassAndClose(t *testing.T) {
	f := &fakeChat{incoming: make(chan []byte, 2), sent: make(chan []byte, 1)}
	f.incoming <- []byte("ping")
	passed, closed := make(chan struct{}), make(chan struct{})
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	result := make(chan error, 1)
	go func() {
		defer close(closed)
		err := encryptedChat(ctx, f)
		if err == nil {
			close(passed)
		}
		result <- err
	}()
	select {
	case echo := <-f.sent:
		if string(echo) != "echo: ping" {
			t.Fatalf("echo = %q", echo)
		}
	case <-ctx.Done():
		t.Fatal("echo not sent")
	}
	select {
	case <-passed:
		t.Fatal("premature PASS")
	case <-closed:
		t.Fatal("premature close")
	default:
	}
	f.incoming <- []byte("ack: echo: ping")
	if err := <-result; err != nil {
		t.Fatal(err)
	}
	<-passed
	<-closed
}

func TestEncryptedChatAckTimeoutFailsNotSkips(t *testing.T) {
	f := &fakeChat{incoming: make(chan []byte, 1), sent: make(chan []byte, 1)}
	f.incoming <- []byte("ping")
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
	defer cancel()
	if err := encryptedChat(ctx, f); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("missing ACK must fail on deadline: %v", err)
	}
}

func TestEncryptedChatRejectsWrongAck(t *testing.T) {
	f := &fakeChat{incoming: make(chan []byte, 2), sent: make(chan []byte, 1)}
	f.incoming <- []byte("ping")
	f.incoming <- []byte("ack: wrong")
	if err := encryptedChat(context.Background(), f); err == nil || !strings.Contains(err.Error(), "chat ack") {
		t.Fatalf("wrong ACK accepted: %v", err)
	}
}

// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"strings"
	"testing"
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

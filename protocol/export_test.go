// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package protocol

import (
	"testing"
)

// Unit tests for internal helper functions.

func TestGoConstPrefix(t *testing.T) {
	tests := []struct {
		actor string
		want  string
	}{
		{"ios", "App"},
		{"cli", "CLI"},
		{"server", "Server"},
		{"client", "Client"},
		{"", ""},
		{"x", "X"},
	}
	for _, tc := range tests {
		got := goConstPrefix(tc.actor)
		if got != tc.want {
			t.Errorf("goConstPrefix(%q) = %q, want %q", tc.actor, got, tc.want)
		}
	}
}

func TestSwiftTypeName(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"server", "Server"},
		{"ios", "Ios"},
		{"", ""},
	}
	for _, tc := range tests {
		got := swiftTypeName(tc.in)
		if got != tc.want {
			t.Errorf("swiftTypeName(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestSwiftCase(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"Idle", "idle"},
		{"ServerIdle", "serverIdle"},
		{"", ""},
		{"PairBegin", "pairBegin"},
	}
	for _, tc := range tests {
		got := swiftCase(tc.in)
		if got != tc.want {
			t.Errorf("swiftCase(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// PairingCeremony-based export tests removed along with the old
// pairing artifacts. Codegen export is still exercised through the
// session-protocol pipeline in protocol_test.go.

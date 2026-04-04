// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package protocol

import (
	"bytes"
	"strings"
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

func TestExportGoStructure(t *testing.T) {
	p := PairingCeremony()
	var buf bytes.Buffer
	if err := p.ExportGo(&buf, "protocol", "PairingCeremony"); err != nil {
		t.Fatalf("ExportGo: %v", err)
	}

	out := buf.String()

	checks := []string{
		"package protocol",
		"ServerIdle",
		"MsgPairBegin",
		"GuardTokenValid",
		"ActionGenerateToken",
		"func PairingCeremony",
		"ChannelBound",
		"OneShot",
	}

	for _, want := range checks {
		if !strings.Contains(out, want) {
			t.Errorf("ExportGo output missing %q", want)
		}
	}
}

func TestExportSwiftStructure(t *testing.T) {
	p := PairingCeremony()
	var buf bytes.Buffer
	if err := p.ExportSwift(&buf); err != nil {
		t.Fatalf("ExportSwift: %v", err)
	}

	out := buf.String()

	checks := []string{
		"MessageType",
		"ServerState",
		"IosState",
		"CliState",
		"Transitions",
		"public",
	}

	for _, want := range checks {
		if !strings.Contains(out, want) {
			t.Errorf("ExportSwift output missing %q", want)
		}
	}
}

func TestExportPlantUMLStructure(t *testing.T) {
	p := PairingCeremony()
	var buf bytes.Buffer
	if err := p.ExportPlantUML(&buf); err != nil {
		t.Fatalf("ExportPlantUML: %v", err)
	}

	out := buf.String()

	checks := []string{
		"@startuml",
		"@enduml",
		"PairingCeremony",
		"server",
		"ios",
		"cli",
	}

	for _, want := range checks {
		if !strings.Contains(out, want) {
			t.Errorf("ExportPlantUML output missing %q", want)
		}
	}
}

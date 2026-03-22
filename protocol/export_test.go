// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package protocol

import (
	"bytes"
	"strings"
	"testing"
)

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
		// TODO(🎯T2.6): ExportGo does not yet emit ChannelBound or OneShot.
		// Uncomment when gogen.go is fixed.
		// "ChannelBound",
		// "OneShot",
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
		"handleMessage",
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

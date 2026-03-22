// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package protocol

import (
	"testing"
)

func TestLoadYAML(t *testing.T) {
	p, err := LoadYAML("pairing.yaml")
	if err != nil {
		t.Fatalf("LoadYAML: %v", err)
	}
	if p.Name != "PairingCeremony" {
		t.Fatalf("expected Name %q, got %q", "PairingCeremony", p.Name)
	}
	if err := p.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
}

func TestParseYAMLInvalid(t *testing.T) {
	_, err := ParseYAML([]byte("{\x00garbage: [[["))
	if err == nil {
		t.Fatal("expected error parsing garbage YAML, got nil")
	}
}

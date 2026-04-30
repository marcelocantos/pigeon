// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package protocol

import (
	"testing"
)

func TestLoadYAML(t *testing.T) {
	p, err := LoadYAML("session.yaml")
	if err != nil {
		t.Fatalf("LoadYAML: %v", err)
	}
	if p.Name == "" {
		t.Fatalf("expected non-empty Name")
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

func TestLoadYAMLFileNotFound(t *testing.T) {
	if _, err := LoadYAML("/nonexistent/path/to/file.yaml"); err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestParseYAMLMessagesNonMapping(t *testing.T) {
	// messages: must be a mapping node; a sequence should cause a parse error.
	yaml := `
name: Bad
messages:
  - foo
`
	if _, err := ParseYAML([]byte(yaml)); err == nil {
		t.Fatal("expected error when messages is a sequence, not a mapping")
	}
}

func TestParseYAMLActorsNonMapping(t *testing.T) {
	yaml := `
name: Bad
actors:
  - foo
`
	if _, err := ParseYAML([]byte(yaml)); err == nil {
		t.Fatal("expected error when actors is a sequence, not a mapping")
	}
}

func TestParseYAMLLivenessProperty(t *testing.T) {
	yaml := `
name: Simple
properties:
  - name: EventuallyDone
    kind: liveness
    expr: <>(done)
    desc: Eventually reaches done
`
	p, err := ParseYAML([]byte(yaml))
	if err != nil {
		t.Fatalf("ParseYAML: %v", err)
	}
	if len(p.Properties) != 1 {
		t.Fatalf("expected 1 property, got %d", len(p.Properties))
	}
	if p.Properties[0].Kind != Liveness {
		t.Fatalf("expected Liveness kind, got %v", p.Properties[0].Kind)
	}
}

func TestParseYAMLGuardsNonMapping(t *testing.T) {
	yaml := `
name: Bad
guards:
  - foo
`
	if _, err := ParseYAML([]byte(yaml)); err == nil {
		t.Fatal("expected error when guards is a sequence")
	}
}

// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestSwiftOutputPaths exercises the CLI rather than only the Swift exporters:
// the compiled module must receive exactly the generated files in PigeonCore.
func TestSwiftOutputPaths(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	out := t.TempDir()
	binary := filepath.Join(out, "protogen")
	build := exec.Command("go", "build", "-o", binary, ".")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build protogen: %v\n%s", err, output)
	}

	for _, args := range [][]string{
		{filepath.Join(root, "protocol", "pairing.yaml")},
		{"--wireformats", filepath.Join(root, "protocol", "wireformats.yaml")},
	} {
		cmd := exec.Command(binary, args...)
		cmd.Dir = out
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("protogen %v: %v\n%s", args, err, output)
		}
	}

	for _, file := range []string{"PairingCeremonyMachine.swift", "WireGen.swift"} {
		canonical := filepath.Join("Sources", "PigeonCore", file)
		generated, err := os.ReadFile(filepath.Join(out, canonical))
		if err != nil {
			t.Fatalf("generated %s: %v", canonical, err)
		}
		tracked, err := os.ReadFile(filepath.Join(root, canonical))
		if err != nil {
			t.Fatalf("tracked %s: %v", canonical, err)
		}
		if !bytes.Equal(generated, tracked) {
			t.Errorf("generated %s differs from tracked artifact", canonical)
		}
		stray := filepath.Join(out, "Sources", "Pigeon", file)
		if _, err := os.Stat(stray); !os.IsNotExist(err) {
			t.Errorf("stray generator output %s: stat error %v", stray, err)
		}
	}
}

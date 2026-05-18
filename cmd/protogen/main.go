// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// protogen generates Go, Swift, Kotlin, TypeScript, C, TLA+, and PlantUML from a
// YAML protocol definition.
//
// Usage:
//
//	protogen [--root-pkg=<pkg>] protocol/session.yaml
//	protogen --wireformats protocol/wireformats.yaml
//
// Outputs (relative to working directory):
//
//	protocol/<name>_gen.go  (skipped when --root-pkg is set)
//	formal/<Name>.tla
//	docs/<name>.puml
//	Sources/Pigeon/<Name>Machine.swift
//	android/pigeon/src/main/kotlin/com.marcelocantos.pigeon/crypto/<Name>Machine.kt
//	web/src/<Name>Machine.ts
//	<name>_gen.go           (only when --root-pkg is set)
//
// --wireformats mode (one-shot byte formats; no FSM, no TLA+):
//
//	wire_gen.go             (root package, default "pigeon")
//	c/include/pigeon/wire_gen.h
//	c/src/wire_gen.c
//	Sources/Pigeon/WireGen.swift
//	android/pigeon/src/main/kotlin/com/marcelocantos/pigeon/crypto/WireGen.kt
//	web/src/wireGen.ts
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/marcelocantos/pigeon/protocol"
)

func main() {
	var rootPkg, rootOut string
	var wireformats bool
	args := os.Args[1:]
	for len(args) > 0 && strings.HasPrefix(args[0], "--") {
		switch {
		case strings.HasPrefix(args[0], "--root-pkg="):
			rootPkg = strings.TrimPrefix(args[0], "--root-pkg=")
		case strings.HasPrefix(args[0], "--root-out="):
			rootOut = strings.TrimPrefix(args[0], "--root-out=")
		case args[0] == "--wireformats":
			wireformats = true
		default:
			fmt.Fprintf(os.Stderr, "unknown flag: %s\n", args[0])
			os.Exit(1)
		}
		args = args[1:]
	}

	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: protogen [--root-pkg=<pkg>] <protocol.yaml>")
		fmt.Fprintln(os.Stderr, "       protogen --wireformats <wireformats.yaml>")
		os.Exit(1)
	}

	if wireformats {
		if err := runWireformats(args[0], rootPkg); err != nil {
			fmt.Fprintf(os.Stderr, "wireformats: %v\n", err)
			os.Exit(1)
		}
		return
	}

	p, err := protocol.LoadYAML(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "load: %v\n", err)
		os.Exit(1)
	}

	if err := p.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "validate: %v\n", err)
		os.Exit(1)
	}

	lowerName := strings.ToLower(p.Name)

	generators := []struct {
		path string
		gen  func() error
	}{
		{
			// Skip protocol/ output when --root-pkg is set (the root-level
			// file replaces it, avoiding redeclaration conflicts when a
			// unified protocol subsumes an older one).
			path: filepath.Join("protocol", lowerName+"_gen.go"),
			gen: func() func() error {
				if rootPkg != "" {
					return nil
				}
				return func() error {
					return writeFile(
						filepath.Join("protocol", lowerName+"_gen.go"),
						func(f *os.File) error {
							return p.ExportGo(f, "protocol", p.Name)
						},
					)
				}
			}(),
		},
		{
			// The composed (cross-phase) TLA+ spec is suppressed for the
			// Session protocol: a composed pairing+session spec exploded
			// TLC's state space and was abandoned in favour of two
			// independent specs joined by a ValidPairingRecord interface
			// contract — see docs/session-protocol.md and 🎯T39.2.
			path: filepath.Join("formal", p.Name+".tla"),
			gen: func() func() error {
				if p.Name == "Session" {
					return nil
				}
				return func() error {
					return writeFile(
						filepath.Join("formal", p.Name+".tla"),
						func(f *os.File) error { return p.ExportTLA(f) },
					)
				}
			}(),
		},
		{
			path: filepath.Join("docs", "transport.puml"),
			gen: func() error {
				return writeFile(
					filepath.Join("docs", "transport.puml"),
					func(f *os.File) error {
						return p.ExportPlantUMLActors(f, "Transport", []string{"backend", "client"})
					},
				)
			},
		},
		{
			path: filepath.Join("docs", "relay.puml"),
			gen: func() error {
				return writeFile(
					filepath.Join("docs", "relay.puml"),
					func(f *os.File) error {
						return p.ExportPlantUMLActors(f, "Relay", []string{"relay"})
					},
				)
			},
		},
		{
			path: filepath.Join("Sources", "Pigeon", p.Name+"Machine.swift"),
			gen: func() error {
				return writeFile(
					filepath.Join("Sources", "Pigeon", p.Name+"Machine.swift"),
					func(f *os.File) error { return p.ExportSwift(f) },
				)
			},
		},
		{
			path: filepath.Join("android", "pigeon", "src", "main", "kotlin",
				"com", "marcelocantos", "pigeon", "crypto", p.Name+"Machine.kt"),
			gen: func() error {
				return writeFile(
					filepath.Join("android", "pigeon", "src", "main", "kotlin",
						"com", "marcelocantos", "pigeon", "crypto", p.Name+"Machine.kt"),
					func(f *os.File) error {
						return p.ExportKotlin(f, "com.marcelocantos.pigeon.crypto")
					},
				)
			},
		},
		{
			path: filepath.Join("web", "src", p.Name+"Machine.ts"),
			gen: func() error {
				return writeFile(
					filepath.Join("web", "src", p.Name+"Machine.ts"),
					func(f *os.File) error { return p.ExportTypeScript(f) },
				)
			},
		},
		{
			path: filepath.Join("c", "include", "pigeon", lowerName+"_gen.h"),
			gen: func() error {
				return writeFile(
					filepath.Join("c", "include", "pigeon", lowerName+"_gen.h"),
					func(f *os.File) error { return p.ExportCHeader(f) },
				)
			},
		},
		{
			path: filepath.Join("c", "src", lowerName+"_gen.c"),
			gen: func() error {
				return writeFile(
					filepath.Join("c", "src", lowerName+"_gen.c"),
					func(f *os.File) error { return p.ExportCImpl(f) },
				)
			},
		},
	}

	// Optional root-level Go file for a different package.
	// --root-pkg=NAME emits a typed runtime machine in package NAME.
	// --root-out=DIR (optional) puts the file under DIR (default: ".").
	if rootPkg != "" {
		dir := rootOut
		if dir == "" {
			dir = "."
		}
		outPath := filepath.Join(dir, lowerName+"_gen.go")
		funcName := p.Name + "Protocol"
		generators = append(generators, struct {
			path string
			gen  func() error
		}{
			path: outPath,
			gen: func() error {
				return writeFile(
					outPath,
					func(f *os.File) error {
						return p.ExportGo(f, rootPkg, funcName)
					},
				)
			},
		})
	}

	for _, g := range generators {
		if g.gen == nil {
			continue
		}
		if err := g.gen(); err != nil {
			fmt.Fprintf(os.Stderr, "generate %s: %v\n", g.path, err)
			os.Exit(1)
		}
		fmt.Printf("wrote %s\n", g.path)
	}

	// Phase-specific TLA+ specs. Two pieces of phase-export policy
	// for the Session protocol live here (🎯T39.2):
	//   * The Pairing phase is *not* exported. PairingCeremony.tla
	//     (generated separately from protocol/pairing.yaml) is the
	//     authoritative pairing spec; emitting Session_Pairing.tla
	//     alongside it would just duplicate that work.
	//   * The Transport phase exports as SessionMachine.tla, the
	//     canonical name for the post-pairing transport state
	//     machine. Other protocols keep the default <Name>_<Phase>.tla
	//     scheme.
	for _, ph := range p.Phases {
		if p.Name == "Session" && ph.Name == "Pairing" {
			continue
		}
		name := p.Name + "_" + strings.ReplaceAll(ph.Name, " ", "_")
		if p.Name == "Session" && ph.Name == "Transport" {
			name = "SessionMachine"
		}
		path := filepath.Join("formal", name+".tla")
		if err := writeFile(path, func(f *os.File) error {
			return p.ExportTLAPhase(f, ph.Name)
		}); err != nil {
			fmt.Fprintf(os.Stderr, "generate %s: %v\n", path, err)
			os.Exit(1)
		}
		fmt.Printf("wrote %s\n", path)
	}
}

// runWireformats drives the byte-format generator across all 5 SDK targets.
// Unlike the FSM mode, there is no TLA+ or PlantUML output — these formats
// don't carry state-machine semantics. The per-language file paths are
// fixed (one file per language) since the set lives in a single YAML.
func runWireformats(specPath, rootPkg string) error {
	set, err := protocol.LoadWireFormatsYAML(specPath)
	if err != nil {
		return fmt.Errorf("load: %w", err)
	}
	if err := set.Validate(); err != nil {
		return fmt.Errorf("validate: %w", err)
	}
	pkg := rootPkg
	if pkg == "" {
		pkg = "pigeon"
	}
	targets := []struct {
		path string
		gen  func(f *os.File) error
	}{
		{"wire_gen.go", func(f *os.File) error { return set.ExportGo(f, pkg) }},
		{filepath.Join("c", "include", "pigeon", "wire_gen.h"), func(f *os.File) error { return set.ExportCHeader(f) }},
		{filepath.Join("c", "src", "wire_gen.c"), func(f *os.File) error { return set.ExportCImpl(f) }},
		{filepath.Join("Sources", "Pigeon", "WireGen.swift"), func(f *os.File) error { return set.ExportSwift(f) }},
		{filepath.Join("android", "pigeon", "src", "main", "kotlin",
			"com", "marcelocantos", "pigeon", "crypto", "WireGen.kt"),
			func(f *os.File) error { return set.ExportKotlin(f, "com.marcelocantos.pigeon.crypto") }},
		{filepath.Join("web", "src", "wireGen.ts"), func(f *os.File) error { return set.ExportTypeScript(f) }},
	}
	for _, t := range targets {
		if err := writeFile(t.path, t.gen); err != nil {
			return fmt.Errorf("generate %s: %w", t.path, err)
		}
		fmt.Printf("wrote %s\n", t.path)
	}
	return nil
}

func writeFile(path string, fn func(*os.File) error) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	if err := fn(f); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	// Run gofmt on .go outputs so freshly-regenerated files don't trip
	// the bullseye gofmt check.
	if strings.HasSuffix(path, ".go") {
		if out, err := exec.Command("gofmt", "-w", path).CombinedOutput(); err != nil {
			return fmt.Errorf("gofmt %s: %w (%s)", path, err, out)
		}
	}
	return nil
}

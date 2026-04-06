// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// protogen generates Go, Swift, Kotlin, TypeScript, TLA+, and PlantUML from a
// YAML protocol definition.
//
// Usage:
//
//	protogen [--root-pkg=<pkg>] protocol/session.yaml
//
// Outputs (relative to working directory):
//
//	protocol/<name>_gen.go  (skipped when --root-pkg is set)
//	formal/<Name>.tla
//	docs/<name>.puml
//	Sources/Tern/<Name>Machine.swift
//	android/tern/src/main/kotlin/com/marcelocantos/tern/crypto/<Name>Machine.kt
//	web/src/<Name>Machine.ts
//	<name>_gen.go           (only when --root-pkg is set)
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/marcelocantos/tern/protocol"
)

func main() {
	var rootPkg string
	args := os.Args[1:]
	for len(args) > 0 && strings.HasPrefix(args[0], "--") {
		if rest, ok := strings.CutPrefix(args[0], "--root-pkg="); ok {
			rootPkg = rest
		} else {
			fmt.Fprintf(os.Stderr, "unknown flag: %s\n", args[0])
			os.Exit(1)
		}
		args = args[1:]
	}

	if len(args) < 1 {
		fmt.Fprintln(os.Stderr, "usage: protogen [--root-pkg=<pkg>] <protocol.yaml>")
		os.Exit(1)
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
			path: filepath.Join("formal", p.Name+".tla"),
			gen: func() error {
				return writeFile(
					filepath.Join("formal", p.Name+".tla"),
					func(f *os.File) error { return p.ExportTLA(f) },
				)
			},
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
			path: filepath.Join("Sources", "Tern", p.Name+"Machine.swift"),
			gen: func() error {
				return writeFile(
					filepath.Join("Sources", "Tern", p.Name+"Machine.swift"),
					func(f *os.File) error { return p.ExportSwift(f) },
				)
			},
		},
		{
			path: filepath.Join("android", "tern", "src", "main", "kotlin",
				"com", "marcelocantos", "tern", "crypto", p.Name+"Machine.kt"),
			gen: func() error {
				return writeFile(
					filepath.Join("android", "tern", "src", "main", "kotlin",
						"com", "marcelocantos", "tern", "crypto", p.Name+"Machine.kt"),
					func(f *os.File) error {
						return p.ExportKotlin(f, "com.marcelocantos.tern.crypto")
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
	}

	// Optional root-level Go file for a different package.
	if rootPkg != "" {
		funcName := p.Name + "Protocol"
		generators = append(generators, struct {
			path string
			gen  func() error
		}{
			path: lowerName + "_gen.go",
			gen: func() error {
				return writeFile(
					lowerName+"_gen.go",
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

	// Phase-specific TLA+ specs.
	for _, ph := range p.Phases {
		name := p.Name + "_" + strings.ReplaceAll(ph.Name, " ", "_")
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

func writeFile(path string, fn func(*os.File) error) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return fn(f)
}

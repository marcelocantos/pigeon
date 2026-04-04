// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package protocol

import (
	"fmt"
	"io"
	"strings"
)

// ExportTypeScript writes a TypeScript source file with:
//   - Const enums for states (per actor) and message types
//   - Guard and action ID constants
//   - Transition table as static data
//
// Does NOT generate executor logic. The generic Machine class in
// web/src/machine.ts interprets the table at runtime.
func (p *Protocol) ExportTypeScript(w io.Writer) error {
	var b strings.Builder

	b.WriteString("// Copyright 2026 Marcelo Cantos\n")
	b.WriteString("// SPDX-License-Identifier: Apache-2.0\n\n")
	b.WriteString("// Auto-generated from protocol definition. Do not edit.\n")
	b.WriteString("// Source of truth: protocol/*.yaml\n\n")

	// Message type enum.
	b.WriteString("export enum MessageType {\n")
	for _, m := range p.Messages {
		fmt.Fprintf(&b, "    %s = \"%s\",\n", kotlinPascalCase(string(m.Type)), m.Type)
	}
	b.WriteString("}\n\n")

	// Per-actor state enum.
	for _, a := range p.Actors {
		typeName := tsTypeName(a.Name)
		states := collectStates(a)

		fmt.Fprintf(&b, "export enum %sState {\n", typeName)
		for _, s := range states {
			fmt.Fprintf(&b, "    %s = \"%s\",\n", string(s), s)
		}
		b.WriteString("}\n\n")
	}

	// Guard ID enum.
	if len(p.Guards) > 0 {
		b.WriteString("export enum GuardID {\n")
		for _, g := range p.Guards {
			fmt.Fprintf(&b, "    %s = \"%s\",\n", kotlinPascalCase(string(g.ID)), g.ID)
		}
		b.WriteString("}\n\n")
	}

	// Action ID enum.
	actions := collectActions(p)
	if len(actions) > 0 {
		b.WriteString("export enum ActionID {\n")
		for _, id := range actions {
			fmt.Fprintf(&b, "    %s = \"%s\",\n", kotlinPascalCase(id), id)
		}
		b.WriteString("}\n\n")
	}

	// Transition table interface.
	b.WriteString("export interface Transition {\n")
	b.WriteString("    readonly from: string;\n")
	b.WriteString("    readonly to: string;\n")
	b.WriteString("    readonly on: string;\n")
	b.WriteString("    readonly onKind: \"recv\" | \"internal\";\n")
	b.WriteString("    readonly guard?: string;\n")
	b.WriteString("    readonly action?: string;\n")
	b.WriteString("    readonly sends?: ReadonlyArray<{ readonly to: string; readonly msg: string }>;\n")
	b.WriteString("}\n\n")

	b.WriteString("export interface ActorTable {\n")
	b.WriteString("    readonly initial: string;\n")
	b.WriteString("    readonly transitions: ReadonlyArray<Transition>;\n")
	b.WriteString("}\n\n")

	// Per-actor table.
	for _, a := range p.Actors {
		typeName := tsTypeName(a.Name)
		fmt.Fprintf(&b, "/** %s transition table. */\n", a.Name)
		fmt.Fprintf(&b, "export const %sTable: ActorTable = {\n", strings.ToLower(typeName[:1])+typeName[1:])
		fmt.Fprintf(&b, "    initial: %sState.%s,\n", typeName, a.Initial)
		b.WriteString("    transitions: [\n")

		for _, t := range a.Transitions {
			onKind := "internal"
			onValue := t.On.Desc
			if t.On.Kind == TriggerRecv {
				onKind = "recv"
				onValue = string(t.On.Msg)
			}

			b.WriteString("        { ")
			fmt.Fprintf(&b, "from: %q, to: %q, on: %q, onKind: %q", t.From, t.To, onValue, onKind)
			if t.Guard != "" {
				fmt.Fprintf(&b, ", guard: %q", string(t.Guard))
			}
			if t.Do != "" {
				fmt.Fprintf(&b, ", action: %q", string(t.Do))
			}
			if len(t.Sends) > 0 {
				b.WriteString(", sends: [")
				for i, s := range t.Sends {
					if i > 0 {
						b.WriteString(", ")
					}
					fmt.Fprintf(&b, "{ to: %q, msg: %q }", s.To, s.Msg)
				}
				b.WriteString("]")
			}
			b.WriteString(" },\n")
		}

		b.WriteString("    ],\n")
		b.WriteString("};\n\n")
	}

	result := strings.TrimRight(b.String(), "\n") + "\n"
	_, err := io.WriteString(w, result)
	return err
}

func tsTypeName(name string) string {
	if len(name) == 0 {
		return name
	}
	return strings.ToUpper(name[:1]) + name[1:]
}

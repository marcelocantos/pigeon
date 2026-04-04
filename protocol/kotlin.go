// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package protocol

import (
	"fmt"
	"io"
	"strings"
	"unicode"
)

// ExportKotlin writes a Kotlin source file with:
//   - Enum classes for states (per actor) and message types
//   - Guard and action ID constants
//   - Transition table as static data
//
// Does NOT generate executor logic. The generic Machine class in
// the Kotlin library interprets the table at runtime.
func (p *Protocol) ExportKotlin(w io.Writer, pkg string) error {
	var b strings.Builder

	b.WriteString("// Copyright 2026 Marcelo Cantos\n")
	b.WriteString("// SPDX-License-Identifier: Apache-2.0\n\n")
	b.WriteString("// Auto-generated from protocol definition. Do not edit.\n")
	b.WriteString("// Source of truth: protocol/*.yaml\n\n")
	fmt.Fprintf(&b, "package %s\n\n", pkg)

	// Message type enum.
	b.WriteString("enum class MessageType(val value: String) {\n")
	for i, m := range p.Messages {
		comma := ","
		if i == len(p.Messages)-1 {
			comma = ";"
		}
		fmt.Fprintf(&b, "    %s(\"%s\")%s\n", kotlinPascalCase(string(m.Type)), m.Type, comma)
	}
	b.WriteString("}\n\n")

	// Per-actor state enum.
	for _, a := range p.Actors {
		typeName := kotlinTypeName(a.Name)
		states := collectStates(a)

		fmt.Fprintf(&b, "enum class %sState(val value: String) {\n", typeName)
		for i, s := range states {
			comma := ","
			if i == len(states)-1 {
				comma = ";"
			}
			fmt.Fprintf(&b, "    %s(\"%s\")%s\n", string(s), s, comma)
		}
		b.WriteString("}\n\n")
	}

	// Guard ID enum.
	if len(p.Guards) > 0 {
		b.WriteString("enum class GuardID(val value: String) {\n")
		for i, g := range p.Guards {
			comma := ","
			if i == len(p.Guards)-1 {
				comma = ";"
			}
			fmt.Fprintf(&b, "    %s(\"%s\")%s\n", kotlinPascalCase(string(g.ID)), g.ID, comma)
		}
		b.WriteString("}\n\n")
	}

	// Action ID enum.
	actions := collectActions(p)
	if len(actions) > 0 {
		b.WriteString("enum class ActionID(val value: String) {\n")
		for i, id := range actions {
			comma := ","
			if i == len(actions)-1 {
				comma = ";"
			}
			fmt.Fprintf(&b, "    %s(\"%s\")%s\n", kotlinPascalCase(id), id, comma)
		}
		b.WriteString("}\n\n")
	}

	// Transition table per actor.
	for _, a := range p.Actors {
		typeName := kotlinTypeName(a.Name)
		fmt.Fprintf(&b, "/** %s transition table. */\n", a.Name)
		fmt.Fprintf(&b, "object %sTable {\n", typeName)
		fmt.Fprintf(&b, "    val initial = %sState.%s\n\n", typeName, a.Initial)

		fmt.Fprintf(&b, "    data class Transition(\n")
		b.WriteString("        val from: String,\n")
		b.WriteString("        val to: String,\n")
		b.WriteString("        val on: String,\n")
		b.WriteString("        val onKind: String,\n")
		b.WriteString("        val guard: String? = null,\n")
		b.WriteString("        val action: String? = null,\n")
		b.WriteString("        val sends: List<Pair<String, String>> = emptyList(),\n")
		b.WriteString("    )\n\n")

		b.WriteString("    val transitions = listOf(\n")
		for _, t := range a.Transitions {
			onKind := "internal"
			onValue := t.On.Desc
			if t.On.Kind == TriggerRecv {
				onKind = "recv"
				onValue = string(t.On.Msg)
			}

			guardStr := "null"
			if t.Guard != "" {
				guardStr = fmt.Sprintf("%q", string(t.Guard))
			}
			actionStr := "null"
			if t.Do != "" {
				actionStr = fmt.Sprintf("%q", string(t.Do))
			}

			sends := "emptyList()"
			if len(t.Sends) > 0 {
				var parts []string
				for _, s := range t.Sends {
					parts = append(parts, fmt.Sprintf("%q to %q", s.To, s.Msg))
				}
				sends = "listOf(" + strings.Join(parts, ", ") + ")"
			}

			fmt.Fprintf(&b, "        Transition(%q, %q, %q, %q, %s, %s, %s),\n",
				t.From, t.To, onValue, onKind, guardStr, actionStr, sends)
		}
		b.WriteString("    )\n")
		b.WriteString("}\n\n")
	}

	_, err := io.WriteString(w, b.String())
	return err
}

func kotlinTypeName(name string) string {
	if len(name) == 0 {
		return name
	}
	return strings.ToUpper(name[:1]) + name[1:]
}

func kotlinPascalCase(s string) string {
	var result []rune
	nextUpper := true
	for _, r := range s {
		if r == '_' {
			nextUpper = true
			continue
		}
		if nextUpper {
			result = append(result, unicode.ToUpper(r))
			nextUpper = false
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}

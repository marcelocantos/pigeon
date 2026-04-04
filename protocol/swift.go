// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package protocol

import (
	"fmt"
	"io"
	"strings"
	"unicode"
)

// ExportSwift writes a Swift source file with:
//   - Enums for states (per actor) and message types
//   - Guard and action ID constants
//   - A static table literal (the Protocol struct as Swift data)
//
// It does NOT generate executor logic (handleMessage, step). The
// generic Machine executor in Sources/Tern/Machine.swift interprets
// the table at runtime.
func (p *Protocol) ExportSwift(w io.Writer) error {
	var b strings.Builder

	b.WriteString("// Copyright 2026 Marcelo Cantos\n")
	b.WriteString("// SPDX-License-Identifier: Apache-2.0\n\n")
	b.WriteString("// Auto-generated from protocol definition. Do not edit.\n")
	b.WriteString("// Source of truth: protocol/*.yaml\n\n")
	b.WriteString("import Foundation\n\n")

	// Message type enum.
	b.WriteString("public enum MessageType: String, Sendable {\n")
	for _, m := range p.Messages {
		fmt.Fprintf(&b, "    case %s = \"%s\"\n", swiftCase(string(m.Type)), m.Type)
	}
	b.WriteString("}\n\n")

	// Per-actor state enum.
	for _, a := range p.Actors {
		typeName := swiftTypeName(a.Name)
		states := collectStates(a)

		fmt.Fprintf(&b, "public enum %sState: String, Sendable {\n", typeName)
		for _, s := range states {
			fmt.Fprintf(&b, "    case %s = \"%s\"\n", swiftCase(string(s)), s)
		}
		b.WriteString("}\n\n")
	}

	// Guard ID enum.
	if len(p.Guards) > 0 {
		b.WriteString("public enum GuardID: String, Sendable {\n")
		for _, g := range p.Guards {
			fmt.Fprintf(&b, "    case %s = \"%s\"\n", swiftCase(string(g.ID)), g.ID)
		}
		b.WriteString("}\n\n")
	}

	// Action ID enum.
	actions := collectActions(p)
	if len(actions) > 0 {
		b.WriteString("public enum ActionID: String, Sendable {\n")
		for _, id := range actions {
			fmt.Fprintf(&b, "    case %s = \"%s\"\n", swiftCase(id), id)
		}
		b.WriteString("}\n\n")
	}

	// Protocol table as a static struct.
	b.WriteString("/// The protocol transition table. Fed to Machine for execution.\n")
	fmt.Fprintf(&b, "public enum %sProtocol {\n", swiftTypeName(p.Name))

	for _, a := range p.Actors {
		typeName := swiftTypeName(a.Name)
		fmt.Fprintf(&b, "\n    /// %s transitions.\n", a.Name)
		fmt.Fprintf(&b, "    public static let %sInitial: %sState = .%s\n\n",
			strings.ToLower(a.Name[:1])+a.Name[1:],
			typeName,
			swiftCase(string(a.Initial)))

		fmt.Fprintf(&b, "    public static let %sTransitions: [(from: String, to: String, on: String, onKind: String, guard: String?, action: String?, sends: [(to: String, msg: String)])] = [\n",
			strings.ToLower(a.Name[:1])+a.Name[1:])

		for _, t := range a.Transitions {
			onKind := "internal"
			onValue := t.On.Desc
			if t.On.Kind == TriggerRecv {
				onKind = "recv"
				onValue = string(t.On.Msg)
			}

			guardStr := "nil"
			if t.Guard != "" {
				guardStr = fmt.Sprintf("%q", string(t.Guard))
			}
			actionStr := "nil"
			if t.Do != "" {
				actionStr = fmt.Sprintf("%q", string(t.Do))
			}

			sends := "[]"
			if len(t.Sends) > 0 {
				var parts []string
				for _, s := range t.Sends {
					parts = append(parts, fmt.Sprintf("(to: %q, msg: %q)", s.To, s.Msg))
				}
				sends = "[" + strings.Join(parts, ", ") + "]"
			}

			fmt.Fprintf(&b, "        (from: %q, to: %q, on: %q, onKind: %q, guard: %s, action: %s, sends: %s),\n",
				t.From, t.To, onValue, onKind, guardStr, actionStr, sends)
		}
		b.WriteString("    ]\n")
	}

	b.WriteString("}\n")

	_, err := io.WriteString(w, b.String())
	return err
}

// collectActions returns a sorted list of unique action IDs from all actors.
func collectActions(p *Protocol) []string {
	seen := map[string]bool{}
	var result []string
	for _, a := range p.Actors {
		for _, t := range a.Transitions {
			if t.Do != "" && !seen[string(t.Do)] {
				seen[string(t.Do)] = true
				result = append(result, string(t.Do))
			}
		}
	}
	return result
}

func swiftTypeName(name string) string {
	if len(name) == 0 {
		return name
	}
	return strings.ToUpper(name[:1]) + name[1:]
}

func swiftCase(s string) string {
	if len(s) == 0 {
		return s
	}
	var result []rune
	prevUpper := false
	for i, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			prevUpper = false
			continue
		}
		if i == 0 || (result == nil) {
			result = append(result, unicode.ToLower(r))
			prevUpper = unicode.IsUpper(r)
			continue
		}
		if unicode.IsUpper(r) && !prevUpper {
			result = append(result, r)
		} else if prevUpper && unicode.IsLower(r) {
			if len(result) > 1 {
				last := result[len(result)-1]
				result[len(result)-1] = unicode.ToUpper(last)
			}
			result = append(result, r)
		} else {
			result = append(result, r)
		}
		prevUpper = unicode.IsUpper(r)
	}
	return string(result)
}

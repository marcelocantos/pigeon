// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package protocol

import (
	"fmt"
	"io"
	"strings"
)

// ExportTypeScript writes a TypeScript source file with enums for states and
// message types, and a table-driven state machine class for each actor.
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

	// Per-actor state enum + machine.
	for _, a := range p.Actors {
		typeName := tsTypeName(a.Name)
		states := collectStates(a)

		// State enum.
		fmt.Fprintf(&b, "export enum %sState {\n", typeName)
		for _, s := range states {
			fmt.Fprintf(&b, "    %s = \"%s\",\n", string(s), s)
		}
		b.WriteString("}\n\n")

		// Machine class.
		fmt.Fprintf(&b, "export class %sMachine {\n", typeName)
		fmt.Fprintf(&b, "    private _state: %sState = %sState.%s;\n\n", typeName, typeName, a.Initial)
		fmt.Fprintf(&b, "    get state(): %sState {\n", typeName)
		b.WriteString("        return this._state;\n")
		b.WriteString("    }\n\n")

		// handleMessage
		b.WriteString("    /** Process a received message. Returns the new state, or null if rejected. */\n")
		fmt.Fprintf(&b, "    handleMessage(msg: MessageType, guard?: (name: string) => boolean): %sState | null {\n", typeName)
		b.WriteString("        const check = guard ?? (() => true);\n")
		fmt.Fprintf(&b, "        let newState: %sState | null = null;\n", typeName)

		firstRecv := true
		for _, t := range a.Transitions {
			if t.On.Kind != TriggerRecv {
				continue
			}
			guardExpr := ""
			if t.Guard != "" {
				guardExpr = fmt.Sprintf(" && check(\"%s\")", t.Guard)
			}
			ifKeyword := "        } else if"
			if firstRecv {
				ifKeyword = "        if"
				firstRecv = false
			}
			fmt.Fprintf(&b, "%s (this._state === %sState.%s && msg === MessageType.%s%s) {\n",
				ifKeyword, typeName, t.From,
				kotlinPascalCase(string(t.On.Msg)),
				guardExpr)
			fmt.Fprintf(&b, "            newState = %sState.%s;\n", typeName, t.To)
		}
		if !firstRecv {
			b.WriteString("        }\n")
		}
		b.WriteString("        if (newState !== null) this._state = newState;\n")
		b.WriteString("        return newState;\n")
		b.WriteString("    }\n\n")

		// step
		b.WriteString("    /** Attempt an internal transition. Returns the new state, or null if none available. */\n")
		fmt.Fprintf(&b, "    step(guard?: (name: string) => boolean): %sState | null {\n", typeName)
		b.WriteString("        const check = guard ?? (() => true);\n")

		// Group internal transitions by from-state.
		byFrom := map[State][]Transition{}
		for _, t := range a.Transitions {
			if t.On.Kind == TriggerInternal {
				byFrom[t.From] = append(byFrom[t.From], t)
			}
		}

		firstInternal := true
		for _, s := range states {
			ts := byFrom[s]
			if len(ts) == 0 {
				continue
			}
			for _, t := range ts {
				ifKeyword := "        } else if"
				if firstInternal {
					ifKeyword = "        if"
					firstInternal = false
				}
				guardExpr := ""
				if t.Guard != "" {
					guardExpr = fmt.Sprintf(" && check(\"%s\")", t.Guard)
				}
				fmt.Fprintf(&b, "%s (this._state === %sState.%s%s) {\n",
					ifKeyword, typeName, s, guardExpr)
				fmt.Fprintf(&b, "            this._state = %sState.%s;\n", typeName, t.To)
				b.WriteString("            return this._state;\n")
			}
		}
		if !firstInternal {
			b.WriteString("        }\n")
		}

		b.WriteString("        return null;\n")
		b.WriteString("    }\n")

		b.WriteString("}\n\n")
	}

	// Trim trailing newline to keep file clean.
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

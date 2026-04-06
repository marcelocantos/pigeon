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
//   - EventID and CmdID enums
//   - Transition table as static data
//   - Per-actor typed machine classes with handleEvent methods
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

	// EventID enum — internal events + recv_* events for messages.
	events := tsCollectEvents(p)
	if len(events) > 0 {
		b.WriteString("export enum EventID {\n")
		for _, id := range events {
			fmt.Fprintf(&b, "    %s = \"%s\",\n", kotlinPascalCase(id), id)
		}
		b.WriteString("}\n\n")
	}

	// CmdID enum.
	if len(p.Commands) > 0 {
		b.WriteString("export enum CmdID {\n")
		for _, c := range p.Commands {
			fmt.Fprintf(&b, "    %s = \"%s\",\n", kotlinPascalCase(string(c.ID)), c.ID)
		}
		b.WriteString("}\n\n")
	}

	// Wire constants.
	if len(p.WireConsts) > 0 {
		b.WriteString("/** Protocol wire constants shared across all platforms. */\n")
		b.WriteString("export const Wire = {\n")
		for _, wc := range p.WireConsts {
			name := tsConstName(wc.Name)
			switch wc.Type {
			case "byte":
				fmt.Fprintf(&b, "    %s: 0x%02X,\n", name, wireInt(wc.Value))
			case "int":
				fmt.Fprintf(&b, "    %s: %d,\n", name, wireInt(wc.Value))
			case "duration_ms":
				fmt.Fprintf(&b, "    %s: %d, // ms\n", name, wireInt(wc.Value))
			case "string":
				fmt.Fprintf(&b, "    %s: %q,\n", name, wc.Value)
			}
		}
		b.WriteString("} as const;\n\n")
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

		for _, t := range a.FlattenedTransitions() {
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

	// Per-actor typed machine classes.
	for _, a := range p.Actors {
		typeName := tsTypeName(a.Name)

		// Collect vars updated by this actor's transitions.
		actorVarSet := map[string]bool{}
		for _, t := range a.FlattenedTransitions() {
			for _, u := range t.Updates {
				actorVarSet[u.Var] = true
			}
		}

		// Machine class.
		fmt.Fprintf(&b, "/** %sMachine is the generated state machine for the %s actor. */\n", typeName, a.Name)
		fmt.Fprintf(&b, "export class %sMachine {\n", typeName)
		fmt.Fprintf(&b, "    state: %sState;\n", typeName)

		// Typed variable fields owned by this actor.
		for _, v := range p.Vars {
			if !actorVarSet[v.Name] {
				continue
			}
			tsT := tsVarType(v.Type)
			init := tsInitialValue(v)
			comment := ""
			if v.Desc != "" {
				comment = " // " + v.Desc
			}
			fmt.Fprintf(&b, "    %s: %s = %s;%s\n", tsVarFieldName(v.Name), tsT, init, comment)
		}

		if len(p.Guards) > 0 {
			b.WriteString("    guards: Map<GuardID, () => boolean> = new Map();\n")
		}
		if len(actions) > 0 {
			b.WriteString("    actions: Map<ActionID, () => void> = new Map();\n")
		}
		b.WriteString("\n")

		// Constructor.
		fmt.Fprintf(&b, "    constructor() {\n")
		fmt.Fprintf(&b, "        this.state = %sState.%s;\n", typeName, a.Initial)
		b.WriteString("    }\n\n")

		// handleEvent method.
		b.WriteString("    handleEvent(ev: EventID): CmdID[] {\n")
		b.WriteString("        switch (true) {\n")

		for _, t := range a.FlattenedTransitions() {
			// Determine event constant name.
			var eventID string
			if t.On.Kind == TriggerRecv {
				eventID = "recv_" + string(t.On.Msg)
			} else {
				eventID = t.On.Desc
			}
			eventVal := kotlinPascalCase(eventID)

			// Build guard condition.
			guardCond := ""
			if t.Guard != "" {
				guardVal := kotlinPascalCase(string(t.Guard))
				guardCond = fmt.Sprintf(" && this.guards.get(GuardID.%s)?.() === true", guardVal)
			}

			fmt.Fprintf(&b, "            case this.state === %sState.%s && ev === EventID.%s%s: {\n",
				typeName, t.From, eventVal, guardCond)

			// Action call.
			if t.Do != "" {
				actionVal := kotlinPascalCase(string(t.Do))
				fmt.Fprintf(&b, "                this.actions.get(ActionID.%s)?.();\n", actionVal)
			}

			// Variable updates.
			for _, u := range t.Updates {
				varField := tsVarFieldName(u.Var)
				if lit, ok := tsSimpleLiteral(u.Expr); ok {
					fmt.Fprintf(&b, "                this.%s = %s;\n", varField, lit)
				} else {
					fmt.Fprintf(&b, "                // %s: %s (set by action)\n", u.Var, u.Expr)
				}
			}

			// State transition.
			fmt.Fprintf(&b, "                this.state = %sState.%s;\n", typeName, t.To)

			// Return commands.
			if len(t.Emits) > 0 {
				b.WriteString("                return [")
				for i, cmd := range t.Emits {
					if i > 0 {
						b.WriteString(", ")
					}
					fmt.Fprintf(&b, "CmdID.%s", kotlinPascalCase(string(cmd)))
				}
				b.WriteString("];\n")
			} else {
				b.WriteString("                return [];\n")
			}

			b.WriteString("            }\n")
		}

		b.WriteString("        }\n")
		b.WriteString("        return [];\n")
		b.WriteString("    }\n")
		b.WriteString("}\n\n")
	}

	result := strings.TrimRight(b.String(), "\n") + "\n"
	_, err := io.WriteString(w, result)
	return err
}

// tsCollectEvents returns a deduplicated, ordered list of all event IDs:
// declared events, internal transition events, and recv_* events.
func tsCollectEvents(p *Protocol) []string {
	seen := map[string]bool{}
	var result []string
	add := func(id string) {
		if !seen[id] {
			seen[id] = true
			result = append(result, id)
		}
	}
	// Declared events first.
	for _, e := range p.Events {
		add(string(e.ID))
	}
	// Internal events from transitions.
	for _, a := range p.Actors {
		for _, t := range a.FlattenedTransitions() {
			if t.On.Kind == TriggerInternal {
				add(t.On.Desc)
			}
		}
	}
	// recv_* events.
	for _, a := range p.Actors {
		for _, t := range a.FlattenedTransitions() {
			if t.On.Kind == TriggerRecv {
				add("recv_" + string(t.On.Msg))
			}
		}
	}
	return result
}

// tsVarType converts a VarType to a TypeScript type string.
func tsVarType(t VarType) string {
	switch t {
	case VarInt:
		return "number"
	case VarBool:
		return "boolean"
	case VarSetString:
		return "Set<string>"
	default:
		return "string"
	}
}

// tsInitialValue returns a TypeScript initial value expression for a VarDef.
func tsInitialValue(v VarDef) string {
	if lit, ok := tsSimpleLiteral(v.Initial); ok {
		return lit
	}
	switch v.Type {
	case VarInt:
		return "0"
	case VarBool:
		return "false"
	case VarSetString:
		return "new Set()"
	default:
		return `""`
	}
}

// tsSimpleLiteral converts a TLA+ expression to a TypeScript literal when
// it is simple enough (string, int, bool). Returns ("", false) otherwise.
func tsSimpleLiteral(expr string) (string, bool) {
	expr = strings.TrimSpace(expr)
	switch expr {
	case "TRUE":
		return "true", true
	case "FALSE":
		return "false", true
	}
	if strings.HasPrefix(expr, "\"") && strings.HasSuffix(expr, "\"") {
		return expr, true // already a string literal
	}
	var n int
	if _, err := fmt.Sscanf(expr, "%d", &n); err == nil {
		return fmt.Sprintf("%d", n), true
	}
	return "", false
}

// tsVarFieldName converts a snake_case var name to camelCase for TypeScript.
func tsConstName(s string) string {
	return strings.ToUpper(s)
}

func tsVarFieldName(name string) string {
	parts := strings.Split(name, "_")
	if len(parts) == 0 {
		return name
	}
	var b strings.Builder
	for i, p := range parts {
		if i == 0 {
			b.WriteString(strings.ToLower(p))
		} else if len(p) > 0 {
			b.WriteString(strings.ToUpper(p[:1]) + p[1:])
		}
	}
	return b.String()
}

func tsTypeName(name string) string {
	if len(name) == 0 {
		return name
	}
	return strings.ToUpper(name[:1]) + name[1:]
}

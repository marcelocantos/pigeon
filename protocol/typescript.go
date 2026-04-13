// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package protocol

import (
	"fmt"
	"io"
	"strings"
)

// ExportTypeScript writes a TypeScript source file with:
//   - Per-actor state enums at module scope (collision-free, actor-prefixed)
//   - A <ProtoName>Protocol namespace containing: MessageType, GuardID,
//     ActionID, EventID, CmdID, Wire constants, Transition/ActorTable
//     interfaces, and per-actor transition tables
//   - Per-actor typed machine classes with handleEvent methods
//   - For composed actors: per-sub-machine classes and a composite class
//     with a route() method
func (p *Protocol) ExportTypeScript(w io.Writer) error {
	var b strings.Builder

	b.WriteString("// Copyright 2026 Marcelo Cantos\n")
	b.WriteString("// SPDX-License-Identifier: Apache-2.0\n\n")
	b.WriteString("// Auto-generated from protocol definition. Do not edit.\n")
	b.WriteString("// Source of truth: protocol/*.yaml\n\n")

	protoName := tsTypeName(p.Name)
	nsName := protoName + "Protocol"

	// Per-actor state enums at module scope — prefixed with protocol name
	// to avoid collisions when multiple protocols share actor names.
	// For composed actors, emit one enum per sub-machine.
	for _, a := range p.Actors {
		if a.IsComposed() {
			for _, m := range a.Machines {
				typeName := tsTypeName(a.Name)
				machName := tsTypeName(m.Name)
				stateEnum := protoName + typeName + machName + "State"
				states := collectSubMachineStates(m)

				fmt.Fprintf(&b, "export enum %s {\n", stateEnum)
				for _, s := range states {
					fmt.Fprintf(&b, "    %s = \"%s\",\n", string(s), s)
				}
				b.WriteString("}\n\n")
			}
			continue
		}

		typeName := tsTypeName(a.Name)
		stateEnum := protoName + typeName + "State"
		states := collectStates(a)

		fmt.Fprintf(&b, "export enum %s {\n", stateEnum)
		for _, s := range states {
			fmt.Fprintf(&b, "    %s = \"%s\",\n", string(s), s)
		}
		b.WriteString("}\n\n")
	}

	// Protocol namespace — contains shared enums and transition tables.
	// Using a namespace prevents name collisions when multiple protocol files
	// are imported into the same module (e.g. PairingCeremonyMachine.ts and
	// SessionMachine.ts both define MessageType, GuardID, etc.).
	b.WriteString("/** The protocol transition table and shared type enums. */\n")
	fmt.Fprintf(&b, "export namespace %s {\n\n", nsName)

	// Message type enum (nested).
	b.WriteString("    export enum MessageType {\n")
	for _, m := range p.Messages {
		fmt.Fprintf(&b, "        %s = \"%s\",\n", kotlinPascalCase(string(m.Type)), m.Type)
	}
	b.WriteString("    }\n\n")

	// Guard ID enum (nested).
	if len(p.Guards) > 0 {
		b.WriteString("    export enum GuardID {\n")
		for _, g := range p.Guards {
			fmt.Fprintf(&b, "        %s = \"%s\",\n", kotlinPascalCase(string(g.ID)), g.ID)
		}
		b.WriteString("    }\n\n")
	}

	// Action ID enum (nested).
	actions := collectActions(p)
	if len(actions) > 0 {
		b.WriteString("    export enum ActionID {\n")
		for _, id := range actions {
			fmt.Fprintf(&b, "        %s = \"%s\",\n", kotlinPascalCase(id), id)
		}
		b.WriteString("    }\n\n")
	}

	// EventID enum (nested) — internal events + recv_* events for messages.
	events := tsCollectEvents(p)
	if len(events) > 0 {
		b.WriteString("    export enum EventID {\n")
		for _, id := range events {
			fmt.Fprintf(&b, "        %s = \"%s\",\n", kotlinPascalCase(id), id)
		}
		b.WriteString("    }\n\n")
	}

	// CmdID enum (nested).
	if len(p.Commands) > 0 {
		b.WriteString("    export enum CmdID {\n")
		for _, c := range p.Commands {
			fmt.Fprintf(&b, "        %s = \"%s\",\n", kotlinPascalCase(string(c.ID)), c.ID)
		}
		b.WriteString("    }\n\n")
	}

	// Wire constants (nested).
	if len(p.WireConsts) > 0 {
		b.WriteString("    /** Protocol wire constants shared across all platforms. */\n")
		b.WriteString("    export const Wire = {\n")
		for _, wc := range p.WireConsts {
			name := tsConstName(wc.Name)
			switch wc.Type {
			case "byte":
				fmt.Fprintf(&b, "        %s: 0x%02X,\n", name, wireInt(wc.Value))
			case "int":
				fmt.Fprintf(&b, "        %s: %d,\n", name, wireInt(wc.Value))
			case "duration_ms":
				fmt.Fprintf(&b, "        %s: %d, // ms\n", name, wireInt(wc.Value))
			case "string":
				fmt.Fprintf(&b, "        %s: %q,\n", name, wc.Value)
			}
		}
		b.WriteString("    } as const;\n\n")
	}

	// Transition table interfaces (nested).
	b.WriteString("    export interface Transition {\n")
	b.WriteString("        readonly from: string;\n")
	b.WriteString("        readonly to: string;\n")
	b.WriteString("        readonly on: string;\n")
	b.WriteString("        readonly onKind: \"recv\" | \"internal\";\n")
	b.WriteString("        readonly guard?: string;\n")
	b.WriteString("        readonly action?: string;\n")
	b.WriteString("        readonly sends?: ReadonlyArray<{ readonly to: string; readonly msg: string }>;\n")
	b.WriteString("    }\n\n")

	b.WriteString("    export interface ActorTable {\n")
	b.WriteString("        readonly initial: string;\n")
	b.WriteString("        readonly transitions: ReadonlyArray<Transition>;\n")
	b.WriteString("    }\n\n")

	// Per-actor transition tables (nested).
	// For composed actors, emit one table per sub-machine.
	for _, a := range p.Actors {
		if a.IsComposed() {
			for _, m := range a.Machines {
				typeName := tsTypeName(a.Name)
				machName := tsTypeName(m.Name)
				stateEnum := protoName + typeName + machName + "State"
				tableVar := strings.ToLower(typeName[:1]) + typeName[1:] + machName
				fmt.Fprintf(&b, "    /** %s/%s transition table. */\n", a.Name, m.Name)
				fmt.Fprintf(&b, "    export const %sTable: ActorTable = {\n", tableVar)
				fmt.Fprintf(&b, "        initial: %s.%s,\n", stateEnum, m.Initial)
				b.WriteString("        transitions: [\n")

				for _, t := range m.FlattenedTransitions() {
					onKind := "internal"
					onValue := t.On.Desc
					if t.On.Kind == TriggerRecv {
						onKind = "recv"
						onValue = string(t.On.Msg)
					}

					b.WriteString("            { ")
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

				b.WriteString("        ],\n")
				b.WriteString("    };\n\n")
			}
			continue
		}

		typeName := tsTypeName(a.Name)
		stateEnum := protoName + typeName + "State"
		fmt.Fprintf(&b, "    /** %s transition table. */\n", a.Name)
		fmt.Fprintf(&b, "    export const %sTable: ActorTable = {\n", strings.ToLower(typeName[:1])+typeName[1:])
		fmt.Fprintf(&b, "        initial: %s.%s,\n", stateEnum, a.Initial)
		b.WriteString("        transitions: [\n")

		for _, t := range a.FlattenedTransitions() {
			onKind := "internal"
			onValue := t.On.Desc
			if t.On.Kind == TriggerRecv {
				onKind = "recv"
				onValue = string(t.On.Msg)
			}

			b.WriteString("            { ")
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

		b.WriteString("        ],\n")
		b.WriteString("    };\n\n")
	}

	// Close namespace.
	b.WriteString("}\n\n")

	// Per-actor typed machine classes.
	for _, a := range p.Actors {
		if a.IsComposed() {
			writeTSComposedActor(&b, p, a, protoName, nsName, actions)
			continue
		}

		typeName := tsTypeName(a.Name)

		// Collect vars updated by this actor's transitions.
		actorVarSet := map[string]bool{}
		for _, t := range a.FlattenedTransitions() {
			for _, u := range t.Updates {
				actorVarSet[u.Var] = true
			}
		}

		// Machine class — prefix with protocol name to avoid collisions when
		// multiple protocols share actor names (e.g. session and pathswitch
		// both have "relay").
		machinePrefix := protoName + typeName
		stateEnum := protoName + typeName + "State"
		fmt.Fprintf(&b, "/** %sMachine is the generated state machine for the %s actor. */\n", machinePrefix, a.Name)
		fmt.Fprintf(&b, "export class %sMachine {\n", machinePrefix)
		// Convenience type aliases so callers can write machine.EventID etc.
		fmt.Fprintf(&b, "    readonly protocol = %s;\n", nsName)
		fmt.Fprintf(&b, "    state: %s;\n", stateEnum)

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
			fmt.Fprintf(&b, "    guards: Map<%s.GuardID, () => boolean> = new Map();\n", nsName)
		}
		if len(actions) > 0 {
			fmt.Fprintf(&b, "    actions: Map<%s.ActionID, () => void> = new Map();\n", nsName)
		}
		b.WriteString("\n")

		// Constructor.
		fmt.Fprintf(&b, "    constructor() {\n")
		fmt.Fprintf(&b, "        this.state = %s.%s;\n", stateEnum, a.Initial)
		b.WriteString("    }\n\n")

		// handleEvent method.
		if len(p.Commands) > 0 {
			fmt.Fprintf(&b, "    handleEvent(ev: %s.EventID): %s.CmdID[] {\n", nsName, nsName)
		} else {
			fmt.Fprintf(&b, "    handleEvent(ev: %s.EventID): string[] {\n", nsName)
		}
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
				guardCond = fmt.Sprintf(" && this.guards.get(%s.GuardID.%s)?.() === true", nsName, guardVal)
			}

			fmt.Fprintf(&b, "            case this.state === %s.%s && ev === %s.EventID.%s%s: {\n",
				stateEnum, t.From, nsName, eventVal, guardCond)

			// Action call.
			if t.Do != "" {
				actionVal := kotlinPascalCase(string(t.Do))
				fmt.Fprintf(&b, "                this.actions.get(%s.ActionID.%s)?.();\n", nsName, actionVal)
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
			fmt.Fprintf(&b, "                this.state = %s.%s;\n", stateEnum, t.To)

			// Return commands.
			if len(t.Emits) > 0 {
				b.WriteString("                return [")
				for i, cmd := range t.Emits {
					if i > 0 {
						b.WriteString(", ")
					}
					fmt.Fprintf(&b, "%s.CmdID.%s", nsName, kotlinPascalCase(string(cmd)))
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

// writeTSComposedActor emits per-sub-machine classes and a composite class
// with a route() method for a composed actor.
func writeTSComposedActor(b *strings.Builder, p *Protocol, a Actor, protoName, nsName string, actions []string) {
	actorTypeName := tsTypeName(a.Name)
	actorClass := protoName + actorTypeName

	// --- Per-sub-machine machine classes ---
	for _, m := range a.Machines {
		machName := tsTypeName(m.Name)
		machClass := actorClass + machName
		stateEnum := actorClass + machName + "State"

		// Collect vars updated by this sub-machine.
		varSet := map[string]bool{}
		for _, t := range m.FlattenedTransitions() {
			for _, u := range t.Updates {
				varSet[u.Var] = true
			}
		}

		fmt.Fprintf(b, "/** %sMachine is the generated state machine for %s/%s. */\n",
			machClass, a.Name, m.Name)
		fmt.Fprintf(b, "export class %sMachine {\n", machClass)
		fmt.Fprintf(b, "    readonly protocol = %s;\n", nsName)
		fmt.Fprintf(b, "    state: %s;\n", stateEnum)

		for _, v := range p.Vars {
			if !varSet[v.Name] {
				continue
			}
			tsT := tsVarType(v.Type)
			init := tsInitialValue(v)
			comment := ""
			if v.Desc != "" {
				comment = " // " + v.Desc
			}
			fmt.Fprintf(b, "    %s: %s = %s;%s\n", tsVarFieldName(v.Name), tsT, init, comment)
		}

		if len(p.Guards) > 0 {
			fmt.Fprintf(b, "    guards: Map<%s.GuardID, () => boolean> = new Map();\n", nsName)
		}
		if len(actions) > 0 {
			fmt.Fprintf(b, "    actions: Map<%s.ActionID, () => void> = new Map();\n", nsName)
		}
		b.WriteString("\n")

		// Constructor.
		fmt.Fprintf(b, "    constructor() {\n")
		fmt.Fprintf(b, "        this.state = %s.%s;\n", stateEnum, m.Initial)
		b.WriteString("    }\n\n")

		// step() — internal event dispatch.
		fmt.Fprintf(b, "    step(ev: %s.EventID): boolean {\n", nsName)
		b.WriteString("        switch (true) {\n")
		for _, t := range m.FlattenedTransitions() {
			if t.On.Kind != TriggerInternal {
				continue
			}
			eventVal := kotlinPascalCase(t.On.Desc)
			guardCond := ""
			if t.Guard != "" {
				guardVal := kotlinPascalCase(string(t.Guard))
				guardCond = fmt.Sprintf(" && this.guards.get(%s.GuardID.%s)?.() === true", nsName, guardVal)
			}
			fmt.Fprintf(b, "            case this.state === %s.%s && ev === %s.EventID.%s%s: {\n",
				stateEnum, t.From, nsName, eventVal, guardCond)
			if t.Do != "" {
				actionVal := kotlinPascalCase(string(t.Do))
				fmt.Fprintf(b, "                this.actions.get(%s.ActionID.%s)?.();\n", nsName, actionVal)
			}
			for _, u := range t.Updates {
				varField := tsVarFieldName(u.Var)
				if lit, ok := tsSimpleLiteral(u.Expr); ok {
					fmt.Fprintf(b, "                this.%s = %s;\n", varField, lit)
				} else {
					fmt.Fprintf(b, "                // %s: %s (set by action)\n", u.Var, u.Expr)
				}
			}
			fmt.Fprintf(b, "                this.state = %s.%s;\n", stateEnum, t.To)
			b.WriteString("                return true;\n")
			b.WriteString("            }\n")
		}
		b.WriteString("        }\n")
		b.WriteString("        return false;\n")
		b.WriteString("    }\n\n")

		// handleMessage() — recv message dispatch.
		b.WriteString("    handleMessage(msg: string): boolean {\n")
		b.WriteString("        switch (true) {\n")
		for _, t := range m.FlattenedTransitions() {
			if t.On.Kind != TriggerRecv {
				continue
			}
			msgVal := kotlinPascalCase(string(t.On.Msg))
			guardCond := ""
			if t.Guard != "" {
				guardVal := kotlinPascalCase(string(t.Guard))
				guardCond = fmt.Sprintf(" && this.guards.get(%s.GuardID.%s)?.() === true", nsName, guardVal)
			}
			fmt.Fprintf(b, "            case this.state === %s.%s && msg === %s.MessageType.%s%s: {\n",
				stateEnum, t.From, nsName, msgVal, guardCond)
			if t.Do != "" {
				actionVal := kotlinPascalCase(string(t.Do))
				fmt.Fprintf(b, "                this.actions.get(%s.ActionID.%s)?.();\n", nsName, actionVal)
			}
			for _, u := range t.Updates {
				varField := tsVarFieldName(u.Var)
				if lit, ok := tsSimpleLiteral(u.Expr); ok {
					fmt.Fprintf(b, "                this.%s = %s;\n", varField, lit)
				} else {
					fmt.Fprintf(b, "                // %s: %s (set by action)\n", u.Var, u.Expr)
				}
			}
			fmt.Fprintf(b, "                this.state = %s.%s;\n", stateEnum, t.To)
			b.WriteString("                return true;\n")
			b.WriteString("            }\n")
		}
		b.WriteString("        }\n")
		b.WriteString("        return false;\n")
		b.WriteString("    }\n")
		b.WriteString("}\n\n")
	}

	// --- Composite class ---
	compositeClass := actorClass + "Composite"
	fmt.Fprintf(b, "/** %s holds all sub-machines for the %s actor. */\n", compositeClass, a.Name)
	fmt.Fprintf(b, "export class %s {\n", compositeClass)
	for _, m := range a.Machines {
		machName := tsTypeName(m.Name)
		machClass := actorClass + machName
		fieldName := strings.ToLower(m.Name[:1]) + m.Name[1:]
		fmt.Fprintf(b, "    %s = new %sMachine();\n", fieldName, machClass)
	}
	if len(p.Guards) > 0 {
		fmt.Fprintf(b, "    routeGuards: Map<%s.GuardID, () => boolean> = new Map();\n", nsName)
	}
	b.WriteString("\n")

	// route() method.
	b.WriteString("    /** route dispatches inter-machine events according to the routing table. */\n")
	fmt.Fprintf(b, "    route(from: string, event: %s.EventID): void {\n", nsName)
	if len(a.Routes) > 0 {
		b.WriteString("        switch (true) {\n")
		for _, r := range a.Routes {
			eventVal := kotlinPascalCase(string(r.On))
			guardCond := ""
			if r.Guard != "" {
				guardVal := kotlinPascalCase(string(r.Guard))
				guardCond = fmt.Sprintf(" && this.routeGuards.get(%s.GuardID.%s)?.() === true", nsName, guardVal)
			}
			fmt.Fprintf(b, "            case from === %q && event === %s.EventID.%s%s: {\n",
				r.From, nsName, eventVal, guardCond)
			for _, s := range r.Sends {
				targetField := strings.ToLower(s.To[:1]) + s.To[1:]
				sendEventVal := kotlinPascalCase(string(s.Event))
				fmt.Fprintf(b, "                this.%s.step(%s.EventID.%s);\n",
					targetField, nsName, sendEventVal)
			}
			b.WriteString("                break;\n")
			b.WriteString("            }\n")
		}
		b.WriteString("        }\n")
	}
	b.WriteString("    }\n")
	b.WriteString("}\n\n")
}

// tsCollectEvents returns a deduplicated, ordered list of all event IDs:
// declared events, internal transition events, and recv_* events.
// Also collects events from sub-machines of composed actors.
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
	// Internal events from flat actor transitions and sub-machines.
	for _, a := range p.Actors {
		for _, t := range a.FlattenedTransitions() {
			if t.On.Kind == TriggerInternal {
				add(t.On.Desc)
			}
		}
		for _, m := range a.Machines {
			for _, t := range m.FlattenedTransitions() {
				if t.On.Kind == TriggerInternal {
					add(t.On.Desc)
				}
			}
		}
	}
	// recv_* events from flat actor transitions and sub-machines.
	for _, a := range p.Actors {
		for _, t := range a.FlattenedTransitions() {
			if t.On.Kind == TriggerRecv {
				add("recv_" + string(t.On.Msg))
			}
		}
		for _, m := range a.Machines {
			for _, t := range m.FlattenedTransitions() {
				if t.On.Kind == TriggerRecv {
					add("recv_" + string(t.On.Msg))
				}
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

// tsConstName converts a name to UPPER_SNAKE for TypeScript wire constants.
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

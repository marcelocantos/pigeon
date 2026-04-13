// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package protocol

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"unicode"
)

// ExportKotlin writes a Kotlin source file with:
//   - Enum classes for states (per actor) and message types
//   - EventID and CmdID enum classes
//   - Guard and action ID constants
//   - Transition table as static data
//   - Per-actor typed machine classes with handleEvent methods
func (p *Protocol) ExportKotlin(w io.Writer, pkg string) error {
	var b strings.Builder

	b.WriteString("// Copyright 2026 Marcelo Cantos\n")
	b.WriteString("// SPDX-License-Identifier: Apache-2.0\n\n")
	b.WriteString("// Auto-generated from protocol definition. Do not edit.\n")
	b.WriteString("// Source of truth: protocol/*.yaml\n\n")
	fmt.Fprintf(&b, "package %s\n\n", pkg)

	protoName := kotlinTypeName(p.Name)
	protoPrefix := protoName + "Protocol"

	// Per-actor state enums at package scope — prefixed with protocol name
	// to avoid collisions when multiple protocols share actor names.
	// For composed actors, emit one enum per sub-machine instead.
	for _, a := range p.Actors {
		typeName := kotlinTypeName(a.Name)
		if a.IsComposed() {
			for _, m := range a.Machines {
				machTypeName := kotlinTypeName(m.Name)
				stateEnum := protoName + typeName + machTypeName + "State"
				states := collectSubMachineStates(m)
				fmt.Fprintf(&b, "enum class %s(val value: String) {\n", stateEnum)
				for i, s := range states {
					comma := ","
					if i == len(states)-1 {
						comma = ";"
					}
					fmt.Fprintf(&b, "    %s(\"%s\")%s\n", string(s), s, comma)
				}
				b.WriteString("}\n\n")
			}
			continue
		}
		stateEnum := protoName + typeName + "State"
		states := collectStates(a)
		fmt.Fprintf(&b, "enum class %s(val value: String) {\n", stateEnum)
		for i, s := range states {
			comma := ","
			if i == len(states)-1 {
				comma = ";"
			}
			fmt.Fprintf(&b, "    %s(\"%s\")%s\n", string(s), s, comma)
		}
		b.WriteString("}\n\n")
	}

	// Protocol namespace — contains shared enums (MessageType, GuardID, ActionID,
	// EventID, CmdID), wire constants, and per-actor transition tables.
	// Nesting these prevents name collisions when multiple protocol files share
	// the same Kotlin package.
	b.WriteString("/** The protocol transition table and shared type enums. */\n")
	fmt.Fprintf(&b, "object %s {\n\n", protoPrefix)

	// Message type enum (nested).
	b.WriteString("    enum class MessageType(val value: String) {\n")
	for i, m := range p.Messages {
		comma := ","
		if i == len(p.Messages)-1 {
			comma = ";"
		}
		fmt.Fprintf(&b, "        %s(\"%s\")%s\n", kotlinPascalCase(string(m.Type)), m.Type, comma)
	}
	b.WriteString("    }\n\n")

	// Guard ID enum (nested).
	if len(p.Guards) > 0 {
		b.WriteString("    enum class GuardID(val value: String) {\n")
		for i, g := range p.Guards {
			comma := ","
			if i == len(p.Guards)-1 {
				comma = ";"
			}
			fmt.Fprintf(&b, "        %s(\"%s\")%s\n", kotlinPascalCase(string(g.ID)), g.ID, comma)
		}
		b.WriteString("    }\n\n")
	}

	// Action ID enum (nested).
	actions := collectActions(p)
	if len(actions) > 0 {
		b.WriteString("    enum class ActionID(val value: String) {\n")
		for i, id := range actions {
			comma := ","
			if i == len(actions)-1 {
				comma = ";"
			}
			fmt.Fprintf(&b, "        %s(\"%s\")%s\n", kotlinPascalCase(id), id, comma)
		}
		b.WriteString("    }\n\n")
	}

	// EventID enum (nested) — declared events + internal transition events + recv_* events.
	events := collectKotlinEvents(p)
	if len(events) > 0 {
		b.WriteString("    enum class EventID(val value: String) {\n")
		for i, id := range events {
			comma := ","
			if i == len(events)-1 {
				comma = ";"
			}
			fmt.Fprintf(&b, "        %s(\"%s\")%s\n", kotlinPascalCase(id), id, comma)
		}
		b.WriteString("    }\n\n")
	}

	// CmdID enum (nested) — from commands: section.
	if len(p.Commands) > 0 {
		b.WriteString("    enum class CmdID(val value: String) {\n")
		for i, c := range p.Commands {
			comma := ","
			if i == len(p.Commands)-1 {
				comma = ";"
			}
			fmt.Fprintf(&b, "        %s(\"%s\")%s\n", kotlinPascalCase(string(c.ID)), c.ID, comma)
		}
		b.WriteString("    }\n\n")
	}

	// Wire constants (nested).
	if len(p.WireConsts) > 0 {
		b.WriteString("    /** Protocol wire constants shared across all platforms. */\n")
		b.WriteString("    object Wire {\n")
		for _, wc := range p.WireConsts {
			name := kotlinConstName(wc.Name)
			switch wc.Type {
			case "byte":
				fmt.Fprintf(&b, "        const val %s: Byte = 0x%02X.toByte()\n", name, wireInt(wc.Value))
			case "int":
				fmt.Fprintf(&b, "        const val %s = %d\n", name, wireInt(wc.Value))
			case "duration_ms":
				fmt.Fprintf(&b, "        const val %s = %dL // ms\n", name, wireInt(wc.Value))
			case "string":
				fmt.Fprintf(&b, "        const val %s = %q\n", name, wc.Value)
			}
		}
		b.WriteString("    }\n\n")
	}

	// Transition table per actor (nested).
	// For composed actors, emit one table per sub-machine.
	for _, a := range p.Actors {
		typeName := kotlinTypeName(a.Name)
		if a.IsComposed() {
			for _, m := range a.Machines {
				machTypeName := kotlinTypeName(m.Name)
				stateEnum := protoName + typeName + machTypeName + "State"
				tableObj := typeName + machTypeName
				fmt.Fprintf(&b, "    /** %s/%s transition table. */\n", a.Name, m.Name)
				fmt.Fprintf(&b, "    object %sTable {\n", tableObj)
				fmt.Fprintf(&b, "        val initial = %s.%s\n\n", stateEnum, m.Initial)
				b.WriteString(kotlinTransitionDataClass("        "))
				b.WriteString("        val transitions = listOf(\n")
				for _, t := range m.FlattenedTransitions() {
					kotlinWriteTransitionEntry(&b, t, "            ")
				}
				b.WriteString("        )\n")
				b.WriteString("    }\n\n")
			}
			continue
		}
		stateEnum := protoName + typeName + "State"
		fmt.Fprintf(&b, "    /** %s transition table. */\n", a.Name)
		fmt.Fprintf(&b, "    object %sTable {\n", typeName)
		fmt.Fprintf(&b, "        val initial = %s.%s\n\n", stateEnum, a.Initial)
		b.WriteString(kotlinTransitionDataClass("        "))
		b.WriteString("        val transitions = listOf(\n")
		for _, t := range a.FlattenedTransitions() {
			kotlinWriteTransitionEntry(&b, t, "            ")
		}
		b.WriteString("        )\n")
		b.WriteString("    }\n\n")
	}

	// Close the protocol namespace object.
	b.WriteString("}\n\n")

	// Per-actor typed machine classes with handleEvent.
	// For composed actors, emit per-sub-machine classes and a composite class.
	for _, a := range p.Actors {
		typeName := kotlinTypeName(a.Name)
		if a.IsComposed() {
			writeKotlinComposedActor(&b, p, a, protoName, protoPrefix, actions)
			continue
		}

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
		fmt.Fprintf(&b, "class %sMachine {\n", machinePrefix)
		fmt.Fprintf(&b, "    var state: %s = %s.%s\n", stateEnum, stateEnum, a.Initial)
		b.WriteString("        private set\n")

		// Typed variable fields owned by this actor.
		for _, v := range p.Vars {
			if !actorVarSet[v.Name] {
				continue
			}
			comment := ""
			if v.Desc != "" {
				comment = " // " + v.Desc
			}
			fmt.Fprintf(&b, "    var %s: %s = %s%s\n",
				kotlinCamelCase(v.Name), kotlinType(v.Type), kotlinInitialValue(v), comment)
		}

		if len(p.Guards) > 0 {
			fmt.Fprintf(&b, "    val guards = mutableMapOf<%s.GuardID, () -> Boolean>()\n", protoPrefix)
		}
		if len(actions) > 0 {
			fmt.Fprintf(&b, "    val actions = mutableMapOf<%s.ActionID, () -> Unit>()\n", protoPrefix)
		}

		b.WriteString("\n")

		// handleEvent method — unified entry point returning commands.
		b.WriteString("    /** Handle an event and return the list of commands to execute. */\n")
		cmdType := "String"
		if len(p.Commands) > 0 {
			cmdType = protoPrefix + ".CmdID"
		}
		fmt.Fprintf(&b, "    fun handleEvent(ev: %s.EventID): List<%s> {\n", protoPrefix, cmdType)
		fmt.Fprintf(&b, "        val cmds: List<%s> = when {\n", cmdType)

		for _, t := range a.FlattenedTransitions() {
			// Determine event constant.
			var eventConst string
			if t.On.Kind == TriggerRecv {
				eventConst = protoPrefix + ".EventID." + kotlinPascalCase("recv_"+string(t.On.Msg))
			} else {
				eventConst = protoPrefix + ".EventID." + kotlinPascalCase(t.On.Desc)
			}

			// Guard condition.
			guardCond := ""
			if t.Guard != "" {
				guardCond = fmt.Sprintf(" && guards[%s.GuardID.%s]?.invoke() == true",
					protoPrefix, kotlinPascalCase(string(t.Guard)))
			}

			fmt.Fprintf(&b, "            state == %s.%s && ev == %s%s ->\n",
				stateEnum, t.From, eventConst, guardCond)

			// Transition body.
			b.WriteString("                run {\n")

			// Action call.
			if t.Do != "" {
				fmt.Fprintf(&b, "                    actions[%s.ActionID.%s]?.invoke()\n",
					protoPrefix, kotlinPascalCase(string(t.Do)))
			}

			// Variable updates.
			for _, u := range t.Updates {
				if lit, ok := kotlinSimpleLiteral(u.Expr); ok {
					fmt.Fprintf(&b, "                    %s = %s\n", kotlinCamelCase(u.Var), lit)
				} else {
					fmt.Fprintf(&b, "                    // %s: %s (set by action)\n", u.Var, u.Expr)
				}
			}

			// State transition.
			fmt.Fprintf(&b, "                    state = %s.%s\n", stateEnum, t.To)

			// Return commands.
			if len(t.Emits) > 0 {
				b.WriteString("                    listOf(")
				for i, cmd := range t.Emits {
					if i > 0 {
						b.WriteString(", ")
					}
					fmt.Fprintf(&b, "%s.CmdID.%s", protoPrefix, kotlinPascalCase(string(cmd)))
				}
				b.WriteString(")\n")
			} else {
				b.WriteString("                    emptyList()\n")
			}

			b.WriteString("                }\n")
		}

		b.WriteString("            else -> emptyList()\n")
		b.WriteString("        }\n")
		b.WriteString("        return cmds\n")
		b.WriteString("    }\n")
		b.WriteString("}\n\n")
	}

	_, err := io.WriteString(w, b.String())
	return err
}

// collectKotlinEvents returns a sorted, deduplicated list of all event IDs:
// declared events, internal transition events, and recv_* events.
// For composed actors, sub-machine transitions are included.
func collectKotlinEvents(p *Protocol) []string {
	seen := map[string]bool{}
	for _, e := range p.Events {
		seen[string(e.ID)] = true
	}
	addTransition := func(t Transition) {
		if t.On.Kind == TriggerInternal {
			seen[t.On.Desc] = true
		} else if t.On.Kind == TriggerRecv {
			seen["recv_"+string(t.On.Msg)] = true
		}
	}
	for _, a := range p.Actors {
		if a.IsComposed() {
			for _, m := range a.Machines {
				for _, t := range m.FlattenedTransitions() {
					addTransition(t)
				}
			}
			// Also include route event IDs (events reported by sub-machines).
			for _, r := range a.Routes {
				seen[string(r.On)] = true
				for _, s := range r.Sends {
					seen[string(s.Event)] = true
				}
			}
		} else {
			for _, t := range a.FlattenedTransitions() {
				addTransition(t)
			}
		}
	}
	result := make([]string, 0, len(seen))
	for id := range seen {
		result = append(result, id)
	}
	sort.Strings(result)
	return result
}

// kotlinType converts a VarType to its Kotlin equivalent.
func kotlinType(t VarType) string {
	switch t {
	case VarInt:
		return "Int"
	case VarBool:
		return "Boolean"
	default:
		return "String"
	}
}

// kotlinInitialValue converts a VarDef's initial expression to a Kotlin literal,
// falling back to a zero value if the expression is too complex.
func kotlinInitialValue(v VarDef) string {
	if lit, ok := kotlinSimpleLiteral(v.Initial); ok {
		return lit
	}
	switch v.Type {
	case VarInt:
		return "0"
	case VarBool:
		return "false"
	default:
		return "\"\""
	}
}

// kotlinSimpleLiteral converts a TLA+ expression to a Kotlin literal if it
// is simple enough (bool, int, string). Returns ("", false) for complex
// expressions that need action callbacks.
func kotlinSimpleLiteral(expr string) (string, bool) {
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

// kotlinCamelCase converts a snake_case or kebab-case string to lowerCamelCase.
func kotlinCamelCase(s string) string {
	var result []rune
	nextUpper := false
	for i, r := range s {
		if r == '_' || r == '-' {
			nextUpper = true
			continue
		}
		if nextUpper {
			result = append(result, unicode.ToUpper(r))
			nextUpper = false
		} else if i == 0 {
			result = append(result, unicode.ToLower(r))
		} else {
			result = append(result, r)
		}
	}
	return string(result)
}

func kotlinConstName(s string) string {
	return strings.ToUpper(s)
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
		if r == '_' || r == '-' || r == ' ' {
			nextUpper = true
			continue
		}
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			// Skip non-alphanumeric characters (e.g. '--' flags, punctuation).
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

// kotlinTransitionDataClass returns the shared Transition data class definition
// used in every transition table, indented by the given prefix.
func kotlinTransitionDataClass(indent string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%sdata class Transition(\n", indent)
	fmt.Fprintf(&b, "%s    val from: String,\n", indent)
	fmt.Fprintf(&b, "%s    val to: String,\n", indent)
	fmt.Fprintf(&b, "%s    val on: String,\n", indent)
	fmt.Fprintf(&b, "%s    val onKind: String,\n", indent)
	fmt.Fprintf(&b, "%s    val guard: String? = null,\n", indent)
	fmt.Fprintf(&b, "%s    val action: String? = null,\n", indent)
	fmt.Fprintf(&b, "%s    val sends: List<Pair<String, String>> = emptyList(),\n", indent)
	fmt.Fprintf(&b, "%s)\n\n", indent)
	return b.String()
}

// kotlinWriteTransitionEntry writes a single Transition(...) list entry for the table.
func kotlinWriteTransitionEntry(b *strings.Builder, t Transition, indent string) {
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
	fmt.Fprintf(b, "%sTransition(%q, %q, %q, %q, %s, %s, %s),\n",
		indent, t.From, t.To, onValue, onKind, guardStr, actionStr, sends)
}

// writeKotlinComposedActor emits per-sub-machine machine classes and a
// composite class with a route method for a composed actor.
func writeKotlinComposedActor(b *strings.Builder, p *Protocol, a Actor, protoName, protoPrefix string, actions []string) {
	typeName := kotlinTypeName(a.Name)
	cmdType := "String"
	if len(p.Commands) > 0 {
		cmdType = protoPrefix + ".CmdID"
	}

	// Per-sub-machine machine classes.
	for _, m := range a.Machines {
		machTypeName := kotlinTypeName(m.Name)
		machineClass := protoName + typeName + machTypeName + "Machine"
		stateEnum := protoName + typeName + machTypeName + "State"

		// Collect vars updated by this sub-machine.
		varSet := map[string]bool{}
		for _, t := range m.FlattenedTransitions() {
			for _, u := range t.Updates {
				varSet[u.Var] = true
			}
		}

		fmt.Fprintf(b, "/** %s is the generated state machine for %s/%s. */\n", machineClass, a.Name, m.Name)
		fmt.Fprintf(b, "class %s {\n", machineClass)
		fmt.Fprintf(b, "    var state: %s = %s.%s\n", stateEnum, stateEnum, m.Initial)
		b.WriteString("        private set\n")

		for _, v := range p.Vars {
			if !varSet[v.Name] {
				continue
			}
			comment := ""
			if v.Desc != "" {
				comment = " // " + v.Desc
			}
			fmt.Fprintf(b, "    var %s: %s = %s%s\n",
				kotlinCamelCase(v.Name), kotlinType(v.Type), kotlinInitialValue(v), comment)
		}
		if len(p.Guards) > 0 {
			fmt.Fprintf(b, "    val guards = mutableMapOf<%s.GuardID, () -> Boolean>()\n", protoPrefix)
		}
		if len(actions) > 0 {
			fmt.Fprintf(b, "    val actions = mutableMapOf<%s.ActionID, () -> Unit>()\n", protoPrefix)
		}
		b.WriteString("\n")

		// handleEvent — handles both internal and recv triggers.
		b.WriteString("    /** Handle an event and return the list of commands to execute. */\n")
		fmt.Fprintf(b, "    fun handleEvent(ev: %s.EventID): List<%s> {\n", protoPrefix, cmdType)
		fmt.Fprintf(b, "        val cmds: List<%s> = when {\n", cmdType)

		for _, t := range m.FlattenedTransitions() {
			var eventConst string
			if t.On.Kind == TriggerRecv {
				eventConst = protoPrefix + ".EventID." + kotlinPascalCase("recv_"+string(t.On.Msg))
			} else {
				eventConst = protoPrefix + ".EventID." + kotlinPascalCase(t.On.Desc)
			}
			guardCond := ""
			if t.Guard != "" {
				guardCond = fmt.Sprintf(" && guards[%s.GuardID.%s]?.invoke() == true",
					protoPrefix, kotlinPascalCase(string(t.Guard)))
			}
			fmt.Fprintf(b, "            state == %s.%s && ev == %s%s ->\n",
				stateEnum, t.From, eventConst, guardCond)
			b.WriteString("                run {\n")
			if t.Do != "" {
				fmt.Fprintf(b, "                    actions[%s.ActionID.%s]?.invoke()\n",
					protoPrefix, kotlinPascalCase(string(t.Do)))
			}
			for _, u := range t.Updates {
				if lit, ok := kotlinSimpleLiteral(u.Expr); ok {
					fmt.Fprintf(b, "                    %s = %s\n", kotlinCamelCase(u.Var), lit)
				} else {
					fmt.Fprintf(b, "                    // %s: %s (set by action)\n", u.Var, u.Expr)
				}
			}
			fmt.Fprintf(b, "                    state = %s.%s\n", stateEnum, t.To)
			if len(t.Emits) > 0 {
				b.WriteString("                    listOf(")
				for i, cmd := range t.Emits {
					if i > 0 {
						b.WriteString(", ")
					}
					fmt.Fprintf(b, "%s.CmdID.%s", protoPrefix, kotlinPascalCase(string(cmd)))
				}
				b.WriteString(")\n")
			} else {
				b.WriteString("                    emptyList()\n")
			}
			b.WriteString("                }\n")
		}

		b.WriteString("            else -> emptyList()\n")
		b.WriteString("        }\n")
		b.WriteString("        return cmds\n")
		b.WriteString("    }\n")
		b.WriteString("}\n\n")
	}

	// Composite class.
	compositeClass := protoName + typeName + "Composite"
	fmt.Fprintf(b, "/** %s holds all sub-machines for the %s actor. */\n", compositeClass, a.Name)
	fmt.Fprintf(b, "class %s {\n", compositeClass)
	for _, m := range a.Machines {
		machTypeName := kotlinTypeName(m.Name)
		machineClass := protoName + typeName + machTypeName + "Machine"
		fmt.Fprintf(b, "    val %s = %s()\n", kotlinCamelCase(m.Name), machineClass)
	}
	b.WriteString("\n")

	// route method.
	b.WriteString("    /** Route dispatches inter-machine events according to the routing table. */\n")
	fmt.Fprintf(b, "    fun route(from: String, event: %s.EventID) {\n", protoPrefix)
	if len(a.Routes) > 0 {
		b.WriteString("        when {\n")
		for _, r := range a.Routes {
			guardCond := ""
			if r.Guard != "" {
				// Route guards are not yet supported in Kotlin composite — emit a comment.
				guardCond = fmt.Sprintf(" /* guard: %s */", r.Guard)
			}
			eventConst := protoPrefix + ".EventID." + kotlinPascalCase(string(r.On))
			fmt.Fprintf(b, "            from == %q && event == %s%s -> {\n", r.From, eventConst, guardCond)
			for _, s := range r.Sends {
				deliverEvent := protoPrefix + ".EventID." + kotlinPascalCase(string(s.Event))
				fmt.Fprintf(b, "                %s.handleEvent(%s)\n", kotlinCamelCase(s.To), deliverEvent)
			}
			b.WriteString("            }\n")
		}
		b.WriteString("        }\n")
	}
	b.WriteString("    }\n")
	b.WriteString("}\n\n")
}

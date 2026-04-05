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

	// EventID enum — declared events + internal transition events + recv_* events.
	events := collectKotlinEvents(p)
	if len(events) > 0 {
		b.WriteString("enum class EventID(val value: String) {\n")
		for i, id := range events {
			comma := ","
			if i == len(events)-1 {
				comma = ";"
			}
			fmt.Fprintf(&b, "    %s(\"%s\")%s\n", kotlinPascalCase(id), id, comma)
		}
		b.WriteString("}\n\n")
	}

	// CmdID enum — from commands: section.
	if len(p.Commands) > 0 {
		b.WriteString("enum class CmdID(val value: String) {\n")
		for i, c := range p.Commands {
			comma := ","
			if i == len(p.Commands)-1 {
				comma = ";"
			}
			fmt.Fprintf(&b, "    %s(\"%s\")%s\n", kotlinPascalCase(string(c.ID)), c.ID, comma)
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

	// Per-actor typed machine classes with handleEvent.
	for _, a := range p.Actors {
		typeName := kotlinTypeName(a.Name)

		// Collect vars updated by this actor's transitions.
		actorVarSet := map[string]bool{}
		for _, t := range a.Transitions {
			for _, u := range t.Updates {
				actorVarSet[u.Var] = true
			}
		}

		// Machine class.
		fmt.Fprintf(&b, "/** %sMachine is the generated state machine for the %s actor. */\n", typeName, a.Name)
		fmt.Fprintf(&b, "class %sMachine {\n", typeName)
		fmt.Fprintf(&b, "    var state: %sState = %sState.%s\n", typeName, typeName, a.Initial)
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
			b.WriteString("    val guards = mutableMapOf<GuardID, () -> Boolean>()\n")
		}
		if len(actions) > 0 {
			b.WriteString("    val actions = mutableMapOf<ActionID, () -> Unit>()\n")
		}

		b.WriteString("\n")

		// handleEvent method — unified entry point returning commands.
		b.WriteString("    /** Handle an event and return the list of commands to execute. */\n")
		b.WriteString("    fun handleEvent(ev: EventID): List<CmdID> {\n")
		b.WriteString("        val cmds = when {\n")

		for _, t := range a.Transitions {
			// Determine event constant.
			var eventConst string
			if t.On.Kind == TriggerRecv {
				eventConst = "EventID." + kotlinPascalCase("recv_"+string(t.On.Msg))
			} else {
				eventConst = "EventID." + kotlinPascalCase(t.On.Desc)
			}

			// Guard condition.
			guardCond := ""
			if t.Guard != "" {
				guardCond = fmt.Sprintf(" && guards[GuardID.%s]?.invoke() == true",
					kotlinPascalCase(string(t.Guard)))
			}

			fmt.Fprintf(&b, "            state == %sState.%s && ev == %s%s ->\n",
				typeName, t.From, eventConst, guardCond)

			// Transition body.
			b.WriteString("                run {\n")

			// Action call.
			if t.Do != "" {
				fmt.Fprintf(&b, "                    actions[ActionID.%s]?.invoke()\n",
					kotlinPascalCase(string(t.Do)))
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
			fmt.Fprintf(&b, "                    state = %sState.%s\n", typeName, t.To)

			// Return commands.
			if len(t.Emits) > 0 {
				b.WriteString("                    listOf(")
				for i, cmd := range t.Emits {
					if i > 0 {
						b.WriteString(", ")
					}
					fmt.Fprintf(&b, "CmdID.%s", kotlinPascalCase(string(cmd)))
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
func collectKotlinEvents(p *Protocol) []string {
	seen := map[string]bool{}
	for _, e := range p.Events {
		seen[string(e.ID)] = true
	}
	for _, a := range p.Actors {
		for _, t := range a.Transitions {
			if t.On.Kind == TriggerInternal {
				seen[t.On.Desc] = true
			} else if t.On.Kind == TriggerRecv {
				seen["recv_"+string(t.On.Msg)] = true
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

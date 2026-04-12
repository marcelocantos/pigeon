// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package protocol

import (
	"fmt"
	"io"
	"strings"
)

// ExportC writes a C header and implementation file pair for the protocol.
// The header contains enums, machine structs, and function declarations.
// The implementation contains the state machine dispatch functions.
func (p *Protocol) ExportCHeader(w io.Writer) error {
	var b strings.Builder
	upper := strings.ToUpper(p.Name)
	prefix := cSnakePrefix(p.Name) // e.g. "pairing_ceremony"

	b.WriteString("// Copyright 2026 Marcelo Cantos\n")
	b.WriteString("// SPDX-License-Identifier: Apache-2.0\n\n")
	b.WriteString("// Code generated from protocol/*.yaml. DO NOT EDIT.\n\n")
	fmt.Fprintf(&b, "#ifndef PIGEON_%s_GEN_H\n", upper)
	fmt.Fprintf(&b, "#define PIGEON_%s_GEN_H\n\n", upper)
	b.WriteString("#include <stdbool.h>\n")
	b.WriteString("#include <stdint.h>\n\n")

	// --- Per-actor state enums ---
	for _, a := range p.Actors {
		states := collectStates(a)
		actorUpper := cConstPrefix(a.Name)
		fmt.Fprintf(&b, "// %s %s states.\n", p.Name, a.Name)
		fmt.Fprintf(&b, "typedef enum {\n")
		for i, s := range states {
			fmt.Fprintf(&b, "\tPIGEON_%s_%s", actorUpper, cConstName(string(s)))
			if i == 0 {
				b.WriteString(" = 0")
			}
			b.WriteString(",\n")
		}
		fmt.Fprintf(&b, "\tPIGEON_%s_STATE_COUNT\n", actorUpper)
		fmt.Fprintf(&b, "} pigeon_%s_state;\n\n", cSnake(a.Name))
	}

	// --- Message type enum ---
	fmt.Fprintf(&b, "// %s message types.\n", p.Name)
	b.WriteString("typedef enum {\n")
	for i, m := range p.Messages {
		fmt.Fprintf(&b, "\tPIGEON_MSG_%s", cConstName(string(m.Type)))
		if i == 0 {
			b.WriteString(" = 0")
		}
		b.WriteString(",\n")
	}
	b.WriteString("\tPIGEON_MSG_COUNT\n")
	fmt.Fprintf(&b, "} %s_msg_type;\n\n", prefix)

	// --- Guard ID enum ---
	if len(p.Guards) > 0 {
		fmt.Fprintf(&b, "// %s guards.\n", p.Name)
		b.WriteString("typedef enum {\n")
		for i, g := range p.Guards {
			fmt.Fprintf(&b, "\tPIGEON_GUARD_%s", cConstName(string(g.ID)))
			if i == 0 {
				b.WriteString(" = 0")
			}
			b.WriteString(",\n")
		}
		b.WriteString("\tPIGEON_GUARD_COUNT\n")
		fmt.Fprintf(&b, "} %s_guard_id;\n\n", prefix)
	}

	// --- Action ID enum ---
	actions := collectActions(p)
	if len(actions) > 0 {
		fmt.Fprintf(&b, "// %s actions.\n", p.Name)
		b.WriteString("typedef enum {\n")
		for i, id := range actions {
			fmt.Fprintf(&b, "\tPIGEON_ACTION_%s", cConstName(id))
			if i == 0 {
				b.WriteString(" = 0")
			}
			b.WriteString(",\n")
		}
		b.WriteString("\tPIGEON_ACTION_COUNT\n")
		fmt.Fprintf(&b, "} %s_action_id;\n\n", prefix)
	}

	// --- Event ID enum ---
	events := collectAllEvents(p)
	if len(events) > 0 {
		fmt.Fprintf(&b, "// %s events.\n", p.Name)
		b.WriteString("typedef enum {\n")
		for i, id := range events {
			fmt.Fprintf(&b, "\tPIGEON_EVENT_%s", cConstName(id))
			if i == 0 {
				b.WriteString(" = 0")
			}
			b.WriteString(",\n")
		}
		b.WriteString("\tPIGEON_EVENT_COUNT\n")
		fmt.Fprintf(&b, "} %s_event_id;\n\n", prefix)
	}

	// --- Command ID enum ---
	if len(p.Commands) > 0 {
		fmt.Fprintf(&b, "// %s commands.\n", p.Name)
		b.WriteString("typedef enum {\n")
		for i, c := range p.Commands {
			fmt.Fprintf(&b, "\tPIGEON_CMD_%s", cConstName(string(c.ID)))
			if i == 0 {
				b.WriteString(" = 0")
			}
			b.WriteString(",\n")
		}
		b.WriteString("\tPIGEON_CMD_COUNT\n")
		fmt.Fprintf(&b, "} %s_cmd_id;\n\n", prefix)
	}

	// --- Wire constants ---
	if len(p.WireConsts) > 0 {
		b.WriteString("// Wire constants.\n")
		for _, wc := range p.WireConsts {
			name := "PIGEON_WIRE_" + cConstName(wc.Name)
			switch wc.Type {
			case "byte":
				fmt.Fprintf(&b, "#define %s ((uint8_t)0x%02X)", name, wireInt(wc.Value))
			case "int":
				fmt.Fprintf(&b, "#define %s %d", name, wireInt(wc.Value))
			case "duration_ms":
				fmt.Fprintf(&b, "#define %s %d /* ms */", name, wireInt(wc.Value))
			case "string":
				fmt.Fprintf(&b, "#define %s %q", name, wc.Value)
			}
			if wc.Desc != "" {
				fmt.Fprintf(&b, " // %s", wc.Desc)
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	// --- Callback types ---
	b.WriteString("// Guard and action callback types.\n")
	b.WriteString("typedef bool (*pigeon_guard_fn)(void *ctx);\n")
	b.WriteString("typedef int  (*pigeon_action_fn)(void *ctx);\n")
	b.WriteString("typedef void (*pigeon_change_fn)(const char *var_name, void *ctx);\n\n")

	// --- Per-actor machine structs ---
	for _, a := range p.Actors {
		actorSnake := cSnake(a.Name)

		// Collect vars updated by this actor.
		actorVarSet := map[string]bool{}
		for _, t := range a.FlattenedTransitions() {
			for _, u := range t.Updates {
				actorVarSet[u.Var] = true
			}
		}

		fmt.Fprintf(&b, "// %s %s state machine.\n", p.Name, a.Name)
		fmt.Fprintf(&b, "typedef struct {\n")
		fmt.Fprintf(&b, "\tpigeon_%s_state state;\n", actorSnake)

		// Typed variable fields (skip set types — managed by actions).
		for _, v := range p.Vars {
			if !actorVarSet[v.Name] {
				continue
			}
			ct := cType(v.Type)
			if ct == "" {
				continue // set<string> etc. — action-managed
			}
			fmt.Fprintf(&b, "\t%s %s;", ct, cSnake(v.Name))
			if v.Desc != "" {
				fmt.Fprintf(&b, " // %s", v.Desc)
			}
			b.WriteString("\n")
		}

		guardCount := "PIGEON_GUARD_COUNT"
		actionCount := "PIGEON_ACTION_COUNT"
		if len(p.Guards) > 0 {
			fmt.Fprintf(&b, "\tpigeon_guard_fn guards[%s];\n", guardCount)
		}
		if len(actions) > 0 {
			fmt.Fprintf(&b, "\tpigeon_action_fn actions[%s];\n", actionCount)
		}
		b.WriteString("\tpigeon_change_fn on_change;\n")
		b.WriteString("\tvoid *userdata;\n")
		fmt.Fprintf(&b, "} pigeon_%s_machine;\n\n", actorSnake)

		// Function declarations.
		fmt.Fprintf(&b, "void pigeon_%s_machine_init(pigeon_%s_machine *m);\n", actorSnake, actorSnake)
		fmt.Fprintf(&b, "int  pigeon_%s_handle_message(pigeon_%s_machine *m, %s_msg_type msg);\n",
			actorSnake, actorSnake, prefix)
		fmt.Fprintf(&b, "int  pigeon_%s_step(pigeon_%s_machine *m, %s_event_id event);\n",
			actorSnake, actorSnake, prefix)

		b.WriteString("\n")
	}

	fmt.Fprintf(&b, "#endif // PIGEON_%s_GEN_H\n", upper)

	_, err := io.WriteString(w, b.String())
	return err
}

// ExportCImpl writes the C implementation file for the protocol state machines.
func (p *Protocol) ExportCImpl(w io.Writer) error {
	var b strings.Builder
	prefix := cSnakePrefix(p.Name)
	lowerName := strings.ToLower(p.Name)

	b.WriteString("// Copyright 2026 Marcelo Cantos\n")
	b.WriteString("// SPDX-License-Identifier: Apache-2.0\n\n")
	b.WriteString("// Code generated from protocol/*.yaml. DO NOT EDIT.\n\n")
	fmt.Fprintf(&b, "#include \"%s_gen.h\"\n", lowerName)
	b.WriteString("#include <string.h>\n\n")

	for _, a := range p.Actors {
		actorSnake := cSnake(a.Name)
		actorUpper := cConstPrefix(a.Name)

		// Collect vars updated by this actor.
		actorVarSet := map[string]bool{}
		for _, t := range a.FlattenedTransitions() {
			for _, u := range t.Updates {
				actorVarSet[u.Var] = true
			}
		}

		// --- Init ---
		fmt.Fprintf(&b, "void pigeon_%s_machine_init(pigeon_%s_machine *m)\n{\n", actorSnake, actorSnake)
		b.WriteString("\tmemset(m, 0, sizeof(*m));\n")
		fmt.Fprintf(&b, "\tm->state = PIGEON_%s_%s;\n", actorUpper, cConstName(string(a.Initial)))

		// Initial values for simple vars.
		for _, v := range p.Vars {
			if !actorVarSet[v.Name] {
				continue
			}
			ct := cType(v.Type)
			if ct == "" {
				continue
			}
			if init := cInitialValue(v); init != "" {
				fmt.Fprintf(&b, "\tm->%s = %s;\n", cSnake(v.Name), init)
			}
		}
		b.WriteString("}\n\n")

		// --- HandleMessage ---
		fmt.Fprintf(&b, "int pigeon_%s_handle_message(pigeon_%s_machine *m, %s_msg_type msg)\n{\n",
			actorSnake, actorSnake, prefix)

		hasRecv := false
		for _, t := range a.FlattenedTransitions() {
			if t.On.Kind == TriggerRecv {
				hasRecv = true
				break
			}
		}

		if hasRecv {
			for _, t := range a.FlattenedTransitions() {
				if t.On.Kind != TriggerRecv {
					continue
				}
				guard := ""
				if t.Guard != "" {
					guard = fmt.Sprintf(" && m->guards[PIGEON_GUARD_%s] && m->guards[PIGEON_GUARD_%s](m->userdata)",
						cConstName(string(t.Guard)), cConstName(string(t.Guard)))
				}
				fmt.Fprintf(&b, "\tif (m->state == PIGEON_%s_%s && msg == PIGEON_MSG_%s%s) {\n",
					actorUpper, cConstName(string(t.From)),
					cConstName(string(t.On.Msg)), guard)

				writeCTransitionBody(&b, t, actorUpper)
				b.WriteString("\t\treturn 1;\n")
				b.WriteString("\t}\n")
			}
		}

		b.WriteString("\treturn 0;\n}\n\n")

		// --- Step ---
		fmt.Fprintf(&b, "int pigeon_%s_step(pigeon_%s_machine *m, %s_event_id event)\n{\n",
			actorSnake, actorSnake, prefix)

		hasInternal := false
		for _, t := range a.FlattenedTransitions() {
			if t.On.Kind == TriggerInternal {
				hasInternal = true
				break
			}
		}

		if hasInternal {
			for _, t := range a.FlattenedTransitions() {
				if t.On.Kind != TriggerInternal {
					continue
				}
				guard := ""
				if t.Guard != "" {
					guard = fmt.Sprintf(" && m->guards[PIGEON_GUARD_%s] && m->guards[PIGEON_GUARD_%s](m->userdata)",
						cConstName(string(t.Guard)), cConstName(string(t.Guard)))
				}
				fmt.Fprintf(&b, "\tif (m->state == PIGEON_%s_%s && event == PIGEON_EVENT_%s%s) {\n",
					actorUpper, cConstName(string(t.From)),
					cConstName(t.On.Desc), guard)

				writeCTransitionBody(&b, t, actorUpper)
				b.WriteString("\t\treturn 1;\n")
				b.WriteString("\t}\n")
			}
		}

		b.WriteString("\treturn 0;\n}\n\n")
	}

	_, err := io.WriteString(w, b.String())
	return err
}

// writeCTransitionBody emits the action call, variable updates, and state
// change for a single transition.
func writeCTransitionBody(b *strings.Builder, t Transition, actorUpper string) {
	// Action callback.
	if t.Do != "" {
		fmt.Fprintf(b, "\t\tif (m->actions[PIGEON_ACTION_%s]) {\n", cConstName(string(t.Do)))
		fmt.Fprintf(b, "\t\t\tint err = m->actions[PIGEON_ACTION_%s](m->userdata);\n", cConstName(string(t.Do)))
		b.WriteString("\t\t\tif (err) return -err;\n")
		b.WriteString("\t\t}\n")
	}

	// Variable updates — emit simple assignments directly.
	for _, u := range t.Updates {
		field := cSnake(u.Var)
		if lit, ok := cSimpleLiteral(u.Expr); ok {
			fmt.Fprintf(b, "\t\tm->%s = %s;\n", field, lit)
			fmt.Fprintf(b, "\t\tif (m->on_change) m->on_change(%q, m->userdata);\n", u.Var)
		} else if cExpr, ok := cSelfUpdate(u.Var, u.Expr); ok {
			fmt.Fprintf(b, "\t\tm->%s = %s;\n", field, cExpr)
			fmt.Fprintf(b, "\t\tif (m->on_change) m->on_change(%q, m->userdata);\n", u.Var)
		} else {
			// Complex expression — handled by action callback.
			fmt.Fprintf(b, "\t\t// %s: %s (set by action)\n", u.Var, u.Expr)
		}
	}

	// State transition.
	fmt.Fprintf(b, "\t\tm->state = PIGEON_%s_%s;\n", actorUpper, cConstName(string(t.To)))
}

// cSnakePrefix returns the snake_case protocol prefix (e.g. "pairing_ceremony").
func cSnakePrefix(name string) string {
	return cSnake(name)
}

// cSnake converts CamelCase or snake_case to lower_snake_case.
// Handles acronyms correctly: "ScanQR" → "scan_qr", "E2EReady" → "e2e_ready",
// "ECDH" → "ecdh". Digits don't trigger word breaks.
func cSnake(s string) string {
	var b strings.Builder
	runes := []rune(s)
	for i, r := range runes {
		if r == '_' || r == '-' || r == ' ' {
			b.WriteByte('_')
			continue
		}
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				prev := runes[i-1]
				prevIsLetter := (prev >= 'a' && prev <= 'z') || (prev >= 'A' && prev <= 'Z')
				prevLower := prev >= 'a' && prev <= 'z'
				nextLower := i+1 < len(runes) && runes[i+1] >= 'a' && runes[i+1] <= 'z'
				// Insert underscore when:
				// - previous letter was lowercase ("nQ" → "n_q"), OR
				// - previous was a letter AND next is lowercase (end of
				//   acronym: "QRp" → "qr_p" but not "QR$" → "qr")
				if prevLower || (prevIsLetter && nextLower) {
					b.WriteByte('_')
				}
			}
			b.WriteByte(byte(r - 'A' + 'a'))
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// cConstName converts a name to UPPER_SNAKE_CASE for C constants.
// Collapses runs of underscores and trims leading/trailing underscores.
func cConstName(s string) string {
	snake := strings.ToUpper(cSnake(s))
	// Collapse multiple underscores.
	for strings.Contains(snake, "__") {
		snake = strings.ReplaceAll(snake, "__", "_")
	}
	return strings.Trim(snake, "_")
}

// cConstPrefix returns the UPPER_SNAKE actor prefix.
func cConstPrefix(actor string) string {
	switch actor {
	case "ios":
		return "APP"
	case "cli":
		return "CLI"
	default:
		return strings.ToUpper(actor)
	}
}

// cType maps VarType to a C type string. Returns "" for types that
// can't be directly represented (set<string>).
func cType(t VarType) string {
	switch t {
	case VarInt:
		return "int"
	case VarBool:
		return "bool"
	case VarString, "":
		return "const char *"
	default:
		return "" // set<string> etc. — not representable
	}
}

// cSimpleLiteral converts a TLA+ expression to a C literal if simple.
func cSimpleLiteral(expr string) (string, bool) {
	expr = strings.TrimSpace(expr)
	switch expr {
	case "TRUE":
		return "true", true
	case "FALSE":
		return "false", true
	}
	if strings.HasPrefix(expr, "\"") && strings.HasSuffix(expr, "\"") {
		return expr, true // C string literal
	}
	var n int
	if _, err := fmt.Sscanf(expr, "%d", &n); err == nil {
		return fmt.Sprintf("%d", n), true
	}
	return "", false
}

// cSelfUpdate handles "var_name + N" / "var_name - N" patterns.
func cSelfUpdate(varName, expr string) (string, bool) {
	expr = strings.TrimSpace(expr)
	for _, op := range []string{" + ", " - "} {
		if rest, ok := strings.CutPrefix(expr, varName+op); ok {
			rest = strings.TrimSpace(rest)
			var n int
			if _, err := fmt.Sscanf(rest, "%d", &n); err == nil {
				return fmt.Sprintf("m->%s%s%d", cSnake(varName), op, n), true
			}
		}
	}
	return "", false
}

// cInitialValue converts a VarDef's initial value to a C literal.
func cInitialValue(v VarDef) string {
	if lit, ok := cSimpleLiteral(v.Initial); ok {
		return lit
	}
	// Default zero values — memset already zeroed, so skip.
	return ""
}

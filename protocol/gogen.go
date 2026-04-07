// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package protocol

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// ExportGo writes a Go source file with:
//   - State/message/guard/action/event enum constants
//   - Protocol table literal (*Protocol)
//   - Typed structs from YAML struct definitions
//   - Per-actor typed state machines with HandleMessage/Step
func (p *Protocol) ExportGo(w io.Writer, pkgName, funcName string) error {
	var b strings.Builder

	b.WriteString("// Copyright 2026 Marcelo Cantos\n")
	b.WriteString("// SPDX-License-Identifier: Apache-2.0\n\n")
	b.WriteString("// Code generated from protocol/*.yaml. DO NOT EDIT.\n\n")
	fmt.Fprintf(&b, "package %s\n\n", pkgName)

	// Detect if frozen import is needed.
	needsFrozen := false
	for _, v := range p.Vars {
		if v.Type == VarSetString {
			needsFrozen = true
			break
		}
	}
	if !needsFrozen {
		for _, sd := range p.Structs {
			for _, f := range sd.Fields {
				if f.Type == VarSetString {
					needsFrozen = true
					break
				}
			}
		}
	}

	// When generating into the protocol package itself, skip the
	// self-import and type aliases.
	selfPkg := pkgName == "protocol"
	if selfPkg {
		if needsFrozen {
			b.WriteString("import \"github.com/arr-ai/frozen\"\n\n")
			b.WriteString("var _ frozen.Set[string] // suppress unused import\n\n")
		}
	} else {
		b.WriteString("import (\n")
		b.WriteString("\t\"github.com/marcelocantos/pigeon/protocol\"\n")
		if needsFrozen {
			b.WriteString("\t\"github.com/arr-ai/frozen\"\n")
		}
		b.WriteString(")\n\n")

		b.WriteString("type (\n")
		b.WriteString("\tState      = protocol.State\n")
		b.WriteString("\tMsgType    = protocol.MsgType\n")
		b.WriteString("\tGuardID    = protocol.GuardID\n")
		b.WriteString("\tActionID   = protocol.ActionID\n")
		b.WriteString("\tEventID    = protocol.EventID\n")
		b.WriteString("\tCmdID      = protocol.CmdID\n")
		b.WriteString("\tProtocol   = protocol.Protocol\n")
		b.WriteString("\tActor      = protocol.Actor\n")
		b.WriteString("\tTransition = protocol.Transition\n")
		b.WriteString("\tSend       = protocol.Send\n")
		b.WriteString("\tMessage    = protocol.Message\n")
		b.WriteString("\tVarDef     = protocol.VarDef\n")
		b.WriteString("\tVarUpdate  = protocol.VarUpdate\n")
		b.WriteString("\tGuardDef   = protocol.GuardDef\n")
		b.WriteString("\tOperator   = protocol.Operator\n")
		b.WriteString("\tAdvAction  = protocol.AdvAction\n")
		b.WriteString("\tProperty   = protocol.Property\n")
		b.WriteString(")\n\n")

		// Constructor aliases.
		b.WriteString("var (\n")
		b.WriteString("\tRecv     = protocol.Recv\n")
		b.WriteString("\tInternal = protocol.Internal\n")
		b.WriteString("\tInvariant = protocol.Invariant\n")
		b.WriteString("\tLiveness  = protocol.Liveness\n")
		b.WriteString(")\n\n")

		if needsFrozen {
			b.WriteString("var _ frozen.Set[string] // suppress unused import\n\n")
		}
	}

	// --- Enum constants ---
	// All constants are prefixed with the protocol name to avoid
	// collisions when multiple protocols are generated into the same package.
	cPrefix := funcName // e.g. "PairingCeremony", "Session"

	for _, a := range p.Actors {
		states := collectStates(a)
		prefix := cPrefix + goConstPrefix(a.Name)
		fmt.Fprintf(&b, "// %s %s states.\n", funcName, a.Name)
		b.WriteString("const (\n")
		for _, s := range states {
			fmt.Fprintf(&b, "\t%s%s State = %q\n", prefix, s, s)
		}
		b.WriteString(")\n\n")
	}

	fmt.Fprintf(&b, "// %s message types.\nconst (\n", funcName)
	for _, m := range p.Messages {
		fmt.Fprintf(&b, "\t%sMsg%s MsgType = %q\n", cPrefix, goCamel(string(m.Type)), m.Type)
	}
	b.WriteString(")\n\n")

	fmt.Fprintf(&b, "// %s guards.\nconst (\n", funcName)
	for _, g := range p.Guards {
		fmt.Fprintf(&b, "\t%sGuard%s GuardID = %q\n", cPrefix, goCamel(string(g.ID)), g.ID)
	}
	b.WriteString(")\n\n")

	actions := map[string]bool{}
	events := map[string]bool{}
	for _, a := range p.Actors {
		for _, t := range a.FlattenedTransitions() {
			if t.Do != "" {
				actions[string(t.Do)] = true
			}
			if t.On.Kind == TriggerInternal {
				events[t.On.Desc] = true
			}
		}
	}
	if len(actions) > 0 {
		sortedActions := sortedKeys(actions)
		fmt.Fprintf(&b, "// %s actions.\nconst (\n", funcName)
		for _, id := range sortedActions {
			fmt.Fprintf(&b, "\t%sAction%s ActionID = %q\n", cPrefix, goCamel(id), id)
		}
		b.WriteString(")\n\n")
	}

	// Collect message-receipt events (recv X → EventRecvX).
	for _, a := range p.Actors {
		for _, t := range a.FlattenedTransitions() {
			if t.On.Kind == TriggerRecv {
				events["recv_"+string(t.On.Msg)] = true
			}
		}
	}
	// Add declared events from the events: section.
	for _, e := range p.Events {
		events[string(e.ID)] = true
	}

	if len(events) > 0 {
		sortedEvents := sortedKeys(events)
		fmt.Fprintf(&b, "// %s events.\nconst (\n", funcName)
		for _, id := range sortedEvents {
			fmt.Fprintf(&b, "\t%sEvent%s EventID = %q\n", cPrefix, goCamel(id), id)
		}
		b.WriteString(")\n\n")
	}

	// Command IDs from the commands: section.
	if len(p.Commands) > 0 {
		fmt.Fprintf(&b, "// %s commands.\nconst (\n", funcName)
		for _, c := range p.Commands {
			fmt.Fprintf(&b, "\t%sCmd%s CmdID = %q\n", cPrefix, goCamel(string(c.ID)), c.ID)
		}
		b.WriteString(")\n\n")
	}

	// Wire constants — protocol constants for all platforms.
	if len(p.WireConsts) > 0 {
		b.WriteString("// Wire constants — protocol-level values shared across all platforms.\n")
		// Group constants.
		lastGroup := ""
		for _, wc := range p.WireConsts {
			if wc.Group != lastGroup {
				if lastGroup != "" {
					b.WriteString(")\n\n")
				}
				fmt.Fprintf(&b, "// %s\n", goCamel(wc.Group))
				b.WriteString("const (\n")
				lastGroup = wc.Group
			}
			name := goCamel(wc.Name)
			switch wc.Type {
			case "byte":
				fmt.Fprintf(&b, "\t%s byte = 0x%02X", name, wireInt(wc.Value))
			case "int":
				fmt.Fprintf(&b, "\t%s = %d", name, wireInt(wc.Value))
			case "duration_ms":
				fmt.Fprintf(&b, "\t%s = %d // ms", name, wireInt(wc.Value))
			case "string":
				fmt.Fprintf(&b, "\t%s = %q", name, wc.Value)
			}
			if wc.Desc != "" {
				fmt.Fprintf(&b, " // %s", wc.Desc)
			}
			b.WriteString("\n")
		}
		if lastGroup != "" {
			b.WriteString(")\n\n")
		}
	}

	// --- Protocol table literal ---

	fmt.Fprintf(&b, "func %s() *Protocol {\n", funcName)
	b.WriteString("\treturn &Protocol{\n")
	fmt.Fprintf(&b, "\t\tName: %q,\n", p.Name)

	b.WriteString("\t\tActors: []Actor{\n")
	for _, a := range p.Actors {
		fmt.Fprintf(&b, "\t\t\t{Name: %q, Initial: %q, Transitions: []Transition{\n", a.Name, a.Initial)
		for _, t := range a.FlattenedTransitions() {
			b.WriteString("\t\t\t\t{")
			fmt.Fprintf(&b, "From: %q, To: %q, ", t.From, t.To)
			if t.On.Kind == TriggerRecv {
				fmt.Fprintf(&b, "On: Recv(%q)", t.On.Msg)
			} else {
				fmt.Fprintf(&b, "On: Internal(%q)", t.On.Desc)
			}
			if t.Guard != "" {
				fmt.Fprintf(&b, ", Guard: %q", t.Guard)
			}
			if t.Do != "" {
				fmt.Fprintf(&b, ", Do: %q", t.Do)
			}
			if len(t.Sends) > 0 {
				b.WriteString(", Sends: []Send{")
				for _, s := range t.Sends {
					fmt.Fprintf(&b, "{To: %q, Msg: %q", s.To, s.Msg)
					if len(s.Fields) > 0 {
						b.WriteString(", Fields: map[string]string{")
						for _, k := range sortedKeys(s.Fields) {
							fmt.Fprintf(&b, "%q: %q, ", k, s.Fields[k])
						}
						b.WriteString("}")
					}
					b.WriteString("}, ")
				}
				b.WriteString("}")
			}
			if len(t.Updates) > 0 {
				b.WriteString(", Updates: []VarUpdate{")
				for _, u := range t.Updates {
					fmt.Fprintf(&b, "{Var: %q, Expr: %q}, ", u.Var, u.Expr)
				}
				b.WriteString("}")
			}
			b.WriteString("},\n")
		}
		b.WriteString("\t\t\t}},\n")
	}
	b.WriteString("\t\t},\n")

	b.WriteString("\t\tMessages: []Message{\n")
	for _, m := range p.Messages {
		fmt.Fprintf(&b, "\t\t\t{Type: %q, From: %q, To: %q, Desc: %q},\n", m.Type, m.From, m.To, m.Desc)
	}
	b.WriteString("\t\t},\n")

	b.WriteString("\t\tVars: []VarDef{\n")
	for _, v := range p.Vars {
		fmt.Fprintf(&b, "\t\t\t{Name: %q, Initial: %q, Desc: %q},\n", v.Name, v.Initial, v.Desc)
	}
	b.WriteString("\t\t},\n")

	b.WriteString("\t\tGuards: []GuardDef{\n")
	for _, g := range p.Guards {
		fmt.Fprintf(&b, "\t\t\t{ID: %q, Expr: %q},\n", g.ID, g.Expr)
	}
	b.WriteString("\t\t},\n")

	b.WriteString("\t\tOperators: []Operator{\n")
	for _, op := range p.Operators {
		fmt.Fprintf(&b, "\t\t\t{Name: %q, Params: %q, Expr: %q, Desc: %q},\n",
			op.Name, op.Params, op.Expr, op.Desc)
	}
	b.WriteString("\t\t},\n")

	b.WriteString("\t\tAdvActions: []AdvAction{\n")
	for _, aa := range p.AdvActions {
		fmt.Fprintf(&b, "\t\t\t{Name: %q, Desc: %q, Code: %q},\n", aa.Name, aa.Desc, aa.Code)
	}
	b.WriteString("\t\t},\n")

	b.WriteString("\t\tProperties: []Property{\n")
	for _, prop := range p.Properties {
		kindStr := "Invariant"
		if prop.Kind == Liveness {
			kindStr = "Liveness"
		}
		fmt.Fprintf(&b, "\t\t\t{Name: %q, Kind: %s, Expr: %q, Desc: %q},\n",
			prop.Name, kindStr, prop.Expr, prop.Desc)
	}
	b.WriteString("\t\t},\n")

	fmt.Fprintf(&b, "\t\tChannelBound: %d,\n", p.ChannelBound)
	fmt.Fprintf(&b, "\t\tOneShot: %v,\n", p.OneShot)

	b.WriteString("\t}\n}\n\n")

	// --- Structs ---

	if len(p.Structs) > 0 {
		for _, sd := range p.Structs {
			if sd.Desc != "" {
				fmt.Fprintf(&b, "// %s\n", sd.Desc)
			}
			fmt.Fprintf(&b, "type %s struct {\n", goCamel(sd.Name))
			for _, f := range sd.Fields {
				fmt.Fprintf(&b, "\t%s %s", goCamel(f.Name), goType(f.Type))
				if f.Desc != "" {
					fmt.Fprintf(&b, " // %s", f.Desc)
				}
				b.WriteString("\n")
			}
			b.WriteString("}\n\n")
		}
	}

	// --- Per-actor typed state machines ---

	for _, a := range p.Actors {
		prefix := cPrefix + goCamel(goConstPrefix(a.Name))
		stateType := cPrefix + goConstPrefix(a.Name)

		// Collect vars updated by this actor's transitions.
		actorVarSet := map[string]bool{}
		for _, t := range a.FlattenedTransitions() {
			for _, u := range t.Updates {
				actorVarSet[u.Var] = true
			}
		}

		// Machine struct.
		fmt.Fprintf(&b, "// %sMachine is the generated state machine for the %s actor.\n", prefix, a.Name)
		fmt.Fprintf(&b, "type %sMachine struct {\n", prefix)
		fmt.Fprintf(&b, "\tState State\n")

		// Typed variable fields owned by this actor.
		for _, v := range p.Vars {
			if !actorVarSet[v.Name] {
				continue
			}
			fmt.Fprintf(&b, "\t%s %s", goCamel(v.Name), goType(v.Type))
			if v.Desc != "" {
				fmt.Fprintf(&b, " // %s", v.Desc)
			}
			b.WriteString("\n")
		}

		b.WriteString("\n\tGuards  map[GuardID]func() bool\n")
		b.WriteString("\tActions map[ActionID]func() error\n")
		b.WriteString("\tOnChange func(varName string)\n")
		b.WriteString("}\n\n")

		// Constructor.
		fmt.Fprintf(&b, "func New%sMachine() *%sMachine {\n", prefix, prefix)
		fmt.Fprintf(&b, "\treturn &%sMachine{\n", prefix)
		fmt.Fprintf(&b, "\t\tState: %s%s,\n", stateType, goCamel(string(a.Initial)))

		// Initial values for owned vars.
		for _, v := range p.Vars {
			if !actorVarSet[v.Name] {
				continue
			}
			init := goInitialValue(v)
			if init != "" {
				fmt.Fprintf(&b, "\t\t%s: %s,\n", goCamel(v.Name), init)
			}
		}

		b.WriteString("\t\tGuards:  make(map[GuardID]func() bool),\n")
		b.WriteString("\t\tActions: make(map[ActionID]func() error),\n")
		b.WriteString("\t}\n}\n\n")

		// HandleMessage.
		fmt.Fprintf(&b, "func (m *%sMachine) HandleMessage(msg MsgType) (bool, error) {\n", prefix)
		b.WriteString("\tswitch {\n")
		for _, t := range a.FlattenedTransitions() {
			if t.On.Kind != TriggerRecv {
				continue
			}
			guard := ""
			if t.Guard != "" {
				guard = fmt.Sprintf(" && m.Guards[%sGuard%s] != nil && m.Guards[%sGuard%s]()",
					cPrefix, goCamel(string(t.Guard)), cPrefix, goCamel(string(t.Guard)))
			}
			fmt.Fprintf(&b, "\tcase m.State == %s%s && msg == %sMsg%s%s:\n",
				stateType, goCamel(string(t.From)),
				cPrefix, goCamel(string(t.On.Msg)), guard)

			writeGoTransitionBody(&b, t, stateType, cPrefix)
			b.WriteString("\t\treturn true, nil\n")
		}
		b.WriteString("\t}\n\treturn false, nil\n}\n\n")

		// Step — takes an EventID to disambiguate internal transitions.
		fmt.Fprintf(&b, "func (m *%sMachine) Step(event EventID) (bool, error) {\n", prefix)
		b.WriteString("\tswitch {\n")
		for _, t := range a.FlattenedTransitions() {
			if t.On.Kind != TriggerInternal {
				continue
			}
			guard := ""
			if t.Guard != "" {
				guard = fmt.Sprintf(" && m.Guards[%sGuard%s] != nil && m.Guards[%sGuard%s]()",
					cPrefix, goCamel(string(t.Guard)), cPrefix, goCamel(string(t.Guard)))
			}
			fmt.Fprintf(&b, "\tcase m.State == %s%s && event == %sEvent%s%s:\n",
				stateType, goCamel(string(t.From)),
				cPrefix, goCamel(t.On.Desc), guard)

			writeGoTransitionBody(&b, t, stateType, cPrefix)
			b.WriteString("\t\treturn true, nil\n")
		}
		b.WriteString("\t}\n\treturn false, nil\n}\n\n")

		// HandleEvent — unified entry point returning commands.
		// Maps recv messages to EventRecv* and internal events to Event*.
		fmt.Fprintf(&b, "func (m *%sMachine) HandleEvent(ev EventID) ([]CmdID, error) {\n", prefix)
		b.WriteString("\tswitch {\n")
		for _, t := range a.FlattenedTransitions() {
			guard := ""
			if t.Guard != "" {
				guard = fmt.Sprintf(" && m.Guards[%sGuard%s] != nil && m.Guards[%sGuard%s]()",
					cPrefix, goCamel(string(t.Guard)), cPrefix, goCamel(string(t.Guard)))
			}

			// Event ID: recv messages become EventRecv*, internal events become Event*.
			var eventConst string
			if t.On.Kind == TriggerRecv {
				eventConst = fmt.Sprintf("%sEventRecv%s", cPrefix, goCamel(string(t.On.Msg)))
			} else {
				eventConst = fmt.Sprintf("%sEvent%s", cPrefix, goCamel(t.On.Desc))
			}

			fmt.Fprintf(&b, "\tcase m.State == %s%s && ev == %s%s:\n",
				stateType, goCamel(string(t.From)), eventConst, guard)

			writeGoEventTransitionBody(&b, t, stateType, cPrefix)
		}
		b.WriteString("\t}\n\treturn nil, nil\n}\n\n")
	}

	_, err := io.WriteString(w, b.String())
	return err
}

// writeGoTransitionBody emits the action call, variable updates, state
// change, and change notifications for a single transition.
func writeGoTransitionBody(b *strings.Builder, t Transition, stateType, cPrefix string) {
	// Action.
	if t.Do != "" {
		fmt.Fprintf(b, "\t\tif fn := m.Actions[%sAction%s]; fn != nil {\n", cPrefix, goCamel(string(t.Do)))
		b.WriteString("\t\t\tif err := fn(); err != nil { return false, err }\n")
		b.WriteString("\t\t}\n")
	}

	// Variable updates — emit simple assignments directly.
	for _, u := range t.Updates {
		field := goCamel(u.Var)
		if lit, ok := goSimpleLiteral(u.Expr); ok {
			fmt.Fprintf(b, "\t\tm.%s = %s\n", field, lit)
			fmt.Fprintf(b, "\t\tif m.OnChange != nil { m.OnChange(%q) }\n", u.Var)
		} else if goExpr, ok := goSelfUpdate(u.Var, u.Expr); ok {
			fmt.Fprintf(b, "\t\tm.%s = %s\n", field, goExpr)
			fmt.Fprintf(b, "\t\tif m.OnChange != nil { m.OnChange(%q) }\n", u.Var)
		} else {
			// Complex expression — must be handled by the action callback.
			fmt.Fprintf(b, "\t\t// %s: %s (set by action)\n", u.Var, u.Expr)
		}
	}

	// State transition.
	fmt.Fprintf(b, "\t\tm.State = %s%s\n", stateType, goCamel(string(t.To)))
}

// writeGoEventTransitionBody emits the action call, variable updates,
// state change, and command list return for HandleEvent.
func writeGoEventTransitionBody(b *strings.Builder, t Transition, stateType, cPrefix string) {
	// Action.
	if t.Do != "" {
		fmt.Fprintf(b, "\t\tif fn := m.Actions[%sAction%s]; fn != nil {\n", cPrefix, goCamel(string(t.Do)))
		b.WriteString("\t\t\tif err := fn(); err != nil { return nil, err }\n")
		b.WriteString("\t\t}\n")
	}

	// Variable updates.
	for _, u := range t.Updates {
		field := goCamel(u.Var)
		if lit, ok := goSimpleLiteral(u.Expr); ok {
			fmt.Fprintf(b, "\t\tm.%s = %s\n", field, lit)
			fmt.Fprintf(b, "\t\tif m.OnChange != nil { m.OnChange(%q) }\n", u.Var)
		} else if goExpr, ok := goSelfUpdate(u.Var, u.Expr); ok {
			fmt.Fprintf(b, "\t\tm.%s = %s\n", field, goExpr)
			fmt.Fprintf(b, "\t\tif m.OnChange != nil { m.OnChange(%q) }\n", u.Var)
		} else {
			fmt.Fprintf(b, "\t\t// %s: %s (set by action)\n", u.Var, u.Expr)
		}
	}

	// State transition.
	fmt.Fprintf(b, "\t\tm.State = %s%s\n", stateType, goCamel(string(t.To)))

	// Return emitted commands.
	if len(t.Emits) > 0 {
		b.WriteString("\t\treturn []CmdID{")
		for i, cmd := range t.Emits {
			if i > 0 {
				b.WriteString(", ")
			}
			fmt.Fprintf(b, "%sCmd%s", cPrefix, goCamel(string(cmd)))
		}
		b.WriteString("}, nil\n")
	} else {
		b.WriteString("\t\treturn nil, nil\n")
	}
}

// wireInt extracts an integer from a YAML value that may be int or float64.
func wireInt(v any) int {
	switch n := v.(type) {
	case int:
		return n
	case float64:
		return int(n)
	default:
		return 0
	}
}

// goSimpleLiteral converts a TLA+ expression to a Go literal if it's
// simple enough (string, int, bool). Returns ("", false) for complex
// expressions that need action callbacks.
func goSimpleLiteral(expr string) (string, bool) {
	expr = strings.TrimSpace(expr)
	switch expr {
	case "TRUE":
		return "true", true
	case "FALSE":
		return "false", true
	}
	if strings.HasPrefix(expr, "\"") && strings.HasSuffix(expr, "\"") {
		return expr, true // already a Go string literal
	}
	// Simple integer.
	var n int
	if _, err := fmt.Sscanf(expr, "%d", &n); err == nil {
		return fmt.Sprintf("%d", n), true
	}
	return "", false
}

// goSelfUpdate checks if expr is a simple self-referential update like
// "var_name + 1" or "var_name - 1" and returns the Go equivalent
// (e.g., "m.VarName + 1"). Returns ("", false) for complex expressions.
func goSelfUpdate(varName, expr string) (string, bool) {
	expr = strings.TrimSpace(expr)
	// Match "var_name + N" or "var_name - N".
	for _, op := range []string{" + ", " - "} {
		if rest, ok := strings.CutPrefix(expr, varName+op); ok {
			rest = strings.TrimSpace(rest)
			var n int
			if _, err := fmt.Sscanf(rest, "%d", &n); err == nil {
				return fmt.Sprintf("m.%s%s%d", goCamel(varName), op, n), true
			}
		}
	}
	return "", false
}

// goInitialValue converts a VarDef's TLA+ initial value to a Go literal,
// or returns "" if it can't be simply expressed.
func goInitialValue(v VarDef) string {
	if lit, ok := goSimpleLiteral(v.Initial); ok {
		return lit
	}
	// Default zero values for types.
	switch v.Type {
	case VarInt:
		return "0"
	case VarBool:
		return "false"
	case VarSetString:
		return "" // zero value of frozen.Set[string] is empty set
	default:
		return `""`
	}
}

func goType(t VarType) string {
	switch t {
	case VarInt:
		return "int"
	case VarBool:
		return "bool"
	case VarSetString:
		return "frozen.Set[string]"
	default:
		return "string"
	}
}

func goConstPrefix(actor string) string {
	switch actor {
	case "ios":
		return "App"
	case "cli":
		return "CLI"
	default:
		if len(actor) == 0 {
			return ""
		}
		return strings.ToUpper(actor[:1]) + actor[1:]
	}
}

func goCamel(s string) string {
	var b strings.Builder
	upper := true
	for _, r := range s {
		if r == '_' || r == '-' || r == ' ' {
			upper = true
			continue
		}
		if upper {
			b.WriteRune(rune(strings.ToUpper(string(r))[0]))
			upper = false
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

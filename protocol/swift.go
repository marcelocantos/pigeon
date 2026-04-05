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
//   - EventID and CmdID enums
//   - Guard and action ID constants
//   - A static table literal (the Protocol struct as Swift data)
//   - Per-actor typed machine classes with handleEvent/handleMessage/step methods
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

	// EventID enum — declared events + internal events + recv_* events.
	events := collectAllEvents(p)
	if len(events) > 0 {
		b.WriteString("public enum EventID: String, Sendable {\n")
		for _, id := range events {
			fmt.Fprintf(&b, "    case %s = \"%s\"\n", swiftCase(id), id)
		}
		b.WriteString("}\n\n")
	}

	// CmdID enum.
	if len(p.Commands) > 0 {
		b.WriteString("public enum CmdID: String, Sendable {\n")
		for _, c := range p.Commands {
			fmt.Fprintf(&b, "    case %s = \"%s\"\n", swiftCase(string(c.ID)), c.ID)
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

	b.WriteString("}\n\n")

	// Per-actor typed state machines.
	for _, a := range p.Actors {
		typeName := swiftTypeName(a.Name)

		// Collect vars updated by this actor's transitions.
		actorVarSet := map[string]bool{}
		for _, t := range a.Transitions {
			for _, u := range t.Updates {
				actorVarSet[u.Var] = true
			}
		}

		// Machine class.
		fmt.Fprintf(&b, "/// %sMachine is the generated state machine for the %s actor.\n", typeName, a.Name)
		fmt.Fprintf(&b, "public final class %sMachine: @unchecked Sendable {\n", typeName)
		fmt.Fprintf(&b, "    public private(set) var state: %sState\n", typeName)

		// Typed variable fields owned by this actor.
		for _, v := range p.Vars {
			if !actorVarSet[v.Name] {
				continue
			}
			swiftT := swiftType(v.Type)
			comment := ""
			if v.Desc != "" {
				comment = " // " + v.Desc
			}
			fmt.Fprintf(&b, "    public var %s: %s%s\n", swiftCamelField(v.Name), swiftT, comment)
		}

		b.WriteString("\n")
		b.WriteString("    public var guards: [GuardID: () -> Bool] = [:]\n")
		b.WriteString("    public var actions: [ActionID: () throws -> Void] = [:]\n")
		b.WriteString("\n")

		// Constructor.
		fmt.Fprintf(&b, "    public init() {\n")
		fmt.Fprintf(&b, "        self.state = .%s\n", swiftCase(string(a.Initial)))
		for _, v := range p.Vars {
			if !actorVarSet[v.Name] {
				continue
			}
			init_ := swiftInitialValue(v)
			if init_ != "" {
				fmt.Fprintf(&b, "        self.%s = %s\n", swiftCamelField(v.Name), init_)
			}
		}
		b.WriteString("    }\n\n")

		// handleEvent — unified entry point returning commands.
		b.WriteString("    /// Handle any event (message receipt or internal). Returns emitted commands.\n")
		b.WriteString("    @discardableResult\n")
		b.WriteString("    public func handleEvent(_ ev: EventID) throws -> [CmdID] {\n")
		b.WriteString("        switch (state, ev) {\n")
		for _, t := range a.Transitions {
			// Compute EventID case name.
			var eventCase string
			if t.On.Kind == TriggerRecv {
				eventCase = swiftCase("recv_" + string(t.On.Msg))
			} else {
				eventCase = swiftCase(t.On.Desc)
			}

			if t.Guard != "" {
				fmt.Fprintf(&b, "        case (.%s, .%s) where guards[.%s]?() == true:\n",
					swiftCase(string(t.From)), eventCase, swiftCase(string(t.Guard)))
			} else {
				fmt.Fprintf(&b, "        case (.%s, .%s):\n",
					swiftCase(string(t.From)), eventCase)
			}

			writeSwiftEventTransitionBody(&b, t, p)
		}
		b.WriteString("        default:\n")
		b.WriteString("            return []\n")
		b.WriteString("        }\n")
		b.WriteString("    }\n\n")

		// handleMessage — backward compat.
		b.WriteString("    /// Process a received message. Returns the new state, or nil if rejected.\n")
		fmt.Fprintf(&b, "    @discardableResult\n")
		fmt.Fprintf(&b, "    public func handleMessage(_ msg: MessageType) throws -> %sState? {\n", typeName)
		b.WriteString("        switch (state, msg) {\n")
		for _, t := range a.Transitions {
			if t.On.Kind != TriggerRecv {
				continue
			}
			msgCase := swiftCase(string(t.On.Msg))
			if t.Guard != "" {
				fmt.Fprintf(&b, "        case (.%s, .%s) where guards[.%s]?() == true:\n",
					swiftCase(string(t.From)), msgCase, swiftCase(string(t.Guard)))
			} else {
				fmt.Fprintf(&b, "        case (.%s, .%s):\n",
					swiftCase(string(t.From)), msgCase)
			}
			writeSwiftTransitionBody(&b, t, p)
			b.WriteString("            return state\n")
		}
		b.WriteString("        default:\n")
		b.WriteString("            return nil\n")
		b.WriteString("        }\n")
		b.WriteString("    }\n\n")

		// step — backward compat.
		// Only handles states that have a single unambiguous auto-progression
		// (one internal event, possibly with guards). States that have multiple
		// distinct internal events are event-driven and require handleEvent.
		b.WriteString("    /// Attempt an internal transition. Returns the new state, or nil if none available.\n")
		fmt.Fprintf(&b, "    @discardableResult\n")
		fmt.Fprintf(&b, "    public func step() throws -> %sState? {\n", typeName)
		b.WriteString("        switch state {\n")

		// Group internal transitions by From state.
		// key: From state; value: set of distinct event descriptors in use.
		type internalT struct {
			guard GuardID
			t     Transition
		}
		fromMap := map[State][]internalT{}
		fromEventCount := map[State]map[string]bool{} // from → distinct event descs
		fromOrder := []State{}
		seenFrom := map[State]bool{}
		for _, t := range a.Transitions {
			if t.On.Kind != TriggerInternal {
				continue
			}
			if !seenFrom[t.From] {
				seenFrom[t.From] = true
				fromOrder = append(fromOrder, t.From)
				fromEventCount[t.From] = map[string]bool{}
			}
			fromMap[t.From] = append(fromMap[t.From], internalT{t.Guard, t})
			fromEventCount[t.From][t.On.Desc] = true
		}

		for _, from := range fromOrder {
			ts := fromMap[from]
			// Only emit a step case when all internal transitions from this state
			// share the same event (i.e., they're guard variants of one trigger).
			if len(fromEventCount[from]) != 1 {
				continue // multiple distinct events — use handleEvent instead
			}
			fmt.Fprintf(&b, "        case .%s:\n", swiftCase(string(from)))
			for _, it := range ts {
				if it.guard != "" {
					fmt.Fprintf(&b, "            if guards[.%s]?() == true {\n", swiftCase(string(it.guard)))
					writeSwiftTransitionBodyIndented(&b, it.t, p, "                ")
					b.WriteString("                return state\n")
					b.WriteString("            }\n")
				} else {
					writeSwiftTransitionBodyIndented(&b, it.t, p, "            ")
					b.WriteString("            return state\n")
				}
			}
			// If all transitions had guards, fall through to nil.
			allGuarded := true
			for _, it := range ts {
				if it.guard == "" {
					allGuarded = false
					break
				}
			}
			if allGuarded {
				b.WriteString("            return nil\n")
			}
		}

		b.WriteString("        default:\n")
		b.WriteString("            return nil\n")
		b.WriteString("        }\n")
		b.WriteString("    }\n")
		b.WriteString("}\n\n")
	}

	_, err := io.WriteString(w, b.String())
	return err
}

// writeSwiftEventTransitionBody emits action, var updates, state change,
// and command return for a handleEvent case.
func writeSwiftEventTransitionBody(b *strings.Builder, t Transition, p *Protocol) {
	if t.Do != "" {
		fmt.Fprintf(b, "            try actions[.%s]?()\n", swiftCase(string(t.Do)))
	}
	for _, u := range t.Updates {
		if lit, ok := swiftSimpleLiteral(u.Expr); ok {
			fmt.Fprintf(b, "            %s = %s\n", swiftCamelField(u.Var), lit)
		} else {
			fmt.Fprintf(b, "            // %s: %s (set by action)\n", u.Var, u.Expr)
		}
	}
	fmt.Fprintf(b, "            state = .%s\n", swiftCase(string(t.To)))
	if len(t.Emits) > 0 {
		b.WriteString("            return [")
		for i, cmd := range t.Emits {
			if i > 0 {
				b.WriteString(", ")
			}
			fmt.Fprintf(b, ".%s", swiftCase(string(cmd)))
		}
		b.WriteString("]\n")
	} else {
		b.WriteString("            return []\n")
	}
}

// writeSwiftTransitionBody emits action, var updates, and state change
// for handleMessage cases (no return — caller returns state).
func writeSwiftTransitionBody(b *strings.Builder, t Transition, p *Protocol) {
	writeSwiftTransitionBodyIndented(b, t, p, "            ")
}

func writeSwiftTransitionBodyIndented(b *strings.Builder, t Transition, p *Protocol, indent string) {
	if t.Do != "" {
		fmt.Fprintf(b, "%stry actions[.%s]?()\n", indent, swiftCase(string(t.Do)))
	}
	for _, u := range t.Updates {
		if lit, ok := swiftSimpleLiteral(u.Expr); ok {
			fmt.Fprintf(b, "%s%s = %s\n", indent, swiftCamelField(u.Var), lit)
		} else {
			fmt.Fprintf(b, "%s// %s: %s (set by action)\n", indent, u.Var, u.Expr)
		}
	}
	fmt.Fprintf(b, "%sstate = .%s\n", indent, swiftCase(string(t.To)))
}

// collectAllEvents returns a deduplicated, ordered list of all event IDs:
// declared events + internal transition descriptors + recv_* for messages.
func collectAllEvents(p *Protocol) []string {
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
	// Internal event descriptors.
	for _, a := range p.Actors {
		for _, t := range a.Transitions {
			if t.On.Kind == TriggerInternal {
				add(t.On.Desc)
			}
		}
	}
	// recv_* events for message receipts.
	for _, a := range p.Actors {
		for _, t := range a.Transitions {
			if t.On.Kind == TriggerRecv {
				add("recv_" + string(t.On.Msg))
			}
		}
	}
	return result
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

// swiftSimpleLiteral converts a TLA+ expression to a Swift literal
// for simple cases (bool, int, string). Returns ("", false) otherwise.
func swiftSimpleLiteral(expr string) (string, bool) {
	expr = strings.TrimSpace(expr)
	switch expr {
	case "TRUE":
		return "true", true
	case "FALSE":
		return "false", true
	}
	if strings.HasPrefix(expr, "\"") && strings.HasSuffix(expr, "\"") {
		return expr, true
	}
	var n int
	if _, err := fmt.Sscanf(expr, "%d", &n); err == nil {
		return fmt.Sprintf("%d", n), true
	}
	return "", false
}

// swiftInitialValue returns a Swift literal for a VarDef's initial value.
func swiftInitialValue(v VarDef) string {
	if lit, ok := swiftSimpleLiteral(v.Initial); ok {
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

// swiftType converts a VarType to a Swift type name.
func swiftType(t VarType) string {
	switch t {
	case VarInt:
		return "Int"
	case VarBool:
		return "Bool"
	default:
		return "String"
	}
}

// swiftCamelField converts a snake_case variable name to lowerCamelCase
// for Swift property names (same as swiftCase but cleaner for fields).
func swiftCamelField(s string) string {
	return swiftCase(s)
}

func swiftTypeName(name string) string {
	if len(name) == 0 {
		return name
	}
	return strings.ToUpper(name[:1]) + name[1:]
}

func swiftCase(s string) string {
	// Convert snake_case, PascalCase, or camelCase to lowerCamelCase.
	if len(s) == 0 {
		return s
	}
	runes := []rune(s)
	var result []rune
	prevUpper := false
	atBoundary := false // true when last skipped char was non-alphanumeric
	for i, r := range runes {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			prevUpper = false
			atBoundary = true
			continue
		}
		if len(result) == 0 {
			// First character is always lowercase.
			result = append(result, unicode.ToLower(r))
			prevUpper = unicode.IsUpper(r)
			atBoundary = false
			continue
		}
		// After a word boundary (underscore, hyphen, space, etc.):
		// capitalise the first letter of the new word.
		if atBoundary {
			result = append(result, unicode.ToUpper(r))
			prevUpper = unicode.IsUpper(r)
			atBoundary = false
			continue
		}
		// PascalCase boundary: transition from lower to upper.
		_ = i
		if unicode.IsUpper(r) && !prevUpper {
			result = append(result, r)
		} else if prevUpper && unicode.IsLower(r) {
			// e.g. "LANActive" — capitalise the penultimate letter.
			if len(result) > 1 {
				last := result[len(result)-1]
				result[len(result)-1] = unicode.ToUpper(last)
			}
			result = append(result, r)
		} else {
			result = append(result, r)
		}
		prevUpper = unicode.IsUpper(r)
		atBoundary = false
	}
	return string(result)
}

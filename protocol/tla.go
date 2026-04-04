// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package protocol

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

// ExportTLA writes a pure TLA+ spec (no PlusCal) for the full protocol.
func (p *Protocol) ExportTLA(w io.Writer) error {
	return p.ExportTLAPhase(w, "")
}

// ExportTLAPhase writes a pure TLA+ spec for a specific phase, or the
// full protocol if phaseName is empty. Each YAML transition becomes a
// named TLA+ action. No PlusCal, no processes, no program counters.
func (p *Protocol) ExportTLAPhase(w io.Writer, phaseName string) error {
	var phase *Phase
	if phaseName != "" {
		for i := range p.Phases {
			if p.Phases[i].Name == phaseName {
				phase = &p.Phases[i]
				break
			}
		}
		if phase == nil {
			return fmt.Errorf("phase %q not found", phaseName)
		}
	}

	phaseStates := map[State]bool{}
	phaseVars := map[string]bool{}
	if phase != nil {
		for _, s := range phase.States {
			phaseStates[s] = true
		}
		for _, v := range phase.Vars {
			phaseVars[v] = true
		}
	}

	var b strings.Builder

	moduleName := sanitiseTLA(p.Name)
	if phase != nil {
		moduleName += "_" + sanitiseTLA(phase.Name)
	}

	fmt.Fprintf(&b, "---- MODULE %s ----\n", moduleName)
	b.WriteString("\\* Auto-generated from protocol YAML. Do not edit.\n")
	if phase != nil {
		fmt.Fprintf(&b, "\\* Phase: %s\n", phase.Name)
	}
	b.WriteString("\nEXTENDS Integers, Sequences, FiniteSets, TLC\n\n")

	// --- State constants ---
	for _, a := range p.Actors {
		fmt.Fprintf(&b, "\\* States for %s\n", a.Name)
		for _, s := range collectStates(a) {
			if len(phaseStates) > 0 && !phaseStates[s] && !isNeighbour(a, s, phaseStates) {
				continue
			}
			fmt.Fprintf(&b, "%s_%s == \"%s_%s\"\n",
				sanitiseTLA(a.Name), sanitiseTLA(string(s)), a.Name, s)
		}
		b.WriteString("\n")
	}

	// --- Message constants ---
	if len(p.Messages) > 0 {
		b.WriteString("\\* Message types\n")
		for _, m := range p.Messages {
			fmt.Fprintf(&b, "MSG_%s == \"%s\"\n", sanitiseTLA(string(m.Type)), m.Type)
		}
		b.WriteString("\n")
	}

	// --- Operators ---
	for _, op := range p.Operators {
		if op.Desc != "" {
			fmt.Fprintf(&b, "\\* %s\n", op.Desc)
		}
		if op.Params != "" {
			fmt.Fprintf(&b, "%s(%s) == %s\n", sanitiseTLA(op.Name), op.Params, op.Expr)
		} else {
			fmt.Fprintf(&b, "%s == %s\n", sanitiseTLA(op.Name), op.Expr)
		}
	}
	if len(p.Operators) > 0 {
		b.WriteString("\n")
	}

	// --- CONSTANTS ---
	b.WriteString("CONSTANTS MaxChanLen\n\n")

	// --- Variables ---
	b.WriteString("VARIABLES\n")
	var allVarNames []string

	// Actor state variables.
	for _, a := range p.Actors {
		name := sanitiseTLA(a.Name) + "_state"
		allVarNames = append(allVarNames, name)
		fmt.Fprintf(&b, "    %s,\n", name)
	}

	// Channels.
	channels := channelPairs(p)
	for _, ch := range channels {
		name := channelName(ch.from, ch.to)
		allVarNames = append(allVarNames, name)
		fmt.Fprintf(&b, "    %s,\n", name)
	}

	// Protocol variables (phase-filtered).
	for _, v := range p.Vars {
		if phase != nil && !phaseVars[v.Name] && v.Name != "recv_msg" {
			continue
		}
		allVarNames = append(allVarNames, sanitiseTLA(v.Name))
		fmt.Fprintf(&b, "    %s,\n", sanitiseTLA(v.Name))
	}

	// Remove trailing comma.
	s := b.String()
	if idx := strings.LastIndex(s, ",\n"); idx >= 0 {
		b.Reset()
		b.WriteString(s[:idx])
		b.WriteString("\n")
	}
	b.WriteString("\n")

	// vars tuple.
	fmt.Fprintf(&b, "vars == <<%s>>\n\n", strings.Join(allVarNames, ", "))

	// Helper.
	b.WriteString("CanSend(ch) == Len(ch) < MaxChanLen\n\n")

	// --- Init ---
	b.WriteString("Init ==\n")
	for _, a := range p.Actors {
		init := initialForPhase(a, phase)
		fmt.Fprintf(&b, "    /\\ %s_state = %s_%s\n",
			sanitiseTLA(a.Name), sanitiseTLA(a.Name), sanitiseTLA(string(init)))
	}
	for _, ch := range channels {
		fmt.Fprintf(&b, "    /\\ %s = <<>>\n", channelName(ch.from, ch.to))
	}
	for _, v := range p.Vars {
		if phase != nil && !phaseVars[v.Name] && v.Name != "recv_msg" {
			continue
		}
		fmt.Fprintf(&b, "    /\\ %s = %s\n", sanitiseTLA(v.Name), v.Initial)
	}
	b.WriteString("\n")

	// --- Actions ---
	// Each YAML transition becomes a named TLA+ action.
	var actionNames []string

	for _, a := range p.Actors {
		fmt.Fprintf(&b, "\\* ================================================================\n")
		fmt.Fprintf(&b, "\\* %s actions\n", a.Name)
		fmt.Fprintf(&b, "\\* ================================================================\n\n")

		for _, t := range a.Transitions {
			// Phase filter.
			if len(phaseStates) > 0 && !phaseStates[t.From] && !phaseStates[t.To] {
				continue
			}

			actionName := fmt.Sprintf("%s_%s_to_%s",
				sanitiseTLA(a.Name),
				sanitiseTLA(string(t.From)),
				sanitiseTLA(string(t.To)))

			// Disambiguate if multiple transitions share from→to.
			if t.On.Kind == TriggerRecv {
				actionName += "_on_" + sanitiseTLA(string(t.On.Msg))
			} else if t.On.Desc != "" {
				actionName += "_on_" + sanitiseTLA(t.On.Desc)
			}
			if t.Guard != "" {
				actionName += "_" + sanitiseTLA(string(t.Guard))
			}

			actionNames = append(actionNames, actionName)

			// Comment.
			fmt.Fprintf(&b, "\\* %s: %s -> %s", a.Name, t.From, t.To)
			if t.On.Kind == TriggerRecv {
				fmt.Fprintf(&b, " on recv %s", t.On.Msg)
			} else if t.On.Desc != "" {
				fmt.Fprintf(&b, " (%s)", t.On.Desc)
			}
			if t.Guard != "" {
				fmt.Fprintf(&b, " [%s]", t.Guard)
			}
			b.WriteString("\n")

			fmt.Fprintf(&b, "%s ==\n", actionName)

			// State guard.
			fmt.Fprintf(&b, "    /\\ %s_state = %s_%s\n",
				sanitiseTLA(a.Name),
				sanitiseTLA(a.Name), sanitiseTLA(string(t.From)))

			// Message guard.
			if t.On.Kind == TriggerRecv {
				fromActor := msgSender(p, t.On.Msg)
				ch := channelName(fromActor, a.Name)
				fmt.Fprintf(&b, "    /\\ Len(%s) > 0\n", ch)
				fmt.Fprintf(&b, "    /\\ Head(%s).type = MSG_%s\n", ch, sanitiseTLA(string(t.On.Msg)))
			}

			// Guard expression.
			if t.Guard != "" {
				expr := guardExpr(p, t.Guard)
				if t.On.Kind == TriggerRecv {
					fromActor := msgSender(p, t.On.Msg)
					ch := channelName(fromActor, a.Name)
					expr = strings.ReplaceAll(expr, "recv_msg", "Head("+ch+")")
				}
				fmt.Fprintf(&b, "    /\\ %s\n", expr)
			}

			// --- Primed assignments ---

			// Consume message.
			if t.On.Kind == TriggerRecv {
				fromActor := msgSender(p, t.On.Msg)
				ch := channelName(fromActor, a.Name)
				fmt.Fprintf(&b, "    /\\ %s' = Tail(%s)\n", ch, ch)
			}

			// Sends.
			for _, s := range t.Sends {
				ch := channelName(a.Name, s.To)
				fmt.Fprintf(&b, "    /\\ %s' = Append(%s, ", ch, ch)
				writeRecord(&b, s.Msg, s.Fields)
				b.WriteString(")\n")
			}

			// State update.
			fmt.Fprintf(&b, "    /\\ %s_state' = %s_%s\n",
				sanitiseTLA(a.Name),
				sanitiseTLA(a.Name), sanitiseTLA(string(t.To)))

			// Variable updates.
			for _, u := range t.Updates {
				fmt.Fprintf(&b, "    /\\ %s' = %s\n", sanitiseTLA(u.Var), u.Expr)
			}

			// UNCHANGED for everything not modified.
			modified := map[string]bool{
				sanitiseTLA(a.Name) + "_state": true,
			}
			if t.On.Kind == TriggerRecv {
				fromActor := msgSender(p, t.On.Msg)
				modified[channelName(fromActor, a.Name)] = true
			}
			for _, s := range t.Sends {
				modified[channelName(a.Name, s.To)] = true
			}
			for _, u := range t.Updates {
				modified[sanitiseTLA(u.Var)] = true
			}

			var unchanged []string
			for _, v := range allVarNames {
				if !modified[v] {
					unchanged = append(unchanged, v)
				}
			}
			if len(unchanged) > 0 {
				fmt.Fprintf(&b, "    /\\ UNCHANGED <<%s>>\n", strings.Join(unchanged, ", "))
			}

			b.WriteString("\n")
		}
	}

	// --- Next ---
	b.WriteString("Next ==\n")
	for i, name := range actionNames {
		if i == 0 {
			fmt.Fprintf(&b, "    \\/ %s\n", name)
		} else {
			fmt.Fprintf(&b, "    \\/ %s\n", name)
		}
	}
	b.WriteString("\n")

	// --- Spec ---
	b.WriteString("Spec == Init /\\ [][Next]_vars /\\ WF_vars(Next)\n\n")

	// --- Properties ---
	b.WriteString("\\* ================================================================\n")
	b.WriteString("\\* Invariants and properties\n")
	b.WriteString("\\* ================================================================\n\n")
	for _, prop := range p.Properties {
		if phase != nil && !propertyRelevant(prop, phaseVars, p.Vars) {
			continue
		}
		if prop.Desc != "" {
			fmt.Fprintf(&b, "\\* %s\n", prop.Desc)
		}
		switch prop.Kind {
		case Invariant:
			fmt.Fprintf(&b, "%s == %s\n", sanitiseTLA(prop.Name), prop.Expr)
		case Liveness:
			fmt.Fprintf(&b, "%s == <>(%s)\n", sanitiseTLA(prop.Name), prop.Expr)
		case LeadsTo:
			fmt.Fprintf(&b, "%s == (%s) ~> (%s)\n", sanitiseTLA(prop.Name), prop.FromExpr, prop.ToExpr)
		}
	}
	b.WriteString("\n====\n")

	_, err := io.WriteString(w, b.String())
	return err
}

func isNeighbour(a Actor, s State, phaseStates map[State]bool) bool {
	for _, t := range a.Transitions {
		if t.From == s && phaseStates[t.To] {
			return true
		}
		if t.To == s && phaseStates[t.From] {
			return true
		}
	}
	return false
}

func initialForPhase(a Actor, phase *Phase) State {
	if phase == nil {
		return a.Initial
	}
	phaseStates := map[State]bool{}
	for _, s := range phase.States {
		phaseStates[s] = true
	}
	if phaseStates[a.Initial] {
		return a.Initial
	}
	for _, t := range a.Transitions {
		if !phaseStates[t.From] && phaseStates[t.To] {
			return t.To
		}
	}
	return a.Initial
}

func writeRecord(b *strings.Builder, msg MsgType, fields map[string]string) {
	b.WriteString("[type |-> MSG_")
	b.WriteString(sanitiseTLA(string(msg)))
	if len(fields) > 0 {
		keys := make([]string, 0, len(fields))
		for k := range fields {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Fprintf(b, ", %s |-> %s", k, fields[k])
		}
	}
	b.WriteString("]")
}

func guardExpr(p *Protocol, id GuardID) string {
	for _, g := range p.Guards {
		if g.ID == id {
			return g.Expr
		}
	}
	return string(id)
}

func propertyRelevant(prop Property, phaseVars map[string]bool, allVars []VarDef) bool {
	if len(phaseVars) == 0 {
		return true
	}
	expr := prop.Expr + prop.FromExpr + prop.ToExpr
	for _, v := range allVars {
		if !phaseVars[v.Name] && v.Name != "recv_msg" && strings.Contains(expr, v.Name) {
			return false
		}
	}
	if strings.Contains(expr, "adversary_knowledge") || strings.Contains(expr, "adversary_keys") {
		if !phaseVars["adversary_keys"] {
			return false
		}
	}
	return true
}

// Helpers.

type channelPair struct{ from, to string }

func channelPairs(p *Protocol) []channelPair {
	seen := map[string]bool{}
	var pairs []channelPair
	for _, m := range p.Messages {
		key := m.From + "_" + m.To
		if !seen[key] {
			seen[key] = true
			pairs = append(pairs, channelPair{from: m.From, to: m.To})
		}
	}
	return pairs
}

func channelName(from, to string) string {
	return "chan_" + from + "_" + to
}

func msgSender(p *Protocol, msg MsgType) string {
	for _, m := range p.Messages {
		if m.Type == msg {
			return m.From
		}
	}
	return "unknown"
}

func sanitiseTLA(s string) string {
	r := strings.NewReplacer(" ", "_", "-", "_", ".", "_")
	return r.Replace(s)
}

// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package protocol

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

// ExportTLA writes a pure TLA+ spec for the full protocol.
func (p *Protocol) ExportTLA(w io.Writer) error {
	return p.ExportTLAPhase(w, "")
}

// ExportTLAPhase writes a pure TLA+ spec for a specific phase, or the
// full protocol if phaseName is empty. The generated spec is optimised:
// only phase-relevant variables, messages, and channels are included.
// recv_msg is eliminated by inlining Head(channel) in guards.
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

	// Collect transitions relevant to this phase.
	type actorTransition struct {
		actor      Actor
		transition Transition
	}
	var phaseTransitions []actorTransition
	for _, a := range p.Actors {
		for _, t := range a.Transitions {
			if len(phaseStates) > 0 {
				if !phaseStates[t.From] || !phaseStates[t.To] {
					// Only include transitions that stay within the phase.
					continue
				}
			}
			phaseTransitions = append(phaseTransitions, actorTransition{a, t})
		}
	}

	// Collect messages actually used in phase transitions.
	usedMsgs := map[MsgType]bool{}
	for _, at := range phaseTransitions {
		if at.transition.On.Kind == TriggerRecv {
			usedMsgs[at.transition.On.Msg] = true
		}
		for _, s := range at.transition.Sends {
			usedMsgs[s.Msg] = true
		}
	}

	// Collect channel routes actually used.
	usedChannels := map[string]channelPair{}
	for _, at := range phaseTransitions {
		t := at.transition
		a := at.actor
		if t.On.Kind == TriggerRecv {
			from := msgSender(p, t.On.Msg)
			key := from + "_" + a.Name
			usedChannels[key] = channelPair{from, a.Name}
		}
		for _, s := range t.Sends {
			key := a.Name + "_" + s.To
			usedChannels[key] = channelPair{a.Name, s.To}
		}
	}
	var channels []channelPair
	for _, ch := range usedChannels {
		channels = append(channels, ch)
	}
	sort.Slice(channels, func(i, j int) bool {
		return channels[i].from+channels[i].to < channels[j].from+channels[j].to
	})

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

	// --- State constants (phase-scoped) ---
	for _, a := range p.Actors {
		states := collectStates(a)
		var emitted bool
		for _, s := range states {
			if len(phaseStates) > 0 && !phaseStates[s] && !isNeighbour(a, s, phaseStates) {
				continue
			}
			if !emitted {
				fmt.Fprintf(&b, "\\* States for %s\n", a.Name)
				emitted = true
			}
			fmt.Fprintf(&b, "%s_%s == \"%s_%s\"\n",
				sanitiseTLA(a.Name), sanitiseTLA(string(s)), a.Name, s)
		}
		if emitted {
			b.WriteString("\n")
		}
	}

	// --- Message constants (only used messages) ---
	var emittedMsgs bool
	for _, m := range p.Messages {
		if len(usedMsgs) > 0 && !usedMsgs[m.Type] {
			continue
		}
		if !emittedMsgs {
			b.WriteString("\\* Message types\n")
			emittedMsgs = true
		}
		fmt.Fprintf(&b, "MSG_%s == \"%s\"\n", sanitiseTLA(string(m.Type)), m.Type)
	}
	if emittedMsgs {
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

	// Detect which vars are actually modified by phase transitions.
	modifiedVars := map[string]bool{}
	for _, at := range phaseTransitions {
		for _, u := range at.transition.Updates {
			modifiedVars[u.Var] = true
		}
	}

	// --- CONSTANTS ---
	// Variables that exist in the phase but are never updated become constants.
	b.WriteString("CONSTANTS MaxChanLen")
	var constVars []string
	for _, v := range p.Vars {
		if v.Name == "recv_msg" {
			continue
		}
		if phase != nil && !phaseVars[v.Name] {
			continue
		}
		if !modifiedVars[v.Name] {
			constVars = append(constVars, sanitiseTLA(v.Name))
			fmt.Fprintf(&b, ", %s", sanitiseTLA(v.Name))
		}
	}
	b.WriteString("\n\n")

	constSet := map[string]bool{}
	for _, c := range constVars {
		constSet[c] = true
	}

	// Detect actors with transitions that change state within this phase.
	// Actors whose state is constant become TLA+ CONSTANTS.
	actorChangesState := map[string]bool{}
	for _, at := range phaseTransitions {
		if at.transition.From != at.transition.To {
			actorChangesState[at.actor.Name] = true
		}
	}

	// Add constant-state actors to the CONSTANTS line.
	for _, a := range p.Actors {
		hasTransitions := false
		for _, at := range phaseTransitions {
			if at.actor.Name == a.Name {
				hasTransitions = true
				break
			}
		}
		if hasTransitions && !actorChangesState[a.Name] {
			name := sanitiseTLA(a.Name) + "_state"
			constSet[name] = true
			fmt.Fprintf(&b, ", %s", name)
		}
	}
	b.WriteString("\n\n")

	// --- Variables (no recv_msg — inlined as Head(channel)) ---
	b.WriteString("VARIABLES\n")
	var allVarNames []string

	for _, a := range p.Actors {
		hasTransitions := false
		for _, at := range phaseTransitions {
			if at.actor.Name == a.Name {
				hasTransitions = true
				break
			}
		}
		if !hasTransitions || !actorChangesState[a.Name] {
			continue
		}
		name := sanitiseTLA(a.Name) + "_state"
		allVarNames = append(allVarNames, name)
		fmt.Fprintf(&b, "    %s,\n", name)
	}

	for _, ch := range channels {
		name := channelName(ch.from, ch.to)
		allVarNames = append(allVarNames, name)
		fmt.Fprintf(&b, "    %s,\n", name)
	}

	for _, v := range p.Vars {
		if v.Name == "recv_msg" {
			continue // eliminated — inlined as Head(channel)
		}
		if phase != nil && !phaseVars[v.Name] {
			continue
		}
		if constSet[sanitiseTLA(v.Name)] {
			continue // promoted to CONSTANT
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

	fmt.Fprintf(&b, "vars == <<%s>>\n\n", strings.Join(allVarNames, ", "))
	b.WriteString("CanSend(ch) == Len(ch) < MaxChanLen\n\n")

	// --- Init ---
	b.WriteString("Init ==\n")
	emittedActors := map[string]bool{}
	for _, at := range phaseTransitions {
		if emittedActors[at.actor.Name] {
			continue
		}
		emittedActors[at.actor.Name] = true
		init := initialForPhase(at.actor, phase)
		fmt.Fprintf(&b, "    /\\ %s_state = %s_%s\n",
			sanitiseTLA(at.actor.Name),
			sanitiseTLA(at.actor.Name), sanitiseTLA(string(init)))
	}
	for _, ch := range channels {
		fmt.Fprintf(&b, "    /\\ %s = <<>>\n", channelName(ch.from, ch.to))
	}
	for _, v := range p.Vars {
		if v.Name == "recv_msg" {
			continue
		}
		if phase != nil && !phaseVars[v.Name] {
			continue
		}
		if constSet[sanitiseTLA(v.Name)] {
			continue
		}
		fmt.Fprintf(&b, "    /\\ %s = %s\n", sanitiseTLA(v.Name), v.Initial)
	}
	b.WriteString("\n")

	// --- Actions ---
	var actionNames []string

	for _, a := range p.Actors {
		var actorActions []string
		for _, t := range a.Transitions {
			if len(phaseStates) > 0 && (!phaseStates[t.From] || !phaseStates[t.To]) {
				continue
			}

			actionName := makeActionName(a.Name, t)
			actionNames = append(actionNames, actionName)
			actorActions = append(actorActions, actionName)

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

			// Message guard — inline Head(channel) instead of recv_msg.
			var recvChan string
			if t.On.Kind == TriggerRecv {
				fromActor := msgSender(p, t.On.Msg)
				recvChan = channelName(fromActor, a.Name)
				fmt.Fprintf(&b, "    /\\ Len(%s) > 0\n", recvChan)
				fmt.Fprintf(&b, "    /\\ Head(%s).type = MSG_%s\n", recvChan, sanitiseTLA(string(t.On.Msg)))
			}

			// Guard expression — substitute recv_msg with Head(channel).
			if t.Guard != "" {
				expr := guardExpr(p, t.Guard)
				if recvChan != "" {
					expr = strings.ReplaceAll(expr, "recv_msg", "Head("+recvChan+")")
				}
				fmt.Fprintf(&b, "    /\\ %s\n", expr)
			}

			// --- Primed assignments ---
			modified := map[string]bool{
				sanitiseTLA(a.Name) + "_state": true,
			}

			// Consume message (Tail the channel).
			if recvChan != "" {
				fmt.Fprintf(&b, "    /\\ %s' = Tail(%s)\n", recvChan, recvChan)
				modified[recvChan] = true
			}

			// Sends.
			for _, s := range t.Sends {
				ch := channelName(a.Name, s.To)
				fmt.Fprintf(&b, "    /\\ %s' = Append(%s, ", ch, ch)
				writeRecord(&b, s.Msg, s.Fields)
				b.WriteString(")\n")
				modified[ch] = true
			}

			// State update.
			fmt.Fprintf(&b, "    /\\ %s_state' = %s_%s\n",
				sanitiseTLA(a.Name),
				sanitiseTLA(a.Name), sanitiseTLA(string(t.To)))

			// Variable updates.
			for _, u := range t.Updates {
				fmt.Fprintf(&b, "    /\\ %s' = %s\n", sanitiseTLA(u.Var), u.Expr)
				modified[sanitiseTLA(u.Var)] = true
			}

			// UNCHANGED.
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

		if len(actorActions) > 0 {
			b.WriteString("\n")
		}
	}

	// --- Next ---
	b.WriteString("Next ==\n")
	for _, name := range actionNames {
		fmt.Fprintf(&b, "    \\/ %s\n", name)
	}
	b.WriteString("\nSpec == Init /\\ [][Next]_vars /\\ WF_vars(Next)\n\n")

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

func makeActionName(actorName string, t Transition) string {
	name := fmt.Sprintf("%s_%s_to_%s",
		sanitiseTLA(actorName),
		sanitiseTLA(string(t.From)),
		sanitiseTLA(string(t.To)))
	if t.On.Kind == TriggerRecv {
		name += "_on_" + sanitiseTLA(string(t.On.Msg))
	} else if t.On.Desc != "" {
		name += "_" + sanitiseTLA(t.On.Desc)
	}
	if t.Guard != "" {
		name += "_" + sanitiseTLA(string(t.Guard))
	}
	return name
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

type channelPair struct{ from, to string }

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

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
		for _, t := range a.FlattenedTransitions() {
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

	// Collect events and commands actually used in phase transitions.
	usedEvents := map[EventID]bool{}
	usedCmds := map[CmdID]bool{}
	for _, at := range phaseTransitions {
		// The trigger event: either an internal event ID or a recv message event.
		if at.transition.On.Kind == TriggerInternal && at.transition.On.Desc != "" {
			usedEvents[EventID(at.transition.On.Desc)] = true
		}
		if at.transition.On.Kind == TriggerRecv {
			usedEvents[EventID("recv_"+string(at.transition.On.Msg))] = true
		}
		for _, cmd := range at.transition.Emits {
			usedCmds[cmd] = true
		}
	}

	// Channel elimination: instead of channel sequences, each receivable
	// message type becomes a struct variable. Senders write directly to
	// the struct; receivers guard on it and clear it after processing.
	// This eliminates sequences from the state space entirely.

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

	// --- Event constants (only events used in phase transitions) ---
	if len(usedEvents) > 0 {
		var sortedEvts []string
		for e := range usedEvents {
			sortedEvts = append(sortedEvts, string(e))
		}
		sort.Strings(sortedEvts)
		b.WriteString("\\* Event types\n")
		for _, e := range sortedEvts {
			fmt.Fprintf(&b, "EVT_%s == \"%s\"\n", sanitiseTLA(e), e)
		}
		b.WriteString("\n")
	}

	// --- Command constants (only commands emitted in phase transitions) ---
	if len(usedCmds) > 0 {
		var sortedCmds []string
		for c := range usedCmds {
			sortedCmds = append(sortedCmds, string(c))
		}
		sort.Strings(sortedCmds)
		b.WriteString("\\* Command types\n")
		for _, c := range sortedCmds {
			fmt.Fprintf(&b, "CMD_%s == \"%s\"\n", sanitiseTLA(c), c)
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

	// Detect which vars are actually modified by phase transitions.
	modifiedVars := map[string]bool{}
	for _, at := range phaseTransitions {
		for _, u := range at.transition.Updates {
			modifiedVars[u.Var] = true
		}
	}

	// --- CONSTANTS ---
	// Variables that exist in the phase but are never updated become constants.
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

	// Add constant-state actors.
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
			constVars = append(constVars, name)
		}
	}

	// Emit CONSTANTS line if any.
	if len(constVars) > 0 {
		fmt.Fprintf(&b, "CONSTANTS %s\n\n", strings.Join(constVars, ", "))
	}

	// Collect received message types per actor → received_<msg> variables.
	// Each receivable message becomes a struct variable that stores the
	// whole message on receipt.
	type recvInfo struct {
		actorName string
		msgType   MsgType
		varName   string
	}
	var recvVars []recvInfo
	recvVarSet := map[string]bool{}
	for _, at := range phaseTransitions {
		if at.transition.On.Kind == TriggerRecv {
			varName := "received_" + sanitiseTLA(string(at.transition.On.Msg))
			if !recvVarSet[varName] {
				recvVarSet[varName] = true
				recvVars = append(recvVars, recvInfo{
					actorName: at.actor.Name,
					msgType:   at.transition.On.Msg,
					varName:   varName,
				})
			}
		}
	}

	// --- Variables ---
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

	// No channel variables — channels eliminated via received_<msg> structs.

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

	// Received message struct variables.
	for _, rv := range recvVars {
		allVarNames = append(allVarNames, rv.varName)
		fmt.Fprintf(&b, "    %s,\n", rv.varName)
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
	// No channel init — channels eliminated.
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
	// Init received message vars to empty record.
	for _, rv := range recvVars {
		fmt.Fprintf(&b, "    /\\ %s = [type |-> \"none\"]\n", rv.varName)
	}
	b.WriteString("\n")

	// --- Actions ---
	var actionNames []string

	for _, a := range p.Actors {
		var actorActions []string
		for _, t := range a.FlattenedTransitions() {
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

			// Message guard: check that the received_<msg> struct has
			// been written (type /= "none") by the sender.
			var recvVar string
			if t.On.Kind == TriggerRecv {
				recvVar = "received_" + sanitiseTLA(string(t.On.Msg))
				fmt.Fprintf(&b, "    /\\ %s.type = MSG_%s\n", recvVar, sanitiseTLA(string(t.On.Msg)))
			}

			// Guard expression — substitute recv_msg with the received struct.
			if t.Guard != "" {
				expr := guardExpr(p, t.Guard)
				if recvVar != "" {
					expr = strings.ReplaceAll(expr, "recv_msg", recvVar)
				}
				fmt.Fprintf(&b, "    /\\ %s\n", expr)
			}

			// --- Primed assignments ---
			modified := map[string]bool{
				sanitiseTLA(a.Name) + "_state": true,
			}

			// Consume received message: clear the struct.
			if recvVar != "" {
				fmt.Fprintf(&b, "    /\\ %s' = [type |-> \"none\"]\n", recvVar)
				modified[recvVar] = true
			}

			// Sends: write directly to the received_<msg> struct on the
			// receiver side. No channel — the struct IS the communication.
			for _, s := range t.Sends {
				targetRecvVar := "received_" + sanitiseTLA(string(s.Msg))
				var sb strings.Builder
				writeRecord(&sb, s.Msg, s.Fields)
				fmt.Fprintf(&b, "    /\\ %s' = %s\n", targetRecvVar, sb.String())
				modified[targetRecvVar] = true
			}

			// State update.
			fmt.Fprintf(&b, "    /\\ %s_state' = %s_%s\n",
				sanitiseTLA(a.Name),
				sanitiseTLA(a.Name), sanitiseTLA(string(t.To)))

			// Variable updates — substitute recv_msg with received struct.
			for _, u := range t.Updates {
				expr := u.Expr
				if recvVar != "" {
					expr = strings.ReplaceAll(expr, "recv_msg", recvVar)
				}
				fmt.Fprintf(&b, "    /\\ %s' = %s\n", sanitiseTLA(u.Var), expr)
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

			// Command set operator — declares what commands this
			// transition emits. Used for verification properties.
			if len(t.Emits) > 0 {
				fmt.Fprintf(&b, "Cmds_%s == {", actionName)
				for i, cmd := range t.Emits {
					if i > 0 {
						b.WriteString(", ")
					}
					fmt.Fprintf(&b, "CMD_%s", sanitiseTLA(string(cmd)))
				}
				b.WriteString("}\n\n")
			}
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
	// --- Command-consistency invariants ---
	// These verify that state changes implied by commands are
	// reflected in the transition's variable updates. Generated
	// from the relationship between commands and state variables.
	hasCommands := false
	for _, a := range p.Actors {
		for _, t := range a.FlattenedTransitions() {
			if len(phaseStates) > 0 && (!phaseStates[t.From] || !phaseStates[t.To]) {
				continue
			}
			if len(t.Emits) > 0 {
				hasCommands = true
				break
			}
		}
		if hasCommands {
			break
		}
	}
	if hasCommands {
		b.WriteString("\n\\* ================================================================\n")
		b.WriteString("\\* Command-consistency: state after transition matches emitted commands\n")
		b.WriteString("\\* These are verified by construction (the same YAML defines both\n")
		b.WriteString("\\* the variable updates and the command list), but documenting\n")
		b.WriteString("\\* them as TLA+ operators makes the relationship explicit.\n")
		b.WriteString("\\* ================================================================\n\n")

		// Emit one consistency operator per actor showing what commands
		// each transport state implies when entered.
		for _, a := range p.Actors {
			for _, t := range a.FlattenedTransitions() {
				if len(phaseStates) > 0 && (!phaseStates[t.From] || !phaseStates[t.To]) {
					continue
				}
				if len(t.Emits) == 0 {
					continue
				}
				actionName := makeActionName(a.Name, t)
				cmds := make([]string, len(t.Emits))
				for i, c := range t.Emits {
					cmds[i] = "CMD_" + sanitiseTLA(string(c))
				}
				fmt.Fprintf(&b, "\\* %s emits: %s\n",
					actionName,
					strings.Join(cmds, ", "))
			}
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
	for _, t := range a.FlattenedTransitions() {
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
	for _, t := range a.FlattenedTransitions() {
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

func sanitiseTLA(s string) string {
	r := strings.NewReplacer(" ", "_", "-", "_", ".", "_")
	return r.Replace(s)
}

// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package protocol

import (
	"fmt"
	"io"
	"strings"
)

// ExportPlantUML writes a PlantUML state diagram for all actors.
func (p *Protocol) ExportPlantUML(w io.Writer) error {
	return p.ExportPlantUMLActors(w, "", nil)
}

// ExportPlantUMLActors writes a PlantUML state diagram for a subset
// of actors. If actors is nil, all actors are included. The title
// suffix is appended to the diagram title.
func (p *Protocol) ExportPlantUMLActors(w io.Writer, titleSuffix string, actors []string) error {
	actorSet := map[string]bool{}
	for _, a := range actors {
		actorSet[a] = true
	}
	includeActor := func(name string) bool {
		return len(actorSet) == 0 || actorSet[name]
	}
	var b strings.Builder

	diagramName := strings.ToLower(sanitisePUML(titleSuffix))
	if diagramName == "" {
		diagramName = strings.ToLower(sanitisePUML(p.Name))
	}
	b.WriteString("@startuml ")
	b.WriteString(diagramName)
	b.WriteString("\n!theme plain\nskinparam backgroundColor white\n")
	b.WriteString("skinparam state {\n  BackgroundColor #f8f8f8\n  BorderColor #888\n}\n\n")
	title := p.Name
	if titleSuffix != "" {
		title += " — " + titleSuffix
	}
	fmt.Fprintf(&b, "title %s\n\n", title)

	// Build phase lookup: state -> phase name.
	phaseOf := map[State]string{}
	for _, ph := range p.Phases {
		for _, s := range ph.States {
			phaseOf[s] = ph.Name
		}
	}
	hasPhases := len(p.Phases) > 0

	// Identify actors with cross-actor interactions (sends or receives
	// messages). Actors that only use internal events are infrastructure
	// and clutter the diagram.
	hasInteraction := map[string]bool{}
	for _, a := range p.Actors {
		for _, t := range a.FlattenedTransitions() {
			if t.On.Kind == TriggerRecv {
				hasInteraction[a.Name] = true
			}
			for _, s := range t.Sends {
				hasInteraction[a.Name] = true
				hasInteraction[s.To] = true
			}
		}
	}

	// Actor alias mapping.
	aliases := []string{"B", "C", "R", "D", "E", "F"}
	actorAlias := make(map[string]string)

	for i, a := range p.Actors {
		// When no explicit actor list is given, skip actors that have
		// no cross-actor interactions (infrastructure-only actors).
		if !includeActor(a.Name) || (len(actorSet) == 0 && !hasInteraction[a.Name]) {
			continue
		}
		alias := aliases[i%len(aliases)]
		actorAlias[a.Name] = alias

		fmt.Fprintf(&b, "state \"%s\" as %s {\n", a.Name, alias)

		if hasPhases {
			// Group this actor's states by phase.
			actorStates := collectStates(a)
			statesByPhase := map[string][]State{}
			var ungrouped []State
			for _, s := range actorStates {
				if ph, ok := phaseOf[s]; ok {
					statesByPhase[ph] = append(statesByPhase[ph], s)
				} else {
					ungrouped = append(ungrouped, s)
				}
			}

			// Emit phase superstates in definition order.
			for _, ph := range p.Phases {
				states := statesByPhase[ph.Name]
				if len(states) == 0 {
					continue
				}

				phAlias := fmt.Sprintf("%s_%s", alias, sanitisePUML(ph.Name))
				fmt.Fprintf(&b, "  state \"%s\" as %s {\n", ph.Name, phAlias)

				// Initial state: first state in this phase that matches
				// the actor's initial state, or the first listed state.
				initial := states[0]
				for _, s := range states {
					if s == a.Initial {
						initial = s
						break
					}
				}
				fmt.Fprintf(&b, "    [*] --> %s_%s\n", alias, sanitisePUML(string(initial)))

				// Transitions within this phase (coalesced).
				var phaseTransitions []Transition
				for _, t := range a.FlattenedTransitions() {
					fromPhase := phaseOf[t.From]
					toPhase := phaseOf[t.To]
					if fromPhase == ph.Name && toPhase == ph.Name {
						phaseTransitions = append(phaseTransitions, t)
					}
				}
				for _, line := range coalesceTransitions(phaseTransitions, alias) {
					fmt.Fprintf(&b, "    %s\n", line)
				}
				b.WriteString("  }\n\n")
			}

			// Cross-phase transitions (coalesced).
			var crossPhaseTransitions []Transition
			for _, t := range a.FlattenedTransitions() {
				fromPhase := phaseOf[t.From]
				toPhase := phaseOf[t.To]
				if fromPhase != toPhase {
					crossPhaseTransitions = append(crossPhaseTransitions, t)
				}
			}
			for _, line := range coalesceTransitions(crossPhaseTransitions, alias) {
				fmt.Fprintf(&b, "  %s\n", line)
			}

			// Ungrouped states.
			for _, s := range ungrouped {
				sid := fmt.Sprintf("%s_%s", alias, sanitisePUML(string(s)))
				fmt.Fprintf(&b, "  state %s\n", sid)
			}
		} else {
			// No phases — flat diagram (coalesced).
			fmt.Fprintf(&b, "  [*] --> %s_%s\n", alias, sanitisePUML(string(a.Initial)))
			for _, line := range coalesceTransitions(a.FlattenedTransitions(), alias) {
				fmt.Fprintf(&b, "  %s\n", line)
			}
		}

		b.WriteString("}\n\n")
	}

	// Cross-actor interaction arrows.
	b.WriteString("' === Cross-actor interactions ===\n\n")
	for _, a := range p.Actors {
		if !includeActor(a.Name) {
			continue
		}
		srcAlias := actorAlias[a.Name]
		for _, t := range a.FlattenedTransitions() {
			for _, s := range t.Sends {
				if !includeActor(s.To) {
					continue
				}
				dstAlias := actorAlias[s.To]
				from := fmt.Sprintf("%s_%s", srcAlias, sanitisePUML(string(t.From)))
				to := findRecvState(p, s.To, s.Msg, dstAlias)
				label := string(s.Msg)
				if len(s.Fields) > 0 {
					var fields []string
					for k := range s.Fields {
						fields = append(fields, k)
					}
					label += "\\n{" + strings.Join(fields, ", ") + "}"
				}
				fmt.Fprintf(&b, "%s -[#888,dashed]-> %s : %s\n", from, to, label)
			}
		}
	}

	// Leads-to properties are documented in the design doc, not in the
	// diagram — floating notes add clutter without connecting to states.

	b.WriteString("\n@enduml\n")
	_, err := io.WriteString(w, b.String())
	return err
}

// transitionGroupKey returns a key for coalescing transitions that share
// the same from/to states and qualifiers (guard, action, fairness).
type transitionGroupKey struct {
	From, To       string
	Guard          GuardID
	Do             ActionID
	Fairness       FairnessKind
}

func groupKey(t Transition) transitionGroupKey {
	return transitionGroupKey{
		From:     string(t.From),
		To:       string(t.To),
		Guard:    t.Guard,
		Do:       t.Do,
		Fairness: t.Fairness,
	}
}

// triggerName returns the event/message name for a transition trigger.
func triggerName(t Transition) string {
	if t.On.Kind == TriggerRecv {
		return "recv " + string(t.On.Msg)
	}
	return t.On.Desc
}

// coalesceTransitions groups transitions by (from, to, guard, do, fairness)
// and returns one label per group with events stacked vertically.
func coalesceTransitions(transitions []Transition, alias string) []string {
	type group struct {
		key      transitionGroupKey
		triggers []string
		style    string
	}
	var groups []group
	index := map[transitionGroupKey]int{}

	for _, t := range transitions {
		k := groupKey(t)
		if idx, ok := index[k]; ok {
			groups[idx].triggers = append(groups[idx].triggers, triggerName(t))
		} else {
			index[k] = len(groups)
			groups = append(groups, group{
				key:      k,
				triggers: []string{triggerName(t)},
				style:    transitionStyle(t),
			})
		}
	}

	var lines []string
	for _, g := range groups {
		from := fmt.Sprintf("%s_%s", alias, sanitisePUML(g.key.From))
		to := fmt.Sprintf("%s_%s", alias, sanitisePUML(g.key.To))

		// Build label: stacked trigger names, then qualifiers.
		var parts []string
		parts = append(parts, g.triggers...)
		if g.key.Guard != "" {
			parts = append(parts, "["+string(g.key.Guard)+"]")
		}
		if g.key.Do != "" {
			parts = append(parts, string(g.key.Do))
		}
		if g.key.Fairness == StrongFair {
			parts = append(parts, "«SF»")
		}
		label := strings.Join(parts, "\\n")
		lines = append(lines, fmt.Sprintf("%s %s %s : %s", from, g.style, to, label))
	}
	return lines
}

// transitionStyle returns PlantUML arrow style based on fairness.
func transitionStyle(t Transition) string {
	if t.Fairness == StrongFair {
		return "-[bold]->"
	}
	return "-->"
}

func findRecvState(p *Protocol, actorName string, msg MsgType, alias string) string {
	for _, a := range p.Actors {
		if a.Name != actorName {
			continue
		}
		for _, t := range a.FlattenedTransitions() {
			if t.On.Kind == TriggerRecv && t.On.Msg == msg {
				return fmt.Sprintf("%s_%s", alias, sanitisePUML(string(t.From)))
			}
		}
	}
	return fmt.Sprintf("%s_%s", alias, "unknown")
}

func sanitisePUML(s string) string {
	return strings.ReplaceAll(s, " ", "_")
}

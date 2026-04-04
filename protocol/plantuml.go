// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package protocol

import (
	"fmt"
	"io"
	"strings"
)

// ExportPlantUML writes a PlantUML state diagram. If the protocol has
// phases defined, states are grouped into hierarchical superstates.
// Each actor gets its own concurrent region.
func (p *Protocol) ExportPlantUML(w io.Writer) error {
	var b strings.Builder

	b.WriteString("@startuml ")
	b.WriteString(sanitisePUML(p.Name))
	b.WriteString("\n!theme plain\nskinparam backgroundColor white\n")
	b.WriteString("skinparam state {\n  BackgroundColor #f8f8f8\n  BorderColor #888\n}\n\n")
	fmt.Fprintf(&b, "title %s\n\n", p.Name)

	// Build phase lookup: state -> phase name.
	phaseOf := map[State]string{}
	for _, ph := range p.Phases {
		for _, s := range ph.States {
			phaseOf[s] = ph.Name
		}
	}
	hasPhases := len(p.Phases) > 0

	// Actor alias mapping.
	aliases := []string{"B", "C", "R", "D", "E", "F"}
	actorAlias := make(map[string]string)

	for i, a := range p.Actors {
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

				// Transitions within this phase.
				for _, t := range a.Transitions {
					fromPhase := phaseOf[t.From]
					toPhase := phaseOf[t.To]
					if fromPhase == ph.Name && toPhase == ph.Name {
						from := fmt.Sprintf("%s_%s", alias, sanitisePUML(string(t.From)))
						to := fmt.Sprintf("%s_%s", alias, sanitisePUML(string(t.To)))
						fmt.Fprintf(&b, "    %s %s %s : %s\n", from, transitionStyle(t), to, transitionLabel(t))
					}
				}
				b.WriteString("  }\n\n")
			}

			// Cross-phase transitions.
			for _, t := range a.Transitions {
				fromPhase := phaseOf[t.From]
				toPhase := phaseOf[t.To]
				if fromPhase != toPhase {
					from := fmt.Sprintf("%s_%s", alias, sanitisePUML(string(t.From)))
					to := fmt.Sprintf("%s_%s", alias, sanitisePUML(string(t.To)))
					fmt.Fprintf(&b, "  %s %s %s : %s\n", from, transitionStyle(t), to, transitionLabel(t))
				}
			}

			// Ungrouped states.
			for _, s := range ungrouped {
				sid := fmt.Sprintf("%s_%s", alias, sanitisePUML(string(s)))
				fmt.Fprintf(&b, "  state %s\n", sid)
			}
		} else {
			// No phases — flat diagram.
			fmt.Fprintf(&b, "  [*] --> %s_%s\n", alias, sanitisePUML(string(a.Initial)))
			for _, t := range a.Transitions {
				from := fmt.Sprintf("%s_%s", alias, sanitisePUML(string(t.From)))
				to := fmt.Sprintf("%s_%s", alias, sanitisePUML(string(t.To)))
				fmt.Fprintf(&b, "  %s %s %s : %s\n", from, transitionStyle(t), to, transitionLabel(t))
			}
		}

		b.WriteString("}\n\n")
	}

	// Cross-actor interaction arrows.
	colors := []string{"#blue", "#green", "#orange", "#purple", "#gray", "#red"}
	b.WriteString("' === Cross-actor interactions ===\n\n")
	colorIdx := 0
	for _, a := range p.Actors {
		srcAlias := actorAlias[a.Name]
		for _, t := range a.Transitions {
			for _, s := range t.Sends {
				dstAlias := actorAlias[s.To]
				from := fmt.Sprintf("%s_%s", srcAlias, sanitisePUML(string(t.From)))
				to := findRecvState(p, s.To, s.Msg, dstAlias)
				color := colors[colorIdx%len(colors)]
				label := string(s.Msg)
				if len(s.Fields) > 0 {
					var fields []string
					for k := range s.Fields {
						fields = append(fields, k)
					}
					label += "\\n{" + strings.Join(fields, ", ") + "}"
				}
				fmt.Fprintf(&b, "%s -[%s,dashed]-> %s : %s\n", from, color, to, label)
				colorIdx++
			}
		}
	}

	// Leads-to properties as dashed arrows.
	hasLeadsTo := false
	for _, prop := range p.Properties {
		if prop.Kind == LeadsTo {
			if !hasLeadsTo {
				b.WriteString("\n' === Leads-to properties ===\n\n")
				hasLeadsTo = true
			}
			fmt.Fprintf(&b, "note \"%s\\n%s ~> %s\" as N_%s\n",
				prop.Name,
				prop.FromExpr, prop.ToExpr,
				sanitisePUML(prop.Name))
		}
	}

	b.WriteString("\n@enduml\n")
	_, err := io.WriteString(w, b.String())
	return err
}

func transitionLabel(t Transition) string {
	var parts []string
	if t.On.Kind == TriggerRecv {
		parts = append(parts, "recv "+string(t.On.Msg))
	} else if t.On.Desc != "" {
		parts = append(parts, t.On.Desc)
	}
	if t.Guard != "" {
		parts = append(parts, "["+string(t.Guard)+"]")
	}
	if t.Do != "" {
		parts = append(parts, string(t.Do))
	}
	if t.Fairness == StrongFair {
		parts = append(parts, "«SF»")
	}
	return strings.Join(parts, "\\n")
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
		for _, t := range a.Transitions {
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

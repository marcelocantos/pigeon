// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package protocol

import (
	"fmt"
	"sync"
)

// GuardFunc evaluates whether a transition's guard condition is met.
// The context parameter carries protocol-specific state.
type GuardFunc func(ctx any) bool

// ActionFunc executes a side-effect when a transition fires.
type ActionFunc func(ctx any) error

// Machine is a runtime state machine executor for one actor in a protocol.
// It enforces the transition table — any message or event not matching a
// valid transition from the current state is rejected.
type Machine struct {
	mu      sync.Mutex
	actor   Actor
	state   State
	guards  map[GuardID]GuardFunc
	actions map[ActionID]ActionFunc

	// index: (state, msgType) → applicable transitions
	recvIndex     map[stateMsg][]Transition
	internalIndex map[stateEvent][]Transition
}

type stateMsg struct {
	state State
	msg   MsgType
}

type stateEvent struct {
	state State
	event EventID
}

// NewMachine creates a runtime state machine for the named actor.
// Guards and actions are registered separately via the returned Machine.
func NewMachine(p *Protocol, actorName string) (*Machine, error) {
	var actor *Actor
	for i := range p.Actors {
		if p.Actors[i].Name == actorName {
			actor = &p.Actors[i]
			break
		}
	}
	if actor == nil {
		return nil, fmt.Errorf("actor %q not found in protocol %q", actorName, p.Name)
	}

	m := &Machine{
		actor:         *actor,
		state:         actor.Initial,
		guards:        make(map[GuardID]GuardFunc),
		actions:       make(map[ActionID]ActionFunc),
		recvIndex:     make(map[stateMsg][]Transition),
		internalIndex: make(map[stateEvent][]Transition),
	}

	for _, t := range actor.FlattenedTransitions() {
		switch t.On.Kind {
		case TriggerRecv:
			key := stateMsg{t.From, t.On.Msg}
			m.recvIndex[key] = append(m.recvIndex[key], t)
		case TriggerInternal:
			key := stateEvent{t.From, EventID(t.On.Desc)}
			m.internalIndex[key] = append(m.internalIndex[key], t)
		}
	}

	return m, nil
}

// RegisterGuard binds a guard function to a GuardID.
func (m *Machine) RegisterGuard(id GuardID, fn GuardFunc) {
	m.guards[id] = fn
}

// RegisterAction binds an action function to an ActionID.
func (m *Machine) RegisterAction(id ActionID, fn ActionFunc) {
	m.actions[id] = fn
}

// State returns the current state.
func (m *Machine) State() State {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.state
}

// HandleMessage processes a received message. Returns the new state
// or an error if no valid transition exists.
func (m *Machine) HandleMessage(msg MsgType, ctx any) (State, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := stateMsg{m.state, msg}
	transitions := m.recvIndex[key]
	if len(transitions) == 0 {
		return m.state, fmt.Errorf("no transition from %s on message %s", m.state, msg)
	}

	return m.tryTransitions(transitions, ctx)
}

// Step attempts an internal transition from the current state.
// If event is non-empty, only transitions triggered by that event are
// considered. If event is empty, all internal transitions from the
// current state are tried (first matching guard wins).
func (m *Machine) Step(event EventID, ctx any) (State, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if event != "" {
		key := stateEvent{m.state, event}
		transitions := m.internalIndex[key]
		if len(transitions) == 0 {
			return m.state, fmt.Errorf("no internal transition from %s on event %s", m.state, event)
		}
		return m.tryTransitions(transitions, ctx)
	}

	// No event specified — try all internal transitions from this state.
	var all []Transition
	for key, ts := range m.internalIndex {
		if key.state == m.state {
			all = append(all, ts...)
		}
	}
	if len(all) == 0 {
		return m.state, fmt.Errorf("no internal transition from %s", m.state)
	}
	return m.tryTransitions(all, ctx)
}

func (m *Machine) tryTransitions(transitions []Transition, ctx any) (State, error) {
	for _, t := range transitions {
		if t.Guard != "" {
			guard, ok := m.guards[t.Guard]
			if !ok {
				return m.state, fmt.Errorf("unregistered guard: %s", t.Guard)
			}
			if !guard(ctx) {
				continue
			}
		}

		if t.Do != "" {
			action, ok := m.actions[t.Do]
			if !ok {
				return m.state, fmt.Errorf("unregistered action: %s", t.Do)
			}
			if err := action(ctx); err != nil {
				return m.state, fmt.Errorf("action %s failed: %w", t.Do, err)
			}
		}

		m.state = t.To
		return m.state, nil
	}

	return m.state, fmt.Errorf("all guards failed for transitions from %s", m.state)
}

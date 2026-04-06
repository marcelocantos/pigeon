// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Package protocol provides a declarative state machine framework for
// security protocols. State machines are defined as data (transition
// tables) that serve as the single source of truth for both runtime
// execution and TLA+ model generation.
package protocol

import "fmt"

// State is a named state in an actor's state machine.
type State string

// MsgType identifies a message kind exchanged between actors.
type MsgType string

// GuardID identifies a named guard predicate.
type GuardID string

// ActionID identifies a named side-effect action.
type ActionID string

// EventID identifies an internal event that triggers a transition.
type EventID string

// CmdID identifies a command emitted by a transition.
type CmdID string

// PropertyKind classifies a verification property.
type PropertyKind int

const (
	Invariant PropertyKind = iota
	Liveness              // <>P
	LeadsTo               // P ~> Q
)

// FairnessKind specifies when TLC must force a transition to fire.
type FairnessKind int

const (
	WeakFair   FairnessKind = iota // if continuously enabled, eventually taken
	StrongFair                     // if enabled infinitely often, eventually taken
)

// Transition defines a single edge in an actor's state machine.
type Transition struct {
	From     State
	To       State
	On       Trigger      // what causes this transition
	Guard    GuardID      // optional: must be true to take this edge
	Do       ActionID     // optional: side-effect on transition
	Fairness FairnessKind // weak (default) or strong
	Sends    []Send       // messages emitted when this transition fires
	Updates  []VarUpdate  // auxiliary variable updates
	Emits    []CmdID      // commands emitted when this transition fires
}

// Send describes a message emitted during a transition.
type Send struct {
	To     string            // receiver actor name
	Msg    MsgType           // message type
	Fields map[string]string // TLA+ expressions for message record fields
}

// VarUpdate sets an auxiliary variable to a TLA+ expression.
type VarUpdate struct {
	Var  string // variable name
	Expr string // TLA+ expression
}

// TriggerKind distinguishes message receipt from internal events.
type TriggerKind int

const (
	TriggerRecv    TriggerKind = iota // message received from another actor
	TriggerInternal                   // internal event (timeout, completion)
)

// Trigger describes what causes a transition.
type Trigger struct {
	Kind TriggerKind
	Msg  MsgType // for TriggerRecv: the message type
	Desc string  // human-readable label (used in TLA+ comments)
}

// Recv creates a trigger for receiving a message.
func Recv(msg MsgType) Trigger {
	return Trigger{Kind: TriggerRecv, Msg: msg}
}

// Internal creates a trigger for an internal event.
func Internal(desc string) Trigger {
	return Trigger{Kind: TriggerInternal, Desc: desc}
}

// Message defines a message type in the protocol, with its sender
// and receiver. This determines TLA+ channel topology.
type Message struct {
	Type MsgType
	From string // actor name
	To   string // actor name
	Desc string // human-readable description
}

// VarDef defines an auxiliary state variable for TLA+ model checking.
// These track protocol-level state that isn't part of any single actor.
// VarType represents the type of a state variable for code generation.
// TLA+ doesn't use these (it's untyped); they're for the executor.
type VarType string

const (
	VarString    VarType = "string"
	VarInt       VarType = "int"
	VarBool      VarType = "bool"
	VarSetString VarType = "set<string>"
)

type VarDef struct {
	Name    string  // TLA+ variable name
	Type    VarType // executor type (string, int, bool, set<string>, or struct name)
	Initial string  // TLA+ expression for initial value
	Desc    string  // human-readable description
}

// StructDef defines a named group of variables that are logically
// related and often updated together. In the executor, each struct
// becomes a native struct/class. In TLA+, each struct is a record.
type StructDef struct {
	Name   string
	Fields []StructField
	Desc   string
}

// StructField is a single field within a StructDef.
type StructField struct {
	Name    string
	Type    VarType
	Initial string
	Desc    string
}

// GuardDef maps a GuardID to a TLA+ expression, binding the
// declarative guard used in transitions to a checkable predicate.
type GuardDef struct {
	ID   GuardID
	Expr string // TLA+ boolean expression
}

// EventDef declares an event that can trigger transitions.
type EventDef struct {
	ID   EventID
	Desc string
}

// CommandDef declares a command that transitions can emit.
type CommandDef struct {
	ID   CmdID
	Desc string
}

// Operator defines a TLA+ helper operator used in guard expressions,
// variable updates, or properties. These capture domain-specific
// logic (e.g., symbolic crypto operations).
type Operator struct {
	Name   string // operator name
	Params string // comma-separated parameter list (e.g., "a, b")
	Expr   string // TLA+ expression body
	Desc   string // human-readable description
}

// AdvAction defines an adversary capability as a PlusCal code block.
// Each action becomes an either/or branch in the adversary process,
// in addition to the standard Dolev-Yao eavesdrop/drop/replay.
type AdvAction struct {
	Name string // identifier (used in comments)
	Desc string // what attack this models
	Code string // PlusCal code (indented inside either/or branch)
}

// Property defines a verification property for TLA+ model checking.
type Property struct {
	Name     string
	Kind     PropertyKind
	Expr     string // TLA+ expression (for invariant and liveness)
	FromExpr string // TLA+ expression for leads-to LHS (P ~> Q: P)
	ToExpr   string // TLA+ expression for leads-to RHS (P ~> Q: Q)
	Desc     string // human-readable description
}

// WireConstant defines a protocol constant shared across all platforms.
// These are emitted as typed constants/enums in every generated language.
type WireConstant struct {
	Name  string
	Value any    // int, float64, string, or bool
	Type  string // "int", "byte", "duration_ms", "string"
	Desc  string
	Group string // optional grouping key (e.g., "datagram_framing")
}

// ConstantDef defines a parameterised constant for model checking.
// TLC substitutes values from the set to verify the protocol works
// for all possible values, not just one hardcoded scenario.
type ConstantDef struct {
	Name   string
	Type   VarType
	Values []string // set of values to check (TLA+ expressions)
	Desc   string
}

// StateNode represents a state in the hierarchy. Leaf states have no
// children. Superstates have children and inherited transitions that
// apply to all descendants (self-loops).
type StateNode struct {
	Name        State
	Parent      *StateNode
	Children    []*StateNode
	Transitions []Transition // inherited by all descendants
}

// IsLeaf reports whether this state has no children.
func (n *StateNode) IsLeaf() bool { return len(n.Children) == 0 }

// LeafStates returns all leaf descendants (or self if leaf).
func (n *StateNode) LeafStates() []*StateNode {
	if n.IsLeaf() {
		return []*StateNode{n}
	}
	var leaves []*StateNode
	for _, c := range n.Children {
		leaves = append(leaves, c.LeafStates()...)
	}
	return leaves
}

// AncestorChain returns [self, parent, grandparent, ...] up to root.
func (n *StateNode) AncestorChain() []*StateNode {
	var chain []*StateNode
	for cur := n; cur != nil; cur = cur.Parent {
		chain = append(chain, cur)
	}
	return chain
}

// Actor defines one participant in the protocol.
type Actor struct {
	Name        string
	Initial     State
	Transitions []Transition          // leaf-level transitions
	StateIndex  map[State]*StateNode  // hierarchy (nil = flat)
	Roots       []*StateNode          // top-level state nodes
}

// FlattenedTransitions returns all transitions including superstate
// transitions expanded into explicit per-leaf-state self-loops.
// The result is suitable for generators that need a flat transition list.
func (a *Actor) FlattenedTransitions() []Transition {
	if len(a.StateIndex) == 0 {
		return a.Transitions
	}

	// Start with the explicit leaf transitions.
	result := make([]Transition, len(a.Transitions))
	copy(result, a.Transitions)

	// Expand each superstate's transitions into leaf self-loops.
	// Walk the hierarchy in definition order (roots first).
	var expand func(node *StateNode)
	expand = func(node *StateNode) {
		if !node.IsLeaf() {
			leaves := node.LeafStates()
			for _, t := range node.Transitions {
				for _, leaf := range leaves {
					expanded := t
					expanded.From = leaf.Name
					expanded.To = leaf.Name
					result = append(result, expanded)
				}
			}
			for _, child := range node.Children {
				expand(child)
			}
		}
	}
	for _, root := range a.Roots {
		expand(root)
	}

	return result
}

// Protocol is the complete definition of a multi-actor protocol.
// This is the single source of truth for runtime and TLA+ generation.
// Phase groups states for diagramming (hierarchical superstates in
// PlantUML) and for optimising TLA+ verification (variables scoped to
// phases are frozen as UNCHANGED outside their phase).
type Phase struct {
	Name     string
	States   []State  // states belonging to this phase
	Vars     []string // variables that change in this phase (empty = all)
	Adversary bool    // whether the adversary is active in this phase
}

type Protocol struct {
	Name         string
	Actors       []Actor
	Messages     []Message
	Events       []EventDef   // declared event types (triggers for transitions)
	Commands     []CommandDef // declared command types (emitted by transitions)
	Structs      []StructDef  // named variable groups
	Vars         []VarDef     // auxiliary state variables (may reference structs)
	Guards       []GuardDef   // guard TLA+ expressions
	Operators    []Operator   // TLA+ helper operators
	AdvActions   []AdvAction  // adversary capabilities beyond Dolev-Yao
	AdvGuard     string       // TLA+ expression gating the adversary (empty = always active)
	Phases       []Phase      // named groupings of states for diagramming and TLA+ splitting
	WireConsts   []WireConstant // protocol constants emitted in all languages
	Constants    []ConstantDef  // parameterised constants for model checking
	Properties   []Property
	ChannelBound int // max messages per channel (0 = unbounded)
	OneShot      bool // if true, actors run once then terminate (no loop)
}

// collectStates returns all unique leaf states for an actor in definition order.
// Superstate names are excluded — they are containers, not states.
func collectStates(a Actor) []State {
	seen := map[State]bool{}
	var states []State
	add := func(s State) {
		if !seen[s] {
			seen[s] = true
			states = append(states, s)
		}
	}
	add(a.Initial)
	// Include states from flattened transitions (covers hierarchy).
	for _, t := range a.FlattenedTransitions() {
		add(t.From)
		add(t.To)
	}
	return states
}

// Validate checks the protocol definition for internal consistency:
// all states reachable, all message types declared, sends reference
// valid actors and message types, guards are defined.
func (p *Protocol) Validate() error {
	actorNames := map[string]bool{}
	for _, a := range p.Actors {
		if actorNames[a.Name] {
			return fmt.Errorf("duplicate actor name: %s", a.Name)
		}
		actorNames[a.Name] = true
	}

	msgTypes := map[MsgType]bool{}
	msgSenders := map[MsgType]string{}
	for _, m := range p.Messages {
		msgTypes[m.Type] = true
		msgSenders[m.Type] = m.From
		if !actorNames[m.From] {
			return fmt.Errorf("message %s: unknown sender %q", m.Type, m.From)
		}
		if !actorNames[m.To] {
			return fmt.Errorf("message %s: unknown receiver %q", m.Type, m.To)
		}
	}

	guardDefs := map[GuardID]bool{}
	for _, g := range p.Guards {
		guardDefs[g.ID] = true
	}

	for _, a := range p.Actors {
		// Validate hierarchy consistency.
		for name, node := range a.StateIndex {
			for _, child := range node.Children {
				if child.Parent != node {
					return fmt.Errorf("actor %s: state %s child %s has wrong parent", a.Name, name, child.Name)
				}
			}
		}

		// Validate both explicit and superstate transitions.
		for _, t := range a.FlattenedTransitions() {
			if t.On.Kind == TriggerRecv && !msgTypes[t.On.Msg] {
				return fmt.Errorf("actor %s: transition %s->%s triggers on undeclared message %s",
					a.Name, t.From, t.To, t.On.Msg)
			}
			if t.Guard != "" && !guardDefs[t.Guard] {
				return fmt.Errorf("actor %s: transition %s->%s uses undefined guard %s",
					a.Name, t.From, t.To, t.Guard)
			}
			for _, s := range t.Sends {
				if !actorNames[s.To] {
					return fmt.Errorf("actor %s: transition %s->%s sends to unknown actor %q",
						a.Name, t.From, t.To, s.To)
				}
				if !msgTypes[s.Msg] {
					return fmt.Errorf("actor %s: transition %s->%s sends undeclared message %s",
						a.Name, t.From, t.To, s.Msg)
				}
				if sender := msgSenders[s.Msg]; sender != a.Name {
					return fmt.Errorf("actor %s: transition %s->%s sends message %s but declared sender is %s",
						a.Name, t.From, t.To, s.Msg, sender)
				}
			}
		}
	}

	return nil
}

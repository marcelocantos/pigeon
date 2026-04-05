// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package protocol

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// YAML parse types — mirrors the YAML schema.

type yamlProtocol struct {
	Name         string                     `yaml:"name"`
	Messages     yaml.Node                  `yaml:"messages"`
	Events       yaml.Node                  `yaml:"events"`
	Commands     yaml.Node                  `yaml:"commands"`
	Actors       yaml.Node                  `yaml:"actors"`
	Structs      yaml.Node                  `yaml:"structs"`
	Vars         yaml.Node                  `yaml:"vars"`
	Guards       yaml.Node                  `yaml:"guards"`
	Operators    yaml.Node                  `yaml:"operators"`
	Phases       yaml.Node                  `yaml:"phases"`
	Constants    []yamlConstant             `yaml:"constants"`
	AdvGuard     string                     `yaml:"adversary_guard"`
	Adversary    []yamlAdvAction            `yaml:"adversary"`
	Properties   []yamlProperty             `yaml:"properties"`
	ChannelBound int                        `yaml:"channel_bound"`
	OneShot      bool                       `yaml:"one_shot"`
}

type yamlMessage struct {
	From string `yaml:"from"`
	To   string `yaml:"to"`
	Desc string `yaml:"desc"`
}

type yamlActor struct {
	Initial     string           `yaml:"initial"`
	Transitions []yamlTransition `yaml:"transitions"`
}

type yamlTransition struct {
	From     string              `yaml:"from"`
	To       string              `yaml:"to"`
	On       string              `yaml:"on"`
	Guard    string              `yaml:"guard"`
	Do       string              `yaml:"do"`
	Fairness string              `yaml:"fairness"` // "weak" (default) or "strong"
	Sends    []yamlSend          `yaml:"sends"`
	Updates  yaml.Node           `yaml:"updates"`
	Emits    []string            `yaml:"emits"`
}

type yamlSend struct {
	To     string            `yaml:"to"`
	Msg    string            `yaml:"msg"`
	Fields map[string]string `yaml:"fields"`
}

type yamlVar struct {
	Initial string `yaml:"initial"`
	Type    string `yaml:"type"`
	Desc    string `yaml:"desc"`
}

type yamlOperator struct {
	Params string `yaml:"params"`
	Expr   string `yaml:"expr"`
	Desc   string `yaml:"desc"`
}

type yamlConstant struct {
	Name   string   `yaml:"name"`
	Type   string   `yaml:"type"`
	Values []string `yaml:"values"`
	Desc   string   `yaml:"desc"`
}

type yamlAdvAction struct {
	Name string `yaml:"name"`
	Desc string `yaml:"desc"`
	Code string `yaml:"code"`
}

type yamlProperty struct {
	Name     string `yaml:"name"`
	Kind     string `yaml:"kind"`      // invariant, liveness, leads_to
	Expr     string `yaml:"expr"`      // for invariant and liveness
	FromExpr string `yaml:"from_expr"` // for leads_to: P in P ~> Q
	ToExpr   string `yaml:"to_expr"`   // for leads_to: Q in P ~> Q
	Desc     string `yaml:"desc"`
}

// LoadYAML reads a protocol definition from a YAML file and returns
// a Protocol struct suitable for use with the runtime executor and
// all code generators.
func LoadYAML(path string) (*Protocol, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseYAML(data)
}

// ParseYAML parses a YAML protocol definition.
func ParseYAML(data []byte) (*Protocol, error) {
	var yp yamlProtocol
	if err := yaml.Unmarshal(data, &yp); err != nil {
		return nil, fmt.Errorf("parse YAML: %w", err)
	}

	p := &Protocol{
		Name:         yp.Name,
		ChannelBound: yp.ChannelBound,
		OneShot:      yp.OneShot,
	}

	// Events — declared event types.
	evts, err := parseOrderedMapString(&yp.Events)
	if err != nil {
		return nil, fmt.Errorf("events: %w", err)
	}
	for _, kv := range evts {
		p.Events = append(p.Events, EventDef{
			ID:   EventID(kv.key),
			Desc: kv.val,
		})
	}

	// Commands — declared command types.
	cmds, err := parseOrderedMapString(&yp.Commands)
	if err != nil {
		return nil, fmt.Errorf("commands: %w", err)
	}
	for _, kv := range cmds {
		p.Commands = append(p.Commands, CommandDef{
			ID:   CmdID(kv.key),
			Desc: kv.val,
		})
	}

	// Messages — preserve YAML key order.
	msgs, err := parseOrderedMap[yamlMessage](&yp.Messages)
	if err != nil {
		return nil, fmt.Errorf("messages: %w", err)
	}
	for _, kv := range msgs {
		p.Messages = append(p.Messages, Message{
			Type: MsgType(kv.key),
			From: kv.val.From,
			To:   kv.val.To,
			Desc: kv.val.Desc,
		})
	}

	// Actors — preserve YAML key order.
	actors, err := parseOrderedMap[yamlActor](&yp.Actors)
	if err != nil {
		return nil, fmt.Errorf("actors: %w", err)
	}
	for _, kv := range actors {
		actor := Actor{
			Name:    kv.key,
			Initial: State(kv.val.Initial),
		}
		for i, yt := range kv.val.Transitions {
			t, err := convertTransition(yt)
			if err != nil {
				return nil, fmt.Errorf("actor %s transition %d: %w", kv.key, i, err)
			}
			actor.Transitions = append(actor.Transitions, t)
		}
		p.Actors = append(p.Actors, actor)
	}

	// Structs — named variable groups.
	if yp.Structs.Kind == yaml.MappingNode {
		for i := 0; i+1 < len(yp.Structs.Content); i += 2 {
			structName := yp.Structs.Content[i].Value
			fieldsNode := yp.Structs.Content[i+1]

			sd := StructDef{Name: structName}

			if fieldsNode.Kind == yaml.MappingNode {
				fields, err := parseOrderedMap[yamlVar](fieldsNode)
				if err != nil {
					return nil, fmt.Errorf("struct %q: %w", structName, err)
				}
				for _, kv := range fields {
					sd.Fields = append(sd.Fields, StructField{
						Name:    kv.key,
						Type:    parseVarType(kv.val.Type),
						Initial: kv.val.Initial,
						Desc:    kv.val.Desc,
					})
				}
			}

			p.Structs = append(p.Structs, sd)
		}
	}

	// Vars — preserve YAML key order.
	vars, err := parseOrderedMap[yamlVar](&yp.Vars)
	if err != nil {
		return nil, fmt.Errorf("vars: %w", err)
	}
	for _, kv := range vars {
		p.Vars = append(p.Vars, VarDef{
			Name:    kv.key,
			Type:    parseVarType(kv.val.Type),
			Initial: kv.val.Initial,
			Desc:    kv.val.Desc,
		})
	}

	// Guards — preserve YAML key order.
	guards, err := parseOrderedMapString(&yp.Guards)
	if err != nil {
		return nil, fmt.Errorf("guards: %w", err)
	}
	for _, kv := range guards {
		p.Guards = append(p.Guards, GuardDef{
			ID:   GuardID(kv.key),
			Expr: kv.val,
		})
	}

	// Operators — preserve YAML key order.
	ops, err := parseOrderedMap[yamlOperator](&yp.Operators)
	if err != nil {
		return nil, fmt.Errorf("operators: %w", err)
	}
	for _, kv := range ops {
		p.Operators = append(p.Operators, Operator{
			Name:   kv.key,
			Params: kv.val.Params,
			Expr:   kv.val.Expr,
			Desc:   kv.val.Desc,
		})
	}

	// Adversary actions.
	for _, ya := range yp.Adversary {
		code := strings.TrimRight(ya.Code, "\n")
		// Indent code lines for PlusCal either/or block.
		var indented []string
		for _, line := range strings.Split(code, "\n") {
			indented = append(indented, "      "+line)
		}
		p.AdvActions = append(p.AdvActions, AdvAction{
			Name: ya.Name,
			Desc: ya.Desc,
			Code: strings.Join(indented, "\n"),
		})
	}

	// Phases — parse either simple list or rich object form.
	if yp.Phases.Kind == yaml.MappingNode {
		for i := 0; i+1 < len(yp.Phases.Content); i += 2 {
			name := yp.Phases.Content[i].Value
			valueNode := yp.Phases.Content[i+1]

			ph := Phase{Name: name}

			if valueNode.Kind == yaml.SequenceNode {
				// Simple form: phases: { Pairing: [Idle, ...] }
				for _, n := range valueNode.Content {
					ph.States = append(ph.States, State(n.Value))
				}
			} else if valueNode.Kind == yaml.MappingNode {
				// Rich form: phases: { Pairing: { states: [...], vars: [...], adversary: true } }
				var rich struct {
					States    []string `yaml:"states"`
					Vars      []string `yaml:"vars"`
					Adversary bool     `yaml:"adversary"`
				}
				if err := valueNode.Decode(&rich); err != nil {
					return nil, fmt.Errorf("parse phase %q: %w", name, err)
				}
				for _, s := range rich.States {
					ph.States = append(ph.States, State(s))
				}
				ph.Vars = rich.Vars
				ph.Adversary = rich.Adversary
			}

			p.Phases = append(p.Phases, ph)
		}
	}

	p.AdvGuard = yp.AdvGuard

	// Constants.
	for _, yc := range yp.Constants {
		p.Constants = append(p.Constants, ConstantDef{
			Name:   yc.Name,
			Type:   parseVarType(yc.Type),
			Values: yc.Values,
			Desc:   yc.Desc,
		})
	}

	// Properties.
	for _, ypr := range yp.Properties {
		kind := Invariant
		switch ypr.Kind {
		case "liveness":
			kind = Liveness
		case "leads_to":
			kind = LeadsTo
		}
		p.Properties = append(p.Properties, Property{
			Name:     ypr.Name,
			Kind:     kind,
			Expr:     ypr.Expr,
			FromExpr: ypr.FromExpr,
			ToExpr:   ypr.ToExpr,
			Desc:     ypr.Desc,
		})
	}

	return p, nil
}

func convertTransition(yt yamlTransition) (Transition, error) {
	fairness := WeakFair
	if yt.Fairness == "strong" {
		fairness = StrongFair
	}

	t := Transition{
		From:     State(yt.From),
		To:       State(yt.To),
		Guard:    GuardID(yt.Guard),
		Do:       ActionID(yt.Do),
		Fairness: fairness,
	}

	// Parse trigger: "recv <msg>" or free-form internal description.
	if strings.HasPrefix(yt.On, "recv ") {
		t.On = Recv(MsgType(strings.TrimPrefix(yt.On, "recv ")))
	} else {
		t.On = Internal(yt.On)
	}

	// Sends.
	for _, ys := range yt.Sends {
		t.Sends = append(t.Sends, Send{
			To:     ys.To,
			Msg:    MsgType(ys.Msg),
			Fields: ys.Fields,
		})
	}

	// Emits — commands emitted by this transition.
	for _, e := range yt.Emits {
		t.Emits = append(t.Emits, CmdID(e))
	}

	// Updates — preserve YAML key order.
	if yt.Updates.Kind != 0 {
		updates, err := parseOrderedMapString(&yt.Updates)
		if err != nil {
			return t, fmt.Errorf("updates: %w", err)
		}
		for _, kv := range updates {
			t.Updates = append(t.Updates, VarUpdate{
				Var:  kv.key,
				Expr: kv.val,
			})
		}
	}

	return t, nil
}

// Ordered map parsing — YAML maps don't guarantee order, but
// yaml.Node preserves it.

type kv[V any] struct {
	key string
	val V
}

func parseOrderedMap[V any](node *yaml.Node) ([]kv[V], error) {
	if node.Kind == 0 {
		return nil, nil
	}
	if node.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("expected mapping, got %d", node.Kind)
	}
	var result []kv[V]
	for i := 0; i < len(node.Content); i += 2 {
		key := node.Content[i].Value
		var val V
		if err := node.Content[i+1].Decode(&val); err != nil {
			return nil, fmt.Errorf("key %s: %w", key, err)
		}
		result = append(result, kv[V]{key: key, val: val})
	}
	return result, nil
}

func parseVarType(s string) VarType {
	switch s {
	case "int":
		return VarInt
	case "bool":
		return VarBool
	case "set<string>":
		return VarSetString
	default:
		return VarString
	}
}

func parseOrderedMapString(node *yaml.Node) ([]kv[string], error) {
	if node.Kind == 0 {
		return nil, nil
	}
	if node.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("expected mapping, got %d", node.Kind)
	}
	var result []kv[string]
	for i := 0; i < len(node.Content); i += 2 {
		result = append(result, kv[string]{
			key: node.Content[i].Value,
			val: node.Content[i+1].Value,
		})
	}
	return result, nil
}

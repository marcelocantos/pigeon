// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package protocol

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestPairingCeremonyValidates(t *testing.T) {
	p := PairingCeremony()
	if err := p.Validate(); err != nil {
		t.Fatalf("PairingCeremony validation failed: %v", err)
	}
}

func noopActions(m *Machine, ids ...ActionID) {
	for _, id := range ids {
		m.RegisterAction(id, func(any) error { return nil })
	}
}

func TestMachineHappyPath(t *testing.T) {
	p := PairingCeremony()

	m, err := NewMachine(p, "server")
	if err != nil {
		t.Fatalf("NewMachine: %v", err)
	}

	m.RegisterGuard(PairingCeremonyGuardTokenValid, func(any) bool { return true })
	m.RegisterGuard(PairingCeremonyGuardTokenInvalid, func(any) bool { return false })
	m.RegisterGuard(PairingCeremonyGuardCodeCorrect, func(any) bool { return true })
	m.RegisterGuard(PairingCeremonyGuardCodeWrong, func(any) bool { return false })
	m.RegisterGuard(PairingCeremonyGuardDeviceKnown, func(any) bool { return true })
	m.RegisterGuard(PairingCeremonyGuardDeviceUnknown, func(any) bool { return false })
	m.RegisterGuard(PairingCeremonyGuardNonceFresh, func(any) bool { return true })

	noopActions(m,
		PairingCeremonyActionGenerateToken, PairingCeremonyActionRegisterRelay, PairingCeremonyActionDeriveSecret,
		PairingCeremonyActionStoreDevice, PairingCeremonyActionVerifyDevice,
	)

	assertState := func(expected State) {
		t.Helper()
		if got := m.State(); got != expected {
			t.Fatalf("expected state %s, got %s", expected, got)
		}
	}

	assertState(PairingCeremonyServerIdle)
	mustHandle(t, m, PairingCeremonyMsgPairBegin)
	assertState(PairingCeremonyServerGenerateToken)
	mustStep(t, m) // -> RegisterRelay
	mustStep(t, m) // -> WaitingForClient
	mustHandle(t, m, PairingCeremonyMsgPairHello)
	assertState(PairingCeremonyServerDeriveSecret)
	mustStep(t, m) // -> SendAck
	mustStep(t, m) // -> WaitingForCode
	mustHandle(t, m, PairingCeremonyMsgCodeSubmit)
	assertState(PairingCeremonyServerValidateCode)
	mustStep(t, m) // -> StorePaired
	mustStep(t, m) // -> Paired
	assertState(PairingCeremonyServerPaired)

	// Reconnection.
	mustHandle(t, m, PairingCeremonyMsgAuthRequest)
	mustStep(t, m) // -> SessionActive
	assertState(PairingCeremonyServerSessionActive)
}

func TestMachineTokenRejection(t *testing.T) {
	p := PairingCeremony()
	m, err := NewMachine(p, "server")
	if err != nil {
		t.Fatalf("NewMachine: %v", err)
	}

	m.RegisterGuard(PairingCeremonyGuardTokenValid, func(any) bool { return false })
	m.RegisterGuard(PairingCeremonyGuardTokenInvalid, func(any) bool { return true })
	noopActions(m, PairingCeremonyActionGenerateToken, PairingCeremonyActionRegisterRelay)

	mustHandle(t, m, PairingCeremonyMsgPairBegin)
	mustStep(t, m) // -> RegisterRelay
	mustStep(t, m) // -> WaitingForClient
	mustHandle(t, m, PairingCeremonyMsgPairHello)

	if got := m.State(); got != PairingCeremonyServerIdle {
		t.Fatalf("expected Idle after invalid token, got %s", got)
	}
}

func TestMachineCodeRejection(t *testing.T) {
	p := PairingCeremony()
	m, err := NewMachine(p, "server")
	if err != nil {
		t.Fatalf("NewMachine: %v", err)
	}

	m.RegisterGuard(PairingCeremonyGuardTokenValid, func(any) bool { return true })
	m.RegisterGuard(PairingCeremonyGuardTokenInvalid, func(any) bool { return false })
	m.RegisterGuard(PairingCeremonyGuardCodeCorrect, func(any) bool { return false })
	m.RegisterGuard(PairingCeremonyGuardCodeWrong, func(any) bool { return true })
	noopActions(m, PairingCeremonyActionGenerateToken, PairingCeremonyActionRegisterRelay, PairingCeremonyActionDeriveSecret)

	mustHandle(t, m, PairingCeremonyMsgPairBegin)
	mustStep(t, m) // -> RegisterRelay
	mustStep(t, m) // -> WaitingForClient
	mustHandle(t, m, PairingCeremonyMsgPairHello)
	mustStep(t, m) // -> SendAck
	mustStep(t, m) // -> WaitingForCode
	mustHandle(t, m, PairingCeremonyMsgCodeSubmit)
	mustStep(t, m) // -> Idle (wrong code)

	if got := m.State(); got != PairingCeremonyServerIdle {
		t.Fatalf("expected Idle after wrong code, got %s", got)
	}
}

func TestMachineRejectsInvalidMessage(t *testing.T) {
	p := PairingCeremony()
	m, err := NewMachine(p, "server")
	if err != nil {
		t.Fatalf("NewMachine: %v", err)
	}

	if _, err := m.HandleMessage(PairingCeremonyMsgAuthRequest, nil); err == nil {
		t.Fatal("expected error for invalid message in Idle state")
	}
}

func TestIOSMachineHappyPath(t *testing.T) {
	p := PairingCeremony()
	m, err := NewMachine(p, "ios")
	if err != nil {
		t.Fatalf("NewMachine: %v", err)
	}

	noopActions(m, PairingCeremonyActionSendPairHello, PairingCeremonyActionDeriveSecret, PairingCeremonyActionStoreSecret)

	mustStep(t, m) // -> ScanQR
	mustStep(t, m) // -> ConnectRelay
	mustStep(t, m) // -> GenKeyPair
	mustStep(t, m) // -> WaitAck
	mustHandle(t, m, PairingCeremonyMsgPairHelloAck)
	mustHandle(t, m, PairingCeremonyMsgPairConfirm)

	if got := m.State(); got != PairingCeremonyAppShowCode {
		t.Fatalf("expected ShowCode, got %s", got)
	}

	mustStep(t, m) // -> WaitPairComplete
	mustHandle(t, m, PairingCeremonyMsgPairComplete)

	if got := m.State(); got != PairingCeremonyAppPaired {
		t.Fatalf("expected Paired, got %s", got)
	}

	mustStep(t, m) // -> Reconnect
	mustStep(t, m) // -> SendAuth
	mustHandle(t, m, PairingCeremonyMsgAuthOk)

	if got := m.State(); got != PairingCeremonyAppSessionActive {
		t.Fatalf("expected SessionActive, got %s", got)
	}
}

func TestCLIMachineHappyPath(t *testing.T) {
	p := PairingCeremony()
	m, err := NewMachine(p, "cli")
	if err != nil {
		t.Fatalf("NewMachine: %v", err)
	}

	mustStep(t, m)
	mustStep(t, m)
	mustHandle(t, m, PairingCeremonyMsgTokenResponse)
	mustHandle(t, m, PairingCeremonyMsgWaitingForCode)
	mustStep(t, m)
	mustHandle(t, m, PairingCeremonyMsgPairStatus)

	if got := m.State(); got != PairingCeremonyCLIDone {
		t.Fatalf("expected Done, got %s", got)
	}
}

func TestExportTLA(t *testing.T) {
	p := PairingCeremony()
	var buf bytes.Buffer
	if err := p.ExportTLA(&buf); err != nil {
		t.Fatalf("ExportTLA: %v", err)
	}

	spec := buf.String()

	checks := map[string][]string{
		"structural": {
			"MODULE PairingCeremony",
			"EXTENDS Integers",
			"server_state", "ios_state", "cli_state",
			"received_", "Init ==", "Next ==", "Spec ==",
		},
		"variables": {
			"VARIABLES",
			"vars ==",
		},
		"operators": {
			"DeriveKey(a, b)", "DeriveCode(a, b)",
		},
		"guards_inlined": {
			`.token \in active_tokens`,
			`.token \notin active_tokens`,
			"received_code = server_code",
			`received_device_id \in paired_devices`,
		},
		"sends": {
			"received_", "[type |->",
		},
		"transitions": {
			"_state' =",
			"UNCHANGED",
		},
		"properties": {
			"NoTokenReuse", "MitMDetectedByCodeMismatch", "MitMPrevented",
			"AuthRequiresCompletedPairing", "NoNonceReuse",
			"WrongCodeDoesNotPair", "DeviceSecretSecrecy",
			"HonestPairingCompletes",
		},
	}

	for category, items := range checks {
		for _, check := range items {
			if !strings.Contains(spec, check) {
				t.Errorf("TLA+ spec missing %s: %q", category, check)
			}
		}
	}
}

func TestExportTLAKeyBoundCodes(t *testing.T) {
	p := PairingCeremony()
	var buf bytes.Buffer
	if err := p.ExportTLA(&buf); err != nil {
		t.Fatalf("ExportTLA: %v", err)
	}

	spec := buf.String()

	// Server computes code from its view of pubkeys (pure TLA+ uses ' =).
	if !strings.Contains(spec, `server_code' = DeriveCode("server_pub"`) {
		t.Error("TLA+ spec missing server-side DeriveCode")
	}

	// iOS computes code from its view of pubkeys.
	if !strings.Contains(spec, `ios_code' = DeriveCode(received_server_pub`) {
		t.Error("TLA+ spec missing ios-side DeriveCode")
	}

	// CLI sends ios_code (what the user read from the phone).
	if !strings.Contains(spec, "code |-> ios_code") {
		t.Error("TLA+ spec missing CLI sending ios_code")
	}
}

func TestValidateDetectsDuplicateActor(t *testing.T) {
	p := &Protocol{
		Name: "Bad",
		Actors: []Actor{
			{Name: "a", Initial: "S0"},
			{Name: "a", Initial: "S1"},
		},
	}
	if err := p.Validate(); err == nil {
		t.Fatal("expected error for duplicate actor")
	}
}

func TestValidateDetectsUndeclaredMessage(t *testing.T) {
	p := &Protocol{
		Name: "Bad",
		Actors: []Actor{
			{Name: "a", Initial: "S0", Transitions: []Transition{
				{From: "S0", To: "S1", On: Recv("nonexistent")},
			}},
		},
	}
	if err := p.Validate(); err == nil {
		t.Fatal("expected error for undeclared message")
	}
}

func TestValidateDetectsUndefinedGuard(t *testing.T) {
	p := &Protocol{
		Name: "Bad",
		Actors: []Actor{
			{Name: "a", Initial: "S0", Transitions: []Transition{
				{From: "S0", To: "S1", On: Internal("x"), Guard: "missing_guard"},
			}},
		},
	}
	if err := p.Validate(); err == nil {
		t.Fatal("expected error for undefined guard")
	}
}

func TestValidateDetectsWrongSender(t *testing.T) {
	p := &Protocol{
		Name: "Bad",
		Actors: []Actor{
			{Name: "a", Initial: "S0", Transitions: []Transition{
				{From: "S0", To: "S1", On: Internal("x"), Sends: []Send{
					{To: "b", Msg: "msg1"},
				}},
			}},
			{Name: "b", Initial: "S0"},
		},
		Messages: []Message{
			{Type: "msg1", From: "b", To: "a"},
		},
	}
	if err := p.Validate(); err == nil {
		t.Fatal("expected error for wrong sender")
	}
}

// Machine error path tests.

func TestNewMachineUnknownActor(t *testing.T) {
	p := PairingCeremony()
	if _, err := NewMachine(p, "nonexistent"); err == nil {
		t.Fatal("expected error for unknown actor")
	}
}

func TestStepNoInternalTransition(t *testing.T) {
	// A machine in ServerPaired has no internal transitions from that state.
	p := PairingCeremony()
	m, err := NewMachine(p, "server")
	if err != nil {
		t.Fatal(err)
	}
	// Advance machine to Paired state via the full happy path.
	m.RegisterGuard(PairingCeremonyGuardTokenValid, func(any) bool { return true })
	m.RegisterGuard(PairingCeremonyGuardTokenInvalid, func(any) bool { return false })
	m.RegisterGuard(PairingCeremonyGuardCodeCorrect, func(any) bool { return true })
	m.RegisterGuard(PairingCeremonyGuardCodeWrong, func(any) bool { return false })
	m.RegisterGuard(PairingCeremonyGuardDeviceKnown, func(any) bool { return true })
	m.RegisterGuard(PairingCeremonyGuardDeviceUnknown, func(any) bool { return false })
	m.RegisterGuard(PairingCeremonyGuardNonceFresh, func(any) bool { return true })
	noopActions(m,
		PairingCeremonyActionGenerateToken, PairingCeremonyActionRegisterRelay, PairingCeremonyActionDeriveSecret,
		PairingCeremonyActionStoreDevice, PairingCeremonyActionVerifyDevice,
	)
	mustHandle(t, m, PairingCeremonyMsgPairBegin)
	mustStep(t, m)
	mustStep(t, m)
	mustHandle(t, m, PairingCeremonyMsgPairHello)
	mustStep(t, m)
	mustStep(t, m)
	mustHandle(t, m, PairingCeremonyMsgCodeSubmit)
	mustStep(t, m)
	mustStep(t, m)
	// Now in ServerPaired — no internal transition.
	if _, err := m.Step("", nil); err == nil {
		t.Fatal("expected error for Step with no internal transition")
	}
}

func TestTryTransitionsUnregisteredGuard(t *testing.T) {
	// Build a minimal protocol with a guarded internal transition.
	p := &Protocol{
		Name: "Test",
		Actors: []Actor{
			{
				Name:    "a",
				Initial: "S0",
				Transitions: []Transition{
					{From: "S0", To: "S1", On: Internal("x"), Guard: "myguard"},
				},
			},
		},
		Guards: []GuardDef{
			{ID: "myguard", Expr: "TRUE"},
		},
	}
	m, err := NewMachine(p, "a")
	if err != nil {
		t.Fatal(err)
	}
	// Do NOT register the guard — should get "unregistered guard" error.
	if _, err := m.Step("", nil); err == nil {
		t.Fatal("expected error for unregistered guard")
	}
}

func TestTryTransitionsUnregisteredAction(t *testing.T) {
	p := &Protocol{
		Name: "Test",
		Actors: []Actor{
			{
				Name:    "a",
				Initial: "S0",
				Transitions: []Transition{
					{From: "S0", To: "S1", On: Internal("x"), Do: "myaction"},
				},
			},
		},
	}
	m, err := NewMachine(p, "a")
	if err != nil {
		t.Fatal(err)
	}
	// Do NOT register the action.
	if _, err := m.Step("", nil); err == nil {
		t.Fatal("expected error for unregistered action")
	}
}

func TestTryTransitionsActionFailure(t *testing.T) {
	p := &Protocol{
		Name: "Test",
		Actors: []Actor{
			{
				Name:    "a",
				Initial: "S0",
				Transitions: []Transition{
					{From: "S0", To: "S1", On: Internal("x"), Do: "failaction"},
				},
			},
		},
	}
	m, err := NewMachine(p, "a")
	if err != nil {
		t.Fatal(err)
	}
	m.RegisterAction("failaction", func(any) error {
		return errors.New("action exploded")
	})
	if _, err := m.Step("", nil); err == nil {
		t.Fatal("expected error when action fails")
	}
}

func TestTryTransitionsAllGuardsFail(t *testing.T) {
	p := &Protocol{
		Name: "Test",
		Actors: []Actor{
			{
				Name:    "a",
				Initial: "S0",
				Transitions: []Transition{
					{From: "S0", To: "S1", On: Internal("x"), Guard: "alwaysfalse"},
				},
			},
		},
		Guards: []GuardDef{
			{ID: "alwaysfalse", Expr: "FALSE"},
		},
	}
	m, err := NewMachine(p, "a")
	if err != nil {
		t.Fatal(err)
	}
	m.RegisterGuard("alwaysfalse", func(any) bool { return false })
	if _, err := m.Step("", nil); err == nil {
		t.Fatal("expected error when all guards fail")
	}
}

// Validate error path tests.

func TestValidateUnknownMessageSender(t *testing.T) {
	p := &Protocol{
		Name:   "Bad",
		Actors: []Actor{{Name: "a", Initial: "S0"}},
		Messages: []Message{
			{Type: "m", From: "nobody", To: "a"},
		},
	}
	if err := p.Validate(); err == nil {
		t.Fatal("expected error for unknown message sender")
	}
}

func TestValidateUnknownMessageReceiver(t *testing.T) {
	p := &Protocol{
		Name:   "Bad",
		Actors: []Actor{{Name: "a", Initial: "S0"}},
		Messages: []Message{
			{Type: "m", From: "a", To: "nobody"},
		},
	}
	if err := p.Validate(); err == nil {
		t.Fatal("expected error for unknown message receiver")
	}
}

func TestValidateSendsToUnknownActor(t *testing.T) {
	p := &Protocol{
		Name: "Bad",
		Actors: []Actor{
			{Name: "a", Initial: "S0", Transitions: []Transition{
				{From: "S0", To: "S1", On: Internal("x"), Sends: []Send{
					{To: "nobody", Msg: "msg1"},
				}},
			}},
		},
		Messages: []Message{
			{Type: "msg1", From: "a", To: "a"},
		},
	}
	if err := p.Validate(); err == nil {
		t.Fatal("expected error for send to unknown actor")
	}
}

func TestValidateSendsUndeclaredMessage(t *testing.T) {
	p := &Protocol{
		Name: "Bad",
		Actors: []Actor{
			{Name: "a", Initial: "S0", Transitions: []Transition{
				{From: "S0", To: "S1", On: Internal("x"), Sends: []Send{
					{To: "a", Msg: "undeclared"},
				}},
			}},
		},
	}
	if err := p.Validate(); err == nil {
		t.Fatal("expected error for send of undeclared message")
	}
}

// Code generator tests.

func TestExportKotlinStructure(t *testing.T) {
	p := PairingCeremony()
	var buf bytes.Buffer
	if err := p.ExportKotlin(&buf, "com.example.pigeon"); err != nil {
		t.Fatalf("ExportKotlin: %v", err)
	}
	out := buf.String()
	checks := []string{
		"package com.example.pigeon",
		"MessageType",
		"enum class",
		"transitions",
	}
	for _, want := range checks {
		if !strings.Contains(out, want) {
			t.Errorf("ExportKotlin output missing %q", want)
		}
	}
}

func TestExportTypeScriptStructure(t *testing.T) {
	p := PairingCeremony()
	var buf bytes.Buffer
	if err := p.ExportTypeScript(&buf); err != nil {
		t.Fatalf("ExportTypeScript: %v", err)
	}
	out := buf.String()
	checks := []string{
		"export enum MessageType",
		"transitions",
	}
	for _, want := range checks {
		if !strings.Contains(out, want) {
			t.Errorf("ExportTypeScript output missing %q", want)
		}
	}
}

// Hierarchy tests.

func TestParseHierarchyFromYAML(t *testing.T) {
	yaml := `
name: HierTest
one_shot: true
channel_bound: 1
messages: {}
events:
  ping: 'ping event'
  data_in: 'data in'
  special_in: 'special data in'
commands:
  deliver: 'deliver data'
  deliver_special: 'deliver special'
  pong: 'send pong'
actors:
  server:
    initial: Idle
    states:
      Active:
        children: [Normal, SubActive]
        transitions:
          - {on: data_in, emits: [deliver]}
      SubActive:
        children: [Special]
        transitions:
          - {on: special_in, emits: [deliver_special]}
    transitions:
      - from: Idle
        to: Normal
        on: ping
        emits: [pong]
vars: {}
guards: {}
`
	p, err := ParseYAML([]byte(yaml))
	if err != nil {
		t.Fatalf("ParseYAML: %v", err)
	}

	a := p.Actors[0]

	// Check hierarchy: Active -> [Normal, SubActive], SubActive -> [Special]
	// StateIndex: Active, Normal, SubActive, Special = 4 entries
	if len(a.StateIndex) != 4 {
		t.Fatalf("expected 4 state nodes, got %d", len(a.StateIndex))
	}
	if len(a.Roots) != 1 {
		t.Fatalf("expected 1 root, got %d", len(a.Roots))
	}
	if a.Roots[0].Name != "Active" {
		t.Fatalf("expected root Active, got %s", a.Roots[0].Name)
	}

	// Check flattened transitions.
	flat := a.FlattenedTransitions()
	// Explicit: Idle->Normal on ping (1)
	// Active inherited: data_in on Normal and Special (2)
	// SubActive inherited: special_in on Special (1)
	// Total: 4
	if len(flat) != 4 {
		t.Fatalf("expected 4 flattened transitions, got %d", len(flat))
	}

	// Verify specific expansions.
	found := map[string]bool{}
	for _, tr := range flat {
		key := string(tr.From) + ":" + tr.On.Desc
		found[key] = true
	}
	for _, want := range []string{
		"Idle:ping",
		"Normal:data_in",
		"Special:data_in",
		"Special:special_in",
	} {
		if !found[want] {
			t.Errorf("missing expected transition: %s", want)
		}
	}
}

func TestFlattenedTransitionsNoHierarchy(t *testing.T) {
	a := Actor{
		Name:    "simple",
		Initial: "S0",
		Transitions: []Transition{
			{From: "S0", To: "S1", On: Internal("go")},
		},
	}
	flat := a.FlattenedTransitions()
	if len(flat) != 1 {
		t.Fatalf("expected 1, got %d", len(flat))
	}
}

func TestHierarchyMultipleParentsRejected(t *testing.T) {
	yaml := `
name: BadHierarchy
one_shot: true
channel_bound: 1
messages: {}
events: {}
commands: {}
actors:
  a:
    initial: S0
    states:
      Parent1:
        children: [Child]
      Parent2:
        children: [Child]
    transitions:
      - {from: S0, to: Child, on: go}
vars: {}
guards: {}
`
	_, err := ParseYAML([]byte(yaml))
	if err == nil {
		t.Fatal("expected error for multiple parents")
	}
	if !strings.Contains(err.Error(), "multiple parents") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMachineHierarchyDispatch(t *testing.T) {
	p := &Protocol{
		Name: "HierDispatch",
		Actors: []Actor{{
			Name:    "a",
			Initial: "S1",
			Transitions: []Transition{
				// Leaf-level transition.
				{From: "S1", To: "S2", On: Internal("advance")},
			},
			StateIndex: map[State]*StateNode{
				"Parent": {
					Name: "Parent",
					Transitions: []Transition{
						// Superstate self-loop (will be expanded).
						{On: Internal("tick")},
					},
				},
				"S1": {Name: "S1"},
				"S2": {Name: "S2"},
			},
			Roots: nil, // will set up below
		}},
	}
	// Wire up hierarchy.
	parent := p.Actors[0].StateIndex["Parent"]
	s1 := p.Actors[0].StateIndex["S1"]
	s2 := p.Actors[0].StateIndex["S2"]
	parent.Children = []*StateNode{s1, s2}
	s1.Parent = parent
	s2.Parent = parent
	p.Actors[0].Roots = []*StateNode{parent}

	m, err := NewMachine(p, "a")
	if err != nil {
		t.Fatalf("NewMachine: %v", err)
	}

	// "tick" is inherited — should work from S1.
	if _, err := m.Step("tick", nil); err != nil {
		t.Fatalf("expected tick to work from S1 via hierarchy: %v", err)
	}
	if m.State() != "S1" { // self-loop
		t.Fatalf("expected S1 after tick, got %s", m.State())
	}

	// "advance" is a leaf transition S1->S2.
	if _, err := m.Step("advance", nil); err != nil {
		t.Fatalf("advance failed: %v", err)
	}
	if m.State() != "S2" {
		t.Fatalf("expected S2, got %s", m.State())
	}

	// "tick" should also work from S2.
	if _, err := m.Step("tick", nil); err != nil {
		t.Fatalf("expected tick to work from S2 via hierarchy: %v", err)
	}
}

func TestStateNodeLeafStates(t *testing.T) {
	grandchild := &StateNode{Name: "GC"}
	child1 := &StateNode{Name: "C1", Children: []*StateNode{grandchild}}
	child2 := &StateNode{Name: "C2"}
	root := &StateNode{Name: "R", Children: []*StateNode{child1, child2}}

	leaves := root.LeafStates()
	if len(leaves) != 2 {
		t.Fatalf("expected 2 leaves, got %d", len(leaves))
	}
	if leaves[0].Name != "GC" || leaves[1].Name != "C2" {
		t.Fatalf("unexpected leaves: %v, %v", leaves[0].Name, leaves[1].Name)
	}
}

func TestAncestorChain(t *testing.T) {
	root := &StateNode{Name: "R"}
	child := &StateNode{Name: "C", Parent: root}
	grandchild := &StateNode{Name: "GC", Parent: child}

	chain := grandchild.AncestorChain()
	if len(chain) != 3 {
		t.Fatalf("expected 3, got %d", len(chain))
	}
	if chain[0].Name != "GC" || chain[1].Name != "C" || chain[2].Name != "R" {
		t.Fatal("wrong ancestor chain order")
	}
}

func TestSessionProtocolHierarchy(t *testing.T) {
	// Verify session.yaml loads and validates with the hierarchy.
	p, err := LoadYAML("session.yaml")
	if err != nil {
		t.Fatalf("LoadYAML: %v", err)
	}
	if err := p.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}

	// Backend should have hierarchy.
	var backend *Actor
	for i := range p.Actors {
		if p.Actors[i].Name == "backend" {
			backend = &p.Actors[i]
			break
		}
	}
	if backend == nil {
		t.Fatal("backend actor not found")
	}
	if len(backend.StateIndex) == 0 {
		t.Fatal("backend should have state hierarchy")
	}
	if _, ok := backend.StateIndex["Connected"]; !ok {
		t.Fatal("backend missing Connected superstate")
	}

	// Flattened transitions should include the expanded self-loops.
	flat := backend.FlattenedTransitions()
	// Count app_send transitions — should exist for each leaf under Connected.
	appSendCount := 0
	for _, tr := range flat {
		if tr.On.Desc == "app_send" {
			appSendCount++
		}
	}
	// Connected has children: RelayConnected, LANOffered, LANPath(->LANActive, LANDegraded), RelayBackoff
	// = 5 leaf states
	if appSendCount != 5 {
		t.Fatalf("expected 5 app_send transitions (one per leaf), got %d", appSendCount)
	}
}

// Helpers.

func mustHandle(t *testing.T, m *Machine, msg MsgType) {
	t.Helper()
	if _, err := m.HandleMessage(msg, nil); err != nil {
		t.Fatalf("HandleMessage(%s): %v", msg, err)
	}
}

func mustStep(t *testing.T, m *Machine) {
	t.Helper()
	if _, err := m.Step("", nil); err != nil {
		t.Fatalf("Step from %s: %v", m.State(), err)
	}
}

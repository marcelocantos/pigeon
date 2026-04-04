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

	m.RegisterGuard(GuardTokenValid, func(any) bool { return true })
	m.RegisterGuard(GuardTokenInvalid, func(any) bool { return false })
	m.RegisterGuard(GuardCodeCorrect, func(any) bool { return true })
	m.RegisterGuard(GuardCodeWrong, func(any) bool { return false })
	m.RegisterGuard(GuardDeviceKnown, func(any) bool { return true })
	m.RegisterGuard(GuardDeviceUnknown, func(any) bool { return false })
	m.RegisterGuard(GuardNonceFresh, func(any) bool { return true })

	noopActions(m,
		ActionGenerateToken, ActionRegisterRelay, ActionDeriveSecret,
		ActionStoreDevice, ActionVerifyDevice,
	)

	assertState := func(expected State) {
		t.Helper()
		if got := m.State(); got != expected {
			t.Fatalf("expected state %s, got %s", expected, got)
		}
	}

	assertState(ServerIdle)
	mustHandle(t, m, MsgPairBegin)
	assertState(ServerGenerateToken)
	mustStep(t, m) // -> RegisterRelay
	mustStep(t, m) // -> WaitingForClient
	mustHandle(t, m, MsgPairHello)
	assertState(ServerDeriveSecret)
	mustStep(t, m) // -> SendAck
	mustStep(t, m) // -> WaitingForCode
	mustHandle(t, m, MsgCodeSubmit)
	assertState(ServerValidateCode)
	mustStep(t, m) // -> StorePaired
	mustStep(t, m) // -> Paired
	assertState(ServerPaired)

	// Reconnection.
	mustHandle(t, m, MsgAuthRequest)
	mustStep(t, m) // -> SessionActive
	assertState(ServerSessionActive)
}

func TestMachineTokenRejection(t *testing.T) {
	p := PairingCeremony()
	m, err := NewMachine(p, "server")
	if err != nil {
		t.Fatalf("NewMachine: %v", err)
	}

	m.RegisterGuard(GuardTokenValid, func(any) bool { return false })
	m.RegisterGuard(GuardTokenInvalid, func(any) bool { return true })
	noopActions(m, ActionGenerateToken, ActionRegisterRelay)

	mustHandle(t, m, MsgPairBegin)
	mustStep(t, m) // -> RegisterRelay
	mustStep(t, m) // -> WaitingForClient
	mustHandle(t, m, MsgPairHello)

	if got := m.State(); got != ServerIdle {
		t.Fatalf("expected Idle after invalid token, got %s", got)
	}
}

func TestMachineCodeRejection(t *testing.T) {
	p := PairingCeremony()
	m, err := NewMachine(p, "server")
	if err != nil {
		t.Fatalf("NewMachine: %v", err)
	}

	m.RegisterGuard(GuardTokenValid, func(any) bool { return true })
	m.RegisterGuard(GuardTokenInvalid, func(any) bool { return false })
	m.RegisterGuard(GuardCodeCorrect, func(any) bool { return false })
	m.RegisterGuard(GuardCodeWrong, func(any) bool { return true })
	noopActions(m, ActionGenerateToken, ActionRegisterRelay, ActionDeriveSecret)

	mustHandle(t, m, MsgPairBegin)
	mustStep(t, m) // -> RegisterRelay
	mustStep(t, m) // -> WaitingForClient
	mustHandle(t, m, MsgPairHello)
	mustStep(t, m) // -> SendAck
	mustStep(t, m) // -> WaitingForCode
	mustHandle(t, m, MsgCodeSubmit)
	mustStep(t, m) // -> Idle (wrong code)

	if got := m.State(); got != ServerIdle {
		t.Fatalf("expected Idle after wrong code, got %s", got)
	}
}

func TestMachineRejectsInvalidMessage(t *testing.T) {
	p := PairingCeremony()
	m, err := NewMachine(p, "server")
	if err != nil {
		t.Fatalf("NewMachine: %v", err)
	}

	if _, err := m.HandleMessage(MsgAuthRequest, nil); err == nil {
		t.Fatal("expected error for invalid message in Idle state")
	}
}

func TestIOSMachineHappyPath(t *testing.T) {
	p := PairingCeremony()
	m, err := NewMachine(p, "ios")
	if err != nil {
		t.Fatalf("NewMachine: %v", err)
	}

	noopActions(m, ActionSendPairHello, ActionDeriveSecret, ActionStoreSecret)

	mustStep(t, m) // -> ScanQR
	mustStep(t, m) // -> ConnectRelay
	mustStep(t, m) // -> GenKeyPair
	mustStep(t, m) // -> WaitAck
	mustHandle(t, m, MsgPairHelloAck)
	mustHandle(t, m, MsgPairConfirm)

	if got := m.State(); got != AppShowCode {
		t.Fatalf("expected ShowCode, got %s", got)
	}

	mustStep(t, m) // -> WaitPairComplete
	mustHandle(t, m, MsgPairComplete)

	if got := m.State(); got != AppPaired {
		t.Fatalf("expected Paired, got %s", got)
	}

	mustStep(t, m) // -> Reconnect
	mustStep(t, m) // -> SendAuth
	mustHandle(t, m, MsgAuthOk)

	if got := m.State(); got != AppSessionActive {
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
	mustHandle(t, m, MsgTokenResponse)
	mustHandle(t, m, MsgWaitingForCode)
	mustStep(t, m)
	mustHandle(t, m, MsgPairStatus)

	if got := m.State(); got != CLIDone {
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
	m.RegisterGuard(GuardTokenValid, func(any) bool { return true })
	m.RegisterGuard(GuardTokenInvalid, func(any) bool { return false })
	m.RegisterGuard(GuardCodeCorrect, func(any) bool { return true })
	m.RegisterGuard(GuardCodeWrong, func(any) bool { return false })
	m.RegisterGuard(GuardDeviceKnown, func(any) bool { return true })
	m.RegisterGuard(GuardDeviceUnknown, func(any) bool { return false })
	m.RegisterGuard(GuardNonceFresh, func(any) bool { return true })
	noopActions(m,
		ActionGenerateToken, ActionRegisterRelay, ActionDeriveSecret,
		ActionStoreDevice, ActionVerifyDevice,
	)
	mustHandle(t, m, MsgPairBegin)
	mustStep(t, m)
	mustStep(t, m)
	mustHandle(t, m, MsgPairHello)
	mustStep(t, m)
	mustStep(t, m)
	mustHandle(t, m, MsgCodeSubmit)
	mustStep(t, m)
	mustStep(t, m)
	// Now in ServerPaired — no internal transition.
	if _, err := m.Step(nil); err == nil {
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
	if _, err := m.Step(nil); err == nil {
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
	if _, err := m.Step(nil); err == nil {
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
	if _, err := m.Step(nil); err == nil {
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
	if _, err := m.Step(nil); err == nil {
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
	if err := p.ExportKotlin(&buf, "com.example.tern"); err != nil {
		t.Fatalf("ExportKotlin: %v", err)
	}
	out := buf.String()
	checks := []string{
		"package com.example.tern",
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

// Helpers.

func mustHandle(t *testing.T, m *Machine, msg MsgType) {
	t.Helper()
	if _, err := m.HandleMessage(msg, nil); err != nil {
		t.Fatalf("HandleMessage(%s): %v", msg, err)
	}
}

func mustStep(t *testing.T, m *Machine) {
	t.Helper()
	if _, err := m.Step(nil); err != nil {
		t.Fatalf("Step from %s: %v", m.State(), err)
	}
}

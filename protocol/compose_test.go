// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package protocol

import (
	"bytes"
	"testing"
)

const composeTestYAML = `
name: TransportTest
one_shot: true

messages:
  data:
    from: client
    to: backend
    desc: application data

actors:
  backend:
    machines:
      relay:
        initial: Connecting
        reports: [ready, lost]
        accepts: [quiesce, activate]
        transitions:
          - from: Connecting
            to: Active
            on: connected
          - from: Active
            to: Quiescent
            on: quiesce received
          - from: Quiescent
            to: Active
            on: activate received
          - from: Active
            to: Connecting
            on: disconnected
      lan:
        initial: Idle
        reports: [ready, lost]
        accepts: [quiesce, activate]
        transitions:
          - from: Idle
            to: Discovering
            on: start discovery
          - from: Discovering
            to: Active
            on: peer found
          - from: Active
            to: Idle
            on: peer lost
      session:
        initial: WaitTransport
        accepts: [transport_ready, transport_lost]
        transitions:
          - from: WaitTransport
            to: Ready
            on: transport available
          - from: Ready
            to: WaitTransport
            on: all transports lost
    routes:
      - on: ready
        from: relay
        sends:
          - to: session
            event: transport_ready
      - on: ready
        from: lan
        sends:
          - to: relay
            event: quiesce
          - to: session
            event: transport_ready
      - on: lost
        from: lan
        sends:
          - to: relay
            event: activate

  client:
    initial: Idle
    transitions:
      - from: Idle
        to: Connected
        on: recv data
`

func TestParseComposedActor(t *testing.T) {
	p, err := ParseYAML([]byte(composeTestYAML))
	if err != nil {
		t.Fatalf("ParseYAML: %v", err)
	}

	if err := p.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}

	// Find the backend actor.
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

	if !backend.IsComposed() {
		t.Fatal("backend should be composed")
	}

	if len(backend.Machines) != 3 {
		t.Fatalf("expected 3 machines, got %d", len(backend.Machines))
	}

	// Verify machine names and initial states.
	expected := map[string]State{
		"relay":   "Connecting",
		"lan":     "Idle",
		"session": "WaitTransport",
	}
	for _, m := range backend.Machines {
		want, ok := expected[m.Name]
		if !ok {
			t.Errorf("unexpected machine: %s", m.Name)
			continue
		}
		if m.Initial != want {
			t.Errorf("machine %s: initial = %q, want %q", m.Name, m.Initial, want)
		}
	}

	// Verify relay machine has reports and accepts.
	relay := backend.Machines[0]
	if len(relay.Reports) != 2 {
		t.Errorf("relay reports: got %d, want 2", len(relay.Reports))
	}
	if len(relay.Accepts) != 2 {
		t.Errorf("relay accepts: got %d, want 2", len(relay.Accepts))
	}
	if len(relay.Transitions) != 4 {
		t.Errorf("relay transitions: got %d, want 4", len(relay.Transitions))
	}

	// Verify routes.
	if len(backend.Routes) != 3 {
		t.Fatalf("expected 3 routes, got %d", len(backend.Routes))
	}

	r0 := backend.Routes[0]
	if r0.On != "ready" || r0.From != "relay" {
		t.Errorf("route 0: on=%q from=%q, want ready/relay", r0.On, r0.From)
	}
	if len(r0.Sends) != 1 || r0.Sends[0].To != "session" || r0.Sends[0].Event != "transport_ready" {
		t.Errorf("route 0 sends: %+v", r0.Sends)
	}

	r1 := backend.Routes[1]
	if r1.On != "ready" || r1.From != "lan" {
		t.Errorf("route 1: on=%q from=%q, want ready/lan", r1.On, r1.From)
	}
	if len(r1.Sends) != 2 {
		t.Errorf("route 1: expected 2 sends, got %d", len(r1.Sends))
	}

	// Verify the client actor is flat (not composed).
	var client *Actor
	for i := range p.Actors {
		if p.Actors[i].Name == "client" {
			client = &p.Actors[i]
			break
		}
	}
	if client == nil {
		t.Fatal("client actor not found")
	}
	if client.IsComposed() {
		t.Fatal("client should not be composed")
	}
}

func TestExportGoComposed(t *testing.T) {
	p, err := ParseYAML([]byte(composeTestYAML))
	if err != nil {
		t.Fatalf("ParseYAML: %v", err)
	}

	var buf bytes.Buffer
	if err := p.ExportGo(&buf, "protocol", p.Name); err != nil {
		t.Fatalf("ExportGo: %v", err)
	}

	out := buf.String()

	// Should contain per-sub-machine state constants.
	for _, want := range []string{
		"TransportTestBackendRelayConnecting",
		"TransportTestBackendRelayActive",
		"TransportTestBackendLanIdle",
		"TransportTestBackendLanDiscovering",
		"TransportTestBackendSessionWaitTransport",
		"TransportTestBackendSessionReady",
	} {
		if !bytes.Contains(buf.Bytes(), []byte(want)) {
			t.Errorf("missing state constant %q", want)
		}
	}

	// Should contain sub-machine types.
	for _, want := range []string{
		"type TransportTestBackendRelayMachine struct",
		"type TransportTestBackendLanMachine struct",
		"type TransportTestBackendSessionMachine struct",
		"type TransportTestBackendComposite struct",
	} {
		if !bytes.Contains(buf.Bytes(), []byte(want)) {
			t.Errorf("missing type: %q", want)
		}
	}

	// Should contain Route method.
	if !bytes.Contains(buf.Bytes(), []byte("func (c *TransportTestBackendComposite) Route(")) {
		t.Error("missing Route method")
	}

	// Should still contain flat client actor.
	if !bytes.Contains(buf.Bytes(), []byte("type TransportTestClientMachine struct")) {
		t.Error("missing flat ClientMachine")
	}

	_ = out
}

func TestExportCComposed(t *testing.T) {
	p, err := ParseYAML([]byte(composeTestYAML))
	if err != nil {
		t.Fatalf("ParseYAML: %v", err)
	}

	var hdr bytes.Buffer
	if err := p.ExportCHeader(&hdr); err != nil {
		t.Fatalf("ExportCHeader: %v", err)
	}

	var impl bytes.Buffer
	if err := p.ExportCImpl(&impl); err != nil {
		t.Fatalf("ExportCImpl: %v", err)
	}

	h := hdr.String()
	c := impl.String()

	// State enum types for each sub-machine.
	for _, want := range []string{
		"pigeon_backend_relay_state",
		"pigeon_backend_lan_state",
		"pigeon_backend_session_state",
	} {
		if !bytes.Contains(hdr.Bytes(), []byte(want)) {
			t.Errorf("header: missing state type %q", want)
		}
	}

	// State enum constants.
	for _, want := range []string{
		"PIGEON_BACKEND_RELAY_CONNECTING",
		"PIGEON_BACKEND_RELAY_ACTIVE",
		"PIGEON_BACKEND_LAN_IDLE",
		"PIGEON_BACKEND_LAN_DISCOVERING",
		"PIGEON_BACKEND_SESSION_WAIT_TRANSPORT",
		"PIGEON_BACKEND_SESSION_READY",
	} {
		if !bytes.Contains(hdr.Bytes(), []byte(want)) {
			t.Errorf("header: missing state constant %q", want)
		}
	}

	// Per-sub-machine machine struct types.
	for _, want := range []string{
		"pigeon_backend_relay_machine",
		"pigeon_backend_lan_machine",
		"pigeon_backend_session_machine",
	} {
		if !bytes.Contains(hdr.Bytes(), []byte(want)) {
			t.Errorf("header: missing machine struct %q", want)
		}
	}

	// Composite struct and functions.
	for _, want := range []string{
		"pigeon_backend_composite",
		"pigeon_backend_composite_init",
		"pigeon_backend_route",
	} {
		if !bytes.Contains(hdr.Bytes(), []byte(want)) {
			t.Errorf("header: missing composite declaration %q", want)
		}
	}

	// Flat client actor still present.
	if !bytes.Contains(hdr.Bytes(), []byte("pigeon_client_machine")) {
		t.Error("header: missing flat client machine struct")
	}

	// Implementation: per-sub-machine init/handle_message/step.
	for _, want := range []string{
		"pigeon_backend_relay_machine_init",
		"pigeon_backend_relay_handle_message",
		"pigeon_backend_relay_step",
		"pigeon_backend_lan_machine_init",
		"pigeon_backend_session_machine_init",
		"pigeon_backend_composite_init",
		"pigeon_backend_route",
	} {
		if !bytes.Contains(impl.Bytes(), []byte(want)) {
			t.Errorf("impl: missing function %q", want)
		}
	}

	_ = h
	_ = c
}

func TestExportSwiftComposed(t *testing.T) {
	p, err := ParseYAML([]byte(composeTestYAML))
	if err != nil {
		t.Fatalf("ParseYAML: %v", err)
	}

	var buf bytes.Buffer
	if err := p.ExportSwift(&buf); err != nil {
		t.Fatalf("ExportSwift: %v", err)
	}

	// Per-sub-machine state enums.
	for _, want := range []string{
		"public enum TransportTestBackendRelayState: String, Sendable",
		"public enum TransportTestBackendLanState: String, Sendable",
		"public enum TransportTestBackendSessionState: String, Sendable",
	} {
		if !bytes.Contains(buf.Bytes(), []byte(want)) {
			t.Errorf("missing state enum %q", want)
		}
	}

	// Per-sub-machine machine classes.
	for _, want := range []string{
		"public final class TransportTestBackendRelayMachine",
		"public final class TransportTestBackendLanMachine",
		"public final class TransportTestBackendSessionMachine",
	} {
		if !bytes.Contains(buf.Bytes(), []byte(want)) {
			t.Errorf("missing machine class %q", want)
		}
	}

	// Composite class.
	if !bytes.Contains(buf.Bytes(), []byte("public final class TransportTestBackendComposite")) {
		t.Error("missing composite class TransportTestBackendComposite")
	}

	// route(from:event:) method.
	if !bytes.Contains(buf.Bytes(), []byte("public func route(from: String, event:")) {
		t.Error("missing route(from:event:) method")
	}

	// Flat client actor still present.
	if !bytes.Contains(buf.Bytes(), []byte("public final class TransportTestClientMachine")) {
		t.Error("missing flat ClientMachine")
	}
}

func TestValidateComposedBadRoute(t *testing.T) {
	yaml := `
name: BadRoute
messages:
  ping:
    from: a
    to: b
actors:
  a:
    machines:
      m1:
        initial: Idle
        transitions:
          - from: Idle
            to: Idle
            on: tick
    routes:
      - on: ready
        from: nonexistent
        sends:
          - to: m1
            event: tick
  b:
    initial: Idle
    transitions:
      - from: Idle
        to: Done
        on: recv ping
`
	p, err := ParseYAML([]byte(yaml))
	if err != nil {
		t.Fatalf("ParseYAML: %v", err)
	}
	err = p.Validate()
	if err == nil {
		t.Fatal("expected validation error for bad route source")
	}
	t.Logf("got expected error: %v", err)
}

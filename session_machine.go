// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package pigeon

import (
	"fmt"

	"github.com/marcelocantos/pigeon/protocol"
)

// sessionMachine is a thin per-Session handle over the generated
// SessionMachine — exactly one of backend or client is non-nil. It
// exists so api.go's Session struct can hold the post-activation
// machine state behind one interface without depending on which
// actor it was for.
//
// docs/session-protocol.md §Boundaries explicitly puts per-stream
// I/O above the executor, so a Session-level machine wrapper has a
// narrower remit than Conn's executor: it persists activation state,
// drives session_established / disconnect transitions on lifecycle
// boundaries, and routes datagrams (the one I/O surface the spec
// keeps inside the loop). Stream lifecycle (clientAcceptLoop) stays
// above this wrapper.
type sessionMachine struct {
	backend *SessionProtocolBackendMachine // nil on client side
	client  *SessionProtocolClientMachine  // nil on backend side
}

// newBackendSessionMachine wraps the post-activation backend machine
// and advances it through SessionActive → RelayConnected so the
// Session's persisted state mirrors what the spec expects an
// activated backend session to be in.
func newBackendSessionMachine(m *SessionProtocolBackendMachine) (*sessionMachine, error) {
	if m == nil {
		return nil, fmt.Errorf("nil backend machine")
	}
	if _, err := m.HandleEvent(SessionProtocolEventSessionEstablished); err != nil {
		return nil, fmt.Errorf("session_established: %w", err)
	}
	if m.State != SessionProtocolBackendRelayConnected {
		return nil, fmt.Errorf("backend post-activation state %q (want RelayConnected)", m.State)
	}
	return &sessionMachine{backend: m}, nil
}

// newClientSessionMachine mirrors newBackendSessionMachine for the
// client side.
func newClientSessionMachine(m *SessionProtocolClientMachine) (*sessionMachine, error) {
	if m == nil {
		return nil, fmt.Errorf("nil client machine")
	}
	if _, err := m.HandleEvent(SessionProtocolEventSessionEstablished); err != nil {
		return nil, fmt.Errorf("session_established: %w", err)
	}
	if m.State != SessionProtocolClientRelayConnected {
		return nil, fmt.Errorf("client post-activation state %q (want RelayConnected)", m.State)
	}
	return &sessionMachine{client: m}, nil
}

// State returns the current SessionMachine state — used by tests and
// by Session.Close to decide whether disconnect is meaningful (we
// shouldn't fire it twice or after the machine has already left
// RelayConnected for some other reason).
func (sm *sessionMachine) State() State {
	if sm.backend != nil {
		return sm.backend.State
	}
	return sm.client.State
}

// handleEvent forwards to the wrapped machine's HandleEvent.
func (sm *sessionMachine) handleEvent(ev EventID) ([]protocol.CmdID, error) {
	if sm.backend != nil {
		return sm.backend.HandleEvent(ev)
	}
	return sm.client.HandleEvent(ev)
}

// disconnect drives the spec's RelayConnected → Paired transition on
// session shutdown. It's idempotent: firing on a machine that's no
// longer at RelayConnected (already disconnected, or in some future
// LAN-active state) is a no-op rather than an error, since Close can
// race with executor-driven state transitions.
//
// Backend and client RelayConnected constants both project to the
// underlying string "RelayConnected", so a single equality is enough
// to gate either actor.
func (sm *sessionMachine) disconnect() error {
	if sm.State() != SessionProtocolBackendRelayConnected {
		return nil
	}
	_, err := sm.handleEvent(SessionProtocolEventDisconnect)
	return err
}

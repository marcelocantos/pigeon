// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package cwire

// #include "pigeon.h"
// #include <stdlib.h>
//
// static void *cwire_calloc_acceptor(void) { return calloc(1, sizeof(pigeon_acceptor_machine));  }
// static void *cwire_calloc_initiator(void){ return calloc(1, sizeof(pigeon_initiator_machine)); }
import "C"

import (
	"errors"
	"runtime"
	"unsafe"
)

// AcceptorState mirrors pigeon_acceptor_state. State string values match
// the YAML names so they round-trip across Go / C / TLA+.
type AcceptorState int

const (
	AcceptorIdle AcceptorState = iota
	AcceptorGeneratingEphemeral
	AcceptorRegisteringRelay
	AcceptorWaitingForHello
	AcceptorDerivingCode
	AcceptorAwaitingUserConfirm
	AcceptorAwaitingPeerConfirm
	AcceptorPaired
	AcceptorAborted
)

func (s AcceptorState) String() string {
	switch s {
	case AcceptorIdle:
		return "Idle"
	case AcceptorGeneratingEphemeral:
		return "GeneratingEphemeral"
	case AcceptorRegisteringRelay:
		return "RegisteringRelay"
	case AcceptorWaitingForHello:
		return "WaitingForHello"
	case AcceptorDerivingCode:
		return "DerivingCode"
	case AcceptorAwaitingUserConfirm:
		return "AwaitingUserConfirm"
	case AcceptorAwaitingPeerConfirm:
		return "AwaitingPeerConfirm"
	case AcceptorPaired:
		return "Paired"
	case AcceptorAborted:
		return "Aborted"
	}
	return "Unknown"
}

// InitiatorState mirrors pigeon_initiator_state.
type InitiatorState int

const (
	InitiatorIdle InitiatorState = iota
	InitiatorDecodingToken
	InitiatorGeneratingEphemeral
	InitiatorConnectingRelay
	InitiatorAwaitingWelcome
	InitiatorDerivingCode
	InitiatorAwaitingUserConfirm
	InitiatorAwaitingPeerConfirm
	InitiatorPaired
	InitiatorAborted
)

func (s InitiatorState) String() string {
	switch s {
	case InitiatorIdle:
		return "Idle"
	case InitiatorDecodingToken:
		return "DecodingToken"
	case InitiatorGeneratingEphemeral:
		return "GeneratingEphemeral"
	case InitiatorConnectingRelay:
		return "ConnectingRelay"
	case InitiatorAwaitingWelcome:
		return "AwaitingWelcome"
	case InitiatorDerivingCode:
		return "DerivingCode"
	case InitiatorAwaitingUserConfirm:
		return "AwaitingUserConfirm"
	case InitiatorAwaitingPeerConfirm:
		return "AwaitingPeerConfirm"
	case InitiatorPaired:
		return "Paired"
	case InitiatorAborted:
		return "Aborted"
	}
	return "Unknown"
}

// CeremonyEvent is the Go mirror of pairing_ceremony_event_id.
type CeremonyEvent int

const (
	EvPairBegin CeremonyEvent = iota
	EvEphemeralReady
	EvRelayRegistered
	EvCodeReady
	EvUserConfirm
	EvUserCancel
	EvTokenReceived
	EvTokenDecoded
	EvRelayConnected
	EvRecvHello
	EvRecvConfirmToAcceptor
	EvRecvWelcome
	EvRecvConfirmToInitiator
)

// CeremonyMessage mirrors pairing_ceremony_msg_type.
type CeremonyMessage int

const (
	MsgHello CeremonyMessage = iota
	MsgWelcome
	MsgConfirmToInitiator
	MsgConfirmToAcceptor
)

// AcceptorMachine wraps pigeon_acceptor_machine.
type AcceptorMachine struct {
	c *C.pigeon_acceptor_machine
}

func NewAcceptorMachine() *AcceptorMachine {
	m := &AcceptorMachine{c: (*C.pigeon_acceptor_machine)(C.cwire_calloc_acceptor())}
	C.pigeon_acceptor_machine_init(m.c)
	runtime.SetFinalizer(m, func(m *AcceptorMachine) {
		if m.c != nil {
			C.free(unsafe.Pointer(m.c))
			m.c = nil
		}
	})
	return m
}

// State returns the current state.
func (m *AcceptorMachine) State() AcceptorState {
	return AcceptorState(m.c.state)
}

// Step dispatches an internal event.
func (m *AcceptorMachine) Step(ev CeremonyEvent) error {
	rv := C.pigeon_acceptor_step(m.c, C.pairing_ceremony_event_id(ev))
	if rv != 1 {
		return errors.New("cwire: acceptor: spec rejected event in current state")
	}
	return nil
}

// HandleMessage dispatches an inbound wire-message arrival.
func (m *AcceptorMachine) HandleMessage(msg CeremonyMessage) error {
	rv := C.pigeon_acceptor_handle_message(m.c, C.pairing_ceremony_msg_type(msg))
	if rv != 1 {
		return errors.New("cwire: acceptor: spec rejected message in current state")
	}
	return nil
}

// InitiatorMachine wraps pigeon_initiator_machine.
type InitiatorMachine struct {
	c *C.pigeon_initiator_machine
}

func NewInitiatorMachine() *InitiatorMachine {
	m := &InitiatorMachine{c: (*C.pigeon_initiator_machine)(C.cwire_calloc_initiator())}
	C.pigeon_initiator_machine_init(m.c)
	runtime.SetFinalizer(m, func(m *InitiatorMachine) {
		if m.c != nil {
			C.free(unsafe.Pointer(m.c))
			m.c = nil
		}
	})
	return m
}

// State returns the current state.
func (m *InitiatorMachine) State() InitiatorState {
	return InitiatorState(m.c.state)
}

// Step dispatches an internal event.
func (m *InitiatorMachine) Step(ev CeremonyEvent) error {
	rv := C.pigeon_initiator_step(m.c, C.pairing_ceremony_event_id(ev))
	if rv != 1 {
		return errors.New("cwire: initiator: spec rejected event in current state")
	}
	return nil
}

// HandleMessage dispatches an inbound wire-message arrival.
func (m *InitiatorMachine) HandleMessage(msg CeremonyMessage) error {
	rv := C.pigeon_initiator_handle_message(m.c, C.pairing_ceremony_msg_type(msg))
	if rv != 1 {
		return errors.New("cwire: initiator: spec rejected message in current state")
	}
	return nil
}

// DeriveConfirmationCode wraps pigeon_derive_confirmation_code: the
// 6-digit MitM-detection code computed from both ephemeral pubkeys.
// pubA and pubB are 32-byte X25519 public keys.
func DeriveConfirmationCode(pubA, pubB []byte) (string, error) {
	if len(pubA) != 32 || len(pubB) != 32 {
		return "", errors.New("cwire: pubkeys must be 32 bytes")
	}
	var out [7]C.char // 6 digits + NUL
	rv := C.pigeon_derive_confirmation_code(
		(*C.uint8_t)(unsafe.Pointer(&pubA[0])),
		(*C.uint8_t)(unsafe.Pointer(&pubB[0])),
		&out[0])
	if rv != 0 {
		return "", errors.New("cwire: derive_confirmation_code failed")
	}
	return C.GoString(&out[0]), nil
}

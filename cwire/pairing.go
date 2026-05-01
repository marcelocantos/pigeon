// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package cwire

// #include "pigeon.h"
// #include <stdlib.h>
// #include <string.h>
//
// static void *cwire_calloc_acceptor(void) { return calloc(1, sizeof(pigeon_acceptor_machine));  }
// static void *cwire_calloc_initiator(void){ return calloc(1, sizeof(pigeon_initiator_machine)); }
//
// // Trampoline defined in cwire_pigeon.c; bridges the C confirm callback to
// // the //export'd cwireGoConfirm in gotransport.go.
// extern int cwire_confirm_trampoline(void *udata, const char *code);
//
// // Defined in cwire_pigeon.c; also referenced by gotransport.go.
// extern void cwire_make_go_transport(void *udata, pigeon_transport *out);
//
// typedef struct cwire_go_udata { uintptr_t handle; } cwire_go_udata;
// extern cwire_go_udata *cwire_alloc_go_udata(uintptr_t handle);
// extern void cwire_free_go_udata(cwire_go_udata *u);
//
// // Wrappers that pass cwire_confirm_trampoline as the confirm_fn so we
// // don't need to convert a C function pointer to *[0]byte on the Go side.
// static int cwire_run_acceptor(
//     const pigeon_transport *t,
//     const uint8_t *priv, const uint8_t *pub,
//     const uint8_t *idpub, const char *iid,
//     void *udata,
//     pigeon_pairing_record *rec, char *code)
// {
//     return pigeon_pair_acceptor(t, priv, pub, idpub, iid,
//                                 cwire_confirm_trampoline, udata, rec, code);
// }
// static int cwire_run_initiator(
//     const pigeon_transport *t,
//     const uint8_t *priv, const uint8_t *pub,
//     const uint8_t *idpub, const char *iid,
//     const uint8_t *accpub, const char *acciid,
//     void *udata,
//     pigeon_pairing_record *rec, char *code)
// {
//     return pigeon_pair_initiator(t, priv, pub, idpub, iid,
//                                  accpub, acciid,
//                                  cwire_confirm_trampoline, udata, rec, code);
// }
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

// PairingRecord mirrors pigeon_pairing_record for return values from the
// C pairing drivers. relay_url is not filled by the C driver (the caller
// must supply the relay URL if needed).
type PairingRecord struct {
	PeerInstanceID string
	LocalPrivKey   []byte // 32 bytes, X25519
	LocalPubKey    []byte // 32 bytes, X25519
	PeerPubKey     []byte // 32 bytes, X25519
}

// RunAcceptorArgs bundles the inputs for RunAcceptor so the signature
// stays flat and extensible without functional-options.
type RunAcceptorArgs struct {
	// Ref provides the Go transport (open_stream / accept_stream /
	// send_on_stream / recv_on_stream). pigeon_pair_acceptor calls
	// accept_stream internally to wait for the initiator's stream.
	Ref          *GoTransportRef
	LocalEphPriv []byte // 32 bytes
	LocalEphPub  []byte // 32 bytes
	IdentityPub  []byte // 32 bytes
	InstanceID   string
	// ConfirmFn is called with the 6-digit code once it is derivable.
	// Return true to confirm, false to abort.
	ConfirmFn func(code string) bool
}

// RunAcceptor drives the C-side acceptor pairing driver. It blocks until
// the initiator connects, the ceremony completes, and ConfirmFn returns.
func RunAcceptor(args *RunAcceptorArgs) (*PairingRecord, string, error) {
	if args == nil {
		return nil, "", errors.New("cwire: RunAcceptor: args is nil")
	}
	if args.Ref == nil || args.Ref.cudata == nil {
		return nil, "", errors.New("cwire: RunAcceptor: nil ref")
	}
	if len(args.LocalEphPriv) != 32 || len(args.LocalEphPub) != 32 || len(args.IdentityPub) != 32 {
		return nil, "", errors.New("cwire: RunAcceptor: keys must be 32 bytes")
	}
	if args.ConfirmFn == nil {
		return nil, "", errors.New("cwire: RunAcceptor: ConfirmFn required")
	}
	if args.InstanceID == "" {
		return nil, "", errors.New("cwire: RunAcceptor: InstanceID required")
	}

	// Build a pigeon_transport vtable pointing at the Go transport.
	var t C.pigeon_transport
	C.cwire_make_go_transport(unsafe.Pointer(args.Ref.cudata), &t)

	cInstanceID := C.CString(args.InstanceID)
	defer C.free(unsafe.Pointer(cInstanceID))

	var rec C.pigeon_pairing_record
	var code [7]C.char

	// ConfirmFn is invoked synchronously from pigeon_pair_acceptor; bridge
	// it through a cgo.Handle so the C callback can call back into Go.
	confirmH := newConfirmHandle(args.ConfirmFn)
	defer confirmH.delete()

	rv := C.cwire_run_acceptor(
		&t,
		(*C.uint8_t)(unsafe.Pointer(&args.LocalEphPriv[0])),
		(*C.uint8_t)(unsafe.Pointer(&args.LocalEphPub[0])),
		(*C.uint8_t)(unsafe.Pointer(&args.IdentityPub[0])),
		cInstanceID,
		confirmH.ptr(),
		&rec,
		&code[0])
	if rv != 0 {
		return nil, "", errors.New("cwire: pigeon_pair_acceptor failed")
	}

	pr := recordFromC(&rec)
	codeStr := C.GoString(&code[0])
	return pr, codeStr, nil
}

// RunInitiatorArgs bundles the inputs for RunInitiator.
type RunInitiatorArgs struct {
	// Ref provides the Go transport. pigeon_pair_initiator calls
	// open_stream internally to open a stream toward the acceptor.
	Ref          *GoTransportRef
	LocalEphPriv []byte
	LocalEphPub  []byte
	IdentityPub  []byte
	InstanceID   string
	// AccEphPub is the acceptor's ephemeral public key decoded from the token.
	AccEphPub   []byte // 32 bytes
	AccInstance string
	ConfirmFn   func(code string) bool
}

// RunInitiator drives the C-side initiator pairing driver, blocking until
// paired or error.
func RunInitiator(args *RunInitiatorArgs) (*PairingRecord, string, error) {
	if args == nil {
		return nil, "", errors.New("cwire: RunInitiator: args is nil")
	}
	if args.Ref == nil || args.Ref.cudata == nil {
		return nil, "", errors.New("cwire: RunInitiator: nil ref")
	}
	if len(args.LocalEphPriv) != 32 || len(args.LocalEphPub) != 32 ||
		len(args.IdentityPub) != 32 || len(args.AccEphPub) != 32 {
		return nil, "", errors.New("cwire: RunInitiator: keys must be 32 bytes")
	}
	if args.ConfirmFn == nil {
		return nil, "", errors.New("cwire: RunInitiator: ConfirmFn required")
	}
	if args.InstanceID == "" || args.AccInstance == "" {
		return nil, "", errors.New("cwire: RunInitiator: InstanceIDs required")
	}

	var t C.pigeon_transport
	C.cwire_make_go_transport(unsafe.Pointer(args.Ref.cudata), &t)

	cInstanceID := C.CString(args.InstanceID)
	defer C.free(unsafe.Pointer(cInstanceID))
	cAccInstance := C.CString(args.AccInstance)
	defer C.free(unsafe.Pointer(cAccInstance))

	var rec C.pigeon_pairing_record
	var code [7]C.char

	confirmH := newConfirmHandle(args.ConfirmFn)
	defer confirmH.delete()

	rv := C.cwire_run_initiator(
		&t,
		(*C.uint8_t)(unsafe.Pointer(&args.LocalEphPriv[0])),
		(*C.uint8_t)(unsafe.Pointer(&args.LocalEphPub[0])),
		(*C.uint8_t)(unsafe.Pointer(&args.IdentityPub[0])),
		cInstanceID,
		(*C.uint8_t)(unsafe.Pointer(&args.AccEphPub[0])),
		cAccInstance,
		confirmH.ptr(),
		&rec,
		&code[0])
	if rv != 0 {
		return nil, "", errors.New("cwire: pigeon_pair_initiator failed")
	}

	pr := recordFromC(&rec)
	codeStr := C.GoString(&code[0])
	return pr, codeStr, nil
}

func recordFromC(rec *C.pigeon_pairing_record) *PairingRecord {
	pr := &PairingRecord{
		PeerInstanceID: C.GoString(&rec.peer_instance_id[0]),
		LocalPrivKey:   make([]byte, 32),
		LocalPubKey:    make([]byte, 32),
		PeerPubKey:     make([]byte, 32),
	}
	copy(pr.LocalPrivKey, (*[32]byte)(unsafe.Pointer(&rec.local_private_key[0]))[:])
	copy(pr.LocalPubKey, (*[32]byte)(unsafe.Pointer(&rec.local_public_key[0]))[:])
	copy(pr.PeerPubKey, (*[32]byte)(unsafe.Pointer(&rec.peer_public_key[0]))[:])
	return pr
}

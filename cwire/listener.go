// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package cwire

// #include <stdlib.h>
// #include <string.h>
// #include <stdint.h>
// #include "pigeon.h"
//
// // Forward-declare the resolve trampoline (defined in cwire_listener.c).
// extern int cwire_resolve_trampoline(void *udata, const char *device_id, void *out_record);
//
// // cwire_go_udata is fully defined in gotransport.go's preamble (which
// // appears in _cgo_export.c because gotransport.go has //export'd functions).
// // This file's preamble is also merged into _cgo_export.c (because
// // listener.go exports cwireGoResolve), so we must NOT redefine the struct.
// // Forward-declare via an opaque pointer alias instead.
// struct cwire_go_udata;
// typedef struct cwire_go_udata cwire_go_udata;
// extern cwire_go_udata *cwire_alloc_go_udata(uintptr_t handle);
// extern void cwire_free_go_udata(cwire_go_udata *u);
// extern void cwire_make_go_transport(void *udata, pigeon_transport *out);
// extern void *cwire_calloc_session(void);
import "C"

import (
	"errors"
	"runtime/cgo"
	"unsafe"
)

// PairingResolver looks up the PairingRecord for a given device ID.
// Return (nil, false) to reject the client.
type PairingResolver func(deviceID string) (*PairingRecord, bool)

// Listener wraps pigeon_listener. It accepts paired clients over a
// registered Go transport and produces a *Session per client.
//
// The transport must already have completed the PIGEON_ROLE_REGISTER_MUX
// greeting — NewListener does NOT perform registration.
//
// Memory: the Listener owns every *Session it returns from Accept/Step.
// Do not call Close on a Listener-returned session; Listener.Close frees
// them all.
type Listener struct {
	c       *C.pigeon_listener
	ref     *GoTransportRef // keeps the Go transport alive
	resolveH *resolveHandle // holds the resolve callback
}

// NewListener constructs a Listener over a registered Go transport.
// resolve is invoked synchronously on the accept thread when a new client
// primary arrives. datagrams declares the named channels available on
// each accepted session (both peers must declare the same map).
func NewListener(
	tr GoTransport,
	instanceID string,
	resolve PairingResolver,
	datagrams []DatagramChannel,
) (*Listener, error) {
	if tr == nil {
		return nil, errors.New("cwire: NewListener: nil transport")
	}
	if resolve == nil {
		return nil, errors.New("cwire: NewListener: nil resolver")
	}
	if len(datagrams) > 16 {
		return nil, errors.New("cwire: NewListener: too many datagram channels")
	}

	ref := NewGoTransportRef(tr)

	var t C.pigeon_transport
	C.cwire_make_go_transport(unsafe.Pointer(ref.cudata), &t)

	// Marshal datagram channel definitions.
	var defs [16]C.pigeon_dgchannel_def
	for i, dc := range datagrams {
		if len(dc.Name) >= 64 {
			ref.Close()
			return nil, errors.New("cwire: NewListener: datagram channel name too long")
		}
		nb := []byte(dc.Name)
		for j := range nb {
			defs[i].name[j] = C.char(nb[j])
		}
		defs[i].channel_id = C.uint64_t(dc.ID)
	}

	var cInstanceID *C.char
	if instanceID != "" {
		cInstanceID = C.CString(instanceID)
		defer C.free(unsafe.Pointer(cInstanceID))
	}

	// Bridge the PairingResolver through a cgo.Handle so the C callback
	// can call back into Go without unsafe.Pointer ↔ uintptr round-trips.
	rh := newResolveHandle(resolve)

	var defsPtr *C.pigeon_dgchannel_def
	if len(datagrams) > 0 {
		defsPtr = &defs[0]
	}

	var cListener *C.pigeon_listener
	rv := C.pigeon_listener_init(
		&cListener,
		&t,
		cInstanceID,
		(*[0]byte)(C.cwire_resolve_trampoline),
		rh.ptr(),
		defsPtr,
		C.size_t(len(datagrams)),
	)
	if rv != 0 {
		rh.delete()
		ref.Close()
		return nil, errors.New("cwire: pigeon_listener_init failed")
	}

	return &Listener{c: cListener, ref: ref, resolveH: rh}, nil
}

// InstanceID returns the listener's relay-assigned instance ID.
func (l *Listener) InstanceID() string {
	if l == nil || l.c == nil {
		return ""
	}
	p := C.pigeon_listener_instance_id(l.c)
	if p == nil {
		return ""
	}
	return C.GoString(p)
}

// Accept blocks until the next paired client completes the activation
// handshake and returns its session. Sub-streams that arrive for
// already-accepted clients during the pump are dispatched into those
// sessions' incoming-stream queues.
//
// Returns (nil, err) on transport failure or listener shutdown.
// The returned session is owned by the Listener — do not call Close on it.
func (l *Listener) Accept() (*Session, error) {
	if l == nil || l.c == nil {
		return nil, errors.New("cwire: Listener.Accept: closed")
	}
	var cSess *C.pigeon_session
	rv := C.pigeon_listener_accept(l.c, &cSess)
	if rv != 0 {
		return nil, errors.New("cwire: pigeon_listener_accept failed")
	}
	return listenerSession(cSess), nil
}

// Step consumes exactly one inbound stream from the transport and
// dispatches it. Returns:
//   - (*Session, nil) when a new client primary completes activation
//   - (nil, nil)      when a sub-stream was dispatched, a malformed header
//                     was dropped, or a primary was rejected
//   - (nil, err)      on transport failure or listener shutdown
//
// The returned session is owned by the Listener — do not call Close on it.
func (l *Listener) Step() (*Session, error) {
	if l == nil || l.c == nil {
		return nil, errors.New("cwire: Listener.Step: closed")
	}
	var cSess *C.pigeon_session
	rv := C.pigeon_listener_step(l.c, &cSess)
	switch rv {
	case 1:
		return listenerSession(cSess), nil
	case 0:
		return nil, nil
	default:
		return nil, errors.New("cwire: pigeon_listener_step failed")
	}
}

// Close tears down the listener and all child sessions. Safe to call
// multiple times. The underlying transport is not closed — the caller
// owns it.
func (l *Listener) Close() error {
	if l == nil {
		return nil
	}
	if l.c != nil {
		C.pigeon_listener_close(l.c)
		l.c = nil
	}
	if l.resolveH != nil {
		l.resolveH.delete()
		l.resolveH = nil
	}
	if l.ref != nil {
		l.ref.Close()
		l.ref = nil
	}
	return nil
}

// AcceptIncomingStream polls the session's incoming-stream queue for
// a buffered sub-stream with the given name. Returns (nil, nil) if none
// is queued yet — this is non-blocking. Application code that needs to
// wait should call Listener.Step/Accept to advance the pump, then retry.
func (s *Session) AcceptIncomingStream(name string) (*Stream, error) {
	if s == nil || s.c == nil {
		return nil, errors.New("cwire: AcceptIncomingStream: nil session")
	}
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	st := &Stream{session: s}
	rv := C.pigeon_session_accept_incoming_stream(s.c, cName, &st.c)
	if rv != 0 {
		return nil, nil // not queued yet; caller should retry after pump
	}
	return st, nil
}

// RunClientActivation drives the client-side of the activation exchange
// on an already-opened stream: sends auth_request{device_id}, reads
// auth_ok, and returns whether the activation was accepted.
// The stream handle must have already been posted to the backend's
// accept queue (i.e. the backend's transport->accept_stream will return it).
//
// This is used by tests that need to simulate a client without depending
// on the parallel cwire.Connect implementation.
func RunClientActivation(ref *GoTransportRef, handle unsafe.Pointer, deviceID string) error {
	if ref == nil || ref.cudata == nil {
		return errors.New("cwire: RunClientActivation: nil ref")
	}
	if handle == nil {
		return errors.New("cwire: RunClientActivation: nil handle")
	}
	if deviceID == "" {
		return errors.New("cwire: RunClientActivation: empty deviceID")
	}

	var t C.pigeon_transport
	C.cwire_make_go_transport(unsafe.Pointer(ref.cudata), &t)

	cDeviceID := C.CString(deviceID)
	defer C.free(unsafe.Pointer(cDeviceID))

	// pigeon_client_machine is only needed as an output; we pass a
	// zero-initialised one and discard it post-activation.
	var machine C.pigeon_client_machine

	rv := C.pigeon_run_client_activation(
		unsafe.Pointer(&t),
		handle,
		cDeviceID,
		unsafe.Pointer(&machine),
		nil, 0,
	)
	if rv != 0 {
		return errors.New("cwire: pigeon_run_client_activation failed (rejected or wire error)")
	}
	return nil
}

// ConnectOnTransport drives the client-side of the activation handshake on
// an already-opened primary stream, then initialises and returns a Session.
// This is a thin wrapper around pigeon_connect_on_transport and is the
// client-side counterpart used to test Listener without depending on the
// parallel cwire.Connect implementation.
//
// ref must already be registered. primaryHandle is the opaque stream
// handle obtained from ref's transport.OpenStream (or wherever the caller
// obtained it). peerInstanceID and deviceID must be non-empty.
// record supplies the pre-shared AEAD keying material; datagrams declares
// the session's datagram channels (both sides must declare the same map).
//
// The returned *Session is heap-allocated and owned by the caller; call
// Close when done.
func ConnectOnTransport(
	ref *GoTransportRef,
	primaryHandle unsafe.Pointer,
	peerInstanceID string,
	deviceID string,
	record *PairingRecord,
	datagrams []DatagramChannel,
) (*Session, error) {
	if ref == nil || ref.cudata == nil {
		return nil, errors.New("cwire: ConnectOnTransport: nil ref")
	}
	if primaryHandle == nil {
		return nil, errors.New("cwire: ConnectOnTransport: nil primary handle")
	}
	if peerInstanceID == "" {
		return nil, errors.New("cwire: ConnectOnTransport: empty peerInstanceID")
	}
	if record == nil {
		return nil, errors.New("cwire: ConnectOnTransport: nil record")
	}
	if deviceID == "" {
		return nil, errors.New("cwire: ConnectOnTransport: empty deviceID")
	}
	if len(datagrams) > 16 {
		return nil, errors.New("cwire: ConnectOnTransport: too many datagram channels")
	}

	var t C.pigeon_transport
	C.cwire_make_go_transport(unsafe.Pointer(ref.cudata), &t)

	cPeerID := C.CString(peerInstanceID)
	defer C.free(unsafe.Pointer(cPeerID))
	cDeviceID := C.CString(deviceID)
	defer C.free(unsafe.Pointer(cDeviceID))

	var crec C.pigeon_pairing_record
	C.memset(unsafe.Pointer(&crec), 0, C.sizeof_pigeon_pairing_record)
	if len(record.PeerInstanceID) > 0 {
		src := []byte(record.PeerInstanceID)
		n := len(src)
		if n >= 64 {
			n = 63
		}
		for i := 0; i < n; i++ {
			crec.peer_instance_id[i] = C.char(src[i])
		}
	}
	if len(record.LocalPrivKey) == 32 {
		for i := 0; i < 32; i++ {
			crec.local_private_key[i] = C.uint8_t(record.LocalPrivKey[i])
		}
	}
	if len(record.LocalPubKey) == 32 {
		for i := 0; i < 32; i++ {
			crec.local_public_key[i] = C.uint8_t(record.LocalPubKey[i])
		}
	}
	if len(record.PeerPubKey) == 32 {
		for i := 0; i < 32; i++ {
			crec.peer_public_key[i] = C.uint8_t(record.PeerPubKey[i])
		}
	}

	var defs [16]C.pigeon_dgchannel_def
	for i, dc := range datagrams {
		if len(dc.Name) >= 64 {
			return nil, errors.New("cwire: ConnectOnTransport: datagram channel name too long")
		}
		nb := []byte(dc.Name)
		for j := range nb {
			defs[i].name[j] = C.char(nb[j])
		}
		defs[i].channel_id = C.uint64_t(dc.ID)
	}

	s := &Session{c: (*C.pigeon_session)(C.cwire_calloc_session())}
	var defsPtr *C.pigeon_dgchannel_def
	if len(datagrams) > 0 {
		defsPtr = &defs[0]
	}

	rv := C.pigeon_connect_on_transport(
		&t,
		(*C.pigeon_stream_handle)(primaryHandle),
		cPeerID,
		cDeviceID,
		&crec,
		defsPtr,
		C.size_t(len(datagrams)),
		s.c,
	)
	if rv != 0 {
		C.free(unsafe.Pointer(s.c))
		return nil, errors.New("cwire: pigeon_connect_on_transport failed")
	}
	s.open = true
	return s, nil
}

// listenerSession wraps a C pigeon_session pointer returned from the
// listener. These sessions are owned by the listener; their Close is a
// no-op so the caller can still defer-close them safely.
func listenerSession(c *C.pigeon_session) *Session {
	if c == nil {
		return nil
	}
	// open=false so Session.Close will not call pigeon_session_close or
	// free the pointer — the listener owns both.
	return &Session{c: c, open: false}
}

// --- Resolve trampoline bridge ---

// resolveHandle wraps a Go PairingResolver so it can be invoked from C
// via cwire_resolve_trampoline.
type resolveHandle struct {
	handle cgo.Handle
	cudata *C.cwire_go_udata
}

func newResolveHandle(fn PairingResolver) *resolveHandle {
	h := cgo.NewHandle(fn)
	cu := C.cwire_alloc_go_udata(C.uintptr_t(h))
	return &resolveHandle{handle: h, cudata: cu}
}

func (r *resolveHandle) ptr() unsafe.Pointer {
	return unsafe.Pointer(r.cudata)
}

func (r *resolveHandle) delete() {
	if r == nil {
		return
	}
	if r.cudata != nil {
		C.cwire_free_go_udata(r.cudata)
		r.cudata = nil
	}
	if r.handle != 0 {
		r.handle.Delete()
		r.handle = 0
	}
}

//export cwireGoResolve
func cwireGoResolve(udata unsafe.Pointer, deviceID *C.char, outRecord unsafe.Pointer) C.int {
	cu := (*C.cwire_go_udata)(udata)
	h := cgo.Handle(cu.handle)
	fn := h.Value().(PairingResolver)

	rec, ok := fn(C.GoString(deviceID))
	if !ok || rec == nil {
		return -1
	}

	// Write the PairingRecord into the C-side pigeon_pairing_record.
	cr := (*C.pigeon_pairing_record)(outRecord)

	// Zero first (callee guarantee: the output struct is value-passed,
	// stack-allocated in the C caller; we write all fields explicitly).
	C.memset(unsafe.Pointer(cr), 0, C.sizeof_pigeon_pairing_record)

	// peer_instance_id (NUL-terminated, 64 bytes)
	if len(rec.PeerInstanceID) > 0 {
		src := []byte(rec.PeerInstanceID)
		n := len(src)
		if n >= 64 {
			n = 63
		}
		for i := 0; i < n; i++ {
			cr.peer_instance_id[i] = C.char(src[i])
		}
	}

	// Keys — only copy if provided (tests may supply zero keys).
	if len(rec.LocalPrivKey) == 32 {
		for i := 0; i < 32; i++ {
			cr.local_private_key[i] = C.uint8_t(rec.LocalPrivKey[i])
		}
	}
	if len(rec.LocalPubKey) == 32 {
		for i := 0; i < 32; i++ {
			cr.local_public_key[i] = C.uint8_t(rec.LocalPubKey[i])
		}
	}
	if len(rec.PeerPubKey) == 32 {
		for i := 0; i < 32; i++ {
			cr.peer_public_key[i] = C.uint8_t(rec.PeerPubKey[i])
		}
	}

	return 0
}

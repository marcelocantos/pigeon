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

// AcceptIncomingStream pulls the next peer-opened sub-stream with the
// given name off the session's own QUIC pipe (T45: each session owns
// its connection, so there is no listener demux to pump). Streams that
// arrive with a different name are buffered for a later matching call.
// Returns (nil, err) on transport failure or shutdown.
func (s *Session) AcceptIncomingStream(name string) (*Stream, error) {
	if s == nil || s.c == nil {
		return nil, errors.New("cwire: AcceptIncomingStream: nil session")
	}
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	st := &Stream{session: s}
	rv := C.pigeon_session_accept_incoming_stream(s.c, cName, &st.c)
	if rv != 0 {
		return nil, errors.New("cwire: pigeon_session_accept_incoming_stream failed")
	}
	return st, nil
}

// RunClientActivation drives the client-side of the activation exchange on
// an already-opened stream: mints a fresh per-session nonce, sends
// auth_request{device_id, nonce}, reads auth_ok, and returns the nonce
// (nonceLen bytes) for DeriveSessionChannel. The stream handle must have
// already been posted to the backend's accept queue (i.e. the backend's
// transport->accept_stream will return it).
//
// This is used by tests that need to simulate a client without depending
// on the parallel cwire.Connect implementation.
func RunClientActivation(ref *GoTransportRef, handle unsafe.Pointer, deviceID, route string) ([]byte, error) {
	if ref == nil || ref.cudata == nil {
		return nil, errors.New("cwire: RunClientActivation: nil ref")
	}
	if handle == nil {
		return nil, errors.New("cwire: RunClientActivation: nil handle")
	}
	if deviceID == "" {
		return nil, errors.New("cwire: RunClientActivation: empty deviceID")
	}

	var t C.pigeon_transport
	C.cwire_make_go_transport(unsafe.Pointer(ref.cudata), &t)

	cDeviceID := C.CString(deviceID)
	defer C.free(unsafe.Pointer(cDeviceID))
	cRoute := C.CString(route)
	defer C.free(unsafe.Pointer(cRoute))

	// pigeon_client_machine is only needed as an output; we pass a
	// zero-initialised one and discard it post-activation.
	var machine C.pigeon_client_machine
	var outNonce [nonceLen]C.uint8_t

	rv := C.pigeon_run_client_activation(
		unsafe.Pointer(&t),
		handle,
		cDeviceID,
		cRoute,
		unsafe.Pointer(&machine),
		nil, 0,
		&outNonce[0],
	)
	if rv != 0 {
		return nil, errors.New("cwire: pigeon_run_client_activation failed (rejected or wire error)")
	}
	return C.GoBytes(unsafe.Pointer(&outNonce[0]), C.int(nonceLen)), nil
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

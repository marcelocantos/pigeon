// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package cwire

// #include <stdlib.h>
// #include <string.h>
// #include "pigeon.h"
//
// // Defined in cwire_pigeon.c.
// extern void cwire_make_go_transport(void *udata, pigeon_transport *out);
// extern void *cwire_calloc_session(void);
//
// // cwire_go_udata is defined in gotransport.go's preamble; forward-declare
// // here so we can use pointer-to it without redefining the struct.
// struct cwire_go_udata;
// typedef struct cwire_go_udata cwire_go_udata;
// extern cwire_go_udata *cwire_alloc_go_udata(uintptr_t handle);
// extern void cwire_free_go_udata(cwire_go_udata *u);
//
// // cwire_resolve_trampoline is defined in cwire_listener.c (bridges the
// // pigeon_resolve_device_fn callback to cwireGoResolve in listener.go).
// extern int cwire_resolve_trampoline(void *udata, const char *device_id,
//                                     void *out_record);
//
// static int cwire_run_backend_activation(
//     const pigeon_transport *t,
//     pigeon_stream_handle *stream,
//     void *resolve_udata,
//     pigeon_backend_machine *out_machine,
//     char *out_device_id, size_t out_device_id_cap,
//     pigeon_pairing_record *out_record,
//     uint8_t *out_nonce)
// {
//     return pigeon_run_backend_activation(t, stream,
//                                          cwire_resolve_trampoline,
//                                          resolve_udata,
//                                          out_machine,
//                                          out_device_id, out_device_id_cap,
//                                          out_record,
//                                          out_nonce);
// }
import "C"

import (
	"errors"
	"fmt"
	"unsafe"
)

// StreamHandle is an opaque stream pointer returned by GoTransport.OpenStream.
// It wraps the C-side pigeon_stream_handle* as an unsafe.Pointer.
type StreamHandle = unsafe.Pointer

// nonceLen is the per-session activation nonce length. Must equal
// PIGEON_AUTH_NONCE_LEN in c/include/pigeon/activation.h. The nonce is
// folded into the session-key HKDF info so concurrent sessions under one
// PairingRecord derive distinct keys (🎯T44.1).
const nonceLen = 16

// ConnectArgs bundles inputs for Connect so the signature stays flat and
// extensible without functional options.
type ConnectArgs struct {
	// Ref is the cgo handle for the Go transport. The transport must
	// remain valid for the lifetime of the returned Session.
	Ref *GoTransportRef

	// Primary is the already-opened primary stream handle (obtained via
	// tr.OpenStream()). pigeon_connect_on_transport writes the empty-name
	// stream header on this handle and runs client activation over it.
	// In activation mode the handle is consumed by the handshake; in
	// pairing mode the caller drives the ceremony via Session.Primary.
	Primary StreamHandle

	// PeerInstanceID is the relay instance ID of the backend peer.
	PeerInstanceID string

	// DeviceID identifies this client to the backend (activation mode).
	// Must be empty when Record is nil (pairing mode).
	DeviceID string

	// Record is the PairingRecord used to derive the AEAD channel keys.
	// Pass nil to enter pairing mode (activation is skipped; the session's
	// channel stays unestablished until the pairing ceremony completes).
	Record *PairingRecord

	// Datagrams lists the pre-agreed name→id datagram channels. Both
	// peers must declare an identical list.
	Datagrams []DatagramChannel
}

// Connect runs the client-side activation handshake (or skips it in pairing
// mode) on top of a Go-implemented transport. The caller has already opened
// the primary stream via tr.OpenStream() and passes the resulting handle as
// args.Primary.
//
// Pass args.Record = nil to enter pairing mode; the caller then drives the
// pairing ceremony over the returned session's primary stream and re-derives
// the channel from the resulting PairingRecord.
func Connect(args *ConnectArgs) (*Session, error) {
	if args == nil {
		return nil, errors.New("cwire: Connect: args is nil")
	}
	if args.Ref == nil || args.Ref.cudata == nil {
		return nil, errors.New("cwire: Connect: nil GoTransportRef")
	}
	if args.Primary == nil {
		return nil, errors.New("cwire: Connect: nil primary stream handle")
	}
	if args.PeerInstanceID == "" {
		return nil, errors.New("cwire: Connect: PeerInstanceID required")
	}
	if args.Record != nil && args.DeviceID == "" {
		return nil, errors.New("cwire: Connect: DeviceID required in activation mode")
	}
	if args.Record == nil && args.DeviceID != "" {
		return nil, errors.New("cwire: Connect: DeviceID must be empty in pairing mode")
	}
	if len(args.Datagrams) > 16 {
		return nil, errors.New("cwire: Connect: too many datagram channels")
	}

	var t C.pigeon_transport
	C.cwire_make_go_transport(unsafe.Pointer(args.Ref.cudata), &t)

	cPeerIID := C.CString(args.PeerInstanceID)
	defer C.free(unsafe.Pointer(cPeerIID))

	var cDeviceID *C.char
	if args.DeviceID != "" {
		cDeviceID = C.CString(args.DeviceID)
		defer C.free(unsafe.Pointer(cDeviceID))
	}

	// Marshal the Go PairingRecord to a C struct if present.
	var cRec C.pigeon_pairing_record
	var cRecPtr *C.pigeon_pairing_record
	if args.Record != nil {
		r := args.Record
		if len(r.LocalPrivKey) != 32 || len(r.LocalPubKey) != 32 || len(r.PeerPubKey) != 32 {
			return nil, errors.New("cwire: Connect: PairingRecord keys must be 32 bytes")
		}
		iid := r.PeerInstanceID
		if len(iid) >= 64 {
			return nil, errors.New("cwire: Connect: PeerInstanceID too long in PairingRecord")
		}
		for i := range iid {
			cRec.peer_instance_id[i] = C.char(iid[i])
		}
		cRec.peer_instance_id[len(iid)] = 0
		copy((*[32]byte)(unsafe.Pointer(&cRec.local_private_key[0]))[:], r.LocalPrivKey)
		copy((*[32]byte)(unsafe.Pointer(&cRec.local_public_key[0]))[:], r.LocalPubKey)
		copy((*[32]byte)(unsafe.Pointer(&cRec.peer_public_key[0]))[:], r.PeerPubKey)
		cRecPtr = &cRec
	}

	// Marshal the datagram channel definitions.
	var defs [16]C.pigeon_dgchannel_def
	for i, dc := range args.Datagrams {
		if len(dc.Name) >= 64 {
			return nil, errors.New("cwire: Connect: datagram channel name too long")
		}
		for j := range dc.Name {
			defs[i].name[j] = C.char(dc.Name[j])
		}
		defs[i].channel_id = C.uint64_t(dc.ID)
	}
	var defsPtr *C.pigeon_dgchannel_def
	if len(args.Datagrams) > 0 {
		defsPtr = &defs[0]
	}

	s := &Session{c: (*C.pigeon_session)(C.cwire_calloc_session())}
	rv := C.pigeon_connect_on_transport(
		&t,
		(*C.pigeon_stream_handle)(args.Primary),
		cPeerIID,
		cDeviceID,
		cRecPtr,
		defsPtr, C.size_t(len(args.Datagrams)),
		s.c,
	)
	if rv != 0 {
		C.free(unsafe.Pointer(s.c))
		return nil, errors.New("cwire: pigeon_connect_on_transport failed")
	}
	s.open = true
	return s, nil
}

// BackendActivationResult carries the outputs of RunBackendActivation.
type BackendActivationResult struct {
	// DeviceID is the client device ID decoded from the auth_request.
	// Populated even when Accepted is false (device looked up but
	// rejected), but empty on wire-decode failure.
	DeviceID string

	// Accepted is true when the client was looked up and accepted.
	// When false, the backend has already sent the reject reply.
	Accepted bool

	// Record is the resolved PairingRecord for the accepted client.
	// Nil when Accepted is false.
	Record *PairingRecord

	// Nonce is the per-session nonce decoded from the client's
	// auth_request (nonceLen bytes). Pass it to DeriveSessionChannel so
	// the backend derives the same per-session keys as the client.
	// Populated only when Accepted is true.
	Nonce []byte
}

// RunBackendActivationArgs bundles inputs for RunBackendActivation.
type RunBackendActivationArgs struct {
	// Ref is the cgo handle for the backend's Go transport.
	Ref *GoTransportRef

	// Stream is the opaque handle for the primary (already accepted) stream.
	Stream StreamHandle

	// ResolveFn looks up a client's PairingRecord by device ID.
	// Return (record, true) to accept, (nil, false) to reject.
	ResolveFn func(deviceID string) (*PairingRecord, bool)

	// SkipPrimaryHeader, when true, reads and discards one leading
	// message on the primary stream before running the activation
	// handshake.
	//
	// Under T45 activation mode no longer writes a primary stream header
	// (the client's auth_request is the first message on the bridged
	// pipe), so this should normally be left false. It is retained for
	// callers that interpose a leading framing message of their own.
	SkipPrimaryHeader bool
}

// RunBackendActivation drives the backend's per-client activation handshake:
// reads auth_request from the primary stream, resolves the device ID via
// ResolveFn, and writes the auth_ok reply. It wraps pigeon_run_backend_activation.
//
// Returns BackendActivationResult. The tri-valued C return (0 / 1 / -1) maps
// to: 0 → Accepted=true; 1 → Accepted=false, DeviceID populated; -1 → error.
func RunBackendActivation(args *RunBackendActivationArgs) (*BackendActivationResult, error) {
	if args == nil {
		return nil, errors.New("cwire: RunBackendActivation: args is nil")
	}
	if args.Ref == nil || args.Ref.cudata == nil {
		return nil, errors.New("cwire: RunBackendActivation: nil GoTransportRef")
	}
	if args.Stream == nil {
		return nil, errors.New("cwire: RunBackendActivation: nil stream handle")
	}
	if args.ResolveFn == nil {
		return nil, errors.New("cwire: RunBackendActivation: ResolveFn required")
	}

	t := args.Ref.handle.Value().(GoTransport)

	if args.SkipPrimaryHeader {
		// Drain the empty-name primary stream header that
		// pigeon_connect_on_transport writes before the auth_request.
		// pigeon_run_backend_activation expects to read auth_request
		// as the first message; the production listener strips the header
		// externally (pigeon_listener_step → read_backend_header).
		if _, err := t.RecvOnStream(args.Stream); err != nil {
			return nil, fmt.Errorf("cwire: RunBackendActivation: read primary header: %w", err)
		}
	}

	var ct C.pigeon_transport
	C.cwire_make_go_transport(unsafe.Pointer(args.Ref.cudata), &ct)

	resolveH := newResolveHandle(PairingResolver(args.ResolveFn))
	defer resolveH.delete()

	var machine C.pigeon_backend_machine
	C.pigeon_backend_machine_init(&machine)

	var outDeviceID [129]C.char
	var outRec C.pigeon_pairing_record
	var outNonce [nonceLen]C.uint8_t

	rv := C.cwire_run_backend_activation(
		&ct,
		(*C.pigeon_stream_handle)(args.Stream),
		resolveH.ptr(),
		&machine,
		&outDeviceID[0], C.size_t(len(outDeviceID)),
		&outRec,
		&outNonce[0],
	)

	deviceID := C.GoString(&outDeviceID[0])

	switch rv {
	case 0:
		return &BackendActivationResult{
			DeviceID: deviceID,
			Accepted: true,
			Record:   recordFromC(&outRec),
			Nonce:    C.GoBytes(unsafe.Pointer(&outNonce[0]), C.int(nonceLen)),
		}, nil
	case 1:
		return &BackendActivationResult{
			DeviceID: deviceID,
			Accepted: false,
		}, nil
	default:
		return nil, errors.New("cwire: pigeon_run_backend_activation failed")
	}
}

// Keypair holds an X25519 private/public key pair.
type Keypair struct {
	PrivKey []byte // 32 bytes
	PubKey  []byte // 32 bytes
}

// GenerateKeypair generates a fresh X25519 key pair via libsodium.
func GenerateKeypair() (*Keypair, error) {
	var kp C.pigeon_keypair
	if C.pigeon_generate_keypair(&kp) != 0 {
		return nil, errors.New("cwire: GenerateKeypair failed")
	}
	out := &Keypair{
		PrivKey: make([]byte, 32),
		PubKey:  make([]byte, 32),
	}
	copy(out.PrivKey, (*[32]byte)(unsafe.Pointer(&kp.private_key[0]))[:])
	copy(out.PubKey, (*[32]byte)(unsafe.Pointer(&kp.public_key[0]))[:])
	return out, nil
}

// DeriveSessionChannel derives the AEAD channel from a PairingRecord and the
// per-session nonce exchanged during activation. The HKDF info is
// direction-label || nonce (🎯T44.1), so concurrent/reconnecting sessions
// under one PairingRecord derive distinct keys:
//   - Client (isBackend=false): send="client->backend", recv="backend->client"
//   - Backend (isBackend=true):  send="backend->client", recv="client->backend"
//
// `nonce` must be nonceLen bytes: the client's nonce returned by
// RunClientActivation, or (on the backend) BackendActivationResult.Nonce.
// Both sides must hold the other's public key as PeerPubKey for the ECDH to
// agree, and both must fold in the same nonce. In tests where both sides hold
// the same record, passing opposite isBackend flags with the same nonce makes
// the keys match (send on one side becomes recv on the other).
func DeriveSessionChannel(rec *PairingRecord, isBackend bool, nonce []byte) (*Channel, error) {
	if rec == nil {
		return nil, errors.New("cwire: DeriveSessionChannel: nil record")
	}
	if len(rec.LocalPrivKey) != 32 || len(rec.PeerPubKey) != 32 {
		return nil, errors.New("cwire: DeriveSessionChannel: keys must be 32 bytes")
	}
	if len(nonce) != nonceLen {
		return nil, fmt.Errorf("cwire: DeriveSessionChannel: nonce must be %d bytes", nonceLen)
	}

	var cRec C.pigeon_pairing_record
	copy((*[32]byte)(unsafe.Pointer(&cRec.local_private_key[0]))[:], rec.LocalPrivKey)
	copy((*[32]byte)(unsafe.Pointer(&cRec.peer_public_key[0]))[:], rec.PeerPubKey)

	const clientToBackend = "client->backend"
	const backendToClient = "backend->client"

	var sendLabel, recvLabel string
	if isBackend {
		sendLabel, recvLabel = backendToClient, clientToBackend
	} else {
		sendLabel, recvLabel = clientToBackend, backendToClient
	}
	// info = direction-label || per-session nonce.
	sendInfo := append([]byte(sendLabel), nonce...)
	recvInfo := append([]byte(recvLabel), nonce...)

	var sendKey [32]C.uint8_t
	var recvKey [32]C.uint8_t

	if C.pigeon_derive_session_key(
		&cRec.local_private_key[0],
		&cRec.peer_public_key[0],
		(*C.uint8_t)(unsafe.Pointer(&sendInfo[0])), C.size_t(len(sendInfo)),
		&sendKey[0]) != 0 {
		return nil, errors.New("cwire: DeriveSessionChannel: derive send key failed")
	}
	if C.pigeon_derive_session_key(
		&cRec.local_private_key[0],
		&cRec.peer_public_key[0],
		(*C.uint8_t)(unsafe.Pointer(&recvInfo[0])), C.size_t(len(recvInfo)),
		&recvKey[0]) != 0 {
		return nil, errors.New("cwire: DeriveSessionChannel: derive recv key failed")
	}

	return NewChannel(
		(*[32]byte)(unsafe.Pointer(&sendKey[0]))[:],
		(*[32]byte)(unsafe.Pointer(&recvKey[0]))[:],
		"stream",
	)
}

// cwireGoResolve / resolveHandle / newResolveHandle live in listener.go —
// both slices share the same pigeon_resolve_device_fn bridge, owned by the
// Listener side.

// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

//go:build csdke2e

// Package csdke2e holds the cgo-driven end-to-end tests for the C
// SDK against an in-process Go relay (🎯T32.4).
//
// Build tag: this package only compiles when -tags csdke2e is in
// effect. The default Go build skips it because the link line
// requires the vendored ngtcp2 + quictls static libraries (built by
// `make build-vendor-deps`), which most callers (and `go test ./...`
// inside `make bullseye`) do not have on hand.
//
// Build / test invocation:
//
//	make test-c-ngtcp2          # bullseye-friendly entry point
//
// The Makefile target builds the vendored deps if needed, then runs
//
//	go test -count=1 -tags csdke2e ./c/test/csdke2e/
//
// with the cgo CFLAGS / LDFLAGS picking up the static libs.
package csdke2e

// #cgo CFLAGS: -I${SRCDIR}/../../../dist
// #cgo CFLAGS: -I${SRCDIR}/../../../c/include
// #cgo CFLAGS: -I${SRCDIR}/../../../c/src
// #cgo CFLAGS: -I${SRCDIR}/../../../c/vendor/build/include
// #cgo CFLAGS: -DPIGEON_CRYPTO_LIBSODIUM
// #cgo CFLAGS: -DCSDKE2E_BUILD
// #cgo darwin CFLAGS: -I/opt/homebrew/include
// #cgo darwin LDFLAGS: -L/opt/homebrew/lib -lsodium
// #cgo linux  LDFLAGS: -lsodium
// #cgo LDFLAGS: ${SRCDIR}/../../../c/vendor/build/lib/libngtcp2_crypto_quictls.a
// #cgo LDFLAGS: ${SRCDIR}/../../../c/vendor/build/lib/libngtcp2.a
// #cgo LDFLAGS: ${SRCDIR}/../../../c/vendor/build/lib/libssl.a
// #cgo LDFLAGS: ${SRCDIR}/../../../c/vendor/build/lib/libcrypto.a
// #cgo LDFLAGS: -lpthread
//
// // Forward the resolve-device trampoline that csdke2e_resolve_bridge.go
// // exports — the C resolver function pointer needs a stable name to
// // marshal into the pigeon_resolve_device_fn slot.
// #include <stdlib.h>
// #include <stdint.h>
// #include <stddef.h>
// #include "pigeon.h"
//
// extern int csdke2eResolveTrampoline(void *userdata, char *device_id, void *out_record);
import "C"

import (
	"errors"
	"fmt"
	"runtime/cgo"
	"unsafe"
)

// --- pigeon_register ---

// registerBackend brings up a backend listener against the relay at
// host:port with the given self-assigned instance ID. The resolver
// is invoked synchronously from pigeon_listener_accept whenever a
// new client primary arrives.
func registerBackend(host, port, selfInstanceID string,
	datagrams map[string]uint64,
	resolver func(deviceID string) (*PairingRecord, bool),
) (*listenerHandle, error) {
	cHost := C.CString(host)
	defer C.free(unsafe.Pointer(cHost))
	cPort := C.CString(port)
	defer C.free(unsafe.Pointer(cPort))
	cID := C.CString(selfInstanceID)
	defer C.free(unsafe.Pointer(cID))

	defs, n := dgDefs(datagrams)
	var defsPtr *C.pigeon_dgchannel_def
	if n > 0 {
		defsPtr = (*C.pigeon_dgchannel_def)(unsafe.Pointer(&defs[0]))
	}

	rh := cgo.NewHandle(resolver)
	// Smuggle the handle through an *opaque* heap-allocated box so the
	// resolver trampoline can recover it via a regular C pointer
	// dereference (no unsafe.Pointer ↔ uintptr round-trip in C).
	box := (*resolveBox)(C.malloc(C.size_t(unsafe.Sizeof(resolveBox{}))))
	box.handle = C.uintptr_t(rh)

	var l *C.pigeon_listener
	var assigned [128]C.char
	rv := C.pigeon_register(cHost, cPort, cID /*token*/, nil,
		C.pigeon_resolve_device_fn(C.csdke2eResolveTrampoline),
		unsafe.Pointer(box),
		defsPtr, C.size_t(n),
		&l, &assigned[0], C.size_t(len(assigned)))
	if rv != 0 {
		rh.Delete()
		C.free(unsafe.Pointer(box))
		return nil, errors.New("pigeon_register failed")
	}
	return &listenerHandle{
		c:          l,
		handle:     rh,
		box:        box,
		instanceID: C.GoString(&assigned[0]),
	}, nil
}

type resolveBox struct {
	handle C.uintptr_t
}

type listenerHandle struct {
	c          *C.pigeon_listener
	handle     cgo.Handle
	box        *resolveBox
	instanceID string
	// pendingSessions buffers sessions surfaced by step() while the
	// caller was pumping for a sub-stream on an already-accepted
	// session. Accept() drains this first before calling into C.
	pendingSessions []*sessionHandle
}

func (l *listenerHandle) Close() {
	if l == nil || l.c == nil {
		return
	}
	C.pigeon_listener_close(l.c)
	l.c = nil
	if l.box != nil {
		C.free(unsafe.Pointer(l.box))
		l.box = nil
	}
	if l.handle != 0 {
		l.handle.Delete()
		l.handle = 0
	}
}

// Accept blocks until the next paired client connects and returns
// the resulting session. Sessions surfaced by step() while the caller
// was pumping for sub-streams are returned first (FIFO). The session
// is owned by the listener — do not call sessionClose on it;
// pigeon_listener_close tears down all sessions.
func (l *listenerHandle) Accept() (*sessionHandle, error) {
	if len(l.pendingSessions) > 0 {
		s := l.pendingSessions[0]
		l.pendingSessions = l.pendingSessions[1:]
		return s, nil
	}
	var s *C.pigeon_session
	if rv := C.pigeon_listener_accept(l.c, &s); rv != 0 {
		return nil, errors.New("pigeon_listener_accept failed")
	}
	return &sessionHandle{c: s, ownedByListener: true}, nil
}

// step runs one pigeon_listener_step iteration. If a brand-new
// client primary completes, the session is parked on
// pendingSessions so the next Accept() returns it; the caller of
// step() doesn't need to handle it. Returns nil on a normal step
// (sub-stream dispatched, header dropped, or session parked) and
// an error on transport failure / listener shutdown.
//
// Used by pumpAcceptStream in tests that need to drive the demux
// pump while waiting for a peer-opened sub-stream to land in an
// already-accepted session's incoming-stream queue.
func (l *listenerHandle) step() error {
	var s *C.pigeon_session
	rv := C.pigeon_listener_step(l.c, &s)
	if rv < 0 {
		return errors.New("pigeon_listener_step failed")
	}
	if rv == 1 {
		l.pendingSessions = append(l.pendingSessions,
			&sessionHandle{c: s, ownedByListener: true})
	}
	return nil
}

// --- pigeon_connect ---

func connectClient(host, port, peerInstanceID, deviceID string,
	rec *PairingRecord, datagrams map[string]uint64) (*connectionHandle, error) {
	cHost := C.CString(host)
	defer C.free(unsafe.Pointer(cHost))
	cPort := C.CString(port)
	defer C.free(unsafe.Pointer(cPort))
	cPeer := C.CString(peerInstanceID)
	defer C.free(unsafe.Pointer(cPeer))
	cDev := C.CString(deviceID)
	defer C.free(unsafe.Pointer(cDev))

	defs, n := dgDefs(datagrams)
	var defsPtr *C.pigeon_dgchannel_def
	if n > 0 {
		defsPtr = (*C.pigeon_dgchannel_def)(unsafe.Pointer(&defs[0]))
	}

	cRec := rec.toC()
	var conn *C.pigeon_connection
	rv := C.pigeon_connect(cHost, cPort, cPeer, cDev /*token*/, nil,
		&cRec, defsPtr, C.size_t(n), &conn)
	if rv != 0 {
		return nil, errors.New("pigeon_connect failed")
	}
	return &connectionHandle{c: conn}, nil
}

type connectionHandle struct {
	c *C.pigeon_connection
}

func (c *connectionHandle) Close() {
	if c == nil || c.c == nil {
		return
	}
	C.pigeon_connect_close(c.c)
	c.c = nil
}

func (c *connectionHandle) Session() *sessionHandle {
	return &sessionHandle{c: C.pigeon_connect_session(c.c)}
}

// --- session / stream / datagram ---

type sessionHandle struct {
	c               *C.pigeon_session
	ownedByListener bool
}

// openStream opens a new named stream toward the peer. Mirrors
// Go's Session.OpenStream.
func (s *sessionHandle) openStream(name string) (*streamHandle, error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	st := &streamHandle{}
	if rv := C.pigeon_session_open_stream(s.c, cName, &st.c); rv != 0 {
		return nil, fmt.Errorf("pigeon_session_open_stream(%q) failed", name)
	}
	return st, nil
}

// acceptStream pulls the next peer-opened sub-stream of the given
// name from the listener-fed incoming queue. Returns an error if
// no matching stream is buffered. The session API is single-
// threaded; the typical pattern is to call listener.Accept (which
// pumps the demux) and then this in a loop until the stream
// arrives.
func (s *sessionHandle) acceptStream(name string) (*streamHandle, error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	st := &streamHandle{}
	if rv := C.pigeon_session_accept_incoming_stream(s.c, cName, &st.c); rv != 0 {
		return nil, fmt.Errorf("pigeon_session_accept_incoming_stream(%q): no buffered stream", name)
	}
	return st, nil
}

// datagram looks up the pre-declared datagram channel by name.
func (s *sessionHandle) datagram(name string) (*datagramHandle, error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	d := &datagramHandle{}
	if rv := C.pigeon_session_get_datagram(s.c, cName, &d.c); rv != 0 {
		return nil, fmt.Errorf("pigeon_session_get_datagram(%q) failed", name)
	}
	return d, nil
}

type streamHandle struct {
	c C.pigeon_stream
}

func (s *streamHandle) Send(msg []byte) error {
	var p *C.uint8_t
	if len(msg) > 0 {
		p = (*C.uint8_t)(unsafe.Pointer(&msg[0]))
	}
	if rv := C.pigeon_stream_send(&s.c, p, C.size_t(len(msg))); rv != 0 {
		return errors.New("pigeon_stream_send failed")
	}
	return nil
}

func (s *streamHandle) Recv() ([]byte, error) {
	// PIGEON_MAX_MSG is 1MB; allocate the buffer in Go so the cgo
	// call can stamp directly into it.
	buf := make([]byte, 1<<20)
	n := C.pigeon_stream_recv(&s.c,
		(*C.uint8_t)(unsafe.Pointer(&buf[0])),
		C.size_t(len(buf)))
	if n < 0 {
		return nil, errors.New("pigeon_stream_recv failed")
	}
	return buf[:int(n)], nil
}

type datagramHandle struct {
	c C.pigeon_datagram
}

func (d *datagramHandle) Send(msg []byte) error {
	var p *C.uint8_t
	if len(msg) > 0 {
		p = (*C.uint8_t)(unsafe.Pointer(&msg[0]))
	}
	if rv := C.pigeon_datagram_send(&d.c, p, C.size_t(len(msg))); rv != 0 {
		return errors.New("pigeon_datagram_send failed")
	}
	return nil
}

func (d *datagramHandle) Recv() ([]byte, error) {
	buf := make([]byte, 1<<20)
	n := C.pigeon_datagram_recv(&d.c,
		(*C.uint8_t)(unsafe.Pointer(&buf[0])),
		C.size_t(len(buf)))
	if n < 0 {
		return nil, errors.New("pigeon_datagram_recv failed")
	}
	return buf[:int(n)], nil
}

// --- PairingRecord conversion ---

// PairingRecord is the cgo-side mirror of pigeon_pairing_record.
type PairingRecord struct {
	PeerInstanceID string
	RelayURL       string
	LocalPrivKey   [32]byte
	LocalPubKey    [32]byte
	PeerPubKey     [32]byte
}

func (r *PairingRecord) toC() C.pigeon_pairing_record {
	var c C.pigeon_pairing_record
	for i, b := range []byte(r.PeerInstanceID) {
		if i >= len(c.peer_instance_id)-1 {
			break
		}
		c.peer_instance_id[i] = C.char(b)
	}
	for i, b := range []byte(r.RelayURL) {
		if i >= len(c.relay_url)-1 {
			break
		}
		c.relay_url[i] = C.char(b)
	}
	for i, b := range r.LocalPrivKey {
		c.local_private_key[i] = C.uint8_t(b)
	}
	for i, b := range r.LocalPubKey {
		c.local_public_key[i] = C.uint8_t(b)
	}
	for i, b := range r.PeerPubKey {
		c.peer_public_key[i] = C.uint8_t(b)
	}
	return c
}

// dgDefs builds a fixed-size C array of pigeon_dgchannel_def from a Go
// map. Returns the array and the number of populated entries. The
// array is heap-allocated by Go so the pointer is stable for the
// duration of the C call.
func dgDefs(m map[string]uint64) ([16]C.pigeon_dgchannel_def, int) {
	var defs [16]C.pigeon_dgchannel_def
	i := 0
	for name, id := range m {
		if i >= len(defs) {
			break
		}
		for j, b := range []byte(name) {
			if j >= len(defs[i].name)-1 {
				break
			}
			defs[i].name[j] = C.char(b)
		}
		defs[i].channel_id = C.uint64_t(id)
		i++
	}
	return defs, i
}

// --- Resolver trampoline (called from C) ---

//export csdke2eResolveTrampoline
func csdke2eResolveTrampoline(udata unsafe.Pointer, deviceID *C.char,
	outRecord unsafe.Pointer) C.int {
	box := (*resolveBox)(udata)
	h := cgo.Handle(box.handle)
	fn := h.Value().(func(string) (*PairingRecord, bool))
	rec, ok := fn(C.GoString(deviceID))
	if !ok || rec == nil {
		return -1
	}
	*(*C.pigeon_pairing_record)(outRecord) = rec.toC()
	return 0
}

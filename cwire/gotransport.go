// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package cwire

// #include <stdlib.h>
// #include <string.h>
// #include <stdint.h>
// #include "pigeon.h"
//
// // The trampolines below are //export'd from gotransport.go's Go side.
// // Forward-declare them here so the .c file's cwire_make_go_transport
// // can reference them by name.
// extern int cwireGoOpenStream(void *udata, pigeon_stream_handle **out);
// extern int cwireGoAcceptStream(void *udata, pigeon_stream_handle **out);
// extern int cwireGoSendOnStream(void *udata, pigeon_stream_handle *h, uint8_t *data, size_t len);
// extern int cwireGoRecvOnStream(void *udata, pigeon_stream_handle *h, uint8_t *buf, size_t buf_len, size_t *out_len);
// extern int cwireGoCloseStream(void *udata, pigeon_stream_handle *h);
// extern int cwireGoSendDatagram(void *udata, uint8_t *data, size_t len);
// extern int cwireGoRecvDatagram(void *udata, uint8_t *buf, size_t buf_len, size_t *out_len);
//
// // Defined in cwire_pigeon.c.
// extern void *cwire_calloc_session(void);
// extern void cwire_make_go_transport(void *udata, pigeon_transport *out);
//
// typedef struct cwire_go_udata {
//     uintptr_t handle;
// } cwire_go_udata;
//
// extern cwire_go_udata *cwire_alloc_go_udata(uintptr_t handle);
// extern void cwire_free_go_udata(cwire_go_udata *u);
import "C"

import (
	"errors"
	"runtime/cgo"
	"unsafe"
)

// GoTransport is the Go-side counterpart of pigeon_transport. A
// concrete implementation provides in-process or networked I/O for a
// pigeon_session driven through cgo.
//
// Stream handles are opaque pointers: cwire stores them behind the
// C-side `pigeon_stream_handle*` and hands them back to the Go
// transport on every per-stream call. The transport must mint each
// handle as a real C pointer (e.g. via C.malloc) so the value can
// round-trip through cgo without uintptr conversions that go vet
// flags. The transport is responsible for freeing the underlying
// allocation when CloseStream is called.
type GoTransport interface {
	OpenStream() (handle unsafe.Pointer, err error)
	AcceptStream() (handle unsafe.Pointer, err error)
	SendOnStream(handle unsafe.Pointer, msg []byte) error
	RecvOnStream(handle unsafe.Pointer) ([]byte, error)
	CloseStream(handle unsafe.Pointer) error
	SendDatagram(payload []byte) error
	RecvDatagram() ([]byte, error)
}

// GoTransportRef ties a GoTransport to a cgo.Handle. The C-side
// transport vtable holds a pointer to a small heap-allocated box
// (cwire_go_udata) that carries the cgo.Handle bit-pattern; the
// //export'd trampolines below recover it via a normal C-pointer
// dereference (no unsafe.Pointer ↔ uintptr round-trip, no vet flag).
//
// Lifecycle: the ref must outlive every pigeon_session that copies
// its transport vtable (pigeon_session_init copies the vtable +
// userdata pointer into the session struct). Call Close() once no
// session references the transport.
type GoTransportRef struct {
	handle cgo.Handle
	cudata *C.cwire_go_udata
}

// NewGoTransportRef registers a Go transport with the cgo handle
// table and allocates the C-side userdata box.
func NewGoTransportRef(t GoTransport) *GoTransportRef {
	h := cgo.NewHandle(t)
	cu := C.cwire_alloc_go_udata(C.uintptr_t(h))
	return &GoTransportRef{handle: h, cudata: cu}
}

// Close releases the cgo.Handle and frees the C-side userdata box.
// Safe to call multiple times.
func (r *GoTransportRef) Close() {
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

// NewGoSession constructs a pigeon_session over a GoTransport. The
// supplied ref must remain valid for the lifetime of the returned
// Session. `datagrams` mirrors the loopback constructor. Under T45
// both peers are symmetric — there is no clientTag.
func NewGoSession(ref *GoTransportRef, ch *Channel, datagrams []DatagramChannel) (*Session, error) {
	if ref == nil || ref.cudata == nil {
		return nil, errors.New("cwire: nil GoTransportRef")
	}
	if ch == nil {
		return nil, errors.New("cwire: nil channel")
	}

	var t C.pigeon_transport
	C.cwire_make_go_transport(unsafe.Pointer(ref.cudata), &t)

	if len(datagrams) > 16 {
		return nil, errors.New("cwire: too many datagram channels")
	}
	var defs [16]C.pigeon_dgchannel_def
	for i, dc := range datagrams {
		if len(dc.Name) >= 64 {
			return nil, errors.New("cwire: datagram channel name too long")
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
	rv := C.pigeon_session_init(s.c, &t, ch.c,
		defsPtr, C.size_t(len(datagrams)))
	if rv != 0 {
		C.free(unsafe.Pointer(s.c))
		return nil, errors.New("cwire: pigeon_session_init failed")
	}
	s.open = true
	return s, nil
}

// AcceptStreamFromGo pulls the next peer-opened stream from the Go
// transport, then reads the [varint name-len][name] header so the
// caller has a fully-attached *Stream. Mirrors the loopback helper.
func (s *Session) AcceptStreamFromGo(ref *GoTransportRef) (*Stream, string, error) {
	if ref == nil || ref.cudata == nil {
		return nil, "", errors.New("cwire: nil GoTransportRef")
	}
	t := ref.handle.Value().(GoTransport)
	h, err := t.AcceptStream()
	if err != nil {
		return nil, "", err
	}
	hdr, err := t.RecvOnStream(h)
	if err != nil {
		return nil, "", err
	}
	name, _, derr := DecodeStreamHeader(hdr)
	if derr != nil {
		return nil, "", derr
	}
	st := &Stream{session: s}
	st.c.session = s.c
	st.c.handle = (*C.pigeon_stream_handle)(h)
	if len(name) >= 64 {
		return nil, "", errors.New("cwire: accepted name too long")
	}
	for i := range name {
		st.c.name[i] = C.char(name[i])
	}
	st.c.name[len(name)] = 0
	return st, name, nil
}

// --- //export'd trampolines ---
//
// Each one recovers the GoTransport from the cgo.Handle stashed in
// the C transport's userdata, dispatches to the Go method, and
// translates errors to a -1 return.

//export cwireGoOpenStream
func cwireGoOpenStream(udata unsafe.Pointer, out **C.pigeon_stream_handle) C.int {
	t := recoverGoTransport(udata)
	h, err := t.OpenStream()
	if err != nil {
		return -1
	}
	*out = (*C.pigeon_stream_handle)(h)
	return 0
}

//export cwireGoAcceptStream
func cwireGoAcceptStream(udata unsafe.Pointer, out **C.pigeon_stream_handle) C.int {
	t := recoverGoTransport(udata)
	h, err := t.AcceptStream()
	if err != nil {
		return -1
	}
	*out = (*C.pigeon_stream_handle)(h)
	return 0
}

//export cwireGoSendOnStream
func cwireGoSendOnStream(udata unsafe.Pointer, h *C.pigeon_stream_handle, data *C.uint8_t, n C.size_t) C.int {
	t := recoverGoTransport(udata)
	src := unsafe.Slice((*byte)(unsafe.Pointer(data)), int(n))
	msg := make([]byte, len(src))
	copy(msg, src)
	if err := t.SendOnStream(unsafe.Pointer(h), msg); err != nil {
		return -1
	}
	return 0
}

//export cwireGoRecvOnStream
func cwireGoRecvOnStream(udata unsafe.Pointer, h *C.pigeon_stream_handle, buf *C.uint8_t, buflen C.size_t, outlen *C.size_t) C.int {
	t := recoverGoTransport(udata)
	msg, err := t.RecvOnStream(unsafe.Pointer(h))
	if err != nil {
		return -1
	}
	if C.size_t(len(msg)) > buflen {
		return -1
	}
	if len(msg) > 0 {
		dst := unsafe.Slice((*byte)(unsafe.Pointer(buf)), int(buflen))
		copy(dst, msg)
	}
	*outlen = C.size_t(len(msg))
	return 0
}

//export cwireGoCloseStream
func cwireGoCloseStream(udata unsafe.Pointer, h *C.pigeon_stream_handle) C.int {
	t := recoverGoTransport(udata)
	if err := t.CloseStream(unsafe.Pointer(h)); err != nil {
		return -1
	}
	return 0
}

//export cwireGoSendDatagram
func cwireGoSendDatagram(udata unsafe.Pointer, data *C.uint8_t, n C.size_t) C.int {
	t := recoverGoTransport(udata)
	src := unsafe.Slice((*byte)(unsafe.Pointer(data)), int(n))
	msg := make([]byte, len(src))
	copy(msg, src)
	if err := t.SendDatagram(msg); err != nil {
		return -1
	}
	return 0
}

//export cwireGoRecvDatagram
func cwireGoRecvDatagram(udata unsafe.Pointer, buf *C.uint8_t, buflen C.size_t, outlen *C.size_t) C.int {
	t := recoverGoTransport(udata)
	msg, err := t.RecvDatagram()
	if err != nil {
		return -1
	}
	if C.size_t(len(msg)) > buflen {
		return -1
	}
	if len(msg) > 0 {
		dst := unsafe.Slice((*byte)(unsafe.Pointer(buf)), int(buflen))
		copy(dst, msg)
	}
	*outlen = C.size_t(len(msg))
	return 0
}

func recoverGoTransport(udata unsafe.Pointer) GoTransport {
	cu := (*C.cwire_go_udata)(udata)
	h := cgo.Handle(cu.handle)
	return h.Value().(GoTransport)
}

// confirmHandle wraps a Go confirm callback so it can be invoked from C
// via the cwire_confirm_trampoline in cwire_pigeon.c.
type confirmHandle struct {
	handle cgo.Handle
	cudata *C.cwire_go_udata
}

// newConfirmHandle registers fn and allocates the C-side userdata box.
func newConfirmHandle(fn func(code string) bool) *confirmHandle {
	h := cgo.NewHandle(fn)
	cu := C.cwire_alloc_go_udata(C.uintptr_t(h))
	return &confirmHandle{handle: h, cudata: cu}
}

// ptr returns the userdata pointer to pass as the confirm_fn userdata arg.
func (c *confirmHandle) ptr() unsafe.Pointer {
	return unsafe.Pointer(c.cudata)
}

// delete releases the cgo.Handle and C-side box.
func (c *confirmHandle) delete() {
	if c == nil {
		return
	}
	if c.cudata != nil {
		C.cwire_free_go_udata(c.cudata)
		c.cudata = nil
	}
	if c.handle != 0 {
		c.handle.Delete()
		c.handle = 0
	}
}

//export cwireGoConfirm
func cwireGoConfirm(udata unsafe.Pointer, code *C.char) C.int {
	cu := (*C.cwire_go_udata)(udata)
	h := cgo.Handle(cu.handle)
	fn := h.Value().(func(code string) bool)
	if fn(C.GoString(code)) {
		return 1
	}
	return 0
}

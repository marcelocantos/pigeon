// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package cwire

// #include <stdlib.h>
// #include "pigeon.h"
// #include "loopback.h"
//
// // Inline allocators / sizeof to dodge cgo's parser sometimes failing
// // to surface stdlib calloc / sizeof in struct types.
// static void *cwire_calloc_session(void)  { return calloc(1, sizeof(pigeon_session)); }
// static void *cwire_calloc_channel(void)  { return calloc(1, sizeof(pigeon_channel)); }
import "C"

import (
	"errors"
	"runtime"
	"unsafe"
)

// Loopback is a paired in-process transport endpoint. Two endpoints
// route streams and datagrams to each other via the C-side loopback
// (no I/O); used by tests to exercise Session/Stream/Datagram without
// ngtcp2 or a live relay.
type Loopback struct {
	c *C.pigeon_loopback_endpoint
}

// NewLoopbackPair allocates two paired loopback endpoints.
func NewLoopbackPair() (*Loopback, *Loopback) {
	a := C.pigeon_loopback_new()
	b := C.pigeon_loopback_new()
	C.pigeon_loopback_pair(a, b)
	la := &Loopback{c: a}
	lb := &Loopback{c: b}
	runtime.SetFinalizer(la, func(l *Loopback) { C.pigeon_loopback_free(l.c) })
	runtime.SetFinalizer(lb, func(l *Loopback) { C.pigeon_loopback_free(l.c) })
	return la, lb
}

// Close eagerly frees the endpoint. Safe to call multiple times.
func (l *Loopback) Close() {
	if l == nil || l.c == nil {
		return
	}
	C.pigeon_loopback_free(l.c)
	l.c = nil
	runtime.SetFinalizer(l, nil)
}

// Channel is a thin wrapper around pigeon_channel for cgo-bridge tests.
// Production callers should use the Go pigeon package's *crypto.Channel
// (this exists only so tests can wire the loopback API end-to-end).
//
// The C struct is heap-allocated via C.calloc so it can be passed
// across cgo without tripping Go's "Go pointer to Go pointer" rule.
type Channel struct {
	c *C.pigeon_channel
}

// NewChannel constructs a Channel with separate send/recv keys (each
// 32 bytes). Mode is one of "stream" (PIGEON_MODE_STRICT) or
// "datagram" (PIGEON_MODE_DATAGRAMS).
func NewChannel(sendKey, recvKey []byte, mode string) (*Channel, error) {
	if len(sendKey) != 32 || len(recvKey) != 32 {
		return nil, errors.New("cwire: channel keys must be 32 bytes")
	}
	var cmode C.pigeon_channel_mode = C.PIGEON_MODE_STRICT
	if mode == "datagram" {
		cmode = C.PIGEON_MODE_DATAGRAMS
	}
	ch := &Channel{
		c: (*C.pigeon_channel)(C.cwire_calloc_channel()),
	}
	C.pigeon_channel_init(ch.c,
		(*C.uint8_t)(unsafe.Pointer(&sendKey[0])),
		(*C.uint8_t)(unsafe.Pointer(&recvKey[0])),
		cmode)
	runtime.SetFinalizer(ch, func(ch *Channel) {
		if ch.c != nil {
			C.free(unsafe.Pointer(ch.c))
			ch.c = nil
		}
	})
	return ch, nil
}

// Session wraps a pigeon_session. Stream + datagram I/O routes through
// the underlying transport (here: a Loopback) and the AEAD channel.
//
// The pigeon_session struct is allocated on the C heap (calloc) so
// passing &Session.c to cgo doesn't trip Go's "Go pointer to Go
// pointer" rule (the C struct contains a transport vtable whose
// userdata field is a C pointer; Go's runtime check is conservative
// about anything within Go-allocated memory).
type Session struct {
	c    *C.pigeon_session
	open bool
}

// DatagramChannel describes a pre-agreed name → varint id mapping.
// Both peers must declare an identical list.
type DatagramChannel struct {
	Name string
	ID   uint64
}

// NewLoopbackSession constructs a Session over a loopback endpoint.
// `isBackend` toggles the 4-byte clientTag prefix on outbound stream
// headers and datagrams.
func NewLoopbackSession(lb *Loopback, ch *Channel, isBackend bool, clientTag uint32, datagrams []DatagramChannel) (*Session, error) {
	if lb == nil || lb.c == nil {
		return nil, errors.New("cwire: nil loopback")
	}
	if ch == nil {
		return nil, errors.New("cwire: nil channel")
	}

	var t C.pigeon_transport
	C.pigeon_loopback_fill_transport(&t, lb.c)

	// Marshal the (name, id) array into a C-side fixed-size buffer.
	if len(datagrams) > 16 { // PIGEON_MAX_DATAGRAM_CHANNELS
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

	s := &Session{
		c: (*C.pigeon_session)(C.cwire_calloc_session()),
	}
	var cIsBackend C.bool
	if isBackend {
		cIsBackend = C.bool(true)
	}
	var defsPtr *C.pigeon_dgchannel_def
	if len(datagrams) > 0 {
		defsPtr = &defs[0]
	}
	rv := C.pigeon_session_init(s.c, &t, ch.c,
		cIsBackend, C.uint32_t(clientTag),
		defsPtr, C.size_t(len(datagrams)))
	if rv != 0 {
		C.free(unsafe.Pointer(s.c))
		return nil, errors.New("cwire: pigeon_session_init failed")
	}
	s.open = true
	runtime.SetFinalizer(s, func(s *Session) { s.Close() })
	return s, nil
}

// Close frees the heap scratch buffers attached to the session. Safe
// to call repeatedly.
func (s *Session) Close() {
	if s == nil || !s.open {
		return
	}
	C.pigeon_session_close(s.c)
	if s.c != nil {
		C.free(unsafe.Pointer(s.c))
		s.c = nil
	}
	s.open = false
	runtime.SetFinalizer(s, nil)
}

// Stream wraps a pigeon_stream.
type Stream struct {
	c       C.pigeon_stream
	session *Session
}

// OpenStream opens a fresh named stream on the session.
func (s *Session) OpenStream(name string) (*Stream, error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	st := &Stream{session: s}
	rv := C.pigeon_session_open_stream(s.c, cName, &st.c)
	if rv != 0 {
		return nil, errors.New("cwire: pigeon_session_open_stream failed")
	}
	return st, nil
}

// AcceptStream is a tiny convenience used in tests: drain the next
// peer-opened stream from the loopback transport and read its
// header. Returns the handle wrapped in a *Stream and the (decoded)
// header name. Production code would do this via
// pigeon_decode_*_stream_header on the raw header buffer.
func (lb *Loopback) AcceptStreamWithHeader(s *Session) (*Stream, uint32, string, error) {
	var hdr [4 + 10 + 256]byte
	var hdrLen C.size_t
	st := &Stream{session: s}
	rv := C.pigeon_loopback_accept_with_header(lb.c, &st.c.handle,
		(*C.uint8_t)(unsafe.Pointer(&hdr[0])), C.size_t(len(hdr)), &hdrLen)
	if rv != 0 {
		return nil, 0, "", errors.New("cwire: accept_with_header failed")
	}
	st.c.session = s.c
	tag, name, _, err := DecodeBackendStreamHeader(hdr[:int(hdrLen)])
	if err != nil {
		// Fall back to client-side header (no tag).
		name2, _, err2 := DecodeClientStreamHeader(hdr[:int(hdrLen)])
		if err2 != nil {
			return nil, 0, "", err
		}
		return st, 0, name2, nil
	}
	// Copy name into the C struct's name field (NUL-terminated).
	if len(name) >= 64 {
		return nil, 0, "", errors.New("cwire: accepted name too long")
	}
	for i := range name {
		st.c.name[i] = C.char(name[i])
	}
	st.c.name[len(name)] = 0
	return st, tag, name, nil
}

// Send AEAD-encrypts and writes a message on the stream.
func (st *Stream) Send(msg []byte) error {
	var p *C.uint8_t
	if len(msg) > 0 {
		p = (*C.uint8_t)(unsafe.Pointer(&msg[0]))
	}
	rv := C.pigeon_stream_send(&st.c, p, C.size_t(len(msg)))
	if rv != 0 {
		return errors.New("cwire: stream_send failed")
	}
	return nil
}

// Recv reads + AEAD-decrypts the next message on the stream.
func (st *Stream) Recv() ([]byte, error) {
	out := make([]byte, 1<<20)
	n := C.pigeon_stream_recv(&st.c,
		(*C.uint8_t)(unsafe.Pointer(&out[0])), C.size_t(len(out)))
	if n < 0 {
		return nil, errors.New("cwire: stream_recv failed")
	}
	res := make([]byte, int(n))
	copy(res, out[:int(n)])
	return res, nil
}

// Close releases the underlying transport stream.
func (st *Stream) Close() {
	if st == nil || st.c.handle == nil {
		return
	}
	C.pigeon_stream_close(&st.c)
	st.c.handle = nil
}

// Datagram wraps a pigeon_datagram on a session's pre-declared channel.
type Datagram struct {
	c C.pigeon_datagram
}

// Datagram looks up a pre-declared datagram channel by name.
func (s *Session) Datagram(name string) (*Datagram, error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	d := &Datagram{}
	rv := C.pigeon_session_get_datagram(s.c, cName, &d.c)
	if rv != 0 {
		return nil, errors.New("cwire: pigeon_session_get_datagram failed")
	}
	return d, nil
}

// Send transmits a datagram payload on this channel.
func (d *Datagram) Send(payload []byte) error {
	var p *C.uint8_t
	if len(payload) > 0 {
		p = (*C.uint8_t)(unsafe.Pointer(&payload[0]))
	}
	rv := C.pigeon_datagram_send(&d.c, p, C.size_t(len(payload)))
	if rv != 0 {
		return errors.New("cwire: datagram_send failed")
	}
	return nil
}

// Recv reads + AEAD-decrypts the next datagram on this channel. If
// the next datagram on the wire is for a different channel, returns
// errors.New("cwire: wrong channel"); the caller should then
// re-dispatch by reading the connection-level recv directly.
func (d *Datagram) Recv() ([]byte, error) {
	out := make([]byte, 1<<20)
	n := C.pigeon_datagram_recv(&d.c,
		(*C.uint8_t)(unsafe.Pointer(&out[0])), C.size_t(len(out)))
	switch {
	case n == -2:
		return nil, errors.New("cwire: wrong channel")
	case n < 0:
		return nil, errors.New("cwire: datagram_recv failed")
	}
	res := make([]byte, int(n))
	copy(res, out[:int(n)])
	return res, nil
}

// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package pigeon

// #include <stdlib.h>
import "C"

import (
	"context"
	"io"
	"sync"
	"unsafe"

	"github.com/marcelocantos/pigeon/cwire"
)

// goTransportAdapter wraps the Go-side *transport (quic-go or WebTransport)
// as a cwire.GoTransport so libpigeon can drive it via cgo.
//
// Per-stream handles are minted as C-malloc'd 1-byte stubs whose addresses
// key into a package-level Go map. The cgo side treats the pointer as
// opaque; the Go side recovers the io.ReadWriteCloser on each Send/Recv/
// Close call. CloseStream frees both the malloc'd byte and the map entry.
type goTransportAdapter struct {
	t   *transport
	ctx context.Context // used for AcceptStream / ReceiveDatagram
}

func newGoTransportAdapter(ctx context.Context, t *transport) *goTransportAdapter {
	return &goTransportAdapter{t: t, ctx: ctx}
}

var (
	streamMu      sync.Mutex
	streamHandles = make(map[unsafe.Pointer]io.ReadWriteCloser)
)

func registerStream(rwc io.ReadWriteCloser) unsafe.Pointer {
	p := C.malloc(1)
	streamMu.Lock()
	streamHandles[p] = rwc
	streamMu.Unlock()
	return p
}

func lookupStream(p unsafe.Pointer) io.ReadWriteCloser {
	streamMu.Lock()
	defer streamMu.Unlock()
	return streamHandles[p]
}

func removeStream(p unsafe.Pointer) io.ReadWriteCloser {
	streamMu.Lock()
	rwc := streamHandles[p]
	delete(streamHandles, p)
	streamMu.Unlock()
	C.free(p)
	return rwc
}

var _ cwire.GoTransport = (*goTransportAdapter)(nil)

func (a *goTransportAdapter) OpenStream() (unsafe.Pointer, error) {
	rwc, err := a.t.OpenStream()
	if err != nil {
		return nil, err
	}
	return registerStream(rwc), nil
}

func (a *goTransportAdapter) AcceptStream() (unsafe.Pointer, error) {
	rwc, err := a.t.AcceptStream(a.ctx)
	if err != nil {
		return nil, err
	}
	return registerStream(rwc), nil
}

func (a *goTransportAdapter) SendOnStream(handle unsafe.Pointer, msg []byte) error {
	rwc := lookupStream(handle)
	if rwc == nil {
		return io.ErrClosedPipe
	}
	return writeMessage(rwc, msg)
}

func (a *goTransportAdapter) RecvOnStream(handle unsafe.Pointer) ([]byte, error) {
	rwc := lookupStream(handle)
	if rwc == nil {
		return nil, io.ErrClosedPipe
	}
	return readMessage(rwc)
}

func (a *goTransportAdapter) CloseStream(handle unsafe.Pointer) error {
	rwc := removeStream(handle)
	if rwc == nil {
		return nil
	}
	return rwc.Close()
}

func (a *goTransportAdapter) SendDatagram(payload []byte) error {
	return a.t.SendDatagram(payload)
}

func (a *goTransportAdapter) RecvDatagram() ([]byte, error) {
	return a.t.ReceiveDatagram(a.ctx)
}

// adoptPrimary registers an already-opened primary stream under a fresh
// handle and returns its opaque pointer. Used by Connect/Register where
// the relay greeting has already created the primary stream as part of
// the dial flow.
func (a *goTransportAdapter) adoptPrimary(rwc io.ReadWriteCloser) unsafe.Pointer {
	return registerStream(rwc)
}

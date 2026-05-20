// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package pairing

import (
	"encoding/binary"
	"errors"
	"io"
	"sync/atomic"
	"unsafe"
)

// singleStreamTransport adapts a single io.ReadWriteCloser as a
// cwire.GoTransport for the pairing ceremony. OpenStream and AcceptStream
// both return the same handle once; subsequent calls return ErrClosed.
// Datagrams are not used by the pairing wire.
//
// Used by runAcceptorOnStream / runInitiatorOnStream to hand an
// already-opened pigeon.Stream to libpigeon's pigeon_pair_acceptor /
// pigeon_pair_initiator via cwire.RunAcceptor / cwire.RunInitiator.
type singleStreamTransport struct {
	rwc    io.ReadWriteCloser
	served atomic.Bool
}

// pairHandleStub is a stable address used as the opaque stream handle.
// Pinned in BSS so its pointer round-trips through cgo without GC
// concerns. One stub is enough because every singleStreamTransport
// serves exactly one stream.
var pairHandleStub [1]byte

func (t *singleStreamTransport) stream() (unsafe.Pointer, error) {
	if t.served.Swap(true) {
		return nil, errors.New("pairing: stream already served")
	}
	return unsafe.Pointer(&pairHandleStub[0]), nil
}

func (t *singleStreamTransport) OpenStream() (unsafe.Pointer, error)   { return t.stream() }
func (t *singleStreamTransport) AcceptStream() (unsafe.Pointer, error) { return t.stream() }

func (t *singleStreamTransport) SendOnStream(_ unsafe.Pointer, msg []byte) error {
	var hdr [4]byte
	binary.BigEndian.PutUint32(hdr[:], uint32(len(msg)))
	if _, err := t.rwc.Write(hdr[:]); err != nil {
		return err
	}
	_, err := t.rwc.Write(msg)
	return err
}

func (t *singleStreamTransport) RecvOnStream(_ unsafe.Pointer) ([]byte, error) {
	var hdr [4]byte
	if _, err := io.ReadFull(t.rwc, hdr[:]); err != nil {
		return nil, err
	}
	n := binary.BigEndian.Uint32(hdr[:])
	buf := make([]byte, n)
	if _, err := io.ReadFull(t.rwc, buf); err != nil {
		return nil, err
	}
	return buf, nil
}

func (t *singleStreamTransport) CloseStream(_ unsafe.Pointer) error { return t.rwc.Close() }

func (t *singleStreamTransport) SendDatagram(_ []byte) error {
	return errors.New("pairing: datagrams not used by ceremony")
}

func (t *singleStreamTransport) RecvDatagram() ([]byte, error) {
	return nil, errors.New("pairing: datagrams not used by ceremony")
}

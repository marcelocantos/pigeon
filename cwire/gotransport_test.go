// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package cwire_test

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"unsafe"

	"github.com/marcelocantos/pigeon/cwire"
)

// pipeHandleStash mints unique opaque pointers for stream handles.
// Each address inside this package-level array is stable for the
// program's lifetime (the array sits in BSS, not the GC-managed
// heap), which lets us pass them through cgo without uintptr
// conversions.
var (
	pipeHandleStash [4096]byte
	pipeHandleNext  atomic.Uint32
)

func mintPipeHandle() unsafe.Pointer {
	i := pipeHandleNext.Add(1) - 1
	if int(i) >= len(pipeHandleStash) {
		panic("pipeHandleStash exhausted")
	}
	return unsafe.Pointer(&pipeHandleStash[i])
}

// pipeWire is a paired in-memory wire shared by two pipeTransport
// endpoints. Each pipeWire holds the inbound queue for the side
// that owns it. Stream messages are keyed by an opaque handle (a
// C-malloc'd 1-byte stub minted in OpenStream); both sides share
// the same handle value for the same logical stream — the opener
// allocates and writes the pointer into the peer's accept queue.
type pipeWire struct {
	mu sync.Mutex

	streamMsgs  map[unsafe.Pointer][][]byte
	acceptQueue []unsafe.Pointer
	datagrams   [][]byte
}

func newPipeWire() *pipeWire {
	return &pipeWire{streamMsgs: map[unsafe.Pointer][][]byte{}}
}

// pipeTransport is one endpoint. `out` is where this endpoint writes
// (peer's inbound). `in` is where this endpoint reads from (own
// inbound).
type pipeTransport struct {
	in  *pipeWire // wire I read from
	out *pipeWire // wire I write to
}

func newPipeTransports() (a, b *pipeTransport) {
	wireA := newPipeWire()
	wireB := newPipeWire()
	a = &pipeTransport{in: wireA, out: wireB}
	b = &pipeTransport{in: wireB, out: wireA}
	return
}

func (t *pipeTransport) OpenStream() (unsafe.Pointer, error) {
	h := mintPipeHandle()
	t.out.mu.Lock()
	t.out.acceptQueue = append(t.out.acceptQueue, h)
	t.out.mu.Unlock()
	return h, nil
}

func (t *pipeTransport) AcceptStream() (unsafe.Pointer, error) {
	t.in.mu.Lock()
	defer t.in.mu.Unlock()
	if len(t.in.acceptQueue) == 0 {
		return nil, errors.New("pipe: no pending stream")
	}
	h := t.in.acceptQueue[0]
	t.in.acceptQueue = t.in.acceptQueue[1:]
	return h, nil
}

func (t *pipeTransport) SendOnStream(handle unsafe.Pointer, msg []byte) error {
	t.out.mu.Lock()
	defer t.out.mu.Unlock()
	cp := make([]byte, len(msg))
	copy(cp, msg)
	t.out.streamMsgs[handle] = append(t.out.streamMsgs[handle], cp)
	return nil
}

func (t *pipeTransport) RecvOnStream(handle unsafe.Pointer) ([]byte, error) {
	t.in.mu.Lock()
	defer t.in.mu.Unlock()
	q := t.in.streamMsgs[handle]
	if len(q) == 0 {
		return nil, errors.New("pipe: empty")
	}
	msg := q[0]
	t.in.streamMsgs[handle] = q[1:]
	return msg, nil
}

func (t *pipeTransport) CloseStream(handle unsafe.Pointer) error {
	t.out.mu.Lock()
	delete(t.out.streamMsgs, handle)
	t.out.mu.Unlock()
	t.in.mu.Lock()
	delete(t.in.streamMsgs, handle)
	t.in.mu.Unlock()
	return nil
}

func (t *pipeTransport) SendDatagram(payload []byte) error {
	t.out.mu.Lock()
	defer t.out.mu.Unlock()
	cp := make([]byte, len(payload))
	copy(cp, payload)
	t.out.datagrams = append(t.out.datagrams, cp)
	return nil
}

func (t *pipeTransport) RecvDatagram() ([]byte, error) {
	t.in.mu.Lock()
	defer t.in.mu.Unlock()
	if len(t.in.datagrams) == 0 {
		return nil, errors.New("pipe: empty")
	}
	d := t.in.datagrams[0]
	t.in.datagrams = t.in.datagrams[1:]
	return d, nil
}

// TestGoTransportStreamRoundTrip drives the C pigeon_session through
// a Go-implemented transport and confirms the wire framing
// (encode_stream_header + AEAD-encrypted message) round-trips
// correctly across the cgo boundary in both directions.
func TestGoTransportStreamRoundTrip(t *testing.T) {
	master := bytes32(0xAB)
	chA, err := cwire.NewChannel(master, master, "stream")
	if err != nil {
		t.Fatal(err)
	}
	chB, err := cwire.NewChannel(master, master, "stream")
	if err != nil {
		t.Fatal(err)
	}

	pa, pb := newPipeTransports()
	refA := cwire.NewGoTransportRef(pa)
	defer refA.Close()
	refB := cwire.NewGoTransportRef(pb)
	defer refB.Close()

	sa, err := cwire.NewGoSession(refA, chA, false, 0, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer sa.Close()
	sb, err := cwire.NewGoSession(refB, chB, true, 0x01020304, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer sb.Close()

	// A opens "control"; B accepts and reads the header.
	saStream, err := sa.OpenStream("control")
	if err != nil {
		t.Fatalf("OpenStream: %v", err)
	}
	sbStream, _, name, err := sb.AcceptStreamFromGo(refB)
	if err != nil {
		t.Fatalf("AcceptStream: %v", err)
	}
	if name != "control" {
		t.Fatalf("got name %q, want control", name)
	}

	if err := saStream.Send([]byte("hello from A")); err != nil {
		t.Fatalf("A.Send: %v", err)
	}
	got, err := sbStream.Recv()
	if err != nil {
		t.Fatalf("B.Recv: %v", err)
	}
	if string(got) != "hello from A" {
		t.Fatalf("got %q", got)
	}

	if err := sbStream.Send([]byte("hi back from B")); err != nil {
		t.Fatalf("B.Send: %v", err)
	}
	got, err = saStream.Recv()
	if err != nil {
		t.Fatalf("A.Recv: %v", err)
	}
	if string(got) != "hi back from B" {
		t.Fatalf("got %q", got)
	}
}

// TestGoTransportDatagramRoundTrip exercises the datagram half of
// the vtable.
func TestGoTransportDatagramRoundTrip(t *testing.T) {
	master := bytes32(0xCD)
	chA, err := cwire.NewChannel(master, master, "datagram")
	if err != nil {
		t.Fatal(err)
	}
	chB, err := cwire.NewChannel(master, master, "datagram")
	if err != nil {
		t.Fatal(err)
	}

	pa, pb := newPipeTransports()
	refA := cwire.NewGoTransportRef(pa)
	defer refA.Close()
	refB := cwire.NewGoTransportRef(pb)
	defer refB.Close()

	dgChans := []cwire.DatagramChannel{{Name: "telemetry", ID: 7}}

	// Both sides are configured as clients so the wire has no
	// 4-byte tag prefix in either direction. The smoke test isn't
	// modelling a relay between them — it's exercising the cgo
	// callbacks for send_datagram / recv_datagram.
	sa, err := cwire.NewGoSession(refA, chA, false, 0, dgChans)
	if err != nil {
		t.Fatal(err)
	}
	defer sa.Close()
	sb, err := cwire.NewGoSession(refB, chB, false, 0, dgChans)
	if err != nil {
		t.Fatal(err)
	}
	defer sb.Close()

	dgA, err := sa.Datagram("telemetry")
	if err != nil {
		t.Fatal(err)
	}
	dgB, err := sb.Datagram("telemetry")
	if err != nil {
		t.Fatal(err)
	}

	if err := dgA.Send([]byte("ping")); err != nil {
		t.Fatalf("A.Send dg: %v", err)
	}
	got, err := dgB.Recv()
	if err != nil {
		t.Fatalf("B.Recv dg: %v", err)
	}
	if string(got) != "ping" {
		t.Fatalf("got %q", got)
	}

	if err := dgB.Send([]byte("pong")); err != nil {
		t.Fatalf("B.Send dg: %v", err)
	}
	got, err = dgA.Recv()
	if err != nil {
		t.Fatalf("A.Recv dg: %v", err)
	}
	if string(got) != "pong" {
		t.Fatalf("got %q", got)
	}
}

func bytes32(b byte) []byte {
	out := make([]byte, 32)
	for i := range out {
		out[i] = b
	}
	return out
}

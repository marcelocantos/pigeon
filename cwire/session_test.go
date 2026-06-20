// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package cwire_test

import (
	"bytes"
	"testing"

	"github.com/marcelocantos/pigeon/cwire"
)

// symKey is a fixed 32-byte symmetric key used by both endpoints in
// these in-process loopback tests. Production uses derives the key
// from a PairingRecord; here we just want to drive Session/Stream/
// Datagram round-trips through the C ABI.
func symKey() []byte {
	k := make([]byte, 32)
	for i := range k {
		k[i] = 0x42
	}
	return k
}

// TestSessionStreamRoundTripViaCgo drives pigeon_session_open_stream +
// pigeon_stream_send + pigeon_stream_recv from Go via cgo. Two
// loopback endpoints route the messages in-process.
func TestSessionStreamRoundTripViaCgo(t *testing.T) {
	la, lb := cwire.NewLoopbackPair()
	defer la.Close()
	defer lb.Close()

	chA, err := cwire.NewChannel(symKey(), symKey(), "stream")
	if err != nil {
		t.Fatal(err)
	}
	chB, err := cwire.NewChannel(symKey(), symKey(), "stream")
	if err != nil {
		t.Fatal(err)
	}

	sA, err := cwire.NewLoopbackSession(la, chA, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer sA.Close()
	sB, err := cwire.NewLoopbackSession(lb, chB, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer sB.Close()

	chat, err := sA.OpenStream("chat")
	if err != nil {
		t.Fatal(err)
	}
	defer chat.Close()

	st, name, err := lb.AcceptStreamWithHeader(sB)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	if name != "chat" {
		t.Fatalf("name: got %q want %q", name, "chat")
	}

	if err := chat.Send([]byte("hello")); err != nil {
		t.Fatal(err)
	}
	got, err := st.Recv()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, []byte("hello")) {
		t.Fatalf("recv on B: got %q", got)
	}

	if err := st.Send([]byte("world")); err != nil {
		t.Fatal(err)
	}
	got, err = chat.Recv()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, []byte("world")) {
		t.Fatalf("recv on A: got %q", got)
	}
}

// TestSessionDatagramRoundTripViaCgo drives pigeon_datagram_send /
// pigeon_datagram_recv through the cgo bridge.
func TestSessionDatagramRoundTripViaCgo(t *testing.T) {
	la, lb := cwire.NewLoopbackPair()
	defer la.Close()
	defer lb.Close()

	chA, err := cwire.NewChannel(symKey(), symKey(), "datagram")
	if err != nil {
		t.Fatal(err)
	}
	chB, err := cwire.NewChannel(symKey(), symKey(), "datagram")
	if err != nil {
		t.Fatal(err)
	}

	dgs := []cwire.DatagramChannel{
		{Name: "ping", ID: 1},
		{Name: "metric", ID: 2},
	}

	// Under T45 both peers are symmetric — the datagram wire is just
	// AEAD([varint channel-id][payload]) with no tag prefix.
	sA, err := cwire.NewLoopbackSession(la, chA, dgs)
	if err != nil {
		t.Fatal(err)
	}
	defer sA.Close()
	sB, err := cwire.NewLoopbackSession(lb, chB, dgs)
	if err != nil {
		t.Fatal(err)
	}
	defer sB.Close()

	pingA, err := sA.Datagram("ping")
	if err != nil {
		t.Fatal(err)
	}
	pingB, err := sB.Datagram("ping")
	if err != nil {
		t.Fatal(err)
	}

	if err := pingA.Send([]byte("p1")); err != nil {
		t.Fatal(err)
	}
	got, err := pingB.Recv()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, []byte("p1")) {
		t.Fatalf("ping recv: got %q", got)
	}

	if err := pingB.Send([]byte("p2")); err != nil {
		t.Fatal(err)
	}
	got, err = pingA.Recv()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, []byte("p2")) {
		t.Fatalf("ping recv: got %q", got)
	}
}

// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package pigeon

import (
	"context"
	"errors"
	"testing"
)

// fakeRelaySession is a minimal relaySession whose only meaningful
// behaviour is a cancelable Context (the instance lifecycle derives from
// it). All I/O methods are unused by the hub-ownership tests.
type fakeRelaySession struct {
	ctx    context.Context
	cancel context.CancelFunc
}

func newFakeRelaySession() *fakeRelaySession {
	ctx, cancel := context.WithCancel(context.Background())
	return &fakeRelaySession{ctx: ctx, cancel: cancel}
}

func (f *fakeRelaySession) ReadMessage() ([]byte, error) { return nil, errors.New("unused") }
func (f *fakeRelaySession) WriteMessage([]byte) error    { return nil }
func (f *fakeRelaySession) SendDatagram([]byte) error    { return nil }
func (f *fakeRelaySession) Context() context.Context     { return f.ctx }
func (f *fakeRelaySession) Close() error                 { f.cancel(); return nil }
func (f *fakeRelaySession) AcceptStream(context.Context) (readWriteCloserPair, error) {
	return nil, errors.New("unused")
}
func (f *fakeRelaySession) OpenStream() (readWriteCloserPair, error) {
	return nil, errors.New("unused")
}
func (f *fakeRelaySession) ReceiveDatagram(context.Context) ([]byte, error) {
	return nil, errors.New("unused")
}

// TestHubRegisterRejectsLiveIDHijack reproduces Fable-5 F8 (🎯T56): on the
// default open relay, hub.register was last-writer-wins on a client-chosen
// instance ID, so a peer could overwrite a live backend's slot and hijack
// every client that connects to that ID. register must refuse to displace
// a live instance.
func TestHubRegisterRejectsLiveIDHijack(t *testing.T) {
	h := newHub()

	ctrlA := newFakeRelaySession()
	a := newInstance("X", ctrlA)
	if err := h.register(a); err != nil {
		t.Fatalf("register(a): unexpected error %v", err)
	}

	// Attacker (or racing reconnect) tries to claim the same live ID.
	ctrlB := newFakeRelaySession()
	b := newInstance("X", ctrlB)
	if err := h.register(b); err == nil {
		t.Fatalf("register(b) claimed a live ID — hijack not prevented")
	}
	if got := h.get("X"); got != a {
		t.Fatalf("live victim displaced: get(X)=%p want a=%p", got, a)
	}
	if a.ctx.Err() != nil {
		t.Fatalf("victim instance was cancelled by a rejected hijack")
	}
}

// TestHubUnregisterOwnershipNoCrossCancel reproduces Fable-5 F9 (🎯T56):
// hub.unregister deleted/cancelled whatever was mapped to an ID regardless
// of ownership, so an old connection's teardown could cancel a freshly
// re-registered live instance. A dead slot may be taken over, but the stale
// owner's unregister must be a no-op that never disturbs the new instance.
func TestHubUnregisterOwnershipNoCrossCancel(t *testing.T) {
	h := newHub()

	// Instance A registers, then its connection dies (ctx cancelled) —
	// but its deferred unregister has not run yet.
	ctrlA := newFakeRelaySession()
	a := newInstance("X", ctrlA)
	if err := h.register(a); err != nil {
		t.Fatalf("register(a): %v", err)
	}
	ctrlA.cancel() // connection lost; a.ctx now Done

	// Backend reconnects and re-registers the same stable ID on a new,
	// live connection. The dead slot is taken over.
	ctrlB := newFakeRelaySession()
	b := newInstance("X", ctrlB)
	if err := h.register(b); err != nil {
		t.Fatalf("register(b) rejected despite dead prior owner: %v", err)
	}
	if got := h.get("X"); got != b {
		t.Fatalf("takeover failed: get(X)=%p want b=%p", got, b)
	}

	// Now A's long-delayed teardown fires. It must NOT evict or cancel B.
	h.unregister(a)

	if got := h.get("X"); got != b {
		t.Fatalf("stale unregister evicted the live instance: get(X)=%p want b=%p", got, b)
	}
	if b.ctx.Err() != nil {
		t.Fatalf("stale unregister cross-cancelled the live instance B")
	}

	// B's own teardown removes exactly its slot.
	h.unregister(b)
	if h.get("X") != nil {
		t.Fatalf("owner unregister did not remove its own slot")
	}
	if b.ctx.Err() == nil {
		t.Fatalf("owner unregister did not cancel its own instance")
	}
}

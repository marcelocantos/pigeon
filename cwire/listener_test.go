// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package cwire_test

import (
	"crypto/ecdh"
	"crypto/rand"
	"testing"
	"unsafe"

	"github.com/marcelocantos/pigeon/cwire"
)

// genX25519 generates a fresh X25519 keypair, fatal on error.
func genX25519(t *testing.T) (*ecdh.PrivateKey, *ecdh.PublicKey) {
	t.Helper()
	k, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("keygen: %v", err)
	}
	return k, k.PublicKey()
}

// openPrimaryStream simulates what a relay would do: opens a stream from
// the client transport, sends a backend-format header ([4-byte tag][varint
// 0-length name]) on it, and returns the stream handle. The listener's
// accept_stream will dequeue this handle, read the header, see name_len==0
// and run activation.
func openPrimaryStream(t *testing.T, tr *chanTransport, clientTag uint32) unsafe.Pointer {
	t.Helper()
	h, err := tr.OpenStream()
	if err != nil {
		t.Fatalf("OpenStream: %v", err)
	}
	// Backend-format header: [4-byte tag][varint 0][no name] — simulates
	// what the relay prepends before forwarding the stream to the backend.
	hdr, err := cwire.EncodeStreamHeader(true, clientTag, "")
	if err != nil {
		t.Fatalf("EncodeStreamHeader: %v", err)
	}
	if err := tr.SendOnStream(h, hdr); err != nil {
		t.Fatalf("SendOnStream header: %v", err)
	}
	return h
}

// TestListenerAcceptNewPrimary verifies:
//   - A client completing activation triggers PairingResolver with the right device_id.
//   - A valid record yields a non-nil Session from Accept.
func TestListenerAcceptNewPrimary(t *testing.T) {
	t.Parallel()

	const (
		deviceID   = "test-device-abc123"
		listenerID = "backend-instance-1"
	)

	backendPriv, backendPub := genX25519(t)
	clientPriv, clientPub := genX25519(t)

	// backend record: local=backend, peer=client.
	backendRecord := &cwire.PairingRecord{
		LocalPrivKey: backendPriv.Bytes(),
		LocalPubKey:  backendPub.Bytes(),
		PeerPubKey:   clientPub.Bytes(),
	}
	_ = clientPriv // used only to derive clientPub

	resolverCalled := make(chan string, 1)
	resolve := func(id string) (*cwire.PairingRecord, bool) {
		resolverCalled <- id
		if id == deviceID {
			return backendRecord, true
		}
		return nil, false
	}

	backendTr, clientTr := newChanTransportPair()

	listener, err := cwire.NewListener(backendTr, listenerID, resolve, nil)
	if err != nil {
		t.Fatalf("NewListener: %v", err)
	}
	defer listener.Close()

	if got := listener.InstanceID(); got != listenerID {
		t.Errorf("InstanceID: got %q, want %q", got, listenerID)
	}

	clientRef := cwire.NewGoTransportRef(clientTr)
	defer clientRef.Close()

	type sessResult struct {
		sess *cwire.Session
		err  error
	}
	serverCh := make(chan sessResult, 1)
	clientCh := make(chan error, 1)

	// Listener goroutine: blocks in Accept.
	go func() {
		sess, err := listener.Accept()
		serverCh <- sessResult{sess: sess, err: err}
	}()

	// Client goroutine: open primary stream with backend-format header,
	// then drive the activation exchange.
	go func() {
		h := openPrimaryStream(t, clientTr, 0x00000001)
		clientCh <- cwire.RunClientActivation(clientRef, h, deviceID)
	}()

	clientErr := <-clientCh
	serverRes := <-serverCh

	if clientErr != nil {
		t.Fatalf("client activation: %v", clientErr)
	}
	if serverRes.err != nil {
		t.Fatalf("listener Accept: %v", serverRes.err)
	}
	if serverRes.sess == nil {
		t.Fatal("listener session is nil")
	}

	// Resolver must have been called with the right device ID.
	select {
	case id := <-resolverCalled:
		if id != deviceID {
			t.Errorf("resolver called with %q, want %q", id, deviceID)
		}
	default:
		t.Error("resolver was never called")
	}
}

// TestListenerResolverRejectClient verifies that when PairingResolver returns
// false, Step returns (nil, nil) — rejected clients don't materialise a Session.
func TestListenerResolverRejectClient(t *testing.T) {
	t.Parallel()

	const (
		deviceID   = "unknown-device"
		listenerID = "backend-instance-2"
	)

	resolve := func(id string) (*cwire.PairingRecord, bool) {
		return nil, false // reject everyone
	}

	backendTr, clientTr := newChanTransportPair()

	listener, err := cwire.NewListener(backendTr, listenerID, resolve, nil)
	if err != nil {
		t.Fatalf("NewListener: %v", err)
	}
	defer listener.Close()

	clientRef := cwire.NewGoTransportRef(clientTr)
	defer clientRef.Close()

	clientErrCh := make(chan error, 1)
	go func() {
		h := openPrimaryStream(t, clientTr, 0x00000001)
		clientErrCh <- cwire.RunClientActivation(clientRef, h, deviceID)
	}()

	// Step once — the listener should consume the primary, reject it, and
	// return (nil, nil) without a transport failure.
	sess, stepErr := listener.Step()
	if stepErr != nil {
		t.Fatalf("Step: unexpected transport error: %v", stepErr)
	}
	if sess != nil {
		t.Errorf("Step: got session for rejected client, want nil")
	}

	// The client's activation should have failed (auth_ok ok=false).
	clientErr := <-clientErrCh
	if clientErr == nil {
		t.Error("client: expected rejection error, got nil")
	}
}

// TestListenerSubstreamAcceptIncomingStream verifies:
//   - A named sub-stream opened after activation lands in the session's queue.
//   - AcceptIncomingStream retrieves it by name.
//   - AcceptIncomingStream returns (nil, nil) for an unknown name.
func TestListenerSubstreamAcceptIncomingStream(t *testing.T) {
	t.Parallel()

	const (
		deviceID      = "device-substream"
		listenerID    = "backend-instance-3"
		subStreamName = "data"
		clientTag     = uint32(0x00000002)
	)

	backendPriv, backendPub := genX25519(t)
	clientPriv, clientPub := genX25519(t)
	_ = clientPriv

	backendRecord := &cwire.PairingRecord{
		LocalPrivKey: backendPriv.Bytes(),
		LocalPubKey:  backendPub.Bytes(),
		PeerPubKey:   clientPub.Bytes(),
	}

	resolve := func(id string) (*cwire.PairingRecord, bool) {
		if id == deviceID {
			return backendRecord, true
		}
		return nil, false
	}

	backendTr, clientTr := newChanTransportPair()

	listener, err := cwire.NewListener(backendTr, listenerID, resolve, nil)
	if err != nil {
		t.Fatalf("NewListener: %v", err)
	}
	defer listener.Close()

	clientRef := cwire.NewGoTransportRef(clientTr)
	defer clientRef.Close()

	type sessResult struct {
		sess *cwire.Session
		err  error
	}
	serverCh := make(chan sessResult, 1)
	clientDone := make(chan error, 1)

	// Listener goroutine: block in Accept for the primary.
	go func() {
		sess, err := listener.Accept()
		serverCh <- sessResult{sess: sess, err: err}
	}()

	// Client goroutine: activate, then open a named sub-stream.
	go func() {
		// Step 1: primary activation.
		h := openPrimaryStream(t, clientTr, clientTag)
		if err := cwire.RunClientActivation(clientRef, h, deviceID); err != nil {
			clientDone <- err
			return
		}

		// Step 2: open sub-stream. The relay would prepend the same clientTag
		// plus the name. We simulate that here with is_backend=true.
		sh, err := clientTr.OpenStream()
		if err != nil {
			clientDone <- err
			return
		}
		subHdr, err := cwire.EncodeStreamHeader(true, clientTag, subStreamName)
		if err != nil {
			clientDone <- err
			return
		}
		if err := clientTr.SendOnStream(sh, subHdr); err != nil {
			clientDone <- err
			return
		}
		clientDone <- nil
	}()

	serverRes := <-serverCh
	if serverRes.err != nil {
		t.Fatalf("listener Accept: %v", serverRes.err)
	}
	if serverRes.sess == nil {
		t.Fatal("listener session is nil after Accept")
	}

	if err := <-clientDone; err != nil {
		t.Fatalf("client: %v", err)
	}

	// Pump the listener once to dispatch the sub-stream.
	subSess, stepErr := listener.Step()
	if stepErr != nil {
		t.Fatalf("listener Step for sub-stream: %v", stepErr)
	}
	if subSess != nil {
		t.Errorf("Step returned a session for sub-stream; want nil")
	}

	// Sub-stream should now be in the session's queue.
	st, err := serverRes.sess.AcceptIncomingStream(subStreamName)
	if err != nil {
		t.Fatalf("AcceptIncomingStream: %v", err)
	}
	if st == nil {
		t.Fatal("AcceptIncomingStream: got nil, want sub-stream handle")
	}

	// Unknown name returns (nil, nil).
	st2, err2 := serverRes.sess.AcceptIncomingStream("nonexistent")
	if err2 != nil {
		t.Fatalf("AcceptIncomingStream(nonexistent): unexpected error: %v", err2)
	}
	if st2 != nil {
		t.Error("AcceptIncomingStream(nonexistent): want nil, got stream")
	}
}

// TestListenerNewListenerValidation verifies that NewListener rejects
// invalid arguments eagerly.
func TestListenerNewListenerValidation(t *testing.T) {
	t.Parallel()

	resolve := func(id string) (*cwire.PairingRecord, bool) { return nil, false }
	tr, _ := newChanTransportPair()

	tests := []struct {
		name    string
		tr      cwire.GoTransport
		resolve cwire.PairingResolver
	}{
		{"nil transport", nil, resolve},
		{"nil resolver", tr, nil},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := cwire.NewListener(tc.tr, "iid", tc.resolve, nil)
			if err == nil {
				t.Errorf("%s: expected error, got nil", tc.name)
			}
		})
	}
}

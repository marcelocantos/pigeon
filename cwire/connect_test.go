// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package cwire_test

import (
	"bytes"
	"errors"
	"sync"
	"testing"
	"unsafe"

	"github.com/marcelocantos/pigeon/cwire"
)

// bpipeWire is a blocking in-memory wire shared by two bpipeTransport
// endpoints. Unlike pipeTransport (which returns "empty" immediately),
// each channel blocks until a message is enqueued. This is needed for
// the activation handshake which requires true bidirectional blocking I/O:
// the client sends auth_request then blocks waiting for auth_ok, while
// the backend reads auth_request then sends auth_ok.
type bpipeWire struct {
	// streams is indexed by the same handle pointer used by the sender.
	// Each map value is a buffered channel of byte slices.
	mu      sync.Mutex
	streams map[unsafe.Pointer]chan []byte // per-stream message queue
	accept  chan unsafe.Pointer            // stream accept queue
	dgs     chan []byte                    // datagram queue
}

func newBpipeWire() *bpipeWire {
	return &bpipeWire{
		streams: make(map[unsafe.Pointer]chan []byte),
		accept:  make(chan unsafe.Pointer, 64),
		dgs:     make(chan []byte, 64),
	}
}

func (w *bpipeWire) streamChan(h unsafe.Pointer) chan []byte {
	w.mu.Lock()
	defer w.mu.Unlock()
	ch, ok := w.streams[h]
	if !ok {
		ch = make(chan []byte, 128)
		w.streams[h] = ch
	}
	return ch
}

// bpipeTransport implements GoTransport over two bpipeWires.
// Writes go to `out`; reads come from `in`.
type bpipeTransport struct {
	in  *bpipeWire
	out *bpipeWire
}

func newBpipeTransportPair() (a, b *bpipeTransport) {
	wireA := newBpipeWire()
	wireB := newBpipeWire()
	a = &bpipeTransport{in: wireA, out: wireB}
	b = &bpipeTransport{in: wireB, out: wireA}
	return
}

func (t *bpipeTransport) OpenStream() (unsafe.Pointer, error) {
	h := mintPipeHandle() // reuse the BSS-backed handle allocator from gotransport_test.go
	t.out.accept <- h
	return h, nil
}

func (t *bpipeTransport) AcceptStream() (unsafe.Pointer, error) {
	h, ok := <-t.in.accept
	if !ok {
		return nil, errors.New("bpipe: accept channel closed")
	}
	return h, nil
}

func (t *bpipeTransport) SendOnStream(handle unsafe.Pointer, msg []byte) error {
	cp := make([]byte, len(msg))
	copy(cp, msg)
	t.out.streamChan(handle) <- cp
	return nil
}

func (t *bpipeTransport) RecvOnStream(handle unsafe.Pointer) ([]byte, error) {
	msg, ok := <-t.in.streamChan(handle)
	if !ok {
		return nil, errors.New("bpipe: stream closed")
	}
	return msg, nil
}

func (t *bpipeTransport) CloseStream(handle unsafe.Pointer) error {
	return nil
}

func (t *bpipeTransport) SendDatagram(payload []byte) error {
	cp := make([]byte, len(payload))
	copy(cp, payload)
	t.out.dgs <- cp
	return nil
}

func (t *bpipeTransport) RecvDatagram() ([]byte, error) {
	d, ok := <-t.in.dgs
	if !ok {
		return nil, errors.New("bpipe: datagram channel closed")
	}
	return d, nil
}

// TestConnectActivationRoundTrip runs pigeon_connect_on_transport (client) and
// pigeon_run_backend_activation (backend) against each other over the blocking
// in-memory transport pair, then verifies stream and datagram round-trips on
// the resulting sessions.
//
// Topology:
//
//	client side:  ta (bpipeTransport) → refA → Connect
//	backend side: tb (bpipeTransport) → refB → RunBackendActivation + NewGoSession
//
// Both activation functions perform blocking I/O so they run concurrently.
func TestConnectActivationRoundTrip(t *testing.T) {
	// Generate real X25519 key pairs so ECDH produces a real shared secret.
	// The HKDF used by pigeon_connect_on_transport and DeriveSessionChannel
	// is keyed from DH(client_priv, backend_pub) == DH(backend_priv, client_pub),
	// so both sides derive the same session keys when they hold the correct
	// peer public key.
	clientKP, err := cwire.GenerateKeypair()
	if err != nil {
		t.Fatalf("GenerateKeypair (client): %v", err)
	}
	backendKP, err := cwire.GenerateKeypair()
	if err != nil {
		t.Fatalf("GenerateKeypair (backend): %v", err)
	}

	// Client's PairingRecord: local = client keys, peer = backend pubkey.
	clientRec := &cwire.PairingRecord{
		PeerInstanceID: "backend-001",
		LocalPrivKey:   clientKP.PrivKey,
		LocalPubKey:    clientKP.PubKey,
		PeerPubKey:     backendKP.PubKey,
	}
	// Backend's PairingRecord: local = backend keys, peer = client pubkey.
	backendRec := &cwire.PairingRecord{
		PeerInstanceID: "client-001",
		LocalPrivKey:   backendKP.PrivKey,
		LocalPubKey:    backendKP.PubKey,
		PeerPubKey:     clientKP.PubKey,
	}

	ta, tb := newBpipeTransportPair()

	refA := cwire.NewGoTransportRef(ta)
	defer refA.Close()
	refB := cwire.NewGoTransportRef(tb)
	defer refB.Close()

	const deviceID = "device-abc123"
	const peerIID = "backend-001"

	dgs := []cwire.DatagramChannel{{Name: "events", ID: 5}}

	var (
		clientSess *cwire.Session
		backendRes *cwire.BackendActivationResult
		wg         sync.WaitGroup
		clientErr  error
		backendErr error
	)

	wg.Add(2)

	go func() {
		defer wg.Done()
		primary, err := ta.OpenStream()
		if err != nil {
			clientErr = err
			return
		}
		clientSess, clientErr = cwire.Connect(&cwire.ConnectArgs{
			Ref:            refA,
			Primary:        primary,
			PeerInstanceID: peerIID,
			DeviceID:       deviceID,
			Record:         clientRec,
			Datagrams:      dgs,
		})
	}()

	go func() {
		defer wg.Done()
		stream, err := tb.AcceptStream()
		if err != nil {
			backendErr = err
			return
		}
		backendRes, backendErr = cwire.RunBackendActivation(&cwire.RunBackendActivationArgs{
			Ref:               refB,
			Stream:            stream,
			SkipPrimaryHeader: true, // pigeon_connect_on_transport writes the
			// primary header first; drain it before reading auth_request.
			ResolveFn: func(id string) (*cwire.PairingRecord, bool) {
				if id == deviceID {
					return backendRec, true
				}
				return nil, false
			},
		})
	}()

	wg.Wait()

	if clientErr != nil {
		t.Fatalf("Connect: %v", clientErr)
	}
	if backendErr != nil {
		t.Fatalf("RunBackendActivation: %v", backendErr)
	}
	if !backendRes.Accepted {
		t.Fatalf("backend: device not accepted (device_id=%q)", backendRes.DeviceID)
	}
	if backendRes.DeviceID != deviceID {
		t.Fatalf("backend: got device_id %q, want %q", backendRes.DeviceID, deviceID)
	}

	defer clientSess.Close()

	// Derive the backend session channel: isBackend=true swaps send/recv info
	// strings so the keys are complementary to the client's channel.
	backendCh, err := cwire.DeriveSessionChannel(backendRec, true)
	if err != nil {
		t.Fatalf("DeriveSessionChannel (backend): %v", err)
	}
	// In a direct (no-relay) in-process test, the backend session must use
	// is_backend=false and clientTag=0, just like pigeon_connect_on_transport
	// sets on the client side. In production, the relay adds the 4-byte
	// clientTag prefix; here we bypass the relay so both sides use the same
	// framing.
	backendSess, err := cwire.NewGoSession(refB, backendCh, false, 0, dgs)
	if err != nil {
		t.Fatalf("NewGoSession (backend): %v", err)
	}
	defer backendSess.Close()

	// Client opens a named stream; backend accepts it.
	cStream, err := clientSess.OpenStream("chat")
	if err != nil {
		t.Fatalf("client OpenStream: %v", err)
	}
	defer cStream.Close()

	bStream, _, name, err := backendSess.AcceptStreamFromGo(refB)
	if err != nil {
		t.Fatalf("backend AcceptStreamFromGo: %v", err)
	}
	defer bStream.Close()
	if name != "chat" {
		t.Fatalf("stream name: got %q, want %q", name, "chat")
	}

	// Stream round-trip: client → backend.
	if err := cStream.Send([]byte("hello from client")); err != nil {
		t.Fatalf("client Send: %v", err)
	}
	got, err := bStream.Recv()
	if err != nil {
		t.Fatalf("backend Recv: %v", err)
	}
	if !bytes.Equal(got, []byte("hello from client")) {
		t.Fatalf("backend Recv: got %q", got)
	}

	// Stream round-trip: backend → client.
	if err := bStream.Send([]byte("hi from backend")); err != nil {
		t.Fatalf("backend Send: %v", err)
	}
	got, err = cStream.Recv()
	if err != nil {
		t.Fatalf("client Recv: %v", err)
	}
	if !bytes.Equal(got, []byte("hi from backend")) {
		t.Fatalf("client Recv: got %q", got)
	}

	// Datagram round-trip: client → backend.
	cDg, err := clientSess.Datagram("events")
	if err != nil {
		t.Fatalf("client Datagram: %v", err)
	}
	bDg, err := backendSess.Datagram("events")
	if err != nil {
		t.Fatalf("backend Datagram: %v", err)
	}

	if err := cDg.Send([]byte("event-1")); err != nil {
		t.Fatalf("client dg Send: %v", err)
	}
	got, err = bDg.Recv()
	if err != nil {
		t.Fatalf("backend dg Recv: %v", err)
	}
	if !bytes.Equal(got, []byte("event-1")) {
		t.Fatalf("backend dg Recv: got %q", got)
	}

	// Datagram round-trip: backend → client.
	if err := bDg.Send([]byte("ack-1")); err != nil {
		t.Fatalf("backend dg Send: %v", err)
	}
	got, err = cDg.Recv()
	if err != nil {
		t.Fatalf("client dg Recv: %v", err)
	}
	if !bytes.Equal(got, []byte("ack-1")) {
		t.Fatalf("client dg Recv: got %q", got)
	}
}

// TestConnectPairingMode verifies that Connect in pairing mode (nil record)
// succeeds and returns a session whose channel is unestablished. This exercises
// the pairing-mode branch of pigeon_connect_on_transport (no auth handshake).
// In pairing mode there is no activation exchange, so a blocking transport is
// not required.
func TestConnectPairingMode(t *testing.T) {
	ta, _ := newBpipeTransportPair()

	refA := cwire.NewGoTransportRef(ta)
	defer refA.Close()

	primary, err := ta.OpenStream()
	if err != nil {
		t.Fatalf("OpenStream: %v", err)
	}

	sess, err := cwire.Connect(&cwire.ConnectArgs{
		Ref:            refA,
		Primary:        primary,
		PeerInstanceID: "backend-002",
		DeviceID:       "",
		Record:         nil,
		Datagrams:      nil,
	})
	if err != nil {
		t.Fatalf("Connect (pairing mode): %v", err)
	}
	defer sess.Close()
	// Session should be valid; channel is unestablished until ceremony.
}

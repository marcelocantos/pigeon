// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

//go:build csdke2e

package csdke2e

// 🎯T33 cross-language wire interop. The four tests below mix the Go
// peer (pigeon.Register / pigeon.Connect — uses Go's quic-go transport
// plus libpigeon-via-cwire for activation, session, and AEAD) with
// the C peer (pigeon_register / pigeon_connect — uses C/ngtcp2 for
// transport and libpigeon directly for everything else). Both peers
// talk through the same in-process Go relay (pigeon.NewQUICServer).
//
// Why this matters: post-T34 the Go peer library is a thin shim over
// libpigeon, but it still rides a different transport (Go quic-go,
// not ngtcp2). The wire-format primitives — stream headers, datagram
// framing, activation auth_request/auth_ok, per-client tag demux —
// must agree byte-for-byte between the two transports. Pure-C tests
// (TestCSDK*) verify ngtcp2 ↔ ngtcp2, and pure-Go tests (TestE2E*)
// verify quic-go ↔ quic-go. These mixed tests verify ngtcp2 ↔
// quic-go through the relay, which is the only configuration that
// catches drift between the two transport adapters' wire layouts.
//
// Channel parity: each single-client topology round-trips a chat
// stream, a control stream, a ping datagram, and a metric datagram —
// the four-channel set defined by the multi-stream API
// (pigeon_session_open_stream / pigeon_session_get_datagram on the
// C side, *Session.OpenStream / *Session.Datagram on the Go side).
//
// Multi-client topology: two clients (of the opposite language to the
// backend) connect concurrently and exercise the relay's clientTag
// demux. The backend tags each reply with the message it received so
// the test can confirm no cross-talk.

import (
	"context"
	"crypto/ecdh"
	"crypto/tls"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/marcelocantos/pigeon"
	"github.com/marcelocantos/pigeon/crypto"
)

// pairingPairFor builds matching backend and client crypto.PairingRecord
// views plus a csdke2e.PairingRecord view of each, for use by the C side.
// All four records describe the same ECDH bond — pigeon.Register/Connect
// and connectClient/registerBackend derive identical AEAD keys from them.
type interopBond struct {
	backendID    crypto.Identity
	clientID     crypto.Identity
	backendGo    *crypto.PairingRecord // backend's view, for pigeon.Register's resolver
	clientGo     *crypto.PairingRecord // client's view, for pigeon.Connect
	backendCView *PairingRecord        // backend's view, for csdke2e registerBackend's resolver
	clientCView  *PairingRecord        // client's view, for csdke2e connectClient
}

func mintInteropBond(t *testing.T, tmp, relayURL, suffix string) *interopBond {
	t.Helper()
	bid, err := crypto.NewFileIdentity(tmp + "/bond-backend-" + suffix + ".json")
	if err != nil {
		t.Fatalf("backend identity: %v", err)
	}
	cid, err := crypto.NewFileIdentity(tmp + "/bond-client-" + suffix + ".json")
	if err != nil {
		t.Fatalf("client identity: %v", err)
	}
	bkp, err := crypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("backend kp: %v", err)
	}
	ckp, err := crypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("client kp: %v", err)
	}
	cPubB, _ := ecdh.X25519().NewPublicKey(ckp.Public.Bytes())
	bPubC, _ := ecdh.X25519().NewPublicKey(bkp.Public.Bytes())

	bond := &interopBond{
		backendID: bid,
		clientID:  cid,
		backendGo: crypto.NewPairingRecord(cid.InstanceID(), relayURL, bkp, cPubB),
		clientGo:  crypto.NewPairingRecord(bid.InstanceID(), relayURL, ckp, bPubC),
		backendCView: &PairingRecord{
			PeerInstanceID: cid.InstanceID(),
			RelayURL:       relayURL,
		},
		clientCView: &PairingRecord{
			PeerInstanceID: bid.InstanceID(),
			RelayURL:       relayURL,
		},
	}
	copy(bond.backendCView.LocalPrivKey[:], bkp.Private.Bytes())
	copy(bond.backendCView.LocalPubKey[:], bkp.Public.Bytes())
	copy(bond.backendCView.PeerPubKey[:], ckp.Public.Bytes())
	copy(bond.clientCView.LocalPrivKey[:], ckp.Private.Bytes())
	copy(bond.clientCView.LocalPubKey[:], ckp.Public.Bytes())
	copy(bond.clientCView.PeerPubKey[:], bkp.Public.Bytes())
	return bond
}

// --- C-as-backend ↔ Go-as-client (single client, four channels) ---

func TestInteropCBackendGoClient(t *testing.T) {
	relay := startRelay(t)
	defer relay.stop()

	bond := mintInteropBond(t, t.TempDir(), relay.url, "cb-gc")
	dgs := map[string]uint64{"ping": 1, "metric": 2}

	// Backend (C side, single goroutine — C SDK is single-threaded
	// per transport). Drains chat + control + ping + metric in
	// fixed order matched by the client; signals readiness through
	// backendReady and completion through backendDone.
	backendReady := make(chan struct{}, 1)
	backendDone := make(chan error, 1)
	backendShutdown := make(chan struct{})

	go func() {
		resolver := func(deviceID string) (*PairingRecord, bool) {
			if deviceID != bond.clientID.InstanceID() {
				return nil, false
			}
			return bond.backendCView, true
		}
		l, err := registerBackend(relay.host, relay.port,
			bond.backendID.InstanceID(), dgs, resolver)
		if err != nil {
			backendReady <- struct{}{}
			backendDone <- fmt.Errorf("registerBackend: %w", err)
			return
		}
		defer func() {
			<-backendShutdown
			l.Close()
		}()
		backendReady <- struct{}{}

		sess, err := l.Accept()
		if err != nil {
			backendDone <- fmt.Errorf("Accept: %w", err)
			return
		}

		// chat: echo
		chat, err := pumpAcceptStream(l, sess, "chat", 10*time.Second)
		if err != nil {
			backendDone <- fmt.Errorf("accept chat: %w", err)
			return
		}
		msg, err := chat.Recv()
		if err != nil {
			backendDone <- fmt.Errorf("chat Recv: %w", err)
			return
		}
		if err := chat.Send([]byte("echo: " + string(msg))); err != nil {
			backendDone <- fmt.Errorf("chat Send: %w", err)
			return
		}

		// control: ack
		ctrl, err := pumpAcceptStream(l, sess, "control", 10*time.Second)
		if err != nil {
			backendDone <- fmt.Errorf("accept control: %w", err)
			return
		}
		cmsg, err := ctrl.Recv()
		if err != nil {
			backendDone <- fmt.Errorf("control Recv: %w", err)
			return
		}
		if err := ctrl.Send([]byte("ack:" + string(cmsg))); err != nil {
			backendDone <- fmt.Errorf("control Send: %w", err)
			return
		}

		// ping datagram: pong
		pingDg, err := sess.datagram("ping")
		if err != nil {
			backendDone <- fmt.Errorf("get ping: %w", err)
			return
		}
		pmsg, err := pingDg.Recv()
		if err != nil {
			backendDone <- fmt.Errorf("ping Recv: %w", err)
			return
		}
		if err := pingDg.Send([]byte("pong:" + string(pmsg))); err != nil {
			backendDone <- fmt.Errorf("ping Send: %w", err)
			return
		}

		// metric datagram: n=<x>
		metricDg, err := sess.datagram("metric")
		if err != nil {
			backendDone <- fmt.Errorf("get metric: %w", err)
			return
		}
		mmsg, err := metricDg.Recv()
		if err != nil {
			backendDone <- fmt.Errorf("metric Recv: %w", err)
			return
		}
		if err := metricDg.Send([]byte("n=" + string(mmsg))); err != nil {
			backendDone <- fmt.Errorf("metric Send: %w", err)
			return
		}
		backendDone <- nil
	}()
	<-backendReady

	// Client (Go side via pigeon.Connect). Single goroutine drives
	// chat → control → ping → metric in that order so the C backend
	// can match the sequence without per-channel concurrency.
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	tlsCfg := &tls.Config{InsecureSkipVerify: true}

	sess, err := pigeon.Connect(ctx, &pigeon.ConnectArgs{
		InstanceID: bond.backendID.InstanceID(),
		Record:     bond.clientGo,
		Identity:   bond.clientID,
		Relay:      relay.url,
		TLS:        tlsCfg,
		Datagrams:  dgs,
	})
	if err != nil {
		t.Fatalf("pigeon.Connect: %v", err)
	}

	chat, err := sess.OpenStream(ctx, "chat")
	if err != nil {
		t.Fatalf("OpenStream(chat): %v", err)
	}
	if err := chat.Send([]byte("hello")); err != nil {
		t.Fatalf("chat Send: %v", err)
	}
	got, err := chat.Recv(ctx)
	if err != nil {
		t.Fatalf("chat Recv: %v", err)
	}
	if string(got) != "echo: hello" {
		t.Fatalf("chat: got %q want %q", got, "echo: hello")
	}

	ctrl, err := sess.OpenStream(ctx, "control")
	if err != nil {
		t.Fatalf("OpenStream(control): %v", err)
	}
	if err := ctrl.Send([]byte("ping")); err != nil {
		t.Fatalf("control Send: %v", err)
	}
	got, err = ctrl.Recv(ctx)
	if err != nil {
		t.Fatalf("control Recv: %v", err)
	}
	if string(got) != "ack:ping" {
		t.Fatalf("control: got %q want %q", got, "ack:ping")
	}

	ping := sess.Datagram("ping")
	if err := ping.Send([]byte("p1")); err != nil {
		t.Fatalf("ping Send: %v", err)
	}
	part, err := ping.Recv(ctx)
	if err != nil {
		t.Fatalf("ping Recv: %v", err)
	}
	if string(part.Payload) != "pong:p1" {
		t.Fatalf("ping: got %q want %q", part.Payload, "pong:p1")
	}

	metric := sess.Datagram("metric")
	if err := metric.Send([]byte("42")); err != nil {
		t.Fatalf("metric Send: %v", err)
	}
	part, err = metric.Recv(ctx)
	if err != nil {
		t.Fatalf("metric Recv: %v", err)
	}
	if string(part.Payload) != "n=42" {
		t.Fatalf("metric: got %q want %q", part.Payload, "n=42")
	}

	_ = sess.Close()
	close(backendShutdown)
	select {
	case err := <-backendDone:
		if err != nil {
			t.Fatalf("backend: %v", err)
		}
	case <-time.After(20 * time.Second):
		t.Fatalf("backend did not finish within 20s")
	}
}

// --- Go-as-backend ↔ C-as-client (single client, four channels) ---

func TestInteropGoBackendCClient(t *testing.T) {
	relay := startRelay(t)
	defer relay.stop()

	bond := mintInteropBond(t, t.TempDir(), relay.url, "gb-cc")
	dgs := map[string]uint64{"ping": 1, "metric": 2}

	// Backend (Go side via pigeon.Register). Each channel runs its
	// own goroutine — the Go side is concurrent-friendly. We track
	// per-channel completion so the test can detect missing replies.
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	tlsCfg := &tls.Config{InsecureSkipVerify: true}

	pairings := map[string]*crypto.PairingRecord{
		bond.clientID.InstanceID(): bond.backendGo,
	}
	listener, _, err := pigeon.Register(ctx, &pigeon.RegisterArgs{
		Identity: bond.backendID,
		Pairing: func(id string) (*crypto.PairingRecord, error) {
			rec, ok := pairings[id]
			if !ok {
				return nil, fmt.Errorf("unknown client %q", id)
			}
			return rec, nil
		},
		Relay:     relay.url,
		TLS:       tlsCfg,
		Datagrams: dgs,
	})
	if err != nil {
		t.Fatalf("pigeon.Register: %v", err)
	}
	defer listener.Close()

	backendDone := make(chan error, 1)
	go func() {
		bsess, err := listener.Accept(ctx)
		if err != nil {
			backendDone <- fmt.Errorf("Accept: %w", err)
			return
		}
		defer bsess.Close()

		// Chat echo loop.
		go func() {
			chat, err := bsess.AcceptStream(ctx, "chat")
			if err != nil {
				return
			}
			for {
				msg, err := chat.Recv(ctx)
				if err != nil {
					return
				}
				_ = chat.Send(append([]byte("echo: "), msg...))
			}
		}()

		// Control ack loop.
		go func() {
			ctrl, err := bsess.AcceptStream(ctx, "control")
			if err != nil {
				return
			}
			for {
				msg, err := ctrl.Recv(ctx)
				if err != nil {
					return
				}
				_ = ctrl.Send(append([]byte("ack:"), msg...))
			}
		}()

		// Ping pong loop.
		go func() {
			ping := bsess.Datagram("ping")
			for {
				msg, err := ping.Recv(ctx)
				if err != nil {
					return
				}
				_ = ping.Send(append([]byte("pong:"), msg.Payload...))
			}
		}()

		// Metric responder.
		go func() {
			metric := bsess.Datagram("metric")
			for {
				if _, err := metric.Recv(ctx); err != nil {
					return
				}
				_ = metric.Send([]byte("n=42"))
			}
		}()

		<-ctx.Done()
		backendDone <- nil
	}()

	// Client (C side via connectClient). Single goroutine, in-order
	// round-trip across all four channels.
	conn, err := connectClient(relay.host, relay.port,
		bond.backendID.InstanceID(), bond.clientID.InstanceID(),
		bond.clientCView, dgs)
	if err != nil {
		t.Fatalf("connectClient: %v", err)
	}
	csess := conn.Session()

	chat, err := csess.openStream("chat")
	if err != nil {
		t.Fatalf("openStream(chat): %v", err)
	}
	if err := chat.Send([]byte("hello")); err != nil {
		t.Fatalf("chat Send: %v", err)
	}
	got, err := chat.Recv()
	if err != nil {
		t.Fatalf("chat Recv: %v", err)
	}
	if string(got) != "echo: hello" {
		t.Fatalf("chat: got %q want %q", got, "echo: hello")
	}

	ctrl, err := csess.openStream("control")
	if err != nil {
		t.Fatalf("openStream(control): %v", err)
	}
	if err := ctrl.Send([]byte("ping")); err != nil {
		t.Fatalf("control Send: %v", err)
	}
	got, err = ctrl.Recv()
	if err != nil {
		t.Fatalf("control Recv: %v", err)
	}
	if string(got) != "ack:ping" {
		t.Fatalf("control: got %q want %q", got, "ack:ping")
	}

	ping, err := csess.datagram("ping")
	if err != nil {
		t.Fatalf("get ping: %v", err)
	}
	if err := ping.Send([]byte("p1")); err != nil {
		t.Fatalf("ping Send: %v", err)
	}
	got, err = ping.Recv()
	if err != nil {
		t.Fatalf("ping Recv: %v", err)
	}
	if string(got) != "pong:p1" {
		t.Fatalf("ping: got %q want %q", got, "pong:p1")
	}

	metric, err := csess.datagram("metric")
	if err != nil {
		t.Fatalf("get metric: %v", err)
	}
	if err := metric.Send([]byte("get")); err != nil {
		t.Fatalf("metric Send: %v", err)
	}
	got, err = metric.Recv()
	if err != nil {
		t.Fatalf("metric Recv: %v", err)
	}
	if string(got) != "n=42" {
		t.Fatalf("metric: got %q want %q", got, "n=42")
	}

	conn.Close()

	// Cancel the listener context to let the backend echo loops exit
	// cleanly. backendDone will fire when the deferred sess.Close runs.
	cancel()
	select {
	case <-backendDone:
	case <-time.After(5 * time.Second):
		// Backend may still be blocked on a Recv after the cancel
		// propagates; don't fail the test for that. The channel
		// round-trips above are the assertion.
	}
}

// --- Go-as-backend with two C-clients (clientTag demux) ---

func TestInteropGoBackendTwoCClients(t *testing.T) {
	relay := startRelay(t)
	defer relay.stop()

	const numClients = 2

	tmp := t.TempDir()
	bid, err := crypto.NewFileIdentity(tmp + "/twoc-backend.json")
	if err != nil {
		t.Fatalf("backend id: %v", err)
	}
	bkp, err := crypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("backend kp: %v", err)
	}

	type clientPair struct {
		id      crypto.Identity
		cView   *PairingRecord
		backRec *crypto.PairingRecord
	}
	clients := make([]clientPair, numClients)
	pairings := make(map[string]*crypto.PairingRecord, numClients)

	for i := range numClients {
		ci, err := crypto.NewFileIdentity(fmt.Sprintf("%s/twoc-client-%d.json", tmp, i))
		if err != nil {
			t.Fatalf("client %d id: %v", i, err)
		}
		ckp, err := crypto.GenerateKeyPair()
		if err != nil {
			t.Fatalf("client %d kp: %v", i, err)
		}
		cPubB, _ := ecdh.X25519().NewPublicKey(ckp.Public.Bytes())
		backRec := crypto.NewPairingRecord(ci.InstanceID(), relay.url, bkp, cPubB)
		// The C-side cv (csdke2e.PairingRecord) is built directly from
		// the raw key bytes below; no Go-side client view is needed
		// because the client is a C peer.

		cv := &PairingRecord{
			PeerInstanceID: bid.InstanceID(),
			RelayURL:       relay.url,
		}
		copy(cv.LocalPrivKey[:], ckp.Private.Bytes())
		copy(cv.LocalPubKey[:], ckp.Public.Bytes())
		copy(cv.PeerPubKey[:], bkp.Public.Bytes())

		clients[i] = clientPair{id: ci, cView: cv, backRec: backRec}
		pairings[ci.InstanceID()] = backRec
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	tlsCfg := &tls.Config{InsecureSkipVerify: true}

	listener, _, err := pigeon.Register(ctx, &pigeon.RegisterArgs{
		Identity: bid,
		Pairing: func(id string) (*crypto.PairingRecord, error) {
			r, ok := pairings[id]
			if !ok {
				return nil, fmt.Errorf("unknown client %q", id)
			}
			return r, nil
		},
		Relay: relay.url,
		TLS:   tlsCfg,
	})
	if err != nil {
		t.Fatalf("pigeon.Register: %v", err)
	}
	defer listener.Close()

	// Backend accept loop: spawn one echo goroutine per session.
	// The reply tags both the message and the peer ID so the test
	// can verify per-client demux.
	go func() {
		for {
			bsess, err := listener.Accept(ctx)
			if err != nil {
				return
			}
			go func(s *pigeon.Session) {
				defer s.Close()
				chat, err := s.AcceptStream(ctx, "chat")
				if err != nil {
					return
				}
				for {
					msg, err := chat.Recv(ctx)
					if err != nil {
						return
					}
					reply := fmt.Appendf(nil, "[%s] %s", s.PeerID(), msg)
					if err := chat.Send(reply); err != nil {
						return
					}
				}
			}(bsess)
		}
	}()

	// Two C clients drive in parallel.
	var wg sync.WaitGroup
	errs := make(chan error, numClients)
	for i := range numClients {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			conn, err := connectClient(relay.host, relay.port,
				bid.InstanceID(), clients[i].id.InstanceID(),
				clients[i].cView, nil)
			if err != nil {
				errs <- fmt.Errorf("client %d connect: %w", i, err)
				return
			}
			defer conn.Close()
			st, err := conn.Session().openStream("chat")
			if err != nil {
				errs <- fmt.Errorf("client %d openStream: %w", i, err)
				return
			}
			msg := fmt.Sprintf("c%d", i)
			if err := st.Send([]byte(msg)); err != nil {
				errs <- fmt.Errorf("client %d Send: %w", i, err)
				return
			}
			got, err := st.Recv()
			if err != nil {
				errs <- fmt.Errorf("client %d Recv: %w", i, err)
				return
			}
			want := fmt.Sprintf("[%s] %s", clients[i].id.InstanceID(), msg)
			if string(got) != want {
				errs <- fmt.Errorf("client %d cross-talk: got %q want %q", i, got, want)
				return
			}
		}(i)
	}
	wg.Wait()
	close(errs)
	for e := range errs {
		t.Error(e)
	}
}

// --- C-as-backend with two Go-clients (symmetric clientTag demux) ---

func TestInteropCBackendTwoGoClients(t *testing.T) {
	relay := startRelay(t)
	defer relay.stop()

	const numClients = 2

	tmp := t.TempDir()
	bid, err := crypto.NewFileIdentity(tmp + "/twog-backend.json")
	if err != nil {
		t.Fatalf("backend id: %v", err)
	}
	bkp, err := crypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("backend kp: %v", err)
	}

	type clientPair struct {
		id       crypto.Identity
		cv       *crypto.PairingRecord // client's own view, for pigeon.Connect
		backView *PairingRecord        // backend's view of this client, for the C resolver
	}
	clients := make([]clientPair, numClients)
	resolveTable := map[string]*PairingRecord{}

	for i := range numClients {
		ci, err := crypto.NewFileIdentity(fmt.Sprintf("%s/twog-client-%d.json", tmp, i))
		if err != nil {
			t.Fatalf("client %d id: %v", i, err)
		}
		ckp, err := crypto.GenerateKeyPair()
		if err != nil {
			t.Fatalf("client %d kp: %v", i, err)
		}
		cPubB, _ := ecdh.X25519().NewPublicKey(ckp.Public.Bytes())
		bPubC, _ := ecdh.X25519().NewPublicKey(bkp.Public.Bytes())
		cv := crypto.NewPairingRecord(bid.InstanceID(), relay.url, ckp, bPubC)

		bv := &PairingRecord{
			PeerInstanceID: ci.InstanceID(),
			RelayURL:       relay.url,
		}
		copy(bv.LocalPrivKey[:], bkp.Private.Bytes())
		copy(bv.LocalPubKey[:], bkp.Public.Bytes())
		copy(bv.PeerPubKey[:], cPubB.Bytes())

		clients[i] = clientPair{id: ci, cv: cv, backView: bv}
		resolveTable[ci.InstanceID()] = bv
	}

	// C-side backend pinned to one goroutine (single-threaded
	// transport invariant). Drives an N-deep sequential
	// Accept-then-echo loop, just like TestCSDKTwoConcurrentClients.
	backendReady := make(chan struct{}, 1)
	backendDone := make(chan error, 1)
	backendShutdown := make(chan struct{})

	go func() {
		resolver := func(deviceID string) (*PairingRecord, bool) {
			r, ok := resolveTable[deviceID]
			return r, ok
		}
		l, err := registerBackend(relay.host, relay.port,
			bid.InstanceID(), nil, resolver)
		if err != nil {
			backendReady <- struct{}{}
			backendDone <- fmt.Errorf("registerBackend: %w", err)
			return
		}
		defer func() {
			<-backendShutdown
			l.Close()
		}()
		backendReady <- struct{}{}

		for i := 0; i < numClients; i++ {
			sess, err := l.Accept()
			if err != nil {
				backendDone <- fmt.Errorf("Accept #%d: %w", i, err)
				return
			}
			st, err := pumpAcceptStream(l, sess, "chat", 15*time.Second)
			if err != nil {
				backendDone <- fmt.Errorf("pumpAcceptStream #%d: %w", i, err)
				return
			}
			got, err := st.Recv()
			if err != nil {
				backendDone <- fmt.Errorf("backend Recv #%d: %w", i, err)
				return
			}
			reply := fmt.Appendf(nil, "ack[%s]", got)
			if err := st.Send(reply); err != nil {
				backendDone <- fmt.Errorf("backend Send #%d: %w", i, err)
				return
			}
		}
		backendDone <- nil
	}()
	<-backendReady

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	tlsCfg := &tls.Config{InsecureSkipVerify: true}

	// Two Go clients connect concurrently. Each opens a "chat"
	// stream, sends a unique tag, expects the C backend's
	// ack[<tag>] reply. A mismatch indicates clientTag demux failure
	// on the C side.
	var wg sync.WaitGroup
	var ok atomic.Int32
	errs := make(chan error, numClients)
	for i := range numClients {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			sess, err := pigeon.Connect(ctx, &pigeon.ConnectArgs{
				InstanceID: bid.InstanceID(),
				Record:     clients[i].cv,
				Identity:   clients[i].id,
				Relay:      relay.url,
				TLS:        tlsCfg,
			})
			if err != nil {
				errs <- fmt.Errorf("client %d Connect: %w", i, err)
				return
			}
			defer sess.Close()
			chat, err := sess.OpenStream(ctx, "chat")
			if err != nil {
				errs <- fmt.Errorf("client %d OpenStream: %w", i, err)
				return
			}
			msg := fmt.Sprintf("c%d", i)
			if err := chat.Send([]byte(msg)); err != nil {
				errs <- fmt.Errorf("client %d Send: %w", i, err)
				return
			}
			got, err := chat.Recv(ctx)
			if err != nil {
				errs <- fmt.Errorf("client %d Recv: %w", i, err)
				return
			}
			want := fmt.Sprintf("ack[%s]", msg)
			if string(got) != want {
				errs <- fmt.Errorf("client %d cross-talk: got %q want %q", i, got, want)
				return
			}
			ok.Add(1)
		}(i)
	}
	wg.Wait()
	close(errs)
	for e := range errs {
		t.Error(e)
	}
	close(backendShutdown)
	select {
	case err := <-backendDone:
		if err != nil {
			t.Fatalf("backend: %v", err)
		}
	case <-time.After(20 * time.Second):
		t.Fatalf("backend did not finish within 20s")
	}
	if got := ok.Load(); got != numClients {
		t.Fatalf("expected %d clients to succeed; got %d", numClients, got)
	}
}

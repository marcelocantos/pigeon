// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

//go:build csdke2e

package csdke2e

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"math/big"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/marcelocantos/pigeon"
	"github.com/marcelocantos/pigeon/crypto"
)

// 🎯T32.4 live round-trips: the C SDK's pigeon_register /
// pigeon_connect / pigeon_listener_accept driving real QUIC traffic
// against an in-process Go relay (pigeon.NewQUICServer).
//
// Threading model: the C ABI is single-threaded per transport. Each
// test pins backend operations to one goroutine and each client to
// its own goroutine. The backend goroutine drives both the listener
// pump (pigeon_listener_step) and the session work (recv/send on
// streams and datagrams) on the same OS thread — pigeon_session
// shares the listener's transport, so any cross-goroutine call would
// race the ngtcp2 state.

// --- Fixtures (mirror e2e_test.go but tuned for the C SDK) ---

type relayFixture struct {
	host string
	port string
	url  string
	stop func()
}

func startRelay(t *testing.T) *relayFixture {
	t.Helper()
	cert := selfSignedCert(t)
	tlsCfg := &tls.Config{Certificates: []tls.Certificate{cert}}

	wtSrv, err := pigeon.NewWebTransportServer("127.0.0.1:0", tlsCfg, "")
	if err != nil {
		t.Fatalf("new wt server: %v", err)
	}

	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		t.Fatalf("listen udp: %v", err)
	}
	port := udpConn.LocalAddr().(*net.UDPAddr).Port
	qSrv := pigeon.NewQUICServer(fmt.Sprintf("127.0.0.1:%d", port), tlsCfg, "", wtSrv.Hub())
	go func() { _ = qSrv.ServeWithTLS(udpConn, tlsCfg) }()

	// Give the QUIC listener a moment to bind before clients dial.
	time.Sleep(200 * time.Millisecond)

	return &relayFixture{
		host: "127.0.0.1",
		port: fmt.Sprintf("%d", port),
		url:  fmt.Sprintf("https://127.0.0.1:%d", port),
		stop: func() {
			_ = qSrv.Close()
			_ = wtSrv.Close()
		},
	}
}

func selfSignedCert(t *testing.T) tls.Certificate {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	serial, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	notBefore := time.Now().Add(-time.Hour)
	tmpl := &x509.Certificate{
		SerialNumber: serial,
		DNSNames:     []string{"localhost"},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		NotBefore:    notBefore,
		NotAfter:     notBefore.Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create cert: %v", err)
	}
	return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key}
}

type bond struct {
	backendID   crypto.Identity
	clientID    crypto.Identity
	backendView *PairingRecord
	clientView  *PairingRecord
}

func mintBond(t *testing.T, tmp, relayURL, suffix string) *bond {
	t.Helper()
	bid, err := crypto.NewFileIdentity(tmp + "/backend-" + suffix + ".json")
	if err != nil {
		t.Fatalf("backend identity: %v", err)
	}
	cid, err := crypto.NewFileIdentity(tmp + "/client-" + suffix + ".json")
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

	bv := &PairingRecord{
		PeerInstanceID: cid.InstanceID(),
		RelayURL:       relayURL,
	}
	copy(bv.LocalPrivKey[:], bkp.Private.Bytes())
	copy(bv.LocalPubKey[:], bkp.Public.Bytes())
	copy(bv.PeerPubKey[:], ckp.Public.Bytes())

	cv := &PairingRecord{
		PeerInstanceID: bid.InstanceID(),
		RelayURL:       relayURL,
	}
	copy(cv.LocalPrivKey[:], ckp.Private.Bytes())
	copy(cv.LocalPubKey[:], ckp.Public.Bytes())
	copy(cv.PeerPubKey[:], bkp.Public.Bytes())

	return &bond{backendID: bid, clientID: cid, backendView: bv, clientView: cv}
}

// pumpAcceptStream drives pigeon_listener_step in a loop until the
// peer-opened sub-stream identified by `name` lands in the session's
// incoming-stream queue. Each step processes one inbound transport
// stream — either dispatching a sub-stream to a queue (rc=0) or
// surfacing an unexpected second primary (rc=1, which we treat as an
// error in tests that pre-baked the client count).
//
// The C SDK is single-threaded per transport, so all calls happen on
// the calling goroutine. transport_accept_stream has an internal
// 5-second timeout — in a healthy round-trip the sub-stream packets
// have already arrived by the time we call here, so step returns
// quickly. The deadline below is the all-up budget.
func pumpAcceptStream(l *listenerHandle, sess *sessionHandle, name string, deadline time.Duration) (*streamHandle, error) {
	end := time.Now().Add(deadline)
	for {
		if st, err := sess.acceptStream(name); err == nil {
			return st, nil
		}
		if time.Now().After(end) {
			return nil, fmt.Errorf("acceptStream(%q): deadline exceeded", name)
		}
		if err := l.step(); err != nil {
			return nil, fmt.Errorf("listener.step: %w", err)
		}
	}
}

// --- Test 1: single client ↔ single backend, named-stream round-trip ---

func TestCSDKSingleClientChat(t *testing.T) {
	relay := startRelay(t)
	defer relay.stop()

	p := mintBond(t, t.TempDir(), relay.url, "1")

	backendDone := make(chan error, 1)
	backendReady := make(chan struct{}, 1)
	backendShutdown := make(chan struct{})

	go runBackendSingle(t, relay, p, backendReady, backendShutdown, backendDone)
	<-backendReady // wait for the backend to be registered before connecting

	conn, err := connectClient(relay.host, relay.port,
		p.backendID.InstanceID(), p.clientID.InstanceID(),
		p.clientView, nil)
	if err != nil {
		t.Fatalf("connectClient: %v", err)
	}

	st, err := conn.Session().openStream("chat")
	if err != nil {
		t.Fatalf("client openStream: %v", err)
	}
	if err := st.Send([]byte("hello backend")); err != nil {
		t.Fatalf("client Send: %v", err)
	}
	got, err := st.Recv()
	if err != nil {
		t.Fatalf("client Recv: %v", err)
	}
	if string(got) != "hello client" {
		t.Fatalf("client got %q want %q", got, "hello client")
	}

	// Close the client BEFORE signalling backend shutdown so the
	// session's primary stream closes cleanly while the bridge is
	// still running, then let the backend tear down.
	conn.Close()
	close(backendShutdown)

	select {
	case err := <-backendDone:
		if err != nil {
			t.Fatalf("backend: %v", err)
		}
	case <-time.After(15 * time.Second):
		t.Fatalf("backend did not finish within 15s")
	}
}

// runBackendSingle pins the backend's listener + session work to this
// goroutine and signals completion via backendDone. The shutdown
// channel lets the test's main goroutine hold the listener open
// until the client side has confirmed all expected I/O — closing
// the listener prematurely tears the QUIC connection down before
// the relay can finish flushing the reply bytes to the client.
func runBackendSingle(t *testing.T, relay *relayFixture, p *bond, ready chan<- struct{}, shutdown <-chan struct{}, done chan<- error) {
	resolver := func(deviceID string) (*PairingRecord, bool) {
		if deviceID != p.clientID.InstanceID() {
			return nil, false
		}
		return p.backendView, true
	}
	l, err := registerBackend(relay.host, relay.port,
		p.backendID.InstanceID(), nil, resolver)
	if err != nil {
		ready <- struct{}{}
		done <- fmt.Errorf("registerBackend: %w", err)
		return
	}
	defer func() {
		<-shutdown
		l.Close()
	}()
	ready <- struct{}{} // backend is registered; client may now connect

	sess, err := l.Accept()
	if err != nil {
		done <- fmt.Errorf("listener Accept: %w", err)
		return
	}
	st, err := pumpAcceptStream(l, sess, "chat", 10*time.Second)
	if err != nil {
		done <- fmt.Errorf("backend pumpAcceptStream: %w", err)
		return
	}
	got, err := st.Recv()
	if err != nil {
		done <- fmt.Errorf("backend Recv: %w", err)
		return
	}
	if string(got) != "hello backend" {
		done <- fmt.Errorf("backend got %q want %q", got, "hello backend")
		return
	}
	if err := st.Send([]byte("hello client")); err != nil {
		done <- fmt.Errorf("backend Send: %w", err)
		return
	}
	done <- nil
}

// --- Test 2: one session, two named streams + two datagram channels ---

func TestCSDKMultiStreamMultiDatagram(t *testing.T) {
	relay := startRelay(t)
	defer relay.stop()

	p := mintBond(t, t.TempDir(), relay.url, "2")
	dgs := map[string]uint64{"ping": 1, "metric": 2}

	backendDone := make(chan error, 1)
	backendReady := make(chan struct{}, 1)
	backendShutdown := make(chan struct{})
	chatReady := make(chan struct{}, 1)

	go runBackendMulti(t, relay, p, dgs, backendReady, backendShutdown, chatReady, backendDone)
	<-backendReady

	conn, err := connectClient(relay.host, relay.port,
		p.backendID.InstanceID(), p.clientID.InstanceID(),
		p.clientView, dgs)
	if err != nil {
		t.Fatalf("connectClient: %v", err)
	}
	sess := conn.Session()

	// --- chat round-trip ---
	chat, err := sess.openStream("chat")
	if err != nil {
		t.Fatalf("openStream(chat): %v", err)
	}
	if err := chat.Send([]byte("hi")); err != nil {
		t.Fatalf("chat Send: %v", err)
	}
	gotChat, err := chat.Recv()
	if err != nil {
		t.Fatalf("chat Recv: %v", err)
	}
	if string(gotChat) != "echo: hi" {
		t.Fatalf("chat got %q want %q", gotChat, "echo: hi")
	}
	chatReady <- struct{}{} // signal backend to open control phase

	// --- control round-trip ---
	ctrl, err := sess.openStream("control")
	if err != nil {
		t.Fatalf("openStream(control): %v", err)
	}
	if err := ctrl.Send([]byte("ping")); err != nil {
		t.Fatalf("control Send: %v", err)
	}
	gotCtrl, err := ctrl.Recv()
	if err != nil {
		t.Fatalf("control Recv: %v", err)
	}
	if string(gotCtrl) != "ack:ping" {
		t.Fatalf("control got %q want %q", gotCtrl, "ack:ping")
	}

	// --- ping datagram round-trip (strictly serialised with metric) ---
	pingDg, err := sess.datagram("ping")
	if err != nil {
		t.Fatalf("get ping: %v", err)
	}
	if err := pingDg.Send([]byte("p1")); err != nil {
		t.Fatalf("ping Send: %v", err)
	}
	gotPing, err := pingDg.Recv()
	if err != nil {
		t.Fatalf("ping Recv: %v", err)
	}
	if string(gotPing) != "pong:p1" {
		t.Fatalf("ping got %q want %q", gotPing, "pong:p1")
	}

	// --- metric datagram round-trip ---
	metricDg, err := sess.datagram("metric")
	if err != nil {
		t.Fatalf("get metric: %v", err)
	}
	if err := metricDg.Send([]byte("42")); err != nil {
		t.Fatalf("metric Send: %v", err)
	}
	gotMetric, err := metricDg.Recv()
	if err != nil {
		t.Fatalf("metric Recv: %v", err)
	}
	if string(gotMetric) != "n=42" {
		t.Fatalf("metric got %q want %q", gotMetric, "n=42")
	}

	// All assertions complete — close the client connection
	// cleanly while the bridge is still alive, then release the
	// backend's listener shutdown.
	conn.Close()
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

func runBackendMulti(t *testing.T, relay *relayFixture, p *bond, dgs map[string]uint64, ready chan<- struct{}, shutdown <-chan struct{}, chatReady <-chan struct{}, done chan<- error) {
	resolver := func(deviceID string) (*PairingRecord, bool) {
		if deviceID != p.clientID.InstanceID() {
			return nil, false
		}
		return p.backendView, true
	}
	l, err := registerBackend(relay.host, relay.port,
		p.backendID.InstanceID(), dgs, resolver)
	if err != nil {
		ready <- struct{}{}
		done <- fmt.Errorf("registerBackend: %w", err)
		return
	}
	defer func() {
		<-shutdown
		l.Close()
	}()
	ready <- struct{}{}

	sess, err := l.Accept()
	if err != nil {
		done <- fmt.Errorf("Accept: %w", err)
		return
	}

	// chat: echo with prefix.
	chat, err := pumpAcceptStream(l, sess, "chat", 10*time.Second)
	if err != nil {
		done <- fmt.Errorf("accept chat: %w", err)
		return
	}
	got, err := chat.Recv()
	if err != nil {
		done <- fmt.Errorf("chat Recv: %w", err)
		return
	}
	if err := chat.Send([]byte("echo: " + string(got))); err != nil {
		done <- fmt.Errorf("chat Send: %w", err)
		return
	}

	<-chatReady // pause so the client can verify chat before control opens

	// control: respond ack:msg.
	ctrl, err := pumpAcceptStream(l, sess, "control", 10*time.Second)
	if err != nil {
		done <- fmt.Errorf("accept control: %w", err)
		return
	}
	got, err = ctrl.Recv()
	if err != nil {
		done <- fmt.Errorf("control Recv: %w", err)
		return
	}
	if err := ctrl.Send([]byte("ack:" + string(got))); err != nil {
		done <- fmt.Errorf("control Send: %w", err)
		return
	}

	// ping datagram round-trip (matched 1-for-1 with client).
	pingDg, err := sess.datagram("ping")
	if err != nil {
		done <- fmt.Errorf("get ping: %w", err)
		return
	}
	pmsg, err := pingDg.Recv()
	if err != nil {
		done <- fmt.Errorf("ping Recv: %w", err)
		return
	}
	if err := pingDg.Send([]byte("pong:" + string(pmsg))); err != nil {
		done <- fmt.Errorf("ping Send: %w", err)
		return
	}

	// metric datagram round-trip.
	metricDg, err := sess.datagram("metric")
	if err != nil {
		done <- fmt.Errorf("get metric: %w", err)
		return
	}
	mmsg, err := metricDg.Recv()
	if err != nil {
		done <- fmt.Errorf("metric Recv: %w", err)
		return
	}
	if err := metricDg.Send([]byte("n=" + string(mmsg))); err != nil {
		done <- fmt.Errorf("metric Send: %w", err)
		return
	}

	done <- nil
}

// --- Test 3: one backend, two concurrent clients, demux by clientTag ---

func TestCSDKTwoConcurrentClients(t *testing.T) {
	relay := startRelay(t)
	defer relay.stop()

	const numClients = 2

	tmp := t.TempDir()
	bid, err := crypto.NewFileIdentity(tmp + "/backend.json")
	if err != nil {
		t.Fatalf("backend id: %v", err)
	}
	bkp, err := crypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("backend kp: %v", err)
	}

	type clientBond struct {
		id     crypto.Identity
		client *PairingRecord
	}
	clients := make([]clientBond, numClients)
	resolveTable := map[string]*PairingRecord{}

	for i := range numClients {
		ci, err := crypto.NewFileIdentity(fmt.Sprintf("%s/client-%d.json", tmp, i))
		if err != nil {
			t.Fatalf("client %d id: %v", i, err)
		}
		ckp, err := crypto.GenerateKeyPair()
		if err != nil {
			t.Fatalf("client %d kp: %v", i, err)
		}
		bv := &PairingRecord{
			PeerInstanceID: ci.InstanceID(),
			RelayURL:       relay.url,
		}
		copy(bv.LocalPrivKey[:], bkp.Private.Bytes())
		copy(bv.LocalPubKey[:], bkp.Public.Bytes())
		copy(bv.PeerPubKey[:], ckp.Public.Bytes())

		cv := &PairingRecord{
			PeerInstanceID: bid.InstanceID(),
			RelayURL:       relay.url,
		}
		copy(cv.LocalPrivKey[:], ckp.Private.Bytes())
		copy(cv.LocalPubKey[:], ckp.Public.Bytes())
		copy(cv.PeerPubKey[:], bkp.Public.Bytes())

		clients[i] = clientBond{id: ci, client: cv}
		resolveTable[ci.InstanceID()] = bv
	}

	backendDone := make(chan error, 1)
	backendReady := make(chan struct{}, 1)
	backendShutdown := make(chan struct{})
	go runBackendTwoClients(t, relay, bid, resolveTable, numClients, backendReady, backendShutdown, backendDone)
	<-backendReady

	// Spin up both clients concurrently. Each opens its own
	// pigeon_connection (separate ngtcp2 transport per goroutine —
	// no per-client race) and round-trips one message on "chat".
	// The backend serves them sequentially in tag-arrival order
	// and tags the reply with the message so we can verify per-
	// client demux didn't cross-talk.
	var wg sync.WaitGroup
	errs := make(chan error, numClients)
	for i := range numClients {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			conn, err := connectClient(relay.host, relay.port,
				bid.InstanceID(), clients[i].id.InstanceID(),
				clients[i].client, nil)
			if err != nil {
				errs <- fmt.Errorf("client %d connect: %w", i, err)
				return
			}

			st, err := conn.Session().openStream("chat")
			if err != nil {
				conn.Close()
				errs <- fmt.Errorf("client %d openStream: %w", i, err)
				return
			}
			msg := fmt.Sprintf("c%d", i)
			if err := st.Send([]byte(msg)); err != nil {
				conn.Close()
				errs <- fmt.Errorf("client %d Send: %w", i, err)
				return
			}
			got, err := st.Recv()
			if err != nil {
				conn.Close()
				errs <- fmt.Errorf("client %d Recv: %w", i, err)
				return
			}
			conn.Close()
			want := fmt.Sprintf("ack[%s]", msg)
			if string(got) != want {
				errs <- fmt.Errorf("client %d got %q want %q (cross-talk between clients)", i, got, want)
				return
			}
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
}

// runBackendTwoClients single-threaded backend driver. Holds the
// listener and serves N clients in tag-arrival order. The clients
// connect and send on their own goroutines; the backend's pump
// state-machine accepts each session, drains its "chat" sub-stream
// (which may be already buffered if the client raced ahead), and
// echoes back a tagged reply.
func runBackendTwoClients(t *testing.T, relay *relayFixture, bid crypto.Identity, resolveTable map[string]*PairingRecord, n int, ready chan<- struct{}, shutdown <-chan struct{}, done chan<- error) {
	resolver := func(deviceID string) (*PairingRecord, bool) {
		rec, ok := resolveTable[deviceID]
		return rec, ok
	}
	l, err := registerBackend(relay.host, relay.port,
		bid.InstanceID(), nil, resolver)
	if err != nil {
		ready <- struct{}{}
		done <- fmt.Errorf("registerBackend: %w", err)
		return
	}
	defer func() {
		<-shutdown
		l.Close()
	}()
	ready <- struct{}{}

	for i := 0; i < n; i++ {
		sess, err := l.Accept()
		if err != nil {
			done <- fmt.Errorf("Accept #%d: %w", i, err)
			return
		}
		st, err := pumpAcceptStream(l, sess, "chat", 15*time.Second)
		if err != nil {
			done <- fmt.Errorf("pumpAcceptStream #%d: %w", i, err)
			return
		}
		got, err := st.Recv()
		if err != nil {
			done <- fmt.Errorf("backend Recv #%d: %w", i, err)
			return
		}
		reply := fmt.Appendf(nil, "ack[%s]", got)
		if err := st.Send(reply); err != nil {
			done <- fmt.Errorf("backend Send #%d: %w", i, err)
			return
		}
	}
	done <- nil
}

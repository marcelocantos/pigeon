// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package pigeon_test

import (
	"context"
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"math/big"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/marcelocantos/pigeon"
	"github.com/marcelocantos/pigeon/crypto"
)

// startRelay brings up a raw QUIC pigeon relay on a random port and
// returns the relay URL plus a teardown function. The TLS config uses
// a self-signed cert that the test clients accept via InsecureSkipVerify.
func startRelay(t *testing.T) (relayURL string, teardown func()) {
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

	go qSrv.ServeWithTLS(udpConn, tlsCfg)

	teardown = func() {
		_ = qSrv.Close()
		_ = wtSrv.Close()
	}
	return fmt.Sprintf("https://127.0.0.1:%d", port), teardown
}

// selfSignedCert creates a fresh self-signed TLS certificate.
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

// pairingPair generates a fresh pair of crypto.Identity-backed
// PairingRecords representing a completed pairing between a backend
// and a client. Skips the ceremony — the test mints both records
// directly from a shared ECDH exchange.
func pairingPair(t *testing.T, relayURL string) (backendID crypto.Identity, clientID crypto.Identity, backendRec, clientRec *crypto.PairingRecord) {
	t.Helper()
	tmp := t.TempDir()
	bid, err := crypto.NewFileIdentity(tmp + "/backend-id.json")
	if err != nil {
		t.Fatalf("backend identity: %v", err)
	}
	cid, err := crypto.NewFileIdentity(tmp + "/client-id.json")
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

	backendRec = crypto.NewPairingRecord(cid.InstanceID(), relayURL, bkp, cPubB)
	clientRec = crypto.NewPairingRecord(bid.InstanceID(), relayURL, ckp, bPubC)
	return bid, cid, backendRec, clientRec
}

// TestSessionMachineLifecycle proves the per-Session SessionMachine
// (T39.4) is wired through the activation flow and the disconnect
// transition: post-Connect / post-Accept the machine sits at the
// spec's RelayConnected state on both sides; post-Close it has
// advanced through the spec's RelayConnected → Paired transition.
func TestSessionMachineLifecycle(t *testing.T) {
	t.Parallel()
	relayURL, teardown := startRelay(t)
	defer teardown()
	time.Sleep(200 * time.Millisecond)

	bid, cid, brec, crec := pairingPair(t, relayURL)
	pairings := map[string]*crypto.PairingRecord{cid.InstanceID(): brec}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	tlsCfg := &tls.Config{InsecureSkipVerify: true}

	listener, _, err := pigeon.Register(ctx, &pigeon.RegisterArgs{
		Identity: bid,
		Pairing: func(id string) (*crypto.PairingRecord, error) {
			rec, ok := pairings[id]
			if !ok {
				return nil, fmt.Errorf("unknown client %q", id)
			}
			return rec, nil
		},
		Relay: relayURL,
		TLS:   tlsCfg,
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	defer listener.Close()

	bsess := make(chan *pigeon.Session, 1)
	go func() {
		s, err := listener.Accept(ctx)
		if err != nil {
			t.Errorf("accept: %v", err)
			bsess <- nil
			return
		}
		bsess <- s
	}()

	csess, err := pigeon.Connect(ctx, &pigeon.ConnectArgs{
		InstanceID: bid.InstanceID(),
		Record:     crec,
		Identity:   cid,
		Relay:      relayURL,
		TLS:        tlsCfg,
	})
	if err != nil {
		t.Fatalf("connect: %v", err)
	}

	bs := <-bsess
	if bs == nil {
		t.Fatal("backend accept returned nil")
	}

	const wantConnected = "RelayConnected"
	if got := string(csess.MachineState()); got != wantConnected {
		t.Errorf("client machine post-Connect: got %q, want %q", got, wantConnected)
	}
	if got := string(bs.MachineState()); got != wantConnected {
		t.Errorf("backend machine post-Accept: got %q, want %q", got, wantConnected)
	}

	_ = csess.Close()
	_ = bs.Close()

	const wantClosed = "Paired"
	if got := string(csess.MachineState()); got != wantClosed {
		t.Errorf("client machine post-Close: got %q, want %q (RelayConnected → Paired on disconnect)", got, wantClosed)
	}
	if got := string(bs.MachineState()); got != wantClosed {
		t.Errorf("backend machine post-Close: got %q, want %q (RelayConnected → Paired on disconnect)", got, wantClosed)
	}
}

// TestE2ESingleClientChat verifies the new pigeon.Register / pigeon.Connect
// shape and the multi-stream channel API end-to-end.
func TestE2ESingleClientChat(t *testing.T) {
	t.Parallel()
	relayURL, teardown := startRelay(t)
	defer teardown()
	time.Sleep(200 * time.Millisecond) // let the relay bind

	bid, cid, brec, crec := pairingPair(t, relayURL)

	pairings := map[string]*crypto.PairingRecord{cid.InstanceID(): brec}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	tlsCfg := &tls.Config{InsecureSkipVerify: true}

	listener, _, err := pigeon.Register(ctx, &pigeon.RegisterArgs{
		Identity: bid,
		Pairing: func(clientInstanceID string) (*crypto.PairingRecord, error) {
			rec, ok := pairings[clientInstanceID]
			if !ok {
				return nil, fmt.Errorf("unknown client %q", clientInstanceID)
			}
			return rec, nil
		},
		Relay: relayURL,
		TLS:   tlsCfg,
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	defer listener.Close()

	// Backend goroutine: accept one client, echo on chat until the
	// client closes the stream (EOF). Driving the lifecycle from the
	// client side avoids racing the backend's deferred Session.Close
	// against the last echo's bytes still in flight on the QUIC stream.
	bdone := make(chan error, 1)
	go func() {
		sess, err := listener.Accept(ctx)
		if err != nil {
			bdone <- fmt.Errorf("accept: %w", err)
			return
		}
		defer sess.Close()
		chat, err := sess.AcceptStream(ctx, "chat")
		if err != nil {
			bdone <- fmt.Errorf("accept chat: %w", err)
			return
		}
		for {
			msg, err := chat.Recv(ctx)
			if err != nil {
				// EOF / cancelled / context-done: the client is
				// finishing up. The N successful round-trips before
				// this point are what the test asserts on.
				bdone <- nil
				return
			}
			if err := chat.Send(append([]byte("echo: "), msg...)); err != nil {
				bdone <- fmt.Errorf("send: %w", err)
				return
			}
		}
	}()

	sess, err := pigeon.Connect(ctx, &pigeon.ConnectArgs{
		InstanceID: bid.InstanceID(),
		Record:     crec,
		Identity:   cid,
		Relay:      relayURL,
		TLS:        tlsCfg,
	})
	if err != nil {
		t.Fatalf("connect: %v", err)
	}

	chat, err := sess.OpenStream(ctx, "chat")
	if err != nil {
		t.Fatalf("open chat: %v", err)
	}

	for _, want := range []string{"echo: hello", "echo: world", "echo: bye"} {
		msg := want[len("echo: "):]
		if err := chat.Send([]byte(msg)); err != nil {
			t.Fatalf("send %q: %v", msg, err)
		}
		got, err := chat.Recv(ctx)
		if err != nil {
			t.Fatalf("recv %q: %v", msg, err)
		}
		if string(got) != want {
			t.Fatalf("recv %q: got %q want %q", msg, got, want)
		}
	}

	// Closing the client session signals EOF to the backend's Recv
	// loop, which then exits and writes bdone. We then wait on bdone,
	// guaranteeing the backend has observed the close *after* the
	// last echo was successfully delivered to us.
	_ = sess.Close()

	if err := <-bdone; err != nil {
		t.Fatalf("backend: %v", err)
	}
}

// TestE2EMultiStreamMultiDatagram exercises 2 stream channels (chat,
// control) and 2 datagram channels (ping, metric) over a single client
// session.
func TestE2EMultiStreamMultiDatagram(t *testing.T) {
	t.Parallel()
	relayURL, teardown := startRelay(t)
	defer teardown()
	time.Sleep(200 * time.Millisecond)

	bid, cid, brec, crec := pairingPair(t, relayURL)
	pairings := map[string]*crypto.PairingRecord{cid.InstanceID(): brec}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	tlsCfg := &tls.Config{InsecureSkipVerify: true}

	dgChannels := map[string]uint64{"ping": 1, "metric": 2}

	listener, _, err := pigeon.Register(ctx, &pigeon.RegisterArgs{
		Identity:  bid,
		Pairing:   func(s string) (*crypto.PairingRecord, error) { return pairings[s], nil },
		Relay:     relayURL,
		TLS:       tlsCfg,
		Datagrams: dgChannels,
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	defer listener.Close()

	bdone := make(chan error, 1)
	go func() {
		sess, err := listener.Accept(ctx)
		if err != nil {
			bdone <- err
			return
		}
		defer sess.Close()

		// chat: echo
		go func() {
			chat, err := sess.AcceptStream(ctx, "chat")
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

		// control: respond "stats" with count of echoes (atomic — see
		// the ping goroutine below that increments concurrently).
		var echoes atomic.Int64
		go func() {
			ctrl, err := sess.AcceptStream(ctx, "control")
			if err != nil {
				return
			}
			for {
				msg, err := ctrl.Recv(ctx)
				if err != nil {
					return
				}
				if string(msg) == "stats" {
					_ = ctrl.Send(fmt.Appendf(nil, "echoes=%d", echoes.Load()))
				}
			}
		}()

		// ping datagram: echo back; metric datagram: respond with "n=42"
		go func() {
			pingCh := sess.Datagram("ping")
			for {
				p, err := pingCh.Recv(ctx)
				if err != nil {
					return
				}
				_ = pingCh.Send(append([]byte("pong:"), p...))
				echoes.Add(1)
			}
		}()
		go func() {
			metric := sess.Datagram("metric")
			for {
				_, err := metric.Recv(ctx)
				if err != nil {
					return
				}
				_ = metric.Send([]byte("n=42"))
			}
		}()

		<-ctx.Done()
		bdone <- nil
	}()

	sess, err := pigeon.Connect(ctx, &pigeon.ConnectArgs{
		InstanceID: bid.InstanceID(),
		Record:     crec,
		Identity:   cid,
		Relay:      relayURL,
		TLS:        tlsCfg,
		Datagrams:  dgChannels,
	})
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer sess.Close()

	chat, err := sess.OpenStream(ctx, "chat")
	if err != nil {
		t.Fatalf("open chat: %v", err)
	}
	ctrl, err := sess.OpenStream(ctx, "control")
	if err != nil {
		t.Fatalf("open control: %v", err)
	}
	ping := sess.Datagram("ping")
	metric := sess.Datagram("metric")

	// Round-trip on chat.
	if err := chat.Send([]byte("hello")); err != nil {
		t.Fatalf("chat send: %v", err)
	}
	got, err := chat.Recv(ctx)
	if err != nil {
		t.Fatalf("chat recv: %v", err)
	}
	if string(got) != "echo: hello" {
		t.Fatalf("chat: got %q", got)
	}

	// Round-trip on ping (datagram).
	if err := ping.Send([]byte("p1")); err != nil {
		t.Fatalf("ping send: %v", err)
	}
	got, err = ping.Recv(ctx)
	if err != nil {
		t.Fatalf("ping recv: %v", err)
	}
	if string(got) != "pong:p1" {
		t.Fatalf("ping: got %q", got)
	}

	// Round-trip on metric (datagram).
	if err := metric.Send([]byte("get")); err != nil {
		t.Fatalf("metric send: %v", err)
	}
	got, err = metric.Recv(ctx)
	if err != nil {
		t.Fatalf("metric recv: %v", err)
	}
	if string(got) != "n=42" {
		t.Fatalf("metric: got %q", got)
	}

	// Round-trip on control. Backend's echo count should now be 1
	// (one ping echoed). Wait briefly for the ping echo to update the
	// counter on the backend before asking.
	time.Sleep(50 * time.Millisecond)
	if err := ctrl.Send([]byte("stats")); err != nil {
		t.Fatalf("ctrl send: %v", err)
	}
	got, err = ctrl.Recv(ctx)
	if err != nil {
		t.Fatalf("ctrl recv: %v", err)
	}
	if string(got) != "echoes=1" {
		t.Fatalf("ctrl: got %q want echoes=1", got)
	}
}

// TestE2EMultiClient brings up two clients against the same backend and
// verifies they can chat independently without interfering.
func TestE2EMultiClient(t *testing.T) {
	t.Parallel()
	relayURL, teardown := startRelay(t)
	defer teardown()
	time.Sleep(200 * time.Millisecond)

	const numClients = 2

	bid, _, _, _ := pairingPair(t, relayURL)
	clientIDs := make([]crypto.Identity, numClients)
	clientRecs := make([]*crypto.PairingRecord, numClients)
	pairings := make(map[string]*crypto.PairingRecord, numClients)

	tmp := t.TempDir()
	bkp, err := crypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("backend kp: %v", err)
	}

	for i := range numClients {
		ci, err := crypto.NewFileIdentity(fmt.Sprintf("%s/client-%d.json", tmp, i))
		if err != nil {
			t.Fatalf("client %d identity: %v", i, err)
		}
		ckp, err := crypto.GenerateKeyPair()
		if err != nil {
			t.Fatalf("client %d kp: %v", i, err)
		}
		cPubB, _ := ecdh.X25519().NewPublicKey(ckp.Public.Bytes())
		bPubC, _ := ecdh.X25519().NewPublicKey(bkp.Public.Bytes())
		brec := crypto.NewPairingRecord(ci.InstanceID(), relayURL, bkp, cPubB)
		crec := crypto.NewPairingRecord(bid.InstanceID(), relayURL, ckp, bPubC)
		clientIDs[i] = ci
		clientRecs[i] = crec
		pairings[ci.InstanceID()] = brec
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	tlsCfg := &tls.Config{InsecureSkipVerify: true}

	listener, _, err := pigeon.Register(ctx, &pigeon.RegisterArgs{
		Identity: bid,
		Pairing:  func(s string) (*crypto.PairingRecord, error) { return pairings[s], nil },
		Relay:    relayURL,
		TLS:      tlsCfg,
	})
	if err != nil {
		t.Fatalf("register: %v", err)
	}
	defer listener.Close()

	// Backend accept loop: echoes on chat, tagging with peer ID prefix
	// so the test can verify the right session got the right traffic.
	go func() {
		for {
			sess, err := listener.Accept(ctx)
			if err != nil {
				return
			}
			go func(sess *pigeon.Session) {
				defer sess.Close()
				chat, err := sess.AcceptStream(ctx, "chat")
				if err != nil {
					return
				}
				for {
					msg, err := chat.Recv(ctx)
					if err != nil {
						return
					}
					reply := fmt.Appendf(nil, "[%s] %s", sess.PeerID(), msg)
					if err := chat.Send(reply); err != nil {
						return
					}
				}
			}(sess)
		}
	}()

	// Drive both clients concurrently; each sends 3 messages and
	// verifies the echoes come back correctly.
	var wg sync.WaitGroup
	errs := make(chan error, numClients)
	for i := range numClients {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			sess, err := pigeon.Connect(ctx, &pigeon.ConnectArgs{
				InstanceID: bid.InstanceID(),
				Record:     clientRecs[i],
				Identity:   clientIDs[i],
				Relay:      relayURL,
				TLS:        tlsCfg,
			})
			if err != nil {
				errs <- fmt.Errorf("client %d connect: %w", i, err)
				return
			}
			defer sess.Close()
			chat, err := sess.OpenStream(ctx, "chat")
			if err != nil {
				errs <- fmt.Errorf("client %d open chat: %w", i, err)
				return
			}
			for j := range 3 {
				msg := fmt.Sprintf("c%d-%d", i, j)
				if err := chat.Send([]byte(msg)); err != nil {
					errs <- fmt.Errorf("client %d send %d: %w", i, j, err)
					return
				}
				got, err := chat.Recv(ctx)
				if err != nil {
					errs <- fmt.Errorf("client %d recv %d: %w", i, j, err)
					return
				}
				want := fmt.Sprintf("[%s] %s", clientIDs[i].InstanceID(), msg)
				if string(got) != want {
					errs <- fmt.Errorf("client %d msg %d: got %q want %q", i, j, got, want)
					return
				}
			}
		}(i)
	}
	wg.Wait()
	close(errs)
	for e := range errs {
		t.Error(e)
	}
}

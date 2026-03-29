// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package tern

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"math/big"
	"net"
	"net/url"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/marcelocantos/tern/crypto"
)

// generateTestCert creates a self-signed TLS certificate for testing.
func generateTestCert(t testing.TB) (tls.Certificate, *x509.CertPool) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		DNSNames:     []string{"localhost"},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}

	cert := tls.Certificate{
		Certificate: [][]byte{certDER},
		PrivateKey:  key,
	}

	pool := x509.NewCertPool()
	parsedCert, err := x509.ParseCertificate(certDER)
	if err != nil {
		t.Fatal(err)
	}
	pool.AddCert(parsedCert)

	return cert, pool
}

// relayEnv holds the URL and options needed to connect to a relay.
type relayEnv struct {
	url  string
	opts []Option
}

// localRelay starts a test relay (WebTransport + raw QUIC) and returns
// a relayEnv configured for raw QUIC connections.
func localRelay(t *testing.T) relayEnv {
	t.Helper()
	// If an external instrumented relay is running (e.g. for coverage),
	// use it instead of starting our own.
	if qURL := os.Getenv("TERN_TEST_QUIC_URL"); qURL != "" {
		u, _ := url.Parse(qURL)
		return relayEnv{
			url: qURL,
			opts: []Option{
				WithTLS(&tls.Config{InsecureSkipVerify: true}),
				WithQUICPort(u.Port()),
			},
		}
	}
	return localRelayTB(t)
}

// localRelayTB is the shared implementation for tests and benchmarks.
func localRelayTB(t testing.TB) relayEnv {
	t.Helper()

	cert, pool := generateTestCert(t)
	tlsCfg := &tls.Config{Certificates: []tls.Certificate{cert}}

	srv, err := NewWebTransportServer("127.0.0.1:0", tlsCfg, "")
	if err != nil {
		t.Fatal(err)
	}

	// Start WebTransport server.
	wtUDP, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatal(err)
	}
	go srv.Serve(wtUDP)
	t.Cleanup(func() { srv.Close() })

	// Start raw QUIC server sharing the same hub.
	qsrv := NewQUICServer("127.0.0.1:0", tlsCfg, "", srv.Hub())
	qUDP, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatal(err)
	}
	go qsrv.ServeWithTLS(qUDP, tlsCfg)
	t.Cleanup(func() { qsrv.Close() })

	wtPort := wtUDP.LocalAddr().(*net.UDPAddr).Port
	qPort := qUDP.LocalAddr().(*net.UDPAddr).Port

	// Default: raw QUIC. The URL host is used by both WT and QUIC paths.
	return relayEnv{
		url: "https://127.0.0.1:" + strconv.Itoa(wtPort),
		opts: []Option{
			WithTLS(&tls.Config{RootCAs: pool}),
			WithQUICPort(strconv.Itoa(qPort)),
		},
	}
}

// localRelayWT starts a test relay and returns a relayEnv configured
// for WebTransport connections (backward compat / browser path).
func localRelayWT(t *testing.T) relayEnv {
	t.Helper()

	// If an external instrumented relay is running, use it.
	if wtURL := os.Getenv("TERN_TEST_WT_URL"); wtURL != "" {
		return relayEnv{
			url: wtURL,
			opts: []Option{
				WithTLS(&tls.Config{InsecureSkipVerify: true}),
				WithWebTransport(),
			},
		}
	}

	cert, pool := generateTestCert(t)
	tlsCfg := &tls.Config{Certificates: []tls.Certificate{cert}}

	srv, err := NewWebTransportServer("127.0.0.1:0", tlsCfg, "")
	if err != nil {
		t.Fatal(err)
	}

	udpAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		t.Fatal(err)
	}

	go srv.Serve(conn)
	t.Cleanup(func() { srv.Close() })

	addr := conn.LocalAddr().(*net.UDPAddr)
	return relayEnv{
		url: "https://127.0.0.1:" + strconv.Itoa(addr.Port),
		opts: []Option{
			WithTLS(&tls.Config{RootCAs: pool}),
			WithWebTransport(),
		},
	}
}

// liveRelay returns a relayEnv for tern.fly.dev if TERN_TOKEN is set.
// Skips the test otherwise. Uses WebTransport since the live relay may
// not yet have a raw QUIC port.
func liveRelay(t *testing.T) relayEnv {
	t.Helper()
	token := os.Getenv("TERN_TOKEN")
	if token == "" {
		t.Skip("TERN_TOKEN not set; skipping live test")
	}

	env := relayEnv{
		url: "https://tern.fly.dev:443",
		opts: []Option{
			WithToken(token),
			WithWebTransport(),
		},
	}

	// Probe connectivity — skip if the relay isn't reachable over QUIC/UDP.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	probe, err := Register(ctx, env.url, env.opts...)
	if err != nil {
		t.Skipf("live relay not reachable: %v", err)
	}
	probe.CloseNow()

	return env
}

// connectPair registers a backend and connects a client, returning both.
func connectPair(t *testing.T, env relayEnv) (*Conn, *Conn) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	t.Cleanup(cancel)

	b, err := Register(ctx, env.url, env.opts...)
	if err != nil {
		t.Fatal("register:", err)
	}
	t.Cleanup(func() { b.CloseNow() })

	c, err := Connect(ctx, env.url, b.InstanceID(), env.opts...)
	if err != nil {
		t.Fatal("connect:", err)
	}
	t.Cleanup(func() { c.CloseNow() })

	return b, c
}

// setupEncryption creates matching E2E channels on both sides.
func setupEncryption(t *testing.T, b, c *Conn) {
	t.Helper()
	bKP, err := crypto.GenerateKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	cKP, err := crypto.GenerateKeyPair()
	if err != nil {
		t.Fatal(err)
	}

	bSendKey, err := crypto.DeriveSessionKey(bKP.Private, cKP.Public, []byte("b-to-c"))
	if err != nil {
		t.Fatal(err)
	}
	bRecvKey, err := crypto.DeriveSessionKey(bKP.Private, cKP.Public, []byte("c-to-b"))
	if err != nil {
		t.Fatal(err)
	}
	cSendKey, err := crypto.DeriveSessionKey(cKP.Private, bKP.Public, []byte("c-to-b"))
	if err != nil {
		t.Fatal(err)
	}
	cRecvKey, err := crypto.DeriveSessionKey(cKP.Private, bKP.Public, []byte("b-to-c"))
	if err != nil {
		t.Fatal(err)
	}

	bCh, err := crypto.NewChannel(bSendKey, bRecvKey)
	if err != nil {
		t.Fatal(err)
	}
	cCh, err := crypto.NewChannel(cSendKey, cRecvKey)
	if err != nil {
		t.Fatal(err)
	}

	b.SetChannel(bCh)
	c.SetChannel(cCh)
}

// setupDatagramEncryption creates matching datagram channels on both sides.
func setupDatagramEncryption(t *testing.T, b, c *Conn) {
	t.Helper()
	bKP, err := crypto.GenerateKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	cKP, err := crypto.GenerateKeyPair()
	if err != nil {
		t.Fatal(err)
	}

	bSendKey, err := crypto.DeriveSessionKey(bKP.Private, cKP.Public, []byte("dg-b-to-c"))
	if err != nil {
		t.Fatal(err)
	}
	bRecvKey, err := crypto.DeriveSessionKey(bKP.Private, cKP.Public, []byte("dg-c-to-b"))
	if err != nil {
		t.Fatal(err)
	}
	cSendKey, err := crypto.DeriveSessionKey(cKP.Private, bKP.Public, []byte("dg-c-to-b"))
	if err != nil {
		t.Fatal(err)
	}
	cRecvKey, err := crypto.DeriveSessionKey(cKP.Private, bKP.Public, []byte("dg-b-to-c"))
	if err != nil {
		t.Fatal(err)
	}

	bCh, err := crypto.NewChannel(bSendKey, bRecvKey)
	if err != nil {
		t.Fatal(err)
	}
	cCh, err := crypto.NewChannel(cSendKey, cRecvKey)
	if err != nil {
		t.Fatal(err)
	}

	b.SetDatagramChannel(bCh)
	c.SetDatagramChannel(cCh)
}

// liveRelayEnv returns the token and URL for the live relay, or empty
// strings if TERN_TOKEN is not set.
func liveRelayEnv() (token, url string) {
	token = os.Getenv("TERN_TOKEN")
	if token == "" {
		return "", ""
	}
	return token, "https://tern.fly.dev:443"
}

// localRelayB is localRelay for benchmarks.
func localRelayB(b *testing.B) relayEnv {
	b.Helper()
	return localRelayTB(b)
}

// forEachRelay runs a subtest against local (QUIC), local (WebTransport),
// and live relay environments.
func forEachRelay(t *testing.T, fn func(t *testing.T, env relayEnv)) {
	t.Run("local/quic", func(t *testing.T) { fn(t, localRelay(t)) })
	t.Run("local/webtransport", func(t *testing.T) { fn(t, localRelayWT(t)) })
	t.Run("live", func(t *testing.T) { fn(t, liveRelay(t)) })
}

// --- Tests ---

func TestStreamRoundTrip(t *testing.T) {
	forEachRelay(t, func(t *testing.T, env relayEnv) {
		ctx := context.Background()
		b, c := connectPair(t, env)

		if err := c.Send(ctx, []byte("hello")); err != nil {
			t.Fatal(err)
		}
		data, err := b.Recv(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != "hello" {
			t.Fatalf("got %q, want hello", data)
		}

		if err := b.Send(ctx, []byte("reply")); err != nil {
			t.Fatal(err)
		}
		data, err = c.Recv(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != "reply" {
			t.Fatalf("got %q, want reply", data)
		}
	})
}

func TestEncryptedStreamRoundTrip(t *testing.T) {
	forEachRelay(t, func(t *testing.T, env relayEnv) {
		ctx := context.Background()
		b, c := connectPair(t, env)
		setupEncryption(t, b, c)

		if err := c.Send(ctx, []byte("secret")); err != nil {
			t.Fatal(err)
		}
		data, err := b.Recv(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != "secret" {
			t.Fatalf("got %q, want secret", data)
		}
	})
}

func TestDatagramRoundTrip(t *testing.T) {
	forEachRelay(t, func(t *testing.T, env relayEnv) {
		ctx := context.Background()
		b, c := connectPair(t, env)

		if err := c.SendDatagram([]byte("dgram")); err != nil {
			t.Fatal(err)
		}
		data, err := b.RecvDatagram(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != "dgram" {
			t.Fatalf("got %q, want dgram", data)
		}
	})
}

func TestEncryptedDatagramRoundTrip(t *testing.T) {
	forEachRelay(t, func(t *testing.T, env relayEnv) {
		ctx := context.Background()
		b, c := connectPair(t, env)
		setupDatagramEncryption(t, b, c)

		if err := c.SendDatagram([]byte("encrypted-dgram")); err != nil {
			t.Fatal(err)
		}
		data, err := b.RecvDatagram(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != "encrypted-dgram" {
			t.Fatalf("got %q, want encrypted-dgram", data)
		}
	})
}

// TestOpenStream verifies that OpenStream succeeds on an established Conn.
// The relay currently only bridges the primary stream; additional streams
// are not forwarded to the peer. This test confirms the stream opens without
// error. End-to-end multi-stream relay requires server-side support —
// see the TODO in session.go (bridgeClient).
func TestOpenStream(t *testing.T) {
	forEachRelay(t, func(t *testing.T, env relayEnv) {
		b, _ := connectPair(t, env)

		stream, err := b.OpenStream()
		if err != nil {
			t.Fatalf("OpenStream: %v", err)
		}
		defer stream.Close()
	})
}

func TestMultipleMessages(t *testing.T) {
	forEachRelay(t, func(t *testing.T, env relayEnv) {
		ctx := context.Background()
		b, c := connectPair(t, env)

		const n = 10
		for i := range n {
			if err := c.Send(ctx, []byte("msg-"+strconv.Itoa(i))); err != nil {
				t.Fatalf("send %d: %v", i, err)
			}
		}

		for i := range n {
			expected := "msg-" + strconv.Itoa(i)
			data, err := b.Recv(ctx)
			if err != nil {
				t.Fatalf("recv %d: %v", i, err)
			}
			if string(data) != expected {
				t.Fatalf("msg %d: got %q, want %q", i, data, expected)
			}
		}
	})
}

func TestPersistentInstanceID(t *testing.T) {
	env := localRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	myUUID := "test-device-uuid-12345"

	// Register with a persistent instance ID.
	b, err := Register(ctx, env.url, append(env.opts, WithInstanceID(myUUID))...)
	if err != nil {
		t.Fatal("register:", err)
	}
	if b.InstanceID() != myUUID {
		t.Fatalf("got instance ID %q, want %q", b.InstanceID(), myUUID)
	}

	// Client connects using the persistent ID.
	c, err := Connect(ctx, env.url, myUUID, env.opts...)
	if err != nil {
		t.Fatal("connect:", err)
	}

	// Verify messaging works.
	if err := c.Send(ctx, []byte("persistent")); err != nil {
		t.Fatal(err)
	}
	data, err := b.Recv(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "persistent" {
		t.Fatalf("got %q, want persistent", data)
	}

	c.CloseNow()
	b.CloseNow()
}

func TestStreamingChannel(t *testing.T) {
	env := localRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	b, err := Register(ctx, env.url, env.opts...)
	if err != nil {
		t.Fatal("register:", err)
	}
	defer b.CloseNow()

	c, err := Connect(ctx, env.url, b.InstanceID(), env.opts...)
	if err != nil {
		t.Fatal("connect:", err)
	}
	defer c.CloseNow()

	// Client opens a named channel.
	ch, err := c.OpenChannel("game-state")
	if err != nil {
		t.Fatal("open channel:", err)
	}
	defer ch.Close()

	// Backend accepts the channel.
	bch, err := b.AcceptChannel(ctx)
	if err != nil {
		t.Fatal("accept channel:", err)
	}
	defer bch.Close()

	if bch.Name() != "game-state" {
		t.Fatalf("got channel name %q, want game-state", bch.Name())
	}

	// Send/recv on the channel.
	if err := ch.Send(ctx, []byte("player moved")); err != nil {
		t.Fatal("send:", err)
	}
	data, err := bch.Recv(ctx)
	if err != nil {
		t.Fatal("recv:", err)
	}
	if string(data) != "player moved" {
		t.Fatalf("got %q, want 'player moved'", data)
	}

	// Reverse direction.
	if err := bch.Send(ctx, []byte("state updated")); err != nil {
		t.Fatal("send:", err)
	}
	data, err = ch.Recv(ctx)
	if err != nil {
		t.Fatal("recv:", err)
	}
	if string(data) != "state updated" {
		t.Fatalf("got %q, want 'state updated'", data)
	}
}

func TestMultipleChannels(t *testing.T) {
	env := localRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	b, err := Register(ctx, env.url, env.opts...)
	if err != nil {
		t.Fatal(err)
	}
	defer b.CloseNow()

	c, err := Connect(ctx, env.url, b.InstanceID(), env.opts...)
	if err != nil {
		t.Fatal(err)
	}
	defer c.CloseNow()

	// Open two channels.
	ch1, _ := c.OpenChannel("control")
	ch2, _ := c.OpenChannel("data")
	defer ch1.Close()
	defer ch2.Close()

	bch1, _ := b.AcceptChannel(ctx)
	bch2, _ := b.AcceptChannel(ctx)
	defer bch1.Close()
	defer bch2.Close()

	// Messages on different channels are independent.
	ch1.Send(ctx, []byte("ctrl-msg"))
	ch2.Send(ctx, []byte("data-msg"))

	d1, _ := bch1.Recv(ctx)
	d2, _ := bch2.Recv(ctx)

	// Channel names tell us which is which (order may vary due to concurrency).
	msgs := map[string]string{bch1.Name(): string(d1), bch2.Name(): string(d2)}
	if msgs["control"] != "ctrl-msg" {
		t.Fatalf("control channel got %q", msgs["control"])
	}
	if msgs["data"] != "data-msg" {
		t.Fatalf("data channel got %q", msgs["data"])
	}
}

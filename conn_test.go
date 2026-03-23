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
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/marcelocantos/tern/crypto"
)

// generateTestCert creates a self-signed TLS certificate for testing.
func generateTestCert(t *testing.T) (tls.Certificate, *x509.CertPool) {
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

// localRelay starts a test WebTransport relay and returns a relayEnv.
func localRelay(t *testing.T) relayEnv {
	t.Helper()

	cert, pool := generateTestCert(t)

	srv, err := NewWebTransportServer("127.0.0.1:0", cert, "")
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
		url:  "https://127.0.0.1:" + strconv.Itoa(addr.Port),
		opts: []Option{WithTLS(&tls.Config{RootCAs: pool})},
	}
}

// liveRelay returns a relayEnv for tern.fly.dev if TERN_TOKEN is set.
// Skips the test otherwise.
func liveRelay(t *testing.T) relayEnv {
	t.Helper()
	token := os.Getenv("TERN_TOKEN")
	if token == "" {
		t.Skip("TERN_TOKEN not set; skipping live test")
	}
	// Use TERN_RELAY_URL if set, otherwise default to tern.fly.dev.
	relayURL := os.Getenv("TERN_RELAY_URL")
	if relayURL == "" {
		relayURL = "https://tern.fly.dev:443"
	}

	env := relayEnv{
		url: relayURL,
		opts: []Option{
			WithToken(token),
			// InsecureSkipVerify because the relay generates a self-signed cert.
			WithTLS(&tls.Config{InsecureSkipVerify: true}),
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
	bKP, _ := crypto.GenerateKeyPair()
	cKP, _ := crypto.GenerateKeyPair()

	bSendKey, _ := crypto.DeriveSessionKey(bKP.Private, cKP.Public, []byte("b-to-c"))
	bRecvKey, _ := crypto.DeriveSessionKey(bKP.Private, cKP.Public, []byte("c-to-b"))
	cSendKey, _ := crypto.DeriveSessionKey(cKP.Private, bKP.Public, []byte("c-to-b"))
	cRecvKey, _ := crypto.DeriveSessionKey(cKP.Private, bKP.Public, []byte("b-to-c"))

	bCh, _ := crypto.NewChannel(bSendKey, bRecvKey)
	cCh, _ := crypto.NewChannel(cSendKey, cRecvKey)

	b.SetChannel(bCh)
	c.SetChannel(cCh)
}

// setupDatagramEncryption creates matching datagram channels on both sides.
func setupDatagramEncryption(t *testing.T, b, c *Conn) {
	t.Helper()
	bKP, _ := crypto.GenerateKeyPair()
	cKP, _ := crypto.GenerateKeyPair()

	bSendKey, _ := crypto.DeriveSessionKey(bKP.Private, cKP.Public, []byte("dg-b-to-c"))
	bRecvKey, _ := crypto.DeriveSessionKey(bKP.Private, cKP.Public, []byte("dg-c-to-b"))
	cSendKey, _ := crypto.DeriveSessionKey(cKP.Private, bKP.Public, []byte("dg-c-to-b"))
	cRecvKey, _ := crypto.DeriveSessionKey(cKP.Private, bKP.Public, []byte("dg-b-to-c"))

	bCh, _ := crypto.NewChannel(bSendKey, bRecvKey)
	cCh, _ := crypto.NewChannel(cSendKey, cRecvKey)

	b.SetDatagramChannel(bCh)
	c.SetDatagramChannel(cCh)
}

// forEachRelay runs a subtest against both the local and live relay.
func forEachRelay(t *testing.T, fn func(t *testing.T, env relayEnv)) {
	t.Run("local", func(t *testing.T) { fn(t, localRelay(t)) })
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

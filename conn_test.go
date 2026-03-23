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

// startTestRelay starts a WebTransport relay server on an ephemeral port
// and returns the URL (https://...) and TLS config for connecting.
func startTestRelay(t *testing.T) (string, *tls.Config) {
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

	go func() {
		srv.Serve(conn)
	}()
	t.Cleanup(func() { srv.Close() })

	addr := conn.LocalAddr().(*net.UDPAddr)
	url := "https://127.0.0.1:" + strconv.Itoa(addr.Port)
	tlsConfig := &tls.Config{RootCAs: pool}

	return url, tlsConfig
}

func TestRawModeRoundTrip(t *testing.T) {
	url, tlsConfig := startTestRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	b, err := Register(ctx, url, WithTLS(tlsConfig))
	if err != nil {
		t.Fatal(err)
	}
	defer b.CloseNow()

	c, err := Connect(ctx, url, b.InstanceID(), WithTLS(tlsConfig))
	if err != nil {
		t.Fatal(err)
	}
	defer c.CloseNow()

	// Client -> backend (raw mode).
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
}

func TestEncryptedModeRoundTrip(t *testing.T) {
	url, tlsConfig := startTestRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	b, err := Register(ctx, url, WithTLS(tlsConfig))
	if err != nil {
		t.Fatal(err)
	}
	defer b.CloseNow()

	c, err := Connect(ctx, url, b.InstanceID(), WithTLS(tlsConfig))
	if err != nil {
		t.Fatal(err)
	}
	defer c.CloseNow()

	// Set up matching encrypted channels.
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

	// Client -> backend (encrypted mode — caller sends plaintext).
	if err := c.Send(ctx, []byte("secret")); err != nil {
		t.Fatal(err)
	}

	data, err := b.Recv(ctx)
	if err != nil {
		t.Fatalf("recv: %v", err)
	}
	if string(data) != "secret" {
		t.Fatalf("got %q, want secret", data)
	}

	// Backend -> client.
	if err := b.Send(ctx, []byte("reply")); err != nil {
		t.Fatal(err)
	}

	data, err = c.Recv(ctx)
	if err != nil {
		t.Fatalf("recv: %v", err)
	}
	if string(data) != "reply" {
		t.Fatalf("got %q, want reply", data)
	}
}

func TestDatagramRoundTrip(t *testing.T) {
	url, tlsConfig := startTestRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	b, err := Register(ctx, url, WithTLS(tlsConfig))
	if err != nil {
		t.Fatal(err)
	}
	defer b.CloseNow()

	c, err := Connect(ctx, url, b.InstanceID(), WithTLS(tlsConfig))
	if err != nil {
		t.Fatal(err)
	}
	defer c.CloseNow()

	// Client -> Backend datagram.
	if err := c.SendDatagram([]byte("dgram-c2b")); err != nil {
		t.Fatal("client send datagram:", err)
	}

	data, err := b.RecvDatagram(ctx)
	if err != nil {
		t.Fatal("backend recv datagram:", err)
	}
	if string(data) != "dgram-c2b" {
		t.Fatalf("got %q, want %q", data, "dgram-c2b")
	}

	// Backend -> Client datagram.
	if err := b.SendDatagram([]byte("dgram-b2c")); err != nil {
		t.Fatal("backend send datagram:", err)
	}

	data, err = c.RecvDatagram(ctx)
	if err != nil {
		t.Fatal("client recv datagram:", err)
	}
	if string(data) != "dgram-b2c" {
		t.Fatalf("got %q, want %q", data, "dgram-b2c")
	}
}

func TestEncryptedDatagramRoundTrip(t *testing.T) {
	url, tlsConfig := startTestRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	b, err := Register(ctx, url, WithTLS(tlsConfig))
	if err != nil {
		t.Fatal(err)
	}
	defer b.CloseNow()

	c, err := Connect(ctx, url, b.InstanceID(), WithTLS(tlsConfig))
	if err != nil {
		t.Fatal(err)
	}
	defer c.CloseNow()

	// Set up matching encrypted datagram channels.
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

	// Client -> Backend encrypted datagram.
	if err := c.SendDatagram([]byte("encrypted-dgram")); err != nil {
		t.Fatal("client send datagram:", err)
	}

	data, err := b.RecvDatagram(ctx)
	if err != nil {
		t.Fatal("backend recv datagram:", err)
	}
	if string(data) != "encrypted-dgram" {
		t.Fatalf("got %q, want %q", data, "encrypted-dgram")
	}
}

func TestMultipleMessages(t *testing.T) {
	url, tlsConfig := startTestRelay(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	backend, err := Register(ctx, url, WithTLS(tlsConfig))
	if err != nil {
		t.Fatal("register:", err)
	}
	defer backend.CloseNow()

	client, err := Connect(ctx, url, backend.InstanceID(), WithTLS(tlsConfig))
	if err != nil {
		t.Fatal("connect:", err)
	}
	defer client.CloseNow()

	const n = 10
	for i := range n {
		msg := []byte("msg-" + strconv.Itoa(i))
		if err := client.Send(ctx, msg); err != nil {
			t.Fatalf("send %d: %v", i, err)
		}
	}

	for i := range n {
		expected := "msg-" + strconv.Itoa(i)
		data, err := backend.Recv(ctx)
		if err != nil {
			t.Fatalf("recv %d: %v", i, err)
		}
		if string(data) != expected {
			t.Fatalf("msg %d: got %q, want %q", i, data, expected)
		}
	}
}

// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package main

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

	"github.com/marcelocantos/tern"
)

// testCert creates a self-signed TLS certificate and CA pool for testing.
func testCert(t *testing.T) (tls.Certificate, *x509.CertPool) {
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
// and returns the URL and TLS config for connecting.
func startTestRelay(t *testing.T, token string) (string, *tls.Config) {
	t.Helper()

	cert, pool := testCert(t)

	srv, err := tern.NewWebTransportServer("127.0.0.1:0", &tls.Config{
		Certificates: []tls.Certificate{cert},
	}, token)
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

func TestRegisterAssignsID(t *testing.T) {
	url, tlsConfig := startTestRelay(t, "")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	backend, err := tern.Register(ctx, url, tern.WithTLS(tlsConfig))
	if err != nil {
		t.Fatal("register:", err)
	}
	defer backend.CloseNow()

	if backend.InstanceID() == "" {
		t.Fatal("expected non-empty instance ID")
	}
	t.Log("instance ID:", backend.InstanceID())
}

func TestBidirectionalBridge(t *testing.T) {
	url, tlsConfig := startTestRelay(t, "")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	backend, err := tern.Register(ctx, url, tern.WithTLS(tlsConfig))
	if err != nil {
		t.Fatal("register:", err)
	}
	defer backend.CloseNow()

	client, err := tern.Connect(ctx, url, backend.InstanceID(), tern.WithTLS(tlsConfig))
	if err != nil {
		t.Fatal("connect:", err)
	}
	defer client.CloseNow()

	// client -> backend
	if err := client.Send(ctx, []byte("hello from client")); err != nil {
		t.Fatal("client send:", err)
	}

	data, err := backend.Recv(ctx)
	if err != nil {
		t.Fatal("backend recv:", err)
	}
	if string(data) != "hello from client" {
		t.Fatalf("got %q, want %q", data, "hello from client")
	}

	// backend -> client
	if err := backend.Send(ctx, []byte("hello from backend")); err != nil {
		t.Fatal("backend send:", err)
	}

	data, err = client.Recv(ctx)
	if err != nil {
		t.Fatal("client recv:", err)
	}
	if string(data) != "hello from backend" {
		t.Fatalf("got %q, want %q", data, "hello from backend")
	}
}

func TestDatagramForwarding(t *testing.T) {
	url, tlsConfig := startTestRelay(t, "")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	backend, err := tern.Register(ctx, url, tern.WithTLS(tlsConfig))
	if err != nil {
		t.Fatal("register:", err)
	}
	defer backend.CloseNow()

	client, err := tern.Connect(ctx, url, backend.InstanceID(), tern.WithTLS(tlsConfig))
	if err != nil {
		t.Fatal("connect:", err)
	}
	defer client.CloseNow()

	// client -> backend datagram
	if err := client.SendDatagram([]byte("dgram-c2b")); err != nil {
		t.Fatal("client send datagram:", err)
	}

	data, err := backend.RecvDatagram(ctx)
	if err != nil {
		t.Fatal("backend recv datagram:", err)
	}
	if string(data) != "dgram-c2b" {
		t.Fatalf("got %q, want %q", data, "dgram-c2b")
	}

	// backend -> client datagram
	if err := backend.SendDatagram([]byte("dgram-b2c")); err != nil {
		t.Fatal("backend send datagram:", err)
	}

	data, err = client.RecvDatagram(ctx)
	if err != nil {
		t.Fatal("client recv datagram:", err)
	}
	if string(data) != "dgram-b2c" {
		t.Fatalf("got %q, want %q", data, "dgram-b2c")
	}
}

func TestTokenAuth(t *testing.T) {
	url, tlsConfig := startTestRelay(t, "secret-token")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// No token -> should fail.
	_, err := tern.Register(ctx, url, tern.WithTLS(tlsConfig))
	if err == nil {
		t.Fatal("expected error without token")
	}

	// Wrong token -> should fail.
	_, err = tern.Register(ctx, url, tern.WithTLS(tlsConfig), tern.WithToken("wrong"))
	if err == nil {
		t.Fatal("expected error with wrong token")
	}

	// Correct token -> should succeed.
	backend, err := tern.Register(ctx, url, tern.WithTLS(tlsConfig), tern.WithToken("secret-token"))
	if err != nil {
		t.Fatal("register with token:", err)
	}
	defer backend.CloseNow()

	if backend.InstanceID() == "" {
		t.Fatal("expected non-empty instance ID")
	}
}

func TestSecondClientRejected(t *testing.T) {
	url, tlsConfig := startTestRelay(t, "")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	backend, err := tern.Register(ctx, url, tern.WithTLS(tlsConfig))
	if err != nil {
		t.Fatal("register:", err)
	}
	defer backend.CloseNow()

	// First client connects.
	client1, err := tern.Connect(ctx, url, backend.InstanceID(), tern.WithTLS(tlsConfig))
	if err != nil {
		t.Fatal("connect first:", err)
	}
	defer client1.CloseNow()

	// Second client should be rejected.
	_, err = tern.Connect(ctx, url, backend.InstanceID(), tern.WithTLS(tlsConfig))
	if err == nil {
		t.Fatal("expected error for second client")
	}
}

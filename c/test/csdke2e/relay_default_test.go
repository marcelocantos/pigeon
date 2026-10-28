// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package csdke2e

// This untagged Go peer oracle keeps the relay/package in the default
// post-merge test set. The tagged C peers in this directory require vendored
// ngtcp2/quictls and are separately exercised by c/test/full-duplex.mk.

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
	"testing"
	"time"

	"github.com/marcelocantos/pigeon"
	"github.com/marcelocantos/pigeon/crypto"
)

// TestRelayRecvFirstRequestReply proves a real QUIC relay routes an encrypted
// request on one stream while a peer is already waiting for its reply on
// another. The C peer's analogous scenario runs under the csdke2e tag.
func TestRelayRecvFirstRequestReply(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	template := &x509.Certificate{
		SerialNumber: serial, IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1)},
		NotBefore: now.Add(-time.Hour), NotAfter: now.Add(time.Hour),
		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	tlsServer := &tls.Config{Certificates: []tls.Certificate{{Certificate: [][]byte{certDER}, PrivateKey: key}}}
	tlsClient := &tls.Config{InsecureSkipVerify: true} // ephemeral, self-signed test relay
	wt, err := pigeon.NewWebTransportServer("127.0.0.1:0", tlsServer, pigeon.Auth{})
	if err != nil {
		t.Fatal(err)
	}
	defer wt.Close()
	udp, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	if err != nil {
		t.Fatal(err)
	}
	port := udp.LocalAddr().(*net.UDPAddr).Port
	relay := pigeon.NewQUICServer(fmt.Sprintf("127.0.0.1:%d", port), tlsServer, pigeon.Auth{}, wt.Hub())
	go func() { _ = relay.ServeWithTLS(udp, tlsServer) }()
	defer relay.Close()
	relayURL := fmt.Sprintf("https://127.0.0.1:%d", port)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	tmp := t.TempDir()
	backend, err := crypto.NewFileIdentity(tmp + "/backend.json")
	if err != nil {
		t.Fatal(err)
	}
	client, err := crypto.NewFileIdentity(tmp + "/client.json")
	if err != nil {
		t.Fatal(err)
	}
	backendKP, err := crypto.GenerateKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	clientKP, err := crypto.GenerateKeyPair()
	if err != nil {
		t.Fatal(err)
	}
	clientPub, err := ecdh.X25519().NewPublicKey(clientKP.Public.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	backendPub, err := ecdh.X25519().NewPublicKey(backendKP.Public.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	backendRecord := crypto.NewPairingRecord(client.InstanceID(), relayURL, backendKP, clientPub)
	clientRecord := crypto.NewPairingRecord(backend.InstanceID(), relayURL, clientKP, backendPub)
	listener, _, err := pigeon.Register(ctx, &pigeon.RegisterArgs{
		Identity: backend, Relay: relayURL, TLS: tlsClient,
		Pairing: func(id string) (*crypto.PairingRecord, error) {
			if id != client.InstanceID() {
				return nil, fmt.Errorf("unexpected client %q", id)
			}
			return backendRecord, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	done := make(chan error, 1)
	go func() {
		session, err := listener.Accept(ctx)
		if err != nil {
			done <- fmt.Errorf("accept: %w", err)
			return
		}
		defer session.Close()
		request, err := session.AcceptStream(ctx, "request")
		if err != nil {
			done <- fmt.Errorf("request stream: %w", err)
			return
		}
		message, err := request.Recv(ctx)
		if err != nil {
			done <- fmt.Errorf("request recv: %w", err)
			return
		}
		if string(message) != "ping" {
			done <- fmt.Errorf("request = %q", message)
			return
		}
		reply, err := session.OpenStream(ctx, "reply")
		if err != nil {
			done <- fmt.Errorf("reply stream: %w", err)
			return
		}
		if err := reply.Send([]byte("pong")); err != nil {
			done <- fmt.Errorf("reply send: %w", err)
			return
		}
		ack, err := request.Recv(ctx)
		if err != nil {
			done <- fmt.Errorf("ack recv: %w", err)
			return
		}
		if string(ack) != "ack" {
			done <- fmt.Errorf("ack = %q", ack)
			return
		}
		done <- nil
	}()
	session, err := pigeon.Connect(ctx, &pigeon.ConnectArgs{
		InstanceID: backend.InstanceID(), Record: clientRecord, Identity: client,
		Relay: relayURL, TLS: tlsClient,
	})
	if err != nil {
		t.Fatal(err)
	}
	defer session.Close()
	// Queue the receive before sending the request: the relay and session
	// must not serialize the waiting read ahead of the independent send.
	replyDone := make(chan error, 1)
	receiveStarted := make(chan struct{})
	go func() {
		close(receiveStarted)
		reply, err := session.AcceptStream(ctx, "reply")
		if err != nil {
			replyDone <- err
			return
		}
		got, err := reply.Recv(ctx)
		if err == nil && string(got) != "pong" {
			err = fmt.Errorf("reply = %q", got)
		}
		replyDone <- err
	}()
	<-receiveStarted
	request, err := session.OpenStream(ctx, "request")
	if err != nil {
		t.Fatal(err)
	}
	if err := request.Send([]byte("ping")); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-replyDone:
		if err != nil {
			t.Fatal(err)
		}
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
	if err := request.Send([]byte("ack")); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-ctx.Done():
		t.Fatal(ctx.Err())
	}
}

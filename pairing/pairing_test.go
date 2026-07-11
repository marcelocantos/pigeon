// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package pairing_test

import (
	"context"
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
	"github.com/marcelocantos/pigeon/pairing"
)

// TestCeremonyEndToEnd drives the full pairing ceremony through the
// generated state machine on both sides, in-process, with auto-confirm.
// Asserts:
//   - both Codes match (no MitM in this honest run);
//   - both PairingRecords reference the right peer InstanceIDs;
//   - both records' DeriveChannel produces a working AEAD pair (i.e.
//     the records carry compatible key material).
//
// This is the in-Go equivalent of the cross-language test that used
// to live in Tests/PigeonRelayE2ETests; it's the regression check that
// the YAML-generated PairingCeremonyProtocol*Machine sequences the
// ceremony correctly and that runAcceptor / runInitiator drive those
// machines through their full state graph without spec-rejected events.
//
// 🎯T57: was flaky with "cwire: pigeon_pair_initiator failed" when the
// acceptor Confirm returned and Close tore down the backend Session
// while the initiator was still draining the final confirm through the
// async relay bridge. Fixed in production by acceptor post-success
// drain + initiator stream FIN before deliverResult. T54's ctx-threading
// of ConfirmFn did not create this race (it predates T54 on master) but
// Close-cancel can shrink the drain window; the drain/FIN ordering is
// what makes success deterministic.
func TestCeremonyEndToEnd(t *testing.T) {
	t.Parallel()

	relayURL := startTestRelay(t)
	tmp := t.TempDir()

	bid, err := crypto.NewFileIdentity(tmp + "/acceptor-id.json")
	if err != nil {
		t.Fatalf("acceptor identity: %v", err)
	}
	cid, err := crypto.NewFileIdentity(tmp + "/initiator-id.json")
	if err != nil {
		t.Fatalf("initiator identity: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pairer, err := pairing.Register(ctx, &pairing.Args{Relay: relayURL, Identity: bid})
	if err != nil {
		t.Fatalf("pairing.Register: %v", err)
	}
	defer pairer.Close()

	type accResult struct {
		rec  *crypto.PairingRecord
		code string
		err  error
	}
	accCh := make(chan accResult, 1)
	tokenCh := make(chan string, 1)

	go func() {
		cer, err := pairer.Accept(ctx)
		if err != nil {
			accCh <- accResult{err: fmt.Errorf("pairer.Accept: %w", err)}
			return
		}
		defer cer.Close()
		tokenCh <- cer.Token

		code, err := cer.Code(ctx)
		if err != nil {
			accCh <- accResult{err: fmt.Errorf("acceptor code: %w", err)}
			return
		}
		rec, err := cer.Confirm(ctx)
		accCh <- accResult{rec: rec, code: code, err: err}
	}()

	var token string
	select {
	case token = <-tokenCh:
	case <-time.After(5 * time.Second):
		t.Fatalf("timed out waiting for acceptor token")
	}

	icer, err := pairing.Initiate(ctx, &pairing.InitiateArgs{Rendezvous: token, Identity: cid})
	if err != nil {
		t.Fatalf("pairing.Initiate: %v", err)
	}
	defer icer.Close()

	icode, err := icer.Code(ctx)
	if err != nil {
		t.Fatalf("initiator code: %v", err)
	}
	clientRec, err := icer.Confirm(ctx)
	if err != nil {
		t.Fatalf("initiator confirm: %v", err)
	}

	res := <-accCh
	if res.err != nil {
		t.Fatalf("acceptor: %v", res.err)
	}

	if res.code != icode {
		t.Fatalf("confirmation code mismatch: acceptor=%q initiator=%q", res.code, icode)
	}
	if res.code == "" || len(res.code) != 6 {
		t.Fatalf("unexpected code shape %q (want 6 digits)", res.code)
	}

	if res.rec.PeerInstanceID != cid.InstanceID() {
		t.Errorf("acceptor's record peer = %q, want %q", res.rec.PeerInstanceID, cid.InstanceID())
	}
	if clientRec.PeerInstanceID != bid.InstanceID() {
		t.Errorf("initiator's record peer = %q, want %q", clientRec.PeerInstanceID, bid.InstanceID())
	}

	// Verify the records produce compatible AEAD channels — proving the
	// ceremony exchanged matching ECDH key material under the spec.
	bch, err := res.rec.DeriveChannel([]byte("a->b"), []byte("b->a"), nil)
	if err != nil {
		t.Fatalf("acceptor DeriveChannel: %v", err)
	}
	cch, err := clientRec.DeriveChannel([]byte("b->a"), []byte("a->b"), nil)
	if err != nil {
		t.Fatalf("initiator DeriveChannel: %v", err)
	}
	pt := []byte("hello over a derived channel")
	ct := bch.Encrypt(pt)
	got, err := cch.Decrypt(ct)
	if err != nil {
		t.Fatalf("cross-side decrypt: %v", err)
	}
	if string(got) != string(pt) {
		t.Fatalf("cross-side payload mismatch: got %q want %q", got, pt)
	}
}

func startTestRelay(t *testing.T) string {
	t.Helper()
	cert := selfSignedCert(t)
	tlsCfg := &tls.Config{Certificates: []tls.Certificate{cert}}

	wtSrv, err := pigeon.NewWebTransportServer("127.0.0.1:0", tlsCfg, pigeon.Auth{})
	if err != nil {
		t.Fatalf("wt server: %v", err)
	}
	udpAddr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		t.Fatalf("listen udp: %v", err)
	}
	port := udpConn.LocalAddr().(*net.UDPAddr).Port
	qSrv := pigeon.NewQUICServer(fmt.Sprintf("127.0.0.1:%d", port), tlsCfg, pigeon.Auth{}, wtSrv.Hub())
	go qSrv.ServeWithTLS(udpConn, tlsCfg)
	t.Cleanup(func() {
		_ = qSrv.Close()
		_ = wtSrv.Close()
	})
	time.Sleep(150 * time.Millisecond)
	return fmt.Sprintf("https://127.0.0.1:%d", port)
}

func selfSignedCert(t *testing.T) tls.Certificate {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("gen key: %v", err)
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
		t.Fatalf("cert: %v", err)
	}
	return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key}
}

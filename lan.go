// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package pigeon

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math/big"
	"net"
	"sync"
	"time"

	"github.com/marcelocantos/pigeon/qr"
	"github.com/quic-go/quic-go"
)

// lanControlStreamName is the reserved sub-stream the backend opens to
// advertise a LAN offer after activation (🎯T48). The NUL prefix keeps
// it out of the application name space (apps use plain names like "chat").
const lanControlStreamName = "\x00pigeon.lan"

// LANServer is a local QUIC listener that accepts direct connections
// from clients on the same LAN. It is the local counterpart of the
// relay server — same protocol, no relay in between.
//
// The backend creates a LANServer at startup and passes it via
// RegisterArgs.LAN. After each client activates, the Session advertises
// the LAN address over an encrypted control stream. A client that set
// ConnectArgs.PreferLAN dials the backend directly, verifies via
// challenge/response, and the Session transparently swaps its transport
// to the direct path.
//
// Usage:
//
//	lan, _ := pigeon.NewLANServer("", nil) // random port, self-signed cert
//	defer lan.Close()
//
//	listener, _, _ := pigeon.Register(ctx, &pigeon.RegisterArgs{
//	    Identity: id, Pairing: resolve, Relay: relayURL, LAN: lan,
//	})
type LANServer struct {
	listener *quic.Listener
	addr     string // "ip:port" on the LAN
	certHash []byte // SHA-256 of DER cert for browser serverCertificateHashes
	mu       sync.Mutex
	conns    map[string]*pendingLAN // pending key → challenge + callback
}

// pendingLAN tracks a client that should connect via LAN.
type pendingLAN struct {
	challenge []byte
	onVerify  func(stream io.ReadWriteCloser, conn *quic.Conn)
}

// NewLANServer creates a LAN QUIC listener. The addr parameter
// specifies the listen address (e.g., ":0" for a random port,
// "127.0.0.1:44333" for a fixed address). If addr is empty, ":0"
// is used. If tlsConfig is nil, a self-signed certificate is generated.
func NewLANServer(addr string, tlsConfig *tls.Config) (*LANServer, error) {
	if addr == "" {
		addr = ":0"
	}

	if tlsConfig == nil {
		cert, err := generateSelfSigned()
		if err != nil {
			return nil, fmt.Errorf("generate LAN cert: %w", err)
		}
		tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
			NextProtos:   []string{"pigeon-lan"},
		}
	} else {
		tlsConfig = tlsConfig.Clone()
		tlsConfig.NextProtos = []string{"pigeon-lan"}
	}

	listener, err := quic.ListenAddr(addr, tlsConfig, &quic.Config{
		EnableDatagrams:      true,
		MaxIdleTimeout:       60 * time.Second,
		KeepAlivePeriod:      10 * time.Second,
		HandshakeIdleTimeout: 30 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("LAN listen: %w", err)
	}

	// Determine the advertised address. If the listen address doesn't
	// specify a host (e.g., ":0" or ":44333"), use the LAN IP.
	listenAddr := listener.Addr().(*net.UDPAddr)
	host := listenAddr.IP.String()
	if listenAddr.IP.IsUnspecified() {
		host = qr.LanIP()
	}
	advertised := fmt.Sprintf("%s:%d", host, listenAddr.Port)

	// Compute cert hash for browser serverCertificateHashes.
	var certHash []byte
	if len(tlsConfig.Certificates) > 0 && len(tlsConfig.Certificates[0].Certificate) > 0 {
		h := sha256.Sum256(tlsConfig.Certificates[0].Certificate[0])
		certHash = h[:]
	}

	s := &LANServer{
		listener: listener,
		addr:     advertised,
		certHash: certHash,
		conns:    make(map[string]*pendingLAN),
	}

	go s.acceptLoop()

	slog.Info("LAN server started", "addr", advertised)
	return s, nil
}

// Addr returns the LAN address (ip:port) that clients should dial.
func (s *LANServer) Addr() string { return s.addr }

// CertHash returns the SHA-256 hash of the LAN server's DER-encoded
// certificate. Browser clients pass this in serverCertificateHashes
// when connecting via WebTransport.
func (s *LANServer) CertHash() []byte { return s.certHash }

// Close stops the LAN server.
func (s *LANServer) Close() error {
	return s.listener.Close()
}

// RegisterPending records a challenge for a pending LAN upgrade keyed by
// id (typically the client's stable InstanceID). When a dialer presents a
// matching lanVerify, onVerify runs with the handshake stream and the
// established QUIC connection.
func (s *LANServer) RegisterPending(id string, challenge []byte, onVerify func(stream io.ReadWriteCloser, conn *quic.Conn)) {
	if s == nil || id == "" {
		return
	}
	s.mu.Lock()
	s.conns[id] = &pendingLAN{
		challenge: append([]byte(nil), challenge...),
		onVerify:  onVerify,
	}
	s.mu.Unlock()
}

// UnregisterPending drops a pending entry (e.g. on offer timeout).
func (s *LANServer) UnregisterPending(id string) {
	if s == nil {
		return
	}
	s.mu.Lock()
	delete(s.conns, id)
	s.mu.Unlock()
}

// acceptLoop accepts incoming LAN connections and verifies them.
func (s *LANServer) acceptLoop() {
	for {
		conn, err := s.listener.Accept(context.Background())
		if err != nil {
			return
		}
		go s.handleConn(conn)
	}
}

func (s *LANServer) handleConn(conn *quic.Conn) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stream, err := conn.AcceptStream(ctx)
	if err != nil {
		conn.CloseWithError(1, "no stream")
		return
	}

	data, err := readMessage(stream)
	if err != nil {
		conn.CloseWithError(1, "read verify failed")
		return
	}

	var verify lanVerify
	if err := json.Unmarshal(data, &verify); err != nil {
		conn.CloseWithError(1, "bad verify")
		return
	}

	// Look up the pending connection by client identity key.
	s.mu.Lock()
	pending, ok := s.conns[verify.InstanceID]
	if ok {
		delete(s.conns, verify.InstanceID)
	}
	s.mu.Unlock()

	if !ok {
		conn.CloseWithError(1, "unknown instance")
		return
	}

	// Verify the challenge.
	if !challengeEqual(pending.challenge, verify.Challenge) {
		conn.CloseWithError(1, "bad challenge")
		return
	}

	// Send confirmation.
	if err := writeMessage(stream, []byte("ok")); err != nil {
		conn.CloseWithError(1, "write confirm failed")
		return
	}

	slog.Info("LAN connection verified", "peer", verify.InstanceID)

	if pending.onVerify != nil {
		pending.onVerify(stream, conn)
	}
	slog.Info("upgraded to LAN", "peer", verify.InstanceID)
}

// --- lanOffer / lanVerify wire types ---

// lanOffer is sent via the encrypted relay channel to advertise the
// LAN server address.
type lanOffer struct {
	Addr      string `json:"addr"`
	Challenge []byte `json:"challenge"`
	CertHash  []byte `json:"cert_hash,omitempty"` // SHA-256 of DER cert for serverCertificateHashes
}

// lanVerify is sent on the direct LAN connection to prove identity.
// InstanceID is the client's stable identifier (the key used with
// RegisterPending on the backend).
type lanVerify struct {
	Challenge  []byte `json:"challenge"`
	InstanceID string `json:"instance_id"`
}

// --- Helpers ---

func challengeEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	return subtle.ConstantTimeCompare(a, b) == 1
}

func generateSelfSigned() (tls.Certificate, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, err
	}

	notBefore := time.Now().Add(-time.Hour)
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		NotBefore:    notBefore,
		// 14 days from NotBefore — Chromium's serverCertificateHashes
		// requires validity ≤ 14 days measured from NotBefore.
		NotAfter:    notBefore.Add(14 * 24 * time.Hour),
		KeyUsage:    x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		DNSNames:    []string{"localhost"},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		return tls.Certificate{}, err
	}

	return tls.Certificate{
		Certificate: [][]byte{certDER},
		PrivateKey:  key,
	}, nil
}

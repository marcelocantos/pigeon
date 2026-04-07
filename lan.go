// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package pigeon

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
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

// LANServer is a local QUIC listener that accepts direct connections
// from clients on the same LAN. It is the local counterpart of the
// relay server — same protocol, no relay in between.
//
// The backend creates a LANServer at startup. When a client connects
// via the relay, the backend's Conn advertises the LAN address. The
// client attempts a direct connection; if successful, the Conn
// transparently switches to the LAN path.
//
// Usage:
//
//	lan, _ := pigeon.NewLANServer(tlsConfig)  // random port
//	defer lan.Close()
//
//	// Register with the relay, passing the LAN server.
//	b, _ := pigeon.Register(ctx, relayURL, pigeon.WithLANServer(lan))
//	// The LAN address is automatically advertised to connecting clients.
type LANServer struct {
	listener *quic.Listener
	addr     string // "ip:port" on the LAN
	mu       sync.Mutex
	conns    map[string]*pendingLAN // instance ID → pending connection
}

// pendingLAN tracks a client that should connect via LAN.
type pendingLAN struct {
	challenge []byte
	conn      *Conn // the relay Conn to upgrade (nil when executor-driven)
	onVerify  func(stream io.ReadWriteCloser, conn *quic.Conn) // executor callback
}

// NewLANServer creates a LAN QUIC listener. The addr parameter
// specifies the listen address (e.g., ":0" for a random port,
// "localhost:44333" for a fixed address). If addr is empty, ":0"
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
		EnableDatagrams: true,
		MaxIdleTimeout:  60 * time.Second,
		KeepAlivePeriod: 10 * time.Second,
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

	s := &LANServer{
		listener: listener,
		addr:     advertised,
		conns:    make(map[string]*pendingLAN),
	}

	go s.acceptLoop()

	slog.Info("LAN server started", "addr", advertised)
	return s, nil
}

// Addr returns the LAN address (ip:port) that clients should dial.
func (s *LANServer) Addr() string { return s.addr }

// Close stops the LAN server.
func (s *LANServer) Close() error {
	return s.listener.Close()
}

// registerConn records a Conn for LAN upgrade. When a client connects
// directly and presents the correct challenge, the Conn's transport
// is swapped. Returns the lanOffer to send to the client via relay.
func (s *LANServer) registerConn(c *Conn) (lanOffer, error) {
	challenge := make([]byte, 32)
	if _, err := rand.Read(challenge); err != nil {
		return lanOffer{}, err
	}

	s.mu.Lock()
	s.conns[c.instanceID] = &pendingLAN{
		challenge: challenge,
		conn:      c,
	}
	s.mu.Unlock()

	return lanOffer{
		Addr:      s.addr,
		Challenge: challenge,
	}, nil
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

	// Look up the pending connection by instance ID.
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
}

// lanVerify is sent on the direct LAN connection to prove identity.
type lanVerify struct {
	Challenge  []byte `json:"challenge"`
	InstanceID string `json:"instance_id"`
}

// --- Conn integration ---
// All LAN lifecycle (advertiseLAN, handleLANOffer, setDirectPath,
// sendControl) is now handled by the executor. See executor.go.

// --- Helpers ---

func challengeEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func generateSelfSigned() (tls.Certificate, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, err
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1)},
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

// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package pigeon

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"crypto/tls"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"github.com/quic-go/webtransport-go"
)

// maxMessageSize aliases the generated wire constant for relay message frames.
const maxMessageSize = MaxMessageSize

// wtSession wraps a WebTransport session to implement relaySession.
type wtSession struct {
	session *webtransport.Session
	stream  *webtransport.Stream
	writeMu sync.Mutex // serialises writes to the stream
}

type wtStreamWrapper struct {
	stream  *webtransport.Stream
	writeMu sync.Mutex
}

func (w *wtStreamWrapper) ReadMessage() ([]byte, error)  { return readMessage(w.stream) }
func (w *wtStreamWrapper) WriteMessage(data []byte) error { w.writeMu.Lock(); defer w.writeMu.Unlock(); return writeMessage(w.stream, data) }
func (w *wtStreamWrapper) Close() error {
	// Set a past read deadline to unblock any pending ReadMessage call
	// (including handleSessionGoneError waits inside the WT library).
	// This avoids hanging indefinitely when the session close signal
	// is delayed through the QUIC stack.
	_ = w.stream.SetReadDeadline(time.Unix(0, 1))
	return w.stream.Close()
}

func (s *wtSession) ReadMessage() ([]byte, error) {
	return readMessage(s.stream)
}

func (s *wtSession) WriteMessage(data []byte) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	return writeMessage(s.stream, data)
}

func (s *wtSession) SendDatagram(data []byte) error {
	return s.session.SendDatagram(data)
}

func (s *wtSession) ReceiveDatagram(ctx context.Context) ([]byte, error) {
	return s.session.ReceiveDatagram(ctx)
}

func (s *wtSession) AcceptStream(ctx context.Context) (readWriteCloserPair, error) {
	stream, err := s.session.AcceptStream(ctx)
	if err != nil {
		return nil, err
	}
	return &wtStreamWrapper{stream: stream}, nil
}

func (s *wtSession) OpenStream() (readWriteCloserPair, error) {
	stream, err := s.session.OpenStream()
	if err != nil {
		return nil, err
	}
	return &wtStreamWrapper{stream: stream}, nil
}

func (s *wtSession) Context() context.Context {
	return s.session.Context()
}

func (s *wtSession) Close() error {
	return s.session.CloseWithError(0, "")
}

// generateID generates a random 128-bit instance ID as a hex string.
func generateID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic("crypto/rand failed: " + err.Error())
	}
	return hex.EncodeToString(b)
}

// writeMessage writes a length-prefixed binary message to a stream.
// Format: [4-byte big-endian length][payload]
func writeMessage(stream io.Writer, data []byte) error {
	if len(data) > maxMessageSize {
		return fmt.Errorf("message too large: %d > %d", len(data), maxMessageSize)
	}
	var hdr [4]byte
	binary.BigEndian.PutUint32(hdr[:], uint32(len(data)))
	if _, err := stream.Write(hdr[:]); err != nil {
		return err
	}
	_, err := stream.Write(data)
	return err
}

// readMessage reads a length-prefixed binary message from a stream.
func readMessage(stream io.Reader) ([]byte, error) {
	var hdr [4]byte
	if _, err := io.ReadFull(stream, hdr[:]); err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint32(hdr[:])
	if length > maxMessageSize {
		return nil, fmt.Errorf("message too large: %d > %d", length, maxMessageSize)
	}
	buf := make([]byte, length)
	if _, err := io.ReadFull(stream, buf); err != nil {
		return nil, err
	}
	return buf, nil
}

// WebTransportServer provides a WebTransport relay. Backends register via
// /register; clients connect via /ws/{id}. Traffic is bridged
// bidirectionally, including datagrams.
type WebTransportServer struct {
	wtServer *webtransport.Server
	hub      *hub
	token    string // bearer token for /register auth; empty = open
	addr     string
	conn     net.PacketConn
}

// NewWebTransportServer creates a WebTransport relay server listening on addr.
// The provided TLS config is used for the QUIC/HTTP3 connection (it may use
// static certificates or a dynamic GetCertificate callback such as certmagic).
// If token is non-empty, /register requires a matching Bearer token.
func NewWebTransportServer(addr string, tlsConfig *tls.Config, token string) (*WebTransportServer, error) {
	return NewWebTransportServerWithHub(addr, tlsConfig, token, newHub())
}

// NewWebTransportServerWithHub creates a WebTransport relay server that
// shares the provided hub with other server types (e.g. raw QUIC).
func NewWebTransportServerWithHub(addr string, tlsConfig *tls.Config, token string, h *hub) (*WebTransportServer, error) {
	mux := http.NewServeMux()
	s := &WebTransportServer{
		hub:   h,
		token: token,
		addr:  addr,
	}

	// Clone to avoid mutating the caller's config.
	serverTLS := tlsConfig.Clone()
	serverTLS.NextProtos = []string{http3.NextProtoH3}

	wtServer := &webtransport.Server{
		H3: &http3.Server{
			Addr:            addr,
			Handler:         mux,
			TLSConfig:       serverTLS,
			EnableDatagrams: true,
			QUICConfig: &quic.Config{
				EnableDatagrams: true,
				MaxIdleTimeout:  60 * time.Second,
				KeepAlivePeriod: 10 * time.Second,
			},
		},
		CheckOrigin: func(r *http.Request) bool { return true },
	}
	webtransport.ConfigureHTTP3Server(wtServer.H3)
	s.wtServer = wtServer

	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	mux.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		s.handleRegister(w, r)
	})

	mux.HandleFunc("/ws/", func(w http.ResponseWriter, r *http.Request) {
		s.handleClient(w, r)
	})

	return s, nil
}

func (s *WebTransportServer) handleRegister(w http.ResponseWriter, r *http.Request) {
	if s.token != "" {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			if qtoken := r.URL.Query().Get("token"); qtoken != "" {
				auth = "Bearer " + qtoken
			}
		}
		if subtle.ConstantTimeCompare([]byte(auth), []byte("Bearer "+s.token)) != 1 {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
	}

	session, err := s.wtServer.Upgrade(w, r)
	if err != nil {
		slog.Error("register: upgrade failed", "err", err)
		return
	}
	// Accept the bidirectional stream opened by the backend client.
	stream, err := session.AcceptStream(session.Context())
	if err != nil {
		slog.Error("register: accept stream failed", "err", err)
		session.CloseWithError(0, "failed to accept stream")
		return
	}

	// Read the handshake message. May contain a requested instance ID:
	// "register" or "register::INSTANCE_ID"
	handshake, err := readMessage(stream)
	if err != nil {
		slog.Error("register: read handshake failed", "err", err)
		session.CloseWithError(0, "failed to read handshake")
		return
	}

	var id string
	msg := string(handshake)
	if strings.HasPrefix(msg, "register::") {
		id = msg[len("register::"):]
	}
	if id == "" {
		id = generateID()
	}

	// Send the instance ID to the backend.
	if err := writeMessage(stream, []byte(id)); err != nil {
		slog.Error("register: write ID failed", "err", err)
		session.CloseWithError(0, "failed to write ID")
		return
	}

	sess := &wtSession{session: session, stream: stream}
	inst := &instance{id: id, session: sess}
	s.hub.register(inst)
	defer s.hub.unregister(id)

	slog.Info("instance registered", "id", id, "transport", "webtransport")

	// Keep alive until backend disconnects.
	<-session.Context().Done()
	slog.Info("instance disconnected", "id", id)
}

func (s *WebTransportServer) handleClient(w http.ResponseWriter, r *http.Request) {
	// Extract instance ID from path: /ws/{id}
	instanceID := r.URL.Path[len("/ws/"):]
	if len(instanceID) == 0 || len(instanceID) > 64 {
		http.Error(w, `{"error":"invalid instance ID"}`, http.StatusBadRequest)
		return
	}

	inst := s.hub.get(instanceID)
	if inst == nil {
		http.Error(w, `{"error":"instance not found"}`, http.StatusNotFound)
		return
	}

	// Enforce single client per instance.
	inst.mu.Lock()
	if inst.occupied {
		inst.mu.Unlock()
		http.Error(w, `{"error":"instance already has a connected client"}`, http.StatusConflict)
		return
	}
	inst.occupied = true
	inst.mu.Unlock()
	defer func() {
		inst.mu.Lock()
		inst.occupied = false
		inst.mu.Unlock()
	}()

	session, err := s.wtServer.Upgrade(w, r)
	if err != nil {
		slog.Error("client: upgrade failed", "err", err)
		return
	}
	defer session.CloseWithError(0, "")

	// Accept the bidirectional stream opened by the client.
	clientStream, err := session.AcceptStream(session.Context())
	if err != nil {
		slog.Error("client: accept stream failed", "err", err)
		return
	}

	// Read the handshake message.
	if _, err := readMessage(clientStream); err != nil {
		slog.Error("client: read handshake failed", "err", err)
		return
	}

	// Send acknowledgment so the client knows the relay has processed the
	// handshake. This ensures any additional streams opened after Connect()
	// returns are enqueued after the primary stream in the relay's accept queue,
	// avoiding stream ordering races in the WebTransport session manager.
	if err := writeMessage(clientStream, []byte("ok")); err != nil {
		slog.Error("client: write ack failed", "err", err)
		return
	}

	slog.Info("client connected", "instance", instanceID, "transport", "webtransport")

	clientSess := &wtSession{session: session, stream: clientStream}
	bridgeClient(inst, clientSess)
}

// Hub returns the shared hub for use by other server types.
func (s *WebTransportServer) Hub() *hub {
	return s.hub
}

// Serve starts the WebTransport server using the provided PacketConn.
func (s *WebTransportServer) Serve(conn net.PacketConn) error {
	s.conn = conn
	return s.wtServer.Serve(conn)
}

// ListenAndServe starts the WebTransport server.
func (s *WebTransportServer) ListenAndServe() error {
	addr := s.addr
	if addr == "" {
		addr = ":443"
	}
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return err
	}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return err
	}
	return s.Serve(conn)
}

// Close shuts down the server.
func (s *WebTransportServer) Close() error {
	return s.wtServer.Close()
}

// Addr returns the local address the server is listening on, or nil if
// Serve/ListenAndServe has not been called.
func (s *WebTransportServer) Addr() net.Addr {
	if s.conn != nil {
		return s.conn.LocalAddr()
	}
	return nil
}

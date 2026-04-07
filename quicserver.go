// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package pigeon

import (
	"context"
	"crypto/subtle"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/quic-go/quic-go"
)

// pigeonALPN is the ALPN protocol identifier for raw QUIC pigeon connections.
const pigeonALPN = "pigeon"

// quicSession wraps a raw QUIC connection to implement relaySession.
type quicSession struct {
	conn    *quic.Conn
	stream  *quic.Stream
	writeMu sync.Mutex // serialises writes to the stream
}

// quicStreamWrapper wraps a quic.Stream as a readWriteCloserPair.
type quicStreamWrapper struct {
	stream  *quic.Stream
	writeMu sync.Mutex
}

func (w *quicStreamWrapper) ReadMessage() ([]byte, error)  { return readMessage(w.stream) }
func (w *quicStreamWrapper) WriteMessage(data []byte) error { w.writeMu.Lock(); defer w.writeMu.Unlock(); return writeMessage(w.stream, data) }
func (w *quicStreamWrapper) Close() error                   { return w.stream.Close() }

func (s *quicSession) ReadMessage() ([]byte, error) {
	return readMessage(s.stream)
}

func (s *quicSession) WriteMessage(data []byte) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	return writeMessage(s.stream, data)
}

func (s *quicSession) SendDatagram(data []byte) error {
	return s.conn.SendDatagram(data)
}

func (s *quicSession) ReceiveDatagram(ctx context.Context) ([]byte, error) {
	return s.conn.ReceiveDatagram(ctx)
}

func (s *quicSession) AcceptStream(ctx context.Context) (readWriteCloserPair, error) {
	stream, err := s.conn.AcceptStream(ctx)
	if err != nil {
		return nil, err
	}
	return &quicStreamWrapper{stream: stream}, nil
}

func (s *quicSession) OpenStream() (readWriteCloserPair, error) {
	stream, err := s.conn.OpenStream()
	if err != nil {
		return nil, err
	}
	return &quicStreamWrapper{stream: stream}, nil
}

func (s *quicSession) Context() context.Context {
	return s.conn.Context()
}

func (s *quicSession) Close() error {
	return s.conn.CloseWithError(0, "")
}

// QUICServer provides a raw QUIC relay for native clients. It shares a
// hub with the WebTransport server so that a raw QUIC backend can talk
// to a WebTransport browser client and vice versa.
type QUICServer struct {
	hub      *hub
	token    string
	addr     string
	listener *quic.Listener
	conn     net.PacketConn
	tlsConfig *tls.Config
}

// NewQUICServer creates a raw QUIC relay server. The hub is shared with
// a WebTransport server so instances are visible to both protocols.
func NewQUICServer(addr string, tlsConfig *tls.Config, token string, h *hub) *QUICServer {
	return &QUICServer{
		hub:   h,
		token: token,
		addr:  addr,
	}
}

// ServeWithTLS starts the QUIC server on the provided PacketConn with
// the given TLS config.
func (s *QUICServer) ServeWithTLS(conn net.PacketConn, tlsConfig *tls.Config) error {
	s.conn = conn

	serverTLS := tlsConfig.Clone()
	serverTLS.NextProtos = []string{pigeonALPN}

	tr := &quic.Transport{Conn: conn}
	ln, err := tr.Listen(serverTLS, &quic.Config{
		EnableDatagrams: true,
		MaxIdleTimeout:  60 * time.Second,
		KeepAlivePeriod: 10 * time.Second,
	})
	if err != nil {
		return fmt.Errorf("quic listen: %w", err)
	}
	s.listener = ln

	return s.acceptLoop()
}

// ListenAndServe starts the QUIC server on the configured address.
func (s *QUICServer) ListenAndServe(tlsConfig *tls.Config) error {
	addr := s.addr
	if addr == "" {
		addr = ":4433"
	}
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return err
	}
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return err
	}
	return s.ServeWithTLS(conn, tlsConfig)
}

func (s *QUICServer) acceptLoop() error {
	for {
		conn, err := s.listener.Accept(context.Background())
		if err != nil {
			return fmt.Errorf("quic accept: %w", err)
		}
		go s.handleConnection(conn)
	}
}

func (s *QUICServer) handleConnection(conn *quic.Conn) {
	// Accept the first bidirectional stream (the handshake stream).
	stream, err := conn.AcceptStream(conn.Context())
	if err != nil {
		slog.Error("quic: accept stream failed", "err", err)
		conn.CloseWithError(1, "failed to accept stream")
		return
	}

	// Read the handshake message to determine role.
	handshake, err := readMessage(stream)
	if err != nil {
		slog.Error("quic: read handshake failed", "err", err)
		conn.CloseWithError(1, "failed to read handshake")
		return
	}

	msg := string(handshake)
	switch {
	case msg == "register" || strings.HasPrefix(msg, "register:"):
		s.handleRegister(conn, stream, msg)
	case strings.HasPrefix(msg, "connect:"):
		s.handleConnect(conn, stream, msg)
	default:
		slog.Error("quic: unknown handshake", "msg", msg)
		conn.CloseWithError(1, "unknown handshake")
	}
}

func (s *QUICServer) handleRegister(conn *quic.Conn, stream *quic.Stream, msg string) {
	// Parse handshake: "register[:TOKEN[:INSTANCE_ID]]"
	var token, requestedID string
	if strings.HasPrefix(msg, "register:") {
		parts := strings.SplitN(msg[len("register:"):], ":", 2)
		token = parts[0]
		if len(parts) > 1 {
			requestedID = parts[1]
		}
	}

	if s.token != "" {
		if subtle.ConstantTimeCompare([]byte(token), []byte(s.token)) != 1 {
			slog.Warn("quic register: unauthorized")
			conn.CloseWithError(1, "unauthorized")
			return
		}
	}

	id := requestedID
	if id == "" {
		id = generateID()
	}

	// Send the instance ID back to the backend.
	if err := writeMessage(stream, []byte(id)); err != nil {
		slog.Error("quic register: write ID failed", "err", err)
		conn.CloseWithError(1, "failed to write ID")
		return
	}

	sess := &quicSession{conn: conn, stream: stream}
	inst := &instance{id: id, session: sess}
	s.hub.register(inst)
	defer s.hub.unregister(id)

	slog.Info("instance registered", "id", id, "transport", "quic")

	// Keep alive until backend disconnects.
	<-conn.Context().Done()
	slog.Info("instance disconnected", "id", id)
}

func (s *QUICServer) handleConnect(conn *quic.Conn, stream *quic.Stream, msg string) {
	instanceID := msg[len("connect:"):]
	if len(instanceID) == 0 || len(instanceID) > 64 {
		slog.Error("quic connect: invalid instance ID", "id", instanceID)
		conn.CloseWithError(1, "invalid instance ID")
		return
	}

	inst := s.hub.get(instanceID)
	if inst == nil {
		slog.Error("quic connect: instance not found", "id", instanceID)
		conn.CloseWithError(1, "instance not found")
		return
	}

	// Enforce single client per instance.
	inst.mu.Lock()
	if inst.occupied {
		inst.mu.Unlock()
		slog.Warn("quic connect: instance already occupied", "id", instanceID)
		conn.CloseWithError(1, "instance already has a connected client")
		return
	}
	inst.occupied = true
	inst.mu.Unlock()
	defer func() {
		inst.mu.Lock()
		inst.occupied = false
		inst.mu.Unlock()
	}()

	slog.Info("client connected", "instance", instanceID, "transport", "quic")

	clientSess := &quicSession{conn: conn, stream: stream}
	defer conn.CloseWithError(0, "")

	bridgeClient(inst, clientSess)
}

// Close shuts down the QUIC server.
func (s *QUICServer) Close() error {
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

// Addr returns the local address the server is listening on, or nil if
// not yet started.
func (s *QUICServer) Addr() net.Addr {
	if s.conn != nil {
		return s.conn.LocalAddr()
	}
	return nil
}

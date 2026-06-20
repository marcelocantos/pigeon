// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package pigeon

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"
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

func (w *quicStreamWrapper) ReadMessage() ([]byte, error) { return readMessage(w.stream) }
func (w *quicStreamWrapper) WriteMessage(data []byte) error {
	w.writeMu.Lock()
	defer w.writeMu.Unlock()
	return writeMessage(w.stream, data)
}
func (w *quicStreamWrapper) Close() error { return w.stream.Close() }

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
	hub       *hub
	auth      Auth
	addr      string
	listener  *quic.Listener
	conn      net.PacketConn
	tlsConfig *tls.Config
}

// NewQUICServer creates a raw QUIC relay server. The hub is shared with
// a WebTransport server so instances are visible to both protocols. The
// zero Auth means "accept all" (open relay); see BearerTokenAuth and
// MutualTLSAuth for the bundled defaults.
func NewQUICServer(addr string, tlsConfig *tls.Config, auth Auth, h *hub) *QUICServer {
	return &QUICServer{
		hub:  h,
		auth: auth,
		addr: addr,
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
		EnableDatagrams:      true,
		MaxIdleTimeout:       60 * time.Second,
		KeepAlivePeriod:      10 * time.Second,
		HandshakeIdleTimeout: 30 * time.Second,
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
	stream, err := conn.AcceptStream(conn.Context())
	if err != nil {
		slog.Error("quic: accept stream failed", "err", err)
		conn.CloseWithError(1, "failed to accept stream")
		return
	}
	handshake, err := readMessage(stream)
	if err != nil {
		slog.Error("quic: read handshake failed", "err", err)
		conn.CloseWithError(1, "failed to read handshake")
		return
	}
	dec, err := DecodeRelayGreeting(handshake)
	if err != nil {
		slog.Error("quic: bad handshake", "err", err)
		conn.CloseWithError(1, "bad handshake")
		return
	}
	sess := &quicSession{conn: conn, stream: stream}
	switch dec.Variant {
	case RelayGreetingRegister:
		s.handleRegister(conn, sess, dec.Token, dec.InstanceId)
	case RelayGreetingListen:
		s.handleListen(conn, sess, dec.Token, dec.InstanceId)
	case RelayGreetingConnect:
		s.handleConnect(conn, sess, dec.InstanceId)
	default:
		slog.Error("quic: unknown greeting variant", "variant", dec.Variant)
		conn.CloseWithError(1, "unknown greeting")
	}
}

func (s *QUICServer) handleRegister(conn *quic.Conn, sess *quicSession, token, requestedID string) {
	if s.auth.VerifyRegister != nil {
		tlsState := conn.ConnectionState().TLS
		req := &RegisterRequest{
			Token:      token,
			InstanceID: requestedID,
			TLS:        &tlsState,
			QUICConn:   conn,
		}
		if err := s.auth.VerifyRegister(conn.Context(), req); err != nil {
			slog.Warn("quic register: unauthorized", "err", err)
			conn.CloseWithError(1, "unauthorized")
			return
		}
	}
	id := requestedID
	if id == "" {
		id = generateID()
	}
	if err := sess.WriteMessage([]byte(id)); err != nil {
		slog.Error("quic register: write ID failed", "err", err)
		conn.CloseWithError(1, "failed to write ID")
		return
	}
	inst := newInstance(id, sess)
	s.hub.register(inst)
	defer s.hub.unregister(id)
	slog.Info("instance registered", "id", id, "transport", "quic")
	<-conn.Context().Done()
	slog.Info("instance disconnected", "id", id)
}

func (s *QUICServer) handleListen(conn *quic.Conn, sess *quicSession, token, instanceID string) {
	if instanceID == "" {
		slog.Error("quic listen: missing instance ID")
		conn.CloseWithError(1, "missing instance ID")
		return
	}
	if s.auth.VerifyRegister != nil {
		tlsState := conn.ConnectionState().TLS
		req := &RegisterRequest{
			Token:      token,
			InstanceID: instanceID,
			TLS:        &tlsState,
			QUICConn:   conn,
		}
		if err := s.auth.VerifyRegister(conn.Context(), req); err != nil {
			slog.Warn("quic listen: unauthorized", "err", err)
			conn.CloseWithError(1, "unauthorized")
			return
		}
	}
	inst := s.hub.get(instanceID)
	if inst == nil {
		slog.Warn("quic listen: instance not found", "id", instanceID)
		conn.CloseWithError(1, "instance not found")
		return
	}
	if err := sess.WriteMessage([]byte(instanceID)); err != nil {
		slog.Error("quic listen: write ack failed", "err", err)
		conn.CloseWithError(1, "write ack failed")
		return
	}
	if err := inst.parkListen(conn.Context(), sess); err != nil {
		slog.Info("quic listen: park ended", "id", instanceID, "err", err)
		conn.CloseWithError(0, "")
	}
}

func (s *QUICServer) handleConnect(conn *quic.Conn, sess *quicSession, instanceID string) {
	if len(instanceID) == 0 || len(instanceID) > 64 {
		slog.Error("quic connect: invalid instance ID", "id", instanceID)
		conn.CloseWithError(1, "invalid instance ID")
		return
	}
	if s.auth.VerifyConnect != nil {
		tlsState := conn.ConnectionState().TLS
		req := &ConnectRequest{
			InstanceID: instanceID,
			TLS:        &tlsState,
			QUICConn:   conn,
		}
		if err := s.auth.VerifyConnect(conn.Context(), req); err != nil {
			slog.Warn("quic connect: unauthorized", "err", err)
			conn.CloseWithError(1, "unauthorized")
			return
		}
	}
	inst := s.hub.get(instanceID)
	if inst == nil {
		slog.Warn("quic connect: instance not found", "id", instanceID)
		conn.CloseWithError(1, "instance not found")
		return
	}
	listen, err := inst.matchListen(conn.Context())
	if err != nil {
		slog.Warn("quic connect: no listener", "id", instanceID, "err", err)
		conn.CloseWithError(1, "no listener")
		return
	}
	if err := sess.WriteMessage([]byte("ok")); err != nil {
		slog.Error("quic connect: write ack failed", "err", err)
		conn.CloseWithError(1, "write ack failed")
		_ = listen.Close()
		return
	}
	slog.Info("client connected", "instance", instanceID, "transport", "quic")
	bridge(inst, sess, listen)
}

// Close shuts down the QUIC server. It closes the listener and the
// underlying PacketConn (if ServeWithTLS was called), releasing the
// quic-go transport's read goroutines and OS sockets.
func (s *QUICServer) Close() error {
	var err error
	if s.listener != nil {
		err = s.listener.Close()
	}
	if s.conn != nil {
		if cerr := s.conn.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}
	return err
}

// Addr returns the local address the server is listening on, or nil if
// not yet started.
func (s *QUICServer) Addr() net.Addr {
	if s.conn != nil {
		return s.conn.LocalAddr()
	}
	return nil
}

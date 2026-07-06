// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package pigeon

import (
	"context"
	"crypto/rand"
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

func (w *wtStreamWrapper) ReadMessage() ([]byte, error) { return readMessage(w.stream) }
func (w *wtStreamWrapper) WriteMessage(data []byte) error {
	w.writeMu.Lock()
	defer w.writeMu.Unlock()
	return writeMessage(w.stream, data)
}
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

// bearerToken extracts a relay credential from a WebTransport upgrade
// request: the "Bearer " Authorization header, or failing that the
// ?token= query parameter. Returns "" when neither is present.
func bearerToken(r *http.Request) string {
	const prefix = "Bearer "
	if h := r.Header.Get("Authorization"); strings.HasPrefix(h, prefix) {
		return h[len(prefix):]
	}
	return r.URL.Query().Get("token")
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
	auth     Auth
	addr     string
	conn     net.PacketConn
}

// NewWebTransportServer creates a WebTransport relay server listening on addr.
// The provided TLS config is used for the QUIC/HTTP3 connection (it may use
// static certificates or a dynamic GetCertificate callback such as certmagic).
// The zero Auth means "accept all" (open relay); see BearerTokenAuth and
// MutualTLSAuth for the bundled defaults.
func NewWebTransportServer(addr string, tlsConfig *tls.Config, auth Auth) (*WebTransportServer, error) {
	return NewWebTransportServerWithHub(addr, tlsConfig, auth, newHub())
}

// NewWebTransportServerWithHub creates a WebTransport relay server that
// shares the provided hub with other server types (e.g. raw QUIC).
func NewWebTransportServerWithHub(addr string, tlsConfig *tls.Config, auth Auth, h *hub) (*WebTransportServer, error) {
	mux := http.NewServeMux()
	s := &WebTransportServer{
		hub:  h,
		auth: auth,
		addr: addr,
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
				EnableDatagrams:      true,
				MaxIdleTimeout:       60 * time.Second,
				KeepAlivePeriod:      10 * time.Second,
				HandshakeIdleTimeout: 30 * time.Second,
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

	mux.HandleFunc("/pigeon", func(w http.ResponseWriter, r *http.Request) {
		s.handlePigeon(w, r)
	})

	return s, nil
}

// handlePigeon is the single WebTransport entry point under T45's
// remote-Listen L1 model. The primary stream's greeting decides the
// role: register / listen / connect.
func (s *WebTransportServer) handlePigeon(w http.ResponseWriter, r *http.Request) {
	session, err := s.wtServer.Upgrade(w, r)
	if err != nil {
		slog.Error("pigeon: upgrade failed", "err", err)
		return
	}
	stream, err := session.AcceptStream(session.Context())
	if err != nil {
		slog.Error("pigeon: accept primary stream failed", "err", err)
		session.CloseWithError(0, "failed to accept stream")
		return
	}
	handshake, err := readMessage(stream)
	if err != nil {
		slog.Error("pigeon: read handshake failed", "err", err)
		session.CloseWithError(0, "failed to read handshake")
		return
	}
	dec, err := DecodeRelayGreeting(handshake)
	if err != nil {
		slog.Error("pigeon: bad handshake", "err", err)
		session.CloseWithError(0, "bad handshake")
		return
	}
	// Browser WebTransport clients can't set a greeting token but can
	// set the Authorization header; fall back to it so BearerTokenAuth
	// (which reads req.Token) works for both. See RegisterRequest.Token.
	token := dec.Token
	if token == "" {
		token = bearerToken(r)
	}
	sess := &wtSession{session: session, stream: stream}
	switch dec.Variant {
	case RelayGreetingRegister:
		s.handleRegister(r, sess, token, dec.InstanceId)
	case RelayGreetingListen:
		s.handleListen(r, sess, token, dec.InstanceId)
	case RelayGreetingConnect:
		s.handleConnect(r, sess, dec.InstanceId)
	default:
		slog.Error("pigeon: unknown greeting variant", "variant", dec.Variant)
		session.CloseWithError(0, "unknown greeting")
	}
}

func (s *WebTransportServer) handleRegister(r *http.Request, sess *wtSession, token, requestedID string) {
	if s.auth.VerifyRegister != nil {
		req := &RegisterRequest{
			Token:       token,
			InstanceID:  requestedID,
			TLS:         r.TLS,
			HTTPRequest: r,
		}
		if err := s.auth.VerifyRegister(r.Context(), req); err != nil {
			slog.Warn("wt register: unauthorized", "err", err)
			_ = sess.Close()
			return
		}
	}
	id := requestedID
	if id == "" {
		id = generateID()
	}
	inst := newInstance(id, sess)
	if err := s.hub.register(inst); err != nil {
		slog.Warn("wt register: instance ID in use", "id", id, "err", err)
		_ = sess.Close()
		return
	}
	defer s.hub.unregister(inst)
	if err := sess.WriteMessage([]byte(id)); err != nil {
		slog.Error("register: write ID failed", "err", err)
		_ = sess.Close()
		return
	}
	slog.Info("instance registered", "id", id, "transport", "webtransport")
	<-sess.Context().Done()
	slog.Info("instance disconnected", "id", id)
}

func (s *WebTransportServer) handleListen(r *http.Request, sess *wtSession, token, instanceID string) {
	if instanceID == "" {
		slog.Error("wt listen: missing instance ID")
		_ = sess.Close()
		return
	}
	if s.auth.VerifyRegister != nil {
		req := &RegisterRequest{
			Token:       token,
			InstanceID:  instanceID,
			TLS:         r.TLS,
			HTTPRequest: r,
		}
		if err := s.auth.VerifyRegister(r.Context(), req); err != nil {
			slog.Warn("wt listen: unauthorized", "err", err)
			_ = sess.Close()
			return
		}
	}
	inst := s.hub.get(instanceID)
	if inst == nil {
		slog.Warn("wt listen: instance not found", "id", instanceID)
		_ = sess.Close()
		return
	}
	if err := sess.WriteMessage([]byte(instanceID)); err != nil {
		slog.Error("wt listen: write ack failed", "err", err)
		_ = sess.Close()
		return
	}
	if err := inst.parkListen(r.Context(), sess); err != nil {
		slog.Info("wt listen: park ended", "id", instanceID, "err", err)
		_ = sess.Close()
	}
}

func (s *WebTransportServer) handleConnect(r *http.Request, sess *wtSession, instanceID string) {
	if instanceID == "" {
		slog.Error("wt connect: missing instance ID")
		_ = sess.Close()
		return
	}
	if s.auth.VerifyConnect != nil {
		req := &ConnectRequest{
			InstanceID:  instanceID,
			TLS:         r.TLS,
			HTTPRequest: r,
		}
		if err := s.auth.VerifyConnect(r.Context(), req); err != nil {
			slog.Warn("wt connect: unauthorized", "err", err)
			_ = sess.Close()
			return
		}
	}
	inst := s.hub.get(instanceID)
	if inst == nil {
		slog.Warn("wt connect: instance not found", "id", instanceID)
		_ = sess.Close()
		return
	}
	listen, err := inst.matchListen(r.Context())
	if err != nil {
		slog.Warn("wt connect: no listener", "id", instanceID, "err", err)
		_ = sess.Close()
		return
	}
	if err := sess.WriteMessage([]byte("ok")); err != nil {
		slog.Error("wt connect: write ack failed", "err", err)
		_ = sess.Close()
		_ = listen.Close()
		return
	}
	slog.Info("client connected", "instance", instanceID, "transport", "webtransport")
	bridge(inst, sess, listen)
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

// Close shuts down the server. It closes the WebTransport server and the
// underlying PacketConn (if Serve was called), releasing read goroutines
// and OS sockets held by the QUIC stack.
func (s *WebTransportServer) Close() error {
	err := s.wtServer.Close()
	if s.conn != nil {
		if cerr := s.conn.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}
	return err
}

// Addr returns the local address the server is listening on, or nil if
// Serve/ListenAndServe has not been called.
func (s *WebTransportServer) Addr() net.Addr {
	if s.conn != nil {
		return s.conn.LocalAddr()
	}
	return nil
}

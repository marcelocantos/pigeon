// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package tern

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"sync"

	"github.com/quic-go/quic-go/http3"
	"github.com/quic-go/webtransport-go"
)

// maxWTMessageSize is the maximum message size for WebTransport relay
// frames.
const maxWTMessageSize = 1 << 20 // 1 MiB

// wtHub manages WebTransport backend instances.
type wtHub struct {
	mu        sync.RWMutex
	instances map[string]*wtInstance
}

type wtInstance struct {
	id      string
	session *webtransport.Session
	stream  *webtransport.Stream // primary bidirectional stream

	mu       sync.Mutex
	occupied bool // true when a client is connected
}

func newWTHub() *wtHub {
	return &wtHub{instances: make(map[string]*wtInstance)}
}

func (h *wtHub) register(inst *wtInstance) {
	h.mu.Lock()
	h.instances[inst.id] = inst
	h.mu.Unlock()
}

func (h *wtHub) unregister(id string) {
	h.mu.Lock()
	delete(h.instances, id)
	h.mu.Unlock()
}

func (h *wtHub) get(id string) *wtInstance {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.instances[id]
}

// generateID generates a random instance ID.
func generateID() string {
	b := make([]byte, 4)
	rand.Read(b)
	n := uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
	return strconv.FormatUint(uint64(n), 36)
}

// writeWTMessage writes a length-prefixed binary message to a stream.
// Format: [4-byte big-endian length][payload]
func writeWTMessage(stream io.Writer, data []byte) error {
	if len(data) > maxWTMessageSize {
		return fmt.Errorf("message too large: %d > %d", len(data), maxWTMessageSize)
	}
	var hdr [4]byte
	binary.BigEndian.PutUint32(hdr[:], uint32(len(data)))
	if _, err := stream.Write(hdr[:]); err != nil {
		return err
	}
	_, err := stream.Write(data)
	return err
}

// readWTMessage reads a length-prefixed binary message from a stream.
func readWTMessage(stream io.Reader) ([]byte, error) {
	var hdr [4]byte
	if _, err := io.ReadFull(stream, hdr[:]); err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint32(hdr[:])
	if length > maxWTMessageSize {
		return nil, fmt.Errorf("message too large: %d > %d", length, maxWTMessageSize)
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
	hub      *wtHub
	token    string // bearer token for /register auth; empty = open
	addr     string
	conn     net.PacketConn
}

// NewWebTransportServer creates a WebTransport relay server listening on addr.
// The provided TLS certificate is used for the QUIC/HTTP3 connection.
// If token is non-empty, /register requires a matching Bearer token.
func NewWebTransportServer(addr string, tlsCert tls.Certificate, token string) (*WebTransportServer, error) {
	hub := newWTHub()

	mux := http.NewServeMux()
	s := &WebTransportServer{
		hub:   hub,
		token: token,
		addr:  addr,
	}

	wtServer := &webtransport.Server{
		H3: &http3.Server{
			Addr:    addr,
			Handler: mux,
			TLSConfig: &tls.Config{
				Certificates: []tls.Certificate{tlsCert},
				NextProtos:   []string{http3.NextProtoH3},
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
		if auth != "Bearer "+s.token {
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

	// Read and discard the handshake message.
	if _, err := readWTMessage(stream); err != nil {
		slog.Error("register: read handshake failed", "err", err)
		session.CloseWithError(0, "failed to read handshake")
		return
	}

	id := generateID()

	// Send the instance ID to the backend.
	if err := writeWTMessage(stream, []byte(id)); err != nil {
		slog.Error("register: write ID failed", "err", err)
		session.CloseWithError(0, "failed to write ID")
		return
	}

	inst := &wtInstance{id: id, session: session, stream: stream}
	s.hub.register(inst)
	defer s.hub.unregister(id)

	slog.Info("instance registered", "id", id)

	// Keep alive until backend disconnects.
	<-session.Context().Done()
	slog.Info("instance disconnected", "id", id)
}

func (s *WebTransportServer) handleClient(w http.ResponseWriter, r *http.Request) {
	// Extract instance ID from path: /ws/{id}
	instanceID := r.URL.Path[len("/ws/"):]
	if instanceID == "" {
		http.Error(w, `{"error":"missing instance ID"}`, http.StatusBadRequest)
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

	// Read and discard the handshake message.
	if _, err := readWTMessage(clientStream); err != nil {
		slog.Error("client: read handshake failed", "err", err)
		return
	}

	ctx := session.Context()
	slog.Info("client connected", "instance", instanceID)

	errCh := make(chan error, 3)

	// backend stream -> client stream
	go func() {
		for {
			msg, err := readWTMessage(inst.stream)
			if err != nil {
				errCh <- fmt.Errorf("read backend: %w", err)
				return
			}
			if err := writeWTMessage(clientStream, msg); err != nil {
				errCh <- fmt.Errorf("write client: %w", err)
				return
			}
		}
	}()

	// client stream -> backend stream
	go func() {
		for {
			msg, err := readWTMessage(clientStream)
			if err != nil {
				errCh <- fmt.Errorf("read client: %w", err)
				return
			}
			inst.mu.Lock()
			err = writeWTMessage(inst.stream, msg)
			inst.mu.Unlock()
			if err != nil {
				errCh <- fmt.Errorf("write backend: %w", err)
				return
			}
		}
	}()

	// backend datagrams -> client datagrams
	go func() {
		for {
			data, err := inst.session.ReceiveDatagram(ctx)
			if err != nil {
				return
			}
			if err := session.SendDatagram(data); err != nil {
				return
			}
		}
	}()

	// client datagrams -> backend datagrams
	go func() {
		for {
			data, err := session.ReceiveDatagram(ctx)
			if err != nil {
				return
			}
			if err := inst.session.SendDatagram(data); err != nil {
				return
			}
		}
	}()

	// Wait for stream relay to end or session to close.
	select {
	case err := <-errCh:
		slog.Info("client disconnected", "instance", instanceID, "reason", err)
	case <-ctx.Done():
		slog.Info("client session ended", "instance", instanceID)
	case <-inst.session.Context().Done():
		slog.Info("backend session ended", "instance", instanceID)
	}
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

// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package tern

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
)

// relaySession abstracts a relay peer's connection. Both WebTransport
// sessions and raw QUIC connections implement this interface.
type relaySession interface {
	// ReadMessage reads a length-prefixed message from the primary stream.
	ReadMessage() ([]byte, error)
	// WriteMessage writes a length-prefixed message to the primary stream.
	WriteMessage(data []byte) error
	// AcceptStream accepts an incoming bidirectional stream.
	AcceptStream(ctx context.Context) (readWriteCloserPair, error)
	// OpenStream opens a new bidirectional stream to the peer.
	OpenStream() (readWriteCloserPair, error)
	// SendDatagram sends an unreliable datagram.
	SendDatagram(data []byte) error
	// ReceiveDatagram receives the next datagram.
	ReceiveDatagram(ctx context.Context) ([]byte, error)
	// Context returns the session lifecycle context.
	Context() context.Context
	// Close closes the session.
	Close() error
}

// readWriteCloserPair is a bidirectional stream that supports
// length-prefixed message framing.
type readWriteCloserPair interface {
	ReadMessage() ([]byte, error)
	WriteMessage(data []byte) error
	Close() error
}

// hub manages registered backend instances. It is shared between the
// WebTransport and raw QUIC server paths.
type hub struct {
	mu        sync.RWMutex
	instances map[string]*instance
}

type instance struct {
	id      string
	session relaySession

	mu       sync.Mutex
	occupied bool // true when a client is connected
}

func newHub() *hub {
	return &hub{instances: make(map[string]*instance)}
}

func (h *hub) register(inst *instance) {
	h.mu.Lock()
	h.instances[inst.id] = inst
	h.mu.Unlock()
}

func (h *hub) unregister(id string) {
	h.mu.Lock()
	delete(h.instances, id)
	h.mu.Unlock()
}

func (h *hub) get(id string) *instance {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.instances[id]
}

// bridgeClient connects a client session to a registered backend instance.
// It relays messages and datagrams bidirectionally until one side disconnects.
//
// Note: a slow consumer can block the producer (backpressure propagates
// through the relay). This is inherent to the 1:1 relay design. For
// protection against malicious slow clients, add bounded message buffers
// with drop-on-overflow.
func bridgeClient(inst *instance, clientSession relaySession) {
	ctx, cancel := context.WithCancel(clientSession.Context())
	defer cancel()

	var wg sync.WaitGroup
	errCh := make(chan error, 2)

	// backend stream -> client stream
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			msg, err := inst.session.ReadMessage()
			if err != nil {
				errCh <- fmt.Errorf("read backend: %w", err)
				return
			}
			if err := clientSession.WriteMessage(msg); err != nil {
				errCh <- fmt.Errorf("write client: %w", err)
				return
			}
		}
	}()

	// client stream -> backend stream
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			msg, err := clientSession.ReadMessage()
			if err != nil {
				errCh <- fmt.Errorf("read client: %w", err)
				return
			}
			if err := inst.session.WriteMessage(msg); err != nil {
				errCh <- fmt.Errorf("write backend: %w", err)
				return
			}
		}
	}()

	// backend datagrams -> client datagrams
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			data, err := inst.session.ReceiveDatagram(ctx)
			if err != nil {
				return
			}
			if err := clientSession.SendDatagram(data); err != nil {
				return
			}
		}
	}()

	// client datagrams -> backend datagrams
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			data, err := clientSession.ReceiveDatagram(ctx)
			if err != nil {
				return
			}
			if err := inst.session.SendDatagram(data); err != nil {
				return
			}
		}
	}()

	// Bridge additional streams opened by either side.
	// backend opens stream -> forward to client
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			backendStream, err := inst.session.AcceptStream(ctx)
			if err != nil {
				return
			}
			clientStream, err := clientSession.OpenStream()
			if err != nil {
				backendStream.Close()
				return
			}
			// Bridge this stream pair in both directions.
			wg.Add(2)
			go func() {
				defer wg.Done()
				bridgeStream(backendStream, clientStream)
			}()
			go func() {
				defer wg.Done()
				bridgeStream(clientStream, backendStream)
			}()
		}
	}()

	// client opens stream -> forward to backend
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			clientStream, err := clientSession.AcceptStream(ctx)
			if err != nil {
				return
			}
			backendStream, err := inst.session.OpenStream()
			if err != nil {
				clientStream.Close()
				return
			}
			wg.Add(2)
			go func() {
				defer wg.Done()
				bridgeStream(clientStream, backendStream)
			}()
			go func() {
				defer wg.Done()
				bridgeStream(backendStream, clientStream)
			}()
		}
	}()

	// Wait for stream relay to end or session to close.
	select {
	case err := <-errCh:
		slog.Info("client disconnected", "instance", inst.id, "reason", err)
	case <-ctx.Done():
		slog.Info("client session ended", "instance", inst.id)
	case <-inst.session.Context().Done():
		slog.Info("backend session ended", "instance", inst.id)
	}

	cancel() // stop datagram goroutines and stream accept loops
	wg.Wait()
}

// bridgeStream copies messages from src to dst until an error occurs.
func bridgeStream(src, dst readWriteCloserPair) {
	defer src.Close()
	defer dst.Close()
	for {
		msg, err := src.ReadMessage()
		if err != nil {
			return
		}
		if err := dst.WriteMessage(msg); err != nil {
			return
		}
	}
}

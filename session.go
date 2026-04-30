// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package pigeon

import (
	"context"
	"encoding/binary"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
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

// instance is a registered backend. In mux mode, multiple clients may
// bridge to the same instance concurrently; the relay tags each client
// with a 4-byte clientTag and prepends it to every stream and datagram
// it forwards to the backend, so the backend can demux per-client. In
// pair mode (the legacy 1:1 path used by the pairing ceremony), the
// relay does not tag — it bridges client.primary ↔ backend.primary
// directly.
type instance struct {
	id      string
	session relaySession
	muxMode bool

	mu      sync.Mutex
	clients map[uint32]*clientSlot

	nextTag       atomic.Uint32
	dispatchOnce  sync.Once
	dispatchCtx   context.Context
	dispatchStop  context.CancelFunc
	dispatchReady chan struct{}
}

// clientSlot pairs a clientTag with the relay-side session for that
// client. It is registered on bridgeClient entry and removed on exit.
type clientSlot struct {
	tag    uint32
	client relaySession
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

// startDispatch lazily starts the per-instance goroutines that demux
// backend-side streams and datagrams to the right client by the
// 4-byte clientTag prefix.
func (inst *instance) startDispatch() {
	inst.dispatchOnce.Do(func() {
		inst.dispatchCtx, inst.dispatchStop = context.WithCancel(inst.session.Context())
		inst.dispatchReady = make(chan struct{})
		if inst.clients == nil {
			inst.clients = make(map[uint32]*clientSlot)
		}
		go inst.dispatchStreams()
		go inst.dispatchDatagrams()
		close(inst.dispatchReady)
	})
}

// dispatchStreams reads streams the backend opens (rare) and routes
// them to the corresponding client by the 4-byte tag header.
func (inst *instance) dispatchStreams() {
	for {
		stream, err := inst.session.AcceptStream(inst.dispatchCtx)
		if err != nil {
			return
		}
		header, err := stream.ReadMessage()
		if err != nil {
			_ = stream.Close()
			continue
		}
		if len(header) < 4 {
			_ = stream.Close()
			continue
		}
		tag := binary.BigEndian.Uint32(header[:4])
		inst.mu.Lock()
		slot, ok := inst.clients[tag]
		inst.mu.Unlock()
		if !ok {
			_ = stream.Close()
			continue
		}
		clientStream, err := slot.client.OpenStream()
		if err != nil {
			_ = stream.Close()
			continue
		}
		// First message to client: original header without the 4-byte tag.
		if err := clientStream.WriteMessage(header[4:]); err != nil {
			_ = stream.Close()
			_ = clientStream.Close()
			continue
		}
		go bridgeStream(stream, clientStream)
		go bridgeStream(clientStream, stream)
	}
}

// dispatchDatagrams reads datagrams the backend sends and routes them
// to the corresponding client by the 4-byte tag prefix.
func (inst *instance) dispatchDatagrams() {
	for {
		data, err := inst.session.ReceiveDatagram(inst.dispatchCtx)
		if err != nil {
			return
		}
		if len(data) < 4 {
			continue
		}
		tag := binary.BigEndian.Uint32(data[:4])
		inst.mu.Lock()
		slot, ok := inst.clients[tag]
		inst.mu.Unlock()
		if !ok {
			continue
		}
		_ = slot.client.SendDatagram(data[4:])
	}
}

func (inst *instance) addClient(slot *clientSlot) {
	inst.mu.Lock()
	if inst.clients == nil {
		inst.clients = make(map[uint32]*clientSlot)
	}
	inst.clients[slot.tag] = slot
	inst.mu.Unlock()
}

func (inst *instance) removeClient(tag uint32) {
	inst.mu.Lock()
	delete(inst.clients, tag)
	inst.mu.Unlock()
}

// streamTracker tracks active bridge streams so they can be closed
// externally, unblocking goroutines stuck in pending reads.
type streamTracker struct {
	mu      sync.Mutex
	streams []readWriteCloserPair
}

func (t *streamTracker) add(s readWriteCloserPair) {
	t.mu.Lock()
	t.streams = append(t.streams, s)
	t.mu.Unlock()
}

func (t *streamTracker) closeAll() {
	t.mu.Lock()
	streams := t.streams
	t.streams = nil
	t.mu.Unlock()
	for _, s := range streams {
		_ = s.Close()
	}
}

// bridgeClient dispatches to the right bridge implementation based on
// the instance's registration mode.
func bridgeClient(inst *instance, clientSession relaySession) {
	if inst.muxMode {
		bridgeClientMux(inst, clientSession)
	} else {
		bridgeClientPair(inst, clientSession)
	}
}

// bridgeClientMux connects a client session to a mux-mode instance. The
// relay tags this client with a unique 4-byte clientTag (per-instance
// counter) and prepends the tag to:
//   - every stream it opens on the backend (one for the client's primary,
//     plus one per sub-stream the client opens), and
//   - every datagram it forwards from this client to the backend.
//
// The backend strips the tag to demux to the right Session.
func bridgeClientMux(inst *instance, clientSession relaySession) {
	inst.startDispatch()

	ctx, cancel := context.WithCancel(clientSession.Context())
	defer cancel()

	tag := inst.nextTag.Add(1)
	tagBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(tagBytes, tag)

	slot := &clientSlot{tag: tag, client: clientSession}
	inst.addClient(slot)
	defer inst.removeClient(tag)

	tracker := &streamTracker{}

	// Open the per-client primary stream on the backend. The first
	// message: 4-byte tag prefix + the empty name header the client wrote.
	backendPrimary, err := inst.session.OpenStream()
	if err != nil {
		slog.Warn("bridgeClient: open backend primary", "err", err)
		return
	}
	tracker.add(backendPrimary)

	// Read first message off client primary (the empty-name stream header)
	// and forward it to backend primary with the tag prefix.
	clientHeader, err := clientSession.ReadMessage()
	if err != nil {
		slog.Debug("bridgeClient: read client primary header", "err", err)
		_ = backendPrimary.Close()
		return
	}
	if err := backendPrimary.WriteMessage(append(tagBytes, clientHeader...)); err != nil {
		slog.Debug("bridgeClient: write backend primary header", "err", err)
		_ = backendPrimary.Close()
		return
	}

	var wg sync.WaitGroup
	errCh := make(chan error, 2)

	// Pump primary client → backend (tag-less; bytes already framed by client).
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			msg, err := clientSession.ReadMessage()
			if err != nil {
				errCh <- fmt.Errorf("read client primary: %w", err)
				return
			}
			if err := backendPrimary.WriteMessage(msg); err != nil {
				errCh <- fmt.Errorf("write backend primary: %w", err)
				return
			}
		}
	}()
	// Pump primary backend → client.
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			msg, err := backendPrimary.ReadMessage()
			if err != nil {
				errCh <- fmt.Errorf("read backend primary: %w", err)
				return
			}
			if err := clientSession.WriteMessage(msg); err != nil {
				errCh <- fmt.Errorf("write client primary: %w", err)
				return
			}
		}
	}()

	// Forward client-opened sub-streams: read first message from client,
	// open a backend stream, write [tag][clientFirstMessage] as the
	// header, then byte-pump messages both directions.
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			clientStream, err := clientSession.AcceptStream(ctx)
			if err != nil {
				return
			}
			header, err := clientStream.ReadMessage()
			if err != nil {
				_ = clientStream.Close()
				continue
			}
			backendStream, err := inst.session.OpenStream()
			if err != nil {
				_ = clientStream.Close()
				continue
			}
			if err := backendStream.WriteMessage(append(tagBytes, header...)); err != nil {
				_ = clientStream.Close()
				_ = backendStream.Close()
				continue
			}
			tracker.add(clientStream)
			tracker.add(backendStream)
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

	// Forward client-side datagrams to backend with tag prefix.
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			data, err := clientSession.ReceiveDatagram(ctx)
			if err != nil {
				return
			}
			framed := make([]byte, 4+len(data))
			copy(framed, tagBytes)
			copy(framed[4:], data)
			if err := inst.session.SendDatagram(framed); err != nil {
				return
			}
		}
	}()

	// Wait for either side to disconnect.
	select {
	case err := <-errCh:
		slog.Info("client disconnected", "instance", inst.id, "tag", tag, "reason", err)
	case <-ctx.Done():
		slog.Info("client session ended", "instance", inst.id, "tag", tag)
	case <-inst.session.Context().Done():
		slog.Info("backend session ended", "instance", inst.id)
	}

	cancel()
	tracker.closeAll()
	wg.Wait()
}

// bridgeClientPair connects a client session to a pair-mode instance.
// This is the legacy 1:1 bridge used by the pairing ceremony. The relay
// bridges client.primary ↔ backend.primary directly, and forwards each
// stream the client opens to a fresh stream on the backend (and vice
// versa). No tag prefix is added.
func bridgeClientPair(inst *instance, clientSession relaySession) {
	ctx, cancel := context.WithCancel(clientSession.Context())
	defer cancel()

	var wg sync.WaitGroup
	errCh := make(chan error, 2)
	tracker := &streamTracker{}

	// backend stream → client stream
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

	// client stream → backend stream
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

	// backend datagrams → client datagrams
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

	// client datagrams → backend datagrams
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

	// Forward streams either side opens.
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
				_ = backendStream.Close()
				return
			}
			tracker.add(backendStream)
			tracker.add(clientStream)
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
				_ = clientStream.Close()
				return
			}
			tracker.add(clientStream)
			tracker.add(backendStream)
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

	select {
	case err := <-errCh:
		slog.Info("client disconnected (pair)", "instance", inst.id, "reason", err)
	case <-ctx.Done():
		slog.Info("client session ended (pair)", "instance", inst.id)
	case <-inst.session.Context().Done():
		slog.Info("backend session ended (pair)", "instance", inst.id)
	}

	cancel()
	tracker.closeAll()
	wg.Wait()
}

// bridgeStream copies messages from src to dst until an error occurs.
func bridgeStream(src, dst readWriteCloserPair) {
	defer src.Close()
	defer dst.Close()
	for {
		msg, err := src.ReadMessage()
		if err != nil {
			slog.Debug("bridgeStream: read error", "err", err)
			return
		}
		if err := dst.WriteMessage(msg); err != nil {
			slog.Debug("bridgeStream: write error", "err", err)
			return
		}
	}
}

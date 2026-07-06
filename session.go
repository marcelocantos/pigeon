// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package pigeon

import (
	"context"
	"errors"
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

// instance is a registered backend under T45's remote-Listen L1 model.
// The control connection (from `register:<id>`) is kept open purely for
// lifetime tracking; no traffic flows on it. Each `listen:<id>` from
// the backend parks a fresh QUIC connection on `listens`, where the
// next arriving client pairs with it.
type instance struct {
	id      string
	control relaySession

	ctx     context.Context
	cancel  context.CancelFunc
	listens chan relaySession
}

func newHub() *hub {
	return &hub{instances: make(map[string]*instance)}
}

// register claims the instance's ID for its owning connection. It
// refuses to displace an instance that is still live (its context not
// yet cancelled), so a peer — malicious or a racing reconnect — cannot
// hijack a live backend's ID via last-writer-wins. A dead-but-not-yet-
// unregistered slot may be taken over; the departing owner's unregister
// is ownership-checked and will not disturb the new instance.
func (h *hub) register(inst *instance) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if cur, ok := h.instances[inst.id]; ok && cur.ctx.Err() == nil {
		return fmt.Errorf("instance %q already registered", inst.id)
	}
	h.instances[inst.id] = inst
	return nil
}

// unregister removes and cancels inst only if it still owns its ID slot
// (pointer identity). A stale teardown whose slot was taken over by a
// re-registered instance is a no-op, so an old connection's departure
// can never cancel the new live instance.
func (h *hub) unregister(inst *instance) {
	h.mu.Lock()
	cur, ok := h.instances[inst.id]
	owns := ok && cur == inst
	if owns {
		delete(h.instances, inst.id)
	}
	h.mu.Unlock()
	if owns {
		inst.cancel()
	}
}

func (h *hub) get(id string) *instance {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.instances[id]
}

// newInstance constructs an instance with its lifecycle context tied to
// the register-side control connection.
func newInstance(id string, control relaySession) *instance {
	ctx, cancel := context.WithCancel(control.Context())
	return &instance{
		id:      id,
		control: control,
		ctx:     ctx,
		cancel:  cancel,
		listens: make(chan relaySession),
	}
}

// parkListen blocks until the listen is matched to an arriving client
// or the instance is torn down. Returns nil if matched (caller no
// longer owns the session), or an error if the park was cancelled.
func (inst *instance) parkListen(ctx context.Context, listen relaySession) error {
	select {
	case inst.listens <- listen:
		return nil
	case <-inst.ctx.Done():
		return errors.New("instance unregistered")
	case <-ctx.Done():
		return ctx.Err()
	}
}

// matchListen blocks until the backend parks a listen for this
// instance. Returns the parked relaySession (now owned by the caller)
// or an error if the wait was cancelled.
func (inst *instance) matchListen(ctx context.Context) (relaySession, error) {
	select {
	case listen := <-inst.listens:
		return listen, nil
	case <-inst.ctx.Done():
		return nil, errors.New("instance unregistered")
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// bridge byte-shovels two QUIC-shaped relaySessions together. Each
// pipe is opaque to the relay: streams, datagrams, and the post-
// greeting traffic on the primary stream are forwarded both ways
// without inspection. The bridge runs until either side disconnects.
func bridge(inst *instance, client, backend relaySession) {
	ctx, cancel := context.WithCancel(client.Context())
	defer cancel()

	var wg sync.WaitGroup
	errCh := make(chan error, 4)
	tracker := &streamTracker{}

	// backend primary → client primary
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			msg, err := backend.ReadMessage()
			if err != nil {
				errCh <- fmt.Errorf("read backend primary: %w", err)
				return
			}
			if err := client.WriteMessage(msg); err != nil {
				errCh <- fmt.Errorf("write client primary: %w", err)
				return
			}
		}
	}()

	// client primary → backend primary
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			msg, err := client.ReadMessage()
			if err != nil {
				errCh <- fmt.Errorf("read client primary: %w", err)
				return
			}
			if err := backend.WriteMessage(msg); err != nil {
				errCh <- fmt.Errorf("write backend primary: %w", err)
				return
			}
		}
	}()

	// backend datagrams → client datagrams
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			data, err := backend.ReceiveDatagram(ctx)
			if err != nil {
				return
			}
			if err := client.SendDatagram(data); err != nil {
				return
			}
		}
	}()

	// client datagrams → backend datagrams
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			data, err := client.ReceiveDatagram(ctx)
			if err != nil {
				return
			}
			if err := backend.SendDatagram(data); err != nil {
				return
			}
		}
	}()

	// Streams opened by either side → forwarded to the other.
	wg.Add(2)
	go func() {
		defer wg.Done()
		for {
			srcStream, err := backend.AcceptStream(ctx)
			if err != nil {
				return
			}
			dstStream, err := client.OpenStream()
			if err != nil {
				_ = srcStream.Close()
				return
			}
			tracker.add(srcStream)
			tracker.add(dstStream)
			wg.Add(2)
			go func() {
				defer wg.Done()
				bridgeStream(srcStream, dstStream)
			}()
			go func() {
				defer wg.Done()
				bridgeStream(dstStream, srcStream)
			}()
		}
	}()
	go func() {
		defer wg.Done()
		for {
			srcStream, err := client.AcceptStream(ctx)
			if err != nil {
				return
			}
			dstStream, err := backend.OpenStream()
			if err != nil {
				_ = srcStream.Close()
				return
			}
			tracker.add(srcStream)
			tracker.add(dstStream)
			wg.Add(2)
			go func() {
				defer wg.Done()
				bridgeStream(srcStream, dstStream)
			}()
			go func() {
				defer wg.Done()
				bridgeStream(dstStream, srcStream)
			}()
		}
	}()

	select {
	case err := <-errCh:
		slog.Info("bridge: peer disconnected", "instance", inst.id, "reason", err)
	case <-ctx.Done():
		slog.Info("bridge: client session ended", "instance", inst.id)
	case <-backend.Context().Done():
		slog.Info("bridge: backend session ended", "instance", inst.id)
	}

	cancel()
	tracker.closeAll()
	_ = client.Close()
	_ = backend.Close()
	wg.Wait()
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

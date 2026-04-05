// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package tern

import (
	"context"
	"io"
	"log/slog"
	"math"
	"math/rand/v2"
	"time"

	"github.com/marcelocantos/tern/crypto"
	"github.com/marcelocantos/tern/protocol"
)

// event carries an event ID plus optional payload into the executor.
type event struct {
	id      EventID
	payload any // typed per event
}

// sendRequest is the payload for EvAppSend.
type sendRequest struct {
	data []byte
	done chan<- error
}

// recvResult is delivered via recvWaiters when data arrives.
type recvResult struct {
	data []byte
	err  error
}

// dgRecvResult is delivered via dgRecvWaiters.
type dgRecvResult struct {
	data []byte
	err  error
}

// streamData is the payload for relay/LAN stream data events.
type streamData struct {
	data []byte
}

// streamError is the payload for relay/LAN stream error events.
type streamError struct {
	err error
}

// datagramData is the payload for relay/LAN datagram events.
type datagramData struct {
	data []byte
}

// lanDialResult is the payload for EvLANDialOk.
type lanDialResult struct {
	stream   io.ReadWriteCloser
	dg       datagrammer
	closer   io.Closer
	opener   streamOpener
	acceptor streamAcceptor
}

// executor mediates between the application API, I/O, and the
// transport state machine. All state transitions happen in its
// run() loop — no goroutine makes independent state decisions.
type executor struct {
	machine interface {
		HandleEvent(ev EventID) ([]protocol.CmdID, error)
	}

	// Event channel — all events flow through here.
	events chan event

	// I/O resources.
	relay  *path // permanent
	lan    *path // nil when no LAN path

	// Application response waiters and buffered data.
	recvWaiters   []chan recvResult
	recvBuffer    [][]byte // data received before a waiter was registered
	dgRecvWaiters []chan dgRecvResult
	dgRecvBuffer  [][]byte

	// Encryption (executor-level, not machine-level).
	channel   *crypto.Channel
	dgChannel *crypto.Channel

	// LAN server (backend only — for re-advertisement).
	lanServer  *LANServer
	lanEnabled bool

	// LAN ready signal.
	lanReady chan struct{}

	// Timers managed by the executor.
	monitorCancel  context.CancelFunc
	backoffCancel  context.CancelFunc

	// LAN reader cancellation.
	lanStreamCancel context.CancelFunc
	lanDgCancel     context.CancelFunc

	// Fragment reassembly.
	reasm        *reassembler
	maxDgPayload int

	// Lifecycle.
	ctx    context.Context
	cancel context.CancelFunc
}

func newExecutor(
	ctx context.Context,
	cancel context.CancelFunc,
	machine interface {
		HandleEvent(ev EventID) ([]protocol.CmdID, error)
	},
	relay *path,
) *executor {
	done := make(chan struct{})
	go func() {
		<-ctx.Done()
		close(done)
	}()

	e := &executor{
		machine:      machine,
		events:       make(chan event, 64),
		relay:        relay,
		lanReady:     make(chan struct{}),
		reasm:        newReassembler(DefaultFragmentTimeout, done),
		maxDgPayload: DefaultMaxDatagramPayload,
		ctx:          ctx,
		cancel:       cancel,
	}

	// Start relay readers — they run for the lifetime of the connection.
	go e.streamReader(ctx, relay, EventRelayStreamData, EventRelayStreamError)
	go e.datagramReader(ctx, relay, EventRelayDatagram)

	// Start the event loop.
	go e.run()

	return e
}

// send submits an app_send event and blocks until the write completes.
func (e *executor) send(ctx context.Context, data []byte) error {
	done := make(chan error, 1)
	e.submit(event{id: EventAppSend, payload: &sendRequest{data: data, done: done}})
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	case <-e.ctx.Done():
		return e.ctx.Err()
	}
}

// recv submits an app_recv waiter and blocks until data arrives.
func (e *executor) recv(ctx context.Context) ([]byte, error) {
	result := make(chan recvResult, 1)
	e.submit(event{id: EventAppRecv, payload: result})
	select {
	case r := <-result:
		return r.data, r.err
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-e.ctx.Done():
		return nil, e.ctx.Err()
	}
}

// submit posts an event to the executor. Non-blocking if the channel
// has capacity; blocks if full (backpressure).
func (e *executor) submit(ev event) {
	select {
	case e.events <- ev:
	case <-e.ctx.Done():
	}
}

// run is the main event loop. All state transitions happen here.
func (e *executor) run() {
	for {
		select {
		case ev := <-e.events:
			// Recv is executor-level: register a waiter, deliver when
			// data arrives. Not a machine transition. If buffered data
			// exists, deliver immediately.
			if ev.id == EventAppRecv {
				if ch, ok := ev.payload.(chan recvResult); ok {
					if len(e.recvBuffer) > 0 {
						ch <- recvResult{data: e.recvBuffer[0]}
						e.recvBuffer = e.recvBuffer[1:]
					} else {
						e.recvWaiters = append(e.recvWaiters, ch)
					}
				}
				continue
			}

			cmds, err := e.machine.HandleEvent(ev.id)
			if err != nil {
				slog.Warn("machine error", "event", ev.id, "err", err)
				continue
			}
			for _, cmd := range cmds {
				e.executeCommand(cmd, ev.payload)
			}

			// Events with no matching transition may still need
			// direct handling (e.g., stream data delivery when the
			// machine has no specific transition for this state).
			if cmds == nil {
				e.handleUnmatchedEvent(ev)
			}

		case <-e.ctx.Done():
			return
		}
	}
}

// handleUnmatchedEvent handles events that the machine didn't match
// (no transition from current state for this event). Some events
// need executor-level handling regardless of machine state.
func (e *executor) handleUnmatchedEvent(ev event) {
	switch ev.id {
	case EventRelayStreamData:
		// Data arrived on relay — deliver to waiting Recv if any.
		if sd, ok := ev.payload.(*streamData); ok {
			e.deliverRecv(sd.data)
		}
	case EventLanStreamData:
		// Data arrived on LAN — deliver to waiting Recv if any.
		if sd, ok := ev.payload.(*streamData); ok {
			e.deliverRecv(sd.data)
		}
	case EventRelayDatagram:
		if dd, ok := ev.payload.(*datagramData); ok {
			e.deliverDatagram(dd.data)
		}
	case EventLanDatagram:
		if dd, ok := ev.payload.(*datagramData); ok {
			e.deliverDatagram(dd.data)
		}
	case EventRelayStreamError:
		if se, ok := ev.payload.(*streamError); ok {
			e.deliverRecvError(se.err)
		}
	case EventLanStreamError:
		if se, ok := ev.payload.(*streamError); ok {
			e.deliverRecvError(se.err)
		}
	}
}

// executeCommand carries out a single command from the machine.
func (e *executor) executeCommand(cmd CmdID, payload any) {
	switch cmd {
	case CmdWriteActiveStream:
		if req, ok := payload.(*sendRequest); ok {
			p := e.activePath()
			var msg []byte
			if e.channel != nil {
				framed := make([]byte, 1+len(req.data))
				framed[0] = msgApp
				copy(framed[1:], req.data)
				msg = e.channel.Encrypt(framed)
			} else {
				msg = req.data
			}
			err := writeMessage(p.stream, msg)
			req.done <- err
		}

	case CmdDeliverRecv:
		if sd, ok := payload.(*streamData); ok {
			e.deliverRecv(sd.data)
		}

	case CmdDeliverRecvDatagram:
		if dd, ok := payload.(*datagramData); ok {
			e.deliverDatagram(dd.data)
		}

	case CmdSendPathPing:
		p := e.activePath()
		if p.dg != nil {
			p.dg.SendDatagram([]byte{dgConnWhole})
		}

	case CmdSendPathPong:
		p := e.activePath()
		if p.dg != nil {
			p.dg.SendDatagram([]byte{dgConnWhole})
		}

	case CmdSendLanOffer:
		if e.lanServer != nil {
			if err := e.sendLANOffer(); err != nil {
				slog.Warn("send LAN offer failed", "err", err)
			}
		}

	case CmdDialLan:
		// Dial happens asynchronously — post result as event.
		go e.dialLAN()

	case CmdSendLanVerify:
		// Verification is part of the dial flow — handled in dialLAN.

	case CmdSendLanConfirm:
		// Confirmation is sent by the LAN server handler.

	case CmdStartLanStreamReader:
		if e.lan != nil {
			ctx, cancel := context.WithCancel(e.ctx)
			e.lanStreamCancel = cancel
			go e.streamReader(ctx, e.lan, EventLanStreamData, EventLanStreamError)
		}

	case CmdStopLanStreamReader:
		if e.lanStreamCancel != nil {
			e.lanStreamCancel()
			e.lanStreamCancel = nil
		}

	case CmdStartLanDgReader:
		if e.lan != nil {
			ctx, cancel := context.WithCancel(e.ctx)
			e.lanDgCancel = cancel
			go e.datagramReader(ctx, e.lan, EventLanDatagram)
		}

	case CmdStopLanDgReader:
		if e.lanDgCancel != nil {
			e.lanDgCancel()
			e.lanDgCancel = nil
		}

	case CmdStartMonitor:
		ctx, cancel := context.WithCancel(e.ctx)
		e.monitorCancel = cancel
		go e.monitorLoop(ctx)

	case CmdStopMonitor:
		if e.monitorCancel != nil {
			e.monitorCancel()
			e.monitorCancel = nil
		}

	case CmdStartBackoffTimer:
		ctx, cancel := context.WithCancel(e.ctx)
		e.backoffCancel = cancel
		go e.backoffLoop(ctx)

	case CmdCloseLanPath:
		if e.lan != nil {
			e.lan.close()
			e.lan = nil
		}

	case CmdSignalLanReady:
		select {
		case <-e.lanReady:
		default:
			close(e.lanReady)
		}

	case CmdResetLanReady:
		e.lanReady = make(chan struct{})

	case CmdSetCryptoDatagram:
		if e.channel != nil {
			e.channel.SetMode(crypto.ModeDatagrams)
		}
		if e.dgChannel != nil {
			e.dgChannel.SetMode(crypto.ModeDatagrams)
		}

	default:
		slog.Debug("unhandled command", "cmd", cmd)
	}
}

// activePath returns the path that should be used for I/O.
func (e *executor) activePath() *path {
	if e.lan != nil {
		return e.lan
	}
	return e.relay
}

// deliverRecv sends data to the first waiting Recv caller, or buffers
// it if no waiter is registered.
func (e *executor) deliverRecv(data []byte) {
	if len(e.recvWaiters) > 0 {
		w := e.recvWaiters[0]
		e.recvWaiters = e.recvWaiters[1:]
		w <- recvResult{data: data}
	} else {
		e.recvBuffer = append(e.recvBuffer, data)
	}
}

// deliverRecvError sends an error to the first waiting Recv caller.
func (e *executor) deliverRecvError(err error) {
	if len(e.recvWaiters) > 0 {
		w := e.recvWaiters[0]
		e.recvWaiters = e.recvWaiters[1:]
		w <- recvResult{err: err}
	}
}

// deliverDatagram sends data to the first waiting RecvDatagram caller.
func (e *executor) deliverDatagram(data []byte) {
	if len(e.dgRecvWaiters) > 0 {
		w := e.dgRecvWaiters[0]
		e.dgRecvWaiters = e.dgRecvWaiters[1:]
		w <- dgRecvResult{data: data}
	}
}

// --- I/O reader goroutines (zero state logic) ---

// streamReader reads messages from a path's stream and posts events.
func (e *executor) streamReader(ctx context.Context, p *path, dataEvent, errorEvent EventID) {
	for {
		data, err := readMessage(p.stream)
		if err != nil {
			select {
			case <-ctx.Done():
			default:
				e.submit(event{id: errorEvent, payload: &streamError{err: err}})
			}
			return
		}

		// Decrypt if encrypted.
		if e.channel != nil {
			plaintext, err := e.channel.Decrypt(data)
			if err != nil {
				e.submit(event{id: errorEvent, payload: &streamError{err: err}})
				return
			}
			if len(plaintext) == 0 {
				continue
			}
			// Demux control messages vs application data.
			switch plaintext[0] {
			case msgApp:
				e.submit(event{id: dataEvent, payload: &streamData{data: plaintext[1:]}})
			case msgLANOffer:
				e.submit(event{id: EventRecvLanOffer, payload: &streamData{data: plaintext[1:]}})
			case msgCutover:
				slog.Debug("received cutover marker")
			default:
				slog.Warn("discarding unknown message type", "type", plaintext[0])
			}
		} else {
			e.submit(event{id: dataEvent, payload: &streamData{data: data}})
		}
	}
}

// datagramReader reads datagrams from a path and posts events.
func (e *executor) datagramReader(ctx context.Context, p *path, dataEvent EventID) {
	for {
		data, err := p.dg.ReceiveDatagram(ctx)
		if err != nil {
			return
		}
		if len(data) == 0 {
			continue
		}
		// For now, post raw datagram data. The executor handles
		// framing/fragmentation in deliverDatagram.
		e.submit(event{id: dataEvent, payload: &datagramData{data: data}})
	}
}

// monitorLoop sends ping events at fixed intervals.
func (e *executor) monitorLoop(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			e.submit(event{id: EventPingTick})
		}
	}
}

// backoffLoop waits for the backoff delay, then posts the expired event.
func (e *executor) backoffLoop(ctx context.Context) {
	// Read backoff level from the machine. Since the executor owns
	// the machine and this runs after a transition, read the field directly.
	level := 0
	switch m := e.machine.(type) {
	case *BackendMachine:
		level = m.BackoffLevel
	}

	delay := backoffDelay(level)
	slog.Info("backoff timer started", "delay", delay, "level", level)

	select {
	case <-time.After(delay):
		e.submit(event{id: EventBackoffExpired})
	case <-ctx.Done():
	}
}

// backoffDelay computes the delay for a given backoff level with ±25% jitter.
func backoffDelay(level int) time.Duration {
	if level <= 0 {
		return 0
	}
	base := time.Second * time.Duration(math.Pow(2, float64(level-1)))
	jitter := time.Duration(rand.Int64N(int64(base) / 2)) - base/4
	return base + jitter
}

// sendLANOffer registers with the LAN server and sends the offer
// via the relay control channel.
func (e *executor) sendLANOffer() error {
	offer, err := e.lanServer.registerConn(nil) // TODO: pass Conn ref
	if err != nil {
		return err
	}
	_ = offer // TODO: sendControl via relay stream
	return nil
}

// dialLAN attempts to connect to the LAN address from the most recent
// offer. Posts EvLANDialOk or EvLANDialFailed.
func (e *executor) dialLAN() {
	// TODO: extract offer address from event payload, dial QUIC,
	// exchange challenge, post result event.
	e.submit(event{id: EventLanDialFailed})
}

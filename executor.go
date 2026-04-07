// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package pigeon

import (
	"context"
	cryptoRand "crypto/rand"
	"crypto/tls"
	"encoding/binary"
	"encoding/json"
	"io"
	"log/slog"
	"math"
	"math/rand/v2"
	"time"

	"github.com/marcelocantos/pigeon/crypto"
	"github.com/marcelocantos/pigeon/protocol"
	"github.com/quic-go/quic-go"
)

// event carries an event ID plus optional payload into the executor.
type event struct {
	id      EventID
	payload any          // typed per event
	done    chan struct{} // closed after processing (optional, for sync)
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

// chanDgRecvRequest is the payload for channel datagram recv waiters.
type chanDgRecvRequest struct {
	chanID uint16
	result chan dgRecvResult
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

// lanDialResult is the payload for SessionProtocolEventLanDialOk.
type lanDialResult struct {
	stream   io.ReadWriteCloser
	dg       datagrammer
	closer   io.Closer
	opener   streamOpener
	acceptor streamAcceptor
}

// lanOfferData carries the LAN offer received from the backend.
type lanOfferData struct {
	addr      string
	challenge []byte
}

// lanVerifyData carries the verified LAN connection from the server.
type lanVerifyData struct {
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

	// Per-channel datagram routing.
	chanDgWaiters map[uint16][]chan dgRecvResult
	chanDgBuffers map[uint16][][]byte

	// Encryption (executor-level, not machine-level).
	channel   *crypto.Channel
	dgChannel *crypto.Channel

	// LAN server (backend only — for re-advertisement).
	lanServer  *LANServer
	lanEnabled bool
	lanTLS     *tls.Config
	instanceID string

	// Last LAN offer received (client-side, for dial).
	lastOffer *lanOfferData

	// LAN ready signal.
	lanReady chan struct{}

	// Timers managed by the executor.
	monitorCancel  context.CancelFunc
	backoffCancel  context.CancelFunc
	pongCancel     context.CancelFunc // cancelled when pong received

	// Configurable timing (defaults set in newExecutor).
	pingInterval time.Duration // how often to send health pings
	pongTimeout  time.Duration // how long to wait for a pong reply

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
		machine:       machine,
		events:        make(chan event, 64),
		relay:         relay,
		lanReady:      make(chan struct{}),
		chanDgWaiters: make(map[uint16][]chan dgRecvResult),
		chanDgBuffers: make(map[uint16][][]byte),
		reasm:         newReassembler(DefaultFragmentTimeout, done),
		maxDgPayload:  MaxDatagramPayload,
		pingInterval:  time.Duration(PingIntervalMs) * time.Millisecond,
		pongTimeout:   time.Duration(PongTimeoutMs) * time.Millisecond,
		ctx:           ctx,
		cancel:        cancel,
	}

	// Start relay readers — they run for the lifetime of the connection.
	if relay.stream != nil {
		go e.streamReader(ctx, relay, SessionProtocolEventRelayStreamData, SessionProtocolEventRelayStreamError)
	}
	if relay.dg != nil {
		go e.datagramReader(ctx, relay, SessionProtocolEventRelayDatagram)
	}

	// Start the event loop.
	go e.run()

	return e
}

// send submits an app_send event and blocks until the write completes.
func (e *executor) send(ctx context.Context, data []byte) error {
	done := make(chan error, 1)
	e.submit(event{id: SessionProtocolEventAppSend, payload: &sendRequest{data: data, done: done}})
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	case <-e.ctx.Done():
		return e.ctx.Err()
	}
}

// sendDatagram submits an app_send_datagram event. The executor
// handles encryption and fragmentation.
func (e *executor) sendDatagram(data []byte) error {
	done := make(chan error, 1)
	e.submit(event{id: SessionProtocolEventAppSendDatagram, payload: &sendRequest{data: data, done: done}})
	select {
	case err := <-done:
		return err
	case <-e.ctx.Done():
		return e.ctx.Err()
	}
}

// recvDatagram submits a datagram recv waiter and blocks until data arrives.
func (e *executor) recvDatagram(ctx context.Context) ([]byte, error) {
	result := make(chan dgRecvResult, 1)
	e.submit(event{id: SessionProtocolEventAppRecvDatagram, payload: result})
	select {
	case r := <-result:
		return r.data, r.err
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-e.ctx.Done():
		return nil, e.ctx.Err()
	}
}

// recv submits an app_recv waiter and blocks until data arrives.
func (e *executor) recv(ctx context.Context) ([]byte, error) {
	result := make(chan recvResult, 1)
	e.submit(event{id: SessionProtocolEventAppRecv, payload: result})
	select {
	case r := <-result:
		return r.data, r.err
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-e.ctx.Done():
		return nil, e.ctx.Err()
	}
}

// submitSync posts an event and blocks until the event loop has
// processed it. Used by fallbackToRelay to ensure commands execute
// before returning.
func (e *executor) submitSync(ev event) {
	done := make(chan struct{})
	ev.done = done
	e.submit(ev)
	select {
	case <-done:
	case <-e.ctx.Done():
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
			if ev.id == SessionProtocolEventAppRecv {
				if ch, ok := ev.payload.(chan recvResult); ok {
					if len(e.recvBuffer) > 0 {
						ch <- recvResult{data: e.recvBuffer[0]}
						e.recvBuffer = e.recvBuffer[1:]
					} else {
						e.recvWaiters = append(e.recvWaiters, ch)
					}
				}
				if ev.done != nil {
					close(ev.done)
				}
				continue
			}
			if ev.id == SessionProtocolEventAppRecvDatagram {
				switch p := ev.payload.(type) {
				case chan dgRecvResult:
					// Conn-level datagram recv.
					if len(e.dgRecvBuffer) > 0 {
						p <- dgRecvResult{data: e.dgRecvBuffer[0]}
						e.dgRecvBuffer = e.dgRecvBuffer[1:]
					} else {
						e.dgRecvWaiters = append(e.dgRecvWaiters, p)
					}
				case *chanDgRecvRequest:
					// Per-channel datagram recv.
					id := p.chanID
					if buf := e.chanDgBuffers[id]; len(buf) > 0 {
						p.result <- dgRecvResult{data: buf[0]}
						e.chanDgBuffers[id] = buf[1:]
					} else {
						e.chanDgWaiters[id] = append(e.chanDgWaiters[id], p.result)
					}
				}
				if ev.done != nil {
					close(ev.done)
				}
				continue
			}

			// Stash event-specific data before the machine processes it.
			if ev.id == SessionProtocolEventRecvLanOffer {
				if od, ok := ev.payload.(*lanOfferData); ok {
					e.lastOffer = od
				}
			}
			if ev.id == SessionProtocolEventLanDialOk {
				if ld, ok := ev.payload.(*lanDialResult); ok {
					e.lan = newPath("lan", ld.stream, ld.dg, ld.closer, ld.opener, ld.acceptor)
				}
			}
			if ev.id == SessionProtocolEventRecvLanVerify {
				if vd, ok := ev.payload.(*lanVerifyData); ok {
					// Backend: LAN server verified the client's challenge.
					// Install the LAN path before the machine transitions.
					e.lan = newPath("lan", vd.stream, vd.dg, vd.closer, vd.opener, vd.acceptor)
				}
			}

			cmds, err := e.machine.HandleEvent(ev.id)
			if err != nil {
				slog.Warn("machine error", "event", ev.id, "err", err)
				continue
			}
			for _, cmd := range cmds {
				e.executeCommand(cmd, ev.payload)
			}

			if cmds == nil {
				slog.Debug("unmatched event (no transition)", "event", ev.id)
			}

			// Signal synchronous waiters.
			if ev.done != nil {
				close(ev.done)
			}

		case <-e.ctx.Done():
			return
		}
	}
}

// executeCommand carries out a single command from the machine.
func (e *executor) executeCommand(cmd CmdID, payload any) {
	switch cmd {
	case SessionProtocolCmdWriteActiveStream:
		if req, ok := payload.(*sendRequest); ok {
			p := e.activePath()
			var msg []byte
			if e.channel != nil {
				framed := make([]byte, 1+len(req.data))
				framed[0] = FrameApp
				copy(framed[1:], req.data)
				msg = e.channel.Encrypt(framed)
			} else {
				msg = req.data
			}
			err := writeMessage(p.stream, msg)
			req.done <- err
		}

	case SessionProtocolCmdSendActiveDatagram:
		if req, ok := payload.(*sendRequest); ok {
			p := e.activePath()
			data := req.data
			if e.dgChannel != nil {
				data = e.dgChannel.Encrypt(data)
			}
			// Single datagram or fragment.
			if 1+len(data) <= e.maxDgPayload {
				frame := make([]byte, 1+len(data))
				frame[0] = DgConnWhole
				copy(frame[1:], data)
				req.done <- p.dg.SendDatagram(frame)
			} else {
				msgID := nextMsgID.Add(1)
				req.done <- sendFragmented(p.dg, data, e.maxDgPayload, msgID, DgConnFragment, nil)
			}
		}

	case SessionProtocolCmdDeliverRecv:
		if sd, ok := payload.(*streamData); ok {
			e.deliverRecv(sd.data)
		}

	case SessionProtocolCmdDeliverRecvError:
		if se, ok := payload.(*streamError); ok {
			e.deliverRecvError(se.err)
		}

	case SessionProtocolCmdDeliverRecvDatagram:
		if dd, ok := payload.(*datagramData); ok {
			e.processDatagram(dd.data)
		}

	case SessionProtocolCmdSendPathPing:
		p := e.activePath()
		if p.dg != nil {
			p.dg.SendDatagram([]byte{DgPing})
		}

	case SessionProtocolCmdSendPathPong:
		p := e.activePath()
		if p.dg != nil {
			p.dg.SendDatagram([]byte{DgPong})
		}

	case SessionProtocolCmdSendLanOffer:
		if e.lanServer != nil {
			if err := e.sendLANOffer(); err != nil {
				slog.Warn("send LAN offer failed", "err", err)
			}
		}

	case SessionProtocolCmdDialLan:
		// Dial happens asynchronously — post result as event.
		go e.dialLAN()

	case SessionProtocolCmdSendLanVerify:
		// Verification is part of the dial flow — handled in dialLAN.

	case SessionProtocolCmdSendLanConfirm:
		// Confirmation is sent by the LAN server handler.

	case SessionProtocolCmdStartLanStreamReader:
		if e.lan != nil {
			ctx, cancel := context.WithCancel(e.ctx)
			e.lanStreamCancel = cancel
			go e.streamReader(ctx, e.lan, SessionProtocolEventLanStreamData, SessionProtocolEventLanStreamError)
		}

	case SessionProtocolCmdStopLanStreamReader:
		if e.lanStreamCancel != nil {
			e.lanStreamCancel()
			e.lanStreamCancel = nil
		}

	case SessionProtocolCmdStartLanDgReader:
		if e.lan != nil {
			ctx, cancel := context.WithCancel(e.ctx)
			e.lanDgCancel = cancel
			go e.datagramReader(ctx, e.lan, SessionProtocolEventLanDatagram)
		}

	case SessionProtocolCmdStopLanDgReader:
		if e.lanDgCancel != nil {
			e.lanDgCancel()
			e.lanDgCancel = nil
		}

	case SessionProtocolCmdStartMonitor:
		ctx, cancel := context.WithCancel(e.ctx)
		e.monitorCancel = cancel
		go e.monitorLoop(ctx)

	case SessionProtocolCmdStopMonitor:
		if e.monitorCancel != nil {
			e.monitorCancel()
			e.monitorCancel = nil
		}

	case SessionProtocolCmdStartPongTimeout:
		if e.pongCancel != nil {
			e.pongCancel()
		}
		ctx, cancel := context.WithCancel(e.ctx)
		e.pongCancel = cancel
		go func() {
			select {
			case <-time.After(e.pongTimeout):
				e.submit(event{id: SessionProtocolEventPingTimeout})
			case <-ctx.Done():
			}
		}()

	case SessionProtocolCmdCancelPongTimeout:
		if e.pongCancel != nil {
			e.pongCancel()
			e.pongCancel = nil
		}

	case SessionProtocolCmdStartBackoffTimer:
		ctx, cancel := context.WithCancel(e.ctx)
		e.backoffCancel = cancel
		go e.backoffLoop(ctx)

	case SessionProtocolCmdCloseLanPath:
		if e.lan != nil {
			e.lan.close()
			e.lan = nil
		}
		// Drain stale datagram buffers from the old path.
		e.dgRecvBuffer = nil
		for id := range e.chanDgBuffers {
			e.chanDgBuffers[id] = nil
		}

	case SessionProtocolCmdSignalLanReady:
		select {
		case <-e.lanReady:
		default:
			close(e.lanReady)
		}

	case SessionProtocolCmdResetLanReady:
		e.lanReady = make(chan struct{})

	case SessionProtocolCmdSetCryptoDatagram:
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

// processDatagram decodes framing, reassembles fragments, decrypts,
// and delivers the payload to the application.
func (e *executor) processDatagram(raw []byte) {
	if len(raw) == 0 {
		return
	}

	switch raw[0] {
	case DgPing:
		e.submit(event{id: SessionProtocolEventRecvPathPing})
		return
	case DgPong:
		e.submit(event{id: SessionProtocolEventRecvPathPong})
		return
	case DgConnWhole:
		payload := raw[1:]
		if e.dgChannel != nil {
			decrypted, err := e.dgChannel.Decrypt(payload)
			if err != nil {
				slog.Debug("datagram decrypt failed", "err", err)
				return
			}
			payload = decrypted
		}
		e.deliverDatagram(payload)

	case DgConnFragment:
		if len(raw) < 1+FragHeaderSize {
			return
		}
		msgID := binary.BigEndian.Uint32(raw[1:5])
		fragIdx := int(binary.BigEndian.Uint16(raw[5:7]))
		totalFrags := int(binary.BigEndian.Uint16(raw[7:9]))
		chunk := raw[1+FragHeaderSize:]
		if totalFrags < 2 || fragIdx >= totalFrags {
			return
		}
		assembled := e.reasm.feed(msgID, fragIdx, totalFrags, chunk)
		if assembled == nil {
			return
		}
		if e.dgChannel != nil {
			decrypted, err := e.dgChannel.Decrypt(assembled)
			if err != nil {
				slog.Debug("fragment decrypt failed", "err", err)
				return
			}
			assembled = decrypted
		}
		e.deliverDatagram(assembled)

	case DgChanWhole:
		if len(raw) < 1+ChanIdSize {
			return
		}
		id := binary.BigEndian.Uint16(raw[1:3])
		payload := raw[1+ChanIdSize:]
		e.deliverChannelDatagram(id, payload)

	case DgChanFragment:
		if len(raw) < 1+ChanIdSize+FragHeaderSize {
			return
		}
		id := binary.BigEndian.Uint16(raw[1:3])
		off := 1 + ChanIdSize
		msgID := binary.BigEndian.Uint32(raw[off : off+4])
		fragIdx := int(binary.BigEndian.Uint16(raw[off+4 : off+6]))
		totalFrags := int(binary.BigEndian.Uint16(raw[off+6 : off+8]))
		chunk := raw[off+FragHeaderSize:]
		if totalFrags < 2 || fragIdx >= totalFrags {
			return
		}
		assembled := e.reasm.feed(msgID, fragIdx, totalFrags, chunk)
		if assembled == nil {
			return
		}
		e.deliverChannelDatagram(id, assembled)

	default:
		// Unknown frame type — discard.
	}
}

// deliverDatagram sends data to the first waiting RecvDatagram caller,
// or buffers it if no waiter is registered.
func (e *executor) deliverDatagram(data []byte) {
	if len(e.dgRecvWaiters) > 0 {
		w := e.dgRecvWaiters[0]
		e.dgRecvWaiters = e.dgRecvWaiters[1:]
		w <- dgRecvResult{data: data}
	} else {
		e.dgRecvBuffer = append(e.dgRecvBuffer, data)
	}
}

// deliverChannelDatagram sends data to a waiting channel datagram
// receiver, or buffers it if no waiter is registered.
func (e *executor) deliverChannelDatagram(id uint16, data []byte) {
	if ws := e.chanDgWaiters[id]; len(ws) > 0 {
		w := ws[0]
		e.chanDgWaiters[id] = ws[1:]
		w <- dgRecvResult{data: data}
	} else {
		e.chanDgBuffers[id] = append(e.chanDgBuffers[id], data)
	}
}

// recvChannelDatagram registers a waiter for the next datagram on the
// given channel ID. Delivers immediately from the buffer if available.
func (e *executor) recvChannelDatagram(ctx context.Context, id uint16) ([]byte, error) {
	result := make(chan dgRecvResult, 1)
	e.submit(event{id: SessionProtocolEventAppRecvDatagram, payload: &chanDgRecvRequest{chanID: id, result: result}})
	select {
	case r := <-result:
		return r.data, r.err
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-e.ctx.Done():
		return nil, e.ctx.Err()
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
				e.submit(event{id: dataEvent, payload: &streamData{data: nil}})
				continue
			}
			// Demux control messages vs application data.
			switch plaintext[0] {
			case FrameApp:
				e.submit(event{id: dataEvent, payload: &streamData{data: plaintext[1:]}})
			case FrameLanOffer:
				var offer lanOffer
				if err := json.Unmarshal(plaintext[1:], &offer); err != nil {
					slog.Warn("bad LAN offer", "err", err)
					continue
				}
				e.submit(event{id: SessionProtocolEventRecvLanOffer, payload: &lanOfferData{
					addr: offer.Addr, challenge: offer.Challenge,
				}})
			case FrameCutover:
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
	ticker := time.NewTicker(e.pingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			e.submit(event{id: SessionProtocolEventPingTick})
		}
	}
}

// backoffLoop waits for the backoff delay, then posts the expired event.
func (e *executor) backoffLoop(ctx context.Context) {
	// Read backoff level from the machine. Since the executor owns
	// the machine and this runs after a transition, read the field directly.
	level := 0
	switch m := e.machine.(type) {
	case *SessionProtocolBackendMachine:
		level = m.BackoffLevel
	}

	delay := backoffDelay(level)
	slog.Info("backoff timer started", "delay", delay, "level", level)

	select {
	case <-time.After(delay):
		e.submit(event{id: SessionProtocolEventBackoffExpired})
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
	challenge := make([]byte, 32)
	if _, err := cryptoRand.Read(challenge); err != nil {
		return err
	}

	// Register with the LAN server. When the client verifies, the
	// callback posts SessionProtocolEventRecvLanVerify to the executor.
	e.lanServer.mu.Lock()
	e.lanServer.conns[e.instanceID] = &pendingLAN{
		challenge: challenge,
		onVerify: func(stream io.ReadWriteCloser, conn *quic.Conn) {
			e.submit(event{id: SessionProtocolEventRecvLanVerify, payload: &lanVerifyData{
				stream:   stream,
				dg:       conn,
				closer:   quicCloser{conn},
				opener:   quicOpener{conn},
				acceptor: quicAcceptor{conn},
			}})
		},
	}
	e.lanServer.mu.Unlock()

	offer := lanOffer{
		Addr:      e.lanServer.addr,
		Challenge: challenge,
	}

	// Send the offer as a control message via the relay stream.
	return e.sendControl(FrameLanOffer, offer)
}

// sendControl writes a framed control message on the relay stream.
func (e *executor) sendControl(msgType byte, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	framed := make([]byte, 1+len(data))
	framed[0] = msgType
	copy(framed[1:], data)

	if e.channel != nil {
		framed = e.channel.Encrypt(framed)
	}

	return writeMessage(e.relay.stream, framed)
}

// dialLAN attempts to connect to the LAN address from the most recent
// offer. Posts SessionProtocolEventLanDialOk or SessionProtocolEventLanDialFailed.
func (e *executor) dialLAN() {
	offer := e.lastOffer
	if offer == nil {
		e.submit(event{id: SessionProtocolEventLanDialFailed})
		return
	}

	tlsConfig := e.lanTLS
	if tlsConfig == nil {
		tlsConfig = &tls.Config{
			InsecureSkipVerify: true,
			NextProtos:         []string{"pigeon-lan"},
		}
	} else {
		tlsConfig = tlsConfig.Clone()
		tlsConfig.NextProtos = []string{"pigeon-lan"}
	}

	ctx, cancel := context.WithTimeout(e.ctx, 5*time.Second)
	defer cancel()

	conn, err := quic.DialAddr(ctx, offer.addr, tlsConfig, &quic.Config{
		EnableDatagrams: true,
		MaxIdleTimeout:  60 * time.Second,
		KeepAlivePeriod: 10 * time.Second,
	})
	if err != nil {
		slog.Debug("LAN dial failed", "addr", offer.addr, "err", err)
		e.submit(event{id: SessionProtocolEventLanDialFailed})
		return
	}

	stream, err := conn.OpenStream()
	if err != nil {
		conn.CloseWithError(1, "open stream failed")
		e.submit(event{id: SessionProtocolEventLanDialFailed})
		return
	}

	// Send verification.
	verify := lanVerify{
		Challenge:  offer.challenge,
		InstanceID: e.instanceID,
	}
	data, _ := json.Marshal(verify)
	if err := writeMessage(stream, data); err != nil {
		conn.CloseWithError(1, "write verify failed")
		e.submit(event{id: SessionProtocolEventLanDialFailed})
		return
	}

	// Wait for confirmation.
	resp, err := readMessage(stream)
	if err != nil || string(resp) != "ok" {
		conn.CloseWithError(1, "verify rejected")
		e.submit(event{id: SessionProtocolEventLanDialFailed})
		return
	}

	slog.Info("LAN connection established", "addr", offer.addr)
	result := &lanDialResult{
		stream:   stream,
		dg:       conn,
		closer:   quicCloser{conn},
		opener:   quicOpener{conn},
		acceptor: quicAcceptor{conn},
	}
	// Post both events: dial succeeded, then confirm received.
	// The dialLAN flow does the full handshake (send verify, recv ok)
	// so both transitions fire in sequence.
	e.submit(event{id: SessionProtocolEventLanDialOk, payload: result})
	e.submit(event{id: SessionProtocolEventRecvLanConfirm, payload: result})
}


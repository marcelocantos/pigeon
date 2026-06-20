// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package pigeon

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"time"
	"unsafe"

	"github.com/marcelocantos/pigeon/crypto"
	"github.com/marcelocantos/pigeon/cwire"
)

const (
	// listenPoolSize is how many listen connections the backend keeps
	// parked at the relay (see docs/DESIGN.md §3 L1). Each is an idle
	// QUIC connection awaiting a client; the pool bounds how many
	// clients can be mid-handshake concurrently. Activation is
	// sub-second, so a small pool serves typical connect rates without
	// making clients wait — when all slots are busy, the relay simply
	// blocks the next client until a worker frees up and re-parks.
	listenPoolSize = 4

	// listenRetryBackoff paces re-dials after a transient listen dial
	// failure (relay briefly unavailable) so a flapping relay doesn't
	// spin the pool.
	listenRetryBackoff = 500 * time.Millisecond
)

// RegisterArgs configures a backend's listener.
type RegisterArgs struct {
	// Identity is the backend's long-term keypair. The InstanceID this
	// backend advertises is derived from Identity.PublicKey().
	Identity crypto.Identity

	// Pairing resolves an incoming client's stable identifier to its
	// PairingRecord. Returning (nil, err) rejects the connection.
	Pairing func(clientID string) (*crypto.PairingRecord, error)

	// Relay is the relay URL (e.g. "https://relay.example.com").
	Relay string

	// TLS overrides the default TLS configuration. nil ⇒ insecure-skip-verify
	// (development default; production should pin or trust a CA).
	TLS *tls.Config

	// Token is an optional bearer token presented to the relay.
	Token string

	// Datagrams declares the named datagram channels available on each
	// accepted Session, mapped to their pre-agreed varint channel IDs.
	// Both peers must declare the same map. Channel ID 0 is reserved.
	Datagrams map[string]uint64
}

// ConnectArgs configures a client-side connection to a paired backend.
type ConnectArgs struct {
	// InstanceID is the backend's stable identifier (carried in
	// PairingRecord.PeerInstanceID after pairing).
	InstanceID string

	// Record is the PairingRecord persisted from the pairing ceremony.
	Record *crypto.PairingRecord

	// Identity is the client's long-term keypair.
	Identity crypto.Identity

	// Relay is the relay URL.
	Relay string

	// TLS overrides the default TLS configuration.
	TLS *tls.Config

	// Token is an optional bearer token presented to the relay.
	Token string

	// Datagrams declares the named datagram channels (see RegisterArgs.Datagrams).
	Datagrams map[string]uint64
}

// Listener is the backend-side acceptor for paired clients. Under the
// remote-Listen L1 model (docs/DESIGN.md §3), it holds one register
// control connection open for the instance's lifetime and keeps a pool
// of listen connections parked at the relay. The relay matches each
// arriving client to a parked listen and bridges the two QUIC
// connections end-to-end, so every accepted Session rides its own pipe —
// there is no shared connection or per-client tag demux.
type Listener struct {
	args       *RegisterArgs
	control    *transport
	instanceID string
	relay      string
	cfg        Config

	ctx    context.Context
	cancel context.CancelFunc

	mu     sync.Mutex
	parked map[*transport]struct{} // listen conns not yet owned by a Session

	accepted chan acceptResult

	closeOnce sync.Once
}

type acceptResult struct {
	session *Session
	err     error
}

// Register publishes a backend on the relay and returns a Listener that
// yields one Session per accepted client. Two modes, distinguished by
// whether args.Pairing is supplied:
//
//   - **Activation mode** (args.Pairing != nil): each accepted client
//     runs the auth_request / auth_ok handshake against args.Pairing
//     to resolve its PairingRecord, and the returned Session carries
//     an AEAD channel derived from that record. This is the normal
//     post-pairing operating mode.
//   - **Pairing mode** (args.Pairing == nil): activation is skipped.
//     Each accepted client is handed back as a raw Session with no
//     AEAD channel — pairing.go uses this mode to run the pairing
//     ceremony, which has its own crypto and emits a fresh
//     PairingRecord on completion. Per docs/DESIGN.md §3 L2, this is
//     the no-PairingRecord entry condition that pairing-mode selects.
func Register(ctx context.Context, args *RegisterArgs) (*Listener, string, error) {
	if args == nil {
		return nil, "", fmt.Errorf("pigeon.Register: args is nil")
	}
	if args.Relay == "" {
		return nil, "", fmt.Errorf("pigeon.Register: Relay is required")
	}
	pairingMode := args.Pairing == nil
	if !pairingMode && args.Identity == nil {
		return nil, "", fmt.Errorf("pigeon.Register: Identity is required in activation mode (Pairing supplied)")
	}
	if err := validateDatagrams(args.Datagrams); err != nil {
		return nil, "", fmt.Errorf("pigeon.Register: %w", err)
	}

	tlsCfg := args.TLS
	if tlsCfg == nil {
		tlsCfg = &tls.Config{InsecureSkipVerify: true}
	}
	cfg := Config{
		TLS:   tlsCfg,
		Token: args.Token,
	}
	// In activation mode, the backend's stable instance ID identifies
	// the long-lived listener. In pairing mode, each ceremony wants a
	// fresh random ID that goes into the QR token — let the relay
	// assign one.
	if !pairingMode {
		cfg.InstanceID = args.Identity.InstanceID()
	}
	control, err := dialRegister(ctx, args.Relay, cfg)
	if err != nil {
		return nil, "", fmt.Errorf("relay register: %w", err)
	}
	// Subsequent listen dials reuse the assigned ID, not the requested
	// one (which is empty in pairing mode).
	cfg.InstanceID = control.instanceID

	lctx, cancel := context.WithCancel(ctx)
	l := &Listener{
		args:       args,
		control:    control,
		instanceID: control.instanceID,
		relay:      args.Relay,
		cfg:        cfg,
		ctx:        lctx,
		cancel:     cancel,
		parked:     make(map[*transport]struct{}),
		accepted:   make(chan acceptResult, 8),
	}
	for range listenPoolSize {
		go l.acceptWorker()
	}
	return l, control.instanceID, nil
}

// ID returns the stable instance ID this Listener advertises.
func (l *Listener) ID() string { return l.instanceID }

// Accept blocks until the next paired client connects.
func (l *Listener) Accept(ctx context.Context) (*Session, error) {
	select {
	case r, ok := <-l.accepted:
		if !ok {
			return nil, io.ErrClosedPipe
		}
		return r.session, r.err
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-l.ctx.Done():
		return nil, io.ErrClosedPipe
	}
}

// Close releases the Listener: it tears down the register control
// connection and every idle listen connection still parked at the relay.
// Already-accepted Sessions are independently owned and are not affected
// — close them via Session.Close.
func (l *Listener) Close() error {
	l.closeOnce.Do(func() {
		l.cancel()
		_ = l.control.Close()
		l.mu.Lock()
		for tr := range l.parked {
			_ = tr.Close()
		}
		l.parked = nil
		l.mu.Unlock()
	})
	return nil
}

// acceptWorker maintains one slot in the listen pool: it dials a listen
// connection, parks it at the relay, runs the L2 handshake once a client
// is bridged onto it, hands the resulting Session to Accept, then loops
// to re-park a fresh listen. listenPoolSize of these run concurrently.
func (l *Listener) acceptWorker() {
	for l.ctx.Err() == nil {
		tr, err := dialListen(l.ctx, l.relay, l.instanceID, l.cfg)
		if err != nil {
			if l.ctx.Err() != nil {
				return
			}
			slog.Debug("listener: dial listen", "err", err)
			select {
			case <-time.After(listenRetryBackoff):
			case <-l.ctx.Done():
				return
			}
			continue
		}
		l.trackListen(tr)
		sess, err := l.activate(tr)
		if err != nil {
			l.untrackListen(tr)
			_ = tr.Close()
			if l.ctx.Err() != nil {
				return
			}
			slog.Warn("listener: activation", "err", err)
			select {
			case l.accepted <- acceptResult{err: err}:
			case <-l.ctx.Done():
				return
			}
			continue
		}
		l.untrackListen(tr) // Session owns tr from here.
		select {
		case l.accepted <- acceptResult{session: sess}:
		case <-l.ctx.Done():
			_ = sess.Close()
			return
		}
	}
}

func (l *Listener) trackListen(tr *transport) {
	l.mu.Lock()
	if l.parked != nil {
		l.parked[tr] = struct{}{}
	}
	l.mu.Unlock()
}

func (l *Listener) untrackListen(tr *transport) {
	l.mu.Lock()
	delete(l.parked, tr)
	l.mu.Unlock()
}

// activate runs the L2 handshake on a freshly-bridged listen connection
// and builds the resulting Session. In activation mode (l.args.Pairing
// != nil) it drives the auth_request / auth_ok exchange in libpigeon via
// cwire against args.Pairing's lookup, deriving an AEAD channel from the
// matched record. In pairing mode (l.args.Pairing == nil) the handshake
// is skipped and the bridged primary is wrapped as a no-AEAD Session for
// the pairing ceremony to use directly. See docs/DESIGN.md §3 L2.
func (l *Listener) activate(tr *transport) (*Session, error) {
	if l.args.Pairing == nil {
		// Pairing mode has no activation handshake to block on, so wait
		// for the client's arrival marker on the bridged primary before
		// producing a Session — otherwise a worker would hand back a
		// Session for a listen the relay hasn't matched to any client
		// yet. (Activation mode gets this gating for free: its first
		// primary read is the client's auth_request.) See Connect.
		if _, err := readMessage(tr.primary); err != nil {
			return nil, fmt.Errorf("await client: %w", err)
		}
		sess := newSession(l.ctx, tr, nil, "", l.args.Datagrams, true)
		sess.ownsTransport = true
		sess.bindPrimary(tr.primary)
		go sess.acceptLoop()
		go sess.datagramLoop()
		return sess, nil
	}

	// Each accepted client has its own end-to-end QUIC connection, so —
	// exactly like the client side in Connect — activation runs over a
	// per-connection cwire adapter/ref that loops back to tr.primary
	// through writeMessage/readMessage.
	adapter := newGoTransportAdapter(l.ctx, tr)
	ref := cwire.NewGoTransportRef(adapter)
	cwirePrimary := adapter.adoptPrimary(tr.primary)

	var rec *crypto.PairingRecord
	result, err := cwire.RunBackendActivation(&cwire.RunBackendActivationArgs{
		Ref:    ref,
		Stream: cwirePrimary,
		// Translate the public Pairing func ([id] → *crypto.PairingRecord, error)
		// to the cwire resolver shape ([id] → *cwire.PairingRecord, bool) and
		// capture the matched *crypto.PairingRecord for Go-side channel
		// derivation below.
		ResolveFn: func(deviceID string) (*cwire.PairingRecord, bool) {
			r, err := l.args.Pairing(deviceID)
			if err != nil || r == nil {
				return nil, false
			}
			rec = r
			return pairingRecordToCwire(r), true
		},
	})
	if err != nil {
		removeStream(cwirePrimary)
		ref.Close()
		return nil, fmt.Errorf("activation: %w", err)
	}
	if !result.Accepted {
		removeStream(cwirePrimary)
		ref.Close()
		return nil, fmt.Errorf("activation rejected device %q", result.DeviceID)
	}
	if rec == nil {
		removeStream(cwirePrimary)
		ref.Close()
		return nil, errors.New("activation succeeded but no record captured")
	}
	channel, err := cwire.DeriveSessionChannel(pairingRecordToCwire(rec), true)
	if err != nil {
		removeStream(cwirePrimary)
		ref.Close()
		return nil, fmt.Errorf("derive session channel: %w", err)
	}

	sess := newSession(l.ctx, tr, channel, result.DeviceID, l.args.Datagrams, true)
	sess.ownsTransport = true
	sess.cwireRef = ref
	sess.cwirePrimary = cwirePrimary
	sess.bindPrimary(tr.primary)
	go sess.acceptLoop()
	go sess.datagramLoop()
	return sess, nil
}

// Connect dials the backend identified by args.InstanceID and returns a
// Session ready for OpenStream / Datagram. Two modes, distinguished by
// whether args.Record is supplied:
//
//   - **Activation mode** (args.Record != nil): the client identifies
//     itself to the backend via the auth_request / auth_ok exchange
//     and the resulting Session carries an AEAD channel derived from
//     the supplied PairingRecord.
//   - **Pairing mode** (args.Record == nil, args.Identity == nil):
//     activation is skipped and the Session is returned with no AEAD
//     channel — the pairing ceremony uses this mode to obtain a raw
//     bidirectional bytes pipe to the acceptor and runs its own
//     crypto over it. Per docs/DESIGN.md §3 L2.
func Connect(ctx context.Context, args *ConnectArgs) (*Session, error) {
	if args == nil {
		return nil, fmt.Errorf("pigeon.Connect: args is nil")
	}
	if args.InstanceID == "" {
		return nil, fmt.Errorf("pigeon.Connect: InstanceID is required")
	}
	if args.Relay == "" {
		return nil, fmt.Errorf("pigeon.Connect: Relay is required")
	}
	pairingMode := args.Record == nil
	if !pairingMode && args.Identity == nil {
		return nil, fmt.Errorf("pigeon.Connect: Identity is required in activation mode (Record supplied)")
	}
	if err := validateDatagrams(args.Datagrams); err != nil {
		return nil, fmt.Errorf("pigeon.Connect: %w", err)
	}

	tlsCfg := args.TLS
	if tlsCfg == nil {
		tlsCfg = &tls.Config{InsecureSkipVerify: true}
	}
	cfg := Config{TLS: tlsCfg, Token: args.Token}
	tr, err := dialInitiator(ctx, args.Relay, args.InstanceID, cfg)
	if err != nil {
		return nil, fmt.Errorf("relay connect: %w", err)
	}

	// The dial's "ok" ack means the relay has bridged this connection
	// end-to-end onto a backend listen, so tr.primary is now a clean
	// pipe straight to the backend — no relay-level framing. Activation
	// (or, in pairing mode, the ceremony) is the first thing on it.
	if pairingMode {
		// Pairing mode has no activation handshake; send an empty arrival
		// marker so the backend's accept worker knows a client is bridged
		// onto its listen and produces the Session (see Listener.activate).
		if err := writeMessage(tr.primary, nil); err != nil {
			_ = tr.Close()
			return nil, fmt.Errorf("relay connect: ready: %w", err)
		}
	}
	var (
		channel      *cwire.Channel
		cwireRef     *cwire.GoTransportRef
		cwirePrimary unsafe.Pointer
	)
	if !pairingMode {
		// Activation runs in libpigeon via cwire — it speaks the auth_request /
		// auth_ok wire exchange over the adapter, which loops back to tr.primary
		// through writeMessage/readMessage.
		adapter := newGoTransportAdapter(ctx, tr)
		cwireRef = cwire.NewGoTransportRef(adapter)
		cwirePrimary = adapter.adoptPrimary(tr.primary)
		if err := cwire.RunClientActivation(cwireRef, cwirePrimary, args.Identity.InstanceID()); err != nil {
			removeStream(cwirePrimary)
			cwireRef.Close()
			_ = tr.Close()
			return nil, fmt.Errorf("client activation: %w", err)
		}
		channel, err = cwire.DeriveSessionChannel(pairingRecordToCwire(args.Record), false)
		if err != nil {
			removeStream(cwirePrimary)
			cwireRef.Close()
			_ = tr.Close()
			return nil, fmt.Errorf("derive session channel: %w", err)
		}
	}

	sess := newSession(ctx, tr, channel, args.InstanceID, args.Datagrams, false)
	sess.ownsTransport = true
	sess.cwireRef = cwireRef
	sess.cwirePrimary = cwirePrimary
	sess.bindPrimary(tr.primary)
	go sess.acceptLoop()
	go sess.datagramLoop()
	return sess, nil
}

// validateDatagrams checks the Datagrams map: no duplicate channel IDs,
// channel ID 0 reserved.
func validateDatagrams(m map[string]uint64) error {
	seen := make(map[uint64]string, len(m))
	for name, id := range m {
		if id == 0 {
			return fmt.Errorf("datagram channel %q: id 0 is reserved", name)
		}
		if other, dup := seen[id]; dup {
			return fmt.Errorf("datagram channel id %d used by both %q and %q", id, other, name)
		}
		seen[id] = name
	}
	return nil
}

// Session is an end-to-end encrypted session between two paired peers,
// scoped to a single client. It carries one or more named, reliable
// stream channels (OpenStream) and pre-agreed datagram channels (Datagram).
//
// Each Session owns its own end-to-end QUIC connection to the peer
// (bridged opaquely by the relay), so both sides run a symmetric
// sub-stream accept loop and datagram pump over that connection — there
// is no relay-assigned client tag or shared-connection demux.
type Session struct {
	ctx    context.Context
	cancel context.CancelFunc

	transport     *transport
	channel       *cwire.Channel
	peerID        string
	isBackend     bool
	ownsTransport bool

	primary io.ReadWriteCloser

	// cwireRef pins the goTransportAdapter against the cgo handle table
	// for the lifetime of the session. Owned per-connection on both
	// sides (activation mode); nil on pairing-mode sessions.
	cwireRef *cwire.GoTransportRef

	// cwirePrimary is the opaque stream handle adopted by the adapter
	// during activation. Released when the session closes.
	cwirePrimary unsafe.Pointer

	// Stream rendezvous: pendingOpens map name → waiter; bufferedStreams
	// holds streams that arrived before a local AcceptStream caller.
	streamMu        sync.Mutex
	pendingOpens    map[string][]chan io.ReadWriteCloser
	bufferedStreams map[string][]io.ReadWriteCloser

	// Datagrams are pre-declared by name → channel-id at construction.
	dgConfig  map[string]uint64
	datagrams map[uint64]*Datagram

	closeOnce sync.Once
}

func newSession(ctx context.Context, tr *transport, channel *cwire.Channel, peerID string, dgConfig map[string]uint64, isBackend bool) *Session {
	sCtx, cancel := context.WithCancel(ctx)
	s := &Session{
		ctx:             sCtx,
		cancel:          cancel,
		transport:       tr,
		channel:         channel,
		peerID:          peerID,
		isBackend:       isBackend,
		pendingOpens:    make(map[string][]chan io.ReadWriteCloser),
		bufferedStreams: make(map[string][]io.ReadWriteCloser),
		dgConfig:        make(map[string]uint64, len(dgConfig)),
		datagrams:       make(map[uint64]*Datagram, len(dgConfig)),
	}
	for name, id := range dgConfig {
		s.dgConfig[name] = id
		s.datagrams[id] = &Datagram{
			name:    name,
			id:      id,
			session: s,
			rx:      make(chan []byte, 32),
		}
	}
	return s
}

func (s *Session) bindPrimary(stream io.ReadWriteCloser) {
	s.primary = stream
}

// PeerID returns the InstanceID of the remote peer.
func (s *Session) PeerID() string { return s.peerID }

// OpenStream opens a fresh, reliable, ordered, message-framed channel
// identified by `name`. Both peers must call OpenStream/AcceptStream with
// the same name; the matching call on the peer unblocks when the name
// binds. The stream is a new QUIC stream on this session's connection;
// the relay forwards it opaquely to the peer.
func (s *Session) OpenStream(ctx context.Context, name string) (*Stream, error) {
	if name == "" {
		return nil, errors.New("OpenStream: name must be non-empty")
	}
	rwc, err := s.transport.OpenStream()
	if err != nil {
		return nil, fmt.Errorf("open quic stream: %w", err)
	}
	if err := writeMessage(rwc, EncodeStreamHeader(name)); err != nil {
		_ = rwc.Close()
		return nil, fmt.Errorf("write stream header: %w", err)
	}
	return &Stream{name: name, rwc: rwc, channel: s.channel}, nil
}

// AcceptStream blocks until the peer opens a stream with the given name.
// Streams that arrive before the local AcceptStream call are buffered
// and delivered when the matching name is later requested.
func (s *Session) AcceptStream(ctx context.Context, name string) (*Stream, error) {
	if name == "" {
		return nil, errors.New("AcceptStream: name must be non-empty")
	}
	s.streamMu.Lock()
	if buf := s.bufferedStreams[name]; len(buf) > 0 {
		rwc := buf[0]
		s.bufferedStreams[name] = buf[1:]
		s.streamMu.Unlock()
		return &Stream{name: name, rwc: rwc, channel: s.channel}, nil
	}
	ch := make(chan io.ReadWriteCloser, 1)
	s.pendingOpens[name] = append(s.pendingOpens[name], ch)
	s.streamMu.Unlock()

	select {
	case rwc := <-ch:
		return &Stream{name: name, rwc: rwc, channel: s.channel}, nil
	case <-ctx.Done():
		s.removePending(name, ch)
		return nil, ctx.Err()
	case <-s.ctx.Done():
		return nil, io.ErrClosedPipe
	}
}

func (s *Session) removePending(name string, ch chan io.ReadWriteCloser) {
	s.streamMu.Lock()
	defer s.streamMu.Unlock()
	q := s.pendingOpens[name]
	for i, c := range q {
		if c == ch {
			s.pendingOpens[name] = append(q[:i], q[i+1:]...)
			return
		}
	}
}

// deliverIncomingStream is called by acceptLoop when a new peer-opened
// stream's name handshake has been read.
func (s *Session) deliverIncomingStream(name string, rwc io.ReadWriteCloser) {
	s.streamMu.Lock()
	if q := s.pendingOpens[name]; len(q) > 0 {
		ch := q[0]
		s.pendingOpens[name] = q[1:]
		s.streamMu.Unlock()
		ch <- rwc
		return
	}
	s.bufferedStreams[name] = append(s.bufferedStreams[name], rwc)
	s.streamMu.Unlock()
}

// acceptLoop reads each sub-stream the peer opens on this session's
// connection, reads its name header, and delivers it by name. Runs on
// both sides — each session owns its own connection, so there is no
// shared-connection demux. The primary stream is consumed by activation
// (or used directly in pairing mode) and is never delivered here.
func (s *Session) acceptLoop() {
	for {
		rwc, err := s.transport.AcceptStream(s.ctx)
		if err != nil {
			if s.ctx.Err() == nil {
				slog.Debug("session: accept stream", "err", err)
			}
			return
		}
		header, err := readMessage(rwc)
		if err != nil {
			slog.Warn("session: read sub-stream header", "err", err)
			_ = rwc.Close()
			continue
		}
		name, err := DecodeStreamHeader(header)
		if err != nil {
			slog.Warn("session: parse sub-stream header", "err", err)
			_ = rwc.Close()
			continue
		}
		s.deliverIncomingStream(name, rwc)
	}
}

// datagramLoop pumps incoming datagrams into the right Datagram by
// channel-id. Runs on both sides; each datagram is the AEAD ciphertext
// the peer sent, with no relay framing.
func (s *Session) datagramLoop() {
	for {
		data, err := s.transport.ReceiveDatagram(s.ctx)
		if err != nil {
			if s.ctx.Err() == nil {
				slog.Debug("session: receive datagram", "err", err)
			}
			return
		}
		s.deliverIncomingDatagram(data)
	}
}

// Primary returns the primary stream wrapped as a *Stream. Meaningful
// in pairing-mode where the activation handshake is skipped and the
// primary is otherwise unused — pairing.go and the cross-language
// crypto-peer / pigeon-bridge fixtures use this to talk on the
// primary without opening a sub-stream (the modern client side might
// not support multi-stream QUIC, e.g. Swift NWConnection).
//
// In activation-mode the primary is consumed by activation on Session
// construction; reading from it after activation will block. Primary()
// is safe to call regardless, but only useful in pairing-mode.
func (s *Session) Primary() *Stream {
	return &Stream{name: "", rwc: s.primary, channel: s.channel}
}

// Datagram returns the pre-configured datagram channel by name.
// Calling with an undeclared name is a programmer error.
func (s *Session) Datagram(name string) *Datagram {
	id, ok := s.dgConfig[name]
	if !ok {
		panic(fmt.Sprintf("pigeon: datagram channel %q not declared in Args.Datagrams", name))
	}
	return s.datagrams[id]
}

// deliverIncomingDatagram routes a decrypted, channel-id-prefixed
// datagram to the matching *Datagram's rx queue.
//
// Why not the executor's chan-datagram surface (T39.4 deferred):
// docs/session-protocol.md §Boundaries puts datagram I/O *through*
// the machine's event loop, which would mean threading every inbound
// datagram through HandleEvent(SessionProtocolEventRelayDatagram) and
// dispatching CmdDeliverRecvDatagram via per-channel waiters
// (executor.chanDgWaiters/chanDgBuffers). That migration is non-
// trivial: the executor today is instantiated by Conn with a full
// path-management/relay-reader/lan-reader scaffold that doesn't
// apply to relay-only Session, and Session's per-name Datagram(name)
// returns a *Datagram with a buffered rx channel — a different shape
// from the executor's blocking recvDatagram(ctx) waiter. Standing up
// a Session-shaped executor is its own work item; until then the rx
// channel form below preserves correct demux behaviour without
// pretending the machine drives it.
func (s *Session) deliverIncomingDatagram(payload []byte) {
	if s.channel == nil {
		// Pairing-mode session: datagrams aren't expected. Drop
		// rather than NPE-ing on Decrypt.
		slog.Debug("session: datagram on pairing-mode session, dropping", "peer", s.peerID, "len", len(payload))
		return
	}
	plain, err := s.channel.Decrypt(payload)
	if err != nil {
		slog.Debug("session: datagram decrypt", "peer", s.peerID, "err", err)
		return
	}
	id, body, err := DecodeDatagramPlaintext(plain)
	if err != nil {
		slog.Debug("session: datagram channel id parse", "err", err)
		return
	}
	dg, ok := s.datagrams[id]
	if !ok {
		slog.Debug("session: datagram for unknown channel id", "id", id)
		return
	}
	cp := make([]byte, len(body))
	copy(cp, body)
	select {
	case dg.rx <- cp:
	case <-s.ctx.Done():
	default:
		slog.Warn("session: datagram queue full, dropping", "channel", dg.name)
	}
}

// Close tears down the Session and its underlying QUIC connection.
func (s *Session) Close() error {
	s.closeOnce.Do(func() {
		s.cancel()
		if s.primary != nil {
			_ = s.primary.Close()
		}
		if s.cwirePrimary != nil {
			removeStream(s.cwirePrimary)
			s.cwirePrimary = nil
		}
		if s.cwireRef != nil {
			s.cwireRef.Close()
			s.cwireRef = nil
		}
		if s.ownsTransport && s.transport != nil {
			_ = s.transport.Close()
		}
	})
	return nil
}

// Stream is a reliable, ordered, message-framed channel encrypted with
// the Session's AEAD channel.
type Stream struct {
	name    string
	rwc     io.ReadWriteCloser
	channel *cwire.Channel

	sendMu sync.Mutex
}

// Send writes one message on the stream. In activation-mode sessions
// (s.channel != nil) the message is AEAD-encrypted before going on the
// wire; in pairing-mode sessions (s.channel == nil) the message is sent
// as plaintext — pairing's ceremony does its own crypto and the Session
// is just a bytes pipe. See docs/DESIGN.md §3 L2 for why pairing-mode
// has no AEAD context.
func (s *Stream) Send(msg []byte) error {
	payload := msg
	if s.channel != nil {
		ct, err := s.channel.Encrypt(msg)
		if err != nil {
			return fmt.Errorf("stream: encrypt: %w", err)
		}
		payload = ct
	}
	s.sendMu.Lock()
	defer s.sendMu.Unlock()
	return writeMessage(s.rwc, payload)
}

// Recv reads the next message from the stream. Pairing-mode mirror of
// Send: bytes are returned plaintext when s.channel is nil.
func (s *Stream) Recv(ctx context.Context) ([]byte, error) {
	type result struct {
		data []byte
		err  error
	}
	ch := make(chan result, 1)
	go func() {
		raw, err := readMessage(s.rwc)
		if err != nil {
			ch <- result{err: err}
			return
		}
		if s.channel == nil {
			ch <- result{data: raw}
			return
		}
		plain, err := s.channel.Decrypt(raw)
		ch <- result{data: plain, err: err}
	}()
	select {
	case r := <-ch:
		return r.data, r.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Name returns the channel name supplied to OpenStream.
func (s *Stream) Name() string { return s.name }

// Close closes the underlying QUIC stream.
func (s *Stream) Close() error { return s.rwc.Close() }

// UnderlyingPairingRWC returns the raw bidirectional byte pipe behind a
// pairing-mode stream (one where the Session has no AEAD channel yet,
// because the pairing ceremony itself is what derives the channel). The
// pairing package uses this to hand the stream to libpigeon's
// pigeon_pair_acceptor / pigeon_pair_initiator via a single-stream
// cwire.GoTransport adapter.
//
// Returns nil if the stream is in activation-mode (s.channel != nil), so
// callers can't accidentally bypass AEAD on a live session.
func (s *Stream) UnderlyingPairingRWC() io.ReadWriteCloser {
	if s.channel != nil {
		return nil
	}
	return s.rwc
}

// Datagram is an unreliable, unordered message channel keyed by a
// pre-agreed varint channel ID.
type Datagram struct {
	name    string
	id      uint64
	session *Session
	rx      chan []byte
}

// Send transmits a single datagram on this channel. The wire format is
// AEAD([varint channel-id][payload]); the relay forwards it opaquely on
// this session's own connection, so no routing prefix is added.
func (d *Datagram) Send(payload []byte) error {
	plain := EncodeDatagramPlaintext(d.id, payload)
	ct, err := d.session.channel.Encrypt(plain)
	if err != nil {
		return fmt.Errorf("datagram: encrypt: %w", err)
	}
	return d.session.transport.SendDatagram(ct)
}

// Recv blocks until the next datagram on this channel arrives.
func (d *Datagram) Recv(ctx context.Context) ([]byte, error) {
	select {
	case msg, ok := <-d.rx:
		if !ok {
			return nil, io.ErrClosedPipe
		}
		return msg, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-d.session.ctx.Done():
		return nil, io.ErrClosedPipe
	}
}

// Name returns the channel name supplied to Args.Datagrams.
func (d *Datagram) Name() string { return d.name }

// Wire helpers are protogen-generated in wire_gen.go from
// protocol/wireformats.yaml — EncodeStreamHeader, DecodeStreamHeader,
// EncodeDatagramPlaintext, DecodeDatagramPlaintext,
// EncodeRelayGreetingConnect/Register/Listen, DecodeRelayGreeting. Do
// not add hand-rolled equivalents here — extend the YAML instead.

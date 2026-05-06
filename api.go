// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package pigeon

import (
	"context"
	"crypto/tls"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"sync"

	"github.com/marcelocantos/pigeon/crypto"
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

// Listener is the backend-side acceptor for paired clients.
type Listener struct {
	args       *RegisterArgs
	transport  *transport
	instanceID string

	ctx    context.Context
	cancel context.CancelFunc

	mu       sync.Mutex
	sessions map[uint32]*Session // keyed by relay-assigned clientTag

	accepted chan acceptResult

	closeOnce sync.Once
}

type acceptResult struct {
	session *Session
	err     error
}

// Register publishes a backend on the relay and returns a Listener that
// yields one Session per accepted client. The Listener owns a single
// QUIC connection to the relay across its lifetime; each Accept call
// returns the next paired client without re-registering.
func Register(ctx context.Context, args *RegisterArgs) (*Listener, string, error) {
	if args == nil {
		return nil, "", fmt.Errorf("pigeon.Register: args is nil")
	}
	if args.Identity == nil {
		return nil, "", fmt.Errorf("pigeon.Register: Identity is required")
	}
	if args.Pairing == nil {
		return nil, "", fmt.Errorf("pigeon.Register: Pairing callback is required")
	}
	if args.Relay == "" {
		return nil, "", fmt.Errorf("pigeon.Register: Relay is required")
	}
	if err := validateDatagrams(args.Datagrams); err != nil {
		return nil, "", fmt.Errorf("pigeon.Register: %w", err)
	}

	tlsCfg := args.TLS
	if tlsCfg == nil {
		tlsCfg = &tls.Config{InsecureSkipVerify: true}
	}
	cfg := Config{
		TLS:        tlsCfg,
		Token:      args.Token,
		InstanceID: args.Identity.InstanceID(),
	}
	tr, err := dialAcceptor(ctx, args.Relay, cfg)
	if err != nil {
		return nil, "", fmt.Errorf("relay register: %w", err)
	}

	lctx, cancel := context.WithCancel(ctx)
	l := &Listener{
		args:       args,
		transport:  tr,
		instanceID: tr.instanceID,
		ctx:        lctx,
		cancel:     cancel,
		sessions:   make(map[uint32]*Session),
		accepted:   make(chan acceptResult, 8),
	}
	go l.acceptLoop()
	go l.datagramLoop()
	return l, tr.instanceID, nil
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

// Close releases the Listener and the underlying relay session.
func (l *Listener) Close() error {
	l.closeOnce.Do(func() {
		l.cancel()
		_ = l.transport.Close()
	})
	return nil
}

// acceptLoop reads incoming streams from the relay and demuxes by the
// 4-byte clientTag header the relay prepends. New tag → primary stream
// → run connect-hello/ack and create a Session. Existing tag → sub-stream
// → deliver to the matching Session.
func (l *Listener) acceptLoop() {
	for {
		stream, err := l.transport.AcceptStream(l.ctx)
		if err != nil {
			if l.ctx.Err() == nil {
				slog.Debug("listener: accept stream", "err", err)
			}
			return
		}
		header, err := readMessage(stream)
		if err != nil {
			slog.Warn("listener: read stream header", "err", err)
			_ = stream.Close()
			continue
		}
		tag, name, err := decodeBackendStreamHeader(header)
		if err != nil {
			slog.Warn("listener: parse stream header", "err", err)
			_ = stream.Close()
			continue
		}
		l.mu.Lock()
		sess, exists := l.sessions[tag]
		l.mu.Unlock()
		if exists {
			// Sub-stream for an existing client: dispatch by name into
			// the Session's incoming-stream queue.
			sess.deliverIncomingStream(name, stream)
			continue
		}
		if name != "" {
			slog.Warn("listener: unknown tag with non-empty name", "tag", tag, "name", name)
			_ = stream.Close()
			continue
		}
		// New client: primary stream — run connect-hello/ack inline.
		go l.acceptPrimary(tag, stream)
	}
}

// acceptPrimary runs the auth_request / auth_ok activation handshake
// on a newly arrived client primary stream and registers the resulting
// Session. Activation is driven through a per-client SessionMachine
// (see runBackendActivation) — the spec models the auth flow as
// Paired → AuthCheck → SessionActive, with the device_known guard
// bound from the result of args.Pairing(deviceID).
func (l *Listener) acceptPrimary(tag uint32, stream io.ReadWriteCloser) {
	rawMachine, deviceID, rec, err := runBackendActivation(stream, l.args.Pairing)
	if err != nil {
		slog.Warn("listener: activation", "err", err)
		_ = stream.Close()
		if deviceID != "" {
			l.accepted <- acceptResult{err: err}
		}
		return
	}
	machine, err := newBackendSessionMachine(rawMachine)
	if err != nil {
		_ = stream.Close()
		l.accepted <- acceptResult{err: fmt.Errorf("post-activation transition: %w", err)}
		return
	}
	channel, err := rec.DeriveChannel([]byte("backend->client"), []byte("client->backend"))
	if err != nil {
		_ = stream.Close()
		l.accepted <- acceptResult{err: fmt.Errorf("derive session channel: %w", err)}
		return
	}

	sess := newSession(l.ctx, l.transport, channel, deviceID, tag, l.args.Datagrams, true)
	sess.machine = machine
	sess.bindPrimary(stream)

	l.mu.Lock()
	l.sessions[tag] = sess
	l.mu.Unlock()

	select {
	case l.accepted <- acceptResult{session: sess}:
	case <-l.ctx.Done():
		_ = sess.Close()
	}
}

// datagramLoop reads incoming datagrams from the relay, parses the
// 4-byte clientTag prefix the relay prepends, and dispatches to the
// owning Session's datagram pump.
func (l *Listener) datagramLoop() {
	for {
		data, err := l.transport.ReceiveDatagram(l.ctx)
		if err != nil {
			if l.ctx.Err() == nil {
				slog.Debug("listener: receive datagram", "err", err)
			}
			return
		}
		if len(data) < 4 {
			slog.Warn("listener: datagram too short for tag", "len", len(data))
			continue
		}
		tag := binary.BigEndian.Uint32(data[:4])
		l.mu.Lock()
		sess, ok := l.sessions[tag]
		l.mu.Unlock()
		if !ok {
			slog.Debug("listener: datagram for unknown tag", "tag", tag)
			continue
		}
		sess.deliverIncomingDatagram(data[4:])
	}
}

// removeSession is called by Session.Close to drop the Listener's reference.
func (l *Listener) removeSession(tag uint32) {
	l.mu.Lock()
	delete(l.sessions, tag)
	l.mu.Unlock()
}

// Connect dials the backend identified by args.InstanceID and returns a
// Session ready for OpenStream / Datagram. The client side runs a single
// Session over a fresh QUIC connection to the relay; sub-streams open on
// that same QUIC connection.
func Connect(ctx context.Context, args *ConnectArgs) (*Session, error) {
	if args == nil {
		return nil, fmt.Errorf("pigeon.Connect: args is nil")
	}
	if args.InstanceID == "" {
		return nil, fmt.Errorf("pigeon.Connect: InstanceID is required")
	}
	if args.Record == nil {
		return nil, fmt.Errorf("pigeon.Connect: Record is required")
	}
	if args.Identity == nil {
		return nil, fmt.Errorf("pigeon.Connect: Identity is required")
	}
	if args.Relay == "" {
		return nil, fmt.Errorf("pigeon.Connect: Relay is required")
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

	// Primary stream is the one the relay handshake already used. The
	// header carries an empty name, which (after the relay prepends the
	// clientTag) tells the backend "this is a new client primary".
	if err := writeMessage(tr.primary, encodeStreamHeader(false, 0, "")); err != nil {
		_ = tr.Close()
		return nil, fmt.Errorf("write primary header: %w", err)
	}
	rawMachine, err := runClientActivation(tr.primary, args.Identity.InstanceID())
	if err != nil {
		_ = tr.Close()
		return nil, err
	}
	machine, err := newClientSessionMachine(rawMachine)
	if err != nil {
		_ = tr.Close()
		return nil, fmt.Errorf("post-activation transition: %w", err)
	}

	channel, err := args.Record.DeriveChannel([]byte("client->backend"), []byte("backend->client"))
	if err != nil {
		_ = tr.Close()
		return nil, fmt.Errorf("derive session channel: %w", err)
	}

	sCtx, cancel := context.WithCancel(ctx)
	sess := newSession(sCtx, tr, channel, args.InstanceID, 0, args.Datagrams, false)
	sess.cancel = cancel
	sess.ownsTransport = true
	sess.machine = machine
	sess.bindPrimary(tr.primary)
	go sess.clientAcceptLoop()
	go sess.clientDatagramLoop()
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
type Session struct {
	ctx    context.Context
	cancel context.CancelFunc

	transport     *transport
	channel       *crypto.Channel
	peerID        string
	clientTag     uint32 // 0 on client side; relay-assigned on backend side
	isBackend     bool
	ownsTransport bool // client side owns its transport; backend Sessions don't

	listener *Listener // backend side only

	primary io.ReadWriteCloser

	// machine is the post-activation SessionMachine for this session.
	// Per docs/session-protocol.md §Boundaries, per-stream I/O lives
	// above the machine; the machine drives lifecycle transitions
	// (activation already done by runBackendActivation/runClientActivation
	// when this Session is constructed; disconnect on Close) and any
	// future executor-mediated I/O (datagrams, health monitor, LAN).
	machine *sessionMachine

	// Stream rendezvous: pendingOpens map name → waiter; bufferedStreams
	// holds streams that arrived before a local OpenStream caller.
	streamMu        sync.Mutex
	pendingOpens    map[string][]chan io.ReadWriteCloser
	bufferedStreams map[string][]io.ReadWriteCloser

	// Datagrams are pre-declared by name → channel-id at construction.
	dgConfig  map[string]uint64
	datagrams map[uint64]*Datagram

	closeOnce sync.Once
}

func newSession(ctx context.Context, tr *transport, channel *crypto.Channel, peerID string, tag uint32, dgConfig map[string]uint64, isBackend bool) *Session {
	sCtx, cancel := context.WithCancel(ctx)
	s := &Session{
		ctx:             sCtx,
		cancel:          cancel,
		transport:       tr,
		channel:         channel,
		peerID:          peerID,
		clientTag:       tag,
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
// identified by `name`. Both peers must call OpenStream with the same
// name; the matching call on the peer unblocks when the name binds.
//
// On the backend side this opens a new QUIC stream toward the client
// (with the relay routing by clientTag); on the client side it opens
// a new QUIC stream toward the relay.
func (s *Session) OpenStream(ctx context.Context, name string) (*Stream, error) {
	if name == "" {
		return nil, errors.New("OpenStream: name must be non-empty")
	}
	rwc, err := s.transport.OpenStream()
	if err != nil {
		return nil, fmt.Errorf("open quic stream: %w", err)
	}
	if err := writeMessage(rwc, encodeStreamHeader(s.isBackend, s.clientTag, name)); err != nil {
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

// deliverIncomingStream is called by the transport layer when a new
// peer-opened stream's name handshake has been read.
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
	plain, err := s.channel.Decrypt(payload)
	if err != nil {
		slog.Debug("session: datagram decrypt", "peer", s.peerID, "err", err)
		return
	}
	id, n := binary.Uvarint(plain)
	if n <= 0 {
		slog.Debug("session: datagram channel id parse")
		return
	}
	dg, ok := s.datagrams[id]
	if !ok {
		slog.Debug("session: datagram for unknown channel id", "id", id)
		return
	}
	body := plain[n:]
	cp := make([]byte, len(body))
	copy(cp, body)
	select {
	case dg.rx <- cp:
	case <-s.ctx.Done():
	default:
		slog.Warn("session: datagram queue full, dropping", "channel", dg.name)
	}
}

// clientAcceptLoop runs on the client side: reads each new sub-stream
// the backend opens (rare in v1) and delivers by name.
func (s *Session) clientAcceptLoop() {
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
		name, err := decodeClientStreamHeader(header)
		if err != nil {
			slog.Warn("session: parse sub-stream header", "err", err)
			_ = rwc.Close()
			continue
		}
		s.deliverIncomingStream(name, rwc)
	}
}

// clientDatagramLoop runs on the client side: pumps incoming datagrams
// into the right Datagram by channel-id (no clientTag prefix on the
// client side — relay strips it).
func (s *Session) clientDatagramLoop() {
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

// Close tears down the Session. The SessionMachine is advanced
// through the spec's RelayConnected → Paired transition before the
// underlying QUIC primary is torn down so the spec model has a
// defined post-close state.
func (s *Session) Close() error {
	s.closeOnce.Do(func() {
		if s.machine != nil {
			if err := s.machine.disconnect(); err != nil {
				slog.Debug("session: disconnect transition", "peer", s.peerID, "err", err)
			}
		}
		s.cancel()
		if s.primary != nil {
			_ = s.primary.Close()
		}
		if s.listener != nil {
			s.listener.removeSession(s.clientTag)
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
	channel *crypto.Channel

	sendMu sync.Mutex
}

// Send writes one message on the stream.
func (s *Stream) Send(msg []byte) error {
	ct := s.channel.Encrypt(msg)
	s.sendMu.Lock()
	defer s.sendMu.Unlock()
	return writeMessage(s.rwc, ct)
}

// Recv reads the next message from the stream.
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

// Datagram is an unreliable, unordered message channel keyed by a
// pre-agreed varint channel ID.
type Datagram struct {
	name    string
	id      uint64
	session *Session
	rx      chan []byte
}

// Send transmits a single datagram on this channel. The wire format is
// AEAD([varint channel-id][payload]); on the backend side a 4-byte
// clientTag prefix is added by the Session for relay routing.
func (d *Datagram) Send(payload []byte) error {
	buf := make([]byte, 0, binary.MaxVarintLen64+len(payload))
	var idBuf [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(idBuf[:], d.id)
	buf = append(buf, idBuf[:n]...)
	buf = append(buf, payload...)
	ct := d.session.channel.Encrypt(buf)
	if d.session.isBackend {
		framed := make([]byte, 4+len(ct))
		binary.BigEndian.PutUint32(framed[:4], d.session.clientTag)
		copy(framed[4:], ct)
		return d.session.transport.SendDatagram(framed)
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

// --- wire helpers ---

// encodeStreamHeader builds the per-stream first-message header.
//
//	Backend side: [4-byte clientTag-BE][varint name-len][name].
//	Client side:                       [varint name-len][name].
//
// The whole thing is sent as one length-prefixed message via writeMessage,
// so the relay (which only exposes ReadMessage / WriteMessage on its
// stream wrappers) can prepend the 4-byte tag without raw-byte access.
func encodeStreamHeader(isBackend bool, tag uint32, name string) []byte {
	const tagSize = 4
	cap := binary.MaxVarintLen64 + len(name)
	if isBackend {
		cap += tagSize
	}
	buf := make([]byte, 0, cap)
	if isBackend {
		var tagBuf [tagSize]byte
		binary.BigEndian.PutUint32(tagBuf[:], tag)
		buf = append(buf, tagBuf[:]...)
	}
	var lenBuf [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(lenBuf[:], uint64(len(name)))
	buf = append(buf, lenBuf[:n]...)
	buf = append(buf, []byte(name)...)
	return buf
}

// decodeBackendStreamHeader parses [4-byte tag][varint name-len][name].
func decodeBackendStreamHeader(b []byte) (uint32, string, error) {
	if len(b) < 4 {
		return 0, "", errors.New("header too short for tag")
	}
	tag := binary.BigEndian.Uint32(b[:4])
	name, err := decodeClientStreamHeader(b[4:])
	if err != nil {
		return 0, "", err
	}
	return tag, name, nil
}

// decodeClientStreamHeader parses [varint name-len][name].
func decodeClientStreamHeader(b []byte) (string, error) {
	nameLen, n := binary.Uvarint(b)
	if n <= 0 {
		return "", errors.New("bad name varint")
	}
	if uint64(len(b)-n) < nameLen {
		return "", errors.New("name truncated")
	}
	return string(b[n : n+int(nameLen)]), nil
}

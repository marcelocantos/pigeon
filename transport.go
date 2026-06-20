// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package pigeon

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/webtransport-go"
)

// datagrammer provides unreliable datagram send/receive on the
// underlying QUIC connection or WebTransport session.
type datagrammer interface {
	SendDatagram([]byte) error
	ReceiveDatagram(context.Context) ([]byte, error)
}

// streamOpener can open additional bidirectional streams on the
// underlying QUIC connection or WebTransport session.
type streamOpener interface {
	OpenStream() (io.ReadWriteCloser, error)
}

// streamAcceptor can accept incoming bidirectional streams.
type streamAcceptor interface {
	AcceptStream(context.Context) (io.ReadWriteCloser, error)
}

// deadliner can set read/write deadlines on a stream. The executor's
// pong-timeout / cutover-drain machinery uses this to bound waits.
type deadliner interface {
	SetReadDeadline(time.Time) error
	SetWriteDeadline(time.Time) error
}

// transport is a lightweight wrapper around a QUIC-backed relay
// session (raw QUIC or WebTransport) used by the Session API.
type transport struct {
	primary    io.ReadWriteCloser
	opener     streamOpener
	acceptor   streamAcceptor
	dg         datagrammer
	closer     io.Closer
	instanceID string
}

func (t *transport) OpenStream() (io.ReadWriteCloser, error) {
	return t.opener.OpenStream()
}

func (t *transport) AcceptStream(ctx context.Context) (io.ReadWriteCloser, error) {
	return t.acceptor.AcceptStream(ctx)
}

func (t *transport) SendDatagram(data []byte) error {
	return t.dg.SendDatagram(data)
}

func (t *transport) ReceiveDatagram(ctx context.Context) ([]byte, error) {
	return t.dg.ReceiveDatagram(ctx)
}

func (t *transport) Close() error {
	if t.primary != nil {
		_ = t.primary.Close()
	}
	if t.closer != nil {
		return t.closer.Close()
	}
	return nil
}

// relayQUICConfig is the QUIC configuration used for every relay dial
// (register, listen, connect). Datagrams enabled; keep-alive well under
// the idle timeout so parked listen connections survive while waiting
// for a client.
var relayQUICConfig = &quic.Config{
	EnableDatagrams:      true,
	MaxIdleTimeout:       60 * time.Second,
	KeepAlivePeriod:      10 * time.Second,
	HandshakeIdleTimeout: 30 * time.Second,
}

// dialRegister opens the backend's control connection. The relay holds
// it open for the instance's lifetime as a liveness signal; no client
// traffic flows on it. The relay assigns (or echoes) the instance ID,
// returned in transport.instanceID. Each accepted client rides its own
// listen connection (dialListen), not this one. See docs/DESIGN.md §3 L1.
func dialRegister(ctx context.Context, relayURL string, c Config) (*transport, error) {
	WakeRelay(ctx, relayURL, c)
	if c.WebTransport {
		return dialRegisterWebTransport(ctx, relayURL, c)
	}
	conn, stream, err := dialRelayQUIC(ctx, relayURL, c)
	if err != nil {
		return nil, fmt.Errorf("register: %w", err)
	}
	if err := writeMessage(stream, EncodeRelayGreetingRegister(c.Token, c.InstanceID)); err != nil {
		conn.CloseWithError(0, "send handshake")
		return nil, fmt.Errorf("register: handshake: %w", err)
	}
	idBytes, err := readMessage(stream)
	if err != nil {
		conn.CloseWithError(0, "read ID")
		return nil, fmt.Errorf("register: read ID: %w", err)
	}
	return quicTransport(conn, stream, string(idBytes)), nil
}

// dialListen parks a fresh QUIC connection as an available slot for the
// instance. The relay matches the next arriving client to it and bridges
// the two connections end-to-end; the returned transport then carries
// that one client's session. The backend keeps several of these
// outstanding (see Listener's listen pool) so clients never wait.
func dialListen(ctx context.Context, relayURL, instanceID string, c Config) (*transport, error) {
	if c.WebTransport {
		return dialListenWebTransport(ctx, relayURL, instanceID, c)
	}
	conn, stream, err := dialRelayQUIC(ctx, relayURL, c)
	if err != nil {
		return nil, fmt.Errorf("listen: %w", err)
	}
	if err := writeMessage(stream, EncodeRelayGreetingListen(c.Token, instanceID)); err != nil {
		conn.CloseWithError(0, "send handshake")
		return nil, fmt.Errorf("listen: handshake: %w", err)
	}
	if _, err := readMessage(stream); err != nil {
		conn.CloseWithError(0, "read ack")
		return nil, fmt.Errorf("listen: read ack: %w", err)
	}
	return quicTransport(conn, stream, instanceID), nil
}

// dialInitiator connects to a registered peer through the relay. The
// "ok" ack returns only once the relay has matched a backend listen and
// the end-to-end bridge is live, so the caller can begin the session
// handshake immediately afterward.
func dialInitiator(ctx context.Context, relayURL, instanceID string, c Config) (*transport, error) {
	WakeRelay(ctx, relayURL, c)
	if c.WebTransport {
		return dialInitiatorWebTransport(ctx, relayURL, instanceID, c)
	}
	conn, stream, err := dialRelayQUIC(ctx, relayURL, c)
	if err != nil {
		return nil, fmt.Errorf("connect to %s: %w", instanceID, err)
	}
	if err := writeMessage(stream, EncodeRelayGreetingConnect(instanceID)); err != nil {
		conn.CloseWithError(0, "send handshake")
		return nil, fmt.Errorf("connect: handshake: %w", err)
	}
	if _, err := readMessage(stream); err != nil {
		conn.CloseWithError(0, "read ack")
		return nil, fmt.Errorf("connect: read ack: %w", err)
	}
	return quicTransport(conn, stream, instanceID), nil
}

// dialRelayQUIC dials the relay over raw QUIC and opens the primary
// stream that carries the greeting. Shared by all three roles.
func dialRelayQUIC(ctx context.Context, relayURL string, c Config) (*quic.Conn, *quic.Stream, error) {
	addr, err := quicAddr(relayURL, c)
	if err != nil {
		return nil, nil, err
	}
	conn, err := quic.DialAddr(ctx, addr, quicTLSConfig(c), relayQUICConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("quic dial: %w", err)
	}
	stream, err := conn.OpenStream()
	if err != nil {
		conn.CloseWithError(0, "open stream")
		return nil, nil, fmt.Errorf("open stream: %w", err)
	}
	return conn, stream, nil
}

func quicTransport(conn *quic.Conn, stream *quic.Stream, instanceID string) *transport {
	return &transport{
		primary:    stream,
		opener:     quicOpener{conn},
		acceptor:   quicAcceptor{conn},
		dg:         conn,
		closer:     quicCloser{conn},
		instanceID: instanceID,
	}
}

func dialRegisterWebTransport(ctx context.Context, relayURL string, c Config) (*transport, error) {
	session, stream, err := dialRelayWebTransport(ctx, relayURL, c)
	if err != nil {
		return nil, fmt.Errorf("register: %w", err)
	}
	if err := writeMessage(stream, EncodeRelayGreetingRegister(c.Token, c.InstanceID)); err != nil {
		session.CloseWithError(0, "send handshake")
		return nil, fmt.Errorf("register: handshake: %w", err)
	}
	idBytes, err := readMessage(stream)
	if err != nil {
		session.CloseWithError(0, "read ID")
		return nil, fmt.Errorf("register: read ID: %w", err)
	}
	return wtTransport(session, stream, string(idBytes)), nil
}

func dialListenWebTransport(ctx context.Context, relayURL, instanceID string, c Config) (*transport, error) {
	session, stream, err := dialRelayWebTransport(ctx, relayURL, c)
	if err != nil {
		return nil, fmt.Errorf("listen: %w", err)
	}
	if err := writeMessage(stream, EncodeRelayGreetingListen(c.Token, instanceID)); err != nil {
		session.CloseWithError(0, "send handshake")
		return nil, fmt.Errorf("listen: handshake: %w", err)
	}
	if _, err := readMessage(stream); err != nil {
		session.CloseWithError(0, "read ack")
		return nil, fmt.Errorf("listen: read ack: %w", err)
	}
	return wtTransport(session, stream, instanceID), nil
}

func dialInitiatorWebTransport(ctx context.Context, relayURL, instanceID string, c Config) (*transport, error) {
	session, stream, err := dialRelayWebTransport(ctx, relayURL, c)
	if err != nil {
		return nil, fmt.Errorf("connect to %s: %w", instanceID, err)
	}
	if err := writeMessage(stream, EncodeRelayGreetingConnect(instanceID)); err != nil {
		session.CloseWithError(0, "send handshake")
		return nil, fmt.Errorf("connect: handshake: %w", err)
	}
	if _, err := readMessage(stream); err != nil {
		session.CloseWithError(0, "read ack")
		return nil, fmt.Errorf("connect: read ack: %w", err)
	}
	return wtTransport(session, stream, instanceID), nil
}

// dialRelayWebTransport opens a WebTransport session to the relay's
// single /pigeon endpoint and opens the primary stream that carries the
// greeting. The token rides both the greeting (consumed by
// BearerTokenAuth) and the Authorization header (for browser clients
// that can only set the latter).
func dialRelayWebTransport(ctx context.Context, relayURL string, c Config) (*webtransport.Session, *webtransport.Stream, error) {
	tlsConfig := c.TLS
	if tlsConfig == nil {
		tlsConfig = &tls.Config{}
	}
	d := webtransport.Dialer{TLSClientConfig: tlsConfig}
	hdr := http.Header{}
	if c.Token != "" {
		hdr.Set("Authorization", "Bearer "+c.Token)
	}
	_, session, err := d.Dial(ctx, relayURL+"/pigeon", hdr)
	if err != nil {
		return nil, nil, err
	}
	stream, err := session.OpenStream()
	if err != nil {
		session.CloseWithError(0, "open stream")
		return nil, nil, fmt.Errorf("open stream: %w", err)
	}
	return session, stream, nil
}

func wtTransport(session *webtransport.Session, stream *webtransport.Stream, instanceID string) *transport {
	return &transport{
		primary:    stream,
		opener:     wtOpener{session},
		acceptor:   wtAcceptor{session},
		dg:         &wtDatagrammer{session: session},
		closer:     wtCloser{session},
		instanceID: instanceID,
	}
}

// wtDatagrammer adapts a webtransport.Session to the datagrammer interface.
type wtDatagrammer struct {
	session *webtransport.Session
}

func (w *wtDatagrammer) SendDatagram(data []byte) error {
	return w.session.SendDatagram(data)
}

func (w *wtDatagrammer) ReceiveDatagram(ctx context.Context) ([]byte, error) {
	return w.session.ReceiveDatagram(ctx)
}

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

// dialAcceptor registers the local peer with the relay and returns a
// transport ready for AcceptStream/datagram I/O. The relay assigns (or
// echoes) an instance ID, returned in transport.instanceID.
func dialAcceptor(ctx context.Context, relayURL string, c Config) (*transport, error) {
	WakeRelay(ctx, relayURL, c)
	if c.WebTransport {
		return dialAcceptorWebTransport(ctx, relayURL, c)
	}
	return dialAcceptorQUIC(ctx, relayURL, c)
}

// dialInitiator connects to a registered peer through the relay.
func dialInitiator(ctx context.Context, relayURL, instanceID string, c Config) (*transport, error) {
	WakeRelay(ctx, relayURL, c)
	if c.WebTransport {
		return dialInitiatorWebTransport(ctx, relayURL, instanceID, c)
	}
	return dialInitiatorQUIC(ctx, relayURL, instanceID, c)
}

func dialAcceptorQUIC(ctx context.Context, relayURL string, c Config) (*transport, error) {
	addr, err := quicAddr(relayURL, c)
	if err != nil {
		return nil, err
	}
	conn, err := quic.DialAddr(ctx, addr, quicTLSConfig(c), &quic.Config{
		EnableDatagrams:      true,
		MaxIdleTimeout:       60 * time.Second,
		KeepAlivePeriod:      10 * time.Second,
		HandshakeIdleTimeout: 30 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("register: quic dial: %w", err)
	}
	stream, err := conn.OpenStream()
	if err != nil {
		conn.CloseWithError(0, "open stream")
		return nil, fmt.Errorf("register: open stream: %w", err)
	}
	handshake := EncodeRelayGreetingRegisterMux(c.Token, c.InstanceID)
	if err := writeMessage(stream, handshake); err != nil {
		conn.CloseWithError(0, "send handshake")
		return nil, fmt.Errorf("register: handshake: %w", err)
	}
	idBytes, err := readMessage(stream)
	if err != nil {
		conn.CloseWithError(0, "read ID")
		return nil, fmt.Errorf("register: read ID: %w", err)
	}
	return &transport{
		primary:    stream,
		opener:     quicOpener{conn},
		acceptor:   quicAcceptor{conn},
		dg:         conn,
		closer:     quicCloser{conn},
		instanceID: string(idBytes),
	}, nil
}

func dialInitiatorQUIC(ctx context.Context, relayURL, instanceID string, c Config) (*transport, error) {
	addr, err := quicAddr(relayURL, c)
	if err != nil {
		return nil, err
	}
	conn, err := quic.DialAddr(ctx, addr, quicTLSConfig(c), &quic.Config{
		EnableDatagrams:      true,
		MaxIdleTimeout:       60 * time.Second,
		KeepAlivePeriod:      10 * time.Second,
		HandshakeIdleTimeout: 30 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("connect to %s: quic dial: %w", instanceID, err)
	}
	stream, err := conn.OpenStream()
	if err != nil {
		conn.CloseWithError(0, "open stream")
		return nil, fmt.Errorf("connect: open stream: %w", err)
	}
	if err := writeMessage(stream, EncodeRelayGreetingConnect(instanceID)); err != nil {
		conn.CloseWithError(0, "send handshake")
		return nil, fmt.Errorf("connect: handshake: %w", err)
	}
	return &transport{
		primary:    stream,
		opener:     quicOpener{conn},
		acceptor:   quicAcceptor{conn},
		dg:         conn,
		closer:     quicCloser{conn},
		instanceID: instanceID,
	}, nil
}

func dialAcceptorWebTransport(ctx context.Context, relayURL string, c Config) (*transport, error) {
	tlsConfig := c.TLS
	if tlsConfig == nil {
		tlsConfig = &tls.Config{}
	}
	d := webtransport.Dialer{TLSClientConfig: tlsConfig}
	hdr := http.Header{}
	if c.Token != "" {
		hdr.Set("Authorization", "Bearer "+c.Token)
	}
	_, session, err := d.Dial(ctx, relayURL+"/register", hdr)
	if err != nil {
		return nil, fmt.Errorf("register: %w", err)
	}
	stream, err := session.OpenStream()
	if err != nil {
		session.CloseWithError(0, "open stream")
		return nil, fmt.Errorf("register: open stream: %w", err)
	}
	handshake := EncodeRelayGreetingRegisterMux("", c.InstanceID)
	if err := writeMessage(stream, handshake); err != nil {
		session.CloseWithError(0, "send handshake")
		return nil, fmt.Errorf("register: handshake: %w", err)
	}
	idBytes, err := readMessage(stream)
	if err != nil {
		session.CloseWithError(0, "read ID")
		return nil, fmt.Errorf("register: read ID: %w", err)
	}
	return &transport{
		primary:    stream,
		opener:     wtOpener{session},
		acceptor:   wtAcceptor{session},
		dg:         &wtDatagrammer{session: session},
		closer:     wtCloser{session},
		instanceID: string(idBytes),
	}, nil
}

func dialInitiatorWebTransport(ctx context.Context, relayURL, instanceID string, c Config) (*transport, error) {
	tlsConfig := c.TLS
	if tlsConfig == nil {
		tlsConfig = &tls.Config{}
	}
	d := webtransport.Dialer{TLSClientConfig: tlsConfig}
	_, session, err := d.Dial(ctx, relayURL+"/ws/"+instanceID, nil)
	if err != nil {
		return nil, fmt.Errorf("connect to %s: %w", instanceID, err)
	}
	stream, err := session.OpenStream()
	if err != nil {
		session.CloseWithError(0, "open stream")
		return nil, fmt.Errorf("connect: open stream: %w", err)
	}
	if err := writeMessage(stream, []byte("connect")); err != nil {
		session.CloseWithError(0, "send handshake")
		return nil, fmt.Errorf("connect: handshake: %w", err)
	}
	if _, err := readMessage(stream); err != nil {
		session.CloseWithError(0, "read ack")
		return nil, fmt.Errorf("connect: read ack: %w", err)
	}
	return &transport{
		primary:    stream,
		opener:     wtOpener{session},
		acceptor:   wtAcceptor{session},
		dg:         &wtDatagrammer{session: session},
		closer:     wtCloser{session},
		instanceID: instanceID,
	}, nil
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

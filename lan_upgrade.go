// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package pigeon

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/quic-go/quic-go"
)

// lanUpgradeTimeout bounds how long each side waits for the direct-path
// challenge/response after an offer is sent or received.
const lanUpgradeTimeout = 10 * time.Second

// offerLAN runs on the backend after activation when RegisterArgs.LAN is
// set. It registers a challenge with the LANServer, opens the reserved
// control stream over the relay, sends the encrypted lanOffer, and on
// successful client verification swaps the Session transport to the
// direct QUIC connection.
func (s *Session) offerLAN() {
	if s.lanServer == nil || s.peerID == "" {
		return
	}
	challenge := make([]byte, 32)
	if _, err := rand.Read(challenge); err != nil {
		slog.Warn("session: LAN challenge", "err", err)
		return
	}

	verified := make(chan *transport, 1)
	s.lanServer.RegisterPending(s.peerID, challenge, func(stream io.ReadWriteCloser, conn *quic.Conn) {
		select {
		case verified <- lanTransport(conn, stream, s.peerID):
		default:
		}
	})
	defer s.lanServer.UnregisterPending(s.peerID)

	ctx, cancel := context.WithTimeout(s.ctx, lanUpgradeTimeout)
	defer cancel()

	st, err := s.OpenStream(ctx, lanControlStreamName)
	if err != nil {
		slog.Debug("session: open LAN control stream", "err", err)
		return
	}
	// Keep the control stream open until upgrade completes so the client
	// side does not see an early close while dialing.
	defer st.Close()

	offer := lanOffer{
		Addr:      s.lanServer.Addr(),
		Challenge: challenge,
		CertHash:  s.lanServer.CertHash(),
	}
	payload, err := json.Marshal(offer)
	if err != nil {
		return
	}
	if err := st.Send(payload); err != nil {
		slog.Debug("session: send LAN offer", "err", err)
		return
	}

	select {
	case tr := <-verified:
		s.swapTransport(tr)
		slog.Info("session: backend upgraded to LAN", "peer", s.peerID, "addr", s.lanServer.Addr())
	case <-ctx.Done():
		slog.Debug("session: LAN upgrade timed out (backend)", "peer", s.peerID)
	}
}

// handleLANOffer runs on the client when PreferLAN is set and the
// reserved LAN control stream arrives. It dials the advertised address,
// completes the challenge/response, and swaps the Session transport.
func (s *Session) handleLANOffer(rwc io.ReadWriteCloser) {
	defer rwc.Close()
	st := &Stream{
		name:    lanControlStreamName,
		rwc:     rwc,
		channel: s.streamChannel(lanControlStreamName),
	}

	ctx, cancel := context.WithTimeout(s.ctx, lanUpgradeTimeout)
	defer cancel()

	raw, err := st.Recv(ctx)
	if err != nil {
		slog.Debug("session: recv LAN offer", "err", err)
		return
	}
	var offer lanOffer
	if err := json.Unmarshal(raw, &offer); err != nil {
		slog.Debug("session: bad LAN offer", "err", err)
		return
	}
	if offer.Addr == "" || len(offer.Challenge) == 0 {
		return
	}

	tr, err := dialLAN(ctx, offer, s.localID, s.lanTLS)
	if err != nil {
		slog.Debug("session: LAN dial", "addr", offer.Addr, "err", err)
		return
	}
	s.swapTransport(tr)
	slog.Info("session: client upgraded to LAN", "addr", offer.Addr)
}

// dialLAN connects to the backend's LANServer, sends lanVerify, and
// returns a transport wrapping the verified connection.
func dialLAN(ctx context.Context, offer lanOffer, clientID string, lanTLS *tls.Config) (*transport, error) {
	tlsConfig := lanTLS
	if tlsConfig == nil {
		tlsConfig = &tls.Config{
			InsecureSkipVerify: true,
			NextProtos:         []string{"pigeon-lan"},
		}
	} else {
		tlsConfig = tlsConfig.Clone()
		tlsConfig.NextProtos = []string{"pigeon-lan"}
	}

	conn, err := quic.DialAddr(ctx, offer.Addr, tlsConfig, &quic.Config{
		EnableDatagrams:      true,
		MaxIdleTimeout:       60 * time.Second,
		KeepAlivePeriod:      10 * time.Second,
		HandshakeIdleTimeout: 10 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("dial: %w", err)
	}

	stream, err := conn.OpenStream()
	if err != nil {
		conn.CloseWithError(1, "open stream failed")
		return nil, fmt.Errorf("open stream: %w", err)
	}

	verify := lanVerify{
		Challenge:  offer.Challenge,
		InstanceID: clientID,
	}
	data, err := json.Marshal(verify)
	if err != nil {
		conn.CloseWithError(1, "marshal verify")
		return nil, err
	}
	if err := writeMessage(stream, data); err != nil {
		conn.CloseWithError(1, "write verify failed")
		return nil, fmt.Errorf("write verify: %w", err)
	}

	resp, err := readMessage(stream)
	if err != nil || string(resp) != "ok" {
		conn.CloseWithError(1, "verify rejected")
		if err != nil {
			return nil, fmt.Errorf("read confirm: %w", err)
		}
		return nil, fmt.Errorf("verify rejected")
	}

	return lanTransport(conn, stream, clientID), nil
}

// lanTransport wraps a verified direct QUIC connection as a Session carrier.
// primary is the handshake stream (already consumed for lanVerify); app
// traffic uses additional OpenStream/AcceptStream/datagrams on conn.
func lanTransport(conn *quic.Conn, primary io.ReadWriteCloser, id string) *transport {
	return &transport{
		primary:    primary,
		opener:     quicOpener{conn},
		acceptor:   quicAcceptor{conn},
		dg:         conn,
		closer:     quicCloser{conn},
		instanceID: id,
	}
}

// swapTransport replaces the Session's carrier with a direct LAN
// connection while preserving AEAD material, stream channels, and
// datagram demux state. Existing sub-streams on the old carrier keep
// their rwc handles until the old connection is closed; new
// OpenStream/AcceptStream/Datagram traffic uses the LAN path.
func (s *Session) swapTransport(tr *transport) {
	if tr == nil {
		return
	}

	s.transportMu.Lock()
	if s.onLAN {
		s.transportMu.Unlock()
		_ = tr.Close()
		return
	}
	old := s.transport
	s.transport = tr
	s.ownsTransport = true
	s.onLAN = true
	s.transportMu.Unlock()

	// Accept and datagram loops on the new carrier. The old loops exit
	// when old.Close() returns AcceptStream/ReceiveDatagram errors.
	go s.acceptLoop()
	go s.datagramLoop()

	s.signalLANReady()

	if old != nil {
		// Brief grace so any in-flight write on the old path can finish
		// before the relay bridge is torn down.
		go func() {
			time.Sleep(50 * time.Millisecond)
			_ = old.Close()
		}()
	}
}

func (s *Session) signalLANReady() {
	s.lanReadyOnce.Do(func() {
		if s.lanReady != nil {
			close(s.lanReady)
		}
	})
}

// currentTransport returns the active carrier under the transport lock.
func (s *Session) currentTransport() *transport {
	s.transportMu.RLock()
	defer s.transportMu.RUnlock()
	return s.transport
}

// LANReady returns a channel that is closed when this Session has
// upgraded to a direct LAN path. If LAN upgrade was not configured on
// this side, the returned channel is already closed (non-blocking).
func (s *Session) LANReady() <-chan struct{} {
	if s.lanReady == nil {
		ch := make(chan struct{})
		close(ch)
		return ch
	}
	return s.lanReady
}

// UsingLAN reports whether the Session's active carrier is the direct
// LAN path (post-upgrade).
func (s *Session) UsingLAN() bool {
	s.transportMu.RLock()
	defer s.transportMu.RUnlock()
	return s.onLAN
}

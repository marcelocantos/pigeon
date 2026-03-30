// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package tern

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math/big"
	"net"
	"time"

	"github.com/marcelocantos/tern/crypto"
	"github.com/marcelocantos/tern/qr"
	"github.com/quic-go/quic-go"
)

// lanOffer is sent via the encrypted relay channel to advertise a
// direct LAN address.
type lanOffer struct {
	Addr      string `json:"addr"`      // host:port of the LAN QUIC listener
	Challenge []byte `json:"challenge"` // 32-byte random challenge for verification
}

// lanVerify is sent on the direct LAN connection to prove the peer
// is the same entity from the relay session.
type lanVerify struct {
	Challenge []byte `json:"challenge"` // echo back the challenge from the offer
	InstanceID string `json:"instance_id"`
}

// EnableLAN enables automatic LAN upgrade on a Conn. When both peers
// are on the same LAN, traffic transparently switches to a direct
// QUIC connection, bypassing the relay.
//
// For the backend (the side that called Register), EnableLAN starts a
// local QUIC listener and advertises the address via the relay. For
// the client (the side that called Connect), EnableLAN watches for
// LAN offers and attempts direct connections.
//
// The tlsConfig must contain a certificate for the LAN listener
// (backend) or trust settings for the LAN connection (client). If nil,
// a self-signed certificate is generated and InsecureSkipVerify is
// used (suitable for development).
//
// Call after SetChannel — LAN offers are sent via the encrypted
// primary stream.
func (c *Conn) EnableLAN(ctx context.Context, isBackend bool, tlsConfig *tls.Config) error {
	if isBackend {
		return c.startLANBackend(ctx, tlsConfig)
	}
	return c.startLANClient(ctx, tlsConfig)
}

// startLANBackend starts a QUIC listener on the LAN and advertises
// it to the peer via the relay channel.
func (c *Conn) startLANBackend(ctx context.Context, tlsConfig *tls.Config) error {
	if tlsConfig == nil {
		cert, err := generateSelfSigned()
		if err != nil {
			return fmt.Errorf("generate LAN cert: %w", err)
		}
		tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
			NextProtos:   []string{"tern-lan"},
		}
	} else {
		tlsConfig = tlsConfig.Clone()
		tlsConfig.NextProtos = []string{"tern-lan"}
	}

	// Listen on all interfaces, random port.
	listener, err := quic.ListenAddr("0.0.0.0:0", tlsConfig, &quic.Config{
		EnableDatagrams: true,
	})
	if err != nil {
		return fmt.Errorf("LAN listen: %w", err)
	}

	addr := listener.Addr().(*net.UDPAddr)
	lanAddr := fmt.Sprintf("%s:%d", qr.LanIP(), addr.Port)

	// Generate a challenge for verifying the direct connection.
	challenge := make([]byte, 32)
	if _, err := rand.Read(challenge); err != nil {
		listener.Close()
		return err
	}

	slog.Info("LAN listener started", "addr", lanAddr)

	// Send the LAN offer via the encrypted relay channel.
	offer := lanOffer{Addr: lanAddr, Challenge: challenge}
	if err := c.sendControl(msgLANOffer, offer); err != nil {
		listener.Close()
		return fmt.Errorf("send LAN offer: %w", err)
	}

	// Accept direct connections in the background.
	go func() {
		defer listener.Close()
		for {
			conn, err := listener.Accept(ctx)
			if err != nil {
				return
			}
			go c.handleLANConnection(conn, challenge)
		}
	}()

	return nil
}

// handleLANConnection verifies a direct LAN connection and, if valid,
// swaps the underlying transport.
func (c *Conn) handleLANConnection(conn *quic.Conn, expectedChallenge []byte) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stream, err := conn.AcceptStream(ctx)
	if err != nil {
		conn.CloseWithError(1, "no stream")
		return
	}

	// Read the verification message.
	data, err := readMessage(stream)
	if err != nil {
		conn.CloseWithError(1, "read verify failed")
		return
	}

	var verify lanVerify
	if err := json.Unmarshal(data, &verify); err != nil {
		conn.CloseWithError(1, "bad verify")
		return
	}

	// Check the challenge matches.
	if len(verify.Challenge) != len(expectedChallenge) {
		conn.CloseWithError(1, "bad challenge")
		return
	}
	for i := range verify.Challenge {
		if verify.Challenge[i] != expectedChallenge[i] {
			conn.CloseWithError(1, "bad challenge")
			return
		}
	}

	slog.Info("LAN connection verified", "peer", verify.InstanceID)

	// Send confirmation.
	if err := writeMessage(stream, []byte("ok")); err != nil {
		conn.CloseWithError(1, "write confirm failed")
		return
	}

	// Swap the transport.
	c.swapTransport(stream, conn, quicCloser{conn}, quicOpener{conn}, quicAcceptor{conn})
	slog.Info("switched to LAN transport", "peer", verify.InstanceID)
}

// startLANClient watches for LAN offers from the peer and attempts
// direct connections.
func (c *Conn) startLANClient(ctx context.Context, tlsConfig *tls.Config) error {
	if tlsConfig == nil {
		tlsConfig = &tls.Config{
			InsecureSkipVerify: true,
			NextProtos:         []string{"tern-lan"},
		}
	} else {
		tlsConfig = tlsConfig.Clone()
		tlsConfig.NextProtos = []string{"tern-lan"}
	}

	c.mu.Lock()
	c.lanTLS = tlsConfig
	c.lanEnabled = true
	c.mu.Unlock()

	return nil
}

// handleLANOffer is called when a LAN offer control message is
// received on the encrypted relay channel.
func (c *Conn) handleLANOffer(offer lanOffer) {
	c.mu.Lock()
	enabled := c.lanEnabled
	tlsConfig := c.lanTLS
	c.mu.Unlock()

	if !enabled {
		return
	}

	slog.Info("received LAN offer", "addr", offer.Addr)

	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		conn, err := quic.DialAddr(ctx, offer.Addr, tlsConfig, &quic.Config{
			EnableDatagrams: true,
		})
		if err != nil {
			slog.Debug("LAN dial failed", "addr", offer.Addr, "err", err)
			return
		}

		stream, err := conn.OpenStream()
		if err != nil {
			conn.CloseWithError(1, "open stream failed")
			return
		}

		// Send verification.
		verify := lanVerify{
			Challenge:  offer.Challenge,
			InstanceID: c.instanceID,
		}
		data, _ := json.Marshal(verify)
		if err := writeMessage(stream, data); err != nil {
			conn.CloseWithError(1, "write verify failed")
			return
		}

		// Wait for confirmation.
		resp, err := readMessage(stream)
		if err != nil || string(resp) != "ok" {
			conn.CloseWithError(1, "verify rejected")
			return
		}

		slog.Info("LAN connection established", "addr", offer.Addr)

		// Swap transport.
		c.swapTransport(stream, conn, quicCloser{conn}, quicOpener{conn}, quicAcceptor{conn})
		slog.Info("switched to LAN transport", "addr", offer.Addr)
	}()
}

// swapTransport atomically replaces the underlying transport of the
// Conn. The old stream and closer are left open — QUIC will clean
// them up when the old connection times out.
//
// The encryption channel's sequence counters are synchronized: the
// send counter carries over (so the peer can verify ordering), and
// the recv side switches to ModeDatagrams to tolerate the sequence
// gap from the transport switch.
func (c *Conn) swapTransport(
	stream io.ReadWriteCloser,
	dg datagrammer,
	closer io.Closer,
	opener streamOpener,
	acceptor streamAcceptor,
) {
	c.mu.Lock()
	oldStream := c.stream
	c.stream = stream
	c.dg = dg
	c.closer = closer
	c.opener = opener
	c.acceptor = acceptor
	// Switch the channel to datagram mode to tolerate the sequence
	// gap caused by the transport switch. The send counter is preserved
	// so the peer can still reject replays.
	if c.channel != nil {
		c.channel.SetMode(crypto.ModeDatagrams)
	}
	if c.dgChannel != nil {
		c.dgChannel.SetMode(crypto.ModeDatagrams)
	}
	c.mu.Unlock()

	// Send a cutover marker on the OLD stream so the relay knows
	// we're done with it. Best-effort — the relay will time out anyway.
	c.sendControlOn(oldStream, msgCutover, nil)
}

// sendControl sends an internal control message on the primary stream.
func (c *Conn) sendControl(msgType byte, payload any) error {
	c.mu.Lock()
	ch := c.channel
	stream := c.stream
	c.mu.Unlock()

	return c.sendControlInner(stream, ch, msgType, payload)
}

// sendControlOn sends a control message on a specific stream.
func (c *Conn) sendControlOn(stream io.ReadWriteCloser, msgType byte, payload any) {
	c.mu.Lock()
	ch := c.channel
	c.mu.Unlock()

	c.sendControlInner(stream, ch, msgType, payload)
}

func (c *Conn) sendControlInner(stream io.ReadWriteCloser, ch interface{ Encrypt([]byte) []byte }, msgType byte, payload any) error {
	var data []byte
	if payload != nil {
		var err error
		data, err = json.Marshal(payload)
		if err != nil {
			return err
		}
	}

	framed := make([]byte, 1+len(data))
	framed[0] = msgType
	copy(framed[1:], data)

	if ch != nil {
		framed = ch.Encrypt(framed)
	}

	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	return writeMessage(stream, framed)
}

// generateSelfSigned creates a self-signed TLS certificate for LAN use.
func generateSelfSigned() (tls.Certificate, error) {
	// Reuse the same approach as cmd/tern but simplified.
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, err
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1)},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		return tls.Certificate{}, err
	}

	return tls.Certificate{
		Certificate: [][]byte{certDER},
		PrivateKey:  key,
	}, nil
}

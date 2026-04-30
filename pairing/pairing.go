// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Package pairing performs the device-pairing ceremony for pigeon.
//
// Pairing is decoupled from connecting: pairing.Register / pairing.Initiate
// produce a *crypto.PairingRecord that callers persist and supply to
// pigeon.Connect / pigeon.Register later.
//
// Acceptor side (long-lived backend that accepts new clients):
//
//	pairer, err := pairing.Register(ctx, &pairing.Args{
//	    Relay:    "https://relay.example.com",
//	    Identity: identity,
//	})
//	defer pairer.Close()
//	for {
//	    cer, err := pairer.Accept(ctx)
//	    if err != nil { return }
//	    go func() {
//	        defer cer.Close()
//	        // display cer.Token, await cer.Code(ctx), prompt user,
//	        // then call cer.Confirm(ctx) — or the deferred Close cancels.
//	    }()
//	}
//
// Initiator side (one-shot, given a token from the acceptor):
//
//	cer, err := pairing.Initiate(ctx, &pairing.InitiateArgs{
//	    Rendezvous: token,
//	    Identity:   identity,
//	})
//	defer cer.Close()
//	code, err := cer.Code(ctx)
//	// ...
package pairing

import (
	"context"
	"crypto/ecdh"
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/marcelocantos/pigeon"
	"github.com/marcelocantos/pigeon/crypto"
)

// Args configures the acceptor side of pairing.
type Args struct {
	Relay    string
	Identity crypto.Identity
}

// InitiateArgs configures the initiator side of pairing.
type InitiateArgs struct {
	Rendezvous string
	Identity   crypto.Identity
}

// Pairer is the long-lived acceptor. Each Accept opens a fresh relay
// session for the next pairing ceremony.
type Pairer struct {
	args   *Args
	mu     sync.Mutex
	closed bool
	ctx    context.Context
	cancel context.CancelFunc
}

// Register creates a Pairer ready to accept new pairings.
func Register(ctx context.Context, args *Args) (*Pairer, error) {
	if args == nil {
		return nil, fmt.Errorf("pairing.Register: args is nil")
	}
	if args.Relay == "" {
		return nil, fmt.Errorf("pairing.Register: Relay is required")
	}
	if args.Identity == nil {
		return nil, fmt.Errorf("pairing.Register: Identity is required")
	}
	pCtx, cancel := context.WithCancel(ctx)
	return &Pairer{args: args, ctx: pCtx, cancel: cancel}, nil
}

// InstanceID returns the backend's stable instance ID.
func (p *Pairer) InstanceID() string { return p.args.Identity.InstanceID() }

// Close releases the Pairer.
func (p *Pairer) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.closed {
		p.closed = true
		p.cancel()
	}
	return nil
}

// Accept blocks until the next initiator arrives. The returned Ceremony's
// Token must be displayed to the user (typically as a QR code) and conveyed
// out-of-band to the initiator.
func (p *Pairer) Accept(ctx context.Context) (*Ceremony, error) {
	eph, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("ephemeral keygen: %w", err)
	}

	cfg := pigeon.Config{TLS: &tls.Config{InsecureSkipVerify: true}}
	conn, err := pigeon.DialRelayAcceptor(ctx, p.args.Relay, cfg)
	if err != nil {
		return nil, fmt.Errorf("relay register: %w", err)
	}

	payload := tokenPayload{
		Relay:            p.args.Relay,
		SessionID:        conn.InstanceID(),
		AcceptorEphPub:   eph.PublicKey().Bytes(),
		AcceptorIdPub:    p.args.Identity.PublicKey(),
		AcceptorInstance: p.args.Identity.InstanceID(),
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("marshal token: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(raw)

	cer := newCeremony(token, conn)
	go runAcceptor(p.ctx, conn, eph, p.args.Identity, p.args.Relay, cer)
	return cer, nil
}

// Initiate runs the initiator side against the rendezvous token from
// the acceptor's Pairer.Accept.
func Initiate(ctx context.Context, args *InitiateArgs) (*Ceremony, error) {
	if args == nil {
		return nil, fmt.Errorf("pairing.Initiate: args is nil")
	}
	if args.Rendezvous == "" {
		return nil, fmt.Errorf("pairing.Initiate: Rendezvous token is required")
	}
	if args.Identity == nil {
		return nil, fmt.Errorf("pairing.Initiate: Identity is required")
	}

	raw, err := base64.RawURLEncoding.DecodeString(args.Rendezvous)
	if err != nil {
		return nil, fmt.Errorf("decode token: %w", err)
	}
	var payload tokenPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, fmt.Errorf("parse token: %w", err)
	}

	eph, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("ephemeral keygen: %w", err)
	}

	cfg := pigeon.Config{TLS: &tls.Config{InsecureSkipVerify: true}}
	conn, err := pigeon.DialRelayInitiator(ctx, payload.Relay, payload.SessionID, cfg)
	if err != nil {
		return nil, fmt.Errorf("relay connect: %w", err)
	}

	cer := newCeremony("", conn)
	go runInitiator(context.Background(), conn, eph, args.Identity, &payload, cer)
	return cer, nil
}

// tokenPayload is the rendezvous token's contents (base64-encoded).
type tokenPayload struct {
	Relay            string `json:"relay"`
	SessionID        string `json:"session_id"`
	AcceptorEphPub   []byte `json:"acceptor_eph_pub"`
	AcceptorIdPub    []byte `json:"acceptor_id_pub"`
	AcceptorInstance string `json:"acceptor_instance"`
}

// pairingMessage flows over the relay session during the ceremony.
type pairingMessage struct {
	Kind        string `json:"kind"`
	EphPub      []byte `json:"eph_pub,omitempty"`
	IdentityPub []byte `json:"identity_pub,omitempty"`
	InstanceID  string `json:"instance_id,omitempty"`
}

const (
	msgHello   = "hello"
	msgWelcome = "welcome"
	msgConfirm = "confirm"
)

// Ceremony is a single in-flight pairing exchange.
//
// Always defer Close. Pre-Confirm Close cancels (peer sees abort);
// post-Confirm Close is no-op cleanup. Idempotent.
type Ceremony struct {
	// Token is the rendezvous payload to display (acceptor-side only).
	// Empty on the initiator side.
	Token string

	mu        sync.Mutex
	closed    bool
	conn      *pigeon.Conn
	codeReady chan struct{}
	codeOnce  sync.Once
	code      string
	codeErr   error

	confirmCh chan struct{}
	confirmed bool

	result chan ceremonyResult
}

type ceremonyResult struct {
	rec *crypto.PairingRecord
	err error
}

func newCeremony(token string, conn *pigeon.Conn) *Ceremony {
	return &Ceremony{
		Token:     token,
		conn:      conn,
		codeReady: make(chan struct{}),
		confirmCh: make(chan struct{}),
		result:    make(chan ceremonyResult, 1),
	}
}

// Code blocks until the 6-digit confirmation code is derivable, then
// returns it. Call from the application's UI to display the code so the
// user can compare it against the peer's screen.
func (c *Ceremony) Code(ctx context.Context) (string, error) {
	select {
	case <-c.codeReady:
		return c.code, c.codeErr
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

// Confirm signals "the user said the codes match." Blocks until the peer
// also confirms (or cancels / times out). On success returns the new
// PairingRecord.
func (c *Ceremony) Confirm(ctx context.Context) (*crypto.PairingRecord, error) {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil, fmt.Errorf("ceremony closed")
	}
	if !c.confirmed {
		c.confirmed = true
		close(c.confirmCh)
	}
	c.mu.Unlock()

	select {
	case res := <-c.result:
		return res.rec, res.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// Close releases the Ceremony's resources. Safe to call any time;
// idempotent.
func (c *Ceremony) Close() error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return nil
	}
	c.closed = true
	conn := c.conn
	c.mu.Unlock()
	if conn != nil {
		_ = conn.Close()
	}
	return nil
}

func (c *Ceremony) signalCode(code string, err error) {
	c.codeOnce.Do(func() {
		c.code = code
		c.codeErr = err
		close(c.codeReady)
	})
}

func (c *Ceremony) deliverResult(rec *crypto.PairingRecord, err error) {
	if err != nil {
		c.signalCode("", err)
	}
	select {
	case c.result <- ceremonyResult{rec: rec, err: err}:
	default:
	}
}

// runAcceptor drives the acceptor side of the ceremony.
func runAcceptor(ctx context.Context, conn *pigeon.Conn, eph *ecdh.PrivateKey, identity crypto.Identity, relayURL string, cer *Ceremony) {
	// Read hello.
	raw, err := conn.Recv(ctx)
	if err != nil {
		cer.deliverResult(nil, fmt.Errorf("recv hello: %w", err))
		return
	}
	var hello pairingMessage
	if err := json.Unmarshal(raw, &hello); err != nil || hello.Kind != msgHello {
		cer.deliverResult(nil, fmt.Errorf("bad hello: %v", err))
		return
	}
	initEphPub, err := ecdh.X25519().NewPublicKey(hello.EphPub)
	if err != nil {
		cer.deliverResult(nil, fmt.Errorf("parse initiator eph: %w", err))
		return
	}

	// Send welcome.
	welcome := pairingMessage{
		Kind:        msgWelcome,
		EphPub:      eph.PublicKey().Bytes(),
		IdentityPub: identity.PublicKey(),
		InstanceID:  identity.InstanceID(),
	}
	welcomeBytes, _ := json.Marshal(welcome)
	if err := conn.Send(ctx, welcomeBytes); err != nil {
		cer.deliverResult(nil, fmt.Errorf("send welcome: %w", err))
		return
	}

	// Compute confirmation code.
	code, err := crypto.DeriveConfirmationCode(eph.PublicKey(), initEphPub)
	if err != nil {
		cer.deliverResult(nil, fmt.Errorf("derive code: %w", err))
		return
	}
	cer.signalCode(code, nil)

	if err := exchangeConfirm(ctx, conn, cer); err != nil {
		cer.deliverResult(nil, err)
		return
	}

	rec := crypto.NewPairingRecord(
		hello.InstanceID,
		relayURL,
		&crypto.KeyPair{Private: eph, Public: eph.PublicKey()},
		initEphPub,
	)
	cer.deliverResult(rec, nil)
}

// runInitiator drives the initiator side.
func runInitiator(ctx context.Context, conn *pigeon.Conn, eph *ecdh.PrivateKey, identity crypto.Identity, payload *tokenPayload, cer *Ceremony) {
	// Send hello.
	hello := pairingMessage{
		Kind:        msgHello,
		EphPub:      eph.PublicKey().Bytes(),
		IdentityPub: identity.PublicKey(),
		InstanceID:  identity.InstanceID(),
	}
	helloBytes, _ := json.Marshal(hello)
	if err := conn.Send(ctx, helloBytes); err != nil {
		cer.deliverResult(nil, fmt.Errorf("send hello: %w", err))
		return
	}

	// Read welcome.
	raw, err := conn.Recv(ctx)
	if err != nil {
		cer.deliverResult(nil, fmt.Errorf("recv welcome: %w", err))
		return
	}
	var welcome pairingMessage
	if err := json.Unmarshal(raw, &welcome); err != nil || welcome.Kind != msgWelcome {
		cer.deliverResult(nil, fmt.Errorf("bad welcome"))
		return
	}
	accEphPub, err := ecdh.X25519().NewPublicKey(welcome.EphPub)
	if err != nil {
		cer.deliverResult(nil, fmt.Errorf("parse acc eph: %w", err))
		return
	}

	code, err := crypto.DeriveConfirmationCode(accEphPub, eph.PublicKey())
	if err != nil {
		cer.deliverResult(nil, fmt.Errorf("derive code: %w", err))
		return
	}
	cer.signalCode(code, nil)

	if err := exchangeConfirm(ctx, conn, cer); err != nil {
		cer.deliverResult(nil, err)
		return
	}

	rec := crypto.NewPairingRecord(
		welcome.InstanceID,
		payload.Relay,
		&crypto.KeyPair{Private: eph, Public: eph.PublicKey()},
		accEphPub,
	)
	cer.deliverResult(rec, nil)
}

// exchangeConfirm waits for local Confirm, sends a confirm message to the
// peer, and waits for the peer's confirm message.
func exchangeConfirm(ctx context.Context, conn *pigeon.Conn, cer *Ceremony) error {
	peerCh := make(chan error, 1)
	go func() {
		raw, err := conn.Recv(ctx)
		if err != nil {
			peerCh <- err
			return
		}
		var msg pairingMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			peerCh <- fmt.Errorf("bad peer message: %w", err)
			return
		}
		if msg.Kind != msgConfirm {
			peerCh <- fmt.Errorf("unexpected peer kind: %s", msg.Kind)
			return
		}
		peerCh <- nil
	}()

	select {
	case <-cer.confirmCh:
	case <-ctx.Done():
		return ctx.Err()
	}

	confirmMsg := pairingMessage{Kind: msgConfirm}
	confirmBytes, _ := json.Marshal(confirmMsg)
	if err := conn.Send(ctx, confirmBytes); err != nil {
		return fmt.Errorf("send confirm: %w", err)
	}

	select {
	case err := <-peerCh:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

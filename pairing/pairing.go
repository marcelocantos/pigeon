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
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/marcelocantos/pigeon"
	"github.com/marcelocantos/pigeon/crypto"
	"github.com/marcelocantos/pigeon/cwire"
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
//
// Each Accept call registers a fresh listener with the relay so the QR
// token's SessionID is unique per ceremony. The acceptor then waits on
// listener.Accept for the matching initiator's Connect to land. This runs
// in pigeon.Register's pairing-mode (no Pairing callback supplied) per
// docs/DESIGN.md §3 L2 — the resulting Session is a raw bytes pipe with
// no AEAD context, since the ceremony establishes its own crypto.
func (p *Pairer) Accept(ctx context.Context) (*Ceremony, error) {
	eph, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("ephemeral keygen: %w", err)
	}

	listener, instanceID, err := pigeon.Register(ctx, &pigeon.RegisterArgs{
		Relay: p.args.Relay,
		TLS:   &tls.Config{InsecureSkipVerify: true},
	})
	if err != nil {
		return nil, fmt.Errorf("relay register: %w", err)
	}

	payload := tokenPayload{
		Relay:            p.args.Relay,
		SessionID:        instanceID,
		AcceptorEphPub:   eph.PublicKey().Bytes(),
		AcceptorIdPub:    p.args.Identity.PublicKey(),
		AcceptorInstance: p.args.Identity.InstanceID(),
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		_ = listener.Close()
		return nil, fmt.Errorf("marshal token: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(raw)

	// Root the ceremony ctx at the Pairer ctx so Pairer.Close still
	// aborts every in-flight ceremony; cer.Close additionally cancels
	// just this one.
	cer := newCeremony(p.ctx, token, nil)
	cer.listener = listener
	go runAcceptor(cer.ctx, listener, eph, p.args.Identity, p.args.Relay, cer)
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

	sess, err := pigeon.Connect(ctx, &pigeon.ConnectArgs{
		InstanceID: payload.SessionID,
		Relay:      payload.Relay,
		TLS:        &tls.Config{InsecureSkipVerify: true},
	})
	if err != nil {
		return nil, fmt.Errorf("relay connect: %w", err)
	}

	// Root the ceremony ctx at the caller's ctx and cancel it on Close,
	// so a pre-confirm Close (or a cancelled caller ctx) aborts the
	// C-side confirm wait instead of wedging runInitiator forever.
	cer := newCeremony(ctx, "", sess)
	go runInitiator(cer.ctx, sess, eph, args.Identity, &payload, cer)
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

// Ceremony is a single in-flight pairing exchange.
//
// Always defer Close. Pre-Confirm Close cancels (peer sees abort);
// post-Confirm Close is no-op cleanup. Idempotent.
type Ceremony struct {
	// Token is the rendezvous payload to display (acceptor-side only).
	// Empty on the initiator side.
	Token string

	// ctx scopes the in-flight ceremony (the C driver's confirm wait
	// selects on ctx.Done()). Close cancels it so a pre-confirm Close
	// aborts the ceremony instead of leaking its goroutine + cgo handles.
	ctx    context.Context
	cancel context.CancelFunc

	mu        sync.Mutex
	closed    bool
	sess      *pigeon.Session  // initiator side; nil on acceptor pre-Accept
	listener  *pigeon.Listener // acceptor side; nil on initiator
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

func newCeremony(parent context.Context, token string, sess *pigeon.Session) *Ceremony {
	ctx, cancel := context.WithCancel(parent)
	return &Ceremony{
		Token:     token,
		sess:      sess,
		ctx:       ctx,
		cancel:    cancel,
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
	sess := c.sess
	listener := c.listener
	c.mu.Unlock()
	// Cancel first so the C driver's confirm callback (if parked) returns
	// and runInitiator/runAcceptor unwind, releasing their cgo handles
	// and QUIC session before we close the transport underneath them.
	c.cancel()
	if sess != nil {
		_ = sess.Close()
	}
	if listener != nil {
		_ = listener.Close()
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

// ceremonyStreamName is the named pigeon Stream the pairing ceremony
// rides on. Both peers AcceptStream / OpenStream this name to obtain
// the bidirectional bytes pipe the JSON wire exchange flows over.
const ceremonyStreamName = "ceremony"

// runAcceptor drives the acceptor side of the ceremony. State
// transitions are dispatched through the protogen-generated executor
// (NewPairingCeremonyProtocolAcceptorMachine), so the Go code path is
// provably equivalent to the Swift / Kotlin / TS / C / TLA+ outputs
// generated from the same protocol/pairing.yaml.
func runAcceptor(ctx context.Context, listener *pigeon.Listener, eph *ecdh.PrivateKey, identity crypto.Identity, relayURL string, cer *Ceremony) {
	sess, err := listener.Accept(ctx)
	if err != nil {
		cer.deliverResult(nil, fmt.Errorf("accept ceremony client: %w", err))
		return
	}
	cer.mu.Lock()
	cer.sess = sess
	cer.mu.Unlock()
	stream, err := sess.AcceptStream(ctx, ceremonyStreamName)
	if err != nil {
		cer.deliverResult(nil, fmt.Errorf("accept ceremony stream: %w", err))
		return
	}
	defer stream.Close()
	runAcceptorOnStream(ctx, stream, eph, identity, relayURL, cer)
}

// runAcceptorOnStream drives the acceptor side of the ceremony through
// libpigeon's pigeon_pair_acceptor via cwire.RunAcceptor. The wire
// exchange (hello / welcome / confirm) runs in C; the Go side wraps the
// already-opened pigeon.Stream as a single-stream GoTransport, supplies
// the ephemeral / identity material, and translates the cwire callback
// into the Ceremony's Code() / Confirm() rendezvous.
func runAcceptorOnStream(ctx context.Context, stream *pigeon.Stream, eph *ecdh.PrivateKey, identity crypto.Identity, relayURL string, cer *Ceremony) {
	rwc := stream.UnderlyingPairingRWC()
	if rwc == nil {
		cer.deliverResult(nil, errors.New("acceptor: stream not in pairing mode"))
		return
	}
	ref := cwire.NewGoTransportRef(&singleStreamTransport{rwc: rwc})
	defer ref.Close()

	rec, _, err := cwire.RunAcceptor(&cwire.RunAcceptorArgs{
		Ref:          ref,
		LocalEphPriv: eph.Bytes(),
		LocalEphPub:  eph.PublicKey().Bytes(),
		IdentityPub:  identity.PublicKey(),
		InstanceID:   identity.InstanceID(),
		ConfirmFn: func(code string) bool {
			cer.signalCode(code, nil)
			return waitConfirm(ctx, cer)
		},
	})
	if err != nil {
		cer.deliverResult(nil, fmt.Errorf("acceptor: %w", err))
		return
	}

	// 🎯T57: do not deliverResult (and thus unblock Confirm → caller Close)
	// until the initiator has finished draining our final confirm and closed
	// its stream end. The relay bridge is asynchronous: acceptor's pair_send
	// of confirm can still be in-flight toward the initiator when C returns.
	// If Confirm returns and the caller's defer Close tears down the backend
	// Session, the bridge aborts and the initiator's pair_recv fails with
	// cwire: pigeon_pair_initiator failed. Waiting for peer FIN makes Close
	// safe. Bounded by ctx so a stuck peer still aborts.
	drainPeerStream(ctx, rwc)

	cer.deliverResult(&crypto.PairingRecord{
		PeerInstanceID:  rec.PeerInstanceID,
		RelayURL:        relayURL,
		LocalPrivateKey: rec.LocalPrivKey,
		LocalPublicKey:  rec.LocalPubKey,
		PeerPublicKey:   rec.PeerPubKey,
	}, nil)
}

// runInitiator drives the initiator side via the generated executor.
func runInitiator(ctx context.Context, sess *pigeon.Session, eph *ecdh.PrivateKey, identity crypto.Identity, payload *tokenPayload, cer *Ceremony) {
	stream, err := sess.OpenStream(ctx, ceremonyStreamName)
	if err != nil {
		cer.deliverResult(nil, fmt.Errorf("open ceremony stream: %w", err))
		return
	}
	// Close the stream before deliverResult so the acceptor's post-success
	// drain (🎯T57) observes FIN and unblocks before either side's Confirm
	// returns — ordering that keeps Ceremony.Close from racing the peer.
	defer stream.Close()
	runInitiatorOnStream(ctx, stream, eph, identity, payload, cer)
}

// runInitiatorOnStream drives the initiator side of the ceremony
// through libpigeon's pigeon_pair_initiator via cwire.RunInitiator,
// mirroring runAcceptorOnStream.
func runInitiatorOnStream(ctx context.Context, stream *pigeon.Stream, eph *ecdh.PrivateKey, identity crypto.Identity, payload *tokenPayload, cer *Ceremony) {
	rwc := stream.UnderlyingPairingRWC()
	if rwc == nil {
		cer.deliverResult(nil, errors.New("initiator: stream not in pairing mode"))
		return
	}
	ref := cwire.NewGoTransportRef(&singleStreamTransport{rwc: rwc})
	defer ref.Close()

	rec, _, err := cwire.RunInitiator(&cwire.RunInitiatorArgs{
		Ref:          ref,
		LocalEphPriv: eph.Bytes(),
		LocalEphPub:  eph.PublicKey().Bytes(),
		IdentityPub:  identity.PublicKey(),
		InstanceID:   identity.InstanceID(),
		AccEphPub:    payload.AcceptorEphPub,
		AccInstance:  payload.AcceptorInstance,
		ConfirmFn: func(code string) bool {
			cer.signalCode(code, nil)
			return waitConfirm(ctx, cer)
		},
	})
	// FIN the ceremony stream before delivering the result so the acceptor
	// drainPeerStream unblocks prior to either Confirm returning (🎯T57).
	_ = stream.Close()
	if err != nil {
		cer.deliverResult(nil, fmt.Errorf("initiator: %w", err))
		return
	}

	cer.deliverResult(&crypto.PairingRecord{
		PeerInstanceID:  rec.PeerInstanceID,
		RelayURL:        payload.Relay,
		LocalPrivateKey: rec.LocalPrivKey,
		LocalPublicKey:  rec.LocalPubKey,
		PeerPublicKey:   rec.PeerPubKey,
	}, nil)
}

// waitConfirm blocks until the user confirms (confirmCh closed) or ctx is
// cancelled. When both are ready, prefer confirm over cancel so a Close that
// races with Confirm does not spuriously abort a successful ceremony
// (T54 ctx-threading made this select reachable on Close; the flake was
// primarily the post-success Session teardown race fixed by drainPeerStream,
// but preferring confirmCh keeps Close+Confirm ordering deterministic).
func waitConfirm(ctx context.Context, cer *Ceremony) bool {
	select {
	case <-cer.confirmCh:
		return true
	default:
	}
	select {
	case <-cer.confirmCh:
		return true
	case <-ctx.Done():
		select {
		case <-cer.confirmCh:
			return true
		default:
			return false
		}
	}
}

// drainPeerStream reads until EOF or ctx cancel. Used after a successful
// acceptor pairing so Confirm does not return before the initiator has
// closed its stream end (proof it finished pair_recv of our confirm).
func drainPeerStream(ctx context.Context, r io.Reader) {
	done := make(chan struct{})
	go func() {
		defer close(done)
		_, _ = io.Copy(io.Discard, r)
	}()
	select {
	case <-done:
	case <-ctx.Done():
	}
}

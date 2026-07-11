// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package cwire_test

import (
	"bytes"
	"crypto/ecdh"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"testing"
	"unsafe"

	"github.com/marcelocantos/pigeon/crypto"
	"github.com/marcelocantos/pigeon/cwire"
	"github.com/marcelocantos/pigeon/pairing"
)

// TestAcceptorFSMMatchesGo drives both the C-side acceptor FSM (via
// cgo) and the Go-side acceptor FSM (pairing.NewPairingCeremonyProtocolAcceptorMachine)
// through the same happy-path event sequence. After every step the
// states must agree. Cross-language equivalence by construction.
func TestAcceptorFSMMatchesGo(t *testing.T) {
	cm := cwire.NewAcceptorMachine()
	gm := pairing.NewPairingCeremonyProtocolAcceptorMachine()

	steps := []struct {
		name string
		fire func() error
	}{
		{"PairBegin", func() error {
			ge := cm.Step(cwire.EvPairBegin)
			_, err := gm.Step(pairing.PairingCeremonyProtocolEventPairBegin)
			if err != nil {
				return err
			}
			return ge
		}},
		{"EphemeralReady", func() error {
			ge := cm.Step(cwire.EvEphemeralReady)
			_, err := gm.Step(pairing.PairingCeremonyProtocolEventEphemeralReady)
			if err != nil {
				return err
			}
			return ge
		}},
		{"RelayRegistered", func() error {
			ge := cm.Step(cwire.EvRelayRegistered)
			_, err := gm.Step(pairing.PairingCeremonyProtocolEventRelayRegistered)
			if err != nil {
				return err
			}
			return ge
		}},
		{"recv Hello", func() error {
			ge := cm.HandleMessage(cwire.MsgHello)
			_, err := gm.HandleMessage(pairing.PairingCeremonyProtocolMsgHello)
			if err != nil {
				return err
			}
			return ge
		}},
		{"recv Reveal", func() error {
			ge := cm.HandleMessage(cwire.MsgReveal)
			_, err := gm.HandleMessage(pairing.PairingCeremonyProtocolMsgReveal)
			if err != nil {
				return err
			}
			return ge
		}},
		{"CodeReady", func() error {
			ge := cm.Step(cwire.EvCodeReady)
			_, err := gm.Step(pairing.PairingCeremonyProtocolEventCodeReady)
			if err != nil {
				return err
			}
			return ge
		}},
		{"UserConfirm", func() error {
			ge := cm.Step(cwire.EvUserConfirm)
			_, err := gm.Step(pairing.PairingCeremonyProtocolEventUserConfirm)
			if err != nil {
				return err
			}
			return ge
		}},
		{"recv ConfirmToAcceptor", func() error {
			ge := cm.HandleMessage(cwire.MsgConfirmToAcceptor)
			_, err := gm.HandleMessage(pairing.PairingCeremonyProtocolMsgConfirmToAcceptor)
			if err != nil {
				return err
			}
			return ge
		}},
	}

	for _, s := range steps {
		if err := s.fire(); err != nil {
			t.Fatalf("%s: %v", s.name, err)
		}
		if cm.State().String() != string(gm.State) {
			t.Fatalf("%s: C state %q != Go state %q", s.name, cm.State(), gm.State)
		}
	}
	if cm.State() != cwire.AcceptorPaired {
		t.Fatalf("expected AcceptorPaired, got %v", cm.State())
	}
}

// TestInitiatorFSMMatchesGo is the symmetric drive for the initiator FSM.
func TestInitiatorFSMMatchesGo(t *testing.T) {
	cm := cwire.NewInitiatorMachine()
	gm := pairing.NewPairingCeremonyProtocolInitiatorMachine()

	steps := []struct {
		name string
		fire func() error
	}{
		{"TokenReceived", func() error {
			ge := cm.Step(cwire.EvTokenReceived)
			_, err := gm.Step(pairing.PairingCeremonyProtocolEventTokenReceived)
			if err != nil {
				return err
			}
			return ge
		}},
		{"TokenDecoded", func() error {
			ge := cm.Step(cwire.EvTokenDecoded)
			_, err := gm.Step(pairing.PairingCeremonyProtocolEventTokenDecoded)
			if err != nil {
				return err
			}
			return ge
		}},
		{"EphemeralReady", func() error {
			ge := cm.Step(cwire.EvEphemeralReady)
			_, err := gm.Step(pairing.PairingCeremonyProtocolEventEphemeralReady)
			if err != nil {
				return err
			}
			return ge
		}},
		{"RelayConnected", func() error {
			ge := cm.Step(cwire.EvRelayConnected)
			_, err := gm.Step(pairing.PairingCeremonyProtocolEventRelayConnected)
			if err != nil {
				return err
			}
			return ge
		}},
		{"recv Welcome", func() error {
			ge := cm.HandleMessage(cwire.MsgWelcome)
			_, err := gm.HandleMessage(pairing.PairingCeremonyProtocolMsgWelcome)
			if err != nil {
				return err
			}
			return ge
		}},
		{"RevealSent", func() error {
			ge := cm.Step(cwire.EvRevealSent)
			_, err := gm.Step(pairing.PairingCeremonyProtocolEventRevealSent)
			if err != nil {
				return err
			}
			return ge
		}},
		{"CodeReady", func() error {
			ge := cm.Step(cwire.EvCodeReady)
			_, err := gm.Step(pairing.PairingCeremonyProtocolEventCodeReady)
			if err != nil {
				return err
			}
			return ge
		}},
		{"UserConfirm", func() error {
			ge := cm.Step(cwire.EvUserConfirm)
			_, err := gm.Step(pairing.PairingCeremonyProtocolEventUserConfirm)
			if err != nil {
				return err
			}
			return ge
		}},
		{"recv ConfirmToInitiator", func() error {
			ge := cm.HandleMessage(cwire.MsgConfirmToInitiator)
			_, err := gm.HandleMessage(pairing.PairingCeremonyProtocolMsgConfirmToInitiator)
			if err != nil {
				return err
			}
			return ge
		}},
	}

	for _, s := range steps {
		if err := s.fire(); err != nil {
			t.Fatalf("%s: %v", s.name, err)
		}
		if cm.State().String() != string(gm.State) {
			t.Fatalf("%s: C state %q != Go state %q", s.name, cm.State(), gm.State)
		}
	}
	if cm.State() != cwire.InitiatorPaired {
		t.Fatalf("expected InitiatorPaired, got %v", cm.State())
	}
}

// TestConfirmationCodeMatchesGo proves the C
// pigeon_derive_confirmation_code emits the same 6-digit code as the
// Go crypto.DeriveConfirmationCode for the same pair of ephemeral
// public keys. This is the specific test the cmd/crypto-peer ↔ Swift
// E2E exercises across-process; here we exercise it within one
// process via cgo.
func TestConfirmationCodeMatchesGo(t *testing.T) {
	a, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	b, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}

	pubA := a.PublicKey().Bytes()
	pubB := b.PublicKey().Bytes()

	cCode, err := cwire.DeriveConfirmationCode(pubA, pubB)
	if err != nil {
		t.Fatal(err)
	}
	goCode, err := crypto.DeriveConfirmationCode(a.PublicKey(), b.PublicKey())
	if err != nil {
		t.Fatal(err)
	}
	if cCode != goCode {
		t.Fatalf("code mismatch: C=%q Go=%q", cCode, goCode)
	}
	if len(cCode) != 6 {
		t.Fatalf("unexpected code length %d (want 6)", len(cCode))
	}

	// Order-independence: derive(B, A) must equal derive(A, B).
	swap, err := cwire.DeriveConfirmationCode(pubB, pubA)
	if err != nil {
		t.Fatal(err)
	}
	if swap != cCode {
		t.Fatalf("order dependence: derive(A,B)=%q derive(B,A)=%q", cCode, swap)
	}
}

// ---------------------------------------------------------------------------
// Wire-parity tests
//
// Each sub-test pairs one Go-side driver with one C-side driver over an
// in-process blocking channel transport. The C driver is exercised via
// cwire.RunAcceptor / cwire.RunInitiator; the Go driver is a minimal inline
// hello/welcome/confirm loop. Both sides block-wait on each other, avoiding
// the ordering hazard of the non-blocking pipeTransport.
//
// Acceptance criterion: both sides derive the same 6-digit confirmation code.
// ---------------------------------------------------------------------------

// pairingMsg is the JSON structure the Go/C pairing drivers send on the wire
// after 🎯T52 (commit-then-reveal).
type pairingMsg struct {
	Kind        string `json:"kind"`
	Commit      []byte `json:"commit,omitempty"`
	EphPub      []byte `json:"eph_pub,omitempty"`
	Blind       []byte `json:"blind,omitempty"`
	IdentityPub []byte `json:"identity_pub,omitempty"`
	InstanceID  string `json:"instance_id,omitempty"`
}

// newEphKey generates a fresh X25519 ephemeral key pair, fatal on error.
func newEphKey(t *testing.T) (*ecdh.PrivateKey, *ecdh.PublicKey) {
	t.Helper()
	k, err := ecdh.X25519().GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("keygen: %v", err)
	}
	return k, k.PublicKey()
}

// chanTransport is a simple blocking in-memory transport for pairing tests.
// Two endpoints (A and B) share one bidirectional channel wire. Each endpoint
// has a dedicated send and recv channel so messages from A→B flow through
// aToB and messages from B→A flow through bToA. Stream opens are signalled
// on a shared acceptCh; the opener sends a handle and the acceptor reads it.
type chanTransport struct {
	send     chan []byte         // messages I write (lands in peer's recv)
	recv     chan []byte         // messages I read (from peer's send)
	acceptCh chan unsafe.Pointer // I write here on OpenStream; peer reads on AcceptStream
	// acceptRecv is the channel I read from on AcceptStream.
	acceptRecv chan unsafe.Pointer
}

// newChanTransportPair creates two paired chanTransports. A opens streams
// (sends handles to B's acceptRecv); B opens streams (sends handles to A's
// acceptRecv).
func newChanTransportPair() (a, b *chanTransport) {
	aToB := make(chan []byte, 64)
	bToA := make(chan []byte, 64)
	aAccepts := make(chan unsafe.Pointer, 8) // B opens → A accepts
	bAccepts := make(chan unsafe.Pointer, 8) // A opens → B accepts

	a = &chanTransport{
		send:       aToB,
		recv:       bToA,
		acceptCh:   bAccepts, // A.OpenStream puts handles in bAccepts (B reads)
		acceptRecv: aAccepts, // A.AcceptStream reads from aAccepts (B puts there)
	}
	b = &chanTransport{
		send:       bToA,
		recv:       aToB,
		acceptCh:   aAccepts, // B.OpenStream puts handles in aAccepts (A reads)
		acceptRecv: bAccepts, // B.AcceptStream reads from bAccepts (A puts there)
	}
	return
}

// chanHandle is an opaque pointer per stream. One global store, address is stable.
var chanHandleStore [256]byte
var chanHandleIdx int

func mintChanHandle() unsafe.Pointer {
	i := chanHandleIdx
	chanHandleIdx++
	if chanHandleIdx >= len(chanHandleStore) {
		chanHandleIdx = 0
	}
	return unsafe.Pointer(&chanHandleStore[i])
}

func (t *chanTransport) OpenStream() (unsafe.Pointer, error) {
	h := mintChanHandle()
	t.acceptCh <- h
	return h, nil
}

func (t *chanTransport) AcceptStream() (unsafe.Pointer, error) {
	return <-t.acceptRecv, nil
}

func (t *chanTransport) SendOnStream(_ unsafe.Pointer, msg []byte) error {
	cp := make([]byte, len(msg))
	copy(cp, msg)
	t.send <- cp
	return nil
}

func (t *chanTransport) RecvOnStream(_ unsafe.Pointer) ([]byte, error) {
	return <-t.recv, nil
}

func (t *chanTransport) CloseStream(_ unsafe.Pointer) error { return nil }
func (t *chanTransport) SendDatagram(payload []byte) error  { return nil }
func (t *chanTransport) RecvDatagram() ([]byte, error)      { return nil, nil }

// TestWireParityGoAcceptorCInitiator: Go acceptor ↔ C initiator.
//
// The C initiator (RunInitiator) calls transport->open_stream internally and
// then runs hello→welcome→confirm. The Go acceptor blocks on AcceptStream
// (via the channel transport) and runs the ceremony in parallel. Both sides
// must derive the same confirmation code.
func TestWireParityGoAcceptorCInitiator(t *testing.T) {
	t.Parallel()

	_, accEphPub := newEphKey(t)
	initEphPriv, initEphPub := newEphKey(t)

	accIdPub := make([]byte, 32)
	initIdPub := make([]byte, 32)

	// ca = acceptor side (Go), cb = initiator side (C).
	ca, cb := newChanTransportPair()
	refB := cwire.NewGoTransportRef(cb)
	defer refB.Close()

	type result struct {
		code string
		err  error
	}
	cCh := make(chan result, 1)
	goCh := make(chan result, 1)

	// C initiator goroutine.
	go func() {
		_, code, err := cwire.RunInitiator(&cwire.RunInitiatorArgs{
			Ref:          refB,
			LocalEphPriv: initEphPriv.Bytes(),
			LocalEphPub:  initEphPub.Bytes(),
			IdentityPub:  initIdPub,
			InstanceID:   "c-initiator",
			AccEphPub:    accEphPub.Bytes(),
			AccInstance:  "go-acceptor",
			ConfirmFn:    func(code string) bool { return true },
		})
		cCh <- result{code: code, err: err}
	}()

	// Go acceptor goroutine: blocks on AcceptStream; commit/reveal/confirm.
	go func() {
		h, err := ca.AcceptStream()
		if err != nil {
			goCh <- result{err: err}
			return
		}
		raw, err := ca.RecvOnStream(h)
		if err != nil {
			goCh <- result{err: err}
			return
		}
		var hello pairingMsg
		if err := json.Unmarshal(raw, &hello); err != nil || hello.Kind != "hello" {
			goCh <- result{err: fmt.Errorf("hello: %w", err)}
			return
		}
		if len(hello.Commit) != 32 {
			goCh <- result{err: fmt.Errorf("hello missing 32-byte commit")}
			return
		}
		wb, _ := json.Marshal(pairingMsg{Kind: "welcome", EphPub: accEphPub.Bytes(), IdentityPub: accIdPub, InstanceID: "go-acceptor"})
		if err := ca.SendOnStream(h, wb); err != nil {
			goCh <- result{err: err}
			return
		}
		raw, err = ca.RecvOnStream(h)
		if err != nil {
			goCh <- result{err: err}
			return
		}
		var reveal pairingMsg
		if err := json.Unmarshal(raw, &reveal); err != nil || reveal.Kind != "reveal" {
			goCh <- result{err: fmt.Errorf("reveal: %w", err)}
			return
		}
		if !crypto.SASCommitVerify(reveal.EphPub, reveal.Blind, hello.Commit) {
			goCh <- result{err: fmt.Errorf("reveal does not open commit")}
			return
		}
		initEphPubParsed, err := ecdh.X25519().NewPublicKey(reveal.EphPub)
		if err != nil {
			goCh <- result{err: err}
			return
		}
		code, err := crypto.DeriveConfirmationCode(accEphPub, initEphPubParsed)
		if err != nil {
			goCh <- result{err: err}
			return
		}
		cb, _ := json.Marshal(pairingMsg{Kind: "confirm"})
		if err := ca.SendOnStream(h, cb); err != nil {
			goCh <- result{err: err}
			return
		}
		raw, err = ca.RecvOnStream(h)
		if err != nil {
			goCh <- result{err: err}
			return
		}
		var conf pairingMsg
		if err := json.Unmarshal(raw, &conf); err != nil || conf.Kind != "confirm" {
			goCh <- result{err: fmt.Errorf("confirm: %w", err)}
			return
		}
		goCh <- result{code: code}
	}()

	cRes := <-cCh
	goRes := <-goCh
	if cRes.err != nil {
		t.Fatalf("C initiator error: %v", cRes.err)
	}
	if goRes.err != nil {
		t.Fatalf("Go acceptor error: %v", goRes.err)
	}
	if goRes.code != cRes.code {
		t.Fatalf("code mismatch: Go acceptor=%q C initiator=%q", goRes.code, cRes.code)
	}
	if len(goRes.code) != 6 {
		t.Fatalf("unexpected code length %d", len(goRes.code))
	}
	t.Logf("confirmation code: %s", goRes.code)
}

// TestWireParityCAcceptorGoInitiator: C acceptor ↔ Go initiator.
//
// The Go initiator opens the stream itself and runs hello→welcome→confirm.
// The C acceptor (RunAcceptor) blocks on transport->accept_stream internally
// and then runs the ceremony. Both sides must derive the same confirmation code.
func TestWireParityCAcceptorGoInitiator(t *testing.T) {
	t.Parallel()

	accEphPriv, accEphPub := newEphKey(t)
	_, initEphPub := newEphKey(t)

	accIdPub := make([]byte, 32)
	initIdPub := make([]byte, 32)

	// ca = acceptor side (C), cb = initiator side (Go).
	ca, cb := newChanTransportPair()
	refA := cwire.NewGoTransportRef(ca)
	defer refA.Close()

	type result struct {
		code string
		err  error
	}
	cCh := make(chan result, 1)
	goCh := make(chan result, 1)

	// C acceptor goroutine: blocks on accept_stream internally.
	go func() {
		_, code, err := cwire.RunAcceptor(&cwire.RunAcceptorArgs{
			Ref:          refA,
			LocalEphPriv: accEphPriv.Bytes(),
			LocalEphPub:  accEphPub.Bytes(),
			IdentityPub:  accIdPub,
			InstanceID:   "c-acceptor",
			ConfirmFn:    func(code string) bool { return true },
		})
		cCh <- result{code: code, err: err}
	}()

	// Go initiator goroutine: commit → welcome → reveal → confirm.
	go func() {
		h, err := cb.OpenStream()
		if err != nil {
			goCh <- result{err: err}
			return
		}
		blind := make([]byte, 32)
		if _, err := rand.Read(blind); err != nil {
			goCh <- result{err: err}
			return
		}
		commit, err := crypto.SASCommit(initEphPub.Bytes(), blind)
		if err != nil {
			goCh <- result{err: err}
			return
		}
		hb, _ := json.Marshal(pairingMsg{Kind: "hello", Commit: commit, IdentityPub: initIdPub, InstanceID: "go-initiator"})
		if err := cb.SendOnStream(h, hb); err != nil {
			goCh <- result{err: err}
			return
		}
		raw, err := cb.RecvOnStream(h)
		if err != nil {
			goCh <- result{err: err}
			return
		}
		var welcome pairingMsg
		if err := json.Unmarshal(raw, &welcome); err != nil || welcome.Kind != "welcome" {
			goCh <- result{err: fmt.Errorf("welcome: %w", err)}
			return
		}
		if !bytes.Equal(welcome.EphPub, accEphPub.Bytes()) {
			goCh <- result{err: fmt.Errorf("welcome eph mismatch vs expected acceptor key")}
			return
		}
		rb, _ := json.Marshal(pairingMsg{Kind: "reveal", EphPub: initEphPub.Bytes(), Blind: blind})
		if err := cb.SendOnStream(h, rb); err != nil {
			goCh <- result{err: err}
			return
		}
		code, err := crypto.DeriveConfirmationCode(accEphPub, initEphPub)
		if err != nil {
			goCh <- result{err: err}
			return
		}
		cb2, _ := json.Marshal(pairingMsg{Kind: "confirm"})
		if err := cb.SendOnStream(h, cb2); err != nil {
			goCh <- result{err: err}
			return
		}
		raw, err = cb.RecvOnStream(h)
		if err != nil {
			goCh <- result{err: err}
			return
		}
		var conf pairingMsg
		if err := json.Unmarshal(raw, &conf); err != nil || conf.Kind != "confirm" {
			goCh <- result{err: fmt.Errorf("confirm: %w", err)}
			return
		}
		goCh <- result{code: code}
	}()

	cRes := <-cCh
	goRes := <-goCh
	if cRes.err != nil {
		t.Fatalf("C acceptor error: %v", cRes.err)
	}
	if goRes.err != nil {
		t.Fatalf("Go initiator error: %v", goRes.err)
	}
	if goRes.code != cRes.code {
		t.Fatalf("code mismatch: Go initiator=%q C acceptor=%q", goRes.code, cRes.code)
	}
	if len(goRes.code) != 6 {
		t.Fatalf("unexpected code length %d", len(goRes.code))
	}
	t.Logf("confirmation code: %s", goRes.code)
}

// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package cwire_test

import (
	"crypto/ecdh"
	"crypto/rand"
	"testing"

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

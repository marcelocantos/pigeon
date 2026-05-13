// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package pigeon

import (
	"errors"
	"io"
	"sync"
	"testing"

	"github.com/marcelocantos/pigeon/crypto"
)

// newPipePair returns a two-way in-memory io.ReadWriter pair, suitable
// for driving runBackendActivation against runClientActivation in the
// same process without spinning up a relay or QUIC.
func newPipePair() (clientEnd, backendEnd io.ReadWriter) {
	c2bR, c2bW := io.Pipe()
	b2cR, b2cW := io.Pipe()
	return &pipeEnd{r: b2cR, w: c2bW}, &pipeEnd{r: c2bR, w: b2cW}
}

type pipeEnd struct {
	r io.Reader
	w io.Writer
}

func (p *pipeEnd) Read(b []byte) (int, error)  { return p.r.Read(b) }
func (p *pipeEnd) Write(b []byte) (int, error) { return p.w.Write(b) }

// makeTestPairing returns a fresh PairingRecord usable for activation
// tests. The activation flow only checks the record's presence/absence;
// the keys themselves are exercised downstream by DeriveChannel.
func makeTestPairing(t *testing.T) *crypto.PairingRecord {
	t.Helper()
	clientKP, err := crypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}
	backendKP, err := crypto.GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair: %v", err)
	}
	return crypto.NewPairingRecord("backend-test", "https://relay.example.com", clientKP, backendKP.Public)
}

// TestActivation_KnownDevice walks the happy path: client sends
// auth_request, backend resolves the device, writes auth_ok, both
// machines reach SessionActive.
func TestActivation_KnownDevice(t *testing.T) {
	rec := makeTestPairing(t)
	const deviceID = "device-known-1"
	resolve := func(id string) (*crypto.PairingRecord, error) {
		if id != deviceID {
			return nil, errors.New("unknown")
		}
		return rec, nil
	}

	clientEnd, backendEnd := newPipePair()

	var wg sync.WaitGroup
	var (
		gotMachine  *SessionProtocolBackendMachine
		gotDeviceID string
		gotRecord   *crypto.PairingRecord
		backendErr  error
	)
	wg.Add(1)
	go func() {
		defer wg.Done()
		gotMachine, gotDeviceID, gotRecord, backendErr = runBackendActivation(backendEnd, resolve)
	}()

	clientMachine, clientErr := runClientActivation(clientEnd, deviceID)
	wg.Wait()

	if clientErr != nil {
		t.Fatalf("client activation: %v", clientErr)
	}
	if backendErr != nil {
		t.Fatalf("backend activation: %v", backendErr)
	}
	if gotDeviceID != deviceID {
		t.Errorf("backend received deviceID %q, want %q", gotDeviceID, deviceID)
	}
	if gotRecord != rec {
		t.Errorf("backend record mismatch")
	}
	if got, want := gotMachine.State, SessionProtocolBackendSessionActive; got != want {
		t.Errorf("backend state = %q, want %q", got, want)
	}
	if got, want := clientMachine.State, SessionProtocolClientSessionActive; got != want {
		t.Errorf("client state = %q, want %q", got, want)
	}
	if got := gotMachine.ReceivedDeviceId; got != deviceID {
		t.Errorf("backend ReceivedDeviceId = %q, want %q", got, deviceID)
	}
}

// TestActivation_UnknownDevice walks the rejection path: backend's
// resolve returns nil, machine takes the AuthCheck → Idle branch via
// the device_unknown guard, client sees the rejection.
func TestActivation_UnknownDevice(t *testing.T) {
	resolve := func(string) (*crypto.PairingRecord, error) {
		return nil, errors.New("unknown")
	}
	clientEnd, backendEnd := newPipePair()

	var (
		gotMachine *SessionProtocolBackendMachine
		backendErr error
	)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		gotMachine, _, _, backendErr = runBackendActivation(backendEnd, resolve)
	}()

	_, clientErr := runClientActivation(clientEnd, "device-stranger")
	wg.Wait()

	if clientErr == nil {
		t.Fatal("client activation: want error, got nil")
	}
	if backendErr == nil {
		t.Fatal("backend activation: want error, got nil")
	}
	if got, want := gotMachine.State, SessionProtocolBackendIdle; got != want {
		t.Errorf("backend state = %q, want %q (AuthCheck → Idle on device_unknown)", got, want)
	}
}

// TestActivation_WireRoundtrip locks down the binary wire format so
// future generator changes (T40) can't accidentally diverge.
func TestActivation_WireRoundtrip(t *testing.T) {
	t.Run("auth_request", func(t *testing.T) {
		const id = "device-wire-1"
		got, err := decodeAuthRequest(encodeAuthRequest(id))
		if err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got != id {
			t.Errorf("got %q, want %q", got, id)
		}
	})
	t.Run("auth_ok_accepted", func(t *testing.T) {
		ok, reason, err := decodeAuthOk(encodeAuthOk(true, ""))
		if err != nil {
			t.Fatalf("decode: %v", err)
		}
		if !ok || reason != "" {
			t.Errorf("got (%v, %q), want (true, \"\")", ok, reason)
		}
	})
	t.Run("auth_ok_rejected", func(t *testing.T) {
		ok, reason, err := decodeAuthOk(encodeAuthOk(false, "unknown client"))
		if err != nil {
			t.Fatalf("decode: %v", err)
		}
		if ok || reason != "unknown client" {
			t.Errorf("got (%v, %q), want (false, \"unknown client\")", ok, reason)
		}
	})
}

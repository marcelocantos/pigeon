// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package pigeon

import (
	"encoding/binary"
	"fmt"
	"io"

	"github.com/marcelocantos/pigeon/crypto"
)

// Activation drives the session.yaml pairing-phase auth states
// (Paired → AuthCheck → SessionActive on the backend; Reconnect →
// SendAuth → SessionActive on the client) on a per-client
// SessionMachine. The machine instance is exclusive to this client
// connection — no shared state between clients on the backend side.
//
// The exchange is two messages on the relay's primary stream, immediately
// after the stream-name binding header:
//
//	client → backend: auth_request{device_id}
//	backend → client: auth_ok{ok, reason?}
//
// Both messages are sent via writeMessage / readMessage (4-byte
// big-endian length prefix), wrapping the binary payload defined below.
// The wire format will migrate into the protogen one-shot byte-format
// generator under T40; for T39.1 the encoders live here so that the
// activation flow itself can be wired through the generated executor
// without depending on T40's not-yet-built generator framework.

const (
	// authMsgTagAuthRequest tags an auth_request payload.
	authMsgTagAuthRequest byte = 0x01
	// authMsgTagAuthOk tags an auth_ok payload.
	authMsgTagAuthOk byte = 0x02

	// authOkAccepted is the ok flag value for an accepted activation.
	authOkAccepted byte = 0x01
	// authOkRejected is the ok flag value for a rejected activation
	// (carries a reason string).
	authOkRejected byte = 0x00
)

// encodeAuthRequest builds the binary payload for the client's
// auth_request message: [tag][uvarint device-id length][device-id bytes].
func encodeAuthRequest(deviceID string) []byte {
	buf := make([]byte, 0, 1+binary.MaxVarintLen64+len(deviceID))
	buf = append(buf, authMsgTagAuthRequest)
	var lenBuf [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(lenBuf[:], uint64(len(deviceID)))
	buf = append(buf, lenBuf[:n]...)
	buf = append(buf, deviceID...)
	return buf
}

// decodeAuthRequest parses an auth_request payload and returns the
// device ID. Returns an error if the tag is wrong or the payload is
// malformed / truncated.
func decodeAuthRequest(payload []byte) (string, error) {
	if len(payload) == 0 {
		return "", fmt.Errorf("auth_request: empty payload")
	}
	if payload[0] != authMsgTagAuthRequest {
		return "", fmt.Errorf("auth_request: unexpected tag 0x%02x", payload[0])
	}
	rest := payload[1:]
	idLen, n := binary.Uvarint(rest)
	if n <= 0 {
		return "", fmt.Errorf("auth_request: bad device-id length")
	}
	rest = rest[n:]
	if uint64(len(rest)) != idLen {
		return "", fmt.Errorf("auth_request: device-id length mismatch (got %d want %d)", len(rest), idLen)
	}
	return string(rest), nil
}

// encodeAuthOk builds the binary payload for the backend's auth_ok
// reply. ok=true: [tag][0x01]; ok=false: [tag][0x00][uvarint reason
// length][reason bytes].
func encodeAuthOk(ok bool, reason string) []byte {
	if ok {
		return []byte{authMsgTagAuthOk, authOkAccepted}
	}
	buf := make([]byte, 0, 2+binary.MaxVarintLen64+len(reason))
	buf = append(buf, authMsgTagAuthOk, authOkRejected)
	var lenBuf [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(lenBuf[:], uint64(len(reason)))
	buf = append(buf, lenBuf[:n]...)
	buf = append(buf, reason...)
	return buf
}

// decodeAuthOk parses an auth_ok payload and returns (accepted, reason).
// reason is empty when accepted=true.
func decodeAuthOk(payload []byte) (bool, string, error) {
	if len(payload) < 2 {
		return false, "", fmt.Errorf("auth_ok: short payload")
	}
	if payload[0] != authMsgTagAuthOk {
		return false, "", fmt.Errorf("auth_ok: unexpected tag 0x%02x", payload[0])
	}
	switch payload[1] {
	case authOkAccepted:
		return true, "", nil
	case authOkRejected:
		rest := payload[2:]
		reasonLen, n := binary.Uvarint(rest)
		if n <= 0 {
			return false, "", fmt.Errorf("auth_ok: bad reason length")
		}
		rest = rest[n:]
		if uint64(len(rest)) != reasonLen {
			return false, "", fmt.Errorf("auth_ok: reason length mismatch (got %d want %d)", len(rest), reasonLen)
		}
		return false, string(rest), nil
	default:
		return false, "", fmt.Errorf("auth_ok: unknown ok flag 0x%02x", payload[1])
	}
}

// runBackendActivation drives the backend's per-client SessionMachine
// through Paired → AuthCheck → SessionActive on a valid auth_request,
// or Paired → AuthCheck → Idle on an unknown device. It reads the
// auth_request from stream, looks up the PairingRecord, writes the
// auth_ok reply, and returns the looked-up record on success.
//
// resolve maps a client-supplied device ID to its PairingRecord; it is
// the same callback registered by the application as RegisterArgs.Pairing.
// The lookup itself is a backend-runtime concern (the backend serves N
// already-paired devices), kept outside the spec — the spec sees only
// the device_known / device_unknown guard outcome.
func runBackendActivation(
	stream io.ReadWriter,
	resolve func(deviceID string) (*crypto.PairingRecord, error),
) (*SessionProtocolBackendMachine, string, *crypto.PairingRecord, error) {
	raw, err := readMessage(stream)
	if err != nil {
		return nil, "", nil, fmt.Errorf("read auth_request: %w", err)
	}
	deviceID, err := decodeAuthRequest(raw)
	if err != nil {
		return nil, "", nil, err
	}

	machine := newBackendActivationMachine(deviceID)

	rec, lookupErr := resolve(deviceID)
	known := lookupErr == nil && rec != nil

	machine.Guards[SessionProtocolGuardDeviceKnown] = func() bool { return known }
	machine.Guards[SessionProtocolGuardDeviceUnknown] = func() bool { return !known }

	if _, err := machine.HandleEvent(SessionProtocolEventRecvAuthRequest); err != nil {
		return nil, "", nil, fmt.Errorf("auth_request transition: %w", err)
	}
	if machine.State != SessionProtocolBackendAuthCheck {
		return nil, "", nil, fmt.Errorf("auth_request: machine in %q after recv (want AuthCheck)", machine.State)
	}
	if _, err := machine.HandleEvent(SessionProtocolEventVerify); err != nil {
		return nil, "", nil, fmt.Errorf("verify transition: %w", err)
	}

	if !known {
		_ = writeMessage(stream, encodeAuthOk(false, "unknown client"))
		return machine, deviceID, nil, fmt.Errorf("unknown client %q", deviceID)
	}

	if err := writeMessage(stream, encodeAuthOk(true, "")); err != nil {
		return nil, "", nil, fmt.Errorf("write auth_ok: %w", err)
	}
	if machine.State != SessionProtocolBackendSessionActive {
		return nil, "", nil, fmt.Errorf("verify: machine in %q after verify (want SessionActive)", machine.State)
	}
	return machine, deviceID, rec, nil
}

// runClientActivation drives the client's per-client SessionMachine
// through Reconnect → SendAuth → SessionActive. It writes auth_request
// to stream, reads the auth_ok reply, and returns the machine on
// success. A rejected auth_ok is reported as an error; the caller is
// expected to close the underlying transport.
func runClientActivation(
	stream io.ReadWriter,
	deviceID string,
) (*SessionProtocolClientMachine, error) {
	machine := newClientActivationMachine()

	if _, err := machine.HandleEvent(SessionProtocolEventRelayConnected); err != nil {
		return nil, fmt.Errorf("relay_connected transition: %w", err)
	}
	if machine.State != SessionProtocolClientSendAuth {
		return nil, fmt.Errorf("relay_connected: machine in %q (want SendAuth)", machine.State)
	}

	if err := writeMessage(stream, encodeAuthRequest(deviceID)); err != nil {
		return nil, fmt.Errorf("write auth_request: %w", err)
	}

	raw, err := readMessage(stream)
	if err != nil {
		return nil, fmt.Errorf("read auth_ok: %w", err)
	}
	ok, reason, err := decodeAuthOk(raw)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("backend rejected: %s", reason)
	}

	if _, err := machine.HandleEvent(SessionProtocolEventRecvAuthOk); err != nil {
		return nil, fmt.Errorf("recv_auth_ok transition: %w", err)
	}
	if machine.State != SessionProtocolClientSessionActive {
		return nil, fmt.Errorf("recv_auth_ok: machine in %q (want SessionActive)", machine.State)
	}
	return machine, nil
}

// newBackendActivationMachine builds a per-client backend machine
// pre-seeded at Paired with the supplied device_id already received
// from the wire — i.e. the state the spec assumes the backend is in
// when an auth_request arrives. The lookup from device_id to
// PairingRecord happens outside the spec; see runBackendActivation.
func newBackendActivationMachine(deviceID string) *SessionProtocolBackendMachine {
	m := NewSessionProtocolBackendMachine()
	m.State = SessionProtocolBackendPaired
	m.ReceivedDeviceId = deviceID
	// device_known / device_unknown guards are bound by the caller
	// after the resolve() lookup completes.
	return m
}

// newClientActivationMachine builds a per-client client machine
// pre-seeded at Reconnect — the state the spec uses for an
// already-paired client re-establishing a session. The transition
// from Reconnect to SendAuth fires on relay_connected.
func newClientActivationMachine() *SessionProtocolClientMachine {
	m := NewSessionProtocolClientMachine()
	m.State = SessionProtocolClientReconnect
	return m
}

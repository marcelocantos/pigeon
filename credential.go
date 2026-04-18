// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package pigeon

import (
	"github.com/marcelocantos/pigeon/crypto"
)

// Credential is an opaque blob that a client stores and presents for
// authentication. It wraps a crypto.PairingRecord — apps should treat
// it as opaque bytes (Marshal/Unmarshal).
type Credential = crypto.PairingRecord

// IssueCredential generates a credential for a device without running
// the pairing ceremony. The server creates a fresh keypair and derives
// the shared secret; the returned Credential contains the device's
// side of the keys. The server-side PairingRecord (returned as the
// second value) should be registered with the auth sub-machine so that
// subsequent auth_request messages from the device are accepted.
//
// This enables server-side credential injection: the server can
// pre-provision devices (e.g., via an admin API) and hand them an
// opaque credential blob to store and present on reconnect.
func IssueCredential(peerInstanceID, relayURL string) (deviceCredential *Credential, serverRecord *crypto.PairingRecord, err error) {
	serverKP, err := crypto.GenerateKeyPair()
	if err != nil {
		return nil, nil, err
	}
	deviceKP, err := crypto.GenerateKeyPair()
	if err != nil {
		return nil, nil, err
	}

	// The device credential contains the device's private key and
	// the server's public key — everything the device needs to derive
	// session keys on reconnect.
	deviceCred := crypto.NewPairingRecord(peerInstanceID, relayURL, deviceKP, serverKP.Public)

	// The server record contains the server's private key and the
	// device's public key — everything the server needs to verify
	// auth_request messages from this device.
	serverRec := crypto.NewPairingRecord(peerInstanceID, relayURL, serverKP, deviceKP.Public)

	return deviceCred, serverRec, nil
}

// UnmarshalCredential deserializes a credential from JSON.
// This is the inverse of Credential.Marshal().
func UnmarshalCredential(data []byte) (*Credential, error) {
	return crypto.UnmarshalPairingRecord(data)
}

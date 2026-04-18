// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package pigeon

import (
	"testing"
)

func TestIssueCredential(t *testing.T) {
	deviceCred, serverRec, err := IssueCredential("inst-123", "https://relay.example.com")
	if err != nil {
		t.Fatal("IssueCredential:", err)
	}

	if deviceCred.PeerInstanceID != "inst-123" {
		t.Fatalf("device credential instance ID = %q, want inst-123", deviceCred.PeerInstanceID)
	}
	if serverRec.PeerInstanceID != "inst-123" {
		t.Fatalf("server record instance ID = %q, want inst-123", serverRec.PeerInstanceID)
	}
	if deviceCred.RelayURL != "https://relay.example.com" {
		t.Fatalf("relay URL = %q, want https://relay.example.com", deviceCred.RelayURL)
	}

	// Device credential's peer public key should be the server's local public key.
	if string(deviceCred.PeerPublicKey) != string(serverRec.LocalPublicKey) {
		t.Fatal("device peer key does not match server local key")
	}
	// Server record's peer public key should be the device's local public key.
	if string(serverRec.PeerPublicKey) != string(deviceCred.LocalPublicKey) {
		t.Fatal("server peer key does not match device local key")
	}

	// Both sides should derive the same channel.
	deviceCh, err := deviceCred.DeriveChannel([]byte("client-to-server"), []byte("server-to-client"))
	if err != nil {
		t.Fatal("device derive:", err)
	}
	serverCh, err := serverRec.DeriveChannel([]byte("server-to-client"), []byte("client-to-server"))
	if err != nil {
		t.Fatal("server derive:", err)
	}

	// Round-trip: device encrypts, server decrypts.
	ct := deviceCh.Encrypt([]byte("hello from device"))
	pt, err := serverCh.Decrypt(ct)
	if err != nil {
		t.Fatal("server decrypt:", err)
	}
	if string(pt) != "hello from device" {
		t.Fatalf("decrypted = %q, want hello from device", pt)
	}

	// Round-trip: server encrypts, device decrypts.
	ct2 := serverCh.Encrypt([]byte("hello from server"))
	pt2, err := deviceCh.Decrypt(ct2)
	if err != nil {
		t.Fatal("device decrypt:", err)
	}
	if string(pt2) != "hello from server" {
		t.Fatalf("decrypted = %q, want hello from server", pt2)
	}

	// Credential should marshal/unmarshal.
	data, err := deviceCred.Marshal()
	if err != nil {
		t.Fatal("marshal:", err)
	}
	restored, err := UnmarshalCredential(data)
	if err != nil {
		t.Fatal("unmarshal:", err)
	}
	if restored.PeerInstanceID != deviceCred.PeerInstanceID {
		t.Fatal("restored instance ID mismatch")
	}
}

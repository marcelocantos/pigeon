// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package pigeon

import (
	"bytes"
	"testing"
)

// TestStreamHeaderWireVectors locks the on-the-wire byte representation
// of the post-T22 stream-header encoding so that every cross-language
// SDK (c/test/test_pigeon.c::test_stream_header, the Swift / Kotlin /
// TS suites) can hard-code the same expected bytes and verify cross-
// language equivalence by construction. Post-T40 the encoder /decoder
// are protogen-generated in wire_gen.go; this test exercises the
// generated functions directly.
//
// If this test changes, the corresponding per-language test vectors
// MUST be updated in lockstep — they are intentionally hand-mirrored
// as the cheapest cross-language wire-byte interop check we can run today.
func TestStreamHeaderWireVectors(t *testing.T) {
	cases := []struct {
		name      string
		isBackend bool
		tag       uint32
		channel   string
		want      []byte
	}{
		{
			name:      "backend chat tag=0x01020304",
			isBackend: true, tag: 0x01020304, channel: "chat",
			// 4-byte tag + varint(4) + "chat"
			want: []byte{0x01, 0x02, 0x03, 0x04, 0x04, 'c', 'h', 'a', 't'},
		},
		{
			name:      "client primary (empty name)",
			isBackend: false, tag: 0, channel: "",
			want: []byte{0x00},
		},
		{
			name:      "client control",
			isBackend: false, tag: 0, channel: "control",
			want: []byte{0x07, 'c', 'o', 'n', 't', 'r', 'o', 'l'},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := EncodeStreamHeader(tc.isBackend, tc.tag, tc.channel)
			if !bytes.Equal(got, tc.want) {
				t.Fatalf("EncodeStreamHeader: got %x want %x", got, tc.want)
			}

			// Round-trip via the matching decoder.
			if tc.isBackend {
				tag, name, err := DecodeStreamHeaderBackend(got)
				if err != nil {
					t.Fatalf("decode: %v", err)
				}
				if tag != tc.tag {
					t.Fatalf("decode tag: got 0x%x want 0x%x", tag, tc.tag)
				}
				if name != tc.channel {
					t.Fatalf("decode name: got %q want %q", name, tc.channel)
				}
			} else {
				name, err := DecodeStreamHeaderClient(got)
				if err != nil {
					t.Fatalf("decode: %v", err)
				}
				if name != tc.channel {
					t.Fatalf("decode name: got %q want %q", name, tc.channel)
				}
			}
		})
	}
}

// TestRelayGreetingWireVectors locks the relay greeting variants. The
// register-mux form has four wire shapes depending on which of (token,
// instance_id) are present; the encoder picks the right one.
func TestRelayGreetingWireVectors(t *testing.T) {
	t.Run("connect", func(t *testing.T) {
		got := EncodeRelayGreetingConnect("xyz123")
		want := []byte("connect:xyz123")
		if !bytes.Equal(got, want) {
			t.Fatalf("got %q want %q", got, want)
		}
		dec, err := DecodeRelayGreeting(got)
		if err != nil {
			t.Fatalf("decode: %v", err)
		}
		if dec.Variant != RelayGreetingConnect {
			t.Fatalf("variant: %v", dec.Variant)
		}
		if dec.InstanceId != "xyz123" {
			t.Fatalf("instance: %q", dec.InstanceId)
		}
	})
	cases := []struct {
		name     string
		token    string
		instance string
		want     string
	}{
		{"bare", "", "", "register-mux"},
		{"id only", "", "id-7", "register-mux::id-7"},
		{"token only", "secret", "", "register-mux:secret:"},
		{"both", "tok", "id-2", "register-mux:tok:id-2"},
	}
	for _, tc := range cases {
		t.Run("register-mux/"+tc.name, func(t *testing.T) {
			got := EncodeRelayGreetingRegisterMux(tc.token, tc.instance)
			if string(got) != tc.want {
				t.Fatalf("got %q want %q", got, tc.want)
			}
			dec, err := DecodeRelayGreeting(got)
			if err != nil {
				t.Fatalf("decode: %v", err)
			}
			if dec.Variant != RelayGreetingRegisterMux {
				t.Fatalf("variant: %v", dec.Variant)
			}
			if dec.Token != tc.token {
				t.Fatalf("token: got %q want %q", dec.Token, tc.token)
			}
			if dec.InstanceId != tc.instance {
				t.Fatalf("instance: got %q want %q", dec.InstanceId, tc.instance)
			}
		})
	}
}

// TestDatagramPlaintextRoundTrip exercises EncodeDatagramPlaintext /
// DecodeDatagramPlaintext directly. AEAD wrapping is layered above this
// in the live wire, so the plaintext format is what's nailed down here.
func TestDatagramPlaintextRoundTrip(t *testing.T) {
	cases := []struct {
		name      string
		channelID uint64
		payload   []byte
		wantHead  []byte // expected prefix bytes (varint channel-id) for sanity
	}{
		{"single-byte cid", 1, []byte("hello"), []byte{0x01}},
		{"multi-byte cid", 300, []byte("xyz"), []byte{0xac, 0x02}},
		{"empty payload", 7, nil, []byte{0x07}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := EncodeDatagramPlaintext(tc.channelID, tc.payload)
			if !bytes.HasPrefix(got, tc.wantHead) {
				t.Fatalf("missing varint prefix: got %x want prefix %x", got, tc.wantHead)
			}
			id, body, err := DecodeDatagramPlaintext(got)
			if err != nil {
				t.Fatalf("decode: %v", err)
			}
			if id != tc.channelID {
				t.Fatalf("id: got %d want %d", id, tc.channelID)
			}
			if !bytes.Equal(body, tc.payload) {
				t.Fatalf("payload: got %x want %x", body, tc.payload)
			}
		})
	}
}

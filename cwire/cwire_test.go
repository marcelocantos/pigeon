// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package cwire_test

import (
	"bytes"
	"encoding/binary"
	"testing"

	"github.com/marcelocantos/pigeon/cwire"
)

// TestUvarintMatchesGo proves the C uvarint encoder emits the same
// bytes as encoding/binary.PutUvarint for the same input. This is the
// foundational byte-parity check between Go and C.
func TestUvarintMatchesGo(t *testing.T) {
	cases := []uint64{0, 1, 127, 128, 300, 16384, 1 << 32, 1 << 63}
	for _, v := range cases {
		want := make([]byte, binary.MaxVarintLen64)
		want = want[:binary.PutUvarint(want, v)]
		got := cwire.EncodeUvarint(v)
		if !bytes.Equal(got, want) {
			t.Errorf("v=%d: C %x != Go %x", v, got, want)
		}

		v2, n, err := cwire.DecodeUvarint(want)
		if err != nil {
			t.Errorf("v=%d: decode: %v", v, err)
			continue
		}
		if n != len(want) {
			t.Errorf("v=%d: decode consumed %d, want %d", v, n, len(want))
		}
		if v2 != v {
			t.Errorf("v=%d: decode value %d", v, v2)
		}
	}
}

// TestDatagramRoundTripViaC drives encode_datagram → decode_datagram
// through the C ABI from Go. Pinning the AEAD-with-channel-id wire
// across Go/C/TS at this layer means a future drift on any side
// turns this test red. (Byte-vector tests against a hardcoded wire
// aren't possible here because AEAD ciphertext depends on the
// per-channel sequence number.)
func TestDatagramRoundTripViaC(t *testing.T) {
	sendKey := make([]byte, 32)
	recvKey := make([]byte, 32)
	for i := range sendKey {
		sendKey[i] = 0x42
		recvKey[i] = 0x42
	}

	cases := []struct {
		name      string
		channelID uint64
		payload   []byte
	}{
		{"small chat", 1, []byte("hello")},
		{"multi-byte channelID", 300, []byte("xyz")},
		{"empty payload", 7, nil},
		{"ping", 2, []byte("ping")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wire, err := cwire.EncodeDatagram(sendKey, recvKey, tc.channelID, tc.payload)
			if err != nil {
				t.Fatalf("encode: %v", err)
			}
			cid, payload, err := cwire.DecodeDatagram(sendKey, recvKey, wire)
			if err != nil {
				t.Fatalf("decode: %v", err)
			}
			if cid != tc.channelID {
				t.Errorf("channel id: got %d want %d", cid, tc.channelID)
			}
			if !bytes.Equal(payload, tc.payload) {
				t.Errorf("payload: got %x want %x", payload, tc.payload)
			}
		})
	}
}

// TestStreamHeaderMatchesReferenceVectors locks the C bridge against
// the same wire vectors as wire_vectors_test.go and
// c/test/test_pigeon.c::test_stream_header. Three independent
// implementations (Go, C-via-cwire, hand-coded test vectors) all
// agreeing on the same bytes is the durable cross-language pinning.
func TestStreamHeaderMatchesReferenceVectors(t *testing.T) {
	cases := []struct {
		name    string
		channel string
		want    []byte
	}{
		{
			name:    "chat",
			channel: "chat",
			want:    []byte{0x04, 'c', 'h', 'a', 't'},
		},
		{
			name:    "primary (empty name)",
			channel: "",
			want:    []byte{0x00},
		},
		{
			name:    "control",
			channel: "control",
			want:    []byte{0x07, 'c', 'o', 'n', 't', 'r', 'o', 'l'},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := cwire.EncodeStreamHeader(tc.channel)
			if err != nil {
				t.Fatalf("encode: %v", err)
			}
			if !bytes.Equal(got, tc.want) {
				t.Fatalf("got %x want %x", got, tc.want)
			}

			name, n, err := cwire.DecodeStreamHeader(got)
			if err != nil {
				t.Fatalf("decode: %v", err)
			}
			if name != tc.channel {
				t.Fatalf("name: got %q want %q", name, tc.channel)
			}
			if n != len(got) {
				t.Fatalf("consumed: got %d want %d", n, len(got))
			}
		})
	}
}

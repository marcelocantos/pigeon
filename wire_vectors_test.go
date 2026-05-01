// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package pigeon

import (
	"bytes"
	"testing"
)

// TestStreamHeaderWireVectors locks the on-the-wire byte representation
// of the post-T22 stream-header encoding so that the C SDK
// (c/test/test_pigeon.c::test_stream_header) can hard-code the same
// expected bytes and verify cross-language equivalence by construction.
//
// If this test changes, the corresponding C test vectors MUST be
// updated in lockstep — they are intentionally hand-mirrored as the
// cheapest cross-language wire-byte interop check we can run today.
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
			got := encodeStreamHeader(tc.isBackend, tc.tag, tc.channel)
			if !bytes.Equal(got, tc.want) {
				t.Fatalf("encodeStreamHeader: got %x want %x", got, tc.want)
			}

			// Round-trip via the matching decoder.
			if tc.isBackend {
				tag, name, err := decodeBackendStreamHeader(got)
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
				name, err := decodeClientStreamHeader(got)
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

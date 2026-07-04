// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package pigeon

import (
	"bytes"
	"testing"
)

func TestEncodeDecodeRoutesRoundTrip(t *testing.T) {
	cases := [][]RouteEntry{
		nil,
		{{Route: "game-a", Metadata: []byte("Space Battle")}},
		{
			{Route: "a", Metadata: nil},
			{Route: "b", Metadata: []byte{0x00, 0xff, 0x00}},
			{Route: "", Metadata: []byte("x")},
		},
	}
	for i, in := range cases {
		out, err := decodeRoutes(encodeRoutes(in))
		if err != nil {
			t.Fatalf("case %d: decode: %v", i, err)
		}
		if len(out) != len(in) {
			t.Fatalf("case %d: len=%d want %d", i, len(out), len(in))
		}
		for j := range in {
			if out[j].Route != in[j].Route || !bytes.Equal(out[j].Metadata, in[j].Metadata) {
				t.Fatalf("case %d entry %d: got %+v want %+v", i, j, out[j], in[j])
			}
		}
	}
}

func TestDecodeRoutesRejectsTruncated(t *testing.T) {
	full := encodeRoutes([]RouteEntry{{Route: "abc", Metadata: []byte("meta")}})
	// Every strict prefix is malformed and must error, not panic.
	for n := 0; n < len(full); n++ {
		if _, err := decodeRoutes(full[:n]); err == nil {
			t.Fatalf("truncated to %d bytes decoded without error", n)
		}
	}
}

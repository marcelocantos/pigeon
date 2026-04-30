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

// TestStreamHeaderMatchesReferenceVectors locks the C bridge against
// the same wire vectors as wire_vectors_test.go and
// c/test/test_pigeon.c::test_stream_header. Three independent
// implementations (Go, C-via-cwire, hand-coded test vectors) all
// agreeing on the same bytes is the durable cross-language pinning.
func TestStreamHeaderMatchesReferenceVectors(t *testing.T) {
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
			got, err := cwire.EncodeStreamHeader(tc.isBackend, tc.tag, tc.channel)
			if err != nil {
				t.Fatalf("encode: %v", err)
			}
			if !bytes.Equal(got, tc.want) {
				t.Fatalf("got %x want %x", got, tc.want)
			}

			if tc.isBackend {
				tag, name, n, err := cwire.DecodeBackendStreamHeader(got)
				if err != nil {
					t.Fatalf("decode: %v", err)
				}
				if tag != tc.tag {
					t.Fatalf("tag: got %#x want %#x", tag, tc.tag)
				}
				if name != tc.channel {
					t.Fatalf("name: got %q want %q", name, tc.channel)
				}
				if n != len(got) {
					t.Fatalf("consumed: got %d want %d", n, len(got))
				}
			} else {
				name, n, err := cwire.DecodeClientStreamHeader(got)
				if err != nil {
					t.Fatalf("decode: %v", err)
				}
				if name != tc.channel {
					t.Fatalf("name: got %q want %q", name, tc.channel)
				}
				if n != len(got) {
					t.Fatalf("consumed: got %d want %d", n, len(got))
				}
			}
		})
	}
}

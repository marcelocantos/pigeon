// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package pigeon

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestEncodeDecodeWholePart(t *testing.T) {
	ct := []byte("cipher-blob")
	wire := encodePartWire(0, 0, 1, ct)
	if wire[0] != DgConnWhole {
		t.Fatalf("prefix %x", wire[0])
	}
	msgID, idx, total, gotCT, ok := decodePartWire(wire)
	if !ok || msgID != 0 || idx != 0 || total != 1 {
		t.Fatalf("meta %d %d %d ok=%v", msgID, idx, total, ok)
	}
	if !bytes.Equal(gotCT, ct) {
		t.Fatalf("ct %q", gotCT)
	}
}

func TestEncodeDecodeFragmentParts(t *testing.T) {
	chunks := [][]byte{[]byte("aaa"), []byte("bbbb"), []byte("cc")}
	var wires [][]byte
	const msgID = 42
	total := uint16(len(chunks))
	for i, ch := range chunks {
		wires = append(wires, encodePartWire(msgID, uint16(i), total, ch))
	}
	for i, w := range wires {
		if w[0] != DgConnFragment {
			t.Fatalf("frame %d prefix", i)
		}
		id, idx, tot, ct, ok := decodePartWire(w)
		if !ok || id != msgID || idx != uint16(i) || tot != total {
			t.Fatalf("frame %d meta id=%d idx=%d tot=%d", i, id, idx, tot)
		}
		if !bytes.Equal(ct, chunks[i]) {
			t.Fatalf("frame %d ct", i)
		}
	}
}

func TestDecodeRejectsBadFragmentMeta(t *testing.T) {
	// total=1 is illegal on fragment prefix
	w := make([]byte, 1+FragHeaderSize+1)
	w[0] = DgConnFragment
	binary.BigEndian.PutUint32(w[1:5], 1)
	binary.BigEndian.PutUint16(w[5:7], 0)
	binary.BigEndian.PutUint16(w[7:9], 1) // total must be >= 2
	if _, _, _, _, ok := decodePartWire(w); ok {
		t.Fatal("expected reject")
	}
}

func TestDecodeLegacyUnframed(t *testing.T) {
	// First byte neither 0x00 nor 0x40 — legacy whole AEAD.
	raw := []byte{0x12, 0x34, 0x56}
	id, idx, tot, ct, ok := decodePartWire(raw)
	if !ok || id != 0 || idx != 0 || tot != 1 || !bytes.Equal(ct, raw) {
		t.Fatalf("legacy: id=%d idx=%d tot=%d ct=%x ok=%v", id, idx, tot, ct, ok)
	}
}

func TestMaxAppChunkPositive(t *testing.T) {
	n := maxAppChunk(MaxDatagramPayload, 1+FragHeaderSize, 2)
	if n < 100 {
		t.Fatalf("chunk budget too small: %d", n)
	}
}

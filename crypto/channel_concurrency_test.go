// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package crypto

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"sort"
	"sync"
	"testing"
)

// TestChannelEncryptConcurrent fans out N goroutines all calling Encrypt on
// the same Channel and verifies that every ciphertext decrypts to the
// expected plaintext at a unique, monotonic sequence number. This pins
// the property that Channel.Encrypt is safe for concurrent use across
// the multiple stream pumpers and the datagram pumper that share one
// AEAD inside a Session.
func TestChannelEncryptConcurrent(t *testing.T) {
	const N = 1024

	key := bytes.Repeat([]byte{0x42}, 32)
	send, err := NewChannel(key, key)
	if err != nil {
		t.Fatal(err)
	}
	recv, err := NewChannel(key, key)
	if err != nil {
		t.Fatal(err)
	}

	type item struct {
		seq        uint64
		plaintext  []byte
		ciphertext []byte
	}

	out := make([]item, N)

	var wg sync.WaitGroup
	wg.Add(N)
	for i := range N {
		go func() {
			defer wg.Done()
			plaintext := fmt.Appendf(nil, "msg-%06d", i)
			ct := send.Encrypt(plaintext)
			if len(ct) < 8 {
				t.Errorf("ciphertext too short: %d", len(ct))
				return
			}
			seq := binary.LittleEndian.Uint64(ct[:8])
			out[i] = item{seq: seq, plaintext: plaintext, ciphertext: ct}
		}()
	}
	wg.Wait()

	// Each Encrypt call must have received a unique sequence number in
	// 0..N-1 (atomic monotonic counter, no duplicates, no gaps).
	seen := make(map[uint64]bool, N)
	for _, it := range out {
		if it.seq >= N {
			t.Errorf("seq %d outside expected range 0..%d", it.seq, N-1)
		}
		if seen[it.seq] {
			t.Errorf("duplicate sequence number %d", it.seq)
		}
		seen[it.seq] = true
	}
	if len(seen) != N {
		t.Fatalf("expected %d unique sequence numbers, got %d", N, len(seen))
	}

	// Sort by sequence number and decrypt strictly in order. Each
	// ciphertext must round-trip to its original plaintext.
	sort.Slice(out, func(i, j int) bool { return out[i].seq < out[j].seq })
	for _, it := range out {
		got, err := recv.Decrypt(it.ciphertext)
		if err != nil {
			t.Fatalf("decrypt seq %d: %v", it.seq, err)
		}
		if !bytes.Equal(got, it.plaintext) {
			t.Fatalf("seq %d: plaintext mismatch: got %q want %q", it.seq, got, it.plaintext)
		}
	}
}

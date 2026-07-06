// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package cwire

import (
	"encoding/binary"
	"sync"
	"testing"
)

// TestChannelConcurrentEncryptDistinctNonces reproduces Fable-5 F1 (🎯T49):
// a single *cwire.Channel is shared across every stream pump and the
// datagram pump of a production Session, but pigeon_channel_encrypt advances
// ch->send_seq with a non-atomic C read-modify-write. Concurrent Encrypt
// callers therefore race the counter and can be handed the SAME 8-byte LE
// sequence prefix — which becomes the AES-GCM nonce. Two distinct
// plaintexts sealed under one (key, nonce) is a catastrophic AEAD break.
//
// Oracle: the 8-byte seq prefix of every produced ciphertext must be unique
// across all messages. The Go race detector cannot see the racing write
// (it lives in C), so the mechanism the audit names — a duplicated nonce —
// is asserted directly. Without the mutex added to Channel.Encrypt this
// fails with duplicate prefixes; with it, all prefixes are distinct.
func TestChannelConcurrentEncryptDistinctNonces(t *testing.T) {
	sendKey := make([]byte, 32)
	recvKey := make([]byte, 32)
	for i := range sendKey {
		sendKey[i] = byte(i + 1)
		recvKey[i] = byte(i + 100)
	}
	ch, err := NewChannel(sendKey, recvKey, "stream")
	if err != nil {
		t.Fatalf("NewChannel: %v", err)
	}

	const (
		workers = 64
		perWork = 256
		total   = workers * perWork
	)

	prefixes := make([][8]byte, total)
	var wg sync.WaitGroup
	start := make(chan struct{})
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			<-start // maximise contention on the shared counter
			for i := 0; i < perWork; i++ {
				ct, err := ch.Encrypt([]byte("payload"))
				if err != nil {
					t.Errorf("Encrypt: %v", err)
					return
				}
				var p [8]byte
				copy(p[:], ct[:8])
				prefixes[w*perWork+i] = p
			}
		}(w)
	}
	close(start)
	wg.Wait()

	seen := make(map[uint64]int, total)
	dups := 0
	for _, p := range prefixes {
		seq := binary.LittleEndian.Uint64(p[:])
		seen[seq]++
		if seen[seq] == 2 {
			dups++
		}
	}
	if len(seen) != total {
		t.Fatalf("nonce reuse: %d encrypts produced only %d distinct seq prefixes (%d collided) — (key,nonce) reuse",
			total, len(seen), dups)
	}
}

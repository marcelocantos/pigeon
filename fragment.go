// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package tern

import (
	"encoding/binary"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

// ErrDatagramTooLarge is returned when a payload exceeds the maximum
// fragmentable size (65535 fragments × max chunk size).
var ErrDatagramTooLarge = errors.New("datagram too large to fragment")

// Datagram framing: every datagram has a 1-byte prefix.
//
//	0x00 + payload                         — conn: whole datagram
//	0x40 + frag header + chunk             — conn: fragment
//	0x80 + 2-byte chanID + payload         — channel: whole datagram
//	0xC0 + 2-byte chanID + frag header + chunk — channel: fragment
//
// Fragment header (8 bytes):
//
//	[0:4]  message ID      (uint32, monotonic per sender)
//	[4:6]  fragment index  (uint16, 0-based)
//	[6:8]  total fragments (uint16, >= 2)
const (
	dgConnWhole    byte = 0x00
	dgConnFragment byte = 0x40
	dgChanWhole    byte = 0x80
	dgChanFragment byte = 0xC0
	dgPing         byte = 0x10 // health ping on direct path
	dgPong         byte = 0x11 // health pong on direct path
	fragHeaderSize      = 8
	chanIDSize          = 2
)

// DefaultFragmentTimeout is how long the receiver waits for all
// fragments of a message before discarding the partial assembly.
const DefaultFragmentTimeout = 5 * time.Second

// DefaultMaxDatagramPayload is the maximum payload per QUIC datagram.
// QUIC typically supports ~1200 bytes; we use a conservative default.
const DefaultMaxDatagramPayload = 1200

// reassembler tracks in-flight fragment assemblies for a connection.
type reassembler struct {
	mu         sync.Mutex
	assemblies map[uint32]*assembly
	timeout    time.Duration
	done       chan struct{}
	started    sync.Once
}

type assembly struct {
	fragments [][]byte
	total     int
	received  int
	deadline  time.Time
}

func newReassembler(timeout time.Duration, done chan struct{}) *reassembler {
	return &reassembler{
		assemblies: make(map[uint32]*assembly),
		timeout:    timeout,
		done:       done,
	}
}

// feed processes a fragment and returns the assembled payload if all
// fragments have arrived, or nil if the assembly is still incomplete.
func (r *reassembler) feed(msgID uint32, fragIdx, totalFrags int, payload []byte) []byte {
	r.started.Do(func() { go r.cleanupLoop() })

	r.mu.Lock()
	defer r.mu.Unlock()

	a, ok := r.assemblies[msgID]
	if !ok {
		a = &assembly{
			fragments: make([][]byte, totalFrags),
			total:     totalFrags,
			deadline:  time.Now().Add(r.timeout),
		}
		r.assemblies[msgID] = a
	}

	if fragIdx >= len(a.fragments) || a.fragments[fragIdx] != nil {
		return nil // out of range or duplicate
	}

	frag := make([]byte, len(payload))
	copy(frag, payload)
	a.fragments[fragIdx] = frag
	a.received++

	if a.received < a.total {
		return nil
	}

	// All fragments received — reassemble.
	size := 0
	for _, f := range a.fragments {
		size += len(f)
	}
	assembled := make([]byte, 0, size)
	for _, f := range a.fragments {
		assembled = append(assembled, f...)
	}
	delete(r.assemblies, msgID)
	return assembled
}

func (r *reassembler) cleanupLoop() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-r.done:
			return
		case now := <-ticker.C:
			r.mu.Lock()
			for id, a := range r.assemblies {
				if now.After(a.deadline) {
					delete(r.assemblies, id)
				}
			}
			r.mu.Unlock()
		}
	}
}

// sendFragmented splits data into fragments and sends each as a
// separate datagram with the given prefix byte + fragment header.
// For channel fragments, extraPrefix contains the 2-byte channel ID
// inserted between the prefix byte and the fragment header.
func sendFragmented(dg datagrammer, data []byte, maxPayload int, msgID uint32, prefix byte, extraPrefix []byte) error {
	overhead := 1 + len(extraPrefix) + fragHeaderSize
	maxChunk := maxPayload - overhead
	if maxChunk <= 0 {
		maxChunk = 1
	}

	total := (len(data) + maxChunk - 1) / maxChunk
	if total > 65535 {
		return ErrDatagramTooLarge
	}

	for i := 0; i < total; i++ {
		start := i * maxChunk
		end := start + maxChunk
		if end > len(data) {
			end = len(data)
		}
		chunk := data[start:end]

		frame := make([]byte, overhead+len(chunk))
		frame[0] = prefix
		off := 1
		copy(frame[off:], extraPrefix)
		off += len(extraPrefix)
		binary.BigEndian.PutUint32(frame[off:off+4], msgID)
		binary.BigEndian.PutUint16(frame[off+4:off+6], uint16(i))
		binary.BigEndian.PutUint16(frame[off+6:off+8], uint16(total))
		copy(frame[off+fragHeaderSize:], chunk)

		if err := dg.SendDatagram(frame); err != nil {
			return err
		}
	}
	return nil
}

// nextMsgID is a global counter for fragment message IDs.
var nextMsgID atomic.Uint32

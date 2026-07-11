// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

package pigeon

import (
	"encoding/binary"
	"errors"
	"sync/atomic"
)

// ErrDatagramTooLarge is returned when a payload would need more than
// 65535 parts under the current max datagram payload.
var ErrDatagramTooLarge = errors.New("datagram too large to fragment")

// DatagramPart is one delivered piece of a logical datagram message
// (🎯T59). Whole messages arrive as a single part with Index=0, Total=1.
// Multi-part messages share MsgID; parts are delivered as each QUIC
// datagram arrives — the library does not reassemble.
type DatagramPart struct {
	MsgID   uint32
	Index   uint16
	Total   uint16
	Payload []byte
}

// Outer QUIC-datagram framing. Each part carries its own AEAD ciphertext
// so it can be decrypted and delivered independently:
//
//	0x00 + AEAD(cid||payload)                         — whole (Total=1)
//	0x40 + msgID(4)+idx(2)+total(2) + AEAD(cid||chunk) — fragment
//
// Constants: DgConnWhole, DgConnFragment, FragHeaderSize, MaxDatagramPayload
// from protocol/session.yaml via session_gen.go.

// nextFragMsgID is a process-wide monotonic counter for logical messages.
var nextFragMsgID atomic.Uint32

// aeadOverhead is the AEAD framing cost on the wire (8-byte seq prefix +
// 16-byte GCM tag). Used only to size chunks so framed parts fit in
// MaxDatagramPayload.
const aeadOverhead = 8 + 16

// maxAppChunk returns the largest app-payload chunk that, once AEAD'd and
// wrapped with the given outer overhead, still fits in maxPayload.
func maxAppChunk(maxPayload, outerOverhead, channelIDLen int) int {
	// outer + AEAD(seq||ct||tag) where ct encrypts (cid_varint||chunk).
	// Approximate cid as up to 10 varint bytes worst case; callers pass
	// a tighter bound when known.
	budget := maxPayload - outerOverhead - aeadOverhead - channelIDLen
	if budget < 1 {
		return 1
	}
	return budget
}

// encodePartWire builds one QUIC datagram for a logical part.
// ct is the already-encrypted AEAD blob for this part's plaintext.
func encodePartWire(msgID uint32, index, total uint16, ct []byte) []byte {
	if total <= 1 {
		frame := make([]byte, 1+len(ct))
		frame[0] = DgConnWhole
		copy(frame[1:], ct)
		return frame
	}
	frame := make([]byte, 1+FragHeaderSize+len(ct))
	frame[0] = DgConnFragment
	binary.BigEndian.PutUint32(frame[1:5], msgID)
	binary.BigEndian.PutUint16(frame[5:7], index)
	binary.BigEndian.PutUint16(frame[7:9], total)
	copy(frame[1+FragHeaderSize:], ct)
	return frame
}

// decodePartWire parses one QUIC datagram into outer metadata + AEAD blob.
// For whole frames, Index=0 and Total=1; MsgID is 0 (unused for singles).
func decodePartWire(data []byte) (msgID uint32, index, total uint16, ct []byte, ok bool) {
	if len(data) < 2 {
		return 0, 0, 0, nil, false
	}
	switch data[0] {
	case DgConnWhole:
		ct = data[1:]
		return 0, 0, 1, ct, true
	case DgConnFragment:
		if len(data) < 1+FragHeaderSize {
			return 0, 0, 0, nil, false
		}
		msgID = binary.BigEndian.Uint32(data[1:5])
		index = binary.BigEndian.Uint16(data[5:7])
		total = binary.BigEndian.Uint16(data[7:9])
		if total < 2 || index >= total {
			return 0, 0, 0, nil, false
		}
		ct = data[1+FragHeaderSize:]
		return msgID, index, total, ct, true
	default:
		// Pre-T59 unframed AEAD: treat as a whole message.
		return 0, 0, 1, data, true
	}
}

// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Package cwire is the cgo bridge to libpigeon's wire-format helpers
// (varint, stream-header, AEAD-channel-id datagram framing).
//
// This package is the foundation for 🎯T34 — folding the Go peer
// library into a thin cgo wrapper over libpigeon. Today it covers
// only the pure wire-format primitives; once the multi-stream
// ngtcp2 transport lands (T32 step 3), this same bridge will expose
// pigeon_session_open_stream et al. and the remainder of the Go
// pigeon package can be reimplemented on top.
//
// Build mode: cgo. Pulls in dist/pigeon.c via CFLAGS / LDFLAGS so
// the Go build is fully self-contained — no need for a separate
// `make amalgamate` step before `go test ./cwire`.
package cwire

// #cgo CFLAGS: -I${SRCDIR}/../dist -DPIGEON_CRYPTO_LIBSODIUM
// #cgo CFLAGS: -I${SRCDIR}/../c/vendor/build/include
// #cgo LDFLAGS: ${SRCDIR}/../c/vendor/build/lib/libsodium.a
// #include <stdlib.h>
// #include <string.h>
// #include "pigeon.h"
import "C"

import (
	"errors"
	"unsafe"
)

// EncodeUvarint encodes v in Go's binary.PutUvarint format using the
// C implementation; useful for proving wire-byte equivalence between
// the Go and C peer libraries.
func EncodeUvarint(v uint64) []byte {
	var buf [10]byte
	n := C.pigeon_uvarint_encode(C.uint64_t(v),
		(*C.uint8_t)(unsafe.Pointer(&buf[0])), 10)
	if n < 0 {
		return nil
	}
	out := make([]byte, int(n))
	copy(out, buf[:int(n)])
	return out
}

// DecodeUvarint mirrors EncodeUvarint via the C implementation.
// Returns the decoded value and the number of bytes consumed (1..10),
// or an error if the input is truncated or malformed.
func DecodeUvarint(buf []byte) (uint64, int, error) {
	if len(buf) == 0 {
		return 0, 0, errors.New("cwire: empty buf")
	}
	var v C.uint64_t
	n := C.pigeon_uvarint_decode(
		(*C.uint8_t)(unsafe.Pointer(&buf[0])),
		C.size_t(len(buf)),
		&v)
	switch {
	case n > 0:
		return uint64(v), int(n), nil
	case n == 0:
		return 0, 0, errors.New("cwire: truncated varint")
	default:
		return 0, 0, errors.New("cwire: malformed varint")
	}
}

// EncodeStreamHeader emits the post-T22 first-message stream header
// using the C library's encoder. The bytes are guaranteed
// byte-for-byte identical to Go's internal encodeStreamHeader by the
// shared cross-language test vectors in wire_vectors_test.go and
// c/test/test_pigeon.c::test_stream_header.
func EncodeStreamHeader(isBackend bool, tag uint32, name string) ([]byte, error) {
	var out [4 + 10 + 256]byte // matches PIGEON_MAX_STREAM_HEADER
	var cName *C.char
	var cNameLen C.size_t
	if name != "" {
		cName = C.CString(name)
		defer C.free(unsafe.Pointer(cName))
		cNameLen = C.size_t(len(name))
	}
	var cIsBackend C.bool
	if isBackend {
		cIsBackend = C.bool(true)
	}
	n := C.pigeon_wire_stream_header_encode(
		cIsBackend, C.uint32_t(tag),
		cName, cNameLen,
		(*C.uint8_t)(unsafe.Pointer(&out[0])), C.size_t(len(out)))
	if n < 0 {
		return nil, errors.New("cwire: encode_stream_header failed")
	}
	res := make([]byte, int(n))
	copy(res, out[:int(n)])
	return res, nil
}

// DecodeBackendStreamHeader parses a backend-side header
// ([4-byte tag][varint name-len][name]) via the C decoder. Returns
// the decoded tag, name, and bytes consumed.
func DecodeBackendStreamHeader(buf []byte) (uint32, string, int, error) {
	if len(buf) == 0 {
		return 0, "", 0, errors.New("cwire: empty buf")
	}
	var tag C.uint32_t
	var nameBuf [256]C.char
	var nameLen C.size_t
	n := C.pigeon_wire_stream_header_decode_backend(
		(*C.uint8_t)(unsafe.Pointer(&buf[0])),
		C.size_t(len(buf)),
		&tag,
		&nameBuf[0], C.size_t(len(nameBuf)),
		&nameLen)
	if n < 0 {
		return 0, "", 0, errors.New("cwire: decode_backend_stream_header failed")
	}
	name := C.GoStringN(&nameBuf[0], C.int(nameLen))
	return uint32(tag), name, int(n), nil
}

// DecodeClientStreamHeader parses a client-side header
// ([varint name-len][name]) via the C decoder.
func DecodeClientStreamHeader(buf []byte) (string, int, error) {
	if len(buf) == 0 {
		return "", 0, errors.New("cwire: empty buf")
	}
	var nameBuf [256]C.char
	var nameLen C.size_t
	n := C.pigeon_wire_stream_header_decode_client(
		(*C.uint8_t)(unsafe.Pointer(&buf[0])),
		C.size_t(len(buf)),
		&nameBuf[0], C.size_t(len(nameBuf)),
		&nameLen)
	if n < 0 {
		return "", 0, errors.New("cwire: decode_client_stream_header failed")
	}
	name := C.GoStringN(&nameBuf[0], C.int(nameLen))
	return name, int(n), nil
}

// EncodeDatagram composes the AEAD-encrypted datagram body, with an
// optional 4-byte clientTag prefix on the backend side. The supplied
// (sendKey, recvKey) is used to construct a fresh pigeon_channel.
//
// This is a thin wrapper for cross-language byte-parity testing — Go
// callers that already have a *crypto.Channel should encrypt and
// frame in Go directly.
func EncodeDatagram(sendKey, recvKey []byte, isBackend bool, clientTag uint32, channelID uint64, payload []byte) ([]byte, error) {
	if len(sendKey) != 32 || len(recvKey) != 32 {
		return nil, errors.New("cwire: keys must be 32 bytes")
	}
	var ch C.pigeon_channel
	C.pigeon_channel_init(&ch,
		(*C.uint8_t)(unsafe.Pointer(&sendKey[0])),
		(*C.uint8_t)(unsafe.Pointer(&recvKey[0])),
		C.PIGEON_MODE_DATAGRAMS)

	out := make([]byte, len(payload)+128)
	var pPayload *C.uint8_t
	if len(payload) > 0 {
		pPayload = (*C.uint8_t)(unsafe.Pointer(&payload[0]))
	}
	var cIsBackend C.bool
	if isBackend {
		cIsBackend = C.bool(true)
	}
	n := C.pigeon_encode_datagram(&ch, cIsBackend, C.uint32_t(clientTag),
		C.uint64_t(channelID), pPayload, C.size_t(len(payload)),
		(*C.uint8_t)(unsafe.Pointer(&out[0])), C.size_t(len(out)))
	if n < 0 {
		return nil, errors.New("cwire: encode_datagram failed")
	}
	return out[:int(n)], nil
}

// DecodeDatagram inverts EncodeDatagram. Returns (clientTag — only
// meaningful on backend side; 0 otherwise), the decoded channel ID,
// the application payload, or an error.
func DecodeDatagram(sendKey, recvKey []byte, isBackend bool, wire []byte) (uint32, uint64, []byte, error) {
	if len(sendKey) != 32 || len(recvKey) != 32 {
		return 0, 0, nil, errors.New("cwire: keys must be 32 bytes")
	}
	if len(wire) == 0 {
		return 0, 0, nil, errors.New("cwire: empty wire")
	}
	var ch C.pigeon_channel
	C.pigeon_channel_init(&ch,
		(*C.uint8_t)(unsafe.Pointer(&sendKey[0])),
		(*C.uint8_t)(unsafe.Pointer(&recvKey[0])),
		C.PIGEON_MODE_DATAGRAMS)

	var tag C.uint32_t
	var cid C.uint64_t
	out := make([]byte, len(wire))
	var cIsBackend C.bool
	if isBackend {
		cIsBackend = C.bool(true)
	}
	n := C.pigeon_decode_datagram(&ch, cIsBackend,
		(*C.uint8_t)(unsafe.Pointer(&wire[0])), C.size_t(len(wire)),
		&tag, &cid,
		(*C.uint8_t)(unsafe.Pointer(&out[0])), C.size_t(len(out)))
	if n < 0 {
		return 0, 0, nil, errors.New("cwire: decode_datagram failed")
	}
	return uint32(tag), uint64(cid), out[:int(n)], nil
}

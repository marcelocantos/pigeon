// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0
//
// In-process loopback transport for the pigeon peer library. Two
// endpoints route streams and datagrams to each other in memory, with
// no I/O. All language wrappers (cgo, SwiftPM, JNI) use this to drive
// pigeon_session / pigeon_stream / pigeon_datagram round-trip tests
// without pulling in ngtcp2 or a live relay.
//
// Heap-allocated so host-language callers can manage lifetime through
// their idiomatic patterns (Go finalisers, Swift Unmanaged, JNI
// pointer-return). Pair two endpoints with pigeon_loopback_pair, then
// fill a pigeon_transport from each via pigeon_loopback_fill_transport.

#ifndef PIGEON_LOOPBACK_H
#define PIGEON_LOOPBACK_H

#include "pigeon.h"

#ifdef __cplusplus
extern "C" {
#endif

typedef struct pigeon_loopback_endpoint pigeon_loopback_endpoint;

// Allocate a fresh endpoint. Returns NULL on allocation failure.
pigeon_loopback_endpoint *pigeon_loopback_new(void);

// Pair two endpoints together. Each endpoint's send/open routes to
// the other's receive/accept queue.
void pigeon_loopback_pair(pigeon_loopback_endpoint *a,
                          pigeon_loopback_endpoint *b);

// Fill a pigeon_transport with the loopback vtable for endpoint e.
// The pigeon_transport's userdata pointer is set to e.
void pigeon_loopback_fill_transport(pigeon_transport *t,
                                    pigeon_loopback_endpoint *e);

// Accept the next inbound stream and read its first message (the
// stream-header). Writes the header bytes to hdr (up to hdr_len) and
// the consumed length to *hdr_out_len. Returns the opaque transport
// handle through *out_handle, or -1 if no stream is pending or the
// header read fails.
int pigeon_loopback_accept_with_header(pigeon_loopback_endpoint *e,
                                       pigeon_stream_handle **out_handle,
                                       uint8_t *hdr, size_t hdr_len,
                                       size_t *hdr_out_len);

// Free an endpoint allocated by pigeon_loopback_new. Idempotent on NULL.
void pigeon_loopback_free(pigeon_loopback_endpoint *e);

#ifdef __cplusplus
}
#endif

#endif // PIGEON_LOOPBACK_H

// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

#ifndef PIGEON_LOOPBACK_H
#define PIGEON_LOOPBACK_H

// In-process loopback transport, mirrored from
// c/test/test_pigeon.c::loopback_make_transport. Two endpoints route
// streams and datagrams to each other in memory, with no I/O. The
// Swift bindings use this to drive pigeon_session_*, pigeon_stream_*,
// and pigeon_datagram_* round-trip tests without pulling in ngtcp2.
//
// All the bookkeeping (stream slot table, pending-message ringbuffers,
// accept queues) is allocated on the heap inside pigeon_loopback_new
// and freed by pigeon_loopback_free. The pigeon_transport vtable is
// filled in by pigeon_loopback_fill_transport so callers can pass it
// straight to pigeon_session_init.

#include "pigeon.h"

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
//
// The Swift test harness uses this on the receiving side to mimic
// what a real backend's accept loop would do — strip the unencrypted
// name-binding header before AEAD-decrypting the rest.
int pigeon_loopback_accept_with_header(pigeon_loopback_endpoint *e,
                                       pigeon_stream_handle **out_handle,
                                       uint8_t *hdr, size_t hdr_len,
                                       size_t *hdr_out_len);

// Free an endpoint allocated by pigeon_loopback_new. Idempotent on
// NULL.
void pigeon_loopback_free(pigeon_loopback_endpoint *e);

#endif // PIGEON_LOOPBACK_H

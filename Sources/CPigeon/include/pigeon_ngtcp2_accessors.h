// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0
//
// Small Swift-facing accessor helpers around the ngtcp2 transport
// struct. The `pigeon_ngtcp2_transport` value is ~1.1 MiB (17 × 64
// KiB stream ringbufs); any Swift accessor that materialises a copy
// on the stack (e.g. `withUnsafeBytes(of: &ptr.pointee.last_error)`)
// blows the cooperative-pool thread stack (~544 KiB). These C
// helpers return raw pointers into the heap struct that Swift can
// read with `String(cString:)` without ever copying the parent.
//
// Implementation lives alongside the ngtcp2 transport build unit
// in Sources/CPigeon/cpigeon_ngtcp2.c.

#ifndef PIGEON_NGTCP2_ACCESSORS_H
#define PIGEON_NGTCP2_ACCESSORS_H

#include "pigeon_ngtcp2.h"

#ifdef __cplusplus
extern "C" {
#endif

const char *
pigeon_ngtcp2_transport_last_error(const pigeon_ngtcp2_transport *t);

const char *
pigeon_ngtcp2_transport_instance_id(const pigeon_ngtcp2_transport *t);

#ifdef __cplusplus
}
#endif

#endif // PIGEON_NGTCP2_ACCESSORS_H

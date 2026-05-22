// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0
//
// CPigeon ngtcp2 transport build unit — pulls in the canonical
// ngtcp2 transport implementation (c/src/ngtcp2_transport.c) and
// the high-level listener wrappers (c/src/listener_ngtcp2.c) so the
// Swift Pigeon target gains a real QUIC transport for connecting
// to a Go relay (🎯T37).
//
// dist/pigeon.c (compiled by cpigeon.c) does NOT include this file
// — the ngtcp2 transport opts in at link time and brings in the
// vendored static libraries libngtcp2, libngtcp2_crypto_quictls,
// libssl and libcrypto. The SwiftPM target adds those to its linker
// settings.
//
// Header search paths configured by Package.swift:
//   - ../../c/include            (canonical pigeon headers)
//   - ../../c/vendor/build/include (vendored ngtcp2 + openssl)
//   - ../../dist                  (amalgamated pigeon.h)

// The canonical pigeon.h is supplied via the amalgamated dist/pigeon.h
// (header search path) which cpigeon.c brings in via the pigeon_loopback
// forwarder. Including the listener wrappers requires the canonical
// c/include/pigeon/pigeon.h to be the one in scope — both headers are
// guard-compatible (PIGEON_H) so re-inclusion is a no-op.

#include "../../c/src/ngtcp2_transport.c"
#include "../../c/src/listener_ngtcp2.c"

// ---- Swift-side accessor helpers ----
//
// Swift cannot safely touch fields of pigeon_ngtcp2_transport via
// `pointer.pointee.field`: that struct is ~1.1 MiB (17 × 64 KiB stream
// ringbufs) and any accessor that materialises a copy on the stack
// blows the cooperative-pool ~544 KiB thread stack. These tiny C
// helpers return pointers to the fixed character buffers, which Swift
// can then read with String(cString:) without ever copying the whole
// struct.

const char *
pigeon_ngtcp2_transport_last_error(const pigeon_ngtcp2_transport *t)
{
    return t ? t->last_error : "";
}

const char *
pigeon_ngtcp2_transport_instance_id(const pigeon_ngtcp2_transport *t)
{
    return t ? t->instance_id : "";
}

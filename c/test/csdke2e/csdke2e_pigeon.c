// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0
//
// Wraps the C library's ngtcp2-only pieces into the cgo build of
// package csdke2e:
//
//   * c/src/ngtcp2_transport.c — QUIC transport backed by vendored
//                                ngtcp2 + quictls (linked via
//                                cgo LDFLAGS to the static .a libs).
//   * c/src/listener_ngtcp2.c  — pigeon_register / pigeon_connect
//                                fused wrappers that bind the ngtcp2
//                                transport to the high-level API.
//
// The amalgamated core (dist/pigeon.c) is NOT compiled here — the
// package transitively imports github.com/marcelocantos/pigeon which
// imports cwire, and cwire's cgo build already compiles dist/pigeon.c
// into the same test binary. Duplicating the compilation would emit
// duplicate symbols at link time. Both .c files included below
// reference pigeon_* symbols by linkage only; the actual definitions
// resolve via the cwire compilation unit in the same binary.
//
// Cgo build only — guarded by the same csdke2e build tag as the
// package's .go files so `go vet ./...` / `go build ./...` (which
// don't pass -tags csdke2e) skip it cleanly.

#ifdef CSDKE2E_BUILD

#include "ngtcp2_transport.c"
#include "listener_ngtcp2.c"

#endif

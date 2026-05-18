// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0
//
// Wraps the C library into the cgo build of package csdke2e:
//
//   * dist/pigeon.c       — amalgamated core (crypto, channel,
//                           session, listener, activation, loopback).
//   * c/src/ngtcp2_transport.c — QUIC transport backed by vendored
//                                ngtcp2 + quictls (linked via
//                                cgo LDFLAGS to the static .a libs).
//   * c/src/listener_ngtcp2.c  — pigeon_register / pigeon_connect
//                                fused wrappers that bind the ngtcp2
//                                transport to the high-level API.
//
// All three files are compiled into the cgo TU. Build tag csdke2e
// keeps this file (and the rest of the package) out of the default
// Go build — the vendored ngtcp2 / quictls static libs are not
// present unless `make build-vendor-deps` has run.
//
// Cgo build only — guarded by the same csdke2e build tag as the
// package's .go files so `go vet ./...` / `go build ./...` (which
// don't pass -tags csdke2e) skip it cleanly.

#ifdef CSDKE2E_BUILD

#include "pigeon.c"
#include "ngtcp2_transport.c"
#include "listener_ngtcp2.c"

#endif

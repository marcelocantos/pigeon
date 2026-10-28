#!/usr/bin/env bash
# Copyright 2026 Marcelo Cantos
# SPDX-License-Identifier: Apache-2.0
# Build vendored dependencies in this checkout, and force cgo to recompile
# its #included ngtcp2_transport.c (Go does not track that include in cache).
set -euo pipefail
make build-vendor-deps
bash c/amalgamate.sh dist
make test-c-ngtcp2
# The preceding Make target includes the standalone C unit tests. This
# forced rebuild is the authoritative live relay+peer oracle: startRelay
# creates a real in-process QUIC server and fails (never skips) on errors.
go test -a -tags csdke2e -count=1 -timeout=120s -v ./c/test/csdke2e/

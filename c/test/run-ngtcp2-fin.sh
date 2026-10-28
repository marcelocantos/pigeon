#!/usr/bin/env bash
# Copyright 2026 Marcelo Cantos
# SPDX-License-Identifier: Apache-2.0
set -euo pipefail
if [[ -n "${PIGEON_VENDOR_PREBUILT:-}" ]]; then
  # Optional local prebuilt cache; the gate still builds the exact-tip C
  # source in its clean worktree. Only use with matching pinned submodules.
  mkdir -p c/vendor/build
  cp -R "$PIGEON_VENDOR_PREBUILT/include" "$PIGEON_VENDOR_PREBUILT/lib" c/vendor/build/
else
  make build-vendor-deps
fi
make amalgamate
clang -Ic/vendor/build/include -Ic/include -Idist \
  c/test/test_ngtcp2_fin.c dist/pigeon.c -DPIGEON_CRYPTO_LIBSODIUM \
  c/vendor/build/lib/libngtcp2_crypto_quictls.a \
  c/vendor/build/lib/libngtcp2.a c/vendor/build/lib/libssl.a \
  c/vendor/build/lib/libcrypto.a c/vendor/build/lib/libsodium.a \
  -lpthread -ldl -o c/test/test_ngtcp2_fin
./c/test/test_ngtcp2_fin

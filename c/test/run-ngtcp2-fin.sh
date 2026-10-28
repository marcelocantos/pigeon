#!/usr/bin/env bash
# Copyright 2026 Marcelo Cantos
# SPDX-License-Identifier: Apache-2.0
set -euo pipefail
make build-vendor-deps
make amalgamate
clang -Ic/vendor/build/include -Ic/include -Idist \
  c/test/test_ngtcp2_fin.c dist/pigeon.c -DPIGEON_CRYPTO_LIBSODIUM \
  c/vendor/build/lib/libngtcp2_crypto_quictls.a \
  c/vendor/build/lib/libngtcp2.a c/vendor/build/lib/libssl.a \
  c/vendor/build/lib/libcrypto.a c/vendor/build/lib/libsodium.a \
  -lpthread -ldl -o c/test/test_ngtcp2_fin
./c/test/test_ngtcp2_fin

#!/usr/bin/env bash
# Copyright 2026 Marcelo Cantos
# SPDX-License-Identifier: Apache-2.0
#
# Build vendored C dependencies: quictls/openssl and ngtcp2.
# Outputs static libraries under c/vendor/build/.
#
# Usage:
#   ./c/vendor/build.sh          # build all
#   ./c/vendor/build.sh openssl  # build only openssl
#   ./c/vendor/build.sh ngtcp2   # build only ngtcp2 (requires openssl done)
#
# Artifacts:
#   c/vendor/build/lib/libssl.a
#   c/vendor/build/lib/libcrypto.a
#   c/vendor/build/lib/libngtcp2.a
#   c/vendor/build/lib/libngtcp2_crypto_quictls.a
#   c/vendor/build/include/  (merged headers)
#
# Requirements: cmake, make, perl (for openssl Configure)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

VENDOR="$SCRIPT_DIR"
BUILD="$VENDOR/build"
OPENSSL_SRC="$VENDOR/github.com/quictls/openssl"
NGTCP2_SRC="$VENDOR/github.com/ngtcp2/ngtcp2"

JOBS="$(sysctl -n hw.logicalcpu 2>/dev/null || nproc 2>/dev/null || echo 4)"

mkdir -p "$BUILD/lib" "$BUILD/include"

build_openssl() {
    echo "==> Building quictls/openssl..."
    cd "$OPENSSL_SRC"
    # Configure for the current host, static libs only.
    ./Configure \
        --prefix="$BUILD" \
        --openssldir="$BUILD/etc/ssl" \
        no-shared \
        no-tests \
        enable-quic \
        -fPIC
    make build_libs
    make install_dev   # installs headers + static libs to $BUILD
    echo "==> quictls/openssl built."
}

build_ngtcp2() {
    echo "==> Building ngtcp2..."
    local bdir="$BUILD/ngtcp2-build"
    mkdir -p "$bdir"
    cd "$bdir"
    cmake "$NGTCP2_SRC" \
        -DCMAKE_BUILD_TYPE=Release \
        -DCMAKE_INSTALL_PREFIX="$BUILD" \
        -DENABLE_SHARED_LIB=OFF \
        -DENABLE_STATIC_LIB=ON \
        -DENABLE_OPENSSL=ON \
        -DOPENSSL_ROOT_DIR="$BUILD" \
        -DOPENSSL_INCLUDE_DIR="$BUILD/include" \
        -DOPENSSL_CRYPTO_LIBRARY="$BUILD/lib/libcrypto.a" \
        -DOPENSSL_SSL_LIBRARY="$BUILD/lib/libssl.a" \
        -DENABLE_EXAMPLES=OFF \
        -DENABLE_GNUTLS=OFF \
        -DENABLE_BORINGSSL=OFF \
        -DENABLE_WOLFSSL=OFF \
        -DENABLE_PICOTLS=OFF \
        -DBUILD_TESTING=OFF
    make
    make install
    echo "==> ngtcp2 built."
}

TARGET="${1:-all}"

case "$TARGET" in
    openssl)
        build_openssl
        ;;
    ngtcp2)
        build_ngtcp2
        ;;
    all)
        build_openssl
        build_ngtcp2
        ;;
    *)
        echo "Unknown target: $TARGET" >&2
        echo "Usage: $0 [all|openssl|ngtcp2]" >&2
        exit 1
        ;;
esac

echo ""
echo "==> Vendored libraries ready in $BUILD"
echo "    Libs:    $(ls "$BUILD/lib/"*.a 2>/dev/null | tr '\n' ' ')"

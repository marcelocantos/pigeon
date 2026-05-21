#!/usr/bin/env bash
# Copyright 2026 Marcelo Cantos
# SPDX-License-Identifier: Apache-2.0
#
# Cross-compile vendored libsodium for Android ABIs.
#
# Usage:
#   ./c/vendor/build-android.sh [arm64-v8a|x86_64|all]
#
# Output layout (one per ABI):
#   c/vendor/build-android/arm64-v8a/lib/libsodium.a
#   c/vendor/build-android/arm64-v8a/include/sodium.h
#   c/vendor/build-android/x86_64/lib/libsodium.a
#   c/vendor/build-android/x86_64/include/sodium.h
#
# Requirements:
#   ANDROID_NDK_HOME (or ANDROID_NDK_ROOT) pointing at an NDK r25+ install,
#   plus autoconf, automake, glibtoolize (gnu libtool).
#
# We intentionally do NOT call libsodium's dist-build/android-*.sh
# scripts. macOS host's ar/ranlib choke on ELF object files produced
# by the NDK toolchain, silently turning libsodium.a into a 96-byte
# stub. We invoke ./configure directly with the NDK's llvm-ar /
# llvm-ranlib in $AR/$RANLIB and --disable-shared --enable-static so
# only the static archive is produced.
#
# NDK minimum API: 33 (matches the Kotlin SDK floor in android/pigeon
# build.gradle.kts and the project's JDK 17+ / Android API 33+ note).

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
VENDOR="$SCRIPT_DIR"
LIBSODIUM_SRC="$VENDOR/github.com/jedisct1/libsodium"
OUT_ROOT="$VENDOR/build-android"

# --- NDK resolution -------------------------------------------------------
NDK="${ANDROID_NDK_HOME:-${ANDROID_NDK_ROOT:-}}"
if [ -z "$NDK" ]; then
    sdk_root="${ANDROID_SDK_ROOT:-${ANDROID_HOME:-$HOME/Library/Android/sdk}}"
    if [ -d "$sdk_root/ndk" ]; then
        latest="$(ls -1 "$sdk_root/ndk" | sort -V | tail -1)"
        NDK="$sdk_root/ndk/$latest"
    fi
fi
if [ -z "$NDK" ] || [ ! -d "$NDK" ]; then
    echo "error: Android NDK not found." >&2
    echo "  Set ANDROID_NDK_HOME or install via sdkmanager 'ndk;<version>'." >&2
    exit 1
fi
export ANDROID_NDK_HOME="$NDK"
echo "==> Using NDK: $ANDROID_NDK_HOME"

NDK_API="${ANDROID_NDK_API:-33}"

if [ ! -d "$LIBSODIUM_SRC" ]; then
    echo "error: libsodium submodule not initialised at $LIBSODIUM_SRC" >&2
    echo "  run: git submodule update --init $LIBSODIUM_SRC" >&2
    exit 1
fi

host_tag="$(uname | tr '[:upper:]' '[:lower:]')-x86_64"
toolchain_bin="$NDK/toolchains/llvm/prebuilt/${host_tag}/bin"
toolchain_sysroot="$NDK/toolchains/llvm/prebuilt/${host_tag}/sysroot"
if [ ! -d "$toolchain_bin" ] || [ ! -d "$toolchain_sysroot" ]; then
    echo "error: NDK toolchain not found under $NDK/toolchains/llvm/prebuilt/$host_tag" >&2
    exit 1
fi

# Run autogen once (it caches ./configure). All ABIs share that.
if [ ! -x "$LIBSODIUM_SRC/configure" ]; then
    echo "==> Running libsodium autogen.sh..."
    (cd "$LIBSODIUM_SRC" && ./autogen.sh -s)
fi

build_abi() {
    local abi="$1"            # AGP jniLibs name
    local host_triple="$2"    # autotools --host
    local clang_triple="$3"   # NDK clang prefix (e.g. aarch64-linux-android)
    local march="$4"          # -march flag value

    echo
    echo "==> Building libsodium for Android $abi (API $NDK_API)..."

    local out="$OUT_ROOT/$abi"
    local build_dir="$LIBSODIUM_SRC/build-android-$abi"
    rm -rf "$build_dir" "$out"
    mkdir -p "$build_dir" "$out/lib" "$out/include"

    # Per-ABI env. CC points at the API-versioned NDK clang wrapper;
    # AR/RANLIB/STRIP at the NDK's llvm-* tools so static archives are
    # valid for the host running this script (macOS in our case).
    (
        cd "$build_dir"
        export CC="$toolchain_bin/${clang_triple}${NDK_API}-clang"
        export AR="$toolchain_bin/llvm-ar"
        export RANLIB="$toolchain_bin/llvm-ranlib"
        export STRIP="$toolchain_bin/llvm-strip"
        export CFLAGS="-Os -march=${march} -fPIC"
        # 16K page-size alignment future-proofs the archive for Android
        # 15+ devices; static-link consumers inherit this implicitly.
        export LDFLAGS="-Wl,-z,max-page-size=16384"

        if [ ! -x "$CC" ]; then
            echo "error: NDK clang wrapper not found: $CC" >&2
            echo "  Either bump --api or install an NDK that ships the wrapper." >&2
            exit 1
        fi

        ../configure \
            --host="$host_triple" \
            --prefix="$out" \
            --disable-shared \
            --enable-static \
            --disable-pie \
            --disable-soname-versions \
            --with-sysroot="$toolchain_sysroot"

        local nproc
        nproc="$(sysctl -n hw.logicalcpu 2>/dev/null || getconf _NPROCESSORS_ONLN 2>/dev/null || echo 4)"
        make -j"$nproc"
        make install
    )

    if [ ! -s "$out/lib/libsodium.a" ]; then
        echo "error: libsodium.a missing or empty at $out/lib/libsodium.a" >&2
        exit 1
    fi
    # Sanity: the archive must contain real object members. macOS ar
    # would silently produce a 96-byte SYMDEF-only stub if RANLIB were
    # the host's; check for size as a cheap guard.
    local sz
    sz="$(stat -f%z "$out/lib/libsodium.a" 2>/dev/null || stat -c%s "$out/lib/libsodium.a")"
    if [ "$sz" -lt 100000 ]; then
        echo "error: libsodium.a suspiciously small ($sz bytes) — likely a stub." >&2
        exit 1
    fi
    echo "==> libsodium ($abi) installed at $out  ($sz bytes)"
}

TARGETS="${1:-all}"
case "$TARGETS" in
    arm64-v8a)
        build_abi arm64-v8a aarch64-linux-android aarch64-linux-android armv8-a+crypto
        ;;
    x86_64)
        build_abi x86_64 x86_64-linux-android x86_64-linux-android westmere
        ;;
    all)
        # Sequential — each build uses its own build-android-<abi>/
        # directory but shares the libsodium source tree; configure
        # writes a config.cache that we deliberately avoid by giving
        # each build its own out-of-tree dir.
        build_abi arm64-v8a aarch64-linux-android aarch64-linux-android armv8-a+crypto
        build_abi x86_64 x86_64-linux-android x86_64-linux-android westmere
        ;;
    *)
        echo "Unknown target: $TARGETS" >&2
        echo "Usage: $0 [all|arm64-v8a|x86_64]" >&2
        exit 1
        ;;
esac

echo
echo "==> Android libsodium artefacts ready under $OUT_ROOT/"
ls -1 "$OUT_ROOT"

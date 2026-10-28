#!/usr/bin/env bash
# Copyright 2026 Marcelo Cantos
# SPDX-License-Identifier: Apache-2.0
set -euo pipefail
cd "$(dirname "$0")/.."
# git's safe.directory is set in jitpack.yml before any submodule operation.
# Autotools is needed to cross-compile pinned libsodium; quictls needs Perl.
missing=()
for cmd in autoconf automake libtoolize perl make; do
    command -v "$cmd" >/dev/null || missing+=("$cmd")
done
if ((${#missing[@]})); then
    if ! command -v apt-get >/dev/null; then
        echo "Missing build tools: ${missing[*]} (apt-get unavailable)" >&2
        exit 1
    fi
    if [ "$(id -u)" -eq 0 ]; then
        apt=(apt-get)
    elif command -v sudo >/dev/null && sudo -n true 2>/dev/null; then
        apt=(sudo -n apt-get)
    else
        echo "Missing build tools: ${missing[*]} (no root or noninteractive sudo)" >&2
        exit 1
    fi
    "${apt[@]}" update
    "${apt[@]}" install -y --no-install-recommends autoconf automake libtool perl make
fi
sdk="${ANDROID_HOME:-${ANDROID_SDK_ROOT:-$HOME/android-sdk}}"
mkdir -p "$sdk"
manager="${ANDROID_SDK_MANAGER:-}"
if [ -z "$manager" ]; then
    for path in "$sdk/cmdline-tools/latest/bin/sdkmanager" "$sdk/tools/bin/sdkmanager"; do
        if [ -x "$path" ]; then manager="$path"; break; fi
    done
fi
if [ -z "$manager" ]; then manager="$(command -v sdkmanager || true)"; fi
if [ -z "$manager" ]; then echo "Android sdkmanager unavailable under $sdk" >&2; exit 1; fi
# Licenses and NDK are provisioned in JitPack's ephemeral build container.
# Do not pipe yes into sdkmanager: pipefail would turn a successful install
# into SIGPIPE when sdkmanager exits before consuming all input.
licenses="$(mktemp)"
trap 'rm -f "$licenses"' EXIT
for ((i=0; i<100; i++)); do echo y >> "$licenses"; done
"$manager" --sdk_root="$sdk" --licenses < "$licenses"
"$manager" --sdk_root="$sdk" --install 'platforms;android-34' 'build-tools;34.0.0' 'cmake;3.22.1' 'ndk;29.0.14206865'
test -x "$sdk/ndk/29.0.14206865/toolchains/llvm/prebuilt/linux-x86_64/bin/aarch64-linux-android33-clang"
# sdkmanager's CMake package does not put cmake on PATH; both the native
# archive build and AGP need the pinned SDK binary, not an apt fallback.
test -x "$sdk/cmake/3.22.1/bin/cmake" || { echo "SDK CMake 3.22.1 missing" >&2; exit 1; }
git submodule update --init c/vendor/github.com/jedisct1/libsodium c/vendor/github.com/quictls/openssl c/vendor/github.com/ngtcp2/ngtcp2

#!/usr/bin/env bash
# Copyright 2026 Marcelo Cantos
# SPDX-License-Identifier: Apache-2.0
set -euo pipefail
cd "$(dirname "$0")/.."
# git's safe.directory is set in jitpack.yml before any submodule operation.
# Autotools is needed to cross-compile pinned libsodium; quictls needs Perl.
#
# JitPack's build container has no root and no noninteractive sudo, so
# apt-get install is unavailable. When autoconf/automake/libtoolize are
# missing and the apt path is closed, bootstrap them from pinned,
# checksum-verified GNU source tarballs into a user-writable prefix
# instead of failing closed outright. perl and make are not bootstrapped
# here — they are expected to already be on a JitPack-class build image,
# and building them from source is a much larger undertaking than the
# small, fast-building autotools chain below.
BUILD_TOOLS_PREFIX="${PIGEON_BUILD_TOOLS_PREFIX:-$HOME/.cache/pigeon-jitpack-build-tools}"
export PATH="$BUILD_TOOLS_PREFIX/bin:$PATH"

missing=()
for cmd in autoconf automake libtoolize perl make; do
    command -v "$cmd" >/dev/null || missing+=("$cmd")
done

# GNU release tarballs pinned by version + sha256; mirrors serve the exact
# same immutable bytes as ftp.gnu.org, so a fixed checksum is safe to pin.
# Order matters: autoconf/automake/libtool's own ./configure needs m4.
declare -A GNU_PKG_DIR=(
    [m4]=m4-1.4.19
    [autoconf]=autoconf-2.71
    [automake]=automake-1.16.5
    [libtool]=libtool-2.4.7
)
declare -A GNU_PKG_SHA256=(
    [m4]=3be4a26d825ffdfda52a56fc43246456989a3630093cced3fbddf4771ee58a70
    [autoconf]=431075ad0bf529ef13cb41e9042c542381103e80015686222b8a9d4abef42a1c
    [automake]=07bd24ad08a64bc17250ce09ec56e921d6343903943e99ccf63bbf0705e34605
    [libtool]=04e96c2404ea70c590c546eba4202a4e12722c640016c12b9b2f1ce3d481e9a8
)
GNU_MIRRORS=(
    "https://ftpmirror.gnu.org"
    "https://ftp.gnu.org/gnu"
)

sha256_of() {
    if command -v sha256sum >/dev/null; then sha256sum "$1" | awk '{print $1}'
    else shasum -a 256 "$1" | awk '{print $1}'
    fi
}

bootstrap_gnu_package() {
    local pkg="$1" dirname="${GNU_PKG_DIR[$1]}" want_sha="${GNU_PKG_SHA256[$1]}"
    local cache="$BUILD_TOOLS_PREFIX/src"
    local tarball="$cache/$dirname.tar.gz"
    mkdir -p "$cache"
    if [ ! -f "$tarball" ] || [ "$(sha256_of "$tarball" 2>/dev/null)" != "$want_sha" ]; then
        local fetched=
        for mirror in "${GNU_MIRRORS[@]}"; do
            if curl --fail --location --silent --show-error --proto '=https' \
                    --max-time 180 -o "$tarball.tmp" "$mirror/$pkg/$dirname.tar.gz"; then
                fetched=1
                break
            fi
        done
        [ -n "$fetched" ] || { echo "error: could not download $dirname.tar.gz from any GNU mirror" >&2; return 1; }
        mv "$tarball.tmp" "$tarball"
        local got_sha
        got_sha="$(sha256_of "$tarball")"
        if [ "$got_sha" != "$want_sha" ]; then
            echo "error: $dirname.tar.gz sha256 mismatch: want $want_sha got $got_sha" >&2
            rm -f "$tarball"
            return 1
        fi
    fi
    local build="$BUILD_TOOLS_PREFIX/build-$dirname"
    rm -rf "$build"
    mkdir -p "$build"
    tar -xzf "$tarball" -C "$build" --strip-components=1
    local nproc
    nproc="$(getconf _NPROCESSORS_ONLN 2>/dev/null || echo 4)"
    (cd "$build" && ./configure --prefix="$BUILD_TOOLS_PREFIX" >configure.log 2>&1 \
        && make -j"$nproc" >make.log 2>&1 \
        && make install >install.log 2>&1) || {
        echo "error: building $dirname from source failed; see $build/{configure,make,install}.log" >&2
        return 1
    }
    rm -rf "$build"
}

bootstrap_missing_autotools() {
    local still_missing=()
    for cmd in autoconf automake libtoolize; do
        command -v "$cmd" >/dev/null || still_missing+=("$cmd")
    done
    ((${#still_missing[@]})) || return 0
    command -v curl >/dev/null || { echo "error: curl unavailable, cannot bootstrap autotools from source" >&2; return 1; }
    command -v tar >/dev/null || { echo "error: tar unavailable, cannot bootstrap autotools from source" >&2; return 1; }
    mkdir -p "$BUILD_TOOLS_PREFIX/bin"
    command -v m4 >/dev/null || bootstrap_gnu_package m4
    command -v autoconf >/dev/null || bootstrap_gnu_package autoconf
    command -v automake >/dev/null || bootstrap_gnu_package automake
    command -v libtoolize >/dev/null || bootstrap_gnu_package libtool
}

if ((${#missing[@]})); then
    if command -v apt-get >/dev/null; then
        if [ "$(id -u)" -eq 0 ]; then
            apt=(apt-get)
        elif command -v sudo >/dev/null && sudo -n true 2>/dev/null; then
            apt=(sudo -n apt-get)
        else
            apt=()
        fi
        if ((${#apt[@]})); then
            "${apt[@]}" update
            "${apt[@]}" install -y --no-install-recommends autoconf automake libtool perl make
        fi
    fi
    # Re-check after any apt attempt, then fall back to the source bootstrap
    # for the three tools that chain quickly (autoconf, automake, libtool).
    # perl and make are not bootstrapped from source — if those are still
    # missing, fail closed rather than attempt a much larger build.
    still_missing=()
    for cmd in autoconf automake libtoolize perl make; do
        command -v "$cmd" >/dev/null || still_missing+=("$cmd")
    done
    if ((${#still_missing[@]})); then
        for cmd in perl make; do
            for want in "${still_missing[@]}"; do
                if [ "$want" = "$cmd" ]; then
                    echo "Missing build tools: ${still_missing[*]} (perl/make cannot be bootstrapped without root; no apt-get or no root/noninteractive sudo)" >&2
                    exit 1
                fi
            done
        done
        bootstrap_missing_autotools || { echo "Missing build tools: ${still_missing[*]} (no root or noninteractive sudo, and source bootstrap failed)" >&2; exit 1; }
        still_missing=()
        for cmd in autoconf automake libtoolize perl make; do
            command -v "$cmd" >/dev/null || still_missing+=("$cmd")
        done
        if ((${#still_missing[@]})); then
            echo "Missing build tools: ${still_missing[*]} (source bootstrap into $BUILD_TOOLS_PREFIX did not resolve them)" >&2
            exit 1
        fi
    fi
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

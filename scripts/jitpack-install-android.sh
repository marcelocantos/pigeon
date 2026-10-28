#!/usr/bin/env bash
# Copyright 2026 Marcelo Cantos
# SPDX-License-Identifier: Apache-2.0
set -euo pipefail
cd "$(dirname "$0")/.."
: "${VERSION:?JitPack must supply the immutable requested tag or commit as VERSION}"
# JitPack may continue to install even if before_install failed; check here
# in the same fail-closed shell as Gradle rather than relying on that stage.
bash scripts/jitpack-prepare-android.sh
export ANDROID_HOME="${ANDROID_HOME:-${ANDROID_SDK_ROOT:-$HOME/android-sdk}}"
export ANDROID_NDK_HOME="$ANDROID_HOME/ndk/29.0.14206865"
# jitpack-prepare-android.sh may have bootstrapped autoconf/automake/
# libtoolize from source into this prefix when the container had no root
# and no apt-get path (see that script); it ran as a subprocess, so its
# PATH does not reach here on its own. c/vendor/build-android.sh, invoked
# from the Gradle native build task below, needs those tools too.
BUILD_TOOLS_PREFIX="${PIGEON_BUILD_TOOLS_PREFIX:-$HOME/.cache/pigeon-jitpack-build-tools}"
export PATH="$ANDROID_HOME/cmake/3.22.1/bin:$BUILD_TOOLS_PREFIX/bin:$PATH"
# JitPack harvests the installed module POM and AAR from ~/.m2/repository.
# Do not run the aggregate build: pigeon-host and desktop JVM tests are not
# publication inputs, and the Android module needs its own nested wrapper.
android/gradlew -p android :pigeon:publishReleasePublicationToMavenLocal \
    -P"pigeon.jitpackVersion=$VERSION" --no-daemon --console=plain

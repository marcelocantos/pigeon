# Copyright 2026 Marcelo Cantos
# SPDX-License-Identifier: Apache-2.0
#
# CMake configuration invoked by AGP's externalNativeBuild for the
# Android JNI library libpigeon-jni.so (one per ABI). Driven by
# android/pigeon/build.gradle.kts.
#
# The C sources live at the repo root under c/ — this file lives at
# android/pigeon/src/main/cpp/ as that is the conventional layout AGP
# resolves; it reaches back up the tree via PIGEON_REPO_ROOT.
#
# Expects the following passed via -D from the Gradle build:
#   PIGEON_REPO_ROOT     absolute path to the pigeon repo root.
#   PIGEON_SODIUM_ROOT   absolute path to c/vendor/build-android/<abi>/
#                        (contains lib/libsodium.a + include/sodium.h).
# Falls back to sensible defaults if not provided so the file can also
# be driven manually for debugging.

cmake_minimum_required(VERSION 3.22.1)
project(pigeon-jni C)

set(CMAKE_C_STANDARD 11)
set(CMAKE_POSITION_INDEPENDENT_CODE ON)

if(NOT DEFINED PIGEON_REPO_ROOT)
    # android/pigeon/src/main/cpp -> ../../../../..
    get_filename_component(PIGEON_REPO_ROOT
        "${CMAKE_CURRENT_SOURCE_DIR}/../../../../.." ABSOLUTE)
endif()

if(NOT DEFINED PIGEON_SODIUM_ROOT)
    set(PIGEON_SODIUM_ROOT
        "${PIGEON_REPO_ROOT}/c/vendor/build-android/${ANDROID_ABI}")
endif()

set(PIGEON_DIST_DIR "${PIGEON_REPO_ROOT}/dist")
set(PIGEON_JNI_SRC  "${PIGEON_REPO_ROOT}/c/jni/pigeon_jni.c")
set(SODIUM_LIB      "${PIGEON_SODIUM_ROOT}/lib/libsodium.a")
set(SODIUM_INCLUDE  "${PIGEON_SODIUM_ROOT}/include")

if(NOT EXISTS "${PIGEON_DIST_DIR}/pigeon.c")
    message(FATAL_ERROR
        "Amalgamated pigeon C source missing: ${PIGEON_DIST_DIR}/pigeon.c\n"
        "Run: make amalgamate (or trigger via Gradle :pigeon:amalgamateNative).")
endif()

if(NOT EXISTS "${SODIUM_LIB}")
    message(FATAL_ERROR
        "Vendored libsodium static lib missing for ${ANDROID_ABI}: ${SODIUM_LIB}\n"
        "Run: bash c/vendor/build-android.sh ${ANDROID_ABI}")
endif()

# libsodium static lib as an imported target. CMake's IMPORTED targets
# survive transitive linking cleanly across ABIs and don't pollute the
# default search path.
add_library(sodium_static STATIC IMPORTED)
set_target_properties(sodium_static PROPERTIES
    IMPORTED_LOCATION "${SODIUM_LIB}"
    INTERFACE_INCLUDE_DIRECTORIES "${SODIUM_INCLUDE}")

add_library(pigeon-jni SHARED
    "${PIGEON_DIST_DIR}/pigeon.c"
    "${PIGEON_JNI_SRC}"
)

target_compile_definitions(pigeon-jni PRIVATE
    PIGEON_CRYPTO_LIBSODIUM
)

target_include_directories(pigeon-jni PRIVATE
    "${PIGEON_DIST_DIR}"
)

# Android's NDK ships jni.h in the sysroot — no extra includes needed.
# Link sodium statically (the .so is self-contained on the device) and
# the Android log library for any future LOGI/LOGE calls from the JNI
# shim.
find_library(ANDROID_LOG_LIB log)
target_link_libraries(pigeon-jni
    PRIVATE
        sodium_static
        ${ANDROID_LOG_LIB}
)

# Strip symbols in release builds to shrink the .so. AGP sets
# CMAKE_BUILD_TYPE to Release for assembleRelease.
if(CMAKE_BUILD_TYPE STREQUAL "Release")
    target_link_options(pigeon-jni PRIVATE "-Wl,--gc-sections" "-Wl,--strip-all")
endif()

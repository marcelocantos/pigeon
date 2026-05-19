# Copyright 2026 Marcelo Cantos
# SPDX-License-Identifier: Apache-2.0

cmake_minimum_required(VERSION 3.20)
project(pigeon C)

set(CMAKE_C_STANDARD 11)

option(PIGEON_MAX_MSG "Maximum message size" 1048576)
option(PIGEON_CRYPTO_BACKEND "Crypto backend: libsodium or openssl" "libsodium")

add_library(pigeon STATIC
    src/pigeon.c
    src/crypto.c
    src/pairingceremony_gen.c
    src/session_gen.c
    src/activation.c
    src/listener.c
)

target_include_directories(pigeon PUBLIC include)
target_include_directories(pigeon PRIVATE include/pigeon)

target_compile_definitions(pigeon PRIVATE
    PIGEON_MAX_MSG=${PIGEON_MAX_MSG}
)

if(PIGEON_CRYPTO_BACKEND STREQUAL "libsodium")
    # Vendored static lib: c/vendor/build/lib/libsodium.a, built by
    # `bash c/vendor/build.sh libsodium` (or `make build-vendor-deps`).
    set(SODIUM_VENDOR_ROOT "${CMAKE_CURRENT_SOURCE_DIR}/vendor/build")
    set(SODIUM_VENDOR_LIB "${SODIUM_VENDOR_ROOT}/lib/libsodium.a")
    set(SODIUM_VENDOR_INC "${SODIUM_VENDOR_ROOT}/include")
    if(NOT EXISTS "${SODIUM_VENDOR_LIB}")
        message(FATAL_ERROR
            "vendored libsodium not built: ${SODIUM_VENDOR_LIB}\n"
            "run: bash c/vendor/build.sh libsodium")
    endif()
    target_compile_definitions(pigeon PRIVATE PIGEON_CRYPTO_LIBSODIUM)
    target_include_directories(pigeon PRIVATE "${SODIUM_VENDOR_INC}")
    target_link_libraries(pigeon PRIVATE "${SODIUM_VENDOR_LIB}")
elseif(PIGEON_CRYPTO_BACKEND STREQUAL "openssl")
    find_package(OpenSSL REQUIRED)
    target_compile_definitions(pigeon PRIVATE PIGEON_CRYPTO_OPENSSL)
    target_link_libraries(pigeon PRIVATE OpenSSL::Crypto)
endif()

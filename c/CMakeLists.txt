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
)

target_include_directories(pigeon PUBLIC include)
target_include_directories(pigeon PRIVATE include/pigeon)

target_compile_definitions(pigeon PRIVATE
    PIGEON_MAX_MSG=${PIGEON_MAX_MSG}
)

if(PIGEON_CRYPTO_BACKEND STREQUAL "libsodium")
    find_package(PkgConfig REQUIRED)
    pkg_check_modules(SODIUM REQUIRED libsodium)
    target_compile_definitions(pigeon PRIVATE PIGEON_CRYPTO_LIBSODIUM)
    target_include_directories(pigeon PRIVATE ${SODIUM_INCLUDE_DIRS})
    target_link_libraries(pigeon PRIVATE ${SODIUM_LIBRARIES})
elseif(PIGEON_CRYPTO_BACKEND STREQUAL "openssl")
    find_package(OpenSSL REQUIRED)
    target_compile_definitions(pigeon PRIVATE PIGEON_CRYPTO_OPENSSL)
    target_link_libraries(pigeon PRIVATE OpenSSL::Crypto)
endif()

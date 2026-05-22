// swift-tools-version: 5.9

import PackageDescription

// Header / linker settings for the vendored C stack. CPigeon's amalgamated
// source (dist/pigeon.c) is built with -DPIGEON_CRYPTO_LIBSODIUM and pulls
// AEAD primitives from libsodium; the ngtcp2 transport build unit
// (Sources/CPigeon/cpigeon_ngtcp2.c, added in 🎯T37) compiles
// c/src/ngtcp2_transport.c + c/src/listener_ngtcp2.c against the vendored
// ngtcp2 + quictls headers. All four static libraries (libsodium,
// libngtcp2, libngtcp2_crypto_quictls, libssl, libcrypto) live under
// c/vendor/build/lib/.
//
// The vendored static libs are in-tree at a stable path, so we hard-code
// them rather than asking SwiftPM to discover them dynamically. Callers
// must run `make build-vendor-deps` (or at minimum
// `bash c/vendor/build.sh libsodium openssl ngtcp2`) before `swift build`.
let cpigeonCSettings: [CSetting] = [
    .define("PIGEON_CRYPTO_LIBSODIUM"),
    // dist/pigeon.h is included by cpigeon.c via "../../dist/pigeon.c".
    // The amalgamated header itself self-includes; SwiftPM also needs
    // to find it for the umbrella header, hence the extra search path.
    .headerSearchPath("../../dist"),
    // Vendored libsodium / ngtcp2 / openssl headers under
    // c/vendor/build/include. .headerSearchPath paths are relative to
    // the target source dir (Sources/CPigeon/), so up two levels reaches
    // the package root.
    .headerSearchPath("../../c/vendor/build/include"),
    // Canonical pigeon C headers (c/include) — required by the ngtcp2
    // transport build unit, which #includes "pigeon/ngtcp2_transport.h"
    // and (transitively) the canonical "pigeon/pigeon.h". The amalgamated
    // dist/pigeon.h is guard-compatible (PIGEON_H), so the
    // include-once protection keeps the two coexistent.
    .headerSearchPath("../../c/include"),
]

let cpigeonLinkerSettings: [LinkerSetting] = [
    // Linker unsafeFlags resolve relative to the package root. Order
    // matters: ngtcp2_crypto_quictls depends on ngtcp2 + openssl,
    // ngtcp2 depends on openssl. Libsodium last (no deps).
    .unsafeFlags([
        "c/vendor/build/lib/libngtcp2_crypto_quictls.a",
        "c/vendor/build/lib/libngtcp2.a",
        "c/vendor/build/lib/libssl.a",
        "c/vendor/build/lib/libcrypto.a",
        "c/vendor/build/lib/libsodium.a",
    ]),
]

let package = Package(
    name: "Pigeon",
    platforms: [.iOS(.v16), .macOS(.v13)],
    products: [
        .library(name: "Pigeon", targets: ["Pigeon"]),
    ],
    targets: [
        // C target: amalgamated libpigeon (dist/pigeon.c) + an in-process
        // loopback transport so the Swift tests can exercise the
        // multi-channel session API without ngtcp2.
        .target(
            name: "CPigeon",
            path: "Sources/CPigeon",
            publicHeadersPath: "include",
            cSettings: cpigeonCSettings,
            linkerSettings: cpigeonLinkerSettings
        ),
        .target(
            name: "Pigeon",
            dependencies: ["CPigeon"]
        ),
        .testTarget(name: "PigeonTests", dependencies: ["Pigeon", "CPigeon"]),
        .testTarget(name: "PigeonRelayE2ETests", dependencies: ["Pigeon"]),
        .executableTarget(
            name: "pigeon-e2e-swift",
            dependencies: ["Pigeon"],
            path: "e2e/swift"
        ),
        // pairing-peer-swift drives the Swift PairingCeremony as the
        // initiator side from a Go acceptor's subprocess, used by the
        // cross-language test in pairing/cross_swift_test.go.
        .executableTarget(
            name: "pairing-peer-swift",
            dependencies: ["Pigeon"],
            path: "e2e/pairing-peer-swift"
        ),
    ]
)

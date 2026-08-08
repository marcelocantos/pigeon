// swift-tools-version: 5.9

import Foundation
import PackageDescription

// Product: library "Pigeon" (import Pigeon). Consumable as:
//   - SPM path dependency (local sibling / XcodeGen path:)
//   - SPM git dependency (url: github.com/marcelocantos/pigeon, from: tag)
//   - Go module github.com/marcelocantos/pigeon (separate from this manifest)
// See README.md § "Consuming pigeon from another app".
//
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
//
// Linker paths use absolute paths derived from this Package.swift location.
// Relative "c/vendor/..." flags resolve against the *consumer* CWD when an
// app depends on this package (Xcode link of the app dylib) — that broke
// orthograph/jevons-style path consumers. Absolute paths keep the libs
// found regardless of who links.
//
// unsafeFlags are still required for arbitrary .a paths. App targets
// (Xcode) may depend on this package; pure SPM library dependents may not.
//
// Products:
//   * PigeonCore — pure Swift (pairing artifacts, machines). Links on iOS
//     device without host-only static libs.
//   * Pigeon — full stack (PigeonCore + CPigeon + ngtcp2). Host macOS
//     vendor .a today; iOS once multi-platform vendor build exists.
let packageRoot = URL(fileURLWithPath: #filePath)
    .deletingLastPathComponent()
    .path
let vendorLibDir = packageRoot + "/c/vendor/build/lib"

let cpigeonCSettings: [CSetting] = [
    .define("PIGEON_CRYPTO_LIBSODIUM"),
    .headerSearchPath("../../dist"),
    .headerSearchPath("../../c/vendor/build/include"),
    .headerSearchPath("../../c/include"),
]

let cpigeonLinkerSettings: [LinkerSetting] = [
    .unsafeFlags([
        vendorLibDir + "/libngtcp2_crypto_quictls.a",
        vendorLibDir + "/libngtcp2.a",
        vendorLibDir + "/libssl.a",
        vendorLibDir + "/libcrypto.a",
        vendorLibDir + "/libsodium.a",
    ]),
]

let package = Package(
    name: "Pigeon",
    platforms: [.iOS(.v16), .macOS(.v13)],
    products: [
        .library(name: "Pigeon", targets: ["Pigeon"]),
        .library(name: "PigeonCore", targets: ["PigeonCore"]),
    ],
    targets: [
        .target(
            name: "CPigeon",
            path: "Sources/CPigeon",
            publicHeadersPath: "include",
            cSettings: cpigeonCSettings,
            linkerSettings: cpigeonLinkerSettings
        ),
        // Pure Swift: pairing, machines, wire helpers — no CPigeon.
        .target(
            name: "PigeonCore",
            path: "Sources/PigeonCore"
        ),
        // Full stack: transport + session over CPigeon; re-exports Core.
        .target(
            name: "Pigeon",
            dependencies: ["CPigeon", "PigeonCore"],
            path: "Sources/Pigeon"
        ),
        .testTarget(name: "PigeonTests", dependencies: ["Pigeon", "PigeonCore", "CPigeon"]),
        .testTarget(name: "PigeonRelayE2ETests", dependencies: ["Pigeon"]),
        .executableTarget(
            name: "pigeon-e2e-swift",
            dependencies: ["Pigeon"],
            path: "e2e/swift"
        ),
        .executableTarget(
            name: "pairing-peer-swift",
            dependencies: ["Pigeon"],
            path: "e2e/pairing-peer-swift"
        ),
    ]
)

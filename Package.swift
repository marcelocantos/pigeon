// swift-tools-version: 5.9

import PackageDescription

// Header / linker settings for libsodium. CPigeon's amalgamated source
// (dist/pigeon.c) is built with -DPIGEON_CRYPTO_LIBSODIUM and pulls
// the AEAD primitives from libsodium. We link the static lib built by
// `c/vendor/build.sh libsodium` (or `make build-vendor-deps`); see
// c/vendor/github.com/jedisct1/libsodium for the pinned submodule.
//
// The vendored static lib is in-tree at a stable path, so we hard-code
// it rather than asking SwiftPM to discover libsodium dynamically.
// Callers must run `make build-vendor-deps` (or at minimum
// `bash c/vendor/build.sh libsodium`) before `swift build`.
let cpigeonCSettings: [CSetting] = [
    .define("PIGEON_CRYPTO_LIBSODIUM"),
    // dist/pigeon.h is included by cpigeon.c via "../../dist/pigeon.c".
    // The amalgamated header itself self-includes; SwiftPM also needs
    // to find it for the umbrella header, hence the extra search path.
    .headerSearchPath("../../dist"),
    // Vendored libsodium headers under c/vendor/build/include.
    // .headerSearchPath paths are relative to the target source dir
    // (Sources/CPigeon/), so up two levels reaches the package root.
    .headerSearchPath("../../c/vendor/build/include"),
]

let cpigeonLinkerSettings: [LinkerSetting] = [
    // Linker unsafeFlags resolve relative to the package root.
    .unsafeFlags(["c/vendor/build/lib/libsodium.a"]),
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
    ]
)

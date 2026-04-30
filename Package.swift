// swift-tools-version: 5.9

import PackageDescription

// Header / linker settings for libsodium. CPigeon's amalgamated source
// (dist/pigeon.c) is built with -DPIGEON_CRYPTO_LIBSODIUM and pulls
// the AEAD primitives from a system libsodium. On the dev box that's
// the Homebrew formula at /opt/homebrew (Apple Silicon). x86_64 macOS
// boxes use /usr/local; Linux uses the system search paths.
//
// SwiftPM doesn't have a portable "ask pkg-config for libsodium"
// knob, so we hard-code the Homebrew layout that the rest of this
// repo (cwire/cwire.go, c/CMakeLists.txt) already assumes.
let cpigeonCSettings: [CSetting] = [
    .define("PIGEON_CRYPTO_LIBSODIUM"),
    // dist/pigeon.h is included by cpigeon.c via "../../dist/pigeon.c".
    // The amalgamated header itself self-includes; SwiftPM also needs
    // to find it for the umbrella header, hence the extra search path.
    .headerSearchPath("../../dist"),
    // Header path for sodium.h on Apple Silicon Homebrew. SwiftPM's
    // .headerSearchPath insists on relative paths, so route through
    // -I via .unsafeFlags.
    .unsafeFlags(["-I/opt/homebrew/include"], .when(platforms: [.macOS])),
]

let cpigeonLinkerSettings: [LinkerSetting] = [
    .unsafeFlags(["-L/opt/homebrew/lib"], .when(platforms: [.macOS])),
    .linkedLibrary("sodium"),
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

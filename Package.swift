// swift-tools-version: 5.9

import PackageDescription

let package = Package(
    name: "Tern",
    platforms: [.iOS(.v16), .macOS(.v13)],
    products: [
        .library(name: "TernCrypto", targets: ["TernCrypto"]),
        .library(name: "TernRelay", targets: ["TernRelay"]),
    ],
    targets: [
        .target(name: "TernCrypto"),
        .target(name: "TernRelay"),
        .testTarget(name: "TernCryptoTests", dependencies: ["TernCrypto"]),
        .testTarget(name: "TernRelayTests", dependencies: ["TernRelay", "TernCrypto"]),
    ]
)

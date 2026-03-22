// swift-tools-version: 5.9

import PackageDescription

let package = Package(
    name: "Tern",
    platforms: [.iOS(.v16), .macOS(.v13)],
    products: [
        .library(name: "TernCrypto", targets: ["TernCrypto"]),
    ],
    targets: [
        .target(name: "TernCrypto"),
        .testTarget(name: "TernCryptoTests", dependencies: ["TernCrypto"]),
    ]
)

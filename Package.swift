// swift-tools-version: 5.9

import PackageDescription

let package = Package(
    name: "Pigeon",
    platforms: [.iOS(.v16), .macOS(.v13)],
    products: [
        .library(name: "Pigeon", targets: ["Pigeon"]),
    ],
    targets: [
        .target(name: "Pigeon"),
        .testTarget(name: "PigeonTests", dependencies: ["Pigeon"]),
        .testTarget(name: "PigeonRelayE2ETests", dependencies: ["Pigeon"]),
        .executableTarget(
            name: "pigeon-e2e-swift",
            dependencies: ["Pigeon"],
            path: "e2e/swift"
        ),
    ]
)

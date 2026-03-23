// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

#if canImport(Network)

import XCTest
@testable import TernRelay

final class TernRelayTests: XCTestCase {

    // MARK: - Length-prefix encoding

    func testEncodeLengthPrefixZero() {
        let data = encodeLengthPrefix(0)
        XCTAssertEqual(data, Data([0, 0, 0, 0]))
    }

    func testEncodeLengthPrefixSmall() {
        let data = encodeLengthPrefix(42)
        XCTAssertEqual(data, Data([0, 0, 0, 42]))
    }

    func testEncodeLengthPrefix256() {
        let data = encodeLengthPrefix(256)
        XCTAssertEqual(data, Data([0, 0, 1, 0]))
    }

    func testEncodeLengthPrefixLarge() {
        // 0x01020304 = 16909060
        let data = encodeLengthPrefix(0x01020304)
        XCTAssertEqual(data, Data([0x01, 0x02, 0x03, 0x04]))
    }

    func testEncodeLengthPrefixMax() {
        let data = encodeLengthPrefix(UInt32.max)
        XCTAssertEqual(data, Data([0xFF, 0xFF, 0xFF, 0xFF]))
    }

    // MARK: - Length-prefix decoding

    func testDecodeLengthPrefixZero() {
        let value = decodeLengthPrefix(Data([0, 0, 0, 0]))
        XCTAssertEqual(value, 0)
    }

    func testDecodeLengthPrefixSmall() {
        let value = decodeLengthPrefix(Data([0, 0, 0, 42]))
        XCTAssertEqual(value, 42)
    }

    func testDecodeLengthPrefix256() {
        let value = decodeLengthPrefix(Data([0, 0, 1, 0]))
        XCTAssertEqual(value, 256)
    }

    func testDecodeLengthPrefixLarge() {
        let value = decodeLengthPrefix(Data([0x01, 0x02, 0x03, 0x04]))
        XCTAssertEqual(value, 0x01020304)
    }

    // MARK: - Round-trip

    func testLengthPrefixRoundTrip() {
        for value: UInt32 in [0, 1, 127, 128, 255, 256, 65535, 65536, 1_000_000, UInt32.max] {
            let encoded = encodeLengthPrefix(value)
            let decoded = decodeLengthPrefix(encoded)
            XCTAssertEqual(decoded, value, "Round-trip failed for \(value)")
        }
    }

    // MARK: - Handshake message construction

    func testHandshakeRegisterNoToken() {
        let msg = buildHandshakeMessage(role: "register")
        XCTAssertEqual(msg, "register")
    }

    func testHandshakeRegisterWithToken() {
        let msg = buildHandshakeMessage(role: "register", token: "secret123")
        XCTAssertEqual(msg, "register:secret123")
    }

    func testHandshakeConnect() {
        let msg = buildHandshakeMessage(role: "connect", instanceID: "abc123")
        XCTAssertEqual(msg, "connect:abc123")
    }

    func testHandshakeConnectEmptyID() {
        let msg = buildHandshakeMessage(role: "connect", instanceID: "")
        XCTAssertEqual(msg, "connect:")
    }

    func testHandshakeConnectNilID() {
        let msg = buildHandshakeMessage(role: "connect")
        XCTAssertEqual(msg, "connect:")
    }
}

#endif

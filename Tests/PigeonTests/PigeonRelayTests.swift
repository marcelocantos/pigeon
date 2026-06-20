// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

#if canImport(Network)

import XCTest
@testable import Pigeon

final class PigeonRelayTests: XCTestCase {

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

    // MARK: - Relay greeting construction (T45)
    //
    // Greetings now go through the protogen-generated PigeonWire
    // encoders; PigeonConn.register / .listen / .connect call these
    // directly. Vectors match wire_vectors_test.go.

    private func greeting(_ data: Data) -> String {
        String(decoding: data, as: UTF8.self)
    }

    func testGreetingRegisterBare() {
        XCTAssertEqual(
            greeting(PigeonWire.encodeRelayGreetingRegister(token: "", instanceId: "")),
            "register")
    }

    func testGreetingRegisterWithToken() {
        XCTAssertEqual(
            greeting(PigeonWire.encodeRelayGreetingRegister(token: "secret123", instanceId: "")),
            "register:secret123:")
    }

    func testGreetingListenIDOnly() {
        XCTAssertEqual(
            greeting(PigeonWire.encodeRelayGreetingListen(token: "", instanceId: "id-7")),
            "listen::id-7")
    }

    func testGreetingListenTokenAndID() {
        XCTAssertEqual(
            greeting(PigeonWire.encodeRelayGreetingListen(token: "tok", instanceId: "id-2")),
            "listen:tok:id-2")
    }

    func testGreetingConnect() {
        XCTAssertEqual(
            greeting(PigeonWire.encodeRelayGreetingConnect(instanceId: "abc123")),
            "connect:abc123")
    }

    func testGreetingConnectEmptyID() {
        XCTAssertEqual(
            greeting(PigeonWire.encodeRelayGreetingConnect(instanceId: "")),
            "connect:")
    }
}

#endif

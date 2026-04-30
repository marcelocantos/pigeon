// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0
//
// Tests for the SwiftPM C-target wrapper from T29 — Pigeon's
// Session/Stream/Datagram API on top of CPigeon. We exercise both
// halves through the in-process loopback transport so no real
// network is required; this is the Swift counterpart of
// c/test/test_pigeon.c::test_session_stream_roundtrip and
// ::test_session_datagram_roundtrip.

import XCTest
@testable import Pigeon
import CPigeon

final class SessionTests: XCTestCase {

    // MARK: - Wire vectors (cross-language)

    func testEmptyClientStreamHeader() throws {
        // Empty-name primary client stream encodes to a single 0 byte.
        let bytes = try PigeonWire.encodeStreamHeader(
            isBackend: false, clientTag: 0, name: ""
        )
        XCTAssertEqual([UInt8](bytes), [0x00])
        let decoded = try PigeonWire.decodeClientStreamHeader(bytes)
        XCTAssertEqual(decoded.name, "")
        XCTAssertEqual(decoded.consumed, 1)
    }

    func testNamedClientStreamHeader() throws {
        // "control" -> [0x07, 'c','o','n','t','r','o','l']
        let bytes = try PigeonWire.encodeStreamHeader(
            isBackend: false, clientTag: 0, name: "control"
        )
        XCTAssertEqual([UInt8](bytes),
                       [0x07, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c])
        let decoded = try PigeonWire.decodeClientStreamHeader(bytes)
        XCTAssertEqual(decoded.name, "control")
        XCTAssertEqual(decoded.consumed, 8)
    }

    func testBackendStreamHeader() throws {
        // tag=0x01020304, name="chat" -> 01 02 03 04 04 'c' 'h' 'a' 't'
        let bytes = try PigeonWire.encodeStreamHeader(
            isBackend: true, clientTag: 0x01020304, name: "chat"
        )
        XCTAssertEqual([UInt8](bytes),
                       [0x01, 0x02, 0x03, 0x04, 0x04, 0x63, 0x68, 0x61, 0x74])
        let decoded = try PigeonWire.decodeBackendStreamHeader(bytes)
        XCTAssertEqual(decoded.clientTag, 0x01020304)
        XCTAssertEqual(decoded.name, "chat")
        XCTAssertEqual(decoded.consumed, 9)
    }

    // MARK: - Uvarint round-trips

    func testUvarintRoundTrip() throws {
        for v: UInt64 in [0, 1, 127, 128, 255, 256, 16383, 16384, 1 << 32, UInt64.max] {
            let bytes = PigeonWire.encodeUvarint(v)
            let (decoded, consumed) = try PigeonWire.decodeUvarint(bytes)
            XCTAssertEqual(decoded, v, "round-trip mismatch for \(v)")
            XCTAssertEqual(consumed, bytes.count)
        }
    }

    func testUvarintTruncated() {
        XCTAssertThrowsError(try PigeonWire.decodeUvarint(Data())) { err in
            XCTAssertEqual(err as? PigeonWireError, .truncated)
        }
    }

    // MARK: - Session: stream round-trip via loopback

    func testStreamRoundTripViaLoopback() async throws {
        // Two endpoints share the same symmetric AEAD key; we treat
        // them as is_backend=false on both sides for direct loopback
        // (the C reference test does the same, since the relay-tag
        // prefix is normally stripped by the relay before forwarding).
        let key = Data((0..<32).map { _ in UInt8.random(in: 0...255) })

        let ta = LoopbackTransport()
        let tb = LoopbackTransport()
        LoopbackTransport.pair(ta, tb)

        let sa = try PigeonSession(
            masterKey: key, isBackend: false, clientTag: 0,
            datagramChannels: [], transport: ta
        )
        let sb = try PigeonSession(
            masterKey: key, isBackend: false, clientTag: 0,
            datagramChannels: [], transport: tb
        )

        // A opens a "chat" stream; the loopback queues an inbound
        // stream on B with the unencrypted name-binding header as
        // its first message. We accept it and adopt it on B.
        let aChat = try await sa.openStream(name: "chat")
        let accepted = try tb.acceptWithHeader()
        let decoded = try PigeonWire.decodeClientStreamHeader(accepted.header)
        XCTAssertEqual(decoded.name, "chat")
        let bChat = sb.adoptAcceptedStream(name: "chat", handle: accepted.handle)

        // A → B
        try await aChat.send(Data("hello".utf8))
        let m1 = try await bChat.recv()
        XCTAssertEqual(String(decoding: m1, as: UTF8.self), "hello")

        // B → A
        try await bChat.send(Data("world".utf8))
        let m2 = try await aChat.recv()
        XCTAssertEqual(String(decoding: m2, as: UTF8.self), "world")

        try aChat.close()
    }

    // MARK: - Session: datagram round-trip via loopback

    func testDatagramRoundTripViaLoopback() async throws {
        let key = Data((0..<32).map { _ in UInt8.random(in: 0...255) })

        let ta = LoopbackTransport()
        let tb = LoopbackTransport()
        LoopbackTransport.pair(ta, tb)

        // Pre-declared datagram channels (must be identical on both
        // peers).
        let chans = [
            DatagramChannelDef(name: "ping", channelID: 1),
            DatagramChannelDef(name: "metric", channelID: 2),
        ]

        let sa = try PigeonSession(
            masterKey: key, isBackend: false, clientTag: 0,
            datagramChannels: chans, transport: ta
        )
        let sb = try PigeonSession(
            masterKey: key, isBackend: false, clientTag: 0,
            datagramChannels: chans, transport: tb
        )

        let aPing = try sa.datagram(named: "ping")
        let bPing = try sb.datagram(named: "ping")

        try await aPing.send(Data("p1".utf8))
        let got = try await bPing.recv()
        XCTAssertEqual(got.map { String(decoding: $0, as: UTF8.self) }, "p1")
    }

    func testUnknownDatagramChannelThrows() throws {
        let key = Data(repeating: 0xAB, count: 32)
        let t = LoopbackTransport()
        let s = try PigeonSession(
            masterKey: key, isBackend: false, clientTag: 0,
            datagramChannels: [
                DatagramChannelDef(name: "ping", channelID: 1),
            ],
            transport: t
        )
        XCTAssertThrowsError(try s.datagram(named: "absent"))
    }

    // MARK: - Async sequence of stream messages

    func testStreamAsyncSequence() async throws {
        let key = Data(repeating: 0xCD, count: 32)
        let ta = LoopbackTransport()
        let tb = LoopbackTransport()
        LoopbackTransport.pair(ta, tb)

        let sa = try PigeonSession(
            masterKey: key, isBackend: false, clientTag: 0,
            datagramChannels: [], transport: ta
        )
        let sb = try PigeonSession(
            masterKey: key, isBackend: false, clientTag: 0,
            datagramChannels: [], transport: tb
        )

        let aChat = try await sa.openStream(name: "chat")
        let accepted = try tb.acceptWithHeader()
        let bChat = sb.adoptAcceptedStream(name: "chat", handle: accepted.handle)

        // Push three messages.
        try await aChat.send(Data("one".utf8))
        try await aChat.send(Data("two".utf8))
        try await aChat.send(Data("three".utf8))

        var seen: [String] = []
        var iter = bChat.messages.makeAsyncIterator()
        for _ in 0..<3 {
            guard let m = try await iter.next() else { break }
            seen.append(String(decoding: m, as: UTF8.self))
        }
        XCTAssertEqual(seen, ["one", "two", "three"])
    }
}

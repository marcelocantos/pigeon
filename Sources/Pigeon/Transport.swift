// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0
//
// Transport bridging — Swift code provides a `PigeonTransport`
// implementation, the bindings hand its underlying `pigeon_transport`
// struct to `pigeon_session_init`. There are two flavours today:
//
//   * `LoopbackTransport` — wraps the in-process loopback transport
//     in CPigeon. Used by the unit tests so we can exercise
//     pigeon_session / pigeon_stream / pigeon_datagram without a
//     real QUIC stack.
//
//   * (Coming with T29 part 2) An ngtcp2-backed transport that uses
//     the multi-stream QUIC implementation in c/src/ngtcp2_transport.c
//     once that worker lands.
//
// User-defined transports written in pure Swift can come later — the
// C-callback shape of pigeon_transport makes it awkward to pass
// closures across the boundary, so for now any new transport is
// expected to be implemented C-side and surfaced via CPigeon.

import CPigeon
import Foundation

/// Anything that can hand pigeon a `pigeon_transport` vtable.
///
/// The concrete implementer owns the lifetime of whatever userdata
/// the C-side transport callbacks dereference. The session captures
/// the transport with a strong reference so the userdata pointer
/// stays valid for as long as the session does.
public protocol PigeonTransport: AnyObject {
    /// Run `body` with a transient pointer to the underlying
    /// pigeon_transport vtable. The pointer is valid only for the
    /// duration of the call.
    func withCTransport<R>(_ body: (UnsafePointer<pigeon_transport>) -> R) -> R

    /// A reference-typed anchor that the session uses to derive a
    /// stable identity (for queue labels). Default returns `self`.
    var box: AnyObject { get }
}

public extension PigeonTransport {
    var box: AnyObject { self }
}

/// In-process loopback transport for tests. Two `LoopbackTransport`s
/// can be paired via `pair(_:_:)` so streams and datagrams sent on
/// one end are queued on the other.
public final class LoopbackTransport: PigeonTransport, @unchecked Sendable {
    // Heap-allocated C endpoint (pigeon_loopback_endpoint*). Lifetime
    // tied to this Swift object. We hold it as OpaquePointer because
    // pigeon_loopback_endpoint is a forward-declared incomplete type
    // in the public header.
    fileprivate let endpoint: OpaquePointer
    private var transportStorage: pigeon_transport

    public init() {
        guard let ep = pigeon_loopback_new() else {
            fatalError("pigeon_loopback_new failed (out of memory)")
        }
        self.endpoint = ep
        self.transportStorage = pigeon_transport()
        pigeon_loopback_fill_transport(&self.transportStorage, ep)
    }

    deinit {
        pigeon_loopback_free(endpoint)
    }

    /// Pair two loopback transports so they route to each other.
    public static func pair(_ a: LoopbackTransport, _ b: LoopbackTransport) {
        pigeon_loopback_pair(a.endpoint, b.endpoint)
    }

    public func withCTransport<R>(_ body: (UnsafePointer<pigeon_transport>) -> R) -> R {
        return withUnsafePointer(to: &transportStorage) { p in
            body(p)
        }
    }

    public var box: AnyObject { self }

    /// Accept the next inbound stream and read its first message
    /// (the unencrypted name-binding header). Returns the opaque
    /// stream handle plus the header bytes the peer wrote. Throws
    /// if no stream is pending.
    ///
    /// `pigeon_stream_handle*` is an incomplete struct in the public
    /// header so Swift sees it as `OpaquePointer`; we round-trip
    /// through that type.
    public func acceptWithHeader() throws -> (handle: OpaquePointer, header: Data) {
        var handle: OpaquePointer? = nil
        // Matches PIGEON_MAX_STREAM_HEADER from pigeon.h: 4-byte tag
        // + max varint (10) + max name (256). Hand-encoded because
        // C macros don't translate through Swift's importer.
        let bufLen = 4 + 10 + 256
        let buf = UnsafeMutablePointer<UInt8>.allocate(capacity: bufLen)
        defer { buf.deallocate() }
        var hdrLen: Int = 0
        let rc = pigeon_loopback_accept_with_header(
            endpoint,
            &handle,
            buf, bufLen,
            &hdrLen
        )
        if rc != 0 || handle == nil {
            throw PigeonSessionError.openStream(rc)
        }
        let data = Data(bytes: buf, count: hdrLen)
        return (handle: handle!, header: data)
    }
}

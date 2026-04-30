// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0
//
// Swift bindings for the post-T22 wire helpers in libpigeon
// (`pigeon_uvarint_*`, `pigeon_encode_stream_header`,
// `pigeon_decode_*_stream_header`). The byte vectors are pinned
// across Go/C/Swift/TypeScript by the cross-language wire-vector
// tests; mirroring them in Swift via the same C entry points
// guarantees no drift.
//
// This is the Swift counterpart of cwire/cwire.go on the Go side.

import CPigeon
import Foundation

public enum PigeonWireError: Error {
    case truncated
    case malformed
    case bufferTooSmall
}

public enum PigeonWire {
    /// Mirrors PIGEON_MAX_STREAM_HEADER from pigeon.h (4-byte tag +
    /// varint name-len + 256 bytes name). Hard-coded because the
    /// Clang-Swift importer can't translate the macro expression.
    static let maxStreamHeader = 4 + 10 + 256

    /// Encode `v` as a Go-style unsigned varint.
    public static func encodeUvarint(_ v: UInt64) -> Data {
        var buf = [UInt8](repeating: 0, count: 10)
        let n = buf.withUnsafeMutableBufferPointer { bp -> Int32 in
            pigeon_uvarint_encode(v, bp.baseAddress, bp.count)
        }
        precondition(n >= 0, "uvarint_encode unexpectedly failed")
        return Data(buf.prefix(Int(n)))
    }

    /// Decode a varint at the start of `buf`. Returns the value and
    /// the number of bytes consumed.
    public static func decodeUvarint(_ buf: Data) throws -> (value: UInt64, consumed: Int) {
        if buf.isEmpty { throw PigeonWireError.truncated }
        var v: UInt64 = 0
        let n: Int = buf.withUnsafeBytes { rb -> Int in
            let p = rb.bindMemory(to: UInt8.self).baseAddress!
            return Int(pigeon_uvarint_decode(p, buf.count, &v))
        }
        switch n {
        case 0: throw PigeonWireError.truncated
        case let k where k < 0: throw PigeonWireError.malformed
        default: return (v, n)
        }
    }

    /// Encode the per-stream first-message header.
    ///
    /// Backend side: `[4-byte clientTag-BE][varint name-len][name]`
    /// Client side:  `[varint name-len][name]`
    public static func encodeStreamHeader(
        isBackend: Bool, clientTag: UInt32, name: String
    ) throws -> Data {
        let bufLen = Self.maxStreamHeader
        let buf = UnsafeMutablePointer<UInt8>.allocate(capacity: bufLen)
        defer { buf.deallocate() }
        let nameBytes = Array(name.utf8)
        let n: Int32 = nameBytes.withUnsafeBufferPointer { nb -> Int32 in
            // nb.baseAddress may be nil for empty strings; the C side
            // tolerates name == NULL iff name_len == 0.
            return pigeon_encode_stream_header(
                isBackend, clientTag,
                nb.baseAddress, nameBytes.count,
                buf, bufLen
            )
        }
        if n < 0 { throw PigeonWireError.bufferTooSmall }
        return Data(bytes: buf, count: Int(n))
    }

    /// Decode a backend-side header. Returns clientTag, name, and
    /// total bytes consumed.
    public static func decodeBackendStreamHeader(_ data: Data)
        throws -> (clientTag: UInt32, name: String, consumed: Int)
    {
        if data.isEmpty { throw PigeonWireError.truncated }
        var tag: UInt32 = 0
        let nameBufLen = 256
        let nameBuf = UnsafeMutablePointer<CChar>.allocate(capacity: nameBufLen)
        defer { nameBuf.deallocate() }
        var nameLen: Int = 0
        let n: Int = data.withUnsafeBytes { rb -> Int in
            let p = rb.bindMemory(to: UInt8.self).baseAddress!
            return Int(pigeon_decode_backend_stream_header(
                p, data.count,
                &tag,
                nameBuf, nameBufLen,
                &nameLen
            ))
        }
        if n < 0 { throw PigeonWireError.malformed }
        let name = String(cString: nameBuf)
        return (tag, name, n)
    }

    /// Decode a client-side header. Returns name and bytes consumed.
    public static func decodeClientStreamHeader(_ data: Data)
        throws -> (name: String, consumed: Int)
    {
        if data.isEmpty { throw PigeonWireError.truncated }
        let nameBufLen = 256
        let nameBuf = UnsafeMutablePointer<CChar>.allocate(capacity: nameBufLen)
        defer { nameBuf.deallocate() }
        var nameLen: Int = 0
        let n: Int = data.withUnsafeBytes { rb -> Int in
            let p = rb.bindMemory(to: UInt8.self).baseAddress!
            return Int(pigeon_decode_client_stream_header(
                p, data.count,
                nameBuf, nameBufLen,
                &nameLen
            ))
        }
        if n < 0 { throw PigeonWireError.malformed }
        let name = String(cString: nameBuf)
        return (name, n)
    }
}

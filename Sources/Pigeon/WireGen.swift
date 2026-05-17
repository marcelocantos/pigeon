// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0
//
// Code generated from protocol/wireformats.yaml. DO NOT EDIT.

import Foundation

public enum PigeonWireError: Error {
    case truncated
    case malformed
    case bufferTooSmall
}

public enum PigeonWire {

    public static func encodeUvarint(_ v: UInt64) -> Data {
        var out = Data()
        var x = v
        while x >= 0x80 {
            out.append(UInt8((x & 0x7f) | 0x80))
            x >>= 7
        }
        out.append(UInt8(x))
        return out
    }

    public static func decodeUvarint(_ data: Data) throws -> (value: UInt64, consumed: Int) {
        var v: UInt64 = 0
        var shift: UInt64 = 0
        for (i, b) in data.enumerated() {
            if i >= 10 { throw PigeonWireError.malformed }
            if b & 0x80 == 0 { v |= UInt64(b) << shift; return (v, i + 1) }
            v |= UInt64(b & 0x7f) << shift
            shift += 7
        }
        throw PigeonWireError.truncated
    }

    public static func encodeStreamHeader(isBackend: Bool, clientTag: UInt32, name: String) -> Data {
        var out = Data()
        if isBackend {
            out.append(contentsOf: [UInt8(truncatingIfNeeded: clientTag >> 24), UInt8(truncatingIfNeeded: clientTag >> 16), UInt8(truncatingIfNeeded: clientTag >> 8), UInt8(truncatingIfNeeded: clientTag)])
            do {
                let bytes = Array(name.utf8)
                out.append(PigeonWire.encodeUvarint(UInt64(bytes.count)))
                out.append(contentsOf: bytes)
            }
        } else {
            do {
                let bytes = Array(name.utf8)
                out.append(PigeonWire.encodeUvarint(UInt64(bytes.count)))
                out.append(contentsOf: bytes)
            }
        }
        return out
    }

    public static func decodeStreamHeaderBackend(_ data: Data) throws -> (clientTag: UInt32, name: String, consumed: Int) {
        var off = 0
        if data.count - off < 4 { throw PigeonWireError.truncated }
        let clientTag: UInt32 = (UInt32(data[data.startIndex + off]) << 24)
            | (UInt32(data[data.startIndex + off + 1]) << 16)
            | (UInt32(data[data.startIndex + off + 2]) << 8)
            | UInt32(data[data.startIndex + off + 3])
        off += 4
        let (_lenU, _ln) = try PigeonWire.decodeUvarint(data.subdata(in: (data.startIndex + off)..<data.endIndex))
        off += _ln
        let _len = Int(_lenU)
        if data.count - off < _len { throw PigeonWireError.truncated }
        let name = String(data: data.subdata(in: (data.startIndex + off)..<(data.startIndex + off + _len)), encoding: .utf8) ?? ""
        off += _len
        return (clientTag, name, off)
    }

    public static func decodeStreamHeaderClient(_ data: Data) throws -> (name: String, consumed: Int) {
        var off = 0
        let (_lenU, _ln) = try PigeonWire.decodeUvarint(data.subdata(in: (data.startIndex + off)..<data.endIndex))
        off += _ln
        let _len = Int(_lenU)
        if data.count - off < _len { throw PigeonWireError.truncated }
        let name = String(data: data.subdata(in: (data.startIndex + off)..<(data.startIndex + off + _len)), encoding: .utf8) ?? ""
        off += _len
        return (name, off)
    }

    public static func encodeDatagramPlaintext(channelId: UInt64, payload: Data) -> Data {
        var out = Data()
        out.append(PigeonWire.encodeUvarint(channelId))
        out.append(payload)
        return out
    }

    public static func decodeDatagramPlaintext(_ data: Data) throws -> (channelId: UInt64, payload: Data, consumed: Int) {
        var off = 0
        let (channelId, _vn) = try PigeonWire.decodeUvarint(data.subdata(in: (data.startIndex + off)..<data.endIndex))
        off += _vn
        let payload = data.subdata(in: (data.startIndex + off)..<data.endIndex)
        off = data.count
        return (channelId, payload, off)
    }

    public enum RelayGreetingVariant {
        case connect
        case registerMux
    }

    public struct RelayGreetingDecoded {
        public let variant: RelayGreetingVariant
        public let instanceId: String
        public let token: String
    }

    public static func encodeRelayGreetingConnect(instanceId: String) -> Data {
        var s = "connect:"
        s += instanceId
        return Data(s.utf8)
    }

    public static func encodeRelayGreetingRegisterMux(token: String, instanceId: String) -> Data {
        var s = "register-mux"
        let parts: [String] = [token, instanceId]
        var anyNonEmpty = false
        for p in parts { if !p.isEmpty { anyNonEmpty = true } }
        if anyNonEmpty {
            for p in parts { s += ":"; s += p }
        }
        return Data(s.utf8)
    }

    public static func decodeRelayGreeting(_ data: Data) throws -> RelayGreetingDecoded {
        let s = String(data: data, encoding: .utf8) ?? ""
        if s.hasPrefix("register-mux") {
            let rest = String(s.dropFirst(12))
            var _suffix: [String] = []
            if !rest.isEmpty {
                if !rest.hasPrefix(":") { throw PigeonWireError.malformed }
                let body = String(rest.dropFirst())
                _suffix = body.components(separatedBy: ":")
                while _suffix.count < 2 { _suffix.append("") }
                if _suffix.count > 2 {
                    _suffix[1] = _suffix[1...].joined(separator: ":")
                    _suffix = Array(_suffix.prefix(2))
                }
            } else {
                _suffix = Array(repeating: "", count: 2)
            }
            return RelayGreetingDecoded(
                variant: .registerMux,
                instanceId: _suffix[1],
                token: _suffix[0]
            )
        }
        if s.hasPrefix("connect:") {
            let rest = String(s.dropFirst(8))
            return RelayGreetingDecoded(
                variant: .connect,
                instanceId: rest,
                token: ""
            )
        }
        throw PigeonWireError.malformed
    }

}

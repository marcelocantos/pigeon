// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0
//
// pairing-peer-swift drives the Swift PairingCeremony as the initiator
// against an external Go acceptor that talks the pairing wire over the
// process's stdin/stdout. Used by the Go cross-language pairing test
// (pairing/cross_swift_test.go) to verify that Swift and Go agree on:
//   - the JSON wire (hello/welcome/confirm),
//   - the 6-digit confirmation code derivation, and
//   - the post-ceremony PairingRecord's deriveChannel outputs.
//
// The acceptor's ephemeral pubkey, instance ID, and the initiator's own
// instance ID + identity pubkey are passed via CLI args (base64 / utf8).
// The framing on stdin/stdout matches Go's pairing.singleStreamTransport:
// 4-byte big-endian length prefix, then payload, one logical message per
// frame. The 6-digit code is auto-confirmed (we treat the test acceptor
// as trusted).
//
// On success, prints a JSON line to stdout:
//   {"code":"<6-digit>","record":{...PairingRecord JSON...}}
// and exits 0. The framed wire and the result are disjoint: the
// 4-byte framing only ever flows on stdin (Go→Swift, welcome+confirm)
// and stdout (Swift→Go, hello+confirm), and the result JSON is printed
// to stderr so the parent can read it without colliding with framed wire
// bytes on stdout. Errors go to stderr too.
//
// Usage:
//   pairing-peer-swift \
//     --acceptor-eph-pub-b64=<base64-standard-pad-32-bytes> \
//     --acceptor-instance=<string> \
//     --initiator-instance=<string> \
//     --initiator-identity-pub-b64=<base64-standard-pad-32-bytes> \
//     --initiator-eph-priv-b64=<base64-standard-pad-32-bytes> \
//     --initiator-eph-pub-b64=<base64-standard-pad-32-bytes>

import Foundation
import Pigeon

// MARK: - CLI args

let args = ProcessInfo.processInfo.arguments
func cliArg(_ name: String) -> String? {
    for a in args {
        if a.hasPrefix("--\(name)=") {
            return String(a.dropFirst("--\(name)=".count))
        }
    }
    return nil
}
func requireArg(_ name: String) -> String {
    guard let v = cliArg(name) else {
        FileHandle.standardError.write(Data("missing --\(name)\n".utf8))
        exit(2)
    }
    return v
}

let acceptorEphPubB64 = requireArg("acceptor-eph-pub-b64")
let acceptorInstance  = requireArg("acceptor-instance")
let initiatorInstance = requireArg("initiator-instance")
let initiatorIdPubB64 = requireArg("initiator-identity-pub-b64")
let initiatorEphPrivB64 = requireArg("initiator-eph-priv-b64")
let initiatorEphPubB64  = requireArg("initiator-eph-pub-b64")

func mustB64(_ s: String, _ name: String) -> Data {
    guard let d = Data(base64Encoded: s) else {
        FileHandle.standardError.write(Data("invalid base64 in \(name)\n".utf8))
        exit(2)
    }
    return d
}
let acceptorEphPub  = mustB64(acceptorEphPubB64, "acceptor-eph-pub-b64")
let initiatorIdPub  = mustB64(initiatorIdPubB64, "initiator-identity-pub-b64")
let initiatorEphPriv = mustB64(initiatorEphPrivB64, "initiator-eph-priv-b64")
let initiatorEphPub  = mustB64(initiatorEphPubB64,  "initiator-eph-pub-b64")

// MARK: - Stdio framed transport (4-byte BE length prefix)

/// Drives Go pairing.singleStreamTransport's wire format over stdin/stdout.
/// One full message per send/recv; matches encoding/binary.BigEndian
/// uint32 length prefix.
final class StdioFramedTransport: PairingWireTransport, @unchecked Sendable {
    private let inFD: FileHandle = FileHandle.standardInput
    private let outFD: FileHandle = FileHandle.standardOutput

    func send(_ data: Data) async throws {
        let len = UInt32(data.count)
        var hdr = Data(count: 4)
        hdr[0] = UInt8((len >> 24) & 0xFF)
        hdr[1] = UInt8((len >> 16) & 0xFF)
        hdr[2] = UInt8((len >> 8)  & 0xFF)
        hdr[3] = UInt8(len         & 0xFF)
        try outFD.write(contentsOf: hdr)
        try outFD.write(contentsOf: data)
    }

    func recv() async throws -> Data {
        let hdr = try readExact(4)
        let len = (UInt32(hdr[0]) << 24)
            | (UInt32(hdr[1]) << 16)
            | (UInt32(hdr[2]) << 8)
            | UInt32(hdr[3])
        if len == 0 { return Data() }
        return try readExact(Int(len))
    }

    func close() throws {}

    private func readExact(_ count: Int) throws -> Data {
        var buf = Data()
        while buf.count < count {
            let remaining = count - buf.count
            guard let chunk = try inFD.read(upToCount: remaining), !chunk.isEmpty else {
                throw PairingCeremonyError.malformedMessage("EOF on stdin (need \(remaining) more bytes)")
            }
            buf.append(chunk)
        }
        return buf
    }
}

// MARK: - Run

let transport = StdioFramedTransport()
let identity = PairingIdentity(
    publicKey: initiatorIdPub,
    instanceID: initiatorInstance
)

// We pin the relayURL to a fixed string so the Go side can verify that
// the post-ceremony record's relay_url round-trips correctly across the
// language boundary. The acceptor's relay_url is set by the Go side
// independently.
let relayURL = "https://relay.test"

func runCeremony() async {
    do {
        let res = try await pairInitiator(
            transport: transport,
            identity: identity,
            initiatorEphPriv: initiatorEphPriv,
            initiatorEphPub: initiatorEphPub,
            acceptorEphPub: acceptorEphPub,
            acceptorInstance: acceptorInstance,
            relayURL: relayURL,
            confirm: { _ in true }
        )
        // Emit the result to stderr so the Go parent can parse it
        // independently of the framed wire bytes on stdout.
        let payload: [String: Any] = [
            "code": res.code,
            "record": [
                "peer_instance_id":   res.record.peerInstanceID,
                "relay_url":          res.record.relayURL,
                "local_private_key":  res.record.localPrivateKey.base64EncodedString(),
                "local_public_key":   res.record.localPublicKey.base64EncodedString(),
                "peer_public_key":    res.record.peerPublicKey.base64EncodedString(),
            ],
        ]
        let resultJSON = try JSONSerialization.data(withJSONObject: payload, options: [.sortedKeys])
        FileHandle.standardError.write(Data("RESULT ".utf8))
        FileHandle.standardError.write(resultJSON)
        FileHandle.standardError.write(Data("\n".utf8))
        exit(0)
    } catch {
        FileHandle.standardError.write(Data("ERROR \(error)\n".utf8))
        exit(1)
    }
}

let sem = DispatchSemaphore(value: 0)
Task { await runCeremony(); sem.signal() }
sem.wait()

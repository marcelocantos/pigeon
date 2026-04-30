// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0
//
// Idiomatic Swift façade over the libpigeon multi-channel session
// API (pigeon_session / pigeon_stream / pigeon_datagram + the wire
// helpers in pigeon.h). This is the Swift counterpart to cwire.go in
// Go and the post-T22 multi-channel `Session`/`Stream`/`Datagram`
// types in the Go peer library.
//
// Today the underlying transport is provided by the caller; the
// reference transport is the in-process loopback in CPigeon (see
// LoopbackTransport.swift), which we use for the Swift unit tests.
// Once the multi-stream ngtcp2 wrapper from T32-step-3 lands the C
// side will gain a real QUIC transport and this Swift API will pick
// it up unchanged.

import CPigeon
import Foundation

// MARK: - Errors

public enum PigeonSessionError: LocalizedError {
    case sessionInit(Int32)
    case openStream(Int32)
    case unknownDatagramChannel(String)
    case streamSend(Int32)
    case streamRecv(Int32)
    case streamClose(Int32)
    case datagramSend(Int32)
    case datagramRecv(Int32)
    case nameTooLong(String)
    case channelInit
    case bufferTooSmall(Int)

    public var errorDescription: String? {
        switch self {
        case .sessionInit(let rc):
            return "pigeon_session_init failed (rc=\(rc))"
        case .openStream(let rc):
            return "pigeon_session_open_stream failed (rc=\(rc))"
        case .unknownDatagramChannel(let name):
            return "no datagram channel named '\(name)'"
        case .streamSend(let rc):
            return "pigeon_stream_send failed (rc=\(rc))"
        case .streamRecv(let rc):
            return "pigeon_stream_recv failed (rc=\(rc))"
        case .streamClose(let rc):
            return "pigeon_stream_close failed (rc=\(rc))"
        case .datagramSend(let rc):
            return "pigeon_datagram_send failed (rc=\(rc))"
        case .datagramRecv(let rc):
            return "pigeon_datagram_recv failed (rc=\(rc))"
        case .nameTooLong(let s):
            return "name too long for libpigeon: '\(s)'"
        case .channelInit:
            return "pigeon_channel_init_symmetric failed"
        case .bufferTooSmall(let needed):
            return "receive buffer too small (need \(needed) bytes)"
        }
    }
}

// MARK: - DatagramChannelDef

/// A pre-declared datagram channel name → id mapping. Both peers
/// must declare an identical list at session-init time.
public struct DatagramChannelDef: Sendable, Hashable {
    public var name: String
    public var channelID: UInt64

    public init(name: String, channelID: UInt64) {
        self.name = name
        self.channelID = channelID
    }
}

// MARK: - PigeonSession

/// One peer-to-peer association: backend ↔ paired client (or vice-
/// versa from the client side). Owns the AEAD channel derived from
/// the PairingRecord, the role discriminator (backend/client) and
/// the relay-assigned clientTag (backend side only), plus a fixed-
/// size table of pre-declared datagram channels.
///
/// Mirrors the Go peer-library `Session` — same shape, same
/// constraints (datagrams pre-declared, streams opened on demand).
public final class PigeonSession: @unchecked Sendable {
    // Heap-allocated C session struct. We hold it via
    // UnsafeMutablePointer so its address is stable for the
    // lifetime of this object — pigeon_stream / pigeon_datagram
    // both refer back to it.
    private let sessionPtr: UnsafeMutablePointer<pigeon_session>

    // Each Session owns its own fixed-size key buffers (rather than
    // borrowing from the caller) so the AEAD channel pointer inside
    // the C struct can remain valid for the session's lifetime.
    private let channelStorage: UnsafeMutablePointer<pigeon_channel>

    // The Session serialises access to its underlying transport
    // and the per-session pigeon_channel sequence counters via this
    // worker thread. Every C call into pigeon_session_* /
    // pigeon_stream_* / pigeon_datagram_* is dispatched here.
    //
    // We can't run these on a regular DispatchQueue worker (or on
    // Swift's cooperative thread pool) because pigeon_stream_send /
    // _recv and pigeon_datagram_send / _recv stack-allocate a
    // ~1 MiB ciphertext buffer (PIGEON_MAX_MSG + AEAD overhead),
    // which is bigger than either of those pools' stack budgets and
    // blows the stack on first call. The dedicated thread is
    // configured with a 4 MiB stack which leaves comfortable
    // headroom for the C frames plus any Swift bridging.
    private let worker: BigStackWorker

    public let isBackend: Bool
    public let clientTag: UInt32

    /// Initialise a session from a 32-byte symmetric AEAD master key.
    /// `isBackend == true` means this Session is the server-side half
    /// of the pair; outbound stream/datagram framing then includes
    /// the 4-byte clientTag prefix the relay routes by.
    public init(
        masterKey: Data,
        isBackend: Bool,
        clientTag: UInt32 = 0,
        datagramChannels: [DatagramChannelDef] = [],
        transport: PigeonTransport
    ) throws {
        precondition(masterKey.count == 32, "masterKey must be 32 bytes")

        let session = UnsafeMutablePointer<pigeon_session>.allocate(capacity: 1)
        session.initialize(to: pigeon_session())

        let ch = UnsafeMutablePointer<pigeon_channel>.allocate(capacity: 1)
        ch.initialize(to: pigeon_channel())

        // Initialise the AEAD channel with the same key for send and
        // recv. This mirrors the C reference loopback test in
        // c/test/test_pigeon.c::test_session_*_roundtrip — both peers
        // use a shared symmetric key, so encrypt/decrypt is reversible
        // in either direction. Production callers will instead
        // construct the channel from a PairingRecord-derived
        // send_key/recv_key pair (the keys differ between peers, the
        // channel still uses pigeon_channel_init), and that path
        // will be exposed via a separate initialiser once the real
        // multi-stream ngtcp2 transport lands and we wire the full
        // PairingRecord plumbing through.
        masterKey.withUnsafeBytes { mkBuf in
            let p = mkBuf.bindMemory(to: UInt8.self).baseAddress!
            let mode = datagramChannels.isEmpty
                ? PIGEON_MODE_STRICT
                : PIGEON_MODE_DATAGRAMS
            pigeon_channel_init(ch, p, p, mode)
        }

        // Translate the Swift datagram channel list into a contiguous
        // array of pigeon_dgchannel_def. Names are zero-padded into
        // the fixed 64-byte buffer.
        var dgArray = [pigeon_dgchannel_def](
            repeating: pigeon_dgchannel_def(),
            count: max(1, datagramChannels.count)
        )
        for (i, d) in datagramChannels.enumerated() {
            let bytes = Array(d.name.utf8)
            // PIGEON_MAX_NAME_LEN is 64 in pigeon.h.
            if bytes.count >= 64 {
                ch.deinitialize(count: 1); ch.deallocate()
                session.deinitialize(count: 1); session.deallocate()
                throw PigeonSessionError.nameTooLong(d.name)
            }
            withUnsafeMutableBytes(of: &dgArray[i].name) { rawDest in
                let dest = rawDest.bindMemory(to: UInt8.self).baseAddress!
                bytes.withUnsafeBufferPointer { src in
                    if let s = src.baseAddress {
                        memcpy(dest, s, bytes.count)
                    }
                }
            }
            dgArray[i].channel_id = d.channelID
        }

        let initSessRC: Int32 = transport.withCTransport { tPtr in
            dgArray.withUnsafeBufferPointer { dgBuf -> Int32 in
                pigeon_session_init(
                    session,
                    tPtr,
                    ch,
                    isBackend,
                    clientTag,
                    datagramChannels.isEmpty ? nil : dgBuf.baseAddress,
                    datagramChannels.count
                )
            }
        }
        if initSessRC != 0 {
            ch.deinitialize(count: 1); ch.deallocate()
            session.deinitialize(count: 1); session.deallocate()
            throw PigeonSessionError.sessionInit(initSessRC)
        }

        self.sessionPtr = session
        self.channelStorage = ch
        self.isBackend = isBackend
        self.clientTag = clientTag
        self.worker = BigStackWorker()

        // Anchor the transport so its userdata pointer stays valid.
        self.transportAnchor = transport
    }

    deinit {
        sessionPtr.deinitialize(count: 1)
        sessionPtr.deallocate()
        channelStorage.deinitialize(count: 1)
        channelStorage.deallocate()
    }

    // Strong reference to the transport object so its userdata
    // pointer stays valid as long as the session does.
    private let transportAnchor: PigeonTransport

    // MARK: Streams

    /// Open a new named stream toward the peer. Writes the
    /// `[optional 4-byte tag][varint name-len][name]` header as the
    /// first message on the stream, AEAD-encrypts subsequent
    /// messages.
    public func openStream(name: String) async throws -> PigeonStream {
        if name.utf8.count >= 64 {
            throw PigeonSessionError.nameTooLong(name)
        }
        let streamPtr = UnsafeMutablePointer<pigeon_stream>.allocate(capacity: 1)
        streamPtr.initialize(to: pigeon_stream())
        let rc: Int32 = await worker.run {
            name.withCString { cstr -> Int32 in
                pigeon_session_open_stream(self.sessionPtr, cstr, streamPtr)
            }
        }
        if rc != 0 {
            streamPtr.deinitialize(count: 1)
            streamPtr.deallocate()
            throw PigeonSessionError.openStream(rc)
        }
        return PigeonStream(parent: self, ptr: streamPtr, worker: worker)
    }

    /// Adopt a transport-side stream handle that was already accepted
    /// (e.g. by the loopback's `acceptWithHeader`). The first message
    /// on the wire is assumed to have already been consumed. Returns
    /// a `PigeonStream` that AEAD-encrypts subsequent messages.
    public func adoptAcceptedStream(name: String,
                                    handle: OpaquePointer) -> PigeonStream {
        let streamPtr = UnsafeMutablePointer<pigeon_stream>.allocate(capacity: 1)
        var s = pigeon_stream()
        s.session = sessionPtr
        // pigeon_stream.handle is `pigeon_stream_handle*` (an
        // incomplete C type), imported as OpaquePointer.
        s.handle = handle
        // Copy the name into the fixed buffer (NUL-terminated).
        let bytes = Array(name.utf8)
        precondition(bytes.count < 64, "name too long")
        withUnsafeMutableBytes(of: &s.name) { rawDest in
            let dest = rawDest.bindMemory(to: UInt8.self).baseAddress!
            bytes.withUnsafeBufferPointer { src in
                if let p = src.baseAddress {
                    memcpy(dest, p, bytes.count)
                }
            }
        }
        streamPtr.initialize(to: s)
        return PigeonStream(parent: self, ptr: streamPtr, worker: worker)
    }

    // MARK: Datagrams

    /// Look up a pre-declared datagram channel by name. Returns a
    /// `PigeonDatagram` that AEAD-encrypts and frames each send.
    public func datagram(named name: String) throws -> PigeonDatagram {
        let dgPtr = UnsafeMutablePointer<pigeon_datagram>.allocate(capacity: 1)
        dgPtr.initialize(to: pigeon_datagram())
        let rc = name.withCString { cstr -> Int32 in
            pigeon_session_get_datagram(sessionPtr, cstr, dgPtr)
        }
        if rc != 0 {
            dgPtr.deinitialize(count: 1)
            dgPtr.deallocate()
            throw PigeonSessionError.unknownDatagramChannel(name)
        }
        return PigeonDatagram(parent: self, ptr: dgPtr, worker: worker)
    }
}

// MARK: - PigeonStream

/// A bidirectional stream of AEAD-encrypted application messages.
public final class PigeonStream: @unchecked Sendable {
    private let parent: PigeonSession
    private let ptr: UnsafeMutablePointer<pigeon_stream>
    private let worker: BigStackWorker

    fileprivate init(parent: PigeonSession,
                     ptr: UnsafeMutablePointer<pigeon_stream>,
                     worker: BigStackWorker) {
        self.parent = parent
        self.ptr = ptr
        self.worker = worker
    }

    deinit {
        ptr.deinitialize(count: 1)
        ptr.deallocate()
    }

    public func send(_ data: Data) async throws {
        let rc: Int32 = await worker.run {
            data.withUnsafeBytes { rb -> Int32 in
                let base = rb.bindMemory(to: UInt8.self).baseAddress
                return pigeon_stream_send(self.ptr, base, data.count)
            }
        }
        if rc != 0 {
            throw PigeonSessionError.streamSend(rc)
        }
    }

    public func recv() async throws -> Data {
        let result: Result<Data, PigeonSessionError> = await worker.run {
            let buf = UnsafeMutablePointer<UInt8>.allocate(capacity: Int(PIGEON_MAX_MSG))
            defer { buf.deallocate() }
            let n = pigeon_stream_recv(self.ptr, buf, Int(PIGEON_MAX_MSG))
            if n < 0 {
                return .failure(.streamRecv(n))
            }
            return .success(Data(bytes: buf, count: Int(n)))
        }
        return try result.get()
    }

    /// Async sequence of incoming messages. Yields each AEAD-decrypted
    /// payload as `Data`; finishes when `recv` returns an error.
    public var messages: AsyncThrowingStream<Data, Error> {
        AsyncThrowingStream { cont in
            let task = Task { [self] in
                while !Task.isCancelled {
                    do {
                        let m = try await self.recv()
                        cont.yield(m)
                    } catch {
                        cont.finish(throwing: error)
                        return
                    }
                }
                cont.finish()
            }
            cont.onTermination = { _ in task.cancel() }
        }
    }

    public func close() throws {
        let rc = pigeon_stream_close(ptr)
        if rc != 0 { throw PigeonSessionError.streamClose(rc) }
    }
}

// MARK: - PigeonDatagram

/// One named datagram channel within a session. Each `send` AEAD-
/// encrypts the payload and frames it with the channel id (and the
/// 4-byte clientTag prefix on the backend side).
public final class PigeonDatagram: @unchecked Sendable {
    private let parent: PigeonSession
    private let ptr: UnsafeMutablePointer<pigeon_datagram>
    private let worker: BigStackWorker

    fileprivate init(parent: PigeonSession,
                     ptr: UnsafeMutablePointer<pigeon_datagram>,
                     worker: BigStackWorker) {
        self.parent = parent
        self.ptr = ptr
        self.worker = worker
    }

    deinit {
        ptr.deinitialize(count: 1)
        ptr.deallocate()
    }

    public func send(_ payload: Data) async throws {
        let rc: Int32 = await worker.run {
            payload.withUnsafeBytes { rb -> Int32 in
                let base = rb.bindMemory(to: UInt8.self).baseAddress
                return pigeon_datagram_send(self.ptr, base, payload.count)
            }
        }
        if rc != 0 {
            throw PigeonSessionError.datagramSend(rc)
        }
    }

    /// Receive the next datagram on this named channel. If the next
    /// inbound datagram is for a different channel, returns nil so
    /// the caller can re-dispatch (matches the C semantics where
    /// `pigeon_datagram_recv` returns 0 for "not for me").
    public func recv() async throws -> Data? {
        enum Outcome { case ok(Data), notForMe, fail(Int32) }
        let outcome: Outcome = await worker.run {
            let buf = UnsafeMutablePointer<UInt8>.allocate(capacity: Int(PIGEON_MAX_MSG))
            defer { buf.deallocate() }
            let n = pigeon_datagram_recv(self.ptr, buf, Int(PIGEON_MAX_MSG))
            if n < 0 { return .fail(n) }
            if n == 0 { return .notForMe }
            return .ok(Data(bytes: buf, count: Int(n)))
        }
        switch outcome {
        case .ok(let d): return d
        case .notForMe: return nil
        case .fail(let rc): throw PigeonSessionError.datagramRecv(rc)
        }
    }
}

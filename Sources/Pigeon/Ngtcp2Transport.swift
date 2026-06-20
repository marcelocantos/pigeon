// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0
//
// Real QUIC transport for PigeonSession. Wraps the C
// pigeon_ngtcp2_transport struct (c/src/ngtcp2_transport.c) and
// exposes it to the Swift Session/Stream/Datagram façade through the
// PigeonTransport vtable bridging in Transport.swift.
//
// Two roles are supported, mirroring the C-side `pigeon_role`:
//
//   * .connect(peerInstanceID:) — client-side: opens a QUIC connection
//     to the relay and asks it to route to a registered backend by ID.
//   * .registerMux(selfInstanceID:) — backend-side multi-client: opens
//     the QUIC connection and runs the register-mux handshake. The
//     relay echoes back the assigned instance ID into the transport
//     struct.
//
// T45 status: the relay and the Swift-level relay client (PigeonConn,
// WireGen greetings) are on the remote-Listen L1 wire (register control
// connection + listen pool + per-client end-to-end bridge; see
// docs/DESIGN.md §3 L1). This transport's greetings are emitted by the
// C layer (pigeon_ngtcp2_transport_init → PIGEON_ROLE_*), which still
// speaks the pre-T45 register-mux wire and does not consume the relay's
// connect "ok" ack. The role names below therefore still mirror the C
// `pigeon_role`; they migrate to register / listen / connect once the C
// SDK transport is ported (the cross-language E2E test that drives this
// path is skipped until then — see RelayE2ETests).
//
// Lifetime: Ngtcp2Transport is a reference type (class). The underlying
// C struct lives on the heap (allocated in this Swift object) so its
// address stays stable for the lifetime of the Swift object — that
// pointer is handed to the vtable as userdata. `close()` and `deinit`
// both call pigeon_ngtcp2_transport_close, which is idempotent.
//
// Threading: ngtcp2 is not thread-safe and the underlying transport
// uses blocking select() loops in send/recv. Drive it from a single
// thread (typically PigeonSession's BigStackWorker, which already
// serialises calls into pigeon_session_* / pigeon_stream_* /
// pigeon_datagram_*). Do NOT share an Ngtcp2Transport between two
// concurrent PigeonSession instances.

import CPigeon
import Foundation

/// Errors raised by `Ngtcp2Transport.init`.
public enum Ngtcp2TransportError: LocalizedError {
    /// pigeon_ngtcp2_transport_init failed. `message` is the
    /// transport's `last_error` field (capped at 127 chars in C).
    case initFailed(message: String)

    /// pigeon_ngtcp2_transport_primary_handle returned NULL (no
    /// primary stream open, or the extra-streams slot table is full).
    case noPrimaryHandle

    public var errorDescription: String? {
        switch self {
        case .initFailed(let m): return "ngtcp2 transport init failed: \(m)"
        case .noPrimaryHandle:   return "ngtcp2 transport has no primary stream handle"
        }
    }
}

/// Role discriminator for an Ngtcp2Transport, mirroring the C-side
/// `pigeon_role`.
public enum Ngtcp2Role: Sendable {
    /// Client side. The transport sends `connect:<peerInstanceID>` on
    /// the primary QUIC stream during init and routes through the
    /// relay to the registered backend with that ID.
    case connect(peerInstanceID: String)

    /// Backend side, multi-client. The transport runs the
    /// register-mux handshake; the relay assigns (or echoes) the
    /// instance ID. After init, read it via
    /// `Ngtcp2Transport.instanceID`.
    case registerMux(selfInstanceID: String? = nil, token: String? = nil)
}

/// QUIC transport for PigeonSession, backed by the vendored ngtcp2 +
/// quictls (OpenSSL) stack. Each Ngtcp2Transport owns one QUIC
/// connection.
public final class Ngtcp2Transport: PigeonTransport, @unchecked Sendable {
    // Heap-allocated C transport struct. We own its lifetime; the
    // struct itself is caller-allocated per the C-side memory model.
    // We free it after pigeon_ngtcp2_transport_close.
    private let cTransport: UnsafeMutablePointer<pigeon_ngtcp2_transport>
    private var closed: Bool = false

    /// The instance ID associated with this transport. For
    /// `.connect`, this is the peer's ID. For `.registerMux`, it is
    /// the relay-assigned (or echoed self-assigned) ID after the
    /// handshake completes.
    public let instanceID: String

    /// Create the transport and run the QUIC handshake plus the
    /// pigeon role handshake. Blocks the calling thread for up to
    /// `timeoutMs` milliseconds (default 10s — the C-side default
    /// when 0 is passed).
    ///
    /// - Parameters:
    ///   - host: relay hostname or IP (required).
    ///   - port: relay UDP port (e.g. "4433").
    ///   - role: connect (client) or register-mux (backend).
    ///   - verifyPeer: enable server certificate verification.
    ///     Defaults to false to accept the development relay's
    ///     self-signed certificate.
    ///   - caCertFile: path to a CA bundle PEM. Only consulted when
    ///     `verifyPeer == true`.
    ///   - timeoutMs: overall connect + handshake timeout. Pass 0 to
    ///     use the C-side default (10000 ms).
    public init(
        host: String,
        port: String,
        role: Ngtcp2Role,
        verifyPeer: Bool = false,
        caCertFile: String? = nil,
        timeoutMs: Int32 = 0
    ) throws {
        let tPtr = UnsafeMutablePointer<pigeon_ngtcp2_transport>.allocate(capacity: 1)
        // The struct is ~1.1 MiB (17 × 64 KiB ringbufs + slot tables);
        // do NOT construct it as a Swift value type via
        // `pigeon_ngtcp2_transport()` — that materialises the whole
        // thing on the stack and blows the cooperative-pool thread's
        // ~544 KiB stack. memset the heap allocation directly, which
        // is what the C-side contract asks for anyway ("zero-
        // initialise before calling init").
        memset(UnsafeMutableRawPointer(tPtr), 0,
               MemoryLayout<pigeon_ngtcp2_transport>.size)

        // Marshal the role enum into the C fields.
        let cRole: pigeon_role
        let instanceArg: String?
        let tokenArg: String?
        switch role {
        case .connect(let peerID):
            cRole = PIGEON_ROLE_CONNECT
            instanceArg = peerID
            tokenArg = nil
        case .registerMux(let selfID, let token):
            cRole = PIGEON_ROLE_REGISTER_MUX
            instanceArg = selfID
            tokenArg = token
        }

        // Build the C config inside the closure so the CString pointers
        // live until pigeon_ngtcp2_transport_init returns — it copies
        // the strings into the transport struct internally.
        let rc: Int32 = host.withCString { hostC in
            port.withCString { portC in
                instanceArg.withCStringOrNull { instC in
                    tokenArg.withCStringOrNull { tokC in
                        caCertFile.withCStringOrNull { caC in
                            var cfg = pigeon_ngtcp2_config()
                            cfg.host = hostC
                            cfg.port = portC
                            cfg.instance_id = instC
                            cfg.verify_peer = verifyPeer ? 1 : 0
                            cfg.ca_cert_file = caC
                            cfg.timeout_ms = timeoutMs
                            cfg.role = cRole
                            cfg.token = tokC
                            return pigeon_ngtcp2_transport_init(tPtr, &cfg)
                        }
                    }
                }
            }
        }
        if rc != 0 {
            // pigeon_ngtcp2_transport_init populates last_error on
            // failure and tears down anything it allocated, but does
            // NOT free the struct itself (which is fine — we own it).
            // Read last_error via the C-side accessor — touching the
            // field through `tPtr.pointee.last_error` would force
            // Swift to copy the surrounding ~1.1 MiB transport struct
            // and blow the cooperative-pool thread stack.
            let cstr = pigeon_ngtcp2_transport_last_error(tPtr)
            let msg = cstr.map { String(cString: $0) } ?? ""
            tPtr.deallocate()
            throw Ngtcp2TransportError.initFailed(message: msg)
        }
        self.cTransport = tPtr

        // Read back the (possibly-relay-assigned) instance ID from
        // the transport struct via the C-side accessor. For .connect
        // this echoes the input; for .registerMux this is the value
        // assigned by the relay.
        let idC = pigeon_ngtcp2_transport_instance_id(tPtr)
        self.instanceID = idC.map { String(cString: $0) } ?? ""
    }

    deinit {
        // close() is idempotent; safe to call from deinit even if the
        // caller already invoked it explicitly.
        if !closed {
            pigeon_ngtcp2_transport_close(cTransport)
        }
        // The struct was zero-initialised via memset (NOT via
        // .initialize(to:)) — it holds no Swift-managed values, so
        // we skip .deinitialize() to avoid a phantom struct copy.
        cTransport.deallocate()
    }

    /// Tear down the QUIC connection. Idempotent. After this call the
    /// transport must not be used; any PigeonSession built on it
    /// will see I/O errors on subsequent send/recv calls.
    public func close() {
        if closed { return }
        closed = true
        pigeon_ngtcp2_transport_close(cTransport)
    }

    /// Hand the underlying `pigeon_transport` vtable to the session.
    /// The pointer is valid only while this Ngtcp2Transport is alive,
    /// which the session guarantees by retaining the transport.
    public func withCTransport<R>(_ body: (UnsafePointer<pigeon_transport>) -> R) -> R {
        // `pigeon_transport` is the first field of `pigeon_ngtcp2_transport`
        // (the C-side header guarantees offset 0 for vtable-compatible
        // casting). Using a raw-pointer cast sidesteps Swift's
        // pointee-field accessor, which would otherwise risk copying
        // the ~1.1 MiB parent struct onto the stack.
        let raw = UnsafeRawPointer(cTransport)
        let p = raw.assumingMemoryBound(to: pigeon_transport.self)
        return body(p)
    }

    /// Promote the primary QUIC stream (stream 0, opened during init)
    /// into the multi-channel slot table and return its
    /// `pigeon_stream_handle`. Idempotent: subsequent calls return
    /// the same handle.
    ///
    /// Required for handing the primary off to `pigeon_connect_on_transport`
    /// (the modern client wire). The C-side function returns NULL if
    /// the slot table is full or the primary has not been opened yet;
    /// either case maps to `.noPrimaryHandle`.
    public func primaryHandle() throws -> OpaquePointer {
        guard let h = pigeon_ngtcp2_transport_primary_handle(cTransport) else {
            throw Ngtcp2TransportError.noPrimaryHandle
        }
        // pigeon_stream_handle is forward-declared in the public
        // header, so Swift imports the pointer as OpaquePointer.
        return h
    }

    /// Pointer to the underlying C transport struct, for callers that
    /// need to drop into the C API (e.g. `pigeon_connect_on_transport`).
    /// Use with care: the lifetime of the pointer is tied to `self`.
    public var rawTransport: UnsafeMutablePointer<pigeon_ngtcp2_transport> {
        return cTransport
    }
}

// MARK: - Helpers

private extension Optional where Wrapped == String {
    /// Run `body` with a C string pointer, or NULL if self is nil/empty.
    /// The pointer is valid only for the duration of the call.
    func withCStringOrNull<R>(_ body: (UnsafePointer<CChar>?) -> R) -> R {
        switch self {
        case .some(let s) where !s.isEmpty:
            return s.withCString { body($0) }
        default:
            return body(nil)
        }
    }
}

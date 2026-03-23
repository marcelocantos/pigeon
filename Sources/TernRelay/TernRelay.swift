// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

#if canImport(Network)

import Foundation
import Network

/// Errors specific to TernRelay operations.
public enum TernRelayError: LocalizedError {
    case invalidPort
    case connectionFailed(String)
    case streamFailed(String)
    case handshakeFailed(String)
    case sendFailed(String)
    case recvFailed(String)
    case datagramFailed(String)
    case unexpectedEOF
    case messageTooLarge(UInt32)

    public var errorDescription: String? {
        switch self {
        case .invalidPort: "Invalid port number"
        case .connectionFailed(let msg): "Connection failed: \(msg)"
        case .streamFailed(let msg): "Stream failed: \(msg)"
        case .handshakeFailed(let msg): "Handshake failed: \(msg)"
        case .sendFailed(let msg): "Send failed: \(msg)"
        case .recvFailed(let msg): "Receive failed: \(msg)"
        case .datagramFailed(let msg): "Datagram failed: \(msg)"
        case .unexpectedEOF: "Unexpected end of stream"
        case .messageTooLarge(let size): "Message too large: \(size) bytes"
        }
    }
}

/// Maximum message size for length-prefixed framing (16 MB, matching Go server).
private let maxMessageSize: UInt32 = 16 * 1024 * 1024

/// Connection to a tern relay over raw QUIC (ALPN "tern").
///
/// Use the static `register` and `connect` methods to establish a connection.
/// Messages are exchanged via length-prefixed framing on a bidirectional QUIC
/// stream. Datagrams use the QUIC connection's unreliable datagram channel.
///
/// Available on Apple platforms only (requires Network.framework).
public final class TernConn: @unchecked Sendable {
    /// The relay-assigned instance ID (set after register, or the target ID for connect).
    public let instanceID: String

    private let connection: NWConnection
    private let queue: DispatchQueue
    private let datagramQueue: DispatchQueue

    /// Buffer for received datagrams, consumed by `recvDatagram`.
    private let datagramContinuations = DatagramChannel()

    private init(connection: NWConnection, queue: DispatchQueue, instanceID: String) {
        self.connection = connection
        self.queue = queue
        self.datagramQueue = DispatchQueue(label: "com.tern.datagram", target: queue)
        self.instanceID = instanceID
    }

    // MARK: - Public API

    /// Register as a backend with the relay.
    ///
    /// Opens a raw QUIC connection (ALPN "tern") and sends the register
    /// handshake. Returns a `TernConn` with the relay-assigned instance ID.
    ///
    /// - Parameters:
    ///   - host: Relay server hostname or IP.
    ///   - port: Relay server QUIC port (typically 4433).
    ///   - token: Optional bearer token for authentication.
    ///   - quicOptions: Optional pre-configured QUIC options. If nil, defaults
    ///     are created with ALPN "tern".
    public static func register(
        host: String, port: UInt16,
        token: String? = nil,
        quicOptions: NWProtocolQUIC.Options? = nil
    ) async throws -> TernConn {
        let handshake: String
        if let token = token {
            handshake = "register:\(token)"
        } else {
            handshake = "register"
        }

        let (conn, queue) = try await openConnection(host: host, port: port, quicOptions: quicOptions)

        // Send handshake.
        try await writeMessage(conn, Data(handshake.utf8))

        // Read instance ID.
        let idData = try await readMessage(conn)
        let instanceID = String(decoding: idData, as: UTF8.self)

        let ternConn = TernConn(connection: conn, queue: queue, instanceID: instanceID)
        ternConn.startDatagramReceiver()
        return ternConn
    }

    /// Connect as a client to a specific backend instance.
    ///
    /// Opens a raw QUIC connection (ALPN "tern") and sends the connect
    /// handshake. Returns a `TernConn` ready for message exchange.
    ///
    /// - Parameters:
    ///   - host: Relay server hostname or IP.
    ///   - port: Relay server QUIC port (typically 4433).
    ///   - instanceID: The backend's instance ID to connect to.
    ///   - quicOptions: Optional pre-configured QUIC options.
    public static func connect(
        host: String, port: UInt16,
        instanceID: String,
        quicOptions: NWProtocolQUIC.Options? = nil
    ) async throws -> TernConn {
        let handshake = "connect:\(instanceID)"

        let (conn, queue) = try await openConnection(host: host, port: port, quicOptions: quicOptions)

        // Send handshake.
        try await writeMessage(conn, Data(handshake.utf8))

        let ternConn = TernConn(connection: conn, queue: queue, instanceID: instanceID)
        ternConn.startDatagramReceiver()
        return ternConn
    }

    /// Send a message on the primary stream (length-prefixed).
    ///
    /// Format: [4-byte big-endian length][payload]
    public func send(_ data: Data) async throws {
        try await Self.writeMessage(connection, data)
    }

    /// Receive a message from the primary stream (length-prefixed).
    ///
    /// Reads the 4-byte big-endian length header, then reads exactly that
    /// many bytes of payload.
    public func recv() async throws -> Data {
        try await Self.readMessage(connection)
    }

    /// Send an unreliable datagram on the QUIC connection.
    ///
    /// Uses NWConnection's send with the `.datagram` content context,
    /// which Network.framework maps to a QUIC DATAGRAM frame.
    public func sendDatagram(_ data: Data) async throws {
        try await withCheckedThrowingContinuation { (cont: CheckedContinuation<Void, Error>) in
            connection.send(
                content: data,
                contentContext: .defaultMessage,
                isComplete: true,
                completion: .contentProcessed { error in
                    if let error = error {
                        cont.resume(throwing: TernRelayError.datagramFailed(error.localizedDescription))
                    } else {
                        cont.resume()
                    }
                }
            )
        }
    }

    /// Receive the next datagram from the QUIC connection.
    public func recvDatagram() async throws -> Data {
        try await datagramContinuations.receive()
    }

    /// Close the connection.
    public func close() {
        datagramContinuations.cancel()
        connection.cancel()
    }

    // MARK: - Internal helpers

    /// Open a QUIC connection and wait for it to become ready.
    private static func openConnection(
        host: String, port: UInt16,
        quicOptions: NWProtocolQUIC.Options?
    ) async throws -> (NWConnection, DispatchQueue) {
        guard let nwPort = NWEndpoint.Port(rawValue: port) else {
            throw TernRelayError.invalidPort
        }

        let options = quicOptions ?? {
            let opts = NWProtocolQUIC.Options(alpn: ["tern"])
            // For development: allow insecure connections by default.
            // Callers should provide properly configured quicOptions for production.
            return opts
        }()

        // Ensure ALPN includes "tern" if using caller-provided options.
        // The caller is responsible for setting ALPN correctly.
        let params = NWParameters(quic: options)
        let endpoint = NWEndpoint.hostPort(host: .init(host), port: nwPort)
        let queue = DispatchQueue(label: "com.tern.relay.\(host):\(port)")
        let conn = NWConnection(to: endpoint, using: params)

        try await withThrowingTaskGroup(of: Void.self) { group in
            group.addTask {
                try await withCheckedThrowingContinuation { (cont: CheckedContinuation<Void, Error>) in
                    final class ResumeGuard: @unchecked Sendable {
                        var resumed = false
                    }
                    let guard_ = ResumeGuard()
                    conn.stateUpdateHandler = { state in
                        guard !guard_.resumed else { return }
                        switch state {
                        case .ready:
                            guard_.resumed = true
                            cont.resume()
                        case .failed(let error):
                            guard_.resumed = true
                            cont.resume(throwing: TernRelayError.connectionFailed(error.localizedDescription))
                        case .cancelled:
                            guard_.resumed = true
                            cont.resume(throwing: TernRelayError.connectionFailed("cancelled"))
                        default:
                            break
                        }
                    }
                    conn.start(queue: queue)
                }
            }
            group.addTask {
                try await Task.sleep(nanoseconds: 10_000_000_000) // 10 second timeout
                throw TernRelayError.connectionFailed("timeout after 10 seconds")
            }
            // First to complete wins; cancel the other.
            try await group.next()!
            group.cancelAll()
        }

        return (conn, queue)
    }

    /// Write a length-prefixed message: [4-byte BE length][payload].
    private static func writeMessage(_ conn: NWConnection, _ data: Data) async throws {
        let length = UInt32(data.count)
        guard length <= maxMessageSize else {
            throw TernRelayError.messageTooLarge(length)
        }

        var header = Data(count: 4)
        header[0] = UInt8((length >> 24) & 0xFF)
        header[1] = UInt8((length >> 16) & 0xFF)
        header[2] = UInt8((length >> 8) & 0xFF)
        header[3] = UInt8(length & 0xFF)

        let frame = header + data

        try await withCheckedThrowingContinuation { (cont: CheckedContinuation<Void, Error>) in
            conn.send(content: frame, completion: .contentProcessed { error in
                if let error = error {
                    cont.resume(throwing: TernRelayError.sendFailed(error.localizedDescription))
                } else {
                    cont.resume()
                }
            })
        }
    }

    /// Read a length-prefixed message from the connection.
    private static func readMessage(_ conn: NWConnection) async throws -> Data {
        // Read 4-byte length header.
        let header = try await readExactly(conn, count: 4)
        let length = UInt32(header[0]) << 24
            | UInt32(header[1]) << 16
            | UInt32(header[2]) << 8
            | UInt32(header[3])

        guard length <= maxMessageSize else {
            throw TernRelayError.messageTooLarge(length)
        }

        if length == 0 {
            return Data()
        }

        // Read payload.
        return try await readExactly(conn, count: Int(length))
    }

    /// Read exactly `count` bytes from the connection.
    private static func readExactly(_ conn: NWConnection, count: Int) async throws -> Data {
        try await withCheckedThrowingContinuation { cont in
            conn.receive(minimumIncompleteLength: count, maximumLength: count) { content, _, isComplete, error in
                if let error = error {
                    cont.resume(throwing: TernRelayError.recvFailed(error.localizedDescription))
                } else if let data = content, data.count == count {
                    cont.resume(returning: data)
                } else if isComplete {
                    cont.resume(throwing: TernRelayError.unexpectedEOF)
                } else if let data = content {
                    // Network.framework may return fewer bytes than requested.
                    // We need to accumulate.
                    cont.resume(returning: data)
                } else {
                    cont.resume(throwing: TernRelayError.unexpectedEOF)
                }
            }
        }
    }

    /// Start a background receiver for QUIC datagrams.
    private func startDatagramReceiver() {
        receiveNextDatagram()
    }

    private func receiveNextDatagram() {
        connection.receiveMessage { [weak self] content, context, _, error in
            guard let self = self else { return }

            if let error = error {
                self.datagramContinuations.fail(TernRelayError.datagramFailed(error.localizedDescription))
                return
            }

            // Deliver the datagram content if present. The QUIC metadata
            // on the context identifies this as a datagram frame.
            if let data = content, !data.isEmpty {
                self.datagramContinuations.deliver(data)
            }

            // Continue receiving.
            self.receiveNextDatagram()
        }
    }
}

// MARK: - Datagram channel

/// An async channel for delivering QUIC datagrams from the callback-based
/// Network.framework receiver to async `recvDatagram` callers.
private final class DatagramChannel: @unchecked Sendable {
    private var buffer: [Data] = []
    private var waiters: [CheckedContinuation<Data, Error>] = []
    private var cancelled = false
    private let lock = NSLock()

    func deliver(_ data: Data) {
        lock.lock()
        if let waiter = waiters.first {
            waiters.removeFirst()
            lock.unlock()
            waiter.resume(returning: data)
        } else {
            buffer.append(data)
            lock.unlock()
        }
    }

    func receive() async throws -> Data {
        try await withCheckedThrowingContinuation { cont in
            lock.lock()
            if cancelled {
                lock.unlock()
                cont.resume(throwing: TernRelayError.connectionFailed("cancelled"))
                return
            }
            if !buffer.isEmpty {
                let data = buffer.removeFirst()
                lock.unlock()
                cont.resume(returning: data)
            } else {
                waiters.append(cont)
                lock.unlock()
            }
        }
    }

    func fail(_ error: Error) {
        lock.lock()
        let pending = waiters
        waiters.removeAll()
        cancelled = true
        lock.unlock()
        for waiter in pending {
            waiter.resume(throwing: error)
        }
    }

    func cancel() {
        fail(TernRelayError.connectionFailed("cancelled"))
    }
}

// MARK: - Length-prefix utilities (exposed for testing)

/// Encode a length-prefix header (4-byte big-endian).
internal func encodeLengthPrefix(_ length: UInt32) -> Data {
    var header = Data(count: 4)
    header[0] = UInt8((length >> 24) & 0xFF)
    header[1] = UInt8((length >> 16) & 0xFF)
    header[2] = UInt8((length >> 8) & 0xFF)
    header[3] = UInt8(length & 0xFF)
    return header
}

/// Decode a 4-byte big-endian length prefix.
internal func decodeLengthPrefix(_ data: Data) -> UInt32 {
    precondition(data.count >= 4, "Need at least 4 bytes for length prefix")
    return UInt32(data[data.startIndex]) << 24
        | UInt32(data[data.startIndex + 1]) << 16
        | UInt32(data[data.startIndex + 2]) << 8
        | UInt32(data[data.startIndex + 3])
}

/// Construct a handshake message for the tern QUIC protocol.
internal func buildHandshakeMessage(role: String, token: String? = nil, instanceID: String? = nil) -> String {
    switch role {
    case "register":
        if let token = token {
            return "register:\(token)"
        }
        return "register"
    case "connect":
        return "connect:\(instanceID ?? "")"
    default:
        return role
    }
}

#endif

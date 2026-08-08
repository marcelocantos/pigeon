// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// In-memory paired PairingWireTransport instances for tests and
// in-process demos. Each end's `send` enqueues the payload onto the
// peer's recv queue; each end's `recv` awaits the next payload.
//
// Provides the simplest possible reference implementation of
// PairingWireTransport so the Swift acceptor↔Swift initiator unit test
// can run without any actual relay.

import Foundation

/// A pair of in-process PairingWireTransports that route to each other.
public final class InMemoryPairingTransport: PairingWireTransport, @unchecked Sendable {
    private let lock = NSLock()
    private var inbox: [Data] = []
    private var waiters: [CheckedContinuation<Data, Error>] = []
    private var closed = false
    fileprivate weak var peer: InMemoryPairingTransport?

    public init() {}

    /// Pair two transports so each end's `send` delivers to the other's
    /// `recv` queue. Must be called exactly once for any given pair
    /// before either side calls `send` / `recv`.
    public static func pair() -> (InMemoryPairingTransport, InMemoryPairingTransport) {
        let a = InMemoryPairingTransport()
        let b = InMemoryPairingTransport()
        a.peer = b
        b.peer = a
        return (a, b)
    }

    public func send(_ data: Data) async throws {
        guard let peer = peer else { throw PairingCeremonyError.peerAborted }
        peer.deliver(data)
    }

    public func recv() async throws -> Data {
        try await withCheckedThrowingContinuation { cont in
            lock.lock()
            if closed {
                lock.unlock()
                cont.resume(throwing: PairingCeremonyError.peerAborted)
                return
            }
            if !inbox.isEmpty {
                let next = inbox.removeFirst()
                lock.unlock()
                cont.resume(returning: next)
            } else {
                waiters.append(cont)
                lock.unlock()
            }
        }
    }

    public func close() throws {
        lock.lock()
        let pending = waiters
        waiters.removeAll()
        closed = true
        lock.unlock()
        for w in pending {
            w.resume(throwing: PairingCeremonyError.peerAborted)
        }
    }

    private func deliver(_ data: Data) {
        lock.lock()
        if let w = waiters.first {
            waiters.removeFirst()
            lock.unlock()
            w.resume(returning: data)
        } else {
            inbox.append(data)
            lock.unlock()
        }
    }
}

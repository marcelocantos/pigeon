// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0
//
// BigStackWorker — a serial worker thread with a generously-sized
// stack, used by PigeonSession to host calls into libpigeon.
//
// Why this exists: pigeon_stream_send / pigeon_stream_recv and the
// datagram analogues stack-allocate `uint8_t ct[PIGEON_MAX_MSG + 32]`
// (≈1 MiB) inside dist/pigeon.c. Swift's cooperative thread pool and
// regular DispatchQueue worker threads on Darwin both have stacks
// well under 1 MiB (~64–512 KiB depending on the queue), which the
// big C frame blows out as soon as it's hit. Running these calls on
// a dedicated `Thread` configured with a 4 MiB stack avoids the
// crash with comfortable headroom for nested Swift frames.
//
// The worker also serves as a serialisation point so multiple
// concurrent calls into the same session can't race the
// pigeon_channel sequence counters — the underlying queue is FIFO.

import Foundation

final class BigStackWorker: @unchecked Sendable {
    // 4 MiB. PIGEON_MAX_MSG = 1 MiB, plus AEAD overhead and Swift
    // bridging frames; 4 MiB leaves ample margin.
    private static let stackBytes = 4 * 1024 * 1024

    private let lock = NSLock()
    private var queue: [() -> Void] = []
    private let semaphore = DispatchSemaphore(value: 0)
    private var stopping = false
    private var thread: Thread!

    init() {
        let t = Thread { [weak self] in
            self?.runLoop()
        }
        t.stackSize = BigStackWorker.stackBytes
        t.qualityOfService = .userInitiated
        t.name = "com.pigeon.BigStackWorker"
        self.thread = t
        t.start()
    }

    deinit {
        // Tell the loop to exit and wake it.
        lock.lock()
        stopping = true
        lock.unlock()
        semaphore.signal()
        // Don't join — Thread joining isn't directly supported and
        // the daemon thread will exit on its own once the loop
        // sees `stopping`.
    }

    private func runLoop() {
        while true {
            semaphore.wait()
            lock.lock()
            if stopping && queue.isEmpty {
                lock.unlock()
                return
            }
            let job = queue.isEmpty ? nil : queue.removeFirst()
            lock.unlock()
            job?()
        }
    }

    /// Schedule `body` on the worker thread and await its result.
    /// `body` runs synchronously on a thread with a 4 MiB stack; the
    /// continuation is resumed back in the caller's executor.
    func run<T>(_ body: @escaping () -> T) async -> T {
        await withCheckedContinuation { (cont: CheckedContinuation<T, Never>) in
            lock.lock()
            queue.append { [body] in
                let v = body()
                cont.resume(returning: v)
            }
            lock.unlock()
            semaphore.signal()
        }
    }
}

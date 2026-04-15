// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Manages a Go pigeon relay server as a subprocess for local E2E tests.
// Modeled after android/pigeon/src/test/kotlin/.../GoRelayProcess.kt.

import { execSync, spawn, ChildProcess } from "node:child_process";
import { createSocket } from "node:dgram";
import path from "node:path";

export interface RelayProcess {
  /** The WebTransport (HTTP/3) port the relay is listening on. */
  wtPort: number;
  /** The raw QUIC port the relay is listening on. */
  quicPort: number;
  /**
   * SHA-256 fingerprint (hex) of the self-signed TLS certificate.
   * Pass this to WebTransport's serverCertificateHashes for Chromium.
   */
  certHash: string;
  /** Tear down the relay server. */
  close: () => void;
}

const REPO_ROOT = path.resolve(import.meta.dirname, "../../..");
const RELAY_BIN = "/tmp/pigeon-e2e-server";
const BRIDGE_BIN = "/tmp/pigeon-bridge";
export const CRYPTO_PEER_BIN = "/tmp/pigeon-crypto-peer";

/**
 * Find an available UDP port by briefly binding a socket to port 0.
 */
function findAvailablePort(): Promise<number> {
  return new Promise((resolve, reject) => {
    const sock = createSocket("udp4");
    sock.bind(0, "127.0.0.1", () => {
      const addr = sock.address();
      sock.close(() => resolve(addr.port));
    });
    sock.on("error", reject);
  });
}

/**
 * Build all Go binaries (relay server, pigeon-bridge, and crypto-peer) and
 * start the relay on ephemeral ports. Waits for the "pigeon starting"
 * log line before resolving.
 *
 * @param fixedWtPort - Optional fixed WebTransport port. When omitted, an
 *   ephemeral port is allocated. Use a fixed port when Chromium launch args
 *   like --origin-to-force-quic-on require a known port at config time.
 */
export async function startRelay(fixedWtPort?: number): Promise<RelayProcess> {
  // Build relay, bridge, and crypto-peer binaries.
  execSync(`go build -o ${RELAY_BIN} ./cmd/pigeon/`, {
    cwd: REPO_ROOT,
    stdio: "inherit",
  });
  execSync(`go build -o ${BRIDGE_BIN} ./cmd/pigeon-bridge/`, {
    cwd: REPO_ROOT,
    stdio: "inherit",
  });
  execSync(`go build -o ${CRYPTO_PEER_BIN} ./cmd/crypto-peer`, {
    cwd: REPO_ROOT,
    stdio: "inherit",
  });

  // Allocate ports: use a fixed WT port if provided, else allocate ephemeral.
  const [wtPort, quicPort] = await Promise.all([
    fixedWtPort !== undefined ? Promise.resolve(fixedWtPort) : findAvailablePort(),
    findAvailablePort(),
  ]);

  const proc = spawn(
    RELAY_BIN,
    [
      "--port", String(wtPort),
      "--quic-port", String(quicPort),
      // Use a 14-day cert so Chromium accepts it via serverCertificateHashes.
      "--cert-validity", "14",
    ],
    {
      cwd: REPO_ROOT,
      stdio: ["ignore", "pipe", "pipe"],
    },
  );

  // Wait for "pigeon starting" on stderr (slog writes to stderr).
  // Also capture cert-hash= from the self-signed cert log line.
  let certHash = "";
  await new Promise<void>((resolve, reject) => {
    const timeout = setTimeout(() => {
      proc.kill();
      reject(new Error("Timed out waiting for relay to start (30s)"));
    }, 30000);

    let stderrBuf = "";

    proc.stderr!.on("data", (chunk: Buffer) => {
      const text = chunk.toString();
      stderrBuf += text;
      process.stderr.write(`[pigeon-relay] ${text}`);

      // Parse cert-hash from slog key=value output.
      // slog TextHandler format: key=value or key="value"
      if (!certHash) {
        const m = stderrBuf.match(/cert-hash=([0-9a-f]{64})/);
        if (m) {
          certHash = m[1];
        }
      }

      // "pigeon starting" appears before the goroutine binds. Wait a short
      // grace period then check if the process is still alive.
      if (
        stderrBuf.includes('msg="pigeon starting"') ||
        stderrBuf.includes("msg=pigeon starting")
      ) {
        // Give the goroutine a brief moment to bind and report any error.
        setTimeout(() => {
          if (!proc.killed && proc.exitCode === null) {
            clearTimeout(timeout);
            resolve();
          }
          // If proc already exited/failed, the "exit" handler below fires.
        }, 300);
      }

      // Treat "pigeon failed" as immediate failure (e.g. port in use).
      if (
        stderrBuf.includes('msg="pigeon failed"') ||
        stderrBuf.includes("msg=pigeon failed")
      ) {
        clearTimeout(timeout);
        proc.kill();
        reject(new Error("Relay failed to start (port in use or other error)"));
      }
    });

    proc.on("error", (err) => {
      clearTimeout(timeout);
      reject(err);
    });

    proc.on("exit", (code) => {
      clearTimeout(timeout);
      reject(new Error(`Relay exited prematurely with code ${code}`));
    });
  });

  if (!certHash) {
    proc.kill();
    throw new Error("Relay started but cert-hash not found in log output");
  }

  return {
    wtPort,
    quicPort,
    certHash,
    close() {
      proc.kill();
    },
  };
}

/**
 * Convenience wrapper: start the relay on a specific WebTransport port.
 * Useful when Chromium's --origin-to-force-quic-on must reference a fixed port.
 */
export function startRelayOnPort(wtPort: number): Promise<RelayProcess> {
  return startRelay(wtPort);
}

/**
 * Spawn a pigeon-bridge process with PIGEON_INSECURE=1 and return
 * helpers for length-prefixed message I/O.
 */
export function spawnBridge(
  ...args: string[]
): {
  proc: ChildProcess;
  send: (data: string | Uint8Array) => void;
  recv: () => Promise<string>;
  recvBytes: () => Promise<Uint8Array>;
  close: () => void;
} {
  const proc = spawn(BRIDGE_BIN, args, {
    stdio: ["pipe", "pipe", "pipe"],
    env: { ...process.env, PIGEON_INSECURE: "1" },
  });

  let buffer = Buffer.alloc(0);
  const waiters: ((data: Buffer) => void)[] = [];

  proc.stdout!.on("data", (chunk: Buffer) => {
    buffer = Buffer.concat([buffer, chunk]);
    drain();
  });

  function drain() {
    while (buffer.length >= 4) {
      const len = buffer.readUInt32BE(0);
      if (buffer.length < 4 + len) break;
      const msg = buffer.subarray(4, 4 + len);
      buffer = buffer.subarray(4 + len);
      const waiter = waiters.shift();
      if (waiter) waiter(msg);
    }
  }

  function send(data: string | Uint8Array) {
    const payload = typeof data === "string" ? Buffer.from(data, "utf-8") : Buffer.from(data);
    const hdr = Buffer.alloc(4);
    hdr.writeUInt32BE(payload.length, 0);
    proc.stdin!.write(hdr);
    proc.stdin!.write(payload);
  }

  function recvBytes(): Promise<Uint8Array> {
    return new Promise((resolve, reject) => {
      const timeout = setTimeout(
        () => reject(new Error("recv timeout")),
        10000,
      );
      waiters.push((msg) => {
        clearTimeout(timeout);
        resolve(new Uint8Array(msg));
      });
      drain(); // check buffer in case data already arrived
    });
  }

  function recv(): Promise<string> {
    return recvBytes().then((bytes) => Buffer.from(bytes).toString("utf-8"));
  }

  function close() {
    proc.stdin!.end();
    proc.kill();
  }

  return { proc, send, recv, recvBytes, close };
}

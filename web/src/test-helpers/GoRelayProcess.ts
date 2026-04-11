// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Manages a Go pigeon relay server as a subprocess for local E2E tests.
// Modeled after android/pigeon/src/test/kotlin/.../GoRelayProcess.kt.

import { execSync, spawn, ChildProcess } from "node:child_process";
import { createSocket } from "node:dgram";
import path from "node:path";

export interface RelayProcess {
  /** The raw QUIC port the relay is listening on. */
  quicPort: number;
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
 */
export async function startRelay(): Promise<RelayProcess> {
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

  // Allocate two ephemeral ports: one for WebTransport (HTTP/3) and one for raw QUIC.
  const [wtPort, quicPort] = await Promise.all([
    findAvailablePort(),
    findAvailablePort(),
  ]);

  const proc = spawn(
    RELAY_BIN,
    ["--port", String(wtPort), "--quic-port", String(quicPort)],
    {
      cwd: REPO_ROOT,
      stdio: ["ignore", "pipe", "pipe"],
    },
  );

  // Wait for "pigeon starting" on stderr (slog writes to stderr).
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
      if (
        stderrBuf.includes('msg="pigeon starting"') ||
        stderrBuf.includes("msg=pigeon starting")
      ) {
        clearTimeout(timeout);
        resolve();
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

  return {
    quicPort,
    close() {
      proc.kill();
    },
  };
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

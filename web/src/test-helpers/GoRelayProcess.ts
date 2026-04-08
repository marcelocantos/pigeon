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
 * Build both Go binaries (relay server and pigeon-bridge) and start
 * the relay on an ephemeral port. Waits for the "pigeon starting"
 * log line before resolving.
 */
export async function startRelay(): Promise<RelayProcess> {
  // Build relay and bridge binaries.
  execSync(`go build -o ${RELAY_BIN} ./cmd/pigeon/`, {
    cwd: REPO_ROOT,
    stdio: "inherit",
  });
  execSync(`go build -o ${BRIDGE_BIN} ./cmd/pigeon-bridge/`, {
    cwd: REPO_ROOT,
    stdio: "inherit",
  });

  const quicPort = await findAvailablePort();

  const proc = spawn(
    RELAY_BIN,
    ["--quic-port", String(quicPort)],
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
  send: (data: string) => void;
  recv: () => Promise<string>;
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

  function send(data: string) {
    const payload = Buffer.from(data, "utf-8");
    const hdr = Buffer.alloc(4);
    hdr.writeUInt32BE(payload.length, 0);
    proc.stdin!.write(hdr);
    proc.stdin!.write(payload);
  }

  function recv(): Promise<string> {
    return new Promise((resolve, reject) => {
      const timeout = setTimeout(
        () => reject(new Error("recv timeout")),
        10000,
      );
      waiters.push((msg) => {
        clearTimeout(timeout);
        resolve(msg.toString("utf-8"));
      });
      drain(); // check buffer in case data already arrived
    });
  }

  function close() {
    proc.stdin!.end();
    proc.kill();
  }

  return { proc, send, recv, close };
}

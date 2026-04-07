// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Node.js E2E test for the pigeon relay.
// Uses pigeon-bridge (Go binary) to connect via QUIC and bridges
// messages over stdin/stdout with length-prefixed framing.
//
// Usage:
//   PIGEON_TOKEN=<tok> npx tsx --test src/relay.e2e.ts

import { describe, it, before } from "node:test";
import assert from "node:assert/strict";
import { execSync, spawn, ChildProcess } from "node:child_process";
import path from "node:path";

const TOKEN = process.env.PIGEON_TOKEN || "";
const RELAY_URL = process.env.PIGEON_RELAY_URL || "https://carrier-pigeon.fly.dev:4433";

const REPO_ROOT = path.resolve(import.meta.dirname, "../..");
const BRIDGE_BIN = "/tmp/pigeon-bridge";

// Build the bridge binary once.
before(() => {
  execSync(`go build -o ${BRIDGE_BIN} ./cmd/pigeon-bridge`, {
    cwd: REPO_ROOT,
    stdio: "inherit",
  });
});

// Spawn a bridge process and return helpers for message I/O.
function spawnBridge(
  ...args: string[]
): { proc: ChildProcess; send: (data: string) => void; recv: () => Promise<string>; close: () => void } {
  const proc = spawn(BRIDGE_BIN, args, { stdio: ["pipe", "pipe", "pipe"] });

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
        10000
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

describe(
  "Node.js relay E2E (via pigeon-bridge)",
  { skip: !TOKEN ? "PIGEON_TOKEN not set" : false },
  () => {
    it("register assigns instance ID", async () => {
      const backend = spawnBridge("register", RELAY_URL, TOKEN);
      try {
        const id = await backend.recv();
        assert.ok(id.length > 0, "instance ID should be non-empty");
      } finally {
        backend.close();
      }
    });

    it("bidirectional stream round-trip", async () => {
      // Register backend
      const backend = spawnBridge("register", RELAY_URL, TOKEN);
      const id = await backend.recv();

      // Connect client
      const client = spawnBridge("connect", RELAY_URL, id);

      try {
        // Client → backend
        client.send("hello from node");
        const msg = await backend.recv();
        assert.equal(msg, "hello from node");

        // Backend → client
        backend.send("reply from node");
        const reply = await client.recv();
        assert.equal(reply, "reply from node");
      } finally {
        client.close();
        backend.close();
      }
    });

    it("10 messages in order", async () => {
      const backend = spawnBridge("register", RELAY_URL, TOKEN);
      const id = await backend.recv();
      const client = spawnBridge("connect", RELAY_URL, id);

      try {
        // Send and receive one at a time to avoid buffering issues.
        for (let i = 0; i < 10; i++) {
          client.send(`msg-${i}`);
          const msg = await backend.recv();
          assert.equal(msg, `msg-${i}`);
        }
      } finally {
        client.close();
        backend.close();
      }
    });
  }
);

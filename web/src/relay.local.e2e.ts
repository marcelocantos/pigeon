// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Local E2E test for the pigeon relay.
// Starts a local Go relay server, runs tests via pigeon-bridge, and
// tears down. No PIGEON_TOKEN or remote relay required.
//
// Usage:
//   npx tsx --test src/relay.local.e2e.ts

import { describe, it, before, after } from "node:test";
import assert from "node:assert/strict";
import { startRelay, spawnBridge, type RelayProcess } from "./test-helpers/GoRelayProcess.js";

let relay: RelayProcess;

before(async () => {
  relay = await startRelay();
});

after(() => {
  relay?.close();
});

function relayURL(): string {
  return `https://127.0.0.1:${relay.quicPort}`;
}

describe("Local relay E2E (via pigeon-bridge)", () => {
  it("register assigns instance ID", async () => {
    const backend = spawnBridge("register", relayURL());
    try {
      const id = await backend.recv();
      assert.ok(id.length > 0, "instance ID should be non-empty");
    } finally {
      backend.close();
    }
  });

  it("bidirectional stream round-trip", async () => {
    // Register backend
    const backend = spawnBridge("register", relayURL());
    const id = await backend.recv();

    // Connect client
    const client = spawnBridge("connect", relayURL(), id);

    try {
      // Client -> backend
      client.send("hello from node");
      const msg = await backend.recv();
      assert.equal(msg, "hello from node");

      // Backend -> client
      backend.send("reply from node");
      const reply = await client.recv();
      assert.equal(reply, "reply from node");
    } finally {
      client.close();
      backend.close();
    }
  });

  it("10 messages in order", async () => {
    const backend = spawnBridge("register", relayURL());
    const id = await backend.recv();
    const client = spawnBridge("connect", relayURL(), id);

    try {
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
});

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
import { spawn } from "node:child_process";
import {
  startRelay,
  spawnBridge,
  CRYPTO_PEER_BIN,
  type RelayProcess,
} from "./test-helpers/GoRelayProcess.js";
import { E2EKeyPair, deriveConfirmationCode } from "./crypto.js";

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

  it("cross-language confirmation code", async () => {
    // Start crypto-peer (Go binary) with the local relay URL.
    const peerProc = spawn(CRYPTO_PEER_BIN, [relayURL()], {
      stdio: ["ignore", "pipe", "pipe"],
      env: { ...process.env, PIGEON_INSECURE: "1" },
    });
    let instanceID = "";

    // Read instance ID from crypto-peer's stderr (first line).
    await new Promise<void>((resolve, reject) => {
      const timeout = setTimeout(
        () => reject(new Error("timeout waiting for crypto-peer instance ID")),
        10000,
      );
      let stderrBuf = "";
      let resolved = false;
      peerProc.stderr!.on("data", (chunk: Buffer) => {
        stderrBuf += chunk.toString();
        const nl = stderrBuf.indexOf("\n");
        if (nl !== -1 && !resolved) {
          instanceID = stderrBuf.slice(0, nl).trim();
          resolved = true;
          clearTimeout(timeout);
          resolve();
        }
      });
      peerProc.on("error", (err) => {
        clearTimeout(timeout);
        reject(err);
      });
    });

    assert.ok(instanceID.length > 0, "instance ID from crypto-peer should be non-empty");

    // Connect pigeon-bridge in connect mode to the crypto-peer instance.
    const bridge = spawnBridge("connect", relayURL(), instanceID);

    try {
      // Receive crypto-peer's 32-byte public key.
      const peerPubKey = await bridge.recvBytes();
      assert.equal(peerPubKey.length, 32, "peer public key should be 32 bytes");

      // Generate own X25519 keypair.
      const myKeyPair = await E2EKeyPair.create();

      // Send our 32-byte public key to crypto-peer.
      bridge.send(myKeyPair.publicKeyData);

      // Receive crypto-peer's 6-byte ASCII confirmation code.
      const peerCodeStr = await bridge.recv();
      assert.equal(peerCodeStr.length, 6, "peer confirmation code should be 6 characters");

      // Derive our own confirmation code.
      const myCode = await deriveConfirmationCode(myKeyPair.publicKeyData, peerPubKey);

      // Assert both codes match.
      assert.equal(myCode, peerCodeStr, "TypeScript and Go confirmation codes should match");
    } finally {
      bridge.close();
      peerProc.kill();
    }
  });
});

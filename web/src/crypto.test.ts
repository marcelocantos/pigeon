// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

import { describe, it } from "node:test";
import assert from "node:assert/strict";
import {
  E2EKeyPair,
  E2EChannel,
  deriveKeyFromSecret,
  deriveConfirmationCode,
  createPairingRecord,
  deriveChannelFromRecord,
} from "./crypto.js";

describe("E2EKeyPair", () => {
  it("generates a 32-byte public key", async () => {
    const kp = await E2EKeyPair.create();
    assert.equal(kp.publicKeyData.length, 32);
  });

  it("two key pairs derive the same session key", async () => {
    const alice = await E2EKeyPair.create();
    const bob = await E2EKeyPair.create();
    const info = new TextEncoder().encode("test-session");

    const keyA = await alice.deriveSessionKey(bob.publicKeyData, info);
    const keyB = await bob.deriveSessionKey(alice.publicKeyData, info);

    assert.deepEqual(keyA, keyB);
    assert.equal(keyA.length, 32);
  });
});

describe("E2EChannel", () => {
  it("encrypt/decrypt round trip", async () => {
    const key = globalThis.crypto.getRandomValues(new Uint8Array(32));
    const ch = await E2EChannel.fromSharedKey(key, false);
    // Create a matching channel for the other side
    const ch2 = await E2EChannel.fromSharedKey(key, true);

    const plaintext = new TextEncoder().encode("hello, pigeon!");
    const encrypted = await ch.encrypt(plaintext);
    const decrypted = await ch2.decrypt(encrypted);

    assert.deepEqual(decrypted, plaintext);
  });

  it("bidirectional channel communication", async () => {
    const key = globalThis.crypto.getRandomValues(new Uint8Array(32));
    const client = await E2EChannel.fromSharedKey(key, false);
    const server = await E2EChannel.fromSharedKey(key, true);

    // Client -> Server
    const msg1 = new TextEncoder().encode("from client");
    const enc1 = await client.encrypt(msg1);
    const dec1 = await server.decrypt(enc1);
    assert.deepEqual(dec1, msg1);

    // Server -> Client
    const msg2 = new TextEncoder().encode("from server");
    const enc2 = await server.encrypt(msg2);
    const dec2 = await client.decrypt(enc2);
    assert.deepEqual(dec2, msg2);

    // Multiple messages in same direction
    const msg3 = new TextEncoder().encode("second from client");
    const enc3 = await client.encrypt(msg3);
    const dec3 = await server.decrypt(enc3);
    assert.deepEqual(dec3, msg3);
  });

  it("rejects replay attack (repeated ciphertext)", async () => {
    const key = globalThis.crypto.getRandomValues(new Uint8Array(32));
    const client = await E2EChannel.fromSharedKey(key, false);
    const server = await E2EChannel.fromSharedKey(key, true);

    const plaintext = new TextEncoder().encode("test");
    const encrypted = await client.encrypt(plaintext);

    // First decrypt succeeds
    await server.decrypt(encrypted);

    // Replay should fail (sequence number already consumed)
    await assert.rejects(() => server.decrypt(encrypted), {
      message: "unexpected sequence number",
    });
  });

  it("rejects ciphertext too short", async () => {
    const key = globalThis.crypto.getRandomValues(new Uint8Array(32));
    const ch = await E2EChannel.fromSharedKey(key, true);

    await assert.rejects(() => ch.decrypt(new Uint8Array(4)), {
      message: "ciphertext too short",
    });
  });

  it("datagram mode accepts messages with gaps", async () => {
    const key = globalThis.crypto.getRandomValues(new Uint8Array(32));
    const client = await E2EChannel.fromSharedKey(key, false);
    const server = await E2EChannel.fromSharedKey(key, true);
    server.mode = "datagrams";

    // Encrypt seq 0..5 on sender.
    const cts: Uint8Array[] = [];
    for (let i = 0; i < 6; i++) {
      cts.push(await client.encrypt(new TextEncoder().encode(`msg${i}`)));
    }

    // Receiver gets seq 0, 1, 5 (skipping 2, 3, 4).
    for (const idx of [0, 1, 5]) {
      const pt = await server.decrypt(cts[idx]);
      assert.deepEqual(pt, new TextEncoder().encode(`msg${idx}`));
    }
  });

  it("datagram mode rejects replay", async () => {
    const key = globalThis.crypto.getRandomValues(new Uint8Array(32));
    const client = await E2EChannel.fromSharedKey(key, false);
    const server = await E2EChannel.fromSharedKey(key, true);
    server.mode = "datagrams";

    const ct = await client.encrypt(new TextEncoder().encode("hello"));
    await server.decrypt(ct);

    // Replay same ciphertext — must fail.
    await assert.rejects(() => server.decrypt(ct), {
      message: "sequence number replayed or too old",
    });
  });

  it("datagram mode rejects old sequence numbers", async () => {
    const key = globalThis.crypto.getRandomValues(new Uint8Array(32));
    const client = await E2EChannel.fromSharedKey(key, false);
    const server = await E2EChannel.fromSharedKey(key, true);
    server.mode = "datagrams";

    // Encrypt seq 0..5.
    const cts: Uint8Array[] = [];
    for (let i = 0; i < 6; i++) {
      cts.push(await client.encrypt(new TextEncoder().encode(`msg${i}`)));
    }

    // Receive seq=5 first (jump forward).
    await server.decrypt(cts[5]);

    // seq=3 is now in the past — must be rejected.
    await assert.rejects(() => server.decrypt(cts[3]), {
      message: "sequence number replayed or too old",
    });
  });

  it("strict mode rejects gaps", async () => {
    const key = globalThis.crypto.getRandomValues(new Uint8Array(32));
    const client = await E2EChannel.fromSharedKey(key, false);
    const server = await E2EChannel.fromSharedKey(key, true);

    // Encrypt two packets but skip delivering seq=0.
    await client.encrypt(new TextEncoder().encode("first"));
    const ct1 = await client.encrypt(new TextEncoder().encode("second"));

    // seq=1 without seq=0 must fail in strict mode.
    await assert.rejects(() => server.decrypt(ct1), {
      message: "unexpected sequence number",
    });
  });

  it("direct key construction", async () => {
    const sendKey = globalThis.crypto.getRandomValues(new Uint8Array(32));
    const recvKey = globalThis.crypto.getRandomValues(new Uint8Array(32));

    const sender = await E2EChannel.create(sendKey, recvKey);
    const receiver = await E2EChannel.create(recvKey, sendKey);

    const msg = new TextEncoder().encode("direct keys");
    const enc = await sender.encrypt(msg);
    const dec = await receiver.decrypt(enc);
    assert.deepEqual(dec, msg);
  });
});

describe("deriveKeyFromSecret", () => {
  it("derives a 32-byte key", async () => {
    const secret = globalThis.crypto.getRandomValues(new Uint8Array(32));
    const info = new TextEncoder().encode("test-info");
    const key = await deriveKeyFromSecret(secret, info);
    assert.equal(key.length, 32);
  });

  it("same inputs produce same output", async () => {
    const secret = new Uint8Array(32).fill(0x42);
    const info = new TextEncoder().encode("deterministic");
    const k1 = await deriveKeyFromSecret(secret, info);
    const k2 = await deriveKeyFromSecret(secret, info);
    assert.deepEqual(k1, k2);
  });

  it("different info produces different output", async () => {
    const secret = new Uint8Array(32).fill(0x42);
    const k1 = await deriveKeyFromSecret(
      secret,
      new TextEncoder().encode("info-a"),
    );
    const k2 = await deriveKeyFromSecret(
      secret,
      new TextEncoder().encode("info-b"),
    );
    assert.notDeepEqual(k1, k2);
  });
});

describe("deriveConfirmationCode", () => {
  it("cross-platform vector matches Go", async () => {
    // Fixed 32-byte keys matching TestConfirmationCodeCrossplatformVector in Go.
    const keyA = new Uint8Array(32).fill(0x01);
    const keyB = new Uint8Array(32).fill(0x02);

    const code = await deriveConfirmationCode(keyA, keyB);
    assert.equal(code, "629624");

    // Order-independence.
    const codeBA = await deriveConfirmationCode(keyB, keyA);
    assert.equal(codeBA, "629624");
  });

  it("produces a 6-digit string", async () => {
    const alice = await E2EKeyPair.create();
    const bob = await E2EKeyPair.create();
    const code = await deriveConfirmationCode(
      alice.publicKeyData,
      bob.publicKeyData,
    );
    assert.match(code, /^\d{6}$/);
  });

  it("is order-independent", async () => {
    const alice = await E2EKeyPair.create();
    const bob = await E2EKeyPair.create();
    const code1 = await deriveConfirmationCode(
      alice.publicKeyData,
      bob.publicKeyData,
    );
    const code2 = await deriveConfirmationCode(
      bob.publicKeyData,
      alice.publicKeyData,
    );
    assert.equal(code1, code2);
  });

  it("different key pairs produce different codes (probabilistic)", async () => {
    const a = await E2EKeyPair.create();
    const b = await E2EKeyPair.create();
    const c = await E2EKeyPair.create();
    const code1 = await deriveConfirmationCode(a.publicKeyData, b.publicKeyData);
    const code2 = await deriveConfirmationCode(a.publicKeyData, c.publicKeyData);
    // With 1M possible codes, collision is ~0.0001% — safe to assert inequality
    assert.notEqual(code1, code2);
  });
});

describe("PairingRecord", () => {
  it("round-trips through JSON and derives a working channel", async () => {
    // Simulate a pairing: two key pairs, record one side.
    const serverKP = await E2EKeyPair.create();
    const clientKP = await E2EKeyPair.create();

    const record = await createPairingRecord(
      "server-uuid",
      "https://relay.example.com",
      clientKP,
      serverKP.publicKeyData,
    );

    // Marshal/unmarshal via JSON.
    const json = JSON.stringify(record);
    const restored = JSON.parse(json);

    // Derive channel from restored record.
    const enc = new TextEncoder();
    const ch = await deriveChannelFromRecord(
      restored,
      enc.encode("c2s"),
      enc.encode("s2c"),
    );

    // Derive the same channel from the original keys (server side).
    const serverSendKey = await serverKP.deriveSessionKey(
      clientKP.publicKeyData,
      enc.encode("s2c"),
    );
    const serverRecvKey = await serverKP.deriveSessionKey(
      clientKP.publicKeyData,
      enc.encode("c2s"),
    );
    const serverCh = await E2EChannel.create(serverSendKey, serverRecvKey);

    // Verify they can communicate.
    const ct = await ch.encrypt(enc.encode("hello from restored record"));
    const pt = await serverCh.decrypt(ct);
    assert.deepEqual(pt, enc.encode("hello from restored record"));
  });
});

describe("full ECDH pairing flow", () => {
  it("completes key exchange and establishes encrypted channel", async () => {
    // 1. Both sides generate key pairs
    const alice = await E2EKeyPair.create();
    const bob = await E2EKeyPair.create();

    // 2. Exchange public keys and derive session keys
    const info = new TextEncoder().encode("pairing-session");
    const aliceKey = await alice.deriveSessionKey(bob.publicKeyData, info);
    const bobKey = await bob.deriveSessionKey(alice.publicKeyData, info);
    assert.deepEqual(aliceKey, bobKey);

    // 3. Derive confirmation codes — both should match
    const codeA = await deriveConfirmationCode(
      alice.publicKeyData,
      bob.publicKeyData,
    );
    const codeB = await deriveConfirmationCode(
      bob.publicKeyData,
      alice.publicKeyData,
    );
    assert.equal(codeA, codeB);

    // 4. Create symmetric channels
    const aliceCh = await E2EChannel.fromSharedKey(aliceKey, false);
    const bobCh = await E2EChannel.fromSharedKey(bobKey, true);

    // 5. Bidirectional communication
    const msg1 = new TextEncoder().encode("Hello from Alice");
    const enc1 = await aliceCh.encrypt(msg1);
    const dec1 = await bobCh.decrypt(enc1);
    assert.deepEqual(dec1, msg1);

    const msg2 = new TextEncoder().encode("Hello from Bob");
    const enc2 = await bobCh.encrypt(msg2);
    const dec2 = await aliceCh.decrypt(enc2);
    assert.deepEqual(dec2, msg2);
  });
});

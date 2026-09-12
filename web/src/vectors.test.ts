// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

import { describe, it } from "node:test";
import assert from "node:assert/strict";
import { readFileSync, existsSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";
import {
  E2EChannel,
  deriveConfirmationCode,
  deriveChannelFromRecord,
  type PairingRecord,
} from "./crypto.js";

type Vectors = Record<string, string>;

function loadVectors(): Vectors {
  let dir = dirname(fileURLToPath(import.meta.url));
  for (let i = 0; i < 8; i++) {
    const candidate = join(dir, "c/test/vectors.json");
    if (existsSync(candidate)) {
      return JSON.parse(readFileSync(candidate, "utf8")) as Vectors;
    }
    dir = dirname(dir);
  }
  throw new Error("c/test/vectors.json not found above web/src");
}

function hex(s: string): Uint8Array {
  const out = new Uint8Array(s.length / 2);
  for (let i = 0; i < out.length; i++) {
    out[i] = parseInt(s.slice(i * 2, i * 2 + 2), 16);
  }
  return out;
}

function b64(bytes: Uint8Array): string {
  return Buffer.from(bytes).toString("base64");
}

function record(
  peer: string,
  priv: string,
  pub: string,
  peerPub: string,
): PairingRecord {
  return {
    peerInstanceID: peer,
    relayURL: "test",
    localPrivateKey: b64(hex(priv)),
    localPublicKey: b64(hex(pub)),
    peerPublicKey: b64(hex(peerPub)),
  };
}

describe("cross-language vectors (ENT-003)", () => {
  const v = loadVectors();
  const enc = new TextEncoder();

  it("confirmation code matches Go vector", async () => {
    const code = await deriveConfirmationCode(
      hex(v.alice_public),
      hex(v.bob_public),
    );
    assert.equal(code, v.confirmation_code);
  });

  it("ECDH ciphertext decrypts and re-encrypts", async () => {
    const alice = await deriveChannelFromRecord(
      record("bob", v.alice_private, v.alice_public, v.bob_public),
      enc.encode("alice-to-bob"),
      enc.encode("bob-to-alice"),
    );
    const bob = await deriveChannelFromRecord(
      record("alice", v.bob_private, v.bob_public, v.alice_public),
      enc.encode("bob-to-alice"),
      enc.encode("alice-to-bob"),
    );
    const pt = enc.encode(v.plaintext);
    assert.deepEqual(await bob.decrypt(hex(v.ciphertext)), pt);
    assert.deepEqual(await alice.encrypt(pt), hex(v.ciphertext));
  });

  it("symmetric ciphertexts match Go vector", async () => {
    const master = hex(v.symmetric_master);
    const client = await E2EChannel.fromSharedKey(master, false);
    const server = await E2EChannel.fromSharedKey(master, true);
    const pt = enc.encode(v.symmetric_plaintext);
    assert.deepEqual(
      await server.decrypt(hex(v.symmetric_ciphertext_c2s)),
      pt,
    );
    assert.deepEqual(
      await client.decrypt(hex(v.symmetric_ciphertext_s2c)),
      pt,
    );
    assert.deepEqual(
      await client.encrypt(pt),
      hex(v.symmetric_ciphertext_c2s),
    );
    assert.deepEqual(
      await server.encrypt(pt),
      hex(v.symmetric_ciphertext_s2c),
    );
  });
});

// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0
//
// Unit tests for the wire helpers in relay.ts.
//
// These tests intentionally avoid any live WebTransport: Node's
// WebTransport support is unstable, and the bullseye target stays
// pure-unit so it remains fast and deterministic. Browser/E2E
// validation is out of scope for T31 and lives in the demo app.
//
// The reference byte vectors below are pinned against the equivalent
// C tests in c/test/test_pigeon.c (test_stream_header,
// test_datagram_framing). Cross-language byte parity follows by
// construction: Go, C, and TS all share the same wire shape.

import { describe, it } from "node:test";
import assert from "node:assert/strict";

import { E2EChannel } from "./crypto.js";
import {
  Datagram,
  Session,
  Stream,
  decodeDatagram,
  decodeStreamHeader,
  decodeUvarint,
  encodeDatagram,
  encodeStreamHeader,
  encodeUvarint,
} from "./relay.js";

describe("uvarint", () => {
  // Reference vectors pinned against Go's encoding/binary.PutUvarint
  // and the C uvarint tests in c/test/test_pigeon.c::test_uvarint.
  const cases: Array<[bigint, number[]]> = [
    [0n, [0x00]],
    [1n, [0x01]],
    [127n, [0x7f]],
    [128n, [0x80, 0x01]],
    [300n, [0xac, 0x02]],
    [16384n, [0x80, 0x80, 0x01]],
    [
      1n << 63n,
      [0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x01],
    ],
  ];

  for (const [value, bytes] of cases) {
    it(`encodes ${value} to ${bytes.length} bytes`, () => {
      const got = encodeUvarint(value);
      assert.deepEqual(Array.from(got), bytes);
    });

    it(`decodes ${value} from ${bytes.length} bytes`, () => {
      const [v, n] = decodeUvarint(new Uint8Array(bytes));
      assert.equal(v, value);
      assert.equal(n, bytes.length);
    });
  }

  it("returns 0 bytes consumed on truncated input", () => {
    const [, n] = decodeUvarint(new Uint8Array([0x80]));
    assert.equal(n, 0);
  });

  it("rejects negative values", () => {
    assert.throws(() => encodeUvarint(-1n));
  });
});

describe("stream-header", () => {
  // Reference vectors pinned against the C test
  // c/test/test_pigeon.c::test_stream_header (client side).
  it("encodes the empty name as [0x00]", () => {
    const out = encodeStreamHeader("");
    assert.deepEqual(Array.from(out), [0x00]);
  });

  it("decodes [0x00] as the empty name", () => {
    const [name, n] = decodeStreamHeader(new Uint8Array([0x00]));
    assert.equal(name, "");
    assert.equal(n, 1);
  });

  it("encodes \"control\" as [0x07, 'c', 'o', 'n', 't', 'r', 'o', 'l']", () => {
    const out = encodeStreamHeader("control");
    assert.deepEqual(
      Array.from(out),
      [0x07, 0x63, 0x6f, 0x6e, 0x74, 0x72, 0x6f, 0x6c],
    );
  });

  it("round-trips a multi-byte name", () => {
    const out = encodeStreamHeader("chat");
    assert.deepEqual(Array.from(out), [0x04, 0x63, 0x68, 0x61, 0x74]);
    const [name, n] = decodeStreamHeader(out);
    assert.equal(name, "chat");
    assert.equal(n, out.length);
  });

  it("rejects a truncated name buffer", () => {
    // Header claims length 5 but only 3 bytes follow.
    const buf = new Uint8Array([0x05, 0x61, 0x62, 0x63]);
    assert.throws(() => decodeStreamHeader(buf));
  });
});

describe("datagram framing", () => {
  // Mirrors c/test/test_pigeon.c::test_datagram_framing (client side):
  //   AEAD([varint channel-id][payload]) round-trips with the
  //   correct channel-id and payload.
  async function makeChannel(): Promise<E2EChannel> {
    // Same key on both sides simulates the symmetric two-end channel
    // the C test sets up (datagrams mode, send_ch+recv_ch share a key).
    const key = new Uint8Array(32);
    crypto.getRandomValues(key);
    const ch = await E2EChannel.create(key, key);
    ch.mode = "datagrams";
    return ch;
  }

  it("round-trips channel-id 1 with a small payload", async () => {
    const ch = await makeChannel();
    const payload = new TextEncoder().encode("hello");
    const wire = await encodeDatagram(ch, 1n, payload);
    const [id, decoded] = await decodeDatagram(ch, wire);
    assert.equal(id, 1n);
    assert.deepEqual(decoded, payload);
  });

  it("round-trips a multi-byte varint channel-id (300)", async () => {
    const ch = await makeChannel();
    const payload = new TextEncoder().encode("ping");
    const wire = await encodeDatagram(ch, 300n, payload);
    const [id, decoded] = await decodeDatagram(ch, wire);
    assert.equal(id, 300n);
    assert.deepEqual(decoded, payload);
  });

  it("round-trips an empty payload", async () => {
    const ch = await makeChannel();
    const wire = await encodeDatagram(ch, 7n, new Uint8Array(0));
    const [id, decoded] = await decodeDatagram(ch, wire);
    assert.equal(id, 7n);
    assert.equal(decoded.length, 0);
  });

  it("rejects ciphertext encrypted under a different key", async () => {
    const a = await makeChannel();
    const b = await makeChannel();
    const wire = await encodeDatagram(a, 1n, new TextEncoder().encode("x"));
    await assert.rejects(decodeDatagram(b, wire));
  });
});

// Smoke test for the public type exports — Session/Stream/Datagram are
// directly instantiable below in a real wire-up, but here we just
// verify the constructors are available on the module surface so a
// caller can `new Stream(...)` (e.g. in a mock) without TS surprises.
describe("module surface", () => {
  it("exports Session, Stream, Datagram", () => {
    assert.equal(typeof Session, "function");
    assert.equal(typeof Stream, "function");
    assert.equal(typeof Datagram, "function");
  });
});

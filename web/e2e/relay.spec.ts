// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

import { test, expect } from "@playwright/test";
import { createServer } from "http";
import type { Server } from "http";

// Live relay coordinates. Tests are skipped if PIGEON_TOKEN is not set.
const RELAY_URL = process.env.PIGEON_RELAY_URL || "https://carrier-pigeon.fly.dev";
const TOKEN = process.env.PIGEON_TOKEN || "";

/**
 * Start a minimal HTTP server on localhost that serves a blank page.
 * Chromium treats localhost as a secure context, so WebTransport works.
 */
function startLocalServer(): Promise<{ server: Server; port: number }> {
  return new Promise((resolve) => {
    const server = createServer((_req, res) => {
      res.writeHead(200, { "Content-Type": "text/html" });
      res.end("<html><body></body></html>");
    });
    server.listen(0, "localhost", () => {
      const addr = server.address();
      const port = typeof addr === "object" && addr ? addr.port : 0;
      resolve({ server, port });
    });
  });
}

/**
 * Browser-side relay protocol helpers, injected via new Function().
 * Mirrors the length-prefixed framing from relay.ts.
 */
const RELAY_HELPERS = [
  "async function writeMessage(writer, data) {",
  "  const frame = new Uint8Array(4 + data.length);",
  "  new DataView(frame.buffer).setUint32(0, data.length, false);",
  "  frame.set(data, 4);",
  "  await writer.write(frame);",
  "}",
  "",
  "async function readExact(state, reader, n) {",
  "  const buf = new Uint8Array(n);",
  "  let offset = 0;",
  "  if (state.remainder) {",
  "    const take = Math.min(state.remainder.length, n);",
  "    buf.set(state.remainder.subarray(0, take), 0);",
  "    offset = take;",
  "    state.remainder = take < state.remainder.length",
  "      ? state.remainder.subarray(take) : null;",
  "  }",
  "  while (offset < n) {",
  "    const { value, done } = await reader.read();",
  "    if (done || !value) throw new Error('stream ended');",
  "    const take = Math.min(value.length, n - offset);",
  "    buf.set(value.subarray(0, take), offset);",
  "    offset += take;",
  "    if (take < value.length) {",
  "      state.remainder = value.subarray(take);",
  "    }",
  "  }",
  "  return buf;",
  "}",
  "",
  "async function readMessage(state, reader) {",
  "  const hdr = await readExact(state, reader, 4);",
  "  const len = new DataView(hdr.buffer, hdr.byteOffset, 4).getUint32(0, false);",
  "  return readExact(state, reader, len);",
  "}",
  "",
  "async function openSession(url, handshake) {",
  "  // Prime Alt-Svc cache: browser needs to learn HTTP/3 is available",
  "  // via a regular HTTPS fetch before WebTransport works.",
  "  try { await fetch(url.split('/register')[0].split('/ws/')[0] + '/health'); } catch(e) {}",
  "  await new Promise(r => setTimeout(r, 500));",
  "  const transport = new WebTransport(url);",
  "  await transport.ready;",
  "  const stream = await transport.createBidirectionalStream();",
  "  const writer = stream.writable.getWriter();",
  "  const reader = stream.readable.getReader();",
  "  await writeMessage(writer, new TextEncoder().encode(handshake));",
  "  return { transport, writer, reader };",
  "}",
].join("\n");

test.describe("WebTransport relay E2E", () => {
  let server: Server;
  let pageUrl: string;

  test.beforeAll(async () => {
    const local = await startLocalServer();
    server = local.server;
    pageUrl = "http://localhost:" + local.port;
  });

  test.afterAll(async () => {
    if (server) {
      await new Promise<void>((r) => server.close(() => r()));
    }
  });

  test("register assigns a non-empty instance ID", async ({ page }) => {
    test.skip(!TOKEN, "PIGEON_TOKEN not set");
    await page.goto(pageUrl);

    const instanceID = await page.evaluate(
      async ([relayUrl, token, helpers]: string[]) => {
        const body = [
          helpers,
          "const registerUrl = relayUrl + '/register?token=' + encodeURIComponent(token);",
          "const { transport, writer, reader } = await openSession(registerUrl, 'register');",
          "const state = { remainder: null };",
          "const idBytes = await readMessage(state, reader);",
          "const id = new TextDecoder().decode(idBytes);",
          "transport.close();",
          "return id;",
        ].join("\n");
        const AsyncFunction = Object.getPrototypeOf(async function(){}).constructor;
        const fn = new AsyncFunction("relayUrl", "token", body);
        return await fn(relayUrl, token);
      },
      [RELAY_URL, TOKEN, RELAY_HELPERS],
    );

    expect(instanceID).toBeTruthy();
    expect(typeof instanceID).toBe("string");
    expect((instanceID as string).length).toBeGreaterThan(0);
  });

  test("bidirectional stream round-trip", async ({ page }) => {
    test.skip(!TOKEN, "PIGEON_TOKEN not set");
    await page.goto(pageUrl);

    const result = await page.evaluate(
      async ([relayUrl, token, helpers]: string[]) => {
        const body = [
          helpers,
          "const enc = new TextEncoder();",
          "const dec = new TextDecoder();",
          "",
          "const registerUrl = relayUrl + '/register?token=' + encodeURIComponent(token);",
          "const backend = await openSession(registerUrl, 'register');",
          "const backendState = { remainder: null };",
          "const idBytes = await readMessage(backendState, backend.reader);",
          "const instanceID = dec.decode(idBytes);",
          "",
          "const connectUrl = relayUrl + '/ws/' + encodeURIComponent(instanceID);",
          "const client = await openSession(connectUrl, 'connect');",
          "const clientState = { remainder: null };",
          "",
          "await writeMessage(client.writer, enc.encode('hello from browser'));",
          "const received = await readMessage(backendState, backend.reader);",
          "",
          "await writeMessage(backend.writer, enc.encode('hello from backend'));",
          "const reply = await readMessage(clientState, client.reader);",
          "",
          "client.transport.close();",
          "backend.transport.close();",
          "",
          "return { received: dec.decode(received), reply: dec.decode(reply) };",
        ].join("\n");
        const AsyncFunction = Object.getPrototypeOf(async function(){}).constructor;
        const fn = new AsyncFunction("relayUrl", "token", body);
        return await fn(relayUrl, token);
      },
      [RELAY_URL, TOKEN, RELAY_HELPERS],
    );

    expect(result).toHaveProperty("received", "hello from browser");
    expect(result).toHaveProperty("reply", "hello from backend");
  });

  test("datagram round-trip", async ({ page }) => {
    test.skip(!TOKEN, "PIGEON_TOKEN not set");
    await page.goto(pageUrl);

    const result = await page.evaluate(
      async ([relayUrl, token, helpers]: string[]) => {
        const body = [
          helpers,
          "const enc = new TextEncoder();",
          "const dec = new TextDecoder();",
          "",
          "const registerUrl = relayUrl + '/register?token=' + encodeURIComponent(token);",
          "const backend = await openSession(registerUrl, 'register');",
          "const backendState = { remainder: null };",
          "const idBytes = await readMessage(backendState, backend.reader);",
          "const instanceID = dec.decode(idBytes);",
          "",
          "const connectUrl = relayUrl + '/ws/' + encodeURIComponent(instanceID);",
          "const client = await openSession(connectUrl, 'connect');",
          "",
          "const clientDgWriter = client.transport.datagrams.writable.getWriter();",
          "const backendDgReader = backend.transport.datagrams.readable.getReader();",
          "await clientDgWriter.write(enc.encode('dg-from-browser'));",
          "const { value: dg } = await backendDgReader.read();",
          "",
          "const backendDgWriter = backend.transport.datagrams.writable.getWriter();",
          "const clientDgReader = client.transport.datagrams.readable.getReader();",
          "await backendDgWriter.write(enc.encode('dg-from-backend'));",
          "const { value: dgReply } = await clientDgReader.read();",
          "",
          "client.transport.close();",
          "backend.transport.close();",
          "",
          "return { received: dec.decode(dg), reply: dec.decode(dgReply) };",
        ].join("\n");
        const AsyncFunction = Object.getPrototypeOf(async function(){}).constructor;
        const fn = new AsyncFunction("relayUrl", "token", body);
        return await fn(relayUrl, token);
      },
      [RELAY_URL, TOKEN, RELAY_HELPERS],
    );

    expect(result).toHaveProperty("received", "dg-from-browser");
    expect(result).toHaveProperty("reply", "dg-from-backend");
  });

  test("encrypted stream round-trip", async ({ page }) => {
    test.skip(!TOKEN, "PIGEON_TOKEN not set");
    await page.goto(pageUrl);

    const result = await page.evaluate(
      async ([relayUrl, token, helpers]: string[]) => {
        const body = [
          helpers,
          "const enc = new TextEncoder();",
          "const dec = new TextDecoder();",
          "const subtle = crypto.subtle;",
          "",
          "async function deriveKey(secret, info) {",
          "  const ikm = await subtle.importKey('raw', secret, 'HKDF', false, ['deriveBits']);",
          "  const bits = await subtle.deriveBits(",
          "    { name: 'HKDF', hash: 'SHA-256', salt: new Uint8Array(0), info },",
          "    ikm, 256",
          "  );",
          "  return new Uint8Array(bits);",
          "}",
          "",
          "async function e2eEncrypt(key, seq, plaintext) {",
          "  const seqBytes = new Uint8Array(8);",
          "  new DataView(seqBytes.buffer).setBigUint64(0, BigInt(seq), true);",
          "  const nonce = new Uint8Array(12);",
          "  nonce.set(seqBytes, 0);",
          "  const aesKey = await subtle.importKey('raw', key, 'AES-GCM', false, ['encrypt']);",
          "  const ct = await subtle.encrypt(",
          "    { name: 'AES-GCM', iv: nonce, additionalData: seqBytes },",
          "    aesKey, plaintext",
          "  );",
          "  const result = new Uint8Array(8 + ct.byteLength);",
          "  result.set(seqBytes, 0);",
          "  result.set(new Uint8Array(ct), 8);",
          "  return result;",
          "}",
          "",
          "async function e2eDecrypt(key, data) {",
          "  const seqBytes = data.slice(0, 8);",
          "  const ct = data.slice(8);",
          "  const nonce = new Uint8Array(12);",
          "  nonce.set(seqBytes, 0);",
          "  const aesKey = await subtle.importKey('raw', key, 'AES-GCM', false, ['decrypt']);",
          "  const pt = await subtle.decrypt(",
          "    { name: 'AES-GCM', iv: nonce, additionalData: seqBytes },",
          "    aesKey, ct",
          "  );",
          "  return new Uint8Array(pt);",
          "}",
          "",
          "const sharedKey = new Uint8Array(32);",
          "for (let i = 0; i < 32; i++) sharedKey[i] = (i * 7 + 13) & 0xff;",
          "",
          "const clientSendKey = await deriveKey(sharedKey, enc.encode('client-to-server'));",
          "const clientRecvKey = await deriveKey(sharedKey, enc.encode('server-to-client'));",
          "const backendSendKey = await deriveKey(sharedKey, enc.encode('server-to-client'));",
          "const backendRecvKey = await deriveKey(sharedKey, enc.encode('client-to-server'));",
          "",
          "const registerUrl = relayUrl + '/register?token=' + encodeURIComponent(token);",
          "const backend = await openSession(registerUrl, 'register');",
          "const backendState = { remainder: null };",
          "const idBytes = await readMessage(backendState, backend.reader);",
          "const instanceID = dec.decode(idBytes);",
          "",
          "const connectUrl = relayUrl + '/ws/' + encodeURIComponent(instanceID);",
          "const client = await openSession(connectUrl, 'connect');",
          "const clientState = { remainder: null };",
          "",
          "const plaintext = enc.encode('encrypted hello from browser');",
          "const ciphertext = await e2eEncrypt(clientSendKey, 0, plaintext);",
          "await writeMessage(client.writer, ciphertext);",
          "const receivedCt = await readMessage(backendState, backend.reader);",
          "const receivedPt = await e2eDecrypt(backendRecvKey, receivedCt);",
          "",
          "const replyPt = enc.encode('encrypted reply from backend');",
          "const replyCt = await e2eEncrypt(backendSendKey, 0, replyPt);",
          "await writeMessage(backend.writer, replyCt);",
          "const receivedReplyCt = await readMessage(clientState, client.reader);",
          "const receivedReplyPt = await e2eDecrypt(clientRecvKey, receivedReplyCt);",
          "",
          "client.transport.close();",
          "backend.transport.close();",
          "",
          "return {",
          "  received: dec.decode(receivedPt),",
          "  reply: dec.decode(receivedReplyPt),",
          "};",
        ].join("\n");
        const AsyncFunction = Object.getPrototypeOf(async function(){}).constructor;
        const fn = new AsyncFunction("relayUrl", "token", body);
        return await fn(relayUrl, token);
      },
      [RELAY_URL, TOKEN, RELAY_HELPERS],
    );

    expect(result).toHaveProperty("received", "encrypted hello from browser");
    expect(result).toHaveProperty("reply", "encrypted reply from backend");
  });
});

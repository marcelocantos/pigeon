// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

// Browser-only WebTransport client for the pigeon relay.
//
// Uses the global WebTransport API. The shape mirrors the Go side
// (Session/Stream/Datagram) so callers can think of pigeon as one
// protocol with several language bindings.
//
// Wire (post-T22, client side only — the relay strips/adds the
// 4-byte clientTag prefix on the backend side, so the browser sees no
// tag):
//
//   - Bidi stream framing: each application message is sent as a
//     length-prefixed frame [4-byte big-endian length][payload].
//   - First message on every stream is the unencrypted stream-header
//     [varint name-len][name-bytes]; the empty-name header [0x00]
//     identifies the per-client primary stream. Sub-stream headers
//     carry a non-empty name.
//   - After the header, every subsequent message on a stream is
//     AEAD-encrypted with the Session's crypto.Channel.
//   - Datagrams are AEAD([varint channel-id][payload]); the channel-id
//     is looked up from the pre-agreed name → id map both peers share.
//
// Connect handshake:
//   1. WebTransport `${relayURL}/ws/${instanceID}`.
//   2. Open a bidirectional stream (the primary).
//   3. Write length-prefixed "connect" handshake; read length-prefixed
//      "ok" ack.
//   4. Write length-prefixed [varint 0] empty-name stream header (this
//      tells the backend, via the relay's tag prefix, "new client").
//   5. Write length-prefixed connect-hello JSON
//      {"client_instance_id": <client identity ID>}; read connect-ack
//      JSON {"ok": true}. (Both unencrypted — the relay-tag → backend
//      Session demux happens before the channel exists.)

import {
  E2EChannel,
  PairingRecord,
  deriveChannelFromRecord,
} from "./crypto.js";

// --- Public API types ----------------------------------------------

/**
 * The identity an app passes to connect(). Only `instanceID` is
 * actually consumed by the wire today (it's sent in the connect-hello
 * so the backend can look up the matching PairingRecord). `publicKey`
 * and `deriveSharedSecret` are kept on the interface to mirror the Go
 * side and reserve room for future negotiation that doesn't go through
 * a stored PairingRecord.
 */
export interface Identity {
  publicKey: Uint8Array;
  instanceID: string;
  deriveSharedSecret(
    peerPublicKey: Uint8Array,
    info: Uint8Array,
  ): Promise<Uint8Array>;
}

/** Options for connect(). */
export interface ConnectArgs {
  /** Relay base URL, e.g. "https://relay.example.com:4433". */
  relayURL: string;
  /** Backend's stable instance ID. */
  instanceID: string;
  /** Client identity (provides instanceID and ECDH primitives). */
  identity: Identity;
  /** PairingRecord persisted from the pairing ceremony. */
  record: PairingRecord;
  /** Pre-agreed name → channel-id map for datagrams. ID 0 is reserved. */
  datagrams: Map<string, bigint>;
  /**
   * Optional WebTransport server certificate hashes (for self-signed
   * certs in dev). Each entry: { algorithm: "sha-256", value: ArrayBuffer }.
   */
  serverCertificateHashes?: WebTransportHash[];
}

// --- Wire constants and helpers ------------------------------------

/**
 * Maximum size of a length-prefixed stream message. Mirrors the Go
 * side's MaxMessageSize. Frames larger than this are rejected on both
 * ends — keeping the limit tight bounds peer memory usage and matches
 * what the relay enforces.
 */
const MAX_MESSAGE_SIZE = 1 << 20; // 1 MiB

/** Encode a non-negative bigint as Go-style uvarint into a Uint8Array. */
export function encodeUvarint(value: bigint): Uint8Array {
  if (value < 0n) throw new Error("uvarint: negative value");
  const out: number[] = [];
  let v = value;
  while (v >= 0x80n) {
    out.push(Number((v & 0x7fn) | 0x80n));
    v >>= 7n;
  }
  out.push(Number(v));
  return new Uint8Array(out);
}

/**
 * Decode a uvarint at offset 0. Returns [value, bytesConsumed].
 * Throws on overflow (>10 bytes) and returns bytesConsumed=0 on
 * truncated input (need-more-bytes).
 */
export function decodeUvarint(buf: Uint8Array): [bigint, number] {
  let v = 0n;
  let shift = 0n;
  for (let i = 0; i < buf.length; i++) {
    if (i >= 10) throw new Error("uvarint: too many bytes");
    const b = buf[i];
    if ((b & 0x80) === 0) {
      v |= BigInt(b) << shift;
      return [v, i + 1];
    }
    v |= BigInt(b & 0x7f) << shift;
    shift += 7n;
  }
  return [0n, 0]; // truncated
}

/**
 * Encode a client-side stream-header: [varint name-len][name-bytes].
 * Empty name → just [0x00] (the per-client primary stream marker).
 * Backend headers (with a 4-byte tag prefix) are not produced here —
 * the browser only ever speaks the client side of the wire.
 */
export function encodeStreamHeader(name: string): Uint8Array {
  const nameBytes = new TextEncoder().encode(name);
  const lenBytes = encodeUvarint(BigInt(nameBytes.length));
  const out = new Uint8Array(lenBytes.length + nameBytes.length);
  out.set(lenBytes, 0);
  out.set(nameBytes, lenBytes.length);
  return out;
}

/** Decode a client-side stream-header. Returns [name, bytesConsumed]. */
export function decodeStreamHeader(buf: Uint8Array): [string, number] {
  const [nameLen, n] = decodeUvarint(buf);
  if (n === 0) throw new Error("stream header: truncated varint");
  const total = n + Number(nameLen);
  if (buf.length < total) throw new Error("stream header: truncated name");
  const name = new TextDecoder().decode(buf.subarray(n, total));
  return [name, total];
}

/**
 * Encode a client-side datagram payload: AEAD([varint channel-id][payload]).
 * Returns the ciphertext to put on the wire.
 */
export async function encodeDatagram(
  channel: E2EChannel,
  channelId: bigint,
  payload: Uint8Array,
): Promise<Uint8Array> {
  const idBytes = encodeUvarint(channelId);
  const plain = new Uint8Array(idBytes.length + payload.length);
  plain.set(idBytes, 0);
  plain.set(payload, idBytes.length);
  return channel.encrypt(plain);
}

/**
 * Decode a client-side datagram from the wire.
 * Returns [channelId, payload].
 */
export async function decodeDatagram(
  channel: E2EChannel,
  wire: Uint8Array,
): Promise<[bigint, Uint8Array]> {
  const plain = await channel.decrypt(wire);
  const [id, n] = decodeUvarint(plain);
  if (n === 0) throw new Error("datagram: truncated channel-id varint");
  return [id, plain.subarray(n)];
}

// --- Length-prefixed frame I/O on a single stream ------------------

/**
 * Per-stream reader that owns a ReadableStreamDefaultReader and a
 * leftover-bytes buffer. Length-prefixed frames don't necessarily
 * align with WebTransport read chunks, so the reader has to be able
 * to carry remainder bytes between calls.
 */
class FramedReader {
  private reader: ReadableStreamDefaultReader<Uint8Array>;
  private remainder: Uint8Array | null = null;

  constructor(reader: ReadableStreamDefaultReader<Uint8Array>) {
    this.reader = reader;
  }

  async readMessage(): Promise<Uint8Array> {
    const hdr = await this.readExact(4);
    const length = new DataView(
      hdr.buffer,
      hdr.byteOffset,
      hdr.byteLength,
    ).getUint32(0, false);
    if (length > MAX_MESSAGE_SIZE) {
      throw new Error(`message too large: ${length} > ${MAX_MESSAGE_SIZE}`);
    }
    return this.readExact(length);
  }

  private async readExact(n: number): Promise<Uint8Array> {
    const buf = new Uint8Array(n);
    let offset = 0;

    if (this.remainder !== null) {
      const rem = this.remainder;
      const take = Math.min(rem.length, n);
      buf.set(rem.subarray(0, take), 0);
      offset = take;
      this.remainder = take < rem.length ? rem.subarray(take) : null;
    }

    while (offset < n) {
      const { value, done } = await this.reader.read();
      if (done || !value) {
        throw new Error("stream ended before expected bytes were read");
      }
      const take = Math.min(value.length, n - offset);
      buf.set(value.subarray(0, take), offset);
      offset += take;
      if (take < value.length) {
        this.remainder = value.subarray(take);
      }
    }
    return buf;
  }
}

async function writeFrame(
  writer: WritableStreamDefaultWriter<Uint8Array>,
  data: Uint8Array,
): Promise<void> {
  if (data.length > MAX_MESSAGE_SIZE) {
    throw new Error(
      `message too large: ${data.length} > ${MAX_MESSAGE_SIZE}`,
    );
  }
  const frame = new Uint8Array(4 + data.length);
  new DataView(frame.buffer).setUint32(0, data.length, false);
  frame.set(data, 4);
  await writer.write(frame);
}

// --- Stream / Datagram / Session -----------------------------------

/**
 * A reliable, ordered, message-framed channel encrypted with the
 * Session's AEAD channel. Created via Session.openStream(name) or
 * Session.acceptStream(name).
 */
export class Stream {
  readonly name: string;
  private reader: FramedReader;
  private writer: WritableStreamDefaultWriter<Uint8Array>;
  private channel: E2EChannel;
  private closed = false;

  /** @internal Use Session.openStream / Session.acceptStream. */
  constructor(
    name: string,
    reader: FramedReader,
    writer: WritableStreamDefaultWriter<Uint8Array>,
    channel: E2EChannel,
  ) {
    this.name = name;
    this.reader = reader;
    this.writer = writer;
    this.channel = channel;
  }

  /** Encrypt and send one message on the stream. */
  async send(msg: Uint8Array): Promise<void> {
    if (this.closed) throw new Error("stream closed");
    const ct = await this.channel.encrypt(msg);
    await writeFrame(this.writer, ct);
  }

  /** Read the next message from the stream and decrypt it. */
  async recv(): Promise<Uint8Array> {
    if (this.closed) throw new Error("stream closed");
    const ct = await this.reader.readMessage();
    return this.channel.decrypt(ct);
  }

  async close(): Promise<void> {
    if (this.closed) return;
    this.closed = true;
    try {
      await this.writer.close();
    } catch {
      // Closing a writer that's already errored is fine.
    }
  }
}

/**
 * A pre-agreed unreliable, unordered channel keyed by a varint
 * channel-id. Both peers must declare the same name → id map.
 */
export class Datagram {
  readonly name: string;
  private id: bigint;
  private session: Session;
  /** Inbound queue + a single waiter (FIFO). The Session pump pushes here. */
  private rxQueue: Uint8Array[] = [];
  private rxWaiters: Array<{
    resolve: (v: Uint8Array) => void;
    reject: (err: unknown) => void;
  }> = [];

  /** @internal Use Session.datagram(name). */
  constructor(name: string, id: bigint, session: Session) {
    this.name = name;
    this.id = id;
    this.session = session;
  }

  /** Encrypt and send one datagram. */
  async send(payload: Uint8Array): Promise<void> {
    const wire = await encodeDatagram(
      this.session.cryptoChannel,
      this.id,
      payload,
    );
    const writer = this.session.datagramWriter;
    await writer.write(wire);
  }

  /** Block until the next datagram on this channel arrives. */
  recv(): Promise<Uint8Array> {
    if (this.rxQueue.length > 0) {
      return Promise.resolve(this.rxQueue.shift()!);
    }
    return new Promise<Uint8Array>((resolve, reject) => {
      this.rxWaiters.push({ resolve, reject });
    });
  }

  /** @internal Called by the Session datagram pump. */
  deliver(payload: Uint8Array): void {
    const w = this.rxWaiters.shift();
    if (w) {
      w.resolve(payload);
    } else {
      this.rxQueue.push(payload);
    }
  }

  /** @internal Called by Session.close() to unblock pending recv()s. */
  failPending(err: unknown): void {
    const waiters = this.rxWaiters;
    this.rxWaiters = [];
    for (const w of waiters) w.reject(err);
  }
}

/**
 * An end-to-end encrypted session to a paired backend through the
 * pigeon WebTransport relay. Carries one or more named, AEAD-encrypted
 * streams (openStream/acceptStream) and pre-agreed datagram channels
 * (datagram).
 */
export class Session {
  readonly peerID: string;
  /** @internal */
  readonly cryptoChannel: E2EChannel;

  private transport: WebTransport;
  /** Per-client primary stream — kept open for session lifetime. */
  private primaryWriter: WritableStreamDefaultWriter<Uint8Array>;
  private primaryReader: FramedReader;

  /** @internal — Datagrams writes through this. */
  readonly datagramWriter: WritableStreamDefaultWriter<Uint8Array>;
  private datagramReader: ReadableStreamDefaultReader<Uint8Array>;

  /** datagram-id → Datagram (both directions of the name → id map). */
  private datagramsById: Map<bigint, Datagram>;
  /** name → Datagram (the user's index). */
  private datagramsByName: Map<string, Datagram>;

  /** Stream rendezvous — same shape as Go's pendingOpens / bufferedStreams. */
  private pendingAccepts: Map<string, Array<(s: Stream) => void>> = new Map();
  private bufferedStreams: Map<string, Stream[]> = new Map();

  private closed = false;
  private datagramPumpRunning = true;
  private acceptPumpRunning = true;

  /** @internal Use connect(). */
  constructor(
    transport: WebTransport,
    primaryWriter: WritableStreamDefaultWriter<Uint8Array>,
    primaryReader: FramedReader,
    channel: E2EChannel,
    peerID: string,
    datagrams: Map<string, bigint>,
  ) {
    this.transport = transport;
    this.primaryWriter = primaryWriter;
    this.primaryReader = primaryReader;
    this.cryptoChannel = channel;
    this.peerID = peerID;
    this.datagramWriter = transport.datagrams.writable.getWriter();
    this.datagramReader = transport.datagrams.readable.getReader();

    this.datagramsById = new Map();
    this.datagramsByName = new Map();
    for (const [name, id] of datagrams) {
      if (id === 0n) {
        throw new Error(`datagram channel "${name}": id 0 is reserved`);
      }
      if (this.datagramsById.has(id)) {
        throw new Error(`datagram channel id ${id} used twice`);
      }
      const dg = new Datagram(name, id, this);
      this.datagramsByName.set(name, dg);
      this.datagramsById.set(id, dg);
    }

    void this.datagramPump();
    void this.acceptPump();
  }

  /**
   * Open a fresh, reliable, ordered stream identified by `name`. Both
   * peers must call openStream/acceptStream with the same name; the
   * backend's matching call unblocks when the named stream arrives.
   */
  async openStream(name: string): Promise<Stream> {
    if (name === "") throw new Error("openStream: name must be non-empty");
    if (this.closed) throw new Error("session closed");

    const stream = await this.transport.createBidirectionalStream();
    const writer = stream.writable.getWriter();
    const reader = new FramedReader(stream.readable.getReader());

    // First message on the stream: unencrypted client stream-header.
    await writeFrame(writer, encodeStreamHeader(name));

    return new Stream(name, reader, writer, this.cryptoChannel);
  }

  /**
   * Block until the peer opens a stream with the given name. Streams
   * that arrive before this call are buffered and delivered in FIFO
   * order when their name is later requested.
   */
  acceptStream(name: string): Promise<Stream> {
    if (name === "") {
      return Promise.reject(new Error("acceptStream: name must be non-empty"));
    }
    if (this.closed) {
      return Promise.reject(new Error("session closed"));
    }
    const buf = this.bufferedStreams.get(name);
    if (buf && buf.length > 0) {
      const s = buf.shift()!;
      if (buf.length === 0) this.bufferedStreams.delete(name);
      return Promise.resolve(s);
    }
    return new Promise<Stream>((resolve) => {
      const q = this.pendingAccepts.get(name) ?? [];
      q.push(resolve);
      this.pendingAccepts.set(name, q);
    });
  }

  /** Look up a pre-declared datagram channel by name. */
  datagram(name: string): Datagram {
    const dg = this.datagramsByName.get(name);
    if (!dg) {
      throw new Error(
        `pigeon: datagram channel "${name}" not declared in connect args`,
      );
    }
    return dg;
  }

  /** Tear down the Session and the underlying WebTransport. */
  async close(): Promise<void> {
    if (this.closed) return;
    this.closed = true;
    this.datagramPumpRunning = false;
    this.acceptPumpRunning = false;

    // Unblock any pending datagram and stream waiters.
    const closedErr = new Error("session closed");
    for (const dg of this.datagramsById.values()) dg.failPending(closedErr);
    for (const queue of this.pendingAccepts.values()) {
      for (const _resolve of queue) {
        // We deliberately do not reject acceptStream waiters — the spec
        // says it returns a Stream. We just leave them pending; the
        // caller's surrounding logic (e.g. an AbortController) is
        // expected to wrap acceptStream if cancellation is needed.
        void _resolve;
      }
    }
    this.pendingAccepts.clear();

    try { await this.primaryWriter.close(); } catch { /* ignore */ }
    try { this.datagramWriter.releaseLock(); } catch { /* ignore */ }
    try { this.datagramReader.releaseLock(); } catch { /* ignore */ }
    try { this.transport.close(); } catch { /* ignore */ }
  }

  /** Pump incoming datagrams: decrypt, dispatch by channel-id. */
  private async datagramPump(): Promise<void> {
    while (this.datagramPumpRunning) {
      let value: Uint8Array | undefined;
      try {
        const r = await this.datagramReader.read();
        if (r.done) return;
        value = r.value;
      } catch {
        return;
      }
      if (!value) continue;
      try {
        const [id, payload] = await decodeDatagram(this.cryptoChannel, value);
        const dg = this.datagramsById.get(id);
        if (dg) dg.deliver(payload);
        // Unknown channel-id: drop silently (matches Go session.go).
      } catch {
        // Drop bad datagrams (matches Go session.go: AEAD failures
        // and varint failures are logged-and-dropped, never fatal).
      }
    }
  }

  /**
   * Pump incoming sub-streams the backend opens. Reads each new
   * stream's first-message header, then dispatches by name. In
   * practice the backend rarely opens streams against the client in
   * v1, but the API stays symmetric with Go's clientAcceptLoop.
   */
  private async acceptPump(): Promise<void> {
    const incoming = this.transport.incomingBidirectionalStreams.getReader();
    try {
      while (this.acceptPumpRunning) {
        let stream: WebTransportBidirectionalStream | undefined;
        try {
          const r = await incoming.read();
          if (r.done) return;
          stream = r.value;
        } catch {
          return;
        }
        if (!stream) continue;

        const writer = stream.writable.getWriter();
        const reader = new FramedReader(stream.readable.getReader());
        let name: string;
        try {
          const hdr = await reader.readMessage();
          [name] = decodeStreamHeader(hdr);
        } catch {
          try { await writer.close(); } catch { /* ignore */ }
          continue;
        }
        const s = new Stream(name, reader, writer, this.cryptoChannel);
        this.deliverIncomingStream(name, s);
      }
    } finally {
      try { incoming.releaseLock(); } catch { /* ignore */ }
    }
  }

  private deliverIncomingStream(name: string, s: Stream): void {
    const q = this.pendingAccepts.get(name);
    if (q && q.length > 0) {
      const resolve = q.shift()!;
      if (q.length === 0) this.pendingAccepts.delete(name);
      resolve(s);
      return;
    }
    const buf = this.bufferedStreams.get(name) ?? [];
    buf.push(s);
    this.bufferedStreams.set(name, buf);
  }
}

// --- connect -------------------------------------------------------

/** Wire shape sent on the primary just after the relay handshake. */
interface ConnectHello {
  client_instance_id: string;
}

interface ConnectAck {
  ok: boolean;
  reason?: string;
}

/**
 * Connect to a paired backend through the pigeon WebTransport relay.
 *
 * Browser-only — uses the global WebTransport API. Returns a Session
 * ready for openStream/datagram traffic. The returned Session's
 * cryptoChannel is derived from the supplied PairingRecord (client →
 * backend / backend → client direction).
 */
export async function connect(args: ConnectArgs): Promise<Session> {
  const baseURL = args.relayURL.replace(/^https?:/, "https:").replace(/\/$/, "");
  const url = `${baseURL}/ws/${encodeURIComponent(args.instanceID)}`;

  const wtOpts: WebTransportOptions = {};
  if (args.serverCertificateHashes) {
    wtOpts.serverCertificateHashes = args.serverCertificateHashes;
  }

  const transport = new WebTransport(url, wtOpts);
  await transport.ready;

  // Open the primary stream, run the relay-level handshake
  // ("connect"/"ok") and the application-level handshake (empty-name
  // stream-header + connect-hello/ack). All four exchanges are sent
  // length-prefixed via writeFrame; the relay only inspects the first
  // ("connect" → "ok"), then transparently forwards the rest.
  const stream = await transport.createBidirectionalStream();
  const writer = stream.writable.getWriter();
  const reader = new FramedReader(stream.readable.getReader());

  // Relay handshake.
  await writeFrame(writer, new TextEncoder().encode("connect"));
  const ack = await reader.readMessage();
  const ackStr = new TextDecoder().decode(ack);
  if (ackStr !== "ok") {
    transport.close();
    throw new Error(`relay handshake: expected "ok", got ${JSON.stringify(ackStr)}`);
  }

  // Application handshake — empty-name stream-header marks "this is the
  // client primary" so the backend's listener can demux on the relay's
  // tag prefix.
  await writeFrame(writer, encodeStreamHeader(""));
  const hello: ConnectHello = { client_instance_id: args.identity.instanceID };
  await writeFrame(writer, new TextEncoder().encode(JSON.stringify(hello)));
  const helloAckRaw = await reader.readMessage();
  const helloAck: ConnectAck = JSON.parse(new TextDecoder().decode(helloAckRaw));
  if (!helloAck.ok) {
    transport.close();
    throw new Error(`backend rejected connection: ${helloAck.reason ?? "(no reason)"}`);
  }

  // Derive the AEAD channel from the PairingRecord. Same info strings
  // as Go's api.go Connect (client side: send=client->backend,
  // recv=backend->client).
  const channel = await deriveChannelFromRecord(
    args.record,
    new TextEncoder().encode("client->backend"),
    new TextEncoder().encode("backend->client"),
  );

  return new Session(transport, writer, reader, channel, args.instanceID, args.datagrams);
}

/**
 * Wake a Fly.io relay that may be auto-stopped. Sends an HTTPS
 * request to /health, which triggers Fly's proxy to start the
 * machine. No-op if the relay is already running. Best-effort —
 * errors are silently ignored. Useful to call before connect().
 */
export async function wakeRelay(relayURL: string): Promise<void> {
  const healthURL = relayURL.replace(/\/$/, "") + "/health";
  try {
    await fetch(healthURL);
  } catch {
    // Best-effort.
  }
}

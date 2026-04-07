// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

/** Options for relay connections. */
export interface ConnectOptions {
  /** Bearer token for authentication on /register. */
  token?: string;
  /**
   * Optional server certificate hashes for development (self-signed certs).
   * Each entry must have { algorithm: "sha-256", value: ArrayBuffer }.
   */
  serverCertificateHashes?: WebTransportHash[];
}

/** Maximum relay frame size (must match server's maxWTMessageSize). */
const maxMessageSize = 1 << 20; // 1 MiB

/**
 * Write a length-prefixed message to a WritableStream.
 * Format: [4-byte big-endian length][payload]
 */
async function writeMessage(
  writer: WritableStreamDefaultWriter<Uint8Array>,
  data: Uint8Array,
): Promise<void> {
  if (data.length > maxMessageSize) {
    throw new Error(`message too large: ${data.length} > ${maxMessageSize}`);
  }
  const frame = new Uint8Array(4 + data.length);
  const view = new DataView(frame.buffer);
  view.setUint32(0, data.length, false); // big-endian
  frame.set(data, 4);
  await writer.write(frame);
}

/**
 * Read a length-prefixed message from a ReadableStream.
 * Returns the payload (without the 4-byte header).
 */
async function readMessage(
  conn: Conn,
  reader: ReadableStreamBYOBReader | ReadableStreamDefaultReader<Uint8Array>,
): Promise<Uint8Array> {
  const hdr = await readExact(conn, reader, 4);
  const length = new DataView(
    hdr.buffer,
    hdr.byteOffset,
    hdr.byteLength,
  ).getUint32(0, false);
  if (length > maxMessageSize) {
    throw new Error(`message too large: ${length} > ${maxMessageSize}`);
  }
  return readExact(conn, reader, length);
}

/**
 * Read exactly `n` bytes from a stream, assembling from multiple chunks
 * if necessary. Leftover bytes from oversized chunks are preserved in
 * `conn.remainder` for subsequent reads.
 */
async function readExact(
  conn: Conn,
  reader: ReadableStreamBYOBReader | ReadableStreamDefaultReader<Uint8Array>,
  n: number,
): Promise<Uint8Array> {
  const buf = new Uint8Array(n);
  let offset = 0;

  // Consume any leftover bytes from a previous read.
  if (conn["remainder"] !== null) {
    const rem = conn["remainder"] as Uint8Array;
    const take = Math.min(rem.length, n);
    buf.set(rem.subarray(0, take), 0);
    offset = take;
    conn["remainder"] = take < rem.length ? rem.subarray(take) : null;
  }

  while (offset < n) {
    const { value, done } = await reader.read() as ReadableStreamReadResult<Uint8Array>;
    if (done || !value) {
      throw new Error("stream ended before expected bytes were read");
    }
    const take = Math.min(value.length, n - offset);
    buf.set(value.subarray(0, take), offset);
    offset += take;
    if (take < value.length) {
      conn["remainder"] = value.subarray(take);
    }
  }
  return buf;
}

// TODO(T12): Add setChannel/setDatagramChannel for automatic E2E encryption.
// Currently callers must encrypt/decrypt manually using E2EChannel.

/**
 * A connection to a peer through the pigeon WebTransport relay.
 */
export class Conn {
  /** The relay-assigned instance ID. */
  readonly instanceID: string;

  private transport: WebTransport;
  private writer: WritableStreamDefaultWriter<Uint8Array>;
  private reader: ReadableStreamDefaultReader<Uint8Array>;
  private remainder: Uint8Array | null = null;
  private datagramWriter: WritableStreamDefaultWriter<Uint8Array> | null = null;
  private datagramReader: ReadableStreamDefaultReader<Uint8Array> | null = null;
  private closed = false;

  /** @internal Use register() or connect() instead. */
  constructor(
    transport: WebTransport,
    writer: WritableStreamDefaultWriter<Uint8Array>,
    reader: ReadableStreamDefaultReader<Uint8Array>,
    instanceID: string,
  ) {
    this.transport = transport;
    this.writer = writer;
    this.reader = reader;
    this.instanceID = instanceID;
  }

  /** Send a message to the peer on the reliable stream. */
  async send(data: Uint8Array): Promise<void> {
    if (this.closed) {
      throw new Error("connection is closed");
    }
    await writeMessage(this.writer, data);
  }

  /** Receive the next message from the peer on the reliable stream. */
  async recv(): Promise<Uint8Array> {
    if (this.closed) {
      throw new Error("connection is closed");
    }
    return readMessage(this, this.reader);
  }

  /** Send an unreliable datagram to the peer. */
  sendDatagram(data: Uint8Array): void {
    if (this.closed) {
      throw new Error("connection is closed");
    }
    if (!this.datagramWriter) {
      this.datagramWriter = this.transport.datagrams.writable.getWriter();
    }
    this.datagramWriter.write(data);
  }

  /** Receive the next unreliable datagram from the peer. */
  async recvDatagram(): Promise<Uint8Array> {
    if (this.closed) {
      throw new Error("connection is closed");
    }
    if (!this.datagramReader) {
      this.datagramReader = this.transport.datagrams.readable.getReader();
    }
    const { value, done } = await this.datagramReader.read();
    if (done || !value) {
      throw new Error("datagram stream ended");
    }
    return value;
  }

  /** Close the connection. */
  close(): void {
    if (!this.closed) {
      this.closed = true;
      this.writer.close().catch(() => {});
      if (this.datagramWriter) {
        this.datagramWriter.close().catch(() => {});
      }
      this.transport.close();
    }
  }
}

/**
 * Open a WebTransport session and create a bidirectional stream with
 * the length-prefixed handshake.
 */
async function openSession(
  url: string,
  handshake: string,
  opts?: ConnectOptions,
): Promise<{
  transport: WebTransport;
  writer: WritableStreamDefaultWriter<Uint8Array>;
  reader: ReadableStreamDefaultReader<Uint8Array>;
}> {
  const wtOpts: WebTransportOptions = {};
  if (opts?.serverCertificateHashes) {
    wtOpts.serverCertificateHashes = opts.serverCertificateHashes;
  }

  const transport = new WebTransport(url, wtOpts);
  await transport.ready;

  const stream = await transport.createBidirectionalStream();
  const writer = stream.writable.getWriter();
  const reader = stream.readable.getReader();

  // Send the handshake message (length-prefixed).
  await writeMessage(writer, new TextEncoder().encode(handshake));

  return { transport, writer, reader };
}

/**
 * Register as a backend with the relay. Returns a Conn whose instanceID
 * is the relay-assigned instance ID.
 */
export async function register(
  url: string,
  opts?: ConnectOptions,
): Promise<Conn> {
  // Wake the relay if auto-stopped (best-effort).
  await wakeRelay(url);

  let registerURL = url.replace(/\/$/, "") + "/register";

  // WebTransport supports headers via URL params or protocol-level auth.
  // Pass token as a query parameter since WebTransport doesn't support
  // arbitrary request headers in all browsers.
  if (opts?.token) {
    const sep = registerURL.includes("?") ? "&" : "?";
    registerURL += sep + "token=" + encodeURIComponent(opts.token);
  }

  const { transport, writer, reader } = await openSession(
    registerURL,
    "register",
    opts,
  );

  // Create the Conn first so readMessage can use its remainder buffer.
  const conn = new Conn(transport, writer, reader, "");
  const idBytes = await readMessage(conn, reader);
  const instanceID = new TextDecoder().decode(idBytes);

  // Patch the instance ID onto the connection (bypass readonly for init).
  (conn as unknown as { instanceID: string }).instanceID = instanceID;
  return conn;
}

/**
 * Connect to a backend instance through the relay.
 */
export async function connect(
  url: string,
  instanceID: string,
  opts?: ConnectOptions,
): Promise<Conn> {
  // Wake the relay if auto-stopped (best-effort).
  await wakeRelay(url);

  const connectURL =
    url.replace(/\/$/, "") + "/ws/" + encodeURIComponent(instanceID);

  const { transport, writer, reader } = await openSession(
    connectURL,
    "connect",
    opts,
  );

  return new Conn(transport, writer, reader, instanceID);
}

/**
 * Wake a Fly.io relay that may be auto-stopped. Sends an HTTPS
 * request to /health, which triggers Fly's proxy to start the machine.
 * No-op if the relay is already running. Best-effort — errors are
 * silently ignored.
 */
export async function wakeRelay(url: string): Promise<void> {
  const healthURL = url.replace(/\/$/, "") + "/health";
  try {
    await fetch(healthURL);
  } catch {
    // Best-effort: the relay may not support HTTPS health checks
    // (e.g., local development with self-signed certs).
  }
}

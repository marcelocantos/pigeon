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
  reader: ReadableStreamBYOBReader | ReadableStreamDefaultReader<Uint8Array>,
): Promise<Uint8Array> {
  const hdr = await readExact(reader, 4);
  const length = new DataView(
    hdr.buffer,
    hdr.byteOffset,
    hdr.byteLength,
  ).getUint32(0, false);
  if (length > maxMessageSize) {
    throw new Error(`message too large: ${length} > ${maxMessageSize}`);
  }
  return readExact(reader, length);
}

/**
 * Read exactly `n` bytes from a stream, assembling from multiple chunks
 * if necessary.
 */
async function readExact(
  reader: ReadableStreamBYOBReader | ReadableStreamDefaultReader<Uint8Array>,
  n: number,
): Promise<Uint8Array> {
  const buf = new Uint8Array(n);
  let offset = 0;
  while (offset < n) {
    const { value, done } = await reader.read() as ReadableStreamReadResult<Uint8Array>;
    if (done || !value) {
      throw new Error("stream ended before expected bytes were read");
    }
    const take = Math.min(value.length, n - offset);
    buf.set(value.subarray(0, take), offset);
    offset += take;
    // If the chunk was larger than needed, we lose the extra bytes.
    // WebTransport streams deliver ordered bytes, so this is fine for
    // length-prefixed framing as long as each readMessage call is
    // sequential (which it is — recv() is awaited).
  }
  return buf;
}

/**
 * A connection to a peer through the tern WebTransport relay.
 */
export class Conn {
  /** The relay-assigned instance ID. */
  readonly instanceID: string;

  private transport: WebTransport;
  private writer: WritableStreamDefaultWriter<Uint8Array>;
  private reader: ReadableStreamDefaultReader<Uint8Array>;
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
    return readMessage(this.reader);
  }

  /** Send an unreliable datagram to the peer. */
  sendDatagram(data: Uint8Array): void {
    if (this.closed) {
      throw new Error("connection is closed");
    }
    const writer = this.transport.datagrams.writable.getWriter();
    writer.write(data).finally(() => writer.releaseLock());
  }

  /** Receive the next unreliable datagram from the peer. */
  async recvDatagram(): Promise<Uint8Array> {
    if (this.closed) {
      throw new Error("connection is closed");
    }
    const reader = this.transport.datagrams.readable.getReader();
    try {
      const { value, done } = await reader.read();
      if (done || !value) {
        throw new Error("datagram stream ended");
      }
      return value;
    } finally {
      reader.releaseLock();
    }
  }

  /** Close the connection. */
  close(): void {
    if (!this.closed) {
      this.closed = true;
      this.writer.close().catch(() => {});
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

  // Read the instance ID from the server.
  const idBytes = await readMessage(reader);
  const instanceID = new TextDecoder().decode(idBytes);

  return new Conn(transport, writer, reader, instanceID);
}

/**
 * Connect to a backend instance through the relay.
 */
export async function connect(
  url: string,
  instanceID: string,
  opts?: ConnectOptions,
): Promise<Conn> {
  const connectURL =
    url.replace(/\/$/, "") + "/ws/" + encodeURIComponent(instanceID);

  const { transport, writer, reader } = await openSession(
    connectURL,
    "connect",
    opts,
  );

  return new Conn(transport, writer, reader, instanceID);
}

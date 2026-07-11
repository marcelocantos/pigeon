// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

const subtle = globalThis.crypto.subtle;

/** Generate 32 cryptographically random bytes suitable for use as a nonce. */
export function generateNonce(): Uint8Array {
  const b = new Uint8Array(32);
  crypto.getRandomValues(b);
  return b;
}

/** Generate 32 cryptographically random bytes suitable for use as a secret. */
export function generateSecret(): Uint8Array {
  return generateNonce();
}

/**
 * Derives a 256-bit key from input keying material via HKDF-SHA256.
 * Uses empty salt (matching Go's hkdf.New(sha256.New, ikm, nil, info)).
 */
export async function deriveKeyFromSecret(
  secret: Uint8Array,
  info: Uint8Array,
): Promise<Uint8Array> {
  const ikm = await subtle.importKey("raw", secret, "HKDF", false, [
    "deriveBits",
  ]);
  const bits = await subtle.deriveBits(
    {
      name: "HKDF",
      hash: "SHA-256",
      salt: new Uint8Array(0),
      info,
    },
    ikm,
    256,
  );
  return new Uint8Array(bits);
}

/**
 * Derives a 6-digit confirmation code from two ECDH public keys.
 * The code is order-independent: deriveConfirmationCode(a, b) === deriveConfirmationCode(b, a).
 */
export async function deriveConfirmationCode(
  pubA: Uint8Array,
  pubB: Uint8Array,
): Promise<string> {
  // Sort lexicographically
  let a = pubA;
  let b = pubB;
  if (compareBytes(a, b) > 0) {
    [a, b] = [b, a];
  }
  // Concat
  const ikm = new Uint8Array(a.length + b.length);
  ikm.set(a, 0);
  ikm.set(b, a.length);

  const derived = await deriveKeyFromSecret(
    ikm,
    new TextEncoder().encode("pairing-confirmation"),
  );
  // Take first 4 bytes big-endian, mod 1000000
  const view = new DataView(derived.buffer, derived.byteOffset, 4);
  const code = view.getUint32(0, false) % 1000000;
  return code.toString().padStart(6, "0");
}

function compareBytes(a: Uint8Array, b: Uint8Array): number {
  const len = Math.min(a.length, b.length);
  for (let i = 0; i < len; i++) {
    if (a[i] !== b[i]) return a[i] - b[i];
  }
  return a.length - b.length;
}

/**
 * X25519 ECDH key pair for key exchange.
 */
export class E2EKeyPair {
  private privateKey: CryptoKey;
  /** 32-byte raw X25519 public key. */
  public publicKeyData: Uint8Array;

  private constructor(privateKey: CryptoKey, publicKeyData: Uint8Array) {
    this.privateKey = privateKey;
    this.publicKeyData = publicKeyData;
  }

  /** Generate a new X25519 key pair. */
  static async create(): Promise<E2EKeyPair> {
    const keyPair = await subtle.generateKey("X25519", true, [
      "deriveBits",
    ]) as CryptoKeyPair;
    const rawPub = await subtle.exportKey("raw", keyPair.publicKey);
    return new E2EKeyPair(keyPair.privateKey, new Uint8Array(rawPub));
  }

  /** Export the raw private key bytes (32 bytes). */
  async exportPrivateKey(): Promise<Uint8Array> {
    // Node.js doesn't support "raw" export for X25519 private keys;
    // use PKCS#8 and strip the 16-byte ASN.1 header.
    const pkcs8 = await subtle.exportKey("pkcs8", this.privateKey);
    const bytes = new Uint8Array(pkcs8);
    // PKCS#8 for X25519: 16-byte header + 32-byte key
    return bytes.slice(bytes.length - 32);
  }

  /**
   * Perform ECDH with a peer's public key, then derive a 256-bit session
   * key via HKDF-SHA256.
   */
  async deriveSessionKey(
    peerPublicKey: Uint8Array,
    info: Uint8Array,
  ): Promise<Uint8Array> {
    const peerKey = await subtle.importKey(
      "raw",
      peerPublicKey,
      "X25519",
      false,
      [],
    );
    const sharedBits = await subtle.deriveBits(
      { name: "X25519", public: peerKey },
      this.privateKey,
      256,
    );
    const shared = new Uint8Array(sharedBits);
    return deriveKeyFromSecret(shared, info);
  }
}

/**
 * Controls how `E2EChannel.decrypt` handles sequence numbers.
 *
 * - `"strict"` (default): sequence numbers must be contiguous with no gaps.
 *   Any out-of-order or replayed packet is rejected. Suitable for reliable
 *   transports (WebTransport streams).
 * - `"datagrams"`: gaps in the sequence number space are accepted, as expected
 *   on lossy transports (UDP, H.264 video). Packets with a sequence number
 *   less than the last accepted one are rejected to prevent replay attacks.
 */
export type ChannelMode = "strict" | "datagrams";

/**
 * AES-256-GCM symmetric encryption channel with counter nonce.
 *
 * Wire format matches Go/Swift/Kotlin implementations:
 *   Nonce: 12 bytes — first 8 = sequence LE, last 4 = zero
 *   Ciphertext: [8-byte seq LE][ciphertext + 16-byte GCM tag]
 *   Additional data: the 8-byte sequence bytes
 */
export class E2EChannel {
  private sendKey: CryptoKey;
  private recvKey: CryptoKey;
  private sendSeq: bigint = 0n;
  private recvSeq: bigint = 0n;
  /** The sequence-number checking mode used by `decrypt`. Defaults to `"strict"`. */
  public mode: ChannelMode = "strict";

  private constructor(sendKey: CryptoKey, recvKey: CryptoKey) {
    this.sendKey = sendKey;
    this.recvKey = recvKey;
  }

  /** Create a channel from raw 32-byte send and receive keys. */
  static async create(
    sendKeyBytes: Uint8Array,
    recvKeyBytes: Uint8Array,
  ): Promise<E2EChannel> {
    const sendKey = await subtle.importKey(
      "raw",
      sendKeyBytes,
      "AES-GCM",
      false,
      ["encrypt"],
    );
    const recvKey = await subtle.importKey(
      "raw",
      recvKeyBytes,
      "AES-GCM",
      false,
      ["decrypt"],
    );
    return new E2EChannel(sendKey, recvKey);
  }

  /**
   * Create a channel from a shared key, deriving directional keys via HKDF.
   * Matches Go's NewSymmetricChannel.
   */
  static async fromSharedKey(
    sharedKey: Uint8Array,
    isServer: boolean,
  ): Promise<E2EChannel> {
    const enc = new TextEncoder();
    let sendInfo = enc.encode("client-to-server");
    let recvInfo = enc.encode("server-to-client");
    if (isServer) {
      [sendInfo, recvInfo] = [recvInfo, sendInfo];
    }
    const sendKeyBytes = await deriveKeyFromSecret(sharedKey, sendInfo);
    const recvKeyBytes = await deriveKeyFromSecret(sharedKey, recvInfo);
    return E2EChannel.create(sendKeyBytes, recvKeyBytes);
  }

  /** Encrypt plaintext. Returns [8-byte seq LE][ciphertext + 16-byte tag]. */
  async encrypt(plaintext: Uint8Array): Promise<Uint8Array> {
    const seq = this.sendSeq;
    this.sendSeq++;

    const seqBytes = new Uint8Array(8);
    new DataView(seqBytes.buffer).setBigUint64(0, seq, true); // LE

    const nonce = new Uint8Array(12);
    nonce.set(seqBytes, 0); // first 8 bytes = seq LE, last 4 = zero

    const ciphertext = await subtle.encrypt(
      { name: "AES-GCM", iv: nonce, additionalData: seqBytes },
      this.sendKey,
      plaintext,
    );

    // Prepend seq bytes
    const result = new Uint8Array(8 + ciphertext.byteLength);
    result.set(seqBytes, 0);
    result.set(new Uint8Array(ciphertext), 8);
    return result;
  }

  /** Decrypt data. Verifies the sequence number according to `mode`. */
  async decrypt(data: Uint8Array): Promise<Uint8Array> {
    if (data.length < 8) {
      throw new Error("ciphertext too short");
    }

    const seqBytes = data.slice(0, 8);
    const seq = new DataView(
      seqBytes.buffer,
      seqBytes.byteOffset,
    ).getBigUint64(0, true);
    const ciphertext = data.slice(8);

    if (this.mode === "strict") {
      if (seq !== this.recvSeq) {
        throw new Error("unexpected sequence number");
      }
      this.recvSeq++;
    } else {
      // datagrams: recvSeq holds highest-accepted-seq + 1 (0n if none yet).
      // Accept seq >= recvSeq (gaps allowed); reject seq < recvSeq (replay).
      if (seq < this.recvSeq) {
        throw new Error("sequence number replayed or too old");
      }
      this.recvSeq = seq + 1n;
    }

    const nonce = new Uint8Array(12);
    nonce.set(seqBytes, 0);

    const plaintext = await subtle.decrypt(
      { name: "AES-GCM", iv: nonce, additionalData: seqBytes },
      this.recvKey,
      ciphertext,
    );
    return new Uint8Array(plaintext);
  }
}

// ---- Session salt (🎯T50) ----

/**
 * Length of the per-session diversifier folded into deriveChannelFromRecord's
 * HKDF info. Matches the activation-path nonce length (T44.1) so reconnects
 * under one PairingRecord never share an AEAD key.
 */
export const SESSION_SALT_LEN = 16;

/**
 * Fixed peer reply to a session-salt handshake on the artifact/reconnect
 * path (plaintext on the primary stream before AEAD is established).
 */
export const SESSION_SALT_ACK = "session_salt_ack";

/** Generate SESSION_SALT_LEN cryptographically random bytes (🎯T50). */
export function generateSessionSalt(): Uint8Array {
  const b = new Uint8Array(SESSION_SALT_LEN);
  crypto.getRandomValues(b);
  return b;
}

/** Concatenate two Uint8Arrays. */
function concatBytes(a: Uint8Array, b: Uint8Array): Uint8Array {
  const out = new Uint8Array(a.length + b.length);
  out.set(a, 0);
  out.set(b, a.length);
  return out;
}

// ---- Pairing record ----

/** Persistent state from a completed pairing ceremony. */
export interface PairingRecord {
  peerInstanceID: string;
  relayURL: string;
  localPrivateKey: string; // base64-encoded raw X25519 private key
  localPublicKey: string; // base64-encoded raw X25519 public key
  peerPublicKey: string; // base64-encoded raw X25519 public key
}

/** Create a pairing record from a completed ceremony. */
export async function createPairingRecord(
  peerInstanceID: string,
  relayURL: string,
  localKeyPair: E2EKeyPair,
  peerPublicKey: Uint8Array,
): Promise<PairingRecord> {
  const privBytes = await localKeyPair.exportPrivateKey();
  return {
    peerInstanceID,
    relayURL,
    localPrivateKey: btoa(String.fromCharCode(...privBytes)),
    localPublicKey: btoa(String.fromCharCode(...localKeyPair.publicKeyData)),
    peerPublicKey: btoa(String.fromCharCode(...peerPublicKey)),
  };
}

/**
 * Derive a channel from a stored pairing record.
 *
 * sessionSalt is a per-session diversifier (typically SESSION_SALT_LEN
 * random bytes exchanged during reconnect). When non-empty it is appended
 * to each direction label (HKDF info = label || salt), so two sessions
 * under one PairingRecord derive distinct AEAD keys (🎯T50). Omit or pass
 * empty only for tests that intentionally exercise the static pre-T50
 * derivation. When non-empty, length must equal SESSION_SALT_LEN.
 */
export async function deriveChannelFromRecord(
  record: PairingRecord,
  sendInfo: Uint8Array,
  recvInfo: Uint8Array,
  sessionSalt: Uint8Array = new Uint8Array(0),
): Promise<E2EChannel> {
  if (sessionSalt.length !== 0 && sessionSalt.length !== SESSION_SALT_LEN) {
    throw new Error(
      `session salt must be ${SESSION_SALT_LEN} bytes, got ${sessionSalt.length}`,
    );
  }
  const privateKeyBytes = Uint8Array.from(atob(record.localPrivateKey), (c) =>
    c.charCodeAt(0),
  );
  const peerPubBytes = Uint8Array.from(atob(record.peerPublicKey), (c) =>
    c.charCodeAt(0),
  );

  // PKCS#8 header for X25519 private keys (16 bytes).
  const pkcs8Header = new Uint8Array([
    0x30, 0x2e, 0x02, 0x01, 0x00, 0x30, 0x05, 0x06, 0x03, 0x2b, 0x65, 0x6e,
    0x04, 0x22, 0x04, 0x20,
  ]);
  const pkcs8 = new Uint8Array(pkcs8Header.length + privateKeyBytes.length);
  pkcs8.set(pkcs8Header, 0);
  pkcs8.set(privateKeyBytes, pkcs8Header.length);

  // Import private key via PKCS#8 (Node.js doesn't support raw X25519 import).
  const privateKey = await subtle.importKey(
    "pkcs8",
    pkcs8,
    "X25519",
    false,
    ["deriveBits"],
  );
  const peerPublicKey = await subtle.importKey(
    "raw",
    peerPubBytes,
    "X25519",
    false,
    [],
  );

  // Derive shared secret via ECDH.
  const sharedBits = await subtle.deriveBits(
    { name: "X25519", public: peerPublicKey },
    privateKey,
    256,
  );
  const shared = new Uint8Array(sharedBits);

  // Derive directional keys via HKDF (info = label || sessionSalt).
  const sendKeyInfo = concatBytes(sendInfo, sessionSalt);
  const recvKeyInfo = concatBytes(recvInfo, sessionSalt);
  const sendKey = await deriveKeyFromSecret(shared, sendKeyInfo);
  const recvKey = await deriveKeyFromSecret(shared, recvKeyInfo);

  return E2EChannel.create(sendKey, recvKey);
}

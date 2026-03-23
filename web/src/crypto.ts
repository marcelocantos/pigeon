// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0

const subtle = globalThis.crypto.subtle;

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
    const keyPair = await subtle.generateKey("X25519", false, [
      "deriveBits",
    ]) as CryptoKeyPair;
    const rawPub = await subtle.exportKey("raw", keyPair.publicKey);
    return new E2EKeyPair(keyPair.privateKey, new Uint8Array(rawPub));
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

  /** Decrypt data. Verifies sequence number to prevent replay attacks. */
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

    if (seq !== this.recvSeq) {
      throw new Error("unexpected sequence number");
    }
    this.recvSeq++;

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

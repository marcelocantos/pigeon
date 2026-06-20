// Copyright 2026 Marcelo Cantos
// SPDX-License-Identifier: Apache-2.0
//
// Code generated from protocol/wireformats.yaml. DO NOT EDIT.

export function encodeUvarint(value: bigint): Uint8Array {
  if (value < 0n) throw new Error("uvarint: negative value");
  const out: number[] = [];
  let v = value;
  while (v >= 0x80n) { out.push(Number((v & 0x7fn) | 0x80n)); v >>= 7n; }
  out.push(Number(v));
  return new Uint8Array(out);
}

export function decodeUvarint(buf: Uint8Array, offset = 0): [bigint, number] {
  let v = 0n; let shift = 0n;
  for (let i = 0; offset + i < buf.length; i++) {
    if (i >= 10) throw new Error("uvarint: too many bytes");
    const b = buf[offset + i];
    if ((b & 0x80) === 0) { v |= BigInt(b) << shift; return [v, i + 1]; }
    v |= BigInt(b & 0x7f) << shift; shift += 7n;
  }
  throw new Error("uvarint: truncated");
}

export function encodeStreamHeader(name: string): Uint8Array {
  const parts: Uint8Array[] = [];
  {
    const bytes = new TextEncoder().encode(name);
    parts.push(encodeUvarint(BigInt(bytes.length)));
    parts.push(bytes);
  }
  let total = 0; for (const p of parts) total += p.length;
  const out = new Uint8Array(total); let off = 0;
  for (const p of parts) { out.set(p, off); off += p.length; }
  return out;
}

export interface StreamHeaderDecoded {
  name: string;
  consumed: number;
}

export function decodeStreamHeader(buf: Uint8Array): StreamHeaderDecoded {
  let off = 0;
  const [_lenU, _ln] = decodeUvarint(buf, off);
  off += _ln;
  const _len = Number(_lenU);
  if (buf.length - off < _len) throw new Error("truncated string");
  const name = new TextDecoder().decode(buf.subarray(off, off + _len));
  off += _len;
  return { name, consumed: off };
}

export function encodeDatagramPlaintext(channelId: bigint, payload: Uint8Array): Uint8Array {
  const parts: Uint8Array[] = [];
  parts.push(encodeUvarint(channelId));
  parts.push(payload);
  let total = 0; for (const p of parts) total += p.length;
  const out = new Uint8Array(total); let off = 0;
  for (const p of parts) { out.set(p, off); off += p.length; }
  return out;
}

export interface DatagramPlaintextDecoded {
  channelId: bigint;
  payload: Uint8Array;
  consumed: number;
}

export function decodeDatagramPlaintext(buf: Uint8Array): DatagramPlaintextDecoded {
  let off = 0;
  const [channelId, _vn] = decodeUvarint(buf, off);
  off += _vn;
  const payload = buf.slice(off);
  off = buf.length;
  return { channelId, payload, consumed: off };
}

export const enum RelayGreetingVariant { Connect = "connect", Register = "register", Listen = "listen" }

export interface RelayGreetingDecoded {
  variant: RelayGreetingVariant;
  instanceId: string;
  token: string;
}

export function encodeRelayGreetingConnect(instanceId: string): Uint8Array {
  let s = "connect:";
  s += instanceId;
  return new TextEncoder().encode(s);
}

export function encodeRelayGreetingRegister(token: string, instanceId: string): Uint8Array {
  let s = "register";
  const parts = [token, instanceId];
  const anyNonEmpty = parts.some((p) => p !== "");
  if (anyNonEmpty) { for (const p of parts) { s += ":"; s += p; } }
  return new TextEncoder().encode(s);
}

export function encodeRelayGreetingListen(token: string, instanceId: string): Uint8Array {
  let s = "listen";
  const parts = [token, instanceId];
  const anyNonEmpty = parts.some((p) => p !== "");
  if (anyNonEmpty) { for (const p of parts) { s += ":"; s += p; } }
  return new TextEncoder().encode(s);
}

export function decodeRelayGreeting(buf: Uint8Array): RelayGreetingDecoded {
  const s = new TextDecoder().decode(buf);
  if (s.startsWith("connect:")) {
    const rest = s.slice(8);
    return { variant: RelayGreetingVariant.Connect, instanceId: rest, token: "" };
  }
  if (s.startsWith("register")) {
    const rest = s.slice(8);
    let suffix: string[];
    if (rest === "") {
      suffix = new Array(2).fill("");
    } else {
      if (!rest.startsWith(":")) throw new Error("relay_greeting: malformed");
      const body = rest.slice(1);
      const parts = body.split(":");
      if (parts.length > 2) {
        const merged = parts.slice(1).join(":");
        suffix = parts.slice(0, 1);
        suffix.push(merged);
      } else {
        suffix = parts.slice();
        while (suffix.length < 2) suffix.push("");
      }
    }
    return { variant: RelayGreetingVariant.Register, instanceId: suffix[1], token: suffix[0] };
  }
  if (s.startsWith("listen")) {
    const rest = s.slice(6);
    let suffix: string[];
    if (rest === "") {
      suffix = new Array(2).fill("");
    } else {
      if (!rest.startsWith(":")) throw new Error("relay_greeting: malformed");
      const body = rest.slice(1);
      const parts = body.split(":");
      if (parts.length > 2) {
        const merged = parts.slice(1).join(":");
        suffix = parts.slice(0, 1);
        suffix.push(merged);
      } else {
        suffix = parts.slice();
        while (suffix.length < 2) suffix.push("");
      }
    }
    return { variant: RelayGreetingVariant.Listen, instanceId: suffix[1], token: suffix[0] };
  }
  throw new Error("relay_greeting: unrecognised prefix");
}


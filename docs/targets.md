# Convergence Targets

## 🎯T1 Tern is a complete library for opaque authenticated relay

All crypto, protocol state machines, code generators, QR helper, and
Swift package live here. Applications import tern rather than duplicating
relay/pairing logic.

**Sub-targets:**

### 🎯T1.1 Crypto library migrated from jevon

Status: done

### 🎯T1.2 Pairing protocol spec migrated from jevon

Status: done

### 🎯T1.3 TLA+ formal model migrated from jevon

Status: done

### 🎯T1.4 Protocol state machine framework migrated from jevon

Status: done

### 🎯T1.5 QR helper migrated from jevon

Status: done

### 🎯T1.6 Swift package (SPM)

Status: done

### 🎯T1.7 E2E integration test

Status: done

### 🎯T1.8 Jevon imports tern's packages

Jevon's `internal/crypto/`, `internal/protocol/`, `internal/qr/`, and
`cmd/protogen/` are replaced by imports from tern. iOS app imports
TernCrypto SPM package.

Status: not started (requires tern to be tagged and pushed)

---

## 🎯T2 Open-source ready — gates v0.1.0

All sub-targets must be done before the v0.1.0 release. No exceptions
without explicit user override.

Audit findings from 2026-03-22 addressed. Tern is a credible public
project with correct code, proper licensing, documentation, and CI.

### 🎯T2.1 Relay enforces single client per instance

Multiple clients connecting to the same instance ID causes concurrent
`inst.conn.Read()` — a data race. The relay rejects a second client
while one is bridged (or closes the previous one).

Status: done


### 🎯T2.2 Swift E2EChannel.encrypt() is concurrency-safe

Lock is released before encryption; concurrent callers can send
out-of-order sequence numbers, causing the receiver to reject. Hold
the lock for the full encrypt operation.

Status: done


### 🎯T2.3 NOTICES file exists with third-party attribution

Apache 2.0 §4(d) requires attribution. All deps (quic-go/quic-go MIT,
quic-go/webtransport-go MIT, skip2/go-qrcode MIT, gopkg.in/yaml.v3
MIT+Apache, golang.org/x/crypto BSD-3, dunglas/httpsfv BSD-3) are
listed with copyright and licence.

Status: done


### 🎯T2.4 README.md exists

Covers what the relay is, trust model, Go package API, Swift SPM
import, deployment, and how to run tests.

Status: done


### 🎯T2.5 GitHub repo settings enforce squash-only merges

`allow_merge_commit: false`, `allow_rebase_merge: false`,
`squash_merge_commit_title: PR_TITLE`, `delete_branch_on_merge: true`.

Status: done


### 🎯T2.6 ExportGo emits ChannelBound and OneShot

Generated `pairingceremony_gen.go` currently drops these fields,
causing the Go representation to differ from the YAML spec. The
generated TLA+ would produce unbounded channels if regenerated from
the Go struct.

Status: done


### 🎯T2.7 ExportGo emits PropertyKind as named constant

Currently emits raw `0`/`1` instead of `Invariant`/`Liveness`.

Status: done


### 🎯T2.8 Relay binary supports --version, --help, --help-agent

Per CLI binary conventions. Includes build-time version injection.

Status: done


### 🎯T2.9 Generated files have SPDX headers

`pairingceremony_gen.go` and `PairingCeremonyMachine.swift` include
copyright and SPDX-License-Identifier lines.

Status: done


### 🎯T2.10 Test coverage for qr, YAML parser, and code generators

`qr/`, `protocol/yaml.go`, `ExportGo`, `ExportSwift`, `ExportPlantUML`
all at 0% coverage. Add basic tests.

Status: done


### 🎯T2.11 CORS wildcard documented as intentional

Both `/register` and `/ws/{id}` use `OriginPatterns: ["*"]`. Document
the design choice and note that deployers should restrict if needed.

Status: done


### 🎯T2.12 go-qrcode dependency evaluated

`github.com/skip2/go-qrcode` last committed 2020, no tagged releases.
Either vendor, replace with a maintained alternative, or accept and
document.

Status: done


### 🎯T2.13 Private project name "jevon" removed from public files

`pairing.yaml` trigger description says `jevon --init`; generated
files and `docs/targets.md` reference jevon.

Status: done


### 🎯T2.14 Health handler w.Write return value handled

`main.go:74` ignores w.Write error.

Status: done


### 🎯T2.15 Test for concurrent clients per instance

Exercises the (now-rejected) second-client scenario to verify the
fix from 🎯T2.1.

Status: done


### 🎯T2.16 formal/tlc uses portable shebang

Change `#!/bin/bash` to `#!/usr/bin/env bash`.

Status: done


### 🎯T2.17 Explicit WebSocket read limit

Call `conn.SetReadLimit()` with a documented value rather than relying
on the implicit 32 KB default.

Status: done


### 🎯T2.18 Instance ID entropy documented

32-bit IDs are adequate for current use but weak for a high-traffic
public relay. Document the trade-off; consider 64-bit if usage grows.

Status: done


### 🎯T2.19 go.mod uses minor version only

Change `go 1.25.7` to `go 1.25`.

Status: done

---

## 🎯T3 Fly.io deployment via CI

Pushes to master auto-deploy to `tern.fly.dev`. Currently deployed
manually with `fly deploy`.

Status: not started

---

## 🎯T4 Bearer token auth on /register

/register requires `Authorization: Bearer <token>` when `TERN_TOKEN`
env var is set. /ws/{id} remains open (instance IDs are unguessable).
Fly.io deployment uses a random 256-bit token.

Status: done

---

## 🎯T5 Multi-transport with LAN upgrade

Devices connected through the relay can discover they're on the same
LAN and upgrade to a direct connection. The `tern.Conn` abstraction
hides this — callers see a single ordered message stream regardless
of transport.

### 🎯T5.1 Reorder-tolerant decryption

`Channel.Decrypt` buffers out-of-order sequence numbers instead of
rejecting them. Required for safe transport cutover.

Status: not started

### 🎯T5.2 LAN discovery via relay

After the encrypted channel is established, both sides exchange their
local IP addresses through the relay. Each attempts a direct WebTransport
connection to the peer's local address using a self-signed certificate.

Status: not started

### 🎯T5.3 Cutover protocol

Each side sends a `CUTOVER` marker as its final message on the old
transport, then sends subsequent messages on the new transport. The
receiver reads from both transports, orders by sequence number, and
closes the old transport after receiving `CUTOVER`.

Status: not started

### 🎯T5.4 Transport-agnostic Conn

`tern.Conn` manages multiple underlying transports. Sends go on the
preferred transport; receives come from any transport and are delivered
in sequence order. Upgrading and downgrading are transparent to the
caller.

Status: not started

---

## 🎯T6 Investigate STUN/NAT hole-punching as a transport

STUN-based peer-to-peer connectivity as a middle tier between relay
and LAN direct. Uses a STUN server to discover public IP/port
mappings, then both peers attempt UDP hole-punching to establish a
direct path without relaying traffic.

Sits between relay and LAN in the transport priority:
1. LAN (same network)
2. STUN/P2P (different networks, direct via hole-punch)
3. Relay (always works, fallback)

Same cutover protocol as 🎯T5.3 applies — it's just another transport.

Needs investigation:
- STUN server requirements (run our own, or use public ones?)
- UDP vs TCP hole-punching (tern already uses QUIC/UDP via WebTransport)
- Success rate across NAT types (symmetric NATs defeat STUN)
- Whether to use ICE (the full WebRTC negotiation framework) or a
  simpler STUN-only approach
- TURN as the fallback when hole-punching fails (essentially our
  existing relay, but standard protocol)
- Go and Swift STUN/ICE libraries (pion/ice, libnice)

Status: not started

---

## 🎯T7 Investigate Bluetooth as proximity oracle

Bluetooth as a proximity signal rather than a data channel. Possible
uses:

- **Pairing trust signal**: if both devices see each other over
  Bluetooth during the pairing ceremony, they're physically co-located.
  Could supplement or replace the 6-digit confirmation code.
- **Transport upgrade hint**: Bluetooth discovery indicates the
  devices are nearby, making LAN direct connection likely to succeed.
- **Proximity-gated handover**: only attempt LAN upgrade when
  Bluetooth confirms proximity, avoiding wasted connection attempts.

Needs investigation: BLE advertising APIs on iOS/Android, power and
battery implications, interaction with existing pairing ceremony,
privacy considerations (Bluetooth MAC address rotation).

Status: not started

---

## 🎯T8 WebTransport relay

WebTransport (QUIC) is the sole transport for the relay path. Enables
both reliable streams (control, pairing) and unreliable datagrams
(H.264 video, real-time data). WebSocket support has been removed.

### 🎯T8.1 WebTransport relay server

WebTransport-only relay server using `quic-go/webtransport-go`.
Self-signed TLS certificate generated at startup for development;
production uses --cert/--key flags. Same endpoints: /register, /ws/{id},
/health. Bridges bidirectional streams + datagrams.

Status: done

### 🎯T8.2 Non-strict Channel.Decrypt for datagrams

Add a windowed/skip mode to Channel.Decrypt that accepts any sequence
number greater than the last seen, instead of requiring exactly the
next one. Required for loss-tolerant datagram delivery (H.264 frames).
The strict sequential mode remains the default for reliable streams.

Status: done

### 🎯T8.3 Go WebTransport client

WebTransport-only Conn in the root tern package. Register and
Connect use WebTransport directly; no WebSocket fallback.

Status: done

### 🎯T8.4 Web/TypeScript WebTransport client

Browser-native WebTransport API in `web/`. Use reliable stream for
control/pairing, datagrams for video.

Status: not started

### 🎯T8.5 LAN direct WebTransport with cert fingerprint

Ephemeral self-signed cert for LAN listener. Include SHA-256 hash in
the LAN offer control message. Browser peers use
`serverCertificateHashes` to accept it (Chromium-only for now; others
fall back to relay-only).

Status: not started

---

## 🎯T9 Raw QUIC protocol for native clients

Native clients (Go, Swift, Kotlin) use raw QUIC (ALPN "tern") instead
of WebTransport. Browser clients continue using WebTransport. The relay
bridges between them transparently via a `relaySession` interface.

### 🎯T9.1 relaySession interface and shared hub

Server-side abstraction: both WebTransport and raw QUIC sessions
implement `relaySession`. Hub stores sessions by interface, not
concrete type. Bridging logic works against the interface.

Status: done

### 🎯T9.2 Raw QUIC server listener

Separate QUIC listener (port 4433) accepting raw QUIC connections
with ALPN "tern". Handshake: client sends "register" or
"connect:<id>" on a bidi stream. Routes to shared hub.

Status: done

### 🎯T9.3 Go client uses raw QUIC

`Register`/`Connect` dial raw QUIC by default. `Conn` wraps
`io.ReadWriteCloser` + datagrammer interface instead of concrete
WebTransport types.

Status: done

### 🎯T9.4 Deployment supports both ports

Fly.io exposes both 443 (WebTransport/browsers) and 4433 (raw QUIC/
native). Dockerfile exposes both. `cmd/tern` starts both servers.

Status: done

### 🎯T9.5 Tests pass against raw QUIC

All existing tests work with raw QUIC client + server path.
Cross-protocol bridging tested (QUIC backend + WebTransport client).

Status: done

---

## 🎯T10 TLA+ model for cutover protocol

Model the transport cutover protocol in TLA+ to verify liveness and
correctness properties:
- No message lost during cutover
- No message duplicated
- No message delivered out of order
- No deadlock (both sides eventually complete the transition)
- Concurrent cutover initiation from both sides is safe

The E2E encryption provides security (verified by the existing pairing
ceremony TLA+ spec). This model focuses on the transport switching
logic: CUTOVER markers, reorder buffer, and the transition from relay
to LAN transport.

Status: not started

---

## 🎯T11 Swift relay client (Network.framework QUIC)

Swift `TernRelay` library using Apple's Network.framework for raw QUIC
connections (ALPN "tern"). Mirrors the Go client's `Register`/`Connect`
API. Available on iOS 16+ / macOS 13+.

### 🎯T11.1 Core TernConn with register/connect/send/recv

`TernConn` with static `register` and `connect` factory methods,
length-prefixed `send`/`recv` on a bidirectional QUIC stream, and
`sendDatagram`/`recvDatagram` for unreliable datagrams.

Status: done

### 🎯T11.2 Integration tests against live QUIC server

End-to-end test connecting the Swift client to a running tern QUIC
server, verifying register, connect, and bidirectional message exchange.

Status: not started


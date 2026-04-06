# Convergence Targets

## Active

### 🎯T1 Tern is a complete library for opaque authenticated relay

All crypto, protocol state machines, code generators, QR helper, and
Swift package live here. Applications import tern rather than duplicating
relay/pairing logic.

- **Weight**: 1.7 (value 5 / cost 3)

**Sub-targets:**

#### 🎯T1.8 Jevon imports tern's packages

Jevon's `internal/crypto/`, `internal/protocol/`, `internal/qr/`, and
`cmd/protogen/` are replaced by imports from tern. iOS app imports
the `Tern` SPM package.

- **Weight**: 1.7 (value 5 / cost 3)
- **Status**: not started (requires tern to be tagged and pushed)

---

---

### 🎯T5 Multi-transport with LAN upgrade

Devices connected through the relay can discover they're on the same
LAN and upgrade to a direct connection. The `tern.Conn` abstraction
hides this — callers see a single ordered message stream regardless
of transport.

- **Weight**: 0.8 (value 5 / cost 6)

#### 🎯T5.1 Reorder-tolerant decryption

`Channel.Decrypt` accepts out-of-order sequence numbers on lossy
transports. `ModeDatagrams` allows gaps (forward jumps) while
rejecting replays. Required for safe transport cutover.

- **Weight**: 1.7 (value 5 / cost 3)
- **Status**: achieved — `ModeDatagrams` in `crypto/crypto.go` accepts
  gaps and rejects replays. `NewDatagramChannel` convenience constructor.
  Full test coverage including reorder, replay, and 50% packet loss.

#### 🎯T5.2 LAN discovery via relay

After the encrypted channel is established, both sides exchange their
local IP addresses through the relay. Each attempts a direct WebTransport
connection to the peer's local address using a self-signed certificate.

- **Weight**: 1.0 (value 5 / cost 5)
- **Status**: not started

#### 🎯T5.3 Cutover protocol

Each side sends a `CUTOVER` marker as its final message on the old
transport, then sends subsequent messages on the new transport. The
receiver reads from both transports, orders by sequence number, and
closes the old transport after receiving `CUTOVER`.

- **Weight**: 1.0 (value 5 / cost 5)
- **Status**: not started
- **Depends on**: 🎯T10

#### 🎯T5.4 Transport-agnostic Conn

`tern.Conn` manages multiple underlying transports. Sends go on the
preferred transport; receives come from any transport and are delivered
in sequence order. Upgrading and downgrading are transparent to the
caller.

- **Weight**: 0.8 (value 5 / cost 6)
- **Status**: not started
- **Depends on**: 🎯T5.2, 🎯T5.3

---

### 🎯T6 Investigate STUN/NAT hole-punching as a transport

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

- **Weight**: 1.5 (value 3 / cost 2)
- **Status**: not started

---

### 🎯T7 Investigate Bluetooth as proximity oracle

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

- **Weight**: 1.0 (value 2 / cost 2)
- **Status**: not started

---

### 🎯T8 WebTransport relay

WebTransport (QUIC) is the sole transport for the relay path. Enables
both reliable streams (control, pairing) and unreliable datagrams
(H.264 video, real-time data). WebSocket support has been removed.

- **Weight**: 1.0 (value 5 / cost 5)

#### 🎯T8.4 Web/TypeScript WebTransport client

Browser-native WebTransport API in `web/`. Use reliable stream for
control/pairing, datagrams for video.

- **Weight**: 1.0 (value 5 / cost 5)
- **Status**: not started

#### 🎯T8.5 LAN direct WebTransport with cert fingerprint

Ephemeral self-signed cert for LAN listener. Include SHA-256 hash in
the LAN offer control message. Browser peers use
`serverCertificateHashes` to accept it (Chromium-only for now; others
fall back to relay-only).

- **Weight**: 1.0 (value 3 / cost 3)
- **Status**: not started
- **Depends on**: 🎯T8.4

---

### 🎯T10 TLA+ model for cutover protocol

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

- **Weight**: 0.6 (value 3 / cost 5)
- **Status**: not started

---

---

### 🎯T14 Browser WebTransport E2E

Prove the browser WebTransport path works end-to-end.

Playwright headless Chromium does NOT support WebTransport/QUIC —
tested against both tern.fly.dev and Google's webtransport.day echo
server; both fail with ERR_QUIC_PROTOCOL_ERROR. The Go WebTransport
client connects to the same server successfully, confirming the server
is correct.

Options:
1. Use Playwright with `headless: false` (headed Chrome) — requires
   a display (CI won't work without Xvfb)
2. Use Selenium + headless Chrome (webtransport-go's own tests do this)
3. Manual verification with a real browser

Server config verified correct: EnableDatagrams on both http3.Server
and quic.Config. Let's Encrypt cert persisted and valid.

- **Weight**: 1.0 (value 3 / cost 3)
- **Status**: blocked on Playwright headless Chromium QUIC support

---

### 🎯T15 Gomobile bindings for iOS and Android

Wrap the Go QUIC client via gomobile to give native iOS and Android
apps the same proven QUIC stack. Replaces the need for platform-specific
QUIC libraries (Network.framework quirks, kwik builder bugs).

Produces:
- iOS: XCFramework importable from Swift
- Android: AAR importable from Kotlin

- **Weight**: 1.0 (value 5 / cost 5)
- **Status**: not started

---

### 🎯T19 Hierarchical state machines in protocol framework

The transport machine has ~25 identical data-forwarding self-loops
per actor because every path state must replicate them. Adding a new
path state (e.g., STUNConnecting for 🎯T6) requires adding 5+ more
copies. The root cause is a flat state machine with no inheritance.

**Desired state**: The YAML protocol framework supports hierarchical
states (Harel statecharts / UML nested states). A superstate defines
transitions inherited by all substates. The transport machine uses a
`Connected` superstate containing all path states, with
data-forwarding transitions defined once on `Connected`.

**Scope**:
- YAML gains nested `states:` with `children:` and `transitions:` at
  each level; substates inherit parent transitions
- Data model (`protocol.go`): states become a tree; `Actor` carries
  the hierarchy
- Runtime (`machine.go`): `HandleEvent` walks up the state hierarchy
  to find matching transitions (innermost wins)
- Go generator (`gogen.go`): emit hierarchical dispatch
- Swift generator (`swift.go`): same
- TLA+ generator (`tla.go`): flatten hierarchy for model checking
  (expand inherited transitions into substates)
- PlantUML generator (`plantuml.go`): emit nested `state` blocks
- Kotlin generator: same as Swift
- Transport YAML refactored to use hierarchy — self-loops eliminated
- Entry/exit actions on superstates (stretch — implement if natural)

**Out of scope (for now)**: Parallel regions (AND-states), history
states. These can be layered on later.

**Forked from**: diagram review during 🎯T18 completion — the
coalesced self-loops made the redundancy visually obvious.

- **Weight**: 1.25 (value 5 / cost 4)
- **Status**: achieved — `StateNode` hierarchy in `protocol.go`,
  `states:` section in YAML parser, `FlattenedTransitions()` expansion,
  all generators updated, session.yaml refactored (~50 self-loops
  eliminated via Connected/LANPath superstates). PlantUML nested
  rendering deferred as cosmetic follow-up.

---

### 🎯T17 Makefile deploy target

`make deploy` that deploys to Fly.io, starts the machine, and waits
for it to be healthy. `make e2e-live` should depend on this so live
tests work without manual machine management.

- **Weight**: 1.5 (value 3 / cost 2)
- **Status**: not started

---

## Achieved

### 🎯T16 Fly.io auto-start for UDP

- **Weight**: 1 (value 1 / cost 1)
- **Status**: done — TCP service on port 443 with auto_start_machines=true.
  Fly proxy wakes the machine on TCP/HTTPS, then UDP flows to the running
  machine. WakeRelay() helper for clients.

### 🎯T3 Fly.io deployment via CI

- **Weight**: 1 (value 1 / cost 1)
- **Status**: done — deploy job in ci.yml runs after tests pass on master push

### 🎯T12 Channel API

- **Weight**: 1 (value 1 / cost 1)
- **Status**: done — streaming and datagram channels implemented in channel.go

### 🎯T12.1 Streaming channels

- **Weight**: 1 (value 1 / cost 1)
- **Status**: done — OpenChannel/AcceptChannel with per-channel encryption

### 🎯T12.2 Datagram channels

- **Weight**: 1 (value 1 / cost 1)
- **Status**: done — DatagramChannel with CRC16 demux and fragmentation support

### 🎯T18 State machine mediates all Conn behavior

- **Weight**: 1 (value 1 / cost 1)
- **Status**: done — Machine drives all I/O via events/commands. Executor
  is a thin event loop. TLA+ emits EVT_*/CMD_* constants. Health monitor
  detects stale LAN via pong timeout. DatagramChannel routed through
  executor. All legacy Conn dispatch removed.

### 🎯T13 Certmagic storage alignment

- **Weight**: 1 (value 1 / cost 1)
- **Status**: done — verified 2026-03-28. ACME account, cert, and key
  persist in /data/certmagic on the Fly volume. Cert is reused on
  restart without re-provisioning.

### 🎯T1.1 Crypto library migrated from jevon

- **Weight**: 1 (value 1 / cost 1)
- **Status**: done

### 🎯T1.2 Pairing protocol spec migrated from jevon

- **Weight**: 1 (value 1 / cost 1)
- **Status**: done

### 🎯T1.3 TLA+ formal model migrated from jevon

- **Weight**: 1 (value 1 / cost 1)
- **Status**: done

### 🎯T1.4 Protocol state machine framework migrated from jevon

- **Weight**: 1 (value 1 / cost 1)
- **Status**: done

### 🎯T1.5 QR helper migrated from jevon

- **Weight**: 1 (value 1 / cost 1)
- **Status**: done

### 🎯T1.6 Swift package (SPM)

- **Weight**: 1 (value 1 / cost 1)
- **Status**: done

### 🎯T1.7 E2E integration test

- **Weight**: 1 (value 1 / cost 1)
- **Status**: done

### 🎯T2 Open-source ready — gates v0.1.0

All sub-targets done. Tern is a credible public project with correct
code, proper licensing, documentation, and CI.

- **Weight**: 1 (value 1 / cost 1)
- **Status**: done

### 🎯T2.1 Relay enforces single client per instance

- **Weight**: 1 (value 1 / cost 1)
- **Status**: done

### 🎯T2.2 Swift E2EChannel.encrypt() is concurrency-safe

- **Weight**: 1 (value 1 / cost 1)
- **Status**: done

### 🎯T2.3 NOTICES file exists with third-party attribution

- **Weight**: 1 (value 1 / cost 1)
- **Status**: done

### 🎯T2.4 README.md exists

- **Weight**: 1 (value 1 / cost 1)
- **Status**: done

### 🎯T2.5 GitHub repo settings enforce squash-only merges

- **Weight**: 1 (value 1 / cost 1)
- **Status**: done

### 🎯T2.6 ExportGo emits ChannelBound and OneShot

- **Weight**: 1 (value 1 / cost 1)
- **Status**: done

### 🎯T2.7 ExportGo emits PropertyKind as named constant

- **Weight**: 1 (value 1 / cost 1)
- **Status**: done

### 🎯T2.8 Relay binary supports --version, --help, --help-agent

- **Weight**: 1 (value 1 / cost 1)
- **Status**: done

### 🎯T2.9 Generated files have SPDX headers

- **Weight**: 1 (value 1 / cost 1)
- **Status**: done

### 🎯T2.10 Test coverage for qr, YAML parser, and code generators

- **Weight**: 1 (value 1 / cost 1)
- **Status**: done

### 🎯T2.11 CORS wildcard documented as intentional

- **Weight**: 1 (value 1 / cost 1)
- **Status**: done

### 🎯T2.12 go-qrcode dependency evaluated

- **Weight**: 1 (value 1 / cost 1)
- **Status**: done

### 🎯T2.13 Private project name "jevon" removed from public files

- **Weight**: 1 (value 1 / cost 1)
- **Status**: done

### 🎯T2.14 Health handler w.Write return value handled

- **Weight**: 1 (value 1 / cost 1)
- **Status**: done

### 🎯T2.15 Test for concurrent clients per instance

- **Weight**: 1 (value 1 / cost 1)
- **Status**: done

### 🎯T2.16 formal/tlc uses portable shebang

- **Weight**: 1 (value 1 / cost 1)
- **Status**: done

### 🎯T2.17 Explicit WebSocket read limit

- **Weight**: 1 (value 1 / cost 1)
- **Status**: done

### 🎯T2.18 Instance ID entropy documented

- **Weight**: 1 (value 1 / cost 1)
- **Status**: done

### 🎯T2.19 go.mod uses minor version only

- **Weight**: 1 (value 1 / cost 1)
- **Status**: done

### 🎯T4 Bearer token auth on /register

- **Weight**: 1 (value 1 / cost 1)
- **Status**: done

### 🎯T8.1 WebTransport relay server

- **Weight**: 1 (value 1 / cost 1)
- **Status**: done

### 🎯T8.2 Non-strict Channel.Decrypt for datagrams

- **Weight**: 1 (value 1 / cost 1)
- **Status**: done

### 🎯T8.3 Go WebTransport client

- **Weight**: 1 (value 1 / cost 1)
- **Status**: done

### 🎯T9 Raw QUIC protocol for native clients

All sub-targets done.

- **Weight**: 1 (value 1 / cost 1)
- **Status**: done

### 🎯T9.1 relaySession interface and shared hub

- **Weight**: 1 (value 1 / cost 1)
- **Status**: done

### 🎯T9.2 Raw QUIC server listener

- **Weight**: 1 (value 1 / cost 1)
- **Status**: done

### 🎯T9.3 Go client uses raw QUIC

- **Weight**: 1 (value 1 / cost 1)
- **Status**: done

### 🎯T9.4 Deployment supports both ports

- **Weight**: 1 (value 1 / cost 1)
- **Status**: done

### 🎯T9.5 Tests pass against raw QUIC

- **Weight**: 1 (value 1 / cost 1)
- **Status**: done

### 🎯T11 Swift relay client (Network.framework QUIC)

All sub-targets done.

- **Weight**: 1 (value 1 / cost 1)
- **Status**: done

### 🎯T11.1 Core TernConn with register/connect/send/recv

- **Weight**: 1 (value 1 / cost 1)
- **Status**: done

### 🎯T11.2 Integration tests against live QUIC server

- **Weight**: 1 (value 1 / cost 1)
- **Status**: done (6/6 local + live via raw NWConnection E2E binary)

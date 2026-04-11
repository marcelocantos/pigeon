# Targets

## Active

### 🎯T1 Pigeon is a complete library for opaque authenticated relay
- **Value**: 5
- **Cost**: 3
- **Acceptance**:
  - All crypto, protocol state machines, code generators, QR helper, and Swift package live in pigeon
  - Applications import pigeon rather than duplicating relay/pairing logic
- **Context**: Parent target for the pigeon library. 7/8 sub-targets achieved — only T1.8 (jevon imports) remains.
- **Depends on**: 🎯T1.1
- **Origin**: bootstrap from targets.md (T1)
- **Status**: Converging
- **Discovered**: 2026-04-08

### 🎯T1.1 Jevon imports pigeon's packages
- **Value**: 5
- **Cost**: 3
- **Acceptance**:
  - Jevon's internal/crypto/, internal/protocol/, internal/qr/, and cmd/protogen/ are replaced by imports from pigeon
  - iOS app imports the Pigeon SPM package
- **Context**: Requires pigeon to be tagged and pushed. v0.15.0 released.
- **Origin**: bootstrap from targets.md (T1.8)
- **Status**: Identified
- **Discovered**: 2026-04-08

### 🎯T2 Multi-transport with LAN upgrade
- **Value**: 5
- **Cost**: 6
- **Acceptance**:
  - Devices connected through relay can discover same LAN and upgrade to direct connection
  - pigeon.Conn abstraction hides transport — callers see single ordered message stream
- **Context**: Parent target. 1/4 sub-targets achieved (T5.1 reorder-tolerant decryption).
- **Depends on**: 🎯T2.1, 🎯T2.2, 🎯T3.1
- **Origin**: bootstrap from targets.md (T5)
- **Status**: Converging
- **Discovered**: 2026-04-08

### 🎯T2.1 Cutover protocol
- **Value**: 5
- **Cost**: 5
- **Acceptance**:
  - Each side sends CUTOVER marker as final message on old transport
  - Receiver reads from both transports, orders by sequence number
  - Old transport closed after receiving CUTOVER
- **Context**: Depends on TLA+ model (T3). Part of multi-transport LAN upgrade.
- **Depends on**: 🎯T3
- **Origin**: bootstrap from targets.md (T5.3)
- **Status**: Identified
- **Discovered**: 2026-04-08

### 🎯T2.2 Transport-agnostic Conn
- **Value**: 5
- **Cost**: 6
- **Acceptance**:
  - pigeon.Conn manages multiple underlying transports
  - Sends go on preferred transport; receives from any transport in sequence order
  - Upgrading and downgrading transparent to caller
- **Context**: Depends on LAN discovery (T3.1) and cutover protocol (T2.1). Final piece of multi-transport.
- **Depends on**: 🎯T3.1, 🎯T2.1
- **Origin**: bootstrap from targets.md (T5.4)
- **Status**: Identified
- **Discovered**: 2026-04-08

### 🎯T3 TLA+ model for cutover protocol
- **Value**: 3
- **Cost**: 5
- **Acceptance**:
  - TLA+ spec verifies no message lost during cutover
  - No message duplicated
  - No message delivered out of order
  - No deadlock
  - Concurrent cutover initiation from both sides is safe
- **Context**: Formal verification of transport switching logic. Blocks T5.3 (cutover protocol implementation).
- **Origin**: bootstrap from targets.md (T10)
- **Status**: Identified
- **Discovered**: 2026-04-08

### 🎯T3.1 LAN discovery via relay
- **Value**: 5
- **Cost**: 5
- **Acceptance**:
  - Both sides exchange local IP addresses through relay after encrypted channel established
  - Each attempts direct WebTransport connection to peer's local address
- **Context**: First step in LAN upgrade path. No dependencies — can start immediately.
- **Origin**: bootstrap from targets.md (T5.2)
- **Status**: Identified
- **Discovered**: 2026-04-08

### 🎯T4 Investigate STUN/NAT hole-punching as a transport
- **Value**: 3
- **Cost**: 2
- **Acceptance**:
  - Investigation complete: STUN server requirements documented
  - UDP vs TCP hole-punching evaluated
  - Success rate across NAT types assessed
  - ICE vs simple STUN approach decided
  - Go and Swift library options identified
- **Context**: Middle-tier transport between relay and LAN. Pure research target.
- **Origin**: bootstrap from targets.md (T6)
- **Status**: Identified
- **Discovered**: 2026-04-08

### 🎯T5 Investigate Bluetooth as proximity oracle
- **Value**: 2
- **Cost**: 2
- **Acceptance**:
  - BLE advertising APIs on iOS/Android evaluated
  - Power/battery implications assessed
  - Interaction with pairing ceremony documented
  - Privacy considerations (MAC rotation) addressed
- **Context**: Bluetooth as proximity signal, not data channel. Could supplement confirmation code or gate LAN upgrade.
- **Origin**: bootstrap from targets.md (T7)
- **Status**: Identified
- **Discovered**: 2026-04-08

### 🎯T6 WebTransport relay
- **Value**: 5
- **Cost**: 5
- **Acceptance**:
  - WebTransport (QUIC) is sole transport for relay path
  - Reliable streams for control/pairing
  - Unreliable datagrams for video/real-time data
- **Context**: Parent target. 3/5 sub-targets achieved. T8.4 (TypeScript client) and T8.5 (LAN direct) remain.
- **Depends on**: 🎯T6.1, 🎯T6.2
- **Origin**: bootstrap from targets.md (T8)
- **Status**: Converging
- **Discovered**: 2026-04-08

### 🎯T6.1 Web/TypeScript WebTransport client
- **Value**: 5
- **Cost**: 5
- **Acceptance**:
  - Browser-native WebTransport API in web/
  - Reliable stream for control/pairing
  - Datagrams for video
- **Context**: Unblocked. Part of WebTransport relay target.
- **Origin**: bootstrap from targets.md (T8.4)
- **Status**: Identified
- **Discovered**: 2026-04-08

### 🎯T6.2 LAN direct WebTransport with cert fingerprint
- **Value**: 3
- **Cost**: 3
- **Acceptance**:
  - Ephemeral self-signed cert for LAN listener
  - SHA-256 hash in LAN offer control message
  - Browser peers use serverCertificateHashes
- **Context**: Depends on T6.1 (Web/TypeScript client). Chromium-only initially.
- **Depends on**: 🎯T6.1
- **Origin**: bootstrap from targets.md (T8.5)
- **Status**: Identified
- **Discovered**: 2026-04-08

### 🎯T7 Browser WebTransport E2E
- **Value**: 3
- **Cost**: 3
- **Acceptance**:
  - Browser WebTransport path works end-to-end against relay
  - Automated or documented manual verification process
- **Context**: Blocked on Playwright headless Chromium QUIC support. Alternatives: headed Chrome, Selenium, or manual.
- **Origin**: bootstrap from targets.md (T14)
- **Status**: Identified
- **Discovered**: 2026-04-08

### 🎯T8 Gomobile bindings for iOS and Android
- **Value**: 5
- **Cost**: 5
- **Acceptance**:
  - iOS XCFramework importable from Swift
  - Android AAR importable from Kotlin
  - Go QUIC stack used by native apps
- **Context**: Replaces platform-specific QUIC libraries (Network.framework quirks, kwik bugs).
- **Origin**: bootstrap from targets.md (T15)
- **Status**: Identified
- **Discovered**: 2026-04-08

### 🎯T9 Makefile deploy target
- **Value**: 3
- **Cost**: 2
- **Acceptance**:
  - make deploy deploys to Fly.io, starts machine, waits for healthy
  - make e2e-live depends on deploy target
- **Context**: Deploy currently broken (Fly.io auth token). Fixing that is a prerequisite.
- **Origin**: bootstrap from targets.md (T17)
- **Status**: Identified
- **Discovered**: 2026-04-08

## Achieved

### 🎯T10 Cross-language E2E test parity
- **Value**: 5
- **Cost**: 2
- **Acceptance**:
  - Every language with generated state machine has automated E2E tests
  - Tests connect to real Go relay, exercise full pairing ceremony
  - Tests run in CI without manual intervention
- **Context**: Parent target. 3/4 sub-targets achieved (T10.1 Swift E2E, T10.2 state machine tests, T10.4 TypeScript local E2E). Only T10.3 (cross-language interop) remains.
- **Depends on**: 🎯T10.1, 🎯T10.3, 🎯T10.4, 🎯T10.2
- **Origin**: bootstrap from targets.md (T20)
- **Status**: Achieved
- **Discovered**: 2026-04-08
- **Achieved**: 2026-04-11
- **Actual-cost**: 8

### 🎯T10.1 Swift E2E integrated into swift test
- **Value**: 1
- **Cost**: 1
- **Acceptance**: TODO
- **Context**: PigeonRelayE2ETests target in Package.swift with RelayE2ETests.swift. 6 tests pass via `swift test`.
- **Status**: Achieved
- **Discovered**: 2026-04-09
- **Achieved**: 2026-04-10
- **Actual-cost**: 1

### 🎯T10.2 State machine unit tests for Swift/Kotlin/TypeScript
- **Value**: 5
- **Cost**: 3
- **Acceptance**:
  - Each language has explicit unit tests for handleEvent
  - Tests verify state transitions for transport phase
  - Tests verify correct command emission per transition
  - Tests verify guard evaluation
- **Context**: Swift (22 tests) and TypeScript (22 tests) pass. Kotlin tests written but blocked on pre-existing codegen compilation errors (CmdID undefined, namespace collisions between SessionMachine/PathSwitchMachine).
- **Origin**: bootstrap from targets.md (T20.2)
- **Status**: Achieved
- **Discovered**: 2026-04-08
- **Achieved**: 2026-04-10
- **Actual-cost**: 3

### 🎯T10.3 Cross-language confirmation code interop test
- **Value**: 5
- **Cost**: 2
- **Acceptance**:
  - Go backend and each non-Go client perform full ECDH key exchange through live relay
  - Both sides independently derive 6-digit confirmation code
  - Test asserts both compute same code
- **Context**: Currently hardcoded '629624' in unit tests. No actual relay interop test. Depends on T10.1 (Swift) and T10.2.
- **Depends on**: 🎯T10.1, 🎯T10.2
- **Origin**: bootstrap from targets.md (T20.3)
- **Status**: Achieved
- **Discovered**: 2026-04-08
- **Achieved**: 2026-04-10
- **Actual-cost**: 3

### 🎯T10.4 TypeScript local E2E tests
- **Value**: 1
- **Cost**: 1
- **Acceptance**: TODO
- **Context**: relay.local.e2e.ts with GoRelayProcess.ts helper. 3 tests (register, round-trip, ordering). Fixed pigeon-bridge port extraction bug.
- **Status**: Achieved
- **Discovered**: 2026-04-09
- **Achieved**: 2026-04-10
- **Actual-cost**: 2

## Graph

```mermaid
graph TD
    T1["Pigeon is a complete library …"]
    T1_1["Jevon imports pigeon's packag…"]
    T2["Multi-transport with LAN upgr…"]
    T2_1["Cutover protocol"]
    T2_2["Transport-agnostic Conn"]
    T3["TLA+ model for cutover protoc…"]
    T3_1["LAN discovery via relay"]
    T4["Investigate STUN/NAT hole-pun…"]
    T5["Investigate Bluetooth as prox…"]
    T6["WebTransport relay"]
    T6_1["Web/TypeScript WebTransport c…"]
    T6_2["LAN direct WebTransport with …"]
    T7["Browser WebTransport E2E"]
    T8["Gomobile bindings for iOS and…"]
    T9["Makefile deploy target"]
    T1 -.->|needs| T1_1
    T2 -.->|needs| T2_1
    T2 -.->|needs| T2_2
    T2 -.->|needs| T3_1
    T2_1 -.->|needs| T3
    T2_2 -.->|needs| T3_1
    T2_2 -.->|needs| T2_1
    T6 -.->|needs| T6_1
    T6 -.->|needs| T6_2
    T6_2 -.->|needs| T6_1
```

# Convergence Report

Evaluated: 2026-04-07

## Standing invariants

- **Tests**: PASSING (CI tests job succeeded on master)
- **CI**: PARTIAL — tests green, **deploy failing** (Fly.io `FLY_API_TOKEN` unauthorized since rename to carrier-pigeon). Not a code issue — needs token rotation in GitHub secrets.

## Movement

- 🎯T20: (new) — added with sub-targets T20.1-T20.4
- 🎯T1, 🎯T1.8, 🎯T5, 🎯T6, 🎯T7, 🎯T8, 🎯T10, 🎯T14, 🎯T15, 🎯T17: (unchanged)

## Gap Report

### 🎯T20 Cross-language E2E test parity  [weight: 2.5]
Gap: not started
Parent target. 0/4 sub-targets achieved. Swift standalone E2E binary exists but is not in `swift test`. Kotlin has local E2E tests (PigeonConnE2ETest starts relay subprocess). TypeScript E2E requires live relay + TERN_TOKEN. No state machine unit tests in any non-Go language.

  [ ] 🎯T20.1 Swift E2E integrated into `swift test` — not started: standalone `e2e/swift/main.swift` exists, Package.swift has `PigeonTests` target but no relay E2E test target
  [ ] 🎯T20.2 State machine unit tests for Swift/Kotlin/TypeScript — not started
  [ ] 🎯T20.3 Cross-language confirmation code interop test — not started (blocked on 🎯T20.1, 🎯T20.2)
  [ ] 🎯T20.4 TypeScript local E2E tests — not started: `relay.e2e.ts` exists but requires TERN_TOKEN, no local relay subprocess

### 🎯T1.8 Jevon imports pigeon's packages  [weight: 1.7]
Gap: not started
v0.14.0 released. Migration in jevon repo still pending — requires updating jevon's imports to pigeon packages.

### 🎯T6 Investigate STUN/NAT hole-punching  [weight: 1.5]
Gap: not started
Pure research target. No code or investigation artifacts exist yet.

### 🎯T17 Makefile deploy target  [weight: 1.5]
Gap: not started
No Makefile deploy target exists. Note: deploy is currently broken (Fly.io auth) — fixing that is a prerequisite.

### 🎯T1 Pigeon is a complete library  [weight: 1.7]
Gap: converging (7/8 sub-targets achieved)

  [x] 🎯T1.1 Crypto library — achieved
  [x] 🎯T1.2 Pairing protocol spec — achieved
  [x] 🎯T1.3 TLA+ formal model — achieved
  [x] 🎯T1.4 Protocol state machine framework — achieved
  [x] 🎯T1.5 QR helper — achieved
  [x] 🎯T1.6 Swift package — achieved
  [x] 🎯T1.7 E2E integration test — achieved
  [ ] 🎯T1.8 Jevon imports pigeon's packages — not started

### 🎯T8 WebTransport relay  [weight: 1.0]
Gap: converging (3/5 sub-targets achieved)

  [x] 🎯T8.1 WebTransport relay server — achieved
  [x] 🎯T8.2 Non-strict Channel.Decrypt — achieved
  [x] 🎯T8.3 Go WebTransport client — achieved
  [ ] 🎯T8.4 Web/TypeScript WebTransport client — not started
  [ ] 🎯T8.5 LAN direct WebTransport — not started (blocked on 🎯T8.4)

### 🎯T14 Browser WebTransport E2E  [weight: 1.0]  (status only)
Status: blocked on Playwright headless Chromium QUIC support

### 🎯T15 Gomobile bindings  [weight: 1.0]  (status only)
Status: not started

### 🎯T7 Investigate Bluetooth as proximity oracle  [weight: 1.0]  (status only)
Status: not started

### 🎯T5 Multi-transport with LAN upgrade  [weight: 0.8]
Gap: converging (1/4 sub-targets achieved)

  [x] 🎯T5.1 Reorder-tolerant decryption — achieved
  [ ] 🎯T5.2 LAN discovery via relay — not started
  [ ] 🎯T5.3 Cutover protocol — not started (blocked on 🎯T10)
  [ ] 🎯T5.4 Transport-agnostic Conn — not started (blocked on 🎯T5.2, 🎯T5.3)

### 🎯T10 TLA+ model for cutover protocol  [weight: 0.6]  (status only)
Status: not started. Low effective weight — blocks 🎯T5.3 but expensive relative to value.

### Blocked targets

- 🎯T20.3 Cross-language confirmation code interop test — blocked on 🎯T20.1, 🎯T20.2
- 🎯T5.3 Cutover protocol — blocked on 🎯T10
- 🎯T5.4 Transport-agnostic Conn — blocked on 🎯T5.2, 🎯T5.3
- 🎯T8.5 LAN direct WebTransport — blocked on 🎯T8.4

## Recommendation

Work on: **🎯T20.1 Swift E2E integrated into `swift test`**
Reason: Highest effective weight (2.5) among unblocked leaf targets. The standalone E2E binary (`e2e/swift/main.swift`) already proves the Swift QUIC path works — this is packaging it as an XCTest target that starts a Go relay subprocess. Low cost, high value, and unblocks 🎯T20.3 (cross-language confirmation code interop).

## Suggested action

Add a new test target in `Package.swift` (e.g., `PigeonRelayE2ETests`) that depends on `Pigeon`. Create a test file that starts a Go relay subprocess (`go run ./cmd/pigeon`), runs the register/connect/stream/crypto round-trip tests from `e2e/swift/main.swift`, and tears down the subprocess. Use `Process` to manage the relay lifecycle. Use `/push` to drive the PR workflow.

<!-- convergence-deps
evaluated: 2026-04-07T00:00:00Z
sha: e27b3c0

🎯T1:
  gap: close
  assessment: "7/8 sub-targets achieved. Only T1.8 (jevon imports) remains."
  read:
    - docs/targets.md

🎯T1.8:
  gap: not started
  assessment: "v0.14.0 released. Jevon migration pending."
  read:
    - docs/targets.md

🎯T5:
  gap: significant
  assessment: "1/4 sub-targets achieved (T5.1). T5.2 unblocked but not started. T5.3 blocked on T10."
  read:
    - docs/targets.md

🎯T6:
  gap: not started
  assessment: "No investigation artifacts exist."
  read:
    - docs/targets.md

🎯T17:
  gap: not started
  assessment: "No Makefile deploy target. Deploy currently broken (Fly.io auth)."
  read:
    - docs/targets.md

🎯T8:
  gap: significant
  assessment: "3/5 sub-targets achieved. T8.4 (TypeScript client) and T8.5 (LAN direct) remain."
  read:
    - docs/targets.md

🎯T20:
  gap: not started
  assessment: "0/4 sub-targets achieved. Swift standalone binary exists but not in swift test. Kotlin has local E2E. TypeScript needs live relay."
  read:
    - docs/targets.md
    - e2e/swift/main.swift
    - web/src/relay.e2e.ts
    - Package.swift
    - android/pigeon/src/test/kotlin/com/marcelocantos/pigeon/relay/PigeonConnE2ETest.kt

🎯T20.1:
  gap: not started
  assessment: "Standalone e2e/swift/main.swift exists with full test coverage. Package.swift has PigeonTests but no relay E2E test target."
  read:
    - e2e/swift/main.swift
    - Package.swift

🎯T20.2:
  gap: not started
  assessment: "No state machine unit tests in any non-Go language."
  read:
    - docs/targets.md

🎯T20.3:
  gap: not started
  assessment: "Blocked on T20.1 and T20.2. No cross-language confirmation code interop test exists."
  read:
    - docs/targets.md

🎯T20.4:
  gap: not started
  assessment: "relay.e2e.ts exists but skips when TERN_TOKEN not set. No local relay subprocess like Kotlin has."
  read:
    - web/src/relay.e2e.ts

standing-invariant:
  gap: partial
  assessment: "Tests pass. Deploy failing — Fly.io FLY_API_TOKEN unauthorized after app rename to carrier-pigeon."
  read:
    - .github/workflows/ci.yml
-->

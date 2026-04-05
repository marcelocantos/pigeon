# Convergence Report

Evaluated: 2026-04-05

## Standing invariants

- **Tests: FAILING** — `TestHealthMonitorFallback` timed out in CI (last 5 runs all failed). The test expects the health monitor to trigger fallback within 25s after closing the LAN server, but it never fires. This has been failing since at least the "Uniform grey for cross-actor message arrows" commit.
- **CI: RED** — all 5 recent master pushes failed with the same test.

## Movement

- 🎯T16: not started -> **achieved** (TCP auto_start_machines in fly.toml confirmed; moved to Achieved)
- 🎯T18: (new target, not started)
- All others: (unchanged)

## Gap Report

### 🎯T1 Tern is a complete library  [weight: 1.7]
Gap: converging (7/8 sub-targets achieved)

  [x] 🎯T1.1 Crypto library — achieved
  [x] 🎯T1.2 Pairing protocol spec — achieved
  [x] 🎯T1.3 TLA+ formal model — achieved
  [x] 🎯T1.4 Protocol state machine framework — achieved
  [x] 🎯T1.5 QR helper — achieved
  [x] 🎯T1.6 Swift package — achieved
  [x] 🎯T1.7 E2E integration test — achieved
  [ ] 🎯T1.8 Jevon imports tern's packages — not started

### 🎯T1.8 Jevon imports tern's packages  [weight: 1.7]
Gap: not started
v0.9.0 released and tagged. Migration in jevon repo still pending.

### 🎯T5.1 Reorder-tolerant decryption  [weight: 1.7]
Gap: not started
No buffering logic in Channel.Decrypt.

### 🎯T6 Investigate STUN/NAT hole-punching  [weight: 1.5]  (status only)
Status: not started

### 🎯T17 Makefile deploy target  [weight: 1.5]  (status only)
Status: not started

### 🎯T18 State machine mediates all Conn behavior  [weight: 1.25]
Gap: not started
Target added. Recent commits (session_gen.go typed state machines, Go export structs) are precursor work but the core integration — machine driving Conn I/O — hasn't begun.

### 🎯T7 Investigate Bluetooth as proximity oracle  [weight: 1.0]  (status only)
Status: not started

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

### 🎯T5 Multi-transport with LAN upgrade  [weight: 0.8]
Gap: not started (0/4 sub-targets achieved)
All sub-targets blocked or not started.

### 🎯T10 TLA+ model for cutover protocol  [weight: 0.6]  (status only)
Status: not started. Low effective weight.

### Blocked targets

- 🎯T5.2 LAN discovery via relay — blocked on 🎯T5.1
- 🎯T5.3 Cutover protocol — blocked on 🎯T5.1, 🎯T10
- 🎯T5.4 Transport-agnostic Conn — blocked on 🎯T5.2, 🎯T5.3
- 🎯T8.5 LAN direct WebTransport — blocked on 🎯T8.4

## Recommendation

Work on: **Fix TestHealthMonitorFallback (standing invariant violation)**
Reason: CI has been red for the last 5 master pushes. All convergence is blocked while the test suite fails — no PR can be merged cleanly. The `TestHealthMonitorFallback` test in `pathswitch_test.go:454` expects the health monitor to detect a dead LAN path and fall back to relay within 25s, but the fallback never triggers. This must be fixed before any target work can proceed.

## Suggested action

Read `pathswitch_test.go` and the health monitor implementation (likely in `pathswitch.go` or `router.go`) to understand why the health monitor isn't detecting the closed LAN server. Check if the health ping interval, failure threshold, or fallback trigger changed in recent commits (the typed state machine generation work may have altered Conn/router behavior). Run the test locally with `-v` to get detailed logs.

<!-- convergence-deps
evaluated: 2026-04-05T08:00:00Z
sha: 921c7ae

🎯T1:
  gap: close
  assessment: "7/8 sub-targets achieved. Only T1.8 (jevon imports) remains."
  read:
    - docs/targets.md

🎯T1.8:
  gap: not started
  assessment: "v0.9.0 released. Jevon migration pending."
  read:
    - docs/targets.md

🎯T5.1:
  gap: not started
  assessment: "No buffering logic in Channel.Decrypt."
  read:
    - docs/targets.md

🎯T16:
  gap: achieved
  assessment: "TCP auto_start_machines confirmed in fly.toml. Moved to Achieved."
  read:
    - fly.toml
    - docs/targets.md

🎯T18:
  gap: not started
  assessment: "Target added. Typed state machine generation is precursor work but core integration not started."
  read:
    - docs/targets.md

standing-invariant:
  gap: failing
  assessment: "TestHealthMonitorFallback times out in CI. 5 consecutive failures on master."
  read:
    - pathswitch_test.go
-->

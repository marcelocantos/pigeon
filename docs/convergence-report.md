# Convergence Report

Evaluated: 2026-04-06

## Standing invariants

- **Tests**: PASSING (test job succeeded on CI run #24001562484)
- **CI**: FAILING — Deploy to Fly.io fails. Dockerfile missing `COPY protocol/ protocol/` and `COPY qr/ qr/` — the root package now imports these after T18 work (executor.go imports protocol, lan.go imports qr). Tests pass; only the deploy step is broken.

## Movement

- 🎯T18: close -> **achieved** (moved to Achieved section with full status)
- 🎯T1, 🎯T5.1, 🎯T6, 🎯T17, 🎯T19: (unchanged)
- All others: (unchanged)

## Gap Report

### 🎯T1.8 Jevon imports tern's packages  [weight: 1.7]
Gap: not started
v0.9.0 released. Migration in jevon repo still pending. This is external work in the jevon codebase.

### 🎯T5.1 Reorder-tolerant decryption  [weight: 1.7]
Gap: not started
No buffering logic in `Channel.Decrypt`. Needed for safe transport cutover.

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

### 🎯T6 Investigate STUN/NAT hole-punching  [weight: 1.5]  (status only)
Status: not started

### 🎯T17 Makefile deploy target  [weight: 1.5]  (status only)
Status: not started
  Implied: CI deploy broken — Dockerfile needs protocol/ and qr/ directories. Fixing the Dockerfile is a prerequisite for any deploy-related work.

### 🎯T19 Parallel regions in protocol state machine  [weight: 1.25]  (status only)
Status: not started

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

Work on: **Fix Dockerfile (CI deploy broken)**
Reason: Standing invariant violation takes priority over all explicit targets. The Dockerfile is missing `COPY protocol/ protocol/` and `COPY qr/ qr/`, causing every deploy to fail. This is a 2-line fix that restores CI to green. After that, the highest-leverage unblocked targets are 🎯T1.8 and 🎯T5.1 (both weight 1.7), followed by 🎯T6 and 🎯T17 (both weight 1.5).

## Suggested action

Add the missing COPY lines to the Dockerfile, then push via `/push` to get CI green:

```dockerfile
COPY protocol/ protocol/
COPY qr/ qr/
```

These go after `COPY crypto/ crypto/` and before the `RUN CGO_ENABLED=0 go build` line. After CI is green, proceed with either 🎯T1.8 (jevon migration) or 🎯T5.1 (reorder-tolerant decryption) based on which is more actionable.

<!-- convergence-deps
evaluated: 2026-04-06T00:00:00Z
sha: bf9d713

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

🎯T17:
  gap: not started
  assessment: "Not started. CI deploy broken — Dockerfile needs protocol/ and qr/."
  read:
    - docs/targets.md
    - Dockerfile

🎯T8:
  gap: close
  assessment: "3/5 sub-targets achieved. T8.4 (TypeScript client) and T8.5 (LAN direct) remain."
  read:
    - docs/targets.md

standing-invariant:
  gap: failing
  assessment: "Tests pass. Deploy fails — Dockerfile missing COPY protocol/ and COPY qr/."
  read:
    - Dockerfile
-->

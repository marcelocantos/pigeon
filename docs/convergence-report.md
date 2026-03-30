# Convergence Report

Evaluated: 2026-03-29

Standing invariants: all green. CI passing (3/3 recent runs succeeded, including v0.9.0 release). No open PRs. Working tree clean.

## Movement

- 🎯T3: not started -> **achieved** (deploy job confirmed in ci.yml)
- 🎯T12: not started -> **achieved** (streaming + datagram channels implemented)
- 🎯T12.1: not started -> **achieved** (OpenChannel/AcceptChannel in channel.go, 32 tests)
- 🎯T12.2: not started -> **achieved** (DatagramChannel with CRC16 demux + fragmentation)
- 🎯T13: (unchanged, achieved)
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
  [ ] 🎯T1.8 Jevon imports tern's packages — not started (requires tern tag + push)

### 🎯T16 Fly.io auto-start for UDP  [weight: 1.7]
Gap: not started
No implementation or investigation artifacts found. Fly machine still requires manual start for UDP traffic.

### 🎯T1.8 Jevon imports tern's packages  [weight: 1.7]
Gap: not started
Requires tagging and pushing tern, then migrating jevon imports. v0.9.0 has been released — this unblocks the migration.

### 🎯T5.1 Reorder-tolerant decryption  [weight: 1.7]
Gap: not started
No buffering logic in Channel.Decrypt. First step in the 🎯T5 multi-transport chain.

### 🎯T6 Investigate STUN/NAT hole-punching  [weight: 1.5]  (status only)
Status: not started

### 🎯T17 Makefile deploy target  [weight: 1.5]  (status only)
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
All sub-targets blocked or not started. 🎯T5.3 blocked on 🎯T10.

### 🎯T10 TLA+ model for cutover protocol  [weight: 0.6]  (status only)
Status: not started. Low effective weight (cost exceeds value).

### Blocked targets

- 🎯T12.2 Datagram channels — **achieved** (no longer blocked)
- 🎯T5.2 LAN discovery via relay — blocked on 🎯T5.1
- 🎯T5.3 Cutover protocol — blocked on 🎯T5.1, 🎯T10
- 🎯T8.5 LAN direct WebTransport — blocked on 🎯T8.4
- 🎯T5.4 Transport-agnostic Conn — blocked on 🎯T5.2, 🎯T5.3

## Recommendation

Work on: **🎯T1.8 Jevon imports tern's packages**
Reason: At weight 1.7 (tied with T16, T5.1), T1.8 is the highest-leverage choice because v0.9.0 has just been released and tagged, removing the last blocker. Completing T1.8 closes out 🎯T1 entirely (the last remaining sub-target). It also validates the library's public API by having a real consumer import it — any issues found now save pain later. 🎯T16 and 🎯T5.1 are equally weighted but don't close a parent target.

## Suggested action

In the jevon repo, replace `internal/crypto/`, `internal/protocol/`, `internal/qr/`, and `cmd/protogen/` with imports from `github.com/marcelocantos/tern`. Update `go.mod` to require `github.com/marcelocantos/tern v0.9.0`. Update the iOS app's Package.swift to add the Tern SPM package dependency. Run tests in both repos to confirm the migration.

<!-- convergence-deps
evaluated: 2026-03-29T10:00:00Z
sha: 4653214

🎯T1:
  gap: close
  assessment: "7/8 sub-targets achieved. Only T1.8 (jevon imports) remains."
  read:
    - docs/targets.md

🎯T3:
  gap: achieved
  assessment: "Deploy job in ci.yml runs after tests pass on master push."
  read:
    - .github/workflows/ci.yml

🎯T12:
  gap: achieved
  assessment: "Both streaming and datagram channels implemented in channel.go with 32 tests."
  read:
    - channel.go
    - channel_test.go
    - conn.go

🎯T12.1:
  gap: achieved
  assessment: "OpenChannel/AcceptChannel with per-channel encryption, full test coverage."
  read:
    - channel.go
    - channel_test.go

🎯T12.2:
  gap: achieved
  assessment: "DatagramChannel with CRC16 demux, fragmentation, full test coverage."
  read:
    - channel.go
    - channel_test.go

🎯T16:
  gap: not started
  assessment: "No implementation or investigation artifacts found."
  read:
    - fly.toml

🎯T1.8:
  gap: not started
  assessment: "Requires tern tag (done: v0.9.0) then jevon migration."
  read:
    - docs/targets.md

🎯T5.1:
  gap: not started
  assessment: "No buffering logic in Channel.Decrypt."
  read:
    - crypto/crypto.go

🎯T14:
  gap: significant
  assessment: "Blocked on Playwright headless Chromium QUIC support. Alternative approaches identified."
  read:
    - docs/targets.md
-->

# Convergence Report

Evaluated: 2026-03-28

Standing invariants: all green. CI passing (5/5 recent runs succeeded). No open PRs.

## Gap Report

### 🎯T13 Certmagic storage alignment  [highest weight: 2.5]
Gap: significant
`fly.toml` sets `XDG_DATA_HOME=/data/certmagic` and mounts a volume at `/data/certmagic`. Dockerfile creates the directory. Configuration looks correct, but the target requires runtime verification that certs, ACME accounts, and OCSP staples persist across deploys and machine restarts. Needs a deploy + restart cycle to confirm.

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

### 🎯T3 Fly.io deployment via CI  [weight: 1.7]
Gap: not started
CI workflow exists (`.github/workflows/ci.yml`) but only runs `go test`. No Fly.io deploy step. Needs a deploy workflow or a step in the existing CI that calls `flyctl deploy`.

### 🎯T5.1 Reorder-tolerant decryption  [weight: 1.7]
Gap: not started
No buffering logic exists in `Channel.Decrypt`. First step in the 🎯T5 multi-transport chain.

### 🎯T12 Channel API  [weight: 1.7]
Gap: not started (0/2 sub-targets achieved)

  [ ] 🎯T12.1 Streaming channels — not started (no OpenChannel/AcceptChannel functions exist)
  [ ] 🎯T12.2 Datagram channels — not started (blocked on 🎯T12.1)

### 🎯T16 Fly.io auto-start for UDP  [weight: 1.7]
Gap: not started (status only)
No implementation or investigation artifacts found.

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
Status: not started. Note: the LE rate limit block (retries after 2026-03-24) has expired — this target is now unblocked.

### 🎯T15 Gomobile bindings  [weight: 1.0]  (status only)
Status: not started

### 🎯T5 Multi-transport with LAN upgrade  [weight: 0.8]
Gap: not started (0/4 sub-targets achieved)
All sub-targets blocked or not started. 🎯T5.3 blocked on 🎯T10 (TLA+ model).

### 🎯T10 TLA+ model for cutover protocol  [weight: 0.6]
Gap: not started (status only)
Low weight (cost exceeds value ratio). Consider whether this is needed before 🎯T5.3 implementation or can be done in parallel.

## Recommendation

Work on: **🎯T13 Certmagic storage alignment**
Reason: Highest effective weight (2.5) among unblocked targets. Low cost (2), high value (5). The infrastructure configuration appears correct but needs runtime verification. Resolving this ensures Let's Encrypt certs persist, which also unblocks 🎯T14 (browser WebTransport E2E) from the cert provisioning side.

## Suggested action

SSH into the Fly.io machine (`fly ssh console`) and verify that `/data/certmagic` contains certmagic's expected directory structure (accounts, certificates, ocsp). Check that `certmagic.NewDefault()` uses the `XDG_DATA_HOME` path. If the directory is empty or missing expected files, inspect certmagic's `DefaultStorage` to confirm it reads `XDG_DATA_HOME`. Deploy, wait for cert provisioning, then restart the machine and verify the cert is reused without re-provisioning.

<!-- convergence-deps
evaluated: 2026-03-28T00:00:00Z
sha: c28f530

🎯T13:
  gap: significant
  assessment: "fly.toml config looks correct (XDG_DATA_HOME + volume mount), but needs runtime verification."
  read:
    - fly.toml
    - cmd/tern/main.go
    - Dockerfile
    - docs/targets.md

🎯T1:
  gap: close
  assessment: "7/8 sub-targets achieved. Only T1.8 (jevon imports) remains — requires tagging."
  read:
    - docs/targets.md

🎯T3:
  gap: not started
  assessment: "CI runs tests only. No fly deploy step."
  read:
    - .github/workflows/ci.yml

🎯T12:
  gap: not started
  assessment: "No OpenChannel/AcceptChannel functions exist."
  read:
    - conn.go
    - tern.go

🎯T5.1:
  gap: not started
  assessment: "No buffering logic in Channel.Decrypt."
  read:
    - crypto/crypto.go
-->

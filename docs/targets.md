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

Apache 2.0 §4(d) requires attribution. All deps (coder/websocket ISC,
skip2/go-qrcode MIT, gopkg.in/yaml.v3 MIT+Apache, golang.org/x/crypto
BSD-3) are listed with copyright and licence.

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


# Entropy audit — pigeon

Date: 2026-08-22
Mode: full (entropy + hygiene)
Auditor: Grok Build subagent (sole entropy-audit owner)

## Executive summary

- **Snapshot:** `/Users/marcelo/work/github.com/marcelocantos/pigeon`
  - Branch: `master` (tracking `origin/master`, **ahead 2**)
  - HEAD: `ffb6319e7a31709b6776ce5aad9d13a2c12a6467`
    (`ffb6319 Add PigeonCore pure-Swift product for iOS consumers`, 2026-08-08)
  - Nearest tag: `v0.32.0` (`v0.32.0-2-gffb6319`)
  - Initial dirty state (`git status --porcelain=v1 -b`): clean working tree;
    only `## master...origin/master [ahead 2]`
- **Scope:** whole repository as shipped. Languages analysed: Go, C, Swift,
  Kotlin/Android, TypeScript/web, Bash/Make, TLA+. Rust, Python, and SQL are
  not product languages here.
- **Exclusions (named, not silent):** `c/vendor/github.com/**` (libsodium,
  ngtcp2, quictls/openssl submodules); `c/vendor/build/**` (local static
  libs); `web/node_modules/**`; `android/**/build/**`; generated FSM/wire
  sources treated as *outputs* of YAML, not independent design; TLC
  `formal/states/` (gitignored); prior security audit
  `docs/audit/fable-2026-07.md` used as history, not as this report's
  denominator.
- **Headline mechanism:** the declared architecture is “YAML is the single
  source of truth; protogen mirrors it into every SDK; C is the peer-library
  core.” Observed reality is a **fan-out that is only half-wired**: generators
  write Swift into a directory the package no longer compiles; `make generate`
  omits two of four YAML specs; CI ratchets only `dist/`; crypto and pairing
  still have live N-way implementations whose lockstep is incomplete.
- **Highest-consequence findings:** ENT-001 (generator path / generate
  target / CI ratchet mismatch), ENT-002 (CI does not run the declared
  `make bullseye` gate), ENT-003 (five crypto implementations, vectors
  consumed only by C), ENT-004 (pairing production driver is hand-rolled C,
  not the generated FSM the comments claim).
- **Unverified residue:** `make bullseye`, Swift/Kotlin/TLC/Playwright, and
  live Fly paths were **not** executed in this audit. Remaining Fable-5
  crypto claims were not re-litigated as P0; post-audit code shows T52 SAS
  commit, T53 per-stream keys, and a cwire mutex — those look remediating,
  not current catastrophic failures. Clone detection was not run (no
  in-repo jscpd; analyzers were not installed).

## Scope and exclusions

| Tree | Role | How treated |
|---|---|---|
| `protocol/*.yaml` | Declared SSoT for FSMs and wire formats | Read as architecture, not generated |
| `*_gen.go`, `*Machine.swift/kt/ts`, `c/*_gen.[ch]`, `dist/pigeon.[ch]` | Protogen / amalgamate outputs | Excluded from “hand-written quality”; checked for *drift gates* |
| `c/vendor/github.com/**` | Vendored C deps (submodules) | Named exclusion |
| `c/vendor/build/**` | Host static libs | Not gitignored (`build-android/` is); local-only |
| `web/node_modules/**`, `android/**/build/**` | Tooling artifacts | Excluded |
| `formal/*_TTrace_*`, `formal/states/` | TLC traces | Gitignored; not audited |
| `formal/Session.old`, `formal/Session_Transport.old` | Stale generated TLA | Residue (ENT-012) |
| `docs/audit/fable-2026-07.md` | Prior security audit | History / counterevidence, not this inventory |
| `faultproxy/` | Test-only fault injector | Noted; not a product surface |

No `AGENTS.md` or `hygiene.yaml` at repo root. `CLAUDE.md` / `Claude.md` are
the same file on this case-insensitive volume.

## Commands run

| Command | Version / notes | Exit | Shipped vs auxiliary | Limitation |
|---|---|---|---|---|
| `git status --porcelain=v1 -b`; `git rev-parse HEAD`; `git log -1` | git 2.55.0 | 0 | provenance | — |
| `git describe --tags --always`; `git tag --sort=-v:refname` | | 0 | provenance | — |
| `git ls-files` counts, generated-marker grep, submodule status | | 0 | inventory | Submodule nested `-` entries (openssl extras) not expanded |
| `GOWORK=off go list ./...` + import SCC | go 1.26.4 darwin/arm64 | 0 | auxiliary graph | Default `go list` fails on a parent `go.work`; `GOWORK=off` is the module |
| `git log --since=2025-08-01 --name-only` churn | | 0 | history | Counts deleted files (`conn.go`, `tern.go`, `docs/targets.md`) |
| `gofmt -l` on tracked `*.go` (sample/head) | | 0, empty | auxiliary format | Not a full-tree CI equivalent |
| `~/.claude/skills/hygiene/hygiene_check.py` | uv-run script | 0 with traceback | hygiene | **No `hygiene.yaml`**; `FileNotFoundError` (see Hygiene posture) |
| Manifest/CI/docs targeted reads | — | — | source | — |
| `make bullseye` / `go test ./...` / `swift test` / gradle / TLC / Playwright | **not run** | — | shipped oracles | Time/cost; residue declared |

No analyzers were installed. jscpd, golangci-lint, and cargo tools are absent
from the repo; proposing them is residue, not a silent metric.

## Observed architecture

### Entry points and deployable units

| Unit | Path | Role |
|---|---|---|
| Relay binary | `cmd/pigeon/main.go` | WebTransport (UDP/TCP 443) + raw QUIC ALPN `pigeon` (UDP 4433); Fly.io via `Dockerfile` + `fly.toml` |
| Go library | module `github.com/marcelocantos/pigeon` | `Register` / `Connect` → `Listener` / `Session` |
| Pairing CLI-less package | `pairing/` | Ceremony → `crypto.PairingRecord`; production driver is C via cgo |
| C peer library | `c/src/*.c` amalgamated to `dist/pigeon.c` + `dist/pigeon.h` | Crypto, pairing wire, session/stream/datagram, ngtcp2 transport |
| Swift | `Package.swift`: products `Pigeon` (CPigeon + PigeonCore) and `PigeonCore` (pure Swift) | iOS/macOS; Core avoids host-only static `.a` |
| Android/Kotlin | `android/pigeon/` AAR + JNI `libpigeon-jni` | JitPack; peer API wraps C; legacy `PigeonConn` remains |
| TypeScript | `web/` `@marcelocantos/pigeon` 0.2.0 | Browser WebTransport client |
| Protogen | `cmd/protogen` | YAML → Go/Swift/Kotlin/TS/C/TLA+/PlantUML |
| Examples | `examples/demo`, `examples/echo` | In-process demo + echo |

Relay HTTP/3 routes actually registered (`webtransport.go:202-207`):
`GET /health`, `GET /status`, `GET /pigeon`. Role is a primary-stream
greeting, not path-based `/register` / `/ws/{id}`.

### Declared layers (DESIGN.md) vs observed packages

```
L4 mux / named streams+datagrams     api.go Session/Stream/Datagram, cwire, C session
L3 path management                   protocol/session.yaml Phase 2, protocol/pathswitch.yaml
                                     (generated machines exist; Go runtime still hand-rolls)
L2 pairing                           protocol/pairing.yaml + C pigeon_pair_* + Swift driver
L1 relay bridge                      session.go hub, webtransport.go, quicserver.go
L0 crypto                            Go crypto/, C libsodium, Swift CryptoKit, Kotlin JCA, TS WebCrypto
```

Go internal import DAG (`GOWORK=off`; no cycles):

```
cmd/pigeon, cmd/pigeon-bridge, cmd/crypto-peer → pigeon (root)
examples/{demo,echo} → root, crypto, pairing [, backchannel, qr]
pairing → root, crypto, cwire, protocol
root → crypto, cwire, protocol, qr
cmd/protogen → protocol
crypto, cwire, protocol, qr, faultproxy, backchannel: leaves / infra
```

**Declared and observed that agree**

- Relay is an opaque byte-shoveler; AEAD is end-to-end (DESIGN.md threat
  model; `api.go` pairing-mode is explicitly a raw pipe).
- Two listener stacks share one `hub` (`session.go:43-48`).
- Protocol YAML is the *intended* SSoT for FSMs; wireformats.yaml for
  one-shot byte layouts; amalgamate CI checks `dist/`.
- C is the production pairing/session crypto path for Go (`pairing/cgo_pair.go`,
  `cwire.RunAcceptor`) and Kotlin JNI (`Session.kt` wraps `PigeonNative`).
- PigeonCore vs Pigeon split is real and recent (HEAD commit).

**Observed, inferred from code**

- Go production Session forks per-stream and datagram AEAD channels from
  `cwire.SessionMaterial` (`api.go:545-552`, 🎯T53) rather than one shared
  channel. `cwire/session.go:53-67` comments still describe the pre-T53
  shared-channel world.
- Pairing Go comments claim the generated executor
  (`pairing/pairing.go:337-341`); the function body hands the stream to
  `pigeon_pair_acceptor` (`pairing.go:360-375`). C `pairing.c` does not
  call `pigeon_acceptor_handle_*`.
- `InsecureSkipVerify: true` is the nil-TLS default on `Register`/`Connect`
  (`api.go:186-188`, `api.go:445-447`) and is **hardcoded** in
  `pairing.Register`/`Initiate` (`pairing.go:122-124`, `180-183`) because
  `pairing.Args` has no TLS field.

**Contradictions**

- `STABILITY.md:13` “Snapshot as of v0.22.0” vs git `v0.32.0`. Catalogue
  still lists `GET /register` and `GET /ws/{id}` (`STABILITY.md:21-22`),
  `DialRelayAcceptor`/`Conn` (no `func DialRelay*` in tree), and a `pair`
  subcommand (no such flag/subcommand in `cmd/pigeon/main.go`).
- `CLAUDE.md:44-51` documents `conn.go`, `*Conn`, `WithToken()`, `WithTLS()`
  — none exist. Register/Connect take `*RegisterArgs` / `*ConnectArgs`.
- `CLAUDE.md:90-94` places E2ECrypto and generated machines under
  `Sources/Pigeon/`; they live in `Sources/PigeonCore/`.
- `cmd/protogen/main.go:153-158` and `:292` still write
  `Sources/Pigeon/*Machine.swift` and `Sources/Pigeon/WireGen.swift`.
- `make generate` (`Makefile:113-115`) runs only `pairing.yaml` and
  `wireformats.yaml`, not `session.yaml` or `pathswitch.yaml`.
- 🎯T28 (achieved) acceptance requires TLC `PathSwitch`
  (`bullseye.yaml:736`); `Makefile:276-277` runs only `PairingCeremony` and
  `SessionMachine`.
- `webtransport.go:150-152` comment still says backends register via
  `/register` and clients via `/ws/{id}`.
- `web/package.json:11` `test:e2e:local` points at
  `src/relay.local.e2e.ts`, which is not in the tree.

**Unknown intent (owner judgment)**

- Whether Swift/Kotlin/TS **must** remain native-crypto (platform APIs) or
  should converge on C/JNI/WASM as the only AEAD implementation.
- Whether generated pairing machines are documentation/oracles or the
  production executor (today: mixed).
- Whether `hygiene.yaml` was deferred or never adopted.
- Whether `STABILITY.md` is meant to be a living catalogue or a frozen
  v0.22 snapshot (the file itself says snapshot, but it is still the
  interaction-surface document).

## Dimension vector

No scalar. Change-from-baseline is “n/a — first report under this contract”
(Fable-5 was a security audit with a different finding set).

| Dimension | State | Evidence summary | Change from baseline |
|---|---|---|---|
| Architecture topology | concern | Clear L1–L4 intent and acyclic Go DAG; generator, Swift layout, and dual SDK topologies have diverged | n/a |
| Redundancy / sources of truth | concern | YAML SSoT is real for FSMs; crypto×5 and pairing C-vs-generated-FSM can disagree; docs contradict code | n/a |
| Change amplification | concern | One YAML/crypto change should fan to 5–7 languages; generate+CI do not cover that fan-out | n/a |
| Local code quality | concern | Substantial, mostly linear Go/C; stale comments (cwire shared-channel, runAcceptor FSM) mislead | n/a |
| Correctness / verification | concern | Strong local oracles (`make bullseye`, vectors, T53/T44.1 tests, TLC two specs); CI is a subset; T55 open | n/a |
| Security / dependencies | concern | E2E story is designed; pairing TLS cannot be verified; CheckOrigin always true; no dependabot/secret-scan | n/a |
| Build / release / operations | concern | Fly deploy on master after `go test` only; JDK Cellar pin; SDK version fields ≠ git tag | n/a |
| Documentation / governance | concern | DESIGN.md is honest; STABILITY/CLAUDE/agents-guide/protogen comments are not; no hygiene.yaml | n/a |

## Findings

### ENT-001: Protogen and `make generate` no longer land on the compiled Swift tree, and CI only ratchets `dist/`

- **Priority:** P1
- **Dimensions:** Architecture topology; Redundancy; Change amplification; Build / release
- **Status:** observed fact
- **Evidence:**
  - Generated Swift lives in `Sources/PigeonCore/`
    (`PairingCeremonyMachine.swift:1-5`, `WireGen.swift` header).
  - `Package.swift:77-86` compiles `PigeonCore` from `Sources/PigeonCore`
    and `Pigeon` from `Sources/Pigeon` (four files: Exports, Session,
    Transport, Ngtcp2Transport).
  - `cmd/protogen/main.go:153-158` writes
    `Sources/Pigeon/<Name>Machine.swift`; `:292` writes
    `Sources/Pigeon/WireGen.swift`.
  - `Makefile:113-115` `generate` invokes pairing.yaml + wireformats.yaml
    only; `protocol/session.yaml` and `protocol/pathswitch.yaml` exist and
    are not in that target.
  - `.github/workflows/ci.yml:72-87` `amalgamate` job diffs **only**
    `dist/` after `make amalgamate`.
  - `docs/DESIGN.md:395-398` still documents
    `Sources/Pigeon/*Machine.swift` as the pipeline output.
- **Mechanism:** a YAML or generator fix regenerates into a path SPM does
  not compile (or does not regenerate Session/PathSwitch at all). The only
  standing ratchet (`git diff dist/`) cannot see Swift/Kotlin/TS/Go
  `_gen.go` drift. The next protocol change will be applied by hand in
  PigeonCore or silently omitted.
- **Blast radius:** every SDK consumer of generated machines/wire helpers;
  iOS PigeonCore product (HEAD’s reason for the split); C amalgamation
  indirectly if generate is skipped.
- **Counterevidence checked:** amalgamate CI *does* catch stale `dist/`
  (healthy, narrow). Generated files are committed (good). HEAD commit
  moved files but did not retarget protogen.
- **Smallest coherent remediation:** point protogen Swift outputs at
  `Sources/PigeonCore/`; add `session.yaml` and `pathswitch.yaml` to
  `make generate`; extend the CI drift job to `git diff` the full
  generated set (or `make generate && git diff --exit-code`).
- **Verification:** CI job that runs `make generate && make amalgamate`
  and fails on any dirty generated path.
- **Ratchet candidate:** CI `git diff --exit-code` over the protogen
  output list in `cmd/protogen/main.go` (not only `dist/`).

### ENT-002: GitHub CI does not run the declared one-button gate

- **Priority:** P1
- **Dimensions:** Correctness / verification; Build / release / operations
- **Status:** observed fact
- **Evidence:**
  - `CLAUDE.md:18-30` and 🎯T28 (`bullseye.yaml:730-738`) declare
    `make bullseye` as the durable green signal: gofmt, vet, build, test,
    swift, kotlin `:pigeon:test`, web tsx tests, C tests, TLC
    PairingCeremony **and** PathSwitch.
  - `.github/workflows/ci.yml` jobs: `test` (`go test ./...`, lines 9-33),
    `web-e2e` (Playwright, 35-70), `amalgamate` (72-87), `swift-test`
    (89-122), `deploy` to Fly on master after **`needs: test` only**
    (124-145).
  - No kotlin/gradle job, no TLC, no `gofmt`/`go vet`, no `test-c`, no
    `make bullseye`.
  - `Makefile:276-280` TLC step requires *two* “No error has been found”
    lines (PairingCeremony + SessionMachine) and never invokes PathSwitch,
    contradicting T28’s achieved acceptance (`bullseye.yaml:736`).
  - `web/package.json:9` `npm test` runs only `src/crypto.test.ts`;
    Makefile/bullseye run three tsx files (`Makefile:45`, `:269`).
- **Mechanism:** merge/deploy can be green while kotlin, C, TLC, or gofmt
  are red locally. Fly production (`ci.yml:124-127`) is gated on Go unit
  tests, not the multi-SDK oracle the docs advertise.
- **Blast radius:** every non-Go SDK and the formal specs; production
  relay image.
- **Counterevidence checked:** Swift *is* in CI (macos). Playwright E2E
  *is* in CI (stronger than bullseye’s tsx unit slice). Release workflow
  re-runs `go test` on linux/amd64 (`release.yml:69-71`). Local
  `make bullseye` remains a real umbrella if someone runs it.
- **Smallest coherent remediation:** make `ci.yml` invoke the same steps
  as `make bullseye` (or `make bullseye` itself) on the feasible runners;
  add PathSwitch to the TLC step or amend T28; either run kotlin on a
  runner with JDK/NDK or declare it `skipped` with `absent:` evidence
  once hygiene exists.
- **Verification:** a master push that breaks `protocol/pairing.yaml` TLC
  or kotlin tests must fail required CI.
- **Ratchet candidate:** required CI job wrapping `make bullseye` (or an
  explicit matrix that matches the Makefile step list).

### ENT-003: Five independent AEAD/HKDF stacks; canonical vectors only lock Go→C

- **Priority:** P1
- **Dimensions:** Redundancy; Change amplification; Correctness / verification; Security
- **Status:** observed fact (mechanism); inference (undetected drift probability)
- **Evidence:**
  - Go: `crypto/crypto.go:7-8, 44-51, 156-166` (stdlib AES-GCM, atomic
    `sendSeq`).
  - C: `c/src/crypto.c:7-15, 195-218` (libsodium AES-256-GCM,
    `ch->send_seq++`).
  - Swift: `Sources/PigeonCore/E2ECrypto.swift:8-14, 31-38` (CryptoKit).
  - Kotlin (JVM, still compiled): `android/.../E2ECrypto.kt:17-26`
    (JCA X25519/AES-GCM) **and** JNI `peer/Session.kt:16-18` (“All
    wire-format and crypto logic lives in the C library”).
  - TypeScript: `web/src/crypto.ts:19-21` (WebCrypto HKDF).
  - Canonical vectors: `crypto/testvectors_test.go:14-16, 158` writes
    `c/test/vectors.json`; `c/test/test_pigeon.c:768-773` is the only
    consumer (`Go->C`). Grep of `vectors.json` / confirmation `764140` in
    `*.swift`, `*.kt`, `*.ts`: **no matches**.
  - Swift/Kotlin/TS tests round-trip *internally*
    (`Tests/PigeonTests/E2ECryptoTests.swift:9-47`,
    `E2ECryptoTest.kt:21-31`, `web/src/crypto.test.ts:16-32`) without the
    frozen vector file.
  - 🎯T44.4 remains **identified** (frontier): “Cross-SDK parity and
    cascading for multi-service nodes”.
- **Mechanism:** a nonce, AAD, HKDF-info, or confirmation-code change in
  C/Go can ship while Swift/Kotlin/TS still interoperate with each other
  but not with the relay’s C/Go peers. Kotlin has two stacks (legacy JCA
  + JNI C), so a fix can be applied to the unused one.
- **Blast radius:** any cross-language session (the product’s reason to
  exist). Pairing confirmation codes and post-pair AEAD.
- **Counterevidence checked:** Go↔C vector lock is real and wired into
  `make test-c`. `pairing/cross_swift_test.go` is a live Go-acceptor /
  Swift-initiator ceremony (darwin-only). `cwire` exists specifically to
  fold Go onto C (🎯T34). DESIGN.md:414 notes AEAD wrapping is
  “hand-rolled per language”. This duplication is partly *deliberate*
  (platform crypto APIs, PigeonCore without libsodium).
- **Smallest coherent remediation:** consume `c/test/vectors.json` (or a
  generated sibling) in Swift, Kotlin, and TS unit tests; treat JNI C as
  Kotlin’s AEAD oracle and quarantine or delete JCA `E2EChannel` once
  `PigeonConn` is gone.
- **Verification:** mutating one byte of `vectors.json` ciphertext fails
  unit tests in every SDK, not only `test_pigeon`.
- **Ratchet candidate:** CI step that greps each SDK test tree for the
  vector filename or confirmation-code constant.

### ENT-004: Production pairing is a hand-rolled C JSON driver, not the generated FSM

- **Priority:** P1
- **Dimensions:** Redundancy; Correctness / verification; Change amplification
- **Status:** observed fact
- **Evidence:**
  - `pairing/pairing.go:337-341` comment: “State transitions are
    dispatched through the protogen-generated executor
    (`NewPairingCeremonyProtocolAcceptorMachine`), so the Go code path is
    provably equivalent to the Swift / Kotlin / TS / C / TLA+ outputs.”
  - `pairing/pairing.go:360-375` implementation: `cwire.RunAcceptor` →
    `pigeon_pair_acceptor`.
  - `c/src/pairing.c:4-21` is an explicit hand-rolled
    hello/welcome/reveal/confirm JSON driver; grep of `c/src/*.c` for
    `pigeon_acceptor_handle` / `pigeon_initiator_handle`: **no matches**.
  - Swift `Sources/PigeonCore/PairingCeremony.swift:4-10` *does* drive
    the generated `PairingCeremonyAcceptorMachine` /
    `PairingCeremonyInitiatorMachine`.
  - `pairing/cross_swift_test.go:31-34` still cites
    `Sources/Pigeon/PairingCeremony.swift`.
  - DESIGN.md:416-419: “The pairing ceremony is the only wire interaction
    currently flowing through protogen” — true for *codegen*, false for
    the C/Go *executor*.
- **Mechanism:** TLC and generated machines can stay green while
  `pairing.c` (the Go production path) diverges on field names, commit
  order, or abort behaviour. Swift’s generated-FSM driver and C’s
  hand-rolled driver are two sources of pairing truth that the
  darwin-only interop test only samples, not specifies.
- **Blast radius:** every pairing ceremony involving a Go or C peer
  (i.e. the relay-side backend).
- **Counterevidence checked:** C driver documents byte-for-byte JSON
  compatibility with Go `encoding/json` (`pairing.c:8-10`). T52 SAS
  commit is implemented in both C and Go crypto. Cross-language test
  exists but is skipped off-darwin (`cross_swift_test.go:46-47`).
- **Smallest coherent remediation:** either (a) make `pigeon_pair_*` step
  the generated C machines, matching the comment, or (b) rewrite the
  comments/DESIGN/TLA fidelity claim to “C driver is SSoT; generated
  machines are oracles” and add a driver↔machine conformance test.
- **Verification:** a YAML transition change (e.g. extra abort message)
  fails C `pigeon_pair_*` tests without a matching `pairing.c` edit —
  or, if C is SSoT, fails a conformance test that replays C traces
  through the generated machine.
- **Ratchet candidate:** test that the generated acceptor machine
  accepts exactly the message sequence `pairing.c` sends (and rejects
  mutations).

### ENT-005: Interaction-surface docs are a competing, stale catalogue

- **Priority:** P2
- **Dimensions:** Documentation / governance; Change amplification
- **Status:** observed fact
- **Evidence:**
  - `STABILITY.md:13` v0.22.0 snapshot; repo tag `v0.32.0`.
  - `STABILITY.md:21-22` `/register` and `/ws/{id}` vs
    `webtransport.go:205-207` `/pigeon`.
  - `STABILITY.md` DialRelay/Conn API vs no `func DialRelay*` in `*.go`.
  - `STABILITY.md:50-54` `pair` subcommand vs `cmd/pigeon/main.go:98-129`
    (version, help-agent, port, quic-port, cert, key, domain, acme,
    lan, cert-validity only).
  - `CLAUDE.md:44-51` `conn.go`, `WithToken`, `WithTLS`, `*Conn`.
  - `CLAUDE.md:90-94` Swift sources under `Sources/Pigeon/`.
  - `webtransport.go:150-152` stale route comment next to the correct
    handler.
  - `agents-guide.md:335` still names `ConnectWithArtifact` /
    `PigeonConn.connect(artifact:)`.
- **Mechanism:** integrators and agents following STABILITY/CLAUDE will
  call APIs that do not exist and miss `RegisterArgs`/`PigeonCore`.
  STABILITY churn is already high (32 commits in the log window) without
  tracking the tag.
- **Blast radius:** every new consumer and every agent session that
  loads `CLAUDE.md`.
- **Counterevidence checked:** `README.md` Register/Connect snippets
  match `api.go`. `pigeon.go:1-21` package comment is current. DESIGN.md
  is largely honest about runtime gaps.
- **Smallest coherent remediation:** retitle STABILITY with the current
  tag and delete or mark removed surfaces; fix CLAUDE.md package map in
  the same commit as the next API change (do not wait for a docs PR).
- **Verification:** a grep gate that `WithToken` / `conn.go` /
  `GET /register` do not appear as *current* API in CLAUDE.md/STABILITY
  unless behind a “removed in v0.xx” heading.
- **Ratchet candidate:** CI grep or bullseye acceptance on those strings.

### ENT-006: Session/PathSwitch YAML is generated and model-checked, then unused by the Go executor; PathSwitch TLC is not in bullseye

- **Priority:** P2
- **Dimensions:** Architecture topology; Correctness / verification
- **Status:** observed fact
- **Evidence:**
  - DESIGN.md:407-408, 424-429: `api.go` still drives per-pipe I/O;
    YAML is SSoT for protogen-emitted machines; folding onto the FSM is
    remaining work.
  - `Makefile:276-277` TLC: PairingCeremony + SessionMachine only.
  - `formal/PathSwitch.tla` + `PathSwitch.cfg` are tracked; T28
    acceptance names PathSwitch (`bullseye.yaml:736`).
  - 🎯T55 active: “Session Phase-1 security invariants are actually
    model-checked, with no tautological or dead guards”.
  - 🎯T43 identified 112 days with no progress (STUN path-tier).
- **Mechanism:** a green SessionMachine TLC run does not constrain
  `api.go`. PathSwitch can rot without failing `make bullseye`. T55
  already records the tautology risk.
- **Blast radius:** path upgrade / cutover correctness (LAN, future STUN);
  false confidence from “TLA+ says the session is safe”.
- **Counterevidence checked:** DESIGN.md discloses the gap (do not treat
  the gap as accidental). PairingCeremony TLC *is* in bullseye. This is
  sequenced work, not a silent omission of the pairing spec.
- **Smallest coherent remediation:** add PathSwitch to the bullseye TLC
  step (or reopen T28); keep T55 as the fidelity gate before claiming
  session security from TLC.
- **Verification:** `make bullseye` fails if `formal/PathSwitch.tla` does
  not model-check; a T55 oracle that the checked invariants are
  non-tautological.
- **Ratchet candidate:** Makefile TLC completed-count 2 → 3, or an
  explicit PathSwitch step.

### ENT-007: Pairing cannot present a verified TLS identity to the relay

- **Priority:** P2
- **Dimensions:** Security / dependencies; Architecture topology
- **Status:** observed fact
- **Evidence:**
  - `pairing.Args` (`pairing.go:55-59`) is `{Relay, Identity}` only.
  - `pairing.go:122-124` and `:180-183` always pass
    `&tls.Config{InsecureSkipVerify: true}`.
  - Library default is the same when `RegisterArgs.TLS` / `ConnectArgs.TLS`
    is nil (`api.go:60-61, 186-188, 445-447`), but those structs *can*
    override; pairing cannot.
  - `webtransport.go:197` `CheckOrigin: func(...) bool { return true }`.
  - `cmd/pigeon/main.go:147-149`: empty `PIGEON_TOKEN` ⇒ open register.
- **Mechanism:** the pairing ceremony is the moment a network attacker
  most wants a split pipe. SAS commit (T52) binds ephemeral keys; it
  does not bind the relay’s TLS identity or stop token capture on the
  pairing transport. Hardcoding skip-verify removes the only hook an
  integrator would use to pin the relay.
- **Blast radius:** all `pairing.Register` / `Initiate` callers; QR
  rendezvous tokens in transit.
- **Counterevidence checked:** DESIGN.md and `api.go` comments call
  skip-verify a development default. Payload confidentiality after a
  *successful, code-checked* pair still rests on AEAD. Fable-5 already
  flagged this cluster; it remains structurally true. `cmd/crypto-peer`
  uses `PIGEON_INSECURE==1` (opt-in), which is the better pattern.
- **Smallest coherent remediation:** add `TLS *tls.Config` to
  `pairing.Args` / `InitiateArgs` and thread it; keep skip-verify only
  as explicit nil default, matching Register/Connect.
- **Verification:** unit test that a pairing `Args.TLS` with
  `InsecureSkipVerify: false` and a custom `RootCAs` is the config
  passed into `pigeon.Register`.
- **Ratchet candidate:** grep that `pairing.go` contains no literal
  `InsecureSkipVerify: true` (must come from args or a named
  `DevelopmentTLS` helper).

### ENT-008: Dual live client APIs (legacy PigeonConn vs Session) plus dual Go Channel types

- **Priority:** P2
- **Dimensions:** Redundancy; Change amplification; Local code quality
- **Status:** observed fact
- **Evidence:**
  - Swift `Sources/PigeonCore/PigeonRelay.swift:50-56` deprecated
    `PigeonConn` (Network.framework, pre-T22 wire); tests still depend
    (`PigeonRelayE2ETests`).
  - Kotlin `PigeonConn.kt:19-26` deprecated pre-T22 wire; peer
    `Session.kt` is JNI.
  - Go `crypto.Channel` (`crypto.go:156-166`) and `cwire.Channel`
    (`cwire/session.go:53-71`) both implement the same wire; comments
    still say production Session stores `Session.channel` — the field is
    now per-stream (`api.go:545-552, 932-937`).
- **Mechanism:** a wire-format or nonce fix must be considered on the
  deprecated path or the deprecated path silently speaks a different
  protocol (T27 already quarantined some tests). Reviewers following
  cwire comments will mis-model locking/sharing.
- **Blast radius:** remaining PigeonConn callers; anyone reading cwire
  comments while changing Session AEAD.
- **Counterevidence checked:** deprecation attributes are explicit.
  T53 tests (`cwire/t53_channel_isolation_test.go:13-17`) document the
  new isolation invariant. Removing PigeonConn is blocked on ngtcp2
  production transport (comments say so).
- **Smallest coherent remediation:** delete or `//go:build legacy` the
  PigeonConn types once E2E tests use Session; rewrite cwire Channel
  comments to T53.
- **Verification:** grep for `class PigeonConn` / `Session.channel`
  comment returns only history, not live API.
- **Ratchet candidate:** file-rule or test that PigeonRelayE2ETests
  import Session, not PigeonConn (when ready).

### ENT-009: Published SDK version fields are unrelated to git tags

- **Priority:** P2
- **Dimensions:** Build / release / operations; Documentation
- **Status:** observed fact
- **Evidence:**
  - Git tag `v0.32.0`.
  - `android/pigeon/build.gradle.kts:13` `version = "0.6.0"`.
  - `web/package.json:3` `"version": "0.2.0"`.
  - `STABILITY.md:13` v0.22.0.
  - JitPack consumes git tags (`CLAUDE.md:102-103`) while Gradle
    `version` is a different number.
- **Mechanism:** a consumer comparing npm/Gradle coordinates to GitHub
  releases cannot tell which tree they have. Release automation that
  reads `package.json` / Gradle will mint the wrong name.
- **Blast radius:** JitPack/npm consumers; Homebrew is tag-driven
  (`release.yml:106`) so the relay binary is consistent.
- **Counterevidence checked:** Go module is tag-versioned (no version
  in `go.mod` beyond the language). Pre-1.0, independent SDK versions
  can be deliberate — then STABILITY/README must say so. They do not.
- **Smallest coherent remediation:** either drive all SDK version
  fields from the git tag in `/release`, or document three independent
  version lines in STABILITY.
- **Verification:** `web/package.json` version, Gradle `version`, and
  `git describe` match after a release (or a documented mapping table
  is grepped in STABILITY).
- **Ratchet candidate:** release skill check that those three strings
  equal `vX.Y.0`.

### ENT-010: Hygiene posture is not declared; SDLC scanners are absent

- **Priority:** P3
- **Dimensions:** Documentation / governance; Security; Build
- **Status:** observed fact
- **Evidence:**
  - No `hygiene.yaml`.
  - `hygiene_check.py` from repo root:
    `FileNotFoundError: .../pigeon/hygiene.yaml`.
  - `.github/` contains only `workflows/ci.yml` and `release.yml` — no
    CODEOWNERS, dependabot, codeql, or secret-scan workflow.
  - No `.golangci.yml`.
- **Mechanism:** fleet “which repos lack X?” cannot see pigeon; CI
  dependency refresh is manual; secret scanning is not a standing gate.
- **Blast radius:** governance/fleet aggregation; supply chain of
  `go.mod` and npm lockfile.
- **Counterevidence checked:** LICENSE is Apache-2.0; README exists;
  `go.mod`/`go.sum` and `web/package-lock.json` exist; vendored C deps
  are submodules (aligned with `cpp.md`). Local `make bullseye` is
  stronger than many repos’ CI.
- **Smallest coherent remediation:** add `hygiene.yaml` in a later
  hygiene-init (not this audit) reflecting *actual* floors, with
  kotlin/TLC gaps as `planned`/`skipped` if CI does not run them.
- **Verification:** `hygiene_check.py` exit 0 against declared floors.
- **Ratchet candidate:** `hygiene.yaml` itself, once authored.

### ENT-011: Web package scripts diverge from Makefile and point at a missing file

- **Priority:** P3
- **Dimensions:** Correctness / verification; Documentation
- **Status:** observed fact
- **Evidence:**
  - `web/package.json:8-11`: `test` = crypto.test.ts only;
    `test:e2e:local` = `src/relay.local.e2e.ts`.
  - `web/src/relay.local.e2e.ts` is **absent**; `web/src/test-helpers/GoRelayProcess.ts`
    remains; Playwright lives in `web/e2e/relay.spec.ts`.
  - T10.4 context still describes `relay.local.e2e.ts` (`bullseye.yaml`
    T10.4).
- **Mechanism:** `npm test` in `web/` is a weaker oracle than
  `make test-web`; `npm run test:e2e:local` fails on a missing file.
- **Blast radius:** contributors following package.json instead of
  Makefile.
- **Counterevidence checked:** CI uses `npm run test:e2e` (Playwright),
  which exists. Makefile lists the three tsx files explicitly.
- **Smallest coherent remediation:** align `package.json` scripts with
  Makefile; delete or restore `test:e2e:local`.
- **Verification:** `npm test` in `web/` runs the same three files as
  `make test-web`; `npm run test:e2e:local` either exists or is removed.
- **Ratchet candidate:** `command:` evidence in a future hygiene.yaml.

### ENT-012: Residue files and pins that amplify confusion

- **Priority:** P3
- **Dimensions:** Local code quality; Build
- **Status:** observed fact
- **Evidence:**
  - Tracked `formal/Session.old` and `formal/Session_Transport.old`
    (headers: auto-generated from “internal/protocol/”, a path that
    does not exist).
  - `Makefile:4` `JDK21 ?= /opt/homebrew/Cellar/openjdk@21/21.0.11/...`
    — host-specific Cellar patch version.
  - `cwire/cwire.go:7-12` still says the package covers “only the pure
    wire-format primitives”; it now owns connect, pairing, session,
    listener.
  - `c/vendor/build/` is **not** gitignored (only `build-android/`);
    local static libs can appear as untracked noise (status was clean
    at audit start).
- **Mechanism:** old TLA modules and stale package comments are extra
  sources of “how does this work?”. The JDK pin breaks `make test-kotlin`
  on any machine without that exact Cellar path.
- **Blast radius:** kotlin local tests; newcomers reading cwire docs;
  TLC users opening `Session.old`.
- **Counterevidence checked:** JDK pin exists *on this host* (path
  present). `.old` files are not referenced by `Makefile` TLC.
- **Smallest coherent remediation:** delete `formal/*.old`; gitignore
  `c/vendor/build/`; find JDK via `JAVA_HOME`/`/usr/libexec/java_home`;
  rewrite cwire’s package comment.
- **Verification:** `make test-kotlin` with only `JAVA_HOME` set;
  `git ls-files 'formal/*.old'` empty.
- **Ratchet candidate:** `.gitignore` line + Makefile JDK discovery.

### ENT-013: `faultproxy` uses the functional-options pattern

- **Priority:** P3
- **Dimensions:** Local code quality
- **Status:** observed fact
- **Evidence:** `faultproxy/faultproxy.go:75-89` `type Option func(*Profile)`
  plus `WithLatency` / `WithPacketLoss` / `WithReorder`. Language rule
  (`go.md`): never `...Option`.
- **Mechanism:** a test helper that trains the forbidden API into this
  repo; copy-paste risk into product packages.
- **Blast radius:** test-only package (`STABILITY.md` marks it testing
  only).
- **Counterevidence checked:** no product API uses `...Option` (grep
  `type Option func` → this file only). Register/Connect already use
  struct-pointer args.
- **Smallest coherent remediation:** `New(target string, profile *Profile)`.
- **Verification:** grep `type Option func` is empty.
- **Ratchet candidate:** grep in CI or a sawmill convention check.

## Redundancy and competing-source-of-truth inventory

| Concept | Live authorities | Can they disagree? | Owner if converged |
|---|---|---|---|
| Pairing FSM | `protocol/pairing.yaml` → generated machines; C `pairing.c`; Swift `PairingCeremony.swift`; comments claim Go generated executor | **Yes** (ENT-004) | Pick C driver *or* generated machine |
| Session / PathSwitch FSM | YAML + generated SDK machines + TLA; `api.go` hand-rolled I/O | **Yes** (ENT-006) | YAML, once runtime dispatches it |
| Wire bytes (greeting, stream header, datagram plaintext) | `protocol/wireformats.yaml` + protogen outputs | Intended no; CI only checks C amalgamation (ENT-001) | YAML |
| AEAD / HKDF / SAS | Go `crypto/`, C libsodium, Swift CryptoKit, Kotlin JCA, TS WebCrypto | **Yes** (ENT-003) | vectors.json in every SDK; C for cgo/JNI |
| Relay HTTP routes | Code `/pigeon`; STABILITY `/register` `/ws/{id}`; webtransport comment | **Yes** (ENT-005) | `webtransport.go` handlers |
| Public Go API | `api.go` RegisterArgs; CLAUDE WithToken/Conn; STABILITY DialRelay | **Yes** (ENT-005) | `api.go` + README |
| Swift generated sources | Protogen `Sources/Pigeon/`; package `Sources/PigeonCore/` | **Yes** (ENT-001) | PigeonCore |
| SDK version | git tag v0.32.0; Gradle 0.6.0; npm 0.2.0; STABILITY v0.22.0 | **Yes** (ENT-009) | git tag |
| “Tests are green” | `make bullseye` vs GitHub `ci.yml` | **Yes** (ENT-002) | bullseye, required in CI |
| Legacy vs Session client | PigeonConn + Session in Swift/Kotlin | **Yes** (ENT-008) | Session |

**Deliberate duplication (tried to invalidate, kept):**

- Native crypto APIs per platform so PigeonCore can link on iOS without
  host `libsodium.a` (`Package.swift:35-39`). Keep, but pin with shared
  vectors (ENT-003).
- WebTransport vs raw QUIC listeners sharing one hub — two network
  shapes, one routing table (`session.go:43-48`).
- Amalgamated `dist/` plus per-file `c/src` — distribution vs
  development; amalgamate CI exists.

## Healthy structure worth retaining

- **Acyclic Go module graph** (`GOWORK=off go list`); pairing depends
  downward onto cwire/C rather than a cycle through proto runtime.
- **Struct-pointer args** on `Register`/`Connect` (`api.go:37-78`) —
  matches `go.md`; do not regress to `WithToken`.
- **C as the Go/Kotlin production peer library** (`cwire`, JNI) is the
  right fold (🎯T34); finish it rather than adding a sixth crypto.
- **PigeonCore vs Pigeon** split (HEAD) — iOS can import pairing/crypto
  without ngtcp2 static libs. Retarget protogen at Core (ENT-001), do
  not undo the split.
- **T53 per-stream / datagram key fork** (`api.go:545-552`,
  `cwire/t53_channel_isolation_test.go`) — closes the shared-strict-
  channel contradiction Fable-5 recorded. Keep the tests; fix the
  stale cwire comments.
- **T52 SAS commit** in Go (`crypto.go:110-117`) and C (`pairing.c:12-17`).
- **T44.1 per-session nonce** (`cwire/persession_test.go:13-21`) and
  cwire `Channel.mu` (`cwire/session.go:61-70, 107-110`).
- **Hub ownership tests** (`hub_ownership_test.go:40`) for Fable-5
  register/unregister integrity.
- **Amalgamate drift job** on `dist/` (`ci.yml:72-87`) — extend, do not
  drop.
- **Playwright WebTransport E2E** in CI (`ci.yml:35-70`, `web/e2e/`) —
  owner-visible browser path; stronger than bullseye’s tsx units.
- **DESIGN.md** as an honest architecture note (runtime-vs-YAML gap is
  written down). Prefer fixing STABILITY/CLAUDE over adding more prose.
- **Vendored C deps as submodules** under `c/vendor/github.com/...`
  (matches `cpp.md`); libsodium build cached in CI.
- **Apache-2.0** `LICENSE`; `NOTICES` for bundled attribution.
- **Formal PairingCeremony + SessionMachine** actually invoked from
  `make bullseye` (incomplete vs T28, but not vapor).

## Hygiene posture

**Hygiene posture not declared.** There is no root `hygiene.yaml`.

Validator invocation (mandatory even though undeclared):

```
$ /Users/marcelo/.claude/skills/hygiene/hygiene_check.py
FileNotFoundError: [Errno 2] No such file or directory:
  '/Users/marcelo/work/github.com/marcelocantos/pigeon/hygiene.yaml'
```

No per-dimension held tiers or floors. This audit did **not** initialize
`hygiene.yaml`.

Observed reality that a future init would have to declare (not floors
until validated):

| Dim | What exists | Gap |
|---|---|---|
| correctness | `go test ./...` in CI; local `make bullseye`; Playwright E2E; C tests locally; TLC two specs locally | kotlin/TLC/C/gofmt not in CI; PathSwitch TLC not in bullseye |
| security | E2E crypto tests; hub ownership tests; no secret-scan/CodeQL/dependabot | pairing TLS hardcoded skip-verify |
| quality | gofmt in bullseye only | no golangci-lint |
| docs | README, DESIGN, LICENSE | STABILITY/CLAUDE drift; no hygiene.yaml |
| release | `release.yml` on published GitHub release; Fly deploy on master | deploy gated on Go tests only; comment still says `gh release create` |
| build | Dockerfile multi-stage; amalgamate check | JDK Cellar pin; generate incomplete |

Entropy findings suitable for later hygiene items: ENT-001 (generate
diff), ENT-002 (bullseye in CI), ENT-003 (vectors in every SDK),
ENT-005 (doc grep), ENT-007 (no hardcoded skip-verify in pairing).

## Oracle coverage and residue

| Load-bearing property | Decision path |
|---|---|
| Go packages test | CI `go test ./...` (shipped CI) + bullseye `go test -short` (local) |
| Swift unit + package tests | CI `swift test`; bullseye same |
| Kotlin unit tests | **local bullseye only** — not CI |
| C amalgamated tests | **local** `test-c` / bullseye `test-c-only` — not CI |
| C ASan/UBSan, Go race | **local** `make bullseye-strict` — not CI |
| Go↔C crypto vectors | C `test_cross_language_vectors` (local C suite) |
| Swift/Kotlin/TS crypto vs vectors | **nothing** (ENT-003) |
| Go↔Swift pairing interop | `pairing/cross_swift_test.go` (darwin; default `go test`) |
| C SDK live interop | `go test -tags csdke2e` — bullseye skips if ngtcp2 not built |
| Browser WebTransport E2E | CI Playwright (`web/e2e`) |
| PairingCeremony TLC | local bullseye only |
| SessionMachine TLC | local bullseye only |
| PathSwitch TLC | **spec present, not in bullseye or CI** |
| Generated code freshness (dist) | CI amalgamate |
| Generated code freshness (Swift/Kotlin/TS/Go machines) | **nothing** (ENT-001) |
| Relay hub ownership | `hub_ownership_test.go` (CI via `go test`) |
| T53 channel isolation | `cwire/t53_channel_isolation_test.go` (CI via `go test`) |
| T44.1 per-session keys | `cwire/persession_test.go` (CI via `go test`) |
| Live Fly relay E2E | `make e2e-live` requires `PIGEON_TOKEN` — **not run** |
| Session YAML fidelity vs api.go | **accepted gap** (DESIGN.md); T55 open |
| Hygiene floors | **undeclared** |

**Failed / skipped checks in this audit:** `make bullseye` not executed;
full `gofmt -l .` not waited as CI; jscpd not available; live relay not
contacted; hygiene validator cannot succeed without a yaml.

**Owner residue (intent only — not mechanical homework):**

1. Is C the unique production crypto/pairing executor, with generated
   machines as oracles — or must every SDK step the generated FSM?
2. Must PigeonCore keep native CryptoKit, or is a future iOS C/libsodium
   vendor build the convergence point?
3. Is STABILITY.md a frozen v0.22 museum or a living 1.0 catalogue?
4. Accept CI without kotlin/TLC, or make them required (T28 already
   claimed the latter)?

## Remediation sequence

1. **Repair the generate/CI seam (ENT-001, ENT-002).** Retarget protogen
   Swift to `Sources/PigeonCore`; put all four YAML specs in
   `make generate`; CI `git diff` that set; run (or subset-match)
   `make bullseye` in CI; add PathSwitch to TLC or reopen T28.
2. **Converge pairing truth (ENT-004) and crypto vectors (ENT-003).**
   Decide C-driver-vs-generated-FSM; lock `vectors.json` into Swift,
   Kotlin, and TS tests; stop treating JCA `E2EChannel` as a second
   Kotlin crypto.
3. **Close doc and TLS holes that amplify the wrong API (ENT-005,
   ENT-007, ENT-009).** Refresh STABILITY to v0.32 surfaces; fix
   CLAUDE.md; thread TLS through pairing.Args; align SDK version
   fields or document them.
4. **Remove residue after consumers are gone (ENT-008, ENT-011,
   ENT-012, ENT-013).** PigeonConn, missing npm script, `formal/*.old`,
   Cellar JDK pin, faultproxy options, stale comments.
5. **Ratchet** generate-diff, bullseye-in-CI, and vector consumption.
   Author `hygiene.yaml` only when those commands exist — do not
   declare floors the CI does not hold (ENT-010).
6. **Re-run this audit** on the same finding IDs and the same
   generate/CI/vector definitions.

Do not start with a sixth SDK or a Session-machine rewrite (ENT-006)
until the generator and CI actually protect the YAML that already
exists.

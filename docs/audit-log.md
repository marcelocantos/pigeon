# Audit Log

Chronological record of audits, releases, documentation passes, and other
maintenance activities. Append-only — newest entries at the bottom.

## 2026-03-22 — /open-source tern v0.1.0

- **Commit**: `6782a9c`
- **Outcome**: Open-sourced tern. Migrated all library code from jevon (crypto, protocol framework, QR helper, protogen tool, Swift package). Audit: 19 findings (T2.1–T2.19) all addressed. Docs: README with integration examples and pairing flow, CLAUDE.md, agents-guide.md (wired into --help-agent), STABILITY.md, NOTICES, pairing ceremony SVG diagram. Renamed 'jevond' actor to 'server' in protocol spec and all generated files. Released v0.1.0 (darwin-arm64, linux-amd64, linux-arm64). Homebrew formula published to marcelocantos/homebrew-tap. CI release workflow configured.
- **Deferred**:
  - Protocol framework `Example_test.go` (Priority 4)
  - Swift confirmation code documentation (Priority 4 — depends on Swift getting DeriveConfirmationCode)

## 2026-03-23 — /release v0.3.0

- **Commit**: `91426ec`
- **Outcome**: Released v0.3.0. Major changes: WebSocket replaced with WebTransport (QUIC/HTTP3), datagram support, Let's Encrypt via certmagic, TypeScript client + codegen, DeriveConfirmationCode on all 4 platforms, CI on push/PR, benchmarks, stress tests. Five audit passes (48 findings total, all resolved). Deployed to tern.fly.dev with dedicated IPv4 + Let's Encrypt.
- **Deferred**:
  - Protocol framework `Example_test.go`
  - LAN upgrade re-implementation on WebTransport (🎯T5)
  - TLA+ model for cutover protocol (🎯T9)

## 2026-03-24 — /release v0.4.0

- **Commit**: `1f193cf`
- **Outcome**: Released v0.4.0. Raw QUIC protocol for native clients (ALPN "tern", port 4433) alongside WebTransport (port 443, browsers). Swift relay client (TernRelay via Network.framework), Kotlin relay client (ternrelay with QuicTransport interface), tern-bridge for cross-language E2E. OpenStream on Conn. Makefile. E2E tests on all 4 platforms (local + live). Cert fallback + Fly volume. 128-bit instance IDs, timing-safe token auth.
- **Deferred**:
  - Browser WebTransport E2E (blocked on Let's Encrypt rate limit, resolves ~2026-03-24 20:00 UTC)
  - LAN upgrade re-implementation on QUIC (🎯T5)
  - TLA+ model for cutover protocol (🎯T10)
  - Channel API (streaming + datagram channels)

## 2026-04-02 — /release v0.11.0

- **Commit**: `35f9372`
- **Outcome**: Released v0.11.0. Fly.io auto-start (TCP wake trigger), transparent wakeRelay in all 4 client libs, LANReady() channel, encrypt+write atomicity fix.

## 2026-03-30 — /release v0.10.0

- **Commit**: `f91e752`
- **Outcome**: Released v0.10.0. LAN upgrade (LANServer, Config.LAN), Config struct replaces options pattern, --lan CLI flag. 13 LAN tests. Homebrew formula updated.

## 2026-03-30 — /release v0.9.0

- **Commit**: `cd8c35b`
- **Outcome**: Released v0.9.0. Transparent large datagram fragmentation/reassembly folded into SendDatagram/RecvDatagram. 1-byte framing prefix. Homebrew formula updated.

## 2026-03-30 — /release v0.8.0

- **Commit**: `e0d6555`
- **Outcome**: Released v0.8.0. Channel API (streaming + datagram), faultproxy package, CI auto-deploy, WebTransport fixes, test coverage 89%/92%/98%/94%. Homebrew formula updated.

## 2026-03-25 — /release v0.7.0

- **Commit**: `0e2fab0`
- **Outcome**: Released v0.7.0. Persistent device pairing: `WithInstanceID` for stable relay identity, `PairingRecord` on all 4 platforms for save/restore of pairing state across reboots and network changes.

## 2026-03-25 — /release v0.6.0

- **Commit**: TBD
- **Outcome**: Released v0.6.0. 24 audit findings fixed (2 high, 7 medium, 15 low): Swift readExactly accumulation, goroutine leak in datagram relay, write deadline race, graceful shutdown, self-signed cert random serial, protogen output paths, maxMessageSize alignment, tern-bridge secure by default, datagram mode tests on Swift/Kotlin, generateNonce/generateSecret on all 4 platforms.

## 2026-03-24 — /release v0.5.0

- **Commit**: `70b55e6`
- **Outcome**: Released v0.5.0. Renamed Swift and Kotlin packages from TernCrypto/TernRelay to just Tern (single package per platform). Added convergence targets T12-T17.

## 2026-04-06 — /release v0.12.0

- **Commit**: `4ba3ca0`
- **Outcome**: Released v0.12.0 (darwin-arm64, linux-amd64, linux-arm64). Major release: machine-driven executor (🎯T18), hierarchical state machines (🎯T19), TLA+ rewrite with channel elimination, unified session protocol, wire constants, session protocol design doc. Homebrew formula updated.

## 2026-04-06 — /release v0.13.0

- **Commit**: `7ad9b9c`
- **Outcome**: Released v0.13.0. Project renamed from tern to pigeon. GitHub repo, Go module, Swift/Kotlin/TypeScript packages, ALPN protocol, env vars, Fly.io app, all documentation updated. Homebrew formula updated to pigeon.

## 2026-04-07 — /release v0.14.0

- **Commit**: `56b9224`
- **Outcome**: Released v0.14.0. Fly.io app migrated from tern to carrier-pigeon.fly.dev. Homebrew formula updated.
## 2026-04-07 — /release v0.15.0

- **Commit**: `7bed6ca`
- **Outcome**: Released v0.15.0 (darwin-arm64, linux-amd64, linux-arm64). Codegen namespace collisions fixed across all four generators (Go, Swift, Kotlin, TypeScript) — multiple protocols now coexist safely. Complete tern→pigeon rename (zero stale references). Swift E2E relay tests added (6 tests via XCTest). Flaky TestChaosMultiPair fixed (faultproxy reset bug, QUIC keepalives, CI UDP buffer sizing). Homebrew formula updated.

## 2026-04-12 — /release v0.16.0

- **Commit**: `fb01463`
- **Outcome**: Released v0.16.0 (darwin-arm64, linux-amd64, linux-arm64). Pure C client library added — zero-allocation struct-based API, distributed as amalgamated pigeon.h/pigeon.c pair. C code generator (cgen.go) added to protogen. 15 C tests including cross-language crypto vector validation (Go→C). CI amalgamation staleness check. Homebrew formula updated.

## 2026-04-19 — /release v0.17.0

- **Commit**: `af71c04`
- **Outcome**: Released v0.17.0 (darwin-arm64, linux-amd64, linux-arm64). Major work: multi-client relay + one-time token pairing + credential API (🎯T15); cert-hash in LAN offer for browser direct connections (🎯T6.2); ngtcp2-based QUIC transport for the C client library (🎯T11.4); E2E encryption wired into C send/recv (🎯T11.2); PairingRecord fixed-schema serialisation (🎯T11.3); pairing ceremony decomposed into pairing + auth sub-machines (🎯T14) with composable sub-machine support across all five codegen targets (Go, Swift, Kotlin, TypeScript, C) plus TLA+; drain-window cutover protocol with TLA+ cutover model (🎯T2.1, 🎯T2.2, 🎯T3); browser WebTransport E2E tests now runnable unattended (🎯T7); `make deploy` target for Fly.io (🎯T9); STUN and Bluetooth proximity investigations (🎯T4, 🎯T5); cascading relay architecture exploration doc (🎯T12). C codegen naming fix to match composed-actor types (🎯T11.1). 🎯T8 (gomobile) dropped as speculative.
- **Breaking changes**: C state-enum names refactored for composed actors (`pigeon_server_pairing_*` / `pigeon_server_auth_*` / `pigeon_app_*` — was `pigeon_ios_*` / `pigeon_cli_*`). Swift generated machine class names refactored to `PairingCeremony*` prefix with per-sub-machine + Composite types. STABILITY.md catalogue updated; settling clock reset.
- **Deferred**:
  - 🎯T16 — Swift composed-actor sub-machines have unit-test coverage. `Tests/PigeonTests/PairingCeremonyMachineTests.swift` still references the pre-T14 monolithic types and fails to compile. Not caught by CI (no swift test job in ci.yml).
  - 🎯T1.1 — Jevon imports pigeon's packages (verified externally on Jevon side).

## 2026-04-20 — /release v0.18.0

- **Commit**: `d47aa2a`
- **Outcome**: Released v0.18.0 (darwin-arm64, linux-amd64, linux-arm64). Test/CI-only release: 🎯T16 Swift `PairingCeremonyMachineTests` rewritten for the composed-actor API (PR #11); 🎯T17 `testCrossLanguageConfirmationCode` fixed on macos-15 CI runner by printing crypto-peer's instance ID to stdout instead of stderr, so quic-go's UDP-buffer warning can't contaminate the first stderr line (PR #12). `--skip testCrossLanguageConfirmationCode` removed from ci.yml. No public API changes. No breaking changes — settling clock from v0.17.0 continues to tick.
- **Deferred**:
  - 🎯T1.1 — Jevon imports pigeon's packages (verified externally on Jevon side).
- **Known issues**:
  - `ci.yml` `Deploy to Fly.io` job fails on master with `Error: unauthorized` — the `FLY_API_TOKEN` secret has expired. Orthogonal to release artifacts; does not gate `release.yml`. Rotate when convenient.

## 2026-04-25 — /release v0.19.0

- **Commit**: `694e7c0`
- **Outcome**: Released v0.19.0 (darwin-arm64, linux-amd64, linux-arm64). Adds `PairingArtifact` (persistable+expirable envelope around `crypto.PairingRecord`) + `CredentialStore` interface with platform-specific reference implementations (Keychain on iOS/macOS, file on JVM/desktop) + `PairingHost` server-side artifact minter with configurable TTL (default 30 days). All three SDKs (Go, Swift, Kotlin) now have wire-compatible canonical JSON encoding (snake_case) and a single-line base64url text encoding for the artifact, suitable for QR-payload transport, xcrun-injected deploys, launch arguments, environment variables, and pasteboard transport. Adds one-call reconnect helpers — `ConnectWithArtifact` (Go), `PigeonConn.connect(artifact:)` (Swift), `connectWithArtifact` (Kotlin) — that check expiry up front and route through typed expiry errors (`ErrPairingExpired` / `PairingError.expired` / `PairingExpiredException`). Adds `pigeon pair` CLI subcommand for deploy-script use, replacing the standalone `cmd/pigeon-pair` binary. Adds `bullseye` Makefile target for standing-invariant checks.
- **Breaking changes**: Swift `PairingRecord` JSON wire format switched from camelCase to snake_case for cross-SDK interoperability — affects any consumer that had previously serialised PairingRecord JSON in Swift. The Swift type's API surface (property names) is unchanged. Settling clock reset to 2026-04-25; earliest 1.0 eligibility 2026-06-25.
- **Deferred**:
  - ~~🎯T1.3 — End-to-end re-pair lifecycle guide.~~ Shipped in #18 the same day as `docs/pairing-lifecycle.md` (covering both QR-scan and developer-deploy via xcrun delivery channels with side-by-side Go/Swift/Kotlin samples).
  - 🎯T19 — Kotlin `PairingCeremonyMachineTest` is regenerated and compiles. Orphan test references state-machine symbols renamed when the generator split sub-machines; excluded from the test source set in `build.gradle.kts` until rewritten.
  - 🎯T20 — `PigeonConnE2ETest` `cross-language confirmation code via relay` is reliable. Live-relay handshake test intermittently times out; pre-existing but only newly visible because compileTestKotlin couldn't run before the 🎯T19 exclude.
- **Known issues**:
  - `ci.yml` `Deploy to Fly.io` job continues to fail on master with `FLY_API_TOKEN` expired. Carried over from v0.18.0; orthogonal to release artifacts.

## 2026-04-26 — /release v0.20.0

- **Commit**: `53dc3e2`
- **Outcome**: Released v0.20.0 (darwin-arm64, linux-amd64, linux-arm64). Documentation + test-quality release. 🎯T1.3 `docs/pairing-lifecycle.md` end-to-end re-pair lifecycle guide shipped on the same day as v0.19.0 (#18). 🎯T20 `PigeonConnE2ETest` cross-language confirmation code stabilised — instance ID now read from crypto-peer's stdout (was stderr) and stderr drained on a background thread, eliminating intermittent `recv pubkey: context deadline exceeded` failures (#20). 🎯T19 Kotlin `PairingCeremonyMachineTest` rewritten as five test classes mirroring the v0.17 sub-machine split (server pairing/auth, ios pairing/auth, cli) and the `compileTestKotlin` source-set exclude removed (#21). No public API changes. No breaking changes — settling clock from v0.19.0 continues to tick.
- **Deferred**: (none)
- **Known issues**:
  - `ci.yml` `Deploy to Fly.io` job continues to fail on master with `FLY_API_TOKEN` expired. Carried over from v0.18.0; orthogonal to release artifacts.

## 2026-04-30 — multi-channel pigeon API (🎯T22, 🎯T22.5–T22.8)

- **Commit**: working tree dirty — `0f0d87c`
- **Outcome**: Landed the redesigned multi-channel pigeon API. New surface: `pigeon.Register` returns a `*Listener` over a single relay registration, `Listener.Accept` yields one `*Session` per connecting paired client. `Session.OpenStream(ctx, name)` opens a fresh QUIC stream per named channel with a length-prefixed `[varint name-len][name]` first-message handshake; the relay forwards every QUIC stream end-to-end opaquely. `Session.Datagram(name)` returns a pre-declared `*Datagram` whose wire format is `AEAD([varint channel-id][payload])` — channel-ids inside the AEAD envelope. Multi-client routing on the backend is done by a 4-byte `clientTag` the relay prepends to every stream/datagram it forwards to backend (mux mode); the legacy `register`-handshake path stays 1:1 (pair mode) so the pairing ceremony in `pairing/` keeps working unchanged. `crypto.Channel` is concurrency-safe (atomic monotonic seq, mutex-guarded Decrypt) and is now exercised by `crypto/channel_concurrency_test.go`. Three new e2e tests (`e2e_test.go`) drive the in-process raw-QUIC relay: single-client chat, multi-stream + multi-datagram in one Session, and two concurrent clients sharing one backend without cross-talk. `examples/echo/{backend,client}/main.go` rewritten to demonstrate all four channels (chat, control, ping, metric). Retired 🎯T22, 🎯T22.5, 🎯T22.6, 🎯T22.7, 🎯T22.8.
- **Deferred**:
  - Per-Session state-machine instance (lifting the executor inside Session). Not a precondition for the rich demo; tracked separately.
  - 🎯T23 (macOS Keychain identity) — explicitly deferred per scope.
  - Backend-initiated streams (Session.OpenStream from the backend side) work by the same tagged protocol; demo only exercises client-initiated.
- **Known issues**:
  - `ci.yml` `Deploy to Fly.io` job continues to fail on master with `FLY_API_TOKEN` expired. Carried over from v0.18.0; orthogonal to release artifacts.

## 2026-04-30 — architectural pivot: one C peer library; vendor build live (T32 + T33 step + T31 + T34 starter)

- **Commit**: `882a58b` (in-progress; multiple commits since `0e14b44`)
- **Outcome**: Pivot of the cross-language SDK strategy after the user locked five architectural decisions in dialogue:
  - **D1A**: Swift binds the vendored ngtcp2 via a SwiftPM C-target.
  - **D2a**: Kotlin binds vendored ngtcp2 via JNI.
  - **D3**: TS becomes browser-only via the native `WebTransport` API; Node.js reserved for a future, possibly-different SDK.
  - **D4 (the big one)**: ONE peer-library implementation, in C (`libpigeon.a`). Every other language is a thin idiomatic wrapper. The relay (`cmd/pigeon`, Go + quic-go) stays untouched because it's server-side.
  - **D5 (extension of D4)**: even the Go peer library will be folded — `pigeon.Register/Connect/Session/Stream/Datagram` reimplemented as a cgo wrapper over libpigeon. Only the relay binary stays pure Go.

  The session's deliverables, in commit order:
  - **🎯T31 retired**: Browser TS rebuilt on `WebTransport`. Dropped Node-side `relay.e2e.ts` / `relay.local.e2e.ts`. New `web/src/relay.test.ts` adds 26 wire-format unit tests pinning the same byte vectors as the C side.
  - **🎯T32 step 1** (commit `3ab14f4`): post-T22 wire-format helpers in C (`pigeon_uvarint_*`, `pigeon_encode_stream_header`, `pigeon_decode_{backend,client}_stream_header`, `pigeon_encode_datagram`, `pigeon_decode_datagram`), plus `wire_vectors_test.go` locking Go-emitted bytes against the C-side hardcoded vectors.
  - **🎯T32 step 2** (commit `acd3225`): multi-channel C session API (`pigeon_session`, `pigeon_stream`, `pigeon_datagram`), `pigeon_transport` vtable extended with optional multi-stream callbacks (`open_stream`, `accept_stream`, `send_on_stream`, `recv_on_stream`, `close_stream`), in-process loopback transport for tests. 25/25 C tests pass.
  - **🎯T32 step 3** (commit `e057765`): ngtcp2 transport gains multi-stream support — slot 0 stays the legacy primary, slots 1..16 are dynamic. Wired `stream_open_cb` / `stream_close_cb` ngtcp2 callbacks; new vtable callbacks invoke `ngtcp2_conn_open_bidi_stream`, drain an accept queue with a QUIC event-loop pump, and route per-stream recv data to per-slot ringbufs. Vendor submodules initialised (`git submodule update --init --recursive`) and quictls/openssl + ngtcp2 + ngtcp2_crypto_quictls built (the build.sh's example target failed for a dynamic-link reason but the static libs we need are in place; fix or split the script later). 6/6 ngtcp2 unit tests pass against the rebuilt vendor libs.
  - **🎯T34 starter** (commit `7633280`): `cwire/` cgo bridge from Go to libpigeon's wire helpers. Tests prove byte-by-byte equivalence between Go's native encoder, the C encoder via cgo, and the hardcoded reference vectors. Foundation for the full Go-as-cgo-wrapper.
  - **🎯T33 step** (commit `882a58b`): cwire bridge gains datagram-framing helpers (`EncodeDatagram` / `DecodeDatagram`) so the cross-language byte-parity check now covers the full post-T22 wire across Go-native, C-via-cgo, hardcoded vectors, and (via wire_vectors_test.go ↔ test_pigeon.c ↔ relay.test.ts) the TS browser side.
  - **`make bullseye` parallelised** (commit `7231dce`): SDK suites + TLC fan out concurrently with per-step status files. Wall-clock dropped from ~28s sequential to ~10s parallel on M4 Max. Added the C suite (`make test-c`) to the standing invariants. Added `bullseye-prereq` to fix a gofmt-vs-codegen race introduced by the parallel fan-out.
  - **In flight**: 🎯T29 Swift wrapper, 🎯T30 Kotlin JNI wrapper, 🎯T34 full Go cgo wrapper — running in parallel sub-agents at session end.
- **Deferred**:
  - Live multi-stream interop tests (Go relay ↔ C peer over ngtcp2) — the wire layer compiles and unit tests pass; live integration tests need a relay running and weren't run in this session.
  - Vendored libsodium for Android NDK builds (T30).
  - The `c/vendor/build.sh` failure on the qtlsclient example target — the libs we need built fine, but the script ought to either skip examples or fix the link flags.
- **Known issues**:
  - `ci.yml` `Deploy to Fly.io` job continues to fail on master with `FLY_API_TOKEN` expired. Carried over from v0.18.0; orthogonal to release artifacts.

## 2026-04-30 — clean bill of health on testing (🎯T26 + T27 + T28 + T18)

- **Commit**: `0e14b44`
- **Outcome**: Brought every active SDK test suite from yellow/red to green and made `make bullseye` the durable one-button validator.
  - **🎯T26** — Rewrote the orphaned FSM test fixtures in Swift (`Tests/PigeonTests/PairingCeremonyMachineTests.swift`), Kotlin (`android/.../PairingCeremonyMachineTest.kt`), and TypeScript (`web/src/PairingCeremonyMachine.test.ts`) against the new acceptor / initiator actors generated from `protocol/pairing.yaml`. Each suite is now ~10 tests covering: initial state, full happy path with action-firing-order assertions, user_cancel from both AwaitingUserConfirm and AwaitingPeerConfirm, action-throws propagation, and wire-string round-trips for MessageType / ActionID. Fixed a Swift codegen bug in protocol/swift.go (was emitting `public typealias GuardID = …` even when the YAML had no guards — added `len(p.Guards) > 0` checks).
  - **🎯T27** — Audited every cross-language test that talks to a relay. Quarantine count: zero. The relay's bridgeClientPair preserves the legacy register / connect: wire, so Swift / Kotlin / TS continue to work at single-channel scope without porting. The only two real failures (Tests/PigeonRelayE2ETests/RelayE2ETests.swift::testCrossLanguageConfirmationCode and android/.../PigeonConnE2ETest.kt::cross-language confirmation code via relay) were unblocked by reviving cmd/crypto-peer/main.go: dropped its //go:build ignore tag and switched from the deprecated pigeon.Register signature to the still-exported pigeon.DialRelayAcceptor. Cross-language Go ↔ Swift and Go ↔ Kotlin confirmation-code agreement re-verified.
  - **🎯T28** — Extended `make bullseye` to run gofmt → go vet → go build → go test (all packages, -short -count=1) → swift build → swift test → kotlin :pigeon:test → web (tsx --test crypto.test.ts + PairingCeremonyMachine.test.ts) → TLC PairingCeremony. Every line green on master HEAD. Added `--root-out=DIR` to cmd/protogen and made writeFile auto-run gofmt on freshly-emitted .go files so a regen leaves the tree clean — the previous behaviour was for fresh codegen to fail bullseye gofmt the next run.
  - **T18** (already retired earlier; this is a related new flake) — A new TestE2ESingleClientChat flake (`recv "bye": EOF`) appeared after the T22 rewrite: the backend goroutine's deferred Session.Close raced against the client's third Recv, dropping the last echo's QUIC bytes mid-flush. Reshaped the test so the backend reads chat until EOF (driven by the client closing first) and the test explicitly closes the client session after receiving all echoes. Stress: 30/30 single-client and 20/20 all-three-E2E in parallel runs are green; previously ~1-in-N runs would fail.
- **Deferred**:
  - 🎯T29 — port Swift PigeonRelay to the post-T22 multi-channel wire protocol (register-mux + named-stream first-message + 4-byte clientTag prefix + AEAD-with-channel-id datagrams). Apple's Network.framework QUIC stack opens one stream per NWConnection by default; multi-stream multiplexing needs investigation into NWMultiplexGroup or per-stream NWConnections. Substantive port; not in scope for the bill-of-health pass.
  - 🎯T30 — same as T29 for Kotlin.
  - 🎯T31 — same as T29 for TypeScript.
  - 🎯T25 — Swift cross-language pairing E2E (Go acceptor ↔ Swift initiator with matching codes via the regenerated Swift PairingCeremonyMachine). Layered on top of T29.
- **Known issues**:
  - `ci.yml` `Deploy to Fly.io` job continues to fail on master with `FLY_API_TOKEN` expired. Carried over from v0.18.0; orthogonal to release artifacts.

## 2026-05-02 — /release v0.22.0

- **Commit**: `pending`
- **Outcome**: Released v0.22.0 (darwin-arm64, linux-amd64, linux-arm64). Two pieces of work, both landed in PR #28 after a parallel-agent fan-out from `/cv`. (1) **Vendor build fix** (🎯T32 unblocker): `c/vendor/build.sh` now uses `ENABLE_LIB_ONLY=ON` for ngtcp2's cmake. The previous `ENABLE_EXAMPLES=OFF` is not a valid ngtcp2 option; cmake silently ignored it, tried to build the examples, and `make install` bailed before staging the static libraries because the example link step needed brew-installed `libev`/`libnghttp3`. (2) **C-side full pairing wire driver** (🎯T34b.2): new `pigeon_pair_acceptor`/`pigeon_pair_initiator` in `c/src/pairing.c` run the hello/welcome/confirm exchange end-to-end against a `pigeon_transport` and produce a `pigeon_pairing_record`. Exposed from Go as `cwire.RunAcceptor`/`cwire.RunInitiator` via cgo wrappers that bridge the confirm callback through a `cgo.Handle` trampoline. Wire-byte parity locked by two new tests `TestWireParityGoAcceptorCInitiator`/`TestWireParityCAcceptorGoInitiator` driving both Go↔C and C↔Go halves of the same handshake over an in-process blocking `chanTransport`. Bonus root-cause fix: `b64_decode` in `c/src/pairing.c` used `0` as the past-end sentinel, but `0` is also a valid base64 char ('A'). For 32-byte X25519 keys (43 chars after stripping `=` padding) the final group's 4th slot was always present, causing an attempted 33rd-byte write into a 32-byte buffer. Switched to `-1`.
- **Process notes**:
  - First time the global "one PR per session" feedback was applied: the `/cv` fan-out spawned two worktree-isolated agents that each opened their own PR (#26 vendor fix + #27 pairing driver). User flagged this as costly to merge; consolidated both commits into a single branch and opened PR #28 in their place. Saved as `feedback_one_pr_at_a_time.md`. Still need to apply a matching directive to `~/.claude/CLAUDE.md` — pending user approval.
  - 🎯T34b.2 was discovered to be the tractable sub-target after `/cv` initially recommended fanning out on the bare 🎯T34 parent, which is documented to stall single agents.
- **Deferred**:
  - 🎯T32 steps 1–4 (multi-stream `ngtcp2_transport`, register-mux variant, Listener/Session ABI, live in-process Go-relay round-trip) — the vendor unblocker is the first block of the substantive multi-day work; nothing else of T32 landed in this release.
  - 🎯T34c.1 — cgo wrap of Listener / Register / Connect (the Pairing-callback piece). Context says best done after T34a + T34b have settled; T34b.2 settling now means this is unblocked for the next session.
  - 🎯T34d — delete the native Go peer-library files once T34a/b/c provide full coverage. Still blocked on T34c.1.
  - 🎯T25 — Swift cross-language pairing E2E (carried over from v0.21.0).
  - 🎯T35 — Kotlin JNI Android NDK matrix (carried over).
  - 🎯T36 — Java-callback transport (carried over).
  - 🎯T37 — Swift `Ngtcp2Transport` (carried over).
- **Known issues**:
  - `c/vendor/build.sh` does not auto-init nested git submodules; the ngtcp2 source tree's `third-party/urlparse` submodule must already be initialised (e.g. via `git submodule update --init` in the ngtcp2 directory, or by cloning the parent repo with `--recurse-submodules`) before the script will succeed. Documented in PR #28 description; not blocking CI which uses fresh clones with full submodule init.

## 2026-05-01 — /release v0.21.0

- **Commit**: `a9cf09d`
- **Outcome**: Released v0.21.0. The release re-bases all four language wrappers on a single C peer library: Swift (🎯T29) and Kotlin/JVM (🎯T30) are now thin shims over `libpigeon` (CPigeon SwiftPM C-target / JNI), Go (🎯T34, T34a, T34b, T34c.1, T38) is migrating to a cgo wrapper with the wire helpers, Session/Stream/Datagram, pairing FSM, confirmation code and `pigeon_transport` vtable bridged. C library extended with multi-channel session API, wire helpers, multi-stream callbacks, in-process loopback transport, PairingRecord binary serialisation and a restructured acceptor/initiator pairing FSM (🎯T32 steps 1-3, 🎯T33). TypeScript browser SDK gains the multi-channel API and wire-format unit tests locked to the Go byte vectors (🎯T31). New `crypto.Identity` interface and `NewFileIdentity`. New `Listener`/`Session`/`Stream`/`Datagram` Go types; `Register`/`Connect` now take struct-args (breaking). Visual `examples/demo` binary. `make bullseye` runs all SDK suites + TLC concurrently; `make bullseye-strict` adds ASan/UBSan + Go race. `protocol/pairing.yaml` is again the single source of truth (🎯T24). Released darwin-arm64, linux-amd64, linux-arm64. Homebrew formula updated.
- **Bullseye fixes shipped in this release**:
  - `8387f0c` — serialised `make generate` + `make amalgamate` before the parallel branches to fix the gofmt-vs-codegen race introduced when `make bullseye` was parallelised.
  - `61a471f` — split `test-c` into a deps-bearing wrapper and a body-only `test-c-only` target, and made `c/include/pigeon/loopback.h` self-contained (`<stddef.h>` + `<stdint.h>`). The bullseye fan-out's `test-c` branch was re-running codegen mid-fanout and rewriting Swift sources while `swift test` was compiling — "input file PairingCeremonyMachine.swift was modified during the build" — and `dist/loopback.h` only got `size_t` from a transitively-included `pigeon.h`, which broke the `import CPigeon` umbrella when the header was parsed first.
- **Deferred**:
  - 🎯T25 — Swift cross-language pairing E2E (Go acceptor ↔ Swift initiator) via the regenerated PairingCeremonyMachine. Now unblocked by 🎯T29 (CPigeon-backed Swift) but not yet implemented.
  - 🎯T35 — Kotlin JNI Android NDK build matrix (arm64-v8a + x86_64) with vendored libsodium AAR. Desktop JVM works (🎯T30); Android still pending.
  - 🎯T36 — Java-callback transport: C side calls back into a JVM/Kotlin `QuicTransport` via JNI.
  - 🎯T37 — Swift `Ngtcp2Transport` built on `c/src/ngtcp2_transport.c`.
  - Vendored libsodium for Android NDK (carried over from v0.20.0).
  - `c/vendor/build.sh` qtlsclient example link failure (carried over from v0.20.0; the static libs we need build fine).
- **Known issues**:
  - None new. The `Deploy to Fly.io` CI job is now green on master HEAD.

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

## 2026-03-25 — /release v0.7.0

- **Commit**: `0e2fab0`
- **Outcome**: Released v0.7.0. Persistent device pairing: `WithInstanceID` for stable relay identity, `PairingRecord` on all 4 platforms for save/restore of pairing state across reboots and network changes.

## 2026-03-25 — /release v0.6.0

- **Commit**: TBD
- **Outcome**: Released v0.6.0. 24 audit findings fixed (2 high, 7 medium, 15 low): Swift readExactly accumulation, goroutine leak in datagram relay, write deadline race, graceful shutdown, self-signed cert random serial, protogen output paths, maxMessageSize alignment, tern-bridge secure by default, datagram mode tests on Swift/Kotlin, generateNonce/generateSecret on all 4 platforms.

## 2026-03-24 — /release v0.5.0

- **Commit**: `70b55e6`
- **Outcome**: Released v0.5.0. Renamed Swift and Kotlin packages from TernCrypto/TernRelay to just Tern (single package per platform). Added convergence targets T12-T17.

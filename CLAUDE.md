# CLAUDE.md

Pigeon is an opaque authenticated WebTransport/QUIC relay library and server
(Go + Swift + Kotlin + C + TypeScript) for reaching backends behind NAT.

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

Pigeon is a WebTransport relay library and server (Go + Swift) that enables
connections between devices where the backend is on a private network
with no ingress. It establishes a virtual WebTransport session over QUIC
such that the relay itself cannot inspect the traffic.

Applications import pigeon's packages rather than implementing
relay/pairing/crypto logic themselves.

## Build, Test & Run

```bash
git submodule update --init --recursive  # vendored libsodium / quictls / ngtcp2
make build-vendor-deps            # static libs cwire's cgo build links against
make bullseye                     # one-button validation: every SDK + TLC
go build -o pigeon ./cmd/pigeon   # build relay server
go test ./...                     # all Go tests (relay, crypto, protocol, E2E)
go test -run TestE2E              # E2E integration test only
swift test                        # Swift crypto + state machine tests
JAVA_HOME=<jdk21> android/gradlew -p android test  # Kotlin tests
go run ./cmd/protogen protocol/pairing.yaml  # regenerate from YAML spec
make demo                         # in-process relay + N clients + web UI
```

**Prerequisites are not optional.** The root package and `cwire/` are cgo
packages that `#include` libsodium headers and link
`c/vendor/build/lib/libsodium.a` (`cwire/cwire.go:19-21`). Without the
submodule checkout and `make build-vendor-deps`, `go test ./...` fails with
`fatal error: 'sodium.h' file not found` for the root package, `cwire`,
`pairing`, `cmd/*` and `examples/*` — only `backchannel`, `crypto`,
`faultproxy`, `protocol` and `qr` build.

TLC also needs `formal/tla2tools.jar`, which is gitignored and not fetched by
any make target — download a TLA+ tools release into `formal/` before running
`./formal/tlc <Module>`.

`make bullseye` is the durable green signal: gofmt, go vet, go build, go test
(all packages, `-short`), swift build, swift test, kotlin :pigeon:test, web (tsx
--test), C tests, and TLC on both `PairingCeremony` and `SessionMachine`
(Makefile:248-280). If anything goes red across SDKs, this is the first thing
to run.

## Package Structure

### Relay Server (`cmd/pigeon/main.go`)

WebTransport relay server over QUIC/HTTP3. Backends register and get an
instance ID; clients connect by ID and traffic is bridged opaquely
(both streams and datagrams). Generates a self-signed TLS certificate
at startup for development; use --cert/--key for production certificates.

**Endpoints (HTTP/3 + HTTPS):** `GET /health`, `GET /status`, `GET /pigeon`

### Root Package (`api.go`, `pigeon.go`, `session.go`, `webtransport.go`)

Client-side connectivity and server library. Backends call
`pigeon.Register(ctx, *RegisterArgs)` (`api.go:171`) to get a `*Listener`
plus an instance ID; each `Listener.Accept` yields one `*Session`. Clients
call `pigeon.Connect(ctx, *ConnectArgs)` (`api.go:427`) and get a `*Session`
directly. A `Session` exposes `Primary()`, `OpenStream`/`AcceptStream`
(named reliable streams) and `Datagram(name)` (unreliable). Bearer token and
custom TLS are struct fields (`RegisterArgs.Token`, `RegisterArgs.TLS`), not
functional options.

The `WebTransportServer` type provides the relay server library used
by `cmd/pigeon`.

### E2E Crypto (`crypto/`)

Application-level encryption so the relay sees only ciphertext:
- **Key exchange:** X25519 ECDH
- **Symmetric encryption:** AES-256-GCM with monotonic counter nonce
- **Key derivation:** HKDF-SHA256
- **Confirmation codes:** 6-digit order-independent code from both ECDH pubkeys (MitM detection)

### Protocol Framework (`protocol/`)

Declarative state machine framework. Protocols are defined in YAML
(`protocol/pairing.yaml`) and serve as the single source of truth for:
- Go runtime executor (`machine.go`)
- Go code generator (`gogen.go`)
- Swift code generator (`swift.go`)
- TLA+ spec generator (`tla.go`)
- PlantUML diagram generator (`plantuml.go`)

### Code Generator (`cmd/protogen/`)

Reads YAML, validates, and generates Go/Swift/TLA+/PlantUML outputs.

### QR Helper (`qr/`)

Terminal QR code rendering and LAN IP detection for device pairing flows.

### Formal Model (`formal/`)

TLA+ specification (`PairingCeremony.tla`) with adversary model verifying:
no token reuse, MitM detection via code mismatch, device secret secrecy,
auth requires completed pairing, no nonce reuse.

Run with `./formal/tlc PairingCeremony`.

### Swift Package (`Package.swift`, `Sources/Pigeon/`)

Two SPM library products (`Package.swift:65-68`): `PigeonCore`
(`Sources/PigeonCore/` — E2ECrypto.swift, PigeonRelay.swift, PairingArtifact.swift,
CredentialStore.swift, and the generated PairingCeremonyMachine.swift) and
`Pigeon` (`Sources/Pigeon/` — Session/Transport over the `CPigeon` C target).
iOS apps add the GitHub repo as a package dependency.

### Android/Kotlin Library (`android/pigeon/`)

Android Library (AAR, AGP `com.android.library`) containing `E2EKeyPair`,
`E2EChannel`, `Hkdf`, `PigeonConn`, and the generated
`PairingCeremonyMachine.kt`. The packaged AAR includes
`libpigeon-jni.so` for `arm64-v8a` and `x86_64`, each statically linked
against a vendored libsodium built per-ABI by `c/vendor/build-android.sh`.
Consumed via JitPack (`com.github.marcelocantos.pigeon:pigeon:<tag>`).
Requires JDK 17+ / Android API 33+ (for X25519 support) and an
Android NDK r27+ install on the build host (pinned to
`29.0.14206865` via `android.ndkVersion` in `android/pigeon/build.gradle.kts`).
Desktop JVM unit tests still run on the host platform via
`gradlew :pigeon:testDebugUnitTest`; the AAR is built via
`gradlew :pigeon:assembleRelease`.

## Deployment

Fly.io config in `fly.toml`. Multi-stage Dockerfile (`golang:1.25-alpine`
-> `alpine:3.21`). WebTransport over QUIC on UDP port 443. Shared-cpu-1x,
256MB.

## Delivery

Merged to master.

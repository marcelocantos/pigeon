# CLAUDE.md

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
make bullseye                     # one-button validation: every SDK + TLC
go build -o pigeon ./cmd/pigeon   # build relay server
go test ./...                     # all Go tests (relay, crypto, protocol, E2E)
go test -run TestE2E              # E2E integration test only
swift test                        # Swift crypto + state machine tests
JAVA_HOME=<jdk21> android/gradlew -p android test  # Kotlin tests
go run ./cmd/protogen protocol/pairing.yaml  # regenerate from YAML spec
make demo                         # in-process relay + N clients + web UI
```

`make bullseye` is the durable green signal: gofmt, go vet, go build, go test
(all packages), swift build, swift test, kotlin :pigeon:test, web (tsx
--test), TLC PairingCeremony. If anything goes red across SDKs, this is
the first thing to run.

## Package Structure

### Relay Server (`cmd/pigeon/main.go`)

WebTransport relay server over QUIC/HTTP3. Backends register and get an
instance ID; clients connect by ID and traffic is bridged opaquely
(both streams and datagrams). Generates a self-signed TLS certificate
at startup for development; use --cert/--key for production certificates.

**Endpoints (HTTP/3):** `GET /health`, `GET /register`, `GET /ws/{id}`

### Root Package (`pigeon.go`, `conn.go`, `webtransport.go`)

Client-side connectivity and server library. Backends call
`pigeon.Register()` to get an instance ID; clients call
`pigeon.Connect()` with the ID. Both return a `*Conn` wrapping a
WebTransport session with `Send`/`Recv` (reliable stream) and
`SendDatagram`/`RecvDatagram` (unreliable). Supports bearer token
auth via `pigeon.WithToken()` and custom TLS via `pigeon.WithTLS()`.

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

SPM library (`Pigeon`) containing E2ECrypto.swift, PigeonRelay.swift, and the
generated PairingCeremonyMachine.swift. iOS apps add the GitHub repo as a
package dependency.

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

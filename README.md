# Tern

Tern is a WebSocket relay library and server (Go + Swift) that enables
connections between devices where the backend sits on a private network
with no ingress. The relay forwards opaque ciphertext — it never sees
plaintext traffic. Applications import tern's packages rather than
implementing relay, pairing, or crypto logic themselves.

## Trust Model

All application traffic is end-to-end encrypted:

- **Key exchange:** X25519 ECDH — each side generates an ephemeral key pair
  and derives a shared secret.
- **Symmetric encryption:** AES-256-GCM with monotonic counter nonces and
  directional key derivation via HKDF-SHA256.
- **MitM detection:** A 6-digit confirmation code derived from both public
  keys is displayed on each device. Users verify the codes match during
  the pairing ceremony.

The relay server handles only ciphertext and has no access to session keys.

## How It Works

1. A **backend** connects to `GET /register` via WebSocket. The relay
   assigns a unique instance ID and sends it back as the first message.
2. A **client** connects to `GET /ws/<instance-id>`. The relay bridges
   all traffic bidirectionally between the two WebSocket connections.
3. Pairing and encryption happen above the relay layer, in the
   application, using tern's crypto and protocol packages.

## Go Library

```go
import (
    "github.com/marcelocantos/tern/crypto"
    "github.com/marcelocantos/tern/protocol"
    "github.com/marcelocantos/tern/qr"
)
```

| Package    | Purpose                                                     |
|------------|-------------------------------------------------------------|
| `crypto/`  | X25519 key exchange, AES-256-GCM channel, confirmation code |
| `protocol/`| Declarative state machine framework and pairing ceremony     |
| `qr/`      | Terminal QR code rendering and LAN IP detection              |

## Swift Package

Add the GitHub repo as an SPM dependency:

```
https://github.com/marcelocantos/tern
```

The package provides the `TernCrypto` library (iOS 16+, macOS 13+)
containing `E2ECrypto.swift` (key exchange and encrypted channel) and
the generated `PairingCeremonyMachine.swift`.

## Running the Relay Server

```bash
go build -o tern .
PORT=8080 ./tern
```

The server is also deployable via Fly.io (`fly.toml` and `Dockerfile`
are included).

**Endpoints:**

| Route              | Description                          |
|--------------------|--------------------------------------|
| `GET /health`      | Health check (returns `{"status":"ok"}`) |
| `GET /register`    | Backend registers (WebSocket upgrade)|
| `GET /ws/{id}`     | Client connects by instance ID       |

## Running Tests

```bash
# Go — relay, crypto, protocol, and E2E integration tests
go test ./...

# Swift — crypto and state machine tests
swift test
```

## Protocol Code Generation

Protocols are defined in YAML (`protocol/pairing.yaml`) and used to
generate Go, Swift, TLA+, and PlantUML outputs:

```bash
go run ./cmd/protogen protocol/pairing.yaml
```

## Formal Model

A TLA+ specification (`formal/PairingCeremony.tla`) models the pairing
ceremony with an active adversary. Verified security properties include:

- No token reuse
- MitM detection via confirmation code mismatch
- Device secret secrecy
- Authentication requires completed pairing
- No nonce reuse

Run the model checker:

```bash
./formal/tlc PairingCeremony
```

## Licence

Apache 2.0 — see [LICENSE](LICENSE).

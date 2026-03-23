# Tern

Tern is a WebTransport relay library and server (Go + Swift) that enables
connections between devices where the backend sits on a private network
with no ingress. The relay forwards opaque ciphertext over QUIC — it never
sees plaintext traffic. Applications import tern's packages rather than
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

1. A **backend** connects to `GET /register` via WebTransport. The relay
   assigns a unique instance ID and sends it back as the first message.
2. A **client** connects to `GET /ws/<instance-id>`. The relay bridges
   all traffic bidirectionally between the two WebTransport sessions —
   both reliable streams and unreliable datagrams.
3. Pairing and encryption happen above the relay layer, in the
   application, using tern's crypto and protocol packages.

## Go Library

```bash
go get github.com/marcelocantos/tern
```

```go
import (
    "github.com/marcelocantos/tern"
    "github.com/marcelocantos/tern/crypto"
    "github.com/marcelocantos/tern/protocol"
    "github.com/marcelocantos/tern/qr"
)
```

| Package    | Purpose                                                     |
|------------|-------------------------------------------------------------|
| root       | Client-side relay connectivity (register, connect, send/recv)|
| `crypto/`  | X25519 key exchange, AES-256-GCM channel, confirmation code |
| `protocol/`| Declarative state machine framework and pairing ceremony     |
| `qr/`      | Terminal QR code rendering and LAN IP detection              |

**Quick integration — relay + encrypted channel:**

```go
// Backend registers with the relay.
backend, _ := tern.Register(ctx, "https://tern.fly.dev",
    tern.WithToken(os.Getenv("TERN_TOKEN")),
    tern.WithTLS(tlsConfig))
fmt.Println("Instance ID:", backend.InstanceID()) // share via QR code

// Client connects by instance ID (obtained from QR scan).
client, _ := tern.Connect(ctx, "https://tern.fly.dev", instanceID,
    tern.WithTLS(tlsConfig))

// Send/receive through the relay (reliable stream).
client.Send(ctx, ciphertext)
data, _ := backend.Recv(ctx)

// Unreliable datagrams (for latency-sensitive data like video frames).
client.SendDatagram(data)
data, _ = backend.RecvDatagram(ctx)
```

**Encrypted channel:**

```go
// Both sides generate an ephemeral key pair and exchange public keys.
kp, _ := crypto.GenerateKeyPair()
// ... send kp.Public.Bytes() to peer; receive peerPubBytes ...
peerPub, _ := ecdh.X25519().NewPublicKey(peerPubBytes)

// Derive directional session keys and open an encrypted channel.
sendKey, _ := crypto.DeriveSessionKey(kp.Private, peerPub, []byte("client-to-server"))
recvKey, _ := crypto.DeriveSessionKey(kp.Private, peerPub, []byte("server-to-client"))
ch, _ := crypto.NewChannel(sendKey, recvKey)

// Verify the pairing is MitM-free (show 6-digit codes on both devices).
code, _ := crypto.DeriveConfirmationCode(kp.Public, peerPub)
fmt.Println("Confirmation code:", code) // e.g. "042857"

// Encrypt / decrypt messages sent through the relay.
encrypted := ch.Encrypt([]byte("hello"))
plaintext, _ := ch.Decrypt(encrypted)
```

## Swift Package

Add the GitHub repo as an SPM dependency:

```
https://github.com/marcelocantos/tern
```

The package provides the `TernCrypto` library (iOS 16+, macOS 13+)
containing `E2ECrypto.swift` (key exchange and encrypted channel) and
the generated `PairingCeremonyMachine.swift`.

```swift
// Both sides exchange public key bytes through the relay.
let kp = E2EKeyPair()
// ... send kp.publicKeyData; receive peerPubBytes ...
let sessionKey = try kp.deriveSessionKey(peerPublicKey: peerPubBytes,
                                         info: Data("client-to-server".utf8))
let channel = E2EChannel(sharedKey: sessionKey, isServer: false)
let encrypted = try channel.encrypt(plaintext)
let plaintext  = try channel.decrypt(ciphertext)
```

## Android/Kotlin Library

Add via [JitPack](https://jitpack.io) (Gradle):

```kotlin
// settings.gradle.kts
dependencyResolutionManagement {
    repositories {
        maven("https://jitpack.io")
    }
}

// build.gradle.kts
dependencies {
    implementation("com.github.marcelocantos.tern:terncrypto:v0.3.0")
}
```

Requires JDK 17+ / Android API 33+ (for X25519).

```kotlin
// Key exchange
val kp = E2EKeyPair()
// ... send kp.publicKeyData (32 bytes); receive peerPubBytes ...
val sessionKey = kp.deriveSessionKey(peerPubBytes, "client-to-server".toByteArray())

// Encrypted channel from shared key
val channel = E2EChannel(sharedKey, isServer = false)
val encrypted = channel.encrypt(plaintext)
val plaintext = channel.decrypt(ciphertext)
```

## Pairing Ceremony

The full ceremony involves three actors — **server** (backend daemon),
**mobile** (iOS client), and **CLI** (initiator):

1. CLI sends `pair_begin` to server; server generates a one-time token,
   connects to the relay (`/register`), and receives an instance ID.
2. Server displays a QR code encoding the relay URL, token, and instance ID.
3. Mobile scans the QR, connects to `/ws/{id}`, generates an X25519 key pair,
   and sends `{token, pubkey}` to the server through the relay.
4. Server verifies the token, performs ECDH, derives the session key, and sends
   `pair_hello_ack {pubkey}` back. Mobile performs ECDH and derives the same key.
5. Both sides independently compute the 6-digit confirmation code from the two
   public keys. The server signals CLI to show the code; mobile shows it on screen.
   The user verifies the codes match — a mismatch means a MitM is present.
6. CLI submits the code the user entered. If correct, the server sends
   `pair_complete {secret, key}` to mobile and `pair_status` to CLI. Pairing done.

![Pairing ceremony state machines](docs/PairingCeremony.svg)

## Running the Relay Server

```bash
go build -o tern ./cmd/tern
PORT=443 ./tern                           # self-signed cert (development)
./tern --cert cert.pem --key key.pem      # production TLS certificate
```

The server is also deployable via Fly.io (`fly.toml` and `Dockerfile`
are included).

**Endpoints (HTTP/3 over WebTransport):**

| Route              | Description                               |
|--------------------|-------------------------------------------|
| `GET /health`      | Health check (returns `{"status":"ok"}`)  |
| `GET /register`    | Backend registers (WebTransport session)  |
| `GET /ws/{id}`     | Client connects by instance ID            |

## Configuration

| Flag / Env var | Default | Description |
|----------------|---------|-------------|
| `--port` / `PORT` | `443` | Listening port (UDP + TCP) |
| `--domain` | — | Domain for automatic Let's Encrypt TLS (e.g. `tern.fly.dev`) |
| `--acme-email` | — | Email for Let's Encrypt account |
| `--cert` | — | TLS certificate file (PEM); if `--domain` is not set |
| `--key` | — | TLS private key file (PEM); used with `--cert` |
| `TERN_TOKEN` | — | Bearer token required for `/register`; open if unset |
| `--version` | — | Print version and exit |
| `--help-agent` | — | Print usage + agent guide |

Build-time version injection: `go build -ldflags "-X main.version=v1.0.0" ./cmd/tern`

Max message frame size: 1 MiB (constant `maxMessageSize`).

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

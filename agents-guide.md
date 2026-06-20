# Pigeon — Agent Guide

## What Is Pigeon?

Pigeon is a Go + Swift + Kotlin + C + TypeScript library for opaque authenticated WebTransport relay. It provides:

- A relay server that bridges WebTransport sessions without seeing plaintext
- E2E encryption (X25519 ECDH + AES-256-GCM) with a pairing ceremony and MitM detection
- A declarative protocol state machine framework with code generation (Go, Swift, Kotlin, C, TypeScript, TLA+, PlantUML)
- A Swift package (`Pigeon`) for iOS 16+/macOS 13+
- A pure C client library (`dist/pigeon.h` + `dist/pigeon.c`) with zero heap allocations

## Go Packages

| Package | Import | Purpose |
|---------|--------|---------|
| `crypto` | `github.com/marcelocantos/pigeon/crypto` | Key exchange, encrypted channel, confirmation code |
| `protocol` | `github.com/marcelocantos/pigeon/protocol` | State machine framework + pairing ceremony |
| `qr` | `github.com/marcelocantos/pigeon/qr` | Terminal QR rendering, LAN IP detection |

### crypto/

```go
// Key exchange
kp, err := crypto.GenerateKeyPair()   // *KeyPair{Private, Public}
sessionKey, err := crypto.DeriveSessionKey(kp.Private, peerPub, info)

// MitM detection — both sides compute; if codes differ, abort
code, err := crypto.DeriveConfirmationCode(myPub, peerPub) // "123456"

// Encrypted channel (separate send/recv keys for directional nonces)
ch, err := crypto.NewChannel(sendKey, recvKey)
ch, err := crypto.NewSymmetricChannel(sessionKey, isServer)
encrypted := ch.Encrypt(plaintext)   // []byte (seq prefix + ciphertext)
plaintext, err := ch.Decrypt(data)   // []byte

// Utilities
nonce, err := crypto.GenerateNonce()   // 32 random bytes
secret, err := crypto.GenerateSecret() // 32 random bytes
```

### protocol/

```go
p, err := protocol.LoadYAML("protocol/pairing.yaml")
m := protocol.NewMachine(actor, p)
m.RegisterGuard("guard_name", func(ctx context.Context) bool { ... })
m.RegisterAction("action_name", func(ctx context.Context) error { ... })
newState, err := m.Handle(ctx, message) // process recv trigger
newState, err := m.Step(ctx)            // fire internal trigger
```

### qr/

```go
qr.Print(os.Stdout, url) // render QR to terminal (Unicode half-blocks)
ip := qr.LanIP()         // "192.168.1.5" or "localhost" on error
```

### Root package (relay client)

The root package exposes `Register`, `Connect`, `Listener`, `Session`, `Stream`, and `Datagram`.

**`Register`** publishes a backend on the relay and returns a `*Listener` plus the stable `instanceID` to advertise. Each `listener.Accept(ctx)` call returns one `*Session` for a newly paired client.

**`Connect`** dials an already-paired backend by `instanceID` and returns a `*Session` ready for I/O.

**`Session`** is the per-peer handle. It carries:
- `Primary() *Stream` — the default reliable stream (use in pairing-mode or as the main message pipe).
- `OpenStream(ctx, name) (*Stream, error)` — open a named reliable stream; the peer calls `AcceptStream` with the same name.
- `AcceptStream(ctx, name) (*Stream, error)` — accept a named stream opened by the peer.
- `Datagram(name) *Datagram` — retrieve a pre-declared datagram channel by name.
- `PeerID() string` — InstanceID of the remote peer.
- `Close() error` — tear down the session.

**`Stream`** is a reliable, ordered, AEAD-encrypted message channel:
- `Send(msg []byte) error`
- `Recv(ctx context.Context) ([]byte, error)`
- `Close() error`

**`Datagram`** is an unreliable, unordered channel keyed by pre-agreed varint ID:
- `Send(payload []byte) error`
- `Recv(ctx context.Context) ([]byte, error)`

```go
// Backend side.
listener, instanceID, err := pigeon.Register(ctx, &pigeon.RegisterArgs{
    Identity:  identity,          // crypto.Identity (long-term keypair)
    Pairing:   resolvePairing,    // func(clientID string) (*crypto.PairingRecord, error)
    Relay:     "https://relay.example.com",
    Token:     bearerToken,       // optional; wires through BearerTokenAuth
    Datagrams: map[string]uint64{"video": 1},
})
defer listener.Close()
// Advertise instanceID (e.g. as QR payload), then loop:
session, err := listener.Accept(ctx)
defer session.Close()

// Client side.
session, err := pigeon.Connect(ctx, &pigeon.ConnectArgs{
    InstanceID: instanceID,
    Record:     pairingRecord,    // *crypto.PairingRecord from the pairing ceremony
    Identity:   identity,
    Relay:      "https://relay.example.com",
    Datagrams:  map[string]uint64{"video": 1},
})
defer session.Close()

// Reliable stream I/O.
primary := session.Primary()
if err := primary.Send([]byte("hello")); err != nil { ... }
msg, err := primary.Recv(ctx)

// Named extra stream (both peers call matching Open/Accept).
stream, err := session.OpenStream(ctx, "control")
// peer: stream, err := session.AcceptStream(ctx, "control")

// Datagram I/O.
video := session.Datagram("video")
video.Send(frame)
frame, err := video.Recv(ctx)
```

## Relay Authentication

The relay server accepts a `pigeon.Auth` hook that controls which backends and clients it admits. Both fields are optional; the zero `Auth` means "accept all" (open relay).

```go
type Auth struct {
    VerifyRegister func(ctx context.Context, req *RegisterRequest) error
    VerifyConnect  func(ctx context.Context, req *ConnectRequest) error
}
```

`Auth` is passed to `NewWebTransportServer` or `NewQUICServer` when constructing a relay. Client-side auth is separate: pass a `Token` in `ConnectArgs` / `RegisterArgs` and the relay checks it against the configured verifier.

**Built-in verifiers:**

`BearerTokenAuth(token string) Auth` — admits only registrations whose greeting token matches `token` (constant-time compare). An empty token returns the zero `Auth`. `VerifyConnect` is left unset — clients are admitted transitively by backend pairing logic.

`MutualTLSAuth(pool *x509.CertPool) Auth` — requires a TLS client certificate signed by one of the roots in `pool`. The caller must also set `tls.Config.ClientAuth = tls.RequireAndVerifyClientCert` on the server TLS config.

**`PIGEON_TOKEN`** wires through `BearerTokenAuth` by default. Adopt a custom `Auth` to replace it with any admission policy.

**`RegisterRequest` / `ConnectRequest`** expose the live transport handle for verifiers: `Token`, `InstanceID`, `TLS` (negotiated state + peer certs), `QUICConn` (raw QUIC; nil for WebTransport), and `HTTPRequest` (WebTransport upgrade request; nil for raw QUIC).

**Custom verifier example** (fingerprint allowlist):

```go
allowlist := map[string]bool{
    "AA:BB:CC:...": true,
}
auth := pigeon.Auth{
    VerifyRegister: func(ctx context.Context, req *pigeon.RegisterRequest) error {
        if req.TLS == nil || len(req.TLS.PeerCertificates) == 0 {
            return pigeon.ErrUnauthorized
        }
        fp := fingerprintSHA256(req.TLS.PeerCertificates[0])
        if !allowlist[fp] {
            return pigeon.ErrUnauthorized
        }
        return nil
    },
}
```

See [docs/DESIGN.md §2](docs/DESIGN.md) for the broader threat model that motivates these hooks.

## Relay Endpoints

| Route | Description |
|-------|-------------|
| `GET /health` | Returns `{"status":"ok"}` (HTTP/3) |
| `GET /pigeon` | Single WebTransport entry point; the role (register / listen / connect) is set by the greeting on the primary stream |

Native clients use raw QUIC (ALPN `"pigeon"`) instead of WebTransport; the
same greeting variant (register / listen / connect) selects the role.

Multiple clients can connect to the same instance ID. The backend parks a
pool of listen connections at the relay; the relay matches each client to a
parked listen and bridges the two QUIC connections end-to-end, so backends
serve many peers concurrently — each over its own opaque pipe.

## Swift (SPM)

```
https://github.com/marcelocantos/pigeon
```

Product: `Pigeon`. Platforms: iOS 16+, macOS 13+.

```swift
let kp = E2EKeyPair()
let sessionKey = try kp.deriveSessionKey(peerPublicKey: peerPubBytes, info: info)
let channel = E2EChannel(sharedKey: sessionKey, isServer: false)
let encrypted = try channel.encrypt(plaintext)
let plaintext = try channel.decrypt(ciphertext)
```

## C Client Library

Distributed as two files: `dist/pigeon.h` + `dist/pigeon.c`. Compile with
`-DPIGEON_CRYPTO_LIBSODIUM` and link `-lsodium`.

```c
// All state lives in a struct — allocate on stack, static, or embedded.
pigeon_ctx ctx;
pigeon_init(&ctx, &transport);  // transport = user-provided QUIC callbacks

// Crypto
pigeon_keypair kp;
pigeon_generate_keypair(&kp);
pigeon_derive_session_key(kp.private_key, peer_pub, info, info_len, out_key);

char code[7];
pigeon_derive_confirmation_code(pub_a, pub_b, code);  // "123456"

// Encrypted channel
pigeon_channel ch;
pigeon_channel_init(&ch, send_key, recv_key, PIGEON_MODE_STRICT);
pigeon_channel_encrypt(&ch, plaintext, pt_len, out, out_len);
pigeon_channel_decrypt(&ch, ciphertext, ct_len, out, out_len);

// Wire framing (4-byte big-endian length prefix)
pigeon_frame_message(payload, len, buf, buf_len);
pigeon_send(&ctx, data, len);     // length-prefixed stream message
pigeon_recv(&ctx, buf, buf_len);  // returns message length
```

Zero heap allocations. `PIGEON_MAX_MSG` (default 1 MiB) is the sole build-time knob.
Generated pairing state machine included — all three actors (server, ios, cli).

## Pairing Ceremony Flow

The three actors are **server** (backend daemon), **mobile** (iOS client), and **CLI** (initiator):

```
CLI               Server              Relay              Mobile
---               ------              -----              ------
cli --init
  └─ pair_begin ─→
                  generate token
                  register ────────────────────────────→ relay
                  ←──────────────────────────────── instance_id
                  show QR(url+token+id)
                                                     scan QR
                                                     connect ──→ relay (by id)
                                                     send {token, pubkey}
                  ←────────────────────────────────────────────
                  verify token
                  ECDH → session key
                  send pair_hello_ack ─────────────────────────→
                                                     ECDH → session key
                  ←── send waiting_for_code ──→
  show code (6d)  show code (6d)                     show code (6d)
user verifies codes match on both devices
  enter code ──→ code_submit ─→
                  code correct?
                  send pair_complete ──────────────────────────→
                                                     store device secret
  ← pair_status ←
```

MitM detection: the 6-digit confirmation code is `HKDF(min(a,b) || max(a,b), "pairing-confirmation")`. An adversary who substituted their own public key gets a different code — both devices show mismatched codes and the user aborts.

## Persistent Pairing

After the first pairing ceremony, save a `PairingRecord` for reconnection
without repeating the ceremony:

```go
// After first pairing — save securely (e.g., Keychain, EncryptedSharedPreferences)
record := crypto.NewPairingRecord(peerInstanceID, relayURL, myKeyPair, peerPubKey)
data, _ := record.Marshal()

// On reconnect — load the record and connect; Connect derives the
// encrypted channel from it.
record, _ := crypto.UnmarshalPairingRecord(data)
session, _ := pigeon.Connect(ctx, &pigeon.ConnectArgs{
    InstanceID: record.PeerInstanceID,
    Record:     record,
    Identity:   identity,
    Relay:      record.RelayURL,
})
defer session.Close()
```

The shared secret is re-derived on each reconnect (never stored). Available
on all platforms: Go (`crypto.PairingRecord`), Swift (`PairingRecord`),
Kotlin (`PairingRecord`), TypeScript (`PairingRecord` /
`createPairingRecord` / `deriveChannelFromRecord`).

## Pairing Artifact Lifecycle (v0.19.0+)

For the consumer-app integration path — pairing once, persisting the
credential across launches, reconnecting via a one-call helper, and
re-pairing on expiry — pigeon ships a higher-level set of primitives:

- **`PairingArtifact`** wraps a `PairingRecord` with a token, an
  issued-at timestamp, and an expires-at timestamp. Canonical JSON
  encoding (snake_case keys, ISO-8601 timestamps) and a single-line
  base64url text encoding interoperate across Go, Swift, and Kotlin.
- **`CredentialStore`** is a uniform persistence interface with
  reference implementations: `FileCredentialStore` (Go/Swift/Kotlin)
  and `KeychainCredentialStore` (Swift, iOS/macOS).
- **`PairingHost`** is the server-side artifact minter. Configurable
  TTL (default 30 days) and optional bearer-token issuer.
- **`pigeon pair`** is a CLI subcommand for deploy scripts. Mints an
  artifact and emits it on stdout (`--out=-`, default); writes the
  companion server-side `PairingRecord` to stderr.
- **`ConnectWithArtifact`** (Go), **`PigeonConn.connect(artifact:)`**
  (Swift), **`connectWithArtifact`** (Kotlin) check expiry up front
  and route through typed errors (`ErrPairingExpired`,
  `PairingError.expired`, `PairingExpiredException`).

Two delivery flows are first-class: QR scan from a paired-screen UX
(the artifact's text encoding fits in a scannable QR payload) and
developer-deploy via `xcrun` (mint on the laptop, inject into the
iOS app's launch environment as `PIGEON_PAIRING_ARTIFACT`).

The full step-by-step guide — server-side acceptance pattern,
side-by-side Go/Swift/Kotlin samples, security notes — lives in
**[docs/pairing-lifecycle.md](docs/pairing-lifecycle.md)**. Read it
before integrating pigeon into a new application.

## Common Commands

```bash
go test ./...                             # all Go tests (relay, crypto, protocol, qr, E2E)
swift test                                # Swift tests
go build -o pigeon ./cmd/pigeon               # build relay binary
go run ./cmd/protogen protocol/pairing.yaml   # regenerate state machine code
./formal/tlc PairingCeremony              # run TLA+ model checker
PORT=443 ./pigeon                           # run relay server (self-signed cert)
./pigeon --cert cert.pem --key key.pem      # run with real TLS certificate
```

## Configuration

| Flag/Env | Default | Description |
|----------|---------|-------------|
| `pair` (subcommand) | — | Mint a `PairingArtifact` for a peer instance ID and emit it. See [docs/pairing-lifecycle.md](docs/pairing-lifecycle.md). |
| `--port` / `PORT` | `443` | WebTransport listening port (UDP) |
| `--quic-port` / `QUIC_PORT` | `4433` | Raw QUIC listening port (native clients) |
| `--domain` | — | Domain for automatic Let's Encrypt TLS |
| `--acme-email` | — | Email for Let's Encrypt account |
| `--cert` | — | TLS certificate file (PEM); if omitted, generates self-signed |
| `--key` | — | TLS private key file (PEM) |
| `--cert-validity` | `365` | Self-signed certificate validity in days (use ≤14 for WebTransport `serverCertificateHashes`) |
| `--lan` | — | LAN listener address for direct connections (e.g. `:0`); not yet wired into the relay flow |
| `--version` | — | Print version and exit |
| `--help-agent` | — | Print this guide |
| `PIGEON_TOKEN` | — | Bearer token for backend registration auth. Wires through the default `BearerTokenAuth` verifier; replace with any custom `pigeon.Auth` for more complex admission policies. |

Build-time version injection: `-ldflags "-X main.version=<version>"`.

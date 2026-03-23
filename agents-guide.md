# Tern — Agent Guide

## What Is Tern?

Tern is a Go + Swift library for opaque authenticated WebTransport relay. It provides:

- A relay server that bridges WebTransport sessions without seeing plaintext
- E2E encryption (X25519 ECDH + AES-256-GCM) with a pairing ceremony and MitM detection
- A declarative protocol state machine framework with code generation (Go, Swift, TLA+, PlantUML)
- A Swift package (`TernCrypto`) for iOS 16+/macOS 13+

## Go Packages

| Package | Import | Purpose |
|---------|--------|---------|
| `crypto` | `github.com/marcelocantos/tern/crypto` | Key exchange, encrypted channel, confirmation code |
| `protocol` | `github.com/marcelocantos/tern/protocol` | State machine framework + pairing ceremony |
| `qr` | `github.com/marcelocantos/tern/qr` | Terminal QR rendering, LAN IP detection |

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

```go
// Register as a backend — returns a Conn with an assigned instance ID.
conn, err := tern.Register(ctx, "https://relay.example.com",
    tern.WithToken("bearer-token"),
    tern.WithTLS(&tls.Config{RootCAs: pool}))
defer conn.Close()
id := conn.InstanceID()

// Connect as a client to a specific backend instance.
conn, err := tern.Connect(ctx, "https://relay.example.com", instanceID,
    tern.WithTLS(&tls.Config{RootCAs: pool}))
defer conn.Close()

// Stream messages (reliable, ordered).
conn.Send(ctx, []byte("hello"))
data, err := conn.Recv(ctx)

// Datagrams (unreliable, unordered — for real-time data).
conn.SendDatagram([]byte("frame"))
data, err := conn.RecvDatagram(ctx)

// E2E encryption — set after key exchange.
conn.SetChannel(ch)            // encrypts/decrypts stream messages
conn.SetDatagramChannel(dgCh)  // encrypts/decrypts datagrams
```

## Relay Endpoints

| Route | Description |
|-------|-------------|
| `GET /health` | Returns `{"status":"ok"}` (HTTP/3) |
| `GET /register` | Backend registers (WebTransport session); receives assigned instance ID |
| `GET /ws/{id}` | Client connects by instance ID (WebTransport session); bridged bidirectionally to backend |

One client per instance. A second client connection returns HTTP 409.

## Swift (SPM)

```
https://github.com/marcelocantos/tern
```

Product: `TernCrypto`. Platforms: iOS 16+, macOS 13+.

```swift
let kp = E2EKeyPair()
let sessionKey = try kp.deriveSessionKey(peerPublicKey: peerPubBytes, info: info)
let channel = E2EChannel(sharedKey: sessionKey, isServer: false)
let encrypted = try channel.encrypt(plaintext)
let plaintext = try channel.decrypt(ciphertext)
```

## Pairing Ceremony Flow

The three actors are **server** (backend daemon), **mobile** (iOS client), and **CLI** (initiator):

```
CLI               Server              Relay              Mobile
---               ------              -----              ------
cli --init
  └─ pair_begin ─→
                  generate token
                  connect ─────────────────────────────→ /register
                  ←──────────────────────────────── instance_id
                  show QR(url+token+id)
                                                     scan QR
                                                     connect ──→ /ws/{id}
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

## Common Commands

```bash
go test ./...                             # all Go tests (relay, crypto, protocol, qr, E2E)
swift test                                # Swift tests
go build -o tern .                        # build relay binary
go run ./cmd/protogen protocol/pairing.yaml   # regenerate state machine code
./formal/tlc PairingCeremony              # run TLA+ model checker
PORT=443 ./tern                           # run relay server (self-signed cert)
./tern --cert cert.pem --key key.pem      # run with real TLS certificate
```

## Configuration

| Flag/Env | Default | Description |
|----------|---------|-------------|
| `--port` / `PORT` | `443` | Relay listening port (UDP) |
| `--cert` | — | TLS certificate file (PEM); if omitted, generates self-signed |
| `--key` | — | TLS private key file (PEM) |
| `--version` | — | Print version and exit |
| `--help-agent` | — | Print this guide |

Build-time version injection: `-ldflags "-X main.version=v0.1.0"`.

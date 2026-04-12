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

```go
// Register as a backend — returns a Conn with an assigned instance ID.
conn, err := pigeon.Register(ctx, "https://relay.example.com",
    pigeon.WithToken("bearer-token"),
    pigeon.WithTLS(&tls.Config{RootCAs: pool}))
defer conn.Close()
id := conn.InstanceID()

// Connect as a client to a specific backend instance.
conn, err := pigeon.Connect(ctx, "https://relay.example.com", instanceID,
    pigeon.WithTLS(&tls.Config{RootCAs: pool}))
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

## Persistent Pairing

After the first pairing ceremony, save a `PairingRecord` for reconnection
without repeating the ceremony:

```go
// After first pairing — save securely (e.g., Keychain, EncryptedSharedPreferences)
record := crypto.NewPairingRecord(backend.InstanceID(), relayURL, myKeyPair, peerPubKey)
data, _ := record.Marshal()

// On reconnect — load and derive channel
record, _ := crypto.UnmarshalPairingRecord(data)
ch, _ := record.DeriveChannel([]byte("client-to-server"), []byte("server-to-client"))
conn, _ := pigeon.Connect(ctx, record.RelayURL, record.PeerInstanceID)
conn.SetChannel(ch)
```

The shared secret is re-derived on each reconnect (never stored). Available
on all platforms: Go (`crypto.PairingRecord`), Swift (`PairingRecord`),
Kotlin (`PairingRecord`), TypeScript (`PairingRecord` /
`createPairingRecord` / `deriveChannelFromRecord`).

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
| `--port` / `PORT` | `443` | Relay listening port (UDP) |
| `--domain` | — | Domain for automatic Let's Encrypt TLS |
| `--acme-email` | — | Email for Let's Encrypt account |
| `--cert` | — | TLS certificate file (PEM); if omitted, generates self-signed |
| `--key` | — | TLS private key file (PEM) |
| `--version` | — | Print version and exit |
| `--help-agent` | — | Print this guide |
| `PIGEON_TOKEN` | — | Bearer token for /register auth |

Build-time version injection: `-ldflags "-X main.version=<version>"`.

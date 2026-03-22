# Stability

This document tracks tern's readiness for a 1.0 release — the point at which
backwards compatibility becomes a binding commitment. Once v1.0.0 ships,
breaking changes to any public surface listed below require a major version bump.

The pre-1.0 period (currently v0.x.x) exists to get the interaction surface right.

---

## Interaction Surface Catalogue

*Snapshot as of v0.1.0.*

### Relay HTTP API (the binary's external interface)

| Route | Protocol | Response |
|-------|----------|----------|
| `GET /health` | HTTP | `{"status":"ok"}` |
| `GET /register` | WebSocket | First text message is the assigned instance ID |
| `GET /ws/{id}` | WebSocket | Bridged bidirectionally to registered backend |

Constraints: one client per instance; second client returns HTTP 409.
Max frame size: 1 MiB.
CORS: `OriginPatterns: ["*"]` (intentional default).

*Stability: Stable.*

### CLI interface (the `tern` binary)

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--port` | string | `""` | Listening port (overrides `PORT` env var) |
| `--version` | bool | `false` | Print version and exit |
| `--help-agent` | bool | `false` | Print usage + agents-guide.md and exit |

Environment variables: `PORT` (default `8080`).

Build-time version injection: `-ldflags "-X main.version=<version>"`.

*Stability: Stable.*

### Wire format (encrypted message frame)

```
[8-byte sequence number, little-endian uint64]
[ciphertext (variable length)]
[16-byte AES-GCM authentication tag]
```

The sequence number doubles as both the replay-prevention counter and the
AES-GCM nonce (first 8 bytes of the 12-byte nonce, remaining 4 bytes zero).

*Stability: Stable.*

### `crypto/` Go package

```go
// Types
type KeyPair struct {
    Private *ecdh.PrivateKey
    Public  *ecdh.PublicKey
}
type Channel struct { /* unexported fields */ }

// Key exchange
func GenerateKeyPair() (*KeyPair, error)
func DeriveSessionKey(priv *ecdh.PrivateKey, peerPub *ecdh.PublicKey, info []byte) ([]byte, error)
func DeriveKeyFromSecret(secret, nonce []byte) ([]byte, error)

// Utilities
func GenerateNonce() ([]byte, error)    // 32 random bytes
func GenerateSecret() ([]byte, error)  // 32 random bytes
func DeriveConfirmationCode(pubA, pubB *ecdh.PublicKey) (string, error) // 6-digit code

// Channel construction
func NewChannel(sendKey, recvKey []byte) (*Channel, error)
func NewSymmetricChannel(key []byte, isServer bool) (*Channel, error)

// Channel methods
func (*Channel) Encrypt(plaintext []byte) []byte
func (*Channel) Decrypt(data []byte) ([]byte, error)
```

*Stability: Stable — `NewSymmetricChannel` may be renamed for clarity before 1.0.*

### `protocol/` Go package

```go
// Core types
type Protocol struct {
    Name         string
    Actors       []Actor
    Messages     []Message
    Vars         []VarDef
    Guards       []GuardDef
    Operators    []Operator
    AdvActions   []AdvAction
    Properties   []Property
    ChannelBound int
    OneShot      bool
}
type Actor struct { /* ... */ }
type Machine struct { /* unexported fields */ }
type State string
type MsgType string
type ActionID string
type GuardID string
type TriggerKind int
type Trigger struct { /* ... */ }
type Transition struct { /* ... */ }
type Message struct { /* ... */ }
type Property struct { /* ... */ }
type PropertyKind int

// Protocol loading
func LoadYAML(path string) (*Protocol, error)
func ParseYAML(data []byte) (*Protocol, error)
func PairingCeremony() *Protocol

// Protocol validation and export
func (*Protocol) Validate() error
func (*Protocol) ExportGo(w io.Writer, pkgName, funcName string) error
func (*Protocol) ExportSwift(w io.Writer) error
func (*Protocol) ExportTLA(w io.Writer) error
func (*Protocol) ExportPlantUML(w io.Writer) error

// Machine runtime
func NewMachine(p *Protocol, actorName string) (*Machine, error)
func (*Machine) RegisterGuard(id GuardID, fn GuardFunc)
func (*Machine) RegisterAction(id ActionID, fn ActionFunc)
func (*Machine) HandleMessage(msg MsgType, ctx any) (State, error)
func (*Machine) Step(ctx any) (State, error)
func (*Machine) State() State

// Pairing ceremony state constants (ServerXxx, AppXxx, CLIXxx)
// ActionID and GuardID constants (ActionSendPairHello, GuardTokenValid, etc.)
// MsgType constants (MsgPairBegin, MsgPairHello, etc.)
```

*Stability: `Machine` API is Stable. Export functions (ExportGo, ExportSwift,
ExportTLA, ExportPlantUML) are Needs Review — generated output format may evolve.*

### `qr/` Go package

```go
func Print(w io.Writer, url string)
func LanIP() string
```

*Stability: Stable.*

### Swift `TernCrypto` package (SPM)

```swift
// E2EKeyPair
public struct E2EKeyPair {
    public init()
    public var publicKeyData: Data
    public func deriveSessionKey(peerPublicKey: Data, info: Data) throws -> SymmetricKey
}

// E2EChannel
public final class E2EChannel: @unchecked Sendable {
    public init(sendKey: SymmetricKey, recvKey: SymmetricKey)
    public convenience init(sharedKey: Data, isServer: Bool)
    public func encrypt(_ plaintext: Data) throws -> Data
    public func decrypt(_ data: Data) throws -> Data
    public enum E2EError: LocalizedError { ... }
}

// Standalone functions
public func deriveKeyFromSecret(_ secret: Data, info: Data) -> SymmetricKey

// Generated state machines (from pairing.yaml)
// ServerMachine, AppMachine, CLIMachine
// MessageType enum, ServerState/AppState/CLIState enums
```

*Stability: E2EKeyPair and E2EChannel are Stable. Generated state machines are
Needs Review — names depend on pairing.yaml actor names.*

---

## Gaps and Prerequisites for 1.0

- **Actor names in pairing.yaml** (`ios`, `cli`) are app-specific. The generated
  Swift classes `AppMachine` and `CLIMachine` are fine for jevon's use but may
  not suit other applications. Consider making actor names configurable, or
  documenting that consumers should define their own protocol YAML.
- **No confirmation code function in Swift** — `DeriveConfirmationCode` exists
  in Go but not in `TernCrypto`. iOS apps currently rely on the backend for
  code display.
- **`protocol.ExportGo` output format** is not yet documented as stable; the
  generated code structure may change if the generator is improved.
- **No published Go module docs** until the first tag is pushed (pkg.go.dev
  indexes on tags).
- **No protocol framework usage example** (`Example_test.go` in `protocol/`).
- **Settling period** (see below): 2-month minimum required after last breaking
  change before 1.0 eligibility.

## Out of Scope for 1.0

- TLS termination at the relay (intended to run behind a proxy or on Fly.io)
- Multi-instance relay (clustering, state sharing)
- Protocol hot-reload without restart
- Bidirectional streaming beyond the current single-client-per-instance model

---

## 1.0 Readiness

**Not yet eligible.** The settling threshold requires 2 months from the last
breaking change (20–50 surface items). Clock starts 2026-03-22 (first public
API surface, v0.1.0). Earliest 1.0 eligibility: 2026-05-22, provided all gaps
above are resolved.

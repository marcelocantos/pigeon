# Stability

This document tracks pigeon's readiness for a 1.0 release — the point at which
backwards compatibility becomes a binding commitment. Once v1.0.0 ships,
breaking changes to any public surface listed below require a major version bump.

The pre-1.0 period (currently v0.x.x) exists to get the interaction surface right.

---

## Interaction Surface Catalogue

*Snapshot as of v0.16.0.*

### Relay API (the binary's external interface)

| Route | Protocol | Response |
|-------|----------|----------|
| `GET /health` | HTTP/3 | `{"status":"ok"}` |
| `GET /register` | WebTransport (QUIC) | First stream message is the assigned instance ID |
| `GET /ws/{id}` | WebTransport (QUIC) | Bridged bidirectionally (streams + datagrams) to registered backend |

Supports both reliable streams (via `Send`/`Recv`) and unreliable datagrams (via `SendDatagram`/`RecvDatagram`).
Relay bridges additional streams opened by either side (for channel API).

Constraints: one client per instance; second client returns HTTP 409.
Max message frame size: 1 MiB.
CORS: `Access-Control-Allow-Origin: *` on health endpoint (for browser Alt-Svc priming).

*Stability: Stable.*

### CLI interface (the `pigeon` binary)

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--port` | string | `""` | Listening port (overrides `PORT` env var) |
| `--cert` | string | `""` | TLS certificate file (PEM); generates self-signed if omitted |
| `--key` | string | `""` | TLS private key file (PEM) |
| `--domain` | string | `""` | Domain name for ACME (Let's Encrypt) certificate provisioning |
| `--acme-email` | string | `""` | Contact email for ACME certificate registration |
| `--lan` | string | `""` | LAN listener address for direct connections (e.g. `:0`, `localhost:44333`) |
| `--version` | bool | `false` | Print version and exit |
| `--help-agent` | bool | `false` | Print usage + agents-guide.md and exit |

Environment variables: `PORT` (default `443`).

Build-time version injection: `-ldflags "-X main.version=<version>"`.

*Stability: Stable.*

### Wire format (encrypted message frame — streams)

```
[8-byte sequence number, little-endian uint64]
[ciphertext (variable length)]
[16-byte AES-GCM authentication tag]
```

The sequence number doubles as both the replay-prevention counter and the
AES-GCM nonce (first 8 bytes of the 12-byte nonce, remaining 4 bytes zero).

*Stability: Stable.*

### Wire format (datagram framing)

Every datagram has a 1-byte prefix for type discrimination and automatic
fragmentation of payloads exceeding the QUIC datagram frame size:

```
0x00 + payload                                — conn: whole datagram
0x40 + frag header (8B) + chunk               — conn: fragment
0x80 + channel ID (2B) + payload              — channel: whole datagram
0xC0 + channel ID (2B) + frag header (8B) + chunk — channel: fragment
```

Fragment header: `[4B msg ID][2B frag index][2B total fragments]`.
Incomplete assemblies are discarded after 5 seconds (configurable).

*Stability: Stable.*

### Root Go package (`github.com/marcelocantos/pigeon`)

```go
// Conn type
type Conn struct { /* unexported fields */ }
func (c *Conn) InstanceID() string
func (c *Conn) Send(ctx context.Context, data []byte) error
func (c *Conn) Recv(ctx context.Context) ([]byte, error)
func (c *Conn) SendDatagram(data []byte) error
func (c *Conn) RecvDatagram(ctx context.Context) ([]byte, error)
func (c *Conn) SetChannel(ch *crypto.Channel)
func (c *Conn) SetDatagramChannel(ch *crypto.Channel)
func (c *Conn) SetPairingRecord(rec *crypto.PairingRecord)
func (c *Conn) OpenChannel(name string) (*StreamChannel, error)
func (c *Conn) AcceptChannel(ctx context.Context) (*StreamChannel, error)
func (c *Conn) DatagramChannel(name string) *DatagramChannel
func (c *Conn) LANReady() <-chan struct{}
func (c *Conn) Close() error
func (c *Conn) CloseNow() error

// StreamChannel type
type StreamChannel struct { /* unexported fields */ }
func (sc *StreamChannel) Name() string
func (sc *StreamChannel) Send(ctx context.Context, data []byte) error
func (sc *StreamChannel) Recv(ctx context.Context) ([]byte, error)
func (sc *StreamChannel) Close() error

// DatagramChannel type
type DatagramChannel struct { /* unexported fields */ }
func (dc *DatagramChannel) Send(data []byte) error
func (dc *DatagramChannel) Recv(ctx context.Context) ([]byte, error)

// Config struct
type Config struct {
    Token        string
    InstanceID   string
    TLS          *tls.Config
    WebTransport bool
    QUICPort     string
    LANServer    *LANServer
    LAN          bool
    LANTLS       *tls.Config
}

// Client-side connectivity (both call WakeRelay automatically)
func Register(ctx context.Context, relayURL string, c Config) (*Conn, error)
func Connect(ctx context.Context, relayURL, instanceID string, c Config) (*Conn, error)
func WakeRelay(ctx context.Context, relayURL string, c Config) error

// LAN server
type LANServer struct { /* unexported fields */ }
func NewLANServer(addr string, tlsConfig *tls.Config) (*LANServer, error)
func (s *LANServer) Addr() string
func (s *LANServer) Close() error

// Server library — WebTransport (browsers)
type WebTransportServer struct { /* unexported fields */ }
func NewWebTransportServer(addr string, tlsConfig *tls.Config, token string) (*WebTransportServer, error)
func NewWebTransportServerWithHub(addr string, tlsConfig *tls.Config, token string, h *hub) (*WebTransportServer, error)
func (s *WebTransportServer) ListenAndServe() error
func (s *WebTransportServer) Serve(conn net.PacketConn) error
func (s *WebTransportServer) Close() error
func (s *WebTransportServer) Addr() net.Addr
func (s *WebTransportServer) Hub() *hub

// Server library — raw QUIC (native clients)
type QUICServer struct { /* unexported fields */ }
func NewQUICServer(addr string, tlsConfig *tls.Config, token string, h *hub) *QUICServer
func (s *QUICServer) ListenAndServe(tlsConfig *tls.Config) error
func (s *QUICServer) ServeWithTLS(conn net.PacketConn, tlsConfig *tls.Config) error
func (s *QUICServer) Close() error
func (s *QUICServer) Addr() net.Addr
```

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

// Channel mode
type ChannelMode int
const (
    ModeStrict   ChannelMode = iota // sequential, no gaps (default)
    ModeDatagrams                   // gaps allowed, replay rejected
)

// Channel construction
func NewChannel(sendKey, recvKey []byte) (*Channel, error)
func NewSymmetricChannel(key []byte, isServer bool) (*Channel, error)
func NewDatagramChannel(sendKey, recvKey []byte) (*Channel, error)

// Channel methods
func (*Channel) Encrypt(plaintext []byte) []byte
func (*Channel) Decrypt(data []byte) ([]byte, error)
func (*Channel) SetMode(mode ChannelMode)

// PairingRecord — persistent pairing state
type PairingRecord struct {
    PeerInstanceID  string `json:"peer_instance_id"`
    RelayURL        string `json:"relay_url"`
    LocalPrivateKey []byte `json:"local_private_key"`
    LocalPublicKey  []byte `json:"local_public_key"`
    PeerPublicKey   []byte `json:"peer_public_key"`
}
func NewPairingRecord(peerInstanceID, relayURL string, localKP *KeyPair, peerPubKey *ecdh.PublicKey) *PairingRecord
func (*PairingRecord) DeriveChannel(sendInfo, recvInfo []byte) (*Channel, error)
func (*PairingRecord) Marshal() ([]byte, error)
func UnmarshalPairingRecord(data []byte) (*PairingRecord, error)
```

*Stability: Stable — `NewSymmetricChannel` may be renamed for clarity before 1.0.*

### `protocol/` Go package

```go
// Core types
type State string
type MsgType string
type ActionID string
type GuardID string
type EventID string
type CmdID string
type TriggerKind int
type Trigger struct { Kind TriggerKind; Msg MsgType; Desc string }
type FairnessKind int // WeakFair, StrongFair
type PropertyKind int // Invariant, Liveness, LeadsTo
type VarType string   // VarString, VarInt, VarBool, VarSetString
type ChannelMode int  // ModeStrict, ModeDatagrams

type Protocol struct {
    Name         string
    Actors       []Actor
    Messages     []Message
    Events       []EventDef
    Commands     []CommandDef
    Structs      []StructDef
    Vars         []VarDef
    Guards       []GuardDef
    Operators    []Operator
    AdvActions   []AdvAction
    AdvGuard     string
    Phases       []Phase
    WireConsts   []WireConstant
    Constants    []ConstantDef
    Properties   []Property
    ChannelBound int
    OneShot      bool
}

type Actor struct {
    Name        string
    Initial     State
    Transitions []Transition
    StateIndex  map[State]*StateNode
    Roots       []*StateNode
}

type Transition struct {
    From, To State
    On       Trigger
    Guard    GuardID
    Do       ActionID
    Fairness FairnessKind
    Sends    []Send
    Updates  []VarUpdate
    Emits    []CmdID
}

// Hierarchy
type StateNode struct {
    Name        State
    Parent      *StateNode
    Children    []*StateNode
    Transitions []Transition
}
func (*StateNode) IsLeaf() bool
func (*StateNode) LeafStates() []*StateNode
func (*StateNode) AncestorChain() []*StateNode
func (*Actor) FlattenedTransitions() []Transition

// Supporting types
type EventDef struct { ID EventID; Desc string }
type CommandDef struct { ID CmdID; Desc string }
type StructDef struct { Name string; Fields []StructField; Desc string }
type StructField struct { Name string; Type VarType; Initial, Desc string }
type VarDef struct { Name string; Type VarType; Initial, Desc string }
type WireConstant struct { Name string; Value any; Type, Desc, Group string }
type ConstantDef struct { Name string; Type VarType; Values []string; Desc string }
type Phase struct { Name string; States []State; Vars []VarDef; Structs []StructDef }
type Property struct { Name string; Kind PropertyKind; Expr, FromExpr, ToExpr, Desc string }
type Machine struct { /* unexported fields */ }

// Protocol loading
func LoadYAML(path string) (*Protocol, error)
func ParseYAML(data []byte) (*Protocol, error)
func PairingCeremony() *Protocol

// Protocol validation and export
func (*Protocol) Validate() error
func (*Protocol) ExportGo(w io.Writer, pkgName, funcName string) error
func (*Protocol) ExportSwift(w io.Writer) error
func (*Protocol) ExportTLA(w io.Writer) error
func (*Protocol) ExportTLAPhase(w io.Writer, phase *Phase) error
func (*Protocol) ExportPlantUML(w io.Writer) error
func (*Protocol) ExportPlantUMLActors(w io.Writer, title string, actors []string) error
func (*Protocol) ExportKotlin(w io.Writer) error
func (*Protocol) ExportTypeScript(w io.Writer) error

// Machine runtime
func NewMachine(p *Protocol, actorName string) (*Machine, error)
func (*Machine) RegisterGuard(id GuardID, fn GuardFunc)
func (*Machine) RegisterAction(id ActionID, fn ActionFunc)
func (*Machine) HandleMessage(msg MsgType, ctx any) (State, error)
func (*Machine) HandleEvent(ev EventID) ([]CmdID, error)
func (*Machine) Step(ev EventID) (State, error)
func (*Machine) State() State
```

*Stability: `Machine` API is Stable (HandleEvent is the preferred entry point;
HandleMessage/Step retained for backward compatibility). Export functions are
Needs Review — generated output format may evolve. Hierarchy types (StateNode,
FlattenedTransitions) are Needs Review — new in v0.12.0.*

### `qr/` Go package

```go
func Print(w io.Writer, url string)
func LanIP() string
```

*Stability: Stable.*

### C client library (`dist/pigeon.h` + `dist/pigeon.c`)

Distributed as an amalgamated single-header/single-source pair. Requires
libsodium for crypto primitives. Zero heap allocations — all state lives in
a `pigeon_ctx` struct sized at compile time.

```c
// Configuration
#define PIGEON_MAX_MSG 1048576  // sole build-time knob

// Crypto types
typedef struct { uint8_t private_key[32]; uint8_t public_key[32]; } pigeon_keypair;
typedef enum { PIGEON_MODE_STRICT, PIGEON_MODE_DATAGRAMS } pigeon_channel_mode;
typedef struct { uint8_t send_key[32]; uint8_t recv_key[32]; uint64_t send_seq, recv_seq; pigeon_channel_mode mode; } pigeon_channel;
typedef struct { char peer_instance_id[64]; char relay_url[256]; uint8_t local_private_key[32]; uint8_t local_public_key[32]; uint8_t peer_public_key[32]; } pigeon_pairing_record;

// Transport abstraction (user-provided callbacks)
typedef struct {
    void *userdata;
    int (*send_stream)(void *, const uint8_t *, size_t);
    int (*recv_stream)(void *, uint8_t *, size_t, size_t *);
    int (*send_datagram)(void *, const uint8_t *, size_t);
    int (*recv_datagram)(void *, uint8_t *, size_t, size_t *);
} pigeon_transport;

// Client context — all library state
typedef struct {
    pigeon_keypair keypair;
    uint8_t peer_pubkey[32];
    pigeon_channel stream_channel, datagram_channel;
    uint8_t hkdf_scratch[96];
    pigeon_pairing_record record;
    pigeon_ios_machine pairing;
    pigeon_transport transport;
    uint8_t read_buf[PIGEON_MAX_MSG];
    uint8_t write_buf[PIGEON_MAX_MSG];
} pigeon_ctx;

// API
void pigeon_init(pigeon_ctx *ctx, const pigeon_transport *transport);
int  pigeon_generate_keypair(pigeon_keypair *kp);
int  pigeon_derive_session_key(const uint8_t *priv, const uint8_t *peer_pub, const uint8_t *info, size_t info_len, uint8_t *out);
int  pigeon_derive_confirmation_code(const uint8_t *pub_a, const uint8_t *pub_b, char *out_code);
void pigeon_channel_init(pigeon_channel *ch, const uint8_t *send_key, const uint8_t *recv_key, pigeon_channel_mode mode);
int  pigeon_channel_init_symmetric(pigeon_channel *ch, const uint8_t *master_key, bool is_server);
int  pigeon_channel_encrypt(pigeon_channel *ch, const uint8_t *pt, size_t pt_len, uint8_t *out, size_t out_len);
int  pigeon_channel_decrypt(pigeon_channel *ch, const uint8_t *data, size_t data_len, uint8_t *out, size_t out_len);
int  pigeon_send(pigeon_ctx *ctx, const uint8_t *data, size_t len);
int  pigeon_recv(pigeon_ctx *ctx, uint8_t *out, size_t out_len);
int  pigeon_send_datagram(pigeon_ctx *ctx, const uint8_t *data, size_t len);
int  pigeon_recv_datagram(pigeon_ctx *ctx, uint8_t *out, size_t out_len);
int  pigeon_frame_message(const uint8_t *payload, size_t len, uint8_t *buf, size_t buf_len);
uint32_t pigeon_read_frame_length(const uint8_t *buf);

// Callback types
typedef bool (*pigeon_guard_fn)(void *ctx);
typedef int  (*pigeon_action_fn)(void *ctx);
typedef void (*pigeon_change_fn)(const char *var_name, void *ctx);

// Generated state machines (per-actor: server, ios, cli)
// Each has: _machine struct, _machine_init(), _handle_message(), _step()
```

*Stability: Fluid — new in v0.16.0, API surface may evolve before 1.0.*

### `protocol/` C code generator

```go
func (*Protocol) ExportCHeader(w io.Writer) error
func (*Protocol) ExportCImpl(w io.Writer) error
```

*Stability: Fluid — new in v0.16.0.*

### Swift `Pigeon` package (SPM)

```swift
// E2EKeyPair
public struct E2EKeyPair {
    public init()
    public var publicKeyData: Data
    public func deriveSessionKey(peerPublicKey: Data, info: Data) throws -> SymmetricKey
}

// ChannelMode
public enum ChannelMode {
    case strict    // sequential, no gaps (default)
    case datagrams // gaps allowed, replay rejected
}

// E2EChannel
public final class E2EChannel: @unchecked Sendable {
    public init(sendKey: SymmetricKey, recvKey: SymmetricKey)
    public convenience init(sharedKey: Data, isServer: Bool)
    public var mode: ChannelMode
    public func encrypt(_ plaintext: Data) throws -> Data
    public func decrypt(_ data: Data) throws -> Data
    public enum E2EError: LocalizedError { ... }
}

// PairingRecord
public struct PairingRecord: Codable, Sendable {
    public init(peerInstanceID: String, relayURL: String, localKeyPair: E2EKeyPair, peerPublicKey: Data)
    public func deriveChannel(sendInfo: Data, recvInfo: Data) throws -> E2EChannel
}

// Standalone functions
public func deriveKeyFromSecret(_ secret: Data, info: Data) -> SymmetricKey

// Generated state machines (from pairing.yaml)
// ServerMachine, AppMachine, CLIMachine
// MessageType enum, ServerState/AppState/CLIState enums
```

*Stability: E2EKeyPair and E2EChannel are Stable. Generated state machines are
Needs Review — names depend on pairing.yaml actor names.*

### `faultproxy/` Go package (testing only)

```go
type Proxy struct { /* unexported fields */ }
type Profile struct { /* see source */ }
type Option func(*Profile)
type Action int   // Forward, Drop
type Stats struct { /* atomic counters */ }

func New(target string, opts ...Option) (*Proxy, error)
func (*Proxy) Addr() string
func (*Proxy) GetStats() *Stats
func (*Proxy) PacketCount() int
func (*Proxy) UpdateProfile(opts ...Option)
func (*Proxy) Close() error

func WithLatency(base, jitter time.Duration) Option
func WithPacketLoss(rate float64) Option
func WithReorder(rate float64) Option
func WithCorrupt(rate float64) Option
func WithBandwidth(bytesPerSec int) Option
func WithBlackhole(duration, interval time.Duration) Option
func WithDropAfter(n int) Option
func WithDropWindow(start, end int) Option
func WithPacketHook(fn func(pktNum int, data []byte) Action) Option
```

*Stability: Fluid — testing utility, API may evolve freely.*

---

## Gaps and Prerequisites for 1.0

- **Actor names in pairing.yaml** (`ios`, `cli`) are app-specific. The generated
  Swift classes `AppMachine` and `CLIMachine` are fine for the reference
  application but may not suit other applications. Consider making actor names configurable, or
  documenting that consumers should define their own protocol YAML.
- **`protocol.ExportGo` output format** is not yet documented as stable; the
  generated code structure may change if the generator is improved.
- **Hierarchy API** (`StateNode`, `FlattenedTransitions`) is new in v0.12.0 and
  may evolve — the PlantUML rendering of hierarchy is not yet complete.
- **No published Go module docs** until the first tag is pushed (pkg.go.dev
  indexes on tags).
- **No protocol framework usage example** (`Example_test.go` in `protocol/`).
- **C library API stabilisation**: New in v0.16.0, the C API surface is Fluid
  and needs real-world usage feedback before freezing. No version macros yet.
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

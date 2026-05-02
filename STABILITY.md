# Stability

This document tracks pigeon's readiness for a 1.0 release — the point at which
backwards compatibility becomes a binding commitment. Once v1.0.0 ships,
breaking changes to any public surface listed below require a major version bump.

The pre-1.0 period (currently v0.x.x) exists to get the interaction surface right.

---

## Interaction Surface Catalogue

*Snapshot as of v0.22.0.*

### Relay API (the binary's external interface)

| Route | Protocol | Response |
|-------|----------|----------|
| `GET /health` | HTTP/3 | `{"status":"ok"}` |
| `GET /register` | WebTransport (QUIC) | First stream message is the assigned instance ID |
| `GET /ws/{id}` | WebTransport (QUIC) | Bridged bidirectionally (streams + datagrams) to registered backend |

Supports both reliable streams (via `Send`/`Recv`) and unreliable datagrams (via `SendDatagram`/`RecvDatagram`).
Relay bridges additional streams opened by either side (for channel API).

Constraints: multiple clients may connect to the same instance ID; the relay
maintains independent bridges for each. Stream/datagram traffic is fanned
out per client.
Max message frame size: 1 MiB.
CORS: `Access-Control-Allow-Origin: *` on health endpoint (for browser Alt-Svc priming).

*Stability: Stable.*

### CLI interface (the `pigeon` binary)

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--port` | string | `""` | WebTransport listening port (overrides `PORT` env var) |
| `--quic-port` | string | `""` | Raw QUIC listening port (overrides `QUIC_PORT` env var) |
| `--cert` | string | `""` | TLS certificate file (PEM); generates self-signed if omitted |
| `--key` | string | `""` | TLS private key file (PEM) |
| `--domain` | string | `""` | Domain name for ACME (Let's Encrypt) certificate provisioning |
| `--acme-email` | string | `""` | Contact email for ACME certificate registration |
| `--lan` | string | `""` | LAN listener address for direct connections (e.g. `:0`, `localhost:44333`) |
| `--cert-validity` | int | `365` | Self-signed certificate validity in days (use ≤14 for browser `serverCertificateHashes`) |
| `--version` | bool | `false` | Print version and exit |
| `--help-agent` | bool | `false` | Print usage + agents-guide.md and exit |

Subcommands:

| Subcommand | Description |
|------------|-------------|
| `pair` | Mint a `PairingArtifact` for a peer instance ID and emit it (default to stdout); writes companion server-side `PairingRecord` to stderr. Flags: `--relay`, `--instance`, `--ttl`, `--token`, `--format=json|text`, `--out`, `--server-record-out`. |

Environment variables: `PORT` (default `443`).

Build-time version injection: `-ldflags "-X main.version=<version>"`.

*Stability: Stable. The `pair` subcommand is Needs Review — added in v0.19.0; flag set may settle further as deploy-script use cases land.*

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

Substantially restructured in v0.21.0. The package now exposes both the new
multi-channel peer API (`Register`/`Connect`/`Session`/`Stream`/`Datagram`/`Listener`)
and retains the legacy relay-dialer/`Conn` path (`DialRelayAcceptor`,
`DialRelayInitiator`) for the pairing ceremony and other single-stream uses.
`PairingArtifact`, `PairingHost`, `CredentialStore`, `IssueCredential`,
`ConnectWithArtifact`, and the old `Register(relayURL, Config)` /
`Connect(relayURL, instanceID, Config)` signatures have been removed from this
package; the artifact/credential surface lives in Swift and Kotlin SDKs only,
or is handled out-of-band.

`crypto.Identity` is a new interface (v0.21.0) providing long-term peer
identity (X25519 ECDH + Ed25519 signing) without exposing private bytes to pigeon.

```go
// --- Multi-channel peer API (new in v0.21) ---

// Identity — long-term peer identity (new in v0.21)
// crypto.Identity interface (in sub-package github.com/marcelocantos/pigeon/crypto)
type Identity interface {
    PublicKey() []byte                                     // X25519 public key (32 bytes)
    InstanceID() string                                    // stable ID derived from public key
    DeriveSharedSecret(peerPub, info []byte) ([]byte, error) // ECDH + HKDF
    Sign(msg []byte) ([]byte, error)                       // Ed25519
}
func NewFileIdentity(path string) (Identity, error)        // file-backed reference implementation

// RegisterArgs / ConnectArgs — structured config replacing the old Config+relayURL pair
type RegisterArgs struct {
    Identity  crypto.Identity
    Pairing   func(clientID string) (*crypto.PairingRecord, error)
    Relay     string
    TLS       *tls.Config
    Token     string
    Datagrams map[string]uint64  // name → pre-agreed channel ID; 0 reserved
}
type ConnectArgs struct {
    InstanceID string
    Record     *crypto.PairingRecord
    Identity   crypto.Identity
    Relay      string
    TLS        *tls.Config
    Token      string
    Datagrams  map[string]uint64
}

// Listener — backend-side acceptor; yields one Session per paired client
type Listener struct { /* unexported fields */ }
func Register(ctx context.Context, args *RegisterArgs) (*Listener, string, error)
func (l *Listener) ID() string
func (l *Listener) Accept(ctx context.Context) (*Session, error)
func (l *Listener) Close() error

// Connect — client-side; returns one Session to the paired backend
func Connect(ctx context.Context, args *ConnectArgs) (*Session, error)

// Session — one end-to-end encrypted peer ↔ peer association
type Session struct { /* unexported fields */ }
func (s *Session) PeerID() string
func (s *Session) OpenStream(ctx context.Context, name string) (*Stream, error)
func (s *Session) AcceptStream(ctx context.Context, name string) (*Stream, error)
func (s *Session) Datagram(name string) *Datagram
func (s *Session) Close() error

// Stream — reliable, ordered, AEAD-encrypted named channel
type Stream struct { /* unexported fields */ }
func (s *Stream) Name() string
func (s *Stream) Send(msg []byte) error
func (s *Stream) Recv(ctx context.Context) ([]byte, error)
func (s *Stream) Close() error

// Datagram — unreliable, pre-declared AEAD-encrypted channel
type Datagram struct { /* unexported fields */ }
func (d *Datagram) Send(payload []byte) error
func (d *Datagram) Recv(ctx context.Context) ([]byte, error)

// --- Legacy relay-dialer path (retained; single-stream for pairing ceremony etc.) ---

// Config — relay connection config for DialRelay* and WakeRelay
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

func WakeRelay(ctx context.Context, relayURL string, c Config) error
func DialRelayAcceptor(ctx context.Context, relayURL string, c Config) (*Conn, error)
func DialRelayInitiator(ctx context.Context, relayURL, instanceID string, c Config) (*Conn, error)

// Conn — legacy single-stream relay connection
type Conn struct { /* unexported fields */ }
func (c *Conn) InstanceID() string
func (c *Conn) Send(ctx context.Context, data []byte) error
func (c *Conn) Recv(ctx context.Context) ([]byte, error)
func (c *Conn) SendDatagram(data []byte) error
func (c *Conn) RecvDatagram(ctx context.Context) ([]byte, error)
func (c *Conn) SetChannel(ch *crypto.Channel)
func (c *Conn) SetDatagramChannel(ch *crypto.Channel)
func (c *Conn) SetPairingRecord(rec *crypto.PairingRecord)
func (c *Conn) OpenStream() (io.ReadWriteCloser, error)
func (c *Conn) LANReady() <-chan struct{}
func (c *Conn) FallbackToRelay()
func (c *Conn) IsDirectActive() bool
func (c *Conn) Close() error
func (c *Conn) CloseNow() error

// LANServer — LAN listener for direct-connect upgrade
type LANServer struct { /* unexported fields */ }
func NewLANServer(addr string, tlsConfig *tls.Config) (*LANServer, error)
func (s *LANServer) Addr() string
func (s *LANServer) CertHash() []byte
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

*Stability: The new multi-channel peer API (Register/Connect/Listener/Session/Stream/
Datagram/RegisterArgs/ConnectArgs) is Fluid — new in v0.21.0 and actively evolving
as T34 folds more of the Go peer library onto the C library. The legacy relay-dialer
path (DialRelayAcceptor/DialRelayInitiator/Conn/Config) is Needs Review — the rename
from Register/Connect in v0.21.0 is breaking. WebTransportServer, QUICServer,
LANServer, and WakeRelay are Stable.*

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
libsodium for crypto primitives.

**Pairing-ceremony FSM (T32, v0.21.0 restructuring):**
The generated state machines now use role-neutral actor names from `pairing.yaml`:
`pigeon_acceptor_machine` / `pigeon_initiator_machine` (replacing the previous
`server/app/cli` decomposition). State enums: `pigeon_acceptor_state` /
`pigeon_initiator_state`. Constants: `PIGEON_ACCEPTOR_*`, `PIGEON_INITIATOR_*`.

**Multi-channel session API (T32):**
`pigeon_session`, `pigeon_stream`, and `pigeon_datagram` provide the post-T22
multi-channel wire: each named stream is a distinct QUIC stream; datagram
channels are pre-declared at session init. `pigeon_ctx` is retained for the
legacy single-stream path; the `pigeon_ios_machine pairing` field has been
removed (callers drive the FSM via the generated machine structs directly).

**Multi-channel wire helpers (T32):**
`pigeon_uvarint_encode`, `pigeon_uvarint_decode`, `pigeon_encode_stream_header`,
`pigeon_decode_backend_stream_header`, `pigeon_decode_client_stream_header`,
`pigeon_encode_datagram`, `pigeon_decode_datagram`.

**PairingRecord serialisation (T32):**
Fixed-schema binary format (420 bytes, magic `PGR\x01`):
`pigeon_pairing_record_serialize`, `pigeon_pairing_record_deserialize`.

```c
// Build-time knob
#define PIGEON_MAX_MSG 1048576

// Crypto types (unchanged since v0.16)
typedef struct { uint8_t private_key[32]; uint8_t public_key[32]; } pigeon_keypair;
typedef enum { PIGEON_MODE_STRICT, PIGEON_MODE_DATAGRAMS } pigeon_channel_mode;
typedef struct {
    uint8_t send_key[32]; uint8_t recv_key[32];
    uint64_t send_seq, recv_seq;
    pigeon_channel_mode mode;
    bool established;
} pigeon_channel;
typedef struct { char peer_instance_id[64]; char relay_url[256];
    uint8_t local_private_key[32]; uint8_t local_public_key[32]; uint8_t peer_public_key[32];
} pigeon_pairing_record;

// Transport abstraction (extended in v0.21 with multi-stream callbacks)
typedef struct pigeon_stream_handle pigeon_stream_handle;  // opaque
typedef struct {
    void *userdata;
    // Legacy single-stream callbacks (still present; NULL in multi-stream-only transports)
    int (*send_stream)(void *userdata, const uint8_t *data, size_t len);
    int (*recv_stream)(void *userdata, uint8_t *buf, size_t buf_len, size_t *out_len);
    int (*send_datagram)(void *userdata, const uint8_t *data, size_t len);
    int (*recv_datagram)(void *userdata, uint8_t *buf, size_t buf_len, size_t *out_len);
    // Multi-stream callbacks (post-T22; NULL if not supported)
    int (*open_stream)(void *userdata, pigeon_stream_handle **out_handle);
    int (*accept_stream)(void *userdata, pigeon_stream_handle **out_handle);
    int (*send_on_stream)(void *userdata, pigeon_stream_handle *handle, const uint8_t *data, size_t len);
    int (*recv_on_stream)(void *userdata, pigeon_stream_handle *handle, uint8_t *buf, size_t buf_len, size_t *out_len);
    int (*close_stream)(void *userdata, pigeon_stream_handle *handle);
} pigeon_transport;

// Legacy single-stream client context (pigeon_ios_machine field removed)
typedef struct {
    pigeon_keypair keypair;
    uint8_t peer_pubkey[32];
    pigeon_channel stream_channel, datagram_channel;
    uint8_t hkdf_scratch[96];
    pigeon_pairing_record record;
    pigeon_transport transport;
    uint8_t read_buf[PIGEON_MAX_MSG];
    uint8_t write_buf[PIGEON_MAX_MSG];
} pigeon_ctx;

// Multi-channel session types (post-T22, new in v0.21)
#define PIGEON_MAX_DATAGRAM_CHANNELS 16
#define PIGEON_MAX_NAME_LEN 64
typedef struct { char name[PIGEON_MAX_NAME_LEN]; uint64_t channel_id; } pigeon_dgchannel_def;
typedef struct { pigeon_channel channel; bool is_backend; uint32_t client_tag;
    pigeon_transport transport;
    pigeon_dgchannel_def datagrams[PIGEON_MAX_DATAGRAM_CHANNELS]; size_t datagram_count;
    uint8_t *scratch_a, *scratch_b; size_t scratch_size;  // heap-allocated, freed by pigeon_session_close
} pigeon_session;
typedef struct { pigeon_session *session; pigeon_stream_handle *handle; char name[PIGEON_MAX_NAME_LEN]; } pigeon_stream;
typedef struct { pigeon_session *session; uint64_t channel_id; char name[PIGEON_MAX_NAME_LEN]; } pigeon_datagram;

// Callback types (unchanged)
typedef bool (*pigeon_guard_fn)(void *ctx);
typedef int  (*pigeon_action_fn)(void *ctx);
typedef void (*pigeon_change_fn)(const char *var_name, void *ctx);

// Generated state machines (pairing ceremony — acceptor/initiator roles)
typedef struct { pigeon_acceptor_state state; /* vars */ pigeon_action_fn actions[PIGEON_ACTION_COUNT]; pigeon_change_fn on_change; void *userdata; } pigeon_acceptor_machine;
typedef struct { pigeon_initiator_state state; /* vars */ pigeon_action_fn actions[PIGEON_ACTION_COUNT]; pigeon_change_fn on_change; void *userdata; } pigeon_initiator_machine;

// Crypto API (unchanged since v0.16)
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

// Multi-channel wire helpers (new in v0.21)
#define PIGEON_MAX_VARINT_LEN 10
#define PIGEON_MAX_STREAM_HEADER (4 + PIGEON_MAX_VARINT_LEN + 256)
int pigeon_uvarint_encode(uint64_t v, uint8_t *buf, size_t buf_len);
int pigeon_uvarint_decode(const uint8_t *buf, size_t buf_len, uint64_t *out);
int pigeon_encode_stream_header(bool is_backend, uint32_t client_tag, const char *name, size_t name_len, uint8_t *out, size_t out_len);
int pigeon_decode_backend_stream_header(const uint8_t *buf, size_t buf_len, uint32_t *client_tag, char *name_buf, size_t name_buf_len, size_t *name_len_out);
int pigeon_decode_client_stream_header(const uint8_t *buf, size_t buf_len, char *name_buf, size_t name_buf_len, size_t *name_len_out);
int pigeon_encode_datagram(pigeon_channel *ch, bool is_backend, uint32_t client_tag, uint64_t channel_id, const uint8_t *payload, size_t payload_len, uint8_t *out, size_t out_len);
int pigeon_decode_datagram(pigeon_channel *ch, bool is_backend, const uint8_t *wire, size_t wire_len, uint32_t *client_tag, uint64_t *channel_id, uint8_t *payload_buf, size_t payload_buf_len);

// Multi-channel session API (new in v0.21)
int  pigeon_session_init(pigeon_session *s, const pigeon_transport *transport, const pigeon_channel *channel, bool is_backend, uint32_t client_tag, const pigeon_dgchannel_def *datagrams, size_t datagram_count);
int  pigeon_session_open_stream(pigeon_session *s, const char *name, pigeon_stream *out_stream);
int  pigeon_session_get_datagram(pigeon_session *s, const char *name, pigeon_datagram *out);
int  pigeon_stream_send(pigeon_stream *s, const uint8_t *msg, size_t msg_len);
int  pigeon_stream_recv(pigeon_stream *s, uint8_t *buf, size_t buf_len);
int  pigeon_stream_close(pigeon_stream *s);
void pigeon_session_close(pigeon_session *s);
int  pigeon_datagram_send(pigeon_datagram *d, const uint8_t *payload, size_t payload_len);
int  pigeon_datagram_recv(pigeon_datagram *d, uint8_t *buf, size_t buf_len);

// PairingRecord binary serialisation (new in v0.21)
#define PIGEON_PAIRING_RECORD_SIZE 420
int pigeon_pairing_record_serialize(const pigeon_pairing_record *rec, uint8_t *buf, size_t buf_len);
int pigeon_pairing_record_deserialize(pigeon_pairing_record *rec, const uint8_t *buf, size_t buf_len);

// Pairing ceremony FSM (new in v0.21 — replaces per-actor server/app/cli machines)
void pigeon_acceptor_machine_init(pigeon_acceptor_machine *m);
int  pigeon_acceptor_handle_message(pigeon_acceptor_machine *m, pairing_ceremony_msg_type msg);
int  pigeon_acceptor_step(pigeon_acceptor_machine *m, pairing_ceremony_event_id event);
void pigeon_initiator_machine_init(pigeon_initiator_machine *m);
int  pigeon_initiator_handle_message(pigeon_initiator_machine *m, pairing_ceremony_msg_type msg);
int  pigeon_initiator_step(pigeon_initiator_machine *m, pairing_ceremony_event_id event);

// Pairing wire driver (new in v0.22 — runs the hello/welcome/confirm exchange end-to-end)
typedef int (*pigeon_confirm_fn)(void *userdata, const char *code);
int pigeon_pair_acceptor(const pigeon_transport *transport, const pigeon_keypair *kp,
                         pigeon_confirm_fn confirm, void *confirm_ud,
                         const char *relay_url, pigeon_pairing_record *out);
int pigeon_pair_initiator(const pigeon_transport *transport, const pigeon_keypair *kp,
                          pigeon_confirm_fn confirm, void *confirm_ud,
                          const char *relay_url, pigeon_pairing_record *out);
```

*Stability: Fluid — substantially extended in v0.21.0 (multi-channel session API,
wire helpers, PairingRecord serialisation, FSM restructuring) and again in v0.22.0
with the C-side pairing wire driver (`pigeon_pair_acceptor`/`pigeon_pair_initiator`)
that runs the hello/welcome/confirm exchange end-to-end against a `pigeon_transport`.
The crypto primitives (`pigeon_channel_*`, `pigeon_keypair`, key derivation) are
stable in content but not yet frozen by commitment. No project-version macros yet
(only protocol-version `PIGEON_PR_VERSION`).*

### C QUIC transport (`c/`, ngtcp2 backend)

Native QUIC transport for C clients (added v0.17, T11.4). Built via CMake
under `c/`, vendoring ngtcp2 and OpenSSL as submodules under `c/vendor/`.
Provides the `pigeon_transport` callbacks expected by the amalgamated
`pigeon_ctx`, so C apps can connect over raw QUIC without implementing the
QUIC stack themselves.

Extended in v0.21.0 (T32 step 3) with multi-stream support:
`pigeon_ngtcp2_transport` now tracks up to `PIGEON_NGTCP2_MAX_EXTRA_STREAMS`
(16) concurrent extra streams in `pigeon_ngtcp2_stream_slot` slots, alongside
the primary stream. The public init/close API is unchanged:

```c
int  pigeon_ngtcp2_transport_init(pigeon_ngtcp2_transport *t, const pigeon_ngtcp2_config *cfg);
void pigeon_ngtcp2_transport_close(pigeon_ngtcp2_transport *t);
pigeon_transport *pigeon_ngtcp2_as_transport(pigeon_ngtcp2_transport *t); // inline helper
```

Also added in v0.21.0: the loopback transport (`c/include/pigeon/loopback.h`) for
in-process unit tests without a QUIC stack, shared with Go (cwire/), Swift (CPigeon),
and Kotlin (JNI):

```c
pigeon_loopback_endpoint *pigeon_loopback_new(void);
void pigeon_loopback_pair(pigeon_loopback_endpoint *a, pigeon_loopback_endpoint *b);
void pigeon_loopback_fill_transport(pigeon_transport *t, pigeon_loopback_endpoint *e);
int  pigeon_loopback_accept_with_header(pigeon_loopback_endpoint *e, pigeon_stream_handle **out_handle, uint8_t *hdr, size_t hdr_len, size_t *hdr_out_len);
void pigeon_loopback_free(pigeon_loopback_endpoint *e);
```

*Stability: Fluid — API and callback contracts still settling before 1.0.*

### `protocol/` C code generator

```go
func (*Protocol) ExportCHeader(w io.Writer) error
func (*Protocol) ExportCImpl(w io.Writer) error
```

*Stability: Fluid — new in v0.16.0.*

### Swift `Pigeon` package (SPM)

The Swift package was refactored in v0.21.0 (T29) from a Network.framework
single-stream client (`PigeonConn`) into a CPigeon C-wrapper with a multi-channel
session API (`PigeonSession`/`PigeonStream`/`PigeonDatagram`). The E2E crypto
types, PairingArtifact, and CredentialStore are unchanged.

`PigeonConn` is still present but is annotated `@available(*, deprecated)`.
`PigeonConn.connect(artifact:)` is also still present under the same deprecation;
new code should use `PigeonSession` with a loopback or ngtcp2 transport.

```swift
// --- E2E crypto (unchanged) ---

public struct E2EKeyPair {
    public init()
    public var publicKeyData: Data
    public func deriveSessionKey(peerPublicKey: Data, info: Data) throws -> SymmetricKey
}

public enum ChannelMode { case strict; case datagrams }

public final class E2EChannel: @unchecked Sendable {
    public init(sendKey: SymmetricKey, recvKey: SymmetricKey)
    public convenience init(sharedKey: Data, isServer: Bool)
    public var mode: ChannelMode
    public func encrypt(_ plaintext: Data) throws -> Data
    public func decrypt(_ data: Data) throws -> Data
    public enum E2EError: LocalizedError { ... }
}

// PairingRecord (JSON wire format snake_case — unchanged since v0.19)
public struct PairingRecord: Codable, Sendable {
    public init(peerInstanceID: String, relayURL: String, localKeyPair: E2EKeyPair, peerPublicKey: Data)
    public func deriveChannel(sendInfo: Data, recvInfo: Data) throws -> E2EChannel
}

// Standalone functions
public func generateNonce() -> Data
public func generateSecret() -> Data
public func deriveKeyFromSecret(_ secret: Data, info: Data) -> SymmetricKey
public func deriveConfirmationCode(_ pubA: Data, _ pubB: Data) -> String

// --- PairingArtifact / CredentialStore (unchanged since v0.19) ---

public let defaultPairingTTL: TimeInterval  // 30 days

public enum PairingError: LocalizedError, Equatable {
    case expired(at: Date)
    case missingField(String)
    case malformedText
}

public struct PairingArtifact: Codable, Sendable {
    public var record: PairingRecord
    public var token: String
    public var issuedAt: Date
    public var expiresAt: Date?
    public init(record: PairingRecord, token: String, issuedAt: Date, ttl: TimeInterval)
    public func isExpired(now: Date) -> Bool
    public func toJSON() throws -> Data
    public static func fromJSON(_ data: Data) throws -> PairingArtifact
    public func toText() throws -> String                   // base64url(json), no padding
    public static func fromText(_ text: String) throws -> PairingArtifact
}

public enum CredentialStoreError: LocalizedError { case noCredential; case backingStore(String) }

public protocol CredentialStore {
    func save(_ artifact: PairingArtifact) throws
    func load() throws -> PairingArtifact
    func delete() throws
    func isExpired() throws -> Bool
}

public final class KeychainCredentialStore: CredentialStore { /* iOS/macOS Keychain */ }
public final class FileCredentialStore: CredentialStore    { /* file-backed fallback */ }

// Deprecated: single-channel Network.framework client
@available(*, deprecated, message: "Use PigeonSession; this predates the post-T22 wire.")
public final class PigeonConn: @unchecked Sendable {
    public let instanceID: String
    public static func register(host:port:token:quicOptions:) async throws -> PigeonConn
    public static func connect(host:port:instanceID:quicOptions:) async throws -> PigeonConn
    public func send(_ data: Data) async throws
    public func recv() async throws -> Data
    public func sendDatagram(_ data: Data) async throws
    public func recvDatagram() async throws -> Data
    public func close()
    public static func wakeRelay(host:port:) async
    // deprecated artifact-driven reconnect:
    public static func connect(artifact:quicOptions:) async throws -> (PigeonConn, E2EChannel)
}

// --- Multi-channel session API (new in v0.21, T29) ---
// Backed by CPigeon (dist/pigeon.h via SPM C-target wrapper).

public protocol PigeonTransport: AnyObject {
    func withCTransport<R>(_ body: (UnsafePointer<pigeon_transport>) -> R) -> R
    var box: AnyObject { get }
}

// In-process loopback transport for tests (no QUIC stack required).
public final class LoopbackTransport: PigeonTransport, @unchecked Sendable {
    public init()
    public static func pair(_ a: LoopbackTransport, _ b: LoopbackTransport)
    public func withCTransport<R>(_ body: (UnsafePointer<pigeon_transport>) -> R) -> R
    public func acceptWithHeader() throws -> (handle: OpaquePointer, header: Data)
}

public struct DatagramChannelDef: Sendable, Hashable {
    public var name: String
    public var channelID: UInt64
    public init(name: String, channelID: UInt64)
}

public enum PigeonSessionError: LocalizedError {
    case sessionInit(Int32); case openStream(Int32); case unknownDatagramChannel(String)
    case streamSend(Int32); case streamRecv(Int32); case streamClose(Int32)
    case datagramSend(Int32); case datagramRecv(Int32)
    case nameTooLong(String); case channelInit; case bufferTooSmall(Int)
}

public final class PigeonSession: @unchecked Sendable {
    public let isBackend: Bool
    public let clientTag: UInt32
    public init(masterKey: Data, isBackend: Bool, clientTag: UInt32,
                datagramChannels: [DatagramChannelDef], transport: PigeonTransport) throws
    public func openStream(name: String) async throws -> PigeonStream
    public func adoptAcceptedStream(name: String, handle: OpaquePointer) -> PigeonStream
    public func datagram(named name: String) throws -> PigeonDatagram
}

public final class PigeonStream: @unchecked Sendable {
    public func send(_ data: Data) async throws
    public func recv() async throws -> Data
    public var messages: AsyncThrowingStream<Data, Error> { get }
    public func close() throws
}

public final class PigeonDatagram: @unchecked Sendable {
    public func send(_ payload: Data) async throws
    public func recv() async throws -> Data?  // nil = datagram not for this channel
}

// Generated state machines (from pairing.yaml — acceptor/initiator as of v0.21)
// PairingCeremonyAcceptorMachine, PairingCeremonyInitiatorMachine
// PairingCeremonyAcceptorState, PairingCeremonyInitiatorState

// Generated state machines (from session.yaml and path_switch.yaml — added v0.17)
// SessionBackendMachine, SessionClientMachine, SessionRelayMachine
// PathSwitchBackendMachine, PathSwitchClientMachine, PathSwitchRelayMachine
```

*Stability: E2EKeyPair, E2EChannel, PairingRecord (including JSON wire format),
PairingArtifact, CredentialStore, PairingError, and the standalone crypto functions
are Stable. The multi-channel session API (PigeonSession / PigeonStream /
PigeonDatagram / PigeonTransport / LoopbackTransport / DatagramChannelDef) is Fluid —
new in v0.21.0; ngtcp2 transport not yet wired to Swift. Generated state machines
are Needs Review — names track YAML actor names. Session/PathSwitch machines are
Fluid — new in v0.17.*

### Kotlin/JVM `Pigeon` library (`android/pigeon/`)

Kotlin/JVM package consumed via JitPack as
`com.github.marcelocantos.pigeon:pigeon:<tag>`. Requires JDK 17+ /
Android API 33+ (for X25519).

Refactored in v0.21.0 (T30) from a native JVM implementation to a JNI wrapper
over `libpigeon-jni` (the C peer library). The legacy `PigeonConn` / `QuicTransport`
interface is retained but deprecated. The new session API lives in
`com.marcelocantos.pigeon.peer` (`Session`, `Stream`, `Datagram`, `Channel`,
`DatagramChannelDef`). Low-level JNI bindings are in `PigeonNative` (internal).

```kotlin
// --- com.marcelocantos.pigeon.crypto (unchanged since v0.19) ---

class E2EKeyPair { ... }
class E2EChannel { ... }
object Hkdf { ... }

// PairingRecord (JSON keys are snake_case — unchanged)
data class PairingRecord(
    val peerInstanceID: String, val relayURL: String,
    val localPrivateKey: ByteArray, val localPublicKey: ByteArray, val peerPublicKey: ByteArray,
)

// PairingArtifact — unchanged since v0.19
val DEFAULT_PAIRING_TTL: java.time.Duration  // 30 days
class PairingExpiredException(val expiresAt: Instant) : Exception
data class PairingArtifact(
    val record: PairingRecord,
    val token: String = "",
    val issuedAt: Instant = Instant.now(),
    val expiresAt: Instant? = null,
) {
    fun isExpired(now: Instant = Instant.now()): Boolean
    fun toJson(): String
    fun toText(): String                                    // base64url(json), no padding
    companion object {
        fun mint(...): PairingArtifact
        fun fromJson(json: String): PairingArtifact
        fun fromText(text: String): PairingArtifact
    }
}

// CredentialStore — unchanged since v0.19
class NoCredentialException : Exception
interface CredentialStore {
    fun save(artifact: PairingArtifact)
    fun load(): PairingArtifact
    fun delete()
    fun isExpired(): Boolean
}
class FileCredentialStore(val path: File) : CredentialStore

// --- com.marcelocantos.pigeon.relay (deprecated, retained for compat) ---

@Deprecated("Use com.marcelocantos.pigeon.peer.Session (T30).")
interface QuicTransport { val inputStream: InputStream; val outputStream: OutputStream; ... }

@Deprecated("Use com.marcelocantos.pigeon.peer.Session (T30).")
class PigeonConn internal constructor(...) { ... }
fun register(transport: QuicTransport, token: String?, host: String?): PigeonConn  // deprecated
fun connect(transport: QuicTransport, instanceID: String, host: String?): PigeonConn  // deprecated

// --- com.marcelocantos.pigeon.peer (new in v0.21, T30) ---
// JNI wrapper over libpigeon-jni; today only the loopback transport is
// available for tests; production ngtcp2 transport pending.

data class DatagramChannelDef(val name: String, val id: Long)

class Channel internal constructor(handle: Long) : AutoCloseable {
    companion object {
        fun directional(sendKey: ByteArray, recvKey: ByteArray, mode: Int = MODE_STRICT): Channel
        fun shared(key: ByteArray, mode: Int = MODE_STRICT): Channel
        const val MODE_STRICT = 0; const val MODE_DATAGRAMS = 1
    }
    fun encrypt(plaintext: ByteArray): ByteArray
    fun decrypt(ciphertext: ByteArray): ByteArray
    override fun close()
}

class Session internal constructor(...) : AutoCloseable {
    val isBackend: Boolean; val clientTag: Int
    suspend fun openStream(name: String): Stream
    fun openStreamBlocking(name: String): Stream
    suspend fun acceptStream(): Stream?
    fun acceptStreamBlocking(): Stream?
    fun getDatagram(name: String): Datagram
    override fun close()
    companion object {
        fun loopbackPair(channelA: Channel, channelB: Channel,
                         aIsBackend: Boolean, aTag: Int,
                         bIsBackend: Boolean, bTag: Int,
                         datagramChannels: List<DatagramChannelDef> = emptyList()): Pair<Session, Session>
    }
}

class Stream internal constructor(...) : AutoCloseable {
    val name: String
    suspend fun send(msg: ByteArray)
    fun sendBlocking(msg: ByteArray)
    suspend fun recv(): ByteArray
    fun recvBlocking(): ByteArray
    val incoming: Flow<ByteArray>
    override fun close()
}

class Datagram internal constructor(...) : AutoCloseable {
    val name: String
    suspend fun send(payload: ByteArray)
    fun sendBlocking(payload: ByteArray)
    suspend fun recv(): ByteArray
    fun recvBlocking(): ByteArray
    override fun close()
}

// Generated machines mirror Swift/Go names with PascalCase PairingCeremony*
// (acceptor/initiator as of v0.21)
```

*Stability: E2EKeyPair, E2EChannel, Hkdf, PairingRecord, PairingArtifact,
CredentialStore, FileCredentialStore, and PairingExpiredException are Needs
Review — wire format settled since v0.19 but the JNI implementation path is
new. The peer-library session API (Session / Stream / Datagram / Channel /
DatagramChannelDef) is Fluid — new in v0.21.0; ngtcp2 production transport
not yet wired in. Generated state machines are Needs Review.*

### `web/` TypeScript package (`@marcelocantos/pigeon`)

Browser-oriented WebTransport relay client and generated state machines
(added v0.17 codegen target). Lives under `web/src/`; tested with
DOM tsconfig and unit test vectors (T31). Provides crypto primitives,
relay connection helpers, and the same generated PairingCeremony/Session/
PathSwitch machines as the other targets.

The relay module (`relay.ts`) was substantially extended in v0.21.0 to
match the post-T22 multi-channel wire: `Session`, `Stream`, `Datagram`,
`connect()`, plus wire helpers. All exports go through `index.ts`.

```typescript
// --- crypto.ts ---
export class E2EKeyPair {
    static async create(): Promise<E2EKeyPair>
    publicKeyData: Uint8Array
    async deriveSessionKey(peerPublicKey: Uint8Array, info: Uint8Array): Promise<Uint8Array>
    async exportPrivateKey(): Promise<Uint8Array>
}
export type ChannelMode = "strict" | "datagrams";
export class E2EChannel {
    static async create(sendKeyBytes: Uint8Array, recvKeyBytes: Uint8Array): Promise<E2EChannel>
    static async fromSharedKey(sharedKey: Uint8Array, isServer: boolean): Promise<E2EChannel>
    mode: ChannelMode
    async encrypt(plaintext: Uint8Array): Promise<Uint8Array>
    async decrypt(data: Uint8Array): Promise<Uint8Array>
}
export interface PairingRecord { peerInstanceID: string; relayURL: string;
    localPrivateKey: string; localPublicKey: string; peerPublicKey: string; }  // base64-encoded keys
export async function generateNonce(): Uint8Array
export async function generateSecret(): Uint8Array
export async function deriveKeyFromSecret(secret: Uint8Array, info: Uint8Array): Promise<Uint8Array>
export async function deriveConfirmationCode(pubA: Uint8Array, pubB: Uint8Array): Promise<string>
export async function createPairingRecord(...): Promise<PairingRecord>
export async function deriveChannelFromRecord(record, sendInfo, recvInfo): Promise<E2EChannel>

// --- relay.ts (new multi-channel API in v0.21) ---
export interface Identity { publicKey: Uint8Array; instanceID: string;
    deriveSharedSecret(peerPublicKey: Uint8Array, info: Uint8Array): Promise<Uint8Array>; }
export interface ConnectArgs { relayURL: string; instanceID: string; identity: Identity;
    record: PairingRecord; datagrams: Map<string, bigint>;
    serverCertificateHashes?: WebTransportHash[]; }
export async function connect(args: ConnectArgs): Promise<Session>
export async function wakeRelay(relayURL: string): Promise<void>

export class Session {
    readonly peerID: string
    async openStream(name: string): Promise<Stream>
    acceptStream(name: string): Promise<Stream>
    datagram(name: string): Datagram
    async close(): Promise<void>
}
export class Stream {
    readonly name: string
    async send(msg: Uint8Array): Promise<void>
    async recv(): Promise<Uint8Array>
    async close(): Promise<void>
}
export class Datagram {
    readonly name: string
    async send(payload: Uint8Array): Promise<void>
    recv(): Promise<Uint8Array>
}

// Wire helpers (new in v0.21 — cross-language pin tested in T31)
export function encodeUvarint(value: bigint): Uint8Array
export function decodeUvarint(buf: Uint8Array): [bigint, number]
export function encodeStreamHeader(name: string): Uint8Array
export function decodeStreamHeader(buf: Uint8Array): [string, number]
export async function encodeDatagram(channel: E2EChannel, channelId: bigint, payload: Uint8Array): Promise<Uint8Array>
export async function decodeDatagram(channel: E2EChannel, wire: Uint8Array): Promise<[bigint, Uint8Array]>
```

*Stability: Fluid — npm publication and module layout still being shaped.
Crypto primitives (E2EKeyPair, E2EChannel, key derivation) match the other
SDKs and are likely stable in content but not yet committed. The multi-channel
session API (Session/Stream/Datagram/connect) is new in v0.21.0.*

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

- **Root Go package API restructuring (T34)**: The v0.21.0 rename of
  `Register`/`Connect` (from relayURL+Config to RegisterArgs/ConnectArgs) is
  breaking. The Go peer library is mid-migration to a cgo wrapper over libpigeon
  (🎯T34); the API shape is not yet stable. Must settle before 1.0.
- **Actor names in pairing.yaml**: The acceptor/initiator rename (v0.21.0) from
  `server/app/cli` decomposition is breaking for generated code consumers. The
  generated Swift/Kotlin/C machine names changed accordingly. Must verify no
  further actor renames are needed.
- **`protocol.ExportGo` output format** is not yet documented as stable; the
  generated code structure may change if the generator is improved.
- **Hierarchy API** (`StateNode`, `FlattenedTransitions`) is new in v0.12.0 and
  may evolve — the PlantUML rendering of hierarchy is not yet complete.
- **No published Go module docs** until the first tag is pushed (pkg.go.dev
  indexes on tags).
- **No protocol framework usage example** (`Example_test.go` in `protocol/`).
- **C library API stabilisation**: Substantially extended in v0.21.0 (session API,
  wire helpers, FSM restructuring); surface is Fluid and needs real-world usage
  feedback before freezing. No project-version macros yet (only protocol-version
  `PIGEON_PR_VERSION`).
- **C QUIC transport (ngtcp2) stabilisation**: Multi-stream callback slots added in
  v0.21.0; transport callback contracts and CMake build layout still settling.
- **Swift multi-channel transport**: `PigeonSession` and `PigeonStream` are backed
  by the loopback transport only; ngtcp2 transport not yet wired into Swift (T29
  part 2 pending). Until the real QUIC transport lands, PigeonSession cannot replace
  PigeonConn in production.
- **Kotlin JNI production transport**: `Session`/`Stream` backed by loopback only;
  ngtcp2 Java-callback transport bridge pending (T30). Until it lands, the peer
  session API is test-only.
- **TypeScript/web package**: Module layout and npm publication still being shaped.
  Wire helpers and Session API are new in v0.21.0.
- **PairingArtifact / CredentialStore (Swift and Kotlin)**: Wire format is stable
  since v0.19.0; the API surface has had minor changes. Needs at least one
  production consumer before freezing.
- **Web package PairingArtifact parity**: The web/TypeScript SDK does not have
  PairingArtifact / CredentialStore equivalents (browser deploy story is QR-only).
  Revisit when the web package stabilises.
- **Settling period** (see below): 2-month minimum required after last breaking
  change before 1.0 eligibility.

## Out of Scope for 1.0

- TLS termination at the relay (intended to run behind a proxy or on Fly.io)
- Multi-instance relay (clustering, state sharing)
- Protocol hot-reload without restart
- Bidirectional streaming beyond the current single-client-per-instance model

---

## 1.0 Readiness

**Not yet eligible.** v0.21.0 contains multiple breaking changes: the root Go
package renamed `Register`/`Connect` (old form removed), removed `PairingArtifact`/
`PairingHost`/`CredentialStore`/`ConnectWithArtifact` from the Go package, and
restructured the C pairing-ceremony FSM actor names (server/app/cli → acceptor/
initiator). These resets the settling clock to 2026-05-01.

Additionally, the Go peer library (🎯T34) is mid-migration to a cgo wrapper over
libpigeon; the Swift (T29 part 2) and Kotlin (T30) production QUIC transports are
not yet wired in. Several surfaces are Fluid with no production consumer yet.

With a 2-month minimum settling period after the last breaking change, earliest
1.0 eligibility is 2026-07-01, provided all gaps above are resolved.

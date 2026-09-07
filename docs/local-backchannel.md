# Local Back-Channel Protocol

> **Status (verified 2026-09-07):** **Implemented (Go)** — the `backchannel/`
> package ships the Unix-socket transport described here:
> `backchannel.Listen` (`backchannel/server.go:44`),
> `backchannel.SocketPath` (`backchannel/path.go:23`),
> `backchannel.ServeSession` (`backchannel/dispatch.go:35`), wire types in
> `backchannel/wire.go`. Check the source for exact signatures; this document
> is the rationale and protocol design, not a generated API reference. No
> Swift/Kotlin/C equivalent exists.

## Problem

The pairing ceremony in `pairing.yaml` defines three actors: **cli**
(the user's terminal), **server** (the daemon), and **ios** (the
mobile app). The server↔ios channel is well-defined: it runs over
the pigeon relay with ECDH encryption, and the state machine governs
every message.

The cli↔server channel is not. The YAML describes five messages
between them (`pair_begin`, `token_response`, `waiting_for_code`,
`code_submit`, `pair_status`) and annotates them with HTTP-like hints
(`POST /api/pair/begin`), but pigeon provides no transport, framing,
or discovery mechanism for this exchange.

Every consumer reinvents this back channel. The common case is
identical across all of them: a CLI binary and a daemon binary on the
same machine, where the daemon runs as a system service (launchd,
systemd, Homebrew service) and the CLI is a one-shot command the user
types. Without a standard, consumers either:

1. **Collapse cli and server into one process** — the CLI command
   registers with the relay itself, does the full ceremony, and saves
   a pairing record that the daemon loads later. This works for the
   happy path but breaks the three-actor model: the daemon isn't
   involved in pairing, re-pairing requires stopping the daemon, and
   the protocol's security properties (which assume the server holds
   state across the ceremony) don't apply as designed.

2. **Invent ad-hoc IPC** — a REST endpoint on localhost, a temp file,
   a named pipe. Each approach has different failure modes, discovery
   mechanisms, and security properties, none of which are formally
   specified or verified.

## Proposal

Add a local back-channel transport to pigeon that:

1. Provides a Unix domain socket server the daemon listens on
2. Defines the wire format for the five cli↔server messages
3. Handles socket path discovery (well-known location)
4. Integrates with the existing pairing state machine — the daemon's
   `HandleMessage` receives `pair_begin` from the socket the same way
   it receives `pair_hello` from the relay

### Socket path

```
$XDG_RUNTIME_DIR/pigeon/<app-name>.sock    (Linux)
~/Library/Application Support/pigeon/<app-name>.sock  (macOS)
```

The app name is supplied by the consumer (e.g., `pairdroid`). The
daemon creates the directory and socket on startup; the CLI connects
to it.

### Wire format

Length-prefixed JSON over the Unix socket, using the same framing as
pigeon's relay stream (`uint32 big-endian length + payload`). Message
types map directly to the YAML message names:

```json
{"type": "pair_begin"}
{"type": "token_response", "instance_id": "abc123", "token": "tok_xyz"}
{"type": "waiting_for_code"}
{"type": "code_submit", "code": "482901"}
{"type": "pair_status", "status": "paired"}
```

No encryption on the Unix socket — it's local-only and protected by
filesystem permissions (0700 directory, 0600 socket).

### API surface

```go
package backchannel

// Server listens on the Unix socket and feeds messages into the
// pairing state machine.
type Server struct { ... }

func NewServer(appName string) (*Server, error)
func (s *Server) Accept(ctx context.Context) (*Session, error)
func (s *Server) Close() error

// Session represents a single CLI connection.
type Session struct { ... }

func (s *Session) Recv(ctx context.Context) (Message, error)
func (s *Session) Send(ctx context.Context, msg Message) error
func (s *Session) Close() error

// Client connects to a running daemon's Unix socket.
type Client struct { ... }

func Dial(appName string) (*Client, error)
func (c *Client) Send(ctx context.Context, msg Message) error
func (c *Client) Recv(ctx context.Context) (Message, error)
func (c *Client) Close() error

// Message is a typed union of the five cli↔server messages.
type Message struct {
    Type       string `json:"type"`
    InstanceID string `json:"instance_id,omitempty"`
    Token      string `json:"token,omitempty"`
    Code       string `json:"code,omitempty"`
    Status     string `json:"status,omitempty"`
}
```

### Integration with the state machine

The daemon's event loop currently receives events from the relay
connection. With this change, it also receives events from the
back-channel socket:

```
  CLI (local)  ──▶ Unix socket ──▶ ┌──────────┐ ──▶ relay ──▶ Mobile
                                    │  Machine  │
  Mobile ──▶ relay ──▶              │           │ ──▶ Unix socket ──▶ CLI
                                    └──────────┘
```

The executor treats back-channel messages as events, the same as
relay messages. `pair_begin` from the socket triggers the same
transition as it does in the YAML. The machine doesn't know or care
which transport delivered the event.

### Lifecycle

1. `pairdroid serve` starts → creates Unix socket, registers with
   relay, enters `Idle` state
2. User runs `pairdroid` → dials Unix socket, sends `pair_begin`
3. Daemon receives `pair_begin` → generates token, registers with
   relay, sends `token_response` back over socket
4. CLI receives `token_response` → renders QR code (relay URL +
   instance ID + token)
5. Mobile scans QR → `pair_hello` arrives over relay
6. Daemon does ECDH → sends `pair_hello_ack` over relay, sends
   `waiting_for_code` over socket
7. CLI displays confirmation code, user enters code → `code_submit`
   over socket
8. Daemon verifies code → sends `pair_complete` over relay, sends
   `pair_status` over socket
9. CLI exits, daemon remains running with the paired device

### Scope

This proposal covers only the back-channel transport and framing.
It does not change the pairing protocol itself, the relay protocol,
or the TLA+ spec. The back-channel is a local transport detail, not
a security-relevant protocol change — the Unix socket is trusted
(same user, filesystem ACL).

### Cross-platform

The design uses Unix domain sockets, which work on macOS and Linux.
Windows support (named pipes) is out of scope for v1 but the API
abstracts the transport, so it can be added later without changing
consumers.

### Future: general-purpose local control

The back-channel pattern extends beyond pairing. Consumers may want
to query daemon status, trigger re-pairing, or disconnect a device —
all via the same Unix socket. The message format is extensible (add
new `type` values), and the socket server can multiplex control
messages alongside pairing. This design doc covers only pairing;
control messages are a natural follow-on.

# Session Protocol

Tern's session protocol manages the entire lifecycle of a connection
between two devices: from initial pairing through ongoing transport
path management. It is defined as a single state machine in
[`protocol/session.yaml`](../protocol/session.yaml) and verified
by TLA+ model checking.

## Overview

The protocol has two phases connected by a single choke point:

1. **Pairing** — ECDH key exchange, confirmation code verification,
   device secret establishment. One-shot: runs once per device pair.
2. **Transport** — relay (permanent baseline) + LAN (direct, optional).
   Continuous: runs for the lifetime of the session, adapting to
   network conditions.

The transition between phases is `SessionActive → RelayConnected`.
For reconnection with a saved `PairingRecord`, the executor starts
directly at `RelayConnected`, skipping Phase 1.

## Actors

Three actors participate across both phases:

| Actor | Role |
|---|---|
| **backend** | The device behind the NAT/firewall. Registers with the relay, starts a LAN server, advertises LAN address. Includes CLI interactions (QR display, code entry) as local actions. |
| **client** | The mobile device. Scans QR, connects via relay, dials LAN when offered. |
| **relay** | The intermediary server. Bridges traffic between backend and client. Permanent baseline — never closes while the session is active. |

## State Machine Diagrams

### Transport (backend + client)

![Transport State Machine](transport.svg)

App I/O transitions (`app_send`, `relay_stream_data`, etc.) are
self-loops on transport states. Commands emitted by transitions
drive the executor.

### Relay

![Relay State Machine](relay.svg)

## Architecture: Machine as I/O Mediator

The state machine mediates between two event surfaces:

```
  App calls ──▶ ┌──────────┐ ──▶ Commands ──▶ I/O writes
                │  Machine  │
  I/O events ─▶ │(generated)│ ──▶ Commands ──▶ App responses
                └──────────┘
```

**Top (application):** `app_send`, `app_recv`, `app_send_datagram`,
`app_recv_datagram`, `app_close`

**Bottom (I/O):** `relay_stream_data`, `lan_stream_data`,
`relay_stream_error`, `lan_stream_error`, `relay_datagram`,
`lan_datagram`, `lan_dial_ok`, `lan_dial_failed`, `ping_timeout`,
`backoff_expired`

The machine is a pure function: **(State, Event) → (State, []Command)**.
The executor is a thin event loop that waits for events from both
surfaces, feeds them to the machine, and executes the returned
commands. No goroutine makes independent state decisions.

### Events

Declared in the `events:` section of session.yaml. Two categories:

- **Application events**: intent from the caller (send data, receive
  data, close connection)
- **I/O events**: completions from the runtime (data arrived on
  socket, connection error, timer expired)

Message receipts (`recv lan_offer`, `recv path_ping`, etc.) are also
events — the stream reader detects the message type and posts the
corresponding event.

### Commands

Declared in the `commands:` section of session.yaml. The `emits:`
field on each transition lists the commands it produces. Categories:

| Category | Examples |
|---|---|
| **I/O writes** | `write_active_stream`, `send_active_datagram`, `send_path_ping` |
| **App delivery** | `deliver_recv`, `deliver_recv_datagram` |
| **Resource lifecycle** | `start_lan_stream_reader`, `stop_monitor`, `close_lan_path` |
| **LAN lifecycle** | `send_lan_offer`, `dial_lan`, `send_lan_verify` |
| **Notifications** | `signal_lan_ready`, `reset_lan_ready` |
| **Crypto** | `set_crypto_datagram` |

### Executor

The executor (`executor.go`) holds the machine, an event channel,
I/O resources, and app response channels. Its `run()` loop:

1. Wait for event from channel
2. Handle app-level events (register waiters for `app_recv`)
3. Feed event to `machine.HandleEvent(ev)` → get `[]CmdID`
4. Execute each command

I/O reader goroutines (relay stream, relay datagram, LAN stream,
LAN datagram) post events — they contain zero state logic. App
methods (`Send`, `Recv`) submit events and block on response channels.

### Boundaries

- **Encryption/decryption**: executor-level (data transform, not
  state decision). Stream reader decrypts and demuxes control
  messages before posting events.
- **Fragment reassembly**: executor-level (stateless data-plane
  logic in `processDatagram`).
- **StreamChannel I/O**: above the machine (per-stream, not routed
  through event loop).
- **DatagramChannel I/O**: through the event loop (needs demux).
- **Relay reader**: always runs (relay is permanent, control messages
  arrive on it even when LAN is active).

## Phase 1: Pairing

The pairing ceremony establishes mutual trust between two devices
that have never communicated before.

### Flow

1. Backend generates a pairing token and registers with the relay
2. Token published via backchannel (QR code, NFC, Bluetooth, manual
   entry — the protocol is agnostic to the mechanism)
3. Client obtains token via backchannel, connects to relay, generates
   ECDH key pair
4. Client sends `pair_hello` with public key + token
5. Backend verifies token, derives shared secret, sends `pair_hello_ack`
6. Both sides independently compute a 6-digit confirmation code from
   the two public keys
7. User reads code from client device, enters on backend CLI
8. Backend verifies code match → stores device, sends `pair_complete`
9. Session established

### Security Properties (verified by TLA+)

- **NoTokenReuse**: revoked tokens are never accepted again
- **MitMDetectedByCodeMismatch**: if an attacker substitutes public
  keys, the confirmation codes differ — the user sees a mismatch
- **MitMPrevented**: if the shared key is compromised, pairing
  never completes
- **AuthRequiresCompletedPairing**: sessions require prior pairing
- **NoNonceReuse**: each authentication nonce accepted at most once
- **DeviceSecretSecrecy**: adversary never learns the device secret

### Adversary Model

The TLA+ spec includes a Dolev-Yao adversary with 8 specific attack
capabilities: backchannel observation, MitM key substitution, secret
re-encryption, concurrent pairing race, token brute-force, code
guessing, and session replay. All attacks are verified to be detected
or prevented by the protocol.

## Phase 2: Transport

After pairing, the transport phase manages which network path carries
traffic. The relay is always connected as a fallback; LAN provides a
direct low-latency path when both devices are on the same network.

### Flow

1. Backend starts a LAN server and sends `lan_offer` via relay
   (includes LAN address + random challenge)
2. Client receives offer, dials the LAN address directly
3. Client sends `lan_verify` with the challenge echoed back
4. Backend verifies challenge, sends `lan_confirm`
5. Both sides switch to the LAN path
6. Health monitor pings the LAN path at fixed intervals
7. If pings fail (3 consecutive), fall back to relay
8. After fallback, wait with exponential backoff, then re-advertise
9. If LAN becomes available again, re-establish

### Resource Lifecycle via Commands

Resource management is not inferred from variable diffs. Each
transition explicitly declares the commands it emits. The executor
mechanically executes them.

Example: `LANOffered → LANActive` (on `recv lan_verify`):

```yaml
emits:
  - send_lan_confirm
  - start_lan_stream_reader
  - start_lan_dg_reader
  - start_monitor
  - signal_lan_ready
  - set_crypto_datagram
```

Example: `LANDegraded → RelayBackoff` (on `ping_timeout`, guard
`at_max_failures`):

```yaml
emits:
  - stop_monitor
  - stop_lan_stream_reader
  - stop_lan_dg_reader
  - close_lan_path
  - reset_lan_ready
  - start_backoff_timer
```

Every resource start has a corresponding stop. The machine's
transition table is the single source of truth for when resources
are created and destroyed.

### App I/O Transitions

Application Send/Recv are self-loop transitions on every transport
state. The machine routes them to the active path:

```yaml
- from: LANActive
  to: LANActive
  on: app_send
  emits: [write_active_stream]

- from: LANActive
  to: LANActive
  on: lan_stream_data
  emits: [deliver_recv]
```

When on relay, `relay_stream_data` delivers. When on LAN, both
`relay_stream_data` and `lan_stream_data` deliver (relay carries
control messages even when LAN is active).

### Backoff Strategy

After falling back to relay, the backend waits before re-advertising
LAN. The delay is exponential: `2^(level-1) × 1s` with ±25% jitter,
capped at level 5 (~16s). The backoff level:

- Resets to 0 on successful LAN establishment
- Increments on offer timeout or max-failure fallback
- Never exceeds `max_backoff_level`

### Transport Properties (verified by TLA+)

- **PathConsistency**: active path is always "relay" or "lan"
- **BackoffBounded**: backoff level never exceeds cap
- **BackoffResetsOnSuccess**: LAN active implies backoff = 0
- **DispatcherAlwaysBound**: dispatchers always on a valid path
- **BackendDispatcherMatchesActive**: backend dispatcher on LAN when
  LAN is active
- **ClientDispatcherMatchesActive**: client dispatcher on LAN when
  LAN is active
- **MonitorOnlyWhenLAN**: health monitor only pings when LAN is
  active or degraded

### Leads-to Properties

- **FallbackLeadsToReadvertise**: after fallback, the backend
  eventually re-advertises LAN
- **DegradedLeadsToResolutionOrFallback**: a degraded LAN path
  eventually either recovers or falls back

## Typed Variables

Every state variable has an explicit type (`string`, `int`, `bool`,
`set<string>`) declared in the YAML. These types are used by code
generators to emit typed structs in each target language. TLA+ ignores
types (it's untyped).

## Structs

Variables that are logically related and updated together are grouped
into structs:

| Struct | Fields | Purpose |
|---|---|---|
| `ECDHState` | backend_pub, client_pub, shared_key, code | ECDH key exchange state |
| `TokenState` | current, active, used | Pairing token lifecycle |
| `BackendPathState` | active_path, dispatcher_path, monitor_target, lan_signal | Backend transport resources |
| `ClientPathState` | active_path, dispatcher_path | Client transport resources |

## Code Generation

The YAML spec is the single source of truth. `protogen` generates:

| Output | What's generated |
|---|---|
| **Go** | Typed machine structs with `HandleEvent(ev) → []CmdID`. Event/command/state/message/guard/action constants. Generic `Machine` executor in `machine.go`. |
| **Swift** | State/message enum constants. Typed machines TBD. |
| **Kotlin** | State/message enum constants. Typed machines TBD. |
| **TypeScript** | State/message enum constants. Typed machines TBD. |
| **TLA+** | Pure TLA+ spec (named actions, UNCHANGED, no PlusCal). Phase-aware export. |
| **PlantUML** | Hierarchical state diagram with phase superstates. |

The Go generator emits `HandleEvent` — a unified entry point that
maps `(state, event, guards) → (new state, variable updates, commands)`.
This replaces the earlier `HandleMessage`/`Step` split. The other
language generators will follow the same pattern.

## TLA+ Verification

The generated TLA+ is verified by TLC (the TLA+ model checker) in
under 1 second:

- **Transport phase**: 121 distinct states, 7 invariants, all pass
- **Pairing phase**: verified separately (with adversary model)

### Channel Elimination

The TLA+ generator uses **channel elimination**: instead of modelling
message passing as sequences (which create combinatorial state space
explosion), each receivable message type becomes a struct variable.
Senders write directly to the struct; receivers guard on it and clear
it after processing. This reduces the state space from millions to
hundreds.

---

## Appendix: The Journey

This appendix chronicles the evolution from ad-hoc implementation to
a formally verified, command-emitting state machine architecture.
It's a record of decisions, wrong turns, and the progressive
discovery that more formalism — not less — makes the system simpler.

### Starting Point: Hand-Written Everything

The pairing ceremony was the first formally modelled protocol. It was
defined in `protocol/pairing.yaml` and generated code for Go, Swift,
Kotlin, TypeScript, and TLA+. The TLA+ spec (`PairingCeremony.tla`)
was generated as PlusCal and verified with TLC. The pairing state
machine was complete and correct.

Everything after pairing — relay connection, message sending, datagram
handling — was ad-hoc Go code. No state machine, no formal model,
no verification.

### LAN Upgrade: The First Transport Feature

When LAN upgrade was added, it was implemented as a `swapTransport`
function on `Conn`. The relay connection was closed; the LAN
connection replaced it. One-shot, one direction, no going back.

This worked for the happy path but created problems:
- No fallback if LAN died
- No health monitoring
- No re-establishment after walking away and coming back

### Path Router: Keeping Both Connections

The design shifted to maintaining both the relay and LAN connections.
A `pathRouter` held a permanent relay path and an optional direct
path. Traffic routed through the best one. Fallback was automatic.

But the path router was ad-hoc Go code. It managed goroutines,
mutexes, and channel (Go channel) dispatch — none of which was
formally specified.

### The Bugs Arrive

Deterministic path-switching tests found 4 bugs:
1. `LANReady` channel not reset on re-establishment
2. Datagram dispatcher reading from dead LAN connection
3. Relay closing backend stream when client disconnected
4. Health monitor ping not failing on closed LAN server

All 4 bugs were in **resource lifecycle management** — goroutines
holding references to the wrong path, signals not reset, dispatchers
not rebound. The state machine (such as it was) described the protocol
correctly. The bugs were in the plumbing that connected the protocol
to the Go runtime.

### Insight: Resource Lifecycle Belongs in the State Machine

The key realisation: if the state machine tracked resource bindings
(which path the dispatcher reads from, which path the monitor pings,
the LANReady signal state), then every transition that changed paths
would be **forced** to update those bindings. The Go code would just
execute the state machine's instructions — no independent decisions.

New state variables were added: `b_dispatcher_path`,
`c_dispatcher_path`, `monitor_target`, `lan_signal`. The TLA+ spec
was updated. TLC verified 17 invariants including:
- `DispatcherMatchesActive`: dispatcher is on LAN when LAN is active
- `MonitorOffOnFallback`: monitor stops when falling back
- `LANSignalPendingOnFallback`: signal resets on fallback

These invariants would have caught all 4 bugs before they were written.

### Per-Actor Variables

TLC found a violation: `DispatcherMatchesActive` failed because the
backend and client switch at different times. The dispatcher path was
a single shared variable, but each actor has its own dispatcher. The
fix: split into `b_dispatcher_path` and `c_dispatcher_path`.

Similarly, `lan_signal` was shared but each `Conn` has its own
`lanReady` channel. The `LANSignalPendingOnFallback` invariant was
removed because valid states exist where actors disagree on the
signal (the backend fell back but the client hasn't processed the
change yet).

### Backoff: Formally Modelled

Exponential backoff for LAN re-advertisement was added as a state
machine concern, not an implementation detail. `backoff_level` became
a state variable. Transitions specified when to increment it (on
fallback), when to reset (on success), and the cap. TLC verified:
- `BackoffBounded`: level ≤ max
- `BackoffResetsOnSuccess`: LAN active ⟹ level = 0
- `FallbackEntersBackoff`: fallback ⟹ level ≥ 1

### Unifying Pairing and Transport

The two protocols (pairing and transport) were merged into a single
YAML (`session.yaml`). The CLI actor from the pairing spec was folded
into the backend as local actions. Actor names were unified: `server`
→ `backend`, `ios` → `client`. The transition `SessionActive →
RelayConnected` bridges the two phases.

A single YAML, one transition table, one executor per language. For
reconnection, the executor starts at `RelayConnected`, skipping
pairing.

### The PlusCal Problem

The TLA+ generator originally produced PlusCal — an imperative
language that compiles to TLA+. PlusCal was chosen by a previous
session because it looks like pseudocode and was easy to generate
mechanically.

The unified spec with PlusCal exploded to hundreds of millions of
states. TLC ran for hours and never finished. Investigation revealed:

- PlusCal introduces process interleaving (4 processes × all branches)
- Program counter variables add state dimensions
- Channel sequences create combinatorial content combinations
- The adversary's knowledge set grows with every eavesdropped message

**Archaeology of the decision**: searching the JSONL transcripts from
the jevon repo revealed that PlusCal was never a deliberate choice.
The previous session fought with PlusCal-specific issues (guard
ordering, operator placement, recv_msg scoping) for multiple
iterations — every one an artifact of PlusCal that doesn't exist in
pure TLA+.

### Rewriting the Generator: Pure TLA+

The generator was rewritten to emit pure TLA+: named actions, primed
variables, UNCHANGED. Each YAML transition maps to one TLA+ action.
No PlusCal, no processes, no program counters.

This was structurally correct but still too large — millions of states
from channel content combinations. The same 5 message types × 3 slots
× 2 channels created tens of thousands of channel state combinations.

### Phase-Aware Export

The generator learned to emit phase-specific specs: only the
transitions, variables, and properties relevant to one phase. The
adversary was restricted to the pairing phase (it has no meaningful
actions during transport). Constants were promoted for variables that
never change within a phase.

This reduced the transport spec from 21 to 12 variables and the
adversary was eliminated entirely for the transport phase.

### Channel Elimination: The Breakthrough

The final insight: channels (sequences) are the state space killer.
The hand-written `PathSwitch.tla` had 979 states because it abstracted
channels. The generated spec faithfully modelled channels and exploded.

The solution: **replace channels with struct variables.** Each
receivable message type becomes a `received_<msg>` struct variable.
Senders write directly to the struct (no Append). Receivers guard on
the struct's type field (no Head) and clear it after processing (no
Tail). No sequences anywhere.

This is semantically different from channels (it models at most one
pending message per type, not a queue). But for the protocols we
verify, this is sufficient — each message type has at most one
in-flight instance.

Result: **121 distinct states, verified in under 1 second.** Down
from hundreds of millions that never finished.

### The OnChange Mistake

With the state machine tracking resource bindings as variables, the
first attempt at wiring the generated machines into `Conn` used
`OnChange` callbacks: when a variable changed (e.g., `b_active_path`
from `"relay"` to `"lan"`), the callback inferred what to do (switch
the active path, restart the dispatcher, etc.).

This created a new class of bugs:
- Deadlocks when `OnChange` tried to lock the same mutex the machine
  was called from
- Missing inferences (stale datagrams not drained because the
  executor didn't know `b_dispatcher_path` changing implied flushing)
- The `Recv` retry hack (if the LAN stream died, Recv had ad-hoc
  logic to fall back and retry instead of letting the machine drive
  the fallback)

The fundamental problem: **the machine declared what state the system
should be in, but the executor guessed how to get there.** Variable
diffs are the wrong abstraction — each diff could imply multiple
actions, and the executor had to infer them.

### The Insight: Machine Emits Commands

The key shift: instead of the machine declaring state and the
executor inferring actions, **the machine explicitly emits commands.**
Each transition's `emits:` field lists the commands the executor
should execute. The executor doesn't infer anything — it mechanically
carries out the command list.

```yaml
- from: LANDegraded
  to: RelayBackoff
  on: ping_timeout
  guard: at_max_failures
  do: fallback_to_relay
  emits:
    - stop_monitor
    - stop_lan_stream_reader
    - stop_lan_dg_reader
    - close_lan_path
    - reset_lan_ready
    - start_backoff_timer
```

This eliminated all the OnChange inference bugs. The command list is
explicit, ordered, and complete. No guessing.

### The Event Loop

With commands explicit, the executor became a simple event loop:

1. Wait for event (from app API or I/O reader goroutine)
2. Feed to `machine.HandleEvent(event)` → get `[]CmdID`
3. Execute each command

Application methods become thin wrappers:
```go
func (c *Conn) Send(ctx context.Context, data []byte) error {
    return c.exec.send(ctx, data)
}
```

I/O reader goroutines contain zero state logic — they read bytes and
post events. The event loop is the single place where transitions
happen.

### Unified Event Space

`HandleEvent` replaced both `HandleMessage` (for received protocol
messages) and `Step` (for internal events). All inputs are events:
- `recv lan_offer` → `EventRecvLanOffer`
- `ping_timeout` → `EventPingTimeout`
- `app_send` → `EventAppSend`
- `relay_stream_data` → `EventRelayStreamData`

One method, one switch, one event space.

### The Final State

| Layer | Formalism | Status |
|---|---|---|
| YAML spec | Single source of truth | `session.yaml` |
| Events | Declared in `events:` section | App surface + I/O surface |
| Commands | Declared in `commands:` section | I/O writes, resource lifecycle, notifications |
| Emits | Per-transition command lists | Every transition declares its commands |
| Phases | Named groupings with scoped vars | Pairing (21 states, 24 vars), Transport (8 states, 16 vars) |
| Types | `string`, `int`, `bool`, `set<string>` | All 40 variables typed |
| Structs | Named variable groups | 4 structs (ECDH, Token, BackendPath, ClientPath) |
| Constants | Parameterised for model checking | Challenges set |
| Fairness | Per-transition weak/strong | `backoff_expired` is strong-fair |
| Properties | Invariants, liveness, leads-to | 14 total (6 security + 8 transport) |
| TLA+ generation | Pure TLA+, channel-free | 121 states, <1s |
| Go code generation | `HandleEvent` + typed machines | Event/command/state constants |
| Other languages | Enum constants (machines TBD) | Swift, Kotlin, TypeScript |
| Executor | Thin event loop in `executor.go` | Conn delegates all I/O |
| PlantUML | Hierarchical state diagram | Phase superstates |

The progression: hand-written code → hand-written TLA+ →
generated PlusCal (failed) → generated pure TLA+ (slow) → generated
pure TLA+ with channel elimination (fast) → OnChange callbacks
(wrong abstraction) → **explicit command emission (correct).**

Each step moved more logic into the state machine. The final step —
commands — moved the last category of ad-hoc logic (resource
lifecycle) into the machine. The executor is now a mechanical
applier of machine instructions. The bugs that prompted this journey
are impossible by construction: the machine's transition table is the
single source of truth for what happens, when, and in what order.

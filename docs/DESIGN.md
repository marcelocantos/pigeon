# Pigeon — Design

This is the top-level architecture document for pigeon. It is intentionally
short: deep specifications live in the per-topic documents under `docs/`.
Read this first to get the conceptual map; then follow the links into the
detailed specs.

If you change the architecture, change this document. If you change a wire
format or state machine, change the relevant deep spec (and the protogen
YAML it derives from). Both should stay in sync — drift is a defect.

---

## 1. Goals and non-goals

**Goal.** Provide a reusable library and relay so that two devices can
establish an end-to-end-encrypted, QUIC-*shaped* session even when neither
has direct network ingress to the other, and even when their best path
between each other changes over the lifetime of the session.

The shape an application sees is a logical *session* between two paired
peers. A session carries:

- one or more reliable, ordered, message-framed *named streams*, and
- one or more unreliable, unordered, message-framed *named datagram
  channels*.

This is QUIC-shaped on purpose — QUIC is the right primitive for
multiplexed peer-to-peer messaging — but the session is not literally a
single QUIC connection. The carrier underneath the session can change; the
session is the stable thing the application holds onto.

**Non-goals.**

- **Generic VPN.** Pigeon is not a routed network — it provides logical
  sessions between specifically paired peer pairs, not arbitrary IP
  reachability.
- **Presence / pubsub / discovery.** Pigeon does not advertise who is
  online, distribute pairings, or do app-layer fanout. Applications are
  expected to layer their own semantics on top.
- **Anonymity.** Pigeon hides session *content* from the relay. It does
  not hide *that two parties are talking* or *which networks they are
  on*. See the threat model.
- **Multi-party sessions.** A session has exactly two peers. Group
  semantics are an application-layer concern.

**Current state vs. target.** Goals reflect the target. The pieces
shipped today are pairing (full), relay-tunnelled session carrier (full,
single-client and multi-client), named streams + named datagram
channels (full), and the **generic candidate-pair sub-FSM** that
underpins LAN-direct and future STUN-reflexive paths
(`protocol/session.yaml` Phase 2, verified by `formal/SessionMachine.tla`
under 🎯T39.7). LAN-direct *carrier code* on the executor side, path
optimisation in the runtime, and `host`-candidate-only STUN gathering
(🎯T43) remain follow-up work — the spec layer is no longer the
bottleneck.

---

## 2. Threat model

**Application data is opaque to the relay; the relay is not invisible.**

What the relay sees:

- That a pair (backend instance ID, client) is talking.
- The client's and backend's *network paths* (source IPs, ports, when
  they change). QUIC migration moves the path; that fact is observable
  to the relay.
- The size, timing, and direction of each stream message and datagram.
- Routing tags (the 4-byte clientTag prefix on multi-client backends).
- The relay greeting (`connect:<id>`, `register-mux:[token]:[id]`).

What the relay does **not** see:

- Application payloads. Every stream message and every datagram payload
  is AEAD-encrypted with a per-session key derived from the
  PairingRecord; the relay sees only ciphertext.
- The contents of the pairing ceremony beyond the ephemeral public keys
  exchanged. (The pairing ceremony itself happens out-of-band — the
  relay is not involved.)
- The peers' long-term identity private keys. Pairing exchanges *public*
  keys; the AEAD channel derives from a fresh ECDH per session. A
  compromised relay cannot decrypt past or future sessions.

**Trusted parties.** The two paired peers trust each other (that's what
pairing means). They do not trust the relay with payload data, nor any
third party.

**Authentication to the relay.** Pigeon authenticates to the relay via
an optional bearer token presented in the relay greeting
(`register-mux:<token>:[<id>]` on the backend side; clients are
authenticated transitively by the backend's pairing callback). This is
out of scope for the design proper — it is a connection-level concern
that should ride on QUIC/TLS rather than appearing in pigeon's
application greeting. The current bearer-in-greeting form is a
provisional placeholder; the target is mTLS or per-connection token
exchange at the QUIC handshake layer (see §9). Whatever the mechanism,
relay auth is orthogonal to the AEAD security between peers — a relay
that authenticates a backend impersonator still cannot decrypt session
data, because the AEAD channel is keyed from the PairingRecord, not
from anything the relay knows.

**Out of scope of the threat model.**

- Side-channel attacks against the host device (memory disclosure,
  malicious peripherals, debugger attachment).
- Application-layer auth/authz between paired peers — pigeon ensures the
  channel is authentic and confidential; the application decides what
  it does with it.
- Adversaries in possession of a peer's long-term identity private key
  (the security model for that scenario is "rotate identity, re-pair").

---

## 3. Layered architecture

Four layers, bottom-up. Each layer's contract is what the layer above
sees; the implementation of the layer is free to change without breaking
the contract. Each layer has a single sharp responsibility.

```
  ┌─────────────────────────────────────────────────────────────┐
  │ L4. Multiplexing        Named streams, per-channel          │
  │                          datagrams, AEAD per-channel keys   │
  │                          — the public application surface   │
  ├─────────────────────────────────────────────────────────────┤
  │ L3. Path management     Upgrade engine — owns N alternative │
  │                          pipes (LAN-direct, STUN-reflexive, │
  │                          ...) above an established session, │
  │                          monitors, promotes, demotes        │
  ├─────────────────────────────────────────────────────────────┤
  │ L2. Session             Per-pipe handshake (auth_request /  │
  │                          auth_ok) that turns a raw pipe     │
  │                          into an AEAD-keyed session.        │
  │                          Pairing is the no-PairingRecord    │
  │                          mode of the same handshake — same  │
  │                          layer, different mode              │
  ├─────────────────────────────────────────────────────────────┤
  │ L1. Relay transport     Remote net.Listen / Accept over     │
  │                          QUIC. Backend registers, calls     │
  │                          listen N times; relay matches      │
  │                          arriving clients to pending        │
  │                          listens; per-client opaque QUIC    │
  │                          pipe. The sole rendezvous —        │
  │                          every session starts here          │
  └─────────────────────────────────────────────────────────────┘
```

**Three architectural invariants pin this stack:**

1. **Identity invariant.** A *session* is identified by its AEAD context
   (derived from the PairingRecord). The pipe carrying the session is
   not part of the session's identity. L3 can swap pipes under the
   session — the AEAD context, the named-stream demux table, the
   datagram-channel table all carry over.
2. **Single-rendezvous invariant.** L1 is the relay, full stop. Every
   session starts on a relay-mediated pipe. There is no path-discovery
   complexity at the entry layer; the relay is the universal
   bootstrap. L3's job is *upgrade only* — promote relay → LAN /
   STUN / ... — never "start on an alternative."
3. **Permanent-baseline invariant.** The relay pipe stays open for the
   lifetime of the session, even when L3 has promoted to a faster
   alternative. Demote-to-relay is instant and free; control messages
   (LAN re-advertise, candidate refresh, ...) ride on the relay pipe;
   re-bootstrapping after NAT shuffles is not necessary. Do not describe
   the relay as a "fallback that goes away when LAN is up."

**Why these layers, not others:**

- **L1 stays sharp by being relay-only.** Folding alternative providers
  (LAN, STUN) into L1 hides huge complexity (NAT traversal, candidate
  gathering, ICE-style negotiation) behind one verb. Keeping L1 = relay
  means L1's contract is small and provable, and pairing — which has
  no `PairingRecord` and therefore can't authenticate alternative
  paths — uses identical L1 plumbing as a fully-paired session does.
- **L2 stays a single layer with two modes** rather than splitting
  pairing and activation into separate layers. Pairing is "L2 with no
  `PairingRecord`, runs the pairing ceremony, produces a record."
  Activation is "L2 with a `PairingRecord`, runs auth_request/auth_ok,
  produces an AEAD context." Both consume an L1 pipe and emit an
  L2-established session. One layer, two entry conditions.
- **L3 is the upgrade engine, not a chooser.** L3 never decides "start
  on LAN"; the session is already running on relay (L1+L2). L3
  discovers candidate alternatives (the relay pipe is the signalling
  channel for candidate exchange), validates them (the AEAD context
  L2 established authenticates the LAN/STUN challenge handshakes),
  promotes, monitors, demotes. The promotion direction is always
  relay → alternative; demote direction is alternative → relay.
- **L4 doesn't know which pipe is active.** It says "open a stream
  named X" or "send a datagram on channel Y"; L3 routes onto whichever
  pipe is currently nominated. A path swap is transparent to L4 by
  design — application-visible swaps would force every consumer to
  carry handling code for them.

**Two state machines, not four.** The four-layer diagram above shows
*responsibilities*, not state-machine count. There are exactly two
state machines in pigeon, both protogen-generated:

- **Pairing machine** (`protocol/pairing.yaml` → `PairingCeremony.tla`)
  — runs once per device pair; produces a `PairingRecord`.
- **Session machine** (`protocol/session.yaml` Transport phase →
  `SessionMachine.tla`) — runs once per logical session; encompasses
  the L2 activation handshake, L3 path management, and health
  monitoring as sub-states of one FSM.

L3 path management is a *sub-FSM of the session machine*, not a
separate top-level state machine. Layering it separately in the
diagram reflects the responsibility split (L3 operates on "which pipe";
L2 operates on "AEAD + naming") — but in execution they share one
executor and one event loop. The two machines join through the
`ValidPairingRecord(r)` interface contract: the pairing machine proves
it as a postcondition; the session machine assumes it as an axiom over
its initial state.

**Per-layer notes:**

L1. **Relay transport.** Single bridging code path on the relay: an
   incoming client connection is matched to one of the backend's
   pending `listen` calls and the two QUIC connections are
   byte-shovelled together. Each accepted client gets its own
   end-to-end QUIC connection (backend has K outstanding listens →
   K independent client pipes). No per-client tag prefix, no demux
   layer, no shared QUIC connection between clients. Built on top of
   the per-language QUIC library (`quic-go` in Go, `ngtcp2` in C,
   `Network.framework` in Swift, etc.) — pigeon does not reimplement
   QUIC. The wire greeting is one verb set: `register` (backend
   announces its identity), `listen` (backend opens a slot), `connect`
   (client requests bridging to a known backend identity).

L2. **Session establishment.** Per-pipe handshake. Two modes
   distinguished by whether the caller supplies a `PairingRecord`:
   - *Activation mode* (record supplied): runs the auth_request /
     auth_ok exchange specified in `protocol/session.yaml`; on success
     hands the layer above an AEAD-keyed session.
   - *Pairing mode* (no record): runs the pairing ceremony specified
     in `protocol/pairing.yaml`; on success produces a fresh
     `PairingRecord` for both sides to persist. The session layer
     above either tears down (pairing-only ceremony) or proceeds to
     activation on a subsequent pipe.

L3. **Path management.** The upgrade engine. Owns the relay pipe
   (always) plus zero-or-more alternative pipes (LAN candidates from
   local interfaces, future STUN-reflexive candidates). Discovers
   candidates by exchanging them with the peer over the existing
   AEAD-keyed session. Validates candidates with a
   challenge-response handshake authenticated by the L2 AEAD
   context. Probes connectivity, nominates the best working
   alternative, and routes L4's stream/datagram surface onto the
   nominee. Monitors with periodic ping/pong; on degradation, demotes
   to the next-best alternative or back to the relay; on relay-only
   stability windows, periodically re-probes for upgrades.
   See `docs/session-protocol.md` for the full FSM.

L4. **Multiplexing.** The application surface — `OpenStream("name")`,
   `AcceptStream("name")`, `Datagram("name")`. Named streams ride on
   QUIC streams within the L3-active pipe; per-channel datagrams ride
   on QUIC datagrams with channel-id framing. AEAD per-channel keys
   are derived from the L2-established `PairingRecord`. Same
   semantics across languages; idiomatic per-language surface.

**Current state vs. target.** L1 currently has two parallel bridging
modes (`bridgeClientPair` for 1:1, `bridgeClientMux` for N-clients
with per-client tag prefix) and pigeon has two parallel session-
shaped types (legacy `Conn` carrying L2+L3 plumbing; modern
`Session` carrying L4 over a thinner L1 with no L3). The target above
collapses both: single bridging mode at L1 (mux subsumed into the
remote-Listen model), single session type carrying the L4 surface
above an L3 layer that explicitly upgrades from relay. Migration is
tracked under T39 sub-targets. Pairing's transition from
`pigeon.Conn` to the modern API rides on L2's pairing-mode landing.

---

## 4. Wire protocols — declarative, code-generated

**Principle: anything that goes on the wire is specified in protogen and
code-generated for every target language. Nothing on the wire is hand-
rolled per language.**

Why, in order of importance:

1. **Formal verification.** A protocol can only be model-checked if it
   has a formal specification. Protogen emits TLA+ alongside every
   language target, so every state machine in pigeon is verifiable by
   TLC against properties expressed in the same source: no token reuse,
   MitM detection via code mismatch, path consistency, backoff bounded,
   and so on (see `docs/session-protocol.md` for the full property
   list). Verification catches whole classes of protocol bugs that
   testing cannot reach — race conditions, adversary-driven state
   sequences, liveness violations under message reordering. This is the
   primary motivation, and the one that demands rigour: a hand-rolled
   per-language implementation cannot be verified, full stop.
2. **Multi-language consistency.** As a corollary of (1), every SDK
   that consumes a generated executor speaks the same wire by
   construction. Each hand-rolled wire helper would be N×language
   copies of the same logic, each capable of drifting; drift produces
   silent interop failures that only surface when two peers run
   different SDKs. Generation from a single spec eliminates the class.
3. **Documentation lockstep.** Protogen also emits PlantUML diagrams
   and the TLA+ spec is itself self-documenting. The diagram is never
   out of date because it is generated from the same source as the
   code.

Protogen's input language describes:

- **Multi-state FSMs with transitions.** Pigeon has two of these: the
  *pairing machine* and the *session machine*. The session machine
  contains path selection (§8) and health monitoring as integrated
  sub-states. Output: an executor for each target language,
  encoders/decoders for every transitionable message, a TLA+ spec, a
  PlantUML diagram.
- **One-shot byte formats.** Examples: relay greeting strings, the
  4-byte clientTag prefix, the per-stream `[varint name-len][name]`
  binding header. Output: encoder + decoder for each target language.

**Two state machines, verified independently.** Pairing and session
have completely different lifecycles — pairing is one-shot per device
pair (lasts the seconds of the ceremony); session is many-shot per pair
(one per logical connection). They communicate via a persistence
boundary: pairing's output (PairingRecord) is persisted via
`CredentialStore`, and that record becomes session's input. So in code
they are two distinct executors.

**Do not compose them into a single TLA+ spec for verification.** A
prior attempt did exactly this and TLC's state space blew up — the
combinatorics of pairing's adversary model multiplied with the
session's path-selection and health-monitor sub-states is intractable
for finite model checking, even at small parameter sizes. Composition
is the wrong tool for cross-machine reasoning here.

**Use interface contracts for cross-cutting properties.** A property
like "no session reaches RelayConnected without prior successful
pairing" decomposes into two local obligations:

- *Pairing spec proves:* "every reached final-success state produces a
  PairingRecord satisfying invariant `ValidPairingRecord(r)`."
- *Session spec assumes:* "session is initialised with an `r`
  satisfying `ValidPairingRecord(r)`" and proves its own properties
  modulo that assumption.

The cross-machine property is then a one-line argument by
modus-ponens, not a model check. Each spec stays tractable for TLC
because it explores only its own state space. The interface contract
(`ValidPairingRecord`) is the join point and the only thing both specs
need to agree on.

**When composition is genuinely unavoidable** (e.g., a property over a
liveness sequence that crosses both machines), parameterise the
composed spec aggressively: bound message channels to length 1-2,
symmetry-reduce identical components, and treat the result as a
sanity check rather than primary verification. Independent
verification stays the default.

**Concretely for pigeon:** `formal/PairingCeremony.tla` exists today.
A separate `formal/SessionMachine.tla` will exist when the session
machine is in YAML. There is no plan for a `formal/Joined.tla`. The
"joined machine" wording in `docs/session-protocol.md` describing the
two phases predates the state-space-explosion incident and should be
read as a *conceptual* join (the PairingRecord hand-off), not an
operational or model-checking fusion.

Both shapes go through the same pipeline: `protocol/*.yaml` →
`cmd/protogen` → `protocol/*_gen.go`, `Sources/Pigeon/*Machine.swift`,
`android/pigeon/.../*Machine.kt`, `c/src/*_gen.c`, `web/src/*Machine.ts`,
`formal/*.tla`, `docs/*.puml`.

**Catalogue of wire interactions** (some specified, some pending spec).
Items belonging to the *session machine* are folded into one row; their
sub-states are listed for clarity but they are not separate FSMs.

| Interaction                  | Source-of-truth status         |
|------------------------------|--------------------------------|
| **Pairing machine**          | `protocol/pairing.yaml` ✓; verified by `formal/PairingCeremony.tla` ✓ |
| **Session machine** (whole)  | `protocol/session.yaml` ✓; verified by `formal/SessionMachine.tla` (Transport phase, 🎯T39, 🎯T39.7) — runtime consumption is partial; `api.go` still drives the per-pipe I/O directly while the YAML spec is the source of truth for protogen-emitted machines in every SDK |
| &nbsp;&nbsp;⊢ session-open hello/ack | Sub-state — currently hand-rolled JSON (`api.go::connectHello/connectAck`); folding onto the YAML machine is tracked under 🎯T39 sub-targets |
| &nbsp;&nbsp;⊢ path selection / candidate exchange | Sub-state ✓ — generic `Candidate` / `candidates` / `pair_check` / `pair_check_ack` FSM in `session.yaml` Phase 2 (🎯T39.7); LAN-direct is candidate kind `host`, STUN `srflx` plugs in additively (🎯T43) |
| &nbsp;&nbsp;⊢ health monitor ping/pong | Sub-state ✓ — `path_ping` / `path_pong` + degraded/backoff states in `session.yaml` Phase 2 |
| Cross-machine properties     | Interface contract (`ValidPairingRecord`) + local proofs in each spec; **no composed TLA+ spec** (state space explodes) |
| Relay greeting variants      | `protocol/wireformats.yaml::relay_greeting` ✓ (protogen-generated encoder/decoder in every SDK; T40) |
| Stream-name binding header   | `protocol/wireformats.yaml::stream_header` ✓ (protogen-generated; T40) |
| Datagram channel-id framing  | `protocol/wireformats.yaml::datagram_plaintext` ✓ (protogen-generated plaintext; AEAD wrapping stays hand-rolled per language — see §5) |

**Current state vs. target.** The pairing ceremony is the only wire
interaction currently flowing through protogen. Everything else is
hand-rolled in each language and kept in sync by careful review and
shared byte-vector tests (`wire_vectors_test.go` ↔ `c/test/test_pigeon.c`).
The byte-vector lockstep does catch drift, but the *class* of bug is
still possible. The target is for every entry in the catalogue above to
be a protogen spec.

The session+transport FSM (joined pairing + path-selection state
machine) is now in protogen YAML (🎯T39 + 🎯T39.7) and verified by
TLA+. The remaining gap is *runtime consumption*: `api.go` and the
per-SDK executors still drive per-pipe I/O directly rather than
dispatching events through the generated machine. Folding the
executor onto the generated FSM is incremental work tracked under
the T39 sub-targets.

See `docs/session-protocol.md` for the authoritative state-machine
design (state hierarchy, transition table, TLA+ verification properties).

---

## 5. Runtime concerns — hand-written per language

These are the parts that are *not* wire-bearing and so do not belong in
protogen. They are written natively in each language because they
interact with language-specific runtime concerns (memory, threading,
sockets, FFI).

- **QUIC transport bring-up.** UDP socket, TLS handshake (with
  ALPN `pigeon`), QUIC handshake. Every SDK does this once per
  carrier connection.
- **AEAD primitives.** `pigeon_channel_encrypt`/`decrypt` (libsodium in
  C, `crypto/cipher` in Go, CryptoKit in Swift, etc.). Stateless
  functions; no FSM. The wire byte layout (8-byte sequence + ciphertext
  + tag) is canonical and does belong in spec — but the *implementation*
  is per-language by necessity.
- **Demux tables and queues.** Stream name → pending accept-stream
  waiter, datagram channel-id → per-channel queue. Pure data
  structures; no protocol states. (Per-client demux tables on the
  backend disappear in the §3 L1 model — each accepted client gets
  its own end-to-end QUIC pipe rather than sharing one with N
  siblings — but the per-stream-name and per-channel-id maps remain.)
- **Length-prefix framing.** Read/write 4-byte big-endian length + body
  on stream messages. One function per direction; no protocol.
- **Listener accept loop.** A blocking switch statement in C; a
  goroutine reading off a channel in Go. The shape is the same; the
  control-flow primitive differs.
- **Concurrency model.** Single-threaded blocking in C (callers add
  external locking if they want concurrent use). Goroutines + channels
  in Go. Async/await on Swift and Kotlin. Same semantics, idiomatic per
  language.
- **Memory model.** Heap-allocated scratch buffers in C (per-session,
  PIGEON_MAX_MSG-sized; see T38 in the audit log for why per-call
  stack scratch overflows host runtimes). Garbage-collected everywhere
  else.
- **Local backchannel** (proposed, not implemented). A Unix socket
  between the CLI and a local daemon during the pairing ceremony.
  See `docs/local-backchannel.md`.

**Current state vs. target.** All of these are present in shipping
SDKs to varying degrees. The C side has the most explicit boundary
because it's the one where layers are most visible (vtables instead of
interfaces). The local backchannel is a planned addition.

---

## 6. Application API shape

Same conceptual surface across languages, idiomatic per language.

**The four user-facing types:**

- `Listener` — backend-side; produced by `Register`. `Accept` returns
  the next paired client as a `Session`.
- `Session` — the logical, AEAD-bound, two-peer entity. Owns named
  streams and named datagram channels. Identified by the AEAD channel
  derived from the PairingRecord. Hot-swappable carrier underneath.
- `Stream` — opened by name; reliable, ordered, message-framed,
  AEAD-encrypted.
- `Datagram` — pre-declared by name → channel-id at session
  construction; unreliable, unordered, message-framed, AEAD-encrypted.

**Pairing API.** Detailed in `docs/pairing-lifecycle.md`. A
`PairingArtifact` packages a `PairingRecord` + bearer token + expiry;
`CredentialStore` provides the per-platform persistence boundary
(Keychain on iOS, etc.); `ConnectWithArtifact` checks expiry before any
network I/O and surfaces a typed `ErrPairingExpired`.

**Known asymmetries.**

- Swift's `ConnectWithArtifact` returns `(PigeonConn, E2EChannel)`
  rather than wiring the channel into the conn the way Go and Kotlin do.
  See `docs/pairing-lifecycle.md`.

**Current state vs. target.** The Application API is the most stable
layer in shipping pigeon. The named-stream / named-datagram shape is
locked in. The asymmetry above is known and tracked; full parity is a
target but not a release blocker.

---

## 7. Multi-language strategy

One pigeon, six languages on the wire (Go, Swift, Kotlin, C, TS for
the web client, plus TLA+ for verification). The strategy is:

1. **One spec, many generators.** The protogen pipeline is the only
   place a wire interaction is described. Generators emit per-language
   code — never the other way around.
2. **Byte-vector lockstep.** Every wire format has tests that pin its
   byte representation (`wire_vectors_test.go` ↔ `c/test/test_pigeon.c`,
   etc.). Drift is caught at CI time.
3. **`make bullseye` is the durable green signal.** Everything builds
   and tests across all six SDKs (or fails loudly).
4. **Hand-written code is allowed for non-wire concerns** (§5). Native
   idiom wins where the language's strengths matter.

**Audit trail.** `docs/audit-log.md` records significant decisions and
the targets they retired.

**Current state vs. target.** Strategy is in effect. The gap is the
proportion of wire-bearing code that is currently hand-rolled (§4) vs.
protogen-generated. Closing that gap is the largest open architectural
task.

---

## 8. Path optimisation (L3)

**Path management is L3** in the four-layer stack (§3) and is structured
as a sub-FSM of the session machine, not a third top-level machine.
This section describes its responsibilities and shape; the
authoritative spec for the integrated session machine lives in
`docs/session-protocol.md`.

**The problem.** QUIC's connection migration handles "my address
changed; keep the connection going" (RFC 9000 §9). It does *not*
handle "a better route exists between the same two peers." If client
and backend both end up on the same LAN, QUIC alone has no concept of
"bypass the relay" — the topology stays `client ↔ relay ↔ backend`.
Discovering and switching to better paths is L3's job.

**L3 is an upgrade engine, not a chooser** (see the single-rendezvous
invariant in §3). The session is already running on the relay pipe by
the time L3 enters the picture; L3 never decides "start on LAN
instead." The only decisions L3 makes are *promote* (relay → faster
alternative), *demote* (alternative → relay), and *re-probe* (look for
better options on a bounded schedule).

**The shape of the FSM** (ICE-like, simpler):

- **Candidate gathering.** Each peer enumerates reachable alternatives
  to the already-established relay pipe: LAN host candidates from
  local interfaces, future STUN-reflexive candidates (see
  `docs/investigations/stun-hole-punching.md`). The relay itself is
  not a "candidate" — it is the always-present baseline pipe and the
  signalling channel for everything else.
- **Candidate exchange.** Peers swap candidate lists over the existing
  AEAD-keyed session (i.e. on the relay pipe), on a dedicated control
  stream. The L2 AEAD context authenticates the exchange; an attacker
  on the LAN can't substitute candidates without first compromising
  the pairing.
- **Connectivity checks.** Pair the candidates, probe each pair
  (challenge-response handshake on the candidate pipe, MAC-verified
  with the L2 AEAD context — same authentication source as candidate
  exchange).
- **Nomination.** Pick the best working pair; promote it. L4's stream
  / datagram surface routes onto the nominee from this point.
- **Demotion.** If the promoted pipe degrades (3 ping failures →
  degraded → backoff schedule), L4 traffic returns to the relay
  pipe immediately. The relay stays up the entire time; demotion is
  zero-cost (per the permanent-baseline invariant in §3).
- **Periodic re-probe.** When sitting on relay alone, periodically
  re-attempt promotion on a backoff-bounded schedule as conditions
  change.

**Pipe-swap semantics.** Open question, not yet decided. Three
candidate strategies for what happens to in-flight L4 streams when L3
swaps the active pipe:

- **(a) Drain-and-resume** — quiesce open streams, swap, reopen.
  Simplest; introduces latency at swap.
- **(b) Replay window** — track unacked messages per stream, replay on
  the new pipe post-swap. Cleaner UX; more bookkeeping.
- **(c) Application-visible** — surface the swap to the app, let it
  decide.

(a) is the recommended v1 default; (b) is a credible v2; (c) is a
foot-gun unless absolutely required by an application class. The
decision point is deferred until L3 is implemented in code.

**ICE-related future work.** STUN-reflexive candidate gathering for
NAT traversal is a credible third tier between LAN-direct and relay.
~70-80% Wi-Fi coverage with simple STUN; mobile cellular with symmetric
NAT requires full ICE+TURN. See `docs/investigations/stun-hole-punching.md`
for the analysis. STUN candidates fit the same upgrade model as LAN —
the relay session is the signalling channel; the AEAD context
authenticates the connectivity check.

**Current state vs. target.** L3's *spec* layer is fully landed: the
generic candidate-pair sub-FSM lives in `protocol/session.yaml`
Phase 2 (🎯T39.7) and is verified by `formal/SessionMachine.tla`.
LAN-direct is candidate kind `host`; STUN-reflexive (`srflx`) is the
natural second kind and slots into the same `candidates` set / pair-
check handshake without further FSM work (🎯T43). The remaining gap
is *runtime*: the relay-only path is still the only one shipping;
LAN-direct candidate gathering, dial, pair-check, and nomination all
need to land on the executor side (`executor.go`, `lan.go`, the
`pathRouter` inside `Conn`) consuming the generated machine rather
than running ad-hoc Go code.

---

## 9. Open questions and current-vs-target gaps

**Largest open architectural gaps:**

1. **L1 has two parallel bridging modes today.** `bridgeClientPair`
   (1:1, used by pairing) and `bridgeClientMux` (N-clients with
   per-client tag prefix, used by sessions) are parallel
   implementations of what should be a single `register / listen /
   accept / connect` model with each accepted client getting its own
   end-to-end QUIC pipe (§3 L1). Collapsing the two deletes the
   `clientTag` concept across the entire codebase, the dual greeting
   parser on the relay, and the legacy `pigeon.Conn` type. Tracked
   under T39 sub-targets.
2. **L2 doesn't yet support pairing-mode.** Today's `Register` /
   `Connect` always force the auth_request / auth_ok handshake, which
   needs a `PairingRecord`. Pairing therefore can't use the modern
   API and stays on `pigeon.Conn`. Making L2 skippable when no
   `PairingRecord` is supplied collapses the pairing-vs-session
   distinction and is a precondition for `Conn` deletion.
3. **L3 unimplemented as a clean layer.** The path-management *spec*
   is complete — the generic candidate-pair FSM (`Candidate` records,
   `candidates` / `pair_check` / `pair_check_ack` exchange, nomination
   over the pair set, AltActive/AltDegraded/RelayBackoff lifecycle)
   landed in `protocol/session.yaml` under 🎯T39.7 and verifies under
   `formal/SessionMachine.tla`. What remains is the *runtime* — folding
   `executor.go`, `lan.go`, and `pathRouter` onto the generated machine
   so that LAN-direct (candidate kind `host`) actually drives traffic
   end-to-end. STUN-reflexive (`srflx`) plugs into the same machine
   additively (🎯T43).
4. **Pipe-swap semantics undecided.** §8 lists the three candidate
   strategies; v1 needs to commit to one before L3 ships.
5. **Wire interactions not yet protogen specs.** ~~Stream-name binding
   header, datagram channel-id framing, relay greeting variants.~~
   T40 retired these into `protocol/wireformats.yaml` (the one-shot
   byte-format pipeline alongside the FSM pipeline). The remaining
   gap is the session activation hello/ack (sub-state of T39's
   session machine), tracked under that target.
6. **Cross-language API parity.** Swift / Kotlin / C are all behind Go
   on at least one application-API edge (Swift's `setChannel`
   asymmetry is the documented case; others may exist).
7. **Relay authentication mechanism.** Currently a bearer token in the
   `register-mux` greeting. Should move to QUIC/TLS layer (mTLS or
   token exchange at handshake time) — application-greeting auth is
   the wrong layer. Captured in §2; the implementation change is
   bounded once the relay's QUIC config grows a TLS verify hook.

**Future explorations** (out of scope for v1, captured here so they're
not lost):

- **Cascading relay topology** — local relay acting as both backend to
  a public relay and relay to game servers. See
  `docs/cascading-relay.md`.
- **BLE proximity oracle** — Bluetooth RSSI as additive evidence
  alongside the pairing confirmation code. See
  `docs/investigations/bluetooth-proximity-oracle.md`.

**Where to look next:**

- For a wire interaction's spec → `protocol/*.yaml` (or
  `docs/session-protocol.md` if it's not yet in YAML).
- For pairing application API → `docs/pairing-lifecycle.md`.
- For the path-selection FSM → `docs/session-protocol.md` (the
  authoritative source).
- For decisions and target history → `docs/audit-log.md`,
  `bullseye.yaml`.
- For C ABI → `c/include/pigeon/pigeon.h`.
- For Go API → `api.go`.

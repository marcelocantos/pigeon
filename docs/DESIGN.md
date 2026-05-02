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
single-client and multi-client), and named streams + named datagram
channels (full). LAN-direct carriers, path optimisation, and the
candidate-pair FSM are designed (`docs/session-protocol.md`,
`docs/local-backchannel.md`) but not yet implemented across all SDKs.

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

Five layers, bottom-up. Each layer's contract is what the layer above
sees; the implementation of the layer is free to change without breaking
the contract.

```
  ┌─────────────────────────────────────────────────────────────┐
  │ 5. Application API     Listener / Session / Stream /        │
  │                         Datagram (per-language idiomatic)   │
  ├─────────────────────────────────────────────────────────────┤
  │ 4. Path selection      ICE-like FSM: gather candidates,     │
  │                         probe, nominate, swap, fall back    │
  ├─────────────────────────────────────────────────────────────┤
  │ 3. Session             Logical, AEAD-bound, named-stream    │
  │                         + named-datagram channel pair       │
  ├─────────────────────────────────────────────────────────────┤
  │ 2. Carrier             A single QUIC connection currently   │
  │                         bearing the session — relay-tunnel  │
  │                         or LAN-direct. Hot-swappable.       │
  ├─────────────────────────────────────────────────────────────┤
  │ 1. QUIC substrate      Real QUIC, host ↔ relay or peer.     │
  │                         ngtcp2 / quic-go / NWConnection.    │
  └─────────────────────────────────────────────────────────────┘
```

**Critical invariant.** A *session* is identified by its AEAD channel
(derived from the PairingRecord). The carrier underneath is *not* part of
the session's identity. The path-selection layer can swap carriers under
the session — the session, its named streams, its datagram channel
table, all carry over. Streams in flight require draining/replay logic at
the swap moment (see §8); the cryptographic and naming state survives a
swap by design.

This is what makes "we found a better path" tractable. Without this
invariant, every transport change would force re-pairing and lose
in-flight state.

**Two state machines, not five.** The five-layer diagram above shows
*responsibilities*, not state-machine count. There are exactly two
state machines in pigeon, both expressed (or to be expressed) in
protogen specs:

- **Pairing machine** — runs once per device pair; produces a
  PairingRecord; lives in `protocol/pairing.yaml` today.
- **Session machine** — runs once per logical connection; encompasses
  activation, path selection (the §4 layer), and health monitoring as
  sub-states of one FSM; lives in `docs/session-protocol.md` (designed)
  and is the largest pending protogen YAML.

Path selection (layer 4 in the diagram) is a *sub-FSM of the session
machine*, not a separate top-level state machine. Layering it
separately in the diagram reflects the responsibility split — path
selection operates on "which carrier"; session operates on "AEAD +
naming" — but in execution they share one executor and one event loop.

**Per-layer notes:**

1. **QUIC substrate** — owned by per-language QUIC libraries (`ngtcp2`
   in C, `quic-go` in Go, `NWConnection`/`Network.framework` in Swift,
   etc.). Provides streams, datagrams, connection migration,
   congestion control. **Pigeon does not reimplement any of this.**
   Native QUIC connection migration handles "my address changed" without
   pigeon involvement (see §8 on what migration does *not* solve).
2. **Carrier** — a single QUIC connection currently carrying the session.
   In the relay-tunnelled case, the carrier is `host ↔ relay` and the
   relay forwards opaquely to the peer. In the LAN-direct case, the
   carrier is `host ↔ peer` directly. Either way, the carrier exposes
   the same vtable to the layer above (open/accept/send/recv stream;
   send/recv datagram). The C side defines this vtable explicitly
   (`pigeon_transport` in `c/include/pigeon/pigeon.h`).
3. **Session** — AEAD-bound logical entity. Owns the channel keys, the
   clientTag (relay-side routing identifier), the named-stream demux
   table, and the named-datagram channel table. Built on top of a
   carrier; survives carrier swaps.
4. **Path selection** — the FSM that decides which carrier the session
   uses at any moment. Gathers candidates (relay always, LAN candidates
   from local interfaces, future: STUN-reflexive). Probes connectivity.
   Nominates a winning carrier. Swaps. Falls back to the relay
   immediately on failure. Periodically re-probes for a better path.
   The relay carrier is a *permanent baseline* — it stays up while the
   session is active, never dropped, even when a better carrier is in
   use. See `docs/session-protocol.md` for the full FSM (joined
   pairing + transport state machine, TLA+-verified).
5. **Application API** — the surface apps program against. Same
   semantics across languages; idiomatic per-language surface
   (single-threaded blocking in C, goroutines + channels in Go, async
   in Swift / Kotlin).

**Current state vs. target.** Layers 1, 2, 3, 5 are present in Go (full),
Swift (most), Kotlin (most), C (post-T22 wire helpers + sessions; listener
not yet shipped). Layer 4 (path selection) is fully designed in
`docs/session-protocol.md` but not implemented in any SDK. The
relay-permanent-baseline invariant is a binding design decision — do not
describe the relay as a "fallback" that goes away when LAN is up.

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
| **Session machine** (whole)  | `docs/session-protocol.md` (designed; not yet in YAML); verification by an independent `formal/SessionMachine.tla` once in YAML |
| &nbsp;&nbsp;⊢ session-open hello/ack | Sub-state — currently hand-rolled JSON (`api.go::connectHello/connectAck`) |
| &nbsp;&nbsp;⊢ path selection / LAN candidate exchange | Sub-state — designed; not implemented |
| &nbsp;&nbsp;⊢ health monitor ping/pong | Sub-state — designed; not implemented |
| Cross-machine properties     | Interface contract (`ValidPairingRecord`) + local proofs in each spec; **no composed TLA+ spec** (state space explodes) |
| Relay greeting variants      | One-shot byte format — hand-rolled (`pigeon.go`, `transport.go`, `c/src/ngtcp2_transport.c`) |
| Stream-name binding header   | One-shot byte format — hand-rolled (`api.go::encodeStreamHeader`, `c/src/pigeon.c::pigeon_encode_stream_header`) |
| Datagram channel-id framing  | One-shot byte format — hand-rolled (encoder helpers in same files) |

**Current state vs. target.** The pairing ceremony is the only wire
interaction currently flowing through protogen. Everything else is
hand-rolled in each language and kept in sync by careful review and
shared byte-vector tests (`wire_vectors_test.go` ↔ `c/test/test_pigeon.c`).
The byte-vector lockstep does catch drift, but the *class* of bug is
still possible. The target is for every entry in the catalogue above to
be a protogen spec.

The session+transport FSM (joined pairing + path-selection state
machine) is the largest piece of pending YAML work. It is fully
specified prosaically in `docs/session-protocol.md` — translating that
into protogen YAML and re-generating the executors is the path to
collapsing significant chunks of hand-rolled per-language session
plumbing.

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
- **Demux tables and queues.** Mapping clientTag → session, mapping
  stream name → pending accept-stream waiter, datagram channel-id →
  per-channel queue. Pure data structures; no protocol states.
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

## 8. Path optimisation

**Path selection is a sub-FSM of the session machine** (see §3 and §4
on the two-state-machine structure), not a third top-level machine.
This section describes its responsibilities and shape; the
authoritative spec for the integrated session machine lives in
`docs/session-protocol.md`.

**The problem.** QUIC's connection migration handles "my address
changed; keep the connection going" (RFC 9000 §9). It does *not*
handle "a better route exists between the same two peers." If client
and backend both end up on the same LAN, QUIC alone has no concept of
"bypass the relay" — the topology stays `client ↔ relay ↔ backend`.
Optimising for the available path is pigeon's job.

**The shape of the FSM** (ICE-like, simpler):

- **Candidate gathering.** Each peer enumerates reachable addresses:
  the relay (always available; permanent baseline), LAN host candidates
  from local interfaces, future STUN-reflexive candidates (see
  `docs/investigations/stun-hole-punching.md`).
- **Candidate exchange.** Peers swap candidate lists over the existing
  relay-mediated session, on a dedicated control stream.
- **Connectivity checks.** Pair the candidates, probe each pair (a
  challenge-response handshake on the candidate carrier).
- **Nomination.** Pick the best working pair; the carrier swap moves
  the session over to the new carrier.
- **Fallback.** If the optimised carrier degrades (3 ping failures →
  degraded → backoff schedule), drop back to the relay carrier
  immediately. The relay stays up the entire time; fallback is
  zero-cost.
- **Periodic re-probe.** Try to re-optimise on a backoff-bounded
  schedule when conditions change.

**Carrier-swap semantics.** Open question, not yet decided. Three
candidate strategies:

- **(a) Drain-and-resume** — quiesce open streams, swap, reopen.
  Simplest; introduces latency at swap.
- **(b) Replay window** — track unacked messages per stream, replay on
  the new carrier post-swap. Cleaner UX; more bookkeeping.
- **(c) Application-visible** — surface the swap to the app, let it
  decide.

(a) is the recommended v1 default; (b) is a credible v2; (c) is a
foot-gun unless absolutely required by an application class. The
decision point is deferred until path optimisation is implemented.

**ICE-related future work.** STUN-reflexive candidate gathering for
NAT traversal is a credible third tier between LAN-direct and relay.
~70-80% Wi-Fi coverage with simple STUN; mobile cellular with symmetric
NAT requires full ICE+TURN. See `docs/investigations/stun-hole-punching.md`
for the analysis.

**Current state vs. target.** Path optimisation is fully designed
(`docs/session-protocol.md`) and not yet implemented in any SDK. The
relay-tunnelled carrier is the only carrier shipping today; LAN-direct
is the next major capability. STUN-reflexive is a third tier deferred
behind LAN-direct.

---

## 9. Open questions and current-vs-target gaps

**Largest open architectural gaps:**

1. **Wire interactions not yet protogen specs.** Stream-name binding
   header, datagram channel-id framing, relay greeting variants,
   session-open hello/ack, the joined session+transport FSM. Each is a
   class of drift bugs eliminated by moving it into protogen.
2. **Path optimisation unimplemented.** The carrier abstraction (§3)
   and the FSM (§8) are designed; LAN-direct, candidate exchange,
   probing, swap mechanics, and fallback all need first implementations.
3. **Carrier-swap semantics undecided.** §8 lists the three candidate
   strategies; v1 needs to commit to one before path optimisation
   ships.
4. **Cross-language API parity.** Swift / Kotlin / C are all behind Go
   on at least one application-API edge (Swift's `setChannel`
   asymmetry is the documented case; others may exist).
5. **Relay authentication mechanism.** Currently a bearer token in the
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

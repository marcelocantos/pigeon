# Multi-service nodes: sub-addressing + discovery

> **Status:** Implemented (Go). Per-session keys (🎯T44.1), sub-addressing
> (🎯T44.2), and discovery (🎯T44.3) are done and exercised end-to-end
> (`e2e_test.go` `TestE2EMultiServiceNode`). Cross-SDK backend/session parity
> (Swift/Kotlin/TS) and cascading relay-level routing are follow-on — see
> §12. Tracked by 🎯T44 (umbrella) and its sub-targets.
> **Supersedes** the route-substream multiplexing sketch in
> [`cascading-relay.md`](cascading-relay.md) §4a and resolves its §5/§9
> open questions.
> **Motivation:** ge/ged — one laptop (`ged`) fronts several game servers;
> a player pairs *once* with the laptop, discovers the running games, and
> joins them, rather than pairing per game.

---

## 1. The problem: a missing middle level

The scenario wants three nesting levels:

```
pair once             → identity            (player ↔ ged: one PairingRecord)
per-service connection → ???  ← the gap
channels within a service → named streams + datagram channels
```

Pigeon already had the **top** (a `PairingRecord` is one paired identity)
and the **bottom** (a `Session` carries named streams and datagram
channels). What was missing is the **middle**: many concurrent
*connections* under one pairing, each with its own channels. Channels are
the wrong level for "a game" — a game is a connection that itself wants
channels inside it.

## 2. Decision: a service is a `Session`, not a multiplex over channels

Do **not** multiplex services over one session's channels (route-prefixed
substreams + a datagram routing prefix). Instead:

- **service → `Session`**, and **channel → QUIC stream**.

QUIC already gives exactly two levels of multiplexing — connection and
stream. Mapping a service to a connection and a channel to a stream yields
the second nesting level *for free*, and "connection per service, channels
within" maps 1:1 onto "Session per service, streams within." No third
abstraction; the existing L1/L2/L4 stack is reused wholesale.

Why this beats muxing over channels:

- **Real isolation.** Separate QUIC connections give independent
  head-of-line behaviour, independent lifecycle (leave one service while
  others run), and independent channel namespaces (`gs1`'s `state` stream
  is not `gs2`'s).
- **QUIC-aligned.** It does not re-implement a connection layer as a naming
  convention inside one connection.
- **`N=1` falls out.** A single-service node is just one route.

Consequence: **the node is a *router* (route → service), not a stream
multiplexer.** QUIC does the multiplexing; the node only decides where each
incoming session goes.

## 3. Sub-addressing

`Connect` targets a sub-address — `(instance, route)` — not a bare instance
ID:

- The connect greeting carries an optional `route`; empty route is the
  default and is the `N=1` case.
- An accepted backend `Session` exposes the route it was reached on
  (`Session.Route()`); `ConnectRequest` carries the route so `VerifyConnect`
  is route-aware.
- A client may `Connect` with the **same `PairingRecord`/instance multiple
  times**, each with a different route → independent concurrent `Session`s.

Multi-client-per-backend is already the T45 remote-Listen model (the node
parks enough `listen` slots), so no relay-level constraint changes.

## 4. Per-session keys (correctness prerequisite — 🎯T44.1)

Today `cwire.DeriveSessionChannel(record, isBackend)` derives AEAD keys
purely from `(PairingRecord, direction)`. With AES-256-GCM monotonic-counter
nonces that restart at 0, **two concurrent sessions from one record reuse
keys *and* nonces** — catastrophic. The multi-session model is impossible
until this is fixed.

The fix is small because the raw material already exists: the activation
handshake (`protocol/session.yaml`) exchanges a fresh `auth` nonce, used
today only for replay protection. Fold that nonce (ideally both peers
contributing) into the HKDF `info`/salt so every activation yields distinct
keys. This is independently good hygiene — reconnect sessions stop reusing
keys too.

## 5. Discovery: mechanism in the protocol, policy in a hook

A route is useless if you cannot learn it, so discovery is a first-class
companion to sub-addressing — but split **mechanism from policy**, exactly
as relay admission already does with `pigeon.Auth`:

- **Mechanism (protocol).** A stateless `enumerate` request / `routes`
  response in `protocol/wireformats.yaml`. Each response entry is
  `{route-id, opaque metadata bytes}` — the metadata is opaque to pigeon
  (app-defined), like any payload. Spoken on a **reserved control route**
  (empty route). **Snapshot, not pubsub** (poll/refresh; live subscription
  is the §1 presence non-goal and is deferred).
- **Policy (backend hook).** `RegisterArgs.Discover(peer[, filter]) →
  []RouteEntry`, sibling to `pigeon.Auth`. The default returns all routes;
  business logic returns a per-peer subset and fills the metadata. So
  "enumerate everything" and "business-logic-controlled discovery" are the
  *same wire* — only the hook differs.

**Visibility and connectability are independent** decisions over the one
route namespace:

- **Visibility** — which routes a peer can *see* → `Discover`.
- **Connectability** — which routes a peer can *join* → `VerifyConnect`
  (route-aware).

A service can be visible-but-full or joinable-but-hidden (direct-invite by
route id). Conflating them into one gate would foreclose both UXes.

## 6. Auth boundary

In this topology the node is the client's **E2E endpoint**, not a blind
relay for service traffic: it terminates the player's session and fronts
local services. That is precisely what "pair once = trust the laptop"
means — you trust `ged`, and everything it runs. Downstream services never
see player credentials. (If a service were a distinct trust domain that
must be opaque even to the node, you would pair per service — which this
design explicitly avoids.)

## 7. Three tiers of implementation

| Tier | Owns | Examples |
|------|------|----------|
| **Protocol** (protogen, all SDKs) | wire + key derivation | route on the connect greeting; `enumerate`/`routes` wireformat; per-session key derivation in the activation handshake |
| **`pigeon` Go library** (imported by the node) | the API surface | `Session.Route()`; `RegisterArgs.Discover`; route on `ConnectRequest` (route-aware `VerifyConnect`); multiple `Connect` per `PairingRecord` |
| **Node application** (e.g. `ged`) | policy + dispatch | the `Discover` implementation (its service list + visibility) and the route → service dispatch |

The node imports the (enriched) `pigeon` server library — there is **no
separate "multiplexer" library or binary.** The accept loop collapses to
routing:

```go
listener, id, _ := pigeon.Register(ctx, &pigeon.RegisterArgs{
    Identity: ged,
    Discover: ged.listServices, // policy hook
    Auth:     ged.authPlayer,   // route-aware VerifyConnect
    // ...
})
for {
    sess, _ := listener.Accept(ctx)
    go ged.route(sess.Route(), sess) // dispatch to the right service
}
```

The node writes no wire formats, key separation, or stream demux — only
*which services exist, who may see/join them, and where each route goes*.

## 8. Deployment shapes (same primitives)

- **In-process dispatch.** Services are goroutines in the node. `ged.route`
  hands the session to a local handler. No second pigeon hop. Simplest;
  this is the path the 🎯T44 end-to-end test demonstrates.
- **Cascading relay.** Services are separate backends that register with the
  node; the node also runs `pigeon.NewQUICServer`/`NewWebTransportServer`
  and registers upstream as a backend. `ged.route` bridges the player
  session to the local service backend. This is the original
  `cascading-relay.md` framing, retained as an additive extension.

Tier-1/tier-2 additions are identical for both; only `route`'s use differs.
Start in-process; grow into cascading without protocol changes.

## 9. Keep the node's route registry single-sourced

`Discover` (enumeration) and `route` (dispatch) read the same set of live
services. Two parallel structures drift — a service shows up in discovery
but routing 404s, or vice versa. So the node should keep **one** route
registry (register a route once with metadata + handler; serve both
`Discover` and dispatch from it).

Whether pigeon ships that as a thin helper or the node keeps a 20-line map
is a build-time call. Default: the node owns it first; promote to the
library only if a second adopter (or a Swift/Kotlin node-equivalent) needs
the same thing. Do **not** add a generic `Router` type speculatively.

## 10. What pigeon still does not do

- **Global presence / pubsub / fanout** stays a non-goal (DESIGN.md §1).
  Per-backend route *enumeration within an authenticated session* is the
  bounded, in-scope thing this design adds — distinct from advertising who
  is online network-wide.
- **Metadata schema.** Route metadata is opaque bytes; the app defines its
  shape (it may itself be a protogen wireformat the app owns). Pigeon never
  learns what "player count" is.
- **Multi-party sessions.** A `Session` still has exactly two peers; a node
  fronting N services is N two-peer sessions, not a group.

## 11. Target map

| Target | Scope |
|--------|-------|
| 🎯T44 | umbrella — a paired node fronts multiple services over one pairing |
| 🎯T44.1 | per-session AEAD key derivation (correctness blocker) |
| 🎯T44.2 | sub-addressed `Connect` / `route` / `Session.Route()` / multi-session-per-pairing |
| 🎯T44.3 | `enumerate` wireformat + `Discover` hook; visibility vs connectability |

Dependency order: T44.1 → T44.2 → T44.3 → T44.

## 12. Implementation status and deviations from this design

Shipped (Go), verified by `make bullseye` (incl. C↔Go interop) and the
`e2e_test.go` oracles `TestE2ERoutedSessions`, `TestE2EDiscovery`, and
`TestE2EMultiServiceNode`:

- **Per-session keys (T44.1)** — a fresh 16-byte client nonce rides
  `auth_request` and is folded into the HKDF `info` on both sides
  (`c/src/activation.c`, `cwire`, C SDK). Landed across the activation
  ecosystem (C + Go + Swift `PigeonSession`, which routes through the C path).
- **Sub-addressing (T44.2)** — `ConnectArgs.Route` / `Session.Route()`.

Two deviations from the sketch above, made during implementation:

- **The route rides the end-to-end `auth_request`, not the relay greeting
  or a `wireformats.yaml` entry** (§3 left the wire unspecified; this is the
  concrete choice). The relay never sees the route — better privacy, and it
  serves the in-process-dispatch model directly. Cascading relay-level
  routing (where a local relay *does* need the route in a greeting to pick a
  downstream game-server backend) is the remaining piece of the cascading
  deployment, not yet built.
- **The `enumerate` codec is Go-side today, not `protocol/wireformats.yaml`**
  (§5's mechanism-in-protocol goal). `RegisterArgs.Discover` + the reserved
  discovery stream + the length-prefixed `{route, opaque-metadata}` list live
  in `discovery.go`. Promoting the codec into `wireformats.yaml` so
  Swift/Kotlin/TS get a consistent client `Enumerate` is follow-on, tracked
  with the broader Swift/Kotlin/TS backend/session parity gap (DESIGN.md §9).

Follow-on (not blocking the Go ge/ged use case): Swift/Kotlin/TS
backend-side `Discover`/`Route()` parity; cascading relay-level route
dispatch; optional live discovery subscription (still a §1 non-goal today).

# 🎯T43 Phase 1 STUN — blocker

**Status:** Blocked on 🎯T39's path-selection sub-FSM not being in the shape T43 needs.

**Date:** 2026-05-21

---

## Summary

🎯T43 (STUN-reflexive candidate gathering as a third path-tier) cannot
start coding. Its principal acceptance criteria assume that the path
selection sub-FSM from 🎯T39 already models:

- a typed set of **candidates** (LAN-host, server-reflexive, relay),
- a peer-to-peer **candidate-exchange** control message over the relay
  session,
- a generic **candidate-pair handshake** with nomination and fallback
  that can be reused for any candidate type.

None of these exist in the current `protocol/session.yaml`.

## What was actually built under 🎯T39

`protocol/session.yaml` (and its now-superseded sibling
`protocol/pathswitch.yaml`) implements path selection as a flat
**LAN-vs-relay two-state model**:

- States: `RelayConnected`, `LANOffered`, `LANActive`, `LANDegraded`,
  `RelayBackoff` (backend); `RelayConnected`, `LANConnecting`,
  `LANVerifying`, `LANActive`, `RelayFallback` (client).
- Messages: a single hard-coded LAN offer/verify/confirm exchange
  (`lan_offer` / `lan_verify` / `lan_confirm`) — not a generic
  candidate exchange.
- Variables: `lan_addr`, `challenge_bytes`, `offer_challenge`,
  `active_path` (string `"relay"` or `"lan"`).
- No candidate type tag. No candidate list. No general candidate-pair
  handshake. No nomination logic over a set.

Greps for "candidate" across `protocol/session.yaml`,
`protocol/pathswitch.yaml`, `docs/session-protocol.md`, and `session.go`
all return zero matches.

## Cross-check with 🎯T39 acceptance and DESIGN.md

🎯T39's acceptance text says:

> protocol/session.yaml exists and describes the full session state
> machine: activation (replacing the current connectHello/connectAck
> JSON), **path selection sub-FSM (candidate gathering, exchange,
> connectivity checks, nomination, swap, fallback)**, health
> monitoring …

`docs/DESIGN.md` confirms the gap explicitly:

- §3 lines 50–52: "candidate-pair FSM are designed
  (`docs/session-protocol.md`, `docs/local-backchannel.md`) but not
  yet implemented across all SDKs."
- §4 line 373: "⊢ path selection / LAN candidate exchange |
  Sub-state — designed; not implemented."
- §9 lines 608–610: "LAN-direct, candidate exchange, probing, swap
  mechanics, and demotion all need first implementations on the
  cleanly-layered surface."

In other words: 🎯T39 was marked achieved on the basis of its
activation + LAN-vs-relay path subset, but the **candidate-pair sub-FSM
that 🎯T43 depends on did not actually ship**.

## What 🎯T43 needs from a re-opened 🎯T39 (or a new 🎯T39.x) before it can start

1. A first-class `Candidate` record in `protocol/session.yaml` with:
   - kind (`host`, `srflx`, `relay`) — surfaced as a wire constant,
   - transport address (or opaque blob — TBD),
   - a candidate ID stable across the session,
   - optional priority (ICE-shaped — 32-bit unsigned, or simpler).
2. A `candidates_offer` / `candidates_answer` (or single `candidates`)
   message between peers carried over the relay AEAD pipe. This is the
   wire format 🎯T43 must extend with srflx entries; speculating it
   now would lock in the wrong shape.
3. A generic candidate-pair handshake — a `pair_check` /
   `pair_check_ack` exchange (or equivalent) authenticated by the L2
   AEAD context, sent on the candidate-under-test pipe. Acceptance
   criterion three of 🎯T43 says explicitly: *"Connectivity checks
   against STUN-reflexive candidate pairs use the same candidate-pair
   handshake as LAN-direct."* That handshake does not exist yet for
   LAN either — the current LAN path uses a one-shot
   challenge/verify/confirm tied to fields named `lan_*`.
4. A nomination predicate over the candidate-pair set, with a single
   active pair surfaced as an `active_pair` variable instead of the
   current binary `active_path \in {"relay", "lan"}`.
5. A TLA+ verification of the candidate model under 🎯T39's
   verification harness — invariants for "exactly one nominated pair
   when L4 traffic is flowing on a non-relay candidate", "relay always
   reachable", etc.

## Recommendation

- Re-open 🎯T39 (or split a 🎯T39.x) to deliver the candidate-pair
  sub-FSM the original acceptance promised. Refactor the current
  `lan_offer` / `lan_verify` / `lan_confirm` family into the generic
  candidate-pair handshake so that LAN-host becomes the first instance
  of the generic mechanism, not a one-off.
- 🎯T43 stays parked until the above lands. When 🎯T43 resumes, the
  work is purely additive: a new `stun/` Go package, a `srflx`
  candidate kind, candidate-gathering on session open, and a
  reflexive-pair end-to-end test. None of that work depends on Phase 2
  ICE / TURN (which remains a separate future target).

## What was *not* done in this fan-out branch

- No new `stun/` package was created.
- `protocol/session.yaml` was not modified (would have required
  inventing the candidate wire format unilaterally — explicitly
  prohibited by the 🎯T43 brief).
- No code changes beyond this blocker note.

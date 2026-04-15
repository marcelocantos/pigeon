# STUN / NAT Hole-Punching as a Transport

**Status:** Investigation  
**Target:** 🎯T4  
**Date:** 2026-04-15

---

## Motivation

Pigeon currently offers two transport tiers:

| Tier | Latency | Availability | Relay cost |
|------|---------|--------------|------------|
| Relay (webtransport/QUIC) | Medium–high (relay RTT) | Always | Bandwidth × 2 |
| LAN direct (future) | Very low | Same-network only | None |

A third tier — **peer-to-peer across the internet via NAT hole-punching** — would eliminate relay costs and reduce latency for devices on different networks that cannot use a relay for privacy or cost reasons. This investigation evaluates whether STUN/ICE-based hole-punching is viable as that middle tier.

---

## Background: How NAT Hole-Punching Works

When two peers are both behind NAT, neither has a publicly routable address. The technique:

1. Each peer connects to a **STUN server** (a publicly-accessible server that reflects the peer's external IP:port back to it — RFC 5389/8489).
2. Both peers learn their external address/port and exchange them via a **signalling channel** (in pigeon's case, the relay is a natural signalling channel).
3. Each peer sends a UDP packet to the other's external address. The NAT on each side creates a mapping when the packet leaves, and that same mapping accepts the incoming packet from the other peer — the "hole" is punched.

This is codified in **ICE (Interactive Connectivity Establishment, RFC 8445)**, which adds candidate gathering, connectivity checks, and fallback logic on top of bare STUN.

---

## STUN Server Requirements

### What a STUN server does

A STUN server answers `Binding Request` messages with the client's reflexive transport address (external IP:port as seen by the internet). For hole-punching rendezvous, a STUN server only needs to handle these binding requests — it does not relay data. This is a stateless, low-bandwidth service.

Bandwidth per rendezvous: ~160 bytes request + ~160 bytes response = ~320 bytes per peer × 2 peers = ~640 bytes per hole-punch attempt. STUN bandwidth is negligible.

### Public STUN servers

Several public STUN servers are available at no cost:

- `stun.l.google.com:19302` and `stun1.l.google.com:19302` (Google, widely used, no SLA)
- `stun.cloudflare.com:3478` (Cloudflare, announced 2023, more reliable)
- `stun.stunprotocol.org:3478` (community)
- IETF: `stun.ietf.org:3478`

For production use, relying on Google's STUN servers carries a policy risk (no SLA, could be deprecated). Cloudflare's offering is more appropriate for production.

**References:**
- RFC 8489 (STUN): https://www.rfc-editor.org/rfc/rfc8489
- Cloudflare STUN/TURN: https://developers.cloudflare.com/calls/turn/

### Self-hosted: coturn

`coturn` (https://github.com/coturn/coturn) is the canonical open-source STUN+TURN server.

Cost factors for self-hosting:
- **Compute:** Minimal for STUN-only (no data relay). A $5/month VM handles tens of thousands of rendezvous/second.
- **TURN relay bandwidth:** If STUN hole-punching fails and a TURN fallback is needed, the TURN server relays all traffic — bandwidth cost scales with use. At $0.09/GB (typical cloud), a moderate 1 TB/month = ~$90.
- **Operational:** coturn requires TLS certificates, firewall rules (UDP 3478, TCP 443/5349), and monitoring. Fly.io (pigeon's existing host) supports coturn but adds operational overhead.

**Verdict for pigeon:** Use public STUN (Cloudflare) initially. Only self-host if TURN fallback is needed, which requires a separate cost/benefit analysis.

---

## UDP vs TCP Hole-Punching

### UDP hole-punching

UDP is the standard mechanism. Most NATs maintain port mapping state for UDP: once a packet leaves port X → external-port Y, incoming packets to external-port Y are forwarded to port X for a timeout period (typically 30–300 seconds).

**Success rate:** ~80–90% across all NAT types in practice (Bittorrent studies, 2004 Bryan Ford paper). Failures are primarily on symmetric NATs.

**Complexity:** Low. A STUN binding request + coordinated simultaneous open is sufficient.

**References:**
- Ford, Srisuresh, Kegel (2004): https://pdos.csail.mit.edu/papers/p2pnat.pdf
- RFC 5128 (P2P across NATs): https://www.rfc-editor.org/rfc/rfc5128

### TCP hole-punching

TCP hole-punching is possible but significantly harder:

- Requires `SO_REUSEADDR`/`SO_REUSEPORT` to reuse the local port for both listen and connect.
- The SYN packets must cross in flight (simultaneous open), which is timing-sensitive.
- Many NATs and firewalls drop incoming TCP SYNs to unprompted destinations.
- **Success rate:** ~40–60% (significantly lower than UDP).
- Platform support varies: macOS and Linux handle simultaneous TCP open; Windows is more problematic.

**Verdict:** UDP hole-punching is strongly preferred. Pigeon already uses QUIC (which runs over UDP), so UDP-only hole-punching aligns perfectly. TCP hole-punching offers no advantage here.

---

## NAT Type Analysis

### Cone NATs (full cone, restricted, port-restricted)

All cone NAT variants preserve the external port across destinations. Once a mapping is established, any peer that knows the external address can reach the internal host (with varying degrees of restriction on which external address is allowed).

- **Full cone:** Any external host can reach the mapped port. Hole-punching trivially succeeds.
- **Restricted cone:** External host must have received a packet from the internal host first. Standard simultaneous-open hole-punching succeeds.
- **Port-restricted cone:** External host+port must have received a packet. Standard simultaneous-open succeeds.

Roughly 70–80% of residential NATs (home routers) are cone NATs.

### Symmetric NATs

Symmetric NATs assign a **different external port for each destination**. The external address seen by the STUN server differs from the external address assigned for traffic to the peer. Hole-punching fails unless the peer can predict the port, which requires port prediction techniques (fragile, unreliable).

Symmetric NATs are common on:
- **Corporate/enterprise networks** (forced symmetric for security policy)
- **Mobile carriers (4G/5G):** Most mobile carrier NATs are **symmetric by design**. iOS and Android devices on cellular are very likely to be behind symmetric NAT. This is a critical pigeon use case.
- **CGNAT (Carrier-Grade NAT):** RFC 6888 NAT, used by ISPs to share IPv4 addresses, is often symmetric.

### Mobile carrier NAT behaviour

Testing by multiple researchers and the libdatachannel project consistently shows:
- iOS on cellular: ~60–80% behind symmetric or CGNAT
- Android on cellular: similar
- Both platforms on Wi-Fi: ~70–80% behind cone NATs

**Reference:** libdatachannel NAT traversal notes: https://github.com/paullouisageneau/libdatachannel

**Consequence:** Hole-punching without TURN fallback will fail for a significant fraction of pigeon's target use case (mobile ↔ desktop across the internet). A pure STUN approach is insufficient; ICE with TURN fallback is needed for reliable mobile support.

---

## ICE vs Simple STUN

### Simple STUN (no ICE)

Gather one STUN reflexive candidate per peer, exchange via signalling, attempt simultaneous UDP open. No TURN, no candidate prioritisation, no connectivity checks.

**Pros:** Simple to implement (~200 lines of Go). Works for ~70–80% of peer pairs.  
**Cons:** No fallback for symmetric NATs; no TURN integration; reinvents connectivity checking that ICE already solves.

### Full ICE (RFC 8445)

ICE gathers multiple candidate types:
- **Host candidates:** local IP:port
- **Server-reflexive (srflx):** STUN-derived external address
- **Relayed (relay):** TURN-allocated external address (fallback)

ICE performs connectivity checks between all candidate pairs and selects the best working path. It handles symmetric NATs via TURN relay fallback.

**Pros:** Handles all NAT types; well-specified; libraries exist for Go and iOS/Android.  
**Cons:** Complex protocol; TURN server needed for full coverage; ICE negotiation adds 200–500ms latency to connection setup.

### Assessment for pigeon

Given that pigeon's primary use case involves mobile devices (often behind symmetric NAT or CGNAT), simple STUN alone is insufficient. ICE with TURN provides reliable fallback, but the operational cost of running a TURN server is non-trivial.

A pragmatic middle path: implement ICE with TURN, but configure TURN only when a self-hosted server is available. Fall back to the existing pigeon relay (which already handles all cases) when TURN is not configured. This degrades gracefully.

---

## Library Options

### Go

**pion/stun** (https://github.com/pion/stun)  
- RFC 8489-compliant STUN implementation in pure Go.
- Actively maintained (part of the pion project).
- Handles binding requests, STUN attributes, message integrity.
- Low-level; does not include hole-punching logic directly.

**pion/ice** (https://github.com/pion/ice)  
- Full ICE agent in pure Go.
- Supports host, srflx, and relay (TURN) candidates.
- Used by pion/webrtc for WebRTC P2P.
- Well-tested, actively maintained.
- Integrates naturally with pigeon's existing QUIC/UDP stack.
- License: MIT.

**pion/turn** (https://github.com/pion/turn)  
- TURN client and server in Go.
- Can be embedded for self-hosted TURN if needed.

pion is the clear Go choice. It is well-maintained, modular, and already used in production-scale WebRTC deployments.

### iOS / Swift

**Apple Network framework** does not expose raw ICE/STUN APIs.

**WebRTC.framework (Google's prebuilt)**  
- Google distributes a prebuilt `WebRTC.framework` for iOS that includes ICE/STUN/TURN.
- Heavy (~50MB binary), designed for video conferencing, overkill for a signalling+transport primitive.
- Available via Swift Package Manager: https://github.com/stasel/WebRTC

**libdatachannel** (https://github.com/paullouisageneau/libdatachannel)  
- C++ library (with Swift bindings via a thin C wrapper) implementing WebRTC data channels, ICE, STUN, TURN.
- Much lighter than full WebRTC.framework (~2MB).
- Swift wrapper: https://github.com/bprotestle/swift-datachannel (community, less maintained).
- A better option than full WebRTC if a C++ dependency is acceptable.

**Network.framework + custom STUN**  
- For simple STUN (no TURN/ICE), a ~300-line Swift implementation using `Network.framework` UDP sockets is feasible.
- Insufficient for symmetric NAT fallback but adequate for a first pass.

**Verdict for iOS:** The lack of a first-party ICE library is a real gap. libdatachannel with a Swift wrapper is the most viable option. Full WebRTC.framework is acceptable if the binary size cost is tolerable.

### Android / Kotlin

**Android WebRTC** (https://webrtc.googlesource.com/src/)  
- Google publishes `libwebrtc` as a `.aar` on JCenter/Maven Central via `io.getstream:stream-webrtc-android`.
- Includes full ICE/STUN/TURN support.
- ~10MB AAR. Overkill but functional.

**pion/ice via gomobile**  
- pion/ice can be compiled to an Android library via `gomobile bind`. This would share the Go implementation across platforms.
- Feasible but adds gomobile toolchain complexity.

**libdatachannel**  
- JNI bindings exist (https://github.com/elsaland/android-datachannel), less maintained.

**Verdict for Android:** Android WebRTC AAR is the most practical path. gomobile + pion/ice is appealing for code sharing but adds build complexity.

---

## Connection Setup Latency

Approximate timeline for ICE-based hole-punching:

| Step | Time |
|------|------|
| STUN binding request (each peer) | ~50–150ms (RTT to STUN server) |
| Exchange via pigeon relay (signalling) | ~100–300ms (relay RTT) |
| ICE connectivity checks | ~100–200ms |
| **Total** | **~250–650ms** |

Compare to relay: established from the moment the relay connection exists (pigeon relay is always running). Hole-punching has higher setup cost but lower per-packet latency and zero relay bandwidth once established.

---

## Integration with Pigeon

The pigeon relay already exists as a signalling channel — the two peers are already connected to the relay when hole-punching would be initiated. The integration sketch:

1. After relay connection is established, both peers gather STUN candidates.
2. Candidates are exchanged over the existing relay connection (using a new datagram type or a control stream).
3. ICE connectivity checks run; if a direct path is found, the session migrates to the direct QUIC connection.
4. If ICE fails (symmetric NAT, both sides), the relay connection continues as fallback.

QUIC connection migration (RFC 9000 §9) allows an existing QUIC session to migrate to a new path without disruption, which would make the relay→direct transition transparent to the application layer.

---

## Recommendation

**Pursue, but in two phases.**

**Phase 1 — Simple STUN (MVP):** Implement STUN-only hole-punching using pion/stun in Go and Network.framework UDP on iOS. No ICE, no TURN. This handles ~70–80% of Wi-Fi-to-Wi-Fi use cases and is relatively cheap to build (~1–2 weeks). The implementation is entirely additive — if hole-punching fails, the relay continues to work.

**Phase 2 — Full ICE with TURN:** Add pion/ice and a coturn deployment (on Fly.io). This captures the mobile cellular use case. Significantly more complex, and adds operational costs for TURN. Only worth doing if Phase 1 traffic analysis shows significant relay usage on cellular paths.

**Cost:** Go implementation is moderate effort. iOS/Android adds effort due to library gaps — libdatachannel or WebRTC SDK dependencies are non-trivial to vendor. The biggest operational risk is TURN server cost at scale.

**Value:** High for desktop-to-desktop or Wi-Fi-to-Wi-Fi scenarios. Moderate for mobile, which requires TURN to reliably work. If pigeon's primary use case is mobile ↔ desktop on cellular, hole-punching alone is insufficient; TURN is needed for full reliability.

---

## Open Questions

1. **What fraction of pigeon sessions are currently relay-limited?** Relay traffic metrics would clarify whether hole-punching is worth the investment.
2. **Is TURN self-hosting on Fly.io feasible?** Fly.io supports UDP, but coturn's requirements (dynamic port allocation, TLS) need validation.
3. **QUIC connection migration:** Does the current pigeon QUIC stack (quic-go) support path migration sufficiently to allow relay→direct transitions without session teardown?
4. **iOS Network.framework limitations:** Apple's policy restricts raw socket access in App Store apps. Is `NWConnection` with UDP sufficient for simultaneous-open hole-punching, or does it abstract away the timing required?
5. **IPv6 impact:** With IPv6, NAT traversal is often unnecessary (global addresses). What fraction of pigeon deployments are IPv6-only or dual-stack?
6. **gomobile for cross-platform ICE:** Is the added build complexity of gomobile worth the code-sharing benefit over vendoring separate ICE libraries per platform?

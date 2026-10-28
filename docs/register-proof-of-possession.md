# Register proof of possession

> **Status:** Specified, not implemented. No relay or SDK code in this
> change. Tracked by 🎯T65.
>
> **Problem.** A backend's stable instance ID is
> `base64url(sha256(xPub)[:16])` (`crypto/identity.go`,
> `fileIdentity.InstanceID`). Activation-mode `pigeon.Register` requests
> that ID (`api.go`). `WebTransportServer.handleRegister` and
> `QUICServer.handleRegister` accept any requested ID on a first-come
> basis. 🎯T56 stopped last-writer-wins against a *live* connection and
> made unregister ownership-checked; its own note says a dead slot may
> still be taken over. The ID is not a secret: clients put it in
> `connect:<id>`, pairing records and QR payloads carry it, and the
> relay logs it (`"instance registered"`). Anyone who learns an ID can
> register it while the real backend is offline. The real backend then
> gets "instance ID in use" and is locked out until the squatter drops
> or the relay restarts (the hub is in-memory). That is a denial of
> service. Payloads stay end-to-end encrypted, so confidentiality is
> unchanged. `PIGEON_TOKEN` / `Auth.VerifyRegister` only decides who may
> register at all. On a shared relay, or an open one, it does not bind
> an ID to a key. This bites backends that register only for the few
> minutes of a run (finance-bank-fetch is the shape).

## Decision

Proof of possession at register time for every **requested**
(non-empty) ID, using the X25519 key the ID is already derived from.
The relay checks `requestedID == InstanceID(xPub)` before the ID can
occupy a slot. A proof that verifies **evicts** whoever is sitting in
that slot. Relay-assigned IDs (empty request, pairing mode) stay
ungated.

### Why ECDH, not the Ed25519 signing key

Both options were weighed. The proof is static-ephemeral X25519, not
`Identity.Sign`.

The ID commits only to the X25519 public key. `fileIdentity` generates
the X25519 key and the Ed25519 key with two independent
`GenerateKey` calls, and `Identity` does not expose the Ed25519 public
key. An Ed25519 signature proves possession of a key the ID does not
commit to.

Binding that signing key into the ID (`H(xPub || edPub)`, or replacing
the derivation) would change every existing instance ID and invalidate
pairing records, artifacts, and QR payloads. An ECDH proof does not.

X25519 cannot sign. An existing X25519 private key cannot be turned
into an Ed25519 key. libsodium converts Ed25519 *to* X25519, and pigeon
did not generate identities that way.

A binding of the form "some Ed25519 key signs xPub" is forgeable by
anyone who knows xPub. xPub is already in pairing records and in the
QR token's `AcceptorIdPub`.

`Identity.DeriveSharedSecret` already performs X25519 plus
HKDF-SHA256 with a nil salt (`crypto/crypto.go` `hkdfDerive`). File
identities, the macOS keychain identity, and any later enclave
identity can produce the tag without exporting private bytes. The tag
is a confirmation, not a signature the relay can replay elsewhere.

`Identity.Sign` stays the peer-authentication operation it already is.
This target does not add a second identity key and does not change
`InstanceID`.

## Handshake

ASCII greetings stay protogen union variants in
`protocol/wireformats.yaml` (DESIGN.md §4). The challenge and the proof
are the next length-prefixed messages on the register primary stream
(the same 4-byte big-endian framing as today). They are not colon
fields: the proof is binary, and the colon grammar splits on `:`.

One helper serves both `webtransport.go` and `quicserver.go`.

### Relay-assigned ID (pairing mode) — unchanged

`pigeon.Register` leaves the instance ID empty when `Pairing == nil`
(`api.go`). `pairing.Accept` uses that path and then puts the
relay-assigned session ID in the QR token, separate from the stable
`AcceptorInstance`.

Greeting stays `register` or `register:<token>:`. The relay calls
`generateID()` (128-bit hex) and writes that id back, as today. No
challenge, no capability. `listen[:token[:id]]` for that instance stays
as today.

Those hex IDs are not later claimable with a proof: a proof is admitted
only when `requestedID == InstanceID(xPub)`, and a random hex id is not
that encoding. The slot dies when the ceremony listener closes. 🎯T56
still refuses to displace the live ceremony slot.

### Requested ID — `register-pop`

New union variant:

```
register-pop[:<token>[:<id>]]
```

Same colon suffix shape as `register`. Non-empty `id` is required.
Empty `id` on `register-pop` is rejected.

`Auth.VerifyRegister` runs first, exactly as today (token, mTLS, or a
custom hook). A nil `Auth` (open relay) does not skip the proof. No
verifier can waive it. Failing the token check does not consume a
challenge.

Then, on that connection only:

1. Relay writes a challenge, 65 bytes: `0x01 || relayEphPub(32) || nonce(32)`.
   The ephemeral key and the nonce are fresh from `crypto/rand` for this
   attempt, discarded after the check, and never logged. Any version
   byte other than `0x01` is a hard failure.
2. Backend writes a proof, 64 bytes: `xPub(32) || tag(32)`.
3. Relay admits or closes the connection. There is no second attempt
   on the same connection.

The tag uses the same HKDF the identity already uses, so every SDK can
call its existing derive:

```
info = "pigeon-register-pop-v1" || nonce || relayEphPub || xPub || requestedID
tag  = HKDF-SHA256(ikm = X25519(myPriv, relayEphPub), salt = nil, info = info)
```

`requestedID` is the raw UTF-8 bytes from the greeting. The label,
nonce, and both public keys are fixed length, so appending the id is
unambiguous. The nil salt is deliberate: `DeriveSharedSecret(relayEphPub, info)`
on the backend and `hkdfDerive(shared, info)` on the relay must match.
The nonce is inside `info`, not the salt.

The relay recomputes the tag with `X25519(relayEphPriv, xPub)` and
compares in constant time. It rejects an xPub or a shared secret that
the platform X25519 refuses. It admits the connection only when the tag
matches **and**

```
requestedID == base64.RawURLEncoding(sha256(xPub)[:16])
```

using one shared definition of that derivation (today it is inlined in
`fileIdentity.InstanceID` and in the macOS keychain identity). A proof
for key A presented with key B's id is an `id mismatch`, not a free
registration under whichever id matches the key.

On success the relay writes two length-prefixed messages:

1. The instance id as UTF-8 (same payload as today's register reply).
2. A listen capability: 32 random bytes, base64url, no padding.

On failure the relay closes. Reason strings are `proof required`,
`bad proof`, and `id mismatch`. They do not contain the id, the public
key, or the tag.

### Listen capability

The register connection is only a liveness signal. Clients are bridged
onto parked `listen` connections (`session.go`). Today `handleListen`
admits anyone who knows the id and passes `VerifyRegister`. On a shared
token that is the same hole as register.

A proof-gated instance accepts listens only with the capability from
step 2, via a new greeting:

```
listen-pop[:<token>[:<id>[:<capability>]]]
```

The relay compares the capability in constant time to the one stored on
that instance. The old `listen` greeting is rejected for a proof-gated
instance. `listen-pop` is rejected for a relay-assigned instance.
`connect:<id>` is unchanged: clients must be able to name a backend,
and the id was never a credential. The capability is not part of the
instance id, is not put in a pairing record or a QR payload, and is
not logged. It rotates on every successful register, including an
eviction.

### Versioning

| Peer | Requested id today | After this target |
| --- | --- | --- |
| Go `pigeon.Register` | Activation mode sends `Identity.InstanceID()` (`api.go`, `transport.go` `dialRegister` / `dialRegisterWebTransport`) | Speaks `register-pop` and the proof via `Identity.DeriveSharedSecret`. `Listener` keeps the capability and sends `listen-pop` from `dialListen`. Pairing mode stays on the empty-id greeting. |
| C `pigeon_register` / ngtcp2 `PIGEON_ROLE_REGISTER` (`c/src/ngtcp2_transport.c` `build_handshake`, `c/src/listener_ngtcp2.c`) | Optional `self_instance_id` string, no key | A non-empty id requires the X25519 identity (public key plus a derive callback, or equivalent that never needs the relay to see private bytes). Non-empty id without a key is an error, not a silent old greeting. The listen dialer stores and presents the capability. |
| Swift `Ngtcp2Transport` `.register(selfInstanceID:)` | Passes the id through to C | Follows the C path. |
| Swift `PigeonCore.register` (`Sources/PigeonCore/PigeonRelay.swift`) | Always sends an empty id | Stays on that greeting. It does not claim a stable identity id. |
| Kotlin `register` (`PigeonConn.kt`) | Optional `instanceID` string | Same rule as C: non-empty id requires the identity and speaks `register-pop`; empty id is unchanged. |
| TypeScript `web/src/relay.ts` | Connect only (`connect:<id>`) | Connect greeting unchanged. Protogen still regenerates `web/src/wireGen.ts`, and the TS suite round-trips the new variants and the tag vector. No browser register client is added. |

`register`, `register-pop`, and `listen-pop` are generated into Go
(`wire_gen.go`), Swift, Kotlin, C (`c/src/wire_gen.c`), and TypeScript
(`web/src/wireGen.ts`). Wire-vector tests move with the generator, the
same way `wire_vectors_test.go` already locks stream headers.

Fail closed across versions:

- New relay, old activation backend (`register:<token>:<id>`): close
  `proof required`. The slot stays empty. A squatter on the old
  greeting cannot hold it either. The owner is offline until that
  backend is upgraded. Pairing-mode empty-id register keeps working.
- Old relay, new activation backend (`register-pop`): the old relay
  closes on the unknown prefix. Deploy the relay first, then the
  backends. The cutover locks owners out of their own ids until they
  upgrade; it does not hand the id to a squatter.
- Clients speaking `connect:<id>` do not upgrade for this.

## Eviction

🎯T56 stays the rule for anyone who has not proved the key: `hub.register`
refuses to displace a live instance, and `hub.unregister` deletes only
on pointer identity.

A **valid** proof is the exception, including when the occupant is live:

1. Under the hub lock, install the new instance.
2. Cancel and close the old one: control connection, listens still
   parked, and sessions already handed to `bridge`. `bridge` today
   follows the client context and would otherwise leave the incumbent
   holding live pipes (`session.go`). After eviction a new client
   connect reaches the new owner, and an in-flight client of the
   incumbent observes a disconnect and can reconnect.
3. The old connection's `unregister` is a no-op because it no longer
   owns the slot (the 🎯T56 check). It must not cancel the new instance.

Same-key overlap (a crashed backend whose QUIC idle timeout has not
fired, or two processes with one identity) is the same path: the later
proof wins. An unproven claimant never evicts.

An occupant who proved the key earlier and is still connected is
evicted by the key holder. That is the intended recovery when a
squatter registered before this relay binary was running and the
process was not restarted — and it is the recovery for a stale
connection of the real owner. The hub is in-memory, so a relay restart
already drops every squatter; eviction covers the relay that stayed up.

Someone who has stolen the identity private key can evict the owner.
DESIGN.md §2 already puts that case out of scope: rotate the identity
and re-pair.

## Logs

Info and warn lines on register, listen, and connect must not contain
the full instance id. The sites that do today are the `"id", id` and
`"id", instanceID` attributes in `webtransport.go` and `quicserver.go`
(`instance registered`, `instance disconnected`, `instance ID in use`,
listen and connect not-found / park lines). A prefix of at most four
characters is enough to correlate a line. The challenge nonce, both
ephemeral and identity public keys, the tag, and the listen capability
are not logged.

## Formal model

Nothing under `formal/` models hub admission.

- `PairingCeremony.tla` `relay_registered` is an abstract acceptor
  event. It does not mention who is allowed to claim an id.
- `PathSwitch.tla` `backend_register` moves the relay actor from Idle
  to BackendRegistered with no id and no key.
- `SessionMachine.tla` assumes a `ValidPairingRecord` and does not
  model the relay hub.

Leave those three specs alone. Add a small hand-written
`formal/RelayRegister.tla` (the hub is not a protogen machine) with at
least:

- `UnprovenClaimRejected` — a claimant without a valid proof does not
  become the occupant of a requested id, whether or not the slot was
  free.
- `IdBindsKey` — the occupant of a proof-gated id is the holder of the
  key that hashes to it.
- `OwnerEvicts` — a valid proof becomes the occupant even if the slot
  was live.
- `NoEvictWithoutProof` — an unproven claimant leaves a live occupant
  in place.
- `AssignedIdUngated` — an empty request is admitted without a proof
  and does not satisfy a later proof for that id.

`make bullseye` runs TLC on it alongside `PairingCeremony` and
`SessionMachine`, and the success count in the Makefile goes from 2 to
3. Do not add `PathSwitch` here; that gap is 🎯T60.6.

## Tests

- `TestE2ERegisterSquatRejected` — owner registers with a proof, drops
  the control connection, and is offline. A second party who knows the
  id and (if configured) the bearer token, but not the identity key,
  is rejected on both the old `register:<id>` greeting and a
  `register-pop` proof under a different key. The owner registers again
  and a client `connect` reaches them.
- `TestE2ERegisterOwnerEvictsIncumbent` — while a proven registration
  is live, a second proof from the same identity takes the slot; the
  first control connection and its listens close; a client reaches the
  new listener. A proof from a different identity does not evict and
  does not park a listen.
- Unit tests for a wrong key, an id that does not match `xPub`, a
  replayed tag against a fresh nonce, and a relay-assigned id that
  cannot be claimed later with a proof.
- A single tag vector (fixed keys, nonce, id, expected 32-byte tag)
  checked in Go, Swift, Kotlin, C, and TypeScript. Hand-mirror it the
  way `wire_vectors_test.go` already does. Do not block on
  `vectors.json` being wired into every SDK (that is the entropy-audit
  follow-up).
- The existing pairing ceremony still receives a relay-assigned id and
  does not speak `register-pop`.
- `make bullseye` stays green, including the third TLC run.

## Docs when implementing

`docs/DESIGN.md` §2 gains the admission rule next to the `Auth` hook:
the hook decides whether the peer may register at all; the proof
decides whether they may hold this id. `agents-guide.md` says the same
and says `PIGEON_TOKEN` does not substitute for the proof.
`STABILITY.md` names `register-pop` and `listen-pop`. This note's
status line flips to implemented and names the tests above.

## Out of scope

- Hiding the instance id. `connect:<id>` keeps carrying it.
- Changing `InstanceID` derivation, pairing records, or the Ed25519
  signing key.
- A bearer token, mTLS, or a custom `Auth` as a stand-in for the proof.
- An adversary who already has the identity private key.
- Persistent hub state across relay restarts.
- Pulling `PathSwitch.tla` into `make bullseye`.

# Adversarial Modeling with Formal State Machines

**Companion to:** [`formal-state-machines.md`](formal-state-machines.md)  
**Audience:** engineers and architects who already accept declarative
state machines + TLA+ (or are evaluating them) and need to reason about
**who can break what**, including powerful insiders—not only external
network attackers.

**Status:** method essay grounded in pigeon’s pairing / session formal
models (YAML `adversary:` blocks, TLC properties, and the evolution from
a fuller Dolev–Yao-style model to a focused MitM model). Not a threat
catalogue for any particular industry.

**How to use this document**

| Reader | Start here |
|--------|------------|
| Human, 15 minutes | §§1–3, §6 |
| Human designing a model | §§3–5, §7 |
| Language model (threat / architecture strategy) | §1 claims, §4 taxonomy, §5 method, §7 playbook, §9 agent checklist |

---

## 1. Thesis

1. **Correctness under honesty is not enough.** A state machine that
   only models cooperative parties proves “happy path + some faults.”
   Security and integrity properties require an explicit **adversary
   process** whose actions TLC can interleave with honest steps.
2. **An adversary is just another set of transitions** over shared
   state (channels, stores, knowledge sets). It is not a prose
   paragraph in a threat model PDF; it is code the model checker
   executes exhaustively (within bounds).
3. **The question is not “can we stop every adversary?”** It is:
   *given this adversary’s modalities, which invariants still hold, and
   which residual powers remain?* Residual powers drive architecture
   (detection, dual control, cryptographic binding, audit, narrow
   privileges)—not wishful deletion of the adversary.
4. **Network adversaries and privileged internal adversaries share a
   modelling pattern.** Network MitM rewrites messages in flight;
   privileged insiders rewrite **persistent state** or **control
   inputs** the machine trusts. Same technique: adversary actions +
   knowledge/capability sets + properties that forbid bad outcomes.
5. **Adversary models must obey the same fidelity discipline as crypto
   operators.** An idealised adversary that cannot do what a real
   attacker can (or a model that assumes detection the product does not
   implement) produces **false assurance**—the dual of over-strong
   crypto primitives in the honest model.
6. **Adversary modelling is where formal methods pay for security
   design, not only for concurrency bugs.** Path-switch bugs are found
   by resource invariants; *who can force a bad pairing or a forged
   approval* is found by adversary actions + security properties.

These claims extend [`formal-state-machines.md`](formal-state-machines.md)
§1 (pure machines, codegen, TLA+, fidelity). Adversarial modelling is
how security *questions* enter that closed pipeline.

---

## 2. Why tests alone miss this

Tests pick **particular schedules**. An adversary model asks TLC to
explore **all interleavings** of honest transitions with attacker
transitions (bounded).

| Assurance style | What it explores | What it typically misses |
|-----------------|------------------|---------------------------|
| Unit / E2E tests | Fixed scripts | Rare orderings; creative attacker steps |
| Fuzzing | Random inputs | Structured multi-step protocol goals |
| Code review / LLM review | Opinion | Exhaustive residual power analysis |
| **TLC + adversary actions** | Reachable graph under stated attacker | Only what the adversary *cannot* do in the model (abstraction gap) |

The method is not “replace tests.” It is: **tests for codecs and
regression; adversary model checking for residual attack surface.**

---

## 3. Pattern: adversary as a first-class process

### 3.1 Structure in the formal pipeline

In the declarative spec (pigeon: YAML under `protocol/*.yaml`):

```yaml
# Adversary-only variables (not part of production runtime state)
variables:
  adversary_keys: { initial: '{}' }
  adversary_knowledge: { initial: '{}' }   # historical / fuller models
  adv_eph_pub: { initial: '"adv_eph"' }

adversary:
  - name: MitM_hello
    desc: intercept hello and substitute adversary ephemeral pubkey
    code: |
      await Len(chan_initiator_acceptor) > 0 /\ ...
      # rewrite message on the wire; update knowledge
  - name: MitM_welcome
    ...

properties:
  - name: MitMDetectedByCodeMismatch
    kind: invariant
    expr: '(adv_eph_pub \in adversary_keys /\ ...) => acceptor_code /= initiator_code'
```

`protogen` lifts `adversary:` into TLA+ actions that appear in `Next`
alongside honest actor steps. TLC explores sequences such as:

```
Honest₁ → Adv_MitM_hello → Honest₂ → Adv_MitM_welcome → …
```

Production executors **do not** implement the adversary. The adversary
exists only in the verification artefact. That split is intentional:
honest machines generate real code; attacker modalities generate only
the model that stresses them.

### 3.2 Knowledge and capability sets

Two complementary state components:

| Component | Role |
|-----------|------|
| **Knowledge** | What the adversary has observed or derived (tokens, messages, secrets, ciphertexts). Grows by eavesdrop / decrypt / shoulder-surf actions. |
| **Capability / keys** | What the adversary can *use* as if it were a peer (injected pubkeys, session keys under MitM, forged credentials). |

Properties then say things like:

- *If* MitM keys are in `adversary_keys` *and* both sides show codes,
  *then* codes differ (detection).
- *If* MitM succeeded at the key layer, *then* pairing never reaches
  the trusted terminal state (prevention / human abort).
- No element of type `plaintext_secret` ever appears in
  `adversary_knowledge` (secrecy).
- `used_tokens ∩ active_tokens = {}` (no token reuse), even when the
  adversary races or replays.

### 3.3 Properties are residual-power statements

A green invariant under adversary actions means:

> **Across the entire explored attack surface of this model, this bad
> thing does not happen.**

It does *not* mean “no attacker exists in the world.” It means the
architecture **closes** that class of attack *as modelled*, or forces
detection (e.g. human code comparison) before a trusted state is
reached.

When TLC finds a counterexample, the output is a **trace**: a concrete
sequence of honest and adversary steps to a violated state. That trace
is an architectural finding—often more valuable than a bug ticket—because
it shows *which modality* was sufficient.

---

## 4. Taxonomy of adversary modalities

The same formal machinery covers multiple **trust positions**. Name the
position; do not conflate them.

### 4.1 Network / on-path adversary (classical)

**Position:** controls or sits on channels between honest parties
(relay, WAN, compromised proxy).

**Typical actions (pigeon pairing lineage):**

| Action (illustrative) | Modality |
|-----------------------|----------|
| Eavesdrop / learn message | Observation |
| Replay captured message | Freshness attack |
| Drop / delay | Liveness / DoS (often separate properties) |
| **MitM substitute pubkey** in `hello` / `welcome` | Active impersonation on key exchange |
| Re-encrypt secrets under MitM keys | Pivot after key substitution |
| Concurrent pair with shoulder-surfed token | Race using out-of-band leak |
| Fabricated token | Brute / guess enrollment secret |
| Session replay of `auth_request` | Stale authenticator |

**Properties that answer residual power:**

- Detection: `MitMDetectedByCodeMismatch`
- Prevention of trusted completion: `MitMPreventsPairing` / `MitMPrevented`
- Secrecy: `DeviceSecretSecrecy`
- Enrollment hygiene: `NoTokenReuse`
- Auth binding: `AuthRequiresCompletedPairing`, `NoNonceReuse` (session phase)

Pigeon’s **current** `protocol/pairing.yaml` concentrates on **MitM
pubkey substitution** on hello/welcome (the load-bearing active attack
for SAS). A **richer historical model** (session/pairing lineage
documented in `docs/session-protocol.md` and residual tables in
generated session artefacts) enumerated on the order of **eight**
attack capabilities including observation, re-encryption, concurrent
pair, token brute-force, code guessing, and session replay. Both
shapes are valid: start rich to discover residual power; slim the
adversary when state space or phase split demands it—**document what
was dropped**.

### 4.2 Privileged / internal adversary (state-tampering)

**Position:** can write values that honest machines *treat as truth*:
databases, config stores, message queues, admin APIs, batch jobs,
“support” tools—not necessarily the network path.

**Typical actions (general pattern, not pigeon-specific code):**

| Action class | Example modality |
|--------------|------------------|
| **Corrupt durable state** | Flip a status enum, rewrite balances, reassign ownership IDs |
| **Inject control events** | Forge “approved”, “settled”, “paired” as if from a trusted actor |
| **Advance or rewind versions** | Reset sequence numbers, reuse nonces, resurrect revoked tokens |
| **Widen privilege** | Grant self a role bit the machine later trusts |
| **Clock / schedule abuse** | Jump timeouts, expire others’ locks early |
| **Partial delete** | Remove audit rows or revoke flags while leaving live grants |

**Formal encoding:**

```text
Adv_CorruptStatus ==
  /\ \E id \in EntityIds :
       store' = [store EXCEPT ![id].status = "APPROVED"]
  /\ UNCHANGED << ... honest-only vars ... >>
```

Or, more tightly, parameterise *what* the insider may touch:

```text
Adv_WriteField(f) ==
  /\ f \in InsiderWritableFields   \* architecture: minimise this set
  /\ ...
```

**Properties that answer residual power:**

| Property shape | Intent |
|----------------|--------|
| `TrustedState ⇒ CryptographicWitness` | Terminal states require unforgeable evidence, not only a DB flag |
| `NoOrphanGrant` | Privilege bits cannot exist without a justified machine path |
| `RevocationMonotone` | Once revoked, never re-active without a full re-enrollment path |
| `AuditComplete` | Every privilege-changing transition appends an append-only log entry the adversary *cannot* delete (separate store with harder trust) |
| `DualControl` | High-impact transitions require two independent events (maker–checker) |
| `SeparationOfDuties` | Adversary process is forbidden from holding two roles in one trace |

**Closeness to pairing MitM:** high. MitM rewrites **messages in
channels**; the internal adversary rewrites **variables the machine
reads as inputs**. Both are *untrusted writes into the model*. The
pairing token race (`concurrent_pair` with a surfed token) is already
halfway between network and insider: the secret left the intended
channel (QR) into adversary knowledge, then re-entered as a normal
message.

### 4.3 Compromised legitimate principal

**Position:** honest *code* path, stolen credentials or malware on a
real device/user.

Modelled either as:

- the honest actor’s actions **plus** extra actions using their keys, or
- a second “malicious initiator” process with the same interface.

Properties shift from “MitM detected” to “blast radius bounds”
(rate limits, step-up auth, device binding, anomaly thresholds)—often
statistical in production, but **safety cores** (e.g. cannot approve
own grant) remain invariants.

### 4.4 Supply-chain / generator adversary (meta)

Out of band for TLC of the protocol, but in scope for the *method*:
trust the generator and CI. Mitigations are process (reproducible
builds, reviewed protogen, standing TLC)—see fidelity in
[`formal-state-machines.md`](formal-state-machines.md) §5.3.

---

## 5. Method: from threat story to checked residual power

### 5.1 Steps

1. **Name the honest machine** (states, events, commands, durable
   variables)—per the companion essay.
2. **Name the adversary’s trust position** (§4). One model may include
   several positions as *separate* action groups; do not merge them
   into one omnipotent god without saying so (omnipotence proves
   nothing useful).
3. **List modalities as actions** with preconditions (`await` /
   guards). Prefer *small, named* attacks over one `Adv_DoAnything`.
4. **Define knowledge / capability evolution** for each action.
5. **Write properties as residual-power claims** (“even if X, not Y”).
6. **Run TLC.** On violation: read the trace; fix **architecture**
   (stronger binding, dual control, crypto witness) or **narrow the
   writable set**—not only the property text.
7. **Fidelity check:** can a real attacker in that position do more
   than the model? Less? Update actions until the model is a
   *sound over-approximation* of the real modality you care about.
8. **Phase-split** if adversary × honest state space explodes
   (pigeon: pairing with adversary; transport often without—document
   why).

### 5.2 Designing for residual power (the architectural payoff)

Model checking an adversary is most valuable when the answer is:

> The adversary **can** still do *A*, therefore we will change the
> system so *A* requires *B* (second factor, HSM signature, two-person
> rule, append-only log, air-gapped ceremony…).

Examples of architecture moves driven by residual power:

| Residual finding | Architectural response |
|------------------|------------------------|
| MitM can align keys if SAS too weak / no commitment | Commitment round; more entropy; out-of-band channel assumptions documented (pigeon open item T52) |
| Insider can set `status=APPROVED` in DB | Approval is a signature over payload, verified in the machine; DB flag is cache only |
| Token can be reused after revoke | Explicit `used_tokens`; machine rejects non-active tokens |
| Replay of auth | Nonces / session diversifiers in the machine and on the wire |
| Single admin can grant and settle | Dual-control events from two roles |

This is the loop: **model → residual power → tighter architecture →
model again.**

### 5.3 Honest liveness vs adversarial safety

Always keep both:

- **Safety under adversary:** bad states unreachable even when Adv
  steps fire.
- **Liveness under honesty:** without Adv (or with Adv restricted),
  the protocol still completes (`HonestPairingCompletes`).

A system that is “secure” because it never progresses is useless. TLC
should check both classes (invariants vs leads-to / liveness under
fairness).

### 5.4 Fidelity traps (adversary edition)

| Trap | Symptom | Fix |
|------|---------|-----|
| Under-powered adversary | Green properties, real attack works | Add the real modality as an action |
| Over-powered adversary | Everything fails; no design signal | Split positions; parameterise writable sets |
| Detection not implemented | Model assumes human abort; UI skips compare | Product must implement the detect path |
| Idealised crypto in Adv path | Model MitM “always” breaks codes; 20-bit SAS grindable | Align bit widths / commitment with implementation ([`formal-state-machines.md`](formal-state-machines.md) §5.3) |
| Adv only on network, prod trust is DB | Classic external threat model, insider free | Add §4.2 actions on stores |

---

## 6. Worked sketch: pigeon pairing (network MitM)

**Honest goal:** two devices share a pairing record only if a human
accepts matching confirmation codes derived from both ephemeral
pubkeys.

**Adversary position:** on-path between initiator and acceptor
(e.g. malicious relay or network attacker).

**Actions (current focused model):**

1. `MitM_hello` — replace initiator ephemeral pubkey with `adv_eph_pub`.
2. `MitM_welcome` — replace acceptor ephemeral pubkey with `adv_eph_pub`.

**Knowledge / keys:** `adversary_keys` gains `adv_eph_pub`; saved real
ephemerals for potential further pivots in richer models.

**Checked residual power:**

- If injection occurred and both codes exist ⇒ codes **differ**
  (`MitMDetectedByCodeMismatch`).
- If codes differ ⇒ both sides do **not** both sit in `Paired`
  (`MitMPreventsPairing`)—relying on human cancel.
- Without adversary ⇒ codes **match** (`HonestPairingMatchesCodes`).
- Honest completion still possible (`HonestPairingCompletes`).

**Architectural implication already in the product:** confirmation code
ceremony is not UX chrome; it is the **detection mechanism** the proof
depends on. Weak SAS / missing commitment (audit T52) is therefore a
**fidelity and residual-power** issue: the model’s detection claim is
stronger than a grindable 6-digit code under a patient MitM.

**Richer historical modalities** (token surf, concurrent pair, replay,
secret re-encrypt) show how the same framework stretches toward
enrollment theft and session attacks—exactly the bridge toward
“insider who can read a QR off a screen” or “operator who can resubmit
a captured auth.”

---

## 7. Playbook: adding an internal (state) adversary

Use this when the concern is *trusted writes*, not only *trusted wires*.

### 7.1 Inventory trust anchors

List every durable field and event source the machine treats as true:

- DB columns / documents  
- Caches that can become source of truth  
- Admin APIs and “break glass” tools  
- Message buses that can be published to with high privilege  
- Cron / batch that mutates state  

### 7.2 Partition fields

| Class | Rule |
|-------|------|
| **Machine-only** | Only honest transitions may write; Adv has no action on these (enforce in model and in DB permissions) |
| **Insider-writable** | Explicit `Adv_Write*` actions; properties must still hold |
| **Evidence-bearing** | Values must be accompanied by unforgeable witness (signature, HSM, second party); Adv can write the cache flag but not the witness |

The architectural win is **shrinking Insider-writable** and **moving
high-impact outcomes into evidence-bearing**.

### 7.3 Minimal internal-adversary package

1. One honest lifecycle machine (e.g. grant → approve → activate →
   revoke).
2. `Adv_FlipStatus` on the status field only.
3. Invariants:
   - `Active ⇒ ValidWitness(grant)`  
   - `Revoked ⇒ ¬Active` (monotone)  
4. TLC until green **or** a trace that forces dual control / witness.
5. Implement the architectural fix; regenerate; re-check.
6. Only then expand Adv (delete audit, clock skew, multi-row).

### 7.4 What “success” looks like

Not “no insider can ever cause harm.” Success is:

> We can **state** the insider’s modalities, **show** which harms are
> impossible under those modalities, and **list** residual harms with
> compensating controls outside the machine (monitoring, background
> checks, physical security)—each residual explicitly accepted.

That is an auditable security argument. Opinions—human or model—are
not.

---

## 8. Relationship to the rest of the method

| Piece | Role |
|-------|------|
| [`formal-state-machines.md`](formal-state-machines.md) | Pure command-emitting machines; codegen; TLC in CI; fidelity; adoption |
| **This document** | Adversary as transitions; residual power; network vs privileged positions |
| [`session-protocol.md`](session-protocol.md) | Concrete pairing/session properties and journey |
| [`DESIGN.md`](DESIGN.md) §2, §4 | Threat model (relay opacity); verification-first wire principle |
| `protocol/pairing.yaml` `adversary:` | Live example of MitM actions + properties |
| Fable-5 audit | Case where idealised detection/crypto outran implementation |

**Composition note:** adversary-rich pairing models **multiply** state
space. Do not fold them into transport models without need (see
companion essay on interface contracts and state-space explosion).

---

## 9. Checklist for an agent producing a strategy

When asked to plan adversarial modelling for a system, emit:

1. **Honest machines** in scope (names, durable variables).
2. **Adversary positions** (§4) in priority order—not one god-mode Adv.
3. **Modality catalogue** as named actions with preconditions.
4. **Knowledge / capability variables** and how each action updates them.
5. **Property list** phrased as residual-power claims (safety + honest
   liveness).
6. **Field partition** (machine-only / insider-writable / evidence-bearing).
7. **Fidelity table:** real attacker vs modelled actions.
8. **TLC budget** and phase split plan if needed.
9. **Architecture back-pressure:** expected control types if TLC finds
   traces (witnesses, dual control, narrower grants).
10. **Acceptance:** CI runs TLC with adversary enabled for the security
    phase; traces stored as artefacts on failure.

Do not invent sector-specific compliance labels unless the user supplies
that context. Stay in the vocabulary of machines, modalities, residual
power, and fidelity.

---

## 10. One-paragraph summary

> Put the attacker **inside the formal model** as explicit transitions
> over channels and stores, with knowledge and capability sets that
> grow under those actions. Use TLC to explore interleavings with honest
> state machines and to prove **residual-power invariants**: what still
> cannot happen even when the adversary steps. The same pattern covers
> classical on-path MitM (rewrite messages) and privileged insiders
> (rewrite trusted state). Counterexample traces drive architectural
> hardening—cryptographic witnesses, dual control, monotone
> revocation—not only bug fixes. Keep adversary models fidelity-checked
> against real modalities, and keep them phase-scoped so verification
> stays tractable. Security arguments then attach to **checked artefacts
> under a stated attacker**, not to the hope that tests hit the right
> schedule.

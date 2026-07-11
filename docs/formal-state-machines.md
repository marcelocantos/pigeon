# Declarative State Machines, Formal Verification, and Deterministic Code Generation

**Audience:** engineers, architects, and agents evaluating how to design
systems whose control logic must be *correct by construction* across
languages and over long maintenance horizons.

**Status:** architectural essay grounded in the pigeon project’s lived
history (2026-03 → 2026-07). Not a protocol reference. Protocol detail
lives in [`session-protocol.md`](session-protocol.md) and
[`DESIGN.md`](DESIGN.md) §4. **Adversarial modelling** (network MitM and
privileged state-tampering as TLA+ actions, residual power) is covered
in the companion [`adversarial-modeling.md`](adversarial-modeling.md).

**How to use this document**

| Reader | Start here |
|--------|------------|
| Human, 15 minutes | §§1–3, §6, §8 |
| Human, full case | All sections |
| Language model (strategy / adoption plan) | Treat §1 claims as axioms, §3 as the normative pattern, §5 as failure modes, §7 as adoption playbook, §9 as glossary |
| Security / residual attack surface | Read with [`adversarial-modeling.md`](adversarial-modeling.md) |

---

## 1. Thesis (normative claims)

These claims are the product of building and breaking a multi-language
protocol stack. They are stated as propositions so an agent or reviewer
can accept, reject, or refine them explicitly.

1. **Control complexity belongs in a declarative state machine**, not in
   ad-hoc control flow scattered across handlers, goroutines, and
   language-specific “glue.”
2. **The machine is a pure function**  
   `(state, event, guards) → (state′, variable updates, commands)`.  
   Side effects are not inferred; they are **emitted** as an ordered
   command list. The runtime is a thin executor: wait for event →
   step machine → run commands.
3. **A formal specification is the single source of truth** for that
   machine. The file format (YAML, DSL, whatever) is incidental. What
   matters is: one parse tree → all language executors **and** a TLA+
   model.
4. **Formal verification (TLA+ / TLC) is a first-class motivation**, not
   a documentation afterthought. Multi-language reuse is a corollary of
   generation; verification is the reason the pipeline exists.
5. **Deterministic code generation closes the loop.** Spec → executable
   tables/machines + TLA+ + diagrams. Correctness arguments attach to
   the *spec*, not to “an LLM’s opinion of the code” or to N hand ports
   that silently drift.
6. **Composition is a model-checking hazard.** Independent machines with
   **interface contracts** (e.g. `ValidPairingRecord(r)`) beat one giant
   joined model when state space explodes.
7. **Verification is only as strong as model–implementation fidelity.**
   A green TLC run that idealizes a crypto primitive (injective codes,
   infinite nonces, challenge–response that the C never sends) is a
   *false sense of assurance*. Treat every TLA+ property as a contract
   the concrete runtime must honour with matching parameters.
8. **More formalism, carefully chosen, makes systems simpler** at the
   seams that actually fail: resource lifecycle, path switching, timers,
   fallbacks—not at the byte-plumbing layer (sockets, varints, AEAD
   seal/open).

---

## 2. Problem this architecture solves

### 2.1 Failure modes of ad-hoc control logic

Without a declarative machine mediating I/O:

| Failure mode | Typical symptom |
|--------------|-----------------|
| **Resource lifecycle bugs** | Dispatcher still reading a dead path; monitor not stopped; signal not reset after fallback |
| **Cross-language drift** | Go and Swift disagree on allowed transitions; interop only fails in the field |
| **Unverified security ordering** | “Auth after pairing” holds in tests for one schedule, fails under reordering / MitM |
| **Executor invents policy** | `OnChange` callbacks infer actions from variable diffs; deadlocks and missing steps |
| **Silent event drops** | Unmatched handlers deliver data outside the machine, masking missing transitions |
| **Review-as-proof** | Humans or LLMs “look correct” over N ports; no exhaustive exploration of adversary interleavings |

Pigeon’s path-switch work produced four deterministic bugs *all* in
resource lifecycle—not in the protocol story itself. Invariants such as
`DispatcherMatchesActive` and `MonitorOffOnFallback` would have forbidden
those states *before code was written* if the bindings had been machine
variables from the start. That is the empirical core of claim 8.

### 2.2 What testing alone cannot do

Unit and E2E tests exercise **particular schedules**. TLC explores the
**reachable state graph** under a stated fairness and adversary model
(within bounded parameters). For security and concurrency protocols,
those are different tools:

- Tests: regression oracles, byte-vector parity, soak, race detectors.
- Model checking: exhaustive search for invariant / liveness violations
  under message reordering, loss, and (when modelled) Dolev–Yao style
  interference.

Neither replaces the other. The architecture insists on both, wired to
the same specification.

### 2.3 Why “LLM review” is the wrong primary assurance

Language models are excellent at *drafting* specs, *explaining* designs,
and *proposing* strategies. They are **not** a substitute for:

- exhaustive model checking of a formal transition system, or
- deterministic generation that keeps N language ports bit-aligned.

An organisational strategy that leans on model opinions for “is this
secure?” without a closed formal loop reintroduces the same drift and
false confidence this architecture was built to eliminate.

---

## 3. The normative architecture pattern

### 3.1 Machine as I/O mediator

```
  Application API  ──events──▶  ┌──────────────────┐  ──commands──▶  I/O / timers / app replies
                                │  Pure state       │
  I/O completions  ──events──▶  │  machine          │
  Timers / cancel  ──events──▶  └──────────────────┘
```

**Rules**

1. No goroutine / thread / task makes independent *policy* decisions
   about lifecycle. Readers post events; writers execute commands.
2. Application methods are thin: submit event, await response channel
   if needed.
3. If the machine has no transition for `(state, event)`, the event is
   **logged and dropped**—never silently fulfilled by a backdoor
   handler. Missing transitions are defects in the table, not
   opportunities for “helpful” default behaviour.
4. Crypto transforms, fragment reassembly, and pure data-plane codecs
   may sit *beside* the machine (stateless transforms). They must not
   *choose* path, retry, or session phase.

### 3.2 Command emission (not OnChange inference)

**Anti-pattern:** machine updates variables; executor watches diffs and
*guesses* which side effects to run.

**Pattern:** each transition lists `emits: [cmd…]` in order. The
executor runs that list and nothing else.

Why: variable diffs under-specify actions (one change may imply flush +
rebind + cancel timer + notify app). Inference reintroduces the bugs
the machine was meant to prevent (deadlocks on re-entrant locks,
missed flushes, ad-hoc `Recv` retry hacks).

### 3.3 Hierarchical states

When many leaf states share identical data-forwarding transitions,
**superstates with inherited transitions** (Harel-style hierarchy)
collapse N×M duplication into definitions on the container. Flatten for
codegen and for TLA+ export so executors and model checkers see only
leaf transitions.

### 3.4 Two concerns, two machines (when needed)

Prefer **small machines joined by contracts** over one mega-machine:

| Machine | Lifecycle | Example properties |
|---------|-----------|--------------------|
| Pairing / enrollment | One-shot per peer relationship | No token reuse, MitM via SAS mismatch, device secret secrecy |
| Session / transport | Many-shot per connection | Path consistency, backoff bounds, dispatcher binding, health monitor only when alt active |

Cross-cutting property example:

> No session becomes active without a prior successful pairing.

Decompose:

- Pairing proves: final success ⇒ `ValidPairingRecord(r)`.
- Session assumes: initial `r` satisfies `ValidPairingRecord(r)`.
- Composition is modus ponens in prose / review—not a single TLC model
  that multiplies adversary × path × health state spaces.

**Hard lesson:** a prior attempt to join pairing + transport into one
generated TLA+ model produced an intractable state space (hundreds of
millions of states, non-terminating TLC). Independent specs with an
interface contract restored sub-second verification.

### 3.5 What is *not* a state machine

Protogen’s mandate is broader than FSMs: **everything on the wire is
specified once and generated**. That includes:

- Multi-state FSMs (pairing, session, path switch).
- One-shot byte formats (greetings, stream headers, datagram
  plaintext layouts).

Sockets, QUIC stream plumbing, AEAD seal/open, and OS keystores are
**executors and libraries**. Forcing them into FSM tables creates work
without verification payoff. The correct split (from design dialogue):

- **FSM-owned:** negotiation, phases, retries, path preference,
  resource *bindings*, security ordering.
- **Plumbing-owned:** open/read/write/close, framing codecs, crypto
  primitives (with tests and, where applicable, separate crypto
  models).

---

## 4. The closed pipeline

### 4.1 Shape

```
                    ┌─────────────────────────────┐
                    │  Formal protocol spec       │
                    │  (states, events, guards,   │
                    │   actions, commands, props) │
                    └──────────────┬──────────────┘
                                   │ protogen
          ┌────────────┬───────────┼───────────┬────────────┬────────────┐
          ▼            ▼           ▼           ▼            ▼            ▼
       Go tables   Swift/KT/TS   C executor   TLA+      PlantUML     Wire
       + Machine   machines      headers      + TLC      diagrams    constants
```

In pigeon this carrier is YAML under `protocol/*.yaml` and the tool is
`cmd/protogen`. **The carrier is replaceable; the pipeline is not.**

### 4.2 Why generation beats hand ports

| Mechanism | Drift risk | Exhaustive check |
|-----------|------------|------------------|
| Hand-written N× language FSMs | High | None (tests only) |
| Shared library + FFI only | Medium (ABI) | None for protocol logic |
| Spec → N languages + TLA+ | Low (generator bugs) | TLC on properties in the *same* source |

“By construction” byte parity (cross-language vector tests) catches
codec drift. TLC catches **ordering and invariant** defects tests miss.

### 4.3 TLA+ engineering lessons (operational)

Practices that made model checking *usable* day-to-day:

1. **Pure TLA+, not PlusCal**, for generated models. PlusCal process
   interleaving and program counters inflated state without modelling
   value for these protocols. (PlusCal was historically accidental
   scaffolding, not a deliberate methodology choice.)
2. **Phase-aware export.** Generate pairing-phase and transport-phase
   specs separately; drop adversary actions where they have no
   semantics.
3. **Channel elimination.** Model at most one pending message per type
   as a struct variable (`received_msg`) instead of sequence channels.
   Result class: ~10² states / sub-second TLC instead of non-termination.
   Accept the abstraction (one-in-flight per type) only when the
   protocol truly has that shape.
4. **Resource bindings as state variables.** Dispatcher path, monitor
   target, backoff level—so transitions *must* update them.
5. **Fairness and leads-to** for liveness (recovery, re-advertise), not
   only safety invariants.
6. **Standing invariant in CI.** `make bullseye` (or equivalent) runs
   TLC every green build. Verification that is not automatic decays.

### 4.4 Runtime executor pattern

Pseudocode (language-neutral):

```
loop:
  e := waitEvent()                 // app | I/O | timer
  maybeRegisterWaiters(e)          // e.g. app_recv
  cmds := machine.HandleEvent(e)
  for c in cmds:
    execute(c)                     // write, timer, deliver, close, …
```

No second policy path. Timers are started/cancelled only via commands
(`start_pong_timeout`, `cancel_pong_timeout`), never by the executor
“noticing” that a ping was sent.

---

## 5. Case study: what pigeon’s history teaches

This section is evidence, not product marketing. Dates are approximate;
details live in git and `docs/session-protocol.md` appendix.

### 5.1 Timeline (compressed)

| Phase | What happened | Architectural lesson |
|-------|---------------|----------------------|
| Origin | Pairing ceremony as YAML → multi-lang + TLA+; rest of stack ad-hoc | Formalism works for the *hard security ritual* first |
| LAN / path | Ad-hoc swap then path router; four lifecycle bugs | Resource lifecycle must be *in* the machine |
| Unification | Pairing + transport in one conceptual session machine | Useful product model; dangerous as one TLC model |
| TLA+ rewrite | PlusCal → pure TLA+; phase export; channel elimination | Generator engineering is part of the method |
| Command emission (T18) | Kill OnChange inference; force fallback; drop unmatched delivery | Machine must emit *actions*, not only *state* |
| Hierarchy (T19) | Superstates for data forwarding | Scaling the *table* matters |
| Split verification (T39.2) | `SessionMachine.tla` + `PairingCeremony.tla` + `ValidPairingRecord` | Interface contracts replace composition |
| Multi-SDK / C core | One peer library, generated machines, vector parity | Generation + vectors scale languages |
| Security audit (Fable-5) | Critical/high findings at *seams* (model vs C, cgo, hub) | Fidelity and composition are the residual risk |

### 5.2 The design dialogue that sharpened the thesis

When multi-language C plumbing looked “too large,” the corrective
framing was:

> The state machine should define almost all session behaviour.
> Client libraries should not re-implement negotiation noise. Socket
> plumbing is not that noise.

And on terminology:

> “YAML” is metonymy for the **formal state-machine pipeline**. The
> file format is incidental. The substance is declarative FSM →
> generated executors → TLA+.

And on priority:

> **TLA+ verification is a primary motivation**—more important than
> multi-language reuse alone.

And on composition:

> Last time pairing + transport were joined for verification, state
> space exploded. Do not do that again.

These statements are the project’s internal doctrine; this document
elevates them to transferable method.

### 5.3 What green TLC does *not* prove (fidelity)

A deep audit later found cases where:

- the **model** treated confirmation-code derivation as injective over
  key material, while the **implementation** truncated to ~20 bits with
  no commitment round;
- the **model** described nonce/challenge-style auth, while the
  **runtime** accepted a bare device id on the wire.

The architecture’s response is not “abandon TLA+.” It is claim 7:

> Every verified property must be checked for **abstraction fidelity**
> against concrete parameters (bit widths, truncation, which fields
> actually appear on the wire, atomicity of counters).

Without that discipline, formal verification becomes theatre.

### 5.4 Payoffs observed

- **Bugs forbidden by invariant** rather than found late (dispatcher /
  monitor / signal class).
- **Cross-language behavioural consistency** when the machine is the
  sole policy authority (no unmatched-event backdoor).
- **Sub-second TLC** in CI after channel elimination—verification as
  *standing invariant*, not a heroics weekend.
- **Explicit backlog** when runtime still bypasses the generated
  machine in places: gaps are named targets, not tribal knowledge.

---

## 6. Benefits (structured for strategy docs)

### 6.1 Correctness and security

| Benefit | Mechanism |
|---------|-----------|
| Exhaustive exploration of interleavings (bounded) | TLC |
| Security ordering as invariants | Properties co-located with transitions in the spec |
| MitM / adversary modelling (where applicable) | Explicit adversary actions in pairing-phase models |
| No silent policy outside the table | Drop unmatched events; delete backdoor handlers |
| Reduced “looks secure” reliance | Arguments attach to checked properties |

### 6.2 Multi-implementation consistency

| Benefit | Mechanism |
|---------|-----------|
| Same transitions in every language | Codegen from one parse tree |
| Wire layout agreement | Generated codecs + shared test vectors |
| Diagrams that cannot rot | PlantUML (or equivalent) from the same source |

### 6.3 Operability and change

| Benefit | Mechanism |
|---------|-----------|
| Localised change | Edit one transition; regenerate; re-run TLC + tests |
| Onboarding | Diagrams + properties + table > archaeological reading of N ports |
| Agent / automation friendly | Spec is structured data; CI is deterministic |

### 6.4 Organisational (domain-agnostic)

| Benefit | Mechanism |
|---------|-----------|
| Reviewable policy | Security and product owners read tables and properties, not mutexes |
| Audit trail | Spec diff is the change; generated artefacts are rebuildable |
| Separates *policy* from *platform* | Platforms (QUIC, OS, crypto libs) swap under a stable machine |

---

## 7. Adoption playbook (for generating a strategy)

Use this as a checklist when applying the method outside pigeon.

### 7.1 Select the first machine

**Good first candidates**

- Multi-step protocols with security or money movement ordering.
- Lifecycle with retries, timeouts, and resource acquisition/release.
- Logic currently duplicated across services or languages.
- Areas with a history of “impossible” intermittent bugs.

**Poor first candidates**

- Pure data transforms (encoding, hashing).
- CRUD with no ordering constraints.
- Hot loops where table dispatch is the bottleneck (profile first;
  most control planes are not).

### 7.2 Minimum viable pipeline

1. Write the **spec** (states, events, transitions, guards, commands).
2. State **invariants and at least one liveness/leads-to** property.
3. Generate **one** production language executor + **TLA+**.
4. Put TLC in **CI** with a hard time budget; fix explosions via
   abstraction (phase split, channel elimination), not by deleting
   properties.
5. Delete or quarantine **backdoor handlers** that bypass the machine.
6. Add **fidelity checklist** for any crypto/numeric abstraction in
   the model vs code.
7. Only then fan out to additional languages.

### 7.3 Interface contracts between machines

For each hand-off artefact `R` (record, token, capability):

- Define `ValidR(r)` in the producer machine (postcondition).
- Assume `ValidR(r)` at the consumer machine’s initial state.
- Do **not** join machines in TLC unless a property truly cannot be
  decomposed and parameters can be tiny.

### 7.4 Executor rules (non-negotiable)

1. Event loop is single-threaded w.r.t. the machine (mutex or actor).
2. Commands are the only side-effect channel.
3. Timers are commands, not executor heuristics.
4. Unmatched events never deliver application data.
5. Tests include “force failure / force fallback” paths that only go
   through machine events.

### 7.5 Fidelity gates (non-negotiable for security properties)

For each invariant involving crypto or identifiers, document:

| Model concept | Concrete parameters | Evidence |
|---------------|---------------------|----------|
| Confirmation code | Bit width, commitment, who chooses ephemerals | Spec + tests |
| Nonce / sequence | Atomicity, per-session key diversifier | Race tests + model |
| Auth proof | What is on the wire vs model variables | Wire vectors |

Fail the gate if the model is strictly stronger than the runtime.

### 7.6 Anti-patterns to ban

| Anti-pattern | Replace with |
|--------------|--------------|
| OnChange / variable-diff side effects | `emits` command lists |
| “Helpful” default delivery on unknown events | Explicit transitions or drop |
| One TLC model for everything | Contracts + independent models |
| PlusCal generation without measuring state | Pure TLA+ actions; measure TLC |
| Hand-duplicated constants across languages | Generated wire_constants |
| “TLC green ⇒ ship crypto” | Fidelity review + tests |
| LLM-only security sign-off | Spec properties + TLC + vectors |

### 7.7 Success metrics

- TLC runtime in CI (budget; e.g. < 10s per machine).
- Number of production languages driven from one spec.
- Count of residual “policy outside machine” sites (drive to zero).
- Bugs found by TLC vs production (track both).
- Time from protocol change → all languages + green TLC.
- Fidelity checklist coverage for security properties.

---

## 8. Intellectual lineage (orientation, not a literature survey)

This method sits at a known intersection; pigeon’s contribution is a
**worked multi-language + CI-bound pipeline**, not a new formalism.

| Tradition | Relevance |
|-----------|-----------|
| **Lamport / TLA+** | Spec as transition system; TLC model checking; invariants and liveness |
| **Harel statecharts** | Hierarchy, orthogonal regions; inherited transitions for scaling tables |
| **Mealy machines** | Outputs (commands) on transitions, not only state labels |
| **SCXML / statechart interpreters** | Run-to-completion semantics; industrial FSM engines |
| **Protocol verification** | Adversary models; separate security ceremony from session lifetime |
| **Code generation / model-driven engineering** | Avoid N hand ports; risk of generator bugs (mitigate with tests + TLC) |

Agents drafting strategies should cite these traditions when mapping
the pattern into enterprise architecture languages, without claiming
pigeon invents them.

---

## 9. Glossary

| Term | Meaning here |
|------|----------------|
| **Spec** | Declarative description of a machine (states, events, transitions, guards, commands, properties). |
| **Executor** | Runtime that feeds events to the machine and performs emitted commands. |
| **Command** | Side effect the executor must perform (I/O, timer, app delivery, resource close). |
| **Event** | Input to the machine (app intent, I/O completion, timer, received message type). |
| **Interface contract** | Predicate shared by two machines (e.g. `ValidPairingRecord`) enabling independent verification. |
| **Channel elimination** | Modelling pending messages as single-slot structs instead of sequence channels to shrink TLC state. |
| **Fidelity** | Agreement between model abstractions and concrete runtime parameters/behaviours. |
| **Protogen** | Pigeon’s generator (`cmd/protogen`); stand-in for any tool that implements the closed pipeline. |
| **Standing invariant** | Check that must stay green on every merge (tests, TLC, formatters). |

---

## 10. Related documents in this repository

| Document | Role |
|----------|------|
| [`adversarial-modeling.md`](adversarial-modeling.md) | Adversary as transitions; residual power; network vs privileged insiders |
| [`session-protocol.md`](session-protocol.md) | Concrete session/pairing machine design + full journey appendix |
| [`DESIGN.md`](DESIGN.md) §4 | Wire protocols: generate everything; verification first; no composed TLA+ |
| [`DESIGN.md`](DESIGN.md) §1–2 | Goals, non-goals, threat model (product context) |
| [`protocol/*.yaml`](../protocol/) | Live formal specs (`adversary:` on pairing) |
| [`formal/`](../formal/) | Generated / maintained TLA+ and TLC configs |
| [`audit/fable-2026-07.md`](audit/fable-2026-07.md) | Case study of seam failures and fidelity gaps |

---

## 11. One-paragraph summary (for paste into briefs)

> Treat protocol and lifecycle control as **declarative state machines**
> that are **pure** with respect to events and that **emit commands**
> for all side effects. Keep a **single formal specification** as source
> of truth; **deterministically generate** language executors, wire
> constants, diagrams, and a **TLA+ model** checked by TLC in CI. Prefer
> **small machines joined by interface contracts** over monolithic
> models. Pair verification with a **fidelity discipline** so idealised
> model operators cannot outrun the concrete implementation. Use tests
> and cross-language vectors for codecs and races; use model checking
> for ordering and invariants. The goal is correctness arguments that
> attach to **checked artefacts**, not to opinions—human or machine—
> about hand-written control flow.

---

## 12. Checklist for an agent producing an adoption strategy

When asked to “generate a strategy from this document,” emit:

1. **Scope:** which subsystems become machines first (with rationale §7.1).
2. **Machine map:** names, events, commands, contracts between machines.
3. **Property list:** safety + liveness candidates; adversary model if any
   (detail: [`adversarial-modeling.md`](adversarial-modeling.md)).
4. **Pipeline plan:** tooling choice, CI budgets, generation targets.
5. **Fidelity plan:** table from §7.5 for every security property.
6. **Migration plan:** how to delete backdoor handlers; interim dual-run.
7. **Metrics:** from §7.7 with baselines.
8. **Risks:** state-space explosion, generator bugs, false TLC greens.
9. **Non-goals:** what remains plumbing / libraries.
10. **90-day milestones:** one machine green in CI end-to-end before
    multi-language fan-out.

Do not invent banking or industry-specific compliance language unless
the user adds that context; keep the strategy in the vocabulary of this
document.

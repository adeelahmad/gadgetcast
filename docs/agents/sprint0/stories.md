---
type: stories
sprint: 0
---

# Stories — Sprint 0: hardware-free Go core

The pure-logic, stdlib-only Go core of gadgetcast: the parts that need neither the
Raspberry Pi, WebRTC/GStreamer, nor the USB configfs gadget to exist. Four feature
areas (switcher, signaling codec, config, pairing) plus a shared foundation, all
exercised by unit tests with no external process, no socket, and no hardware.

Source of truth: `README.md` (§3.1, §3.2, §3.3, §4), `docs/agents/sprint0/intake.md`,
`docs/agents/standards.md`. Resolved open questions from the dispatch are baked in and
are NOT re-raised here.

## Sprint goal

Deliver a dependency-light (stdlib-only) Go core — module
`github.com/adeelahmad/gadgetcast`, go 1.22 — that later sprints wire hardware and
transport around: (1) a source-switcher state machine enforcing "exactly one client
on-air at a time," (2) a typed signaling message set with JSON codec + validation,
(3) a config model that parses/validates the HTTPS/bind/PIN/format knobs, and (4) a
per-client PIN pairing/auth state machine with a configurable attempt limit and
durable lockout. Pairing gates the switcher (an unpaired client can appear but cannot
go on-air), and the signaling `on-air-state` message is a projection of switcher
state — no duplicated on-air logic.

## Sprint demo

`go test ./...` passes and, in a scratch `main`, we: add three clients to a switcher,
show none on-air; set client A on-air (query reports only A); set B on-air (A is
displaced, only B on-air); remove B (on-air reports nobody); attempt to set an
unpaired client on-air and get a typed error; round-trip all seven signaling messages
through JSON; parse a well-formed config with zero errors and reject an HTTPS-disabled
config; and drive a client through `pending → paired` on the right PIN and
`pending → rejected` (durable) after exhausting the 5-attempt limit.

## Definition of Done

- Module `github.com/adeelahmad/gadgetcast` (go 1.22) exists with a single `go.mod`
  created in Wave 1 before any feature story starts (avoids parallel worktrees each
  minting a conflicting `go.mod`; see `docs/agents/memory.md` M-001).
- Stdlib only — no third-party dependencies anywhere in Sprint 0 (config uses
  `encoding/json`; resolved decision #7).
- All four feature areas implemented as pure-logic packages under `internal/`
  (`internal/core`, `internal/signaling`, `internal/config`, `internal/pairing`,
  `internal/switcher`), organized by feature not by type (`rules/golang/coding-style.md`).
- The gate matrix in `docs/agents/standards.md` is green: `gofmt -l .` empty,
  `go vet ./...` clean, `go build ./...` compiles, `go test ./...` passes, coverage
  ≥ 80% on these pure-logic packages.
- Every story's five-part Intent is honored by tests, including the four invariant
  tests: never-two-on-air, unknown-type-decode-errors-not-panics, brute-force-lockout,
  unpaired-cannot-go-on-air.
- No production code is written by the planner; execution happens under TDD (RED →
  SCAFFOLD → GREEN) per the playbook.

## Out of scope

Deferred to later sprints (verbatim from intake §"Out of scope"):

- Actual WebRTC/DTLS-SRTP negotiation and media transport.
- GStreamer pipeline (decode/convert/encode).
- V4L2/UVC/UAC2 gadget I/O (configfs, `/dev/videoN`, ALSA).
- Real HTTPS server wiring (TLS termination, cert issuance/trust) — the config model
  only *represents and validates* the HTTPS requirement; it stands up no listener.
- mDNS advertisement.
- Tailscale integration — config only *represents and validates* the bind-interface
  choice; it opens no socket and resolves no real interface.
- QR pairing (resolved decision #4: Sprint 0 is PIN-only).
- Any goroutine/channel concurrency: Sprint 0 state machines are single-threaded
  value logic, so `-race` stays advisory (standards Gate-matrix note).
- Anything needing hardware or a heavy external dependency.

## User stories

Each story carries the five-part Intent (What's wanted · Constraints · Failure
scenarios · Success scenarios · Connections). Failure-scenarios and Connections are
mandatory. Per-story task/validate/plan artifacts live in the story's `sN-NN-<slug>/`
directory.

---

### S0-00 — Core foundation (module + shared domain types)

Wave 1. Slug: `s0-00-core-foundation`. Package: `internal/core`, plus `go.mod`.

**What's wanted.** A single Go module (`go mod init github.com/adeelahmad/gadgetcast`,
go 1.22) and a shared `internal/core` package holding the value types every feature
package needs — principally `ClientID` — so the four Wave-2 stories all compile against
one agreed vocabulary and one `go.mod`.

**Constraints.**
- The module and `go.mod` are created exactly once, in Wave 1, and merged before any
  Wave-2 feature story starts (M-001; avoids conflicting parallel `go.mod` files).
- Stdlib only. `internal/core` imports nothing outside the standard library.
- `ClientID` is a strongly-typed value (not a bare `string`), usable as a map key and
  comparable with `==`, with a distinguishable zero/invalid value.
- Minimal surface (YAGNI): only what the feature packages provably need this sprint.

**Failure scenarios (NOT done if any hold).**
- `NewClientID("")` (or an all-whitespace id) returns a zero value with no error — an
  empty client identity must be rejected, not silently accepted.
- The zero `ClientID` reports itself as valid, letting "no client" masquerade as a
  real client downstream in the switcher.
- The package pulls in a third-party dependency, breaking the stdlib-only constraint.

**Success scenarios (done when).**
- `go build ./...` compiles the fresh module; `go test ./...` runs.
- `NewClientID("cam-1")` yields a valid `ClientID` whose `String()` is `"cam-1"`; two
  IDs built from the same string compare equal and key the same map slot.
- An empty/whitespace id is rejected with a typed error; the zero `ClientID` reports
  invalid.

**Connections.** `ClientID` is the pointer the switcher (S0-04) tracks as "who is
on-air" and the identity the pairing machine (S0-03) pairs; it appears in signaling
messages (S0-01, e.g. `join`, `on-air-state`). A wrong or ambiguous identity here
propagates to every other package — this is the shared vocabulary the whole sprint
builds on (intake §Connections).

---

### S0-01 — Signaling message types + JSON codec

Wave 2 (parallel). Slug: `s0-01-signaling-codec`. Package: `internal/signaling`.

**What's wanted.** The seven typed WebSocket signaling messages — `join`, `offer`,
`answer`, `ice-candidate`, `switch-source`, `on-air-state`, `error` — with a
`type`-tagged JSON envelope, encode/decode, per-message field validation, and a pure
projection helper that builds the `on-air-state` message from switcher-level values
(README §3.2; resolved decision #2 — this seven-message set is the authoritative
Sprint 0 wire contract).

**Constraints.**
- Decode produces strongly-typed structs, never a loose `map[string]interface{}`.
- Decode of any input NEVER panics and never silently substitutes a zero value — an
  unknown `type`, truncated JSON, or a type/payload mismatch returns a typed error.
- Stdlib `encoding/json` only (decision #7).
- The `on-air-state` message is a projection of switcher state (decision #8): the
  package exposes `OnAirStateFrom(id core.ClientID, live bool)` consuming
  switcher-level values (`core.ClientID` + bool) so it needs NO compile-time import of
  `internal/switcher` — the two stay parallel in Wave 2, and there is one copy of the
  on-air fact, owned by the switcher.

**Failure scenarios (NOT done if any hold).**
- Decoding a message with an unrecognized `type` panics, or returns a zero-value
  message with no error (intake failure #4; invariant: unknown-type-decode-errors).
- Truncated/malformed JSON panics instead of returning a typed decode error.
- An `offer` missing its SDP (or `join` missing its client id, `ice-candidate`
  missing its candidate, `switch-source` missing its target) validates as OK.
- A decoded message comes back as `map[string]interface{}` rather than a concrete
  struct.

**Success scenarios (done when).**
- Every one of the seven types round-trips encode → JSON → decode without loss and
  decodes to a validated, strongly-typed struct (intake success #4).
- `OnAirStateFrom(A, true)` yields a message naming A on-air; `OnAirStateFrom(zero,
  false)` yields a message stating nobody is on-air; both round-trip.
- Every malformed input in the decode table returns a typed error and no panic.

**Connections.** This codec is the wire contract between browser clients and the Pi's
future WebSocket server (README §3.2). `switch-source` and `on-air-state` must stay in
lockstep with the switcher's (S0-04) transition API and on-air query, or a later
sprint's transport is built against a codec that doesn't match the state machine it
drives (intake §Connections). `join`/`on-air-state` carry `core.ClientID` from S0-00.

---

### S0-02 — Configuration model (parse + validate)

Wave 2 (parallel). Slug: `s0-02-config-model`. Package: `internal/config`.

**What's wanted.** A JSON config model (via `encoding/json`) that parses and validates
the security/bind/format knobs README §4/§3.1/§3.2 call out: HTTPS on/off (+ cert/key),
bind-interface (all vs. Tailscale-only), PIN-required, PIN attempt limit, and the
video/audio format defaults — so a later sprint's real HTTPS server, Tailscale
integration, and pairing read a validated config at startup.

**Constraints.**
- HTTPS is a hard precondition: a config with HTTPS disabled fails validation, and
  HTTPS enabled with no cert/key material fails validation (README §3.2 —
  `getUserMedia`/`getDisplayMedia` refuse insecure contexts; §4 security posture).
- `bind=tailscale-only` combined with an all-interfaces / unresolvable target is a
  contradiction and is rejected — never a silent fall-back to binding everywhere
  (README §4; intake failure #6).
- PIN-required is an explicit field; bind-interface NEVER auto-disables PIN (resolved
  decision #6) — the two are independent mitigations.
- Default PIN attempt limit is 5 (resolved decision #3); omitted fields take documented
  defaults. Video default MJPEG primary / YUYV fallback at 1080p30, 720p30, 480p30
  (§3.1); audio default Opus at 48 kHz (§3.2/§4).
- Stdlib `encoding/json` only; validate at this boundary (standards §Security).

**Failure scenarios (NOT done if any hold).**
- A config with HTTPS disabled (or HTTPS on with no cert/key) parses and validates
  clean (intake failure #5).
- `bind=tailscale-only` with an all-interfaces / unresolvable target validates clean
  instead of being rejected (intake failure #6).
- Choosing `tailscale-only` silently flips PIN-required off (violates decision #6).
- Malformed JSON panics rather than returning a typed parse error.

**Success scenarios (done when).**
- A well-formed config — HTTPS on with cert/key, bind=all (or tailscale with a
  tailscale-shaped target), PIN required, video MJPEG/YUYV at 1080p30/720p30/480p30,
  audio Opus/48 kHz — parses and validates with zero errors (intake success #5).
- Defaults resolve correctly when fields are omitted, including attempt-limit = 5.
- PIN-required stays independent of bind choice.

**Connections.** The bind-interface and PIN-required fields are exactly what a later
sprint's real HTTPS server, Tailscale integration, and mDNS advertisement (§4, §8)
read at startup — the device's whole trust boundary rests on this validation being
correct before any real network code exists (intake §Connections). The attempt-limit
default feeds the pairing machine (S0-03); the video/audio defaults describe the UVC/
UAC2 gadget a later sprint builds.

---

### S0-03 — Client pairing / PIN auth state machine

Wave 2 (parallel). Slug: `s0-03-pairing-auth`. Package: `internal/pairing`.

**What's wanted.** A per-client PIN pairing/auth state machine: `pending → paired` on
the correct PIN, `pending → rejected` (durable) once a configurable attempt limit is
exhausted, plus an `IsPaired(core.ClientID) bool` query the switcher consumes (README
§4; resolved decisions #3, #4, #5).

**Constraints.**
- Attempt limit is configurable; default 5 (decision #3). Once the limit is exhausted
  the client is `rejected` and the rejection is durable — further attempts (even with
  the correct PIN) are refused, never allowed to retry indefinitely (README §4; intake
  failure #7; invariant: brute-force-lockout).
- PIN-only (decision #4 — QR deferred).
- A correct PIN transitions `pending → paired` exactly once and the client stays
  `paired` thereafter (idempotent).
- `IsPaired` is true only in the `paired` state (false for `pending` and `rejected`) —
  this is the predicate that couples pairing to the switcher (decision #5).
- Stdlib only; no hardcoded PIN in source (standards §Security — real PINs come from
  config/env in a later sprint; tests inject a known PIN).

**Failure scenarios (NOT done if any hold).**
- A client submitting wrong PINs past the configured limit is still allowed to keep
  retrying (unbounded retries defeat the pairing purpose; intake failure #7).
- A `rejected` client silently reverts to `pending` on the next attempt, or a correct
  PIN after lockout re-pairs it — rejection must be durable/queryable (intake success #7).
- `IsPaired` returns true for a `pending` or `rejected` client.

**Success scenarios (done when).**
- A client supplying the correct PIN within the limit transitions `pending → paired`
  exactly once and stays paired (intake success #6).
- A client exhausting the limit transitions `pending → rejected`, and the machine keeps
  reporting `rejected` (durable, queryable) on later queries/attempts (intake success #7).
- The attempt limit is honored as configured (e.g. limit 3 rejects on the 4th wrong
  PIN), defaulting to 5.

**Connections.** `IsPaired` determines which clients are even eligible to become the
switcher's on-air source (S0-04): README §4's security intent (an unauthenticated
gadgetcast on an open LAN is "a genuinely bad time") holds only if unpaired clients
cannot reach the on-air position (intake §Connections; decision #5). The attempt-limit
value originates in the config model (S0-02). The client identity is `core.ClientID`
(S0-00).

---

### S0-04 — Source-switcher state machine

Wave 2 (parallel; logically downstream — see dependency graph). Slug:
`s0-04-source-switcher`. Package: `internal/switcher`.

**What's wanted.** The source-switcher state machine: add-client, remove-client,
set-on-air, and an on-air query, enforcing "exactly one client on-air at a time" and
refusing to put an unpaired client on-air (README §3.3; intake (a); decisions #5, #8).

**Constraints.**
- Exactly one client on-air at any instant is a hard invariant that holds across every
  add / remove / set-on-air sequence, not just the steady state (README §3.3; intake
  failure #1; invariant: never-two-on-air). `set-on-air(B)` while A is on-air displaces
  A atomically — never a moment with both.
- `set-on-air` with an unknown client id returns a typed error and leaves on-air state
  unchanged — no phantom on-air client (intake failure #2).
- Removing the currently on-air client transitions on-air to "nobody," never a dangling
  reference to a removed client (intake failure #3).
- An unpaired client MAY be added/appear in the client set but CANNOT be set on-air:
  attempting to do so is a typed error (decision #5; intake failure #8). The switcher
  consumes pairing via a small locally-defined interface (accept interfaces — standards
  §API design), e.g. `PairingOracle{ IsPaired(core.ClientID) bool }`, so it does NOT
  import `internal/pairing` and stays parallel-buildable in Wave 2; tests inject a stub.
- The switcher is the single source of truth for on-air state; the `on-air-state`
  signaling message (S0-01) is a projection of `OnAir() (core.ClientID, bool)` — no
  duplicated on-air logic (decision #8). The switcher does NOT import
  `internal/signaling`; the contract is the shared `(core.ClientID, bool)` shape.
- Stdlib only; single-threaded value logic (no goroutines this sprint).

**Failure scenarios (NOT done if any hold).**
- Any observable point where two clients are simultaneously on-air across a
  set-on-air / switch / remove sequence (intake failure #1).
- `set-on-air` on an id not in the client set mutates on-air state or creates a phantom
  on-air client instead of returning a typed error (intake failure #2).
- Removing the on-air client leaves the on-air pointer referencing the removed client
  (intake failure #3).
- An unpaired (`pending`/`rejected`) client can be set on-air (intake failure #8).

**Success scenarios (done when).**
- N clients added, none on-air by default; `set-on-air(A)` makes A on-air and the query
  reflects A and only A (intake success #1).
- `set-on-air(B)` while A is on-air transitions A off and B on with the invariant
  holding at every step (intake success #2).
- `remove-client` on a non-on-air client leaves the on-air pointer untouched (intake
  success #3); removing the on-air client sets on-air to nobody.
- A paired client can be set on-air; an unpaired one is refused with a typed error.
- `OnAir()` returns the `(core.ClientID, bool)` source of truth that S0-01's
  `OnAirStateFrom` projects to the wire.

**Connections.** The switcher's on-air `ClientID` is the pointer the later WebRTC/
GStreamer pipeline reads to decide whose decoded A/V is written to the UVC/UAC2 gadget
(README §3.1, §3.3) — a bug here means the wrong participant's camera/mic reaches the
host, the core promise of the project (intake §Connections). It consumes pairing's
`IsPaired` (S0-03) and the `ClientID` vocabulary (S0-00); its on-air query is what
S0-01's `on-air-state` message projects. README §7 puts switcher/on-air in Phase 3 and
pairing in Phase 4 — Sprint 0 pulls both forward ahead of the UI/hardware that will
gate them.

## Story dependency graph

```
Wave 1 (foundation, merged before Wave 2 starts):
  S0-00 core-foundation ── creates go.mod + internal/core (ClientID)
        │
        │  every Wave-2 package imports internal/core
        ▼
Wave 2 (run in parallel — 4 stories, ~15 tasks):
  S0-01 signaling-codec ───────────┐
  S0-02 config-model               │  (independent of each other)
  S0-03 pairing-auth ──┐           │
  S0-04 source-switcher┘           │
```

Edges:

- `S0-00 → S0-01`, `S0-00 → S0-02`, `S0-00 → S0-03`, `S0-00 → S0-04`: hard compile
  dependency on `internal/core` (`ClientID`). This is why S0-00 is a Wave-1 gate: the
  single `go.mod` and `ClientID` must exist and be merged before the parallel feature
  stories branch, or each Wave-2 worktree mints a conflicting `go.mod` (M-001).
- `S0-04 ⇢ S0-03` (LOGICAL, not compile): the switcher needs pairing's "is this client
  paired?" predicate (decision #5). Resolved at compile time by a switcher-local
  `PairingOracle` interface (accept interfaces), so S0-04 does NOT import
  `internal/pairing`. Tests inject a stub oracle. Because there is no compile edge,
  S0-04 stays in Wave 2 and runs in parallel with S0-03 — it does not fall to a later
  wave.
- `S0-04 ⇢ S0-01` (LOGICAL, not compile): the `on-air-state` message is a projection of
  the switcher's `OnAir() (core.ClientID, bool)` (decision #8). The projection helper
  `OnAirStateFrom` lives in `internal/signaling` and consumes only `core.ClientID`+bool,
  so neither package imports the other; the shared `(core.ClientID, bool)` shape (owned
  by `internal/core`) is the whole contract. No compile edge → both stay parallel in
  Wave 2.

Explicit note (per dispatch): the two switcher couplings are real and are reflected as
LOGICAL edges above. They do NOT force S0-04 into a later wave because both are broken
at compile time via `internal/core` value types plus a switcher-local pairing
interface. All four Wave-2 stories therefore build and enter RED concurrently.

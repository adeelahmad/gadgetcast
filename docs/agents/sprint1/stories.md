---
type: stories
sprint: 1
---

# Stories — Sprint 1: hardware-free session hub + audio pure-logic pipeline

The pure-logic, stdlib-only layer that wires the Sprint 0 packages into a runtime
shape and adds the pure-math half of the audio pipeline — deferring all real
WebRTC/DTLS-SRTP, Opus, GStreamer, ALSA/UAC2, the WebSocket server, HTTPS/TLS, and
mDNS/Tailscale to a later hardware sprint. Three independent packages, each unit-
testable with no external process, no socket, and no hardware — the same bar Sprint 0
set.

Source of truth: `README.md` (§3.2, §3.3, §4, §7 Phase 1), `docs/agents/sprint1/intake.md`,
`docs/agents/standards.md`, `docs/agents/memory.md`. The eight open questions the intake
raised were RESOLVED by the orchestrator before this planning and are baked in below as
constraints; they are NOT re-raised here.

## Resolved decisions (baked in — not re-litigated)

- **D1 Admission timing.** On `join`, the client is admitted to the `switcher` set
  (visible). Pairing gates ON-AIR ONLY: an unpaired client MAY appear but `SetOnAir`
  errors. Consistent with Sprint 0's source-switcher story and decision #5. NOT "admit on
  pair."
- **D2 No clock.** No `Clock` interface in Sprint 1. `internal/drift` takes explicit
  `(fillLevel, dt)` inputs; the session hub is purely event-driven with no time-based
  logic (no session timeouts this sprint).
- **D3 Offer/Answer/ICE.** The hub RELAYS them as OPAQUE passthrough to the other
  connected peers (an addressed Effect each); it never interprets SDP/ICE.
- **D4 Disconnect.** On disconnect the hub removes the client from `switcher` (clearing
  on-air + broadcasting nobody-on-air if it was live); pairing state is DURABLE and
  retained on reconnect (Sprint 0 pairing story). `pairing.Manager` has no deregister —
  none is added.
- **D5 Broadcast shape.** The hub emits `[]Effect`, each an addressed outbound message
  (`To core.ClientID` + `signaling.Message`). A broadcast is expanded by the hub into
  one addressed Effect per connected client, in a deterministic roster order, so tests
  assert the exact Effect list. The transport is a dumb sink and is NOT built this sprint.
- **D6 Jitter depth.** Buffer capacity/depth is an injectable constructor parameter with
  a documented default.
- **D7 (b)↔(c) decoupled.** `internal/jitter` and `internal/drift` are independent this
  sprint; drift takes a synthetic fill-level series. Real wiring is a later integration.
- **D8 Drift control law.** A PROPORTIONAL controller with EMA smoothing, dead-band, and
  clamp: `ratio = 1 + Kp*fillError`, `fillError = (smoothed - target)/target`, clamped to
  a small injectable bound (default ±2%), with injectable `Kp`, `target`, clamp, and an
  EMA time constant driven by `dt`. Sign convention: sustained positive fill drift
  (buffer filling, fill > target) → ratio > 1 (drain faster); documented in code.

## Sprint goal

Deliver three pure-logic, stdlib-only packages that let a later hardware sprint attach
real transport/codec/ALSA around already-proven logic instead of discovering state-machine
and DSP bugs against real sockets and real audio hardware: (1) `internal/session` — a
single-threaded, event-driven hub `Process(event) -> []Effect` that drives a client
through connect → `join` → PIN-pair → switcher admission → (conditional) on-air →
disconnect over an injected transport, wiring all four Sprint 0 packages together for the
first time; (2) `internal/jitter` — a pure reordering buffer for seq/timestamped frames
that drops late + duplicate frames, emits in playout order, and reports underrun/overrun;
(3) `internal/drift` — the clock-drift resampler control law (README §4) turning
`(fillLevel, dt)` samples into a resample-ratio correction. All three import only Sprint 0
packages (already on the base), so they form ONE parallel wave.

## Sprint demo

`go test ./internal/session/ ./internal/jitter/ ./internal/drift/` passes and, in a
scratch `main`, we: build a hub over a `pairing.Manager` + `switcher.Switcher`, drive a
scripted event sequence (client A joins → is visible in the switcher but blocked from
on-air → submits the correct PIN → is routed on-air, and every connected client receives
an `on-air-state(A, live)` effect → A disconnects and every remaining client receives an
`on-air-state(nobody)` effect); push audio frames `3, 1, 2` into a jitter buffer and pop
them back as `1, 2, 3`, then show a late frame and a duplicate each dropped-and-reported
and a pop-on-empty reported as underrun; and feed the drift controller a flat fill series
(ratio 1.0), a sustained-rising series (ratio > 1.0), a sustained-falling series (ratio <
1.0), and a flat series with a single outlier (ratio stays in the neutral band).

## Definition of Done

- Three new pure-logic packages under `internal/` — `internal/session`,
  `internal/jitter`, `internal/drift` — organized by feature, stdlib-only, importing only
  Sprint 0 packages (`internal/core`, `internal/signaling`, `internal/config`,
  `internal/pairing`, `internal/switcher`) and nothing third-party (standards §Dependencies).
- The Sprint 0 public surfaces are reused EXACTLY as shipped; no Sprint 0 package is
  modified this sprint. `pairing.Manager.IsPaired` already satisfies
  `switcher.PairingOracle`, so the hub wires `switcher.New(pairingManager)` with zero
  adapter code.
- The `standards.md` gate matrix is green on all three packages: `gofmt -l .` empty,
  `go vet ./...` clean, `go build ./...` compiles, `go test ./...` passes, coverage
  ≥ 80% on each pure-logic package.
- Every story's five-part Intent is honored by tests, including the named invariants:
  disconnect-of-on-air clears and broadcasts nobody, unpaired-cannot-go-on-air,
  malformed/unknown-message decode never panics, late + duplicate frame dropped and never
  reordered, underrun/overrun reported and distinguishable, positive-drift → ratio > 1,
  clamp bounds honored, single-outlier does not swing the ratio.
- No goroutines/channels are introduced (session is single-threaded event-driven; audio
  is pure math), so `-race` stays advisory per standards Gate-matrix note.
- No production code is written by the planner; execution happens under TDD (RED →
  SCAFFOLD → GREEN) per the playbook.

## Out of scope

Verbatim from the intake §"Out of scope", plus the decoupling decisions above:

- Real WebRTC/DTLS-SRTP negotiation and media transport (offer/answer/ice are opaque
  relay only — D3).
- Opus encode/decode; the GStreamer pipeline; V4L2/UVC/UAC2 device I/O (configfs,
  `/dev/videoN`, ALSA).
- The actual WebSocket network server (the injected transport is a dumb sink, not built).
- HTTPS/TLS termination; mDNS/Tailscale integration.
- Any `Clock` interface or time-based hub behavior (D2 — no session timeouts this sprint).
- Wiring the jitter buffer into the drift controller (D7 — they stay decoupled).
- `libsamplerate`/ALSA `plug` real resampling — `internal/drift` computes the ratio only.
- Any goroutine/channel concurrency; anything needing hardware or a third-party dependency.

## User stories

Each story carries the five-part Intent (What's wanted · Constraints · Failure scenarios ·
Success scenarios · Connections). Failure-scenarios and Connections are mandatory.
Per-story task/validate/plan artifacts live in the story's `sN-NN-<slug>/` directory.

---

### S1-01 — Session hub (event-driven state machine)

Wave 1 (parallel). Slug: `s1-01-session-hub`. Package: `internal/session`.

**What's wanted.** A single-threaded, event-driven hub — `Process(event) -> []Effect`
(no goroutines, no real network) — that wires all four Sprint 0 packages into one runtime.
Constructed over an injected `pairing.Manager` and `switcher.Switcher`
(`session.New(pm, sw)`; `sw := switcher.New(pm)` since `pm.IsPaired` satisfies
`switcher.PairingOracle` with zero adapter). It drives a client through connect → `join`
(admit to switcher, register with pairing) → PIN-pair → (conditional) on-air → disconnect;
consumes raw signaling bytes via `signaling.Decode`; emits addressed outbound Effects; and
broadcasts on-air state via `signaling.OnAirStateFrom(switcher.OnAir())`, expanded into one
addressed Effect per connected client (D5).

**Constraints.**
- Single-threaded and dependency-injected: `Process` is one call in, one deterministic
  `[]Effect` out, no goroutines/channels/`time.Now()` (intake §Constraints; standards
  §Testing — `-race` stays advisory).
- Reuses Sprint 0 surfaces exactly; imports only `internal/core`, `internal/signaling`,
  `internal/pairing`, `internal/switcher` (+ stdlib). Modifies no Sprint 0 package.
- **Admission (D1).** `join` admits to `switcher.AddClient` immediately (visible) and
  `pairing.Register`s the client; pairing gates on-air only. An unpaired admitted client
  is refused on-air by `switcher.SetOnAir`'s existing `PairingOracle` gate — the hub never
  mutates on-air state directly (decision #5, decision #8).
- **On-air truth (D8/decision #8).** On-air state is a projection of `switcher.OnAir()`
  via `signaling.OnAirStateFrom` — the hub never keeps a second copy of the on-air fact.
- **Roster.** `switcher` exposes no enumeration (only `Has`), so the hub maintains its own
  ordered roster of connected clients purely to fan a broadcast out deterministically; the
  roster mirrors switcher membership (append on `join`, remove on disconnect).
- **Every error is a typed Effect, never a panic or a silent no-op.** Decode failures,
  duplicate join, unknown/unpaired switch-source target, and PIN submission for a
  never-joined client each yield an addressed `signaling.ErrorMessage` Effect to the
  requester (standards §Errors — never swallow).
- **Offer/Answer/ICE (D3).** Validated and relayed verbatim (opaque) as addressed Effects
  to the OTHER connected peers; SDP/ICE is never parsed or interpreted.
- **Disconnect (D4).** Removes the client from `switcher` (clearing on-air + broadcasting
  nobody if it was live) and from the roster; pairing state is retained (no deregister).

**Failure scenarios (NOT done if any hold).**
- A disconnect of the on-air client leaves the switcher/hub still reporting that client
  on-air, or omits the nobody-on-air broadcast, so remaining clients see a stale on-air
  participant (intake failure #1; the request's own worked example).
- An unpaired client's `join` or a `switch-source` naming it results in it becoming on-air
  — by bypassing `SetOnAir`'s gate or mutating on-air directly — instead of a typed refusal
  (intake failure #2; invariant: unpaired-cannot-go-on-air).
- A `join` for a `core.ClientID` already in the switcher's set is processed as a fresh
  silent admit instead of surfacing `switcher.ErrDuplicateClient` as a typed Effect
  (intake failure #3).
- A `switch-source` naming a target not connected (or not paired) silently no-ops instead
  of a typed error Effect (`ErrUnknownClient`/`ErrNotPaired`, never swallowed) (intake
  failure #4).
- A PIN submission for a `core.ClientID` that never sent `join` (never `Register`ed) is
  treated as success instead of surfacing `pairing.ErrUnknownClient` (intake failure #5).
- Malformed/undecodable inbound bytes (anything `signaling.Decode` returns a `*DecodeError`
  for) panic `Process` instead of producing a typed error Effect (intake failure #6;
  invariant: decode-never-panics).
- A disconnect for a `core.ClientID` that never joined panics or corrupts hub state instead
  of being a safe no-op (intake failure #7).

**Success scenarios (done when).**
- A client `join`s, submits the correct PIN within the limit; `pairing.IsPaired` is true
  and the client is in `switcher`'s set (intake success #1).
- Routing a paired, connected client on-air succeeds via `switcher.SetOnAir` and the
  returned Effects include `signaling.OnAirStateFrom(client, true)` addressed to every
  connected client (intake success #2).
- Disconnecting the on-air client removes it (`RemoveClient` clears on-air per its
  Sprint 0 contract) and the Effects include `signaling.OnAirStateFrom(<zero>, false)` to
  every
  remaining client (intake success #3; the worked example).
- An unpaired client's `join` is accepted (visible), but any on-air attempt is refused with
  a typed error Effect and never reaches on-air (intake success #4).
- A wrong PIN within the limit yields a typed `ErrIncorrectPIN` Effect; after the limit is
  exhausted the client is durably `StateRejected` and further PIN/on-air attempts are
  refused (intake success #5).
- Disconnecting a NOT-on-air client leaves on-air untouched and produces no on-air-state
  broadcast (intake success #6).

**Connections.** This hub is the FIRST place all four Sprint 0 packages are wired into one
runtime: `core.ClientID` (identity), `signaling` (`Decode`/`Encode`/`Validate` +
`OnAirStateFrom`), `pairing.Manager` (consumed as the switcher's `PairingOracle` with zero
adapter), and `switcher.Switcher` (client set + on-air + the `SetOnAir` pairing gate). A
bug in how the hub sequences these calls reintroduces exactly the failure modes Sprint 0's
invariant tests closed at the package level (never-two-on-air, unpaired-cannot-go-on-air) —
Sprint 1 must not re-derive or bypass them. A later hardware sprint replaces the injected
transport with a real WebSocket server (feeding real bytes into `Process` and writing its
Effects back out) and attaches the Pion/GStreamer endpoint that terminates the
offer/answer/ice this sprint only relays (intake §Connections).

---

### S1-02 — Jitter / reorder buffer

Wave 1 (parallel). Slug: `s1-02-jitter-buffer`. Package: `internal/jitter`.

**What's wanted.** A pure reordering buffer keyed on a per-frame sequence number (and
timestamp) that accepts out-of-order frames, drops late (seq < playout cursor) and
duplicate frames, emits frames in playout order, and reports underrun (pop with nothing
available) and overrun (push at capacity) as distinguishable, queryable conditions. Pure
logic over synthetic frame descriptors — no real audio, no Opus, no ALSA (README §4).

**Constraints.**
- Buffer capacity/depth is an injectable constructor parameter with a documented default
  (D6): `jitter.New(capacity)` (or a functional option) with a named default constant.
- `Push` returns a typed result distinguishing Accepted / DroppedLate / DroppedDuplicate /
  Overrun — a drop is never silent (standards §Errors; intake failures #8/#9/#10).
- `Pop` emits in playout order and reports underrun distinguishably from an ordinary empty
  read (intake failure #11): a `(Frame, false)` return plus a queryable underrun counter.
- The playout cursor never moves backward: once a seq has been emitted (or the cursor has
  passed it), a later frame with a lower seq is late and dropped, never reordered in
  (intake failure #8).
- Stdlib only; pure value logic, deterministic given a scripted input sequence — no
  wall-clock, no goroutines.

**Failure scenarios (NOT done if any hold).**
- A late frame (`seq < playout cursor`) is accepted and shifts the cursor backward instead
  of being dropped and never reordered (intake failure #8; the worked example).
- A duplicate frame (same seq already emitted or already buffered) is emitted a second time
  into playout order instead of being dropped (intake failure #9).
- A frame arriving at capacity is silently accepted (evicting arbitrary state) or panics/
  blocks instead of being reported as an overrun (intake failure #10).
- Pop with nothing available at/after the cursor is unobservable — a caller cannot
  distinguish "legitimately empty" from "underrun" because underrun is not reported/
  queryable (intake failure #11).

**Success scenarios (done when).**
- Frames pushed out of order (e.g. 3, 1, 2) are emitted in playout order (1, 2, 3) once the
  reorder window is satisfied (intake success #7).
- A late frame is dropped and reported as dropped — never emitted, never counted as
  reordered-in (intake success #8).
- A duplicate frame is dropped without a second emission (intake success #9).
- Overrun and underrun are each independently observable/queryable and distinguishable from
  ordinary pushes/pulls (intake success #10).

**Connections.** The buffer's emitted playout-order frames are the shape a later sprint's
real pipeline feeds toward Opus/ALSA; its underrun/overrun/depth readouts are naturally the
signal the drift controller (S1-03) would consume once the two are wired — but this sprint
they stay decoupled (D7): the buffer is unit-tested in isolation and the drift controller
takes a synthetic fill series. A later hardware sprint attaches real Opus decode (producing
the PCM frames this buffer reorders) and the UAC2/ALSA sink (intake §Connections).

---

### S1-03 — Clock-drift resampler control law

Wave 1 (parallel). Slug: `s1-03-drift-control`. Package: `internal/drift`.

**What's wanted.** The drift-tracking control law implied by README §4 ("the browser's
48 kHz and the USB host's 48 kHz are different crystals… an async resampler… with drift
tracking"): given a series of buffer fill-level samples over time — `(fillLevel, dt)` — a
controller computes the resample-ratio correction that keeps a UAC2 sink from under/over-
running. Pure math over numeric input — no `libsamplerate`, no ALSA, no real resampling.

**Constraints.**
- Control law (D8): a PROPORTIONAL controller with EMA smoothing, dead-band, and clamp.
  `fillError = (smoothed - target)/target`; `ratio = clamp(1 + Kp*fillError, 1-bound,
  1+bound)`. `Kp`, `target`, the clamp bound, the dead-band, and the EMA time constant are
  injectable with documented defaults; `dt` drives the EMA (the sole use of `dt`; D2 — no
  `Clock`).
- Sign convention (documented in code): sustained positive fill drift (buffer filling,
  `fill > target`) → `ratio > 1` (drain faster); sustained negative drift → `ratio < 1`.
- Noise damping: the EMA smooths fill before the proportional law so a single-sample
  outlier does not swing the ratio; a flat series (at target, or within the dead-band)
  yields exactly `ratio = 1.0` (intake failures #12/#13/#14).
- Stdlib `math` only; deterministic given a scripted fill series — no wall-clock, no real
  I/O jitter, no goroutines. Decoupled from S1-02 (D7).

**Failure scenarios (NOT done if any hold).**
- A fill series trending upward (sustained positive drift) yields `ratio ≤ 1.0` (or no
  correction), letting the buffer grow toward overrun (intake failure #12; the worked
  example).
- A fill series trending downward yields a ratio failing to go `< 1.0`, allowing repeated
  underruns (intake failure #13).
- A single noisy/outlier sample amid an otherwise flat series produces a large, immediate
  swing in the ratio — the loop reacting to single-sample noise as if it were sustained
  drift, which would produce audible artifacts in a real resampler (intake failure #14).

**Success scenarios (done when).**
- A fill series trending upward over the tracking window yields `ratio > 1.0` (intake
  success #11; the worked example).
- A fill series trending downward yields `ratio < 1.0` (intake success #12).
- A flat/stable series yields `ratio = 1.0` (or within the defined neutral/dead-band)
  (intake success #13).
- A single outlier amid an otherwise flat series does not swing the ratio outside the
  neutral band — the loop damps single-sample noise (intake success #14).

**Connections.** This is the pure-math half of README §4's audio pipeline. Its
`(fillLevel, dt)` input is the kind of signal S1-02's underrun/overrun/depth reporting
could produce once the two are wired — but that wiring is deferred (D7). A later hardware
sprint attaches the real UAC2/ALSA sink (whose real fill level this controller would track)
and `libsamplerate`/ALSA `plug` (the real resampling engine consuming this controller's
ratio), turning the proven control law into a live drift compensator (intake §Connections;
README §4 "a solved problem, but it must be in the design from day one").

## Story dependency graph

```
Sprint 0 base (already merged: go.mod + internal/core, internal/signaling,
               internal/config, internal/pairing, internal/switcher)
        │
        │  all three Sprint 1 packages import ONLY Sprint 0 packages
        ▼
Wave 1 (run in parallel — 3 independent stories, ~11 tasks):
  S1-01 session-hub  (imports core, signaling, pairing, switcher)
  S1-02 jitter-buffer (imports core only — synthetic frames)
  S1-03 drift-control (imports stdlib math only)
```

Edges:

- `Sprint0 → S1-01`, `Sprint0 → S1-02`, `Sprint0 → S1-03`: hard compile dependency on
  already-merged Sprint 0 packages. Because `go.mod` and every Sprint 0 package already
  exist on the base, there is NO serial foundation wave this sprint — all three stories
  build and enter RED concurrently in a single wave (natural wave width 3).
- `S1-01 ⇢ pairing + switcher` (compile): the hub imports both and wires
  `switcher.New(pairingManager)`. `pairing.Manager.IsPaired(core.ClientID) bool` already
  satisfies `switcher.PairingOracle`, so the coupling needs zero adapter code.
- `S1-02 ⇔ S1-03` (NONE this sprint): decoupled by D7. The buffer's readouts and the
  controller's fill input are the same signal in a later integration, but neither imports
  the other now; the controller takes a synthetic fill series. No compile edge → both stay
  parallel.
- `S1-01 ⇔ S1-02`, `S1-01 ⇔ S1-03` (NONE): the session hub has no compile dependency on the
  audio packages this sprint (the injected transport carries no audio) — all three are
  mutually independent and parallel.

---
sprint: 1
phase: planning
activity: intake
status: draft
---

# Intake — Sprint 1: hardware-free session hub + audio pure-logic pipeline

## Request (verbatim, restated)

Wire the Sprint 0 packages (`internal/core`, `internal/signaling`, `internal/config`,
`internal/pairing`, `internal/switcher`) into a hardware-free, unit-testable SESSION
layer plus the pure-logic half of the audio pipeline — deferring real WebRTC/
DTLS-SRTP, Opus codec, GStreamer, ALSA/UAC2 device I/O, the actual WebSocket network
server, HTTPS/TLS, and mDNS/Tailscale to a later hardware sprint. Three concrete
pieces: (a) a session hub (`internal/session`) — a single-threaded, event-driven state
machine over injected transport + clock interfaces that drives a client through
connect → join → PIN pairing → switcher admission → (conditionally) on-air, and
handles disconnect; (b) an audio jitter buffer (`internal/audio`) — a pure reordering
buffer for sequence-numbered/timestamped frames that drops late/duplicate frames and
reports underrun/overrun; (c) a clock-drift resampler control loop (`internal/audio`,
README §4) — pure math that turns buffer fill-level samples over time into a
resample-ratio correction. No third-party dependencies; reuses Sprint 0 packages as-is.

## What's wanted

**(a) Session hub (`internal/session`).** A single-threaded `process(event) ->
outbound effects` state machine, parameterized over an injected transport interface
and an injected clock interface (no goroutines, no real sockets, no wall-clock
sleeps) — so a unit test can drive a scripted sequence of events and assert on the
returned effects synchronously. Lifecycle: a client connects; sends a `join`
(`signaling.JoinMessage`, carrying `core.ClientID`); pairs by PIN through
`internal/pairing`'s `Manager`; on pairing outcome, the client's relationship to
`internal/switcher`'s client set is established (see Open Q1 for the exact timing);
a paired, connected client may be routed on-air only if `switcher.SetOnAir` accepts it
(decision #5 — already enforced by the `PairingOracle` gate, not re-litigated here);
the hub consumes the seven typed `internal/signaling` messages and produces typed
outbound messages, including broadcasting on-air state via
`signaling.OnAirStateFrom(switcher.OnAir())` (decision #8 — the switcher stays the
single source of truth, no duplicated on-air logic); a disconnect event removes the
client from the switcher (`RemoveClient`) and, if that client was on-air, clears
on-air and broadcasts nobody-on-air.

**(b) Audio jitter buffer (`internal/audio`).** A pure reordering buffer keyed on a
per-frame sequence number and/or timestamp: accepts frames arriving out of order,
drops frames that are late (behind the current playout cursor) or duplicate (already
played out or already buffered), emits frames in correct playout order, and reports
underrun (playout requested, nothing available) and overrun (push arrives at
capacity) as distinguishable, queryable conditions. Pure logic operating on synthetic
frame descriptors — no real audio samples, no Opus, no ALSA.

**(c) Clock-drift resampler control (`internal/audio`, README §4).** The
drift-tracking control loop implied by README §4 ("the browser's 48 kHz and the USB
host's 48 kHz are different crystals... an async resampler... with drift tracking"):
given a series of buffer fill-level samples over time, compute the resample-ratio
correction that would keep a UAC2 sink from under/over-running — sustained positive
(growing) drift should push the ratio above 1.0 to drain the buffer faster than it
fills; sustained negative (shrinking) drift should push it below 1.0. Pure math over
numeric input — no `libsamplerate`, no ALSA, no real resampling.

## Constraints

- Hardware-free, verbatim from the request: no real WebRTC/DTLS-SRTP negotiation, no
  Opus encode/decode, no GStreamer pipeline, no V4L2/UVC/UAC2 device I/O, no real
  WebSocket network server, no HTTPS/TLS termination, no mDNS/Tailscale. Everything
  in (a)/(b)/(c) must be exercisable by `go test` with no external process, no real
  socket, and no hardware — the same bar Sprint 0 set.
- No third-party dependencies added anywhere (stdlib only), consistent with Sprint
  0's Definition of Done.
- The session hub is single-threaded and dependency-injected: transport and clock are
  interfaces the hub calls through, never concrete goroutine/network/time.Now() code
  — this is what makes `process(event) -> outbound effects` deterministically
  unit-testable (one call in, one set of effects out, no timing races).
- Reuses the Sprint 0 public surfaces exactly as shipped, adding no changes to them
  this sprint unless a genuine gap blocks the hub (flagged as an Open question, not
  silently patched):
  - `internal/core.ClientID` — the identity threaded through every package.
  - `internal/signaling` — `Message`/`Envelope`, `Encode`/`Decode` (+ typed
    `DecodeError`), each message's `Validate()`, and the `OnAirStateFrom(id, live)`
    projection helper.
  - `internal/pairing.Manager` — `New`, `Register`, `SubmitPIN`, `State`, `Attempts`,
    `IsPaired`. Its `IsPaired(core.ClientID) bool` method already satisfies
    `switcher.PairingOracle` with no adapter needed.
  - `internal/switcher.Switcher` — `New(oracle)`, `AddClient`, `RemoveClient`, `Has`,
    `SetOnAir`, `OnAir`, and the `PairingOracle` interface it defines.
- Decision #5 (an unpaired client cannot be set on-air — enforced today by
  `switcher.SetOnAir`'s `PairingOracle` check) and decision #8 (`on-air-state` is a
  pure projection of `switcher.OnAir()`, never a second copy of on-air truth) are
  already resolved in Sprint 0 and bind Sprint 1's hub design; they are cited here as
  constraints, not reopened as questions.
- The jitter buffer (b) and clock-drift control (c) are pure logic / pure math:
  deterministic given a scripted sequence of synthetic inputs (frame descriptors,
  fill-level samples) — no wall-clock sleeps, no real audio codec, no floating
  non-determinism from real I/O jitter.
- Package boundary: (b) and (c) both live under `internal/audio`; (a) lives under
  `internal/session`. Whether (a) has any compile dependency on (b)/(c), or whether
  (b) and (c) depend on each other, is not fixed by the request and is called out
  below (Open Q7) rather than assumed.

## Failure scenarios (result does NOT count as done)

**Session hub (a)**
1. A disconnect of the currently on-air client leaves the switcher/hub still
   reporting that client on-air, or omits the nobody-on-air broadcast, so remaining
   clients' UI shows a stale on-air participant (violates the request's own worked
   example: "a disconnect of the on-air client clears on-air and broadcasts
   nobody-on-air").
2. An unpaired client's `join` (or a `switch-source` naming it) results in that
   client becoming on-air — whether by bypassing `switcher.SetOnAir`'s pairing gate
   or by the hub mutating on-air state directly — instead of being refused (violates
   the request's own worked example: "an unpaired join cannot reach on-air").
3. A `join` for a `core.ClientID` already present in the switcher's connected set is
   processed as a fresh, silent admit instead of surfacing
   `switcher.ErrDuplicateClient` as a typed outbound effect.
4. A `switch-source` naming a target not in the connected set (or not yet paired)
   silently no-ops instead of producing a typed error effect back to the requester
   (`switcher.ErrUnknownClient` / `ErrNotPaired`, never swallowed).
5. A PIN submission event for a `core.ClientID` that never sent `join` (never
   `Register`ed with `pairing.Manager`) is processed as if it succeeded instead of
   surfacing `pairing.ErrUnknownClient`.
6. Malformed/undecodable inbound signaling bytes (anything `signaling.Decode` would
   return a typed `*DecodeError` for) panic the hub's `process()` call instead of
   producing a typed error effect.
7. A disconnect event for a `core.ClientID` that never joined / was never admitted to
   the switcher panics or corrupts hub state instead of being a safe no-op.

**Audio jitter buffer (b)**
8. A late frame (`seq < playout cursor`) is accepted and shifts the playout cursor
   backward instead of being dropped and never re-ordered in (the request's own
   worked example).
9. A duplicate frame (same `seq` already emitted, or already sitting in the buffer)
   is emitted a second time into playout order instead of being dropped.
10. An incoming frame that arrives when the buffer is at capacity is silently
    accepted (evicting arbitrary state) or causes a panic/block, instead of being
    handled as a reported overrun.
11. Playout is requested with no frame available at/after the cursor and this is
    unobservable — a caller cannot distinguish "buffer legitimately empty" from
    "underrun," because underrun is not reported/queryable.

**Clock-drift resampler control (c)**
12. A fill-level sample series that trends upward (sustained positive drift, buffer
    growing) yields a resample ratio ≤ 1.0 (or no correction at all), letting the
    buffer keep growing toward overrun — contradicts the request's own worked
    example ("sustained positive fill drift yields a resample ratio >1 to drain the
    buffer").
13. A fill-level sample series that trends downward (sustained negative drift,
    buffer draining) fails to yield a ratio < 1.0, allowing repeated underruns.
14. A single noisy/outlier fill-level sample amid an otherwise flat series produces
    a large, immediate swing in the resample ratio — i.e. the control loop reacts to
    single-sample noise as if it were sustained drift, which would produce audible
    artifacts in a real resampler.

## Success scenarios (result DOES count as done)

**Session hub (a)**
1. A client connects, sends `join`, and submits the correct PIN within
   `pairing.Manager`'s configured attempt limit; `pairing.IsPaired` reports true for
   that client afterward, and the client is present in the switcher's connected set.
2. Routing a paired, connected client on-air succeeds via `switcher.SetOnAir`, and
   the hub's returned effects include `signaling.OnAirStateFrom(client, true)`.
3. Disconnecting the on-air client removes it from the switcher (`RemoveClient`
   clears on-air per its existing Sprint 0 contract) and the hub's effects include
   `signaling.OnAirStateFrom(<zero>, false)` — the request's own worked example.
4. An unpaired client's `join` is accepted (visible in whatever form Open Q1
   resolves), but any attempt to route it on-air is refused with a typed error
   effect (`switcher.ErrNotPaired` or equivalent) — it never reaches on-air. The
   request's own worked example: "an unpaired join cannot reach on-air."
5. A wrong PIN within the attempt limit yields a typed `pairing.ErrIncorrectPIN`
   effect; after the limit is exhausted the client is durably `StateRejected`, and
   further PIN submissions or on-air attempts from it are refused.
6. Disconnecting a client that is NOT on-air leaves on-air state untouched
   (`RemoveClient`'s existing contract) and produces no on-air-state broadcast.

**Audio jitter buffer (b)**
7. Frames pushed out of order (e.g. seq 3, 1, 2) are emitted in playout order (1, 2,
   3) once the reorder window is satisfied.
8. A late frame (`seq < playout cursor`) is dropped and reported as dropped — never
   emitted, never counted as reordered-in (the request's own worked example).
9. A duplicate frame is dropped without a second emission.
10. Overrun and underrun are each independently observable/queryable (e.g. distinct
    counters or reported conditions) and distinguishable from ordinary pushes/pulls.

**Clock-drift resampler control (c)**
11. A fill-level series trending upward over the tracking window yields a resample
    ratio > 1.0 — the request's own worked example.
12. A fill-level series trending downward yields a ratio < 1.0.
13. A flat/stable fill-level series yields a ratio of 1.0 (or within a defined
    neutral/dead-band).
14. A single noisy outlier amid an otherwise flat series does not swing the ratio
    outside the neutral band — the control loop damps single-sample noise.

## Connections (what a change here touches)

- The session hub is the FIRST place all four Sprint 0 packages are wired together
  into one runtime shape: `core.ClientID` (identity), `signaling` (`Decode`/`Encode`/
  `Validate` + the `OnAirStateFrom` projection), `pairing.Manager` (consumed as the
  switcher's `PairingOracle` with zero adapter code, since `IsPaired` already matches
  the interface), and `switcher.Switcher` (client set + on-air + the pairing gate
  built into `SetOnAir`). A bug in how the hub sequences these calls reintroduces
  exactly the failure modes Sprint 0's invariant tests already closed at the package
  level (never-two-on-air, unpaired-cannot-go-on-air) — Sprint 1 must not silently
  re-derive or bypass those guarantees.
- A later hardware sprint attaches: the real WebSocket network server (replaces the
  injected transport interface with real `net/http` + a WS library, feeding real
  inbound bytes into `hub.process()` and writing the hub's outbound effects back out
  over real sockets); real wall-clock time (replaces the injected clock, if any
  session-level timed behavior is confirmed in scope — see Open Q2); the Pion/
  GStreamer WebRTC endpoint (actually terminates the `offer`/`answer`/`ice-candidate`
  messages the hub can only route/relay this sprint — see Open Q3); HTTPS/TLS
  termination and mDNS/Tailscale discovery (the `config.HTTPSConfig`/`BindConfig`
  knobs Sprint 0 already models and validates — nothing new to build for those here).
- (b) and (c) are the pure-logic half of README §4's audio pipeline. A later hardware
  sprint attaches: real Opus decode (produces the PCM frames (b) would reorder in a
  real pipeline), the actual UAC2/ALSA sink (whose real fill level (c) would track,
  and whose real resample the ratio (c) computes would parameterize), and
  `libsamplerate`/ALSA `plug` (the real resampling engine consuming (c)'s output).
- (b)'s emitted playout-order frames are the shape a later sprint's real pipeline
  would feed toward Opus/ALSA; (c)'s fill-level input is naturally the kind of signal
  (b)'s underrun/overrun/depth reporting could produce. Whether (b) and (c) are
  wired together THIS sprint (c) consuming (b)'s live readouts, or stay fully
  decoupled with (c) taking a synthetic fill-level series, is an open question
  (below) — not assumed either way here.
- README §7 places "multi-client switcher UI; on-air states" in Phase 3 and audio
  end-to-end in Phase 1, both gated on hardware Sprint 0/1 pull the underlying logic
  of forward, ahead of the hardware. Sprint 1 is the layer that will let a later
  hardware sprint attach real transport/codec/ALSA around already-proven pure logic
  rather than discovering the state-machine bugs for the first time against real
  sockets and real audio hardware.

## In scope

- (a) `internal/session`: single-threaded, event-driven session hub over injected
  transport + clock interfaces (connect/join/PIN-pair/admit/on-air-route/disconnect).
- (b) `internal/audio`: pure jitter/reorder buffer for sequence-numbered/timestamped
  frames (drop late/duplicate, emit in playout order, report underrun/overrun).
- (c) `internal/audio`: pure clock-drift resampler control loop (fill-level samples
  in, resample-ratio correction out), per README §4.
- No third-party dependencies.

## Out of scope

Verbatim from the request:

- Real WebRTC/DTLS-SRTP negotiation and media transport.
- Opus encode/decode.
- The GStreamer pipeline.
- V4L2/UVC/UAC2 device I/O (configfs, `/dev/videoN`, ALSA).
- The actual WebSocket network server.
- HTTPS/TLS termination.
- mDNS/Tailscale integration.

## Open questions (blocking — hold for human/orchestrator confirmation before Stage 2 planning)

1. **Admission timing.** Is a joined-but-not-yet-paired client added to
   `switcher.AddClient` immediately at `join` — visible in the client set (e.g. as a
   greyed-out tile) and blocked from on-air solely by `SetOnAir`'s existing
   `PairingOracle` gate, matching Sprint 0's S0-04 resolution ("an unpaired client
   MAY be added/appear in the client set but CANNOT be set on-air") — or does the hub
   withhold `AddClient` until PIN pairing succeeds, matching the Sprint 1 request's
   literal wording ("on pair, is admitted to the switcher's client set")? These
   describe two different, differently-testable state machines and the request text
   and Sprint 0's resolved model currently disagree.
2. **Clock interface purpose.** `pairing.Manager`'s rejection is permanent with no
   timeout or retry-after-cooldown path (Sprint 0 has no lockout-duration concept).
   What concrete, time-dependent hub behavior in Sprint 1 actually needs the injected
   clock interface (a session/idle timeout? a heartbeat? something else)? If nothing
   in Sprint 1's scope needs real time yet, should the clock interface still be
   built as unused scaffolding, or deferred until a real timed behavior exists?
3. **Offer/Answer/ICE handling.** Since real WebRTC negotiation is out of scope, does
   the hub's `process(event)` validate and relay `offer`/`answer`/`ice-candidate` as
   opaque outbound effects (store-and-forward, no interpretation of the SDP/ICE
   payload), or are these three message types entirely unhandled/rejected this
   sprint, with routing deferred to the hardware sprint that terminates WebRTC?
4. **Pairing state on disconnect.** `pairing.Manager` has no remove/deregister
   method — its client map is never pruned. On disconnect, must the hub clear that
   client's pairing state (so a same-`ClientID` reconnect starts fresh), or is it
   acceptable/by design that a disconnected-then-reconnected client keeps its prior
   paired/rejected state for the life of the process?
5. **Broadcast fan-out shape.** When the hub emits an on-air-state effect, is it
   addressed to "all currently connected clients" as a hub-computed target list, or
   does the hub emit a single unaddressed "broadcast" effect and leave fan-out/
   addressing to the (not-yet-built) transport layer? This determines what shape the
   injected transport interface needs and what a unit test can assert about it.
6. **Jitter buffer depth.** Is the reorder window/buffer depth a fixed, injectable
   parameter Sprint 1 must expose (e.g. "N frames" or "N ms"), or is the exact depth
   an implementation detail left to the planner/task level with no constraint from
   intake?
7. **(b) → (c) wiring.** Does Sprint 1 require the clock-drift control loop (c) to
   actually consume live fill-level readouts from the jitter buffer (b) — i.e. a real
   integration point between the two — or are (b) and (c) independently unit-tested
   this sprint, with (c) taking a synthetic fill-level series with no dependency on
   (b) at all, deferring real wiring to a later sprint?
8. **Control-law choice for (c).** README §4 states the drift-compensation problem
   ("a solved problem... must be in the design from day one") but names no specific
   algorithm (proportional, PI, one-pole IIR/EMA, etc.). Is choosing the exact
   control law the planner's call to make and document, or does the human want to
   pin a specific approach now?

None of the above are resolved in this intake — resolving any of them here would be
the intake activity assuming scope to "keep moving," which is out of bounds for this
role. They should be closed by the human (or explicitly deferred with a stated
default) before the `planner` is dispatched for Sprint 1 Stage 2.

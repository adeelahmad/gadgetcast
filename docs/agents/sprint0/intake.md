---
sprint: 0
phase: planning
activity: intake
status: draft
---

# Intake — Sprint 0: hardware-free Go core (switcher, signaling codec, config, pairing)

## Request (verbatim, restated)

Build the foundational, unit-testable Go core of gadgetcast — the parts that need
neither the Raspberry Pi, WebRTC/GStreamer, nor the USB configfs gadget to exist or
run: (a) the source-switcher state machine, (b) the signaling message types + codec,
(c) the configuration model, and (d) the client pairing/PIN auth state machine.
Everything requiring hardware, a WebRTC/DTLS-SRTP stack, GStreamer, V4L2/UVC/UAC2
I/O, a real HTTPS server, mDNS, or Tailscale is explicitly deferred to a later
sprint.

## What's wanted

A pure-logic, dependency-light core — no Pi, no network sockets, no WebRTC/GStreamer
runtime, no USB gadget — that later sprints wire hardware and transport around,
covering four pieces: (1) a **source-switcher** state machine that tracks N
connected clients and enforces "exactly one on-air to the USB gadget at a time,"
with add-client, remove-client, set-on-air, and an on-air indicator query (README
§3.3, "Multiple browser clients can be connected simultaneously; exactly one stream
is 'on air' to the USB gadget at a time"); (2) a **signaling protocol codec** — typed
WebSocket message structs (join, offer, answer, ICE candidate, switch-source,
on-air-state, error) with JSON encode/decode and field validation, standing in for
the "WebSocket signaling channel" of README §3.2; (3) a **configuration model** that
parses and validates the security/bind knobs README §4 calls out — HTTPS on/off,
bind-interface (all vs. Tailscale-only), PIN/pairing requirement, and video/audio
format defaults (MJPEG/YUYV + 1080p30/720p30/480p30 per §3.1, Opus/48 kHz per §3.2);
and (4) a **client pairing/auth** state machine — per-client PIN pairing,
`pending → paired`/`pending → rejected`, with an attempt limit, standing in for
README §4's "PIN or QR pairing on first connect per client." All four must be fully
exercised by unit tests with no external process, no real socket, and no hardware.

## Constraints

- No hardware, no WebRTC/DTLS-SRTP, no GStreamer, no V4L2/UVC/UAC2 I/O, no real
  HTTPS server, no mDNS, no Tailscale, "anything needing hardware or heavy external
  deps" — verbatim out-of-scope list from the request.
- Exactly one on-air client at any instant is a hard invariant of the switcher
  (README §3.3) — it must hold across every add/remove/set-on-air sequence, not just
  in the steady state.
- HTTPS is a hard precondition elsewhere in the system ("`getUserMedia`/
  `getDisplayMedia` refuse to run in insecure contexts," README §3.2) — the config
  model must be able to represent and validate that requirement even though standing
  up a real HTTPS listener is out of scope here.
- Security baseline from README §4: "HTTPS everywhere, PIN or QR pairing on first
  connect per client, on-air indicator in the web UI, and an option to bind the web
  app to the Tailscale interface only." The config + pairing pieces exist to make
  that baseline checkable and enforceable in code, not just descriptive.
- Implementation language: the request specifies **Go**. README itself does not
  commit to a language for the whole project — §3.2 lists Pion (Go) as one of three
  candidate WebRTC endpoints ("Pion (Go) — clean, static binary...") alongside
  GStreamer `webrtcbin` and aiortc (Python), and §8's open questions explicitly say
  "GStreamer-only vs Pion+GStreamer split — prototype both in Phase 1 and pick on
  maintainability." Building Sprint 0 in Go does not by itself decide that later
  question; see Open questions.
- The exact seven signaling message types (join, offer, answer, ICE candidate,
  switch-source, on-air-state, error) come from the request, not a literal list in
  the README — §3.2 only says "a WebSocket signaling channel." join/offer/answer/ICE
  are standard WebRTC signaling vocabulary; switch-source/on-air-state map onto the
  switcher described in §3.3. Treated as in-scope and authoritative for Sprint 0
  unless the human says otherwise (see Open questions).

## Failure scenarios (result does NOT count as done)

1. Two clients call set-on-air concurrently/in sequence without an intervening
   remove → the system ends with the SECOND caller on-air and the first displaced;
   at no observable point are two clients simultaneously on-air (violates README
   §3.3's "exactly one stream is 'on air'").
2. set-on-air is called with a client ID not in the connected-client set → returns a
   typed error; on-air state is unchanged (no phantom on-air client is created).
3. The currently on-air client is removed → the on-air indicator transitions to "no
   one on-air," never continues to reference a removed client.
4. Decoding a signaling message with an unrecognized `type` field, truncated JSON, or
   a type/payload mismatch (e.g., an `offer` message missing its SDP field) returns a
   typed decode/validation error — it never panics and never silently substitutes a
   zero-value.
5. A config is parsed with HTTPS disabled (or enabled with no cert/key material
   supplied) → validation fails, because README states HTTPS is a hard requirement
   for `getUserMedia`/`getDisplayMedia` elsewhere in the system (§3.2) and the whole
   security posture assumes it (§4).
6. A config requests bind-interface `tailscale-only` combined with an unresolvable
   or all-interfaces bind target → validation rejects the contradiction rather than
   silently falling back to binding everywhere (the entire point of §4's
   Tailscale-only option is to NOT be reachable on the open LAN).
7. A client submits an incorrect PIN repeatedly past the configured attempt limit →
   further pairing attempts from that client are rejected outright (locked out),
   never allowed to retry indefinitely (unbounded retries would defeat the "PIN...
   pairing" security purpose stated in §4).
8. A client whose pairing state is `pending` (not yet `paired`) is passed to
   set-on-air or otherwise able to become the on-air source → this is a failure
   unless intake's open question about pairing/switcher coupling (below) is
   explicitly answered otherwise — an unpaired client reaching on-air is exactly the
   "unauthenticated gadgetcast on an open LAN" scenario §4 calls "a genuinely bad
   time."

## Success scenarios (result DOES count as done)

1. N clients are added with none on-air by default; set-on-air(A) makes A on-air and
   an on-air-state query reflects A, and only A.
2. set-on-air(B) while A is on-air atomically transitions the system so A is off-air
   and B is on-air, with the "exactly one on-air" invariant holding at every
   intermediate step (§3.3).
3. remove-client on a client that is NOT currently on-air leaves the on-air pointer
   untouched.
4. Every one of the seven message types (join, offer, answer, ICE candidate,
   switch-source, on-air-state, error) round-trips through encode → JSON →
   decode without loss, and each decodes to a validated, strongly-typed struct (not a
   loose `map[string]interface{}`).
5. A well-formed config — HTTPS on with cert/key set, bind=all (or bind=tailscale
   with a Tailscale-shaped target), PIN pairing required, video defaults MJPEG
   primary / YUYV fallback at 1080p30, 720p30, 480p30 (§3.1), audio default Opus at
   48 kHz (§3.2/§4) — parses and validates with zero errors.
6. A client that supplies the correct PIN within the attempt limit transitions
   `pending → paired` exactly once, and stays `paired` afterward.
7. A client that exhausts the attempt limit transitions `pending → rejected` and
   this rejection is durable/queryable (the state machine reports `rejected`, not
   silently `pending` again).

## Connections (what a change here touches)

- The switcher's on-air client ID is the pointer the (later) WebRTC/GStreamer
  pipeline reads to decide whose decoded audio/video gets written to the UVC/UAC2
  gadget devices (README §3.1, §3.3) — a bug here means the wrong participant's
  camera/mic reaches the host computer, which is the core promise of the whole
  project.
- The signaling codec (b) is the wire contract between browser clients and the Pi's
  (future) WebSocket server (§3.2); the switch-source and on-air-state message
  shapes must stay in lockstep with the switcher's (a) actual transition API, or a
  later sprint's transport layer will be built against a codec that doesn't match
  the state machine it's meant to drive.
- The config model's bind-interface and PIN-required fields are exactly what a later
  sprint's real HTTPS server, Tailscale integration, and mDNS advertisement (§4, §8)
  will read at startup — the device's whole trust boundary rests on this validation
  being correct before any real network code exists to test it against.
- The pairing/auth state machine (d) determines which clients are even eligible to
  appear in the switcher's connected-client set (a) — README §4's security intent
  (PIN pairing exists so an "unauthenticated gadgetcast on an open LAN" is not "a
  genuinely bad time") only holds if unpaired clients cannot reach the switcher or
  the on-air position; see Open question on this coupling.
- README §7 (MVP roadmap) assigns "multi-client switcher UI; on-air states" to Phase
  3 and pairing/PIN pairing flow to Phase 4 "product polish," both AFTER audio (Phase
  1) and video (Phase 2) are working end to end with hardware. Sprint 0 pulls the
  underlying logic of Phase 3/4 forward, ahead of the UI and the hardware it will
  eventually gate — the planner's dependency graph should make this pull-forward
  explicit so later hardware/UI stories build on this core rather than re-deriving
  it.

## In scope

- (a) Source-switcher state machine: add/remove client, set-on-air, on-air
  transitions, on-air indicator query (README §3.3).
- (b) Signaling message types + codec: join, offer, answer, ICE candidate,
  switch-source, on-air-state, error — JSON encode/decode + validation (README §3.2).
- (c) Configuration model: parse/validate HTTPS on/off, bind-interface (all vs.
  Tailscale-only), PIN/pairing requirement, video/audio format defaults (README §4,
  §3.1, §3.2).
- (d) Client pairing/auth state machine: per-client PIN pairing, `pending →
  paired`/`rejected`, attempt limits (README §4).

## Out of scope

- Actual WebRTC/DTLS-SRTP negotiation and media transport.
- GStreamer pipeline (decode/convert/encode).
- V4L2/UVC/UAC2 gadget I/O (configfs, `/dev/videoN`, ALSA).
- Real HTTPS server wiring (TLS termination, cert issuance/trust).
- mDNS advertisement.
- Tailscale integration.
- Anything else needing hardware or a heavy external dependency.

## Open questions (blocking — hold for human/orchestrator confirmation before Stage 2 planning)

1. **Language commitment.** The request specifies Go for this core. README §8 still
   lists "GStreamer-only vs Pion+GStreamer split — prototype both... and pick" as an
   open project-level question. Does building Sprint 0 in Go commit the whole
   project to Go/Pion for the WebRTC endpoint too, or is Sprint 0's language
   independent of that later decision (e.g., core logic in Go, WebRTC endpoint
   decided separately later)?
2. **Signaling message list is authoritative or illustrative?** README §3.2 never
   enumerates message types; the seven (join, offer, answer, ICE candidate,
   switch-source, on-air-state, error) come from the request. Is this the complete,
   fixed set for Sprint 0 (planner should not add/rename types without a new
   intake), or illustrative, leaving room to add types the planner discovers are
   needed (e.g., a "leave"/"heartbeat" message)?
3. **PIN attempt limit and lockout.** README §4 requires "PIN or QR pairing on first
   connect per client" but states no numeric attempt limit, no lockout duration, and
   no retry/reset path after rejection. What is the limit (a fixed constant, or a
   config field the operator sets), and can a `rejected` client ever retry (new PIN,
   timeout, manual admin reset), or is rejection terminal for that client's session?
4. **QR pairing in or out of scope.** README §4 says "PIN **or** QR pairing" — is
   QR-code pairing part of Sprint 0's auth state machine (as an alternate
   credential/transport modeled alongside PIN), or is Sprint 0 PIN-only with QR
   deferred to a later sprint?
5. **Pairing ↔ switcher coupling.** Is an unpaired (`pending`/`rejected`) client
   excluded from the switcher's connected-client set entirely (invisible to
   add-client/tile UI data), or can it be "connected" (visible, e.g. as a
   greyed-out tile) but merely blocked from becoming on-air? README §4 implies the
   former but does not state it for the switcher's data model, and this changes the
   switcher's (a) public API shape.
6. **Bind-interface vs. PIN interaction.** Does binding to `tailscale-only` ever make
   PIN pairing optional (since Tailscale itself is an auth boundary), or is PIN
   pairing always required regardless of bind interface? README §4 lists both as
   independent baseline mitigations without saying whether either can substitute for
   the other.
7. **Config source format.** README does not specify a config file format (YAML,
   JSON, env vars, flags) or a defaults source. Is choosing this format Sprint 0's
   call to make and document, or is there an existing convention the human wants
   followed?
8. **"On-air indicator" scope split between (a) and (b).** The request lists an
   "on-air indicator state" under the switcher (a) and a separate `on-air-state`
   signaling message under the codec (b). Is (a)'s indicator strictly the internal
   state-machine query (who is on-air, queried in-process), with the wire
   representation belonging entirely to (b)'s message type — or does Sprint 0 need
   (a) itself to know how to serialize/broadcast that state, duplicating part of
   (b)?

None of the above are resolved in this intake — resolving any of them here would be
the intake activity assuming scope to "keep moving," which is out of bounds for this
role. They should be closed by the human (or explicitly deferred with a stated
default) before the `planner` is dispatched for Sprint 0 Stage 2.

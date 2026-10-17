# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

gadgetcast turns a Raspberry Pi into a USB composite gadget (UVC webcam + UAC2 mic/speaker) whose media comes from WebRTC sessions served by the Pi to browsers on the LAN. `docs/design.md` (formerly the README) is the design doc (architecture, latency budget, USB bandwidth/MJPEG rationale, audio drift, security baseline, phased roadmap); code comments cite its sections as "README §N" (e.g. "README §3.2").

Current state: `cmd/gadgetcast` serves the REST/WebSocket API and the Studio UI (`web/`), creates the USB gadget (`gadgetcast gadget up`), and plays the on-air browser's mic into the UAC2 gadget and its camera or screen into the UVC gadget over WebRTC, and can send the host's speaker audio back to the on-air browser (video: H.264 or VP8 → MJPEG via GStreamer `uvcsink`). Go dependencies are `coder/websocket` and Pion WebRTC (Pion approved by Adeel 2026-10-17); add no others without asking.

## Commands

```sh
go build ./...
go vet ./...
gofmt -l .                                   # must print nothing
go test ./...                                # full suite
go test ./internal/session/                  # one package
go test ./internal/jitter/ -run TestName -v  # one test
go test -cover ./...                         # coverage (last reported ~93.5%)
```

`go.mod` declares `go 1.25.0` (x/crypto and Pion security updates need it; run with `GOTOOLCHAIN=auto` if your local Go is older). These four (`gofmt`, `vet`, `build`, `test`) are the gate matrix every story must pass.

## Architecture

All code is under `internal/`, one abstraction per package:

- `core` — `ClientID`, a comparable typed identity used as a map key everywhere.
- `signaling` — the closed set of WebSocket signaling messages (join, offer, answer, ice-candidate, switch-source, on-air-state, error). `Encode`/`Decode` use a `{type, payload}` JSON envelope with private `*Wire` structs (because `ClientID` has unexported fields); `Decode` returns `*DecodeError`; each message has `Validate()` wrapping `ErrValidation`.
- `config` — JSON config `Parse` → `applyDefaults` → `Validate` (HTTPS with cert+key is a hard precondition).
- `pairing` — per-client PIN pairing state machine with attempt limit and durable lockout. The expected PIN is injected, never hardcoded.
- `switcher` — connected-client set plus the single on-air selection. Depends on pairing only through the `PairingOracle` interface (`pairing.Manager` satisfies it directly).
- `session` — the integration surface: `Hub.Process(Event) []Effect` is a synchronous, side-effect-free event processor over pairing + switcher. Events are `Inbound` (raw frame bytes), `SubmitPIN`, `Disconnect`; effects are addressed `signaling.Message`s. SDP/ICE are relayed to every other roster client; a successful switch-source or the on-air client disconnecting broadcasts on-air state. Future transport (WS server, WebRTC) is meant to sit outside the hub, feed it events, and deliver its effects.
- `jitter` — sequence-number reordering buffer with underrun/overrun counters (stdlib only).
- `gadget` — configfs USB composite gadget (UVC MJPEG webcam + UAC2 mic/speaker) as an ordered `Plan` of filesystem ops; `Up`/`Down` run it (teardown is the reverse) through an `Executor`, so it is tested without a kernel. `gadgetcast gadget up|down` is run by `gadgetcast-gadget.service`.
- `media` — Pion WebRTC bridge: answers a paired client's `media-offer` (non-trickle) and forwards the on-air client's Opus RTP to an `AudioSink` and its H.264/VP8 RTP to `Options.Video` (with PLI keyframe requests); `GstSink` supervises a `gst-launch-1.0` pipeline playing audio into the UAC2 gadget, `NewGstVideoSink` one encoding MJPEG into the UVC gadget (restarting on a codec change). With `Options.Speaker`, `WriteSpeakerRTP` sends host audio (captured by `GstSource`) to the on-air client if its offer is sendrecv. `server.Options.Media` takes it as the `MediaBridge` interface.
- `drift` — proportional control law mapping buffer fill fraction to a clamped resample ratio, for audio clock-drift compensation (stdlib `math` only).

Dependency direction: `session` → {`core`, `signaling`, `pairing`, `switcher`}; `jitter` and `drift` import no internal packages. Avoid new cross-package coupling.

## Process conventions

Work is done in sprints via an "agentic-agile" TDD pipeline; per-sprint records live in `docs/agents/sprintN/` (`stories.md` holds the contracts and resolved decisions like D1–D8; each story dir has `plan.md`/`tasks.md`/`validate.md`/`output.md`). Read the relevant `stories.md` before changing a package's behavior. `docs/agents/memory.md` holds lessons learned; the Go-relevant ones:

- Strict red/green TDD: a RED test must fail on its assertion, not on a missing symbol (add a minimal compile shim). GREEN implements only what the tests require; don't widen the exported API.
- During parallel work, run tests scoped to your package; run `go test ./...` once everything has merged.
- The pipeline's gate scripts are Rust-oriented and give false results on Go (symbol-conflict checks, anti-cheat scan). Use `go vet`/`go build` and an independent check for skipped tests, `//nolint`, or build-tag exclusions instead.

`.agentic/`, `.transcripts/`, `.claude/`, and `docs/agents/standards.md` are gitignored pipeline artifacts.

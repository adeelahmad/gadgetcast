---
type: validate
story: S1-01
---

# Validate — S1-01 session-hub

PASS/FAIL rubric mirroring the task headings. Literal commands (PACKAGE-SCOPED per M-011),
literal expected output. Failure modes cite the rule each violates. Run from repo root.

## Pre-flight

- [ ] Worktree inherits merged base (M-010): `git reset --hard sprint1` before the first
      edit, so `internal/core`, `internal/signaling`, `internal/pairing`, `internal/switcher`
      are present.
- [ ] Sprint 0 packages build: `go build ./internal/core/ ./internal/signaling/ ./internal/pairing/ ./internal/switcher/` exits 0.
- [ ] `go test ./internal/session/` compiles.
- [ ] Scoped commands only (M-011): use `./internal/session/`, never `./...` — sibling
      packages `internal/jitter` / `internal/drift` are panic-stubbed mid-wave.
- [ ] `internal/session` imports only stdlib + `internal/core`, `internal/signaling`,
      `internal/pairing`, `internal/switcher`; no third-party:
      `go list -deps ./internal/session/ | grep -vE '^(internal|std|C$)' | grep '\.'` prints nothing
      outside `github.com/adeelahmad/gadgetcast/internal/*`.
- [ ] The hub wires `switcher.New(pairingManager)` with no adapter type (reuse note:
      `pairing.Manager.IsPaired` already satisfies `switcher.PairingOracle`).

## T1 — construction + join admission

| Check | Command | Expected | Fails rule |
|---|---|---|---|
| join admits + registers | `go test ./internal/session/ -run TestJoin_AdmitsClientAndRegistersPairing -v` | `--- PASS` | D1 / intake success #1 |
| duplicate join errors | `go test ./internal/session/ -run TestJoin_DuplicateClient_ErrorEffect -v` | `--- PASS` | intake failure #3 (surface ErrDuplicateClient) |
| unpaired join visible | `go test ./internal/session/ -run TestJoin_UnpairedClient_VisibleButNotPaired -v` | `--- PASS` | D1 (admit on join, gate on-air only) |

## T2 — PIN pairing events

| Check | Command | Expected | Fails rule |
|---|---|---|---|
| correct PIN pairs | `go test ./internal/session/ -run TestSubmitPIN_Correct_PairsClient -v` | `--- PASS` | intake success #1 |
| incorrect PIN error | `go test ./internal/session/ -run TestSubmitPIN_Incorrect_ErrorEffect -v` | `--- PASS` | intake success #5 |
| exhaust → durable reject | `go test ./internal/session/ -run TestSubmitPIN_ExhaustLimit_DurablyRejected -v` | `--- PASS` | intake success #5 (durable lockout) |
| PIN for never-joined | `go test ./internal/session/ -run TestSubmitPIN_UnknownClient_ErrorEffect -v` | `--- PASS` | intake failure #5 (pairing.ErrUnknownClient) |

## T3 — switch-source on-air + broadcast

| Check | Command | Expected | Fails rule |
|---|---|---|---|
| paired goes on-air + broadcast | `go test ./internal/session/ -run TestSwitchSource_PairedClient_GoesOnAirAndBroadcasts -v` | `--- PASS` | intake success #2 / decision #8 |
| unpaired never on-air | `go test ./internal/session/ -run TestSwitchSource_UnpairedTarget_ErrorEffect_NotOnAir -v` | `--- PASS` | intake failure #2 (invariant: unpaired-cannot-go-on-air) |
| unknown target errors | `go test ./internal/session/ -run TestSwitchSource_UnknownTarget_ErrorEffect -v` | `--- PASS` | intake failure #4 (switcher.ErrUnknownClient) |
| broadcast to all connected | `go test ./internal/session/ -run TestSwitchSource_BroadcastAddressedToAllConnected -v` | `--- PASS` | D5 (one addressed Effect per connected client) |

FAIL if an unpaired (`pending`/`rejected`) client can be routed on-air (intake failure #2;
README §4 "unauthenticated gadgetcast on an open LAN"), or if the hub mutates on-air state
outside `switcher.SetOnAir`/`RemoveClient` (decision #8 — on-air is a projection of
`switcher.OnAir()`).

## T4 — disconnect handling

| Check | Command | Expected | Fails rule |
|---|---|---|---|
| on-air disconnect → nobody | `go test ./internal/session/ -run TestDisconnect_OnAirClient_ClearsAndBroadcastsNobody -v` | `--- PASS` | intake failure #1 / success #3 (worked example) |
| non-on-air disconnect quiet | `go test ./internal/session/ -run TestDisconnect_NonOnAirClient_NoBroadcast -v` | `--- PASS` | intake success #6 |
| unknown disconnect no-op | `go test ./internal/session/ -run TestDisconnect_UnknownClient_SafeNoop -v` | `--- PASS` | intake failure #7 (safe no-op, no panic) |
| pairing retained | `go test ./internal/session/ -run TestDisconnect_PairingStateRetained -v` | `--- PASS` | D4 (pairing durable, no deregister) |

FAIL if a disconnect of the on-air client leaves a stale on-air participant or omits the
nobody-on-air broadcast (intake failure #1).

## T5 — decode robustness + opaque relay

| Check | Command | Expected | Fails rule |
|---|---|---|---|
| malformed bytes no panic | `go test ./internal/session/ -run TestInbound_MalformedBytes_ErrorEffectNoPanic -v` | `--- PASS` | intake failure #6 (invariant: decode-never-panics) |
| unknown type no panic | `go test ./internal/session/ -run TestInbound_UnknownType_ErrorEffectNoPanic -v` | `--- PASS` | intake failure #6 |
| unexpected server msg | `go test ./internal/session/ -run TestInbound_UnexpectedServerMessage_ErrorEffect -v` | `--- PASS` | standards §Errors (never swallow) |
| offer relayed opaquely | `go test ./internal/session/ -run TestInbound_OfferRelayedOpaquelyToPeers -v` | `--- PASS` | D3 (opaque passthrough to peers) |
| ice relayed opaquely | `go test ./internal/session/ -run TestInbound_IceCandidateRelayedOpaquely -v` | `--- PASS` | D3 |

FAIL if `Process` panics on any decode-table input (intake failure #6), or if the hub parses
/ interprets SDP or ICE payloads (D3 — opaque only).

## Final sign-off

- [ ] Every T1–T5 row PASS; `go test ./internal/session/` exits 0.
- [ ] `gofmt -l internal/session` empty; `go vet ./internal/session/` clean.
- [ ] `go test ./internal/session/ -cover` ≥ 80%.
- [ ] `internal/session` imports only stdlib + the four Sprint 0 packages named above; no
      third-party dependency; no Sprint 0 package modified.
- [ ] No goroutines/channels introduced (single-threaded event-driven hub) — `-race` stays
      advisory (standards Gate-matrix note).

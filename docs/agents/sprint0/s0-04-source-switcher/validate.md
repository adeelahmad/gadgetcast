---
type: validate
story: S0-04
---

# Validate — S0-04 source-switcher

PASS/FAIL rubric mirroring the task headings. Literal commands, literal expected output.
Failure modes cite the rule each violates. Run from repo root.

## Pre-flight

- [ ] S0-00 merged: `go build ./internal/core/` exits 0.
- [ ] `go test ./internal/switcher/` compiles.
- [ ] The switcher does NOT import pairing or signaling (dependency graph — logical
      edges only): `go list -deps ./internal/switcher/ | grep -E 'internal/(pairing|signaling)'`
      prints nothing. Imports are stdlib + `internal/core` only.

## T1 — client set

| Check | Command | Expected | Fails rule |
|---|---|---|---|
| add appears | `go test ./internal/switcher/ -run TestAddClient_AppearsInSet -v` | `--- PASS` | S0-04 (add-client) |
| duplicate add errors | `go test ./internal/switcher/ -run TestAddClient_Duplicate_Errors -v` | `--- PASS` | standards §Errors (typed error) |
| remove unknown errors | `go test ./internal/switcher/ -run TestRemoveClient_Unknown_Errors -v` | `--- PASS` | standards §Errors |

## T2 — set-on-air + never-two-on-air

| Check | Command | Expected | Fails rule |
|---|---|---|---|
| none by default | `go test ./internal/switcher/ -run TestNewSwitcher_NoneOnAirByDefault -v` | `--- PASS` | intake success #1 |
| set → only that client | `go test ./internal/switcher/ -run TestSetOnAir_MakesOnlyThatClientOnAir -v` | `--- PASS` | intake success #1 |
| switch displaces A | `go test ./internal/switcher/ -run TestSetOnAir_Switch_A_to_B_DisplacesA -v` | `--- PASS` | intake success #2 |
| exactly one across sequence | `go test ./internal/switcher/ -run TestOnAir_ExactlyOne_AcrossSequence -v` | `--- PASS` | README §3.3 / intake failure #1 (invariant: never-two-on-air) |

FAIL if any observable step leaves two clients on-air (README §3.3 hard invariant —
the wrong participant's camera/mic would reach the host).

## T3 — unknown-client + remove-on-air

| Check | Command | Expected | Fails rule |
|---|---|---|---|
| unknown set-on-air unchanged | `go test ./internal/switcher/ -run TestSetOnAir_UnknownClient_Errors_StateUnchanged -v` | `--- PASS` | intake failure #2 (no phantom on-air) |
| remove on-air → nobody | `go test ./internal/switcher/ -run TestRemoveClient_OnAir_TransitionsToNobody -v` | `--- PASS` | intake failure #3 (no dangling reference) |
| remove non-on-air untouched | `go test ./internal/switcher/ -run TestRemoveClient_NotOnAir_LeavesOnAirUntouched -v` | `--- PASS` | intake success #3 |

## T4 — pairing gate + source of truth

| Check | Command | Expected | Fails rule |
|---|---|---|---|
| unpaired cannot go on-air | `go test ./internal/switcher/ -run TestSetOnAir_UnpairedClient_Errors -v` | `--- PASS` | decision #5 / intake failure #8 (invariant: unpaired-cannot-go-on-air) |
| paired succeeds | `go test ./internal/switcher/ -run TestSetOnAir_PairedClient_Succeeds -v` | `--- PASS` | decision #5 |
| may connect, not on-air | `go test ./internal/switcher/ -run TestUnpairedClient_MayConnect_ButNotOnAir -v` | `--- PASS` | decision #5 (couples pairing → switcher) |
| OnAir() is source of truth | `go test ./internal/switcher/ -run TestOnAir_ReturnsSourceOfTruthShape -v` | `--- PASS` | decision #8 (on-air-state is a projection of OnAir()) |

FAIL if an unpaired (`pending`/`rejected`) client can be set on-air (intake failure #8 —
"unauthenticated gadgetcast on an open LAN", README §4), or if the switcher imports
`internal/pairing`/`internal/signaling` (breaks the parallel-wave decoupling).

## Final sign-off

- [ ] Every T1–T4 row PASS; `go test ./internal/switcher/` exits 0.
- [ ] `gofmt -l internal/switcher` empty; `go vet ./internal/switcher/` clean.
- [ ] `go test ./internal/switcher/ -cover` ≥ 80%.
- [ ] `internal/switcher` imports only stdlib + `internal/core`; the pairing dependency
      is satisfied by the injected `PairingOracle` stub in tests (dependency graph).

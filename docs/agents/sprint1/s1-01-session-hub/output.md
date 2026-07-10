---
type: output
story: S1-01
---

## S1-01 · attempt 1 · red-worker · 2026-07-10T00:19:22Z
status: ok

### Summary
Wrote all 20 failing tests for `internal/session` at their exact `internal/session/session_test.go::TestName` paths from plan.md (T1 join admission incl. duplicate + unpaired-visible; T2 SubmitPIN correct/incorrect/exhaust-limit/unknown; T3 switch-source on-air + exact per-roster broadcast fan-out; T4 disconnect incl. on-air-clears-broadcasts-nobody, non-on-air-no-broadcast, safe no-op, retained pairing; T5 decode-never-panics for malformed/unknown/unexpected-server plus opaque offer/ice relay). Every test builds `pm := pairing.New(pin, limit)`, `sw := switcher.New(pm)`, `h := session.New(pm, sw)`, drives `Event`s through `h.Process`, and asserts on the returned `[]Effect` and on `pm`/`sw` state. Added one `// agentic:shim` file `internal/session/shim.go` declaring `Hub`, `New`, the `Event` interface + `Inbound`/`SubmitPIN`/`Disconnect`, `Effect{To core.ClientID; Msg signaling.Message}`, and `(*Hub).Process`. `Process` has a deliberately wrong body: it mutates no injected state and returns a single bogus non-error effect, so effect-list assertions, error-effect type assertions, and pm/sw state assertions all FAIL BY ASSERTION — including the safe-no-op case (which would trivially pass on a `nil` return). Re-based onto latest sprint1 (`5f0fc61`, which advanced to include the sibling `internal/jitter` after the initial reset) so the diff touches only `internal/session`. `gofmt` clean, `go vet ./internal/session` clean, `go build ./...` OK; `go test ./internal/session` fails: 20 FAIL / 0 PASS, zero runtime panics.

### Result
| Check | Status | Detail |
|---|---|---|
| `session_test.go::TestJoin_AdmitsClientAndRegistersPairing` | FAIL | fails by assertion (sw.Has false, no-effects) |
| `session_test.go::TestJoin_DuplicateClient_ErrorEffect` | FAIL | fails by assertion (Msg not ErrorMessage) |
| `session_test.go::TestJoin_UnpairedClient_VisibleButNotPaired` | FAIL | fails by assertion (sw.Has false) |
| `session_test.go::TestSubmitPIN_Correct_PairsClient` | FAIL | fails by assertion (not paired) |
| `session_test.go::TestSubmitPIN_Incorrect_ErrorEffect` | FAIL | fails by assertion (Msg not ErrorMessage) |
| `session_test.go::TestSubmitPIN_ExhaustLimit_DurablyRejected` | FAIL | fails by assertion (state not Rejected) |
| `session_test.go::TestSubmitPIN_UnknownClient_ErrorEffect` | FAIL | fails by assertion (Msg not ErrorMessage) |
| `session_test.go::TestSwitchSource_PairedClient_GoesOnAirAndBroadcasts` | FAIL | fails by assertion (OnAir not set) |
| `session_test.go::TestSwitchSource_UnpairedTarget_ErrorEffect_NotOnAir` | FAIL | fails by assertion (Msg not ErrorMessage) |
| `session_test.go::TestSwitchSource_UnknownTarget_ErrorEffect` | FAIL | fails by assertion (Msg not ErrorMessage) |
| `session_test.go::TestSwitchSource_BroadcastAddressedToAllConnected` | FAIL | fails by assertion (1 effect, want 3) |
| `session_test.go::TestDisconnect_OnAirClient_ClearsAndBroadcastsNobody` | FAIL | fails by assertion (wrong effect list) |
| `session_test.go::TestDisconnect_NonOnAirClient_NoBroadcast` | FAIL | fails by assertion (OnAir not A,true) |
| `session_test.go::TestDisconnect_UnknownClient_SafeNoop` | FAIL | fails by assertion (got 1 effect, want 0) |
| `session_test.go::TestDisconnect_PairingStateRetained` | FAIL | fails by assertion (state not Paired) |
| `session_test.go::TestInbound_MalformedBytes_ErrorEffectNoPanic` | FAIL | fails by assertion (no ErrorMessage, no panic) |
| `session_test.go::TestInbound_UnknownType_ErrorEffectNoPanic` | FAIL | fails by assertion (no ErrorMessage, no panic) |
| `session_test.go::TestInbound_UnexpectedServerMessage_ErrorEffect` | FAIL | fails by assertion (no ErrorMessage) |
| `session_test.go::TestInbound_OfferRelayedOpaquelyToPeers` | FAIL | fails by assertion (1 effect, want 2 to B,C) |
| `session_test.go::TestInbound_IceCandidateRelayedOpaquely` | FAIL | fails by assertion (wrong effect to B) |

### Next
Scaffold the named symbols in `internal/session`: replace the `shim.go` `Process` wrong-body with `panic("SUB-AGENT-TODO: ...")` stubs (type `Hub`, `New`, `Event`/`Inbound`/`SubmitPIN`/`Disconnect`, `Effect`, `(*Hub).Process`), delete the `// agentic:shim` marker, then hand to GREEN to implement event routing over `pairing.Manager` + `switcher.Switcher` (ordered roster for broadcast/relay, opaque offer/ice fan-out to peers, ErrorMessage surfacing of the pairing/switcher sentinels, decode-never-panics).

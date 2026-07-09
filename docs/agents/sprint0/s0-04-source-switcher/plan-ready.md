---
type: plan-ready
story: S0-04
from_red_at: 2026-07-09T22:42:05Z
---

# plan-ready — S0-04 source-switcher

RED verified on sprint0 (gate-red-verify green). Boxes tick ONLY on GREEN pass.

- [x] `internal/switcher/switcher_test.go::TestAddClient_AppearsInSet` — input: fresh switcher; act: `AddClient("cam-1")`; assert: `cam-1` is a member.
- [x] `internal/switcher/switcher_test.go::TestAddClient_Duplicate_Errors` — input: switcher with `cam-1`; act: `AddClient("cam-1")` again; assert: typed error, set unchanged.
- [x] `internal/switcher/switcher_test.go::TestRemoveClient_Unknown_Errors` — input: fresh switcher; act: `RemoveClient("ghost")`; assert: typed error.
- [x] `internal/switcher/switcher_test.go::TestNewSwitcher_NoneOnAirByDefault` — input: switcher with N paired clients; act: `OnAir()`; assert: live=false, nobody on-air (intake success #1).
- [x] `internal/switcher/switcher_test.go::TestSetOnAir_MakesOnlyThatClientOnAir` — input: paired A,B; act: `SetOnAir(A)`; assert: `OnAir()` reports A and only A (intake success #1).
- [x] `internal/switcher/switcher_test.go::TestSetOnAir_Switch_A_to_B_DisplacesA` — input: A on-air, B paired; act: `SetOnAir(B)`; assert: `OnAir()` reports B, A is off-air (intake success #2).
- [x] `internal/switcher/switcher_test.go::TestOnAir_ExactlyOne_AcrossSequence` — input: paired A,B,C; act: a scripted add/set/switch/remove sequence; assert: after every step at most one client is on-air — never two (README §3.3; intake failure #1; invariant: never-two-on-air).
- [x] `internal/switcher/switcher_test.go::TestSetOnAir_UnknownClient_Errors_StateUnchanged` — input: A on-air; act: `SetOnAir("ghost")` (not in set); assert: typed error and `OnAir()` still reports A — no phantom on-air client (intake failure #2).
- [x] `internal/switcher/switcher_test.go::TestRemoveClient_OnAir_TransitionsToNobody` — input: A on-air; act: `RemoveClient(A)`; assert: `OnAir()` live=false, nobody on-air, no dangling reference to A (intake failure #3).
- [x] `internal/switcher/switcher_test.go::TestRemoveClient_NotOnAir_LeavesOnAirUntouched` — input: A on-air, B present; act: `RemoveClient(B)`; assert: `OnAir()` still reports A (intake success #3).
- [x] `internal/switcher/switcher_test.go::TestSetOnAir_UnpairedClient_Errors` — input: switcher with a stub oracle where U is unpaired, U added to the set; act: `SetOnAir(U)`; assert: typed error, on-air unchanged (decision #5; intake failure #8).
- [x] `internal/switcher/switcher_test.go::TestSetOnAir_PairedClient_Succeeds` — input: stub oracle where P is paired, P in the set; act: `SetOnAir(P)`; assert: `OnAir()` reports P, no error.
- [x] `internal/switcher/switcher_test.go::TestUnpairedClient_MayConnect_ButNotOnAir` — input: unpaired U; act: `AddClient(U)` then `SetOnAir(U)`; assert: `AddClient` succeeds (U is a member/visible) but `SetOnAir` errors (decision #5 coupling).
- [x] `internal/switcher/switcher_test.go::TestOnAir_ReturnsSourceOfTruthShape` — input: P on-air; act: `OnAir()`; assert: returns `(core.ClientID, bool)` == `(P, true)` — the exact shape `signaling.OnAirStateFrom` consumes, confirming one source of truth (decision #8).

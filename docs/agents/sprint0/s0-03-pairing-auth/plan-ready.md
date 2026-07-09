---
type: plan-ready
story: S0-03
from_red_at: 2026-07-09T22:45:14Z
---

# plan-ready — S0-03 pairing-auth

RED verified on sprint0 (gate-red-verify green). Boxes tick ONLY on GREEN pass.

- [x] `internal/pairing/pairing_test.go::TestNewClient_StartsPending` — input: a fresh client; act: query state; assert: state is `pending`, `IsPaired` false.
- [x] `internal/pairing/pairing_test.go::TestSubmitPIN_Correct_TransitionsToPaired` — input: expected PIN `"1234"`; act: `SubmitPIN(id,"1234")`; assert: state `paired`, no error (intake success #6).
- [x] `internal/pairing/pairing_test.go::TestSubmitPIN_Correct_IsIdempotent` — input: already-paired client; act: `SubmitPIN` correct PIN again; assert: stays `paired`, transition counted once (intake success #6 "exactly once").
- [x] `internal/pairing/pairing_test.go::TestSubmitPIN_UnknownClient_Errors` — input: an id not registered; act: `SubmitPIN`; assert: typed error, no state created.
- [x] `internal/pairing/pairing_test.go::TestSubmitPIN_Wrong_StaysPending_IncrementsAttempts` — input: limit 5, one wrong PIN; act: `SubmitPIN` wrong; assert: state still `pending`, attempts == 1.
- [x] `internal/pairing/pairing_test.go::TestSubmitPIN_ExhaustLimit_TransitionsToRejected` — input: limit 5, five wrong PINs; act: submit wrong ×5; assert: state `rejected` (intake success #7).
- [x] `internal/pairing/pairing_test.go::TestSubmitPIN_AfterRejected_FurtherAttemptsRejected` — input: a rejected client; act: `SubmitPIN` again (any PIN); assert: refused, state stays `rejected`, never reverts to `pending` (invariant: brute-force-lockout; intake failure #7).
- [x] `internal/pairing/pairing_test.go::TestSubmitPIN_CorrectAfterLockout_StillRejected` — input: a rejected client, then the correct PIN; act: `SubmitPIN` correct; assert: stays `rejected` — lockout is durable, a late-correct PIN does not re-pair (intake success #7).
- [x] `internal/pairing/pairing_test.go::TestIsPaired_TrueOnlyWhenPaired` — input: a paired client; act: `IsPaired(id)`; assert: true.
- [x] `internal/pairing/pairing_test.go::TestIsPaired_FalseWhenPending` — input: a pending client; act: `IsPaired(id)`; assert: false (decision #5 — unpaired ineligible).
- [x] `internal/pairing/pairing_test.go::TestIsPaired_FalseWhenRejected` — input: a rejected client; act: `IsPaired(id)`; assert: false.
- [x] `internal/pairing/pairing_test.go::TestAttemptLimit_Configurable` — input: limit configured to 3; act: submit wrong ×3; assert: `rejected` on the 3rd (limit honored, not hardwired to 5) (decision #3).
- [x] `internal/pairing/pairing_test.go::TestRejectedState_QueryableAfterRejection` — input: a rejected client; act: query state later; assert: still reports `rejected` (durable/queryable, not silently `pending`) (intake success #7).

---
type: validate
story: S0-03
---

# Validate — S0-03 pairing-auth

PASS/FAIL rubric mirroring the task headings. Literal commands, literal expected output.
Failure modes cite the rule each violates. Run from repo root.

## Pre-flight

- [ ] S0-00 merged: `go build ./internal/core/` exits 0.
- [ ] `go test ./internal/pairing/` compiles.
- [ ] No hardcoded PIN literal ships as a default in non-test source:
      `grep -rn 'PIN' internal/pairing/*.go | grep -v _test.go` shows no baked-in secret
      value (standards §Security — secrets come from config/env; tests inject the PIN).
- [ ] `internal/pairing` imports only stdlib + `internal/core`.

## T1 — pending → paired

| Check | Command | Expected | Fails rule |
|---|---|---|---|
| starts pending | `go test ./internal/pairing/ -run TestNewClient_StartsPending -v` | `--- PASS` | S0-03 constraint (default state) |
| correct PIN pairs | `go test ./internal/pairing/ -run TestSubmitPIN_Correct_TransitionsToPaired -v` | `--- PASS` | intake success #6 |
| idempotent pairing | `go test ./internal/pairing/ -run TestSubmitPIN_Correct_IsIdempotent -v` | `--- PASS` | intake success #6 (exactly once, stays paired) |
| unknown client errors | `go test ./internal/pairing/ -run TestSubmitPIN_UnknownClient_Errors -v` | `--- PASS` | standards §Errors (typed error) |

## T2 — attempt limit + durable lockout

| Check | Command | Expected | Fails rule |
|---|---|---|---|
| wrong stays pending | `go test ./internal/pairing/ -run TestSubmitPIN_Wrong_StaysPending_IncrementsAttempts -v` | `--- PASS` | S0-03 constraint (counter) |
| exhaust → rejected | `go test ./internal/pairing/ -run TestSubmitPIN_ExhaustLimit_TransitionsToRejected -v` | `--- PASS` | intake success #7 |
| lockout durable | `go test ./internal/pairing/ -run TestSubmitPIN_AfterRejected_FurtherAttemptsRejected -v` | `--- PASS` | intake failure #7 (invariant: brute-force-lockout) |
| late-correct still rejected | `go test ./internal/pairing/ -run TestSubmitPIN_CorrectAfterLockout_StillRejected -v` | `--- PASS` | intake success #7 (durable) |

FAIL if a rejected client can revert to `pending` or re-pair with a late-correct PIN
(intake failure #7 — defeats the pairing security purpose, README §4).

## T3 — IsPaired + configurable limit

| Check | Command | Expected | Fails rule |
|---|---|---|---|
| IsPaired true when paired | `go test ./internal/pairing/ -run TestIsPaired_TrueOnlyWhenPaired -v` | `--- PASS` | decision #5 (predicate the switcher uses) |
| IsPaired false when pending | `go test ./internal/pairing/ -run TestIsPaired_FalseWhenPending -v` | `--- PASS` | decision #5 |
| IsPaired false when rejected | `go test ./internal/pairing/ -run TestIsPaired_FalseWhenRejected -v` | `--- PASS` | decision #5 |
| limit configurable | `go test ./internal/pairing/ -run TestAttemptLimit_Configurable -v` | `--- PASS` | decision #3 (configurable, not hardwired) |
| rejection queryable | `go test ./internal/pairing/ -run TestRejectedState_QueryableAfterRejection -v` | `--- PASS` | intake success #7 (durable/queryable) |

## Final sign-off

- [ ] Every T1–T3 row PASS; `go test ./internal/pairing/` exits 0.
- [ ] `gofmt -l internal/pairing` empty; `go vet ./internal/pairing/` clean.
- [ ] `go test ./internal/pairing/ -cover` ≥ 80%.
- [ ] No hardcoded PIN in non-test source; imports only stdlib + `internal/core`.

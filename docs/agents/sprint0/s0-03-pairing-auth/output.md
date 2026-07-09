---
type: output
story: S0-03
---

## S0-03 · attempt 1 · red-worker · 2026-07-09T22:42:09Z
status: ok

### Summary
Wrote all 13 failing tests for package `internal/pairing` at their exact
`internal/pairing/pairing_test.go::TestName` paths, plus one `// agentic:shim`
file (`internal/pairing/shim.go`) with deliberately-wrong bodies so the package
compiles and every test FAILS BY ASSERTION rather than by a missing symbol. The
shim declares `State` (zero value `StateUnknown`, plus `StatePending`/`StatePaired`/
`StateRejected`), typed errors `ErrUnknownClient`/`ErrIncorrectPIN`/`ErrRejected`,
and `Manager` with `New(expectedPIN, limit)`, `Register`, `State`, `Attempts`,
`SubmitPIN`, `IsPaired`. The zero-value `StateUnknown` shim return makes every
state assertion fail while remaining honest under a correct implementation. The
expected PIN is injected by tests (`"1234"`) — no PIN is hardcoded in source.
`gofmt -l` clean, `go vet ./internal/pairing` clean, `go test ./internal/pairing`
FAILS with assertion messages only. Diff vs sprint0 touches only the pairing test
file and the marked shim.

### Result
| Check | Status | Detail |
|---|---|---|
| `pairing_test.go::TestNewClient_StartsPending` | FAIL | State()=unknown, want pending |
| `pairing_test.go::TestSubmitPIN_Correct_TransitionsToPaired` | FAIL | State()=unknown, want paired |
| `pairing_test.go::TestSubmitPIN_Correct_IsIdempotent` | FAIL | State()=unknown, want paired |
| `pairing_test.go::TestSubmitPIN_UnknownClient_Errors` | FAIL | err=nil, want ErrUnknownClient |
| `pairing_test.go::TestSubmitPIN_Wrong_StaysPending_IncrementsAttempts` | FAIL | err=nil want ErrIncorrectPIN; Attempts=0 want 1 |
| `pairing_test.go::TestSubmitPIN_ExhaustLimit_TransitionsToRejected` | FAIL | State()=unknown, want rejected |
| `pairing_test.go::TestSubmitPIN_AfterRejected_FurtherAttemptsRejected` | FAIL | err=nil want ErrRejected; state not durable |
| `pairing_test.go::TestSubmitPIN_CorrectAfterLockout_StillRejected` | FAIL | State()=unknown, want rejected (durable lockout) |
| `pairing_test.go::TestIsPaired_TrueOnlyWhenPaired` | FAIL | setup State()=unknown, want paired |
| `pairing_test.go::TestIsPaired_FalseWhenPending` | FAIL | setup State()=unknown, want pending |
| `pairing_test.go::TestIsPaired_FalseWhenRejected` | FAIL | setup State()=unknown, want rejected |
| `pairing_test.go::TestAttemptLimit_Configurable` | FAIL | State()=unknown, want pending then rejected |
| `pairing_test.go::TestRejectedState_QueryableAfterRejection` | FAIL | later State()=unknown, want rejected |
| `gofmt -l internal/pairing` | PASS | no output |
| `go vet ./internal/pairing` | PASS | exit 0 |

### Next
Scaffold the named symbols in `internal/pairing`: `Manager` with real per-client
state, `Register`, `State`, `Attempts`, `SubmitPIN` (correct→paired idempotently;
wrong→ErrIncorrectPIN + increment; at `limit`→rejected; rejected→ErrRejected and
durable, correct-after-lockout stays rejected), `IsPaired`, and the `State`
constants + typed errors — as `panic("SUB-AGENT-TODO: ...")` stubs, deleting shim.go.

## S0-03 · attempt 1 · scaffolder · 2026-07-09T22:47:37Z
status: ok
### Summary
Replaced the RED compile shim with the canonical production skeleton in
`internal/pairing/pairing.go`. Declared the lifecycle `type State` with its four
constants (`StateUnknown` zero-valued), the three sentinel errors
(`ErrUnknownClient`/`ErrIncorrectPIN`/`ErrRejected`), the `Manager` struct, the
`New(expectedPIN string, limit int) *Manager` constructor, and the five method
signatures the tests call. Every function/method body is a single
`panic("SUB-AGENT-TODO: …")`; type/const/var-sentinel declarations carry no body
and no PIN is hardcoded. Deleted `internal/pairing/shim.go` (no `// agentic:shim`
marker remains). Package formats, builds (exit 0), and vets clean, so every test
now fails by panic (absence) rather than by a wrong-body assertion.
### Scaffold
+ create State (+StateUnknown/StatePending/StatePaired/StateRejected) @ internal/pairing/pairing.go
+ create ErrUnknownClient/ErrIncorrectPIN/ErrRejected @ internal/pairing/pairing.go
+ create Manager @ internal/pairing/pairing.go
+ create New @ internal/pairing/pairing.go
+ create Register @ internal/pairing/pairing.go
+ create State (method) @ internal/pairing/pairing.go
+ create Attempts @ internal/pairing/pairing.go
+ create SubmitPIN @ internal/pairing/pairing.go
+ create IsPaired @ internal/pairing/pairing.go
- delete shim @ internal/pairing/shim.go
### Result
| Check | Status | Detail |
|---|---|---|
| stubs panic+TODO | PASS | 6 funcs/methods (New, Register, State, Attempts, SubmitPIN, IsPaired) |
| shim removed | PASS | no `// agentic:shim`; shim.go deleted |
| `gofmt -l internal/pairing` | PASS | no output |
| `go build ./internal/pairing` | PASS | exit 0 |
| `go vet ./internal/pairing` | PASS | exit 0 |
### Next
green: fill the stubs — construct Manager with a per-client state store; correct
PIN → StatePaired (idempotent, no attempt count); wrong → increment + ErrIncorrectPIN,
StateRejected at limit; rejected → ErrRejected and durable (correct-after-lockout
stays rejected); IsPaired true only when StatePaired.

## S0-03 · attempt 1 · green-worker · 2026-07-09T22:59:55Z
status: ok
### Summary
Filled the four panic-TODO bodies in `internal/pairing/pairing.go` with the minimum
needed to pass all 13 RED tests. Manager now holds a per-client `map[core.ClientID]*record`
(state + attempt count). `New` injects expectedPIN and limit (no hardcoded PIN). `SubmitPIN`
enforces: unknown → ErrUnknownClient (no state created); rejected → ErrRejected (durable);
paired → idempotent no-op (attempts untouched); correct-while-pending → StatePaired; wrong →
increment attempts, transition to StateRejected on reaching limit else ErrIncorrectPIN. The
durable-lockout invariant holds: a correct PIN after lockout returns ErrRejected and never
re-pairs. `IsPaired` is true only in StatePaired. Scoped gate matrix clean; coverage 95.7%.
### Result
| Check | Status | Detail |
|---|---|---|
| `pairing_test.go::TestNewClient_StartsPending` | PASS | green |
| `pairing_test.go::TestSubmitPIN_Correct_TransitionsToPaired` | PASS | green |
| `pairing_test.go::TestSubmitPIN_Correct_IsIdempotent` | PASS | green |
| `pairing_test.go::TestSubmitPIN_UnknownClient_Errors` | PASS | green |
| `pairing_test.go::TestSubmitPIN_Wrong_StaysPending_IncrementsAttempts` | PASS | green |
| `pairing_test.go::TestSubmitPIN_ExhaustLimit_TransitionsToRejected` | PASS | green |
| `pairing_test.go::TestSubmitPIN_AfterRejected_FurtherAttemptsRejected` | PASS | green |
| `pairing_test.go::TestSubmitPIN_CorrectAfterLockout_StillRejected` | PASS | green |
| `pairing_test.go::TestIsPaired_TrueOnlyWhenPaired` | PASS | green |
| `pairing_test.go::TestIsPaired_FalseWhenPending` | PASS | green |
| `pairing_test.go::TestIsPaired_FalseWhenRejected` | PASS | green |
| `pairing_test.go::TestAttemptLimit_Configurable` | PASS | green |
| `pairing_test.go::TestRejectedState_QueryableAfterRejection` | PASS | green |
| gofmt -l / go vet / go build | PASS | clean |
| coverage | PASS | 95.7% (>=80%) |
### Next
structural review — verify no orphan/parallel pairing implementation; then next task.

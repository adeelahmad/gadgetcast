---
type: output
story: S1-02
---

## S1-02 · attempt 1 · red-worker · 2026-07-10T00:41:00Z
status: ok
### Summary
Wrote all 12 failing tests for package `internal/jitter` at their exact `internal/jitter/jitter_test.go::TestName` paths, covering construction/default+injectable capacity, in-order push/pop and cursor advance, out-of-order reordering, late-frame `DroppedLate` (never reordered, cursor unchanged), buffered-duplicate `DroppedDuplicate`, already-played duplicate `DroppedLate`, pop-empty underrun `(zero,false)`+`Underruns()==1`, push-at-capacity `Overrun`+`Overruns()==1`, overrun-leaves-buffer-intact, and underrun/overrun distinguishability. One `// agentic:shim`-marked `internal/jitter/shim.go` declares `Buffer`, `New(capacity int) *Buffer`, `Frame{Seq int}`, `PushResult` with `OK`/`Overrun`/`DroppedLate`/`DroppedDuplicate`, `(*Buffer).Push`/`Pop`/`Len`/`Underruns`/`Overruns`, and `DefaultCapacity` — all with WRONG bodies (Push→OK, Pop→(zero,true), counters→0) so the package compiles and every test fails by assertion, not by a missing symbol. Drain loops are bounded so the always-true shim Pop cannot hang. gofmt applied; `go vet ./internal/jitter` clean; stdlib only (no internal/core needed — Frame references only Seq).
### Result
| Check | Status | Detail |
|---|---|---|
| `internal/jitter/jitter_test.go::TestNew_DefaultCapacity` | FAIL | push beyond default cap got OK, want Overrun |
| `internal/jitter/jitter_test.go::TestNew_InjectableCapacity` | FAIL | 9th push into cap-8 got OK, want Overrun |
| `internal/jitter/jitter_test.go::TestPushPop_InOrder_EmitsInOrder` | FAIL | pop seq 0, want 1 |
| `internal/jitter/jitter_test.go::TestPop_AdvancesCursor` | FAIL | first pop (0,true), want (1,true) |
| `internal/jitter/jitter_test.go::TestPush_OutOfOrder_EmitsInPlayoutOrder` | FAIL | playout pop (0,true), want (1,true) |
| `internal/jitter/jitter_test.go::TestPush_LateFrame_DroppedNotReordered` | FAIL | late push got OK, want DroppedLate |
| `internal/jitter/jitter_test.go::TestPush_DuplicateBuffered_Dropped` | FAIL | dup push got OK, want DroppedDuplicate |
| `internal/jitter/jitter_test.go::TestPush_DuplicateAlreadyPlayed_Dropped` | FAIL | already-played push got OK, want DroppedLate |
| `internal/jitter/jitter_test.go::TestPop_Empty_ReportsUnderrun` | FAIL | pop empty ok=true, want false |
| `internal/jitter/jitter_test.go::TestPush_AtCapacity_ReportsOverrun` | FAIL | push at cap got OK, want Overrun |
| `internal/jitter/jitter_test.go::TestPush_Overrun_LeavesBufferIntact` | FAIL | pop1 after overrun (0,true), want (1,true) |
| `internal/jitter/jitter_test.go::TestUnderrunOverrun_Distinguishable` | FAIL | Underruns()==0, want 1 |
### Next
scaffold the named symbols: `Buffer`, `New`, `Frame`, `PushResult` (`OK`/`Overrun`/`DroppedLate`/`DroppedDuplicate`), `(*Buffer).Push`/`Pop`/`Len`/`Underruns`/`Overruns`, `DefaultCapacity` — replace shim bodies with `panic("SUB-AGENT-TODO: ...")` and delete the `// agentic:shim` file.

## S1-02 · attempt 1 · scaffolder · 2026-07-10T00:20:04Z
status: ok
### Summary
Created the canonical `internal/jitter/jitter.go` declaring every production symbol the RED tests reference exactly once: `type Buffer`, `type Frame{Seq int}`, `type PushResult` with consts `OK`/`Overrun`/`DroppedLate`/`DroppedDuplicate`, `const DefaultCapacity = 64`, `New(capacity int) *Buffer`, and methods `(*Buffer).Push(Frame) PushResult`, `Pop() (Frame,bool)`, `Len() int`, `Underruns() int`, `Overruns() int`. Type and const declarations carry no body; every function/method body is a `panic("SUB-AGENT-TODO: ...")` describing the recipe GREEN will implement — no real logic. Deleted the RED `// agentic:shim` file `internal/jitter/shim.go`; no shim marker remains. Wrote the package-distinctive bare symbol names to `.agentic/scaffold-symbols` (`Buffer`, `PushResult`, `DefaultCapacity`, `Underruns`, `Overruns`, `DroppedLate`, `DroppedDuplicate`), excluding tree-wide collision-prone names per M-007 (`New` counts 3; `Push`/`Pop`/`Len`/`Frame`/`OK`/`Overrun` are generic and stay stubbed but out of the uniqueness list). Stdlib only. `gofmt -w`, `go build ./internal/jitter` (exit 0), and `go vet ./internal/jitter` all clean.
### Scaffold
+ create internal/jitter/jitter.go (Buffer, Frame, PushResult+consts, DefaultCapacity, New, Push, Pop, Len, Underruns, Overruns — all fn/method bodies panic+TODO)
- delete internal/jitter/shim.go (// agentic:shim removed)
+ create .agentic/scaffold-symbols (7 package-distinctive bare names)
+ create .agentic/task.env
### Result
| Check | Status | Detail |
|---|---|---|
| stubs panic+TODO | PASS | 6 fn/method bodies, all panic("SUB-AGENT-TODO: ...") |
| symbols unique tree-wide | PASS | 7 distinctive names, ctx-symbols count==1 each |
| no // agentic:shim remains | PASS | shim.go deleted |
| go build ./internal/jitter | PASS | exit 0 |
| go vet ./internal/jitter | PASS | clean |
### Next
green: fill the stubs with the least code that passes exactly this task's tests

## S1-02 · attempt 1 · green-worker · 2026-07-10T00:24:32Z
status: ok
### Summary
Filled the six panic-TODO bodies in the canonical `internal/jitter/jitter.go` with the least code that passes exactly the 12 RED tests — no new tests, no unforced refactor, no suppression. `Buffer` now carries `capacity`, a `cursor` (next-expected seq; 0 = nothing played), a `frames map[int]Frame` reorder store, and `underruns`/`overruns` counters. `New` clamps capacity ≤ 0 to `DefaultCapacity` and initializes the frame map. `Push` checks in the exact order the tests assert: `Seq < cursor` → `DroppedLate` (late frame never reordered, cursor never moves backward), already-buffered `Seq` → `DroppedDuplicate`, `len(frames) >= capacity` → increment `overruns` + `Overrun` (buffer left intact, nothing evicted), else buffer + `OK`. `Pop` emits the lowest buffered seq, deletes it, and advances `cursor = seq+1`; on empty it increments `underruns` and returns `(Frame{}, false)`. `Len`/`Underruns`/`Overruns` are direct reads. Stdlib only; touches only `internal/jitter/jitter.go`.
### Result
| Check | Status | Detail |
|---|---|---|
| `internal/jitter/jitter_test.go::TestNew_DefaultCapacity` | PASS | green |
| `internal/jitter/jitter_test.go::TestNew_InjectableCapacity` | PASS | green |
| `internal/jitter/jitter_test.go::TestPushPop_InOrder_EmitsInOrder` | PASS | green |
| `internal/jitter/jitter_test.go::TestPop_AdvancesCursor` | PASS | green |
| `internal/jitter/jitter_test.go::TestPush_OutOfOrder_EmitsInPlayoutOrder` | PASS | green |
| `internal/jitter/jitter_test.go::TestPush_LateFrame_DroppedNotReordered` | PASS | green |
| `internal/jitter/jitter_test.go::TestPush_DuplicateBuffered_Dropped` | PASS | green |
| `internal/jitter/jitter_test.go::TestPush_DuplicateAlreadyPlayed_Dropped` | PASS | green |
| `internal/jitter/jitter_test.go::TestPop_Empty_ReportsUnderrun` | PASS | green |
| `internal/jitter/jitter_test.go::TestPush_AtCapacity_ReportsOverrun` | PASS | green |
| `internal/jitter/jitter_test.go::TestPush_Overrun_LeavesBufferIntact` | PASS | green |
| `internal/jitter/jitter_test.go::TestUnderrunOverrun_Distinguishable` | PASS | green |
| `go test ./internal/jitter` | PASS | 12/12, coverage 96.4% |
| gofmt / go vet / go build (internal/jitter) | PASS | clean |
### Next
structural review — orphan/parallel-implementation scan over internal/jitter, then final gate.

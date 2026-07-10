---
type: validate
story: S1-02
---

# Validate — S1-02 jitter-buffer

PASS/FAIL rubric mirroring the task headings. Literal commands (PACKAGE-SCOPED per M-011),
literal expected output. Failure modes cite the rule each violates. Run from repo root.

## Pre-flight

- [ ] Worktree inherits merged base (M-010): `git reset --hard sprint1` before the first
      edit (so `internal/core` is present).
- [ ] `go build ./internal/core/` exits 0.
- [ ] `go test ./internal/jitter/` compiles.
- [ ] Scoped commands only (M-011): use `./internal/jitter/`, never `./...` — sibling
      packages `internal/session` / `internal/drift` are panic-stubbed mid-wave.
- [ ] `internal/jitter` imports only stdlib (+ at most `internal/core`); no third-party and
      no `internal/drift`: `go list -deps ./internal/jitter/ | grep internal/drift` prints
      nothing (D7 decoupling).

## T1 — construction, frame type, in-order push/pop

| Check | Command | Expected | Fails rule |
|---|---|---|---|
| default capacity | `go test ./internal/jitter/ -run TestNew_DefaultCapacity -v` | `--- PASS` | D6 (documented default) |
| injectable capacity | `go test ./internal/jitter/ -run TestNew_InjectableCapacity -v` | `--- PASS` | D6 (injectable param) |
| in-order emits in order | `go test ./internal/jitter/ -run TestPushPop_InOrder_EmitsInOrder -v` | `--- PASS` | intake success #7 baseline |
| pop advances cursor | `go test ./internal/jitter/ -run TestPop_AdvancesCursor -v` | `--- PASS` | README §4 (playout order) |

## T2 — reordering + late/duplicate drops

| Check | Command | Expected | Fails rule |
|---|---|---|---|
| out-of-order → playout order | `go test ./internal/jitter/ -run TestPush_OutOfOrder_EmitsInPlayoutOrder -v` | `--- PASS` | intake success #7 |
| late frame dropped | `go test ./internal/jitter/ -run TestPush_LateFrame_DroppedNotReordered -v` | `--- PASS` | intake failure #8 (cursor never moves back) |
| duplicate (buffered) dropped | `go test ./internal/jitter/ -run TestPush_DuplicateBuffered_Dropped -v` | `--- PASS` | intake failure #9 |
| duplicate (played) dropped | `go test ./internal/jitter/ -run TestPush_DuplicateAlreadyPlayed_Dropped -v` | `--- PASS` | intake failure #9 |

FAIL if a late frame (`seq < cursor`) is emitted or shifts the cursor backward (intake
failure #8), or if any duplicate seq is emitted twice (intake failure #9).

## T3 — underrun + overrun reporting

| Check | Command | Expected | Fails rule |
|---|---|---|---|
| pop-empty underrun | `go test ./internal/jitter/ -run TestPop_Empty_ReportsUnderrun -v` | `--- PASS` | intake failure #11 (underrun queryable) |
| push-at-capacity overrun | `go test ./internal/jitter/ -run TestPush_AtCapacity_ReportsOverrun -v` | `--- PASS` | intake failure #10 (overrun reported) |
| overrun leaves buffer intact | `go test ./internal/jitter/ -run TestPush_Overrun_LeavesBufferIntact -v` | `--- PASS` | intake failure #10 (no arbitrary eviction/corruption) |
| underrun/overrun distinct | `go test ./internal/jitter/ -run TestUnderrunOverrun_Distinguishable -v` | `--- PASS` | intake success #10 (independently observable) |

FAIL if pop-on-empty is indistinguishable from a legitimate read (intake failure #11), or if
a push at capacity silently evicts state, panics, or blocks instead of reporting overrun
(intake failure #10).

## Final sign-off

- [ ] Every T1–T3 row PASS; `go test ./internal/jitter/` exits 0.
- [ ] `gofmt -l internal/jitter` empty; `go vet ./internal/jitter/` clean.
- [ ] `go test ./internal/jitter/ -cover` ≥ 80%.
- [ ] `internal/jitter` imports only stdlib (+ at most `internal/core`); no third-party, no
      `internal/drift` (D7).
- [ ] No goroutines/channels introduced (pure value logic) — `-race` stays advisory.

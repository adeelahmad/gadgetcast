---
type: init
story: S1-02
---

# init — S1-02 (inbound comms, append-only)

## S1-02 · attempt 1 · red-worker · 2026-07-10T00:11:59Z

### Mandate
Write the FAILING tests for package `internal/jitter` per the story plan.md, across ALL its tasks.
Tests-only RED against minimal compile shims so every test FAILS BY ASSERTION (not a compile
error, not a pass).

### Scope (May / May Not)
- **May:** create `internal/jitter/jitter_test.go` with the tests at their EXACT `path::fn`; add ONE
  `// agentic:shim`-marked file `internal/jitter/shim.go` declaring the referenced symbols (type Buffer, New(capacity int) *Buffer, Frame{Seq ...}, PushResult constants (OK/Overrun/DroppedLate/DroppedDuplicate), (*Buffer).Push(Frame) PushResult, (*Buffer).Pop() (Frame,bool), Len()/Underruns()/Overruns() int, default-capacity const)
  with WRONG bodies so the package COMPILES and tests fail by assertion.
- **May Not:** implement real behavior; touch `go.mod` or any package outside `internal/jitter`;
  skip/ignore/weaken a test. Imports allowed: stdlib + internal/core.

### Inputs (absolute main-tree paths)
- plan (exact test bullets): /Users/adeelahmad/work/gadgetcast/docs/agents/sprint1/s1-02-jitter-buffer/plan.md
- tasks: /Users/adeelahmad/work/gadgetcast/docs/agents/sprint1/s1-02-jitter-buffer/tasks.md
- validate: /Users/adeelahmad/work/gadgetcast/docs/agents/sprint1/s1-02-jitter-buffer/validate.md
- Sprint 0 package surfaces you consume are on the base (read their .go for exact signatures).

### Acceptance
- Every test at its exact path::fn; `go test ./internal/jitter` FAILS with ASSERTION failures; `git diff`
  vs sprint1 touches only internal/jitter test file + the marked shim; gofmt clean; vet clean.

### Memory (apply these — from the Sprint 0 retrospective)
- M-010: the harness makes your worktree a commit BEHIND the tip — run `git reset --hard sprint1` FIRST.
- M-006: write `.agentic/task.env` with SCOPE_GLOBS QUOTED + subtree glob: `SCOPE_GLOBS="internal/jitter internal/jitter/*"`.
- M-002: keep RED failing on the ASSERTION not a missing symbol — that is why the shims exist.

## S1-02 · attempt 1 · scaffolder · 2026-07-10T00:17:43Z

### Mandate
Stub every production symbol the RED `internal/jitter` tests reference EXACTLY ONCE as
`panic("SUB-AGENT-TODO: ...")`; delete the `// agentic:shim` file; write bare symbol names to
`.agentic/scaffold-symbols`. Implement NO real body.

### Scope (May / May Not)
- **May:** create `internal/jitter/jitter.go` (real decls: `Buffer`, `Frame`, `PushResult`+consts,
  `DefaultCapacity`; method sigs with panic-TODO bodies); delete `internal/jitter/shim.go`.
- **May Not:** implement logic; add/modify tests; define a symbol twice; touch other packages.

### Inputs
- RED tests: `internal/jitter/jitter_test.go`; plan-ready: `docs/agents/sprint1/s1-02-jitter-buffer/plan-ready.md`

### Acceptance
- Each function/method defined once with panic-TODO (type/const decls have no body); no
  `// agentic:shim` remains; `go build ./internal/jitter` compiles; vet clean.

### Memory
- M-007: SCAFFOLD_SYMBOLS is a literal space-separated list of BARE names; EXCLUDE names that
  recur across packages or as type-vs-method (e.g. `New`, `Push`/`Pop` if the counter false-flags) —
  they stay stubbed but out of the uniqueness list. Set quoted or leave empty to use the file.
- M-010: `git reset --hard sprint1` first.

## S1-02 · attempt 1 · green-worker · 2026-07-10T00:21:36Z

### Mandate
Fill the panic-TODO bodies in `internal/jitter/jitter.go` with the BARE MINIMUM to pass the 12
RED tests: in-order + reordering emit, late-frame DroppedLate (cursor never moves backward),
duplicate DroppedDuplicate/DroppedLate, underrun on pop-empty, overrun at capacity leaving the
buffer intact, distinguishable counters. No new tests, no refactor, no suppression.

### Scope (May / May Not)
- **May:** edit `internal/jitter/jitter.go` bodies only.
- **May Not:** add/modify/weaken tests; touch other packages/go.mod; leave any panic-TODO.

### Inputs
- plan-ready (12 boxes): `docs/agents/sprint1/s1-02-jitter-buffer/plan-ready.md`; tests: `internal/jitter/jitter_test.go`; standards.md

### Acceptance
- `go test ./internal/jitter` passes all 12; regressed=0; gofmt/vet clean; diff only jitter.go; coverage ≥80%.

### Memory
- M-010 (git reset --hard sprint1 first); M-011/M-005 (write a PACKAGE-SCOPED fenced .agentic/gate-matrix.md
  and point STANDARDS_FILE at it — module-wide test would fail on still-stubbed sibling packages).

---
type: init
story: S1-03
---

# init — S1-03 (inbound comms, append-only)

## S1-03 · attempt 1 · red-worker · 2026-07-10T00:11:59Z

### Mandate
Write the FAILING tests for package `internal/drift` per the story plan.md, across ALL its tasks.
Tests-only RED against minimal compile shims so every test FAILS BY ASSERTION (not a compile
error, not a pass).

### Scope (May / May Not)
- **May:** create `internal/drift/drift_test.go` with the tests at their EXACT `path::fn`; add ONE
  `// agentic:shim`-marked file `internal/drift/shim.go` declaring the referenced symbols (type Controller, New(opts ...Option) *Controller + Option funcs WithKp/WithTarget/WithClamp, (*Controller).Update(fill, dt float64) float64 (ratio), documented default consts (Kp/target/clamp/deadband/tau))
  with WRONG bodies so the package COMPILES and tests fail by assertion.
- **May Not:** implement real behavior; touch `go.mod` or any package outside `internal/drift`;
  skip/ignore/weaken a test. Imports allowed: stdlib + (stdlib math only).

### Inputs (absolute main-tree paths)
- plan (exact test bullets): /Users/adeelahmad/work/gadgetcast/docs/agents/sprint1/s1-03-drift-control/plan.md
- tasks: /Users/adeelahmad/work/gadgetcast/docs/agents/sprint1/s1-03-drift-control/tasks.md
- validate: /Users/adeelahmad/work/gadgetcast/docs/agents/sprint1/s1-03-drift-control/validate.md
- Sprint 0 package surfaces you consume are on the base (read their .go for exact signatures).

### Acceptance
- Every test at its exact path::fn; `go test ./internal/drift` FAILS with ASSERTION failures; `git diff`
  vs sprint1 touches only internal/drift test file + the marked shim; gofmt clean; vet clean.

### Memory (apply these — from the Sprint 0 retrospective)
- M-010: the harness makes your worktree a commit BEHIND the tip — run `git reset --hard sprint1` FIRST.
- M-006: write `.agentic/task.env` with SCOPE_GLOBS QUOTED + subtree glob: `SCOPE_GLOBS="internal/drift internal/drift/*"`.
- M-002: keep RED failing on the ASSERTION not a missing symbol — that is why the shims exist.

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

## S1-03 · attempt 1 · scaffolder · 2026-07-10T00:25:31Z

### Mandate
Stub every production symbol the RED `internal/drift` tests reference EXACTLY ONCE as
`panic("SUB-AGENT-TODO: ...")`; delete the `// agentic:shim` file; write bare unique symbol
names to `.agentic/scaffold-symbols`. Implement NO real body.

### Scope (May / May Not)
- **May:** create `internal/drift/drift.go` — real decls `type Controller`, `type Option`,
  option funcs `WithKp`/`WithTarget`/`WithClamp`, default consts (DefaultKp/Target/Clamp/Deadband/Tau),
  `New(opts ...Option) *Controller`, and `(*Controller).Update(fill, dt float64) float64` with panic-TODO body;
  delete `internal/drift/shim.go`.
- **May Not:** implement logic; add/modify tests; define a symbol twice; touch other packages.

### Inputs
- RED tests: `internal/drift/drift_test.go`; plan-ready: `docs/agents/sprint1/s1-03-drift-control/plan-ready.md`

### Acceptance
- Each function/method once with panic-TODO (const/type decls no body); no `// agentic:shim` remains;
  `go build ./internal/drift` compiles; vet clean.

## S1-03 · attempt 1 · green-worker · 2026-07-10T00:28:59Z

### Mandate
Fill the `New`/`Update` panic-TODO bodies in `internal/drift/drift.go` with the BARE MINIMUM to
pass the 11 RED tests: defaults, flat/dead-band → ratio 1.0, sustained-up → ratio>1, sustained-down
→ ratio<1, larger-error→larger-correction, clamp ±bound exact, single-outlier stays neutral (EMA
damps), sustained-vs-transient contrast. EMA-smoothed proportional law with dead-band + clamp. No
new tests, no refactor, no suppression.

### Scope (May / May Not)
- **May:** edit `internal/drift/drift.go` bodies only.
- **May Not:** add/modify/weaken tests; touch other packages/go.mod; leave any panic-TODO.

### Inputs
- plan-ready (11 boxes): `docs/agents/sprint1/s1-03-drift-control/plan-ready.md`; tests: `internal/drift/drift_test.go`; standards.md

### Acceptance
- `go test ./internal/drift` passes all 11; regressed=0; gofmt/vet clean; diff only drift.go; coverage ≥80%.

### Memory
- M-010 (reset --hard sprint1 first); M-011/M-005 (PACKAGE-SCOPED fenced .agentic/gate-matrix.md + STANDARDS_FILE);
  STORY_DIR = MAIN tree; git add scoped to internal/drift.

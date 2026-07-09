---
type: init
story: S0-02
---

# init — S0-02 (inbound comms, append-only)

## S0-02 · attempt 1 · red-worker · 2026-07-09T22:34:43Z

### Mandate
Write the FAILING tests for package `internal/config` per the story plan. Tests-only RED against
minimal compile shims so every test FAILS BY ASSERTION (not a compile error, not a pass).

### Scope (May / May Not)
- **May:** create the test files (config_test.go, validate_test.go, defaults_test.go) at the EXACT `path::fn` in the story plan.md;
  add minimal shim(s) in one `// agentic:shim`-marked file (internal/config/shim.go) declaring the
  referenced production symbols (Config, Parse, Validate, defaults/constants) with WRONG bodies so tests compile + fail by assertion.
- **May Not:** implement real behavior; touch `go.mod`, `internal/core`, or any other package;
  skip/ignore/weaken a test. Import only `internal/core` + stdlib.

### Inputs (absolute main-tree paths — docs/agents is not in your worktree)
- plan (exact test bullets): /Users/adeelahmad/work/gadgetcast/docs/agents/sprint0/s0-02-config-model/plan.md
- tasks: /Users/adeelahmad/work/gadgetcast/docs/agents/sprint0/s0-02-config-model/tasks.md
- validate: /Users/adeelahmad/work/gadgetcast/docs/agents/sprint0/s0-02-config-model/validate.md
- `internal/core` (available on sprint0): ClientID, NewClientID, ErrInvalidClientID

### Acceptance
- Every test exists at its exact path::fn; `go test ./internal/config` FAILS with ASSERTION failures;
  `git diff` vs sprint0 touches only internal/config test files + the marked shim file; gofmt clean; vet clean.

### Memory
- M-002 (red-worker): keep RED failing on the ASSERTION not a missing symbol — that is why the shims exist.
- Harness note: write `.agentic/task.env` with SCOPE_GLOBS QUOTED and a subtree glob, e.g.
  `SCOPE_GLOBS="internal/config internal/config/*"`. Reset your worktree onto `sprint0` first (it now has internal/core).

## S0-02 · attempt 1 · scaffolder · 2026-07-09T22:42:38Z

### Mandate
Stub every production symbol the RED `internal/config` tests reference EXACTLY ONCE as
`panic("SUB-AGENT-TODO: ...")`; delete the `// agentic:shim` file; write bare symbol names
to `.agentic/scaffold-symbols`. Implement NO real body.

### Scope (May / May Not)
- **May:** create real production file(s) (e.g. `internal/config/config.go`, `validate.go`,
  `defaults.go`) with panic-TODO bodies; delete `internal/config/shim.go`.
- **May Not:** implement logic; add/modify tests; define a symbol twice; touch other packages.
  (Struct fields + default constants have no body — a `const`/struct decl is fine; the
  function/method bodies are the panic-TODO stubs.)

### Inputs
- RED tests: `internal/config/*_test.go`; plan-ready: `docs/agents/sprint0/s0-02-config-model/plan-ready.md`

### Acceptance
- Each function/method symbol defined once with panic-TODO; no `// agentic:shim` remains;
  `go build ./internal/config` compiles; vet clean.

### Memory
- M-007 (scaffolder): SCAFFOLD_SYMBOLS is a literal space-separated list of BARE names; set quoted or leave empty.

## S0-02 · attempt 1 · green-worker · 2026-07-09T22:49:41Z

### Mandate
Fill the panic-TODO bodies in `internal/config/{config.go,validate.go,defaults.go}` with the BARE
MINIMUM to pass the 16 RED tests (parse/malformed, HTTPS+bind validation, defaults). No new tests,
no refactor, no suppression.

### Scope (May / May Not)
- **May:** edit `internal/config/config.go`, `validate.go`, `defaults.go` bodies.
- **May Not:** add/modify/weaken tests; touch other packages/go.mod; leave any panic-TODO.

### Inputs
- plan-ready (16 boxes): `docs/agents/sprint0/s0-02-config-model/plan-ready.md`; tests: `internal/config/*_test.go`; standards.md

### Acceptance
- `go test ./internal/config` passes all 16; regressed=0; gofmt/vet clean; diff only within internal/config; coverage ≥80%.

### Memory
- M-003; M-005/M-006 (per-package scoped gate matrix; stdlib encoding/json only).

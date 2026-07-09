---
type: init
story: S0-03
---

# init — S0-03 (inbound comms, append-only)

## S0-03 · attempt 1 · red-worker · 2026-07-09T22:34:43Z

### Mandate
Write the FAILING tests for package `internal/pairing` per the story plan. Tests-only RED against
minimal compile shims so every test FAILS BY ASSERTION (not a compile error, not a pass).

### Scope (May / May Not)
- **May:** create the test files (pairing_test.go) at the EXACT `path::fn` in the story plan.md;
  add minimal shim(s) in one `// agentic:shim`-marked file (internal/pairing/shim.go) declaring the
  referenced production symbols (State (pending/paired/rejected), New, SubmitPIN, IsPaired, State query) with WRONG bodies so tests compile + fail by assertion.
- **May Not:** implement real behavior; touch `go.mod`, `internal/core`, or any other package;
  skip/ignore/weaken a test. Import only `internal/core` + stdlib.

### Inputs (absolute main-tree paths — docs/agents is not in your worktree)
- plan (exact test bullets): /Users/adeelahmad/work/gadgetcast/docs/agents/sprint0/s0-03-pairing-auth/plan.md
- tasks: /Users/adeelahmad/work/gadgetcast/docs/agents/sprint0/s0-03-pairing-auth/tasks.md
- validate: /Users/adeelahmad/work/gadgetcast/docs/agents/sprint0/s0-03-pairing-auth/validate.md
- `internal/core` (available on sprint0): ClientID, NewClientID, ErrInvalidClientID

### Acceptance
- Every test exists at its exact path::fn; `go test ./internal/pairing` FAILS with ASSERTION failures;
  `git diff` vs sprint0 touches only internal/pairing test files + the marked shim file; gofmt clean; vet clean.

### Memory
- M-002 (red-worker): keep RED failing on the ASSERTION not a missing symbol — that is why the shims exist.
- Harness note: write `.agentic/task.env` with SCOPE_GLOBS QUOTED and a subtree glob, e.g.
  `SCOPE_GLOBS="internal/pairing internal/pairing/*"`. Reset your worktree onto `sprint0` first (it now has internal/core).

## S0-03 · attempt 1 · scaffolder · 2026-07-09T22:45:14Z

### Mandate
Stub every production symbol the RED `internal/pairing` tests reference EXACTLY ONCE as
`panic("SUB-AGENT-TODO: ...")`; delete the `// agentic:shim` file; write bare symbol names to
`.agentic/scaffold-symbols`. Implement NO real body.

### Scope (May / May Not)
- **May:** create `internal/pairing/pairing.go` with real declarations — `type State` +
  `StateUnknown`(zero)/`StatePending`/`StatePaired`/`StateRejected`; sentinels
  `ErrUnknownClient`/`ErrIncorrectPIN`/`ErrRejected`; `type Manager` + `New(expectedPIN string, limit int)`,
  and method sigs `Register`, `State`, `Attempts`, `SubmitPIN`, `IsPaired` — every METHOD/FUNCTION
  body `panic("SUB-AGENT-TODO: ...")`. Delete `internal/pairing/shim.go`.
- **May Not:** implement logic; add/modify tests; define a symbol twice; touch other packages.

### Inputs
- RED tests: `internal/pairing/pairing_test.go`; plan-ready: `docs/agents/sprint0/s0-03-pairing-auth/plan-ready.md`

### Acceptance
- Each symbol once; function bodies panic-TODO (const/type/var sentinel decls have no body);
  no `// agentic:shim` remains; `go build ./internal/pairing` compiles; vet clean.

### Memory
- M-007 (scaffolder): SCAFFOLD_SYMBOLS is a literal space-separated list of BARE names; set quoted or leave empty.

## S0-03 · attempt 1 · green-worker · 2026-07-09T22:56:02Z

### Mandate
Fill the panic-TODO bodies in `internal/pairing/pairing.go` with the BARE MINIMUM to pass the 13 RED
tests: pending→paired on correct PIN (idempotent), unknown-client error, wrong-PIN increments while
pending, exhaust limit → durable `rejected` (a later correct PIN NEVER re-pairs), IsPaired true only
when paired, configurable limit. No new tests, no refactor, no suppression.

### Scope (May / May Not)
- **May:** edit `internal/pairing/pairing.go` bodies only.
- **May Not:** add/modify/weaken tests; touch other packages/go.mod; leave any panic-TODO; hardcode a PIN.

### Inputs
- plan-ready (13 boxes): `docs/agents/sprint0/s0-03-pairing-auth/plan-ready.md`; tests: `internal/pairing/pairing_test.go`; standards.md

### Acceptance
- `go test ./internal/pairing` passes all 13; regressed=0; gofmt/vet clean; diff only pairing.go;
  durable-lockout invariant holds; coverage ≥80%.

### Memory
- M-003; M-005/M-006 (per-package scoped gate matrix; stdlib only).

---
type: init
story: S0-04
---

# init — S0-04 (inbound comms, append-only)

## S0-04 · attempt 1 · red-worker · 2026-07-09T22:34:43Z

### Mandate
Write the FAILING tests for package `internal/switcher` per the story plan. Tests-only RED against
minimal compile shims so every test FAILS BY ASSERTION (not a compile error, not a pass).

### Scope (May / May Not)
- **May:** create the test files (switcher_test.go) at the EXACT `path::fn` in the story plan.md;
  add minimal shim(s) in one `// agentic:shim`-marked file (internal/switcher/shim.go) declaring the
  referenced production symbols (Switcher, New, AddClient, RemoveClient, SetOnAir, OnAir, PairingOracle interface, membership query) with WRONG bodies so tests compile + fail by assertion.
- **May Not:** implement real behavior; touch `go.mod`, `internal/core`, or any other package;
  skip/ignore/weaken a test. Import only `internal/core` + stdlib.

### Inputs (absolute main-tree paths — docs/agents is not in your worktree)
- plan (exact test bullets): /Users/adeelahmad/work/gadgetcast/docs/agents/sprint0/s0-04-source-switcher/plan.md
- tasks: /Users/adeelahmad/work/gadgetcast/docs/agents/sprint0/s0-04-source-switcher/tasks.md
- validate: /Users/adeelahmad/work/gadgetcast/docs/agents/sprint0/s0-04-source-switcher/validate.md
- `internal/core` (available on sprint0): ClientID, NewClientID, ErrInvalidClientID

### Acceptance
- Every test exists at its exact path::fn; `go test ./internal/switcher` FAILS with ASSERTION failures;
  `git diff` vs sprint0 touches only internal/switcher test files + the marked shim file; gofmt clean; vet clean.

### Memory
- M-002 (red-worker): keep RED failing on the ASSERTION not a missing symbol — that is why the shims exist.
- Harness note: write `.agentic/task.env` with SCOPE_GLOBS QUOTED and a subtree glob, e.g.
  `SCOPE_GLOBS="internal/switcher internal/switcher/*"`. Reset your worktree onto `sprint0` first (it now has internal/core).

## S0-04 · attempt 1 · scaffolder · 2026-07-09T22:42:05Z

### Mandate
Stub every production symbol the RED `internal/switcher` tests reference EXACTLY ONCE as
`panic("SUB-AGENT-TODO: ...")`; delete the `// agentic:shim` file; write the bare symbol
names to `.agentic/scaffold-symbols`. Implement NO real body.

### Scope (May / May Not)
- **May:** create `internal/switcher/switcher.go` with real signatures + panic-TODO bodies;
  delete `internal/switcher/shim.go`.
- **May Not:** implement logic; add/modify tests; define a symbol twice; touch other packages.

### Inputs
- RED tests: `internal/switcher/switcher_test.go`; plan-ready: `docs/agents/sprint0/s0-04-source-switcher/plan-ready.md`

### Acceptance
- Each symbol defined once; new bodies are panic-TODO (interface `PairingOracle` is a type decl,
  no body — fine); no `// agentic:shim` remains; `go build ./internal/switcher` compiles; vet clean.

### Memory
- M-007 (scaffolder): SCAFFOLD_SYMBOLS is a literal space-separated list of BARE names (e.g. `AddClient`,
  method `OnAir` not `Switcher.OnAir`). Set it quoted in task.env or leave empty to use the file.

## S0-04 · attempt 1 · green-worker · 2026-07-09T22:48:24Z

### Mandate
Fill the panic-TODO bodies in `internal/switcher/switcher.go` with the BARE MINIMUM to pass the 14
RED tests. No new tests, no unforced refactor, no suppression. Enforce the never-two-on-air invariant
and the pairing gate (unpaired cannot go on-air).

### Scope (May / May Not)
- **May:** edit `internal/switcher/switcher.go` bodies only.
- **May Not:** add/modify/weaken tests; touch other packages or go.mod; leave any panic-TODO;
  widen exported surface beyond the tests.

### Inputs
- plan-ready (14 boxes): `docs/agents/sprint0/s0-04-source-switcher/plan-ready.md`; tests: `internal/switcher/switcher_test.go`
- Standards (authoritative): /Users/adeelahmad/work/gadgetcast/docs/agents/standards.md

### Acceptance
- `go test ./internal/switcher` passes all 14; regressed=0; gofmt/vet clean; diff only switcher.go;
  coverage of internal/switcher ≥80%.

### Memory
- M-003 (green-worker): implement only what RED requires; don't widen exported API.
- M-005/M-006 (harness): GREEN gate reads the standards matrix from FENCED blocks; per-package gate
  matrix avoids the module-wide test failing on still-stubbed sibling packages.

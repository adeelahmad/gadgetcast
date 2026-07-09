---
type: init
story: S0-00
---

# init — S0-00 core-foundation (inbound comms, append-only)

## S0-00 · attempt 1 · red-worker · 2026-07-09T22:10:44Z

### Mandate
Write the FAILING tests for `internal/core` (the `ClientID` value type, task T2). Tests-only
RED. `go.mod` already exists in the base (T1 done as infra) — do NOT touch it.

### Scope (May / May Not)
- **May:** create `internal/core/client_id_test.go` with the 5 tests named in the story
  `plan.md`; add MINIMAL compile shims for `ClientID`, `NewClientID`, `String`, `IsValid`,
  `IsZero`, `ErrInvalidClientID` so the package COMPILES and every test FAILS BY ASSERTION.
  Put shims in a file marked `// agentic:shim` (e.g. `internal/core/client_id_shim.go`) so
  the scaffolder can find and replace them.
- **May Not:** implement real ClientID behavior (shims must make tests FAIL, not pass);
  touch `go.mod`; touch any package other than `internal/core`; skip/ignore/weaken a test.

### Inputs
- Test plan (5 bullets, exact `path::fn`): `docs/agents/sprint0/s0-00-core-foundation/plan.md`
- Tasks: `docs/agents/sprint0/s0-00-core-foundation/tasks.md`
- Validate rubric: `docs/agents/sprint0/s0-00-core-foundation/validate.md`

### Acceptance
- All 5 tests exist at their exact `internal/core/client_id_test.go::TestName`.
- `go test ./internal/core` FAILS, and each failure is an ASSERTION failure (not a compile
  error and not a passing test — a RED test that passes exercises nothing).
- `git diff` vs base touches only `internal/core/*_test.go` + the marked shim file.

### Memory
- M-001 (scaffolder/all): fresh Go module needs `go mod init` before RED — already handled at base.
- M-002 (red-worker): `go test ./...` is the suite; keep RED failing on the ASSERTION not a
  missing symbol — that is exactly why the minimal shims exist.

## S0-00 · attempt 1 · scaffolder · 2026-07-09T22:12:00Z

### Mandate
Stub every production symbol the RED `internal/core` tests reference EXACTLY ONCE as
`panic("SUB-AGENT-TODO: ...")`, delete the `// agentic:shim` file, write the stubbed
symbol list to `.agentic/scaffold-symbols`. Implement NO real body.

### Scope (May / May Not)
- **May:** create `internal/core/client_id.go` with real signatures + panic-TODO bodies for
  `ClientID`, `NewClientID`, `String`, `IsValid`, `IsZero`, `ErrInvalidClientID`; delete
  `internal/core/client_id_shim.go`.
- **May Not:** implement any real logic; add/modify tests; touch `go.mod` or other packages;
  define any symbol twice.

### Inputs
- RED tests (verified, on sprint0): `internal/core/client_id_test.go`
- plan-ready: `docs/agents/sprint0/s0-00-core-foundation/plan-ready.md`

### Acceptance
- Each referenced symbol defined EXACTLY once; every new body is `panic("SUB-AGENT-TODO: ...")`;
  no `// agentic:shim` file remains; `go build ./internal/core` compiles; `go vet` clean.
- `.agentic/scaffold-symbols` lists the stubbed symbols (one per line).

### Memory
- M-001 (scaffolder): module already initialized at base.

## S0-00 · attempt 1 · green-worker · 2026-07-09T22:15:00Z

### Mandate
Fill the four `SUB-AGENT-TODO` panic bodies in `internal/core/client_id.go` with the BARE
MINIMUM to pass exactly the 5 RED tests. No new tests, no unforced refactor, no suppression.

### Scope (May / May Not)
- **May:** edit `internal/core/client_id.go` bodies only (`NewClientID`, `String`, `IsValid`, `IsZero`).
- **May Not:** add/modify/weaken any test; touch other packages or `go.mod`; add exported surface
  beyond what the tests use; leave any `panic("SUB-AGENT-TODO")`.

### Inputs
- plan-ready (5 tests + their FAIL messages): `docs/agents/sprint0/s0-00-core-foundation/plan-ready.md`
- Tests (authoritative spec): `internal/core/client_id_test.go`

### Acceptance
- `go test ./internal/core` PASSES all 5; `regressed=0`; `gofmt -l` empty; `go vet` clean;
  diff touches only `internal/core/client_id.go`.

### Memory
- M-003 (green-worker): implement only what RED requires; don't widen exported API beyond the task's tests.

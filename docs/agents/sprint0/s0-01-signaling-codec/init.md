---
type: init
story: S0-01
---

# init — S0-01 (inbound comms, append-only)

## S0-01 · attempt 1 · red-worker · 2026-07-09T22:34:43Z

### Mandate
Write the FAILING tests for package `internal/signaling` per the story plan. Tests-only RED against
minimal compile shims so every test FAILS BY ASSERTION (not a compile error, not a pass).

### Scope (May / May Not)
- **May:** create the test files (codec_test.go, validate_test.go, onair_test.go) at the EXACT `path::fn` in the story plan.md;
  add minimal shim(s) in one `// agentic:shim`-marked file (internal/signaling/shim.go) declaring the
  referenced production symbols (MessageType, Envelope, the 7 message structs (Join/Offer/Answer/IceCandidate/SwitchSource/OnAirState/Error), Encode, Decode, DecodeError, Validate, OnAirStateFrom) with WRONG bodies so tests compile + fail by assertion.
- **May Not:** implement real behavior; touch `go.mod`, `internal/core`, or any other package;
  skip/ignore/weaken a test. Import only `internal/core` + stdlib.

### Inputs (absolute main-tree paths — docs/agents is not in your worktree)
- plan (exact test bullets): /Users/adeelahmad/work/gadgetcast/docs/agents/sprint0/s0-01-signaling-codec/plan.md
- tasks: /Users/adeelahmad/work/gadgetcast/docs/agents/sprint0/s0-01-signaling-codec/tasks.md
- validate: /Users/adeelahmad/work/gadgetcast/docs/agents/sprint0/s0-01-signaling-codec/validate.md
- `internal/core` (available on sprint0): ClientID, NewClientID, ErrInvalidClientID

### Acceptance
- Every test exists at its exact path::fn; `go test ./internal/signaling` FAILS with ASSERTION failures;
  `git diff` vs sprint0 touches only internal/signaling test files + the marked shim file; gofmt clean; vet clean.

### Memory
- M-002 (red-worker): keep RED failing on the ASSERTION not a missing symbol — that is why the shims exist.
- Harness note: write `.agentic/task.env` with SCOPE_GLOBS QUOTED and a subtree glob, e.g.
  `SCOPE_GLOBS="internal/signaling internal/signaling/*"`. Reset your worktree onto `sprint0` first (it now has internal/core).

## S0-01 · attempt 1 · scaffolder · 2026-07-09T22:43:07Z

### Mandate
Stub every production symbol the RED `internal/signaling` tests reference EXACTLY ONCE as
`panic("SUB-AGENT-TODO: ...")`; delete the `// agentic:shim` file; write bare symbol names
to `.agentic/scaffold-symbols`. Implement NO real body. NOTE: `core.ClientID` is a STRUCT —
tests build ids via `core.NewClientID`; `Decode` returns `(Message, error)` where `Message`
is an interface (keep that shape). Message structs + `MessageType` consts + the `DecodeError`
type are declarations (no panic body); the FUNCTION bodies (`Encode`, `Decode`, per-message
`Validate`, `OnAirStateFrom`) are the panic-TODO stubs.

### Scope (May / May Not)
- **May:** create real production files (e.g. `message.go`, `codec.go`, `validate.go`, `onair.go`)
  with panic-TODO function bodies; delete `internal/signaling/shim.go`.
- **May Not:** implement logic; add/modify tests; define a symbol twice; touch other packages.

### Inputs
- RED tests: `internal/signaling/*_test.go`; plan-ready: `docs/agents/sprint0/s0-01-signaling-codec/plan-ready.md`

### Acceptance
- Each symbol once; function bodies are panic-TODO; no `// agentic:shim` remains;
  `go build ./internal/signaling` compiles; vet clean.

### Memory
- M-007 (scaffolder): SCAFFOLD_SYMBOLS is a literal space-separated list of BARE names; set quoted or leave empty.

## S0-01 · attempt 1 · green-worker · 2026-07-09T22:50:14Z

### Mandate
Fill the panic-TODO bodies in `internal/signaling/{message.go,codec.go,validate.go,onair.go}` with
the BARE MINIMUM to pass the 21 RED tests (7-message round-trip, decode-never-panics + typed
DecodeError, per-message Validate, OnAirStateFrom projection). No new tests, no refactor, no suppression.

### Scope (May / May Not)
- **May:** edit the four `internal/signaling/*.go` production files' bodies.
- **May Not:** add/modify/weaken tests; touch other packages/go.mod; leave any panic-TODO.

### Inputs
- plan-ready (21 boxes): `docs/agents/sprint0/s0-01-signaling-codec/plan-ready.md`; tests: `internal/signaling/*_test.go`; standards.md

### Acceptance
- `go test ./internal/signaling` passes all 21; regressed=0; gofmt/vet clean; diff only within
  internal/signaling; unknown-type Decode returns typed *DecodeError and NEVER panics; coverage ≥80%.

### Memory
- M-003; M-005/M-006 (per-package scoped gate matrix; stdlib encoding/json only; Decode returns interface Message).

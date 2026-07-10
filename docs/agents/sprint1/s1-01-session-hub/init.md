---
type: init
story: S1-01
---

# init — S1-01 (inbound comms, append-only)

## S1-01 · attempt 1 · red-worker · 2026-07-10T00:11:59Z

### Mandate
Write the FAILING tests for package `internal/session` per the story plan.md, across ALL its tasks.
Tests-only RED against minimal compile shims so every test FAILS BY ASSERTION (not a compile
error, not a pass).

### Scope (May / May Not)
- **May:** create `internal/session/session_test.go` with the tests at their EXACT `path::fn`; add ONE
  `// agentic:shim`-marked file `internal/session/shim.go` declaring the referenced symbols (type Hub, New(pm *pairing.Manager, sw *switcher.Switcher) *Hub, Event interface + Inbound{From core.ClientID; Data []byte}/SubmitPIN{Client core.ClientID; PIN string}/Disconnect{Client core.ClientID}, Effect{To core.ClientID; Msg signaling.Message}, (*Hub).Process(Event) []Effect)
  with WRONG bodies so the package COMPILES and tests fail by assertion.
- **May Not:** implement real behavior; touch `go.mod` or any package outside `internal/session`;
  skip/ignore/weaken a test. Imports allowed: stdlib + internal/core internal/signaling internal/pairing internal/switcher.

### Inputs (absolute main-tree paths)
- plan (exact test bullets): /Users/adeelahmad/work/gadgetcast/docs/agents/sprint1/s1-01-session-hub/plan.md
- tasks: /Users/adeelahmad/work/gadgetcast/docs/agents/sprint1/s1-01-session-hub/tasks.md
- validate: /Users/adeelahmad/work/gadgetcast/docs/agents/sprint1/s1-01-session-hub/validate.md
- Sprint 0 package surfaces you consume are on the base (read their .go for exact signatures).

### Acceptance
- Every test at its exact path::fn; `go test ./internal/session` FAILS with ASSERTION failures; `git diff`
  vs sprint1 touches only internal/session test file + the marked shim; gofmt clean; vet clean.

### Memory (apply these — from the Sprint 0 retrospective)
- M-010: the harness makes your worktree a commit BEHIND the tip — run `git reset --hard sprint1` FIRST.
- M-006: write `.agentic/task.env` with SCOPE_GLOBS QUOTED + subtree glob: `SCOPE_GLOBS="internal/session internal/session/*"`.
- M-002: keep RED failing on the ASSERTION not a missing symbol — that is why the shims exist.

## S1-01 · attempt 1 · scaffolder · 2026-07-10T00:22:20Z

### Mandate
Stub every production symbol the RED `internal/session` tests reference EXACTLY ONCE as
`panic("SUB-AGENT-TODO: ...")`; delete the `// agentic:shim` file; write bare unique symbol
names to `.agentic/scaffold-symbols`. Implement NO real body.

### Scope (May / May Not)
- **May:** create `internal/session/session.go` — real decls `type Hub`, the `Event` interface +
  `Inbound`/`SubmitPIN`/`Disconnect` structs, `Effect{To core.ClientID; Msg signaling.Message}`,
  `New(pm *pairing.Manager, sw *switcher.Switcher) *Hub`, and `(*Hub).Process(Event) []Effect` with a
  panic-TODO body; delete `internal/session/shim.go`.
- **May Not:** implement logic; add/modify tests; define a symbol twice; touch other packages.

### Inputs
- RED tests: `internal/session/session_test.go`; plan-ready: `docs/agents/sprint1/s1-01-session-hub/plan-ready.md`

### Acceptance
- Each symbol once; `Process` (+ any constructor) body is panic-TODO (interface/struct decls have no body);
  no `// agentic:shim` remains; `go build ./internal/session` compiles; vet clean.

### Memory
- M-007: exclude `New` (recurs across packages) from the uniqueness list — still stubbed. Prefer unique names
  (`Hub`, `Effect`, `Inbound`, `SubmitPIN`, `Disconnect`, `Process`). Set SCAFFOLD_SYMBOLS quoted or empty.
- M-010: `git reset --hard sprint1` first.

---
type: sprint-plan
sprint: 1
stage: "2"
---

# Sprint plan — Sprint 1 (Stage 2)

Waves, dependency graph, parallelism, critical path, per-story plan pointers, and the
cross-cutting gate matrix. All three stories are Stage-2 complete: each has a
`s1-NN-<slug>/` dir with `tasks.md` + `validate.md` + `plan.md` (tests only), no
placeholders remaining.

## Sprint goal

Ship three pure-logic, stdlib-only packages — `internal/session` (event-driven hub),
`internal/jitter` (reorder buffer), `internal/drift` (drift control law) — that wire the
Sprint 0 core into a runtime shape and add the pure-math half of the audio pipeline, all
unit-testable with no hardware, no socket, no third-party dependency. See `stories.md` for
the full contract and the eight resolved decisions (D1–D8).

## Dependency graph

```
Sprint 0 base (merged): go.mod + internal/{core,signaling,config,pairing,switcher}
        │  every Sprint 1 package imports ONLY already-merged Sprint 0 packages
        ▼
Wave 1 (parallel — 3 independent stories):
  S1-01 session-hub   → internal/session  (imports core, signaling, pairing, switcher)
  S1-02 jitter-buffer → internal/jitter   (imports core)
  S1-03 drift-control → internal/drift     (imports stdlib math)
```

There is NO serial foundation wave: `go.mod` and all five Sprint 0 packages already exist
on the integration base, so the three new packages compile against them from the first RED
and have no compile edges among themselves. The natural wave width is **3** — three
independent packages, one story each, dispatched together.

## Waves

| Wave | Stories | Package | Tasks | Runs |
|---|---|---|---|---|
| 1 | S1-01, S1-02, S1-03 | `internal/session`, `internal/jitter`, `internal/drift` | 5 + 3 + 3 = 11 | fully parallel (3-wide) |

After Wave 1's three stories all pass GREEN and merge, run STRUCTURAL-REVIEW across the
wave, then the once-per-sprint FINAL-GATE.

## Parallelism table

| Story | Package | Depends on (compile) | Parallel-safe with | Tasks | Tests |
|---|---|---|---|---|---|
| S1-01 | `internal/session` | core, signaling, pairing, switcher (all merged) | S1-02, S1-03 | 5 | 20 |
| S1-02 | `internal/jitter` | core (merged) | S1-01, S1-03 | 3 | 12 |
| S1-03 | `internal/drift` | stdlib `math` only | S1-01, S1-02 | 3 | 11 |

Each story owns a distinct new package directory (`internal/session`, `internal/jitter`,
`internal/drift`) with no shared files, so parallel worktrees never collide. Under resource
pressure, serialize the three (one worktree at a time) — this is a SCHEDULE change, never a
relaxation of an invariant.

## Critical path

Every story is a single wave with no inter-story edge, so the critical path is the single
longest story, **S1-01 session-hub (5 tasks, 20 tests)** — the only story wiring four
Sprint 0 packages together and the one carrying the named runtime invariants
(disconnect-of-on-air clears + broadcasts nobody, unpaired-cannot-go-on-air,
decode-never-panics). S1-02 (3 tasks) and S1-03 (3 tasks) finish inside S1-01's span.
Sprint wall-clock ≈ the S1-01 RED→SCAFFOLD→GREEN chain; the audio stories add no serial
length.

## Per-story plan pointers

- S1-01 → `s1-01-session-hub/plan.md` (tests only) · `tasks.md` · `validate.md`
- S1-02 → `s1-02-jitter-buffer/plan.md` (tests only) · `tasks.md` · `validate.md`
- S1-03 → `s1-03-drift-control/plan.md` (tests only) · `tasks.md` · `validate.md`

Test totals: S1-01 = 20, S1-02 = 12, S1-03 = 11 — **43 tests** across the sprint.

## Execution notes (memory-derived — bind the dispatch)

- **M-011 (package-scoped commands).** During this parallel wave, every worker scopes its
  test/gate commands to its OWN package (`go test ./internal/session/`, `./internal/jitter/`,
  `./internal/drift/`) — NOT module-wide `go test ./...`. Sibling packages are panic-stubbed
  mid-wave and would fail the whole-module run even though your package is done. All three
  `validate.md` files are already package-scoped. Run the module-wide `./...` sweep only at
  FINAL-GATE after the three stories merge.
- **M-010 (worktree base ref).** The harness creates each worktree a commit BEHIND the
  integration branch tip. Before the first edit in a fresh worktree, `git reset --hard
  <integration-branch>` (e.g. `sprint1`) so already-merged prior work is inherited. Recorded
  in each `validate.md` Pre-flight and to be echoed in the `init.md` dispatch block.
- **M-005 (fenced gate matrix).** `standards.md` keeps the Go gate commands in a FENCED
  block; the GREEN/FINAL gates parse fences, not markdown tables. Do not let the matrix
  below drift from `standards.md`.
- **M-002 / M-003 (RED/GREEN discipline).** RED fails by ASSERTION against minimal shims
  (not a missing symbol); GREEN implements only what the RED assertions require and never
  widens the exported API beyond the task's tests.

## Cross-cutting gates

The authoritative gate matrix is `docs/agents/standards.md` (Sprint 1 reaffirmation). The
GREEN and FINAL gates run the FENCED command block below, copied verbatim from
`standards.md` — do not let the two drift apart.

| Gate | Command | Pass condition | Status |
|---|---|---|---|
| **fmt** | `gofmt -l .` | Empty stdout (no file listed) | HARD gate |
| **vet** | `go vet ./...` | Exit code 0, no diagnostics | HARD gate |
| **typecheck** | `go build ./...` | Exit code 0 (the Go compiler is the type checker) | HARD gate |
| **test** | `go test ./...` | Exit code 0, all tests pass | HARD gate |
| **coverage** | `go test ./... -coverprofile=coverage.out` then `go tool cover -func=coverage.out \| tail -1` | Total ≥ **80%** on the pure-logic packages (`internal/session`, `internal/jitter`, `internal/drift`) | HARD gate |
| **audit** | `go vet ./...` + `govulncheck ./...` if present | `go vet` clean; `govulncheck` **N/A** (not installed) | HARD (vet) / N-A |
| **lint (extended)** | `golangci-lint run` | not installed — OPTIONAL, never blocks | OPTIONAL |
| **lint (static)** | `staticcheck ./...` | not installed — OPTIONAL | OPTIONAL |
| **security (extended)** | `gosec ./...` | not installed — OPTIONAL | OPTIONAL |

Machine-readable gate commands (the GREEN/FINAL run, verbatim from `standards.md`; at
FINAL-GATE these run module-wide, and per M-011 each mid-wave worker scopes the same
commands to its own package):

```bash
gofmt -l .
go vet ./...
go build ./...
go test ./...
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out
```

Notes:

- `-race` stays ADVISORY: no Sprint 1 package spawns goroutines/channels (session is
  single-threaded event-driven; audio is pure math). The promotion trigger is unchanged —
  the instant any package spawns goroutines/channels, `-race` becomes a HARD gate
  (`standards.md` Gate-matrix note).
- Stdlib-only holds: the three new packages import only Sprint 0 packages + stdlib, adding
  no `require` block to `go.mod` and no `go.sum` (standards §Dependencies).

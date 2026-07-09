---
type: sprint-plan
sprint: 0
stage: "2"
---

# Sprint 0 plan — waves, parallelism, gates

Stage-2 plan for the hardware-free Go core. Authored by the `planner`; execution runs
under TDD (RED → SCAFFOLD → GREEN → STRUCTURAL-REVIEW → FINAL-GATE). No production code
here. All commands assume repo root `/Users/adeelahmad/work/gadgetcast`.

## Sprint goal

Ship a stdlib-only Go core (module `github.com/adeelahmad/gadgetcast`, go 1.22) with
five packages — `internal/core`, `internal/signaling`, `internal/config`,
`internal/pairing`, `internal/switcher` — enforcing: exactly-one-on-air, a seven-message
typed signaling codec that never panics on bad input, a config model that makes the
HTTPS/bind/PIN posture checkable, and a PIN pairing machine with durable lockout that
gates the switcher. Green means the standards gate matrix passes and the four invariant
tests hold.

## Waves

**Wave 1 — foundation (serial, merged before Wave 2).**

- `S0-00` core-foundation. Creates the single `go.mod` (`go mod init
  github.com/adeelahmad/gadgetcast`, go 1.22) and `internal/core` (`ClientID`). Must be
  merged before Wave 2 branches so parallel worktrees do not mint conflicting `go.mod`
  files (`docs/agents/memory.md` M-001). 2 tasks.

**Wave 2 — features (parallel, 4 stories, 15 tasks).**

- `S0-01` signaling-codec — 4 tasks (`internal/signaling`).
- `S0-02` config-model — 4 tasks (`internal/config`).
- `S0-03` pairing-auth — 3 tasks (`internal/pairing`).
- `S0-04` source-switcher — 4 tasks (`internal/switcher`).

All 15 Wave-2 tasks are mutually independent at compile time (each imports only
`internal/core`; the switcher's pairing/signaling couplings are logical, resolved via a
local interface and a shared value shape — see the dependency graph in `stories.md`).
RED can therefore run ≥ 5 workers concurrently.

## Dependency graph

```
        S0-00 core-foundation  (Wave 1: go.mod + internal/core)
        ┌───────┬───────────┬───────────┐
        ▼       ▼           ▼           ▼
     S0-01   S0-02       S0-03       S0-04       (Wave 2: parallel)
   signaling config     pairing     switcher
                           ⇡            ⇡
                           └── logical ─┘  (S0-04 uses IsPaired via a local
                                            PairingOracle interface — no compile edge)
     S0-01 ⇠························ S0-04   (on-air-state = projection of OnAir();
                                             OnAirStateFrom takes core types — no edge)
```

Solid arrows = hard compile dependencies (on `internal/core`). Dotted/⇡ = logical
contract edges resolved without a compile import, so they do not serialize the wave.

## Parallelism table

| Wave | Story | Tasks | Compile deps | Parallel with | RED workers |
|---|---|---|---|---|---|
| 1 | S0-00 core-foundation | T1–T2 | none | — (gates Wave 2) | 1 (serial) |
| 2 | S0-01 signaling-codec | T1–T4 | internal/core | S0-02, S0-03, S0-04 | up to 4 |
| 2 | S0-02 config-model | T1–T4 | internal/core | S0-01, S0-03, S0-04 | up to 4 |
| 2 | S0-03 pairing-auth | T1–T3 | internal/core | S0-01, S0-02, S0-04 | up to 3 |
| 2 | S0-04 source-switcher | T1–T4 | internal/core | S0-01, S0-02, S0-03 | up to 4 |

Wave 2 total: 15 independent tasks → comfortably ≥ 5 concurrent RED workers.

## Critical path

`S0-00 (T1 go.mod → T2 ClientID)` → merge → `S0-04 source-switcher` (the widest Wave-2
story and the one carrying both logical couplings + the never-two-on-air invariant).

- Wave 1 is the only serialization point; it exists solely to create `go.mod` +
  `ClientID` once. Length: 2 tasks.
- Wave 2 length is bounded by the single longest story (4 tasks), not the sum, because
  the four stories run concurrently. Critical path length ≈ 2 (Wave 1) + 4 (longest
  Wave-2 story) = 6 tasks deep, 17 tasks wide.
- No story appears in two waves. No Wave-2 story blocks another Wave-2 story.

## Per-story plan pointers

| Story | tasks.md | validate.md | plan.md (tests-only) |
|---|---|---|---|
| S0-00 | `s0-00-core-foundation/tasks.md` | `s0-00-core-foundation/validate.md` | `s0-00-core-foundation/plan.md` |
| S0-01 | `s0-01-signaling-codec/tasks.md` | `s0-01-signaling-codec/validate.md` | `s0-01-signaling-codec/plan.md` |
| S0-02 | `s0-02-config-model/tasks.md` | `s0-02-config-model/validate.md` | `s0-02-config-model/plan.md` |
| S0-03 | `s0-03-pairing-auth/tasks.md` | `s0-03-pairing-auth/validate.md` | `s0-03-pairing-auth/plan.md` |
| S0-04 | `s0-04-source-switcher/tasks.md` | `s0-04-source-switcher/validate.md` | `s0-04-source-switcher/plan.md` |

## Cross-cutting gates

Copied verbatim from `docs/agents/standards.md` §"Gate matrix" (authoritative — the
GREEN and FINAL gates read that table). Run from repo root after `go mod init`.

| Gate | Command | Pass condition | Status |
|---|---|---|---|
| **fmt** | `gofmt -l .` | Empty stdout (no file listed) | HARD gate |
| **vet** | `go vet ./...` | Exit code 0, no diagnostics | HARD gate |
| **typecheck** | `go build ./...` | Exit code 0 (Go's compiler is the type checker) | HARD gate |
| **test** | `go test ./...` | Exit code 0, all tests pass | HARD gate |
| **coverage** | `go test ./... -coverprofile=coverage.out` then `go tool cover -func=coverage.out \| tail -1` | Total ≥ **80%** on pure-logic packages (core, signaling, config, pairing, switcher) | HARD gate |
| **audit** | `go vet ./...` + `govulncheck ./...` if present | `go vet` clean; `govulncheck` **N/A** — not installed, do not block | HARD (vet) / N-A (govulncheck) |
| **lint (extended)** | `golangci-lint run` | n/a — not installed, OPTIONAL, never blocks | OPTIONAL / not on PATH |
| **lint (extended, static)** | `staticcheck ./...` | n/a — not installed, OPTIONAL | OPTIONAL / not on PATH |
| **security (extended)** | `gosec ./...` | n/a — not installed, OPTIONAL | OPTIONAL / not on PATH |

Notes (from standards.md):

- Before any gate runs on this repo for the first time, `go mod init
  github.com/adeelahmad/gadgetcast` must have been run — this is S0-00 T1 (M-001).
- All five Sprint-0 packages are pure-logic, so the 80% coverage threshold applies to
  every one of them; none is a cgo/hardware wrapper exempt from the numeric threshold.
- `-race` is advisory this sprint (no goroutines/channels); it becomes a HARD gate only
  once a package spawns concurrency across the future capture → gadget-write pipeline.
- If `golangci-lint`/`staticcheck`/`gosec`/`govulncheck` are installed later, promote
  their rows to HARD in a follow-up standards revision — do not silently enforce an
  undeclared tool.

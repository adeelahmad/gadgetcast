---
type: memory
updated_at: "2026-07-09T23:07:16Z"
last_retro_sprint: 0
---

# Entries

| id    | applies_to    | text                                                                                                                              | origin | recurrence | added      |
| ----- | ------------- | --------------------------------------------------------------------------------------------------------------------------------- | ------ | ---------- | ---------- |
| M-001 | scaffolder    | A fresh Go module has no `go.mod` — run `go mod init <module>` before the first RED dispatch, or `go test` fails at compile before any assertion runs.        | seed   | seed       | 2026-07-09 |
| M-002 | red-worker    | `go test ./...` is the suite command. Keep a RED test failing on the ASSERTION, not a missing symbol: add the minimal type/func/zero-value shim needed to compile. | seed   | seed       | 2026-07-09 |
| M-003 | green-worker  | Implement only what the RED assertions require; do not widen the exported API beyond the task's tests — keeps the diff reviewable and avoids scope creep.       | seed   | seed       | 2026-07-09 |
| M-004 | all           | Sprint 0 for this project: no prior failure/feedback trail exists yet. Treat these as generic Go-TDD priors, not repo-specific lessons — replace once real recurrence (>=2) is observed. | seed   | seed       | 2026-07-09 |
| M-005 | green-worker  | `standards.md` gate matrix must live in a FENCED code block: `run_standards_matrix` parses fences only and ignores markdown tables, falling back to the default (Rust) matrix that fails a Go repo. Keep a fenced Go command block in standards.md.  | execution | 1 | 2026-07-09 |
| M-006 | red-worker    | In `.agentic/task.env`, write SCOPE_GLOBS QUOTED with a subtree glob: `SCOPE_GLOBS="internal/pkg internal/pkg/*"`. A bare dir path doesn't match files under the gate's case-glob matcher; an unquoted multi-value line breaks `. task.env` sourcing.  | execution | 1 | 2026-07-09 |
| M-007 | scaffolder    | `SCAFFOLD_SYMBOLS` is a literal space-separated list of BARE symbol names (ctx-symbols keys Go methods by bare name, e.g. `String` not `ClientID.String`), NOT a file path. Set it to the real symbols or leave empty to use `.agentic/scaffold-symbols`.  | execution | 1 | 2026-07-09 |
| M-008 | structural-reviewer | `ctx-symbols conflicts --tree` is Rust-semantic and false-flags HIGH on Go: same-named methods on different receivers (`Type`/`Validate`), package-scoped repeats (`New`/`ErrUnknownClient`), type-vs-accessor (`State`), and even local vars/test helpers; `Test*` funcs show as LOW unreferenced. Do the structural review via `go vet`/`go build` + import-graph reasoning; escalate the gate's Go-incompatibility, don't HALT on false HIGHs. | execution | 1 | 2026-07-09 |

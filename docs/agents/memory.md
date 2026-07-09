---
type: memory
updated_at: "2026-07-09T21:44:45Z"
last_retro_sprint: 0
---

# Entries

| id    | applies_to    | text                                                                                                                              | origin | recurrence | added      |
| ----- | ------------- | --------------------------------------------------------------------------------------------------------------------------------- | ------ | ---------- | ---------- |
| M-001 | scaffolder    | A fresh Go module has no `go.mod` — run `go mod init <module>` before the first RED dispatch, or `go test` fails at compile before any assertion runs.        | seed   | seed       | 2026-07-09 |
| M-002 | red-worker    | `go test ./...` is the suite command. Keep a RED test failing on the ASSERTION, not a missing symbol: add the minimal type/func/zero-value shim needed to compile. | seed   | seed       | 2026-07-09 |
| M-003 | green-worker  | Implement only what the RED assertions require; do not widen the exported API beyond the task's tests — keeps the diff reviewable and avoids scope creep.       | seed   | seed       | 2026-07-09 |
| M-004 | all           | Sprint 0 for this project: no prior failure/feedback trail exists yet. Treat these as generic Go-TDD priors, not repo-specific lessons — replace once real recurrence (>=2) is observed. | seed   | seed       | 2026-07-09 |

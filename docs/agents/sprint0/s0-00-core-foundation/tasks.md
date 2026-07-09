---
type: tasks
story: S0-00
---

# Tasks — S0-00 core-foundation

Wave 1. Establishes the single Go module and the shared `internal/core` vocabulary.
Must be merged before any Wave-2 feature story branches (`docs/agents/memory.md` M-001).
Stdlib only.

## T1 — Initialize the Go module

Run `go mod init github.com/adeelahmad/gadgetcast` at repo root and pin `go 1.22` in
`go.mod`. This is the one-and-only `go.mod` for the sprint (M-001): it exists so that
`go build ./...` / `go test ./...` compile and so parallel Wave-2 worktrees inherit one
module rather than minting conflicting ones.

- Atomic: touches only `go.mod` (and `go.sum` if ever needed — it will not, stdlib only).
- Testable/demoable: `go build ./...` exits 0 on the empty module; `go vet ./...` clean;
  `gofmt -l .` empty.
- No third-party requires appear in `go.mod`.

## T2 — `internal/core` ClientID value type

Add `internal/core/client_id.go` defining `ClientID` as a strongly-typed comparable
value (usable as a map key, comparable with `==`) with `NewClientID(string)
(ClientID, error)`, `String() string`, and an `IsValid()`/`IsZero()` distinction. Empty
or all-whitespace input is rejected with a typed sentinel error (e.g.
`ErrInvalidClientID`); the zero value reports invalid. Minimal surface — only what
S0-01..S0-04 need.

- Atomic: one file + its test; no other package touched.
- Testable/demoable: table-driven tests in `internal/core/client_id_test.go` cover
  valid construction, empty/whitespace rejection, equality/map-key usability, and the
  zero-value-invalid case (this is the "at least one real behavioral test" the
  foundation story requires).
- Constraint: imports only the standard library.

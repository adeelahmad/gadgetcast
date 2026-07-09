---
type: validate
story: S0-00
---

# Validate — S0-00 core-foundation

PASS/FAIL rubric mirroring the task headings. Literal commands, literal expected output.
Failure modes cite the rule each violates. Run from repo root
`/Users/adeelahmad/work/gadgetcast`.

## Pre-flight

- [ ] `go version` prints `go1.22...` (standards §Detected stack; go 1.22 required).
- [ ] Working tree is on the Wave-1 branch for S0-00; no `go.mod` exists yet before T1.
- [ ] No third-party tooling assumed: only `go` and `gofmt` are used (standards
      Gate-matrix — `golangci-lint`/`staticcheck`/`gosec`/`govulncheck` are OPTIONAL/absent).

## T1 — Initialize the Go module

| Check | Command | Expected | Fails rule |
|---|---|---|---|
| module created | `head -1 go.mod` | `module github.com/adeelahmad/gadgetcast` | dispatch decision #1 (module path) / M-001 |
| go version pinned | `grep '^go ' go.mod` | `go 1.22` | standards §Detected stack (Go 1.22) |
| no third-party deps | `grep -c '^require' go.mod` | `0` | decision #7 (stdlib only) |
| compiles | `go build ./...` | exit 0, no output | standards Gate-matrix **typecheck** |
| vet clean | `go vet ./...` | exit 0, no output | standards Gate-matrix **vet** |
| formatted | `gofmt -l .` | empty stdout | standards Gate-matrix **fmt** |

FAIL if `go.mod` names any other module path, pins a `go` line other than `1.22`, lists
any `require`, or a second `go.mod` appears anywhere (violates M-001 "one go.mod").

## T2 — internal/core ClientID value type

| Check | Command | Expected | Fails rule |
|---|---|---|---|
| tests pass | `go test ./internal/core/` | `ok  github.com/adeelahmad/gadgetcast/internal/core` | standards Gate-matrix **test** |
| valid construction | `go test ./internal/core/ -run TestNewClientID_Valid -v` | `--- PASS: TestNewClientID_Valid` | S0-00 success (valid id) |
| empty rejected | `go test ./internal/core/ -run TestNewClientID_EmptyRejected -v` | `--- PASS: TestNewClientID_EmptyRejected` | S0-00 failure (empty accepted) |
| whitespace rejected | `go test ./internal/core/ -run TestNewClientID_WhitespaceRejected -v` | `--- PASS: TestNewClientID_WhitespaceRejected` | S0-00 failure (empty/blank identity) |
| equality + map key | `go test ./internal/core/ -run TestClientID_EqualityAndMapKey -v` | `--- PASS: TestClientID_EqualityAndMapKey` | S0-00 constraint (comparable, map-key) |
| zero invalid | `go test ./internal/core/ -run TestClientID_ZeroValueInvalid -v` | `--- PASS: TestClientID_ZeroValueInvalid` | S0-00 failure (zero masquerades as client) |
| core coverage | `go test ./internal/core/ -cover` | `coverage: ` ≥ `80.0% of statements` | standards Gate-matrix **coverage** |
| stdlib only | `go list -deps ./internal/core/ \| grep -v '^internal/\|^github.com/adeelahmad/gadgetcast'` prints only stdlib paths | no third-party module path | decision #7 (stdlib only) |

FAIL if `ClientID` is a bare `string` alias with no validation, if the zero value
reports valid (S0-00 failure scenario), or if `internal/core` imports any non-stdlib
package (decision #7 / standards §Security stdlib-only).

## Final sign-off

- [ ] All T1 and T2 rows PASS.
- [ ] `go build ./...`, `go vet ./...`, `go test ./...` all exit 0; `gofmt -l .` empty.
- [ ] `internal/core` coverage ≥ 80%.
- [ ] Exactly one `go.mod` in the tree, module `github.com/adeelahmad/gadgetcast`,
      `go 1.22`, zero `require` entries.
- [ ] S0-00 merged to the integration branch BEFORE any Wave-2 story branches (M-001).

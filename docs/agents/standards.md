---
type: standards
sprint: 0
stack: go
detected_at: "2026-07-10T00:00:00Z"
---

# Standards — gadgetcast

Rule digest and gate matrix this sprint's execution binds to. Emitted by the
`standards` activity; cited sources are the only ones the planner and workers
may rely on. See `pipeline/planning/01-standards/gate.md` for the enforcement
gate (`bin/gate-standards-cited`).

## Detected stack

- **Language:** Go 1.22.2 (`go version` → `go1.22.2 darwin/arm64`; `go` is on `$PATH`).
- **Evidence:** `.gitignore` at repo root is the GitHub Go template (`*.exe`, `*.test`, `coverage.*`, `go.work` entries); `README.md` names Pion (a Go WebRTC library) as the primary WebRTC-endpoint candidate ("WebRTC endpoint (Pion / GStreamer webrtcbin / aiortc)").
- **Repo state:** no `go.mod` yet — this is Sprint 0 / Phase 0 (feasibility spike), pre-first-commit-of-code. `docs/agents/memory.md` (M-001) already anticipates this: the scaffolder runs `go mod init <module>` before the first RED dispatch.
- **Toolchain audit (what's actually on this machine — do not gate on absent tools):**

  | Tool | On `$PATH`? | Use in gates below |
  |---|---|---|
  | `go` (fmt/vet/test/build) | yes | fmt, vet, test, coverage, typecheck |
  | `golangci-lint` | **no** | OPTIONAL only, never a hard gate |
  | `staticcheck` | no | OPTIONAL only, never a hard gate |
  | `gosec` | no | OPTIONAL only, never a hard gate |
  | `govulncheck` | no | OPTIONAL only, marked N/A in audit gate |
  | `goimports` | no | OPTIONAL only; `gofmt -l` is the hard gate |

## Active rules

Every bullet cites the local rule file that binds it plus, where the local file itself doesn't spell out the underlying idiom, the canonical upstream Go source the local rule is built on.

**Formatting**
- `gofmt` (and `goimports` where available) is mandatory; no style debates, do not hand-format. Source: `rules/golang/coding-style.md` ("gofmt and goimports are mandatory — no style debates"); background: Effective Go, "Formatting" (https://go.dev/doc/effective_go#formatting).

**Errors**
- Wrap errors with `%w` and calling context; never return a bare upstream error with no context added. Source: `rules/golang/coding-style.md` ("Always wrap errors with context" / `fmt.Errorf("failed to create user: %w", err)`).
- Never silently swallow an error; handle explicitly at every level, log detailed context server-side. Source: `rules/common/coding-style.md` ("Error Handling") and `CLAUDE.md` Coding Rules.

**API design**
- Accept interfaces, return structs; keep interfaces small (1-3 methods) and defined at their point of use, not next to the implementation. Source: `rules/golang/coding-style.md` ("Accept interfaces, return structs" / "Keep interfaces small") and `rules/golang/patterns.md` ("Small Interfaces").
- Use constructor functions / functional options for configurable components (e.g. the composite-gadget builder, the WebRTC endpoint config) instead of large parameter lists. Source: `rules/golang/patterns.md` ("Functional Options", "Dependency Injection").

**Testing**
- Standard `go test`, table-driven tests as the default shape. Source: `rules/golang/testing.md` ("Use the standard `go test` with table-driven tests"); background: the `testing` package (https://pkg.go.dev/testing).
- Run the RED test failing on the *assertion*, not a missing symbol — add the minimal type/func/zero-value shim needed to compile first. Source: `docs/agents/memory.md` M-002 (repo-specific seed lesson).
- Implement only what the RED assertions require; do not widen the exported API beyond the task's tests. Source: `docs/agents/memory.md` M-003.
- `-race` is recommended for any package touching goroutines/channels (the WebRTC/gadget pipeline is inherently concurrent: capture, resample, USB gadget write are separate stages). Source: `rules/golang/testing.md` ("Always run with the `-race` flag"). Advisory, not a hard gate, until Sprint 0 has real concurrent code (see Gate matrix note).
- Minimum 80% coverage on pure-logic packages (protocol/signaling parsing, resampling math, config). Source: `rules/common/testing.md` ("Minimum Test Coverage: 80%"). Packages that are thin wrappers over cgo/hardware (`configfs`, V4L2 M2M, ALSA) are exempt from the numeric threshold — see Gate matrix.

**Security**
- No hardcoded secrets; read config (pairing PIN, Tailscale auth, etc.) from environment variables and fail fast at startup if a required one is missing. Source: `rules/golang/security.md` ("Secret Management") and `rules/common/security.md` ("Secret Management").
- Use `context.Context` for timeout control on every network/IO call (WebRTC session setup, WS signaling, USB gadget writes). Source: `rules/golang/security.md` ("Context & Timeouts").
- This device is a network-controlled camera/mic injector for the host per the project's own threat model (`README.md` §4 "Security"); validate all input at the WS/HTTP signaling boundary, never trust an unauthenticated LAN client. Source: `rules/common/security.md` ("Mandatory Security Checks") and `CLAUDE.md` Coding Rules ("Only validate at system boundaries").

**File & project organization**
- Many small files over few large ones: 200-400 lines typical, 800 max; organize by feature/domain (e.g. `gadget/`, `webrtc/`, `signaling/`), not by type. Source: `rules/common/coding-style.md` ("File Organization").
- Every project needs `CLAUDE.md` + `DEVLOG.md` at the repo root; gadgetcast currently has neither — flagged as an open gap for the `planner`/`intake` activities, not fixed here (this activity is read-only). Source: `rules/project-docs.md` ("Required Docs by Project Type") and `CLAUDE.md` Workflow section ("DEVLOG.md in every project root").

**Commits**
- Conventional commit format `<type>: <description>` with types `feat|fix|refactor|docs|test|chore|perf|ci`. Source: `rules/common/git-workflow.md` ("Commit Message Format").

## Gate matrix (authoritative — GREEN and FINAL gates read this table verbatim)

| Gate | Command | Pass condition | Status |
|---|---|---|---|
| **fmt** | `gofmt -l .` | Empty stdout (no file listed) | HARD gate |
| **vet** | `go vet ./...` | Exit code 0, no diagnostics | HARD gate |
| **typecheck** | `go build ./...` | Exit code 0 (Go's compiler is the type checker; no separate typecheck tool exists in this toolchain) | HARD gate |
| **test** | `go test ./...` | Exit code 0, all tests pass | HARD gate |
| **coverage** | `go test ./... -coverprofile=coverage.out` then `go tool cover -func=coverage.out \| tail -1` | Total ≥ **80%** on pure-logic packages (protocol, resampling, config, signaling). Direct cgo/hardware-binding packages (`configfs`, V4L2, ALSA) are exempt from the numeric threshold but must still have `go test` passing (compile + smoke) | HARD gate (threshold scoped as above) |
| **audit** | `go vet ./...` (repeat of the vet gate) + `govulncheck ./...` if present | `go vet` must be clean. `govulncheck`: **N/A** — not installed on this machine; do not block on it | HARD (vet) / N-A (govulncheck) |
| **lint (extended)** | `golangci-lint run` | n/a — **not installed**, OPTIONAL, never blocks GREEN/FINAL | OPTIONAL / not on PATH |
| **lint (extended, static)** | `staticcheck ./...` | n/a — **not installed**, OPTIONAL | OPTIONAL / not on PATH |
| **security (extended)** | `gosec ./...` | n/a — **not installed**, OPTIONAL | OPTIONAL / not on PATH |

Notes:
- Before any gate runs on this repo for the first time: `go mod init <module>` must have been run (no `go.mod` exists yet). Source: `docs/agents/memory.md` M-001.
- Race detection (`go test -race ./...`, per `rules/golang/testing.md`) is advisory for Sprint 0 and becomes a HARD gate the moment a package spawns goroutines/channels across the capture → resample → gadget-write pipeline described in `README.md` §3 (Architecture).
- If `golangci-lint`, `staticcheck`, `gosec`, or `govulncheck` are installed later, promote their rows from OPTIONAL to HARD in a follow-up standards revision — do not silently start enforcing an undeclared tool.

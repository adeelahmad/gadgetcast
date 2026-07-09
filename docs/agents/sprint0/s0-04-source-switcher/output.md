---
type: output
story: S0-04
---

## S0-04 · attempt 1 · red-worker · 2026-07-09T22:40:09Z
status: ok

### Summary
Wrote all 14 failing tests for package `internal/switcher` at their exact
`internal/switcher/switcher_test.go::TestName` paths per plan.md (T1–T4), each
injecting a switcher-local `PairingOracle` stub. Added one `// agentic:shim`
file (`internal/switcher/shim.go`) declaring the minimal production surface
(`Switcher`, `New(PairingOracle) *Switcher`, `AddClient`, `RemoveClient`,
`Has`, `SetOnAir`, `OnAir() (core.ClientID, bool)`, and the `PairingOracle`
interface) with deliberately wrong bodies: `AddClient`/`RemoveClient`/`SetOnAir`
never mutate and never error, `Has` always denies membership, and `OnAir`
always claims the zero client is live. That single broken shim makes every one
of the 14 tests fail on an assertion (not a missing symbol, not a panic). Imports
only `internal/core` + stdlib. `gofmt` and `go vet ./internal/switcher` are clean.

### Result
| Check | Status | Detail |
|---|---|---|
| `switcher_test.go::TestAddClient_AppearsInSet` | FAIL | fails by assertion (Has=false after add) |
| `switcher_test.go::TestAddClient_Duplicate_Errors` | FAIL | fails by assertion (dup add returns nil) |
| `switcher_test.go::TestRemoveClient_Unknown_Errors` | FAIL | fails by assertion (remove-unknown returns nil) |
| `switcher_test.go::TestNewSwitcher_NoneOnAirByDefault` | FAIL | fails by assertion (OnAir live=true by default) |
| `switcher_test.go::TestSetOnAir_MakesOnlyThatClientOnAir` | FAIL | fails by assertion (OnAir != a) |
| `switcher_test.go::TestSetOnAir_Switch_A_to_B_DisplacesA` | FAIL | fails by assertion (OnAir != b) |
| `switcher_test.go::TestOnAir_ExactlyOne_AcrossSequence` | FAIL | fails by assertion (never-two invariant / identity) |
| `switcher_test.go::TestSetOnAir_UnknownClient_Errors_StateUnchanged` | FAIL | fails by assertion (set-unknown returns nil) |
| `switcher_test.go::TestRemoveClient_OnAir_TransitionsToNobody` | FAIL | fails by assertion (still live after remove) |
| `switcher_test.go::TestRemoveClient_NotOnAir_LeavesOnAirUntouched` | FAIL | fails by assertion (OnAir != a) |
| `switcher_test.go::TestSetOnAir_UnpairedClient_Errors` | FAIL | fails by assertion (unpaired set returns nil) |
| `switcher_test.go::TestSetOnAir_PairedClient_Succeeds` | FAIL | fails by assertion (OnAir != p) |
| `switcher_test.go::TestUnpairedClient_MayConnect_ButNotOnAir` | FAIL | fails by assertion (Has=false after add) |
| `switcher_test.go::TestOnAir_ReturnsSourceOfTruthShape` | FAIL | fails by assertion (OnAir != (p,true)) |

### Next
Scaffold the named symbols in `internal/switcher`: stub `Switcher`, `New`,
`AddClient`, `RemoveClient`, `Has`, `SetOnAir`, `OnAir`, and the `PairingOracle`
interface as `panic("SUB-AGENT-TODO: ...")` bodies (deleting the shim), then
GREEN the membership set, single-on-air selection with displacement, unknown/
remove-on-air transitions, and the paired-only on-air gate.

## S0-04 · attempt 1 · scaffolder · 2026-07-09T22:45:31Z
status: ok

### Summary
Created the single canonical skeleton `internal/switcher/switcher.go` declaring
the full production surface the RED tests reference: the `PairingOracle`
interface (`IsPaired(core.ClientID) bool` — type decl, no body), the `Switcher`
struct, the `New(PairingOracle) *Switcher` constructor, and methods `AddClient`,
`RemoveClient`, `Has`, `SetOnAir`, and `OnAir() (core.ClientID, bool)`. Every
function/method body is `panic("SUB-AGENT-TODO: ...")` with a recipe describing
what GREEN must implement; no real switching, membership, gating, or on-air
logic exists yet, so the 14 tests still fail by absence of behavior (panic).
Deleted the now-redundant `internal/switcher/shim.go`; no `// agentic:shim`
marker remains. Imports only `internal/core` + stdlib. Wrote the eight bare
symbol names to `.agentic/scaffold-symbols`. `gofmt -w`, `go build
./internal/switcher` (exit 0), and `go vet` are all clean.

### Scaffold
+ create internal/switcher/switcher.go (Switcher, New, AddClient, RemoveClient, Has, SetOnAir, OnAir, PairingOracle)
- delete internal/switcher/shim.go (// agentic:shim)
+ create .agentic/scaffold-symbols (8 bare names)

### Result
| Check | Status | Detail |
|---|---|---|
| stubs panic+TODO | PASS | 6 funcs/methods panic("SUB-AGENT-TODO: …") |
| interface decl no body | PASS | PairingOracle is a type decl |
| each symbol defined once | PASS | 8 symbols, single definition |
| shim removed | PASS | no // agentic:shim remains |
| go build ./internal/switcher | PASS | exit 0 |
| go vet ./internal/switcher | PASS | clean |
| gofmt | PASS | clean |

### Next
green: fill the stubs — membership set with duplicate/unknown typed errors,
single on-air selection with displacement, remove-on-air → nobody transition,
and the known+paired on-air gate.

## S0-04 · attempt 1 · green-worker · 2026-07-09T22:50:51Z
status: ok
### Summary
Filled the six panic-TODO bodies in `internal/switcher/switcher.go` with the
minimum logic to pass all 14 RED tests. `New` builds the Switcher with the
injected oracle and an empty `map[core.ClientID]bool` client set. `AddClient`
records membership and returns `ErrDuplicateClient` on a repeat, leaving the set
unchanged. `RemoveClient` returns `ErrUnknownClient` for a client not present,
else deletes it and, if it was the live on-air client, transitions on-air to
nobody (zero id, live=false). `Has` reports set membership. `SetOnAir` gates on
membership (`ErrUnknownClient`) then on `oracle.IsPaired` (`ErrNotPaired`),
leaving on-air state untouched on either error; otherwise it sets the single
on-air id and live=true, displacing any prior selection so there is never two
on-air. `OnAir()` returns the stored `(core.ClientID, bool)` as the single
source of truth. Single-threaded, no locks. Only `internal/switcher/switcher.go`
changed; imports are stdlib `errors` + `internal/core`.
### Result
| Check | Status | Detail |
|---|---|---|
| `internal/switcher::14 tests` | PASS | `go test ./internal/switcher` ok, all 14 pass |
| regressed | NONE | no prior-green tests touched |
| diff-scope | PASS | only internal/switcher/switcher.go (sprint0...HEAD) |
| fmt | PASS | `gofmt -l internal/switcher` empty |
| vet | PASS | `go vet ./internal/switcher` clean |
| coverage | PASS | 100.0% of statements (≥80%) |
### Next
structural review — confirm no orphan/parallel switcher implementation; then S0-04 done.

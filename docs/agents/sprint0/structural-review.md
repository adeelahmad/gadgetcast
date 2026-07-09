# Structural review — Sprint 0, Wave 2 (supervisor-performed)

Read-only structural review after all Wave-2 GREEN merges. Performed by the supervisor
directly (not the `structural-reviewer` sub-agent) because `ctx-symbols conflicts` is
Rust-semantic and produces false HIGH findings on any valid Go program — dispatching the
reviewer would have spuriously HALTed the chain (see Tooling gap below). The review INTENT
(orphan modules · parallel implementations · duplicate helpers) is satisfied by the
analysis here, cross-checked against `go build`/`go vet` (the real Go duplicate detectors).

## Verdict: CLEAN — no foundation-poisoning, no HIGH structural defect.

### Import graph (verified via `go list`)
- `internal/core` → stdlib only
- `internal/signaling` → only `internal/core`
- `internal/config` → stdlib only
- `internal/pairing` → only `internal/core`
- `internal/switcher` → only `internal/core` (does NOT import pairing/signaling — the
  `PairingOracle` interface (decision #5) and `OnAirStateFrom` projection (decision #8)
  keep the coupling logical, not a compile edge). Confirmed.
- No third-party dependencies anywhere (stdlib-only; `go.mod` has no `require`).

### Orphan modules: none.
Every package is exercised by its own tests and is a declared foundation for later sprints
(WebRTC/GStreamer pipeline reads `switcher.OnAir()`; the WS server uses the signaling codec;
startup reads config; the switcher gates on pairing). Nothing is dead.

### Parallel implementations of one abstraction: none.
Each abstraction is defined in exactly one package: `ClientID` (core), the 7-message
codec (signaling), config parse/validate/defaults (config), PIN pairing (pairing),
the exactly-one-on-air switcher (switcher). `IsPaired` appearing in both `pairing`
(`Manager.IsPaired`) and `switcher` (`PairingOracle` interface + test stub) is the INTENDED
contract coupling, not two implementations of one thing.

### Duplicate helpers: one, idiomatic and acceptable.
- `mustID` test helper is defined in 3 `_test.go` packages (pairing, signaling, switcher).
  Unexported Go test helpers cannot be shared across packages without an exported testutil
  package; three 3-line helpers is the idiomatic choice for Sprint 0 ("three similar lines >
  premature abstraction"). Not a production concern.
- `ErrUnknownClient` sentinel exists in both `pairing` and `switcher` as distinct
  package-scoped errors. Idiomatic; no shared-error abstraction warranted yet.

## Tooling gap (ESCALATION for next planning session)
`ctx-symbols conflicts --tree` flags as HIGH "duplicate definition": same-named methods on
different receiver types (`Type`, `Validate`), package-scoped identifiers repeated across
packages (`New`, `ErrUnknownClient`, `IsPaired`), a type + its accessor sharing a name
(`State`/`State()`), and even LOCAL VARIABLES (`de`, `w`, `limit`) and per-package test
helpers (`mustID`). Its LOW "never referenced" hits are all `Test*` funcs (run by `go test`,
not referenced in code). None is a real Go defect — `go build`/`go vet` are clean. The
`gate-structural-integrity` gate is therefore unusable for Go as-is: it would HALT every
valid Go codebase. Next planning session should either point the structural gate at a
Go-aware duplicate detector (e.g. `go vet`, `staticcheck U1000` for unused, or an
AST-package-aware tool) or scope `ctx-symbols` conflict detection by package + receiver.

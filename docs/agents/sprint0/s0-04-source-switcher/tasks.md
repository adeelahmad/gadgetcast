---
type: tasks
story: S0-04
---

# Tasks — S0-04 source-switcher

Wave 2 (parallel; logically downstream). Package `internal/switcher`. Stdlib only;
imports only `internal/core`. Single-threaded value logic (no goroutines). Consumes
pairing via a switcher-local interface (accept interfaces) and never imports
`internal/pairing` or `internal/signaling` (see dependency graph in `stories.md`).

## T1 — Client set: add / remove

Define the switcher in `internal/switcher/switcher.go` with `New(...)`,
`AddClient(core.ClientID)`, `RemoveClient(core.ClientID)`, and a membership view. Adding
a duplicate id and removing an unknown id return typed errors. No on-air logic yet.

- Atomic: the client-set operations.
- Testable: `switcher_test.go` covers add-appears, duplicate-add error, remove-unknown
  error.

## T2 — set-on-air + exactly-one invariant

Implement `SetOnAir(core.ClientID)` and `OnAir() (core.ClientID, bool)`. None on-air by
default; `SetOnAir(A)` makes A on-air; `SetOnAir(B)` while A is on-air displaces A
atomically so at no observable point are two clients on-air (README §3.3; intake
failures #1, success #1/#2; invariant: never-two-on-air).

- Atomic: the on-air pointer + set/switch transition + `OnAir()` query.
- Testable: `switcher_test.go` covers none-by-default, set makes only-A, switch A→B
  displaces A, and an exactly-one property across an add/set/switch sequence.

## T3 — Unknown-client + remove-on-air behavior

`SetOnAir` with an id not in the set returns a typed error and leaves on-air unchanged
(no phantom on-air client; intake failure #2). Removing the currently on-air client
transitions on-air to "nobody" (intake failure #3); removing a non-on-air client leaves
the on-air pointer untouched (intake success #3).

- Atomic: the unknown-id guard + remove-on-air transition.
- Testable: `switcher_test.go` covers set-unknown-errors-state-unchanged,
  remove-on-air→nobody, remove-non-on-air-untouched, remove-unknown-errors.

## T4 — Pairing gate (unpaired cannot go on-air)

The switcher accepts a small locally-defined `PairingOracle` interface
(`IsPaired(core.ClientID) bool`); tests inject a stub. An unpaired client MAY be added
(appears in the set) but `SetOnAir` on it returns a typed error; a paired client
succeeds (decision #5; intake failure #8). `OnAir()` remains the single source of truth
that S0-01's `OnAirStateFrom` projects (decision #8).

- Atomic: the oracle check in `SetOnAir` + the interface declaration.
- Testable: `switcher_test.go` covers unpaired-set-on-air-errors,
  paired-set-on-air-succeeds, unpaired-may-connect-but-not-on-air, and that `OnAir()`
  returns the `(core.ClientID, bool)` contract shape.

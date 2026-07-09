---
type: tasks
story: S0-03
---

# Tasks — S0-03 pairing-auth

Wave 2 (parallel). Package `internal/pairing`. Stdlib only; imports only `internal/core`.
Per-client PIN pairing (decision #4 — PIN-only, QR deferred). No hardcoded PIN in source
(standards §Security — tests inject a known PIN).

## T1 — State type + pending→paired transition

Define the pairing state (`pending`/`paired`/`rejected`) and per-client machine in
`internal/pairing/pairing.go`, plus a constructor taking the expected PIN and the
configurable attempt limit. `SubmitPIN(id, pin)` on a correct PIN transitions
`pending → paired` exactly once and stays `paired` (idempotent). An unknown client id
returns a typed error.

- Atomic: state type + `New` + the correct-PIN transition.
- Testable: `pairing_test.go` covers start-pending, correct-PIN→paired, idempotent
  re-submit, unknown-client error.

## T2 — Attempt limit + durable rejection (lockout)

Wrong PINs increment an attempt counter while `pending`; on exhausting the configured
limit the client transitions `pending → rejected`. Rejection is durable: further
attempts — including a subsequently-correct PIN — are refused, never reverting to
`pending` (README §4; intake failure #7; invariant: brute-force-lockout).

- Atomic: the counter + lockout transition + durability guard.
- Testable: `pairing_test.go` covers wrong-PIN-stays-pending, exhaust→rejected,
  post-rejection attempts still rejected, correct-PIN-after-lockout still rejected.

## T3 — IsPaired query + configurable attempt limit

Expose `IsPaired(core.ClientID) bool` — true only in `paired` (false for `pending` and
`rejected`) — the predicate the switcher (S0-04) consumes (decision #5). Confirm the
attempt limit is honored as configured (not hardwired), defaulting to 5 when the config
default is used (decision #3).

- Atomic: the query + configurable-limit wiring.
- Testable: `pairing_test.go` covers IsPaired true/false per state, a non-default limit
  (e.g. 3), and that rejected state stays queryable.

---
type: tasks
story: S0-01
---

# Tasks — S0-01 signaling-codec

Wave 2 (parallel). Package `internal/signaling`. Stdlib `encoding/json` only. Imports
only `internal/core`. The seven messages: `join`, `offer`, `answer`, `ice-candidate`,
`switch-source`, `on-air-state`, `error` (decision #2 — authoritative set).

## T1 — Message types + type-tagged envelope

Define the seven message structs and the JSON envelope in
`internal/signaling/message.go`: a `MessageType` string constant per wire name and an
`Envelope{ Type MessageType; ... }` (or equivalent tagged union) that carries exactly
one typed payload. `join`/`on-air-state` reference `core.ClientID`. No behavior yet
beyond the type declarations and their `type` constants.

- Atomic: one file of type/const declarations.
- Testable: compiled and exercised by the round-trip tests in T2.

## T2 — Encode + decode (round-trip, strongly typed)

Implement `Encode(msg) ([]byte, error)` and `Decode([]byte) (<typed>, error)` in
`internal/signaling/codec.go`. Decode dispatches on the `type` tag to a concrete struct
(never `map[string]interface{}`). All seven types round-trip encode → JSON → decode
without loss.

- Atomic: the codec functions + their dispatch table.
- Testable: `codec_test.go` round-trips each of the seven types and asserts the decoded
  value is the concrete struct type (not a map).

## T3 — Decode error handling (never panic)

Harden `Decode`: an unrecognized `type`, truncated/malformed JSON, or a type/payload
mismatch returns a typed decode error and NEVER panics or substitutes a zero value
(intake failure #4; invariant: unknown-type-decode-errors-not-panics).

- Atomic: error paths in `codec.go` + a typed `DecodeError` (wraps context with `%w`,
  standards §Errors).
- Testable: a malformed-input table in `codec_test.go` asserts a typed error and no
  panic for every entry.

## T4 — Per-message field validation + on-air projection

Add `internal/signaling/validate.go` with `Validate()` per message (offer/answer need
SDP, join needs client id, ice-candidate needs candidate, switch-source needs target)
and the projection helper `OnAirStateFrom(id core.ClientID, live bool)
OnAirStateMessage` in `internal/signaling/onair.go` (decision #8 — the on-air-state
message is a projection of switcher state; takes core-level values, imports no switcher).

- Atomic: validation rules + the pure projection function.
- Testable: `validate_test.go` and `onair_test.go` cover each missing-field rejection
  and both projection cases (live / nobody), plus a projection round-trip.

---
type: validate
story: S0-01
---

# Validate — S0-01 signaling-codec

PASS/FAIL rubric mirroring the task headings. Literal commands, literal expected output.
Failure modes cite the rule each violates. Run from repo root.

## Pre-flight

- [ ] S0-00 is merged: `go build ./internal/core/` exits 0 (this story imports
      `internal/core`).
- [ ] `go test ./internal/signaling/` compiles (RED tests present).
- [ ] `internal/signaling` imports only stdlib + `internal/core`:
      `go list -deps ./internal/signaling/ | grep -v '^internal/\|^github.com/adeelahmad/gadgetcast'`
      lists only standard-library paths (decision #7; no import of `internal/switcher`).

## T1 — message types + envelope

| Check | Command | Expected | Fails rule |
|---|---|---|---|
| compiles | `go build ./internal/signaling/` | exit 0 | standards Gate-matrix **typecheck** |
| seven type constants present | `go doc ./internal/signaling` | lists a `MessageType` const per wire name (join…error) | decision #2 (seven-message set) |

## T2 — encode/decode round-trip

| Check | Command | Expected | Fails rule |
|---|---|---|---|
| all round-trips pass | `go test ./internal/signaling/ -run 'TestEncodeDecode_' -v` | seven `--- PASS: TestEncodeDecode_*_RoundTrip` lines | intake success #4 (lossless round-trip) |
| decode is concrete struct | `go test ./internal/signaling/ -run TestDecode_ReturnsConcreteStruct_NotMap -v` | `--- PASS` | intake failure #4 / success #4 (not map[string]interface{}) |

FAIL if any type decodes to `map[string]interface{}`, or if any round-trip drops a
field (intake success #4).

## T3 — decode never panics

| Check | Command | Expected | Fails rule |
|---|---|---|---|
| unknown type errors | `go test ./internal/signaling/ -run TestDecode_UnknownType_ReturnsError -v` | `--- PASS` | intake failure #4 (invariant: unknown-type-decode-errors) |
| truncated JSON errors | `go test ./internal/signaling/ -run TestDecode_TruncatedJSON_ReturnsError -v` | `--- PASS` | intake failure #4 |
| never panics (table) | `go test ./internal/signaling/ -run TestDecode_NeverPanics_MalformedTable -v` | `--- PASS` | intake failure #4 (no panic, no zero-substitution) |
| type/payload mismatch | `go test ./internal/signaling/ -run TestDecode_TypePayloadMismatch_ReturnsError -v` | `--- PASS` | intake failure #4 |

FAIL if any malformed input panics or returns a nil error with a zero-value message
(intake failure #4; standards §Errors — never silently swallow).

## T4 — validation + on-air projection

| Check | Command | Expected | Fails rule |
|---|---|---|---|
| missing-field rejections | `go test ./internal/signaling/ -run TestValidate_ -v` | all `TestValidate_*_Errors` PASS, `TestValidate_WellFormedMessages_OK` PASS | intake success #4 / failure #4 (validated struct) |
| projection live | `go test ./internal/signaling/ -run TestOnAirStateFrom_Live -v` | `--- PASS` | decision #8 (projection of switcher state) |
| projection nobody | `go test ./internal/signaling/ -run TestOnAirStateFrom_Nobody -v` | `--- PASS` | intake success (nobody on-air) |
| projection round-trip | `go test ./internal/signaling/ -run TestOnAirState_Projection_RoundTrip -v` | `--- PASS` | decision #8 (lockstep with OnAir() shape) |

FAIL if `OnAirStateFrom` imports `internal/switcher` (breaks the no-compile-edge
decoupling in the dependency graph) or if a missing required field validates OK.

## Final sign-off

- [ ] Every T1–T4 row PASS; `go test ./internal/signaling/` exits 0.
- [ ] `gofmt -l internal/signaling` empty; `go vet ./internal/signaling/` clean.
- [ ] `go test ./internal/signaling/ -cover` reports ≥ 80% (standards Gate-matrix).
- [ ] `internal/signaling` imports only stdlib + `internal/core` — no `internal/switcher`,
      no third-party (decision #7, dependency graph).

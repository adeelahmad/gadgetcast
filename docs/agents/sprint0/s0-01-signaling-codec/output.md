---
type: output
story: S0-01
---

## S0-01 · attempt 1 · red-worker · 2026-07-09T22:40:36Z
status: ok

### Summary
Wrote all 21 RED tests from plan.md at their exact `internal/signaling/<file>_test.go::TestName` across `codec_test.go` (12), `validate_test.go` (6), and `onair_test.go` (3), plus one `// agentic:shim`-marked `internal/signaling/shim.go` declaring the minimal referenced symbols (`MessageType` + type constants, `Envelope`, the 7 message structs, the `Message` interface, `Encode`, `Decode`, `DecodeError`, `ErrValidation`, per-message `Validate`, `OnAirStateFrom`) with deliberately WRONG bodies. `gofmt -l` clean, `go vet ./internal/signaling` exit 0, `go build ./...` exit 0. `go test ./internal/signaling` FAILS with 21/21 assertion failures — no compile errors, no panics, no trivial passes. Decode returns `(nil, nil)` so the never-panic table fails because no typed `*DecodeError` is returned (not because of a panic); Validate returns a generic non-sentinel error so BOTH the "should error" and "well-formed OK" checks fail. Uses `core.NewClientID` (ClientID is a struct, not a string type). Only `internal/signaling` + `internal/core` + stdlib touched.

### Result
| Check | Status | Detail |
|---|---|---|
| `codec_test.go::TestEncodeDecode_Join_RoundTrip` | FAIL | fails by assertion (got nil, want concrete JoinMessage) |
| `codec_test.go::TestEncodeDecode_Offer_RoundTrip` | FAIL | fails by assertion (got nil) |
| `codec_test.go::TestEncodeDecode_Answer_RoundTrip` | FAIL | fails by assertion (got nil) |
| `codec_test.go::TestEncodeDecode_IceCandidate_RoundTrip` | FAIL | fails by assertion (got nil) |
| `codec_test.go::TestEncodeDecode_SwitchSource_RoundTrip` | FAIL | fails by assertion (got nil) |
| `codec_test.go::TestEncodeDecode_OnAirState_RoundTrip` | FAIL | fails by assertion (got nil) |
| `codec_test.go::TestEncodeDecode_Error_RoundTrip` | FAIL | fails by assertion (got nil) |
| `codec_test.go::TestDecode_ReturnsConcreteStruct_NotMap` | FAIL | fails by assertion (not concrete OfferMessage) |
| `codec_test.go::TestDecode_UnknownType_ReturnsError` | FAIL | fails by assertion (no typed error) |
| `codec_test.go::TestDecode_TruncatedJSON_ReturnsError` | FAIL | fails by assertion (no typed error) |
| `codec_test.go::TestDecode_NeverPanics_MalformedTable` | FAIL | fails by assertion (no typed error; no panic) |
| `codec_test.go::TestDecode_TypePayloadMismatch_ReturnsError` | FAIL | fails by assertion (no typed error) |
| `validate_test.go::TestValidate_Offer_MissingSDP_Errors` | FAIL | fails by assertion (not ErrValidation) |
| `validate_test.go::TestValidate_Answer_MissingSDP_Errors` | FAIL | fails by assertion (not ErrValidation) |
| `validate_test.go::TestValidate_Join_MissingClientID_Errors` | FAIL | fails by assertion (not ErrValidation) |
| `validate_test.go::TestValidate_IceCandidate_MissingCandidate_Errors` | FAIL | fails by assertion (not ErrValidation) |
| `validate_test.go::TestValidate_SwitchSource_MissingTarget_Errors` | FAIL | fails by assertion (not ErrValidation) |
| `validate_test.go::TestValidate_WellFormedMessages_OK` | FAIL | fails by assertion (shim errors on well-formed) |
| `onair_test.go::TestOnAirStateFrom_Live` | FAIL | fails by assertion (no on-air id) |
| `onair_test.go::TestOnAirStateFrom_Nobody` | FAIL | fails by assertion (phantom live=true) |
| `onair_test.go::TestOnAirState_Projection_RoundTrip` | FAIL | fails by assertion (got nil) |

### Next
Scaffold the named production symbols (`MessageType`, `Envelope`, the 7 message structs, `Message`, `Encode`, `Decode`, `DecodeError`, `ErrValidation`, per-message `Validate`, `OnAirStateFrom`) as `panic("SUB-AGENT-TODO: ...")` stubs and delete `internal/signaling/shim.go`, then hand to GREEN to implement JSON envelope encode/decode with typed `*DecodeError` on unknown/malformed/mismatch input, `ErrValidation`-wrapped field validation, and the on-air projection.

## S0-01 · attempt 1 · scaffolder · 2026-07-09T22:45:53Z
status: ok

### Summary
Created the canonical production skeleton for `internal/signaling` split across four files: `message.go` (the `MessageType` discriminator + 7 type constants, the `Envelope` wire wrapper, the `Message` interface, the 7 message structs, and their panic-TODO `Type()` methods), `codec.go` (`Encode`, `Decode` returning `(Message, error)`, and the `DecodeError` type with panic-TODO `Error()`/`Unwrap()`), `validate.go` (the `ErrValidation` sentinel + 7 per-message panic-TODO `Validate()` methods), and `onair.go` (panic-TODO `OnAirStateFrom`). All type/const/struct/interface declarations are real decls; every function/method body is `panic("SUB-AGENT-TODO: <what GREEN implements>")` — no real logic. Deleted `internal/signaling/shim.go`; no `// agentic:shim` marker remains in-scope (the shims left in `internal/config`/`internal/switcher` belong to other stories, outside my scope glob). `core.ClientID` is used as the struct built via `core.NewClientID`; `Decode` keeps its `(Message, error)` interface shape. `gofmt -l` clean, `go build ./internal/signaling` exit 0, `go vet ./internal/signaling` exit 0. `go test` now panics on the stubs as expected.

### Result
| Check | Status | Detail |
|---|---|---|
| stubs panic+TODO | PASS | 4 funcs + 16 methods panic-TODO; 15 tracked symbols |
| each symbol once | PASS | no duplicate/parallel definitions |
| shim removed | PASS | internal/signaling/shim.go deleted; no in-scope agentic:shim marker |
| build compiles | PASS | go build ./internal/signaling exit 0 |
| vet clean | PASS | go vet ./internal/signaling exit 0 |
| gofmt clean | PASS | gofmt -l empty |

### Scaffold
+ create Encode @ internal/signaling/codec.go
+ create Decode @ internal/signaling/codec.go
+ create DecodeError @ internal/signaling/codec.go
+ create Envelope @ internal/signaling/message.go
+ create Message @ internal/signaling/message.go
+ create MessageType @ internal/signaling/message.go
+ create JoinMessage @ internal/signaling/message.go
+ create OfferMessage @ internal/signaling/message.go
+ create AnswerMessage @ internal/signaling/message.go
+ create IceCandidateMessage @ internal/signaling/message.go
+ create SwitchSourceMessage @ internal/signaling/message.go
+ create OnAirStateMessage @ internal/signaling/message.go
+ create ErrorMessage @ internal/signaling/message.go
+ create Validate (7 per-message methods) @ internal/signaling/validate.go
+ create ErrValidation @ internal/signaling/validate.go
+ create OnAirStateFrom @ internal/signaling/onair.go
+ delete shim.go @ internal/signaling/shim.go

### Next
green: fill the stubs — JSON envelope Encode/Decode with typed *DecodeError on unknown/malformed/mismatch input, ErrValidation-wrapped per-message Validate, and the OnAirStateFrom projection.

## S0-01 · attempt 1 · green-worker · 2026-07-10T09:05:00Z
status: ok
### Summary
Filled the four panic-TODO stub files with the minimum production code to pass all 21 RED tests. `Encode` projects each of the 7 messages onto a JSON `Envelope{Type, Payload}`, using string wire-payload structs because `core.ClientID` carries an unexported field and has no JSON marshaler. `Decode` unmarshals the envelope, dispatches on the `type` tag to the concrete struct, then runs `Validate`, returning a typed `*DecodeError` (wrapping context via `Unwrap`/`%w`) and a nil message on any unknown type, truncated/malformed JSON, or type/payload mismatch — never a panic. Per-message `Validate` methods wrap `ErrValidation` on missing SDP/client-id/candidate/target; `OnAirStateMessage`/`ErrorMessage` accept well-formed input. The 7 `Type()` methods return their constants and `OnAirStateFrom` builds `OnAirStateMessage{OnAir: id, Live: live}` with no phantom id when nobody is on air. Stdlib `encoding/json` only; no new tests, no suppression.
### Result
| Check | Status | Detail |
|---|---|---|
| `codec_test.go` (13 tests) | PASS | round-trips for all 7 types; unknown/truncated/malformed-table/mismatch → typed *DecodeError, nil message |
| `validate_test.go` (6 tests) | PASS | missing SDP/client/candidate/target wrap ErrValidation; 7 well-formed accepted |
| `onair_test.go` (3 tests) | PASS | live, nobody (no phantom id), projection round-trip |
| `go test ./internal/signaling` | PASS | 21/21 tests pass |
| coverage | PASS | 84.3% ≥ 80% |
| `gofmt -l` / `go vet` / `go build` | PASS | clean |
| `git diff --stat sprint0...HEAD` | PASS | only the 4 signaling production files |
### Next
structural review — confirm no orphan modules or parallel codec implementations across the signaling package.

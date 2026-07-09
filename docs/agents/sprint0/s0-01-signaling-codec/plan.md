---
type: plan
story: S0-01
scope: "tests only"
---

# Test plan — S0-01 signaling-codec (tests only)

One checkbox = one test; every bullet names a real `path::fn`. Covers all seven message
types (acceptance) plus the decode/validation edge and failure cases (including the
unknown-type-never-panics invariant).

## T2 — round-trip, one per message type (`internal/signaling/codec_test.go`)

- [ ] `internal/signaling/codec_test.go::TestEncodeDecode_Join_RoundTrip` — input: a `join` with `core.ClientID("cam-1")`; act: Encode then Decode; assert: decoded concrete `JoinMessage` equals the original, no loss.
- [ ] `internal/signaling/codec_test.go::TestEncodeDecode_Offer_RoundTrip` — input: an `offer` with a non-empty SDP; act: Encode→Decode; assert: concrete `OfferMessage` round-trips with SDP intact.
- [ ] `internal/signaling/codec_test.go::TestEncodeDecode_Answer_RoundTrip` — input: an `answer` with SDP; act: Encode→Decode; assert: `AnswerMessage` round-trips.
- [ ] `internal/signaling/codec_test.go::TestEncodeDecode_IceCandidate_RoundTrip` — input: an `ice-candidate` with a candidate string; act: Encode→Decode; assert: `IceCandidateMessage` round-trips.
- [ ] `internal/signaling/codec_test.go::TestEncodeDecode_SwitchSource_RoundTrip` — input: a `switch-source` targeting `core.ClientID("cam-2")`; act: Encode→Decode; assert: `SwitchSourceMessage` round-trips with target intact.
- [ ] `internal/signaling/codec_test.go::TestEncodeDecode_OnAirState_RoundTrip` — input: an `on-air-state` naming A live; act: Encode→Decode; assert: `OnAirStateMessage` round-trips.
- [ ] `internal/signaling/codec_test.go::TestEncodeDecode_Error_RoundTrip` — input: an `error` message with code+reason; act: Encode→Decode; assert: `ErrorMessage` round-trips.
- [ ] `internal/signaling/codec_test.go::TestDecode_ReturnsConcreteStruct_NotMap` — input: encoded `offer` bytes; act: Decode; assert: dynamic type is the concrete `OfferMessage` struct, not `map[string]interface{}`.

## T3 — decode error handling / never panic (`internal/signaling/codec_test.go`)

- [ ] `internal/signaling/codec_test.go::TestDecode_UnknownType_ReturnsError` — input: valid JSON with `"type":"bogus"`; act: Decode; assert: typed decode error, no panic, no zero-value message returned (invariant: unknown-type-decode-errors).
- [ ] `internal/signaling/codec_test.go::TestDecode_TruncatedJSON_ReturnsError` — input: `{"type":"offer","sd` (truncated); act: Decode; assert: typed decode error, no panic.
- [ ] `internal/signaling/codec_test.go::TestDecode_NeverPanics_MalformedTable` — input: a table of malformed payloads (empty bytes, `null`, `[]`, wrong-typed fields); act: Decode each under a recover-guard; assert: every entry returns a typed error and none panics.
- [ ] `internal/signaling/codec_test.go::TestDecode_TypePayloadMismatch_ReturnsError` — input: `"type":"offer"` but the payload carries only answer/none fields (missing SDP shape); act: Decode; assert: typed error, never a silently zero-valued `OfferMessage`.

## T4 — field validation (`internal/signaling/validate_test.go`)

- [ ] `internal/signaling/validate_test.go::TestValidate_Offer_MissingSDP_Errors` — input: an `offer` with empty SDP; act: `Validate()`; assert: typed validation error.
- [ ] `internal/signaling/validate_test.go::TestValidate_Answer_MissingSDP_Errors` — input: an `answer` with empty SDP; act: `Validate()`; assert: typed validation error.
- [ ] `internal/signaling/validate_test.go::TestValidate_Join_MissingClientID_Errors` — input: a `join` with the zero `core.ClientID`; act: `Validate()`; assert: typed validation error.
- [ ] `internal/signaling/validate_test.go::TestValidate_IceCandidate_MissingCandidate_Errors` — input: an `ice-candidate` with empty candidate; act: `Validate()`; assert: typed validation error.
- [ ] `internal/signaling/validate_test.go::TestValidate_SwitchSource_MissingTarget_Errors` — input: a `switch-source` with the zero target id; act: `Validate()`; assert: typed validation error.
- [ ] `internal/signaling/validate_test.go::TestValidate_WellFormedMessages_OK` — input: one well-formed instance of each of the seven types; act: `Validate()` each; assert: all return nil error.

## T4 — on-air projection (`internal/signaling/onair_test.go`)

- [ ] `internal/signaling/onair_test.go::TestOnAirStateFrom_Live` — input: `(core.ClientID("cam-1"), true)`; act: `OnAirStateFrom`; assert: message reports `cam-1` on-air and live=true.
- [ ] `internal/signaling/onair_test.go::TestOnAirStateFrom_Nobody` — input: `(zero ClientID, false)`; act: `OnAirStateFrom`; assert: message reports nobody on-air (live=false, no phantom id).
- [ ] `internal/signaling/onair_test.go::TestOnAirState_Projection_RoundTrip` — input: message from `OnAirStateFrom(A,true)`; act: Encode→Decode; assert: decoded `OnAirStateMessage` equals the projected one (contract stays lockstep with the switcher's `OnAir()` shape, decision #8).

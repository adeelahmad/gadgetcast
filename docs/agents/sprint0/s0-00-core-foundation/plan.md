---
type: plan
story: S0-00
scope: "tests only"
---

# Test plan — S0-00 core-foundation (tests only)

One checkbox = one test. Every bullet names a real `path::fn`. Covers T2's behavioral
acceptance and edge cases. (T1 is the `go mod init` setup task; its "test" is that
`go build ./...` compiles, which every other test transitively depends on — it has no
unit-test bullet by design.)

## internal/core — ClientID

- [ ] `internal/core/client_id_test.go::TestNewClientID_Valid` — input `"cam-1"`; act: `NewClientID("cam-1")`; assert: no error, returned `ClientID.String() == "cam-1"`, `IsValid()` true.
- [ ] `internal/core/client_id_test.go::TestNewClientID_EmptyRejected` — input `""`; act: `NewClientID("")`; assert: returns `ErrInvalidClientID` and the zero `ClientID`.
- [ ] `internal/core/client_id_test.go::TestNewClientID_WhitespaceRejected` — input `"   "`; act: `NewClientID("   ")`; assert: returns `ErrInvalidClientID` (all-whitespace id is not a valid identity).
- [ ] `internal/core/client_id_test.go::TestClientID_EqualityAndMapKey` — input: two IDs from the same string `"a"` and one from `"b"`; act: build all three, use as map keys; assert: the two `"a"` IDs compare `==` and index the same map slot, `"b"` indexes a different slot.
- [ ] `internal/core/client_id_test.go::TestClientID_ZeroValueInvalid` — input: the zero `ClientID{}`; act: call `IsValid()`/`IsZero()`; assert: `IsZero()` true and `IsValid()` false, so "no client" cannot masquerade as a real client.

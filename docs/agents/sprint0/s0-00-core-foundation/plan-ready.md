---
type: plan-ready
story: S0-00
from_red_at: 2026-07-09T22:10:44Z
---

# plan-ready — S0-00 core-foundation

RED verified on `sprint0` @ 9761fb1 (gate-red-verify green). Boxes tick ONLY on GREEN pass.

## internal/core — ClientID

- [x] `internal/core/client_id_test.go::TestNewClientID_Valid` — FAIL: `id.String()="" want "cam-1"; IsValid()=false want true`
- [x] `internal/core/client_id_test.go::TestNewClientID_EmptyRejected` — FAIL: `error=<nil>, want ErrInvalidClientID`
- [x] `internal/core/client_id_test.go::TestNewClientID_WhitespaceRejected` — FAIL: `error=<nil>, want ErrInvalidClientID`
- [x] `internal/core/client_id_test.go::TestClientID_EqualityAndMapKey` — FAIL: `map has 1 key, want 2`
- [x] `internal/core/client_id_test.go::TestClientID_ZeroValueInvalid` — FAIL: `ClientID{}.IsZero()=false, want true`

---
type: output
story: S1-03
---

## S1-03 · attempt 1 · red-worker · 2026-07-10T00:16:21Z
status: ok

### Summary
Wrote all 11 failing tests for `internal/drift` at their exact `path::fn` in
`internal/drift/drift_test.go`, plus one `// agentic:shim` file `internal/drift/shim.go`
declaring `Controller`, `New(opts ...Option)`, `Option` with `WithKp`/`WithTarget`/`WithClamp`,
`(*Controller).Update(fill, dt float64) float64`, and the documented default consts
(`DefaultKp` 0.5, `DefaultTarget` 0.5, `DefaultClamp` ±0.02, `DefaultDeadband` 0.01,
`DefaultTau` 2.0). The shim bodies are deliberately wrong — `New` ignores defaults/options
(zero-value struct) and `Update` returns `math.NaN()` (stdlib `math` only) — so every test
fails by assertion, never by a missing symbol. Assertions are written in `if !(good)` form so a
NaN result always reports failure. `gofmt` clean, `go vet ./internal/drift` clean, and
`go test ./internal/drift` FAILS with 11 assertion failures.

### Result
| Check | Status | Detail |
|---|---|---|
| `internal/drift/drift_test.go::TestNew_Defaults` | FAIL | fields 0 != documented defaults |
| `internal/drift/drift_test.go::TestNew_InjectableParams` | FAIL | injected params not applied |
| `internal/drift/drift_test.go::TestFlatAtTarget_RatioIsOne` | FAIL | NaN != 1.0 |
| `internal/drift/drift_test.go::TestWithinDeadband_RatioIsOne` | FAIL | NaN != 1.0 in dead-band |
| `internal/drift/drift_test.go::TestSustainedUpwardFill_RatioAboveOne` | FAIL | NaN not > 1.0 |
| `internal/drift/drift_test.go::TestSustainedDownwardFill_RatioBelowOne` | FAIL | NaN not < 1.0 |
| `internal/drift/drift_test.go::TestLargerError_LargerCorrection` | FAIL | NaN devs not ordered |
| `internal/drift/drift_test.go::TestClamp_UpperBound` | FAIL | NaN != 1+bound |
| `internal/drift/drift_test.go::TestClamp_LowerBound` | FAIL | NaN != 1-bound |
| `internal/drift/drift_test.go::TestSingleOutlier_StaysInNeutralBand` | FAIL | NaN not within neutral band |
| `internal/drift/drift_test.go::TestSustainedTrend_vs_Transient_Contrast` | FAIL | NaN devs not contrasted |
| `gofmt -l internal/drift/` | PASS | no diffs |
| `go vet ./internal/drift` | PASS | clean |

### Next
Scaffold the named symbols: `Controller`, `New`/`Option`/`WithKp`/`WithTarget`/`WithClamp`,
`(*Controller).Update` as `panic("SUB-AGENT-TODO")` stubs and delete the `// agentic:shim`
file, then GREEN implements the EMA-smoothed proportional control law (dead-band, clamp ±0.02)
so every assertion passes.

## S1-03 · attempt 1 · scaffolder · 2026-07-10T00:27:38Z
status: ok
### Summary
Created the canonical `internal/drift/drift.go` with real declarations for the control-law surface the RED tests reference: exported fields on `Controller` (Kp, Target, Clamp, Deadband, Tau), the `Option` functional-options type with real closure bodies for `WithKp`/`WithTarget`/`WithClamp`, and the documented default consts matching the RED expectations (DefaultKp 0.5, DefaultTarget 0.5, DefaultClamp 0.02, DefaultDeadband 0.01, DefaultTau 2.0). `New` and `(*Controller).Update` carry `panic("SUB-AGENT-TODO: ...")` bodies only — no logic. Deleted the RED `shim.go`; no `// agentic:shim` marker remains. gofmt clean, `go build ./internal/drift` and `go vet` both exit 0.
### Scaffold
+ create Controller @ internal/drift/drift.go
+ create Option @ internal/drift/drift.go
+ create WithKp @ internal/drift/drift.go
+ create WithTarget @ internal/drift/drift.go
+ create WithClamp @ internal/drift/drift.go
+ create New (panic+TODO) @ internal/drift/drift.go
+ create Update (panic+TODO) @ internal/drift/drift.go
+ create DefaultKp/DefaultTarget/DefaultClamp/DefaultDeadband/DefaultTau @ internal/drift/drift.go
- delete shim.go @ internal/drift/shim.go
### Result
| Check | Status | Detail |
|---|---|---|
| stubs panic+TODO | PASS | New, Update panic("SUB-AGENT-TODO: ...") |
| go build | PASS | exit 0 |
| go vet | PASS | exit 0 |
| shim removed | PASS | no // agentic:shim remains |
| single canonical decl | PASS | each symbol defined once |
### Next
green: fill the New/Update stubs so the RED tests pass

## S1-03 · attempt 1 · green-worker · 2026-07-10T00:31:43Z
status: ok
### Summary
Filled the `New`/`Update` panic-TODO bodies in the canonical stub `internal/drift/drift.go` with the bare-minimum EMA proportional control law. `New` seeds the `Default*` params, applies options, and primes the EMA state at `Target` so the first sample does not jolt. `Update` advances an EMA of the fill (`alpha = dt/(Tau+dt)`), computes `fillError = emaFill - Target`, returns `1.0` inside the dead-band, else applies `Kp*fillError` clamped to `[-Clamp,+Clamp]` and returns `1+clamped`. Tests are authoritative: the raw (non-normalized) `fillError` is what makes `TestWithinDeadband` (boundary at 0.01) and `TestLargerError` (0.52 vs 0.53 both stay inside the ±0.02 clamp with distinct corrections) pass — a `/target` normalization would clamp both and fail those two. Stdlib `math` only; no new tests, no suppression, no refactor.
### Result
| Check | Status | Detail |
|---|---|---|
| `drift_test.go::TestNew_Defaults` | PASS | green |
| `drift_test.go::TestNew_InjectableParams` | PASS | green |
| `drift_test.go::TestFlatAtTarget_RatioIsOne` | PASS | green |
| `drift_test.go::TestWithinDeadband_RatioIsOne` | PASS | green |
| `drift_test.go::TestSustainedUpwardFill_RatioAboveOne` | PASS | green |
| `drift_test.go::TestSustainedDownwardFill_RatioBelowOne` | PASS | green |
| `drift_test.go::TestLargerError_LargerCorrection` | PASS | green |
| `drift_test.go::TestClamp_UpperBound` | PASS | green |
| `drift_test.go::TestClamp_LowerBound` | PASS | green |
| `drift_test.go::TestSingleOutlier_StaysInNeutralBand` | PASS | green |
| `drift_test.go::TestSustainedTrend_vs_Transient_Contrast` | PASS | green |
| `gofmt -l internal/drift` | PASS | clean |
| `go vet ./internal/drift` | PASS | exit 0 |
| `go build ./internal/drift` | PASS | exit 0 |
| coverage | PASS | 100.0% (≥80%) |
### Next
structural review of internal/drift; then S1-03 final gate.

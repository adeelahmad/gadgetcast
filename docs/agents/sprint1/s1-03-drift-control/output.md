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

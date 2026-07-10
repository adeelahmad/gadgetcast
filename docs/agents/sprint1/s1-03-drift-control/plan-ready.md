---
type: plan-ready
story: S1-03
from_red_at: 2026-07-10T00:25:31Z
---

# plan-ready — S1-03 drift-control

RED verified on sprint1 (gate-red-verify green). Boxes tick ONLY on GREEN pass.

- [x] `internal/drift/drift_test.go::TestNew_Defaults` — input: `New()`; act: inspect the controller; assert: `Kp`, `target`, clamp bound (±0.02), dead-band, and `tau` equal the documented defaults (D8).
- [x] `internal/drift/drift_test.go::TestNew_InjectableParams` — input: `New(WithKp(...), WithTarget(...), WithClamp(...))`; act: inspect; assert: the injected values override the defaults (D8 injectable).
- [x] `internal/drift/drift_test.go::TestFlatAtTarget_RatioIsOne` — input: a controller at target; act: `Update(target, dt)` repeatedly; assert: `ratio == 1.0` (±eps) every sample (success #13).
- [x] `internal/drift/drift_test.go::TestWithinDeadband_RatioIsOne` — input: fill a hair above target but inside the dead-band; act: `Update`; assert: `ratio == 1.0` (dead-band zeroes tiny error) (success #13).
- [x] `internal/drift/drift_test.go::TestSustainedUpwardFill_RatioAboveOne` — input: a series sustained above target (buffer filling); act: `Update` across the window; assert: converged `ratio > 1.0` — drain faster (success #11; failure #12; sign convention).
- [x] `internal/drift/drift_test.go::TestSustainedDownwardFill_RatioBelowOne` — input: a series sustained below target (buffer draining); act: `Update` across the window; assert: converged `ratio < 1.0` (success #12; failure #13).
- [x] `internal/drift/drift_test.go::TestLargerError_LargerCorrection` — input: two sustained series, one further above target than the other; act: converge both; assert: the larger error yields a larger (correctly-signed, still-clamped) deviation from 1.0 (proportional response).
- [x] `internal/drift/drift_test.go::TestClamp_UpperBound` — input: fill far above target driving `Kp*fillError` past the bound; act: converge; assert: `ratio == 1 + bound` exactly (clamp upper honored).
- [x] `internal/drift/drift_test.go::TestClamp_LowerBound` — input: fill far below target (near zero); act: converge; assert: `ratio == 1 - bound` exactly (clamp lower honored).
- [x] `internal/drift/drift_test.go::TestSingleOutlier_StaysInNeutralBand` — input: a flat-at-target series with ONE outlier sample (tau >> dt); act: `Update` the outlier; assert: `ratio` stays within the neutral band (does not swing) — the EMA damps single-sample noise (failure #14; success #14).
- [x] `internal/drift/drift_test.go::TestSustainedTrend_vs_Transient_Contrast` — input: the same peak value delivered once (transient) vs. sustained over the window; act: compare resulting ratios; assert: the sustained trend moves the ratio materially while the single transient does not — noise ≠ drift (failure #14).

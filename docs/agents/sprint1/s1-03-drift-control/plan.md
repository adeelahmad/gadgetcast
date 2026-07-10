---
type: plan
story: S1-03
scope: "tests only"
---

# Test plan — S1-03 drift-control (tests only)

One checkbox = one test; every bullet names a real `path::fn`. Covers acceptance AND the
edge/failure cases the intent names: positive-drift → ratio > 1, negative-drift → ratio < 1,
flat → ratio 1.0, clamp bounds honored, single outlier does not swing the ratio. All tests
live in `internal/drift/drift_test.go`. Float assertions use a small epsilon.

## T1 — construction, defaults, neutral behavior

- [ ] `internal/drift/drift_test.go::TestNew_Defaults` — input: `New()`; act: inspect the controller; assert: `Kp`, `target`, clamp bound (±0.02), dead-band, and `tau` equal the documented defaults (D8).
- [ ] `internal/drift/drift_test.go::TestNew_InjectableParams` — input: `New(WithKp(...), WithTarget(...), WithClamp(...))`; act: inspect; assert: the injected values override the defaults (D8 injectable).
- [ ] `internal/drift/drift_test.go::TestFlatAtTarget_RatioIsOne` — input: a controller at target; act: `Update(target, dt)` repeatedly; assert: `ratio == 1.0` (±eps) every sample (success #13).
- [ ] `internal/drift/drift_test.go::TestWithinDeadband_RatioIsOne` — input: fill a hair above target but inside the dead-band; act: `Update`; assert: `ratio == 1.0` (dead-band zeroes tiny error) (success #13).

## T2 — proportional response + sign convention

- [ ] `internal/drift/drift_test.go::TestSustainedUpwardFill_RatioAboveOne` — input: a series sustained above target (buffer filling); act: `Update` across the window; assert: converged `ratio > 1.0` — drain faster (success #11; failure #12; sign convention).
- [ ] `internal/drift/drift_test.go::TestSustainedDownwardFill_RatioBelowOne` — input: a series sustained below target (buffer draining); act: `Update` across the window; assert: converged `ratio < 1.0` (success #12; failure #13).
- [ ] `internal/drift/drift_test.go::TestLargerError_LargerCorrection` — input: two sustained series, one further above target than the other; act: converge both; assert: the larger error yields a larger (correctly-signed, still-clamped) deviation from 1.0 (proportional response).

## T3 — clamp bounds + noise damping

- [ ] `internal/drift/drift_test.go::TestClamp_UpperBound` — input: fill far above target driving `Kp*fillError` past the bound; act: converge; assert: `ratio == 1 + bound` exactly (clamp upper honored).
- [ ] `internal/drift/drift_test.go::TestClamp_LowerBound` — input: fill far below target (near zero); act: converge; assert: `ratio == 1 - bound` exactly (clamp lower honored).
- [ ] `internal/drift/drift_test.go::TestSingleOutlier_StaysInNeutralBand` — input: a flat-at-target series with ONE outlier sample (tau >> dt); act: `Update` the outlier; assert: `ratio` stays within the neutral band (does not swing) — the EMA damps single-sample noise (failure #14; success #14).
- [ ] `internal/drift/drift_test.go::TestSustainedTrend_vs_Transient_Contrast` — input: the same peak value delivered once (transient) vs. sustained over the window; act: compare resulting ratios; assert: the sustained trend moves the ratio materially while the single transient does not — noise ≠ drift (failure #14).

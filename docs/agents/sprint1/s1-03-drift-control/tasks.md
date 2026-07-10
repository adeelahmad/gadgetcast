---
type: tasks
story: S1-03
---

# Tasks — S1-03 drift-control

Wave 1 (parallel). Package `internal/drift`. Stdlib `math` only. Pure math over a scripted
`(fillLevel, dt)` series — deterministic, no goroutines, no wall-clock. Decoupled from
`internal/jitter` (D7 — synthetic fill series).

Control law (D8): a PROPORTIONAL controller with EMA smoothing, dead-band, and clamp.
`Controller` holds injectable `Kp`, `target`, clamp bound (default ±0.02 = ±2%), dead-band,
and an EMA time constant `tau`; the EMA state `smoothed` starts at `target`. `Update(fill,
dt float64) float64` low-passes fill (`alpha = dt/(tau+dt)`; `smoothed += alpha*(fill -
smoothed)`), forms `fillError = (smoothed - target)/target`, zeroes it inside the dead-band,
and returns `ratio = clamp(1 + Kp*fillError, 1-bound, 1+bound)`. `dt` is used solely for the
EMA (D2 — no `Clock`). Constructor `New(opts...)` (functional options) with documented
defaults. Sign convention documented in code: `fill > target` (buffer filling) → `ratio > 1`
(drain faster); `fill < target` → `ratio < 1`.

## T1 — construction, defaults, neutral behavior

Define `Controller`, `New` (+ documented defaults for `Kp`, `target`, clamp, dead-band,
`tau`), and `Update`. A series flat at `target` returns `ratio == 1.0` exactly; a small
deviation within the dead-band returns `ratio == 1.0` (intake success #13). The sign
convention is documented in a package comment.

- Atomic: the type + constructor/options + the EMA/dead-band/clamp `Update` skeleton +
  neutral (at-target and within-dead-band) behavior.
- Testable: `drift_test.go` covers defaults, injectable-params, flat-at-target-ratio-one,
  and within-dead-band-ratio-one.

## T2 — proportional response + sign convention

A fill series sustained above `target` drives `ratio > 1.0` (drain faster); a series
sustained below `target` drives `ratio < 1.0`; a larger sustained error produces a larger
(correctly-signed) deviation from 1.0 (intake failures #12/#13; success #11/#12).

- Atomic: the proportional term applied to the smoothed error with the documented sign.
- Testable: `drift_test.go` covers sustained-upward-ratio-above-one, sustained-downward-
  ratio-below-one, and larger-error-larger-correction.

## T3 — clamp bounds + noise damping

The clamp caps the ratio: an extreme fill (far above `target`) returns exactly `1 + bound`;
an extreme low fill returns exactly `1 - bound` (clamp bounds honored). The EMA damps
single-sample noise: a single outlier amid an otherwise flat-at-target series does NOT swing
the ratio outside the neutral band, whereas a sustained trend does (intake failure #14;
success #14).

- Atomic: the clamp at both bounds + the EMA damping of a transient.
- Testable: `drift_test.go` covers clamp-upper-bound, clamp-lower-bound, single-outlier-
  stays-in-neutral-band, and sustained-trend-vs-transient-contrast.

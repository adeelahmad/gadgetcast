---
type: validate
story: S1-03
---

# Validate — S1-03 drift-control

PASS/FAIL rubric mirroring the task headings. Literal commands (PACKAGE-SCOPED per M-011),
literal expected output. Failure modes cite the rule each violates. Run from repo root.

## Pre-flight

- [ ] Worktree inherits merged base (M-010): `git reset --hard sprint1` before the first
      edit.
- [ ] `go test ./internal/drift/` compiles.
- [ ] Scoped commands only (M-011): use `./internal/drift/`, never `./...` — sibling
      packages `internal/session` / `internal/jitter` are panic-stubbed mid-wave.
- [ ] `internal/drift` imports only stdlib (`math`); no third-party and no `internal/jitter`:
      `go list -deps ./internal/drift/ | grep internal/jitter` prints nothing (D7 decoupling).

## T1 — construction, defaults, neutral behavior

| Check | Command | Expected | Fails rule |
|---|---|---|---|
| documented defaults | `go test ./internal/drift/ -run TestNew_Defaults -v` | `--- PASS` | D8 (documented defaults) |
| injectable params | `go test ./internal/drift/ -run TestNew_InjectableParams -v` | `--- PASS` | D8 (injectable Kp/target/clamp) |
| flat at target → 1.0 | `go test ./internal/drift/ -run TestFlatAtTarget_RatioIsOne -v` | `--- PASS` | intake success #13 |
| within dead-band → 1.0 | `go test ./internal/drift/ -run TestWithinDeadband_RatioIsOne -v` | `--- PASS` | intake success #13 (dead-band) |

## T2 — proportional response + sign convention

| Check | Command | Expected | Fails rule |
|---|---|---|---|
| upward → ratio > 1 | `go test ./internal/drift/ -run TestSustainedUpwardFill_RatioAboveOne -v` | `--- PASS` | intake failure #12 / success #11 (sign convention) |
| downward → ratio < 1 | `go test ./internal/drift/ -run TestSustainedDownwardFill_RatioBelowOne -v` | `--- PASS` | intake failure #13 / success #12 |
| larger error → larger correction | `go test ./internal/drift/ -run TestLargerError_LargerCorrection -v` | `--- PASS` | D8 (proportional response) |

FAIL if a sustained upward fill series yields `ratio ≤ 1.0` (intake failure #12 — buffer
grows toward overrun) or a sustained downward series fails to yield `ratio < 1.0` (intake
failure #13 — repeated underruns).

## T3 — clamp bounds + noise damping

| Check | Command | Expected | Fails rule |
|---|---|---|---|
| clamp upper | `go test ./internal/drift/ -run TestClamp_UpperBound -v` | `--- PASS` | D8 (clamp bound honored) |
| clamp lower | `go test ./internal/drift/ -run TestClamp_LowerBound -v` | `--- PASS` | D8 (clamp bound honored) |
| single outlier damped | `go test ./internal/drift/ -run TestSingleOutlier_StaysInNeutralBand -v` | `--- PASS` | intake failure #14 / success #14 |
| trend vs transient | `go test ./internal/drift/ -run TestSustainedTrend_vs_Transient_Contrast -v` | `--- PASS` | intake failure #14 (noise ≠ drift) |

FAIL if a single noisy/outlier sample amid a flat series produces a large immediate swing in
the ratio (intake failure #14 — audible artifacts in a real resampler), or if the ratio ever
exceeds the clamp bound.

## Final sign-off

- [ ] Every T1–T3 row PASS; `go test ./internal/drift/` exits 0.
- [ ] `gofmt -l internal/drift` empty; `go vet ./internal/drift/` clean.
- [ ] `go test ./internal/drift/ -cover` ≥ 80%.
- [ ] `internal/drift` imports only stdlib (`math`); no third-party, no `internal/jitter` (D7).
- [ ] The sign convention (positive fill drift → ratio > 1) is documented in a package/code
      comment.
- [ ] No goroutines/channels introduced (pure math) — `-race` stays advisory.

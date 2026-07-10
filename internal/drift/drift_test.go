package drift

import (
	"math"
	"testing"
)

const eps = 1e-9

func approx(a, b, tol float64) bool { return math.Abs(a-b) <= tol }

// converge runs the controller on a constant fill until the EMA settles and
// returns the final ratio.
func converge(c *Controller, fill, dt float64, n int) float64 {
	var r float64
	for i := 0; i < n; i++ {
		r = c.Update(fill, dt)
	}
	return r
}

// T1 — construction, defaults, neutral behavior

func TestNew_Defaults(t *testing.T) {
	c := New()
	if !approx(c.Kp, DefaultKp, eps) {
		t.Errorf("Kp = %v, want default %v", c.Kp, DefaultKp)
	}
	if !approx(c.Target, DefaultTarget, eps) {
		t.Errorf("Target = %v, want default %v", c.Target, DefaultTarget)
	}
	if !approx(c.Clamp, DefaultClamp, eps) {
		t.Errorf("Clamp = %v, want default %v", c.Clamp, DefaultClamp)
	}
	if !approx(c.Deadband, DefaultDeadband, eps) {
		t.Errorf("Deadband = %v, want default %v", c.Deadband, DefaultDeadband)
	}
	if !approx(c.Tau, DefaultTau, eps) {
		t.Errorf("Tau = %v, want default %v", c.Tau, DefaultTau)
	}
}

func TestNew_InjectableParams(t *testing.T) {
	const (
		kp     = 0.9
		target = 0.4
		bound  = 0.05
	)
	c := New(WithKp(kp), WithTarget(target), WithClamp(bound))
	if !approx(c.Kp, kp, eps) {
		t.Errorf("Kp = %v, want injected %v", c.Kp, kp)
	}
	if !approx(c.Target, target, eps) {
		t.Errorf("Target = %v, want injected %v", c.Target, target)
	}
	if !approx(c.Clamp, bound, eps) {
		t.Errorf("Clamp = %v, want injected %v", c.Clamp, bound)
	}
}

func TestFlatAtTarget_RatioIsOne(t *testing.T) {
	c := New()
	for i := 0; i < 200; i++ {
		r := c.Update(DefaultTarget, 0.1)
		if !approx(r, 1.0, eps) {
			t.Fatalf("sample %d: ratio = %v, want 1.0 at target", i, r)
		}
	}
}

func TestWithinDeadband_RatioIsOne(t *testing.T) {
	c := New()
	fill := DefaultTarget + DefaultDeadband/2 // a hair above target, inside the dead-band
	for i := 0; i < 200; i++ {
		r := c.Update(fill, 0.1)
		if !approx(r, 1.0, eps) {
			t.Fatalf("sample %d: ratio = %v, want 1.0 inside dead-band", i, r)
		}
	}
}

// T2 — proportional response + sign convention

func TestSustainedUpwardFill_RatioAboveOne(t *testing.T) {
	c := New()
	r := converge(c, 0.7, 0.1, 2000) // buffer sustained above target -> drain faster
	if !(r > 1.0+eps) {
		t.Errorf("converged ratio = %v, want > 1.0 for sustained upward fill", r)
	}
}

func TestSustainedDownwardFill_RatioBelowOne(t *testing.T) {
	c := New()
	r := converge(c, 0.3, 0.1, 2000) // buffer sustained below target -> drain slower
	if !(r < 1.0-eps) {
		t.Errorf("converged ratio = %v, want < 1.0 for sustained downward fill", r)
	}
}

func TestLargerError_LargerCorrection(t *testing.T) {
	// Both fills stay inside the clamp so the proportional response is visible.
	small := converge(New(), 0.52, 0.1, 5000)
	large := converge(New(), 0.53, 0.1, 5000)
	devSmall := math.Abs(small - 1.0)
	devLarge := math.Abs(large - 1.0)
	if !(devLarge > devSmall+eps) {
		t.Errorf("larger error must yield larger correction: devLarge=%v devSmall=%v", devLarge, devSmall)
	}
	// Sign convention: both are above target, so both ratios exceed 1.
	if !(small > 1.0+eps) || !(large > 1.0+eps) {
		t.Errorf("both corrections must be above 1.0: small=%v large=%v", small, large)
	}
}

// T3 — clamp bounds + noise damping

func TestClamp_UpperBound(t *testing.T) {
	c := New()
	r := converge(c, 1.0, 0.1, 5000) // far above target, drives Kp*error past the bound
	want := 1.0 + DefaultClamp
	if !approx(r, want, eps) {
		t.Errorf("ratio = %v, want clamped upper bound %v", r, want)
	}
}

func TestClamp_LowerBound(t *testing.T) {
	c := New()
	r := converge(c, 0.0, 0.1, 5000) // far below target, drives Kp*error past the bound
	want := 1.0 - DefaultClamp
	if !approx(r, want, eps) {
		t.Errorf("ratio = %v, want clamped lower bound %v", r, want)
	}
}

func TestSingleOutlier_StaysInNeutralBand(t *testing.T) {
	c := New()
	// Prime flat at target (tau >> dt so the EMA is well settled).
	for i := 0; i < 50; i++ {
		c.Update(DefaultTarget, 0.01)
	}
	r := c.Update(1.0, 0.01) // one outlier sample
	if !(math.Abs(r-1.0) < DefaultClamp) {
		t.Errorf("single outlier swung ratio to %v; EMA should keep |ratio-1| < %v", r, DefaultClamp)
	}
}

func TestSustainedTrend_vs_Transient_Contrast(t *testing.T) {
	const peak = 0.7
	// Transient: prime at target, then one peak sample (tau >> dt).
	ct := New()
	for i := 0; i < 50; i++ {
		ct.Update(DefaultTarget, 0.01)
	}
	transient := ct.Update(peak, 0.01)

	// Sustained: the same peak held across the whole window.
	sustained := converge(New(), peak, 0.01, 20000)

	devTransient := math.Abs(transient - 1.0)
	devSustained := math.Abs(sustained - 1.0)

	if !(devSustained > devTransient+eps) {
		t.Errorf("sustained trend must move ratio more than a transient: sustained=%v transient=%v", devSustained, devTransient)
	}
	if !(devSustained > DefaultClamp/2) {
		t.Errorf("sustained trend should move the ratio materially, dev=%v", devSustained)
	}
	if !(devTransient < DefaultClamp/2) {
		t.Errorf("single transient should not move the ratio materially, dev=%v", devTransient)
	}
}

// Package drift is a resampler drift controller: it turns an observed buffer
// fill fraction into a bounded resample ratio via a proportional control law
// smoothed by an exponential moving average.
package drift

// Documented default control-law parameters (D8).
const (
	// DefaultKp is the proportional gain applied to the smoothed fill error.
	DefaultKp = 0.5
	// DefaultTarget is the target buffer fill fraction the controller holds.
	DefaultTarget = 0.5
	// DefaultClamp is the symmetric bound on the resample-ratio deviation (±0.02).
	DefaultClamp = 0.02
	// DefaultDeadband zeroes corrections for fill errors smaller than this.
	DefaultDeadband = 0.01
	// DefaultTau is the EMA time constant (seconds) that damps single-sample noise.
	DefaultTau = 2.0
)

// Controller computes a resample ratio from the observed buffer fill fraction.
type Controller struct {
	Kp       float64
	Target   float64
	Clamp    float64
	Deadband float64
	Tau      float64
}

// Option configures a Controller at construction time.
type Option func(*Controller)

// WithKp overrides the proportional gain.
func WithKp(kp float64) Option { return func(c *Controller) { c.Kp = kp } }

// WithTarget overrides the target fill fraction.
func WithTarget(target float64) Option { return func(c *Controller) { c.Target = target } }

// WithClamp overrides the symmetric ratio-deviation bound.
func WithClamp(bound float64) Option { return func(c *Controller) { c.Clamp = bound } }

// New constructs a Controller seeded with the documented defaults, then applies
// any options.
func New(opts ...Option) *Controller {
	panic("SUB-AGENT-TODO: seed Controller with Default* params, apply opts, return it")
}

// Update returns the resample ratio for the latest fill sample over interval dt.
func (c *Controller) Update(fill, dt float64) float64 {
	panic("SUB-AGENT-TODO: EMA-smooth fill (Tau/dt), apply deadband, proportional Kp*error, clamp to 1±Clamp")
}

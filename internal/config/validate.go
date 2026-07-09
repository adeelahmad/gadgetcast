package config

// Validate enforces the hard preconditions: HTTPS must be enabled with cert+key,
// tailscale-only bind requires a tailscale-shaped (non-all-interfaces, non-empty)
// target, and a required PIN must have configured source material.
func (c Config) Validate() error {
	panic("SUB-AGENT-TODO: enforce HTTPS-on+cert/key (#5), tailscale-only non-all/non-empty target (#6), PIN-required has source (§4)")
}

package config

import (
	"errors"
	"fmt"
)

// Validate enforces the hard preconditions: HTTPS must be enabled with cert+key,
// tailscale-only bind requires a tailscale-shaped (non-all-interfaces, non-empty)
// target, and a required PIN must have configured source material.
func (c Config) Validate() error {
	if !c.HTTPS.Enabled {
		return errors.New("config: HTTPS must be enabled (hard precondition)")
	}
	if c.HTTPS.CertPath == "" || c.HTTPS.KeyPath == "" {
		return errors.New("config: HTTPS enabled requires cert_path and key_path")
	}
	if c.Bind.Mode == BindTailscaleOnly {
		if c.Bind.Target == "" || c.Bind.Target == "0.0.0.0" {
			return fmt.Errorf("config: bind=tailscale-only requires a tailscale-shaped target, got %q", c.Bind.Target)
		}
	}
	if c.PIN.Required && c.PIN.Source == "" {
		return errors.New("config: PIN required but no source configured")
	}
	return nil
}

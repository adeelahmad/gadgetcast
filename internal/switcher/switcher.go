// Package switcher owns the set of connected clients and the single on-air
// selection, gating on-air eligibility through a switcher-local PairingOracle.
package switcher

import "github.com/adeelahmad/gadgetcast/internal/core"

// PairingOracle reports whether a client has completed pairing. The switcher
// consumes pairing through this switcher-local interface (it does not import
// internal/pairing) so the on-air gate can be exercised with a stub.
type PairingOracle interface {
	IsPaired(id core.ClientID) bool
}

// Switcher owns the set of connected clients and the single on-air selection.
type Switcher struct {
	oracle PairingOracle
}

// New builds a Switcher backed by the given pairing oracle.
func New(oracle PairingOracle) *Switcher {
	panic("SUB-AGENT-TODO: construct Switcher with oracle and empty client set")
}

// AddClient registers a client in the connected set. Duplicate adds must return
// a typed error and leave the set unchanged.
func (s *Switcher) AddClient(id core.ClientID) error {
	panic("SUB-AGENT-TODO: record client membership; error on duplicate")
}

// RemoveClient deregisters a client. Unknown clients must return a typed error.
// Removing the on-air client transitions on-air to nobody.
func (s *Switcher) RemoveClient(id core.ClientID) error {
	panic("SUB-AGENT-TODO: drop client membership; error on unknown; clear on-air if it was this client")
}

// Has reports whether the client is in the connected set.
func (s *Switcher) Has(id core.ClientID) bool {
	panic("SUB-AGENT-TODO: report set membership")
}

// SetOnAir selects the sole on-air client, displacing any prior selection.
// The client must be known and paired, else a typed error is returned and
// on-air state is left unchanged.
func (s *Switcher) SetOnAir(id core.ClientID) error {
	panic("SUB-AGENT-TODO: gate on membership + oracle.IsPaired; set single on-air, displacing prior")
}

// OnAir returns the on-air client and whether one is live. It is the single
// source of truth: at most one client is live at any time.
func (s *Switcher) OnAir() (core.ClientID, bool) {
	panic("SUB-AGENT-TODO: return current on-air client and liveness")
}

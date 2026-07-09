// agentic:shim
// Minimal compile shim for the RED phase of S0-04 source-switcher.
// Bodies are DELIBERATELY WRONG so the package compiles and every test in
// switcher_test.go FAILS BY ASSERTION (not by a missing symbol). No real
// switching, membership, gating, or on-air logic lives here. Stdlib only.
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
	return &Switcher{oracle: oracle}
}

// AddClient registers a client. WRONG: never records membership.
func (s *Switcher) AddClient(id core.ClientID) error {
	_ = id
	return nil
}

// RemoveClient deregisters a client. WRONG: never errors, never mutates.
func (s *Switcher) RemoveClient(id core.ClientID) error {
	_ = id
	return nil
}

// Has reports set membership. WRONG: always denies membership.
func (s *Switcher) Has(id core.ClientID) bool {
	_ = id
	return false
}

// SetOnAir selects the sole on-air client. WRONG: never errors, no effect.
func (s *Switcher) SetOnAir(id core.ClientID) error {
	_ = id
	return nil
}

// OnAir returns the on-air client and whether one is live. WRONG: always
// claims the zero client is live, so both the default-nobody and the
// selected-identity assertions fail.
func (s *Switcher) OnAir() (core.ClientID, bool) {
	return core.ClientID{}, true
}

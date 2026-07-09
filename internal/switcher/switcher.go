// Package switcher owns the set of connected clients and the single on-air
// selection, gating on-air eligibility through a switcher-local PairingOracle.
package switcher

import (
	"errors"

	"github.com/adeelahmad/gadgetcast/internal/core"
)

// Typed errors returned by the switcher.
var (
	ErrDuplicateClient = errors.New("switcher: client already connected")
	ErrUnknownClient   = errors.New("switcher: client not connected")
	ErrNotPaired       = errors.New("switcher: client not paired")
)

// PairingOracle reports whether a client has completed pairing. The switcher
// consumes pairing through this switcher-local interface (it does not import
// internal/pairing) so the on-air gate can be exercised with a stub.
type PairingOracle interface {
	IsPaired(id core.ClientID) bool
}

// Switcher owns the set of connected clients and the single on-air selection.
type Switcher struct {
	oracle  PairingOracle
	clients map[core.ClientID]bool
	onAir   core.ClientID
	live    bool
}

// New builds a Switcher backed by the given pairing oracle.
func New(oracle PairingOracle) *Switcher {
	return &Switcher{
		oracle:  oracle,
		clients: make(map[core.ClientID]bool),
	}
}

// AddClient registers a client in the connected set. Duplicate adds must return
// a typed error and leave the set unchanged.
func (s *Switcher) AddClient(id core.ClientID) error {
	if s.clients[id] {
		return ErrDuplicateClient
	}
	s.clients[id] = true
	return nil
}

// RemoveClient deregisters a client. Unknown clients must return a typed error.
// Removing the on-air client transitions on-air to nobody.
func (s *Switcher) RemoveClient(id core.ClientID) error {
	if !s.clients[id] {
		return ErrUnknownClient
	}
	delete(s.clients, id)
	if s.live && s.onAir == id {
		s.onAir = core.ClientID{}
		s.live = false
	}
	return nil
}

// Has reports whether the client is in the connected set.
func (s *Switcher) Has(id core.ClientID) bool {
	return s.clients[id]
}

// SetOnAir selects the sole on-air client, displacing any prior selection.
// The client must be known and paired, else a typed error is returned and
// on-air state is left unchanged.
func (s *Switcher) SetOnAir(id core.ClientID) error {
	if !s.clients[id] {
		return ErrUnknownClient
	}
	if !s.oracle.IsPaired(id) {
		return ErrNotPaired
	}
	s.onAir = id
	s.live = true
	return nil
}

// OnAir returns the on-air client and whether one is live. It is the single
// source of truth: at most one client is live at any time.
func (s *Switcher) OnAir() (core.ClientID, bool) {
	return s.onAir, s.live
}

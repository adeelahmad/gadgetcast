// Package pairing tracks the pairing lifecycle of clients against an expected
// PIN, enforcing a configurable attempt limit and a durable lockout.
package pairing

import (
	"errors"

	"github.com/adeelahmad/gadgetcast/internal/core"
)

// State is a pairing lifecycle state. The zero value (StateUnknown) is not a
// real state: an unregistered client reports StateUnknown.
type State int

const (
	StateUnknown State = iota
	StatePending
	StatePaired
	StateRejected
)

// Typed errors the SubmitPIN contract returns.
var (
	ErrUnknownClient = errors.New("pairing: unknown client")
	ErrIncorrectPIN  = errors.New("pairing: incorrect pin")
	ErrRejected      = errors.New("pairing: client rejected")
)

// Manager tracks pairing state for a set of clients against an expected PIN.
type Manager struct {
	expectedPIN string
	limit       int
}

// New builds a Manager. The expected PIN is injected (never hardcoded here).
func New(expectedPIN string, limit int) *Manager {
	panic("SUB-AGENT-TODO: construct Manager with injected expectedPIN + attempt limit and per-client state store")
}

// Register creates a fresh pending record for a client.
func (m *Manager) Register(id core.ClientID) {
	panic("SUB-AGENT-TODO: create a fresh StatePending record for id (idempotent per client)")
}

// State reports a client's pairing state (StateUnknown if never registered).
func (m *Manager) State(id core.ClientID) State {
	panic("SUB-AGENT-TODO: return the client's current State, StateUnknown when unregistered")
}

// Attempts reports the number of failed PIN attempts for a client.
func (m *Manager) Attempts(id core.ClientID) int {
	panic("SUB-AGENT-TODO: return the count of failed PIN attempts for id")
}

// SubmitPIN submits a PIN for a registered client, enforcing the attempt limit
// and durable lockout. Returns ErrUnknownClient/ErrIncorrectPIN/ErrRejected.
func (m *Manager) SubmitPIN(id core.ClientID, pin string) error {
	panic("SUB-AGENT-TODO: unknown->ErrUnknownClient; rejected->ErrRejected; correct->StatePaired (idempotent, no attempt count); wrong->increment, StateRejected at limit else ErrIncorrectPIN; lockout is durable")
}

// IsPaired reports whether a client is paired.
func (m *Manager) IsPaired(id core.ClientID) bool {
	panic("SUB-AGENT-TODO: return true only when the client's State is StatePaired")
}

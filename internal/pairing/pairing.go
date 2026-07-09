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

type record struct {
	state    State
	attempts int
}

// Manager tracks pairing state for a set of clients against an expected PIN.
type Manager struct {
	expectedPIN string
	limit       int
	clients     map[core.ClientID]*record
}

// New builds a Manager. The expected PIN is injected (never hardcoded here).
func New(expectedPIN string, limit int) *Manager {
	return &Manager{
		expectedPIN: expectedPIN,
		limit:       limit,
		clients:     make(map[core.ClientID]*record),
	}
}

// Register creates a fresh pending record for a client.
func (m *Manager) Register(id core.ClientID) {
	m.clients[id] = &record{state: StatePending}
}

// State reports a client's pairing state (StateUnknown if never registered).
func (m *Manager) State(id core.ClientID) State {
	if r, ok := m.clients[id]; ok {
		return r.state
	}
	return StateUnknown
}

// Attempts reports the number of failed PIN attempts for a client.
func (m *Manager) Attempts(id core.ClientID) int {
	if r, ok := m.clients[id]; ok {
		return r.attempts
	}
	return 0
}

// SubmitPIN submits a PIN for a registered client, enforcing the attempt limit
// and durable lockout. Returns ErrUnknownClient/ErrIncorrectPIN/ErrRejected.
func (m *Manager) SubmitPIN(id core.ClientID, pin string) error {
	r, ok := m.clients[id]
	if !ok {
		return ErrUnknownClient
	}
	if r.state == StateRejected {
		return ErrRejected
	}
	if r.state == StatePaired {
		return nil
	}
	if pin == m.expectedPIN {
		r.state = StatePaired
		return nil
	}
	r.attempts++
	if r.attempts >= m.limit {
		r.state = StateRejected
	}
	return ErrIncorrectPIN
}

// IsPaired reports whether a client is paired.
func (m *Manager) IsPaired(id core.ClientID) bool {
	return m.State(id) == StatePaired
}

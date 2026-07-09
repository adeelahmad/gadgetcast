// agentic:shim
// RED compile shim only. Every symbol here is declared with a deliberately
// wrong body so the package COMPILES and every test in pairing_test.go FAILS
// BY ASSERTION (not by a missing symbol). The scaffolder replaces this file.
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

// String renders the state for readable test failures.
func (s State) String() string {
	switch s {
	case StatePending:
		return "pending"
	case StatePaired:
		return "paired"
	case StateRejected:
		return "rejected"
	default:
		return "unknown"
	}
}

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
	return &Manager{expectedPIN: expectedPIN, limit: limit}
}

// Register creates a fresh pending record for a client. (shim: no-op)
func (m *Manager) Register(id core.ClientID) {}

// State reports a client's pairing state. (shim: always StateUnknown)
func (m *Manager) State(id core.ClientID) State { return StateUnknown }

// Attempts reports the number of failed PIN attempts. (shim: always 0)
func (m *Manager) Attempts(id core.ClientID) int { return 0 }

// SubmitPIN submits a PIN for a registered client. (shim: always nil)
func (m *Manager) SubmitPIN(id core.ClientID, pin string) error { return nil }

// IsPaired reports whether a client is paired. (shim: always false)
func (m *Manager) IsPaired(id core.ClientID) bool { return false }

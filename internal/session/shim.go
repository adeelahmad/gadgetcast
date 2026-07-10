// agentic:shim
package session

import (
	"github.com/adeelahmad/gadgetcast/internal/core"
	"github.com/adeelahmad/gadgetcast/internal/pairing"
	"github.com/adeelahmad/gadgetcast/internal/signaling"
	"github.com/adeelahmad/gadgetcast/internal/switcher"
)

// Event is the closed set of inputs the Hub processes.
type Event interface {
	isEvent()
}

// Inbound is a raw signaling frame received from a connected client.
type Inbound struct {
	From core.ClientID
	Data []byte
}

// SubmitPIN is a client's PIN submission.
type SubmitPIN struct {
	Client core.ClientID
	PIN    string
}

// Disconnect signals that a client left the session.
type Disconnect struct {
	Client core.ClientID
}

func (Inbound) isEvent()    {}
func (SubmitPIN) isEvent()  {}
func (Disconnect) isEvent() {}

// Effect is a message the Hub wants delivered to a specific client.
type Effect struct {
	To  core.ClientID
	Msg signaling.Message
}

// Hub is the event-driven session coordinator over pairing and the switcher.
type Hub struct {
	pm *pairing.Manager
	sw *switcher.Switcher
}

// New builds a Hub over the given pairing manager and switcher.
func New(pm *pairing.Manager, sw *switcher.Switcher) *Hub {
	return &Hub{pm: pm, sw: sw}
}

// Process handles one event and returns the effects to deliver.
//
// SHIM: intentionally wrong — it mutates no injected state and always returns a
// single bogus non-error effect. Every behavioral test therefore fails BY
// ASSERTION (wrong effect list / wrong effect type / unchanged pm-sw state),
// never by a missing symbol and never by trivially passing a "no effects" case.
func (h *Hub) Process(ev Event) []Effect {
	return []Effect{{To: core.ClientID{}, Msg: signaling.JoinMessage{}}}
}

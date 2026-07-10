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
	pm     *pairing.Manager
	sw     *switcher.Switcher
	roster []core.ClientID
}

// New builds a Hub over the given pairing manager and switcher.
func New(pm *pairing.Manager, sw *switcher.Switcher) *Hub {
	return &Hub{pm: pm, sw: sw}
}

// Process handles one event and returns the effects to deliver.
func (h *Hub) Process(ev Event) []Effect {
	switch e := ev.(type) {
	case Inbound:
		return h.handleInbound(e)
	case SubmitPIN:
		return h.handleSubmitPIN(e)
	case Disconnect:
		return h.handleDisconnect(e)
	}
	return nil
}

func errorEffect(to core.ClientID, reason string) []Effect {
	return []Effect{{To: to, Msg: signaling.ErrorMessage{Reason: reason}}}
}

func (h *Hub) handleInbound(e Inbound) []Effect {
	msg, err := signaling.Decode(e.Data)
	if err != nil {
		return errorEffect(e.From, err.Error())
	}

	switch m := msg.(type) {
	case signaling.JoinMessage:
		if err := h.sw.AddClient(m.Client); err != nil {
			return errorEffect(e.From, err.Error())
		}
		h.pm.Register(m.Client)
		h.roster = append(h.roster, m.Client)
		return nil
	case signaling.SwitchSourceMessage:
		if err := h.sw.SetOnAir(m.Target); err != nil {
			return errorEffect(e.From, err.Error())
		}
		return h.broadcast(signaling.OnAirStateFrom(h.sw.OnAir()))
	case signaling.OfferMessage, signaling.AnswerMessage, signaling.IceCandidateMessage:
		return h.relay(e.From, msg)
	default:
		return errorEffect(e.From, "unexpected server message")
	}
}

func (h *Hub) handleSubmitPIN(e SubmitPIN) []Effect {
	if err := h.pm.SubmitPIN(e.Client, e.PIN); err != nil {
		return errorEffect(e.Client, err.Error())
	}
	return nil
}

func (h *Hub) handleDisconnect(e Disconnect) []Effect {
	idx := -1
	for i, c := range h.roster {
		if c == e.Client {
			idx = i
			break
		}
	}
	if idx == -1 {
		return nil
	}

	onAir, live := h.sw.OnAir()
	wasOnAir := live && onAir == e.Client
	h.sw.RemoveClient(e.Client)
	h.roster = append(h.roster[:idx], h.roster[idx+1:]...)

	if !wasOnAir {
		return nil
	}
	return h.broadcast(signaling.OnAirStateFrom(core.ClientID{}, false))
}

// broadcast addresses one copy of msg to every roster client, in roster order.
func (h *Hub) broadcast(msg signaling.Message) []Effect {
	effects := make([]Effect, 0, len(h.roster))
	for _, c := range h.roster {
		effects = append(effects, Effect{To: c, Msg: msg})
	}
	return effects
}

// relay forwards msg verbatim to every roster client other than the sender.
func (h *Hub) relay(from core.ClientID, msg signaling.Message) []Effect {
	effects := make([]Effect, 0, len(h.roster))
	for _, c := range h.roster {
		if c == from {
			continue
		}
		effects = append(effects, Effect{To: c, Msg: msg})
	}
	return effects
}

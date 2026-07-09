package signaling

import (
	"encoding/json"

	"github.com/adeelahmad/gadgetcast/internal/core"
)

// MessageType is the discriminator carried on the wire envelope.
type MessageType string

const (
	TypeJoin         MessageType = "join"
	TypeOffer        MessageType = "offer"
	TypeAnswer       MessageType = "answer"
	TypeIceCandidate MessageType = "ice-candidate"
	TypeSwitchSource MessageType = "switch-source"
	TypeOnAirState   MessageType = "on-air-state"
	TypeError        MessageType = "error"
)

// Envelope is the wire wrapper: a type tag plus a raw payload.
type Envelope struct {
	Type    MessageType     `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// Message is the closed set of signaling messages.
type Message interface {
	Type() MessageType
	Validate() error
}

// JoinMessage announces a client joining the session.
type JoinMessage struct {
	Client core.ClientID
}

// OfferMessage carries a WebRTC SDP offer.
type OfferMessage struct {
	SDP string
}

// AnswerMessage carries a WebRTC SDP answer.
type AnswerMessage struct {
	SDP string
}

// IceCandidateMessage carries a single ICE candidate.
type IceCandidateMessage struct {
	Candidate string
}

// SwitchSourceMessage requests switching the live source to Target.
type SwitchSourceMessage struct {
	Target core.ClientID
}

// OnAirStateMessage projects the current on-air state.
type OnAirStateMessage struct {
	OnAir core.ClientID
	Live  bool
}

// ErrorMessage carries a coded failure reason.
type ErrorMessage struct {
	Code   string
	Reason string
}

func (JoinMessage) Type() MessageType { return TypeJoin }

func (OfferMessage) Type() MessageType { return TypeOffer }

func (AnswerMessage) Type() MessageType { return TypeAnswer }

func (IceCandidateMessage) Type() MessageType { return TypeIceCandidate }

func (SwitchSourceMessage) Type() MessageType { return TypeSwitchSource }

func (OnAirStateMessage) Type() MessageType { return TypeOnAirState }

func (ErrorMessage) Type() MessageType { return TypeError }

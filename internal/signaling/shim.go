// agentic:shim
// Minimal compile shims for the S0-01 RED tests. Every body here is deliberately
// WRONG so the package compiles and every test FAILS BY ASSERTION (never a panic,
// never a pass). The scaffolder replaces this file; do not implement real logic.
package signaling

import (
	"encoding/json"
	"errors"

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

// ErrValidation is the sentinel every message's Validate wraps on bad input.
var ErrValidation = errors.New("signaling: validation error")

// Envelope is the wire wrapper: a type tag plus a raw payload.
type Envelope struct {
	Type    MessageType     `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// DecodeError is the typed error Decode returns for any malformed / unknown input.
type DecodeError struct {
	Type   MessageType
	Reason string
	Err    error
}

func (e *DecodeError) Error() string { return "signaling: decode error: " + e.Reason }
func (e *DecodeError) Unwrap() error { return e.Err }

// Message is the closed set of signaling messages.
type Message interface {
	Type() MessageType
	Validate() error
}

type JoinMessage struct {
	Client core.ClientID
}

type OfferMessage struct {
	SDP string
}

type AnswerMessage struct {
	SDP string
}

type IceCandidateMessage struct {
	Candidate string
}

type SwitchSourceMessage struct {
	Target core.ClientID
}

type OnAirStateMessage struct {
	OnAir core.ClientID
	Live  bool
}

type ErrorMessage struct {
	Code   string
	Reason string
}

func (JoinMessage) Type() MessageType         { return TypeJoin }
func (OfferMessage) Type() MessageType        { return TypeOffer }
func (AnswerMessage) Type() MessageType       { return TypeAnswer }
func (IceCandidateMessage) Type() MessageType { return TypeIceCandidate }
func (SwitchSourceMessage) Type() MessageType { return TypeSwitchSource }
func (OnAirStateMessage) Type() MessageType   { return TypeOnAirState }
func (ErrorMessage) Type() MessageType        { return TypeError }

// shimErr is a generic (untyped-relative-to-ErrValidation) error so that BOTH the
// "should error" validation tests AND the "well-formed is OK" test fail: bad-input
// tests demand errors.Is(err, ErrValidation) (false here), and the OK test demands
// nil (non-nil here).
func shimErr() error { return errors.New("agentic:shim Validate unimplemented") }

func (JoinMessage) Validate() error         { return shimErr() }
func (OfferMessage) Validate() error        { return shimErr() }
func (AnswerMessage) Validate() error       { return shimErr() }
func (IceCandidateMessage) Validate() error { return shimErr() }
func (SwitchSourceMessage) Validate() error { return shimErr() }
func (OnAirStateMessage) Validate() error   { return shimErr() }
func (ErrorMessage) Validate() error        { return shimErr() }

// Encode WRONGLY drops the payload so every round-trip mismatches.
func Encode(msg Message) ([]byte, error) {
	_ = msg
	return []byte("{}"), nil
}

// Decode WRONGLY returns a nil message and a nil error: round-trip tests fail on the
// concrete-type assertion, and the decode-error / never-panic tests fail because no
// typed *DecodeError is returned (not because of a panic).
func Decode(data []byte) (Message, error) {
	_ = data
	return nil, nil
}

// OnAirStateFrom WRONGLY reports live with no on-air id, so both projection tests fail.
func OnAirStateFrom(id core.ClientID, live bool) OnAirStateMessage {
	_ = id
	_ = live
	return OnAirStateMessage{Live: true}
}

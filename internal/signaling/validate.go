package signaling

import "errors"

// ErrValidation is the sentinel every message's Validate wraps on bad input.
var ErrValidation = errors.New("signaling: validation error")

func (m JoinMessage) Validate() error {
	panic("SUB-AGENT-TODO: GREEN rejects the zero Client with an ErrValidation-wrapped error")
}

func (m OfferMessage) Validate() error {
	panic("SUB-AGENT-TODO: GREEN rejects an empty SDP with an ErrValidation-wrapped error")
}

func (m AnswerMessage) Validate() error {
	panic("SUB-AGENT-TODO: GREEN rejects an empty SDP with an ErrValidation-wrapped error")
}

func (m IceCandidateMessage) Validate() error {
	panic("SUB-AGENT-TODO: GREEN rejects an empty Candidate with an ErrValidation-wrapped error")
}

func (m SwitchSourceMessage) Validate() error {
	panic("SUB-AGENT-TODO: GREEN rejects the zero Target with an ErrValidation-wrapped error")
}

func (m OnAirStateMessage) Validate() error {
	panic("SUB-AGENT-TODO: GREEN accepts a well-formed on-air-state message")
}

func (m ErrorMessage) Validate() error {
	panic("SUB-AGENT-TODO: GREEN accepts a well-formed error message")
}

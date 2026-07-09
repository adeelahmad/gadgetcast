package signaling

import (
	"errors"
	"fmt"
)

// ErrValidation is the sentinel every message's Validate wraps on bad input.
var ErrValidation = errors.New("signaling: validation error")

func (m JoinMessage) Validate() error {
	if m.Client.IsZero() {
		return fmt.Errorf("join: missing client id: %w", ErrValidation)
	}
	return nil
}

func (m OfferMessage) Validate() error {
	if m.SDP == "" {
		return fmt.Errorf("offer: missing sdp: %w", ErrValidation)
	}
	return nil
}

func (m AnswerMessage) Validate() error {
	if m.SDP == "" {
		return fmt.Errorf("answer: missing sdp: %w", ErrValidation)
	}
	return nil
}

func (m IceCandidateMessage) Validate() error {
	if m.Candidate == "" {
		return fmt.Errorf("ice-candidate: missing candidate: %w", ErrValidation)
	}
	return nil
}

func (m SwitchSourceMessage) Validate() error {
	if m.Target.IsZero() {
		return fmt.Errorf("switch-source: missing target: %w", ErrValidation)
	}
	return nil
}

func (m OnAirStateMessage) Validate() error {
	return nil
}

func (m ErrorMessage) Validate() error {
	return nil
}

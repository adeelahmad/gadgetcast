package signaling

import (
	"errors"
	"testing"

	"github.com/adeelahmad/gadgetcast/internal/core"
)

func TestValidate_Offer_MissingSDP_Errors(t *testing.T) {
	err := OfferMessage{SDP: ""}.Validate()
	if err == nil {
		t.Fatalf("expected a validation error for an offer with empty SDP, got nil")
	}
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected err to wrap ErrValidation, got %v", err)
	}
}

func TestValidate_Answer_MissingSDP_Errors(t *testing.T) {
	err := AnswerMessage{SDP: ""}.Validate()
	if err == nil {
		t.Fatalf("expected a validation error for an answer with empty SDP, got nil")
	}
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected err to wrap ErrValidation, got %v", err)
	}
}

func TestValidate_Join_MissingClientID_Errors(t *testing.T) {
	err := JoinMessage{Client: core.ClientID{}}.Validate()
	if err == nil {
		t.Fatalf("expected a validation error for a join with the zero ClientID, got nil")
	}
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected err to wrap ErrValidation, got %v", err)
	}
}

func TestValidate_IceCandidate_MissingCandidate_Errors(t *testing.T) {
	err := IceCandidateMessage{Candidate: ""}.Validate()
	if err == nil {
		t.Fatalf("expected a validation error for an ice-candidate with empty candidate, got nil")
	}
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected err to wrap ErrValidation, got %v", err)
	}
}

func TestValidate_SwitchSource_MissingTarget_Errors(t *testing.T) {
	err := SwitchSourceMessage{Target: core.ClientID{}}.Validate()
	if err == nil {
		t.Fatalf("expected a validation error for a switch-source with the zero target, got nil")
	}
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("expected err to wrap ErrValidation, got %v", err)
	}
}

func TestValidate_WellFormedMessages_OK(t *testing.T) {
	msgs := []Message{
		JoinMessage{Client: mustID(t, "cam-1")},
		OfferMessage{SDP: "v=0\r\no=- 1 1 IN IP4 0.0.0.0\r\n"},
		AnswerMessage{SDP: "v=0\r\no=- 2 2 IN IP4 0.0.0.0\r\n"},
		IceCandidateMessage{Candidate: "candidate:1 1 udp 2130706431 192.0.2.1 54321 typ host"},
		SwitchSourceMessage{Target: mustID(t, "cam-2")},
		OnAirStateFrom(mustID(t, "cam-1"), true),
		ErrorMessage{Code: "E_PAIR_DENIED", Reason: "pairing pin rejected"},
	}

	for _, m := range msgs {
		if err := m.Validate(); err != nil {
			t.Fatalf("%T is well-formed but Validate returned: %v", m, err)
		}
	}
}

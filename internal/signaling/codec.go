package signaling

import (
	"encoding/json"
	"fmt"

	"github.com/adeelahmad/gadgetcast/internal/core"
)

// DecodeError is the typed error Decode returns for any malformed / unknown input.
type DecodeError struct {
	Type   MessageType
	Reason string
	Err    error
}

func (e *DecodeError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("signaling: decode %q: %s: %v", e.Type, e.Reason, e.Err)
	}
	return fmt.Sprintf("signaling: decode %q: %s", e.Type, e.Reason)
}

func (e *DecodeError) Unwrap() error { return e.Err }

// Wire payload shapes: core.ClientID has an unexported field and no JSON
// marshaler, so client-bearing messages are projected onto string wire forms.
type joinWire struct {
	Client string `json:"client"`
}

type sdpWire struct {
	SDP string `json:"sdp"`
}

type iceWire struct {
	Candidate string `json:"candidate"`
}

type switchWire struct {
	Target string `json:"target"`
}

type onAirWire struct {
	OnAir string `json:"onair"`
	Live  bool   `json:"live"`
}

type errorWire struct {
	Code   string `json:"code"`
	Reason string `json:"reason"`
}

// Encode serializes a Message into a typed JSON envelope.
func Encode(msg Message) ([]byte, error) {
	var wire any
	switch m := msg.(type) {
	case JoinMessage:
		wire = joinWire{Client: m.Client.String()}
	case OfferMessage:
		wire = sdpWire{SDP: m.SDP}
	case AnswerMessage:
		wire = sdpWire{SDP: m.SDP}
	case IceCandidateMessage:
		wire = iceWire{Candidate: m.Candidate}
	case SwitchSourceMessage:
		wire = switchWire{Target: m.Target.String()}
	case OnAirStateMessage:
		wire = onAirWire{OnAir: m.OnAir.String(), Live: m.Live}
	case ErrorMessage:
		wire = errorWire{Code: m.Code, Reason: m.Reason}
	default:
		return nil, fmt.Errorf("signaling: cannot encode unknown message %T", msg)
	}

	payload, err := json.Marshal(wire)
	if err != nil {
		return nil, fmt.Errorf("signaling: marshal payload: %w", err)
	}
	return json.Marshal(Envelope{Type: msg.Type(), Payload: payload})
}

// Decode parses a JSON envelope back into its concrete Message, returning a typed
// *DecodeError for unknown types, truncated/malformed JSON, or type/payload mismatch.
func Decode(data []byte) (Message, error) {
	var env Envelope
	if err := json.Unmarshal(data, &env); err != nil {
		return nil, &DecodeError{Reason: "malformed envelope", Err: err}
	}

	msg, err := decodePayload(env.Type, env.Payload)
	if err != nil {
		return nil, err
	}

	if err := msg.Validate(); err != nil {
		return nil, &DecodeError{Type: env.Type, Reason: "payload failed validation", Err: err}
	}
	return msg, nil
}

func decodePayload(typ MessageType, payload json.RawMessage) (Message, error) {
	unmarshal := func(v any) error {
		if err := json.Unmarshal(payload, v); err != nil {
			return &DecodeError{Type: typ, Reason: "malformed payload", Err: err}
		}
		return nil
	}

	switch typ {
	case TypeJoin:
		var w joinWire
		if err := unmarshal(&w); err != nil {
			return nil, err
		}
		id, _ := core.NewClientID(w.Client)
		return JoinMessage{Client: id}, nil
	case TypeOffer:
		var w sdpWire
		if err := unmarshal(&w); err != nil {
			return nil, err
		}
		return OfferMessage{SDP: w.SDP}, nil
	case TypeAnswer:
		var w sdpWire
		if err := unmarshal(&w); err != nil {
			return nil, err
		}
		return AnswerMessage{SDP: w.SDP}, nil
	case TypeIceCandidate:
		var w iceWire
		if err := unmarshal(&w); err != nil {
			return nil, err
		}
		return IceCandidateMessage{Candidate: w.Candidate}, nil
	case TypeSwitchSource:
		var w switchWire
		if err := unmarshal(&w); err != nil {
			return nil, err
		}
		id, _ := core.NewClientID(w.Target)
		return SwitchSourceMessage{Target: id}, nil
	case TypeOnAirState:
		var w onAirWire
		if err := unmarshal(&w); err != nil {
			return nil, err
		}
		id, _ := core.NewClientID(w.OnAir)
		return OnAirStateMessage{OnAir: id, Live: w.Live}, nil
	case TypeError:
		var w errorWire
		if err := unmarshal(&w); err != nil {
			return nil, err
		}
		return ErrorMessage{Code: w.Code, Reason: w.Reason}, nil
	default:
		return nil, &DecodeError{Type: typ, Reason: "unknown message type", Err: fmt.Errorf("%w: %q", ErrValidation, typ)}
	}
}

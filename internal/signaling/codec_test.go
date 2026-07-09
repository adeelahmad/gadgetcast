package signaling

import (
	"errors"
	"testing"

	"github.com/adeelahmad/gadgetcast/internal/core"
)

// mustID builds a valid core.ClientID or fails the test. Shared across the
// signaling test files (same package).
func mustID(t *testing.T, raw string) core.ClientID {
	t.Helper()
	id, err := core.NewClientID(raw)
	if err != nil {
		t.Fatalf("NewClientID(%q): %v", raw, err)
	}
	return id
}

func TestEncodeDecode_Join_RoundTrip(t *testing.T) {
	original := JoinMessage{Client: mustID(t, "cam-1")}

	data, err := Encode(original)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	decoded, err := Decode(data)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}

	got, ok := decoded.(JoinMessage)
	if !ok {
		t.Fatalf("expected concrete JoinMessage, got %T", decoded)
	}
	if got != original {
		t.Fatalf("join round-trip mismatch: got %+v want %+v", got, original)
	}
}

func TestEncodeDecode_Offer_RoundTrip(t *testing.T) {
	original := OfferMessage{SDP: "v=0\r\no=- 1 1 IN IP4 0.0.0.0\r\n"}

	data, err := Encode(original)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	decoded, err := Decode(data)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}

	got, ok := decoded.(OfferMessage)
	if !ok {
		t.Fatalf("expected concrete OfferMessage, got %T", decoded)
	}
	if got != original {
		t.Fatalf("offer round-trip mismatch: got %+v want %+v", got, original)
	}
}

func TestEncodeDecode_Answer_RoundTrip(t *testing.T) {
	original := AnswerMessage{SDP: "v=0\r\no=- 2 2 IN IP4 0.0.0.0\r\n"}

	data, err := Encode(original)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	decoded, err := Decode(data)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}

	got, ok := decoded.(AnswerMessage)
	if !ok {
		t.Fatalf("expected concrete AnswerMessage, got %T", decoded)
	}
	if got != original {
		t.Fatalf("answer round-trip mismatch: got %+v want %+v", got, original)
	}
}

func TestEncodeDecode_IceCandidate_RoundTrip(t *testing.T) {
	original := IceCandidateMessage{Candidate: "candidate:1 1 udp 2130706431 192.0.2.1 54321 typ host"}

	data, err := Encode(original)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	decoded, err := Decode(data)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}

	got, ok := decoded.(IceCandidateMessage)
	if !ok {
		t.Fatalf("expected concrete IceCandidateMessage, got %T", decoded)
	}
	if got != original {
		t.Fatalf("ice-candidate round-trip mismatch: got %+v want %+v", got, original)
	}
}

func TestEncodeDecode_SwitchSource_RoundTrip(t *testing.T) {
	original := SwitchSourceMessage{Target: mustID(t, "cam-2")}

	data, err := Encode(original)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	decoded, err := Decode(data)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}

	got, ok := decoded.(SwitchSourceMessage)
	if !ok {
		t.Fatalf("expected concrete SwitchSourceMessage, got %T", decoded)
	}
	if got != original {
		t.Fatalf("switch-source round-trip mismatch: got %+v want %+v", got, original)
	}
}

func TestEncodeDecode_OnAirState_RoundTrip(t *testing.T) {
	original := OnAirStateMessage{OnAir: mustID(t, "cam-A"), Live: true}

	data, err := Encode(original)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	decoded, err := Decode(data)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}

	got, ok := decoded.(OnAirStateMessage)
	if !ok {
		t.Fatalf("expected concrete OnAirStateMessage, got %T", decoded)
	}
	if got != original {
		t.Fatalf("on-air-state round-trip mismatch: got %+v want %+v", got, original)
	}
}

func TestEncodeDecode_Error_RoundTrip(t *testing.T) {
	original := ErrorMessage{Code: "E_PAIR_DENIED", Reason: "pairing pin rejected"}

	data, err := Encode(original)
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}
	decoded, err := Decode(data)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}

	got, ok := decoded.(ErrorMessage)
	if !ok {
		t.Fatalf("expected concrete ErrorMessage, got %T", decoded)
	}
	if got != original {
		t.Fatalf("error round-trip mismatch: got %+v want %+v", got, original)
	}
}

func TestDecode_ReturnsConcreteStruct_NotMap(t *testing.T) {
	data, err := Encode(OfferMessage{SDP: "v=0\r\n"})
	if err != nil {
		t.Fatalf("Encode: %v", err)
	}

	decoded, err := Decode(data)
	if err != nil {
		t.Fatalf("Decode: %v", err)
	}

	var raw interface{} = decoded
	if _, isMap := raw.(map[string]interface{}); isMap {
		t.Fatalf("Decode returned a map[string]interface{}, want a concrete OfferMessage")
	}
	if _, ok := decoded.(OfferMessage); !ok {
		t.Fatalf("expected concrete OfferMessage, got %T", decoded)
	}
}

func TestDecode_UnknownType_ReturnsError(t *testing.T) {
	decoded, err := Decode([]byte(`{"type":"bogus","payload":{}}`))
	if err == nil {
		t.Fatalf("expected a decode error for unknown type, got nil")
	}

	var de *DecodeError
	if !errors.As(err, &de) {
		t.Fatalf("expected a typed *DecodeError, got %T: %v", err, err)
	}
	if decoded != nil {
		t.Fatalf("expected nil message on unknown-type error, got %+v", decoded)
	}
}

func TestDecode_TruncatedJSON_ReturnsError(t *testing.T) {
	decoded, err := Decode([]byte(`{"type":"offer","sd`))
	if err == nil {
		t.Fatalf("expected a decode error for truncated JSON, got nil")
	}

	var de *DecodeError
	if !errors.As(err, &de) {
		t.Fatalf("expected a typed *DecodeError, got %T: %v", err, err)
	}
	if decoded != nil {
		t.Fatalf("expected nil message on truncated-JSON error, got %+v", decoded)
	}
}

func TestDecode_NeverPanics_MalformedTable(t *testing.T) {
	cases := []struct {
		name string
		data []byte
	}{
		{"empty", []byte(``)},
		{"null", []byte(`null`)},
		{"array", []byte(`[]`)},
		{"wrong-typed-fields", []byte(`{"type":123,"payload":"not-an-object"}`)},
		{"garbage", []byte(`{not json at all`)},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var (
				msg Message
				err error
			)
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Fatalf("Decode panicked on %q payload: %v", tc.name, r)
					}
				}()
				msg, err = Decode(tc.data)
			}()

			if err == nil {
				t.Fatalf("expected a typed decode error for %q, got nil", tc.name)
			}
			var de *DecodeError
			if !errors.As(err, &de) {
				t.Fatalf("expected a typed *DecodeError for %q, got %T: %v", tc.name, err, err)
			}
			if msg != nil {
				t.Fatalf("expected nil message for %q, got %+v", tc.name, msg)
			}
		})
	}
}

func TestDecode_TypePayloadMismatch_ReturnsError(t *testing.T) {
	// Envelope says offer, but the payload carries no SDP shape at all.
	decoded, err := Decode([]byte(`{"type":"offer","payload":{}}`))
	if err == nil {
		t.Fatalf("expected a typed error for type/payload mismatch, got nil")
	}

	var de *DecodeError
	if !errors.As(err, &de) {
		t.Fatalf("expected a typed *DecodeError, got %T: %v", err, err)
	}
	if _, ok := decoded.(OfferMessage); ok {
		t.Fatalf("expected an error, not a silently zero-valued OfferMessage")
	}
}

package signaling

import (
	"testing"

	"github.com/adeelahmad/gadgetcast/internal/core"
)

func TestOnAirStateFrom_Live(t *testing.T) {
	id := mustID(t, "cam-1")

	msg := OnAirStateFrom(id, true)

	if !msg.Live {
		t.Fatalf("expected live=true, got %+v", msg)
	}
	if msg.OnAir != id {
		t.Fatalf("expected on-air id %v, got %v", id, msg.OnAir)
	}
}

func TestOnAirStateFrom_Nobody(t *testing.T) {
	msg := OnAirStateFrom(core.ClientID{}, false)

	if msg.Live {
		t.Fatalf("expected live=false when nobody is on air, got %+v", msg)
	}
	if !msg.OnAir.IsZero() {
		t.Fatalf("expected no phantom on-air id, got %v", msg.OnAir)
	}
}

func TestOnAirState_Projection_RoundTrip(t *testing.T) {
	projected := OnAirStateFrom(mustID(t, "cam-A"), true)

	data, err := Encode(projected)
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
	if got != projected {
		t.Fatalf("on-air projection round-trip mismatch: got %+v want %+v", got, projected)
	}
}

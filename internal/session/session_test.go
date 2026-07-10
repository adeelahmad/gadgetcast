package session_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/adeelahmad/gadgetcast/internal/core"
	"github.com/adeelahmad/gadgetcast/internal/pairing"
	"github.com/adeelahmad/gadgetcast/internal/session"
	"github.com/adeelahmad/gadgetcast/internal/signaling"
	"github.com/adeelahmad/gadgetcast/internal/switcher"
)

const (
	testPIN   = "1234"
	testLimit = 5
)

func id(t *testing.T, s string) core.ClientID {
	t.Helper()
	c, err := core.NewClientID(s)
	if err != nil {
		t.Fatalf("NewClientID(%q): %v", s, err)
	}
	return c
}

func enc(t *testing.T, m signaling.Message) []byte {
	t.Helper()
	b, err := signaling.Encode(m)
	if err != nil {
		t.Fatalf("Encode(%T): %v", m, err)
	}
	return b
}

func newHub(pin string, limit int) (*pairing.Manager, *switcher.Switcher, *session.Hub) {
	pm := pairing.New(pin, limit)
	sw := switcher.New(pm)
	return pm, sw, session.New(pm, sw)
}

func join(t *testing.T, h *session.Hub, c core.ClientID) {
	t.Helper()
	h.Process(session.Inbound{From: c, Data: enc(t, signaling.JoinMessage{Client: c})})
}

func pairUp(t *testing.T, h *session.Hub, c core.ClientID) {
	t.Helper()
	h.Process(session.SubmitPIN{Client: c, PIN: testPIN})
}

func assertNoEffects(t *testing.T, got []session.Effect) {
	t.Helper()
	if len(got) != 0 {
		t.Fatalf("expected no effects, got %d: %+v", len(got), got)
	}
}

func assertErrorEffect(t *testing.T, got []session.Effect, to core.ClientID, wantReasonSubstr string) {
	t.Helper()
	if len(got) != 1 {
		t.Fatalf("expected exactly 1 error effect, got %d: %+v", len(got), got)
	}
	e := got[0]
	if e.To != to {
		t.Errorf("error effect To = %v, want %v", e.To, to)
	}
	em, ok := e.Msg.(signaling.ErrorMessage)
	if !ok {
		t.Fatalf("effect Msg = %#v, want signaling.ErrorMessage", e.Msg)
	}
	if wantReasonSubstr != "" && !strings.Contains(em.Reason, wantReasonSubstr) {
		t.Errorf("error message reason = %q, want to contain %q", em.Reason, wantReasonSubstr)
	}
}

func assertEffects(t *testing.T, got, want []session.Effect) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %d effects, want %d\n got=%+v\nwant=%+v", len(got), len(want), got, want)
	}
	for i := range want {
		if got[i].To != want[i].To {
			t.Errorf("effect[%d].To = %v, want %v", i, got[i].To, want[i].To)
		}
		if !reflect.DeepEqual(got[i].Msg, want[i].Msg) {
			t.Errorf("effect[%d].Msg = %#v, want %#v", i, got[i].Msg, want[i].Msg)
		}
	}
}

// ---- T1: construction + join admission ----

func TestJoin_AdmitsClientAndRegistersPairing(t *testing.T) {
	pm, sw, h := newHub(testPIN, testLimit)
	a := id(t, "A")

	effects := h.Process(session.Inbound{From: a, Data: enc(t, signaling.JoinMessage{Client: a})})

	if !sw.Has(a) {
		t.Errorf("sw.Has(A) = false, want true after join")
	}
	if got := pm.State(a); got != pairing.StatePending {
		t.Errorf("pm.State(A) = %v, want StatePending", got)
	}
	if _, live := sw.OnAir(); live {
		t.Errorf("sw.OnAir() live = true, want false after a bare join")
	}
	assertNoEffects(t, effects)
}

func TestJoin_DuplicateClient_ErrorEffect(t *testing.T) {
	_, sw, h := newHub(testPIN, testLimit)
	a := id(t, "A")
	join(t, h, a)

	effects := h.Process(session.Inbound{From: a, Data: enc(t, signaling.JoinMessage{Client: a})})

	assertErrorEffect(t, effects, a, switcher.ErrDuplicateClient.Error())
	if !sw.Has(a) {
		t.Errorf("sw.Has(A) = false, want true (a duplicate join must leave the connected set unchanged)")
	}
}

func TestJoin_UnpairedClient_VisibleButNotPaired(t *testing.T) {
	pm, sw, h := newHub(testPIN, testLimit)
	a := id(t, "A")

	join(t, h, a)

	if !sw.Has(a) {
		t.Errorf("sw.Has(A) = false, want true (a joined client is visible)")
	}
	if pm.IsPaired(a) {
		t.Errorf("pm.IsPaired(A) = true, want false (joining must not pair a client)")
	}
}

// ---- T2: PIN pairing events ----

func TestSubmitPIN_Correct_PairsClient(t *testing.T) {
	pm, _, h := newHub(testPIN, testLimit)
	a := id(t, "A")
	join(t, h, a)

	effects := h.Process(session.SubmitPIN{Client: a, PIN: testPIN})

	if !pm.IsPaired(a) {
		t.Errorf("pm.IsPaired(A) = false, want true after a correct PIN")
	}
	assertNoEffects(t, effects)
}

func TestSubmitPIN_Incorrect_ErrorEffect(t *testing.T) {
	pm, _, h := newHub(testPIN, 5)
	a := id(t, "A")
	join(t, h, a)

	effects := h.Process(session.SubmitPIN{Client: a, PIN: "wrong"})

	assertErrorEffect(t, effects, a, pairing.ErrIncorrectPIN.Error())
	if got := pm.State(a); got != pairing.StatePending {
		t.Errorf("pm.State(A) = %v, want StatePending after one wrong PIN under the limit", got)
	}
}

func TestSubmitPIN_ExhaustLimit_DurablyRejected(t *testing.T) {
	pm, _, h := newHub(testPIN, 3)
	a := id(t, "A")
	join(t, h, a)

	h.Process(session.SubmitPIN{Client: a, PIN: "no"})
	h.Process(session.SubmitPIN{Client: a, PIN: "no"})
	h.Process(session.SubmitPIN{Client: a, PIN: "no"})

	if got := pm.State(a); got != pairing.StateRejected {
		t.Errorf("pm.State(A) = %v, want StateRejected after 3 wrong PINs (limit 3)", got)
	}

	effects := h.Process(session.SubmitPIN{Client: a, PIN: testPIN})
	assertErrorEffect(t, effects, a, pairing.ErrRejected.Error())
}

func TestSubmitPIN_UnknownClient_ErrorEffect(t *testing.T) {
	_, _, h := newHub(testPIN, testLimit)
	z := id(t, "Z")

	effects := h.Process(session.SubmitPIN{Client: z, PIN: testPIN})

	assertErrorEffect(t, effects, z, pairing.ErrUnknownClient.Error())
}

// ---- T3: switch-source on-air routing + broadcast ----

func TestSwitchSource_PairedClient_GoesOnAirAndBroadcasts(t *testing.T) {
	_, sw, h := newHub(testPIN, testLimit)
	a := id(t, "A")
	join(t, h, a)
	pairUp(t, h, a)

	effects := h.Process(session.Inbound{From: a, Data: enc(t, signaling.SwitchSourceMessage{Target: a})})

	onAir, live := sw.OnAir()
	if !live || onAir != a {
		t.Errorf("sw.OnAir() = (%v, %v), want (A, true)", onAir, live)
	}
	want := signaling.OnAirStateFrom(a, true)
	found := false
	for _, e := range effects {
		if e.To == a && reflect.DeepEqual(e.Msg, want) {
			found = true
		}
	}
	if !found {
		t.Errorf("effects %+v do not include OnAirStateFrom(A,true) addressed to A", effects)
	}
}

func TestSwitchSource_UnpairedTarget_ErrorEffect_NotOnAir(t *testing.T) {
	_, sw, h := newHub(testPIN, testLimit)
	a := id(t, "A")
	join(t, h, a) // joined but NOT paired

	effects := h.Process(session.Inbound{From: a, Data: enc(t, signaling.SwitchSourceMessage{Target: a})})

	assertErrorEffect(t, effects, a, switcher.ErrNotPaired.Error())
	if _, live := sw.OnAir(); live {
		t.Errorf("sw.OnAir() live = true, want false (an unpaired client must never reach on-air)")
	}
}

func TestSwitchSource_UnknownTarget_ErrorEffect(t *testing.T) {
	_, sw, h := newHub(testPIN, testLimit)
	a := id(t, "A")
	g := id(t, "G")
	join(t, h, a)
	pairUp(t, h, a)

	effects := h.Process(session.Inbound{From: a, Data: enc(t, signaling.SwitchSourceMessage{Target: g})})

	assertErrorEffect(t, effects, a, switcher.ErrUnknownClient.Error())
	if _, live := sw.OnAir(); live {
		t.Errorf("sw.OnAir() live = true, want false (switching to an unknown target must not change on-air)")
	}
}

func TestSwitchSource_BroadcastAddressedToAllConnected(t *testing.T) {
	_, _, h := newHub(testPIN, testLimit)
	a := id(t, "A")
	b := id(t, "B")
	c := id(t, "C")
	for _, x := range []core.ClientID{a, b, c} {
		join(t, h, x)
		pairUp(t, h, x)
	}

	effects := h.Process(session.Inbound{From: a, Data: enc(t, signaling.SwitchSourceMessage{Target: a})})

	on := signaling.OnAirStateFrom(a, true)
	want := []session.Effect{
		{To: a, Msg: on},
		{To: b, Msg: on},
		{To: c, Msg: on},
	}
	assertEffects(t, effects, want)
}

// ---- T4: disconnect handling ----

func TestDisconnect_OnAirClient_ClearsAndBroadcastsNobody(t *testing.T) {
	_, sw, h := newHub(testPIN, testLimit)
	a := id(t, "A")
	b := id(t, "B")
	join(t, h, a)
	pairUp(t, h, a)
	join(t, h, b)
	pairUp(t, h, b)
	h.Process(session.Inbound{From: a, Data: enc(t, signaling.SwitchSourceMessage{Target: a})}) // A on-air

	effects := h.Process(session.Disconnect{Client: a})

	onAir, live := sw.OnAir()
	if live || !onAir.IsZero() {
		t.Errorf("sw.OnAir() = (%v, %v), want (zero, false) after the on-air client disconnects", onAir, live)
	}
	if sw.Has(a) {
		t.Errorf("sw.Has(A) = true, want false after disconnect")
	}
	want := []session.Effect{
		{To: b, Msg: signaling.OnAirStateFrom(core.ClientID{}, false)},
	}
	assertEffects(t, effects, want)
}

func TestDisconnect_NonOnAirClient_NoBroadcast(t *testing.T) {
	_, sw, h := newHub(testPIN, testLimit)
	a := id(t, "A")
	b := id(t, "B")
	join(t, h, a)
	pairUp(t, h, a)
	join(t, h, b)
	pairUp(t, h, b)
	h.Process(session.Inbound{From: a, Data: enc(t, signaling.SwitchSourceMessage{Target: a})}) // A on-air

	effects := h.Process(session.Disconnect{Client: b})

	onAir, live := sw.OnAir()
	if !live || onAir != a {
		t.Errorf("sw.OnAir() = (%v, %v), want (A, true) unchanged after a non-on-air disconnect", onAir, live)
	}
	if sw.Has(b) {
		t.Errorf("sw.Has(B) = true, want false after disconnect")
	}
	for _, e := range effects {
		if e.Msg.Type() == signaling.TypeOnAirState {
			t.Errorf("unexpected on-air-state effect on a non-on-air disconnect: %+v", e)
		}
	}
}

func TestDisconnect_UnknownClient_SafeNoop(t *testing.T) {
	_, sw, h := newHub(testPIN, testLimit)
	z := id(t, "Z")

	effects := h.Process(session.Disconnect{Client: z})

	assertNoEffects(t, effects)
	if sw.Has(z) {
		t.Errorf("sw.Has(Z) = true, want false")
	}
	if _, live := sw.OnAir(); live {
		t.Errorf("sw.OnAir() live = true, want false after a no-op disconnect")
	}
}

func TestDisconnect_PairingStateRetained(t *testing.T) {
	pm, _, h := newHub(testPIN, testLimit)
	a := id(t, "A")
	join(t, h, a)
	pairUp(t, h, a)
	h.Process(session.Inbound{From: a, Data: enc(t, signaling.SwitchSourceMessage{Target: a})})

	h.Process(session.Disconnect{Client: a})

	if got := pm.State(a); got != pairing.StatePaired {
		t.Errorf("pm.State(A) = %v, want StatePaired (pairing is durable across disconnect)", got)
	}
}

// ---- T5: decode robustness + opaque relay ----

func TestInbound_MalformedBytes_ErrorEffectNoPanic(t *testing.T) {
	_, _, h := newHub(testPIN, testLimit)
	a := id(t, "A")
	join(t, h, a)

	effects := h.Process(session.Inbound{From: a, Data: []byte("{not json")})

	assertErrorEffect(t, effects, a, "")
}

func TestInbound_UnknownType_ErrorEffectNoPanic(t *testing.T) {
	_, _, h := newHub(testPIN, testLimit)
	a := id(t, "A")
	join(t, h, a)

	raw := []byte(`{"type":"nonsense","payload":{}}`)
	effects := h.Process(session.Inbound{From: a, Data: raw})

	assertErrorEffect(t, effects, a, "")
}

func TestInbound_UnexpectedServerMessage_ErrorEffect(t *testing.T) {
	_, sw, h := newHub(testPIN, testLimit)
	a := id(t, "A")
	join(t, h, a)

	data := enc(t, signaling.OnAirStateFrom(a, true)) // a server->client type sent by a client
	effects := h.Process(session.Inbound{From: a, Data: data})

	assertErrorEffect(t, effects, a, "")
	if _, live := sw.OnAir(); live {
		t.Errorf("sw.OnAir() live = true, want false (an unexpected server message must not change state)")
	}
}

func TestInbound_OfferRelayedOpaquelyToPeers(t *testing.T) {
	_, _, h := newHub(testPIN, testLimit)
	a := id(t, "A")
	b := id(t, "B")
	c := id(t, "C")
	for _, x := range []core.ClientID{a, b, c} {
		join(t, h, x)
	}

	const sdp = "v=0 opaque-offer-sdp"
	effects := h.Process(session.Inbound{From: a, Data: enc(t, signaling.OfferMessage{SDP: sdp})})

	offer := signaling.OfferMessage{SDP: sdp}
	want := []session.Effect{
		{To: b, Msg: offer},
		{To: c, Msg: offer},
	}
	assertEffects(t, effects, want)
}

func TestInbound_IceCandidateRelayedOpaquely(t *testing.T) {
	_, _, h := newHub(testPIN, testLimit)
	a := id(t, "A")
	b := id(t, "B")
	join(t, h, a)
	join(t, h, b)

	const cand = "candidate:opaque 1 udp"
	effects := h.Process(session.Inbound{From: a, Data: enc(t, signaling.IceCandidateMessage{Candidate: cand})})

	want := []session.Effect{
		{To: b, Msg: signaling.IceCandidateMessage{Candidate: cand}},
	}
	assertEffects(t, effects, want)
}

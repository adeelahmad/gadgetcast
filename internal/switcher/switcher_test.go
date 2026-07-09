package switcher_test

import (
	"testing"

	"github.com/adeelahmad/gadgetcast/internal/core"
	"github.com/adeelahmad/gadgetcast/internal/switcher"
)

// mustID builds a valid core.ClientID or fails the test.
func mustID(t *testing.T, raw string) core.ClientID {
	t.Helper()
	id, err := core.NewClientID(raw)
	if err != nil {
		t.Fatalf("build client id %q: %v", raw, err)
	}
	return id
}

// stubOracle is an injectable PairingOracle. When allPaired is true every
// client is considered paired; otherwise only ids present in paired are.
type stubOracle struct {
	allPaired bool
	paired    map[core.ClientID]bool
}

func (s stubOracle) IsPaired(id core.ClientID) bool {
	if s.allPaired {
		return true
	}
	return s.paired[id]
}

// everyonePaired returns an oracle that pairs every client, for tests that
// exercise set/on-air behavior rather than the pairing gate itself.
func everyonePaired() stubOracle { return stubOracle{allPaired: true} }

// ---- T1: client set ----------------------------------------------------

func TestAddClient_AppearsInSet(t *testing.T) {
	sw := switcher.New(everyonePaired())
	cam := mustID(t, "cam-1")

	if err := sw.AddClient(cam); err != nil {
		t.Fatalf("AddClient(cam-1): unexpected error: %v", err)
	}
	if !sw.Has(cam) {
		t.Fatalf("Has(cam-1) = false, want true after AddClient")
	}
}

func TestAddClient_Duplicate_Errors(t *testing.T) {
	sw := switcher.New(everyonePaired())
	cam := mustID(t, "cam-1")

	if err := sw.AddClient(cam); err != nil {
		t.Fatalf("first AddClient(cam-1): unexpected error: %v", err)
	}
	if err := sw.AddClient(cam); err == nil {
		t.Fatalf("duplicate AddClient(cam-1) = nil, want typed error")
	}
	if !sw.Has(cam) {
		t.Fatalf("Has(cam-1) = false after duplicate add, set must be unchanged")
	}
}

func TestRemoveClient_Unknown_Errors(t *testing.T) {
	sw := switcher.New(everyonePaired())

	if err := sw.RemoveClient(mustID(t, "ghost")); err == nil {
		t.Fatalf("RemoveClient(ghost) = nil, want typed error for unknown client")
	}
}

// ---- T2: set-on-air + never-two-on-air invariant -----------------------

func TestNewSwitcher_NoneOnAirByDefault(t *testing.T) {
	sw := switcher.New(everyonePaired())
	for _, raw := range []string{"a", "b", "c"} {
		if err := sw.AddClient(mustID(t, raw)); err != nil {
			t.Fatalf("AddClient(%q): %v", raw, err)
		}
	}

	got, live := sw.OnAir()
	if live {
		t.Fatalf("OnAir() live = true, want false with nobody set on-air; got %v", got)
	}
	if !got.IsZero() {
		t.Fatalf("OnAir() id = %v, want zero client when nobody is on-air", got)
	}
}

func TestSetOnAir_MakesOnlyThatClientOnAir(t *testing.T) {
	sw := switcher.New(everyonePaired())
	a, b := mustID(t, "a"), mustID(t, "b")
	for _, id := range []core.ClientID{a, b} {
		if err := sw.AddClient(id); err != nil {
			t.Fatalf("AddClient(%v): %v", id, err)
		}
	}

	if err := sw.SetOnAir(a); err != nil {
		t.Fatalf("SetOnAir(a): unexpected error: %v", err)
	}

	got, live := sw.OnAir()
	if !live {
		t.Fatalf("OnAir() live = false, want true after SetOnAir(a)")
	}
	if got != a {
		t.Fatalf("OnAir() = %v, want a (only a on-air)", got)
	}
}

func TestSetOnAir_Switch_A_to_B_DisplacesA(t *testing.T) {
	sw := switcher.New(everyonePaired())
	a, b := mustID(t, "a"), mustID(t, "b")
	for _, id := range []core.ClientID{a, b} {
		if err := sw.AddClient(id); err != nil {
			t.Fatalf("AddClient(%v): %v", id, err)
		}
	}
	if err := sw.SetOnAir(a); err != nil {
		t.Fatalf("SetOnAir(a): %v", err)
	}

	if err := sw.SetOnAir(b); err != nil {
		t.Fatalf("SetOnAir(b): unexpected error: %v", err)
	}

	got, live := sw.OnAir()
	if !live {
		t.Fatalf("OnAir() live = false, want true after SetOnAir(b)")
	}
	if got != b {
		t.Fatalf("OnAir() = %v, want b (a must be displaced)", got)
	}
	if got == a {
		t.Fatalf("OnAir() still reports a; b did not displace a")
	}
}

func TestOnAir_ExactlyOne_AcrossSequence(t *testing.T) {
	sw := switcher.New(everyonePaired())
	a, b, c := mustID(t, "a"), mustID(t, "b"), mustID(t, "c")

	// assertOnAir enforces the never-two-on-air invariant: OnAir is the single
	// source of truth, so after every step at most one client is live and it is
	// exactly the expected one (README §3.3).
	assertOnAir := func(step string, want core.ClientID, wantLive bool) {
		t.Helper()
		got, live := sw.OnAir()
		if live != wantLive {
			t.Fatalf("[%s] OnAir() live = %v, want %v", step, live, wantLive)
		}
		if wantLive && got != want {
			t.Fatalf("[%s] OnAir() = %v, want %v (never two on-air)", step, got, want)
		}
		if !wantLive && !got.IsZero() {
			t.Fatalf("[%s] OnAir() = %v, want nobody on-air", step, got)
		}
	}

	for _, id := range []core.ClientID{a, b, c} {
		if err := sw.AddClient(id); err != nil {
			t.Fatalf("AddClient(%v): %v", id, err)
		}
		assertOnAir("after add "+id.String(), core.ClientID{}, false)
	}

	if err := sw.SetOnAir(a); err != nil {
		t.Fatalf("SetOnAir(a): %v", err)
	}
	assertOnAir("set a", a, true)

	if err := sw.SetOnAir(b); err != nil {
		t.Fatalf("SetOnAir(b): %v", err)
	}
	assertOnAir("switch a->b", b, true)

	if err := sw.SetOnAir(c); err != nil {
		t.Fatalf("SetOnAir(c): %v", err)
	}
	assertOnAir("switch b->c", c, true)

	if err := sw.RemoveClient(c); err != nil {
		t.Fatalf("RemoveClient(c) while on-air: %v", err)
	}
	assertOnAir("remove on-air c", core.ClientID{}, false)

	if err := sw.SetOnAir(a); err != nil {
		t.Fatalf("SetOnAir(a) again: %v", err)
	}
	assertOnAir("set a again", a, true)

	if err := sw.RemoveClient(b); err != nil {
		t.Fatalf("RemoveClient(b) not on-air: %v", err)
	}
	assertOnAir("remove non-on-air b", a, true)
}

// ---- T3: unknown-client + remove-on-air --------------------------------

func TestSetOnAir_UnknownClient_Errors_StateUnchanged(t *testing.T) {
	sw := switcher.New(everyonePaired())
	a := mustID(t, "a")
	if err := sw.AddClient(a); err != nil {
		t.Fatalf("AddClient(a): %v", err)
	}
	if err := sw.SetOnAir(a); err != nil {
		t.Fatalf("SetOnAir(a): %v", err)
	}

	if err := sw.SetOnAir(mustID(t, "ghost")); err == nil {
		t.Fatalf("SetOnAir(ghost) = nil, want typed error for unknown client")
	}

	got, live := sw.OnAir()
	if !live || got != a {
		t.Fatalf("OnAir() = (%v, %v), want (a, true); no phantom on-air client", got, live)
	}
}

func TestRemoveClient_OnAir_TransitionsToNobody(t *testing.T) {
	sw := switcher.New(everyonePaired())
	a := mustID(t, "a")
	if err := sw.AddClient(a); err != nil {
		t.Fatalf("AddClient(a): %v", err)
	}
	if err := sw.SetOnAir(a); err != nil {
		t.Fatalf("SetOnAir(a): %v", err)
	}

	if err := sw.RemoveClient(a); err != nil {
		t.Fatalf("RemoveClient(a) while on-air: unexpected error: %v", err)
	}

	got, live := sw.OnAir()
	if live {
		t.Fatalf("OnAir() live = true after removing on-air a, want false (nobody on-air)")
	}
	if !got.IsZero() {
		t.Fatalf("OnAir() = %v, want zero client; no dangling reference to a", got)
	}
	if sw.Has(a) {
		t.Fatalf("Has(a) = true after RemoveClient(a)")
	}
}

func TestRemoveClient_NotOnAir_LeavesOnAirUntouched(t *testing.T) {
	sw := switcher.New(everyonePaired())
	a, b := mustID(t, "a"), mustID(t, "b")
	for _, id := range []core.ClientID{a, b} {
		if err := sw.AddClient(id); err != nil {
			t.Fatalf("AddClient(%v): %v", id, err)
		}
	}
	if err := sw.SetOnAir(a); err != nil {
		t.Fatalf("SetOnAir(a): %v", err)
	}

	if err := sw.RemoveClient(b); err != nil {
		t.Fatalf("RemoveClient(b) not on-air: unexpected error: %v", err)
	}

	got, live := sw.OnAir()
	if !live || got != a {
		t.Fatalf("OnAir() = (%v, %v), want (a, true); removing b must not touch on-air", got, live)
	}
}

// ---- T4: pairing gate + on-air source of truth -------------------------

func TestSetOnAir_UnpairedClient_Errors(t *testing.T) {
	u := mustID(t, "unpaired")
	oracle := stubOracle{paired: map[core.ClientID]bool{}} // u is NOT paired
	sw := switcher.New(oracle)
	if err := sw.AddClient(u); err != nil {
		t.Fatalf("AddClient(u): %v", err)
	}

	if err := sw.SetOnAir(u); err == nil {
		t.Fatalf("SetOnAir(unpaired) = nil, want typed error (decision #5)")
	}

	got, live := sw.OnAir()
	if live || !got.IsZero() {
		t.Fatalf("OnAir() = (%v, %v), want nobody; unpaired client must not go on-air", got, live)
	}
}

func TestSetOnAir_PairedClient_Succeeds(t *testing.T) {
	p := mustID(t, "paired")
	oracle := stubOracle{paired: map[core.ClientID]bool{p: true}}
	sw := switcher.New(oracle)
	if err := sw.AddClient(p); err != nil {
		t.Fatalf("AddClient(p): %v", err)
	}

	if err := sw.SetOnAir(p); err != nil {
		t.Fatalf("SetOnAir(paired): unexpected error: %v", err)
	}

	got, live := sw.OnAir()
	if !live || got != p {
		t.Fatalf("OnAir() = (%v, %v), want (p, true)", got, live)
	}
}

func TestUnpairedClient_MayConnect_ButNotOnAir(t *testing.T) {
	u := mustID(t, "unpaired")
	oracle := stubOracle{paired: map[core.ClientID]bool{}} // u is NOT paired
	sw := switcher.New(oracle)

	// Unpaired clients may connect (join the set / be visible)...
	if err := sw.AddClient(u); err != nil {
		t.Fatalf("AddClient(unpaired): want success, got error: %v", err)
	}
	if !sw.Has(u) {
		t.Fatalf("Has(unpaired) = false, want true; unpaired may connect and be visible")
	}

	// ...but must not be selectable on-air (decision #5 coupling).
	if err := sw.SetOnAir(u); err == nil {
		t.Fatalf("SetOnAir(unpaired) = nil, want typed error")
	}
	if _, live := sw.OnAir(); live {
		t.Fatalf("OnAir() live = true, want false; unpaired client must not be on-air")
	}
}

func TestOnAir_ReturnsSourceOfTruthShape(t *testing.T) {
	p := mustID(t, "paired")
	oracle := stubOracle{paired: map[core.ClientID]bool{p: true}}
	sw := switcher.New(oracle)
	if err := sw.AddClient(p); err != nil {
		t.Fatalf("AddClient(p): %v", err)
	}
	if err := sw.SetOnAir(p); err != nil {
		t.Fatalf("SetOnAir(p): %v", err)
	}

	// OnAir() is the exact (core.ClientID, bool) shape signaling.OnAirStateFrom
	// consumes — one source of truth (decision #8).
	var (
		id   core.ClientID
		live bool
	)
	id, live = sw.OnAir()
	if id != p || live != true {
		t.Fatalf("OnAir() = (%v, %v), want (p, true) as the source-of-truth shape", id, live)
	}
}

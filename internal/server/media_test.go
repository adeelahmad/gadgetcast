package server

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/adeelahmad/gadgetcast/internal/core"
)

type fakeMedia struct {
	mu      sync.Mutex
	offers  []string
	onAir   string
	live    bool
	dropped []string
	fail    bool
}

func (f *fakeMedia) Offer(id core.ClientID, sdp string) (string, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.fail {
		return "", errors.New("media: bad offer")
	}
	f.offers = append(f.offers, id.String()+":"+sdp)
	return "answer-for-" + id.String(), nil
}

func (f *fakeMedia) SetOnAir(id core.ClientID, live bool) {
	f.mu.Lock()
	f.onAir, f.live = id.String(), live
	f.mu.Unlock()
}

func (f *fakeMedia) Drop(id core.ClientID) {
	f.mu.Lock()
	f.dropped = append(f.dropped, id.String())
	f.mu.Unlock()
}

func (f *fakeMedia) state() (string, bool, []string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.onAir, f.live, append([]string(nil), f.dropped...)
}

func TestMediaOfferRequiresPairingAndFollowsOnAir(t *testing.T) {
	fm := &fakeMedia{}
	e := newEnv(t, nil, func(o *Options) { o.Media = fm })
	a := e.dial(t, "/ws")

	a.send(`{"type":"media-offer","payload":{"sdp":"v=0"}}`)
	a.expect("error", "join before")

	e.join(t, a, "alice")
	a.send(`{"type":"media-offer","payload":{"sdp":"v=0"}}`)
	a.expect("error", "pair with the PIN")

	a.send(`{"type":"submit-pin","payload":{"pin":"1234"}}`)
	a.send(`{"type":"media-offer","payload":{}}`)
	a.expect("error", "media-offer: payload")
	a.send(`{"type":"media-offer","payload":{"sdp":"v=0 alice"}}`)
	a.expect("media-answer", `"sdp":"answer-for-alice"`)

	a.send(`{"type":"switch-source","payload":{"target":"alice"}}`)
	a.expect("on-air-state", `"live":true`)
	if on, live, _ := fm.state(); on != "alice" || !live {
		t.Fatalf("bridge on air = %q %v", on, live)
	}

	a.c.CloseNow()
	deadline := time.Now().Add(5 * time.Second)
	for {
		_, live, dropped := fm.state()
		if !live && len(dropped) == 1 && dropped[0] == "alice" {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("after disconnect: live=%v dropped=%v", live, dropped)
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestMediaOfferErrors(t *testing.T) {
	e := newEnv(t, nil)
	a := e.dial(t, "/ws")
	e.join(t, a, "alice")
	a.send(`{"type":"submit-pin","payload":{"pin":"1234"}}`)
	a.send(`{"type":"media-offer","payload":{"sdp":"v=0"}}`)
	a.expect("error", "not available")

	fm := &fakeMedia{fail: true}
	e2 := newEnv(t, nil, func(o *Options) { o.Media = fm })
	b := e2.dial(t, "/ws")
	e2.join(t, b, "bob")
	b.send(`{"type":"submit-pin","payload":{"pin":"1234"}}`)
	b.send(`{"type":"media-offer","payload":{"sdp":"v=0"}}`)
	b.expect("error", "bad offer")
}

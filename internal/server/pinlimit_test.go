package server

import (
	"fmt"
	"testing"
	"time"
)

func TestPinLimiterLocksOutAndBacksOff(t *testing.T) {
	now := time.Unix(0, 0)
	l := newPinLimiter(func() time.Time { return now })
	for i := 0; i < pinMaxFails-1; i++ {
		l.fail("a")
		if w := l.wait("a"); w != 0 {
			t.Fatalf("locked after %d fails", i+1)
		}
	}
	l.fail("a")
	if w := l.wait("a"); w != time.Minute {
		t.Fatalf("first lockout = %v", w)
	}
	if l.wait("b") != 0 {
		t.Fatal("other address locked")
	}
	now = now.Add(time.Minute)
	if l.wait("a") != 0 {
		t.Fatal("still locked after lockout")
	}
	for i := 0; i < pinMaxFails; i++ {
		l.fail("a")
	}
	if w := l.wait("a"); w != 2*time.Minute {
		t.Fatalf("second lockout = %v, want doubling", w)
	}
	for i := 0; i < 20*pinMaxFails; i++ {
		l.fail("a")
	}
	if w := l.wait("a"); w != pinMaxLockout {
		t.Fatalf("capped lockout = %v", w)
	}
	l.success("a")
	if l.wait("a") != 0 {
		t.Fatal("success did not reset")
	}
}

func TestPinLimiterPrunes(t *testing.T) {
	now := time.Unix(0, 0)
	l := newPinLimiter(func() time.Time { return now })
	for i := 0; i < pinMaxTracked; i++ {
		l.fail(fmt.Sprint(i))
	}
	l.fail("new")
	if len(l.m) != 1 {
		t.Fatalf("tracked = %d after prune, want 1", len(l.m))
	}
	if newPinLimiter(nil).now == nil {
		t.Fatal("default clock missing")
	}
}

func TestSignaling_PINGuessesRateLimitedAcrossClientIDs(t *testing.T) {
	e := newEnv(t, nil)
	// A fresh client id per guess must not reset the attempt budget.
	for i := 0; i < pinMaxFails; i++ {
		c := e.dial(t, "/ws")
		e.join(t, c, fmt.Sprintf("guess%d", i))
		c.send(`{"type":"submit-pin","payload":{"pin":"0000"}}`)
		c.expect("error", "incorrect pin")
	}
	c := e.dial(t, "/ws")
	e.join(t, c, "guess-final")
	c.send(`{"type":"submit-pin","payload":{"pin":"1234"}}`) // even the right PIN waits
	c.expect("error", "too many wrong PINs")
	c.send(`{"type":"switch-source","payload":{"target":"guess-final"}}`)
	c.expect("error", "not paired")
}

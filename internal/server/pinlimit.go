package server

import (
	"sync"
	"time"
)

// pinLimiter rate-limits PIN guesses per remote address. The pairing state
// machine locks out a client id, but ids are chosen by the client, so without
// this a fresh id per guess would brute-force the PIN (README §4 security).
// Every pinMaxFails wrong PINs from one address lock it out for pinBaseLockout,
// doubling each time up to pinMaxLockout.
type pinLimiter struct {
	mu  sync.Mutex
	now func() time.Time
	m   map[string]*pinRecord
}

type pinRecord struct {
	fails    int
	lockouts int
	until    time.Time
}

const (
	pinMaxFails    = 5
	pinBaseLockout = time.Minute
	pinMaxLockout  = time.Hour
	pinMaxTracked  = 4096
)

func newPinLimiter(now func() time.Time) *pinLimiter {
	if now == nil {
		now = time.Now
	}
	return &pinLimiter{now: now, m: map[string]*pinRecord{}}
}

// wait returns how long addr must wait before its next guess (0 = allowed).
func (l *pinLimiter) wait(addr string) time.Duration {
	l.mu.Lock()
	defer l.mu.Unlock()
	r := l.m[addr]
	if r == nil {
		return 0
	}
	if d := r.until.Sub(l.now()); d > 0 {
		return d
	}
	return 0
}

// fail records a wrong PIN from addr.
func (l *pinLimiter) fail(addr string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	r := l.m[addr]
	if r == nil {
		if len(l.m) >= pinMaxTracked {
			l.pruneLocked()
		}
		r = &pinRecord{}
		l.m[addr] = r
	}
	r.fails++
	if r.fails < pinMaxFails {
		return
	}
	r.fails = 0
	d := pinBaseLockout << r.lockouts
	if d > pinMaxLockout || d <= 0 {
		d = pinMaxLockout
	}
	r.lockouts++
	r.until = l.now().Add(d)
}

// success forgets addr's history once it pairs.
func (l *pinLimiter) success(addr string) {
	l.mu.Lock()
	delete(l.m, addr)
	l.mu.Unlock()
}

// pruneLocked drops addresses that are not locked out. Caller holds mu.
func (l *pinLimiter) pruneLocked() {
	now := l.now()
	for a, r := range l.m {
		if !r.until.After(now) {
			delete(l.m, a)
		}
	}
}

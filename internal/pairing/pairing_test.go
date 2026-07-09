package pairing_test

import (
	"errors"
	"testing"

	"github.com/adeelahmad/gadgetcast/internal/core"
	"github.com/adeelahmad/gadgetcast/internal/pairing"
)

// PINs are injected by the tests — never hardcoded in production source.
const (
	correctPIN = "1234"
	wrongPIN   = "0000"
)

func mustID(t *testing.T, raw string) core.ClientID {
	t.Helper()
	id, err := core.NewClientID(raw)
	if err != nil {
		t.Fatalf("core.NewClientID(%q) failed: %v", raw, err)
	}
	return id
}

// submitWrong submits an incorrect PIN n times and fails the test on a panic.
func submitWrong(m *pairing.Manager, id core.ClientID, n int) {
	for i := 0; i < n; i++ {
		_ = m.SubmitPIN(id, wrongPIN)
	}
}

// --- T1: pending -> paired ---------------------------------------------------

func TestNewClient_StartsPending(t *testing.T) {
	m := pairing.New(correctPIN, 5)
	id := mustID(t, "cam-fresh")

	m.Register(id)

	if got := m.State(id); got != pairing.StatePending {
		t.Errorf("fresh client State() = %v, want %v", got, pairing.StatePending)
	}
	if m.IsPaired(id) {
		t.Errorf("fresh client IsPaired() = true, want false")
	}
}

func TestSubmitPIN_Correct_TransitionsToPaired(t *testing.T) {
	m := pairing.New(correctPIN, 5)
	id := mustID(t, "cam-correct")
	m.Register(id)

	if err := m.SubmitPIN(id, correctPIN); err != nil {
		t.Fatalf("SubmitPIN(correct) returned error: %v", err)
	}
	if got := m.State(id); got != pairing.StatePaired {
		t.Errorf("after correct PIN State() = %v, want %v", got, pairing.StatePaired)
	}
}

func TestSubmitPIN_Correct_IsIdempotent(t *testing.T) {
	m := pairing.New(correctPIN, 5)
	id := mustID(t, "cam-idem")
	m.Register(id)

	if err := m.SubmitPIN(id, correctPIN); err != nil {
		t.Fatalf("first SubmitPIN(correct) returned error: %v", err)
	}
	if got := m.State(id); got != pairing.StatePaired {
		t.Fatalf("after first correct PIN State() = %v, want %v", got, pairing.StatePaired)
	}

	// A second correct submission must not error and must not re-transition.
	if err := m.SubmitPIN(id, correctPIN); err != nil {
		t.Errorf("idempotent SubmitPIN(correct) returned error: %v", err)
	}
	if got := m.State(id); got != pairing.StatePaired {
		t.Errorf("after repeat correct PIN State() = %v, want still %v", got, pairing.StatePaired)
	}
	if got := m.Attempts(id); got != 0 {
		t.Errorf("Attempts() = %d after correct PINs, want 0 (correct never counts as failure)", got)
	}
}

func TestSubmitPIN_UnknownClient_Errors(t *testing.T) {
	m := pairing.New(correctPIN, 5)
	unknown := mustID(t, "cam-ghost") // never registered

	err := m.SubmitPIN(unknown, correctPIN)
	if !errors.Is(err, pairing.ErrUnknownClient) {
		t.Errorf("SubmitPIN(unknown) error = %v, want ErrUnknownClient", err)
	}
	// No state must be created for an unknown client.
	if got := m.State(unknown); got != pairing.StateUnknown {
		t.Errorf("unknown client State() = %v, want %v (no state created)", got, pairing.StateUnknown)
	}
	if m.IsPaired(unknown) {
		t.Errorf("unknown client IsPaired() = true, want false")
	}
}

// --- T2: attempt limit + durable lockout ------------------------------------

func TestSubmitPIN_Wrong_StaysPending_IncrementsAttempts(t *testing.T) {
	m := pairing.New(correctPIN, 5)
	id := mustID(t, "cam-wrong1")
	m.Register(id)

	err := m.SubmitPIN(id, wrongPIN)
	if !errors.Is(err, pairing.ErrIncorrectPIN) {
		t.Errorf("SubmitPIN(wrong) error = %v, want ErrIncorrectPIN", err)
	}
	if got := m.State(id); got != pairing.StatePending {
		t.Errorf("after one wrong PIN State() = %v, want %v", got, pairing.StatePending)
	}
	if got := m.Attempts(id); got != 1 {
		t.Errorf("Attempts() = %d after one wrong PIN, want 1", got)
	}
}

func TestSubmitPIN_ExhaustLimit_TransitionsToRejected(t *testing.T) {
	const limit = 5
	m := pairing.New(correctPIN, limit)
	id := mustID(t, "cam-exhaust")
	m.Register(id)

	submitWrong(m, id, limit)

	if got := m.State(id); got != pairing.StateRejected {
		t.Errorf("after %d wrong PINs State() = %v, want %v", limit, got, pairing.StateRejected)
	}
}

func TestSubmitPIN_AfterRejected_FurtherAttemptsRejected(t *testing.T) {
	const limit = 5
	m := pairing.New(correctPIN, limit)
	id := mustID(t, "cam-afterreject")
	m.Register(id)
	submitWrong(m, id, limit) // drive to rejected

	// Any further attempt is refused and must not revert to pending.
	err := m.SubmitPIN(id, wrongPIN)
	if !errors.Is(err, pairing.ErrRejected) {
		t.Errorf("SubmitPIN after rejection error = %v, want ErrRejected", err)
	}
	if got := m.State(id); got != pairing.StateRejected {
		t.Errorf("State() = %v after further attempt, want still %v (never reverts to pending)", got, pairing.StateRejected)
	}
}

func TestSubmitPIN_CorrectAfterLockout_StillRejected(t *testing.T) {
	const limit = 5
	m := pairing.New(correctPIN, limit)
	id := mustID(t, "cam-latecorrect")
	m.Register(id)
	submitWrong(m, id, limit) // drive to rejected

	// A late-correct PIN must NOT re-pair: lockout is durable.
	_ = m.SubmitPIN(id, correctPIN)

	if got := m.State(id); got != pairing.StateRejected {
		t.Errorf("State() = %v after correct PIN post-lockout, want %v (durable lockout)", got, pairing.StateRejected)
	}
	if m.IsPaired(id) {
		t.Errorf("IsPaired() = true after correct PIN post-lockout, want false")
	}
}

// --- T3: IsPaired query + configurable limit --------------------------------

func TestIsPaired_TrueOnlyWhenPaired(t *testing.T) {
	m := pairing.New(correctPIN, 5)
	id := mustID(t, "cam-paired")
	m.Register(id)
	if err := m.SubmitPIN(id, correctPIN); err != nil {
		t.Fatalf("SubmitPIN(correct) returned error: %v", err)
	}

	if got := m.State(id); got != pairing.StatePaired {
		t.Fatalf("setup: State() = %v, want %v", got, pairing.StatePaired)
	}
	if !m.IsPaired(id) {
		t.Errorf("IsPaired() = false for paired client, want true")
	}
}

func TestIsPaired_FalseWhenPending(t *testing.T) {
	m := pairing.New(correctPIN, 5)
	id := mustID(t, "cam-pending")
	m.Register(id)

	if got := m.State(id); got != pairing.StatePending {
		t.Fatalf("setup: State() = %v, want %v", got, pairing.StatePending)
	}
	if m.IsPaired(id) {
		t.Errorf("IsPaired() = true for pending client, want false")
	}
}

func TestIsPaired_FalseWhenRejected(t *testing.T) {
	const limit = 5
	m := pairing.New(correctPIN, limit)
	id := mustID(t, "cam-rejected")
	m.Register(id)
	submitWrong(m, id, limit)

	if got := m.State(id); got != pairing.StateRejected {
		t.Fatalf("setup: State() = %v, want %v", got, pairing.StateRejected)
	}
	if m.IsPaired(id) {
		t.Errorf("IsPaired() = true for rejected client, want false")
	}
}

func TestAttemptLimit_Configurable(t *testing.T) {
	const limit = 3
	m := pairing.New(correctPIN, limit)
	id := mustID(t, "cam-limit3")
	m.Register(id)

	// Two wrong attempts stay pending (limit not hardwired to 5).
	submitWrong(m, id, limit-1)
	if got := m.State(id); got != pairing.StatePending {
		t.Errorf("after %d wrong PINs State() = %v, want %v", limit-1, got, pairing.StatePending)
	}

	// The limit-th wrong attempt rejects.
	_ = m.SubmitPIN(id, wrongPIN)
	if got := m.State(id); got != pairing.StateRejected {
		t.Errorf("after %d wrong PINs State() = %v, want %v (limit honored)", limit, got, pairing.StateRejected)
	}
}

func TestRejectedState_QueryableAfterRejection(t *testing.T) {
	const limit = 5
	m := pairing.New(correctPIN, limit)
	id := mustID(t, "cam-queryable")
	m.Register(id)
	submitWrong(m, id, limit)

	// Queried again later: still rejected, not silently pending.
	if got := m.State(id); got != pairing.StateRejected {
		t.Errorf("later State() = %v, want %v (durable/queryable)", got, pairing.StateRejected)
	}
}

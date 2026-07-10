package jitter

import "testing"

// ---------------------------------------------------------------------------
// T1 — construction, frame type, in-order push/pop
// ---------------------------------------------------------------------------

func TestNew_DefaultCapacity(t *testing.T) {
	b := New(DefaultCapacity)
	if got := b.Len(); got != 0 {
		t.Fatalf("fresh buffer Len() = %d, want 0", got)
	}
	for i := 1; i <= DefaultCapacity; i++ {
		if res := b.Push(Frame{Seq: i}); res != OK {
			t.Fatalf("push %d within default capacity: got %v, want OK", i, res)
		}
	}
	if res := b.Push(Frame{Seq: DefaultCapacity + 1}); res != Overrun {
		t.Fatalf("push beyond default capacity: got %v, want Overrun", res)
	}
}

func TestNew_InjectableCapacity(t *testing.T) {
	b := New(8)
	for i := 1; i <= 8; i++ {
		if res := b.Push(Frame{Seq: i}); res != OK {
			t.Fatalf("push %d within injected capacity 8: got %v, want OK", i, res)
		}
	}
	if res := b.Push(Frame{Seq: 9}); res != Overrun {
		t.Fatalf("9th push into capacity-8 buffer: got %v, want Overrun", res)
	}
}

func TestPushPop_InOrder_EmitsInOrder(t *testing.T) {
	b := New(8)
	for _, seq := range []int{1, 2, 3} {
		if res := b.Push(Frame{Seq: seq}); res != OK {
			t.Fatalf("push seq %d: got %v, want OK", seq, res)
		}
	}
	for _, want := range []int{1, 2, 3} {
		f, ok := b.Pop()
		if !ok {
			t.Fatalf("pop for seq %d: ok=false, want true", want)
		}
		if f.Seq != want {
			t.Fatalf("pop seq = %d, want %d", f.Seq, want)
		}
	}
}

func TestPop_AdvancesCursor(t *testing.T) {
	b := New(8)
	b.Push(Frame{Seq: 1})
	b.Push(Frame{Seq: 2})

	f1, ok := b.Pop()
	if !ok || f1.Seq != 1 {
		t.Fatalf("first pop = (%d,%v), want (1,true)", f1.Seq, ok)
	}
	f2, ok := b.Pop()
	if !ok || f2.Seq != 2 {
		t.Fatalf("second pop = (%d,%v), want (2,true) — cursor did not advance", f2.Seq, ok)
	}
}

// ---------------------------------------------------------------------------
// T2 — reordering + late/duplicate drops
// ---------------------------------------------------------------------------

func TestPush_OutOfOrder_EmitsInPlayoutOrder(t *testing.T) {
	b := New(8)
	for _, seq := range []int{3, 1, 2} {
		if res := b.Push(Frame{Seq: seq}); res != OK {
			t.Fatalf("push seq %d: got %v, want OK", seq, res)
		}
	}
	for _, want := range []int{1, 2, 3} {
		f, ok := b.Pop()
		if !ok || f.Seq != want {
			t.Fatalf("playout pop = (%d,%v), want (%d,true)", f.Seq, ok, want)
		}
	}
}

func TestPush_LateFrame_DroppedNotReordered(t *testing.T) {
	b := New(8)
	b.Push(Frame{Seq: 1})
	b.Push(Frame{Seq: 2})
	if f, ok := b.Pop(); !ok || f.Seq != 1 {
		t.Fatalf("pop1 = (%d,%v), want (1,true)", f.Seq, ok)
	}
	if f, ok := b.Pop(); !ok || f.Seq != 2 {
		t.Fatalf("pop2 = (%d,%v), want (2,true)", f.Seq, ok)
	}

	// Cursor is now at 3; a seq-1 frame is late.
	if res := b.Push(Frame{Seq: 1}); res != DroppedLate {
		t.Fatalf("late seq-1 push: got %v, want DroppedLate", res)
	}

	// Cursor must be unchanged (never moves backward): the next in-order frame
	// is seq 3, and the late seq 1 must never be emitted.
	b.Push(Frame{Seq: 3})
	f, ok := b.Pop()
	if !ok || f.Seq != 3 {
		t.Fatalf("post-late pop = (%d,%v), want (3,true) — late frame was reordered or cursor moved", f.Seq, ok)
	}
}

func TestPush_DuplicateBuffered_Dropped(t *testing.T) {
	b := New(8)
	if res := b.Push(Frame{Seq: 2}); res != OK {
		t.Fatalf("first seq-2 push: got %v, want OK", res)
	}
	if res := b.Push(Frame{Seq: 2}); res != DroppedDuplicate {
		t.Fatalf("second seq-2 push while buffered: got %v, want DroppedDuplicate", res)
	}

	// Exactly one emission of seq 2 on drain (bounded so a broken Pop can't hang).
	count := 0
	for i := 0; i < 8; i++ {
		f, ok := b.Pop()
		if !ok {
			break
		}
		if f.Seq == 2 {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("seq-2 emissions = %d, want exactly 1", count)
	}
}

func TestPush_DuplicateAlreadyPlayed_Dropped(t *testing.T) {
	b := New(8)
	b.Push(Frame{Seq: 2})
	if f, ok := b.Pop(); !ok || f.Seq != 2 {
		t.Fatalf("pop = (%d,%v), want (2,true)", f.Seq, ok)
	}

	// Cursor is now past 2; re-offering seq 2 is late, not a buffered duplicate.
	if res := b.Push(Frame{Seq: 2}); res != DroppedLate {
		t.Fatalf("already-played seq-2 push: got %v, want DroppedLate", res)
	}
}

// ---------------------------------------------------------------------------
// T3 — underrun + overrun reporting
// ---------------------------------------------------------------------------

func TestPop_Empty_ReportsUnderrun(t *testing.T) {
	b := New(8)
	f, ok := b.Pop()
	if ok {
		t.Fatalf("pop on empty buffer: ok=true, want false")
	}
	if f != (Frame{}) {
		t.Fatalf("pop on empty buffer: frame=%+v, want zero Frame{}", f)
	}
	if got := b.Underruns(); got != 1 {
		t.Fatalf("Underruns() = %d, want 1", got)
	}
}

func TestPush_AtCapacity_ReportsOverrun(t *testing.T) {
	b := New(2)
	b.Push(Frame{Seq: 1})
	b.Push(Frame{Seq: 2})
	if res := b.Push(Frame{Seq: 3}); res != Overrun {
		t.Fatalf("push at capacity: got %v, want Overrun", res)
	}
	if got := b.Overruns(); got != 1 {
		t.Fatalf("Overruns() = %d, want 1", got)
	}
}

func TestPush_Overrun_LeavesBufferIntact(t *testing.T) {
	b := New(2)
	b.Push(Frame{Seq: 1})
	b.Push(Frame{Seq: 2})
	b.Push(Frame{Seq: 3}) // overrun — must not evict or corrupt

	f1, ok1 := b.Pop()
	if !ok1 || f1.Seq != 1 {
		t.Fatalf("pop1 after overrun = (%d,%v), want (1,true)", f1.Seq, ok1)
	}
	f2, ok2 := b.Pop()
	if !ok2 || f2.Seq != 2 {
		t.Fatalf("pop2 after overrun = (%d,%v), want (2,true)", f2.Seq, ok2)
	}
}

func TestUnderrunOverrun_Distinguishable(t *testing.T) {
	under := New(4)
	under.Pop() // trigger one underrun

	over := New(1)
	over.Push(Frame{Seq: 1})
	over.Push(Frame{Seq: 2}) // trigger one overrun

	if got := under.Underruns(); got != 1 {
		t.Fatalf("under.Underruns() = %d, want 1", got)
	}
	if got := under.Overruns(); got != 0 {
		t.Fatalf("under.Overruns() = %d, want 0", got)
	}
	if got := over.Overruns(); got != 1 {
		t.Fatalf("over.Overruns() = %d, want 1", got)
	}
	if got := over.Underruns(); got != 0 {
		t.Fatalf("over.Underruns() = %d, want 0", got)
	}
}

---
type: plan
story: S1-02
scope: "tests only"
---

# Test plan — S1-02 jitter-buffer (tests only)

One checkbox = one test; every bullet names a real `path::fn`. Covers acceptance AND the
edge/failure cases the intent names: out-of-order reorder, late-frame dropped and never
reordered, duplicate dropped, pop-empty underrun reported, push-at-capacity overrun
reported, underrun/overrun distinguishable. All tests live in `internal/jitter/jitter_test.go`.

## T1 — construction, frame type, in-order push/pop

- [ ] `internal/jitter/jitter_test.go::TestNew_DefaultCapacity` — input: `New` with the documented default; act: inspect a fresh buffer; assert: `Len()==0` and capacity equals the default constant (D6).
- [ ] `internal/jitter/jitter_test.go::TestNew_InjectableCapacity` — input: `New(8)`; act: push 8 frames; assert: the 9th push reports `Overrun` (capacity is the injected value, D6).
- [ ] `internal/jitter/jitter_test.go::TestPushPop_InOrder_EmitsInOrder` — input: push frames seq 1,2,3; act: `Pop` three times; assert: emitted seqs are 1,2,3 in order (success #7 baseline).
- [ ] `internal/jitter/jitter_test.go::TestPop_AdvancesCursor` — input: push seq 1,2; act: `Pop` once; assert: first pop is seq 1 and the next pop is seq 2 (cursor advanced).

## T2 — reordering + late/duplicate drops

- [ ] `internal/jitter/jitter_test.go::TestPush_OutOfOrder_EmitsInPlayoutOrder` — input: push seq 3,1,2; act: `Pop` three times; assert: emitted 1,2,3 in playout order (success #7).
- [ ] `internal/jitter/jitter_test.go::TestPush_LateFrame_DroppedNotReordered` — input: push+pop through seq 2 (cursor now 3); act: `Push(Frame{Seq:1})`; assert: result `DroppedLate`, seq 1 never emitted, cursor unchanged (never moves backward) (failure #8; success #8).
- [ ] `internal/jitter/jitter_test.go::TestPush_DuplicateBuffered_Dropped` — input: `Push(Frame{Seq:2})` while seq 2 is still buffered; act: `Push(Frame{Seq:2})` again; assert: second result `DroppedDuplicate`, exactly one emission of seq 2 on drain (failure #9; success #9).
- [ ] `internal/jitter/jitter_test.go::TestPush_DuplicateAlreadyPlayed_Dropped` — input: push+pop seq 2 (already emitted); act: `Push(Frame{Seq:2})`; assert: result `DroppedLate` (already played, seq < cursor), not re-emitted (failure #9).

## T3 — underrun + overrun reporting

- [ ] `internal/jitter/jitter_test.go::TestPop_Empty_ReportsUnderrun` — input: fresh empty buffer; act: `Pop()`; assert: returns `(zero, false)` and `Underruns()==1` — distinguishable from a legitimate read (failure #11; success #10).
- [ ] `internal/jitter/jitter_test.go::TestPush_AtCapacity_ReportsOverrun` — input: `New(2)` filled with 2 frames; act: `Push` a third; assert: result `Overrun` and `Overruns()==1` (failure #10; success #10).
- [ ] `internal/jitter/jitter_test.go::TestPush_Overrun_LeavesBufferIntact` — input: `New(2)` filled with seq 1,2; act: `Push(Frame{Seq:3})` (overrun); assert: buffer still holds seq 1,2 (no arbitrary eviction, no corruption) and popping yields 1,2 (failure #10).
- [ ] `internal/jitter/jitter_test.go::TestUnderrunOverrun_Distinguishable` — input: trigger one underrun (pop empty) and one overrun (push at capacity) on separate buffers; act: query both counters; assert: `Underruns()` and `Overruns()` are independent and reflect the right condition each (success #10).

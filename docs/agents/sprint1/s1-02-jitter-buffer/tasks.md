---
type: tasks
story: S1-02
---

# Tasks — S1-02 jitter-buffer

Wave 1 (parallel). Package `internal/jitter`. Stdlib only; imports at most `internal/core`
(synthetic frame descriptors — no real audio). Pure value logic, deterministic given a
scripted push/pop sequence — no goroutines, no wall-clock. Decoupled from `internal/drift`
(D7).

Surface: a `Frame{Seq uint64; Timestamp uint64}` descriptor; `jitter.New(capacity int)
*Buffer` with a documented default capacity constant (D6); `Push(Frame) PushResult` where
`PushResult` distinguishes `Accepted`/`DroppedLate`/`DroppedDuplicate`/`Overrun`; `Pop()
(Frame, bool)` emitting in playout order with `false` signalling underrun; and queryable
`Underruns() int`, `Overruns() int`, `Len() int`. The playout cursor is the next expected
seq and never moves backward.

## T1 — construction, frame type, in-order push/pop

Define `Frame`, `PushResult`, `Buffer`, `New(capacity)` (+ a documented default), and the
happy-path `Push`/`Pop`/`Len`. A fresh buffer is empty; frames pushed in order are popped in
order and each pop advances the cursor (README §4; intake success #7 baseline).

- Atomic: the types + constructor + in-order enqueue/dequeue + `Len`.
- Testable: `jitter_test.go` covers default-capacity, injectable-capacity, in-order-emits-
  in-order, and pop-advances-cursor.

## T2 — reordering + late/duplicate drops

`Push` accepts out-of-order frames and `Pop` emits them in playout (ascending-seq) order. A
frame with `seq < cursor` is `DroppedLate` — never emitted, and the cursor never moves
backward (intake failure #8; success #8). A frame whose seq is already buffered, or was
already emitted, is `DroppedDuplicate`/`DroppedLate` and never emitted twice (intake failure
#9; success #9).

- Atomic: the ordered-insert + the late-drop guard (cursor monotonic) + the duplicate guard.
- Testable: `jitter_test.go` covers out-of-order-emits-in-playout-order, late-frame-dropped-
  not-reordered, duplicate-buffered-dropped, and duplicate-already-played-dropped.

## T3 — underrun + overrun reporting

`Pop` on an empty buffer (nothing at/after the cursor) returns `(zero, false)` and
increments a queryable underrun counter — distinguishable from an ordinary read (intake
failure #11; success #10). `Push` when `Len == capacity` returns `Overrun`, increments a
queryable overrun counter, and leaves buffer contents uncorrupted (intake failure #10;
success #10). Underrun and overrun are separate, independently observable conditions.

- Atomic: the pop-empty underrun path + the push-at-capacity overrun path + the two counters.
- Testable: `jitter_test.go` covers pop-empty-reports-underrun, push-at-capacity-reports-
  overrun, overrun-leaves-buffer-intact, and underrun-overrun-distinguishable.

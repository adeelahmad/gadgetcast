package jitter

// DefaultCapacity is the documented default jitter-buffer capacity used when a
// caller wants the standard playout window.
const DefaultCapacity = 64

// PushResult reports the disposition of a Push.
type PushResult int

const (
	// OK means the frame was buffered for playout.
	OK PushResult = iota
	// Overrun means the buffer was full and the frame was rejected.
	Overrun
	// DroppedLate means the frame's seq is at or before the play cursor.
	DroppedLate
	// DroppedDuplicate means an equal seq is already buffered.
	DroppedDuplicate
)

// Frame is a single sequenced media frame.
type Frame struct {
	Seq int
}

// Buffer is a reordering jitter buffer.
type Buffer struct {
	capacity int
}

// New builds a Buffer with the given capacity.
func New(capacity int) *Buffer {
	panic("SUB-AGENT-TODO: construct a Buffer with the given capacity and an initialized play cursor")
}

// Push buffers a frame for playout, returning its disposition.
func (b *Buffer) Push(f Frame) PushResult {
	panic("SUB-AGENT-TODO: buffer frame in playout order; reject at capacity (Overrun), drop seq<=cursor (DroppedLate) and buffered dup (DroppedDuplicate)")
}

// Pop returns the next frame in playout order and advances the cursor.
func (b *Buffer) Pop() (Frame, bool) {
	panic("SUB-AGENT-TODO: emit next in-order frame and advance cursor; on empty return (Frame{}, false) and count an underrun")
}

// Len is the number of buffered frames.
func (b *Buffer) Len() int {
	panic("SUB-AGENT-TODO: return the count of currently buffered frames")
}

// Underruns counts pops on an empty buffer.
func (b *Buffer) Underruns() int {
	panic("SUB-AGENT-TODO: return the cumulative count of pops on an empty buffer")
}

// Overruns counts pushes rejected at capacity.
func (b *Buffer) Overruns() int {
	panic("SUB-AGENT-TODO: return the cumulative count of pushes rejected at capacity")
}

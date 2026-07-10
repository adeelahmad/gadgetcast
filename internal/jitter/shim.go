// agentic:shim
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
	return &Buffer{capacity: capacity}
}

// Push buffers a frame for playout.
//
// SHIM: always reports OK regardless of state.
func (b *Buffer) Push(f Frame) PushResult {
	return OK
}

// Pop returns the next frame in playout order.
//
// SHIM: always reports a zero frame as present.
func (b *Buffer) Pop() (Frame, bool) {
	return Frame{}, true
}

// Len is the number of buffered frames.
//
// SHIM: always zero.
func (b *Buffer) Len() int { return 0 }

// Underruns counts pops on an empty buffer.
//
// SHIM: always zero.
func (b *Buffer) Underruns() int { return 0 }

// Overruns counts pushes rejected at capacity.
//
// SHIM: always zero.
func (b *Buffer) Overruns() int { return 0 }

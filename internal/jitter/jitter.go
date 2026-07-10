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
	capacity  int
	cursor    int
	frames    map[int]Frame
	underruns int
	overruns  int
}

// New builds a Buffer with the given capacity.
func New(capacity int) *Buffer {
	if capacity <= 0 {
		capacity = DefaultCapacity
	}
	return &Buffer{
		capacity: capacity,
		frames:   make(map[int]Frame),
	}
}

// Push buffers a frame for playout, returning its disposition.
func (b *Buffer) Push(f Frame) PushResult {
	if f.Seq < b.cursor {
		return DroppedLate
	}
	if _, ok := b.frames[f.Seq]; ok {
		return DroppedDuplicate
	}
	if len(b.frames) >= b.capacity {
		b.overruns++
		return Overrun
	}
	b.frames[f.Seq] = f
	return OK
}

// Pop returns the next frame in playout order and advances the cursor.
func (b *Buffer) Pop() (Frame, bool) {
	if len(b.frames) == 0 {
		b.underruns++
		return Frame{}, false
	}
	next := 0
	first := true
	for seq := range b.frames {
		if first || seq < next {
			next = seq
			first = false
		}
	}
	f := b.frames[next]
	delete(b.frames, next)
	b.cursor = next + 1
	return f, true
}

// Len is the number of buffered frames.
func (b *Buffer) Len() int {
	return len(b.frames)
}

// Underruns counts pops on an empty buffer.
func (b *Buffer) Underruns() int {
	return b.underruns
}

// Overruns counts pushes rejected at capacity.
func (b *Buffer) Overruns() int {
	return b.overruns
}

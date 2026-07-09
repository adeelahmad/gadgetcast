package signaling

// DecodeError is the typed error Decode returns for any malformed / unknown input.
type DecodeError struct {
	Type   MessageType
	Reason string
	Err    error
}

func (e *DecodeError) Error() string {
	panic("SUB-AGENT-TODO: GREEN formats the decode error message")
}

func (e *DecodeError) Unwrap() error {
	panic("SUB-AGENT-TODO: GREEN returns the wrapped underlying error")
}

// Encode serializes a Message into a typed JSON envelope.
func Encode(msg Message) ([]byte, error) {
	panic("SUB-AGENT-TODO: GREEN marshals msg into an Envelope{Type, Payload} JSON document")
}

// Decode parses a JSON envelope back into its concrete Message, returning a typed
// *DecodeError for unknown types, truncated/malformed JSON, or type/payload mismatch.
func Decode(data []byte) (Message, error) {
	panic("SUB-AGENT-TODO: GREEN unmarshals the envelope, dispatches on Type to the concrete message, and returns (nil, *DecodeError) on any malformed/unknown/mismatched input without panicking")
}

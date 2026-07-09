package core

import "errors"

// ErrInvalidClientID is the sentinel returned when a ClientID cannot be built.
var ErrInvalidClientID = errors.New("core: invalid client id")

// ClientID is a strongly-typed, comparable client identity (usable as a map key
// and with ==). The zero value is invalid and must not masquerade as a real client.
type ClientID struct {
	value string
}

// NewClientID builds a ClientID from raw input.
func NewClientID(_ string) (ClientID, error) {
	panic("SUB-AGENT-TODO: trim input; reject empty/whitespace-only with ErrInvalidClientID returning zero ClientID; otherwise return ClientID carrying the trimmed value")
}

// String returns the underlying client identity string.
func (c ClientID) String() string {
	panic("SUB-AGENT-TODO: return the stored value")
}

// IsValid reports whether the ClientID holds a real (non-empty) identity.
func (c ClientID) IsValid() bool {
	panic("SUB-AGENT-TODO: return true only when value is non-empty (zero value must be false)")
}

// IsZero reports whether the ClientID is the zero value.
func (c ClientID) IsZero() bool {
	panic("SUB-AGENT-TODO: return true when value is empty (the zero value)")
}

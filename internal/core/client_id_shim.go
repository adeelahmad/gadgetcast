// agentic:shim
package core

import "errors"

// ErrInvalidClientID is the sentinel returned when a ClientID cannot be built.
var ErrInvalidClientID = errors.New("core: invalid client id")

// ClientID is a strongly-typed, comparable client identity (usable as a map key).
type ClientID struct {
	value string
}

// NewClientID is a RED shim: it never rejects and never carries the input, so
// every behavioral assertion fails on the value, not on a missing symbol.
func NewClientID(_ string) (ClientID, error) {
	return ClientID{}, nil
}

// String is a RED shim: always empty.
func (c ClientID) String() string {
	return ""
}

// IsValid is a RED shim: always false.
func (c ClientID) IsValid() bool {
	return false
}

// IsZero is a RED shim: always false.
func (c ClientID) IsZero() bool {
	return false
}

package core

import (
	"errors"
	"strings"
)

// ErrInvalidClientID is the sentinel returned when a ClientID cannot be built.
var ErrInvalidClientID = errors.New("core: invalid client id")

// ClientID is a strongly-typed, comparable client identity (usable as a map key
// and with ==). The zero value is invalid and must not masquerade as a real client.
type ClientID struct {
	value string
}

// NewClientID builds a ClientID from raw input.
func NewClientID(raw string) (ClientID, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ClientID{}, ErrInvalidClientID
	}
	return ClientID{value: trimmed}, nil
}

// String returns the underlying client identity string.
func (c ClientID) String() string {
	return c.value
}

// IsValid reports whether the ClientID holds a real (non-empty) identity.
func (c ClientID) IsValid() bool {
	return c.value != ""
}

// IsZero reports whether the ClientID is the zero value.
func (c ClientID) IsZero() bool {
	return c.value == ""
}

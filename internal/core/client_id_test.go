package core

import (
	"errors"
	"testing"
)

func TestNewClientID_Valid(t *testing.T) {
	const in = "cam-1"

	id, err := NewClientID(in)
	if err != nil {
		t.Fatalf("NewClientID(%q) returned unexpected error: %v", in, err)
	}
	if got := id.String(); got != in {
		t.Errorf("id.String() = %q, want %q", got, in)
	}
	if !id.IsValid() {
		t.Errorf("id.IsValid() = false, want true for %q", in)
	}
}

func TestNewClientID_EmptyRejected(t *testing.T) {
	id, err := NewClientID("")
	if !errors.Is(err, ErrInvalidClientID) {
		t.Errorf("NewClientID(%q) error = %v, want ErrInvalidClientID", "", err)
	}
	if id != (ClientID{}) {
		t.Errorf("NewClientID(%q) returned non-zero ClientID %#v, want zero value", "", id)
	}
}

func TestNewClientID_WhitespaceRejected(t *testing.T) {
	cases := []string{"   ", "\t", "\n", " \t\n "}
	for _, in := range cases {
		in := in
		t.Run("blank", func(t *testing.T) {
			id, err := NewClientID(in)
			if !errors.Is(err, ErrInvalidClientID) {
				t.Errorf("NewClientID(%q) error = %v, want ErrInvalidClientID", in, err)
			}
			if id != (ClientID{}) {
				t.Errorf("NewClientID(%q) returned non-zero ClientID %#v, want zero value", in, id)
			}
		})
	}
}

func TestClientID_EqualityAndMapKey(t *testing.T) {
	a1, err := NewClientID("a")
	if err != nil {
		t.Fatalf("NewClientID(%q): %v", "a", err)
	}
	a2, err := NewClientID("a")
	if err != nil {
		t.Fatalf("NewClientID(%q): %v", "a", err)
	}
	b, err := NewClientID("b")
	if err != nil {
		t.Fatalf("NewClientID(%q): %v", "b", err)
	}

	if a1 != a2 {
		t.Errorf("two ClientIDs from %q are not ==: %#v vs %#v", "a", a1, a2)
	}
	if a1 == b {
		t.Errorf("ClientIDs from %q and %q compare ==, want distinct", "a", "b")
	}

	m := map[ClientID]int{}
	m[a1] = 1
	m[a2] = 2 // same identity as a1 -> must overwrite the same slot
	m[b] = 3

	if len(m) != 2 {
		t.Errorf("map has %d keys, want 2 (a1/a2 share a slot, b is distinct)", len(m))
	}
	if m[a1] != 2 {
		t.Errorf("m[a1] = %d, want 2 (a2 must index the same slot as a1)", m[a1])
	}
	if m[b] != 3 {
		t.Errorf("m[b] = %d, want 3 (b must index its own slot)", m[b])
	}
}

func TestClientID_ZeroValueInvalid(t *testing.T) {
	var zero ClientID

	if !zero.IsZero() {
		t.Errorf("ClientID{}.IsZero() = false, want true")
	}
	if zero.IsValid() {
		t.Errorf("ClientID{}.IsValid() = true, want false (zero must not masquerade as a real client)")
	}
}

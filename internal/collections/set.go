package collections

import (
	"encoding"
	"strings"
)

// Set is a generic set collection using a map-like interface.
type Set[
	T interface {
		comparable
		encoding.TextMarshaler
	},
	PT interface {
		*T
		encoding.TextUnmarshaler
	},
] map[T]struct{}

// Add adds an element to the set. If the element is already in the set, it has no effect.
func (s Set[T, P]) Add(item T) {
	s[item] = struct{}{}
}

// Remove removes an element from the set. If the element is not in the set, it has no effect.
func (s Set[T, P]) Remove(item T) {
	delete(s, item)
}

// Contains reports whether an element is in the set.
func (s Set[T, P]) Contains(item T) bool {
	_, ok := s[item]
	return ok
}

// MarshalText implements encoding.TextMarshaler, returning comma-separated encoded elements.
func (s Set[T, PT]) MarshalText() ([]byte, error) {
	if len(s) == 0 {
		return []byte{}, nil
	}

	var parts []string
	for elem := range s {
		data, err := elem.MarshalText()
		if err != nil {
			return nil, err
		}
		parts = append(parts, string(data))
	}

	return []byte(strings.Join(parts, ",")), nil
}

// UnmarshalText implements encoding.TextUnmarshaler, parsing comma-separated elements.
func (s *Set[T, PT]) UnmarshalText(text []byte) error {
	str := string(text)
	if str == "" {
		*s = make(Set[T, PT])
		return nil
	}

	parts := strings.Split(str, ",")
	*s = make(Set[T, PT], len(parts))

	for _, part := range parts {
		var item T
		var ptr PT = &item
		if err := ptr.UnmarshalText([]byte(strings.TrimSpace(part))); err != nil {
			return err
		}
		s.Add(item)
	}

	return nil
}

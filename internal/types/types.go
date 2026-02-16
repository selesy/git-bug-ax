package types

import (
	"encoding"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/git-bug/git-bug/entity"
)

// TextCodec combines encoding.TextMarshaler and encoding.TextUnmarshaler
// for types that support bidirectional text serialization.
type TextCodec interface {
	encoding.TextMarshaler
	encoding.TextUnmarshaler
}

var _ TextCodec = (*ID)(nil)

// ID wraps git-bug's entity.Id and implements TextCodec for
// bidirectional text serialization.
type ID struct {
	entity.Id
}

// NewID creates an ID from a full 64-character hex hash.
// It accepts both uppercase and lowercase hex characters,
// normalizing to lowercase for storage.
// It returns an error if the hash is not a valid hex string or fails validation.
func NewID(hash string) (ID, error) {
	lower := strings.ToLower(hash)
	// TODO: Submit a bug report to git-bug and remove this when Validate() works
	if _, err := hex.DecodeString(lower); err != nil {
		return ID{}, fmt.Errorf("invalid hex in hash: %w", err)
	}
	id := entity.Id(lower)
	if err := id.Validate(); err != nil {
		return ID{}, fmt.Errorf("invalid hash: %w", err)
	}
	return ID{Id: id}, nil
}

// MarshalText implements encoding.TextMarshaler, returning the ID as a hex-encoded string.
func (id ID) MarshalText() ([]byte, error) {
	return []byte(id.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler, parsing a hex-encoded ID string.
func (id *ID) UnmarshalText(text []byte) error {
	parsed, err := NewID(string(text))
	if err != nil {
		return err
	}
	*id = parsed
	return nil
}

var _ TextCodec = (*IDs)(nil)

// IDs is a slice of ID values that implements TextCodec with
// comma-separated serialization.
type IDs []ID

// MarshalText implements encoding.TextMarshaler, returning comma-separated ID hashes.
func (ids IDs) MarshalText() ([]byte, error) {
	if len(ids) == 0 {
		return []byte{}, nil
	}
	strs := make([]string, len(ids))
	for i, id := range ids {
		strs[i] = id.String()
	}
	return []byte(strings.Join(strs, ",")), nil
}

// UnmarshalText implements encoding.TextUnmarshaler, parsing comma-separated ID hashes.
func (ids *IDs) UnmarshalText(text []byte) error {
	s := string(text)
	if s == "" {
		*ids = nil
		return nil
	}
	parts := strings.Split(s, ",")
	result := make(IDs, len(parts))
	for i, part := range parts {
		parsed, err := NewID(strings.TrimSpace(part))
		if err != nil {
			return fmt.Errorf("invalid ID at position %d: %w", i, err)
		}
		result[i] = parsed
	}
	*ids = result
	return nil
}

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

	for i, part := range parts {
		var item T
		var ptr PT = &item
		if err := ptr.UnmarshalText([]byte(strings.TrimSpace(part))); err != nil {
			return fmt.Errorf("invalid element at position %d: %w", i, err)
		}
		s.Add(item)
	}

	return nil
}

var _ TextCodec = (*Label)(nil)

// Label is a string type that implements TextCodec for bidirectional text serialization.
type Label string

// MarshalText implements encoding.TextMarshaler, returning the label as a string.
func (l Label) MarshalText() ([]byte, error) {
	return []byte(l), nil
}

// UnmarshalText implements encoding.TextUnmarshaler, parsing a label string.
func (l *Label) UnmarshalText(text []byte) error {
	*l = Label(text)
	return nil
}

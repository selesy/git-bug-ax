package entity

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/git-bug/git-bug/entity"
)

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

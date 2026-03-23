package issue

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/git-bug/git-bug/entity"

	"github.com/selesy/git-bug-agent/internal/codec"
)

// ID wraps git-bug's entity.Id and implements TextCodec for
// bidirectional text serialization.
type ID struct {
	entity.Id
}

var _ codec.TextCodec = (*ID)(nil)

// NewID creates an ID from a full 64-character hex hash.
// It accepts both uppercase and lowercase hex characters,
// normalizing to lowercase for storage.
func WrapID(id entity.Id) (ID, error) {
	return ID{Id: id}, nil
}

// ParseID validates a git-bug hash and wraps it in an issue.ID.
func ParseID(hash string) (ID, error) {
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

// MarshalText implements encoding.TextMarshaler.
func (id ID) MarshalText() ([]byte, error) {
	return []byte(id.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (id *ID) UnmarshalText(text []byte) error {
	parsed, err := ParseID(string(text))
	if err != nil {
		return err
	}
	*id = parsed

	return nil
}

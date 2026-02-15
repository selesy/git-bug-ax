package issue

import (
	"fmt"
	"strings"

	"github.com/selesy/git-bug-ax/internal/types"
)

// Priority represents the priority level of an issue.
// Use one of the exported Priority* constants.
type Priority struct {
	value string
}

var (
	PriorityHighest = Priority{"highest"}
	PriorityHigh    = Priority{"high"}
	PriorityMedium  = Priority{"medium"}
	PriorityLow     = Priority{"low"}
	PriorityLowest  = Priority{"lowest"}
)

// MarshalText implements encoding.TextMarshaler.
func (p Priority) MarshalText() ([]byte, error) {
	return []byte(p.value), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (p *Priority) UnmarshalText(text []byte) error {
	lower := strings.ToLower(string(text))
	switch lower {
	case "highest":
		*p = PriorityHighest
	case "high":
		*p = PriorityHigh
	case "medium":
		*p = PriorityMedium
	case "low":
		*p = PriorityLow
	case "lowest":
		*p = PriorityLowest
	default:
		return fmt.Errorf("invalid priority: %q", lower)
	}
	return nil
}

// String implements fmt.Stringer.
func (p Priority) String() string {
	return p.value
}

// Verify that Priority implements TextCodec.
var _ types.TextCodec = (*Priority)(nil)

// ParsePriority returns a Priority from a string, or an error if the string is invalid.
func ParsePriority(s string) (Priority, error) {
	var p Priority
	if err := p.UnmarshalText([]byte(s)); err != nil {
		return Priority{}, err
	}
	return p, nil
}

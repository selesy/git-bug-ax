package issue

import (
	"fmt"
	"strings"

	"github.com/selesy/git-bug-agent/internal/codec"
)

// Priority represents the priority level of an issue.
// Use one of the exported Priority* constants.
type Priority struct {
	value string
}

var (
	// PriorityHighest represents the highest priority level.
	PriorityHighest = Priority{"highest"}
	// PriorityHigh represents a high priority level.
	PriorityHigh = Priority{"high"}
	// PriorityMedium represents a medium priority level.
	PriorityMedium = Priority{"medium"}
	// PriorityLow represents a low priority level.
	PriorityLow = Priority{"low"}
	// PriorityLowest represents the lowest priority level.
	PriorityLowest = Priority{"lowest"}
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

// String implements fmt.Stringer, returning the string representation of the priority.
func (p Priority) String() string {
	return p.value
}

// Verify that Priority implements TextCodec.
var _ codec.TextCodec = (*Priority)(nil)

// ParsePriority returns a Priority from a string, or an error if the string is invalid.
// Valid values are: "highest", "high", "medium", "low", "lowest" (case-insensitive).
func ParsePriority(s string) (Priority, error) {
	var p Priority
	if err := p.UnmarshalText([]byte(s)); err != nil {
		return Priority{}, err
	}
	return p, nil
}

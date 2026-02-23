package issue

import (
	"fmt"
	"strings"

	"github.com/selesy/git-bug-ax/internal/codec"
)

// Resolution represents the resolution of an issue.
// Use one of the exported Resolution* constants.
type Resolution struct {
	value string
}

var (
	// ResolutionFixed represents an issue that has been fixed.
	ResolutionFixed = Resolution{"fixed"}
	// ResolutionWontFix represents an issue that will not be fixed.
	ResolutionWontFix = Resolution{"wont-fix"}
	// ResolutionDuplicate represents an issue that is a duplicate of another.
	ResolutionDuplicate = Resolution{"duplicate"}
	// ResolutionCannotRepro represents an issue that cannot be reproduced.
	ResolutionCannotRepro = Resolution{"cannot-repro"}
	// ResolutionNotNeeded represents an issue that is no longer needed.
	ResolutionNotNeeded = Resolution{"not-needed"}
)

// MarshalText implements encoding.TextMarshaler.
func (r Resolution) MarshalText() ([]byte, error) {
	return []byte(r.value), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (r *Resolution) UnmarshalText(text []byte) error {
	lower := strings.ToLower(string(text))
	switch lower {
	case "fixed":
		*r = ResolutionFixed
	case "wont-fix":
		*r = ResolutionWontFix
	case "duplicate":
		*r = ResolutionDuplicate
	case "cannot-repro":
		*r = ResolutionCannotRepro
	case "not-needed":
		*r = ResolutionNotNeeded
	default:
		return fmt.Errorf("invalid resolution: %q", lower)
	}
	return nil
}

// String implements fmt.Stringer, returning the string representation of the resolution.
func (r Resolution) String() string {
	return r.value
}

// Verify that Resolution implements TextCodec.
var _ codec.TextCodec = (*Resolution)(nil)

// ParseResolution returns a Resolution from a string, or an error if the string is invalid.
// Valid values are: "fixed", "wont-fix", "duplicate", "cannot-repro", "not-needed"
// (case-insensitive).
func ParseResolution(s string) (Resolution, error) {
	var r Resolution
	if err := r.UnmarshalText([]byte(s)); err != nil {
		return Resolution{}, err
	}
	return r, nil
}

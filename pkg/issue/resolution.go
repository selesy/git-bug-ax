package issue

import (
	"fmt"
	"strings"

	"github.com/selesy/git-bug-ax/internal/types"
)

// Resolution represents the resolution of an issue.
// Use one of the exported Resolution* constants.
type Resolution struct {
	value string
}

var (
	ResolutionFixed       = Resolution{"fixed"}
	ResolutionWontFix     = Resolution{"wont-fix"}
	ResolutionDuplicate   = Resolution{"duplicate"}
	ResolutionCannotRepro = Resolution{"cannot-repro"}
	ResolutionNotNeeded   = Resolution{"not-needed"}
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

// String implements fmt.Stringer.
func (r Resolution) String() string {
	return r.value
}

// Verify that Resolution implements TextCodec.
var _ types.TextCodec = (*Resolution)(nil)

// ParseResolution returns a Resolution from a string, or an error if the string is invalid.
func ParseResolution(s string) (Resolution, error) {
	var r Resolution
	if err := r.UnmarshalText([]byte(s)); err != nil {
		return Resolution{}, err
	}
	return r, nil
}

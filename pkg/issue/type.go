package issue

import (
	"fmt"
	"strings"

	"github.com/selesy/git-bug-ax/internal/types"
)

// Type represents the type of an issue.
// Use one of the exported Type* constants.
type Type struct {
	value string
}

var (
	TypeEpic     = Type{"epic"}
	TypeFeature  = Type{"feature"}
	TypeTask     = Type{"task"}
	TypeBug      = Type{"bug"}
	TypeSpike    = Type{"spike"}
	TypeTechDebt = Type{"tech-debt"}
	TypeFix      = Type{"fix"}
	TypeFeat     = Type{"feat"}
	TypeBuild    = Type{"build"}
	TypeChore    = Type{"chore"}
	TypeCI       = Type{"ci"}
	TypeDocs     = Type{"docs"}
	TypeStyle    = Type{"style"}
	TypeRefactor = Type{"refactor"}
	TypeTest     = Type{"test"}
	TypePerf     = Type{"perf"}
)

// MarshalText implements encoding.TextMarshaler.
func (t Type) MarshalText() ([]byte, error) {
	return []byte(t.value), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (t *Type) UnmarshalText(text []byte) error {
	lower := strings.ToLower(string(text))
	switch lower {
	case "epic":
		*t = TypeEpic
	case "feature":
		*t = TypeFeature
	case "task":
		*t = TypeTask
	case "bug":
		*t = TypeBug
	case "spike":
		*t = TypeSpike
	case "tech-debt":
		*t = TypeTechDebt
	case "fix":
		*t = TypeFix
	case "feat":
		*t = TypeFeat
	case "build":
		*t = TypeBuild
	case "chore":
		*t = TypeChore
	case "ci":
		*t = TypeCI
	case "docs":
		*t = TypeDocs
	case "style":
		*t = TypeStyle
	case "refactor":
		*t = TypeRefactor
	case "test":
		*t = TypeTest
	case "perf":
		*t = TypePerf
	default:
		return fmt.Errorf("invalid type: %q", lower)
	}
	return nil
}

// String implements fmt.Stringer.
func (t Type) String() string {
	return t.value
}

// Verify that Type implements TextCodec.
var _ types.TextCodec = (*Type)(nil)

// ParseType returns a Type from a string, or an error if the string is invalid.
func ParseType(s string) (Type, error) {
	var t Type
	if err := t.UnmarshalText([]byte(s)); err != nil {
		return Type{}, err
	}
	return t, nil
}

package issue

import (
	"fmt"
	"strings"

	"github.com/selesy/git-bug-ax/internal/codec"
)

// Type represents the type of an issue.
// Use one of the exported Type* constants.
type Type struct {
	value string
}

var (
	// TypeEpic represents a large body of work that spans multiple issues.
	TypeEpic = Type{"epic"}
	// TypeFeature represents a new feature request.
	TypeFeature = Type{"feature"}
	// TypeTask represents a general task or work item.
	TypeTask = Type{"task"}
	// TypeBug represents a bug or defect.
	TypeBug = Type{"bug"}
	// TypeSpike represents a time-boxed investigation or research task.
	TypeSpike = Type{"spike"}
	// TypeTechDebt represents technical debt that needs to be addressed.
	TypeTechDebt = Type{"tech-debt"}
	// TypeFix represents a bug fix.
	TypeFix = Type{"fix"}
	// TypeFeat represents a new feature (alternative naming).
	TypeFeat = Type{"feat"}
	// TypeBuild represents a build or deployment task.
	TypeBuild = Type{"build"}
	// TypeChore represents a maintenance or housekeeping task.
	TypeChore = Type{"chore"}
	// TypeCI represents a continuous integration task.
	TypeCI = Type{"ci"}
	// TypeDocs represents a documentation task.
	TypeDocs = Type{"docs"}
	// TypeStyle represents a code style or formatting task.
	TypeStyle = Type{"style"}
	// TypeRefactor represents a refactoring task.
	TypeRefactor = Type{"refactor"}
	// TypeTest represents a testing task.
	TypeTest = Type{"test"}
	// TypePerf represents a performance optimization task.
	TypePerf = Type{"perf"}
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

// String implements fmt.Stringer, returning the string representation of the type.
func (t Type) String() string {
	return t.value
}

// Verify that Type implements TextCodec.
var _ codec.TextCodec = (*Type)(nil)

// ParseType returns a Type from a string, or an error if the string is invalid.
// Valid values are: "epic", "feature", "task", "bug", "spike", "tech-debt", "fix",
// "feat", "build", "chore", "ci", "docs", "style", "refactor", "test", "perf"
// (case-insensitive).
func ParseType(s string) (Type, error) {
	var t Type
	if err := t.UnmarshalText([]byte(s)); err != nil {
		return Type{}, err
	}
	return t, nil
}

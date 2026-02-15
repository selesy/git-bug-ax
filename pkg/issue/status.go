package issue

import (
	"fmt"
	"strings"

	"github.com/selesy/git-bug-ax/internal/types"
)

// Status represents the status of an issue.
// Use one of the exported Status* constants.
type Status struct {
	value string
}

var (
	StatusDraft              = Status{"draft"}
	StatusReady              = Status{"ready"}
	StatusClaimed            = Status{"claimed"}
	StatusInProgress         = Status{"in-progress"}
	StatusBlocked            = Status{"blocked"}
	StatusReview             = Status{"review"}
	StatusDone               = Status{"done"}
	StatusAbandoned          = Status{"abandoned"}
	StatusFailed             = Status{"failed"}
	StatusStale              = Status{"stale"}
	StatusNeedsDecomposition = Status{"needs-decomposition"}
	StatusNeedsReplanning    = Status{"needs-replanning"}
	StatusContested          = Status{"contested"}
)

// MarshalText implements encoding.TextMarshaler.
func (s Status) MarshalText() ([]byte, error) {
	return []byte(s.value), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (s *Status) UnmarshalText(text []byte) error {
	lower := strings.ToLower(string(text))
	switch lower {
	case "draft":
		*s = StatusDraft
	case "ready":
		*s = StatusReady
	case "claimed":
		*s = StatusClaimed
	case "in-progress":
		*s = StatusInProgress
	case "blocked":
		*s = StatusBlocked
	case "review":
		*s = StatusReview
	case "done":
		*s = StatusDone
	case "abandoned":
		*s = StatusAbandoned
	case "failed":
		*s = StatusFailed
	case "stale":
		*s = StatusStale
	case "needs-decomposition":
		*s = StatusNeedsDecomposition
	case "needs-replanning":
		*s = StatusNeedsReplanning
	case "contested":
		*s = StatusContested
	default:
		return fmt.Errorf("invalid status: %q", lower)
	}
	return nil
}

// String implements fmt.Stringer.
func (s Status) String() string {
	return s.value
}

// Verify that Status implements TextCodec.
var _ types.TextCodec = (*Status)(nil)

// ParseStatus returns a Status from a string, or an error if the string is invalid.
func ParseStatus(s string) (Status, error) {
	var st Status
	if err := st.UnmarshalText([]byte(s)); err != nil {
		return Status{}, err
	}
	return st, nil
}

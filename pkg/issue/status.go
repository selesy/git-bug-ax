package issue

import (
	"fmt"
	"strings"

	"github.com/selesy/git-bug-ax/internal/codec"
)

// Status represents the status of an issue.
// Use one of the exported Status* constants.
type Status struct {
	value string
}

var (
	// StatusDraft represents a draft issue that is not yet ready.
	StatusDraft = Status{"draft"}
	// StatusReady represents an issue that is ready to be claimed.
	StatusReady = Status{"ready"}
	// StatusClaimed represents an issue that has been claimed by someone.
	StatusClaimed = Status{"claimed"}
	// StatusInProgress represents an issue that is currently being worked on.
	StatusInProgress = Status{"in-progress"}
	// StatusBlocked represents an issue that is blocked and cannot proceed.
	StatusBlocked = Status{"blocked"}
	// StatusReview represents an issue that is under review.
	StatusReview = Status{"review"}
	// StatusDone represents an issue that is completed.
	StatusDone = Status{"done"}
	// StatusAbandoned represents an issue that has been abandoned.
	StatusAbandoned = Status{"abandoned"}
	// StatusFailed represents an issue where work has failed.
	StatusFailed = Status{"failed"}
	// StatusStale represents an issue that is no longer current.
	StatusStale = Status{"stale"}
	// StatusNeedsDecomposition represents an issue that needs to be broken down.
	StatusNeedsDecomposition = Status{"needs-decomposition"}
	// StatusNeedsReplanning represents an issue that needs to be replanned.
	StatusNeedsReplanning = Status{"needs-replanning"}
	// StatusContested represents an issue that has conflicting opinions.
	StatusContested = Status{"contested"}
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

// String implements fmt.Stringer, returning the string representation of the status.
func (s Status) String() string {
	return s.value
}

// Verify that Status implements TextCodec.
var _ codec.TextCodec = (*Status)(nil)

// ParseStatus returns a Status from a string, or an error if the string is invalid.
// Valid values are: "draft", "ready", "claimed", "in-progress", "blocked", "review",
// "done", "abandoned", "failed", "stale", "needs-decomposition", "needs-replanning",
// "contested" (case-insensitive).
func ParseStatus(s string) (Status, error) {
	var st Status
	if err := st.UnmarshalText([]byte(s)); err != nil {
		return Status{}, err
	}
	return st, nil
}

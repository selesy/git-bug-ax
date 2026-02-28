package issue_test

import (
	"testing"

	"github.com/selesy/git-bug-ax/pkg/issue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatusConstants(t *testing.T) {
	tests := []struct {
		name   string
		status issue.Status
		want   string
	}{
		{"Draft", issue.StatusDraft, "draft"},
		{"Ready", issue.StatusReady, "ready"},
		{"Claimed", issue.StatusClaimed, "claimed"},
		{"InProgress", issue.StatusInProgress, "in-progress"},
		{"Blocked", issue.StatusBlocked, "blocked"},
		{"Review", issue.StatusReview, "review"},
		{"Done", issue.StatusDone, "done"},
		{"Abandoned", issue.StatusAbandoned, "abandoned"},
		{"Failed", issue.StatusFailed, "failed"},
		{"Stale", issue.StatusStale, "stale"},
		{"NeedsDecomposition", issue.StatusNeedsDecomposition, "needs-decomposition"},
		{"NeedsReplanning", issue.StatusNeedsReplanning, "needs-replanning"},
		{"Contested", issue.StatusContested, "contested"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.status.String())
		})
	}
}

func TestStatusMarshalText(t *testing.T) {
	tests := []struct {
		name   string
		status issue.Status
		want   string
	}{
		{"Draft", issue.StatusDraft, "draft"},
		{"InProgress", issue.StatusInProgress, "in-progress"},
		{"Done", issue.StatusDone, "done"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.status.MarshalText()
			require.NoError(t, err)
			assert.Equal(t, tt.want, string(data))
		})
	}
}

func TestStatusUnmarshalText(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    issue.Status
		wantErr bool
	}{
		{"Draft", "draft", issue.StatusDraft, false},
		{"InProgress", "IN-PROGRESS", issue.StatusInProgress, false},
		{"Done", "DoNe", issue.StatusDone, false},
		{"Invalid", "invalid", issue.Status{}, true},
		{"Empty", "", issue.Status{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got issue.Status
			err := got.UnmarshalText([]byte(tt.input))
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseStatus(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    issue.Status
		wantErr bool
	}{
		{"Draft", "draft", issue.StatusDraft, false},
		{"Ready", "READY", issue.StatusReady, false},
		{"InProgress", "in-progress", issue.StatusInProgress, false},
		{"Done", "done", issue.StatusDone, false},
		{"Invalid", "unknown", issue.Status{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := issue.ParseStatus(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestStatusRoundTrip(t *testing.T) {
	originals := []issue.Status{
		issue.StatusDraft, issue.StatusReady, issue.StatusDone,
		issue.StatusInProgress, issue.StatusBlocked,
	}
	for _, orig := range originals {
		data, err := orig.MarshalText()
		require.NoError(t, err)
		var restored issue.Status
		err = restored.UnmarshalText(data)
		require.NoError(t, err)
		assert.Equal(t, orig, restored)
	}
}

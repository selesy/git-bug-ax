package issue_test

import (
	"testing"

	"github.com/selesy/git-bug-ax/pkg/issue"
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
			if got := tt.status.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
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
			if err != nil {
				t.Errorf("MarshalText() error = %v", err)
				return
			}
			if got := string(data); got != tt.want {
				t.Errorf("MarshalText() = %q, want %q", got, tt.want)
			}
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
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalText() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("UnmarshalText() = %v, want %v", got, tt.want)
			}
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
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseStatus() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ParseStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStatusRoundTrip(t *testing.T) {
	originals := []issue.Status{
		issue.StatusDraft, issue.StatusReady, issue.StatusDone,
		issue.StatusInProgress, issue.StatusBlocked,
	}
	for _, orig := range originals {
		data, _ := orig.MarshalText()
		var restored issue.Status
		_ = restored.UnmarshalText(data)
		if restored != orig {
			t.Errorf("Status round-trip failed: %v != %v", restored, orig)
		}
	}
}

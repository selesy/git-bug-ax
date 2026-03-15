package issue_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/selesy/git-bug-ax/pkg/issue"
)

func TestPriorityConstants(t *testing.T) {
	tests := []struct {
		name     string
		priority issue.Priority
		expected string
	}{
		{"Highest", issue.PriorityHighest, "highest"},
		{"High", issue.PriorityHigh, "high"},
		{"Medium", issue.PriorityMedium, "medium"},
		{"Low", issue.PriorityLow, "low"},
		{"Lowest", issue.PriorityLowest, "lowest"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.priority.String())
		})
	}
}

func TestPriorityMarshalText(t *testing.T) {
	tests := []struct {
		name     string
		priority issue.Priority
		expected string
	}{
		{"Highest", issue.PriorityHighest, "highest"},
		{"High", issue.PriorityHigh, "high"},
		{"Medium", issue.PriorityMedium, "medium"},
		{"Low", issue.PriorityLow, "low"},
		{"Lowest", issue.PriorityLowest, "lowest"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.priority.MarshalText()
			require.NoError(t, err)
			assert.Equal(t, tt.expected, string(data))
		})
	}
}

func TestPriorityUnmarshalText(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    issue.Priority
		wantErr bool
	}{
		{"Exact", "highest", issue.PriorityHighest, false},
		{"Lowercase", "high", issue.PriorityHigh, false},
		{"Uppercase", "MEDIUM", issue.PriorityMedium, false},
		{"MixedCase", "LoW", issue.PriorityLow, false},
		{"Invalid", "invalid", issue.Priority{}, true},
		{"Empty", "", issue.Priority{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got issue.Priority
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

func TestParsePriority(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    issue.Priority
		wantErr bool
	}{
		{"Highest", "highest", issue.PriorityHighest, false},
		{"High", "HIGH", issue.PriorityHigh, false},
		{"Medium", "MeDiUm", issue.PriorityMedium, false},
		{"Low", "low", issue.PriorityLow, false},
		{"Lowest", "lowest", issue.PriorityLowest, false},
		{"Invalid", "critical", issue.Priority{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := issue.ParsePriority(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPriorityRoundTrip(t *testing.T) {
	originals := []issue.Priority{
		issue.PriorityHighest, issue.PriorityHigh,
		issue.PriorityMedium, issue.PriorityLow, issue.PriorityLowest,
	}
	for _, orig := range originals {
		data, err := orig.MarshalText()
		require.NoError(t, err)
		var restored issue.Priority
		err = restored.UnmarshalText(data)
		require.NoError(t, err)
		assert.Equal(t, orig, restored)
	}
}

package issue_test

import (
	"testing"

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
			if got := tt.priority.String(); got != tt.expected {
				t.Errorf("String() = %q, want %q", got, tt.expected)
			}
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
			if err != nil {
				t.Errorf("MarshalText() error = %v", err)
				return
			}
			if got := string(data); got != tt.expected {
				t.Errorf("MarshalText() = %q, want %q", got, tt.expected)
			}
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
			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePriority() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ParsePriority() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPriorityRoundTrip(t *testing.T) {
	originals := []issue.Priority{
		issue.PriorityHighest, issue.PriorityHigh,
		issue.PriorityMedium, issue.PriorityLow, issue.PriorityLowest,
	}
	for _, orig := range originals {
		data, _ := orig.MarshalText()
		var restored issue.Priority
		_ = restored.UnmarshalText(data)
		if restored != orig {
			t.Errorf("Priority round-trip failed: %v != %v", restored, orig)
		}
	}
}

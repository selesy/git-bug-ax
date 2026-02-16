package issue_test

import (
	"testing"

	"github.com/selesy/git-bug-ax/pkg/issue"
)

func TestResolutionConstants(t *testing.T) {
	tests := []struct {
		name       string
		resolution issue.Resolution
		want       string
	}{
		{"Fixed", issue.ResolutionFixed, "fixed"},
		{"WontFix", issue.ResolutionWontFix, "wont-fix"},
		{"Duplicate", issue.ResolutionDuplicate, "duplicate"},
		{"CannotRepro", issue.ResolutionCannotRepro, "cannot-repro"},
		{"NotNeeded", issue.ResolutionNotNeeded, "not-needed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.resolution.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestResolutionMarshalText(t *testing.T) {
	tests := []struct {
		name       string
		resolution issue.Resolution
		want       string
	}{
		{"Fixed", issue.ResolutionFixed, "fixed"},
		{"WontFix", issue.ResolutionWontFix, "wont-fix"},
		{"Duplicate", issue.ResolutionDuplicate, "duplicate"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.resolution.MarshalText()
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

func TestResolutionUnmarshalText(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    issue.Resolution
		wantErr bool
	}{
		{"Fixed", "fixed", issue.ResolutionFixed, false},
		{"WontFix", "WONT-FIX", issue.ResolutionWontFix, false},
		{"Duplicate", "DuPlIcAtE", issue.ResolutionDuplicate, false},
		{"Invalid", "invalid", issue.Resolution{}, true},
		{"Empty", "", issue.Resolution{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got issue.Resolution
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

func TestParseResolution(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    issue.Resolution
		wantErr bool
	}{
		{"Fixed", "fixed", issue.ResolutionFixed, false},
		{"WontFix", "WONT-FIX", issue.ResolutionWontFix, false},
		{"Duplicate", "duplicate", issue.ResolutionDuplicate, false},
		{"CannotRepro", "cannot-repro", issue.ResolutionCannotRepro, false},
		{"Invalid", "unknown", issue.Resolution{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := issue.ParseResolution(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseResolution() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ParseResolution() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResolutionRoundTrip(t *testing.T) {
	originals := []issue.Resolution{
		issue.ResolutionFixed, issue.ResolutionWontFix,
		issue.ResolutionDuplicate, issue.ResolutionCannotRepro,
	}
	for _, orig := range originals {
		data, _ := orig.MarshalText()
		var restored issue.Resolution
		_ = restored.UnmarshalText(data)
		if restored != orig {
			t.Errorf("Resolution round-trip failed: %v != %v", restored, orig)
		}
	}
}

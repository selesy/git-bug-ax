package issue_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
			assert.Equal(t, tt.want, tt.resolution.String())
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
			require.NoError(t, err)
			assert.Equal(t, tt.want, string(data))
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
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
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
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestResolutionRoundTrip(t *testing.T) {
	originals := []issue.Resolution{
		issue.ResolutionFixed, issue.ResolutionWontFix,
		issue.ResolutionDuplicate, issue.ResolutionCannotRepro,
	}
	for _, orig := range originals {
		data, err := orig.MarshalText()
		require.NoError(t, err)
		var restored issue.Resolution
		err = restored.UnmarshalText(data)
		require.NoError(t, err)
		assert.Equal(t, orig, restored)
	}
}

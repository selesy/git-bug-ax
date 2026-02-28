package issue_test

import (
	"testing"

	"github.com/selesy/git-bug-ax/pkg/issue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTypeConstants(t *testing.T) {
	tests := []struct {
		name string
		typ  issue.Type
		want string
	}{
		{"Epic", issue.TypeEpic, "epic"},
		{"Feature", issue.TypeFeature, "feature"},
		{"Task", issue.TypeTask, "task"},
		{"Bug", issue.TypeBug, "bug"},
		{"Spike", issue.TypeSpike, "spike"},
		{"TechDebt", issue.TypeTechDebt, "tech-debt"},
		{"Fix", issue.TypeFix, "fix"},
		{"Feat", issue.TypeFeat, "feat"},
		{"Build", issue.TypeBuild, "build"},
		{"Chore", issue.TypeChore, "chore"},
		{"CI", issue.TypeCI, "ci"},
		{"Docs", issue.TypeDocs, "docs"},
		{"Style", issue.TypeStyle, "style"},
		{"Refactor", issue.TypeRefactor, "refactor"},
		{"Test", issue.TypeTest, "test"},
		{"Perf", issue.TypePerf, "perf"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.typ.String())
		})
	}
}

func TestTypeMarshalText(t *testing.T) {
	tests := []struct {
		name string
		typ  issue.Type
		want string
	}{
		{"Epic", issue.TypeEpic, "epic"},
		{"Bug", issue.TypeBug, "bug"},
		{"Task", issue.TypeTask, "task"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.typ.MarshalText()
			require.NoError(t, err)
			assert.Equal(t, tt.want, string(data))
		})
	}
}

func TestTypeUnmarshalText(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    issue.Type
		wantErr bool
	}{
		{"Epic", "epic", issue.TypeEpic, false},
		{"Bug", "BUG", issue.TypeBug, false},
		{"Task", "TaSk", issue.TypeTask, false},
		{"TechDebt", "tech-debt", issue.TypeTechDebt, false},
		{"Invalid", "invalid", issue.Type{}, true},
		{"Empty", "", issue.Type{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got issue.Type
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

func TestParseType(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    issue.Type
		wantErr bool
	}{
		{"Epic", "epic", issue.TypeEpic, false},
		{"Feature", "FEATURE", issue.TypeFeature, false},
		{"Bug", "bug", issue.TypeBug, false},
		{"TechDebt", "tech-debt", issue.TypeTechDebt, false},
		{"Invalid", "unknown", issue.Type{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := issue.ParseType(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTypeRoundTrip(t *testing.T) {
	originals := []issue.Type{
		issue.TypeEpic, issue.TypeBug, issue.TypeTask,
		issue.TypeFeature, issue.TypeSpike,
	}
	for _, orig := range originals {
		data, err := orig.MarshalText()
		require.NoError(t, err)
		var restored issue.Type
		err = restored.UnmarshalText(data)
		require.NoError(t, err)
		assert.Equal(t, orig, restored)
	}
}

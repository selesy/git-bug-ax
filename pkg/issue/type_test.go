package issue_test

import (
	"testing"

	"github.com/selesy/git-bug-ax/pkg/issue"
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
			if got := tt.typ.String(); got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
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
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ParseType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTypeRoundTrip(t *testing.T) {
	originals := []issue.Type{
		issue.TypeEpic, issue.TypeBug, issue.TypeTask,
		issue.TypeFeature, issue.TypeSpike,
	}
	for _, orig := range originals {
		data, _ := orig.MarshalText()
		var restored issue.Type
		_ = restored.UnmarshalText(data)
		if restored != orig {
			t.Errorf("Type round-trip failed: %v != %v", restored, orig)
		}
	}
}

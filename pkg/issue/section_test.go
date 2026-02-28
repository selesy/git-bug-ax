package issue

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSectionIsValid(t *testing.T) {
	tests := []struct {
		section Section
		valid   bool
	}{
		{SectionOverview, true},
		{SectionScope, true},
		{SectionFilesAffected, true},
		{SectionEnvironment, true},
		{SectionImplementation, true},
		{SectionAcceptanceCriteria, true},
		{SectionVerification, true},
		{Section("custom_section"), false},
		{Section(""), false},
	}

	for _, tt := range tests {
		t.Run(string(tt.section), func(t *testing.T) {
			assert.Equal(t, tt.valid, tt.section.IsValid())
		})
	}
}

func TestParseSection(t *testing.T) {
	tests := []struct {
		input     string
		expected  Section
		shouldErr bool
	}{
		// Canonical names
		{"scope", SectionScope, false},
		{"SCOPE", SectionScope, false},
		{"Scope", SectionScope, false},
		{"overview", SectionOverview, false},
		{"files_affected", SectionFilesAffected, false},
		{"environment", SectionEnvironment, false},
		{"implementation", SectionImplementation, false},
		{"acceptance_criteria", SectionAcceptanceCriteria, false},
		{"verification", SectionVerification, false},

		// Aliases with variations
		{"Implementation Notes", SectionImplementation, false},
		{"implementation notes", SectionImplementation, false},
		{"IMPLEMENTATION NOTES", SectionImplementation, false},
		{"Impl-Notes", SectionImplementation, false},
		{"impl", SectionImplementation, false},

		{"Files Affected", SectionFilesAffected, false},
		{"files affected", SectionFilesAffected, false},
		{"files", SectionFilesAffected, false},
		{"FILES-AFFECTED", SectionFilesAffected, false},

		{"Acceptance Criteria", SectionAcceptanceCriteria, false},
		{"acceptance criteria", SectionAcceptanceCriteria, false},
		{"acceptance", SectionAcceptanceCriteria, false},
		{"criteria", SectionAcceptanceCriteria, false},

		// Custom sections
		{"custom_section", Section("custom_section"), false},
		{"Custom Section", Section("custom_section"), false},
		{"custom-research-log", Section("custom_research_log"), false},

		// Error cases
		{"", Section(""), true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := ParseSection(tt.input)
			if tt.shouldErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSectionHeading(t *testing.T) {
	tests := []struct {
		section Section
		heading string
	}{
		{SectionOverview, ""},
		{SectionScope, "Scope"},
		{SectionFilesAffected, "Files Affected"},
		{SectionEnvironment, "Environment"},
		{SectionImplementation, "Implementation Notes"},
		{SectionAcceptanceCriteria, "Acceptance Criteria"},
		{SectionVerification, "Verification"},
		{Section("custom_research_log"), "Custom Research Log"},
		{Section("my_section"), "My Section"},
	}

	for _, tt := range tests {
		t.Run(string(tt.section), func(t *testing.T) {
			assert.Equal(t, tt.heading, tt.section.Heading())
		})
	}
}

func TestSectionMarshalUnmarshal(t *testing.T) {
	tests := []Section{
		SectionScope,
		SectionImplementation,
		SectionAcceptanceCriteria,
	}

	for _, original := range tests {
		t.Run(string(original), func(t *testing.T) {
			marshaled, err := original.MarshalText()
			require.NoError(t, err)

			var unmarshaled Section
			err = unmarshaled.UnmarshalText(marshaled)
			require.NoError(t, err)

			assert.Equal(t, original, unmarshaled)
		})
	}
}

func TestSectionString(t *testing.T) {
	s := SectionScope
	assert.Equal(t, "scope", string(s))
}

func TestSections(t *testing.T) {
	sections := Sections()
	assert.Len(t, sections, 7)

	expected := []Section{
		SectionOverview,
		SectionScope,
		SectionFilesAffected,
		SectionEnvironment,
		SectionImplementation,
		SectionAcceptanceCriteria,
		SectionVerification,
	}

	assert.Equal(t, expected, sections)
}

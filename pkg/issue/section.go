package issue

import (
	"fmt"

	"github.com/selesy/git-bug-agent/internal/codec"
)

var _ codec.TextCodec = (*Section)(nil)

// Section represents a named section within an issue description.
// Sections are ordered and can be either canonical (predefined) or custom.
type Section string

const (
	// Canonical sections as defined in README.md
	SectionOverview           Section = "overview"
	SectionScope              Section = "scope"
	SectionFilesAffected      Section = "files_affected"
	SectionEnvironment        Section = "environment"
	SectionImplementation     Section = "implementation"
	SectionAcceptanceCriteria Section = "acceptance_criteria"
	SectionVerification       Section = "verification"
)

// IsValid returns true if the section is a known canonical section.
func (s Section) IsValid() bool {
	switch s {
	case SectionOverview, SectionScope, SectionFilesAffected, SectionEnvironment,
		SectionImplementation, SectionAcceptanceCriteria, SectionVerification:
		return true
	}
	return false
}

// String implements fmt.Stringer, returning the canonical section name.
func (s Section) String() string {
	return string(s)
}

// MarshalText implements encoding.TextMarshaler.
func (s Section) MarshalText() ([]byte, error) {
	return []byte(s), nil
}

// UnmarshalText implements encoding.TextUnmarshaler, parsing case-insensitively.
func (s *Section) UnmarshalText(text []byte) error {
	parsed, err := ParseSection(string(text))
	if err != nil {
		return err
	}
	*s = parsed
	return nil
}

// ParseSection parses a section name, recognizing aliases case-insensitively.
// Returns the canonical Section name or an error if unrecognized.
// Examples:
//   - "scope" → SectionScope
//   - "SCOPE" → SectionScope
//   - "files affected" → SectionFilesAffected
//   - "implementation notes" → SectionImplementation
//   - "custom-section" → Section("custom_section") (custom sections allowed)
func ParseSection(input string) (Section, error) {
	if input == "" {
		return "", fmt.Errorf("section name cannot be empty")
	}

	// Normalize the input: lowercase, replace separators with underscores
	normalized := normalizeSection(input)

	// Map common aliases to canonical section names
	aliases := map[string]Section{
		// Exact canonical names
		"overview":            SectionOverview,
		"scope":               SectionScope,
		"files_affected":      SectionFilesAffected,
		"environment":         SectionEnvironment,
		"implementation":      SectionImplementation,
		"acceptance_criteria": SectionAcceptanceCriteria,
		"verification":        SectionVerification,

		// Common aliases
		"implementation_notes": SectionImplementation,
		"impl_notes":           SectionImplementation,
		"impl":                 SectionImplementation,
		"files":                SectionFilesAffected,
		"acceptance":           SectionAcceptanceCriteria,
		"criteria":             SectionAcceptanceCriteria,
	}

	if canonical, ok := aliases[normalized]; ok {
		return canonical, nil
	}

	// Allow custom sections (anything not recognized as an alias)
	// Custom sections are returned as-is (normalized)
	return Section(normalized), nil
}

// Heading returns the display heading text for the section.
// For canonical sections, returns the standard heading.
// For custom sections, returns title-cased version of the name.
func (s Section) Heading() string {
	switch s {
	case SectionOverview:
		return "" // Overview has no heading
	case SectionScope:
		return "Scope"
	case SectionFilesAffected:
		return "Files Affected"
	case SectionEnvironment:
		return "Environment"
	case SectionImplementation:
		return "Implementation Notes"
	case SectionAcceptanceCriteria:
		return "Acceptance Criteria"
	case SectionVerification:
		return "Verification"
	default:
		// Custom section: denormalize to title case
		return denormalizeSection(string(s))
	}
}

// Sections returns all canonical sections in order.
func Sections() []Section {
	return []Section{
		SectionOverview,
		SectionScope,
		SectionFilesAffected,
		SectionEnvironment,
		SectionImplementation,
		SectionAcceptanceCriteria,
		SectionVerification,
	}
}

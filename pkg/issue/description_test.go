package issue

import (
	"strings"
	"testing"
)

func TestDescriptionParsing(t *testing.T) {
	markdown := `This is an overview paragraph describing the work.

## Scope

- Implement token validation
- Add middleware

## Files Affected

- pkg/api/handler.go
- pkg/validate/rules.go

## Implementation Notes

- Use existing validator package
- Add validation rules

## Acceptance Criteria

- [ ] All inputs validated
- [ ] Tests pass

## Verification

- go test ./pkg/api

## Environment

- Requires GO_API_KEY`

	desc := &Description{}
	err := desc.UnmarshalText([]byte(markdown))
	if err != nil {
		t.Fatalf("Failed to parse markdown: %v", err)
	}

	tests := []struct {
		name     string
		getter   func() []string
		expected int
	}{
		{"Overview", desc.Overview, 1},
		{"Scope", desc.Scope, 2},
		{"FilesAffected", desc.FilesAffected, 2},
		{"Implementation", desc.Implementation, 2},
		{"AcceptanceCriteria", desc.AcceptanceCriteria, 2},
		{"Verification", desc.Verification, 1},
		{"Environment", desc.Environment, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.getter()
			if len(result) == 0 {
				t.Errorf("Expected %d items, got 0", tt.expected)
			}
		})
	}
}

func TestDescriptionMarshal(t *testing.T) {
	markdown := `## Scope

- Test scope

## Implementation Notes

- Test impl`

	desc := &Description{}
	_ = desc.UnmarshalText([]byte(markdown))

	// Modify a section
	desc.SetScope([]string{"- Updated scope"})

	// Marshal back
	marshaled, err := desc.MarshalText()
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	if len(marshaled) == 0 {
		t.Error("Expected marshaled output, got empty")
	}

	// Parse again
	desc2 := &Description{}
	_ = desc2.UnmarshalText(marshaled)
	scope := desc2.Scope()
	if len(scope) == 0 {
		t.Error("Expected scope after re-parse")
	}
}

func TestDescriptionIgnoresLevel1HeadersDuringUnmarshal(t *testing.T) {
	// When unmarshaling markdown with a level-1 header, it should be silently ignored
	markdown := `# Title that should be ignored

This is an overview paragraph.

## Scope

- Task scope`

	desc := &Description{}
	_ = desc.UnmarshalText([]byte(markdown))

	// Verify the level-1 header text is NOT in any section
	overview := desc.Overview()
	if len(overview) > 0 {
		for _, line := range overview {
			if strings.Contains(line, "Title that should be ignored") {
				t.Error("Level-1 header content should not be in overview")
			}
		}
	}

	// Verify overview was properly captured (the real overview text)
	found := false
	for _, line := range overview {
		if strings.Contains(line, "overview paragraph") {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected overview paragraph to be extracted after level-1 header")
	}

	// Verify scope was captured
	scope := desc.Scope()
	if len(scope) == 0 {
		t.Error("Expected scope to be extracted")
	}
}

func TestDescriptionNeverOutputsLevel1Headers(t *testing.T) {
	// When marshaling, the Description should never output level-1 headers
	desc := &Description{sections: make(map[Section][]string)}
	desc.SetOverview([]string{"This is an overview."})
	desc.SetScope([]string{"- Scope item"})
	desc.SetImplementation([]string{"- Implementation item"})

	marshaled, _ := desc.MarshalText()
	output := string(marshaled)

	// Check that output has no single-# headers (level-1)
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Look for single # followed by space (level-1), but allow ## (level-2)
		if strings.HasPrefix(trimmed, "# ") {
			t.Errorf("Marshaled output should not contain level-1 headers, found: %s", line)
		}
	}

	// Verify it does contain level-2 headers
	if !strings.Contains(output, "## Scope") {
		t.Error("Marshaled output should contain level-2 headers like ## Scope")
	}
	if !strings.Contains(output, "## Implementation") {
		t.Error("Marshaled output should contain level-2 headers like ## Implementation")
	}
}

func TestDescriptionLevel1HeaderWithoutOverview(t *testing.T) {
	// Test that level-1 header is stripped even when there's no overview text
	markdown := `# Title to ignore

## Scope

- Scope content`

	desc := &Description{}
	_ = desc.UnmarshalText([]byte(markdown))

	// Overview should be empty or not contain the title
	overview := desc.Overview()
	for _, line := range overview {
		if strings.Contains(line, "Title to ignore") {
			t.Error("Level-1 header should be stripped")
		}
	}

	// Scope should be captured
	scope := desc.Scope()
	if len(scope) == 0 {
		t.Error("Expected scope to be extracted")
	}
}

func TestDescriptionPreservesOverviewBeforeFirstLevel2Header(t *testing.T) {
	// Test that multiple paragraphs before first ## are all captured as overview
	markdown := `First paragraph of overview.

Second paragraph of overview with more context.

## Scope

- Scope item`

	desc := &Description{}
	_ = desc.UnmarshalText([]byte(markdown))

	overview := desc.Overview()
	if len(overview) < 2 {
		t.Errorf("Expected at least 2 overview paragraphs, got %d", len(overview))
	}

	// Both paragraphs should be in overview
	overviewText := strings.Join(overview, " ")
	if !strings.Contains(overviewText, "First paragraph") {
		t.Error("Expected first paragraph in overview")
	}
	if !strings.Contains(overviewText, "Second paragraph") {
		t.Error("Expected second paragraph in overview")
	}
}

func TestDescriptionCapturesUnknownSections(t *testing.T) {
	// Test that custom level-2 headers are captured and preserved
	markdown := `Overview text.

## Scope

- Scope item

## Custom Decision Log

- Decision 1
- Decision 2

## Scope

- Additional scope

## Known Gotchas

- Gotcha 1
- Gotcha 2`

	desc := &Description{}
	_ = desc.UnmarshalText([]byte(markdown))

	// Known section should exist
	scope := desc.Scope()
	if len(scope) == 0 {
		t.Error("Expected scope section to be captured")
	}

	// Custom sections should be accessible
	if _, ok := desc.sections["custom_decision_log"]; !ok {
		t.Error("Expected unknown section 'custom_decision_log' to be captured")
	}
	if _, ok := desc.sections["known_gotchas"]; !ok {
		t.Error("Expected unknown section 'known_gotchas' to be captured")
	}

	// Order should be preserved
	expectedOrder := []Section{Section("overview"), Section("scope"), Section("custom_decision_log"), Section("scope"), Section("known_gotchas")}
	if len(desc.order) != len(expectedOrder) {
		t.Errorf("Expected %d sections in order, got %d", len(expectedOrder), len(desc.order))
	}
	for i, expected := range expectedOrder {
		if i >= len(desc.order) {
			break
		}
		if desc.order[i] != expected {
			t.Errorf("Expected section %d to be %q, got %q", i, expected, desc.order[i])
		}
	}
}

func TestDescriptionRoundripsUnknownSections(t *testing.T) {
	// Test that unknown sections survive marshal/unmarshal roundtrip
	markdown := `Overview context.

## Scope

- Standard section

## Custom Notes

- Note 1
- Note 2

## Implementation Notes

- Impl item`

	desc := &Description{}
	_ = desc.UnmarshalText([]byte(markdown))

	// Marshal and parse again
	marshaled, _ := desc.MarshalText()
	desc2 := &Description{}
	_ = desc2.UnmarshalText(marshaled)

	// Check that custom section survived
	if _, ok := desc2.sections["custom_notes"]; !ok {
		t.Error("Expected unknown section to survive roundtrip")
	}

	// Check that order is preserved
	if len(desc2.order) != len(desc.order) {
		t.Errorf("Expected same order length after roundtrip: %v vs %v", desc.order, desc2.order)
	}
}

func TestDescriptionDenormalizesCustomSections(t *testing.T) {
	// Test that custom section names are properly denormalized in output
	desc := &Description{sections: make(map[Section][]string)}
	desc.sections[Section("custom_research_log")] = []string{"- Research item"}
	desc.order = []Section{Section("custom_research_log")}

	marshaled, _ := desc.MarshalText()
	output := string(marshaled)

	// Should have denormalized heading like "## Custom Research Log"
	if !strings.Contains(output, "## Custom Research Log") {
		t.Errorf("Expected denormalized heading, got: %s", output)
	}
}

func TestDescriptionAliasesCaseInsensitive(t *testing.T) {
	// Test that aliases are recognized case-insensitively and normalized to canonical
	tests := []struct {
		name     string
		markdown string
		expected string // expected canonical section name in output
	}{
		{
			"Implementation Notes (title case)",
			`## Implementation Notes
- Item`,
			"## Implementation Notes",
		},
		{
			"implementation notes (lowercase)",
			`## implementation notes
- Item`,
			"## Implementation Notes",
		},
		{
			"IMPLEMENTATION NOTES (uppercase)",
			`## IMPLEMENTATION NOTES
- Item`,
			"## Implementation Notes",
		},
		{
			"Impl-Notes (hyphenated)",
			`## Impl-Notes
- Item`,
			"## Implementation Notes",
		},
		{
			"Files Affected",
			`## Files Affected
- file.go`,
			"## Files Affected",
		},
		{
			"FILES AFFECTED (uppercase)",
			`## FILES AFFECTED
- file.go`,
			"## Files Affected",
		},
		{
			"Acceptance Criteria",
			`## Acceptance Criteria
- [ ] Done`,
			"## Acceptance Criteria",
		},
		{
			"ACCEPTANCE-CRITERIA (hyphenated)",
			`## ACCEPTANCE-CRITERIA
- [ ] Done`,
			"## Acceptance Criteria",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			desc := &Description{}
			_ = desc.UnmarshalText([]byte(tt.markdown))

			marshaled, _ := desc.MarshalText()
			output := string(marshaled)

			if !strings.Contains(output, tt.expected) {
				t.Errorf("Expected canonical heading %q in output, got: %s", tt.expected, output)
			}
		})
	}
}

func TestDescriptionEmptySection(t *testing.T) {
	// Test that empty sections are handled properly
	markdown := `## Scope

## Files Affected

- file.go`

	desc := &Description{}
	_ = desc.UnmarshalText([]byte(markdown))

	// Empty scope should not be in sections
	scope := desc.Scope()
	if len(scope) > 0 {
		t.Error("Expected empty scope section to be omitted")
	}

	// Files affected should still be captured
	files := desc.FilesAffected()
	if len(files) == 0 {
		t.Error("Expected files affected section to be captured")
	}
}

func TestDescriptionHeadingWithoutContent(t *testing.T) {
	// Test heading followed immediately by another heading (no content between)
	markdown := `## Scope
## Files Affected
- file.go`

	desc := &Description{}
	_ = desc.UnmarshalText([]byte(markdown))

	scope := desc.Scope()
	if len(scope) > 0 {
		t.Error("Expected empty scope between headings")
	}

	files := desc.FilesAffected()
	if len(files) == 0 {
		t.Error("Expected files affected to be captured")
	}
}

func TestDescriptionOnlyOverview(t *testing.T) {
	// Test markdown with only overview (no level-2 sections)
	markdown := `This is just an overview paragraph with no sections.`

	desc := &Description{}
	_ = desc.UnmarshalText([]byte(markdown))

	// Should be stored as raw
	raw := desc.Raw()
	if !strings.Contains(raw, "overview paragraph") {
		t.Error("Expected overview to be stored as raw")
	}
}

func TestDescriptionMultipleParagraphsInSection(t *testing.T) {
	// Test section with multiple paragraphs
	markdown := `## Scope

First paragraph of scope.

Second paragraph of scope.

Third paragraph of scope.`

	desc := &Description{}
	_ = desc.UnmarshalText([]byte(markdown))

	scope := desc.Scope()
	if len(scope) < 3 {
		t.Errorf("Expected 3 paragraphs in scope, got %d", len(scope))
	}

	// All paragraphs should be present
	scopeText := strings.Join(scope, " ")
	if !strings.Contains(scopeText, "First") || !strings.Contains(scopeText, "Second") || !strings.Contains(scopeText, "Third") {
		t.Error("Expected all three paragraphs in scope")
	}
}

func TestDescriptionSpecialCharactersInHeading(t *testing.T) {
	// Test that special characters in headings are normalized correctly
	markdown := `## Files @#$% Affected!

- file.go`

	desc := &Description{}
	_ = desc.UnmarshalText([]byte(markdown))

	// Should be normalized and stored
	marshaled, _ := desc.MarshalText()
	output := string(marshaled)

	// Custom section should be denormalized back
	if !strings.Contains(output, "Files") {
		t.Error("Expected denormalized custom section heading in output")
	}
}

func TestDescriptionPreservesSectionOrder(t *testing.T) {
	// Test that sections are output in the order they appear in input
	markdown := `## Verification

- Run tests

## Scope

- Define scope

## Implementation Notes

- Implement feature`

	desc := &Description{}
	_ = desc.UnmarshalText([]byte(markdown))

	marshaled, _ := desc.MarshalText()
	output := string(marshaled)

	// Find positions of headings
	verificationPos := strings.Index(output, "## Verification")
	scopePos := strings.Index(output, "## Scope")
	implPos := strings.Index(output, "## Implementation Notes")

	if verificationPos == -1 || scopePos == -1 || implPos == -1 {
		t.Error("Expected all headings in output")
		return
	}

	// Verify order is preserved (Verification before Scope before Implementation)
	if verificationPos >= scopePos || scopePos >= implPos {
		t.Error("Expected sections in original order: Verification, Scope, Implementation")
	}
}

func TestDescriptionContentBetweenHeadings(t *testing.T) {
	// Test the condition: currentSection != "" && len(currentContent) > 0
	// This tests that BOTH currentSection is set AND there's content to save
	markdown := `## Scope

Content under scope heading.

## Files Affected

- file.go`

	desc := &Description{}
	_ = desc.UnmarshalText([]byte(markdown))

	// Scope should have its content saved when we hit the next heading
	scope := desc.Scope()
	if len(scope) == 0 {
		t.Error("Expected scope content to be saved before next heading")
	}

	// Files should be captured in next section
	files := desc.FilesAffected()
	if len(files) == 0 {
		t.Error("Expected files affected to be captured")
	}
}

func TestDescriptionOverviewContentBeforeFirstHeading(t *testing.T) {
	// Test the condition: !foundFirstHeading && currentSection == ""
	// This tests that we capture overview content BEFORE any level-2 heading
	markdown := `This is overview paragraph 1.

This is overview paragraph 2.

## Scope

- Scope content`

	desc := &Description{}
	_ = desc.UnmarshalText([]byte(markdown))

	overview := desc.Overview()
	if len(overview) < 2 {
		t.Errorf("Expected at least 2 overview paragraphs, got %d", len(overview))
	}

	// Verify both paragraphs are present
	overviewText := strings.Join(overview, " ")
	if !strings.Contains(overviewText, "paragraph 1") || !strings.Contains(overviewText, "paragraph 2") {
		t.Error("Expected both overview paragraphs")
	}
}

func TestDescriptionContentAfterFirstHeading(t *testing.T) {
	// Test that content in first section is properly captured
	// and that after finding first heading, we don't accumulate in overview
	markdown := `## Scope

First item in scope.

Second item in scope.

## Files Affected

- file.go`

	desc := &Description{}
	_ = desc.UnmarshalText([]byte(markdown))

	scope := desc.Scope()
	if len(scope) < 2 {
		t.Errorf("Expected at least 2 scope items, got %d", len(scope))
	}
}

func TestDescriptionNormalizeSeparators(t *testing.T) {
	// Test normalizeSection with various separator characters
	// Tests rune == ' ' || r == '-' || r == '_' condition
	tests := []struct {
		input    string
		contains string // expected normalized substring
	}{
		{"files affected", "files_affected"},   // space to underscore
		{"files-affected", "files_affected"},   // hyphen to underscore
		{"files_affected", "files_affected"},   // underscore stays
		{"files  affected", "files_affected"},  // multiple spaces
		{"files--affected", "files_affected"},  // multiple hyphens
		{"files_-_affected", "files_affected"}, // mixed separators
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalizeSection(tt.input)
			if result != tt.contains {
				t.Errorf("normalizeSection(%q) = %q, expected %q", tt.input, result, tt.contains)
			}
		})
	}
}

func TestDescriptionNormalizeAlphanumeric(t *testing.T) {
	// Test that alphanumeric characters are preserved
	// Tests (r >= '0' && r <= '9') condition
	tests := []struct {
		input    string
		expected string
	}{
		{"section123", "section123"},
		{"123section", "123section"},
		{"sec123tion", "sec123tion"},
		{"section-1", "section_1"},
		{"1", "1"},
		{"123", "123"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalizeSection(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeSection(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestDescriptionEmptyInput(t *testing.T) {
	// Test with completely empty input
	// Tests len(s) > 0 condition
	desc := &Description{}
	err := desc.UnmarshalText([]byte(""))
	if err != nil {
		t.Errorf("UnmarshalText should not error on empty input: %v", err)
	}

	raw := desc.Raw()
	if raw != "" {
		t.Error("Expected empty output for empty input")
	}
}

func TestDescriptionSingleCharacterSection(t *testing.T) {
	// Test with minimal content
	// Tests boundary conditions for len > 0
	markdown := `## S

a`

	desc := &Description{}
	_ = desc.UnmarshalText([]byte(markdown))

	// Custom section "s" should be captured
	if _, ok := desc.sections[Section("s")]; !ok {
		t.Error("Expected single-character section to be captured")
	}
}

func TestDescriptionHeadingWithoutSavingContent(t *testing.T) {
	// Test: if currentSection != "" && len(currentContent) > 0
	// When we hit a new heading with NO currentContent, it should NOT save
	// This tests the && logic - both must be true to save
	markdown := `## Scope

## Files Affected

- file.go`

	desc := &Description{}
	_ = desc.UnmarshalText([]byte(markdown))

	// Scope should be empty (no content before next heading)
	scope := desc.Scope()
	if len(scope) > 0 {
		t.Error("Expected scope to be empty when no content follows heading")
	}

	// Files should have content
	files := desc.FilesAffected()
	if len(files) == 0 {
		t.Error("Expected files to have content")
	}
}

func TestDescriptionCurrentSectionEmptyString(t *testing.T) {
	// Test: if currentSection != "" condition
	// currentSection should be empty string initially
	// This verifies that without a section set, content doesn't get saved
	markdown := `Some overview text that appears before any heading.

More overview text.`

	desc := &Description{}
	_ = desc.UnmarshalText([]byte(markdown))

	// Should be stored as raw since no level-2 sections
	raw := desc.Raw()
	if !strings.Contains(raw, "overview text") {
		t.Error("Expected text without sections to be stored as raw")
	}

	// Should NOT be in any named section
	scope := desc.Scope()
	if len(scope) > 0 {
		t.Error("Expected scope to be empty when no ## Scope heading")
	}
}

func TestDescriptionFoundFirstHeadingTrueFalse(t *testing.T) {
	// Test: !foundFirstHeading && currentSection == "" condition
	// This tests that once we find first heading, we don't accumulate to overview anymore
	markdown := `Overview before any section.

## Scope

Content in scope.

## Files Affected

- file.go`

	desc := &Description{}
	_ = desc.UnmarshalText([]byte(markdown))

	overview := desc.Overview()
	if len(overview) == 0 {
		t.Error("Expected overview before first heading")
	}

	// The "Content in scope" should NOT be in overview
	overviewText := strings.Join(overview, " ")
	if strings.Contains(overviewText, "Content in scope") {
		t.Error("Expected content after first heading to NOT be in overview")
	}
}

func TestDescriptionRawSectionConditions(t *testing.T) {
	// Test: !hasStructuredSections && len(s) > 0 condition
	// Only store as raw if BOTH no structured sections AND input is not empty

	// Test 1: No sections, with content -> should be raw
	markdown1 := `Just some plain text without any sections.`
	desc1 := &Description{}
	_ = desc1.UnmarshalText([]byte(markdown1))
	if _, ok := desc1.sections[SectionRaw]; !ok {
		t.Error("Expected raw content when no sections")
	}

	// Test 2: No sections, empty input -> should NOT be raw
	desc2 := &Description{}
	_ = desc2.UnmarshalText([]byte(""))
	if _, ok := desc2.sections[SectionRaw]; ok {
		t.Error("Expected no raw section for empty input")
	}

	// Test 3: Has sections, no matter what -> should NOT be raw
	markdown3 := `## Scope

- item`
	desc3 := &Description{}
	_ = desc3.UnmarshalText([]byte(markdown3))
	if _, ok := desc3.sections[SectionRaw]; ok {
		t.Error("Expected no raw section when structured sections exist")
	}
}

func TestDescriptionReassembleRawCondition(t *testing.T) {
	// Test: if raw, ok := d.sections[SectionRaw]; ok && len(raw) > 0
	// Must be both in map AND have content to return as raw

	// Create a description with only raw content
	markdown := `Plain text without sections.`
	desc := &Description{}
	_ = desc.UnmarshalText([]byte(markdown))

	// Reassemble should return the raw text
	reassembled := desc.Raw()
	if reassembled != markdown {
		t.Errorf("Expected raw content to reassemble exactly, got: %q", reassembled)
	}

	// Test the condition by manually creating a section with empty content
	desc2 := &Description{sections: make(map[Section][]string)}
	desc2.sections[SectionRaw] = []string{} // Empty slice in raw
	reassembled2 := desc2.Raw()
	if reassembled2 != "" {
		// If empty raw is treated as truthy, this would fail
		t.Errorf("Empty raw should not be returned, got: %q", reassembled2)
	}
}

func TestDescriptionSectionContentCondition(t *testing.T) {
	// Test: if content, ok := d.sections[section]; ok && len(content) > 0
	// Must be in map AND have content to output

	markdown := `## Scope

- item 1

- item 2

## Files Affected`

	desc := &Description{}
	_ = desc.UnmarshalText([]byte(markdown))

	reassembled, _ := desc.MarshalText()
	output := string(reassembled)

	// Scope should be in output (has content)
	if !strings.Contains(output, "## Scope") {
		t.Error("Expected Scope section in output when it has content")
	}

	// Files Affected should NOT be in output (no content)
	if strings.Contains(output, "## Files Affected") {
		t.Error("Expected Files Affected to be omitted when empty")
	}
}

func TestDescriptionNormalizeRuneConditions(t *testing.T) {
	// Test the rune conditions in normalizeSection:
	// r == ' ' || r == '-' || r == '_' (separators)
	// (r >= 'a' && r <= 'z') (lowercase letters)
	// (r >= '0' && r <= '9') (digits)

	tests := []struct {
		input    string
		expected string
	}{
		// Separators should become underscores
		{"my section", "my_section"},
		{"my-section", "my_section"},
		{"my_section", "my_section"},

		// Uppercase should become lowercase
		{"MySection", "mysection"},
		{"SECTION", "section"},

		// Numbers should be preserved
		{"section123", "section123"},
		{"123section", "123section"},

		// Special chars should become underscores
		{"my@section", "my_section"},
		{"my#section", "my_section"},
		{"my$section", "my_section"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalizeSection(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeSection(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestDescriptionDenormalizeWordBoundary(t *testing.T) {
	// Test denormalizeSection with word boundaries
	// Tests len(word) > 0 condition
	tests := []struct {
		input    string
		expected string
	}{
		{"simple", "Simple"},
		{"multiple_words", "Multiple Words"},
		{"three_word_example", "Three Word Example"},
		{"a_b_c", "A B C"},
		{"single_", "Single "}, // trailing underscore becomes space
		{"_single", " Single"}, // leading underscore becomes space
		{"word_", "Word "},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := denormalizeSection(tt.input)
			if result != tt.expected {
				t.Errorf("denormalizeSection(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestDescriptionOverviewSectionMapping(t *testing.T) {
	// Test: if section == SectionOverview vs other sections
	// Overview should have no heading in output, others should

	desc := &Description{sections: make(map[Section][]string)}
	desc.sections[SectionOverview] = []string{"Overview text"}
	desc.sections[SectionScope] = []string{"Scope text"}
	desc.order = []Section{SectionOverview, SectionScope}

	output := string(desc.Raw())

	// Overview should NOT have a heading
	if strings.Contains(output, "## Overview") {
		t.Error("Expected Overview section to have no heading")
	}

	// Overview content should still be there
	if !strings.Contains(output, "Overview text") {
		t.Error("Expected overview text in output")
	}

	// Scope should have a heading
	if !strings.Contains(output, "## Scope") {
		t.Error("Expected Scope heading in output")
	}
}

func TestDescriptionKnownSectionMapping(t *testing.T) {
	// Test: if standardHeading, isKnown := sectionToHeading[section]
	// Known sections should use standard heading, unknown should be denormalized

	desc := &Description{sections: make(map[Section][]string)}
	desc.sections[SectionImplementation] = []string{"- impl"}
	desc.sections[Section("custom_decision")] = []string{"- decision"}
	desc.order = []Section{SectionImplementation, Section("custom_decision")}

	output := string(desc.Raw())

	// Known section should use standard heading
	if !strings.Contains(output, "## Implementation Notes") {
		t.Error("Expected standard heading for Implementation section")
	}

	// Unknown section should be denormalized
	if !strings.Contains(output, "## Custom Decision") {
		t.Error("Expected denormalized heading for custom section")
	}
}

func TestDescriptionHeadingLevels(t *testing.T) {
	// Test that level-1, level-2, and level-3 headers are handled differently
	// Level-1: should be ignored
	// Level-2: should start structured sections
	// Level-3: should reset current section

	markdown := `# Title Level 1

## Scope

- scope item

### Subsection Level 3

Some text

## Files Affected

- file.go`

	desc := &Description{}
	_ = desc.UnmarshalText([]byte(markdown))

	scope := desc.Scope()
	if len(scope) == 0 {
		t.Error("Expected Scope section to be captured")
	}

	files := desc.FilesAffected()
	if len(files) == 0 {
		t.Error("Expected Files Affected section to be captured")
	}

	// "Some text" should not be in Files since level-3 heading resets section
	filesText := strings.Join(files, " ")
	if strings.Contains(filesText, "Some text") {
		t.Error("Expected level-3 heading to reset current section")
	}
}

func TestDescriptionContentAccumulation(t *testing.T) {
	// Test that content with exactly one item and zero items are handled differently
	// len(content) > 0 vs len(content) == 0

	// Case 1: Heading with exactly one item
	markdown1 := `## Scope
- single item`

	desc1 := &Description{}
	_ = desc1.UnmarshalText([]byte(markdown1))
	scope1 := desc1.Scope()
	if len(scope1) == 0 {
		t.Error("Expected scope with 1 item to be saved")
	}

	// Case 2: Empty overview (0 items)
	markdown2 := `## Scope
- item`

	desc2 := &Description{}
	_ = desc2.UnmarshalText([]byte(markdown2))
	overview2 := desc2.Overview()
	if len(overview2) > 0 {
		t.Error("Expected empty overview to not be saved")
	}
}

func TestDescriptionSingleCharacterEdgeCases(t *testing.T) {
	// Test the boundary between zero and one bytes/characters

	// Markdown with 1 byte
	desc1 := &Description{}
	_ = desc1.UnmarshalText([]byte("a"))
	raw1 := desc1.Raw()
	if raw1 != "a" {
		t.Error("Expected single character to be stored as raw")
	}

	// Empty markdown (0 bytes)
	desc2 := &Description{}
	_ = desc2.UnmarshalText([]byte(""))
	raw2 := desc2.Raw()
	if raw2 != "" {
		t.Error("Expected empty input to result in empty output")
	}
}

func TestDescriptionNormalizeDoubleUnderscore(t *testing.T) {
	// Test the boundary condition in the "__" cleanup loop
	// The loop: for strings.Contains(normalized, "__")

	markdown := `## My__Section__Name

- content`

	desc := &Description{}
	_ = desc.UnmarshalText([]byte(markdown))

	// Should normalize double underscores to single
	reassembled := string(desc.Raw())
	// After normalization and denormalization, should be "My Section Name"
	if !strings.Contains(reassembled, "My Section Name") {
		t.Errorf("Expected normalized heading in output, got: %s", reassembled)
	}
}

func TestDescriptionMustReturnRawFirst(t *testing.T) {
	// Test that raw content is returned immediately without further processing
	// if raw, ok := d.sections[SectionRaw]; ok { if len(raw) > 0 { return raw[0] } }

	markdown := `Just plain text, no sections here.`
	desc := &Description{}
	_ = desc.UnmarshalText([]byte(markdown))

	output := desc.Raw()

	// Should return exactly the raw text, not process it further
	if output != markdown {
		t.Errorf("Expected exact raw content, got %q", output)
	}

	// Should not have any markdown heading syntax added
	if strings.Contains(output, "##") {
		t.Error("Expected raw content to be returned as-is without headers")
	}
}

func TestDescriptionEmptySectionNotInOutput(t *testing.T) {
	// Test the content output condition: if len(content) > 0
	// Empty sections should not appear in output at all

	markdown := `## Scope

- scope item

## Files Affected

## Environment

- env var`

	desc := &Description{}
	_ = desc.UnmarshalText([]byte(markdown))

	reassembled, _ := desc.MarshalText()
	output := string(reassembled)

	// Scope should be in output
	if !strings.Contains(output, "## Scope") {
		t.Error("Expected Scope in output")
	}

	// Files Affected should NOT be in output (empty)
	if strings.Contains(output, "## Files Affected") {
		t.Error("Expected empty Files Affected to be omitted from output")
	}

	// Environment should be in output
	if !strings.Contains(output, "## Environment") {
		t.Error("Expected Environment in output")
	}
}

func TestDescriptionRuneClassification(t *testing.T) {
	// Test the rune classification conditions in normalizeSection
	// Tests that uppercase letters are converted to lowercase
	// and that the condition (r >= 'a' && r <= 'z') is checked correctly

	tests := []struct {
		input         string
		expectedLower bool
	}{
		{"UPPERCASE", true}, // all should be lowercase
		{"lowercase", true}, // already lowercase
		{"MiXeD", true},     // mixed case should become lowercase
		{"A", true},         // single uppercase
		{"a", true},         // single lowercase
		{"123", true},       // numbers should be preserved as-is
	}

	for _, tt := range tests {
		result := normalizeSection(tt.input)
		// Check that result doesn't contain uppercase
		for _, r := range result {
			if r >= 'A' && r <= 'Z' {
				t.Errorf("normalizeSection(%q) = %q contains uppercase", tt.input, result)
			}
		}
	}
}

func TestDescriptionSeparatorCharacters(t *testing.T) {
	// Test that specific separator characters (space, dash, underscore) are handled
	// Tests: r == ' ' || r == '-' || r == '_'

	tests := []struct {
		input    string
		expected string
		desc     string
	}{
		{"test space", "test_space", "space becomes underscore"},
		{"test-dash", "test_dash", "dash becomes underscore"},
		{"test_underscore", "test_underscore", "underscore stays"},
		{"test space-dash_mix", "test_space_dash_mix", "all separators normalized"},
		{"space-hyphen_under", "space_hyphen_under", "all three separator types"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			result := normalizeSection(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeSection(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestDescriptionEmptyOverviewNotSaved(t *testing.T) {
	// Boundary test: len(overviewContent) > 0
	// If the check was removed or changed to >= 0, this would fail
	// Empty overview should NOT be saved at all

	markdown := `## Scope
- scope item`

	desc := &Description{}
	_ = desc.UnmarshalText([]byte(markdown))

	overview := desc.Overview()
	if len(overview) > 0 {
		t.Error("Empty overview should not be in sections map")
	}

	// Verify by checking the sections map directly
	if _, exists := desc.sections[SectionOverview]; exists {
		t.Error("SectionOverview should not exist when no overview content")
	}
}

func TestDescriptionEmptyCurrentContentNotSaved(t *testing.T) {
	// Boundary test: len(currentContent) > 0
	// If the check was removed or changed to >= 0, empty sections would be saved
	// Consecutive headings with no content between should not save empty section

	markdown := `## Scope
## Files Affected
- file.go`

	desc := &Description{}
	_ = desc.UnmarshalText([]byte(markdown))

	scope := desc.Scope()
	if len(scope) > 0 {
		t.Error("Empty Scope section should not be saved")
	}

	// But Files Affected should exist
	files := desc.FilesAffected()
	if len(files) == 0 {
		t.Error("Files Affected should be saved")
	}
}

func TestDescriptionEmptyInputNotStored(t *testing.T) {
	// Boundary test: len(s) > 0
	// If the check was removed or changed to >= 0, empty input would create raw section
	// Empty input should result in no sections at all

	desc := &Description{}
	_ = desc.UnmarshalText([]byte(""))

	if _, exists := desc.sections[SectionRaw]; exists {
		t.Error("Empty input should not create raw section")
	}

	if len(desc.sections) > 0 {
		t.Error("Empty input should result in no sections")
	}
}

func TestDescriptionOneItemSection(t *testing.T) {
	// Boundary test: len(content) > 0
	// Verify that single item (len == 1) IS saved

	markdown := `## Scope
- single item`

	desc := &Description{}
	_ = desc.UnmarshalText([]byte(markdown))

	scope := desc.Scope()
	if len(scope) != 1 {
		t.Errorf("Expected 1 item in scope, got %d", len(scope))
	}
}

func TestDescriptionEmptySectionNotOutput(t *testing.T) {
	// Boundary test: if len(content) > 0 in reassembleMarkdown
	// If the check was removed or changed to >= 0, empty sections would appear in output

	desc := &Description{sections: make(map[Section][]string)}
	desc.sections[SectionScope] = []string{}                  // Empty scope
	desc.sections[SectionFilesAffected] = []string{"file.go"} // Non-empty files
	desc.order = []Section{SectionScope, SectionFilesAffected}

	output := string(desc.Raw())

	// Empty scope should not appear in output at all
	if strings.Contains(output, "## Scope") {
		t.Error("Empty Scope section should not appear in output")
	}

	// Non-empty files should appear
	if !strings.Contains(output, "## Files Affected") {
		t.Error("Non-empty Files Affected should appear in output")
	}
}

func TestDescriptionOneSectionOutput(t *testing.T) {
	// Boundary test: verify that single-item section (len == 1) IS output

	desc := &Description{sections: make(map[Section][]string)}
	desc.sections[SectionScope] = []string{"- single"}
	desc.order = []Section{SectionScope}

	output := string(desc.Raw())

	if !strings.Contains(output, "## Scope") {
		t.Error("Section with 1 item should be in output")
	}
}

func TestDescriptionRawWithOneItem(t *testing.T) {
	// Boundary test: if len(raw) > 0 in reassembleMarkdown
	// If check was removed, would still return early from reassembleMarkdown

	desc := &Description{sections: make(map[Section][]string)}
	desc.sections[SectionRaw] = []string{"raw content"} // 1 item

	output := desc.Raw()
	if output != "raw content" {
		t.Errorf("Raw with 1 item should be returned, got: %q", output)
	}
}

func TestDescriptionRawWithZeroItems(t *testing.T) {
	// Boundary test: if len(raw) > 0
	// If removed or changed to >= 0, would return empty string instead of processing sections

	desc := &Description{sections: make(map[Section][]string)}
	desc.sections[SectionRaw] = []string{} // 0 items
	desc.sections[SectionScope] = []string{"scope"}
	desc.order = []Section{SectionScope}

	output := desc.Raw()

	// Should NOT short-circuit on empty raw, should process sections
	if !strings.Contains(output, "## Scope") {
		t.Error("Should process sections when raw has 0 items")
	}
}

func TestDescriptionRuneBoundaries(t *testing.T) {
	// Test rune boundaries: (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
	// Boundary mutations would change >= to > or <= to <
	// Test edge runes that should pass the boundary check
	tests := []struct {
		input    string
		expected string
		desc     string
	}{
		{"a", "a", "lowercase a at start of range"},
		{"z", "z", "lowercase z at end of range"},
		{"0", "0", "digit 0 at start of range"},
		{"9", "9", "digit 9 at end of range"},
		{"abc0", "abc0", "lowercase and digits together"},
		{"abc@def", "abc_def", "@ converted to underscore"},
		{"my[section", "my_section", "[ converted to underscore"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			result := normalizeSection(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeSection(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestDescriptionCurrentContentBoundary(t *testing.T) {
	// Explicit test for len(currentContent) > 0 boundary at line 69
	// Boundary mutation would change > to >= or <=
	// Test that EXACTLY 1 item gets saved (should pass > 0 but not > 1)

	markdown := `## Scope
- exactly one item`

	desc := &Description{}
	_ = desc.UnmarshalText([]byte(markdown))

	scope := desc.Scope()
	if len(scope) != 1 {
		t.Errorf("Expected exactly 1 item in scope, got %d", len(scope))
	}

	// And verify 0 items would not be saved
	markdown2 := `## Scope
## Files Affected
- file`

	desc2 := &Description{}
	_ = desc2.UnmarshalText([]byte(markdown2))

	scope2 := desc2.Scope()
	if len(scope2) > 0 {
		t.Error("Expected 0 items in scope (empty section)")
	}
}

func TestIsSeparatorRune(t *testing.T) {
	// Test isSeparatorRune helper function
	// Tests: r == ' ' || r == '-' || r == '_'

	tests := []struct {
		input    rune
		expected bool
	}{
		{' ', true},   // space
		{'-', true},   // dash
		{'_', true},   // underscore
		{'a', false},  // letter
		{'0', false},  // digit
		{'@', false},  // special char
		{'[', false},  // bracket
		{'\n', false}, // newline
	}

	for _, tt := range tests {
		result := isSeparatorRune(tt.input)
		if result != tt.expected {
			t.Errorf("isSeparatorRune(%q) = %v, expected %v", tt.input, result, tt.expected)
		}
	}
}

func TestIsAlphanumericRune(t *testing.T) {
	// Test isAlphanumericRune helper function
	// Tests: (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_'

	tests := []struct {
		input    rune
		expected bool
	}{
		{'a', true},  // lowercase start
		{'z', true},  // lowercase end
		{'m', true},  // lowercase middle
		{'0', true},  // digit start
		{'9', true},  // digit end
		{'5', true},  // digit middle
		{'_', true},  // underscore
		{' ', false}, // space
		{'-', false}, // dash
		{'A', false}, // uppercase
		{'@', false}, // special char
		{'[', false}, // bracket
	}

	for _, tt := range tests {
		result := isAlphanumericRune(tt.input)
		if result != tt.expected {
			t.Errorf("isAlphanumericRune(%q) = %v, expected %v", tt.input, result, tt.expected)
		}
	}
}

func TestDescriptionAllSeparators(t *testing.T) {
	// Test that ALL three separator conditions in normalizeSection work
	// r == ' ' || r == '-' || r == '_'
	// If one condition was removed, this test would fail

	tests := []struct {
		input    string
		expected string
	}{
		{"with space", "with_space"},                           // tests r == ' '
		{"with-dash", "with_dash"},                             // tests r == '-'
		{"with_underscore", "with_underscore"},                 // tests r == '_'
		{"all three-types_and-mix", "all_three_types_and_mix"}, // all three
	}

	for _, tt := range tests {
		result := normalizeSection(tt.input)
		if result != tt.expected {
			t.Errorf("normalizeSection(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}

func TestDescriptionSeparatorVsSpecialChar(t *testing.T) {
	// Test that separators are handled differently from other special chars
	// Separators (space, dash, underscore) become underscores
	// Other chars become underscores too, but verify the distinction

	// Separators should result in underscore
	sep := normalizeSection("a-b c_d")
	if sep != "a_b_c_d" {
		t.Errorf("Expected 'a_b_c_d' for separators, got %q", sep)
	}

	// Non-separators should also become underscores
	special := normalizeSection("a@b#c$d")
	if special != "a_b_c_d" {
		t.Errorf("Expected 'a_b_c_d' for special chars, got %q", special)
	}
}

func TestDescriptionOverviewVsNormalSections(t *testing.T) {
	// Test that Overview section gets special treatment (no heading)
	// vs all other sections which get "## Heading" format

	desc := &Description{sections: make(map[Section][]string)}
	desc.sections[SectionOverview] = []string{"Overview text"}
	desc.sections[SectionScope] = []string{"Scope text"}
	desc.order = []Section{SectionOverview, SectionScope}

	output := string(desc.Raw())

	// Verify Overview has no heading
	if strings.Contains(output, "## Overview") {
		t.Error("Expected Overview to have no heading")
	}

	// Verify overview content is still in output
	if !strings.Contains(output, "Overview text") {
		t.Error("Expected overview text to be in output")
	}

	// Verify Scope has a heading
	if !strings.Contains(output, "## Scope") {
		t.Error("Expected Scope to have ## heading")
	}
}

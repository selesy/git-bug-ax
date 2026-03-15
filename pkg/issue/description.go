package issue

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"

	"github.com/selesy/git-bug-ax/internal/codec"
)

var _ codec.TextCodec = (*Description)(nil)

// Description represents a Markdown document with structured sections.
// Sections are defined as level-2 headers followed by content.
// Both known sections (scope, files-affected, etc.) and unknown custom sections
// are preserved, with order maintained from the source document.
type Description struct {
	sections map[Section][]string // normalized section name → content lines
	order    []Section            // section order as they appear in source (preserves user intent)
}

// SectionRaw is an internal-only pseudo-section for unstructured content
const SectionRaw Section = "_raw"

// Raw returns the raw Markdown text by reassembling sections.
func (d *Description) Raw() string {
	return d.reassembleMarkdown()
}

// MarshalText converts the Description back to Markdown.
func (d *Description) MarshalText() ([]byte, error) {
	return []byte(d.reassembleMarkdown()), nil
}

// UnmarshalText parses Markdown text and extracts sections.
// If the text doesn't contain structured sections, it's stored as raw content.
// Both known and unknown sections are captured; order is preserved.
func (d *Description) UnmarshalText(s []byte) error {
	d.sections = make(map[Section][]string)
	d.order = []Section{}

	parser := goldmark.DefaultParser()
	root := parser.Parse(text.NewReader(s))

	// Walk the AST and extract sections
	var currentSection Section
	var currentContent []string
	var overviewContent []string
	hasStructuredSections := false
	foundFirstHeading := false

	_ = ast.Walk(root, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		switch n := n.(type) {
		case *ast.Heading:
			// Silently skip level-1 headers (title is stored separately)
			if n.Level == 1 {
				return ast.WalkContinue, nil
			}

			// If we encounter a new heading and have accumulated content, save it
			if currentSection != "" {
				if len(currentContent) >= 1 {
					d.sections[currentSection] = currentContent
					currentContent = []string{}
				}
			}

			// Check if this is a level-2 heading (first one marks start of structured sections)
			if n.Level == 2 {
				if !foundFirstHeading {
					// This is the first level-2 heading; save any overview content
					if len(overviewContent) > 0 {
						d.sections[SectionOverview] = overviewContent
						d.order = append(d.order, SectionOverview)
					}
					foundFirstHeading = true
				}
				headingText := extractHeadingText(n, s)
				normalizedStr := normalizeSection(headingText)
				currentSection = Section(normalizedStr)
				d.order = append(d.order, currentSection) // Track order
				hasStructuredSections = true
			} else {
				// Non-level-2 headings (that aren't level-1) reset the current section
				currentSection = ""
			}

		case *ast.Paragraph, *ast.List, *ast.CodeBlock, *ast.Blockquote:
			nodeText := extractNodeText(n, s)
			if nodeText != "" {
				// Accumulate in appropriate section
				if !foundFirstHeading {
					if currentSection == "" {
						// Before any level-2 heading: this is overview
						overviewContent = append(overviewContent, nodeText)
					}
				} else if currentSection != "" {
					// Inside a level-2 section (after first heading found)
					currentContent = append(currentContent, nodeText)
				}
			}
		}
		return ast.WalkContinue, nil
	})

	// Save the last accumulated section
	if currentSection != "" {
		if len(currentContent) >= 1 {
			d.sections[currentSection] = currentContent
		}
	}

	// If no structured sections found, treat the entire input as raw content
	if !hasStructuredSections {
		if len(s) > 0 {
			d.sections[SectionRaw] = []string{string(s)}
			d.order = append(d.order, SectionRaw)
		}
	}

	return nil
}

// extractHeadingText extracts text from a heading node.
func extractHeadingText(h *ast.Heading, src []byte) string {
	var buf bytes.Buffer
	_ = ast.Walk(h, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if text, ok := n.(*ast.Text); ok {
			buf.Write(text.Segment.Value(src))
		}
		return ast.WalkContinue, nil
	})
	return buf.String()
}

// extractNodeText extracts text content from various node types.
func extractNodeText(node ast.Node, src []byte) string {
	var buf bytes.Buffer
	_ = ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if text, ok := n.(*ast.Text); ok {
			buf.Write(text.Segment.Value(src))
		}
		return ast.WalkContinue, nil
	})
	return buf.String()
}

// isSeparatorRune returns true if the rune is a separator character
func isSeparatorRune(r rune) bool {
	return r == ' ' || r == '-' || r == '_'
}

// isAlphanumericRune returns true if the rune is alphanumeric or underscore
func isAlphanumericRune(r rune) bool {
	if r >= 'a' && r <= 'z' {
		return true
	}
	if r >= '0' && r <= '9' {
		return true
	}
	return r == '_'
}

// normalizeSection converts heading text to a section key (lowercase, underscores).
// Section headers are recognized case-insensitively with flexible spacing/punctuation.
// Common aliases (e.g., "Implementation Notes" for "implementation") are mapped to canonical names.
func normalizeSection(heading string) string {
	// Normalize: trim, lowercase, replace whitespace/punctuation with underscores
	normalized := strings.ToLower(strings.TrimSpace(heading))
	// Replace any whitespace, hyphens, or other word separators with underscores
	normalized = strings.Map(func(r rune) rune {
		if isSeparatorRune(r) {
			return '_'
		}
		// Keep only alphanumeric and underscores
		if isAlphanumericRune(r) {
			return r
		}
		return '_'
	}, normalized)
	// Clean up multiple consecutive underscores
	for strings.Contains(normalized, "__") {
		normalized = strings.ReplaceAll(normalized, "__", "_")
	}
	normalized = strings.Trim(normalized, "_")

	// Map common aliases to canonical section names (case-insensitively matched)
	aliases := map[string]string{
		"implementation_notes": string(SectionImplementation),
		"impl_notes":           string(SectionImplementation),
		"impl":                 string(SectionImplementation),
		"files_affected":       string(SectionFilesAffected),
		"files":                string(SectionFilesAffected),
		"acceptance_criteria":  string(SectionAcceptanceCriteria),
		"acceptance":           string(SectionAcceptanceCriteria),
		"criteria":             string(SectionAcceptanceCriteria),
	}

	canonical, found := aliases[normalized]
	if found {
		return canonical
	}

	return normalized
}

// Overview returns the overview paragraph (content before first ## section).
func (d *Description) Overview() []string {
	return d.sections[SectionOverview]
}

// SetOverview sets the overview paragraph.
func (d *Description) SetOverview(lines []string) {
	d.sections[SectionOverview] = lines
}

// Scope returns the lines from the "## Scope" section.
func (d *Description) Scope() []string {
	return d.sections[SectionScope]
}

// SetScope sets the "## Scope" section.
func (d *Description) SetScope(lines []string) {
	d.sections[SectionScope] = lines
}

// FilesAffected returns the lines from the "## Files Affected" section.
func (d *Description) FilesAffected() []string {
	return d.sections[SectionFilesAffected]
}

// SetFilesAffected sets the "## Files Affected" section.
func (d *Description) SetFilesAffected(lines []string) {
	d.sections[SectionFilesAffected] = lines
}

// Implementation returns the lines from the "## Implementation Notes" section.
func (d *Description) Implementation() []string {
	return d.sections[SectionImplementation]
}

// SetImplementation sets the "## Implementation Notes" section.
func (d *Description) SetImplementation(lines []string) {
	d.sections[SectionImplementation] = lines
}

// AcceptanceCriteria returns the lines from the "## Acceptance Criteria" section.
func (d *Description) AcceptanceCriteria() []string {
	return d.sections[SectionAcceptanceCriteria]
}

// SetAcceptanceCriteria sets the "## Acceptance Criteria" section.
func (d *Description) SetAcceptanceCriteria(lines []string) {
	d.sections[SectionAcceptanceCriteria] = lines
}

// Verification returns the lines from the "## Verification" section.
func (d *Description) Verification() []string {
	return d.sections[SectionVerification]
}

// SetVerification sets the "## Verification" section.
func (d *Description) SetVerification(lines []string) {
	d.sections[SectionVerification] = lines
}

// Environment returns the lines from the "## Environment" section.
func (d *Description) Environment() []string {
	return d.sections[SectionEnvironment]
}

// SetEnvironment sets the "## Environment" section.
func (d *Description) SetEnvironment(lines []string) {
	d.sections[SectionEnvironment] = lines
}

// reassembleMarkdown reconstructs the markdown from sections, preserving order.
func (d *Description) reassembleMarkdown() string {
	var buf bytes.Buffer

	// If we have raw unstructured content, return it as-is
	if raw, ok := d.sections[SectionRaw]; ok {
		if len(raw) > 0 {
			return raw[0]
		}
	}

	// Helper to convert section key to heading text (known sections only)
	sectionToHeading := map[Section]string{
		SectionOverview:           "", // No heading for overview
		SectionScope:              "Scope",
		SectionFilesAffected:      "Files Affected",
		SectionEnvironment:        "Environment",
		SectionImplementation:     "Implementation Notes",
		SectionAcceptanceCriteria: "Acceptance Criteria",
		SectionVerification:       "Verification",
	}

	// Output sections in order they appeared (or canonical order if no order tracked)
	outputOrder := d.order
	if len(outputOrder) == 0 {
		// Fallback to canonical order if no explicit order was tracked
		outputOrder = []Section{
			SectionOverview,
			SectionScope,
			SectionFilesAffected,
			SectionEnvironment,
			SectionImplementation,
			SectionAcceptanceCriteria,
			SectionVerification,
		}
	}

	for _, section := range outputOrder {
		if content, ok := d.sections[section]; ok {
			if len(content) > 0 {
				// Overview has no heading
				if section == SectionOverview {
					for _, line := range content {
						buf.WriteString(line)
						if !strings.HasSuffix(line, "\n") {
							buf.WriteString("\n")
						}
					}
					buf.WriteString("\n")
				} else {
					// All other sections are level-2 headers
					// Known sections: use standard heading
					// Unknown sections: denormalize the section name to a heading
					var heading string
					if standardHeading, isKnown := sectionToHeading[section]; isKnown {
						heading = standardHeading
					} else {
						// Unknown section: denormalize from snake_case to Title Case
						heading = denormalizeSection(string(section))
					}

					fmt.Fprintf(&buf, "## %s\n\n", heading)
					for _, line := range content {
						buf.WriteString(line)
						if !strings.HasSuffix(line, "\n") {
							buf.WriteString("\n")
						}
					}
					buf.WriteString("\n")
				}
			}
		}
	}

	return buf.String()
}

// denormalizeSection converts a section key back to heading text.
// Reverse of normalizeSection: "implementation_notes" → "Implementation Notes"
func denormalizeSection(key string) string {
	// Split on underscore, capitalize each word, join with space
	words := strings.Split(key, "_")
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
		}
	}
	return strings.Join(words, " ")
}

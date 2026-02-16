package types_test

import (
	"encoding"
	"strings"
	"testing"

	"github.com/selesy/git-bug-ax/internal/types"
)

// TestNewID_ValidHexHash tests NewID with valid hex hashes.
func TestNewID_ValidHexHash(t *testing.T) {
	tests := map[string]struct {
		hash string
	}{
		"lowercase": {
			hash: "abcd1234567890abcd1234567890abcd1234567890abcd1234567890abcd1234",
		},
		"uppercase": {
			hash: "ABCD1234567890ABCD1234567890ABCD1234567890ABCD1234567890ABCD1234",
		},
		"mixed case": {
			hash: "AbCd1234567890AbCd1234567890AbCd1234567890AbCd1234567890AbCd1234",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			id, err := types.NewID(tt.hash)
			if err != nil {
				t.Fatalf("NewID(%q) error = %v, want nil", tt.hash, err)
			}
			if id.String() == "" {
				t.Error("NewID returned empty ID")
			}
		})
	}
}

// TestNewID_InvalidHexHash tests NewID with invalid hex hashes.
func TestNewID_InvalidHexHash(t *testing.T) {
	tests := map[string]struct {
		hash string
	}{
		"non-hex characters": {
			hash: "zzzz1234567890zzzz1234567890zzzz1234567890zzzz1234567890zzzz1234",
		},
		"too short": {
			hash: "abcd1234",
		},
		"empty string": {
			hash: "",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := types.NewID(tt.hash)
			if err == nil {
				t.Errorf("NewID(%q) error = nil, want error", tt.hash)
			}
		})
	}
}

// TestID_MarshalUnmarshalText tests ID TextCodec round-trip.
func TestID_MarshalUnmarshalText(t *testing.T) {
	originalHash := "abcd1234567890abcd1234567890abcd1234567890abcd1234567890abcd1234"
	id, err := types.NewID(originalHash)
	if err != nil {
		t.Fatalf("NewID error: %v", err)
	}

	// Marshal to text
	data, err := id.MarshalText()
	if err != nil {
		t.Fatalf("MarshalText error: %v", err)
	}
	if len(data) == 0 {
		t.Error("MarshalText returned empty bytes")
	}

	// Unmarshal from text
	var id2 types.ID
	err = id2.UnmarshalText(data)
	if err != nil {
		t.Fatalf("UnmarshalText error: %v", err)
	}

	// Verify round-trip
	if id.String() != id2.String() {
		t.Errorf("Round-trip failed: %s != %s", id.String(), id2.String())
	}
}

// TestID_TextCodec verifies ID implements TextCodec.
func TestID_TextCodec(t *testing.T) {
	var _ types.TextCodec = (*types.ID)(nil)
	// Compile-time check passes if this test runs
}

// TestIDs_MarshalUnmarshalText tests IDs TextCodec round-trip.
func TestIDs_MarshalUnmarshalText(t *testing.T) {
	tests := map[string]struct {
		input string
	}{
		"single ID": {
			input: "abcd1234567890abcd1234567890abcd1234567890abcd1234567890abcd1234",
		},
		"empty": {
			input: "",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var ids types.IDs
			err := ids.UnmarshalText([]byte(tt.input))
			if err != nil {
				t.Fatalf("UnmarshalText error: %v", err)
			}

			// Marshal back
			data, err := ids.MarshalText()
			if err != nil {
				t.Fatalf("MarshalText error: %v", err)
			}

			// Unmarshal again and verify
			var ids2 types.IDs
			err = ids2.UnmarshalText(data)
			if err != nil {
				t.Fatalf("UnmarshalText (round-trip) error: %v", err)
			}

			if len(ids) != len(ids2) {
				t.Errorf("Round-trip length mismatch: %d != %d", len(ids), len(ids2))
			}

			for i := range ids {
				if ids[i].String() != ids2[i].String() {
					t.Errorf("Round-trip ID mismatch at index %d: %s != %s", i, ids[i].String(), ids2[i].String())
				}
			}
		})
	}
}

// TestIDs_Multiple tests IDs with multiple values.
func TestIDs_Multiple(t *testing.T) {
	id1, err := types.NewID("abcd1234567890abcd1234567890abcd1234567890abcd1234567890abcd1234")
	if err != nil {
		t.Fatalf("NewID error: %v", err)
	}

	id2, err := types.NewID("fedcba9876543210fedcba9876543210fedcba9876543210fedcba9876543210")
	if err != nil {
		t.Fatalf("NewID error: %v", err)
	}

	ids := types.IDs{id1, id2}

	// Marshal to text
	data, err := ids.MarshalText()
	if err != nil {
		t.Fatalf("MarshalText error: %v", err)
	}
	if !strings.Contains(string(data), ",") {
		t.Error("Multiple IDs should be comma-separated")
	}

	// Unmarshal from text
	var ids2 types.IDs
	err = ids2.UnmarshalText(data)
	if err != nil {
		t.Fatalf("UnmarshalText error: %v", err)
	}

	// Verify
	if len(ids2) != 2 {
		t.Errorf("Expected 2 IDs, got %d", len(ids2))
	}
	if ids2[0].String() != id1.String() || ids2[1].String() != id2.String() {
		t.Error("Round-trip mismatch for multiple IDs")
	}
}

// TestIDs_EmptyMarshaling tests marshaling empty IDs.
func TestIDs_EmptyMarshaling(t *testing.T) {
	var ids types.IDs
	data, err := ids.MarshalText()
	if err != nil {
		t.Fatalf("MarshalText error: %v", err)
	}
	if len(data) != 0 {
		t.Errorf("Empty IDs should marshal to empty bytes, got: %s", string(data))
	}
}

// TestIDs_InvalidMarshaling tests unmarshaling invalid IDs.
func TestIDs_InvalidMarshaling(t *testing.T) {
	invalidInputs := []string{
		"not-a-valid-hex",
		"abcd,1234", // too short
	}

	for _, input := range invalidInputs {
		t.Run(input, func(t *testing.T) {
			var ids types.IDs
			err := ids.UnmarshalText([]byte(input))
			if err == nil {
				t.Errorf("UnmarshalText(%q) error = nil, want error", input)
			}
		})
	}
}

// TestIDs_TextCodec verifies IDs implements TextCodec.
func TestIDs_TextCodec(t *testing.T) {
	var _ types.TextCodec = (*types.IDs)(nil)
	// Compile-time check passes if this test runs
}

// TestLabel_MarshalUnmarshalText tests Label TextCodec round-trip.
func TestLabel_MarshalUnmarshalText(t *testing.T) {
	original := types.Label("test-label")

	// Marshal to text
	data, err := original.MarshalText()
	if err != nil {
		t.Fatalf("MarshalText error: %v", err)
	}
	if string(data) != string(original) {
		t.Errorf("MarshalText mismatch: %s != %s", string(data), string(original))
	}

	// Unmarshal from text
	var label types.Label
	err = label.UnmarshalText(data)
	if err != nil {
		t.Fatalf("UnmarshalText error: %v", err)
	}

	// Verify round-trip
	if label != original {
		t.Errorf("Round-trip failed: %s != %s", label, original)
	}
}

// TestLabel_EmptyString tests Label with empty string.
func TestLabel_EmptyString(t *testing.T) {
	label := types.Label("")
	data, err := label.MarshalText()
	if err != nil {
		t.Fatalf("MarshalText error: %v", err)
	}
	if len(data) != 0 {
		t.Errorf("Empty label should marshal to empty bytes, got: %s", string(data))
	}

	var label2 types.Label
	err = label2.UnmarshalText(data)
	if err != nil {
		t.Fatalf("UnmarshalText error: %v", err)
	}
	if label2 != "" {
		t.Errorf("Unmarshaled empty label should be empty, got: %q", label2)
	}
}

// TestLabel_TextCodec verifies Label implements TextCodec.
func TestLabel_TextCodec(t *testing.T) {
	var _ types.TextCodec = (*types.Label)(nil)
	// Compile-time check passes if this test runs
}

// TestLabel_SpecialCharacters tests Label with special characters.
func TestLabel_SpecialCharacters(t *testing.T) {
	testCases := []string{
		"label-with-dashes",
		"label_with_underscores",
		"label with spaces",
		"label,with,commas",
		"label\nwith\nnewlines",
	}

	for _, tc := range testCases {
		t.Run(tc, func(t *testing.T) {
			label := types.Label(tc)
			data, err := label.MarshalText()
			if err != nil {
				t.Fatalf("MarshalText error: %v", err)
			}

			var label2 types.Label
			err = label2.UnmarshalText(data)
			if err != nil {
				t.Fatalf("UnmarshalText error: %v", err)
			}

			if label2 != label {
				t.Errorf("Round-trip failed: %q != %q", label2, label)
			}
		})
	}
}

// TestTextCodec verifies TextCodec interface.
func TestTextCodec_Interface(t *testing.T) {
	var _ types.TextCodec = (*types.ID)(nil)
	var _ types.TextCodec = (*types.IDs)(nil)
	var _ types.TextCodec = (*types.Label)(nil)
	// Compile-time checks pass if this test runs
}

// TestTextCodec_ImplementsStandardInterfaces verifies TextCodec implements standard encoding interfaces.
func TestTextCodec_ImplementsStandardInterfaces(t *testing.T) {
	var _ encoding.TextMarshaler = (*types.ID)(nil)
	var _ encoding.TextUnmarshaler = (*types.ID)(nil)
	var _ encoding.TextMarshaler = (*types.IDs)(nil)
	var _ encoding.TextUnmarshaler = (*types.IDs)(nil)
	var _ encoding.TextMarshaler = (*types.Label)(nil)
	var _ encoding.TextUnmarshaler = (*types.Label)(nil)
	// Compile-time checks pass if this test runs
}

// TestSet_Add tests Set.Add method.
func TestSet_Add(t *testing.T) {
	s := make(types.Set[types.Label, *types.Label])
	label := types.Label("test")

	// Add should not error
	s.Add(label)

	// Adding same element again should not error
	s.Add(label)
}

// TestSet_Contains tests Set.Contains method.
func TestSet_Contains(t *testing.T) {
	s := make(types.Set[types.Label, *types.Label])
	label := types.Label("test")

	// Element not yet in set
	if s.Contains(label) {
		t.Error("Contains should return false for element not in set")
	}

	// Add element
	s.Add(label)

	// Element should now be in set
	if !s.Contains(label) {
		t.Error("Contains should return true for element in set")
	}
}

// TestSet_Remove tests Set.Remove method.
func TestSet_Remove(t *testing.T) {
	s := make(types.Set[types.Label, *types.Label])
	label := types.Label("test")

	// Add element
	s.Add(label)
	if !s.Contains(label) {
		t.Fatal("Failed to add element to set")
	}

	// Remove element
	s.Remove(label)
	if s.Contains(label) {
		t.Error("Contains should return false after Remove")
	}

	// Removing non-existent element should not error
	s.Remove(label)
}

// TestSet_MarshalUnmarshalText tests Set TextCodec round-trip.
func TestSet_MarshalUnmarshalText(t *testing.T) {
	s := make(types.Set[types.Label, *types.Label])
	s.Add(types.Label("label1"))
	s.Add(types.Label("label2"))
	s.Add(types.Label("label3"))

	// Marshal to text
	data, err := s.MarshalText()
	if err != nil {
		t.Fatalf("MarshalText error: %v", err)
	}
	if len(data) == 0 {
		t.Error("MarshalText returned empty bytes for non-empty set")
	}

	// Unmarshal from text
	var s2 types.Set[types.Label, *types.Label]
	err = s2.UnmarshalText(data)
	if err != nil {
		t.Fatalf("UnmarshalText error: %v", err)
	}

	// Verify all elements are present
	if len(s) != len(s2) {
		t.Errorf("Round-trip length mismatch: %d != %d", len(s), len(s2))
	}

	for label := range s {
		if !s2.Contains(label) {
			t.Errorf("Round-trip failed: element %q missing in s2", label)
		}
	}
}

// TestSet_EmptyMarshaling tests marshaling empty Set.
func TestSet_EmptyMarshaling(t *testing.T) {
	s := make(types.Set[types.Label, *types.Label])
	data, err := s.MarshalText()
	if err != nil {
		t.Fatalf("MarshalText error: %v", err)
	}
	if len(data) != 0 {
		t.Errorf("Empty set should marshal to empty bytes, got: %s", string(data))
	}

	var s2 types.Set[types.Label, *types.Label]
	err = s2.UnmarshalText(data)
	if err != nil {
		t.Fatalf("UnmarshalText error: %v", err)
	}
	if len(s2) != 0 {
		t.Errorf("Unmarshaled empty set should be empty, got length: %d", len(s2))
	}
}

// TestSet_SetMultipleOperations tests multiple operations on Set.
func TestSet_SetMultipleOperations(t *testing.T) {
	s := make(types.Set[types.Label, *types.Label])

	// Add multiple elements
	labels := []types.Label{"a", "b", "c", "d", "e"}
	for _, label := range labels {
		s.Add(label)
	}
	if len(s) != 5 {
		t.Fatalf("Expected 5 elements, got %d", len(s))
	}

	// Verify all are present
	for _, label := range labels {
		if !s.Contains(label) {
			t.Errorf("Set should contain %q", label)
		}
	}

	// Remove some elements
	s.Remove(types.Label("a"))
	s.Remove(types.Label("c"))
	if len(s) != 3 {
		t.Fatalf("After removing 2 elements, expected 3, got %d", len(s))
	}

	// Verify correct elements remain
	if s.Contains(types.Label("a")) || s.Contains(types.Label("c")) {
		t.Error("Removed elements should not be in set")
	}
	if !s.Contains(types.Label("b")) || !s.Contains(types.Label("d")) || !s.Contains(types.Label("e")) {
		t.Error("Remaining elements should be in set")
	}
}

package collections_test

import (
	"testing"

	"github.com/selesy/git-bug-ax/internal/codec"
	"github.com/selesy/git-bug-ax/internal/collections"
)

// TestSet_Add tests Set.Add method.
func TestSet_Add(t *testing.T) {
	s := make(collections.Set[codec.Label, *codec.Label])
	label := codec.Label("test")

	// Add should not error
	s.Add(label)

	// Adding same element again should not error
	s.Add(label)
}

// TestSet_Contains tests Set.Contains method.
func TestSet_Contains(t *testing.T) {
	s := make(collections.Set[codec.Label, *codec.Label])
	label := codec.Label("test")

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
	s := make(collections.Set[codec.Label, *codec.Label])
	label := codec.Label("test")

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
	s := make(collections.Set[codec.Label, *codec.Label])
	s.Add(codec.Label("label1"))
	s.Add(codec.Label("label2"))
	s.Add(codec.Label("label3"))

	// Marshal to text
	data, err := s.MarshalText()
	if err != nil {
		t.Fatalf("MarshalText error: %v", err)
	}
	if len(data) == 0 {
		t.Error("MarshalText returned empty bytes for non-empty set")
	}

	// Unmarshal from text
	var s2 collections.Set[codec.Label, *codec.Label]
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
	s := make(collections.Set[codec.Label, *codec.Label])
	data, err := s.MarshalText()
	if err != nil {
		t.Fatalf("MarshalText error: %v", err)
	}
	if len(data) != 0 {
		t.Errorf("Empty set should marshal to empty bytes, got: %s", string(data))
	}

	var s2 collections.Set[codec.Label, *codec.Label]
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
	s := make(collections.Set[codec.Label, *codec.Label])

	// Add multiple elements
	labels := []codec.Label{"a", "b", "c", "d", "e"}
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
	s.Remove(codec.Label("a"))
	s.Remove(codec.Label("c"))
	if len(s) != 3 {
		t.Fatalf("After removing 2 elements, expected 3, got %d", len(s))
	}

	// Verify correct elements remain
	if s.Contains(codec.Label("a")) || s.Contains(codec.Label("c")) {
		t.Error("Removed elements should not be in set")
	}
	if !s.Contains(codec.Label("b")) || !s.Contains(codec.Label("d")) || !s.Contains(codec.Label("e")) {
		t.Error("Remaining elements should be in set")
	}
}

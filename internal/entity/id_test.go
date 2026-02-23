package entity_test

import (
	"encoding"
	"strings"
	"testing"

	"github.com/selesy/git-bug-ax/internal/entity"
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
			id, err := entity.NewID(tt.hash)
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
			_, err := entity.NewID(tt.hash)
			if err == nil {
				t.Errorf("NewID(%q) error = nil, want error", tt.hash)
			}
		})
	}
}

// TestID_MarshalUnmarshalText tests ID TextCodec round-trip.
func TestID_MarshalUnmarshalText(t *testing.T) {
	originalHash := "abcd1234567890abcd1234567890abcd1234567890abcd1234567890abcd1234"
	id, err := entity.NewID(originalHash)
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
	var id2 entity.ID
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
	var _ encoding.TextMarshaler = (*entity.ID)(nil)
	var _ encoding.TextUnmarshaler = (*entity.ID)(nil)
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
			var ids entity.IDs
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
			var ids2 entity.IDs
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
	id1, err := entity.NewID("abcd1234567890abcd1234567890abcd1234567890abcd1234567890abcd1234")
	if err != nil {
		t.Fatalf("NewID error: %v", err)
	}

	id2, err := entity.NewID("fedcba9876543210fedcba9876543210fedcba9876543210fedcba9876543210")
	if err != nil {
		t.Fatalf("NewID error: %v", err)
	}

	ids := entity.IDs{id1, id2}

	// Marshal to text
	data, err := ids.MarshalText()
	if err != nil {
		t.Fatalf("MarshalText error: %v", err)
	}
	if !strings.Contains(string(data), ",") {
		t.Error("Multiple IDs should be comma-separated")
	}

	// Unmarshal from text
	var ids2 entity.IDs
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
	var ids entity.IDs
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
			var ids entity.IDs
			err := ids.UnmarshalText([]byte(input))
			if err == nil {
				t.Errorf("UnmarshalText(%q) error = nil, want error", input)
			}
		})
	}
}

// TestIDs_TextCodec verifies IDs implements TextCodec.
func TestIDs_TextCodec(t *testing.T) {
	var _ encoding.TextMarshaler = (*entity.IDs)(nil)
	var _ encoding.TextUnmarshaler = (*entity.IDs)(nil)
	// Compile-time check passes if this test runs
}

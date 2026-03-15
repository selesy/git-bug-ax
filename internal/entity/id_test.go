package entity_test

import (
	"encoding"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
			require.NoError(t, err)
			assert.NotEmpty(t, id.String())
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
			require.Error(t, err)
		})
	}
}

// TestID_MarshalUnmarshalText tests ID TextCodec round-trip.
func TestID_MarshalUnmarshalText(t *testing.T) {
	originalHash := "abcd1234567890abcd1234567890abcd1234567890abcd1234567890abcd1234"
	id, err := entity.NewID(originalHash)
	require.NoError(t, err)

	// Marshal to text
	data, err := id.MarshalText()
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	// Unmarshal from text
	var id2 entity.ID
	err = id2.UnmarshalText(data)
	require.NoError(t, err)

	// Verify round-trip
	assert.Equal(t, id.String(), id2.String())
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
			require.NoError(t, err)

			// Marshal back
			data, err := ids.MarshalText()
			require.NoError(t, err)

			// Unmarshal again and verify
			var ids2 entity.IDs
			err = ids2.UnmarshalText(data)
			require.NoError(t, err)

			assert.Len(t, ids2, len(ids))

			for i := range ids {
				assert.Equal(t, ids[i].String(), ids2[i].String())
			}
		})
	}
}

// TestIDs_Multiple tests IDs with multiple values.
func TestIDs_Multiple(t *testing.T) {
	id1, err := entity.NewID("abcd1234567890abcd1234567890abcd1234567890abcd1234567890abcd1234")
	require.NoError(t, err)

	id2, err := entity.NewID("fedcba9876543210fedcba9876543210fedcba9876543210fedcba9876543210")
	require.NoError(t, err)

	ids := entity.IDs{id1, id2}

	// Marshal to text
	data, err := ids.MarshalText()
	require.NoError(t, err)
	assert.Contains(t, string(data), ",")

	// Unmarshal from text
	var ids2 entity.IDs
	err = ids2.UnmarshalText(data)
	require.NoError(t, err)

	// Verify
	assert.Len(t, ids2, 2)
	assert.Equal(t, id1.String(), ids2[0].String())
	assert.Equal(t, id2.String(), ids2[1].String())
}

// TestIDs_EmptyMarshaling tests marshaling empty IDs.
func TestIDs_EmptyMarshaling(t *testing.T) {
	var ids entity.IDs
	data, err := ids.MarshalText()
	require.NoError(t, err)
	assert.Empty(t, data)
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
			require.Error(t, err)
		})
	}
}

// TestIDs_TextCodec verifies IDs implements TextCodec.
func TestIDs_TextCodec(t *testing.T) {
	var _ encoding.TextMarshaler = (*entity.IDs)(nil)
	var _ encoding.TextUnmarshaler = (*entity.IDs)(nil)
	// Compile-time check passes if this test runs
}

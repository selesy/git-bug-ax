package codec_test

import (
	"encoding"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/selesy/git-bug-ax/internal/codec"
)

// TestLabel_MarshalUnmarshalText tests Label TextCodec round-trip.
func TestLabel_MarshalUnmarshalText(t *testing.T) {
	original := codec.Label("test-label")

	// Marshal to text
	data, err := original.MarshalText()
	require.NoError(t, err)
	assert.Equal(t, string(original), string(data))

	// Unmarshal from text
	var label codec.Label
	err = label.UnmarshalText(data)
	require.NoError(t, err)

	// Verify round-trip
	assert.Equal(t, original, label)
}

// TestLabel_EmptyString tests Label with empty string.
func TestLabel_EmptyString(t *testing.T) {
	label := codec.Label("")
	data, err := label.MarshalText()
	require.NoError(t, err)
	assert.Empty(t, data)

	var label2 codec.Label
	err = label2.UnmarshalText(data)
	require.NoError(t, err)
	assert.Equal(t, codec.Label(""), label2)
}

// TestLabel_TextCodec verifies Label implements TextCodec.
func TestLabel_TextCodec(t *testing.T) {
	var _ codec.TextCodec = (*codec.Label)(nil)
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
			label := codec.Label(tc)
			data, err := label.MarshalText()
			require.NoError(t, err)

			var label2 codec.Label
			err = label2.UnmarshalText(data)
			require.NoError(t, err)

			assert.Equal(t, label, label2)
		})
	}
}

// TestTextCodec_ImplementsStandardInterfaces verifies TextCodec implements standard encoding interfaces.
func TestTextCodec_ImplementsStandardInterfaces(t *testing.T) {
	var _ encoding.TextMarshaler = (*codec.Label)(nil)
	var _ encoding.TextUnmarshaler = (*codec.Label)(nil)
	// Compile-time checks pass if this test runs
}

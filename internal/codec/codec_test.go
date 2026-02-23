package codec_test

import (
	"encoding"
	"testing"

	"github.com/selesy/git-bug-ax/internal/codec"
)

// TestLabel_MarshalUnmarshalText tests Label TextCodec round-trip.
func TestLabel_MarshalUnmarshalText(t *testing.T) {
	original := codec.Label("test-label")

	// Marshal to text
	data, err := original.MarshalText()
	if err != nil {
		t.Fatalf("MarshalText error: %v", err)
	}
	if string(data) != string(original) {
		t.Errorf("MarshalText mismatch: %s != %s", string(data), string(original))
	}

	// Unmarshal from text
	var label codec.Label
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
	label := codec.Label("")
	data, err := label.MarshalText()
	if err != nil {
		t.Fatalf("MarshalText error: %v", err)
	}
	if len(data) != 0 {
		t.Errorf("Empty label should marshal to empty bytes, got: %s", string(data))
	}

	var label2 codec.Label
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
			if err != nil {
				t.Fatalf("MarshalText error: %v", err)
			}

			var label2 codec.Label
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

// TestTextCodec_ImplementsStandardInterfaces verifies TextCodec implements standard encoding interfaces.
func TestTextCodec_ImplementsStandardInterfaces(t *testing.T) {
	var _ encoding.TextMarshaler = (*codec.Label)(nil)
	var _ encoding.TextUnmarshaler = (*codec.Label)(nil)
	// Compile-time checks pass if this test runs
}

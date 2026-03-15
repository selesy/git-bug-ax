package collections_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

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
	assert.False(t, s.Contains(label))

	// Add element
	s.Add(label)

	// Element should now be in set
	assert.True(t, s.Contains(label))
}

// TestSet_Remove tests Set.Remove method.
func TestSet_Remove(t *testing.T) {
	s := make(collections.Set[codec.Label, *codec.Label])
	label := codec.Label("test")

	// Add element
	s.Add(label)
	require.True(t, s.Contains(label))

	// Remove element
	s.Remove(label)
	assert.False(t, s.Contains(label))

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
	require.NoError(t, err)
	assert.NotEmpty(t, data)

	// Unmarshal from text
	var s2 collections.Set[codec.Label, *codec.Label]
	err = s2.UnmarshalText(data)
	require.NoError(t, err)

	// Verify all elements are present
	assert.Len(t, s2, len(s))

	for label := range s {
		assert.True(t, s2.Contains(label))
	}
}

// TestSet_EmptyMarshaling tests marshaling empty Set.
func TestSet_EmptyMarshaling(t *testing.T) {
	s := make(collections.Set[codec.Label, *codec.Label])
	data, err := s.MarshalText()
	require.NoError(t, err)
	assert.Empty(t, data)

	var s2 collections.Set[codec.Label, *codec.Label]
	err = s2.UnmarshalText(data)
	require.NoError(t, err)
	assert.Empty(t, s2)
}

// TestSet_SetMultipleOperations tests multiple operations on Set.
func TestSet_SetMultipleOperations(t *testing.T) {
	s := make(collections.Set[codec.Label, *codec.Label])

	// Add multiple elements
	labels := []codec.Label{"a", "b", "c", "d", "e"}
	for _, label := range labels {
		s.Add(label)
	}
	require.Len(t, s, 5)

	// Verify all are present
	for _, label := range labels {
		assert.True(t, s.Contains(label))
	}

	// Remove some elements
	s.Remove(codec.Label("a"))
	s.Remove(codec.Label("c"))
	require.Len(t, s, 3)

	// Verify correct elements remain
	assert.False(t, s.Contains(codec.Label("a")))
	assert.False(t, s.Contains(codec.Label("c")))
	assert.True(t, s.Contains(codec.Label("b")))
	assert.True(t, s.Contains(codec.Label("d")))
	assert.True(t, s.Contains(codec.Label("e")))
}

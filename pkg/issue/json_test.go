package issue_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/selesy/git-bug-ax/pkg/issue"
	"github.com/selesy/git-bug-ax/pkg/issue/issuetest"
)

func TestIssueMarshalJSON(t *testing.T) {
	t.Parallel()

	t.Run("marshals issue with history", func(t *testing.T) {
		t.Parallel()

		_, iss := issuetest.NewTestIssue(
			t,
			issue.WithTitle("Test Issue"),
			issue.WithDescription(newTestDescription(t, "Test description")),
		)

		data, err := json.Marshal(iss)
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(data, &result)
		require.NoError(t, err)

		assert.NotNil(t, result["id"])
		assert.Equal(t, "Test Issue", result["title"])
		assert.NotNil(t, result["history"])
		assert.NotNil(t, result["created"])
		assert.NotNil(t, result["updated"])
	})

	t.Run("history contains operations", func(t *testing.T) {
		t.Parallel()

		_, iss := issuetest.NewTestIssue(
			t,
			issue.WithTitle("Test Issue"),
			issue.WithDescription(newTestDescription(t, "Test description")),
		)

		data, err := json.Marshal(iss)
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(data, &result)
		require.NoError(t, err)

		history, ok := result["history"].([]interface{})
		require.True(t, ok, "history should be an array")
		assert.Len(t, history, 1) // CreateOperation

		op, ok := history[0].(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "CreateOperation", op["name"])
		assert.NotNil(t, op["time"])
		assert.NotNil(t, op["author"])
	})
}

func TestExcerptMarshalJSON(t *testing.T) {
	t.Parallel()

	t.Run("marshals excerpt without history", func(t *testing.T) {
		t.Parallel()

		_, iss := issuetest.NewTestIssue(
			t,
			issue.WithTitle("Test Issue"),
			issue.WithDescription(newTestDescription(t, "Test description")),
		)

		excerpt := iss.Excerpt()
		data, err := json.Marshal(excerpt)
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(data, &result)
		require.NoError(t, err)

		assert.NotNil(t, result["id"])
		assert.Equal(t, "Test Issue", result["title"])
		assert.Nil(t, result["history"])
		assert.Nil(t, result["created"])
		assert.Nil(t, result["updated"])
	})

	t.Run("excerpt contains metadata fields", func(t *testing.T) {
		t.Parallel()

		_, iss := issuetest.NewTestIssue(
			t,
			issue.WithTitle("Test Issue"),
			issue.WithDescription(newTestDescription(t, "Test description")),
			issue.WithPriority(issue.PriorityHigh),
			issue.WithStatus(issue.StatusReady),
			issue.WithType(issue.TypeBug),
		)

		excerpt := iss.Excerpt()
		data, err := json.Marshal(excerpt)
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(data, &result)
		require.NoError(t, err)

		assert.NotNil(t, result["id"])
		assert.Equal(t, "Test Issue", result["title"])
		assert.NotNil(t, result["type"])
		assert.NotNil(t, result["status"])
		assert.NotNil(t, result["priority"])
	})
}

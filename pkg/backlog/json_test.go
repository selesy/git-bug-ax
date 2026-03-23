package backlog_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/selesy/git-bug-agent/pkg/backlog"
	"github.com/selesy/git-bug-agent/pkg/backlog/backlogtest"
	"github.com/selesy/git-bug-agent/pkg/issue"
)

func TestIssueMarshalJSON(t *testing.T) {
	t.Parallel()

	t.Run("marshals issue with history", func(t *testing.T) {
		t.Parallel()

		desc := newTestDescriptionJSON(t, "Test description")

		_, iss := newTestIssueJSON(
			t,
			issue.TypeTask,
			"Test Issue",
			&desc,
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

		desc := newTestDescriptionJSON(t, "Test description")

		_, iss := newTestIssueJSON(
			t,
			issue.TypeTask,
			"Test Issue",
			&desc,
		)

		data, err := json.Marshal(iss)
		require.NoError(t, err)

		var result map[string]interface{}
		err = json.Unmarshal(data, &result)
		require.NoError(t, err)

		history, ok := result["history"].([]interface{})
		require.True(t, ok, "history should be an array")
		assert.Len(t, history, 2) // CreateOperation and SetMetadataOperation

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

		desc := newTestDescriptionJSON(t, "Test description")

		_, iss := newTestIssueJSON(
			t,
			issue.TypeTask,
			"Test Issue",
			&desc,
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

		desc := newTestDescriptionJSON(t, "Test description")

		_, iss := newTestIssueJSON(
			t,
			issue.TypeTask,
			"Test Issue",
			&desc,
			backlog.WithPriority(issue.PriorityHigh),
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

func newTestIssueJSON(t *testing.T, typ issue.Type, title string, description *issue.Description, opts ...backlog.CreateOption) (*backlog.Index, *backlog.Issue) {
	t.Helper()

	// env, _ := testenv.NewTestEnvAndUser(t)
	// user, err := env.Backend.GetUserIdentity()
	// require.NoError(t, err)

	// iss, err := backlog.Create(env, opts...)
	// require.NoError(t, err)
	// require.NoError(t, iss.Commit(user))

	path := backlogtest.NewGitbugRepoPath(t, backlogtest.WithIdentityCount(1))

	ind := backlogtest.NewIndex(
		t,
		backlog.WithRepoPath(path),
	)

	iss, err := ind.Create(typ, title, description, opts...)
	require.NoError(t, err)
	require.NoError(t, iss.Commit())

	return ind, iss
}

func newTestDescriptionJSON(t *testing.T, description string) issue.Description {
	t.Helper()

	var d issue.Description
	require.NoError(t, (&d).UnmarshalText([]byte(description)))

	return d
}

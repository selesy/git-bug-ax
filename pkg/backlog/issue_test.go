package backlog_test

import (
	"maps"
	"testing"

	"github.com/git-bug/git-bug/cache"
	"github.com/git-bug/git-bug/commands/bug/testenv"
	"github.com/git-bug/git-bug/commands/execenv"
	"github.com/git-bug/git-bug/entities/bug"
	"github.com/git-bug/git-bug/entity/dag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/selesy/git-bug-agent/internal/metadata"
	"github.com/selesy/git-bug-agent/pkg/backlog"
	"github.com/selesy/git-bug-agent/pkg/issue"
)

func TestCreate(t *testing.T) {
	t.Parallel()

	const (
		testTitle       = "Issue with a title and a description"
		testDescription = "This description should appear as the issue's first comment"
	)

	t.Run("with missing WithTitle option", func(t *testing.T) {
		t.Parallel()

		env, _ := testenv.NewTestEnvAndUser(t)

		_, err := backlog.Create(env)
		require.ErrorIs(t, err, issue.ErrNoTitle)
	})

	t.Run("with empty WithTitle option", func(t *testing.T) {
		t.Parallel()

		env, _ := testenv.NewTestEnvAndUser(t)

		_, err := backlog.Create(env, backlog.WithTitle(""))
		require.ErrorIs(t, err, issue.ErrNoTitle)
	})

	t.Run("with only valid WithTitle option", func(t *testing.T) {
		t.Parallel()

		const testTitle = "Issue with only a title"

		env, iss := newTestIssue(
			t,
			backlog.WithTitle(testTitle),
		)

		bug, err := env.Backend.Bugs().ResolvePrefix(iss.ID().String())
		require.NoError(t, err)
		snap := bug.Snapshot()
		assert.Len(t, snap.Operations, 1)
		assert.Equal(t, testTitle, snap.Title)
		assert.Len(t, snap.Comments, 1)
		assert.Empty(t, snap.Comments[0].Message)
	})

	t.Run("with a valid WithTitle option and an empty WithDescription option", func(t *testing.T) {
		t.Parallel()

		const testTitle = "Issue with a title and an empty description"

		env, iss := newTestIssue(
			t,
			backlog.WithTitle(testTitle),
			backlog.WithDescription(newTestDescription(t, "")),
		)

		bug, err := env.Backend.Bugs().ResolvePrefix(iss.ID().String())
		require.NoError(t, err)
		snap := bug.Snapshot()
		assert.Len(t, snap.Operations, 1)
		assert.Equal(t, testTitle, snap.Title)
		assert.Len(t, snap.Comments, 1)
		assert.Empty(t, snap.Comments[0].Message)
	})

	t.Run("with valid WithTitle and WithDescription options", func(t *testing.T) {
		t.Parallel()

		env, iss := newTestIssue(
			t,
			backlog.WithTitle(testTitle),
			backlog.WithDescription(newTestDescription(t, testDescription)),
		)

		bug, err := env.Backend.Bugs().ResolvePrefix(iss.ID().String())
		require.NoError(t, err)
		snap := bug.Snapshot()
		assert.Len(t, snap.Operations, 1)
		assert.Equal(t, testTitle, snap.Title)
		assert.Len(t, snap.Comments, 1)
		assert.Equal(t, testDescription, snap.Comments[0].Message)
	})

	t.Run("with valid WithTitle, WithDescription and metadata options", func(t *testing.T) {
		t.Parallel()

		env, iss := newTestIssue(
			t,
			backlog.WithTitle(testTitle),
			backlog.WithDescription(newTestDescription(t, testDescription)),
			backlog.WithPriority(issue.PriorityHigh),
			backlog.WithStatus(issue.StatusDraft),
			backlog.WithType(issue.TypeBug),
		)

		bug, err := env.Backend.Bugs().ResolvePrefix(iss.ID().String())
		require.NoError(t, err)
		snap := bug.Snapshot()
		assert.Len(t, snap.Operations, 2)
		assert.Equal(t, testTitle, snap.Title)
		assert.Len(t, snap.Comments, 1)
		assert.Equal(t, testDescription, snap.Comments[0].Message)
		assert.Contains(t, mutableMetadata(t, bug), metadata.KeyPriority)
		assert.Equal(t, issue.PriorityHigh.String(), mutableMetadata(t, bug)[metadata.KeyPriority])
		assert.Contains(t, mutableMetadata(t, bug), metadata.KeyStatus)
		assert.Equal(t, issue.StatusDraft.String(), mutableMetadata(t, bug)[metadata.KeyStatus])
		assert.Contains(t, mutableMetadata(t, bug), metadata.KeyType)
		assert.Equal(t, issue.TypeBug.String(), mutableMetadata(t, bug)[metadata.KeyType])
	})
}

func TestIssue_Parent(t *testing.T) {
	t.Run("Issue has no parent", func(t *testing.T) {
		t.Parallel()

		env, bugID, commentID := testenv.NewTestEnvAndBugWithComment(t)
		_ = commentID

		bug, err := env.Backend.Bugs().ResolvePrefix(bugID.String())
		require.NoError(t, err)

		iss, err := backlog.Wrap(bug)
		require.NoError(t, err)

		_, err = iss.Parent()
		require.ErrorIs(t, err, issue.ErrNoParent)
	})

	t.Run("Issue has a parent", func(t *testing.T) {
		t.Parallel()

		env, bugID, _ := testenv.NewTestEnvAndBugWithComment(t)
		bug, err := env.Backend.Bugs().ResolvePrefix(bugID.String())
		require.NoError(t, err)

		_, err = bug.SetMetadata(bugID, map[string]string{"gba_parent": bugID.String()})
		require.NoError(t, err)

		iss, err := backlog.Wrap(bug)
		require.NoError(t, err)

		issID, err := issue.NewID(bugID.String())
		require.NoError(t, err)

		parentID, err := iss.Parent()
		require.NoError(t, err)
		assert.Equal(t, issID, parentID)
	})
}

func TestIssue_SetParent(t *testing.T) {
	t.Parallel()

	env, bugID, commentID := testenv.NewTestEnvAndBugWithComment(t)
	_ = commentID

	bug, err := env.Backend.Bugs().ResolvePrefix(bugID.String())
	require.NoError(t, err)
	require.NotContains(t, mutableMetadata(t, bug), metadata.KeyParent)

	iss, err := backlog.Wrap(bug)
	require.NoError(t, err)
	issID, err := issue.NewID(bugID.String())
	require.NoError(t, err)

	user, err := env.Backend.GetUserIdentity()
	require.NoError(t, err)

	iss.SetParent(issID)

	require.NoError(t, iss.Commit(user))

	updatedBug, err := env.Backend.Bugs().ResolvePrefix(bugID.String())
	require.NoError(t, err)
	assert.Contains(t, mutableMetadata(t, updatedBug), metadata.KeyParent)
}

func mutableMetadata(t *testing.T, bugCache *cache.BugCache) map[string]string {
	t.Helper()

	metadata := map[string]string{}
	for _, op := range bugCache.Snapshot().Operations {
		metaOp, ok := op.(*dag.SetMetadataOperation[*bug.Snapshot])
		if !ok {
			continue
		}

		maps.Copy(metadata, metaOp.NewMetadata)
	}

	return metadata
}

func newTestDescription(t *testing.T, description string) issue.Description {
	t.Helper()

	var d issue.Description
	require.NoError(t, (&d).UnmarshalText([]byte(description)))

	return d
}

func newTestIssue(t *testing.T, opts ...backlog.IssueOption) (*execenv.Env, *backlog.Issue) {
	t.Helper()

	env, _ := testenv.NewTestEnvAndUser(t)
	user, err := env.Backend.GetUserIdentity()
	require.NoError(t, err)

	iss, err := backlog.Create(env, opts...)
	require.NoError(t, err)
	require.NoError(t, iss.Commit(user))

	return env, iss
}

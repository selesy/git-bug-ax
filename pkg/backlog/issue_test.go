package backlog_test

import (
	"maps"
	"testing"

	"github.com/git-bug/git-bug/cache"
	"github.com/git-bug/git-bug/entities/bug"
	"github.com/git-bug/git-bug/entity/dag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/selesy/git-bug-agent/internal/gitbug"
	"github.com/selesy/git-bug-agent/internal/metadata"
	"github.com/selesy/git-bug-agent/pkg/backlog"
	"github.com/selesy/git-bug-agent/pkg/backlog/backlogtest"
	"github.com/selesy/git-bug-agent/pkg/issue"
)

func TestCreate(t *testing.T) {
	t.Parallel()

	const (
		testTitle       = "Issue with a title and a description"
		testDescription = "This description should appear as the issue's first comment"
	)

	t.Run("with empty title", func(t *testing.T) {
		t.Parallel()

		path := backlogtest.NewGitbugRepoPath(t, backlogtest.WithIdentityCount(1))
		ind := backlogtest.NewIndex(t, backlog.WithRepoPath(path))

		_, err := ind.Create(issue.TypeTask, "", nil)
		require.ErrorIs(t, err, issue.ErrNoTitle)
	})

	t.Run("with only a title", func(t *testing.T) {
		t.Parallel()

		bug := createIssueAndGetAssociatedBug(t, issue.TypeTask, testTitle, nil)
		snap := bug.Snapshot()
		assert.Len(t, snap.Operations, 2)
		assert.Equal(t, testTitle, snap.Title)
		assert.Len(t, snap.Comments, 1)
		assert.Empty(t, snap.Comments[0].Message)
	})

	t.Run("with a valid WithTitle option and an empty WithDescription option", func(t *testing.T) {
		t.Parallel()

		desc := newTestDescription(t, "")
		bug := createIssueAndGetAssociatedBug(t, issue.TypeTask, testTitle, desc)
		snap := bug.Snapshot()
		assert.Len(t, snap.Operations, 2)
		assert.Equal(t, testTitle, snap.Title)
		assert.Len(t, snap.Comments, 1)
		assert.Empty(t, snap.Comments[0].Message)
	})

	t.Run("with valid WithTitle and WithDescription options", func(t *testing.T) {
		t.Parallel()

		desc := newTestDescription(t, testDescription)
		bug := createIssueAndGetAssociatedBug(t, issue.TypeTask, testTitle, desc)
		snap := bug.Snapshot()
		assert.Len(t, snap.Operations, 2)
		assert.Equal(t, testTitle, snap.Title)
		assert.Len(t, snap.Comments, 1)
		assert.Equal(t, testDescription, snap.Comments[0].Message)
	})

	t.Run("with valid WithTitle, WithDescription and metadata options", func(t *testing.T) {
		t.Parallel()

		desc := newTestDescription(t, testDescription)
		bug := createIssueAndGetAssociatedBug(t, issue.TypeBug, testTitle, desc, backlog.WithPriority(issue.PriorityHigh))

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

		path := backlogtest.NewGitbugRepoPath(t, backlogtest.WithIdentityCount(1))

		cfg, err := gitbug.NewConfig(gitbug.WithRepoPath(path))
		require.NoError(t, err)
		backend, _, err := gitbug.NewBackendAndRepo(t.Context(), cfg)
		require.NoError(t, err)

		bug, _, err := backend.Bugs().New("test title", "")
		require.NoError(t, err)
		require.NoError(t, backend.Close())

		ind := backlogtest.NewIndex(t, backlog.WithRepoPath(path))
		iss, err := ind.ResolvePrefix(bug.Id().String())
		require.NoError(t, err)

		_, err = iss.Parent()
		require.ErrorIs(t, err, issue.ErrNoParent)
	})

	t.Run("Issue has a parent", func(t *testing.T) {
		t.Parallel()

		path := backlogtest.NewGitbugRepoPath(t, backlogtest.WithIdentityCount(1))

		cfg, err := gitbug.NewConfig(gitbug.WithRepoPath(path))
		require.NoError(t, err)
		backend, _, err := gitbug.NewBackendAndRepo(t.Context(), cfg)
		require.NoError(t, err)

		bug, _, err := backend.Bugs().New("test title", "")
		require.NoError(t, err)
		_, err = bug.SetMetadata(bug.Id(), map[string]string{"gba_parent": bug.Id().String()})
		require.NoError(t, err)
		require.NoError(t, backend.Close())

		ind := backlogtest.NewIndex(t, backlog.WithRepoPath(path))
		iss, err := ind.ResolvePrefix(bug.Id().String())
		require.NoError(t, err)

		_, err = iss.Parent()
		require.ErrorIs(t, err, issue.ErrNoParent)
	})
}

func TestIssue_SetParent(t *testing.T) {
	t.Parallel()

	path := backlogtest.NewGitbugRepoPath(t, backlogtest.WithIssueCount(2))
	ind := backlogtest.NewIndex(t, backlog.WithRepoPath(path))
	issueIDs, err := ind.AllIDs()
	require.NoError(t, err)
	require.Len(t, issueIDs, 2)

	iss, err := ind.Resolve(issueIDs[0])
	require.NoError(t, err)
	iss.SetParent(issueIDs[1])
	require.NoError(t, iss.Commit())

	require.NoError(t, ind.Close())

	cfg, err := gitbug.NewConfig(gitbug.WithRepoPath(path))
	require.NoError(t, err)
	backend, _, err := gitbug.NewBackendAndRepo(t.Context(), cfg)
	require.NoError(t, err)

	bug, err := backend.Bugs().Resolve(issueIDs[0].Id)
	require.NoError(t, err)
	meta := mutableMetadata(t, bug)
	assert.Contains(t, meta, metadata.KeyParent)
	assert.Equal(t, issueIDs[1].String(), meta[metadata.KeyParent])
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

func newTestDescription(t *testing.T, description string) *issue.Description {
	t.Helper()

	var d issue.Description
	require.NoError(t, (&d).UnmarshalText([]byte(description)))

	return &d
}

func createIssueAndGetAssociatedBug(t *testing.T, typ issue.Type, title string, description *issue.Description, opts ...backlog.CreateOption) *cache.BugCache {
	path := backlogtest.NewGitbugRepoPath(t, backlogtest.WithIdentityCount(1))
	ind := backlogtest.NewIndex(t, backlog.WithRepoPath(path))

	iss, err := ind.Create(typ, title, description, opts...)
	require.NoError(t, err)
	require.NoError(t, iss.Commit())
	require.NoError(t, ind.Close())

	cfg, err := gitbug.NewConfig(gitbug.WithRepoPath(path))
	require.NoError(t, err)
	backend, _, err := gitbug.NewBackendAndRepo(t.Context(), cfg)
	require.NoError(t, err)

	bug, err := backend.Bugs().ResolvePrefix(iss.ID().String())
	require.NoError(t, err)

	return bug
}

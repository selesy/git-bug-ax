package backlog_test

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/selesy/deterministic"

	"github.com/selesy/git-bug-agent/internal/gitbug"
	"github.com/selesy/git-bug-agent/internal/metadata"
	"github.com/selesy/git-bug-agent/pkg/backlog"
	"github.com/selesy/git-bug-agent/pkg/backlog/backlogtest"
	"github.com/selesy/git-bug-agent/pkg/issue"
)

func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("with no options", func(t *testing.T) {
		// t.Parallel()

		backlog, err := backlog.New(t.Context())
		require.NoError(t, err)

		t.Cleanup(func() {
			require.NoError(t, backlog.Close())
		})
	})

	t.Run("with WithEnsureUser", func(t *testing.T) {
		// t.Parallel()

		backlog, err := backlog.New(t.Context(), backlog.WithEnsureUser(true))
		require.NoError(t, err)

		t.Cleanup(func() {
			require.NoError(t, backlog.Close())
		})
	})

	t.Run("with WithEnsureUser fails", func(t *testing.T) {
		t.Parallel()

		const exp = "No identity is set.\nTo interact with bugs, an identity first needs to be created using \"git bug user new\" or adopted with \"git bug user adopt\""

		_, _, path := backlogtest.NewGitbugBackendRepoAndRepoPath(t)
		backlog, err := backlog.New(t.Context(), backlog.WithRepoPath(path), backlog.WithEnsureUser(true))
		require.EqualError(t, err, exp)
		assert.Nil(t, backlog)
	})

	t.Run("with WithEnsureUser passes", func(t *testing.T) {
		t.Parallel()

		path := backlogtest.NewGitbugRepoPath(t, backlogtest.WithIdentityCount(1))
		backlog, err := backlog.New(t.Context(), backlog.WithRepoPath(path), backlog.WithEnsureUser(true))

		require.NoError(t, err)
		assert.NotNil(t, backlog)
	})
}

func TestBacklog_Close(t *testing.T) {
	t.Parallel()

	recorder := deterministic.NewSlogRecorder()
	logger := slog.New(recorder)
	path := backlogtest.NewGitbugRepoPath(t, backlogtest.WithIdentityCount(1))

	backlog, err := backlog.New(
		t.Context(),
		backlog.WithRepoPath(path),
		backlog.WithLogger(logger),
	)
	require.NoError(t, err)
	require.NoError(t, backlog.Close())
	require.Equal(t, 2, recorder.Len())

	assertLoggerAttrs := func(t *testing.T, r assertableSlogRecord) {
		assert.Equal(t, "00000000000000000000000000000000", r.Attr(t, "trace_id").String())
		assert.Equal(t, "0000000000000000", r.Attr(t, "span_id").String())
		assert.NotEmpty(t, r.Attr(t, "identity").String())
	}

	rec, err := recorder.Get(0)
	require.NoError(t, err)
	assert.Equal(t, "backend started", rec.Message)
	assert.Equal(t, slog.LevelDebug, rec.Level)
	assertRec := newAssertableSlogRecord(rec)
	assertLoggerAttrs(t, assertRec)
	assert.Zero(t, assertRec.Attr(t, "bugs").Int64())
	assert.Equal(t, int64(1), assertRec.Attr(t, "identities").Int64())

	rec, err = recorder.Get(1)
	require.NoError(t, err)
	assert.Equal(t, "backend closed", rec.Message)
	assert.Equal(t, slog.LevelDebug, rec.Level)
	assertRec = newAssertableSlogRecord(rec)
	assertLoggerAttrs(t, assertRec)
}

func TestIndex_Create(t *testing.T) {
	t.Parallel()

	path := backlogtest.NewGitbugRepoPath(t, backlogtest.WithIdentityCount(1))
	ind, err := backlog.New(t.Context(), backlog.WithRepoPath(path))
	require.NoError(t, err)

	desc := newTestDescription(t, "test description")

	iss, err := ind.Create(issue.TypeEpic, "test issue", desc)
	require.NoError(t, err)
	require.NoError(t, iss.Commit())
	require.NoError(t, ind.Close())

	cfg, err := gitbug.NewConfig(gitbug.WithRepoPath(path))
	require.NoError(t, err)
	backend, _, err := gitbug.NewBackendAndRepo(t.Context(), cfg)
	require.NoError(t, err)
	bug, err := backend.Bugs().Resolve(iss.ID().Id)
	require.NoError(t, err)
	snap := bug.Snapshot()
	assert.Equal(t, "test issue", snap.Title)
	assert.Equal(t, "test description", snap.Comments[0].Message)
	meta := mutableMetadata(t, bug)
	assert.Equal(t, issue.TypeEpic.String(), meta[metadata.KeyType])
	assert.Equal(t, issue.StatusDraft.String(), meta[metadata.KeyStatus])
}

func TestIndex_Resolve(t *testing.T) {
	t.Parallel()

	path := backlogtest.NewGitbugRepoPath(t, backlogtest.WithIssueCount(1))
	cfg, err := gitbug.NewConfig(gitbug.WithRepoPath(path))
	require.NoError(t, err)
	backend, _, err := gitbug.NewBackendAndRepo(t.Context(), cfg)
	require.NoError(t, err)
	require.Len(t, backend.Bugs().AllIds(), 1)
	bug, err := backend.Bugs().Resolve(backend.Bugs().AllIds()[0])
	require.NoError(t, err)
	_, err = bug.SetMetadata(bug.Id(), map[string]string{
		metadata.KeyType:   issue.TypeEpic.String(),
		metadata.KeyStatus: issue.StatusDraft.String(),
	})
	require.NoError(t, err)
	require.NoError(t, bug.Commit())
	require.NoError(t, backend.Close())

	id, err := issue.WrapID(bug.Id())
	require.NoError(t, err)
	ind, err := backlog.New(t.Context(), backlog.WithRepoPath(path))
	require.NoError(t, err)
	iss, err := ind.Resolve(id)
	require.NoError(t, err)
	assert.Equal(t, "issue #0", iss.Title())
	desc := iss.Description()
	assert.Equal(t, "This is test issue #0", (&desc).Raw())
	assert.Equal(t, issue.TypeEpic, iss.Type())
	assert.Equal(t, issue.StatusDraft, iss.Status())
}

type assertableSlogRecord struct {
	slog.Record
	attrs map[string]slog.Value
}

func newAssertableSlogRecord(record slog.Record) assertableSlogRecord {
	attrs := make(map[string]slog.Value, record.NumAttrs())
	record.Attrs(func(attr slog.Attr) bool {
		attrs[attr.Key] = attr.Value

		return true
	})

	return assertableSlogRecord{
		Record: record,
		attrs:  attrs,
	}
}

func (r assertableSlogRecord) Attr(t *testing.T, key string) slog.Value {
	v, ok := r.attrs[key]
	require.True(t, ok)

	return v
}

package ax_test

import (
	"bytes"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/selesy/deterministic"

	"github.com/selesy/git-bug-ax/pkg/ax"
	"github.com/selesy/git-bug-ax/pkg/ax/axtest"
)

func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("with no options", func(t *testing.T) {
		// t.Parallel()

		backlog, err := ax.New(t.Context())
		require.NoError(t, err)

		t.Cleanup(func() {
			require.NoError(t, backlog.Close())
		})
	})

	t.Run("with WithEnsureUser", func(t *testing.T) {
		// t.Parallel()

		backlog, err := ax.New(t.Context(), ax.WithEnsureUser(true))
		require.NoError(t, err)

		t.Cleanup(func() {
			require.NoError(t, backlog.Close())
		})
	})

	t.Run("with WithEnsureUser fails", func(t *testing.T) {
		t.Parallel()

		const exp = "No identity is set.\nTo interact with bugs, an identity first needs to be created using \"git bug user new\" or adopted with \"git bug user adopt\""

		path := axtest.NewRepo(t)
		backlog, err := ax.New(t.Context(), ax.WithRepoPath(path), ax.WithEnsureUser(true))
		require.EqualError(t, err, exp)
		assert.Nil(t, backlog)
	})

	t.Run("with WithEnsureUser passes", func(t *testing.T) {
		t.Parallel()

		path := axtest.NewRepo(t, axtest.WithIdentityCount(1))
		backlog, err := ax.New(t.Context(), ax.WithRepoPath(path), ax.WithEnsureUser(true))

		require.NoError(t, err)
		assert.NotNil(t, backlog)
	})
}

func TestBacklog_Close(t *testing.T) {
	t.Parallel()

	const exp = `time=2006-01-02T15:04:05.000Z level=DEBUG msg="backend started" identities=1 bugs=0
time=2006-01-02T15:04:06.000Z level=DEBUG msg="backend closed"
`

	buf := &bytes.Buffer{}
	handler := deterministic.NewSlogHandler(
		slog.NewTextHandler(buf, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		}),
		deterministic.NowFunc(),
	)
	logger := slog.New(handler)

	backlog, err := ax.New(
		t.Context(),
		ax.WithRepoPath(axtest.NewRepo(t, axtest.WithIdentityCount(1))),
		ax.WithLogger(logger),
	)
	require.NoError(t, err)
	require.NoError(t, backlog.Close())
	assert.Equal(t, exp, buf.String())
}

// func testBacklog(t *testing.T, opts ...ax.Option) *ax.Backlog {
// 	t.Helper()

// 	backlog, err := ax.New(t.Context(), opts...)
// 	require.NoError(t, err)

// 	t.Cleanup(func() {
// 		require.NoError(t, backlog.Close())
// 	})
// }

package backlogtest_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/selesy/git-bug-ax/pkg/backlog/backlogtest"
)

func TestNew(t *testing.T) {
	t.Parallel()

	t.Run("with additional path elements", func(t *testing.T) {
		path := backlogtest.NewRepo(t, backlogtest.WithSubdirCount(5))
		require.True(t, strings.HasSuffix(path, "/a/b/c/d/e"))
	})

	t.Run("with generated issues", func(t *testing.T) {
		t.Parallel()

		t.Run("multiple identities and issues", func(t *testing.T) {
			t.Parallel()

			path := backlogtest.NewRepo(
				t,
				// The order of thesee options matter to verify the identity count isn't overwritten
				backlogtest.WithIdentityCount(5),
				backlogtest.WithIssueCount(20),
			)

			assertRefCount(t, filepath.Join(path, ".git", "refs", "bugs"), 20)
			assertRefCount(t, filepath.Join(path, ".git", "refs", "identities"), 5)
		})

		t.Run("identity requested for issue creation", func(t *testing.T) {
			t.Parallel()

			path := backlogtest.NewRepo(
				t,
				backlogtest.WithIssueCount(20),
				backlogtest.WithIdentityCount(1),
			)

			assertRefCount(t, filepath.Join(path, ".git", "refs", "bugs"), 20)
			assertRefCount(t, filepath.Join(path, ".git", "refs", "identities"), 1)
		})

		t.Run("no identity for issue creation", func(t *testing.T) {
			t.Parallel()

			path := backlogtest.NewRepo(t, backlogtest.WithIssueCount(20))

			assertRefCount(t, filepath.Join(path, ".git", "refs", "bugs"), 20)
			assertRefCount(t, filepath.Join(path, ".git", "refs", "identities"), 1)
		})

		t.Run("no identities or issues", func(t *testing.T) {
			t.Parallel()

			path := backlogtest.NewRepo(t)

			assertRefCount(t, filepath.Join(path, ".git", "refs", "bugs"), 0)
			assertRefCount(t, filepath.Join(path, ".git", "refs", "identities"), 0)
		})
	})
}

func assertRefCount(t *testing.T, path string, count int) {
	t.Helper()

	de, err := os.ReadDir(path)
	if err != nil && count == 0 && errors.Is(err, os.ErrNotExist) {
		return
	}

	assert.Len(t, de, count)
}

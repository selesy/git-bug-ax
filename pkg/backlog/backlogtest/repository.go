package backlogtest

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/git-bug/git-bug/cache"
	"github.com/git-bug/git-bug/repository"
	"github.com/stretchr/testify/require"

	"github.com/selesy/git-bug-agent/internal/gitbug"
)

func NewGitbugBackendRepoAndRepoPath(t *testing.T, opts ...Option) (*cache.RepoCache, *repository.GoGitRepo, string) {
	t.Helper()

	cfg := newConfig(t, opts...)
	backend, repo, path := newGitbugBackendRepoAndRepoPath(t, cfg)

	t.Cleanup(func() {
		require.NoError(t, backend.Close())
	})

	return backend, repo, filepath.Join(append([]string{path}, cfg.subdir...)...)
}

func NewGitbugRepoPath(t *testing.T, opts ...Option) string {
	t.Helper()

	cfg := newConfig(t, opts...)
	backend, _, path := newGitbugBackendRepoAndRepoPath(t, cfg)
	require.NoError(t, backend.Close())

	return path
}

func newGitbugBackendRepoAndRepoPath(t *testing.T, cfg *config) (*cache.RepoCache, *repository.GoGitRepo, string) {
	t.Helper()

	path := t.TempDir()

	repo, err := repository.InitGoGitRepo(path, gitbug.Namespace)
	require.NoError(t, err)

	backend, events := cache.NewRepoCache(repo)

	for event := range events {
		require.NoError(t, event.Err)
	}

	for i := range cfg.userCount {
		user, err := backend.Identities().New(fmt.Sprintf("identity-%d", i), fmt.Sprintf("identity-%d@example.com", i))
		require.NoError(t, err)

		if i == 0 {
			require.NoError(t, backend.SetUserIdentity(user))
		}
	}

	for i := range cfg.bugCount {
		_, _, err := backend.Bugs().New(fmt.Sprintf("issue #%d", i), fmt.Sprintf("This is test issue #%d", i))
		require.NoError(t, err)
	}

	return backend, repo, path
}

type config struct {
	subdir    []string
	bugCount  int
	userCount int
}

func newConfig(t *testing.T, opts ...Option) *config {
	t.Helper()

	var cfg config

	for _, opt := range opts {
		opt(&cfg)
	}

	return &cfg
}

type Option func(*config)

func WithIdentityCount(n int) Option {
	return func(cfg *config) {
		cfg.userCount = n
	}
}

func WithIssueCount(n int) Option {
	return func(cfg *config) {
		if n != 0 && cfg.userCount == 0 {
			cfg.userCount = 1
		}

		cfg.bugCount = n
	}
}

func WithSubdirCount(n int) Option {
	return func(cfg *config) {
		for i := range n {
			cfg.subdir = append(cfg.subdir, string('a'+rune(i)))
		}
	}
}

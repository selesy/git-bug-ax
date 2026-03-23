package gitbug

import (
	"context"
	"errors"

	"github.com/git-bug/git-bug/cache"
	"github.com/git-bug/git-bug/entities/identity"
	"github.com/git-bug/git-bug/repository"
)

// Namespace is the git-bug namespace identifier.
const Namespace = "git-bug"

func NewBackendAndRepo(ctx context.Context, cfg *Config) (*cache.RepoCache, *repository.GoGitRepo, error) {
	var repo *repository.GoGitRepo
	repo, err := repository.OpenGoGitRepo(cfg.RepoPath(), Namespace, []repository.ClockLoader{})
	if err != nil {
		return nil, nil, err
	}

	_, err = identity.GetUserIdentity(repo)
	if (cfg.ensureUser && errors.Is(err, identity.ErrNoIdentitySet)) || (err != nil && !errors.Is(err, identity.ErrNoIdentitySet)) {
		return nil, nil, err
	}

	if cfg.noBackend {
		return nil, repo, nil
	}

	// var events chan cache.BuildEvent
	backend, events := cache.NewRepoCache(repo)
	if err = cfg.EventConsumer()(ctx, events); err != nil {
		return nil, nil, err
	}

	// b.logger.DebugContext(
	// 	ctx,
	// 	"backend started",
	// 	slog.Int("identities", len(b.backend.Identities().AllIds())),
	// 	slog.Int("bugs", len(b.backend.Bugs().AllIds())),
	// )

	return backend, repo, nil
}

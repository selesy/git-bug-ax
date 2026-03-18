package backlog

import (
	"context"
	"log/slog"

	"github.com/git-bug/git-bug/cache"
	"github.com/git-bug/git-bug/entities/identity"
	"github.com/git-bug/git-bug/entity"
	"github.com/git-bug/git-bug/repository"
	"github.com/lmittmann/tint"
	"go.opentelemetry.io/otel/trace"

	"github.com/selesy/git-bug-agent/internal/app"
	"github.com/selesy/git-bug-agent/internal/gitbug"
	"github.com/selesy/git-bug-agent/internal/otel"
)

// type Interface interface {
// 	New(...Option) (*Issue, error)
// 	Update(...Option) (*Issue, error)
// 	Remove(entity.Id) error
// }

type Index struct {
	// Git-bug
	repo    repository.ClockedRepo
	backend *cache.RepoCache
	// Observability
	logger *slog.Logger
	tracer trace.Tracer
	span   trace.Span
}

func New(ctx context.Context, opts ...Option) (*Index, error) {
	var err error

	cfg, err := newConfig(ctx, opts...)
	if err != nil {
		return nil, err
	}

	ctx, gbaSpan := cfg.otel.Tracer().Start(ctx, app.Name)
	defer func() { otel.EndSpanOnError(gbaSpan, err) }()
	ctx, newSpan := cfg.otel.Tracer().Start(ctx, "new")
	defer otel.EndSpan(newSpan, err)

	var repo repository.ClockedRepo
	repo, err = repository.OpenGoGitRepo(cfg.gitbug.RepoPath(), gitbug.Namespace, []repository.ClockLoader{})
	if err != nil {
		return nil, err
	}

	b := &Index{
		repo:   repo,
		logger: cfg.otel.Logger(),
		tracer: cfg.otel.Tracer(),
		span:   gbaSpan,
	}

	if !cfg.gitbug.NoBackend() {
		var events chan cache.BuildEvent
		b.backend, events = cache.NewRepoCache(b.repo)

		if err = cfg.gitbug.EventConsumer()(ctx, events); err != nil {
			return nil, err
		}

		b.logger.DebugContext(
			ctx,
			"backend started",
			slog.Int("identities", len(b.backend.Identities().AllIds())),
			slog.Int("bugs", len(b.backend.Bugs().AllIds())),
		)
	}

	if !cfg.gitbug.EnsureUser() {
		return b, nil
	}

	if _, err = identity.GetUserIdentity(b.repo); err != nil {
		return nil, err
	}

	return b, nil
}

func (b *Index) Close() error {
	var err error

	switch {
	case b.backend != nil:
		err = b.backend.Close()
	case b.backend == nil && b.repo != nil:
		err = b.repo.Close()
	}

	if err != nil {
		b.logger.Error("backend closed", tint.Err(err))
	} else {
		b.logger.Debug("backend closed")
	}

	otel.EndSpan(b.span, err)
	b.backend = nil
	b.repo = nil

	return err
}

func (b *Index) Resolve(id entity.Id) (*Issue, error) {
	bug, err := b.backend.Bugs().Resolve(id)
	if err != nil {
		return nil, err
	}

	return Wrap(bug)
}

func (b *Index) ResolvePrefix(prefix string) (*Issue, error) {
	// TODO: check for backend

	bug, err := b.backend.Bugs().ResolvePrefix(prefix)
	if err != nil {
		return nil, err
	}

	return Wrap(bug)
}

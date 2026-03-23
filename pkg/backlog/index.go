package backlog

import (
	"context"
	"errors"
	"log/slog"

	"github.com/git-bug/git-bug/cache"
	"github.com/git-bug/git-bug/entities/identity"
	"github.com/git-bug/git-bug/repository"
	"github.com/lmittmann/tint"
	"go.opentelemetry.io/otel/trace"

	"github.com/selesy/git-bug-agent/internal/app"
	"github.com/selesy/git-bug-agent/internal/gitbug"
	"github.com/selesy/git-bug-agent/internal/otel"
	"github.com/selesy/git-bug-agent/pkg/issue"
)

type Index struct {
	repo    repository.ClockedRepo
	backend *cache.RepoCache
	logger  *slog.Logger
	tracer  trace.Tracer
	span    trace.Span
}

func New(ctx context.Context, opts ...IndexOption) (*Index, error) {
	var err error

	cfg, err := newConfig(ctx, opts...)
	if err != nil {
		return nil, err
	}

	ctx, gbaSpan := cfg.otel.Tracer().Start(ctx, app.Name)
	defer func() { otel.EndSpanOnError(gbaSpan, err) }()
	ctx, newSpan := cfg.otel.Tracer().Start(ctx, "new")
	defer otel.EndSpan(newSpan, err)

	backend, repo, err := gitbug.NewBackendAndRepo(ctx, cfg.gitbug)
	if err != nil {
		return nil, err
	}

	user, err := identity.GetUserIdentity(repo)
	if err != nil && !errors.Is(err, identity.ErrNoIdentitySet) {
		return nil, err
	}

	var human string
	if user != nil {
		human = user.Id().Human()
	}

	logger := cfg.otel.Logger().With(
		slog.String("trace_id", gbaSpan.SpanContext().TraceID().String()),
		slog.String("span_id", string(gbaSpan.SpanContext().SpanID().String())),
		slog.String("identity", human),
	)

	logger.DebugContext(
		ctx,
		"backend started",
		slog.Int("identities", len(backend.Identities().AllIds())),
		slog.Int("bugs", len(backend.Bugs().AllIds())),
	)

	return &Index{
		repo:    repo,
		backend: backend,
		logger:  logger,
		tracer:  cfg.otel.Tracer(),
		span:    gbaSpan,
	}, nil
}

func (i *Index) AllIDs() ([]issue.ID, error) {
	bugIDs := i.backend.Bugs().AllIds()
	issueIDs := make([]issue.ID, len(bugIDs))

	var errs error
	for i, bugID := range bugIDs {
		issueID, err := issue.ParseID(bugID.String())
		if err != nil {
			errs = errors.Join(errs, err)

			continue
		}

		issueIDs[i] = issueID
	}

	return issueIDs, errs
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

func (b *Index) Resolve(id issue.ID) (*Issue, error) {
	bug, err := b.backend.Bugs().Resolve(id.Id)
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

func (b *Index) SpanID() trace.SpanID {
	return b.span.SpanContext().SpanID()
}

func (b *Index) TraceID() trace.TraceID {
	return b.span.SpanContext().TraceID()
}

func (b *Index) UserIdentity() (*cache.IdentityCache, error) {
	return b.backend.GetUserIdentity()
}

package backlog

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/selesy/git-bug-agent/internal/gitbug"
	"github.com/selesy/git-bug-agent/internal/otel"
)

type config struct {
	gitbug *gitbug.Config
	otel   *otel.Config
}

func newConfig(ctx context.Context, opts ...IndexOption) (*config, error) {
	gitbugCfg, err := gitbug.NewConfig()
	if err != nil {
		return nil, err
	}

	cfg := config{
		gitbug: gitbugCfg,
		otel:   otel.NewConfig(),
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	if cfg.gitbug.NoBackend() && cfg.gitbug.EventConsumer() != nil {
		cfg.otel.Logger().WarnContext(ctx, "the WithEventConsumer option will have no effect with the WithNoBackend option")
	}

	if !cfg.gitbug.HasEventConsumer() {
		gitbug.WithEventConsumer(newObservedEventConsumer(cfg.otel))
	}

	return &cfg, nil
}

// IndexOption is a functional option for configuring the gba package.
type IndexOption func(*config)

// WithEnsureUser returns an Option that sets whether to ensure the user is configured.
func WithEnsureUser(ensureUser bool) IndexOption {
	return func(cfg *config) {
		gitbug.WithEnsureUser(ensureUser)(cfg.gitbug)
	}
}

// WithEventsConsumer returns an Option that sets the events consumer.
func WithEventsConsumer(eventConsumer gitbug.EventConsumer) IndexOption {
	return func(cfg *config) {
		gitbug.WithEventConsumer(eventConsumer)(cfg.gitbug)
	}
}

// WithNoBackend returns an Option that disables the backend.
func WithNoBackend(noBackend bool) IndexOption {
	return func(cfg *config) {
		gitbug.WithNoBackend(noBackend)(cfg.gitbug)
	}
}

// WithRepoPath returns an Option that sets the repository path.
func WithRepoPath(repoPath string) IndexOption {
	return func(cfg *config) {
		gitbug.WithRepoPath(repoPath)(cfg.gitbug)
	}
}

// WithLogger returns an Option that sets the logger.
func WithLogger(logger *slog.Logger) IndexOption {
	return func(cfg *config) {
		otel.WithLogger(logger)(cfg.otel)
	}
}

// WithMeter returns an Option that sets the meter.
func WithMeter(meter metric.Meter) IndexOption {
	return func(cfg *config) {
		otel.WithMeter(meter)(cfg.otel)
	}
}

// WithTracer returns an Option that sets the tracer.
func WithTracer(tracer trace.Tracer) IndexOption {
	return func(cfg *config) {
		otel.WithTracer(tracer)(cfg.otel)
	}
}

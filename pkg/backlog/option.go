package backlog

import (
	"context"
	"log/slog"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"

	"github.com/selesy/git-bug-ax/internal/gitbug"
	"github.com/selesy/git-bug-ax/internal/otel"
)

type config struct {
	gitbug *gitbug.Config
	otel   *otel.Config
}

func newConfig(ctx context.Context, opts ...Option) (*config, error) {
	// wd, err := os.Getwd()
	// if err != nil {
	// 	return nil, err
	// }

	gitbugCfg, err := gitbug.NewConfig()
	if err != nil {
		return nil, err
	}

	cfg := config{
		gitbug: gitbugCfg,
		otel:   otel.NewConfig(),
		// path:           wd,
		// ensureUser:     false,
		// noBackend:      false,
		// eventsConsumer: nil,
		// logger:         slog.New(slog.NewTextHandler(io.Discard, nil)),
		// meter:          metricnoop.NewMeterProvider().Meter("ax"),
		// tracer:         tracenoop.NewTracerProvider().Tracer("ax"),
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

// Option is a functional option for configuring the ax package.
type Option func(*config)

// WithEnsureUser returns an Option that sets whether to ensure the user is configured.
func WithEnsureUser(ensureUser bool) Option {
	return func(cfg *config) {
		gitbug.WithEnsureUser(ensureUser)(cfg.gitbug)
	}
}

// WithEventsConsumer returns an Option that sets the events consumer.
func WithEventsConsumer(eventConsumer gitbug.EventConsumer) Option {
	return func(cfg *config) {
		gitbug.WithEventConsumer(eventConsumer)(cfg.gitbug)
	}
}

// WithNoBackend returns an Option that disables the backend.
func WithNoBackend(noBackend bool) Option {
	return func(cfg *config) {
		gitbug.WithNoBackend(noBackend)(cfg.gitbug)
	}
}

// WithRepoPath returns an Option that sets the repository path.
func WithRepoPath(repoPath string) Option {
	return func(cfg *config) {
		gitbug.WithRepoPath(repoPath)(cfg.gitbug)
	}
}

// WithLogger returns an Option that sets the logger.
func WithLogger(logger *slog.Logger) Option {
	return func(cfg *config) {
		otel.WithLogger(logger)(cfg.otel)
	}
}

// WithMeter returns an Option that sets the meter.
func WithMeter(meter metric.Meter) Option {
	return func(cfg *config) {
		otel.WithMeter(meter)(cfg.otel)
	}
}

// WithTracer returns an Option that sets the tracer.
func WithTracer(tracer trace.Tracer) Option {
	return func(cfg *config) {
		otel.WithTracer(tracer)(cfg.otel)
	}
}

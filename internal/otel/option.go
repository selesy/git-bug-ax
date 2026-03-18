package otel

import (
	"io"
	"log/slog"

	"go.opentelemetry.io/otel/metric"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/trace"
	tracenoop "go.opentelemetry.io/otel/trace/noop"

	"github.com/selesy/git-bug-agent/internal/app"
)

type Config struct {
	logger   *slog.Logger
	meter    metric.Meter
	tracer   trace.Tracer
	observed bool
}

func NewConfig(opts ...Option) *Config {
	cfg := Config{
		logger: slog.New(slog.NewTextHandler(io.Discard, nil)),
		meter:  metricnoop.NewMeterProvider().Meter(app.Name),
		tracer: tracenoop.NewTracerProvider().Tracer(app.Name),
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	return &cfg
}

func (c *Config) Logger() *slog.Logger {
	return c.logger
}

func (c *Config) Meter() metric.Meter {
	return c.meter
}

func (c *Config) Tracer() trace.Tracer {
	return c.tracer
}

func (c *Config) IsObserved() bool {
	return c.observed
}

type Option func(*Config)

func WithLogger(logger *slog.Logger) Option {
	return func(cfg *Config) {
		cfg.logger = logger
		cfg.observed = true
	}
}

func WithMeter(meter metric.Meter) Option {
	return func(cfg *Config) {
		cfg.meter = meter
		cfg.observed = true
	}
}

func WithTracer(tracer trace.Tracer) Option {
	return func(cfg *Config) {
		cfg.tracer = tracer
		cfg.observed = true
	}
}

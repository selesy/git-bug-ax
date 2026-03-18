package otel_test

import (
	"io"
	"log/slog"
	"testing"

	metricnoop "go.opentelemetry.io/otel/metric/noop"
	tracenoop "go.opentelemetry.io/otel/trace/noop"

	"github.com/selesy/git-bug-agent/internal/otel"
)

func TestNewConfig_DefaultValues(t *testing.T) {
	cfg := otel.NewConfig()

	if cfg.Logger() == nil {
		t.Error("expected logger to be set")
	}
	if cfg.Meter() == nil {
		t.Error("expected meter to be set")
	}
	if cfg.Tracer() == nil {
		t.Error("expected tracer to be set")
	}
	if cfg.IsObserved() {
		t.Error("expected IsObserved to be false for default config")
	}
}

func TestNewConfig_WithLogger(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	cfg := otel.NewConfig(otel.WithLogger(logger))

	if cfg.Logger() != logger {
		t.Error("expected logger to be set to provided instance")
	}
	if !cfg.IsObserved() {
		t.Error("expected IsObserved to be true after setting logger")
	}
}

func TestNewConfig_WithMeter(t *testing.T) {
	meter := metricnoop.NewMeterProvider().Meter("test")
	cfg := otel.NewConfig(otel.WithMeter(meter))

	if cfg.Meter() != meter {
		t.Error("expected meter to be set to provided instance")
	}
	if !cfg.IsObserved() {
		t.Error("expected IsObserved to be true after setting meter")
	}
}

func TestNewConfig_WithTracer(t *testing.T) {
	tracer := tracenoop.NewTracerProvider().Tracer("test")
	cfg := otel.NewConfig(otel.WithTracer(tracer))

	if cfg.Tracer() != tracer {
		t.Error("expected tracer to be set to provided instance")
	}
	if !cfg.IsObserved() {
		t.Error("expected IsObserved to be true after setting tracer")
	}
}

func TestNewConfig_WithMultipleOptions(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	meter := metricnoop.NewMeterProvider().Meter("test")
	tracer := tracenoop.NewTracerProvider().Tracer("test")

	cfg := otel.NewConfig(
		otel.WithLogger(logger),
		otel.WithMeter(meter),
		otel.WithTracer(tracer),
	)

	if cfg.Logger() != logger {
		t.Error("expected logger to be set")
	}
	if cfg.Meter() != meter {
		t.Error("expected meter to be set")
	}
	if cfg.Tracer() != tracer {
		t.Error("expected tracer to be set")
	}
	if !cfg.IsObserved() {
		t.Error("expected IsObserved to be true")
	}
}

func TestNewConfig_OptionsOverridePreviousOptions(t *testing.T) {
	logger1 := slog.New(slog.NewTextHandler(io.Discard, nil))
	logger2 := slog.New(slog.NewTextHandler(io.Discard, nil))

	cfg := otel.NewConfig(
		otel.WithLogger(logger1),
		otel.WithLogger(logger2),
	)

	if cfg.Logger() != logger2 {
		t.Error("expected later option to override earlier option")
	}
}

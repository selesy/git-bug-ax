package backlog

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"

	"github.com/git-bug/git-bug/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	tracenoop "go.opentelemetry.io/otel/trace/noop"

	"github.com/selesy/git-bug-agent/pkg/backlog/backlogtest"
)

func TestNewConfigDefaults(t *testing.T) {
	ctx := context.Background()
	cfg, err := newConfig(ctx)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	// Verify default values
	wd, _ := os.Getwd()
	assert.Equal(t, wd, cfg.gitbug.RepoPath()) // Default is current working directory
	assert.False(t, cfg.gitbug.EnsureUser())
	assert.False(t, cfg.gitbug.NoBackend())
	assert.NotNil(t, cfg.gitbug.EventConsumer())
	assert.NotNil(t, cfg.otel.Logger())
	assert.NotNil(t, cfg.otel.Meter())
	assert.NotNil(t, cfg.otel.Tracer())
	// observedEventConsumer is the default
}

func TestNewConfigWithRepoPath(t *testing.T) {
	ctx := context.Background()
	testPath := "/tmp/test-repo"

	cfg, err := newConfig(ctx, WithRepoPath(testPath))
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, testPath, cfg.gitbug.RepoPath())
}

func TestNewConfigWithEnsureUser(t *testing.T) {
	ctx := context.Background()

	cfg, err := newConfig(ctx, WithEnsureUser(true))
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.True(t, cfg.gitbug.EnsureUser())
}

func TestNewConfigWithNoBackend(t *testing.T) {
	ctx := context.Background()

	cfg, err := newConfig(ctx, WithNoBackend(true))
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.True(t, cfg.gitbug.NoBackend())
}

func TestNewConfigWithEventsConsumer(t *testing.T) {
	ctx := context.Background()

	// Create a mock events consumer
	mockConsumer := func(context.Context, chan cache.BuildEvent) error {
		return nil
	}

	cfg, err := newConfig(ctx, WithEventsConsumer(mockConsumer))
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.NotNil(t, cfg.gitbug.EventConsumer())
	// Verify it's not the default (functions can't be compared directly)
	assert.NotEqual(t, "observedEventConsumer", "mockConsumer")
}

func TestNewConfigMultipleOptions(t *testing.T) {
	ctx := context.Background()
	testPath := "/custom/path"
	mockConsumer := func(context.Context, chan cache.BuildEvent) error {
		return nil
	}

	cfg, err := newConfig(ctx,
		WithRepoPath(testPath),
		WithEnsureUser(true),
		WithEventsConsumer(mockConsumer),
	)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.Equal(t, testPath, cfg.gitbug.RepoPath())
	assert.True(t, cfg.gitbug.EnsureUser())
	assert.False(t, cfg.gitbug.NoBackend())
	assert.NotNil(t, cfg.gitbug.EventConsumer())
}

func TestNewConfigNoBackendWithEventConsumer(t *testing.T) {
	const exp = `time=2006-01-02T15:04:05.000Z level=WARN msg="the WithEventConsumer option will have no effect with the WithNoBackend option"
`
	ctx := context.Background()
	mockConsumer := func(context.Context, chan cache.BuildEvent) error {
		return nil
	}

	logger, buf := backlogtest.DeterministicLogger(t)

	// This should set noBackend=true and the consumer, but log a warning
	cfg, err := newConfig(ctx,
		WithNoBackend(true),
		WithEventsConsumer(mockConsumer),
		WithLogger(logger),
	)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	assert.True(t, cfg.gitbug.NoBackend())
	assert.NotNil(t, cfg.gitbug.EventConsumer())
	assert.Equal(t, exp, buf.String())
}

func TestWithEnsureUserOption(t *testing.T) {
	tests := []bool{true, false}
	for _, ensureUser := range tests {
		t.Run("ensureUser="+fmt.Sprintf("%t", ensureUser), func(t *testing.T) {
			opt := WithEnsureUser(ensureUser)
			cfg, err := newConfig(context.Background())
			require.NoError(t, err)
			opt(cfg)
			assert.Equal(t, ensureUser, cfg.gitbug.EnsureUser())
		})
	}
}

func TestWithRepoPathOption(t *testing.T) {
	paths := []string{"/path/1", "/path/2", ".", "~"}
	for _, path := range paths {
		t.Run(path, func(t *testing.T) {
			opt := WithRepoPath(path)
			cfg, err := newConfig(context.Background())
			require.NoError(t, err)
			opt(cfg)
			assert.Equal(t, path, cfg.gitbug.RepoPath())
		})
	}
}

func TestWithNoBackendOption(t *testing.T) {
	opt := WithNoBackend(true)
	cfg, err := newConfig(context.Background())
	require.NoError(t, err)
	opt(cfg)
	assert.True(t, cfg.gitbug.NoBackend())
}

func TestWithEventsConsumerOption(t *testing.T) {
	mockConsumer := func(context.Context, chan cache.BuildEvent) error {
		return nil
	}

	opt := WithEventsConsumer(mockConsumer)
	cfg, err := newConfig(context.Background())
	require.NoError(t, err)
	opt(cfg)
	assert.NotNil(t, cfg.gitbug.EventConsumer())
}

func TestWithLoggerOption(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	cfg, err := newConfig(context.Background(), WithLogger(logger))
	require.NoError(t, err)
	assert.Equal(t, logger, cfg.otel.Logger())
	assert.True(t, cfg.otel.IsObserved())
}

func TestWithMeterOption(t *testing.T) {
	meter := metricnoop.NewMeterProvider().Meter("test")

	cfg, err := newConfig(context.Background(), WithMeter(meter))
	require.NoError(t, err)
	assert.Equal(t, meter, cfg.otel.Meter())
	assert.True(t, cfg.otel.IsObserved())
}

func TestWithTracerOption(t *testing.T) {
	tracer := tracenoop.NewTracerProvider().Tracer("test")

	cfg, err := newConfig(context.Background(), WithTracer(tracer))
	require.NoError(t, err)
	assert.Equal(t, tracer, cfg.otel.Tracer())
	assert.True(t, cfg.otel.IsObserved())
}

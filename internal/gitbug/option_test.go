package gitbug

import (
	"context"
	"errors"
	"testing"

	"github.com/git-bug/git-bug/cache"
)

func TestNewConfigErrorGettingWorkingDirectory(t *testing.T) {
	// This test is difficult to exercise naturally since os.Getwd() rarely fails.
	// It's documented here for future reference if test framework improvements allow it.
	// The error path is: os.Getwd() fails -> NewConfig returns nil, error
}

func TestNewConfigDefaults(t *testing.T) {
	cfg, err := NewConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.RepoPath() == "" {
		t.Error("repoPath should be set to current working directory")
	}
	if cfg.EnsureUser() {
		t.Error("ensureUser should default to false")
	}
	if cfg.NoBackend() {
		t.Error("noBackend should default to false")
	}
	if cfg.HasEventConsumer() {
		t.Error("eventConsumer should not be set by default")
	}
}

func TestNewConfigWithRepoPath(t *testing.T) {
	customPath := "/custom/repo/path"
	cfg, err := NewConfig(WithRepoPath(customPath))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.RepoPath() != customPath {
		t.Errorf("expected repoPath %q, got %q", customPath, cfg.RepoPath())
	}
}

func TestNewConfigWithEnsureUser(t *testing.T) {
	cfg, err := NewConfig(WithEnsureUser(true))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !cfg.EnsureUser() {
		t.Error("ensureUser should be true")
	}
}

func TestNewConfigWithNoBackend(t *testing.T) {
	cfg, err := NewConfig(WithNoBackend(true))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !cfg.NoBackend() {
		t.Error("noBackend should be true")
	}
}

func TestNewConfigWithEventConsumer(t *testing.T) {
	customConsumer := func(ctx context.Context, events chan cache.BuildEvent) error {
		return nil
	}

	cfg, err := NewConfig(WithEventConsumer(customConsumer))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !cfg.HasEventConsumer() {
		t.Error("eventConsumer should be set")
	}
}

func TestNewConfigMultipleOptions(t *testing.T) {
	customPath := "/custom/path"
	customConsumer := func(ctx context.Context, events chan cache.BuildEvent) error {
		return nil
	}

	cfg, err := NewConfig(
		WithRepoPath(customPath),
		WithEnsureUser(true),
		WithNoBackend(true),
		WithEventConsumer(customConsumer),
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.RepoPath() != customPath {
		t.Errorf("expected repoPath %q, got %q", customPath, cfg.RepoPath())
	}
	if !cfg.EnsureUser() {
		t.Error("ensureUser should be true")
	}
	if !cfg.NoBackend() {
		t.Error("noBackend should be true")
	}
	if !cfg.HasEventConsumer() {
		t.Error("eventConsumer should be set")
	}
}

func TestEventConsumerReturnsDefaultWhenNotSet(t *testing.T) {
	cfg, err := NewConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	consumer := cfg.EventConsumer()
	if consumer == nil {
		t.Error("EventConsumer() should never return nil")
	}

	// Verify it's the noop consumer by testing its behavior
	events := make(chan cache.BuildEvent)
	close(events)
	err = consumer(context.Background(), events)
	if err != nil {
		t.Errorf("noop consumer should not return error, got %v", err)
	}
}

func TestEventConsumerReturnsCustomWhenSet(t *testing.T) {
	expectedErr := errors.New("custom error")
	customConsumer := func(ctx context.Context, events chan cache.BuildEvent) error {
		return expectedErr
	}

	cfg, err := NewConfig(WithEventConsumer(customConsumer))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	consumer := cfg.EventConsumer()
	events := make(chan cache.BuildEvent)
	close(events)
	err = consumer(context.Background(), events)
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
}

func TestNoopEventConsumerClosesChannelGracefully(t *testing.T) {
	ctx := context.Background()
	events := make(chan cache.BuildEvent)

	go func() {
		close(events)
	}()

	err := noopEventConsumer(ctx, events)
	if err != nil {
		t.Errorf("noop consumer should not return error, got %v", err)
	}
}

func TestNoopEventConsumerStopsOnError(t *testing.T) {
	ctx := context.Background()
	events := make(chan cache.BuildEvent)
	expectedErr := errors.New("test error")

	go func() {
		events <- cache.BuildEvent{Err: expectedErr}
		events <- cache.BuildEvent{Err: nil} // This shouldn't be processed
		close(events)
	}()

	err := noopEventConsumer(ctx, events)
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
}

func TestWithRepoPathOption(t *testing.T) {
	tests := []string{
		"/path/1",
		"/path/2",
		".",
		"~",
	}

	for _, path := range tests {
		t.Run(path, func(t *testing.T) {
			cfg, err := NewConfig(WithRepoPath(path))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if cfg.RepoPath() != path {
				t.Errorf("expected repoPath %q, got %q", path, cfg.RepoPath())
			}
		})
	}
}

func TestWithEnsureUserOption(t *testing.T) {
	tests := []bool{true, false}

	for _, val := range tests {
		t.Run("ensureUser="+func() string {
			if val {
				return "1"
			}
			return "0"
		}(), func(t *testing.T) {
			cfg, err := NewConfig(WithEnsureUser(val))
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if cfg.EnsureUser() != val {
				t.Errorf("expected ensureUser %v, got %v", val, cfg.EnsureUser())
			}
		})
	}
}

func TestWithNoBackendOption(t *testing.T) {
	cfg, err := NewConfig(WithNoBackend(true))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !cfg.NoBackend() {
		t.Error("noBackend should be true")
	}
}

func TestWithEventsConsumerOption(t *testing.T) {
	customConsumer := func(ctx context.Context, events chan cache.BuildEvent) error {
		return nil
	}

	cfg, err := NewConfig(WithEventConsumer(customConsumer))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !cfg.HasEventConsumer() {
		t.Error("eventConsumer should be set")
	}
}

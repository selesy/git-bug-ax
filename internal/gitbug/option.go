// Package gitbug provides configuration and utilities for interacting with git-bug.
package gitbug

import "os"

// Config holds the configuration for git-bug operations.
type Config struct {
	repoPath      string
	ensureUser    bool
	noBackend     bool
	eventConsumer EventConsumer
}

// NewConfig creates a new Config with the given options.
// It defaults to the current working directory as the repository path.
func NewConfig(opts ...Option) (*Config, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	cfg := Config{
		repoPath:   wd,
		ensureUser: false,
		noBackend:  false,
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	return &cfg, nil
}

// RepoPath returns the repository path.
func (c *Config) RepoPath() string {
	return c.repoPath
}

// EnsureUser returns whether to ensure the git user is configured.
func (c *Config) EnsureUser() bool {
	return c.ensureUser
}

// NoBackend returns whether to skip backend initialization.
func (c *Config) NoBackend() bool {
	return c.noBackend
}

// EventConsumer returns the event consumer.
// If no custom consumer was set, it returns the noop consumer.
func (c *Config) EventConsumer() EventConsumer {
	if c.eventConsumer != nil {
		return c.eventConsumer
	}
	return noopEventConsumer
}

// HasEventConsumer returns whether a custom event consumer was set.
func (c *Config) HasEventConsumer() bool {
	return c.eventConsumer != nil
}

// Option is a function that modifies a Config.
type Option func(*Config)

// WithRepoPath returns an Option that sets the repository path.
func WithRepoPath(repoPath string) Option {
	return func(cfg *Config) {
		cfg.repoPath = repoPath
	}
}

// WithEnsureUser returns an Option that sets whether to ensure the git user is configured.
func WithEnsureUser(ensureUser bool) Option {
	return func(cfg *Config) {
		cfg.ensureUser = ensureUser
	}
}

// WithNoBackend returns an Option that sets whether to skip backend initialization.
func WithNoBackend(noBackend bool) Option {
	return func(cfg *Config) {
		cfg.noBackend = noBackend
	}
}

// WithEventConsumer returns an Option that sets a custom event consumer.
func WithEventConsumer(eventConsumer EventConsumer) Option {
	return func(cfg *Config) {
		cfg.eventConsumer = eventConsumer
	}
}

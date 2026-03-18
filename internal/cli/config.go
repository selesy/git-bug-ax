// Package cli provides the command-line interface for git-bug-agent.
package cli

import (
	"crypto/sha256"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lmittmann/tint"
	"github.com/spf13/cobra"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/selesy/git-bug-agent/internal/app"
	"github.com/selesy/git-bug-agent/pkg/backlog"
)

const logFileName = app.Name + ".log"

// config holds the global CLI configuration state including logging and index.
type config struct {
	index        *backlog.Index
	gitDir       string
	human        bool
	logger       *slog.Logger
	logLevel     string
	logCloseFunc func() error
	logDir       string
	logFormat    string
	pretty       bool
	session      string
}

// Initialize sets up application state that must happen before any CLI command executes,
// including logging, observability, and session tracking.
func (cfg *config) Initialize() func(*cobra.Command, []string) error {
	return func(_ *cobra.Command, _ []string) error {
		var level slog.Level
		if err := level.UnmarshalText([]byte(cfg.logLevel)); err != nil {
			return err
		}

		rotator := &lumberjack.Logger{
			Filename:   filepath.Join(cfg.logDir, logFileName),
			MaxSize:    10,
			MaxBackups: 3,
			MaxAge:     28,
			Compress:   true,
		}

		cfg.logCloseFunc = rotator.Close

		var (
			handler slog.Handler
			warnMsg string
		)

		switch format := strings.ToLower(cfg.logFormat); format {
		case "json":
			handler = slog.NewJSONHandler(rotator, &slog.HandlerOptions{
				Level: level,
			})
		case "text":
			handler = slog.NewTextHandler(rotator, &slog.HandlerOptions{
				Level: level,
			})
		default:
			if format != "colorized" {
				warnMsg = "unknown --log-format specified"
			}

			handler = tint.NewHandler(rotator, &tint.Options{
				Level: level,
			})
		}

		// Create the slog handler using the rotator as the destination
		cfg.logger = slog.New(handler)

		// Add the session attribute
		pid := os.Getpid()
		now := time.Now().UnixNano()
		cfg.session = fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%d-%d", pid, now))))[:7]
		cfg.logger = cfg.logger.With(slog.String("session", cfg.session))

		// Set as default so you can use slog.Info() globally
		slog.SetDefault(cfg.logger)

		slog.Debug(app.Name + " started")
		if warnMsg != "" {
			slog.Warn(warnMsg, slog.String("value", cfg.logFormat))
		}

		return nil
	}
}

// LoadBacklog returns a Cobra PersistentPreRunE function that loads the index
// from the git repository. The index is required by most CLI commands.
func (c *config) LoadBacklog() func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		opts := []backlog.Option{backlog.WithLogger(c.logger)}
		if c.gitDir != "" {
			opts = append(opts, backlog.WithRepoPath(c.gitDir))
		}

		var err error
		c.index, err = backlog.New(cmd.Context(), opts...)

		return err
	}
}

// LoadBacklogEnsureUser returns a Cobra PersistentPreRunE function that loads the index
// and ensures a git user is configured. Used by commands that create new bug entries.
func (c *config) LoadBacklogEnsureUser() func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, _ []string) error {
		opts := []backlog.Option{
			backlog.WithLogger(c.logger),
			backlog.WithEnsureUser(true),
		}
		if c.gitDir != "" {
			opts = append(opts, backlog.WithRepoPath(c.gitDir))
		}

		var err error
		c.index, err = backlog.New(cmd.Context(), opts...)

		return err
	}
}

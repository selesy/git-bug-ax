package cli

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/git-bug/git-bug/entities/identity"
	"github.com/lmittmann/tint"
	"github.com/spf13/cobra"

	"github.com/selesy/git-bug-agent/internal/app"
	"github.com/selesy/git-bug-agent/pkg/backlog"
)

// Execute initializes and executes the CLI application, handling startup and cleanup.
func Execute() (err error) {
	ctx := context.Background() // TODO: this should be cancellable and tied to signals

	var cfg config
	defer func() {
		logResult(ctx, cfg, err)

		if err := closeIndex(cfg.index); err != nil && cfg.logger != nil {
			cfg.logger.ErrorContext(ctx, "failed to close index", tint.Err(err))
		}

		if cfg.logger != nil {
			cfg.logger.DebugContext(ctx, app.Name+" stopped")
		}

		if cfg.logCloseFunc == nil {
			return
		}

		if err := cfg.logCloseFunc(); err != nil && cfg.logger != nil {
			cfg.logger.ErrorContext(ctx, "lumberjack failed to close log file(s)")
		}
	}()

	var cmd *cobra.Command
	cmd, err = RootCmd(&cfg)
	if err != nil {
		return err
	}

	err = cmd.ExecuteContext(ctx)

	return // err
}

// RootCmd returns the root Cobra command for the git-bug-agent CLI application.
// It configures all persistent flags and registers subcommands.
func RootCmd(cfg *config) (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:               app.Name,
		Short:             "Git-Bug's Agent Interface",
		Aliases:           []string{"list", "ls"},
		PersistentPreRunE: cfg.Initialize(),
		Run:               func(_ *cobra.Command, _ []string) {},
	}

	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}

	// Repository configuration
	cmd.PersistentFlags().StringVar(&cfg.gitDir, "git-dir", wd, "path to a Git repository")

	// Observability configuration
	cmd.PersistentFlags().StringVar(&cfg.logLevel, "log-level", "INFO", "log level must be one of ERROR, WARN, INFO or DEBUG")
	cmd.PersistentFlags().StringVar(&cfg.logDir, "log-dir", filepath.Join(cacheDir, app.Name, "logs"), "path where logs will be written")
	cmd.PersistentFlags().StringVar(&cfg.logFormat, "log-format", "colorized", "one of colorized, json or text")

	// Output formatting
	cmd.PersistentFlags().BoolVar(&cfg.human, "human", false, "display output in human-readable form instead of JSON")
	cmd.PersistentFlags().BoolVar(&cfg.pretty, "pretty", false, "pretty-print JSON output with indentation and new-lines")

	cmd.AddCommand(ShowCmd(cfg))

	return cmd, err
}

func closeIndex(index *backlog.Index) error {
	if index == nil {
		return nil
	}

	return index.Close()
}

// In normal operation, there is one "wide" log entry written for each CLI
// call as described at https://loggingsucks.com/.  This is clearly far
// more valuable with structured logs but still nicely augments the records
// written to the local log file.
func logResult(ctx context.Context, cfg config, err error) {
	// The logger is not loaded if a help command was run
	if cfg.logger == nil {
		return
	}

	// Sanitize the command line to avoid log injection (gosec G706)
	var command []string
	for _, tkn := range os.Args {
		tkn = strings.ReplaceAll(tkn, "\n", "")
		tkn = strings.ReplaceAll(tkn, "\r", "")
		command = append(command, strings.TrimSpace(tkn))
	}

	level := slog.LevelInfo
	attrs := []any{
		slog.String("command", strings.Join(command, " ")),
	}

	if err != nil {
		level = slog.LevelError
		attrs = append(attrs, tint.Err(err))
	}

	// If the index isn't loaded, the identity, trace ID and span ID attributes
	// make no sense.
	if cfg.index == nil {
		cfg.logger.Log(ctx, level, app.Name+" CLI executed", attrs...)

		return
	}

	userIdentity, idErr := cfg.index.UserIdentity()
	if idErr != nil && !errors.Is(err, identity.ErrNoIdentitySet) {
		cfg.logger.ErrorContext(ctx, "failed to retrieve user identity", tint.Err(err))
	}

	var userHuman string
	if idErr == nil {
		userHuman = userIdentity.Id().Human()
	}

	attrs = append(attrs, []any{
		slog.String("identity", userHuman),
		slog.String("trace_id", cfg.index.TraceID().String()),
		slog.String("span_id", cfg.index.SpanID().String()),
	}...)

	// if identity := userIdentity(cfg.index); identity != "" {
	// 	attrs = append(attrs, slog.String("identity", identity))
	// }

	cfg.logger.Log(ctx, level, app.Name+" CLI executed", attrs...)
}

// func userIdentity(index *backlog.Index) string {
// 	if index == nil {
// 		return ""
// 	}

// 	identity, err := index.GetUserIdentity()
// 	if err != nil {
// 		return "" // TODO: report warning?
// 	}

// 	return identity.Id().Human()
// }

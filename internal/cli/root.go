package cli

import (
	"errors"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// Execute initializes and executes the CLI application, handling startup and cleanup.
func Execute() (err error) {
	var cfg config

	defer func() {

		if cfg.index != nil {
			err = errors.Join(err, cfg.index.Close())
		}

		if err != nil {
			slog.Error(err.Error())
		}

		slog.Debug("ax stopped")

		if cfg.logCloseFunc == nil {
			return
		}

		if err := cfg.logCloseFunc(); err != nil {
			slog.Error("lumberjack failed to close log file(s)")
		}
	}()

	var cmd *cobra.Command
	cmd, err = RootCmd(&cfg)
	if err != nil {
		return err
	}

	err = cmd.Execute()

	return // err
}

// RootCmd returns the root Cobra command for the git-bug-ax CLI application.
// It configures all persistent flags and registers subcommands.
func RootCmd(cfg *config) (*cobra.Command, error) {

	cmd := &cobra.Command{
		Use:               "gbax",
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
	cmd.PersistentFlags().StringVar(&cfg.logDir, "log-dir", filepath.Join(cacheDir, "ax", "logs"), "path where logs will be written")
	cmd.PersistentFlags().StringVar(&cfg.logFormat, "log-format", "colorized", "one of colorized, json or text")

	// Output formatting
	cmd.PersistentFlags().BoolVar(&cfg.human, "human", false, "display output in human-readable form instead of JSON")
	cmd.PersistentFlags().BoolVar(&cfg.pretty, "pretty", false, "pretty-print JSON output with indentation and new-lines")

	cmd.AddCommand(ShowCmd(cfg))

	return cmd, err
}

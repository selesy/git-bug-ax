package cli

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/git-bug/git-bug/commands/execenv"
	"github.com/lmittmann/tint"
	"github.com/spf13/cobra"
	"gopkg.in/natefinch/lumberjack.v2"
)

func Execute() (err error) {
	var cfg config
	cfg.env = execenv.NewEnv()

	defer func() {
		if cfg.env.Backend != nil {
			slog.Debug("git-bug backend stopped")
			err = errors.Join(err, cfg.env.Backend.Close())
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

func RootCmd(cfg *config) (*cobra.Command, error) {

	cmd := &cobra.Command{
		Use:               "gbax",
		Short:             "Git-Bug's Agent Interface",
		PersistentPreRunE: setupConfig(cfg),
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

	cmd.PersistentFlags().StringVar(&cfg.gitDir, "git-dir", wd, "path to a Git repository")

	// Logging configuration
	cmd.PersistentFlags().StringVar(&cfg.logLevel, "log-level", "INFO", "log level must be one of ERROR, WARN, INFO or DEBUG")
	cmd.PersistentFlags().StringVar(&cfg.logDir, "log-dir", filepath.Join(cacheDir, "ax", "logs"), "path where logs will be written")
	cmd.PersistentFlags().StringVar(&cfg.logFormat, "log-format", "colorized", "one of colorized, json or text")

	// Output formatting
	cmd.PersistentFlags().BoolVar(&cfg.human, "human", false, "display output in human-readable form instead of JSON")
	cmd.PersistentFlags().BoolVar(&cfg.pretty, "pretty", false, "pretty-print JSON output with indentation and new-lines")
	cmd.MarkFlagsMutuallyExclusive("human", "pretty")

	_ = cfg

	// cmd.AddCommand(ShowCmd(cfg))
	// cmd.AddCommand(StatusCmd())
	// cmd.AddCommand(UpdateCmd(cfg.env))

	return cmd, err
}

type config struct {
	env          *execenv.Env
	gitDir       string
	human        bool
	log          *slog.Logger
	logLevel     string
	logCloseFunc func() error
	logDir       string
	logFormat    string
	pretty       bool
	session      string
}

func setupConfig(cfg *config) func(*cobra.Command, []string) error {
	return func(_ *cobra.Command, _ []string) error {
		var level slog.Level
		if err := level.UnmarshalText([]byte(cfg.logLevel)); err != nil {
			return err
		}

		rotator := &lumberjack.Logger{
			Filename:   filepath.Join(cfg.logDir, "ax.log"),
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
			if format != "colorize" {
				warnMsg = "unknown --log-format specified"
			}

			handler = tint.NewHandler(rotator, &tint.Options{
				Level: level,
			})
		}

		// Create the slog handler using the rotator as the destination
		cfg.log = slog.New(handler)

		// Add the session attribute

		pid := os.Getpid()
		now := time.Now().UnixNano()
		cfg.session = fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("%d-%d", pid, now))))[:7]
		cfg.log = cfg.log.With(slog.String("session", cfg.session))

		// Set as default so you can use slog.Info() globally
		slog.SetDefault(cfg.log)

		slog.Debug("ax started")
		if warnMsg != "" {
			slog.Warn(warnMsg, slog.String("value", cfg.logFormat))
		}

		return nil
	}
}

package cli

import (
	"os"
	"path/filepath"

	"github.com/git-bug/git-bug/commands/execenv"
	"github.com/spf13/cobra"
)

func RootCmd() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "gbax",
		Short: "Git-Bug's Agent Interface",
		Run:   func(_ *cobra.Command, _ []string) {},
	}

	wd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return nil, err
	}

	cmd.PersistentFlags().String("repo-dir", wd, "path to a Git repository")
	cmd.PersistentFlags().Bool("human", false, "display output in human-readable form instead of JSON")
	cmd.PersistentFlags().String("log-level", "INFO", "log level must be one of ERROR, WARN, INFO or DEBUG")
	cmd.PersistentFlags().String("log-dir", filepath.Join(cacheDir, "gbax", "logs"), "path where logs will be written")

	env := execenv.NewEnv()

	_ = env // TODO

	// cmd.AddCommand(ShowCmd(env))
	// cmd.AddCommand(StatusCmd())
	// cmd.AddCommand(UpdateCmd(env))

	return cmd, err
}

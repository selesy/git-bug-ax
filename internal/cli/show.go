package cli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

const showLong = `Display detailed information about a specific issue.

The ID can be a full identifier or a prefix. The issue details will be
formatted according to the output format flags (--human, --pretty, or --json).`

const showExample = `  # Show an issue by prefix
  git-bug-ax show abc123

  # Show an issue by full hash
  git-bug-ax show abc123ef45670890abcdef1234567890abcdef

  # Show an issue with human-readable output
  git-bug-ax show abc123 --human

  # Show an issue with CRDT history
  git-bug-ax show abc123 --history

  # Show an issue with pretty-printed JSON
  git-bug-ax show abc123 --pretty`

// ShowCmd returns a Cobra command that displays detailed information about an issue.
// The issue can be identified by full ID or prefix.
func ShowCmd(cfg *config) *cobra.Command {
	opts := &showOptions{}

	cmd := &cobra.Command{
		Use:     "show <id>",
		Short:   "Display detailed information about an issue",
		Long:    showLong,
		Example: showExample,
		PreRunE: cfg.LoadBacklog(),
		RunE:    showCmdFunc(cfg, opts),
	}

	cmd.Flags().BoolVar(&opts.history, "history", false, "include CRDT operations in output")

	return cmd
}

type showOptions struct {
	history bool
}

func showCmdFunc(cfg *config, opts *showOptions) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return ErrWrongPrefixCount
		}

		issue, err := cfg.backlog.ResolvePrefix(args[0])
		if err != nil {
			return err
		}

		var v json.Marshaler
		if opts.history {
			v = issue
		} else {
			v = issue.Excerpt()
		}

		displayFn := json.Marshal
		if cfg.pretty {
			displayFn = func(v any) ([]byte, error) {
				return json.MarshalIndent(v, "", "  ")
			}
		}

		data, err := displayFn(v)
		if err != nil {
			return err
		}

		_, err = fmt.Println(string(data))

		return err
	}
}

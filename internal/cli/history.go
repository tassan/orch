package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tassan/orch/internal/history"
)

const timeLayout = "2006-01-02 15:04"

func historyCmd() *cobra.Command {
	var limit int
	var asJSON bool
	var files []string
	cmd := &cobra.Command{
		Use:   "history <command> [args...]",
		Short: "Show past calls of a command from your shell history",
		Long: "Search your shell history (zsh, bash, fish) for calls of a command, oldest first.\n" +
			"Give leading arguments to narrow it down: `orch history git commit` matches calls that\n" +
			"start with `git commit`. The command is matched anywhere a command can start: after\n" +
			"env assignments, sudo/time/env, in pipelines and after && or ;. Flags for orch go before the command.\n\n" +
			"History files are read from $HISTFILE, ~/.zsh_history, ~/.bash_history and fish's\n" +
			"fish_history, or from --file. Shells often write history only on exit, so the current\n" +
			"session may be missing. History can contain secrets typed on the command line.",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if limit < 0 {
				return fmt.Errorf("--limit must be 0 (all) or more")
			}
			paths := files
			if len(paths) == 0 {
				home, _ := os.UserHomeDir()
				paths = history.Sources(home, os.Getenv("HISTFILE"), os.Getenv("XDG_DATA_HOME"))
			}
			if len(paths) == 0 {
				return fmt.Errorf("no shell history files found; pass one with --file")
			}
			all, err := history.Load(paths)
			if err != nil {
				return err
			}
			calls := history.Filter(all, args[0], args[1:])
			label := strings.Join(args, " ")
			if len(calls) == 0 {
				return fmt.Errorf("no calls of %q found in %s", label, names(paths))
			}
			shown := calls
			if limit > 0 && len(shown) > limit {
				shown = shown[len(shown)-limit:]
			}
			out := cmd.OutOrStdout()
			if asJSON {
				enc := json.NewEncoder(out)
				enc.SetIndent("", "  ")
				if err := enc.Encode(shown); err != nil {
					return err
				}
			} else {
				pad := strings.Repeat(" ", len(timeLayout)+2)
				for _, e := range shown {
					when := strings.Repeat(" ", len(timeLayout))
					if e.Time != nil {
						when = e.Time.Local().Format(timeLayout)
					}
					fmt.Fprintf(out, "%s  %s\n", when, strings.ReplaceAll(e.Line, "\n", "\n"+pad))
				}
			}
			fmt.Fprintf(cmd.ErrOrStderr(), "showing %d of %d calls of %q from %s\n", len(shown), len(calls), label, names(paths))
			return nil
		},
	}
	// Everything after the command belongs to the command, not to orch.
	cmd.Flags().SetInterspersed(false)
	cmd.Flags().IntVarP(&limit, "limit", "n", 20, "show only the last N calls (0 = all)")
	cmd.Flags().BoolVar(&asJSON, "json", false, "JSON output")
	cmd.Flags().StringArrayVar(&files, "file", nil, "history file to read instead of the defaults (repeatable)")
	return cmd
}

func names(paths []string) string {
	var b []string
	for _, p := range paths {
		b = append(b, filepath.Base(p))
	}
	return strings.Join(b, ", ")
}

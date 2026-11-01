package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tassan/orch/internal/provider"
	"github.com/tassan/orch/internal/scan"
)

// selfPath returns the running binary's path; a var so tests can inject one.
var selfPath = os.Executable

func updateCmd() *cobra.Command {
	var yes, dry bool
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update orch through the package manager that installed it",
		Long: "Update finds the package manager that owns this orch binary and runs its upgrade command\n" +
			"(`brew upgrade orch` for a Homebrew install). orch never replaces its own binary. If no\n" +
			"upgrade-capable manager owns it, orch prints the path and does nothing.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			exe, err := selfPath()
			if err != nil {
				return fmt.Errorf("cannot locate orch binary: %w", err)
			}
			real, err := filepath.EvalSymlinks(exe)
			if err != nil {
				real = exe
			}
			p, pkg, ok := provider.Owner(providers, scan.Tool{Name: "orch", Path: exe, Real: real})
			if !ok {
				return fmt.Errorf("orch at %s is not owned by a known package manager; update it the way you installed it", real)
			}
			up, ok := p.(provider.Upgrader)
			if !ok {
				return fmt.Errorf("%s cannot upgrade packages; update orch with it manually (%s)", p.Name(), pkg)
			}
			fmt.Fprintf(out, "current:  %s\n", version())
			fmt.Fprintf(out, "package:  %s\n", pkg)
			c := up.UpgradeCmd(pkg)
			fmt.Fprintf(out, "would run: %s\n", provider.Display(c))
			if dry {
				return nil
			}
			if !yes {
				fmt.Fprint(out, "proceed? [y/N] ")
				ans, _ := bufio.NewReader(cmd.InOrStdin()).ReadString('\n')
				if a := strings.ToLower(strings.TrimSpace(ans)); a != "y" && a != "yes" {
					return fmt.Errorf("aborted")
				}
			}
			c.Stdout, c.Stderr = out, cmd.ErrOrStderr()
			return c.Run()
		},
	}
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "skip confirmation")
	cmd.Flags().BoolVar(&dry, "dry-run", false, "print command, do not run")
	return cmd
}

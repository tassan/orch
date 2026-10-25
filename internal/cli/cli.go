// Package cli wires the cobra commands.
package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/tassan/orch/internal/classify"
	"github.com/tassan/orch/internal/info"
	"github.com/tassan/orch/internal/provider"
	"github.com/tassan/orch/internal/scan"
)

type entry struct {
	scan.Tool
	Kind   classify.Kind `json:"kind"`
	Source string        `json:"source,omitempty"`
}

func inventory() []entry {
	home, _ := os.UserHomeDir()
	var out []entry
	for _, t := range scan.Scan(scan.PathDirs(os.Getenv("PATH"))) {
		e := entry{Tool: t, Kind: classify.Classify(t.Real)}
		if e.Kind == classify.User {
			e.Source = classify.Source(t.Real, home)
		}
		out = append(out, e)
	}
	return out
}

func find(name string) (entry, bool) {
	for _, e := range inventory() {
		if e.Name == name {
			return e, true
		}
	}
	return entry{}, false
}

func Root() *cobra.Command {
	root := &cobra.Command{Use: "orch", Short: "Inspect and uninstall CLI tools", SilenceUsage: true}
	root.AddCommand(listCmd(), infoCmd(), uninstallCmd())
	return root
}

func listCmd() *cobra.Command {
	var native, user, asJSON bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List CLI tools found on PATH",
		RunE: func(cmd *cobra.Command, _ []string) error {
			var es []entry
			for _, e := range inventory() {
				if (native && e.Kind != classify.Native) || (user && e.Kind != classify.User) {
					continue
				}
				es = append(es, e)
			}
			if asJSON {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(es)
			}
			w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 4, 2, ' ', 0)
			fmt.Fprintln(w, "NAME\tKIND\tSOURCE\tPATH")
			for _, e := range es {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", e.Name, e.Kind, e.Source, e.Path)
			}
			return w.Flush()
		},
	}
	cmd.Flags().BoolVar(&native, "native", false, "only OS-shipped tools")
	cmd.Flags().BoolVar(&user, "user", false, "only user-installed tools")
	cmd.Flags().BoolVar(&asJSON, "json", false, "JSON output")
	return cmd
}

func infoCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "info <name>",
		Short: "Show details about a tool",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			e, ok := find(args[0])
			if !ok {
				return fmt.Errorf("%s not found on PATH", args[0])
			}
			w := cmd.OutOrStdout()
			fmt.Fprintf(w, "name:     %s\npath:     %s\nreal:     %s\nkind:     %s\n", e.Name, e.Path, e.Real, e.Kind)
			if e.Source != "" {
				fmt.Fprintf(w, "source:   %s\n", e.Source)
			}
			if p := provider.Find(provider.All(), e.Tool); p != nil {
				fmt.Fprintf(w, "owner:    %s\n", p.Name())
			}
			if v := info.Version(e.Path); v != "" {
				fmt.Fprintf(w, "version:  %s\n", v)
			}
			if d := info.Whatis(e.Name); d != "" {
				fmt.Fprintf(w, "about:    %s\n", d)
			}
			return nil
		},
	}
}

func uninstallCmd() *cobra.Command {
	var yes, dry bool
	cmd := &cobra.Command{
		Use:   "uninstall <name>",
		Short: "Uninstall a tool via the package manager that owns it",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			e, ok := find(args[0])
			if !ok {
				return fmt.Errorf("%s not found on PATH", args[0])
			}
			if e.Kind == classify.Native {
				return fmt.Errorf("%s is a native OS tool; refusing to uninstall", e.Name)
			}
			p := provider.Find(provider.All(), e.Tool)
			if p == nil {
				return fmt.Errorf("no package manager owns %s; remove manually: %s", e.Name, e.Real)
			}
			c, err := p.UninstallCmd(e.Tool)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "would run: %s\n", provider.Display(c))
			if dry {
				return nil
			}
			if !yes {
				fmt.Fprint(cmd.OutOrStdout(), "proceed? [y/N] ")
				ans, _ := bufio.NewReader(cmd.InOrStdin()).ReadString('\n')
				if a := strings.ToLower(strings.TrimSpace(ans)); a != "y" && a != "yes" {
					return fmt.Errorf("aborted")
				}
			}
			c.Stdout, c.Stderr = cmd.OutOrStdout(), cmd.ErrOrStderr()
			return c.Run()
		},
	}
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "skip confirmation")
	cmd.Flags().BoolVar(&dry, "dry-run", false, "print command, do not run")
	return cmd
}

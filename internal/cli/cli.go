// Package cli wires the cobra commands.
package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"runtime/debug"
	"sort"
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
	Kind    classify.Kind     `json:"kind"`
	Source  string            `json:"source,omitempty"`
	Package *provider.Package `json:"package,omitempty"`
}

// providers is a var so tests can inject fakes.
var providers = provider.All()

// inventory scans PATH; with all it also scans well-known tool dirs that are not on PATH.
func inventory(all bool) []entry {
	home, _ := os.UserHomeDir()
	dirs := scan.PathDirs(os.Getenv("PATH"))
	tools := scan.Scan(dirs)
	if all && home != "" {
		onPath := map[string]bool{}
		for _, t := range tools {
			onPath[t.Name] = true
		}
		for _, t := range scan.Scan(scan.ExtraDirs(home, dirs)) {
			if !onPath[t.Name] {
				t.OffPath = true
				tools = append(tools, t)
			}
		}
		sort.Slice(tools, func(i, j int) bool { return tools[i].Name < tools[j].Name })
	}
	var out []entry
	for _, t := range tools {
		e := entry{Tool: t, Kind: classify.Classify(t.Real)}
		if e.Kind == classify.User {
			e.Source = classify.Source(t.Real, home)
			// A package manager that owns the tool is a better source than a path guess.
			if p, pkg, ok := provider.Owner(providers, t); ok {
				e.Source, e.Package = p.Name(), &pkg
			}
		}
		out = append(out, e)
	}
	return out
}

func find(name string) (entry, bool) {
	for _, e := range inventory(true) {
		if e.Name == name {
			return e, true
		}
	}
	return entry{}, false
}

// Version is set at release time with -ldflags "-X github.com/tassan/orch/internal/cli.Version=...".
var (
	Version = ""
	Commit  = ""
	Date    = ""
)

func version() string {
	if Version != "" {
		if Commit != "" && Date != "" {
			return fmt.Sprintf("%s (%s, %s)", Version, Commit, Date)
		}
		return Version
	}
	if bi, ok := debug.ReadBuildInfo(); ok && bi.Main.Version != "" && bi.Main.Version != "(devel)" {
		return bi.Main.Version
	}
	return "dev"
}

func Root() *cobra.Command {
	root := &cobra.Command{Use: "orch", Short: "Inspect and uninstall CLI tools", SilenceUsage: true, Version: version()}
	root.AddCommand(listCmd(), infoCmd(), uninstallCmd(), historyCmd())
	return root
}

func listCmd() *cobra.Command {
	var native, user, asJSON, all bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List CLI tools found on PATH",
		RunE: func(cmd *cobra.Command, _ []string) error {
			var es []entry
			for _, e := range inventory(all) {
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
	cmd.Flags().BoolVar(&all, "all", false, "also scan well-known tool dirs not on PATH (~/.cargo/bin, ~/go/bin, ~/.local/bin, ...)")
	return cmd
}

type detail struct {
	entry
	Version string `json:"version,omitempty"`
	About   string `json:"about,omitempty"`
}

func infoCmd() *cobra.Command {
	var asJSON, noRun bool
	cmd := &cobra.Command{
		Use:   "info <name>",
		Short: "Show details about a tool",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			e, ok := find(args[0])
			if !ok {
				return fmt.Errorf("%s not found", args[0])
			}
			d := detail{entry: e, About: info.Whatis(e.Name)}
			if !noRun {
				d.Version = info.Version(e.Path)
			}
			w := cmd.OutOrStdout()
			if asJSON {
				enc := json.NewEncoder(w)
				enc.SetIndent("", "  ")
				return enc.Encode(d)
			}
			fmt.Fprintf(w, "name:     %s\npath:     %s\nreal:     %s\nkind:     %s\n", e.Name, e.Path, e.Real, e.Kind)
			if e.OffPath {
				fmt.Fprintln(w, "on PATH:  no")
			}
			if e.Source != "" {
				fmt.Fprintf(w, "source:   %s\n", e.Source)
			}
			if e.Package != nil {
				fmt.Fprintf(w, "owner:    %s\n", e.Package)
			}
			for _, sh := range e.Shadows {
				fmt.Fprintf(w, "shadows:  %s\n", sh)
			}
			if d.Version != "" {
				fmt.Fprintf(w, "version:  %s\n", d.Version)
			}
			if d.About != "" {
				fmt.Fprintf(w, "about:    %s\n", d.About)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&asJSON, "json", false, "JSON output")
	cmd.Flags().BoolVar(&noRun, "no-run", false, "do not execute the tool to read its version")
	return cmd
}

const maxListed = 10

// siblings lists the PATH commands that belong to pkg, sorted by name.
func siblings(es []entry, pkg provider.Package) []string {
	var names []string
	for _, e := range es {
		if e.Package != nil && *e.Package == pkg {
			names = append(names, e.Name)
		}
	}
	return names
}

// resolve turns the argument into a package to remove. A command on PATH wins;
// otherwise the argument is tried as an installed package name.
func resolve(name string, es []entry) (provider.Provider, provider.Package, error) {
	for _, e := range es {
		if e.Name != name {
			continue
		}
		if e.Kind == classify.Native {
			return nil, provider.Package{}, fmt.Errorf("%s is a native OS tool; refusing to uninstall", e.Name)
		}
		p, pkg, ok := provider.Owner(providers, e.Tool)
		if !ok {
			return nil, provider.Package{}, fmt.Errorf("no package manager owns %s; remove manually: %s", e.Name, e.Real)
		}
		return p, pkg, nil
	}
	if p, pkg, ok := provider.ByName(providers, name); ok {
		return p, pkg, nil
	}
	return nil, provider.Package{}, fmt.Errorf("%s is neither a command on PATH nor an installed package", name)
}

func uninstallCmd() *cobra.Command {
	var yes, dry bool
	cmd := &cobra.Command{
		Use:   "uninstall <command|package>",
		Short: "Uninstall a tool via the package manager that owns it",
		Long: "Uninstall removes the whole package that provides the command, not just the command.\n" +
			"The argument is a command on PATH or an installed package name.",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			out := cmd.OutOrStdout()
			es := inventory(true)
			p, pkg, err := resolve(args[0], es)
			if err != nil {
				return err
			}
			names := siblings(es, pkg)
			fmt.Fprintf(out, "package:  %s\n", pkg)
			if len(names) > 0 {
				shown := names
				if len(shown) > maxListed {
					shown = shown[:maxListed]
				}
				more := ""
				if len(names) > maxListed {
					more = fmt.Sprintf(", +%d more", len(names)-maxListed)
				}
				fmt.Fprintf(out, "removes:  %d command(s) on PATH: %s%s\n", len(names), strings.Join(shown, ", "), more)
			}
			c := p.UninstallCmd(pkg)
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

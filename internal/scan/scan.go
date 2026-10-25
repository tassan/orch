// Package scan discovers executables on PATH.
package scan

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Tool struct {
	Name    string   `json:"name"`
	Path    string   `json:"path"`              // location found on PATH
	Real    string   `json:"real"`              // symlinks resolved
	Shadows []string `json:"shadows,omitempty"` // same-named executables hidden behind this one
	OffPath bool     `json:"off_path,omitempty"`
}

// Scan walks the given dirs; the first dir wins on duplicate names and later
// copies with a different real path are recorded as Shadows.
func Scan(dirs []string) []Tool {
	idx := map[string]int{}
	var tools []Tool
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			p := filepath.Join(dir, e.Name())
			st, err := os.Stat(p) // follows symlinks
			if err != nil || st.IsDir() || st.Mode()&0o111 == 0 {
				continue
			}
			real, err := filepath.EvalSymlinks(p)
			if err != nil {
				real = p
			}
			if i, dup := idx[e.Name()]; dup {
				if real != tools[i].Real {
					tools[i].Shadows = append(tools[i].Shadows, p)
				}
				continue
			}
			idx[e.Name()] = len(tools)
			tools = append(tools, Tool{Name: e.Name(), Path: p, Real: real})
		}
	}
	sort.Slice(tools, func(i, j int) bool { return tools[i].Name < tools[j].Name })
	return tools
}

// PathDirs returns the absolute dirs in a PATH-style string, deduplicated.
// A leading "~" is expanded (shells do it for assignments, the kernel does not).
// Relative entries are skipped: they depend on the working directory.
func PathDirs(path string) []string {
	home, _ := os.UserHomeDir()
	var out []string
	seen := map[string]bool{}
	for _, d := range filepath.SplitList(path) {
		if d == "~" || strings.HasPrefix(d, "~/") {
			if home == "" {
				continue
			}
			d = filepath.Join(home, strings.TrimPrefix(d, "~"))
		}
		if !filepath.IsAbs(d) || seen[d] {
			continue
		}
		seen[d] = true
		out = append(out, d)
	}
	return out
}

// ExtraDirs lists well-known tool dirs that exist but are not in have.
func ExtraDirs(home string, have []string) []string {
	skip := map[string]bool{}
	for _, d := range have {
		skip[d] = true
	}
	var out []string
	for _, rel := range []string{".cargo/bin", "go/bin", ".local/bin", ".bun/bin", ".deno/bin", ".dotnet/tools"} {
		d := filepath.Join(home, rel)
		if st, err := os.Stat(d); err == nil && st.IsDir() && !skip[d] {
			out = append(out, d)
		}
	}
	return out
}

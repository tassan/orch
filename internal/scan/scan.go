// Package scan discovers executables on PATH.
package scan

import (
	"os"
	"path/filepath"
	"sort"
)

type Tool struct {
	Name string `json:"name"`
	Path string `json:"path"` // location found on PATH
	Real string `json:"real"` // symlinks resolved
}

// Scan walks the given dirs; first dir wins on duplicate names.
func Scan(dirs []string) []Tool {
	seen := map[string]bool{}
	var tools []Tool
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if seen[e.Name()] {
				continue
			}
			p := filepath.Join(dir, e.Name())
			st, err := os.Stat(p) // follows symlinks
			if err != nil || st.IsDir() || st.Mode()&0o111 == 0 {
				continue
			}
			real, err := filepath.EvalSymlinks(p)
			if err != nil {
				real = p
			}
			seen[e.Name()] = true
			tools = append(tools, Tool{Name: e.Name(), Path: p, Real: real})
		}
	}
	sort.Slice(tools, func(i, j int) bool { return tools[i].Name < tools[j].Name })
	return tools
}

// PathDirs returns the dirs in a PATH-style string, deduplicated.
func PathDirs(path string) []string {
	var out []string
	seen := map[string]bool{}
	for _, d := range filepath.SplitList(path) {
		if d == "" || seen[d] {
			continue
		}
		seen[d] = true
		out = append(out, d)
	}
	return out
}

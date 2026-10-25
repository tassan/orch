// Package info probes a tool for version and description.
package info

import (
	"context"
	"os/exec"
	"strings"
	"time"
)

// Version probes --version only: other flags (-v, -V, version) are valid args for some tools and have side effects or garbage output.
func Version(path string) string {
	for _, args := range [][]string{{"--version"}} {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		out, err := exec.CommandContext(ctx, path, args...).CombinedOutput()
		cancel()
		if err != nil {
			continue
		}
		if line, _, _ := strings.Cut(strings.TrimSpace(string(out)), "\n"); line != "" {
			return line
		}
	}
	return ""
}

// Whatis returns the man-page one-liner, if any.
func Whatis(name string) string {
	out, err := exec.Command("whatis", name).Output()
	if err != nil {
		return ""
	}
	return pickWhatis(string(out), name)
}

// pickWhatis keeps the entry for exactly name: whatis matches loosely (ls -> npm-ls).
func pickWhatis(out, name string) string {
	for _, line := range strings.Split(out, "\n") {
		if strings.HasPrefix(line, name+"(") || strings.HasPrefix(line, name+" ") {
			return line
		}
	}
	return ""
}

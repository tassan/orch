// Package classify decides whether a tool ships with the OS or was user-installed (macOS rules).
package classify

import (
	"path/filepath"
	"strings"
)

type Kind string

const (
	Native Kind = "native"
	User   Kind = "user"
)

var nativePrefixes = []string{"/usr/bin", "/bin", "/usr/sbin", "/usr/libexec", "/sbin", "/System", "/Library/Apple"}

// Classify uses the symlink-resolved path.
func Classify(real string) Kind {
	for _, p := range nativePrefixes {
		if real == p || strings.HasPrefix(real, p+"/") {
			return Native
		}
	}
	return User
}

// Source guesses where a user tool came from, from its resolved path.
func Source(real, home string) string {
	switch {
	case strings.HasPrefix(real, "/opt/homebrew/"), strings.Contains(real, "/Cellar/"), strings.HasPrefix(real, "/usr/local/Homebrew/"):
		return "brew"
	case home != "" && strings.HasPrefix(real, filepath.Join(home, ".cargo")+"/"):
		return "cargo"
	case home != "" && strings.HasPrefix(real, filepath.Join(home, "go", "bin")+"/"):
		return "go"
	case strings.Contains(real, "/node_modules/"):
		return "npm"
	case strings.HasPrefix(real, "/usr/local/"):
		return "manual"
	}
	return "unknown"
}

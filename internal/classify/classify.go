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
// A package manager that owns the tool overrides this guess.
func Source(real, home string) string {
	under := func(rel string) bool { return home != "" && strings.HasPrefix(real, filepath.Join(home, rel)+"/") }
	switch {
	case strings.HasPrefix(real, "/opt/homebrew/"), strings.Contains(real, "/Cellar/"), strings.HasPrefix(real, "/usr/local/Homebrew/"):
		return "brew"
	case under(".cargo"):
		return "cargo"
	case under("go/bin"):
		return "go"
	case under(".nvm"):
		return "nvm"
	case under(".asdf"):
		return "asdf"
	case under(".local/share/mise"):
		return "mise"
	case under(".pyenv"):
		return "pyenv"
	case under(".rbenv"):
		return "rbenv"
	case strings.Contains(real, "/pipx/venvs/"):
		return "pipx"
	case strings.Contains(real, "/node_modules/"):
		return "npm"
	case strings.Contains(real, ".app/"):
		return "app"
	case under(".local"):
		return "local"
	case strings.HasPrefix(real, "/usr/local/"):
		return "manual"
	}
	return "unknown"
}

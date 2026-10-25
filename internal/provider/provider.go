// Package provider delegates uninstalls to the package manager that owns a tool.
package provider

import (
	"os/exec"
	"strings"

	"github.com/tassan/orch/internal/scan"
)

type Provider interface {
	Name() string
	Available() bool
	// Owns reports whether this manager installed the tool.
	Owns(t scan.Tool) bool
	// UninstallCmd returns the command that would uninstall the tool.
	UninstallCmd(t scan.Tool) (*exec.Cmd, error)
}

// All returns the known providers.
func All() []Provider { return []Provider{Brew{}} }

// Find returns the available provider that owns t, or nil.
func Find(ps []Provider, t scan.Tool) Provider {
	for _, p := range ps {
		if p.Available() && p.Owns(t) {
			return p
		}
	}
	return nil
}

func lookPath(bin string) bool {
	_, err := exec.LookPath(bin)
	return err == nil
}

func joinCmd(c *exec.Cmd) string { return strings.Join(c.Args, " ") }

// Display renders a command for dry-run output.
func Display(c *exec.Cmd) string { return joinCmd(c) }

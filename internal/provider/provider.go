// Package provider delegates uninstalls to the package manager that owns a tool.
package provider

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/tassan/orch/internal/scan"
)

// Package identifies something a package manager installed.
type Package struct {
	Provider string `json:"provider"`
	Name     string `json:"name"`
	Cask     bool   `json:"cask,omitempty"`
}

func (p Package) String() string {
	switch {
	case p.Cask:
		return fmt.Sprintf("%s cask %s", p.Provider, p.Name)
	case p.Provider == "brew":
		return fmt.Sprintf("%s formula %s", p.Provider, p.Name)
	}
	return fmt.Sprintf("%s %s", p.Provider, p.Name)
}

type Provider interface {
	Name() string
	Available() bool
	// Owner reports the package that installed the tool.
	Owner(t scan.Tool) (Package, bool)
	// Lookup finds an installed package by its own name (not a command name).
	Lookup(name string) (Package, bool)
	// UninstallCmd returns the command that would uninstall the package.
	UninstallCmd(p Package) *exec.Cmd
}

// All returns the known providers.
func All() []Provider { return []Provider{&Brew{}, &Cargo{}, &Npm{}, &Pipx{}} }

// Owner returns the available provider and package that own t.
func Owner(ps []Provider, t scan.Tool) (Provider, Package, bool) {
	for _, p := range ps {
		if !p.Available() {
			continue
		}
		if pkg, ok := p.Owner(t); ok {
			return p, pkg, true
		}
	}
	return nil, Package{}, false
}

// ByName finds an installed package by name across available providers.
func ByName(ps []Provider, name string) (Provider, Package, bool) {
	for _, p := range ps {
		if !p.Available() {
			continue
		}
		if pkg, ok := p.Lookup(name); ok {
			return p, pkg, true
		}
	}
	return nil, Package{}, false
}

type runner func(args ...string) ([]byte, error)

func command(name string, run runner, args ...string) ([]byte, error) {
	if run != nil {
		return run(args...)
	}
	return exec.Command(name, args...).Output()
}

func lookPath(bin string) bool {
	_, err := exec.LookPath(bin)
	return err == nil
}

// Display renders a command for dry-run output.
func Display(c *exec.Cmd) string { return strings.Join(c.Args, " ") }

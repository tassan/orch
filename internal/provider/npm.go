package provider

import (
	"encoding/json"
	"os/exec"
	"strings"
	"sync"

	"github.com/tassan/orch/internal/scan"
)

// Npm handles global npm packages (including those under nvm).
type Npm struct {
	run runner

	once sync.Once
	pkgs map[string]bool
}

func (n *Npm) Name() string    { return "npm" }
func (n *Npm) Available() bool { return n.run != nil || lookPath("npm") }

// npmPackage parses ".../lib/node_modules/<pkg>/..." (or "@scope/pkg") from a resolved path.
func npmPackage(real string) string {
	_, rest, ok := strings.Cut(real, "/lib/node_modules/")
	if !ok {
		return ""
	}
	parts := strings.Split(rest, "/")
	if strings.HasPrefix(parts[0], "@") && len(parts) > 1 {
		return parts[0] + "/" + parts[1]
	}
	return parts[0]
}

func (n *Npm) Owner(t scan.Tool) (Package, bool) {
	if p := npmPackage(t.Real); p != "" {
		return Package{Provider: "npm", Name: p}, true
	}
	return Package{}, false
}

func (n *Npm) Lookup(name string) (Package, bool) {
	n.once.Do(func() {
		n.pkgs = map[string]bool{}
		out, err := command("npm", n.run, "ls", "-g", "--depth=0", "--json")
		var doc struct {
			Dependencies map[string]json.RawMessage `json:"dependencies"`
		}
		if err == nil && json.Unmarshal(out, &doc) == nil {
			for k := range doc.Dependencies {
				n.pkgs[k] = true
			}
		}
	})
	if n.pkgs[name] {
		return Package{Provider: "npm", Name: name}, true
	}
	return Package{}, false
}

func (*Npm) UninstallCmd(p Package) *exec.Cmd {
	return exec.Command("npm", "uninstall", "-g", p.Name)
}

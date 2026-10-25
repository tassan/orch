package provider

import (
	"os/exec"
	"strings"
	"sync"

	"github.com/tassan/orch/internal/scan"
)

// Pipx handles Python CLI apps installed with pipx.
type Pipx struct {
	run runner

	once sync.Once
	pkgs map[string]bool
}

func (p *Pipx) Name() string    { return "pipx" }
func (p *Pipx) Available() bool { return p.run != nil || lookPath("pipx") }

func (p *Pipx) Owner(t scan.Tool) (Package, bool) {
	_, rest, ok := strings.Cut(t.Real, "/pipx/venvs/")
	if !ok {
		return Package{}, false
	}
	name, _, _ := strings.Cut(rest, "/")
	return Package{Provider: "pipx", Name: name}, name != ""
}

func (p *Pipx) Lookup(name string) (Package, bool) {
	p.once.Do(func() {
		p.pkgs = map[string]bool{}
		out, err := command("pipx", p.run, "list", "--short")
		if err != nil {
			return
		}
		for _, l := range strings.Split(string(out), "\n") {
			if f := strings.Fields(l); len(f) > 0 {
				p.pkgs[f[0]] = true
			}
		}
	})
	if p.pkgs[name] {
		return Package{Provider: "pipx", Name: name}, true
	}
	return Package{}, false
}

func (*Pipx) UninstallCmd(p Package) *exec.Cmd { return exec.Command("pipx", "uninstall", p.Name) }

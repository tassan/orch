package provider

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/tassan/orch/internal/scan"
)

// Cargo handles crates installed with `cargo install`, read from $CARGO_HOME/.crates2.json.
type Cargo struct {
	dir string // override for tests; default $CARGO_HOME or ~/.cargo

	once   sync.Once
	bins   map[string]string // binary name -> crate
	crates map[string]bool
}

func (*Cargo) Name() string    { return "cargo" }
func (*Cargo) Available() bool { return lookPath("cargo") }

func (c *Cargo) home() string {
	if c.dir != "" {
		return c.dir
	}
	if d := os.Getenv("CARGO_HOME"); d != "" {
		return d
	}
	h, _ := os.UserHomeDir()
	return filepath.Join(h, ".cargo")
}

func (c *Cargo) load() {
	c.once.Do(func() {
		c.bins, c.crates = map[string]string{}, map[string]bool{}
		data, err := os.ReadFile(filepath.Join(c.home(), ".crates2.json"))
		if err != nil {
			return
		}
		var doc struct {
			Installs map[string]struct {
				Bins []string `json:"bins"`
			} `json:"installs"`
		}
		if json.Unmarshal(data, &doc) != nil {
			return
		}
		for key, inst := range doc.Installs {
			name, _, _ := strings.Cut(key, " ") // "ripgrep 14.1.0 (registry+...)"
			c.crates[name] = true
			for _, b := range inst.Bins {
				c.bins[b] = name
			}
		}
	})
}

func (c *Cargo) Owner(t scan.Tool) (Package, bool) {
	if filepath.Dir(t.Real) != filepath.Join(c.home(), "bin") {
		return Package{}, false
	}
	c.load()
	if crate, ok := c.bins[filepath.Base(t.Real)]; ok {
		return Package{Provider: "cargo", Name: crate}, true
	}
	return Package{}, false
}

func (c *Cargo) Lookup(name string) (Package, bool) {
	c.load()
	if c.crates[name] {
		return Package{Provider: "cargo", Name: name}, true
	}
	return Package{}, false
}

func (*Cargo) UninstallCmd(p Package) *exec.Cmd { return exec.Command("cargo", "uninstall", p.Name) }

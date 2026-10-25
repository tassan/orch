package provider

import (
	"encoding/json"
	"os/exec"
	"path"
	"strings"
	"sync"

	"github.com/tassan/orch/internal/scan"
)

// Brew handles Homebrew formulae and casks. The zero value runs the real brew.
// Cask data and the formula list each cost a brew call, so they load lazily, once.
type Brew struct {
	run runner

	caskOnce sync.Once
	casks    map[string]bool   // installed cask tokens
	apps     map[string]string // "Cursor.app" -> token
	bins     map[string]string // cask binary path -> token

	formulaOnce sync.Once
	formulae    map[string]bool
}

func (b *Brew) Name() string    { return "brew" }
func (b *Brew) Available() bool { return b.run != nil || lookPath("brew") }

func (b *Brew) exec(args ...string) ([]byte, error) {
	return command("brew", b.run, args...)
}

// formula parses ".../Cellar/<formula>/<version>/..." from a resolved path.
func formula(real string) string {
	_, rest, ok := strings.Cut(real, "/Cellar/")
	if !ok {
		return ""
	}
	name, _, _ := strings.Cut(rest, "/")
	return name
}

// caskroom parses ".../Caskroom/<token>/..." from a resolved path.
func caskroom(real string) string {
	_, rest, ok := strings.Cut(real, "/Caskroom/")
	if !ok {
		return ""
	}
	name, _, _ := strings.Cut(rest, "/")
	return name
}

// appBundle returns the first "X.app" path segment, if any.
func appBundle(real string) string {
	for _, seg := range strings.Split(real, "/") {
		if strings.HasSuffix(seg, ".app") {
			return seg
		}
	}
	return ""
}

func (b *Brew) loadCasks() {
	b.caskOnce.Do(func() {
		b.casks, b.apps, b.bins = map[string]bool{}, map[string]string{}, map[string]string{}
		out, err := b.exec("info", "--cask", "--json=v2", "--installed")
		if err != nil {
			return
		}
		var doc struct {
			Casks []struct {
				Token     string                       `json:"token"`
				Artifacts []map[string]json.RawMessage `json:"artifacts"`
			} `json:"casks"`
		}
		if json.Unmarshal(out, &doc) != nil {
			return
		}
		for _, c := range doc.Casks {
			b.casks[c.Token] = true
			for _, art := range c.Artifacts {
				for kind, raw := range art {
					if kind != "app" && kind != "binary" {
						continue
					}
					var items []json.RawMessage
					if json.Unmarshal(raw, &items) != nil {
						continue
					}
					for _, it := range items {
						var s string
						if json.Unmarshal(it, &s) != nil {
							continue // {"target": ...} option object
						}
						if kind == "app" {
							b.apps[path.Base(s)] = c.Token
						} else {
							b.bins[s] = c.Token
						}
					}
				}
			}
		}
	})
}

func (b *Brew) Owner(t scan.Tool) (Package, bool) {
	if f := formula(t.Real); f != "" {
		return Package{Provider: "brew", Name: f}, true
	}
	if c := caskroom(t.Real); c != "" {
		return Package{Provider: "brew", Name: c, Cask: true}, true
	}
	if app := appBundle(t.Real); app != "" {
		b.loadCasks()
		if tok, ok := b.apps[app]; ok {
			return Package{Provider: "brew", Name: tok, Cask: true}, true
		}
		if tok, ok := b.bins[t.Real]; ok {
			return Package{Provider: "brew", Name: tok, Cask: true}, true
		}
	}
	return Package{}, false
}

func (b *Brew) Lookup(name string) (Package, bool) {
	b.formulaOnce.Do(func() {
		b.formulae = map[string]bool{}
		out, err := b.exec("list", "--formula", "-1")
		if err != nil {
			return
		}
		for _, l := range strings.Split(string(out), "\n") {
			if l = strings.TrimSpace(l); l != "" {
				b.formulae[l] = true
			}
		}
	})
	if b.formulae[name] {
		return Package{Provider: "brew", Name: name}, true
	}
	b.loadCasks()
	if b.casks[name] {
		return Package{Provider: "brew", Name: name, Cask: true}, true
	}
	return Package{}, false
}

func (*Brew) UninstallCmd(p Package) *exec.Cmd {
	if p.Cask {
		return exec.Command("brew", "uninstall", "--cask", p.Name)
	}
	return exec.Command("brew", "uninstall", "--formula", p.Name)
}

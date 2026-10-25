package provider

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tassan/orch/internal/scan"
)

func TestCargo(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, ".crates2.json"), []byte(
		`{"installs":{"ripgrep 14.1.0 (registry+https://github.com/rust-lang/crates.io-index)":{"bins":["rg"]}}}`), 0o644)
	c := &Cargo{dir: dir}
	if p, ok := c.Owner(scan.Tool{Real: filepath.Join(dir, "bin", "rg")}); !ok || p.Name != "ripgrep" {
		t.Errorf("rg: %+v %v", p, ok)
	}
	if _, ok := c.Owner(scan.Tool{Real: filepath.Join(dir, "bin", "rustc")}); ok {
		t.Error("rustup proxy must not be owned")
	}
	if _, ok := c.Owner(scan.Tool{Real: "/usr/bin/rg"}); ok {
		t.Error("outside cargo bin must not be owned")
	}
	if _, ok := c.Lookup("ripgrep"); !ok {
		t.Error("crate lookup failed")
	}
	if got := Display(c.UninstallCmd(Package{Name: "ripgrep"})); got != "cargo uninstall ripgrep" {
		t.Errorf("got %q", got)
	}
}

func TestCargoMissingIndex(t *testing.T) {
	c := &Cargo{dir: t.TempDir()}
	if _, ok := c.Lookup("x"); ok {
		t.Error("no index should own nothing")
	}
}

func TestNpm(t *testing.T) {
	n := &Npm{run: func(args ...string) ([]byte, error) {
		return []byte(`{"dependencies":{"homebridge":{},"@scope/tool":{}}}`), nil
	}}
	for real, want := range map[string]string{
		"/Users/a/.nvm/versions/node/v24/lib/node_modules/homebridge/bin/homebridge.js": "homebridge",
		"/x/lib/node_modules/@scope/tool/dist/cli.js":                                   "@scope/tool",
	} {
		if p, ok := n.Owner(scan.Tool{Real: real}); !ok || p.Name != want {
			t.Errorf("%s: %+v %v", real, p, ok)
		}
	}
	if _, ok := n.Owner(scan.Tool{Real: "/usr/bin/node"}); ok {
		t.Error("plain binary must not be owned")
	}
	if _, ok := n.Lookup("@scope/tool"); !ok {
		t.Error("lookup failed")
	}
	if got := Display(n.UninstallCmd(Package{Name: "homebridge"})); got != "npm uninstall -g homebridge" {
		t.Errorf("got %q", got)
	}
}

func TestPipx(t *testing.T) {
	p := &Pipx{run: func(args ...string) ([]byte, error) { return []byte("black 24.1.0\nhttpie 3.2\n"), nil }}
	if pkg, ok := p.Owner(scan.Tool{Real: "/Users/a/.local/pipx/venvs/black/bin/black"}); !ok || pkg.Name != "black" {
		t.Errorf("%+v %v", pkg, ok)
	}
	if _, ok := p.Lookup("httpie"); !ok {
		t.Error("lookup failed")
	}
	if _, ok := p.Lookup("nope"); ok {
		t.Error("nope resolved")
	}
	if got := Display(p.UninstallCmd(Package{Name: "black"})); got != "pipx uninstall black" {
		t.Errorf("got %q", got)
	}
}

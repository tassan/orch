package provider

import (
	"errors"
	"strings"
	"testing"

	"github.com/tassan/orch/internal/scan"
)

const caskJSON = `{"casks":[
 {"token":"cursor","artifacts":[{"app":["Cursor.app"]},{"binary":["/Applications/Cursor.app/Contents/Resources/app/bin/code",{"target":"cursor"}]}]},
 {"token":"dotnet-sdk","artifacts":[{"binary":["/usr/local/share/dotnet/dotnet"]}]},
 {"token":"font-hack","artifacts":[]}]}`

func fakeBrew(calls *[]string) *Brew {
	return &Brew{run: func(args ...string) ([]byte, error) {
		*calls = append(*calls, strings.Join(args, " "))
		switch args[0] {
		case "info":
			return []byte(caskJSON), nil
		case "list":
			return []byte("jq\nsvt-av1\n"), nil
		}
		return nil, errors.New("unexpected")
	}}
}

func TestFormula(t *testing.T) {
	if got := formula("/opt/homebrew/Cellar/jq/1.7/bin/jq"); got != "jq" {
		t.Errorf("got %q", got)
	}
	if got := formula("/usr/bin/ls"); got != "" {
		t.Errorf("got %q", got)
	}
}

func TestOwner(t *testing.T) {
	var calls []string
	b := fakeBrew(&calls)
	for real, want := range map[string]Package{
		"/opt/homebrew/Cellar/svt-av1/4.1.0/bin/x":                 {Provider: "brew", Name: "svt-av1"},
		"/opt/homebrew/Caskroom/vlc/3.0/vlc.wrapper.sh":            {Provider: "brew", Name: "vlc", Cask: true},
		"/Applications/Cursor.app/Contents/Resources/app/bin/code": {Provider: "brew", Name: "cursor", Cask: true},
	} {
		got, ok := b.Owner(scan.Tool{Real: real})
		if !ok || got != want {
			t.Errorf("%s: got %+v ok=%v want %+v", real, got, ok, want)
		}
	}
	for _, real := range []string{"/Applications/Docker.app/Contents/bin/docker", "/usr/bin/ls", "/Users/a/.nvm/bin/node"} {
		if got, ok := b.Owner(scan.Tool{Real: real}); ok {
			t.Errorf("%s: unexpected owner %+v", real, got)
		}
	}
}

func TestOwnerLoadsCasksLazilyOnce(t *testing.T) {
	var calls []string
	b := fakeBrew(&calls)
	b.Owner(scan.Tool{Real: "/opt/homebrew/Cellar/jq/1/bin/jq"})
	b.Owner(scan.Tool{Real: "/usr/bin/ls"})
	if len(calls) != 0 {
		t.Fatalf("brew called for non-app paths: %v", calls)
	}
	b.Owner(scan.Tool{Real: "/Applications/A.app/x"})
	b.Owner(scan.Tool{Real: "/Applications/B.app/x"})
	if len(calls) != 1 {
		t.Fatalf("want 1 cask load, got %v", calls)
	}
}

func TestLookup(t *testing.T) {
	var calls []string
	b := fakeBrew(&calls)
	if p, ok := b.Lookup("svt-av1"); !ok || p.Cask {
		t.Errorf("svt-av1: %+v %v", p, ok)
	}
	if p, ok := b.Lookup("cursor"); !ok || !p.Cask {
		t.Errorf("cursor: %+v %v", p, ok)
	}
	if _, ok := b.Lookup("nope"); ok {
		t.Error("nope should not resolve")
	}
}

func TestUninstallCmd(t *testing.T) {
	b := &Brew{}
	if got := Display(b.UninstallCmd(Package{Name: "jq"})); got != "brew uninstall --formula jq" {
		t.Errorf("got %q", got)
	}
	if got := Display(b.UninstallCmd(Package{Name: "cursor", Cask: true})); got != "brew uninstall --cask cursor" {
		t.Errorf("got %q", got)
	}
}

func TestPackageString(t *testing.T) {
	for p, want := range map[Package]string{
		{Provider: "brew", Name: "jq"}:              "brew formula jq",
		{Provider: "brew", Name: "vlc", Cask: true}: "brew cask vlc",
		{Provider: "cargo", Name: "ripgrep"}:        "cargo ripgrep",
	} {
		if got := p.String(); got != want {
			t.Errorf("got %q want %q", got, want)
		}
	}
}

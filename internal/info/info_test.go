package info

import (
	"os"
	"path/filepath"
	"testing"
)

func script(t *testing.T, body string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "tool")
	if err := os.WriteFile(p, []byte("#!/bin/sh\n"+body), 0o755); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestVersionFirstLineOfDashDashVersion(t *testing.T) {
	p := script(t, `[ "$1" = "--version" ] && printf 'tool 1.2.3\nextra\n'`)
	if got := Version(p); got != "tool 1.2.3" {
		t.Errorf("got %q", got)
	}
}

func TestVersionEmptyOnFailureAndOtherFlagsNotTried(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "other-flag")
	p := script(t, `[ "$1" != "--version" ] && touch `+marker+`; exit 1`)
	if got := Version(p); got != "" {
		t.Errorf("got %q", got)
	}
	if _, err := os.Stat(marker); err == nil {
		t.Error("probed a flag other than --version")
	}
}

func TestPickWhatis(t *testing.T) {
	out := "ls(1)                    - list directory contents\nnpm-ls(1) - List installed packages\n"
	if got := pickWhatis(out, "ls"); got != "ls(1)                    - list directory contents" {
		t.Errorf("got %q", got)
	}
	if got := pickWhatis("npm-ls(1) - x\n", "ls"); got != "" {
		t.Errorf("loose match leaked: %q", got)
	}
}

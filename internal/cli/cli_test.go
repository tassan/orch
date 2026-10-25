package cli

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tassan/orch/internal/provider"
	"github.com/tassan/orch/internal/scan"
)

// fake owns tools by resolved-path suffix and knows installed package names.
type fake struct {
	owners   map[string]provider.Package // base name of Real -> package
	packages map[string]provider.Package
}

func (fake) Name() string    { return "fake" }
func (fake) Available() bool { return true }
func (f fake) Owner(t scan.Tool) (provider.Package, bool) {
	p, ok := f.owners[filepath.Base(t.Real)]
	return p, ok
}
func (f fake) Lookup(n string) (provider.Package, bool) { p, ok := f.packages[n]; return p, ok }
func (fake) UninstallCmd(p provider.Package) *exec.Cmd {
	return exec.Command("fake-uninstall", p.Name)
}

func setup(t *testing.T, tools ...string) {
	t.Helper()
	dir := t.TempDir()
	for _, n := range tools {
		if err := os.WriteFile(filepath.Join(dir, n), []byte("x"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", dir+":/bin")
	pkg := provider.Package{Provider: "fake", Name: "multi"}
	providers = []provider.Provider{fake{
		owners:   map[string]provider.Package{"a": pkg, "b": pkg},
		packages: map[string]provider.Package{"multi": pkg, "idle": {Provider: "fake", Name: "idle", Cask: true}},
	}}
	t.Cleanup(func() { providers = provider.All() })
}

func run(args ...string) (string, error) {
	var buf bytes.Buffer
	cmd := Root()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetIn(strings.NewReader(""))
	cmd.SetArgs(args)
	err := cmd.Execute()
	return buf.String(), err
}

func TestUninstallByCommandListsSiblings(t *testing.T) {
	setup(t, "a", "b", "loner")
	out, err := run("uninstall", "a", "--dry-run")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"package:  fake multi", "2 command(s) on PATH: a, b", "would run: fake-uninstall multi"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
}

func TestUninstallByPackageName(t *testing.T) {
	setup(t, "a", "b")
	out, err := run("uninstall", "idle", "--dry-run")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "fake cask idle") || !strings.Contains(out, "fake-uninstall idle") {
		t.Errorf("got:\n%s", out)
	}
	if strings.Contains(out, "removes:") {
		t.Errorf("package with no PATH commands should not list removals:\n%s", out)
	}
}

func TestUninstallRefusals(t *testing.T) {
	setup(t, "loner")
	for arg, want := range map[string]string{
		"ls":      "native OS tool",
		"loner":   "no package manager owns loner",
		"missing": "neither a command on PATH nor an installed package",
	} {
		out, err := run("uninstall", arg, "--dry-run")
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("%s: err=%v out=%s", arg, err, out)
		}
		if strings.Contains(out, "would run") {
			t.Errorf("%s: must not plan a command", arg)
		}
	}
}

func TestUninstallAbortsWithoutYes(t *testing.T) {
	setup(t, "a")
	out, err := run("uninstall", "a")
	if err == nil || !strings.Contains(err.Error(), "aborted") {
		t.Fatalf("err=%v out=%s", err, out)
	}
}

func TestListJSONIncludesPackage(t *testing.T) {
	setup(t, "a", "loner")
	out, err := run("list", "--user", "--json")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, `"name": "multi"`) || !strings.Contains(out, `"source": "fake"`) {
		t.Errorf("owned tool should carry package and source:\n%s", out)
	}
}

func TestInfoJSONShadowsAndNoRun(t *testing.T) {
	setup(t, "loner")
	marker := filepath.Join(t.TempDir(), "ran")
	dir := t.TempDir()
	must(t, os.WriteFile(filepath.Join(dir, "probe"), []byte("#!/bin/sh\n/usr/bin/touch "+marker+"\necho v1\n"), 0o755))
	other := t.TempDir()
	must(t, os.WriteFile(filepath.Join(other, "probe"), []byte("#!/bin/sh\necho v2\n"), 0o755))
	t.Setenv("PATH", dir+":"+other)

	out, err := run("info", "probe", "--no-run", "--json")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(marker); err == nil {
		t.Error("--no-run executed the tool")
	}
	if !strings.Contains(out, `"shadows"`) || !strings.Contains(out, other) || strings.Contains(out, `"version"`) {
		t.Errorf("json:\n%s", out)
	}
	out, _ = run("info", "probe")
	if _, err := os.Stat(marker); err != nil || !strings.Contains(out, "version:  v1") || !strings.Contains(out, "shadows:  "+filepath.Join(other, "probe")) {
		t.Errorf("default info should run the tool:\n%s", out)
	}
}

func TestListAllFindsOffPathTools(t *testing.T) {
	setup(t, "onpath")
	home := t.TempDir()
	t.Setenv("HOME", home)
	cargoBin := filepath.Join(home, ".cargo", "bin")
	must(t, os.MkdirAll(cargoBin, 0o755))
	must(t, os.WriteFile(filepath.Join(cargoBin, "offtool"), []byte("x"), 0o755))

	out, _ := run("list", "--user", "--json")
	if strings.Contains(out, "offtool") {
		t.Error("off-PATH tool listed without --all")
	}
	out, _ = run("list", "--user", "--json", "--all")
	if !strings.Contains(out, `"name": "offtool"`) || !strings.Contains(out, `"off_path": true`) {
		t.Errorf("--all should list it:\n%s", out)
	}
	if out, err := run("info", "offtool", "--no-run"); err != nil || !strings.Contains(out, "on PATH:  no") {
		t.Errorf("info should find off-PATH tool: %v\n%s", err, out)
	}
}

func TestVersionIncludesBuildInfo(t *testing.T) {
	v, c, d := Version, Commit, Date
	t.Cleanup(func() { Version, Commit, Date = v, c, d })
	Version, Commit, Date = "1.2.3", "abc1234", "2026-10-25"
	if got := version(); got != "1.2.3 (abc1234, 2026-10-25)" {
		t.Errorf("got %q", got)
	}
	Commit = ""
	if got := version(); got != "1.2.3" {
		t.Errorf("got %q", got)
	}
}

func TestVersionFlag(t *testing.T) {
	out, err := run("--version")
	if err != nil || !strings.Contains(out, "orch version") {
		t.Errorf("err=%v out=%q", err, out)
	}
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func historyHome(t *testing.T, zsh string) {
	t.Helper()
	home := t.TempDir()
	must(t, os.WriteFile(filepath.Join(home, ".zsh_history"), []byte(zsh), 0o600))
	t.Setenv("HOME", home)
	t.Setenv("HISTFILE", "")
	t.Setenv("XDG_DATA_HOME", "")
}

const sampleHistory = ": 1700000000:0;git status\n" +
	": 1700000100:0;git commit -m one\n" +
	": 1700000200:0;ls\n" +
	": 1700000300:0;FOO=1 git commit -m two\n" +
	": 1700000400:0;make && git push\n"

func TestHistoryFiltersAndLimits(t *testing.T) {
	historyHome(t, sampleHistory)
	out, err := run("history", "git")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(out, "git") < 4 || strings.Contains(out, "ls\n") {
		t.Errorf("want 4 git calls only:\n%s", out)
	}
	out, _ = run("history", "git", "commit")
	if !strings.Contains(out, "git commit -m one") || !strings.Contains(out, "FOO=1 git commit -m two") || strings.Contains(out, "git status") {
		t.Errorf("args should narrow to commit:\n%s", out)
	}
	out, _ = run("history", "-n", "1", "git", "commit")
	if strings.Contains(out, "one") || !strings.Contains(out, "two") {
		t.Errorf("-n 1 should keep the newest only:\n%s", out)
	}
	if !strings.Contains(out, `showing 1 of 2 calls of "git commit"`) {
		t.Errorf("footer missing:\n%s", out)
	}
}

func TestHistoryArgsAreNotOrchFlags(t *testing.T) {
	historyHome(t, ": 1700000000:0;ls -la\n: 1700000100:0;ls\n")
	out, err := run("history", "ls", "-la")
	if err != nil || !strings.Contains(out, "ls -la") {
		t.Fatalf("err=%v out=%s", err, out)
	}
	if !strings.Contains(out, "showing 1 of 1 calls") {
		t.Errorf("only `ls -la` should match:\n%s", out)
	}
}

func TestHistoryJSONAndErrors(t *testing.T) {
	historyHome(t, sampleHistory)
	out, err := run("history", "--json", "ls")
	if err != nil || !strings.Contains(out, `"line": "ls"`) || !strings.Contains(out, `"time"`) {
		t.Errorf("err=%v out=%s", err, out)
	}
	if _, err := run("history", "nosuchcmd"); err == nil || !strings.Contains(err.Error(), "no calls of") {
		t.Errorf("err=%v", err)
	}
	if _, err := run("history", "-n", "-1", "git"); err == nil {
		t.Error("negative limit accepted")
	}
	t.Setenv("HOME", t.TempDir())
	if _, err := run("history", "git"); err == nil || !strings.Contains(err.Error(), "no shell history files") {
		t.Errorf("err=%v", err)
	}
}

func TestHistoryFileFlag(t *testing.T) {
	historyHome(t, "")
	f := filepath.Join(t.TempDir(), "h")
	must(t, os.WriteFile(f, []byte("echo hi\n"), 0o600))
	out, err := run("history", "--file", f, "echo")
	if err != nil || !strings.Contains(out, "echo hi") {
		t.Errorf("err=%v out=%s", err, out)
	}
}

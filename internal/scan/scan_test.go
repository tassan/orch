package scan

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScanFirstWinsAndSkipsNonExec(t *testing.T) {
	a, b := t.TempDir(), t.TempDir()
	write := func(dir, name string, mode os.FileMode) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), mode); err != nil {
			t.Fatal(err)
		}
	}
	write(a, "tool", 0o755)
	write(b, "tool", 0o755)
	write(b, "other", 0o755)
	write(b, "data", 0o644)

	got := Scan([]string{a, b})
	if len(got) != 2 {
		t.Fatalf("want 2 tools, got %+v", got)
	}
	if got[1].Name != "tool" || filepath.Dir(got[1].Path) != a {
		t.Errorf("first dir should win: %+v", got[1])
	}
}

func TestScanShadowsSymlinksAndSkips(t *testing.T) {
	a, b := t.TempDir(), t.TempDir()
	must(t, os.WriteFile(filepath.Join(a, "tool"), []byte("x"), 0o755))
	must(t, os.WriteFile(filepath.Join(b, "tool"), []byte("x"), 0o755))
	must(t, os.Symlink(filepath.Join(a, "tool"), filepath.Join(b, "same"))) // alias of a/tool
	must(t, os.Symlink(filepath.Join(a, "tool"), filepath.Join(a, "alias")))
	must(t, os.Symlink(filepath.Join(a, "gone"), filepath.Join(a, "broken")))
	must(t, os.Mkdir(filepath.Join(a, "subdir"), 0o755))

	got := map[string]Tool{}
	for _, tl := range Scan([]string{a, b}) {
		got[tl.Name] = tl
	}
	if len(got["tool"].Shadows) != 1 || filepath.Dir(got["tool"].Shadows[0]) != b {
		t.Errorf("want one shadow in b: %+v", got["tool"])
	}
	if got["alias"].Real != got["tool"].Real {
		t.Errorf("symlink should resolve to real: %+v", got["alias"])
	}
	if _, ok := got["broken"]; ok {
		t.Error("broken symlink listed")
	}
	if _, ok := got["subdir"]; ok {
		t.Error("directory listed")
	}
}

func TestPathDirs(t *testing.T) {
	t.Setenv("HOME", "/home/u")
	got := PathDirs("/a:~/bin:relative/dir::/a:~")
	want := []string{"/a", "/home/u/bin", "/home/u"}
	if len(got) != len(want) {
		t.Fatalf("got %v want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("got %v want %v", got, want)
		}
	}
}

func TestExtraDirs(t *testing.T) {
	home := t.TempDir()
	must(t, os.MkdirAll(filepath.Join(home, ".cargo/bin"), 0o755))
	must(t, os.MkdirAll(filepath.Join(home, "go/bin"), 0o755))
	got := ExtraDirs(home, []string{filepath.Join(home, "go/bin")})
	if len(got) != 1 || got[0] != filepath.Join(home, ".cargo/bin") {
		t.Errorf("got %v", got)
	}
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

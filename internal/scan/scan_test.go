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

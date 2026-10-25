package history

import (
	"os"
	"path/filepath"
	"testing"
)

func write(t *testing.T, dir, name, body string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func lines(es []Entry) []string {
	var out []string
	for _, e := range es {
		out = append(out, e.Line)
	}
	return out
}

func eq(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("got %q want %q", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %q want %q", got, want)
		}
	}
}

func TestZshExtendedMultilineAndMeta(t *testing.T) {
	d := t.TempDir()
	body := ": 1700000000:0;git status\n" +
		": 1700000010:0;echo one \\\ntwo\n" +
		": 1700000020:0;echo a\xe2\x80\x83\xb4b\n" // zsh stores em dash (e2 80 94) as e2 80 83 (94^32): bytes 0x83-0x9f are escaped
	es, err := Load([]string{write(t, d, ".zsh_history", body)})
	if err != nil {
		t.Fatal(err)
	}
	eq(t, lines(es), []string{"git status", "echo one \ntwo", "echo a—b"})
	if es[0].Time == nil || es[0].Time.Unix() != 1700000000 {
		t.Errorf("time: %+v", es[0].Time)
	}
}

func TestBashPlainAndTimestamped(t *testing.T) {
	d := t.TempDir()
	es, _ := Load([]string{write(t, d, ".bash_history", "ls\n#1700000005\ngit log\n\n")})
	eq(t, lines(es), []string{"ls", "git log"})
	if es[0].Time != nil || es[1].Time == nil || es[1].Time.Unix() != 1700000005 {
		t.Errorf("times: %+v %+v", es[0].Time, es[1].Time)
	}
}

func TestFish(t *testing.T) {
	d := t.TempDir()
	body := "- cmd: git commit -m \"a\\\\b\"\n  when: 1700000100\n- cmd: echo x\\ny\n  when: 1700000200\n  paths:\n    - foo\n"
	es, _ := Load([]string{write(t, d, "fish_history", body)})
	eq(t, lines(es), []string{`git commit -m "a\b"`, "echo x\ny"})
	if es[1].Time.Unix() != 1700000200 {
		t.Errorf("time: %v", es[1].Time)
	}
}

func TestSniffsUnknownFileName(t *testing.T) {
	d := t.TempDir()
	es, _ := Load([]string{write(t, d, "myhist", ": 1700000000:0;ls -la\n")})
	eq(t, lines(es), []string{"ls -la"})
}

func TestLoadMergesByTimeUntimedFirst(t *testing.T) {
	d := t.TempDir()
	z := write(t, d, ".zsh_history", ": 200:0;second\n: 100:0;first\n")
	b := write(t, d, ".bash_history", "untimed\n")
	es, _ := Load([]string{z, b})
	eq(t, lines(es), []string{"untimed", "first", "second"})
}

func TestSources(t *testing.T) {
	home := t.TempDir()
	write(t, home, ".zsh_history", "x")
	custom := write(t, home, "custom", "x")
	got := Sources(home, custom, "")
	eq(t, got, []string{custom, filepath.Join(home, ".zsh_history")})
	if got := Sources(home, filepath.Join(home, ".zsh_history"), ""); len(got) != 1 {
		t.Errorf("duplicate path not collapsed: %v", got)
	}
}

func TestMatch(t *testing.T) {
	for _, c := range []struct {
		line string
		cmd  string
		args []string
		want bool
	}{
		{"git status", "git", nil, true},
		{"git commit -m 'x y'", "git", []string{"commit"}, true},
		{"git commit -m 'x y'", "git", []string{"commit", "-m", "x y"}, true},
		{"git commit", "git", []string{"commit", "-m"}, false},
		{"git commit", "git", []string{"push"}, false},
		{"gitk", "git", nil, false},
		{"/opt/homebrew/bin/git log", "git", []string{"log"}, true},
		{"git log", "/usr/bin/git", nil, true},
		{"FOO=1 BAR=2 git pull", "git", []string{"pull"}, true},
		{"sudo -E brew upgrade", "brew", []string{"upgrade"}, true},
		{"time env -i git fetch", "git", []string{"fetch"}, true},
		{"ls | grep foo", "grep", []string{"foo"}, true},
		{"make && git push origin main", "git", []string{"push"}, true},
		{"cd x; git diff", "git", nil, true},
		{"(git stash)", "git", []string{"stash"}, true},
		{"echo $(git rev-parse HEAD)", "git", []string{"rev-parse"}, true},
		{`\git status`, "git", nil, true},
		{"echo git status", "git", nil, false},
		{"echo 'a && git push'", "git", nil, false},
		{"grep git file", "git", nil, false},
		{"make 2>&1 | tee log", "tee", nil, true},
		{"make 2>&1 | tee log", "make", nil, true},
		{"", "git", nil, false},
	} {
		if got := Match(c.line, c.cmd, c.args); got != c.want {
			t.Errorf("Match(%q, %q, %q) = %v, want %v", c.line, c.cmd, c.args, got, c.want)
		}
	}
}

func TestFilter(t *testing.T) {
	es := []Entry{{Line: "git a"}, {Line: "ls"}, {Line: "git b"}}
	eq(t, lines(Filter(es, "git", nil)), []string{"git a", "git b"})
	eq(t, lines(Filter(es, "git", []string{"b"})), []string{"git b"})
}

package classify

import "testing"

func TestClassify(t *testing.T) {
	for path, want := range map[string]Kind{
		"/bin/ls":                          Native,
		"/usr/bin/git":                     Native,
		"/usr/bin2/x":                      User,
		"/opt/homebrew/Cellar/jq/1/bin/jq": User,
		"/usr/local/bin/foo":               User,
	} {
		if got := Classify(path); got != want {
			t.Errorf("%s: got %s want %s", path, got, want)
		}
	}
}

func TestSource(t *testing.T) {
	for path, want := range map[string]string{
		"/opt/homebrew/Cellar/jq/1.7/bin/jq": "brew",
		"/Users/a/.cargo/bin/rg":             "cargo",
		"/Users/a/go/bin/x":                  "go",
		"/usr/local/bin/foo":                 "manual",
		"/somewhere/else":                    "unknown",
	} {
		if got := Source(path, "/Users/a"); got != want {
			t.Errorf("%s: got %s want %s", path, got, want)
		}
	}
}

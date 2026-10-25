package provider

import "testing"

func TestFormula(t *testing.T) {
	if got := formula("/opt/homebrew/Cellar/jq/1.7/bin/jq"); got != "jq" {
		t.Errorf("got %q", got)
	}
	if got := formula("/usr/bin/ls"); got != "" {
		t.Errorf("got %q", got)
	}
}

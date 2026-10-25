package provider

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/tassan/orch/internal/scan"
)

type Brew struct{}

func (Brew) Name() string    { return "brew" }
func (Brew) Available() bool { return lookPath("brew") }

// formula parses ".../Cellar/<formula>/<version>/..." from a resolved path.
func formula(real string) string {
	_, rest, ok := strings.Cut(real, "/Cellar/")
	if !ok {
		return ""
	}
	name, _, _ := strings.Cut(rest, "/")
	return name
}

func (Brew) Owns(t scan.Tool) bool { return formula(t.Real) != "" }

func (Brew) UninstallCmd(t scan.Tool) (*exec.Cmd, error) {
	f := formula(t.Real)
	if f == "" {
		return nil, fmt.Errorf("%s is not a brew formula", t.Name)
	}
	return exec.Command("brew", "uninstall", f), nil
}

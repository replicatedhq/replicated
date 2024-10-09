package homebrew

import (
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/version/pkgmgr"
)

type HomebrewExternalPackageManager struct {
	formula string
}

var _ pkgmgr.ExternalPackageManager = (*HomebrewExternalPackageManager)(nil)

type homebrewInfoOutput struct {
	Installed []struct {
		Version string `json:"version"`
	} `json:"installed"`
}

func NewHomebrewExternalPackageManager(formula string) pkgmgr.ExternalPackageManager {
	return HomebrewExternalPackageManager{
		formula: formula,
	}
}

func (m HomebrewExternalPackageManager) UpgradeCommand() string {
	return fmt.Sprintf("brew upgrade %s", m.formula)
}

// IsInstalled will return true if the formula is installed using homebrew
func (m HomebrewExternalPackageManager) IsInstalled() (bool, error) {
	path, err := exec.LookPath("brew")
	if err != nil {
		// we just assume that it wasn't installed via brew if there's no brew command
		return false, nil
	}

	out, err := exec.Command(
		path,
		"info",
		m.formula,
		"--json",
	).Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			if exitError.ExitCode() == 1 {
				// brew info with an invalid (not installed) package name returns an error
				return false, nil
			}
		}
		return false, errors.Wrap(err, "exec brew")
	}

	unmarshaled := []homebrewInfoOutput{}
	if err := json.Unmarshal(out, &unmarshaled); err != nil {
		return false, errors.Wrap(err, "unmarshal brew output")
	}

	if len(unmarshaled) == 0 {
		return false, nil
	}

	if len(unmarshaled[0].Installed) == 0 {
		return false, nil
	}

	return true, nil
}

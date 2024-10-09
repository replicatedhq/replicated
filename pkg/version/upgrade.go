package version

import (
	"fmt"

	"github.com/pkg/errors"
)

func VerifyCanUpgrade() error {
	// we support in place upgrades
	// only if not installed from a package manager

	updateChecker, err := NewUpdateChecker(build.Version, "replicatedhq/replicated/cli")
	if err != nil {
		return errors.New("create update checker")
	}

	canSupportUpgrade, err := updateChecker.CanSupportUpgrade()
	if err != nil {
		return errors.Wrap(err, "check if can support upgrade")
	}

	if !canSupportUpgrade {
		upgradeCmd := updateChecker.ExternalUpgradeCommand()
		fmt.Println("replicated was install using a package manager.")
		if upgradeCmd != "" {
			fmt.Printf("To upgrade, try running %q\n", upgradeCmd)
		}

		return errors.New("Upgrade not supported on this platform")
	}

	return nil
}

func PerformUpgrade() error {
	updateChecker, err := NewUpdateChecker(build.Version, "replicatedhq/replicated/cli")
	if err != nil {
		return errors.New("create update checker")
	}

	return updateChecker.Upgrade()
}

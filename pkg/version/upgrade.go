package version

import (
	"fmt"

	"github.com/pkg/errors"
)

func VerifyCanUpgrade() error {
	// we support in place upgrades
	// only if not installed from a package manager

	usrbinsdk, err := NewUsrbinSDK(build.Version)
	if err != nil {
		return errors.Wrap(err, "create usrbin")
	}

	canSupportUpgrade, err := usrbinsdk.CanSupportUpgrade()
	if err != nil {
		return errors.Wrap(err, "check if can support upgrade")
	}

	if !canSupportUpgrade {
		upgradeCmd := usrbinsdk.ExternalUpgradeCommand()
		fmt.Println("replicated was install using a package manager.")
		if upgradeCmd != "" {
			fmt.Printf("To upgrade, try running %q\n", upgradeCmd)
		}

		return errors.New("Upgrade not supported on this platform")
	}

	return nil
}

func PerformUpgrade() error {
	usrbinsdk, err := NewUsrbinSDK(build.Version)
	if err != nil {
		return errors.Wrap(err, "create usrbin")
	}

	return usrbinsdk.Upgrade()
}

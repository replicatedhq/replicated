package version

import (
	"os"
	"time"

	"github.com/minio/selfupdate"
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/version/homebrew"
	"github.com/replicatedhq/replicated/pkg/version/pkgmgr"
)

type UpdateChecker struct {
	version                 string
	externalPackageManagers []pkgmgr.ExternalPackageManager
	httpTimeout             time.Duration
}

type Option func(*UpdateChecker) error

func NewUpdateChecker(version string, homebrewFormula string) (*UpdateChecker, error) {
	updateChecker := UpdateChecker{
		version:     version,
		httpTimeout: 3 * time.Second, // Default timeout
	}

	if homebrewFormula != "" {
		updateChecker.externalPackageManagers = append(updateChecker.externalPackageManagers, homebrew.NewHomebrewExternalPackageManager(homebrewFormula))
	}

	return &updateChecker, nil
}

func (s UpdateChecker) CanSupportUpgrade() (bool, error) {
	for _, epm := range s.externalPackageManagers {
		isInstalled, err := epm.IsInstalled()
		if err != nil {
			return false, err
		}
		if isInstalled {
			return false, nil
		}
	}

	return true, nil
}

func (s UpdateChecker) ExternalUpgradeCommand() string {
	for _, epm := range s.externalPackageManagers {
		isInstalled, err := epm.IsInstalled()
		if err != nil {
			return ""
		}
		if isInstalled {
			return epm.UpgradeCommand()
		}
	}

	return ""
}

func (s UpdateChecker) Upgrade() error {
	// assume the latest
	updateInfo, err := s.GetUpdateInfo()
	if err != nil {
		return errors.Wrap(err, "get update info")
	}

	if updateInfo == nil {
		return errors.New("no update info")
	}

	newVersionPath, err := downloadVersion(updateInfo.LatestVersion, true)
	if err != nil {
		return errors.Wrap(err, "download version")
	}

	f, err := os.Open(newVersionPath)
	if err != nil {
		return errors.Wrap(err, "open new version")
	}
	defer func() {
		if err := f.Close(); err != nil {
			panic(err)
		}
	}()

	err = selfupdate.Apply(f, selfupdate.Options{})
	if err != nil {
		return errors.Wrap(err, "apply update")
	}

	return nil
}

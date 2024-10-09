package version

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Masterminds/semver"
	"github.com/pkg/errors"
)

const (
	latestVersionURI = `https://replicated.app/ping`
)

// GetUpdateInfo will return the latest version
func (s UpdateChecker) GetUpdateInfo() (*UpdateInfo, error) {
	checkedAt := time.Now()

	latestVersion, err := getLatestVersion(s.httpTimeout)
	if err != nil {
		return nil, errors.Wrap(err, "get latest version")
	}

	if latestVersion == nil {
		return nil, nil
	}

	updateInfo, err := UpdateInfoFromVersions(s.version, latestVersion)
	if err != nil {
		return nil, errors.Wrap(err, "update info from versions")
	}
	if updateInfo == nil {
		return nil, nil
	}

	updateInfo.ExternalUpgradeCommand = s.ExternalUpgradeCommand()
	updateInfo.CanUpgradeInPlace = updateInfo.ExternalUpgradeCommand == ""
	updateInfo.CheckedAt = &checkedAt

	return updateInfo, nil
}

type PingResponse struct {
	ClientIP       string         `json:"client_ip"`
	ClientVersions ClientVersions `json:"client_versions"`
}

type ClientVersions struct {
	ReplicatedCLI string `json:"replicated_cli"`
	ReplicatedSDK string `json:"replicated_sdk"`
	KOTS          string `json:"kots"`
}

func getLatestVersion(timeout time.Duration) (*VersionInfo, error) {
	client := &http.Client{
		Timeout: timeout,
	}

	resp, err := client.Get(latestVersionURI)
	if err != nil {
		return nil, errors.Wrap(err, "get latest version")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	pingResponse := PingResponse{}
	err = json.NewDecoder(resp.Body).Decode(&pingResponse)
	if err != nil {
		return nil, errors.Wrap(err, "decode response")
	}

	if pingResponse.ClientVersions.ReplicatedCLI == "" {
		return nil, nil
	}

	return &VersionInfo{
		Version: pingResponse.ClientVersions.ReplicatedCLI,
	}, nil
}

type VersionInfo struct {
	Version string `json:"version"`
}

type UpdateInfo struct {
	LatestVersion   string     `json:"latestVersion"`
	LatestReleaseAt *time.Time `json:"latestReleaseAt"`

	CheckedAt *time.Time `json:"checkedAt"`

	CanUpgradeInPlace      bool   `json:"canUpgradeInPlace"`
	ExternalUpgradeCommand string `json:"externalUpgradeCommand"`
}

func UpdateInfoFromVersions(currentVersion string, latestVersion *VersionInfo) (*UpdateInfo, error) {
	if latestVersion == nil {
		return nil, errors.New("latest version is nil")
	}

	latestSemver, err := semver.NewVersion(latestVersion.Version)
	if err != nil {
		return nil, errors.Wrap(err, "latest semver")
	}

	currentSemver, err := semver.NewVersion(currentVersion)
	if err != nil {
		return nil, errors.Wrap(err, "current semver")
	}

	if latestSemver.LessThan(currentSemver) || latestSemver.Equal(currentSemver) {
		return nil, nil
	}

	updateInfo := UpdateInfo{
		LatestVersion: latestVersion.Version,
	}

	return &updateInfo, nil
}

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
	debugLog("Getting update info for version %s", s.version)
	checkedAt := time.Now()

	debugLog("Fetching latest version with timeout %v", s.httpTimeout)
	latestVersion, err := getLatestVersion(s.httpTimeout)
	if err != nil {
		debugLog("Error getting latest version: %v", err)
		return nil, errors.Wrap(err, "get latest version")
	}

	if latestVersion == nil {
		debugLog("No latest version info returned (nil)")
		return nil, nil
	}

	debugLog("Got latest version: %s", latestVersion.Version)
	updateInfo, err := UpdateInfoFromVersions(s.version, latestVersion)
	if err != nil {
		debugLog("Error creating update info: %v", err)
		return nil, nil
	}
	if updateInfo == nil {
		debugLog("No update needed, current version is latest or newer")
		return nil, nil
	}

	debugLog("Update available: %s", updateInfo.LatestVersion)
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
	startTime := time.Now()
	debugLog("Making HTTP request to %s with timeout %v", latestVersionURI, timeout)
	
	client := &http.Client{
		Timeout: timeout,
	}

	resp, err := client.Get(latestVersionURI)
	if err != nil {
		debugLog("HTTP request failed: %v", err)
		return nil, errors.Wrap(err, "get latest version")
	}
	defer resp.Body.Close()

	elapsed := time.Since(startTime)
	debugLog("HTTP request completed in %v with status %d", elapsed, resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		debugLog("Unexpected status code: %d", resp.StatusCode)
		return nil, errors.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	pingResponse := PingResponse{}
	err = json.NewDecoder(resp.Body).Decode(&pingResponse)
	if err != nil {
		debugLog("Failed to decode response: %v", err)
		return nil, errors.Wrap(err, "decode response")
	}

	if pingResponse.ClientVersions.ReplicatedCLI == "" {
		debugLog("No CLI version found in response")
		return nil, nil
	}

	debugLog("Latest CLI version from server: %s", pingResponse.ClientVersions.ReplicatedCLI)
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

	// Try to parse the latest version
	latestSemver, err := semver.NewVersion(latestVersion.Version)
	if err != nil {
		debugLog("Failed to parse latest version as semver: %v", err)
		return nil, errors.Wrap(err, "latest semver")
	}

	// Special handling for "unknown" or otherwise invalid current versions
	// In this case, we'll always return update info since we can't compare versions
	if currentVersion == "" || currentVersion == "unknown" || currentVersion == "development" {
		debugLog("Current version is '%s', assuming update is available", currentVersion)
		return &UpdateInfo{
			LatestVersion: latestVersion.Version,
		}, nil
	}

	// Parse the current version, but handle errors gracefully
	currentSemver, err := semver.NewVersion(currentVersion)
	if err != nil {
		debugLog("Failed to parse current version '%s' as semver: %v", currentVersion, err)
		// Since we can't parse the current version, we'll assume an update is available
		// This is safer than failing silently
		return &UpdateInfo{
			LatestVersion: latestVersion.Version,
		}, nil
	}

	// Compare versions and only return update info if a newer version is available
	if latestSemver.LessThan(currentSemver) || latestSemver.Equal(currentSemver) {
		debugLog("Current version %s is equal to or newer than latest %s, no update needed", 
			currentVersion, latestVersion.Version)
		return nil, nil
	}

	debugLog("Update available: current=%s, latest=%s", currentVersion, latestVersion.Version)
	updateInfo := UpdateInfo{
		LatestVersion: latestVersion.Version,
	}

	return &updateInfo, nil
}

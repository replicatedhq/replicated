package version

import (
	"fmt"
	"os"
)

var (
	build Build
)

// Build holds details about this build of the replicated cli binary
type Build struct {
	Version    string      `json:"version,omitempty"`
	UpdateInfo *UpdateInfo `json:"updateInfo,omitempty"`
}

// initBuild sets up the version info from build args
func initBuild() {
	build.Version = version

	// Don't check for updates in CI environments
	if os.Getenv("CI") == "true" {
		return
	}

	// Try to load update info from cache
	updateCache, _ := LoadUpdateCache()
	if updateCache != nil {
		// Only use cached update info if it's for the current version
		if updateCache.Version == build.Version {
			debugLog("Using cached update info for version %s", build.Version)
			build.UpdateInfo = &updateCache.UpdateInfo
		} else {
			debugLog("Cached update info is for version %s, current version is %s - clearing cache", 
				updateCache.Version, build.Version)
			// Clear the cache when CLI version has changed
			ClearUpdateCache()
		}
	}

	// Start background update check (never blocks)
	CheckForUpdatesInBackground(build.Version, "replicatedhq/replicated/cli")
}

func GetBuild() Build {
	return build
}

func Version() string {
	return build.Version
}

// SetUpdateInfo updates the current build's update info
func SetUpdateInfo(updateInfo *UpdateInfo) {
	build.UpdateInfo = updateInfo
}

func Print() {
	fmt.Printf("replicated version %s\n", build.Version)

	if build.UpdateInfo != nil {
		fmt.Printf("Update available: %s\n", build.UpdateInfo.LatestVersion)
		if build.UpdateInfo.CanUpgradeInPlace {
			fmt.Printf("To automatically upgrade, run \"replicated version upgrade\"\n")
		} else {
			fmt.Printf("To upgrade, run \"%s\"\n", build.UpdateInfo.ExternalUpgradeCommand)
		}
	}
}

// PrintToStdErrIfUpgradeAvailable prints the update info to stderr if available
func PrintIfUpgradeAvailable() {
	if build.UpdateInfo != nil {
		fmt.Fprintf(os.Stderr, "Update available: %s\n", build.UpdateInfo.LatestVersion)
		if build.UpdateInfo.CanUpgradeInPlace {
			fmt.Fprintf(os.Stderr, "To automatically upgrade, run \"replicated version upgrade\"\n")
		} else {
			fmt.Fprintf(os.Stderr, "To upgrade, run \"%s\"\n", build.UpdateInfo.ExternalUpgradeCommand)
		}
	}
}
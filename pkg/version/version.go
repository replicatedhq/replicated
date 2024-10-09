package version

import (
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/pkg/errors"
)

var (
	build Build
)

// Build holds details about this build of the replicated cli binary
type Build struct {
	Version      string      `json:"version,omitempty"`
	GitSHA       string      `json:"git,omitempty"`
	BuildTime    time.Time   `json:"buildTime,omitempty"`
	TimeFallback string      `json:"buildTimeFallback,omitempty"`
	GoInfo       GoInfo      `json:"go,omitempty"`
	UpdateInfo   *UpdateInfo `json:"updateInfo,omitempty"`
}

type GoInfo struct {
	Version  string `json:"version,omitempty"`
	Compiler string `json:"compiler,omitempty"`
	OS       string `json:"os,omitempty"`
	Arch     string `json:"arch,omitempty"`
}

// initBuild sets up the version info from build args
func initBuild() {
	build.Version = version
	if len(gitSHA) >= 7 {
		build.GitSHA = gitSHA[:7]
	}
	var err error
	build.BuildTime, err = time.Parse(time.RFC3339, buildTime)
	if err != nil {
		build.TimeFallback = buildTime
	}

	build.GoInfo = getGoInfo()

	updateChecker, err := NewUpdateChecker(build.Version, "replicatedhq/replicated/cli")
	if err != nil {
		return
	}

	build.UpdateInfo, err = updateChecker.GetUpdateInfo()
	if err != nil {
		if errors.Cause(err) == ErrTimeoutExceeded {
			// i'm going to leave this println out for now because it could be really noisy
			// for someone with a slow connection
			// fmt.Fprintln(os.Stderr, "Unable to check for updates, timeout exceeded.")
			return
		}
		fmt.Fprintf(os.Stderr, "Error getting update info: %s", err)
	}
}

// GetBuild gets the build
func GetBuild() Build {
	return build
}

// Version gets the version
func Version() string {
	return build.Version
}

// GitSHA gets the gitsha
func GitSHA() string {
	return build.GitSHA
}

// BuildTime gets the build time
func BuildTime() time.Time {
	return build.BuildTime
}

func getGoInfo() GoInfo {
	return GoInfo{
		Version:  runtime.Version(),
		Compiler: runtime.Compiler,
		OS:       runtime.GOOS,
		Arch:     runtime.GOARCH,
	}
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

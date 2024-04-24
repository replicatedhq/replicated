package version

import (
	"os"
	"path/filepath"

	"github.com/usrbinapp/usrbin-go"
)

func NewUsrbinSDK(currentVersion string) (*usrbin.SDK, error) {
	return usrbin.New(
		currentVersion,
		usrbin.UsingGitHubUpdateChecker("github.com/replicatedhq/replicated"),
		usrbin.UsingHomebrewFormula("replicatedhq/replicated/cli"),
		usrbin.UsingCacheDir(filepath.Join(homeDir(), ".replicated", "cache.json")),
	)
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE")
}

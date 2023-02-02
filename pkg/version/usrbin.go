package version

import "github.com/usrbinapp/usrbin-go"

func NewUsrbinSDK(currentVersion string) (*usrbin.SDK, error) {
	return usrbin.New(
		currentVersion,
		usrbin.UsingGitHubUpdateChecker("github.com/replicatedhq/replicated"),
		usrbin.UsingHomebrewFormula("replicatedhq/replicated/cli"),
	)
}

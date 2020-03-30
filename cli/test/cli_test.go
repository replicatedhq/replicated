package test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
)

func TestCLI(t *testing.T) {
	if os.Getenv("SKIP_INTEGRATION_TESTING") != "" {
		return
	}
	RunSpecs(t, "CLI Suite")
}

var _ = AfterSuite(cleanupApps)

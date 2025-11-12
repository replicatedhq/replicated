//go:build integration
// +build integration

package tools

import (
	"testing"
)

// TestGetLatestStableVersion_EmbeddedCluster verifies that the ping API
// returns a version for embedded-cluster and that we can parse it correctly.
func TestGetLatestStableVersion_EmbeddedCluster(t *testing.T) {
	version, err := getLatestStableVersion(ToolEmbeddedCluster)
	if err != nil {
		t.Fatalf("getLatestStableVersion failed for embedded-cluster: %v", err)
	}

	if version == "" {
		t.Fatal("expected non-empty version, got empty string")
	}

	t.Logf("✓ Got embedded-cluster version from ping API: %s", version)

	// Verify version format (should be like "2.12.0+k8s-1.33")
	// At minimum, should contain digits
	hasDigit := false
	for _, ch := range version {
		if ch >= '0' && ch <= '9' {
			hasDigit = true
			break
		}
	}

	if !hasDigit {
		t.Errorf("version %q doesn't contain any digits", version)
	}
}

// TestGetLatestStableVersion_AllTools verifies that all supported tools
// can retrieve versions from the ping API, including embedded-cluster and KOTS.
func TestGetLatestStableVersion_AllTools(t *testing.T) {
	tools := []struct {
		name string
		tool string
	}{
		{"Helm", ToolHelm},
		{"Preflight", ToolPreflight},
		{"Support Bundle", ToolSupportBundle},
		{"Embedded Cluster", ToolEmbeddedCluster},
		{"KOTS", ToolKots},
	}

	for _, tc := range tools {
		t.Run(tc.name, func(t *testing.T) {
			version, err := getLatestStableVersion(tc.tool)
			if err != nil {
				t.Fatalf("getLatestStableVersion failed for %s: %v", tc.name, err)
			}

			if version == "" {
				t.Fatalf("expected non-empty version for %s, got empty string", tc.name)
			}

			t.Logf("✓ %s version: %s", tc.name, version)
		})
	}
}

package tools

import (
	"testing"
)

// Note: FetchRecommendedVersions() is tested via integration tests in cli/cmd/lint_test.go
// (TestLint_VersionWarnings* tests validate the full flow including API calls)

func TestCompareVersions_MajorVersionBehind(t *testing.T) {
	result := CompareVersions("3.0.0", "4.0.0")
	if !result {
		t.Error("Expected true when major version is behind, got false")
	}
}

func TestCompareVersions_MinorVersionBehind(t *testing.T) {
	result := CompareVersions("3.14.0", "3.15.0")
	if !result {
		t.Error("Expected true when minor version is behind, got false")
	}
}

func TestCompareVersions_PatchVersionBehind(t *testing.T) {
	result := CompareVersions("3.14.4", "3.14.5")
	if result {
		t.Error("Expected false for patch-only difference, got true")
	}
}

func TestCompareVersions_SameVersion(t *testing.T) {
	result := CompareVersions("3.14.0", "3.14.0")
	if result {
		t.Error("Expected false when versions are equal, got true")
	}
}

func TestCompareVersions_AheadOfRecommended(t *testing.T) {
	result := CompareVersions("3.16.0", "3.15.0")
	if result {
		t.Error("Expected false when ahead of recommended, got true")
	}
}

func TestCompareVersions_WithVPrefix(t *testing.T) {
	result := CompareVersions("v3.14.0", "3.15.0")
	if !result {
		t.Error("Expected true when minor version is behind (with v prefix), got false")
	}

	result = CompareVersions("3.14.0", "v3.15.0")
	if !result {
		t.Error("Expected true when minor version is behind (recommended has v prefix), got false")
	}
}

func TestCompareVersions_InvalidConfiguredVersion(t *testing.T) {
	result := CompareVersions("not-a-version", "3.15.0")
	if result {
		t.Error("Expected false for invalid configured version (silent failure), got true")
	}
}

func TestCompareVersions_InvalidRecommendedVersion(t *testing.T) {
	result := CompareVersions("3.14.0", "not-a-version")
	if result {
		t.Error("Expected false for invalid recommended version (silent failure), got true")
	}
}

func TestCompareVersions_BothInvalid(t *testing.T) {
	result := CompareVersions("invalid", "also-invalid")
	if result {
		t.Error("Expected false when both versions are invalid, got true")
	}
}

func TestCompareVersions_ComplexVersions(t *testing.T) {
	// Test with prerelease versions
	result := CompareVersions("3.14.0-beta.1", "3.14.0")
	// Prerelease versions are considered less than release versions in semver
	// but we only care about major/minor, so this should return false
	if result {
		t.Error("Expected false for prerelease vs release (same minor), got true")
	}

	// Test major version behind with prerelease
	result = CompareVersions("3.14.0-beta.1", "4.0.0")
	if !result {
		t.Error("Expected true for major version behind (with prerelease), got false")
	}
}

func TestCompareVersions_MinorVersionAheadButMajorBehind(t *testing.T) {
	// 2.99.0 vs 3.0.0 - major version behind should warn
	result := CompareVersions("2.99.0", "3.0.0")
	if !result {
		t.Error("Expected true when major version is behind (even if minor is high), got false")
	}
}

func TestCompareVersions_EmbeddedClusterVersionFormat(t *testing.T) {
	// EC versions can have + (e.g., "1.33.0+k8s-1.33")
	// Semver treats everything after + as metadata (ignored in comparison)
	result := CompareVersions("1.32.0+k8s-1.32", "1.33.0+k8s-1.33")
	if !result {
		t.Error("Expected true when minor version is behind (EC format), got false")
	}

	// Same version with different metadata should not warn
	result = CompareVersions("1.33.0+k8s-1.33", "1.33.0+k8s-1.34")
	if result {
		t.Error("Expected false when only metadata differs (same version), got true")
	}
}

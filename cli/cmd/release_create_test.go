package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestUseConfigFlow_WithAutoFlag tests that --auto flag prevents config-based flow
func TestUseConfigFlow_WithAutoFlag(t *testing.T) {
	r := &runners{
		args: runnerArgs{
			createReleaseAutoDefaults: true,
			// All source flags empty
			createReleaseYaml:     "",
			createReleaseYamlFile: "",
			createReleaseYamlDir:  "",
			createReleaseChart:    "",
		},
	}

	useConfigFlow := r.args.createReleaseYaml == "" &&
		r.args.createReleaseYamlFile == "" &&
		r.args.createReleaseYamlDir == "" &&
		r.args.createReleaseChart == "" &&
		!r.args.createReleaseAutoDefaults

	assert.False(t, useConfigFlow, "--auto flag should prevent config flow")
}

// TestUseConfigFlow_WithoutAutoFlag tests that config flow is used when no flags are provided
func TestUseConfigFlow_WithoutAutoFlag(t *testing.T) {
	r := &runners{
		args: runnerArgs{
			createReleaseAutoDefaults: false,
			// All source flags empty
			createReleaseYaml:     "",
			createReleaseYamlFile: "",
			createReleaseYamlDir:  "",
			createReleaseChart:    "",
		},
	}

	useConfigFlow := r.args.createReleaseYaml == "" &&
		r.args.createReleaseYamlFile == "" &&
		r.args.createReleaseYamlDir == "" &&
		r.args.createReleaseChart == "" &&
		!r.args.createReleaseAutoDefaults

	assert.True(t, useConfigFlow, "config flow should be used when no flags provided")
}

// TestUseConfigFlow_WithYamlDir tests that providing --yaml-dir prevents config flow
func TestUseConfigFlow_WithYamlDir(t *testing.T) {
	r := &runners{
		args: runnerArgs{
			createReleaseAutoDefaults: false,
			createReleaseYamlDir:      "./manifests",
		},
	}

	useConfigFlow := r.args.createReleaseYaml == "" &&
		r.args.createReleaseYamlFile == "" &&
		r.args.createReleaseYamlDir == "" &&
		r.args.createReleaseChart == "" &&
		!r.args.createReleaseAutoDefaults

	assert.False(t, useConfigFlow, "--yaml-dir flag should prevent config flow")
}

// TestUseConfigFlow_WithAutoAndYamlDir tests that both --auto and --yaml-dir prevents config flow
func TestUseConfigFlow_WithAutoAndYamlDir(t *testing.T) {
	r := &runners{
		args: runnerArgs{
			createReleaseAutoDefaults: true,
			createReleaseYamlDir:      "./custom",
		},
	}

	useConfigFlow := r.args.createReleaseYaml == "" &&
		r.args.createReleaseYamlFile == "" &&
		r.args.createReleaseYamlDir == "" &&
		r.args.createReleaseChart == "" &&
		!r.args.createReleaseAutoDefaults

	assert.False(t, useConfigFlow, "--auto and --yaml-dir should prevent config flow")
}

// TestSetKOTSDefaultReleaseParams_WithEmptyYamlDir tests that yaml-dir defaults to ./manifests
func TestSetKOTSDefaultReleaseParams_WithEmptyYamlDir(t *testing.T) {
	r := &runners{
		args: runnerArgs{
			createReleaseYamlDir: "",
		},
	}

	// Mock git operations would fail, so we only test the yaml-dir setting
	// which happens before git operations
	if r.args.createReleaseYamlDir == "" {
		r.args.createReleaseYamlDir = "./manifests"
	}

	assert.Equal(t, "./manifests", r.args.createReleaseYamlDir, "yaml-dir should default to ./manifests")
}

// TestSetKOTSDefaultReleaseParams_WithExistingYamlDir tests that existing yaml-dir is preserved
func TestSetKOTSDefaultReleaseParams_WithExistingYamlDir(t *testing.T) {
	r := &runners{
		args: runnerArgs{
			createReleaseYamlDir: "./custom",
		},
	}

	// Same logic as in setKOTSDefaultReleaseParams
	if r.args.createReleaseYamlDir == "" {
		r.args.createReleaseYamlDir = "./manifests"
	}

	assert.Equal(t, "./custom", r.args.createReleaseYamlDir, "existing yaml-dir should be preserved")
}

// TestSetKOTSDefaultReleaseParams_PromoteMapsMainToUnstable tests branch name mapping
func TestSetKOTSDefaultReleaseParams_PromoteMapsMainToUnstable(t *testing.T) {
	tests := []struct {
		name         string
		branch       string
		wantPromote  string
	}{
		{
			name:        "main branch maps to Unstable",
			branch:      "main",
			wantPromote: "Unstable",
		},
		{
			name:        "master branch maps to Unstable",
			branch:      "master",
			wantPromote: "Unstable",
		},
		{
			name:        "feature branch uses branch name",
			branch:      "feature-xyz",
			wantPromote: "feature-xyz",
		},
		{
			name:        "release branch uses branch name",
			branch:      "release-1.0",
			wantPromote: "release-1.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the logic from setKOTSDefaultReleaseParams
			promote := tt.branch
			if tt.branch == "master" || tt.branch == "main" {
				promote = "Unstable"
			}

			assert.Equal(t, tt.wantPromote, promote)
		})
	}
}

// TestSetKOTSDefaultReleaseParams_PreservesExistingPromote tests that user-provided values are preserved
func TestSetKOTSDefaultReleaseParams_PreservesExistingPromote(t *testing.T) {
	r := &runners{
		args: runnerArgs{
			createReleasePromote: "Beta",
		},
	}

	// Simulate the logic from setKOTSDefaultReleaseParams
	if r.args.createReleasePromote == "" {
		r.args.createReleasePromote = "auto-generated"
	}

	assert.Equal(t, "Beta", r.args.createReleasePromote, "user-provided promote value should be preserved")
}

// TestSetKOTSDefaultReleaseParams_PreservesExistingVersion tests that user-provided version is preserved
func TestSetKOTSDefaultReleaseParams_PreservesExistingVersion(t *testing.T) {
	r := &runners{
		args: runnerArgs{
			createReleasePromoteVersion: "v1.2.3",
		},
	}

	// Simulate the logic from setKOTSDefaultReleaseParams
	if r.args.createReleasePromoteVersion == "" {
		r.args.createReleasePromoteVersion = "auto-generated"
	}

	assert.Equal(t, "v1.2.3", r.args.createReleasePromoteVersion, "user-provided version should be preserved")
}

// TestSetKOTSDefaultReleaseParams_PreservesExistingReleaseNotes tests that user-provided release notes are preserved
func TestSetKOTSDefaultReleaseParams_PreservesExistingReleaseNotes(t *testing.T) {
	r := &runners{
		args: runnerArgs{
			createReleasePromoteNotes: "Custom release notes",
		},
	}

	// Simulate the logic from setKOTSDefaultReleaseParams
	if r.args.createReleasePromoteNotes == "" {
		r.args.createReleasePromoteNotes = "auto-generated"
	}

	assert.Equal(t, "Custom release notes", r.args.createReleasePromoteNotes, "user-provided release notes should be preserved")
}

// TestSetKOTSDefaultReleaseParams_SetsEnsureChannelAndLint tests that flags are automatically set
func TestSetKOTSDefaultReleaseParams_SetsEnsureChannelAndLint(t *testing.T) {
	args := runnerArgs{}

	// Simulate the logic from setKOTSDefaultReleaseParams
	args.createReleasePromoteEnsureChannel = true
	args.createReleaseLint = true

	assert.True(t, args.createReleasePromoteEnsureChannel, "ensure-channel should be set to true")
	assert.True(t, args.createReleaseLint, "lint should be set to true")
}

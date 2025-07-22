package cmd

import (
	"errors"
	"fmt"
	"testing"

	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestCleanImageName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "images.shortrib.io/anonymous/index.docker.io/library/nginx:1.25.3",
			expected: "docker.io/library/nginx:1.25.3",
		},
		{
			input:    "images.shortrib.io/proxy/testapp/ghcr.io/example/app:v1.0.0",
			expected: "ghcr.io/example/app:v1.0.0",
		},
		{
			input:    "docker.io/library/postgres:14",
			expected: "docker.io/library/postgres:14",
		},
		{
			input:    "index.docker.io/replicated/replicated-sdk:1.0.0-beta.32",
			expected: "docker.io/replicated/replicated-sdk:1.0.0-beta.32",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := cleanImageName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFindReleaseLogic(t *testing.T) {
	tests := []struct {
		name               string
		releases           []*types.ChannelRelease
		requestedVersion   string
		expectedSequence   int32
		expectError        bool
		errorMsg           string
	}{
		{
			name: "find current release (highest channel sequence)",
			releases: []*types.ChannelRelease{
				{ChannelSequence: 5, Semver: "1.0.0"},
				{ChannelSequence: 10, Semver: "1.1.0"},
				{ChannelSequence: 8, Semver: "1.0.5"},
			},
			requestedVersion: "",
			expectedSequence: 10,
		},
		{
			name: "find specific version",
			releases: []*types.ChannelRelease{
				{ChannelSequence: 5, Semver: "1.0.0"},
				{ChannelSequence: 10, Semver: "1.1.0"},
				{ChannelSequence: 8, Semver: "1.0.5"},
			},
			requestedVersion: "1.0.5",
			expectedSequence: 8,
		},
		{
			name: "version not found",
			releases: []*types.ChannelRelease{
				{ChannelSequence: 5, Semver: "1.0.0"},
				{ChannelSequence: 10, Semver: "1.1.0"},
			},
			requestedVersion: "2.0.0",
			expectError:      true,
			errorMsg:         "no release found with version \"2.0.0\"",
		},
		{
			name:             "no releases",
			releases:         []*types.ChannelRelease{},
			requestedVersion: "",
			expectError:      true,
			errorMsg:         "no releases found in channel",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the logic from channelImageLS function
			if len(tt.releases) == 0 {
				err := errors.New("no releases found in channel")
				if tt.expectError {
					assert.Contains(t, err.Error(), tt.errorMsg)
				} else {
					t.Errorf("Unexpected error: %v", err)
				}
				return
			}

			var targetRelease *types.ChannelRelease
			if tt.requestedVersion != "" {
				// Find release by semver
				for _, release := range tt.releases {
					if release.Semver == tt.requestedVersion {
						targetRelease = release
						break
					}
				}
				if targetRelease == nil {
					err := fmt.Errorf("no release found with version %q in channel", tt.requestedVersion)
					if tt.expectError {
						assert.Contains(t, err.Error(), tt.errorMsg)
					} else {
						t.Errorf("Unexpected error: %v", err)
					}
					return
				}
			} else {
				// Find the current release (highest channel sequence)
				for _, release := range tt.releases {
					if targetRelease == nil || release.ChannelSequence > targetRelease.ChannelSequence {
						targetRelease = release
					}
				}
				if targetRelease == nil {
					err := errors.New("no current release found")
					if tt.expectError {
						assert.Contains(t, err.Error(), tt.errorMsg)
					} else {
						t.Errorf("Unexpected error: %v", err)
					}
					return
				}
			}

			if tt.expectError {
				t.Errorf("Expected error but got none")
			} else {
				assert.Equal(t, tt.expectedSequence, targetRelease.ChannelSequence)
			}
		})
	}
}
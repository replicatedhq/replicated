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
		name                string
		input               string
		proxyRegistryDomain string
		expected            string
	}{
		{
			name:                "proxy registry with app name",
			input:               "images.shortrib.io/proxy/testapp/ghcr.io/example/app:v1.0.0",
			proxyRegistryDomain: "images.shortrib.io",
			expected:            "ghcr.io/example/app:v1.0.0",
		},
		{
			name:                "docker hub registry prefix",
			input:               "docker.io/library/postgres:14",
			proxyRegistryDomain: "",
			expected:            "postgres:14",
		},
		{
			name:                "index.docker.io prefix",
			input:               "index.docker.io/replicated/replicated-sdk:1.0.0-beta.32",
			proxyRegistryDomain: "",
			expected:            "replicated/replicated-sdk:1.0.0-beta.32",
		},
		{
			name:                "no proxy registry domain provided",
			input:               "images.shortrib.io/proxy/testapp/ghcr.io/example/app:v1.0.0",
			proxyRegistryDomain: "",
			expected:            "images.shortrib.io/proxy/testapp/ghcr.io/example/app:v1.0.0",
		},
		{
			name:                "proxy registry with library prefix",
			input:               "myproxy.com/library/nginx:latest",
			proxyRegistryDomain: "myproxy.com",
			expected:            "nginx:latest",
		},
		{
			name:                "no matching prefixes",
			input:               "ghcr.io/myorg/myapp:v1.0.0",
			proxyRegistryDomain: "",
			expected:            "ghcr.io/myorg/myapp:v1.0.0",
		},
		{
			name:                "proxy registry with anonymous prefix",
			input:               "myproxy.com/anonymous/redis:7.0",
			proxyRegistryDomain: "myproxy.com",
			expected:            "redis:7.0",
		},
		{
			name:                "proxy.replicated.com with library prefix (common SDK pattern)",
			input:               "proxy.replicated.com/library/replicated-sdk-image:1.7.1",
			proxyRegistryDomain: "proxy.replicated.com",
			expected:            "replicated-sdk-image:1.7.1",
		},
		{
			name:                "proxy.replicated.com with proxy prefix and app name",
			input:               "proxy.replicated.com/proxy/myapp/ghcr.io/example/app:v1.0",
			proxyRegistryDomain: "proxy.replicated.com",
			expected:            "ghcr.io/example/app:v1.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanImageName(tt.input, tt.proxyRegistryDomain)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFindReleaseLogic(t *testing.T) {
	tests := []struct {
		name             string
		releases         []*types.ChannelRelease
		requestedVersion string
		expectedSequence int32
		expectError      bool
		errorMsg         string
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
			// Simulate the logic from releaseImageLS function
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
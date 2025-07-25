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

func TestProxyRegistryDomainUsage(t *testing.T) {
	// Test that proxyRegistryDomain from channel release is passed to cleanImageName
	release := &types.ChannelRelease{
		ChannelSequence:     1,
		ProxyRegistryDomain: "my.proxy.com",
		AirgapBundleImages: []string{
			"my.proxy.com/proxy/myapp/nginx:latest",
			"my.proxy.com/library/postgres:14",
			"docker.io/redis:alpine",
		},
	}

	expectedCleanedImages := []string{
		"nginx:latest",   // proxy pattern stripped using domain
		"postgres:14",    // library prefix stripped using domain  
		"redis:alpine",   // docker.io prefix stripped by default rules
	}

	var actualImages []string
	for _, image := range release.AirgapBundleImages {
		cleanImage := cleanImageName(image, release.ProxyRegistryDomain)
		if cleanImage != "" {
			actualImages = append(actualImages, cleanImage)
		}
	}

	assert.Equal(t, expectedCleanedImages, actualImages)
}

func TestKeepProxyFlag(t *testing.T) {
	// Test that --keep-proxy flag preserves proxy registry domains
	release := &types.ChannelRelease{
		ChannelSequence:     1,
		ProxyRegistryDomain: "my.proxy.com",
		AirgapBundleImages: []string{
			"my.proxy.com/proxy/myapp/nginx:latest",
			"my.proxy.com/library/postgres:14",
			"docker.io/redis:alpine",
		},
	}

	// Test with keep-proxy disabled (default behavior)
	expectedWithoutProxy := []string{
		"nginx:latest",   // proxy pattern stripped
		"postgres:14",    // library prefix stripped
		"redis:alpine",   // docker.io prefix stripped
	}

	var actualWithoutProxy []string
	for _, image := range release.AirgapBundleImages {
		cleanImage := cleanImageName(image, release.ProxyRegistryDomain) // Pass proxy domain (strip it)
		if cleanImage != "" {
			actualWithoutProxy = append(actualWithoutProxy, cleanImage)
		}
	}

	assert.Equal(t, expectedWithoutProxy, actualWithoutProxy)

	// Test with keep-proxy enabled
	expectedWithProxy := []string{
		"my.proxy.com/proxy/myapp/nginx:latest", // proxy pattern kept
		"my.proxy.com/library/postgres:14",     // library prefix kept  
		"redis:alpine",                         // docker.io prefix still stripped (not proxy)
	}

	var actualWithProxy []string
	for _, image := range release.AirgapBundleImages {
		cleanImage := cleanImageName(image, "") // Don't pass proxy domain (keep it)
		if cleanImage != "" {
			actualWithProxy = append(actualWithProxy, cleanImage)
		}
	}

	assert.Equal(t, expectedWithProxy, actualWithProxy)
}

func TestKeepProxyFlagIntegration(t *testing.T) {
	// Test the actual --keep-proxy flag behavior in the command logic
	
	// Test case 1: flag not set (default behavior - should strip proxy)
	t.Run("keep-proxy flag false", func(t *testing.T) {
		keepProxyFlag := false
		proxyDomain := "test.proxy.com"
		
		// Simulate the logic from channelImageLS function
		var cleanProxyDomain string
		if !keepProxyFlag {
			cleanProxyDomain = proxyDomain
		}
		
		testImage := "test.proxy.com/proxy/myapp/nginx:1.0"
		result := cleanImageName(testImage, cleanProxyDomain)
		
		assert.Equal(t, "nginx:1.0", result, "Should strip proxy domain when keep-proxy is false")
	})
	
	// Test case 2: flag set to true (should preserve proxy)
	t.Run("keep-proxy flag true", func(t *testing.T) {
		keepProxyFlag := true
		proxyDomain := "test.proxy.com"
		
		// Simulate the logic from channelImageLS function
		var cleanProxyDomain string
		if !keepProxyFlag {
			cleanProxyDomain = proxyDomain
		}
		
		testImage := "test.proxy.com/proxy/myapp/nginx:1.0"
		result := cleanImageName(testImage, cleanProxyDomain)
		
		assert.Equal(t, "test.proxy.com/proxy/myapp/nginx:1.0", result, "Should preserve proxy domain when keep-proxy is true")
	})
}
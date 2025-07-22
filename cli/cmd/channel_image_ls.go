package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/spf13/cobra"
)

func (r *runners) InitChannelImageLS(parent *cobra.Command) {
	imageCmd := &cobra.Command{
		Use:   "image",
		Short: "Manage channel images",
		Long:  "Manage channel images",
	}

	lsCmd := &cobra.Command{
		Use:   "ls --channel CHANNEL_NAME_OR_ID [--version SEMVER] [--keep-proxy]",
		Short: "List images in a channel's current or specified release",
		Long:  "List all container images in the current release or a specific version of a channel",
		Example: `# List images in current release of a channel by name
replicated channel image ls --channel Stable

# List images in a specific version of a channel
replicated channel image ls --channel Stable --version 1.2.1

# List images in a channel by ID  
replicated channel image ls --channel 2abc123

# Keep proxy registry domains in the image names
replicated channel image ls --channel Stable --keep-proxy`,
	}

	lsCmd.Flags().StringVar(&r.args.channelImageLSChannel, "channel", "", "The channel name, slug, or ID (required)")
	lsCmd.Flags().StringVar(&r.args.channelImageLSVersion, "version", "", "The specific semver version to get images for (optional, defaults to current release)")
	lsCmd.Flags().BoolVar(&r.args.channelImageLSKeepProxy, "keep-proxy", false, "Keep proxy registry domain in image names instead of stripping it")
	lsCmd.MarkFlagRequired("channel")

	parent.AddCommand(imageCmd)
	imageCmd.AddCommand(lsCmd)
	lsCmd.RunE = r.channelImageLS
}

func (r *runners) channelImageLS(cmd *cobra.Command, args []string) error {
	if !r.hasApp() {
		return errors.New("no app specified")
	}

	if r.args.channelImageLSChannel == "" {
		return errors.New("channel is required")
	}

	// Get the channel to find its current release
	channel, err := r.api.GetChannelByName(r.appID, r.appType, r.args.channelImageLSChannel)
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}

	// Get all releases for this channel
	channelReleases, err := r.api.ListChannelReleases(r.appID, r.appType, channel.ID)
	if err != nil {
		return fmt.Errorf("failed to list channel releases: %w", err)
	}

	if len(channelReleases) == 0 {
		return errors.New("no releases found in channel")
	}

	// Find the target release
	var targetRelease *types.ChannelRelease
	if r.args.channelImageLSVersion != "" {
		// Find release by semver
		for _, release := range channelReleases {
			if release.Semver == r.args.channelImageLSVersion {
				targetRelease = release
				break
			}
		}
		if targetRelease == nil {
			return fmt.Errorf("no release found with version %q in channel", r.args.channelImageLSVersion)
		}
	} else {
		// Find the current release (highest channel sequence)
		for _, release := range channelReleases {
			if targetRelease == nil || release.ChannelSequence > targetRelease.ChannelSequence {
				targetRelease = release
			}
		}
		if targetRelease == nil {
			return errors.New("no current release found")
		}
	}

	// Extract and clean up image names
	images := make([]string, 0)
	
	// Get proxy domain from multiple sources in order of preference:
	// 1. Channel release ProxyRegistryDomain field
	// 2. Channel customHostnameOverrides.proxy.hostname
	proxyDomain := targetRelease.ProxyRegistryDomain
	if proxyDomain == "" {
		// Try to get from channel custom hostname overrides
		if r.appType == "kots" {
			customHostnames, err := r.api.GetCustomHostnames(r.appID, r.appType, channel.ID)
			if err == nil && customHostnames.Proxy.Hostname != "" {
				proxyDomain = customHostnames.Proxy.Hostname
			}
		}
	}
	
	for _, image := range targetRelease.AirgapBundleImages {
		// Remove registry prefixes and clean up image names
		var cleanProxyDomain string
		if !r.args.channelImageLSKeepProxy {
			cleanProxyDomain = proxyDomain
		}
		cleanImage := cleanImageName(image, cleanProxyDomain)
		if cleanImage != "" {
			images = append(images, cleanImage)
		}
	}

	// Print images
	return print.ChannelImages(r.w, images)
}

func cleanImageName(image string, proxyRegistryDomain string) string {
	cleaned := image

	// Remove proxy registry domain if provided and present
	if proxyRegistryDomain != "" {
		// Handle proxy registry patterns like "proxyRegistryDomain/proxy/app-name/"
		proxyPrefix := proxyRegistryDomain + "/proxy/"
		if strings.HasPrefix(cleaned, proxyPrefix) {
			// Remove prefix and find first occurrence after app name
			withoutPrefix := strings.TrimPrefix(cleaned, proxyPrefix)
			parts := strings.SplitN(withoutPrefix, "/", 2)
			if len(parts) == 2 {
				cleaned = parts[1] // Take everything after the app name
			}
		}
		// Also handle anonymous proxy registry domain prefixes
		if strings.HasPrefix(cleaned, proxyRegistryDomain+"/anonymous/") {
			cleaned = strings.TrimPrefix(cleaned, proxyRegistryDomain+"/anonymous/")
		}

		// And library proxy domain prefixes
		if strings.HasPrefix(cleaned, proxyRegistryDomain+"/library/") {
			cleaned = strings.TrimPrefix(cleaned, proxyRegistryDomain+"/library/")
		}
	}

	// Remove other common registry prefixes
	prefixes := []string{
		"registry-1.docker.io/library/",
		"registry-1.docker.io/",
		"docker.io/library/",
		"docker.io/",
		"index.docker.io/library/",
		"index.docker.io/",
		"hub.docker.com/library/",
		"hub.docker.com/",
		"registry.hub.docker.com/library/",
		"registry.hub.docker.com/",
	}

	for _, prefix := range prefixes {
		if strings.HasPrefix(cleaned, prefix) {
			cleaned = strings.TrimPrefix(cleaned, prefix)
		}
	}

	return cleaned
}

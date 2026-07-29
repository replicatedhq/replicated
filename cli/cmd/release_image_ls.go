package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/replicatedhq/replicated/cli/print"
	"github.com/replicatedhq/replicated/pkg/types"
	"github.com/spf13/cobra"
)

func (r *runners) InitReleaseImageLS(parent *cobra.Command) {
	imageCmd := &cobra.Command{
		Use:   "image",
		Short: "Manage release images",
		Long:  "Manage release images",
	}

	lsCmd := &cobra.Command{
		Use:   "ls --channel CHANNEL_NAME_OR_ID [--version SEMVER] [--keep-proxy]",
		Short: "List images in a channel's current or specified release",
		Long:  "List all container images in the current release or a specific version of a channel",
		Example: `# List images in current release of a channel by name
replicated release image ls --channel Stable

# List images in a specific version of a channel
replicated release image ls --channel Stable --version 1.2.1

# List images in a channel by ID  
replicated release image ls --channel 2abc123

# Keep proxy registry domains in the image names
replicated release image ls --channel Stable --keep-proxy`,
	}

	lsCmd.Flags().StringVar(&r.args.releaseImageLSChannel, "channel", "", "The channel name, slug, or ID (required)")
	lsCmd.Flags().StringVar(&r.args.releaseImageLSVersion, "version", "", "The specific semver version to get images for (optional, defaults to current release)")
	lsCmd.Flags().BoolVar(&r.args.releaseImageLSKeepProxy, "keep-proxy", false, "Keep proxy registry domain in image names instead of stripping it")
	lsCmd.Flags().StringVar(&r.args.releaseImageLSIncludeInstallerImages, "include-installer-images", "", "Include installer images in the output (valid values: online, airgap)")
	lsCmd.MarkFlagRequired("channel")

	parent.AddCommand(imageCmd)
	imageCmd.AddCommand(lsCmd)
	lsCmd.RunE = r.releaseImageLS
}

func (r *runners) releaseImageLS(cmd *cobra.Command, args []string) error {
	if !r.hasApp() {
		return errors.New("no app specified")
	}

	if r.args.releaseImageLSChannel == "" {
		return errors.New("channel is required")
	}

	// Get the channel to find its current release
	channel, err := r.api.GetChannelByName(r.appID, r.appType, r.args.releaseImageLSChannel)
	if err != nil {
		return fmt.Errorf("failed to get channel: %w", err)
	}

	var targetRelease *types.ChannelRelease
	var proxyDomain string

	if r.args.releaseImageLSVersion != "" {
		// For specific versions, try the server-side versionLabel filter first.
		// This avoids downloading every page of a large channel history.
		targetRelease, err = r.findReleaseByVersion(channel.ID, r.args.releaseImageLSVersion)
		if err != nil {
			return err
		}

		// Get proxy domain for version-specific releases
		proxyDomain = targetRelease.ProxyRegistryDomain
		if proxyDomain == "" && r.appType == "kots" {
			customHostnames, err := r.api.GetCustomHostnames(r.appID, r.appType, channel.ID)
			if err == nil && customHostnames.Proxy.Hostname != "" {
				proxyDomain = customHostnames.Proxy.Hostname
			}
			// Check embedded cluster proxy domain if still empty
			if proxyDomain == "" && targetRelease.InstallationTypes.EmbeddedCluster.ProxyRegistryDomain != "" {
				proxyDomain = targetRelease.InstallationTypes.EmbeddedCluster.ProxyRegistryDomain
			}
			// If no explicit proxy domain is configured, fall back to default custom hostname
			if proxyDomain == "" {
				defaultProxy, err := r.api.GetDefaultProxyHostname(r.appID)
				if err == nil && defaultProxy != "" {
					proxyDomain = defaultProxy
				} else {
					// Final fallback to default Replicated proxy
					proxyDomain = "proxy.replicated.com"
				}
			}
		}
	} else {
		// For current release, use optimized method that tries to avoid extra API call
		var err error
		targetRelease, proxyDomain, err = r.api.GetCurrentChannelRelease(r.appID, r.appType, channel.ID, r.args.releaseImageLSIncludeInstallerImages)
		if err != nil {
			return fmt.Errorf("failed to get current channel release: %w", err)
		}
	}

	// Extract and clean up image names
	images := make([]string, 0)

	for _, image := range targetRelease.AirgapBundleImages {
		// Remove registry prefixes and clean up image names
		var cleanProxyDomain string
		if r.args.releaseImageLSKeepProxy {
			// Keep proxy domain - don't strip it
			cleanProxyDomain = ""
		} else {
			// Strip proxy domain
			cleanProxyDomain = proxyDomain
		}
		cleanImage := cleanImageName(image, cleanProxyDomain)
		if cleanImage != "" {
			images = append(images, cleanImage)
		}
	}

	// Print images
	return print.ChannelImages(r.outputFormat, r.w, images)
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
		cleaned = strings.TrimPrefix(cleaned, prefix)
	}

	return cleaned
}

// findReleaseByVersion returns the channel release matching the requested version.
// ListChannelReleasesByVersion uses a server-side versionLabel filter and returns
// matching releases sorted by channel sequence, so the first result is the target.
func (r *runners) findReleaseByVersion(channelID string, version string) (*types.ChannelRelease, error) {
	releases, err := r.api.ListChannelReleasesByVersion(r.appID, r.appType, channelID, version, r.args.releaseImageLSIncludeInstallerImages)
	if err != nil {
		return nil, fmt.Errorf("failed to list channel releases: %w", err)
	}
	if len(releases) == 0 {
		return nil, fmt.Errorf("no release found with version %q in channel", version)
	}
	return releases[0], nil
}

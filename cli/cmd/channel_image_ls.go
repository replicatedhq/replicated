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
		Use:   "ls --channel CHANNEL_NAME_OR_ID [--version SEMVER]",
		Short: "List images in a channel's current or specified release",
		Long:  "List all container images in the current release or a specific version of a channel",
		Example: `# List images in current release of a channel by name
replicated channel image ls --channel Stable

# List images in a specific version of a channel
replicated channel image ls --channel Stable --version 1.2.1

# List images in a channel by ID  
replicated channel image ls --channel 2abc123`,
	}
	
	lsCmd.Flags().StringVar(&r.args.channelImageLSChannel, "channel", "", "The channel name, slug, or ID (required)")
	lsCmd.Flags().StringVar(&r.args.channelImageLSVersion, "version", "", "The specific semver version to get images for (optional, defaults to current release)")
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
	for _, image := range targetRelease.AirgapBundleImages {
		// Remove registry prefixes and clean up image names
		cleanImage := cleanImageName(image)
		if cleanImage != "" {
			images = append(images, cleanImage)
		}
	}

	// Print images
	return print.ChannelImages(r.w, images)
}

func cleanImageName(image string) string {
	cleaned := image
	
	// Handle proxy registry patterns first - "images.shortrib.io/proxy/app-name/"
	if strings.HasPrefix(cleaned, "images.shortrib.io/proxy/") {
		// Remove prefix and find first occurrence after app name
		withoutPrefix := strings.TrimPrefix(cleaned, "images.shortrib.io/proxy/")
		parts := strings.SplitN(withoutPrefix, "/", 2)
		if len(parts) == 2 {
			cleaned = parts[1] // Take everything after the app name
		}
	}
	
	// Remove other common registry prefixes
	prefixes := []string{
		"images.shortrib.io/anonymous/index.",
		"images.shortrib.io/anonymous/",
		"index.",
	}
	
	for _, prefix := range prefixes {
		if strings.HasPrefix(cleaned, prefix) {
			cleaned = strings.TrimPrefix(cleaned, prefix)
		}
	}
	
	return cleaned
}
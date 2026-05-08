package cmd

import (
	"errors"

	"github.com/replicatedhq/replicated/cli/print"
	"github.com/spf13/cobra"
)

func (r *runners) InitChannelReleases(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "releases CHANNEL_ID_OR_NAME",
		Short: "List all releases in a channel",
		Long:  "List all releases promoted to a channel, including demoted releases. Accepts a channel ID or name.",
		Example: `# List releases for a channel by name
replicated channel releases Stable

# List releases for a channel by ID
replicated channel releases 2abc123

# JSON output for scripting or AI agents
replicated channel releases Stable --output json

# Paginate (second page of 50)
replicated channel releases Stable --page 1 --page-size 50`,
	}
	parent.AddCommand(cmd)
	cmd.Flags().StringVarP(&r.outputFormat, "output", "o", "table", "The output format to use. One of: json|table")
	cmd.Flags().IntVar(&r.args.channelReleasesPage, "page", 0, "The page to fetch (KOTS apps only).")
	cmd.Flags().IntVar(&r.args.channelReleasesPageSize, "page-size", 0, "The number of releases per page (KOTS apps only).")

	cmd.RunE = r.channelReleases
}

func (r *runners) channelReleases(cmd *cobra.Command, args []string) error {
	if !r.hasApp() {
		return errors.New("no app specified")
	}

	if len(args) != 1 {
		return errors.New("channel name or ID is required")
	}
	channelNameOrID := args[0]

	if r.args.channelReleasesPage != 0 && r.args.channelReleasesPageSize == 0 {
		return errors.New("--page requires --page-size")
	}

	channel, err := r.api.GetChannelByName(r.appID, r.appType, channelNameOrID)
	if err != nil {
		return err
	}

	if r.appType == "platform" {
		_, releases, err := r.platformAPI.GetChannel(r.appID, channel.ID)
		if err != nil {
			return err
		}

		return print.ChannelReleases(r.outputFormat, r.w, releases)
	} else if r.appType == "kots" {
		releases, err := r.api.ListChannelReleasesPaged(r.appID, r.appType, channel.ID, "", r.args.channelReleasesPage, r.args.channelReleasesPageSize)
		if err != nil {
			return err
		}

		return print.KotsChannelReleases(r.outputFormat, r.w, releases)
	}

	return errors.New("unknown app type")
}

package cmd

import (
	"errors"
	"fmt"

	"github.com/replicatedhq/replicated/client"
	"github.com/spf13/cobra"
)

func (r *runners) InitChannelReleaseDemote(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "demote CHANNEL_ID_OR_NAME",
		Short: "Demote a release from a channel",
		Long: `Demote a release from a channel.

  Example:
  replicated channel release demote Beta --channel-sequence 15
  replicated channel release demote Beta --release-sequence 12`,
		Args: cobra.ExactArgs(1),
	}

	cmd.Flags().Int64Var(&r.args.demoteChannelSequence, "channel-sequence", 0, "The channel sequence to demote")
	cmd.Flags().Int64Var(&r.args.demoteReleaseSequence, "release-sequence", 0, "The release sequence to demote")

	parent.AddCommand(cmd)
	cmd.RunE = r.channelReleaseDemote
}

func (r *runners) channelReleaseDemote(cmd *cobra.Command, args []string) error {
	if !r.hasApp() {
		return errors.New("no app specified")
	}

	if r.args.demoteChannelSequence == 0 && r.args.demoteReleaseSequence == 0 {
		return errors.New("one of --channel-sequence or --release-sequence is required")
	}

	if r.args.demoteChannelSequence != 0 && r.args.demoteReleaseSequence != 0 {
		fmt.Fprintf(r.w, "Warning: Both --channel-sequence and --release-sequence provided. Using --channel-sequence %d\n", r.args.demoteChannelSequence)
	}

	chanID := args[0]

	opts := client.GetOrCreateChannelOptions{
		AppID:          r.appID,
		AppType:        r.appType,
		NameOrID:       chanID,
		CreateIfAbsent: false,
	}
	foundChannel, err := r.api.GetOrCreateChannelByName(opts)
	if err != nil {
		return err
	}

	channelSequence := r.args.demoteChannelSequence
	if r.args.demoteReleaseSequence != 0 {
		kotsChannel, err := r.api.KotsClient.GetKotsChannel(r.appID, foundChannel.ID)
		if err != nil {
			return err
		}

		for _, channelRelease := range kotsChannel.Releases {
			if int64(channelRelease.Sequence) == r.args.demoteReleaseSequence {
				channelSequence = int64(channelRelease.ChannelSequence)
				break
			}
		}

		if channelSequence == 0 {
			return fmt.Errorf("release sequence %d not found in channel %s", r.args.demoteReleaseSequence, foundChannel.ID)
		}
	}

	demotedRelease, err := r.api.ChannelReleaseDemote(r.appID, r.appType, foundChannel.ID, channelSequence)
	if err != nil {
		return err
	}

	fmt.Fprintf(r.w, "Channel sequence %d (version %s, release %d) demoted in channel %s\n", channelSequence, demotedRelease.Semver, demotedRelease.Sequence, chanID)
	r.w.Flush()

	return nil
}

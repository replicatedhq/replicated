package cmd

import (
	"errors"
	"fmt"

	"github.com/replicatedhq/replicated/client"
	"github.com/spf13/cobra"
)

func (r *runners) InitChannelReleaseUnDemote(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:   "un-demote CHANNEL_ID_OR_NAME",
		Short: "Un-demote a release from a channel",
		Long: `Un-demote a release from a channel.

  Example:
  replicated channel release un-demote Beta --channel-sequence 15
  replicated channel release un-demote Beta --release-sequence 12`,
		Args: cobra.ExactArgs(1),
	}

	cmd.Flags().Int64Var(&r.args.unDemoteChannelSequence, "channel-sequence", 0, "The channel sequence to un-demote")
	cmd.Flags().Int64Var(&r.args.unDemoteReleaseSequence, "release-sequence", 0, "The release sequence to un-demote")

	parent.AddCommand(cmd)
	cmd.RunE = r.channelReleaseUnDemote
}

func (r *runners) channelReleaseUnDemote(cmd *cobra.Command, args []string) error {
	if !r.hasApp() {
		return errors.New("no app specified")
	}

	if r.args.unDemoteChannelSequence == 0 && r.args.unDemoteReleaseSequence == 0 {
		return errors.New("one of --channel-sequence or --release-sequence is required")
	}

	if r.args.unDemoteChannelSequence != 0 && r.args.unDemoteReleaseSequence != 0 {
		fmt.Fprintf(r.w, "Warning: Both --channel-sequence and --release-sequence provided. Using --channel-sequence %d\n", r.args.unDemoteChannelSequence)
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

	channelSequence := r.args.unDemoteChannelSequence
	if r.args.unDemoteReleaseSequence != 0 {
		kotsChannel, err := r.api.KotsClient.GetKotsChannel(r.appID, foundChannel.ID)
		if err != nil {
			return err
		}

		for _, channelRelease := range kotsChannel.Releases {
			if int64(channelRelease.Sequence) == r.args.unDemoteReleaseSequence {
				channelSequence = int64(channelRelease.ChannelSequence)
				break
			}
		}

		if channelSequence == 0 {
			return fmt.Errorf("release sequence %d not found in channel %s", r.args.unDemoteReleaseSequence, foundChannel.ID)
		}
	}

	unDemotedRelease, err := r.api.ChannelReleaseUnDemote(r.appID, r.appType, foundChannel.ID, channelSequence)
	if err != nil {
		return err
	}

	fmt.Fprintf(r.w, "Channel sequence %d (version %s, release %d) un-demoted in channel %s\n", channelSequence, unDemotedRelease.Semver, unDemotedRelease.Sequence, chanID)
	r.w.Flush()

	return nil
}

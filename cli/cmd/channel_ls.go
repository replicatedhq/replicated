package cmd

import (
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/cli/print"
	"github.com/spf13/cobra"
)

func (r *runners) InitChannelList(parent *cobra.Command) {
	cmd := &cobra.Command{
		Use:     "ls",
		Aliases: []string{"list"},
		Short:   "List all channels in your app",
		Long:    "List all channels in your app",
	}

	parent.AddCommand(cmd)
	cmd.Flags().StringVar(&r.outputFormat, "output", "table", "The output format to use. One of: json|table (default: table)")

	cmd.RunE = r.channelList
}

func (r *runners) channelList(cmd *cobra.Command, args []string) error {
	if !r.hasApp() {
		return errors.New("no app specified")
	}

	channels, err := r.api.ListChannels(r.appID, r.appType, "")
	if err != nil {
		return err
	}

	return print.Channels(r.outputFormat, r.w, channels)
}

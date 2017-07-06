package cmd

import (
	"github.com/spf13/cobra"
)

// channelCmd represents the channel command
var channelCmd = &cobra.Command{
	Use:   "channel",
	Short: "Manage and review channels",
	Long:  "Manage and review channels",
}

func init() {
	RootCmd.AddCommand(channelCmd)
}

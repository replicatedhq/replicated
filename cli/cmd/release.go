package cmd

import (
	"github.com/spf13/cobra"
)

// releaseCmd represents the release command
var releaseCmd = &cobra.Command{
	Use:   "release",
	Short: "manage app releases",
	Long:  `The release command allows vendors to create, display, modify, and archive their releases.`,
}

func init() {
	RootCmd.AddCommand(releaseCmd)
}

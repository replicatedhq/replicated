package cmd

import (
	"github.com/spf13/cobra"
)

// releaseCmd represents the release command
var shipReleaseCommand = &cobra.Command{
	Use:    "shiprelease",
	Short:  "Manage ship releases",
	Long:   `The shiprelease command allows vendors to create, display, modify, and archive their Ship releases.`,
	Hidden: true,
}

func init() {
	RootCmd.AddCommand(shipReleaseCommand)
}

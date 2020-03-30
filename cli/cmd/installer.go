package cmd

import (
	"github.com/spf13/cobra"
)

func (r *runners) InitInstallerCommand(parent *cobra.Command) *cobra.Command {
	installerCommand := &cobra.Command{
		Use:          "installer",
		Short:        "Manage Kubernetes installers",
		Long:         `The installers command allows vendors to create, display, modify and promote kurl.sh specs for managing the installation of Kubernetes.`,
		SilenceUsage: true,
	}
	parent.AddCommand(installerCommand)

	return installerCommand
}

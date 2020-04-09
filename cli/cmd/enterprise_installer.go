package cmd

import (
	"github.com/spf13/cobra"
)

func (r *runners) InitEnterpriseInstaller(parent *cobra.Command) *cobra.Command {
	enterpriseInstallerCommand := &cobra.Command{
		Use:   "installer",
		Short: "Manage enterprise installers",
		Long:  `The installer command allows approved enterprise to create custom installers`,
	}
	parent.AddCommand(enterpriseInstallerCommand)

	return enterpriseInstallerCommand
}
